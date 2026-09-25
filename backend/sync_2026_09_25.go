package main

import (
	"encoding/json"
	"log"
	"os"
)

// BİRDƏFƏLİK (2026-09-25): Excel-in son 1 ayındakı YENİ hərəkətlər DB-yə köçürülür.
// Sahibkarın göstərişi: «əvvəlki məhsullara müdaxilə olunmasın, sadəcə yenilikləri et».
// Ona görə burada qiymət düzəlişi, dublikat birləşdirmə və köhnə açıq məsələlər YOXDUR —
// yalnız yeni gələn mal, yeni satış, filial transferi və realizasiya hərəkətləri.
//
// ⚠️ MƏHSUL SİLİNMİR. Kreditə toxunulmur. Hər əməliyyat gözlənilən vəziyyəti yoxlayır;
// uyğun gəlməsə keçir və loglayır (təkrar işləsə heç nəyi ikiqat etmir).

type sync25Plan struct {
	AddStock []struct {
		Name       string  `json:"name"`
		Serial     string  `json:"serial"`
		BranchID   uint    `json:"branch_id"`
		CategoryID uint    `json:"category_id"`
		Cost       float64 `json:"cost"`
		CreatedAt  string  `json:"created_at"`
	} `json:"add_stock"`
	// admin malı əl ilə yaradıb, mayasını boş qoyub → YALNIZ boş mayanı tamamlayır
	SetCost []struct {
		ItemID uint    `json:"item_id"`
		Cost   float64 `json:"cost"`
	} `json:"set_cost"`
	SetBranch []struct {
		ItemID   uint `json:"item_id"`
		BranchID uint `json:"branch_id"`
	} `json:"set_branch"`
	SellItem []struct {
		ItemID    uint    `json:"item_id"`
		SalePrice float64 `json:"sale_price"`
		SoldAt    string  `json:"sold_at"`
	} `json:"sell_item"`
	AddItemAndSale []struct {
		Name       string  `json:"name"`
		Serial     string  `json:"serial"`
		BranchID   uint    `json:"branch_id"`
		CategoryID uint    `json:"category_id"`
		Cost       float64 `json:"cost"`
		SalePrice  float64 `json:"sale_price"`
		SoldAt     string  `json:"sold_at"`
	} `json:"add_item_and_sale"`
	ConsignItem []struct {
		ItemID     uint    `json:"item_id"`
		StoreName  string  `json:"store_name"`
		GivenAt    string  `json:"given_at"`
		GivenPrice float64 `json:"given_price"`
	} `json:"consign_item"`
	AddItemAndConsign []struct {
		Name       string  `json:"name"`
		Serial     string  `json:"serial"`
		BranchID   uint    `json:"branch_id"`
		CategoryID uint    `json:"category_id"`
		Cost       float64 `json:"cost"`
		StoreName  string  `json:"store_name"`
		GivenAt    string  `json:"given_at"`
		GivenPrice float64 `json:"given_price"`
	} `json:"add_item_and_consign"`
	CloseConsignmentSold []struct {
		ConsignmentID uint    `json:"consignment_id"`
		SalePrice     float64 `json:"sale_price"`
		SoldAt        string  `json:"sold_at"`
		// malın mayası DB-də boşdursa Excel-dəki maya (mənfəət uydurma çıxmasın deyə)
		Cost float64 `json:"cost"`
	} `json:"close_consignment_sold"`
}

