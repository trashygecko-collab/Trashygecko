# -*- coding: utf-8 -*-
"""Ersetzt die Level 86–100 durch 6-Kisten-Level und glättet die Reihenfolge.

Aufruf: python sokoban_letzte_bauen.py
"""

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


def main():
    bau = lade("bau", "sokoban_level_bauen.py")
    sol = lade("sol", "sokoban_solver.py")
    text = io.open(ZIEL, encoding="utf-8").read()
    bloecke = re.findall(r"\{karte:\[(.*?)\],\s*loesung:\s*(.*?),\s*pushes:\s*(.*?)\}",
                         text.replace("\n", ""))
    level = [{"karte": json.loads("[" + k + "]"), "loesung": json.loads(l), "pushes": int(p)}
             for k, l, p in bloecke]
    if len(level) != 100:
        print("FEHLER: %d Level in der Datei" % len(level))
        return 1

    rng = random.Random(777)
    for nr in range(86, 101):
        gebaut = None
        for versuch in range(30):
            stufe = (nr, nr, 6, max(14, 24 - versuch // 3), 36, 12, 10)
            gebaut = bau.erzeuge(stufe, rng, versuche=12, solver_zeit=6.0)
            if gebaut:
                break
        if not gebaut:
            print("FEHLER: Level %d nicht gebaut" % nr)
            return 1
        level[nr - 1] = gebaut
        print("Level %3d  6 Kisten  %3d Schiebe  %s" % (
            nr, gebaut["pushes"], "/".join(gebaut["karte"])), flush=True)

    # Innerhalb jeder Stufe aufsteigend nach Schiebezahl – die Kurve steigt so
    # ohne Sprünge.
    stufen = [(1, 10), (11, 15), (16, 30), (31, 50), (51, 70), (71, 85), (86, 100)]
    for a, b in stufen:
        teil = sorted(level[a - 1:b], key=lambda l: (l["pushes"], len(l["karte"])))
        level[a - 1:b] = teil

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
            json.dumps(l["loesung"]), l["pushes"]))
    zeilen.append("];")
    io.open(ZIEL, "w", encoding="utf-8", newline="\n").write("\n".join(zeilen) + "\n")
    print("\n100 Level geschrieben: %s" % ZIEL)
    return 0


if __name__ == "__main__":
    sys.exit(main())
