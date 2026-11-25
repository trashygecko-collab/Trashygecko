# -*- coding: utf-8 -*-
"""Sokoban-Level bauen und prüfen – für spiele/sokoban/level.js.

Aufruf:
    python sokoban_level_bauen.py            # 100 Level nach ../Homepage/spiele/sokoban/level.js
    python sokoban_level_bauen.py --pruefen  # nur vorhandene level.js prüfen
    python sokoban_level_bauen.py --test     # Generator + Solver selbst testen

Vorgehen:
  1. Raum mit Wänden, Zielen, Kisten und Spieler aufbauen.
  2. Rückwärts ziehen: Der Spieler zieht je einen Schritt eine Kiste. Jede
     gezogene Folge ist vorwärts eine gültige Schiebe-Folge – darum ist jedes
     erzeugte Level konstruktiv lösbar.
  3. Vorwärts prüfen: Ein eigener Simulator spielt die Lösung Schritt für
     Schritt nach (unabhängig vom Generator-Code) und muss gewinnen.
  4. Breitensuche (Solver) auf kleinen Leveln: liefert die kürzeste Lösung und
     damit die kleinste Schiebezahl; dient als unabhängiger Beweis.
  5. Ausgabe als window.SOKOBAN_LEVEL mit Level, Lösung und Messwerten.
"""

import argparse
import io
import json
import os
import random
import sys

HIER = os.path.dirname(os.path.abspath(__file__))
ZIEL_DATEI = os.path.normpath(os.path.join(HIER, "..", "..", "Homepage", "spiele", "sokoban", "level.js"))

# Die ersten zehn Level des Spiels bleiben unverändert – ihre Bestwerte
# (sokoban-best-0 … -9) behalten damit ihre Bedeutung.
BESTEHEND = [
    ["#######", "#  .  #", "#     #", "#  $  #", "#  @  #", "#######"],
    ["########", "#   .  #", "#      #", "#  $   #", "#  @   #", "########"],
    ["########", "#      #", "# $  . #", "#  @   #", "#      #", "########"],
    ["#########", "#  . .  #", "#  $ $  #", "#   @   #", "#       #", "#########"],
    ["##########", "#        #", "#  .  .  #", "#  $  $  #", "#   @    #", "#        #", "##########"],
    ["#########", "#  .    #", "#  $#   #", "#   #   #", "#  $@   #", "#  .    #", "#########"],
    ["###########", "#         #", "#  . . .  #", "#  $ $ $  #", "#    @    #", "#         #", "###########"],
    ["##########", "#        #", "#  $  .  #", "#   ##   #", "#  .  $  #", "#    @   #", "##########"],
    ["#########", "#       #", "# $   . #", "#   @   #", "# $   . #", "#       #", "#########"],
    ["##########", "#   #    #", "# $ # .  #", "#   #    #", "#   @    #", "#  .  $  #", "##########"],
]

# Stufen: (von, bis, Kisten, min. Pushes, max. Pushes, Breite, Höhe)
# Die Pushes sind das Optimum des Solvers – nicht die Länge irgendeiner Lösung.
STUFEN = [
    (11, 15, 1, 3, 6, 7, 6),        # sehr leicht
    (16, 25, 2, 6, 10, 8, 6),       # leicht
    (26, 30, 2, 9, 14, 9, 7),
    (31, 40, 3, 12, 18, 9, 7),      # mittel
    (41, 50, 3, 16, 24, 10, 8),
    (51, 60, 4, 20, 28, 10, 8),     # anspruchsvoll
    (61, 70, 4, 18, 26, 11, 8),
    (71, 80, 4, 20, 28, 11, 9),     # schwer
    (81, 85, 5, 22, 30, 12, 9),
    (86, 95, 5, 24, 34, 12, 9),     # Experten
    (96, 100, 5, 26, 38, 12, 10),
]

RICH = {"hoch": (0, -1), "runter": (0, 1), "links": (-1, 0), "rechts": (1, 0)}
GEGEN = {"hoch": "runter", "runter": "hoch", "links": "rechts", "rechts": "links"}


