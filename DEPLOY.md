# Laptops.az — Fly.io-ya deploy (addım-addım)

Bu layihə **tək Docker image**-dir: bir proqram həm backend-i (API), həm admin paneli, həm müştəri saytını verir.
Deploy edəndən sonra sənə lazım olan **tək şey** — bir dəfə əmrləri yazmaqdır. Sonra özü işləyir.

Deploydan sonra ünvanlar:
- **Müştəri saytı:** `https://laptops-az.fly.dev/`
- **Admin panel:** `https://laptops-az.fly.dev/admin`  (giriş: `admin` / `admin`)
- API: `https://laptops-az.fly.dev/api`

> `laptops-az` sənin seçdiyin app adıdır. Ad tutulubsa başqa ad seçəcəksən, ünvan da ona uyğun olacaq.

---

## 1. flyctl proqramını quraşdır (bir dəfəlik)

PowerShell aç və yaz:

```powershell
pwsh -Command "iwr https://fly.io/install.ps1 -useb | iex"
```

Quraşdıqdan sonra **yeni PowerShell pəncərəsi** aç (PATH yenilənsin) və yoxla:

```powershell
fly version
```

## 2. Fly.io hesabı aç (kart tələb olunur, amma bu ölçüdə pul çəkmir)

```powershell
fly auth signup
```
(Artıq hesabın varsa: `fly auth login`)

Brauzer açılacaq → qeydiyyat → kart əlavə et (təsdiq üçün, kiçik app pulsuz qalır).

## 3. Layihə qovluğuna keç və app yarat

```powershell
cd D:\laptops.az
fly launch --copy-config --no-deploy
```

Soruşduqda:
- **App adı:** `laptops-az` (tutulubsa başqa ad — məs. `laptops-az-baku`)
- **Region:** `fra`
- **Postgres / Redis / başqa baza:** **Xeyr** (lazım deyil)
- Əgər "volume yaradım?" soruşsa → **Bəli**

## 4. Kalıcı disk yarat (datanın itməməsi üçün ƏN VACİB addım)

Əgər 3-cü addımda volume avtomatik yaranmadısa, bunu yaz:

```powershell
fly volumes create laptops_data --region fra --size 1 -y
```

Bu disk sənin bütün datanı (məhsullar, satışlar, şəkillər) saxlayır. Deploy təkrarlansa belə **itmir**.

## 5. Deploy et 🚀

```powershell
fly deploy
```

İlk deploy 3–5 dəqiqə çəkə bilər (uzaqdan build olunur — sənin kompüterində Docker lazım deyil).
İlk açılışda mövcud real baza avtomatik diskə köçürülür.

## 6. Aç və yoxla

```powershell
fly open
```

Sayt açılacaq. Admin panel üçün: ünvanın sonuna `/admin` əlavə et → `admin` / `admin` ilə gir.

---

## ⚠️ İlk işdən sonra MÜTLƏQ et

Admin panel internetdə açıq olacaq, ona görə **standart parolu dərhal dəyiş**:
**Admin panel → Tənzimləmələr → parol dəyiş.** (`admin`/`admin` çox zəifdir.)

---

## Sonradan nə lazımdır?

- **Heç nə.** Deploydan sonra özü işləyir, sən heç nə etməməlisən.
- Uzun müddət heç kim açmasa, proqram "yuxuya" gedir və növbəti açılışda 1–2 saniyə gec qalxır (data itmir). Bu, qənaət üçündür.

### Faydalı əmrlər (istəsən)
```powershell
fly status      # işləyirmi?
fly logs        # canlı jurnal (xəta olsa görünər)
fly open        # saytı brauzerdə aç
```

### Kod dəyişəndə yenidən deploy
Mən layihədə dəyişiklik edəndə, sən sadəcə yenə bunu yazırsan:
```powershell
fly deploy
```
**Datan itmir** — kalıcı diskdə qalır, yalnız proqram yenilənir.

---

## Bir aydan sonra domain alanda
Heç nə yenidən qurulmayacaq. Fly-da bir əmrlə öz domenini bağlayırsan:
```powershell
fly certs add senin-domenin.az
```
və domen provayderində Fly-ın verdiyi DNS qeydini əlavə edirsən. Kod dəyişmir.
