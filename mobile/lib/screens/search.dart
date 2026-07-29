import 'dart:async';
import 'package:flutter/material.dart';
import '../api.dart';
import '../i18n.dart';
import '../store.dart';
import '../theme.dart';
import '../widgets/common.dart';
import '../widgets/product_row.dart';
import 'nav.dart';

class SearchScreen extends StatefulWidget {
  const SearchScreen({super.key});
  @override
  State<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen> {
  final _ctrl = TextEditingController();
  final _focus = FocusNode();
  Timer? _debounce;

  String _query = '';
  bool _loading = false;
  Object? _error;
  List<Product> _results = [];

  @override
  void initState() {
    super.initState();
    _focus.addListener(() {
      if (mounted) setState(() {});
    });
  }

  @override
  void dispose() {
    _debounce?.cancel();
    _ctrl.dispose();
    _focus.dispose();
    super.dispose();
  }

  void _onChanged(String v) {
    setState(() => _query = v);
    _debounce?.cancel();
    if (v.trim().isEmpty) {
      setState(() {
        _results = [];
        _error = null;
        _loading = false;
      });
      return;
    }
    _debounce = Timer(const Duration(milliseconds: 350), () => _run(v));
  }

  Future<void> _run(String v) async {
    final term = v.trim();
    if (term.isEmpty) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final res = await Api.products(q: term);
      if (!mounted || _ctrl.text.trim() != term) return;
      setState(() {
        _results = res;
        _loading = false;
      });
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e;
          _loading = false;
        });
      }
    }
  }

  void _submit() {
    final term = _ctrl.text.trim();
    if (term.isEmpty) return;
    Store.I.addSearch(term);
    _run(term);
  }

  void _useTerm(String term) {
    _ctrl.text = term;
    _ctrl.selection = TextSelection.collapsed(offset: term.length);
    setState(() => _query = term);
    _run(term);
    _focus.requestFocus();
  }

  void _clear() {
    _ctrl.clear();
    _onChanged('');
    _focus.requestFocus();
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
            child: Text(t('search.title'), style: sg(size: 26, weight: FontWeight.w600, spacing: -.4)),
          ),
          _searchField(),
          Expanded(child: _content()),
        ],
      ),
    );
  }

  Widget _searchField() {
    final active = _focus.hasFocus || _query.isNotEmpty;
    return Padding(
      padding: const EdgeInsets.fromLTRB(18, 0, 18, 14),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 4),
        decoration: BoxDecoration(
          color: C.card,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: active ? C.ink : C.line, width: active ? 1.5 : 1),
        ),
        child: Row(
          children: [
            Icon(Icons.search_rounded, size: 18, color: active ? C.ink : C.muted3),
            const SizedBox(width: 9),
            Expanded(
              child: TextField(
                controller: _ctrl,
                focusNode: _focus,
                onChanged: _onChanged,
                onSubmitted: (_) => _submit(),
                textInputAction: TextInputAction.search,
                style: mr(size: 15, weight: FontWeight.w600),
                decoration: InputDecoration(
                  isCollapsed: true,
                  border: InputBorder.none,
                  hintText: t('search.hint'),
                  hintStyle: mr(size: 14.5, weight: FontWeight.w500, color: C.muted3),
                ),
              ),
            ),
            if (_query.isNotEmpty)
              GestureDetector(
                onTap: _clear,
                child: Icon(Icons.cancel_rounded, size: 18, color: C.muted3),
              ),
          ],
        ),
      ),
    );
  }

  Widget _content() {
    if (_query.trim().isEmpty) return _recent();
    if (_loading) return const Center(child: CircularProgressIndicator(color: C.ink));
    if (_error != null) {
      return ErrorState(_error.toString(), onRetry: () => _run(_query));
    }
    if (_results.isEmpty) {
      return EmptyState(t('search.noResult').replaceAll('{q}', _query.trim()));
    }
    return ListView(
      padding: const EdgeInsets.fromLTRB(18, 0, 18, 20),
      children: [
        Padding(
          padding: const EdgeInsets.only(bottom: 10),
          child: Text(t('search.results').replaceAll('{n}', '${_results.length}'),
              style: mr(size: 11, weight: FontWeight.w700, color: C.muted3, spacing: .8)),
        ),
        for (final p in _results) ...[
          ProductRowCompact(p, onTap: () {
            Store.I.addSearch(_query.trim());
            openProduct(context, product: p);
          }),
          const SizedBox(height: 10),
        ],
      ],
    );
  }

  Widget _recent() {
    return AnimatedBuilder(
      animation: Store.I,
      builder: (context, _) {
        final recent = Store.I.recentSearches;
        if (recent.isEmpty) {
          return EmptyState(t('search.prompt'));
        }
        return ListView(
          padding: const EdgeInsets.fromLTRB(18, 4, 18, 20),
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(t('search.recent'),
                    style: mr(size: 11, weight: FontWeight.w700, color: C.muted3, spacing: .8)),
                GestureDetector(
                  onTap: () => Store.I.clearSearches(),
                  child: Text(t('common.clear'),
                      style: mr(size: 12, weight: FontWeight.w700, color: C.muted2)),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                for (final term in recent)
                  GestureDetector(
                    onTap: () => _useTerm(term),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                      decoration: BoxDecoration(
                        color: C.card,
                        borderRadius: BorderRadius.circular(20),
                        border: Border.all(color: C.line),
                      ),
                      child: Text(term, style: mr(size: 13, weight: FontWeight.w600, color: C.muted2)),
                    ),
                  ),
              ],
            ),
          ],
        );
      },
    );
  }
}
