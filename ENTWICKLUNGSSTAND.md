# Entwicklungsstand – TrashyGecko Android

Lebendes Dokument. Nach **jeder** abgeschlossenen Arbeit fortschreiben (oben das Neueste).
Was getestet wurde, gehört zusätzlich in [`PRUEFBERICHT-ANDROID.md`](PRUEFBERICHT-ANDROID.md).

---

## 24.11.2026 · Spielhalle: eigene Spielsymbole, Kategorien-Navigation, sichtbare eigene Runde

**Warum:** Die Spielkarten zeigten nur Buchstabenkürzel, zwischen Kategorien führte
nur der Umweg über den Start, und eine eröffnete Runde war nach dem Schließen des
Tisches nicht mehr sichtbar.

**Was gemacht wurde**
- **Eigene Spielsymbole für alle 94 Karten:** 45 SVG-Zeichen (Würfel, Karten, Krone, Mühle, Bombe,
  Kiste, Juwel, Alien, Galgen …), Zuordnung über Pfadregeln, Rückfall Kategorie.
  Gemessen: 94/94 Karten mit Symbol, Würfelspiele 25/25.
- **Kategorien-Navigation:** unter jeder Kategorie die fünf anderen als Kacheln mit
  Symbol, Name und Spielzahl; gemessen 5 Ziele, beschriftet.
- **Eigene Runde bleibt sichtbar:** Leisten-Knopf „Runde: 10.000“, Panel-Block
  „Offene Runde“ mit Tisch öffnen / Zurückziehen (`/api/angebot` `zurueck`);
  „Partie fortsetzen“ überlebt das Schließen (`letzterTisch`, `tg-halle-tisch`).
- **Textfehler behoben:** `el(tag, klasse)` nimmt keinen dritten Parameter – Stern,
  ›-Pfeil, „Raum + Chat“ und die Kategorien-Beschriftungen waren leer. Jetzt
  `textContent`; Zurück-Pfeil der Suche ist ein echtes `←`.

**Geprüft** (Chromium, Connector 8768, zwei Ursprünge: 127.0.0.1 und
192.168.56.1)
- Raum TG-6348, Anna + Gast YU1, je 2 Spieler; 10.000 eröffnet, Angebot
  angenommen, beide bereit, Host startet; Chat in beide Richtungen (9 Beiträge);
  0 Konsolenfehler.
- Nach dem Schließen: „Runde: 10.000“ in der Leiste, Panel „Deine Runde wartet …“,
  Tisch öffnet ohne zweite Runde.
- Handy 390 × 844: Start/Kategorie/Highlights je 0 px Querscrollen.
- `go test -run Register ./...`: PASS (71 Register-Einträge, 70 Katalogspiele,
  24 Spiele ohne Online-Teil, `malen` ist Werkzeug).
- APK neu gebaut (Assets enthalten `halle.css`, `halle-online.js`), Version
  unverändert 1.16 / Code 17.

**Nachtrag: Bereitschaft und Start in der Halle, drei weitere Spiele geprüft**
- Die Tisch-Kopfzeile hat **Bereit** und **Spiel starten** (Gastgeber, erst wenn
  alle bereit sind). Dieselben Endpunkte wie die Spielseite (`action=ready/start`),
  keine zweite Lobby.
- **Galgen/Kniffel/Schach** vollständig über die Halle gespielt (Raum TG-96F2,
  Loopback + LAN): Beitritt, Bereit, Start, laufende Partie sichtbar, 0 Fehler.
- Behoben: laufender Zustand `spiel` wurde nicht als „Partie läuft“ erkannt;
  offenes Tisch-Angebot blieb nach dem Start stehen; Schach (ohne `players`)
  meldete fälschlich „Warte auf Mitspieler“.
- Handy 390 px mit den neuen Knöpfen: 0 px Querscrollen in Kopfzeile und Seite.

**Offen:** vier Spiele durchgespielt (10.000, Kniffel, Galgen, Schach), 66 weitere
ungetestet; zwei Ursprünge auf einem Rechner statt zwei echter Geräte; kein
Freundesystem im Connector (Anwesenheit nur im Raum); Start aus der Spielseite
räumt das Hallen-Angebot nicht auf (Connector hat keine gemeinsame Statusabfrage).

---

## 24.11.2026 · Web-Spielhalle: neue Optik, neue Highlights, durchgängiger Online-Ablauf

**Warum:** Die Übersicht war grau und schlicht; Räume, Einladungen und Chat lagen nur im
Connector, nicht in der Spielhalle.

**Highlights**
- Flipper und Quizshow sind aus den vier Highlights entfernt, **Mahjong** und **10.000**
  sind neu. Bilder: `spiele/bilder/highlights/{mahjong,zehntausend}.jpg` (1392 × 752,
  91 KB / 86 KB), erzeugt im Stil der vorhandenen zwei Bilder.
- Jede Highlight-Karte zeigt Bild, Kategorie, Kurztext, Modus-Marken und einen direkten
  Einstieg; 10.000 trägt zusätzlich „Online 2–6“. Der Pfad `/spiele/bilder/highlights/…`
  bleibt im Katalog und damit im Android-Bau unverändert.

**Optik** (`spiele/halle.css`, ersetzt das bisherige Inline-CSS in `spiele.html`)
- Dunkler Grund aus drei Radialverläufen, feinem Raster und Vignette; Kategorie-Farben
  wirken als Tint, Rand, Symbol und Akzentkante. Große Bildkarten, sechs
  Kategorie-Kacheln, kompakte Spielkarten mit Modus-Chips.
- Hover/Fokus-Animationen, gestaffelter Auftritt (`--i`, 26 ms), `prefers-reduced-motion`
  schaltet alles ab, `prefers-contrast:more` hebt Texte an, Fokus bleibt sichtbar.
- Raster: Handy 1 Spalte Karten/2 Kategorien, ab 560 px 2, ab 940 px 3, ab 1240 px 4
  Karten und 6 Kategorien. Kopfzeile und Online-Leiste sind auf 390 px einzeilig.
- Mädchen-/Buben-Stil bleiben über die vorhandenen `--tg-*`-Tokens erhalten; die dunklen
  Kartenspiele wurden nicht angefasst.

**Online** (`spiele/halle-online.js`, neue Schicht – kein zweiter Connector)
- Identität wie im Connector: `tg-player-id`, `tg-player-name`, `token.js`; ohne Namen
  entsteht „Gast XYZ“, änderbar im Panel.
- „Mit Freunden spielen“ legt einen **eigenen Raum** an (`POST /api/rooms` mit
  `neuerRaum`), Einladungen laufen über `?raum=CODE`, Raumcode und QR-Bild
  (`/api/qr?url=`). Betreten per Code oder Link; Anwesenheit über `/api/presence` und
  `/api/players` (20 s / 7,5 s).
- Offene Runde über `/api/angebot`: Host „bieten“ (Tischspiele erzeugt der Server),
  Gäste „annehmen“ oder „ablehnen“; eigene Runde lässt sich wieder öffnen. Ein
  laufendes Spiel wird über `/api/<spiel>` erkannt und fortgesetzt statt doppelt eröffnet.
- Der Tisch läuft im Overlay als **dieselbe Spielseite** (`?room&name&id&v`); Bereitschaft
  und Start bleiben dort, wo die Lobby schon war. Die Kopfzeile zeigt Spielerzahl,
  Bereitschaft und Status (alle 2,5 s).
- Chat: `/api/messages`, als Dock rechts (Desktop) bzw. Blatt unten (Handy), ungelesene
  Beiträge am Chat-Knopf, Einladungszeilen mit „Mitspielen/Ablehnen“.
