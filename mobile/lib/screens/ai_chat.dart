import 'package:flutter/material.dart';
import '../api.dart';
import '../i18n.dart';
import '../theme.dart';
import '../widgets/net_image.dart';
import 'nav.dart';

/// AI Köməkçi söhbət ekranı (dizayn: APP 06).
/// Müştəriyə uyğun notebook seçməkdə kömək edir; sifariş/rezervi backend özü yaradır.
class AiChatScreen extends StatefulWidget {
  /// Paylaş vərəqindən gələn ilkin sual (varsa avtomatik göndərilir).
  final String? initialQuestion;
  const AiChatScreen({this.initialQuestion, super.key});
  @override
  State<AiChatScreen> createState() => _AiChatScreenState();
}

class _ChatMsg {
  final String role; // "user" | "assistant"
  final String text;
  final Map<String, Product> products;
  _ChatMsg(this.role, this.text, [this.products = const <String, Product>{}]);
}

/// Söhbət tarixçəsi modul səviyyəsində saxlanılır — ekrandan çıxıb yenidən
/// daxil olanda mesajlar qalır (sessiya boyu; tətbiq tam bağlananda sıfırlanır).
final List<_ChatMsg> _history = <_ChatMsg>[];

class _AiChatScreenState extends State<AiChatScreen> {
  List<_ChatMsg> get _msgs => _history;
  final _input = TextEditingController();
  final _scroll = ScrollController();
  bool _loading = false;

  List<List<String>> get _quick => [
        [t('ai.qGameL'), t('ai.qGameQ')],
        [t('ai.qWorkL'), t('ai.qWorkQ')],
        [t('ai.qDesignL'), t('ai.qDesignQ')],
        [t('ai.qStudentL'), t('ai.qStudentQ')],
      ];