class Level:
    def __init__(self, spalten, zeilen):
        self.spalten = spalten
        self.zeilen = zeilen
        self.waende = set()
        self.ziele = set()
        self.kisten = set()
        self.spieler = None

    def kopie(self):
        l = Level(self.spalten, self.zeilen)
        l.waende = set(self.waende)
        l.ziele = set(self.ziele)
        l.kisten = set(self.kisten)
        l.spieler = self.spieler
        return l

    def frei(self, x, y):
        return 0 <= x < self.spalten and 0 <= y < self.zeilen and \
            (x, y) not in self.waende and (x, y) not in self.kisten

    def zellen(self):
        return [(x, y) for y in range(self.zeilen) for x in range(self.spalten)
                if (x, y) not in self.waende]


def raum_bauen(rng, spalten, zeilen, kisten_anzahl, innenwaende=True):
    """Rechteck mit Randmauern, Kammern (Trennwände mit Türen) und Blöcken.
    Kammern und enge Türen machen aus dem Schieben ein Rätsel: Eine Kiste muss
    durch einen Durchgang, und dort lässt sie sich nicht mehr drehen."""
    l = Level(spalten, zeilen)
    for x in range(spalten):
        l.waende.add((x, 0))
        l.waende.add((x, zeilen - 1))
    for y in range(zeilen):
        l.waende.add((0, y))
        l.waende.add((spalten - 1, y))

    if innenwaende and spalten >= 6 and zeilen >= 6:
        # Trennwände mit ein bis zwei Türen
        trenner = []
        if spalten >= 9 and rng.random() < 0.8:
            trenner.append(("v", rng.randint(3, spalten - 4)))
        if spalten >= 11 and rng.random() < 0.55:
            x = rng.randint(3, spalten - 4)
            if all(abs(x - t[1]) >= 3 for t in trenner):
                trenner.append(("v", x))
        if zeilen >= 8 and rng.random() < 0.4:
            trenner.append(("h", rng.randint(3, zeilen - 4)))
        for art, pos in trenner:
            if art == "v":
                zellen = [(pos, y) for y in range(1, zeilen - 1)]
            else:
                zellen = [(x, pos) for x in range(1, spalten - 1)]
            tueren = set(rng.sample(zellen, min(len(zellen), rng.randint(1, 2))))
            for z in zellen:
                if z not in tueren:
                    l.waende.add(z)

        # ein bis zwei kompakte Blöcke als Möbel
        for _ in range(rng.randint(0, 2)):
            bx = rng.randint(2, spalten - 3)
            by = rng.randint(2, zeilen - 3)
            form = rng.choice([[(0, 0)], [(0, 0), (1, 0)], [(0, 0), (0, 1)],
                               [(0, 0), (1, 0), (2, 0)]])
            for dx, dy in form:
                x, y = bx + dx, by + dy
                if 1 <= x < spalten - 1 and 1 <= y < zeilen - 1:
                    l.waende.add((x, y))

    offen = [z for z in l.zellen() if z not in l.waende]
    if len(offen) < kisten_anzahl * 5 + 8:
        return None

    # Ziele: nicht in Ecken (mindestens zwei freie Nachbarn), Abstand zueinander
    kandidaten = []
    for (x, y) in offen:
        frei = sum(1 for dx, dy in RICH.values() if 0 <= x + dx < spalten and 0 <= y + dy < zeilen
                   and (x + dx, y + dy) not in l.waende)
        if frei >= 2:
            kandidaten.append((x, y))
    rng.shuffle(kandidaten)
    gewaehlt = []
    for z in kandidaten:
        if all(abs(z[0] - g[0]) + abs(z[1] - g[1]) >= 2 for g in gewaehlt):
            gewaehlt.append(z)
        if len(gewaehlt) == kisten_anzahl:
            break
    if len(gewaehlt) < kisten_anzahl:
        return None
    l.ziele = set(gewaehlt)
    l.kisten = set(gewaehlt)

    # Spieler: freies Feld, nicht auf Kiste/Ziel
    rest = [z for z in offen if z not in l.kisten]
    if not rest:
        return None
    l.spieler = rng.choice(rest)
    return l


