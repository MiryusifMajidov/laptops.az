package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Müştəri AI köməkçisi — Gemini proxy + function calling (real sifariş yaradır).
// Açar YALNIZ server tərəfdədir (GEMINI_API_KEY env / fly secret) — frontend-ə heç vaxt düşmür.

type aiMsg struct {
	Role string `json:"role"` // "user" | "assistant"
	Text string `json:"text"`
}

var productMarker = regexp.MustCompile(`\[\[product:(\d+)\]\]`)

// stokda olan malların qısa kataloqu (60 san keşlənir) — AI-nin "bazanı bilməsi" üçün
var (
	catMu   sync.Mutex
	catStr  string
	catTime time.Time
)

func cachedCatalog() string {
	catMu.Lock()
	defer catMu.Unlock()
	if catStr != "" && time.Since(catTime) < 60*time.Second {
		return catStr
	}
	var items []Item
	db.Preload("Category").Preload("Values.Attribute").
		Where("show_on_site = ? AND status = ?", true, "in_stock").
		Order("price asc").Find(&items)
	var b strings.Builder
	for _, it := range items {
		var specs []string
		for _, v := range it.Values {
			switch v.Attribute.Name {
			case "Marka", "Prosessor", "RAM", "SSD", "Ekran kartı", "Ekran", "Yaddaş", "Vəziyyət":
				if v.Value != "" {
					specs = append(specs, v.Value)
				}
			}
		}
		price := it.Price
		if it.Discount > 0 && it.Discount < it.Price {
			price = it.Price - it.Discount // endirimli (real) qiymət
		}
		fmt.Fprintf(&b, "#%d | %s | %s | %s | ₼%.0f\n", it.ID, it.Name, it.Category.Name, strings.Join(specs, ", "), price)
	}
	catStr = b.String()
	catTime = time.Now()
	return catStr
}

const aiSystemPrompt = `Sən Laptops.az-ın müştəri köməkçisisən — Bakıda 2009-cu ildən fəaliyyət göstərən notebook, telefon və aksesuar mağazası.

ROLUN: Müştəriyə səmimi, isti və peşəkar şəkildə uyğun məhsul seçməkdə kömək et. Robot kimi yox — real, mehriban satış məsləhətçisi kimi danış.

DİL: Müştəri hansı dildə yazırsa, HƏMİŞƏ məhz həmin dildə cavab ver (Azərbaycan, rus, türk, ingilis və s.). Dili müştərinin SON mesajından təyin et; müştəri dili dəyişsə, sən də dəyiş. Məhsul adlarını və markaları olduğu kimi saxla.

QAYDALAR:
- YALNIZ aşağıdakı STOK KATALOQU-ndakı məhsulları tövsiyə et. Katalogda olmayan məhsulu UYDURMA.
- Müştərinin ehtiyacını anla (nə üçün: oyun, ofis, tələbə, dizayn; büdcə; marka) və uyğun 1-3 məhsul təklif et.
- Bir məhsul tövsiyə edəndə onu AYRICA sətirdə [[product:ID]] formatında yaz (ID kataloqdakı # nömrəsidir). Sayt onu gözəl məhsul kartı kimi göstərəcək.
- Qiymətləri ₼ ilə de. Ödəniş onlayn deyil, mağazada olur (ətraflı: aşağıdakı ÖDƏNİŞ bölməsi).
- Cavabları qısa və aydın saxla. 1-2 emoji olar, çox yox.
- Markdown, başlıq (###), cədvəl və ya çoxlu ulduz (**) İŞLƏTMƏ — sadə, isti danışıq mətni yaz.
- Büdcə və ya təyinat bilinmirsə, nəzakətlə soruş.
- Yalnız mağaza və məhsullarla bağlı danış.

ÖDƏNİŞ (dəqiq bil və düzgün de, uydurma):
- Ödəniş mağazada olur. Üsullar: Nağd və ya Kart. Taksit YALNIZ taksit kartı ilə mümkündür (məs. Birbank, Bolkart, Tamkart, Albalı).
- Mağazamızda DAXİLİ KREDİT / NİSYƏ YOXDUR. Kimsə "kredit", "nisyə" və ya "borc" soruşsa, olmadığını nəzakətlə de — yalnız taksit kartı keçərlidir.
- Taksit kartı ilə qiymətin ÜZƏRİNƏ faiz gəlir: 12 ay → +16%, 18 ay → +22%. Yalnız bu iki müddət var.
- Müştəri taksiti soruşsa konkret hesabla və aydın de: 12 ay üçün ümumi = qiymət × 1.16, aylıq = ümumi ÷ 12; 18 ay üçün ümumi = qiymət × 1.22, aylıq = ümumi ÷ 18. Rəqəmləri ₼ ilə göstər. Nümunə: 1000 ₼ məhsul → 12 ay: cəmi 1160 ₼, aylıq ~97 ₼; 18 ay: cəmi 1220 ₼, aylıq ~68 ₼.

SİFARİŞ / REZERV:
- Müştəri konkret məhsulu almaq/rezerv etmək istəyəndə, əvvəlcə ad-soyadını və telefon nömrəsini soruş.
- Hər ikisini aldıqdan SONRA "create_order" alətini çağır — həqiqətən çağır, uydurma.
- Alət nəticəsini ALMAYINCA "rezerv olundu" DEMƏ. Uğurlu olsa, verilən referans nömrəsini (məs. LA-000042) müştəriyə de və məhsulun 24 saat saxlanacağını bildir.
- Alət xəta qaytarsa, müştərini mağaza ilə telefonla əlaqəyə yönəlt.

STOK KATALOQU (#ID | ad | kateqoriya | xüsusiyyətlər | qiymət):
`

