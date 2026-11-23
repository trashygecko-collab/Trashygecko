# Prüfbericht Android

Grundregel: Hier steht nur, was **tatsächlich ausgeführt** wurde. Was nicht lief, steht als
„nicht getestet“ – ein erfolgreicher Build ist **kein** Funktionsnachweis.

Letzte Aktualisierung: 23.11.2026 (Etappe 3, Schritt 7: native Startansicht, Zustandssicherung).

---

## 1. Umgebung

| | |
|---|---|
| Bauwerkzeuge | JDK 17.0.21+9, Android SDK (`platforms/android-34`, `build-tools/34.0.0`), Gradle 8.7 – alle unter `Documents\arbeit\werkzeuge\` |
| Gerät A (echt) | Huawei **FIG-LX1**, Android **9 / API 28**, `arm64-v8a`, per `adb` erreichbar |
| Gerät B | Emulator `sdk_gphone64_x86_64`, Android **14 / API 34**, `x86_64` |
| Installiert auf A | `de.trashygecko.tmahjong` **1.15 / versionCode 16**, Erstinstallation 20.11.2026, letztes Update 22.11.2026 23:22 |
| Signatur auf A | eine Signatur, `version:2` – die release-signierte Fassung (`CN=TrashyGecko`) |

Achtung für alle künftigen Tests: Eine **Debug**-APK lässt sich nicht über die
release-signierte 1.15 legen (`INSTALL_FAILED_UPDATE_INCOMPATIBLE`). Entweder mit
`keystore.properties` release bauen oder auf dem Emulator testen. Ein Deinstallieren auf
Gerät A löscht alle lokalen Spielstände.

---

## 1g. Etappe 3, Schritt 7 (23.11.2026, 18:29–18:34)

Geänderter Code: `MainActivity.java` (134 → **193** Zeilen: native Startansicht, Netzstatus),
`SpieleActivity.java` (619 → **630** Zeilen: wiederherstellen vor dem Zeichnen). Debug-Build
auf Emulator B (`adb install -r`), Gerät A unberührt; gelesen über `uiautomator` (Text,
`selected`, `bounds`), `logcat` und `pidof`.

| Prüfung | Ergebnis | Beleg |
|---|---|---|
| `:app:assembleDebug` | **erfolgreich**, `BUILD SUCCESSFUL in 10s`, exit 0 | Terminalausgabe |
| Startansicht nativ | „TRASHYGECKO“ + „WLAN · online bereit“ `[730,197][1030,241]`, Titel `[50,241][1030,394]`, vier Kacheln (y 553/864/1175/1486) mit Beschreibung, Fußzeile endet bei y 2018 (Bildschirm 2340) | `uiautomator` |
| Kachel „Extras“ | Connector-Seite mit Überschrift „Extras“ | `uiautomator` |
| Kachel „Lokales Netz“ | Connector-Seite „Lokales Netzwerk“ | `uiautomator` |
| Zurück von beiden | native Startansicht unverändert (Titel, Status, vier Kacheln) | `uiautomator` |
| **Chip ohne Suchtext** vorher | „Mehrspieler“ `selected="true"`, „70 von 94“, Scroll 5237 | `uiautomator`, `logcat` |
| Prozess wirklich beendet | PID 9573 → `pidof` leer | `adb` |
| Neuer Prozess | PID **9782**, `Zustand wiederhergestellt: „“, 1 Merkmale, Scroll 5237`, `Liste: 70 Zeilen in 103 ms` | `logcat` |
| Liste nachher | Auszug vor/nach **identisch** (`diff` leer), „70 von 94“, Chip `selected="true"` | `uiautomator`, zwei Dumps |
| Suchtext-Wiederherstellung | „a“ → „55 von 94“; nach `am kill` PID **10044**, „55 von 94“, Feld „a“, `Liste: 55 Zeilen in 93 ms` | `uiautomator`, `logcat` |
| Tippen danach | „b“ → „5 von 94“, `Liste: 5 Zeilen in 19 ms` | `uiautomator`, `logcat` |
| „Zurücksetzen“ | „94 Spiele“, `Liste: 94 Zeilen in 60 ms` | `uiautomator`, `logcat` |
| Abstürze | **0** `FATAL EXCEPTION` | `logcat` |

Ursache des vorher falschen Falls (aus dem Code abgeleitet, **nicht** durch einen Lauf mit
der alten Fassung belegt): `onCreate` rief `zeichne()` vor `stelleWieder`; ohne Suchtext
löste danach nichts mehr ein Neuzeichnen aus.

**Nicht getestet:** Online-Kachel, Erreichbarkeit des Render-Servers, echtes Android 9,
Wiederherstellung bei anderer Bildschirmhöhe.

---

## 1f. Etappe 3, Schritt 6 (23.11.2026, 18:12–18:19)

Geänderter Code: `SpieleActivity.java` (560 → **619** Zeilen: `onSaveInstanceState`,
`stelleWieder`). Debug-Build auf Emulator B, Gerät A unberührt. Prozess-Tod echt erzeugt
mit `adb shell am kill`, Rückkehr über die Übersicht (Recents); Zustand über
`uiautomator` (Text, `selected`, `bounds`) und `logcat` gelesen.

| Prüfung | Ergebnis | Beleg |
|---|---|---|
| `:app:assembleDebug` | **erfolgreich**, `BUILD SUCCESSFUL in 10s` | Terminalausgabe |
| Prozess wirklich beendet | PID 8844 vorher, `pidof` nach `am kill` leer | `adb`, `pidof` |
| Neuer Prozess, neuer `onCreate` | PID **9354**, `Katalog: 94 Spiele, 9 Kategorien` neu im Log | `logcat` |
| Sicherung gelesen | `Zustand wiederhergestellt: „a“, 1 Merkmale, Scroll 3210` | `logcat` |
| Suchtext | „a“ im Suchfeld, Zähler „37 von 94“ | `uiautomator` |
| Filter-Chips | „Mehrspieler“ `selected="true"` | `uiautomator` |
| Scrollposition | dieselben Zeilen an denselben Stellen wie vor dem Tod: „Go“ `[83,1689]`, „Halma“ `[83,1882]`, „Kalaha“ `[83,2075]`, „Risiko – Welt erobern“ `[83,2268][997,2327]` | `uiautomator`, zwei Dumps verglichen |
| Weiterfiltern nach der Wiederherstellung | „Zurücksetzen“ → Suchfeld leer, kein Chip gewählt, „94 Spiele“, `Liste: 94 Zeilen in 23 ms` | `uiautomator`, `logcat` |
| Abstürze | **0** `FATAL EXCEPTION` | `logcat` |

Ablauf der Messung: Zustand gesetzt (Suche „a“, Chip „Mehrspieler“, drei Wischgesten →
Scroll 3210) → Home-Taste → `am kill` → Übersicht → Task-Karte. Ein Start der App über
`am start … MAIN/LAUNCHER` taugt für diesen Test **nicht**: er räumt den Task über
`MainActivity` ab (danach `sz=1`), die gesicherte `SpieleActivity` ist dann weg.

**Nicht getestet:** echtes Android 9, Online-Modus, Wiederherstellung bei anderer
Bildschirmhöhe.

---

## 1e. Etappe 3, Schritt 5 (23.11.2026, 18:00–18:10)

Geänderter Code: `SpieleActivity.java` (560 Zeilen: `onConfigurationChanged`, Breite in
`baueChips`). Debug-Build auf Emulator B, Gerät A unberührt. Drehung mit `adb emu rotate`;
Zustand über `uiautomator` (Text, `selected`, `bounds`) gelesen.

| Prüfung | Ergebnis | Beleg |
|---|---|---|
| `:app:assembleDebug` | **erfolgreich**, `BUILD SUCCESSFUL in 9s` | Terminalausgabe |
| Drehung **vor** der Änderung (Gegenprobe) | quer 2296 px → hochkant: „Lokal am Gerät“ auf 74 px gequetscht `[940,444][1014,1082]`, „Online“, „Gegen Computer“, „Highscore“ fehlen ganz, Liste beginnt erst bei y = 1115 | `uiautomator` |
| hochkant → quer (nach der Änderung) | alle 7 Chips in **einer** Zeile, letzter endet bei x = 2080 (Breite 2160) | `uiautomator` |
| quer → hochkant | wieder 3 Zeilen, alle innerhalb x ≤ 1036, jeder Chip normal breit | `uiautomator` |
| Suchtext | „a“ vor und nach beiden Drehungen | `uiautomator` |
| Filter-Chips | „Mehrspieler“ `selected="true"` vor und nach beiden Drehungen | `uiautomator` |
| Zähler | „37 von 94“ unverändert | `uiautomator` |
| Scrollposition | oberste sichtbare Zeile „Risiko – Welt erobern“ (quer y = 499, hochkant y = 753) | `uiautomator` |
| Abstürze | **0** `FATAL EXCEPTION` | `logcat` |

Drehfolge der Messung: hochkant (1080 × 2340) → quer (2340 × 1080) → hochkant
(1080 × 2340); die Liste blieb dabei dieselbe Instanz (kein neuer `onCreate`-Log).

**Nicht getestet:** echtes Android 9, Drehung mit geöffneter Tastatur. Der damals offene
Prozess-Tod ist in Abschnitt 1f nachgeholt.

---

## 1d. Etappe 3, Schritt 4 (23.11.2026, 17:58–18:04)

Geänderter Code: `ConnectorActivity.java` (Karte „Alle Spiele (offline)“ und
`spieleAnzahl()` entfernt, 461 → 442 Zeilen). Debug-Build auf Emulator B, Gerät A unberührt.

| Prüfung | Ergebnis | Beleg |
|---|---|---|
| `:app:assembleDebug` | **erfolgreich**, `BUILD SUCCESSFUL in 9s` | Terminalausgabe |
| Connector-Startseite | nur „Lokales Netzwerk“ und „Online verbinden“, **kein** „Alle Spiele (offline)“ | `uiautomator` |
| Lokal-Bildschirm | unverändert: „Räume suchen“, „Raum erstellen“, Raumcode, Chat, „Galgenmännchen im Raum“ | `uiautomator` |
| Startseite → „Spiele“ | `SpieleActivity`, Kopf „94 Spiele“, Filterleiste | `dumpsys window`, `uiautomator` |
| Abstürze | **0** `FATAL EXCEPTION` | `logcat` |

**Nicht getestet:** echtes Android 9, Rotation, Online-Modus.

---

## 1c. Etappe 3, Schritt 3 (23.11.2026, 17:44–17:55)

Geänderter Code: `SpieleActivity.java` (Merkmalsfilter, Zeilen-Cache),
`ConnectorActivity.java` (Kartentext, `spieleAnzahl()`). Keine Änderung an Spielen,
Assets, Manifest oder Version. Debug-Build auf Emulator B, Gerät A unberührt.

| Prüfung | Ergebnis | Beleg |
|---|---|---|
| `:app:assembleDebug` | **erfolgreich**, `BUILD SUCCESSFUL in 10s`, keine Warnung/Fehler im Filter | Terminalausgabe |
| Start → „Spiele“ | `SpieleActivity`, Log `Katalog: 94 Spiele, 9 Kategorien` | `logcat`, `uiautomator` |
| Filterleiste | 6 Chips in 3 Zeilen: Zurücksetzen, Einzelspieler, Mehrspieler / Lokal am Gerät, Online, Gegen Computer / Highscore – `controller` fehlt (kommt im Katalog nicht vor) | `uiautomator` |
| „Mehrspieler“ | Zähler **„70 von 94“**, Köpfe „Kartenspiele · 10 / 11“ | `uiautomator` |
| „Mehrspieler“ + „Gegen Computer“ | Zähler **„5 von 94“** (UND-Verknüpfung), Köpfe „Brettspiele · 4 / 22“, „Arcade und Klassiker · 1 / 14“ | `uiautomator` |
| „Zurücksetzen“ | „94 Spiele“, Suchfeld leer, alle Kategorien wieder da | `uiautomator` |
| Suche „zzz“ | „0 von 94“ und Hinweis „Kein Spiel passt zu Suche und Filtern …“ | `uiautomator` |
| Filterwechsel (Log `Liste: … Zeilen in … ms`) | 94 Zeilen **31 ms**, 23 Zeilen **39 ms**, 5 Zeilen **14 ms**; erster Aufbau 94 Zeilen **82 ms** | `logcat` |
| Scrollen, 6 Wischgesten, `gfxinfo` | 94 Zeilen: 237 Frames, Median **36 ms**, 90er **57 ms**, 33 % Janky | `dumpsys gfxinfo` |
| Scrollen, Vergleich 23 Zeilen | 140 Frames, Median **40 ms**, 90er **61 ms**, 39 % Janky – **kein** Unterschied durch die Eintragszahl | `dumpsys gfxinfo` |
| Renderer des Emulators | `ro.hardware.egl=emulation` (Software-GL) – erklärt die 36–40 ms | `getprop` |
| Connector-Karte | Connector-Startseite: „Alle Spiele (offline)“, „Die komplette Sammlung aus dieser App: **94 Spiele**, ohne Internet nutzbar …“ (Zahl aus `katalog.json`) | `uiautomator` |
| Karte antippen | „Spiele öffnen“ → `SpieleActivity` (94 Spiele, Filterleiste) | `dumpsys window`, `uiautomator` |
| Zurück aus der Liste | `ConnectorActivity` mit unveränderter Startseite | `dumpsys window` |
| Abstürze | **0** `FATAL EXCEPTION` | `logcat` |

**Weg zur Karte:** App-Startseite → „Lokales Netz“ öffnet den Connector auf dem
Lokal-Bildschirm; die Karte liegt auf dem Connector-Start (Kopfknopf „Start“). Das ist ein
Umweg, kein Fehler – im Entwicklungsstand als offener Punkt notiert.

**Nicht getestet:** echtes Android 9, Rotation, Online-Modus, Flüssigkeit auf echter
Hardware (nur softwaregerenderter Emulator), Antippen eines Spiels aus der über die Karte
geöffneten Liste (dieselbe Liste, bereits separat geprüft).

---

## 1b. Etappe 3, Schritt 2 (23.11.2026, 17:28–17:36)

Geänderter Code: `WebNavigation.java` (neu), `SpieleActivity.java` (neu),
`MainActivity.java` (stark gekürzt), `GameActivity.java` (Routing ausgelagert),
`WebBasis.java`, `AndroidManifest.xml`, `ConnectorActivity.java` (Text + Ziel),
`start/app.html` (Zahl).

| Prüfung | Ergebnis | Beleg |
|---|---|---|
| `:app:assembleDebug` | **erfolgreich**, `BUILD SUCCESSFUL in 16s` | Terminalausgabe |
| Start landet auf `app.html` | ja, Startseite zeigt „94 Spiele“ | `logcat`, Screenshot |
| Tipp auf „Spiele“ | öffnet **`SpieleActivity`** (`ResumedActivity`), Log `Katalog: 94 Spiele, 9 Kategorien` | `logcat`, `dumpsys activity` |
| Native Liste | Kopf „Spiele / 94 Spiele“, Suchfeld, „Kartenspiele · 11“, Zeilen mit Merkmalen | Screenshot |
| Suche „kniffel“ | Zähler „1 von 94“, nur „Würfelspiele · 1“ mit Kniffel | Screenshot |
| „Solitär“ (nur Solo) | ohne Dialog `GameActivity`, `Seite geladen: …/solitaer.html` | `logcat` |
| „Baccara“ (Online) | Dialog → „Lokal / offline“ → `GameActivity`, `…/baccara.html` | `logcat`, `uiautomator` |
| Zurück aus dem Spiel | zurück in `SpieleActivity` (Liste bleibt stehen) | `dumpsys activity` |
| Abstürze | **0** `FATAL EXCEPTION` im Log | `logcat` |

**Nicht getestet:** Merkmals-Filter (die Liste filtert nur per Textsuche), Online-Modus,
Rotation, echtes Android 9, Connector-Karte „Alle Spiele (offline)“ (Ziel geändert, nicht
angetippt).

---

## 1a. Etappe 3, Schritt 1 (23.11.2026, 17:15–17:25)

Geänderter Code: `GameActivity.java` (neu), `WebBasis.java` (neu), `MainActivity.java`
(gekürzt), `AndroidManifest.xml`, `app/build.gradle` (Aufgabe `katalogBauen`),
`.gitignore`. Keine Änderung an Spiele-HTML, Assets oder Connector.

| Prüfung | Ergebnis | Beleg |
|---|---|---|
| `:app:assembleDebug` | **erfolgreich**, `BUILD SUCCESSFUL in 17s` | Terminalausgabe |
| Debug-APK | **11.377.723 B** (vorher 11.346.443 B) | `ls -la` |
| `katalogBauen` gelaufen | 94 Spiele, 9 Kategorien → `assets/katalog.json` (33.896 B), im APK enthalten | `unzip -l`, JSON gelesen |
| Katalog-Daten | 94 Spiele, keine doppelten Pfade; Kategorien karten/wuerfel/brett/denk/arcade/action/jump/abenteuer (+ sonstige in der Liste) | JSON-Auswertung |
| Start landet auf `app.html` | **ja** – `I TrashyGeckoHome: Seite geladen: https://appassets.androidplatform.net/app.html` | `logcat` |
| Katalog öffnet weiterhin in `MainActivity` | `Seite geladen: …/spiele/spiele.html`, Anzeige „94 Spiele in 6 Kategorien“ | `logcat`, Screenshot |
| Spielklick → `GameActivity` | Tipp auf „Baccara“ → Dialog → „Lokal / offline“: `I TrashyGeckoGame: Spielseite geladen: …/baccara.html`, `ResumedActivity = GameActivity` | `logcat`, `dumpsys activity` |
| Zurück aus dem Spiel | kurzer Druck auf Zurück → `ResumedActivity = MainActivity`, Katalog mit aufgeklappter Kategorie und gleicher Scrollposition | `dumpsys activity`, Screenshot |
| Mahjong | öffnet in `GameActivity` (`…/mahjong.html`), Brett und Einstellungen werden gezeichnet | `logcat`, Screenshot |
| Abstürze | keine `FATAL`/`AndroidRuntime`-Fehler der App im Log | `logcat` |