- Reconnect: 409 → gleicher Platz neu, 403 → neue Kennung (höchstens alle 30 s),
  404 → neuer Raum; gespeicherter Raum und Tisch werden beim Laden wieder aufgenommen.
- Nur Spiele aus `multiplayer-register.js` bekommen den Online-Einstieg; alle anderen
  starten unverändert sofort. Der Filter „Online spielbar“ zeigt die echten Fälle.

**Geprüft** (Chromium/Playwright, zwei getrennte Kontexte, Connector lokal Port 8769)
- Startseite 1440 × 900: 4 Highlights, 6 Kategorien, keine Konsolenfehler.
- Raum TG-5BD3 über „Mit Freunden spielen“, Gast über `?raum=TG-5BD3`: beide sehen
  2 Spieler. 10.000 → „Online-Runde“ → Tisch; Gast nimmt das Angebot an; beide „Bereit“,
  Host startet. Danach in beiden Köpfen „Partie läuft · 2 Spieler“.
- Chat in beide Richtungen („Hallo von A …“, „Hallo von B …“), 10 Beiträge im Verlauf.
- Handy 390 × 844: gleicher Ablauf, Chat als Blatt unten (62 % Höhe), Screenshots geprüft.
- Neuladen des Gasts → „Partie fortsetzen“ → laufende Partie wieder da. Schließen und
  erneut 10.000 wählen → „Eigene Runde öffnen“, **1 am Tisch** (keine zweite Runde).
- „Jetzt spielen“ führt weiterhin auf `…/zehntausend.html` **ohne** `?room`.
- `node --check` für `halle-online.js` und die ausgelagerte Inline-JS: grün.

**Android**
- `:app:katalogBauen` neu: 94 Spiele, 6 Kategorien, Highlights Mahjong/10.000/Risiko/
  Labyrinth. `:app:packageTrashyGeckoApk`: 14.690.966 B,
  md5 `0c0450069e852ff4151b9e79e7658ddb`; APK enthält `halle.css`, `halle-online.js` und
  beide neuen Bilder. Version bleibt **1.16 / Code 17** (Änderung nur mit Rückfrage).
- Die App selbst wurde in dieser Runde nicht neu installiert/getestet.

**Offen**
- Ende-zu-Ende nur mit 10.000 geprüft; die übrigen Online-Spiele des Registers nutzen
  denselben Weg, wurden aber nicht durchgespielt.
- Zwei Browser-Kontexte auf einem Rechner, **kein** echtes zweites Gerät im WLAN.
- Bereitschaft/Start bleiben bewusst in der Spielseite; die Halle zeigt nur den Stand.
- Ohne laufenden Connector bleibt die Halle beim Einzelspieler-Teil.

---

## 23.11.2026 · App-Spielhalle: Startseite, Kategorien und Suche aus `katalog.json`

**Warum:** Die Web-Spielhalle hatte Kategorien, Highlights und eine Suche, die App zeigte
weiterhin die flache Liste mit 94 Einträgen.

**Was gemacht wurde** (D-26)
- `katalogBauen` überträgt jetzt auch die Kurzbeschreibungen (`TEXTE`), die
  Kategorie-Texte samt `akzent`-Farbe und die vier `HIGHLIGHTS` nach `katalog.json`
  (94 Spiele mit Text, 6 Kategorien, 4 Highlights). Die Zeilenform in `spiele.html`
  bleibt unverändert; Highlight-Pfade werden relativ zu `spiele/` aufgelöst.
- `SpieleActivity` (614 → **1055** Zeilen) hat drei Ansichten: Start mit vier
  Highlight-Karten (2 Spalten, ab 900 dp 4) und sechs Kategorie-Kacheln (2/3/6 Spalten je
  Breite), eine Kategorie mit Beschreibung und **nur** ihren Spielen sowie die
  Trefferliste für Suche/Filter, nach Kategorie gruppiert. Jede Spielkarte zeigt Name,
  Kurztext und vorhandene Modi; die Suche prüft Name, Suchwörter, Text und
  Kategoriename (Umlaute egal), die Chips bleiben unverändert (UND).
- Zurück führt in dieser Reihenfolge: Filter löschen → Kategorie verlassen → Activity
  beenden; der Pfeil in der Kopfzeile nimmt denselben Weg.
- Prozess-Tod: `onSaveInstanceState` sichert zusätzlich Ansicht und offene Kategorie
  (`ZUSTAND_ANSICHT`, `ZUSTAND_KATEGORIE`); Suche, Merkmale und Scroll bleiben wie
  bisher. Die Reihenfolge `stelleWieder` → einmal `zeichne` ist unverändert.
- Highlight-Bilder kommen aus `assets/www/spiele/bilder/highlights/*.jpg`, verkleinert
  gelesen (`inSampleSize = 3`, ~464 × 251); Karten einer Rasterzeile bekommen die Höhe
  der höchsten Karte.

**Geprüft (Emulator 1080 × 2340, Debug-APK)**
- Start: „94 Spiele“, 7 Filter-Chips, 4 Highlights (Karten einer Zeile gleich hoch),
  6 Kacheln mit 25/11/22/15/20/1 Spielen, keine abgeschnittenen Texte.
- „Würfelspiele“ → 25 Einträge mit Beschreibung; Suche „kniffel“ → „1 Treffer“.
- Prozess-Tod in der Kategorie bei Scroll 1014: `am kill` (PID 18129 → leer), Start
  über die Übersicht → PID 18451, Log „Ansicht kategorie (wuerfel), Scroll 1014“,
  Liste an derselben Stelle. Suche „kniffel“ überlebte ebenso („1 Treffer“).
- Prozess-Tod mit dem endgültigen Bau: Suche „bejeweled“ gesetzt, `am kill`
  (PID 4883 → leer), Task zurückgeholt → PID 5308, Log „Zustand wiederhergestellt:
  „bejeweled“, 0 Merkmale, Ansicht start“, „1 Treffer“.
- Dialog „Lokal / offline“ / „Online-Spiel“ erscheint, „Lokal / offline“ öffnet
  `GameActivity`; Bejeweled lädt („Neues Spiel“, Anleitung, Spielfeld), der Prozess
  bleibt, 0 `FATAL` im Logcat. (Beim ersten Versuch verschwand der Emulator selbst –
  Adb „device offline“, kein App-Absturz; nach einem Neustart lief es.)
- `:app:assembleDebug` und `:app:packageTrashyGeckoApk` grün; `katalog.json` in beiden
  Ablagen md5-gleich (`440ee0e5…`), ebenso `SpieleActivity.java` (`005b4b3d…`) und
  `build.gradle` (`9a5f009d…`).
- App **1.16** (versionCode 17), signiert (CN=TrashyGecko), **14.475.624 B**;
  `tmahjong.apk` und `TrashyGecko.apk` in beiden Ablagen md5-gleich (`ed23415b…`).

**Grenze:** Keine ganze Partie durchgespielt; Mehrspieler nur bis zum Modus-Dialog
geprüft (die Spielseiten selbst sind unverändert). Nicht committet, nicht veröffentlicht.

## 23.11.2026 · Spielhalle: Kategorien, Highlights und Suche statt langer Liste

**Warum:** Die Spieleübersicht war eine einzige, immer länger werdende Liste mit
aufklappbaren Abschnitten. Am Handy musste man weit scrollen, Kategorien und Spielmodi
waren nicht auf einen Blick zu erkennen.

**Was gemacht wurde** (D-25)
- `spiele/spiele.html` neu aufgebaut: Startseite mit vier großen Highlight-Karten, sechs
  Kategorie-Kacheln (Würfel, Karten, Brett, Denken, Arcade/Geschicklichkeit,
  Quiz/Partyspiele), zentrale Suche und je Kategorie eine eigene Ansicht mit kompakten
  Spielkarten (Name, Symbol, Kurzbeschreibung, vorhandene Spielmodi, „Spielen“).
