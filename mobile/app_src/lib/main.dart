import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'config.dart';
import 'store.dart';
import 'theme.dart';
import 'screens/root.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await AppConfig.load();
  await Store.I.load();
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
      home: const RootScaffold(),
    );
  }
}
