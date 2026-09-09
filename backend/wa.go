package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// WhatsApp Business Cloud API inteqrasiyası — gələn mesajı eyni AI beyninə (aiReply)
// verir, cavabı geri göndərir. Söhbətlər admin paneldə "AI söhbətləri"ndə görünür (source=whatsapp).

const waAPIVersion = "v21.0"

// DEV-mode körpüsü: unpublished app-da Meta müştəri nömrəsini gizlədir və scoped id
// (məs. "AZ.209...") göndərir. Test üçün bilinən scoped id-ləri real nömrəyə bağlayırıq.
// App PUBLISH olandan sonra webhook real nömrəni verəcək və bu map LAZIM DEYİL (boş qala bilər).
var waDevMap = map[string]string{
	"AZ.2098399877455670": "994708151283",
}

func waVerifyToken() string {
	if v := os.Getenv("WA_VERIFY_TOKEN"); v != "" {
		return v
	}
	return "laptopsaz-wa-2026-verify" // Meta-da "Verify token" xanasına EYNİ bunu yaz
}

// GET /api/public/whatsapp — Meta webhook doğrulaması
func waVerify(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("hub.mode") == "subscribe" && q.Get("hub.verify_token") == waVerifyToken() {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(q.Get("hub.challenge")))
		return
	}
	w.WriteHeader(http.StatusForbidden)
}

// POST /api/public/whatsapp — gələn mesaj. Meta tez 200 gözləyir → arxa planda emal.
func waWebhook(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	log.Printf("[wa] webhook POST: %d bayt | RAW: %s", len(body), string(body))
	w.WriteHeader(http.StatusOK)
	go waProcess(body)
}

func waProcess(body []byte) {
	defer func() { _ = recover() }()
	var p struct {
		Entry []struct {
			Changes []struct {
				Value struct {
					Contacts []struct {
						WaID   string `json:"wa_id"`
						UserID string `json:"user_id"`
					} `json:"contacts"`
					Messages []struct {
						From       string `json:"from"`
						FromUserID string `json:"from_user_id"` // Meta yeni privacy: nömrə əvəzinə scoped id
						Type       string `json:"type"`
						Text       struct {
							Body string `json:"body"`
						} `json:"text"`
					} `json:"messages"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}
	if json.Unmarshal(body, &p) != nil {
		return
	}
	for _, e := range p.Entry {
		for _, ch := range e.Changes {
			fallbackFrom := ""
			if len(ch.Value.Contacts) > 0 {
				c := ch.Value.Contacts[0]
				fallbackFrom = c.WaID
				if fallbackFrom == "" {
					fallbackFrom = c.UserID
				}
			}
			for _, m := range ch.Value.Messages {
				from := m.From
				if from == "" {
					from = m.FromUserID
				}
				if from == "" {
					from = fallbackFrom
				}
				if m.Type == "text" && strings.TrimSpace(m.Text.Body) != "" && from != "" {
					handleWaMessage(from, m.Text.Body)
				}
			}
		}
	}
}

func handleWaMessage(from, text string) {
	convID := "wa-" + from
	var msgs []aiMsg
	var c AiConversation
	if db.First(&c, "id = ?", convID).Error == nil {
		_ = json.Unmarshal([]byte(c.Messages), &msgs)
	}
	log.Printf("[wa] gələn mesaj from=%s: %q", from, text)
	msgs = append(msgs, aiMsg{Role: "user", Text: text})
	reply, err := aiReply(msgs)
	if err != nil {
		reply = "Bağışlayın, indi cavab verə bilmirəm. Bir azdan yenidən yazın 🙏"
	}
	// DEV-mode körpüsü: unpublished app-da Meta nömrəni gizlədir (scoped id).
	// Bilinən test scoped id → real nömrə. Publish olandan sonra LAZIM DEYİL.
	sendTo := from
	if mapped, ok := waDevMap[from]; ok {
		sendTo = mapped
	}
	waSend(sendTo, waFormat(reply))
	logAiConversation(convID, "whatsapp", "", from, msgs, reply)
}

// [[product:ID]] markerlərini oxunaqlı mətnə (ad + qiymət + link) çevir, markdown təmizlə
func waFormat(reply string) string {
	out := productMarker.ReplaceAllStringFunc(reply, func(mk string) string {
		sub := productMarker.FindStringSubmatch(mk)
		if len(sub) < 2 {
			return ""
		}
		var it Item
		if db.Where("show_on_site = ? AND status = ?", true, "in_stock").First(&it, sub[1]).Error != nil {
			return ""
		}
		price := it.Price
		if it.Discount > 0 && it.Discount < it.Price {
			price = it.Price - it.Discount
		}
		return fmt.Sprintf("\n\n🛒 %s — ₼%.0f\nhttps://laptops.az/mehsul/%s", it.Name, price, sub[1])
	})
	out = strings.ReplaceAll(out, "**", "")
	out = strings.TrimSpace(out)
	if len(out) > 3900 {
		out = out[:3900]
	}
	return out
}

// waSend — WhatsApp Cloud API ilə mətn cavabı göndər
func waSend(to, text string) {
	token := os.Getenv("WA_TOKEN")
	phoneID := os.Getenv("WA_PHONE_ID")
	if token == "" || phoneID == "" || text == "" {
		log.Printf("wa send atlanıldı (WA_TOKEN/WA_PHONE_ID yox və ya boş mətn)")
		return
	}
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text":              map[string]any{"body": text, "preview_url": true},
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://graph.facebook.com/"+waAPIVersion+"/"+phoneID+"/messages", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		log.Printf("wa send err: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		rb, _ := io.ReadAll(resp.Body)
		log.Printf("wa send status %d: %s", resp.StatusCode, string(rb))
	}
}
