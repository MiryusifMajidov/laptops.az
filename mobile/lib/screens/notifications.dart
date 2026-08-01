import 'package:flutter/material.dart';
import '../api.dart';
import '../i18n.dart';
import '../store.dart';
import '../theme.dart';
import '../widgets/common.dart';
import '../widgets/net_image.dart';
import 'nav.dart';

class NotificationsScreen extends StatefulWidget {
  const NotificationsScreen({super.key});
  @override
  State<NotificationsScreen> createState() => _NotificationsScreenState();
}

class _NotificationsScreenState extends State<NotificationsScreen> {
  late Future<List<NotifItem>> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<NotifItem>> _load() async {
    final items = await Api.notifications();
    if (items.isNotEmpty) {
      await Store.I.markNotifSeen(items.first.id); // ən yeni id → baxıldı (qırmızı nöqtə itir)
    }
    return items;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: C.appBg,
      appBar: AppBar(
        backgroundColor: C.appBg,
        elevation: 0,
        foregroundColor: C.ink,
        title: Text(t('notif.title'), style: mr(size: 17, weight: FontWeight.w800, color: C.ink)),
      ),
      body: FutureBuilder<List<NotifItem>>(
        future: _future,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator(color: C.ink));
          }
          if (snap.hasError) {
            return ErrorState(snap.error.toString(), onRetry: () => setState(() => _future = _load()));
          }
          final items = snap.data ?? [];
          if (items.isEmpty) return EmptyState(t('notif.empty'));
          return ListView.separated(
            padding: const EdgeInsets.all(14),
            itemCount: items.length,
            separatorBuilder: (_, __) => const SizedBox(height: 10),
            itemBuilder: (context, i) {
              final n = items[i];
              return GestureDetector(
                onTap: () => openProduct(context, id: n.id),
                child: Container(
                  padding: const EdgeInsets.all(11),
                  decoration: BoxDecoration(
                    color: C.card,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: C.line),
                  ),
                  child: Row(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(10),
                        child: SizedBox(width: 54, height: 54, child: NetImage(n.cardImage)),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
                              decoration: BoxDecoration(color: C.ink, borderRadius: BorderRadius.circular(6)),
                              child: Text(t('notif.newTag'), style: mr(size: 9, weight: FontWeight.w800, color: Colors.white, spacing: .3)),
                            ),
                            const SizedBox(height: 5),
                            Text(n.name, maxLines: 2, overflow: TextOverflow.ellipsis, style: mr(size: 13.5, weight: FontWeight.w600, color: C.ink)),
                            const SizedBox(height: 3),
                            Row(
                              children: [
                                Text(money(n.finalPrice), style: sg(size: 14, weight: FontWeight.w700)),
                                if (n.hasDiscount) ...[
                                  const SizedBox(width: 6),
                                  Text(money(n.price),
                                      style: mr(size: 11, weight: FontWeight.w500, color: C.muted)
                                          .copyWith(decoration: TextDecoration.lineThrough)),
                                ],
                              ],
                            ),
                          ],
                        ),
                      ),
                      Icon(Icons.chevron_right_rounded, color: C.muted3),
                    ],
                  ),
                ),
              );
            },
          );
        },
      ),
    );
  }
}