**Nicht getestet:** „SPIEL VERLASSEN“ im Mahjong-Menü (Brücke wird wie vorher injiziert,
aber nicht ausgelöst), „ZURÜCK ZUM CONNECTOR“, Online-Modus, Rotation, echtes Android-9-Gerät.

**Umgebungsproblem am Rand:** Der Emulator zeigte nach dem Start erneut „System UI isn't
responding“ (Fokus im ANR-Fenster). Nach Tipp auf „Wait“ war die App bedienbar; dieser
Dialog gehört zur System-UI des Emulators, nicht zur App.

---

## 2. Was am 23.11.2026 wirklich lief (Etappe 1)

| Prüfung | Ergebnis | Beleg |
|---|---|---|
| Gradle-Build `:app:assembleDebug` | **erfolgreich**, `BUILD SUCCESSFUL in 46s`, 34 Tasks | Terminalausgabe |
| Debug-APK entstanden | **11.346.443 B**, `app/build/outputs/apk/debug/app-debug.apk` | `ls -la` |
| `homepageKopieren` beim Bauen | gelaufen (hängt an `preBuild`), keine `GradleException` | Buildprotokoll |
| Asset-Baum danach | **353 Dateien, 108 HTML-Seiten** unter `assets/www` | `find … | wc -l` |
| Katalogprüfung `verifyHomepageGames` | **nicht** gelaufen (hängt an `packageTrashyGeckoApk`, nicht an `assembleDebug`) | – |
| Geräteverbindung | Gerät A und Emulator B von `adb devices` erkannt | `adb devices` |
| Versionsstand auf Gerät A | 1.15 / 16 abgefragt | `dumpsys package` |
| Debug-APK auf Emulator B installieren | **erst gescheitert** (`INSTALL_FAILED_UPDATE_INCOMPATIBLE`: dort lag die release-signierte Fassung), nach `adb uninstall` → `Success` | `adb install` |
| App auf Emulator B starten | Prozess lief (`pidof` = 2991), **kein** `AndroidRuntime`-Absturz im Log | `logcat` |
| Sichtprüfung auf dem Emulator | **nicht möglich**: Emulator meldete nach dem Neustart „System UI isn't responding“, das ANR-Fenster hielt den Fokus, ein Screenshot zeigt nur den Dialog | `dumpsys window`, Screenshot |

