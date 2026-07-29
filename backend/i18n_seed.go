package main

import "log"

// kod → İngiliscə dil adı (AI tərcümə promptu üçün)
var langEnglishName = map[string]string{
	"az": "Azerbaijani", "ru": "Russian", "tr": "Turkish", "en": "English",
	"ar": "Arabic", "fa": "Persian", "de": "German", "fr": "French", "es": "Spanish",
}

func langNameEN(code string) string {
	if n, ok := langEnglishName[code]; ok {
		return n
	}
	return code
}

// kv — açar + əsas (AZ) dəyər. Sıra admin "Sayt mətnləri" tabında qorunur.
type kv struct{ Key, Value string }

// defaultUiStrings — saytdakı bütün statik mətnlər (AZ əsas dil).
// Yeni açar əlavə olunanda seedI18n onu avtomatik yaradır (mövcud dəyərləri pozmadan).
func defaultUiStrings() []kv {
	return []kv{
		// header / naviqasiya
		{"promo.banner", "Onlayn sifariş et → ödənişi mağazada et: Nağd · Kart · Taksit kartı"},
		{"nav.notebook", "Notebooklar"},
		{"nav.used", "İşlənmiş"},
		{"nav.desktop", "Stolüstü PC"},
		{"nav.phone", "Telefon"},
		{"nav.accessory", "Aksesuar"},
		{"search.placeholder", "Axtar…"},
		{"search.placeholderLong", "Marka, model və ya xüsusiyyət axtar…"},

		// ümumi (bir çox yerdə)
		{"common.available", "Mövcuddur"},
		{"common.order", "Sifariş et"},
		{"common.details", "Ətraflı"},
		{"common.call", "Zəng et"},
		{"common.home", "Ana səhifə"},
		{"common.loading", "Yüklənir…"},

		// ana səhifə (hero + bölmələr)
		{"home.heroEyebrow", "2009-cu ildən Bakıda"},
		{"home.heroTitle", "Sizə uyğun\nnotebook. Zəmanətlə."},
		{"home.heroText", "Yeni və işlənmiş notebooklar, telefon və aksesuarlar. Onlayn seçin və sifariş edin — ödənişi mağazada edin."},
		{"home.ctaBrowse", "Notebooklara bax"},
		{"home.ctaVisit", "Mağazaya gəl"},
		{"home.liveText", "Stok real vaxtda yenilənir — «Mövcuddur» görünən məhsul mağazada var"},
		{"home.featuredEyebrow", "SEÇİLMİŞ"},
		{"home.chipAll", "Hamısı"},
		{"home.featuredTitle", "Seçilmiş məhsullar"},
		{"home.seeAll", "Hamısına bax"},
		{"home.loadingProducts", "Məhsullar yüklənir…"},
		{"home.tileNotebookTitle", "Notebooklar"},
		{"home.tileNotebookDesc", "HP · Lenovo · ASUS · Apple · Dell · MSI · Huawei"},
		{"home.tileUsedTitle", "İşlənmiş"},
		{"home.tileUsedDesc", "Yoxlanılmış, zəmanətli"},
		{"home.tileDesktopTitle", "Stolüstü PC"},
		{"home.tileDesktopDesc", "Oyun və iş üçün yığım kompüterlər"},
		{"home.tileCollection", "Kolleksiyaya bax"},
		{"home.tileView", "Bax"},
		{"home.tilePhoneTitle", "Telefonlar"},
		{"home.tilePhoneDesc", "iPhone · Samsung · Xiaomi"},
		{"home.tileAccessoryTitle", "Aksesuarlar"},
		{"home.tileAccessoryDesc", "Siçan · klaviatura · SSD · monitor · çanta"},

		// məhsul səhifəsi
		{"product.notFound", "Məhsul tapılmadı."},
		{"product.orderNote", "Onlayn ödəniş yoxdur. Sifariş edin — məhsulu mağazada yoxlayıb Nağd, Kart və ya Taksit kartı ilə ödəyin."},
		{"product.buybarNote", "Ödəniş mağazada — Nağd / Kart / Taksit"},

		// kateqoriya / filtr
		{"cat.sortPopular", "Populyar"},
		{"cat.sortCheap", "Ucuz → Baha"},
		{"cat.sortExp", "Baha → Ucuz"},
		{"cat.searchTitle", "Axtarış"},
		{"cat.allProducts", "Bütün məhsullar"},
		{"cat.productsSuffix", "məhsul · real-vaxt stok"},
		{"cat.filter", "Filtr"},
		{"cat.filters", "Filtrlər"},
		{"cat.clear", "Təmizlə"},
		{"cat.price", "Qiymət"},
		{"cat.sort", "Sıralama"},
		{"cat.reset", "Sıfırla"},
		{"cat.noFilter", "Filtr yoxdur"},
		{"cat.noProducts", "Uyğun məhsul yoxdur."},
		{"cat.showResults", "nəticəni göstər"},

		// footer
		{"footer.tagline", "2009-cu ildən etibarən Bakıda etibarlı notebook mağazası."},
		{"footer.storeLabel", "MAĞAZA"},
		{"footer.address", "28 May, Süleyman Rəhimov 201\nBakı, Azərbaycan\nHər gün: 10:00 – 20:00"},
		{"footer.contactLabel", "ƏLAQƏ"},
		{"footer.phones", "+994 70 815 12 83\n+994 55 654 56 95\n+994 12 525 21 09\ninfo@laptops.az"},
		{"footer.copyright", "© 2026 Laptops.az — Bütün hüquqlar qorunur."},
		{"menu.storeInfo", "28 May, Süleyman Rəhimov 201\nHər gün 10:00 – 20:00"},
		{"map.directions", "Yol göstər"},
		{"map.view", "Xəritədə bax"},

		// AI köməkçi
		{"ai.button", "AI Köməkçi"},
		{"ai.status", "Onlayn · dərhal cavablayır"},
		{"ai.welcome", "Salam! 👋 Mən Laptops.az köməkçisiyəm. Sizə uyğun notebook tapmaqda kömək edim — nə üçün istifadə edəcəksiniz?"},
		{"ai.inputPlaceholder", "Sualınızı yazın…"},
		{"ai.error", "Bağışlayın, bir problem oldu. Bir azdan yenidən yoxlayın 🙏"},
		{"ai.quickGameLabel", "🎮 Oyun"},
		{"ai.quickGameQ", "Oyun üçün notebook axtarıram, tövsiyə edin"},
		{"ai.quickOfficeLabel", "💼 İş / ofis"},
		{"ai.quickOfficeQ", "İş və ofis işləri üçün notebook lazımdır"},
		{"ai.quickStudentLabel", "🎓 Tələbə"},
		{"ai.quickStudentQ", "Tələbə üçün büdcəyə uyğun notebook"},
		{"ai.quickDesignLabel", "🎨 Dizayn"},
		{"ai.quickDesignQ", "Dizayn və video montaj üçün güclü notebook"},

		// sifariş modalı
		{"order.title", "Sifarişi tamamla"},
		{"order.qty", "1 ədəd"},
		{"order.nameLabel", "AD, SOYAD"},
		{"order.namePlaceholder", "Adınız və soyadınız"},
		{"order.phoneLabel", "TELEFON"},
		{"order.noteLabel", "QEYD (İSTƏYƏ BAĞLI)"},
		{"order.notePlaceholder", "Məs: sabah günorta gələ bilərəm…"},
		{"order.validation", "Ad və telefon vacibdir"},
		{"order.paymentNote", "Ödəniş yoxdur. Mağaza zəng edib məhsulu 24 saat saxlayacaq."},
		{"order.submit", "Sifarişi təsdiqlə"},
		{"order.submitting", "Göndərilir…"},
		{"order.received", "Sifarişiniz alındı"},
		{"order.receivedText", "mağaza qısa zamanda sizinlə əlaqə saxlayacaq. Məhsul 24 saat sizin üçün saxlanılır."},
		{"order.close", "Bağla"},
	}
}

