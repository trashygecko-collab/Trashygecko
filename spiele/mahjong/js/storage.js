
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
            hints: true
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

        return {
            ...createDefaultStorage(),
            ...data
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
function saveLevelProgress(data, level) {
    data.currentLevel = level;

    data.unlockedLevel = Math.max(
        data.unlockedLevel,
        level
    );

    return saveStorage(data);
}

// Speichert einen abgeschlossenen Level
function saveLevelResult(data, level, seconds) {
    data.gamesPlayed++;
    data.gamesWon++;

    const previousBest = data.bestTimes[level];

    if (
        previousBest === undefined ||
        seconds < previousBest
    ) {
        data.bestTimes[level] = seconds;
    }

    data.currentLevel = level + 1;

    data.unlockedLevel = Math.max(
        data.unlockedLevel,
        level + 1
    );

    return saveStorage(data);
}

// Speichert Einstellungen
function saveSettings(data, settings) {
    data.settings = {
        ...data.settings,
        ...settings
    };

    return saveStorage(data);
}

// Setzt alle Spieldaten zurück
function resetStorage() {
    const freshData = createDefaultStorage();

    saveStorage(freshData);

    return freshData;
}
