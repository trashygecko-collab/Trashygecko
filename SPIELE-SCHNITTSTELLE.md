# Spielstandard – TrashyGecko

**Verbindlich für jedes Spiel.** Diese Datei ist vor der Arbeit an einem Spiel
zu lesen. Sie beschreibt die Drei-Schichten-Architektur und die öffentliche
Schnittstelle, über die Konnektor, Sammlung und Tests ein Spiel bedienen.

- **Standard:** 2.0 (26.11.2026)
- **Grundlage:** Architektur-Audit von Super Pimmelchen (Commit `2bbdebb`,
  Prüfbericht `48a7af9`)
- **Umsetzungsreferenz:** `spiele/pimmel/spimmel.html`, Schnittstelle 1.1.0 –
  erfüllt den Standard noch **nicht** vollständig, siehe Abschnitt 11
- **Prüfungen:** `arbeit/tmp/pimmel_p1_p5_pruefen.js` (41) und
  `arbeit/tmp/pimmel_tastatur_pruefen.js` (15)

Zwingende Anforderungen und optionale Erweiterungen sind in Abschnitt 8
getrennt. Wo „muss“ steht, ist es Prüfgegenstand.

## 1. Drei Schichten

| Schicht | Zuständig für | Nicht zuständig für |
|---|---|---|
| **Spiel** | Mechanik, Regeln, Grafik, Sound, Steuerung, Bots, Schwierigkeitsgrade, Wertung, eigene Einstellungen, Startbildschirm, Start über den gemeinsamen Initialisierungsweg | Räume, Einladungen, Spieleridentitäten, Netzwerk, `/api/`, Spielerlisten |
| **Konnektor** | Räume, Einladungen, Spieleridentitäten, Verbindungsaufbau und -abbruch, Nachrichtenübertragung, Synchronisierung, Koordination von Start/Pause/Ende, Weitergabe der Einstellungen, Wiederverbindung, Host-Rolle | Spielregeln, Mechanik, Botlogik, Bedeutung einzelner Werte, Leveldaten, Grafik |
| **Sammlung** | Spielübersicht, Kategorien und Navigation, Öffnen und Schließen, Kataloginformationen, Anzeige von Solo-, Online-, Bot- und Rekordfunktionen | Spielmechanik, Zustandslogik, Regeln, Räume, Netzwerk |

Der Konnektor reicht `stufe: 'schwer'` als Zeichenkette durch und weiß nicht,
dass das 20 s Special-Takt bedeutet. Das Spiel weiß nicht, in welchem Raum es
läuft. Die Sammlung verlinkt ein Spiel, kennt aber seine Regeln nicht.

Keine Schicht ersetzt eine andere: Ein Spiel darf ohne Konnektor und ohne
Sammlung laufen; Konnektor und Sammlung dürfen ohne Spielmechanik arbeiten.

## 2. Eigenständiges Spiel

Jedes Spiel muss:

- **direkt im Browser laufen** – per Doppelklick (`file://`), über GitHub Pages
  oder im Konnektor-Rahmen, ohne Server und ohne Anmeldung;
- seine **vollständige Mechanik und seine eigenen Regeln** besitzen;
- Grafik, Sound, Steuerung, Bots, Schwierigkeitsgrade, Wertung und
  Einstellungen **selbst** verwalten;
- einen **eigenen Startbildschirm** haben (Abschnitt 9);
- über **einen gemeinsamen Initialisierungsweg** startbar sein:
  `start(konfiguration?)` – derselbe Weg für den Startknopf, den Konnektor,
  Bots und Tests. Kein zweiter Startpfad, keine Sonderfunktion für den
  Konnektor;
- **keine Raumverwaltung, keine Spieleridentitäten** und **keine eigene
  Netzwerkkommunikation** enthalten;
- **kein `/api/` aufrufen**, keine WebSockets öffnen und `room`, `name`, `id`
  nicht aus der Adresse lesen;
- den Rückweg zur Sammlung **ausschließlich** über den gemeinsamen Helfer
  `spiele/zurueck.js` bzw. `window.tgZurueck()` lösen, nicht selbst gebaut.

Ein Spiel, das diese Punkte erfüllt, ist auch ohne Konnektor vollständig
spielbar – das ist die Abnahmebedingung.

## 3. Öffentliche Spielschnittstelle

Jedes Spiel stellt **genau einen** eindeutig benannten globalen Namensraum
bereit: `window.<SpielName>` in PascalCase, abgeleitet aus dem Spielnamen
(`SuperPimmelchen`, `Frogger`, `Arkanoid`). Weitere Globals sind verboten;
gemeinsame Helfer heißen `TG…` (`TGStil`, `TGSpiel`, `TGW`).

| Feld | Typ | Bedeutung |
|---|---|---|
| `version` | string | SemVer der **Schnittstelle**, z. B. `'1.1.0'` |
| `bereit` | boolean | `true`, sobald vollständig initialisiert |
| `whenReady()` | Funktion → Promise | löst mit `{ok:true, version}` auf, sobald bereit; sofort, wenn schon bereit; lehnt nie ab |
| `start(konfiguration?)` | Funktion | Partie beginnen, Antwort unten |
| `einstellungen()` | Funktion | Startmöglichkeiten, **ohne** Bedeutung der Werte |
| `zustand()` | Funktion | Momentaufnahme als **Kopie** |
| `startbildschirm()` | Funktion | zurück zur Auswahl |
| `stop()` | Funktion | Partie abbrechen, zurück zur Auswahl |
| `pause()` | Funktion | Partie anhalten |
| `resume()` | Funktion | Partie fortsetzen |
| `aktion(name, daten)` | Funktion | eine dokumentierte Spielaktion auslösen |

Die Schnittstelle kann mit `Object.freeze` eingefroren werden – empfohlen für
neue Spiele, optional für Bestandsspiele (Abschnitt 8).

### 3.1 Rückgaben

Alle Aufrufe **antworten** und **werfen nie**:

