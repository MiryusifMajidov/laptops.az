package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// əsas dilin kodu (adətən az)
func defaultLangCode() string {
	var dl Language
	if err := db.Where("is_default = ?", true).First(&dl).Error; err == nil && dl.Code != "" {
		return dl.Code
	}
	return "az"
}

func sectionOf(key string) string {
	if i := strings.IndexByte(key, '.'); i > 0 {
		return key[:i]
	}
	return key
}

// itemTr — məhsul adının bir dildəki tərcüməsi (create/update input-u)
type itemTr struct {
	Lang string `json:"lang"`
	Name string `json:"name"`
}

// saveItemTranslations — məhsulun tərcümələrini yenidən yazır (əsas dil və boşlar atılır).
func saveItemTranslations(itemID uint, trs []itemTr) {
	def := defaultLangCode()
	db.Where("item_id = ?", itemID).Delete(&ItemTranslation{})
	for _, t := range trs {
		lang := strings.TrimSpace(t.Lang)
		name := strings.TrimSpace(t.Name)
		if lang == "" || lang == def || name == "" {
			continue
		}
		db.Create(&ItemTranslation{ItemID: itemID, Lang: lang, Name: name})
	}
}

// applyItemLang — PUBLIC üçün: adı seçilmiş dilə çevirir (yoxdursa əsas dil qalır),
// daxili sahələri (alış, topdan) gizlədir və tərcümə siyahısını təmizləyir.
// def — əsas dil kodu (çağıran bir dəfə hesablayıb ötürür).
func applyItemLang(it *Item, lang, def string) {
	if lang != "" && lang != def {
		for _, t := range it.Translations {
			if t.Lang == lang && t.Name != "" {
				it.Name = t.Name
				break
			}
		}
	}
	it.Translations = nil
	it.Cost = 0
	it.WholesalePrice = 0
}

// ---- PUBLIC (sayt üçün, auth yoxdur) ----

// GET /api/public/languages — saytda göstəriləcək aktiv dillər (sıralı)
func publicLanguages(w http.ResponseWriter, r *http.Request) {
	var langs []Language
	db.Where("enabled = ?", true).Order("sort").Find(&langs)
	writeJSON(w, 200, langs)
}

// GET /api/public/i18n?lang=ru — həmin dil üçün mətn lüğəti (əsas dilə fallback ilə)
func publicI18n(w http.ResponseWriter, r *http.Request) {
	def := defaultLangCode()
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = def
	}
	m := map[string]string{}
	var base []UiString
	db.Where("lang = ?", def).Find(&base)
	for _, s := range base {
		m[s.Key] = s.Value
	}
	if lang != def {
		var loc []UiString
		db.Where("lang = ?", lang).Find(&loc)
		for _, s := range loc {
			if s.Value != "" {
				m[s.Key] = s.Value // yalnız dolu tərcümə fallback-i üstələyir
			}
		}
	}
	writeJSON(w, 200, m)
}

// ---- ADMIN: dillər ----

// GET /api/languages — hamısı (deaktiv daxil)
func listLanguages(w http.ResponseWriter, r *http.Request) {
	var langs []Language
	db.Order("sort").Find(&langs)
	writeJSON(w, 200, langs)
}

// POST /api/languages {code,name,icon,enabled}
func createLanguage(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Code, Name, Icon string
		Enabled          bool
	}
	if err := decodeBody(r, &in); err != nil || strings.TrimSpace(in.Code) == "" || strings.TrimSpace(in.Name) == "" {
		writeJSON(w, 400, map[string]string{"error": "kod və ad vacibdir"})
		return
	}
	code := strings.ToLower(strings.TrimSpace(in.Code))
	var ex int64
	db.Model(&Language{}).Where("code = ?", code).Count(&ex)
	if ex > 0 {
		writeJSON(w, 409, map[string]string{"error": "bu dil kodu artıq var"})
		return
	}
	var maxSort int
	db.Model(&Language{}).Select("COALESCE(MAX(sort),0)").Scan(&maxSort)
	l := Language{Code: code, Name: strings.TrimSpace(in.Name), Icon: in.Icon, Enabled: in.Enabled, Sort: maxSort + 1}
	db.Create(&l)
	writeJSON(w, 201, l)
}

