import 'dart:convert';
import 'package:http/http.dart' as http;
import 'config.dart';
import 'store.dart';
import 'i18n.dart';

/// ------- Modellər (backend public API-yə uyğun) -------

class AttrValue {
  final String attribute; // xüsusiyyət adı (məs. "Marka")
  final String value; // dəyəri (məs. "Apple")
  AttrValue(this.attribute, this.value);

  factory AttrValue.fromJson(Map<String, dynamic> j) {
    final a = j['attribute'];
    final name = (a is Map && a['name'] != null) ? a['name'].toString() : '';
    return AttrValue(name, (j['value'] ?? '').toString());
  }
}

class Product {
  final int id;
  final String name;
  final String serial;
  final num price;
  final num discount;
  final String categoryName;
  final String cardImage;
  final List<String> gallery;
  final List<AttrValue> values;

  Product({
    required this.id,
    required this.name,
    required this.serial,
    required this.price,
    this.discount = 0,
    required this.categoryName,
    required this.cardImage,
    required this.gallery,
    required this.values,
  });

  factory Product.fromJson(Map<String, dynamic> j) {
    // gallery — backend-də JSON array string kimi saxlanılır
    List<String> gal = [];
    final g = j['gallery'];
    if (g is String && g.trim().isNotEmpty) {
      try {
        final parsed = jsonDecode(g);
        if (parsed is List) gal = parsed.map((e) => e.toString()).toList();
      } catch (_) {}
    } else if (g is List) {
      gal = g.map((e) => e.toString()).toList();
    }

    final vals = <AttrValue>[];
    final vraw = j['values'];
    if (vraw is List) {
      for (final v in vraw) {
        if (v is Map<String, dynamic>) vals.add(AttrValue.fromJson(v));
      }
    }

    final cat = j['category'];
    final catName = (cat is Map && cat['name'] != null) ? cat['name'].toString() : '';

    return Product(
      id: (j['id'] ?? 0) as int,
      name: (j['name'] ?? '').toString(),
      serial: (j['serial'] ?? '').toString(),
      price: (j['price'] ?? 0) as num,
      discount: (j['discount'] ?? 0) as num,
      categoryName: catName,
      cardImage: (j['card_image'] ?? '').toString(),
      gallery: gal,
      values: vals,
    );
  }

  /// Endirim varmı (₼ ilə, price-dən kiçik)
  bool get hasDiscount => discount > 0 && discount < price;

  /// Göstəriləcək real qiymət — endirim varsa price - discount
  num get finalPrice => hasDiscount ? price - discount : price;

  /// Adı verilmiş xüsusiyyətin dəyəri (yoxdursa boş sətir).
  String attr(String name) {
    for (final v in values) {
      if (v.attribute == name) return v.value;
    }
    return '';
  }

  /// Marka — «Marka» xüsusiyyəti yoxdursa, addakı ilk sözdən götür.
  String get brand {
    final m = attr('Marka');
    if (m.isNotEmpty) return m;
    final parts = name.trim().split(RegExp(r'\s+'));
    return parts.isNotEmpty ? parts.first : '';
  }

  /// Kart üçün qısa spesifikasiya sətri (dəyərlər seçilmiş dilə çevrilir).
  String get specLine {
    const order = ['Prosessor', 'RAM', 'SSD', 'Ekran kartı', 'Yaddaş', 'Ölçü'];
    final parts = <String>[];
    for (final n in order) {
      final v = attr(n);
      if (v.isNotEmpty) parts.add(tt(v));
      if (parts.length == 3) break;
    }
    return parts.join(' · ');
  }

  /// Detал ekranı üçün şəkil siyahısı (kart şəkli + qalereya, təkrarsız).
  List<String> get images {
    final out = <String>[];
    if (cardImage.isNotEmpty) out.add(cardImage);
    for (final g in gallery) {
      if (g.isNotEmpty && !out.contains(g)) out.add(g);
    }
    return out;
  }
}

class CatCount {
  final String name;
  final int count;
  CatCount(this.name, this.count);
  factory CatCount.fromJson(Map<String, dynamic> j) =>
      CatCount((j['name'] ?? '').toString(), (j['count'] ?? 0) as int);
}

/// Bildiriş elementi (yeni məhsul) — /api/public/notifications
class NotifItem {
  final int id;
  final String name;
  final String cardImage;
  final num price;
  final num discount;
  final String createdAt;
  NotifItem({required this.id, required this.name, required this.cardImage, required this.price, this.discount = 0, required this.createdAt});
  factory NotifItem.fromJson(Map<String, dynamic> j) => NotifItem(
        id: (j['id'] ?? 0) as int,
        name: (j['name'] ?? '').toString(),
        cardImage: (j['card_image'] ?? '').toString(),
        price: (j['price'] ?? 0) as num,
        discount: (j['discount'] ?? 0) as num,
        createdAt: (j['created_at'] ?? '').toString(),
      );