- Ansichten über Hash-Adressen (`#/k/wuerfel`, `#/highlights`, `#/werkzeuge`); der
  Zurück-Pfeil bleibt beim Scrollen oben stehen. Keine waagerechte Scrollleiste.
- Spielmodi kommen ausschließlich aus den vorhandenen Merkmalen (`solo`, `lokal`,
  `online`, `computer`); nicht vorhandene Modi werden nicht angezeigt. Die Auswahl
  passiert weiterhin erst im Spiel.
- Kategorie-Zuordnung gestrafft: Action, Jump & Run und Abenteuer gehören jetzt zu
  „Arcade und Geschicklichkeit“, die Scherzshow zu „Quiz und Partyspiele“. Das zweite
  „Solitär“ heißt zur Unterscheidung „Steine-Solitär“ (gleiche URL, gleiches Spiel).
- Katalogzeilen in `SPIELE`/`WERKZEUGE` unverändert im Format – `katalogBauen` und
  `homepageKopieren` lesen sie weiterhin. Neu sind nur `TEXTE` (Kurzbeschreibungen),
  `HIGHLIGHTS` und vier Vorschaubilder unter `spiele/bilder/highlights/`.
- Nichts an Spielen, Connector, Spielständen, URLs oder der Grafikstil-Auswahl geändert.

**Tatsächlich getestet** (Playwright gegen `python -m http.server`, Desktop-Chromium)
- 360×740: 6 Kategorie-Kacheln, 4 Highlights, Startseite **1274 px** hoch (vorher 1809),
  keine waagerechte Scrollleiste; Würfelspiele 25 Karten, davon 16 mit Einzelspieler.
- Suche: „mau“ → Mau Mau, „kniffel“ → Kniffel, „muhle“ → Mühle (ohne Umlaut),
  „qqqq“ → Hinweis; Leeren stellt die Übersicht wieder her.
- Modi: UNO zeigt „Mehrspieler: lokal & online“, Solitär nur „Einzelspieler“.
- 1280×800: 6 Kacheln und 4 Highlights je in einer Reihe, Spielkarten in 4 Spalten.
- Alle **94** Spiele per HTTP **200**; UNO, Schach, Jule und Pong laden ohne Konsolen-,
  Seiten- oder Ladefehler. Grafikstil „Mädchen“ greift weiterhin.
- `:app:katalogBauen` und `:app:verifyHomepageGames`: **94 Spiele, 6 Kategorien**,
  `BUILD SUCCESSFUL in 41s`; die APK-Kopie enthält die beiden Installer-Zeilen nicht mehr.

**Offen**
- Nur Browserprüfung bei zwei Größen; kein Emulator- oder Gerätetest in dieser Runde.
- Die zwei zufallsabhängigen Go-Tests (Quiz-Antwortzitate, Zehntausend-Wurf) bleiben
  sporadisch – unabhängig von dieser Arbeit.

---

## 23.11.2026 · Etappe 3, Schritt 11: Homepage-Seiten liegen im APK

**Warum:** Im APK lag nur `index.html` (als `homepage.html`). Fast jede Kachel der
Startseite (Tagebuch, Bücher, Rezepte, Tipps …) lief deshalb auf 404 – die Seite war ohne
Netz weitgehend tot.

**Was gemacht wurde** (D-24)
- `app/build.gradle`: neue Liste `homepageSeiten` mit den 19 Ordnern, die `index.html`
  verlinkt; `homepageKopieren` kopiert sie samt `main.css` nach `www/<ordner>/`. Draußen
  bleiben der alte Seiten-Spiegel unter `css/` (in `.gitignore`, u. a. mit alten
  Zugangsdaten), `musik/.backup-*` und `tagebuch/Textdokument (neu).txt`.
- Sonst keine Codeänderung: `HomepageActivity`, `WebNavigation` und `HomepageAssets`
  bleiben, weil interne Links schon im WebView bleiben.

**Tatsächlich getestet** (Emulator Android 14, Flugmodus, `Active default network: none`)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 18s`; APK 11.460.085 → **11.959.743 B**
  (+499.658 B), `assets/www` 13 → 14 MB, **303** HTML-Seiten im Paket.
- 23 lokale Adressen aus der Seite heraus abgerufen (`fetch`): alle **200**.
- Echte Klicks: Tagebuch, Rezepte, Bücher, Tipps öffnen (Titel „Tagebuch“, „Rezepte“,
  „Bücher“, „Tipps“), zwei Ebenen tiefer `tipps/hoersturz.html` („Hörsturz (Gehörsturz)“)
  und `seiten/programmieren.html` → `../css/css.html` („CSS“).
- 0 `Seitenfehler` im Protokoll, 0 kaputte Bilder, 0 `FATAL EXCEPTION`.

**Offen**
- Neun Ziele gibt es auch in der Quelle nicht (Beispielpfade der Anleitungen:
  `bilder/beispiel.jpg`, `home.png`, `html/home.png`, `css/bild.gif`, `css/bild.png`,
  „Adresse des Bildes“) – die bleiben tot, im Browser wie im APK.
- Android 9 auf Gerät A (gesperrt) und eine Sichtprüfung des Seitenbildes.

---

## 23.11.2026 · Etappe 3, Schritt 10: Online-Kachel sagt die Wahrheit

**Warum:** Ohne Netz zeigte die Online-Kachel Chromiums englische Fehlerseite
(`net::ERR_INTERNET_DISCONNECTED`), während die Statuszeile „Online-Connector · bestehender
Server“ behauptete. Ein Nutzer ohne Netz sah also eine englische Technikmeldung plus eine
falsche Zusage.

**Was gemacht wurde** (D-23)
- `ConnectorActivity` (458 → **494**): `istOnline()` prüft vor dem Laden das aktive Netz;
  ohne Netz wird gar kein WebView gebaut.
- `zeigeOnlineFehler(...)` verwirft den WebView und zeigt „Online-Connector nicht erreichbar“
  mit deutschem Erklärungssatz und Knopf „Erneut versuchen“.
- Statuszeile folgt dem Verlauf: „wird geladen …“ → „Server erreichbar“; bei Fehler
  „Offline · keine Verbindung“ bzw. „Online-Connector · Server nicht erreichbar“.
- `onReceivedError` im Hauptframe greift auch, wenn das Netz da ist, der Server aber nicht
  antwortet.

**Tatsächlich getestet** (Emulator Android 14)
- Flugmodus (`Active default network: none`): Kachel → deutsche Meldung mit Knopf, Status
  „Offline · keine Verbindung“.
- Flugmodus aus, Knopf gedrückt → Seite lädt, Status „Online-Connector · Server erreichbar“.
- Server weg (Testbau mit unerreichbarer Adresse): Status „Server nicht erreichbar“, gleiche
  Meldung; Testbau zurückgebaut, Endfassung erneut im Flugmodus geprüft.
- 0 Abstürze in allen Läufen.

**Offen**
- Gerätetest Android 9 für diesen Weg (Gerät gesperrt, siehe Schritt 9) und eine Sichtprüfung
  der geladenen Connector-Seite (WebView-Inhalt steht nicht im UI-Baum).

---

## 23.11.2026 · Etappe 3, Schritt 9: Startansicht und Spieleliste teilen Farben und Bausteine

**Warum:** Die native Startansicht hatte ihr eigenes Farbbild (leicht andere Grün- und
Grautöne als die Spieleliste) und eigene Kopien von `dp` und `flaeche`; `onResume` prüfte
außerdem `status` auf null, obwohl die Ansicht in `onCreate` immer entsteht.

**Was gemacht wurde**
- Neu `Ansicht.java` (53 Zeilen): `BG`, `KARTE`, `RAND`, `GOLD`, `TEXT`, `MUTED` an einer
  Stelle, dazu `flaeche(farbe,radius)`, `flaeche(farbe,radius,rand,breite)`, `text(...)` und
  `dp(context,wert)`.
- `MainActivity` (193 → **173**): benutzt nur noch `Ansicht`; eigene Farbkonstanten,
  `text()`, `flaeche()`, `dp()` und der `status != null`-Test sind weg. Die Netzstatus-Logik
  ist unverändert, nur mit einer Notiz zur API-Lage.
- `SpieleActivity` (630 → **614**): Farben und Helfer ebenfalls aus `Ansicht`, lokal bleibt
  nur der Chip-Füllton. `CHIP_RAND` ist entfallen, die Chip-Kante nimmt `Ansicht.RAND`.
- Startseite nutzt jetzt dieselben Töne wie die Spieleliste (Hintergrund `#0B241D` statt
  `#07140F`, Text `#E9F1EB`, Grau `#9AA7A0`, Karte `#123028`, Kante `#2B4C40`).

