#!/usr/bin/env bash
# Baut die Debug-APK von TMahjong ohne Android Studio.
#
#   bash bauen.sh            # baut app/build/outputs/apk/debug/app-debug.apk
#   bash bauen.sh install    # baut und installiert auf einem angeschlossenen Geraet
#
# Werkzeuge: JDK 17, Android-SDK und Gradle liegen ausserhalb der Homepage
# in ../../../arbeit/werkzeuge (t-spiele/ und gradle-8.7/), damit das Projekt
# schlank bleibt. Gradle wird nur geladen, wenn es dort fehlt.
#
# Eigene Pfade:  TMAHJONG_JDK=... TMAHJONG_SDK=... bash bauen.sh

set -e

HIER="$(cd "$(dirname "$0")" && pwd)"
ARBEIT="$(cd "$HIER/../../.." && pwd)/arbeit/werkzeuge"
WERK="$ARBEIT"
SPIELE_WERKZEUG="$ARBEIT/t-spiele"

cd "$HIER"

# ------------------------------------------------------------------ JDK 17
if [ -n "${TMAHJONG_JDK:-}" ]; then
    export JAVA_HOME="$TMAHJONG_JDK"
elif ls -d "$SPIELE_WERKZEUG"/jdk-17* >/dev/null 2>&1; then
    export JAVA_HOME="$(ls -d "$SPIELE_WERKZEUG"/jdk-17* | head -1)"
elif [ -d "$WERK/jdk-17" ]; then
    export JAVA_HOME="$WERK/jdk-17"
else
    echo "Kein JDK 17 gefunden. Setze TMAHJONG_JDK=/pfad/zum/jdk-17" >&2
    exit 1
fi

# ---------------------------------------------------------------- Android SDK
if [ -n "${TMAHJONG_SDK:-}" ]; then
    export ANDROID_HOME="$TMAHJONG_SDK"
elif [ -d "$SPIELE_WERKZEUG/sdk/platforms/android-34" ]; then
    export ANDROID_HOME="$SPIELE_WERKZEUG/sdk"
elif [ -d "$WERK/sdk/platforms/android-34" ]; then
    export ANDROID_HOME="$WERK/sdk"
else
    echo "Kein Android-SDK mit platforms/android-34 gefunden." >&2
    exit 1
fi

export ANDROID_SDK_ROOT="$ANDROID_HOME"
echo "JDK:   $JAVA_HOME"
echo "SDK:   $ANDROID_HOME"

# ------------------------------------------------------------------- Gradle
if [ ! -x "$WERK/gradle-8.7/bin/gradle" ]; then
    mkdir -p "$WERK"
    echo "Gradle 8.7 wird geladen (einmalig, ca. 128 MB) ..."
    curl -L -o "$WERK/gradle-8.7-bin.zip" https://services.gradle.org/distributions/gradle-8.7-bin.zip
    python -c "import zipfile; zipfile.ZipFile('$WERK/gradle-8.7-bin.zip').extractall('$WERK')"
    rm -f "$WERK/gradle-8.7-bin.zip"
fi

# --------------------------------------------------------------------- Bau
"$WERK/gradle-8.7/bin/gradle" --no-daemon --console=plain clean assembleDebug

APK="$HIER/app/build/outputs/apk/debug/app-debug.apk"

if [ ! -f "$APK" ]; then
    echo "Kein APK gebaut." >&2
    exit 1
fi

echo
echo "Fertig: $APK"

# ---------------------------------------------------------------- Installieren
if [ "${1:-}" = "install" ]; then
    "$ANDROID_HOME/platform-tools/adb" install -r "$APK"
fi
