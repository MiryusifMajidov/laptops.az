import 'package:flutter/material.dart';
import '../api.dart';
import '../i18n.dart';
import '../theme.dart';

/// Seçilmiş filtrlər.
class Filters {
  Set<String> brands; // «Marka» dəyərləri
  Map<String, Set<String>> attrs; // digər xüsusiyyət → seçilmiş dəyərlər
  RangeValues? price; // null = tam diapazon (filtr yoxdur)

  Filters({Set<String>? brands, Map<String, Set<String>>? attrs, this.price})
      : brands = brands ?? {},
        attrs = attrs ?? {};

  Filters copy() => Filters(
        brands: {...brands},
        attrs: {for (final e in attrs.entries) e.key: {...e.value}},
        price: price,
      );

  bool get isEmpty => brands.isEmpty && attrs.values.every((s) => s.isEmpty) && price == null;

  bool matches(Product p) {
    if (brands.isNotEmpty && !brands.contains(p.brand)) return false;
    for (final e in attrs.entries) {
      if (e.value.isNotEmpty && !e.value.contains(p.attr(e.key))) return false;
    }
    if (price != null && (p.price < price!.start || p.price > price!.end)) return false;
    return true;
  }
}

/// Məhsul siyahısından filtr fasetlərini (mümkün dəyərləri) hesabla.
class Facets {
  final List<String> brands;
  final Map<String, List<String>> attrs; // Marka istisna olmaqla
  final double minPrice;
  final double maxPrice;

  Facets(this.brands, this.attrs, this.minPrice, this.maxPrice);

  factory Facets.from(List<Product> items) {
    final brandSet = <String>{};
    final attrMap = <String, Set<String>>{};
    double lo = double.infinity, hi = 0;
    const skip = {'Marka'};
    for (final p in items) {
      if (p.brand.isNotEmpty) brandSet.add(p.brand);
      for (final v in p.values) {
        if (v.value.isEmpty || skip.contains(v.attribute)) continue;
        (attrMap[v.attribute] ??= <String>{}).add(v.value);
      }
      if (p.price < lo) lo = p.price.toDouble();
      if (p.price > hi) hi = p.price.toDouble();
    }
    if (lo == double.infinity) lo = 0;
    // yalnız 2+ fərqli dəyəri olan xüsusiyyətləri filtrdə göstər
    final attrs = <String, List<String>>{};
    for (final e in attrMap.entries) {
      if (e.value.length >= 2) attrs[e.key] = e.value.toList()..sort();
    }
    return Facets(brandSet.toList()..sort(), attrs, lo, hi);
  }
}

/// Filtr panelini aç. Tətbiq edilərsə yeni [Filters] qaytarır, əks halda null.
Future<Filters?> showFilterSheet(
  BuildContext context, {
  required Facets facets,
  required Filters current,
  required int Function(Filters) countFor,
}) {
  return showModalBottomSheet<Filters>(
    context: context,
    isScrollControlled: true,
    backgroundColor: Colors.transparent,
    builder: (ctx) => _FilterSheet(facets: facets, current: current.copy(), countFor: countFor),
  );
}

class _FilterSheet extends StatefulWidget {
  final Facets facets;
  final Filters current;
  final int Function(Filters) countFor;
  const _FilterSheet({required this.facets, required this.current, required this.countFor});

  @override
  State<_FilterSheet> createState() => _FilterSheetState();
}

class _FilterSheetState extends State<_FilterSheet> {
  late Filters f;
  late RangeValues _range;
  bool _priceTouched = false;

  double get _lo => widget.facets.minPrice;
  double get _hi => widget.facets.maxPrice <= widget.facets.minPrice
      ? widget.facets.minPrice + 1
      : widget.facets.maxPrice;

  @override
  void initState() {
    super.initState();
    f = widget.current;
    _range = f.price ?? RangeValues(_lo, _hi);
    _priceTouched = f.price != null;
  }

  void _toggleBrand(String b) {
    setState(() {
      if (f.brands.contains(b)) {
        f.brands.remove(b);
      } else {
        f.brands.add(b);
      }
    });
  }

  void _toggleAttr(String attr, String v) {
    setState(() {
      final set = f.attrs[attr] ??= <String>{};
      if (set.contains(v)) {
        set.remove(v);
      } else {
        set.add(v);
      }
    });
  }

