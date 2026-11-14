#!/usr/bin/env bash
# Baut T.Spiele.apk ohne Android Studio und ohne Gradle.
#
#   bash bauen.sh
#
# Beim ersten Mal lädt das Skript ein portables JDK 17 und die Android-Werkzeuge
# (zusammen etwa 350 MB) in den Ordner .werkzeuge und lässt sie dort liegen.
# Danach dauert ein Bau nur noch Sekunden.
#
# Eigenen Werkzeugordner benutzen:  TSPIELE_WERKZEUGE=/pfad bash bauen.sh

set -e

HIER="$(cd "$(dirname "$0")" && pwd)"
WERK="${TSPIELE_WERKZEUGE:-$HIER/.werkzeuge}"
mkdir -p "$WERK"
cd "$HIER"

# ---------------------------------------------------------------- Werkzeuge
if ! ls -d "$WERK"/jdk-17* >/dev/null 2>&1; then
  echo "== JDK 17 laden (einmalig)"
  curl -fL --retry 3 -o "$WERK/jdk.zip" \
    "https://api.adoptium.net/v3/binary/latest/17/ga/windows/x64/jdk/hotspot/normal/eclipse"
  unzip -q -o "$WERK/jdk.zip" -d "$WERK"
fi
JAVA_HOME="$(ls -d "$WERK"/jdk-17* | head -1)"
export JAVA_HOME
PATH="$JAVA_HOME/bin:$PATH"

if [ ! -d "$WERK/sdk/build-tools/34.0.0" ]; then
  if [ ! -d "$WERK/sdk/cmdline-tools/latest" ]; then
    echo "== Android-Werkzeuge laden (einmalig)"
    curl -fL --retry 3 -o "$WERK/cmdline.zip" \
      "https://dl.google.com/android/repository/commandlinetools-win-11076708_latest.zip"
    rm -rf "$WERK/auspack" && mkdir -p "$WERK/auspack"
    unzip -q -o "$WERK/cmdline.zip" -d "$WERK/auspack"
    mkdir -p "$WERK/sdk/cmdline-tools"
    mv "$WERK/auspack/cmdline-tools" "$WERK/sdk/cmdline-tools/latest"
  fi
  SDKM="$WERK/sdk/cmdline-tools/latest/bin/sdkmanager.bat"
  echo "== Lizenzen und SDK-Teile (einmalig)"
  printf 'y\n%.0s' $(seq 1 60) | "$SDKM" --sdk_root="$WERK/sdk" --licenses >/dev/null 2>&1 || true
  printf 'y\n%.0s' $(seq 1 60) | "$SDKM" --sdk_root="$WERK/sdk" \
    "platforms;android-34" "build-tools;34.0.0"
fi

SDK="$WERK/sdk"
ANDROID_JAR="$SDK/platforms/android-34/android.jar"
BT="$SDK/build-tools/34.0.0"
MINDESTENS=14      # Android 4.0 – damit läuft es auf praktisch jedem Gerät
ZIEL=34

BAU="$HIER/.bau"
rm -rf "$BAU"
mkdir -p "$BAU/klassen"

# ---------------------------------------------------------------- Bauen
echo "== Ressourcen"
"$BT/aapt2.exe" compile --dir res -o "$BAU/res.zip"
"$BT/aapt2.exe" link -o "$BAU/app-ohne-signatur.apk" \
  -I "$ANDROID_JAR" \
  --manifest AndroidManifest.xml \
  --java "$BAU/gen" \
  --min-sdk-version "$MINDESTENS" \
  --target-sdk-version "$ZIEL" \
  --version-code 1 --version-name 1.0 \
  "$BAU/res.zip"

echo "== Java"
javac -source 8 -target 8 -nowarn -bootclasspath "$ANDROID_JAR" \
  -d "$BAU/klassen" src/de/trashygecko/tspiele/MainActivity.java

echo "== Dex"
"$BT/d8.bat" --lib "$ANDROID_JAR" --min-api "$MINDESTENS" \
  --output "$BAU" $(find "$BAU/klassen" -name '*.class')

python - "$BAU" <<'PY'
import shutil, sys, zipfile
bau = sys.argv[1]
shutil.copyfile(bau + "/app-ohne-signatur.apk", bau + "/app-mit-dex.apk")
with zipfile.ZipFile(bau + "/app-mit-dex.apk", "a", zipfile.ZIP_DEFLATED) as apk:
    apk.write(bau + "/classes.dex", "classes.dex")
PY

echo "== Ausrichten und signieren"
"$BT/zipalign.exe" -f 4 "$BAU/app-mit-dex.apk" "$BAU/app-ausgerichtet.apk"
if [ ! -f schluessel.jks ]; then
  echo "== Eigenen Schlüssel anlegen (einmalig, bleibt im Projekt)"
  keytool -genkeypair -keystore schluessel.jks -storepass tspiele123 \
    -keypass tspiele123 -alias tspiele -keyalg RSA -keysize 2048 \
    -validity 10000 -dname "CN=T.Spiele, O=TrashyGecko, C=DE" >/dev/null 2>&1
fi
"$BT/apksigner.bat" sign --ks schluessel.jks --ks-pass pass:tspiele123 \
  --key-pass pass:tspiele123 --min-sdk-version "$MINDESTENS" \
  --out "$HIER/T.Spiele.apk" "$BAU/app-ausgerichtet.apk"

echo "== Prüfen"
"$BT/apksigner.bat" verify --min-sdk-version "$MINDESTENS" --print-certs "$HIER/T.Spiele.apk" | head -4
"$BT/aapt2.exe" dump badging "$HIER/T.Spiele.apk" | head -4

# Die Kachel "Meine Android App" auf spiele.html holt die App aus
# spiele/android – dorthin kommt die neue Fassung gleich mit.
ZIEL="$HIER/../../spiele/android"
if [ -d "$ZIEL" ]; then
  cp "$HIER/T.Spiele.apk" "$ZIEL/tspiele.apk"
  echo "Kopie für die Seite: $(cd "$ZIEL" && pwd)/tspiele.apk"
else
  echo "Hinweis: Ordner $ZIEL fehlt – Kopie für die Seite übersprungen."
fi

echo
echo "Fertig: $HIER/T.Spiele.apk"
