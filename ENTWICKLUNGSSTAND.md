# Entwicklungsstand – TrashyGecko Android

Lebendes Dokument. Nach **jeder** abgeschlossenen Arbeit fortschreiben (oben das Neueste).
Was getestet wurde, gehört zusätzlich in [`PRUEFBERICHT-ANDROID.md`](PRUEFBERICHT-ANDROID.md).

---

## 27.11.2026 · Pac-Man: großes PC-Feld, Daumensteuerung, ruhigeres Tempo

**Auftrag:** Pac-Man gezielt überarbeiten: großes PC-Spielfeld, feste
Daumensteuerung rechts unten am Handy, spürbar langsameres Spieltempo,
dunklerer Mädchenstil, Start/Pause/Neustart/Ende im Spiel und vollständige
Schnittstelle 2.0. Normal- und Bubenstil sowie Labyrinth, Punkte, Leben,
Geisterregeln und Kraftpillen bleiben erhalten. Keine anderen Spiele ändern.

**Umgesetzt**

- **PC-Layout:** `main` ist jetzt bis 1440 px breit; das Spielfeld sitzt in
  einer großen Bühne, Status und Hinweis stehen darüber, die Steuerung und
  der Tastaturhinweis in einer schmalen rechten Leiste. Das Seitenverhältnis
  bleibt exakt 19:21 und wird aus dem freien Platz berechnet. Gemessen:
  **1280×800 → Feld 586×648**, Bühne 964×666; **1920×1080 → Feld 839×928**,
  Bühne 1121×946; kein Abschneiden, kein Querscrollen.
- **Startbildschirm:** eigener Bereich mit Titel, Figurenvorschau, kurzer
  Erklärung, Grafikstilwahl (Normal/Mädchen/Buben), großem Startknopf und
  aufklappbaren Regeln. Der Spielbereich ist bis zum Start verborgen. Die
  Wahl wird unter `pacman-stil` gemerkt; beim direkten Start ist sie sichtbar,
  bei einem von außen übergebenen Stil nicht.
- **Handysteuerung:** Das Steuerkreuz sitzt fest rechts unten
  (`position:fixed`, Safe-Area-Abstand), nicht mehr mittig über dem Feld.
  Gemessen bei 390×844: Feld 360×397, Pad **175,8×108,4 px** bei
  x 204,2/y 725,6 (rechts 380, unten 834), kein Überlappen. Bei 320×568:
  Feld 243×269, Pad 156×99; bei 260×563 (150-%-Zoom-Ersatz): Feld 230×254,
  Pad 156×99, kein Überlappen. Querformat 844×390: Feld 256×284, Pad 152×92,
  Tasten 48×44 px, ebenfalls ohne Überlappen. Im Querformat sitzt das Pad
  rechts neben der Bühne.
- **Tempo tatsächlich langsamer:** Pac-Man **6,1 → 3,6 Kacheln/s** (Level-
  Zuwachs 1,2 → 0,75, pro Level 0,2 → 0,12). Geister **5,2/5,02/4,84/4,66 →
  3,15/3,00/2,85/2,70 Kacheln/s** (Level-Zuwachs 1,4 → 0,6, pro Level 0,18 →
  0,10), Heimweg **9 → 4,8**; Angstfaktor bleibt 0,62. Geändert ist die in
  `bewegePac`/`bewegeGeist` benutzte Geschwindigkeit, nicht nur die Animation.
  Gemessen: Pac-Man 72 px in 1 s = **3,6 Kacheln/s**; Geister im geraden Lauf
  rund 3,0–3,3 Kacheln/s (Rasterabtastung).
- **Grafik:** Normal und Buben bleiben in ihrer vorhandenen Zeichenwelt.
  Der Mädchenstil wurde von Leuchtrosa auf **dunkles Rosenholz** umgestellt:
  Seitenfläche `#9a446a → #6d2948 → #3b1228`, Bühne `#240b18`, Wände
  `#c07a9c/#8f4a6c/#6d3250/#4a2038`, Punkte warmgold, Figuren hell und klar
  erkennbar. Der Start-Pac-Man ist jetzt ein echter Kreis mit Mund und Auge
  statt eines Dreiecks. Keine neuen Dekorationen im Spielfeld.
- **Schnittstelle 2.0:** `window.PacMan` mit `version 2.0.0`, `bereit`,
  `whenReady`, `start`, `einstellungen`, `zustand`, `startbildschirm`, `stop`,
  `pause`, `resume` und `aktion`. Erlaubtes Startfeld nur `stil`
  (`normal`/`maedchen`/`buben`); ein reiner `start({stil})` richtet nur die
  Welt ein. Aktionen `richtung`, `start`, `pause`, `weiter`, `neu`; Ereignisse
  `pacman:bereit|ende|pausiert|weiter`. Pause/Weiter sind idempotent,
  `zustand()`/`einstellungen()` liefern Kopien, Fehlercodes nach Schnittstelle.
- **Connector:** `PacMan` steht jetzt in `SPIEL_API_NAMEN` beider
  `connector.html`; der Connector ruft `start({stil})` auf. Die eigene
  Stilwahl bleibt dort verborgen; ein Connector-Stil hat Vorrang, auch vor
  einer im Spiel gespeicherten Wahl. Keine CDN- oder Netzabhängigkeit.

**Geprüft**

- **PC 1280×800 und 1920×1080**, Handy **390×844, 320×568, 260×563** und
  Querformat **844×390**, jeweils alle drei Stile: Feld vollständig sichtbar,
  Seitenverhältnis erhalten, kein Querscrollen, Pad innerhalb des Viewports
  und ohne Überlappung mit dem Feld. Startbildschirm bei 390 und 320 ohne
  Scrollen; Stilwahl bleibt nach Neuladen erhalten (Mädchen getestet).
- **Tastatur:** Pfeil links bewegt Pac-Man (190 → 145,6 px in 600 ms),
  Pfeil hoch wird angenommen; Pause/Weiter-Knopf wechselt `laeuft → pause →
  laeuft`; „Neu“ startet sofort eine neue Runde, „Start“ öffnet den
  Startbildschirm. Keine Konsolen- oder Seitenfehler.
- **Touch:** Tipp auf das rechte Pad-Feld bewegt Pac-Man (390: 190 → 227,2 px
  in 500 ms; 320 und Querformat ebenso); Pad ist sichtbar und erreichbar.
- **Schnittstelle:** Version/`whenReady`; `start({stil})` ohne Start;
  unbekannter Stil → `normal`; fremdes Feld, falscher Typ und laufende Partie
  liefern die vorgesehenen Fehler; `pause`/`resume` zweimal hintereinander
  antworten mit demselben Pausenzustand; `aktion` prüft Richtung und Felder;
  Kopien bleiben unabhängig; Ereignisse `bereit`, `pausiert`, `weiter` und
  `ende` wurden beobachtet. Ein Testlauf bis **GAME OVER** (3 Leben, 130
  Punkte, „Nochmal“) und drei kurze Neustart-Runden (je 60 Punkte) liefen ohne
  Fehler; Kraftpille an der Ecke: 130 Punkte, 3 Leben.
- **Connector:** Pac-Man im Rahmen geöffnet, `stilAnSpiel()` meldet Erfolg;
  Mädchen aus dem Connector kommt an (`stil:maedchen`, `stilWahl:false`,
  Kachel verborgen), Startknopf startet die Partie; Connector `klassisch`
  überschreibt eine gespeicherte Mädchenwahl auf `normal`. Die statische
  Connector-Prüfseite meldete nur die erwarteten 501/404 des nicht laufenden
  Go-Backends, keine Pac-Man-Skriptfehler.

**Offen / Grenzen**

- Kein vollständiger Levelabschluss und keine Frucht wurde separat
  durchgespielt; Levelstruktur, Punkte, Leben, Kollisionen, Geisterregeln und
  Kraftpillen wurden nicht umgebaut. Ein echter Browser-Zoom wurde nicht
  separat gemessen; 260×563 diente als 150-%-Ersatz für 390×844.
- **Nicht gepusht, nicht deployed.** Nur der Commit folgt nach den Tests.

**Sicherung:** `arbeit/sicherung/pacman-2026-11-27/` (GitHub- und
Homepage-Kopie vor der Änderung). Bilder: `arbeit/screens/pacman-*.png`.

**Dateien:** `spiele/pacman/pacman.html` in beiden Ablagen;
`spiele/connector/connector.html` und `spiele/connector-online/connector.html`
(nur `PacMan` in der API-Liste); `ENTWICKLUNGSSTAND.md`.

---


## 27.11.2026 · Mensch ärgere dich nicht: Eckmedaillons an die Figuren gesetzt

**Auftrag:** Die Spielfiguren wieder in ihre ursprünglichen Parkplätze
setzen und die Hintergrundgrafiken der vier Ecken daran ausrichten – nicht
umgekehrt. Nur diese Korrektur; Mädchenstil, Blinken, Würfelposition,
Spielmechanik, Laufwege, Bots und Schnittstelle 2.0 bleiben unverändert.

**Ursache und Umgesetzt**

- Die letzte Änderung hatte `PARKPLATZ` von `1,25/2,75` auf `0,75/2,25`
  Zellen verschoben. Dadurch standen die Figuren links oben im 3×3-Feld,
  passend zu den bereits links oben stehenden Medaillons. Das ist
  rückgängig gemacht.
- **Figuren:** `PARKPLATZ` liegt wieder exakt bei `1,25` und `2,75` Zellen,
  also mit der Mitte bei `2,0` Zellen. Die vier Parkplätze sind damit wieder
  die ursprünglichen; die Startfelder und Laufwege sind unverändert.
- **Eckgrafiken:** Die zentralen Medaillons/Symbole sitzen jetzt bei
  `2,0` Zellen, nicht mehr bei `1,5`. Das betrifft den Stern (Normal), das
  Rosenmedaillon samt Symbol (Mädchen) und das Symbol (Buben) in allen vier
  Ecken.
- **Sichtbare Bildmitte:** Neu ist `zentriereEckSymbol(basis)`: Nach dem
  Einsetzen wird die sichtbare SVG-Zeichnung über ihre `getBBox()`-Mitte
  ausgerichtet. Dadurch verschieben leere Ränder bzw. asymmetrische Formen
  wie Krone und Kleeblatt die optische Mitte nicht.
- **Unverändert:** Mädchenstil, Normal-/Bubenstil, Blinkanimation,
  Würfelposition in der Brettmitte, Regeln, Laufwege, Bots, Mehrspieler und
  Schnittstelle 2.0.

**Geprüft**

- **PC 1280×800 und Handy 390×844**, jeweils **Normal, Mädchen und Buben**,
  jeweils alle vier Ecken (`yellow`, `green`, `blue`, `red`).
- **Figuren:** 12 Figuren in den Parkplätzen exakt auf den vorgesehenen
  Mittelpunkten, maximale Abweichung **0,00** SVG-Einheiten; die übrigen
  4 Figuren stehen erwartungsgemäß auf den Startfeldern und wurden nicht
  verändert. Geprüfte Parkplatzmitten:
  `yellow 145,45/145,45`, `green 654,55/145,45`,
  `blue 145,45/654,55`, `red 654,55/654,55`.
- **Grafiken:** sichtbare Symbolmitte bzw. Medaillonmitte in allen vier
  Ecken und allen drei Stilen exakt auf der jeweiligen Parkplatzmitte;
  maximale Abweichung **0,00**.
- Keine Konsolen- oder Seitenfehler; keine Figuren wurden verschoben, um
  einen Grafikfehler zu kaschieren. Sichtprüfung aller vier Ecken in beiden
  Größen über die Bilder unten.
- **Nicht geprüft:** Eine vollständige Partie wurde für diese reine
  Ausrichtungsänderung nicht erneut durchgespielt; Regeln und Spielablauf
  sind nicht angefasst.

**Sicherung:** `arbeit/sicherung/menschaerger-ecken-2026-11-27/` (beide
Ablagen vor der Änderung). Bilder: `arbeit/screens/maen-ecken-normal-pc.png`,
`…-maedchen-pc.png`, `…-buben-pc.png`, `…-normal-handy.png`,
`…-maedchen-handy.png`, `…-buben-handy.png`.

**Dateien:** `spiele/brettspiele/menschaerger/menschaerger.html` in beiden
Ablagen, `ENTWICKLUNGSSTAND.md`.

---

## 27.11.2026 · Mensch ärgere dich nicht: Würfel wieder in der Brettmitte

**Auftrag:** Den Würfel auf PC und Handy zurück in den freien zentralen
Bereich des Spielfelds holen – nicht in die Fußzeile, nicht links oder rechts
unten. Die zuletzt umgesetzten Eckpositionen, Stile, Blinkanimation,
Spielmechanik, Bots, Mehrspieler und Schnittstelle 2.0 bleiben unverändert.

**Umgesetzt**

- **Ursprüngliche Position wiedergefunden:** In der früheren Fassung lag der
  Würfel absolut mit `left:50%;top:50%;transform:translate(-50%,-50%)` im
  Brettbereich. Genau diese zentrale Overlay-Position ist wiederhergestellt.
- **Markup:** Der Würfel steht jetzt als `.wuerfel-mitte` innerhalb von
  `.buehne` direkt über dem SVG-Brett. Er wird nicht mehr im `.fuss`
  gerendert; die Fußzeile enthält nur noch Meldung und Spielerkarten.
- **Responsiv:** Die Position bindet an die Mitte des Brettbereichs, nicht an
  eine feste Pixelkoordinate. Die Würfelgröße bleibt 56 px am PC, 50 px bei
  höchstens 520 px Breite und 42 px in flachen Fenstern; Maus- und
  Touch-Bedienung sowie die Touchfläche bleiben erhalten.
- **Fußzeile:** Auf schmalen Fenstern stehen Meldung und Karten jetzt in zwei
  Zeilen statt in drei; unter dem Brett entsteht kein zusätzlicher
  Würfelbereich.
- **Unverändert:** Ecken und Figuren, dunkler Mädchenstil, Normal- und
  Bubenstil, Blinkanimation, Regeln, Laufwege, Bots, Mehrspieler und
  Schnittstelle 2.0.

**Geprüft**

- **1280 × 800:** Brett 641 × 641 px, Würfel 56 × 56 px, beide Mittelpunkte
  exakt `(640, 379,95)`; Abstand dx/dy **0**. Würfel vollständig im Brett,
  `footerContainsDie` **false**, Fußzeile 7,5 px unter dem Brett.
- **390 × 844:** Brett 380 × 380 px, Würfel 50 × 50 px, Mittelpunkte beide
  `(195, 402,23)`; dx/dy **0**. Keine Figur überlappt den Würfel; die
  Fußzeile beginnt 140,3 px unter dem Brett, kein Querscrollen.
- **Alle drei Stile** (Normal, Mädchen, Buben) auf PC und Handy: Würfel jeweils
  sichtbar und zentriert, keine Konsolen- oder Seitenfehler.
- **Kleinere Größen und Zoom:** 320 × 568, 844 × 390 (quer) sowie 260 × 562
  (entspricht 150 % Zoom auf 390 × 844) und 195 × 422 (200 % Zoom): Würfel
  jeweils mittig im Brett, vollständig sichtbar, kein Querscrollen. Bei
  200 % bleibt die Touchfläche 42 × 42 px.
- **Bedienung:** PC-Mausklick und Handy-Touch auf den Würfel starten den Wurf
  (`phase` `wartenWurf` → `wurfLaeuft`, Klasse `rollt`, Meldung „Der Würfel
  rollt …“). Trefferpunkt ist der Würfel selbst.
- **Nicht geprüft:** Eine vollständige Partie wurde für diese reine
  Positionsänderung nicht erneut durchgespielt; Regeln und Spielablauf wurden
  nicht angefasst.

**Sicherung:** `arbeit/sicherung/menschaerger-wuerfel-2026-11-27/` (beide
Ablagen vor der Änderung). Bilder: `arbeit/screens/maen-wuerfel-mitte-pc.png`,
`arbeit/screens/maen-wuerfel-mitte-handy.png`.

**Dateien:** `spiele/brettspiele/menschaerger/menschaerger.html` in beiden
Ablagen, `ENTWICKLUNGSSTAND.md`.

---

## 27.11.2026 · Gold Miner: Nuggets, Gestein und klarer Startbildschirm

**Auftrag:** Die Startseite von Gold Miner auf das Nötige reduzieren und
Gold sowie Steine als echte, plastische Objekte zeichnen – ohne Änderung an
Seil, Haken, Treffern, Einholen, Punkten, Zeit, Leveln, Schnittstelle oder
Mehrspieler.

**Umgesetzt**

- **Startseite:** Die Marke „Arcade · ein Spieler“ sowie der Erklärungstext,
  die Steuerungsliste und die Info-Karten unter der Grafikstil-Auswahl sind
  entfernt. Es bleiben Titel, kurze Unterzeile, Minenvorschau, die drei
  Stile **Normal / Mädchen / Buben**, die Startfunktion und die
  Vollbild-/Rückwege. Die Inhalte werden vertikal gleichmäßig verteilt; auf
  flachen Geräten bleibt der Startknopf sichtbar.
- **Gold:** Jedes Stück bekommt beim Anlegen eine feste, aus seinem Startwert
  erzeugte Silhouette aus 10–12 Punkten mit drei Buckeln, einer tiefen Kerbe
  und unterschiedlicher Streckung. Dazu kommen eine helle, kantige
  Goldfacette, eine zweite kleine Facette, eine dunkle Unterseite,
  Goldadern, eingesprengte Körner und zwei scharfe Glanzpunkte. Die Pfade
  werden einmal als `Path2D` gebaut und nicht pro Bild neu erzeugt.
- **Stile der Nuggets:** Normal warmes Amber-/Goldbraun, Mädchen ein
  gedämpftes Roségold mit warmem Goldkern, Buben ein kräftiges Messinggold
  mit harten Facetten. Das sind eigene Paletten und Formkanten, keine
  bloßen Farbfilter über einem Kreis.
- **Steine:** 7–10 unregelmäßige, kantige Eckpunkte, graue bis lavendelgraue
  bzw. kalte Schiefer-Paletten, eine helle Bruchfläche, eine dunkle Flanke,
  feine Gesteinskrümel, Risse mit Gegenlicht und ein dezenter Glanzreflex.
  Buben bekommen zusätzlich eine rote Erzader, Mädchen eine weichere, aber
  weiterhin eckige Silhouette. Gold und Gestein sind damit auch klein klar
  unterscheidbar.
- **Beutel:** Der frühere Kreis mit „?“ ist jetzt ein kleiner Lederbeutel
  mit Band, Knoten und Volumen; der Wert bleibt verdeckt. Der Diamant behält
  seine bisherige Form.
- **Startvorschau:** Die Minenvorschau zeigt einige der neuen Gold- und
  Gesteinsstücke. Sie wird einmal je Vorschaugröße gebaut und dann nur
  gezeichnet.
- **Objektplatzierung:** Das Spielfeld wird jetzt **vor** dem Erzeugen eines
  Levels gemessen; vorher entstanden die Objekte für die alte, versteckte
  Fläche und wurden beim ersten `ResizeObserver` an den Rand gedrückt. Bei
  einer späteren Größenänderung werden die Positionen mit dem Feld skaliert,
  danach mit einem Zeichenrand begrenzt. Dadurch überlappen die größeren
  Brocken nicht mehr und ragen nicht aus dem Feld.
- **Unverändert:** Regeln, Hakenfahrt, Trefferradius, Einholtempo, Punkte,
  Zeit, Level, Pause, Neustart, Speicher und die Schnittstelle 2.0.0.

**Geprüft**

- Drei Stile auf PC 1280×800 und Handy 390×844 gestartet und fotografiert;
  Gold, Steine, Beutel und Diamant sichtbar, keine Konsolenfehler.
- **Treffer und Einholen** über die öffentliche Aktion `werfen` in allen drei
  Stilen: Normal holte in 20 Versuchen u. a. `goldM` (+120), `goldS` (+50),
  `goldG` (+260) und `steinS` (+11); Mädchen in 13 Versuchen `goldS` (+50),
  `goldM` (+120) und `steinS` (+11); Buben in 13 Versuchen `goldS` (+50),
  `goldM` (+120) und `steinS` (+11). Die Punkte entsprachen den Objektwerten,
  die Objekte verschwanden nach dem Einholen.
- Startseite: Marke, Erklärung, Steuerungsliste und Info-Karten sind nicht
  mehr vorhanden; alle drei Stilknöpfe und der Startknopf funktionieren;
  Stilwechsel aktualisiert Vorschau und Welt. Geprüft bei 320×568,
  360×640, 390×844, 844×390, 1024×600 und 1280×800: kein Scrollen,
  Startknopf im Bild.
- PC-Mausklick und Handy-Touch starten den Haken (`tick` 1); Pause/Weiter,
  Neu und Zum Start funktionieren. Bei einer Größenänderung von 390×844 auf
  844×390 bleiben Level und Objektzahl erhalten, keine Fehler.
- Schnittstelle **2.0.0**: `version`, `whenReady()`, `zustand()`-Kopie,
  Fehlercodes `unbekannte-aktion`, `daten-muessen-objekt-sein`,
  `unbekanntes-feld`, `level-muss-1-bis-50-sein`; direkter Start über die
  Seite funktioniert.
- **Connector-Rahmenpfad:** In einem Rahmen mit `TGStil.lokal()` = Mädchen
  öffnete das Spiel mit `stilWahl:false`, verborgener Stilkachel und
  Mädchenwelt; `start()` startete die Partie. `start({stil:'buben'})`
  wechselte bei laufender Partie auf Buben, Level 1 und Takt 0 blieben
  gleich, die Stilkachel blieb verborgen. Keine doppelte Auswahl.
- **Offen:** Eine vollständige 50-Level-Partie wurde nicht durchgespielt. Der
  laufende Connector auf 8766 zeigt weiterhin die alten 404-Aufrufe und
  öffnet keinen Spielkatalog; deshalb wurde der Connector-Pfad über den
  beschriebenen `TGStil`-Rahmen geprüft, nicht über eine echte Connector-Runde.

**Sicherung:** `arbeit/sicherung/goldminer-2026-11-27/` (beide Ablagen vor
der Änderung). Bilder: `arbeit/screens/gold-final-*`, `gold-start-final-*`,
`gold-game-final-*`, `gold-game-touch-maedchen.png`.

**Dateien:** `spiele/goldminer/goldminer.html` in beiden Ablagen,
byte-gleich; `SPIELE-SCHNITTSTELLE.md` (Grafik-Absatz aktualisiert). Keine
anderen Spiele geändert.

---

## 27.11.2026 · Mensch ärgere dich nicht: Ecken, Mädchenstil, Blinken, Würfel

**Auftrag:** In allen drei Stilen die vier Eckbereiche und Figuren mittig
ausrichten, den zu hellen Mädchenstil hochwertiger machen, das Blinken
aktiver Figuren sichtbar machen und den Würfel wieder in die Mitte des
Spielbereichs holen – ohne Änderung an Regeln, Laufwegen, Bot oder
Mehrspieler.

**Ursache der versetzten Ecken:** `PARKPLATZ` lag mit den Offsets
`[1.25,1.25] … [2.75,2.75]` um **0,5 Zellen** (bei 800 Einheiten
**36,36 px**) rechts und unten neben der Mitte des 3×3-Startbereichs.
Gemessen lag der Figurenmittelpunkt vorher **36,4 px** neben dem
Bereichsmittelpunkt; Eckmedaillon und Parkkreise hatten dadurch keine
gemeinsame Mitte.

**Umgesetzt**

- **Ecken:** `PARKPLATZ` auf `[0.75,0.75] … [2.25,2.25]` gesetzt – also
  symmetrisch um 1,5 Zellen. Gemessene Home-Transforms:
  `563.64 / 672.73`; Gruppenmittelpunkte **109.095 / 618.185** in allen vier
  Ecken und allen drei Stilen. Die Figuren selbst wurden nicht einzeln
  verschoben.
- **Mädchenstil:** tiefes Pflaumen-/Beerenbild statt grellem Rosa:
  Seite `#160912`, Panel `#2a1020`, Felder `#24101a`, Parkflächen `#1d0b14`,
  Text `#f8edf1`, Gold/Rosé `#f0c9a8`/`#c98a9d`. Ruhiger Abendgarten mit
  Bach, Goldrosetten, Rosenmedaillon in jeder Basis, Blütenfeldern und
  Goldrahmen statt großer weißer Flächen. Die Figurenfarben sind tiefer
  (Rosa `#c9366b`, Zitrone `#c9971b`, Minze `#238a68`, Veilchen `#7140b8`).
- **Warum die Seite am PC zunächst hell blieb:** `/spiele/stil.js` setzt für
  Mädchen am PC eine helle Herz-Kulisse und `color-scheme:light`. Die
  spezifischere Regel `html[data-stil="maedchen"][data-stil] body` und
  `stilSetzen()` (setzt `color-scheme:dark` und `theme-color`) übernehmen
  jetzt die Spielpalette. Kontraste gemessen: Text/Panel **15,40 : 1**,
  gedämpft/Panel **8,65 : 1**, Gold/Panel **11,42 : 1**, Ring/Park
  **16,24 : 1**.
- **Blinken:** pro aktiver Figur ein weicher Schein hinter dem Stein, ein
  dunkler Halo (11 Einheiten) und ein heller Ring (6 Einheiten), dazu
  helligkeitspulsierender Stein. Takt **0,62 s**; Ringopazität
  **1 → 0,28**, Schein **0,30 → 0,06**. Gemessene Ringopazitäten im Verlauf:
  `0,974 · 0,844 · 0,660 · 0,442 · 0,320 · 0,281 · 0,329` – der Ring
  verschwindet nicht mehr im hellen Feld.
- **Würfel:** Fuß wieder als Grid `minmax(0,1fr) auto minmax(0,1fr)`:
  Meldung links, Würfel Mitte, Karten rechts. Am Handy gestapelt Würfel →
  Meldung → Karten. Gemessen: PC Würfelmittelpunkt **640 / 726**, Fußmitte
  **640**; Handy **195 / 707,5**, Fußmitte **195**; Touchfläche **50 px**
  (PC 56 px). Maus- und Touch-Klick zeigen danach „Deine leuchtende Figur
  anklicken (oder 1–4).“.
- **Feldquadrat:** `.buehne` ist Größencontainer, `.brett` nimmt
  `min(100cqw,100cqh)`; dadurch bleibt das Brett auch bei 320×568 quadratisch
  (vorher 266 × 258,6). Gemessen: 320 → **256 × 256**, 360 → **350 × 350**,
  390 → **380 × 380**, 768 → **752 × 752**, 1280 → **624 × 624**; überall
  `scrollWidth = Fensterbreite`, kein Quer- oder Hochscrollen.

**Geprüft**

- Drei Stile × PC 1280×800 und Handy 390×844: Start, Vier-Spieler-Partie,
  alle 16 Figuren, vier Ecken, keine Konsolenfehler.
- Schnittstelle **2.0.0**: `version`, `whenReady()`, `start({anzahl:4,
  stil, seed})`, `zustand()`, `aktion('wuerfeln')` und
  `aktion('figur',{figur})` in allen drei Stilen; Figur zog von `0` nach
  `4`, der Zug wechselte zum nächsten Spieler.
- Würfel mit Maus (PC) und `touchscreen.tap` (390×844 und 320×568);
  Auswahlphase und Meldung danach korrekt.
- Größen 320×568, 360×640, 390×844, 768×1024 und 1280×800 in allen drei
  Stilen; Brett quadratisch, Karten im Fuß sichtbar, kein Überlauf.
- Connector-Seite geladen; der laufende Connector-Prozess auf 8766 liefert
  für `/api/maensync` weiterhin **404** (alter Prozess), daher keine echte
  Online-Runde gestartet. Der direkte Spielstart und die Connector-
  Adressform (`?modus=online&room=…&id=…&anzahl=…`) wurden geprüft.

**Sicherung:** `arbeit/sicherung/menschaerger-2026-11-27/` (beide Ablagen
vor der Änderung). Bilder: `arbeit/screens/maen-start-*`,
`maen-neu-*`, `maen-ringpeak-*`, `maen-size-*`.

**Dateien:** `spiele/brettspiele/menschaerger/menschaerger.html` in beiden
Ablagen, byte-gleich; keine anderen Spiele geändert.

---

## 27.11.2026 · 10.000: eigenständig, Bot und Schnittstelle 2.0

**Auftrag:** Das vorhandene Würfelspiel 10.000 vollständig überarbeiten –
eigener Startbildschirm mit Stil-, Modus-, Spielerzahl- und Bot-Wahl, drei
eigenständige Grafikstile, Bot, lokale Mehrspieler-Runden, Pause, Regeln,
Schnittstelle 2.0 und Connector-Anbindung. Danach testen, dokumentieren,
beide Ablagen spiegeln und committen.

**Umgesetzt**

- **Sicherung vor dem Umbau:**
  `sicherung/zehntausend-schnittstelle-2026-11-27/zehntausend.html`
  (71.147 Bytes, alte Fassung mit `/api/zehntausend` und `?room&id`).
- **Mechanik und Wertung unverändert:** sechs Würfel, 1 = 100, 5 = 50,
  Dreierpasch 1er = 1.000, sonst Augenzahl × 100, jeder weitere gleiche
  Würfel zählt den Satz erneut; nur 1/5/ab drei gleiche zählen; Einstieg
  350, Ziel 10.000. Die Wertung steht als reine `REGELN`-Funktion im Spiel
  und wechselt während einer Partie nicht.
- **Startbildschirm im Spiel:** Stilwahl (Normal/Mädchen/Buben), Spielart
  (Einzelspieler gegen Bot oder Mehrspieler an einem Gerät), Spielerzahl
  2–6, je Platz Name und Mensch/Bot, „Spiel starten“ und „Laufendes Spiel
  fortsetzen“. Ohne Spielstand öffnet immer diese Auswahl.
- **Modi:** Einzelspieler gegen einen oder mehrere Bots, zwei bis sechs
  Menschen an einem Gerät, gemischte Sitze. Wer am Zug ist, steht in der
  Kopfzeile und wird in der Spielerliste hervorgehoben.
- **Bot gehört zum Spiel:** Er spielt über dieselben Funktionen wie ein
  Mensch, legt alle wertbaren Würfel beiseite (1er, 5er, vollständige Sätze),
  würfelt weiter und sichert ab 350 Punkten – früher bei wenigen freien
  Würfeln oder ab 1.000 Punkten. Seine Auswahl ist kurz sichtbar, damit der
  Zug nachvollziehbar bleibt. Ein Wächter verhindert doppelte Bot-Züge.
- **Drei eigenständige Stile:** Normal ist grüner Filz mit Elfenbeinwürfeln
  und Gold; Mädchen ist Aubergine/Rosé mit gepunkteten Würfeln, weichem
  Innenrahmen und Serifentiteln; Buben ist Anthrazit/Stahl mit gebürsteten
  Würfeln, roter Kante und klaren Linien. Die Augenzahlen bleiben in allen
  Stilen dunkel bzw. hell genug und werden von keiner Dekoration überdeckt.
