
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
    id: "trashygecko-mahjong-144",
    level: 1,
    columns: 12,
    rows: 8,
    tileCount: 144,
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
    tileCount: 144,
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
    tileCount: 144,
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
    tileCount: 144,
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
    tileCount: 144,
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
    tileCount: 144,
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
        tileCount: 144,
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
// LEVEL 16 – KRONE
// --------------------------------------------------

const kroneBase = [
    createRanges(0, [[2, 3], [5, 6], [8, 9]], 0),   // drei Spitzen
    createRanges(1, [[2, 3], [5, 6], [8, 9]], 0),
    createRanges(2, [[2, 3], [5, 6], [8, 9]], 0),
    createRanges(3, [[1, 10]], 0),                  // Krone
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[0, 11]], 0),                  // breites Band
    createRanges(7, [[0, 11]], 0),
    createRanges(8, [[0, 11]], 0),
    createRanges(9, [[1, 10]], 0)                   // Abschluss
].flat();

const KRONE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-krone",
    level: 16,
    rows: 10,
    base: kroneBase,
    layerStart: 3
});

// --------------------------------------------------
// LEVEL 17 – SCHMETTERLING
// --------------------------------------------------

const schmetterlingBase = [
    createRanges(0, [[2, 3], [8, 9]], 0),           // Fühlerspitzen
    createRanges(1, [[1, 4], [7, 10]], 0),          // obere Flügel
    createRanges(2, [[0, 4], [7, 11]], 0),
    createRanges(3, [[0, 4], [7, 11]], 0),
    createRanges(4, [[1, 4], [7, 10]], 0),
    createRanges(5, [[3, 4], [7, 8]], 0),           // Flügelansatz
    createRanges(6, [[4, 7]], 0),                   // Körper
    createRanges(7, [[4, 7]], 0),
    createRanges(8, [[3, 8]], 0),                   // untere Flügel
    createRanges(9, [[2, 9]], 0),
    createRanges(10, [[1, 10]], 0),
    createRanges(11, [[2, 9]], 0)
].flat();

const SCHMETTERLING_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-schmetterling",
    level: 17,
    rows: 12,
    base: schmetterlingBase,
    layerStart: 6
});

// --------------------------------------------------
// LEVEL 18 – RAKETE
// --------------------------------------------------

const raketeBase = [
    createRanges(0, [[5, 6]], 0),                   // Spitze
    createRanges(1, [[5, 6]], 0),
    createRanges(2, [[4, 7]], 0),
    createRanges(3, [[4, 7]], 0),
    createRanges(4, [[3, 8]], 0),                   // Rumpf
    createRanges(5, [[3, 8]], 0),
    createRanges(6, [[2, 9]], 0),
    createRanges(7, [[2, 9]], 0),
    createRanges(8, [[1, 10]], 0),
    createRanges(9, [[1, 4], [7, 10]], 0),          // Flossen
    createRanges(10, [[3, 8]], 0),
    createRanges(11, [[3, 8]], 0),
    createRanges(12, [[4, 7]], 0),                  // Triebwerk
    createRanges(13, [[3, 8]], 0)
].flat();

const RAKETE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-rakete",
    level: 18,
    rows: 14,
    base: raketeBase,
    layerStart: 4
});

// --------------------------------------------------
// LEVEL 19 – SCHNEEMANN
// --------------------------------------------------

const schneemannBase = [
    createRanges(0, [[4, 7]], 0),                   // Kopf
    createRanges(1, [[3, 8]], 0),
    createRanges(2, [[3, 8]], 0),
    createRanges(3, [[4, 7]], 0),                   // Hals
    createRanges(4, [[2, 9]], 0),                   // Schal
    createRanges(5, [[2, 9]], 0),
    createRanges(6, [[1, 10]], 0),                  // Bauch
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[2, 9]], 0),
    createRanges(9, [[2, 9]], 0),
    createRanges(10, [[1, 10]], 0),                 // unterer Bauch
    createRanges(11, [[1, 10]], 0)
].flat();

const SCHNEEMANN_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-schneemann",
    level: 19,
    rows: 12,
    base: schneemannBase,
    layerStart: 4
});

// --------------------------------------------------
// LEVEL 20 – BAUM
// --------------------------------------------------

