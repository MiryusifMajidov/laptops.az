import 'package:flutter/material.dart';
import '../api.dart';
import '../theme.dart';
import 'net_image.dart';
import 'common.dart';

/// Kateqoriya siyahısı üçün üfüqi sətir kartı.
class ProductRow extends StatelessWidget {
  final Product p;
  final VoidCallback onTap;
  const ProductRow(this.p, {required this.onTap, super.key});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: C.card,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: C.line2),
        ),
        child: Row(
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(11),
              child: SizedBox(width: 78, height: 64, child: NetImage(p.cardImage)),
            ),
            const SizedBox(width: 13),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(p.name,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: mr(size: 14.5, weight: FontWeight.w700, color: C.ink)),
                  if (p.specLine.isNotEmpty) ...[
                    const SizedBox(height: 2),
                    Text(p.specLine,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: mr(size: 12, weight: FontWeight.w500, color: C.muted)),
                  ],
                  const SizedBox(height: 9),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(money(p.price), style: sg(size: 17, weight: FontWeight.w600)),
                      const AvailableTag(),
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

/// Axtarış nəticəsi üçün kompakt sətir.
class ProductRowCompact extends StatelessWidget {
  final Product p;
  final VoidCallback onTap;
  const ProductRowCompact(this.p, {required this.onTap, super.key});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(11),
        decoration: BoxDecoration(
          color: C.card,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: C.line2),
        ),
        child: Row(
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(10),
              child: SizedBox(width: 56, height: 48, child: NetImage(p.cardImage)),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(p.name,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: mr(size: 14, weight: FontWeight.w700, color: C.ink)),
                  const SizedBox(height: 2),
                  Text(money(p.price), style: sg(size: 13, weight: FontWeight.w600, color: C.muted2)),
                ],
              ),
            ),
            const SizedBox(width: 8),
            const Dot(size: 8),
          ],
        ),
      ),
    );
  }
}
