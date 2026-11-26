# Entwicklungsstand – TrashyGecko Android

Lebendes Dokument. Nach **jeder** abgeschlossenen Arbeit fortschreiben (oben das Neueste).
Was getestet wurde, gehört zusätzlich in [`PRUEFBERICHT-ANDROID.md`](PRUEFBERICHT-ANDROID.md).

---

## 26.11.2026 · Chat am Handy tiefer, Smiley-Auswahl passt ins Bild

**Fehler:** Im Handy-Layout stand `#chatbox` per `order:1` direkt unter dem Kopf
(Chatbeginn bei 46 px). Das Smiley-Fenster hing fest über der Eingabezeile
(`position:absolute; bottom:calc(100% + 6px)`) – weil die Eingabezeile ganz oben
lag, öffnete es nach oben aus dem Bild: hochkant `top −2 px`, quer `−24 px`, die
obersten Kacheln waren abgeschnitten und nicht antippbar.

**Geändert** (`spiele/connector/connector.html`, byte-gleich in `connector-online/`):
Reihenfolge im Handy-Block jetzt Spieler (1), Einstellungen (2), Chat (3); dazu
`smileyPanelPlatzieren()`, das die Eingabezeile und den sichtbaren Bereich
(`visualViewport`) misst und das Fenster über der Eingabe (ab 240 px Platz), sonst
darunter, sonst über die ganze Höhe öffnet – Lage und Höhe auf 8 px Rand begrenzt,
`position:fixed` nur in diesem Bereich, neu gerechnet bei Drehen, Tastatur,
Adressleiste und Scrollen. Solange es offen ist, ist der schwebende Chat-Knopf aus.
Am PC (≥1024 px) steigt die Rechnung sofort aus; dort bleibt alles beim CSS.

**Gemessen** (Chromium-Emulation, echter Connector 8767): Chat beginnt statt bei
46 px bei **320 px**; Fenster hochkant **8 … 306** (vorher `−2 … 338`), quer
**8 … 348** (vorher `−24 … 171`), bei auf 560 px verengtem Fenster (Tastatur)
**227 … 552**, mit offener Partie (Chat 287 px) **274 … 614** – überall `passt: true`.

Oberste und unterste Kachel angetippt → Auswahl übernommen, Nachricht mit Bild
gesendet (`naturalWidth > 0`), Escape und Klick daneben schließen, Drehen bei
offenem Fenster rechnet neu (298 → 340 → 298), Partie öffnen/schließen unverändert,
keine Seitenfehler. **PC 1280 × 800 unverändert:** Chat 147 … 800 (653 px), Fenster
381 … 721 (340 px). Grenze: Geräte-Emulation statt echtem Telefon, Tastatur über
kleinere Fensterhöhe nachgestellt.

## 25.11.2026 · Snake wieder Einzelspieler, Würfelspiele als Mehrspieler

**Snake:** Die Spielhalle öffnet jede Seite mit `?room=…&id=…`; Snake schaltete daraufhin
in die Zwei-Spieler-Lobby um – der Einzelspieler war aus der Halle nicht erreichbar. Das
Duell gilt jetzt nur mit `&art=duell` (der Connector setzt das an seinen vier
Snake-Aufrufen, `spielAdresse(path, art)`); ohne den Zusatz startet Snake allein, bei
Raum+Kennung steht zusätzlich „Duell im Raum“ im Menü.
Geprüft (headless Chromium, ohne Testhaken): Startbild „SNAKE · EINZELSPIELER“, HUD nur
Punkte/Länge/Tempo/Rekord, kein Mitspielertext im DOM, Game Over nach 3,9 s, „Nochmal“
setzt auf 0 Punkte / Länge 3, zweiter Lauf läuft. Apfel: 10 Punkte, Länge 3 → 4, Rekord
10 im `localStorage`. Duell im Raum TG-33D5 gegen Bot „Mumpitz“ (`action=bot`): Menü
ausgeblendet, Stand „Namenloser 2 0 : Mumpitz 1“, Runde 1 abgeschlossen.

**Würfelspiele:** Alle 25 (7 unter `brettspiele/`, 18 in `wuerfelspiele/`) waren als
`solo` geführt und standen damit unter „Einzelspieler“, obwohl sie einen Tisch für 2–6
Personen und über den Connector Computergegner haben. Merkmale jetzt
`lokal zwei computer`, auf den Hallenkarten „Solo“ → „Bots“.
Gemessen im Connector: „Einzelspieler“ 52 → **36** Spiele (0 Würfelspiele),
„Mehrspieler“ **72** mit allen 25; je Würfelkarte die Knöpfe „Am Gerät spielen“,
„Gegen Computer“, „Online spielen“. Lokal zu zweit: Beelzebub, 5 Runden,
**40 : 20** mit Siegmeldung. Gegen Computer: Beelzebub mit Bot „Vroni“, Bot würfelt und
punktet (27). `go test ./...` in `spiele/connector` grün (229 s, u. a.
`bot_wuerfel_test.go` über alle 18 Spiele, `bot_snake_test.go`).