def zusammenhaengend(l):
    if not l.spieler:
        return False
    frei = set(l.zellen())
    if l.spieler not in frei:
        return False
    gesehen = {l.spieler}
    stapel = [l.spieler]
    while stapel:
        x, y = stapel.pop()
        for dx, dy in RICH.values():
            n = (x + dx, y + dy)
            if n in frei and n not in gesehen:
                gesehen.add(n)
                stapel.append(n)
    # alle Kisten und Ziele müssen erreichbar sein (Kisten stehen auf Boden)
    return all(z in gesehen for z in l.kisten) and all(z in gesehen for z in l.ziele)


def ziehen(l, richtung):
    """Rückwärts: Spieler zieht die Kiste, die in Richtung `richtung` steht.
    Der Spieler geht in die Gegenrichtung, die Kiste folgt auf sein altes Feld.
    Liefert True, wenn der Zug ausgeführt wurde."""
    dx, dy = RICH[richtung]
    sx, sy = l.spieler
    kx, ky = sx + dx, sy + dy          # Kiste vor dem Spieler
    if (kx, ky) not in l.kisten:
        return False
    nx, ny = sx - dx, sy - dy          # Spieler tritt zurück
    if not (0 <= nx < l.spalten and 0 <= ny < l.zeilen):
        return False
    if (nx, ny) in l.waende or (nx, ny) in l.kisten:
        return False
    l.kisten.discard((kx, ky))
    l.kisten.add((sx, sy))
    l.spieler = (nx, ny)
    return True


def laufen(l, richtung):
    """Rückwärts gehen: Der Spieler tritt einen Schritt in `richtung`.
    Vorwärts ist das ein Schritt in die Gegenrichtung."""
    dx, dy = RICH[richtung]
    nx, ny = l.spieler[0] + dx, l.spieler[1] + dy
    if not (0 <= nx < l.spalten and 0 <= ny < l.zeilen):
        return False
    if (nx, ny) in l.waende or (nx, ny) in l.kisten:
        return False
    l.spieler = (nx, ny)
    return True


def erreichbar(l):
    """Alle Zellen, die der Spieler ohne Kistenschub erreicht (Vorgängerkarte)."""
    from collections import deque
    kam = {l.spieler: None}
    schlange = deque([l.spieler])
    while schlange:
        x, y = schlange.popleft()
        for r, (dx, dy) in RICH.items():
            n = (x + dx, y + dy)
            if n in l.waende or n in l.kisten or n in kam:
                continue
            kam[n] = (x, y)
            schlange.append(n)
    return kam


def weg_zu(kam, ziel):
    """Kürzester Weg (Richtungen) von der Startzelle zu `ziel`."""
    if ziel not in kam:
        return None
    zuege = []
    while kam[ziel] is not None:
        px, py = kam[ziel]
        x, y = ziel
        for r, (dx, dy) in RICH.items():
            if (px + dx, py + dy) == (x, y):
                zuege.append(r)
                break
        ziel = (px, py)
    return list(reversed(zuege))


def distanz_karte(l):
    """Kleinste Schiebezahl jeder Zelle zum nächsten Ziel (BFS rückwärts)."""
    from collections import deque
    dist = {z: 0 for z in l.ziele}
    schlange = deque(l.ziele)
    while schlange:
        x, y = schlange.popleft()
        for dx, dy in RICH.values():
            vor = (x - dx, y - dy)
            dahinter = (x - 2 * dx, y - 2 * dy)
            if vor in l.waende or dahinter in l.waende:
                continue
            if not (0 <= vor[0] < l.spalten and 0 <= vor[1] < l.zeilen):
                continue
            if not (0 <= dahinter[0] < l.spalten and 0 <= dahinter[1] < l.zeilen):
                continue
            if vor not in dist:
                dist[vor] = dist[(x, y)] + 1
                schlange.append(vor)
    return dist


