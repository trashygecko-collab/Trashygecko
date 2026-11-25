# -*- coding: utf-8 -*-
"""Unabhängiger Sokoban-Solver (A*) für spiele/sokoban/level.js.

Aufruf:
    python sokoban_solver.py              # alle Level, je 10 s Zeitlimit
    python sokoban_solver.py --nr 37      # nur Level 37 (1-basiert)
    python sokoban_solver.py --zeit 30

Der Solver kennt keine Lösung aus der Datei. Er sucht selbst: Zustand ist die
Kistenmenge plus die erreichbare Spielerregion (Spielerpositionen innerhalb
einer Region sind gleichwertig). Kosten sind Schiebezüge, darum ist die
gefundene Lösung push-optimal; die Zugzahl ergibt sich aus den Laufwegen.

Erkannt werden die beiden einfachen Endstellungen: Kiste in einer Ecke und
2×2-Blöcke aus Kisten und Wänden (jeweils ohne Ziel).
"""

import argparse
import heapq
import io
import json
import os
import re
import sys
import time
from collections import deque

HIER = os.path.dirname(os.path.abspath(__file__))
LEVEL_DATEI = os.path.normpath(os.path.join(HIER, "..", "..", "Homepage", "spiele", "sokoban", "level.js"))

RICH = {"hoch": (0, -1), "runter": (0, 1), "links": (-1, 0), "rechts": (1, 0)}


def level_lesen(datei):
    text = io.open(datei, encoding="utf-8").read()
    # Nur Zeilenumbrüche entfernen – Leerzeichen sind Teil der Karten!
    bloecke = re.findall(
        r"\{karte:\[(.*?)\],\s*loesung:\s*(.*?),\s*pushes:\s*(.*?)\}",
        text.replace("\n", ""))
    level = []
    for karte_roh, _, _ in bloecke:
        level.append(json.loads("[" + karte_roh + "]"))
    return level


class Brett:
    def __init__(self, karte):
        self.breite = max(len(z) for z in karte)
        self.hoehe = len(karte)
        self.waende = set()
        self.ziele = frozenset()
        kisten = set()
        spieler = None
        for y, zeile in enumerate(karte):
            for x, c in enumerate(zeile):
                if c == "#":
                    self.waende.add((x, y))
                if c in ".+*":
                    self.ziele |= {(x, y)}
                if c in "$*":
                    kisten.add((x, y))
                if c in "@+":
                    spieler = (x, y)
        self.kisten = frozenset(kisten)
        self.spieler = spieler
        self.frei = {(x, y) for y in range(self.hoehe) for x in range(self.breite)} - self.waende
        self.nachbar = {r: (dx, dy) for r, (dx, dy) in RICH.items()}

    def im_brett(self, z):
        return 0 <= z[0] < self.breite and 0 <= z[1] < self.hoehe

    def region(self, spieler, kisten):
        """Erreichbare Zellen ohne Kistenschub."""
        if spieler in self.waende or spieler in kisten:
            return frozenset()
        gesehen = {spieler}
        schlange = deque([spieler])
        while schlange:
            x, y = schlange.popleft()
            for dx, dy in RICH.values():
                n = (x + dx, y + dy)
                if n in gesehen or n in self.waende or n in kisten:
                    continue
                gesehen.add(n)
                schlange.append(n)
        return frozenset(gesehen)

    def distanzen(self):
        """Für jede Zelle: kleinste Schiebezahl bis zu irgendeinem Ziel."""
        dist = {z: 0 for z in self.ziele}
        schlange = deque(self.ziele)
        while schlange:
            x, y = schlange.popleft()
            d = dist[(x, y)]
            # Eine Kiste kommt von (x-dx, y-dy) und wurde nach (x,y) geschoben:
            # der Spieler stand dahinter, also muss (x-2dx, y-2dy) frei sein.
            for dx, dy in RICH.values():
                vor = (x - dx, y - dy)
                dahinter = (x - 2 * dx, y - 2 * dy)
                if vor in self.waende or dahinter in self.waende:
                    continue
                if not self.im_brett(vor) or not self.im_brett(dahinter):
                    continue
                if vor not in dist:
                    dist[vor] = d + 1
                    schlange.append(vor)
        return dist

    def heuristik(self, kisten, dist):
        summe = 0
        for k in kisten:
            if k in self.ziele:
                continue
            d = dist.get(k)
            if d is None:                 # Kiste kann nie mehr auf ein Ziel
                return None
            summe += d
        return summe

    def tote_stellung(self, kisten):
        for (x, y) in kisten:
            if (x, y) in self.ziele:
                continue
            w = [((x + dx, y + dy) in self.waende) for dx, dy in RICH.values()]
            if (w[0] and w[2]) or (w[0] and w[3]) or (w[1] and w[2]) or (w[1] and w[3]):
                return True
        # 2×2-Block aus Kisten und Wänden: einmal fest, immer fest
        for (x, y) in kisten:
            for ex, ey in ((0, 0), (-1, 0), (0, -1), (-1, -1)):
                zellen = [(x + ex, y + ey), (x + ex + 1, y + ey),
                          (x + ex, y + ey + 1), (x + ex + 1, y + ey + 1)]
                if all((z in self.waende or z in kisten) for z in zellen):
                    if not any(z in self.ziele for z in zellen):
                        return True
        return False


