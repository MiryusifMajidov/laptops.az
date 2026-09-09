package main

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// BİRDƏFƏLİK: Excel «KREDIT» vərəqinin laptops.az-a köçürülməsi.
// Plan /data/credit_backfill_2026_09.json faylından oxunur, tətbiqdən sonra fayl silinir.
//
// Sahibkarın qərarı ilə: İLKİN ÖDƏNİŞ = 0, ilk ödəniş tarixi = satış tarixi + 1 ay.
//
// ⚠️ HEÇ BİR MƏHSUL SİLİNMİR. Mövcud məhsul varsa o istifadə olunur (yenisi yaradılmır),
// yoxdursa yeni məhsul yaradılır və «kreditə atılır» (status=sold + credit_plan).

type creditRow struct {
	XLRow         int     `json:"xl_row"`
	Serial        string  `json:"serial"`
	Name          string  `json:"name"`
	Customer      string  `json:"customer"`
	CategoryID    uint    `json:"category_id"`
	Cost          float64 `json:"cost"`
	SalePrice     float64 `json:"sale_price"`
	SoldAt        string  `json:"sold_at"`
	NextDue       string  `json:"next_due"`
	ExistingItem  uint    `json:"existing_item_id"`
	DateApprox    bool    `json:"date_approx"`
}

func cbDate(s string) time.Time {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	return time.Now()
}

func applyCreditBackfill202609() {
	if getSetting("credit_backfill_2026_09_done") == "1" {
		return
	}
	raw, err := os.ReadFile("/data/credit_backfill_2026_09.json")
	if err != nil {
		return
	}
	var plan struct {
		Credits []creditRow `json:"credits"`
	}
	if err := json.Unmarshal(raw, &plan); err != nil {
		log.Printf("credit_backfill: plan oxunmadı: %v", err)
		return
	}

	custCache := map[string]uint{}
	var nCust, nItemNew, nItemReuse, nSale, nPlan, nSkip int
	now := time.Now()

	for _, cr := range plan.Credits {
		// ---- 1) müştəri (ada görə təkrarsız)
		var custID uint
		if id, ok := custCache[cr.Customer]; ok {
			custID = id
		} else {
			var cust Customer
			if db.Where("name = ?", cr.Customer).First(&cust).Error == nil {
				custID = cust.ID
			} else {
				cust = Customer{Name: cr.Customer, CreatedAt: now}
				db.Create(&cust)
				custID = cust.ID
				nCust++
			}
			custCache[cr.Customer] = custID
		}

		// ---- 2) məhsul: mövcudu tap, yoxdursa YARAT (heç vaxt silmə)
		var it Item
		found := false
		if cr.ExistingItem > 0 && db.First(&it, cr.ExistingItem).Error == nil {
			found = true
		} else if cr.Serial != "" {
			if db.Where("lower(replace(serial,' ','')) = lower(replace(?,' ',''))", cr.Serial).
				First(&it).Error == nil {
				found = true
			}
		}
		if found {
			nItemReuse++
		} else {
			it = Item{
				Name: cr.Name, Serial: cr.Serial, CategoryID: cr.CategoryID, BranchID: 1,
				Cost: cr.Cost, Price: cr.SalePrice, Quantity: 1,
				Status: "in_stock", ShowOnSite: false, CreatedAt: cbDate(cr.SoldAt),
			}
			db.Create(&it)
			nItemNew++
		}

		// ---- 3) artıq kredit qeydi varsa keç (idempotent)
		var already int64
		db.Model(&CreditPlan{}).Where("item_name = ? AND customer_id = ?", it.Name, custID).Count(&already)
		if already > 0 {
			nSkip++
			continue
		}

		// ---- 4) satış (kredit kanalı; pul gəlmədiyi üçün hesabata sayılmır)
		var saleID *uint
		var haveSale int64
		db.Model(&Sale{}).Where("item_id = ?", it.ID).Count(&haveSale)
		if haveSale == 0 {
			s := Sale{
				ItemID: it.ID, SalePrice: cr.SalePrice, Quantity: 1,
				Profit: cr.SalePrice - cr.Cost, Channel: "credit",
				CustomerID: &custID, BranchID: 1,
				SoldAt: cbDate(cr.SoldAt), Counted: false, Pending: false,
			}
			db.Create(&s)
			saleID = &s.ID
			nSale++
		}
		// məhsulu «kreditə at» — silmə yox, status dəyişikliyi
		db.Model(&it).Updates(map[string]any{"status": "sold", "show_on_site": false})

		// ---- 5) kredit planı: ilkin ödəniş 0, ilk ödəniş 1 ay sonra
		due := cbDate(cr.NextDue)
		st := "ontime"
		if due.Before(time.Now().Truncate(24 * time.Hour)) {
			st = "overdue"
		} else if due.Format("2006-01-02") == time.Now().Format("2006-01-02") {
			st = "duetoday"
		}
		db.Create(&CreditPlan{
			CustomerID: custID, ItemName: it.Name,
			Total: cr.SalePrice, Paid: 0, NextDue: due,
			Status: st, Closed: false, SaleID: saleID,
		})
		nPlan++
	}

	log.Printf("credit_backfill: müştəri=%d yeni_məhsul=%d mövcud_məhsul=%d satış=%d kredit=%d keçildi=%d",
		nCust, nItemNew, nItemReuse, nSale, nPlan, nSkip)
	setSetting("credit_backfill_2026_09_done", "1")
	os.Remove("/data/credit_backfill_2026_09.json")
}