// seedI18n — dilləri və statik mətnlərin AZ dəyərlərini yaradır (idempotent, əlavə edən).
// Mövcud dəyərləri POZMUR — yalnız çatışmayan açarları/dilləri əlavə edir.
func seedI18n() {
	var n int64
	db.Model(&Language{}).Count(&n)
	if n == 0 {
		db.Create(&[]Language{
			{Code: "az", Name: "Azərbaycan", Enabled: true, IsDefault: true, Sort: 0},
			{Code: "ru", Name: "Русский", Enabled: true, Sort: 1},
			{Code: "tr", Name: "Türkçe", Enabled: true, Sort: 2},
			{Code: "en", Name: "English", Enabled: true, Sort: 3},
		})
		log.Println("i18n: dillər yaradıldı (az, ru, tr, en)")
	}
	// əsas dilin kodu (adətən az)
	def := "az"
	var dl Language
	if err := db.Where("is_default = ?", true).First(&dl).Error; err == nil && dl.Code != "" {
		def = dl.Code
	}
	// çatışmayan AZ mətnlərini əlavə et
	added := 0
	for _, e := range defaultUiStrings() {
		var c int64
		db.Model(&UiString{}).Where("key = ? AND lang = ?", e.Key, def).Count(&c)
		if c == 0 {
			db.Create(&UiString{Key: e.Key, Lang: def, Value: e.Value})
			added++
		}
	}
	if added > 0 {
		log.Printf("i18n: %d yeni mətn açarı əlavə olundu (%s)", added, def)
	}
	// digər dillərə hazır tərcümələr — yalnız çatışmayan (key,lang) sətirləri əlavə edir,
	// admin paneldən edilmiş redaktələri POZMUR.
	tadd := 0
	for lang, m := range defaultTranslations() {
		for key, val := range m {
			var c int64
			db.Model(&UiString{}).Where("key = ? AND lang = ?", key, lang).Count(&c)
			if c == 0 {
				db.Create(&UiString{Key: key, Lang: lang, Value: val})
				tadd++
			}
		}
	}
	if tadd > 0 {
		log.Printf("i18n: %d hazır tərcümə əlavə olundu", tadd)
	}
}

