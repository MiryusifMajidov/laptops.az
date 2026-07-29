package main

// Language — saytın dəstəklədiyi dil (admin paneldən idarə olunur)
type Language struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Code      string `gorm:"uniqueIndex" json:"code"` // az, ru, tr, en
	Name      string `json:"name"`                    // ana dildə adı: Azərbaycan, Русский, Türkçe, English
	Icon      string `json:"icon"`                    // bayraq şəkli URL (/uploads/..) — boşdursa kod göstərilir
	Enabled   bool   `json:"enabled"`                 // saytda görünürmü
	IsDefault bool   `json:"is_default"`              // əsas dil (tərcümə yoxdursa buna düşür)
	Sort      int    `json:"sort"`                    // sıralama
}

// UiString — saytın statik mətni. Hər açar hər dil üçün bir sətir.
type UiString struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Key   string `gorm:"uniqueIndex:idx_uistr_key_lang" json:"key"`  // məs. home.heroTitle
	Lang  string `gorm:"uniqueIndex:idx_uistr_key_lang" json:"lang"` // dil kodu
	Value string `json:"value"`
}

// ItemTranslation — məhsulun adının başqa dildə variantı (əsas dil Item.Name-dədir).
type ItemTranslation struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	ItemID uint   `gorm:"uniqueIndex:idx_itemtr_item_lang;index" json:"item_id"`
	Lang   string `gorm:"uniqueIndex:idx_itemtr_item_lang" json:"lang"`
	Name   string `json:"name"`
}

// Setting — sayt üzrə açar-dəyər tənzimləmələri (dilə bağlı deyil): xəritə koordinatı, seçilmiş məhsul və s.
type Setting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `json:"value"`
}

// TermTranslation — kataloq terminlərinin (kateqoriya adı, xüsusiyyət adı, seçim dəyəri) tərcüməsi.
// Əsas (AZ) dəyər öz cədvəlindədir; burada yalnız digər dillərin variantı saxlanılır.
type TermTranslation struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Kind  string `gorm:"uniqueIndex:idx_term_krl" json:"kind"` // category | attribute | option
	RefID uint   `gorm:"uniqueIndex:idx_term_krl" json:"ref_id"`
	Lang  string `gorm:"uniqueIndex:idx_term_krl" json:"lang"`
	Value string `json:"value"`
}
