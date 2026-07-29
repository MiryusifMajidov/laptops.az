package main

// PRIXOD Excel-dən çıxarılmış JSON-u real data kimi DB-yə yazır.
// İki yol:
//   1) CLI:  go run . -import <import_data.json>
//   2) HTTP: POST /api/import  (admin token) — atomik (transaction), saytı dayandırmadan.
// DİQQƏT: köhnə əməliyyat datalarını (məhsul/satış/realizasiya/kredit...) SİLİR;
//   istifadəçilər, filiallar, mail şablonları, audit jurnalı, sayt tənzimləmələri SAXLANILIR.

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type impItem struct {
	Name      string            `json:"name"`
	Serial    string            `json:"serial"`
	Cost      float64           `json:"cost"`
	Price     float64           `json:"price"`
	Category  string            `json:"category"`
	Branch    string            `json:"branch"`
	Attrs     map[string]string `json:"attrs"`
	CreatedAt *string           `json:"created_at"`
}

type impConsignment struct {
	Store      string  `json:"store"`
	GivenAt    *string `json:"given_at"`
	GivenPrice float64 `json:"given_price"`
	Item       impItem `json:"item"`
}

type impCredit struct {
	Customer string  `json:"customer"`
	Phone    string  `json:"phone"`
	Total    float64 `json:"total"`
	Paid     float64 `json:"paid"`
	SoldAt   *string `json:"sold_at"`
	Note     string  `json:"note"`
	Item     impItem `json:"item"`
}

type impData struct {
	Items        []impItem        `json:"items"`
	Consignments []impConsignment `json:"consignments"`
	Credits      []impCredit      `json:"credits"`
}

func impDate(s *string, fallback time.Time) time.Time {
	if s == nil || *s == "" {
		return fallback
	}
	d, err := time.ParseInLocation("2006-01-02", *s, time.Local)
	if err != nil {
		return fallback
	}
	return d.Add(12 * time.Hour)
}

// CLI giriş nöqtəsi
func runImport(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("import faylı açılmadı: %v", err)
	}
	var data impData
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Fatalf("import JSON oxunmadı: %v", err)
	}
	log.Println(importData(db, data))
}

// HTTP: POST /api/import  (yalnız admin) — body = impData JSON
func importHandler(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromReq(r)
	if !ok || u.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "yalnız admin"})
		return
	}
	var data impData
	if err := decodeBody(r, &data); err != nil {
		writeJSON(w, 400, map[string]string{"error": "JSON oxunmadı: " + err.Error()})
		return
	}
	if len(data.Items) == 0 && len(data.Consignments) == 0 {
		writeJSON(w, 400, map[string]string{"error": "boş data"})
		return
	}
	var summary string
	defer func() {
		if rec := recover(); rec != nil {
			writeJSON(w, 500, map[string]string{"error": fmt.Sprintf("import xətası (geri qaytarıldı): %v", rec)})
		}
	}()
	if err := db.Transaction(func(tx *gorm.DB) error {
		summary = importData(tx, data)
		return nil
	}); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "summary": summary})
}

