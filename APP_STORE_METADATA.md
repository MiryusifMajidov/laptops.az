# App Store Connect — Laptops.az listinq mətnləri

> Bunları App Store Connect → App Information / Localization sahələrinə kopyala.
> Simvol limitləri Apple-ın tələbidir.

## Əsas sahələr
| Sahə | Dəyər |
|---|---|
| **App Name** (≤30) | `Laptops.az` |
| **Bundle ID** | `az.laptops.laptopsAz` |
| **Primary Category** | Shopping |
| **Secondary Category** | (istəyə bağlı) Business |
| **Age Rating** | 4+ |
| **Support URL** | https://laptops.az |
| **Marketing URL** | https://laptops.az |
| **Privacy Policy URL** | https://laptops.az/privacy (domen işləməsə: https://laptops-az.fly.dev/privacy) |
| **Copyright** | © 2026 Laptops.az |

---

## 🇦🇿 Azərbaycan (əsas dil)

**Subtitle** (≤30): `Notbuk və kompüter mağazası`

**Promotional Text** (≤170):
```
Notbuk, kompüter və aksesuarları rahat seçin. AI köməkçi ilə sizə uyğun modeli tapın, ad və nömrə ilə sifariş edin — ödəniş yoxdur, mağaza zəng edir.
```

**Description** (≤4000):
```
Laptops.az — notbuk, kompüter və aksesuar almağın ən rahat yolu.

• AI KÖMƏKÇİ — nə üçün istifadə edəcəyinizi yazın (oyun, iş, dizayn, təhsil), sizə uyğun modeli tövsiyə etsin.
• SADƏ SİFARİŞ — kart/ödəniş yoxdur. Ad və telefon nömrəsi kifayətdir, məhsul 24 saat saxlanılır və mağaza sizə zəng edir.
• AXTARIŞ VƏ FİLTR — marka, qiymət və xüsusiyyətlərə görə tez tapın.
• YENİ MƏHSUL BİLDİRİŞLƏRİ — yeni gələnlərdən ilk siz xəbər tutun.
• 4 DİL — Azərbaycan, Rus, Türk, İngilis.
• TOPDAN PARTNYORLAR — mağazanızsa, partnyor girişi ilə optavoy qiymətləri görün.

Ödəniş tətbiq daxilində alınmır. Bütün alışlar mağazada (nağd/kart/taksit) tamamlanır.

laptops.az
```

**Keywords** (≤100, vergüllə):
```
notbuk,laptop,kompüter,noutbuk,macbook,asus,hp,lenovo,dell,aksesuar,mağaza,texnika,gaming
```

**What's New** (ilk versiya):
```
İlk versiya: məhsul kataloqu, AI köməkçi, sadə sifariş, çoxdilli interfeys.
```

---

## 🇬🇧 English (recommended — review is in English)

**Subtitle** (≤30): `Laptops & computers store`

**Promotional Text** (≤170):
```
Browse laptops, computers and accessories. Let the AI assistant find your match, then order with just your name and phone — no payment, the store calls you.
```

**Description** (≤4000):
```
Laptops.az — the easiest way to buy laptops, computers and accessories.

• AI ASSISTANT — tell it what you need (gaming, work, design, study) and get the right model recommended.
• SIMPLE ORDERING — no card, no payment. Just your name and phone; the item is held for 24h and the store calls you.
• SEARCH & FILTER — find fast by brand, price and specs.
• NEW-PRODUCT ALERTS — be first to know about new arrivals.
• 4 LANGUAGES — Azerbaijani, Russian, Turkish, English.
• WHOLESALE PARTNERS — store owners can sign in to see wholesale prices.

No payments are taken inside the app. All purchases are completed in-store (cash/card/installment).

laptops.az
```

**Keywords** (≤100):
```
laptop,notebook,computer,macbook,asus,hp,lenovo,dell,accessories,store,tech,gaming,shop
```

**What's New**:
```
First release: product catalog, AI assistant, simple ordering, multi-language UI.
```

