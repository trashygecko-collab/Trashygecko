# Eigene Smilies und GIFs

Alles, was hier liegt, erscheint im Connector-Chat hinter dem Smiley-Knopf 🙂
neben dem Senden-Knopf. Eine Liste muss nirgends gepflegt werden: Der Server
liest diesen Ordner bei jeder Abfrage neu – Datei hineinlegen, Seite neu laden,
fertig.

## Dateien ablegen

```
spiele/connector/smilies/
├── katze.gif
├── daumen-hoch.png
├── tiere/            ← Unterordner = Kategorie
│   ├── hund.webp
│   └── eule.jpg
└── party/
    └── konfetti.gif
```

- Erlaubt: `.png`, `.jpg`, `.jpeg`, `.webp`, `.gif` (GIFs bleiben animiert).
- Höchstens **1 MiB** pro Datei – größere Dateien werden ignoriert.
- Höchstens **500** Grafiken insgesamt, höchstens ein Unterordner tief.
- Unterordner sind Kategorien; sie erscheinen als Reiter im Auswahlpanel.
  Dateien direkt in `smilies/` laufen unter „Ohne Ordner“.
- Empfehlung: quadratisch, etwa 64–128 px. Kleinere Dateien laden schneller.

## Wie es im Chat funktioniert

1. Smiley-Knopf 🙂 drücken – das Panel öffnet sich über der Eingabezeile.
2. Suchen (Feld oben) oder Kategorie wählen, dann eine Grafik anklicken.
3. Die Grafik erscheint als kleines Vorschaukästchen über der Eingabe; mit ×
   wieder entfernen. Text dazu ist möglich, z. B. „schau mal 😀“ + GIF.
4. „Senden“ schickt beides zusammen. Alle im Raum sehen dieselbe Grafik, GIFs
   animiert, auch nach dem Neuladen der Seite.

## Grenzen

Nur Dateien aus diesem Ordner dürfen verschickt werden; der Server prüft jeden
Bildnamen gegen den Bestand (Endung, Pfad, Größe). Fremde Adressen, HTML oder
ausführbare Dateien werden abgewiesen. Neue Grafiken brauchen deshalb keinen
Neustart, aber ein Neuladen der Chatseite im Browser.