- **Animationen:** Würfelrollen mit gestaffelter Bewegung, deutliche
  Auswahl-/Sichert-Marken, Punktepuls bei Änderungen, Fehlwurmkarte mit
  den stehen gebliebenen Würfeln. `prefers-reduced-motion` schaltet die
  Bewegung ab.
- **Bedienung im Spiel:** Würfeln/Weiterwürfeln, Würfel antippen, Punkte
  sichern, Pause (Knopf und Taste P), Neues Spiel, Regeln (Kopfknopf und
  „Spielregeln“), Spielstand (ständige Spielerliste), Rückweg über
  `zurueck.js`/`window.tgZurueck()`.
- **Schnittstelle 2.0:** `window.Zehntausend`, Version **2.0.0**, eingefroren;
  `version`, `bereit`, `whenReady()`, `start()`, `einstellungen()`,
  `zustand()`, `startbildschirm()`, `stop()`, `pause()`, `resume()`,
  `aktion(name, daten)`. Ereignisse `zehntausend:bereit|ende|pausiert|weiter`
  (zusätzlich `wurf`, `zug`, `neu`); `:ende` kommt vor dem Ergebnisfenster.
- **Aktionen:** `wuerfeln`, `weiterwuerfeln`, `sichern`, `waehlen {index}`,
  `wertung {wuerfel}` (reine Prüf-/Testfunktion), `pause`, `fortsetzen`,
  `neu`, `start`, `regeln`. Unbekannte Namen antworten
  `unbekannte-aktion`, falsche Zustände `aktion-nicht-erlaubt`, `daten`
  außerhalb eines Objekts `daten-muessen-objekt-sein`.
- **Konfiguration:** `start({stil})` stellt nur die Welt ein und lässt den
  Startbildschirm stehen (Connector-Weg); `start({spieler, modus, anzahl})`
  beginnt eine Partie; `start()` ohne Angaben übernimmt die sichtbare Auswahl. Erlaubt sind `stil`, `spieler`, `modus` (`bot`/`lokal`)
  und `anzahl` (2–6); eigene Fehlercodes
  `spieler-muessen-2-bis-6-sein`, `spieler-braucht-mindestens-einen-menschen`,
  `modus-muss-bot-oder-lokal-sein`, `anzahl-muss-2-bis-6-sein`.
- **Speicher:** laufende Partie `tg-zehntausend-spiel` (alter Schlüssel
  bewusst behalten), eigene Stilwahl `zehntausendStil`. Eine beendete oder
  abgebrochene Partie hinterlässt kein „Fortsetzen“.
- **Connector:** `Zehntausend` steht in `SPIEL_API_NAMEN` beider
  `connector.html` (byte-gleich, je 2 NUL-Bytes unverändert). Der Rahmen
  übergibt den Stil über `start({stil})`; die Stilkachel im Spiel ist dann
  verborgen. Kein `/api/`, kein `room`/`id` aus der Adresse, kein
  `token.js`/`ergebnis.js` mehr.
- **Regelblatt:** `regeln.js` öffnet die vorhandene Regelliste; die neue
  Zeile erklärt Spielarten und die feste Wertung. Der geschlossene Schleier
  wird in dieser Seite ausgeblendet, damit er keine Klicks abfängt.

**Geprüft (Chromium, lokaler HTTP-Server, 103 Prüfungen grün)**

- Startbildschirm mit drei Stilen, zwei Spielarten, 2–6 Spielern und
  Namens-/Bot-Zeilen; `window.Zehntausend` 2.0.0, `whenReady()`, Kopien von
  `einstellungen()`/`zustand()`, nur ein neuer globaler Name.
- Wertungen: 1, 5, 1+5, 3×1, 4×2, 5×6, 4×1, unwertbare 2/3/4/6, ungültige
  Würfel; unbekannte Aktion und falsche Daten.
- Alle Startfehler ohne Zustandsänderung; `start` während einer Partie,
  `pause` ohne Partie, `stop` und Fortsetzen.
- Einzelspieler gegen Bot: Fehlwurf mit stehenden Würfeln, Verlust der
  Zugpunkte, Bot würfelt selbstständig und gibt den Zug zurück; 1 und 5
  sichern, unter 350 gesperrt, „Weiterwürfeln“ beschriftet.
- Vier Menschen, Pause über Kopfzeile und API, genau ein
  `pausiert`/`weiter`-Ereignis, Regeln öffnen/schließen, Neues Spiel mit
  Fortsetzen.
- Sieg über einen gespeicherten Stand: `:ende` mit `grund: gewonnen`,
  Punkte und Gewinner, Rangliste, Neustart ohne Fortsetzen-Leiche.
- Drei Stile und Connector-Weg (`start({stil})` startet nicht, Kachel weg)
  geprüft; Prüfbilder für Start, Normal, Mädchen, Buben (390×844) und
  Normal (1280×800) angesehen: Würfel groß und lesbar, keine abgeschnittenen
  Elemente, kein Querscrollen, Handy-Würfel 3 × 2, Knöpfe ≥ 44 px.
- Keine Konsolen- oder Seitenfehler in den Läufen.

**Offen**

- **Connector-Online für 10.000:** Die alte Fassung holte Raum und Kennung
  aus der Adresse und sprach `/api/zehntausend` an. Das ist nach
  Schnittstelle 2.0 verboten und wurde entfernt. Eine 2.0-Brücke im Connector
  (Lobby, `start({online})`, `aktion("stand")`, `zehntausend:aktion`) gibt es
  für dieses Spiel noch nicht. Die Online-Kachel im Mehrspieler-Register
  bleibt deshalb vorerst bestehen, öffnet aber die lokale Auswahl; das ist
  der einzige offene Punkt. Lokal, gegen Bot und im Connector-Rahmen (Stil)
  ist die Fassung vollständig.

---

## 27.11.2026 · Solitär und Spider: Grafik, Karten und Layout

**Auftrag:** Karten vollständig deckend machen, die drei Grafikstile sichtbar
hochwertiger und eigenständiger gestalten und den Spider-Fuß auf PC und Handy
sauber vom Tableau trennen. Beide Spielvarianten, alle drei Stile und beide
Geräte prüfen; Schnittstelle 2.0 und Regeln unverändert lassen.

**Umgesetzt**

- **Deckende Karten:** Die Ursache war `opacity` auf `.karte.verdeckt`
  (0,52), `.karte.bedeckt` (0,8) sowie auf Kartenzeichen und Zieh-Geist.
  Karten, Zeichen und Rückseiten stehen jetzt in allen Stilen auf `opacity:1`;
  Vorderseiten haben ein deckendes Elfenbein, Rückseiten einen deckenden
  Grund. Keine Transparenz, kein Filter, kein Blend-Modus auf Karten.
- **Normal:** tiefer dunkelgrüner Tisch mit Goldkante, Elfenbeinkarten,
  Goldmedaillon-Rückseite, ruhige Licht- und Schattenflächen.
- **Mädchen:** dunkles Beeren-/Pflaumengarn statt heller Fläche, Rosé- und
  Violettakzente, Beerenrosette als Rückseite, eigene Kartenkante und
  Dekorlinien; die Karten bleiben hell und gut lesbar.
- **Buben:** Anthrazit/Teal statt Blau-Einfärbung, eigenes Chevron-Muster auf
  der Rückseite, harte Kanten und türkise Akzente.
- **Spider-Fuß:** eigener Bereich mit den zwei beschrifteten Zonen „Ablagen“
  (8 Plätze) und „Vorrat“; 30 px Abstand zum Tableau, eigene Tischkante und
  klare Trennung. Auf dem Handy liegen die acht Ablagen und der Vorrat
  nebeneinander in der Fußzeile; auf dem PC rechts neben den Ablagen.
- **Handy-Spider:** Das Tableau bricht unter 521 px auf zwei Reihen mit je
  fünf Spalten um. Dadurch bleiben die Karten 69 px breit (statt 35 px) und
  die Werte lesbar; auf dem PC bleiben zehn Spalten nebeneinander. Die
  Kartenabstände im Stapel werden für die schmale Ansicht leicht verringert.
- **Hintergrund:** Die von `stil.js` gesetzte 264-px-Herzkachel und ihre
  `background-size` werden für Solitär überschrieben; der Mädchen-PC zeigt
  dadurch keinen waagerechten Kachelstreifen mehr.
- **Schnittstelle 2.0 und Regeln:** unverändert. Keine externen Ressourcen,
  keine neuen Bibliotheken.
- **Sicherungen:** `sicherung/solitaer-2026-11-27-vor-final/` (Arbeitsstand
  vor der letzten Layoutrunde) und `sicherung/solitaer-2026-11-27-mirror-vorher/`
  (Spiegel vor dem Angleichen).

**Geprüft (echte Läufe)**

- `solitaer_pruefen.js`: **236/236** Prüfungen bestanden, alle sechs
  Kombinationen aus Normal/Mädchen/Buben und Klondike/Spider, jeweils PC
  1280×800 und Handy 390×844.
- Zusätzlich 360×740 als Layout- und Sichtprüfung: kein Querscrollen,
  Karten 62,8 px breit, Fuß vollständig sichtbar.
- Geprüft wurden Deckkraft und Alpha-Werte aller Karten, Vorder-/Rückseiten,
  Rot/Schwarz-Kontrast, Drag, Nachziehen, Aufdecken, Neustart, keine
  Querscrollung, Fußabstand, acht Ablagen und Vorrat, keine Konsolenfehler.
- Zusätzlicher Smoke-Test aller fünf Varianten in allen drei Stilen auf PC
  und Handy: keine transparenten Karten, kein Querscrollen, keine Fehler.
- Sichtprüfung per Screenshots: `screens/sol-final-*.png`,
  `screens/sol-start-*.png`, `screens/sol-zoom-handy-spider.png`.
- Beide Ablagen (`Documents/Homepage` und `Documents/GitHub/Homepage`)
  byte-gleich; `sha256sum` ohne Unterschied.

**Offen:** nichts aus dem Auftrag.

---

## 26.11.2026 · Solitär: eigenständig und Schnittstelle 2.0

**Auftrag:** Solitär vollständig eigenständig machen – eigener
Startbildschirm mit Stil- und Variantenwahl, drei eigenständige Kartenbilder,
Pause, Statistik, Neumischen – und über die Spiele-Schnittstelle 2.0 in den
Connector aufnehmen. Danach testen, dokumentieren, beide Ablagen spiegeln,
committen.

**Umgesetzt**

- **Neue Schnittstelle** `spiele/kartenspiele/solitaer/js/schnittstelle.js`
  (zuletzt geladen): `window.Solitaer` mit `version` 2.0.0, `bereit`,
  `whenReady()`, `start()`, `einstellungen()`, `zustand()`,
  `startbildschirm()`, `stop()`, `pause()`, `resume()`, `aktion()`; Ereignisse
  `solitaer:bereit|ende|pausiert|weiter`. `:ende` kommt vor dem
  Ergebnisfenster und trägt `grund`, `punkte`, `level`, `levelAnzahl`, `tick`,
  `modus`, `zuege`, `zeit`, `stil`.
- **`start({stil})`** stellt nur die Welt ein (Connector-Weg, keine
  Doppelwahl); `start({modus})` beginnt eine Partie. Unbekannte Felder,
  falsche Typen und unbekannte Varianten antworten mit den festen Codes;
  eigener Code `unbekannter-modus` statt `unbekannte-stufe`.
- **Rahmen** `js/shell.js`: Pause/Fortsetzen (`#btnPause`, Taste P, Uhr steht,
  Brett gesperrt, `#pauseSchleier`), Statistikfenster je Variante
  (`#btnStat`), Stilwahl (`#stilKnoepfe`, eigenes Gedächtnis
  `solitaer-stil`), Ergebnismeldung vor dem Dialog, Eingabesperre in der
  Pause; Tastenkürzel halten sich auf Knöpfen und Feldern zurück.
- **Startbildschirm** `solitaer.html`: Stilwahl-Kachel über den fünf
  Varianten, „Neues Spiel“, Pause, Rückgängig, Tipp, Automatisch, Regeln,
  Statistik, Pausenfenster; Rückweg über `zurueck.js`/`window.tgZurueck()`.
- **Drei eigenständige Kartenbilder** in CSS: Normal Elfenbein/Goldrand,
  Nachtblau-Rückseite mit Ovalmedaillon; Mädchen Rosé, runde Ecken,
  gepunktete Innenlinie, Rautengitter mit Schleifenmitte; Buben Anthrazit,
  harte Kante, rote Schrägkante, Kohlefaser mit Winkeln. Je Stil eigene
  Vorderseite, Rückseite, Radius, Schrift und Dekor – keine Umfärbung.
- **Connector:** `Solitaer` in `SPIEL_API_NAMEN` beider `connector.html`
  (byte-gleich, 2 NUL-Bytes unverändert).
- **Sicherung:** `sicherung/solitaer-schnittstelle-2026-11-27/` mit
  `solitaer.html` und allen `js/`-Dateien vor dem Umbau.

**Geprüft (echte Läufe)**

- Direktstart ohne Connector: Startbildschirm, Stilwahl sichtbar, fünf
  Varianten, `window.Solitaer` vollständig, `whenReady()` und
  `solitaer:bereit` je einmal.
- Drei Stile: Datenattribut, gespeicherte Wahl, sichtbarer Knopf und
  Kartenvorderseite/-rückseite unterscheiden sich messbar; gespeicherter
  Stil überlebt das Neuladen.
- Connector (Port 8765, IPv6): Katalog enthält Solitär; Rahmen öffnete die
  Seite; Stil „Mädchen“ kam an, Kachel verborgen; Partie im Rahmen, 1 Zug;
  Umschalten auf „Buben“ wechselte das Bild bei laufender Partie.
- TriPeaks echt gespielt: 7 Verluste („Keine Züge mehr“) und ein Gewinn in
  51 Zügen mit Ergebnisfenster; `:ende` jeweils mit richtigem `grund`.
- Klondike: Nachziehstapel, Rückgängig (1 → 0), Tipp, Neumischen (Züge 0,
  Zeit 0); Pause hält die Uhr an, `resume` startet sie wieder.
- Handy 390×844 (Spider 33-px-Karten, Klondike 46 px): kein horizontaler
  Überlauf, keine Karte außerhalb, Stapel sichtbar, Rang 12,5 px; Tippen ohne
  Zug. PC 1280×800: Karten 151 px, Bahn 1118 px.
- Keine Seitenfehler in allen Läufen.

**Offen:** nichts aus dem Auftrag.

---

## 26.11.2026 · Verrücktes Labyrinth: zwei neue Kachelwelten und Schnittstelle 2.0

**Auftrag:** Das Kachelspiel „Das verrückte Labyrinth“ neu gestalten – Mädchen-
und Bubenstil vollständig eigene Kachelsätze, Normalstil unverändert – und auf
die Spiele-Schnittstelle 2.0 umstellen, inklusive Connector, Mehrspieler,
Sicherung, Doku und Commit.

**Wichtig vorab:** `spiele/brettspiele/labyrinth/labyrinth.html` ist das
Kugel-Labyrinth „Labyrinth Plus!“ und blieb unangetastet. Umgebaut wurde
`spiele/brettspiele/verruecktes-labyrinth/verruecktes-labyrinth.html`.

### Grafik

- **Normalstil unverändert.** Der Zeichenpfad für Klassisch, Schloss, Dungeon,
  Wald und Eis ist byte-gleich geblieben. Gegenprobe: Brett-Canvas im
  Sicherungsstand und im neuen Stand mit gleichem Zufallsstart gerendert –
  beide PNGs sind identisch (271.035 B, `NORMAL-GLEICH: true`).
- **Mädchen – „Porzellan & Rosengarten“:** eigener Kachelmaler
  (`kachelMaedchen`). Rosé-Marmor als Mauerwerk mit Adern, Perlmuttpunkten und
  Porzellankante, Elfenbein-Gang mit Rosenholz-Fassung, Perlenreihe an beiden
  Gangrändern, Goldrahmen mit Eckherzen auf festen Kacheln, runde
  Porzellan-Medaillons für die Gegenstände.
- **Buben – „Werkstatt & Kommandobasis“:** eigener Kachelmaler
  (`kachelBuben`). Gebürstete Stahlplatten mit Nieten, helle Stahlbahn mit
  Trittprofil und gestrichelter Leitlinie, sechseckige Stahlplaketten für die
  Gegenstände; feste Kacheln mit verschraubtem Rahmen.
- **Neue Motive statt Palettentausch:** Mädchen Herzkrönchen, Seidenschleife,
  Rosenblüte, Schmetterling, Handspiegel, Teetasse, Perlenkette, Märchenbuch;
  Buben Kompass, Zahnrad, Anker, Energieblitz, Eiskristall, Signalflagge,
  Seilrolle, Taschenlampe. Alle als Vektorpfade im Einheitsquadrat, DPR-scharf.
- **Mechanik unberührt:** Beide Maler benutzen dieselbe `wegUnion`-Geometrie
  und dieselbe Gangbreite (46 % der Kachel) wie der Normalstil. Wege, Drehungen
  und Schieben ändern sich dadurch nicht – nur Material und Zierrat.
- Seitenfarben der beiden Welten (CSS-Variablen) auf Rosé-Gold bzw. Stahl-Cyan
  abgestimmt.

### Schnittstelle 2.0

- Genau ein Global `window.Labyrinth` mit `version` (`2.0.0`), `bereit`,
  `whenReady()`, `start()`, `einstellungen()`, `zustand()`, `startbildschirm()`,
  `stop()`, `pause()`, `resume()`, `aktion()`; alle Aufrufe antworten, werfen
  nie, Fehlercodes wie im Standard. `zustand()`/`einstellungen()` liefern tiefe
  Kopien.
- Ereignisse `labyrinth:bereit`, `:ende` (mit `grund`, `tick`, `sieger`,
  `aufgaben`), `:pausiert`, `:weiter`.
- Aktionen: `schiebe`, `ziehe`, `drehen`, `beenden`, `stand`, `neu`, `thema`,
  `platz`. `start()` prüft Stil, Feldgröße (5/7/9), Aufgaben (3/5/7),
  Spielerzahl (2–4), Modus, Stufe und Kreuzungen, bevor irgendetwas übernommen
  wird.
- **Kein `/api/` mehr im Spiel:** `online-sync.js`, `token.js` und
  `TG_ONLINE_SPIEL` sind aus der Seite entfernt; sie liest weder `room` noch
  `id` aus der Adresse. Das Brett bleibt bei fehlender Serverantwort stehen.
- Vollbild-Knopf in der Titelleiste (echtes Fullscreen, Pseudo-Vollbild als
  Rückfall, Escape verlässt es).
- Startbildschirm bleibt der eigene Einrichtungsbereich (Titel, Beschreibung,
  Spielart, Feld, Aufgaben, STARTEN).

**Bewusste Abweichungen (in der Spieldoku genannt):**
1. Die Startbildschirm-Vorlage aus Abschnitt 9 (`spiel-start.js`,
   `TGSpiel.feldPassen`) wurde **nicht** übernommen, weil sie Raster und
   Feldmaß ändert – das würde den bestandsgeschützten Normalstil verschieben.
   Der vorhandene Einrichtungsbereich erfüllt „eigener Startbildschirm“.
2. Für Werte ohne Levelbedeutung (Stil, Feldgröße, Aufgaben, Spielerzahl,
   Modus) gibt es den zusätzlichen Code `ungueltiger-wert` mit `feld`; die
   Level-Codes der Vorlage passen hier nicht.

### Connector

`spiele/connector/connector.html` bekam eine Brücke `labBrueckeAn` /
`labTaktStarten`: Sie sichert den Sitz (`rejoin`, sonst `create`), zeigt die
Lobby (Bereit, + Computer, Start), baut als Gastgeber die Aufstellung und
schickt sie als Stand, schiebt Serverstände über `aktion("stand")` hinein und
meldet eigene Züge als vollständigen Stand zurück. Der gewählte Connector-Stil
hat Vorrang (`aktion("thema")`, auch bei Stilwechsel). Ein mitlaufender Schutz
verhindert Senden und Laden mitten in einer Animation – sonst könnte ein
Serverstand einen laufenden Figurenzug verschlucken.

### Tests (Playwright, Chromium)

- **Vertrag:** alle Pflichtfelder vorhanden; `start(5)` →
  `konfiguration-muss-objekt-sein`, unbekanntes Feld, ungültiger Stil,
  Feldgröße 6 und Stufe als Zahl werden abgelehnt; `start()` während einer
  Partie → `spiel-laeuft`; `pause()/resume()` mit bzw. ohne Partie korrekt;
  `schiebe` in der Pause → `aktion-nicht-erlaubt`; `zustand()`-Kopien
  unabhängig; `stop()` setzt den Startbildschirm zurück.
- **Kacheln und Drehungen:** künstliches 4×4-Brett mit allen 16
  Seitenkombinationen (Gerade, Kurve, T-Kreuz, Kreuzung, geschlossen) in
  Klassisch, Mädchen und Buben; je Kachel alle vier Kanten pixelweise gegen
  `offen` geprüft – **0 Abweichungen**. Freier Stein: vier Drehungen, jede
  Kante pixelweise gegen `dreheSeiten` – korrekt in allen drei Stilen.
- **Ganze Partie:** 5×5, zwei Spieler am Gerät, geschieht über `aktion()`;
  alle Aufgaben eingesammelt, zurück ins Startfeld, `labyrinth:ende` genau
  einmal mit `grund: "gewonnen"`, Figuren immer im Feld, keine Ablehnung.
- **Handy 390×844:** STARTEN sichtbar ohne Scrollen, Brett quadratisch und
  innerhalb der Breite, kein Querscroll, Vollbild-Knopf vorhanden.
- **Connector live** (Port 8779, echter Go-Server): Labyrinth-Runde im Raum
  geöffnet, Lobby 1/2, Computer ergänzt, Bereit, Start; Aufstellung wurde vom
  Gastgeber gebaut, Schieben und Figurenzug kamen über den Server an
  (Stand-Version 7, Figur auf Feld 1, Bot zog danach), keine Seitenfehler.
- **Normalstil:** Brett-PNG aus Sicherung und neuem Stand identisch.

### Dateien

- `spiele/brettspiele/verruecktes-labyrinth/verruecktes-labyrinth.html`
- `spiele/connector/connector.html`
- `sicherung/labyrinth-2026-11-26/` (beide Dateien vor dem Umbau)
- `ENTWICKLUNGSSTAND.md`

Keine Go-Dateien, keine anderen Spiele, keine gemeinsamen Helfer geändert.

---

## 26.11.2026 · Alter Endpunkt /api/menschaerger entfernt

**Auftrag:** Der alte Endpunkt `/api/menschaerger` lief live noch neben
`/api/maensync`, obwohl ihn keine Seite mehr aufrief. Prüfen, wer ihn noch
braucht, und ihn samt Handler entfernen, wenn er wirklich tot ist – Labyrinth
und die übrigen Spiele bleiben unberührt.

**Wer rief ihn noch auf?** Keine ausgelieferte Seite: `connector.html` spricht
seit dem Standweg-Umbau ausschließlich `/api/maensync` an. Aufrufer waren nur
`bot_steuerung_test.go` (beide Fassungen) und die **veraltete** Android-Kopie
`spiele/android/app/src/main/assets/www/spiele/brettspiele/menschaerger/menschaerger.html`
(40.359 B vom 25.11., während die echte Seite 110.037 B vom 26.11. hat) – der
Asset-Spiegel ist ohnehin breit veraltet (84 Dateien gleich, 97 anders,
412 Seiten fehlen) und wurde deshalb nicht als „braucht den Endpunkt“ gewertet.

**Beweis am laufenden Router** (`httptest` um `routes()`, beide Fassungen):

| Anfrage | Antwort |
| --- | --- |
| `GET /api/menschaerger?room=ABC123` | 404 `404 page not found` (ServeMux – Route weg) |
| `GET /api/maensync?room=ABC123` | 404 `Spielraum nicht gefunden` (Handler antwortet) |
| `GET /api/maensync` | 400 `Spielraum-Code fehlt` |

**Entfernt (in `connector/` und `connector-online/`, je identisch):**

- Route in `server.go` und der Handler samt Altspiel in `menschaerger.go`
  (`MenschaergerGame`, `MaenPlayer`, Regeln, `maenSync`-fremde Sitzlogik).
- `bot_mensch.go` (`maenBotTreiben`, `maenBotSitz`, `maenRoll`, `maenBotFigur`,
  `maenBotWert`) – der alte Bot hing allein am alten Handler.
- `types.go`: Feld `Room.Menschaerger`, der Statustext-Zweig und
  `cloneMenschaerger`.
- `leave.go`: der Sitz-Block „Mensch ärgere dich nicht“, die Zweige in
  `aktivesSpiel` und `playerInGame` sowie `maenRemoveSeat`.
- `lobby.go`: Raumliste zeigte für den alten Stand „Mensch ärgere dich nicht“.
- `umbenennen.go`: Umbenennen im alten Spielstand.
- Tests: Menschaerger-Block in `bot_steuerung_test.go`,
  `TestBotMenschGanzePartien` in `bot_brett2_test.go`; `tabellen_test.go`
  prüft das Spiel jetzt über das neue Feld `Room.Maen` (das trägt die
  Sitz-/Umbenenn-Pflege weiter).

**Unberührt:** `/api/maensync` (Handler `maenSyncHandler`, `Room.Maen`,
`syncArten`, `angebot.go`-Zuordnung), `/api/labyrinth` und alle übrigen Spiele.

**Geprüft:** `gofmt` sauber (außer der lokalen, nicht getrackten
`bot_steuerung_test.go` mit CRLF), `go vet ./...` und `go build` in beiden
Fassungen; volle Suiten `go test -count=1 ./...` in `connector` (221 s) und
`connector-online` (232 s), beide **exit 0**; die Routen-Probe oben lief in
beiden Fassungen grün und wurde danach gelöscht. Beide Ablagen
(`Documents\Homepage` und `Documents\GitHub\Homepage`) sind für diese Dateien
byte-gleich.

## 26.11.2026 · Acht Spiele nach der Grafik-Überarbeitung technisch geprüft

**Auftrag:** Arkanoid, Flappy Bird, Frogger, Pac-Man, Snake, Space Invaders,
Doodle Jump und Pong sollten nach den optischen Überarbeitungen wieder
zuverlässig spielbar sein – ohne die neuen Grafiken anzufassen.

**Vorgehen:** Statt jede Datei nur zu öffnen, wurden Steuerung und Mechanik
gemessen (Playwright/Chromium, echter lokaler Server): je Spiel und Stil
(Klassisch/Mädchen/Buben) auf PC 1280 × 800 und Handy 390 × 844, mit Tastatur,
Mausklick und Wischen, dazu Spielstart, Neustart und Spielende. Die Spiele
haben keine Test-Globals (alles in IIFEs); deshalb saßen für die Messung
kurzzeitig Zustandshaken in den acht Dateien und wurden danach entfernt –
alle acht sind wieder hakenfrei (`window.__t` gibt es nirgends), die
Rückkehr wurde per md5 gegen die Sicherungen geprüft.

**Drei echte Fehler gefunden und behoben:**

1. **Arkanoid: Feld zu hoch (der gemeldete Fehler).** Die Leinwand ist
   480 × 520, stand aber mit `width:100%` in einer 760 px breiten Spalte:
   am PC **742 × 804 px**, Unterkante 898 bei 800 px Fensterhöhe, die Seite
   scrollte 204 px. Neu: `passeFeld()` in `arkanoid.html` begrenzt die
   Feldbreite auf die freie Fensterhöhe (Kopf, Fuß und Regelzeile werden
   abgezogen); die Bühne bekommt genau die Feldbreite, damit der Startschleier
   deckungsgleich bleibt; neu gerechnet bei `resize`, Drehung und
   `visualViewport`. Gemessen: 1280 × 800 → **551 × 596**, Seite genau 800;
   1024 × 768 → 521 × 564 / 768; 1920 × 1080 → 742 × 804 / 1080;
   Handy 390 × 844 unverändert 360 × 390 / 844. Zeigerzuordnung geprüft:
   Klick bei 15 / 50 / 85 % der Feldbreite setzt den Schläger auf
   72 / 240 / 408 von 480 – links ist links.
2. **Space Invaders am Handy: 8 px zu lange Seite.** `body` hatte 18 px
   Unterrand, die Seite war 852 statt 844 px hoch. Jetzt 2 px unter 430 px
   Breite: Seite 844 px, kein Quer- oder Hochscrollen, Feld 374 × 442.
3. **Frogger: erste Pfeiltaste nach „Starten“ wurde geschluckt.** Der
   Tastaturwächter (`bedienelementFokussiert()`) lässt Tasten liegen, solange
   ein Bedienelement den Fokus hat – nach dem Klick auf „Starten“ war das der
   Knopf. `starten()` gibt den Fokus jetzt ab (nur `BUTTON`; ein Auswahlfeld
   behält ihn). Gemessen: nach dem Klick bewegen alle vier Richtungen sofort
   (hoch −32, runter +32, links −32, rechts +32 px).

**Gemessen in Ordnung (keine Änderung nötig):**

- **Flappy Bird:** Flattern hebt (y 204 → 148, vy −2), Rohre, Tod und
  „Nochmal“ (danach wieder `laeuft`, y 202) – in allen drei Stilen.
- **Frogger:** vier Richtungen, Wischen, echte Vollbild-API (an: Feld
  553 → 565 px, Knopf „Vollbild beenden“, Escape aus, danach bewegt sich der
  Frosch weiter), alle drei Stile.
- **Pac-Man:** links/rechts/hoch aus je einem frischen Spiel (−51 / +51 / −51
  px). „Runter“ ist am Startfeld **absichtlich** durch die Wand blockiert
  (Reihe 20 ist `###`), nach dem Losfahren greift es.
- **Snake:** hoch → links → runter → rechts, ohne Umkehrfehler,
  Einzelspieler läuft.
- **Space Invaders:** Schuss existiert (alle 10 ms gemessen: 1, 1, 1, 0 – er
  trifft zuerst den eigenen Bunker), nach dem Durchschießen fliegt er weiter,
  Punkte 0 → 90/105, Gegner werden zerstört.
- **Doodle Jump:** Sprung −10,8 → +10,8 (Sprunghöhe 146 px bei 0,40
  Schwerkraft, größter Plattformabstand 98 px), Seitensteuerung
  200 → 132 → 216, Gegner exakt 24 × 18 px in allen drei Stilen,
  Trefferellipse 23,4 × 20,7 px (Treffer bis 23 px Abstand, bei 26 px keiner)
  – also nie größer als Gegner plus Spielfigur, kein unfairer Treffer.
- **Pong:** Schläger links/rechts, Ball prallt, Solo läuft.
- **Neustart:** `#btnNeu` in allen acht Spielen führt zurück in den
  Startzustand (Punkte 0). Arkanoid bis zum Ende gespielt: 3 → 2 → 1 → 0
  Leben, „Nochmal“ gibt 3 Leben und 0 Punkte.
