package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// məhsul formasında xüsusiyyətlərin görünmə sırası
var attrRank = map[string]int{
	"Marka": 0, "Prosessor": 1, "RAM": 2, "SSD": 3, "Ekran kartı": 4,
	"Ekran": 5, "Yaddaş": 6, "Rəng": 7, "Ölçü": 8, "Tezlik": 9,
	"Növ": 10, "Vəziyyət": 11,
}

func attrRankOf(name string) int {
	if r, ok := attrRank[name]; ok {
		return r
	}
	return 99
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// GET /api/dashboard
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// filial təcridi — satıcı yalnız öz filialının rəqəmlərini görür; admin hamısını
	bid, restricted := branchScope(r)

	// stok/anbar bütün filiallara açıqdır (satış üçün lazımdır) — qlobal
	var stockCount, reservedCount int64
	var stockValue float64
	db.Model(&Item{}).Where("status = ?", "in_stock").Count(&stockCount)
	db.Model(&Item{}).Where("status = ?", "reserved").Count(&reservedCount)
	db.Model(&Item{}).Where("status = ?", "in_stock").Select("COALESCE(SUM(cost * quantity),0)").Scan(&stockValue)

	var openCredit float64
	var overdue int64
	db.Model(&CreditPlan{}).Select("COALESCE(SUM(total - paid),0)").Scan(&openCredit)
	db.Model(&CreditPlan{}).Where("status = ?", "overdue").Count(&overdue)

	var salesCount int64
	var turnover, profit float64
	sq1 := db.Model(&Sale{}).Where("counted = ?", true)
	sq2 := db.Model(&Sale{}).Where("counted = ?", true)
	sq3 := db.Model(&Sale{}).Where("counted = ?", true)
	if restricted {
		sq1 = sq1.Where("branch_id = ?", bid)
		sq2 = sq2.Where("branch_id = ?", bid)
		sq3 = sq3.Where("branch_id = ?", bid)
	}
	sq1.Count(&salesCount)
	sq2.Select("COALESCE(SUM(sale_price),0)").Scan(&turnover)
	sq3.Select("COALESCE(SUM(profit),0)").Scan(&profit)

	var expenses float64
	db.Model(&Expense{}).Select("COALESCE(SUM(amount),0)").Scan(&expenses)

	type catRow struct {
		Name     string  `json:"name"`
		Turnover float64 `json:"turnover"`
	}
	var byCat []catRow
	catQ := db.Model(&Sale{}).Select("categories.name as name, COALESCE(SUM(sales.sale_price),0) as turnover").
		Joins("JOIN items ON items.id = sales.item_id").
		Joins("JOIN categories ON categories.id = items.category_id").
		Where("sales.counted = ?", true)
	if restricted {
		catQ = catQ.Where("sales.branch_id = ?", bid)
	}
	catQ.Group("categories.name").Order("turnover desc").Limit(5).Scan(&byCat)

	writeJSON(w, 200, map[string]any{
		"stock_count":    stockCount,
		"stock_value":    stockValue,
		"reserved_count": reservedCount,
		"open_credit":    openCredit,
		"overdue_count":  overdue,
		"sales_count":    salesCount,
		"turnover":       turnover,
		"gross_profit":   profit,
		"expenses":       expenses,
		"net_profit":     profit - expenses,
		"by_category":    byCat,
	})
}

// GET /api/items?status=&category_id=&branch_id=&q=
func listItems(w http.ResponseWriter, r *http.Request) {
	q := db.Preload("Category").Preload("Branch").Preload("Values.Attribute").Preload("Translations").Order("created_at desc")
	if s := r.URL.Query().Get("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if c := r.URL.Query().Get("category_id"); c != "" {
		q = q.Where("category_id = ?", c)
	}
	if b := r.URL.Query().Get("branch_id"); b != "" {
		q = q.Where("branch_id = ?", b)
	}
	if term := r.URL.Query().Get("q"); term != "" {
		like := "%" + term + "%"
		// ad, seriya VƏ YA xüsusiyyət dəyəri üzrə
		q = q.Where("name LIKE ? OR serial LIKE ? OR id IN (?)", like, like,
			db.Model(&ItemAttributeValue{}).Select("item_id").Where("value LIKE ?", like))
	}
	var items []Item
	q.Find(&items)
	// satılan cihazlar üçün satış tarixini əlavə et (məhsul redaktəsində göstərmək üçün)
	var soldIDs []uint
	for _, it := range items {
		if it.Status == "sold" {
			soldIDs = append(soldIDs, it.ID)
		}
	}
	if len(soldIDs) > 0 {
		type saleRow struct {
			ItemID uint
			SoldAt time.Time
		}
		var srs []saleRow
		db.Model(&Sale{}).Select("item_id, sold_at").Where("item_id IN ?", soldIDs).Find(&srs)
		m := map[uint]time.Time{}
		for _, s := range srs {
			m[s.ItemID] = s.SoldAt
		}
		for i := range items {
			if t, ok := m[items[i].ID]; ok {
				tt := t
				items[i].SoldAt = &tt
			}
		}
	}
	writeJSON(w, 200, items)
}

