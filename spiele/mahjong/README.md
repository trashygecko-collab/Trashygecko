# TrashyGecko Mahjong

Mahjong-Solitär mit 100 Formen, 72 Motiven und deutschen Sprüchen. Jedes Level hat
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
  `http://localhost:8765/spiele/brettspiele/mahjong/mahjong.html`.

Der Spielstand (Level, Punkte, Bestzeiten) liegt im `localStorage` des Browsers.

## Steuerung

| Aktion | Bedienung |
|---|---|
| Stein wählen | Tippen / Klick auf einen freien Stein |
| Paar abräumen | zweiten Stein mit gleichem Motiv antippen |
| Brett verschieben | mit einem Finger wischen bzw. mit gedrückter Maustaste ziehen |
| Zoomen | zwei Finger auseinander/zusammen oder Mausrad |
| Gesamtansicht | Doppelklick / Doppeltipp auf eine freie Stelle |
| Level wählen | auf `HANDY 37 ▾` tippen, dann ein Feld im Raster; oben zwischen HANDY-LEVEL und NORMAL-LEVEL umschalten, unten eine Farbe antippen, um in deren Runde zu springen |
| Tipp | Knopf `TIPP` markiert ein freies Paar in Türkis – kostet 200 Punkte, erst ab 200 Punkten |
| Vollbild | Knopf `⛶` |

Der Startzoom zeigt immer das ganze Muster; das Feld nutzt den gesamten Platz
zwischen Kopfzeile und Bildschirmrand (1 px seitlich, 3 px oben, 2 px unten –
Notch/Safe-Area bleibt zusätzlich frei).

## Spielablauf und Punkte

- Ein Stein ist frei, wenn nichts darauf liegt und links oder rechts daneben
  kein Stein mehr steht. Nicht erreichbare Steine sind deutlich dunkler, in
  Runde 7 (Schiefer) zusätzlich leicht durchscheinend. In den beiden schwarzen
  Runden 8 und 9 bleiben sie deckend, weil halb durchsichtige schwarze Steine
  vor dem dunklen Brett wie Löcher wirken: Runde 8 dunkelt nur sanft ab
  (Helligkeit 0,88) und gibt den freien Steinen einen helleren Glanz an der
  Innenkante, Runde 9 dunkelt kräftiger ab (0,78) und trennt zusätzlich über
  den Sockel.
- **Statusmeldung:** Spruch und Punktzahl werden als Textzeile **über dem
  Brett** eingeblendet (dunkler, halbtransparenter Kasten, `pointer-events:
  none`), statt Platz zu reservieren – das Brett nutzt die volle Höhe. Tipps
  gehen durch die Einblendung hindurch.
- **Gleiches Paar:** +100 Punkte, dazu ein zufälliger Spruch aus `lob.txt`; die
  Punktzahl steht in einer eigenen Zeile unter dem Spruch. Der
  Spruch erscheint jedes Mal in einer anderen, gut lesbaren Farbe (Gold, Mint,
  Himmelblau, Rosa, Lavendel, Sand, Aqua, Apricot, nie zweimal dieselbe
  hintereinander); die Punkteangabe dahinter bleibt grau.
- **Zwei ungleiche Steine:** −100 Punkte, die beiden Steine blitzen kurz rot auf,
  die Punkteanzeige leuchtet rot und es kommt ein Spruch aus `pech.txt`.
  Bei **0 Punkten** hört der Spott auf: dann steht dort nur
  „Die 2 Steine passen nicht zusammen." – ohne Abzug und ohne rote Punkte.
- **Stein wieder abwählen:** nochmal auf denselben Stein tippen hebt die Auswahl
  auf. Dafür kommt ein Spruch aus `pech.txt`, aber **ohne Punktabzug**, ohne
  Blinken und ohne Fehlgriff-Zählung – ein zweiter Tipp auf denselben Stein ist
  kein Fehler.
