# Arkanoid: zwei Designwelten, elf Extras (25.11.2026)

## Bestandsaufnahme vor dem Umbau

`spiele/arkanoid/arkanoid.html` (1106 Zeilen) hatte drei Ansichten
(`klassisch`, `maedchen`, `jungs`) mit je einer Zeichenfunktion für Hintergrund,
Stein, Kapsel, Laser, Schläger, Kugel und Rahmen. Vorhandene Extras: **E**
(breiter Schläger), **C** (Ball klebt), **L** (Laser), **D** (drei Bälle),
**S** (langsamer), **P** (Extraleben). Alle fielen als dasselbe 18×22-Rechteck
mit einem Buchstaben. Sound gab es nicht.

**Fehler im Bestand:**

1. **S wirkte dauerhaft.** `powerGeben("S")` rechnete `vx*=0.65, vy*=0.65` einmal
   auf die Kugel; der Ablauf nach 18 s nahm nichts zurück. Die Anzeige lief ab,
   der Ball blieb langsam.
2. **C klebte für immer.** Beim Ablauf wurde nur `fangBall=null` gesetzt, aber
   `b.fest` blieb `true` – der Ball hing am Schläger, bis ein Leben verloren ging.
3. **Globale Canvas-Zustände.** `powerSymbol` setzte `lineJoin`/`lineCap` ohne
   `save()/restore()`; danach bekamen alle Steine runde Ecken.

## Was dazugekommen ist

**Fünf neue Extras** (fehlten gegenüber den klassischen Arkanoid-Power-ups):
**K** schmaler Schläger (12 s), **F** schneller (14 s), **T** durchdringt Steine
(10 s), **G** Schild am unteren Rand (16 s), **B** Bonus (sofort, 250 + 50·Level).

**Jedes Extra hat eine eigene Silhouette und ein eigenes Symbol** (Linien, keine
Schrift): E doppelter Pfeil nach außen, K nach innen, D drei Kugeln, L zwei
Strahlen, C Netz, T Kugel mit Spur, G Schild, P Herz, S Sanduhr, F Blitz,
B Stern. Jungs: dunkle Hexkapsel mit Neonrand und Puls. Mädchen: Perlmuttkreis
mit farbigem Kern, Goldrand und dunklem Symbol. Klassisch: farbige Kapsel mit
dunklem Symbol.

**Jungen-Version:** Chromschläger mit Verlauf, Chromkante, atmender Cyan-Leiste,
Endkappen und Lichtschweif; Kugel als Chromkugel mit Plasmakern, dessen Glow mit
dem Tempo wächst; Hintergrund mit Raster, Platinenbahnen, Vignette und pulsierendem
Cyan-Saum; Steine als abgeschrägte Metallplatten mit Neonkante, Riffelung und
Riss nach dem ersten Treffer.

**Mädchen-Version:** Perlmutt-Schläger mit Regenbogenverlauf, Goldkante und
wanderndem Schimmer; Kugel als facettierte Kristallkugel mit pinkem Glow;
Hintergrund aus Creme, Rosen und Lavendel mit weichen Schleiern und aufsteigendem
Glitzer; Bonbonsteine mit Goldkante, Perlmuttglanz und Herz bei Doppelsteinen.

**Gemeinsam:** Partikel (max. 140) und kurze Lichtringe bei Treffer, Zerstörung
und Einsammeln; Steinblitz; Hintergrund einmal pro Stil in ein eigenes Canvas
gecacht, pro Bild nur eine kleine lebende Schicht; kleiner WebAudio-Ton
(Schläger, Stein, Zerstörung, Extra, Schild, Verlust) mit Knopf „Ton an/aus" und
gespeicherter Wahl.

**Balance:** Breite und Tempo werden aus allen laufenden Extras berechnet
(E=126, K=58, beide=84; S=×0,72, F=×1,28), damit sich Extras nicht überschreiben.
Gleiche Extras verlängern ihre Restzeit. Anzeige unten: `E 18s · S 14s · G 16s`.

## Messungen

**Extras einzeln** (über die Prüfschnittstelle, danach wieder entfernt):

| Extra | Messung |
|---|---|
| E | Breite 84 → 126, Anzeige „E 18s", nach Ablauf 84 |
| K | 84 → 58, nach 13 s wieder 84 |
| E + K | 84 (neutral) |
| S | 300 → 216 px/s (×0,72) |
| F | 300 → 384 px/s (×1,28) |
| S + F | 300 → 276 px/s (×0,9216) |
| C | Ball klebt; nach 17,95 s gelöst, v = 306 px/s |
| L | 2 Schüsse, Laser zerstört 3 Steine |
| D | 1 → 3 Bälle |
| T | Kugel prallt nicht ab (vy bleibt −200), 2 Steine in 0,16 s |
| G | 49 Bilder mit vy < 0, Leben bleibt 3 |
| P | 3 → 5 Leben (Obergrenze) |
| B | +300 Punkte (Level 1) |
| Kapsel | fällt, wird eingesammelt, Schläger 126, Kapsel verschwindet |
| Mehrere | E + S + G gleichzeitig, Anzeige „E 18s · S 14s · G 16s" |
| Level | alle Steine weg → „Level 1 geschafft", Weiter → Level 2 mit 50 Steinen |

**Kontrast (mittlere Helligkeit 0–255):** Jungs Feld 8 / Stein 72 (vorher 54,
aufgehellt), Klassisch 25 / 183, Mädchen Feld 232 / Stein 221 – dort trägt der
dunkle Rand und die Sättigung den Kontrast, nicht die Helligkeit.

**Mobil 390×844:** Canvas 361×391 px, Dokumentbreite 390 = Fensterbreite, kein
seitliches Scrollen; Kapseln und Steine bleiben unterscheidbar.

**Playwright:** `tests/spiele.spec.js -g arkanoid` in 1024×600, 800×480 und
1280×720 – 3 Tests bestanden, Bildreferenzen erneuert, keine Konsolen- oder
Ladefehler.

## Grenzen

- Die Extras wurden über eine kurzzeitig eingebaute Prüfschnittstelle
  (`window.__ak`) ausgelöst; sie ist entfernt, die Datei enthält kein `__ak`.
- Nicht von Hand durchgespielt: eine komplette Runde über mehrere Level mit
  natürlichem Power-up-Fall (16 % je zerstörtem Stein).
- Der Ton wurde nicht abgehört (kein Mikrofonzugriff im Test); geprüft ist nur,
  dass der Knopf schaltet und die Tonaufrufe ohne Fehler laufen.