```js
start({stufe:'schwer', level:4})  → {ok:true, stufe:'schwer', level:4, levelAnzahl:61}
start({level:'4'})                → {ok:false, fehler:'level-muss-1-bis-61-sein'}
pause()                           → {ok:true, pausiert:true}
resume()                          → {ok:true, pausiert:false}
stop() / startbildschirm()        → {ok:true}
aktion('richtung', {x:0,y:-1})    → {ok:true}
aktion('springen')                → {ok:false, fehler:'unbekannte-aktion'}
```

- `start()` antwortet bei Erfolg mit den **tatsächlich übernommenen** Werten
  (fehlende Felder bleiben auf der aktuellen Auswahl) und der Anzahl der
  Level/Stufen, wenn das Spiel welche hat.
- `pause()`/`resume()` antworten mit dem **neuen** Pausenzustand, auch wenn er
  sich nicht geändert hat.
- `whenReady()` ist der einzige asynchrone Weg; der Konnektor wartet darauf,
  statt zu pollen.

### 3.2 Feste Fehlercodes

| Code | Wann |
|---|---|
| `spiel-nicht-bereit` | Aufruf vor der Initialisierung |
| `konfiguration-muss-objekt-sein` | Zahl, Array, Funktion statt Objekt |
| `unbekanntes-feld` | Feld außerhalb der erlaubten Liste (Antwort nennt `feld`) |
| `unbekannte-stufe` | Stufe nicht in `einstellungen().stufen` |
| `stufe-muss-text-sein` | Stufe kein Text |
| `level-muss-1-bis-<n>-sein` | Level keine ganze Zahl im Bereich (Code nennt den Bereich) |
| `spiel-laeuft` | `start()` während einer laufenden Partie |
| `spiel-laeuft-nicht` | `pause()`/`resume()` ohne Partie (`stop()` ist dagegen wiederholbar und antwortet immer `{ok:true}`) |
| `unbekannte-aktion` | `aktion()` mit einem Namen außerhalb der Spielliste |
| `aktion-nicht-erlaubt` | Aktion im aktuellen Zustand nicht möglich |
| `daten-muessen-objekt-sein` | `daten` ist kein Objekt |

Regeln für die Prüfung:

- nur **bekannte Felder**, nur **echte Typen** (`'12'` ist kein Level, `true`
  ist kein Level, `null` ist kein Wert);
- `Object.prototype.hasOwnProperty` statt `in`;
- fehlende Felder bleiben auf der aktuellen Auswahl, `null`/`undefined` als
  ganzes Argument ebenfalls;
- bei einem Fehler bleibt der Spielzustand **unverändert** – keine halb
  übernommene Konfiguration, keine gestartete Partie.

### 3.3 Kapselung

- `zustand()` und `einstellungen()` geben **Kopien** zurück; verschachtelte
  Objekte und Listen (`richtung`, `extras`, `spieler`) werden ebenfalls
  kopiert. Mutationen an der Antwort ändern das Spiel nicht.
- Die Schnittstelle gibt **keine internen Spielobjekte** frei: keine Referenz
  auf Figuren, Spielfeld, Timer, Canvas oder Konfigurationsobjekte.
- Interne Namen bleiben intern; nur die Tabelle oben ist Zusage.

## 4. Ereignisse

Ereignisse laufen als `CustomEvent` auf `window`. Name: `<spiel>:<was>`, alles
klein geschrieben, abgeleitet vom globalen Namen (`SuperPimmelchen` →
`superpimmelchen`).

| Ereignis | Detail | Wann |
|---|---|---|
| `<spiel>:bereit` | `{version}` | genau einmal, nach der Initialisierung |
| `<spiel>:ende` | `{grund:'gewonnen'\|'verloren'\|'unentschieden', punkte, level, levelAnzahl, tick}` | genau einmal je ausgespielter Partie |
| `<spiel>:pausiert` | `{tick}` | wenn eine laufende Partie angehalten wird |
| `<spiel>:weiter` | `{tick}` | wenn eine angehaltene Partie fortgesetzt wird |

- `:ende` kommt **bevor** Highscore- oder Ergebnisfenster den Faden anhalten,
  damit der Konnektor das Ergebnis zuerst erhält.
- Kein Ereignis ohne Zustandsänderung: `pause()` in der Pause und `resume()`
  im Lauf lösen nichts aus.
- Die Details sind Kopien und enthalten die für das Ergebnis nötigen Zahlen.
  Spiele dürfen Felder ergänzen, nicht ersetzen.
- Ereignisse feuern auch ohne Zuhörer und auch im eigenständigen Betrieb; der
  Konnektor pollt nicht, er hört zu.
- Ein Abbruch über `stop()`/`startbildschirm()` ist **kein Ergebnis** und löst
  `:ende` nicht aus – der Aufrufer hat ihn selbst veranlasst.

## 5. Eingaben und Aktionen

- **Ein Eingabe-Seam:** Tastatur, Touch, Zeiger, Bot und spätere externe
  Spieleraktionen rufen **dieselben** spielinternen Aktionsfunktionen auf
  (`setDirection`, `fire`, `pause` …). Kein zweiter Weg an der Mechanik vorbei.
- **Externe Komponenten** – Konnektor, Sammlung, Test, Bot – lösen
  ausschließlich über `aktion(name, daten)` aus. Direkte Änderungen an
  internen Variablen sind verboten.
- Jedes Spiel dokumentiert seine Aktionsnamen (im Kopf der eigenen Datei oder
  in einem kurzen Abschnitt der Spieldoku): Name, erwartete `daten`, erlaubte
  Zustände.
- `aktion()` prüft wie `start()`: unbekannter Name → `unbekannte-aktion`,
  falscher Zustand → `aktion-nicht-erlaubt`, `daten` kein Objekt →
  `daten-muessen-objekt-sein`; der Zustand bleibt unverändert.
