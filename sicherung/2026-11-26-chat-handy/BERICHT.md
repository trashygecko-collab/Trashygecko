# Chat und Smiley-Auswahl am Handy

Stand 26.11.2026. Gesichert vor der Änderung: `connector.html` (dieselbe Datei
liegt byte-gleich in `connector/` und `connector-online/`).

## Gefundene Fehler

1. **Chat klebte am oberen Rand.** Im Handy-Layout stand `#chatbox` per
   `order:1` direkt unter dem Kopf – also über Spielern, Einstellungen und
   allem anderen. Der Chat begann bei 46 px Seitenhöhe.
2. **Smiley-Fenster ragte aus dem Bild.** `.smiley-panel` stand fest
   `position:absolute; bottom:calc(100% + 6px)` über der Eingabezeile, mit
   `max-height:min(50vh,340px)`. Weil die Eingabezeile ganz oben lag, öffnete
   das Fenster nach oben aus dem Bildschirm hinaus: hochkant `top = −2 px`,
   quer `top = −24 px` – die obersten Kacheln waren abgeschnitten und nicht
   antippbar. Die Höhe kannte den tatsächlich freien Platz nicht.

## Änderungen (`spiele/connector/connector.html`, Kopie `connector-online/`)

- **Reihenfolge im Handy-Block (≤1023 px):** `#zentrale` (Spieler) 1,
  `#einstellungen` 2, `#chatbox` 3. Der Chat steht jetzt unter den Spielern
  und Einstellungen statt direkt unter dem Kopf.
- **Smiley-Fenster wird gerechnet, nicht geraten.** Neu
  `smileyPanelPlatzieren()`: misst die Eingabezeile und den sichtbaren
  Bereich (`visualViewport`, sonst `innerHeight`), öffnet über der Eingabe,
  wenn dort mindestens 240 px sind, sonst darunter, und wenn beide Seiten eng
  sind (quer gehalten, Tastatur offen), über die ganze Höhe. Höhe und Lage
  werden auf 8 px Rand begrenzt; `position:fixed` in diesem Bereich, damit
  kein übergeordneter Kasten mitschneidet.
- Das Fenster bleibt am Platz bei `resize`, `orientationchange`, Scrollen und
  Änderungen am `visualViewport` (Adressleiste, Tastatur).
- Beim Schließen werden die gerechneten Werte entfernt (`style` weg), die
  Desktop-Regel greift wieder.
- Solange das Fenster offen ist, ist der schwebende Chat-Knopf
  (`#tg-mobil-knopf`) ausgeblendet (`body.smiley-offen`), damit er keine
  Kachel verdeckt.
- Am PC ändert sich nichts: dort bleibt es bei `position:absolute` über der
  Eingabezeile, `smileyPanelPlatzieren()` steigt bei ≥1024 px sofort aus.

## Messungen (Chromium, Geräte-Emulation, echter Connector auf 8767)

| Probe | vorher | nachher |
|---|---|---|
| Chat im Seitenfluss, 390 × 844 | y = 46 px (unter dem Kopf) | **y = 320 px** (unter Spielern + Einstellungen) |
| Smiley-Fenster hochkant | `top −2`, Höhe 340 | **`top 8 … 306`**, Höhe 298, passt |
| Smiley-Fenster quer 844 × 390 | `top −24 … 171`, 24 px abgeschnitten | **`8 … 348`**, Höhe 340, passt |
| Fenster bei offener Tastatur (390 × 560) | – | **`227 … 552`**, Höhe 325, passt |
| Fenster mit offener Partie (Chat 287 px) | – | **`274 … 614`**, passt, Knopf `none` |
| Kacheln | 340 im Bestand, oberste unerreichbar | 24–75 gleichzeitig erreichbar, Raster scrollt (3414 px Inhalt) |

Weitere Proben: oberste und unterste Kachel angetippt → Auswahl übernommen,
Nachricht mit Bild gesendet (`img.chat-bild`, `naturalWidth > 0`), Escape
schließt, Klick daneben schließt, Drehen bei offenem Fenster rechnet neu
(hochkant 298 → quer 340 → hochkant 298), Partie öffnen/schließen unverändert,
keine Seitenfehler.

**PC 1280 × 800 unverändert:** Chat 147 … 800 (653 px), Fenster `absolute`
381 … 721 (340 px) – dieselben Werte wie vor der Änderung.

## Grenzen

- Getestet in der Geräte-Emulation (390 × 844, 844 × 390, 390 × 560), nicht auf
  einem echten Telefon; die eingeblendete Tastatur wurde über eine kleinere
  Fensterhöhe nachgestellt, nicht über eine echte Tastatur.
- Nur die Oberfläche geändert, keine Chat-, Smiley- oder Serverlogik.
