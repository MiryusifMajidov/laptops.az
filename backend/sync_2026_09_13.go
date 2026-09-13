package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

// BİRDƏFƏLİK (2026-09-13): Excel-in 10→13 sentyabr dəyişiklikləri DB-yə köçürülür —
// yeni gələn mallar, satışlar, realizasiyaya verilən / realizasiyadan çıxan mallar,
// filiallar arası transfer və maya düzəlişi.
// Plan /data/sync_2026_09_13.json-dan oxunur (biznes datası repo-da saxlanmır),
// tətbiqdən sonra fayl silinir, marker qoyulur — təkrar işləmir.
//
// ⚠️ MƏHSUL SİLİNMİR. Kreditə toxunulmur (kredit satışını mağaza özü yazır).
// Hər əməliyyat gözlənilən vəziyyəti yoxlayır; uyğun gəlməsə keçir və loglayır.
// Satış/realizasiya məntiqi tətbiqin öz handler-ləri ilə eynidir (createSale, createConsignment,
// updateConsignment → sold_paid).
// Seriyasız addımlar (4 və 6) təkrar işləsə dublikat yaradardı — onlar ayrıca addım-marker alır.

type sync13Plan struct {
	SetCost []struct {
		ItemID uint    `json:"item_id"`
		Cost   float64 `json:"cost"`
	} `json:"set_cost"`
	SetBranch []struct {
		ItemID   uint `json:"item_id"`
		BranchID uint `json:"branch_id"`
	} `json:"set_branch"`
	AddStock []struct {
		Name       string  `json:"name"`
		Serial     string  `json:"serial"`
		BranchID   uint    `json:"branch_id"`
		CategoryID uint    `json:"category_id"`
		Cost       float64 `json:"cost"`
		CreatedAt  string  `json:"created_at"`
	} `json:"add_stock"`
	AddItemAndSale []struct {
		Name       string  `json:"name"`
		Serial     string  `json:"serial"`
		BranchID   uint    `json:"branch_id"`
		CategoryID uint    `json:"category_id"`
		Cost       float64 `json:"cost"`
		SalePrice  float64 `json:"sale_price"`
		SoldAt     string  `json:"sold_at"`
	} `json:"add_item_and_sale"`
	SellItem []struct {
		ItemID    uint    `json:"item_id"`
		SalePrice float64 `json:"sale_price"`
		SoldAt    string  `json:"sold_at"`
	} `json:"sell_item"`
	SellFromStock []struct {
		ItemID    uint    `json:"item_id"`
		SalePrice float64 `json:"sale_price"` // BİR ədədin qiyməti
		Qty       int     `json:"qty"`
		SoldAt    string  `json:"sold_at"`
	} `json:"sell_from_stock"`
	CloseConsignmentSold []struct {
		ConsignmentID uint    `json:"consignment_id"`
		SalePrice     float64 `json:"sale_price"`
		SoldAt        string  `json:"sold_at"`
	} `json:"close_consignment_sold"`
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
	ConsignLoose []struct {
		ItemName   string  `json:"item_name"`
		Serial     string  `json:"serial"`
		Cost       float64 `json:"cost"`
		GivenPrice float64 `json:"given_price"`
		StoreName  string  `json:"store_name"`
		GivenAt    string  `json:"given_at"`
	} `json:"consign_loose"`
}

// serialTaken — seriya DB-də (silinmişlər daxil) artıq varmı. Dublikat yaratmamaq üçün.
func serialTaken(serial string) bool {
	if strings.TrimSpace(serial) == "" {
		return false
	}
	var n int64
	db.Unscoped().Model(&Item{}).
		Where("lower(replace(serial,' ','')) = lower(replace(?,' ',''))", serial).Count(&n)
	return n > 0
}

