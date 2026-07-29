import 'package:flutter/material.dart';
import '../api.dart';
import '../store.dart';
import '../theme.dart';
import '../widgets/net_image.dart';

/// APP05 — sifariş təsdiqi. Ödəniş yoxdur; ad + telefon kifayətdir.
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

  Future<void> _submit() async {
    final name = _name.text.trim();
    final phone = _phone.text.trim();
    if (name.isEmpty) {
      _snack('Zəhmət olmasa ad, soyad yaz');
      return;
    }
    if (phone.isEmpty) {
      _snack('Zəhmət olmasa telefon nömrəsi yaz');
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
        _snack(e.toString());
      }
    }
  }

  void _snack(String m) =>
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(m)));

  void _showSuccess(String ref) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => Dialog(
        backgroundColor: C.appBg,
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
                child: Icon(Icons.check_rounded, color: C.greenTx, size: 34),
              ),
              const SizedBox(height: 18),
              Text('Sifariş qəbul edildi', style: sg(size: 20, weight: FontWeight.w600)),
              const SizedBox(height: 8),
              Text(
                'Nömrə: $ref\nMəhsul 24 saat saxlanılır — mağaza sizə zəng edəcək.',
                textAlign: TextAlign.center,
                style: mr(size: 13, weight: FontWeight.w500, color: C.muted2, height: 1.5),
              ),
              const SizedBox(height: 22),
              SizedBox(
                width: double.infinity,
                child: GestureDetector(
                  onTap: () {
                    Navigator.pop(ctx); // dialoq
                    Navigator.pop(context); // sifariş forması
                    Navigator.pop(context); // məhsul detalı
                  },
                  child: Container(
                    alignment: Alignment.center,
                    padding: const EdgeInsets.symmetric(vertical: 15),
                    decoration: BoxDecoration(color: C.ink, borderRadius: BorderRadius.circular(14)),
                    child: Text('Bağla', style: mr(size: 15, weight: FontWeight.w700, color: Colors.white)),
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
      body: SafeArea(
        bottom: false,
        child: Column(
          children: [
            _header(),
            Expanded(
              child: ListView(
                padding: const EdgeInsets.fromLTRB(18, 0, 18, 18),
                children: [
                  _summary(p),
                  const SizedBox(height: 16),
                  _field('AD, SOYAD', _name, hint: 'məs. Elvin Məmmədov'),
                  const SizedBox(height: 12),
                  _field('TELEFON', _phone,
                      hint: '+994 50 000 00 00', keyboard: TextInputType.phone, mono: true),
                  const SizedBox(height: 12),
                  _field('QEYD (İSTƏYƏ BAĞLI)', _note,
                      hint: 'Nə vaxt gələ bilərəm…', maxLines: 3),
                  const SizedBox(height: 16),
                  _infoBox(),
                ],
              ),
            ),
          ],
        ),
      ),
      bottomNavigationBar: _bottomBar(),
    );
  }

  Widget _header() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(18, 6, 18, 10),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text('Sifariş', style: sg(size: 22, weight: FontWeight.w600, spacing: -.3)),
          GestureDetector(
            onTap: () => Navigator.pop(context),
            child: Container(
              width: 34,
              height: 34,
              decoration: BoxDecoration(
                color: C.card,
                shape: BoxShape.circle,
                border: Border.all(color: C.line),
              ),
              child: Icon(Icons.close_rounded, size: 18, color: C.ink),
            ),
          ),
        ],
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
                Text('1 ədəd', style: mr(size: 12, weight: FontWeight.w500, color: C.muted)),
              ],
            ),
          ),
          const SizedBox(width: 8),
          Text(money(p.price), style: sg(size: 17, weight: FontWeight.w600)),
        ],
      ),
    );
  }

  Widget _field(
    String label,
    TextEditingController ctrl, {
    String? hint,
    TextInputType? keyboard,
    int maxLines = 1,
    bool mono = false,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 2, bottom: 6),
          child: Text(label, style: mr(size: 11, weight: FontWeight.w700, color: C.muted, spacing: .3)),
        ),
        TextField(
          controller: ctrl,
          keyboardType: keyboard,
          maxLines: maxLines,
          style: mono ? sg(size: 14.5) : mr(size: 14.5, weight: FontWeight.w600),
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
        ),
      ],
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
          Icon(Icons.check_circle_outline_rounded, size: 16, color: C.greenTx),
          const SizedBox(width: 9),
          Expanded(
            child: RichText(
              text: TextSpan(
                style: mr(size: 12.5, weight: FontWeight.w600, color: C.greenTx, height: 1.5),
                children: const [
                  TextSpan(text: 'Ödəniş yoxdur. ', style: TextStyle(fontWeight: FontWeight.w800)),
                  TextSpan(text: 'Məhsul 24 saat saxlanılır, mağaza sizə zəng edir.'),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _bottomBar() {
    return Container(
      decoration: const BoxDecoration(
        color: C.card,
        border: Border(top: BorderSide(color: C.line, width: .5)),
      ),
      padding: EdgeInsets.fromLTRB(18, 13, 18, 16 + MediaQuery.of(context).padding.bottom),
      child: GestureDetector(
        onTap: _sending ? null : _submit,
        child: Container(
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(vertical: 16),
          decoration: BoxDecoration(
            color: _sending ? C.ink.withOpacity(.55) : C.ink,
            borderRadius: BorderRadius.circular(14),
            boxShadow: _sending
                ? null
                : [BoxShadow(color: C.ink.withOpacity(.35), blurRadius: 20, offset: const Offset(0, 8))],
          ),
          child: _sending
              ? const SizedBox(
                  width: 22,
                  height: 22,
                  child: CircularProgressIndicator(strokeWidth: 2.4, color: Colors.white),
                )
              : Text('Sifarişi təsdiqlə', style: mr(size: 15.5, weight: FontWeight.w700, color: Colors.white)),
        ),
      ),
    );
  }
}
