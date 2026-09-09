# Laptops API — laptops.az

Laptop siyahısını çəkmək üçün sadə, açıq (açarsız) API.
Serverdə həmişə tətbiq olunan filtr: **Mərkəz filial + stokda + saytda aktiv + Notebook**.
Daxili qiymətlər (alış/topdan) verilmir. Şəkillər tam URL kimi gəlir. CORS açıqdır (brauzerdən birbaşa çağırmaq olar).

**Baza URL:** `https://laptops.az`

---

## 1. Siyahı
```
GET /api/public/laptops
```

Query parametrləri (hamısı könüllü):

| Param | Nə edir | Nümunə |
|-------|---------|--------|
| `q` | Ad və ya xüsusiyyət üzrə axtarış | `?q=dell` |
| `limit` | Neçə ədəd qaytarsın | `?limit=20` |
| `offset` | Neçəsini ötsün (səhifələmə) | `?limit=20&offset=20` |
| `lang` | Ad dili: `az`/`ru`/`tr`/`en` | `?lang=ru` |

Cavab: laptop obyektlərindən ibarət **massiv**.

## 2. Detal
```
GET /api/public/laptops/{id}
```
Tək laptop obyekti. Tapılmasa → **404**. `?lang=` burada da işləyir.

---

## Laptop obyekti

```json
{
  "id": 948,
  "name": "DELL ALIENWARE 16 AURORA",
  "price": 2399,
  "discount": 100,
  "final_price": 2299,
  "image": "https://laptops.az/uploads/178...png",
  "images": [
    "https://laptops.az/uploads/178...png",
    "https://laptops.az/uploads/178...png"
  ],
  "category": "Notebook",
  "specs": [
    { "name": "Marka", "value": "Dell" },
    { "name": "Prosessor", "value": "Intel Core 7 - 240H" },
    { "name": "RAM", "value": "16 GB" },
    { "name": "SSD", "value": "1 TB" },
    { "name": "Ekran kartı", "value": "Nvidia GeForce RTX 5050" }
  ],
  "created_at": "2026-08-15T..."
}
```

| Sahə | İzah |
|------|------|
| `id` | Laptopun id-si (detal URL-i üçün) |
| `name` | Ad |
| `price` | Sayt qiyməti (₼) |
| `discount` | Endirim (₼) |
| `final_price` | Ödəniləcək qiymət = `price − discount` |
| `image` | Əsas şəkil (tam URL) |
| `images` | Bütün şəkillər (tam URL massivi) |
| `category` | Həmişə `"Notebook"` |
| `specs` | Xüsusiyyətlər: `{name, value}` massivi |
| `created_at` | Əlavə olunma tarixi |

---

## Nümunə (JavaScript)

```javascript
const BASE = "https://laptops.az";

// Siyahı
const laptops = await fetch(`${BASE}/api/public/laptops`).then(r => r.json());
laptops.forEach(l => {
  console.log(l.name, l.final_price, l.image);
});

// Detal
const item = await fetch(`${BASE}/api/public/laptops/948`).then(r => r.json());
console.log(item.name, item.images, item.specs);
```

Şəkillər hazır tam URL olduğu üçün birbaşa `<img src={l.image}>` yazmaq kifayətdir.
