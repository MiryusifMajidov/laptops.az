import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../api.dart';
import '../config.dart';
import '../theme.dart';
import '../widgets/common.dart';
import '../widgets/net_image.dart';
import 'order_form.dart';

/// APP03 — məhsul detalı. [initial] siyahıdan gələn məhsul (dərhal göstərmək üçün),
/// tam məlumat (qalereya, bütün xüsusiyyətlər) [id] ilə serverdən yüklənir.
class ProductScreen extends StatefulWidget {
  final int? id;
  final Product? initial;
  const ProductScreen({this.id, this.initial, super.key});

  @override
  State<ProductScreen> createState() => _ProductScreenState();
}

class _ProductScreenState extends State<ProductScreen> {
  Product? _p;
  Object? _error;
  final _pageCtrl = PageController();
  int _page = 0;

  int get _pid => widget.id ?? widget.initial!.id;

  @override
  void initState() {
    super.initState();
    _p = widget.initial;
    _fetch();
  }

  Future<void> _fetch() async {
    try {
      final full = await Api.product(_pid);
      if (mounted) setState(() => _p = full);
    } catch (e) {
      if (mounted && _p == null) setState(() => _error = e);
    }
  }

  @override
  void dispose() {
    _pageCtrl.dispose();
    super.dispose();
  }

  Future<void> _call() async {
    final uri = Uri.parse('tel:$kStorePhone');
    try {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Zəng: $kStorePhone')),
        );
      }
    }
  }

  void _order() {
    final p = _p!;
    Navigator.of(context).push(MaterialPageRoute(builder: (_) => OrderFormScreen(product: p)));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: C.appBg,
      body: SafeArea(
        bottom: false,
        child: _error != null && _p == null
            ? Column(children: [
                _topButtons(),
                Expanded(child: ErrorState(_error.toString(), onRetry: () {
                  setState(() => _error = null);
                  _fetch();
                })),
              ])
            : _p == null
                ? const Center(child: CircularProgressIndicator(color: C.ink))
                : _body(_p!),
      ),
      bottomNavigationBar: _p == null ? null : _bottomBar(),
    );
  }

  Widget _topButtons() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(18, 6, 18, 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          RoundIconButton(icon: Icons.arrow_back_ios_new_rounded, onTap: () => Navigator.pop(context)),
          RoundIconButton(icon: Icons.ios_share_rounded, onTap: () {}),
        ],
      ),
    );
  }

  Widget _body(Product p) {
    final imgs = p.images;
    return Column(
      children: [
        _topButtons(),
        Expanded(
          child: ListView(
            padding: const EdgeInsets.fromLTRB(18, 0, 18, 18),
            children: [
              // şəkil karuseli
              ClipRRect(
                borderRadius: BorderRadius.circular(20),
                child: SizedBox(
                  height: 250,
                  child: imgs.isEmpty
                      ? const NetImage('')
                      : PageView.builder(
                          controller: _pageCtrl,
                          onPageChanged: (i) => setState(() => _page = i),
                          itemCount: imgs.length,
                          itemBuilder: (_, i) => NetImage(imgs[i]),
                        ),
                ),
              ),
              if (imgs.length > 1) ...[
                const SizedBox(height: 10),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    for (var i = 0; i < imgs.length; i++)
                      Container(
                        margin: const EdgeInsets.symmetric(horizontal: 3.5),
                        width: 7,
                        height: 7,
                        decoration: BoxDecoration(
                          color: i == _page ? C.ink : const Color(0xFFD8D5CD),
                          shape: BoxShape.circle,
                        ),
                      ),
                  ],
                ),
              ],
              const SizedBox(height: 14),
              if (p.brand.isNotEmpty)
                Text(p.brand.toUpperCase(),
                    style: mr(size: 12, weight: FontWeight.w700, color: C.muted, spacing: .5)),
              const SizedBox(height: 4),
              Text(p.name, style: sg(size: 25, weight: FontWeight.w600, spacing: -.5)),
              const SizedBox(height: 12),
              Row(
                children: [
                  Text(money(p.price), style: sg(size: 26, weight: FontWeight.w600)),
                  const SizedBox(width: 11),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 5),
                    decoration: BoxDecoration(color: C.greenBg, borderRadius: BorderRadius.circular(20)),
                    child: const AvailableTag(),
                  ),
                ],
              ),
              const SizedBox(height: 18),
              _specGrid(p),
              if (p.serial.isNotEmpty) ...[
                const SizedBox(height: 14),
                Text('Seriya: ${p.serial}',
                    style: mr(size: 12, weight: FontWeight.w600, color: C.muted3)),
              ],
            ],
          ),
        ),
      ],
    );
  }

  Widget _specGrid(Product p) {
    // ilk 4 (boş olmayan) xüsusiyyəti 2x2 grid-də göstər
    final specs = p.values.where((v) => v.value.isNotEmpty).take(4).toList();
    if (specs.isEmpty) return const SizedBox.shrink();
    return GridView.count(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      crossAxisCount: 2,
      crossAxisSpacing: 9,
      mainAxisSpacing: 9,
      childAspectRatio: 2.9,
      children: [
        for (final s in specs)
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 13, vertical: 11),
            decoration: BoxDecoration(
              color: C.card,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: C.line2),
            ),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(s.attribute.toUpperCase(),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: mr(size: 10.5, weight: FontWeight.w700, color: C.muted)),
                const SizedBox(height: 2),
                Text(s.value,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: mr(size: 13, weight: FontWeight.w700, color: C.ink)),
              ],
            ),
          ),
      ],
    );
  }

  Widget _bottomBar() {
    return Container(
      decoration: const BoxDecoration(
        color: C.card,
        border: Border(top: BorderSide(color: C.line, width: .5)),
      ),
      padding: EdgeInsets.fromLTRB(18, 12, 18, 16 + MediaQuery.of(context).padding.bottom),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.check_circle_outline_rounded, size: 14, color: C.greenTx),
              const SizedBox(width: 7),
              Text('Ödəniş yoxdur — mağazada Nağd/Kart/Taksit',
                  style: mr(size: 11.5, weight: FontWeight.w600, color: C.greenTx)),
            ],
          ),
          const SizedBox(height: 11),
          Row(
            children: [
              Expanded(
                child: GestureDetector(
                  onTap: _order,
                  child: Container(
                    alignment: Alignment.center,
                    padding: const EdgeInsets.symmetric(vertical: 15),
                    decoration: BoxDecoration(
                      color: C.ink,
                      borderRadius: BorderRadius.circular(14),
                      boxShadow: [
                        BoxShadow(color: C.ink.withOpacity(.35), blurRadius: 20, offset: const Offset(0, 8)),
                      ],
                    ),
                    child: Text('Sifariş et',
                        style: mr(size: 15.5, weight: FontWeight.w700, color: Colors.white)),
                  ),
                ),
              ),
              const SizedBox(width: 11),
              GestureDetector(
                onTap: _call,
                child: Container(
                  width: 54,
                  height: 52,
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(color: const Color(0xFFD8D5CD)),
                  ),
                  child: Icon(Icons.call_outlined, color: C.ink, size: 20),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
