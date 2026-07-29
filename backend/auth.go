package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Sessiya müddəti — 7 gün (istifadəçi "ən azı 6 saat" istədi; 7 gün rahat ödəyir).
const sessionTTL = 7 * 24 * time.Hour

type sessEntry struct {
	u   User
	exp time.Time
}

// Yaddaş keşi (sürət üçün) + DB-də saxlanır (Session modeli) → restart-dan sağ çıxır.
var sessions = struct {
	sync.RWMutex
	m map[string]sessEntry
}{m: map[string]sessEntry{}}

func newToken() string {
	b := make([]byte, 24)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashPassword(p string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(h)
}

func checkPassword(hash, p string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) == nil
}

func bearer(r *http.Request) string {
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}

func userFromReq(r *http.Request) (User, bool) {
	tok := bearer(r)
	if tok == "" {
		return User{}, false
	}
	now := time.Now()
	sessions.RLock()
	e, ok := sessions.m[tok]
	sessions.RUnlock()
	if ok {
		if now.After(e.exp) {
			sessions.Lock()
			delete(sessions.m, tok)
			sessions.Unlock()
			db.Where("token = ?", tok).Delete(&Session{})
			return User{}, false
		}
		return e.u, true
	}
	// yaddaşda yoxdur (server restart/auto-stop olub) → DB-dən bərpa et
	var s Session
	if err := db.Where("token = ?", tok).First(&s).Error; err != nil {
		return User{}, false
	}
	if now.After(s.ExpiresAt) {
		db.Where("token = ?", tok).Delete(&Session{})
		return User{}, false
	}
	u := User{ID: s.UserID, Username: s.Username, Role: s.Role, Name: s.Name}
	sessions.Lock()
	sessions.m[tok] = sessEntry{u, s.ExpiresAt}
	sessions.Unlock()
	return u, true
}

// POST /api/login
func login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Username, Password string }
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	var u User
	uname := strings.ToLower(strings.TrimSpace(in.Username))
	if err := db.Where("username = ?", uname).First(&u).Error; err != nil || !checkPassword(u.PassHash, in.Password) {
		writeJSON(w, 401, map[string]string{"error": "İstifadəçi adı və ya parol yanlışdır"})
		return
	}
	tok := newToken()
	exp := time.Now().Add(sessionTTL)
	sessions.Lock()
	sessions.m[tok] = sessEntry{u, exp}
	sessions.Unlock()
	db.Create(&Session{Token: tok, UserID: u.ID, Username: u.Username, Role: u.Role, Name: u.Name, ExpiresAt: exp})
	db.Where("expires_at < ?", time.Now()).Delete(&Session{}) // vaxtı keçmişləri təmizlə
	writeJSON(w, 200, map[string]any{"token": tok, "role": u.Role, "name": u.Name, "username": u.Username})
}

// POST /api/logout
func logout(w http.ResponseWriter, r *http.Request) {
	tok := bearer(r)
	sessions.Lock()
	delete(sessions.m, tok)
	sessions.Unlock()
	db.Where("token = ?", tok).Delete(&Session{})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// POST /api/change-password  {old, new}
func changePassword(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromReq(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "giriş tələb olunur"})
		return
	}
	var in struct{ Old, New string }
	if err := decodeBody(r, &in); err != nil || len(in.New) < 3 {
		writeJSON(w, 400, map[string]string{"error": "yeni parol ən az 3 simvol olmalıdır"})
		return
	}
	var full User
	db.First(&full, u.ID)
	if !checkPassword(full.PassHash, in.Old) {
		writeJSON(w, 401, map[string]string{"error": "köhnə parol yanlışdır"})
		return
	}
	db.Model(&full).Update("pass_hash", hashPassword(in.New))
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// GET /api/audit — YALNIZ admin
func listAudit(w http.ResponseWriter, r *http.Request) {
	u, _ := userFromReq(r)
	if u.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "yalnız admin görə bilər"})
		return
	}
	var logs []AuditLog
	db.Order("created_at desc").Limit(300).Find(&logs)
	writeJSON(w, 200, logs)
}

// auth middleware — login/health/uploads açıqdır, qalanı token tələb edir; write-lar audit olunur
func authMiddleware(next http.Handler) http.Handler {
	open := map[string]bool{"/api/login": true, "/api/health": true}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		// Yalnız /api/* qorunur (public istisna). Statik UI (/, /admin, /assets…) və /uploads açıqdır.
		if r.Method == http.MethodOptions || open[p] || !strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/api/public/") {
			next.ServeHTTP(w, r)
			return
		}
		u, ok := userFromReq(r)
		if !ok {
			writeJSON(w, 401, map[string]string{"error": "giriş tələb olunur"})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			// JSON body-ni oxu (detal üçün) və handler-ə çatması üçün geri qoy
			var bodyBytes []byte
			if r.Body != nil && strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
			db.Create(&AuditLog{Username: u.Username, Role: u.Role, Action: r.Method, Path: p, Detail: auditDetail(bodyBytes), CreatedAt: time.Now()})
		}
		next.ServeHTTP(w, r)
	})
}