**Zahlen und Wege:** `spiele.html` lädt jetzt `spieldaten.js` und setzt „N Spiele ·
N Kategorien“ daraus (**95 / 6**, vorher 94); Werkzeugkachel und
`android/start/app.html` ebenfalls 95. `python spiele/pruefe-spieldaten.py` läuft ohne
Abweichung (vorher 3). Keine CDNs, keine neuen Bibliotheken.

**Noch offen:** Das verlinkte APK ist der alte Bau; `assets/www` und `katalog.json`
entstehen beim Bauen und wurden nicht von Hand geändert – die App zeigt die neue
Zuordnung erst nach einem Neubau.

## 24.11.2026 · Schritt 7: Cache- und Ausfallverhalten geprüft

Geprüft auf Emulator B (Android 14) mit dem Debug-Bau 1.17/18, Flugmodus
(`Active default network: none`), echte Tipps über `uiautomator`, WebView-Inhalt über
DevTools. Kein Quellcode geändert – nur Messungen; Testkatalog und Testseite danach
zurückgebaut (Katalog `fa50d261…`, Mahjong-Seite `1eaab258…`).

- **Ohne Netz:** Start 722 ms (`am start -W`, COLD), Spielhalle 94 Spiele / 6 Kategorien
  in 54 ms, lokales Spiel (Mahjong) 3,52 s nach dem Tipp geladen, 144 Bilder / 0 kaputt,
  `readyState complete`. Einziges Log: `www/favicon.ico` fehlt (404). Ein Online-Spiel
  ohne Netz zeigt die englische Chromium-Seite „Webpage not available“
  (`net::ERR_INTERNET_DISCONNECTED` −2); die App schreibt danach trotzdem
  „Spielseite geladen“.
- **Veralteter Katalog:** 95 Einträge – einer auf eine fehlende Seite, einer mit
  unbekannter Kategorie – werden ungeprüft gelistet. Fehlende Seite → 404 aus
  `HomepageAssets`, Chromium-Seite `net::ERR_HTTP_RESPONSE_CODE_FAILURE`, keine
  App-Meldung und kein Weg außer „Zurück“. Unbekannte Kategorie → Kopf „Weitere · 1“.
  Unlesbares JSON → „Der Spielekatalog fehlt im App-Paket. Bitte die App neu
  installieren.“, kein Absturz.
- **Liegengebliebene Seiten:** 303 HTML-Seiten im Paket, 105 unter `spiele/`, 94 im
  Katalog; die 11 übrigen sind Werkzeuge/Infoseiten und alle verlinkt (0 verwaist). Eine
  künstlich abgelegte Alt-Seite (60 B) verschwindet beim nächsten `homepageKopieren`.
  Nach einem App-Update erscheint ein geänderter Seitentitel sofort – kein alter
  Cache-Stand.
- **Cache:** nach dem Leeren und einem Spiel + Halle 324 KB in 22 Dateien, davon 216 KB
  V8-Code-Cache; keine Seitenkopien im HTTP-Cache.

Offen (nur festgestellt, nicht geändert): Für Online-Spiele ohne Netz und für
Katalogeinträge ohne Seite gibt es keine deutsche Meldung in der App.

---

## 24.11.2026 · Schritt 2: Alte Zahlen und Bauskripte bereinigt

- **Spielzahlen:** Die App-Kachel in `spiele.html` sagte „91 Spiele offline“ →
  **94**. Der App-Start (`start/app.html`) sagte „65 Spiele mit mehreren“ →
  **71** (Online-Spiele; 72 ist die Vereinigung mit den lokalen
  Zwei-Personen-Spielen). `spiele/android/README.md` nannte zweimal 91 →
  **94**. `ANDROID-ARCHITEKTUR.md`: 91 → 94, 9 → **6 Kategorien** (so viele
  stehen in `spieldaten.js` und `katalog.json`). `ENTSCHEIDUNGEN.md`: 91 → 94,
  Stand 24.11.2026 mit 6 Kategorien; der alte Widerspruch bleibt als Hinweis
  („früher stand dort 91, im Katalog 94“) erhalten.
- **Alte Bauskripte aus dem App-Paket:** `gewonnen_js_bauen.py` und
  `lob_js_bauen.py` lagen als Kopien unter
  `assets/www/spiele/` und damit im APK. Entfernt; `homepageKopieren`
  schließt jetzt `**/*.py`, `**/*.sh` und `**/*.ps1` aus. Die Quellwerkzeuge
  unter `spiele/` bleiben (sie bauen `gewonnen.js`/`lob.js` aus den lokalen
  Textdateien).
- **Keine eingebettete Spieleliste mehr:** `KATEGORIEN`, `SPIELE`, `TEXTE`
  und `HIGHLIGHTS` werden nur noch in `spiele/spieldaten.js` definiert.
  `spiele.html` und beide `connector.html` laden die Datei je einmal
  (`<script src…spieldaten.js>`), enthalten keine Pfad-Zeilen der alten Liste
  (0 Treffer) und keine Inline-Definition. Gemessen im laufenden Connector:
  Halle **94 Karten**, Connector **95 Karten**, je **1** Daten-Skript,
  **0** Inline-Listen, **0** Seitenfehler.
- **Noch nicht neu gebaut:** Das vorhandene APK (17 / 1.16) enthält die zwei
  `.py`-Kopien und die alte Cabo-Zuordnung weiterhin; beides verschwindet
  erst mit dem Neubau in Schritt 4. Nichts veröffentlicht, nichts gepusht.

