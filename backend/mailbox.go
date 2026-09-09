package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"
)

// IncomingMail — Cloudflare Email Worker vasitəsilə gələn maillər (admin panelin inbox-u).
// Worker xam email-i /api/public/mail-in-ə POST edir; backend Go stdlib ilə parse edib saxlayır.
type IncomingMail struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Box        string    `gorm:"default:inbox" json:"box"` // inbox | sent
	FromAddr   string    `json:"from"`
	ToAddr     string    `json:"to"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	Attach     string    `json:"attach"` // JSON: əlavə fayl adları
	ReceivedAt time.Time `json:"received_at"`
	Seen       bool      `json:"seen"`
}

func mailInSecret() string {
	if v := os.Getenv("MAIL_IN_SECRET"); v != "" {
		return v
	}
	return "laptopsaz-mailin-2026"
}

// POST /api/public/mail-in — Cloudflare Worker xam RFC822 email POST edir (X-Mail-Secret ilə)
func mailIn(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Mail-Secret") != mailInSecret() {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	raw, _ := io.ReadAll(io.LimitReader(r.Body, 8<<20)) // 8MB
	w.WriteHeader(http.StatusOK)
	go func() {
		defer func() { _ = recover() }()
		from, to, subj, body := parseIncoming(raw)
		db.Create(&IncomingMail{FromAddr: from, ToAddr: to, Subject: subj, Body: body, ReceivedAt: time.Now()})
	}()
}

func parseIncoming(raw []byte) (from, to, subject, text string) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return "", "", "(başlıq oxunmadı)", string(raw)
	}
	dec := new(mime.WordDecoder)
	subject, _ = dec.DecodeHeader(msg.Header.Get("Subject"))
	from = decodeAddr(dec, msg.Header.Get("From"))
	to = decodeAddr(dec, msg.Header.Get("To"))
	ctype := msg.Header.Get("Content-Type")
	if ctype == "" {
		ctype = "text/plain"
	}
	text = extractText(ctype, msg.Header.Get("Content-Transfer-Encoding"), msg.Body)
	return
}

func decodeAddr(dec *mime.WordDecoder, h string) string {
	h = strings.TrimSpace(h)
	if a, e := mail.ParseAddress(h); e == nil {
		name, _ := dec.DecodeHeader(a.Name)
		if name != "" {
			return name + " <" + a.Address + ">"
		}
		return a.Address
	}
	d, _ := dec.DecodeHeader(h)
	return d
}

func decodeBodyBytes(cte string, r io.Reader) string {
	switch strings.ToLower(strings.TrimSpace(cte)) {
	case "base64":
		d, _ := io.ReadAll(base64.NewDecoder(base64.StdEncoding, r))
		return string(d)
	case "quoted-printable":
		d, _ := io.ReadAll(quotedprintable.NewReader(r))
		return string(d)
	default:
		d, _ := io.ReadAll(r)
		return string(d)
	}
}

var reHTMLTag = regexp.MustCompile(`(?s)<[^>]*>`)
var reScriptStyle = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
var reSpaces = regexp.MustCompile(`\n{3,}`)

func stripHTML(s string) string {
	s = reScriptStyle.ReplaceAllString(s, "")
	s = regexp.MustCompile(`(?i)<br\s*/?>`).ReplaceAllString(s, "\n")
	s = regexp.MustCompile(`(?i)</p>`).ReplaceAllString(s, "\n\n")
	s = reHTMLTag.ReplaceAllString(s, "")
	rep := strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", "\"", "&#39;", "'")
	s = rep.Replace(s)
	s = reSpaces.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

func extractText(ctype, cte string, body io.Reader) string {
	mediaType, params, err := mime.ParseMediaType(ctype)
	if err != nil {
		return strings.TrimSpace(decodeBodyBytes(cte, body))
	}
	return extractPart(mediaType, params, cte, body)
}

func extractPart(mediaType string, params map[string]string, cte string, body io.Reader) string {
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(body, params["boundary"])
		var htmlPart string
		for {
			p, e := mr.NextPart()
			if e != nil {
				break
			}
			pmt, pp, _ := mime.ParseMediaType(p.Header.Get("Content-Type"))
			pcte := p.Header.Get("Content-Transfer-Encoding")
			if strings.HasPrefix(pmt, "multipart/") {
				if inner := extractPart(pmt, pp, pcte, p); inner != "" {
					return inner
				}
				continue
			}
			if strings.HasPrefix(pmt, "text/plain") {
				return strings.TrimSpace(decodeBodyBytes(pcte, p))
			}
			if strings.HasPrefix(pmt, "text/html") && htmlPart == "" {
				htmlPart = decodeBodyBytes(pcte, p)
			}
		}
		if htmlPart != "" {
			return stripHTML(htmlPart)
		}
		return ""
	}
	data := decodeBodyBytes(cte, body)
	if strings.HasPrefix(mediaType, "text/html") {
		return stripHTML(data)
	}
	return strings.TrimSpace(data)
}

// GET /api/mail/inbox — gələn maillər siyahısı (admin)
func listInbox(w http.ResponseWriter, r *http.Request) {
	if !adminOnly(r) {
		writeJSON(w, 403, map[string]string{"error": "yalnız admin"})
		return
	}
	box := r.URL.Query().Get("box")
	if box == "" {
		box = "inbox"
	}
	var mails []IncomingMail
	q := db.Order("received_at desc").Limit(300)
	if box == "inbox" {
		q = q.Where("box = ? OR box = '' OR box IS NULL", "inbox")
	} else {
		q = q.Where("box = ?", box)
	}
	q.Find(&mails)
	for i := range mails { // siyahı üçün qısa snippet (rune-təhlükəsiz)
		rs := []rune(mails[i].Body)
		if len(rs) > 160 {
			mails[i].Body = string(rs[:160])
		}
	}
	writeJSON(w, 200, mails)
}

// GET /api/mail/inbox/{id} — tam mail (admin) + oxundu işarələ
func getInboxMail(w http.ResponseWriter, r *http.Request) {
	if !adminOnly(r) {
		writeJSON(w, 403, map[string]string{"error": "yalnız admin"})
		return
	}
	var m IncomingMail
	if db.First(&m, r.PathValue("id")).Error != nil {
		writeJSON(w, 404, map[string]string{"error": "tapılmadı"})
		return
	}
	if !m.Seen {
		db.Model(&m).Update("seen", true)
	}
	writeJSON(w, 200, m)
}
