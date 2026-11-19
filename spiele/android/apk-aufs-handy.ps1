# Kopiert die Debug-APK ueber MTP (Datei-Modus) in den Download-Ordner des Handys.
# Aufruf:  powershell -NoProfile -ExecutionPolicy Bypass -File apk-aufs-handy.ps1

$ErrorActionPreference = "Stop"

$apk = Join-Path $PSScriptRoot "app\build\outputs\apk\debug\app-debug.apk"

if (-not (Test-Path $apk)) {
    Write-Host "APK fehlt: $apk"
    exit 1
}

$shell = New-Object -ComObject Shell.Application
$arbeitsplatz = $shell.NameSpace(17)   # "Dieser PC"

# Geraet suchen: Eintrag im Arbeitsplatz, der nicht zu den Laufwerken gehoert
$geraet = $null
foreach ($eintrag in $arbeitsplatz.Items()) {
    if ($eintrag.IsFolder -and $eintrag.Path -notmatch "^[A-Z]:") {
        $geraet = $eintrag
        break
    }
}

if (-not $geraet) {
    Write-Host "Kein Handy im Datei-Modus gefunden."
    exit 2
}

Write-Host "Geraet: $($geraet.Name)"

# MTP-Geraete haben keinen Dateipfad - deshalb ueber GetFolder weiterhangeln.
$wurzel = $geraet.GetFolder

$speicher = $null
foreach ($eintrag in $wurzel.Items()) {
    Write-Host "  Speicher-Kandidat: $($eintrag.Name)"
    if (-not $speicher -and $eintrag.Name -match "intern|speicher|phone|storage") {
        $speicher = $eintrag
    }
}

if (-not $speicher) {
    Write-Host "Kein interner Speicher sichtbar."
    exit 3
}

$inhalt = $speicher.GetFolder

$ziel = $null
foreach ($eintrag in $inhalt.Items()) {
    if ($eintrag.Name -match "^Download$") {
        $ziel = $eintrag
        break
    }
}

if (-not $ziel) {
    Write-Host "Kein Download-Ordner auf dem Handy gefunden."
    exit 4
}

Write-Host "Ziel: $($speicher.Name)\$($ziel.Name)"
$zielOrdner = $ziel.GetFolder

# 16 = "Ja zu allen" bei Rueckfragen, 4 = keine Fortschrittsanzeige
$zielOrdner.CopyHere($apk, 16)

# Kurz warten und pruefen, ob die Datei angekommen ist
for ($i = 0; $i -lt 20; $i++) {
    Start-Sleep -Milliseconds 500
    foreach ($eintrag in $zielOrdner.Items()) {
        if ($eintrag.Name -eq "app-debug.apk") {
            Write-Host "Kopiert: $($speicher.Name)\$($ziel.Name)\app-debug.apk ($([math]::Round($eintrag.Size / 1MB, 2)) MB)"
            exit 0
        }
    }
}

Write-Host "Kopie nicht bestaetigt - bitte im Handy-Download-Ordner pruefen."
exit 5
