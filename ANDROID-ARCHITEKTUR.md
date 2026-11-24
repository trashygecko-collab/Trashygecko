# TrashyGecko Android – Architektur

Stand: 23.11.2026. Dieses Dokument beschreibt **zuerst den tatsächlichen Ist-Zustand**
(im Code gelesen, nicht vermutet) und danach das **Zielbild**. Was getestet ist und was
nicht, steht ausschließlich in [`PRUEFBERICHT-ANDROID.md`](PRUEFBERICHT-ANDROID.md).

---

## 1. Es gibt zwei Android-Projekte

| | `spiele/android` | `apps/t-spiele` |
|---|---|---|
| Rolle | **die App** (TrashyGecko) | Altlast: Startbildschirm mit zwei Knöpfen |
| Paket | `de.trashygecko.tmahjong` | `de.trashygecko.tspiele` |
| Fassung | 1.15 / versionCode 16 | 1.4 / versionCode 5 |
| Bau | Gradle 8.7 + AGP 8.5.2, `app/build.gradle` | `bauen.sh` mit aapt2/d8/apksigner, **kein Gradle** |
| minSdk | 21 | 14 |
| Inhalt | Spiele-Baum als Assets, lokale Räume, Online-Connector, Fernbedienung | zwei Links in den Browser |
| Größe | ~10,5 MB APK | 144.671 B |

Der Paketname `tmahjong` ist historisch (die App war zuerst nur das Mahjong-Spiel) und
bleibt erhalten, weil Paketidentität und Signatur über Updates hinweg gleich bleiben
müssen. Der Anzeigename ist **TrashyGecko**.

`apps/t-spiele` ist die „Mini-App, die nur eine Webseite öffnet“ – sie baut **keine**
Spiele ein. Sie wird nicht weiterentwickelt, bleibt aber als Verteilkopie bestehen
(siehe Entscheidungen D-08).

---

## 2. Ist-Zustand: Aufbau der App

```
                    ┌───────────────────────────────────────────┐
   Launcher  ──────▶│ MainActivity  (193 Zeilen, nativ)          │
                    │  vier Kacheln: Spiele, Lokales Netz,       │
                    │  Online-Räume, Extras · Netzstatus         │
                    └───────┬───────────────────────────────────┘
                            │ Kachel öffnet Activity
          Spielseiten rufen WebNavigation.pruefe() (eine Regel für alle Activities)
          ┌─────────────────┼──────────────────┬──────────────────┐
          ▼                 ▼                  ▼                  ▼
 ┌────────────────┐ ┌───────────────┐ ┌──────────────┐ ┌──────────────────┐
 │ SpieleActivity │ │ GameActivity  │ │ Connector    │ │ Browser          │
 │ (nativ, 363)   │ │ (WebView, 219)│ │ Activity     │ │ (fremde Hosts)   │
 │ Liste aus      │ │ eine          │ │ (443, nativ) │ │                  │
 │ katalog.json   │ │ Spielseite    │ │              │ │                  │
 └───────┬────────┘ └───────┬───────┘ └───┬──────┬───┘ └──────────────────┘
         │ Tipp auf Spiel   │             │      │
         └─────────────────▶│             │      │
                            │             │      │
              ┌─────────────▼──┐  ┌───────▼───┐  │
              │ spiele/…/*.html│  │LocalNetwork│  │
              │ aus den Assets │  │Coordinator │  │
              │ oder vom Server│  │ (NSD+TCP)  │  │
              └────────────────┘  └────────────┘  │
                                        ┌─────────▼──────────┐
                                        │ RemoteControl      │
                                        │ Activity (+ remote)│
                                        │ Android TV / Roku  │
                                        └────────────────────┘
```

### 2.1 Native Teile (Java)

