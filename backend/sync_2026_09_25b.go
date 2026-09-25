package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"
)

// BİRDƏFƏLİK (2026-09-25, 2-ci dalğa): sahibkarın göstərişi — «ən doğru olan Exceldir,
// dublikatları həll et, ayın tarixi tam uyğun olmalıdır».
// Burada DB Excel-ə uyğunlaşdırılır: satış tarixi/qiyməti, maya dəyəri və dublikat kartların
// birləşdirilməsi. Dublikat halında SATIŞ əsl karta köçürülür, boş kart arxivə göndərilir
// (soft-delete — «Silinmiş məhsullar» səhifəsindən bərpa oluna bilər). Şəkillər itmir.
//
// ⚠️ TOXUNULMAYAN CƏDVƏLLƏR: credit_plans, credit_payments, consignments.
//   • Kredit: «kreditə verilənlər üçün satış qeydi yaradılmır, kredit tam ödəndikdə özümüz edəcəyik».
//   • Realizasiya: statusu «sold_paid» etmək konsiqnatorun borcunu sıfırlayır (= «pulu verdik»),
//     bu, sahibkarın qərarıdır; üstəlik updateConsignment hər göndərişdə yeni satış yarada bilər.
// ⚠️ Məhsul fiziki olaraq silinmir (yalnız soft-delete).
// ⚠️ Hər əməliyyat öz tranzaksiyasındadır: yarımçıq vəziyyət qalmır (yenidən işə düşəndə
//    qoruyucular əməliyyatı təkrarlamır, ona görə yarımçıqlıq birdəfəlik itki olardı).

type truthPlan struct {
	SetSalePrice []struct {
		SaleID    uint    `json:"sale_id"`
		SalePrice float64 `json:"sale_price"`
	} `json:"set_sale_price"`
	SetSaleDate []struct {
		SaleID uint   `json:"sale_id"`
		SoldAt string `json:"sold_at"`
	} `json:"set_sale_date"`
	SetItemCost []struct {
		ItemID uint    `json:"item_id"`
		Cost   float64 `json:"cost"`
		SaleID uint    `json:"sale_id"` // həmin satışın mənfəəti yenidən hesablanır
	} `json:"set_item_cost"`
	// Excel-də seriya var, DB kartı seriyasız yazılıb → cihaz seriya ilə izlənə bilsin
	SetItemSerial []struct {
		ItemID uint   `json:"item_id"`
		Serial string `json:"serial"`
	} `json:"set_item_serial"`
	AddItemAndSale []struct {
		Name       string  `json:"name"`
		Serial     string  `json:"serial"`
		BranchID   uint    `json:"branch_id"`
		CategoryID uint    `json:"category_id"`
		Cost       float64 `json:"cost"`
		SalePrice  float64 `json:"sale_price"`
		SoldAt     string  `json:"sold_at"`
	} `json:"add_item_and_sale"`
	// Excel-də olmayan satış silinir. Lazım gələrsə mal stoka qaytarılır (Excel hansı filialı göstərirsə),
	// satışın bağlı olduğu təkrar kart isə arxivə göndərilir.
	DeleteSale []struct {
		SaleID        uint `json:"sale_id"`
		RestoreItemID uint `json:"restore_item_id"`
		RestoreBranch uint `json:"restore_branch"`
		ArchiveItemID uint `json:"archive_item_id"`
	} `json:"delete_sale"`
	// Eyni cihazın iki kartı: satış «boş» kartdadır, əsl kart (şəkilli) ayrıca durur.
	MergeDuplicate []struct {
		KeepItemID uint    `json:"keep_item_id"`
		JunkItemID uint    `json:"junk_item_id"`
		SaleID     uint    `json:"sale_id"`
		Serial     string  `json:"serial"`     // Excel-dəki seriya (boşdursa kartınkı saxlanılır)
		SalePrice  float64 `json:"sale_price"` // Excel-dəki qiymət
		SoldAt     string  `json:"sold_at"`    // Excel-dəki tarix
	} `json:"merge_duplicate"`
}

// saleProfit — satışın mənfəətini malın mayasına görə yenidən hesablayır.
func saleProfit(s *Sale, cost float64) float64 {
	q := s.Quantity
	if q < 1 {
		q = 1
	}
	return s.SalePrice - cost*float64(q)
}

