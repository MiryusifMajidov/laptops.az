package main

import (
	"encoding/json"
	"log"
	"os"

	"gorm.io/gorm"
)

// BİRDƏFƏLİK (2026-09-25, 3-cü dalğa): sahibkarın qaydası —
// «Excel-də realizasiyadan çıxarılıbsa və satışa atılıbsa, deməkki artıq pulu ödənilib».
// Excel-in REALIZACIYA vərəqindən çıxıb «продажа» vərəqinə keçən və DB-də satış qeydi olan
// realizasiya sətirləri «satılıb·ödənilib» kimi bağlanır: borc 0, mövcud satışa bağlanır.
//
// ⚠️ YENİ SATIŞ YARADILMIR — mövcud satışın id-si yazılır (bir fiziki satış = bir pul qeydi).
// ⚠️ Yalnız hər üç şərt ödənəndə: Excel REALIZACIYA-da YOX + продажа-da VAR + DB-də satış VAR.

type consPlan struct {
	CloseConsignmentPaid []struct {
		ConsignmentID uint    `json:"consignment_id"`
		SaleID        uint    `json:"sale_id"`
		ItemID        uint    `json:"item_id"`
		GivenPrice    float64 `json:"given_price"`
		Serial        string  `json:"serial"`
	} `json:"close_consignment_paid"`
}

func applySync20260925c() {
	const marker = "sync_2026_09_25c"
	const path = "/data/sync_2026_09_25c.json"
	if getSetting(marker+"_done") == "1" {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Printf("%s: plan faylı oxunmadı (%s): %v", marker, path, err)
		return
	}
	var p consPlan
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Printf("%s: plan oxunmadı: %v", marker, err)
		return
	}
	var n int
	for _, m := range p.CloseConsignmentPaid {
		var cs Consignment
		if db.First(&cs, m.ConsignmentID).Error != nil {
			log.Printf("%s: KEÇİLDİ realizasiya %d (tapılmadı)", marker, m.ConsignmentID)
			continue
		}
		if cs.Status == "sold_paid" || cs.SaleID != nil {
			log.Printf("%s: KEÇİLDİ realizasiya %d (artıq bağlıdır)", marker, m.ConsignmentID)
			continue
		}
		// satış həmin cihaza aid olmalıdır — səhv sətri bağlamamaq üçün
		var s Sale
		if db.First(&s, m.SaleID).Error != nil || cs.ItemID == nil || s.ItemID != *cs.ItemID ||
			*cs.ItemID != m.ItemID {
			log.Printf("%s: KEÇİLDİ realizasiya %d (satış #%d uyğun gəlmir)", marker, m.ConsignmentID, m.SaleID)
			continue
		}
		sid := s.ID
		oldDebt := cs.Debt // Updates(map) cs struktunu da dəyişir — log üçün əvvəlcədən saxlanılır
		upd := map[string]any{"status": "sold_paid", "debt": 0, "sale_id": sid}
		if m.GivenPrice > 0 {
			upd["given_price"] = m.GivenPrice
		}
		if m.Serial != "" {
			upd["serial"] = m.Serial // consignments.serial öz kopyasıdır — sinxron saxlanılır
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if e := tx.Model(&cs).Updates(upd).Error; e != nil {
				return e
			}
			// cihazın statusu satışla uyğun olsun
			return tx.Model(&Item{}).Where("id = ?", *cs.ItemID).Update("status", "sold").Error
		})
		if err != nil {
			log.Printf("%s: XƏTA realizasiya %d: %v", marker, m.ConsignmentID, err)
			continue
		}
		log.Printf("%s: realizasiya #%d bağlandı — borc %.0f₼ → 0, satış #%d (%s)",
			marker, cs.ID, oldDebt, sid, cs.StoreName)
		n++
	}
	log.Printf("%s: baglanan_realizasiya=%d", marker, n)
	setSetting(marker+"_done", "1")
	os.Remove(path)
}