// POST /api/items
func createItem(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name           string   `json:"name"`
		Serial         string   `json:"serial"`
		CategoryID     uint     `json:"category_id"`
		BranchID       uint     `json:"branch_id"`
		Cost           float64  `json:"cost"`
		Price          float64  `json:"price"`
		WholesalePrice float64  `json:"wholesale_price"`
		Discount       float64  `json:"discount"`
		Quantity       int      `json:"quantity"` // stokdakı ədəd (boş/0 → 1)
		ShowOnSite     bool     `json:"show_on_site"`
		CardImage      string   `json:"card_image"`
		Gallery        string   `json:"gallery"`
		Translations   []itemTr `json:"translations"`
		Values         []struct {
			AttributeID uint   `json:"attribute_id"`
			Value       string `json:"value"`
		} `json:"values"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	// satıcı (admin deyil) yalnız ÖZ filialına məhsul əlavə edə bilər — filial məcburi öz filialı
	if u, ok := userFromReq(r); ok && u.Role != "admin" && u.BranchID != 0 {
		in.BranchID = u.BranchID
	}
	qty := in.Quantity
	if qty < 1 {
		qty = 1
	}
	it := Item{
		Name: in.Name, Serial: in.Serial, CategoryID: in.CategoryID,
		BranchID: in.BranchID, Cost: in.Cost, Price: in.Price, Status: "in_stock",
		WholesalePrice: in.WholesalePrice, Discount: in.Discount, Quantity: qty,
		ShowOnSite: in.ShowOnSite, CardImage: in.CardImage, Gallery: in.Gallery, CreatedAt: time.Now(),
	}
	for _, v := range in.Values {
		it.Values = append(it.Values, ItemAttributeValue{AttributeID: v.AttributeID, Value: v.Value})
	}
	if err := db.Omit("Values.Attribute").Create(&it).Error; err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	saveItemTranslations(it.ID, in.Translations)
	db.Preload("Category").Preload("Branch").Preload("Values.Attribute").Preload("Translations").First(&it, it.ID)
	// yeni məhsul saytda görünürsə → telefonlara push (arxa planda)
	if it.ShowOnSite {
		go sendNewProductPush("Yeni məhsul 🔔", it.Name)
	}
	writeJSON(w, 201, it)
}

// GET /api/categories
func listCategories(w http.ResponseWriter, r *http.Request) {
	var cats []Category
	db.Preload("Attributes.Options", optOrder).Find(&cats)
	// xüsusiyyətləri sabit sıraya sal (Marka öndə, Vəziyyət sonda)
	for i := range cats {
		sort.SliceStable(cats[i].Attributes, func(a, b int) bool {
			return attrRankOf(cats[i].Attributes[a].Name) < attrRankOf(cats[i].Attributes[b].Name)
		})
	}
	writeJSON(w, 200, cats)
}

// GET /api/attributes
func listAttributes(w http.ResponseWriter, r *http.Request) {
	var attrs []Attribute
	db.Preload("Options", optOrder).Find(&attrs)
	writeJSON(w, 200, attrs)
}

// GET /api/branches
func listBranches(w http.ResponseWriter, r *http.Request) {
	var b []Branch
	db.Find(&b)
	writeJSON(w, 200, b)
}

// GET /api/sales?limit=&offset=&q=  (pagination — 40k+ sətir üçün)
func listSales(w http.ResponseWriter, r *http.Request) {
	limit, offset := 50, 0
	if v, e := strconv.Atoi(r.URL.Query().Get("limit")); e == nil && v > 0 && v <= 500 {
		limit = v
	}
	if v, e := strconv.Atoi(r.URL.Query().Get("offset")); e == nil && v >= 0 {
		offset = v
	}
	term := r.URL.Query().Get("q")
	cat := r.URL.Query().Get("category")
	ch := r.URL.Query().Get("channel")
	// "sayılan" satışlar + təsdiq gözləyənlər (pending) — pending-lər ən üstdə.
	// bağlanmamış kredit satışları (counted=false, pending=false) görünmür.
	q := db.Preload("Item", unscoped).Preload("Item.Category").Preload("Item.Branch").Preload("Customer").
		Where("sales.counted = ? OR sales.pending = ?", true, true).
		Order("sales.pending desc, sales.sold_at desc")
	if term != "" || (cat != "" && cat != "all") {
		q = q.Joins("JOIN items ON items.id = sales.item_id")
	}
	if term != "" {
		like := "%" + term + "%"
		q = q.Where("items.name LIKE ? OR items.serial LIKE ?", like, like)
	}
	if cat != "" && cat != "all" {
		q = q.Joins("JOIN categories ON categories.id = items.category_id").Where("categories.name = ?", cat)
	}
	if ch != "" && ch != "all" {
		q = q.Where("sales.channel = ?", ch)
	}
	// filial təcridi — satıcı yalnız öz filialının satışını görür; admin hamısını (istəsə seçər)
	if bid, restricted := branchScope(r); restricted {
		q = q.Where("sales.branch_id = ?", bid)
	}
	var sales []Sale
	q.Limit(limit).Offset(offset).Find(&sales)
	writeJSON(w, 200, sales)
}

// PUT /api/sales/{id}/confirm  {sale_price} — təsdiq gözləyən satışı təsdiqlə.
// Qiymət ƏL İLƏ yazılır; təsdiq anı satış tarixi olur; mənfəət = qiymət − alış.
func confirmSale(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		SalePrice float64 `json:"sale_price"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	var sale Sale
	if err := db.First(&sale, id).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "satış tapılmadı"})
		return
	}
	if !sale.Pending {
		writeJSON(w, 409, map[string]string{"error": "bu satış artıq təsdiqlənib"})
		return
	}
	qty := sale.Quantity
	if qty < 1 {
		qty = 1
	}
	var cost float64
	db.Unscoped().Model(&Item{}).Select("cost").Where("id = ?", sale.ItemID).Scan(&cost)
	db.Model(&sale).Updates(map[string]any{
		"sale_price": in.SalePrice,
		"profit":     in.SalePrice - cost*float64(qty),
		"pending":    false,
		"counted":    true,
		"sold_at":    time.Now(), // təsdiq anı = satış tarixi
	})
	db.Preload("Item", unscoped).Preload("Item.Category").Preload("Item.Branch").Preload("Customer").First(&sale, sale.ID)
	writeJSON(w, 200, sale)
}