| Datei | Zeilen | Aufgabe |
|---|---|---|
| `MainActivity.java` | 173 | **native Startansicht** (D-12): Marke, Titel, vier Kacheln (Spiele, Lokales Netz, Online-Räume, Extras), Netzstatus des Geräts in der Kopfzeile, Vollbild – kein WebView; Farben und Bausteine aus `Ansicht` |
| `HomepageActivity.java` | 142 | die **echte Startseite der Website** (`assets/www/homepage.html`) im WebView, nur über „Extras → Homepage“ erreichbar, Zurück führt in die App (D-12) |
| `SpieleActivity.java` | 614 | **native Spieleliste** aus `assets/katalog.json`: Textsuche, Merkmalsfilter (UND), Kategorien mit Anzahl, Dialog „Lokal / Online“, startet `GameActivity`; `onConfigurationChanged` bricht die Chips für die neue Breite um (D-21), `onSaveInstanceState`/`stelleWieder` sichern Suche, Merkmale und Scrollposition über den Prozess-Tod (D-22, erst wiederherstellen, dann einmal zeichnen) |
| `GameActivity.java` | 219 | eine einzelne Spielseite (lokal aus den Assets oder vom Render-Server), Mahjong-Brücke, Zurück = Spiel verlassen |
| `WebNavigation.java` | 142 | **die eine Link-Regel**: was im WebView bleibt, was `GameActivity`, `SpieleActivity` oder `ConnectorActivity` öffnet, was an den Browser geht |
| `WebBasis.java` | 85 | gemeinsame WebView-Einstellungen, Vollbild, Konsolenprotokoll, Online-Host und Pfadprüfung |
| `Ansicht.java` | 53 | gemeinsames Farbbild (`BG`, `KARTE`, `RAND`, `GOLD`, `TEXT`, `MUTED`) und die Bausteine `flaeche`, `text`, `dp` für Startansicht und Spieleliste |
| `ConnectorActivity.java` | 494 | nativer Bereich „Connector“: Karten für Lokales Netz und Online, Extras; lokaler Chat; Galgen-WebView; Online-Kachel lädt ohne Netz gar nicht erst und zeigt statt Chromiums englischer Fehlerseite eine deutsche Meldung mit „Erneut versuchen“ (D-23); Zurück schließt direkt, wenn der Bereich über eine Kachel der Startansicht geöffnet wurde (`vonAussen`) – **kein** Einstieg zur Spieleliste (D-20) |
| `LocalNetworkCoordinator.java` | 455 | NSD-Dienst `_trashygecko._tcp`, TCP-Relay im Gastgebergerät, Raumcode, Ereignis-Nummern, Deduplizierung, Wiederverbindung (12 Versuche, 120 s Kulanz) |
| `GalgenLokal.java` | 355 | Regeln des Galgens für den lokalen Relay (Port von `spiele/connector/galgen.go`) |
| `RemoteControlActivity.java` | 173 | Fernbedienung (WebView + native Brücke) |
| `remote/*.java` | 572 | Android-TV-Remote-v2 (TLS/PIN), Roku-ECP-Erkennung, SSDP, Protobuf-Framing |
| `HomepageAssets.java` | 92 | liefert `assets/www` unter einer echten `https`-Herkunft aus (nötig für `localStorage`) |

### 2.2 Native Ansichten und eingebettete Webteile

Start und Spieleliste sind nativ (seit 23.11.2026); HTML aus `assets/www` sind die Spiele,
die Fernbedienung und die alte Startseite als Rückweg:

- `MainActivity` – **native Startansicht** (D-12): Marke, Titel, vier Kacheln (Spiele,
  Lokales Netz, Online-Räume, Extras) und in der Kopfzeile der Netzstatus des Geräts
  („offline spielbar“ / „WLAN · online bereit“, in `onResume` neu gelesen).
- `app.html` – die alte HTML-Startseite (Quelle `spiele/android/start/app.html`) liegt nur
  noch als `index.html` im Paket, damit „zurück zum Start“ aus einer Spielseite landet.
  Die App selbst öffnet sie nicht mehr.
- `homepage.html` – die **echte Startseite der Website** (`index.html`), beim Bauen als
  `homepage.html` mitkopiert (neben `main.css`); nur über „Extras → Homepage“
  (`HomepageActivity`) erreichbar, Zurück führt in die App. Weitere Website-Seiten
  (Tagebuch, Bücher, Rezepte …) liegen nicht im Paket.
- `spiele/spiele.html` – **die Quelle des Spielekatalogs**: eine Desktop-Seite mit Suche,
  Filtern und der JS-Liste `SPIELE`; daraus entsteht beim Bauen `assets/katalog.json`.
  In der App zeigt `SpieleActivity` den Katalog nativ; die HTML-Seite bleibt im Paket
  (Sicherheitsnetz), wird aber von der App nicht mehr verlinkt. Beim Bauen werden darin
  zwei Zeilen (APK-Installer, Fernbedienung) entfernt.