- **Tastatur:** Leertaste und Enter starten nur ohne fokussiertes
  Bedienelement. Steht der Fokus auf Knopf, Auswahl oder Feld (`button`,
  `select`, `input`, `textarea`, `a[href]`, `[contenteditable]`), löst der
  Browser selbst aus; das Spiel hält sich heraus. Kein `preventDefault` für
  Tasten, die das Spiel nicht benutzt.

## 6. Spielzustand

`zustand()` liefert eine flache Kopie mit Pflichtfeldern; Spiele dürfen Felder
ergänzen, nicht ersetzen:

```js
{
  laeuft, pausiert, tick, richtung: {x, y},
  level, punkte, extras: [{x, y, type}],
  startbildschirm
}
```

- `tick` zählt ausgeführte Spielzüge seit Partiebeginn (Neustart = 0). Er ist
  die spätere Vergleichsnummer für synchronisierte Partien.
- `startbildschirm` ist `true`, solange die Auswahl steht.
- `einstellungen().level` = **Auswahl** für den nächsten Start,
  `zustand().level` = **laufendes** Level. Nicht verwechseln.
- `einstellungen()` enthält keine Bedeutung der Werte (nicht „schwer = 20 s“),
  sondern nur die Möglichkeiten (Stufen, Levelbereich, Anzahl).

## 7. Spielübergreifende Regeln

- **Eindeutige globale Namen.** Ein Global je Spiel, keine generischen Helfer
  auf `window`. Gemeinsame Module behalten ihr `TG`-Präfix.
- **Eigene Speicherpräfixe.** Jeder `localStorage`-Schlüssel beginnt mit dem
  Spielnamen (`frogger-rekord`, `frogger-erreicht`, `superPimmelchenHighscore`);
  kein `rekord`, `level` oder `stil` als nackter Schlüssel. Neue Spiele nehmen
  die Schreibweise `<spiel>-…`. Ausnahme ist die bewusst spielweite Stilwahl
  `tg-stil` über `spiele/stil.js`. Die Schlüssel stehen in der Spieldoku;
  bestehende Schlüssel werden nicht umbenannt, sonst sind Spielstände weg.
- **Keine doppelten Spielkopien** mit derselben API-Version und denselben
  Speicherschlüsseln. Ist eine Sonderfassung nötig (Einzeldatei für den
  Offline-Betrieb), wird sie eingefroren, mit eigener Version und eigenem
  Speicherpräfix versehen und in der Doku genannt (bekannte Ausnahme:
  Abschnitt 11).
- **Keine blockierenden `alert()`-, `prompt()`- oder `confirm()`-Dialoge** für
  reguläre Spielergebnisse. Game over, Rekord, Pause und Levelwechsel laufen
  als Karte in der Seite und über `:ende`. Dialoge des Browsers sind auch im
  eingebetteten Rahmen nicht bedienbar.
- **Saubere Timer- und Listener-Verwaltung.** Timer-IDs werden gesammelt und
  bei `stop()`, Neustart und Spielende gelöscht; `requestAnimationFrame` läuft
  höchstens eine Schleife; Fenster- und Dokument-Listener werden mit
  `removeEventListener` abgemeldet. Pause hält den Spielfluss an (keine
  Zustandsänderung, keine Punkte, keine Zeit), auch wenn die Zeichenschleife
  weiterläuft. Keine Listener pro Bild, kein Wachstum über mehrere Partien.
- **Keine ungefragten Änderungen an anderen Spielen.** Wer ein Spiel umbaut,
  lässt die übrigen unangetastet; gemeinsame Dateien (`stil.js`,
  `spiel-start.js`, `zurueck.js`) werden nur erweitert, nicht umgedeutet.
- **Schrittweise Migration.** Bestehende Spiele werden im Rahmen ausdrücklich
  beauftragter Arbeiten angepasst, nicht pauschal und nicht „nebenbei“.

## 8. Zwingend und optional

**Zwingend (Abnahme):**

1. läuft eigenständig im Browser, ohne Konnektor und ohne `/api/`;
2. genau ein globaler Namensraum mit allen Feldern aus Abschnitt 3;
3. alle Aufrufe antworten, werfen nie, Fehlercodes aus Abschnitt 3.2;
4. ungültige Konfigurationen verändern den Zustand nicht;
5. `zustand()` und `einstellungen()` liefern unabhängige Kopien;
6. `:bereit`, `:ende`, `:pausiert`, `:weiter` mit den genannten Daten;
7. ein Eingabe-Seam für Tastatur, Touch und `aktion()`;
8. eigener Startbildschirm;
9. eigener Speicherpräfix, keine blockierenden Dialoge, aufgeräumte Timer;
10. Rückweg über `zurueck.js`.

**Optional (mit Begründung, darf Zwingendes nicht brechen):**

- `Object.freeze` auf der Schnittstelle (für neue Spiele empfohlen);
- Seed in `start()` für wiederholbare Partien;
- weitere Ereignisse (`:level`, `:punkte`) und weitere `zustand()`-Felder;
- weitere Aktionen über die Grundbedienung hinaus (Bot-Züge, Tipp, Undo);
- zusätzliche Konfigurationsfelder neben `stufe`/`level`;
- mehrere Grafstile über `TGStil` (Pflicht nur dort, wo das Spiel sie hat).

Abweichungen werden in der Spieldoku genannt, nicht stillschweigend geduldet.

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

## 10. Verboten

- Spiel ruft selbst `/api/` oder öffnet WebSockets; Spiel liest `room`, `name`,
  `id` aus der Adresse.
- Konnektor oder Sammlung kennen Levelbedeutung, Punkteregeln, Botlogik.
- Spiel öffnet `prompt`/`alert`/`confirm` für Ergebnisse (Ergebnis über
  `:ende`).
- Zwei Kopien desselben Spiels mit derselben API-Version und denselben
  Speicherschlüsseln.
- Externe Komponenten ändern interne Spielvariablen oder rufen interne
  Funktionen auf, die nicht in Abschnitt 3 stehen.
- Zweiter Startpfad neben `start(konfiguration?)`.
- `zustand()` liefert Referenzen statt Kopien.

## 11. Übergang und Migrationsstand