**Tatsächlich getestet** (Emulator Android 14; Android 9 nur Statuslogik)
- Build `BUILD SUCCESSFUL in 10s`; Startansicht mit gleichen Maßen (Titel y 241, Kacheln
  553/864/1175/1486), Status „WLAN · online bereit“.
- Spieleliste: „94 Spiele“ → „Mehrspieler“ „70 von 94“ → „Zurücksetzen“ „94 Spiele“,
  `Liste: 94 Zeilen in 39 ms`, 0 Abstürze.
- Android 9 (Gerät A, API 28): Testbau mit eigener Paket-ID, Gerät gesperrt, deshalb über
  eine Protokollzeile gemessen – `Netzstatus: WLAN · online bereit`, keine Ausnahme.
  Testbau wieder deinstalliert, die release-signierte 1.15 blieb unangetastet.

**Offen**
- Startansicht und Offline-Zweig auf Android 9 mit Bildschirm (Gerät gesperrt; Flugmodus
  dort nicht geschaltet).

**Nächster konkreter Arbeitsschritt**
1. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.
2. Danach Statusanzeige mit Serverprüfung und Einstellungen.

---

## 23.11.2026 · Etappe 3, Schritt 8: „Extras → Homepage“ mit der echten Startseite

**Warum:** D-12 verlangt die echte Website als eigenen Eintrag unter Extras; bis hierher
gab es den Eintrag nicht.

**Was gemacht wurde**
- `homepageKopieren` kopiert die echte `index.html` als `www/homepage.html`; `main.css`
  liegt dort schon. `www/index.html` bleibt die alte App-Startseite (`app.html`) als
  Rückweg aus Spielseiten – die beiden Namen kollidieren also nicht.
- Neu `HomepageActivity` (142 Zeilen): WebView über denselben `HomepageAssets`-Lader,
  Vollbild, `WebBasis`-Einstellungen und die Link-Regel `WebNavigation.pruefe(uri, true)`
  („Spiele“ öffnet die native Liste, fremde Adressen den Browser); Zurück geht erst im
  Verlauf zurück, sonst in die App.
- `ConnectorActivity.showExtras` (442 → **453** Zeilen) hat die Karte „Homepage“; sie
  erscheint nur, wenn `www/homepage.html` im Paket liegt (`isAssetAvailable`), sonst steht
  dort ein Hinweis. Die Fernbedienungs-Prüfung nutzt jetzt denselben Helfer.
- Manifest: `HomepageActivity` mit denselben `configChanges` wie die übrigen Activities.

**Tatsächlich getestet** (Emulator Android 14, Debug-Build)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 17s`; `assets/www/homepage.html`
  md5-gleich zu `index.html` (`08e7a1b8…`), `main.css` 17.579 B daneben, `www/index.html`
  (App-Start) unverändert 4.249 B.
- Extras zeigt die Karte „Homepage“ (Titel `[94,1145]`, Knopf „Homepage öffnen“ `[94,1383]`).
- Tipp auf den Knopf: `HomepageActivity` oben, Log
  „Homepage geladen: https://appassets.androidplatform.net/homepage.html“; Inhalt:
  Titel „TrashyGecko“, „Google Search“, Links (ChatGPT, deepseek, Gemini …), Uhr
  „18:51:14 Uhr“. Einziger Fehlzugriff: `www/favicon.ico` (harmlos).
- Zurück-Taste: wieder `ConnectorActivity` (Extras-Kopfzeile), 0 `FATAL EXCEPTION`.

**Offen / nicht getestet**
- Die weiteren Seiten der Website (Tagebuch, Bücher, Rezepte …) liegen seit Schritt 11
  (D-24) im Paket; dieser Abschnitt bleibt als Stand von Schritt 8 stehen.
- Das Google-Formular zielt auf `http://` und erzeugt eine Mixed-Content-Warnung im Log;
  ob die Suche danach funktioniert, ist nicht geprüft.
- Echtes Android 9.

**Nachtrag (Online-Kachel, 18:52–18:59):** Die Kachel öffnet
`https://trashygecko.onrender.com/connector` (Ziel über die WebView-Debug-Schnittstelle
ausgelesen). Zurück führte vorher zuerst auf die Connector-Startseite und erst mit dem
zweiten Druck auf den nativen Start; `ConnectorActivity` (453 → **458** Zeilen) merkt sich
jetzt in `vonAussen`, ob sie über eine Kachel geöffnet wurde, stoppt das lokale Netz und
schließt dann. Ohne Netz zeigt der WebView die englische Chromium-Fehlerseite
(`net::ERR_INTERNET_DISCONNECTED`), die Statuszeile bleibt bei „bestehender Server“ – ein
deutscher Fehlertext fehlt noch.

**Nächster konkreter Arbeitsschritt**
1. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.
2. Danach Statusanzeige mit Serverprüfung und Einstellungen.

---

## 23.11.2026 · Etappe 3, Schritt 7: native Startansicht, Zustandssicherung vereinfacht

**Warum:** Der Start der App war wieder eine Webseite (`app.html` im WebView) – genau der
Eindruck, den D-12 vermeiden will. Außerdem hatte die Zustandssicherung aus Schritt 6 einen
Fehler: `onCreate` zeichnete die Liste **vor** `stelleWieder`. Wurde nur ein Merkmal ohne
Suchtext gesichert, kam die Liste ungefiltert zurück, während der Chip schon gold war.

**Was gemacht wurde**
- `MainActivity` (134 → **193** Zeilen): kein WebView mehr. Marke, Titel, Unterzeile und
  vier Kacheln (Spiele, Lokales Netz, Online-Räume, Extras) als Views, im Farbbild von
  `start/app.html`; jede Kachel öffnet `SpieleActivity` bzw. `ConnectorActivity` mit
  `EXTRA_INITIAL_SCREEN = local | online | extras`.
- Kopfzeile zeigt den **Netzstatus des Geräts**: „offline spielbar“, „WLAN · online
  bereit“, „Mobilfunk · online bereit“ (neu gelesen in `onResume`). Er sagt nur, ob ein
  Netz da ist – nicht, ob der Render-Server antwortet.
- `start/app.html` bleibt im Paket, aber nur noch als `index.html` für „zurück zum Start“
  aus einer Spielseite; die App selbst öffnet sie nicht mehr.
- `SpieleActivity` (619 → **630** Zeilen): `stelleWieder` stellt erst her und zeichnet dann
  **einmal**; während `suchfeld.setText` läuft, ist `stelleWiederLaeuft` gesetzt, damit der
  Textwächter nicht ein zweites Mal baut. Nullprüfungen im Bundle-Weg sind entfallen (die
  Ansichten existieren, wenn gesichert wird), unbekannte Merkmale werden weiter verworfen.