// AI-nin çağıra biləcəyi alət: real onlayn sifariş/rezerv yaradır (admin paneldə görünür)
var orderTool = map[string]any{
	"functionDeclarations": []any{
		map[string]any{
			"name":        "create_order",
			"description": "Müştəri konkret məhsulu rezerv/sifariş etmək istədikdə çağır. YALNIZ ad-soyad, telefon və məhsul təsdiqləndikdən sonra. Mağazada 24 saat rezerv yaradır.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"customer_name": map[string]any{"type": "string", "description": "Müştərinin ad və soyadı"},
					"phone":         map[string]any{"type": "string", "description": "Əlaqə telefon nömrəsi"},
					"product_id":    map[string]any{"type": "integer", "description": "Kataloqdakı məhsulun # nömrəsi (ID)"},
				},
				"required": []string{"customer_name", "phone", "product_id"},
			},
		},
	},
}

// create_order alətini icra et — həqiqi OnlineOrder yaradır (publicOrder ilə eyni məntiq)
func executeCreateOrder(args map[string]any) map[string]any {
	name, _ := args["customer_name"].(string)
	phone, _ := args["phone"].(string)
	var pid uint
	switch v := args["product_id"].(type) {
	case float64:
		pid = uint(v)
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			pid = uint(n)
		}
	}
	if strings.TrimSpace(name) == "" || pid == 0 {
		return map[string]any{"status": "error", "message": "ad və məhsul mütləqdir"}
	}
	var it Item
	if err := db.Where("show_on_site = ? AND status = ?", true, "in_stock").First(&it, pid).Error; err != nil {
		return map[string]any{"status": "error", "message": "məhsul stokda tapılmadı"}
	}
	o := OnlineOrder{
		CustomerName: name, Phone: phone, ItemID: pid, Note: "AI köməkçi vasitəsilə",
		Status: "pending", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	db.Create(&o)
	return map[string]any{"status": "ok", "ref": fmt.Sprintf("LA-%06d", o.ID), "product": it.Name, "reserved_hours": 24}
}

