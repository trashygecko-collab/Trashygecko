# TrashyGecko – WLAN-Fernbedienung

Eine lokale Weboberfläche und ein isoliertes Android-Bridge-Modul in der bestehenden TrashyGecko-App. Netzwerkzugriffe laufen asynchron in Java; der WebView stellt nur UI und Gerätemetadaten bereit. Keine CDN- oder Cloud-Abhängigkeit.

## Start

Im Projektordner ausführen:

```powershell
python -m http.server 8000
```

Dann `http://localhost:8000` öffnen. Im normalen Browser sind automatische WLAN-Suche und Native-Adapter nicht verfügbar; die Seite sagt das offen und schaltet keine Schein-Tasten frei. **In der App ist die Fernbedienung ausschließlich über „Extras“ erreichbar** (Startbildschirm → Extras → WLAN-Fernbedienung); die Kachel im Spielekatalog und auf der Startseite wird beim APK-Build entfernt. Beim Bau wird dieses Verzeichnis nach `assets/www/fernbedienung` kopiert. Keine Android-Projektarchitektur wird dupliziert.

## Implementierungsstand

- Gerätemanagement: mehrere Geräte, Räume, Anzeigenamen, manueller Host/IP, Auswahl, lokaler Metadatenspeicher, Fähigkeiten-/Verbindungsstatus und begrenzter Wiederverbindungsversuch.
- Android-Bridge v2: asynchrone Aufrufe, Herkunftsprüfung der WebView-Assets, Eingabegrößenbegrenzung, sicherer AssetLoader-Pfad, keine Datei-/Content-Zugriffe aus dem WebView.
- Discovery: SSDP-Suche für Roku ECP und DIAL sowie mDNS/NSD für `_androidtvremote2._tcp.`. Roku/DIAL werden als Erkennung markiert; SSDP allein beweist keine Fernsteuerbarkeit.
- Android TV Remote v2: TLS 1.2 auf 6467 zum PIN-Pairing und 6466 für Fernsteuerung; HEX-PIN, Zertifikats-Fingerabdruck-Pinning nach erfolgreicher Kopplung, Clientidentität im Android Keystore, Status-/Capability-Aushandlung. Tasten und Texteingabe werden nur bei ausgehandelten Gerätefunktionen freigeschaltet. Voraussetzung: kompatibler Remote-v2-Dienst und aktivierte TV-Freigabe. Dieses Protokoll ist reverse-engineered und keine zugesicherte öffentliche Google-Drittanbieter-API; Modell-/Firmwaretests stehen aus.
- Unterstützte Profile für diesen Android-TV-Adapter: Android/Google TV, TCL Android/Google TV, Hisense Android/Google TV, Sony Android/Google TV und Philips Android/Google TV – ausschließlich, wenn der konkrete Fernseher Android TV Remote v2 anbietet.

## Bewusst nicht aktivierte Protokolle

| System | Stand und Grund |
| --- | --- |
| Roku / TCL Roku TV | SSDP-Erkennung ja, Steuerung nein. Roku ECP ist lokales unverschlüsseltes HTTP; die aktuelle Roku-Dokumentation beschränkt ECP-Befehle von Drittanbieter-Mobilplattformen. Kein Cleartext- oder HTTP-Fallback in der App. |
| Samsung Tizen | Kein generationsübergreifend verifiziertes lokales Drittanbieter-Fernbedienungsprotokoll. Modell-/Freigabeverfahren unterscheiden sich. |
| LG webOS | Kein aktivierter SSAP-WebSocket: verbreitete lokale Implementierungen sind undokumentiert/modellabhängig und nicht durch die offizielle API als allgemeine Drittanbieterfernbedienung zugesichert. |
| Sony BRAVIA (nicht Android TV) | Consumer-IP-Control hängt von Modell, Einstellung und Authentisierung ab; verbreitete Verfahren benötigen lokales HTTP/PSK. Nur Android-TV-Modelle mit Remote v2 werden über das gemeinsame Android-TV-Protokoll unterstützt. |
| Philips (nicht Android TV) | JointSPACE-Generationen unterscheiden sich in Ports, HTTP/TLS und Pairing. Kein generischer Adapter ohne modell- und zertifikatsspezifische Validierung. |
| Hisense VIDAA | Lokales Fernbedienungsprotokoll ist generation-/firmwareabhängig und nicht als einheitliche öffentliche Drittanbieter-API bestätigt. |
| Amazon Fire TV / Fire OS | ADB ist laut Amazon für Entwicklung/Debugging dokumentiert, erfordert aktiviertes Debugging und TV-Autorisierung; es wird nicht als allgemeine Fernbedienung aktiviert. Vega OS ebenfalls nicht unterstützt. |
| DIAL | Discovery und App-Start-Erkennung, keine allgemeine Remote-Steuerung. Die Bridge ruft keine unverschlüsselten DIAL-HTTP-Ressourcen ab. |

Ein manuell ausgewähltes, nicht unterstütztes Profil antwortet mit einer konkreten Grenze. Nicht bestätigte Tasten und App-Funktionen bleiben deaktiviert.

