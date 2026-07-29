import 'package:flutter/foundation.dart';

/// Tətbiq ön plandaykən (açıq) yeni məhsul push-u gəldikdə artırılır.
/// Ana səhifə bunu dinləyir və içəri bildiriş bannerini dərhal göstərir.
final ValueNotifier<int> foregroundPushTick = ValueNotifier<int>(0);
