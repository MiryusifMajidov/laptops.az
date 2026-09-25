package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// POST /api/upload  (multipart form, field "file") → {url}
func uploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		writeJSON(w, 400, map[string]string{"error": "fayl oxunmadı"})
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "fayl yoxdur"})
		return
	}
	defer file.Close()
	os.MkdirAll("uploads", 0755)
	ext := filepath.Ext(hdr.Filename)
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst, err := os.Create("uploads/" + name)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "yaddaşa yazılmadı"})
		return
	}
	defer dst.Close()
	io.Copy(dst, file)
	writeJSON(w, 201, map[string]string{"url": "/uploads/" + name})
}

// POST /api/credits — yeni kredit müqaviləsi
func createCredit(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CustomerID uint    `json:"customer_id"`
		ItemName   string  `json:"item_name"`
		Total      float64 `json:"total"`
		Paid       float64 `json:"paid"`
		NextDue    string  `json:"next_due"`
	}
	if err := decodeBody(r, &in); err != nil || in.CustomerID == 0 {
		writeJSON(w, 400, map[string]string{"error": "müştəri və məbləğ vacibdir"})
		return
	}
	due := time.Now().AddDate(0, 1, 0)
	if in.NextDue != "" {
		if t, e := time.Parse("2006-01-02", in.NextDue); e == nil {
			due = t
		}
	}
	c := CreditPlan{CustomerID: in.CustomerID, ItemName: in.ItemName, Total: in.Total, Paid: in.Paid, NextDue: due, Status: creditStatus(due, in.Paid, in.Total)}
	db.Create(&c)
	db.Preload("Customer").First(&c, c.ID)
	writeJSON(w, 201, c)
}

// POST /api/credits/{id}/pay  {amount}
func payCredit(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Amount  float64 `json:"amount"`
		NextDue string  `json:"next_due"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	var c CreditPlan
	if err := db.First(&c, id).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "tapılmadı"})
		return
	}
	c.Paid += in.Amount
	if in.NextDue != "" {
		if t, e := time.Parse("2006-01-02", in.NextDue); e == nil {
			c.NextDue = t
		}
	}
	c.Status = creditStatus(c.NextDue, c.Paid, c.Total)
	db.Save(&c)
	// ödəniş tarixçəsinə ayrıca qeyd
	db.Create(&CreditPayment{CreditPlanID: c.ID, Amount: in.Amount, CreatedAt: time.Now()})
	// tam ödənildisə bağlı satış «sayılan» olur
	syncCreditSale(&c)
	db.Preload("Customer").Preload("Payments").First(&c, c.ID)
	writeJSON(w, 200, c)
}

func creditStatus(due time.Time, paid, total float64) string {
	if paid >= total {
		return "paid"
	}
	days := int(time.Until(due).Hours() / 24)
	if days < 0 {
		return "overdue"
	}
	if days == 0 {
		return "duetoday"
	}
	return "ontime"
}

// kredit "gerçəkləşib"mi — tam ödənib VƏ YA əl ilə bağlanıb
func creditRealized(cp *CreditPlan) bool {
	return cp.Closed || cp.Paid >= cp.Total
}

// bağlı satışın Counted (sayılır) sahəsini kreditin vəziyyətinə uyğunlaşdır
func syncCreditSale(cp *CreditPlan) {
	if cp.SaleID == nil {
		return
	}
	db.Model(&Sale{}).Where("id = ?", *cp.SaleID).Update("counted", creditRealized(cp))
}

// PUT /api/credits/{id}  {closed} — "biz özümüz bağladıq" checkbox
func updateCredit(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Closed *bool `json:"closed"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	var c CreditPlan
	if err := db.First(&c, id).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "tapılmadı"})
		return
	}
	if in.Closed != nil {
		c.Closed = *in.Closed
		db.Model(&c).Update("closed", c.Closed)
	}
	syncCreditSale(&c)
	db.Preload("Customer").Preload("Payments").First(&c, c.ID)
	writeJSON(w, 200, c)
}

