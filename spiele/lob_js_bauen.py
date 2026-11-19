# -*- coding: utf-8 -*-
"""Baut lob.js aus der aufgeraeumten lob.txt neu.

Das Array LOB_SPRUECHE kommt komplett aus lob.txt.  Die Mahjong-Sprueche
(LOB_MAHJONG) bleiben wie sie sind und werden aus der Hauptliste herausgehalten,
damit sie weiter nur im Mahjong-Spiel auftauchen.  Alles ab dem Mahjong-Absatz
wird unveraendert uebernommen.

Aufruf: python lob_js_bauen.py [--schreiben]
"""
import sys

JS = "lob.js"
TXT = "connector/lob.txt"
MARKER = "// Mahjong-Sprüche"


def liste(zeilen, name):
    """Eintraege des Arrays const <name> = [ ... ]."""
    start = naechste = None
    for i, z in enumerate(zeilen):
        if z.startswith("const %s = [" % name):
            start = i
        elif start is not None and z.strip() == "];":
            naechste = i
            break
    if start is None or naechste is None:
        raise SystemExit("Array %s nicht gefunden" % name)
    aus = []
    for z in zeilen[start + 1:naechste]:
        s = z.strip()
        if not s.startswith('"'):
            continue
        rumpf = s[1:s.rindex('"')]
        aus.append(rumpf.replace('\\"', '"').replace('\\\\', '\\'))
    return aus, start, naechste


js = open(JS, encoding="utf-8").read().splitlines()
txt = [z[1:-1] for z in open(TXT, encoding="utf-8").read().splitlines()]

haupt, h_start, h_ende = liste(js, "LOB_SPRUECHE")
steine, s_start, s_ende = liste(js, "LOB_MAHJONG")

print("lob.js: LOB_SPRUECHE %d, LOB_MAHJONG %d, Zeilen %d"
      % (len(haupt), len(steine), len(js)))
print("lob.txt: %d Spruche, eindeutig %d" % (len(txt), len(set(txt))))
print("Mahjong-Spruche, die schon in LOB_SPRUECHE stehen:", len(set(steine) & set(haupt)))
print("Mahjong-Spruche in lob.txt:", len(set(steine) & set(txt)), "von", len(set(steine)))

neu = [t for t in txt if t not in set(steine)]
print("neue LOB_SPRUECHE:", len(neu), "| davon neu gegenueber jetzt:",
      len(set(neu) - set(haupt)), "| entfallen:", len(set(haupt) - set(neu)))
fehlt = [t for t in haupt if t not in set(txt)]
print("bisherige Spruche, die in lob.txt fehlen:", len(fehlt), fehlt[:3])

if "--schreiben" not in sys.argv:
    print("\n(Probelauf - nichts geschrieben)")
    raise SystemExit

kopf = [z for z in js[:h_start] if z.strip()]
zeilen = ["%s" % z for z in kopf]
zeilen.append("const LOB_SPRUECHE = [")
for t in neu:
    zeilen.append('    "%s",' % t.replace("\\", "\\\\").replace('"', '\\"'))
zeilen.append("];")
zeilen.extend(js[s_start - 1:])          # Leerzeile + Mahjong-Absatz bis Ende

open(JS, "w", encoding="utf-8", newline="\n").write("\n".join(zeilen) + "\n")
print("\ngeschrieben:", JS, len(zeilen), "Zeilen")
