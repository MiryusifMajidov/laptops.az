import 'package:audioplayers/audioplayers.dart';
import 'package:flutter/material.dart';
import '../api.dart';
import '../store.dart';
import '../i18n.dart';
import '../notif_bus.dart';
import '../theme.dart';
import '../widgets/common.dart';
import '../widgets/net_image.dart';
import '../widgets/notif_banner.dart';
import '../widgets/product_card.dart';
import 'nav.dart';
import 'root.dart';

class HomeScreen extends StatefulWidget {
  final VoidCallback onSeeSearch;
  final VoidCallback onSeeCategories;
  const HomeScreen({required this.onSeeSearch, required this.onSeeCategories, super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  late Future<_HomeData> _future;
  int _newestId = 0; // ən yeni məhsulun id-si (bildiriş qırmızı nöqtəsi üçün)
  final AudioPlayer _player = AudioPlayer();

  @override
  void initState() {
    super.initState();
    // tətbiq açıqkən push gələndə — yuxarıda popup banner göstər
    foregroundPushTick.addListener(_onForegroundPush);
    _future = _load();
    _future.then((d) {
      // ilk açılışda baxılmamış yeni məhsul varsa — səssiz popup göstər
      if (mounted && d.products.isNotEmpty && Store.I.hasUnread(_newestId)) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) {
            _showNamedBanner(name: d.products.first.name, itemId: d.products.first.id, sound: false);
          }
        });
      }
    }).catchError((_) {});
  }

  @override
  void dispose() {
    foregroundPushTick.removeListener(_onForegroundPush);
    _player.dispose();
    super.dispose();
  }

  Future<_HomeData> _load() async {
    final results = await Future.wait([Api.products(), Api.categories(), Api.featuredItemId()]);
    final products = results[0] as List<Product>;
    _newestId = products.isNotEmpty ? products.first.id : 0;
    return _HomeData(products: products, cats: results[1] as List<CatCount>, featuredId: results[2] as int);
  }

  Future<void> _refresh() async {
    final data = await _load();
    if (mounted) setState(() => _future = Future.value(data));
  }

  // Tətbiq AÇIQ ikən yeni məhsul push-u gəldi — SƏHİFƏNİ YENİLƏMƏDƏN yuxarıda popup + səs.
  Future<void> _onForegroundPush() async {
    try {
      final items = await Api.notifications();
      if (!mounted || items.isEmpty) return;
      final n = items.first;
      setState(() => _newestId = n.id); // yalnız zəng qırmızı nöqtəsi (eyni _future → ekran reload olmur)
      _showNamedBanner(name: n.name, itemId: n.id, sound: true);
    } catch (_) {}
  }

  // Yuxarıda telefon bildirişi kimi popup (heç bir layout-a təsir etmir; 5 san, sürüşdürüb silmək olar).
  void _showNamedBanner({required String name, required int itemId, required bool sound}) {
    if (sound) _ding();
    showTopNotifBanner(context, message: name, onTap: () => openProduct(context, id: itemId));
  }

  Future<void> _ding() async {
    try {
      await _player.stop();
      await _player.play(AssetSource('notif.wav'));
    } catch (_) {}
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
                    child: ListView(children: [
                      const SizedBox(height: 160),
                      EmptyState(t('home.empty')),
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

  void _showLangSheet() {
    showModalBottomSheet(
      context: context,
      backgroundColor: C.appBg,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(22))),
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(height: 16),
            Text(t('lang.title'), style: mr(size: 15, weight: FontWeight.w800, color: C.ink)),
            const SizedBox(height: 6),
            for (final l in L.I.langs)
              ListTile(
                title: Text(l.name, style: mr(size: 15, weight: FontWeight.w600, color: C.ink)),
                trailing: l.code == L.I.lang ? const Icon(Icons.check_rounded, color: C.ink) : null,
                onTap: () async {
                  Navigator.pop(ctx);
                  if (l.code != L.I.lang) {
                    await L.I.setLang(l.code);
                    if (mounted) {
                      Navigator.of(context).pushAndRemoveUntil(
                          MaterialPageRoute(builder: (_) => const RootScaffold()), (r) => false);
                    }
                  }
                },
              ),
            const SizedBox(height: 12),
          ],
        ),
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
          Row(
            children: [
              GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTap: _showLangSheet,
                child: Padding(
                  padding: const EdgeInsets.all(4),
                  child: Text(L.I.lang.toUpperCase(),
                      style: mr(size: 13, weight: FontWeight.w800, color: C.ink)),
                ),
              ),
              const SizedBox(width: 10),
              GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTap: () => openProfile(context),
                child: const Padding(
                  padding: EdgeInsets.all(4),
                  child: Icon(Icons.person_outline_rounded, size: 25, color: C.ink),
                ),
              ),
              const SizedBox(width: 6),
              GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTap: () => openNotifications(context),
                child: AnimatedBuilder(
                  animation: Store.I,
                  builder: (context, _) {
                    final unread = _newestId > 0 && Store.I.hasUnread(_newestId);
                    return Stack(
                      clipBehavior: Clip.none,
                      children: [
                        const Padding(
                          padding: EdgeInsets.all(4),
                          child: Icon(Icons.notifications_none_rounded, size: 25, color: C.ink),
                        ),
                        if (unread)
                          Positioned(
                            top: 2,
                            right: 3,
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
                    );
                  },
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _content(_HomeData data) {
    final hero = _pickHero(data); // böyük kart — Site Settings-də seçilmiş məhsul
    final popular = data.products.where((p) => p.id != hero.id).toList();

    return ListView(
      padding: const EdgeInsets.only(bottom: 24),
      children: [
        _searchBar(),
        _hero(hero),
        if (data.cats.isNotEmpty) _catChips(data.cats),
        if (popular.isNotEmpty) _popular(popular), // hero məhsulu təkrar göstərmə
      ],
    );
  }

  // seçilmiş məhsul (varsa) — yoxdursa ən yeni
  Product _pickHero(_HomeData data) {
    if (data.featuredId > 0) {
      for (final p in data.products) {
        if (p.id == data.featuredId) return p;
      }
    }
    return data.products.first;
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
              Text(t('home.searchHint'),
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
                      '${tt(p.brand)} · ${t('home.featured')}'.toUpperCase(),
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
                        GestureDetector(
                          behavior: HitTestBehavior.opaque,
                          onTap: () => openOrder(context, p), // birbaşa sifariş
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 11),
                            decoration: BoxDecoration(
                                color: C.ink, borderRadius: BorderRadius.circular(12)),
                            child: Text(t('common.order'),
                                style: mr(size: 13.5, weight: FontWeight.w700, color: Colors.white)),
                          ),
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
                child: Text(tt(c.name),
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
          Text(t('home.popular'), style: mr(size: 15, weight: FontWeight.w800)),
          const SizedBox(height: 12),
          GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: items.length,
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              crossAxisSpacing: 12,
              mainAxisSpacing: 12,
              mainAxisExtent: 250, // sabit hündürlük — overflow olmasın, kartlar bərabər
            ),
            itemBuilder: (context, i) => ProductCard(
              items[i],
              onTap: () => openProduct(context, product: items[i]),
              onOrder: () => openOrder(context, items[i]),
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
  final int featuredId;
  _HomeData({required this.products, required this.cats, this.featuredId = 0});
}