## Datenschutz und Netzwerkgrenzen

Im Browser-/WebView-Speicher liegen nur Geräte-Metadaten, keine PINs oder Schlüssel. Android-TV-Clientschlüssel bleiben im Android Keystore; der bestätigte TV-Zertifikats-Fingerabdruck wird in app-privaten Preferences gehalten. App-Deinstallation löscht diese Identität. Adressen werden vor TLS-Verbindung auf private/lokale IPv4-Bereiche begrenzt. Es gibt keine app-weite Cleartext-Freigabe und keine dynamische Network-Security-Ausnahme für wechselnde Roku-IP-Adressen.

Ab Android 17/API 37 ist für Apps mit entsprechendem Target die neue lokale Netzwerkberechtigung zu berücksichtigen. Der aktuelle APK-Target bleibt 34; Android 9/API 28 bleibt Mindestversion. Für Zielversion 37 muss die Runtime-Berechtigungsabfrage ergänzt werden, bevor die App dorthin angehoben wird.

## Test

Im vorhandenen Android-Projekt:

```powershell
gradle :app:remoteProtocolTest :app:assembleDebug
```

Die Selbsttests prüfen Protokollparser, SSDP-/DIAL-Erkennung, lokale Adressgrenzen, Android-TV-Tastenmapping und Frame-Codierung. Ein echter kompatibler Fernseher war beim Entwicklungslauf nicht vorhanden; Pairing, TLS-Gerätekompatibilität und reale Tastenausführung sind daher nicht als Hardware-erfolgreich bestätigt.

## Struktur

- `index.html`, `styles.css`: mobile Oberfläche
- `js/app.js`: Gerätemanagement und UI-/Bridge-Orchestrierung
- `js/bridge.js`: gemeinsamer Web-Bridge-Vertrag
- `js/brands.js`: Profile und Grenzen
- `js/protocols/`: browserseitige, sichere Adapterbasis
- Native Bridge: `spiele/android/app/src/main/java/de/trashygecko/tmahjong/remote/`

## Quellen

