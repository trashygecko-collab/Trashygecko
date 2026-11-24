# Nachtarbeit – Spielhalle, 23./24.11.2026

Arbeitsstand: sicher, nichts committet, nichts veröffentlicht. Beide Ablagen
(`Documents\Homepage\…` und `Documents\GitHub\Homepage\…`) sind nach dem Lauf
byte-gleich; Belege unten.

---

## 1. Was tatsächlich verändert wurde

### 1.1 Spielhalle: eigene Symbole statt Buchstabenkacheln
Jede der **94 Spielkarten** hat jetzt ein gezeichnetes Symbol (SVG, 24er-Raster,
Kategorie-Farbe) statt eines Monogramms. 45 Zeichen stehen zur Verfügung
(Würfel, Karten, Brett, Krone, Mühle, Dame, Domino, Ziel, Leiter, Globus,
Schiff, Bombe, Mine, Kiste, Puzzle, Juwel, Spross, Flipper, Alien, Schlange,
Pac, Auto, Vogel, Flasche, Galgen, Buchstabe, Rechnen, Kartenpaar, Block,
Turm, Sprung, Rakete, Irrgarten, Gebiet, Schale, Punkte, Linie, Streichholz,
Säule, Tic-Tac, Steine, Schläger, Ball). Zuordnung über Pfadregeln, Rückfall
ist das Kategorie-Symbol. Damit ist keine Karte mehr nur ein Kürzel.

### 1.2 Kategorien-Navigation
Unter den Spielen einer Kategorie stehen die fünf anderen Kategorien als
beschriftete Kacheln mit Symbol, Name und Spielzahl („Würfelspiele · 25 Spiele“).
Kein Umweg mehr über den Startbildschirm.

### 1.3 Online: eigene Runde bleibt sichtbar
- Nach dem Schließen des Tisches bleibt die eröffnete Runde in der Online-Leiste
  als Knopf **„Runde: 10.000“**; ein Klick öffnet denselben Tisch.
- Im Online-Panel gibt es dafür den Block **„Offene Runde“** mit
  **Tisch öffnen** und **Zurückziehen** (`/api/angebot`, Aktion `zurueck`);
  bei fremden Runden **Mitspielen** und **Ablehnen**.
- **„Partie fortsetzen“** bleibt nach dem Schließen erhalten (vorher ging der
  Verweis mit dem geschlossenen Overlay verloren) und wird nach einem Neuladen
  aus `tg-halle-tisch` wiederhergestellt.

### 1.4 Kleinere Fehler und Politur
- **Stern, Pfeil und Beschriftungen fehlten:** Der Seitenhelfer `el(tag, klasse)`
  nimmt keinen dritten Textparameter; mehrere Aufrufe übergaben trotzdem einen.
  Dadurch waren der ★ im Highlight-Abzeichen, der ›-Pfeil im Startknopf, der
  Text „Raum + Chat“ und die neuen Kategorien-Beschriftungen leer. Alle vier
  Stellen setzen den Text jetzt ausdrücklich (`textContent`).
- Zurück-Pfeil in der Suchleiste ist ein echtes `←` statt eines dünnen `‹`.
- Symbolgröße 26 px im 44-px-Feld, leichte Vergrößerung beim Überfahren.

### 1.5 Schon in dieser Nacht davor (gleicher Stand, hier zur Vollständigkeit)
- Neue Optik `spiele/halle.css` (dunkler Grund mit Tiefe, kräftige
  Kategorie-Akzente, große Bildkarten, 6 Kacheln, 1–4 Spalten, reduzierte
  Bewegung, Kontrastmodus).
- Highlights: Flipper und Quizshow raus, **Mahjong** und **10.000** rein, neue
  Bilder (1392 × 752) in beiden Ablagen.
- `spiele/halle-online.js`: Räume, Einladungslink `?raum=`, QR, Anwesenheit,
  Angebote, Chat (Dock/Blatt), Reconnect, Tisch-Overlay mit der vorhandenen
  Spielseite; keine zweite Kennung, keine Go-Änderung.

### 1.6 Bereitschaft und Start direkt in der Halle (Fortsetzung der Nacht)
Die Tisch-Kopfzeile der Halle hat jetzt **Bereit** (mit Häkchen, wenn gesetzt) und
für den Gastgeber **Spiel starten** (erst aktiv, wenn mindestens zwei Spieler da
und alle bereit sind). Beides ruft **dieselben** Endpunkte wie die Spielseite
(`?action=ready` / `?action=start`), es gibt also keine zweite Lobby und keinen
zweiten Zustand. Damit ist der geforderte Ablauf vollständig in der Halle
bedienbar: auswählen → Raum → einladen → beitreten → **bereit** → **starten** →
spielen → Chat. „Zurück zur Halle“ lässt die Partie laufen.

Dabei drei echte Fehler gefunden und behoben:
- **Galgen meldet `status:"spiel"`**, nicht `active`. Die Halle zeigte deshalb
  „2 am Tisch“ statt „Partie läuft“ und hätte den Start nicht erkannt. Jetzt gilt:
  alles außer `invited`/`finished`/`ende`/`declined` ist eine laufende Partie.
- **Angebot blieb nach dem Start offen** (Server lässt es für weitere Spieler
  absichtlich stehen, aber `join` geht nur vor dem Start). Die Halle zieht das
  eigene Tisch-Angebot jetzt zurück, sobald die Runde läuft – geprüft: Angebot
  danach `null`.
- **Schach hat keine `players`-Liste** (nur Weiß/Schwarz). Die Halle behauptete
  „Warte auf Mitspieler“, während beide schon spielten. Jetzt „Partie läuft.“,
  und eine offene Einladung heißt „Einladung offen – im Chat oder im Spiel
  annehmen.“

---

## 2. Geänderte Dateien