---
## 24.11.2026 · Spielzuordnung abgeschlossen: Cabo ist Einzelspieler

**Prüfung (alle 94 Spiele der Halle):**
- **Online (71):** Register-Eintrag, Seitenaufruf des eigenen Endpunkts (direkt oder über
  das jeweilige Modul, z. B. `wuerfelspiele/online.js`, `halma/js/main.js`,
  `risiko-fern.js`, `paare` über `API+"/paare"`), Route als `mux.HandleFunc` in
  `connector/server.go`; `go test -run Register` grün. Kein Spiel außerhalb des
  Registers ruft einen Spiel-Endpunkt auf.
- **Lokal (55):** 54 Spiele mit ausdrücklichem Zwei-Personen-Modus (Spielerzahl 2,
  „2 Spieler“, „Lokal 2–4“, geteilte Steuerung/abwechselnde Seiten); Arschloch lokal
  3–9 (Optionen 3…9), also lokal, aber nicht zu zweit.
- **Einzelspieler (56):** Solo-Modus, „1 Spieler“ oder Gegner-KI.

**Korrektur:** `Cabo` hat einen eigenen Knopf **„Gegen Computer“** (`#btnSolo` →
`tu("solo")`, Schwierigkeit Leicht/Mittel/Schwer) und war fälschlich ohne
Einzelspieler geführt. In `spieldaten.js` jetzt `solo`. Damit: **56 Einzelspieler**,
55 lokal, 54 zu zweit, 71 online, Mehrspieler 72. Gemessen: Connector 95/56/72/71/54,
Halle Solo 56, Cabo in beiden „Solo“+„Online“.

**Arschloch und Acia – kein Widerspruch, zwei verschiedene Fragen:**
- Arschloch: lokal 3–9 → „Lokal“ ja, „Zu zweit“ nein.
- Acia: 1 oder 2 Spieler → „Lokal“ ja, „Zu zweit“ ja.
- Die Halle kennzeichnet „Lokal“ (Mehrspieler am Gerät, beliebige Zahl), der Connector
  filtert „Zu zweit“ (genau zwei). In der App müssen beide Kennzeichen nebeneinander
  stehen: Acia „Lokal“ **und** „Zu zweit“, Arschloch nur „Lokal“ (plus Online).

**Nicht angefasst:** Versionsnummer (weiter 17 / 1.16), APK und App-Assets – die werden
erst beim abschließenden Neubau erzeugt und enthalten deshalb noch Cabo ohne
Einzelspieler. Nichts veröffentlicht, nichts gepusht.

---

## 24.11.2026 · APK mit den gemeinsamen Spieldaten neu gebaut

**Anlass:** Die Spielhalle in der App soll dieselben Merks zeigen wie Website und
Connector. Der Android-Bau las den Katalog aber noch aus `spiele/spiele.html`, wo die
Liste seit dem Umbau nicht mehr steht – ohne Anpassung wäre der Bau gar nicht
durchgelaufen.

**Geändert** (`android/app/build.gradle`, beide Ablagen)
- `verifyHomepageGames` und `katalogBauen` lesen die Spieleliste, Kategorien, Texte und
  Highlights aus `spiele/spieldaten.js`; die Spielzeichen (`SPIELSYMBOLE`, `IKONEN`,
  `SYMBOLREGELN`) weiter aus `spiele.html`.
- Die native Spieleliste kennt kein `zwei`: der Bau lässt es weg und ergänzt bei
  `lokal` das Merkmal `multi` (Mehrspieler am Gerät). `katalog.json` nennt als Quelle
  jetzt `spiele/spieldaten.js`.

**Gebaut** (signierte Release-APK, gleicher Schlüssel wie zuvor: CN=TrashyGecko,
SHA-256 `73019c01…`): `gradle clean packageTrashyGeckoApk`, **BUILD SUCCESSFUL in 37 s**,
94 Spiele geprüft. APK **14.669.149 B**, md5 `c41dbb0e…`, versionCode 17 / 1.16, in
beiden Ablagen als `spiele/android/tmahjong.apk`.

**Im APK geprüft**
- `assets/www/spiele/spiele.html` lädt `spieldaten.js`; `assets/www/spiele/spieldaten.js`
  enthält Arschloch `lokal` und Acia `solo lokal zwei`.
- `assets/katalog.json`: 94 Spiele, Quelle `spiele/spieldaten.js`; Arschloch
  `[lokal, multi, online]`, Acia `[solo, lokal, multi]`, Bingo `[solo, multi, online]`.
- App-Halle im Testbrowser (aus den Build-Assets): 94 Karten, Arschloch „Lokal & Online“,
  Acia „Solo“+„Lokal“, Bingo/Galgenmännchen „Solo“+„Online“, 0 Fehler, 0 Antworten ≥ 400.
- Die Installer-Kacheln (`android/*.apk`) und die Fernbedienungs-Kachel fehlen in der
  App-Halle wie bisher (der Bau entfernt sie aus der Kopie).

**Aufgeräumt:** `app/build`, `.gradle` und `build` wieder entfernt. APK nicht auf einem
Gerät installiert (kein Gerät angeschlossen), nichts veröffentlicht, nichts gepusht.

---

