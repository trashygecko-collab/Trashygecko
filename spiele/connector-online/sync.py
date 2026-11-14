# -*- coding: utf-8 -*-
"""Holt die gemeinsamen Dateien aus ../connector hierher und baut die Online-Fassung.

Aufruf:  python sync.py            (nur kopieren und Windows-exe bauen)
         python sync.py --linux    (zusätzlich eine Linux-Datei für den Server)
         python sync.py --nur-kopieren

Was hier bleibt und nicht überschrieben wird: main.go, netz.go, status.go,
qr.go, sync.py, README.md und alles, was nicht aus ../connector stammt.
"""

import io
import os
import shutil
import subprocess
import sys

HIER = os.path.dirname(os.path.abspath(__file__))
QUELLE = os.path.normpath(os.path.join(HIER, "..", "connector"))

# Diese Dateien gibt es nur in der Online-Fassung; die lokalen Fassungen
# (main.go, status.go, qr.go) dürfen sie nicht überschreiben.
EIGENE = {"main.go", "netz.go", "status.go", "qr.go"}
# Zusätzlich werden diese Datendateien mitgenommen.
DATEN = ["connector.html", "go.mod", "go.sum"]
# Nicht kopieren, auch wenn sie in ../connector liegen.
IGNORIEREN = {"sync.py", "README.md", ".DS_Store"}


def gemeinsame_dateien():
    namen = []
    for name in sorted(os.listdir(QUELLE)):
        if not name.endswith(".go"):
            continue
        if name in EIGENE or name in IGNORIEREN:
            continue
        namen.append(name)
    namen.extend(DATEN)
    return [n for n in namen if os.path.isfile(os.path.join(QUELLE, n))]


def gleich(a, b):
    if not os.path.isfile(b):
        return False
    with io.open(a, "rb") as fa, io.open(b, "rb") as fb:
        return fa.read() == fb.read()


def kopieren():
    neu, geaendert, gleichgeblieben = [], [], []
    for name in gemeinsame_dateien():
        quelle = os.path.join(QUELLE, name)
        ziel = os.path.join(HIER, name)
        if not os.path.isfile(ziel):
            neu.append(name)
        elif not gleich(quelle, ziel):
            geaendert.append(name)
        else:
            gleichgeblieben.append(name)
        shutil.copy2(quelle, ziel)

    if neu:
        print("neu:        " + ", ".join(neu))
    if geaendert:
        print("aktualisiert: " + ", ".join(geaendert))
    if not neu and not geaendert:
        print("alles schon aktuell (%d Dateien)" % len(gleichgeblieben))
    return neu, geaendert


def bauen(linux=False):
    print("go vet …")
    subprocess.run(["go", "vet", "./..."], cwd=HIER, check=True)

    print("baue trashygecko-connector-online.exe …")
    subprocess.run(
        ["go", "build", "-o", "trashygecko-connector-online.exe", "."],
        cwd=HIER,
        check=True,
    )

    if linux:
        print("baue trashygecko-connector-online-linux (amd64) …")
        umgebung = dict(os.environ, GOOS="linux", GOARCH="amd64", CGO_ENABLED="0")
        subprocess.run(
            ["go", "build", "-o", "trashygecko-connector-online-linux", "."],
            cwd=HIER,
            check=True,
            env=umgebung,
        )


def main():
    if not os.path.isdir(QUELLE):
        raise SystemExit("Quellordner fehlt: %s" % QUELLE)

    kopieren()
    if "--nur-kopieren" in sys.argv:
        return
    bauen(linux="--linux" in sys.argv)


if __name__ == "__main__":
    main()