- **Referenz:** `spiele/pimmel/spimmel.html`, Schnittstelle **2.0.0** –
  vollständig: `version`, `bereit`, `whenReady()`, `start()`, `einstellungen()`,
  `zustand()`, `startbildschirm()`, `stop()`, `pause()`, `resume()`,
  `aktion(name, daten)` sowie `superpimmelchen:bereit|ende|pausiert|weiter`.
  `einstellungen()` nennt nur die Möglichkeiten (`stufe`, `stufen`, `level`,
  `levelAnzahl`), keine Sekunden; das Objekt ist eingefroren, alle Antworten
  sind Kopien.
- **Pimmel-Aktionen:** `richtung {x,y}` (genau eine Achse ±1, nur im Spiel),
  `feuer`, `start`, `pause`, `weiter`, `neu` – alle über dieselben internen
  Funktionen wie Tastatur und Touch.
- **Ein `richtung`-Aufruf ist genau ein Schritt (26.11.2026):** Er dreht und
  geht **einen** Takt weit, danach steht die Schlange wieder – ein Aufruf hält
  keine Richtung. Tastatur und Steuerkreuz drehen über dieselbe Prüfung und
  laufen nur weiter, solange sie gehalten werden (`keydown`/`pointerdown` bis
  `keyup`/`pointerup`). Vorher startete `aktion('richtung')` denselben
  Dauerlauf, der beim Fortsetzen schon entfernt wurde. Geprüft: Ist eine
  Richtung gesperrt (Wand oder Brücke), bleibt der Schritt aus – genau wie bei
  der Taste; die Antwort nennt das in `schritt: false`. Kehrtwenden verwirft
  die Prüfung wie bei der Tastatur, sobald die Schlange länger als ein Feld
  ist. Schnittstelle bleibt **2.0.0** (neues Feld nur in der Antwort).
- **Fortsetzen bewegt nicht von selbst (26.11.2026):** `resume()` und
  `aktion('weiter')` heben nur die Pause auf und stellen keinen Bewegungstakt
  auf. Gefahren wird wie außerhalb der Pause ausschließlich, solange eine
  Richtung gehalten wird (`keydown`/`pointerdown` bis `keyup`/`pointerup`);
  `pause()` vergisst gehaltene Steuerkreuz-Finger. `zustand().richtung` bleibt
  über die Pause hinweg gleich, die erste Eingabe danach zählt genau einen
  Takt. Gilt für alle vier Richtungen gleich; Schnittstelle bleibt **2.0.0**.
- **Drei Spielwelten bei Pimmel (26.11.2026):** `start({ stil })` nimmt
  `normal`, `maedchen` oder `buben`; unbekannt oder fehlend heißt `normal`
  (kein Fehler). Eine Konfiguration mit ausschließlich `stil` richtet die Welt
  ein und lässt den Startbildschirm stehen – so übergibt der Connector die
  Wahl aus `spiele/stil.js` beim Öffnen. Das Spiel liest keinen globalen Stil
  mehr und hat keine eigene Stilwahl; `einstellungen()` nennt `stil` und
  `stile`, `zustand()` nennt `stil`. `normal` ist der grün-goldene
  Kasino-Look; `stil.js` bleibt bei seinen drei Werten.
- **Stilwahl beim direkten Start (26.11.2026):** Ohne Connector zeigt
  Pimmel auf dem Startbildschirm die Kachel `GRAFIKSTIL` mit `NORMAL`,
  `MÄDCHEN`, `BUBEN` (`#stil-wahl`). Ein Klick stellt die Welt sofort um
  (Vorschau auf Startbildschirm und Leinwand) und merkt sie unter
  `superPimmelchenStil`; der nächste direkte Start öffnet damit. Kommt der
  Stil von außen – `start({ stil })` mit `stil`-Feld –, hat er Vorrang: Die
  Kachel verschwindet (`#stil-block` bleibt verborgen, auch nach
  `startbildschirm()`), und die lokale Wahl wird weder gelesen noch
  überschrieben. Eine Konfiguration ohne `stil` (z. B. nur `stufe`) lässt die
  Kachel ebenfalls weg. `einstellungen()` und `zustand()` nennen weiterhin nur
  `stil` und `stile`; die Schnittstelle bleibt **2.0.0**.
- **Pimmel-Highscore ohne Browserdialog:** Der Highscore läuft seit 26.11.2026
  als Karte in der Seite (`#highscore-blatt`) mit Rangliste, Namensfeld,
  „Speichern“ und „Ohne Namen“; `prompt` und `alert` sind im Spiel vollständig
  entfernt. Für fehlerhafte Aktionsdaten (diagonal, `0/0`, Text) gibt es in
  Abschnitt 3.2 weiter keinen Code – Pimmel antwortet mit
  `aktion-nicht-erlaubt`, fremde Felder mit `unbekanntes-feld`.
- **Zweite Fassung mit eingebetteten Grafiken:**
  `spiele/pimmel/Super-Pimmelchen-61-Level-Einzeldatei.html` hat dieselbe
  Schnittstelle 2.0.0, dieselbe Spiellogik, dieselben 61 Level, dieselben
  Aktionen und dieselben drei Spielwelten wie die Referenz; unterschiedlich
  ist nur die Einbindung: Grafiken, Schriften und `zurueck.js` stecken in der
  Datei. Sie wird aus `spimmel.html` gebaut (Bauskript in
  `Documents\arbeit\tmp`), damit die beiden Fassungen nicht auseinanderlaufen.
- **Bekannte Abweichung:** Diese Einzeldatei teilt Stufe und Highscore mit
  `spimmel.html` (`superPimmelchenStufe`, `superPimmelchenHighscore`). Beide
  Seiten im selben Browser beeinflussen sich also gegenseitig. Getrennte
  Schlüssel wären ein eigener Auftrag, weil sie bestehende Spielstände
  betreffen – bis dahin bleibt es so.
