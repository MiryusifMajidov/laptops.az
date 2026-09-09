package main

import (
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func initDB() {
	var err error
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}
	// prod: DATABASE_URL (Postgres) qoyulubsa onu işlət; yoxdursa lokal SQLite
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err = gorm.Open(postgres.Open(dsn), cfg)
		log.Println("DB: PostgreSQL")
	} else {
		db, err = gorm.Open(sqlite.Open("laptops.db"), cfg)
	}
	if err != nil {
		log.Fatalf("db aç: %v", err)
	}
	// quantity sütunu artıq varmı? (ilk dəfə əlavə olunanda köhnə sətirləri 1-ə köçürmək üçün)
	hadItemQty := db.Migrator().HasColumn(&Item{}, "quantity")
	hadSaleQty := db.Migrator().HasColumn(&Sale{}, "quantity")

	if err := db.AutoMigrate(
		&Branch{}, &Category{}, &Attribute{}, &AttributeOption{},
		&Item{}, &ItemAttributeValue{}, &Customer{},
		&Sale{}, &CreditPlan{}, &Consignment{}, &Expense{},
		&OnlineOrder{}, &DailyClose{}, &SupplyBatch{},
		&User{}, &AuditLog{}, &EmailTemplate{}, &CreditPayment{},
		&Language{}, &UiString{}, &ItemTranslation{}, &TermTranslation{}, &Setting{},
		&PartnerApplication{}, &Session{}, &AiConversation{}, &Visit{}, &IncomingMail{},
	); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// BİRDƏFƏLİK: quantity sütunu bu deploy-da yeni yarandısa, mövcud bütün
	// məhsul/satışları 1 ədəd et (hər köhnə qeyd = 1 fiziki ədəd). Sonrakı
	// deploy-larda işləmir → aksesuarın həqiqi 0 stokunu sıfırlamaz.
	if !hadItemQty {
		db.Exec("UPDATE items SET quantity = 1")
		log.Println("migrasiya: mövcud məhsullar quantity=1 edildi")
	}
	if !hadSaleQty {
		db.Exec("UPDATE sales SET quantity = 1")
		log.Println("migrasiya: mövcud satışlar quantity=1 edildi")
	}
	seed()
	seedAuth()
	seedI18n()
	seedSettings()
	// köhnə "pending satış" obyektlərini sil — təsdiq gözləyənlər artıq birbaşa items
	// cədvəlindən (status=sold, satışı olmayan) hesablanır, ayrıca satış obyekti yaradılmır.
	db.Where("pending = ?", true).Delete(&Sale{})

	// BİRDƏFƏLİK: köhnə legacy zibil — qiyməti=0 & alışı=0, satışı olmayan «satıldı»
	// cihazlar (səhvən sold kimi idxal olunmuş) → silinmişə köçürülür. Marker ilə yalnız
	// bir dəfə; qiyməti olan real cihazlara TOXUNMUR.
	if getSetting("junk_sold_cleanup_done") != "1" {
		res := db.Where("status = ? AND price = 0 AND cost = 0 AND id NOT IN (?)", "sold",
			db.Model(&Sale{}).Select("item_id")).Delete(&Item{})
		setSetting("junk_sold_cleanup_done", "1")
		log.Printf("migrasiya: %d köhnə qiymətsiz «satıldı» cihaz silinmişə köçürüldü", res.RowsAffected)
	}
	syncNotebookFromExcel() // Excel → Mərkəz komputer uyğunlaşdırması (bir dəfəlik)
	notebookCleanupV2()     // köhnə satıldı sil + DELL PRO bərpa + adlardan NOTEBOOK sil
	notebookRestoreNames()      // adlardakı «NOTEBOOK» sözünü geri qaytar (v2 səhvi düzəlişi)
	addMissingAttributeOptions() // istifadədəki dəyərləri option kimi əlavə et (prosessor və s.)
	applyIncrementalSync()       // yeni Excel versiyasından inkremental əlavələr (add/sales/realiz)
	applyBranchSync()            // Zaur/Elçin filial uyğunlaşdırması (moves/adds/completes)
	deleteSerialessZaurLaptops() // Excel-də seriyası olmayan seriyasız Zaur laptopları sil
	applyReconcile202609()       // 2026-09 uzlaşdırma auditi: realizasiya/filial/maya düzəlişləri
	applyCreditBackfill202609()  // Excel «KREDIT» vərəqi → credit_plans (ilkin ödəniş 0, ilk ödəniş +1 ay)
	applyFix2202609()            // kredit satışlarını sil + son 10 günün əksik satış/stokunu tamamla
	applyFix3202609()            // seriyasız (aksesuar/sborka/monitor) əksik satışlar
	applyFix4202609()            // fix3-də səhvən yaranmış dublikat satışların təmizlənməsi
	applyFix5202609()            // son dublikat satışın təmizlənməsi
	applyFix6202609()            // səhvən yaradılmış 10 boş məhsul qeydi (istifadəçi göstərişi)
	applyFix7202609()            // kredit köçürməsinin təmizlənməsi (47 sil + 3 bərpa)
}