const baumBase = [
    createRanges(0, [[5, 6]], 0),                   // Spitze
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[3, 8]], 0),
    createRanges(3, [[4, 7]], 0),                   // Astlücke
    createRanges(4, [[2, 9]], 0),
    createRanges(5, [[3, 8]], 0),
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[2, 9]], 0),
    createRanges(8, [[0, 11]], 0),                  // breiteste Stelle
    createRanges(9, [[2, 9]], 0),
    createRanges(10, [[5, 6]], 0),                  // Stamm
    createRanges(11, [[5, 6]], 0)
].flat();

const BAUM_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-baum",
    level: 20,
    rows: 12,
    base: baumBase,
    layerStart: 4
});

// --------------------------------------------------
// LEVEL 21 – TASSE
// --------------------------------------------------

const tasseBase = [
    createRanges(0, [[1, 10]], 0),                  // Rand
    createRanges(1, [[1, 10]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[2, 9]], 0),                   // Wandung
    createRanges(7, [[2, 9]], 0),
    createRanges(8, [[3, 8]], 0),
    createRanges(9, [[3, 8]], 0)
].flat();

const TASSE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-tasse",
    level: 21,
    rows: 10,
    base: tasseBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 22 – KLEID
// --------------------------------------------------

const kleidBase = [
    createRanges(0, [[4, 7]], 0),                   // Ausschnitt
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[3, 8]], 0),                   // Oberteil
    createRanges(3, [[3, 8]], 0),
    createRanges(4, [[2, 9]], 0),                   // Taille
    createRanges(5, [[1, 10]], 0),                  // Rock
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[0, 11]], 0),
    createRanges(8, [[0, 11]], 0),
    createRanges(9, [[1, 10]], 0)                   // Saum
].flat();

const KLEID_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-kleid",
    level: 22,
    rows: 10,
    base: kleidBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 23 – SCHLEIFE
// --------------------------------------------------

const schleifeBase = [
    createRanges(0, [[1, 3], [8, 10]], 0),          // Schlaufen oben
    createRanges(1, [[0, 4], [7, 11]], 0),
    createRanges(2, [[0, 4], [7, 11]], 0),
    createRanges(3, [[1, 4], [7, 10]], 0),
    createRanges(4, [[4, 7]], 0),                   // Knoten
    createRanges(5, [[3, 8]], 0),                   // Bänder
    createRanges(6, [[2, 9]], 0),
    createRanges(7, [[3, 8]], 0),
    createRanges(8, [[4, 7]], 0),
    createRanges(9, [[3, 8]], 0)                    // Bandenden
].flat();

const SCHLEIFE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-schleife",
    level: 23,
    rows: 10,
    base: schleifeBase,
    layerStart: 4
});

// --------------------------------------------------
// LEVEL 24 – EISTÜTE
// --------------------------------------------------

const eistueteBase = [
    createRanges(0, [[3, 8]], 0),                   // Kugel oben
    createRanges(1, [[2, 9]], 0),
    createRanges(2, [[1, 10]], 0),                  // breite Kugel
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[2, 9]], 0),
    createRanges(5, [[3, 8]], 0),                   // Waffelrand
    createRanges(6, [[3, 8]], 0),
    createRanges(7, [[4, 7]], 0),                   // Tüte
    createRanges(8, [[4, 7]], 0),
    createRanges(9, [[5, 6]], 0),
    createRanges(10, [[5, 6]], 0)                   // Spitze
].flat();

const EISTUETE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-eistuete",
    level: 24,
    rows: 11,
    base: eistueteBase,
    layerStart: 1
});

// --------------------------------------------------
// LEVEL 25 – HEISSLUFTBALLON
// --------------------------------------------------

const ballonBase = [
    createRanges(0, [[5, 6]], 0),                   // Kuppe
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[3, 8]], 0),
    createRanges(3, [[2, 9]], 0),
    createRanges(4, [[1, 10]], 0),                  // breiteste Stelle
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[2, 9]], 0),
    createRanges(7, [[3, 8]], 0),
    createRanges(8, [[4, 7]], 0),                   // Hals
    createRanges(9, [[5, 6]], 0),                   // Seile
    createRanges(10, [[4, 7]], 0)                   // Korb
].flat();

const BALLON_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-balloon",
    level: 25,
    rows: 11,
    base: ballonBase,
    layerStart: 1
});

// --------------------------------------------------
// LEVEL 26 – MONDSICHEL
// --------------------------------------------------