- **Level geschafft:** Siegerehrung mit `Level · Runde · Farbe`
  (z. B. „Level 101 · Runde 2 · Bernstein"), Punkten, Spruch aus `gewonnen.txt`,
  Zeit, Highscore und „WEITER"-Knopf. Der Spruch wechselt bei jedem Sieg die
  Farbe (warme Palette: Orange, Gold, Amber, Coral, Rosa, Honig, nie zweimal
  dieselbe hintereinander). Am Ende zählen ein Zeitbonus
  (1000 − 2 pro Sekunde) und ein Abzug von 25 Punkten je Fehlgriff.
- **Tipp:** `TIPP` markiert ein freies Paar mit türkisem Ring und pulsierendem
  Schein und kostet **200 Punkte** – mit einem Spruch aus `pech.txt` in Rot und
  `- 200 Punkte` dahinter, wie beim Fehlgriff. Unter 200 Punkten gibt es keinen
  Tipp mehr: Der Knopf `TIPP` ist dann ausgegraut und nicht drückbar
  (Tooltip „Tipp erst ab 200 Punkten").
- **Kein Zug mehr möglich:** Nach jedem Paar prüft das Spiel, ob überhaupt noch
  zwei freie Steine mit gleichem Motiv da sind. Wenn nicht, verteilt es die
  Motive der restlichen Steine still neu – ohne Knopf, ohne Meldung, ohne
  Punkteabzug. Die Form des Bretts und die Zahl der abgeräumten Steine bleiben,
  und die neue Verteilung geht garantiert wieder auf.
- Der Timer pausiert, wenn der Tab im Hintergrund ist.

## Formen 1–100

Dieselben 100 Formen kommen in jeder Runde wieder, nur in einer anderen
Steinfarbe (siehe unten). Die Level 1–56 sind unverändert; 57–100 sind 44 neue
Formen (siehe unten).

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
| 57 | Fahrrad | 79 | Apfel |
| 58 | Zug | 80 | Birne |
| 59 | Bus | 81 | Traube |
| 60 | Flugzeug | 82 | Erdbeere |
| 61 | U-Boot | 83 | Ananas |
| 62 | Fledermaus | 84 | Kürbis |
| 63 | Fuchs | 85 | Kuchen |
| 64 | Bär | 86 | Kanne |
| 65 | Elefant | 87 | Gabel |
| 66 | Giraffe | 88 | Schere |
| 67 | Pferd | 89 | Brille |
| 68 | Hund | 90 | Hut |
| 69 | Frosch | 91 | Schuh |
| 70 | Schildkröte | 92 | Tasche |
| 71 | Krake | 93 | Koffer |
| 72 | Wal | 94 | Bleistift |
| 73 | Delfin | 95 | Welle |
| 74 | Biene | 96 | Berg |
| 75 | Libelle | 97 | Zelt |
| 76 | Kaktus | 98 | Windmühle |
| 77 | Palme | 99 | Riesenrad |
| 78 | Blatt | 100 | Tempel |

Die Level 51–56 sind bewusst einfache Formen: Würfel (drei gleiche Ebenen),
Rechteck, Ball, Pyramide, ein Feld aus zwölf Säulen und eine Brücke, deren Steg
in der Mitte über einer Lücke liegt.

## Formen 2 – die Handy-Level

Der goldene Level-Knopf zeigt den laufenden Satz mit (`HANDY 01 ▾`); er öffnet die
Level-Auswahl. Darin schaltet der Umschalter oben zwischen **HANDY-LEVEL** und
**NORMAL-LEVEL** um: gleiche Runden, gleiche Steinfarben, gleiche Sprüche,
gleiche Punktzahl – nur **100 andere Formen** (6 × 8 statt 12 × 8). Jedes Spiel
beginnt mit den Handy-Leveln, auch wenn zuletzt ein Normal-Level lief.

- Jede Form ist **höchstens 6 Steine breit und 8 Steine hoch** (Handyformat) und
  hat trotzdem genau 144 Steine. Gestapelt wird immer genau übereinander, damit
  unten am Rand von jedem Stein ein Streifen sichtbar bleibt; die Tiefe ist frei.
- Die Formen stehen als kleine Bilder (8 Zeilen mit je 6 Zeichen, `#` = Stein)
  in `js/board-layout-2.js` (Level 1–56) und `js/board-layout-100.js`
  (Level 57–100); daraus werden die 144 Steine gleichmäßig gestapelt.
- Fortschritt, Häkchen und Bestzeiten der beiden Sätze sind **getrennt**: Satz 1
  liegt wie bisher ganz oben im Spielstand, Satz 2 in `satz2`.
- Der schwere Modus hat eigene Sammlungen: **100 Formen für den Rechner**
  (12 × 8, höchstens vier Ebenen) und **100 für das Handy** (6 × 8, höchstens
  sieben Ebenen). Beide bestehen aus 25 Umrissen zu je vier Höhenprofilen
  (Kuppe, Grat, Terrassen, Mulde); die Level 57–100 bringen elf neue Umrisse je
  Sammlung. Leicht und Schwer bleiben damit deutlich verschieden: hier liegen
  36 Motive zu je vier Steinen aus.

## Runden und Steinfarben

Die Formen wiederholen sich, die Steinfarbe nicht: Level 1–100 sind Runde 1,
Level 101–200 sind dieselben 100 Formen in Runde 2, usw. Welche Form ein Level
spielt, ergibt sich aus `((Level − 1) mod 100) + 1`, die Runde aus
`floor((Level − 1) / 100) + 1`.

| Runde | Steinfarbe | ab Level |
|---|---|---|
| 1 | Elfenbein (hell, klassisch) | 1 |
| 2 | Bernstein (warm) | 101 |
| 3 | Rosé | 201 |
| 4 | Jade | 301 |
| 5 | Azur | 401 |
| 6 | Amethyst | 501 |
| 7 | Schiefer (dunkel) | 601 |
| 8 | Weinrot (dunkel) | 701 |
| 9 | Tiefschwarz (richtiges Schwarz) | 801 |
| 10 | Regenbogen (alle Farben, jedes Paar eine) | 901 |

### Die Regenbogen-Runde

Nach dem letzten schwarzen Level kommt eine Zugabe: Level 901–1000 zeigen alle
neun Farben **gleichzeitig** – jedes Paar in einer eigenen Farbe. Die Farbe
kommt aus dem Motiv (`((Motivnummer − 1) mod 9) + 1`), und weil beide Steine
eines Paares dasselbe Motiv haben, passen sie immer farblich zusammen; es
werden genau 8 Paare je Farbe. In der Runde setzt nicht das Brett eine Farbe
(`body[data-runde]`), sondern jeder Stein selbst (`.mahjong-tile[data-farbe]`,
neun Paletten-Blöcke am Ende von `css/mahjong.css`).

In der Farbwahl steht sie als zehnter Punkt (Farbkreis) und führt zu Level 901.
Wer wenig sieht, kommt damit ohne Runden-Klettern direkt in die Runde, in der
jedes Paar eine eigene Farbe hat. Beim Start steht in der Statuszeile
„Runde 10 – alle Farben. Jedes Paar hat seine eigene Farbe."
Danach geht es mit Elfenbein weiter (Level 1001 = Runde 11), die Tafel „ALLE
FARBEN DURCH" erscheint jetzt am Ende der Regenbogen-Runde (Level 1000, 2000, …).

Bildet ein Level den Anfang einer neuen Runde (101, 201, …, 901), steht beim
Start in der Statuszeile der Rundenhinweis.

Die Farben sind zurückhaltend gewählt, damit die Motive gut lesbar bleiben;
die letzten drei sind dunkel, dort wird die Schrift auf den Steinen hell
(`--stein-text`). In Runde 7 bleiben die gesperrten Steine leicht durchscheinend,
in den dunklen Runden 8 und 9 sind sie deckend und nur abgedunkelt – sonst
wirken sie wie Löcher im Brett. Die Kopfzeile zeigt nur die Levelnummer; welche
Runde und Farbe läuft, steht in der Level-Auswahl über dem Raster.

Die Farbe steckt in CSS-Variablen am Stein (`--stein-hell`, `--stein-mitte`,
`--stein-rand`, `--stein-sockel`, `--stein-linie`, `--stein-text`), gesetzt wird
sie über `body[data-runde]` in `main.js`. Weitere Farben brauchen nur einen
zusätzlichen Block in `css/mahjong.css` und einen Namen in `RUNDEN_FARBEN`.

In der Level-Auswahl steht über dem Raster die gewählte Runde mit Farbname
und Fortschritt („Runde 9 · Tiefschwarz · 12 von 100"), darunter die 100 Formen
dieser Runde mit ihren Level-Nummern. Geschaffte Level tragen ein grünes
Häkchen und darunter ihre Bestzeit (mm:ss); die Tooltips der Farbpunkte nennen
den Fortschritt der jeweiligen Runde. Am Ende der Auswahl steht die Farbwahl – eine Farbe
antippen springt an den Anfang dieser Runde (Level 1, 101, 201, …).

Ganz oben in der Auswahl stehen zwei Umschalter: **FORM** (`RUND` / `ECKIG`)
und **ZUSTAND** (`NEU` / `ALT`). Sie wirken sofort auf das Brett und werden im
Spielstand unter `settings.steinForm` und `settings.steinZustand` gespeichert.
`ECKIG` setzt nur die Ecken des Steins auf 2 px (statt 8 px) und den inneren
Rahmen auf 0. `ALT` legt drei Ebenen über das Motiv: den Schmutz in den Poren (Ausbrüche am
Rand, `multiply`, 60 % Deckkraft, 0,5 px weichgezeichnet), feine Haarrisse
(`multiply`, 85 %, 0,5 px weich, per `mask-image` zur Mitte hin ausgedünnt –
echte Kratzer sitzen an Kanten und Ecken) und eine Körnung aus 50 % Grau mit
Rauschen (`stein-alt-korn.png`, `overlay`, 20 %), die dem Stein die raue
Oberfläche gibt, auf der die Risse erst echt wirken. Die Korn-Textur erzeugt
`werkzeuge/korn-erzeugen.mjs` (128 × 128, deterministisch, 11 kB). Jeder Stein
spiegelt die beiden Bild-Ebenen je nach Lage waagerecht und/oder senkrecht,
damit nicht auf allen Steinen dasselbe Muster klebt.
Die Risse nutzen `multiply` auf den hellen Runden und `screen` auf den dunklen,
weil sie sonst im Schwarz verschwinden; dort bleibt der Ausbruch deckend
(80 %), weil Multiplizieren auf Schwarz unsichtbar wäre. Die Ausbrüche gibt es dreifach:
warmbraun für Elfenbein, Bernstein und Rosé (`stein-alt-kanten-hell.webp`),
graugrün für Jade, Azur und Amethyst (`stein-alt-kanten-kuehl.webp`) und
tiefschwarz für die dunklen Runden (`stein-alt-kanten.webp`).

## Aufbau der Dateien

```
mahjong/
├── mahjong.html          Grundgerüst: Kopfzeile, Spielfeld, Level-Auswahl, Siegerehrung
├── README.md             diese Beschreibung
├── css/mahjong.css       gesamte Optik (auch Handy/Tablet und Querformat)
├── css/fonts/            Atkinson Hyperlegible 400/700 (woff2) + OFL-Lizenz
├── js/
│   ├── board-layout.js   Level 1–56 (PC, 12 × 8) als Positionen (x, y, Ebene)
│   ├── board-layout-2.js Level 1–56 des Handy-Satzes (6 × 8, gestapelt)
│   ├── board-layout-schwer.js schwere Formen 1–56 (PC 12 × 8, Handy 6 × 8)
│   ├── board-layout-100.js Level 57–100 aller vier Sammlungen
│   ├── board.js          Regeln: freie Steine, Paare entfernen
│   ├── generator.js      verteilt die 72 Motive, findet die Abräum-Reihenfolge
│   ├── renderer.js       Brett zeichnen, Zoom, Verschieben, Blinkeffekte
│   ├── hud.js            Punkte, Zeit, Meldungen und Sprüche
│   ├── dialoge.js        Level-Auswahl und Einstellungen
│   ├── groesse.js        Grundschrift je Bildschirmdiagonale
│   ├── storage.js        Spielstand und Bestzeiten (localStorage)
│   ├── pech.js           1379 Sprüche für einen Fehlgriff
│   ├── gewonnen.js       624 Sprüche und Symbole für die Siegerehrung
│   └── main.js           Ablauf: Timer, Punkte, Level, Siegerehrung
└── assets/              72 Motive als WebP (160 × 160, verlustfrei)
```

Die **Lob-Sprüche** liegen nicht bei den Spiel-Dateien, sondern zentral in
`spiele/lob.js` (`LOB_SPRUECHE` plus der Mahjong-Zusatz `LOB_MAHJONG`) und
werden von mehreren Spielen geteilt.

## Technische Hinweise

- **144 Steine, garantiert lösbar:** `generator.js` baut das Brett so auf, dass
  eine vollständige Abräum-Reihenfolge existiert; `findRemovalOrder()` findet sie
  und prüft damit das Brett.
  Unlösbare Stellungen entstehen dadurch nicht. Durch eigene Züge kann man sich
  trotzdem festfahren; dann verteilt `verteileMotiveNeu()` in `main.js` die
  restlichen Motive über `findRemovalOrder` neu (die Funktion arbeitet auch auf
  einem angefangenen Brett). Findet sie keine Reihenfolge, sagt die Statuszeile
  es – das ist der einzige Fall, in dem noch ein Level-Neustart nötig ist.
- **Formen:** Jede Form beschreibt nur Positionen (`x`, `y`, Ebene `z`). Höhere
  Ebenen liegen versetzt darüber, dadurch wirken die Steine dreidimensional.
  Neue Formen kommen als weiterer Eintrag in `js/board-layout-100.js` dazu:
  PC-Leicht als 12-x-8-Bild, Handy-Leicht als 6-x-8-Bild, Schwer als Umriss mit
  vier Höhenprofilen.
- **Steine:** 72 WebP-Dateien, verlustfrei aus PNG umgewandelt (2,72 → 2,06 MB,
  Pixel identisch). Im Startzoom sind sie je nach Form und Fenster rund
  38 × 45 px groß (gemessen bei 390 und 1280 px Breite), die 160 px Auflösung
  bleibt auch beim Hineinzoomen scharf.
- **Schrift:** **Atkinson Hyperlegible** (Braille Institute, SIL Open Font
  License) liegt lokal unter `css/fonts/` – 17 kB je Schnitt, eingebunden per
  `@font-face` und im `body`-Stapel vor Arial. Das Spiel braucht dafür kein Netz,
  auch die App nicht. Die Lizenz steht in `css/fonts/OFL.txt`.
- **Sprüche:** kommen aus `lob.txt`, `pech.txt` und `gewonnen.txt`, alle drei
  alphabetisch sortiert und zeilenweise in `"…"` gesetzt (Umlaute zählen als
  Grundbuchstabe, also Ä wie A). Die JS-Dateien enthalten dieselben Sprüche in
  derselben Reihenfolge und ziehen daraus ohne direkte Wiederholung.
- **Zwei Ablageorte:** Der Ordner liegt parallel unter
  `Documents\Homepage\spiele\mahjong` und `Documents\GitHub\Homepage\spiele\mahjong`.
  Änderungen gehören in beide, damit die Kopien gleich bleiben.
