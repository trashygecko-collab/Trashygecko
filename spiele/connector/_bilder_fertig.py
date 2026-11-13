"""Bereitet Motiv-Raster aus I:/Downloads/arbeit fuer das Mahjong-Spiel auf.

Jede Zelle wird ausgeschnitten, der einfarbige Hintergrund freigestellt und das
Motiv auf 160 x 160 mit transparentem Rand gesetzt. Zellen, deren Hintergrund
ein echtes Foto ist, landen in `fertig/unsauber/`.

Aufruf:  python _bilder_fertig.py
"""
import os

import numpy as np
from PIL import Image, ImageDraw, ImageFilter

QUELLE = r"I:/Downloads/arbeit"
ZIEL = os.path.join(QUELLE, "fertig")
AUSSCHUSS = os.path.join(ZIEL, "unsauber")
FINAL = 160
RAND = 0.06
SPALTEN, ZEILEN = 6, 6
RANDZONE = 13


def grundfarbe(im):
    klein = np.asarray(im.resize((120, 120), Image.NEAREST).convert("RGB"), np.int32)
    q = klein // 10
    schluessel = q[..., 0] * 10000 + q[..., 1] * 100 + q[..., 2]
    werte, zahlen = np.unique(schluessel, return_counts=True)
    haeufigste = werte[int(np.argmax(zahlen))]
    return klein[schluessel == haeufigste].mean(axis=0).astype(np.float32)


def faecher(im, grund):
    """Zellgrenzen aus den einfarbigen Zwischenstreifen ableiten."""
    arr = np.asarray(im.convert("RGB"), np.float32)
    leer = np.sqrt(((arr - grund) ** 2).sum(axis=2)) < 45

    def laeufe(anteil):
        streifen = anteil > 0.75
        gebiete, start = [], None
        for i, v in enumerate(streifen):
            if v and start is None:
                start = i
            elif not v and start is not None:
                gebiete.append((start, i))
                start = None
        if start is not None:
            gebiete.append((start, len(streifen)))

        zellen, letzte = [], 0
        for a, b in gebiete:
            if a - letzte > 60:
                zellen.append((letzte, a))
            letzte = b
        if len(streifen) - letzte > 60:
            zellen.append((letzte, len(streifen)))
        return zellen

    spalten = laeufe(leer.mean(axis=0))
    zeilen = laeufe(leer.mean(axis=1))

    if len(spalten) != SPALTEN or len(zeilen) != ZEILEN:
        breite, hoehe = im.size
        spalten = [(int(i * breite / SPALTEN), int((i + 1) * breite / SPALTEN))
                   for i in range(SPALTEN)]
        zeilen = [(int(i * hoehe / ZEILEN), int((i + 1) * hoehe / ZEILEN))
                  for i in range(ZEILEN)]

    return spalten, zeilen


