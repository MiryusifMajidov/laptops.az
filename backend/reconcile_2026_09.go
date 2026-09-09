package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// BİRDƏFƏLİK: 2026-09 Excel ↔ DB uzlaşdırma auditinin nəticələrini tətbiq edir.
// Plan /data/reconcile_2026_09.json faylından oxunur (biznes datası repo-da saxlanmır),
// tətbiqdən sonra fayl silinir və marker qoyulur — təkrar işləmir.
//
// Plan əməliyyatları:
//   set_branch          — məhsulu düzgün filiala keçir (transfer qeyd olunmayıb)
//   set_status          — status düzəlişi (məs. realizasiyaya verilib, hələ in_stock görünür)
//   set_cost            — maya dəyəri boş/səhv (mənfəət hesabatını pozur)
//   create_sale         — Excel-də satılıb, DB-də satış qeydi yoxdur (fantom stok)
//   close_consignment   — realizasiya sətri bağlanmayıb (saxta açıq borc)
//   create_consignment  — realizasiyaya verilib, qeyd yaradılmayıb
//   add_item            — Excel-də var, DB-də ümumiyyətlə yoxdur

type recPlan struct {
	SetBranch []struct {
		ItemID   uint `json:"item_id"`
		BranchID uint `json:"branch_id"`
	} `json:"set_branch"`
	SetStatus []struct {
		ItemID uint   `json:"item_id"`
		Status string `json:"status"`
	} `json:"set_status"`
	SetCost []struct {
		ItemID uint    `json:"item_id"`
		Cost   float64 `json:"cost"`
	} `json:"set_cost"`
	CreateSale []struct {
		ItemID    uint    `json:"item_id"`
		SalePrice float64 `json:"sale_price"`
		SoldAt    string  `json:"sold_at"` // YYYY-MM-DD
		Channel   string  `json:"channel"`
		BranchID  uint    `json:"branch_id"`
	} `json:"create_sale"`
	CloseConsignment []struct {
		ID     uint   `json:"id"`
		Status string `json:"status"`
		SaleID *uint  `json:"sale_id"`
	} `json:"close_consignment"`
	CreateConsignment []struct {
		ItemID     uint    `json:"item_id"`
		StoreName  string  `json:"store_name"`
		GivenAt    string  `json:"given_at"`
		GivenPrice float64 `json:"given_price"`
	} `json:"create_consignment"`
	AddItem []struct {
		Name       string  `json:"name"`
		Serial     string  `json:"serial"`
		BranchID   uint    `json:"branch_id"`
		CategoryID uint    `json:"category_id"`
		Cost       float64 `json:"cost"`
		Status     string  `json:"status"`
	} `json:"add_item"`
}

func recDate(s string) time.Time {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	return time.Now()
}

func applyReconcile202609() {
	if getSetting("reconcile_2026_09_done") == "1" {
		return
	}
	raw, err := os.ReadFile("/data/reconcile_2026_09.json")
	if err != nil {
		return // plan yoxdur — sükutla keç
	}
	var p recPlan
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Printf("reconcile_2026_09: plan oxunmadı: %v", err)
		return
	}
	var nBranch, nStatus, nCost, nSale, nCloseC, nNewC, nAdd int

	// 1) filial düzəlişi
	for _, m := range p.SetBranch {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.BranchID == m.BranchID {
			continue
		}
		db.Model(&it).Update("branch_id", m.BranchID)
		nBranch++
	}

	// 2) status düzəlişi (realizasiya və s.) — saytdan da çıxarılır
	for _, m := range p.SetStatus {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.Status == m.Status {
			continue
		}
		upd := map[string]any{"status": m.Status}
		if m.Status != "in_stock" {
			upd["show_on_site"] = false
		}
		db.Model(&it).Updates(upd)
		nStatus++
	}

	// 3) maya dəyəri
	for _, m := range p.SetCost {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.Cost == m.Cost {
			continue
		}
		db.Model(&it).Update("cost", m.Cost)
		nCost++
	}

	// 4) əksik satışlar — məhsul "satıldı" olur, saytdan çıxır
	for _, m := range p.CreateSale {
		var it Item
		if db.First(&it, m.ItemID).Error != nil {
			continue
		}
		var exists int64
		db.Model(&Sale{}).Where("item_id = ?", m.ItemID).Count(&exists)
		if exists > 0 {
			continue // artıq satış qeydi var — idempotent
		}
		br := m.BranchID
		if br == 0 {
			br = it.BranchID
		}
		ch := m.Channel
		if ch == "" {
			ch = "cash"
		}
		db.Create(&Sale{
			ItemID: m.ItemID, SalePrice: m.SalePrice, Quantity: 1,
			Profit: m.SalePrice - it.Cost, Channel: ch, BranchID: br,
			SoldAt: recDate(m.SoldAt), Counted: true, Pending: false,
		})
		db.Model(&it).Updates(map[string]any{"status": "sold", "show_on_site": false})
		nSale++
	}

	// 5) açıq qalmış realizasiya sətirlərini bağla
	for _, m := range p.CloseConsignment {
		var cs Consignment
		if db.First(&cs, m.ID).Error != nil || cs.Status == m.Status {
			continue
		}
		upd := map[string]any{"status": m.Status, "debt": 0}
		if m.SaleID != nil {
			upd["sale_id"] = *m.SaleID
		}
		db.Model(&cs).Updates(upd)
		nCloseC++
	}

	// 6) əksik realizasiya qeydləri
	for _, m := range p.CreateConsignment {
		var it Item
		if db.First(&it, m.ItemID).Error != nil {
			continue
		}
		var exists int64
		db.Model(&Consignment{}).Where("item_id = ? AND status = ?", m.ItemID, "out").Count(&exists)
		if exists > 0 {
			continue
		}
		db.Create(&Consignment{
			StoreName: m.StoreName, ItemID: &m.ItemID, ItemName: it.Name, Serial: it.Serial,
			GivenAt: recDate(m.GivenAt), Cost: it.Cost, GivenPrice: m.GivenPrice,
			Status: "out", Debt: m.GivenPrice,
		})
		db.Model(&it).Updates(map[string]any{"status": "consignment", "show_on_site": false})
		nNewC++
	}

	// 7) DB-də ümumiyyətlə olmayan məhsullar
	for _, m := range p.AddItem {
		if m.Serial != "" {
			var exists int64
			db.Model(&Item{}).Where("lower(replace(serial,' ','')) = lower(?)", m.Serial).Count(&exists)
			if exists > 0 {
				continue
			}
		}
		st := m.Status
		if st == "" {
			st = "in_stock"
		}
		db.Create(&Item{
			Name: m.Name, Serial: m.Serial, BranchID: m.BranchID, CategoryID: m.CategoryID,
			Cost: m.Cost, Quantity: 1, Status: st, ShowOnSite: false, CreatedAt: time.Now(),
		})
		nAdd++
	}

	log.Printf("reconcile_2026_09: filial=%d status=%d maya=%d satış=%d realiz.bağlandı=%d realiz.yeni=%d yeni məhsul=%d",
		nBranch, nStatus, nCost, nSale, nCloseC, nNewC, nAdd)
	setSetting("reconcile_2026_09_done", "1")
	os.Remove("/data/reconcile_2026_09.json")
}