**Nicht** gelaufen (und deshalb nicht behauptet): Navigation, Spieleliste, Spielstart,
Connector-Verbindung, LAN-Raum, Hotspot, Online-Multiplayer, Fehlerbehandlung, Rotation,
Android-9-Tauglichkeit dieser Änderung. Grund: Etappe 1 war Bestandsaufnahme; die App
wurde in dieser Runde nicht verändert, und der Emulator war danach nicht mehr bedienbar
(System-UI-ANR).

**Zustand der Geräte nach dieser Runde:** Gerät A unverändert (1.15, release-signiert,
Daten intakt). Emulator B hat jetzt die Debug-APK ohne Daten (die alte Installation wurde
entfernt); er braucht vor weiteren Tests einen Neustart, weil die System-UI hängt.

---

## 3. Was aus früheren Runden dokumentiert ist (fremde Messungen, nicht von mir wiederholt)

Aus `spiele/android/README.md` (Fassungen 1.10–1.14, 20./21.11.2026) – **nicht** neu
geprüft, hier nur als Ausgangsstand zitiert:

- Release-APK 1.14: `BUILD SUCCESSFUL in 16s`, 10.492.045 B, md5 `cc5484bb…`,
  release-signiert (`CN=TrashyGecko`, SHA-256 `73:01:9C:01:…:BA:FF`), 470 APK-Einträge.
