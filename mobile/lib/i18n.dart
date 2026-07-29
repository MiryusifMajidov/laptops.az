import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import 'config.dart';

/// Sadə çoxdilli sistem:
///  t(key)  → UI mətnləri (lokal, aşağıdakı _s xəritəsi)
///  tt(az)  → kataloq terminləri (kateqoriya/xüsusiyyət/dəyər) — backend /terms-dən
/// Dil dəyişəndə tətbiq yenidən qurulur (məhsullar ?lang= ilə yenidən yüklənir).

class LangOpt {
  final String code, name, icon;
  LangOpt(this.code, this.name, this.icon);
}

class L extends ChangeNotifier {
  static final L I = L._();
  L._();

  static const _kLang = 'app_lang';
  String lang = 'az';
  List<LangOpt> langs = [];
  Map<String, String> _terms = {};

  Future<void> load() async {
    final p = await SharedPreferences.getInstance();
    lang = p.getString(_kLang) ?? 'az';
    await fetchRemote();
  }

  Future<void> fetchRemote() async {
    try {
      final r = await http.get(Uri.parse('${AppConfig.apiBase}/languages')).timeout(const Duration(seconds: 10));
      if (r.statusCode == 200) {
        final list = jsonDecode(utf8.decode(r.bodyBytes)) as List;
        langs = list
            .map((e) => LangOpt((e['code'] ?? '').toString(), (e['name'] ?? '').toString(), (e['icon'] ?? '').toString()))
            .toList();
        if (langs.isNotEmpty && !langs.any((l) => l.code == lang)) lang = langs.first.code;
      }
    } catch (_) {}
    try {
      final r = await http.get(Uri.parse('${AppConfig.apiBase}/terms?lang=$lang')).timeout(const Duration(seconds: 10));
      if (r.statusCode == 200) {
        final m = jsonDecode(utf8.decode(r.bodyBytes)) as Map<String, dynamic>;
        _terms = m.map((k, v) => MapEntry(k, v.toString()));
      }
    } catch (_) {}
  }

  Future<void> setLang(String code) async {
    lang = code;
    final p = await SharedPreferences.getInstance();
    await p.setString(_kLang, code);
    await fetchRemote();
    notifyListeners();
  }

  String t(String key) {
    final m = _s[key];
    if (m == null) return key;
    return m[lang] ?? m['az'] ?? key;
  }

  String tt(String azText) => azText.isEmpty ? azText : (_terms[azText] ?? azText);
}

// qısa köməkçilər
String t(String key) => L.I.t(key);
String tt(String s) => L.I.tt(s);

