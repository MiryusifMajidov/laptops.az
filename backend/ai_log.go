package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// AiConversation — müştəri AI köməkçisi ilə bir söhbət (tam yazışma JSON kimi).
// Client hər sorğuda sabit conversation_id + source ("web"/"app") göndərir; hər turda upsert olunur.
type AiConversation struct {
	ID        string    `gorm:"primaryKey" json:"id"` // client tərəfindən yaradılan id
	Source    string    `json:"source"`               // web | app
	Lang      string    `json:"lang"`
	IP        string    `json:"ip"`
	Messages  string    `json:"messages"` // JSON massiv: [{role, text}]
	Count     int       `json:"count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// logAiConversation — söhbəti saxla/yenilə (tam tarixçə + yeni AI cavabı).
func logAiConversation(id, source, lang, ip string, msgs []aiMsg, reply string) {
	if id == "" {
		return
	}
	if source != "app" {
		source = "web"
	}
	full := make([]aiMsg, 0, len(msgs)+1)
	full = append(full, msgs...)
	full = append(full, aiMsg{Role: "assistant", Text: reply})
	b, _ := json.Marshal(full)
	now := time.Now()

	var c AiConversation
	if db.First(&c, "id = ?", id).Error == nil {
		db.Model(&c).Updates(map[string]any{
			"source": source, "lang": lang, "ip": ip,
			"messages": string(b), "count": len(full), "updated_at": now,
		})
		return
	}
	db.Create(&AiConversation{
		ID: id, Source: source, Lang: lang, IP: ip,
		Messages: string(b), Count: len(full), CreatedAt: now, UpdatedAt: now,
	})
}

// GET /api/ai-chats — söhbətlər siyahısı (ən yeni yuxarıda, ilk sual önizləməsi ilə). Admin.
func listAiChats(w http.ResponseWriter, r *http.Request) {
	var cs []AiConversation
	db.Order("updated_at desc").Limit(1000).Find(&cs)
	out := make([]map[string]any, 0, len(cs))
	for _, c := range cs {
		var msgs []aiMsg
		_ = json.Unmarshal([]byte(c.Messages), &msgs)
		preview := ""
		for _, m := range msgs {
			if m.Role == "user" {
				preview = m.Text
				break
			}
		}
		out = append(out, map[string]any{
			"id": c.ID, "source": c.Source, "lang": c.Lang, "count": c.Count,
			"created_at": c.CreatedAt, "updated_at": c.UpdatedAt, "preview": preview,
		})
	}
	writeJSON(w, 200, out)
}

// GET /api/ai-chats/{id} — tam yazışma. Admin.
func getAiChat(w http.ResponseWriter, r *http.Request) {
	var c AiConversation
	if db.First(&c, "id = ?", r.PathValue("id")).Error != nil {
		writeJSON(w, 404, map[string]string{"error": "söhbət tapılmadı"})
		return
	}
	var msgs []aiMsg
	_ = json.Unmarshal([]byte(c.Messages), &msgs)
	writeJSON(w, 200, map[string]any{
		"id": c.ID, "source": c.Source, "lang": c.Lang, "ip": c.IP,
		"created_at": c.CreatedAt, "updated_at": c.UpdatedAt, "messages": msgs,
	})
}
