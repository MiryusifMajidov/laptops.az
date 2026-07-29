import 'package:flutter/material.dart';
import '../i18n.dart';
import '../theme.dart';

/// Yuxarıda "telefon bildirişi" kimi popup banner.
/// - Overlay-də göstərilir → heç bir səhifənin layout-una təsir etmir.
/// - 5 saniyə görünür, sonra öz-özünə bağlanır (altında azalan xətt).
/// - Sağa/sola sürüşdürüb silmək olar (swipe-to-dismiss).
/// - Üstünə basanda [onTap] işə düşür.

OverlayEntry? _active;

void showTopNotifBanner(BuildContext context,
    {required String message, required VoidCallback onTap}) {
  final overlay = Overlay.of(context);
  // əvvəlki banner qalıbsa sil — üst-üstə düşməsin
  _active?.remove();
  _active = null;

  late OverlayEntry entry;
  void close() {
    if (_active == entry) _active = null;
    if (entry.mounted) entry.remove();
  }

  entry = OverlayEntry(
    builder: (_) => _TopBanner(message: message, onTap: onTap, onClose: close),
  );
  _active = entry;
  overlay.insert(entry);
}

class _TopBanner extends StatefulWidget {
  final String message;
  final VoidCallback onTap;
  final VoidCallback onClose;
  const _TopBanner({required this.message, required this.onTap, required this.onClose});

  @override
  State<_TopBanner> createState() => _TopBannerState();
}

class _TopBannerState extends State<_TopBanner> with TickerProviderStateMixin {
  late final AnimationController _enter =
      AnimationController(vsync: this, duration: const Duration(milliseconds: 280))..forward();
  late final AnimationController _life =
      AnimationController(vsync: this, duration: const Duration(seconds: 5))
        ..addStatusListener((s) {
          if (s == AnimationStatus.completed) _autoClose();
        })
        ..forward();
  bool _closing = false;

  Future<void> _autoClose() async {
    if (_closing) return;
    _closing = true;
    _life.stop();
    if (mounted) await _enter.reverse(); // yuxarı sürüşərək itir
    widget.onClose();
  }

  void _tap() {
    if (_closing) return;
    _closing = true;
    _life.stop();
    widget.onTap();
    widget.onClose();
  }

  @override
  void dispose() {
    _enter.dispose();
    _life.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final slide = Tween<Offset>(begin: const Offset(0, -1.2), end: Offset.zero)
        .animate(CurvedAnimation(parent: _enter, curve: Curves.easeOutCubic));
    return Positioned(
      top: 0,
      left: 0,
      right: 0,
      child: SafeArea(
        bottom: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(14, 8, 14, 0),
          child: SlideTransition(
            position: slide,
            child: FadeTransition(
              opacity: _enter,
              child: Dismissible(
                key: const ValueKey('notif-banner'),
                direction: DismissDirection.horizontal,
                resizeDuration: null,
                onDismissed: (_) {
                  _closing = true;
                  _life.stop();
                  widget.onClose();
                },
                child: _card(),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _card() {
    return GestureDetector(
      onTap: _tap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 13, vertical: 11),
        decoration: BoxDecoration(
          color: Colors.white.withOpacity(.94),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: C.ink.withOpacity(.06)),
          boxShadow: [
            BoxShadow(color: C.ink.withOpacity(.16), blurRadius: 26, offset: const Offset(0, 10)),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
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
                          Text(t('home.now'), style: mr(size: 11, weight: FontWeight.w600, color: C.muted)),
                        ],
                      ),
                      const SizedBox(height: 1),
                      Text(t('home.newArrived').replaceAll('{p}', widget.message),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: mr(size: 12.5, weight: FontWeight.w500, color: const Color(0xFF3A4252))),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            // 5 saniyə ərzində soldan azalan geri sayım xətti
            SizedBox(
              width: double.infinity,
              height: 3,
              child: Stack(
                children: [
                  Positioned.fill(
                    child: DecoratedBox(
                      decoration: BoxDecoration(
                        color: C.ink.withOpacity(.08),
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                  ),
                  AnimatedBuilder(
                    animation: _life,
                    builder: (_, __) => FractionallySizedBox(
                      alignment: Alignment.centerLeft,
                      widthFactor: (1 - _life.value).clamp(0.0, 1.0),
                      heightFactor: 1,
                      child: DecoratedBox(
                        decoration: BoxDecoration(
                          color: C.ink,
                          borderRadius: BorderRadius.circular(2),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