**Tatsächlich getestet** (Emulator Android 14, Debug-Build)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 10s` (exit 0).
- Start nativ: „TRASHYGECKO“ + „WLAN · online bereit“, Titel, vier Kacheln mit
  Beschreibung, Fußzeile; alles innerhalb 1080 × 2340, 0 `FATAL EXCEPTION`.
- Kacheln: „Extras“ → Connector-Seite „Extras“, „Lokales Netz“ → „Lokales Netzwerk“,
  Zurück führt jeweils auf den nativen Start.
- **Chip ohne Suchtext** (der vorher fehlerhafte Fall): „Mehrspieler“ → „70 von 94“, drei
  Wischgesten (Scroll 5237) → Home → `am kill` (PID 9573 → leer) → Übersicht: neuer
  Prozess 9782, Log „Zustand wiederhergestellt: „“, 1 Merkmale, Scroll 5237“,
  `Liste: 70 Zeilen in 103 ms`; `uiautomator`-Ausdruck vor/nach **identisch**, Chip
  `selected="true"`, „70 von 94“, 0 Abstürze.
- Suchtext: „a“ → „55 von 94“ → `am kill` → neuer Prozess 10044, „55 von 94“, Feld zeigt
  „a“; danach „b“ getippt → „5 von 94“ (Textwächter wieder scharf). „Zurücksetzen“ →
  „94 Spiele“, `Liste: 94 Zeilen in 60 ms`.

**Offen / nicht getestet**
- Online-Kachel (öffnet nur die Connector-Seite gegen `trashygecko.onrender.com`).
- Server-Erreichbarkeit und „erneut versuchen“ – der Netzstatus prüft nur das Gerät.
- „Extras → Homepage“ aus D-12 fehlt noch; dafür müsste die echte Website in die Assets.
- Echtes Android 9, Wiederherstellung bei anderer Bildschirmhöhe.

**Nächster konkreter Arbeitsschritt**
1. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.
2. Danach Statusanzeige mit Serverprüfung und Einstellungen.

---

## 23.11.2026 · Etappe 3, Schritt 6: Zustand überlebt den Prozess-Tod

**Warum:** Die Drehung lässt die Activity stehen (D-21), ein beendeter Prozess nicht. Bis
hierher sicherte `SpieleActivity` nichts: nach „Aktivitäten nicht behalten“ oder einem
Prozess-Tod im Hintergrund wären Suchtext, Filter und Scrollposition weg.

**Was gemacht wurde**
- `SpieleActivity` (560 → **619** Zeilen): `onSaveInstanceState` legt Suchtext
  (`suche`), die gewählten Merkmale (`merkmale`, `StringArrayList`) und
  `rollen.getScrollY()` (`scroll`) in das Bundle.
- `stelleWieder(bundle)` läuft in `onCreate` nach `ladeKatalog()` und `zeichne()`:
  Merkmale werden nur übernommen, wenn sie in `MERKMAL_FILTER` stehen (keine Chips aus
  alten Ständen), der Suchtext geht über `suchfeld.setText(...)` durch dieselbe
  `TextWatcher`-Kette wie eine Eingabe, `chipAussehen()` malt die Auswahl an.
- Die Scrollposition wird erst nach dem ersten Layout gesetzt
  (`ViewTreeObserver.addOnGlobalLayoutListener`, danach wieder abgemeldet) – ein
  `post()` davor liefe gegen eine Inhaltshöhe von 0 und würde auf 0 geklemmt.
- Ohne Bundle bleibt alles beim Anfang; die Drehung braucht den Weg nicht (D-21).

**Tatsächlich getestet** (Emulator Android 14)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 10s`.
- Zustand gesetzt: Suche „a“, Chip „Mehrspieler“ `[GEWÄHLT]`, „37 von 94“, Liste drei
  Wischgesten nach unten (Scroll 3210).
- Home-Taste → `adb shell am kill de.trashygecko.tmahjong` → **Prozess weg** (PID 8844),
  Rückkehr über die Übersicht (Recents): **neuer Prozess 9354**, neuer `onCreate`
  („Katalog: 94 Spiele“ im Log), Logzeile
  „Zustand wiederhergestellt: „a“, 1 Merkmale, Scroll 3210“.
- Danach dieselben Zeilen an denselben Stellen wie vorher: „Risiko – Welt erobern“
  `[83,2268][997,2327]`, „Halma“ `[83,1882]`, „Go“ `[83,1689]`, „Kalaha“ `[83,2075]`;
  Zähler „37 von 94“, „Mehrspieler“ `selected="true"`, Suchfeld „a“.
- Weiterfilterbar: „Zurücksetzen“ → Suchfeld leer, kein Chip gewählt, „94 Spiele“
  (`Liste: 94 Zeilen in 23 ms`). 0 `FATAL EXCEPTION`.

**Offen / nicht getestet**
- Echtes Android 9, Online-Modus.
- Scrollposition auf einem Gerät mit anderer Bildschirmhöhe (der gesicherte Wert wird
  unverändert gesetzt; bei kleinerem Bildschirm klemmt die Ansicht selbst ab).

**Nächster konkreter Arbeitsschritt**
1. Native Startansicht (D-12) und Verbindungsstatus.
2. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.

---

## 23.11.2026 · Etappe 3, Schritt 5: Drehung – Zustand bleibt, Chips brechen neu um

**Warum:** Prüfen, ob Suchtext, Filter-Chips und Scrollposition eine Drehung überstehen.

**Befund**
- Die `SpieleActivity` wird bei Drehung **nicht neu gebaut** (`configChanges` im Manifest,
  D-21). Suchtext, gewählte Merkmale und Scrollposition bleiben dadurch von selbst erhalten.
- **Aber die Chip-Leiste wurde nur einmal umbrochen**, mit der Breite vom Start. Im
  Querformat (2296 px) passen alle sieben Chips in eine Zeile; nach dem Zurückdrehen auf
  1080 px blieb diese eine Zeile stehen: „Lokal am Gerät“ wurde auf 74 px zusammengedrückt
  (Bounds `[940,444][1014,1082]`, Buchstaben übereinander), „Online“, „Gegen Computer“ und
  „Highscore“ lagen ganz außerhalb des Bildes – diese Filter waren nicht mehr erreichbar,
  und die Liste begann erst bei y = 1115.

**Was gemacht wurde**
- `SpieleActivity` (541 → **560** Zeilen): `onConfigurationChanged` baut die Chip-Leiste
  mit der neuen Anzeigebreite neu (`baueChips()`) und stellt die Scrollposition wieder her
  (`rollen.post(scrollTo)`). Die Auswahl steht in `aktiveMerkmale`, `chipAussehen()` malt
  sie danach wieder an; `WebBasis.vollbild` wird erneut gesetzt.
- `baueChips()` rechnet die verfügbare Breite jetzt immer aus
  `getResources().getDisplayMetrics().widthPixels - dp(32)` – die Breite der Ansicht ist im
  Callback noch die alte.