// optOrder — attribute option-larını sıraya (position, sonra id) görə düzmək üçün Preload köməkçisi
func optOrder(d *gorm.DB) *gorm.DB { return d.Order("position").Order("id") }

// unscoped — preload-da soft-deleted (silinmiş) yazıları da daxil et.
// Məs. satılmış cihaz sonradan silinsə belə, satış sətrində adı görünsün.
func unscoped(d *gorm.DB) *gorm.DB { return d.Unscoped() }

func seedAuth() {
	var n int64
	db.Model(&User{}).Count(&n)
	if n == 0 {
		db.Create(&[]User{
			{Username: "admin", PassHash: hashPassword("admin"), Role: "admin", Name: "Səid"},
			{Username: "user", PassHash: hashPassword("user"), Role: "user", Name: "Satıcı"},
		})
		log.Println("istifadəçilər yaradıldı: admin/admin (admin), user/user (satıcı)")
	}
	// filial id-ləri (satıcıları filiallara bağlamaq üçün)
	branchID := func(where string, args ...any) uint {
		var b Branch
		if db.Where(where, args...).First(&b).Error == nil {
			return b.ID
		}
		return 0
	}
	mainB := branchID("is_main = ?", true)
	zaurB := branchID("name = ?", "Zaur")
	elcinB := branchID("name LIKE ?", "%Elçin%")

	// işçi hesabları (satıcı rolu) — additive: yoxdursa yaradılır, mövcud parolu POZMUR.
	// Hər işçi öz filialına bağlıdır (yalnız o filialın satışını görür).
	emps := []struct {
		u, name string
		branch  uint
	}{
		{"ammar", "Ammar", mainB}, {"nicat", "Nicat", mainB}, {"seyran", "Seyran", mainB}, {"yusif", "Yusif", mainB},
		{"zaur", "Zaur", zaurB},     // Zaur filialı — user: zaur / parol: zaur
		{"elcin", "Elçin", elcinB},  // Elçin & Rəşid filialı — user: elcin / parol: elcin
	}
	for _, e := range emps {
		var existing User
		if db.Where("username = ?", e.u).First(&existing).Error != nil {
			db.Create(&User{Username: e.u, PassHash: hashPassword(e.u), Role: "user", Name: e.name, BranchID: e.branch})
			log.Printf("işçi hesabı yaradıldı: %s / %s (satıcı, filial=%d)", e.u, e.u, e.branch)
		} else if existing.BranchID == 0 && e.branch != 0 {
			db.Model(&existing).Update("branch_id", e.branch)
			log.Printf("işçi filialı təyin edildi: %s → filial %d", e.u, e.branch)
		}
	}
	// filialı təyin edilməyən qalan istifadəçilər (admin, user…) → Mərkəz
	if mainB != 0 {
		db.Model(&User{}).Where("branch_id = 0 OR branch_id IS NULL").Update("branch_id", mainB)
	}
	var m int64
	db.Model(&EmailTemplate{}).Count(&m)
	if m == 0 {
		db.Create(&[]EmailTemplate{
			{Name: "Öz mağazamız", Emails: "merkez@laptops.az"},
			{Name: "Digər filiallar", Emails: "elcin@laptops.az, zaur@laptops.az"},
		})
	}
}

func daysAgo(d int) time.Time { return time.Now().AddDate(0, 0, -d) }