func applySync20260925b() {
	const marker = "sync_2026_09_25b"
	const path = "/data/sync_2026_09_25b.json"
	if getSetting(marker+"_done") == "1" {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		// səssiz qayıtma olmasın: «plan yüklənməyib»i «uğurla bitdi»dən ayırmaq üçün jurnala yazılır
		log.Printf("%s: plan faylı oxunmadı (%s): %v", marker, path, err)
		return
	}
	var p truthPlan
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Printf("%s: plan oxunmadı: %v", marker, err)
		return
	}
	skip := func(what string, id any) { log.Printf("%s: KEÇİLDİ %s %v", marker, what, id) }
	fail := func(what string, id any, err error) { log.Printf("%s: XƏTA %s %v: %v", marker, what, id, err) }
	var nPrice, nDate, nCost, nNew, nMerge, nDel, nSerial int

	// 0) Excel-də olmayan satışların silinməsi (mal stoka qayıdır / təkrar kart arxivə)
	// Satış qeydində soft-delete YOXDUR → əvvəl bərpa/arxiv, ƏN SONDA silmə, hamısı bir tranzaksiyada.
	for _, m := range p.DeleteSale {
		var s Sale
		if db.First(&s, m.SaleID).Error != nil {
			skip("delete_sale", m.SaleID)
			continue
		}
		// silinən pul qeydi jurnala düşsün — lazım olsa əl ilə bərpa edilə bilsin
		log.Printf("%s: SİLİNİR satış#%d mal#%d %.0f₼ %s mənfəət %.0f",
			marker, s.ID, s.ItemID, s.SalePrice, s.SoldAt.Format("2006-01-02"), s.Profit)
		err := db.Transaction(func(tx *gorm.DB) error {
			if m.RestoreItemID != 0 {
				upd := map[string]any{"status": "in_stock", "quantity": 1}
				if m.RestoreBranch != 0 {
					upd["branch_id"] = m.RestoreBranch
				}
				if e := tx.Model(&Item{}).Where("id = ?", m.RestoreItemID).Updates(upd).Error; e != nil {
					return e
				}
			}
			if m.ArchiveItemID != 0 { // təkrar kart — bu satışdan başqa satışı yoxdursa arxivə
				var left int64
				tx.Model(&Sale{}).Where("item_id = ? AND id <> ?", m.ArchiveItemID, s.ID).Count(&left)
				if left == 0 {
					if e := tx.Delete(&Item{}, m.ArchiveItemID).Error; e != nil {
						return e
					}
				} else {
					skip("delete_sale(arxiv: başqa satışı var)", m.ArchiveItemID)
				}
			}
			return tx.Delete(&Sale{}, m.SaleID).Error
		})
		if err != nil {
			fail("delete_sale", m.SaleID, err)
			continue
		}
		nDel++
	}

	// 1) satış qiyməti Excel-ə uyğunlaşdırılır (mənfəət də yenidən hesablanır)
	for _, m := range p.SetSalePrice {
		var s Sale
		if db.First(&s, m.SaleID).Error != nil {
			skip("set_sale_price", m.SaleID)
			continue
		}
		var it Item
		db.Unscoped().First(&it, s.ItemID)
		s.SalePrice = m.SalePrice
		db.Model(&s).Updates(map[string]any{"sale_price": m.SalePrice, "profit": saleProfit(&s, it.Cost)})
		nPrice++
	}

	// 2) satış tarixi Excel-ə uyğunlaşdırılır
	for _, m := range p.SetSaleDate {
		var s Sale
		if db.First(&s, m.SaleID).Error != nil {
			skip("set_sale_date", m.SaleID)
			continue
		}
		db.Model(&s).Update("sold_at", f2date(m.SoldAt))
		nDate++
	}

	// 3) maya dəyəri Excel-ə uyğunlaşdırılır + həmin satışın mənfəəti
	for _, m := range p.SetItemCost {
		var it Item
		if db.Unscoped().First(&it, m.ItemID).Error != nil {
			skip("set_item_cost", m.ItemID)
			continue
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			// Unscoped — arxivdəki kartın da mayası yazılsın (adi Model() soft-delete filtri qoyur)
			if e := tx.Unscoped().Model(&Item{}).Where("id = ?", it.ID).Update("cost", m.Cost).Error; e != nil {
				return e
			}
			var s Sale
			if m.SaleID != 0 && tx.First(&s, m.SaleID).Error == nil && s.ItemID == it.ID {
				return tx.Model(&s).Update("profit", saleProfit(&s, m.Cost)).Error
			}
			return nil
		})
		if err != nil {
			fail("set_item_cost", m.ItemID, err)
			continue
		}
		nCost++
	}

	// 3b) Excel-dəki seriya kartın üzərinə yazılır (yalnız kart seriyasızdırsa)
	for _, m := range p.SetItemSerial {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || strings.TrimSpace(it.Serial) != "" {
			skip("set_item_serial", m.ItemID)
			continue
		}
		if serialTaken(m.Serial) {
			skip("set_item_serial(seriya başqa maldadır)", m.Serial)
			continue
		}
		db.Model(&it).Update("serial", m.Serial)
		nSerial++
	}

	// 4) Excel-də satılıb, DB-də heç bir qeydi olmayan mal
	for _, m := range p.AddItemAndSale {
		if serialTaken(m.Serial) {
			skip("add_item_and_sale", m.Serial)
			continue
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			it := Item{Name: m.Name, Serial: m.Serial, BranchID: m.BranchID, CategoryID: m.CategoryID,
				Cost: m.Cost, Price: m.SalePrice, Quantity: 1, Status: "in_stock", ShowOnSite: false,
				CreatedAt: f2date(m.SoldAt)}
			if e := tx.Create(&it).Error; e != nil {
				return e
			}
			if e := tx.Create(&Sale{ItemID: it.ID, SalePrice: m.SalePrice, Quantity: 1,
				Profit: m.SalePrice - m.Cost, Channel: "cash", BranchID: m.BranchID,
				SoldAt: f2date(m.SoldAt), Counted: true}).Error; e != nil {
				return e
			}
			return tx.Model(&it).Updates(map[string]any{"status": "sold", "quantity": 0, "show_on_site": false}).Error
		})
		if err != nil {
			fail("add_item_and_sale", m.Serial, err)
			continue
		}
		nNew++
	}

	// 5) DUBLİKAT BİRLƏŞDİRMƏSİ
	for _, m := range p.MergeDuplicate {
		var keep, junk Item
		var s Sale
		if db.First(&keep, m.KeepItemID).Error != nil || db.First(&junk, m.JunkItemID).Error != nil ||
			db.First(&s, m.SaleID).Error != nil {
			skip("merge_duplicate", m.JunkItemID)
			continue
		}
		// gözlənilən quruluş: satış BOŞ kartdadır, əsl kartın satışı yoxdur
		if s.ItemID != junk.ID {
			skip("merge_duplicate(satış başqa maldadır)", m.SaleID)
			continue
		}
		var keepSales int64
		db.Model(&Sale{}).Where("item_id = ?", keep.ID).Count(&keepSales)
		if keepSales > 0 {
			skip("merge_duplicate(əsl kartın artıq satışı var)", m.KeepItemID)
			continue
		}
		// seriya YALNIZ kart seriyasızdırsa, ya da eyni seriyadırsa yazılır —
		// fiziki identifikatoru üstündən yazmaq olmaz (Excel-də çap səhvi ola bilər)
		serial := keep.Serial
		if m.Serial != "" {
			if strings.TrimSpace(keep.Serial) == "" || normSerial(keep.Serial) == normSerial(m.Serial) {
				serial = m.Serial
			} else {
				skip("merge_duplicate(seriya fərqlidir, DB-dəki saxlanıldı)", keep.Serial)
			}
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			// satışı əsl karta köçür — pul qeydi itmir, yalnız sahibi dəyişir
			upd := map[string]any{"item_id": keep.ID, "branch_id": keep.BranchID}
			if m.SalePrice > 0 {
				upd["sale_price"] = m.SalePrice
				s.SalePrice = m.SalePrice
			}
			upd["profit"] = saleProfit(&s, keep.Cost)
			if m.SoldAt != "" { // «ayın tarixi tam uyğun olmalıdır»
				upd["sold_at"] = f2date(m.SoldAt)
			}
			if e := tx.Model(&s).Updates(upd).Error; e != nil {
				return e
			}
			if e := tx.Model(&keep).Updates(map[string]any{"status": "sold", "quantity": 0,
				"show_on_site": false, "serial": serial}).Error; e != nil {
				return e
			}
			// boş kart arxivə (soft-delete) — «Silinmiş məhsullar»dan bərpa oluna bilər
			return tx.Delete(&junk).Error
		})
		if err != nil {
			fail("merge_duplicate", m.JunkItemID, err)
			continue
		}
		nMerge++
	}

	log.Printf("%s: silinen_satis=%d qiymet=%d tarix=%d maya=%d seriya=%d yeni_satis=%d dublikat_birlesdi=%d",
		marker, nDel, nPrice, nDate, nCost, nSerial, nNew, nMerge)
	setSetting(marker+"_done", "1")
	os.Remove(path)
}
