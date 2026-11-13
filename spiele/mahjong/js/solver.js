
/*
 * TRASHYGECKO MAHJONG
 * Lösungsprüfung
 */

function getBoardKey(board) {
    return board
        .filter(tile => !tile.removed)
        .map(tile => tile.id)
        .sort((a, b) => a - b)
        .join(",");
}

function findSolution(originalBoard) {
    const board = originalBoard.map(tile => ({ ...tile }));
    const failedStates = new Set();

    function solve() {
        const remaining = board.filter(tile => !tile.removed);

        if (remaining.length === 0) {
            return [];
        }

        const key = getBoardKey(board);

        if (failedStates.has(key)) {
            return null;
        }

        const freeTiles = getFreeTiles(board);
        const groups = new Map();

        for (const tile of freeTiles) {
            if (!groups.has(tile.symbolId)) {
                groups.set(tile.symbolId, []);
            }

            groups.get(tile.symbolId).push(tile);
        }

        const pairs = [];

        for (const [symbolId, tiles] of groups) {
            for (let i = 0; i < tiles.length; i++) {
                for (let j = i + 1; j < tiles.length; j++) {
                    pairs.push({
                        first: tiles[i],
                        second: tiles[j],
                        symbolId
                    });
                }
            }
        }

        pairs.sort((a, b) =>
            groups.get(a.symbolId).length -
            groups.get(b.symbolId).length
        );

        for (const pair of pairs) {
            pair.first.removed = true;
            pair.second.removed = true;

            const result = solve();

            if (result !== null) {
                pair.first.removed = false;
                pair.second.removed = false;

                return [
                    [pair.first.id, pair.second.id],
                    ...result
                ];
            }

            pair.first.removed = false;
            pair.second.removed = false;
        }

        failedStates.add(key);

        return null;
    }

    if (board.length !== 144) {
        return null;
    }

    if (board.some(tile => tile.symbolId === null)) {
        return null;
    }

    return solve();
}
