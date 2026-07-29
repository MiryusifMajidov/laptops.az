import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../config.dart';
import '../theme.dart';

/// Şəbəkədən şəkil — keşlənir (cached_network_image), boş/yüklənir/xəta halları düzgün göstərilir.
/// [width] — kiçildilmiş versiya (backend /img). Kart/siyahı üçün kiçik saxla; detaildə 0 (tam).
class NetImage extends StatelessWidget {
  final String path; // API-dən gələn yol (card_image və s.)
  final BoxFit fit;
  final int width; // hədəf en (piksel); <=0 → tam ölçü
  const NetImage(this.path, {this.fit = BoxFit.cover, this.width = 500, super.key});

  @override
  Widget build(BuildContext context) {
    final url = AppConfig.thumb(path, w: width);
    if (url.isEmpty) return const _Placeholder();
    return CachedNetworkImage(
      imageUrl: url,
      fit: fit,
      fadeInDuration: const Duration(milliseconds: 150),
      placeholder: (ctx, _) => const _Placeholder(loading: true),
      errorWidget: (ctx, _, __) => const _Placeholder(),
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
