package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

var nsAlnum = regexp.MustCompile(`[^a-z0-9]`)

// nsNorm — seriya normalizasiyası (kiçik hərf, yalnız hərf/rəqəm)
func nsNorm(s string) string {
	return nsAlnum.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "")
}

// syncNotebookFromExcel — BİRDƏFƏLİK: Excel (NOTEBOOK/продажа/REALIZACIYA) əsasında
// Mərkəz (branch_id=1) komputerlərini uyğunlaşdırır. Plan /data/notebook_sync.json
// faylından oxunur (data public repo-ya düşmür). Marker ilə yalnız bir dəfə; idempotent
// (seriya ilə uyğunlaşır). Yalnız Mərkəzə toxunur; qiymətə (satış/optavoy) toxunmur.
func syncNotebookFromExcel() {
	if getSetting("notebook_sync_v1_done") == "1" {
		return
	}
	raw, err := os.ReadFile("/data/notebook_sync.json")
	if err != nil {
		return // plan faylı yoxdursa — heç nə etmə
	}
	var plan struct {
		Sales []struct {
			Serial string  `json:"serial"`
			Cost   float64 `json:"cost"`
			Sale   float64 `json:"sale"`
			Date   string  `json:"date"`
		} `json:"sales"`
		Deletes []struct {
			Serial string `json:"serial"`
		} `json:"deletes"`
		Adds []struct {
			Name   string            `json:"name"`
			Serial string            `json:"serial"`
			Cost   float64           `json:"cost"`
			Specs  map[string]string `json:"specs"`
		} `json:"adds"`
		Completes []struct {
			Serial string            `json:"serial"`
			Specs  map[string]string `json:"specs"`
		} `json:"completes"`
	}
	if err := json.Unmarshal(raw, &plan); err != nil {
		log.Printf("notebook sync: json xəta: %v", err)
		return
	}

	// atribut adı → id
	attrID := map[string]uint{}
	var attrs []Attribute
	db.Find(&attrs)
	for _, a := range attrs {
		attrID[a.Name] = a.ID
	}
	var nbCat Category
	db.Where("name = ?", "Notebook").First(&nbCat)

	loadMerkez := func() map[string]uint {
		var items []Item
		db.Where("branch_id = ?", 1).Select("id, serial").Find(&items)
		m := map[string]uint{}
		for _, it := range items {
			if it.Serial != "" {
				m[nsNorm(it.Serial)] = it.ID
			}
		}
		return m
	}
	byser := loadMerkez()

	// bütün mövcud seriyalar (adds-də dublikat olmasın)
	allSer := map[string]bool{}
	{
		var items []Item
		db.Select("serial").Find(&items)
		for _, it := range items {
			if it.Serial != "" {
				allSer[nsNorm(it.Serial)] = true
			}
		}
	}

	sales, dels, adds, comps := 0, 0, 0, 0

	// 1) SALES — satışa at (Excel продажа alış/satış qiyməti + tarixi ilə)
	for _, s := range plan.Sales {
		id, ok := byser[nsNorm(s.Serial)]
		if !ok {
			continue
		}
		var it Item
		db.First(&it, id)
		var cnt int64
		db.Model(&Sale{}).Where("item_id = ?", id).Count(&cnt)
		if cnt == 0 {
			cost := it.Cost
			if cost == 0 {
				cost = s.Cost
				db.Model(&Item{}).Where("id = ?", id).Update("cost", s.Cost)
			}
			soldAt := time.Now()
			if t, e := time.Parse("2006-01-02", s.Date); e == nil {
				soldAt = t
			}
			db.Create(&Sale{
				ItemID: id, SalePrice: s.Sale, Quantity: 1, Profit: s.Sale - cost,
				Channel: "cash", BranchID: 1, SoldAt: soldAt, Counted: true,
			})
			sales++
		}
		db.Model(&Item{}).Where("id = ?", id).Update("status", "sold")
	}

	// 2) DELETES — köhnə satılıb in_stock qalan cihazlar → soft delete
	for _, d := range plan.Deletes {
		if id, ok := byser[nsNorm(d.Serial)]; ok {
			db.Delete(&Item{}, id)
			dels++
		}
	}

	// 3) ADDS — Excel-də olub saytda olmayan yeni cihazlar (Mərkəz, in_stock, saytda gizli)
	if nbCat.ID != 0 {
		for _, a := range plan.Adds {
			if allSer[nsNorm(a.Serial)] {
				continue // artıq var
			}
			it := Item{
				Name: a.Name, Serial: a.Serial, CategoryID: nbCat.ID, BranchID: 1,
				Cost: a.Cost, Status: "in_stock", Quantity: 1, ShowOnSite: false, CreatedAt: time.Now(),
			}
			for an, v := range a.Specs {
				if aid, ok := attrID[an]; ok {
					it.Values = append(it.Values, ItemAttributeValue{AttributeID: aid, Value: v})
				}
			}
			db.Omit("Values.Attribute").Create(&it)
			allSer[nsNorm(a.Serial)] = true
			adds++
		}
	}

	// 4) COMPLETES — yalnız ƏSKİK atributları əlavə et (mövcuda toxunma)
	byser = loadMerkez()
	for _, cm := range plan.Completes {
		id, ok := byser[nsNorm(cm.Serial)]
		if !ok {
			continue
		}
		for an, v := range cm.Specs {
			aid, ok := attrID[an]
			if !ok {
				continue
			}
			var ec int64
			db.Model(&ItemAttributeValue{}).Where("item_id = ? AND attribute_id = ?", id, aid).Count(&ec)
			if ec == 0 {
				db.Create(&ItemAttributeValue{ItemID: id, AttributeID: aid, Value: v})
				comps++
			}
		}
	}

	setSetting("notebook_sync_v1_done", "1")
	os.Remove("/data/notebook_sync.json") // datanı volume-dan təmizlə
	log.Printf("notebook sync tamam: sales=%d deletes=%d adds=%d specs=%d", sales, dels, adds, comps)
}