| Datei | Änderung |
|---|---|
| `spiele/spiele.html` | Spielsymbole + Zuordnung, Kategorien-Navigation, Textfehler (Stern/Pfeil/Raum+Chat), Zurück-Pfeil; Online-Einstieg je Karte, Modus-Blatt, Filter „Online spielbar“ |
| `spiele/halle.css` | komplette Spielhallen-Optik; neu: Symbolgrößen, Kategorien-Wechsel, Karten-/Highlight-Regeln, Chat, Overlay, Panels, Handy-Feinheiten |
| `spiele/halle-online.js` | Online-Schicht (Räume, Presence, Angebote, Chat, Tisch, Reconnect); neu: eigene Runde in Leiste/Panel, dauerhaftes „Partie fortsetzen“, `letzterTisch` |
| `spiele/bilder/highlights/mahjong.jpg`, `…/zehntausend.jpg` | neue Highlight-Bilder |
| `spiele/android/app/src/main/assets/…` | Katalog und Web-Assets neu erzeugt (APK) |
| `spiele/android/tmahjong.apk`, `…/TrashyGecko.apk` | neu gebaut |
| `ENTWICKLUNGSSTAND.md`, `ENTSCHEIDUNGEN.md` (D-27), `AGENTS.md` | Dokumentation |
| `NACHTARBEIT-BERICHT.md` | dieser Bericht |

Nicht angefasst: Go-Quellen, Spielmechaniken, Spielpfade, Katalogzeilen,
Kartenspiel-Optik, `design.css`, `multiplayer-register.js`.

---

## 3. Was jetzt wirklich funktioniert

**Spielhalle**
- 94 Spiele in 6 Kategorien, Suche, Kategorien, Highlights, Extras.
- 94/94 Karten mit eigenem Symbol; 5 Wechsel-Ziele je Kategorie.
- Kein Querscrollen bei 390 px auf Start, Kategorie und Highlights (gemessen
  390/390).
- Jede Karte mit Online-Einstieg fragt den Modus ab (Einzelspieler/lokal/Online);
  Spiele ohne Online-Einstieg starten sofort. Der Filter „Online spielbar“
  zeigt genau die 70 Spiele aus dem Register.

**Multiplayer (über den vorhandenen Connector)**
- Identität `tg-player-id` / `tg-player-name`, Token über `token.js`; ohne Namen
  „Gast XYZ“, im Panel änderbar.
- Eigener Raum beim Verbinden, Beitritt per Link (`?raum=CODE`) oder Raumcode,
  QR über `/api/qr`.
- Anwesenheit (20 s), Spielerliste, Chat (`/api/messages`) mit Ungelesen-Zähler,
  Einladungszeilen mit Mitspielen/Ablehnen.
- Runde eröffnen (`/api/angebot` bieten), beitreten (annehmen), ablehnen,
  zurückziehen; Tisch-Overlay mit der echten Spielseite; Statuszeile mit
  Spielerzahl/Bereitschaft; Bereit und Start in der Spiel-Lobby.
- Reconnect: 409 → gleicher Platz, 403 → neue Kennung (max. 1×/30 s),
  404 → neuer Raum; gespeicherter Raum und Tisch werden wieder aufgenommen.

**Weitere Tischspiele über die Halle geprüft** (Raum `TG-96F2`, A über Loopback,
B über LAN): **Kniffel**, **Galgenmännchen** und **Schach** – jeweils Runde
eröffnen, Beitritt, Bereit/Kniffel+Galgen, Start, laufende Partie im Rahmen.
Schach läuft als Einladung (Beitritt = Annahme, kein Bereit-Knopf).

**10.000 als Referenz**
- Kategorie → 10.000 → Modi → Online-Runde → Tisch → Mitspieler → Bereit →
  Host-Start → Würfeln/Punkte auf dem Server → Chat → zurück zur Halle →
  „Runde: 10.000“ / „Partie fortsetzen“.

---

## 4. Durchgeführte Tests (alle ausgeführt, keine Annahmen)

**Bestandsaufnahme**
- Register: **71** Einträge, **0** doppelte IDs, **0** fehlende Pfade,
  Arten `tisch/einladung/gemeinsam`.
- Katalog: **94** Spiele, **6** Kategorien; **70** Katalogspiele sind im Register
  (ein Register-Eintrag, `malen`, ist Werkzeug); **24** Spiele ohne Online-Teil.
- `go test -run Register ./...` in `spiele/connector`: **PASS**
  (`TestRegisterDecktAlleSpieleAb`).

**Zwei Geräte, zwei Ursprünge** (Playwright, Chromium; Connector lokal Port 8768):
- Gerät A `http://127.0.0.1:8768`, Gerät B `http://192.168.56.1:8768` (LAN-Adresse
  des Rechners, damit der Connector B wie ein echtes zweites Gerät behandelt).
- Raum **TG-6348**, A = „Anna“, B = „Gast YU1“; beide sehen **2 Spieler**.
- A eröffnet 10.000; Tisch zeigt „1 am Tisch · 0 bereit“; B bekommt das Angebot
  „ANNA LÄDT EIN … 2–6 Spieler“ und tritt bei.
- Beide „Bereit“, A startet: „2 am Tisch · 2 bereit“, danach Partie aktiv in
  beiden Fenstern.
- Chat beide Richtungen („Anna über Loopback.“ / „Bert über LAN.“), 9 Beiträge.
- A schließt den Tisch, Leiste zeigt **„Runde: 10.000“**, Panel zeigt
  **„Deine Runde ‚10.000‘ wartet im Raum TG-6348“** mit Tisch öffnen/Zurückziehen;
  Tisch öffnet sich ohne zweite Runde.
- **Keine** JavaScript-Fehler, keine Konsolenmeldungen.

**Optik/Überlauf**
- Start 1440 × 900: 4 Highlights, 6 Kategorien; Highlights-Ansicht 4 Karten,
  Abzeichen ★ HIGHLIGHT und ONLINE 2–6 sichtbar.
- Handy 390 × 844: Start, Kategorie, Highlights je `scrollWidth` 390 =
  kein Querscrollen; 25/25 Symbole in Würfelspiele.
- Einzelspieler-Einstieg: „Jetzt spielen“ → `…/zehntausend.html` **ohne** `?room`.

**Syntax**
- `node --check` für `halle-online.js` und beide Inline-Skriptblöcke aus
  `spiele.html`: grün.