- **Rahmen (Connector-Start):** Spiel im iframe geöffnet und dort gestartet –
  die Tastatur kommt im Spiel an (Schläger 240 → 42), sobald ein Klick im
  Rahmen war; Klick und Wischen funktionieren dort genauso. Der laufende
  Go-Connector wurde dafür nicht gestartet (iframe-Nachbau im Browser).

**Offen:** Kein Spiel wurde in einem durchgehenden Durchlauf bis zum Sieg
gespielt (Ausnahme Arkanoid bis Game over); Power-ups bei Arkanoid und
Invaders wurden nur indirekt über Punkte geprüft. Doodle-Jump-Gegner
erscheinen erst ab 150 m und dann mit 0,07 % je Plattform – gemessen wurde
deshalb mit kurzzeitig gesetztem Abstand, nicht im freien Spiel.

## 26.11.2026 · Live-Stand zusammengeführt: Mensch ärgere dich nicht am Standweg

Vor dem Live-Test im Mehrspieler lief die Go-Suite über den unfertigen Stand.
Sie fand zwei Lücken zwischen Seite und Server:

- **Mensch ärgere dich nicht:** Die Seite sprach schon den Standweg
  (`TG_ONLINE_SPIEL = "maensync"`, `online-sync.js`), die Verwaltung sagte
  `"api": "maensync"`, aber der Server führte das Spiel weiter unter
  `"menschaerger"`: `spielApi` (Runde anbieten, bereit, Anwesenheit), die
  Tischart (`tischspiel`) und die Sitzpflege in `leave.go` kannten nur den alten
  Weg. Jetzt sind beide Seiten auf `maensync`: `spielApi` und `tischspiel` zeigen
  darauf, `angebot.go` liest den Status aus `room.Maen`, `leave.go` räumt Sitze
  und „laufende Partie“ auf, `umbenennen.go` benennt mit um, und der
  Register-Eintrag steht auf `"art": "tisch"` wie das Labyrinth. Damit läuft
  „bieten → beitreten → bereit → Start → Zug“ im Test durch; der `create`-Aufruf
  beim Anbieten legt den Stand an, der Beitritt nimmt den Sitz.
- **Schiffe:** Das Spiel und der Server setzen seit dem Umbau die Flotte
  5, 4, 3, 2, 1 (Flugzeugträger, Zerstörer, Patrouillenboot, U-Boot, Ente); der
  Zug-Test schickte noch die alte Flotte 5, 4, 3, 3, 2. Der Test folgt jetzt der
  Flotte aus `schiffe.go`; die nicht mehr benutzten Schiffsbilder
  (`ansicht_schlachtschiff.png`, `bilder/kleines-boot.png` u. a.) sind entfallen,
  die Seite zeichnet die Schiffe selbst.

Geprüft: `go vet`, `go build` und die gezielten Läufe
(`TestSpieleverwaltungPasstZuGo`, `TestSpieltabellenVollstaendig`,
`TestJedesAngeboteneSpielZuZweit`, `TestJedesSpielIstImRaumSichtbar`,
`TestZehnTausendHTTP`) sind grün. Im Gesamtlauf bleiben zwei bekannte
Wackelkandidaten (`TestZehnTausendHTTP`, `TestBotDomino` – Zufall bzw. geteilter
Raumzustand); einzeln laufen sie durch.

### Live-Kontrolle (nach dem Push, 26.11.2026)

Gegen `https://trashygecko.onrender.com` geprüft, nachdem Render das Abbild neu
gebaut hatte:

- `/api/maensync` antwortet (vorher „404 page not found“, jetzt die Meldung des
  Standwegs) – der neue Connector läuft.
- Homepage `200`, `/connector` `200` ohne Konsolenfehler,
  `spiele/goldminer/goldminer.html` mit `window.GoldMiner` 2.0.0, Startbildschirm
  und den drei Stilknöpfen; in der Seite umgeschaltet: `normal`, `maedchen`,
  `buben` (Zustand meldet jeweils den Stil), `start()` läuft (Level 1, Phase
  „schwingt“), 0 Fehler.
- Register live: `"api": "maensync", "art": "tisch"`.
- Mehrspieler mit zwei getrennten Browsersitzungen im Raum `TG-HALLE`:
  „bieten“ und „annehmen“ je `204`, der Stand steht danach mit zwei Sitzen auf
  `invited`; beide „bereit“ und „Start“ `200` → `active`; ein Zug von A `200`,
  B sieht denselben Stand mit beiden Spielern. (Die Prüfpartie hing danach im
  Speicher, weil die Testsitzung geschlossen war; der Neuaufbau durch diesen
  Commit räumt sie weg.)

Veröffentlicht mit dem Commit dieser Änderung auf `origin/main`; Render baut
daraus das Docker-Abbild (Website aus dem Repo, `connector.html` aus
`connector-online`).

---

## 26.11.2026 · Goldminer auf Schnittstelle 2.0 – Grafik unangetastet

`spiele/goldminer/goldminer.html` (beide Ablagen und die Android-Assets
byte-gleich, md5 `e06b155c…`): Das Spiel hat jetzt den Standard 2.0.0 –
`window.GoldMiner` mit `version`, `bereit`, `whenReady()`, `start()`,
`einstellungen()`, `zustand()`, `startbildschirm()`, `stop()`, `pause()`,
`resume()`, `aktion()` und den Ereignissen
`goldminer:bereit|ende|pausiert|weiter`.

- **Grafik unverändert:** Himmel, Erde, Vordergrund, Winde und Fundstücke sind
  Canvas-Funktionen in der Datei (kein CDN, keine Bilddateien). Der Vergleich
  mit der Sicherung `sicherung/goldminer-2.0-2026-11-26/` ergibt für den
  gesamten Zeichenblock (17.648 Zeichen) **0 Unterschiede**; `LEVEL` und
  `ARTEN` sind ebenfalls gleich. Mädchen- und Bubenansicht sind damit
  nachweislich nicht angefasst (Pixelprobe der Seitenfläche und des Himmels
  vor/nach dem Umbau gleich: Mädchen `247,228,236` / `243,236,248`, Buben
  `10,10,11` / `4,4,5`).
- **Regeln unverändert:** Der Regelblock unterscheidet sich nur um drei
  Stellen – Pausen-/Startbildschirmsperre in `auswerfen()`, `tick++` und zwei
  `meldeEnde()`-Aufrufe. Physik, Tempo, Werte, Zielrechnung, Zeit und
  Levelaufbau sind gleich.
- **Startbildschirm:** eigener Bereich (`spiel-start.css`/`-js`) mit Titel,
  echter Minenvorschau (dieselben Ansichtsfunktionen, `#startSzene`),
  Steuerung, Infos und „Graben beginnen“. Gemessen 0 px innerer Überlauf und
  STARTEN im Bild bei 360×640, 390×844, 844×390, 1024×600 und 1280×800, Seite
  jeweils 0 px Quer- und Hochscrollen; auf Höhen ≤ 680 entfällt die Vorschau.
- **Stil:** `start({stil})` stellt nur die Welt ein (Connector beim Öffnen und
  Wechseln), `start()`/`start({level})` beginnt. Direkt wählt der Spieler über
  `#stilKnoepfe` (gemerkt in `goldminerStil`); kommt der Stil von außen, ist
  die Wahl verborgen. `buben` zeichnet weiter die Ansicht `jungs`, `normal`
  die Ansicht `klassisch`.
- **Geprüft (Playwright, Chromium, 1280×800 und 390×844):** 16 Fehlerfälle mit
  den Codes aus Abschnitt 3.2, Zustand nach jedem Fehler unverändert; Kopien
  von `zustand()` und `einstellungen()` unabhängig; `:bereit` genau einmal;
  `:pausiert`/`:weiter` je einmal pro Wechsel (zweites `pause()` löst nichts
  aus); `:ende` mit vorgespulter Uhr: `{grund:"verloren", punkte:0, level:1,
  levelAnzahl:50, tick:0}` samt Ergebniskarte; Leertaste wirft (Takt 0→1),
  `P` pausiert, `Escape` setzt fort, Klick auf die Leinwand wirft (1→2);
  Fang: 17 → 16 Fundstücke, +50 Geld. 0 Konsolen- und Seitenfehler.
- **Connector (echt, Port 8779):** `GoldMiner` steht jetzt in
  `SPIEL_API_NAMEN`. Im Rahmen kam „Mädchen“ an (`stilWahl false`, Kachel
  verborgen), `start()` lief; Umschalten auf „Buben“ wechselte die Welt bei
  **laufender** Partie (Level 1, Takt 0 unverändert). 0 Fehler.
- **Speicher:** neu `goldminerStil`; `tg-stil` wird weiter gelesen, aber nicht
  geschrieben. Der Prüf-Global `window.goldMinerStand` ist entfallen (jetzt
  `GoldMiner.zustand()`).

---

## 26.11.2026 · Super Pimmelchen: Pause/Weiter nachgesehen und den Takt an eine Stelle gezogen

`spiele/pimmel/spimmel.html` und die daraus gebaute
`Super-Pimmelchen-61-Level-Einzeldatei.html` (beide Ablagen byte-gleich):
Regeln, Level, Stufen und die Schnittstelle **2.0.0** unverändert; nur der
Code um Pause und Fortsetzen ist aufgeräumt.

- **Die Abfrage stand richtig:** `resumeGame()` prüft im Zweig
  `state === 'playing'` nach dem Neustart von Uhr und Special-Takt, ob eine
  Richtung gehalten wird – genau dort, wo auch vorher der Takt gesetzt wurde.
  Am Verhalten ändert sich nichts: gehaltene Taste läuft weiter (tick 1 → 3),
  ein unter dem Pausenblatt losgelassener Finger nicht mehr (2 → 2), ohne
  gehaltene Richtung keine Eigenbewegung (1 → 1).
- **Überflüssig war das `!moveTimer`:** In der Pause kann kein Takt entstehen
  (`setDirection` und `schrittFuerRichtung` steigen vorher aus), der Zusatz
  prüfte also etwas, das nie eintritt. Dasselbe `clearInterval`+`setInterval`
  stand an vier Stellen (`setDirection`, `resumeGame`, `activateSpeed` zwei
  Mal) – jetzt einmal in `starteMove()` neben `stopMove()`.
- **`heldDirections` bleibt, wo es ist:** Die Zuordnung steht bei den
  Tastaturzuständen, weil `haeltRichtung()` beides liest; `pauseGame()` leert
  sie weiterhin (ein Finger bekommt unter dem Pausenblatt kein `pointerup`).
  `pressedDirs` wird bewusst **nicht** geleert – eine Taste, die über die Pause
  gehalten wird, soll danach weiterlaufen; `blur` räumt beide Wege ab.
- **Kommentar gekürzt:** Der lange Absatz mit der Vorgeschichte („vorher
  startete hier immer ein Takt …“) steht jetzt in zwei Zeilen als Regel; die
  Historie gehört in dieses Dokument.
- **Geprüft:** eigenes Prüfskript Pause/Weiter **13 grün, 0 rot** in beiden
  Fassungen (Taste gehalten, Steuerkreuz losgelassen, nichts gehalten),
  `pimmel_richtung_pruefung.js` **43 grün, 0 rot**,
  `pimmel_2_0_pruefen.js` **119 grün, 0 rot**, Skript-Syntax (`node --check`)
  grün.

## 26.11.2026 · Super Pimmelchen: `aktion('richtung')` macht einen Schritt statt eines Dauerlaufs

`spiele/pimmel/spimmel.html` und die daraus gebaute
`Super-Pimmelchen-61-Level-Einzeldatei.html` (beide Ablagen byte-gleich):
Regeln, Level, Stufen und die Schnittstelle **2.0.0** unverändert.

- **Fehler:** `aktion('richtung', {x,y})` rief `setDirection()`, und das setzt
  neben dem Schritt auch den Bewegungstakt – ein einzelner Aufruf startete
  damit denselben Dauerlauf, der beim Fortsetzen (`resume`) schon entfernt
  wurde. Die Doku verspricht „ein Schritt“.
- **Jetzt:** neue Funktion `schrittFuerRichtung()` – dieselben Prüfungen wie
  Tastatur und Steuerkreuz (nur im Spiel, nicht in der Pause, keine
  Kehrtwende bei Länge > 1), aber ohne Takt. Ein Aufruf dreht und geht genau
  einen Takt; die Antwort lautet `{ok: true, schritt: true|false}`, wobei
  `false` eine gesperrte Richtung meldet (Wand/Brücke) – dort bleibt auch die
  Taste stehen.
- **Gemessen** (Level 1, mittel, beide Fassungen): rechts und links je
  **tick +1** und ein Feld weiter, danach 900 ms ohne weitere Bewegung
  (vorher wären es ~4 Schritte gewesen); hoch und runter **tick +0** an der
  Startposition – genau wie ein kurzer Tastendruck dort; gedrückte Taste läuft
  weiter (tick 1 → 5) und hält nach dem Loslassen an (5 → 5).
- **Geprüft:** neues Prüfskript **43 grün, 0 rot** über beide Fassungen
  (vier Richtungen gegen die Tastatur, kein Dauerlauf, Sperre in der Pause,
  Fortsetzen ohne Eigenbewegung, keine Konsolenfehler).

## 26.11.2026 · Super Pimmelchen: Startbildschirm aufgeräumt

`spiele/pimmel/spimmel.html` und die daraus gebaute
`Super-Pimmelchen-61-Level-Einzeldatei.html` (beide Ablagen byte-gleich,
`05647b1cee9126156a51ad78d94cd0dd` / `d4465f7e3979896f1e73c67866ce3cbf`):
Regeln, die 61 Level, Stufen, Punkte, Highscore und die Connector-Schnittstelle
**2.0.0** sind unverändert – geändert ist nur der Startbildschirm.

- **Weg:** der Hinweis „Gilt beim nächsten Start ohne Connector.“ unter der
  Stilwahl und der ganze Platzhalterbereich „EINSTELLUNGEN“ (Leben 3, Zeit je
  Leben 120 s, Feld 20 × 11, Specials, Triebwerk). Dazu die zugehörigen
  CSS-Regeln (Liste, Farben in Mädchen/Buben, Querformat-Sonderfälle) und der
  tote Verweis auf `#start-takt` im Skript. Die drei Grafikstile Normal,
  Mädchen und Buben bleiben samt Wahl.
- **Neu ausgerichtet:** Im Querformat (zwei Spalten) deckt die Stilwahl jetzt
  beide Zeilen der zweiten Spalte ab und trägt ihre Überschrift wieder – vorher
  saß dort der Einstellungsblock, darunter blieb eine Lücke von **24 px**.
  Gemessen bei 640 × 360 und 844 × 390: Stilblock **169 px** hoch über beide
  Zeilen, Unterkante bündig (Lücke 0 px), kein Scrollen.
- **Kürzer:** Der Startbildschirm passt bei 360 × 640 jetzt in ein Fenster
  (Dokument **640 px** statt vorher 738 px); 390 × 844, 640 × 360, 844 × 390
  und 1280 × 800 bleiben ohne Hoch- und Querscrollen.
- **Geprüft:** eigenes Prüfskript in beiden Fassungen **39 grün, 0 rot**
  (Texte weg, drei Stile vorhanden und umschaltbar, Spiel startet, fünf
  Fenstergrößen ohne Scrollen, Spalte im Querformat geschlossen, Stil vom
  Connector blendet die Wahl in jeder Größe aus, keine 404 und keine
  Konsolenfehler); `pimmel_2_0_pruefen.js` **119 grün, 0 rot**;
  `pimmel_datei_pruefen.js` lädt die Einzeldatei über `file://` fehlerfrei.
  Die vier verbleibenden Meldungen in `pimmel_einzel_pruefen.js`
  (`specialsTaktMs`, Highscore) treten mit dem Stand **vor** dem Aufräumen
  genauso auf – ältere Prüferwartung, kein Rückschritt.

## 26.11.2026 · Tic Tac Toe: Spielfeldsymbole je Grafikstil (SVG)

`spiele/brettspiele/tictac/tictactoe.html` (beide Ablagen byte-gleich,
`34196dad3be54508fad6a1fe273d513d`): Regeln, Gewinnprüfung, Bot, Online-Synchronisierung und
Schnittstelle unverändert; Sicherung davor unter
`sicherung/tictac-symbole-2026-11-26/tictactoe-vor-symbolen.html`
(`28cd70eda2e5ba119364db952b04ac81`).

- **Paare je Stil:** `normal` Kreuz (X) und Kreis (O) wie bisher, `maedchen`
  Herz (X) und Teddy (O), `buben` Pentagramm (X, durchgezogene Linien) und
  Totenkopf (O) – alle als gezeichnetes SVG, keine Schriftzeichen, keine
  Emojis.
- **Gleiche Maße:** jedes Symbol sitzt in `viewBox="0 0 100 100"` und füllt
  62 % der Zelle, mittig; gemessen 103 px in 167-px-Zellen (1280 × 800),
  66 px in 110 px (390 × 844), 45 px in 77 px (640 × 360). Kein Symbol
  berührt den Zellrand, X und O sind gleich groß.
- **Wechsel:** `stilSetzen()` zeichnet neu, also tauscht auch der
  Connector-Stil (`start({stil})`) mitten in der Partie die Symbole, ohne
  Brett oder Züge anzufassen (geprüft: Brett „XO“, Züge 2 unverändert,
  alle drei Fassungen nacheinander).
- **Online:** jeder Rahmen zeichnet seine eigene Fassung – geprüft mit zwei
  Sitzungen: A Mädchen (Herz/Teddy), B Buben (Pentagramm/Totenkopf), gleiches
  Brett, keine Vermischung.
- **Texte:** Statuszeile und Ergebniskarte nennen das Symbol des eigenen Stils
  („Herz ist dran.“, „Totenkopf hat drei in einer Reihe.“), die Startzeile bei
  Mädchen heißt „Herz und Teddy, ganz lieb – drei in einer Reihe.“, das
  `aria-label` der Felder ebenso („Feld 5: Teddy“).
- **Geprüft:** `pruefung_symbole.js` **57** grün (drei Stile, Größe und Mitte,
  Wechsel mitten in der Partie, Sieg mit Linie und Ergebniskarte, Remis,
  neue Runde, Bot, 1280 × 800 / 390 × 844 / 640 × 360), `pruefung_start.js`
  **73**, `pruefung_solo.js` **78**, `pruefung_multi.js` **35** und
  `pruefung_umgebung.js` **16** grün (vier neue Symbolprüfungen über den
  echten Connector).

- **Nachtrag 320 px und Kontrast (26.11.2026):** Bei **320 px** Breite
  (320 × 568 und 320 × 640) schob das Brett die Seite um 25 px auseinander
  (Brett 326 px im 320-px-Fenster, rechte Spalte angeschnitten). Ursache war
  das Seitenraster: ohne `grid-template-columns` bestimmte das größte Kind die
  Spaltenbreite. Mit `grid-template-columns:minmax(0,1fr)` misst die Seite
  jetzt 320 × 568 bzw. 320 × 640 **ohne Scrollen**, das Brett **274–280 px**,
  Zellen 87–93 px, Symbole **51–57 px** (62 %), Rand 17–18 px; die drei
  Leistenknöpfe sind 34 px hoch. Alle sechs Symbole sitzen mittig in der
  Zelle, kein Randkontakt.
- **Kontrast an gerenderten Zellbildern gemessen** (Zelle 2× und 3×
  aufgenommen, Hintergrund je Bildzeile aus den Seitenrändern, Symbolfläche
  gegen diesen Grund, WCAG-Verhältnis): Kreuz **4,35 : 1**, Kreis
  **7,82 : 1**, Herz **4,66 : 1**, Teddy **4,90 : 1**, Pentagramm
  **4,56 : 1**, Totenkopf **12,47 : 1** – bei 320 px und 1280 px gleich
  (Abweichung ≤ 0,04). Zwei Farben wurden dafür nachgezogen: Herz `#e8407e`
  → `#de2f6c` (3,9 → 4,7 : 1) und Teddys Augen/Nase/Mund `#3b2216` →
  `#241209` (2,3 → 4,1 : 1 auf dem Fell). Innere Details: Augen des Totenkopfs
  8,08 : 1, Kreuzungen des Pentagramms 3,3–3,6 : 1, Teddy-Kontur 2,29 : 1
  (nur eine Linie). Werkzeuge: `arbeit/tmp/tictac/symbole_320.js`
  (30 Prüfungen, 18 Zellbilder) und `kontrast_messen.py` (Schwelle 3 : 1 für
  die Hauptfläche, alle grün).
- **Nach dem Eingriff erneut geprüft:** `pruefung_symbole.js` **57**,
  `pruefung_start.js` **73**, `pruefung_solo.js` **78**, `mess_layout.js`
  **24/24**, `pruefung_multi.js` **35** und `pruefung_umgebung.js` **16**
  grün; 1280 × 800 unverändert (Zellen 167 px, Brett 502 px).
- **Offen bei 320 px:** Die Knöpfe der Kopfzeile („Vollbild“, „Zurück zum
  Connector“) überlappen den Titel „TIC TAC TOE“. Nicht angefasst – betrifft
  die Spielfeldsymbole nicht.

## 26.11.2026 · Tic Tac Toe: Startbildschirm auf zwei Modi vereinfacht

`spiele/brettspiele/tictac/tictactoe.html` (beide Ablagen byte-gleich,
`28cd70eda2e5ba119364db952b04ac81`): Regeln, Gewinnprüfung, Bot und
Online-Weg unverändert; Sicherung des Stands davor unter
`sicherung/tictac-startbildschirm-2026-11-26/tictactoe-vor-vereinfachung.html`
(`01e1c3fc24180bb429d1850aba741aec`). Der Abschnitt darunter beschreibt die
Moduswahl davor.

- **Zwei Modi** statt drei: „Zwei am Gerät“ und „Gegen den Bot“. Entfernt sind
  die Schaltfläche „Online zu zweit im Connector“, ihre Erklärung, der Block
  „Deine Wahl“ (Modus/Feld/Bilanz samt Tipp) und der Hinweis unter dem
  Startknopf zur Stilwahl – dazu die zugehörigen CSS-Regeln und die toten
  JS-Zweige für einen gewählten Modus `online`; `einstellungen().modi` nennt
  jetzt `["lokal","bot"]`.
- **Online unverändert:** `start({online:{…}})` des Connectors setzt weiter
  Marke, Gegner und Stil, zeigt den Wartetext ohne Wahl und blendet den
  Starten-Knopf aus; die Partie startet wie bisher, sobald die eigene Marke
  kommt. Geprüft mit zwei Sitzungen über Port 8778.
- **Stil und Bot bleiben:** drei Stile beim direkten Start (gespeichert in
  `ticTacToeStil`), im Connector-Rahmen vom Rahmen; drei Bot-Stufen wie zuvor.
  Die Bilanz wird weiter gebucht (`ticTacToeBilanz`), nur nicht mehr im
  Startbildschirm angezeigt.
- **Layout (24 Messungen: 1280 × 800, 640 × 360, 390 × 844, 360 × 640, je drei
  Stile und zwei Modi):** überall ohne Hoch- und Querscrollen, Startknopf im
  Bild; Startbildschirm bei 390 × 844 **796 px** und damit genau ein Fenster,
  bei 640 × 360 **327 px** (Knopf unten 311/360), bei 360 × 640 592 px,
  bei 1280 × 800 Knopf unten 592/800. Im Querformat stehen die drei Wahlzeilen
  einspaltig untereinander, damit im Modus „Zwei am Gerät“ keine leere Spalte
  bleibt.
- **Geprüft:** `pruefung_start.js` **73** grün (neu geprüft: keine
  Online-Wahl, kein Block „Deine Wahl“, kein Tipp, keine Fußnote im DOM; der
  Connector-Start übernimmt Stil und Marke und zeigt nur den Wartetext),
  `pruefung_solo.js` **78**, `pruefung_multi.js` **35** und
  `pruefung_umgebung.js` **12** grün.

## 26.11.2026 · Tic Tac Toe: Startbildschirm, drei Spielmodi, Bot in drei Stufen

`spiele/brettspiele/tictac/tictactoe.html` (beide Ablagen byte-gleich,
`01e1c3fc24180bb429d1850aba741aec`): Regeln, Gewinnprüfung, Online-Weg und
Schnittstelle **2.0.0** unverändert; Sicherung des Ausgangsstands liegt unter
`sicherung/tictac-start-2026-11-26/tictactoe-vor-start.html`
(`840a6c65a2b65978d29a2ed0ea0ef7a2`).

- **Drei Modi** auf einem neuen Startbildschirm: „Zwei Spieler an einem Gerät“
  (`lokal`), „Gegen den Bot“ (`bot`), „Online über den Connector“ (`online`).
  `start()` ohne Daten startet den gewählten Modus (Vorgabe `lokal`);
  `start({stil, online})` verhält sich wie bisher. Die Stufenleiste erscheint
  nur bei `bot`, die Stilwahl nur beim direkten Start ohne Connector; eine
  Zeile unter den Knöpfen nennt die Wahl in Worten. Kein `alert`/`prompt`.
- **Bot vollständig im Spiel** (kein Go-Code angefasst, `tictactoe.go` und
  `bot_tictactoe.go` unverändert): Der Bot spielt Kreis, der Mensch Kreuz und
  beginnt, und zieht nach 320 ms über dieselbe Prüfung wie ein Mensch.
  `leicht` nimmt einen Sieg in etwa drei von vier Fällen und blockt nie;
  `mittel` spielt Sieg → Block → Mitte → Ecke; `schwer` rechnet vollständiges
  Minimax mit Zufall unter gleich guten Zügen.
- **Gemessen (je 60 Partien, Zufallsspieler als Mensch):** `leicht` blockt
  31 %, nutzt 89 % der Siege, Mensch 25 / Bot 26 / Remis 9; `mittel` und
  `schwer` blocken und siegen 100 % (mittel 1 / 50 / 9, schwer 0 / 47 / 13).
  Jäger-Test (Drohung erzwingt Doppeldrohung): gegen `mittel` gewinnt der
  Mensch **40 von 40**, gegen `schwer` **0 von 40** (alle Remis) – der
  Unterschied ist damit nicht nur eine Zahl, sondern spielbar.
- **Drei Stile** unverändert in den Farben, jetzt auch wählbar: `normal`
  (dunkelgrün/gold), `maedchen` (rosa, Herzen), `buben` (düster, blutig).
  Eigene Wahl beim Direktstart liegt in `ticTacToeStil`; im Connector-Rahmen
  kommt der Stil von außen (`TGStil`/`data-stil` des Rahmens), dann ist die
  Wahl im Spiel verborgen (`stilWahl: false`). X und O bleiben überall
  erkennbar – Kontrast zum Feld: normal 3,99 : 1 / 6,96 : 1,
  Mädchen 5,04 : 1 / 7,81 : 1, Buben 4,80 : 1 / 12,55 : 1.
- **Layout:** 1280 × 800, 640 × 360, 390 × 844 und 360 × 640, je drei Stile und
  zwei Modi – überall ohne Hoch- und Querscrollen, Startknopf und Fußnote im
  Bild; Startbildschirm bei 390 × 844 **844 px** hoch (genau ein Fenster).
- **Geprüft:** `pruefung_solo.js` **78** grün, `pruefung_start.js` **73** grün
  (Startbildschirm, Modus-/Stufenwechsel, drei Stile in beiden Modi,
  Bot-Statistik, Tastatur/Maus/Touch, vier Fenstergrößen, keine
  Konsolenfehler), `pruefung_multi.js` **35** grün und `pruefung_umgebung.js`
  **12** grün über den echten Connector (Port 8778, zwei getrennte Sitzungen:
  Einladung, Annahme, Züge in beide Richtungen, gleiches Brett, A Mädchen /
  B Buben ohne Überschwappen, keine Stil- und Moduswahl im Rahmen).
- **Fallstrick beim Prüfen:** Der Connector merkt sich Räume und Angebote; nach
  mehreren Läufen lagen alte Karten im selben Raum und die Testannahme traf
  ein altes Angebot. Jede Online-Prüfung braucht einen **frisch gestarteten**
  Connector, sonst sind die Ergebnisse nicht eindeutig.
- **Offen:** nicht auf zwei echten Handys geprüft (nur zwei Browser-Sitzungen
  auf einem Rechner); nichts committet.

## 26.11.2026 · Super Pimmelchen: „WEITERSPIELEN“ setzt die Schlange nicht mehr in Bewegung

`spiele/pimmel/spimmel.html` und die daraus gebaute
`Super-Pimmelchen-61-Level-Einzeldatei.html` (beide Ablagen byte-gleich):
Nach einer Pause lief die Schlange ohne jede Eingabe in ihrer letzten Richtung
weiter, bis sie an die Wand stieß.

**Ursache:** `resumeGame()` stellte beim Fortsetzen immer
`moveTimer = setInterval(step, moveDelay)` und damit einen eigenen Takt auf,
obwohl das Spiel sonst nur fährt, solange eine Richtung gehalten wird
(`keydown`/`pointerdown` startet, `keyup`/`pointerup` beendet über
`stopMove()`). Mit der Startrichtung `{x:1,y:0}` wurde daraus „läuft von
selbst nach rechts“; mit einer zuletzt gehaltenen Richtung links war es
genauso (4 statt 5 Takte, weil die Wand näher lag), und oben/unten war der
erste Schritt am Startplatz ohnehin Wand – deshalb fiel der Fehler nur rechts
auf. Kein Sonderfall im Code, sondern ein zusätzlicher Taktgeber.

**Korrektur:** `resumeGame()` startet den Takt nur noch bei gehaltener
Richtung (`haeltRichtung()` prüft `pressedDirs` und `heldDirections`);
`pauseGame()` leert zusätzlich `heldDirections`, damit ein unter dem
Pausenblatt losgelassener Finger nicht als gehalten gilt. `heldDirections`
steht jetzt bei `pressedDirs`. `dir`, `facing`, `step()`, Leveldaten und
Spielregeln sind unverändert – die Rechtsbewegung ist kein Sonderfall.

**Vorher gemessen** (Chromium; Sicherung `sicherung/pimmel-pause-2026-11-26/`
mit den md5 `bc5b1e63…` und `330385ab…`, also exakt der Stand vor der
Korrektur, gegengeprüft): Pause bei Fahrt nach rechts → nach „WEITERSPIELEN“
5 Takte in 1,2 s ohne Eingabe, nach links 4 Takte, oben/unten 0 (Wand am
Startplatz).

**Nachher** (140 Prüfungen grün, `arbeit/tmp/pruefung_pimmel_pause.js`, je
Fassung 70): oben, unten, links, rechts je 0 Takte in 1,4 s nach dem Weiter,
Richtung und Kopf unverändert; fünf Pausen hintereinander 0 Takte; Weiter per
Mausklick, per Tipp, per `P`, nach Fensterverlust und nach Tabwechsel jeweils
0 Takte; danach zählt die erste Taste genau einen Takt in derselben Richtung,
Querrichtungen greifen weiter; 6 Schritte je Sekunde vor und nach der Pause
(kein doppelter Timer). Die bestehenden 69 Prüfungen (Stilwahl/Direktstart)
bleiben grün. Schnittstelle weiter **2.0.0**.