// bir Gemini çağırışı — contents göndərir; cavab content-i (raw), parts-ı, HTTP statusu və xətanı qaytarır.
// status retry qərarı üçün lazımdır (503/429 → təkrar cəhd).
func geminiCall(apiKey, model string, contents []any, system string) (json.RawMessage, []geminiPart, int, error) {
	reqBody := map[string]any{
		"system_instruction": map[string]any{"parts": []map[string]string{{"text": system}}},
		"contents":           contents,
		"tools":              []any{orderTool},
		// Qeyd: thinkingConfig göndərilmir — cari flash modelləri bəzən onu rədd edir
		// (400 INVALID_ARGUMENT). maxOutputTokens düşünmə + cavabı əhatə etsin deyə yüksəkdir.
		"generationConfig": map[string]any{"temperature": 0.6, "maxOutputTokens": 3072},
	}
	bb, _ := json.Marshal(reqBody)
	url := "https://generativelanguage.googleapis.com/v1beta/models/" + model + ":generateContent"
	httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(bb))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", apiKey)
	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, nil, 0, err // şəbəkə/timeout — transient sayılır (status 0)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		snippet := string(raw)
		if len(snippet) > 1500 {
			snippet = snippet[:1500]
		}
		log.Printf("gemini error %d (%s): %s", resp.StatusCode, model, snippet)
		return nil, nil, resp.StatusCode, fmt.Errorf("gemini status %d", resp.StatusCode)
	}
	var gr struct {
		Candidates []struct {
			Content json.RawMessage `json:"content"`
		} `json:"candidates"`
	}
	json.Unmarshal(raw, &gr)
	if len(gr.Candidates) == 0 {
		return nil, nil, 200, fmt.Errorf("boş cavab")
	}
	var parsed struct {
		Parts []geminiPart `json:"parts"`
	}
	json.Unmarshal(gr.Candidates[0].Content, &parsed)
	return gr.Candidates[0].Content, parsed.Parts, 200, nil
}

// geminiGenerate — geminiCall üstündə etibarlılıq qatı:
// transient xətalarda (503 overloaded / 429 quota / 500 / şəbəkə) qısa gözləmə ilə təkrar cəhd,
// davam edərsə ehtiyat modelə keçir. Beləcə "AI cavab vermir" kəskin azalır.
func geminiGenerate(apiKey string, models []string, contents []any, system string) (json.RawMessage, []geminiPart, error) {
	var lastErr error
	for _, model := range models {
		for attempt := 0; attempt < 3; attempt++ {
			raw, parts, status, err := geminiCall(apiKey, model, contents, system)
			if err == nil {
				return raw, parts, nil
			}
			lastErr = err
			// 503 overloaded / 500 / şəbəkə → qısa gözləyib təkrar cəhd et.
			// 429 (kvota) və 400/404 → təkrar ETMƏ (kvotaya retry onu daha da yeyir); birbaşa ehtiyat modelə.
			transient := status == 500 || status == 503 || status == 0
			if !transient {
				break
			}
			if attempt < 2 {
				time.Sleep(time.Duration(500*(attempt+1)) * time.Millisecond) // 0.5s, 1s
			}
		}
	}
	return nil, nil, lastErr
}

type geminiPart struct {
	Text         string `json:"text"`
	FunctionCall *struct {
		Name string         `json:"name"`
		Args map[string]any `json:"args"`
	} `json:"functionCall"`
}