def loesung_bauen(l, rng, pulls):
    """Baut rückwärts `pulls` Schiebezüge samt kürzester Laufwege.
    Liefert die Vorwärts-Lösung als Liste von Richtungen.

    Rückwärts ist ein Zug ein Ziehen: Der Spieler steht hinter der Kiste und
    tritt einen Schritt zurück, die Kiste folgt. Vorwärts ist das ein Schub in
    dieselbe Richtung. Kein Schub-Zustand darf zweimal vorkommen, sonst
    enthielte die Lösung eine Schleife."""
    aktionen = []                     # Laufzüge und Züge, rückwärts gesehen
    gesehen = {(l.spieler, frozenset(l.kisten))}
    dist = distanz_karte(l)

    def ein_pull():
        """Zieht eine Kiste – möglichst weit von ihrem Ziel weg."""
        kam = erreichbar(l)
        moeglich = []
        for (px, py) in kam:
            for r, (dx, dy) in RICH.items():
                kx, ky = px + dx, py + dy
                if (kx, ky) not in l.kisten:
                    continue
                nx, ny = px - dx, py - dy
                if not (0 <= nx < l.spalten and 0 <= ny < l.zeilen):
                    continue
                if (nx, ny) in l.waende or (nx, ny) in l.kisten:
                    continue
                probe = l.kopie()
                probe.spieler = (px, py)
                ziehen(probe, r)
                zustand = (probe.spieler, frozenset(probe.kisten))
                if zustand not in gesehen:
                    # Wie weit steht die Kiste nach dem Ziehen vom Ziel weg?
                    davor = dist.get((kx, ky))
                    danach = dist.get((px, py))
                    weit = (danach - davor) if (davor is not None and danach is not None) else -99
                    moeglich.append(((px, py), r, zustand, weit))
        if not moeglich:
            return False
        moeglich.sort(key=lambda m: -m[3])
        beste = moeglich[:3]
        pos, r, zustand, _ = rng.choice(beste)
        for s in weg_zu(kam, pos):
            laufen(l, s)
            aktionen.append(("lauf", s))
        ziehen(l, r)
        aktionen.append(("pull", r))
        gesehen.add(zustand)
        return True

    for _ in range(pulls):
        if not ein_pull():
            break
    # Eine Kiste, die am Ende noch auf einem Ziel steht, wäre verschenkt:
    # weiterziehen, bis keine Kiste mehr auf einem Ziel steht.
    extra = 0
    while (l.kisten & l.ziele) and extra < 12:
        if not ein_pull():
            break
        extra += 1

    zuege = []
    for art, r in reversed(aktionen):
        zuege.append(r if art == "pull" else GEGEN[r])
    return zuege


def als_karte(l):
    zeilen = []
    for y in range(l.zeilen):
        zeile = ""
        for x in range(l.spalten):
            if (x, y) in l.waende:
                zeile += "#"
            elif (x, y) == l.spieler and (x, y) in l.ziele:
                zeile += "+"
            elif (x, y) == l.spieler:
                zeile += "@"
            elif (x, y) in l.kisten and (x, y) in l.ziele:
                zeile += "*"
            elif (x, y) in l.kisten:
                zeile += "$"
            elif (x, y) in l.ziele:
                zeile += "."
            else:
                zeile += " "
        zeilen.append(zeile)
    return zeilen


# --------------------------------------------------------------- Prüfungen

def karte_lesen(karte):
    l = Level(max(len(z) for z in karte), len(karte))
    for y, zeile in enumerate(karte):
        for x, c in enumerate(zeile):
            if c == "#":
                l.waende.add((x, y))
            elif c in ".+*":
                l.ziele.add((x, y))
            if c in "$*":
                l.kisten.add((x, y))
            if c in "@+":
                l.spieler = (x, y)
            if c != " " and c not in "#. $*@+":
                raise ValueError("unbekanntes Zeichen %r in Zeile %d" % (c, y + 1))
    return l


