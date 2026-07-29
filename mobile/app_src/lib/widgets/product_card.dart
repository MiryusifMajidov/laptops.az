import 'package:flutter/material.dart';
import '../api.dart';
import '../theme.dart';
import 'net_image.dart';
import 'common.dart';

/// «Populyar» üçün şaquli kart (2 sütunlu grid).
class ProductCard extends StatelessWidget {
  final Product p;
  final VoidCallback onTap;
  const ProductCard(this.p, {required this.onTap, super.key});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        decoration: BoxDecoration(
          color: C.card,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: C.line2),
        ),
        clipBehavior: Clip.antiAlias,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            AspectRatio(aspectRatio: 1.55, child: NetImage(p.cardImage)),
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 10, 12, 12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    p.name,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: mr(size: 12.5, weight: FontWeight.w700, color: C.ink, height: 1.3),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Flexible(
                        child: Text(money(p.price),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: sg(size: 15, weight: FontWeight.w600)),
                      ),
                      const SizedBox(width: 6),
                      const Dot(),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
