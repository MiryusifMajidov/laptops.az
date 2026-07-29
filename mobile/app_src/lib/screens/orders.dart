import 'package:flutter/material.dart';
import '../store.dart';
import '../theme.dart';
import '../widgets/common.dart';

/// «Sifariş» tabı — telefonda saxlanmış sifariş qeydləri.
class OrdersScreen extends StatelessWidget {
  const OrdersScreen({super.key});

  String _dateAz(String iso) {
    final d = DateTime.tryParse(iso);
    if (d == null) return '';
    const months = [
      'yan', 'fev', 'mar', 'apr', 'may', 'iyn',
      'iyl', 'avq', 'sen', 'okt', 'noy', 'dek'
    ];
    final two = (int n) => n.toString().padLeft(2, '0');
    return '${d.day} ${months[d.month - 1]}, ${two(d.hour)}:${two(d.minute)}';
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      bottom: false,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(18, 8, 18, 12),
            child: Text('Sifarişlərim', style: sg(size: 26, weight: FontWeight.w600, spacing: -.4)),
          ),
          Expanded(
            child: AnimatedBuilder(
              animation: Store.I,
              builder: (context, _) {
                final orders = Store.I.orders;
                if (orders.isEmpty) {
                  return const EmptyState(
                      'Hələ sifarişin yoxdur.\nMəhsul seçib «Sifariş et» düyməsinə bas.');
                }
                return ListView.separated(
                  padding: const EdgeInsets.fromLTRB(18, 4, 18, 20),
                  itemCount: orders.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 12),
                  itemBuilder: (context, i) {
                    final o = orders[i];
                    return Container(
                      padding: const EdgeInsets.all(14),
                      decoration: BoxDecoration(
                        color: C.card,
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: C.line2),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(o.ref, style: sg(size: 14, weight: FontWeight.w600, color: C.muted2)),
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                                decoration: BoxDecoration(
                                  color: const Color(0xFFFFF4E5),
                                  borderRadius: BorderRadius.circular(20),
                                ),
                                child: Text('Gözləyir',
                                    style: mr(size: 11, weight: FontWeight.w700, color: C.warn)),
                              ),
                            ],
                          ),
                          const SizedBox(height: 10),
                          Text(o.productName,
                              style: mr(size: 15, weight: FontWeight.w700, height: 1.3)),
                          const SizedBox(height: 8),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(_dateAz(o.date),
                                  style: mr(size: 12, weight: FontWeight.w500, color: C.muted)),
                              Text(money(o.price), style: sg(size: 16, weight: FontWeight.w600)),
                            ],
                          ),
                        ],
                      ),
                    );
                  },
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}