const mondBase = [
    createRanges(0, [[3, 9]], 0),                   // obere Rundung
    createRanges(1, [[2, 10]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 9]], 0),                   // Einkerbung rechts
    createRanges(4, [[1, 8]], 0),
    createRanges(5, [[1, 8]], 0),
    createRanges(6, [[1, 9]], 0),
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[2, 10]], 0),
    createRanges(9, [[3, 9]], 0)                    // untere Rundung
].flat();

const MOND_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-mond",
    level: 26,
    rows: 10,
    base: mondBase,
    layerStart: 3
});

// --------------------------------------------------
// LEVEL 27 – HERZ MIT FLÜGELN
// --------------------------------------------------

const fluegelherzBase = [
    createRanges(0, [[1, 2], [9, 10]], 0),          // Flügelspitzen
    createRanges(1, [[0, 3], [8, 11]], 0),
    createRanges(2, [[0, 4], [7, 11]], 0),          // Flügel wachsen ins Herz
    createRanges(3, [[3, 8]], 0),                   // Herzschultern
    createRanges(4, [[2, 9]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[2, 9]], 0),
    createRanges(8, [[3, 8]], 0),
    createRanges(9, [[4, 7]], 0)                    // Herzspitze
].flat();

const FLUEGELHERZ_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-fluegelherz",
    level: 27,
    rows: 10,
    base: fluegelherzBase,
    layerStart: 3
});

// --------------------------------------------------
// LEVEL 28 – KIRSCHBLÜTE
// --------------------------------------------------

const kirschblueteBase = [
    createRanges(0, [[4, 7]], 0),                   // oberes Blütenblatt
    createRanges(1, [[3, 8]], 0),
    createRanges(2, [[0, 3], [8, 11]], 0),          // seitliche Blütenblätter
    createRanges(3, [[0, 3], [8, 11]], 0),
    createRanges(4, [[1, 10]], 0),                  // Blütenmitte
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[0, 11]], 0),                  // untere Blätter
    createRanges(8, [[1, 10]], 0),
    createRanges(9, [[2, 9]], 0)
].flat();

const KIRSCHBLUETE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-kirschbluete",
    level: 28,
    rows: 10,
    base: kirschblueteBase,
    layerStart: 4
});

// --------------------------------------------------
// LEVEL 29 – SCHWAN
// --------------------------------------------------

const schwanBase = [
    createRanges(0, [[6, 8]], 0),                   // Kopf
    createRanges(1, [[5, 7]], 0),                   // Hals
    createRanges(2, [[5, 7]], 0),
    createRanges(3, [[4, 6]], 0),
    createRanges(4, [[1, 7]], 0),                   // Rücken
    createRanges(5, [[0, 8]], 0),
    createRanges(6, [[0, 9]], 0),
    createRanges(7, [[1, 10]], 0),                  // Körper
    createRanges(8, [[2, 9]], 0),
    createRanges(9, [[3, 8]], 0),
    createRanges(10, [[4, 7]], 0)                   // Wasserlinie
].flat();

const SCHWAN_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-schwan",
    level: 29,
    rows: 11,
    base: schwanBase,
    layerStart: 4
});

// --------------------------------------------------
// LEVEL 30 – BALLERINA
// --------------------------------------------------

const ballerinaBase = [
    createRanges(0, [[5, 6]], 0),                   // Kopf
    createRanges(1, [[5, 6]], 0),
    createRanges(2, [[4, 7]], 0),                   // Schultern
    createRanges(3, [[3, 8]], 0),
    createRanges(4, [[2, 9]], 0),                   // Tutu
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[0, 11]], 0),                  // weiteste Stelle
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[4, 7]], 0),                   // Beine
    createRanges(9, [[4, 7]], 0),
    createRanges(10, [[5, 6]], 0),                  // Spitze
    createRanges(11, [[4, 7]], 0)
].flat();

const BALLERINA_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-ballerina",
    level: 30,
    rows: 12,
    base: ballerinaBase,
    layerStart: 4
});

// --------------------------------------------------
// LEVEL 31 – MUSCHEL
// --------------------------------------------------

const muschelBase = [
    createRanges(0, [[0, 11]], 0),                  // Fächerrand
    createRanges(1, [[0, 11]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[2, 9]], 0),
    createRanges(5, [[2, 9]], 0),
    createRanges(6, [[3, 8]], 0),
    createRanges(7, [[3, 8]], 0),
    createRanges(8, [[4, 7]], 0),                   // Scharnier
    createRanges(9, [[4, 7]], 0)
].flat();