// UI mətnləri: açar → {az, ru, tr, en}
const Map<String, Map<String, String>> _s = {
  // aşağı naviqasiya
  'nav.home': {'az': 'Ana', 'ru': 'Главная', 'tr': 'Ana', 'en': 'Home'},
  'nav.category': {'az': 'Kateqoriya', 'ru': 'Категории', 'tr': 'Kategori', 'en': 'Category'},
  'nav.chat': {'az': 'Söhbət', 'ru': 'Чат', 'tr': 'Sohbet', 'en': 'Chat'},
  'nav.search': {'az': 'Axtarış', 'ru': 'Поиск', 'tr': 'Arama', 'en': 'Search'},
  'nav.orders': {'az': 'Sifariş', 'ru': 'Заказы', 'tr': 'Sipariş', 'en': 'Orders'},
  // ümumi
  'common.available': {'az': 'Mövcuddur', 'ru': 'В наличии', 'tr': 'Mevcut', 'en': 'In stock'},
  'common.order': {'az': 'Sifariş et', 'ru': 'Заказать', 'tr': 'Sipariş ver', 'en': 'Order'},
  'common.filter': {'az': 'Filtr', 'ru': 'Фильтр', 'tr': 'Filtre', 'en': 'Filter'},
  'common.reset': {'az': 'Sıfırla', 'ru': 'Сбросить', 'tr': 'Sıfırla', 'en': 'Reset'},
  'common.clear': {'az': 'Təmizlə', 'ru': 'Очистить', 'tr': 'Temizle', 'en': 'Clear'},
  'common.retry': {'az': 'Yenidən', 'ru': 'Повторить', 'tr': 'Tekrar', 'en': 'Retry'},
  'common.close': {'az': 'Bağla', 'ru': 'Закрыть', 'tr': 'Kapat', 'en': 'Close'},
  'common.today': {'az': 'Bu gün', 'ru': 'Сегодня', 'tr': 'Bugün', 'en': 'Today'},
  // ana səhifə
  'home.searchHint': {'az': 'Notebook, marka, model axtar…', 'ru': 'Поиск: ноутбук, бренд, модель…', 'tr': 'Notebook, marka, model ara…', 'en': 'Search laptop, brand, model…'},
  'home.now': {'az': 'indi', 'ru': 'сейчас', 'tr': 'şimdi', 'en': 'now'},
  'home.newArrived': {'az': 'Yeni gəldi: {p} — stokda!', 'ru': 'Новинка: {p} — в наличии!', 'tr': 'Yeni geldi: {p} — stokta!', 'en': 'New arrival: {p} — in stock!'},
  'home.featured': {'az': 'Seçilmiş', 'ru': 'Избранное', 'tr': 'Öne çıkan', 'en': 'Featured'},
  'home.popular': {'az': 'Populyar', 'ru': 'Популярное', 'tr': 'Popüler', 'en': 'Popular'},
  'home.empty': {'az': 'Hazırda saytda məhsul yoxdur', 'ru': 'Пока нет товаров', 'tr': 'Şu an ürün yok', 'en': 'No products yet'},
  // kateqoriyalar
  'cat.title': {'az': 'Kateqoriyalar', 'ru': 'Категории', 'tr': 'Kategoriler', 'en': 'Categories'},
  'cat.count': {'az': '{n} məhsul', 'ru': '{n} товаров', 'tr': '{n} ürün', 'en': '{n} products'},
  'cat.empty': {'az': 'Kateqoriya yoxdur', 'ru': 'Нет категорий', 'tr': 'Kategori yok', 'en': 'No categories'},
  'cat.all': {'az': 'Hamısı', 'ru': 'Все', 'tr': 'Tümü', 'en': 'All'},
  'cat.noProducts': {'az': 'Bu kateqoriyada məhsul yoxdur', 'ru': 'В этой категории нет товаров', 'tr': 'Bu kategoride ürün yok', 'en': 'No products in this category'},
  'cat.noMatch': {'az': 'Bu filtrlərə uyğun məhsul tapılmadı', 'ru': 'Нет товаров по этим фильтрам', 'tr': 'Bu filtrelere uygun ürün yok', 'en': 'No products match these filters'},
  // axtarış
  'search.title': {'az': 'Axtarış', 'ru': 'Поиск', 'tr': 'Arama', 'en': 'Search'},
  'search.hint': {'az': 'Notebook, marka, model…', 'ru': 'Ноутбук, бренд, модель…', 'tr': 'Notebook, marka, model…', 'en': 'Laptop, brand, model…'},
  'search.noResult': {'az': '«{q}» üçün nəticə tapılmadı', 'ru': 'По «{q}» ничего не найдено', 'tr': '«{q}» için sonuç yok', 'en': 'No results for «{q}»'},
  'search.results': {'az': 'NƏTİCƏLƏR · {n}', 'ru': 'РЕЗУЛЬТАТЫ · {n}', 'tr': 'SONUÇLAR · {n}', 'en': 'RESULTS · {n}'},
  'search.prompt': {'az': 'Məhsul adı və ya markanı yazıb axtar', 'ru': 'Введите название или бренд', 'tr': 'Ürün adı veya marka yazın', 'en': 'Type a product name or brand'},
  'search.recent': {'az': 'SON AXTARIŞLAR', 'ru': 'НЕДАВНИЕ ПОИСКИ', 'tr': 'SON ARAMALAR', 'en': 'RECENT SEARCHES'},
  // sifarişlər
  'orders.title': {'az': 'Sifarişlərim', 'ru': 'Мои заказы', 'tr': 'Siparişlerim', 'en': 'My orders'},
  'orders.empty': {'az': 'Hələ sifarişin yoxdur.\nMəhsul seçib «Sifariş et» düyməsinə bas.', 'ru': 'Пока нет заказов.\nВыберите товар и нажмите «Заказать».', 'tr': 'Henüz siparişiniz yok.\nÜrün seçip «Sipariş ver»e basın.', 'en': 'No orders yet.\nPick a product and tap «Order».'},
  'orders.pending': {'az': 'Gözləyir', 'ru': 'Ожидает', 'tr': 'Bekliyor', 'en': 'Pending'},
  // məhsul
  'product.serial': {'az': 'Seriya: {s}', 'ru': 'Серийный: {s}', 'tr': 'Seri: {s}', 'en': 'Serial: {s}'},
  'product.payNote': {'az': 'Ödəniş yoxdur — mağazada Nağd/Kart/Taksit', 'ru': 'Без предоплаты — оплата в магазине: Наличные/Карта/Рассрочка', 'tr': 'Ödeme yok — mağazada Nakit/Kart/Taksit', 'en': 'No prepay — pay in store: Cash/Card/Installment'},
  // sifariş forması
  'order.title': {'az': 'Sifariş', 'ru': 'Заказ', 'tr': 'Sipariş', 'en': 'Order'},
  'order.qty': {'az': '1 ədəd', 'ru': '1 шт.', 'tr': '1 adet', 'en': '1 pc'},
  'order.nameLabel': {'az': 'AD, SOYAD', 'ru': 'ИМЯ, ФАМИЛИЯ', 'tr': 'AD, SOYAD', 'en': 'FULL NAME'},
  'order.namePlaceholder': {'az': 'məs. Elvin Məmmədov', 'ru': 'напр. Иван Иванов', 'tr': 'örn. Ali Yılmaz', 'en': 'e.g. John Smith'},
  'order.phoneLabel': {'az': 'TELEFON', 'ru': 'ТЕЛЕФОН', 'tr': 'TELEFON', 'en': 'PHONE'},
  'order.noteLabel': {'az': 'QEYD (İSTƏYƏ BAĞLI)', 'ru': 'ЗАМЕТКА (НЕОБЯЗАТЕЛЬНО)', 'tr': 'NOT (İSTEĞE BAĞLI)', 'en': 'NOTE (OPTIONAL)'},
  'order.notePlaceholder': {'az': 'Nə vaxt gələ bilərəm…', 'ru': 'Когда смогу зайти…', 'tr': 'Ne zaman gelebilirim…', 'en': 'When I can come by…'},
  'order.needName': {'az': 'Zəhmət olmasa ad, soyad yaz', 'ru': 'Пожалуйста, укажите имя и фамилию', 'tr': 'Lütfen ad soyad yazın', 'en': 'Please enter your full name'},
  'order.needPhone': {'az': 'Zəhmət olmasa telefon nömrəsi yaz', 'ru': 'Пожалуйста, укажите телефон', 'tr': 'Lütfen telefon numarası yazın', 'en': 'Please enter your phone number'},
  'order.received': {'az': 'Sifariş qəbul edildi', 'ru': 'Заказ принят', 'tr': 'Siparişiniz alındı', 'en': 'Order received'},
  'order.receivedText': {'az': 'Nömrə: {ref}\nMəhsul 24 saat saxlanılır — mağaza sizə zəng edəcək.', 'ru': 'Номер: {ref}\nТовар удерживается 24 часа — магазин вам позвонит.', 'tr': 'Numara: {ref}\nÜrün 24 saat saklanır — mağaza sizi arayacak.', 'en': 'Ref: {ref}\nThe item is held for 24h — the store will call you.'},
  'order.payNote': {'az': 'Ödəniş yoxdur. Məhsul 24 saat saxlanılır, mağaza sizə zəng edir.', 'ru': 'Без предоплаты. Товар удерживается 24 часа, магазин вам позвонит.', 'tr': 'Ödeme yok. Ürün 24 saat saklanır, mağaza sizi arar.', 'en': 'No payment. The item is held 24h, the store will call you.'},
  'order.submit': {'az': 'Sifarişi təsdiqlə', 'ru': 'Подтвердить заказ', 'tr': 'Siparişi onayla', 'en': 'Confirm order'},
  // AI
  'ai.title': {'az': 'AI Köməkçi', 'ru': 'AI-помощник', 'tr': 'AI Asistan', 'en': 'AI Assistant'},
  'ai.status': {'az': 'Onlayn · dərhal cavablayır', 'ru': 'Онлайн · отвечает сразу', 'tr': 'Çevrimiçi · anında yanıt', 'en': 'Online · instant replies'},
  'ai.welcome': {'az': 'Salam! 👋 Sizə uyğun notebook tapmaqda kömək edim. Nə üçün istifadə edəcəksiniz?', 'ru': 'Здравствуйте! 👋 Помогу подобрать ноутбук. Для чего будете использовать?', 'tr': 'Merhaba! 👋 Size uygun notebook bulmanıza yardım edeyim. Ne için kullanacaksınız?', 'en': 'Hi! 👋 Let me help you find the right laptop. What will you use it for?'},
  'ai.inputHint': {'az': 'Sualınızı yazın…', 'ru': 'Напишите ваш вопрос…', 'tr': 'Sorunuzu yazın…', 'en': 'Type your question…'},
  'ai.error': {'az': 'Bağışlayın, bir problem oldu. Bir azdan yenidən yoxlayın 🙏', 'ru': 'Извините, возникла проблема. Попробуйте чуть позже 🙏', 'tr': 'Üzgünüz, bir sorun oluştu. Biraz sonra tekrar deneyin 🙏', 'en': 'Sorry, something went wrong. Please try again in a moment 🙏'},
  'ai.qGameL': {'az': '🎮 Oyun', 'ru': '🎮 Игры', 'tr': '🎮 Oyun', 'en': '🎮 Gaming'},
  'ai.qGameQ': {'az': 'Oyun üçün notebook axtarıram, tövsiyə edin', 'ru': 'Ищу ноутбук для игр, посоветуйте', 'tr': 'Oyun için notebook arıyorum, önerir misiniz', 'en': "I'm looking for a gaming laptop, please recommend"},
  'ai.qWorkL': {'az': '💼 İş', 'ru': '💼 Работа', 'tr': '💼 İş', 'en': '💼 Work'},
  'ai.qWorkQ': {'az': 'İş və ofis işləri üçün notebook lazımdır', 'ru': 'Нужен ноутбук для работы и офиса', 'tr': 'İş ve ofis için notebook lazım', 'en': 'I need a laptop for work and office'},
  'ai.qDesignL': {'az': '🎨 Dizayn', 'ru': '🎨 Дизайн', 'tr': '🎨 Tasarım', 'en': '🎨 Design'},
  'ai.qDesignQ': {'az': 'Dizayn və video montaj üçün güclü notebook', 'ru': 'Мощный ноутбук для дизайна и видеомонтажа', 'tr': 'Tasarım ve video montaj için güçlü notebook', 'en': 'A powerful laptop for design and video editing'},
  'ai.qStudentL': {'az': '🎓 Tələbə', 'ru': '🎓 Студент', 'tr': '🎓 Öğrenci', 'en': '🎓 Student'},
  'ai.qStudentQ': {'az': 'Tələbə üçün büdcəyə uyğun notebook', 'ru': 'Ноутбук для студента по доступной цене', 'tr': 'Öğrenci için bütçeye uygun notebook', 'en': 'A budget-friendly laptop for a student'},
  // filtr
  'filter.price': {'az': 'Qiymət', 'ru': 'Цена', 'tr': 'Fiyat', 'en': 'Price'},
  'filter.showResults': {'az': '{n} nəticəni göstər', 'ru': 'Показать {n}', 'tr': '{n} sonucu göster', 'en': 'Show {n} results'},
  // bildirişlər
  'notif.title': {'az': 'Bildirişlər', 'ru': 'Уведомления', 'tr': 'Bildirimler', 'en': 'Notifications'},
  'notif.newTag': {'az': 'YENİ GƏLDİ', 'ru': 'НОВИНКА', 'tr': 'YENİ GELDİ', 'en': 'NEW ARRIVAL'},
  'notif.empty': {'az': 'Hələ yeni məhsul yoxdur', 'ru': 'Пока нет новинок', 'tr': 'Henüz yeni ürün yok', 'en': 'No new products yet'},
  // profil / tərəfdaşlıq
  'profile.title': {'az': 'Tərəfdaşlıq', 'ru': 'Партнёрство', 'tr': 'Ortaklık', 'en': 'Partnership'},
  'profile.partner': {'az': 'Partner', 'ru': 'Партнёр', 'tr': 'Partner', 'en': 'Partner'},
  'profile.heading': {'az': 'Bizimlə tərəfdaş olun', 'ru': 'Станьте нашим партнёром', 'tr': 'Bizimle ortak olun', 'en': 'Become our partner'},
  'profile.marketing': {'az': 'Mağazanızsa — bizimlə tərəfdaşlıq edərək kompüterləri və aksesuarları optavoy (topdan) qiymətə ala bilərsiniz.', 'ru': 'Если у вас магазин — станьте партнёром и покупайте компьютеры и аксессуары по оптовым ценам.', 'tr': 'Mağazanız varsa — bizimle ortak olarak bilgisayarları ve aksesuarları toptan fiyata alabilirsiniz.', 'en': 'If you run a store — partner with us and buy computers and accessories at wholesale prices.'},
  'profile.login': {'az': 'Partner girişi', 'ru': 'Вход для партнёра', 'tr': 'Partner girişi', 'en': 'Partner login'},
  'profile.username': {'az': 'İstifadəçi adı', 'ru': 'Имя пользователя', 'tr': 'Kullanıcı adı', 'en': 'Username'},
  'profile.password': {'az': 'Parol', 'ru': 'Пароль', 'tr': 'Şifre', 'en': 'Password'},
  'profile.signin': {'az': 'Daxil ol', 'ru': 'Войти', 'tr': 'Giriş yap', 'en': 'Sign in'},
  'profile.checking': {'az': 'Yoxlanılır…', 'ru': 'Проверка…', 'tr': 'Kontrol ediliyor…', 'en': 'Checking…'},
  'profile.applyLink': {'az': 'Hesabınız yoxdur? Tərəfdaşlıq üçün müraciət edin →', 'ru': 'Нет аккаунта? Подайте заявку на партнёрство →', 'tr': 'Hesabınız yok mu? Ortaklık için başvurun →', 'en': "No account? Apply for partnership →"},
  'profile.applyTitle': {'az': 'Tərəfdaşlıq üçün müraciət', 'ru': 'Заявка на партнёрство', 'tr': 'Ortaklık başvurusu', 'en': 'Partnership application'},
  'profile.applyDesc': {'az': 'Məlumatlarınızı buraxın — biz sizinlə əlaqə saxlayıb hesab açacağıq.', 'ru': 'Оставьте данные — мы свяжемся и создадим аккаунт.', 'tr': 'Bilgilerinizi bırakın — sizinle iletişime geçip hesap açacağız.', 'en': "Leave your details — we'll contact you and set up an account."},
  'profile.name': {'az': 'Ad, soyad', 'ru': 'Имя, фамилия', 'tr': 'Ad, soyad', 'en': 'Full name'},
  'profile.store': {'az': 'Mağaza adı', 'ru': 'Название магазина', 'tr': 'Mağaza adı', 'en': 'Store name'},
  'profile.phone': {'az': 'Telefon', 'ru': 'Телефон', 'tr': 'Telefon', 'en': 'Phone'},
  'profile.applySend': {'az': 'Müraciəti göndər', 'ru': 'Отправить заявку', 'tr': 'Başvuruyu gönder', 'en': 'Send application'},
  'profile.sending': {'az': 'Göndərilir…', 'ru': 'Отправка…', 'tr': 'Gönderiliyor…', 'en': 'Sending…'},
  'profile.backLogin': {'az': '← Girişə qayıt', 'ru': '← Назад ко входу', 'tr': '← Girişe dön', 'en': '← Back to login'},
  'profile.wholesaleActive': {'az': 'Optavoy qiymətlər aktivdir', 'ru': 'Оптовые цены активны', 'tr': 'Toptan fiyatlar aktif', 'en': 'Wholesale prices active'},
  'profile.wholesaleDesc': {'az': 'Bütün məhsullarda topdan (optavoy) qiymətləri görürsünüz.', 'ru': 'Вы видите оптовые цены на все товары.', 'tr': 'Tüm ürünlerde toptan fiyatları görüyorsunuz.', 'en': 'You see wholesale prices on all products.'},
  'profile.logout': {'az': 'Çıxış', 'ru': 'Выйти', 'tr': 'Çıkış', 'en': 'Log out'},
  'profile.needLogin': {'az': 'İstifadəçi adı və parol vacibdir', 'ru': 'Имя пользователя и пароль обязательны', 'tr': 'Kullanıcı adı ve şifre gerekli', 'en': 'Username and password are required'},
  'profile.notPartner': {'az': 'Bu hesab partner deyil', 'ru': 'Этот аккаунт не партнёрский', 'tr': 'Bu hesap partner değil', 'en': 'This account is not a partner'},
  'profile.needFields': {'az': 'Ad və telefon vacibdir', 'ru': 'Имя и телефон обязательны', 'tr': 'Ad ve telefon gerekli', 'en': 'Name and phone are required'},
  'profile.applied': {'az': 'Müraciətiniz göndərildi. Tezliklə sizinlə əlaqə saxlanılacaq.', 'ru': 'Заявка отправлена. Мы скоро свяжемся.', 'tr': 'Başvurunuz gönderildi. En kısa sürede iletişime geçeceğiz.', 'en': 'Your application was sent. We will contact you soon.'},
  // paylaş
  'share.title': {'az': 'Paylaş', 'ru': 'Поделиться', 'tr': 'Paylaş', 'en': 'Share'},
  'share.ai': {'az': 'AI Köməkçi', 'ru': 'AI-помощник', 'tr': 'AI Asistan', 'en': 'AI Assistant'},
  'share.sms': {'az': 'Nömrəyə', 'ru': 'СМС', 'tr': 'SMS', 'en': 'SMS'},
  'share.email': {'az': 'E-poçt', 'ru': 'E-mail', 'tr': 'E-posta', 'en': 'Email'},
  'share.copyLink': {'az': 'Linki kopyala', 'ru': 'Копировать', 'tr': 'Linki kopyala', 'en': 'Copy link'},
  'share.more': {'az': 'Digər', 'ru': 'Ещё', 'tr': 'Diğer', 'en': 'More'},
  'share.copied': {'az': 'Link kopyalandı', 'ru': 'Ссылка скопирована', 'tr': 'Bağlantı kopyalandı', 'en': 'Link copied'},
  'share.askHint': {'az': 'Sual ver: bu model necədir?', 'ru': 'Спросите: как эта модель?', 'tr': 'Sor: bu model nasıl?', 'en': 'Ask: how is this model?'},
  // dil seçici
  'lang.title': {'az': 'Dil', 'ru': 'Язык', 'tr': 'Dil', 'en': 'Language'},
  // paylaşım (əlavə) — mövcud share.* açarları yuxarıda 180-187-də var
  'share.whatsapp': {'az': 'WhatsApp', 'ru': 'WhatsApp', 'tr': 'WhatsApp', 'en': 'WhatsApp'},
  'ai.aboutProduct': {'az': 'Bu məhsul haqqında məlumat verə bilərsən? {p}', 'ru': 'Расскажи об этом товаре: {p}', 'tr': 'Bu ürün hakkında bilgi verir misin? {p}', 'en': 'Can you tell me about this product? {p}'},
};