## 24.11.2026 · Spielarten: eine Quelle für Spielhalle und Connector

**Befund:** Die Trennung war nicht sauber. `lokal` hieß in der Halle „Mehrspieler am
selben Gerät“ (auch Arschloch ab drei), im Connector über `lokal && multi` aber „zu
zweit“. `multi` wurde dort aus dem Online-Register dazuerfunden – deshalb brauchte Acia
ein künstliches `multi`, und Arschloch musste sein `lokal` verlieren, obwohl es lokal
spielbar ist. Dieselben 94 Spiele standen außerdem zweimal (inline in `spiele.html` und
eingebettet in `connector.html`) und waren bereits auseinandergelaufen.

**Geändert**
- Neue gemeinsame Datei `spiele/spieldaten.js` (Kategorien, Spiele, Kurztexte,
  Highlights). Die Halle lädt `spieldaten.js`, der Connector `/spiele/spieldaten.js`;
  die eingebettete Kopie im Connector ist ersatzlos weg (rund 240 Zeilen).
- Merkmale eindeutig: `solo`, `lokal` (am Gerät, beliebige Spielerzahl), `zwei` (lokal
  zu zweit), `computer`, `highscore`, `controller`. `online` und `multi` sind keine
  Merkmale mehr: online kommt ausschließlich aus `multiplayer-register.js`, multi ist
  `online || zwei`.
- Arschloch: `lokal` ohne `zwei` – die Halle zeigt wieder „Lokal & Online“, der
  Connector zählt es nicht zu „zu zweit“. Acia: `solo lokal zwei`.
- Code: Connector liest `solo`/`zwei` direkt aus den Merkmalen (kein `lokal && multi`,
  kein Register-Push mehr); die Halle nutzt `s.online`, ohne `multi`/`online` in die
  Merkmale zu schreiben, und hat den „Mehrspieler“-Zweig in `modiVon` verloren.

**Gemessen** (Playwright, 1440 × 1000): Halle 94 Karten – Solo 55, nur Online 16,
Lokal & Online 54, nur Lokal 1 (Acia), Bots 5; Bingo/Galgenmännchen/Paare
finden/Wortsuche „Solo“+„Online“, Arschloch wieder „Lokal & Online“, Risiko und Domino
„Solo“+„Lokal & Online“+„Bots“; 0 Konsolenfehler, 0 Antworten ≥ 400, `spieldaten.js`
200 (20.598 B). Connector: Alle 95, Einzelspieler 55, Mehrspieler 72, Online 71, Zu
zweit 54; Acia weiter nur lokal, Arschloch nicht unter „zu zweit“, 0 Fehler.

**Nachtrag (gleicher Tag):** Die dritte Kopie der Spieleliste – die Spielhalle in den
Web-Assets der Android-App – trug noch die alten Merks (Arschloch „lokal online multi“,
Acia „solo lokal“). Sie ist jetzt ebenfalls auf die gemeinsame Datei gezogen: in beiden
Ablagen liegen `spiele.html` und `spieldaten.js` byte-gleich zur Quelle, die App-Halle
lädt `spieldaten.js` und zeigt dieselben Kennzeichen (94 Karten, 0 Fehler, 0 Antworten
≥ 400). Alte Merk-Zeilen gibt es außerhalb von Sicherungen und Build-Zwischenständen
nicht mehr; die APK selbst wurde nicht neu gebaut.

**Nicht geändert:** Navigation, Filterbedienung, Spielseiten, Server- und Go-Code.
`spieldaten.js` `b6d27ce4…`, `spiele.html` `87cc12cc…`, `connector.html` `a9bf7a94…` –
in beiden Ablagen und in den App-Assets gleich. Nichts veröffentlicht, nichts gepusht.

---

## 24.11.2026 · Spielhalle: Merks an die geprüfte Zuordnung angeglichen

**Warum:** Die Spielkarten in `spiele/spiele.html` zeigten für Bingo, Galgenmännchen,
Paare finden und Wortsuche „Lokal & Online“, obwohl deren lokaler Modus Einzelspieler ist
(der Connector war bereits korrigiert, die Halle nicht).

**Geändert** (SPIELE-Zeilen in `spiele.html`)
- `Bingo`, `Galgenmännchen`, `Paare finden`, `Wortsuche`: `lokal online multi` →
  `solo online multi`.
- `Risiko – Welt erobern` und `Domino`: `solo` ergänzt (Computergegner) – damit stimmen
  auch diese Karten mit der geprüften Zuordnung überein.
- `Arschloch` bleibt `lokal online multi` (lokal erst ab 3 Spielern, also nicht „zu
  zweit“, aber sehr wohl am Gerät), `Acia` bleibt `solo lokal` (kein Online).

**Gemessen** (Playwright, echte Seite, 1440 × 1000, 0 Konsolenfehler): Bingo,
Galgenmännchen, Paare finden, Wortsuche zeigen **„Solo“ + „Online“** (vorher
„Lokal & Online“); Risiko und Domino **„Solo“ + „Lokal & Online“ + „Bots“**; Arschloch
„Lokal & Online“, Acia „Solo“ + „Lokal“.