**Drei Spiele hintereinander in einem Raum** (ein Durchlauf, keine Neuladungen):
- Raum `TG-96F2`; Kniffel: „2 am Tisch · 2 bereit“, Start, „Partie läuft · 2 Spieler“,
  im Rahmen ist `#game` sichtbar und die Lobby versteckt.
- Galgen: derselbe Ablauf, im Rahmen steht „Läuft _ _ _ _ _ _ … Kategorie: Beruf“.
- Schach: Beitritt startet die Partie, beide sehen „Partie läuft.“, das Brett ist
  sichtbar („Weiß beginnt – wähle eine Figur“).
- **0 Konsolenfehler** in allen drei Partien.
- Handy 390 × 844 mit den neuen Knöpfen: Kopfzeile und Dokument **0 px**
  Querscrollen, Knöpfe „Bereit“ / „Erst alle bereit“ sichtbar.

**Android**
- `:app:katalogBauen`: 94 Spiele, 6 Kategorien, Highlights Mahjong/10.000/
  Risiko/Labyrinth.
- `:app:packageTrashyGeckoApk`: erfolgreich, Assets enthalten `halle.css`,
  `halle-online.js` und beide Bilder: **14.676.385 B**, md5
  `abee9417605e74fc78411ce6b5563e33`.
- Version bleibt **1.16 / Code 17** (Änderung nur mit Rückfrage).

---

## 5. Gefundene und behobene Fehler

1. **Leere Beschriftungen** (Stern, Pfeil, „Raum + Chat“, Kategorien-Namen):
   `el()` ignoriert den dritten Parameter. Behoben, geprüft („★ HIGHLIGHT“,
   „Jetzt spielen ›“, „RAUM + CHAT“, „Kartenspiele · 11 Spiele“).
2. **Eigene Runde unsichtbar**, sobald der Tisch geschlossen war; nur über die
   Spielkarte erreichbar. Jetzt Leisten-Knopf + Panel-Block.
3. **„Partie fortsetzen“ verschwand** mit dem Schließen (`s.tisch = null`).
   Jetzt `letzterTisch` + Wiederaufnahme aus dem Speicher.
4. **Hauptraum statt eigener Raum:** Der erste Entwurf trat dem offenen
   Hauptraum bei (fremde/alte Spieler, alte Angebote). Jetzt legt „Mit Freunden
   spielen“ einen eigenen Raum an; Links betreten gezielt einen Raum.
5. **Anonyme Namen:** Vorher „Namenloser“/„Namenloser1“; jetzt „Gast XYZ“.
6. **Doppelte Runde beim Wiederöffnen:** Ein laufendes Spiel wird über die
   Spiel-API erkannt und fortgesetzt statt neu angeboten (geprüft: 1 am Tisch).

---

## 5a. Weitere gefundene und behobene Fehler (Fortsetzung)

7. **Bereitschaft nur in der Spielseite:** Für den Weg „auswählen → … → starten“
   musste man in die Spielseite wechseln. Jetzt in der Hallen-Kopfzeile, mit
   denselben Endpunkten.
8. **Status `spiel` unbekannt:** Galgen (und andere) nennen den laufenden Zustand
   nicht `active`; die Halle zeigte weiter „am Tisch“. Behoben, geprüft.
9. **Angebot blieb offen:** Nach dem Start stand im Raum weiter eine „offene
   Runde“, obwohl `join` nur vor dem Start möglich ist. Zuerst räumte die Halle
   ihr eigenes Angebot auf – ein Start aus der Spielseite heraus ließ es aber
   stehen. Seit 24.11. entscheidet der Connector beim Abruf selbst (9.1).
10. **Falsche Aussage bei Schach:** „Warte auf Mitspieler“, während gespielt
    wurde. Behoben.

## 6. Was noch offen ist

- **Vier Spiele** sind vollständig durchgespielt: 10.000, Kniffel, Galgenmännchen,
  Schach. Die übrigen 66 Register-Spiele nutzen dieselben Endpunkte
  (`create/join/ready/start` bzw. `invite/accept`), wurden aber nicht einzeln
  geprüft. Das bleibt die größte offene Lücke.
- **Start aus der Spielseite** räumt das Hallen-Angebot nicht auf; nur der Weg
  über die Hallen-Knöpfe tut das. Ursache: Das Angebot liegt im Connector-Raum,
  die Spielseite kennt es nicht. Sauber wäre, der Connector selbst prüfte beim
  Abruf, ob die Runde noch offen ist – dafür fehlt eine gemeinsame
  Statusabfrage über alle Spiele.
- **Kein echtes zweites Gerät:** getestet mit zwei Browser-Kontexten auf einem
  Rechner, aber über zwei verschiedene Ursprünge (Loopback und LAN). Ein
  physisches Handy im WLAN und ein zweiter PC stehen noch aus.
- **Keine Freundesliste:** Der Connector kennt Anwesenheit nur **innerhalb eines
  Raums**; einen globalen Online-Status oder Freunde gibt es nicht. Deshalb
  zeigt die Halle keine „Freunde online“, sondern lädt über Link/Code ein –
  eine erfundene Freundesliste wurde bewusst nicht gebaut.
- **Bereitschaft/Start** sitzen seit 1.6 zusätzlich in der Hallen-Kopfzeile
  (dieselben Endpunkte, keine zweite Lobby); in der Spielseite bleiben sie
  unverändert. Das gilt für Spiele mit `players`/`spieler`/`seats` – Mau Mau
  liefert `seats` und war bis 24.11. ohne Knöpfe (siehe 9.2).
- **Chat ist Raumchat**, kein eigener Kanal je Partie. Für „zusammen spielen“
  reicht das, für mehrere parallele Tische im selben Raum nicht.
- Die App wurde neu gebaut, aber **nicht neu installiert/getestet**.
- Kein Versionssprung der APK (Regel: nur mit Zustimmung).

---

## 7. Nächste sinnvolle Schritte

1. Drei weitere Online-Spiele aus verschiedenen Arten durchspielen:
   Kniffel (tisch), Schach (tisch), Galgen (einladung) – mit demselben
   Zwei-Ursprünge-Test.
