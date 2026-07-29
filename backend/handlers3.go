package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func decodeBody(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// telefonu müqayisə üçün normallaşdır — yalnız rəqəmlər ("050 123" == "0501 23")
func normPhone(p string) string {
	var b strings.Builder
	for _, r := range p {
		if r >= '0' && r <= '9' {
			b.WriteByte(byte(r))
		}
	}
	return b.String()
}

// verilmiş telefonlu müştəri (excludeID istisna) — varsa qaytarır
func customerByPhone(phone string, excludeID uint) *Customer {
	np := normPhone(phone)
	if np == "" {
		return nil
	}
	var all []Customer
	db.Where("phone <> ''").Find(&all)
	for i := range all {
		if all[i].ID != excludeID && normPhone(all[i].Phone) == np {
			return &all[i]
		}
	}
	return nil
}

// POST /api/customers
func createCustomer(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name, Phone string }
	if err := decodeBody(r, &in); err != nil || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "ad vacibdir"})
		return
	}
	if ex := customerByPhone(in.Phone, 0); ex != nil {
		writeJSON(w, 409, map[string]any{
			"error":         "Bu nömrə artıq var: " + ex.Name,
			"existing_id":   ex.ID,
			"existing_name": ex.Name,
		})
		return
	}
	c := Customer{Name: in.Name, Phone: in.Phone, CreatedAt: time.Now()}
	db.Create(&c)
	writeJSON(w, 201, c)
}

// POST /api/expenses
func createExpense(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Category, Note, Date string
		Amount               float64
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	d := time.Now()
	if in.Date != "" {
		if t, e := time.Parse("2006-01-02", in.Date); e == nil {
			d = t
		}
	}
	e := Expense{Date: d, Category: in.Category, Note: in.Note, Amount: in.Amount}
	db.Create(&e)
	writeJSON(w, 201, e)
}

// PUT /api/expenses/{id} — xərci redaktə et
func updateExpense(w http.ResponseWriter, r *http.Request) {
	var e Expense
	if err := db.First(&e, r.PathValue("id")).Error; err != nil {
		writeJSON(w, 404, map[string]string{"error": "xərc tapılmadı"})
		return
	}
	var in struct {
		Category, Note, Date string
		Amount               float64
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	upd := map[string]any{"category": in.Category, "note": in.Note, "amount": in.Amount}
	if in.Date != "" {
		if t, e2 := time.Parse("2006-01-02", in.Date); e2 == nil {
			upd["date"] = t
		}
	}
	db.Model(&e).Updates(upd)
	db.First(&e, e.ID)
	writeJSON(w, 200, e)
}

// DELETE /api/expenses/{id} — bir xərci sil (yalnız admin). id="all" → hamısını sil.
func deleteExpense(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromReq(r)
	if !ok || u.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "yalnız admin"})
		return
	}
	id := r.PathValue("id")
	if id == "all" {
		res := db.Where("1 = 1").Delete(&Expense{})
		if res.Error != nil {
			writeJSON(w, 500, map[string]string{"error": res.Error.Error()})
			return
		}
		writeJSON(w, 200, map[string]int{"deleted": int(res.RowsAffected)})
		return
	}
	res := db.Delete(&Expense{}, id)
	if res.Error != nil {
		writeJSON(w, 500, map[string]string{"error": res.Error.Error()})
		return
	}
	writeJSON(w, 200, map[string]int{"deleted": int(res.RowsAffected)})
}

// POST /api/categories  {name, attribute_ids:[]}
func createCategory(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name         string `json:"name"`
		AttributeIDs []uint `json:"attribute_ids"`
	}
	if err := decodeBody(r, &in); err != nil || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "ad vacibdir"})
		return
	}
	c := Category{Name: in.Name}
	db.Create(&c)
	for _, aid := range in.AttributeIDs {
		db.Exec("INSERT INTO category_attributes (category_id, attribute_id) VALUES (?, ?)", c.ID, aid)
	}
	db.Preload("Attributes.Options").First(&c, c.ID)
	writeJSON(w, 201, c)
}