const MUSCHEL_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-muschel",
    level: 31,
    rows: 10,
    base: muschelBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 32 – EULE
// --------------------------------------------------

const euleBase = [
    createRanges(0, [[2, 3], [8, 9]], 0),
    createRanges(1, [[2, 4], [7, 9]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[2, 9]], 0),
    createRanges(6, [[2, 9]], 0),
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[1, 10]], 0),
    createRanges(9, [[3, 8]], 0)
].flat();

const EULE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-eule",
    level: 32,
    rows: 10,
    base: euleBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 33 – GITARRE
// --------------------------------------------------

const gitarreBase = [
    createRanges(0, [[5, 6]], 0),
    createRanges(1, [[5, 6]], 0),
    createRanges(2, [[4, 7]], 0),
    createRanges(3, [[4, 7]], 0),
    createRanges(4, [[2, 9]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[2, 9]], 0),
    createRanges(9, [[3, 8]], 0)
].flat();

const GITARRE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-gitarre",
    level: 33,
    rows: 10,
    base: gitarreBase,
    layerStart: 4
});

// --------------------------------------------------
// LEVEL 34 – SCHLÜSSEL
// --------------------------------------------------

const schluesselBase = [
    createRanges(0, [[2, 9]], 0),
    createRanges(1, [[1, 10]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[2, 9]], 0),
    createRanges(5, [[4, 7]], 0),
    createRanges(6, [[4, 7]], 0),
    createRanges(7, [[4, 7]], 0),
    createRanges(8, [[4, 9]], 0),
    createRanges(9, [[4, 9]], 0)
].flat();

const SCHLUESSEL_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-schluessel",
    level: 34,
    rows: 10,
    base: schluesselBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 35 – ANKER
// --------------------------------------------------

const ankerBase = [
    createRanges(0, [[4, 7]], 0),
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[4, 7]], 0),
    createRanges(6, [[4, 7]], 0),
    createRanges(7, [[4, 7]], 0),
    createRanges(8, [[2, 9]], 0),
    createRanges(9, [[1, 3], [8, 10]], 0),
    createRanges(10, [[0, 2], [9, 11]], 0)
].flat();

const ANKER_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-anker",
    level: 35,
    rows: 11,
    base: ankerBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 36 – LEUCHTTURM
// --------------------------------------------------

const leuchtturmBase = [
    createRanges(0, [[5, 6]], 0),
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[4, 7]], 0),
    createRanges(3, [[3, 8]], 0),
    createRanges(4, [[3, 8]], 0),
    createRanges(5, [[2, 9]], 0),
    createRanges(6, [[2, 9]], 0),
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[1, 10]], 0),
    createRanges(9, [[0, 11]], 0)
].flat();

const LEUCHTTURM_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-leuchtturm",
    level: 36,
    rows: 10,
    base: leuchtturmBase,
    layerStart: 4
});

// --------------------------------------------------
// LEVEL 37 – SCHLOSS
// --------------------------------------------------

const schlossBase = [
    createRanges(0, [[2, 3], [8, 9]], 0),
    createRanges(1, [[2, 3], [8, 9]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[2, 9]], 0),
    createRanges(8, [[3, 8]], 0),
    createRanges(9, [[4, 7]], 0)
].flat();

const SCHLOSS_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-schloss",
    level: 37,
    rows: 10,
    base: schlossBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 38 – SONNENBLUME
// --------------------------------------------------

const sonnenblumeBase = [
    createRanges(0, [[4, 7]], 0),
    createRanges(1, [[2, 9]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[0, 11]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[2, 9]], 0),
    createRanges(8, [[4, 7]], 0)
].flat();

const SONNENBLUME_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-sonnenblume",
    level: 38,
    rows: 9,
    base: sonnenblumeBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 39 – TULPE
// --------------------------------------------------

const tulpeBase = [
    createRanges(0, [[3, 4], [7, 8]], 0),
    createRanges(1, [[2, 9]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[2, 9]], 0),
    createRanges(6, [[4, 7]], 0),
    createRanges(7, [[4, 7]], 0),
    createRanges(8, [[2, 5], [6, 9]], 0),
    createRanges(9, [[1, 4], [7, 10]], 0)
].flat();

const TULPE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-tulpe",
    level: 39,
    rows: 10,
    base: tulpeBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 40 – VOGEL
// --------------------------------------------------