def weg_b(b, start, ziel, kisten):
    """Kürzester Laufweg um die Kisten herum (Liste von Richtungen)."""
    if start == ziel:
        return []
    kam = {start: None}
    schlange = deque([start])
    while schlange:
        x, y = schlange.popleft()
        for r, (dx, dy) in RICH.items():
            n = (x + dx, y + dy)
            if n in kam or n in b.waende or n in kisten:
                continue
            kam[n] = (x, y)
            if n == ziel:
                schlange.clear()
                break
            schlange.append(n)
    if ziel not in kam:
        return None
    zuege = []
    z = ziel
    while kam[z] is not None:
        px, py = kam[z]
        for r, (dx, dy) in RICH.items():
            if (px + dx, py + dy) == z:
                zuege.append(r)
                break
        z = (px, py)
    return list(reversed(zuege))


def pfad_zuege(b, kisten0, spieler0, eltern, ziel_schluessel):
    """Baut aus der A*-Suche eine vollständige Zugfolge (Laufen und Schieben)."""
    schritte = []
    z = ziel_schluessel
    while z in eltern:
        vorher, kiste, r = eltern[z]
        schritte.append((kiste, r))
        z = vorher
    schritte.reverse()
    zuege = []
    kisten = set(kisten0)
    pos = spieler0
    for (kiste, r) in schritte:
        dx, dy = RICH[r]
        dahinter = (kiste[0] - dx, kiste[1] - dy)
        weg = weg_b(b, pos, dahinter, kisten)
        if weg is None:
            return None
        zuege.extend(weg)
        zuege.append(r)
        kisten.discard(kiste)
        kisten.add((kiste[0] + dx, kiste[1] + dy))
        pos = kiste
    return zuege


def loese(karte, zeitlimit=10.0, max_knoten=400000, mit_loesung=False):
    """A* über Schiebezüge. Liefert dict mit geloest, pushes, zuege, knoten, zeit."""
    start = time.time()
    b = Brett(karte)
    dist = b.distanzen()
    start_region = b.region(b.spieler, b.kisten)
    h0 = b.heuristik(b.kisten, dist)
    if h0 is None:
        return {"geloest": False, "grund": "Kiste kann kein Ziel erreichen", "knoten": 0,
                "zeit": time.time() - start}

    # Zustand: (Kisten, kleinste Zelle der Region) – Regionen sind eindeutig
    start_zustand = (b.kisten, min(start_region))
    offen = [(h0, 0, start_zustand, b.spieler)]
    beste = {start_zustand: 0}
    eltern = {}
    knoten = 0
    while offen:
        if time.time() - start > zeitlimit or knoten > max_knoten:
            break
        f, g, (kisten, _), spieler = heapq.heappop(offen)
        if g > beste.get((kisten, min(b.region(spieler, kisten))), 10**9):
            continue
        knoten += 1
        if kisten == b.ziele:
            ergebnis = {"geloest": True, "pushes": g, "knoten": knoten,
                        "zeit": time.time() - start}
            if mit_loesung:
                ergebnis["zuege"] = pfad_zuege(b, b.kisten, b.spieler, eltern,
                                               (kisten, min(b.region(spieler, kisten))))
            return ergebnis
        region = b.region(spieler, kisten)
        for k in kisten:
            for r, (dx, dy) in RICH.items():
                hinter = (k[0] - dx, k[1] - dy)
                vor = (k[0] + dx, k[1] + dy)
                if hinter not in region:
                    continue
                if vor in b.waende or vor in kisten:
                    continue
                neue = set(kisten)
                neue.discard(k)
                neue.add(vor)
                neue = frozenset(neue)
                if b.tote_stellung(neue):
                    continue
                h = b.heuristik(neue, dist)
                if h is None:
                    continue
                ng = g + 1
                # Spieler steht nach dem Schub auf dem alten Kistenfeld
                reg = b.region(k, neue)
                if not reg:
                    continue
                schluessel = (neue, min(reg))
                if ng >= beste.get(schluessel, 10**9):
                    continue
                beste[schluessel] = ng
                markup = (kisten, min(b.region(spieler, kisten)))
                eltern[schluessel] = (markup, k, r)
                heapq.heappush(offen, (ng + h, ng, (neue, min(reg)), k))
    return {"geloest": False, "grund": "Zeit- oder Knotenlimit", "knoten": knoten,
            "zeit": time.time() - start}


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--datei", default=LEVEL_DATEI)
    p.add_argument("--nr", type=int, default=0)
    p.add_argument("--zeit", type=float, default=10.0)
    p.add_argument("--json", default="")
    args = p.parse_args()

    level = level_lesen(args.datei)
    print("Level in Datei: %d" % len(level))
    auswahl = range(len(level)) if not args.nr else [args.nr - 1]
    ergebnisse = []
    schlecht = 0
    for i in auswahl:
        r = loese(level[i], zeitlimit=args.zeit)
        r["nr"] = i + 1
        ergebnisse.append(r)
        kisten = sum(z.count("$") + z.count("*") for z in level[i])
        if r["geloest"]:
            print("ok      Level %3d  %d Kisten  %3d Schiebe  %7d Knoten  %5.1f s" % (
                i + 1, kisten, r["pushes"], r["knoten"], r["zeit"]))
        else:
            schlecht += 1
            print("OFFEN   Level %3d  %d Kisten  %s  %7d Knoten  %5.1f s" % (
                i + 1, kisten, r["grund"], r["knoten"], r["zeit"]))
    print("\n%d Level, %d offen" % (len(ergebnisse), schlecht))
    if args.json:
        io.open(args.json, "w", encoding="utf-8").write(json.dumps(ergebnisse, indent=1))
    return 0 if schlecht == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
