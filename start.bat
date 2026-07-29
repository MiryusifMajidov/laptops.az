@echo off
title Laptops.az - Baslat
set "PATH=C:\Program Files\Go\bin;C:\Program Files\nodejs;%PATH%"

echo ==================================================
echo    Laptops.az - butun serverler baslayir...
echo ==================================================
echo.
echo Bu 3 qara pencere ACIQ qalmalidir (serverler orada islyir).
echo.

rem 1) Backend / API  (port 8080)
start "Laptops.az API (8080)" cmd /k "cd /d %~dp0backend && go run ."

rem Backend qalxsin deye bir az gozle
timeout /t 4 >nul

rem 2) Admin panel  (port 5173)
start "Laptops.az Admin (5173)" cmd /k "cd /d %~dp0frontend && npm run dev"

rem 3) Musteri sayti  (port 5174)
start "Laptops.az Sayt (5174)" cmd /k "cd /d %~dp0website && npm run dev"

rem Serverler hazir olsun deye gozle, sonra brauzeri ac
timeout /t 8 >nul
start "" "http://localhost:5173"
start "" "http://localhost:5174"

echo.
echo --------------------------------------------------
echo   Admin panel : http://localhost:5173   (admin / admin)
echo   Musteri sayt: http://localhost:5174
echo --------------------------------------------------
echo.
echo Dayandirmaq ucun: stop.bat  (ve ya 3 qara pencereni bagla)
echo.
pause
