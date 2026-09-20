import 'dart:async' show unawaited;
import 'dart:io' show Platform;

import 'package:flutter/foundation.dart' show debugPrint, kIsWeb;
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
    PushDiag.firebase = 'OK';
    FirebaseMessaging.onBackgroundMessage(_fcmBackground);
    final m = FirebaseMessaging.instance;
    final settings = await m.requestPermission();
    PushDiag.permission = settings.authorizationStatus.name;
    // Tətbiq AÇIQ ikən push gələndə — ana səhifə banneri dərhal göstərsin (səslə).
    FirebaseMessaging.onMessage.listen((_) => foregroundPushTick.value++);
    // Abunəlik arxa fonda, təkrar cəhdlə — tətbiqin açılışını gözlətmir.
    unawaited(_subscribeWithRetry(m));
  } catch (e) {
    // Firebase qurulmayıbsa tətbiq yenə işləyir; səbəbi diaqnostikada və logda görünür.
    PushDiag.firebase = 'XƏTA';
    PushDiag.lastError = '$e';
    debugPrint('FCM işə düşmədi: $e');
  }
}

String _tokTail(String t) => t.length <= 8 ? t : '…${t.substring(t.length - 8)}';

// iOS-da APNs token bəzən bir neçə saniyə gecikir; token gəlmədən subscribeToTopic
// «apns-token-not-set» verir və abunəlik heç vaxt yaranmır. Ona görə ~2 dəqiqə ərzində
// təkrar cəhd edirik. Hər mərhələnin nəticəsi PushDiag-a yazılır.
Future<void> _subscribeWithRetry(FirebaseMessaging m) async {
  for (var attempt = 0; attempt < 24; attempt++) {
    try {
      if (Platform.isIOS) {
        final apns = await m.getAPNSToken();
        PushDiag.apns = apns == null ? 'yoxdur' : 'var (${_tokTail(apns)})';
        if (apns == null) {
          await Future.delayed(const Duration(seconds: 5));
          continue;
        }
      } else {
        PushDiag.apns = 'lazım deyil (Android)';
      }
      final token = await m.getToken();
      PushDiag.fcmFull = token ?? '';
      PushDiag.fcm = token == null ? 'yoxdur' : 'var (${_tokTail(token)})';
      await m.subscribeToTopic('new_products');
      PushDiag.topic = 'abunə olundu';
      PushDiag.lastError = '';
      return;
    } catch (e) {
      PushDiag.topic = 'alınmadı';
      PushDiag.lastError = '$e';
      debugPrint('FCM abunəlik cəhdi ${attempt + 1}: $e');
      await Future.delayed(const Duration(seconds: 5));
    }
  }
}

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await AppConfig.load();
  await Store.I.load();
  await L.I.load();
  // FCM push — həm Android, həm iOS. (iOS üçün APNs Auth Key Firebase-ə yüklənib,
  // Push capability + entitlements əlavə olunub.)
  if (!kIsWeb) await _initFcm();
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