- alle Spielseiten unter `spiele/**` (343 Dateien, 105 HTML-Seiten, 91 aktive Katalogeinträge).
- `fernbedienung/**` – die Web-Oberfläche der Fernbedienung.

Der Übergang von Katalog zu Spiel ist ein Tipp auf eine Zeile der nativen Liste. Hat das
Spiel das Merkmal `online`, fragt `SpieleActivity` zuerst „Lokal / offline“ oder
„Online-Spiel“ (wie früher der Dialog aus `app.html`); sonst startet `GameActivity`
sofort lokal. Eigener Verlauf, eigener Zurück-Weg – die Liste bleibt samt Suchtext und
Scrollposition im Hintergrund stehen.

Über der Liste steht eine Filterleiste mit den Merkmalen, die im Katalog wirklich
vorkommen (Chips, umbrechend wie im HTML-Katalog). Gewählte Chips sind UND-verknüpft und
wirken zusammen mit der Textsuche; Kategorieköpfe zeigen dann „x / gesamt“, der Zähler
„x von 94“, und „Zurücksetzen“ leert beides. Die Zeilenansichten entstehen einmal und
werden beim Filtern nur umgehängt – deshalb kostet ein Filterwechsel wenige Millisekunden
und die Liste springt wieder nach oben.

### 2.3 Bauweg (Gradle)

```
:app:homepageKopieren   (Sync, hängt an preBuild)
   kopiert ../../../spiele/**  → app/src/main/assets/www/spiele/
   kopiert außerdem die 19 Ordner aus `homepageSeiten` (anbau, batch, bucher, cheats, css,
      filme, gedichte, html, musik, rezepte, russisch, script, seiten, sprueche, tagebuch,
      tipps, weisheiten, witze, youtube) samt main.css → www/<ordner>/ (D-24);
      ohne den alten css-Spiegel, musik/.backup-* und tagebuch/Textdokument (neu).txt
   kopiert ../../../WLAN-Fernbedienung/{index.html,styles.css,README.md,js} → www/fernbedienung/
   kopiert start/app.html → www/app.html und www/index.html
   entfernt danach zwei Katalogzeilen (APK-Installer, Fernbedienung) – bricht ab, wenn das Muster fehlt
:app:katalogBauen   (hängt an preBuild)
   liest den fertigen Katalog www/spiele/spiele.html und schreibt src/main/assets/katalog.json
   (94 Spiele, 6 Kategorien, Merkmale; bricht bei 0 Spielen oder fehlender Datei ab)
:app:verifyHomepageGames
   prüft jeden Katalogeintrag gegen Quelle und Asset, schreibt build/reports/homepage-integration/spiele-manifest.md
:app:packageTrashyGeckoApk
   assembleRelease (mit keystore.properties) oder assembleDebug, kopiert nach
   app/build/outputs/apk/universal/TrashyGecko.apk
:app:remoteProtocolTest
   47 Protokoll-Assertions (SSDP/DIAL-Parser, Keymapping, Framing)
```

Ausschlüsse beim Kopieren: `spiele/android/**`, `spiele/connector*/**`, Tests, `node_modules`,
`.go`, `.exe`, `*.apk`, `*.jks`, `ANLEITUNG.txt` (enthält Passwörter im Klartext).

### 2.4 Netzwerk

| Weg | Mechanik | Zustand |
|---|---|---|
| Gleiches WLAN | NSD-Suche `_trashygecko._tcp`, TCP-Relay im Gastgeber, 8-stelliger Code | implementiert, nur Chat + Galgen |
| Hotspot | derselbe NSD-Weg; hängt davon ab, ob der Hotspot Multicast weiterleitet | implementiert, **nicht getestet** |
| Unterschiedliche Netze | WebView auf `https://trashygecko.onrender.com` (Online-Connector) | implementiert, hängt am Render-Dienst |
| Spiel-APIs (`/api/…`) | nur gegen den Online-Server; es gibt **keinen** lokalen Go-Server in der App | Grenze |

Der lokale Relay ist **nicht** Ende-zu-Ende-verschlüsselt und für vertrauenswürdige
Heimnetze gedacht. Ports werden dynamisch gewählt und über NSD bekanntgegeben.

