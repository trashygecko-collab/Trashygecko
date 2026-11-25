# Space Invaders – Modernisierung (25.11.2026)

## Ausgangslage

`spiele/invaders/invaders.html`: 896 Zeilen. Ein Schiff, **ein** Schuss gleichzeitig,
11×5-Pixel-Invasoren in drei Arten mit zwei Frames, keine Level, keine Bosse, keine
Extras, kein Sound. Mädchen und Jungen waren nur Farbtoken über derselben Zeichenfunktion.

## Ergebnis

`spiele/invaders/invaders.html`: 1882 Zeilen, 80 KB, eine Datei, keine CDN.
Quellen liegen als drei Bausteine in `werkzeuge/` (`inv_head.html`, `inv_logik.js`,
`inv_stile.js`) und werden per Python zu einer HTML zusammengesetzt.

- **Drei Welten** (`STILE`): `klassisch` (alte Pixelgrafik, unverändert spielbar),
  `maedchen` (Wolkenwesen, Blütenelfen, Wolkenschutz, Schmetterlingsgleiter, Planeten,
  Galaxie), `jungs` (Metalljäger mit Cyan-Ring, rote Alienköpfe, Stahlbunker mit
  rotem Energiekern, Asteroiden, Sternennebel). Eigene Zeichenfunktionen je Welt,
  keine Farbfilter.
- **20 Level** als Formationsraster (`#`-Muster), vier Muster (`takt`, `pendel`,
  `seiten`, `sturz`), Angreifer-Schwärme, Späher (100–300 Punkte), Bosse bei
  Level 5/10/15/20 (Wächter 70, Zerstörer 120, Koloss 170, Herzogin 220 HP) mit
  drei Phasen und eigenen Salven.
- **Fünf Waffen**: Blaster, Doppel, Streu (3), Laser (Spalte), Schnellfeuer –
  jeweils mit Abklingzeit und Restzeit.
- **Fünf Kapseln**: Waffe, Schild, Nova, Punkte, Extraleben. Nova räumt das Feld,
  kostet Bossen 15 HP, max. 3 Ladungen.
- **Effekte**: typabhängige Explosionen, Partikel, Energieringe, Triebwerke,
  Bildschirmwackeln, Bosszerstörung; Parallaxhintergrund je Abschnitt.
- **HUD**: Startbildschirm, Punkte/Level/Leben/Waffe/Extras/Rekord, Bossleiste,
  Pause, Game Over, Levelabschluss, Sieg. Ton an/aus und Lautstärke (WebAudio,
  `invaders-laut`), Rekord weiter über `invaders-rekord`.
- Steuerung: Tasten (←/→, A/D, Leertaste, X, P) und Pad (◀ ▶ FEUER NOVA),
  DPR-Skalierung und `passeAn()` unverändert übernommen.

## Messungen (Prüflauf im Browser, 60-Hz-Ticks)

| Prüfung | Ergebnis |
| --- | --- |
| Start Level 1 | 27 Gegner, 204 Bunkerzellen, 3 Leben |
| Blaster auf Jäger | 1. Treffer 0 Punkte (2 HP), 2. Treffer +30, 26 Gegner |
| Doppelschuss | 2 Schüsse gleichzeitig |
| Streuschuss | 3 Schüsse |
| Laser | 27 → 24 Gegner (nur die Spalte), +60 Punkte |
| Schild gegen Bombe | Leben bleibt 3 |
| Bombe ohne Schild | Leben 3 → 2 |
| Nova | Feld geräumt (27 → 0), +540 Punkte, Ladung 2 → 1 |
| Kapseln | Leben 3 → 4, +250 Punkte, Waffe Laser 13 s |
| Boss Wächter | Phasen 1 → 2 → 3, 11 Bomben in 3,3 s |
| Boss besiegt | +800 Punkte, Level 6 mit 26 Gegnern |
| Level 20 | Herzogin 220 HP, Sieg „GALAXIE GERETTET" |
| Game Over | Leben 0, Meldung mit Level und Punkten |

## Behobene Fehler

1. `STILE` in TDZ: `neuesSpiel()` lief vor der Stil-Definition → Start brach ab.
2. `muster` im `sturz`-Zweig außerhalb seines Blocks (ReferenceError).
3. Levelabschluss prüfte `!spaeher` statt `!boss` – Bosslevel galten sofort als geschafft.
4. Laser tötete alle Gegner statt nur der getroffenen Spalte.
5. Bombentreffer leerte `bomben`, die Schleife lief weiter → Absturz auf `undefined`.
6. Tote Angreifer blieben im Array.
7. Reihenabstand 28 px ließ die drei Reihen überlappen (jetzt 33 px).

## Grenzen der Prüfung

- Getestet auf Desktop-Chromium (540 und 1100 px breit) mit Hook-Ticks und echten
  Klicks; Touch-Pad nur über DOM-Klicks, nicht auf einem echten Smartphone.
- Sound wurde nicht gehörmäßig geprüft (WebAudio startet erst nach Geste).
- Playwright-Snapshot für Invaders nicht aktualisiert.
- Die Pause bei Tabwechsel greift auch beim Screenshot; die Bilder zeigen daher
  teils die Pausemaske über dem Spielfeld.
