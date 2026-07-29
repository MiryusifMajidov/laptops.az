package main

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

// audit detalı üçün oxunaqlı sahə adları (AZ)
var auditLabels = map[string]string{
	"name": "Ad", "serial": "Seriya", "price": "Qiymət", "cost": "Alış",
	"sale_price": "Satış qiyməti", "discount": "Endirim", "wholesale_price": "Topdan",
	"channel": "Kanal", "amount": "Məbləğ", "status": "Status", "note": "Qeyd",
	"customer_name": "Müştəri", "phone": "Telefon", "category_id": "Kateqoriya",
	"branch_id": "Filial", "item_id": "Mal ID", "item_name": "Mal", "store_name": "Mağaza",
	"total": "Cəmi", "paid": "Ödənilən", "down_payment": "İlkin ödəniş",
	"warranty_months": "Zəmanət (ay)", "next_due": "Növbəti ödəniş", "date": "Tarix",
	"item_count": "Say", "total_cost": "Ümumi dəyər", "source": "Mənbə", "supplier": "Təchizatçı",
	"emails": "Ünvanlar", "show_on_site": "Saytda", "closed": "Bağlandı", "given_price": "Verilmə qiyməti",
	"is_default": "Əsas dil", "enabled": "Aktiv", "code": "Kod", "address": "Ünvan",
	"created_at": "Alınma tarixi", "sold_at": "Satılma tarixi", "order_id": "Sifariş",
}

// gizli sahələr — audit-ə heç vaxt düşməsin (parol və s.)
var auditSkip = map[string]bool{"old": true, "new": true, "password": true, "pass": true, "token": true}

// auditDetail — göndərilən JSON body-dən oxunaqlı xülasə qurur ("Satış qiyməti: 1500 · Kanal: cash").
// JSON deyilsə (məs. şəkil yükləmə) və ya boşdursa — boş qaytarır.
func auditDetail(bodyBytes []byte) string {
	if len(bodyBytes) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(bodyBytes, &m); err != nil {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		if auditSkip[k] {
			continue
		}
		s := auditScalar(m[k])
		if s == "" {
			continue // massiv/obyekt/boş dəyər — atla
		}
		label := auditLabels[k]
		if label == "" {
			label = k
		}
		parts = append(parts, label+": "+s)
	}
	detail := strings.Join(parts, " · ")
	if len(detail) > 600 {
		detail = detail[:600] + "…"
	}
	return detail
}

// auditScalar — yalnız sadə dəyərləri mətnə çevirir (massiv/obyekt üçün boş)
func auditScalar(v any) string {
	switch x := v.(type) {
	case string:
		if len(x) > 90 {
			return x[:90] + "…"
		}
		return x
	case bool:
		if x {
			return "bəli"
		}
		return "xeyr"
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return "" // array / object / null
	}
}