**Tatsächlich getestet** (Emulator Android 14, echte Drehung per `adb emu rotate`)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 9s`.
- Hochkant (1080 × 2340) → quer (2340 × 1080) → hochkant: Chips einmal in einer Zeile
  (letzter Chip endet bei 2080 von 2160), danach wieder drei Zeilen innerhalb 1036 px,
  alle sieben erreichbar.
- Unverändert über beide Drehungen: Zähler „37 von 94“, Suchfeld „a“, „Mehrspieler“
  `[GEWÄHLT]`, oberste sichtbare Zeile „Risiko – Welt erobern“. 0 `FATAL EXCEPTION`.

**Offen / nicht getestet**
- Echtes Android 9, Online-Modus; Drehung mit geöffneter Tastatur nicht eigens geprüft.
- Prozess-Tod (Activity wird wirklich neu erstellt): Suchtext und Chips wären weg, weil
  die Ansichten keine IDs haben und `onSaveInstanceState` nichts sichert.

**Nächster konkreter Arbeitsschritt**
1. `onSaveInstanceState` für Suche, Chips und Scrollposition (deckt Prozess-Tod ab).
2. Native Startansicht (D-12) und Verbindungsstatus.
3. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.

---

## 23.11.2026 · Etappe 3, Schritt 4: kein zweiter Einstieg zur Spieleliste

**Warum:** Die Connector-Karte „Alle Spiele (offline)“ lag auf der Connector-Startseite,
die App-Startseite öffnet den Connector aber auf „Lokales Netzwerk“ – man kam nur über den
Kopfknopf „Start“ daran vorbei. Außerdem hat die Startseite die Kachel „Spiele“, die
dieselbe native Liste öffnet.

**Entscheidung (D-20):** Die Karte ist **entfernt**, nicht verschoben. Ein zweiter Einstieg
auf dem Lokal-Bildschirm hätte dieselbe Liste ein zweites Mal angeboten, mitten in einem
Bildschirm, der Räume und Chat koordiniert (dort spielbar ist nur der Galgen). Damit gibt
es genau einen Weg zu den Spielen: die Kachel „Spiele“ auf der Startseite.

**Was gemacht wurde**
- `ConnectorActivity` (461 → **442** Zeilen): Karte „Alle Spiele (offline)“ samt
  `spieleAnzahl()` entfernt (die Hilfsmethode hatte keinen anderen Aufrufer). Die
  Connector-Startseite zeigt nur noch „Lokales Netzwerk“ und „Online verbinden“.
- Keine Änderung an `SpieleActivity`, `WebNavigation`, Manifest, Spielen oder Version.

**Tatsächlich getestet** (Emulator Android 14)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 9s`.
- Connector-Startseite: nur noch zwei Karten, kein „Alle Spiele (offline)“ mehr.
- Lokal-Bildschirm (dort landet man von der Startseite): unverändert – Räume, Chat,
  „Galgenmännchen im Raum“.
- Startseite → Kachel „Spiele“ → `SpieleActivity` („94 Spiele“, Filterleiste). 0 FATAL.

**Offen / nicht getestet**
- Echtes Android 9, Rotation, Online-Modus.

**Nächster konkreter Arbeitsschritt**
1. Native Startansicht (D-12) und Verbindungsstatus.
2. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.

---

## 23.11.2026 · Etappe 3, Schritt 3: Merkmalsfilter, Connector-Karte geprüft

**Warum:** Die native Liste konnte nur per Text suchen; der HTML-Katalog filtert zusätzlich
nach Merkmalen. Und die Connector-Karte „Alle Spiele (offline)“ war nach dem Umbau noch
nicht angesehen worden.

**Was gemacht wurde**
- `SpieleActivity` (363 → **541** Zeilen): Filterleiste aus `katalog.json`.
  - Chips wie im HTML-Katalog: Einzelspieler, Mehrspieler, Lokal am Gerät, Online,
    Gegen Computer, Highscore, Controller. Angeboten wird nur, was im Katalog vorkommt –
    `controller` fehlt deshalb (in den 94 Einträgen nicht vergeben, Stand heute).
  - Mehrere Chips sind **UND**-verknüpft, Suche und Chips wirken zusammen: dieselbe Regel
    wie `passt(s, q)` in `spiele.html`.
  - „Zurücksetzen“ leert Suche und Chips und bleibt immer sichtbar, damit die Leiste beim
    Filtern nicht springt.
  - Kategorieköpfe zeigen bei aktivem Filter „x / gesamt“, der Zähler „x von 94“.
  - Leere Treffermenge: eigener Hinweis statt leerer Fläche.
- Chips **brechen um** (drei Zeilen bei 1080 px) statt seitlich zu scrollen; sonst lägen
  „Gegen Computer“ und „Highscore“ außerhalb des Sichtbereichs.
- Zeilenansichten werden einmal gebaut und beim Filtern nur umgehängt (`zeilen`-Cache);
  nach jedem Filterwechsel steht die Liste wieder oben.
- `ConnectorActivity` (443 → **461** Zeilen): Die Karte „Alle Spiele (offline)“ sagt nicht
  mehr „dazu die Startseite“ – die öffnet die native Liste nicht –, sondern beschreibt die
  Liste. Die Zahl kommt aus `katalog.json` (`spieleAnzahl()`), ist also nicht mehr fest
  eingetippt.

**Tatsächlich getestet** (Emulator Android 14, Details im Prüfbericht)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 10s`.
- „Mehrspieler“ → „70 von 94“, „Kartenspiele · 10 / 11“; zusätzlich „Gegen Computer“ →
  „5 von 94“ (UND-Verknüpfung); „Zurücksetzen“ → 94 Spiele, Suchfeld leer.
- Suche „zzz“ → „0 von 94“ und der Hinweis auf das Zurücksetzen.
- Filterwechsel laut Log (`Liste: … Zeilen in … ms`): 94 Zeilen 31 ms, 23 Zeilen 39 ms,
  5 Zeilen 14 ms; erster Aufbau 82 ms.
- Scrollen mit `dumpsys gfxinfo`: 94 Zeilen 36 ms Median / 57 ms 90er-Perzentil (33 %
  Janky), 23 Zeilen 40 ms / 61 ms (39 % Janky) – die Zahl der Einträge ändert die
  Bildzeit nicht. Der Emulator rendert mit `ro.hardware.egl=emulation` (Software-GL).
- Connector-Karte: Startseite → „Lokales Netz“ → im Connector „Start“ → Karte gelesen
  („94 Spiele, ohne Internet nutzbar“) → „Spiele öffnen“ → `SpieleActivity` → Zurück →
  `ConnectorActivity`. 0 `FATAL EXCEPTION`.

**Offen / nicht getestet**
- Echtes Android 9 (Huawei), Rotation, Online-Modus.
- Flüssigkeit auf echter Hardware: die Emulatorzahlen sind softwaregerendert.
- Die Karte liegt auf der Connector-Startseite (Kopfknopf „Start“); die App-Startseite
  öffnet den Connector direkt auf „Lokales Netzwerk“. Dorthin ist es ein Umweg.

**Nächster konkreter Arbeitsschritt**
1. Entscheiden, ob die Karte „Alle Spiele (offline)“ auf der Connector-Startseite bleibt
   oder auf den Lokal-Bildschirm wandert, wo man von der App-Startseite landet.
2. Native Startansicht (D-12) und Verbindungsstatus.
3. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.

---

## 23.11.2026 · Etappe 3, Schritt 2: ein Routing, native Spieleliste, Zähler

**Warum:** Nachkontrolle mit frischem Blick – und der nächste Schritt zum Zielbild.

**Befund der Nachkontrolle**
- Das URL-Routing stand **doppelt** da (`MainActivity.routeNavigation` und
  `GameActivity.leiteUm`) und war schon auseinandergelaufen: der Online-Sonderfall
  (`onlineGamePage`) gab es nur in `MainActivity`, die Fernbedienungs-Regel prüfte dort
  zusätzlich `/WLAN-Fernbedienung/index.html`. Dazu zwei Kopien der WebView-Einstellungen
  und des Konsolenprotokolls.
- Die Spielzahlen waren veraltet: `start/app.html` sagte „93 Spiele“, die Connector-Karte
  „91 Spiele“, im Katalog stehen **94**.

**Was gemacht wurde**
- Neu `WebNavigation.java` (142 Zeilen): **eine** Regel für alle Activities
  (`pruefe` → `BLEIBT | KATALOG | SPIEL | KONNECTOR_LOKAL/EXTRAS/ONLINE | BROWSER`,
  `starte` führt aus). Spielseiten werden nur umgeleitet, wenn der Aufrufer es erlaubt –
  in einer Spielseite bleiben Links im selben WebView.
- `MainActivity` (267 → **134** Zeilen): nur noch Startansicht. Entfallen sind `TGApp`,
  `injectGameModeChooser`, der Modus-Dialog, `safeGamePath`, `onlineGamePage` und
  `EXTRA_START_PATH` (der Katalog ist nativ, die Brücke hatte keinen Aufrufer mehr).
- `GameActivity` (266 → **219** Zeilen): Routing nur noch über `WebNavigation`,
  Online-Adresse aus `WebBasis`.
- Neu `SpieleActivity.java` (363 Zeilen): **native Spieleliste** aus `assets/katalog.json`
  (Suche, Kategorien mit Anzahl, Merkmale, Dialog „Lokal / Online“, startet `GameActivity`).
- `WebBasis`: `ONLINE_HOST`/`ONLINE_ROOT` und `protokoll(tag)` (Konsolenmeldungen).
- Zahlen: `app.html` 93 → **94**, Connector-Karte 91 → **94** und öffnet jetzt
  `SpieleActivity` statt der HTML-Seite.
- `AndroidManifest.xml`: `SpieleActivity` (`exported=false`, `fullSensor`).

**Tatsächlich getestet** (Emulator Android 14, Details im Prüfbericht)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 16s`.
- Start auf `app.html` mit „94 Spiele“; Tipp auf „Spiele“ → `SpieleActivity`,
  Log `Katalog: 94 Spiele, 9 Kategorien`.
