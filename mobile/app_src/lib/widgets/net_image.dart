import 'package:flutter/material.dart';
import '../config.dart';
import '../theme.dart';

/// Şəbəkədən şəkil — boş / yüklənir / xəta halları düzgün göstərilir.
class NetImage extends StatelessWidget {
  final String path; // API-dən gələn yol (card_image və s.)
  final BoxFit fit;
  const NetImage(this.path, {this.fit = BoxFit.cover, super.key});

  @override
  Widget build(BuildContext context) {
    final url = AppConfig.img(path);
    if (url.isEmpty) return const _Placeholder();
    return Image.network(
      url,
      fit: fit,
      gaplessPlayback: true,
      loadingBuilder: (ctx, child, progress) {
        if (progress == null) return child;
        return const _Placeholder(loading: true);
      },
      errorBuilder: (ctx, err, st) => const _Placeholder(),
    );
  }
}

class _Placeholder extends StatelessWidget {
  final bool loading;
  const _Placeholder({this.loading = false});

  @override
  Widget build(BuildContext context) {
    return Container(
      color: C.surface,
      alignment: Alignment.center,
      child: loading
          ? const SizedBox(
              width: 18,
              height: 18,
              child: CircularProgressIndicator(strokeWidth: 2, color: C.muted3),
            )
          : Icon(Icons.laptop_mac_outlined, color: const Color(0xFFB9B6AD), size: 26),
    );
  }
}