- **Noch ohne öffentliche Schnittstelle:** `spiele/frogger/frogger.html`
  (Neugestaltung 26.11.2026) – beim nächsten beauftragten Schritt bekommt es
  `window.Frogger` nach diesem Standard.
- **Bestehende Online-Spiele** holen Raum und Kennung heute noch aus der
  Adresse (`?room=…&id=…`). Der Standard verbietet das; der Umbau erfolgt
  spielweise mit dem jeweiligen Auftrag, nicht vorher.
- **Neue Spiele** erfüllen den Standard von Anfang an.
- **Bestehende Spiele** werden nur im Rahmen ausdrücklich beauftragter
  Arbeiten angepasst, nicht pauschal.
- **Zweite Fassung nach dem Standard (26.11.2026):**
  `spiele/brettspiele/tictac/tictactoe.html`, Schnittstelle **2.0.0** wie
  Pimmel – `version`, `bereit`, `whenReady()`, `start()`, `einstellungen()`,
  `zustand()`, `startbildschirm()`, `stop()`, `pause()`, `resume()`,
  `aktion(name, daten)`, Ereignisse `tictactoe:bereit|zug|ende|pausiert|weiter|neu`.
  Das Spiel ruft kein `/api/` mehr auf und liest `room`/`id` nicht aus der
  Adresse: Der Connector schiebt den Stand mit `aktion("stand", …)` hinein und
  horcht auf `tictactoe:zug` und `tictactoe:neu`. Ohne Antwort des Servers
  bleibt das Brett stehen, wie es war – ein Verbindungsabriss kann so keinen
  falschen Stand erzeugen.
- **Tic-Tac-Toe-Aktionen:** `setzen {feld}` (0–8, nur im Spiel und nur, wenn
  man dran ist), `stand {mark,gegner,marke,brett,amZug,status,sieger,hinweis}`
  (nur Online), `start`, `pause`, `weiter`, `neu` – dieselben internen
  Funktionen wie Maus, Taste und Touch. Doppelte oder fremde Züge antwortet
  das Spiel mit `aktion-nicht-erlaubt`; der Server weist ein belegtes Feld mit
  **409** „Der andere Spieler ist an der Reihe.“ ab.
- **Startbildschirm bei Tic-Tac-Toe (26.11.2026):** Zwei wählbare Modi –
  `lokal` (zwei Spieler am Gerät) und `bot` (gegen den Computer). Online gibt
  es nur über den Connector: `start({online:{…}})` setzt Marke, Gegner und
  Stil, eine Wahl dafür gibt es auf dem Startbildschirm nicht. `start()` ohne
  Daten startet den gewählten Modus, Vorgabe `lokal`; `start({stil, online})`
  verhält sich wie bisher. Die Stufenleiste erscheint nur bei `bot`, die
  Stilwahl nur beim direkten Start ohne Connector. Neue Felder sind rein
  lesend: `einstellungen()` nennt zusätzlich `modi` (`["lokal","bot"]`),
  `modus`, `botStufen`, `stufe` und `stilWahl`, `zustand()` zusätzlich `modus`,
  `stufe`, `stilWahl`, `botMark`, `botDenkt`. Die Schnittstelle bleibt
  **2.0.0**.
- **Bot gehört zum Spiel:** `tictactoe.html` rechnet alle drei Stufen selbst
  (`leicht`: nimmt einen Sieg meist, blockt nie; `mittel`: Sieg → Block →
  Mitte → Ecke; `schwer`: vollständiges Minimax). Der Bot spielt Kreis und
  zieht über dieselbe Prüfung wie ein Mensch; `tictactoe.go` und
  `bot_tictactoe.go` im Connector sind unverändert und werden vom Spiel nicht
  benutzt.
- **Drei Welten bei Tic-Tac-Toe:** `start({stil})` nimmt `normal`, `maedchen`,
  `buben`. Beim direkten Start wählt der Spieler selbst (gespeichert in
  `ticTacToeStil`, ohne Wahl `normal`); im Connector-Rahmen ist die Wahl
  verborgen und der Stil kommt von außen (`start({stil})` bzw.
  `TGStil`/`data-stil` des Rahmens) – dort meldet `einstellungen().stilWahl`
  `false`. Der Stil ist rein lokal (eigenes `data-stil` im Rahmen), wird nie
  als Zug gesendet und bleibt im Mehrspieler je Spieler getrennt – geprüft mit
  zwei Sitzungen in derselben Partie: A `maedchen`, B `buben`, beide Brett und
  Ergebnis gleich, keine Wahl im Spiel sichtbar.
- **Symbole je Stil (26.11.2026):** Die Marken X und O werden als SVG
  gezeichnet, je eigener Ansicht: `normal` Kreuz/Kreis, `maedchen`
  Herz/Teddy, `buben` Pentagramm/Totenkopf. Die Zuordnung hängt nur an Marke
  und lokalem Stil – im Connector-Rahmen zeichnet jeder Spieler seine Fassung,
  das Brett (`brett`, `amZug`) bleibt unverändert. `zustand().stil` nennt
  weiter den eigenen Stil; die Symbole sind reine Anzeige.
- **Speicher:** `ticTacToeBilanz` (Siege, Niederlagen, Remis am Gerät) und
  `ticTacToeStil` (eigene Stilwahl beim direkten Start) je Browser; im
  Connector-Rahmen gilt weiter allein `tg-stil` des Rahmens.
- **Zweite Fassung nach dem Standard (26.11.2026):**
  `spiele/brettspiele/schiffe/schiffe.html`, Schnittstelle **2.0.0** wie
  Pimmel – `version`, `bereit`, `whenReady()`, `start()`, `einstellungen()`,
  `zustand()`, `startbildschirm()`, `stop()`, `pause()`, `resume()`,
  `aktion(name, daten)`, Ereignisse `schiffe:bereit|ende|pausiert|weiter`.
  Die fünf Schiffsgrafiken stecken als Daten-URI in der Datei; es gibt keinen
  `/api/`-Aufruf und kein `room`/`id` in der Adresse. Online kommt die Partie
  über `start({online:{id,name,gegner,marke}})` und `aktion("stand", …)`;
  `zustand().online` meldet den Rahmen.