- Suche „kniffel“ → „1 von 94“, nur Würfelspiele · Kniffel.
- „Solitär“ (nur Solo) → sofort `GameActivity`; „Baccara“ (Online) → Dialog →
  „Lokal / offline“ → `GameActivity`; Zurück führt beide Male zur Liste. 0 FATAL.

**Offen / nicht getestet**
- HTML-Katalog wird von der App nicht mehr verlinkt (bleibt im Paket und bleibt Quelle).
- Filter nach Merkmalen (nur Textsuche), Online-Modus, Rotation, echtes Android 9.
- Gerät A unverändert (1.15, release-signiert).

**Nächster konkreter Arbeitsschritt**
1. Feinschliff der Liste: Merkmals-Filter wie im HTML-Katalog, zuletzt gespielte Spiele.
2. Native Startansicht (D-12) und Verbindungsstatus.
3. `keystore.properties` wiederherstellen → release bauen → Gerätetest.

---

## 23.11.2026 · Etappe 3, Schritt 1: Spielseiten als eigene Activity, Katalog als JSON

**Warum:** Spiele lagen im selben WebView wie Startseite und Katalog – kein eigener
Zurück-Weg, Katalogzustand ging beim Spielen verloren, und es gab keine maschinenlesbare
Spieleliste für eine spätere native Ansicht.

**Was gemacht wurde**
- Neu `GameActivity.java` (266 Zeilen): zeigt genau eine Spielseite, lokal aus den Assets
  (`EXTRA_PFAD`, `EXTRA_ONLINE`) oder vom Render-Server; eigenes URL-Routing
  (`/lokal`, `/extras`, `/WLAN-Fernbedienung` → `ConnectorActivity`), Mahjong-Brücke,
  Zurück = `goBack()` sonst `finish()`, `onRenderProcessGone` wie bisher.
- Neu `WebBasis.java` (68 Zeilen): gemeinsame WebView-Einstellungen (`LOAD_NO_CACHE`,
  DOM-Speicher, kein Dateizugriff), Vollbild und die Pfadprüfung – damit Start- und
  Spielansicht nicht auseinanderlaufen.
- `MainActivity.java` (296 → 267 Zeilen): behält `app.html` und den Katalog, gibt
  Spielklicks an `GameActivity` ab (auch der Dialog „Lokal / Offline“ / „Online-Spiel“),
  Mahjong-Brücke und doppelte Vollbild-/WebView-Einstellungen sind entfallen.
- `AndroidManifest.xml`: `GameActivity` (`exported=false`, `fullSensor`, `singleTask`,
  `configChanges` wie `MainActivity`).
- `app/build.gradle`: neue Aufgabe `katalogBauen` (hängt an `preBuild`) liest den fertigen
  Katalog `www/spiele/spiele.html` und schreibt `src/main/assets/katalog.json`
  (94 Spiele, 9 Kategorien, Merkmale inkl. `multi`/`online` aus `multiplayer-register.js`).
  Bricht ab, wenn keine Spiele oder eine genannte Datei fehlt.
- `.gitignore`: `app/src/main/assets/katalog.json` (erzeugt, nicht eingecheckt).

**Verhalten:** unverändert für den Nutzer – gleiche Seiten, gleiche Dialoge, gleiches
Routing. Neu ist nur, dass eine Spielseite in einer zweiten Activity liegt.

