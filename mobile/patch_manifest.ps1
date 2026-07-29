# AndroidManifest.xml-i tetbiq ucun hazirlayir:
#  - INTERNET icazesi (release APK-da API-ye qosulmaq ucun VACIBDIR)
#  - usesCleartextTraffic (http:// lokal serverə icaze)
#  - tel: dialer gorunurluyu (zeng dumesi)
#  - tetbiq adi: Laptops.az
$ErrorActionPreference = 'Stop'
$m = Join-Path $PSScriptRoot 'android\app\src\main\AndroidManifest.xml'
if (-not (Test-Path $m)) {
    Write-Host 'XETA: AndroidManifest.xml tapilmadi. Once "flutter create ." isle.'
    exit 1
}
$t = [System.IO.File]::ReadAllText($m)

$inject = @'
<uses-permission android:name="android.permission.INTERNET"/>
    <queries>
        <intent>
            <action android:name="android.intent.action.VIEW" />
            <data android:scheme="tel" />
        </intent>
    </queries>
    <application
'@

if ($t -notmatch 'permission\.INTERNET') {
    $t = $t -replace '<application', $inject
}
if ($t -notmatch 'usesCleartextTraffic') {
    $t = $t -replace '<application', '<application android:usesCleartextTraffic="true"'
}
$t = $t -replace 'android:label="[^"]*"', 'android:label="Laptops.az"'

$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText($m, $t, $utf8NoBom)
Write-Host 'AndroidManifest.xml yenilendi: INTERNET + cleartext + tel + label=Laptops.az'

# gradle.properties: Kotlin incremental compilesini sondur.
# Layihe ile pub cache ferqli disklerdedirse (mes. layihe D:, cache C:),
# Kotlin incremental compiler "different roots" xetasi verir. Bu flag onu hell edir.
$gp = Join-Path $PSScriptRoot 'android\gradle.properties'
if (Test-Path $gp) {
    $g = [System.IO.File]::ReadAllText($gp)
    if ($g -notmatch 'kotlin\.incremental=false') {
        if (-not $g.EndsWith("`n")) { $g += "`r`n" }
        $g += "kotlin.incremental=false`r`n"
        [System.IO.File]::WriteAllText($gp, $g, $utf8NoBom)
        Write-Host 'gradle.properties yenilendi: kotlin.incremental=false'
    }
}
