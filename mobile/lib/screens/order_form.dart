import 'package:flutter/material.dart';
import '../api.dart';
import '../i18n.dart';
import '../store.dart';
import '../theme.dart';
import '../widgets/net_image.dart';

/// Sifariş təsdiqi — sadə, tək ListView struktur (heç bir Expanded/bottomNavigationBar yoxdur).
/// Ödəniş yoxdur; ad + telefon kifayətdir. Uğurlu olsa referans nömrəsi göstərilir.
class OrderFormScreen extends StatefulWidget {
  final Product product;
  const OrderFormScreen({required this.product, super.key});

  @override
  State<OrderFormScreen> createState() => _OrderFormScreenState();
}

class _OrderFormScreenState extends State<OrderFormScreen> {
  final _name = TextEditingController();
  final _phone = TextEditingController();
  final _note = TextEditingController();
  bool _sending = false;

  @override
  void dispose() {
    _name.dispose();
    _phone.dispose();
    _note.dispose();
    super.dispose();
  }

  void _toast(String m) =>
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(m)));

  Future<void> _submit() async {
    final name = _name.text.trim();
    final phone = _phone.text.trim();
    if (name.isEmpty) {
      _toast(t('order.needName'));
      return;
    }
    if (phone.isEmpty) {
      _toast(t('order.needPhone'));
      return;
    }
    setState(() => _sending = true);
    try {
      final ref = await Api.order(
        name: name,
        phone: phone,
        itemId: widget.product.id,
        note: _note.text.trim(),
      );
      await Store.I.addOrder(MyOrder(
        ref: ref.ref,
        itemId: widget.product.id,
        productName: widget.product.name,
        price: widget.product.price,
        date: DateTime.now().toIso8601String(),
      ));
      if (mounted) _showSuccess(ref.ref);
    } catch (e) {
      if (mounted) {
        setState(() => _sending = false);
        _toast(e.toString());
      }
    }
  }

  void _showSuccess(String ref) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => Dialog(
        backgroundColor: Colors.white,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(22)),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(24, 28, 24, 20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 64,
                height: 64,
                decoration: const BoxDecoration(color: C.greenBg, shape: BoxShape.circle),
                child: const Icon(Icons.check_rounded, color: C.greenTx, size: 34),
              ),
              const SizedBox(height: 18),
              Text(t('order.received'), style: sg(size: 20, weight: FontWeight.w600)),
              const SizedBox(height: 8),
              Text(
                t('order.receivedText').replaceAll('{ref}', ref),
                textAlign: TextAlign.center,
                style: mr(size: 13, weight: FontWeight.w500, color: C.muted2, height: 1.5),
              ),
              const SizedBox(height: 22),
              SizedBox(
                width: double.infinity,
                child: GestureDetector(
                  onTap: () {
                    Navigator.of(ctx).pop(); // dialoqu bağla
                    Navigator.of(context).pop(); // sifariş formasını bağla → məhsula qayıt
                  },
                  child: Container(
                    alignment: Alignment.center,
                    padding: const EdgeInsets.symmetric(vertical: 15),
                    decoration: BoxDecoration(color: C.ink, borderRadius: BorderRadius.circular(14)),
                    child: Text(t('common.close'),
                        style: mr(size: 15, weight: FontWeight.w700, color: Colors.white)),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final p = widget.product;
    return Scaffold(
      backgroundColor: C.appBg,
      appBar: AppBar(
        backgroundColor: C.appBg,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        foregroundColor: C.ink,
        title: Text(t('order.title'), style: sg(size: 19, weight: FontWeight.w600)),
        leading: IconButton(
          icon: const Icon(Icons.close_rounded, color: C.ink),
          onPressed: () => Navigator.of(context).maybePop(),
        ),
      ),
      body: SafeArea(
        top: false,
        child: ListView(
          padding: const EdgeInsets.fromLTRB(18, 8, 18, 28),
          children: [
            _summary(p),
            const SizedBox(height: 20),
            _label(t('order.nameLabel')),
            _input(_name, t('order.namePlaceholder')),
            const SizedBox(height: 14),
            _label(t('order.phoneLabel')),
            _input(_phone, '+994 50 000 00 00', keyboard: TextInputType.phone),
            const SizedBox(height: 14),
            _label(t('order.noteLabel')),
            _input(_note, t('order.notePlaceholder'), maxLines: 3),
            const SizedBox(height: 18),
            _infoBox(),
            const SizedBox(height: 22),
            _submitButton(),
          ],
        ),
      ),
    );
  }

  Widget _summary(Product p) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: C.card,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: C.line2),
      ),
      child: Row(
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(11),
            child: SizedBox(width: 64, height: 54, child: NetImage(p.cardImage)),
          ),
          const SizedBox(width: 13),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(p.name,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: mr(size: 14, weight: FontWeight.w700)),
                const SizedBox(height: 2),
                Text(t('order.qty'), style: mr(size: 12, weight: FontWeight.w500, color: C.muted)),
              ],
            ),
          ),
          const SizedBox(width: 8),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            mainAxisSize: MainAxisSize.min,
            children: [
              if (p.hasDiscount)
                Text(money(p.price),
                    style: mr(size: 11, weight: FontWeight.w500, color: C.muted)
                        .copyWith(decoration: TextDecoration.lineThrough)),
              Text(money(p.finalPrice), style: sg(size: 17, weight: FontWeight.w600)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _label(String text) => Padding(
        padding: const EdgeInsets.only(left: 2, bottom: 6),
        child: Text(text, style: mr(size: 11, weight: FontWeight.w700, color: C.muted, spacing: .3)),
      );

  Widget _input(TextEditingController ctrl, String hint,
      {TextInputType? keyboard, int maxLines = 1}) {
    return TextField(
      controller: ctrl,
      keyboardType: keyboard,
      maxLines: maxLines,
      style: mr(size: 14.5, weight: FontWeight.w600),
      decoration: InputDecoration(
        hintText: hint,
        hintStyle: mr(size: 13.5, weight: FontWeight.w500, color: C.muted3),
        filled: true,
        fillColor: C.card,
        contentPadding: const EdgeInsets.symmetric(horizontal: 15, vertical: 13),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(13),
          borderSide: const BorderSide(color: C.line3),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(13),
          borderSide: const BorderSide(color: C.ink, width: 1.5),
        ),
      ),
    );
  }

  Widget _infoBox() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: C.greenBg2,
        borderRadius: BorderRadius.circular(13),
        border: Border.all(color: C.greenLine),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.check_circle_outline_rounded, size: 16, color: C.greenTx),
          const SizedBox(width: 9),
          Expanded(
            child: Text(
              t('order.payNote'),
              style: mr(size: 12.5, weight: FontWeight.w600, color: C.greenTx, height: 1.5),
            ),
          ),
        ],
      ),
    );
  }

  Widget _submitButton() {
    return GestureDetector(
      onTap: _sending ? null : _submit,
      child: Container(
        alignment: Alignment.center,
        padding: const EdgeInsets.symmetric(vertical: 16),
        decoration: BoxDecoration(
          color: _sending ? const Color(0xFF8A93A6) : C.ink,
          borderRadius: BorderRadius.circular(14),
        ),
        child: _sending
            ? const SizedBox(
                width: 22,
                height: 22,
                child: CircularProgressIndicator(strokeWidth: 2.4, color: Colors.white),
              )
            : Text(t('order.submit'),
                style: mr(size: 15.5, weight: FontWeight.w700, color: Colors.white)),
      ),
    );
  }
}
