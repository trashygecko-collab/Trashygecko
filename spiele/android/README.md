# TMahjong – Android-App

Die App zeigt das Mahjong-Spiel aus `../brettspiele/mahjong` in einer randlosen
WebView. Es gibt **keine zweite Kopie des Spiels**: Gradle kopiert beim Bauen
`../brettspiele/mahjong` nach `app/src/main/assets/www/brettspiele/mahjong` und
`../lob.js` nach `app/src/main/assets/www/`. Der Ordnerbaum im APK ist also
derselbe wie auf der Seite – deshalb findet `mahjong.html` seine Spruchliste als
`../../lob.js`. Änderungen am Spiel landen damit automatisch in der App – nur
neu bauen.

## Was die App vom Browser unterscheidet

| | Browser | App |
|---|---|---|
| Vollbild-Knopf | entfernt (das Spiel läuft im Browser wie in der App randlos) | nicht nötig |
| „SPIEL VERLASSEN“ in den Einstellungen | versteckt | sichtbar, beendet die App |
| Ausrichtung | frei | Querformat (`sensorLandscape`) |
| Systemleisten | je nach Browser | dauerhaft ausgeblendet |

Der Knopf erscheint nur, wenn die native Brücke da ist:
`window.TMahjong.verlassen()` → `MainActivity.Bruecke.verlassen()` →
`finishAffinity()`.

## Bauen

**Ein Befehl (nutzt die vorhandenen portablen Werkzeuge)**

```bash
bash bauen.sh            # baut app/build/outputs/apk/debug/app-debug.apk
bash bauen.sh install    # baut und installiert auf einem angeschlossenen Gerät
```

Ohne USB-Debugging lässt sich die APK auch per Datei-Modus aufs Handy kopieren:
`powershell -ExecutionPolicy Bypass -File apk-aufs-handy.ps1` legt sie im
Download-Ordner des angeschlossenen Handys ab.

`bauen.sh` nimmt das portable JDK 17 und das Android-SDK aus
`../../../arbeit/werkzeuge/t-spiele` (dort bauen auch die T.Spiele-APKs), nimmt
Gradle 8.7 aus `../../../arbeit/werkzeuge/gradle-8.7` und lädt es nur nach, wenn es
dort fehlt. Beides liegt ausserhalb der Homepage, damit sie schlank bleibt. Dann
`clean assembleDebug`. Eigene Pfade gehen per
`TMAHJONG_JDK=… TMAHJONG_SDK=… bash bauen.sh`.

**Mit Android Studio**

1. Android Studio öffnen → *Open* → diesen Ordner `android` wählen.
2. Warten, bis Gradle durchgelaufen ist → *Run*.
3. Fertige APK: `app/build/outputs/apk/debug/app-debug.apk`.

Das Projekt nutzt AGP 8.5.2 / Gradle 8.7, `compileSdk 34`, `minSdk 24`. Die
Debug-APK ist mit dem Standard-Debug-Zertifikat signiert und lässt sich per
`adb install -r app-debug.apk` oder direkt am Gerät installieren.

Für eine signierte Release-APK braucht es einen Keystore
(`keytool -genkey …`) und einen `signingConfig`-Block in `app/build.gradle`;
das ist hier bewusst nicht hinterlegt.

## Icons

`icon/icon-quelle.png` ist die gelieferte Vorlage. `python icon-bauen.py`
erzeugt daraus:

- `mipmap-*/ic_launcher.png` – klassisch, quadratisch (48 … 192 px)
- `mipmap-*/ic_launcher_round.png` – rund geschnitten
- `mipmap-*/ic_launcher_foreground.png` – Vordergrund für das adaptive Icon
  (Android 8+, Hintergrund `#0B241D` aus `values/colors.xml`)
- `icon/ic_launcher-playstore.png` – 512 × 512 für den Play Store

## Dateien

```
android/
  bauen.sh                       # ein Befehl: JDK/SDK suchen, Gradle holen, bauen
  .gitignore                     # app/build/, assets/www/ bleiben lokal
  build.gradle, settings.gradle, gradle.properties
  app/
    build.gradle                 # inkl. Kopier-Tasks für die Spieldateien
    src/main/
      AndroidManifest.xml
      java/de/trashygecko/tmahjong/MainActivity.java
      res/values/{strings,colors,themes}.xml
      res/mipmap-*/…             # Icons
      assets/www/…               # entsteht beim Bauen (Kopie von
                                 # ../brettspiele/mahjong und ../lob.js)
  icon-bauen.py, icon/
```
