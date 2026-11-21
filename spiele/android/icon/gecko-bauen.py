# -*- coding: utf-8 -*-
"""Zeichnet das TrashyGecko-Motiv (Gecko von oben) fuer die App-Icons.

Ergebnis: icon/icon-trashygecko.png (1024 px, transparenter Rand).
Weiterverarbeitung mit icon-bauen.py -> mipmap-*/ic_launcher*.png.
Farben aus der Homepage: dunkelgruen #0B241D, Gold #E8CD8E.
"""
import math
import os

from PIL import Image, ImageDraw

HIER = os.path.dirname(os.path.abspath(__file__))
ZIEL = os.path.join(HIER, "icon-trashygecko.png")

KANTE = 1024
S = 4  # Ueberabtastung, danach verkleinern = weiche Kanten

GOLD = (232, 205, 142, 255)
GOLD_DUNKEL = (178, 146, 88, 255)
PUNKT = (150, 120, 70, 255)
AUGE = (11, 36, 29, 255)


def px(wert):
    return int(round(wert * S))


def ellipse(draw, cx, cy, rx, ry, fill, outline=None, breite=0):
    draw.ellipse([px(cx - rx), px(cy - ry), px(cx + rx), px(cy + ry)],
                 fill=fill, outline=outline, width=px(breite))


def kubisch(p0, p1, p2, p3, t):
    u = 1 - t
    return (u * u * u * p0[0] + 3 * u * u * t * p1[0] + 3 * u * t * t * p2[0] + t * t * t * p3[0],
            u * u * u * p0[1] + 3 * u * u * t * p1[1] + 3 * u * t * t * p2[1] + t * t * t * p3[1])


def schwanz(draw, segmente, r_start, r_ende, schritte=160):
    """Zeichnet den Schwanz als Perlenschnur entlang zweier Bezier-Stuecke."""
    gesamt = len(segmente) * schritte
    for i in range(gesamt + 1):
        t = i / gesamt
        stueck = min(len(segmente) - 1, int(t * len(segmente)))
        lokal = (t * len(segmente)) - stueck
        x, y = kubisch(*segmente[stueck], lokal)
        radius = r_start + (r_ende - r_start) * (i / gesamt)
        ellipse(draw, x, y, radius, radius, GOLD)


def zehe(draw, punkt, vorher, radius=19):
    """Drei gespreizte Zehen am Ende einer Gliedmasse."""
    winkel = math.atan2(punkt[1] - vorher[1], punkt[0] - vorher[0])
    for abweichung in (-0.55, 0.0, 0.55):
        a = winkel + abweichung
        x = punkt[0] + math.cos(a) * 30
        y = punkt[1] + math.sin(a) * 30
        ellipse(draw, x, y, radius, radius, GOLD)


def glied(draw, a, b, c, breite=(42, 34)):
    draw.line([px(a[0]), px(a[1]), px(b[0]), px(b[1])], fill=GOLD, width=px(breite[0]))
    draw.line([px(b[0]), px(b[1]), px(c[0]), px(c[1])], fill=GOLD, width=px(breite[1]))
    for p in (a, b, c):
        ellipse(draw, p[0], p[1], breite[1] / 2, breite[1] / 2, GOLD)
    zehe(draw, c, b)


def baue():
    bild = Image.new("RGBA", (KANTE * S, KANTE * S), (0, 0, 0, 0))
    draw = ImageDraw.Draw(bild)

    # Beine zuerst, damit der Koerper sauber darueber liegt.
    glied(draw, (398, 432), (300, 486), (206, 438))          # vorn links
    glied(draw, (626, 432), (724, 486), (818, 438))          # vorn rechts
    glied(draw, (404, 660), (288, 724), (182, 768))          # hinten links
    glied(draw, (620, 660), (736, 724), (842, 768))          # hinten rechts
    ellipse(draw, 384, 648, 58, 66, GOLD)                    # Schenkel links
    ellipse(draw, 640, 648, 58, 66, GOLD)                    # Schenkel rechts

    # Schwanz: nach unten und ohne Schleife zur Seite auslaufend.
    schwanz(draw, [
        ((512, 742), (516, 906), (664, 980), (852, 916)),
    ], 50, 8)

    ellipse(draw, 512, 512, 172, 238, GOLD)                  # Koerper
    ellipse(draw, 512, 336, 118, 104, GOLD)                  # Hals
    ellipse(draw, 512, 250, 114, 134, GOLD)                  # Kopf
    ellipse(draw, 408, 214, 44, 44, GOLD)                    # Augenwulst links
    ellipse(draw, 616, 214, 44, 44, GOLD)                    # Augenwulst rechts
    ellipse(draw, 414, 218, 20, 20, AUGE)                    # Auge links
    ellipse(draw, 610, 218, 20, 20, AUGE)                    # Auge rechts

    for cx, cy, rx, ry in ((468, 462, 26, 32), (566, 520, 21, 27),
                           (492, 604, 24, 30), (556, 664, 18, 24)):
        ellipse(draw, cx, cy, rx, ry, GOLD_DUNKEL)


    return bild.resize((KANTE, KANTE), Image.LANCZOS)


if __name__ == "__main__":
    baue().save(ZIEL, "PNG", optimize=True)
    print("Icon-Quelle:", ZIEL)
