package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// SEO — məhsul səhifələrinə serverdə meta inject (Google + sosial media üçün) + sitemap + robots.

var reTitle = regexp.MustCompile(`(?is)<title>.*?</title>`)
var reDesc = regexp.MustCompile(`(?is)<meta name="description"[^>]*>`)

type seoMeta struct{ Title, Desc, Image, URL, JSONLD string }

func absURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return "https://laptops.az/logo-solid.png"
	}
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u
	}
	if !strings.HasPrefix(u, "/") {
		u = "/" + u
	}
	return "https://laptops.az" + u
}

// productMeta — /mehsul/{id} üçün SEO meta qurur (məhsul stokda + saytda görünürsə)
func productMeta(id string) (seoMeta, bool) {
	var it Item
	if db.Preload("Category").Preload("Values.Attribute").
		Where("show_on_site = ? AND status = ?", true, "in_stock").First(&it, id).Error != nil {
		return seoMeta{}, false
	}
	price := it.Price
	if it.Discount > 0 && it.Discount < it.Price {
		price = it.Price - it.Discount
	}
	brand := ""
	for _, v := range it.Values {
		if v.Attribute.Name == "Marka" {
			brand = v.Value
			break
		}
	}
	u := fmt.Sprintf("https://laptops.az/mehsul/%d", it.ID)
	img := absURL(it.CardImage)
	title := fmt.Sprintf("%s — ₼%.0f | Laptops.az", it.Name, price)
	desc := it.Name
	if it.Category.Name != "" {
		desc += " · " + it.Category.Name
	}
	desc += ". Laptops.az-da onlayn seç, mağazada ödə. Zəmanət və çatdırılma. Bakı."
	ld := map[string]any{
		"@context": "https://schema.org/", "@type": "Product",
		"name": it.Name, "image": img, "description": desc, "url": u,
		"offers": map[string]any{
			"@type": "Offer", "price": fmt.Sprintf("%.0f", price),
			"priceCurrency": "AZN", "availability": "https://schema.org/InStock", "url": u,
		},
	}
	if it.Serial != "" {
		ld["sku"] = it.Serial
	}
	if brand != "" {
		ld["brand"] = map[string]any{"@type": "Brand", "name": brand}
	}
	if it.Category.Name != "" {
		ld["category"] = it.Category.Name
	}
	b, _ := json.Marshal(ld) // encoding/json <, >, & simvollarını escape edir → <script>-ə təhlükəsiz
	return seoMeta{Title: title, Desc: desc, Image: img, URL: u, JSONLD: string(b)}, true
}

// injectSEO — index.html-ə məhsulun meta məlumatını yerləşdirir
func injectSEO(base []byte, m seoMeta) []byte {
	s := string(base)
	s = reTitle.ReplaceAllString(s, "<title>"+html.EscapeString(m.Title)+"</title>")
	s = reDesc.ReplaceAllString(s, `<meta name="description" content="`+html.EscapeString(m.Desc)+`">`)
	extra := `<meta property="og:type" content="product">` +
		`<meta property="og:title" content="` + html.EscapeString(m.Title) + `">` +
		`<meta property="og:description" content="` + html.EscapeString(m.Desc) + `">` +
		`<meta property="og:image" content="` + html.EscapeString(m.Image) + `">` +
		`<meta property="og:url" content="` + html.EscapeString(m.URL) + `">` +
		`<meta property="og:site_name" content="Laptops.az">` +
		`<meta name="twitter:card" content="summary_large_image">` +
		`<link rel="canonical" href="` + html.EscapeString(m.URL) + `">`
	if m.JSONLD != "" {
		extra += `<script type="application/ld+json">` + m.JSONLD + `</script>`
	}
	return []byte(strings.Replace(s, "</head>", extra+"</head>", 1))
}

// GET /robots.txt
func robotsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "User-agent: *\nAllow: /\nDisallow: /sirab/\nDisallow: /api/\n\nSitemap: https://laptops.az/sitemap.xml\n")
}

// GET /sitemap.xml — ana + kateqoriyalar + bütün stokdakı məhsullar
func sitemapHandler(w http.ResponseWriter, r *http.Request) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	add := func(loc, pri string) {
		b.WriteString("  <url><loc>" + html.EscapeString(loc) + "</loc><priority>" + pri + "</priority></url>\n")
	}
	add("https://laptops.az/", "1.0")
	add("https://laptops.az/mehsullar", "0.9")
	var cats []Category
	db.Find(&cats)
	for _, c := range cats {
		add("https://laptops.az/kateqoriya/"+url.PathEscape(c.Name), "0.8")
	}
	var items []Item
	db.Where("show_on_site = ? AND status = ?", true, "in_stock").Order("created_at desc").Find(&items)
	for _, it := range items {
		add(fmt.Sprintf("https://laptops.az/mehsul/%d", it.ID), "0.7")
	}
	b.WriteString("</urlset>\n")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(b.String()))
}
