# T.Spiele – Android-App

Ein festlicher Startbildschirm mit zwei Wegen, beide öffnen sich im Browser:

- **Connector öffnen** → `https://trashygecko.onrender.com/connector`
  (Chat und alle Spiele)
- **Meine Homepage** → `https://trashygecko-collab.github.io/Trashygecko/`

Die Adressen stehen bewusst nicht auf dem Schirm, nur die zwei Knöpfe. Der
Hintergrund ist tiefes Grün mit einem hellen Schein in der Mitte, die Schrift
golden und in Serifen (wie eine Einladung), der Abstand zwischen den Knöpfen
34 dp – die Luft darüber und darunter teilt sich im goldenen Schnitt (1 : 1,618).

Zwischen den beiden Knöpfen steht ein Spruch mit dem heutigen Datum und der
laufenden Uhrzeit: „Wenn die Zeiger am Samstag, 15. November 2026 auf 10:42 Uhr
stehen, erwacht unser Spiel zum Leben. Jetzt treffen sich Taktik, Glück und gute
Laune." Die Uhr tickt im Sekundentakt, solange der Startbildschirm offen ist.

Dazu kommt **jeden Morgen um 7:30 ein Lob als Benachrichtigung** – ohne Server,
der Wecker steckt im Gerät (auch nach einem Neustart wird er neu gestellt).
Abschalten geht in den Android-Einstellungen unter Benachrichtigungen.

**Tablets:** Alle Maße – Schrift und Abstände – wachsen mit der kürzeren
Bildschirmseite. Maßstab 1 ist ein 360 dp breites Handy; ein 7-Zoll-Tablet
(rund 600 dp) bekommt Faktor 1,67 (Titel 38 → 63 sp, Knöpfe 19 → 32 sp),
ein 10-Zoll-Tablet den Deckel 1,9 (Titel 72 sp, Knöpfe 36 sp). Im Querformat
gilt derselbe Faktor, weil die *kleinste* Breite zählt. Wird der Inhalt dabei
höher als das Fenster, laesst sich der Startbildschirm rollen statt unten
abgeschnitten zu sein.

Kein eigener Browser. Berechtigungen: Internet, Benachrichtigungen (ab Android 13
zu bestätigen) und „Neustart erkennen“ für den Wecker. Die App bleibt im
Hintergrund stehen, der Zurück-Knopf führt wieder auf den Startbildschirm.

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

**Uhrzeit oder Sprüche ändern:** in `src/de/trashygecko/tspiele/LobWecker.java`
stehen `STUNDE`/`MINUTE` und die Liste `LOBS`.

**Andere Adressen:** in `src/de/trashygecko/tspiele/MainActivity.java` die
Konstanten `CONNECTOR` und `HOMEPAGE` ändern, dann neu bauen. Die Farben
(`GOLD`, `ELFENBEIN`, `GRUEN_*`) und der Abstand (`PHI`, 34 dp) stehen oben
in derselben Datei.

**Anderes Symbol:** `werkzeuge/monster_icon.py` malen lassen
(`python werkzeuge/monster_icon.py`) – es schreibt alle Größen nach `res/`.
Die Formen stehen oben im Skript (Körper, Augen, Wangen, Mund).

**Neu bauen:**

    bash bauen.sh

Beim ersten Mal lädt das Skript ein portables JDK 17 und die Android-Werkzeuge
nach `../../../arbeit/werkzeuge/t-spiele` (etwa 350 MB) und lässt sie dort
liegen – also ausserhalb der Homepage; danach dauert ein
Bau wenige Sekunden. Android Studio oder Gradle werden nicht gebraucht.

Der Schlüssel `schluessel.jks` (Passwort `tspiele123`) bleibt im Ordner: Nur
mit demselben Schlüssel lässt sich eine neue Fassung **über** die installierte
App legen. Wer ihn verliert, muss die App vorher deinstallieren.

## Aufbau

    AndroidManifest.xml          eine Activity, Startmenü-Eintrag, Internet
    src/…/MainActivity.java      Startbildschirm grün/gold, stellt den Wecker
    src/…/LobWecker.java         Morgen-Lob um 7:30 (Wecker + Anzeige)
    src/…/LobStart.java          stellt den Wecker nach Neustart wieder
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
