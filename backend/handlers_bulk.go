package main

import "net/http"

// POST /api/items/publish-all — stokda olan bütün malları saytda göstər (toplu).
// Import zamanı ShowOnSite=false gəlir; bu, sahibin bir kliklə hamısını yayımlaması üçündür.
// Sahib sonra ayrı-ayrı malları admin paneldən gizlədə bilər.
func publishAllItems(w http.ResponseWriter, r *http.Request) {
	res := db.Model(&Item{}).Where("status = ?", "in_stock").Update("show_on_site", true)
	if res.Error != nil {
		writeJSON(w, 500, map[string]string{"error": res.Error.Error()})
		return
	}
	writeJSON(w, 200, map[string]int64{"published": res.RowsAffected})
}
