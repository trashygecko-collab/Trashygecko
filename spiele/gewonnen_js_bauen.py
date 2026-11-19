# -*- coding: utf-8 -*-
"""Baut das Array GEWONNEN_SPRUECHE in mahjong/js/gewonnen.js aus gewonnen.txt.

Die Sprueche der Sieg-Tafel stehen als Textdatei in connector/gewonnen.txt
(je Zeile ein Spruch in Anfuehrungszeichen).  Dieses Skript uebernimmt sie in
der Reihenfolge der Datei; alles nach dem Array (Auswahl, Symbole, Anzeige)
bleibt unveraendert.  Die Reihenfolge ist egal - die Sprueche werden im Spiel
gemischt.

Aufruf: python gewonnen_js_bauen.py [--schreiben]
"""
import sys

JS = "mahjong/js/gewonnen.js"
TXT = "connector/gewonnen.txt"
NAME = "GEWONNEN_SPRUECHE"


def liste(zeilen, name):
    """Eintraege des Arrays const <name> = [ ... ] samt Zeilennummern."""
    start = ende = None
    for i, z in enumerate(zeilen):
        if z.startswith("const %s = [" % name):
            start = i
        elif start is not None and z.strip() == "];":
            ende = i
            break
    if start is None or ende is None:
        raise SystemExit("Array %s nicht gefunden" % name)
    aus = []
    for z in zeilen[start + 1:ende]:
        s = z.strip()
        if not s.startswith('"'):
            continue
        rumpf = s[1:s.rindex('"')]
        aus.append(rumpf.replace('\\"', '"').replace("\\\\", "\\"))
    return aus, start, ende


def txt_liste():
    aus = []
    for z in open(TXT, encoding="utf-8").read().splitlines():
        s = z.strip()
        if not s:
            continue
        if s.startswith('"') and s.endswith('"') and len(s) > 1:
            s = s[1:-1]
        aus.append(s)
    return aus


js = open(JS, encoding="utf-8").read().splitlines()
txt = txt_liste()
alt, a_start, a_ende = liste(js, NAME)

print("gewonnen.txt: %d Zeilen, eindeutig %d" % (len(txt), len(set(txt))))
print("%s jetzt: %d" % (NAME, len(alt)))
print("neu dazu:", len(set(txt) - set(alt)), "| entfaellt:", len(set(alt) - set(txt)))
for t in list(set(alt) - set(txt))[:5]:
    print("   -", t[:90])

if "--schreiben" not in sys.argv:
    print("\n(Probelauf - nichts geschrieben)")
    raise SystemExit

zeilen = js[:a_start]
zeilen.append("const %s = [" % NAME)
for t in txt:
    zeilen.append('    "%s",' % t.replace("\\", "\\\\").replace('"', '\\"'))
zeilen.append("];")
zeilen.extend(js[a_ende + 1:])

open(JS, "w", encoding="utf-8", newline="\n").write("\n".join(zeilen) + "\n")
print("\ngeschrieben:", JS, len(zeilen), "Zeilen")
