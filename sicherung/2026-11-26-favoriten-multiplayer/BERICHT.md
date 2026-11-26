# Favoriten-Kacheln: Mehrspieler wieder erreichbar (26.11.2026)

## Der Fehler

Die sechs Bild-Kacheln über der Spieleübersicht („Highlights“/Favoriten) hatten
einen eigenen Startweg: `openGame(...)` mit Raum und Kennung, direkt in die
Partie. Damit fehlte der Tisch, den „Online spielen“ auf der Spielkarte anlegt
(`POST /api/angebot`): Andere Spieler sahen kein Angebot, niemand konnte
beitreten; die geöffnete Partie wartete ohne Runde. Dasselbe galt für
`?spiel=…` aus der Spielhalle – der Connector öffnete das Spiel nach 350 ms
selbst.

Die Spielhalle selbst hatte den Weg in den Connector ganz verloren: Beim
Vorab-Rendern von `spiele/spiele.html` (Commit `c24f11c`) fielen
`multiplayer-register.js` und `halle-online.js` aus der Seite. Jede Kachel –
Favorit wie normale Karte – verlinkte danach direkt auf die Spielseite, auch
bei den 71 Spielen mit Mehrspieler-Modus.

## Was geändert wurde

`spiele/connector/connector.html` (byte-gleich in `connector-online/`):

- Favoriten-Kachel ruft `zeigeSpielKarte(pfad)` auf: Kategorie aufklappen, Karte
  in den Blick rollen, kurz umranden (`.u-gewaehlt`), Hauptknopf fokussieren.
  Die Kachel ist damit kein eigener Startweg mehr – alle Modi (am Gerät, gegen
  Computer, online) stehen auf der Karte.
- `?spiel=…` (Deep-Link aus der Spielhalle) macht dasselbe statt `openGame`.
- Jede Spielkarte hat `data-pfad` und eine Sprungmarke (`u-spiel-…`).

`spiele/spiele.html` und `spiele/halle-online.js`:

- Beide Skripte wieder eingebunden; die Kopfzeile zeigt erneut „Online spielen“.
- Klicks auf `a.hl`/`a.karte` werden je Spiel entschieden: Steht das Spiel in
  `multiplayer-register.js`, geht es in den Connector (`?spiel=<kennung>`);
  Einzelspieler-Spiele und Werkzeuge bleiben direkt. In der App (`file://`)
  bleibt alles direkt, weil es dort keinen Connector gibt.
- Die Kachel-Ziele werden zusätzlich umgeschrieben, damit auch „in neuem Tab
  öffnen“ im Raum landet.
- Pfadvergleich normalisiert (`brettspiele/…` ↔ `/spiele/brettspiele/…`); die
  alte Ordnername-Notlösung entfällt, damit kein Einzelspieler-Spiel versehentlich
  im Connector landet.

## Geprüft (Chromium, echter Connector, Port 8775/8776)

- Spielhalle: Favoriten-Kachel 10.000 → `connector.html?spiel=zehntausend`,
  Karte aufgeklappt/hervorgehoben, kein Spiel gestartet, Hinweis „Modus wählen“.
- „Online spielen“ legt den Tisch an (`/api/angebot` enthält `zehntausend`),
  Spiel öffnet mit `?room=…&id=…`.
- Zweiter Browser: Einladungslink `…/connector.html?room=TG-…` → gleicher Raum,
  Tisch mit „Mitspielen“ sichtbar, Beitritt öffnet dieselbe Partie; beide Lobbys
  zeigen beide Spieler, „Bereit“ beider Spieler schaltet „Spiel starten“ frei,
  die Partie beginnt.
- Normale Hallenkarte 10.000 → derselbe Weg (`?spiel=zehntausend`).
- Mahjong (nur Einzelspieler) startet direkt; Würfelbecher (Werkzeug) auch;
  im Connector hat die Mahjong-Karte nur „Einzelspieler“, kein „Online spielen“.
- „Gegen Computer“ (Pig) öffnet die Partie mit Bot am Tisch und legt kein
  Angebot für andere an.
- Handy 390 × 844: Kachel → Risiko-Karte rollt ins Bild (337 px), Modi sichtbar.
- `go test ./...` in `spiele/connector`: grün (197 s).

## Grenzen

- Getestet mit Chromium-Emulation (Desktop und Handy), nicht auf einem echten
  Telefon.
- Der Server führt einen gemeinsamen Spielraum; „Einladungslink“ und
  „?room=“ landen deshalb immer dort. Private Räume entstehen weiterhin nur
  über „Privat einladen“ im Spieler-Menü (unverändert).
- Die Emulation des zweiten Spielers nutzt einen zweiten Browser-Kontext, keine
  zweite Person.

## Sicherung

`connector.html`, `spiele.html` und `halle-online.js` in diesem Ordner sind der
Stand vor der Änderung (Commit `efe47c9`).