- **Schiffe-Aktionen:** `schiff {i}` (0–4), `drehen`, `vorschau {x,y}`,
  `setzen`, `fertig` (nur in der Setzphase), `schuss {x,y}` (nur im Spiel,
  nur wenn man dran ist, nur unbeschossene Felder), `neu`, `stand`, `start`,
  `pause`, `weiter` – alle über dieselben internen Funktionen wie Maus,
  Ziehen, Touch und die Taste R. Online werden `fertig` und `schuss` zu
  Ereignissen (`schiffe:setzen`, `schiffe:zug`), `stand` nimmt den Stand des
  Connectors an; alles andere antwortet dort `aktion-nicht-erlaubt`. Fehlercodes nach Abschnitt 3.2; falscher Zustand
  antwortet mit `aktion-nicht-erlaubt`, ein belegtes Feld mit
  `aktion-nicht-erlaubt` statt eines zweiten Schusses. `zustand()` ergänzt
  `modus`, `stufe`, `stil`, `stilWahl`, `phase` (`start|setzen|spiel|ende`),
  `amZug`, `botDenkt`, `eigen` (`fertig`, `schiffe`, `beschuss`), `feind`
  (`name`, `beschuss`), `treffer`, `versenkt`, `eigeneVerluste`, `sieger`.
- **Flotte 5, 4, 3, 2, 1 (26.11.2026):** Flugzeugträger (5), Zerstörer (4),
  Patrouillenboot (3), U-Boot (2), Ente (1) – insgesamt 15 Trefferflächen auf
  10 × 10 Feldern. Dieselbe Liste steht in `einstellungen().schiffe`, in der
  Aufstellungsanzeige, im Bot und im Connector: `schiffeFlotte` in
  `spiele/connector/schiffe.go` (vorher 5, 4, 3, 3, 2), `schiffeSchiffName`
  (5 Flugzeugträger … 1 Ente) und `schiffeBotFlotte` sind angepasst, beide
  Binaries neu gebaut. Die Grafik füllt die Kästchenlänge; die Höhe folgt
  unverzerrt aus dem Seitenverhältnis (Träger 1,86 Kästchen), die Wasserlinie
  bleibt unten in der eigenen Reihe.
- **Bot gehört zum Spiel:** drei Stufen `leicht`, `mittel`, `schwer`
  (`einstellungen().stufen`). Leicht rät blind, mittel verfolgt Treffer
  (Nachbarfelder, Verlängerung einer Zweierreihe), schwer rechnet die
  Wahrscheinlichkeitsdichte über alle noch möglichen Lagen der Restflotte
  (mit offenen Treffern nur Lagen, die einen davon berühren). Keine Stufe
  liest die gegnerische Aufstellung; sie kennt nur das eigene Beschussbild
  und die noch fahrenden Längen. Gemessen, 400 Partien je Stufe: **94,6 /
  68,8 / 63,5** Schüsse bis alle fünf versenkt (Trefferquote 15,9 / 23,5 /
  26,1 %). `schiffe.go` und `bot_schiffe.go` im Connector bleiben für den
  Online-Betrieb unverändert; lokal werden sie nicht benutzt.
- **Drei Stile bei Schiffe versenken:** `start({stil})` nimmt `normal`,
  `maedchen`, `buben` (unbekannt oder fehlend heißt `normal`). Beim direkten
  Start wählt der Spieler selbst (Kachel `#stilWahl`, gespeichert in
  `schiffeStil`); kommt der Stil von außen, ist die Kachel verborgen, auch
  nach `startbildschirm()`, und `einstellungen().stilWahl` meldet `false`.
  Der Stil ist rein lokal, wird nie als Zug gesendet und bleibt im
  Mehrspieler je Spieler getrennt.
- **Startbildschirm bei Schiffe versenken:** eigener Bereich mit Bot-Stufe,
  Grafikstil, Grafikvorschau und „Spiel starten“; `einstellungen().modi`
  nennt nur `["bot"]`. Online gibt es keine Auswahl – dort kommt die Partie
  vom Connector.
- **Online über Relay wie Tic-Tac-Toe (26.11.2026):** Der Connector besitzt
  Raum, Kennung, Einladungen und die Abfrage beim Server; die Spielregeln
  liegen im Spiel. Er schiebt Kennung, Name, Gegner und den Stand hinein
  (`start({online})`, `aktion("stand")`) und horcht auf die Absichten des
  Spiels: `schiffe:setzen {schiffe}` (Flotte fertig), `schiffe:zug {x,y}`
  (Schuss), `schiffe:neu` (nach dem Ende). Brücke und Abfrage stehen in
  `spiele/connector/connector.html` (`schiffeSpiel`, `schiffeSpielstand`,
  `schiffeBrueckeAn`, 1-s-Takt nur bei offenem Schiffe-Rahmen). Das Spiel
  ruft selbst kein `/api/` auf: ohne neuen Stand bleibt das Brett stehen, ein
  zweiter Schuss vor der Antwort wird verworfen (`zugGesendet`). Aufstellung,
  Zugrecht und Treffer prüft weiter der Server; ein fremder oder doppelter
  Zug kommt als Fehlertext zurück und der Stand bleibt unverändert.
  `stop()` und `startbildschirm()` antworten online mit
  `{ok:true, ignoriert:true}`, „Zum Start“ ist dort ausgeblendet.
- **Goldminer nach dem Standard (26.11.2026):**
  `spiele/goldminer/goldminer.html`, Schnittstelle **2.0.0** wie Pimmel –
  `version`, `bereit`, `whenReady()`, `start()`, `einstellungen()`,
  `zustand()`, `startbildschirm()`, `stop()`, `pause()`, `resume()`,
  `aktion(name, daten)`, Ereignisse `goldminer:bereit|ende|pausiert|weiter`.
  Das Spiel ruft kein `/api/` auf und liest `room`/`id` nicht aus der Adresse;
  der einzige globale Name ist `window.GoldMiner`.
