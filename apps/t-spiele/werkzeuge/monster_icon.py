"""Malt das T.Spiele-Monster und schreibt alle Android-Symbolgrößen.

Aufruf (im Projektordner):  python werkzeuge/monster_icon.py
Alles wird vierfach überabgetastet gezeichnet und danach glatt verkleinert,
damit die Kanten auch als kleines Symbol sauber aussehen.
"""

import os

from PIL import Image, ImageDraw

S = 4                     # Überabtastung
HIER = os.path.dirname(os.path.abspath(__file__))
PROJEKT = os.path.dirname(HIER)
RES = os.path.join(PROJEKT, "res")

KOERPER = "#8fe39b"
KOERPER_RAND = "#57bd74"
BAUCH = "#c6f4cd"
HORN = "#57bd74"
WANGE = (255, 154, 180, 150)
MUND = "#2f2f2f"


def monster_flaeche():
    """Monster auf durchsichtigem Grund, Kantenlänge 1024 * S."""
    n = 1024 * S
    bild = Image.new("RGBA", (n, n), (0, 0, 0, 0))
    d = ImageDraw.Draw(bild)
    w = n * 0.80            # Breite des Monsters
    x0, y0 = (n - w) / 2, (n - w) / 2
    cx = x0 + w / 2

    def p(fx, fy):
        return (x0 + w * fx, y0 + w * fy)

    def kasten(fx, fy, frx, fry):
        x, y = p(fx, fy)
        return [x - w * frx, y - w * fry, x + w * frx, y + w * fry]

    # Fühler mit Kugeln
    for seite in (-1, 1):
        sx = 0.5 + seite * 0.14
        ende = (0.5 + seite * 0.20, 0.215)
        d.line([p(sx, 0.33), p(ende[0], ende[1])], fill=HORN, width=int(w * 0.028))
        ex, ey = p(ende[0], ende[1])
        r = w * 0.048
        d.ellipse([ex - r, ey - r, ex + r, ey + r], fill=HORN)

    # Füße und Arme
    for seite in (-1, 1):
        d.ellipse(kasten(0.5 + seite * 0.13, 0.805, 0.105, 0.062), fill=KOERPER_RAND)
        d.ellipse(kasten(0.5 + seite * 0.255, 0.615, 0.062, 0.10), fill=KOERPER_RAND)

    # Körper
    d.ellipse(kasten(0.5, 0.555, 0.245, 0.265), fill=KOERPER, outline=KOERPER_RAND,
              width=int(w * 0.012))
    d.ellipse(kasten(0.5, 0.66, 0.145, 0.155), fill=BAUCH)

    # Augen
    for seite in (-1, 1):
        auge = kasten(0.5 + seite * 0.10, 0.46, 0.075, 0.085)
        d.ellipse(auge, fill="#ffffff")
        ax = (auge[0] + auge[2]) / 2 + w * 0.006
        ay = (auge[1] + auge[3]) / 2 + w * 0.008
        rp = w * 0.037
        d.ellipse([ax - rp, ay - rp, ax + rp, ay + rp], fill="#22222a")
        gl = w * 0.014
        gx, gy = ax - w * 0.013, ay - w * 0.020
        d.ellipse([gx - gl, gy - gl, gx + gl, gy + gl], fill="#ffffff")

    # Wangen
    for seite in (-1, 1):
        d.ellipse(kasten(0.5 + seite * 0.185, 0.605, 0.048, 0.032), fill=WANGE)

    # Mund mit zwei kleinen Zähnchen
    d.arc(kasten(0.5, 0.578, 0.062, 0.038), start=20, end=160, fill=MUND,
          width=int(w * 0.018))
    for seite in (-1, 1):
        zx, zy = p(0.5 + seite * 0.022, 0.585)
        b = w * 0.016
        d.polygon([(zx - b, zy), (zx + b, zy), (zx, zy + b * 1.5)], fill="#ffffff")

    return bild


def hintergrund(n):
    """Abgerundeter Kachel-Hintergrund mit weichem Verlauf."""
    bild = Image.new("RGBA", (n, n), (0, 0, 0, 0))
    verlauf = Image.new("RGB", (1, n))
    oben, unten = (47, 138, 115), (16, 51, 44)
    for y in range(n):
        t = y / max(1, n - 1)
        verlauf.putpixel((0, y), tuple(
            int(oben[i] + (unten[i] - oben[i]) * t) for i in range(3)))
    bild.paste(verlauf.resize((n, n)), (0, 0))

    maske = Image.new("L", (n, n), 0)
    md = ImageDraw.Draw(maske)
    ecke = int(n * 0.22)
    md.rounded_rectangle([0, 0, n - 1, n - 1], radius=ecke, fill=255)
    bild.putalpha(maske)

    # Weiches Licht hinter dem Monster: viele Ringe mit wenig Deckkraft, damit
    # kein harter Kreisrand entsteht.
    licht = Image.new("RGBA", (n, n), (0, 0, 0, 0))
    ld = ImageDraw.Draw(licht)
    for i in range(30):
        r = n * 0.36 * (1 - i / 30)
        ld.ellipse([n / 2 - r, n / 2 - r, n / 2 + r, n / 2 + r],
                   fill=(255, 255, 255, 3))
    bild.alpha_composite(licht)
    return bild


def schreibe(pfad, bild, groesse):
    os.makedirs(os.path.dirname(pfad), exist_ok=True)
    bild.resize((groesse, groesse), Image.LANCZOS).save(pfad, "PNG")


def main():
    gross = 1024 * S

    monster = monster_flaeche()

    # Altes Symbol: Hintergrund + Monster (Android bis 7.x)
    for dichte, groesse in (("mdpi", 48), ("hdpi", 72), ("xhdpi", 96),
                            ("xxhdpi", 144), ("xxxhdpi", 192)):
        n = groesse * S
        kachel = hintergrund(n)
        m = monster.resize((int(n * 0.86), int(n * 0.86)), Image.LANCZOS)
        kachel.alpha_composite(m, (int((n - m.width) / 2), int((n - m.height) / 2)))
        schreibe(os.path.join(RES, "mipmap-" + dichte, "ic_launcher.png"),
                 kachel, groesse)

    # Adaptives Symbol (Android 8+): nur das Monster, im sicheren Bereich
    for dichte, groesse in (("mdpi", 108), ("hdpi", 162), ("xhdpi", 216),
                            ("xxhdpi", 324), ("xxxhdpi", 432)):
        n = groesse * S
        ebene = Image.new("RGBA", (n, n), (0, 0, 0, 0))
        m = monster.resize((int(n * 0.60), int(n * 0.60)), Image.LANCZOS)
        ebene.alpha_composite(m, (int((n - m.width) / 2), int((n - m.height) / 2)))
        schreibe(os.path.join(RES, "mipmap-" + dichte, "ic_launcher_fg.png"),
                 ebene, groesse)

    # Vorschaubild für README und Laden
    vorschau = hintergrund(gross)
    m = monster.resize((int(gross * 0.86), int(gross * 0.86)), Image.LANCZOS)
    vorschau.alpha_composite(m, (int((gross - m.width) / 2), int((gross - m.height) / 2)))
    vorschau.resize((512, 512), Image.LANCZOS).save(
        os.path.join(PROJEKT, "symbol-vorschau.png"), "PNG")

    print("Symbole geschrieben nach", RES)


if __name__ == "__main__":
    main()
