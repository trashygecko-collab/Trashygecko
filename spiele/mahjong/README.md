# TrashyGecko Mahjong

Mahjong-Solitär mit 56 Formen, 72 Motiven und deutschen Sprüchen. Jedes Level hat
genau 144 Steine (72 Paare) und ist garantiert lösbar. Nach der letzten Form
beginnt die nächste Runde mit Form 1 – in einer anderen Steinfarbe, endlos.

## Starten

Es gibt keinen Build und keine Installation – reine statische Dateien.

- **Direkt öffnen:** `mahjong.html` im Browser (funktioniert auch über `file://`,
  deshalb liegen die Sprüche als JavaScript-Dateien vor).
- **Lokaler Server (empfohlen, z. B. für das Handy im WLAN):**
  im Ordner `spiele` (eine Ebene über `mahjong`) starten:
  ```bash
  python -m http.server 8765
  ```
  Danach `http://localhost:8765/mahjong/mahjong.html` aufrufen. Liegt der
  Server eine Ebene höher (in `Homepage`), heißt die Adresse
  `http://localhost:8765/spiele/mahjong/mahjong.html`.

Der Spielstand (Level, Punkte, Bestzeiten) liegt im `localStorage` des Browsers.

## Steuerung

| Aktion | Bedienung |
|---|---|
| Stein wählen | Tippen / Klick auf einen freien Stein |
| Paar abräumen | zweiten Stein mit gleichem Motiv antippen |
| Brett verschieben | mit einem Finger wischen bzw. mit gedrückter Maustaste ziehen |
| Zoomen | zwei Finger auseinander/zusammen oder Mausrad |
| Gesamtansicht | Doppelklick / Doppeltipp auf eine freie Stelle |
| Level wählen | auf `LV 37 ▾` tippen, dann ein Feld im Raster; unten eine Farbe antippen, um in deren Runde zu springen |
| Tipp | Knopf `TIPP` markiert ein freies Paar – kostet 50 Punkte |
| Neu mischen (nur wenn kein Zug geht) | Knopf `NEU MISCHEN` – einmal pro Level gratis |
| Neu mischen | Knopf `NEU` startet das aktuelle Level neu |
| Vollbild | Knopf `⛶` |

Der Startzoom zeigt immer das ganze Muster; das Feld nutzt den gesamten Platz
zwischen Kopfzeile und Bildschirmrand (1 px seitlich, 3 px oben, 2 px unten –
Notch/Safe-Area bleibt zusätzlich frei).

## Spielablauf und Punkte

- Ein Stein ist frei, wenn nichts darauf liegt und links oder rechts daneben
  kein Stein mehr steht. Nicht erreichbare Steine sind deutlich dunkler.
- **Gleiches Paar:** +100 Punkte, dazu ein zufälliger Spruch aus `lob.txt`. Der
  Spruch erscheint jedes Mal in einer anderen, gut lesbaren Farbe (Gold, Mint,
  Himmelblau, Rosa, Lavendel, Sand, Aqua, Apricot, nie zweimal dieselbe
  hintereinander); die Punkteangabe dahinter bleibt grau.
- **Zwei ungleiche Steine:** −100 Punkte, die beiden Steine blitzen kurz rot auf,
  die Punkteanzeige leuchtet rot und es kommt ein Spruch aus `pech.txt`.
  Bei **0 Punkten** hört der Spott auf: dann steht dort nur
  „Die 2 Steine passen nicht zusammen." – ohne Abzug und ohne rote Punkte.
