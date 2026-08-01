import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';
import 'package:url_launcher/url_launcher.dart';
import '../api.dart';
import '../config.dart';
import '../i18n.dart';
import '../theme.dart';
import '../screens/ai_chat.dart';
import 'common.dart';
import 'net_image.dart';

/// Məhsulun saytdakı linki (paylaşım / kopyalama üçün).
String _productUrl(Product p) => '${AppConfig.host}/mehsul/${p.id}';

/// Məhsulun paylaşıldığı sətir (ad + qiymət + sayt linki).
String _shareText(Product p) => '${p.name} — ${money(p.finalPrice)}\n${_productUrl(p)}';

/// APP07 — «Paylaş» bottom sheet. Link kopyala / OS paylaşım / SMS / E-poçt / AI.
void showShareSheet(BuildContext context, Product product) {
  showModalBottomSheet(
    context: context,
    backgroundColor: C.appBg,
    isScrollControlled: true,
    shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
    builder: (_) => _ShareSheet(parentContext: context, p: product),
  );
}

class _ShareSheet extends StatefulWidget {
  final BuildContext parentContext;
  final Product p;
  const _ShareSheet({required this.parentContext, required this.p});

  @override
  State<_ShareSheet> createState() => _ShareSheetState();
}

class _ShareSheetState extends State<_ShareSheet> {
  final _ask = TextEditingController();

  @override
  void dispose() {
    _ask.dispose();
    super.dispose();
  }

  Product get p => widget.p;

  void _close() => Navigator.of(context).pop();

  void _openAi(String? question) {
    _close();
    Navigator.of(widget.parentContext).push(
      MaterialPageRoute(builder: (_) => AiChatScreen(initialQuestion: question)),
    );
  }

  // AI-yə göndəriləcək mesaj — məhsul HƏMİŞƏ konteksdə olsun.
  // Boş sual → "bu məhsul haqqında məlumat ver"; yazılan sual → sual + məhsul adı.
  String _aiMessage(String? typed) {
    final q = typed?.trim() ?? '';
    if (q.isEmpty) {
      return t('ai.aboutProduct').replaceAll('{p}', p.name);
    }
    return '$q\n\n(${p.name})';
  }

  void _whatsapp() =>
      _launch(Uri.parse('https://wa.me/?text=${Uri.encodeComponent(_shareText(p))}'));

  void _toast(String m) => ScaffoldMessenger.of(widget.parentContext)
      .showSnackBar(SnackBar(content: Text(m)));

  Future<void> _copyLink() async {
    final messenger = ScaffoldMessenger.of(widget.parentContext);
    await Clipboard.setData(ClipboardData(text: _shareText(p)));
    _close();
    messenger.showSnackBar(SnackBar(content: Text(t('share.copied'))));
  }

  Future<void> _osShare() async {
    _close();
    await SharePlus.instance.share(ShareParams(text: _shareText(p)));
  }

  Future<void> _launch(Uri uri) async {
    _close();
    try {
      final ok = await launchUrl(uri, mode: LaunchMode.externalApplication);
      if (!ok) {
        await Clipboard.setData(ClipboardData(text: _shareText(p)));
        _toast(t('share.copied'));
      }
    } catch (_) {
      // əlaqəli tətbiq yoxdursa — linki kopyala
      await Clipboard.setData(ClipboardData(text: _shareText(p)));
      _toast(t('share.copied'));
    }
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      top: false,
      child: Padding(
        padding: EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(height: 10),
            Container(
              width: 38,
              height: 5,
              decoration: BoxDecoration(
                  color: const Color(0xFFDAD7CF), borderRadius: BorderRadius.circular(3)),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 14, 20, 6),
              child: Align(
                alignment: Alignment.centerLeft,
                child: Text(t('share.title'), style: sg(size: 20, weight: FontWeight.w600, spacing: -.3)),
              ),
            ),
            _productRow(),
            const SizedBox(height: 6),
            _options(),
            const SizedBox(height: 8),
            _askBar(),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }

