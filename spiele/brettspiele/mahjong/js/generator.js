
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
            ((originalBoard[0] && originalBoard[0].level) || "?") +
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

    // Beweis der Loesbarkeit: diese Reihenfolge geht mit diesen Motiven auf.
    board.loesung = removalOrder;

    return board;
}

// --------------------------------------------------
// SCHWIERIGER MODUS: VIER KOPIEN JE MOTIV
// --------------------------------------------------

// Ein klassischer Mahjong-Satz hat 144 Steine aus 36 Motiven zu je vier
// Exemplaren.  Genau daraus entsteht die Schwierigkeit: Welche zwei Steine ein
// Paar bilden, entscheidet der Spieler, und eine unglueckliche Wahl kann zwei
// Steine uebrig lassen, die nie gleichzeitig frei werden.  Die Motive selbst
// bleiben dieselben - je Brett kommen 36 der vorhandenen 72 zum Einsatz.
const HARTE_MOTIVE = 36;

function createHardSymbolSet(symbolIds) {
    if (!Array.isArray(symbolIds) || symbolIds.length < HARTE_MOTIVE) {
        throw new Error(`Es werden mindestens ${HARTE_MOTIVE} Motive benoetigt.`);
    }

    if (new Set(symbolIds).size < HARTE_MOTIVE) {
        throw new Error("Die Motive muessen unterschiedlich sein.");
    }

    return shuffle(symbolIds).slice(0, HARTE_MOTIVE);
}

// Gibt es einen Zug?  Nur fuer den Startschutz des schweren Aufbaus.
function hatFreiesPaar(board) {
    const frei = getFreeTiles(board);
    const gesehen = new Set();

    for (const tile of frei) {
        if (gesehen.has(tile.symbolId)) return true;

        gesehen.add(tile.symbolId);
    }

    return false;
}

// Anspruchsvoller Aufbau: 36 Motive zu je vier Steinen.  Jedes Motiv liegt auf
// zwei Abbaustufen, also weit auseinander - die vier Steine muessen ueber
// mehrere Schichten hinweg zusammengefuehrt werden.  Weil jede Stufe fuer sich
// ein passendes Paar bildet, bleibt das Brett entlang dieser Reihenfolge
// vollstaendig abbaubar; die Loesung liegt als Beweis am Brett (board.loesung).
function generateHardBoard(symbolIds, level = 1) {
    const board = createBoard(level);

    if (board.length !== HARTE_MOTIVE * 4) {
        throw new Error(
            `Level ${level} enthaelt ${board.length} statt 144 Positionen.`
        );
    }

    const order = findRemovalOrder(board);

    if (order.length !== HARTE_MOTIVE * 2) {
        throw new Error("Die Abbaureihenfolge ist unvollstaendig.");
    }

    const nachId = new Map(board.map(tile => [tile.id, tile]));

    for (let versuch = 0; versuch < 5; versuch++) {
        const motive = createHardSymbolSet(symbolIds);
        const stufen = shuffle(order.map((_, index) => index));

        motive.forEach((motiv, index) => {
            for (const stufe of [stufen[index * 2], stufen[index * 2 + 1]]) {
                for (const id of order[stufe]) {
                    const tile = nachId.get(id);

                    if (!tile) throw new Error("Ein Spielstein fehlt.");

                    tile.symbolId = motiv;
                }
            }
        });

        // Kein Brett ausliefern, auf dem von Anfang an kein Zug moeglich ist.
        if (hatFreiesPaar(board)) break;
    }

    board.loesung = order;

    return board;
}

// Prueft eine Abbaureihenfolge mit den Regeln des Spiels: jedes Paar muss aus
// zwei freien Steinen mit demselben Motiv bestehen, am Ende muss das Brett leer
// sein.  Das ist der Beweis, dass die Verteilung loesbar ist.
function pruefeAbbau(board, order) {
    const kopie = board.map(tile => ({ ...tile }));

    for (const paar of order) {
        if (!removeTilePair(kopie, paar[0], paar[1])) return false;
    }

    return kopie.every(tile => tile.removed);
}
