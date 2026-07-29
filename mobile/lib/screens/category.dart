import 'package:flutter/material.dart';
import '../api.dart';
import '../i18n.dart';
import '../theme.dart';
import '../widgets/common.dart';
import '../widgets/product_row.dart';
import 'filter_sheet.dart';
import 'nav.dart';

/// APP02 — bir kateqoriyanın məhsulları + marka çipləri + filtr paneli.
class CategoryScreen extends StatefulWidget {
  final String category;
  const CategoryScreen({required this.category, super.key});

  @override
  State<CategoryScreen> createState() => _CategoryScreenState();
}

class _CategoryScreenState extends State<CategoryScreen> {
  late Future<List<Product>> _future;
  List<Product> _all = [];
  Facets _facets = Facets([], {}, 0, 0);
  Filters _filters = Filters();

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<Product>> _load() async {
    final items = await Api.products(category: widget.category);
    _all = items;
    _facets = Facets.from(items);
    return items;
  }

  List<Product> get _visible => _all.where(_filters.matches).toList();

  int _count(Filters f) => _all.where(f.matches).length;

  Future<void> _openFilter() async {
    final res = await showFilterSheet(
      context,
      facets: _facets,
      current: _filters,
      countFor: _count,
    );
    if (res != null) setState(() => _filters = res);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: C.appBg,
      body: SafeArea(
        bottom: false,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _header(),
            Expanded(
              child: FutureBuilder<List<Product>>(
                future: _future,
                builder: (context, snap) {
                  if (snap.connectionState == ConnectionState.waiting) {
                    return const Center(child: CircularProgressIndicator(color: C.ink));
                  }
                  if (snap.hasError) {
                    return ErrorState(snap.error.toString(),
                        onRetry: () => setState(() => _future = _load()));
                  }
                  if (_all.isEmpty) return EmptyState(t('cat.noProducts'));
                  return Column(
                    children: [
                      if (_facets.brands.isNotEmpty) _brandChips(),
                      Expanded(child: _list()),
                    ],
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _header() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(14, 6, 18, 12),
      child: Row(
        children: [
          IconButton(
            onPressed: () => Navigator.pop(context),
            icon: Icon(Icons.arrow_back_ios_new_rounded, size: 20, color: C.ink),
            splashRadius: 22,
          ),
          Expanded(
            child: Text(tt(widget.category),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: sg(size: 24, weight: FontWeight.w600, spacing: -.5)),
          ),
          GestureDetector(
            onTap: _openFilter,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 8),
              decoration: BoxDecoration(
                color: _filters.isEmpty ? C.card : C.ink,
                borderRadius: BorderRadius.circular(11),
                border: _filters.isEmpty ? Border.all(color: C.line) : null,
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.tune_rounded,
                      size: 15, color: _filters.isEmpty ? C.muted2 : Colors.white),
                  const SizedBox(width: 6),
                  Text(t('common.filter'),
                      style: mr(
                          size: 12.5,
                          weight: FontWeight.w600,
                          color: _filters.isEmpty ? C.muted2 : Colors.white)),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _brandChips() {
    final brands = _facets.brands;
    return SizedBox(
      height: 34,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.fromLTRB(18, 0, 18, 0),
        itemCount: brands.length + 1,
        separatorBuilder: (_, __) => const SizedBox(width: 8),
        itemBuilder: (context, i) {
          if (i == 0) {
            final active = _filters.brands.isEmpty;
            return _chip('${t('cat.all')} ${_all.length}', active, () {
              setState(() => _filters.brands.clear());
            });
          }
          final b = brands[i - 1];
          final active = _filters.brands.contains(b);
          return _chip(tt(b), active, () {
            setState(() {
              if (active) {
                _filters.brands.remove(b);
              } else {
                _filters.brands.add(b);
              }
            });
          });
        },
      ),
    );
  }

  Widget _chip(String label, bool active, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        alignment: Alignment.center,
        padding: const EdgeInsets.symmetric(horizontal: 13),
        decoration: BoxDecoration(
          color: active ? C.ink : C.card,
          borderRadius: BorderRadius.circular(20),
          border: active ? null : Border.all(color: C.line),
        ),
        child: Text(label,
            style: mr(
              size: 12.5,
              weight: active ? FontWeight.w700 : FontWeight.w600,
              color: active ? Colors.white : C.muted2,
            )),
      ),
    );
  }

  Widget _list() {
    final items = _visible;
    if (items.isEmpty) {
      return EmptyState(t('cat.noMatch'));
    }
    return ListView.separated(
      padding: const EdgeInsets.fromLTRB(18, 12, 18, 20),
      itemCount: items.length,
      separatorBuilder: (_, __) => const SizedBox(height: 12),
      itemBuilder: (context, i) => ProductRow(
        items[i],
        onTap: () => openProduct(context, product: items[i]),
      ),
    );
  }
}