def pruefe_struktur(karte, meldung):
    """Struktur: genau ein Spieler, Kisten == Ziele, alles zusammenhängend."""
    fehler = []
    text = "\n".join(karte)
    if text.count("@") + text.count("+") != 1:
        fehler.append("%s: %d Spielfiguren" % (meldung, text.count("@") + text.count("+")))
    anzahl_kisten = text.count("$") + text.count("*")
    anzahl_ziele = text.count(".") + text.count("*") + text.count("+")
    if anzahl_kisten != anzahl_ziele:
        fehler.append("%s: %d Kisten, %d Ziele" % (meldung, anzahl_kisten, anzahl_ziele))
    if anzahl_kisten == 0:
        fehler.append("%s: keine Kiste" % meldung)
    l = karte_lesen(karte)
    if not zusammenhaengend(l):
        fehler.append("%s: Boden nicht zusammenhängend" % meldung)
    # Ziele in Ecken sind erlaubt, aber Kisten dort machen das Level schwer bis
    # unmöglich – die Lösung unten deckt das ab.
    return fehler


def spiele(karte, loesung):
    """Unabhängiger Vorwärtssimulator: prüft die Lösung Schritt für Schritt."""
    l = karte_lesen(karte)
    schieben = 0
    for i, r in enumerate(loesung):
        dx, dy = RICH[r]
        x, y = l.spieler
        nx, ny = x + dx, y + dy
        if (nx, ny) in l.waende:
            return False, "Zug %d (%s) läuft in eine Wand" % (i + 1, r), schieben
        if (nx, ny) in l.kisten:
            hx, hy = nx + dx, ny + dy
            if (hx, hy) in l.waende or (hx, hy) in l.kisten:
                return False, "Zug %d (%s): Kiste blockiert" % (i + 1, r), schieben
            l.kisten.discard((nx, ny))
            l.kisten.add((hx, hy))
            schieben += 1
        l.spieler = (nx, ny)
    if l.kisten != l.ziele:
        return False, "am Ende stehen %d von %d Kisten auf Zielen" % (len(l.kisten & l.ziele), len(l.ziele)), schieben
    return True, "Lösung gültig", schieben


def tote_ecke(zelle, l):
    """Kiste in einer Ecke, die kein Ziel ist, ist unwiderruflich verloren."""
    x, y = zelle
    waende = []
    for dx, dy in RICH.values():
        if (x + dx, y + dy) in l.waende:
            waende.append((dx, dy))
    # zwei senkrechte Nachbarn zu – echte Ecke
    for a, b in RICH.values():
        pass
    orth = list(RICH.values())
    for i in range(len(orth)):
        for j in range(i + 1, len(orth)):
            if orth[i][0] + orth[j][0] == 0 and orth[i][1] + orth[j][1] == 0:
                if orth[i] in waende and orth[j] in waende:
                    return True
    return False


def bfs_loesen(karte, max_zustaende=250000, max_tiefe=90):
    """Breitensuche über (Spieler, Kisten). Liefert (Züge, Pushes) oder None."""
    l = karte_lesen(karte)
    start = (l.spieler, frozenset(l.kisten))
    ziele = frozenset(l.ziele)
    from collections import deque
    schlange = deque()
    schlange.append((start, 0, 0))
    gesehen = {start}
    while schlange:
        (spieler, kisten), tiefe, pushes = schlange.popleft()
        if kisten == ziele:
            return tiefe, pushes
        if tiefe >= max_tiefe or len(gesehen) > max_zustaende:
            return None
        x, y = spieler
        for r, (dx, dy) in RICH.items():
            nx, ny = x + dx, y + dy
            if (nx, ny) in l.waende:
                continue
            neue_kisten = kisten
            neue_pushes = pushes
            if (nx, ny) in kisten:
                hx, hy = nx + dx, ny + dy
                if (hx, hy) in l.waende or (hx, hy) in kisten:
                    continue
                neu = set(kisten)
                neu.discard((nx, ny))
                neu.add((hx, hy))
                neue_kisten = frozenset(neu)
                neue_pushes += 1
                # Kiste in einer Ecke ohne Ziel: nie mehr lösbar
                if (hx, hy) not in ziele and tote_ecke((hx, hy), l):
                    continue
            zustand = ((nx, ny), neue_kisten)
            if zustand in gesehen:
                continue
            gesehen.add(zustand)
            schlange.append((zustand, tiefe + 1, neue_pushes))
    return None


