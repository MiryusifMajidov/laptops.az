package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Xarici "laptops" saytı üçün ayrıca public API.
// Yalnız: Mərkəz filial (is_main) + saytda aktiv (show_on_site) + stokda (in_stock) + "Notebook" kateqoriya.
// Daxili qiymətlər (alış/topdan) verilmir. Şəkillər tam URL kimi ötürülür. Auth yoxdur (/api/public/*).
// (absURL — seo.go-da; nisbi /uploads/… → https://laptops.az/uploads/…)

// laptopDTO — xarici sayt üçün təmiz, sabit cavab forması (daxili modeldən asılı deyil).
type laptopDTO struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Price      float64   `json:"price"`       // sayt qiyməti (₼)
	Discount   float64   `json:"discount"`    // endirim (₼)
	FinalPrice float64   `json:"final_price"` // price − discount
	Image      string    `json:"image"`       // əsas şəkil (tam URL)
	Images     []string  `json:"images"`      // bütün şəkillər (tam URL)
	Category   string    `json:"category"`
	Specs      []specKV  `json:"specs"` // xüsusiyyətlər (Marka, Prosessor, RAM…)
	CreatedAt  time.Time `json:"created_at"`
}

type specKV struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// toLaptopDTO — applyItemLang-dən sonra Item-i təmiz DTO-ya çevirir (şəkillər tam URL).
func toLaptopDTO(it Item) laptopDTO {
	d := laptopDTO{
		ID: it.ID, Name: it.Name, Price: it.Price, Discount: it.Discount,
		FinalPrice: it.Price - it.Discount, Category: it.Category.Name, CreatedAt: it.CreatedAt,
	}
	if d.FinalPrice < 0 {
		d.FinalPrice = 0
	}
	// şəkillər: əsas (card) + qalereya, təkrarsız, tam URL
	seen := map[string]bool{}
	add := func(u string) {
		if strings.TrimSpace(u) == "" {
			return
		}
		a := absURL(u)
		if !seen[a] {
			seen[a] = true
			d.Images = append(d.Images, a)
		}
	}
	add(it.CardImage)
	if it.Gallery != "" {
		var raw []string
		if json.Unmarshal([]byte(it.Gallery), &raw) == nil {
			for _, u := range raw {
				add(u)
			}
		}
	}
	d.Image = absURL(it.CardImage) // əsas şəkil (şəkil yoxdursa logo qaytarır)
	for _, v := range it.Values {
		d.Specs = append(d.Specs, specKV{Name: v.Attribute.Name, Value: v.Value})
	}
	return d
}

// laptopScope — 4 şərti bir sorğuya tətbiq edir (siyahı və detal üçün ortaq).
func laptopScope(base *gorm.DB) *gorm.DB {
	return base.
		Joins("JOIN categories ON categories.id = items.category_id").
		Joins("JOIN branches ON branches.id = items.branch_id").
		Where("items.show_on_site = ? AND items.status = ?", true, "in_stock").
		Where("branches.is_main = ?", true).
		Where("categories.name = ?", "Notebook")
}

// GET /api/public/laptops?lang=&q=&limit=&offset=
func laptopsList(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	q := laptopScope(db.Preload("Category").Preload("Branch").Preload("Values.Attribute").Preload("Translations")).
		Order("items.created_at desc")

	// axtarış: ad VƏ YA xüsusiyyət dəyəri (marka, RAM, ekran kartı…)
	if term := r.URL.Query().Get("q"); term != "" {
		like := "%" + term + "%"
		q = q.Where("items.name LIKE ? OR items.id IN (?)", like,
			db.Model(&ItemAttributeValue{}).Select("item_id").Where("value LIKE ?", like))
	}
	// könüllü səhifələmə
	if n, _ := strconv.Atoi(r.URL.Query().Get("limit")); n > 0 {
		q = q.Limit(n)
		if off, _ := strconv.Atoi(r.URL.Query().Get("offset")); off > 0 {
			q = q.Offset(off)
		}
	}

	var items []Item
	q.Find(&items)
	def := defaultLangCode()
	out := make([]laptopDTO, 0, len(items))
	for i := range items {
		applyItemLang(&items[i], lang, def) // adı dilə çevir + sayt-gizli atributları çıxar
		out = append(out, toLaptopDTO(items[i]))
	}
	writeJSON(w, 200, out)
}

// GET /api/public/laptops/{id}
func laptopDetail(w http.ResponseWriter, r *http.Request) {
	var it Item
	err := laptopScope(db.Preload("Category").Preload("Branch").Preload("Values.Attribute").Preload("Translations")).
		First(&it, "items.id = ?", r.PathValue("id")).Error
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "laptop tapılmadı"})
		return
	}
	applyItemLang(&it, r.URL.Query().Get("lang"), defaultLangCode())
	writeJSON(w, 200, toLaptopDTO(it))
}
