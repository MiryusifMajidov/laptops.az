package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// BİRDƏFƏLİK düzəliş (2026-09-09):
//  1) Kredit üçün satış qeydi OLMAMALIDIR — səhvən yaradılmış kredit satışları silinir.
//     (Kredit tam ödəndikdə satışı mağaza özü yazacaq. credit_plans toxunulmur.)
//  2) Excel-də satılıb, DB-də satış qeydi olmayan son 10 günün satışları əlavə olunur.
//  3) Excel stokunda olub DB-də ümumiyyətlə olmayan məhsullar əlavə olunur.
//  4) Bir məhsulun səhv seriyası düzəldilir.
//
// ⚠️ MƏHSUL SİLİNMİR. Yalnız səhv satış qeydləri silinir.

type fix2Plan struct {
	DeleteCreditSales bool `json:"delete_credit_sales"`
	CreateSale        []struct {
		ItemID    uint    `json:"item_id"`
		SalePrice float64 `json:"sale_price"`
		Cost      float64 `json:"cost"`
		SoldAt    string  `json:"sold_at"`
	} `json:"create_sale"`
	AddItemAndSale []struct {
		Name       string  `json:"name"`
		Serial     string  `json:"serial"`
		CategoryID uint    `json:"category_id"`
		Cost       float64 `json:"cost"`
		SalePrice  float64 `json:"sale_price"`
		SoldAt     string  `json:"sold_at"`
	} `json:"add_item_and_sale"`
	AddItem []struct {
		Name       string  `json:"name"`
		Serial     string  `json:"serial"`
		BranchID   uint    `json:"branch_id"`
		CategoryID uint    `json:"category_id"`
		Cost       float64 `json:"cost"`
		CreatedAt  string  `json:"created_at"`
	} `json:"add_item"`
	SetSerial []struct {
		ItemID uint   `json:"item_id"`
		Serial string `json:"serial"`
	} `json:"set_serial"`
	DeleteSaleIDs []uint `json:"delete_sale_ids"` // səhvən yaradılmış satış sətirləri (MƏHSUL silinmir)
}

func f2date(s string) time.Time {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	return time.Now()
}

func applyFix2202609() { runFixPlan("fix2_2026_09", "/data/fix2_2026_09.json") }

// applyFix3202609 — eyni mexanizm, ikinci dalğa: seriyasız (aksesuar və s.) əksik satışlar.
func applyFix3202609() { runFixPlan("fix3_2026_09", "/data/fix3_2026_09.json") }

// applyFix4202609 — fix3-də ad uyğunlaşdırmasının səhvi ucbatından yaranmış dublikat satışları silir.
func applyFix4202609() { runFixPlan("fix4_2026_09", "/data/fix4_2026_09.json") }

// applyFix5202609 — son bir dublikat satışın təmizlənməsi.
func applyFix5202609() { runFixPlan("fix5_2026_09", "/data/fix5_2026_09.json") }

func runFixPlan(marker, path string) {
	if getSetting(marker+"_done") == "1" {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var p fix2Plan
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Printf("%s: plan oxunmadı: %v", marker, err)
		return
	}
	var nDel, nSale, nNewSold, nNewStock, nSer int

	// 1) kredit satışlarını sil — YALNIZ satış sətri.
	// Məhsulun statusuna toxunulmur: mal müştəridədir (stokda deyil),
	// gəlir qeydini kredit tam ödəndikdə mağaza özü yazacaq.
	if p.DeleteCreditSales {
		res := db.Where("channel = ?", "credit").Delete(&Sale{})
		nDel = int(res.RowsAffected)
		// kredit planlarındakı sale_id artıq mövcud deyil — boşalt
		db.Model(&CreditPlan{}).Where("sale_id IS NOT NULL").Update("sale_id", nil)
	}

	// 2) mövcud məhsul üçün əksik satış
	for _, m := range p.CreateSale {
		var it Item
		if db.First(&it, m.ItemID).Error != nil {
			continue
		}
		var have int64
		db.Model(&Sale{}).Where("item_id = ?", m.ItemID).Count(&have)
		if have > 0 {
			continue
		}
		cost := it.Cost
		if cost == 0 {
			cost = m.Cost
		}
		db.Create(&Sale{ItemID: it.ID, SalePrice: m.SalePrice, Quantity: 1,
			Profit: m.SalePrice - cost, Channel: "cash", BranchID: it.BranchID,
			SoldAt: f2date(m.SoldAt), Counted: true, Pending: false})
		db.Model(&it).Updates(map[string]any{"status": "sold", "show_on_site": false})
		nSale++
	}

	// 3) məhsul da yoxdursa: məhsul + satış
	for _, m := range p.AddItemAndSale {
		var exists int64
		if m.Serial != "" {
			db.Model(&Item{}).Where("lower(replace(serial,' ','')) = lower(replace(?,' ',''))", m.Serial).Count(&exists)
		}
		if exists > 0 {
			continue
		}
		it := Item{Name: m.Name, Serial: m.Serial, CategoryID: m.CategoryID, BranchID: 1,
			Cost: m.Cost, Price: m.SalePrice, Quantity: 1, Status: "sold",
			ShowOnSite: false, CreatedAt: f2date(m.SoldAt)}
		db.Create(&it)
		db.Create(&Sale{ItemID: it.ID, SalePrice: m.SalePrice, Quantity: 1,
			Profit: m.SalePrice - m.Cost, Channel: "cash", BranchID: 1,
			SoldAt: f2date(m.SoldAt), Counted: true, Pending: false})
		nNewSold++
	}

	// 4) əksik STOK məhsulları
	for _, m := range p.AddItem {
		var exists int64
		if m.Serial != "" {
			db.Model(&Item{}).Where("lower(replace(serial,' ','')) = lower(replace(?,' ',''))", m.Serial).Count(&exists)
		}
		if exists > 0 {
			continue
		}
		db.Create(&Item{Name: m.Name, Serial: m.Serial, CategoryID: m.CategoryID,
			BranchID: m.BranchID, Cost: m.Cost, Quantity: 1, Status: "in_stock",
			ShowOnSite: false, CreatedAt: f2date(m.CreatedAt)})
		nNewStock++
	}

	// 4b) səhvən yaradılmış dublikat satış sətirlərini sil (yalnız SATIŞ, məhsul qalır)
	if len(p.DeleteSaleIDs) > 0 {
		res := db.Where("id IN ?", p.DeleteSaleIDs).Delete(&Sale{})
		nDel += int(res.RowsAffected)
	}

	// 5) seriya düzəlişi
	for _, m := range p.SetSerial {
		var it Item
		if db.First(&it, m.ItemID).Error == nil && it.Serial != m.Serial {
			db.Model(&it).Update("serial", m.Serial)
			nSer++
		}
	}

	log.Printf("%s: silinen_kredit_satisi=%d yeni_satis=%d yeni_satilmis_mehsul=%d yeni_stok=%d seriya=%d",
		marker, nDel, nSale, nNewSold, nNewStock, nSer)
	setSetting(marker+"_done", "1")
	os.Remove(path)
}
