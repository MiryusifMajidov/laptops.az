//go:build prod

package main

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Prod build zamanı frontend build-ləri binary-ə embed olunur.
// Docker (və ya lokal prod test) go build-dən əvvəl bu qovluqları doldurur:
//   backend/web/admin ← frontend/dist   (Vite base '/admin/')
//   backend/web/site  ← website/dist    (Vite base '/')
//
//go:embed all:web/admin
var adminFiles embed.FS

//go:embed all:web/site
var siteFiles embed.FS

// spaHandler — tələb olunan statik faylı verir; tapılmasa index.html qaytarır (SPA fallback).
func spaHandler(root fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(root))
	index, _ := fs.ReadFile(root, "index.html")
	serveIndex := func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(index)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "" || p == "." {
			serveIndex(w)
			return
		}
		if f, err := root.Open(p); err == nil {
			info, _ := f.Stat()
			f.Close()
			if info != nil && info.IsDir() {
				http.Error(w, "forbidden", http.StatusForbidden) // qovluq siyahısı qadağandır
				return
			}
			fileServer.ServeHTTP(w, r)
			return
		}
		serveIndex(w) // naməlum yol → SPA öz routing-ini etsin
	})
}

// mountStatic — admin paneli /sirab altında, müştəri saytını kökdə verir.
// /api/* və /uploads/* daha spesifik olduğu üçün onlara toxunmur (Go 1.22 mux prioriteti).
func mountStatic(mux *http.ServeMux) {
	adminSub, _ := fs.Sub(adminFiles, "web/admin")
	siteSub, _ := fs.Sub(siteFiles, "web/site")
	mux.Handle("/sirab/", http.StripPrefix("/sirab", spaHandler(adminSub)))
	mux.Handle("/sirab", http.RedirectHandler("/sirab/", http.StatusMovedPermanently))
	mux.Handle("/", siteHandler(siteSub))
}

// siteHandler — müştəri saytı: statik fayl, SPA fallback + /mehsul/{id} üçün SEO meta inject.
func siteHandler(root fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(root))
	index, _ := fs.ReadFile(root, "index.html")
	serveHTML := func(w http.ResponseWriter, b []byte) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(b)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "" || p == "." {
			serveHTML(w, index)
			return
		}
		if f, err := root.Open(p); err == nil { // real statik fayl
			info, _ := f.Stat()
			f.Close()
			if info != nil && info.IsDir() {
				http.Error(w, "forbidden", http.StatusForbidden) // qovluq siyahısı qadağandır
				return
			}
			fileServer.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(p, "mehsul/") { // məhsul səhifəsi → SEO meta
			if m, ok := productMeta(strings.TrimPrefix(p, "mehsul/")); ok {
				serveHTML(w, injectSEO(index, m))
				return
			}
		}
		serveHTML(w, index) // digər SPA yolları
	})
}
