# -*- coding: utf-8 -*-
"""Baut level.js aus dem Protokoll des Levelbaus und ergänzt fehlende Level.

Hintergrund: Der Generator braucht für die letzten Stufen lange. Die bereits
gebauten Level stehen im Protokoll (Karten, aber ohne Lösung). Dieses Skript
liest sie, lässt den unabhängigen Solver die Lösung suchen und baut nur die
noch fehlenden Level – mit einem Band, das der Solver sicher schafft.

Aufruf: python sokoban_nachbauen.py /tmp/sokobau.log [--ab 90]
"""

import argparse
import importlib.util
import io
import json
import os
import random
import re
import sys

HIER = os.path.dirname(os.path.abspath(__file__))
ZIEL = os.path.normpath(os.path.join(HIER, "..", "..", "Homepage", "spiele", "sokoban", "level.js"))


def lade(name, datei):
    spec = importlib.util.spec_from_file_location(name, os.path.join(HIER, datei))
    m = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(m)
    return m


def log_lesen(pfad):
    level = {}
    muster = re.compile(r"^Level\s+(\d+)(\s+\(alt\))?\s+.*?\d+ Schiebe\s+(.+)$")
    for zeile in io.open(pfad, encoding="utf-8", errors="replace").read().splitlines():
        m = muster.match(zeile.strip())
        if not m:
            continue
        nr = int(m.group(1))
        karte = m.group(3).strip().split("/")
        if karte and karte[0].startswith("#"):
            level[nr] = karte
    return level


def main():
    p = argparse.ArgumentParser()
    p.add_argument("log")
    p.add_argument("--ab", type=int, default=90)
    p.add_argument("--ziel", default=ZIEL)
    args = p.parse_args()

    bau = lade("bau", "sokoban_level_bauen.py")
    sol = lade("sol", "sokoban_solver.py")

    gebaut = log_lesen(args.log)
    print("Im Protokoll: %d Level (%s)" % (len(gebaut), sorted(gebaut)[:3]))

    level = []
    for nr in range(1, 101):
        if nr <= 10:
            karte = bau.BESTEHEND[nr - 1]
        elif nr in gebaut:
            karte = gebaut[nr]
        else:
            karte = None
        if karte is None:
            continue
        r = sol.loese(karte, zeitlimit=6.0, mit_loesung=True)
        ok, text, pushes = (False, "keine Lösung", 0)
        if r.get("geloest"):
            ok, text, pushes = bau.spiele(karte, r["zuege"])
        if not ok:
            print("FEHLER Level %d: %s" % (nr, text))
            return 1
        level.append({"karte": karte, "loesung": r["zuege"], "pushes": r["pushes"]})
        print("ok  Level %3d  %2d Kisten  %3d Schiebe" % (
            nr, sum(z.count("$") + z.count("*") for z in karte), r["pushes"]))

    # fehlende Level hinten auffüllen
    rng = random.Random(20261126)
    stufe = (args.ab, 100, 5, 14, 24, 12, 9)
    while len(level) < 100:
        nr = len(level) + 1
        for versuch in range(40):
            s = (stufe[0], stufe[1], stufe[2],
                 max(10, stufe[3] - versuch), stufe[4], stufe[5], stufe[6])
            g = bau.erzeuge(s, rng, versuche=40)
            if g:
                break
        else:
            print("FEHLER: Level %d nicht gebaut" % nr)
            return 1
        level.append(g)
        print("neu Level %3d  %2d Kisten  %3d Schiebe  %s" % (
            nr, sum(z.count("$") + z.count("*") for z in g["karte"]),
            g["pushes"], "/".join(g["karte"])))

    zeilen = [
        "/* Sokoban-Level – erzeugt von werkzeuge/sokoban_level_bauen.py.",
        "   Level 1–10 sind die alten Level des Spiels; alle anderen wurden",
        "   gebaut, vorwärts nachgespielt und vom unabhängigen A*-Solver",
        "   (werkzeuge/sokoban_solver.py) gelöst. `pushes` ist die kleinste",
        "   gefundene Schiebezahl, `loesung` die zugehörige Zugfolge. */",
        "window.SOKOBAN_LEVEL = [",
    ]
    for i, l in enumerate(level):
        zeilen.append(" // Level %02d" % (i + 1))
        zeilen.append(" {karte:[%s], loesung:%s, pushes:%s}," % (
            ", ".join(json.dumps(z, ensure_ascii=False) for z in l["karte"]),
            json.dumps(l["loesung"]),
            l["pushes"]))
    zeilen.append("];")
    io.open(args.ziel, "w", encoding="utf-8", newline="\n").write("\n".join(zeilen) + "\n")
    print("\n%d Level geschrieben: %s" % (len(level), args.ziel))
    return 0


if __name__ == "__main__":
    sys.exit(main())