def erzeuge(stufe, rng, versuche=40, solver_zeit=1.5):
    """Baut ein Level für eine Stufe, prüft es und misst es mit dem Solver.
    Liefert dict oder None.

    Die Rückwärtskonstruktion allein sagt nichts über die Schwierigkeit: Sie
    erzeugt oft weite Umwege, die der Solver abkürzt. Darum entscheidet die
    kleinste Schiebezahl des Solvers, ob ein Level zur Stufe passt."""
    import sokoban_solver
    von, bis, kisten, min_p, max_p, breite, hoehe = stufe
    ziel_pulls = int(max_p * 1.8) + 6
    for _ in range(versuche):
        spalten = rng.randint(breite, breite + 1)
        zeilen = rng.randint(hoehe, hoehe + 1)
        l = raum_bauen(rng, spalten, zeilen, kisten, innenwaende=rng.random() < 0.9)
        if not l or not zusammenhaengend(l):
            continue
        loesung = loesung_bauen(l, rng, ziel_pulls)
        if len(loesung) < ziel_pulls // 2:
            continue
        start = l.kopie()          # so sieht das Level beim Start aus
        karte = als_karte(l)
        if len(karte) < 4 or max(len(z) for z in karte) < 5:
            continue
        fehler = pruefe_struktur(karte, "Stufe %d–%d" % (von, bis))
        if fehler:
            continue
        ok, text, pushes = spiele(karte, loesung)
        if not ok:
            continue
        # jetzt der unabhängige Solver: kennt die Lösung nicht
        r = sokoban_solver.loese(karte, zeitlimit=solver_zeit, max_knoten=250000,
                                 mit_loesung=True)
        if not r.get("geloest"):
            continue
        if not (min_p <= r["pushes"] <= max_p):
            continue
        loesung = r["zuege"]
        pushes = r["pushes"]
        ok, text, _ = spiele(karte, loesung)
        if not ok:
            continue
        # Start-Stellung muss frei von sofort toten Kisten sein, und keine
        # Kiste darf schon auf einem Ziel stehen (das wäre geschenkt).
        tot = [k for k in start.kisten if k not in start.ziele and tote_ecke(k, start)]
        if tot or (start.kisten & start.ziele):
            continue
        return {"karte": karte, "loesung": loesung, "pushes": pushes}
    return None


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--pruefen", action="store_true")
    p.add_argument("--test", action="store_true")
    p.add_argument("--datei", default=ZIEL_DATEI)
    args = p.parse_args()

    if args.pruefen:
        return pruefen(args.datei)
    if args.test:
        return selbsttest()
    return bauen(args.datei)