  void _reset() {
    setState(() {
      f = Filters();
      _range = RangeValues(_lo, _hi);
      _priceTouched = false;
    });
  }

  Filters get _applied {
    final out = f.copy();
    out.price = _priceTouched ? _range : null;
    return out;
  }

  @override
  Widget build(BuildContext context) {
    final count = widget.countFor(_applied);
    return Container(
      constraints: BoxConstraints(maxHeight: MediaQuery.of(context).size.height * .88),
      decoration: const BoxDecoration(
        color: C.appBg,
        borderRadius: BorderRadius.vertical(top: Radius.circular(26)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const SizedBox(height: 12),
          Container(
            width: 38,
            height: 5,
            decoration: BoxDecoration(color: const Color(0xFFDAD7CF), borderRadius: BorderRadius.circular(3)),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 12, 20, 14),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(t('common.filter'), style: sg(size: 20, weight: FontWeight.w600, spacing: -.3)),
                GestureDetector(
                  onTap: _reset,
                  child: Text(t('common.reset'), style: mr(size: 13.5, weight: FontWeight.w700, color: C.muted2)),
                ),
              ],
            ),
          ),
          const Divider(height: 1, color: Color(0xFFF0EEE8)),
          Flexible(
            child: ListView(
              padding: const EdgeInsets.fromLTRB(20, 18, 20, 18),
              children: [
                if (widget.facets.brands.isNotEmpty) ...[
                  _label(tt('Marka')),
                  _chips(widget.facets.brands, (b) => f.brands.contains(b), _toggleBrand),
                  const SizedBox(height: 22),
                ],
                _label(t('filter.price')),
                Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: Text('${money(_range.start)} – ${money(_range.end)}',
                      style: sg(size: 12.5, weight: FontWeight.w600, color: C.muted)),
                ),
                RangeSlider(
                  values: _range,
                  min: _lo,
                  max: _hi,
                  activeColor: C.ink,
                  inactiveColor: C.line,
                  labels: RangeLabels(money(_range.start), money(_range.end)),
                  onChanged: (v) => setState(() {
                    _range = v;
                    _priceTouched = true;
                  }),
                ),
                const SizedBox(height: 16),
                for (final e in widget.facets.attrs.entries) ...[
                  _label(tt(e.key)),
                  _chips(e.value, (v) => (f.attrs[e.key] ?? const {}).contains(v),
                      (v) => _toggleAttr(e.key, v)),
                  const SizedBox(height: 22),
                ],
              ],
            ),
          ),
          Container(
            decoration: const BoxDecoration(
              color: C.card,
              border: Border(top: BorderSide(color: C.line, width: .5)),
            ),
            padding: EdgeInsets.fromLTRB(20, 13, 20, 20 + MediaQuery.of(context).padding.bottom),
            child: GestureDetector(
              onTap: () => Navigator.pop(context, _applied),
              child: Container(
                alignment: Alignment.center,
                padding: const EdgeInsets.symmetric(vertical: 16),
                decoration: BoxDecoration(
                  color: C.ink,
                  borderRadius: BorderRadius.circular(14),
                  boxShadow: [
                    BoxShadow(color: C.ink.withOpacity(.35), blurRadius: 20, offset: const Offset(0, 8)),
                  ],
                ),
                child: Text(t('filter.showResults').replaceAll('{n}', '$count'),
                    style: mr(size: 15.5, weight: FontWeight.w700, color: Colors.white)),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _label(String label) => Padding(
        padding: const EdgeInsets.only(bottom: 12),
        child: Text(label, style: mr(size: 13, weight: FontWeight.w800)),
      );

  Widget _chips(List<String> values, bool Function(String) isOn, void Function(String) onTap) {
    return Wrap(
      spacing: 9,
      runSpacing: 9,
      children: [
        for (final v in values)
          GestureDetector(
            onTap: () => onTap(v),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 9),
              decoration: BoxDecoration(
                color: isOn(v) ? C.ink : C.card,
                borderRadius: BorderRadius.circular(11),
                border: isOn(v) ? null : Border.all(color: C.line3),
              ),
              child: Text(tt(v),
                  style: mr(
                    size: 13,
                    weight: isOn(v) ? FontWeight.w700 : FontWeight.w600,
                    color: isOn(v) ? Colors.white : C.muted2,
                  )),
            ),
          ),
      ],
    );
  }
}
