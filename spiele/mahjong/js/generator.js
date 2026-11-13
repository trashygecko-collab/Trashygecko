
/*
 * TRASHYGECKO MAHJONG
 * Generator für mehrere Level
 */

// --------------------------------------------------
// ZUFALLSMISCHER
// --------------------------------------------------

function shuffle(array) {
    const result = [...array];

    for (let i = result.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));

        [result[i], result[j]] = [result[j], result[i]];
    }

    return result;
}

// --------------------------------------------------
// 72 MOTIVE
// --------------------------------------------------

function createSymbolSet(symbolIds) {
    if (!Array.isArray(symbolIds) || symbolIds.length !== 72) {
        throw new Error("Es werden genau 72 Motive benötigt.");
    }

    if (new Set(symbolIds).size !== 72) {
        throw new Error("Die 72 Motive müssen unterschiedlich sein.");
    }

    return shuffle(symbolIds);
}

// --------------------------------------------------
// SPIELBRETT ERSTELLEN
// --------------------------------------------------

function generateBoard(symbolIds, level = 1) {
    const symbols = createSymbolSet(symbolIds);
    const board = createBoard(level);

    const pairs = [];

    symbols.forEach(symbolId => {
        pairs.push(symbolId, symbolId);
    });

    const mixedPairs = shuffle(pairs);

    board.forEach((tile, index) => {
        tile.symbolId = mixedPairs[index];
    });

    return board;
}

// --------------------------------------------------
// LÖSUNGSREIHENFOLGE SUCHEN
// --------------------------------------------------

function findRemovalOrder(originalBoard) {
    // Bereits entfernte Steine bleiben entfernt - damit lässt sich auch eine
    // Abbaureihenfolge für den Rest eines angefangenen Bretts finden.
    const board = originalBoard.map(tile => ({
        ...tile,
        removed: Boolean(tile.removed)
    }));

    const failedStates = new Set();
    let examinedStates = 0;
    const MAX_STATES = 500000;

    function getStateKey() {
        return board
            .filter(tile => !tile.removed)
            .map(tile => tile.id)
            .join(",");
    }

    function search(remainingCount) {
        if (remainingCount === 0) return [];

        examinedStates++;

        if (examinedStates > MAX_STATES) return null;

        const key = getStateKey();

        if (failedStates.has(key)) return null;

        const freeTiles = getFreeTiles(board);

        if (freeTiles.length < 2) {
            failedStates.add(key);
            return null;
        }

        const candidates = [];

        for (let i = 0; i < freeTiles.length; i++) {
            for (let j = i + 1; j < freeTiles.length; j++) {
                candidates.push([
                    freeTiles[i],
                    freeTiles[j]
                ]);
            }
        }

        for (const [first, second] of shuffle(candidates)) {
            first.removed = true;
            second.removed = true;

            const result = search(remainingCount - 2);

            if (result !== null) {
                return [
                    [first.id, second.id],
                    ...result
                ];
            }

            first.removed = false;
            second.removed = false;

            if (examinedStates > MAX_STATES) break;
        }

        failedStates.add(key);
        return null;
    }

    const solution = search(board.filter(tile => !tile.removed).length);

    if (solution === null) {
        throw new Error(
            "Keine vollständige Abbaureihenfolge für Level " +
            (originalBoard[0]?.level || "?") +
            " gefunden."
        );
    }

    return solution;
}

// --------------------------------------------------
// LÖSBARES SPIELBRETT ERZEUGEN
// --------------------------------------------------

function generateSolvableBoard(symbolIds, level = 1) {
    const symbols = createSymbolSet(symbolIds);
    const board = createBoard(level);

    if (board.length !== 144) {
        throw new Error(
            `Level ${level} enthält ${board.length} statt 144 Positionen.`
        );
    }

    const removalOrder = findRemovalOrder(board);

    if (removalOrder.length !== 72) {
        throw new Error("Die Abbaureihenfolge ist unvollständig.");
    }

    removalOrder.forEach((pair, index) => {
        const symbolId = symbols[index];

        const first = board.find(tile => tile.id === pair[0]);
        const second = board.find(tile => tile.id === pair[1]);

        if (!first || !second) {
            throw new Error("Ein Spielstein fehlt.");
        }

        first.symbolId = symbolId;
        second.symbolId = symbolId;
    });

    return board;
}