> RU/TR lokalizasiyaları istəsən, eyni strukturu tərcümə et — tətbiq onsuz da 4 dillidir.

---

## App Review — Reviewer üçün qeydlər (App Review Information → Notes)
```
- No account is required to browse or place an order. Ordering only needs a name and phone number; no payment is taken in the app (purchases are completed in-store).
- "Partner" login is optional and only for wholesale store partners. Demo partner account (if the reviewer wants to test wholesale prices):
    username: appledemo
    password: <DOLDUR — App Store Connect-də saxlanan parol>
- Account deletion (Guideline 5.1.1(v)): Profile tab -> sign in -> red "Hesabı sil" (Delete account) button
  directly below "Çıxış" (Log out) -> confirm. This performs a full, permanent server-side deletion
  (user record, all sessions, and the partnership application submitted under that name). It is not a
  deactivation, and it requires no call or email to support.
- The AI assistant is powered by Google Gemini via our backend.
- Contact: info@laptops.az
```
> ⚠️ Reviewer partnyor rejimini yoxlaya bilsin deməli test hesabı ver (yoxsa "sign-in required" rədd riski). Admin paneldən bir demo partner hesabı yarat və yuxarıya yaz.

---

## Apple 5.1.1(v) rəddinə CAVAB — Resolution Center-ə olduğu kimi köçür (ingiliscə)

```
Hello,

Thank you for the review. Build 4 of version 1.0.0 adds in-app account deletion, as required by
Guideline 5.1.1(v).

Where to find it:
1. Open the app and select the "Profil" (Profile) tab — the last tab in the bottom bar.
2. Sign in with the demo partner account listed in App Review Information.
3. On the profile screen, directly below "Çıxış" (Log out), tap the red button
   "Hesabı sil" (Delete account).
4. A confirmation dialog appears — tap "Bəli, sil" (Yes, delete).

This immediately and permanently deletes the account together with the personal data associated
with it on our server: the user record, every active session, and the partnership application that
was submitted under that name. It is a full deletion, not a deactivation or a temporary disable,
and it does not require the user to call or email customer support. After deletion the app returns
to the signed-out state.

Please note that because the deletion is real and permanent, the demo account will no longer exist
after you have tested the flow. If you need to sign in again at any point, please let us know in
this thread and we will recreate it immediately.

One further change in this build: the product detail screen now lists all specifications of a
device (previously only the first four were shown).

Thank you for your time.

Laptops.az
```

**Bu cavabla birlikdə nə lazımdır:**
- Apple «screen recording captured on a physical device» istəyib. Ekran yazısı **iPhone-da** çəkilməlidir
  (TestFlight-dan build 4 quraşdırılır → Profil → giriş → «Hesabı sil» → «Bəli, sil» → çıxış ekranı).
  Yazını Resolution Center-də cavaba **attachment** kimi əlavə et.
- Yoxlayıcı silmə axınını test edəndən sonra `appledemo` hesabı **həqiqətən silinir** →
  növbəti submission-dan əvvəl admin panel → Tərəfdaşlıq → partner hesabı yenidən yarat
  (ad: `appledemo`, rol: partner) və parolu App Store Connect-dəki ilə eyni qoy.

## Age Rating anketi — cavablar
Demək olar hamısına **None/No**: zorakılıq yox, cinsi məzmun yox, qumar yox, məhdud maddələr yox → nəticə **4+**.

## Data Privacy (App Privacy) anketi — App Store Connect
| Data | Toplanır? | Nə üçün | Şəxsə bağlıdır? |
|---|---|---|---|
| Ad | Bəli | Sifariş (App Functionality) | Bəli |
| Telefon | Bəli | Sifariş (App Functionality) | Bəli |
| İstifadəçi məzmunu (AI mesajları) | Bəli | App Functionality | Xeyr (istəsən) |
| Device ID (push token) | Bəli | App Functionality (bildiriş) | Xeyr |
> Tracking (reklam izləmə) YOXDUR → "Data used to track you: No".