2. Ein echtes Handy im WLAN als zweiten Client nehmen (Connector bindet auf
   `:port`, LAN-Adresse funktioniert bereits).
3. Optional: Tisch-Chat je Partie (Raum + Spielkennung) – dafür müsste der
   Connector Nachrichten mit Spielkennung filtern; heute geht das nicht.
4. APK auf dem Gerät installieren und die drei Ansichten der App prüfen.
5. Erst danach Version anheben und veröffentlichen (ausdrücklich auf Zuruf).

---

## 8. Belege (gemessen)

- Katalog: 94 Spiele, 6 Kategorien; 71 Register-Einträge.
- Symbole: 94/94 Karten; Würfelspiele 25/25.
- Zwei Ursprünge: Raum TG-6348, 2 Spieler, Chat 9 Beiträge, Start erfolgreich,
  0 Konsolenfehler.
- Handy 390 px: Start/Kategorie/Highlights je 390 = 0 px Querscrollen.
- APK: `spiele/android/tmahjong.apk`, **14.676.485 B**, md5
  `ee604ba10e5f8185253c99f9aa981a71`; Version 1.16 / Code 17 (Stand 24.11.,
  vorher 14.676.385 B / `abee9417…`).

---

## 9. Fortsetzung 24.11.2026: Angebot entscheidet der Server, drei weitere Spiele

### 9.1 Offene Runde erkennt jetzt der Connector

Das Hallen-Angebot blieb stehen, wenn die Partie auf der **Spielseite**
gestartet wurde; bisher räumte nur die Halle ihr eigenes Angebot auf. Jetzt
prüft `GET /api/angebot` den Stand selbst, bevor es etwas liefert:

- `spielStatus(spiel)` liest den Zustand der Runde im Raum: 33 typisierte
  Tischspiele über ihr eigenes Feld, die 18 Würfelspiele über `r.Wuerfel`
  **mit Vergleich auf `Spiel`** (sonst würde ein fremdes Würfelspiel zählen).
- `angebotNochOffen(a)`: Malen bleibt offen (gemeinsame Fläche ohne Plätze),
  Einladungsspiele bleiben offen (sie erledigen sich beim An-/Ablehnen),
  Tischspiele nur bei Status `invited` – genau der Lobby, in die `join`
  aufnimmt. Alles andere (`active`, `spiel`, `runde`, `anschauen`, `ende`,
  `finished`) oder eine fehlende Runde heißt: Angebot weg.
- `action=annehmen` auf eine gestartete Runde antwortet **409**
  „Die Runde ist schon gestartet – hier ist nichts mehr offen.“ statt in einen
  laufenden Tisch zu greifen.
- Die Aufräum-Zeile der Halle blieb stehen, ist aber nur noch zweite
  Absicherung.

Geprüft: neuer Test `angebot_offen_test.go` spielt **alle 51 Tischspiele**
durch den echten Weg (bieten → Angebot da → beitreten → bereit → starten →
Abruf leer, zweiter Abruf bleibt leer) und deckt „Angebot ohne Runde“ ab.
Live mit dem neu gebauten Connector (Port 8768, Raum TG-95CA): bieten 204,
Angebot `galgen` sichtbar, Ben beitreten 204, beide bereit 200, Start 200,
Spielstatus `spiel` → `GET /api/angebot` = `null`, `annehmen` danach 409.
`go vet ./...` 0; `go test -run TestAngebot ./...` grün in allen vier
Ablagen (Homepage/GitHub × connector/connector-online).

### 9.2 Gefundener Fehler: Mau Mau ohne Bereit/Start in der Halle

Mau Mau liefert seine Sitzplätze als **`seats`** (`id`, `name`, `count`,
`ready`); `tischSpieler()` in `halle-online.js` kannte nur `players` und
`spieler`. Folge: In der Hallen-Kopfzeile standen bei Mau Mau **0 Knöpfe** auf
beiden Seiten, obwohl die Spielseite Bereit und Start kennt.

Behoben mit einem dritten Zweig (`seats`, Kommentar im Code). Gemessen:
vorher `#tisch-aktionen` = 0 Knöpfe bei A und B; nachher A
„Bereit / Erst alle bereit“, B „Bereit“, „Spiel starten“ freigegeben, Start
200, danach „Partie läuft · 2 Spieler.“ auf beiden Ursprüngen.

### 9.3 Drei Spiele über die Halle, zwei Ursprünge

A = `http://127.0.0.1:8768`, B = `http://192.168.56.1:8768` (LAN-Adresse des
Rechners), je eigener Browser-Kontext. Weg: A öffnet den Raum, sucht das Spiel,
wählt „Online-Runde“; B tritt über den Einladungslink bei und nimmt die Karte
„Mitspielen“; Bereit/Start in der Hallen-Kopfzeile (wo es sie gibt); dann ein
**echter Zug** und Chat in beide Richtungen.

| Spiel | Raum | Start | Zug (gemessen) | Stand danach |
|---|---|---|---|---|
| Mau Mau | TG-29B2 | Bereit/Start in der Halle 200 | Karte 10♥ gespielt | Anna 5 → 4 Karten, Ablage 10♥, „Bert ist am Zug“, beide „Partie läuft · 2 Spieler.“ |
| Backgammon | TG-D587 | Einladung, Start beim Beitritt (kein Bereit/Start) | „WÜRFELN“ gedrückt | „Würfel liegen bereit“, „WÄHLE WÜRFEL, STEIN UND ZIELFELD“, B sieht denselben Wurf |
| Domino | TG-F6BD | Bereit/Start in der Halle 200 | ersten Stein angelegt (7 → 6) | Brett „LINKS 0 / RECHTS 4“, „Bert ist dran“, beide „Partie läuft · 2 Spieler.“ |

Chat: „Hallo von Anna …“ und „Hallo von Bert …“ kamen auf der jeweils anderen
Seite an (je Raum ein Beitrag mehr). **0 Konsolen-/Ladefehler** auf beiden
Seiten in allen drei Partien.

