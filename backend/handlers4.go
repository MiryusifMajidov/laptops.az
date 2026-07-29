package main

import (
	"net/http"
	"strconv"
	"time"
)

// tarix sətri (YYYY-MM-DD) → günorta saatına time.Time (gün qruplaşdırması üçün neytral)
func parseDayPtr(s *string) (time.Time, bool) {
	if s == nil || *s == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02", *s, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t.Add(12 * time.Hour), true
}

// PUT /api/items/{id} — məhsulu redaktə et (qismən update dəstəklənir)
func updateItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var item Item
	if err := db.First(&item, id).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "cihaz tapılmadı"})
		return
	}
	var in struct {
		Name           *string   `json:"name"`
		Serial         *string   `json:"serial"`
		Cost           *float64  `json:"cost"`
		Price          *float64  `json:"price"`
		WholesalePrice *float64  `json:"wholesale_price"`
		Discount       *float64  `json:"discount"`
		Quantity       *int      `json:"quantity"`
		CategoryID     *uint     `json:"category_id"`
		BranchID       *uint     `json:"branch_id"`
		Status         *string   `json:"status"`
		ShowOnSite     *bool     `json:"show_on_site"`
		CardImage      *string   `json:"card_image"`
		Gallery        *string   `json:"gallery"`
		CreatedAt      *string   `json:"created_at"` // alınma tarixi (YYYY-MM-DD)
		SoldAt         *string   `json:"sold_at"`    // satılma tarixi (YYYY-MM-DD)
		Translations   *[]itemTr `json:"translations"`
		Values         *[]struct {
			AttributeID uint   `json:"attribute_id"`
			Value       string `json:"value"`
		} `json:"values"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	updates := map[string]any{}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Serial != nil {
		updates["serial"] = *in.Serial
	}
	if in.Cost != nil {
		updates["cost"] = *in.Cost
	}
	if in.Price != nil {
		updates["price"] = *in.Price
	}
	if in.WholesalePrice != nil {
		updates["wholesale_price"] = *in.WholesalePrice
	}
	if in.Discount != nil {
		updates["discount"] = *in.Discount
	}
	if in.Quantity != nil {
		q := *in.Quantity
		if q < 0 {
			q = 0
		}
		updates["quantity"] = q
	}
	if in.CategoryID != nil {
		updates["category_id"] = *in.CategoryID
	}
	if in.BranchID != nil {
		updates["branch_id"] = *in.BranchID
	}
	if in.Status != nil {
		updates["status"] = *in.Status
	}
	if in.ShowOnSite != nil {
		updates["show_on_site"] = *in.ShowOnSite
	}
	if in.CardImage != nil {
		updates["card_image"] = *in.CardImage
	}
	if in.Gallery != nil {
		updates["gallery"] = *in.Gallery
	}
	if t, ok := parseDayPtr(in.CreatedAt); ok {
		updates["created_at"] = t
	}
	if len(updates) > 0 {
		db.Model(&item).Updates(updates)
	}
	// satılma tarixi — cihazın satışını yenilə
	if t, ok := parseDayPtr(in.SoldAt); ok {
		db.Model(&Sale{}).Where("item_id = ?", item.ID).Update("sold_at", t)
	}
	if in.Values != nil {
		db.Where("item_id = ?", item.ID).Delete(&ItemAttributeValue{})
		for _, v := range *in.Values {
			db.Create(&ItemAttributeValue{ItemID: item.ID, AttributeID: v.AttributeID, Value: v.Value})
		}
	}
	if in.Translations != nil {
		saveItemTranslations(item.ID, *in.Translations)
	}
	db.Preload("Category").Preload("Branch").Preload("Values.Attribute").Preload("Translations").First(&item, item.ID)
	writeJSON(w, 200, item)
}

// DELETE /api/items/{id}
func deleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	db.Where("item_id = ?", id).Delete(&ItemAttributeValue{})
	db.Where("item_id = ?", id).Delete(&ItemTranslation{})
	db.Delete(&Item{}, id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// PUT /api/customers/{id}
func updateCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct{ Name, Phone string }
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	idn, _ := strconv.Atoi(id)
	if ex := customerByPhone(in.Phone, uint(idn)); ex != nil {
		writeJSON(w, 409, map[string]any{"error": "Bu nömrə artıq var: " + ex.Name, "existing_id": ex.ID, "existing_name": ex.Name})
		return
	}
	db.Model(&Customer{}).Where("id = ?", id).Updates(map[string]any{"name": in.Name, "phone": in.Phone})
	var c Customer
	db.First(&c, id)
	writeJSON(w, 200, c)
}

// PUT /api/sales/{id} — satışı redaktə et (qiymət dəyişəndə mənfəət yenidən hesablanır)
func updateSale(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var sale Sale
	if err := db.First(&sale, id).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "satış tapılmadı"})
		return
	}
	var item Item
	db.First(&item, sale.ItemID)
	var in struct {
		SalePrice      *float64 `json:"sale_price"`
		Channel        *string  `json:"channel"`
		WarrantyMonths *int     `json:"warranty_months"`
		CustomerID     *uint    `json:"customer_id"`
		SoldAt         *string  `json:"sold_at"` // satılma tarixi (YYYY-MM-DD)
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	if in.SalePrice != nil {
		sale.SalePrice = *in.SalePrice
		sale.Profit = *in.SalePrice - item.Cost
	}
	if in.Channel != nil {
		sale.Channel = *in.Channel
	}
	if in.WarrantyMonths != nil {
		sale.WarrantyMonths = *in.WarrantyMonths
	}
	if in.CustomerID != nil {
		sale.CustomerID = in.CustomerID
	}
	if t, ok := parseDayPtr(in.SoldAt); ok {
		sale.SoldAt = t
	}
	db.Save(&sale)
	db.Preload("Item").Preload("Item.Category").Preload("Item.Branch").Preload("Customer").First(&sale, sale.ID)
	writeJSON(w, 200, sale)
}