- **Goldminer-Aktionen:** `werfen` (Haken auswerfen; nur im Spiel, nicht in der
  Pause), `start`, `pause`, `weiter`, `neu` – alle über dieselben internen
  Funktionen wie Knopf, Leertaste, Klick und Touch. `zustand()` ergänzt
  `levelAnzahl`, `ziel`, `zeit`, `phase`, `zieht`; Fehlercodes nach
  Abschnitt 3.2, `stufe` gibt es nicht (keine Schwierigkeitsstufen).
- **Grafik bei Goldminer:** Himmel, Erde, Vordergrund und Winde sind
  Canvas-Funktionen in der Datei (kein CDN, keine Bilddateien). Die drei
  Weltansichten blieben erhalten; am 27.11.2026 wurden Gold, Gestein und der
  Beutel neu gezeichnet: feste, unregelmäßige Silhouetten mit stileigenen
  Paletten, Facetten, Glanz, Rissen und Gesteinskrümeln. Die Pfade werden
  einmal je Objekt gebaut; die Startvorschau zeigt einige der Fundstücke.
  Regeln, Trefferradius und Laufwege blieben unverändert.
- **Stil bei Goldminer:** `start({stil})` stellt **nur** die Welt ein (so
  übergibt der Connector beim Öffnen und beim Wechsel, ohne die Partie neu zu
  starten); `start()` und `start({level})` beginnen eine Partie. Beim direkten
  Start wählt der Spieler auf dem Startbildschirm (`#stilKnoepfe`, gemerkt in
  `goldminerStil`); kommt der Stil von außen (Rahmen/TGStil oder
  `start({stil})`), ist die Wahl verborgen (`einstellungen().stilWahl false`).
  `normal` zeichnet die Ansicht `klassisch`, `buben` die Ansicht `jungs` – die
  Ansichtstabelle bleibt unverändert.
- **Startbildschirm bei Goldminer:** eigener Bereich nach Abschnitt 9 mit
  Titel, echter Minenvorschau (`#startSzene`, dieselben Ansichtsfunktionen),
  der Stilwahl Normal/Mädchen/Buben und „Graben beginnen“. Die frühere
  Marke, der Erklärungstext, die Steuerungsliste und die Info-Karten sind
  entfernt. Ohne inneres Scrollen bei 320×568, 360×640, 390×844, 844×390,
  1024×600 und 1280×800. Auf flachen Geräten (Höhe ≤ 680) entfällt die
  Vorschau, der Knopf bleibt sichtbar.
- **Connector:** `GoldMiner` steht in `SPIEL_API_NAMEN` (`connector.html` in
  `connector` und `connector-online`, byte-gleich). Geprüft im echten
  Connector (Port 8779): Stil „Mädchen“ kam im Rahmen an (Kachel verborgen),
  `start()` lief; Umschalten auf „Buben“ wechselte die Welt bei laufender
  Partie, Level und Takt blieben gleich.

- **Solitär nach dem Standard (26.11.2026):**
  `spiele/kartenspiele/solitaer/solitaer.html`, Schnittstelle **2.0.0** –
  `version`, `bereit`, `whenReady()`, `start()`, `einstellungen()`,
  `zustand()`, `startbildschirm()`, `stop()`, `pause()`, `resume()`,
  `aktion(name, daten)`, Ereignisse `solitaer:bereit|ende|pausiert|weiter`;
  einziger globaler Name `window.Solitaer`. Die fünf Varianten Klondike,
  Spider, FreeCell, Pyramid und TriPeaks bleiben unverändert. `zustand()`
  ergänzt `modus`, `modi`, `zuege`, `zeit`, `stil` und `beendet`; `richtung`
  bleibt `{x:0,y:0}`, `level` ist 1. Nach dem Ergebnis bleiben Züge, Punkte
  und Zeit stehen.
- **Aktionen bei Solitär:** `start` (Partie mit der aktuellen Wahl), `modus`
  `{modus}` (Variante beginnen), `neu`/`mischen` (frisch gemischt), `pause`,
  `weiter`, `tipp`, `zurueck`, `regeln`, `statistik` – alle über dieselben
  internen Funktionen wie Knopf, Taste und Finger. Fehlercodes nach
  Abschnitt 3.2; eine unbekannte Variante antwortet mit dem eigenen Code
  `unbekannter-modus` (Solitär hat keine Stufen – die einzige Abweichung).
- **Drei Stile bei Solitär:** `start({stil})` nimmt `normal`, `maedchen`
  oder `buben` (Unbekanntes heißt `normal`) und stellt **nur** das Kartenbild
  ein, ohne die Partie zu beginnen. Beim direkten Start wählt der Spieler in
  `#stilWahl` (gemerkt in `solitaer-stil`, die zentrale Wahl `tg-stil` wird
  nur gelesen); kommt der Stil von außen, ist die Kachel verborgen, auch nach
  `startbildschirm()`, und `einstellungen().stilWahl` meldet `false`.
- **Eigenständig:** Startbildschirm mit Stilwahl und fünf Varianten, Pause
  als eigener Schleier (`#pauseSchleier`, Taste P, Uhr steht, Brett
  gesperrt), Statistik je Variante, „Neu mischen“, Rückgängig, Tipp, Regeln,
  Automatisch fertigspielen, Gewinn- und Verlustfenster; der Rückweg läuft
  über `window.tgZurueck()` (`zurueck.js` mit `data-kein-knopf="1"`).
- **Kartenbilder:** Normal Elfenbein mit Goldrand und Nachtblau-Rückseite
  (Ovalmedaillon), Mädchen Rosé mit runden Ecken, gepunkteter Innenlinie und
  Rautengitter, Buben Anthrazit mit harter Kante und Kohlefaser-Winkeln –
  eigene Vorderseite, Rückseite, Radius, Schrift und Dekor je Stil. Bei 33 px
  breiten Spider-Karten bleibt der Rang 12,5 px lesbar.