## 26.11.2026 · Super Pimmelchen: Stilwahl beim direkten Start

`spiele/pimmel/spimmel.html` und die daraus gebaute
`Super-Pimmelchen-61-Level-Einzeldatei.html` (beide Ablagen byte-gleich):
Leveldaten, 61 Level, Regeln, Aktionen und Schnittstelle 2.0.0 unverändert.

- **Neue Kachel „GRAFIKSTIL“** auf dem Startbildschirm mit `NORMAL`,
  `MÄDCHEN`, `BUBEN` – nur ohne Connector. Ein Klick stellt die Welt sofort
  um (Startbildschirm und Leinwand ziehen mit) und merkt sie in
  `localStorage` unter `superPimmelchenStil`; der nächste direkte Start öffnet
  mit dieser Welt.
- **Connector hat Vorrang:** `start({ stil })` setzt die Welt und blendet die
  Kachel aus (`#stil-block` verborgen, auch nach `startbildschirm()`); die
  gemerkte Lokalwahl wird dabei weder gelesen noch überschrieben. Eine
  Konfiguration ohne `stil` (z. B. nur `stufe`) lässt die Kachel ebenfalls weg.
- **Feldmaß:** Die Kachel kostet auf dem Startbildschirm eine Zeile. Gemessen:
  844 × 390 → Seite 390 px (kein Scrollen), 740 × 360 und 640 × 360 → 360 px
  (dort weicht die reine Einstellungsliste), 390 × 844 → 844 px; überall kein
  Querscrollen, Startknopf und Stilwahl im Bild.
- **Geprüft** (Chromium, Port 8311): **69 Prüfungen** an beiden Fassungen grün
  – Direktstart mit leerem Speicher (`normal`), Umschalten auf Mädchen und
  Buben (Leinwand 5,18,12 → 248,223,234 → 7,6,6), gespeicherte Wahl nach
  Neuladen, Partie im gewählten Stil, alle drei Connector-Stile über
  `start({ stil })` (Kachel weg, Lokalwahl `buben` unberührt), Handy und
  Querformat ohne Hoch- oder Querscrollen, keine Konsolenfehler. Dazu **16
  Prüfungen über den echten Connector** (Port 8778): Rahmen öffnet Pimmel,
  Stilwechsel klassisch/Mädchen/Buben kommt im Rahmen an, keine Stilwahl im
  Spiel, Lokalwahl unberührt, Partie läuft im Connector-Stil, Stilwechsel
  während der Partie, und ein Direktstart danach zeigt wieder die eigene Wahl.
- **Nebenbei behoben:** Die Initialisierung setzte die Welt, bevor das Feld
  aufgebaut war – bei einer gemerkten Welt lief `draw()` ins Leere und die
  Initialisierung brach ab (Seite blieb auf „nicht bereit“). Jetzt erst
  `resetStage()`, dann `setzeStil()`.

## 26.11.2026 · Neon-Tic-Tac-Toe: Schnittstelle 2.0, drei Welten, Connector-Mehrspieler

`spiele/brettspiele/tictac/tictactoe.html` (beide Ablagen byte-gleich) ist von
der eigenen Abfrage auf die Schnittstelle **2.0.0** umgestellt; Regeln (drei in
einer Reihe, Kreuz beginnt, Remis) unverändert.

- **Kein eigenes Netz mehr:** Das Spiel ruft kein `/api/tictactoe` und liest
  `room`/`id` nicht aus der Adresse. Der Connector schiebt den Stand mit
  `aktion("stand", …)` hinein (1-s-Takt, solange das Brett im Rahmen offen ist)
  und horcht auf `tictactoe:zug` / `tictactoe:neu`; Einladung und Tisch laufen
  über den bestehenden Weg `angebotAktion("bieten")` → `invite`/`accept` am
  Server. Server, `tictactoe.go` und `bot_tictactoe.go` sind unverändert.
- **Drei Welten:** `normal` (Kasino-Grün/Gold, kupfernes X, goldenes O),
  `maedchen` (rosa, weiße Bonbonzellen, Herzen) und `buben` (fast schwarz,
  blutrote Fugen, Tropfen unter dem Rahmen). Wahl ausschließlich im Connector;
  keine Stil-Leiste im Spiel, `start({stil})` richtet nur die Welt ein.
- **Startbildschirm, Ergebnis, Pause** als Karten über `spiel-start.css` und
  `spiel-start.js`; kein `alert`/`prompt`/`confirm`. Ergebnis nennt Sieger und
  Züge, „Neue Runde“ im Online-Spiel fragt die Runde beim Connector an.
- **Feldmaß:** `passeBrett()` rechnet Rahmenbreite, Fensterhöhe und Fußleiste
  ein; Brett 502 px (1280 × 800), 350 px (390 × 844), 320 px (360 × 640),
  281 px (844 × 390) – überall quadratisch, ohne Scrollen.
- **Geprüft:** 78 Einzelprüfungen ohne Connector grün (Schnittstelle,
  Fehlercodes, Ereignisse, Sieg/Remis, Tastatur/Klick, drei Stile, vier
  Fenstergrößen, keine Konsolenfehler) und **35 Prüfungen mit zwei getrennten
  Browsersitzungen** über den echten Connector (Port 8778, Raum TG-HALLE):
  Einladung annehmen, abwechselnde Züge beidseitig synchron, fremder und
  doppelter Zug abgewiesen (Spiel `aktion-nicht-erlaubt`, Server 409), Sieg →
  „Sieg“/„Niederlage“, Remis → „Unentschieden“ auf beiden Seiten, neue Runde,
  Neuladen mitten in der Partie und nach dem Ende (Brett und Ergebnis wieder
  da), eigener Stil je Spieler, keine Konsolenfehler.
- **Offen:** nicht auf zwei echten Handys geprüft (nur zwei Browser-Sitzungen
  auf einem Rechner).
- **Nachtrag (26.11.2026):** Der Stil ist nicht mehr nur im Connector wählbar –
  beim direkten Start gibt es eine eigene Stilwahl, dazu drei Spielmodi und
  einen Bot mit drei Stufen (Abschnitt „Tic Tac Toe: Startbildschirm, drei
  Spielmodi, Bot in drei Stufen“).

## 26.11.2026 · Super Pimmelchen: drei Spielwelten, Highscore-Karte, Feldmaß

`spiele/pimmel/spimmel.html` und die daraus gebaute
`Super-Pimmelchen-61-Level-Einzeldatei.html` (beide Ablagen byte-gleich):
Leveldaten, 61 Level, Stufen und Spielregeln unverändert.

- **Drei Spielwelten** statt reiner Farbvarianten: `normal` (dunkles
  Grün, Gold, Marmor, Zierlinien – der grün-goldene Kasino-Look),
  `maedchen` (Rosa, Herzen, Bögen, weiche Formen) und `buben` (Schwarz,
  Blut, Risse, harte Kanten). Geändert sind
  Boden- und Wandmuster, Brücken, Figurenleuchten, Rahmen, Anzeigen, Knöpfe,
  Startbildschirm, Pause, Spielende und Deko – pro Welt im CSS und auf der
  Leinwand. Keine neuen Bilddateien: die Verzierungen werden gezeichnet, die
  Einzeldatei bleibt ohne Asset-Ordner lauffähig.
- **Stil von außen:** `start({ stil })`; unbekannt oder fehlend heißt `normal`.
  Nur `stil` in der Konfiguration richtet die Welt ein, ohne eine Partie zu
  beginnen – der Connector übergibt so die Wahl aus `stil.js` beim Öffnen und
  bei jedem Wechsel. Das Spiel liest `tg-stil` nicht mehr und hat keine eigene
  Stilwahl; `stil.js` bleibt bei seinen drei Werten.
- **Highscore ohne `prompt`/`alert`:** Karte in der Seite mit Rangliste,
  Namensfeld, „Speichern“ und „Ohne Namen“, per Tastatur (Escape) und Finger
  bedienbar; sie liegt fest über der Seite und bleibt auch nach dem Spielende
  sichtbar, wenn das Spielfeld ausgeblendet ist.
- **Feldmaß:** `passeBuehne()` misst nach und zieht das Feld nach, bis es
  vollständig im Bild liegt; bei 844 × 390 ragt nichts mehr über den Rand.
- **Geprüft** (Chromium, Prüfkopien in `arbeit/tmp/stil/pruefung`,
  131 Prüfungen grün): drei Welten je Fassung (Startbildschirm, laufendes
  Spiel, Pause/Weiter, Highscore, Spielende mit Namenseingabe), 844 × 390
  ohne Überlauf, Touch-Bedienung sichtbar, Connector-Übergabe über
  `start({ stil })` samt Rückfall auf `normal`, Einzeldatei ohne jede
  Fremdanfrage. Die Welten unterscheiden sich messbar in Schriftfarbe,
  Rahmen, Radius und Knopfverlauf sowie im Leinwandbild; angesehen wurden sie
  nicht am Bildschirm, nur gemessen.

- **Korrektur (26.11.2026):** TrashyGecko hat genau drei Spielstile. Der
  zwischenzeitlich eingeführte vierte Stil ist wieder entfernt – aus dem
  Spiel, aus `stil.js`, aus der Connector-Zuordnung, aus Doku und Prüfungen.
  Sein edler Grün-Gold-Look ist jetzt der Stil `normal`; `maedchen` und
  `buben` sind unverändert. Gesichert vorher unter
  `sicherung/pimmel-drei-stile-2026-11-26/`.

- **Visuelle Kontrolle (26.11.2026):** beide Fassungen in Chromium bei 1280 × 800,
  844 × 390, 390 × 844 und 1600 × 900 durchgesehen (Start, Spielfeld, Pause,
  Highscore-Karte samt Formular, Spielende). Behoben: im Mädchen-Stil waren
  Einstellungs-Labels, Werte, Hinweis und „Letzte Runde“ in den dunklen
  Grundfarben auf fast weißem Papier (1,1:1) — jetzt Rosa-Töne; gewählte
  Schwierigkeit dunkler für weiße Schrift; Titel bricht wie in den anderen
  Welten. Hochformat: Anzeige-Labels stießen zusammen und der Connector-Knopf
  lag darüber — Knopf jetzt eigene Zeile, Labels enger. Der Highscore-Schleier
  war neutral und färbte Mädchen grau — jetzt je Welt getönt; Buben-Karte
  hebt sich mit roter Kante ab, Ränge und Formular-Label sind heller. Buben:
  Wände und Risse eine Stufe heller, der Irrgarten ist wieder lesbar; das
  Feld passt in jeder Größe ohne Seitenlauf, auch mit geladenen Schriften.
  Keine Leveldaten, keine Stilarchitektur geändert; Einzeldatei neu gebaut.

## 26.11.2026 · Frogger: Vollbild, Feldmaß und eigener Startbildschirm

`spiele/frogger/frogger.html` neu aufgebaut; Spielregeln, Leveldaten, Tempi,
Zeit, Punkte und Speicherschlüssel (`frogger-rekord`, `frogger-erreicht`)
unverändert. Neu daneben: `spiele/spiel-start.css` und `spiele/spiel-start.js`
als gemeinsames Gerüst für weitere Spiele (siehe
[`SPIELE-SCHNITTSTELLE.md`](SPIELE-SCHNITTSTELLE.md), Abschnitt 9).

- **Vollbild** lag vorher gar nicht vor. Jetzt übernimmt `TGSpiel.vollbild`
  echtes Fullscreen auf `#app`, auf iPhone/Safari ersatzweise ein randloses
  Pseudo-Vollbild mit derselben Bedienung. Der Knopf zeigt Zustand
  (`aria-pressed`, Beschriftung „Vollbild beenden“, Symbol ⤡), Escape beendet
  den Modus, ein zweiter Klick ebenso. Nach dem Aussteigen wird die Darstellung
  über `fullscreenchange` neu berechnet und die Fußzeile wieder eingeblendet.
- **Feldmaß:** Das Feld war 744 px hoch (quadratisch, aber höher als der
  Bildschirm) und schob die Bedienleisten aus dem Bild. `TGSpiel.feldPassen`
  misst jetzt den freien Platz der Bühne (ResizeObserver, `resize`,
  `orientationchange`, `visualViewport`, Vollbildwechsel) und setzt nur die
  CSS-Größe; gezeichnet wird weiter in 416×416, scharf über einen
  DPR-Backing-Store (`ctx.setTransform`). Gemessen: 553 px bei 1280×800,
  832 px bei 1920×1080, 565 px im Vollbild, 344 px bei 390×844, 314 px bei
  360×640, 410 px bei 1024×600, 200 px bei 844×390 – überall quadratisch,
  vollständig sichtbar, ohne Seiten-Scrollen.
- **Startbildschirm** ist kein Overlay über dem Feld mehr: Er besetzt die
  Fläche zwischen Kopfleiste und Menü (Rasterbereich `start`), HUD, Feld und
  D-Pad sind so lange ausgeblendet. Inhalt: großer Titel, SVG-Szene aus Straße,
  Fluss und Frosch, Kurzbeschreibung, Steuerungshilfen, sechs Spielinfos
  (Level, Leben, Zeit, Buchten, Levelanzahl, Rekord), hervorgehobener
  STARTEN-Knopf, Hinweis zur Levelwahl. Passt ohne internes Scrollen: 693/693
  bei 390×844, 489/489 bei 360×640, 282/282 bei 844×390, 492/492 bei
  1024×600, 895/895 bei 768×1024.
- **Gestaltung:** Arcade-Kopfleiste mit LED und Vollbild-Knopf, HUD-Chips mit
  Ziffern in Tabellenbreite, Rahmen mit Schrauben und Innenlicht, Menüleiste,
  D-Pad, Raster- und Vignettenebene über der Seite, Pause-/Level-/Game-over-
  Karten im selben Blattstil. Alle drei Grafstile (klassisch, maedchen, buben)
  mit eigenen neuen Tokens (`--akzent`, `--hud-*`, `--knopf-*`, `--rahmen-a/b`,
  `--gefahr`, `--glanz`).
- **Einbindung unverändert:** Katalogeintrag in `spiel-daten`/`spieldaten.js`,
  Öffnen über Connector-Rahmen oder eigenen Tab, Rückweg über
  `zurueck.js` (`data-ziel=".kopf" data-vor=".kopf-rechts"`) mit
  `postMessage {tg:'spiel', art:'verlassen'}`. Keine zweite Einbindungslogik,
  keine Änderung an `multiplayer-register.js` oder `halle-online.js`.
- **Tastatur:** Pfeile/WASD, P für Pause, Leertaste/Enter nur ohne
  fokussiertes Bedienelement; Escape schließt zuerst Regelblatt oder Pause und
  beendet erst danach (bzw. im Vollbild zuerst) den Vollbildmodus. Das
  Regelblatt merkt sich, ob es die Partie angehalten hat, und setzt nur dann
  fort.
- **Geprüft** (Chromium, 118 Prüfungen grün, `arbeit/tmp/frogger_pruefen.js`):
  Startbildschirm und Spielstart, Zeit läuft, Pfeiltaste bewegt den Frosch,
  Pause über P, Weiter, Escape schließt Pause und Regelblatt, Level neu,
  Vollbild an/aus per Knopf und per Escape inklusive Knopfzustand und
  Wiederherstellung, Pseudo-Vollbild auf dem iPhone-Pfad, sechs Bildschirm-
  größen (Start und Spiel), Game over über eine Testkopie mit 2 s Zeit und
  einem Leben, Neustart, Connector-Rahmen mit Zurück-Meldung, 0 Konsolenfehler.

## 26.11.2026 · Super Pimmelchen: Schnittstelle 1.1.0 (stop/pause/resume, :ende, feste Fehlercodes)

`spiele/pimmel/spimmel.html` (beide Ablagen byte-gleich): Die Startschnittstelle
ist auf **1.1.0** gehoben – `stop()`, `pause()`, `resume()` ergänzt, dazu das
Ereignis `superpimmelchen:ende` mit `{grund, punkte, level, levelAnzahl, tick}`
und ein übertragbarer `zustand()` (`tick`, `richtung`, `extras`). Der Vertrag
steht in [`SPIELE-SCHNITTSTELLE.md`](SPIELE-SCHNITTSTELLE.md).

- **Validierung streng:** nur `stufe` und `level`, nur echte Typen,
  `hasOwnProperty` statt `in`, unbekannte Felder abgelehnt. Feste Codes:
  `spiel-nicht-bereit`, `konfiguration-muss-objekt-sein`, `unbekanntes-feld`
  (Antwort nennt `feld`), `stufe-muss-text-sein`, `unbekannte-stufe`,
  `level-muss-1-bis-61-sein`, `spiel-laeuft`, `spiel-laeuft-nicht`.
- **Tastatur:** Leertaste und Enter starten nur, wenn **kein Bedienelement**
  fokussiert ist; auf Knopf, Auswahl oder Feld löst der Browser selbst aus
  (`button, select, input, textarea, a[href], [contenteditable]`). Gemessen mit
  echten Tastendrücken (15 Prüfungen, `arbeit/tmp/pimmel_tastatur_pruefen.js`):
  ohne Fokus starten Enter und Leertaste, im Level-Feld nicht (Fokus bleibt am
  Feld), HIGHSCORE öffnet mit beiden Tasten den Dialog statt der Partie,
  VOLLBILD startet nicht, „SPIEL STARTEN“ startet mit beiden, im Spiel feuert
  die Leertaste weiter, „WEITERSPIELEN“ setzt mit Leertaste fort.
- **Pause ohne Partie:** `pauseGame()` prüft den Zustand – `zustand().pausiert`
  steht auf dem Startbildschirm nicht mehr auf `true`.
- **Gemessen** (Chromium, 41 Prüfungen grün, `arbeit/tmp/pimmel_p1_p5_pruefen.js`):
  `{level:true}`, `{level:'12'}`, `{level:12.5}`, `{level:null}` →
  `level-muss-1-bis-61-sein`; `{stufe:3}` und `{stufe:null}` →
  `stufe-muss-text-sein`; `{farbe:'blau'}` → `unbekanntes-feld` (Feld `farbe`);
  Zustand danach unverändert. `tick` zählt Züge (0 → 4 nach vier Tasten),
  `pause()` hält Uhr und tick (120/4 → 120/4), `resume()` läuft weiter
  (120 → 118), `stop()` setzt `laeuft:false` und `startbildschirm:true`,
  `pause()`/`resume()` ohne Partie → `spiel-laeuft-nicht`. Enter auf HIGHSCORE
  öffnet den Dialog (vorher startete die Partie), Leertaste ohne Fokus startet
  nicht und feuert im Spiel weiter, 50 Startaufrufe → 0 Starts, P-Taste schaltet
  genau einmal, `blur` auf dem Startbildschirm setzt `pausiert` nicht,
  `zustand()` bleibt Kopie, 0 Konsolenfehler.
- **`:ende`** gemessen an einer Testkopie außerhalb des Repos (nur `start()`
  mit `lives = 1`, sonst unverändert): Selbstkollision nach `tick 56` →
  `{grund:'verloren', punkte:560, level:1, levelAnzahl:61, tick:56}`, danach
  `laeuft:false`, `startbildschirm:true`. Der Zweig „gewonnen" ruft dieselbe
  Funktion (`meldeEnde('gewonnen')`).
- Leveldaten unverändert (Levelblock 36 221 Zeichen, sha256 `0856654…` vor und
  nach der Änderung gleich), Spielablauf und Speicherschlüssel unangetastet.
  Die Einzeldatei bleibt bei Schnittstelle 1.0.0.

## 26.11.2026 · Super Pimmelchen: eigener Startbildschirm + Startschnittstelle

`spiele/pimmel/spimmel.html` (beide Ablagen byte-gleich): Beim Öffnen erscheint
ein eigener Startbildschirm im Retro-Stil – Titel, Schwierigkeit, Start-Level
(1–61), Übersicht der Einstellungen, großer Knopf „SPIEL STARTEN“, Highscore und
Vollbild. HUD, Spielfeld und Touch-Pad sind so lange ausgeblendet; es laufen
keine Timer. Kein Grafikstil-Knopf (KLASSISCH/MÄDCHEN-BOOM bleiben zurückgestellt).

- **Stufen:** LEICHT 5 s (vorher 10 s, jetzt laut Vorgabe), MITTEL 15 s, SCHWER
  20 s. Ein neues Special erscheint je Takt, alle bleiben liegen (Ausnahme
  Totenkopf wie bisher). Gespeichert wird weiter `superPimmelchenStufe`.
- **Ein Startweg:** `starteSpiel()` ist die einzige Stelle, die eine Partie
  beginnt – Startknopf, Enter/Pfeiltasten, Pause-Neustart und die Schnittstelle
  laufen alle darüber. Rückkehr zum Startbildschirm über das Pause-Blatt
  („STARTBILDSCHIRM“), Timer werden dabei gestoppt (`stoppeAlles`). Nach
  Spielende kommt der Startbildschirm von selbst zurück und zeigt „Letzte
  Runde: … Punkte · Level …“.
- **Schnittstelle:** `window.SuperPimmelchen` mit `version` (1.0.0), `bereit`,
  `start({stufe, level})`, `einstellungen()`, `zustand()` (nur Lesen) und
  `startbildschirm()`. `stufe` = `leicht|mittel|schwer`, `level` = 1–61;
  fehlende Felder bleiben auf der aktuellen Auswahl, ungültige Werte werden mit
  `{ok:false, fehler}` abgelehnt, ohne das Spiel anzufassen; bei laufender
  Partie `fehler:'spiel-laeuft'`. Nach der Initialisierung `bereit = true` und
  Ereignis `superpimmelchen:bereit`. Ein gültiger Aufruf überspringt den
  Startbildschirm; unbekannte Felder werden ignoriert (später erweiterbar).
- **Geprüft (Chromium, 24 Einzelprüfungen grün + Takt-/Layout-/Ende-Läufe):**
  Startbildschirm beim Öffnen, MITTEL vorgewählt, LEICHT/MITTEL/SCHWER schalten
  Anzeige und Takt; echte Specials gemessen: 4 s → 0 und 6,1 s → 1 (5 s),
  13 s → 0 und 16 s → 1 (15 s), 18 s → 0 und 21 s → 1 (20 s), nach 22 s liegen
  4; Pause hält Takt und Uhr an; Start-Level 1, 30 und 61 laden; Uhr läuft nach
  Neustart einfach (3 s → 3 s); Pause → Startbildschirm → Neustart sauber;
  Spielende bringt den Startbildschirm mit Ergebniszeile; Highscore-Eintrag
  bleibt lesbar; ungültige Schnittstellenwerte (Level 0/99/„abc“, Stufe
  „ultra“, String, leeres Objekt bei laufender Partie) ändern nichts; 0
  Konsolenfehler. Layout: Startblatt 581 px (PC/Tablet), 604 px (Handy hoch),
  345 px (Handy quer, zweispaltig) – ohne Scrollen, alle Knöpfe im Bild,
  Startknopf 52/40 px, Stufen 44/38 px; Vollbild aus dem Startmenü und im Spiel
  (Feld 1272×699, 382×210, 628×345) ohne Verschiebung; Tastatur 9,9→13,9,
  Touch 13,9→17,9. Leveldaten (36 221 Zeichen) byte-gleich.
- **Einzeldatei nachgezogen:** `Super-Pimmelchen-61-Level-Einzeldatei.html`
  zeigt denselben Startbildschirm (gleiches Blatt, gleiche Blöcke) und hat
  dieselbe Schnittstelle (`window.SuperPimmelchen` 1.0.0 mit `bereit`,
  `start`, `einstellungen`, `zustand`, `startbildschirm`). Ihre Eigenheiten
  bleiben: `:fullscreen` statt `bildschirm-voll`, ein einzelnes Special
  (`extra`) statt Liste, kein Pause-Blatt. Ihr bisheriger Zufallstakt
  (5–20 s) folgt jetzt der Stufe (5/15/20 s, gespeichert in
  `superPimmelchenStufe`). Leveldaten ebenfalls byte-gleich (36 221 Zeichen);
  20 Prüfungen grün (Startbildschirm, 61 Level, MITTEL vor, Takt echt gemessen
  4 s→0/6,2 s→1, 13 s→0/16 s→1, 18 s→0/21 s→1, Level 30/61, Schnittstelle
  gültig/ungültig, Highscore, 0 Konsolenfehler) und Layout 575/565/336/575 px
  ohne Scrollen. Ohne Pause-Blatt gibt es dort keinen Weg zurück mitten in der
  Partie – nur nach Spielende (automatisch) und über
  `SuperPimmelchen.startbildschirm()`.

## 26.11.2026 · Smilie-Auswahl: Kategorien beim ersten Einblenden zusammengedrückt

Die Reiter über dem Smilie-Raster waren beim ersten Öffnen nur noch ein dünner
Streifen (Höhe 10 px, Text abgeschnitten); erst nach einem Klick auf eine Kategorie
hatten sie ihre richtige Höhe. Breite war nie betroffen (41–94 px je Reiter).

- Ursache: `.smiley-panel` ist eine Spalte mit `max-height:min(50vh,340px)`. Beim
  Reiter „Alle“ meldet das Raster 2100 px Inhaltshöhe; da Reiterzeile (durch
  `overflow-x:auto` mit Mindesthöhe 0) und Kopfzeile schrumpfbare Flex-Elemente
  waren, verteilte der Browser die fehlende Höhe auf genau diese beiden – das
  Raster selbst behielt seine Mindesthöhe. Am Handy stand die Schutzregel
  (`flex:0 0 auto`) schon im `@media (max-width:1023px)`-Block, am Desktop fehlte sie.
- Korrektur (nur CSS, `connector.html`): `.smiley-panel .smiley-kopf,
  .smiley-panel .smiley-tabs{flex:0 0 auto}` – Kopf und Reiter behalten ihre Höhe,
  das Raster scrollt wie bisher.
- Gemessen (Chromium, erster Ladevorgang): Reiterhöhe 24 px sofort, nach 900 ms,
  nach Kategorieklick und nach Schließen/erneutem Öffnen – alle acht Reiter gleich
  hoch, Reiterzeile 26 px. Desktop 1280×800 (Panel 383 px, Reiterzeile 365 px breit),
  Desktop 900×700, Handy 390×844 (374/356) und Handy quer 844×390 (828/810) gleich.
  Vorher: 10 px beim ersten Öffnen, 19,8 px nach Klick (Desktop).
- Funktion unverändert: 275/275 Bilder geladen, Suche „tanz“ 15 Treffer, Kategorie
  „tiere“ 15, Auswahl schließt das Panel und setzt die Vorschau, 0 Konsolenfehler,
  Raster scrollt (2100 px Inhalt bei 251 px Höhe).
- Ablagen byte-gleich, Sicherung der Hilfsskripte nicht nötig (Einmalprüfungen).

## 26.11.2026 · Smilies auf den neuen, sortierten Bestand umgestellt

Maßgeblich ist jetzt der neu sortierte Ordner `spiele/connector/smilies` mit
Unterordnern als Kategorien. Er ist in beiden Ablagen byte-gleich übernommen;
der alte Stand (341 Dateien) liegt als Sicherung in
`arbeit/sicherung/2026-11-26-smilies-neu/smilies-alt`.

**Bestand:** 275 Grafiken – **149 png, 101 gif, 24 webp, 1 jpg**; dazu die
unveränderte Anleitung `s (1).md` (per `.gitignore` nicht im Repo). Verteilung:
**böse 56, love 37, tanz 15, tiere 15, traurig 13, xxx 68**, ohne Ordner 71.
Im Repo sind dafür 272 alte Dateien entfallen, 68 an gleicher Stelle ersetzt und
207 neu dazugekommen. Inhaltlich sind alle Bilder des alten Standes bis auf zwei
(`s (138).png`, `s (139).png`, je ~1,1 kB) im neuen Bestand enthalten – die
übrigen sind umbenannt oder ersetzt; 63 Bildinhalte sind neu. Nichts gelöscht,
umbenannt oder ausgetauscht wurde von mir.

**Kein Codeeingriff nötig:** Server (`smilies.go`) liest den Ordner bei jeder
Abfrage neu, ein Unterordner ist die Kategorie, und die Auswahl im Connector
(`connector.html`) baut Reiter, Suche und Vorschau aus dieser Antwort. Ein
Unterordner mit Umlaut (`böse`) funktioniert ohne Sonderfall, weil der Pfad
kodiert und serverseitig wieder dekodiert wird.

