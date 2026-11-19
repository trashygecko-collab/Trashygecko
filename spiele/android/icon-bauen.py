# -*- coding: utf-8 -*-
"""Baut aus dem gelieferten Spielstein-PNG die Android-Icons.

Quelle: icon/icon-quelle.png (transparenter Rand wird beschnitten)
Ziele: app/src/main/res/mipmap-*/ic_launcher.png, ic_launcher_round.png,
       mipmap-*/ic_launcher_foreground.png  (adaptiv, Android 8+)
       icon/ic_launcher-playstore.png
"""
import io
import os

from PIL import Image, ImageDraw

HIER = os.path.dirname(os.path.abspath(__file__))
QUELLE = os.path.join(HIER, "icon", "icon-quelle.png")
RES = os.path.join(HIER, "app", "src", "main", "res")

# Hintergrund der App (css/mahjong.css: #0b241d) - passt zum Spielstein.
HINTERGRUND = (11, 36, 29, 255)

# Dichte -> Kantenlaenge des Legacy-Icons in Pixeln
DICHTEN = {
    "mdpi": 48,
    "hdpi": 72,
    "xhdpi": 96,
    "xxhdpi": 144,
    "xxxhdpi": 192,
}


def lade_stein():
    bild = Image.open(QUELLE).convert("RGBA")
    rand = bild.getbbox()
    return bild.crop(rand) if rand else bild


def setze(stein, kante, anteil, rund=False):
    """Stein mittig auf Hintergrund; rund=True maskiert das Icon kreisfoermig."""
    if rund:
        flaeche = Image.new("RGBA", (kante, kante), (0, 0, 0, 0))
        maske = Image.new("L", (kante, kante), 0)
        ImageDraw.Draw(maske).ellipse((0, 0, kante - 1, kante - 1), fill=255)
        kreis = Image.new("RGBA", (kante, kante), HINTERGRUND)
        flaeche.paste(kreis, (0, 0), maske)
    else:
        flaeche = Image.new("RGBA", (kante, kante), HINTERGRUND)

    ziel = max(1, int(round(kante * anteil)))
    klein = stein.copy()
    klein.thumbnail((ziel, ziel), Image.LANCZOS)

    x = (kante - klein.width) // 2
    y = (kante - klein.height) // 2
    flaeche.paste(klein, (x, y), klein)

    return flaeche


def schreibe(bild, pfad):
    os.makedirs(os.path.dirname(pfad), exist_ok=True)
    bild.save(pfad, "PNG", optimize=True)


stein = lade_stein()
print("Stein nach Beschnitt:", stein.size)

for dichte, kante in DICHTEN.items():
    ordner = os.path.join(RES, "mipmap-" + dichte)

    # Klassisch: quadratisch, Stein mit Rand (Launcher maskieren selbst).
    schreibe(setze(stein, kante, 0.78), os.path.join(ordner, "ic_launcher.png"))
    # Rund (Android 7.1 und aelter, runde Launcher).
    schreibe(setze(stein, kante, 0.66, rund=True), os.path.join(ordner, "ic_launcher_round.png"))

    # Adaptiv: 108dp Flaeche, Motiv innerhalb der sicheren 66dp-Zone.
    adaptiv = int(round(kante * 108 / 48))
    schreibe(setze(stein, adaptiv, 0.56), os.path.join(ordner, "ic_launcher_foreground.png"))
    print("  %-8s Legacy %d, adaptiv %d" % (dichte, kante, adaptiv))

# Play-Store-Icon (512 x 512, ohne Transparenz)
schreibe(setze(stein, 512, 0.76).convert("RGB"), os.path.join(HIER, "icon", "ic_launcher-playstore.png"))
print("fertig")