func applySync20260925() {
	const marker = "sync_2026_09_25"
	const path = "/data/sync_2026_09_25.json"
	if getSetting(marker+"_done") == "1" {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var p sync25Plan
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Printf("%s: plan oxunmadı: %v", marker, err)
		return
	}
	skip := func(what string, id any) { log.Printf("%s: KEÇİLDİ %s %v (gözlənilən vəziyyət deyil)", marker, what, id) }
	fail := func(what string, id any, err error) { log.Printf("%s: XƏTA %s %v: %v", marker, what, id, err) }
	var nStock, nCost, nBranch, nSold, nNewSold, nCons, nNewCons, nCloseC int

	// 1) yeni gələn mallar
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

	// 1b) boş qalmış maya dəyərini tamamla — mövcud düzgün mayanın üstünə YAZMIR
	for _, m := range p.SetCost {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.Cost != 0 || it.Status != "in_stock" {
			skip("set_cost", m.ItemID)
			continue
		}
		db.Model(&it).Update("cost", m.Cost)
		nCost++
	}

	// 2) filiallar arası transfer
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

	// 3) Excel-də satılıb, DB-də stokda qalan mal
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

	// 4) DB-də ümumiyyətlə olmayan, gəlib satılmış mal (seriyasız aksesuar/işlənmiş satışlar da burada).
	// Seriyasızları serialTaken qorumur → bütün addım üçün ayrıca marker: təkrar işləsə dublikat yaranmır.
	if getSetting(marker+"_s4") != "1" {
		for _, m := range p.AddItemAndSale {
			if serialTaken(m.Serial) {
				skip("add_item_and_sale", m.Serial)
				continue
			}
			// Seriyasız sətirləri serialTaken qorumur. Yarımçıq işləmədən sonra təkrar
			// başlasa dublikat yaranmasın deyə: eyni ad + eyni məbləğ + eyni gün varsa keçirik.
			var dup int64
			db.Model(&Sale{}).Joins("JOIN items ON items.id = sales.item_id").
				Where("items.name = ? AND sales.sale_price = ? AND substr(sales.sold_at,1,10) = ?",
					m.Name, m.SalePrice, m.SoldAt).Count(&dup)
			if dup > 0 {
				skip("add_item_and_sale(dublikat)", m.Name)
				continue
			}
			// ƏVVƏL «stokda» yaradılır, satış yazıldıqdan SONRA «satıldı» edilir:
			// arada proses çöksə, satış qeydi olmayan «satıldı» mal qalmır (təsdiq siyahısını zibilləmir).
			it := Item{Name: m.Name, Serial: m.Serial, BranchID: m.BranchID, CategoryID: m.CategoryID,
				Cost: m.Cost, Price: m.SalePrice, Quantity: 1, Status: "in_stock", ShowOnSite: false,
				CreatedAt: f2date(m.SoldAt)}
			if err := db.Create(&it).Error; err != nil || it.ID == 0 {
				fail("add_item_and_sale(item)", m.Serial, err)
				continue
			}
			if err := db.Create(&Sale{ItemID: it.ID, SalePrice: m.SalePrice, Quantity: 1, Profit: m.SalePrice - m.Cost,
				Channel: "cash", BranchID: m.BranchID, SoldAt: f2date(m.SoldAt), Counted: true}).Error; err != nil {
				fail("add_item_and_sale(sale)", it.ID, err)
				continue
			}
			db.Model(&it).Updates(map[string]any{"status": "sold", "quantity": 0})
			nNewSold++
		}
		setSetting(marker+"_s4", "1")
	}

	// 5) mövcud malı realizasiyaya ver (createConsignment ilə eyni: borc = maya)
	for _, m := range p.ConsignItem {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.Status != "in_stock" {
			skip("consign_item", m.ItemID)
			continue
		}
		var open int64
		db.Model(&Consignment{}).Where("item_id = ? AND status = ?", it.ID, "out").Count(&open)
		if open > 0 {
			skip("consign_item(açıq realizasiya var)", m.ItemID)
			continue
		}
		gp := m.GivenPrice
		if gp == 0 {
			gp = it.Price
		}
		id := it.ID
		if err := db.Create(&Consignment{StoreName: m.StoreName, ItemID: &id, ItemName: it.Name, Serial: it.Serial,
			GivenAt: f2date(m.GivenAt), Cost: it.Cost, GivenPrice: gp, Status: "out", Debt: it.Cost}).Error; err != nil {
			fail("consign_item", it.ID, err)
			continue
		}
		db.Model(&it).Updates(map[string]any{"status": "consignment", "show_on_site": false})
		nCons++
	}

	// 6) DB-də olmayan, gəlib birbaşa realizasiyaya verilən mal
	for _, m := range p.AddItemAndConsign {
		if serialTaken(m.Serial) {
			skip("add_item_and_consign", m.Serial)
			continue
		}
		// satış yolu ilə eyni ardıcıllıq: əvvəl «stokda», realizasiya sətri yarandıqdan sonra status dəyişir
		it := Item{Name: m.Name, Serial: m.Serial, BranchID: m.BranchID, CategoryID: m.CategoryID,
			Cost: m.Cost, Price: m.GivenPrice, Quantity: 1, Status: "in_stock", ShowOnSite: false,
			CreatedAt: f2date(m.GivenAt)}
		if err := db.Create(&it).Error; err != nil || it.ID == 0 {
			fail("add_item_and_consign(item)", m.Serial, err)
			continue
		}
		id := it.ID
		if err := db.Create(&Consignment{StoreName: m.StoreName, ItemID: &id, ItemName: it.Name, Serial: it.Serial,
			GivenAt: f2date(m.GivenAt), Cost: m.Cost, GivenPrice: m.GivenPrice, Status: "out", Debt: m.Cost}).Error; err != nil {
			fail("add_item_and_consign(consignment)", it.ID, err)
			continue
		}
		db.Model(&it).Update("status", "consignment")
		nNewCons++
	}

	// 7) realizasiyadakı mal satılıb → sold_paid + satış (updateConsignment ilə eyni)
	for _, m := range p.CloseConsignmentSold {
		var cs Consignment
		if db.First(&cs, m.ConsignmentID).Error != nil || cs.Status != "out" || cs.ItemID == nil || cs.SaleID != nil {
			skip("close_consignment_sold", m.ConsignmentID)
			continue
		}
		var it Item
		if db.First(&it, *cs.ItemID).Error != nil || it.Status != "consignment" {
			skip("close_consignment_sold(mal)", m.ConsignmentID)
			continue
		}
		var have int64
		db.Model(&Sale{}).Where("item_id = ?", it.ID).Count(&have)
		if have > 0 {
			skip("close_consignment_sold(artıq satış var)", m.ConsignmentID)
			continue
		}
		// maya boşdursa Excel-dəkini yazırıq, yoxsa mənfəət bütün satış məbləği qədər uydurma çıxar
		cost := it.Cost
		if cost == 0 && m.Cost > 0 {
			cost = m.Cost
			db.Model(&it).Update("cost", cost)
			db.Model(&cs).Update("cost", cost)
		}
		sale := Sale{ItemID: it.ID, SalePrice: m.SalePrice, Quantity: 1, Profit: m.SalePrice - cost,
			Channel: "cash", BranchID: it.BranchID, SoldAt: f2date(m.SoldAt), Counted: true}
		if err := db.Create(&sale).Error; err != nil || sale.ID == 0 {
			fail("close_consignment_sold", m.ConsignmentID, err)
			continue
		}
		db.Model(&cs).Updates(map[string]any{"status": "sold_paid", "debt": 0, "sale_id": sale.ID, "given_price": m.SalePrice})
		db.Model(&it).Updates(map[string]any{"status": "sold", "quantity": 0, "show_on_site": false})
		nCloseC++
	}

	log.Printf("%s: yeni_stok=%d maya=%d transfer=%d satis=%d yeni_satilan=%d realiz_verildi=%d "+
		"yeni_mal_realiz=%d realiz_satildi=%d",
		marker, nStock, nCost, nBranch, nSold, nNewSold, nCons, nNewCons, nCloseC)
	setSetting(marker+"_done", "1")
	os.Remove(path)
}
