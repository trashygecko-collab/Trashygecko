
/*
 * TRASHYGECKO MAHJONG
 * Spielstände und Einstellungen
 */

const STORAGE_KEY = "trashygecko_mahjong_save";

// Standardwerte für einen neuen Spieler
function createDefaultStorage() {
    return {
        currentLevel: 1,
        unlockedLevel: 1,
        gamesPlayed: 0,
        gamesWon: 0,
        bestTimes: {},
        settings: {
            sound: true,
            animations: true,
            hints: true,
            steinForm: "rund",
            steinZustand: "neu",
            formenSatz: 1,
            hintergrund: "gruen"
        },
        // Zweiter Formensatz: eigener Fortschritt, eigene Bestzeiten.
        satz2: {
            currentLevel: 1,
            unlockedLevel: 1,
            gamesPlayed: 0,
            gamesWon: 0,
            bestTimes: {}
        }
    };
}

// Lädt gespeicherte Daten
function loadStorage() {
    try {
        const saved = localStorage.getItem(STORAGE_KEY);

        if (!saved) {
            return createDefaultStorage();
        }

        const data = JSON.parse(saved);
        const standard = createDefaultStorage();

        return {
            ...standard,
            ...data,
            // Neue Einstellungen fehlen in alten Spielstaenden - ergaenzen.
            settings: {
                ...standard.settings,
                ...(data.settings || {})
            }
        };

    } catch (error) {
        console.error("Spielstand konnte nicht geladen werden:", error);
        return createDefaultStorage();
    }
}

// Speichert alle Daten
function saveStorage(data) {
    try {
        localStorage.setItem(
            STORAGE_KEY,
            JSON.stringify(data)
        );

        return true;

    } catch (error) {
        console.error("Spielstand konnte nicht gespeichert werden:", error);
        return false;
    }
}

// Speichert den erreichten Level

// Speichert einen abgeschlossenen Level. Mit "ziel" landet der Fortschritt in
// einem eigenen Abschnitt (zweiter Formensatz), gespeichert wird immer alles.
function saveLevelResult(data, level, seconds, ziel) {
    const eintrag = ziel || data;

    eintrag.gamesPlayed = (eintrag.gamesPlayed || 0) + 1;
    eintrag.gamesWon = (eintrag.gamesWon || 0) + 1;

    eintrag.bestTimes = eintrag.bestTimes || {};

    const previousBest = eintrag.bestTimes[level];

    if (
        previousBest === undefined ||
        seconds < previousBest
    ) {
        eintrag.bestTimes[level] = seconds;
    }

    eintrag.currentLevel = level + 1;

    eintrag.unlockedLevel = Math.max(
        eintrag.unlockedLevel || 1,
        level + 1
    );

    return saveStorage(data);
}

// Speichert Einstellungen

// Setzt alle Spieldaten zurück