// notebookCleanupV2 — BİRDƏFƏLİK: (1) səhvən silinmiş 3 DELL PRO-nu bərpa;
// (2) Mərkəz status=sold cihazlardan son ayda counted satışı OLMAYANLARı sil (köhnə);
// (3) Mərkəz komputer adlarından «NOTEBOOK» prefiksini sil. Marker ilə bir dəfə.
func notebookCleanupV2() {
	if getSetting("notebook_cleanup_v2_done") == "1" {
		return
	}
	// 1) Excel NOTEBOOK-da olan, səhvən silinmiş DELL PRO cihazlarını bərpa et
	restored := 0
	for _, ser := range []string{"db03cd4", "1c03cd4", "1k03cd4"} {
		res := db.Unscoped().Model(&Item{}).
			Where("branch_id = ? AND LOWER(serial) = ? AND deleted_at IS NOT NULL", 1, ser).
			Update("deleted_at", nil)
		restored += int(res.RowsAffected)
	}

	// 2) köhnə satıldı → sil (Mərkəz status=sold, son 1 ayda counted satışı olmayanlar)
	cutoff := time.Now().AddDate(0, -1, 0)
	res := db.Where("branch_id = ? AND status = ? AND id NOT IN (?)", 1, "sold",
		db.Model(&Sale{}).Select("item_id").Where("counted = ? AND sold_at >= ?", true, cutoff)).
		Delete(&Item{})

	// 3) komputer adlarından «NOTEBOOK» (və yayılmış typoları) prefiksini sil (Mərkəz)
	var items []Item
	db.Where("branch_id = ? AND category_id IN (?)", 1, []uint{3, 5, 7}).Find(&items)
	renamed := 0
	prefixes := []string{"NOTEBOOK ", "NOTEBBOK ", "NETEBOOK ", "NETBOOK ", "NOTEBOOOK ", "NOTEBOK "}
	for _, it := range items {
		up := strings.ToUpper(it.Name)
		for _, pre := range prefixes {
			if strings.HasPrefix(up, pre) {
				n := strings.TrimSpace(it.Name[len(pre):])
				if n != "" && n != it.Name {
					db.Model(&Item{}).Where("id = ?", it.ID).Update("name", n)
					renamed++
				}
				break
			}
		}
	}

	setSetting("notebook_cleanup_v2_done", "1")
	log.Printf("notebook cleanup v2: bərpa=%d köhnə-silindi=%d ad-təmizləndi=%d", restored, res.RowsAffected, renamed)
}