  Widget _productRow() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 6, 20, 10),
      child: Row(
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(10),
            child: SizedBox(width: 52, height: 44, child: NetImage(p.cardImage)),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(p.name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: mr(size: 14, weight: FontWeight.w700)),
                const SizedBox(height: 2),
                Row(
                  children: [
                    Text(money(p.finalPrice),
                        style: sg(size: 13, weight: FontWeight.w600, color: C.muted2)),
                    if (p.hasDiscount) ...[
                      const SizedBox(width: 6),
                      Text(money(p.price),
                          style: mr(size: 11, weight: FontWeight.w500, color: C.muted3)
                              .copyWith(decoration: TextDecoration.lineThrough)),
                    ],
                    const SizedBox(width: 8),
                    const AvailableTag(compact: true),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _options() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12),
      child: Row(
        children: [
          _opt(t('share.ai'), Icons.auto_awesome, () => _openAi(_aiMessage(null)),
              iconColor: const Color(0xFF4C86FF), bg: C.ink, accent: true),
          _opt(t('share.whatsapp'), Icons.chat_rounded, _whatsapp,
              iconColor: const Color(0xFF25D366)),
          _opt(t('share.copyLink'), Icons.link_rounded, _copyLink),
          _opt(t('share.more'), Icons.more_horiz_rounded, _osShare),
        ],
      ),
    );
  }

  Widget _opt(String label, IconData icon, VoidCallback onTap,
      {Color? iconColor, Color? bg, bool accent = false}) {
    return Expanded(
      child: GestureDetector(
        behavior: HitTestBehavior.opaque,
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 6),
          child: Column(
            children: [
              Stack(
                clipBehavior: Clip.none,
                children: [
                  Container(
                    width: 52,
                    height: 52,
                    decoration: BoxDecoration(
                      color: bg ?? C.card,
                      borderRadius: BorderRadius.circular(15),
                      border: accent ? null : Border.all(color: C.line3),
                    ),
                    child: Icon(icon, color: iconColor ?? C.ink, size: 22),
                  ),
                  if (accent)
                    Positioned(
                      bottom: -1,
                      right: -1,
                      child: Container(
                        width: 12,
                        height: 12,
                        decoration: BoxDecoration(
                          color: C.green,
                          shape: BoxShape.circle,
                          border: Border.all(color: C.appBg, width: 2),
                        ),
                      ),
                    ),
                ],
              ),
              const SizedBox(height: 7),
              Text(label,
                  maxLines: 2,
                  textAlign: TextAlign.center,
                  overflow: TextOverflow.ellipsis,
                  style: mr(size: 10.5, weight: FontWeight.w600, color: C.muted2, height: 1.2)),
            ],
          ),
        ),
      ),
    );
  }

  Widget _askBar() {
    return Container(
      margin: const EdgeInsets.fromLTRB(16, 4, 16, 4),
      decoration: BoxDecoration(
        color: const Color(0xFFF5F5F2),
        border: Border.all(color: C.line),
        borderRadius: BorderRadius.circular(14),
      ),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _ask,
              textInputAction: TextInputAction.send,
              onSubmitted: (v) {
                if (v.trim().isNotEmpty) _openAi(_aiMessage(v));
              },
              style: mr(size: 13.5, weight: FontWeight.w500),
              decoration: InputDecoration(
                hintText: t('share.askHint'),
                hintStyle: mr(size: 13, color: C.muted3),
                border: InputBorder.none,
                contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
              ),
            ),
          ),
          GestureDetector(
            onTap: () {
              final v = _ask.text.trim();
              if (v.isNotEmpty) _openAi(_aiMessage(v));
            },
            child: Container(
              margin: const EdgeInsets.all(6),
              width: 38,
              height: 38,
              decoration: BoxDecoration(color: C.ink, borderRadius: BorderRadius.circular(11)),
              child: const Icon(Icons.send_rounded, color: Colors.white, size: 17),
            ),
          ),
        ],
      ),
    );
  }
}