// PUT /api/consignments/{id}  {status} — realizasiya status keçidi
func updateConsignment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Status string `json:"status"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	var c Consignment
	if err := db.First(&c, id).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "tapılmadı"})
		return
	}
	c.Status = in.Status
	if in.Status == "sold_paid" || in.Status == "returned" {
		c.Debt = 0
	} else {
		c.Debt = c.Cost
	}

	// «Satılıb · ödənilib» → pul kassaya gəldi, satış yarat (Satışlara düşür).
	// Başqa statusa keçsə (geri, hələ ödənilməyib) → əvvəl yaradılmış satışı sil.
	if in.Status == "sold_paid" {
		if c.SaleID == nil && c.ItemID != nil {
			var it Item
			db.First(&it, *c.ItemID)
			// Malın artıq satış qeydi varsa TƏKRAR YARATMA — mövcud satışı mənimsə.
			// (Bir fiziki satış üçün iki pul qeydi qəti olmaz.)
			// Şərtlər: satış realizasiyaya verildikdən SONRA olmalı, başqa realizasiyaya
			// bağlanmamış olmalı; çoxsaylı (say>1) aksesuar kartlarında mənimsəmə edilmir.
			var ex Sale
			found := false
			if it.Quantity <= 1 {
				found = db.Where("item_id = ? AND sold_at >= ?", *c.ItemID, c.GivenAt).
					Where("id NOT IN (SELECT sale_id FROM consignments WHERE sale_id IS NOT NULL)").
					Order("id desc").First(&ex).Error == nil
			}
			if found {
				c.SaleID = &ex.ID
				c.SaleOwned = false // bizim yaratmadığımız satış — geri dönəndə silinməyəcək
			} else {
				sale := Sale{
					ItemID: *c.ItemID, SalePrice: c.GivenPrice, Profit: c.GivenPrice - c.Cost,
					Channel: "cash", BranchID: it.BranchID, SoldAt: time.Now(), Counted: true,
				}
				db.Create(&sale)
				c.SaleID = &sale.ID
				c.SaleOwned = true
			}
		}
	} else if c.SaleID != nil {
		// Pul qeydi itməsin: yalnız BU axının yaratdığı satış silinir. Kənardan gələn
		// (Excel ilə təsdiqlənmiş) satış saxlanılır, sadəcə bağlantı açılır.
		var old Sale
		hasOld := db.First(&old, *c.SaleID).Error == nil
		if c.SaleOwned && hasOld {
			log.Printf("realizasiya #%d → «%s»: satış silinir id=%d mal=%d %.2f₼ %s mənfəət=%.2f filial=%d",
				c.ID, in.Status, old.ID, old.ItemID, old.SalePrice,
				old.SoldAt.Format("2006-01-02"), old.Profit, old.BranchID)
			db.Delete(&Sale{}, *c.SaleID)
		} else if hasOld {
			log.Printf("realizasiya #%d → «%s»: satış #%d SAXLANILIR (kənar qeyd), yalnız bağlantı açılır",
				c.ID, in.Status, old.ID)
		}
		c.SaleID = nil
		c.SaleOwned = false
	}

	db.Save(&c)

	// cihazın statusu: qaytarıldı→stoka, satılıb·ödənilib→satıldı, digər→realizasiyada
	if c.ItemID != nil {
		st := "consignment"
		switch in.Status {
		case "returned":
			st = "in_stock"
		case "sold_paid":
			st = "sold"
		}
		db.Model(&Item{}).Where("id = ?", *c.ItemID).Update("status", st)
	}
	writeJSON(w, 200, c)
}

// PUT /api/supplies/{id}  {status} — partiya təsdiqi
func updateSupply(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Status string `json:"status"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	db.Model(&SupplyBatch{}).Where("id = ?", id).Update("status", in.Status)
	var s SupplyBatch
	db.First(&s, id)
	writeJSON(w, 200, s)
}
