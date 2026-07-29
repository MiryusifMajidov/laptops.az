@echo off
title Laptops.az - Dayandir
echo Laptops.az serverleri dayandirilir...

rem Adli pencereleri (ve icindeki go/node prosesini) bagla
taskkill /F /FI "WINDOWTITLE eq Laptops.az API*" /T >nul 2>&1
taskkill /F /FI "WINDOWTITLE eq Laptops.az Admin*" /T >nul 2>&1
taskkill /F /FI "WINDOWTITLE eq Laptops.az Sayt*" /T >nul 2>&1

rem Ehtiyat ucun backend exe-ni de bagla
taskkill /F /IM laptops-backend.exe >nul 2>&1

echo Dayandirildi.
timeout /t 2 >nul
