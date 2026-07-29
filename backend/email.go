package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// cari stokdan Excel qur (mağaza bağlananda göndərilən "cari cədvəl")
func buildStockExcel() (*bytes.Buffer, error) {
	f := excelize.NewFile()
	sh := "Cari Stok"
	f.SetSheetName("Sheet1", sh)
	headers := []string{"Məhsul", "Seriya", "Kateqoriya", "Filial", "Alış (₼)", "Status"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sh, cell, h)
	}
	var items []Item
	db.Preload("Category").Preload("Branch").Where("status = ?", "in_stock").Order("name").Find(&items)
	for i, it := range items {
		row := i + 2
		f.SetCellValue(sh, fmt.Sprintf("A%d", row), it.Name)
		f.SetCellValue(sh, fmt.Sprintf("B%d", row), it.Serial)
		f.SetCellValue(sh, fmt.Sprintf("C%d", row), it.Category.Name)
		f.SetCellValue(sh, fmt.Sprintf("D%d", row), it.Branch.Name)
		f.SetCellValue(sh, fmt.Sprintf("E%d", row), it.Cost)
		f.SetCellValue(sh, fmt.Sprintf("F%d", row), "Stokda")
	}
	f.SetColWidth(sh, "A", "A", 42)
	f.SetColWidth(sh, "B", "D", 20)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

// GET /api/export/stock — xlsx yüklə (frontend token ilə fetch edir)
func exportStock(w http.ResponseWriter, r *http.Request) {
	buf, err := buildStockExcel()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=laptops-cari-stok.xlsx")
	w.Write(buf.Bytes())
}

// ---- Mail şablonları ----

func listEmailTemplates(w http.ResponseWriter, r *http.Request) {
	var t []EmailTemplate
	db.Order("name").Find(&t)
	writeJSON(w, 200, t)
}

func createEmailTemplate(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name, Emails string }
	if err := decodeBody(r, &in); err != nil || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "ad vacibdir"})
		return
	}
	t := EmailTemplate{Name: in.Name, Emails: in.Emails}
	db.Create(&t)
	writeJSON(w, 201, t)
}

func updateEmailTemplate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct{ Name, Emails string }
	decodeBody(r, &in)
	db.Model(&EmailTemplate{}).Where("id = ?", id).Updates(map[string]any{"name": in.Name, "emails": in.Emails})
	var t EmailTemplate
	db.First(&t, id)
	writeJSON(w, 200, t)
}

func deleteEmailTemplate(w http.ResponseWriter, r *http.Request) {
	db.Delete(&EmailTemplate{}, r.PathValue("id"))
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// POST /api/email/send  {template_ids:[]} — seçilmiş şablonlara cari stok Excel-i göndər
func sendStockEmail(w http.ResponseWriter, r *http.Request) {
	var in struct {
		TemplateIDs []uint `json:"template_ids"`
	}
	if err := decodeBody(r, &in); err != nil || len(in.TemplateIDs) == 0 {
		writeJSON(w, 400, map[string]string{"error": "ən azı bir şablon seçin"})
		return
	}
	var tpls []EmailTemplate
	db.Where("id IN ?", in.TemplateIDs).Find(&tpls)
	seen := map[string]bool{}
	var recips []string
	for _, t := range tpls {
		for _, e := range strings.Split(t.Emails, ",") {
			e = strings.TrimSpace(e)
			if e != "" && !seen[e] {
				seen[e] = true
				recips = append(recips, e)
			}
		}
	}
	if len(recips) == 0 {
		writeJSON(w, 400, map[string]string{"error": "şablonlarda ünvan yoxdur"})
		return
	}

	host, port := os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT")
	user, pass, from := os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"), os.Getenv("SMTP_FROM")
	if from == "" {
		from = user
	}
	// SMTP konfiqurasiya olunmayıbsa — dürüst cavab (mexanizm hazırdır)
	if host == "" || user == "" {
		writeJSON(w, 200, map[string]any{
			"sent":       false,
			"recipients": recips,
			"note":       "SMTP hələ qoşulmayıb. Serverdə SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS env dəyişənlərini qoyanda avtomatik göndəriləcək.",
		})
		return
	}

	buf, err := buildStockExcel()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	subject := "Laptops.az — cari stok " + time.Now().Format("02.01.2006")
	msg := buildMIME(from, recips, subject, "Cari stok cədvəli əlavədədir (Excel).", "laptops-cari-stok.xlsx", buf.Bytes())
	auth := smtp.PlainAuth("", user, pass, host)
	if err := smtp.SendMail(host+":"+port, auth, from, recips, msg); err != nil {
		writeJSON(w, 500, map[string]string{"error": "göndərilmədi: " + err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"sent": true, "recipients": recips})
}

func buildMIME(from string, to []string, subject, body, attachName string, attach []byte) []byte {
	b := &bytes.Buffer{}
	boundary := "LAPTOPSAZMIMEBOUNDARY"
	fmt.Fprintf(b, "From: %s\r\n", from)
	fmt.Fprintf(b, "To: %s\r\n", strings.Join(to, ","))
	fmt.Fprintf(b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(b, "Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary)
	fmt.Fprintf(b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString(body + "\r\n\r\n")
	fmt.Fprintf(b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n")
	fmt.Fprintf(b, "Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", attachName)
	enc := base64.StdEncoding.EncodeToString(attach)
	for i := 0; i < len(enc); i += 76 {
		end := i + 76
		if end > len(enc) {
			end = len(enc)
		}
		b.WriteString(enc[i:end] + "\r\n")
	}
	fmt.Fprintf(b, "\r\n--%s--\r\n", boundary)
	return b.Bytes()
}

// POST /api/backup — SQLite bazasının təmiz nüsxəsi
func backupDB(w http.ResponseWriter, r *http.Request) {
	os.MkdirAll("backups", 0755)
	name := fmt.Sprintf("backups/laptops-%s.db", time.Now().Format("2006-01-02-1504"))
	if err := db.Exec("VACUUM INTO '" + name + "'").Error; err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"file": name})
}
