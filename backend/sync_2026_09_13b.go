package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"
)

// BİRDƏFƏLİK (2026-09-13, 2-ci dalğa): Excel ↔ DB arasında ƏVVƏLDƏN qalmış fərqlərin düzəlişi —
// Excel-də satılıb DB-də stokda qalan mallar, filial səhvləri, seriyası yerini dəyişmiş/yazı səhvi olan
// qeydlər, iyun importundan açıq qalmış realizasiyalar, DB-də olmayan stok malları.
// Plan /data/sync_2026_09_13b.json-dan oxunur; hər əməliyyat gözlənilən vəziyyəti yoxlayır
// (təkrar işləsə heç nəyi ikiqat etmir), sonda marker qoyulur və fayl silinir.
//
// ⚠️ MƏHSUL SİLİNMİR. Kreditə toxunulmur. Satış/realizasiya məntiqi tətbiqin handler-ləri ilə eynidir.

type oldFixPlan struct {
	SetBranch []struct {
		ItemID   uint `json:"item_id"`
		BranchID uint `json:"branch_id"`
	} `json:"set_branch"`
	// seriya düzəlişi — yalnız hazırkı seriya From-a bərabərdirsə (dəyiş-dəyiş də təhlükəsiz olur)
	SetSerial []struct {
		ItemID uint   `json:"item_id"`
		From   string `json:"from"`
		To     string `json:"to"`
	} `json:"set_serial"`
	SetCost []struct {
		ItemID uint    `json:"item_id"`
		Cost   float64 `json:"cost"`
	} `json:"set_cost"`
	// satılmış malın maya dəyəri boş/səhv — həm mal, həm satışın mənfəəti düzəlir
	FixSoldCost []struct {
		ItemID uint    `json:"item_id"`
		SaleID uint    `json:"sale_id"`
		Cost   float64 `json:"cost"`
	} `json:"fix_sold_cost"`
	AddStock []struct {
		Name       string  `json:"name"`
		Serial     string  `json:"serial"`
		BranchID   uint    `json:"branch_id"`
		CategoryID uint    `json:"category_id"`
		Cost       float64 `json:"cost"`
		CreatedAt  string  `json:"created_at"`
	} `json:"add_stock"`
	SellItem []struct {
		ItemID    uint    `json:"item_id"`
		SalePrice float64 `json:"sale_price"`
		SoldAt    string  `json:"sold_at"`
	} `json:"sell_item"`
	CloseConsignmentSold []struct {
		ConsignmentID uint    `json:"consignment_id"`
		SalePrice     float64 `json:"sale_price"`
		SoldAt        string  `json:"sold_at"`
	} `json:"close_consignment_sold"`
	// mal artıq satış qeydi ilə satılıb, amma realizasiya sətri açıq qalıb → mövcud satışa bağla
	CloseConsignmentExistingSale []struct {
		ConsignmentID uint `json:"consignment_id"`
		SaleID        uint `json:"sale_id"`
	} `json:"close_consignment_existing_sale"`
	// statusu «realizasiyada» olan, amma realizasiya sətri olmayan mal → açıq sətir yarat
	OpenConsignmentRow []struct {
		ItemID     uint    `json:"item_id"`
		StoreName  string  `json:"store_name"`
		GivenAt    string  `json:"given_at"`
		GivenPrice float64 `json:"given_price"`
	} `json:"open_consignment_row"`
	// statusu «realizasiyada», sətri və satışı olmayan, Excel-də isə satılmış mal → sətir (sold_paid) + satış
	SellConsignmentItem []struct {
		ItemID     uint    `json:"item_id"`
		StoreName  string  `json:"store_name"`
		GivenAt    string  `json:"given_at"`
		GivenPrice float64 `json:"given_price"`
		SalePrice  float64 `json:"sale_price"`
		SoldAt     string  `json:"sold_at"`
	} `json:"sell_consignment_item"`
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]`)

func normSerial(s string) string { return nonAlnum.ReplaceAllString(strings.ToLower(s), "") }

func applySync20260913b() {
	const marker = "sync_2026_09_13b"
	const path = "/data/sync_2026_09_13b.json"
	if getSetting(marker+"_done") == "1" {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var p oldFixPlan
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Printf("%s: plan oxunmadı: %v", marker, err)
		return
	}
	skip := func(what string, id any) { log.Printf("%s: KEÇİLDİ %s %v (gözlənilən vəziyyət deyil)", marker, what, id) }
	fail := func(what string, id any, err error) { log.Printf("%s: XƏTA %s %v: %v", marker, what, id, err) }
	var nBranch, nSerial, nCost, nSoldCost, nStock, nSold, nCloseC, nCloseE, nOpen, nSellC int

	// 1) filiallar arası transfer — yalnız stokdakı mal
	for _, m := range p.SetBranch {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.Status != "in_stock" {
			skip("set_branch", m.ItemID)
			continue
		}
		if it.BranchID != m.BranchID {
			db.Model(&it).Update("branch_id", m.BranchID)
			nBranch++
		}
	}

	// 2) seriya düzəlişi (From yoxlaması ilə)
	for _, m := range p.SetSerial {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || normSerial(it.Serial) != normSerial(m.From) {
			skip("set_serial", m.ItemID)
			continue
		}
		db.Model(&it).Update("serial", m.To)
		// realizasiya sətri seriyanın öz kopyasını saxlayır (Realizasiya səhifəsi onu göstərir) — onu da düzəlt
		db.Model(&Consignment{}).Where("item_id = ?", it.ID).Update("serial", m.To)
		nSerial++
	}

	// 3) stokdakı malın maya dəyəri
	for _, m := range p.SetCost {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.Status != "in_stock" {
			skip("set_cost", m.ItemID)
			continue
		}
		if it.Cost != m.Cost {
			db.Model(&it).Update("cost", m.Cost)
			nCost++
		}
	}

	// 4) satılmış malın maya dəyəri + həmin satışın mənfəəti
	for _, m := range p.FixSoldCost {
		var it Item
		var s Sale
		if db.First(&it, m.ItemID).Error != nil || it.Status != "sold" ||
			db.First(&s, m.SaleID).Error != nil || s.ItemID != it.ID {
			skip("fix_sold_cost", m.ItemID)
			continue
		}
		q := s.Quantity
		if q < 1 {
			q = 1
		}
		db.Model(&it).Update("cost", m.Cost)
		db.Model(&s).Update("profit", s.SalePrice-m.Cost*float64(q))
		nSoldCost++
	}

	// 5) DB-də olmayan stok malları
	for _, m := range p.AddStock {
		if serialTaken(m.Serial) {
			skip("add_stock", m.Serial)
			continue
		}
		if err := db.Create(&Item{Name: m.Name, Serial: m.Serial, BranchID: m.BranchID, CategoryID: m.CategoryID,
			Cost: m.Cost, Quantity: 1, Status: "in_stock", ShowOnSite: false, CreatedAt: f2date(m.CreatedAt)}).Error; err != nil {
			fail("add_stock", m.Serial, err)
			continue
		}
		nStock++
	}

	// 6) Excel-də satılıb, DB-də stokda qalan mal
	for _, m := range p.SellItem {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.Status != "in_stock" {
			skip("sell_item", m.ItemID)
			continue
		}
		var have int64
		db.Model(&Sale{}).Where("item_id = ?", it.ID).Count(&have)
		if have > 0 {
			skip("sell_item(artıq satış var)", m.ItemID)
			continue
		}
		if err := db.Create(&Sale{ItemID: it.ID, SalePrice: m.SalePrice, Quantity: 1, Profit: m.SalePrice - it.Cost,
			Channel: "cash", BranchID: it.BranchID, SoldAt: f2date(m.SoldAt), Counted: true}).Error; err != nil {
			fail("sell_item", it.ID, err)
			continue
		}
		db.Model(&it).Updates(map[string]any{"status": "sold", "quantity": 0, "show_on_site": false})
		nSold++
	}

	// 7) realizasiyadakı mal satılıb → sold_paid + satış (updateConsignment ilə eyni)
	for _, m := range p.CloseConsignmentSold {
		var c Consignment
		if db.First(&c, m.ConsignmentID).Error != nil || c.Status != "out" || c.ItemID == nil || c.SaleID != nil {
			skip("close_consignment_sold", m.ConsignmentID)
			continue
		}
		var it Item
		if db.First(&it, *c.ItemID).Error != nil || it.Status != "consignment" {
			skip("close_consignment_sold(mal)", m.ConsignmentID)
			continue
		}
		var have int64
		db.Model(&Sale{}).Where("item_id = ?", it.ID).Count(&have)
		if have > 0 { // yarımçıq qalmış əvvəlki işləmə — ikinci satış yaratma
			skip("close_consignment_sold(artıq satış var)", m.ConsignmentID)
			continue
		}
		sale := Sale{ItemID: it.ID, SalePrice: m.SalePrice, Quantity: 1, Profit: m.SalePrice - it.Cost,
			Channel: "cash", BranchID: it.BranchID, SoldAt: f2date(m.SoldAt), Counted: true}
		if err := db.Create(&sale).Error; err != nil || sale.ID == 0 {
			fail("close_consignment_sold", m.ConsignmentID, err)
			continue
		}
		db.Model(&c).Updates(map[string]any{"status": "sold_paid", "debt": 0, "sale_id": sale.ID, "given_price": m.SalePrice})
		db.Model(&it).Updates(map[string]any{"status": "sold", "quantity": 0, "show_on_site": false})
		nCloseC++
	}

	// 8) mal artıq satılıb (satış qeydi var), realizasiya sətri açıq qalıb → mövcud satışa bağla
	for _, m := range p.CloseConsignmentExistingSale {
		var c Consignment
		var s Sale
		if db.First(&c, m.ConsignmentID).Error != nil || c.Status != "out" || c.ItemID == nil || c.SaleID != nil ||
			db.First(&s, m.SaleID).Error != nil || s.ItemID != *c.ItemID {
			skip("close_consignment_existing_sale", m.ConsignmentID)
			continue
		}
		db.Model(&c).Updates(map[string]any{"status": "sold_paid", "debt": 0, "sale_id": s.ID})
		db.Model(&Item{}).Where("id = ? AND status <> ?", s.ItemID, "sold").Updates(map[string]any{"status": "sold", "show_on_site": false})
		nCloseE++
	}

	// 9) «realizasiyada» statuslu, sətri olmayan mal → açıq realizasiya sətri (createConsignment ilə eyni)
	for _, m := range p.OpenConsignmentRow {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.Status != "consignment" {
			skip("open_consignment_row", m.ItemID)
			continue
		}
		var open int64
		db.Model(&Consignment{}).Where("item_id = ? AND status = ?", it.ID, "out").Count(&open)
		if open > 0 {
			skip("open_consignment_row(artıq var)", m.ItemID)
			continue
		}
		gp := m.GivenPrice
		if gp == 0 {
			gp = it.Price
		}
		id := it.ID
		if err := db.Create(&Consignment{StoreName: m.StoreName, ItemID: &id, ItemName: it.Name, Serial: it.Serial,
			GivenAt: f2date(m.GivenAt), Cost: it.Cost, GivenPrice: gp, Status: "out", Debt: it.Cost}).Error; err != nil {
			fail("open_consignment_row", it.ID, err)
			continue
		}
		nOpen++
	}

	// 10) «realizasiyada» statuslu, sətri və satışı olmayan, Excel-də satılmış mal → sətir (sold_paid) + satış
	for _, m := range p.SellConsignmentItem {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.Status != "consignment" {
			skip("sell_consignment_item", m.ItemID)
			continue
		}
		var open, have int64
		db.Model(&Consignment{}).Where("item_id = ?", it.ID).Count(&open)
		db.Model(&Sale{}).Where("item_id = ?", it.ID).Count(&have)
		if open > 0 || have > 0 {
			skip("sell_consignment_item(sətir/satış var)", m.ItemID)
			continue
		}
		sale := Sale{ItemID: it.ID, SalePrice: m.SalePrice, Quantity: 1, Profit: m.SalePrice - it.Cost,
			Channel: "cash", BranchID: it.BranchID, SoldAt: f2date(m.SoldAt), Counted: true}
		if err := db.Create(&sale).Error; err != nil || sale.ID == 0 {
			fail("sell_consignment_item", it.ID, err)
			continue
		}
		id, sid := it.ID, sale.ID
		if err := db.Create(&Consignment{StoreName: m.StoreName, ItemID: &id, ItemName: it.Name, Serial: it.Serial,
			GivenAt: f2date(m.GivenAt), Cost: it.Cost, GivenPrice: m.SalePrice, Status: "sold_paid", Debt: 0, SaleID: &sid}).Error; err != nil {
			fail("sell_consignment_item(realizasiya sətri)", it.ID, err) // satış yazılıb — mal yenə «satıldı» olur
		}
		db.Model(&it).Updates(map[string]any{"status": "sold", "quantity": 0, "show_on_site": false})
		nSellC++
	}

	log.Printf("%s: transfer=%d seriya=%d maya=%d satilmis_maya=%d yeni_stok=%d satis=%d "+
		"realiz_satildi=%d realiz_movcud_satisa=%d realiz_setri=%d realiz_mal_satildi=%d",
		marker, nBranch, nSerial, nCost, nSoldCost, nStock, nSold, nCloseC, nCloseE, nOpen, nSellC)
	setSetting(marker+"_done", "1")
	os.Remove(path)
}
