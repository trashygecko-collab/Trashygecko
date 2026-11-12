
/*
 * TRASHYGECKO MAHJONG
 * Spielfeldanordnungen
 */

// --------------------------------------------------
// HILFSFUNKTIONEN
// --------------------------------------------------

function createRow(y, start, count, z) {
    return Array.from({ length: count }, (_, i) => ({
        x: start + i,
        y,
        z
    }));
}

function createRanges(y, ranges, z) {
    const positions = [];

    ranges.forEach(([start, end]) => {
        for (let x = start; x <= end; x++) {
            positions.push({ x, y, z });
        }
    });

    return positions;
}

// --------------------------------------------------
// LEVEL 01 – GRUNDFORM
// --------------------------------------------------

const base = [
    createRow(0, 3, 6, 0),
    createRow(1, 2, 8, 0),
    createRow(2, 1, 10, 0),
    createRow(3, 0, 12, 0),
    createRow(4, 0, 12, 0),
    createRow(5, 1, 10, 0),
    createRow(6, 2, 8, 0),
    createRow(7, 3, 6, 0)
].flat();

const layer2 = [
    createRow(1, 4, 4, 1),
    createRow(2, 4, 4, 1),
    createRow(3, 4, 4, 1),
    createRow(4, 4, 4, 1),
    createRow(5, 4, 4, 1),
    createRow(6, 4, 4, 1)
].flat();

const layer3 = [
    createRow(2, 4, 4, 2),
    createRow(3, 4, 4, 2),
    createRow(4, 4, 4, 2)
].flat();

const layer4 = [
    createRow(2, 5, 2, 3),
    createRow(3, 5, 2, 3),
    createRow(4, 5, 2, 3)
].flat();

const BOARD_LAYOUT = {
    id: "trashygecko-mahjong-114",
    level: 1,
    columns: 12,
    rows: 8,
    tileCount: 114,
    layers: [base, layer2, layer3, layer4],
    positions: [...base, ...layer2, ...layer3, ...layer4]
};

// --------------------------------------------------
// LEVEL 02 – HERZFORM
// --------------------------------------------------

const heartBase = [
    createRanges(0, [[2, 3], [8, 9]], 0),
    createRanges(1, [[1, 4], [7, 10]], 0),
    createRanges(2, [[0, 11]], 0),
    createRanges(3, [[0, 11]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[2, 9]], 0),
    createRanges(7, [[3, 8]], 0),
    createRanges(8, [[5, 6]], 0)
].flat();

const heartLayer2 = [
    createRow(2, 4, 4, 1),
    createRow(3, 4, 4, 1),
    createRow(4, 4, 4, 1),
    createRow(5, 4, 4, 1),
    createRow(6, 4, 4, 1),
    createRow(7, 4, 4, 1)
].flat();

const heartLayer3 = [
    createRow(3, 4, 4, 2),
    createRow(4, 4, 4, 2),
    createRow(5, 4, 4, 2)
].flat();

const heartLayer4 = [
    createRow(3, 5, 2, 3),
    createRow(4, 5, 2, 3),
    createRow(5, 5, 2, 3)
].flat();

const HEART_LAYOUT = {
    id: "trashygecko-mahjong-heart",
    level: 2,
    columns: 12,
    rows: 9,
    tileCount: 114,
    layers: [
        heartBase,
        heartLayer2,
        heartLayer3,
        heartLayer4
    ],
    positions: [
        ...heartBase,
        ...heartLayer2,
        ...heartLayer3,
        ...heartLayer4
    ]
};

// --------------------------------------------------
// LEVEL 03 – JESUSKREUZ
// --------------------------------------------------

const crossBase = [
    // Oberer senkrechter Balken
    createRanges(0, [[4, 7]], 0),
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[4, 7]], 0),
    createRanges(3, [[4, 7]], 0),

    // Breiter Querbalken
    createRanges(4, [[0, 11]], 0),
    createRanges(5, [[0, 11]], 0),
    createRanges(6, [[0, 11]], 0),

    // Unterer senkrechter Balken
    createRanges(7, [[4, 7]], 0),
    createRanges(8, [[4, 7]], 0),
    createRanges(9, [[4, 7]], 0),
    createRanges(10, [[4, 7]], 0),
    createRanges(11, [[4, 7]], 0)
].flat();

// Zweite Ebene: 24 Steine
const crossLayer2 = [
    createRow(3, 4, 4, 1),
    createRow(4, 4, 4, 1),
    createRow(5, 4, 4, 1),
    createRow(6, 4, 4, 1),
    createRow(7, 4, 4, 1),
    createRow(8, 4, 4, 1)
].flat();

// Dritte Ebene: 12 Steine
const crossLayer3 = [
    createRow(4, 4, 4, 2),
    createRow(5, 4, 4, 2),
    createRow(6, 4, 4, 2)
].flat();

// Vierte Ebene: 6 Steine
const crossLayer4 = [
    createRow(4, 5, 2, 3),
    createRow(5, 5, 2, 3),
    createRow(6, 5, 2, 3)
].flat();

const CROSS_LAYOUT = {
    id: "trashygecko-mahjong-cross",
    level: 3,
    columns: 12,
    rows: 12,
    tileCount: 114,
    layers: [
        crossBase,
        crossLayer2,
        crossLayer3,
        crossLayer4
    ],
    positions: [
        ...crossBase,
        ...crossLayer2,
        ...crossLayer3,
        ...crossLayer4
    ]
};

// --------------------------------------------------
// LEVELÜBERSICHT
// --------------------------------------------------

const BOARD_LAYOUTS = {
    1: BOARD_LAYOUT,
    2: HEART_LAYOUT,
    3: CROSS_LAYOUT
};
