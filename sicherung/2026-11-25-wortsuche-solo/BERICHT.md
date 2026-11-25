# Wortsuche – Einzelspieler wiederhergestellt (25.11.2026)

Datei: `spiele/brettspiele/wortsuche/wortsuche.html`
Sicherung des alten Stands: `sicherung/2026-11-25-wortsuche-solo/wortsuche.html`
(md5 alt `f5c267a1e0891dd131781f08a94f1b0b`, neu `d3c750f78ee47aef9b15ec6836c5f249`)

## Fehlerursache

1. **Gewinnmeldungen im Dauerlauf.**
   Commit `7c49989` hatte in `zeichneAlles()` an die Zeile
   `elEnde.classList.toggle("versteckt", status !== "ende");` ein
   `if (window.TGErgebnis) TGErgebnis.gewonnen();` gehängt. `zeichneAlles()`
   läuft bei **jedem** Neuzeichnen – im Online-Betrieb alle 1,5 s durch die
   Serverabfrage. `ergebnis.js` drosselt sich selbst auf 1,2 s, also erschien
   fortlaufend ein neuer Gewinnspruch, unabhängig vom Spielausgang.

2. **Mehrspieler als Zwang.**
   Der Modus hing allein an `?room=` in der Adresse
   (`const ONLINE = /^TG-[A-F0-9]{4}$/.test(ROOM)`). Der Connector öffnet jedes
   Spiel über `spielAdresse()` **immer** mit `&room=…`; damit landete auch ein
   einzelner Spieler in der Lobby („Warte auf Mitspieler“, „Runde eröffnen“,
   Bereit-Knopf). Die Spielseite kannte keinen Weg zurück in den
   Einzelspieler. `multiplayer-register.js` führt Wortsuche mit `"min": 2` –
   richtig für den Mehrspieler, aber es gab keinen Solo-Einstieg mehr.

## Änderungen (nur `wortsuche.html`)

- `TGErgebnis.gewonnen()` aus `zeichneAlles()` entfernt.
- Ergebnis wird **genau einmal je Runde** gemeldet, getrennte Marken für Solo
  (`gemeldetSolo`) und Online (`gemeldetOnline`):
  - Solo: alle Wörter gefunden → `gewonnen`, Zeit abgelaufen → `verloren`.
  - Online: Status `ende` → Sieger bin ich = `gewonnen`, anderer Sieger =
    `verloren`, unentschieden = `lob`.
- Modusschalter: `ONLINE` ist jetzt umschaltbar. Ohne Raum startet die Seite
  im Einzelspieler; mit Raum (Connector) im Mehrspieler, aber mit Knopf
  **„Allein spielen“** in der Online-Leiste. Im Solo erscheint **„Zu zweit“**
  nur, wenn ein Raum in der Adresse steht. Ein Modus-Schild im Kopf zeigt
  „Einzelspieler“ bzw. „Mehrspieler · TG-XXXX“.
- Beim Umschalten werden Online-Takt (1,5 s), Online-Uhr und
  `visibilitychange`-Lauscher gestoppt (`onlineBeenden()`); im Solo läuft kein
  `fetch`, kein `WG.leben`, keine Serverabfrage.
- „Neue Runde“ startet im Solo sofort ein neues Rätsel (vorher: leeres Brett
  und Wartezustand). „Spiel starten“ und „Nächste Runde“ unverändert.
- Leeres Startgitter nutzt die Gerätegröße (`kanteFuerGeraet()`), nicht mehr
  fest 10×10.
- Keine Änderung an Optik (nur eine kleine Modus-Pille), an
  `multiplayer-register.js`, `ergebnis.js`, `connector.html`, Spielständen
  oder Statistiken. Keine neuen Abhängigkeiten.

## Testergebnisse (Browser, 127.0.0.1:8123, Chromium)

### Einzelspieler ohne Raum (`wortsuche.html`)
| Prüfung | Ergebnis |
| --- | --- |
| Startzustand | Lage „Bereit“, Modus „Einzelspieler“, Startleiste sichtbar, Online-Leiste versteckt, „Zu zweit“ versteckt, 100 Zellen |
| „Spiel starten“ | Runde 1, 8 Wörter, 1 Spieler, „Läuft“, Hinweis „Finde die 8 Wörter …“ |
| Wörter markieren (Klick Start-/Endzelle) | 8 von 8 gefunden, alle in der Liste abgehakt |
| Rätsel gelöst | Lage „Gewonnen“, Notiz „Alle Wörter gefunden – 460 Punkte“, „Nächste Runde“ sichtbar |
| Gewinnmeldung | **1** Aufruf `TGErgebnis.gewonnen()`, Text nach 3,5 s unverändert, kein zweiter Spruch |
| Serverabfragen im Solo | **0** (`/api/…`) |
| „Nächste Runde“ | Runde 2, neues Gitter, 8 neue Wörter, keine Markierungen, Ende-Karte weg, kein weiterer Spruch |
| „Neue Runde“ mitten im Spiel | Runde 3, neues Rätsel, „Läuft“ |
| Zeitablauf (Uhr um 10 min vorgespult) | Lage „Runde vorbei“, Notiz „Die Zeit ist um – 0 von 8 Wörtern gefunden“, **1**× `verloren`, alle Zellen gesperrt |
| Zweiter Durchlauf nach Umbau | 8 von 8, 480 Punkte, **1** Meldung nach 2 s Wartezeit, 0 API-Aufrufe |

### Mehrspieler mit Raum (`?room=TG-ABCD`, Serverantworten simuliert)
| Prüfung | Ergebnis |
| --- | --- |
| Startzustand | Modus „Mehrspieler · TG-ABCD“, Online-Leiste sichtbar, Startleiste versteckt, „Allein spielen“ sichtbar, „Zu zweit“ versteckt |
| Beendete Runde, 3 Serverabfragen | **1** Meldung (vorher: eine je Abfrage) |
| „Allein spielen“ | Modus „Einzelspieler“, eigene Runde 1 mit 8 Wörtern, 1 Spieler, **0** Netzaufrufe in 3,5 s |
| Gewinn im umgeschalteten Solo | 8 von 8, „Gewonnen“, genau 1 Meldung |
| „Zu zweit“ zurück | Modus „Mehrspieler · TG-ABCD“, Abfragen laufen wieder, **0** neue Meldungen (gleiche Runde wird nicht erneut gemeldet) |

## Nicht geprüft / offen

- Echter Mehrspielerbetrieb mit zwei Geräten und dem Go-Server (`/api/wortsuche`)
  lief nicht: Auf 127.0.0.1:8123 gibt es nur den statischen Server, die
  Serverantworten wurden deshalb im Browser nachgestellt. Die Spielregeln des
  Servers wurden nicht angefasst.
- Der Connector wurde nicht geändert; er öffnet Wortsuche weiter mit `room`,
  der Einzelspieler ist jetzt aber von dort mit einem Klick erreichbar.
