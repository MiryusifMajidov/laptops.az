import 'package:flutter/material.dart';
import '../store.dart';
import '../i18n.dart';
import '../theme.dart';
import 'home.dart';
import 'categories.dart';
import 'search.dart';
import 'orders.dart';
import 'ai_chat.dart';

/// Alt naviqasiya + 4 tab (dizayndakı kimi: Ana / Kateqoriya / Axtarış / Sifariş).
class RootScaffold extends StatefulWidget {
  const RootScaffold({super.key});
  @override
  State<RootScaffold> createState() => RootScaffoldState();
}

class RootScaffoldState extends State<RootScaffold> {
  int _tab = 0;

  void goTo(int i) => setState(() => _tab = i);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: C.appBg,
      body: IndexedStack(
        index: _tab,
        children: [
          HomeScreen(onSeeSearch: () => goTo(2), onSeeCategories: () => goTo(1)),
          const CategoriesScreen(),
          const SearchScreen(),
          const OrdersScreen(),
        ],
      ),
      bottomNavigationBar: _BottomNav(current: _tab, onTap: goTo),
    );
  }
}

class _BottomNav extends StatelessWidget {
  final int current;
  final ValueChanged<int> onTap;
  const _BottomNav({required this.current, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: Store.I,
      builder: (context, _) {
        final hasOrders = Store.I.orders.isNotEmpty;
        return Container(
          decoration: const BoxDecoration(
            color: C.appBg,
            border: Border(top: BorderSide(color: C.line, width: .5)),
          ),
          padding: EdgeInsets.only(
            top: 10,
            bottom: 12 + MediaQuery.of(context).padding.bottom,
            left: 14,
            right: 14,
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              _NavItem(
                icon: Icons.home_outlined,
                activeIcon: Icons.home_rounded,
                label: t('nav.home'),
                active: current == 0,
                onTap: () => onTap(0),
              ),
              _NavItem(
                icon: Icons.grid_view_outlined,
                activeIcon: Icons.grid_view_rounded,
                label: t('nav.category'),
                active: current == 1,
                onTap: () => onTap(1),
              ),
              // AI "Söhbət" — digər tab-larla eyni ölçüdə
              GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTap: () => Navigator.of(context).push(
                  MaterialPageRoute(builder: (_) => const AiChatScreen()),
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.auto_awesome, size: 23, color: Color(0xFF4C86FF)),
                    const SizedBox(height: 4),
                    Text(t('nav.chat'),
                        style: mr(size: 10, weight: FontWeight.w600, color: C.muted3)),
                  ],
                ),
              ),
              _NavItem(
                icon: Icons.search_rounded,
                activeIcon: Icons.search_rounded,
                label: t('nav.search'),
                active: current == 2,
                onTap: () => onTap(2),
              ),
              _NavItem(
                icon: Icons.receipt_long_outlined,
                activeIcon: Icons.receipt_long_rounded,
                label: t('nav.orders'),
                active: current == 3,
                badge: hasOrders,
                onTap: () => onTap(3),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _NavItem extends StatelessWidget {
  final IconData icon;
  final IconData activeIcon;
  final String label;
  final bool active;
  final bool badge;
  final VoidCallback onTap;
  const _NavItem({
    required this.icon,
    required this.activeIcon,
    required this.label,
    required this.active,
    required this.onTap,
    this.badge = false,
  });

  @override
  Widget build(BuildContext context) {
    final color = active ? C.ink : C.muted3;
    return GestureDetector(
      behavior: HitTestBehavior.opaque,
      onTap: onTap,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Stack(
            clipBehavior: Clip.none,
            children: [
              Icon(active ? activeIcon : icon, size: 23, color: color),
              if (badge)
                Positioned(
                  top: -2,
                  right: -3,
                  child: Container(
                    width: 7,
                    height: 7,
                    decoration: BoxDecoration(
                      color: C.danger,
                      shape: BoxShape.circle,
                      border: Border.all(color: C.appBg, width: 1),
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: 4),
          Text(label,
              style: mr(size: 10, weight: active ? FontWeight.w700 : FontWeight.w600, color: color)),
        ],
      ),
    );
  }
}
