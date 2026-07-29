package main

import (
	"net/http"
	"time"
)

// Visit — müştəri saytına bir giriş (client JS "/api/public/visit" çağırır).
type Visit struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	IP        string    `gorm:"index" json:"ip"`
	Path      string    `json:"path"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// POST /api/public/visit — sayt yüklənəndə çağırılır. IP + səhifə + brauzer loglanır.
// Yalnız JS icra edən real brauzerlər çağırır → botların çoxu avtomatik süzülür.
func trackVisit(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Path string `json:"path"`
	}
	_ = decodeBody(r, &in)
	ip := clientIP(r)
	// təkrar-atışları (React remount və s.) süz — eyni IP son 5 saniyədə
	var recent int64
	db.Model(&Visit{}).Where("ip = ? AND created_at > ?", ip, time.Now().Add(-5*time.Second)).Count(&recent)
	if recent > 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	ua := r.Header.Get("User-Agent")
	if len(ua) > 300 {
		ua = ua[:300]
	}
	path := in.Path
	if len(path) > 200 {
		path = path[:200]
	}
	db.Create(&Visit{IP: ip, Path: path, UserAgent: ua, CreatedAt: time.Now()})
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/visits — admin: statistika + son ziyarətlər.
func listVisits(w http.ResponseWriter, r *http.Request) {
	var total, unique, today int64
	db.Model(&Visit{}).Count(&total)
	db.Model(&Visit{}).Distinct("ip").Count(&unique)

	loc, err := time.LoadLocation("Asia/Baku")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	startAZ := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	db.Model(&Visit{}).Where("created_at >= ?", startAZ.UTC()).Count(&today)

	var list []Visit
	db.Order("created_at desc").Limit(1000).Find(&list)
	writeJSON(w, 200, map[string]any{
		"total": total, "unique_ips": unique, "today": today, "visits": list,
	})
}
