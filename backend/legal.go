package main

import (
	"net/http"
)

// GET /privacy — Google Play tələb etdiyi Məxfilik Siyasəti (Privacy Policy).
// Self-contained HTML; ayrıca dizayn/asılılıq yoxdur.
func privacyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(privacyHTML))
}

const privacyHTML = `<!doctype html>
<html lang="az">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Məxfilik Siyasəti — Laptops.az</title>
<style>
  :root{--ink:#14213A;--muted:#5A6474;--line:#E4E2DB;--bg:#FAFAF8}
  *{box-sizing:border-box}
  body{margin:0;background:var(--bg);color:var(--ink);
    font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;
    line-height:1.65;font-size:16px}
  .wrap{max-width:760px;margin:0 auto;padding:40px 22px 80px}
  header{display:flex;align-items:center;gap:12px;border-bottom:1px solid var(--line);
    padding-bottom:20px;margin-bottom:28px}
  header .badge{width:44px;height:44px;border-radius:11px;background:var(--ink);color:#fff;
    display:flex;align-items:center;justify-content:center;font-weight:800;font-size:18px}
  h1{font-size:26px;margin:0 0 4px;letter-spacing:-.4px}
  h2{font-size:19px;margin:32px 0 8px;letter-spacing:-.2px}
  p,li{color:#2b3550}
  .muted{color:var(--muted);font-size:14px;margin:0}
  ul{padding-left:20px}
  a{color:#1d4ed8}
  .card{background:#fff;border:1px solid var(--line);border-radius:14px;padding:18px 20px;margin-top:14px}
  footer{margin-top:44px;border-top:1px solid var(--line);padding-top:18px;color:var(--muted);font-size:13px}
</style>
</head>
<body>
<div class="wrap">
  <header>
    <div class="badge">L</div>
    <div>
      <h1>Məxfilik Siyasəti</h1>
      <p class="muted">Laptops.az mobil tətbiqi və vebsaytı · Son yenilənmə: 01.08.2026</p>
    </div>
  </header>

  <p>Bu Məxfilik Siyasəti «Laptops.az» (bundan sonra «biz», «tətbiq») mobil tətbiqindən və
  <a href="https://laptops.az">laptops.az</a> saytından istifadə edərkən hansı məlumatların
  toplandığını, necə istifadə olunduğunu və qorunduğunu izah edir.</p>

  <h2>1. Topladığımız məlumatlar</h2>
  <div class="card">
    <ul>
      <li><b>Sifariş məlumatları:</b> məhsul sifariş etdiyiniz zaman qeyd etdiyiniz <b>ad</b> və
        <b>telefon nömrəsi</b> (yalnız sifarişi emal etmək və sizinlə əlaqə saxlamaq üçün).</li>
      <li><b>Bildiriş identifikatoru:</b> yeni məhsul bildirişləri göndərə bilmək üçün cihazınızın
        Firebase Cloud Messaging (FCM) push tokeni.</li>
      <li><b>Texniki məlumat:</b> tətbiqin düzgün işləməsi üçün seçdiyiniz dil və əsas istifadə
        statistikası (anonim).</li>
    </ul>
    <p class="muted" style="margin-top:10px">Tətbiq bank kartı, ödəniş və ya digər maliyyə
    məlumatı TOPLAMIR — sifarişlər yalnız əlaqə əsaslıdır, ödəniş mağazada həyata keçirilir.</p>
  </div>

  <h2>2. Məlumatlardan istifadə</h2>
  <ul>
    <li>Sifarişlərinizi qəbul etmək və emal etmək;</li>
    <li>Yeni məhsul və endirimlər barədə push bildiriş göndərmək;</li>
    <li>Tətbiqin funksionallığını təmin etmək və yaxşılaşdırmaq.</li>
  </ul>

  <h2>3. Məlumatların üçüncü tərəflərlə paylaşılması</h2>
  <p>Şəxsi məlumatlarınızı satmırıq və reklam məqsədilə üçüncü tərəflərə vermirik. Yalnız aşağıdakı
  xidmət təminatçısından istifadə edirik:</p>
  <ul>
    <li><b>Google Firebase (Cloud Messaging):</b> push bildirişlərin çatdırılması üçün. Google-un
      məxfilik siyasəti: <a href="https://firebase.google.com/support/privacy">firebase.google.com/support/privacy</a>.</li>
  </ul>

  <h2>4. Məlumatların saxlanması və təhlükəsizliyi</h2>
  <p>Sifariş məlumatları yalnız xidmətin göstərilməsi üçün lazım olan müddət ərzində saxlanılır və
  müvafiq texniki tədbirlərlə qorunur. Bildiriş tokenini istənilən vaxt tətbiqi silməklə ləğv edə bilərsiniz.</p>

  <h2>5. Uşaqların məxfiliyi</h2>
  <p>Tətbiq 13 yaşından kiçik şəxslərə yönəlməyib və onlardan bilərəkdən şəxsi məlumat toplamırıq.</p>

  <h2>6. Hüquqlarınız</h2>
  <p>Sizinlə bağlı saxlanılan məlumatların silinməsini və ya düzəldilməsini tələb edə bilərsiniz.
  Bunun üçün bizimlə əlaqə saxlayın.</p>

  <h2>7. Əlaqə</h2>
  <div class="card">
    <p style="margin:0">Suallar üçün: <a href="mailto:info@laptops.az">info@laptops.az</a><br>
    Vebsayt: <a href="https://laptops.az">laptops.az</a></p>
  </div>

  <h2 style="margin-top:48px;padding-top:30px;border-top:1px solid var(--line)">Privacy Policy (English)</h2>
  <p>This Privacy Policy explains what data the “Laptops.az” mobile app and the
  <a href="https://laptops.az">laptops.az</a> website collect, how it is used and how it is protected.</p>

  <h2>1. Information we collect</h2>
  <div class="card">
    <ul>
      <li><b>Order details:</b> the <b>name</b> and <b>phone number</b> you enter when placing an order
        (used only to process the order and contact you).</li>
      <li><b>Notification identifier:</b> your device’s Firebase Cloud Messaging (FCM) push token, used
        to send new-product notifications.</li>
      <li><b>Technical data:</b> your chosen language and basic anonymous usage statistics.</li>
    </ul>
    <p class="muted" style="margin-top:10px">The app does NOT collect card, payment or other financial
    data — orders are contact-based and payment is completed in-store.</p>
  </div>

  <h2>2. How we use data</h2>
  <ul>
    <li>To receive and process your orders;</li>
    <li>To send push notifications about new products and discounts;</li>
    <li>To provide and improve the app’s functionality.</li>
  </ul>

  <h2>3. Sharing with third parties</h2>
  <p>We do not sell your personal data or share it for advertising. We use one service provider:
  <b>Google Firebase (Cloud Messaging)</b> to deliver push notifications
  (<a href="https://firebase.google.com/support/privacy">privacy policy</a>).</p>

  <h2>4. Retention &amp; security</h2>
  <p>Order data is kept only as long as needed to provide the service and is protected with appropriate
  technical measures. You can revoke the notification token anytime by deleting the app.</p>

  <h2>5. Children’s privacy</h2>
  <p>The app is not directed to children under 13 and we do not knowingly collect their personal data.</p>

  <h2>6. Your rights</h2>
  <p>You may request deletion or correction of data we hold about you by contacting us.</p>

  <h2>7. Contact</h2>
  <div class="card">
    <p style="margin:0">Questions: <a href="mailto:info@laptops.az">info@laptops.az</a><br>
    Website: <a href="https://laptops.az">laptops.az</a></p>
  </div>

  <footer>
    © 2009–2026 Laptops.az. Bu siyasət dəyişdirilə bilər; yenilənmələr bu səhifədə dərc olunur.
  </footer>
</div>
</body>
</html>`
