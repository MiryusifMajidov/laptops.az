package main

import (
	"fmt"
	"net/http"
	"strconv"
)

// PUT /api/categories/{id}  {name, attribute_ids}
func updateCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	var in struct {
		Name         string `json:"name"`
		AttributeIDs []uint `json:"attribute_ids"`
	}
	if err := decodeBody(r, &in); err != nil || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "ad vacibdir"})
		return
	}
	db.Model(&Category{}).Where("id = ?", id).Update("name", in.Name)
	db.Exec("DELETE FROM category_attributes WHERE category_id = ?", id)
	for _, aid := range in.AttributeIDs {
		db.Exec("INSERT INTO category_attributes (category_id, attribute_id) VALUES (?, ?)", id, aid)
	}
	var c Category
	db.Preload("Attributes.Options").First(&c, id)
	writeJSON(w, 200, c)
}

// DELETE /api/categories/{id}
func deleteCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var n int64
	db.Model(&Item{}).Where("category_id = ?", id).Count(&n)
	if n > 0 {
		writeJSON(w, 409, map[string]string{"error": fmt.Sprintf("%d məhsul bu kateqoriyadadır — əvvəl onları dəyişin", n)})
		return
	}
	db.Exec("DELETE FROM category_attributes WHERE category_id = ?", id)
	db.Delete(&Category{}, id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// PUT /api/attributes/{id}  {name}
func updateAttribute(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Name string `json:"name"`
	}
	if err := decodeBody(r, &in); err != nil || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "ad vacibdir"})
		return
	}
	db.Model(&Attribute{}).Where("id = ?", id).Update("name", in.Name)
	var a Attribute
	db.Preload("Options").First(&a, id)
	writeJSON(w, 200, a)
}

// DELETE /api/attributes/{id}  (option-ları, kateqoriya bağlarını və məhsul dəyərlərini də silir)
func deleteAttribute(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	db.Where("attribute_id = ?", id).Delete(&AttributeOption{})
	db.Where("attribute_id = ?", id).Delete(&ItemAttributeValue{})
	db.Exec("DELETE FROM category_attributes WHERE attribute_id = ?", id)
	db.Delete(&Attribute{}, id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// DELETE /api/attribute-options/{id}
func deleteOption(w http.ResponseWriter, r *http.Request) {
	db.Delete(&AttributeOption{}, r.PathValue("id"))
	writeJSON(w, 200, map[string]bool{"ok": true})
}
