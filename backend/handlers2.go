package main

import (
	"net/http"
	"time"
)

// ---- Onlayn sifarişlər ----

// GET /api/orders
func listOrders(w http.ResponseWriter, r *http.Request) {
	var orders []OnlineOrder
	db.Preload("Item").Preload("Item.Branch").Where("status = ?", "pending").Order("created_at desc").Find(&orders)
	writeJSON(w, 200, orders)
}

// POST /api/orders/{id}/convert — cihazı rezerv et, sifarişi bağla
func convertOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var o OnlineOrder
	if err := db.First(&o, id).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "sifariş tapılmadı"})
		return
	}
	db.Model(&o).Update("status", "converted")
	db.Model(&Item{}).Where("id = ?", o.ItemID).Update("status", "reserved")
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// POST /api/orders/{id}/cancel
func cancelOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	db.Model(&OnlineOrder{}).Where("id = ?", id).Update("status", "cancelled")
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---- Kassa ----

// son bağlanış vaxtı — yoxdursa günün başlanğıcı
func lastCloseTime() time.Time {
	var c DailyClose
	if err := db.Order("date desc").First(&c).Error; err != nil {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}
	return c.Date
}

// son bağlanışdan bəri kanal üzrə cəm (açıq kassa). branchID != 0 → yalnız o filial.
func channelTotalsSince(since time.Time, branchID uint) (map[string]float64, float64) {
	type row struct {
		Channel string
		Sum     float64
	}
	var rows []row
	q := db.Model(&Sale{}).Select("channel, COALESCE(SUM(sale_price),0) as sum").
		Where("sold_at > ? AND counted = ?", since, true)
	if branchID != 0 {
		q = q.Where("branch_id = ?", branchID)
	}
	q.Group("channel").Scan(&rows)
	m := map[string]float64{"cash": 0, "card": 0, "installment": 0, "credit": 0}
	var total float64
	for _, r := range rows {
		m[r.Channel] = r.Sum
		total += r.Sum
	}
	return m, total
}

// GET /api/kassa — açıq sessiya (son bağlanışdan bəri)
func kassaHandler(w http.ResponseWriter, r *http.Request) {
	since := lastCloseTime()
	var kb uint // satıcı → öz filialı; admin → 0 (hamısı)
	if bid, restricted := branchScope(r); restricted {
		kb = bid
	}
	m, total := channelTotalsSince(since, kb)
	var closes []DailyClose
	db.Order("date desc").Limit(10).Find(&closes)
	writeJSON(w, 200, map[string]any{
		"today":  map[string]any{"cash": m["cash"], "card": m["card"], "installment": m["installment"], "credit": m["credit"], "total": total},
		"since":  since,
		"closes": closes,
	})
}

// POST /api/kassa/close — açıq kassanı bağla → sıfırlanır (yenidən bağlamaq mümkün deyil)
func createClose(w http.ResponseWriter, r *http.Request) {
	since := lastCloseTime()
	m, total := channelTotalsSince(since, 0)
	if total == 0 {
		writeJSON(w, 400, map[string]string{"error": "Bağlanacaq satış yoxdur — kassa artıq sıfırdır"})
		return
	}
	c := DailyClose{Date: time.Now(), Cash: m["cash"], Card: m["card"], Installment: m["installment"], Credit: m["credit"], Total: total}
	db.Create(&c)
	writeJSON(w, 201, c)
}

// ---- Təchizat ----

// GET /api/supplies
func listSupplies(w http.ResponseWriter, r *http.Request) {
	var s []SupplyBatch
	db.Order("date desc").Find(&s)
	writeJSON(w, 200, s)
}

// ---- Hesabatlar ----

type topRow struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
type perfRow struct {
	Name   string  `json:"name"`
	Profit float64 `json:"profit"`
}
type deadRow struct {
	Name    string  `json:"name"`
	Cost    float64 `json:"cost"`
	AgeDays int     `json:"age_days"`
}

// GET /api/reports
func reportsHandler(w http.ResponseWriter, r *http.Request) {
	bid, restricted := branchScope(r) // satıcı → yalnız öz filialı

	var top []topRow
	topQ := db.Model(&Sale{}).Select("items.name as name, count(*) as count").
		Joins("JOIN items ON items.id = sales.item_id").
		Where("sales.counted = ?", true)
	if restricted {
		topQ = topQ.Where("sales.branch_id = ?", bid)
	}
	topQ.Group("items.name").Order("count desc").Limit(5).Scan(&top)

	var perf []perfRow
	perfQ := db.Model(&Sale{}).Select("branches.name as name, COALESCE(SUM(sales.profit),0) as profit").
		Joins("JOIN branches ON branches.id = sales.branch_id").
		Where("sales.counted = ?", true)
	if restricted {
		perfQ = perfQ.Where("sales.branch_id = ?", bid)
	}
	perfQ.Group("branches.name").Order("profit desc").Scan(&perf)

	var inStock []Item // ölü stok — anbar bütün filiallara açıq (qlobal)
	db.Where("status = ?", "in_stock").Find(&inStock)
	dead := []deadRow{}
	for _, it := range inStock {
		age := int(time.Since(it.CreatedAt).Hours() / 24)
		if age > 90 {
			dead = append(dead, deadRow{Name: it.Name, Cost: it.Cost, AgeDays: age})
		}
	}

	writeJSON(w, 200, map[string]any{
		"top_models":  top,
		"branch_perf": perf,
		"dead_stock":  dead,
	})
}