const vogelBase = [
    createRanges(0, [[7, 9]], 0),
    createRanges(1, [[1, 9]], 0),
    createRanges(2, [[0, 10]], 0),
    createRanges(3, [[0, 8]], 0),
    createRanges(4, [[1, 7]], 0),
    createRanges(5, [[2, 7]], 0),
    createRanges(6, [[3, 8]], 0),
    createRanges(7, [[4, 9]], 0),
    createRanges(8, [[4, 9]], 0),
    createRanges(9, [[4, 7]], 0)
].flat();

const VOGEL_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-vogel",
    level: 40,
    rows: 10,
    base: vogelBase,
    layerStart: 3
});

// --------------------------------------------------
// LEVEL 41 – SCHNECKE
// --------------------------------------------------

const schneckeBase = [
    createRanges(0, [[3, 7]], 0),
    createRanges(1, [[2, 8]], 0),
    createRanges(2, [[1, 9]], 0),
    createRanges(3, [[1, 9]], 0),
    createRanges(4, [[1, 9]], 0),
    createRanges(5, [[1, 9]], 0),
    createRanges(6, [[0, 9]], 0),
    createRanges(7, [[0, 10]], 0),
    createRanges(8, [[1, 10]], 0),
    createRanges(9, [[2, 9]], 0)
].flat();

const SCHNECKE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-schnecke",
    level: 41,
    rows: 10,
    base: schneckeBase,
    layerStart: 3
});

// --------------------------------------------------
// LEVEL 42 – VULKAN
// --------------------------------------------------

const vulkanBase = [
    createRanges(0, [[4, 7]], 0),
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[3, 8]], 0),
    createRanges(3, [[2, 9]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[0, 11]], 0),
    createRanges(7, [[0, 11]], 0),
    createRanges(8, [[1, 10]], 0)
].flat();

const VULKAN_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-vulkan",
    level: 42,
    rows: 9,
    base: vulkanBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 43 – KOMPASS
// --------------------------------------------------

const kompassBase = [
    createRanges(0, [[5, 6]], 0),
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[3, 8]], 0),
    createRanges(3, [[2, 9]], 0),
    createRanges(4, [[0, 11]], 0),
    createRanges(5, [[0, 11]], 0),
    createRanges(6, [[2, 9]], 0),
    createRanges(7, [[3, 8]], 0),
    createRanges(8, [[4, 7]], 0),
    createRanges(9, [[5, 6]], 0)
].flat();

const KOMPASS_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-kompass",
    level: 43,
    rows: 10,
    base: kompassBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 44 – REGENSCHIRM
// --------------------------------------------------

const regenschirmBase = [
    createRanges(0, [[5, 6]], 0),
    createRanges(1, [[3, 8]], 0),
    createRanges(2, [[2, 9]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[0, 11]], 0),
    createRanges(6, [[0, 11]], 0),
    createRanges(7, [[4, 7]], 0),
    createRanges(8, [[4, 7]], 0),
    createRanges(9, [[4, 7]], 0),
    createRanges(10, [[3, 7]], 0)
].flat();

const REGENSCHIRM_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-regenschirm",
    level: 44,
    rows: 11,
    base: regenschirmBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 45 – KAMERA
// --------------------------------------------------

const kameraBase = [
    createRanges(0, [[3, 8]], 0),
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[0, 11]], 0),
    createRanges(5, [[0, 11]], 0),
    createRanges(6, [[0, 11]], 0),
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[2, 9]], 0)
].flat();

const KAMERA_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-kamera",
    level: 45,
    rows: 9,
    base: kameraBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 46 – LATERNE
// --------------------------------------------------

const laterneBase = [
    createRanges(0, [[5, 6]], 0),
    createRanges(1, [[4, 7]], 0),
    createRanges(2, [[3, 8]], 0),
    createRanges(3, [[2, 9]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[1, 10]], 0),
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[3, 8]], 0)
].flat();

const LATERNE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-laterne",
    level: 46,
    rows: 9,
    base: laterneBase,
    layerStart: 3
});

// --------------------------------------------------
// LEVEL 47 – POKAL
// --------------------------------------------------

const pokalBase = [
    createRanges(0, [[0, 11]], 0),
    createRanges(1, [[0, 11]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[2, 9]], 0),
    createRanges(4, [[4, 7]], 0),
    createRanges(5, [[4, 7]], 0),
    createRanges(6, [[4, 7]], 0),
    createRanges(7, [[2, 9]], 0),
    createRanges(8, [[1, 10]], 0)
].flat();

