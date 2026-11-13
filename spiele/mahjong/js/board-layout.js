
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
// LEVEL 04 – TEDDYBÄR
// --------------------------------------------------

const teddyBase = [
    createRanges(0, [[1, 2], [9, 10]], 0),   // Ohren
    createRanges(1, [[1, 4], [7, 10]], 0),   // Ohren und Kopf oben
    createRanges(2, [[2, 9]], 0),            // Kopf
    createRanges(3, [[2, 9]], 0),            // Kopf
    createRanges(4, [[3, 8]], 0),            // Kinn
    createRanges(5, [[0, 11]], 0),           // ausgestreckte Arme
    createRanges(6, [[1, 10]], 0),           // Bauch
    createRanges(7, [[2, 9]], 0),            // Bauch
    createRanges(8, [[1, 4], [7, 10]], 0)    // Beine
].flat();

const teddyLayer2 = [
    createRow(2, 4, 4, 1),
    createRow(3, 4, 4, 1),
    createRow(4, 4, 4, 1),
    createRow(5, 4, 4, 1),
    createRow(6, 4, 4, 1),
    createRow(7, 4, 4, 1)
].flat();

const teddyLayer3 = [
    createRow(3, 4, 4, 2),
    createRow(4, 4, 4, 2),
    createRow(5, 4, 4, 2)
].flat();

const teddyLayer4 = [
    createRow(3, 5, 2, 3),
    createRow(4, 5, 2, 3),
    createRow(5, 5, 2, 3)
].flat();

const TEDDY_LAYOUT = {
    id: "trashygecko-mahjong-teddy",
    level: 4,
    columns: 12,
    rows: 9,
    tileCount: 114,
    layers: [teddyBase, teddyLayer2, teddyLayer3, teddyLayer4],
    positions: [
        ...teddyBase,
        ...teddyLayer2,
        ...teddyLayer3,
        ...teddyLayer4
    ]
};

// --------------------------------------------------
// LEVEL 05 – STERN
// --------------------------------------------------

const starBase = [
    createRanges(0, [[5, 6]], 0),            // Spitze
    createRanges(1, [[5, 6]], 0),
    createRanges(2, [[4, 7]], 0),
    createRanges(3, [[3, 8]], 0),
    createRanges(4, [[0, 11]], 0),           // ausgestellte Arme
    createRanges(5, [[0, 11]], 0),
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[3, 8]], 0),            // Taille
    createRanges(8, [[3, 8]], 0),
    createRanges(9, [[2, 3], [8, 9]], 0),    // Beine
    createRanges(10, [[1, 2], [9, 10]], 0),  // Füße
    createRanges(11, [[1, 2], [9, 10]], 0)
].flat();

const starLayer2 = [
    createRow(2, 4, 4, 1),
    createRow(3, 4, 4, 1),
    createRow(4, 4, 4, 1),
    createRow(5, 4, 4, 1),
    createRow(6, 4, 4, 1),
    createRow(7, 4, 4, 1)
].flat();

const starLayer3 = [
    createRow(3, 4, 4, 2),
    createRow(4, 4, 4, 2),
    createRow(5, 4, 4, 2)
].flat();

const starLayer4 = [
    createRow(3, 5, 2, 3),
    createRow(4, 5, 2, 3),
    createRow(5, 5, 2, 3)
].flat();

const STAR_LAYOUT = {
    id: "trashygecko-mahjong-star",
    level: 5,
    columns: 12,
    rows: 12,
    tileCount: 114,
    layers: [starBase, starLayer2, starLayer3, starLayer4],
    positions: [
        ...starBase,
        ...starLayer2,
        ...starLayer3,
        ...starLayer4
    ]
};

// --------------------------------------------------
// LEVEL 06 – KATZE
// --------------------------------------------------

const catBase = [
    createRanges(0, [[1, 2], [9, 10]], 0),   // Ohren
    createRanges(1, [[1, 3], [8, 10]], 0),
    createRanges(2, [[2, 9]], 0),            // Kopf
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[2, 9]], 0),
    createRanges(6, [[3, 8]], 0),            // Kinn
    createRanges(7, [[2, 9]], 0),            // Körper
    createRanges(8, [[2, 9]], 0),
    createRanges(9, [[3, 4], [7, 8]], 0)     // Pfoten
].flat();

const catLayer2 = [
    createRow(2, 4, 4, 1),
    createRow(3, 4, 4, 1),
    createRow(4, 4, 4, 1),
    createRow(5, 4, 4, 1),
    createRow(6, 4, 4, 1),
    createRow(7, 4, 4, 1)
].flat();

