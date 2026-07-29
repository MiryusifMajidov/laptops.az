# Laptops.az — Rəqəmsal Sistem

Mağazanın Excel iş axınını əvəz edən sistem. Ardıcıllıq: **Admin panel → Website → Mobil app (Flutter)**.

## Texnologiya
- **Backend:** Go (`net/http` + GORM). Lokal geliştirmə üçün **SQLite** (CGO-suz, `glebarez/sqlite`); prod-da PostgreSQL.
- **Frontend (admin):** React + Vite + TypeScript.
- **Mobil (sonra):** Flutter (iOS + Android), eyni Go API-ni işlədəcək.

## Struktur
```
backend/      Go API (SQLite + seed data)
frontend/     React admin panel
prototype/    HTML dizayn referansı (admin.html)
```

## Əsas prinsip — bir mal = bir qeyd
Excel-də mal iki dəfə yazılırdı (stok + satış). Burada hər cihaz **bir** `Item` qeydidir və statusu dəyişir:
`in_stock → reserved → sold → returned`. Satış bağlananda həmin cihaz avtomatik `sold` olur, mənfəət (`sale_price − cost`) özü hesablanır. Stok və Satış eyni bazadandır.

## Data modeli (qısa)
- **Item** — bir fiziki cihaz (seriya, kateqoriya, filial, alış qiyməti, status, saytda göstər).
- **Category / Attribute / AttributeOption** — çevik xüsusiyyət sistemi (Notebook üçün RAM/SSD/Ekran kartı…). Məhsul əlavə edəndə yalnız seçilən kateqoriyanın xüsusiyyətləri çıxır.
- **Branch** — filiallar (Mərkəz, Elçin & Rəşid, Zaur; artırıla bilər).
- **Sale** — satış (kanal: Nağd/Kart/Taksit/Kredit, müştəri, mənfəət).
- **CreditPlan** — taksit/nisyə borcları.
- **Consignment** (realizasiya) — başqa mağazalara topdan verilən mallar.
- **Customer**, **Expense**.

## API endpoint-ləri
`GET /api/dashboard` · `GET,POST /api/items` · `GET /api/categories` · `GET /api/attributes` · `GET /api/branches` · `GET,POST /api/sales` · `GET /api/customers` · `GET /api/credits` · `GET /api/consignments` · `GET /api/expenses` · `GET /api/orders` · `POST /api/orders/{id}/convert|cancel` · `GET /api/kassa` · `POST /api/kassa/close` · `GET /api/supplies` · `GET /api/reports`

## Admin ekranları (hamısı canlı API-yə bağlı)
Ana Panel · Satışlar · Stok/Anbar (+Yeni məhsul) · Yeni Satış (drawer) · Onlayn Sifarişlər · Kredit/Borclar · Kassa · Xərclər · Müştərilər · Filiallar · Realizasiya · Təchizat · Kateqoriyalar · Hesabatlar.

## Giriş (login)
- **admin / admin** — admin (Səid, mağaza sahibi) · audit jurnalını görür
- **user / user** — satıcı · audit jurnalını görmür (yeganə fərq budur)

Parollar bcrypt ilə hash-lanır, token ilə qorunur. Panel daxilindən parol dəyişmək olar (Tənzimləmələr).

## Konfiqurasiya (env — istəyə görə)
- `DATABASE_URL` — qoyulsa PostgreSQL işlədilir (prod); yoxdursa lokal SQLite.
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM` — mail göndərmək üçün. Qoyulmayınca «Mail & Excel»-də göndərmə mexanizmi hazırdır, amma real göndərmir.

## İşə salmaq
Əvvəlcə Go (1.22+) və Node (20+) quraşdırılmalıdır.

Backend:
```
cd backend
go mod tidy
go run .
# API: http://localhost:8080/api
```
Frontend:
```
cd frontend
npm install
npm run dev
# http://localhost:5173
```
