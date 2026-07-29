import 'package:shared_preferences/shared_preferences.dart';

/// ================== SERVER ÜNVANI ==================
/// Tətbiq canlı serverə (Fly.io) qoşulur — istənilən yerdən, internet ilə işləyir.
///
/// Lokal geliştirmə üçün başqa ünvan lazım olsa (məs. http://192.168.1.35:8080),
/// tətbiqdə ANA səhifədə loqonun üstünə uzun bas (long-press) → ünvanı yaz.
/// ===================================================
const String kDefaultHost = 'https://laptops-az.fly.dev';

/// Mağazanın zəng nömrəsi (məhsul səhifəsindəki telefon düyməsi).
const String kStorePhone = '+994708151283';

class AppConfig {
  static String host = kDefaultHost;

  static const _key = 'server_host';

  /// Tətbiq açılanda yadda saxlanmış ünvanı yüklə.
  static Future<void> load() async {
    final p = await SharedPreferences.getInstance();
    final saved = p.getString(_key);
    if (saved != null && saved.trim().isNotEmpty) host = saved.trim();
  }

  static Future<void> setHost(String value) async {
    var v = value.trim();
    if (v.isEmpty) return;
    if (!v.startsWith('http://') && !v.startsWith('https://')) {
      v = 'http://$v';
    }
    // sondakı «/» simvolunu təmizlə
    while (v.endsWith('/')) {
      v = v.substring(0, v.length - 1);
    }
    host = v;
    final p = await SharedPreferences.getInstance();
    await p.setString(_key, v);
  }

  static String get apiBase => '$host/api/public';

  /// Şəkil yolu tam URL-ə çevrilir.
  static String img(String path) {
    if (path.isEmpty) return '';
    if (path.startsWith('http')) return path;
    return '$host$path';
  }
}