### 2.5 Lokale Speicherung

- **Keine** `SharedPreferences`, keine native Datenhaltung. Alle Spielstände, Namen und
  Stile liegen im `localStorage` der WebView unter der Herkunft
  `https://appassets.androidplatform.net/` (bei Online-Spielen unter der Render-Herkunft).
- Deshalb ist „Daten erhalten“ an die Paket-ID (nicht an die Version) gebunden:
  `adb install -r` behält sie, Deinstallieren löscht sie.

### 2.6 Rechte

`INTERNET`, `ACCESS_NETWORK_STATE`, `ACCESS_WIFI_STATE`, `CHANGE_WIFI_MULTICAST_STATE`,
`VIBRATE`. Keine Laufzeitabfragen, kein Speicher-, Kamera- oder Standortrecht.

---

## 3. Die größten Architekturprobleme (Ist)

1. **Drei Bildschirme, drei Handschriften.** Start und Spieleliste sind nativ, der Connector
   hat eigene Kartenknöpfe und eigene Zurück-Logik; gemeinsame Bausteine für Farben und
   Kopfzeile gibt es noch nicht – jede Datei bringt ihre eigenen Farbkonstanten mit.
2. **Der Start zeigt nur das Gerätenetz.** Ob der Online-Server antwortet, sieht man erst,
   wenn die Connector-Seite lädt. Zustände (prüfe / verbunden / nicht erreichbar),
   Wiederverbindung und „erneut versuchen“ fehlen.
3. **LAN-Multiplayer deckt ein Spiel ab.** Nur Galgen läuft über den lokalen Relay; alle
   anderen Spiele brauchen `/api` und damit den Online-Server. Auf einem Handy ohne
   Internet ist Mehrspieler damit fast nicht vorhanden.
4. **Zwei Apps mit ähnlichem Namen.** `T.Spiele` (nur zwei Links) und `TrashyGecko`
   (die eigentliche App) sorgen für Verwirrung – die kleine wird für „die App“ gehalten.
5. **Release-Signatur ist derzeit nicht baubar.** `keystore.properties` fehlt im Projekt
   (per `.gitignore` ausgenommen); der Keystore selbst liegt außerhalb beider Repos unter
   `Documents\arbeit\werkzeuge\trashygecko-release\`. Ohne diese Datei baut
   `packageTrashyGeckoApk` nur eine Debug-APK, die **nicht** update-fähig über 1.15 ist.

---

## 4. Zielbild

Grundsatz: **Die Spiele bleiben Web-Technik** (sie sind es und sie funktionieren), **die
App drumherum wird nativ.** Kein Nachbau von 94 Spielen, keine zweite Spiele-Engine.

```
┌──────────────────────────────────────────────────────────────────┐
│ MainActivity  (nativ, eine Activity, eigene Ansichtswechsel)      │
│                                                                  │
│  ┌──────────┬──────────┬──────────┬──────────────┐               │
│  │ Start    │ Spiele   │ Netzwerk │ Extras       │  (Tab-Leiste) │
│  └──────────┴──────────┴──────────┴──────────────┘               │
│   Start:    Kacheln mit echtem Status (offline/online, Raum)      │
│   Spiele:   natives Raster/Liste aus katalog.json, Suche/Filter   │
│   Netzwerk: lokaler Raum (NSD) + Online-Verbindung mit Status     │
│   Extras:   Fernbedienung, Einstellungen, Über, Homepage (HTML)   │
└───────────────┬──────────────────────────────────────────────────┘
                │ Spiel öffnen
                ▼
        ┌───────────────────────────────────────┐
        │ GameActivity (nativ + WebView)        │
        │  Titelleiste: Name, Zurück, Neuladen  │
        │  WebView: assets/www/spiele/…         │
        │  oder https://trashygecko.onrender…   │
        │  Brücken: TGApp, TGLokal, TMahjong    │
        └───────────────────────────────────────┘