const catLayer3 = [
    createRow(3, 4, 4, 2),
    createRow(4, 4, 4, 2),
    createRow(5, 4, 4, 2)
].flat();

const catLayer4 = [
    createRow(3, 5, 2, 3),
    createRow(4, 5, 2, 3),
    createRow(5, 5, 2, 3)
].flat();

const CAT_LAYOUT = {
    id: "trashygecko-mahjong-cat",
    level: 6,
    columns: 12,
    rows: 10,
    tileCount: 114,
    layers: [catBase, catLayer2, catLayer3, catLayer4],
    positions: [
        ...catBase,
        ...catLayer2,
        ...catLayer3,
        ...catLayer4
    ]
};

// --------------------------------------------------
// DIE DREI OBEREN EBENEN
// --------------------------------------------------

// Alle neuen Formen tragen denselben Aufbau in der Höhe: sechs Reihen mit vier
// Steinen, darüber drei Reihen mit vier und drei Reihen mit zwei Steinen.
function createTopLayers(firstRow) {
    const second = Array.from(
        { length: 6 },
        (_, i) => createRow(firstRow + i, 4, 4, 1)
    ).flat();

    const third = Array.from(
        { length: 3 },
        (_, i) => createRow(firstRow + 1 + i, 4, 4, 2)
    ).flat();

    const fourth = Array.from(
        { length: 3 },
        (_, i) => createRow(firstRow + 1 + i, 5, 2, 3)
    ).flat();

    return [second, third, fourth];
}

function createShapeLayout({ id, level, rows, base, layerStart }) {
    const [second, third, fourth] = createTopLayers(layerStart);

    return {
        id,
        level,
        columns: 12,
        rows,
        tileCount: 114,
        layers: [base, second, third, fourth],
        positions: [...base, ...second, ...third, ...fourth]
    };
}

// --------------------------------------------------
// LEVEL 07 – HÄSCHEN
// --------------------------------------------------

const bunnyBase = [
    createRanges(0, [[3, 4], [7, 8]], 0),    // Löffelohren
    createRanges(1, [[3, 4], [7, 8]], 0),
    createRanges(2, [[2, 9]], 0),            // Kopf
    createRanges(3, [[2, 9]], 0),
    createRanges(4, [[3, 8]], 0),            // Kinn
    createRanges(5, [[2, 9]], 0),            // Körper
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[2, 9]], 0),
    createRanges(9, [[1, 3], [8, 10]], 0)    // Pfoten
].flat();

const BUNNY_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-bunny",
    level: 7,
    rows: 10,
    base: bunnyBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 08 – SCHAF
// --------------------------------------------------

const sheepBase = [
    createRanges(0, [[3, 8]], 0),            // wolliger Rücken
    createRanges(1, [[2, 9]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[2, 9]], 0),
    createRanges(6, [[2, 9]], 0),            // Bauch
    createRanges(7, [[2, 3], [8, 9]], 0),    // Beine
    createRanges(8, [[2, 3], [8, 9]], 0),
    createRanges(9, [[2, 3], [8, 9]], 0)
].flat();

const SHEEP_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-sheep",
    level: 8,
    rows: 10,
    base: sheepBase,
    layerStart: 1
});

// --------------------------------------------------
// LEVEL 09 – PINGUIN
// --------------------------------------------------

const penguinBase = [
    createRanges(0, [[4, 7]], 0),            // Kopf
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[3, 8]], 0),
    createRanges(3, [[3, 8]], 0),
    createRanges(4, [[2, 9]], 0),            // Schultern
    createRanges(5, [[1, 10]], 0),           // Flügel
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[2, 9]], 0),            // Bauch
    createRanges(8, [[2, 9]], 0),
    createRanges(9, [[4, 7]], 0),            // Standfläche
    createRanges(10, [[2, 3], [8, 9]], 0)    // Füße
].flat();

const PENGUIN_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-penguin",
    level: 9,
    rows: 11,
    base: penguinBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 10 – BLUME
// --------------------------------------------------