func applySync20260913() {
	const marker = "sync_2026_09_13"
	const path = "/data/sync_2026_09_13.json"
	if getSetting(marker+"_done") == "1" {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var p sync13Plan
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Printf("%s: plan oxunmadı: %v", marker, err)
		return
	}
	skip := func(what string, id any) { log.Printf("%s: KEÇİLDİ %s %v (gözlənilən vəziyyət deyil)", marker, what, id) }
	fail := func(what string, id any, err error) { log.Printf("%s: XƏTA %s %v: %v", marker, what, id, err) }
	var nCost, nBranch, nStock, nNewSold, nSold, nAcc, nCloseC, nCons, nNewCons, nLoose int

	// 1) maya düzəlişi — satışdan ƏVVƏL (mənfəət düzgün maya ilə hesablansın)
	for _, m := range p.SetCost {
		var it Item
		if db.First(&it, m.ItemID).Error != nil || it.Status == "sold" {
			skip("set_cost", m.ItemID)
			continue
		}
		if it.Cost != m.Cost {
			db.Model(&it).Update("cost", m.Cost)
			nCost++
		}
	}

	// 2) filiallar arası transfer — yalnız stokdakı mal
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

	// 3) yeni gələn mallar (seriya ilə dublikatdan qorunur)
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

	// 4) DB-də olmayan və elə gəldiyi günlərdə satılan mal: məhsul + satış.
	// Seriyasız aksesuarlar serialTaken ilə qorunmur → addım-marker.
	if getSetting(marker+"_s4") != "1" {
		for _, m := range p.AddItemAndSale {
			if serialTaken(m.Serial) {
				skip("add_item_and_sale", m.Serial)
				continue
			}
			it := Item{Name: m.Name, Serial: m.Serial, BranchID: m.BranchID, CategoryID: m.CategoryID,
				Cost: m.Cost, Price: m.SalePrice, Quantity: 1, Status: "sold", ShowOnSite: false,
				CreatedAt: f2date(m.SoldAt)}
			if err := db.Create(&it).Error; err != nil || it.ID == 0 {
				fail("add_item_and_sale(item)", m.Name, err)
				continue
			}
			// Quantity-nin gorm default:1 tag-ı var — 0-ı Create-də yazmır; createSale kimi sonra 0 edirik
			db.Model(&it).Update("quantity", 0)
			if err := db.Create(&Sale{ItemID: it.ID, SalePrice: m.SalePrice, Quantity: 1, Profit: m.SalePrice - m.Cost,
				Channel: "cash", BranchID: m.BranchID, SoldAt: f2date(m.SoldAt), Counted: true}).Error; err != nil {
				fail("add_item_and_sale(sale)", it.ID, err)
				continue
			}
			nNewSold++
		}
		setSetting(marker+"_s4", "1")
	}

	// 5) mövcud malın satışı (seriyalı, 1 ədəd) — status + mövcud satış yoxlaması ilə qorunur
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

	// 6) çox ədədli stokdan satış (aksesuar) — createSale ilə eyni: miqdar azalır, 0-da «satıldı».
	// Təkrar işləsə miqdarı yenidən azaldardı → addım-marker.
	if getSetting(marker+"_s6") != "1" {
		for _, m := range p.SellFromStock {
			var it Item
			if db.First(&it, m.ItemID).Error != nil || it.Status != "in_stock" {
				skip("sell_from_stock", m.ItemID)
				continue
			}
			qty := m.Qty
			if qty < 1 {
				qty = 1
			}
			if it.Quantity < qty {
				skip("sell_from_stock(miqdar çatmır)", m.ItemID)
				continue
			}
			total := m.SalePrice * float64(qty)
			if err := db.Create(&Sale{ItemID: it.ID, SalePrice: total, Quantity: qty, Profit: total - it.Cost*float64(qty),
				Channel: "cash", BranchID: it.BranchID, SoldAt: f2date(m.SoldAt), Counted: true}).Error; err != nil {
				fail("sell_from_stock", it.ID, err)
				continue
			}
			left := it.Quantity - qty
			if left <= 0 {
				db.Model(&it).Updates(map[string]any{"quantity": 0, "status": "sold"})
			} else {
				db.Model(&it).Update("quantity", left)
			}
			nAcc++
		}
		setSetting(marker+"_s6", "1")
	}

	// 7) realizasiyadakı mal satılıb, pulu gəlib → sold_paid (updateConsignment ilə eyni).
	// updateConsignment satışı given_price ilə yazır; Excel-in real satış qiyməti fərqlidirsə
	// realizasiya sətrinin qiyməti də ona bərabərləşdirilir ki, iki qeyd bir-birinə zidd olmasın.
	for _, m := range p.CloseConsignmentSold {
		var c Consignment
		if db.First(&c, m.ConsignmentID).Error != nil || c.Status != "out" || c.ItemID == nil || c.SaleID != nil {
			skip("close_consignment_sold", m.ConsignmentID)
			continue
		}
		var it Item
		if db.First(&it, *c.ItemID).Error != nil {
			skip("close_consignment_sold(mal yoxdur)", m.ConsignmentID)
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

	// 8) mövcud malı realizasiyaya ver (createConsignment ilə eyni: borc = maya, qiymət boşdursa = satış qiyməti)
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

	// 9) DB-də olmayan, gəlib birbaşa realizasiyaya verilən mal: məhsul + realizasiya
	for _, m := range p.AddItemAndConsign {
		if serialTaken(m.Serial) {
			skip("add_item_and_consign", m.Serial)
			continue
		}
		it := Item{Name: m.Name, Serial: m.Serial, BranchID: m.BranchID, CategoryID: m.CategoryID,
			Cost: m.Cost, Price: m.GivenPrice, Quantity: 1, Status: "consignment", ShowOnSite: false,
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
		nNewCons++
	}

	// 10) stok kartı olmayan aksesuarın realizasiyası (createConsignment-in item_id-siz qolu)
	for _, m := range p.ConsignLoose {
		var dup int64
		db.Model(&Consignment{}).Where("item_id IS NULL AND item_name = ? AND store_name = ? AND status = ?",
			m.ItemName, m.StoreName, "out").Count(&dup)
		if dup > 0 {
			skip("consign_loose(dublikat)", m.ItemName)
			continue
		}
		if err := db.Create(&Consignment{StoreName: m.StoreName, ItemName: m.ItemName, Serial: m.Serial,
			GivenAt: f2date(m.GivenAt), Cost: m.Cost, GivenPrice: m.GivenPrice, Status: "out", Debt: m.Cost}).Error; err != nil {
			fail("consign_loose", m.ItemName, err)
			continue
		}
		nLoose++
	}

	log.Printf("%s: maya=%d transfer=%d yeni_stok=%d yeni_satilan=%d satis=%d aksesuar_satis=%d "+
		"realiz_satildi=%d realiz_verildi=%d yeni_mal_realiz=%d aksesuar_realiz=%d",
		marker, nCost, nBranch, nStock, nNewSold, nSold, nAcc, nCloseC, nCons, nNewCons, nLoose)
	setSetting(marker+"_done", "1")
	os.Remove(path)
}
