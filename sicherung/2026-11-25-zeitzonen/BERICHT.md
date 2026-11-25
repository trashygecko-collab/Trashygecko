# Serverzeit und Zeitzonen (25.11.2026)

Commit `7572e52` (GitHub), beide Ablagen gleich.

## Fehlerbild und Ursache

Am 25.11.2026 zeigte die deutsche Ortszeit ~19 Uhr, der Chat im Connector ~18 Uhr.

Ursache: Alle Server-Zeitstempel entstanden mit

```go
time.Now().Format("15:04:05")
```

`Format` ohne Zone nimmt die Zone **des Prozesses**. Auf dem eigenen PC ist das
`Europe/Berlin`, im Render-Container (Alpine) aber UTC. Im Winter fehlte damit
eine Stunde, im Sommer zwei. Betroffen waren **216 Stellen** in beiden
Connector-Fassungen (`spiele/connector`, `spiele/connector-online`):
Chat-Nachrichten, Begrüßungen, Umbenennen, Admin-Meldung, Spielmeldungen aller
Spiele, Gewinn-/Ergebniszeilen, die Raumliste (`room.Created.Format("15:04")`)
und das Quiz-Datum (`Format("2006-01-02")`).

Kein Fehler war die Rechnung: Fristen, Rundenzeiten, `time.Since`, `Sub`, `Add`,
Bot-Takt und die Quiz-Restzeit arbeiten mit absoluten Zeitpunkten und sind
zonenneutral. Der Fehler saß ausschließlich in der Ausgabe.

## Änderungen

- **Neu `spiele/connector/zeit.go`** (und byte-gleich in `connector-online`):
  - `uhrzeit(t)` → `15:04:05` in `Europe/Berlin`
  - `uhrzeitKurz(t)` → `15:04` (Raumliste)
  - `datum(t)` → `2006-01-02` (Tageswechsel, Quizstand)
  - `anzeigeZone` wird einmalig über `time.LoadLocation("Europe/Berlin")`
    geladen; `import _ "time/tzdata"` bettet die Zeitzonendaten ein, weil das
    Alpine-Abbild keine `/usr/share/zoneinfo` hat (Notnagel `FixedZone` MEZ).
- Alle 216 Aufrufe umgestellt: `time.Now().Format("15:04:05")` → `uhrzeit(time.Now())`,
  `x.Format(...)` → `uhrzeit(x)`, `room.Created.Format("15:04")` → `uhrzeitKurz(...)`,
  Quizstand → `datum(time.Now())`.
- **Gespeichert wird weiter ein Zeitpunkt** (`time.Time` bzw. RFC3339 mit
  Offset in Sitzungen/Risiko-Ständen). Die Anzeige-Zeichenkette im Chat ist die
  einzige Ausnahme – sie wird beim Speichern einmalig lokalisiert.
- **Frontend rechnet nichts um:** `connector.html` zeigt `m.time` roh an
  (vier Stellen), also keine doppelte Verschiebung.
- **`quizshow/fragen.html`:** Export-Datum kam aus `toISOString()` (UTC) und war
  zwischen 00:00 und 01:00 (Sommer bis 02:00) noch der Vortag – jetzt
  `Intl.DateTimeFormat("sv-SE", {timeZone:"Europe/Berlin"})`.
- **`rotlicht/rotlicht-oberflaeche.js`:** Chronik-Uhrzeit mit
  `timeZone:"Europe/Berlin"` statt Browserzone (Datei liegt im noch nicht
  committeten `spiele/rotlicht/` der vorigen Aufgabe).

Nicht angefasst: Serveruhr, NTP, Daten, Regeln, Spielstände, HTTP-Header
(`Date`/`Last-Modified` bleiben GMT nach Standard).

## Tests

`spiele/connector/zeit_test.go` und `zeit_chat_test.go` (in beiden Fassungen):

| Test | Prüfung | Ergebnis |
| --- | --- | --- |
| `TestUhrzeitWinterzeit` | 15.01.2026 18:30 UTC → `19:30:00` | grün |
| `TestUhrzeitSommerzeit` | 15.07.2026 17:30 UTC → `19:30:00` | grün |
| `TestUhrzeitZeitumstellung` | 29.03. 00:59 UTC → `01:59`, 01:00 UTC → `03:00`; 25.10. 00:30/01:30 UTC → `02:30`/`02:30` | grün |
| `TestDatumTageswechsel` | 25.11. 23:30 UTC → `2026-11-26`; 22:30 UTC → `2026-11-25`; 15.07. 22:30 UTC → `2026-07-16` | grün |
| `TestDatumJahreswechsel` | 31.12.2026 23:30 UTC → `2027-01-01`, `00:30:00` | grün |
| `TestUhrzeitKurz` | 15.07. 22:10 UTC → `00:10` | grün |
| `TestAnzeigeUnabhaengigVonProzesszone` | `time.Local` = UTC / MESZ / Berlin → immer `19:05:00`; Berliner Zeit wird nicht doppelt verschoben | grün |
| `TestZeitpunkteBleibenAbsolut` | gleicher Zeitpunkt in zwei Zonen: `Equal`, `Sub` 0, gleiche Anzeige | grün |
| `TestChatUndRaumzeitSindDeutschImUTCBetrieb` | Prozesszone UTC: `POST /api/messages` liefert `20:48:38` statt UTC-`19:48:38`, Verlauf ebenso; `GET /api/rooms` nennt `created` in Ortszeit | grün |

Gesamtläufe: `go test ./...` in `connector` **195 s**, in `connector-online`
**201 s**, beide exit 0; `go vet` 0; `gofmt -l` sauber (außer dem alten
CRLF-Test `bot_steuerung_test.go`, unverändert).

## Live-Prüfung (Render, nach dem Deploy)

- Container läuft in UTC (Render/Alpine). `GET /api/rooms` um **20:52:26**
  Ortszeit: Hauptraum `TG-7AF9` (Besitzer „Tanja", 2 Spieler) mit
  `created: "20:52"` – im Container war es 19:52 UTC. Die Anzeige folgt also
  der deutschen Ortszeit, nicht mehr der Serverzone.
- Browser im Test: `Intl … Europe/Berlin` 20:52:35, `UTC` 19:52:35 – die
  angezeigte Spielzeit stimmt mit der Browserzeit überein.

## Grenzen

- Der Fehler ist nur im Container sichtbar (Prozesszone UTC). Nachgestellt
  wurde er über `time.Local = time.UTC` in den Tests – auf dem Windows-PC ist
  die Prozesszone ohnehin Berlin, dort war die Anzeige schon richtig.
- Nicht geprüft: eine Sommerzeit-Umstellung im laufenden Betrieb (die Tests
  prüfen die Zeitpunkte beiderseits der Umstellung, nicht den Live-Betrieb um
  02:00 Uhr).
