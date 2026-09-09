package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// GET /api/calc?from=YYYY-MM-DD&to=YYYY-MM-DD&branches=1,2&categories=Notebook,Telefon
// Seçilmiş aralıq + filial(lar) + kateqoriya(lar) üzrə dövriyyə, mənfəət,
// xərclər və net qazancı hesablayır. Yalnız "sayılan" satışlar (bağlanmamış
// kredit daxil deyil) — bağlanmamış kredit ayrıca "credit_pending" kimi qayıdır.
func calcHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	from := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	if v := q.Get("from"); v != "" {
		if t, e := time.ParseInLocation("2006-01-02", v, time.Local); e == nil {
			from = t
		}
	}
	toEnd := time.Now().AddDate(0, 0, 1)
	if v := q.Get("to"); v != "" {
		if t, e := time.ParseInLocation("2006-01-02", v, time.Local); e == nil {
			toEnd = t.AddDate(0, 0, 1) // "to" günü daxil olsun
		}
	}

	var branchIDs []uint
	for _, s := range strings.Split(q.Get("branches"), ",") {
		if n, e := strconv.Atoi(strings.TrimSpace(s)); e == nil && n > 0 {
			branchIDs = append(branchIDs, uint(n))
		}
	}
	// satıcı (admin deyil) yalnız öz filialını hesablaya bilər — seçim məcburi öz filialı
	if bid, restricted := branchScope(r); restricted {
		branchIDs = []uint{bid}
	}
	var cats []string
	for _, s := range strings.Split(q.Get("categories"), ",") {
		if t := strings.TrimSpace(s); t != "" {
			cats = append(cats, t)
		}
	}

	// filtrləri tətbiq edən köməkçi (counted parametri ilə)
	build := func(counted bool) *gorm.DB {
		g := db.Model(&Sale{}).
			Joins("JOIN items ON items.id = sales.item_id").
			Where("sales.counted = ? AND sales.sold_at >= ? AND sales.sold_at < ?", counted, from, toEnd)
		if len(branchIDs) > 0 {
			g = g.Where("sales.branch_id IN ?", branchIDs)
		}
		if len(cats) > 0 {
			g = g.Joins("JOIN categories ON categories.id = items.category_id").
				Where("categories.name IN ?", cats)
		}
		return g
	}

	type agg struct {
		Turnover float64
		Profit   float64
		Cost     float64
		Count    int64
	}
	var a agg
	build(true).Select(
		"COALESCE(SUM(sales.sale_price),0) as turnover, " +
			"COALESCE(SUM(sales.profit),0) as profit, " +
			"COALESCE(SUM(items.cost),0) as cost, " +
			"COUNT(*) as count").Scan(&a)

	// bağlanmamış kredit (əlimizə gəlməyən pul) — ayrıca
	var creditPending struct {
		Sum   float64
		Count int64
	}
	build(false).Where("sales.channel = ?", "credit").
		Select("COALESCE(SUM(sales.sale_price),0) as sum, COUNT(*) as count").Scan(&creditPending)

	// xərclər (aralıqda) — filial/kateqoriyaya bağlı deyil
	var expenses float64
	db.Model(&Expense{}).Where("date >= ? AND date < ?", from, toEnd).
		Select("COALESCE(SUM(amount),0)").Scan(&expenses)

	writeJSON(w, 200, map[string]any{
		"turnover":             a.Turnover,
		"profit":               a.Profit,
		"cost":                 a.Cost,
		"count":                a.Count,
		"expenses":             expenses,
		"net":                  a.Profit - expenses, // əlimizdə olmalı xalis
		"credit_pending":       creditPending.Sum,
		"credit_pending_count": creditPending.Count,
	})
}
