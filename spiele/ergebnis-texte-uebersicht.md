# Gewinn-, Lob- und Pechtexte aus dem Connector – Übersicht

Stand: 25. November 2026

## Wie es funktioniert

Alle Texte kommen aus den zentralen Sammlungen, die auch der Connector/Mahjong nutzt.
In den Spielen selbst steht **kein** Spruch und **keine** eigene Liste – nur ein Aufruf.

| Datei | Inhalt |
| --- | --- |
| `spiele/lob.js` | `LOB_SPRUECHE` (Lob, 2420 wirksame Sprüche) + `lobSpruch(mitSteinen)` |
| `spiele/gewonnen.js` | `GEWONNEN_SPRUECHE` (624; 552 ohne Mahjong) + `gewonnenSpruch(ohneMahjong)` |
| `spiele/pech.js` | `PECH_SPRUECHE` (1379; 1330 ohne Mahjong) + `pechSpruch(ohneMahjong)` |
| `spiele/ergebnis.js` | gemeinsamer Anzeiger + API |

Die drei Listen werden erst beim ersten Aufruf geladen (die Spiele bleiben also schnell)
und benutzen die vorhandenen Mischbeutel: kein Spruch kommt direkt hintereinander zweimal.

### Mahjong-Filter

`gewonnen.js` und `pech.js` enthalten auch reine Mahjong-Witze (Steine, Paare,
Zwillinge, „Mahjong“ selbst). Für allgemeine Spiele werden sie herausgefiltert:
`gewonnenSpruch(true)` und `pechSpruch(true)` ziehen aus 552 bzw. 1330 Sprüchen.
Ohne Argument bleibt die volle Liste; Mahjong ruft sie so auf und behält alle
624 Gewinn- und 1379 Pechsprüche. Allgemeine Stein-Redewendungen („in Stein
gemeißelt“, „Meilenstein“, „stein-gut“) bleiben erhalten. `lobSpruch(false)`
war schon vorher getrennt.

### Schnittstelle

```js
TGErgebnis.gewonnen();   // Sieg: Gewinnspruch
TGErgebnis.verloren();   // Niederlage: Pechspruch
TGErgebnis.lob();        // neutrales Lob (z. B. Gleichstand)
TGErgebnis.nachName("Bert", "Anna"); // Sieger + eigener Name → Gewinn oder Pech
```

Der Anzeiger zieht automatisch die gefilterten Fassungen; die Spiele müssen
nichts weiter tun. Wer die Liste direkt nutzt (tictactoe, paare, 1x1, tetris),
ruft ebenfalls `lobSpruch(false)` bzw. `pechSpruch(true)`.

Einbindung in einer Spielseite (Pfad je nach Tiefe):

```html
<script src="../../ergebnis.js"></script>
```

### Anzeige

Eine eigene Karte `#tg-spruch` unten mittig über dem Spielfeld:

* `position: fixed`, `pointer-events: none` – verdeckt keine Bedienelemente und schluckt keine Klicks
* Farbe über CSS-Variablen `--tg-spruch-*`, passt sich dem dunklen Grundlook an
* blendet nach 3,5–7 s von selbst aus, 1,2-s-Sperre gegen Doppelmeldungen
* auf Handy (390 × 844) und Desktop geprüft: im Bild, kein Querscrollen

## Integriert

### Würfelspiele (vollständig)

Alle 13 Spiele über den gemeinsamen Rahmen (`wuerfelspiele/rahmen.js`, online über
`wuerfelspiele/online.js`) – ein Hook deckt lokale Partie und Online-Partie ab:

der-turm, eins-bis-sechzehn, fenster,
gestrichene-zwoelf, hin-und-her, jule, paerchen, plus-minus, rentmeister,
streichhoelzer, teure-sieben, unter-oder-ueber, wuerfelpoker (plus `index.html`).

Weitere Würfelspiele mit eigenem Ende:

| Spiel | Ort des Hooks | Ergebnis |
| --- | --- | --- |
| zehntausend | `beenden()` / `zeigeEndeOnline(g)` | lokal Sieger, online eigener Stand |
| maexchen | `zeigeEnde()` | Sieger / eigener Platz |
| chicago | `zeigeEnde()` | Verlierer wird erkannt |
| shipcrew | `zeigeEnde()` | Siegerpunkte |
| kniffel | `endGame()` | online eigener Platz, lokal Sieger |
| pig | `kopfZeichnen()` bei `status === "ende"` | online eigener Platz, lokal Sieger |

### Kartenspiele

| Spiel | Hook | Ergebnis |
| --- | --- | --- |
| uno | `s.status === "finished"` | eigener Platz gegen `s.winner` |
| baccara | `zeichneWertung()` bei `status === "ende"` | eigenes Netto |
| siebzehn | `zeichneWertung()` bei `status === "ende"` | eigenes Netto |
| cabo | `zeichneWertung()` bei `status === "ende"` | Mensch gegen Bot / Mehrspieler |
| arschloch | `endRound()` | Runde vorbei |
| maumau | Spielende | Siegername |
| sechsundsechzig | `status === "ende"` | Siegername |
| poker | `game.status === "ende"` | eigener Chipstand gegen den Tisch |
| schwimmen | `status === "finished"` | eigener Platz gegen `winner`; Zuschauer → Lob |
| skyjo | `game.status === "finished"` | Gewinner „red"/„blue"/„draw" über `redId`/`blueId`; Zuschauer → Lob |

