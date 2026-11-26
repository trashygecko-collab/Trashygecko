# Spielschnittstelle – TrashyGecko

**Verbindlich für jedes Spiel.** Vorbild ist `spiele/pimmel/spimmel.html`
(Schnittstelle **1.1.0**) mit Prüfungen in `arbeit/tmp/pimmel_p1_p5_pruefen.js`.
Grundlage: Architektur-Audit 3 vom 26.11.2026 (`spimmel.html`).

## 1. Drei Schichten

| Schicht | Darf | Darf nicht |
|---|---|---|
| **Spiel** | Mechanik, Grafik, Sound, Eingaben, Bots, Stufen, Wertung, eigene Einstellungen, Start/Pause/Neustart/Ende | Netzwerk, Räume, Spieler-IDs, Spielregeln anderer Spiele |
| **Konnektor** | Räume, Einladungen, Identitäten, Transport von Aktionen, Verbindungsabbruch, Koordination von Start und Ende, Weiterreichen der Einstellungen | Spielregeln, Mechanik, Botlogik, Bedeutung einzelner Werte (was „schwer" heißt), Leveldaten |
| **Sammlung** | Übersicht, Kategorien, Öffnen/Schließen, Katalog-Metadaten (Solo/Online/Bots/Rekord) | Spielmechanik, Zustandslogik, Regeln |

Der Konnektor reicht `stufe: 'schwer'` als Zeichenkette durch und weiß nicht,
dass das 20 s Special-Takt bedeutet. Das Spiel weiß nicht, in welchem Raum es
läuft.

## 2. Pflicht je Spiel

Ein einziges Global, sonst nichts (`window.<SpielName>`, PascalCase; neue
Spiele frieren es mit `Object.freeze` ein – Pimmel 1.1.0 noch nicht):

| Feld | Typ | Bedeutung |
|---|---|---|
| `version` | string | SemVer der Schnittstelle, z. B. `'1.1.0'` |
| `bereit` | boolean | `true`, sobald initialisiert |
| `start(konfiguration?)` | Funktion | Partie beginnen, Antwort unten |
| `stop()` | Funktion | Partie abbrechen, zurück zur Auswahl (`{ok:true}`) |
| `pause()` | Funktion | Partie anhalten |
| `resume()` | Funktion | Partie fortsetzen |
| `einstellungen()` | Funktion | Startmöglichkeiten, **ohne** Bedeutung der Werte |
| `zustand()` | Funktion | Momentaufnahme als **Kopie** |
| `startbildschirm()` | Funktion | zurück zur Auswahl (Altname, bleibt) |

`start()` antwortet immer, wirft nie:

```js
{ ok: true, stufe: 'schwer', level: 4, levelAnzahl: 61 }
{ ok: false, fehler: 'level-muss-1-bis-61-sein' }
```

`pause()` / `resume()` antworten `{ok:true, pausiert:true|false}`,
`stop()` und `startbildschirm()` `{ok:true}`.

## 3. Feste Fehlercodes

| Code | Wann |
|---|---|
| `spiel-nicht-bereit` | `start()` vor der Initialisierung |
| `konfiguration-muss-objekt-sein` | Zahl, Array, Funktion statt Objekt |
| `unbekanntes-feld` | Feld außerhalb der erlaubten Liste (Antwort nennt `feld`) |
| `unbekannte-stufe` | Stufe nicht in `einstellungen().stufen` |
| `stufe-muss-text-sein` | Stufe kein Text |
| `level-muss-1-bis-61-sein` | Level keine ganze Zahl im Bereich |
| `spiel-laeuft` | `start()` während einer laufenden Partie |
| `spiel-laeuft-nicht` | `pause()`/`resume()` ohne Partie |

Regeln: nur bekannte Felder, nur echte Typen (`'12'` ist kein Level, `true`
ist kein Level), `Object.prototype.hasOwnProperty` statt `in`. Bei einem
Fehler bleibt der Spielzustand **unverändert**. Fehlende Felder bleiben auf
der aktuellen Auswahl; `null`/`undefined` als ganzes Argument ebenfalls.

## 4. Ereignisse

| Ereignis | Detail |
|---|---|
| `<spiel>:bereit` | `{version}` |
| `<spiel>:ende` | `{grund:'gewonnen'|'verloren', punkte, level, levelAnzahl, tick}` |

`ende` kommt genau einmal je Partie, **bevor** Highscore-Dialoge den Faden
anhalten. Der Konnektor pollt nicht, er hört zu.

## 5. Zustand

`zustand()` liefert eine flache Kopie; Mutationen daran ändern das Spiel nicht.
Pflichtfelder (Namen dürfen je Spiel ergänzt, nicht ersetzt werden):

```js
{
  laeuft, pausiert, tick, richtung: {x, y},
  level, punkte, extras: [{x, y, type}], ...
}
```

- `tick` zählt ausgeführte Spielzüge seit Partiebeginn (Neustart = 0). Er ist
  die spätere Vergleichsnummer für synchronisierte Partien.
- `richtung` und `extras` sind Kopien, keine Referenzen auf Spielobjekte.
- `einstellungen().level` = **Auswahl** für den nächsten Start,
  `zustand().level` = **laufendes** Level. Nicht verwechseln.

## 6. Eingaben

- Ein Eingabe-Seam: Tastatur und Touch rufen **dieselben** internen Funktionen
  (`setDirection`, `fire`). Ein späterer Konnektor speist genau dort ein.
- **Leertaste und Enter starten nur ohne fokussiertes Bedienelement.** Steht
  der Fokus auf Knopf, Auswahl oder Feld (`button, select, input, textarea,
  a[href], [contenteditable]`), löst der Browser selbst aus; das Spiel hält
  sich heraus (sonst startet Enter im Level-Auswahlfeld die Partie).
- Kein `preventDefault` für Tasten, die das Spiel nicht benutzt.

## 7. Verboten

- Spiel ruft selbst `/api/` oder öffnet WebSockets; Spiel liest `room`, `name`,
  `id` aus der Adresse.
- Konnektor oder Sammlung kennen Levelbedeutung, Punkteregeln, Botlogik.
- Spiel öffnet `prompt`/`alert` für Ergebnisse (Ergebnis über `:ende`).
- Zwei Kopien desselben Spiels mit derselben API-Version und denselben
  Speicherschlüsseln.

## 8. Übergang

- **Referenz:** `spiele/pimmel/spimmel.html`, 1.1.0.
- **Eingefroren:** `spiele/pimmel/Super-Pimmelchen-61-Level-Einzeldatei.html`
  bleibt Sonderfassung mit Schnittstelle 1.0.0 (kein `stop/pause/resume`, kein
  `:ende`, weiche Validierung). Sie wird nicht nachgezogen und nicht als
  Vorlage verwendet.
- Neue Spiele übernehmen Abschnitt 2–5 vollständig; ältere Spiele werden erst
  beim nächsten echten Umbau angepasst, nicht „nebenbei".
- Offen für später (nicht Teil von 1.1.0): `aktion(name, daten)` als
  öffentlicher Einspeisepunkt, Seed in `start()`, Ereignisse `:pausiert` und
  `:weiter`.

## 9. Startbildschirm, Vollbild und Feldmaß (Vorlage)

**Vorbild:** `spiele/frogger/frogger.html` mit `spiele/spiel-start.css` und
`spiele/spiel-start.js`. Das Gerüst ist absichtlich spielunabhängig: Farben
kommen aus den Tokens der Spielseite (`--panel`, `--gold`, `--hell`, `--muted`,
`--rand`, `--buehne`, `--schatten`), jeder der drei Grafstile bleibt eigen.

### 9.1 Was das Gerüst liefert

| Aufruf | Wirkung |
|---|---|
| `TGSpiel.vollbild.ist()` | läuft Vollbild (echt oder Pseudo)? |
| `TGSpiel.vollbild.an()` / `.aus()` / `.um()` | gezielt schalten |
| `TGSpiel.vollbild.knopf(el, {textAn, textAus})` | Knopf anschließen: Klick, Beschriftung, `aria-pressed`, Titel, Symbol |
| `TGSpiel.vollbild.beiWechsel(fn)` | `fn(an)` bei jedem Wechsel (neu messen) |
| `TGSpiel.feldPassen(canvas, behaelter, {rand, verhaeltnis, beiGroesse})` | quadratisch/verhältnistreu in den freien Platz setzen, bei jeder Größenänderung neu |
| `TGSpiel.zeige(el, an)` | `hidden` setzen/entfernen |

### 9.2 Vollbild

- Ein **sichtbarer** Knopf mit Zustand ist Pflicht; er bleibt auch im Vollbild
  erreichbar (im Spiel: Kopfleiste, nicht im Spielfeld).
- **Escape beendet den Vollbildmodus** – auch dort, wo es kein Fullscreen-API
  gibt (iPhone: Pseudo-Vollbild über `body.ist-vollbild`). Das Modul hört
  selbst auf Escape; Spiele prüfen zuerst `TGSpiel.vollbild.ist()` und
  behandeln Escape erst danach für eigene Dialoge.
- Nach dem Aussteigen wird die Darstellung über `fullscreenchange` neu
  berechnet; `beiWechsel` ist der Ort für eigenes Nachmessen.
- Ein Dialog mit `data-esc-sperre` ist kein Ersatz für den Ausstieg: Der
  Vollbild-Knopf bleibt bedienbar, Escape verlässt zuerst das Vollbild.

### 9.3 Feldmaß

- Gezeichnet wird in **festen logischen Maßen** (Frogger 416×416); nur die
  CSS-Größe ändert sich. Schärfe über einen DPR-Backing-Store und
  `ctx.setTransform(bw/W, 0, 0, bh/H, 0, 0)`.
- `feldPassen` misst den Behälter (ResizeObserver, `resize`,
  `orientationchange`, `visualViewport`, Vollbildwechsel, geladene Schriften)
  und setzt `width`/`height` in Pixeln, nie Prozent – sonst verzerrt das Feld.
- Das Spielfeld darf die Seite **nicht** scrollen: `body{overflow:hidden}`,
  Raster mit `minmax(0,1fr)` für die Bühne, Bedienleisten in eigenen Zeilen.

### 9.4 Startbildschirm

- Eigenständiger Bereich, **nicht** das Spielfeld mit einem Knopf darüber:
  eigener Rasterbereich zwischen Kopfleiste und Menü, HUD/Feld/Steuerung
  ausgeblendet (`body.zeigt-start`).
- Pflichtinhalt: großer Titel, Spielgrafik, Kurzbeschreibung,
  Steuerungshinweise, Spielinfos, hervorgehobener STARTEN-Knopf.
- Der Startbildschirm muss **ohne internes Scrollen** passen, sonst liegt
  STARTEN unter dem Rand. Gemessene Zielgrößen (Frogger): 693 px bei 390×844,
  489 px bei 360×640, 282 px bei 844×390, 492 px bei 1024×600.
- Querformat mit wenig Höhe bekommt ein eigenes Raster (Titel links, Inhalt
  Mitte, STARTEN rechts); sehr flache Geräte lassen die Szene weg, nie den
  Knopf.

### 9.5 Einheitliche Bedienlogik

- Start, Pause, Neustart, Beenden liegen an denselben Stellen und heißen in
  jedem Spiel gleich; Pause und Spielende erscheinen als Karte
  (`arcade-blatt`), nicht als `alert`.
- Tastatur: Pfeile/WASD steuern, `P` pausiert, `Escape` schließt Dialoge;
  **Leertaste und Enter starten nur ohne fokussiertes Bedienelement**.
- Der Rückweg zur Sammlung bleibt `zurueck.js` – keine eigene
  Connector-Logik im Spiel.

### 9.6 Übernahme in ein weiteres Spiel

1. `spiel-start.css` und `spiel-start.js` einbinden.
2. Kopfleiste mit Vollbild-Knopf, Startbildschirm und Karte nach obigem
   Markup einsetzen; Texte und Szene an das Spiel anpassen.
3. `TGSpiel.vollbild.knopf(...)` und `TGSpiel.feldPassen(...)` anschließen,
   Rasterbereich `start` im eigenen CSS anlegen.
4. Gegenprobe bei 390×844, 360×640, 844×390 und 1280×800: kein Scrollen,
   quadratisches Feld, STARTEN sichtbar, Vollbild an/aus per Knopf und Escape.

Nicht alle Spiele gleichzeitig umbauen: Jedes Spiel übernimmt das Gerüst erst
bei seinem eigenen Umbau.