// GET /api/sales/summary?category=&channel=&q= — filtrə uyğun CƏMİ (səhifələmə yox)
// Yuxarıdakı KPI kartları üçün: seçilmiş filtrə uyğun bütün satışların cəmi.
func salesSummary(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("q")
	cat := r.URL.Query().Get("category")
	ch := r.URL.Query().Get("channel")
	q := db.Model(&Sale{}).Where("sales.counted = ?", true)
	if term != "" || (cat != "" && cat != "all") {
		q = q.Joins("JOIN items ON items.id = sales.item_id")
	}
	if term != "" {
		like := "%" + term + "%"
		q = q.Where("items.name LIKE ? OR items.serial LIKE ?", like, like)
	}
	if cat != "" && cat != "all" {
		q = q.Joins("JOIN categories ON categories.id = items.category_id").Where("categories.name = ?", cat)
	}
	if ch != "" && ch != "all" {
		q = q.Where("sales.channel = ?", ch)
	}
	if bid, restricted := branchScope(r); restricted {
		q = q.Where("sales.branch_id = ?", bid)
	}
	var res struct {
		Count    int64   `json:"count"`
		Turnover float64 `json:"turnover"`
		Profit   float64 `json:"profit"`
	}
	q.Select("COUNT(*) as count, COALESCE(SUM(sales.sale_price),0) as turnover, COALESCE(SUM(sales.profit),0) as profit").Scan(&res)
	writeJSON(w, 200, res)
}

