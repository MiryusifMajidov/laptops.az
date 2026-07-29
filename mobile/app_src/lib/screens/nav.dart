import 'package:flutter/material.dart';
import '../api.dart';
import '../config.dart';
import '../theme.dart';
import 'product.dart';
import 'category.dart';

/// Məhsul detalına keç.
void openProduct(BuildContext context, {int? id, Product? product}) {
  Navigator.of(context).push(MaterialPageRoute(
    builder: (_) => ProductScreen(id: id, initial: product),
  ));
}

/// Kateqoriya siyahısına keç.
void openCategory(BuildContext context, String name) {
  Navigator.of(context).push(MaterialPageRoute(
    builder: (_) => CategoryScreen(category: name),
  ));
}

/// Server ünvanını dəyişmək üçün dialoq (loqoya uzun basıldıqda).
Future<void> showServerDialog(BuildContext context) async {
  final ctrl = TextEditingController(text: AppConfig.host);
  final result = await showDialog<bool>(
    context: context,
    builder: (ctx) => AlertDialog(
      backgroundColor: C.appBg,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(18)),
      title: Text('Server ünvanı', style: sg(size: 19, weight: FontWeight.w600)),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Telefon eyni Wi-Fi-dəki kompüterə qoşulur.\nNümunə:  http://192.168.1.35:8080',
            style: mr(size: 12.5, weight: FontWeight.w500, color: C.muted2, height: 1.4),
          ),
          const SizedBox(height: 14),
          TextField(
            controller: ctrl,
            autofocus: true,
            keyboardType: TextInputType.url,
            style: mr(size: 14.5, weight: FontWeight.w600),
            decoration: InputDecoration(
              hintText: 'http://192.168.1.35:8080',
              hintStyle: mr(size: 14, color: C.muted3),
              filled: true,
              fillColor: C.card,
              contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 13),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: const BorderSide(color: C.line3),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: const BorderSide(color: C.ink, width: 1.5),
              ),
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(ctx, false),
          child: Text('Ləğv et', style: mr(size: 14, weight: FontWeight.w700, color: C.muted2)),
        ),
        TextButton(
          onPressed: () => Navigator.pop(ctx, true),
          child: Text('Yadda saxla', style: mr(size: 14, weight: FontWeight.w700, color: C.ink)),
        ),
      ],
    ),
  );
  if (result == true) {
    await AppConfig.setHost(ctrl.text);
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Server ünvanı yadda saxlanıldı: ${AppConfig.host}')),
      );
    }
  }
}
