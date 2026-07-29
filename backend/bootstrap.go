package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// bootstrap — prod-da (DATA_DIR təyin olunubsa) datanı kalıcı volume-dan işlədir.
// İlk boot-da baza yoxdursa, image-ə yığılmış seed-i (real baza + şəkillər) volume-a köçürür,
// sonrakı deploy-larda volume-dakı (yenilənmiş) datanı saxlayır.
// Lokal dev-də DATA_DIR boşdur → heç nə etmir, cari qovluqda işləyir.
func bootstrap() {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		return
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("data dir yarat: %v", err)
	}
	dbPath := filepath.Join(dataDir, "laptops.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if seed := os.Getenv("SEED_DIR"); seed != "" {
			if err := copyFile(filepath.Join(seed, "laptops.db"), dbPath); err == nil {
				copyTree(filepath.Join(seed, "uploads"), filepath.Join(dataDir, "uploads"))
				log.Println("İlk başlanğıc: seed baza + şəkillər volume-a köçürüldü")
			} else {
				log.Printf("seed köçürmə xətası (boş bazadan başlanacaq): %v", err)
			}
		}
	}
	if err := os.Chdir(dataDir); err != nil {
		log.Fatalf("chdir %s: %v", dataDir, err)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// copyTree — src qovluğundakı faylları dst-ə köçürür (uploads bir səviyyəlidir).
func copyTree(src, dst string) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		copyFile(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name()))
	}
}
