"""Schneidet die Motiv-Kontaktboegen in Einzelbilder, stellt den Hintergrund frei
und skaliert auf Spielgroesse (160 x 160, transparenter Rand).

Der Hintergrund wird pro Zelle aus dem Rand als Verlauf geschaetzt; Zellen mit
echtem Fotohintergrund bestehen die Qualitaetspruefung nicht und landen in
`unsauber/` statt in `klein/`.

Aufruf:  python _bilder_klein.py
"""
import os
import sys

import numpy as np
from PIL import Image, ImageDraw, ImageFilter

QUELLE = os.path.join("..", "mahjong", "assets", "arbeit")
ZIEL = os.path.join("..", "mahjong", "assets", "klein")
AUSSCHUSS = os.path.join("..", "mahjong", "assets", "klein", "unsauber")
FINAL = 160
RAND = 0.06
SPALTEN, ZEILEN = 6, 6
T0, T1 = 30.0, 72.0      # Farbabstand: bis T0 Hintergrund, ab T1 Motiv


def gruenstreifen(im):
    arr = np.asarray(im.convert("RGB"), np.float32)
    r, g, b = arr[..., 0], arr[..., 1], arr[..., 2]
    return (g > 110) & (g > r + 40) & (g > b + 40)


def faecher(im):
    """Zellgrenzen aus den gruenen Zwischenstreifen ableiten."""
    gruen = gruenstreifen(im)

    def laeufe(anteil):
        ist_streifen = anteil > 0.75
        gebiete, start = [], None
        for i, v in enumerate(ist_streifen):
            if v and start is None:
                start = i
            elif not v and start is not None:
                gebiete.append((start, i))
                start = None
        if start is not None:
            gebiete.append((start, len(ist_streifen)))
        zellen, letzte = [], 0
        for a, b in gebiete:
            if a - letzte > 60:
                zellen.append((letzte, a))
            letzte = b
        if len(ist_streifen) - letzte > 60:
            zellen.append((letzte, len(ist_streifen)))
        return zellen

    spalten = laeufe(gruen.mean(axis=0))
    zeilen = laeufe(gruen.mean(axis=1))
    if len(spalten) != SPALTEN or len(zeilen) != ZEILEN:
        breite, hoehe = im.size
        spalten = [(int(i * breite / SPALTEN), int((i + 1) * breite / SPALTEN))
                   for i in range(SPALTEN)]
        zeilen = [(int(i * hoehe / ZEILEN), int((i + 1) * hoehe / ZEILEN))
                  for i in range(ZEILEN)]
    return spalten, zeilen