**Nicht geändert:** Filter, Navigation, Spielseiten, Connector. Beide Ablagen byte-gleich
(`spiele.html` md5 `9036603d…`). Die Web-Assets der Android-App
(`spiele/android/app/src/main/assets/www/spiele/spiele.html`) tragen weiterhin die alte
Kopie und wurden nicht angefasst. Nichts veröffentlicht, nichts gepusht.

---

## 24.11.2026 · Verrücktes Labyrinth: Suchgegenstände neu gezeichnet

**Warum:** Die 25 Gegenstandssymbole waren einfarbige Silhouetten in der Tintenfarbe der
Welt und saßen durch einen halben Einheitsversatz (`translate(-.5,-.5)` in
`zeichneGegenstand`) nicht mittig auf ihrem Plättchen, sondern ragten oben links heraus.
Auf dem Feld (16–42 px) war kaum zu erkennen, wonach gesucht wird.

**Geändert** (`spiele/brettspiele/verruecktes-labyrinth/verruecktes-labyrinth.html`,
beide Ablagen byte-gleich, md5 `75b21e73a879a219827f97fd55a81925`)
- Zeichenblock der Gegenstände ersetzt: jede Figur jetzt zwei- bis dreifarbig mit dicker
  dunkler Kontur – Krone, Schlüssel, Edelstein, Münze, Buch, Ring, Glocke, Kelch,
  Heiltrank, Schwert, Schild, Hammer, Kette, Laterne, Schriftrolle, Blatt, Pilz, Eichel,
  Eule, Blume, Fisch, Feder, Stern, Schneeflocke, Eiskristall. Nur große Formen, keine
  feinen Details.
- Eine Zeile: der Versatz in `zeichneGegenstand` entfernt – die drei Aufrufer übergeben
  ohnehin die linke obere Ecke, das Symbol sitzt nun mittig im Plättchen.
- **Unverändert:** Zuordnung Aufgabe → Gegenstand (alle acht Namen je Welt), Karten, Wände,
  Figuren, Regeln, Suchlogik, Spielerreihenfolge, Mehrspieler. Diff zum alten Stand: genau
  zwei Stellen, 235 Zeilen neu / 177 alt.

**Geprüft** (Chromium/Playwright, Server 8797, 1280 × 900 und 390 × 844, alle sieben Welten)
- Feldgrößen: 7 × 7 am Handy Brett 315 px → Kachel 45 px, Plättchen 27 px, Symbol 21 px;
  7 × 7 am PC Brett 630 px → Kachel 90 px, Plättchen 54 px, Symbol 42 px; 9 × 9 am Handy
  Kachel 35 px, Plättchen 21 px, Symbol 16,4 px. Aufgaben-Chips 26 px (Canvas 56 px).
- Kontrast der dunklen Kontur zum Plättchen: **7,95 : 1** ab 24 px, **3,80 : 1** bei 20 px
  (schwächster Wert, Schneeflocke), 3,34 : 1 bei 16 px.
- Je Symbol 1–7 deutlich getrennte Farbtöne (Flächenanteil ≥ 6 %), Deckung 25–76 % der
  Plättchenfläche.
- Alle sieben Welten am PC und am Handy durchgeklickt, 9 × 9 nachgeprüft: keine Konsolen-
  oder Seitenfehler, Gegenstände sitzen mittig und heben sich vom hellen Plättchen ab.
- Kontaktbogen aller 25 Symbole in 56/32/24/20/16 px:
  `arbeit/screens/labyrinth-gegenstaende/bogen-groessen.png`; Brettfotos je Welt daneben.
- Gesichert: `arbeit/sicherung/2026-11-24-labyrinth-gegenstaende/verruecktes-labyrinth.vorher.html`
  (md5 `88712e1729e9a2d69de3f74a426b73e5`).

**Offen**
- Bei 16 px (9 × 9 am Handy) werden Feder und Kette feiner; erkennbar, aber am Rand der
  Deutlichkeit. 5 × 5 und 7 × 7 sind unkritisch.
