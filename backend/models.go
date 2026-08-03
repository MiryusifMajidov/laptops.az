package main

import "time"

// Branch — mağaza / filiallar (Mərkəz, Elçin & Rəşid, Zaur…)
type Branch struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `json:"name"`
	IsMain  bool   `json:"is_main"`
	Address string `json:"address"`
}

// Session — DB-də saxlanan giriş sessiyası. Yaddaşda deyil DB-də olduğu üçün
// server restart/auto-stop-dan sonra da qalır (istifadəçi yenidən login etmir).
type Session struct {
	Token     string    `gorm:"primaryKey" json:"token"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Name      string    `json:"name"`
	BranchID  uint      `json:"branch_id"` // restart-dan sonra filial təcridi üçün
	ExpiresAt time.Time `json:"expires_at"`
}

// Category — Notebook, İşlənmiş, Telefon, Aksesuar, Yığım (PC), Monitor
type Category struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	Name       string      `json:"name"`
	Attributes []Attribute `gorm:"many2many:category_attributes;" json:"attributes"`
}

// Attribute — Marka, RAM, SSD, Ekran kartı, Ekran, Vəziyyət…
type Attribute struct {
	ID          uint              `gorm:"primaryKey" json:"id"`
	Name        string            `json:"name"`
	Multiselect bool              `json:"multiselect"` // çox seçimli — məhsulda bir neçə option seçmək olar
	ShowOnSite  bool              `gorm:"default:true" json:"show_on_site"` // saytda/mobil appda göstərilsin (default: bəli)
	Options     []AttributeOption `json:"options"`
}

// AttributeOption — bir xüsusiyyətin seçimləri (RAM: 8/16/32 GB…)
type AttributeOption struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	AttributeID uint   `json:"attribute_id"`
	Value       string `json:"value"`
	Position    int    `json:"position"` // sıralama (kiçik = əvvəl); 0 = sırasız → id ilə düzülür
}

// Item — BİR fiziki cihaz. Stok və satış eyni qeyddir; status dəyişir.
// status: in_stock | reserved | sold | returned
type Item struct {
	ID         uint                 `gorm:"primaryKey" json:"id"`
	Name       string               `json:"name"`
	Serial     string               `json:"serial"`
	CategoryID uint                 `json:"category_id"`
	Category   Category             `json:"category"`
	BranchID   uint                 `json:"branch_id"`
	Branch     Branch               `json:"branch"`
	Cost           float64          `json:"cost"`            // alış (daxili — müştəri görmür)
	Price          float64          `json:"price"`           // satış/sayt qiyməti (müştəriyə göstərilir)
	WholesalePrice float64          `json:"wholesale_price"` // topdan / optavoy qiymət (daxili — müştəri görmür)
	Discount       float64          `json:"discount"`        // ₼ ilə endirim (0 = yoxdur; sayt qiymətdən çıxır)
	Quantity   int                  `gorm:"default:1" json:"quantity"` // stokdakı ədəd sayı (serialı mal = 1; aksesuar = çox). 0 → bitib
	Status     string               `gorm:"default:in_stock" json:"status"`
	ShowOnSite bool                 `json:"show_on_site"`
	CardImage  string               `json:"card_image"`     // sayt kartında görünən şəkil
	Gallery    string               `json:"gallery"`        // məhsul səhifəsi — JSON array of URLs
	Values     []ItemAttributeValue `json:"values"`
	Translations []ItemTranslation  `json:"translations,omitempty"` // məhsul adının dil tərcümələri
	CreatedAt  time.Time            `json:"created_at"` // alınma / gəlmə tarixi
	// satılan cihaz üçün satış tarixi (DB-də saxlanılmır — satışdan doldurulur)
	SoldAt *time.Time `gorm:"-" json:"sold_at,omitempty"`
}

// ItemAttributeValue — cihazın konkret xüsusiyyət dəyəri (RAM = 16 GB)
type ItemAttributeValue struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ItemID      uint      `json:"item_id"`
	AttributeID uint      `json:"attribute_id"`
	Attribute   Attribute `json:"attribute"`
	Value       string    `json:"value"`
}

// Customer — müştəri profili
type Customer struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

// Sale — satış (kanal: cash | card | installment | credit)
type Sale struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ItemID     uint      `json:"item_id"`
	Item       Item      `json:"item"`
	SalePrice  float64   `json:"sale_price"` // bu satışın CƏMİ məbləği (bir ədədin qiyməti × say)
	Quantity   int       `gorm:"default:1" json:"quantity"` // satılan ədəd sayı
	Profit     float64   `json:"profit"`
	Channel        string    `json:"channel"`
	CustomerID     *uint     `json:"customer_id"`
	Customer       *Customer `json:"customer"`
	BranchID       uint      `json:"branch_id"`
	WarrantyMonths int       `json:"warranty_months"`
	SoldAt         time.Time `json:"sold_at"`
	// satılanlar siyahısına/hesabata sayılırmı. Kredit satışı əlimizə pul gəlmədiyi
	// üçün ödəniş tam olana VƏ YA kredit əl ilə bağlanana qədər false olur.
	// (default tag YOXDUR — əks halda GORM false-u atıb default true yazır;
	// bütün satış yaradan yerlər Counted-i açıq şəkildə təyin edir.)
	Counted bool `json:"counted"`
}

// CreditPlan — taksit / nisyə borcu
type CreditPlan struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CustomerID uint      `json:"customer_id"`
	Customer   Customer  `json:"customer"`
	ItemName   string          `json:"item_name"`
	Total      float64         `json:"total"`
	Paid       float64         `json:"paid"`
	NextDue    time.Time       `json:"next_due"`
	Status     string          `json:"status"` // ontime | duetoday | overdue | paid
	Payments   []CreditPayment `json:"payments"`
	// Closed — "biz özümüz bağladıq" checkbox. Seçilsə, satış hesabata düşür
	// (borc yenə qalır, müştəridən ödəniş adi qaydada alınır).
	Closed bool `json:"closed"`
	// bağlı olduğu satış (varsa) — kredit bağlananda həmin satışı «sayılan» edirik
	SaleID *uint `json:"sale_id"`
}

// Consignment (realizasiya) — başqa mağazalara topdan verilən mallar
// status: out | sold_unpaid | sold_paid | returned
type Consignment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	StoreName  string    `json:"store_name"`
	ItemID     *uint     `json:"item_id"` // stokdakı cihaza bağlantı (qayıdanda geri qaytarmaq üçün)
	ItemName   string    `json:"item_name"`
	Serial     string    `json:"serial"`
	GivenAt    time.Time `json:"given_at"`
	Cost       float64   `json:"cost"`
	GivenPrice float64   `json:"given_price"`
	Status     string    `json:"status"`
	Debt       float64   `json:"debt"`
	// "sold_paid" olanda yaradılan satış (pul kassaya gəlir) — geri dönəndə silmək üçün
	SaleID *uint `json:"sale_id"`
}

// Expense — xərclər (kirayə, maaş, kommunal…)
type Expense struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Date     time.Time `json:"date"`
	Category string    `json:"category"`
	Note     string    `json:"note"`
	Amount   float64   `json:"amount"`
}

// OnlineOrder — saytdan/appdan gələn sifariş (ödəniş yoxdur, 24 saat rezerv)
// status: pending | converted | cancelled
type OnlineOrder struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CustomerName string    `json:"customer_name"`
	Phone        string    `json:"phone"`
	ItemID       uint      `json:"item_id"`
	Item         Item      `json:"item"`
	Note         string    `json:"note"`
	Status       string    `gorm:"default:pending" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// DailyClose — kassa gün bağlanışı
type DailyClose struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Date        time.Time `json:"date"`
	Cash        float64   `json:"cash"`
	Card        float64   `json:"card"`
	Installment float64   `json:"installment"`
	Credit      float64   `json:"credit"`
	Total       float64   `json:"total"`
}

// SupplyBatch — gələn mal partiyası (Dubay, İstanbul, yerli)
type SupplyBatch struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Date      time.Time `json:"date"`
	Source    string    `json:"source"`
	Supplier  string    `json:"supplier"`
	ItemCount int       `json:"item_count"`
	TotalCost float64   `json:"total_cost"`
	Status    string    `json:"status"` // received | pending
}
