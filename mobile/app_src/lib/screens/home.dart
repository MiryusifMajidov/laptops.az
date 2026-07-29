import 'package:flutter/material.dart';
import '../api.dart';
import '../theme.dart';
import '../widgets/common.dart';
import '../widgets/net_image.dart';
import '../widgets/product_card.dart';
import 'nav.dart';

class HomeScreen extends StatefulWidget {
  final VoidCallback onSeeSearch;
  final VoidCallback onSeeCategories;
  const HomeScreen({required this.onSeeSearch, required this.onSeeCategories, super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  late Future<_HomeData> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<_HomeData> _load() async {
    final results = await Future.wait([Api.products(), Api.categories()]);
    return _HomeData(
      products: results[0] as List<Product>,
      cats: results[1] as List<CatCount>,
    );
  }

  Future<void> _refresh() async {
    final data = await _load();
    if (mounted) setState(() => _future = Future.value(data));
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      bottom: false,
      child: Column(
        children: [
          _topBar(),
          Expanded(
            child: FutureBuilder<_HomeData>(
              future: _future,
              builder: (context, snap) {
                if (snap.connectionState == ConnectionState.waiting) {
                  return const Center(child: CircularProgressIndicator(color: C.ink));
                }
                if (snap.hasError) {
                  return ErrorState(
                    snap.error.toString(),
                    onRetry: () => setState(() => _future = _load()),
                  );
                }
                final data = snap.data!;
                if (data.products.isEmpty) {
                  return RefreshIndicator(
                    onRefresh: _refresh,
                    color: C.ink,
                    child: ListView(children: const [
                      SizedBox(height: 160),
                      EmptyState('Hazırda saytda məhsul yoxdur'),
                    ]),
                  );
                }
                return RefreshIndicator(
                  onRefresh: _refresh,
                  color: C.ink,
                  child: _content(data),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _topBar() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 6, 20, 10),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          GestureDetector(
            onLongPress: () => showServerDialog(context),
            child: Image.asset('assets/logo-trim.png', height: 22),
          ),
          Stack(
            clipBehavior: Clip.none,
            children: [
              Icon(Icons.notifications_none_rounded, size: 25, color: C.ink),
              Positioned(
                top: 0,
                right: 1,
                child: Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: C.danger,
                    shape: BoxShape.circle,
                    border: Border.all(color: C.appBg, width: 1.5),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _content(_HomeData data) {
    final hero = data.products.first;
    final popular = data.products.length > 1 ? data.products.sublist(1) : <Product>[];

    return ListView(
      padding: const EdgeInsets.only(bottom: 24),
      children: [
        _pushCard(hero),
        _searchBar(),
        _hero(hero),
        if (data.cats.isNotEmpty) _catChips(data.cats),
        _popular(popular.isEmpty ? data.products : popular),
      ],
    );
  }

  // Push bildiriş kartı (şüşə effekti) — ən yeni məhsul.
  Widget _pushCard(Product newest) {
    return Container(
      margin: const EdgeInsets.fromLTRB(14, 2, 14, 0),
      padding: const EdgeInsets.symmetric(horizontal: 13, vertical: 11),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.82),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: C.ink.withOpacity(.06)),
        boxShadow: [
          BoxShadow(
            color: C.ink.withOpacity(.14),
            blurRadius: 26,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 38,
            height: 38,
            decoration: BoxDecoration(color: C.ink, borderRadius: BorderRadius.circular(10)),
            child: const Icon(Icons.laptop_mac, color: Colors.white, size: 20),
          ),
          const SizedBox(width: 11),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('Laptops.az', style: mr(size: 13, weight: FontWeight.w700)),
                    Text('indi', style: mr(size: 11, weight: FontWeight.w600, color: C.muted)),
                  ],
                ),
                const SizedBox(height: 1),
                Text('Yeni gəldi: ${newest.name} — stokda!',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: mr(size: 12.5, weight: FontWeight.w500, color: const Color(0xFF3A4252))),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _searchBar() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 14, 20, 0),
      child: GestureDetector(
        onTap: widget.onSeeSearch,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
          decoration: BoxDecoration(
            color: C.card,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: C.line),
          ),
          child: Row(
            children: [
              Icon(Icons.search_rounded, size: 18, color: C.muted3),
              const SizedBox(width: 9),
              Text('Notebook, marka, model axtar…',
                  style: mr(size: 14, weight: FontWeight.w500, color: C.muted3)),
            ],
          ),
        ),
      ),
    );
  }

  Widget _hero(Product p) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 16, 20, 0),
      child: GestureDetector(
        onTap: () => openProduct(context, product: p),
        child: Container(
          decoration: BoxDecoration(
            color: C.card,
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: C.line2),
            boxShadow: [
              BoxShadow(
                  color: C.ink.withOpacity(.10),
                  blurRadius: 16,
                  offset: const Offset(0, 4)),
            ],
          ),
          clipBehavior: Clip.antiAlias,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Stack(
                children: [
                  SizedBox(height: 172, width: double.infinity, child: NetImage(p.cardImage)),
                  Positioned(
                    top: 12,
                    left: 12,
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(20),
                        boxShadow: [
                          BoxShadow(color: C.ink.withOpacity(.1), blurRadius: 6),
                        ],
                      ),
                      child: const AvailableTag(),
                    ),
                  ),
                ],
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(17, 15, 17, 17),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '${p.brand} · Seçilmiş'.toUpperCase(),
                      style: mr(size: 12, weight: FontWeight.w700, color: C.muted, spacing: .5),
                    ),
                    const SizedBox(height: 4),
                    Text(p.name,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: mr(size: 18, weight: FontWeight.w800, color: C.ink)),
                    const SizedBox(height: 14),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(money(p.price), style: sg(size: 23, weight: FontWeight.w600)),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 11),
                          decoration: BoxDecoration(
                              color: C.ink, borderRadius: BorderRadius.circular(12)),
                          child: Text('Sifariş et',
                              style: mr(size: 13.5, weight: FontWeight.w700, color: Colors.white)),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _catChips(List<CatCount> cats) {
    return Padding(
      padding: const EdgeInsets.only(top: 18),
      child: SizedBox(
        height: 38,
        child: ListView.separated(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.symmetric(horizontal: 20),
          itemCount: cats.length,
          separatorBuilder: (_, __) => const SizedBox(width: 9),
          itemBuilder: (context, i) {
            final c = cats[i];
            final active = i == 0;
            return GestureDetector(
              onTap: () => openCategory(context, c.name),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 9),
                decoration: BoxDecoration(
                  color: active ? C.ink : C.card,
                  borderRadius: BorderRadius.circular(11),
                  border: active ? null : Border.all(color: C.line),
                ),
                child: Text(c.name,
                    style: mr(
                      size: 13,
                      weight: active ? FontWeight.w700 : FontWeight.w600,
                      color: active ? Colors.white : C.muted2,
                    )),
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _popular(List<Product> items) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 20, 20, 0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Populyar', style: mr(size: 15, weight: FontWeight.w800)),
          const SizedBox(height: 12),
          GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: items.length,
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              crossAxisSpacing: 12,
              mainAxisSpacing: 12,
              childAspectRatio: .74,
            ),
            itemBuilder: (context, i) => ProductCard(
              items[i],
              onTap: () => openProduct(context, product: items[i]),
            ),
          ),
        ],
      ),
    );
  }
}

class _HomeData {
  final List<Product> products;
  final List<CatCount> cats;
  _HomeData({required this.products, required this.cats});
}
