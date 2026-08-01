import 'package:flutter/material.dart';
import '../api.dart';
import '../i18n.dart';
import '../theme.dart';
import 'net_image.dart';
import 'common.dart';

/// «Populyar» üçün şaquli kart (2 sütunlu grid).
/// [onTap] — məhsul detalını açır; [onOrder] — birbaşa sifariş formasını açır.
class ProductCard extends StatelessWidget {
  final Product p;
  final VoidCallback onTap;
  final VoidCallback onOrder;
  const ProductCard(this.p, {required this.onTap, required this.onOrder, super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: C.card,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: C.line2),
      ),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // şəkil + mətn — məhsul detalını açır
          GestureDetector(
            behavior: HitTestBehavior.opaque,
            onTap: onTap,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                SizedBox(height: 118, child: NetImage(p.cardImage)),
                Padding(
                  padding: const EdgeInsets.fromLTRB(11, 9, 11, 0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      SizedBox(
                        height: 34, // 2 sətir üçün sabit yer — kartlar bərabər olsun
                        child: Text(
                          p.name,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: mr(size: 12.5, weight: FontWeight.w700, color: C.ink, height: 1.3),
                        ),
                      ),
                      const SizedBox(height: 6),
                      if (p.hasDiscount)
                        Text(money(p.price),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: mr(size: 10.5, weight: FontWeight.w500, color: C.muted)
                                .copyWith(decoration: TextDecoration.lineThrough)),
                      Row(
                        children: [
                          Flexible(
                            child: Text(money(p.finalPrice),
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
          const Spacer(),
          // «Sifariş et» — birbaşa sifariş formasını açır
          Padding(
            padding: const EdgeInsets.fromLTRB(11, 9, 11, 11),
            child: GestureDetector(
              behavior: HitTestBehavior.opaque,
              onTap: onOrder,
              child: Container(
                alignment: Alignment.center,
                padding: const EdgeInsets.symmetric(vertical: 9),
                decoration: BoxDecoration(
                  color: C.ink,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(t('common.order'),
                    style: mr(size: 12.5, weight: FontWeight.w700, color: Colors.white)),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