// importData — bütün silmə + yazma məntiqi. Xətada panic edir (transaction geri qaytarır).
func importData(g *gorm.DB, data impData) string {
	// ---- 1. köhnə əməliyyat datalarını sil ----
	tables := []string{
		"item_attribute_values", "sales", "credit_payments", "credit_plans",
		"consignments", "online_orders", "items", "customers",
		"expenses", "daily_closes", "supply_batches", "item_translations",
	}
	for _, t := range tables {
		if err := g.Exec("DELETE FROM " + t).Error; err != nil {
			panic(fmt.Sprintf("silinmə xətası (%s): %v", t, err))
		}
	}
	g.Exec("DELETE FROM sqlite_sequence WHERE name IN ('items','sales','customers','credit_plans','credit_payments','consignments','online_orders','expenses','daily_closes','supply_batches','item_attribute_values')")

	// ---- 2. lüğətlər (filial / kateqoriya / xüsusiyyət) ----
	branchID := func(name string) uint {
		if strings.TrimSpace(name) == "" {
			name = "Mərkəz (Mağaza)"
		}
		var b Branch
		if g.Where("name = ?", name).First(&b).Error != nil {
			b = Branch{Name: name}
			g.Create(&b)
		}
		return b.ID
	}
	attrID := func(name string) uint {
		var a Attribute
		if g.Where("name = ?", name).First(&a).Error != nil {
			a = Attribute{Name: name}
			g.Create(&a)
		}
		return a.ID
	}
	catID := func(name string) uint {
		if strings.TrimSpace(name) == "" {
			name = "Aksesuar"
		}
		var c Category
		if g.Where("name = ?", name).First(&c).Error != nil {
			c = Category{Name: name}
			g.Create(&c)
		}
		return c.ID
	}

	mainBranch := branchID("Mərkəz (Mağaza)")

	createItem := func(ii impItem, branch uint, status string, createdAt time.Time) uint {
		it := Item{
			Name: strings.TrimSpace(ii.Name), Serial: ii.Serial,
			CategoryID: catID(ii.Category), BranchID: branch,
			Cost: ii.Cost, Price: ii.Price,
			Status: status, ShowOnSite: false, // sayt görünüşünü sahib özü açır
			CreatedAt: createdAt,
		}
		if err := g.Create(&it).Error; err != nil {
			panic(fmt.Sprintf("item yazılmadı (%s): %v", ii.Name, err))
		}
		vals := make([]ItemAttributeValue, 0, len(ii.Attrs))
		for an, av := range ii.Attrs {
			if strings.TrimSpace(an) == "" || strings.TrimSpace(av) == "" {
				continue
			}
			vals = append(vals, ItemAttributeValue{ItemID: it.ID, AttributeID: attrID(an), Value: av})
		}
		if len(vals) > 0 {
			g.Create(&vals)
		}
		return it.ID
	}

	// ---- 3. stok ----
	for _, ii := range data.Items {
		createItem(ii, branchID(ii.Branch), "in_stock", impDate(ii.CreatedAt, time.Now()))
	}

	// ---- 4. realizasiya ----
	for _, c := range data.Consignments {
		given := impDate(c.GivenAt, time.Now().AddDate(0, 0, -30))
		id := createItem(c.Item, mainBranch, "consignment", given)
		g.Create(&Consignment{
			StoreName: c.Store, ItemID: &id,
			ItemName: c.Item.Name, Serial: c.Item.Serial,
			GivenAt: given, Cost: c.Item.Cost, GivenPrice: c.GivenPrice,
			Status: "out", Debt: c.Item.Cost,
		})
	}

	// ---- 5. kredit borcları ----
	undatedFallback := time.Date(time.Now().Year(), 1, 1, 12, 0, 0, 0, time.Local)
	for _, c := range data.Credits {
		soldAt := impDate(c.SoldAt, undatedFallback)
		cust := Customer{Name: c.Customer, Phone: c.Phone, CreatedAt: soldAt}
		g.Create(&cust)
		itemID := createItem(c.Item, mainBranch, "sold", soldAt)
		sale := Sale{
			ItemID: itemID, SalePrice: c.Total, Profit: c.Total - c.Item.Cost,
			Channel: "credit", CustomerID: &cust.ID, BranchID: mainBranch, SoldAt: soldAt,
			Counted: c.Paid >= c.Total,
		}
		g.Create(&sale)
		nextDue := soldAt.AddDate(0, 1, 0)
		status := "ontime"
		if c.Paid >= c.Total {
			status = "paid"
		} else if nextDue.Before(time.Now()) {
			status = "overdue"
		}
		plan := CreditPlan{
			CustomerID: cust.ID, ItemName: c.Item.Name,
			Total: c.Total, Paid: c.Paid, NextDue: nextDue, Status: status,
			SaleID: &sale.ID,
		}
		g.Create(&plan)
		if c.Paid > 0 {
			g.Create(&CreditPayment{CreditPlanID: plan.ID, Amount: c.Paid, CreatedAt: soldAt})
		}
	}

	// ---- 6. taksonomiyanı real dəyərlərdən yenidən qur ----
	rebuildTaxonomy(g)

	var nItems, nStock, nCons, nAttr, nOpt int64
	g.Model(&Item{}).Count(&nItems)
	g.Model(&Item{}).Where("status = ?", "in_stock").Count(&nStock)
	g.Model(&Consignment{}).Count(&nCons)
	g.Model(&Attribute{}).Count(&nAttr)
	g.Model(&AttributeOption{}).Count(&nOpt)
	return fmt.Sprintf("cihaz: %d (stok: %d), realizasiya: %d, xüsusiyyət: %d, option: %d",
		nItems, nStock, nCons, nAttr, nOpt)
}