// notebookRestoreNames — DÜZƏLİŞ: v2 bəzi komputer adlarından «NOTEBOOK» sözünü sildi.
// Orijinal adları /data/notebook_names.json (seriya → orijinal ad) faylından geri qaytarır.
// Yalnız Mərkəz; yalnız cari ad hələ «kəsilmiş» versiyaya bərabərdirsə (istifadəçi
// arada dəyişməyibsə). Marker + fayl ilə bir dəfə.
func notebookRestoreNames() {
	if getSetting("notebook_names_restore_done") == "1" {
		return
	}
	raw, err := os.ReadFile("/data/notebook_names.json")
	if err != nil {
		return
	}
	var m map[string]string // normalizə seriya → orijinal ad
	if err := json.Unmarshal(raw, &m); err != nil {
		log.Printf("notebook names restore: json xəta: %v", err)
		return
	}
	var items []Item
	db.Where("branch_id = ? AND category_id IN (?)", 1, []uint{3, 5, 7}).Find(&items)
	fixed := 0
	for _, it := range items {
		orig, ok := m[nsNorm(it.Serial)]
		if !ok || orig == "" || orig == it.Name {
			continue
		}
		db.Model(&Item{}).Where("id = ?", it.ID).Update("name", orig)
		fixed++
	}
	setSetting("notebook_names_restore_done", "1")
	os.Remove("/data/notebook_names.json")
	log.Printf("notebook names restore: %d ad geri qaytarıldı", fixed)
}

// addMissingAttributeOptions — BİRDƏFƏLİK: istifadədə olan, lakin options siyahısında
// olmayan atribut dəyərlərini option kimi əlavə edir (yeni prosessorlar və s. dropdown-da
// real seçim olsun). Prosessorda kəsik/zibil dəyərləri («I5», «... - 7») keçir.
func addMissingAttributeOptions() {
	if getSetting("attr_options_backfill_v1_done") == "1" {
		return
	}
	attrs := []uint{2, 3, 7, 8, 9, 13, 14} // RAM, Ekran ölçüsü, Prosessor, SSD, Ekran kartı, keyfiyyət, tezlik
	added := 0
	for _, aid := range attrs {
		var opts []AttributeOption
		db.Where("attribute_id = ?", aid).Find(&opts)
		existing := map[string]bool{}
		maxPos := 0
		for _, o := range opts {
			existing[o.Value] = true
			if o.Position > maxPos {
				maxPos = o.Position
			}
		}
		var vals []string
		db.Raw("SELECT DISTINCT value FROM item_attribute_values WHERE attribute_id = ?", aid).Scan(&vals)
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" || existing[v] {
				continue
			}
			if aid == 7 { // Prosessor — yalnız düzgün formatlı: "... - <ən azı 3 simvol model>"
				idx := strings.LastIndex(v, " - ")
				if idx < 0 || len(strings.TrimSpace(v[idx+3:])) < 3 {
					continue // "I5", "AMD Ryzen 7 - 7" kimi kəsik/zibili keç
				}
			}
			maxPos++
			db.Create(&AttributeOption{AttributeID: aid, Value: v, Position: maxPos})
			existing[v] = true
			added++
		}
	}
	setSetting("attr_options_backfill_v1_done", "1")
	log.Printf("attribute options backfill: %d yeni option əlavə olundu", added)
}