// POST /api/attributes  {name, options:[]}
func createAttribute(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name    string   `json:"name"`
		Options []string `json:"options"`
	}
	if err := decodeBody(r, &in); err != nil || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "ad vacibdir"})
		return
	}
	a := Attribute{Name: in.Name}
	for _, o := range in.Options {
		if v := strings.TrimSpace(o); v != "" {
			a.Options = append(a.Options, AttributeOption{Value: v})
		}
	}
	db.Create(&a)
	writeJSON(w, 201, a)
}

// POST /api/attributes/{id}/options  {value}
func addOption(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	var in struct{ Value string }
	if err := decodeBody(r, &in); err != nil || in.Value == "" {
		writeJSON(w, 400, map[string]string{"error": "dəyər vacibdir"})
		return
	}
	o := AttributeOption{AttributeID: uint(id), Value: in.Value}
	db.Create(&o)
	writeJSON(w, 201, o)
}

// POST /api/branches
func createBranch(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name, Address string }
	if err := decodeBody(r, &in); err != nil || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "ad vacibdir"})
		return
	}
	b := Branch{Name: in.Name, Address: in.Address}
	db.Create(&b)
	writeJSON(w, 201, b)
}

// POST /api/transfers  {item_id, branch_id} — cihazı filiala köçür
func createTransfer(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ItemID   uint `json:"item_id"`
		BranchID uint `json:"branch_id"`
	}
	if err := decodeBody(r, &in); err != nil || in.ItemID == 0 || in.BranchID == 0 {
		writeJSON(w, 400, map[string]string{"error": "cihaz və filial vacibdir"})
		return
	}
	db.Model(&Item{}).Where("id = ?", in.ItemID).Update("branch_id", in.BranchID)
	var it Item
	db.Preload("Branch").First(&it, in.ItemID)
	writeJSON(w, 200, it)
}

// POST /api/supplies
func createSupply(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Source    string  `json:"source"`
		Supplier  string  `json:"supplier"`
		Status    string  `json:"status"`
		ItemCount int     `json:"item_count"`
		TotalCost float64 `json:"total_cost"`
	}
	if err := decodeBody(r, &in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "yanlış məlumat"})
		return
	}
	if in.Status == "" {
		in.Status = "pending"
	}
	s := SupplyBatch{Date: time.Now(), Source: in.Source, Supplier: in.Supplier, ItemCount: in.ItemCount, TotalCost: in.TotalCost, Status: in.Status}
	db.Create(&s)
	writeJSON(w, 201, s)
}

// POST /api/consignments  (realizasiya — başqa mağazaya mal ver)
func createConsignment(w http.ResponseWriter, r *http.Request) {
	var in struct {
		StoreName  string  `json:"store_name"`
		ItemID     *uint   `json:"item_id"`
		ItemName   string  `json:"item_name"`
		Serial     string  `json:"serial"`
		Cost       float64 `json:"cost"`
		GivenPrice float64 `json:"given_price"`
	}
	if err := decodeBody(r, &in); err != nil || in.StoreName == "" {
		writeJSON(w, 400, map[string]string{"error": "mağaza adı vacibdir"})
		return
	}
	c := Consignment{StoreName: in.StoreName, GivenAt: time.Now(), Status: "out"}
	if in.ItemID != nil {
		var it Item
		if err := db.First(&it, *in.ItemID).Error; err != nil {
			writeJSON(w, 404, map[string]string{"error": "cihaz tapılmadı"})
			return
		}
		c.ItemID = in.ItemID
		c.ItemName = it.Name
		c.Serial = it.Serial
		c.Cost = it.Cost
		c.GivenPrice = in.GivenPrice
		if c.GivenPrice == 0 {
			c.GivenPrice = it.Price
		}
		c.Debt = it.Cost
		// cihaz stokdan çıxır → «realizasiyada»
		db.Model(&Item{}).Where("id = ?", *in.ItemID).Update("status", "consignment")
	} else {
		c.ItemName = in.ItemName
		c.Serial = in.Serial
		c.Cost = in.Cost
		c.GivenPrice = in.GivenPrice
		c.Debt = in.Cost
	}
	db.Create(&c)
	writeJSON(w, 201, c)
}
