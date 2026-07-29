package main

import (
	"net/http"
	"strings"
	"time"
)

// ---- PUBLIC: bildiriş lenti (yeni məhsullar) ----

type notifItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CardImage string    `json:"card_image"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

// GET /api/public/notifications — son əlavə olunan məhsullar (mobil zəng/bildiriş üçün)
func publicNotifications(w http.ResponseWriter, r *http.Request) {
	var items []Item
	db.Preload("Translations").Where("show_on_site = ? AND status = ?", true, "in_stock").
		Order("created_at desc").Limit(30).Find(&items)
	lang := r.URL.Query().Get("lang")
	def := defaultLangCode()
	out := make([]notifItem, 0, len(items))
	for i := range items {
		applyItemLang(&items[i], lang, def) // adı seçilmiş dilə çevir
		it := items[i]
		out = append(out, notifItem{ID: it.ID, Name: it.Name, CardImage: it.CardImage, Price: it.Price, CreatedAt: it.CreatedAt})
	}
	writeJSON(w, 200, out)
}

// ---- PUBLIC: tərəfdaşlıq müraciəti (qeydiyyat əvəzi) ----

// POST /api/public/partner-apply {name, store_name, phone}
func publicPartnerApply(w http.ResponseWriter, r *http.Request) {
	if ok, info := orderLimiter.allow(clientIP(r)); !ok {
		writeJSON(w, 429, map[string]string{"error": "Çox sayda müraciət — " + humanWait("az", info.Wait) + " sonra yenidən."})
		return
	}
	var in struct {
		Name      string `json:"name"`
		StoreName string `json:"store_name"`
		Phone     string `json:"phone"`
	}
	if err := decodeBody(r, &in); err != nil || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Phone) == "" {
		writeJSON(w, 400, map[string]string{"error": "ad və telefon vacibdir"})
		return
	}
	a := PartnerApplication{
		Name: strings.TrimSpace(in.Name), StoreName: strings.TrimSpace(in.StoreName),
		Phone: strings.TrimSpace(in.Phone), Status: "pending", CreatedAt: time.Now(),
	}
	db.Create(&a)
	writeJSON(w, 201, map[string]bool{"ok": true})
}

// ---- PARTNER: optavoy qiymətli məhsullar (token tələb olunur) ----

// partnerə görünüş: adı dilə çevir, alışı gizlət, qiyməti = optavoy
func applyPartnerLang(it *Item, lang, def string) {
	if lang != "" && lang != def {
		for _, t := range it.Translations {
			if t.Lang == lang && t.Name != "" {
				it.Name = t.Name
				break
			}
		}
	}
	it.Translations = nil
	it.Cost = 0
	if it.WholesalePrice > 0 {
		it.Price = it.WholesalePrice // partner optavoy qiyməti görür
	}
}

func partnerAuth(r *http.Request) bool {
	u, ok := userFromReq(r)
	return ok && (u.Role == "partner" || u.Role == "admin")
}

// GET /api/partner/products?category=&q=&lang=
func partnerProducts(w http.ResponseWriter, r *http.Request) {
	if !partnerAuth(r) {
		writeJSON(w, 403, map[string]string{"error": "partner girişi tələb olunur"})
		return
	}
	lang := r.URL.Query().Get("lang")
	q := db.Preload("Category").Preload("Branch").Preload("Values.Attribute").Preload("Translations").
		Where("show_on_site = ? AND status = ?", true, "in_stock").Order("created_at desc")
	if c := r.URL.Query().Get("category"); c != "" && c != "all" {
		q = q.Joins("JOIN categories ON categories.id = items.category_id").Where("categories.name = ?", c)
	}
	if term := r.URL.Query().Get("q"); term != "" {
		like := "%" + term + "%"
		q = q.Where("items.name LIKE ? OR items.id IN (?)", like,
			db.Model(&ItemAttributeValue{}).Select("item_id").Where("value LIKE ?", like))
	}
	var items []Item
	q.Find(&items)
	def := defaultLangCode()
	for i := range items {
		applyPartnerLang(&items[i], lang, def)
	}
	writeJSON(w, 200, items)
}

// GET /api/partner/products/{id}?lang=
func partnerProduct(w http.ResponseWriter, r *http.Request) {
	if !partnerAuth(r) {
		writeJSON(w, 403, map[string]string{"error": "partner girişi tələb olunur"})
		return
	}
	var it Item
	if err := db.Preload("Category").Preload("Branch").Preload("Values.Attribute").Preload("Translations").
		Where("show_on_site = ?", true).First(&it, r.PathValue("id")).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "məhsul tapılmadı"})
		return
	}
	applyPartnerLang(&it, r.URL.Query().Get("lang"), defaultLangCode())
	writeJSON(w, 200, it)
}

// ---- ADMIN: tərəfdaşlıq müraciətləri ----

func listPartnerApplications(w http.ResponseWriter, r *http.Request) {
	var apps []PartnerApplication
	db.Order("created_at desc").Find(&apps)
	writeJSON(w, 200, apps)
}

// PUT /api/partner-applications/{id} {status}
func updatePartnerApplication(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status string `json:"status"`
	}
	if err := decodeBody(r, &in); err != nil || in.Status == "" {
		writeJSON(w, 400, map[string]string{"error": "status vacibdir"})
		return
	}
	db.Model(&PartnerApplication{}).Where("id = ?", r.PathValue("id")).Update("status", in.Status)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---- ADMIN: istifadəçi idarəçiliyi (partner/satıcı hesabı yaratmaq) ----

func adminOnly(r *http.Request) bool {
	u, ok := userFromReq(r)
	return ok && u.Role == "admin"
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	if !adminOnly(r) {
		writeJSON(w, 403, map[string]string{"error": "yalnız admin"})
		return
	}
	var users []User
	db.Order("id").Find(&users) // PassHash json:"-" olduğu üçün cavaba düşmür
	writeJSON(w, 200, users)
}

// POST /api/users {username, password, name, role}
func createUser(w http.ResponseWriter, r *http.Request) {
	if !adminOnly(r) {
		writeJSON(w, 403, map[string]string{"error": "yalnız admin"})
		return
	}
	var in struct {
		Username, Password, Name, Role string
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	uname := strings.ToLower(strings.TrimSpace(in.Username))
	if uname == "" || len(in.Password) < 4 {
		writeJSON(w, 400, map[string]string{"error": "istifadəçi adı və ən az 4 simvol parol vacibdir"})
		return
	}
	role := in.Role
	if role != "admin" && role != "user" && role != "partner" {
		role = "partner"
	}
	var n int64
	db.Model(&User{}).Where("username = ?", uname).Count(&n)
	if n > 0 {
		writeJSON(w, 409, map[string]string{"error": "bu istifadəçi adı artıq var"})
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = uname
	}
	u := User{Username: uname, PassHash: hashPassword(in.Password), Role: role, Name: name}
	db.Create(&u)
	writeJSON(w, 201, u)
}

// DELETE /api/users/{id}
func deleteUser(w http.ResponseWriter, r *http.Request) {
	if !adminOnly(r) {
		writeJSON(w, 403, map[string]string{"error": "yalnız admin"})
		return
	}
	var u User
	if err := db.First(&u, r.PathValue("id")).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "tapılmadı"})
		return
	}
	if u.Username == "admin" {
		writeJSON(w, 409, map[string]string{"error": "əsas admin silinə bilməz"})
		return
	}
	db.Delete(&User{}, r.PathValue("id"))
	writeJSON(w, 200, map[string]bool{"ok": true})
}