// aiReply — mesaj tarixçəsindən AI cavabı hazırlayır. Sayt/app/WhatsApp/Instagram —
// HAMISI eyni beyindən keçir (Gemini + stok kataloqu + create_order alət döngüsü).
func aiReply(msgs []aiMsg) (string, error) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return "", fmt.Errorf("GEMINI_API_KEY yoxdur")
	}
	// əsas model (güclü, ən yeni GA flash); GEMINI_MODEL secret-i varsa onu götürür.
	primary := os.Getenv("GEMINI_MODEL")
	if primary == "" {
		primary = "gemini-3.5-flash"
	}
	// ehtiyat model — əsas model yüklənib/xəta versə avtomatik keçilir.
	models := []string{primary}
	if primary != "gemini-flash-latest" {
		models = append(models, "gemini-flash-latest")
	}
	if len(msgs) > 24 {
		msgs = msgs[len(msgs)-24:]
	}
	for i := range msgs {
		if len(msgs[i].Text) > 1500 {
			msgs[i].Text = msgs[i].Text[:1500]
		}
	}
	contents := make([]any, 0, len(msgs)+2)
	for _, m := range msgs {
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]any{"role": role, "parts": []any{map[string]any{"text": m.Text}}})
	}
	system := aiSystemPrompt + cachedCatalog()
	var reply string
	for iter := 0; iter < 3; iter++ {
		rawContent, parts, err := geminiGenerate(key, models, contents, system)
		if err != nil {
			return "", err
		}
		var call *struct {
			Name string
			Args map[string]any
		}
		for _, p := range parts {
			if p.FunctionCall != nil {
				call = &struct {
					Name string
					Args map[string]any
				}{p.FunctionCall.Name, p.FunctionCall.Args}
			}
			reply += p.Text
		}
		if call != nil && call.Name == "create_order" {
			result := executeCreateOrder(call.Args)
			var rawModel map[string]any
			json.Unmarshal(rawContent, &rawModel)
			contents = append(contents, rawModel)
			contents = append(contents, map[string]any{
				"role":  "user",
				"parts": []any{map[string]any{"functionResponse": map[string]any{"name": "create_order", "response": result}}},
			})
			reply = "" // final mətn növbəti dövrdən gələcək
			continue
		}
		break
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		reply = "Bağışlayın, cavab hazırlaya bilmədim. Zəhmət olmasa yenidən soruşun 🙏"
	}
	return reply, nil
}

func aiChat(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	// bot / sui-istifadəyə qarşı: IP üzrə + günlük tavan
	if ok, info := aiLimiter.allow(clientIP(r)); !ok {
		writeJSON(w, 429, map[string]string{"error": rateLimitMsg(lang, info)})
		return
	}
	if os.Getenv("GEMINI_API_KEY") == "" {
		writeJSON(w, 503, map[string]string{"error": "AI köməkçi hələ aktiv deyil"})
		return
	}
	var in struct {
		Messages       []aiMsg `json:"messages"`
		ConversationID string  `json:"conversation_id"` // söhbəti qruplaşdırmaq üçün (client yaradır)
		Source         string  `json:"source"`          // web | app
	}
	if err := decodeBody(r, &in); err != nil || len(in.Messages) == 0 {
		writeJSON(w, 400, map[string]string{"error": "mesaj yoxdur"})
		return
	}
	reply, err := aiReply(in.Messages)
	if err != nil {
		writeJSON(w, 502, map[string]string{"error": "AI cavab vermədi, bir azdan yenidən yoxlayın"})
		return
	}

	// söhbəti admin panel üçün saxla (tam yazışma + yeni cavab)
	logAiConversation(in.ConversationID, in.Source, lang, clientIP(r), in.Messages, reply)

	// tövsiyə olunan məhsulların məlumatlarını qaytar ([[product:ID]] markerləri)
	products := map[string]any{}
	for _, m := range productMarker.FindAllStringSubmatch(reply, -1) {
		id := m[1]
		if _, ok := products[id]; ok {
			continue
		}
		var it Item
		if err := db.Preload("Category").Preload("Values.Attribute").
			Where("show_on_site = ? AND status = ?", true, "in_stock").First(&it, id).Error; err == nil {
			applyItemLang(&it, "", "") // daxili qiymətləri (alış/topdan) gizlə
			products[id] = it
		}
	}
	writeJSON(w, 200, map[string]any{"reply": reply, "products": products})
}
