import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Yerli (telefonda saxlanan) sifariş qeydi — server bunları geri qaytarmır,
/// ona görə təsdiqləndikdən sonra burada saxlayırıq.
class MyOrder {
  final String ref;
  final int itemId;
  final String productName;
  final num price;
  final String date; // ISO
  MyOrder({
    required this.ref,
    required this.itemId,
    required this.productName,
    required this.price,
    required this.date,
  });

  Map<String, dynamic> toJson() => {
        'ref': ref,
        'item_id': itemId,
        'name': productName,
        'price': price,
        'date': date,
      };

  factory MyOrder.fromJson(Map<String, dynamic> j) => MyOrder(
        ref: (j['ref'] ?? '').toString(),
        itemId: (j['item_id'] ?? 0) as int,
        productName: (j['name'] ?? '').toString(),
        price: (j['price'] ?? 0) as num,
        date: (j['date'] ?? '').toString(),
      );
}

/// Sadə qlobal state — provider paketinə ehtiyac yoxdur.
/// Widget-lər `AnimatedBuilder(animation: store, ...)` ilə yenilənir.
class Store extends ChangeNotifier {
  static final Store I = Store._();
  Store._();

  static const _kSearches = 'recent_searches';
  static const _kOrders = 'my_orders';

  List<String> recentSearches = [];
  List<MyOrder> orders = [];

  Future<void> load() async {
    final p = await SharedPreferences.getInstance();
    recentSearches = p.getStringList(_kSearches) ?? [];
    final raw = p.getString(_kOrders);
    if (raw != null && raw.isNotEmpty) {
      try {
        final list = jsonDecode(raw) as List;
        orders = list.map((e) => MyOrder.fromJson(e as Map<String, dynamic>)).toList();
      } catch (_) {
        orders = [];
      }
    }
  }

  Future<void> addSearch(String term) async {
    final t = term.trim();
    if (t.isEmpty) return;
    recentSearches.removeWhere((s) => s.toLowerCase() == t.toLowerCase());
    recentSearches.insert(0, t);
    if (recentSearches.length > 8) {
      recentSearches = recentSearches.sublist(0, 8);
    }
    notifyListeners();
    final p = await SharedPreferences.getInstance();
    await p.setStringList(_kSearches, recentSearches);
  }

  Future<void> clearSearches() async {
    recentSearches = [];
    notifyListeners();
    final p = await SharedPreferences.getInstance();
    await p.remove(_kSearches);
  }

  Future<void> addOrder(MyOrder o) async {
    orders.insert(0, o);
    notifyListeners();
    final p = await SharedPreferences.getInstance();
    await p.setString(_kOrders, jsonEncode(orders.map((e) => e.toJson()).toList()));
  }
}
