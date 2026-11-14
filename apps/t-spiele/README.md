# T.Spiele – Android-App

Eine winzige App, die nichts weiter tut, als die Connector-Adresse im Browser zu
öffnen:

    https://trashygecko.onrender.com/connector

Kein eigener Browser, keine Oberfläche, keine Berechtigungen außer Internet.
Antippen → Browser geht auf → die App ist wieder weg.

| | |
|---|---|
| Datei | `T.Spiele.apk` (in diesem Ordner, ca. 100 KB) |
| Name im Startmenü | **T.Spiele** |
| Paket | `de.trashygecko.tspiele` |
| Läuft ab | Android 4.0 (API 14) bis Android 15 (Ziel API 34) |
| Symbol | süßes grünes Monster (`symbol-vorschau.png`) |

## Installieren

1. Die Datei `T.Spiele.apk` aufs Handy bringen. Am einfachsten über die
   Homepage: Der Connector liefert sie mit aus.
   - im WLAN: `http://<IP-des-PCs>:8765/apps/t-spiele/T.Spiele.apk`
   - im Internet: `https://trashygecko.onrender.com/apps/t-spiele/T.Spiele.apk`
2. Die heruntergeladene Datei antippen. Android fragt, ob die App aus dieser
   Quelle installiert werden darf → **zulassen** („Unbekannte Apps“ bzw.
   „Installation aus dieser Quelle erlauben“).
3. Installieren, fertig. Im Startmenü erscheint **T.Spiele** mit dem Monster.

Alternativ per Kabel: `adb install -r T.Spiele.apk`

## Etwas ändern

**Andere Adresse:** in `src/de/trashygecko/tspiele/MainActivity.java` die Zeile
`ADRESSE` ändern, dann neu bauen.

**Anderes Symbol:** `werkzeuge/monster_icon.py` malen lassen
(`python werkzeuge/monster_icon.py`) – es schreibt alle Größen nach `res/`.
Die Formen stehen oben im Skript (Körper, Augen, Wangen, Mund).

**Neu bauen:**

    bash bauen.sh

Beim ersten Mal lädt das Skript ein portables JDK 17 und die Android-Werkzeuge
nach `.werkzeuge` (etwa 350 MB) und lässt sie dort liegen; danach dauert ein
Bau wenige Sekunden. Android Studio oder Gradle werden nicht gebraucht.

Der Schlüssel `schluessel.jks` (Passwort `tspiele123`) bleibt im Ordner: Nur
mit demselben Schlüssel lässt sich eine neue Fassung **über** die installierte
App legen. Wer ihn verliert, muss die App vorher deinstallieren.

## Aufbau

    AndroidManifest.xml          eine Activity, Startmenü-Eintrag, Internet
    src/…/MainActivity.java      öffnet die Adresse und beendet sich
    res/mipmap-*/ic_launcher.png    Symbol für Android bis 7
    res/mipmap-*/ic_launcher_fg.png Symbol-Vordergrund für Android 8+
    res/mipmap-anydpi-v26/…      adaptives Symbol
    werkzeuge/monster_icon.py    malt das Monster in allen Größen
    bauen.sh                     baut und signiert T.Spiele.apk
    T.Spiele.apk                 fertige App zum Installieren

## Grenzen

- Die App ist nicht im Play Store; sie wird als Datei installiert.
- Sie zeigt keinerlei Ladeanzeige: Der Browser übernimmt sofort. Schläft der
  Render-Dienst (Free-Plan, 15 Minuten ohne Besuch), dauert der erste Aufruf
  etwa eine Minute – das passiert dann im Browser, nicht in der App.
