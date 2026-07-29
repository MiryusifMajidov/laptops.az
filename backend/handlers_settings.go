package main

import "net/http"

// müştəri saytının oxuya biləcəyi açıq tənzimləmə açarları
var publicSettingKeys = []string{"store_lat", "store_lng", "featured_item_id", "tile_notebook_img", "tile_desktop_img"}

func getSetting(key string) string {
	var s Setting
	if err := db.Where("key = ?", key).First(&s).Error; err == nil {
		return s.Value
	}
	return ""
}

func setSetting(key, value string) {
	var s Setting
	if err := db.Where("key = ?", key).First(&s).Error; err == nil {
		db.Model(&s).Update("value", value)
	} else {
		db.Create(&Setting{Key: key, Value: value})
	}
}

// ilkin dəyərlər (mağaza koordinatı) — boşdursa
func seedSettings() {
	if getSetting("store_lat") == "" {
		setSetting("store_lat", "40.3811247")
	}
	if getSetting("store_lng") == "" {
		setSetting("store_lng", "49.8474406")
	}
	// plitə fon şəkilləri — yalnız açar heç yaranmayıbsa (admin təmizləsə boş qalmalıdır)
	var c int64
	db.Model(&Setting{}).Where("key = ?", "tile_notebook_img").Count(&c)
	if c == 0 {
		setSetting("tile_notebook_img", "/tile-notebook.svg")
	}
	db.Model(&Setting{}).Where("key = ?", "tile_desktop_img").Count(&c)
	if c == 0 {
		setSetting("tile_desktop_img", "/tile-desktop.svg")
	}
}

// GET /api/public/settings — saytın oxuyacağı açıq tənzimləmələr
func publicSettings(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	for _, k := range publicSettingKeys {
		m[k] = getSetting(k)
	}
	writeJSON(w, 200, m)
}

// GET /api/settings — admin: bütün tənzimləmələr
func listSettings(w http.ResponseWriter, r *http.Request) {
	var all []Setting
	db.Find(&all)
	m := map[string]string{}
	for _, s := range all {
		m[s.Key] = s.Value
	}
	writeJSON(w, 200, m)
}

// PUT /api/settings {key: value, ...} — toplu yaz
func saveSettings(w http.ResponseWriter, r *http.Request) {
	var in map[string]string
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	for k, v := range in {
		if k != "" {
			setSetting(k, v)
		}
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