Beobachtung ohne Fehler: Backgammon ist ein Einladungsspiel – die Halle zeigt
vor dem Beitritt „Warte auf Mitspieler …“ (es gibt noch keine Runde). Danach
stand dort zunächst nur „Partie läuft.“ **ohne Spielerzahl**, weil der Stand nur
`redId`/`blueId` führt und keine Spielerliste. Das ist behoben und in 9.5
gemessen.

### 9.4 Geänderte Dateien und Belege dieser Fortsetzung

- `spiele/connector/angebot.go`, `spiele/connector-online/angebot.go` und die
  beiden GitHub-Kopien – byte-gleich, md5 `d6be6e1f641446d34056a98bdb04e8b6`.
- `spiele/connector/angebot_offen_test.go` (neu, nur lokal – Tests bleiben aus
  dem Repo).
- `spiele/halle-online.js` (Zweig für `seats`; Zähler `tischSeiten` für Spiele
  ohne Spielerliste), byte-gleich in Homepage, GitHub und beiden
  `android/app/src/main/assets/www/spiele`, jetzt md5
  `fe0c4f46efd327eeb1174dcec0a65fcb`, **59.791 B** (Stand direkt nach 9.3:
  `4228ca8a16379c7b67298c3926221bdc`, 59.174 B).
- Binaries neu gebaut: `trashygecko-connector.exe` `b59f0ebfe6bf3c807dacbd301c47f20a`,
  `trashygecko-connector-online.exe` `357755cbadfe09ecb01dec8551967d0e`,
  `trashygecko-connector-online-linux` `a58d52764c1cf84ea898522911876f7e`
  (je Homepage und GitHub gleich).
- APK neu: **14.676.694 B**, md5 `769b2027de7719a1906a459c8f891a01`, Version
  1.16 / Code 17 (unverändert, in beiden Ablagen gleich);
  `assets/www/spiele/halle-online.js` im APK md5-gleich (`fe0c4f46…`, 59.791 B).
- `stand_pruefen.py`: Binaries gegen Go-Quellen 8× ok, APK-Prüfung ok; die 22
  Baum-Unterschiede zwischen Homepage und GitHub sind vorbestehende Reste
  (Screenshots, `sicherung/…`, `start.json`, `quiz-server.log`) und nicht aus
  dieser Fortsetzung.

### 9.5 Spielerzahl in der Backgammon-Kopfzeile

Der Stand von Backgammon führt keine Spielerliste, nur die beiden Seiten
(`redId`, `blueId`). Für die Zeile „Partie läuft“ zählt die Halle deshalb
zusätzlich diese beiden Felder (`tischSeiten`). Die Zahl kommt weiterhin zuerst
aus `tischSpieler()` – `players`, `spieler` und `seats` bleiben unverändert –,
der neue Zähler greift nur, wenn dort nichts steht, und nur im Zweig „Partie
läuft“, nicht im Einladungstext.

Gemessen mit zwei Ursprüngen (A `127.0.0.1:8768`, B `192.168.56.1:8768`, je
eigener Browser-Kontext). Backgammon: Einladung über
`POST /api/backgammon?action=invite`, Beitritt über `accept`; Galgen und Mau Mau
über `create`/`join`/`ready`/`start` (alle 200).

| Prüfung | Stand | Kopfzeile A | Kopfzeile B |
|---|---|---|---|
| Backgammon, Einladung (Raum TG-8CEA, Spiel TG-CB17) | `status=invited`, nur `redId`/`blueId` | „Einladung offen – im Chat oder im Spiel annehmen.“ (unverändert) | gleich |
| Backgammon nach `accept` | `status=active`, `redId=kopf-a`, `blueId=kopf-b` | **„Partie läuft · 2 Spieler.“** | **„Partie läuft · 2 Spieler.“** |
| Galgen (Raum TG-7B92, Spiel TG-276B, Feld `spieler`) | `status=spiel`, 2 Spieler | „Partie läuft · 2 Spieler.“ | „Partie läuft · 2 Spieler.“ |
| Mau Mau (Raum TG-DAFF, Spiel TG-E1AA, Feld `seats`) | `status=active`, 2 Plätze | „Partie läuft · 2 Spieler.“ | „Partie läuft · 2 Spieler.“ |

Die Stand-Felder von Backgammon sind `id, redId, redName, blueId, blueName,
points, bar, off, player, dice, rolled, status, multiplier, note` – keine
Spielerliste. **0 Konsolen-/Seitenfehler** in allen vier Prüfungen.

### 9.6 Weiter offen
- Alle Messungen auf **einem** Rechner über zwei Ursprünge; ein echtes Handy im
  WLAN und ein zweiter PC stehen weiter aus.
- Die Partien liefen auf einem lokalen Testserver (nach dem Test beendet),
  nichts committet oder veröffentlicht.

## 10. Sichtbare Spielhalle (24.11.2026)

### 10.1 Warum die neuen Funktionen nicht zu sehen waren

Der Connector auf Port 8766 liefert `Documents/Homepage` direkt aus. Dort lag
die Katalogfassung vom 23.11. (61.687 B): **kein** `<link>` auf `halle.css`,
**kein** `<script>` für `halle-online.js`, und die Spielkarten zeigten
Buchstabenkürzel (`monogramm`) statt der SVG-Zeichen. Die Online-Halle, die 45
Spielzeichen und die Kategorie-Navigation steckten nur in der Rework-Fassung.
Dazu ein echter CSS-Fehler: `@media(max-width:420px){.hl-badge.online{display:none}}`
blendete das Online-Abzeichen auf Handys aus.

### 10.2 Was geändert wurde

- Basis ist die Fassung mit allen Funktionen (`sicherung/spielhalle-sichtbar-2026-11-24/spiele-rework.html`),
  die Darstellung folgt wieder der kompakten 23.11-Anordnung: vier Highlights
  nebeneinander ab 980 px, darunter zwei; am Handy 2 × 2.
- Highlight-Bilder **287 × 155** (PC) und **176 × 99** (Handy) statt 586 × 330;
  Startseite **1384 px** statt 2051 (Handy **1789** statt 2674).
- Spielkarten: 46-px-Symbolplatte mit Online-Punkt, Name 15 px, Beschreibung auf
  zwei Zeilen begrenzt, Aktion („Modi ›“ / „Spielen ›“) oben rechts, mehr Luft.
