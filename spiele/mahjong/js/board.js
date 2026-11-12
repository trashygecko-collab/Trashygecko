
/*
 * TRASHYGECKO MAHJONG
 * Spielfeldverwaltung für mehrere Level
 */

// --------------------------------------------------
// SPIELBRETT ERSTELLEN
// --------------------------------------------------

function createBoard(level = 1) {
    const layout = BOARD_LAYOUTS[level] || BOARD_LAYOUT;

    return layout.positions.map((position, index) => ({
        id: index,
        x: position.x,
        y: position.y,
        z: position.z,
        symbolId: null,
        removed: false
    }));
}

// --------------------------------------------------
// STEIN VORHANDEN?
// --------------------------------------------------

function isPresent(tile, board) {
    return board.some(other =>
        other.id !== tile.id &&
        !other.removed &&
        other.x === tile.x &&
        other.y === tile.y &&
        other.z > tile.z
    );
}

// --------------------------------------------------
// STEIN VON OBEN BLOCKIERT?
// --------------------------------------------------

function isBlockedFromAbove(tile, board) {
    return board.some(other =>
        !other.removed &&
        other.z > tile.z &&
        other.x === tile.x &&
        other.y === tile.y
    );
}

// --------------------------------------------------
// BEIDE SEITEN BLOCKIERT?
// --------------------------------------------------

function isBlockedOnBothSides(tile, board) {
    const leftBlocked = board.some(other =>
        !other.removed &&
        other.z === tile.z &&
        other.y === tile.y &&
        other.x === tile.x - 1
    );

    const rightBlocked = board.some(other =>
        !other.removed &&
        other.z === tile.z &&
        other.y === tile.y &&
        other.x === tile.x + 1
    );

    return leftBlocked && rightBlocked;
}

// --------------------------------------------------
// IST EIN STEIN FREI?
// --------------------------------------------------

function isTileFree(tile, board) {
    if (!tile || tile.removed) return false;

    if (isBlockedFromAbove(tile, board)) {
        return false;
    }

    if (isBlockedOnBothSides(tile, board)) {
        return false;
    }

    return true;
}

// --------------------------------------------------
// ALLE FREIEN STEINE
// --------------------------------------------------

function getFreeTiles(board) {
    return board.filter(tile =>
        !tile.removed && isTileFree(tile, board)
    );
}

// --------------------------------------------------
// STEINZUSTÄNDE
// --------------------------------------------------

function getTileStates(board) {
    return board.map(tile => ({
        ...tile,
        free: isTileFree(tile, board)
    }));
}

// --------------------------------------------------
// PAAR ENTFERNEN
// --------------------------------------------------

function removeTilePair(board, firstId, secondId) {
    const first = board.find(tile => tile.id === firstId);
    const second = board.find(tile => tile.id === secondId);

    if (!first || !second) return false;

    if (first.removed || second.removed) return false;

    if (first.symbolId !== second.symbolId) return false;

    if (!isTileFree(first, board)) return false;
    if (!isTileFree(second, board)) return false;

    first.removed = true;
    second.removed = true;

    return true;
}