const flowerBase = [
    createRanges(0, [[4, 7]], 0),            // Blüte oben
    createRanges(1, [[3, 8]], 0),
    createRanges(2, [[2, 9]], 0),
    createRanges(3, [[2, 9]], 0),            // Blütenmitte
    createRanges(4, [[2, 9]], 0),
    createRanges(5, [[3, 8]], 0),
    createRanges(6, [[4, 7]], 0),            // Blüte unten
    createRanges(7, [[1, 10]], 0),           // Blätter und Stiel
    createRanges(8, [[2, 9]], 0),
    createRanges(9, [[3, 8]], 0),
    createRanges(10, [[5, 6]], 0),           // Stiel
    createRanges(11, [[5, 6]], 0)
].flat();

const FLOWER_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-flower",
    level: 10,
    rows: 12,
    base: flowerBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 11 – PILZ
// --------------------------------------------------

const mushroomBase = [
    createRanges(0, [[3, 8]], 0),            // Kappe
    createRanges(1, [[1, 10]], 0),
    createRanges(2, [[0, 11]], 0),
    createRanges(3, [[0, 11]], 0),
    createRanges(4, [[1, 10]], 0),           // Kappenrand
    createRanges(5, [[4, 7]], 0),            // Stiel
    createRanges(6, [[4, 7]], 0),
    createRanges(7, [[4, 7]], 0),
    createRanges(8, [[4, 7]], 0),
    createRanges(9, [[3, 8]], 0)             // Fuß
].flat();

const MUSHROOM_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-mushroom",
    level: 11,
    rows: 10,
    base: mushroomBase,
    layerStart: 1
});

// --------------------------------------------------
// LEVEL 12 – HAUS
// --------------------------------------------------

const houseBase = [
    createRanges(0, [[5, 6]], 0),            // Dachspitze
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[3, 8]], 0),
    createRanges(3, [[2, 9]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[0, 11]], 0),           // Dachrand
    createRanges(6, [[1, 10]], 0),           // Wände
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[1, 10]], 0)
].flat();

const HOUSE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-house",
    level: 12,
    rows: 9,
    base: houseBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 13 – FISCH
// --------------------------------------------------

const fishBase = [
    createRanges(0, [[4, 7]], 0),            // Rückenflosse
    createRanges(1, [[3, 9]], 0),
    createRanges(2, [[2, 9]], 0),
    createRanges(3, [[1, 11]], 0),           // Kopf bis Schwanzansatz
    createRanges(4, [[0, 11]], 0),           // breiteste Stelle
    createRanges(5, [[1, 11]], 0),
    createRanges(6, [[2, 9]], 0),
    createRanges(7, [[3, 9]], 0),
    createRanges(8, [[4, 7]], 0)             // Bauchflosse
].flat();

const FISH_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-fish",
    level: 13,
    rows: 9,
    base: fishBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 14 – WOLKE
// --------------------------------------------------

const cloudBase = [
    createRanges(0, [[3, 4], [8, 9]], 0),    // zwei Buckel
    createRanges(1, [[2, 10]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[0, 11]], 0),
    createRanges(4, [[0, 11]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[2, 9]], 0),
    createRanges(7, [[3, 9]], 0)             // flacher Bauch
].flat();

const CLOUD_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-cloud",
    level: 14,
    rows: 8,
    base: cloudBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 15 – KLEEBLATT
// --------------------------------------------------

const cloverBase = [
    createRanges(0, [[4, 7]], 0),            // oberes Blatt
    createRanges(1, [[3, 8]], 0),
    createRanges(2, [[3, 8]], 0),
    createRanges(3, [[0, 11]], 0),           // Seitenblätter
    createRanges(4, [[0, 11]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[3, 8]], 0),            // unteres Blatt
    createRanges(7, [[3, 8]], 0),
    createRanges(8, [[4, 7]], 0),
    createRanges(9, [[4, 7]], 0),            // Stiel
    createRanges(10, [[5, 6]], 0)
].flat();

const CLOVER_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-clover",
    level: 15,
    rows: 11,
    base: cloverBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVELÜBERSICHT
// --------------------------------------------------

const BOARD_LAYOUTS = {
    1: BOARD_LAYOUT,
    2: HEART_LAYOUT,
    3: CROSS_LAYOUT,
    4: TEDDY_LAYOUT,
    5: STAR_LAYOUT,
    6: CAT_LAYOUT,
    7: BUNNY_LAYOUT,
    8: SHEEP_LAYOUT,
    9: PENGUIN_LAYOUT,
    10: FLOWER_LAYOUT,
    11: MUSHROOM_LAYOUT,
    12: HOUSE_LAYOUT,
    13: FISH_LAYOUT,
    14: CLOUD_LAYOUT,
    15: CLOVER_LAYOUT
};