// POST /api/sales — satış bağla, cihazı "sold" et, mənfəəti hesabla
func createSale(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ItemID         uint    `json:"item_id"`
		SalePrice      float64 `json:"sale_price"` // BİR ədədin qiyməti
		Quantity       int     `json:"quantity"`   // neçə ədəd (boş/0 → 1)
		Channel        string  `json:"channel"`
		CustomerID     *uint   `json:"customer_id"`
		WarrantyMonths int     `json:"warranty_months"`
		DownPayment    float64 `json:"down_payment"` // kredit: ilkin ödəniş
		NextDue        string  `json:"next_due"`     // kredit: növbəti ödəniş tarixi
		OrderID        uint    `json:"order_id"`     // onlayn sifarişdən satılırsa — həmin sifariş
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	var item Item
	if err := db.First(&item, in.ItemID).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "cihaz tapılmadı"})
		return
	}
	// stokda mövcud say (köhnə qeydlərdə 0 ola bilər → 1 kimi qəbul et, geriyə uyğunluq)
	avail := item.Quantity
	if avail <= 0 && item.Status != "sold" {
		avail = 1
	}
	if item.Status == "sold" || avail <= 0 {
		writeJSON(w, 409, map[string]string{"error": "bu məhsul artıq satılıb / stokda yoxdur"})
		return
	}
	qty := in.Quantity
	if qty < 1 {
		qty = 1
	}
	if qty > avail {
		writeJSON(w, 409, map[string]string{"error": fmt.Sprintf("stokda kifayət qədər yoxdur (var: %d)", avail)})
		return
	}
	total := in.SalePrice * float64(qty) // satışın cəmi məbləği
	// kredit satışı əlimizə pul gəlmədiyi üçün "sayılmayan" başlayır (ödəniş
	// tam olana və ya əl ilə bağlanana qədər satılanlarda görünmür)
	counted := true
	if in.Channel == "credit" {
		counted = in.DownPayment >= total
	}
	sale := Sale{
		ItemID: item.ID, SalePrice: total, Quantity: qty, Profit: total - item.Cost*float64(qty),
		Channel: in.Channel, CustomerID: in.CustomerID, BranchID: item.BranchID,
		WarrantyMonths: in.WarrantyMonths, SoldAt: time.Now(), Counted: counted,
	}
	db.Create(&sale)
	// stoku azalt; 0-a düşəndə "satıldı"
	if avail-qty <= 0 {
		db.Model(&item).Updates(map[string]any{"quantity": 0, "status": "sold"})
	} else {
		db.Model(&item).Update("quantity", avail-qty)
	}

	// onlayn sifarişdən satılıbsa, həmin sifarişi bağla (gözləyənlərdən çıxsın)
	if in.OrderID != 0 {
		db.Model(&OnlineOrder{}).Where("id = ?", in.OrderID).Update("status", "converted")
	}

	// Kredit (nisyə) satışı → avtomatik borc qeydi yarat (müştəri lazımdır)
	if in.Channel == "credit" && in.CustomerID != nil {
		due := time.Now().AddDate(0, 1, 0)
		if in.NextDue != "" {
			if t, e := time.Parse("2006-01-02", in.NextDue); e == nil {
				due = t
			}
		}
		cp := CreditPlan{CustomerID: *in.CustomerID, ItemName: item.Name, Total: total, Paid: in.DownPayment, NextDue: due, Status: creditStatus(due, in.DownPayment, total), SaleID: &sale.ID}
		db.Create(&cp)
		if in.DownPayment > 0 {
			db.Create(&CreditPayment{CreditPlanID: cp.ID, Amount: in.DownPayment, CreatedAt: time.Now()})
		}
	}

	db.Preload("Item").Preload("Customer").First(&sale, sale.ID)
	writeJSON(w, 201, sale)
}

// GET /api/customers
func listCustomers(w http.ResponseWriter, r *http.Request) {
	var c []Customer
	db.Order("name").Find(&c)
	writeJSON(w, 200, c)
}

// GET /api/credits
func listCredits(w http.ResponseWriter, r *http.Request) {
	var c []CreditPlan
	db.Preload("Customer").Preload("Payments").Order("next_due").Find(&c)
	writeJSON(w, 200, c)
}

// GET /api/consignments
func listConsignments(w http.ResponseWriter, r *http.Request) {
	var c []Consignment
	db.Order("given_at desc").Find(&c)
	writeJSON(w, 200, c)
}

// GET /api/expenses
func listExpenses(w http.ResponseWriter, r *http.Request) {
	var e []Expense
	db.Order("date desc").Find(&e)
	writeJSON(w, 200, e)
}