- [Android Network Security Configuration](https://developer.android.com/privacy-and-security/security-config)
- [Android Local Network Permission](https://developer.android.com/privacy-and-security/local-network-permission)
- [Android TV Pairing Protocol (AOSP-Protobuf-Schema)](https://android.googlesource.com/platform/external/google-tv-pairing-protocol/%2B/7bcf8af9df8c24687f27fc262764628c4d80ff51/proto/polo.proto)
- [Roku External Control API](https://developer.roku.com/dev/docs/external-control-api)
- [LG webOS TV API references](https://webostv.developer.lge.com/develop/references)
- [Samsung Smart View sender documentation](https://developer.samsung.com/smarttv/develop/extension-libraries/smart-view-sdk/sender-apps/javascript-sender-app.html)
- [Sony BRAVIA REST API (Pro displays; not a consumer-model guarantee)](https://pro-bravia.sony.net/remote-display-control/rest-api/structure/)
- [Amazon Fire TV ADB connection](https://developer.amazon.com/docs/fire-tv/connecting-adb-to-device.html)
## Herstellerkatalog und Plattformzuordnung

Hersteller, TV-Plattform und Steuerprotokoll werden als drei getrennte Metadaten gespeichert. Die UI zeigt sie unabhängig an. Hersteller und Modell werden nach einem erfolgreichen Android-TV-Remote-v2-Handshake, soweit vom TV gemeldet, übernommen. Das Herstellerfeld bietet Vorschläge für Xiaomi/Mi/Redmi, Skyworth/Coocaa, Changhong, Haier, KONKA, Huawei Vision, Hisense/Vidda, Metz, Grundig, Thomson, Telefunken, JVC, Toshiba, Sharp, Blaupunkt, Aiwa, RCA, Vestel sowie die üblichen internationalen Marken. Die Herstellerliste ist ein Eingabe-/Metadatenkatalog, keine Kompatibilitätszusage.

Der Nutzer wählt unabhängig davon das Betriebssystem/Steuerprotokoll. Xiaomi, Redmi, Coocaa, Changhong, Haier, KONKA, Vestel und andere können je nach Modell/Region Android TV/Google TV nutzen; diese laufen über denselben Android-TV-Remote-v2-Adapter und funktionieren nur, falls das Gerät genau diesen Dienst anbietet. Nach erfolgreichem Handshake werden gemeldeter Hersteller und Modell in die getrennten Felder übernommen. Die manuelle Wahl „Android TV / Google TV“ testet nur diesen TLS-Dienst und fällt bei fehlendem Dienst mit einer verständlichen Meldung aus.

Huawei Vision/HarmonyOS, VIDAA/Vidda, Titan OS, Whale TV OS und Coolita sind als Plattformen auswählbar, aber nicht steuerbar. HarmonyOS' offizielle Distributed-Remote-Beispiele verwenden Huawei-HarmonyOS-Systemkomponenten und Cross-Device-Berechtigungen; das ist keine allgemeine Android-LAN-API. Die Whale-TV-Unterlagen zu Remote-Tasten betreffen TV-App-Integration, nicht automatisch eine mobile Fernbedienung. Roku bleibt SSDP-only. Für Handelsmarken wie Metz, Grundig, Thomson, Telefunken, JVC, Toshiba, Sharp, Blaupunkt, Aiwa, RCA und Vestel muss die konkrete Plattform am Modell/Typenschild bzw. in den TV-Netzwerkeinstellungen gewählt werden, sofern Discovery sie nicht eindeutig meldet.

Automatische Plattformerkennung ist derzeit begrenzt auf Android TV Remote v2 via mDNS/NSD sowie Roku/DIAL via SSDP. Android-TV-Herstellerdetails werden erst nach erfolgreichem Protokoll-Handshake aus dem TV-Protokoll gelesen. Andere Geräte werden nicht anhand des Markennamens geraten. Allgemeines SSDP verrät bei vielen TVs weder die aktive Plattform noch ein sicheres Steuerprotokoll; dafür werden pro Plattform geprüfte Discovery-Merkmale und Tests benötigt.
## Oberfläche und Bedienung (21.11.2026)

Die Seite wurde für die App neu aufgebaut: Geräteleiste oben, die Fernbedienung direkt darunter, Verwaltung und WLAN-Suche nur noch eingeklappt. Gestaltung in den Farben der App (dunkelgrün `#0b241d`, Gold `#e8cd8e`), damit der Extras-Bereich nicht wie ein Fremdkörper wirkt.

- **Geräteleiste:** je Gerät eine Karte mit Statuspunkt (grün verbunden, gold verbindet, rot Fehler). Antippen verbindet; „＋ Gerät“ öffnet den Einrichtungsdialog.
- **Fernbedienung:** Steuerkreuz mit OK, Lautstärke (lauter/stumm/leiser), Kanal +/−, Standby, Home/Zurück/Menü/Einstellungen, Wiedergabe (zurückspulen, Play/Pause, vorspulen), Info, Guide, Quelle, Tastatur und ausklappbarer Ziffernblock.
- **Keine Schein-Tasten:** Jede Taste wird gegen die vom Gerät gemeldeten Fähigkeiten geprüft. Nicht gemeldete Tasten sind sichtbar gedimmt und erklären beim Antippen, warum sie nicht senden; verbundene Geräte schalten sie automatisch frei. Die frühere „Apps“-Taste ist entfallen, weil Android TV Remote v2 keine App-Liste liefert und die Taste damit immer tot gewesen wäre.
- **Halten wiederholt:** Steuerkreuz, Lautstärke und Kanal senden nach 460 ms Halten alle 160 ms weiter (wie eine echte Fernbedienung).
- **Tastatur am Rechner:** Pfeiltasten, Enter = OK, Escape/Backspace = Zurück, `h` Home, `m` Menü, `p` Play/Pause, `+`/`-` Lautstärke, Ziffern 0–9. Wird ignoriert, solange ein Eingabefeld oder Dialog offen ist.
- **Kopplung:** Der TV zeigt einen sechsstelligen HEX-Code; der Dialog nimmt nur `0–9A–F` an, prüft die Länge und schickt den Code nach dem Absenden aus dem Feld. Nach Erfolg verbindet die App selbst neu.
- **Quelle:** Der Dialog bietet nur Eingänge, die der Adapter wirklich kennt (HDMI 1–4, TV-Tuner). Der frühere Eintrag „AV“ ist weg, weil er im Adapter auf keinen Tastencode zeigt.
- **Fehlertexte:** Socket-Fehler werden im Klartext erklärt („Der Fernseher nimmt keine Verbindung an (Port 6467 geschlossen) …“) statt als Java-Exception.
- **Bildschirmgrößen:** geprüft bei 360 × 640, 390 × 844 und 844 × 390 (Querformat): kein waagerechtes Scrollen, Steuerkreuz 148–200 px, Tasten 45–63 px. Am Handy stehen Steuerkreuz und die zwei Seitenspalten nebeneinander; im Querformat rückt die Wiedergabezeile auf sieben Spalten.

### Was dabei tatsächlich geprüft wurde

- Echo der Kette WebView → Java-Bridge → Netzwerk: „Suchen“ im Emulator (Android 14) lieferte die echten mDNS/SSDP-Warnungen zurück, und ein Gerät mit der Adresse `10.0.2.2` erzeugte den echten Socket-Fehler `ECONNREFUSED` samt verständlichem Text in der Oberfläche.
- Geräteverwaltung: Anlegen, Speichern (WebView-Speicher und `SharedPreferences`), Auswählen, Umbenennen, Entfernen, Raumfilter.
- Zurück-Navigation: Fernbedienung → Extras → Start → App-Startseite.
- **Nicht geprüft:** eine echte Kopplung und Tastenübertragung an einem Android-TV/Google-TV-Gerät – im Haushalt steht kein passendes Gerät, und der Emulator kann keinen TV ersetzen. Ebenfalls ungeprüft: zweites Handy im selben WLAN, Fire-TV/Samsung/LG (bewusst ohne Adapter) und die Vibration (Berechtigung `VIBRATE` ist gesetzt, im Emulator nicht spürbar).