const POKAL_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-pokal",
    level: 47,
    rows: 9,
    base: pokalBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 48 – SEGELBOOT
// --------------------------------------------------

const segelbootBase = [
    createRanges(0, [[5, 9]], 0),
    createRanges(1, [[4, 9]], 0),
    createRanges(2, [[3, 9]], 0),
    createRanges(3, [[2, 9]], 0),
    createRanges(4, [[1, 9]], 0),
    createRanges(5, [[0, 10]], 0),
    createRanges(6, [[0, 10]], 0),
    createRanges(7, [[1, 9]], 0),
    createRanges(8, [[2, 8]], 0)
].flat();

const SEGELBOOT_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-segelboot",
    level: 48,
    rows: 9,
    base: segelbootBase,
    layerStart: 3
});

// --------------------------------------------------
// LEVEL 49 – MARIENKÄFER
// --------------------------------------------------

const marienkaeferBase = [
    createRanges(0, [[4, 7]], 0),
    createRanges(1, [[3, 8]], 0),
    createRanges(2, [[2, 9]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[1, 10]], 0),
    createRanges(5, [[0, 11]], 0),
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[1, 10]], 0),
    createRanges(8, [[2, 9]], 0),
    createRanges(9, [[3, 8]], 0)
].flat();

const MARIENKAEFER_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-marienkaefer",
    level: 49,
    rows: 10,
    base: marienkaeferBase,
    layerStart: 2
});

// --------------------------------------------------
// LEVEL 50 – GEBURTSTAGSTORTE
// --------------------------------------------------

const torteBase = [
    createRanges(0, [[5, 6]], 0),
    createRanges(1, [[5, 6]], 0),
    createRanges(2, [[1, 10]], 0),
    createRanges(3, [[1, 10]], 0),
    createRanges(4, [[0, 11]], 0),
    createRanges(5, [[0, 11]], 0),
    createRanges(6, [[1, 10]], 0),
    createRanges(7, [[0, 11]], 0),
    createRanges(8, [[0, 11]], 0),
    createRanges(9, [[1, 10]], 0)
].flat();

const TORTE_LAYOUT = createShapeLayout({
    id: "trashygecko-mahjong-torte",
    level: 50,
    rows: 10,
    base: torteBase,
    layerStart: 2
});

// --------------------------------------------------
// ERWEITERUNG AUF 144 STEINE
// --------------------------------------------------

// Ein vollstaendiges Mahjong-Spiel hat 144 Steine, also 72 Paare. Die fehlenden
// Steine kommen als zweite Ebene ueber die freien Positionen der Grundflaeche,
// damit die Silhouette jedes Levels unveraendert bleibt.
const STEINE_PRO_LEVEL = 144;

function erweitereAuf144(layout) {
    const fehlend = STEINE_PRO_LEVEL - layout.positions.length;

    if (fehlend <= 0) return layout;

    const mitteX = (layout.columns - 1) / 2;
    const mitteY = (layout.rows - 1) / 2;
    const abstand = position =>
        Math.pow(position.x - mitteX, 2) + Math.pow(position.y - mitteY, 2);

    const frei = layout.positions
        .filter(position => position.z === 0)
        .filter(position => !layout.positions.some(oben =>
            oben.x === position.x && oben.y === position.y && oben.z > 0
        ))
        .sort((a, b) => abstand(a) - abstand(b));

    const zusatz = frei
        .slice(0, fehlend)
        .map(position => ({ x: position.x, y: position.y, z: 1 }));

    if (zusatz.length < fehlend) {
        throw new Error("Level " + layout.level + " hat zu wenige freie Positionen.");
    }

    layout.layers.push(zusatz);
    layout.positions = [...layout.positions, ...zusatz];
    layout.tileCount = layout.positions.length;

    return layout;
}

// --------------------------------------------------
// EINFACHE FORMEN (Level 51-54)
// --------------------------------------------------

// Diese Formen bringen ihre Ebenen selbst mit, damit die Steinzahl exakt
// aufgeht und keine Fuellsteine noetig sind.
function createLayeredLayout({ id, level, rows, columns, layers }) {
    const positions = layers.flat();

    return {
        id,
        level,
        columns,
        rows,
        tileCount: positions.length,
        layers,
        positions
    };
}