// applyIncrementalSync — BİRDƏFƏLİK: yeni Excel versiyasından inkremental əlavələr.
// /data/inc_plan.json-dan oxuyur: yeni NOTEBOOK cihazları (add), yeni satışlar (sales),
// yeni realizasiya (realiz). Yalnız Mərkəz; idempotent (seriya ilə); marker ilə bir dəfə.
func applyIncrementalSync() {
	if getSetting("inc_sync_v1_done") == "1" {
		return
	}
	raw, err := os.ReadFile("/data/inc_plan.json")
	if err != nil {
		return
	}
	var plan struct {
		Adds []struct {
			Name   string            `json:"name"`
			Serial string            `json:"serial"`
			Cost   float64           `json:"cost"`
			Specs  map[string]string `json:"specs"`
		} `json:"adds"`
		Sales []struct {
			Serial string  `json:"serial"`
			Cost   float64 `json:"cost"`
			Sale   float64 `json:"sale"`
			Date   string  `json:"date"`
		} `json:"sales"`
		Realiz []struct {
			Name   string            `json:"name"`
			Serial string            `json:"serial"`
			Cost   float64           `json:"cost"`
			Cat    uint              `json:"cat"`
			Specs  map[string]string `json:"specs"`
		} `json:"realiz"`
	}
	if err := json.Unmarshal(raw, &plan); err != nil {
		log.Printf("inc sync: json xəta: %v", err)
		return
	}
	attrID := map[string]uint{}
	var attrs []Attribute
	db.Find(&attrs)
	for _, a := range attrs {
		attrID[a.Name] = a.ID
	}
	var nbCat Category
	db.Where("name = ?", "Notebook").First(&nbCat)
	allSer := map[string]bool{}
	{
		var items []Item
		db.Select("serial").Find(&items)
		for _, it := range items {
			if it.Serial != "" {
				allSer[nsNorm(it.Serial)] = true
			}
		}
	}
	mk := func(catID uint, name, serial string, cost float64, specs map[string]string, status string) {
		if allSer[nsNorm(serial)] {
			return
		}
		it := Item{Name: name, Serial: serial, CategoryID: catID, BranchID: 1, Cost: cost,
			Status: status, Quantity: 1, ShowOnSite: false, CreatedAt: time.Now()}
		for an, v := range specs {
			if aid, ok := attrID[an]; ok {
				it.Values = append(it.Values, ItemAttributeValue{AttributeID: aid, Value: v})
			}
		}
		db.Omit("Values.Attribute").Create(&it)
		allSer[nsNorm(serial)] = true
	}
	adds, sales, rz := 0, 0, 0
	for _, a := range plan.Adds {
		if !allSer[nsNorm(a.Serial)] {
			mk(nbCat.ID, a.Name, a.Serial, a.Cost, a.Specs, "in_stock")
			adds++
		}
	}
	for _, r := range plan.Realiz {
		if !allSer[nsNorm(r.Serial)] {
			cat := r.Cat
			if cat == 0 {
				cat = nbCat.ID
			}
			mk(cat, r.Name, r.Serial, r.Cost, r.Specs, "consignment")
			rz++
		}
	}
	// SALES — Mərkəz item tap (seriya ilə), satışı yoxdursa yarat + «satıldı» et
	var merkez []Item
	db.Where("branch_id = ?", 1).Select("id, serial, cost").Find(&merkez)
	byser := map[string]Item{}
	for _, it := range merkez {
		if it.Serial != "" {
			byser[nsNorm(it.Serial)] = it
		}
	}
	for _, s := range plan.Sales {
		it, ok := byser[nsNorm(s.Serial)]
		if !ok {
			continue
		}
		var cnt int64
		db.Model(&Sale{}).Where("item_id = ?", it.ID).Count(&cnt)
		if cnt == 0 {
			cost := it.Cost
			if cost == 0 {
				cost = s.Cost
			}
			soldAt := time.Now()
			if t, e := time.Parse("2006-01-02", s.Date); e == nil {
				soldAt = t
			}
			db.Create(&Sale{ItemID: it.ID, SalePrice: s.Sale, Quantity: 1, Profit: s.Sale - cost,
				Channel: "cash", BranchID: 1, SoldAt: soldAt, Counted: true})
			sales++
		}
		db.Model(&Item{}).Where("id = ?", it.ID).Update("status", "sold")
	}
	setSetting("inc_sync_v1_done", "1")
	os.Remove("/data/inc_plan.json")
	log.Printf("inc sync tamam: adds=%d sales=%d realiz=%d", adds, sales, rz)
}