// rebuildTaxonomy — kateqoriya↔xüsusiyyət bağlantılarını və option-ları REAL cihaz
// dəyərlərindən yenidən qurur; istifadəsiz (fake) xüsusiyyətləri silir.
func rebuildTaxonomy(g *gorm.DB) {
	type triple struct {
		CategoryID  uint
		AttributeID uint
		Value       string
	}
	var rows []triple
	g.Raw(`SELECT it.category_id AS category_id, iav.attribute_id AS attribute_id, iav.value AS value
	        FROM item_attribute_values iav JOIN items it ON it.id = iav.item_id`).Scan(&rows)

	attrName := map[uint]string{}
	{
		type an struct {
			ID   uint
			Name string
		}
		var as []an
		g.Raw(`SELECT id, name FROM attributes`).Scan(&as)
		for _, a := range as {
			attrName[a.ID] = a.Name
		}
	}

	catAttrs := map[uint]map[uint]bool{}
	attrVals := map[uint]map[string]bool{}
	used := map[uint]bool{}
	for _, r := range rows {
		if strings.TrimSpace(r.Value) == "" {
			continue
		}
		used[r.AttributeID] = true
		if catAttrs[r.CategoryID] == nil {
			catAttrs[r.CategoryID] = map[uint]bool{}
		}
		catAttrs[r.CategoryID][r.AttributeID] = true
		if attrVals[r.AttributeID] == nil {
			attrVals[r.AttributeID] = map[string]bool{}
		}
		attrVals[r.AttributeID][r.Value] = true
	}

	g.Exec("DELETE FROM attribute_options")
	g.Exec("DELETE FROM category_attributes")
	g.Exec("DELETE FROM sqlite_sequence WHERE name = 'attribute_options'")

	ordinal := map[string]bool{"RAM": true, "SSD": true, "Yaddaş": true, "Ekran": true, "Tezlik": true}

	for aid, vals := range attrVals {
		list := make([]string, 0, len(vals))
		for v := range vals {
			list = append(list, v)
		}
		if ordinal[attrName[aid]] {
			sort.Slice(list, func(i, j int) bool {
				ni, oki := numKey(list[i])
				nj, okj := numKey(list[j])
				if oki && okj && ni != nj {
					return ni < nj
				}
				return list[i] < list[j]
			})
		} else {
			sort.Strings(list)
		}
		for _, v := range list {
			g.Create(&AttributeOption{AttributeID: aid, Value: v})
		}
	}

	attrOrder := map[string]int{
		"Marka": 0, "Prosessor": 1, "RAM": 2, "SSD": 3, "Ekran kartı": 4,
		"Ekran": 5, "Yaddaş": 6, "Rəng": 7, "Ölçü": 8, "Tezlik": 9,
		"Növ": 10, "Vəziyyət": 11,
	}
	for cid, attrs := range catAttrs {
		ids := make([]uint, 0, len(attrs))
		for aid := range attrs {
			ids = append(ids, aid)
		}
		sort.Slice(ids, func(i, j int) bool {
			oi, ok := attrOrder[attrName[ids[i]]]
			if !ok {
				oi = 99
			}
			oj, ok := attrOrder[attrName[ids[j]]]
			if !ok {
				oj = 99
			}
			return oi < oj
		})
		for _, aid := range ids {
			g.Exec("INSERT INTO category_attributes (category_id, attribute_id) VALUES (?, ?)", cid, aid)
		}
	}

	for aid, name := range attrName {
		if !used[aid] {
			g.Exec("DELETE FROM attributes WHERE id = ?", aid)
			log.Printf("istifadəsiz xüsusiyyət silindi: %s", name)
		}
	}
}

var reNum = regexp.MustCompile(`\d+(?:\.\d+)?`)

// numKey — "512 GB"→512, "1 TB"→1024, "16.0\""→16, "120Hz"→120
func numKey(val string) (float64, bool) {
	m := reNum.FindString(val)
	if m == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return 0, false
	}
	if strings.Contains(val, "TB") {
		f *= 1024
	}
	return f, true
}