- **Level geschafft:** Siegerehrung mit `Level · Runde · Farbe`
  (z. B. „Level 57 · Runde 2 · Bernstein"), Punkten, Spruch aus `gewonnen.txt`,
  Zeit, Highscore und „WEITER"-Knopf. Der Spruch wechselt bei jedem Sieg die
  Farbe (warme Palette: Orange, Gold, Amber, Coral, Rosa, Honig, nie zweimal
  dieselbe hintereinander). Am Ende zählen ein Zeitbonus
  (1000 − 2 pro Sekunde) und ein Abzug von 25 Punkten je Fehlgriff.
- **Tipp:** `TIPP` markiert ein freies Paar und kostet 50 Punkte – mit einem
  Spruch aus `pech.txt` in Rot und `- 50 Punkte` dahinter, wie beim Fehlgriff.
  Bei 0 Punkten bleibt es beim sachlichen Hinweis ohne Abzug.
- **Kein Zug mehr möglich:** Nach jedem Paar prüft das Spiel, ob überhaupt noch
  zwei freie Steine mit gleichem Motiv da sind. Wenn nicht, erscheint über dem
  Brett der Knopf `NEU MISCHEN`. Einmal pro Level ist das gratis: die Form bleibt,
  nur die Motive der restlichen Steine werden neu verteilt – wieder so, dass eine
  Abbaureihenfolge garantiert existiert. Ist das Mischen verbraucht, hilft nur
  noch `NEU` (Level neu starten).
- Der Timer pausiert, wenn der Tab im Hintergrund ist.

## Formen 1–56

Dieselben 56 Formen kommen in jeder Runde wieder, nur in einer anderen
Steinfarbe (siehe unten).

| Level | Form | Level | Form |
|---|---|---|---|
| 1 | Grundform | 29 | Schwan |
| 2 | Herzform | 30 | Ballerina |
| 3 | Jesuskreuz | 31 | Muschel |
| 4 | Teddybär | 32 | Eule |
| 5 | Stern | 33 | Gitarre |
| 6 | Katze | 34 | Schlüssel |
| 7 | Häschen | 35 | Anker |
| 8 | Schaf | 36 | Leuchtturm |
| 9 | Pinguin | 37 | Schloss |
| 10 | Blume | 38 | Sonnenblume |
| 11 | Pilz | 39 | Tulpe |
| 12 | Haus | 40 | Vogel |
| 13 | Fisch | 41 | Schnecke |
| 14 | Wolke | 42 | Vulkan |
| 15 | Kleeblatt | 43 | Kompass |
| 16 | Krone | 44 | Regenschirm |
| 17 | Schmetterling | 45 | Kamera |
| 18 | Rakete | 46 | Laterne |
| 19 | Schneemann | 47 | Pokal |
| 20 | Baum | 48 | Segelboot |
| 21 | Tasse | 49 | Marienkäfer |
| 22 | Kleid | 50 | Geburtstagstorte |
| 23 | Schleife | 51 | Würfel |
| 24 | Eistüte | 52 | Rechteck |
| 25 | Heißluftballon | 53 | Ball |
| 26 | Mondsichel | 54 | Pyramide |
| 27 | Herz mit Flügeln | 55 | Säulen |
| 28 | Kirschblüte | 56 | Brücke |

Die Level 51–56 sind bewusst einfache Formen: Würfel (drei gleiche Ebenen),
Rechteck, Ball, Pyramide, ein Feld aus zwölf Säulen und eine Brücke, deren Steg
in der Mitte über einer Lücke liegt.

## Runden und Steinfarben

Die Formen wiederholen sich, die Steinfarbe nicht: Level 1–56 sind Runde 1,
Level 57–112 sind dieselben 56 Formen in Runde 2, usw. Welche Form ein Level
spielt, ergibt sich aus `((Level − 1) mod 56) + 1`, die Runde aus
`floor((Level − 1) / 56) + 1`.

| Runde | Steinfarbe | ab Level |
|---|---|---|
| 1 | Elfenbein (hell, klassisch) | 1 |
| 2 | Bernstein (warm) | 57 |
| 3 | Rosé | 113 |
| 4 | Jade | 169 |
| 5 | Azur | 225 |
| 6 | Amethyst | 281 |
| 7 | Schiefer (dunkel) | 337 |
| 8 | Schwarz (dunkel) | 393 |
| 9 | Tiefschwarz (richtiges Schwarz) | 449 |

Nach Runde 9 beginnt es wieder mit Elfenbein (Runde 10 = Level 505). Das letzte
Level einer vollen Farbrunde (504, 1008, …) zeigt auf der Siegtafel zusätzlich
die Tafel „ALLE FARBEN DURCH" mit dem Hinweis, dass es wieder mit Elfenbein
weitergeht.

Bildet ein Level den Anfang einer neuen Runde (57, 113, …, 505), steht beim
Start in der Statuszeile: „Neue Runde 10 – Steinfarbe Elfenbein."

Die Farben sind zurückhaltend gewählt, damit die Motive gut lesbar bleiben;
die letzten drei sind dunkel, dort wird die Schrift auf den Steinen hell
(`--stein-text`) und die gesperrten Steine werden über Deckkraft statt über
Helligkeit abgedunkelt. Die Kopfzeile zeigt nur die Levelnummer; welche Runde
und Farbe läuft, steht in der Level-Auswahl über dem Raster.

Die Farbe steckt in CSS-Variablen am Stein (`--stein-hell`, `--stein-mitte`,
`--stein-rand`, `--stein-sockel`, `--stein-linie`, `--stein-text`), gesetzt wird
sie über `body[data-runde]` in `main.js`. Weitere Farben brauchen nur einen
zusätzlichen Block in `css/mahjong.css` und einen Namen in `RUNDEN_FARBEN`.

In der Level-Auswahl steht über dem Raster die gewählte Runde mit Farbname
und Fortschritt („Runde 9 · Tiefschwarz · 12 von 56"), darunter die 56 Formen
dieser Runde mit ihren Level-Nummern. Geschaffte Level tragen ein grünes
Häkchen und darunter ihre Bestzeit (mm:ss); die Tooltips der Farbpunkte nennen
den Fortschritt der jeweiligen Runde. Am Ende der Auswahl steht die Farbwahl – eine Farbe
antippen springt an den Anfang dieser Runde (Level 1, 57, 113, …).

## Aufbau der Dateien

```
mahjong/
├── mahjong.html          Grundgerüst: Kopfzeile, Spielfeld, Level-Auswahl, Siegerehrung
├── css/mahjong.css       gesamte Optik (auch Handy/Tablet und Querformat)
├── js/
│   ├── board-layout.js   alle 56 Formen als Positionen (x, y, Ebene)
│   ├── board.js          Regeln: freie Steine, Paare entfernen
│   ├── generator.js      verteilt die 72 Motive und sorgt für Lösbarkeit
│   ├── solver.js         prüft/ findet eine Abräum-Reihenfolge
│   ├── storage.js        Spielstand und Bestzeiten (localStorage)
│   ├── renderer.js       Brett zeichnen, Zoom, Verschieben, Blinkeffekte
│   ├── lob.js            432 Lob-Sprüche für ein gefundenes Paar
│   ├── pech.js           272 Sprüche für einen Fehlgriff
│   ├── gewonnen.js       Sprüche und Symbole für die Siegerehrung
│   └── main.js           Ablauf: Timer, Punkte, Level, Siegerehrung
└── assets/              72 Motive als WebP (160 × 160, verlustfrei)
```

## Technische Hinweise

- **144 Steine, garantiert lösbar:** `generator.js` baut das Brett so auf, dass
  eine vollständige Abräum-Reihenfolge existiert; `solver.js` kann das prüfen.
  Unlösbare Stellungen entstehen dadurch nicht. Durch eigene Züge kann man sich
  trotzdem festfahren; dafür gibt es das eine Gratis-Mischen pro Level, das die
  restlichen Steine über `findRemovalOrder` neu verteilt (die Funktion arbeitet
  auch auf einem angefangenen Brett).
- **Formen:** Jede Form beschreibt nur Positionen (`x`, `y`, Ebene `z`). Höhere
  Ebenen liegen versetzt darüber, dadurch wirken die Steine dreidimensional.
  Neue Formen kommen als weiterer Eintrag in `BOARD_LAYOUTS` dazu.
- **Steine:** 72 WebP-Dateien, verlustfrei aus PNG umgewandelt (2,72 → 2,06 MB,
  Pixel identisch). Im Startzoom sind sie je nach Form und Fenster rund
  38 × 45 px groß (gemessen bei 390 und 1280 px Breite), die 160 px Auflösung
  bleibt auch beim Hineinzoomen scharf.
- **Sprüche:** kommen aus `lob.txt` (alphabetisch sortiert), `pech.txt` und
  `gewonnen.txt`. Die JS-Dateien enthalten jeweils eine gemischte Ziehung ohne
  direkte Wiederholung.
- **Zwei Ablageorte:** Der Ordner liegt parallel unter
  `Documents\Homepage\spiele\mahjong` und `Documents\GitHub\Homepage\spiele\mahjong`.
  Änderungen gehören in beide, damit die Kopien gleich bleiben.
