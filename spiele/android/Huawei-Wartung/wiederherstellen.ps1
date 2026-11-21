# Aktiviert nur Pakete aus deaktivierte_apps.txt wieder.
# Voraussetzung: genau ein autorisiertes ADB-Geraet verbunden.
$ErrorActionPreference = 'Stop'
$adb = (Get-Command adb.exe -ErrorAction SilentlyContinue).Source
if (-not $adb) {
    $candidate = Join-Path $PSScriptRoot '..\..\..\..\arbeit\werkzeuge\t-spiele\sdk\platform-tools\adb.exe'
    if (Test-Path -LiteralPath $candidate) { $adb = (Resolve-Path -LiteralPath $candidate).Path }
}
if (-not $adb) { throw 'adb.exe nicht gefunden. Android platform-tools in PATH aufnehmen.' }
$devices = @(& $adb devices | Select-String '^\S+\s+device$' | ForEach-Object { ($_.Line -split '\s+')[0] })
if ($devices.Count -ne 1) { throw "Genau ein autorisiertes ADB-Geraet benoetigt; gefunden: $($devices.Count)." }
$serial = $devices[0]
$packages = @(Get-Content -LiteralPath (Join-Path $PSScriptRoot 'deaktivierte_apps.txt') | ForEach-Object { $_.Trim() } | Where-Object { $_ -and -not $_.StartsWith('#') })
if ($packages.Count -eq 0) { Write-Host 'Keine Pakete eingetragen; nichts zu aendern.'; exit 0 }
foreach ($package in $packages) {
    if ($package -notmatch '^[A-Za-z0-9_]+(\.[A-Za-z0-9_]+)+$') { throw "Ungueltiger Paketname: $package" }
    Write-Host "Aktiviere $package auf $serial ..."
    & $adb -s $serial shell pm enable --user 0 $package
    if ($LASTEXITCODE -ne 0) { throw "Aktivierung fehlgeschlagen: $package" }
}
Write-Host 'Wiederherstellung abgeschlossen.'
