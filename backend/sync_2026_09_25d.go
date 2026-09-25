package main

import (
	"encoding/json"
	"log"
	"os"

	"gorm.io/gorm"
)

// BİRDƏFƏLİK (2026-09-25, 4-cü dalğa): sahibkarın göstərişi «dublikatları həll et».
// 2-ci dalğa dublikatları SERİYA üzrə tapırdı, ona görə iki növ dublikatı görməmişdi:
//   a) SERİYASIZ kölgə kartlar — 2026-09-17-də əl ilə yaradılıb, satışları səhvən 2016 ili ilə
//      yazılıb (ona görə heç bir 2026 hesabatına düşmür, amma ümumi dövriyyəyə düşür).
//      Hər birinin seriyalı, düzgün tarixli EYNİ satışı var (Excel «продажа» ilə təsdiqlənib).
//   b) ANAQRAM seriya — «NXJSBER…» ↔ «nxbjser…» (B/J yerdəyişməsi): normallaşdırma
//      yalnız kiçik hərfə salıb durğu işarəsi atdığı üçün hərf yerdəyişməsini tutmur.
//
// ⚠️ Satış qeydi silinir (sales-də soft-delete yoxdur), ona görə silinməzdən ƏVVƏL tam sətir
//    jurnala yazılır — səhv olsa Fly log-undan bərpa oluna bilər.
// ⚠️ Kölgə kart yalnız SOFT-DELETE olunur (arxiv) — məhsul fiziki silinmir.
// ⚠️ Seriyalı əsl kartlara və onların satışlarına TOXUNULMUR.

type dedupPlan struct {
	DropDuplicateSale []struct {
		SaleID        uint   `json:"sale_id"`
		ArchiveItemID uint   `json:"archive_item_id"` // kölgə kart (0 = arxivləmə yoxdur)
		KeepSaleID    uint   `json:"keep_sale_id"`    // əsl satış — yoxlama üçün, TOXUNULMUR
		Why           string `json:"why"`
	} `json:"drop_duplicate_sale"`
}

func applySync20260925d() {
	const marker = "sync_2026_09_25d"
	const path = "/data/sync_2026_09_25d.json"
	if getSetting(marker+"_done") == "1" {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Printf("%s: plan faylı oxunmadı (%s): %v", marker, path, err)
		return
	}
	var p dedupPlan
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Printf("%s: plan oxunmadı: %v", marker, err)
		return
	}
	var n int
	var dropped float64
	for _, m := range p.DropDuplicateSale {
		var s Sale
		if db.First(&s, m.SaleID).Error != nil {
			log.Printf("%s: KEÇİLDİ satış %d (tapılmadı — yəqin artıq silinib)", marker, m.SaleID)
			continue
		}
		// ƏSL satış hələ də yerindədirmi? Yoxsa heç nə silmirik (pul tamam itməsin).
		var keep Sale
		if m.KeepSaleID != 0 && db.First(&keep, m.KeepSaleID).Error != nil {
			log.Printf("%s: KEÇİLDİ satış %d — əsl satış #%d tapılmadı", marker, m.SaleID, m.KeepSaleID)
			continue
		}
		if m.KeepSaleID != 0 && keep.ID == s.ID {
			log.Printf("%s: KEÇİLDİ satış %d — əsl satış ilə eynidir", marker, m.SaleID)
			continue
		}
		// kölgə kart doğrudan da bu satışın malıdırmı?
		if m.ArchiveItemID != 0 && s.ItemID != m.ArchiveItemID {
			log.Printf("%s: KEÇİLDİ satış %d — malı #%d gözlənilirdi, #%d-dir",
				marker, m.SaleID, m.ArchiveItemID, s.ItemID)
			continue
		}
		log.Printf("%s: SİLİNİR dublikat satış#%d mal#%d %.2f₼ %s mənfəət=%.2f filial=%d kanal=%s (əsl: satış#%d) — %s",
			marker, s.ID, s.ItemID, s.SalePrice, s.SoldAt.Format("2006-01-02"),
			s.Profit, s.BranchID, s.Channel, m.KeepSaleID, m.Why)
		err := db.Transaction(func(tx *gorm.DB) error {
			if m.ArchiveItemID != 0 { // ƏVVƏL arxiv, SONRA silmə
				var left int64
				tx.Model(&Sale{}).Where("item_id = ? AND id <> ?", m.ArchiveItemID, s.ID).Count(&left)
				if left == 0 {
					if e := tx.Delete(&Item{}, m.ArchiveItemID).Error; e != nil {
						return e
					}
				} else {
					log.Printf("%s: kart #%d arxivlənmədi — başqa satışı var", marker, m.ArchiveItemID)
				}
			}
			return tx.Delete(&Sale{}, s.ID).Error
		})
		if err != nil {
			log.Printf("%s: XƏTA satış %d: %v", marker, m.SaleID, err)
			continue
		}
		dropped += s.SalePrice
		n++
	}
	log.Printf("%s: silinen_dublikat_satis=%d cemi=%.0f₼", marker, n, dropped)
	setSetting(marker+"_done", "1")
	os.Remove(path)
}
