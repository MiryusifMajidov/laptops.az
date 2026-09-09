package main

import "time"

// User — admin | user (satıcı) | partner (topdan mağaza)
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex" json:"username"`
	PassHash string `json:"-"`
	Role     string `json:"role"` // admin | user | partner
	Name     string `json:"name"`
	BranchID uint   `json:"branch_id"` // aid olduğu filial (satıcı yalnız öz filialını görür; admin hamısını)
}

// PartnerApplication — mobil tətbiqdən tərəfdaşlıq müraciəti (qeydiyyat əvəzi)
type PartnerApplication struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`       // ad soyad
	StoreName string    `json:"store_name"` // mağaza adı
	Phone     string    `json:"phone"`
	Status    string    `json:"status"` // pending | contacted | done
	CreatedAt time.Time `json:"created_at"`
}

// AuditLog — kim nə etdi (yalnız admin görür)
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Action    string    `json:"action"` // POST | PUT | DELETE
	Path      string    `json:"path"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

// EmailTemplate — mail şablonu (ad + vergüllə ünvanlar)
type EmailTemplate struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `json:"name"`
	Emails string `json:"emails"` // comma-separated
}

// CreditPayment — kredit üzrə ayrıca ödəniş qeydi (tarixçə)
type CreditPayment struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreditPlanID uint      `json:"credit_plan_id"`
	Amount       float64   `json:"amount"`
	CreatedAt    time.Time `json:"created_at"`
}