  bool get hasDiscount => discount > 0 && discount < price;
  num get finalPrice => hasDiscount ? price - discount : price;
}

class OrderRef {
  final int id;
  final String ref; // məs. LA-000123
  OrderRef(this.id, this.ref);
}

/// AI köməkçi mesajı və cavabı
class AiMsg {
  final String role; // "user" | "assistant"
  final String text;
  AiMsg(this.role, this.text);
}

class AiResult {
  final String reply;
  final Map<String, Product> products; // [[product:ID]] markerlərinə uyğun
  AiResult(this.reply, this.products);
}

/// ------- API müştərisi -------

class ApiException implements Exception {
  final String message;
  ApiException(this.message);
  @override
  String toString() => message;
}

class Api {
  static const _timeout = Duration(seconds: 12);

  static Uri _u(String path, [Map<String, String>? q]) {
    final base = Uri.parse('${AppConfig.apiBase}$path');
    return q == null ? base : base.replace(queryParameters: {...base.queryParameters, ...q});
  }

  // Partner (topdan) girişi varsa optavoy qiymətli endpoint işlədilir.
  static Future<List<Product>> products({String? category, String? q}) async {
    final params = <String, String>{'lang': L.I.lang};
    if (category != null && category.isNotEmpty && category != 'all') {
      params['category'] = category;
    }
    if (q != null && q.trim().isNotEmpty) params['q'] = q.trim();
    final String res;
    if (Store.I.isPartner) {
      res = await _getAuth(_authU('/partner/products', params));
    } else {
      res = await _get(_u('/products', params));
    }
    final list = jsonDecode(res) as List;
    return list.map((e) => Product.fromJson(e as Map<String, dynamic>)).toList();
  }

  static Future<Product> product(int id) async {
    final p = {'lang': L.I.lang};
    final String res = Store.I.isPartner
        ? await _getAuth(_authU('/partner/products/$id', p))
        : await _get(_u('/products/$id', p));
    return Product.fromJson(jsonDecode(res) as Map<String, dynamic>);
  }

  // son yeni məhsullar (bildiriş lenti) — adlar seçilmiş dilə çevrilir
  static Future<List<NotifItem>> notifications() async {
    final res = await _get(_u('/notifications', {'lang': L.I.lang}));
    final list = jsonDecode(res) as List;
    return list.map((e) => NotifItem.fromJson(e as Map<String, dynamic>)).toList();
  }

  // Site Settings-də seçilmiş məhsulun id-si (0 = seçilməyib)
  static Future<int> featuredItemId() async {
    try {
      final res = await _get(_u('/settings'));
      final m = jsonDecode(res) as Map<String, dynamic>;
      return int.tryParse((m['featured_item_id'] ?? '').toString()) ?? 0;
    } catch (_) {
      return 0;
    }
  }

  // ---- partner giriş + tərəfdaşlıq müraciəti ----

  static Uri _authU(String path, [Map<String, String>? q]) {
    final base = Uri.parse('${AppConfig.host}/api$path');
    return q == null ? base : base.replace(queryParameters: {...base.queryParameters, ...q});
  }

  static Map<String, String> _authHeaders() {
    final t = Store.I.token;
    return {
      'Content-Type': 'application/json',
      if (t != null && t.isNotEmpty) 'Authorization': 'Bearer $t',
    };
  }

  static Future<String> _getAuth(Uri url) async {
    late http.Response r;
    try {
      r = await http.get(url, headers: _authHeaders()).timeout(_timeout);
    } catch (_) {
      throw ApiException('Serverə qoşulmaq alınmadı.');
    }
    if (r.statusCode == 401 || r.statusCode == 403) {
      await Store.I.logout();
      throw ApiException('Giriş vaxtı bitdi — yenidən daxil olun.');
    }
    if (r.statusCode >= 200 && r.statusCode < 300) return r.body;
    throw ApiException(_errOf(r));
  }