- **Connector:** `Solitaer` steht in `SPIEL_API_NAMEN` (`connector.html` in
  `connector` und `connector-online`, byte-gleich, NUL-Zahl unverändert 2).
  Geprüft im echten Connector (Port 8765, IPv6): Katalog enthält „Solitär |
  kartenspiele/solitaer/solitaer.html“, der Rahmen öffnete die Seite, Stil
  „Mädchen“ kam an (`stilWahl:false`, Kachel verborgen), die Partie wurde im
  Rahmen gestartet (1 Zug), Umschalten auf „Buben“ wechselte das Kartenbild
  bei laufender Partie (Züge blieben 1), danach „klassisch“ ebenso.
- **Geprüfte Partien:** TriPeaks echt durchgespielt – 8 Partien, 7-mal
  „Keine Züge mehr“ (`solitaer:ende grund:'verloren'`) und eine gewonnene
  Partie (51 Züge, `grund:'gewonnen'`, Punkte 280, Ergebnisfenster).
  Klondike: Nachziehstapel, Rückgängig, Tipp, Neumischen. Handy 390×844:
  kein Überlauf (Seitenbreite 390), keine Karte außerhalb, alle Stapel
  sichtbar; PC 1280×800: Karten 151 px breit, Bahn 1118 px.

- **Zweite Fassung nach dem Standard (27.11.2026):**
  `spiele/brettspiele/zehntausend/zehntausend.html`, Schnittstelle **2.0.0**
  wie Pimmel – `version`, `bereit`, `whenReady()`, `start()`,
  `einstellungen()`, `zustand()`, `startbildschirm()`, `stop()`, `pause()`,
  `resume()`, `aktion(name, daten)`, Ereignisse
  `zehntausend:bereit|ende|pausiert|weiter` (zusätzlich `wurf`, `zug`, `neu`).
  Kein `/api/`, kein `room`/`id` aus der Adresse, `token.js` und
  `ergebnis.js` sind entfernt; der Rückweg läuft über `zurueck.js`.
- **10.000-Start und -Modi:** eigener Startbildschirm mit drei Stilen,
  `bot`/`lokal`, 2–6 Spielern und je Platz Mensch/Bot. `start({stil})`
  richtet nur die Welt ein und lässt die Auswahl stehen (Connector-Weg);
  `start({spieler, modus, anzahl})` beginnt eine Partie; `start()` ohne Angaben übernimmt die sichtbare Auswahl. `einstellungen()`
  nennt `modi`, `modus`, `anzahl`, `stile`, `stil`, `stilWahl`, `spieler`,
  `ziel`, `einstieg`; `zustand()` ergänzt `modus`, `stil`, `stilWahl`,
  `spieler`, `dran`, `amZug`, `wuerfel`, `gesichert`, `auswahl`,
  `wurfPunkte`, `zugPunkte`, `wuerfe`, `botDenkt`, `ziel`, `einstieg`,
  `beendet`, `sieger`.
- **10.000-Aktionen:** `wuerfeln`, `weiterwuerfeln`, `sichern`,
  `waehlen {index}` (0–5), `wertung {wuerfel}` (reine Prüffunktion),
  `pause`, `fortsetzen`, `neu`, `start`, `regeln` – alle über dieselben
  internen Funktionen wie Maus, Touch, Tastatur und Bot. Eigene
  Startfehlercodes: `spieler-muessen-2-bis-6-sein`,
  `spieler-braucht-mindestens-einen-menschen`,
  `modus-muss-bot-oder-lokal-sein`, `anzahl-muss-2-bis-6-sein`.
- **10.000-Wertung (fest):** 1 = 100, 5 = 50, Dreierpasch 1er = 1.000, sonst
  Augenzahl × 100, jeder weitere gleiche Würfel zählt den Satz erneut; nur
  1/5/ab drei gleiche zählen; Einstieg 350, Ziel 10.000. Die Regel steht im
  Spiel und im Regelblatt.
- **10.000-Bot:** gehört zum Spiel, spielt über dieselben Aktionen, legt alle
  wertbaren Würfel beiseite und sichert ab 350 (früher bei wenigen freien
  Würfeln oder ab 1.000). Kein Server-Bot im lokalen Betrieb.
- **10.000-Speicher:** `tg-zehntausend-spiel` (laufende Partie, alter
  Schlüssel bewusst behalten) und `zehntausendStil` (eigene Stilwahl).
- **10.000-Connector:** `Zehntausend` steht in `SPIEL_API_NAMEN` beider
  `connector.html`; der Rahmen übergibt den Stil, die Stilkachel ist dann
  verborgen. **Offen:** eine 2.0-Online-Brücke (Lobby, `start({online})`,
  `aktion("stand")`) gibt es für 10.000 noch nicht; die alte `/api/`- und
  `?room&id`-Logik ist auftragsgemäß entfernt. Die Online-Kachel des
  Registers bleibt vorerst bestehen und öffnet die lokale Auswahl.

## 12. Prüfliste für ein neues Spiel

1. Ohne Konnektor geöffnet: Startbildschirm, Start, Pause, Neustart, Ende.
2. Globale Namen geprüft: genau ein `window.<SpielName>`, keine weiteren.
3. `start()` mit gültiger, fehlender, falscher und unbekannter Konfiguration –
   Zustand nach jedem Fehler unverändert.
4. `whenReady()` vor und nach der Initialisierung.
5. `pause()`/`resume()` im Lauf und ohne Partie, inklusive Ereignissen
   `:pausiert`/`:weiter` (genau einmal je Wechsel).
6. `:ende` genau einmal, mit Ergebnis, vor dem Ergebnisfenster.
7. `zustand()` zweimal holen, die erste Antwort verändern – Spiel unberührt.
8. Tastatur, Touch und `aktion()` lösen dieselbe Mechanik aus.
9. `localStorage`-Schlüssel mit Spielpräfix; kein `alert`/`prompt` im Pfad.
10. Mehrere Partien hintereinander: keine wachsenden Timer oder Listener.
