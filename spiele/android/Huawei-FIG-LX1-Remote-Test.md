# Huawei FIG-LX1 – Testbericht zur TrashyGecko-Fernbedienung

Stand: 20.11.2026. Die Tests beschränkten sich auf die vorhandene TrashyGecko-App und die geforderten Fernbedienungsfunktionen. Es wurden keine persönlichen Dateien oder sonstigen Apps geöffnet oder verändert.

## Gerät

- Hersteller/Modell: Huawei FIG-LX1
- Android: 9, API 28
- ABIs: `arm64-v8a`, `armeabi-v7a`, `armeabi`
- Auflösung: physisch 1080×2160; bereits gesetzter Android-Override 720×1440 bei Dichte 320 dpi
- System-WebView: Paket `com.google.android.webview`, Version 138.0.7204.179 vorhanden
- WLAN: eingeschaltet
- Zustand: beim ersten Read-Check zunächst gesperrt/asleep. Später ADB meldete das Gerät vorübergehend als awake/unlocked; danach wurde es wieder gesperrt. Es wurde nicht aktiv entsperrt.

## APK-Installation

- Build: `:app:assembleDebug` erfolgreich; eine einzelne Debug-APK `app-debug.apk` (kein ABI-Split).
- Version: `1.10`, `versionCode 11` – damit wurde getestet. Der aktuelle Stand ist `1.12` / `versionCode 13`; dieser Gerätetest ist also zwei Versionen alt. `minSdk 21`, `targetSdk 34`.
- Update: mehrfach mit `adb install -r` über die vorhandene `de.trashygecko.tmahjong`-Installation installiert (die damalige Debug-APK hatte 4.922.437 B, die heutige 12.315.563 B); ADB meldete `Success`. Das vorherige `firstInstallTime` blieb beim ersten Update erhalten, `lastUpdateTime` wurde aktualisiert.
- Kein Deinstallieren, keine Werkseinstellungen, kein Löschen von App-Daten.
- Debug-Signierung wurde für den In-place-Updatepfad akzeptiert. Ein Release-APK mit dauerhaftem Release-Schlüssel wurde nicht erstellt oder geprüft.

## Laufzeit- und UI-Prüfung

- Die reguläre `ConnectorActivity`-Startseite lief. Die Fernbedienungs-Kachel wurde über die normale UI aufgerufen.
- `RemoteControlActivity` wurde als Vordergrund-Activity bestätigt.
- Logcat meldete `Fernbedienungsoberfläche geladen`; im gefilterten `AndroidRuntime:E`-/`TrashyRemote`-Log war kein Crash sichtbar.
- Der Suchbutton wurde angetippt; die Remote-Activity blieb im Vordergrund und es wurde kein Absturz protokolliert. Es gab keinen Fernseher im Testnetz, daher keinen bestätigten Fund und keine echte Discovery-/Pairing-/Steuerverbindung.
- Ein direkter ADB-Start der nicht exportierten `RemoteControlActivity` wurde vom Android-Sicherheitsmodell abgelehnt; der normale Start über die Startbildschirm-Kachel funktionierte.
- Hoch-/Querformat: die logische Displaygröße wurde temporär von 720×1440 auf 1440×720 umgestellt. Android meldete App-Fläche 1355×720, die Remote-Activity blieb aktiv. Danach wurden die ursprünglichen Display- und Rotationswerte wiederhergestellt. Die physische Sensorrotation wurde nicht getestet.
- Ein späterer Close/Reopen-Versuch traf auf das wieder gesperrte Gerät und wurde nicht fortgesetzt. Kein eigenständiges Entsperren.

## Automatisierte Prüfungen

- `:app:remoteProtocolTest`: bestanden, 47 Assertions.
- `:app:assembleDebug`: erfolgreich.
- Alle `*.js` im Fernbedienungsprojekt: `node --check` erfolgreich.
- Assertions decken Parser-/Mappinglogik ab; sie simulieren keine TV-Hardware, TLS-PIN-Kopplung, reale Discoveryantwort, tatsächliche Taste oder Multiplayer.

## Offen

Echter TV-/Streaminggerätetest, Android-TV-Pairing, ausgehandelte Fähigkeiten und Tasten, physische Sensorrotation, Neustart/erneutes Öffnen im entsperrten Zustand sowie Test auf weiteren Android-/ABI-/WebView-Versionen stehen aus. Ein zweites Multiplayergerät war nicht verfügbar.
