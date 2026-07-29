import 'package:flutter/material.dart';
import '../api.dart';
import '../theme.dart';
import '../widgets/common.dart';
import 'nav.dart';

/// «Kateqoriya» tabı — bütün kateqoriyaların siyahısı.
class CategoriesScreen extends StatefulWidget {
  const CategoriesScreen({super.key});
  @override
  State<CategoriesScreen> createState() => _CategoriesScreenState();
}

class _CategoriesScreenState extends State<CategoriesScreen> {
  late Future<List<CatCount>> _future;

  @override
  void initState() {
    super.initState();
    _future = Api.categories();
  }

  IconData _iconFor(String name) {
    final n = name.toLowerCase();
    if (n.contains('telefon')) return Icons.smartphone_rounded;
    if (n.contains('monitor')) return Icons.desktop_windows_outlined;
    if (n.contains('aksesuar')) return Icons.headphones_rounded;
    if (n.contains('yığım') || n.contains('pc')) return Icons.dvr_rounded;
    if (n.contains('işlən')) return Icons.history_rounded;
    return Icons.laptop_mac_rounded;
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      bottom: false,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(18, 8, 18, 14),
            child: Text('Kateqoriyalar', style: sg(size: 26, weight: FontWeight.w600, spacing: -.4)),
          ),
          Expanded(
            child: FutureBuilder<List<CatCount>>(
              future: _future,
              builder: (context, snap) {
                if (snap.connectionState == ConnectionState.waiting) {
                  return const Center(child: CircularProgressIndicator(color: C.ink));
                }
                if (snap.hasError) {
                  return ErrorState(snap.error.toString(),
                      onRetry: () => setState(() => _future = Api.categories()));
                }
                final cats = snap.data ?? [];
                if (cats.isEmpty) return const EmptyState('Kateqoriya yoxdur');
                return GridView.builder(
                  padding: const EdgeInsets.fromLTRB(18, 0, 18, 20),
                  itemCount: cats.length,
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2,
                    crossAxisSpacing: 12,
                    mainAxisSpacing: 12,
                    childAspectRatio: 1.35,
                  ),
                  itemBuilder: (context, i) {
                    final c = cats[i];
                    return GestureDetector(
                      onTap: () => openCategory(context, c.name),
                      child: Container(
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          color: C.card,
                          borderRadius: BorderRadius.circular(18),
                          border: Border.all(color: C.line2),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Container(
                              width: 42,
                              height: 42,
                              decoration: BoxDecoration(
                                color: C.surface,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Icon(_iconFor(c.name), color: C.ink, size: 22),
                            ),
                            Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Text(c.name,
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    style: mr(size: 15, weight: FontWeight.w700)),
                                const SizedBox(height: 2),
                                Text('${c.count} məhsul',
                                    style: mr(size: 12, weight: FontWeight.w600, color: C.muted)),
                              ],
                            ),
                          ],
                        ),
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
