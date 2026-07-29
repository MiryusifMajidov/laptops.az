@echo off
chcp 65001 >nul
setlocal
cd /d "%~dp0"

echo ==================================================
echo    Laptops.az - Mobil tetbiq QURULUSU
echo ==================================================
echo.

where flutter >nul 2>&1
if errorlevel 1 (
  echo [XETA] Flutter tapilmadi.
  echo Once Flutter SDK qurulmalidir - bax: OXU-MENI.txt
  echo.
  pause
  exit /b 1
)

echo == 1/4  Platforma qovluqlari yaradilir (android + ios) ==
call flutter create --project-name laptops_az --org az.laptops --platforms=android,ios .
if errorlevel 1 goto :err

echo.
echo == 2/4  Menbe kodu kopyalanir (app_src -> layihe) ==
xcopy /E /I /Y "app_src\lib" "lib" >nul
copy /Y "app_src\pubspec.yaml" "pubspec.yaml" >nul
xcopy /E /I /Y "app_src\assets" "assets" >nul
if exist "app_src\analysis_options.yaml" copy /Y "app_src\analysis_options.yaml" "analysis_options.yaml" >nul
if exist "test\widget_test.dart" del /Q "test\widget_test.dart"

echo.
echo == 3/4  AndroidManifest.xml yenilenir ==
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0patch_manifest.ps1"
if errorlevel 1 goto :err

echo.
echo == 4/4  Paketler yuklenir (flutter pub get) ==
call flutter pub get
if errorlevel 1 goto :err

echo.
echo ==================================================
echo    HAZIR!
echo ==================================================
echo.
echo VACIB: lib\config.dart faylinda server unvanini
echo        kompyuterinin IP-si ile evez et (mes. http://192.168.1.35:8080)
echo        ( cmd -^> ipconfig -^> IPv4 Address )
echo.
echo Sonra telefonu USB ile qos (USB debugging acik) ve:
echo    flutter run          =^> telefonda canli isletmek
echo    flutter build apk    =^> APK yaratmaq
echo    APK yolu: build\app\outputs\flutter-apk\app-release.apk
echo.
pause
exit /b 0

:err
echo.
echo [XETA] Emeliyyat ugursuz oldu. Yuxaridaki mesajlari yoxla.
echo.
pause
exit /b 1