  @override
  void initState() {
    super.initState();
    if (_history.isEmpty) {
      _history.add(_ChatMsg('assistant', t('ai.welcome')));
    }
    final q = widget.initialQuestion?.trim();
    if (q != null && q.isNotEmpty) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _send(q));
    } else {
      // əvvəlki söhbət varsa aşağıya sürüşdür
      WidgetsBinding.instance.addPostFrameCallback((_) => _scrollToBottom());
    }
  }

  @override
  void dispose() {
    _input.dispose();
    _scroll.dispose();
    super.dispose();
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scroll.hasClients) {
        _scroll.animateTo(_scroll.position.maxScrollExtent,
            duration: const Duration(milliseconds: 250), curve: Curves.easeOut);
      }
    });
  }

  Future<void> _send(String raw) async {
    final txt = raw.trim();
    if (txt.isEmpty || _loading) return;
    _input.clear();
    setState(() {
      _msgs.add(_ChatMsg('user', txt));
      _loading = true;
    });
    _scrollToBottom();
    try {
      final history =
          _msgs.map((m) => AiMsg(m.role, m.text)).toList(growable: false);
      final res = await Api.aiChat(history);
      setState(() => _msgs.add(_ChatMsg('assistant', res.reply, res.products)));
    } catch (_) {
      setState(() => _msgs.add(_ChatMsg('assistant', t('ai.error'))));
    } finally {
      setState(() => _loading = false);
      _scrollToBottom();
    }
  }

  String _clean(String s) => s
      .replaceAll('**', '')
      .replaceAll(RegExp(r'^#{1,6}\s*', multiLine: true), '')
      .replaceAll(RegExp(r'^\s*[-*]\s+', multiLine: true), '• ')
      .trim();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: C.appBg,
      resizeToAvoidBottomInset: true,
      body: SafeArea(
        bottom: false,
        child: Column(
          children: [
            _header(),
            Expanded(
              child: ListView(
                controller: _scroll,
                padding: const EdgeInsets.fromLTRB(15, 16, 15, 10),
                children: [
                  Center(
                    child: Text(t('common.today'),
                        style: mr(
                            size: 11,
                            weight: FontWeight.w600,
                            color: C.muted3)),
                  ),
                  const SizedBox(height: 12),
                  for (int i = 0; i < _msgs.length; i++) ...[
                    _bubble(_msgs[i]),
                    if (i == 0) _quickChips(),
                    const SizedBox(height: 12),
                  ],
                  if (_loading) _typing(),
                ],
              ),
            ),
            _inputBar(),
          ],
        ),
      ),
    );
  }

  Widget _header() {
    return Container(
      padding: const EdgeInsets.fromLTRB(6, 2, 16, 12),
      decoration: const BoxDecoration(
        color: C.appBg,
        border: Border(bottom: BorderSide(color: C.line2)),
      ),
      child: Row(
        children: [
          IconButton(
            onPressed: () => Navigator.of(context).maybePop(),
            icon: const Icon(Icons.arrow_back_ios_new_rounded,
                size: 20, color: C.ink),
          ),
          _avatar(36),
          const SizedBox(width: 11),
          Expanded(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(t('ai.title'), style: mr(size: 15.5, weight: FontWeight.w800)),
                Text(t('ai.status'),
                    style: mr(
                        size: 11.5, weight: FontWeight.w700, color: C.greenTx)),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _avatar(double size) {
    return SizedBox(
      width: size,
      height: size,
      child: Stack(
        clipBehavior: Clip.none,
        children: [
          Container(
            width: size,
            height: size,
            decoration: BoxDecoration(
                color: C.ink, borderRadius: BorderRadius.circular(size * .3)),
            child: Icon(Icons.auto_awesome,
                color: const Color(0xFF4C86FF), size: size * .55),
          ),
          Positioned(
            bottom: -1,
            right: -1,
            child: Container(
              width: 11,
              height: 11,
              decoration: BoxDecoration(
                color: C.green,
                shape: BoxShape.circle,
                border: Border.all(color: C.appBg, width: 2),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _bubble(_ChatMsg m) {
    if (m.role == 'user') {
      return Align(
        alignment: Alignment.centerRight,
        child: Container(
          margin: const EdgeInsets.only(left: 50),
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 11),
          decoration: const BoxDecoration(
            color: C.ink,
            borderRadius: BorderRadius.only(
              topLeft: Radius.circular(16),
              topRight: Radius.circular(5),
              bottomLeft: Radius.circular(16),
              bottomRight: Radius.circular(16),
            ),
          ),
          child: Text(m.text,
              style: mr(
                  size: 13.5,
                  weight: FontWeight.w500,
                  color: Colors.white,
                  height: 1.55)),
        ),
      );
    }
    // assistant
    return Row(
      crossAxisAlignment: CrossAxisAlignment.end,
      children: [
        Container(
          width: 27,
          height: 27,
          margin: const EdgeInsets.only(right: 8),
          decoration:
              BoxDecoration(color: C.ink, borderRadius: BorderRadius.circular(8)),
          child: const Icon(Icons.auto_awesome,
              color: Color(0xFF4C86FF), size: 14),
        ),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: _assistantParts(m),
          ),
        ),
      ],
    );
  }

  List<Widget> _assistantParts(_ChatMsg m) {
    final re = RegExp(r'\[\[product:(\d+)\]\]');
    final out = <Widget>[];
    var last = 0;
    for (final match in re.allMatches(m.text)) {
      final seg = _clean(m.text.substring(last, match.start));
      if (seg.isNotEmpty) out.add(_textBubble(seg));
      final p = m.products[match.group(1)];
      if (p != null) out.add(_productCard(p));
      last = match.end;
    }
    final tail = _clean(m.text.substring(last));
    if (tail.isNotEmpty) out.add(_textBubble(tail));
    if (out.isEmpty) out.add(_textBubble(_clean(m.text)));
    return out
        .expand((w) => [w, const SizedBox(height: 8)])
        .toList()
      ..removeLast();
  }

  Widget _textBubble(String text) {
    return Container(
      constraints: const BoxConstraints(maxWidth: 280),
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 11),
      decoration: BoxDecoration(
        color: C.card,
        border: Border.all(color: C.line2),
        borderRadius: const BorderRadius.only(
          topLeft: Radius.circular(5),
          topRight: Radius.circular(16),
          bottomLeft: Radius.circular(16),
          bottomRight: Radius.circular(16),
        ),
      ),
      child: Text(text,
          style: mr(size: 13.5, weight: FontWeight.w500, color: C.ink, height: 1.55)),
    );
  }

  Widget _productCard(Product p) {
    return GestureDetector(
      onTap: () => openProduct(context, product: p),
      child: Container(
        constraints: const BoxConstraints(maxWidth: 290),
        padding: const EdgeInsets.all(10),
        decoration: BoxDecoration(
          color: C.card,
          border: Border.all(color: C.line2),
          borderRadius: BorderRadius.circular(13),
        ),
        child: Row(
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(9),
              child: SizedBox(width: 66, height: 54, child: NetImage(p.cardImage)),
            ),
            const SizedBox(width: 11),
            Expanded(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(p.name,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: mr(size: 12.5, weight: FontWeight.w700, height: 1.3)),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      Text(money(p.price), style: sg(size: 14, weight: FontWeight.w600)),
                      const SizedBox(width: 6),
                      Container(
                          width: 5,
                          height: 5,
                          decoration: const BoxDecoration(
                              color: C.green, shape: BoxShape.circle)),
                      const SizedBox(width: 3),
                      Text(t('common.available'),
                          style: mr(
                              size: 10,
                              weight: FontWeight.w700,
                              color: C.greenTx)),
                    ],
                  ),
                ],
              ),
            ),
            const Icon(Icons.chevron_right_rounded, color: C.ink, size: 20),
          ],
        ),
      ),
    );
  }

  Widget _quickChips() {
    return Padding(
      padding: const EdgeInsets.only(left: 35, top: 10),
      child: Wrap(
        spacing: 7,
        runSpacing: 7,
        children: [
          for (final q in _quick)
            GestureDetector(
              onTap: () => _send(q[1]),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 7),
                decoration: BoxDecoration(
                  color: C.card,
                  border: Border.all(color: C.line3),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(q[0],
                    style: mr(size: 12, weight: FontWeight.w600, color: C.ink)),
              ),
            ),
        ],
      ),
    );
  }

  Widget _typing() {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.end,
      children: [
        Container(
          width: 27,
          height: 27,
          margin: const EdgeInsets.only(right: 8),
          decoration:
              BoxDecoration(color: C.ink, borderRadius: BorderRadius.circular(8)),
          child: const Icon(Icons.auto_awesome,
              color: Color(0xFF4C86FF), size: 14),
        ),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 13),
          decoration: BoxDecoration(
            color: C.card,
            border: Border.all(color: C.line2),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              for (final o in [.9, .6, .35])
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 2),
                  child: Container(
                    width: 6,
                    height: 6,
                    decoration: BoxDecoration(
                        color: C.muted3.withOpacity(o),
                        shape: BoxShape.circle),
                  ),
                ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _inputBar() {
    return Container(
      padding: EdgeInsets.fromLTRB(
          15, 11, 15, 12 + MediaQuery.of(context).padding.bottom),
      decoration: const BoxDecoration(
        color: C.card,
        border: Border(top: BorderSide(color: C.line2)),
      ),
      child: Row(
        children: [
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                color: const Color(0xFFF5F5F2),
                border: Border.all(color: C.line),
                borderRadius: BorderRadius.circular(14),
              ),
              child: TextField(
                controller: _input,
                textInputAction: TextInputAction.send,
                onSubmitted: _send,
                style: mr(size: 14, weight: FontWeight.w500),
                decoration: InputDecoration(
                  hintText: t('ai.inputHint'),
                  hintStyle: mr(size: 14, color: C.muted3),
                  border: InputBorder.none,
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                ),
              ),
            ),
          ),
          const SizedBox(width: 9),
          GestureDetector(
            onTap: () => _send(_input.text),
            child: Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                  color: C.ink, borderRadius: BorderRadius.circular(11)),
              child: const Icon(Icons.send_rounded, color: Colors.white, size: 18),
            ),
          ),
        ],
      ),
    );
  }
}