// defaultTranslations — statik mətnlərin rus / türk / ingilis tərcümələri.
// Marka/nömrə siyahıları qəsdən dəyişməz saxlanılıb.
func defaultTranslations() map[string]map[string]string {
	return map[string]map[string]string{
		"ru": {
			"promo.banner":            "Закажите онлайн → оплатите в магазине: Наличные · Карта · Рассрочка",
			"nav.notebook":            "Ноутбуки",
			"nav.used":                "Б/у",
			"nav.desktop":             "Настольные ПК",
			"nav.phone":               "Телефоны",
			"nav.accessory":           "Аксессуары",
			"search.placeholder":      "Поиск…",
			"search.placeholderLong":  "Поиск по бренду, модели или характеристике…",
			"common.available":        "В наличии",
			"common.order":            "Заказать",
			"common.details":          "Подробнее",
			"common.call":             "Позвонить",
			"common.home":             "Главная",
			"common.loading":          "Загрузка…",
			"home.heroEyebrow":        "В Баку с 2009 года",
			"home.heroTitle":          "Идеальный ноутбук\nдля вас. С гарантией.",
			"home.heroText":           "Новые и б/у ноутбуки, телефоны и аксессуары. Выбирайте и заказывайте онлайн — оплата в магазине.",
			"home.ctaBrowse":          "Смотреть ноутбуки",
			"home.ctaVisit":           "Прийти в магазин",
			"home.liveText":           "Наличие обновляется в реальном времени — товар с меткой «В наличии» есть в магазине",
			"home.featuredEyebrow":    "ИЗБРАННОЕ",
			"home.chipAll":            "Все",
			"home.featuredTitle":      "Избранные товары",
			"home.seeAll":             "Смотреть все",
			"home.loadingProducts":    "Загрузка товаров…",
			"home.tileNotebookTitle":  "Ноутбуки",
			"home.tileNotebookDesc":   "HP · Lenovo · ASUS · Apple · Dell · MSI · Huawei",
			"home.tileUsedTitle":      "Б/у",
			"home.tileUsedDesc":       "Проверенные, с гарантией",
			"home.tileDesktopTitle":   "Настольные ПК",
			"home.tileDesktopDesc":    "Сборки для игр и работы",
			"home.tileCollection":     "Смотреть коллекцию",
			"home.tileView":           "Смотреть",
			"home.tilePhoneTitle":     "Телефоны",
			"home.tilePhoneDesc":      "iPhone · Samsung · Xiaomi",
			"home.tileAccessoryTitle": "Аксессуары",
			"home.tileAccessoryDesc":  "Мышь · клавиатура · SSD · монитор · сумка",
			"product.notFound":        "Товар не найден.",
			"product.orderNote":       "Онлайн-оплаты нет. Оформите заказ — проверьте товар в магазине и оплатите наличными, картой или в рассрочку.",
			"product.buybarNote":      "Оплата в магазине — Наличные / Карта / Рассрочка",
			"cat.sortPopular":         "Популярные",
			"cat.sortCheap":           "Дешевле → дороже",
			"cat.sortExp":             "Дороже → дешевле",
			"cat.searchTitle":         "Поиск",
			"cat.allProducts":         "Все товары",
			"cat.productsSuffix":      "товаров · склад в реальном времени",
			"cat.filter":              "Фильтр",
			"cat.filters":             "Фильтры",
			"cat.clear":               "Очистить",
			"cat.price":               "Цена",
			"cat.sort":                "Сортировка",
			"cat.reset":               "Сбросить",
			"cat.noFilter":            "Нет фильтров",
			"cat.noProducts":          "Нет подходящих товаров.",
			"cat.showResults":         "результатов",
			"footer.tagline":          "Надёжный магазин ноутбуков в Баку с 2009 года.",
			"footer.storeLabel":       "МАГАЗИН",
			"footer.address":          "28 May, Сулейман Рагимов 201\nБаку, Азербайджан\nЕжедневно: 10:00 – 20:00",
			"footer.contactLabel":     "КОНТАКТЫ",
			"footer.phones":           "+994 70 815 12 83\n+994 55 654 56 95\n+994 12 525 21 09\ninfo@laptops.az",
			"footer.copyright":        "© 2026 Laptops.az — Все права защищены.",
			"menu.storeInfo":          "28 May, Сулейман Рагимов 201\nЕжедневно 10:00 – 20:00",
			"map.directions":          "Построить маршрут",
			"map.view":                "Показать на карте",
			"ai.button":               "AI-помощник",
			"ai.status":               "Онлайн · отвечает сразу",
			"ai.welcome":              "Здравствуйте! 👋 Я помощник Laptops.az. Помогу подобрать подходящий ноутбук — для чего будете использовать?",
			"ai.inputPlaceholder":     "Напишите ваш вопрос…",
			"ai.error":                "Извините, возникла проблема. Попробуйте чуть позже 🙏",
			"ai.quickGameLabel":       "🎮 Игры",
			"ai.quickGameQ":           "Ищу ноутбук для игр, посоветуйте",
			"ai.quickOfficeLabel":     "💼 Работа / офис",
			"ai.quickOfficeQ":         "Нужен ноутбук для работы и офисных задач",
			"ai.quickStudentLabel":    "🎓 Студент",
			"ai.quickStudentQ":        "Ноутбук для студента по доступной цене",
			"ai.quickDesignLabel":     "🎨 Дизайн",
			"ai.quickDesignQ":         "Мощный ноутбук для дизайна и видеомонтажа",
			"order.title":             "Оформить заказ",
			"order.qty":               "1 шт.",
			"order.nameLabel":         "ИМЯ, ФАМИЛИЯ",
			"order.namePlaceholder":   "Ваше имя и фамилия",
			"order.phoneLabel":        "ТЕЛЕФОН",
			"order.noteLabel":         "ЗАМЕТКА (НЕОБЯЗАТЕЛЬНО)",
			"order.notePlaceholder":   "Напр.: смогу зайти завтра днём…",
			"order.validation":        "Имя и телефон обязательны",
			"order.paymentNote":       "Оплаты нет. Магазин позвонит и придержит товар на 24 часа.",
			"order.submit":            "Подтвердить заказ",
			"order.submitting":        "Отправка…",
			"order.received":          "Заказ принят",
			"order.receivedText":      "магазин свяжется с вами в ближайшее время. Товар удерживается для вас 24 часа.",
			"order.close":             "Закрыть",
		},
		"tr": {
			"promo.banner":            "Online sipariş ver → ödemeyi mağazada yap: Nakit · Kart · Taksit",
			"nav.notebook":            "Notebooklar",
			"nav.used":                "İkinci El",
			"nav.desktop":             "Masaüstü PC",
			"nav.phone":               "Telefonlar",
			"nav.accessory":           "Aksesuarlar",
			"search.placeholder":      "Ara…",
			"search.placeholderLong":  "Marka, model veya özellik ara…",
			"common.available":        "Mevcut",
			"common.order":            "Sipariş ver",
			"common.details":          "Detaylar",
			"common.call":             "Ara",
			"common.home":             "Ana sayfa",
			"common.loading":          "Yükleniyor…",
			"home.heroEyebrow":        "2009'dan beri Bakü'de",
			"home.heroTitle":          "Size uygun\nnotebook. Garantili.",
			"home.heroText":           "Yeni ve ikinci el notebooklar, telefonlar ve aksesuarlar. Online seçin ve sipariş verin — ödemeyi mağazada yapın.",
			"home.ctaBrowse":          "Notebooklara bak",
			"home.ctaVisit":           "Mağazaya gel",
			"home.liveText":           "Stok gerçek zamanlı güncellenir — «Mevcut» görünen ürün mağazada var",
			"home.featuredEyebrow":    "ÖNE ÇIKAN",
			"home.chipAll":            "Tümü",
			"home.featuredTitle":      "Öne çıkan ürünler",
			"home.seeAll":             "Tümünü gör",
			"home.loadingProducts":    "Ürünler yükleniyor…",
			"home.tileNotebookTitle":  "Notebooklar",
			"home.tileNotebookDesc":   "HP · Lenovo · ASUS · Apple · Dell · MSI · Huawei",
			"home.tileUsedTitle":      "İkinci El",
			"home.tileUsedDesc":       "Kontrol edilmiş, garantili",
			"home.tileDesktopTitle":   "Masaüstü PC",
			"home.tileDesktopDesc":    "Oyun ve iş için hazır sistemler",
			"home.tileCollection":     "Koleksiyona bak",
			"home.tileView":           "Bak",
			"home.tilePhoneTitle":     "Telefonlar",
			"home.tilePhoneDesc":      "iPhone · Samsung · Xiaomi",
			"home.tileAccessoryTitle": "Aksesuarlar",
			"home.tileAccessoryDesc":  "Fare · klavye · SSD · monitör · çanta",
			"product.notFound":        "Ürün bulunamadı.",
			"product.orderNote":       "Online ödeme yok. Sipariş verin — ürünü mağazada kontrol edip Nakit, Kart veya Taksit ile ödeyin.",
			"product.buybarNote":      "Ödeme mağazada — Nakit / Kart / Taksit",
			"cat.sortPopular":         "Popüler",
			"cat.sortCheap":           "Ucuz → Pahalı",
			"cat.sortExp":             "Pahalı → Ucuz",
			"cat.searchTitle":         "Arama",
			"cat.allProducts":         "Tüm ürünler",
			"cat.productsSuffix":      "ürün · gerçek zamanlı stok",
			"cat.filter":              "Filtre",
			"cat.filters":             "Filtreler",
			"cat.clear":               "Temizle",
			"cat.price":               "Fiyat",
			"cat.sort":                "Sıralama",
			"cat.reset":               "Sıfırla",
			"cat.noFilter":            "Filtre yok",
			"cat.noProducts":          "Uygun ürün yok.",
			"cat.showResults":         "sonucu göster",
			"footer.tagline":          "2009'dan beri Bakü'de güvenilir notebook mağazası.",
			"footer.storeLabel":       "MAĞAZA",
			"footer.address":          "28 May, Süleyman Rəhimov 201\nBakü, Azerbaycan\nHer gün: 10:00 – 20:00",
			"footer.contactLabel":     "İLETİŞİM",
			"footer.phones":           "+994 70 815 12 83\n+994 55 654 56 95\n+994 12 525 21 09\ninfo@laptops.az",
			"footer.copyright":        "© 2026 Laptops.az — Tüm hakları saklıdır.",
			"menu.storeInfo":          "28 May, Süleyman Rəhimov 201\nHer gün 10:00 – 20:00",
			"map.directions":          "Yol tarifi",
			"map.view":                "Haritada göster",
			"ai.button":               "AI Asistan",
			"ai.status":               "Çevrimiçi · anında yanıt",
			"ai.welcome":              "Merhaba! 👋 Ben Laptops.az asistanıyım. Size uygun notebook bulmanıza yardım edeyim — ne için kullanacaksınız?",
			"ai.inputPlaceholder":     "Sorunuzu yazın…",
			"ai.error":                "Üzgünüz, bir sorun oluştu. Biraz sonra tekrar deneyin 🙏",
			"ai.quickGameLabel":       "🎮 Oyun",
			"ai.quickGameQ":           "Oyun için notebook arıyorum, önerir misiniz",
			"ai.quickOfficeLabel":     "💼 İş / ofis",
			"ai.quickOfficeQ":         "İş ve ofis işleri için notebook lazım",
			"ai.quickStudentLabel":    "🎓 Öğrenci",
			"ai.quickStudentQ":        "Öğrenci için bütçeye uygun notebook",
			"ai.quickDesignLabel":     "🎨 Tasarım",
			"ai.quickDesignQ":         "Tasarım ve video montaj için güçlü notebook",
			"order.title":             "Siparişi tamamla",
			"order.qty":               "1 adet",
			"order.nameLabel":         "AD, SOYAD",
			"order.namePlaceholder":   "Adınız ve soyadınız",
			"order.phoneLabel":        "TELEFON",
			"order.noteLabel":         "NOT (İSTEĞE BAĞLI)",
			"order.notePlaceholder":   "Örn: yarın öğlen gelebilirim…",
			"order.validation":        "Ad ve telefon gerekli",
			"order.paymentNote":       "Ödeme yok. Mağaza arayıp ürünü 24 saat saklayacak.",
			"order.submit":            "Siparişi onayla",
			"order.submitting":        "Gönderiliyor…",
			"order.received":          "Siparişiniz alındı",
			"order.receivedText":      "mağaza kısa sürede sizinle iletişime geçecek. Ürün 24 saat sizin için saklanıyor.",
			"order.close":             "Kapat",
		},
		"en": {
			"promo.banner":            "Order online → pay in store: Cash · Card · Installment",
			"nav.notebook":            "Laptops",
			"nav.used":                "Used",
			"nav.desktop":             "Desktops",
			"nav.phone":               "Phones",
			"nav.accessory":           "Accessories",
			"search.placeholder":      "Search…",
			"search.placeholderLong":  "Search by brand, model or spec…",
			"common.available":        "In stock",
			"common.order":            "Order",
			"common.details":          "Details",
			"common.call":             "Call",
			"common.home":             "Home",
			"common.loading":          "Loading…",
			"home.heroEyebrow":        "In Baku since 2009",
			"home.heroTitle":          "The right laptop\nfor you. With warranty.",
			"home.heroText":           "New and used laptops, phones and accessories. Choose and order online — pay in store.",
			"home.ctaBrowse":          "Browse laptops",
			"home.ctaVisit":           "Visit the store",
			"home.liveText":           "Stock updates in real time — items marked «In stock» are available in the store",
			"home.featuredEyebrow":    "FEATURED",
			"home.chipAll":            "All",
			"home.featuredTitle":      "Featured products",
			"home.seeAll":             "See all",
			"home.loadingProducts":    "Loading products…",
			"home.tileNotebookTitle":  "Laptops",
			"home.tileNotebookDesc":   "HP · Lenovo · ASUS · Apple · Dell · MSI · Huawei",
			"home.tileUsedTitle":      "Used",
			"home.tileUsedDesc":       "Checked, with warranty",
			"home.tileDesktopTitle":   "Desktops",
			"home.tileDesktopDesc":    "Desktop builds for gaming and work",
			"home.tileCollection":     "View collection",
			"home.tileView":           "View",
			"home.tilePhoneTitle":     "Phones",
			"home.tilePhoneDesc":      "iPhone · Samsung · Xiaomi",
			"home.tileAccessoryTitle": "Accessories",
			"home.tileAccessoryDesc":  "Mouse · keyboard · SSD · monitor · bag",
			"product.notFound":        "Product not found.",
			"product.orderNote":       "No online payment. Place an order — check the item in store and pay by cash, card or installment.",
			"product.buybarNote":      "Payment in store — Cash / Card / Installment",
			"cat.sortPopular":         "Popular",
			"cat.sortCheap":           "Cheap → Expensive",
			"cat.sortExp":             "Expensive → Cheap",
			"cat.searchTitle":         "Search",
			"cat.allProducts":         "All products",
			"cat.productsSuffix":      "products · real-time stock",
			"cat.filter":              "Filter",
			"cat.filters":             "Filters",
			"cat.clear":               "Clear",
			"cat.price":               "Price",
			"cat.sort":                "Sort",
			"cat.reset":               "Reset",
			"cat.noFilter":            "No filters",
			"cat.noProducts":          "No matching products.",
			"cat.showResults":         "results",
			"footer.tagline":          "A trusted laptop store in Baku since 2009.",
			"footer.storeLabel":       "STORE",
			"footer.address":          "28 May, Suleyman Rahimov 201\nBaku, Azerbaijan\nEvery day: 10:00 – 20:00",
			"footer.contactLabel":     "CONTACT",
			"footer.phones":           "+994 70 815 12 83\n+994 55 654 56 95\n+994 12 525 21 09\ninfo@laptops.az",
			"footer.copyright":        "© 2026 Laptops.az — All rights reserved.",
			"menu.storeInfo":          "28 May, Suleyman Rahimov 201\nEvery day 10:00 – 20:00",
			"map.directions":          "Get directions",
			"map.view":                "View on map",
			"ai.button":               "AI Assistant",
			"ai.status":               "Online · instant replies",
			"ai.welcome":              "Hi! 👋 I'm the Laptops.az assistant. Let me help you find the right laptop — what will you use it for?",
			"ai.inputPlaceholder":     "Type your question…",
			"ai.error":                "Sorry, something went wrong. Please try again in a moment 🙏",
			"ai.quickGameLabel":       "🎮 Gaming",
			"ai.quickGameQ":           "I'm looking for a gaming laptop, please recommend one",
			"ai.quickOfficeLabel":     "💼 Work / office",
			"ai.quickOfficeQ":         "I need a laptop for work and office tasks",
			"ai.quickStudentLabel":    "🎓 Student",
			"ai.quickStudentQ":        "A budget-friendly laptop for a student",
			"ai.quickDesignLabel":     "🎨 Design",
			"ai.quickDesignQ":         "A powerful laptop for design and video editing",
			"order.title":             "Complete order",
			"order.qty":               "1 pc",
			"order.nameLabel":         "FULL NAME",
			"order.namePlaceholder":   "Your first and last name",
			"order.phoneLabel":        "PHONE",
			"order.noteLabel":         "NOTE (OPTIONAL)",
			"order.notePlaceholder":   "E.g.: I can come tomorrow at noon…",
			"order.validation":        "Name and phone are required",
			"order.paymentNote":       "No payment now. The store will call and hold the item for 24 hours.",
			"order.submit":            "Confirm order",
			"order.submitting":        "Sending…",
			"order.received":          "Order received",
			"order.receivedText":      "the store will contact you shortly. The item is held for you for 24 hours.",
			"order.close":             "Close",
		},
	}
}
