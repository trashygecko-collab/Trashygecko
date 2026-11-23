# Entwicklungsstand – TrashyGecko Android

Lebendes Dokument. Nach **jeder** abgeschlossenen Arbeit fortschreiben (oben das Neueste).
Was getestet wurde, gehört zusätzlich in [`PRUEFBERICHT-ANDROID.md`](PRUEFBERICHT-ANDROID.md).

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
