package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	db.Preload("Attributes.Options", optOrder).First(&c, id)
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
		Name        string `json:"name"`
		Multiselect *bool  `json:"multiselect"`
		ShowOnSite  *bool  `json:"show_on_site"`
	}
	if err := decodeBody(r, &in); err != nil || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "ad vacibdir"})
		return
	}
	updates := map[string]any{"name": in.Name}
	if in.Multiselect != nil {
		updates["multiselect"] = *in.Multiselect
	}
	if in.ShowOnSite != nil {
		updates["show_on_site"] = *in.ShowOnSite
	}
	db.Model(&Attribute{}).Where("id = ?", id).Updates(updates)
	var a Attribute
	db.Preload("Options", optOrder).First(&a, id)
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

// PUT /api/attributes/{id}/options/order  {ids: [sıralı option id-ləri]}
// Option-ları verilmiş ardıcıllıqla sıralayır (position = sıra nömrəsi). Data zədələnmir — yalnız sıra.
func reorderOptions(w http.ResponseWriter, r *http.Request) {
	attrID := r.PathValue("id")
	var in struct {
		IDs []uint `json:"ids"`
	}
	if err := decodeBody(r, &in); err != nil || len(in.IDs) == 0 {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	for i, oid := range in.IDs {
		db.Model(&AttributeOption{}).Where("id = ? AND attribute_id = ?", oid, attrID).Update("position", i+1)
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// PUT /api/attribute-options/{id}  {value}
// Option dəyərini dəyişir VƏ məhsullardakı köhnə dəyəri də yeni dəyərə keçirir (data zədələnmir).
// Əgər yeni dəyər həmin xüsusiyyətdə artıq başqa option kimi mövcuddursa → BİRLƏŞDİRİR
// (məhsullar yeni dəyərə keçir, bu təkrar option silinir). Beləliklə "eyni olub fərqli
// yazılan" seçimləri birləşdirmək olur.
func updateOption(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Value string `json:"value"`
	}
	if err := decodeBody(r, &in); err != nil || strings.TrimSpace(in.Value) == "" {
		writeJSON(w, 400, map[string]string{"error": "dəyər vacibdir"})
		return
	}
	newVal := strings.TrimSpace(in.Value)
	var opt AttributeOption
	if err := db.First(&opt, r.PathValue("id")).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "option tapılmadı"})
		return
	}
	oldVal := opt.Value
	if oldVal == newVal {
		writeJSON(w, 200, opt)
		return
	}
	// 1) məhsullarda bu xüsusiyyətin köhnə dəyərini işlədən BÜTÜN qeydləri yeni dəyərə keçir
	updated := db.Model(&ItemAttributeValue{}).
		Where("attribute_id = ? AND value = ?", opt.AttributeID, oldVal).
		Update("value", newVal).RowsAffected
	// 2) yeni dəyər artıq başqa option kimi varsa → birləşdir (bu təkrarı sil); yoxsa adını dəyiş
	var dup int64
	db.Model(&AttributeOption{}).Where("attribute_id = ? AND value = ? AND id <> ?", opt.AttributeID, newVal, opt.ID).Count(&dup)
	merged := dup > 0
	if merged {
		db.Delete(&AttributeOption{}, opt.ID)
	} else {
		db.Model(&AttributeOption{}).Where("id = ?", opt.ID).Update("value", newVal)
	}
	writeJSON(w, 200, map[string]any{"ok": true, "old": oldVal, "new": newVal, "merged": merged, "items_updated": updated})
}