**Tatsächlich getestet** (Details im Prüfbericht):
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 17s`, Debug-APK 11.377.723 B.
- `katalog.json` im APK (`assets/katalog.json`, 33.896 B), 94 Spiele, keine doppelten Pfade.
- Emulator Android 14: Start landet auf `app.html`, Tipp auf „Spiele“ öffnet den Katalog
  (94 Spiele in 6 Kategorien), „Baccara → Lokal / offline“ öffnet `GameActivity`,
  Zurück führt mit erhaltener Kategorie und Scrollposition zum Katalog.
- Mahjong öffnet ebenfalls in `GameActivity`, Brett wird korrekt gezeichnet.

**Offen / nicht getestet**
- `SPIEL VERLASSEN` im Mahjong-Menü nicht gedrückt (Brücke injiziert wie vorher, aber nicht
  ausgelöst); „ZURÜCK ZUM CONNECTOR“ im Mahjong-Menü nicht geprüft.
- Kein Test auf dem echten Android-9-Gerät (Debug-APK passt nicht über die
  release-signierte 1.15, siehe D-14).
- Rotation, Online-Modus, LAN, Hotspot: unverändert offen.
- Zahlen im Text sind veraltet: `start/app.html` sagt „93 Spiele“, `ConnectorActivity`
  „91 Spiele“, der Katalog enthält 94. Noch nicht angefasst (nur Text).

**Nächster konkreter Arbeitsschritt**
1. Native Spieleliste aus `katalog.json` (RecyclerView oder ListView, Suche/Filter),
   damit der Katalog nicht mehr HTML sein muss.
2. Danach native Startansicht (D-12) und Verbindungsstatus.
3. `keystore.properties` wiederherstellen → release-signierter Build → Gerätetest.

---

## 23.11.2026 · Etappe 1: Bestandsaufnahme und Dokumentation

**Warum:** Die App wirkt wie eine verkleinerte Webseite; vor jedem Umbau musste klar
werden, was tatsächlich existiert, was fehlt und warum.

**Was gemacht wurde**
- Beide Android-Projekte gelesen: `spiele/android` (Gradle, `de.trashygecko.tmahjong`,
  1.15 / Code 16, minSdk 21, targetSdk 34) und `apps/t-spiele` (Altlast, `bauen.sh`, 1.4).
- Alle Java-Dateien der App gesichtet (2.445 Zeilen): `MainActivity` (WebView-Wirt der
  Homepage), `ConnectorActivity` (native Connector-Ansicht), `LocalNetworkCoordinator`
  (NSD + TCP-Relay), `GalgenLokal`, `RemoteControlActivity`, `remote/*`, `HomepageAssets`.
- Bauweg verstanden: `homepageKopieren` → `assets/www` (353 Dateien, 108 HTML-Seiten),
  `verifyHomepageGames`, `packageTrashyGeckoApk`.
- Umgebung geprüft: echtes Android-9-Gerät (Huawei FIG-LX1) und Emulator (Android 14)
  sind per `adb` erreichbar; Werkzeuge (JDK 17, SDK 34, Gradle 8.7) liegen in
  `Documents\arbeit\werkzeuge`.
- Echter Build ausgeführt: `:app:assembleDebug` → `BUILD SUCCESSFUL in 46s`,
  Debug-APK 11.346.443 B.
- Neue Dokumente angelegt: `AGENTS.md`, `ANDROID-ARCHITEKTUR.md`, `ENTSCHEIDUNGEN.md`,
  `ENTWICKLUNGSSTAND.md`, `PRUEFBERICHT-ANDROID.md`.

**Bearbeitete Dateien**
- nur Dokumentation (5 neue Markdown-Dateien im Projektwurzelverzeichnis)
- keine Änderung an Code, Gradle, Manifest oder Assets

**Was jetzt funktioniert (Ist-Stand, nicht neu getestet)**
- Die App baut über Gradle und enthält die komplette Spieleabteilung als Assets.
- Start über `app.html` (HTML) mit vier Kacheln; Katalog als `spiele/spiele.html`;
  Spiele öffnen im selben WebView.
- Nativer Connector-Bereich (`ConnectorActivity`): Start, Lokales Netz (Chat + Galgen),
  Online-Connector (WebView auf onrender.com), Extras mit Fernbedienung.
- Lokales Netz technisch vorhanden (NSD + Relay, 8-stelliger Code, Wiederverbindung).

**Erkannte Hauptprobleme**
1. Start und Spieleliste sind HTML → „verkleinerte Webseite“.
2. Zwei getrennte Welten (WebView-Home vs. native Connector-Ansicht) ohne gemeinsame
   Navigation und ohne Verbindungsstatus.
3. Nur Galgen läuft im lokalen Netz; alles andere braucht den Online-Server.
4. `keystore.properties` fehlt → derzeit kein release-signierter Build möglich
   (Keystore liegt außerhalb, Passwort in `passwort.txt` daneben).
5. `apps/t-spiele` (Zwei-Knopf-App) wird mit „der App“ verwechselt.

**Offen**
- Entscheidung des Nutzers zum Zielbild (siehe `ANDROID-ARCHITEKTUR.md`, Abschnitte 4
  und 5, sowie D-12 in `ENTSCHEIDUNGEN.md`): bleibt der Start nativ (empfohlen) und die
  echte Website unter Extras?
- Danach Umsetzung: `GameActivity` herauslösen, `katalog.json` im Build erzeugen,
  native Oberfläche (Tabs, Start, Spieleliste, Einstellungen), Netzwerkstatus.
- `keystore.properties` wiederherstellen, damit `TrashyGecko.apk` release-signiert
  entsteht (ohne die Datei ist kein Update über 1.15 möglich).
- Gerätetests (siehe `PRUEFBERICHT-ANDROID.md`, Abschnitt 4) – in dieser Runde keine.

**Nächster konkreter Arbeitsschritt**
1. `GameActivity` aus `MainActivity` herauslösen (WebView + Asset-Loader + Zurück),
   `MainActivity` behält vorerst `app.html` als Start. Danach bauen und auf dem Emulator
   starten. → Fundament für alles Weitere, ohne sichtbare Änderung.

**Tatsächlich durchgeführte Tests**
- `:app:assembleDebug` erfolgreich (46 s, siehe Prüfbericht).
- `adb devices`: Gerät A und Emulator B erkannt; Version auf Gerät A abgefragt (1.15/16).
- **Keine** Installations-, Start-, Navigations- oder Netzwerktests in dieser Runde.

---

## Vorheriger Stand (aus `spiele/android/README.md`, nicht neu geprüft)

- **1.15 / Code 16** (22.11.2026): letzte Fassung; enthielt die Moderator-/Quiz-Arbeit
  nicht (die betrifft nur den Connector und die Web-Seite).
- **1.14 / Code 15** (21.11.2026): Fernbedienung ausschließlich unter Extras, Startseite
  mit vier Kacheln, `LOAD_NO_CACHE`, `onRenderProcessGone`, Release-Signatur eingeführt.
- **1.13 und früher**: Umbau von „nur Mahjong“ zur Spiele-App, lokale Räume, Galgen im
  LAN, Fernbedienung.

---

## Scherzshow: Zeitsteuerung (24.11.2026)

- **Antwortzeit 30 s in beiden Modi** (`quizZeitTempo` und `quizZeitPunkte` in
  `spiele/connector/quiz.go`, vorher 20 s bzw. 15 s). Die Zeit ist eine Höchstzeit;
  angezeigt wird sie aus `zeitMs`, die Spielseite hat keine eigenen Sekunden.
- **Sofortiger Rundenabschluss**: `quizAntwort` wertet aus, sobald `quizAlleDran()`
  meldet, dass alle Spieler geantwortet haben – es wird nicht auf die Restzeit
  gewartet. Bestand schon, ist jetzt durch Tests festgenagelt; ebenso der Abgang
  eines Spielers (`quizRemoveSeat`) und die Frist als Obergrenze.
- **Spielseite** (`spiele/quizshow/quizshow.html`): Nachladen und eigene Aktion haben
  getrennte Sperren (`busy` / `sendet`). Vorher verschluckte ein laufendes Nachladen
  (alle 900 ms) einen Klick auf eine Antwort – der Spieler stand dann als „hat nicht
  geantwortet“ da und die Runde lief in die Höchstzeit. Wer geantwortet hat, sieht
  Frage, alle vier Antworten und den Quizmaster-Kommentar weiter, kann aber keine
  zweite Antwort abgeben.
- Auflösungsdauer unverändert 7 s (`quizPauseZeit`), Punkte, Fragen, Regeln und
  Gestaltung unverändert.
- Binaries neu gebaut (`trashygecko-connector.exe`, `trashygecko-connector-online.exe`,
  `trashygecko-connector-online-linux`), `connector-online` über `sync.py` abgeglichen,
  APK neu gebaut (14.677.021 B). Sicherung: `Homepage/sicherung/quizzeit-2026-11-24/`.

## Spielehalle ohne Fuß (24.11.2026)

- Der Fußbereich unter dem Spielekatalog ist ganz weg: Links (Startseite, Highlights,
  Werkzeuge & Extras), Datum/Uhrzeit und das Gameboy-Zeichen aus Text. Mit dem Rest der
  Seite hatte er nichts mehr zu tun.
- Entfernt: `<footer class="fuss">` samt Inhalt und das 1-s-Skript für Datum/Uhrzeit in
  `spiele/spiele.html`; die Regeln `.fuss`, `.fuss-links`, `#local-date-time`, `.gameboy`
  (auch die Handy-Regel `.gameboy{display:none}`) in `spiele/halle.css`. Reine
  Anzeige – Katalog, Suche, Filter, Online-Halle unverändert.
- Geprüft (Chromium 390×844 und 1280×800): 0 Elemente `.fuss`/`footer`/`.gameboy`,
  letztes Element ist die Hallen-Auswahl, Seite 1647 px bzw. 1030 px hoch, kein
  Querscrollen, keine Konsolenfehler, 10 Karten wie zuvor.
- Beide Ablagen und die Android-Assets byte-gleich; APK neu 14.676.443 B.
