# Doodle Jump – Gameplay, Balance und Grafik (25.11.2026)

## Ausgangslage (640 Zeilen)

- **Monster-Kollision war ein Rechteck**: geprüft mit `dx < m.w/2 + r*0.7` und
  `dy < m.h/2 + r*0.7`, also 27,5 × 24,5 px um eine gezeichnete Ellipse von
  17 × 14 px. An den Ecken traf es bis zu 10 px neben der sichtbaren Figur.
- **Physik lief pro Frame, nicht pro Zeit**: `vy += 0,42` in jedem Aufruf. Auf
  einem 120-Hz-Handy war das Spiel doppelt so schnell, auf 30 fps halb so schnell.
- **Durchfallen an Plattformen**: Flog der Springer beim Aufsteigen in eine
  Plattform und drehte am Scheitel darin um, war `vorher` schon unter der Kante
  und die Landung griff nie. Gemessen: Läufe endeten nach 19–56 m.
- **Monster ab dem ersten Bild**: drei Stück, 34 × 28 px auf 70-px-Plattformen
  (fast die halbe Landefläche), seitliche Berührung sofort Tod.
- Mädchen und Jungen waren nur Farbtoken über derselben Zeichenfunktion.

## Änderungen

**Kollision und Physik**

- Alles rechnet in Frames zu 60/s: `f = dt*60` skaliert Schwerkraft, Sprung,
  Seitenbewegung und Kollisionsprüfung. 120 Hz und 60 Hz laufen gleich schnell.
- Landung: Oberkante gekreuzt (`vorher <= p.y+1,5 && unten >= p.y`) **oder**
  Scheitel steckt in der Plattform (`unten <= p.y+p.h+4`). Getestet mit
  Fallgeschwindigkeit 8, 12, 15 und 17 px/Frame: jedes Mal Absprung.
- Landetoleranz 14 px über die Plattformkante (vorher 11,2).
- Monster: Ellipse `rx = 0,45*m.w + 0,9*r` (23,4 px), `ry = 0,45*m.h + 0,9*r`
  (20,7 px) – knapp unter der sichtbaren Summe (26 × 23 px).
  Gemessen: Mitte, 18 px und 20 px seitlich treffen; **26 px seitlich und die
  Ecke (20/16 px) treffen nicht** (das alte Rechteck traf dort).
- Monster 34 × 28 → **24 × 18 px**, höchstens 3 (vorher 4), erst ab **150 m**
  (vorher drei ab dem Start).

**Balance**

- Sprung 11,6 → 10,8, Schwerkraft 0,42 → 0,40 → Sprunghöhe 146 px.
- Plattform 70 → **84 px** breit, Abstand 58–92 → 64–92, seitlicher Versatz je
  Abstand begrenzt: `34 + (140-dy)*1,05` ≈ 75–114 px. Gemessen über 6000 Ticks:
  größter Höhenabstand **91 px**, also 62 % der Sprunghöhe.
- Federn 10 → 13 % der grünen Plattformen; Schild und Goldplattform neu.
- Extras alle 0,075 Plattformen; Monsterwahrscheinlichkeit ab 150 m linear bis
  0,12.

**Neue Mechanik**

- **Goldplattform** (7 % ab 60 m): Absprung 15,6 statt 10,8, verliert je Sprung
  18 % Helligkeit.
- **Schild** (Kapsel, 9 s): fängt einen Monsterkontakt ab, Monster platzt,
  +25 Punkte, 1,4 s unverwundbar.
- Zerbrechende Plattformen wackeln 0,55 s und brechen dann (vorher sofort weg).

**Zwei Welten (eigene Zeichenfunktionen, keine Farbfilter)**

- **Mädchen – Zuckersprung**: hellblauer Himmel mit Verlauf, lachende Sonne mit
  zwölf drehenden Strahlen, ziehende Wolken, zwei Hügelreihen, zwei
  Schmetterlinge; Plattformen als flache Wolken (Oberkante genau auf der
  Landekante), Macarons, Kekse mit Punkten, Lollis; Figur ein Marshmallow mit
  großen Augen, Wangen, Blinzel- und Stauch-Animation; Monster ein
  Gummibärchen mit Hüpf-Animation.
- **Jungen – Nachtsprung**: Nachthimmel, Vollmond mit Kratern, dunkle Wolken,
  Bergsilhouette, Gebäude mit roten Lichtern, drei Fledermäuse, zwei
  Nebelbänder; Plattformen als Metallplatten mit Cyan-Kante, schwebende
  Stahlplatten mit Triebwerken, rissige Felsvorsprünge, Energieplatten; Figur
  ein Monster mit Hörnern, Glühaugen und Zähnen; Monster kleine Dämonen mit
  Hörnern und Glühaugen.
- Klassisch behält die alte Optik.

## Prüflauf (Browser, 1/60-Ticks)

| Prüfung | Ergebnis |
| --- | --- |
| Landung bei vy 8 / 12 / 15 / 17 | jedes Mal Absprung (vy −8,4 bis −9,2) |
| Monster Mitte / 18 / 20 px seitlich | Treffer |
| Monster 26 px seitlich / Ecke 20/16 / 30 px oben | kein Treffer |
| Monster ohne Schild seitlich | Spielende |
| Monster mit Schild seitlich | lebt, Schild 0, +25 Punkte, Monster weg |
| Sprung von oben aufs Monster | +75 Punkte, Monster weg, lebt |
| Monster in den ersten 300 Ticks (< 150 m) | 0 |
| Monster nach 900 Ticks bei 220 m | 3 (Deckel) |
| Bot-Läufe 6 × 300 s | 189, 197, 204, 207, 229 und 26 m (Mittel 175) |

Der Bot ist ein einfacher Autopilot ohne Vorausschau (er zielt auf die nächste
Plattform und weicht Monstern zur Seite aus); die 175 m sind eine Untergrenze,
kein Menschenmaß. Vor dem Umbau endeten vergleichbare Läufe nach 19–56 m.

## Grenzen

- Der 5-Minuten-Dauerlauf wurde mit dem Autopiloten simuliert, nicht von Hand
  gespielt; der Testbrowser hat den Tab im Hintergrund und führt kein
  `requestAnimationFrame` aus.
- Touch nicht auf einem echten Smartphone geprüft, nur Klick/Pointer im
  Desktop-Browser.
- Hoch-/Querformat: Das Feld skaliert über CSS (`max-width:400px`, `height:auto`),
  eine echte Drehung wurde nicht gemessen.
- Der Rekord im lokalen Speicher steht durch die Prüfläufe bei **490**; der
  vorherige Wert ließ sich nicht mehr feststellen.
