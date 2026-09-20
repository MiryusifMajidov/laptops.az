import 'package:flutter/foundation.dart';

/// Tətbiq ön plandaykən (açıq) yeni məhsul push-u gəldikdə artırılır.
/// Ana səhifə bunu dinləyir və içəri bildiriş bannerini dərhal göstərir.
final ValueNotifier<int> foregroundPushTick = ValueNotifier<int>(0);

/// Push quraşdırmasının vəziyyəti — bildiriş gəlmədikdə hansı mərhələdə
/// dayandığını cihazın özündə görmək üçün. Loqoya uzun basanda açılan
/// gizli dialoqda göstərilir (istifadəçiyə görünən ekranlarda YOXDUR).
class PushDiag {
  static String firebase = '—';
  static String permission = '—';
  static String apns = '—';
  static String fcm = '—';
  static String topic = '—';
  static String lastError = '';

  /// Tam FCM tokeni — Firebase konsolundan test bildirişi göndərmək üçün kopyalanır.
  static String fcmFull = '';

  static String get summary => [
        'Firebase: $firebase',
        'İcazə: $permission',
        'APNs token: $apns',
        'FCM token: $fcm',
        'Topic (new_products): $topic',
        if (lastError.isNotEmpty) 'Son xəta: $lastError',
      ].join('\n');
}
