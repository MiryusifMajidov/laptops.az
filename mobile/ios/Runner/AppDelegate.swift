import Flutter
import UIKit
import FirebaseMessaging

// Bu layihə UIScene həyat dövrünü (SceneDelegate) işlədir. Bu halda firebase_messaging
// plagininin APNs tokenini avtomatik tutması işləməyə bilir — nəticədə cihaz tokensiz
// qalır və push heç vaxt gəlmir. Ona görə qeydiyyatı özümüz başladır, tokeni birbaşa
// Firebase-ə veririk və nəticəni UserDefaults-a yazırıq (Flutter tərəfdə "apns_status"
// açarı kimi oxunur və gizli diaqnostika pəncərəsində görünür).
@main
@objc class AppDelegate: FlutterAppDelegate, FlutterImplicitEngineDelegate {
  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
    UserDefaults.standard.set("qeydiyyat gözlənilir", forKey: "flutter.apns_status")
    application.registerForRemoteNotifications()
    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }

  override func application(
    _ application: UIApplication,
    didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data
  ) {
    Messaging.messaging().apnsToken = deviceToken
    UserDefaults.standard.set("OK (\(deviceToken.count) bayt)", forKey: "flutter.apns_status")
    super.application(application, didRegisterForRemoteNotificationsWithDeviceToken: deviceToken)
  }

  override func application(
    _ application: UIApplication,
    didFailToRegisterForRemoteNotificationsWithError error: Error
  ) {
    UserDefaults.standard.set("XƏTA: \(error.localizedDescription)", forKey: "flutter.apns_status")
    NSLog("APNs qeydiyyatı alınmadı: \(error)")
    super.application(application, didFailToRegisterForRemoteNotificationsWithError: error)
  }

  func didInitializeImplicitFlutterEngine(_ engineBridge: FlutterImplicitEngineBridge) {
    GeneratedPluginRegistrant.register(with: engineBridge.pluginRegistry)
  }
}