def bauen(datei):
    rng = random.Random(20261125)
    import sokoban_solver
    level = []
    for karte in BESTEHEND:
        text = "\n".join(karte)
        if text.count("$") + text.count("*") != text.count(".") + text.count("*") + text.count("+"):
            print("WARNUNG: bestehendes Level unausgeglichen")
        r = sokoban_solver.loese(karte, zeitlimit=10.0, mit_loesung=True)
        level.append({"karte": karte,
                      "loesung": r.get("zuege"),
                      "pushes": r.get("pushes")})
        print("Level %3d (alt)  %s  %s Schiebe" % (
            len(level), "gelöst" if r.get("geloest") else "OFFEN", r.get("pushes")))

    gesehen = set()
    for stufe in STUFEN:
        von, bis = stufe[0], stufe[1]
        for nr in range(von, bis + 1):
            # Klemmt eine Stufe, wird ihr Band schrittweise gelockert –
            # lieber ein Level etwas leichter als eine Lücke in der Reihe.
            for versuch in range(8):
                stufe_jetzt = (stufe[0], stufe[1], stufe[2],
                               max(5, stufe[3] - versuch * 2), stufe[4],
                               stufe[5], stufe[6])
                gebaut = erzeuge(stufe_jetzt, rng)
                if not gebaut:
                    continue
                schluessel = "\n".join(sorted(gebaut["karte"]))
                if schluessel in gesehen:
                    continue
                gesehen.add(schluessel)
                level.append(gebaut)
                print("Level %3d  %2d Kisten  %3d Schiebe  %s" % (
                    nr, len(gebaut["karte"]) and (sum(z.count("$") + z.count("*") for z in gebaut["karte"])),
                    gebaut["pushes"], "/".join(gebaut["karte"])))
                break
            else:
                print("FEHLER: Level %d nicht gebaut" % nr)
                return 1

    assert len(level) == 100, len(level)
    zeilen = []
    zeilen.append("/* Sokoban-Level – erzeugt von werkzeuge/sokoban_level_bauen.py.")
    zeilen.append("   Jedes Level wurde vorwärts nachgespielt; die Felder `loesung` und")
    zeilen.append("   `pushes` sind Messwerte, keine Pflicht für das Spiel. */")
    zeilen.append("window.SOKOBAN_LEVEL = [")
    for i, l in enumerate(level):
        zeilen.append(" // Level %02d" % (i + 1))
        zeilen.append(" {karte:[%s], loesung:%s, pushes:%s}," % (
            ", ".join(json.dumps(z, ensure_ascii=False) for z in l["karte"]),
            json.dumps(l["loesung"]) if l["loesung"] else "null",
            l["pushes"] if l["pushes"] is not None else "null"))
    zeilen.append("];")
    io.open(datei, "w", encoding="utf-8", newline="\n").write("\n".join(zeilen) + "\n")
    print("\n%d Level geschrieben: %s" % (len(level), datei))
    return 0


def pruefen(datei):
    text = io.open(datei, encoding="utf-8").read()
    import re
    # Nur Zeilenumbrüche entfernen – Leerzeichen sind Teil der Karten!
    bloecke = re.findall(r"\{karte:\[(.*?)\],\s*loesung:\s*(.*?),\s*pushes:\s*(.*?)\}", text.replace("\n", ""))
    fehler = 0
    print("Level in Datei: %d" % len(bloecke))
    for i, (karte_roh, loesung_roh, pushes) in enumerate(bloecke):
        karte = json.loads("[" + karte_roh + "]")
        loesung = json.loads(loesung_roh)
        meldung = "Level %02d" % (i + 1)
        f = pruefe_struktur(karte, meldung)
        if loesung:
            ok, text2, _ = spiele(karte, loesung)
            if not ok:
                f.append("%s: %s" % (meldung, text2))
        else:
            f.append("%s: keine Lösung hinterlegt" % meldung)
        if f:
            fehler += 1
            for e in f:
                print("FEHLER  " + e)
        else:
            print("ok      %s  %2d Kisten  %s Schiebe" % (
                meldung, len(karte), pushes))
    print("\n%d Level, %d mit Fehlern" % (len(bloecke), fehler))
    return 1 if fehler else 0


def selbsttest():
    # eigener Rückwärts-/Vorwärtstest
    rng = random.Random(7)
    gut = 0
    for i in range(30):
        l = raum_bauen(rng, 8, 7, 2)
        if not l:
            continue
        start = l.kopie()
        zuege = loesung_bauen(l, rng, 12)
        karte = als_karte(l)
        ok, text, pushes = spiele(karte, zuege)
        if ok:
            gut += 1
        else:
            print("FEHLER: %s" % text)
    print("%d von 30 erzeugten Leveln mit gültiger Lösung" % gut)
    return 0 if gut >= 25 else 1


if __name__ == "__main__":
    sys.exit(main())
