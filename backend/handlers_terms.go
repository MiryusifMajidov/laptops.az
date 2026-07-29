package main

import (
	"fmt"
	"net/http"
)

// Kataloq terminləri (kateqoriya adı, xüsusiyyət adı, seçim dəyəri) üçün tərcümə.
// Sayt AZ dəyəri məntiq üçün saxlayır; yalnız GÖSTƏRİŞ tt(azText) ilə tərcümə olunur.

type termRow struct {
	ID     uint              `json:"id"`
	Kind   string            `json:"kind"`
	Az     string            `json:"az"`
	Values map[string]string `json:"values"`
}

// bütün AZ termləri: (kind, refID, az)
func allTerms() []struct {
	Kind string
	ID   uint
	Az   string
} {
	var out []struct {
		Kind string
		ID   uint
		Az   string
	}
	var cats []Category
	db.Order("name").Find(&cats)
	for _, c := range cats {
		out = append(out, struct {
			Kind string
			ID   uint
			Az   string
		}{"category", c.ID, c.Name})
	}
	var attrs []Attribute
	db.Order("name").Find(&attrs)
	for _, a := range attrs {
		out = append(out, struct {
			Kind string
			ID   uint
			Az   string
		}{"attribute", a.ID, a.Name})
	}
	var opts []AttributeOption
	db.Order("attribute_id, value").Find(&opts)
	for _, o := range opts {
		out = append(out, struct {
			Kind string
			ID   uint
			Az   string
		}{"option", o.ID, o.Value})
	}
	return out
}

func termKey(kind string, id uint) string { return fmt.Sprintf("%s:%d", kind, id) }

// GET /api/public/terms?lang=ru — {azText: tərcümə} (yalnız dolu tərcümələr)
func publicTerms(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	lang := r.URL.Query().Get("lang")
	def := defaultLangCode()
	if lang == "" || lang == def {
		writeJSON(w, 200, m)
		return
	}
	var trs []TermTranslation
	db.Where("lang = ?", lang).Find(&trs)
	byRef := map[string]string{}
	for _, t := range trs {
		if t.Value != "" {
			byRef[termKey(t.Kind, t.RefID)] = t.Value
		}
	}
	for _, t := range allTerms() {
		if v, ok := byRef[termKey(t.Kind, t.ID)]; ok {
			m[t.Az] = v
		}
	}
	writeJSON(w, 200, m)
}

// GET /api/terms — admin: bütün terminlər dil üzrə dəyərlərlə (kateqoriya/xüsusiyyət/dəyər)
func listTerms(w http.ResponseWriter, r *http.Request) {
	var langs []Language
	db.Order("sort").Find(&langs)
	var trs []TermTranslation
	db.Find(&trs)
	idx := map[string]map[string]string{}
	for _, t := range trs {
		k := termKey(t.Kind, t.RefID)
		if idx[k] == nil {
			idx[k] = map[string]string{}
		}
		idx[k][t.Lang] = t.Value
	}
	cats := []termRow{}
	attrs := []termRow{}
	opts := []termRow{}
	for _, t := range allTerms() {
		vals := idx[termKey(t.Kind, t.ID)]
		if vals == nil {
			vals = map[string]string{}
		}
		row := termRow{ID: t.ID, Kind: t.Kind, Az: t.Az, Values: vals}
		switch t.Kind {
		case "category":
			cats = append(cats, row)
		case "attribute":
			attrs = append(attrs, row)
		case "option":
			opts = append(opts, row)
		}
	}
	writeJSON(w, 200, map[string]any{"languages": langs, "categories": cats, "attributes": attrs, "options": opts})
}

func upsertTerm(kind string, refID uint, lang, value string) {
	var t TermTranslation
	if err := db.Where("kind = ? AND ref_id = ? AND lang = ?", kind, refID, lang).First(&t).Error; err == nil {
		db.Model(&t).Update("value", value)
	} else {
		db.Create(&TermTranslation{Kind: kind, RefID: refID, Lang: lang, Value: value})
	}
}

// PUT /api/terms {entries:[{kind,ref_id,lang,value}]}
func saveTerms(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Entries []struct {
			Kind  string `json:"kind"`
			RefID uint   `json:"ref_id"`
			Lang  string `json:"lang"`
			Value string `json:"value"`
		} `json:"entries"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	for _, e := range in.Entries {
		if e.Kind == "" || e.RefID == 0 || e.Lang == "" {
			continue
		}
		upsertTerm(e.Kind, e.RefID, e.Lang, e.Value)
	}
	writeJSON(w, 200, map[string]any{"ok": true, "count": len(in.Entries)})
}

// POST /api/terms/translate {to, overwrite?} — AZ terminləri AI ilə tərcümə (markaları saxlayır)
func translateTerms(w http.ResponseWriter, r *http.Request) {
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
	var existing []TermTranslation
	db.Where("lang = ?", in.To).Find(&existing)
	have := map[string]string{}
	for _, t := range existing {
		have[termKey(t.Kind, t.RefID)] = t.Value
	}
	type ref struct {
		Kind string
		ID   uint
	}
	var refs []ref
	var texts []string
	for _, t := range allTerms() {
		if t.Az == "" {
			continue
		}
		if !in.Overwrite && have[termKey(t.Kind, t.ID)] != "" {
			continue
		}
		refs = append(refs, ref{t.Kind, t.ID})
		texts = append(texts, t.Az)
	}
	if len(texts) == 0 {
		writeJSON(w, 200, map[string]any{"translated": 0, "note": "tərcümə ediləcək termin yoxdur (hamısı doludur)"})
		return
	}
	translated := 0
	const batch = 50
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
			upsertTerm(refs[i+j].Kind, refs[i+j].ID, in.To, val)
			translated++
		}
	}
	writeJSON(w, 200, map[string]any{"translated": translated})
}