def hintergrund_verlauf(arr):
    """Hintergrundfarbe pro Pixel aus den Zellraendern fortschreiben."""
    hoehe, breite = arr.shape[:2]
    rand_x = max(3, breite // 25)
    rand_y = max(3, hoehe // 25)
    links = np.median(arr[:, :rand_x], axis=1)
    rechts = np.median(arr[:, -rand_x:], axis=1)
    oben = np.median(arr[:rand_y], axis=0)
    unten = np.median(arr[-rand_y:], axis=0)

    waagerecht = (links + rechts) / 2.0
    senkrecht = (oben + unten) / 2.0
    verlauf = (waagerecht[:, None, :] + senkrecht[None, :, :]) / 2.0
    return verlauf, np.concatenate([links, rechts, oben, unten])


def freistellen(kachel, chroma):
    arr = np.asarray(kachel.convert("RGB"), np.float32)
    verlauf, randfarben = hintergrund_verlauf(arr)

    abstand = np.sqrt(((arr - verlauf) ** 2).sum(axis=2))
    alpha = np.clip((abstand - T0) / (T1 - T0), 0.0, 1.0)

    # gruener Kontaktbogen-Hintergrund und Bogenfarbe hart entfernen
    r, g, b = arr[..., 0], arr[..., 1], arr[..., 2]
    gruen = (g > 105) & (g > r + 34) & (g > b + 34)
    alpha[gruen] = 0.0
    if chroma is not None:
        fern = np.sqrt(((arr - chroma) ** 2).sum(axis=2))
        alpha[fern < 50.0] = 0.0
        alpha = np.where(fern < 76.0, alpha * np.clip((fern - 50.0) / 26.0, 0, 1), alpha)
    alpha = np.asarray(Image.fromarray((alpha * 255).astype(np.uint8))
                       .filter(ImageFilter.GaussianBlur(1.0)), np.float32) / 255.0

    a = np.repeat(alpha[..., None], 3, axis=2)
    grund = np.where(gruen[..., None], np.array([40.0, 190.0, 90.0]), verlauf)
    farbe = np.where(a > 0.02, (arr - grund * (1.0 - a)) / np.maximum(a, 0.02), arr)
    farbe = np.clip(farbe, 0, 255)

    # Qualitaet: bleibt der Rand opak, ist der Hintergrund ein Foto
    hoehe, breite = alpha.shape
    d = max(2, min(hoehe, breite) // 25)
    ring = np.concatenate([alpha[:d].ravel(), alpha[-d:].ravel(),
                           alpha[:, :d].ravel(), alpha[:, -d:].ravel()])
    rand_opak = float((ring > 0.9).mean())
    rest_hintergrund = float(ring.mean())
    fuellung = float((alpha > 0.5).mean())
    sauber = rand_opak < 0.32 and rest_hintergrund < 0.24 and 0.03 < fuellung < 0.97
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


def kachel_bauen(kachel, ziel_pfad, chroma=None):
    farbe, alpha, sauber = freistellen(kachel, chroma)
    bild = Image.fromarray(np.dstack([farbe, alpha]), "RGBA")
    ausgerichtet = ausrichten(bild)
    if not sauber:
        ziel_pfad = os.path.join(AUSSCHUSS, os.path.basename(ziel_pfad))
    ausgerichtet.save(ziel_pfad)
    return ausgerichtet, sauber


def einzelbild(pfad, ziel_pfad):
    """Einzelmotiv mit Schachbrett-Hintergrund freistellen."""
    im = Image.open(pfad).convert("RGB")
    arr = np.asarray(im, np.float32)
    hell = (arr.max(axis=2) - arr.min(axis=2) < 16) & (arr.mean(axis=2) > 175)

    marke = Image.fromarray((hell * 255).astype(np.uint8))
    von_rand = marke.copy()
    ImageDraw.floodfill(von_rand, (0, 0), 128, thresh=1)
    aussen = np.asarray(von_rand) == 128
    alpha = np.asarray(Image.fromarray(np.where(aussen, 0, 255).astype(np.uint8))
                       .filter(ImageFilter.GaussianBlur(1.0)), np.float32) / 255.0

    grau = np.where((arr.mean(axis=2) > 240)[..., None], 255.0, 230.0)
    a = np.repeat(alpha[..., None], 3, axis=2)
    farbe = np.where(a > 0.02, (arr - grau * (1.0 - a)) / np.maximum(a, 0.02), arr)
    bild = Image.fromarray(np.dstack([np.clip(farbe, 0, 255).astype(np.uint8),
                                      (alpha * 255).astype(np.uint8)]), "RGBA")
    ausrichten(bild).save(ziel_pfad)


def bogenfarbe(im):
    """Haeufigste Farbe des Bogens (blauer bzw. gruener Kasten)."""
    klein = np.asarray(im.resize((120, 120), Image.NEAREST).convert("RGB"), np.int32)
    q = (klein // 10)
    schluessel = q[..., 0] * 10000 + q[..., 1] * 100 + q[..., 2]
    werte, zahlen = np.unique(schluessel, return_counts=True)
    k = werte[int(np.argmax(zahlen))]
    treffer = klein[schluessel == k].mean(axis=0)
    return treffer.astype(np.float32)


def bogen_zerlegen(pfad, nummer_start):
    im = Image.open(pfad).convert("RGB")
    chroma = bogenfarbe(im)
    spalten, zeilen = faecher(im)
    sauber, unsauber = [], []
    n = nummer_start
    for y0, y1 in zeilen:
        for x0, x1 in spalten:
            kachel = im.crop((x0 + 8, y0 + 8, x1 - 8, y1 - 8))
            ziel = os.path.join(ZIEL, "motiv-%03d.png" % n)
            _, ok = kachel_bauen(kachel, ziel, chroma)
            (sauber if ok else unsauber).append(n)
            n += 1
    return sauber, unsauber, n


if __name__ == "__main__":
    os.makedirs(ZIEL, exist_ok=True)
    os.makedirs(AUSSCHUSS, exist_ok=True)
    for alt in os.listdir(ZIEL):
        if alt.endswith(".png"):
            os.remove(os.path.join(ZIEL, alt))
    for alt in os.listdir(AUSSCHUSS):
        if alt.endswith(".png"):
            os.remove(os.path.join(AUSSCHUSS, alt))

    dateien = [os.path.join(QUELLE, f) for f in sorted(os.listdir(QUELLE))
               if f.lower().endswith((".jpg", ".jpeg", ".png"))
               and not f.startswith(("Gemini_Generated_Image_9sbdsg", "a_clean"))]

    nummer = 1
    alle_sauber, alle_unsauber = [], []
    for pfad in dateien:
        sauber, unsauber, nummer = bogen_zerlegen(pfad, nummer)
        alle_sauber += sauber
        alle_unsauber += unsauber
        print("%-45s sauber %2d / unsauber %2d" % (os.path.basename(pfad),
                                                   len(sauber), len(unsauber)))

    einzelbild(os.path.join(QUELLE, "Gemini_Generated_Image_9sbdsg9sbdsg9sbd.jpg"),
               os.path.join(ZIEL, "motiv-%03d.png" % nummer))
    alle_sauber.append(nummer)
    print("Schmetterling:", nummer)
    print("gesamt sauber:", len(alle_sauber), " unsauber:", len(alle_unsauber))
    with open(os.path.join(QUELLE, "..", "_klein_bericht.txt"), "w", encoding="utf-8") as f:
        f.write("sauber: %s\n\nunsauber: %s\n" % (alle_sauber, alle_unsauber))