- Nicht committet, nicht veröffentlicht (Regel „nichts committen").

## 24.11.2026 · Connector: Spielarten an den Spielen selbst geprüft

**Warum:** „Mehrspieler“ und „Online-Mehrspieler“ zeigten beide 71 Spiele. Die Zuordnung
kam aus dem `merk`-Feld und der Mitgliedschaft in `multiplayer-register.js`, nicht aus
der tatsächlichen Spielintegration.

**Geprüft**
- **Online (71):** Für jedes der 71 Register-Spiele ruft die Spielseite ihren eigenen
  Endpunkt auf (`/api/<id>` bzw. `/api/wuerfel`, 18 Würfelspiele über
  `wuerfelspiele/online.js`). Jede dieser Routen existiert als `mux.HandleFunc` in
  `connector/server.go`; `go test -run Register ./...` grün. Kein Spiel außerhalb des
  Registers ruft einen Spiel-Endpunkt auf → 71 online bestätigt, keine Streichung.
- **Zu zweit (54):** nur Spiele mit ausdrücklichem lokalen Modus für zwei Personen am
  Gerät (Spielerzahl-Auswahl ab 2, „2 Spieler“, „Lokal 2–4 Spieler“, geteilte Steuerung).
  `lokal` im `merk`-Feld bedeutete teils nur „ohne Server“.
- **Einzelspieler (55):** Solo-Modus, Gegner-KI oder „1 Spieler“ vorhanden.

**Korrigiert** (TGSpieldaten in `connector.html`)
- `Acia`: `solo lokal` → `solo lokal multi` (1/2 Spieler, geteilte Tasten) – damit erstmals
  unter „Zu zweit“ und „Mehrspieler“.
- `Arschloch`: `lokal online multi` → `online multi` (lokal erst ab 3 Spielern).
- `Bingo`, `Galgenmännchen`, `Paare finden`, `Wortsuche`: `lokal online multi` →
  `solo online multi` (der lokale Modus ist Einzelspieler, keine zweite Person).
- `Risiko – Welt erobern`: `lokal multi computer` → `solo lokal multi computer`.
- `Domino`: `lokal online multi computer` → `solo lokal online multi computer`
  („Gegen Computer“).

**Gemessen** (Playwright, lokaler Connector, 1920 × 1080, 0 Konsolenfehler): Alle 95,
Einzelspieler 55, Mehrspieler 72, Online 71, Zu zweit 54. Schnittmenge Online ∩ Zu zweit
53, nur lokal `Acia`, Vereinigung 71 + 54 − 53 = 72 = Mehrspieler. Jede Ansicht ohne
Doppelungen.

**Nicht geändert:** Navigation, Spielseiten, Server- und Go-Spielmechanik. Alle vier
`connector.html`-Kopien byte-gleich. Nichts veröffentlicht, nichts gepusht.

---

## 24.11.2026 · Connector: drei Hauptfilter statt doppelter Reiter

**Warum:** Rechts oben standen zwei Reiter („Einzelspieler“, „Multiplayer“) und darunter
noch einmal vier Filter („Alle“, „Einzelspieler“, „Online-Mehrspieler“, „Zu zweit am
Gerät“). „Einzelspieler“ kam doppelt vor, „Multiplayer“/„Mehrspieler“ waren zwei
Bezeichnungen für dasselbe.

**Was gemacht wurde**
- Die Reiterzeile samt Solo-Bereich (`#modusSolo`, `#modusMulti`, `#soloPanel`,
  `#soloListe`) ist entfernt. Es gibt genau **eine** Spieleliste rechts.
- Oben stehen nur noch drei Hauptauswahlen: **Alle Spiele** (alle), **Einzelspieler**
  (solo), **Mehrspieler** (multi = online oder zu zweit).
- Bei „Mehrspieler“ erscheinen darunter die zwei Unterfilter **Online-Mehrspieler** und
  **Zu zweit am Gerät** (mit „Mehrspieler:“ davor); außerhalb dieses Zweigs sind sie weg.
- `ueberPasst()` kennt den Filter `multi`; die Reiter-Reste (`setModus`, `soloAuswahl`)
  sind abgesichert, `sessionStorage` merkt sich nur noch „multi“.

**Gemessen** (Playwright, lokaler Connector, 1920 × 1080)
- Hauptleiste: `["Alle Spiele","Einzelspieler","Mehrspieler"]`, 0 Reiter-Knöpfe mehr.
- Karten: Alle **95** (95 verschiedene Namen), Einzelspieler **49**, Mehrspieler **71**,
  Online-Mehrspieler **71**, Zu zweit am Gerät **58**; jedes Spiel genau einmal je Ansicht.
  Die 49 + 71 ergeben 120, weil 25 Spiele beides können (z. B. Kniffel solo und zu zweit);
  die Sammlung bleibt bei 95 Spielen.
- Unterfilter erscheinen nur bei „Mehrspieler“ (`["Online-Mehrspieler","Zu zweit am Gerät"]`),
  bei „Alle Spiele“ und „Einzelspieler“ ist die Reihe leer; Highlights (6) nur bei
  „Alle Spiele“.
- Zweispaltige Anordnung unverändert: links 614 px, rechts 1306 px, Chat unten links;
  Spiel öffnet rechts (1306 × 1030), „‹ Spieleauswahl“ 131 × 33, Raum bleibt gleich.
- Handy 390 × 844: Hauptleiste 342 px breit, Unterfilter passen, `scrollWidth` 390
  (kein Querscrollen). 0 Skriptfehler auf allen Seiten.

**Offen:** Render und Pages haben den Stand noch nicht (nicht veröffentlicht), nichts
committet. Server- und Spielmechanik unangetastet.

---

## 24.11.2026 · Connector: echte zwei Spalten

**Warum:** Der Chat lief als breite Leiste über den unteren Rand der ganzen Seite, die
Spieleauswahl saß in der oberen rechten Ecke, und die linke Spalte war voller Einstellungen
und Formulare. Die Aufteilung war nicht die, die sie sein sollte.

**Was gemacht wurde**
- Der Chat liegt jetzt **in** der linken Spalte (`#chatbox` im `<main>`), nicht mehr als
  Geschwister unter beiden Spalten. Die alten Regeln (`grid-column:1 / -1`, Chatzeile im
  Raster, `58dvh`-Sondermaß) sind ersetzt, nicht überdeckt.
- Links (30–35 %, Standard 32 %): Kopfzeile mit TRASHYGECKO und Verbindungsstatus, kompakte
  Spielerliste, Namenszeile, zugeklappte Einstellungen („Grafikstil, Einladung & Vollbild“),
  darunter der Chat: Verlauf scrollt, Eingabe sitzt unten links.
- Rechts (65–70 %): Übersicht mit Filtern, Suche, sechs flachen Highlights und allen
  Kategorien; am PC sind die Kategorien offen (95 Karten, sieben je Zeile), am Handy
  zugeklappt. Beim Start einer Partie ersetzt das Spielfeld die Übersicht; „‹ Spieleauswahl“
  führt zurück, ohne den Raum zu verlassen.
- Handy: Chat direkt unter dem Kopf (44 dvh), Spieler/Name/Einstellungen darunter, Spiele
  folgen; kein horizontales Scrollen.

**Geprüft** (Playwright gegen den lokalen Connector)
- **1920 × 1080:** links 614 px (32 %), rechts 1306 px (68 %), beide 1080 px hoch; Chat 816 px
  hoch, Eingabe bei y 1015–1061 unten links; Seite scrollt nicht (Höhe 1080), kein
  Querüberlauf. Übersicht 1306 × 1032 scrollt für sich. Spiel öffnen: Feld 1306 × 1030 rechts,
  Übersicht weg, „‹ Spieleauswahl“ 131 × 33, Raumcode unverändert, Chat bleibt sichtbar.
  Zurück: Übersicht wieder da, „Verbunden“. 0 Skriptfehler.
- **1366 × 768 / 1024 × 768** (Einstellungen aufgeklappt): links 437 / 328 px, Chat 375 / 391 px,
  Eingabe vollständig sichtbar, kein Überlauf.
- **390 × 844:** Breite 390 px ohne Querscrollen; Chat bei y 46–417 mit Eingabe bei 352–398;
  Spiele darunter; Spiel öffnet 801 px hoch, Zurück-Knopf sticky, Raum unverändert,
  0 Skriptfehler.
- Chat senden, Einzelspieler-Reiter (49 Karten), Einstellungen mit „Einladung teilen“
  (115 × 28 px) funktionieren weiter.

**Offen:** Render und Pages haben diesen Stand noch nicht (nicht veröffentlicht); nichts
committet. Spielfenster und Serverlogik (Go) sind unverändert.

---

## 24.11.2026 · Connector: „Einladung teilen“ neben dem QR-Code

**Warum:** Eine Runde per WhatsApp einzuladen hieß bisher: Raumcode abtippen. Der Link lag
nirgends griffbereit.

**Was gemacht wurde**
- `connector.html`: unter dem QR-Bild stehen zwei Knöpfe nebeneinander („Adresse kopieren“,
  „Einladung teilen“); das Aufklappfeld heißt jetzt „QR-Code & Einladung“.
- `einladungTeilen()` nutzt `navigator.share` (Handy: Systemblatt mit WhatsApp, Signal,
  Mail …) und fällt sonst auf die Zwischenablage zurück („Link kopiert – jetzt in WhatsApp
  einfügen.“), zuletzt auf ein Eingabefeld. Text: „TrashyGecko – spiel mit mir! Raum CODE“.
- Der Link nimmt die erreichbare Adresse aus `/api/status` (WLAN-IP bzw. öffentliche
  Adresse) und hängt `?room=CODE` an; `localhost` wird bewusst nicht geteilt.
- Vier Kopien gleich (md5 `629f3603…`), keine Änderung an Räumen, Chat oder Spielen.

**Geprüft** (Playwright, lokaler Connector Port 8769)
- Desktop 1440 × 900 ohne `navigator.share`: Zwischenablage
  „…Raum TG-F5E5 / http://192.168.2.106:8769/connector?room=TG-F5E5“, Hinweis sichtbar,
  0 Skriptfehler, kein Querscrollen.
- Handy 390 × 844: `navigator.share` mit `{title, text, url}` samt Raumcode aufgerufen;
  beide Knöpfe in einer Reihe (65–343 px von 390), kein Querscrollen, 0 Skriptfehler.
- Einladungslink in einem zweiten Kontext geöffnet: **gleicher Raum** (TG-F5E5),
  0 Skriptfehler.
- **Offen:** kein echter WhatsApp-Aufruf auf einem Gerät (nur die System-Schnittstelle
  geprüft); Render und Pages haben den Stand noch nicht (nicht veröffentlicht).

---

## 24.11.2026 · App-Halle: eigene Spielzeichen, Kategorien-Wechsel, Spielstatus repariert

**Warum:** Die frische APK auf dem Emulator zeigte die neue Halle nur halb: eigene
Spielzeichen fehlten, unter einer Kategorie führte kein Weg zu den anderen, und im
Connector (Online-Räume) verschwanden die Reiter unter dem klebenden Kopf, während der
Spielstatus als leerer Kasten stehen blieb.

**Was gemacht wurde**
- `Spielsymbol.java` (neu): zeichnet die Zeichen der Web-Halle im 24er-Raster mit der
  Kategorie-Farbe; liest `<path>`, `<circle>` und `<rect>` über
  `androidx.core.graphics.PathParser`, also dieselben Linien wie `spiele.html`.
- `katalogBauen` liest `SPIELSYMBOLE`, `SYMBOLREGELN` und `IKONEN` aus `spiele.html` und
  schreibt je Spiel `symbol`, je Kategorie `ikon` und alle Zeichen als `symbole` in
  `katalog.json` – eine Quelle, kein zweiter Katalog.
- `SpieleActivity`: Spielzeile mit Zeichen links, Kategorie-Kacheln mit dem echten Symbol
  statt Textzeichen (⚄ ♠ ♞ …), „Andere Kategorien“ unter einer geöffneten Kategorie wie in
  der Web-Halle. Die Suche findet wieder alle Spiele; vorher blieb sie in der offenen
  Kategorie („schach“ → 0 Treffer).
- `connector.html`: `scrollZuDenSpielen()` setzt `scroll-margin-top` auf die Höhe des
  klebenden Kopfs (die Reiter lagen 29 px darunter); `statusAnzeigen()` schreibt bei leerem
  Stand wieder „Noch keine Runde offen.“ statt eines 140 px hohen leeren Kastens.
- Beide Ablagen gleich: vier `connector.html` md5 `b4b3e379…`, `Spielsymbol.java`,
  `SpieleActivity.java` und `build.gradle` je md5-gleich. Version bleibt **1.16 / 17**.

**Geprüft**
- `:app:assembleDebug` grün in beiden Ablagen; `katalogBauen`: 94 Spiele, 6 Kategorien,
  45 Spielzeichen.
- Emulator 1080 × 2340 (Debug-APK installiert): Start mit 94 Spielen und 6 Highlights;
  Würfelspiele 25 Karten mit Zeichen; „Andere Kategorien“ zeigt 5 Kacheln und wechselt
  (Würfelspiele → Kartenspiele); Suche „schach“ → 1 Treffer mit Kronen-Symbol; 0 `FATAL`,
  0 Skriptfehler.
- Connector lokal (Port 8769, Playwright): 390 × 844 Reiter-Oberkante **59 px** bei
  Kopf-Unterkante **43 px** (vorher 11 px bei 40 px, also 29 px verdeckt); Status
  „Noch keine Runde offen.“ bei 390 und 1440 (vorher leer).
- **Offen:** Render und Pages haben den Stand noch nicht (nicht veröffentlicht); die
  „eigene Runde“ im Connector nur bis zur Statuszeile geprüft, keine Partie eröffnet.

---

## 24.11.2026 · Connector: Spielebereiche und Kacheln einzeln klappbar, Chat ganz unten

**Warum:** Die Spieleübersicht im Connector war eine lange, immer offene Liste; der Chat
stand in der linken Spalte zwischen den Elementen und wurde am Handy weggeschoben.

**Was gemacht wurde**
- **Jeder Spielebereich klappt für sich** (native `<details>`, Zustand je Liste in
  `offeneBereiche`): Kopfzeile mit Pfeil, Name, Anzahl und den ersten Spielnamen
  („Würfel · Karten · …“). Standard zu; ein Bereich öffnet keinen anderen.
- **Jede Kachel klappt für sich** (`geschlosseneKarten`): zugeklappt bleiben Name und
  ein ▸-Knopf, Text/Kennzeichen/Knöpfe verschwinden. Standard offen, damit nichts
  versteckt ist; gespielte Ansicht (iframe) wird dabei nicht neu geladen.
- **Chat als eigene Vollbreiten-Zeile** unten im `.zentrale-geruest` (Desktop Rasterzeile 2,
  `--chat-hoehe:clamp(220px,30vh,300px)`; am Handy nach den Spielen). `#spielGriff`
  endet über dem Chat.
- **Kein Seitensprung mehr** bei neuen Nachrichten: nur noch der Chatverlauf rückt nach
  (`if(atBottom || !chatSprungBereit)…`); `chatbox.scrollIntoView` entfällt. Beim ersten
  Aufbau nach dem Beitritt steht der Verlauf immer am Ende.

**Geprüft** (Chromium, Testinstanz 8767, alle vier Kopien byte-gleich,
md5 `7dea5c56664c7a76f955b678ee84abf9`)
- Alle 8 Inline-Skripte `node --check` ohne Fehler.
- Bereiche unabhängig: Würfel und Karten gleichzeitig offen, Schließen des einen lässt den
  anderen offen; Suche „risiko“ öffnet nur Brettspiele (1 Karte), Leeren stellt die vorher
  offenen Bereiche wieder her; Zustand über Größenänderung erhalten.
- Kachel 174 px → 40 px beim Zuklappen, Nachbarkachel bleibt 174 px.
- Desktop 1280 × 800: links 410 × 560, rechts 870 × 560, Chat 1280 × 240 bei y 560–800,
  Seite ohne Scrollen; mit offenem Spiel Fenster 870 × 508, Chat darunter.
- Handy 390 × 844: Chat beginnt bei y 1286 unter der Spielauswahl (372 × 287),
  `scrollWidth` 374 → kein Querscrollen, Kategorie-Kopf 51 px.
- Chat: initial `scrollTop` 48,8 bei max 49; Testkasten 60 px: nach neuer Nachricht
  `top` 234 = `max` 234, `scrollY` bleibt 0.

**Offen:** keine zweite echte Spielersitzung; eine eingehende Fremdnachricht wurde nicht
live erzeugt (nur über den bestehenden Verlauf geprüft). Sicherung vor der Änderung:
`sicherung/2026-11-24-connector-klapp/`.

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