// LEVEL 51 - WUERFEL: drei gleiche Ebenen uebereinander (8 x 6)
const WUERFEL_LAYOUT = createLayeredLayout({
    id: "trashygecko-mahjong-wuerfel",
    level: 51,
    rows: 6,
    columns: 12,
    layers: [0, 1, 2].map(z =>
        Array.from({ length: 6 }, (_, i) => createRow(i, 2, 8, z)).flat()
    )
});

// LEVEL 52 - RECHTECK: 12 x 8 flach, darueber 8 x 6
const RECHTECK_LAYOUT = createLayeredLayout({
    id: "trashygecko-mahjong-rechteck",
    level: 52,
    rows: 8,
    columns: 12,
    layers: [
        Array.from({ length: 8 }, (_, i) => createRow(i, 0, 12, 0)).flat(),
        Array.from({ length: 6 }, (_, i) => createRow(i + 1, 2, 8, 1)).flat()
    ]
});

// LEVEL 53 - BALL: runde Grundflaeche mit einer kleineren inneren Ebene
const BALL_LAYOUT = createLayeredLayout({
    id: "trashygecko-mahjong-ball",
    level: 53,
    rows: 10,
    columns: 12,
    layers: [
        [
            createRow(0, 3, 6, 0),
            createRow(1, 2, 8, 0),
            createRow(2, 1, 10, 0),
            createRow(3, 0, 12, 0),
            createRow(4, 0, 12, 0),
            createRow(5, 0, 12, 0),
            createRow(6, 0, 12, 0),
            createRow(7, 1, 10, 0),
            createRow(8, 2, 8, 0),
            createRow(9, 3, 6, 0)
        ].flat(),
        [
            createRow(2, 3, 6, 1),
            createRow(3, 2, 8, 1),
            createRow(4, 1, 10, 1),
            createRow(5, 1, 10, 1),
            createRow(6, 2, 8, 1),
            createRow(7, 3, 6, 1)
        ].flat()
    ]
});

// LEVEL 54 - PYRAMIDE: nach oben schmaler werdende Treppe
const PYRAMIDE_LAYOUT = createLayeredLayout({
    id: "trashygecko-mahjong-pyramide",
    level: 54,
    rows: 9,
    columns: 12,
    layers: [
        [
            createRow(0, 5, 2, 0),
            createRow(1, 4, 4, 0),
            createRow(2, 3, 6, 0),
            createRow(3, 2, 8, 0),
            createRow(4, 1, 10, 0),
            createRow(5, 0, 12, 0),
            createRow(6, 0, 12, 0),
            createRow(7, 0, 12, 0),
            createRow(8, 0, 12, 0)
        ].flat(),
        [
            createRow(1, 5, 2, 1),
            createRow(2, 4, 4, 1),
            createRow(3, 3, 6, 1),
            createRow(4, 2, 8, 1),
            createRow(5, 1, 10, 1),
            createRow(6, 0, 12, 1)
        ].flat(),
        [
            createRow(2, 4, 4, 2),
            createRow(3, 3, 6, 2),
            createRow(4, 2, 8, 2),
            createRow(5, 3, 6, 2)
        ].flat()
    ]
});

// LEVEL 55 - SAEULEN: zwoelf Tuermchen (2 x 2), je drei Ebenen hoch,
// dazwischen eine Gasse - wie eine kleine Halle.
const SAEULEN_X = [0, 3, 6, 9];
const SAEULEN_Y = [0, 3, 6];

function saeulenEbene(z) {
    const steine = [];

    SAEULEN_Y.forEach(y => {
        SAEULEN_X.forEach(x => {
            steine.push(createRow(y, x, 2, z));
            steine.push(createRow(y + 1, x, 2, z));
        });
    });

    return steine.flat();
}

const SAEULEN_LAYOUT = createLayeredLayout({
    id: "trashygecko-mahjong-saeulen",
    level: 55,
    rows: 8,
    columns: 11,
    layers: [saeulenEbene(0), saeulenEbene(1), saeulenEbene(2)]
});

// LEVEL 56 - BRUECKE: zwei dicke Pfeiler, darueber ein Steg, der die Luecke
// ueberspannt - in der Mitte liegen die Steine also ueber leerem Raum.
function brueckenPfeiler(z) {
    const steine = [];

    for (let y = 0; y < 8; y++) {
        steine.push(createRow(y, 0, 3, z));   // linker Pfeiler
        steine.push(createRow(y, 9, 3, z));   // rechter Pfeiler
    }

    return steine.flat();
}