- Modus-Chips verkürzt (Solo, Lokal, Online, Lokal & Online, Bots, Rekord, Pad),
  der ganze Satz steht als `title` an der Karte.
- Online-Abzeichen auch unter 420 px sichtbar.

### 10.3 Ansichten gemessen (Chromium, echter Connector)

| Ansicht | 1280 × 800 | 390 × 844 |
|---|---|---|
| Startseite | 4 Highlights, 6 Kategorie-Kacheln, Seite 1384 px | 4 Highlights (2 × 2), 6 Kacheln, Seite 1789 px |
| Würfelspiele | 25 Karten, 25 SVG | 25 Karten, 25 SVG |
| Kartenspiele | 11 / 11 | 11 / 11 |
| Brettspiele | 22 / 22 | 22 / 22 |
| Denkspiele | 15 / 15 | 15 / 15 |
| Arcade und Geschicklichkeit | 20 / 20 | 20 / 20 |
| Quiz und Partyspiele | 1 / 1 | 1 / 1 |
| Highlights-Ansicht | 4 große Karten | 4 Karten |
| Werkzeuge | 12 Einträge | 12 Einträge |

Zusammen **94 Spiele**, **94 von 94** Karten mit SVG aus **45** Zeichenformen,
0 Buchstabenkürzel bei Spielen (nur die Werkzeuge behalten Kürzel). Kein
Querscrollen (`scrollWidth` = Fensterbreite) in allen Zeilen.

Suche „Kniffel“ → 1 Karte; Filter „Einzelspieler“ in den Würfelspielen →
„16 von 25 Spielen passen zum Filter“; Direktlink `#/k/wuerfel` → 25 Karten;
Grafikstile Klassisch / Mädchen / Buben je 6 Kategorie-Symbole, umgefärbt.
**0 Konsolen-/Seitenfehler** in allen Ansichten.

### 10.4 Online-Halle durchgespielt (Raum TG-1D83, Galgenmännchen)

Zwei Browser-Kontexte: A `1280 × 800`, B `390 × 844`, beide auf dem echten
Connector. A eröffnet über die Spielkarte „Online-Runde“, B tritt über die
Einladung bei, beide drücken Bereit, A startet.

| Prüfung | Messwert |
|---|---|
| Eigene Runde in der Leiste | Knopf „Runde: Galgenmännchen“, Raumcode TG-1D83, Avatare beider Spieler |
| Einladung bei B | Karte „… lädt ein · Galgenmännchen · 2–6 Spieler“ mit „Mitspielen“ / „Ablehnen“, sichtbar |
| Bereit | A „Bereit“ + gesperrtes „Erst alle bereit“, B „Bereit“; Stand „2 am Tisch · 0 bereit“ |
| Start → Status | „Partie läuft · 2 Spieler.“ bei A nach **69 ms**, bei B nach **569 ms** |
| Tisch A | 1280 × 800, Rahmen 1280 × 744, `galgen.html?room=TG-1D83…` |
| Tisch B | 390 × 844, Rahmen 390 × 795, derselbe Pfad |
| Spiel im Rahmen (beide) | Titel „Galgenmännchen“, 37 Tasten, Zeichnung vorhanden, Text „Läuft / _ / _ / _“ |
| Chat A → B | Nachricht sichtbar nach **452 ms** |
| Chat B → A | Nachricht sichtbar nach **1118 ms** (einschließlich Öffnen des Chat-Feldes bei B) |
| Verläufe | beide Seiten zeigen „Hallo von A“ (09:19:14) und „Hallo von B“ (09:19:16) |

**0 Konsolen-/Seitenfehler** auf beiden Seiten.

### 10.5 Geänderte Dateien und Belege

- `spiele/spiele.html`, byte-gleich in Homepage, GitHub und beiden
  `android/app/src/main/assets/www/spiele`: md5 `69395b907f8ee0face64a09c43cb0749`,
  58.832 B. Im APK: md5 `2a9b04efdbe1ccdcc2b46cd61d700979`, 58.563 B – der Bau
  nimmt die drei gerätefremden Werkzeug-Zeilen (APK-Links, WLAN-Fernbedienung)
  heraus.
- `spiele/halle.css`: md5 `c5196ca05ecf90f5d3fb1e67d67c846a`, 47.128 B,
  byte-gleich in allen vier Ablagen.
- `spiele/halle-online.js` unverändert: md5 `fe0c4f46efd327eeb1174dcec0a65fcb`.
- APK neu: **14.676.865 B**, md5 `7a2a7857ce8af6c36a10070271b36440`, Version
  1.16 / Code 17 unverändert, beide Ablagen gleich.
- Sicherung vor den Änderungen: `sicherung/spielhalle-sichtbar-2026-11-24/`
  (alte Katalogseite, Rework-Fassung, alte `halle.css`).
- `stand_pruefen.py`: weiter 22 vorbestehende Baum-Unterschiede, keine neuen.

### 10.6 Weiter offen

- Alle Messungen auf **einem** Rechner (zwei Browser-Kontexte); echtes Handy im
  WLAN und zweiter PC stehen weiter aus.
- Chat nur mit zwei Teilnehmern geprüft; Gruppenchat mit mehreren und
  Flüsternachrichten nicht.
- Nichts committet oder veröffentlicht.

## 11. Echtes Handy im WLAN (24.11.2026)

Nachtrag zu 10.6: Der Punkt "echtes Handy im WLAN" ist erledigt, ein zweiter
Rechner war nicht verfügbar.

### 11.1 Gerät und Weg

- Huawei FIG-LX1, Android 9, Bildschirm 720 x 1440 (CSS 360 x 720), Firefox
  (Gecko) über `adb` geöffnet.
- WLAN `WLAN-502882`, Ping vom Handy zum PC 11 ms (min 8,8 / max 13,5).
- Adresse `http://192.168.2.106:8766/spiele/spiele.html`, also derselbe
  Connector, der die Seite ausliefert (kein Testserver, keine Kopie).

### 11.2 Was auf dem Handy sichtbar war

