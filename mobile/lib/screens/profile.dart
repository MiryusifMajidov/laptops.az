import 'package:flutter/material.dart';
import '../api.dart';
import '../i18n.dart';
import '../store.dart';
import '../theme.dart';
import 'root.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});
  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  bool _applyMode = false; // false = giriş, true = müraciət
  bool _busy = false;

  final _user = TextEditingController();
  final _pass = TextEditingController();
  final _name = TextEditingController();
  final _store = TextEditingController();
  final _phone = TextEditingController();

  @override
  void dispose() {
    _user.dispose();
    _pass.dispose();
    _name.dispose();
    _store.dispose();
    _phone.dispose();
    super.dispose();
  }

  void _toast(String m) => ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(m)));

  Future<void> _login() async {
    if (_user.text.trim().isEmpty || _pass.text.isEmpty) {
      _toast(t('profile.needLogin'));
      return;
    }
    setState(() => _busy = true);
    try {
      await Api.login(_user.text.trim(), _pass.text);
      if (!Store.I.isPartner) {
        await Store.I.logout();
        _toast(t('profile.notPartner'));
        return;
      }
      if (mounted) {
        // bütün tətbiqi yenidən qur → məhsullar optavoy qiymətlə yüklənir
        Navigator.of(context).pushAndRemoveUntil(
            MaterialPageRoute(builder: (_) => const RootScaffold()), (r) => false);
      }
    } catch (e) {
      _toast(e.toString());
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _apply() async {
    if (_name.text.trim().isEmpty || _phone.text.trim().isEmpty) {
      _toast(t('profile.needFields'));
      return;
    }
    setState(() => _busy = true);
    try {
      await Api.partnerApply(name: _name.text.trim(), storeName: _store.text.trim(), phone: _phone.text.trim());
      _name.clear();
      _store.clear();
      _phone.clear();
      setState(() => _applyMode = false);
      _toast(t('profile.applied'));
    } catch (e) {
      _toast(e.toString());
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _logout() async {
    await Store.I.logout();
    if (mounted) {
      Navigator.of(context).pushAndRemoveUntil(
          MaterialPageRoute(builder: (_) => const RootScaffold()), (r) => false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: C.appBg,
      appBar: AppBar(
        backgroundColor: C.appBg,
        elevation: 0,
        foregroundColor: C.ink,
        title: Text(t('profile.title'), style: mr(size: 17, weight: FontWeight.w800, color: C.ink)),
      ),
      body: AnimatedBuilder(
        animation: Store.I,
        builder: (context, _) {
          if (Store.I.isPartner) return _partnerView();
          return _guestView();
        },
      ),
    );
  }

  // giriş olmuş partner
  Widget _partnerView() {
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        Container(
          width: 66,
          height: 66,
          decoration: BoxDecoration(color: C.ink, borderRadius: BorderRadius.circular(18)),
          child: const Icon(Icons.storefront_rounded, color: Colors.white, size: 32),
        ),
        const SizedBox(height: 16),
        Text(Store.I.userName.isEmpty ? t('profile.partner') : Store.I.userName,
            style: mr(size: 20, weight: FontWeight.w800, color: C.ink)),
        const SizedBox(height: 6),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          margin: const EdgeInsets.only(top: 6, bottom: 24),
          decoration: BoxDecoration(color: const Color(0xFFE7F5EE), borderRadius: BorderRadius.circular(10)),
          child: Row(mainAxisSize: MainAxisSize.min, children: [
            const Icon(Icons.check_circle_rounded, color: Color(0xFF0E9F6E), size: 18),
            const SizedBox(width: 8),
            Text(t('profile.wholesaleActive'), style: mr(size: 13.5, weight: FontWeight.w700, color: const Color(0xFF0A7E57))),
          ]),
        ),
        Text(t('profile.wholesaleDesc'),
            style: mr(size: 13.5, weight: FontWeight.w500, color: C.muted2, height: 1.5)),
        const SizedBox(height: 28),
        _btn(t('profile.logout'), _logout, outline: true),
      ],
    );
  }

  // qonaq (giriş olmayıb)
  Widget _guestView() {
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        // həvəsləndirici mətn
        Container(
          padding: const EdgeInsets.all(18),
          decoration: BoxDecoration(color: C.ink, borderRadius: BorderRadius.circular(18)),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            const Icon(Icons.handshake_rounded, color: Color(0xFF4C86FF), size: 30),
            const SizedBox(height: 12),
            Text(t('profile.heading'), style: mr(size: 19, weight: FontWeight.w800, color: Colors.white)),
            const SizedBox(height: 8),
            Text(t('profile.marketing'),
                style: mr(size: 13.5, weight: FontWeight.w500, color: const Color(0xFFB9C4DC), height: 1.6)),
          ]),
        ),
        const SizedBox(height: 22),

        if (!_applyMode) ..._loginForm() else ..._applyForm(),
      ],
    );
  }

  List<Widget> _loginForm() {
    return [
      Text(t('profile.login'), style: mr(size: 15, weight: FontWeight.w800, color: C.ink)),
      const SizedBox(height: 12),
      _input(_user, t('profile.username')),
      const SizedBox(height: 10),
      _input(_pass, t('profile.password'), obscure: true),
      const SizedBox(height: 16),
      _btn(_busy ? t('profile.checking') : t('profile.signin'), _busy ? null : _login),
      const SizedBox(height: 18),
      Center(
        child: GestureDetector(
          onTap: () => setState(() => _applyMode = true),
          child: Text(t('profile.applyLink'),
              style: mr(size: 13.5, weight: FontWeight.w700, color: const Color(0xFF1A57E0))),
        ),
      ),
    ];
  }

  List<Widget> _applyForm() {
    return [
      Text(t('profile.applyTitle'), style: mr(size: 15, weight: FontWeight.w800, color: C.ink)),
      const SizedBox(height: 4),
      Text(t('profile.applyDesc'),
          style: mr(size: 12.5, weight: FontWeight.w500, color: C.muted2, height: 1.5)),
      const SizedBox(height: 14),
      _input(_name, t('profile.name')),
      const SizedBox(height: 10),
      _input(_store, t('profile.store')),
      const SizedBox(height: 10),
      _input(_phone, t('profile.phone'), keyboard: TextInputType.phone),
      const SizedBox(height: 16),
      _btn(_busy ? t('profile.sending') : t('profile.applySend'), _busy ? null : _apply),
      const SizedBox(height: 16),
      Center(
        child: GestureDetector(
          onTap: () => setState(() => _applyMode = false),
          child: Text(t('profile.backLogin'), style: mr(size: 13.5, weight: FontWeight.w700, color: C.muted2)),
        ),
      ),
    ];
  }

  Widget _input(TextEditingController c, String hint, {bool obscure = false, TextInputType? keyboard}) {
    return TextField(
      controller: c,
      obscureText: obscure,
      keyboardType: keyboard,
      style: mr(size: 15, weight: FontWeight.w600, color: C.ink),
      decoration: InputDecoration(
        hintText: hint,
        hintStyle: mr(size: 14.5, color: C.muted3),
        filled: true,
        fillColor: C.card,
        contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
        enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: C.line)),
        focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: C.ink, width: 1.5)),
      ),
    );
  }

  Widget _btn(String label, VoidCallback? onTap, {bool outline = false}) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(vertical: 15),
        decoration: BoxDecoration(
          color: outline ? Colors.transparent : C.ink,
          borderRadius: BorderRadius.circular(13),
          border: outline ? Border.all(color: C.line2) : null,
        ),
        child: Center(
          child: Text(label, style: mr(size: 15, weight: FontWeight.w700, color: outline ? C.ink : Colors.white)),
        ),
      ),
    );
  }
}