- `adb install -r` auf dem Emulator (Android 14): `Success`, keine Daten gelöscht.
- Auf dem Emulator geprüft: Startseite mit vier Kacheln, Extras-Bildschirm,
  Fernbedienung geladen, Zurück-Weg Fernbedienung → Extras → Start.
- `:app:remoteProtocolTest`: 47 Assertions bestanden.
- Huawei FIG-LX1 (Android 9): Installation und Update ohne Deinstallation bestanden,
  Mahjong-WebView mit lokalen Assets, lokaler NSD-Raum auf demselben Gerät gefunden.
- Ausdrücklich **nicht** bestätigt (auch damals nicht): zweites Gerät im WLAN, Hotspot,
  Chat-/Spielzug-Synchronisierung zwischen zwei Geräten, echte Fernseher, Tablets,
  API 21/24, Play-Store-Installation.

---

## 4. Offene Prüfungen für die nächsten Etappen

1. **Nach dem Umbau**: `:app:packageTrashyGeckoApk` inklusive `verifyHomepageGames`
   (Katalogzahl > 0, jede Datei vorhanden).
2. **Gerät A (Android 9)**: release-signiert bauen, `adb install -r`, Start, native
   Startseite, Spieleliste (alle Kategorien), drei Spiele öffnen und zurück, Rotation,
   Zurück-Taste, App-Neustart, Offline-Betrieb (Flugmodus).