- Startseite: Kopf "Spielhalle · 94 Spiele · 6 Kategorien", Suche, Online-Karte,
  Highlights zwei nebeneinander, sechs Kategorie-Kacheln mit Symbol und Zahl
  (25, 22, 15, 1, 20, 11).
- Kategorie "Brettspiele": 22 Spiele, Beschreibung, Filter "Einzelspieler",
  "Mehrspieler", "Online spielbar", "Gegen Bots"; Karten mit SVG-Zeichen.
- Raum: Einladungskarte "GAST H2T LÄDT EIN · Galgenmännchen · Mitspielen /
  Ablehnen"; nach "Mitspielen" Tischkopf mit Bereit / Chat / Zurück zur Halle,
  Status "Warte auf Mitspieler"; nach Bereit auf beiden Seiten
  "Partie läuft · 2 Spieler.", Spiel im Rahmen (Galgen, Wort, Tastatur).
- Chat: Handy sendet "Hallo vom Handy" 09:44:44, im PC-Verlauf als
  "Gast 0B0 · Hallo vom Handy" sichtbar; PC antwortet "Antwort vom PC", auf dem
  Handy 09:44:46 im Raumchat (8 Beiträge).

### 11.3 Gefundener Fehler und Behebung

- Fehlerbild: Im Spielrahmen lagen Tastaturzeilen und Spielerkarte übereinander;
  der Kartentext schien durch die Buchstaben (H–N, O–U, V–Z). Auf dem Handy
  klar sichtbar, in Chromium bei 390 x 844 nicht.
- Ursache: Die Medienregeln für flache Geräte standen in `galgen.html` **vor**
  den Grundregeln von `.zeichnung` und waren damit wirkungslos; `.mitte` hatte
  `min-height:0`, sodass die Flexbox unter die Inhaltshöhe schrumpfte, ohne dass
  die Seite scrollen konnte.
- Änderung (nur CSS in `spiele/brettspiele/galgen/galgen.html`): Medienregeln
  hinter die Grundregeln verschoben, neue Regel für Hochformat mit wenig Höhe
  (`.zeichnung` schrumpft auf `clamp(90px,22vh,170px)`), `.mitte` auf
  `min-height:auto`, Seite scrollt bei Bedarf, `.app{min-height:100%}`.
- Messung (Tastatur-Unterkante gegen Spielerkarten-Oberkante):
  360 x 560 vorher 59 px Überlappung, nachher 0; 360 x 480 vorher 55 px,
  nachher 0 (Seite scrollt, 546 px); 360 x 640, 390 x 844 und 720 x 400
  unverändert 0; Desktop 1280 x 800 unverändert zweispaltig.
- Auf dem Handy danach nachgesehen: Buchstaben, "Kategorie: Natur",
  "Runde 1 · Klassisch · Du bist dran" und "Gast H2T · 0 Punkte" liegen sauber
  untereinander, keine Überlagerung.
- APK neu gebaut (14.677.008 B, md5 `a437cde99b4d772691e6391f351d7308`); im
  Paket liegt `galgen.html` mit 29.124 B, md5 `821f6d943a3611f65055e1d538056ca4`,
  identisch in beiden Ablagen, GitHub-Kopie und beiden Asset-Kopien.

### 11.4 Grenzen dieses Tests

- Ein zweiter Rechner war nicht verfügbar; der Gegenspieler lief in Chromium auf
  demselben PC (eigener Browser-Kontext, eigener Teilnehmer "Gast H2T").
- Getestet in Firefox auf dem Handy; Huawei-Browser und Chrome auf dem Gerät
  nicht. Gruppenchat mit mehr als zwei Teilnehmern und Flüsternachrichten nicht.
- Nichts committet oder veröffentlicht.

## 12. Weitere Spiele im Hallen-Rahmen am Handy (24.11.2026)

Geprüft auf demselben Gerät (Huawei FIG-LX1, Firefox 160.0.1 und Chrome,
WLAN, 360 x 720 CSS): Mau Mau, Backgammon, Domino, Mahjong. Rahmen wie bei
Galgenmännchen 360 x 671 (Kopf 49 px).

### 12.1 Überlappungen: keine

Gemessen im Rahmen (Oberkante/Unterkante in CSS-Pixeln):

- Mau Mau (online): Inhalt **850 px** hoch; Spielerfeld bis 500,
  Ablagestapel 300–430, Handkarten **540–678**. Nichts liegt übereinander.
- Backgammon (online): Inhalt **1131 px**; Brett **45–500**, Würfel
  **508–574**, Spielerpanel **667–796**, Anleitung **847–1002**, Aktionen
  **1009–1129**. Nichts liegt übereinander (zweispaltig ab 820 px).
- Domino (online): Inhalt **671 px**, passt genau; Kette/Kappen 120–390,
  eigene Hand **412–655**. Keine Überlappung.
- Mahjong: hat nur den Modus „solo“ und erscheint deshalb **gar nicht** im
  Hallen-Rahmen (die Halle bietet dafür keine Online-Runde an). Die Seite
  selbst ist bei 360 x 671 genau 671 px hoch, Spielfeld 608 px – kein
  Überlauf, keine Überlappung.

Zusätzlich in Chromium bei 360 x 640, 360 x 480, 390 x 844 und 1280 x 800
geprüft (Paarvergleich aller sichtbaren Blöcke): **0 Überlappungen** in allen
vier Spielen. Der Galgen-Fall (Tastatur über Spielerkarte) tritt hier nicht
auf, weil diese Spiele keine eingebettete Tastatur im selben Fluss haben.

### 12.2 Gefundener Fehler: unterer Teil im Rahmen unerreichbar

- Bei Mau Mau liegen die **Handkarten ab 678 px** (Rahmen endet bei 671), bei
  Backgammon **alles ab 671 px** (Würfeln, Spielstand, Zug-Hilfe, Regeln).
- Wischen im Rahmen scrollt **nicht**: geprüft mit je 3–4 Wischgesten an
  verschiedenen Stellen und Dauern (400–900 ms, x = 360 und 690), in Firefox
  160.0.1 **und** in Chrome, auch nach `force-stop` und frischem Laden. Der
  Inhalt bewegt sich nicht (Bildvergleich vor/nach).
