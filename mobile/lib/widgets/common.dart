import 'package:flutter/material.dart';
import '../i18n.dart';
import '../theme.dart';

/// «Mövcuddur» yaşıl nöqtə + mətn.
class AvailableTag extends StatelessWidget {
  final bool compact;
  const AvailableTag({this.compact = false, super.key});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 6,
          height: 6,
          decoration: const BoxDecoration(color: C.green, shape: BoxShape.circle),
        ),
        const SizedBox(width: 5),
        Text(t('common.available'),
            style: mr(size: compact ? 11 : 11.5, weight: FontWeight.w700, color: C.greenTx)),
      ],
    );
  }
}

/// Sadə yaşıl nöqtə (kartlarda).
class Dot extends StatelessWidget {
  final Color color;
  final double size;
  const Dot({this.color = C.green, this.size = 7, super.key});
  @override
  Widget build(BuildContext context) => Container(
        width: size,
        height: size,
        decoration: BoxDecoration(color: color, shape: BoxShape.circle),
      );
}

/// Seçilə bilən çip (filtr / marka).
class PillChip extends StatelessWidget {
  final String label;
  final bool active;
  final VoidCallback onTap;
  final bool rounded;
  const PillChip({
    required this.label,
    required this.active,
    required this.onTap,
    this.rounded = false,
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 120),
        padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 9),
        decoration: BoxDecoration(
          color: active ? C.ink : C.card,
          borderRadius: BorderRadius.circular(rounded ? 20 : 11),
          border: active ? null : Border.all(color: C.line),
        ),
        child: Text(
          label,
          style: mr(
            size: 13,
            weight: active ? FontWeight.w700 : FontWeight.w600,
            color: active ? Colors.white : C.muted2,
          ),
        ),
      ),
    );
  }
}

/// Səhv / boş vəziyyət mesajı (retry düyməsi ilə).
class ErrorState extends StatelessWidget {
  final String message;
  final VoidCallback? onRetry;
  const ErrorState(this.message, {this.onRetry, super.key});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(30),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.wifi_off_rounded, color: C.muted3, size: 40),
            const SizedBox(height: 14),
            Text(message,
                textAlign: TextAlign.center,
                style: mr(size: 13.5, weight: FontWeight.w600, color: C.muted2)),
            if (onRetry != null) ...[
              const SizedBox(height: 16),
              OutlinedButton(
                onPressed: onRetry,
                style: OutlinedButton.styleFrom(
                  foregroundColor: C.ink,
                  side: const BorderSide(color: C.line3),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(11)),
                  padding: const EdgeInsets.symmetric(horizontal: 22, vertical: 12),
                ),
                child: Text(t('common.retry'), style: mr(size: 13.5, weight: FontWeight.w700, color: C.ink)),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

/// Boş nəticə (məhsul yoxdur).
class EmptyState extends StatelessWidget {
  final String message;
  const EmptyState(this.message, {super.key});
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(30),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.inventory_2_outlined, color: C.muted3, size: 38),
            const SizedBox(height: 12),
            Text(message,
                textAlign: TextAlign.center,
                style: mr(size: 13.5, weight: FontWeight.w600, color: C.muted2)),
          ],
        ),
      ),
    );
  }
}

/// Dairəvi ikon düyməsi (geri / paylaş və s.).
class RoundIconButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback onTap;
  final double size;
  const RoundIconButton({required this.icon, required this.onTap, this.size = 38, super.key});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          color: C.card,
          shape: BoxShape.circle,
          border: Border.all(color: C.line),
        ),
        child: Icon(icon, size: size * 0.46, color: C.ink),
      ),
    );
  }
}