  static Future<void> login(String username, String password) async {
    late http.Response r;
    try {
      r = await http
          .post(_authU('/login'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'username': username, 'password': password}))
          .timeout(_timeout);
    } catch (_) {
      throw ApiException('Serverə qoşulmaq alınmadı.');
    }
    if (r.statusCode >= 200 && r.statusCode < 300) {
      final j = jsonDecode(utf8.decode(r.bodyBytes)) as Map<String, dynamic>;
      await Store.I.setAuth(
        token: (j['token'] ?? '').toString(),
        role: (j['role'] ?? '').toString(),
        name: (j['name'] ?? username).toString(),
      );
      return;
    }
    throw ApiException(_errOf(r));
  }

  /// Hesabı tamamilə silir (App Store 5.1.1(v)). Uğurlu olsa lokal sessiya da təmizlənir.
  static Future<void> deleteAccount() async {
    late http.Response r;
    try {
      r = await http.delete(_authU('/account'), headers: _authHeaders()).timeout(_timeout);
    } catch (_) {
      throw ApiException('Serverə qoşulmaq alınmadı.');
    }
    if (r.statusCode >= 200 && r.statusCode < 300) {
      await Store.I.logout();
      return;
    }
    throw ApiException(_errOf(r));
  }

  static Future<void> partnerApply({required String name, required String storeName, required String phone}) async {
    late http.Response r;
    try {
      r = await http
          .post(_u('/partner-apply'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'name': name, 'store_name': storeName, 'phone': phone}))
          .timeout(_timeout);
    } catch (_) {
      throw ApiException('Serverə qoşulmaq alınmadı.');
    }
    if (r.statusCode >= 200 && r.statusCode < 300) return;
    throw ApiException(_errOf(r));
  }

  static Future<List<CatCount>> categories() async {
    final res = await _get(_u('/categories'));
    final list = jsonDecode(res) as List;
    return list.map((e) => CatCount.fromJson(e as Map<String, dynamic>)).toList();
  }

  static Future<OrderRef> order({
    required String name,
    required String phone,
    required int itemId,
    String note = '',
  }) async {
    final body = jsonEncode({
      'customer_name': name,
      'phone': phone,
      'item_id': itemId,
      'note': note,
    });
    late http.Response r;
    try {
      r = await http
          .post(_u('/orders'),
              headers: {'Content-Type': 'application/json'}, body: body)
          .timeout(_timeout);
    } catch (_) {
      throw ApiException('Serverə qoşulmaq alınmadı. İnternet / server ünvanını yoxla.');
    }
    if (r.statusCode >= 200 && r.statusCode < 300) {
      final j = jsonDecode(r.body) as Map<String, dynamic>;
      return OrderRef((j['id'] ?? 0) as int, (j['ref'] ?? '').toString());
    }
    throw ApiException(_errOf(r));
  }

  // söhbəti admin-də qruplaşdırmaq üçün app-launch başına sabit id
  static final String _aiConvId =
      'app-${DateTime.now().microsecondsSinceEpoch.toRadixString(36)}';

  /// AI köməkçi — backend Gemini proksisi (/api/public/ai-chat).
  static Future<AiResult> aiChat(List<AiMsg> messages) async {
    final body = jsonEncode({
      'messages': messages.map((m) => {'role': m.role, 'text': m.text}).toList(),
      'conversation_id': _aiConvId,
      'source': 'app',
    });
    late http.Response r;
    try {
      r = await http
          .post(_u('/ai-chat'),
              headers: {'Content-Type': 'application/json'}, body: body)
          .timeout(const Duration(seconds: 40));
    } catch (_) {
      throw ApiException('AI-yə qoşulmaq alınmadı. İnterneti yoxla.');
    }
    if (r.statusCode >= 200 && r.statusCode < 300) {
      final j = jsonDecode(utf8.decode(r.bodyBytes)) as Map<String, dynamic>;
      final prods = <String, Product>{};
      final praw = j['products'];
      if (praw is Map) {
        praw.forEach((k, v) {
          if (v is Map<String, dynamic>) prods[k.toString()] = Product.fromJson(v);
        });
      }
      return AiResult((j['reply'] ?? '').toString(), prods);
    }
    throw ApiException(_errOf(r));
  }

  static Future<String> _get(Uri url) async {
    late http.Response r;
    try {
      r = await http.get(url).timeout(_timeout);
    } catch (_) {
      throw ApiException('Serverə qoşulmaq alınmadı. İnternet / server ünvanını yoxla.');
    }
    if (r.statusCode >= 200 && r.statusCode < 300) return r.body;
    throw ApiException(_errOf(r));
  }

  static String _errOf(http.Response r) {
    try {
      final j = jsonDecode(r.body);
      if (j is Map && j['error'] != null) return j['error'].toString();
    } catch (_) {}
    return 'Xəta baş verdi (${r.statusCode})';
  }
}
