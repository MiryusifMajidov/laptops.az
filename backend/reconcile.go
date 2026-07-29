package main

import (
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// POST /api/reconcile — Excel tutuşdurmasından toplu status/kateqoriya düzəlişi.
// Yalnız admin, atomik (bir transaksiya).
//
//	sold      → hər biri üçün Sale qeydi YARADILIR (satış qiyməti + mənfəət + tarix) VƏ status=sold
//	mark_sold → yalnız status=sold (satış qeydi yaradılmır — köhnə/tarixsiz satışlar)
//	used      → kateqoriya "İşlənmiş"ə keçir (işlənmiş stok; status dəyişmir)
//	new_sales → DB-də olmayan satışlar: YENİ cihaz (status=sold, saytda görünməz) + Sale yaradılır
func reconcileHandler(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromReq(r)
	if !ok || u.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "yalnız admin"})
		return
	}
	var in struct {
		ResetSales    bool   `json:"reset_sales"`     // bütün Sale qeydlərini sil (təkrar import üçün)
		DeleteItemIDs []uint `json:"delete_item_ids"` // bu cihazları sil (əvvəlki avto-yaradılanlar)
		Sold          []struct {
			ItemID    uint    `json:"item_id"`
			SalePrice float64 `json:"sale_price"`
			SoldAt    string  `json:"sold_at"` // YYYY-MM-DD (boşdursa indi)
		} `json:"sold"`
		MarkSold []uint `json:"mark_sold"`
		Used     []uint `json:"used"`
		NewSales []struct {
			Name      string  `json:"name"`
			Serial    string  `json:"serial"`
			Cost      float64 `json:"cost"`
			SalePrice float64 `json:"sale_price"`
			SoldAt    string  `json:"sold_at"` // YYYY-MM-DD
			Category  string  `json:"category"`
		} `json:"new_sales"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "JSON oxunmadı: " + err.Error()})
		return
	}

	// "İşlənmiş" kateqoriyasının id-si
	var usedCatID uint
	{
		var c Category
		if db.Where("name = ?", "İşlənmiş").First(&c).Error == nil {
			usedCatID = c.ID
		}
	}
	// əsas filial (Mərkəz) id-si — new_sales üçün
	var mainBranch uint
	{
		var b Branch
		if db.Where("name = ?", "Mərkəz (Mağaza)").First(&b).Error == nil {
			mainBranch = b.ID
		}
	}

	var nSale, nMark, nUsed, nNew, nDelSale, nDelItem int
	err := db.Transaction(func(tx *gorm.DB) error {
		// 0) təmizlik (təkrar import): əvvəlcə bütün satışlar, sonra avto-cihazlar
		if in.ResetSales {
			res := tx.Where("1 = 1").Delete(&Sale{})
			if res.Error != nil {
				return res.Error
			}
			nDelSale = int(res.RowsAffected)
		}
		for _, id := range in.DeleteItemIDs {
			res := tx.Delete(&Item{}, id)
			if res.Error != nil {
				return res.Error
			}
			nDelItem += int(res.RowsAffected)
		}
		// 1) satış qeydi + sold (təkrar link üçün: artıq satılmışsa da qeyd yaradılır)
		for _, s := range in.Sold {
			var item Item
			if tx.First(&item, s.ItemID).Error != nil {
				continue // tapılmadı — ötür
			}
			soldAt := time.Now()
			if s.SoldAt != "" {
				if t, e := time.Parse("2006-01-02", s.SoldAt); e == nil {
					soldAt = t
				}
			}
			sale := Sale{
				ItemID: item.ID, SalePrice: s.SalePrice, Profit: s.SalePrice - item.Cost,
				Channel: "store", BranchID: item.BranchID, SoldAt: soldAt, Counted: true,
			}
			if err := tx.Create(&sale).Error; err != nil {
				return err
			}
			if err := tx.Model(&item).Update("status", "sold").Error; err != nil {
				return err
			}
			nSale++
		}
		// 2) yalnız status=sold
		for _, id := range in.MarkSold {
			res := tx.Model(&Item{}).Where("id = ? AND status <> ?", id, "sold").Update("status", "sold")
			if res.Error != nil {
				return res.Error
			}
			nMark += int(res.RowsAffected)
		}
		// 3) kateqoriya → İşlənmiş
		if usedCatID != 0 {
			for _, id := range in.Used {
				res := tx.Model(&Item{}).Where("id = ?", id).Update("category_id", usedCatID)
				if res.Error != nil {
					return res.Error
				}
				nUsed += int(res.RowsAffected)
			}
		}
		// 4) DB-də olmayan satışlar → yeni cihaz (sold) + Sale
		catID := func(name string) uint {
			if strings.TrimSpace(name) == "" {
				name = "Aksesuar"
			}
			var c Category
			if tx.Where("name = ?", name).First(&c).Error != nil {
				c = Category{Name: name}
				tx.Create(&c)
			}
			return c.ID
		}
		for _, s := range in.NewSales {
			if strings.TrimSpace(s.Name) == "" {
				continue
			}
			soldAt := time.Now()
			if s.SoldAt != "" {
				if t, e := time.Parse("2006-01-02", s.SoldAt); e == nil {
					soldAt = t
				}
			}
			it := Item{
				Name: strings.TrimSpace(s.Name), Serial: strings.TrimSpace(s.Serial),
				CategoryID: catID(s.Category), BranchID: mainBranch,
				Cost: s.Cost, Price: s.SalePrice, Status: "sold", ShowOnSite: false, CreatedAt: soldAt,
			}
			if err := tx.Create(&it).Error; err != nil {
				return err
			}
			sale := Sale{
				ItemID: it.ID, SalePrice: s.SalePrice, Profit: s.SalePrice - s.Cost,
				Channel: "store", BranchID: mainBranch, SoldAt: soldAt, Counted: true,
			}
			if err := tx.Create(&sale).Error; err != nil {
				return err
			}
			nNew++
		}
		return nil
	})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]int{"sold_recorded": nSale, "marked_sold": nMark, "used": nUsed,
		"new_sales": nNew, "deleted_sales": nDelSale, "deleted_items": nDelItem})
}