const BRUECKE_DECK = Array.from(
    { length: 4 },
    (_, i) => createRow(i + 2, 0, 12, 2)
).flat();

const BRUECKE_LAYOUT = createLayeredLayout({
    id: "trashygecko-mahjong-bruecke",
    level: 56,
    rows: 8,
    columns: 12,
    layers: [brueckenPfeiler(0), brueckenPfeiler(1), BRUECKE_DECK]
});

// --------------------------------------------------
// LEVELUEBERSICHT
// --------------------------------------------------

const BOARD_LAYOUTS = {
    1: erweitereAuf144(BOARD_LAYOUT),
    2: erweitereAuf144(HEART_LAYOUT),
    3: erweitereAuf144(CROSS_LAYOUT),
    4: erweitereAuf144(TEDDY_LAYOUT),
    5: erweitereAuf144(STAR_LAYOUT),
    6: erweitereAuf144(CAT_LAYOUT),
    7: erweitereAuf144(BUNNY_LAYOUT),
    8: erweitereAuf144(SHEEP_LAYOUT),
    9: erweitereAuf144(PENGUIN_LAYOUT),
    10: erweitereAuf144(FLOWER_LAYOUT),
    11: erweitereAuf144(MUSHROOM_LAYOUT),
    12: erweitereAuf144(HOUSE_LAYOUT),
    13: erweitereAuf144(FISH_LAYOUT),
    14: erweitereAuf144(CLOUD_LAYOUT),
    15: erweitereAuf144(CLOVER_LAYOUT),
    16: erweitereAuf144(KRONE_LAYOUT),
    17: erweitereAuf144(SCHMETTERLING_LAYOUT),
    18: erweitereAuf144(RAKETE_LAYOUT),
    19: erweitereAuf144(SCHNEEMANN_LAYOUT),
    20: erweitereAuf144(BAUM_LAYOUT),
    21: erweitereAuf144(TASSE_LAYOUT),
    22: erweitereAuf144(KLEID_LAYOUT),
    23: erweitereAuf144(SCHLEIFE_LAYOUT),
    24: erweitereAuf144(EISTUETE_LAYOUT),
    25: erweitereAuf144(BALLON_LAYOUT),
    26: erweitereAuf144(MOND_LAYOUT),
    27: erweitereAuf144(FLUEGELHERZ_LAYOUT),
    28: erweitereAuf144(KIRSCHBLUETE_LAYOUT),
    29: erweitereAuf144(SCHWAN_LAYOUT),
    30: erweitereAuf144(BALLERINA_LAYOUT),
    31: erweitereAuf144(MUSCHEL_LAYOUT),
    32: erweitereAuf144(EULE_LAYOUT),
    33: erweitereAuf144(GITARRE_LAYOUT),
    34: erweitereAuf144(SCHLUESSEL_LAYOUT),
    35: erweitereAuf144(ANKER_LAYOUT),
    36: erweitereAuf144(LEUCHTTURM_LAYOUT),
    37: erweitereAuf144(SCHLOSS_LAYOUT),
    38: erweitereAuf144(SONNENBLUME_LAYOUT),
    39: erweitereAuf144(TULPE_LAYOUT),
    40: erweitereAuf144(VOGEL_LAYOUT),
    41: erweitereAuf144(SCHNECKE_LAYOUT),
    42: erweitereAuf144(VULKAN_LAYOUT),
    43: erweitereAuf144(KOMPASS_LAYOUT),
    44: erweitereAuf144(REGENSCHIRM_LAYOUT),
    45: erweitereAuf144(KAMERA_LAYOUT),
    46: erweitereAuf144(LATERNE_LAYOUT),
    47: erweitereAuf144(POKAL_LAYOUT),
    48: erweitereAuf144(SEGELBOOT_LAYOUT),
    49: erweitereAuf144(MARIENKAEFER_LAYOUT),
    50: erweitereAuf144(TORTE_LAYOUT),
    51: erweitereAuf144(WUERFEL_LAYOUT),
    52: erweitereAuf144(RECHTECK_LAYOUT),
    53: erweitereAuf144(BALL_LAYOUT),
    54: erweitereAuf144(PYRAMIDE_LAYOUT),
    55: erweitereAuf144(SAEULEN_LAYOUT),
    56: erweitereAuf144(BRUECKE_LAYOUT)
};
