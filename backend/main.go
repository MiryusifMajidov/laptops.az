package main

import (
	"log"
	"net/http"
	"os"
)

// cors — Vite dev server (localhost:5173) üçün
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	bootstrap() // prod: datanı kalıcı volume-a bağla (dev-də no-op)
	initDB()

	// bir dəfəlik Excel köçürməsi:  go run . -import <data.json>
	if len(os.Args) >= 3 && os.Args[1] == "-import" {
		runImport(os.Args[2])
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/dashboard", dashboardHandler)
	mux.HandleFunc("GET /api/items", listItems)
	mux.HandleFunc("GET /api/items/deleted", listDeletedItems)
	mux.HandleFunc("POST /api/items", createItem)
	mux.HandleFunc("POST /api/items/{id}/restore", restoreItem)
	mux.HandleFunc("GET /api/categories", listCategories)
	mux.HandleFunc("GET /api/attributes", listAttributes)
	mux.HandleFunc("GET /api/branches", listBranches)
	mux.HandleFunc("GET /api/sales", listSales)
	mux.HandleFunc("GET /api/sales/summary", salesSummary)
	mux.HandleFunc("GET /api/sales/pending", pendingSaleItems)
	mux.HandleFunc("POST /api/sales", createSale)
	mux.HandleFunc("POST /api/sales/confirm", confirmSale)
	mux.HandleFunc("GET /api/customers", listCustomers)
	mux.HandleFunc("GET /api/credits", listCredits)
	mux.HandleFunc("GET /api/consignments", listConsignments)
	mux.HandleFunc("GET /api/expenses", listExpenses)
	mux.HandleFunc("GET /api/orders", listOrders)
	mux.HandleFunc("POST /api/orders/{id}/convert", convertOrder)
	mux.HandleFunc("POST /api/orders/{id}/cancel", cancelOrder)
	mux.HandleFunc("GET /api/kassa", kassaHandler)
	mux.HandleFunc("POST /api/kassa/close", createClose)
	mux.HandleFunc("GET /api/supplies", listSupplies)
	mux.HandleFunc("GET /api/reports", reportsHandler)
	// auth + create əməliyyatları
	mux.HandleFunc("POST /api/login", login)
	mux.HandleFunc("POST /api/customers", createCustomer)
	mux.HandleFunc("POST /api/expenses", createExpense)
	mux.HandleFunc("POST /api/categories", createCategory)
	mux.HandleFunc("POST /api/attributes", createAttribute)
	mux.HandleFunc("POST /api/attributes/{id}/options", addOption)
	mux.HandleFunc("PUT /api/attributes/{id}/options/order", reorderOptions)
	mux.HandleFunc("POST /api/branches", createBranch)
	mux.HandleFunc("POST /api/transfers", createTransfer)
	mux.HandleFunc("POST /api/supplies", createSupply)
	mux.HandleFunc("POST /api/consignments", createConsignment)
	// update / delete
	mux.HandleFunc("PUT /api/items/{id}", updateItem)
	mux.HandleFunc("POST /api/items/publish-all", publishAllItems)
	mux.HandleFunc("POST /api/import", importHandler)         // Excel→DB toplu köçürmə (yalnız admin, atomik)
	mux.HandleFunc("POST /api/reconcile", reconcileHandler)   // Excel tutuşdurması: toplu sold/used (yalnız admin, atomik)
	mux.HandleFunc("DELETE /api/items/{id}", deleteItem)
	mux.HandleFunc("PUT /api/expenses/{id}", updateExpense)    // xərc redaktə
	mux.HandleFunc("DELETE /api/expenses/{id}", deleteExpense) // xərc sil ({id} və ya "all", yalnız admin)
	mux.HandleFunc("PUT /api/customers/{id}", updateCustomer)
	mux.HandleFunc("PUT /api/sales/{id}", updateSale)
	mux.HandleFunc("PUT /api/categories/{id}", updateCategory)
	mux.HandleFunc("DELETE /api/categories/{id}", deleteCategory)
	mux.HandleFunc("PUT /api/attributes/{id}", updateAttribute)
	mux.HandleFunc("DELETE /api/attributes/{id}", deleteAttribute)
	mux.HandleFunc("PUT /api/attribute-options/{id}", updateOption)
	mux.HandleFunc("DELETE /api/attribute-options/{id}", deleteOption)
	// auth / rol / audit
	mux.HandleFunc("POST /api/logout", logout)
	mux.HandleFunc("POST /api/change-password", changePassword)
	mux.HandleFunc("GET /api/audit", listAudit)
	mux.HandleFunc("GET /api/ai-chats", listAiChats)      // AI köməkçi söhbətləri (admin)
	mux.HandleFunc("GET /api/ai-chats/{id}", getAiChat)
	mux.HandleFunc("GET /api/visits", listVisits)         // sayt ziyarətçiləri (admin)
	// şəkillər
	mux.HandleFunc("POST /api/upload", uploadImage)
	// kredit axını
	mux.HandleFunc("POST /api/credits", createCredit)
	mux.HandleFunc("POST /api/credits/{id}/pay", payCredit)
	mux.HandleFunc("PUT /api/credits/{id}", updateCredit) // "biz bağladıq" checkbox
	// kalkulyator (aralıq + filial + kateqoriya → dövriyyə/mənfəət/net)
	mux.HandleFunc("GET /api/calc", calcHandler)
	// realizasiya / təchizat status
	mux.HandleFunc("PUT /api/consignments/{id}", updateConsignment)
	mux.HandleFunc("PUT /api/supplies/{id}", updateSupply)
	// excel export + mail şablonları + göndərmə + backup
	mux.HandleFunc("GET /api/export/stock", exportStock)
	mux.HandleFunc("GET /api/email-templates", listEmailTemplates)
	mux.HandleFunc("POST /api/email-templates", createEmailTemplate)
	mux.HandleFunc("PUT /api/email-templates/{id}", updateEmailTemplate)
	mux.HandleFunc("DELETE /api/email-templates/{id}", deleteEmailTemplate)
	mux.HandleFunc("POST /api/email/send", sendStockEmail)
	// admin poçt (info@laptops.az) — kompoz + inbox
	mux.HandleFunc("POST /api/mail/send", sendComposedMail)
	mux.HandleFunc("GET /api/mail/inbox", listInbox)
	mux.HandleFunc("GET /api/mail/inbox/{id}", getInboxMail)
	mux.HandleFunc("POST /api/backup", backupDB)
	// çoxdilli — statik mətnlər + dillər (admin)
	mux.HandleFunc("GET /api/languages", listLanguages)
	mux.HandleFunc("POST /api/languages", createLanguage)
	mux.HandleFunc("PUT /api/languages/{id}", updateLanguage)
	mux.HandleFunc("DELETE /api/languages/{id}", deleteLanguage)
	mux.HandleFunc("GET /api/ui-strings", listUiStrings)
	mux.HandleFunc("PUT /api/ui-strings", saveUiStrings)
	mux.HandleFunc("POST /api/ui-strings/translate", translateUiStrings)
	// kataloq terminləri (kateqoriya / xüsusiyyət / dəyər tərcümələri)
	mux.HandleFunc("GET /api/terms", listTerms)
	mux.HandleFunc("PUT /api/terms", saveTerms)
	mux.HandleFunc("POST /api/terms/translate", translateTerms)
	mux.HandleFunc("POST /api/translate", translateText) // bir mətni (məs. məhsul adı) AI ilə tərcümə
	// sayt tənzimləmələri (xəritə koordinatı, seçilmiş məhsul)
	mux.HandleFunc("GET /api/settings", listSettings)
	mux.HandleFunc("PUT /api/settings", saveSettings)
	// tərəfdaşlıq (partner) + istifadəçi idarəçiliyi (admin)
	mux.HandleFunc("GET /api/partner-applications", listPartnerApplications)
	mux.HandleFunc("PUT /api/partner-applications/{id}", updatePartnerApplication)
	mux.HandleFunc("GET /api/users", listUsers)
	mux.HandleFunc("POST /api/users", createUser)
	mux.HandleFunc("DELETE /api/users/{id}", deleteUser)
	// partner (topdan mağaza) — optavoy qiymətli məhsullar (token tələb olunur)
	mux.HandleFunc("GET /api/partner/products", partnerProducts)
	mux.HandleFunc("GET /api/partner/products/{id}", partnerProduct)
	// PUBLIC (sayt üçün — auth yoxdur)
	mux.HandleFunc("GET /api/public/products", publicProducts)
	mux.HandleFunc("GET /api/public/products/{id}", publicProduct)
	mux.HandleFunc("GET /api/public/categories", publicCategories)
	mux.HandleFunc("GET /api/public/laptops", laptopsList)        // xarici laptops saytı: Mərkəz + stokda + aktiv Notebook
	mux.HandleFunc("GET /api/public/laptops/{id}", laptopDetail) // eyni filtrlə detal
	mux.HandleFunc("GET /api/public/languages", publicLanguages)
	mux.HandleFunc("GET /api/public/i18n", publicI18n)
	mux.HandleFunc("GET /api/public/terms", publicTerms)
	mux.HandleFunc("GET /api/public/settings", publicSettings)
	mux.HandleFunc("GET /api/public/notifications", publicNotifications)
	mux.HandleFunc("POST /api/public/partner-apply", publicPartnerApply)
	mux.HandleFunc("POST /api/public/orders", publicOrder)
	mux.HandleFunc("POST /api/public/ai-chat", aiChat) // müştəri AI köməkçisi (Gemini)
	mux.HandleFunc("POST /api/public/visit", trackVisit) // sayt ziyarəti loglama
	mux.HandleFunc("GET /api/public/whatsapp", waVerify)  // WhatsApp webhook doğrulama
	mux.HandleFunc("POST /api/public/whatsapp", waWebhook) // WhatsApp gələn mesaj
	mux.HandleFunc("GET /api/public/instagram", igVerify)  // Instagram webhook doğrulama
	mux.HandleFunc("POST /api/public/instagram", igWebhook) // Instagram gələn DM
	mux.HandleFunc("POST /api/public/mail-in", mailIn)      // Cloudflare Email Worker → gələn mail
	// yüklənmiş şəkillər (statik)
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	mux.HandleFunc("GET /img", imgResizeHandler)
	// SEO
	mux.HandleFunc("GET /sitemap.xml", sitemapHandler)
	mux.HandleFunc("GET /robots.txt", robotsHandler) // şəkil kiçiltmə/keş proksisi (kart/siyahı üçün)
	mux.HandleFunc("GET /privacy", privacyHandler)    // Google Play məxfilik siyasəti

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok"})
	})

	// Prod build-də (`-tags prod`) admin + saytı eyni portdan verir; dev-də no-op.
	mountStatic(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Laptops.az → http://localhost%s  (API: /api, panel: /admin)", addr)
	if err := http.ListenAndServe(addr, cors(authMiddleware(mux))); err != nil {
		log.Fatal(err)
	}
}