**Geprüft am laufenden Connector (Port 8778, echter Server):**
`GET /api/smilies` → **275 Einträge**, Kategorien wie oben, keine doppelten
Namen; **alle 275 Pfade 200** und byte-gleich zur Dateigröße, auch
`/spiele/connector/smilies/böse/böse (1).gif`. Auswahlfenster: Reiter
„Alle · böse · love · tanz · tiere · traurig · xxx · Ohne Ordner“, **275 Bilder
in „Alle“, 0 kaputt, 0 doppelt**, je Kategorie 56/37/15/15/13/68/71 mit allen
Bildern geladen. Suche „love (3“ → 2 Treffer, „böse“ → 56, „tanz“ → 15. Eine
Grafik aus `böse` gewählt, gesendet und im Chat wieder gefunden
(`böse/böse (10).gif`, geladen, nicht als „Nicht gesendet“ markiert, 0
Konsolenfehler). `go vet ./...` sauber, `go test -count=1 ./...` **283,6 s, ok**
(kein Go-Test berührt Smilies).

**Bekannt, bewusst nicht angefasst:** `love/love (1).webp` und
`xxx/xxx (4).webp` sind byte-gleich (12 358 B) – doppelter Inhalt, aber zwei
gewollte Einträge; gelöscht wird hier nichts ohne Ansage.

---

## 26.11.2026 · Super Pimmelchen: Pausenzeile im Hochformat unter dem Feld

Die Leiste mit Pause, Neustart und Vollbild lag im Hochformat als schwebender
Block über den obersten Feldreihen, obwohl unter dem Feld viel Platz frei war.
Jetzt steht sie dort im Textfluss **unter dem Spielfeld** (gemessen 390 × 844:
Feld endet bei 258 px, Leiste beginnt bei 264 px, 30 px hoch, rechtsbündig,
Steuerung darunter ab 306 px) und `passeBuehne()` zieht ihre Höhe von der freien
Fläche ab. Im Querformat bleibt sie die schwebende Leiste (844 × 390: Leiste
`absolute` bei x 699/y 33), im Vollbild ebenfalls. Geprüft: Hochformat,
Drehung hin und zurück, Pause an/aus, keine Konsolenfehler.

---

## 26.11.2026 · Super Pimmelchen fertiggestellt: drei Stufen, Querformat, Vollbild, Pause

**Nur `spiele/pimmel/spimmel.html` angefasst** (beide Ablagen byte-gleich), Sicherung
vorher in `arbeit/sicherung/2026-11-26-super-pimmelchen/`. Die 61 Level sind
unverändert: `LEVEL_1_MAP` plus 60 Einträge in `LEVEL_MAPS`; alle 61 über die
Levelauswahl geladen, 0 Warnungen („Level … fehlt oder hat falsche Dimension“),
Anzeige läuft bis „61“.

**Stufen.** Leicht 10 s, Mittel 15 s, Schwer 20 s – nur der Takt der Specials hängt
daran (`STUFEN`), die Wahl steht in der Levelauswahl und wird als
`superPimmelchenStufe` gespeichert. Gemessen im Browser (Zeitstempel im Taktgeber):
Leicht **10,01 / 9,99 s**, Mittel **15,01 s**, Schwer **20,00 s**.
Der Auftrag nennt in §9 für Leicht „5 Sekunden“, in §3 „10 Sekunden“ – umgesetzt
ist §3 (10 s); eine Zeile in `STUFEN` ändert das.

**Specials bleiben liegen.** Bisher gab es genau einen Platz (`extra`): lag dort
etwas, erschien nichts Neues. Jetzt eine Liste (`extras`), neue Specials suchen sich
ein freies Feld, ein liegendes wird weder ersetzt noch gelöscht; nur der Totenkopf
verschwindet weiter nach 5–10 s (bestehende Regel). Gemessen: nach 11 / 22 / 33 s
lagen **1 / 2 / 3** Specials gleichzeitig.

**Takt ohne Doppelauslösung.** Ein einziger `setInterval` (200 ms) prüft eine
Fälligkeit; er wird beim Pausieren angehalten, beim Weiterlaufen neu gestellt und
von Levelwechsel oder Drehung **nicht** neu gestartet. Gemessen: 10,43 → 9,99 s
über eine Drehung hinweg, 18,09 s über eine 6-s-Pause (Pause + neuer Takt, kein
Nachtrag), 9,99 s über einen Levelwechsel.

**Querformat (390 × 844 → 844 × 390, Touch).** `passeBuehne()` rechnet das Feld aus
der freien Fläche (20:11, nie verzerrt): Hochformat **378 × 207** (Verhältnis 1,826),
Querformat **631 × 347** (1,818) bei 31 px hoher Anzeigezeile, kein Hoch- oder
Querscrollen (Seite genau 844 bzw. 390). Im Querformat liegt die Steuerung
halbtransparent unten rechts über dem Feld (44-px-Tasten, Feuerknopf 84 px),
Überschriften, Startleiste und der Connector-Knopf sind im Spiel ausgeblendet.
Drehen zurück ergibt wieder 378 × 207; Neuberechnung über `resize`,
`orientationchange`, `visualViewport` und `fullscreenchange`.

**Vollbild.** Ein und aus über die Leiste (PC/Handy), zusätzlich aus dem Pause-Blatt;
die Größe wird nach dem Wechsel neu gerechnet (Hochformat **382 × 210**, Verhältnis
1,819; Querformat 345 px hoch). Erlaubt der Browser kein echtes Vollbild, füllt
`body.pseudo-vollbild` das Fenster – gleiche Ansicht, gleicher Weg zurück.

**Pause.** Neu: Knopf in der Spielleiste, `P`/`Escape`, Tabwechsel und Fensterverlust;
das Blatt bietet Weiterspielen, Neustart, Vollbild und Verlassen. Nach dem Weiter
läuft die Bewegung von selbst wieder an.

**Steuerung.** Pfeiltasten und WASD unverändert, Leertaste schießt, Enter startet.
Loslassen hält an (gemessen: x 10 → 13, danach 0,6 s unverändert), Tippen auf ▲ setzt
die Richtung. Der Connector-Knopf bleibt in der Kopfzeile, ohne eine zweite Reihe zu
öffnen (sechste Spalte).

**Geprüft:** alle 61 Level, Takt je Stufe, liegende Specials, Drehung, Pause,
Levelwechsel, Neustart (Punkte 0, Level 01), Tastatur, Touch, Vollbild an/aus,
PC 1280 × 800 (Feld 948 × 521, Seite genau 1280 × 800) – **0 Konsolenfehler**.
Nicht durchgespielt: eine Partie über alle 61 Level am Stück; IOS-Safari und echte
Geräte wurden nicht getestet.

---

## 26.11.2026 · Mädchenstil am PC ruhiger, Rotlicht auf vier Stunden ausgelegt

**Mädchenstil.** Am Handy trägt die kräftige Rosafläche die Seite, am PC wurde
dieselbe Fläche zur Leuchtwand: bei 1440 px füllte `#f7a6d4` alles außer der
Inhaltsspalte, Karten und Tafeln hoben sich kaum ab. Jetzt liegt auf großen
Bildschirmen ein Verlauf aus Weiß, Rosé und Altrosa (`#fdf8fa → #f9e7ee →
#f3d5e1 → #eecad7`) unter einer sehr leisen Herzkulisse (SVG-Kachel, 264 px,
Deckkraft 3,4–4,8 %, keine Animation). Geändert in `spiele/stil.js` – dort
liegt die zentrale Stilwahl, 101 der 104 Seiten mit Mädchenstil laden sie –
und in `spiele/wuerfelspiele/stil.css` (hellere Tafeln und Karten, feinere
Linien, weicherer Schatten, Herz neben der Überschrift, sichtbarer
Tastaturfokus). Alles hängt an `@media (min-width: 641px)` bzw. `561px` und an
`html[data-stil="maedchen"]`: Klassisch und Buben sind unberührt, die
Handy-Ansicht ist **pixelgleich** geblieben (Vergleich 390 × 844 vorher/nachher
mit `cmp`). Geprüft bei 1024, 1440 und 1920 px an Sammlung, Würfelpoker, Dame,
Kniffel, 1×1 und der Spielehalle; Text auf Karte und Tafel bleibt dunkel auf
fast weiß (Kontrast über 7:1), nur Flächen, Rahmen und Hintergrund ändern sich.

**Rotlicht.** Die Partie war nach **58 Minuten** durch (Simulation, optimales
Spiel), weil die Einnahmen am gesamten Multiplikator hingen: Stufe, Upgrades und
Standorte multiplizierten den Lohn, die Preise blieben aber gleich – jede Stufe
machte die nächste billiger, die letzten fünf Stufen fielen im Minutentakt.
Neu:
- **Stufen** kosten gleichmäßig das 43-Fache der vorigen (500 € bis 5,84 Brd. €)
  und heben die Einnahmen nur noch von ×1,5 auf ×3,48 (vorher ×1 bis ×6).
- **Personal** ist teurer (Preise ×1,6 bis ×2,2 ab Ali), zahlt mehr Lohn und
  steigt je Kopf um 1,33–1,39 (vorher 1,23–1,29) – Mengenkauf lohnt nicht mehr.
- **Verbesserungen** wirken schwächer (×1,06 bis ×1,32 statt ×1,15 bis ×1,8),
  **Standorte** geben ×1,05 (vorher ×1,22) und kosten das 2,5-Fache.
- **Erfolge** geben zusammen +30 % (vorher +85 %), **Ereignisse** zahlen 15–45 s
  Ertrag statt 20–90 s.
- **Offline:** zwei Stunden zählen voll, die beiden Buchhaltungen geben je eine
  dazu (maximal vier, vorher acht). Eine Nacht kürzt die Partie damit nicht ab.
- Der Klick-Erfolg „1000 Handgriffe“ war bei 60 s Takt nie erreichbar und heißt
  jetzt „400 Handgriffe“.

Gemessene Spieldauer bis „Stufe 10 + alle Verbesserungen“ (1-s-Simulation,
`pruefung.js`, drei Spielweisen): aktiv **3:55–4:21 h**, normal **3:49–4:30 h**,
vorsichtig **5:25–5:48 h**. Alle 20 Erfolge werden in jedem Lauf erreicht.
Stufenabstände im Normallauf: 9, 31, 27, 22, 18, 42, 39, 23, 35 Minuten – in
jeder halben Stunde passiert etwas. Zwei Stunden Pause nach 20 Minuten Spiel
kürzen die aktive Spielzeit um rund 20 Minuten (4:17 → 3:57); acht Stunden
Pause bringen nichts mehr, weil bei vier Stunden gekappt wird. Neue Prüfung `sicherung/2026-11-26-rotlicht-balance/pruefung.js`
(71 Prüfungen, 0 Fehler) samt `durchspiel.js` und `offline.js`; der alte Kern
liegt als `rotlicht-kern.vorher.js` daneben.

---

## 26.11.2026 · Wiederverbindung: Partie, Chat und Platz überleben den Abbruch

**Fehler (reproduziert, nicht vermutet):** Bei einer laufenden 10.000-Runde zu
zweit erschien „Verbindung wackelt“, danach war der Chat leer, die Partie weg,
beide Spieler standen aber weiter im Raum und die Anzeige sagte „Verbunden“.
Nachgestellt mit einem echten Serverneustart: vorher Raum `TG-F000`, 7
Chatzeilen, Partie `active`; nachher standen beide in einem **neuen** Raum
`TG-4C0C` mit 3 Zeilen (nur Beitritt), der Spielrahmen zeigte weiter auf
`TG-F000` und blieb bei „Verbindung wackelt“ – dauerhaft.

**Ursache (drei Punkte, alle im Code nachweisbar):**
1. Räume und Partien liegen nur im Speicher; `/api/zustand` antwortet nach einem
   Neustart mit 404. `join()` hat das als „Raum weg“ gedeutet, `activeRoom`
   geleert und über `/api/rooms` **still einen neuen Raum** angelegt (dessen
   Kennung war zufällig). Der Chat war damit leer, die Partie unerreichbar – und
   der Rahmen fragte weiter den alten Raum, also für immer „wackelt“.
2. `gameTimeout` stand bei **90 s**, die Platzreservierung bei 150 s: Ein Funkloch
   über 90 s beendete die Partie, obwohl der Platz noch reserviert war.
3. Die Verbindungsanzeige meldete „Verbunden“, sobald irgendein Ruf durchging –
   ohne zu prüfen, ob die Spielsitzung noch existiert. Es gab keinen Weg zurück
   und keinen Endzustand, nur eine Endlosschleife.

**Geändert:**
- `spiele/connector/types.go`: `gameTimeout` 90 → **150 s**, `presenceTimeout`
  150 → **180 s** (erst die Partie, dann der Platz). Nach 60 s Stille meldet der
  Server **einmal** „X ist gerade nicht erreichbar – die Partie bleibt bis zu
  zwei Minuten stehen.“ (neues Feld `WegGemeldet`), statt sofort aufzulösen.
- `spiele/connector/lobby.go`: Der gemeinsame Raum hat die **feste Kennung
  `TG-HALLE`** und wird bei Bedarf neu angelegt. Nach einem Neustart finden
  offene Seiten denselben Raum wieder, statt in einem neuen zu landen.
- `spiele/connector/zustand.go`: `start` (Startzeit des Prozesses) in jeder
  Antwort – daran erkennt die Seite einen Neustart.
- `spiele/connector/idempotenz.go` (neu): Doppelte Spielzüge. Jeder POST trägt
  eine Marke (`X-Aktion`); dieselbe Marke wird innerhalb von 90 s **nicht noch
  einmal ausgeführt**, sondern liefert die gespeicherte Antwort (nur 2xx, max.
  128 KB, max. 400 Einträge). Lebenszeichen, Chat, Raumverwaltung und
  Einladungen sind ausgenommen.
- `spiele/connector/connector.html` (byte-gleich in `connector-online/`):
  Zustandsautomat für die Verbindung – „Verbunden“, „Verbindung unterbrochen –
  Wiederherstellung läuft“, „Wiederverbindung läuft (Versuch n von 6)“,
  „Wieder verbunden“, „Verbindung verloren“ mit Knopf **„Zurück ins Spiel“**.
  Sechs Versuche mit 1/2/4/8 s Pause, ohne die alte 30-s-Sperre; kein stiller
  Raumwechsel mehr (404 → ehrliche Meldung + Knopf). Nach einer Wiederverbindung
  prüft derselbe Statusruf die Partie: Fehlt sie, wird der tote Rahmen
  geschlossen und der Grund genannt. Ein Serverneustart (andere `start`-Zeit)
  wird ausdrücklich gemeldet.
- `spiele/token.js`: Zentral für **alle** Spielseiten – ein 403 „bitte
  beitreten“ meldet den Platz einmal neu an (mit eigenem Token und
  Gerätekennung) und wiederholt dieselbe Anfrage mit derselben Aktionsmarke.

**Geprüft (echter Connector, Chromium, zwei getrennte Browser-Kontexte):**
- Serverneustart mitten in der Partie: beide bleiben in `TG-HALLE`, Status nach
  2 s „Verbunden“, toter Rahmen geschlossen, Meldung „Der Spielserver wurde neu
  gestartet …“; vorher (alter Stand) landeten beide in `TG-4C0C` mit leerem Chat.
- 12 s weg während eines Zuges, Gegner schreibt: gleiche Kennung, Partie
  `active`, Würfel `4,2,2,5,3,4` unverändert, Nachricht des Gegners im Chat.
- 100 s weg: Partie `active`, gleiche Würfel und Punkte, Rahmen offen; der
  Mitspieler sah nach 60 s die „nicht erreichbar“-Meldung.
- Beide gleichzeitig 20 s weg: beide wieder `active`, gleiche Würfel, Rahmen
  offen.
- 50 s ohne Netz: „Wiederverbindung läuft (Versuch 4/6)“ → „Verbindung
  unterbrochen – Wiederherstellung läuft“ → nach 6 Versuchen „Verbindung
  verloren“ **mit Knopf**; Netz zurück → „Verbunden“, gleiche Partie, gleiche
  Kennung.
- `go vet ./...` 0; `go test ./...` connector **237,6 s**, grün; connector-online
  `go vet` + `go build` grün, `connector.html` byte-gleich.
- Neu: `wiederaufnahme_test.go` (lokal, nicht im Repo) – 100 s Aussetzer beendet
  die Partie nicht, 160 s beenden sie, der Platz bleibt bis 180 s; fester
  Hauptraum überlebt einen Neustart; dieselbe Aktionsmarke wird nur einmal
  ausgeführt (`X-Aktion-Wiederholt: 1`).

**Grenzen:** Ein echter Serverneustart löscht den Speicher – Chat und Partie von
vorher sind dann wirklich weg (Render hat keine dauerhafte Platte). Das wird
jetzt gesagt und der Rahmen geschlossen, statt es zu verstecken. Getestet ist
10.000 durchgespielt; die Wiederverbindung sitzt zentral in `connector.html`,
`token.js` und `idempotenz.go` und gilt damit für alle Spiele, einzeln
durchgespielt wurden sie nicht. Zweiter Spieler war ein zweiter Browser-Kontext,
kein echtes zweites Gerät.

---

## 26.11.2026 · Favoriten-Kacheln öffnen wieder den Mehrspieler-Weg

**Fehler:** Die sechs Bild-Kacheln über der Spieleübersicht starteten mit
`openGame(…)` direkt in die Partie – ohne den Tisch, den „Online spielen“ über
`/api/angebot` anlegt. Niemand konnte beitreten, die Partie wartete ohne Runde.
Dasselbe tat `?spiel=…`. Die Spielhalle hatte den Weg in den Connector ganz
verloren: beim Vorab-Rendern von `spiele.html` (Commit `c24f11c`) fielen
`multiplayer-register.js` und `halle-online.js` heraus, jede Kachel verlinkte
direkt auf die Spielseite – auch bei den 71 Spielen mit Mehrspieler-Modus.

**Geändert** (`spiele/connector/connector.html`, byte-gleich in
`connector-online/`): Favoriten-Kachel und `?spiel=…` rufen `zeigeSpielKarte()` –
Kategorie auf, Karte in den Blick, kurz umrandet, Hauptknopf fokussiert. Damit
gilt für Kachel und Karte dieselbe Startlogik mit allen Modi. In
`spiele/spiele.html` und `spiele/halle-online.js` sind beide Skripte wieder
eingebunden; Klicks auf Favoriten- und Hallenkacheln führen für Spiele aus
`multiplayer-register.js` in den Connector (`?spiel=<kennung>`), Einzelspieler
und Werkzeuge bleiben direkt. Kachel-Ziele werden mit umgeschrieben (neuer Tab),
Pfade normalisiert (`brettspiele/…` ↔ `/spiele/brettspiele/…`).

**Geprüft** (Chromium, echter Connector): 10.000 über Favoriten-Kachel →
Karte/Modi, „Online spielen“ legt den Tisch an; zweiter Browser über
`?room=TG-…` im selben Raum, „Mitspielen“ öffnet dieselbe Partie, beide Lobbys
zeigen beide Spieler, „Bereit“ beider schaltet „Spiel starten“ frei. Normale
Hallenkarte nimmt denselben Weg; Mahjong und Würfelbecher starten direkt;
Mahjong-Karte ohne „Online spielen“; „Gegen Computer“ (Pig) mit Bot, ohne
Angebot; Handy 390 × 844 mit Karte im Bild (337 px). `go test ./...` grün.
Grenze: Emulation statt echtem Telefon, zweiter Spieler als zweiter
Browser-Kontext.

## 26.11.2026 · Chat am Handy tiefer, Smiley-Auswahl passt ins Bild

**Fehler:** Im Handy-Layout stand `#chatbox` per `order:1` direkt unter dem Kopf
(Chatbeginn bei 46 px). Das Smiley-Fenster hing fest über der Eingabezeile
(`position:absolute; bottom:calc(100% + 6px)`) – weil die Eingabezeile ganz oben
lag, öffnete es nach oben aus dem Bild: hochkant `top −2 px`, quer `−24 px`, die
obersten Kacheln waren abgeschnitten und nicht antippbar.

**Geändert** (`spiele/connector/connector.html`, byte-gleich in `connector-online/`):
Reihenfolge im Handy-Block jetzt Spieler (1), Einstellungen (2), Chat (3); dazu
`smileyPanelPlatzieren()`, das die Eingabezeile und den sichtbaren Bereich
(`visualViewport`) misst und das Fenster über der Eingabe (ab 240 px Platz), sonst
darunter, sonst über die ganze Höhe öffnet – Lage und Höhe auf 8 px Rand begrenzt,
`position:fixed` nur in diesem Bereich, neu gerechnet bei Drehen, Tastatur,
Adressleiste und Scrollen. Solange es offen ist, ist der schwebende Chat-Knopf aus.
Am PC (≥1024 px) steigt die Rechnung sofort aus; dort bleibt alles beim CSS.

**Gemessen** (Chromium-Emulation, echter Connector 8767): Chat beginnt statt bei
46 px bei **320 px**; Fenster hochkant **8 … 306** (vorher `−2 … 338`), quer
**8 … 348** (vorher `−24 … 171`), bei auf 560 px verengtem Fenster (Tastatur)
**227 … 552**, mit offener Partie (Chat 287 px) **274 … 614** – überall `passt: true`.

Oberste und unterste Kachel angetippt → Auswahl übernommen, Nachricht mit Bild
gesendet (`naturalWidth > 0`), Escape und Klick daneben schließen, Drehen bei
offenem Fenster rechnet neu (298 → 340 → 298), Partie öffnen/schließen unverändert,
keine Seitenfehler. **PC 1280 × 800 unverändert:** Chat 147 … 800 (653 px), Fenster
381 … 721 (340 px). Grenze: Geräte-Emulation statt echtem Telefon, Tastatur über
kleinere Fensterhöhe nachgestellt.

## 25.11.2026 · Snake wieder Einzelspieler, Würfelspiele als Mehrspieler

**Snake:** Die Spielhalle öffnet jede Seite mit `?room=…&id=…`; Snake schaltete daraufhin
in die Zwei-Spieler-Lobby um – der Einzelspieler war aus der Halle nicht erreichbar. Das
Duell gilt jetzt nur mit `&art=duell` (der Connector setzt das an seinen vier
Snake-Aufrufen, `spielAdresse(path, art)`); ohne den Zusatz startet Snake allein, bei
Raum+Kennung steht zusätzlich „Duell im Raum“ im Menü.
Geprüft (headless Chromium, ohne Testhaken): Startbild „SNAKE · EINZELSPIELER“, HUD nur
Punkte/Länge/Tempo/Rekord, kein Mitspielertext im DOM, Game Over nach 3,9 s, „Nochmal“
setzt auf 0 Punkte / Länge 3, zweiter Lauf läuft. Apfel: 10 Punkte, Länge 3 → 4, Rekord
10 im `localStorage`. Duell im Raum TG-33D5 gegen Bot „Mumpitz“ (`action=bot`): Menü
ausgeblendet, Stand „Namenloser 2 0 : Mumpitz 1“, Runde 1 abgeschlossen.

**Würfelspiele:** Alle 25 (7 unter `brettspiele/`, 18 in `wuerfelspiele/`) waren als
`solo` geführt und standen damit unter „Einzelspieler“, obwohl sie einen Tisch für 2–6
Personen und über den Connector Computergegner haben. Merkmale jetzt
`lokal zwei computer`, auf den Hallenkarten „Solo“ → „Bots“.
Gemessen im Connector: „Einzelspieler“ 52 → **36** Spiele (0 Würfelspiele),
„Mehrspieler“ **72** mit allen 25; je Würfelkarte die Knöpfe „Am Gerät spielen“,
„Gegen Computer“, „Online spielen“. Lokal zu zweit: Beelzebub, 5 Runden,
**40 : 20** mit Siegmeldung. Gegen Computer: Beelzebub mit Bot „Vroni“, Bot würfelt und
punktet (27). `go test ./...` in `spiele/connector` grün (229 s, u. a.
`bot_wuerfel_test.go` über alle 18 Spiele, `bot_snake_test.go`).

**Zahlen und Wege:** `spiele.html` lädt jetzt `spieldaten.js` und setzt „N Spiele ·
N Kategorien“ daraus (**95 / 6**, vorher 94); Werkzeugkachel und
`android/start/app.html` ebenfalls 95. `python spiele/pruefe-spieldaten.py` läuft ohne
Abweichung (vorher 3). Keine CDNs, keine neuen Bibliotheken.

**Noch offen:** Das verlinkte APK ist der alte Bau; `assets/www` und `katalog.json`
entstehen beim Bauen und wurden nicht von Hand geändert – die App zeigt die neue
Zuordnung erst nach einem Neubau.

## 24.11.2026 · Schritt 7: Cache- und Ausfallverhalten geprüft

Geprüft auf Emulator B (Android 14) mit dem Debug-Bau 1.17/18, Flugmodus
(`Active default network: none`), echte Tipps über `uiautomator`, WebView-Inhalt über
DevTools. Kein Quellcode geändert – nur Messungen; Testkatalog und Testseite danach
zurückgebaut (Katalog `fa50d261…`, Mahjong-Seite `1eaab258…`).

- **Ohne Netz:** Start 722 ms (`am start -W`, COLD), Spielhalle 94 Spiele / 6 Kategorien
  in 54 ms, lokales Spiel (Mahjong) 3,52 s nach dem Tipp geladen, 144 Bilder / 0 kaputt,
  `readyState complete`. Einziges Log: `www/favicon.ico` fehlt (404). Ein Online-Spiel
  ohne Netz zeigt die englische Chromium-Seite „Webpage not available“
  (`net::ERR_INTERNET_DISCONNECTED` −2); die App schreibt danach trotzdem
  „Spielseite geladen“.
- **Veralteter Katalog:** 95 Einträge – einer auf eine fehlende Seite, einer mit
  unbekannter Kategorie – werden ungeprüft gelistet. Fehlende Seite → 404 aus
  `HomepageAssets`, Chromium-Seite `net::ERR_HTTP_RESPONSE_CODE_FAILURE`, keine
  App-Meldung und kein Weg außer „Zurück“. Unbekannte Kategorie → Kopf „Weitere · 1“.
  Unlesbares JSON → „Der Spielekatalog fehlt im App-Paket. Bitte die App neu
  installieren.“, kein Absturz.
- **Liegengebliebene Seiten:** 303 HTML-Seiten im Paket, 105 unter `spiele/`, 94 im
  Katalog; die 11 übrigen sind Werkzeuge/Infoseiten und alle verlinkt (0 verwaist). Eine
  künstlich abgelegte Alt-Seite (60 B) verschwindet beim nächsten `homepageKopieren`.
  Nach einem App-Update erscheint ein geänderter Seitentitel sofort – kein alter
  Cache-Stand.
- **Cache:** nach dem Leeren und einem Spiel + Halle 324 KB in 22 Dateien, davon 216 KB
  V8-Code-Cache; keine Seitenkopien im HTTP-Cache.

Offen (nur festgestellt, nicht geändert): Für Online-Spiele ohne Netz und für
Katalogeinträge ohne Seite gibt es keine deutsche Meldung in der App.

---

## 24.11.2026 · Schritt 2: Alte Zahlen und Bauskripte bereinigt

- **Spielzahlen:** Die App-Kachel in `spiele.html` sagte „91 Spiele offline“ →
  **94**. Der App-Start (`start/app.html`) sagte „65 Spiele mit mehreren“ →
  **71** (Online-Spiele; 72 ist die Vereinigung mit den lokalen
  Zwei-Personen-Spielen). `spiele/android/README.md` nannte zweimal 91 →
  **94**. `ANDROID-ARCHITEKTUR.md`: 91 → 94, 9 → **6 Kategorien** (so viele
  stehen in `spieldaten.js` und `katalog.json`). `ENTSCHEIDUNGEN.md`: 91 → 94,
  Stand 24.11.2026 mit 6 Kategorien; der alte Widerspruch bleibt als Hinweis
  („früher stand dort 91, im Katalog 94“) erhalten.
- **Alte Bauskripte aus dem App-Paket:** `gewonnen_js_bauen.py` und
  `lob_js_bauen.py` lagen als Kopien unter
  `assets/www/spiele/` und damit im APK. Entfernt; `homepageKopieren`
  schließt jetzt `**/*.py`, `**/*.sh` und `**/*.ps1` aus. Die Quellwerkzeuge
  unter `spiele/` bleiben (sie bauen `gewonnen.js`/`lob.js` aus den lokalen
  Textdateien).
- **Keine eingebettete Spieleliste mehr:** `KATEGORIEN`, `SPIELE`, `TEXTE`
  und `HIGHLIGHTS` werden nur noch in `spiele/spieldaten.js` definiert.
  `spiele.html` und beide `connector.html` laden die Datei je einmal
  (`<script src…spieldaten.js>`), enthalten keine Pfad-Zeilen der alten Liste
  (0 Treffer) und keine Inline-Definition. Gemessen im laufenden Connector:
  Halle **94 Karten**, Connector **95 Karten**, je **1** Daten-Skript,
  **0** Inline-Listen, **0** Seitenfehler.
- **Noch nicht neu gebaut:** Das vorhandene APK (17 / 1.16) enthält die zwei
  `.py`-Kopien und die alte Cabo-Zuordnung weiterhin; beides verschwindet
  erst mit dem Neubau in Schritt 4. Nichts veröffentlicht, nichts gepusht.

---
## 24.11.2026 · Spielzuordnung abgeschlossen: Cabo ist Einzelspieler

**Prüfung (alle 94 Spiele der Halle):**
- **Online (71):** Register-Eintrag, Seitenaufruf des eigenen Endpunkts (direkt oder über
  das jeweilige Modul, z. B. `wuerfelspiele/online.js`, `halma/js/main.js`,
  `risiko-fern.js`, `paare` über `API+"/paare"`), Route als `mux.HandleFunc` in
  `connector/server.go`; `go test -run Register` grün. Kein Spiel außerhalb des
  Registers ruft einen Spiel-Endpunkt auf.
- **Lokal (55):** 54 Spiele mit ausdrücklichem Zwei-Personen-Modus (Spielerzahl 2,
  „2 Spieler“, „Lokal 2–4“, geteilte Steuerung/abwechselnde Seiten); Arschloch lokal
  3–9 (Optionen 3…9), also lokal, aber nicht zu zweit.
- **Einzelspieler (56):** Solo-Modus, „1 Spieler“ oder Gegner-KI.

**Korrektur:** `Cabo` hat einen eigenen Knopf **„Gegen Computer“** (`#btnSolo` →
`tu("solo")`, Schwierigkeit Leicht/Mittel/Schwer) und war fälschlich ohne
Einzelspieler geführt. In `spieldaten.js` jetzt `solo`. Damit: **56 Einzelspieler**,
55 lokal, 54 zu zweit, 71 online, Mehrspieler 72. Gemessen: Connector 95/56/72/71/54,
Halle Solo 56, Cabo in beiden „Solo“+„Online“.

**Arschloch und Acia – kein Widerspruch, zwei verschiedene Fragen:**
- Arschloch: lokal 3–9 → „Lokal“ ja, „Zu zweit“ nein.
- Acia: 1 oder 2 Spieler → „Lokal“ ja, „Zu zweit“ ja.
- Die Halle kennzeichnet „Lokal“ (Mehrspieler am Gerät, beliebige Zahl), der Connector
  filtert „Zu zweit“ (genau zwei). In der App müssen beide Kennzeichen nebeneinander
  stehen: Acia „Lokal“ **und** „Zu zweit“, Arschloch nur „Lokal“ (plus Online).

**Nicht angefasst:** Versionsnummer (weiter 17 / 1.16), APK und App-Assets – die werden
erst beim abschließenden Neubau erzeugt und enthalten deshalb noch Cabo ohne
Einzelspieler. Nichts veröffentlicht, nichts gepusht.

---

## 24.11.2026 · APK mit den gemeinsamen Spieldaten neu gebaut

**Anlass:** Die Spielhalle in der App soll dieselben Merks zeigen wie Website und
Connector. Der Android-Bau las den Katalog aber noch aus `spiele/spiele.html`, wo die
Liste seit dem Umbau nicht mehr steht – ohne Anpassung wäre der Bau gar nicht
durchgelaufen.

**Geändert** (`android/app/build.gradle`, beide Ablagen)
- `verifyHomepageGames` und `katalogBauen` lesen die Spieleliste, Kategorien, Texte und
  Highlights aus `spiele/spieldaten.js`; die Spielzeichen (`SPIELSYMBOLE`, `IKONEN`,
  `SYMBOLREGELN`) weiter aus `spiele.html`.
- Die native Spieleliste kennt kein `zwei`: der Bau lässt es weg und ergänzt bei
  `lokal` das Merkmal `multi` (Mehrspieler am Gerät). `katalog.json` nennt als Quelle
  jetzt `spiele/spieldaten.js`.

**Gebaut** (signierte Release-APK, gleicher Schlüssel wie zuvor: CN=TrashyGecko,
SHA-256 `73019c01…`): `gradle clean packageTrashyGeckoApk`, **BUILD SUCCESSFUL in 37 s**,
94 Spiele geprüft. APK **14.669.149 B**, md5 `c41dbb0e…`, versionCode 17 / 1.16, in
beiden Ablagen als `spiele/android/tmahjong.apk`.

**Im APK geprüft**
- `assets/www/spiele/spiele.html` lädt `spieldaten.js`; `assets/www/spiele/spieldaten.js`
  enthält Arschloch `lokal` und Acia `solo lokal zwei`.
- `assets/katalog.json`: 94 Spiele, Quelle `spiele/spieldaten.js`; Arschloch
  `[lokal, multi, online]`, Acia `[solo, lokal, multi]`, Bingo `[solo, multi, online]`.
- App-Halle im Testbrowser (aus den Build-Assets): 94 Karten, Arschloch „Lokal & Online“,
  Acia „Solo“+„Lokal“, Bingo/Galgenmännchen „Solo“+„Online“, 0 Fehler, 0 Antworten ≥ 400.
- Die Installer-Kacheln (`android/*.apk`) und die Fernbedienungs-Kachel fehlen in der
  App-Halle wie bisher (der Bau entfernt sie aus der Kopie).

**Aufgeräumt:** `app/build`, `.gradle` und `build` wieder entfernt. APK nicht auf einem
Gerät installiert (kein Gerät angeschlossen), nichts veröffentlicht, nichts gepusht.

---

## 24.11.2026 · Spielarten: eine Quelle für Spielhalle und Connector

**Befund:** Die Trennung war nicht sauber. `lokal` hieß in der Halle „Mehrspieler am
selben Gerät“ (auch Arschloch ab drei), im Connector über `lokal && multi` aber „zu
zweit“. `multi` wurde dort aus dem Online-Register dazuerfunden – deshalb brauchte Acia
ein künstliches `multi`, und Arschloch musste sein `lokal` verlieren, obwohl es lokal
spielbar ist. Dieselben 94 Spiele standen außerdem zweimal (inline in `spiele.html` und
eingebettet in `connector.html`) und waren bereits auseinandergelaufen.

**Geändert**
- Neue gemeinsame Datei `spiele/spieldaten.js` (Kategorien, Spiele, Kurztexte,
  Highlights). Die Halle lädt `spieldaten.js`, der Connector `/spiele/spieldaten.js`;
  die eingebettete Kopie im Connector ist ersatzlos weg (rund 240 Zeilen).
- Merkmale eindeutig: `solo`, `lokal` (am Gerät, beliebige Spielerzahl), `zwei` (lokal
  zu zweit), `computer`, `highscore`, `controller`. `online` und `multi` sind keine
  Merkmale mehr: online kommt ausschließlich aus `multiplayer-register.js`, multi ist
  `online || zwei`.
- Arschloch: `lokal` ohne `zwei` – die Halle zeigt wieder „Lokal & Online“, der
  Connector zählt es nicht zu „zu zweit“. Acia: `solo lokal zwei`.
- Code: Connector liest `solo`/`zwei` direkt aus den Merkmalen (kein `lokal && multi`,
  kein Register-Push mehr); die Halle nutzt `s.online`, ohne `multi`/`online` in die
  Merkmale zu schreiben, und hat den „Mehrspieler“-Zweig in `modiVon` verloren.

**Gemessen** (Playwright, 1440 × 1000): Halle 94 Karten – Solo 55, nur Online 16,
Lokal & Online 54, nur Lokal 1 (Acia), Bots 5; Bingo/Galgenmännchen/Paare
finden/Wortsuche „Solo“+„Online“, Arschloch wieder „Lokal & Online“, Risiko und Domino
„Solo“+„Lokal & Online“+„Bots“; 0 Konsolenfehler, 0 Antworten ≥ 400, `spieldaten.js`
200 (20.598 B). Connector: Alle 95, Einzelspieler 55, Mehrspieler 72, Online 71, Zu
zweit 54; Acia weiter nur lokal, Arschloch nicht unter „zu zweit“, 0 Fehler.

**Nachtrag (gleicher Tag):** Die dritte Kopie der Spieleliste – die Spielhalle in den
Web-Assets der Android-App – trug noch die alten Merks (Arschloch „lokal online multi“,
Acia „solo lokal“). Sie ist jetzt ebenfalls auf die gemeinsame Datei gezogen: in beiden
Ablagen liegen `spiele.html` und `spieldaten.js` byte-gleich zur Quelle, die App-Halle
lädt `spieldaten.js` und zeigt dieselben Kennzeichen (94 Karten, 0 Fehler, 0 Antworten
≥ 400). Alte Merk-Zeilen gibt es außerhalb von Sicherungen und Build-Zwischenständen
nicht mehr; die APK selbst wurde nicht neu gebaut.

**Nicht geändert:** Navigation, Filterbedienung, Spielseiten, Server- und Go-Code.
`spieldaten.js` `b6d27ce4…`, `spiele.html` `87cc12cc…`, `connector.html` `a9bf7a94…` –
in beiden Ablagen und in den App-Assets gleich. Nichts veröffentlicht, nichts gepusht.

---

## 24.11.2026 · Spielhalle: Merks an die geprüfte Zuordnung angeglichen

**Warum:** Die Spielkarten in `spiele/spiele.html` zeigten für Bingo, Galgenmännchen,
Paare finden und Wortsuche „Lokal & Online“, obwohl deren lokaler Modus Einzelspieler ist
(der Connector war bereits korrigiert, die Halle nicht).

**Geändert** (SPIELE-Zeilen in `spiele.html`)
- `Bingo`, `Galgenmännchen`, `Paare finden`, `Wortsuche`: `lokal online multi` →
  `solo online multi`.
- `Risiko – Welt erobern` und `Domino`: `solo` ergänzt (Computergegner) – damit stimmen
  auch diese Karten mit der geprüften Zuordnung überein.
- `Arschloch` bleibt `lokal online multi` (lokal erst ab 3 Spielern, also nicht „zu
  zweit“, aber sehr wohl am Gerät), `Acia` bleibt `solo lokal` (kein Online).

**Gemessen** (Playwright, echte Seite, 1440 × 1000, 0 Konsolenfehler): Bingo,
Galgenmännchen, Paare finden, Wortsuche zeigen **„Solo“ + „Online“** (vorher
„Lokal & Online“); Risiko und Domino **„Solo“ + „Lokal & Online“ + „Bots“**; Arschloch
„Lokal & Online“, Acia „Solo“ + „Lokal“.

**Nicht geändert:** Filter, Navigation, Spielseiten, Connector. Beide Ablagen byte-gleich
(`spiele.html` md5 `9036603d…`). Die Web-Assets der Android-App
(`spiele/android/app/src/main/assets/www/spiele/spiele.html`) tragen weiterhin die alte
Kopie und wurden nicht angefasst. Nichts veröffentlicht, nichts gepusht.

---

## 24.11.2026 · Verrücktes Labyrinth: Suchgegenstände neu gezeichnet

**Warum:** Die 25 Gegenstandssymbole waren einfarbige Silhouetten in der Tintenfarbe der
Welt und saßen durch einen halben Einheitsversatz (`translate(-.5,-.5)` in
`zeichneGegenstand`) nicht mittig auf ihrem Plättchen, sondern ragten oben links heraus.
Auf dem Feld (16–42 px) war kaum zu erkennen, wonach gesucht wird.

**Geändert** (`spiele/brettspiele/verruecktes-labyrinth/verruecktes-labyrinth.html`,
beide Ablagen byte-gleich, md5 `75b21e73a879a219827f97fd55a81925`)
- Zeichenblock der Gegenstände ersetzt: jede Figur jetzt zwei- bis dreifarbig mit dicker
  dunkler Kontur – Krone, Schlüssel, Edelstein, Münze, Buch, Ring, Glocke, Kelch,
  Heiltrank, Schwert, Schild, Hammer, Kette, Laterne, Schriftrolle, Blatt, Pilz, Eichel,
  Eule, Blume, Fisch, Feder, Stern, Schneeflocke, Eiskristall. Nur große Formen, keine
  feinen Details.
- Eine Zeile: der Versatz in `zeichneGegenstand` entfernt – die drei Aufrufer übergeben
  ohnehin die linke obere Ecke, das Symbol sitzt nun mittig im Plättchen.
- **Unverändert:** Zuordnung Aufgabe → Gegenstand (alle acht Namen je Welt), Karten, Wände,
  Figuren, Regeln, Suchlogik, Spielerreihenfolge, Mehrspieler. Diff zum alten Stand: genau
  zwei Stellen, 235 Zeilen neu / 177 alt.

**Geprüft** (Chromium/Playwright, Server 8797, 1280 × 900 und 390 × 844, alle sieben Welten)
- Feldgrößen: 7 × 7 am Handy Brett 315 px → Kachel 45 px, Plättchen 27 px, Symbol 21 px;
  7 × 7 am PC Brett 630 px → Kachel 90 px, Plättchen 54 px, Symbol 42 px; 9 × 9 am Handy
  Kachel 35 px, Plättchen 21 px, Symbol 16,4 px. Aufgaben-Chips 26 px (Canvas 56 px).
- Kontrast der dunklen Kontur zum Plättchen: **7,95 : 1** ab 24 px, **3,80 : 1** bei 20 px
  (schwächster Wert, Schneeflocke), 3,34 : 1 bei 16 px.
- Je Symbol 1–7 deutlich getrennte Farbtöne (Flächenanteil ≥ 6 %), Deckung 25–76 % der
  Plättchenfläche.
- Alle sieben Welten am PC und am Handy durchgeklickt, 9 × 9 nachgeprüft: keine Konsolen-
  oder Seitenfehler, Gegenstände sitzen mittig und heben sich vom hellen Plättchen ab.
- Kontaktbogen aller 25 Symbole in 56/32/24/20/16 px:
  `arbeit/screens/labyrinth-gegenstaende/bogen-groessen.png`; Brettfotos je Welt daneben.
- Gesichert: `arbeit/sicherung/2026-11-24-labyrinth-gegenstaende/verruecktes-labyrinth.vorher.html`
  (md5 `88712e1729e9a2d69de3f74a426b73e5`).

**Offen**
- Bei 16 px (9 × 9 am Handy) werden Feder und Kette feiner; erkennbar, aber am Rand der
  Deutlichkeit. 5 × 5 und 7 × 7 sind unkritisch.
- Nicht committet, nicht veröffentlicht (Regel „nichts committen").

## 24.11.2026 · Connector: Spielarten an den Spielen selbst geprüft

**Warum:** „Mehrspieler“ und „Online-Mehrspieler“ zeigten beide 71 Spiele. Die Zuordnung
kam aus dem `merk`-Feld und der Mitgliedschaft in `multiplayer-register.js`, nicht aus
der tatsächlichen Spielintegration.

**Geprüft**
- **Online (71):** Für jedes der 71 Register-Spiele ruft die Spielseite ihren eigenen
  Endpunkt auf (`/api/<id>` bzw. `/api/wuerfel`, 18 Würfelspiele über
  `wuerfelspiele/online.js`). Jede dieser Routen existiert als `mux.HandleFunc` in
  `connector/server.go`; `go test -run Register ./...` grün. Kein Spiel außerhalb des
  Registers ruft einen Spiel-Endpunkt auf → 71 online bestätigt, keine Streichung.
- **Zu zweit (54):** nur Spiele mit ausdrücklichem lokalen Modus für zwei Personen am
  Gerät (Spielerzahl-Auswahl ab 2, „2 Spieler“, „Lokal 2–4 Spieler“, geteilte Steuerung).
  `lokal` im `merk`-Feld bedeutete teils nur „ohne Server“.
- **Einzelspieler (55):** Solo-Modus, Gegner-KI oder „1 Spieler“ vorhanden.

**Korrigiert** (TGSpieldaten in `connector.html`)
- `Acia`: `solo lokal` → `solo lokal multi` (1/2 Spieler, geteilte Tasten) – damit erstmals
  unter „Zu zweit“ und „Mehrspieler“.
- `Arschloch`: `lokal online multi` → `online multi` (lokal erst ab 3 Spielern).
- `Bingo`, `Galgenmännchen`, `Paare finden`, `Wortsuche`: `lokal online multi` →
  `solo online multi` (der lokale Modus ist Einzelspieler, keine zweite Person).
- `Risiko – Welt erobern`: `lokal multi computer` → `solo lokal multi computer`.
- `Domino`: `lokal online multi computer` → `solo lokal online multi computer`
  („Gegen Computer“).

**Gemessen** (Playwright, lokaler Connector, 1920 × 1080, 0 Konsolenfehler): Alle 95,
Einzelspieler 55, Mehrspieler 72, Online 71, Zu zweit 54. Schnittmenge Online ∩ Zu zweit
53, nur lokal `Acia`, Vereinigung 71 + 54 − 53 = 72 = Mehrspieler. Jede Ansicht ohne
Doppelungen.

**Nicht geändert:** Navigation, Spielseiten, Server- und Go-Spielmechanik. Alle vier
`connector.html`-Kopien byte-gleich. Nichts veröffentlicht, nichts gepusht.

---

## 24.11.2026 · Connector: drei Hauptfilter statt doppelter Reiter

**Warum:** Rechts oben standen zwei Reiter („Einzelspieler“, „Multiplayer“) und darunter
noch einmal vier Filter („Alle“, „Einzelspieler“, „Online-Mehrspieler“, „Zu zweit am
Gerät“). „Einzelspieler“ kam doppelt vor, „Multiplayer“/„Mehrspieler“ waren zwei
Bezeichnungen für dasselbe.

**Was gemacht wurde**
- Die Reiterzeile samt Solo-Bereich (`#modusSolo`, `#modusMulti`, `#soloPanel`,
  `#soloListe`) ist entfernt. Es gibt genau **eine** Spieleliste rechts.
- Oben stehen nur noch drei Hauptauswahlen: **Alle Spiele** (alle), **Einzelspieler**
  (solo), **Mehrspieler** (multi = online oder zu zweit).
- Bei „Mehrspieler“ erscheinen darunter die zwei Unterfilter **Online-Mehrspieler** und
  **Zu zweit am Gerät** (mit „Mehrspieler:“ davor); außerhalb dieses Zweigs sind sie weg.
- `ueberPasst()` kennt den Filter `multi`; die Reiter-Reste (`setModus`, `soloAuswahl`)
  sind abgesichert, `sessionStorage` merkt sich nur noch „multi“.

**Gemessen** (Playwright, lokaler Connector, 1920 × 1080)
- Hauptleiste: `["Alle Spiele","Einzelspieler","Mehrspieler"]`, 0 Reiter-Knöpfe mehr.
- Karten: Alle **95** (95 verschiedene Namen), Einzelspieler **49**, Mehrspieler **71**,
  Online-Mehrspieler **71**, Zu zweit am Gerät **58**; jedes Spiel genau einmal je Ansicht.
  Die 49 + 71 ergeben 120, weil 25 Spiele beides können (z. B. Kniffel solo und zu zweit);
  die Sammlung bleibt bei 95 Spielen.
- Unterfilter erscheinen nur bei „Mehrspieler“ (`["Online-Mehrspieler","Zu zweit am Gerät"]`),
  bei „Alle Spiele“ und „Einzelspieler“ ist die Reihe leer; Highlights (6) nur bei
  „Alle Spiele“.
- Zweispaltige Anordnung unverändert: links 614 px, rechts 1306 px, Chat unten links;
  Spiel öffnet rechts (1306 × 1030), „‹ Spieleauswahl“ 131 × 33, Raum bleibt gleich.
- Handy 390 × 844: Hauptleiste 342 px breit, Unterfilter passen, `scrollWidth` 390
  (kein Querscrollen). 0 Skriptfehler auf allen Seiten.

**Offen:** Render und Pages haben den Stand noch nicht (nicht veröffentlicht), nichts
committet. Server- und Spielmechanik unangetastet.

---

## 24.11.2026 · Connector: echte zwei Spalten

**Warum:** Der Chat lief als breite Leiste über den unteren Rand der ganzen Seite, die
Spieleauswahl saß in der oberen rechten Ecke, und die linke Spalte war voller Einstellungen
und Formulare. Die Aufteilung war nicht die, die sie sein sollte.

**Was gemacht wurde**
- Der Chat liegt jetzt **in** der linken Spalte (`#chatbox` im `<main>`), nicht mehr als
  Geschwister unter beiden Spalten. Die alten Regeln (`grid-column:1 / -1`, Chatzeile im
  Raster, `58dvh`-Sondermaß) sind ersetzt, nicht überdeckt.
- Links (30–35 %, Standard 32 %): Kopfzeile mit TRASHYGECKO und Verbindungsstatus, kompakte
  Spielerliste, Namenszeile, zugeklappte Einstellungen („Grafikstil, Einladung & Vollbild“),
  darunter der Chat: Verlauf scrollt, Eingabe sitzt unten links.
- Rechts (65–70 %): Übersicht mit Filtern, Suche, sechs flachen Highlights und allen
  Kategorien; am PC sind die Kategorien offen (95 Karten, sieben je Zeile), am Handy
  zugeklappt. Beim Start einer Partie ersetzt das Spielfeld die Übersicht; „‹ Spieleauswahl“
  führt zurück, ohne den Raum zu verlassen.
- Handy: Chat direkt unter dem Kopf (44 dvh), Spieler/Name/Einstellungen darunter, Spiele
  folgen; kein horizontales Scrollen.

**Geprüft** (Playwright gegen den lokalen Connector)
- **1920 × 1080:** links 614 px (32 %), rechts 1306 px (68 %), beide 1080 px hoch; Chat 816 px
  hoch, Eingabe bei y 1015–1061 unten links; Seite scrollt nicht (Höhe 1080), kein
  Querüberlauf. Übersicht 1306 × 1032 scrollt für sich. Spiel öffnen: Feld 1306 × 1030 rechts,
  Übersicht weg, „‹ Spieleauswahl“ 131 × 33, Raumcode unverändert, Chat bleibt sichtbar.
  Zurück: Übersicht wieder da, „Verbunden“. 0 Skriptfehler.
- **1366 × 768 / 1024 × 768** (Einstellungen aufgeklappt): links 437 / 328 px, Chat 375 / 391 px,
  Eingabe vollständig sichtbar, kein Überlauf.
- **390 × 844:** Breite 390 px ohne Querscrollen; Chat bei y 46–417 mit Eingabe bei 352–398;
  Spiele darunter; Spiel öffnet 801 px hoch, Zurück-Knopf sticky, Raum unverändert,
  0 Skriptfehler.
- Chat senden, Einzelspieler-Reiter (49 Karten), Einstellungen mit „Einladung teilen“
  (115 × 28 px) funktionieren weiter.

**Offen:** Render und Pages haben diesen Stand noch nicht (nicht veröffentlicht); nichts
committet. Spielfenster und Serverlogik (Go) sind unverändert.

---

## 24.11.2026 · Connector: „Einladung teilen“ neben dem QR-Code

**Warum:** Eine Runde per WhatsApp einzuladen hieß bisher: Raumcode abtippen. Der Link lag
nirgends griffbereit.

**Was gemacht wurde**
- `connector.html`: unter dem QR-Bild stehen zwei Knöpfe nebeneinander („Adresse kopieren“,
  „Einladung teilen“); das Aufklappfeld heißt jetzt „QR-Code & Einladung“.
- `einladungTeilen()` nutzt `navigator.share` (Handy: Systemblatt mit WhatsApp, Signal,
  Mail …) und fällt sonst auf die Zwischenablage zurück („Link kopiert – jetzt in WhatsApp
  einfügen.“), zuletzt auf ein Eingabefeld. Text: „TrashyGecko – spiel mit mir! Raum CODE“.
- Der Link nimmt die erreichbare Adresse aus `/api/status` (WLAN-IP bzw. öffentliche
  Adresse) und hängt `?room=CODE` an; `localhost` wird bewusst nicht geteilt.
- Vier Kopien gleich (md5 `629f3603…`), keine Änderung an Räumen, Chat oder Spielen.

**Geprüft** (Playwright, lokaler Connector Port 8769)
- Desktop 1440 × 900 ohne `navigator.share`: Zwischenablage
  „…Raum TG-F5E5 / http://192.168.2.106:8769/connector?room=TG-F5E5“, Hinweis sichtbar,
  0 Skriptfehler, kein Querscrollen.
- Handy 390 × 844: `navigator.share` mit `{title, text, url}` samt Raumcode aufgerufen;
  beide Knöpfe in einer Reihe (65–343 px von 390), kein Querscrollen, 0 Skriptfehler.
- Einladungslink in einem zweiten Kontext geöffnet: **gleicher Raum** (TG-F5E5),
  0 Skriptfehler.
- **Offen:** kein echter WhatsApp-Aufruf auf einem Gerät (nur die System-Schnittstelle
  geprüft); Render und Pages haben den Stand noch nicht (nicht veröffentlicht).

---

## 24.11.2026 · App-Halle: eigene Spielzeichen, Kategorien-Wechsel, Spielstatus repariert

**Warum:** Die frische APK auf dem Emulator zeigte die neue Halle nur halb: eigene
Spielzeichen fehlten, unter einer Kategorie führte kein Weg zu den anderen, und im
Connector (Online-Räume) verschwanden die Reiter unter dem klebenden Kopf, während der
Spielstatus als leerer Kasten stehen blieb.

**Was gemacht wurde**
- `Spielsymbol.java` (neu): zeichnet die Zeichen der Web-Halle im 24er-Raster mit der
  Kategorie-Farbe; liest `<path>`, `<circle>` und `<rect>` über
  `androidx.core.graphics.PathParser`, also dieselben Linien wie `spiele.html`.
- `katalogBauen` liest `SPIELSYMBOLE`, `SYMBOLREGELN` und `IKONEN` aus `spiele.html` und
  schreibt je Spiel `symbol`, je Kategorie `ikon` und alle Zeichen als `symbole` in
  `katalog.json` – eine Quelle, kein zweiter Katalog.
- `SpieleActivity`: Spielzeile mit Zeichen links, Kategorie-Kacheln mit dem echten Symbol
  statt Textzeichen (⚄ ♠ ♞ …), „Andere Kategorien“ unter einer geöffneten Kategorie wie in
  der Web-Halle. Die Suche findet wieder alle Spiele; vorher blieb sie in der offenen
  Kategorie („schach“ → 0 Treffer).
- `connector.html`: `scrollZuDenSpielen()` setzt `scroll-margin-top` auf die Höhe des
  klebenden Kopfs (die Reiter lagen 29 px darunter); `statusAnzeigen()` schreibt bei leerem
  Stand wieder „Noch keine Runde offen.“ statt eines 140 px hohen leeren Kastens.
- Beide Ablagen gleich: vier `connector.html` md5 `b4b3e379…`, `Spielsymbol.java`,
  `SpieleActivity.java` und `build.gradle` je md5-gleich. Version bleibt **1.16 / 17**.

**Geprüft**
- `:app:assembleDebug` grün in beiden Ablagen; `katalogBauen`: 94 Spiele, 6 Kategorien,
  45 Spielzeichen.
- Emulator 1080 × 2340 (Debug-APK installiert): Start mit 94 Spielen und 6 Highlights;
  Würfelspiele 25 Karten mit Zeichen; „Andere Kategorien“ zeigt 5 Kacheln und wechselt
  (Würfelspiele → Kartenspiele); Suche „schach“ → 1 Treffer mit Kronen-Symbol; 0 `FATAL`,
  0 Skriptfehler.
- Connector lokal (Port 8769, Playwright): 390 × 844 Reiter-Oberkante **59 px** bei
  Kopf-Unterkante **43 px** (vorher 11 px bei 40 px, also 29 px verdeckt); Status
  „Noch keine Runde offen.“ bei 390 und 1440 (vorher leer).
- **Offen:** Render und Pages haben den Stand noch nicht (nicht veröffentlicht); die
  „eigene Runde“ im Connector nur bis zur Statuszeile geprüft, keine Partie eröffnet.

---

## 24.11.2026 · Connector: Spielebereiche und Kacheln einzeln klappbar, Chat ganz unten

**Warum:** Die Spieleübersicht im Connector war eine lange, immer offene Liste; der Chat
stand in der linken Spalte zwischen den Elementen und wurde am Handy weggeschoben.

**Was gemacht wurde**
- **Jeder Spielebereich klappt für sich** (native `<details>`, Zustand je Liste in
  `offeneBereiche`): Kopfzeile mit Pfeil, Name, Anzahl und den ersten Spielnamen
  („Würfel · Karten · …“). Standard zu; ein Bereich öffnet keinen anderen.
- **Jede Kachel klappt für sich** (`geschlosseneKarten`): zugeklappt bleiben Name und
  ein ▸-Knopf, Text/Kennzeichen/Knöpfe verschwinden. Standard offen, damit nichts
  versteckt ist; gespielte Ansicht (iframe) wird dabei nicht neu geladen.
- **Chat als eigene Vollbreiten-Zeile** unten im `.zentrale-geruest` (Desktop Rasterzeile 2,
  `--chat-hoehe:clamp(220px,30vh,300px)`; am Handy nach den Spielen). `#spielGriff`
  endet über dem Chat.
- **Kein Seitensprung mehr** bei neuen Nachrichten: nur noch der Chatverlauf rückt nach
  (`if(atBottom || !chatSprungBereit)…`); `chatbox.scrollIntoView` entfällt. Beim ersten
  Aufbau nach dem Beitritt steht der Verlauf immer am Ende.

**Geprüft** (Chromium, Testinstanz 8767, alle vier Kopien byte-gleich,
md5 `7dea5c56664c7a76f955b678ee84abf9`)
- Alle 8 Inline-Skripte `node --check` ohne Fehler.
- Bereiche unabhängig: Würfel und Karten gleichzeitig offen, Schließen des einen lässt den
  anderen offen; Suche „risiko“ öffnet nur Brettspiele (1 Karte), Leeren stellt die vorher
  offenen Bereiche wieder her; Zustand über Größenänderung erhalten.
- Kachel 174 px → 40 px beim Zuklappen, Nachbarkachel bleibt 174 px.
- Desktop 1280 × 800: links 410 × 560, rechts 870 × 560, Chat 1280 × 240 bei y 560–800,
  Seite ohne Scrollen; mit offenem Spiel Fenster 870 × 508, Chat darunter.
- Handy 390 × 844: Chat beginnt bei y 1286 unter der Spielauswahl (372 × 287),
  `scrollWidth` 374 → kein Querscrollen, Kategorie-Kopf 51 px.
- Chat: initial `scrollTop` 48,8 bei max 49; Testkasten 60 px: nach neuer Nachricht
  `top` 234 = `max` 234, `scrollY` bleibt 0.

**Offen:** keine zweite echte Spielersitzung; eine eingehende Fremdnachricht wurde nicht
live erzeugt (nur über den bestehenden Verlauf geprüft). Sicherung vor der Änderung:
`sicherung/2026-11-24-connector-klapp/`.

---

## 24.11.2026 · Spielhalle: eigene Spielsymbole, Kategorien-Navigation, sichtbare eigene Runde

**Warum:** Die Spielkarten zeigten nur Buchstabenkürzel, zwischen Kategorien führte
nur der Umweg über den Start, und eine eröffnete Runde war nach dem Schließen des
Tisches nicht mehr sichtbar.

**Was gemacht wurde**
- **Eigene Spielsymbole für alle 94 Karten:** 45 SVG-Zeichen (Würfel, Karten, Krone, Mühle, Bombe,
  Kiste, Juwel, Alien, Galgen …), Zuordnung über Pfadregeln, Rückfall Kategorie.
  Gemessen: 94/94 Karten mit Symbol, Würfelspiele 25/25.
- **Kategorien-Navigation:** unter jeder Kategorie die fünf anderen als Kacheln mit
  Symbol, Name und Spielzahl; gemessen 5 Ziele, beschriftet.
- **Eigene Runde bleibt sichtbar:** Leisten-Knopf „Runde: 10.000“, Panel-Block
  „Offene Runde“ mit Tisch öffnen / Zurückziehen (`/api/angebot` `zurueck`);
  „Partie fortsetzen“ überlebt das Schließen (`letzterTisch`, `tg-halle-tisch`).
- **Textfehler behoben:** `el(tag, klasse)` nimmt keinen dritten Parameter – Stern,
  ›-Pfeil, „Raum + Chat“ und die Kategorien-Beschriftungen waren leer. Jetzt
  `textContent`; Zurück-Pfeil der Suche ist ein echtes `←`.

**Geprüft** (Chromium, Connector 8768, zwei Ursprünge: 127.0.0.1 und
192.168.56.1)
- Raum TG-6348, Anna + Gast YU1, je 2 Spieler; 10.000 eröffnet, Angebot
  angenommen, beide bereit, Host startet; Chat in beide Richtungen (9 Beiträge);
  0 Konsolenfehler.
- Nach dem Schließen: „Runde: 10.000“ in der Leiste, Panel „Deine Runde wartet …“,
  Tisch öffnet ohne zweite Runde.
- Handy 390 × 844: Start/Kategorie/Highlights je 0 px Querscrollen.
- `go test -run Register ./...`: PASS (71 Register-Einträge, 70 Katalogspiele,
  24 Spiele ohne Online-Teil, `malen` ist Werkzeug).
- APK neu gebaut (Assets enthalten `halle.css`, `halle-online.js`), Version
  unverändert 1.16 / Code 17.

**Nachtrag: Bereitschaft und Start in der Halle, drei weitere Spiele geprüft**
- Die Tisch-Kopfzeile hat **Bereit** und **Spiel starten** (Gastgeber, erst wenn
  alle bereit sind). Dieselben Endpunkte wie die Spielseite (`action=ready/start`),
  keine zweite Lobby.
- **Galgen/Kniffel/Schach** vollständig über die Halle gespielt (Raum TG-96F2,
  Loopback + LAN): Beitritt, Bereit, Start, laufende Partie sichtbar, 0 Fehler.
- Behoben: laufender Zustand `spiel` wurde nicht als „Partie läuft“ erkannt;
  offenes Tisch-Angebot blieb nach dem Start stehen; Schach (ohne `players`)
  meldete fälschlich „Warte auf Mitspieler“.
- Handy 390 px mit den neuen Knöpfen: 0 px Querscrollen in Kopfzeile und Seite.

**Offen:** vier Spiele durchgespielt (10.000, Kniffel, Galgen, Schach), 66 weitere
ungetestet; zwei Ursprünge auf einem Rechner statt zwei echter Geräte; kein
Freundesystem im Connector (Anwesenheit nur im Raum); Start aus der Spielseite
räumt das Hallen-Angebot nicht auf (Connector hat keine gemeinsame Statusabfrage).

---

## 24.11.2026 · Web-Spielhalle: neue Optik, neue Highlights, durchgängiger Online-Ablauf

**Warum:** Die Übersicht war grau und schlicht; Räume, Einladungen und Chat lagen nur im
Connector, nicht in der Spielhalle.

**Highlights**
- Flipper und Quizshow sind aus den vier Highlights entfernt, **Mahjong** und **10.000**
  sind neu. Bilder: `spiele/bilder/highlights/{mahjong,zehntausend}.jpg` (1392 × 752,
  91 KB / 86 KB), erzeugt im Stil der vorhandenen zwei Bilder.
- Jede Highlight-Karte zeigt Bild, Kategorie, Kurztext, Modus-Marken und einen direkten
  Einstieg; 10.000 trägt zusätzlich „Online 2–6“. Der Pfad `/spiele/bilder/highlights/…`
  bleibt im Katalog und damit im Android-Bau unverändert.

**Optik** (`spiele/halle.css`, ersetzt das bisherige Inline-CSS in `spiele.html`)
- Dunkler Grund aus drei Radialverläufen, feinem Raster und Vignette; Kategorie-Farben
  wirken als Tint, Rand, Symbol und Akzentkante. Große Bildkarten, sechs
  Kategorie-Kacheln, kompakte Spielkarten mit Modus-Chips.
- Hover/Fokus-Animationen, gestaffelter Auftritt (`--i`, 26 ms), `prefers-reduced-motion`
  schaltet alles ab, `prefers-contrast:more` hebt Texte an, Fokus bleibt sichtbar.
- Raster: Handy 1 Spalte Karten/2 Kategorien, ab 560 px 2, ab 940 px 3, ab 1240 px 4
  Karten und 6 Kategorien. Kopfzeile und Online-Leiste sind auf 390 px einzeilig.
- Mädchen-/Buben-Stil bleiben über die vorhandenen `--tg-*`-Tokens erhalten; die dunklen
  Kartenspiele wurden nicht angefasst.

**Online** (`spiele/halle-online.js`, neue Schicht – kein zweiter Connector)
- Identität wie im Connector: `tg-player-id`, `tg-player-name`, `token.js`; ohne Namen
  entsteht „Gast XYZ“, änderbar im Panel.
- „Mit Freunden spielen“ legt einen **eigenen Raum** an (`POST /api/rooms` mit
  `neuerRaum`), Einladungen laufen über `?raum=CODE`, Raumcode und QR-Bild
  (`/api/qr?url=`). Betreten per Code oder Link; Anwesenheit über `/api/presence` und
  `/api/players` (20 s / 7,5 s).
- Offene Runde über `/api/angebot`: Host „bieten“ (Tischspiele erzeugt der Server),
  Gäste „annehmen“ oder „ablehnen“; eigene Runde lässt sich wieder öffnen. Ein
  laufendes Spiel wird über `/api/<spiel>` erkannt und fortgesetzt statt doppelt eröffnet.
- Der Tisch läuft im Overlay als **dieselbe Spielseite** (`?room&name&id&v`); Bereitschaft
  und Start bleiben dort, wo die Lobby schon war. Die Kopfzeile zeigt Spielerzahl,
  Bereitschaft und Status (alle 2,5 s).
- Chat: `/api/messages`, als Dock rechts (Desktop) bzw. Blatt unten (Handy), ungelesene
  Beiträge am Chat-Knopf, Einladungszeilen mit „Mitspielen/Ablehnen“.
- Reconnect: 409 → gleicher Platz neu, 403 → neue Kennung (höchstens alle 30 s),
  404 → neuer Raum; gespeicherter Raum und Tisch werden beim Laden wieder aufgenommen.
- Nur Spiele aus `multiplayer-register.js` bekommen den Online-Einstieg; alle anderen
  starten unverändert sofort. Der Filter „Online spielbar“ zeigt die echten Fälle.

**Geprüft** (Chromium/Playwright, zwei getrennte Kontexte, Connector lokal Port 8769)
- Startseite 1440 × 900: 4 Highlights, 6 Kategorien, keine Konsolenfehler.
- Raum TG-5BD3 über „Mit Freunden spielen“, Gast über `?raum=TG-5BD3`: beide sehen
  2 Spieler. 10.000 → „Online-Runde“ → Tisch; Gast nimmt das Angebot an; beide „Bereit“,
  Host startet. Danach in beiden Köpfen „Partie läuft · 2 Spieler“.
- Chat in beide Richtungen („Hallo von A …“, „Hallo von B …“), 10 Beiträge im Verlauf.
- Handy 390 × 844: gleicher Ablauf, Chat als Blatt unten (62 % Höhe), Screenshots geprüft.
- Neuladen des Gasts → „Partie fortsetzen“ → laufende Partie wieder da. Schließen und
  erneut 10.000 wählen → „Eigene Runde öffnen“, **1 am Tisch** (keine zweite Runde).
- „Jetzt spielen“ führt weiterhin auf `…/zehntausend.html` **ohne** `?room`.
- `node --check` für `halle-online.js` und die ausgelagerte Inline-JS: grün.

**Android**
- `:app:katalogBauen` neu: 94 Spiele, 6 Kategorien, Highlights Mahjong/10.000/Risiko/
  Labyrinth. `:app:packageTrashyGeckoApk`: 14.690.966 B,
  md5 `0c0450069e852ff4151b9e79e7658ddb`; APK enthält `halle.css`, `halle-online.js` und
  beide neuen Bilder. Version bleibt **1.16 / Code 17** (Änderung nur mit Rückfrage).
- Die App selbst wurde in dieser Runde nicht neu installiert/getestet.

**Offen**
- Ende-zu-Ende nur mit 10.000 geprüft; die übrigen Online-Spiele des Registers nutzen
  denselben Weg, wurden aber nicht durchgespielt.
- Zwei Browser-Kontexte auf einem Rechner, **kein** echtes zweites Gerät im WLAN.
- Bereitschaft/Start bleiben bewusst in der Spielseite; die Halle zeigt nur den Stand.
- Ohne laufenden Connector bleibt die Halle beim Einzelspieler-Teil.

---

## 23.11.2026 · App-Spielhalle: Startseite, Kategorien und Suche aus `katalog.json`

**Warum:** Die Web-Spielhalle hatte Kategorien, Highlights und eine Suche, die App zeigte
weiterhin die flache Liste mit 94 Einträgen.

**Was gemacht wurde** (D-26)
- `katalogBauen` überträgt jetzt auch die Kurzbeschreibungen (`TEXTE`), die
  Kategorie-Texte samt `akzent`-Farbe und die vier `HIGHLIGHTS` nach `katalog.json`
  (94 Spiele mit Text, 6 Kategorien, 4 Highlights). Die Zeilenform in `spiele.html`
  bleibt unverändert; Highlight-Pfade werden relativ zu `spiele/` aufgelöst.
- `SpieleActivity` (614 → **1055** Zeilen) hat drei Ansichten: Start mit vier
  Highlight-Karten (2 Spalten, ab 900 dp 4) und sechs Kategorie-Kacheln (2/3/6 Spalten je
  Breite), eine Kategorie mit Beschreibung und **nur** ihren Spielen sowie die
  Trefferliste für Suche/Filter, nach Kategorie gruppiert. Jede Spielkarte zeigt Name,
  Kurztext und vorhandene Modi; die Suche prüft Name, Suchwörter, Text und
  Kategoriename (Umlaute egal), die Chips bleiben unverändert (UND).
- Zurück führt in dieser Reihenfolge: Filter löschen → Kategorie verlassen → Activity
  beenden; der Pfeil in der Kopfzeile nimmt denselben Weg.
- Prozess-Tod: `onSaveInstanceState` sichert zusätzlich Ansicht und offene Kategorie
  (`ZUSTAND_ANSICHT`, `ZUSTAND_KATEGORIE`); Suche, Merkmale und Scroll bleiben wie
  bisher. Die Reihenfolge `stelleWieder` → einmal `zeichne` ist unverändert.
- Highlight-Bilder kommen aus `assets/www/spiele/bilder/highlights/*.jpg`, verkleinert
  gelesen (`inSampleSize = 3`, ~464 × 251); Karten einer Rasterzeile bekommen die Höhe
  der höchsten Karte.

**Geprüft (Emulator 1080 × 2340, Debug-APK)**
- Start: „94 Spiele“, 7 Filter-Chips, 4 Highlights (Karten einer Zeile gleich hoch),
  6 Kacheln mit 25/11/22/15/20/1 Spielen, keine abgeschnittenen Texte.
- „Würfelspiele“ → 25 Einträge mit Beschreibung; Suche „kniffel“ → „1 Treffer“.
- Prozess-Tod in der Kategorie bei Scroll 1014: `am kill` (PID 18129 → leer), Start
  über die Übersicht → PID 18451, Log „Ansicht kategorie (wuerfel), Scroll 1014“,
  Liste an derselben Stelle. Suche „kniffel“ überlebte ebenso („1 Treffer“).
- Prozess-Tod mit dem endgültigen Bau: Suche „bejeweled“ gesetzt, `am kill`
  (PID 4883 → leer), Task zurückgeholt → PID 5308, Log „Zustand wiederhergestellt:
  „bejeweled“, 0 Merkmale, Ansicht start“, „1 Treffer“.
- Dialog „Lokal / offline“ / „Online-Spiel“ erscheint, „Lokal / offline“ öffnet
  `GameActivity`; Bejeweled lädt („Neues Spiel“, Anleitung, Spielfeld), der Prozess
  bleibt, 0 `FATAL` im Logcat. (Beim ersten Versuch verschwand der Emulator selbst –
  Adb „device offline“, kein App-Absturz; nach einem Neustart lief es.)
- `:app:assembleDebug` und `:app:packageTrashyGeckoApk` grün; `katalog.json` in beiden
  Ablagen md5-gleich (`440ee0e5…`), ebenso `SpieleActivity.java` (`005b4b3d…`) und
  `build.gradle` (`9a5f009d…`).
- App **1.16** (versionCode 17), signiert (CN=TrashyGecko), **14.475.624 B**;
  `tmahjong.apk` und `TrashyGecko.apk` in beiden Ablagen md5-gleich (`ed23415b…`).

**Grenze:** Keine ganze Partie durchgespielt; Mehrspieler nur bis zum Modus-Dialog
geprüft (die Spielseiten selbst sind unverändert). Nicht committet, nicht veröffentlicht.

## 23.11.2026 · Spielhalle: Kategorien, Highlights und Suche statt langer Liste

**Warum:** Die Spieleübersicht war eine einzige, immer länger werdende Liste mit
aufklappbaren Abschnitten. Am Handy musste man weit scrollen, Kategorien und Spielmodi
waren nicht auf einen Blick zu erkennen.

**Was gemacht wurde** (D-25)
- `spiele/spiele.html` neu aufgebaut: Startseite mit vier großen Highlight-Karten, sechs
  Kategorie-Kacheln (Würfel, Karten, Brett, Denken, Arcade/Geschicklichkeit,
  Quiz/Partyspiele), zentrale Suche und je Kategorie eine eigene Ansicht mit kompakten
  Spielkarten (Name, Symbol, Kurzbeschreibung, vorhandene Spielmodi, „Spielen“).
- Ansichten über Hash-Adressen (`#/k/wuerfel`, `#/highlights`, `#/werkzeuge`); der
  Zurück-Pfeil bleibt beim Scrollen oben stehen. Keine waagerechte Scrollleiste.
- Spielmodi kommen ausschließlich aus den vorhandenen Merkmalen (`solo`, `lokal`,
  `online`, `computer`); nicht vorhandene Modi werden nicht angezeigt. Die Auswahl
  passiert weiterhin erst im Spiel.
- Kategorie-Zuordnung gestrafft: Action, Jump & Run und Abenteuer gehören jetzt zu
  „Arcade und Geschicklichkeit“, die Scherzshow zu „Quiz und Partyspiele“. Das zweite
  „Solitär“ heißt zur Unterscheidung „Steine-Solitär“ (gleiche URL, gleiches Spiel).
- Katalogzeilen in `SPIELE`/`WERKZEUGE` unverändert im Format – `katalogBauen` und
  `homepageKopieren` lesen sie weiterhin. Neu sind nur `TEXTE` (Kurzbeschreibungen),
  `HIGHLIGHTS` und vier Vorschaubilder unter `spiele/bilder/highlights/`.
- Nichts an Spielen, Connector, Spielständen, URLs oder der Grafikstil-Auswahl geändert.

**Tatsächlich getestet** (Playwright gegen `python -m http.server`, Desktop-Chromium)
- 360×740: 6 Kategorie-Kacheln, 4 Highlights, Startseite **1274 px** hoch (vorher 1809),
  keine waagerechte Scrollleiste; Würfelspiele 25 Karten, davon 16 mit Einzelspieler.
- Suche: „mau“ → Mau Mau, „kniffel“ → Kniffel, „muhle“ → Mühle (ohne Umlaut),
  „qqqq“ → Hinweis; Leeren stellt die Übersicht wieder her.
- Modi: UNO zeigt „Mehrspieler: lokal & online“, Solitär nur „Einzelspieler“.
- 1280×800: 6 Kacheln und 4 Highlights je in einer Reihe, Spielkarten in 4 Spalten.
- Alle **94** Spiele per HTTP **200**; UNO, Schach, Jule und Pong laden ohne Konsolen-,
  Seiten- oder Ladefehler. Grafikstil „Mädchen“ greift weiterhin.
- `:app:katalogBauen` und `:app:verifyHomepageGames`: **94 Spiele, 6 Kategorien**,
  `BUILD SUCCESSFUL in 41s`; die APK-Kopie enthält die beiden Installer-Zeilen nicht mehr.

**Offen**
- Nur Browserprüfung bei zwei Größen; kein Emulator- oder Gerätetest in dieser Runde.
- Die zwei zufallsabhängigen Go-Tests (Quiz-Antwortzitate, Zehntausend-Wurf) bleiben
  sporadisch – unabhängig von dieser Arbeit.

---

## 23.11.2026 · Etappe 3, Schritt 11: Homepage-Seiten liegen im APK

**Warum:** Im APK lag nur `index.html` (als `homepage.html`). Fast jede Kachel der
Startseite (Tagebuch, Bücher, Rezepte, Tipps …) lief deshalb auf 404 – die Seite war ohne
Netz weitgehend tot.

**Was gemacht wurde** (D-24)
- `app/build.gradle`: neue Liste `homepageSeiten` mit den 19 Ordnern, die `index.html`
  verlinkt; `homepageKopieren` kopiert sie samt `main.css` nach `www/<ordner>/`. Draußen
  bleiben der alte Seiten-Spiegel unter `css/` (in `.gitignore`, u. a. mit alten
  Zugangsdaten), `musik/.backup-*` und `tagebuch/Textdokument (neu).txt`.
- Sonst keine Codeänderung: `HomepageActivity`, `WebNavigation` und `HomepageAssets`
  bleiben, weil interne Links schon im WebView bleiben.

**Tatsächlich getestet** (Emulator Android 14, Flugmodus, `Active default network: none`)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 18s`; APK 11.460.085 → **11.959.743 B**
  (+499.658 B), `assets/www` 13 → 14 MB, **303** HTML-Seiten im Paket.
- 23 lokale Adressen aus der Seite heraus abgerufen (`fetch`): alle **200**.
- Echte Klicks: Tagebuch, Rezepte, Bücher, Tipps öffnen (Titel „Tagebuch“, „Rezepte“,
  „Bücher“, „Tipps“), zwei Ebenen tiefer `tipps/hoersturz.html` („Hörsturz (Gehörsturz)“)
  und `seiten/programmieren.html` → `../css/css.html` („CSS“).
- 0 `Seitenfehler` im Protokoll, 0 kaputte Bilder, 0 `FATAL EXCEPTION`.

**Offen**
- Neun Ziele gibt es auch in der Quelle nicht (Beispielpfade der Anleitungen:
  `bilder/beispiel.jpg`, `home.png`, `html/home.png`, `css/bild.gif`, `css/bild.png`,
  „Adresse des Bildes“) – die bleiben tot, im Browser wie im APK.
- Android 9 auf Gerät A (gesperrt) und eine Sichtprüfung des Seitenbildes.

---

## 23.11.2026 · Etappe 3, Schritt 10: Online-Kachel sagt die Wahrheit

**Warum:** Ohne Netz zeigte die Online-Kachel Chromiums englische Fehlerseite
(`net::ERR_INTERNET_DISCONNECTED`), während die Statuszeile „Online-Connector · bestehender
Server“ behauptete. Ein Nutzer ohne Netz sah also eine englische Technikmeldung plus eine
falsche Zusage.

**Was gemacht wurde** (D-23)
- `ConnectorActivity` (458 → **494**): `istOnline()` prüft vor dem Laden das aktive Netz;
  ohne Netz wird gar kein WebView gebaut.
- `zeigeOnlineFehler(...)` verwirft den WebView und zeigt „Online-Connector nicht erreichbar“
  mit deutschem Erklärungssatz und Knopf „Erneut versuchen“.
- Statuszeile folgt dem Verlauf: „wird geladen …“ → „Server erreichbar“; bei Fehler
  „Offline · keine Verbindung“ bzw. „Online-Connector · Server nicht erreichbar“.
- `onReceivedError` im Hauptframe greift auch, wenn das Netz da ist, der Server aber nicht
  antwortet.

**Tatsächlich getestet** (Emulator Android 14)
- Flugmodus (`Active default network: none`): Kachel → deutsche Meldung mit Knopf, Status
  „Offline · keine Verbindung“.
- Flugmodus aus, Knopf gedrückt → Seite lädt, Status „Online-Connector · Server erreichbar“.
- Server weg (Testbau mit unerreichbarer Adresse): Status „Server nicht erreichbar“, gleiche
  Meldung; Testbau zurückgebaut, Endfassung erneut im Flugmodus geprüft.
- 0 Abstürze in allen Läufen.

**Offen**
- Gerätetest Android 9 für diesen Weg (Gerät gesperrt, siehe Schritt 9) und eine Sichtprüfung
  der geladenen Connector-Seite (WebView-Inhalt steht nicht im UI-Baum).

---

## 23.11.2026 · Etappe 3, Schritt 9: Startansicht und Spieleliste teilen Farben und Bausteine

**Warum:** Die native Startansicht hatte ihr eigenes Farbbild (leicht andere Grün- und
Grautöne als die Spieleliste) und eigene Kopien von `dp` und `flaeche`; `onResume` prüfte
außerdem `status` auf null, obwohl die Ansicht in `onCreate` immer entsteht.

**Was gemacht wurde**
- Neu `Ansicht.java` (53 Zeilen): `BG`, `KARTE`, `RAND`, `GOLD`, `TEXT`, `MUTED` an einer
  Stelle, dazu `flaeche(farbe,radius)`, `flaeche(farbe,radius,rand,breite)`, `text(...)` und
  `dp(context,wert)`.
- `MainActivity` (193 → **173**): benutzt nur noch `Ansicht`; eigene Farbkonstanten,
  `text()`, `flaeche()`, `dp()` und der `status != null`-Test sind weg. Die Netzstatus-Logik
  ist unverändert, nur mit einer Notiz zur API-Lage.
- `SpieleActivity` (630 → **614**): Farben und Helfer ebenfalls aus `Ansicht`, lokal bleibt
  nur der Chip-Füllton. `CHIP_RAND` ist entfallen, die Chip-Kante nimmt `Ansicht.RAND`.
- Startseite nutzt jetzt dieselben Töne wie die Spieleliste (Hintergrund `#0B241D` statt
  `#07140F`, Text `#E9F1EB`, Grau `#9AA7A0`, Karte `#123028`, Kante `#2B4C40`).

**Tatsächlich getestet** (Emulator Android 14; Android 9 nur Statuslogik)
- Build `BUILD SUCCESSFUL in 10s`; Startansicht mit gleichen Maßen (Titel y 241, Kacheln
  553/864/1175/1486), Status „WLAN · online bereit“.
- Spieleliste: „94 Spiele“ → „Mehrspieler“ „70 von 94“ → „Zurücksetzen“ „94 Spiele“,
  `Liste: 94 Zeilen in 39 ms`, 0 Abstürze.
- Android 9 (Gerät A, API 28): Testbau mit eigener Paket-ID, Gerät gesperrt, deshalb über
  eine Protokollzeile gemessen – `Netzstatus: WLAN · online bereit`, keine Ausnahme.
  Testbau wieder deinstalliert, die release-signierte 1.15 blieb unangetastet.

**Offen**
- Startansicht und Offline-Zweig auf Android 9 mit Bildschirm (Gerät gesperrt; Flugmodus
  dort nicht geschaltet).

**Nächster konkreter Arbeitsschritt**
1. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.
2. Danach Statusanzeige mit Serverprüfung und Einstellungen.

---

## 23.11.2026 · Etappe 3, Schritt 8: „Extras → Homepage“ mit der echten Startseite

**Warum:** D-12 verlangt die echte Website als eigenen Eintrag unter Extras; bis hierher
gab es den Eintrag nicht.

**Was gemacht wurde**
- `homepageKopieren` kopiert die echte `index.html` als `www/homepage.html`; `main.css`
  liegt dort schon. `www/index.html` bleibt die alte App-Startseite (`app.html`) als
  Rückweg aus Spielseiten – die beiden Namen kollidieren also nicht.
- Neu `HomepageActivity` (142 Zeilen): WebView über denselben `HomepageAssets`-Lader,
  Vollbild, `WebBasis`-Einstellungen und die Link-Regel `WebNavigation.pruefe(uri, true)`
  („Spiele“ öffnet die native Liste, fremde Adressen den Browser); Zurück geht erst im
  Verlauf zurück, sonst in die App.
- `ConnectorActivity.showExtras` (442 → **453** Zeilen) hat die Karte „Homepage“; sie
  erscheint nur, wenn `www/homepage.html` im Paket liegt (`isAssetAvailable`), sonst steht
  dort ein Hinweis. Die Fernbedienungs-Prüfung nutzt jetzt denselben Helfer.
- Manifest: `HomepageActivity` mit denselben `configChanges` wie die übrigen Activities.

**Tatsächlich getestet** (Emulator Android 14, Debug-Build)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 17s`; `assets/www/homepage.html`
  md5-gleich zu `index.html` (`08e7a1b8…`), `main.css` 17.579 B daneben, `www/index.html`
  (App-Start) unverändert 4.249 B.
- Extras zeigt die Karte „Homepage“ (Titel `[94,1145]`, Knopf „Homepage öffnen“ `[94,1383]`).
- Tipp auf den Knopf: `HomepageActivity` oben, Log
  „Homepage geladen: https://appassets.androidplatform.net/homepage.html“; Inhalt:
  Titel „TrashyGecko“, „Google Search“, Links (ChatGPT, deepseek, Gemini …), Uhr
  „18:51:14 Uhr“. Einziger Fehlzugriff: `www/favicon.ico` (harmlos).
- Zurück-Taste: wieder `ConnectorActivity` (Extras-Kopfzeile), 0 `FATAL EXCEPTION`.

**Offen / nicht getestet**
- Die weiteren Seiten der Website (Tagebuch, Bücher, Rezepte …) liegen seit Schritt 11
  (D-24) im Paket; dieser Abschnitt bleibt als Stand von Schritt 8 stehen.
- Das Google-Formular zielt auf `http://` und erzeugt eine Mixed-Content-Warnung im Log;
  ob die Suche danach funktioniert, ist nicht geprüft.
- Echtes Android 9.

**Nachtrag (Online-Kachel, 18:52–18:59):** Die Kachel öffnet
`https://trashygecko.onrender.com/connector` (Ziel über die WebView-Debug-Schnittstelle
ausgelesen). Zurück führte vorher zuerst auf die Connector-Startseite und erst mit dem
zweiten Druck auf den nativen Start; `ConnectorActivity` (453 → **458** Zeilen) merkt sich
jetzt in `vonAussen`, ob sie über eine Kachel geöffnet wurde, stoppt das lokale Netz und
schließt dann. Ohne Netz zeigt der WebView die englische Chromium-Fehlerseite
(`net::ERR_INTERNET_DISCONNECTED`), die Statuszeile bleibt bei „bestehender Server“ – ein
deutscher Fehlertext fehlt noch.

**Nächster konkreter Arbeitsschritt**
1. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.
2. Danach Statusanzeige mit Serverprüfung und Einstellungen.

---

## 23.11.2026 · Etappe 3, Schritt 7: native Startansicht, Zustandssicherung vereinfacht

**Warum:** Der Start der App war wieder eine Webseite (`app.html` im WebView) – genau der
Eindruck, den D-12 vermeiden will. Außerdem hatte die Zustandssicherung aus Schritt 6 einen
Fehler: `onCreate` zeichnete die Liste **vor** `stelleWieder`. Wurde nur ein Merkmal ohne
Suchtext gesichert, kam die Liste ungefiltert zurück, während der Chip schon gold war.

**Was gemacht wurde**
- `MainActivity` (134 → **193** Zeilen): kein WebView mehr. Marke, Titel, Unterzeile und
  vier Kacheln (Spiele, Lokales Netz, Online-Räume, Extras) als Views, im Farbbild von
  `start/app.html`; jede Kachel öffnet `SpieleActivity` bzw. `ConnectorActivity` mit
  `EXTRA_INITIAL_SCREEN = local | online | extras`.
- Kopfzeile zeigt den **Netzstatus des Geräts**: „offline spielbar“, „WLAN · online
  bereit“, „Mobilfunk · online bereit“ (neu gelesen in `onResume`). Er sagt nur, ob ein
  Netz da ist – nicht, ob der Render-Server antwortet.
- `start/app.html` bleibt im Paket, aber nur noch als `index.html` für „zurück zum Start“
  aus einer Spielseite; die App selbst öffnet sie nicht mehr.
- `SpieleActivity` (619 → **630** Zeilen): `stelleWieder` stellt erst her und zeichnet dann
  **einmal**; während `suchfeld.setText` läuft, ist `stelleWiederLaeuft` gesetzt, damit der
  Textwächter nicht ein zweites Mal baut. Nullprüfungen im Bundle-Weg sind entfallen (die
  Ansichten existieren, wenn gesichert wird), unbekannte Merkmale werden weiter verworfen.

**Tatsächlich getestet** (Emulator Android 14, Debug-Build)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 10s` (exit 0).
- Start nativ: „TRASHYGECKO“ + „WLAN · online bereit“, Titel, vier Kacheln mit
  Beschreibung, Fußzeile; alles innerhalb 1080 × 2340, 0 `FATAL EXCEPTION`.
- Kacheln: „Extras“ → Connector-Seite „Extras“, „Lokales Netz“ → „Lokales Netzwerk“,
  Zurück führt jeweils auf den nativen Start.
- **Chip ohne Suchtext** (der vorher fehlerhafte Fall): „Mehrspieler“ → „70 von 94“, drei
  Wischgesten (Scroll 5237) → Home → `am kill` (PID 9573 → leer) → Übersicht: neuer
  Prozess 9782, Log „Zustand wiederhergestellt: „“, 1 Merkmale, Scroll 5237“,
  `Liste: 70 Zeilen in 103 ms`; `uiautomator`-Ausdruck vor/nach **identisch**, Chip
  `selected="true"`, „70 von 94“, 0 Abstürze.
- Suchtext: „a“ → „55 von 94“ → `am kill` → neuer Prozess 10044, „55 von 94“, Feld zeigt
  „a“; danach „b“ getippt → „5 von 94“ (Textwächter wieder scharf). „Zurücksetzen“ →
  „94 Spiele“, `Liste: 94 Zeilen in 60 ms`.

**Offen / nicht getestet**
- Online-Kachel (öffnet nur die Connector-Seite gegen `trashygecko.onrender.com`).
- Server-Erreichbarkeit und „erneut versuchen“ – der Netzstatus prüft nur das Gerät.
- „Extras → Homepage“ aus D-12 fehlt noch; dafür müsste die echte Website in die Assets.
- Echtes Android 9, Wiederherstellung bei anderer Bildschirmhöhe.

**Nächster konkreter Arbeitsschritt**
1. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.
2. Danach Statusanzeige mit Serverprüfung und Einstellungen.

---

## 23.11.2026 · Etappe 3, Schritt 6: Zustand überlebt den Prozess-Tod

**Warum:** Die Drehung lässt die Activity stehen (D-21), ein beendeter Prozess nicht. Bis
hierher sicherte `SpieleActivity` nichts: nach „Aktivitäten nicht behalten“ oder einem
Prozess-Tod im Hintergrund wären Suchtext, Filter und Scrollposition weg.

**Was gemacht wurde**
- `SpieleActivity` (560 → **619** Zeilen): `onSaveInstanceState` legt Suchtext
  (`suche`), die gewählten Merkmale (`merkmale`, `StringArrayList`) und
  `rollen.getScrollY()` (`scroll`) in das Bundle.
- `stelleWieder(bundle)` läuft in `onCreate` nach `ladeKatalog()` und `zeichne()`:
  Merkmale werden nur übernommen, wenn sie in `MERKMAL_FILTER` stehen (keine Chips aus
  alten Ständen), der Suchtext geht über `suchfeld.setText(...)` durch dieselbe
  `TextWatcher`-Kette wie eine Eingabe, `chipAussehen()` malt die Auswahl an.
- Die Scrollposition wird erst nach dem ersten Layout gesetzt
  (`ViewTreeObserver.addOnGlobalLayoutListener`, danach wieder abgemeldet) – ein
  `post()` davor liefe gegen eine Inhaltshöhe von 0 und würde auf 0 geklemmt.
- Ohne Bundle bleibt alles beim Anfang; die Drehung braucht den Weg nicht (D-21).

**Tatsächlich getestet** (Emulator Android 14)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 10s`.
- Zustand gesetzt: Suche „a“, Chip „Mehrspieler“ `[GEWÄHLT]`, „37 von 94“, Liste drei
  Wischgesten nach unten (Scroll 3210).
- Home-Taste → `adb shell am kill de.trashygecko.tmahjong` → **Prozess weg** (PID 8844),
  Rückkehr über die Übersicht (Recents): **neuer Prozess 9354**, neuer `onCreate`
  („Katalog: 94 Spiele“ im Log), Logzeile
  „Zustand wiederhergestellt: „a“, 1 Merkmale, Scroll 3210“.
- Danach dieselben Zeilen an denselben Stellen wie vorher: „Risiko – Welt erobern“
  `[83,2268][997,2327]`, „Halma“ `[83,1882]`, „Go“ `[83,1689]`, „Kalaha“ `[83,2075]`;
  Zähler „37 von 94“, „Mehrspieler“ `selected="true"`, Suchfeld „a“.
- Weiterfilterbar: „Zurücksetzen“ → Suchfeld leer, kein Chip gewählt, „94 Spiele“
  (`Liste: 94 Zeilen in 23 ms`). 0 `FATAL EXCEPTION`.

**Offen / nicht getestet**
- Echtes Android 9, Online-Modus.
- Scrollposition auf einem Gerät mit anderer Bildschirmhöhe (der gesicherte Wert wird
  unverändert gesetzt; bei kleinerem Bildschirm klemmt die Ansicht selbst ab).

**Nächster konkreter Arbeitsschritt**
1. Native Startansicht (D-12) und Verbindungsstatus.
2. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.

---

## 23.11.2026 · Etappe 3, Schritt 5: Drehung – Zustand bleibt, Chips brechen neu um

**Warum:** Prüfen, ob Suchtext, Filter-Chips und Scrollposition eine Drehung überstehen.

**Befund**
- Die `SpieleActivity` wird bei Drehung **nicht neu gebaut** (`configChanges` im Manifest,
  D-21). Suchtext, gewählte Merkmale und Scrollposition bleiben dadurch von selbst erhalten.
- **Aber die Chip-Leiste wurde nur einmal umbrochen**, mit der Breite vom Start. Im
  Querformat (2296 px) passen alle sieben Chips in eine Zeile; nach dem Zurückdrehen auf
  1080 px blieb diese eine Zeile stehen: „Lokal am Gerät“ wurde auf 74 px zusammengedrückt
  (Bounds `[940,444][1014,1082]`, Buchstaben übereinander), „Online“, „Gegen Computer“ und
  „Highscore“ lagen ganz außerhalb des Bildes – diese Filter waren nicht mehr erreichbar,
  und die Liste begann erst bei y = 1115.

**Was gemacht wurde**
- `SpieleActivity` (541 → **560** Zeilen): `onConfigurationChanged` baut die Chip-Leiste
  mit der neuen Anzeigebreite neu (`baueChips()`) und stellt die Scrollposition wieder her
  (`rollen.post(scrollTo)`). Die Auswahl steht in `aktiveMerkmale`, `chipAussehen()` malt
  sie danach wieder an; `WebBasis.vollbild` wird erneut gesetzt.
- `baueChips()` rechnet die verfügbare Breite jetzt immer aus
  `getResources().getDisplayMetrics().widthPixels - dp(32)` – die Breite der Ansicht ist im
  Callback noch die alte.

**Tatsächlich getestet** (Emulator Android 14, echte Drehung per `adb emu rotate`)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 9s`.
- Hochkant (1080 × 2340) → quer (2340 × 1080) → hochkant: Chips einmal in einer Zeile
  (letzter Chip endet bei 2080 von 2160), danach wieder drei Zeilen innerhalb 1036 px,
  alle sieben erreichbar.
- Unverändert über beide Drehungen: Zähler „37 von 94“, Suchfeld „a“, „Mehrspieler“
  `[GEWÄHLT]`, oberste sichtbare Zeile „Risiko – Welt erobern“. 0 `FATAL EXCEPTION`.

**Offen / nicht getestet**
- Echtes Android 9, Online-Modus; Drehung mit geöffneter Tastatur nicht eigens geprüft.
- Prozess-Tod (Activity wird wirklich neu erstellt): Suchtext und Chips wären weg, weil
  die Ansichten keine IDs haben und `onSaveInstanceState` nichts sichert.

**Nächster konkreter Arbeitsschritt**
1. `onSaveInstanceState` für Suche, Chips und Scrollposition (deckt Prozess-Tod ab).
2. Native Startansicht (D-12) und Verbindungsstatus.
3. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.

---

## 23.11.2026 · Etappe 3, Schritt 4: kein zweiter Einstieg zur Spieleliste

**Warum:** Die Connector-Karte „Alle Spiele (offline)“ lag auf der Connector-Startseite,
die App-Startseite öffnet den Connector aber auf „Lokales Netzwerk“ – man kam nur über den
Kopfknopf „Start“ daran vorbei. Außerdem hat die Startseite die Kachel „Spiele“, die
dieselbe native Liste öffnet.

**Entscheidung (D-20):** Die Karte ist **entfernt**, nicht verschoben. Ein zweiter Einstieg
auf dem Lokal-Bildschirm hätte dieselbe Liste ein zweites Mal angeboten, mitten in einem
Bildschirm, der Räume und Chat koordiniert (dort spielbar ist nur der Galgen). Damit gibt
es genau einen Weg zu den Spielen: die Kachel „Spiele“ auf der Startseite.

**Was gemacht wurde**
- `ConnectorActivity` (461 → **442** Zeilen): Karte „Alle Spiele (offline)“ samt
  `spieleAnzahl()` entfernt (die Hilfsmethode hatte keinen anderen Aufrufer). Die
  Connector-Startseite zeigt nur noch „Lokales Netzwerk“ und „Online verbinden“.
- Keine Änderung an `SpieleActivity`, `WebNavigation`, Manifest, Spielen oder Version.

**Tatsächlich getestet** (Emulator Android 14)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 9s`.
- Connector-Startseite: nur noch zwei Karten, kein „Alle Spiele (offline)“ mehr.
- Lokal-Bildschirm (dort landet man von der Startseite): unverändert – Räume, Chat,
  „Galgenmännchen im Raum“.
- Startseite → Kachel „Spiele“ → `SpieleActivity` („94 Spiele“, Filterleiste). 0 FATAL.

**Offen / nicht getestet**
- Echtes Android 9, Rotation, Online-Modus.

**Nächster konkreter Arbeitsschritt**
1. Native Startansicht (D-12) und Verbindungsstatus.
2. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.

---

## 23.11.2026 · Etappe 3, Schritt 3: Merkmalsfilter, Connector-Karte geprüft

**Warum:** Die native Liste konnte nur per Text suchen; der HTML-Katalog filtert zusätzlich
nach Merkmalen. Und die Connector-Karte „Alle Spiele (offline)“ war nach dem Umbau noch
nicht angesehen worden.

**Was gemacht wurde**
- `SpieleActivity` (363 → **541** Zeilen): Filterleiste aus `katalog.json`.
  - Chips wie im HTML-Katalog: Einzelspieler, Mehrspieler, Lokal am Gerät, Online,
    Gegen Computer, Highscore, Controller. Angeboten wird nur, was im Katalog vorkommt –
    `controller` fehlt deshalb (in den 94 Einträgen nicht vergeben, Stand heute).
  - Mehrere Chips sind **UND**-verknüpft, Suche und Chips wirken zusammen: dieselbe Regel
    wie `passt(s, q)` in `spiele.html`.
  - „Zurücksetzen“ leert Suche und Chips und bleibt immer sichtbar, damit die Leiste beim
    Filtern nicht springt.
  - Kategorieköpfe zeigen bei aktivem Filter „x / gesamt“, der Zähler „x von 94“.
  - Leere Treffermenge: eigener Hinweis statt leerer Fläche.
- Chips **brechen um** (drei Zeilen bei 1080 px) statt seitlich zu scrollen; sonst lägen
  „Gegen Computer“ und „Highscore“ außerhalb des Sichtbereichs.
- Zeilenansichten werden einmal gebaut und beim Filtern nur umgehängt (`zeilen`-Cache);
  nach jedem Filterwechsel steht die Liste wieder oben.
- `ConnectorActivity` (443 → **461** Zeilen): Die Karte „Alle Spiele (offline)“ sagt nicht
  mehr „dazu die Startseite“ – die öffnet die native Liste nicht –, sondern beschreibt die
  Liste. Die Zahl kommt aus `katalog.json` (`spieleAnzahl()`), ist also nicht mehr fest
  eingetippt.

**Tatsächlich getestet** (Emulator Android 14, Details im Prüfbericht)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 10s`.
- „Mehrspieler“ → „70 von 94“, „Kartenspiele · 10 / 11“; zusätzlich „Gegen Computer“ →
  „5 von 94“ (UND-Verknüpfung); „Zurücksetzen“ → 94 Spiele, Suchfeld leer.
- Suche „zzz“ → „0 von 94“ und der Hinweis auf das Zurücksetzen.
- Filterwechsel laut Log (`Liste: … Zeilen in … ms`): 94 Zeilen 31 ms, 23 Zeilen 39 ms,
  5 Zeilen 14 ms; erster Aufbau 82 ms.
- Scrollen mit `dumpsys gfxinfo`: 94 Zeilen 36 ms Median / 57 ms 90er-Perzentil (33 %
  Janky), 23 Zeilen 40 ms / 61 ms (39 % Janky) – die Zahl der Einträge ändert die
  Bildzeit nicht. Der Emulator rendert mit `ro.hardware.egl=emulation` (Software-GL).
- Connector-Karte: Startseite → „Lokales Netz“ → im Connector „Start“ → Karte gelesen
  („94 Spiele, ohne Internet nutzbar“) → „Spiele öffnen“ → `SpieleActivity` → Zurück →
  `ConnectorActivity`. 0 `FATAL EXCEPTION`.

**Offen / nicht getestet**
- Echtes Android 9 (Huawei), Rotation, Online-Modus.
- Flüssigkeit auf echter Hardware: die Emulatorzahlen sind softwaregerendert.
- Die Karte liegt auf der Connector-Startseite (Kopfknopf „Start“); die App-Startseite
  öffnet den Connector direkt auf „Lokales Netzwerk“. Dorthin ist es ein Umweg.

**Nächster konkreter Arbeitsschritt**
1. Entscheiden, ob die Karte „Alle Spiele (offline)“ auf der Connector-Startseite bleibt
   oder auf den Lokal-Bildschirm wandert, wo man von der App-Startseite landet.
2. Native Startansicht (D-12) und Verbindungsstatus.
3. `keystore.properties` wiederherstellen → release bauen → Gerätetest Android 9.

---

## 23.11.2026 · Etappe 3, Schritt 2: ein Routing, native Spieleliste, Zähler

**Warum:** Nachkontrolle mit frischem Blick – und der nächste Schritt zum Zielbild.

**Befund der Nachkontrolle**
- Das URL-Routing stand **doppelt** da (`MainActivity.routeNavigation` und
  `GameActivity.leiteUm`) und war schon auseinandergelaufen: der Online-Sonderfall
  (`onlineGamePage`) gab es nur in `MainActivity`, die Fernbedienungs-Regel prüfte dort
  zusätzlich `/WLAN-Fernbedienung/index.html`. Dazu zwei Kopien der WebView-Einstellungen
  und des Konsolenprotokolls.
- Die Spielzahlen waren veraltet: `start/app.html` sagte „93 Spiele“, die Connector-Karte
  „91 Spiele“, im Katalog stehen **94**.

**Was gemacht wurde**
- Neu `WebNavigation.java` (142 Zeilen): **eine** Regel für alle Activities
  (`pruefe` → `BLEIBT | KATALOG | SPIEL | KONNECTOR_LOKAL/EXTRAS/ONLINE | BROWSER`,
  `starte` führt aus). Spielseiten werden nur umgeleitet, wenn der Aufrufer es erlaubt –
  in einer Spielseite bleiben Links im selben WebView.
- `MainActivity` (267 → **134** Zeilen): nur noch Startansicht. Entfallen sind `TGApp`,
  `injectGameModeChooser`, der Modus-Dialog, `safeGamePath`, `onlineGamePage` und
  `EXTRA_START_PATH` (der Katalog ist nativ, die Brücke hatte keinen Aufrufer mehr).
- `GameActivity` (266 → **219** Zeilen): Routing nur noch über `WebNavigation`,
  Online-Adresse aus `WebBasis`.
- Neu `SpieleActivity.java` (363 Zeilen): **native Spieleliste** aus `assets/katalog.json`
  (Suche, Kategorien mit Anzahl, Merkmale, Dialog „Lokal / Online“, startet `GameActivity`).
- `WebBasis`: `ONLINE_HOST`/`ONLINE_ROOT` und `protokoll(tag)` (Konsolenmeldungen).
- Zahlen: `app.html` 93 → **94**, Connector-Karte 91 → **94** und öffnet jetzt
  `SpieleActivity` statt der HTML-Seite.
- `AndroidManifest.xml`: `SpieleActivity` (`exported=false`, `fullSensor`).

**Tatsächlich getestet** (Emulator Android 14, Details im Prüfbericht)
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 16s`.
- Start auf `app.html` mit „94 Spiele“; Tipp auf „Spiele“ → `SpieleActivity`,
  Log `Katalog: 94 Spiele, 9 Kategorien`.
- Suche „kniffel“ → „1 von 94“, nur Würfelspiele · Kniffel.
- „Solitär“ (nur Solo) → sofort `GameActivity`; „Baccara“ (Online) → Dialog →
  „Lokal / offline“ → `GameActivity`; Zurück führt beide Male zur Liste. 0 FATAL.

**Offen / nicht getestet**
- HTML-Katalog wird von der App nicht mehr verlinkt (bleibt im Paket und bleibt Quelle).
- Filter nach Merkmalen (nur Textsuche), Online-Modus, Rotation, echtes Android 9.
- Gerät A unverändert (1.15, release-signiert).

**Nächster konkreter Arbeitsschritt**
1. Feinschliff der Liste: Merkmals-Filter wie im HTML-Katalog, zuletzt gespielte Spiele.
2. Native Startansicht (D-12) und Verbindungsstatus.
3. `keystore.properties` wiederherstellen → release bauen → Gerätetest.

---

## 23.11.2026 · Etappe 3, Schritt 1: Spielseiten als eigene Activity, Katalog als JSON

**Warum:** Spiele lagen im selben WebView wie Startseite und Katalog – kein eigener
Zurück-Weg, Katalogzustand ging beim Spielen verloren, und es gab keine maschinenlesbare
Spieleliste für eine spätere native Ansicht.

**Was gemacht wurde**
- Neu `GameActivity.java` (266 Zeilen): zeigt genau eine Spielseite, lokal aus den Assets
  (`EXTRA_PFAD`, `EXTRA_ONLINE`) oder vom Render-Server; eigenes URL-Routing
  (`/lokal`, `/extras`, `/WLAN-Fernbedienung` → `ConnectorActivity`), Mahjong-Brücke,
  Zurück = `goBack()` sonst `finish()`, `onRenderProcessGone` wie bisher.
- Neu `WebBasis.java` (68 Zeilen): gemeinsame WebView-Einstellungen (`LOAD_NO_CACHE`,
  DOM-Speicher, kein Dateizugriff), Vollbild und die Pfadprüfung – damit Start- und
  Spielansicht nicht auseinanderlaufen.
- `MainActivity.java` (296 → 267 Zeilen): behält `app.html` und den Katalog, gibt
  Spielklicks an `GameActivity` ab (auch der Dialog „Lokal / Offline“ / „Online-Spiel“),
  Mahjong-Brücke und doppelte Vollbild-/WebView-Einstellungen sind entfallen.
- `AndroidManifest.xml`: `GameActivity` (`exported=false`, `fullSensor`, `singleTask`,
  `configChanges` wie `MainActivity`).
- `app/build.gradle`: neue Aufgabe `katalogBauen` (hängt an `preBuild`) liest den fertigen
  Katalog `www/spiele/spiele.html` und schreibt `src/main/assets/katalog.json`
  (94 Spiele, 9 Kategorien, Merkmale inkl. `multi`/`online` aus `multiplayer-register.js`).
  Bricht ab, wenn keine Spiele oder eine genannte Datei fehlt.
- `.gitignore`: `app/src/main/assets/katalog.json` (erzeugt, nicht eingecheckt).

**Verhalten:** unverändert für den Nutzer – gleiche Seiten, gleiche Dialoge, gleiches
Routing. Neu ist nur, dass eine Spielseite in einer zweiten Activity liegt.

**Tatsächlich getestet** (Details im Prüfbericht):
- `:app:assembleDebug` → `BUILD SUCCESSFUL in 17s`, Debug-APK 11.377.723 B.
- `katalog.json` im APK (`assets/katalog.json`, 33.896 B), 94 Spiele, keine doppelten Pfade.
- Emulator Android 14: Start landet auf `app.html`, Tipp auf „Spiele“ öffnet den Katalog
  (94 Spiele in 6 Kategorien), „Baccara → Lokal / offline“ öffnet `GameActivity`,
  Zurück führt mit erhaltener Kategorie und Scrollposition zum Katalog.
- Mahjong öffnet ebenfalls in `GameActivity`, Brett wird korrekt gezeichnet.

**Offen / nicht getestet**
- `SPIEL VERLASSEN` im Mahjong-Menü nicht gedrückt (Brücke injiziert wie vorher, aber nicht
  ausgelöst); „ZURÜCK ZUM CONNECTOR“ im Mahjong-Menü nicht geprüft.
- Kein Test auf dem echten Android-9-Gerät (Debug-APK passt nicht über die
  release-signierte 1.15, siehe D-14).
- Rotation, Online-Modus, LAN, Hotspot: unverändert offen.
- Zahlen im Text sind veraltet: `start/app.html` sagt „93 Spiele“, `ConnectorActivity`
  „91 Spiele“, der Katalog enthält 94. Noch nicht angefasst (nur Text).

**Nächster konkreter Arbeitsschritt**
1. Native Spieleliste aus `katalog.json` (RecyclerView oder ListView, Suche/Filter),
   damit der Katalog nicht mehr HTML sein muss.
2. Danach native Startansicht (D-12) und Verbindungsstatus.
3. `keystore.properties` wiederherstellen → release-signierter Build → Gerätetest.

---

## 23.11.2026 · Etappe 1: Bestandsaufnahme und Dokumentation

**Warum:** Die App wirkt wie eine verkleinerte Webseite; vor jedem Umbau musste klar
werden, was tatsächlich existiert, was fehlt und warum.

**Was gemacht wurde**
- Beide Android-Projekte gelesen: `spiele/android` (Gradle, `de.trashygecko.tmahjong`,
  1.15 / Code 16, minSdk 21, targetSdk 34) und `apps/t-spiele` (Altlast, `bauen.sh`, 1.4).
- Alle Java-Dateien der App gesichtet (2.445 Zeilen): `MainActivity` (WebView-Wirt der
  Homepage), `ConnectorActivity` (native Connector-Ansicht), `LocalNetworkCoordinator`
  (NSD + TCP-Relay), `GalgenLokal`, `RemoteControlActivity`, `remote/*`, `HomepageAssets`.
- Bauweg verstanden: `homepageKopieren` → `assets/www` (353 Dateien, 108 HTML-Seiten),
  `verifyHomepageGames`, `packageTrashyGeckoApk`.
- Umgebung geprüft: echtes Android-9-Gerät (Huawei FIG-LX1) und Emulator (Android 14)
  sind per `adb` erreichbar; Werkzeuge (JDK 17, SDK 34, Gradle 8.7) liegen in
  `Documents\arbeit\werkzeuge`.
- Echter Build ausgeführt: `:app:assembleDebug` → `BUILD SUCCESSFUL in 46s`,
  Debug-APK 11.346.443 B.
- Neue Dokumente angelegt: `AGENTS.md`, `ANDROID-ARCHITEKTUR.md`, `ENTSCHEIDUNGEN.md`,
  `ENTWICKLUNGSSTAND.md`, `PRUEFBERICHT-ANDROID.md`.

**Bearbeitete Dateien**
- nur Dokumentation (5 neue Markdown-Dateien im Projektwurzelverzeichnis)
- keine Änderung an Code, Gradle, Manifest oder Assets

**Was jetzt funktioniert (Ist-Stand, nicht neu getestet)**
- Die App baut über Gradle und enthält die komplette Spieleabteilung als Assets.
- Start über `app.html` (HTML) mit vier Kacheln; Katalog als `spiele/spiele.html`;
  Spiele öffnen im selben WebView.
- Nativer Connector-Bereich (`ConnectorActivity`): Start, Lokales Netz (Chat + Galgen),
  Online-Connector (WebView auf onrender.com), Extras mit Fernbedienung.
- Lokales Netz technisch vorhanden (NSD + Relay, 8-stelliger Code, Wiederverbindung).

**Erkannte Hauptprobleme**
1. Start und Spieleliste sind HTML → „verkleinerte Webseite“.
2. Zwei getrennte Welten (WebView-Home vs. native Connector-Ansicht) ohne gemeinsame
   Navigation und ohne Verbindungsstatus.
3. Nur Galgen läuft im lokalen Netz; alles andere braucht den Online-Server.
4. `keystore.properties` fehlt → derzeit kein release-signierter Build möglich
   (Keystore liegt außerhalb, Passwort in `passwort.txt` daneben).
5. `apps/t-spiele` (Zwei-Knopf-App) wird mit „der App“ verwechselt.

**Offen**
- Entscheidung des Nutzers zum Zielbild (siehe `ANDROID-ARCHITEKTUR.md`, Abschnitte 4
  und 5, sowie D-12 in `ENTSCHEIDUNGEN.md`): bleibt der Start nativ (empfohlen) und die
  echte Website unter Extras?
- Danach Umsetzung: `GameActivity` herauslösen, `katalog.json` im Build erzeugen,
  native Oberfläche (Tabs, Start, Spieleliste, Einstellungen), Netzwerkstatus.
- `keystore.properties` wiederherstellen, damit `TrashyGecko.apk` release-signiert
  entsteht (ohne die Datei ist kein Update über 1.15 möglich).
- Gerätetests (siehe `PRUEFBERICHT-ANDROID.md`, Abschnitt 4) – in dieser Runde keine.

**Nächster konkreter Arbeitsschritt**
1. `GameActivity` aus `MainActivity` herauslösen (WebView + Asset-Loader + Zurück),
   `MainActivity` behält vorerst `app.html` als Start. Danach bauen und auf dem Emulator
   starten. → Fundament für alles Weitere, ohne sichtbare Änderung.

**Tatsächlich durchgeführte Tests**
- `:app:assembleDebug` erfolgreich (46 s, siehe Prüfbericht).
- `adb devices`: Gerät A und Emulator B erkannt; Version auf Gerät A abgefragt (1.15/16).
- **Keine** Installations-, Start-, Navigations- oder Netzwerktests in dieser Runde.

---

## Vorheriger Stand (aus `spiele/android/README.md`, nicht neu geprüft)

- **1.15 / Code 16** (22.11.2026): letzte Fassung; enthielt die Moderator-/Quiz-Arbeit
  nicht (die betrifft nur den Connector und die Web-Seite).
- **1.14 / Code 15** (21.11.2026): Fernbedienung ausschließlich unter Extras, Startseite
  mit vier Kacheln, `LOAD_NO_CACHE`, `onRenderProcessGone`, Release-Signatur eingeführt.
- **1.13 und früher**: Umbau von „nur Mahjong“ zur Spiele-App, lokale Räume, Galgen im
  LAN, Fernbedienung.

---

## Scherzshow: Zeitsteuerung (24.11.2026)

- **Antwortzeit 30 s in beiden Modi** (`quizZeitTempo` und `quizZeitPunkte` in
  `spiele/connector/quiz.go`, vorher 20 s bzw. 15 s). Die Zeit ist eine Höchstzeit;
  angezeigt wird sie aus `zeitMs`, die Spielseite hat keine eigenen Sekunden.
- **Sofortiger Rundenabschluss**: `quizAntwort` wertet aus, sobald `quizAlleDran()`
  meldet, dass alle Spieler geantwortet haben – es wird nicht auf die Restzeit
  gewartet. Bestand schon, ist jetzt durch Tests festgenagelt; ebenso der Abgang
  eines Spielers (`quizRemoveSeat`) und die Frist als Obergrenze.
- **Spielseite** (`spiele/quizshow/quizshow.html`): Nachladen und eigene Aktion haben
  getrennte Sperren (`busy` / `sendet`). Vorher verschluckte ein laufendes Nachladen
  (alle 900 ms) einen Klick auf eine Antwort – der Spieler stand dann als „hat nicht
  geantwortet“ da und die Runde lief in die Höchstzeit. Wer geantwortet hat, sieht
  Frage, alle vier Antworten und den Quizmaster-Kommentar weiter, kann aber keine
  zweite Antwort abgeben.
- Auflösungsdauer unverändert 7 s (`quizPauseZeit`), Punkte, Fragen, Regeln und
  Gestaltung unverändert.
- Binaries neu gebaut (`trashygecko-connector.exe`, `trashygecko-connector-online.exe`,
  `trashygecko-connector-online-linux`), `connector-online` über `sync.py` abgeglichen,
  APK neu gebaut (14.677.021 B). Sicherung: `Homepage/sicherung/quizzeit-2026-11-24/`.

## Spielehalle ohne Fuß (24.11.2026)

- Der Fußbereich unter dem Spielekatalog ist ganz weg: Links (Startseite, Highlights,
  Werkzeuge & Extras), Datum/Uhrzeit und das Gameboy-Zeichen aus Text. Mit dem Rest der
  Seite hatte er nichts mehr zu tun.
- Entfernt: `<footer class="fuss">` samt Inhalt und das 1-s-Skript für Datum/Uhrzeit in
  `spiele/spiele.html`; die Regeln `.fuss`, `.fuss-links`, `#local-date-time`, `.gameboy`
  (auch die Handy-Regel `.gameboy{display:none}`) in `spiele/halle.css`. Reine
  Anzeige – Katalog, Suche, Filter, Online-Halle unverändert.
- Geprüft (Chromium 390×844 und 1280×800): 0 Elemente `.fuss`/`footer`/`.gameboy`,
  letztes Element ist die Hallen-Auswahl, Seite 1647 px bzw. 1030 px hoch, kein
  Querscrollen, keine Konsolenfehler, 10 Karten wie zuvor.
- Beide Ablagen und die Android-Assets byte-gleich; APK neu 14.676.443 B.

## Super Pimmelchen auf Schnittstelle 2.0 (26.11.2026)

- **`spiele/pimmel/spimmel.html` erfüllt jetzt `SPIELE-SCHNITTSTELLE.md` 2.0**
  (`version` 2.0.0, Schnittstellenobjekt eingefroren): `whenReady()` löst auch
  dann auf, wenn es erst nach dem Laden gefragt wird; `aktion(name, daten)`
  führt `richtung`, `feuer`, `start`, `pause`, `weiter` und `neu` über
  dieselben internen Funktionen wie Tastatur und Touch aus; `:bereit`, `:ende`,
  `:pausiert` und `:weiter` kommen nur bei echter Zustandsänderung und tragen
  die Details aus dem Standard als Kopie.
- **`einstellungen()` ohne Bedeutung:** `specialsTaktMs` ist entfernt, geliefert
  werden `stufe`, `stufen` (nur die Namen), `level` und `levelAnzahl`.
  `zustand().startbildschirm` folgt jetzt dem Spielzustand statt der
  Body-Klasse.
- **Fehlercodes bleiben in der Tabelle des Standards:** ungültige `daten` bei
  `richtung` (diagonal, `0/0`, Text) → `aktion-nicht-erlaubt`, fremde Felder →
  `unbekanntes-feld`. Ein eigener Code wäre eine Änderung am Standard.
- **Behoben:** Pfeiltasten im fokussierten Level-Feld starteten die Partie
  (Fokusregel gilt jetzt auch für Pfeile, unsichtbare Bedienelemente zählen
  nicht); das automatische Fortsetzen bei `focus` hob eine manuelle Pause beim
  Tabwechsel auf (`blur → pause` bleibt); der Respawn nach verlorenem Leben
  pausiert mit; Totenkopf-Entferner-Timer werden verfolgt und beim Stoppen
  gelöscht; Escape verlässt im Vollbild zuerst das Vollbild, ohne zu pausieren.
- **Unverändert:** alle 61 Level samt Leveldaten, Mechanik, Stufen 5/15/20 s,
  Startbildschirm, Touch-Steuerung, Pause, Vollbild, Highscore und gespeicherte
  Stufe, Einzeldatei-Fassung (1.0.0, gemeinsame Speicherschlüssel bleiben).
- **Offen:** der Highscore nutzt weiter `prompt`/`alert` (Abschnitt 10 verbietet
  das; der Auftrag nahm den Highscore ausdrücklich aus). Für fehlerhafte
  Aktionsdaten fehlt in Abschnitt 3.2 ein eigener Code – im Standard vermerkt,
  nicht eigenmächtig ergänzt.
- **Geprüft** (Chromium, lokaler HTTP-Server, Sicherung vorher): 119 Prüfungen
  grün – Laden, `whenReady()` vor und nach der Bereitschaft, alle drei Stufen,
  Level 1/30/61, alle 61 Level, Kopien von `zustand()`/`einstellungen()`, zehn
  ungültige Konfigurationen und alle Aktionsfehler ohne Zustandsänderung,
  Tastatur/Touch/`aktion()` auf demselben Weg, Pause/Weiter/Stopp/Ende
  inklusive Detailfeldern, 50-fache Wiederholungen ohne Doppel-Timer, Specials
  nach dem Stoppen, Vollbild mit Escape, 1280×800, 390×844, 844×390, 360×640,
  0 Konsolenfehler; dazu die beiden älteren Prüfläufe (41 und 15 Prüfungen) auf
  2.0.0 gezogen, beide grün. Bei 844×390 endet das Feld 2 px unter der
  Fensterhöhe (bestehende Rundung der Bühnenberechnung, nicht angefasst).
  Sicherung: `Homepage/sicherung/pimmel-2.0-2026-11-26/`.

## Einzeldatei auf Schnittstelle 2.0 (26.11.2026)

- **Beide Pimmel-Fassungen sind jetzt eine Version:**
  `Super-Pimmelchen-61-Level-Einzeldatei.html` wird aus `spimmel.html` gebaut
  (Bauskript `Documentsrbeit	mp\pimmel_einzel_bauen.py`) und hat damit
  Schnittstelle 2.0.0, dieselbe Spiellogik, dieselben 61 Level, dieselben
  Stufen 5/15/20 s, dieselben Aktionen und dieselben Fehlercodes. Diff der
  beiden Hauptskripte: nur der Asset-Lader unterscheidet sich.
- **Eingebettet statt geladen:** 19 Spielgrafiken als Data-URI, Fredoka und
  Press Start 2P als Data-URI, dazu die vier Atkinson-Schnitte; `/spiele/stil.js`,
  `/spiele/schrift.css` und `/spiele/zurueck.js` stehen jetzt als Text in der
  Datei. Kein `<link>`, keine Anfrage nach außen – auch `file://` läuft
  (gemessen: Version 2.0.0, Start, 314 Farben auf der Leinwand, 0 Fehler).
- **Vorher (Schnittstelle 1.0.0):** kein `whenReady`, kein `aktion()`, kein
  `pause`/`resume`/`stop`, `specialsTaktMs` in `einstellungen()`, weiche
  Validierung (`'7'` als Level erlaubt). Außerdem war `var(--tg-schrift)`
  nirgends definiert (die Datei band `schrift.css` nicht ein), der Text fiel
  also auf die Browser-Schrift zurück; jetzt stehen die Regeln samt Schriften
  in der Datei, `getComputedStyle(body).fontFamily` ist in beiden Fassungen
  gleich.
- **Entfallen:** der alte Querformat-Helfer der Einzeldatei (eigenes
  `.game-area`-Flex-Layout unter 560 px Höhe) und das Vorladen der Schriften.
  Beides war an das ältere Gerüst gebunden; die Einzeldatei nutzt jetzt
  dasselbe responsive Layout wie die Referenz, damit sich die Fassungen nicht
  unterschiedlich verhalten.
- **Unverändert:** gemeinsame Speicherschlüssel `superPimmelchenStufe` und
  `superPimmelchenHighscore` (bekannte Ausnahme), Startbildschirm, Touch,
  Vollbild, Highscore, Dateiname und Katalogeintrag.
- **Geprüft:** beide Fassungen einzeln mit demselben 2.0-Prüfsatz –
  `spimmel.html` 120 Prüfungen grün, Einzeldatei 125 grün (5 zusätzliche zur
  Eigenständigkeit), je 0 rot; die Einzeldatei dabei aus einem leeren Ordner
  ausgeliefert (keine Anfrage außer der Seite selbst). Direktvergleich über
  21 Punkte (Schnittstellenflächen, `whenReady`, `einstellungen()`, `zustand()`,
  21 Fehlerfälle, `start`, `pause`, `resume`, alle Aktionen, `stop`,
  Ereignisabfolge, Speicherschlüssel, Konsolenfehler): alle gleich.
  Sicherung vorher: `Homepage/sicherung/pimmel-einzeldatei-2026-11-26/`.