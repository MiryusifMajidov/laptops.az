package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// clientIP — həqiqi müştəri IP-si (Fly / Cloudflare proksisini nəzərə alır)
func clientIP(r *http.Request) string {
	for _, h := range []string{"Cf-Connecting-Ip", "Fly-Client-Ip", "X-Forwarded-For"} {
		if v := r.Header.Get(h); v != "" {
			if i := strings.IndexByte(v, ','); i > 0 {
				return strings.TrimSpace(v[:i])
			}
			return strings.TrimSpace(v)
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// limitInfo — hansı limit aşıldı və nə qədər gözləmək lazımdır
type limitInfo struct {
	Scope string // "minute" | "hour" | "day"
	Max   int64
	Wait  int64 // saniyə
}

type ipEntry struct {
	minStart, minCount int64
	hrStart, hrCount   int64
	last               int64
}

// rateLimiter — IP üzrə (dəqiqə + saat) və qlobal günlük tavan
type rateLimiter struct {
	mu       sync.Mutex
	ips      map[string]*ipEntry
	perMin   int64
	perHour  int64
	dayMax   int64
	dayCount int64
	dayStart int64
	calls    int
}

func newLimiter(perMin, perHour, dayMax int) *rateLimiter {
	return &rateLimiter{ips: map[string]*ipEntry{}, perMin: int64(perMin), perHour: int64(perHour), dayMax: int64(dayMax)}
}

// allow — icazə varmı? İcazə yoxdursa hansı limitin aşıldığını qaytarır.
func (rl *rateLimiter) allow(ip string) (bool, limitInfo) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now().Unix()

	if now-rl.dayStart >= 86400 {
		rl.dayStart = now
		rl.dayCount = 0
	}
	if rl.dayMax > 0 && rl.dayCount >= rl.dayMax {
		return false, limitInfo{"day", rl.dayMax, 86400 - (now - rl.dayStart)}
	}

	// arabir köhnə IP-ləri təmizlə (yaddaş şişməsin)
	rl.calls++
	if rl.calls%500 == 0 {
		for k, e := range rl.ips {
			if now-e.last > 3600 {
				delete(rl.ips, k)
			}
		}
	}

	e := rl.ips[ip]
	if e == nil {
		e = &ipEntry{minStart: now, hrStart: now}
		rl.ips[ip] = e
	}
	if now-e.minStart >= 60 {
		e.minStart = now
		e.minCount = 0
	}
	if now-e.hrStart >= 3600 {
		e.hrStart = now
		e.hrCount = 0
	}
	if e.minCount >= rl.perMin {
		return false, limitInfo{"minute", rl.perMin, 60 - (now - e.minStart)}
	}
	if e.hrCount >= rl.perHour {
		return false, limitInfo{"hour", rl.perHour, 3600 - (now - e.hrStart)}
	}

	e.minCount++
	e.hrCount++
	e.last = now
	rl.dayCount++
	return true, limitInfo{}
}

// AI köməkçi — env ilə tənzimlənə bilər (AI_RL_MIN / AI_RL_HOUR / AI_RL_DAY)
var aiLimiter = newLimiter(envInt("AI_RL_MIN", 15), envInt("AI_RL_HOUR", 120), envInt("AI_RL_DAY", 1200))

// Saytdan sifariş — daha sərt (fake sifariş spam-ına qarşı)
var orderLimiter = newLimiter(3, 12, 300)

// ---- limit mesajı (çoxdilli) ----

func pick(m map[string]string, lang string) string {
	if v, ok := m[lang]; ok {
		return v
	}
	return m["az"]
}

func humanWait(lang string, sec int64) string {
	if sec < 1 {
		sec = 1
	}
	if sec >= 60 {
		mins := (sec + 59) / 60
		return fmt.Sprintf("%d %s", mins, pick(map[string]string{"az": "dəqiqə", "ru": "мин", "tr": "dakika", "en": "min"}, lang))
	}
	return fmt.Sprintf("%d %s", sec, pick(map[string]string{"az": "saniyə", "ru": "сек", "tr": "saniye", "en": "sec"}, lang))
}

// rateLimitMsg — AI köməkçi üçün konkret, çoxdilli limit mesajı
func rateLimitMsg(lang string, info limitInfo) string {
	wait := humanWait(lang, info.Wait)
	switch info.Scope {
	case "day":
		return pick(map[string]string{
			"az": "AI köməkçi bu gün üçün çox yükləndi. Zəhmət olmasa sabah yenidən yoxlayın.",
			"ru": "AI-помощник перегружен на сегодня. Пожалуйста, попробуйте завтра.",
			"tr": "AI asistan bugünlük doldu. Lütfen yarın tekrar deneyin.",
			"en": "The AI assistant is at capacity for today. Please try again tomorrow.",
		}, lang)
	case "hour":
		return fmt.Sprintf(pick(map[string]string{
			"az": "Saatlıq limitə çatdınız (saatda %d mesaj). Zəhmət olmasa %s sonra yenidən yazın.",
			"ru": "Вы достигли часового лимита (%d сообщений в час). Пожалуйста, повторите через %s.",
			"tr": "Saatlik sınıra ulaştınız (saatte %d mesaj). Lütfen %s sonra tekrar yazın.",
			"en": "You've reached the hourly limit (%d messages/hour). Please try again in %s.",
		}, lang), info.Max, wait)
	default: // minute
		return fmt.Sprintf(pick(map[string]string{
			"az": "Dəqiqədə %d mesaj limitinə çatdınız. Zəhmət olmasa %s sonra yenidən yazın.",
			"ru": "Вы достигли лимита %d сообщений в минуту. Пожалуйста, повторите через %s.",
			"tr": "Dakikada %d mesaj sınırına ulaştınız. Lütfen %s sonra tekrar yazın.",
			"en": "You've reached the limit of %d messages per minute. Please try again in %s.",
		}, lang), info.Max, wait)
	}
}