def hintergrund_verlauf(arr):
    hoehe, breite = arr.shape[:2]
    rand_x = max(3, breite // 25)
    rand_y = max(3, hoehe // 25)
    links = np.median(arr[:, :rand_x], axis=1)
    rechts = np.median(arr[:, -rand_x:], axis=1)
    oben = np.median(arr[:rand_y], axis=0)
    unten = np.median(arr[-rand_y:], axis=0)
    waagerecht = (links + rechts) / 2.0
    senkrecht = (oben + unten) / 2.0
    return (waagerecht[:, None, :] + senkrecht[None, :, :]) / 2.0


def freistellen(kachel, chroma):
    arr = np.asarray(kachel.convert("RGB"), np.float32)
    verlauf = hintergrund_verlauf(arr)

    abstand = np.sqrt(((arr - verlauf) ** 2).sum(axis=2))
    alpha = np.clip((abstand - 30.0) / 42.0, 0.0, 1.0)

    # gruener Resthintergrund
    r, g, b = arr[..., 0], arr[..., 1], arr[..., 2]
    gruen = (g > 105) & (g > r + 34) & (g > b + 34)
    alpha[gruen] = 0.0

    # Rasterfarbe und zelleigene Randfarbe hart entfernen, mit weichem Uebergang
    ecken = np.concatenate([arr[:3].reshape(-1, 3), arr[-3:].reshape(-1, 3),
                            arr[:, :3].reshape(-1, 3), arr[:, -3:].reshape(-1, 3)])
    eigen = np.median(ecken, axis=0)
    for farbe in (chroma, eigen):
        fern = np.sqrt(((arr - farbe) ** 2).sum(axis=2))
        alpha[fern < 62.0] = 0.0
        alpha = np.where(fern < 92.0, alpha * np.clip((fern - 62.0) / 30.0, 0, 1), alpha)

    a = np.repeat(alpha[..., None], 3, axis=2)
    grund = np.where(gruen[..., None], np.array([40.0, 190.0, 90.0]), verlauf)
    farbe = np.where(a > 0.02, (arr - grund * (1.0 - a)) / np.maximum(a, 0.02), arr)
    farbe = np.clip(farbe, 0, 255)

    hoehe, breite = alpha.shape
    d = max(2, min(hoehe, breite) // 25)
    ring = np.concatenate([alpha[:d].ravel(), alpha[-d:].ravel(),
                           alpha[:, :d].ravel(), alpha[:, -d:].ravel()])
    sauber = float((ring > 0.9).mean()) < 0.32 and float(ring.mean()) < 0.24 \
        and 0.03 < float((alpha > 0.5).mean()) < 0.97

    return farbe.astype(np.uint8), (alpha * 255).astype(np.uint8), sauber


def ausrichten(bild):
    box = bild.getbbox()
    if box:
        bild = bild.crop(box)
    seite = max(bild.size)
    rand = int(seite * RAND)
    leinwand = Image.new("RGBA", (seite + 2 * rand, seite + 2 * rand), (0, 0, 0, 0))
    leinwand.alpha_composite(bild, ((seite + 2 * rand - bild.width) // 2,
                                    (seite + 2 * rand - bild.height) // 2))
    return leinwand.resize((FINAL, FINAL), Image.LANCZOS)


def uebersicht(dateien, pfad, spalten=12):
    zeilen = (len(dateien) + spalten - 1) // spalten
    blatt = Image.new("RGBA", (spalten * 128, zeilen * 148), (247, 243, 236, 255))
    zeichner = ImageDraw.Draw(blatt)
    for i, pfad_datei in enumerate(dateien):
        x, y = (i % spalten) * 128, (i // spalten) * 148
        blatt.alpha_composite(Image.open(pfad_datei), (x + 14, y + 6))
        zeichner.text((x + 14, y + 174 - 22), os.path.basename(pfad_datei)
                      .replace("motiv-", "").replace(".png", ""), fill=(50, 35, 28))
    blatt.convert("RGB").save(pfad, quality=86)


def main():
    os.makedirs(ZIEL, exist_ok=True)
    os.makedirs(AUSSCHUSS, exist_ok=True)

    for alt in os.listdir(ZIEL):
        if alt.endswith(".png"):
            os.remove(os.path.join(ZIEL, alt))
    for alt in os.listdir(AUSSCHUSS):
        if alt.endswith(".png"):
            os.remove(os.path.join(AUSSCHUSS, alt))

    nummer = 1
    sauber, unsauber = [], []

    for name in sorted(os.listdir(QUELLE)):
        if name.startswith("uebersicht") or not name.lower().endswith((".png", ".jpg", ".jpeg")):
            continue

        pfad = os.path.join(QUELLE, name)
        if not os.path.isfile(pfad):
            continue

        im = Image.open(pfad).convert("RGB")
        grund = grundfarbe(im)
        spalten, zeilen = faecher(im, grund)

        for y0, y1 in zeilen:
            for x0, x1 in spalten:
                kachel = im.crop((x0 + RANDZONE, y0 + RANDZONE,
                                  x1 - RANDZONE, y1 - RANDZONE))
                farbe, alpha, ok = freistellen(kachel, grund)
                bild = ausrichten(Image.fromarray(
                    np.dstack([farbe, alpha]), "RGBA"))
                ziel = os.path.join(ZIEL, "motiv-%03d.png" % nummer)
                if not ok:
                    ziel = os.path.join(AUSSCHUSS, "motiv-%03d.png" % nummer)
                bild.save(ziel)
                (sauber if ok else unsauber).append(ziel)
                nummer += 1

        print("%-45s %d x %d Zellen" % (name, len(spalten), len(zeilen)))

    uebersicht(sauber, os.path.join(ZIEL, "uebersicht.jpg"))
    uebersicht(unsauber, os.path.join(ZIEL, "uebersicht-unsauber.jpg"))
    print("sauber:", len(sauber), "| unsauber:", len(unsauber))


if __name__ == "__main__":
    main()
