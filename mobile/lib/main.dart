import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'config.dart';
import 'store.dart';
import 'i18n.dart';
import 'notif_bus.dart';
import 'theme.dart';
import 'screens/root.dart';

// Arxa planda (tətbiq bağlı) gələn push — sistem özü bildirişi göstərir.
@pragma('vm:entry-point')
Future<void> _fcmBackground(RemoteMessage message) async {}

// Push bildiriş — "new_products" topic-inə abunə. Firebase qurulmayıbsa tətbiq yenə işləyir.
Future<void> _initFcm() async {
  try {
    await Firebase.initializeApp();
    FirebaseMessaging.onBackgroundMessage(_fcmBackground);
    final m = FirebaseMessaging.instance;
    await m.requestPermission();
    await m.subscribeToTopic('new_products');
    // Tətbiq AÇIQ ikən push gələndə — ana səhifə banneri dərhal göstərsin (səslə).
    FirebaseMessaging.onMessage.listen((_) => foregroundPushTick.value++);
  } catch (_) {
    // Firebase konfiqurasiyası yoxdursa — səssizcə davam et
  }
}

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await AppConfig.load();
  await Store.I.load();
  await L.I.load();
  if (!kIsWeb) await _initFcm(); // FCM yalnız mobil; web-də Firebase konfiqi yoxdur
  SystemChrome.setSystemUIOverlayStyle(const SystemUiOverlayStyle(
    statusBarColor: Colors.transparent,
    statusBarIconBrightness: Brightness.dark,
    statusBarBrightness: Brightness.light,
  ));
  runApp(const LaptopsApp());
}

class LaptopsApp extends StatelessWidget {
  const LaptopsApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Laptops.az',
      debugShowCheckedModeBanner: false,
      theme: buildTheme(),
      darkTheme: buildTheme(), // sistem qaranlıq rejimdə olsa belə — həmişə açıq (ağ) tema
      themeMode: ThemeMode.light,
      home: const RootScaffold(),
    );
  }
}
