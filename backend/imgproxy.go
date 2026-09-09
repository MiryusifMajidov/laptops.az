package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Şəkil kiçiltmə/keşləmə proksisi.
//
//	GET /img?u=<mənbə>&w=<en>
//
// Kart/siyahı üçün kiçik, sıxılmış JPEG qaytarır (tam ölçü yalnız məhsul detailində).
// Mənbə: lokal "/uploads/.." VƏ YA xarici http(s) URL. Nəticə diskdə (imgcache/) keşlənir.
// Xarici kitabxana yoxdur — sırf stdlib (CGO_ENABLED=0 ilə uyğun).
// SSRF qoruması: HƏR bağlantıda (ilkin + hər redirect hop) real IP yoxlanır və
// məhz həmin IP-yə dial olunur (yenidən resolve yox → DNS rebinding bağlanır).
var imgHTTP = &http.Client{
	Timeout: 15 * time.Second,
	CheckRedirect: func(_ *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return fmt.Errorf("çox redirect")
		}
		return nil // hər hop onsuz da dial anında yoxlanır
	},
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil || len(ips) == 0 {
				return nil, fmt.Errorf("host resolve olunmadı")
			}
			for _, ip := range ips {
				if isBadIP(ip) {
					return nil, fmt.Errorf("qadağan ünvan")
				}
			}
			// yoxlanmış IP-yə birbaşa dial (yenidən DNS sorğusu olmur)
			d := net.Dialer{Timeout: 8 * time.Second}
			return d.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
		},
	},
}

const imgUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"

func imgResizeHandler(w http.ResponseWriter, r *http.Request) {
	src := strings.TrimSpace(r.URL.Query().Get("u"))
	if src == "" {
		http.Error(w, "u yoxdur", http.StatusBadRequest)
		return
	}
	width, _ := strconv.Atoi(r.URL.Query().Get("w"))
	if width <= 0 {
		width = 400
	}
	if width > 1600 {
		width = 1600
	}

	// keş açarı (mənbə + en)
	sum := sha1.Sum([]byte(fmt.Sprintf("%s|%d", src, width)))
	cachePath := filepath.Join("imgcache", hex.EncodeToString(sum[:])+".jpg")
	if b, err := os.ReadFile(cachePath); err == nil {
		writeImg(w, b)
		return
	}

	raw, err := fetchImageBytes(src)
	if err != nil {
		http.Error(w, "şəkil alınmadı", http.StatusBadGateway)
		return
	}
	// yalnız şəkil qəbul et — qeyri-şəkil (HTML/JSON) cavabı geri qaytarma (məlumat sızması)
	if !strings.HasPrefix(http.DetectContentType(raw), "image/") {
		http.Error(w, "şəkil deyil", http.StatusBadGateway)
		return
	}
	out, err := resizeToJPEG(raw, width)
	if err != nil {
		// dekod olunmadı (məs. webp) → orijinalı olduğu kimi ver (artıq image təsdiqlənib)
		writeImg(w, raw)
		return
	}
	_ = os.MkdirAll("imgcache", 0o755)
	_ = os.WriteFile(cachePath, out, 0o644)
	writeImg(w, out)
}

func writeImg(w http.ResponseWriter, b []byte) {
	ct := http.DetectContentType(b)
	if !strings.HasPrefix(ct, "image/") {
		ct = "image/jpeg"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	_, _ = w.Write(b)
}

func fetchImageBytes(src string) ([]byte, error) {
	// lokal yükləmə
	if strings.HasPrefix(src, "/uploads/") {
		if strings.Contains(src, "..") {
			return nil, fmt.Errorf("yanlış yol")
		}
		return os.ReadFile("." + src) // uploads cwd(DATA_DIR)-a nisbi
	}
	// xarici http(s) — SSRF qoruması
	u, err := url.Parse(src)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("yanlış url")
	}
	if isBlockedHost(u.Hostname()) {
		return nil, fmt.Errorf("qadağan host")
	}
	req, _ := http.NewRequest("GET", src, nil)
	req.Header.Set("User-Agent", imgUA)
	resp, err := imgHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 25<<20)) // maks 25MB
}

// isBadIP — daxili/şəbəkə ünvanları (SSRF hədəfləri: loopback, private, link-local…)
func isBadIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

// isBlockedHost — ilkin sürətli yoxlama (əsas müdafiə dial anındadır: hər hop yoxlanır).
func isBlockedHost(host string) bool {
	if host == "" || strings.EqualFold(host, "localhost") {
		return true
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return true
	}
	for _, ip := range ips {
		if isBadIP(ip) {
			return true
		}
	}
	return false
}

// resizeToJPEG — şəkli en=width-ə qədər kiçildir (nisbət qorunur), ağ fona flatten edir, JPEG q82 verir.
func resizeToJPEG(raw []byte, width int) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw <= 0 || sh <= 0 {
		return nil, fmt.Errorf("boş şəkil")
	}
	dw := width
	if sw < dw {
		dw = sw // upscale etmə
	}
	dh := sh * dw / sw
	if dh < 1 {
		dh = 1
	}
	dst := scaleOverWhite(src, dw, dh)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 82}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// scaleOverWhite — sahə-ortalama (box) ilə kiçiltmə + ağ fona kompozisiya (şəffaflıq üçün).
func scaleOverWhite(src image.Image, dw, dh int) *image.RGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for dy := 0; dy < dh; dy++ {
		sy0 := dy * sh / dh
		sy1 := (dy + 1) * sh / dh
		if sy1 <= sy0 {
			sy1 = sy0 + 1
		}
		for dx := 0; dx < dw; dx++ {
			sx0 := dx * sw / dw
			sx1 := (dx + 1) * sw / dw
			if sx1 <= sx0 {
				sx1 = sx0 + 1
			}
			var rs, gs, bs, cnt uint64
			for yy := sy0; yy < sy1; yy++ {
				for xx := sx0; xx < sx1; xx++ {
					cr, cg, cb, ca := src.At(b.Min.X+xx, b.Min.Y+yy).RGBA()
					inv := uint32(65535) - ca // premultiplied → ağ fona kompozisiya
					rs += uint64(cr + inv)
					gs += uint64(cg + inv)
					bs += uint64(cb + inv)
					cnt++
				}
			}
			if cnt == 0 {
				cnt = 1
			}
			dst.SetRGBA(dx, dy, color.RGBA{
				R: uint8((rs / cnt) >> 8),
				G: uint8((gs / cnt) >> 8),
				B: uint8((bs / cnt) >> 8),
				A: 255,
			})
		}
	}
	return dst
}
