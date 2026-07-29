# Laptops.az — Mobile app (Flutter)

Customer-facing mobile app (iOS + Android) for Laptops.az. It consumes the same
**public API** as the website (`/api/public/*` on the Go backend, port 8080):
browse products, filter by category/brand/attributes/price, search, and place a
no-payment reservation order (the shop calls the customer back).

## Layout

```
mobile/
├─ OXU-MENI.txt        Azerbaijani, non-technical setup guide (read this first)
├─ SETUP.bat           One-click: flutter create + copy source + patch manifest + pub get
├─ patch_manifest.ps1  Adds INTERNET + cleartextTraffic + tel: query + app label
├─ README.md           (this file)
└─ app_src/            The actual app source (kept separate so `flutter create` can't clobber it)
   ├─ pubspec.yaml
   ├─ analysis_options.yaml
   ├─ assets/          logo images
   └─ lib/
      ├─ main.dart
      ├─ config.dart   ← SERVER URL + store phone (edit these)
      ├─ theme.dart    brand colors, Manrope/Space Grotesk, money()
      ├─ api.dart      models + API client
      ├─ store.dart    recent searches + local "my orders" (shared_preferences)
      ├─ widgets/      net_image, common, product_card, product_row
      └─ screens/      root (bottom nav), home, categories, category, product,
                       search, orders, order_form, filter_sheet, nav
```

## Why `app_src/`?

`flutter create .` regenerates and **overwrites** `lib/main.dart`, `pubspec.yaml`,
etc. with the demo template. So the real source lives in `app_src/`, and `SETUP.bat`
runs `flutter create` first (to generate the `android/`/`ios/` platform folders for
the installed Flutter version) and *then* copies `app_src/` on top. This avoids
version-mismatch problems from hand-writing Gradle/manifest files.

## Screens (match the design handoff — 5 mobile screens)

- **Home (APP01)** — logo + notification bell, dynamic "new arrival" push card,
  tappable search, featured hero, category chips, popular grid.
- **Categories** — grid of all categories → tap opens a category.
- **Category list (APP02 + 02b)** — brand chips + bottom-sheet **Filtr** panel
  (brand, price range, dynamic attribute chips), product rows.
- **Product (APP03)** — image carousel, brand/name/price, 2×2 spec grid,
  "Sifariş et" + call button. No online payment.
- **Search (APP04)** — debounced query (matches name *and* attribute values,
  same as the website), recent searches.
- **Order (APP05)** — name/phone/note → `POST /api/public/orders`, saves the
  `LA-000123` ref locally and shows it under the **Sifariş** tab.

## Setup / build

See `OXU-MENI.txt`. In short, once Flutter + Android Studio are installed:

```
SETUP.bat
# edit app_src/lib/config.dart -> kDefaultHost = 'http://<your-PC-LAN-IP>:8080'
flutter run          # live on a USB-connected phone
flutter build apk    # -> build/app/outputs/flutter-apk/app-release.apk
```

Phone and PC must be on the same Wi-Fi, and the backend (`start.bat`) must be
running. The server address can also be changed at runtime by long-pressing the
logo on the Home screen.

## Notes / possible later work

- **iOS build** requires a Mac (Xcode). The code is already iOS-ready.
- **Real push notifications** would need Firebase Cloud Messaging (the home "push
  card" is currently an in-app promo banner showing the newest product).
- Custom launcher icon (currently the default Flutter icon) can be added with
  `flutter_launcher_icons`.
- Fonts (Manrope / Space Grotesk) load via `google_fonts` (fetched + cached on
  first run); bundling the TTFs would make first launch fully offline.
