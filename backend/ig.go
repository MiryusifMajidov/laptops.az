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

// Instagram DM inteqrasiyası — gələn birbaşa mesajı (DM) eyni AI beyninə (aiReply)
// verir, cavabı geri göndərir. Söhbətlər admin paneldə "AI söhbətləri"ndə görünür (source=instagram).

// igVerifyToken — webhook doğrulama tokeni. Öz IG_VERIFY_TOKEN, yoxdursa WhatsApp-ınkı, yoxdursa default.
func igVerifyToken() string {
	if v := os.Getenv("IG_VERIFY_TOKEN"); v != "" {
		return v
	}
	if v := os.Getenv("WA_VERIFY_TOKEN"); v != "" {
		return v
	}
	return "laptopsaz-wa-2026-verify"
}

// igAPIBase — Send API host. Default: graph.instagram.com (Instagram Login yolu).
// Facebook Page yolu üçün IG_API_BASE=https://graph.facebook.com/v21.0 qoymaq olar.
func igAPIBase() string {
	if v := os.Getenv("IG_API_BASE"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "https://graph.instagram.com/v21.0"
}

// GET /api/public/instagram — Meta webhook doğrulaması
func igVerify(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("hub.mode") == "subscribe" && q.Get("hub.verify_token") == igVerifyToken() {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(q.Get("hub.challenge")))
		return
	}
	w.WriteHeader(http.StatusForbidden)
}

// POST /api/public/instagram — gələn DM. Meta tez 200 gözləyir → arxa planda emal.
func igWebhook(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	log.Printf("[ig] webhook POST: %d bayt | RAW: %s", len(body), string(body))
	w.WriteHeader(http.StatusOK)
	go igProcess(body)
}

func igProcess(body []byte) {
	defer func() { _ = recover() }()
	var p struct {
		Object string `json:"object"`
		Entry  []struct {
			ID        string `json:"id"`
			Messaging []struct {
				Sender    struct {
					ID string `json:"id"`
				} `json:"sender"`
				Recipient struct {
					ID string `json:"id"`
				} `json:"recipient"`
				Message struct {
					Mid    string `json:"mid"`
					Text   string `json:"text"`
					IsEcho bool   `json:"is_echo"` // öz göndərdiyimiz mesajın əks-sədası
				} `json:"message"`
			} `json:"messaging"`
		} `json:"entry"`
	}
	if json.Unmarshal(body, &p) != nil {
		return
	}
	for _, e := range p.Entry {
		for _, m := range e.Messaging {
			if m.Message.IsEcho { // öz cavabımızın əks-sədası — ötür (loop olmasın)
				continue
			}
			from := m.Sender.ID
			text := strings.TrimSpace(m.Message.Text)
			if from != "" && text != "" {
				handleIgMessage(from, text)
			}
		}
	}
}

func handleIgMessage(from, text string) {
	convID := "ig-" + from
	var msgs []aiMsg
	var c AiConversation
	if db.First(&c, "id = ?", convID).Error == nil {
		_ = json.Unmarshal([]byte(c.Messages), &msgs)
	}
	log.Printf("[ig] gələn mesaj from=%s: %q", from, text)
	msgs = append(msgs, aiMsg{Role: "user", Text: text})
	reply, err := aiReply(msgs)
	if err != nil {
		reply = "Bağışlayın, indi cavab verə bilmirəm. Bir azdan yenidən yazın 🙏"
	}
	igSend(from, igFormat(reply))
	logAiConversation(convID, "instagram", "", from, msgs, reply)
}

// [[product:ID]] markerlərini oxunaqlı mətnə (ad + qiymət + link) çevir, markdown təmizlə
func igFormat(reply string) string {
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
	if len(out) > 950 { // Instagram DM limiti ~1000 simvol
		out = out[:950]
	}
	return out
}

// igSend — Instagram Send API ilə DM cavabı göndər
func igSend(to, text string) {
	token := os.Getenv("IG_TOKEN")
	if token == "" || text == "" {
		log.Printf("ig send atlanıldı (IG_TOKEN yox və ya boş mətn)")
		return
	}
	payload := map[string]any{
		"recipient": map[string]any{"id": to},
		"message":   map[string]any{"text": text},
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", igAPIBase()+"/me/messages", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		log.Printf("ig send err: %v", err)
		return
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		log.Printf("ig send status %d: %s", resp.StatusCode, string(rb))
	}
}