```

### 4.1 Was nativ wird

| Bereich | heute | Ziel |
|---|---|---|
| Startbildschirm | **nativ**, vier Kacheln + Netzstatus (23.11.2026) | Kacheln mit Raum-/Serverstatus |
| Spieleübersicht | **nativ**, Liste mit Suche und Merkmalsfiltern (23.11.2026) | Raster/Liste aus `assets/katalog.json`, Kategorien, Favoriten |
| Navigation | Activity je Bereich, `WebNavigation` für Links aus Spielen | Tab-Leiste + eigener Rücksprung-Stapel |
| Verbindungsstatus | Netzstatus auf dem Start; Server erst beim Laden der Connector-Seite | Zustandsanzeige mit Prüfung, Fehlertext, „erneut versuchen“ |
| Einstellungen | `localStorage` der Spiele | native Einstellungen (Name, Grafikstil, Cache leeren), weitergereicht an die Spiele |
| Spielmodus-Wahl | nativer Dialog aus injiziertem JS | native Wahl beim Antippen im nativen Raster |

### 4.2 Was eingebettet bleibt

- alle Spielseiten unter `spiele/**` samt `main.css`, Schriften, Bildern, Sounds;
- die Fernbedienungs-Oberfläche (ausschließlich über **Extras**);
- die echte Homepage (`index.html`) als eigener Eintrag unter **Extras → Homepage**
  (seit 23.11.2026 umgesetzt; im Paket liegt die Startseite, nicht die ganze Website),
  damit „die richtige TrashyGecko-Startseite“ lokal da ist, ohne die App zu ersetzen;
- die bestehenden nativen Bausteine unverändert: `LocalNetworkCoordinator`, `GalgenLokal`,
  `RemoteControlActivity` + `remote/*`, `HomepageAssets`.

### 4.3 Datenfluss Spieleliste

```
spiele/spiele.html  (eine Quelle, JS-Liste SPIELE, 106 Zeilen)
        │  Gradle-Task :app:katalogBauen (neu)
        ▼
assets/katalog.json  {name, pfad, kategorie, stichworte, online, groesse}
        │
        ▼
native Spieleansicht  ──antippen──▶ GameActivity(WebView)
```

Vorteil: Die Website bleibt unverändert die Wahrheit; die App liest dieselbe Liste, ohne
HTML zu parsen oder JavaScript zu injizieren. Der Build bricht ab, wenn ein Eintrag auf
eine fehlende Datei zeigt (das macht `verifyHomepageGames` schon heute).

### 4.4 Connector-Anbindung (verbindlich)

- **Online**: `https://trashygecko.onrender.com`. Die App prüft die Erreichbarkeit
  (`/api/…` mit kurzem Timeout) und zeigt das Ergebnis als Status. Spiele mit
  Online-Teil öffnen gegen diese Herkunft; Einzelspieler bleibt lokal.
- **Lokal**: NSD + TCP-Relay wie vorhanden, unverändert. Weitere Spiele werden nur
  angebunden, wenn ihre Regeln serverseitig klein genug sind (Vorbild `GalgenLokal`).
- **Kein** Zusammenlegen der beiden Go-Connector-Bäume, kein Go-Server im APK.
- **Keine** neue Parallelimplementierung: die App nutzt die vorhandenen Endpunkte.

---

## 5. Wanderweg (Etappen)

1. **Bestandsaufnahme** – erledigt, siehe oben und `PRUEFBERICHT-ANDROID.md`.
2. **Dokumente** – diese Datei, `ENTSCHEIDUNGEN.md`, `ENTWICKLUNGSSTAND.md`,
   `PRUEFBERICHT-ANDROID.md`, `AGENTS.md`.
3. **Fundament** – `GameActivity` aus `MainActivity` herauslösen (WebView + Leiste),
   `katalog.json` im Build erzeugen. Danach ist die App funktional unverändert, aber
   technisch getrennt.
4. **Native Oberfläche** – native Startansicht und native Spieleliste sind umgesetzt;
   offen sind Tab-Leiste, Einstellungen und die Statusanzeige mit Serverprüfung.
   `app.html` bleibt als `index.html` im Paket (Rückweg aus Spielseiten).
5. **Netzwerk** – Statusprüfung, Wiederverbindung, verständliche Fehler; LAN-Räume aus
   der nativen Ansicht erreichbar.
6. **Qualität** – Gradle-Build, Installation auf dem vorhandenen Android-9-Gerät
   (Huawei FIG-LX1) und im Emulator, Prüfungen siehe `PRUEFBERICHT-ANDROID.md`.
