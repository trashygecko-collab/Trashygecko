# Frogger: Zielplätze repariert, Levelsystem erweitert (25.11.2026)

## Der Fehler

Zwei Ursachen, beide in `spiele/frogger/frogger.html`:

1. **Buchten wurden auf der Zellgrenze gezeichnet.** Der Aufruf war
   `zeichneBucht(i, ((i*2)+1)*Z)`, die Funktion setzt den Mittelpunkt aber auf
   `x` (`x-wB/2`). Die Bucht lag damit 16 px – eine halbe Zelle – links neben
   der Spalte, die die Erkennung prüfte. Wer optisch in die Bucht sprang, stand
   in der Nachbarspalte; `sprungFertig()` fand dort keine Bucht und rief
   `sterben("Zwischen den Buchten gelandet")`. Drei solche Landungen = Game Over.
   **Jetzt:** `zeichneBucht(i, BUCHT_SPALTEN[i]*Z + Z/2)`, und die Erkennung
   liest dieselbe Tabelle (`BUCHT_SPALTEN = [1,3,5,7,9]`), statt die Spalte neu
   auszurechnen.
2. **Die Uhr lief vor dem Sprung.** In `update()` kam die Zeitprüfung vor dem
   Sprungabschluss. Lief die Zeit im letzten Sprungbild ab, starb der Frosch
   genau beim Landen in der Bucht. **Jetzt** wird erst der Sprung abgeschlossen,
   dann die Zeit geprüft.

Dazu: Landung in einer **schon belegten** Bucht meldet jetzt „Diese Bucht ist
schon besetzt" (vorher „Zwischen den Buchten gelandet") und kostet ein Leben,
aber keine Punkte.

## Levelsystem

- **12 Leveldefinitionen** (`LEVEL_DEFS`): je Level eigene Tempi, Richtungen,
  Breiten und Abstände für die fünf Fluss- und fünf Straßenbahnen. Ab Level 13
  wiederholen sich die Formen mit steigendem Tempo.
- **Tempo:** Straße +2,8 % je Level, Fluss +1,2 % je Level. Die Levelform bleibt
  erhalten, nur das Niveau wird auf eine monotone Kurve normalisiert.
- **Fortschritt:** `frogger-erreicht` im Browser; die Auswahl in der Leiste
  bietet nur freigeschaltete Level an. „Neues Spiel" beginnt bei Level 1,
  „Level neu" startet das laufende Level neu, die Auswahl springt jedes Level an.
- **Nach dem Sieg:** Meldung mit Punkten und Level, das nächste Level startet
  nach 4 s automatisch (Knopf „Weiter" oder Enter startet sofort).
- **Anzeige:** Statuszeile zeigt zusätzlich „Buchten n/5".

## Messungen

**Buchten einzeln (Level 1, je ein Sprung von Zeile 1 nach Zeile 0):**

| Bucht | Spalte | Leben vorher/nachher | belegt | Punkte | Status |
|---|---|---|---|---|---|
| 1 | 1 | 3 / 3 | 10000 | 120 | 1/5 |
| 2 | 3 | 3 / 3 | 11000 | 240 | 2/5 |
| 3 | 5 | 3 / 3 | 11100 | 360 | 3/5 |
| 4 | 7 | 3 / 3 | 11110 | 480 | 4/5 |
| 5 | 9 | 3 / 3 | 11111 | 800 (inkl. +200) | 5/5, Level geschafft |

**Fehlerfälle:** belegte Bucht erneut → Leben 3 → 2, Punkte unverändert 920,
Hinweis „Diese Bucht ist schon besetzt"; Spalte 2 (Zwischenraum) → Leben 2 → 1,
„Zwischen den Buchten gelandet"; Auto in Zeile 9 → „Autsch – überfahren",
Leben 3 → 2; Wasser ohne Stamm in Zeile 3 → „Platsch – der Frosch kann nicht
schwimmen", Leben 3 → 2.

**Tempo (Mittel der fünf Bahnen, Level 1–12):**

| Level | Straße | Fluss |
|---|---|---|
| 1 | 48,0 | 36,4 |
| 4 | 52,0 | 37,7 |
| 8 | 57,4 | 39,5 |
| 12 | 62,8 | 41,2 |

Beide Reihen steigen in jedem Schritt (geprüft: `strasseMonoton: true`,
`flussMonoton: true`). Schnellste Straßenbahn in Level 12: 81,9 px/s.

**Weiteres:** Level 1 komplett → `frogger-erreicht = 2`; nach Neuladen bietet die
Auswahl Level 01 und 02; „Weiter" setzt Buchten und Punkte zurück und startet
Level 2; „Level neu" setzt Züge/Buchten zurück; Leertaste startet, Pfeil hoch
springt eine Reihe (12 → 11), Pause/Weiter schaltet korrekt; keine
Konsolenfehler.

## Grenzen

- Die Prüfung der fünf Buchten, Kollisionen und Tempi lief über eine kurzzeitig
  eingebaute Prüfschnittstelle (`window.__fg`), weil `requestAnimationFrame` im
  Testbrowser nicht durchläuft; die Schnittstelle ist **entfernt**, die
  ausgelieferte Datei enthält kein `__fg`.
- Nicht von Hand durchgespielt: eine komplette Runde über alle zwölf Level.
- Kein neues Kunstwerk, keine neuen Abhängigkeiten; Optik, Steuerung, Punkte
  und Rekord unverändert.
