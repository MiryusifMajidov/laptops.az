import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

/// Laptops.az brend rəngləri (dizayndan birbaşa götürülüb).
class C {
  static const canvas = Color(0xFFE7E6E1); // ümumi fon (dizayn kətanı)
  static const appBg = Color(0xFFFAFAF8); // ekran fonu
  static const ink = Color(0xFF14213A); // əsas mətn / tünd
  static const card = Color(0xFFFFFFFF); // ağ kart
  static const surface = Color(0xFFF4F3EF); // şəkil boşluğu / açıq səth

  static const line = Color(0xFFECEAE4); // sərhəd
  static const line2 = Color(0xFFEEEDE8);
  static const line3 = Color(0xFFE4E2DB);

  static const muted = Color(0xFF98A0AC); // solğun mətn
  static const muted2 = Color(0xFF5A6474);
  static const muted3 = Color(0xFFB0AEA8);

  static const green = Color(0xFF0E9F6E); // «mövcuddur» nöqtəsi
  static const greenTx = Color(0xFF0A7E57); // yaşıl mətn
  static const greenBg = Color(0xFFE9F7F1);
  static const greenBg2 = Color(0xFFF5F8F6);
  static const greenLine = Color(0xFFE4EFE9);

  static const danger = Color(0xFFDC2626);
  static const warn = Color(0xFFC2410C);
}

/// Space Grotesk — qiymətlər və başlıqlar üçün (dizayndakı .sg sinfi).
TextStyle sg({
  double size = 16,
  FontWeight weight = FontWeight.w600,
  Color color = C.ink,
  double? spacing,
  double? height,
}) {
  return GoogleFonts.spaceGrotesk(
    fontSize: size,
    fontWeight: weight,
    color: color,
    letterSpacing: spacing,
    height: height,
    fontFeatures: const [FontFeature.tabularFigures()],
  );
}

/// Manrope — əsas mətn şrifti.
TextStyle mr({
  double size = 14,
  FontWeight weight = FontWeight.w500,
  Color color = C.ink,
  double? spacing,
  double? height,
}) {
  return GoogleFonts.manrope(
    fontSize: size,
    fontWeight: weight,
    color: color,
    letterSpacing: spacing,
    height: height,
  );
}

ThemeData buildTheme() {
  final base = ThemeData(
    useMaterial3: true,
    scaffoldBackgroundColor: C.appBg,
    colorScheme: ColorScheme.fromSeed(
      seedColor: C.ink,
      primary: C.ink,
      surface: C.appBg,
      brightness: Brightness.light,
    ),
    splashFactory: InkRipple.splashFactory,
  );
  return base.copyWith(
    textTheme: GoogleFonts.manropeTextTheme(base.textTheme).apply(
      bodyColor: C.ink,
      displayColor: C.ink,
    ),
    appBarTheme: const AppBarTheme(
      backgroundColor: C.appBg,
      surfaceTintColor: Colors.transparent,
      elevation: 0,
      scrolledUnderElevation: 0,
    ),
  );
}

/// ₼1,899 formatı (min ayırıcı vergül).
String money(num n) {
  final v = n.round();
  final s = v.abs().toString();
  final buf = StringBuffer();
  for (var i = 0; i < s.length; i++) {
    if (i > 0 && (s.length - i) % 3 == 0) buf.write(',');
    buf.write(s[i]);
  }
  return '₼${v < 0 ? '-' : ''}$buf';
}
