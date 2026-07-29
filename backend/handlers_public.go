package main

import (
	"fmt"
	"net/http"
	"time"
)

// GET /api/public/products?category=&q=  — YALNIZ saytda göstərilən + stokda olan mallar (auth yoxdur)
func publicProducts(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	q := db.Preload("Category").Preload("Branch").Preload("Values.Attribute").Preload("Translations").
		Where("show_on_site = ? AND status = ?", true, "in_stock").Order("created_at desc")
	if c := r.URL.Query().Get("category"); c != "" && c != "all" {
		q = q.Joins("JOIN categories ON categories.id = items.category_id").Where("categories.name = ?", c)
	}
	if term := r.URL.Query().Get("q"); term != "" {
		like := "%" + term + "%"
		// ad VƏ YA xüsusiyyət dəyəri (marka, RAM, ekran kartı və s.) üzrə axtar
		q = q.Where("items.name LIKE ? OR items.id IN (?)", like,
			db.Model(&ItemAttributeValue{}).Select("item_id").Where("value LIKE ?", like))
	}
	var items []Item
	q.Find(&items)
	def := defaultLangCode()
	for i := range items {
		applyItemLang(&items[i], lang, def) // adı seçilmiş dilə çevir + daxili qiymətləri gizlə
	}
	writeJSON(w, 200, items)
}

// GET /api/public/products/{id}
func publicProduct(w http.ResponseWriter, r *http.Request) {
	var it Item
	if err := db.Preload("Category").Preload("Branch").Preload("Values.Attribute").Preload("Translations").
		Where("show_on_site = ?", true).First(&it, r.PathValue("id")).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "məhsul tapılmadı"})
		return
	}
	applyItemLang(&it, r.URL.Query().Get("lang"), defaultLangCode())
	writeJSON(w, 200, it)
}

// GET /api/public/categories — saytda göstərilən malların kateqoriyaları
func publicCategories(w http.ResponseWriter, r *http.Request) {
	type row struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	var rows []row
	db.Model(&Item{}).Select("categories.name as name, count(*) as count").
		Joins("JOIN categories ON categories.id = items.category_id").
		Where("items.show_on_site = ? AND items.status = ?", true, "in_stock").
		Group("categories.name").Order("count desc").Scan(&rows)
	writeJSON(w, 200, rows)
}

// POST /api/public/orders — saytdan sifariş (ödəniş yoxdur, 24 saat rezerv)
func publicOrder(w http.ResponseWriter, r *http.Request) {
	if ok, info := orderLimiter.allow(clientIP(r)); !ok {
		lang := r.URL.Query().Get("lang")
		msg := fmt.Sprintf(pick(map[string]string{
			"az": "Çox sayda sifariş göndərildi. Zəhmət olmasa %s sonra yenidən cəhd edin.",
			"ru": "Отправлено слишком много заказов. Пожалуйста, повторите через %s.",
			"tr": "Çok fazla sipariş gönderildi. Lütfen %s sonra tekrar deneyin.",
			"en": "Too many orders sent. Please try again in %s.",
		}, lang), humanWait(lang, info.Wait))
		writeJSON(w, 429, map[string]string{"error": msg})
		return
	}
	var in struct {
		CustomerName string `json:"customer_name"`
		Phone        string `json:"phone"`
		ItemID       uint   `json:"item_id"`
		Note         string `json:"note"`
	}
	if err := decodeBody(r, &in); err != nil || in.CustomerName == "" || in.ItemID == 0 {
		writeJSON(w, 400, map[string]string{"error": "ad və məhsul vacibdir"})
		return
	}
	o := OnlineOrder{
		CustomerName: in.CustomerName, Phone: in.Phone, ItemID: in.ItemID, Note: in.Note,
		Status: "pending", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	db.Create(&o)
	writeJSON(w, 201, map[string]any{"id": o.ID, "ref": fmt.Sprintf("LA-%06d", o.ID)})
}
