/*
 * TRASHYGECKO MAHJONG
 * Spielstände und Einstellungen
 */

const STORAGE_KEY = "trashygecko_mahjong_save";

// Ein Formensatz in einer Schwierigkeit: eigener Fortschritt, eigene
// Bestzeiten.  Leicht und Schwer haben verschiedene Bretter - deshalb werden
// Level, Haken und Zeiten getrennt gezaehlt.
function neueSatzDaten() {
    return {
        currentLevel: 1,
        unlockedLevel: 1,
        gamesPlayed: 0,
        gamesWon: 0,
        bestTimes: {}
    };
}

// Alle vier Abschnitte: Satz 1 (PC) und Satz 2 (Handy), je leicht und schwer.
function neuerFortschritt() {
    return {
        1: { leicht: neueSatzDaten(), schwer: neueSatzDaten() },
        2: { leicht: neueSatzDaten(), schwer: neueSatzDaten() }
    };
}

// Standardwerte für einen neuen Spieler
function createDefaultStorage() {
    return {
        fortschritt: neuerFortschritt(),
        settings: {
            sound: true,
            animations: true,
            hints: true,
            steinForm: "rund",
            steinZustand: "neu",
            hintergrund: "gruen"
        }
    };
}

// Zieht die Freischaltung nach statt sie zurueckzusetzen: Wer frueher frei ein
// Level waehlen konnte, eine Bestzeit hat oder dort zuletzt gespielt wurde,
// behaelt den Zugang.  Gesperrt wird nur, was nie erreicht wurde.
function freischalten(eintrag) {
    const geschafft = Object.keys(eintrag.bestTimes || {})
        .map(Number)
        .filter(level => Number.isFinite(level) && level > 0);

    eintrag.unlockedLevel = Math.max(
        Number(eintrag.unlockedLevel) || 1,
        Number(eintrag.currentLevel) || 1,
        ...geschafft.map(level => level + 1)
    );

    return eintrag;
}

// Nimmt einen gespeicherten Abschnitt und fuellt fehlende Felder auf.
function satzAusEintrag(eintrag) {
    const basis = neueSatzDaten();

    if (!eintrag || typeof eintrag !== "object") return basis;

    const zahl = (wert, ersatz) => (Number.isFinite(wert) ? wert : ersatz);

    const daten = {
        currentLevel: zahl(eintrag.currentLevel, basis.currentLevel),
        unlockedLevel: zahl(eintrag.unlockedLevel, basis.unlockedLevel),
        gamesPlayed: zahl(eintrag.gamesPlayed, 0),
        gamesWon: zahl(eintrag.gamesWon, 0),
        bestTimes: { ...(eintrag.bestTimes || {}) }
    };

    return freischalten(daten);
}

// Bringt jeden Spielstand in die Form mit vier Abschnitten.
//
// Alte Staende (bis v64) hatten einen gemeinsamen Fortschritt fuer Leicht und
// Schwer: Satz 1 lag ganz oben, Satz 2 in "satz2".  Diese Daten wandern in den
// Abschnitt der zuletzt gespielten Schwierigkeit - dort stammen die Zeiten
// her; der andere Abschnitt beginnt leer.  Verworfen wird nichts:
// Einstellungen, Punkte und alle Zeiten bleiben erhalten.
function normalisiereSpielstand(daten) {
    if (!daten || typeof daten !== "object") return createDefaultStorage();

    const fortschritt = neuerFortschritt();
    const alt = !daten.fortschritt;
    const stufe = daten.settings && daten.settings.schwierigkeit === "schwer"
        ? "schwer"
        : "leicht";

    [1, 2].forEach(satz => {
        ["leicht", "schwer"].forEach(eigeneStufe => {
            const quelle = alt
                ? (satz === 1 ? daten : daten.satz2)
                : (daten.fortschritt[satz] || {})[eigeneStufe];

            // Beim alten Stand gibt es nur einen Abschnitt - er gehoert der
            // zuletzt gespielten Schwierigkeit.
            if (alt && eigeneStufe !== stufe) return;

            if (quelle) fortschritt[satz][eigeneStufe] = satzAusEintrag(quelle);
        });
    });

    if (!alt) return { ...daten, fortschritt };

    const {
        currentLevel, unlockedLevel, gamesPlayed, gamesWon, bestTimes, satz2, ...rest
    } = daten;

    return { ...rest, fortschritt };
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

        return normalisiereSpielstand({
            ...standard,
            ...data,
            // Bewusst der gespeicherte Wert: fehlt er, ist es ein alter Stand
            // und wird migriert - der Standardwert darf das nicht verdecken.
            fortschritt: data.fortschritt,
            // Neue Einstellungen fehlen in alten Spielstaenden - ergaenzen.
            settings: {
                ...standard.settings,
                ...(data.settings || {})
            }
        });

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

// Speichert einen abgeschlossenen Level. Mit "ziel" landet der Fortschritt in
// einem eigenen Abschnitt (Formensatz und Schwierigkeit), gespeichert wird
// immer alles.
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
    freischalten(eintrag);

    return saveStorage(data);
}

// Speichert Einstellungen

// Setzt alle Spieldaten zurück