// PUT /api/languages/{id} {name,icon,enabled,is_default,sort}
func updateLanguage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var l Language
	if err := db.First(&l, id).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "tapılmadı"})
		return
	}
	var in struct {
		Name      *string `json:"name"`
		Icon      *string `json:"icon"`
		Enabled   *bool   `json:"enabled"`
		IsDefault *bool   `json:"is_default"`
		Sort      *int    `json:"sort"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	upd := map[string]any{}
	if in.Name != nil {
		upd["name"] = *in.Name
	}
	if in.Icon != nil {
		upd["icon"] = *in.Icon
	}
	if in.Sort != nil {
		upd["sort"] = *in.Sort
	}
	if in.Enabled != nil {
		upd["enabled"] = *in.Enabled
	}
	if in.IsDefault != nil && *in.IsDefault {
		db.Model(&Language{}).Where("is_default = ?", true).Update("is_default", false)
		upd["is_default"] = true
		upd["enabled"] = true // əsas dil həmişə aktivdir
	}
	if len(upd) > 0 {
		db.Model(&l).Updates(upd)
	}
	db.First(&l, id)
	writeJSON(w, 200, l)
}

// DELETE /api/languages/{id} — dili və onun tərcümələrini sil (əsas dil silinə bilməz)
func deleteLanguage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var l Language
	if err := db.First(&l, id).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "tapılmadı"})
		return
	}
	if l.IsDefault {
		writeJSON(w, 409, map[string]string{"error": "əsas dili silmək olmaz"})
		return
	}
	db.Where("lang = ?", l.Code).Delete(&UiString{})
	db.Delete(&Language{}, id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---- ADMIN: statik mətnlər ----

// GET /api/ui-strings — bütün açarlar (kanonik sırada), hər dil üzrə dəyər
func listUiStrings(w http.ResponseWriter, r *http.Request) {
	var langs []Language
	db.Order("sort").Find(&langs)
	var all []UiString
	db.Find(&all)
	byKey := map[string]map[string]string{}
	for _, s := range all {
		if byKey[s.Key] == nil {
			byKey[s.Key] = map[string]string{}
		}
		byKey[s.Key][s.Lang] = s.Value
	}
	type row struct {
		Key     string            `json:"key"`
		Section string            `json:"section"`
		Values  map[string]string `json:"values"`
	}
	rows := []row{}
	seen := map[string]bool{}
	for _, e := range defaultUiStrings() {
		seen[e.Key] = true
		vals := byKey[e.Key]
		if vals == nil {
			vals = map[string]string{}
		}
		rows = append(rows, row{Key: e.Key, Section: sectionOf(e.Key), Values: vals})
	}
	for k, vals := range byKey {
		if !seen[k] {
			rows = append(rows, row{Key: k, Section: sectionOf(k), Values: vals})
		}
	}
	writeJSON(w, 200, map[string]any{"languages": langs, "strings": rows})
}

// PUT /api/ui-strings {entries:[{key,lang,value}]} — toplu upsert
func saveUiStrings(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Entries []struct {
			Key   string `json:"key"`
			Lang  string `json:"lang"`
			Value string `json:"value"`
		} `json:"entries"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	for _, e := range in.Entries {
		if e.Key == "" || e.Lang == "" {
			continue
		}
		upsertUiString(e.Key, e.Lang, e.Value)
	}
	writeJSON(w, 200, map[string]any{"ok": true, "count": len(in.Entries)})
}

func upsertUiString(key, lang, value string) {
	var s UiString
	if err := db.Where("key = ? AND lang = ?", key, lang).First(&s).Error; err == nil {
		db.Model(&s).Update("value", value)
	} else {
		db.Create(&UiString{Key: key, Lang: lang, Value: value})
	}
}