func seed() {
	// Real datalar köçürüldükdən sonra demo seed YALNIZ açıq istəklə işləyir:
	//   SEED_DEMO=1 go run .
	// Əks halda boş cədvəl görüb üstünə fake data yazmaz.
	if os.Getenv("SEED_DEMO") != "1" {
		return
	}
	var n int64
	db.Model(&Item{}).Count(&n)
	if n > 0 {
		return // artıq seed olunub
	}
	log.Println("seed başladı…")

	// --- filiallar ---
	branches := []Branch{
		{Name: "Mərkəz (Mağaza)", IsMain: true, Address: "28 May, Süleyman Rəhimov 201"},
		{Name: "Elçin & Rəşid", Address: "Filial anbarı"},
		{Name: "Zaur", Address: "Filial anbarı"},
	}
	db.Create(&branches)

	// --- xüsusiyyətlər + seçimlər ---
	attrDefs := map[string][]string{
		"Marka":       {"HP", "Lenovo", "ASUS", "Apple", "Dell", "Acer", "MSI", "Samsung"},
		"Prosessor":   {"Intel i3", "Intel i5", "Intel i7", "Ultra 5", "Ultra 7", "Ryzen 5", "Ryzen 7"},
		"RAM":         {"8 GB", "16 GB", "32 GB", "64 GB"},
		"SSD":         {"256 GB", "512 GB", "1 TB", "2 TB"},
		"Ekran kartı": {"Intel Iris", "RTX 4050", "RTX 4060", "RTX 4070", "RTX 5050"},
		"Ekran":       {"13.3\"", "14\"", "15.6\"", "16\"", "17.3\""},
		"Vəziyyət":    {"Yeni", "Qutusuz", "İşlənmiş", "Zədəli"},
		"Yaddaş":      {"128 GB", "256 GB", "512 GB", "1 TB"},
		"Rəng":        {"Qara", "Ağ", "Boz", "Göy"},
		"Növ":         {"Siçan", "Klaviatura", "SSD", "RAM", "Çanta", "Adapter", "Xarici disk"},
		"Ölçü":        {"24\"", "27\"", "32\""},
		"Tezlik":      {"75Hz", "144Hz", "165Hz", "180Hz"},
	}
	attrs := map[string]*Attribute{}
	for name, opts := range attrDefs {
		a := Attribute{Name: name}
		for _, o := range opts {
			a.Options = append(a.Options, AttributeOption{Value: o})
		}
		db.Create(&a)
		attrs[name] = &a
	}

	// --- kateqoriyalar (hansı xüsusiyyətlər aiddir) ---
	catDefs := map[string][]string{
		"Notebook":  {"Marka", "Prosessor", "RAM", "SSD", "Ekran kartı", "Ekran", "Vəziyyət"},
		"İşlənmiş":  {"Marka", "Prosessor", "RAM", "SSD", "Ekran", "Vəziyyət"},
		"Telefon":   {"Marka", "Yaddaş", "Rəng", "Vəziyyət"},
		"Aksesuar":  {"Növ", "Marka"},
		"Yığım (PC)": {"Prosessor", "RAM", "SSD", "Ekran kartı"},
		"Monitor":   {"Marka", "Ölçü", "Tezlik"},
	}
	cats := map[string]*Category{}
	for name, anames := range catDefs {
		c := Category{Name: name}
		db.Create(&c)
		// join sətirlərini birbaşa yazırıq ki, GORM mövcud xüsusiyyətləri təsadüfən üzərinə yazmasın
		for _, an := range anames {
			db.Exec("INSERT INTO category_attributes (category_id, attribute_id) VALUES (?, ?)", c.ID, attrs[an].ID)
		}
		cats[name] = &c
	}

	// helper: cihaz + xüsusiyyət dəyərləri əlavə et
	addItem := func(name, serial, cat, branchName string, cost float64, status string, site bool, vals map[string]string, ageDays int) *Item {
		it := Item{
			Name: name, Serial: serial, CategoryID: cats[cat].ID,
			BranchID: branchByName(branches, branchName), Cost: cost,
			Status: status, ShowOnSite: site, CreatedAt: daysAgo(ageDays),
		}
		for an, v := range vals {
			if a, ok := attrs[an]; ok {
				it.Values = append(it.Values, ItemAttributeValue{AttributeID: a.ID, Value: v})
			}
		}
		// Values.Attribute-i omit edirik: yalnız AttributeID (FK) yazılsın, boş Attribute yaranmasın
		db.Omit("Values.Attribute").Create(&it)
		return &it
	}

	hp := addItem("NOTEBOOK HP ProBook 450 G10", "5cd5232k5c", "Notebook", "Mərkəz (Mağaza)", 960, "in_stock", true,
		map[string]string{"Marka": "HP", "Prosessor": "Ultra 7", "RAM": "16 GB", "SSD": "512 GB"}, 6)
	mac := addItem("MacBook Air M3 15\"", "DKQ5YKH43K", "Notebook", "Mərkəz (Mağaza)", 2000, "in_stock", true,
		map[string]string{"Marka": "Apple", "RAM": "8 GB", "SSD": "256 GB"}, 3)
	addItem("NOTEBOOK ASUS Vivobook X1404V", "t1n0cx07k886047", "Notebook", "Zaur", 548, "in_stock", true,
		map[string]string{"Marka": "ASUS", "Prosessor": "Intel i5", "RAM": "8 GB", "SSD": "512 GB"}, 20)
	leg := addItem("NOTEBOOK Lenovo Legion 5 Pro", "PF3A22KD", "Notebook", "Mərkəz (Mağaza)", 3800, "reserved", true,
		map[string]string{"Marka": "Lenovo", "Prosessor": "Intel i7", "RAM": "32 GB", "Ekran kartı": "RTX 4070"}, 2)
	addItem("NOTEBOOK Dell XPS 16 Premium", "DT1TSB4", "Notebook", "Mərkəz (Mağaza)", 2890, "in_stock", false,
		map[string]string{"Marka": "Dell", "Prosessor": "Ultra 7", "RAM": "32 GB", "Vəziyyət": "Qutusuz"}, 118)
	addItem("SBORKA HP EliteDesk 800 G5", "2M04442JWZ", "Yığım (PC)", "Mərkəz (Mağaza)", 221, "in_stock", true,
		map[string]string{"Prosessor": "Intel i5", "RAM": "16 GB", "SSD": "512 GB"}, 9)
	addItem("MONITOR MSI 27\" 180Hz", "MSI-27-180", "Monitor", "Mərkəz (Mağaza)", 197, "in_stock", true,
		map[string]string{"Marka": "MSI", "Ölçü": "27\"", "Tezlik": "180Hz"}, 12)
	addItem("NOTEBOOK Acer Aspire 5 (işlənmiş)", "lxrms01027152", "İşlənmiş", "Elçin & Rəşid", 340, "in_stock", true,
		map[string]string{"Marka": "Acer", "Prosessor": "Intel i3", "RAM": "8 GB", "Vəziyyət": "İşlənmiş"}, 96)
	addItem("SSD Kingston 512GB", "KING-512", "Aksesuar", "Mərkəz (Mağaza)", 50, "in_stock", true,
		map[string]string{"Növ": "SSD"}, 4)
	soldItem := addItem("iPad Air 11\" M2 256GB", "JL7XDYV4G1", "Telefon", "Mərkəz (Mağaza)", 1280, "sold", false,
		map[string]string{"Marka": "Apple", "Yaddaş": "256 GB"}, 30)

	// --- müştərilər ---
	customers := []Customer{
		{Name: "Rauf İsmayılov", Phone: "050 xxx 12 40"},
		{Name: "Vəlixan Məmmədov", Phone: "055 209 89 95"},
		{Name: "Fuad Ələkbərov", Phone: "055 xxx 68 68"},
		{Name: "Günel Məmmədli", Phone: "070 xxx 55 10"},
	}
	db.Create(&customers)

	// --- satışlar (nümunə) ---
	sales := []Sale{
		{ItemID: soldItem.ID, SalePrice: 1500, Profit: 220, Channel: "card", BranchID: 1, SoldAt: daysAgo(1)},
	}
	db.Create(&sales)

	// --- kreditlər ---
	credits := []CreditPlan{
		{CustomerID: customers[1].ID, ItemName: "Lenovo Legion 5", Total: 1600, Paid: 400, NextDue: daysAgo(5), Status: "overdue"},
		{CustomerID: customers[2].ID, ItemName: "iPhone 13", Total: 440, Paid: 385, NextDue: time.Now(), Status: "duetoday"},
		{CustomerID: customers[0].ID, ItemName: "HP ProBook", Total: 1320, Paid: 680, NextDue: daysAgo(-1), Status: "ontime"},
	}
	db.Create(&credits)
	// ödəniş tarixçəsi (seed)
	db.Create(&[]CreditPayment{
		{CreditPlanID: credits[0].ID, Amount: 400, CreatedAt: daysAgo(35)},
		{CreditPlanID: credits[1].ID, Amount: 200, CreatedAt: daysAgo(40)},
		{CreditPlanID: credits[1].ID, Amount: 185, CreatedAt: daysAgo(15)},
		{CreditPlanID: credits[2].ID, Amount: 380, CreatedAt: daysAgo(30)},
		{CreditPlanID: credits[2].ID, Amount: 300, CreatedAt: daysAgo(10)},
	})

	// --- realizasiya (başqa mağazalara) ---
	cons := []Consignment{
		{StoreName: "Salman (mağaza)", ItemName: "ASUS ROG Strix G615", Serial: "T5NRSG00A34", GivenAt: daysAgo(19), Cost: 2569, GivenPrice: 3500, Status: "out", Debt: 2569},
		{StoreName: "Niyaz (mağaza)", ItemName: "HP 14-DQ6015DX", Serial: "1h852635fl", GivenAt: daysAgo(15), Cost: 341, GivenPrice: 500, Status: "sold_unpaid", Debt: 341},
		{StoreName: "Gamemax Cavid", ItemName: "SBORKA EliteDesk 800", Serial: "", GivenAt: daysAgo(8), Cost: 221, GivenPrice: 335, Status: "sold_paid", Debt: 0},
	}
	db.Create(&cons)

	// --- xərclər ---
	exp := []Expense{
		{Date: daysAgo(12), Category: "Kirayə", Note: "Mağaza icarəsi", Amount: 3500},
		{Date: daysAgo(8), Category: "Maaş", Note: "2 satıcı", Amount: 4200},
		{Date: daysAgo(5), Category: "Kommunal", Note: "İşıq + internet", Amount: 480},
		{Date: daysAgo(2), Category: "Nəqliyyat", Note: "Dubay kargo", Amount: 900},
	}
	db.Create(&exp)

	// --- bu günün satışları (kassa üçün) ---
	mouse := addItem("MOUSE 2E sadə", "", "Aksesuar", "Mərkəz (Mağaza)", 12, "sold", true, map[string]string{"Növ": "Siçan"}, 0)
	ssd2 := addItem("SSD 512GB", "", "Aksesuar", "Mərkəz (Mağaza)", 50, "sold", true, map[string]string{"Növ": "SSD"}, 0)
	sbor := addItem("SBORKA CORE I7 13GEN RTX4060", "", "Yığım (PC)", "Mərkəz (Mağaza)", 1895, "sold", true, map[string]string{"Prosessor": "Intel i7", "RAM": "32 GB", "Ekran kartı": "RTX 4060"}, 0)
	db.Create(&[]Sale{
		{ItemID: mouse.ID, SalePrice: 25, Profit: 13, Channel: "cash", BranchID: 1, SoldAt: time.Now(), Counted: true},
		{ItemID: ssd2.ID, SalePrice: 55, Profit: 5, Channel: "card", BranchID: 1, SoldAt: time.Now(), Counted: true},
		{ItemID: sbor.ID, SalePrice: 2480, Profit: 585, Channel: "installment", BranchID: 1, SoldAt: time.Now(), Counted: true},
	})

	// --- onlayn sifarişlər (saytdan) ---
	now := time.Now()
	db.Create(&[]OnlineOrder{
		{CustomerName: "Elvin Q.", Phone: "+994 50 xxx 41 22", ItemID: mac.ID, Status: "pending", CreatedAt: now.Add(-18 * time.Minute), ExpiresAt: now.Add(23 * time.Hour)},
		{CustomerName: "Nicat A.", Phone: "+994 55 xxx 09 87", ItemID: hp.ID, Status: "pending", CreatedAt: now.Add(-41 * time.Minute), ExpiresAt: now.Add(22 * time.Hour)},
		{CustomerName: "Günel M.", Phone: "+994 70 xxx 55 10", ItemID: leg.ID, Status: "pending", CreatedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(21 * time.Hour)},
	})

	// --- kassa keçmiş bağlanışları ---
	db.Create(&[]DailyClose{
		{Date: daysAgo(1), Cash: 4200, Card: 5100, Installment: 940, Credit: 0, Total: 10240},
		{Date: daysAgo(2), Cash: 3100, Card: 4200, Installment: 580, Credit: 0, Total: 7880},
		{Date: daysAgo(3), Cash: 5400, Card: 6300, Installment: 1400, Credit: 0, Total: 13100},
	})

	// --- təchizat partiyaları ---
	db.Create(&[]SupplyBatch{
		{Date: daysAgo(4), Source: "Dubay", Supplier: "Şəxsi partiya", ItemCount: 14, TotalCost: 18400, Status: "received"},
		{Date: daysAgo(12), Source: "Yerli", Supplier: "NoteTech", ItemCount: 6, TotalCost: 7100, Status: "received"},
		{Date: daysAgo(19), Source: "İstanbul", Supplier: "Partiya #204", ItemCount: 9, TotalCost: 12300, Status: "pending"},
	})

	// --- keçmiş satış tarixçəsi (~28 gün, real görünüş üçün) ---
	type stmpl struct {
		name, cat, ch string
		cost, price   float64
		attrs         map[string]string
	}
	tmpls := []stmpl{
		{"NOTEBOOK HP ProBook 450 G10", "Notebook", "card", 900, 1150, map[string]string{"Marka": "HP", "RAM": "16 GB", "SSD": "512 GB"}},
		{"NOTEBOOK Lenovo IdeaPad 3", "Notebook", "cash", 600, 800, map[string]string{"Marka": "Lenovo", "RAM": "8 GB", "SSD": "512 GB"}},
		{"NOTEBOOK ASUS Vivobook 15", "Notebook", "installment", 550, 760, map[string]string{"Marka": "ASUS", "RAM": "16 GB"}},
		{"MacBook Air M2 13\"", "Notebook", "card", 1600, 1950, map[string]string{"Marka": "Apple", "SSD": "256 GB"}},
		{"NOTEBOOK Dell Inspiron 15", "Notebook", "cash", 700, 920, map[string]string{"Marka": "Dell", "RAM": "16 GB"}},
		{"NOTEBOOK Acer Nitro 5", "Notebook", "card", 950, 1250, map[string]string{"Marka": "Acer", "Ekran kartı": "RTX 4060"}},
		{"SBORKA HP EliteDesk i5", "Yığım (PC)", "cash", 400, 560, map[string]string{"Prosessor": "Intel i5"}},
		{"NOTEBOOK HP Omnibook 14", "Notebook", "installment", 850, 1120, map[string]string{"Marka": "HP", "Prosessor": "Ultra 7"}},
		{"NOTEBOOK ASUS TUF Gaming", "Notebook", "card", 1100, 1420, map[string]string{"Marka": "ASUS", "Ekran kartı": "RTX 4070"}},
		{"Aksesuar dəsti (SSD/RAM/siçan)", "Aksesuar", "cash", 60, 95, map[string]string{"Növ": "SSD"}},
	}
	var bulk []Sale
	for i := 0; i < 60; i++ {
		t := tmpls[i%len(tmpls)]
		it := addItem(t.name, "", t.cat, "Mərkəz (Mağaza)", t.cost, "sold", true, t.attrs, 0)
		bulk = append(bulk, Sale{ItemID: it.ID, SalePrice: t.price, Profit: t.price - t.cost, Channel: t.ch, BranchID: 1, SoldAt: daysAgo(i % 28)})
	}
	db.Create(&bulk)

	// sayt/satış qiyməti (alışdan ~%25 yuxarı) — seed üçün
	db.Exec("UPDATE items SET price = ROUND(cost * 1.25) WHERE price = 0 OR price IS NULL")

	log.Println("seed tamamlandı.")
}

func branchByName(branches []Branch, name string) uint {
	for _, b := range branches {
		if b.Name == name {
			return b.ID
		}
	}
	return 1
}