- Kontrolltest: dieselbe Seite in einer Testseite mit einfachem iframe
  (671 px hoch, ohne Halle; Testseite wieder gelöscht) **scrollt** mit einer
  Wischgeste; die Spielseiten
  selbst können also scrollen. Im Online-Modus (`?room=…`) blieb derselbe
  Kontrolltest ohne Bewegung.
- Zwei Behebungsversuche (Rahmenhöhe = Inhaltshöhe + scrollende Körperspalte;
  Tisch am Handy auf Seiten-Ebene statt `position:fixed`) wirken am Rechner
  (Firefox: iframe 1131 px, Dokument 1180 px, `overflow-y:auto`), **am Gerät
  nicht** (Bild unverändert, kein Scrollen). Beide sind wieder
  **zurückgenommen**: `halle-online.js` md5 `fe0c4f46efd327eeb1174dcec0a65fcb`,
  `halle.css` md5 `c5196ca05ecf90f5d3fb1e67d67c846a` – der ausgelieferte Stand.
- Offen bleibt damit: **Wischgesten erreichen in der Halle den Spielinhalt
  nicht** (betrifft Mau Mau und Backgammon, vermutlich jede Seite, die höher
  als der Rahmen ist). Nächster sinnvoller Schritt wäre, die Spiele im Rahmen
  auf die Rahmenhöhe zu bringen (wie beim Galgen) oder den Tisch am Handy als
  eigene Seite zu öffnen – beides ist noch nicht umgesetzt.

### 12.3 Stand

- Am Galgen bleibt die Korrektur aus Abschnitt 11.3 (verifiziert auf dem
  Gerät); Mau Mau, Backgammon, Domino und Mahjong sind **unverändert**.
- APK unverändert (14.677.008 B, md5 `a437cde99b4d772691e6391f351d7308`).
- Bilder: `arbeit/screens/handy/33-maumau-tisch.png`,
  `…/35-maumau-partie.png`, `…/46-bg-tisch.png`, `…/69-bg-fertig.png`,
  `…/66-bg-oben.png`, `…/14-kategorien.png`.
- Nichts committet oder veröffentlicht.

## 13. Ursache des mobilen Scrollproblems gefunden (24.11.2026)

**Befund:** Auf dem Huawei FIG-LX1 lösen Wischgesten, die in einem **iframe**
beginnen, in **keinem** Fall ein Scrollen des Rahmeninhalts aus – weder in
Firefox 160 noch in Chrome. Das ist unabhängig vom Inhalt und von allen
CSS-Stellschrauben.

**Messung am Gerät (360 CSS breit):**

| Prüfung | Ergebnis |
| --- | --- |
| Rahmeninhalt programmatisch scrollen | `300 → 300` von 591 px möglich |
| Wischen im Rahmen (Spielseite) | `touchstart` kommt an, **0** `touchmove`, **0** `touchend`/`touchcancel`, `scrollTop` bleibt 0 |
| Wischen im Rahmen (lange Textseite, 42.207 px) | identisch: `top=0` |
| Trefferkette unter dem Finger | `rect > svg#board > .board-column > .board-frame > section#game`, alle `touch-action: auto`/`manipulation`, `pointer-events: auto` |
| Normale Seite (gleiche Seite auf Seitenebene) | scrollt einwandfrei |

**Bisektion – 11 Varianten, alle ohne Scrollen im Rahmen:** festes Overlay mit
und ohne `body{overflow:hidden}`, statisches Overlay, iframe direkt im Fluss,
iframe absolut positioniert, Grid- vs. Block-Container, `touch-action:auto` vs.
`pan-y` auf Hülle/Körper/iframe, Spielseite vs. lange Textseite, Firefox und
Chrome. Jeweils: `touch-action=… fix=… | innen 42207/540 top=0`.

**Gegenprobe am Rechner:** Chromium 360 × 720 mit echten CDP-Touch-Events
scrollt dieselbe Halle (`scrollTop 0 → 445`). Der Fehler ist also geräte- bzw.
browserseitig im iframe-Kontext, **nicht** in CSS oder JavaScript der Halle
oder der Spiele. `overflow`, `touch-action`, `overscroll-behavior`,
`position:fixed`, feste Höhen, Pointer-/Touch-Handler und unsichtbare Overlays
sind als Ursache ausgeschlossen.

**Lösung (zweite, vom Auftrag erlaubte Möglichkeit):** Am Handy
(`max-width:760px`) öffnet die Halle den Tisch nicht mehr im iframe-Overlay,
sondern lädt die Spieladresse auf **eigener Seitenebene**
(`halle-online.js`, in `tischOeffnen` vor dem Overlay-Zweig). Am Rechner bleibt
alles unverändert. Der zuletzt offene Tisch wird weiter in `tg-halle-tisch`
gemerkt; die Zurück-Taste des Geräts führt in die Halle.

**Am Gerät geprüft:** Tippen auf „Partie fortsetzen“ öffnet die Spielseite
bildschirmfüllend ohne Hallen-Kopf, ohne Überlappung, Bedienknöpfe sichtbar
(Bild `arbeit/screens/handy/176-direkt.png`); Seitenebene scrollt (Beleg
`…/160-seite-vor.png` → `…/161-seite-nach.png`); Mahjong läuft auf Seitenebene
vollständig sichtbar (`…/180-mahjong-vor.png`).

**Nicht geprüft (ehrlich):** eine laufende, dem Handy eigene Partie in Mau Mau,
Backgammon oder Domino auf Seitenebene – die Testpartie gehörte dem Rechner,
das Handy wurde nach wenigen Sekunden als Nicht-Teilnehmer zum Connector
zurückgeleitet. Das Durchwischen einer eigenen Runde muss am Gerät noch
bestätigt werden.

**Stand:** `spiele.html`, `halle.css` unverändert; `halle-online.js` um den
Handy-Zweig ergänzt (beide Ablagen und Android-Assets gleich, SHA256
`6f2e51244247…`); APK neu gebaut, 14.676.879 B, md5 `3b293939ba`.
Sicherung: `Homepage/sicherung/scroll-2026-11-24/`. Nichts committet.