### Brett- und weitere Duelle

| Spiel | Hook |
| --- | --- |
| dame | `gameOver` + `showWinner` |
| muehle | `finishGame(winner)` |
| schach | Schachmatt-Meldung |
| viergewinnt | Vierer-Reihe |
| backgammon | `finishGame()` |
| go | zwei Pässe → `finishGame()` |
| reversi | Spielende (Schwarz/Weiß) |
| tictactoe | Sieg (nutzt `lobSpruch(false)`/`pechSpruch(true)`, zentrale `pech.js`) |
| sim | Dreieck geschlossen |
| hold-the-line | kein Zug mehr |
| kaesekaestche | `finishGame()` |
| menschaerger | alle Figuren im Ziel |
| gaensespiel | Feld 63 |
| schlangen | Feld 100 |
| bingo | `lokal.gewonnen` |
| kalaha | „Spiel beendet" |
| memory | `gewonnen()` (lokale Partie) |
| paare | Duell-Ergebnis (rot/blau, Gleichstand) |
| mastermind | `gewonnen()` / `verloren()` |
| wortsuche | `status === "ende"` |
| galgen | gelöst → Gewinn, Fehlerlimit → Pech |
| schiffe | `gewonnen` der Partie |
| streichholzspiel | letztes Streichholz |
| verruecktes-labyrinth | `gewonnen(f)` |
| mahjong | vorhandene Anzeige, nutzt jetzt `spiele/gewonnen.js` + `spiele/pech.js` |
| domino | Rundenende lokal und online | eigener Sitz gegen `gewinner`; Zuschauer → Lob |
| halma | lokaler Sieg und `status === "finished"` | eigener Platz gegen `sieger`; Zuschauer → Lob |
| risiko | `zeigeSieg()` und fertiger Serverstand | eigener Platz aus `risiko-fern.js`, im Team das eigene Team; Zuschauer → Lob |

Die fünf Online-Zweige (schwimmen, skyjo, domino, halma, risiko) melden ohne
eigenen Sitz Lob statt fälschlich „gewonnen“ oder „verloren“. Bei Domino wird
die Rundenmarke zurückgesetzt, sobald der Serverstand nicht mehr `ende` ist –
sonst bliebe der erste Sieg einer neuen Runde mit derselben Rundennummer stumm.

## Offen

Keine: Die zuletzt fehlenden sechs Spiele (poker, schwimmen, skyjo, domino, halma,
risiko) sind verdrahtet. Domino meldet das Rundenende (das Spiel kennt kein eigenes
Partieziel), alle anderen das Partieende.

## Bewusst nicht ausgestattet

* **solitaer** (Karten), **labyrinth** (Kugel zum Pokal), **1x1**, **malen** und ähnliche
  Einzelspieler-/Werkzeugseiten: es gibt keinen Gegner und damit kein Gewinnen/Verlieren
* **Arcade-Titel** (Tetris, Snake, Pong, Tron, Bomberman, Bubbleshooter, Quizshow …):
  Punkte-/Levelspiele ohne Duell-Ergebnis; Bomberman/Quizshow wurden nicht geprüft

## Prüfungen

* `node --check` über alle geänderten Seiten und die zentralen Dateien: keine Syntaxfehler
* Alle 32 verdrahteten Brett-/Kartenseiten laden im Browser ohne JS-Fehler und ohne
  fehlende Dateien; `TGErgebnis` ist überall vorhanden
* zehntausend, kniffel, tictactoe: Spielende über die echte Spiel-Logik ausgelöst → Spruch
  aus der jeweiligen Zentralliste; tictactoe zusätzlich auf 390 × 844
* poker, schwimmen, skyjo, domino, halma und risiko: Spielende über die echte Logik
  bzw. einen fertigen Serverstand ausgelöst → je genau ein Spruch aus der passenden
  Zentralliste, Karte im Bild; alle sechs Seiten laden ohne JS-Fehler
* Filter: 552 allgemeine Gewinn- und 1330 Pechsprüche gezogen, 0 Mahjong-Reste;
  Mahjong selbst weiter 72 Mahjong-Gewinnsprüche in 624 Zügen; `lobSpruch(false)`
  ohne `LOB_MAHJONG`
* Zuschauer: risiko, halma, domino je Lob; ohne Sitz kein „gewonnen“/„verloren“;
  Risiko-Niederlage am Platz 1 → Pech, Sieg an Platz 0 → Gewinn
* Domino: Neustart mit derselben Rundennummer 1 meldet den Sieg erneut
  (vorher blieb er stumm)
* Nicht durchgespielt: die übrigen Brett-/Kartenspiele (nur Lade- und Syntaxprüfung)

## Sicherung

Vor jeder Änderung wurde die Datei kopiert nach
`spiele/sicherung/2026-11-25-ergebnistexte/`; die Filter- und Platzänderungen
liegen in `spiele/sicherung/2026-11-25-ergebnisfilter/`.