// POST /api/ui-strings/translate {to, overwrite?} — əsas dildən "to" dilinə AI ilə tərcümə
func translateUiStrings(w http.ResponseWriter, r *http.Request) {
	var in struct {
		To        string `json:"to"`
		Overwrite bool   `json:"overwrite"`
	}
	if err := decodeBody(r, &in); err != nil || in.To == "" {
		writeJSON(w, 400, map[string]string{"error": "hədəf dil vacibdir"})
		return
	}
	def := defaultLangCode()
	if in.To == def {
		writeJSON(w, 400, map[string]string{"error": "əsas dili tərcümə etmək lazım deyil"})
		return
	}
	var base, existing []UiString
	db.Where("lang = ?", def).Find(&base)
	db.Where("lang = ?", in.To).Find(&existing)
	have := map[string]string{}
	for _, s := range existing {
		have[s.Key] = s.Value
	}
	var keys, texts []string
	for _, s := range base {
		if s.Value == "" {
			continue
		}
		if !in.Overwrite && have[s.Key] != "" {
			continue // artıq tərcümə var — saxla
		}
		keys = append(keys, s.Key)
		texts = append(texts, s.Value)
	}
	if len(texts) == 0 {
		writeJSON(w, 200, map[string]any{"translated": 0, "note": "tərcümə ediləcək mətn yoxdur (hamısı doludur)"})
		return
	}
	translated := 0
	const batch = 40
	for i := 0; i < len(texts); i += batch {
		end := i + batch
		if end > len(texts) {
			end = len(texts)
		}
		out, err := geminiTranslateBatch(def, in.To, texts[i:end])
		if err != nil {
			writeJSON(w, 502, map[string]string{"error": "AI tərcümə alınmadı: " + err.Error()})
			return
		}
		if len(out) != end-i {
			writeJSON(w, 502, map[string]string{"error": "AI tərcümə uzunluğu uyğun gəlmədi"})
			return
		}
		for j, val := range out {
			upsertUiString(keys[i+j], in.To, val)
			translated++
		}
	}
	writeJSON(w, 200, map[string]any{"translated": translated})
}

// POST /api/translate {text, from?, targets?} — bir mətni hədəf dillərə AI ilə tərcümə
// (məhsul adı və s. üçün). from boşdursa əsas dil; targets boşdursa bütün aktiv qeyri-əsas dillər.
func translateText(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text    string   `json:"text"`
		From    string   `json:"from"`
		Targets []string `json:"targets"`
	}
	if err := decodeBody(r, &in); err != nil || strings.TrimSpace(in.Text) == "" {
		writeJSON(w, 400, map[string]string{"error": "mətn yoxdur"})
		return
	}
	from := in.From
	if from == "" {
		from = defaultLangCode()
	}
	targets := in.Targets
	if len(targets) == 0 {
		var langs []Language
		db.Where("enabled = ? AND is_default = ?", true, false).Order("sort").Find(&langs)
		for _, l := range langs {
			targets = append(targets, l.Code)
		}
	}
	out := map[string]string{}
	for _, to := range targets {
		if to == "" || to == from {
			continue
		}
		res, err := geminiTranslateBatch(from, to, []string{in.Text})
		if err != nil {
			writeJSON(w, 502, map[string]string{"error": "AI tərcümə alınmadı: " + err.Error()})
			return
		}
		if len(res) == 1 {
			out[to] = res[0]
		}
	}
	writeJSON(w, 200, map[string]any{"translations": out})
}

// geminiTranslateBatch — mətn siyahısını fromCode dilindən toCode dilinə tərcümə edir (JSON massiv).
func geminiTranslateBatch(fromCode, toCode string, texts []string) ([]string, error) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("AI açarı (GEMINI_API_KEY) qoyulmayıb")
	}
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-flash-latest"
	}
	payload, _ := json.Marshal(texts)
	prompt := fmt.Sprintf(`You are translating UI strings of a laptop/electronics shop website from %s to %s.
Rules:
- Return ONLY a JSON array of strings, the SAME length and order as the input array.
- Keep emojis, punctuation, line breaks (\n), currency symbols and brand/product names (HP, ASUS, Apple, Lenovo, iPhone, RTX, SSD, RAM…) unchanged.
- Natural, concise retail/marketing tone. No quotes around items, no explanations.
Input JSON array:
%s`, langNameEN(fromCode), langNameEN(toCode), string(payload))

	reqBody := map[string]any{
		"contents": []any{
			map[string]any{"role": "user", "parts": []any{map[string]any{"text": prompt}}},
		},
		"generationConfig": map[string]any{
			"temperature":      0.2,
			"responseMimeType": "application/json",
			"maxOutputTokens":  4096,
			// Qeyd: thinkingConfig göndərilmir — cari flash modelləri (2026) onu rədd edir (400)
		},
	}
	bb, _ := json.Marshal(reqBody)
	url := "https://generativelanguage.googleapis.com/v1beta/models/" + model + ":generateContent"
	httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(bb))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", key)
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gemini status %d", resp.StatusCode)
	}
	var gr struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &gr); err != nil || len(gr.Candidates) == 0 {
		return nil, fmt.Errorf("boş cavab")
	}
	var out string
	for _, p := range gr.Candidates[0].Content.Parts {
		out += p.Text
	}
	var result []string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		return nil, fmt.Errorf("tərcümə formatı yanlış")
	}
	return result, nil
}
