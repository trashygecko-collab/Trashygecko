# Snake als Einzelspieler, Würfelspiele als Mehrspieler

Stand 25.11.2026. Sicherung der Dateien vor der Änderung liegt in diesem Ordner:
`spieldaten.js`, `spiele.html`, `snake.html`, `connector.html`, `app.html`.

## Gefundene Fehler

1. **Snake-Einzelspieler war aus der Halle nicht erreichbar.** Die Spielhalle öffnet jede
   Spielseite mit `?room=…&name=…&id=…` (`spielAdresse()` im Connector). Snake schaltete
   bei vorhandenem Raum **immer** in den Zwei-Spieler-Modus und zeigte die Lobby – auch
   dann, wenn der Knopf „Spielen“ (Einzelspieler) gedrückt wurde. Ein zweiter Spieler war
   damit Pflicht.
2. **25 Würfelspiele standen unter „Einzelspieler“.** In `spieldaten.js` trugen sie das
   Merkmal `solo`, obwohl sie einen Tisch für 2–6 Personen und über den Connector
   Computergegner haben. Der Filter „Einzelspieler“ (Connector und App) zeigte sie
   deshalb als klassische Einzelspielerspiele; auf den Hallenkarten stand „Solo“.
3. **Zahlen liefen auseinander.** `spiele.html` nannte 94 Spiele, der Katalog hat 95;
   `spiele.html` lud `spieldaten.js` gar nicht, die Kopfzeile war fester Text.

## Änderungen

- `spiele/snake/snake.html`: Duell nur noch mit `&art=duell`; sonst Einzelspieler
  (Titel „SNAKE · EINZELSPIELER“). Bei Raum+Kennung zusätzlich der Knopf
  „Duell im Raum“, der die Adresse um `art=duell` ergänzt. Regeln um einen Satz ergänzt.
- `spiele/connector/connector.html` (+ Kopie `connector-online/` über `sync.py`):
  `spielAdresse(path, art)` und `openGame(path, art)`; die vier Snake-Aufrufe setzen
  `duell`. Kartenknopf heißt jetzt „Einzelspieler“ (nur solo), „Am Gerät spielen“
  (lokal/zwei) oder „Spielen“ (beides) statt überall „Spielen“.
- `spiele/spieldaten.js`: alle 25 Würfelspiele `lokal zwei computer` statt `solo …`.
- `spiele/spiele.html`: je Würfelkarte „Bots“ statt „Solo“; lädt `spieldaten.js` und
  setzt die Kopfzeile daraus; Werkzeugkachel und `android/start/app.html` auf 95.
- Regeln, Würfelmechanik, Punkte und Serverlogik wurden **nicht** angefasst.

## Messungen

| Prüfung | Ergebnis |
|---|---|
| Snake allein (headless Chromium) | Startbild „SNAKE · EINZELSPIELER“, Game Over nach 3,9 s, „Nochmal“ → 0 Punkte / Länge 3, zweiter Lauf läuft |
| Mitspielertext im Einzelspiel | keiner (HUD nur Punkte, Länge, Tempo, Rekord) |
| Apfel im Einzelspiel | 10 Punkte, Länge 3 → 4, Rekord 10 im `localStorage` |
| Snake-Adresse mit Raum, ohne `art` | Einzelspieler + Knopf „Duell im Raum“ |
| Snake-Duell (Raum TG-33D5, Bot „Mumpitz“) | Menü ausgeblendet, Stand „Namenloser 2 0 : Mumpitz 1“, Runde 1 beendet |
| Connector-Filter „Einzelspieler“ | 52 → **36** Spiele, **0** Würfelspiele |
| Connector-Filter „Mehrspieler“ | **72** Spiele, alle **25** Würfelspiele |
| Würfelspiel zu zweit am Gerät (Beelzebub) | 5 Runden, Spieler 1 **40** : Spieler 2 **20**, Siegmeldung |
| Würfelspiel gegen Computer (Beelzebub, Bot „Vroni“) | Bot würfelt und punktet (27) |
| `go test ./...` (spiele/connector) | grün, 228,8 s |
| `python spiele/pruefe-spieldaten.py` | keine Abweichung (vorher 3) |
| Handy 390 × 844 | Halle und Snake ohne waagerechten Überlauf, Titel und Startknopf lesbar |
| Andere Spiele (Sudoku, Tetris, Schach) | laden ohne 4xx/5xx und ohne Seitenfehler |

## Offen

- Das verlinkte APK ist der alte Bau. `assets/www` und `assets/katalog.json` entstehen
  beim Bauen; sie wurden nicht von Hand geändert, die App zeigt die neue Zuordnung erst
  nach einem Neubau.
- Der Testlauf im Browserfenster ist kein echtes Touch-Gerät; Wischen wurde über die
  Pointer-Ereignisse des Spiels geprüft, nicht mit einem Finger.
