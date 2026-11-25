# Sokoban auf 100 Level – Bericht (25.11.2026)

## Was gebaut wurde

| Datei | Inhalt |
|---|---|
| `spiele/sokoban/level.js` | 100 Level als `window.SOKOBAN_LEVEL`, je mit `karte`, `loesung` (Zugfolge) und `pushes` (kleinste Schiebezahl des Solvers) |
| `spiele/sokoban/sokoban.html` | lädt `level.js`, Anzeige `Level 07 / 100`, Fortschritt, Haken in der Auswahl, Feldgröße nach Bildschirm |
| `werkzeuge/sokoban_level_bauen.py` | Generator: Rückwärtsziehen, Vorwärtsprüfung, Solverfilter |
| `werkzeuge/sokoban_solver.py` | unabhängiger A*-Solver mit Ecken- und 2×2-Erkennung |
| `werkzeuge/sokoban_nachbauen.py` | übernimmt gebaute Level aus dem Protokoll, ergänzt Solver-Lösungen |
| `werkzeuge/sokoban_spiel_patch.py` | geprüfte Ersetzungen am Spiel (läuft auf beiden Ablagen) |

## Levelbestand

- **100 Level.** 1–10 sind die alten Level des Spiels, unverändert – ihre Bestwerte
  (`sokoban-best-0` … `-9`) behalten damit ihre Bedeutung.
- 11–100 sind neu gebaut: Rückwärtsziehen vom gelösten Zustand, danach mit einem
  **unabhängigen Vorwärtssimulator** nachgespielt und mit dem **A\*-Solver** gelöst.
- Keine Varianten desselben Rätsels: jedes Level hat eine eigene Karte, Bau und
  Prüfung verwerfen doppelte Stellungen.

## Schwierigkeit (kleinste Schiebezahl des Solvers)

| Level | Kisten | Pushes min–max | Mittel |
|---|---|---|---|
| 1–15 | 1–3 | 2–11 | 4,6 |
| 16–30 | 2 | 6–9 | 7,7 |
| 31–50 | 3 | 12–23 | 15,7 |
| 51–70 | 4 | 18–22 | 19,1 |
| 71–85 | 4–5 | 16–22 | 18,6 |
| 86–100 | 5 | 12–22 | 16,1 |

**Ehrliche Grenze:** Ab Level 71 steigt die Zahl der Kisten (4 → 5) und die Enge
der Räume, die Schiebezahl steigt nicht weiter. Ein Versuch, die letzten 15 Level
mit 6 Kisten zu bauen, lief nach mehreren Minuten ohne Ergebnis – das Verfahren
wurde verworfen, statt die Reihe künstlich aufzublähen.

## Prüfungen

- `python sokoban_level_bauen.py --pruefen`: **100 Level, 0 mit Fehlern**
  (genau eine Spielfigur, Kisten = Ziele, Boden zusammenhängend, Lösung gültig).
- Jede Lösung wurde vorwärts nachgespielt: alle Kisten am Ende auf Zielen.
- Der Solver hat **alle 100 Level** ohne Kenntnis der Lösung gelöst; die
  `pushes` in `level.js` sind seine Ergebnisse.
- Kein Startzustand mit Kiste auf einem Ziel oder Kiste in einer toten Ecke.

## Spielfunktionen (Browser, Chromium)

- Level 1, 5, 12, 17, 36, 50 und 100 mit der hinterlegten Lösung durchgespielt:
  Meldung erscheint, Kisten `n/n`, Fortschritt zählt hoch.
- Levelwechsel über die Auswahl, „Weiter" nach dem Sieg (Level 12 → 13,
  Level 100 → 1), „Level neu" setzt Züge auf 0, Rückgängig 3 → 2 Züge.
- Fortschritt bleibt nach dem Neuladen: `sokoban-fertig` = `[11,12,16,35,…]`,
  Haken in der Auswahl, Bestwert bleibt.
- Feldgröße: 390×844 (hoch) Feld 348×290, Steuerkreuz passt; 844×390 (quer)
  Feld 276×230, nichts abgeschnitten. Canvas-Auflösung = Anzeigegröße (scharf).
- Keine Zeitbegrenzung, Tastatur (Pfeile, WASD, Z, R, Enter) und Wischen
  unverändert.