3. **Netzwerk**: Online-Status bei erreichbarem und bei abgeschaltetem Server
   (Flugmodus bzw. falscher Host) – die Fehlermeldung muss stimmen, nicht „verbunden“
   zeigen.
4. **LAN**: zwei Geräte im selben WLAN (Gerät A + Emulator erreichen sich nicht, dafür
   braucht es ein zweites echtes Gerät oder zwei Emulatoren im selben virtuellen Netz).
5. **Hotspot**: echtes Gerät als Hotspot, zweites Gerät als Client.
6. **Online-Multiplayer**: Raum auf `trashygecko.onrender.com` mit zwei Clients.
7. **Android 9-Kompatibilität**: mindestens Start, Spieleliste, ein Spiel und die
   Einstellungen auf Gerät A.

---

## 5. Bekannte Fehler und Stolperfallen (aus Code und Dokumentation)

- `keystore.properties` fehlt im Projekt → `packageTrashyGeckoApk` baut nur Debug.
  Der Keystore liegt unter `Documents\arbeit\werkzeuge\trashygecko-release\`
  (Passwort in `passwort.txt` daneben – **nie** in den Chat oder ins Repo).
- `MainActivity` lädt mit `LOAD_NO_CACHE`; ohne das zeigte der WebView nach einem Update
  alte Dateien (behoben in 1.14, nicht zurückdrehen).
- `onRenderProcessGone` in `MainActivity`/`RemoteControlActivity`: ohne den Neustart
  bleibt die App nach einem WebView-Absturz schwarz.
- Der lokale Relay ist unverschlüsselt; nur im vertrauenswürdigen Heimnetz.
- Emulator B meldete beim Prüfen zeitweise `Can't find service: package` – er war noch
  nicht fertig gebootet. Vor Gerätetests immer `adb wait-for-device` und
  `getprop sys.boot_completed` prüfen.
