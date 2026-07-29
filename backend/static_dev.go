//go:build !prod

package main

import "net/http"

// Dev rejimi: statik UI verilmir — frontend-lər Vite dev server-dən gəlir (5173/5174).
// Prod build (`-tags prod`) static_prod.go-nu işlədir və UI-ı embed edir.
func mountStatic(mux *http.ServeMux) {}
