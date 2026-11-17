/*
 * TRASHYGECKO MAHJONG
 * Zweiter Formensatz – 56 neue Formen
 *
 * Regeln für diesen Satz (so gewünscht):
 *   - höchstens 6 Steine breit, höchstens 8 Steine hoch (Handyformat)
 *   - die Tiefe ist frei: gestapelt wird genau übereinander, damit unten am
 *     Rand von jedem Stein ein Streifen sichtbar bleibt
 *   - jedes Level hat genau 144 Steine (72 Paare) wie im ersten Satz
 *
 * Jede Form ist ein kleines Bild aus 8 Zeilen mit je 6 Zeichen ("#" = Stein,
 * "." = frei). Aus den Steinen eines Bildes werden die 144 Steine gestapelt:
 * gleichmäßig hoch, der Rest kommt in die Mitte, damit der Rand glatt bleibt.
 */

// --------------------------------------------------
// AUS EINEM BILD WIRD EIN BRETT
// --------------------------------------------------

// Verteilt die 144 Steine auf die Steine eines Bildes: der Rand bekommt eine
// Ebene, die Mitte bis zu sechs. Dadurch liegt oben eine kleine Kuppe und der
// Umriss der Form bleibt glatt.
function verteileEbenen(zellen, mitteX, mitteY) {
    // Sechs Ebenen reichen für dichte Formen; dünne Formen brauchen mehr.
    const HOECHSTE_EBENEN = Math.max(6, Math.ceil(STEINE_PRO_LEVEL / zellen.length));
    const rest = STEINE_PRO_LEVEL - zellen.length;

    const abstaende = zellen.map(zelle =>
        Math.hypot(zelle.x - mitteX, zelle.y - mitteY)
    );
    const weitDraussen = Math.max(...abstaende) || 1;

    // Gewicht: in der Mitte viel, am Rand fast nichts.
    const gewichte = abstaende.map(abstand =>
        Math.pow(Math.max(0, 1 - abstand / weitDraussen), 1.5) + 0.02
    );
    const summe = gewichte.reduce((a, b) => a + b, 0);

    const anteil = gewichte.map(gewicht => (rest * gewicht) / summe);
    const zusatz = anteil.map(wert => Math.min(HOECHSTE_EBENEN - 1, Math.floor(wert)));

    let offen = rest - zusatz.reduce((a, b) => a + b, 0);

    // Was übrig bleibt, bekommen die Steine, die am ehesten noch eine Ebene
    // vertragen – so geht die Zahl exakt auf 144 auf.
    while (offen > 0) {
        let bester = -1;
        let besterRest = -Infinity;

        anteil.forEach((wert, index) => {
            if (zusatz[index] >= HOECHSTE_EBENEN - 1) return;

            const uebrig = wert - zusatz[index];

            if (uebrig > besterRest) {
                besterRest = uebrig;
                bester = index;
            }
        });

        if (bester < 0) {
            throw new Error("Die 144 Steine passen nicht in dieses Bild.");
        }

        zusatz[bester]++;
        offen--;
    }

    return zusatz.map(wert => 1 + wert);
}

function formAusBild(level, name, bild) {
    const zeilen = bild.map(zeile => zeile.trim());
    const zellen = [];

    zeilen.forEach((zeile, y) => {
        if (zeile.length !== 6) {
            throw new Error("Form " + level + " (" + name + "): Zeile " + (y + 1) + " ist " + zeile.length + " statt 6 Zeichen breit.");
        }

        [...zeile].forEach((zeichen, x) => {
            if (zeichen === "#") zellen.push({ x, y });
        });
    });

    if (zellen.length === 0) {
        throw new Error("Form " + level + " (" + name + ") hat keine Steine.");
    }

    // Der Rand der Form bleibt flach (eine Ebene), zur Mitte hin wird
    // aufgetürmt – so bleibt die Form in der Draufsicht klar erkennbar.
    const mitteX = zeilen[0].length / 2 - 0.5;
    const mitteY = zeilen.length / 2 - 0.5;

    const hoehen = verteileEbenen(zellen, mitteX, mitteY);

    const positionen = [];

    zellen.forEach((zelle, index) => {
        for (let z = 0; z < hoehen[index]; z++) {
            positionen.push({ x: zelle.x, y: zelle.y, z });
        }
    });

    const ebenen = [];
    const hoechste = Math.max(...hoehen);

    for (let z = 0; z < hoechste; z++) {
        ebenen.push(positionen.filter(position => position.z === z));
    }

    return {
        id: "trashygecko-mahjong-2-" + level,
        level,
        name,
        columns: 6,
        rows: zeilen.length,
        tileCount: positionen.length,
        layers: ebenen,
        positions: positionen
    };
}

// --------------------------------------------------
// DIE 56 BILDER
// --------------------------------------------------

const FORMEN_SATZ2 = [
    ["Herz", [
        "......",
        ".##.##",
        "######",
        "######",
        "######",
        ".####.",
        "..##..",
        "......"
    ]],
    ["Raute", [
        "......",
        "..##..",
        ".####.",
        "######",
        "######",
        ".####.",
        "..##..",
        "......"
    ]],
    ["Kreis", [
        "..##..",
        ".####.",
        "######",
        "######",
        "######",
        "######",
        ".####.",
        "..##.."
    ]],
    ["Ring", [
        ".####.",
        "##..##",
        "#....#",
        "#....#",
        "#....#",
        "#....#",
        "##..##",
        ".####."
    ]],
    ["Rahmen", [
        "######",
        "#....#",
        "#....#",
        "#....#",
        "#....#",
        "#....#",
        "#....#",
        "######"
    ]],
    ["Dreieck", [
        "..##..",
        "..##..",
        ".####.",
        ".####.",
        "######",
        "######",
        "######",
        "######"
    ]],
    ["Stern", [
        "..##..",
        "..##..",
        "######",
        "######",
        ".####.",
        "..##..",
        ".#..#.",
        "#....#"
    ]],
    ["Kreuz", [
        "..##..",
        "..##..",
        "######",
        "######",
        "######",
        "..##..",
        "..##..",
        "..##.."
    ]],
    ["X", [
        "##..##",
        "##..##",
        ".####.",
        "..##..",
        "..##..",
        ".####.",
        "##..##",
        "##..##"
    ]],
    ["Haken", [
        "....##",
        "....##",
        "...###",
        "..###.",
        ".###..",
        "###...",
        "##....",
        "#....."
    ]],
    ["Pfeil hoch", [
        "..##..",
        ".####.",
        "######",
        "..##..",
        "..##..",
        "..##..",
        "..##..",
        "..##.."
    ]],
    ["Pfeil rechts", [
        "..##..",
        "...##.",
        "######",
        "######",
        "######",
        "...##.",
        "..##..",
        "......"
    ]],
    ["Sanduhr", [
        "######",
        "######",
        ".####.",
        "..##..",
        "..##..",
        ".####.",
        "######",
        "######"
    ]],
    ["Blitz", [
        "...###",
        "..###.",
        ".###..",
        "######",
        "######",
        "..###.",
        ".###..",
        "##...."
    ]],
    ["Wolke", [
        "......",
        ".####.",
        "######",
        "######",
        "######",
        ".####.",
        "..##..",
        "......"
    ]],
    ["Sonne", [
        "..##..",
        "#.##.#",
        ".####.",
        "######",
        "######",
        ".####.",
        "#.##.#",
        "..##.."
    ]],
    ["Mond", [
        "..###.",
        ".##..#",
        "##..##",
        "##....",
        "##....",
        "##..##",
        ".##..#",
        "..###."
    ]],
    ["Tropfen", [
        "..##..",
        "..##..",
        ".####.",
        "######",
        "######",
        "######",
        ".####.",
        "..##.."
    ]],
    ["Blume", [
        "..##..",
        ".####.",
        "##..##",
        ".####.",
        "..##..",
        "..##..",
        ".####.",
        "######"
    ]],
    ["Baum", [
        "..##..",
        ".####.",
        "######",
        ".####.",
        "######",
        "..##..",
        "..##..",
        ".####."
    ]],
    ["Tanne", [
        "..##..",
        ".####.",
        "..##..",
        "######",
        ".####.",
        "######",
        "..##..",
        "..##.."
    ]],
    ["Pilz", [
        ".####.",
        "######",
        "######",
        ".####.",
        "..##..",
        "..##..",
        ".####.",
        "......"
    ]],
    ["Haus", [
        "..##..",
        ".####.",
        "######",
        "######",
        "######",
        "######",
        "##..##",
        "##..##"
    ]],
    ["Burg", [
        "#.##.#",
        "######",
        "######",
        "######",
        "######",
        "#.##.#",
        "#.##.#",
        "######"
    ]],
    ["Turm", [
        "..##..",
        ".####.",
        ".####.",
        ".####.",
        ".####.",
        ".####.",
        ".####.",
        "######"
    ]],
    ["Bruecke", [
        "######",
        "######",
        "#....#",
        "#....#",
        "#....#",
        "#....#",
        "#....#",
        "#....#"
    ]],
    ["Leiter", [
        "##..##",
        "##..##",
        "######",
        "##..##",
        "##..##",
        "######",
        "##..##",
        "##..##"
    ]],
    ["Anker", [
        "..##..",
        ".####.",
        "..##..",
        "..##..",
        "######",
        "..##..",
        "#.##.#",
        ".####."
    ]],
    ["Schluessel", [
        ".####.",
        ".#..#.",
        ".####.",
        "...#..",
        "...#..",
        ".###..",
        ".#....",
        ".##..."
    ]],
    ["Glocke", [
        "..##..",
        ".####.",
        "######",
        "######",
        "######",
        "######",
        "######",
        "..##.."
    ]],
    ["Flasche", [
        "..##..",
        "..##..",
        ".####.",
        ".####.",
        "######",
        "######",
        "######",
        "######"
    ]],
    ["Tasse", [
        "######",
        "######",
        "######",
        "######",
        ".####.",
        "..##..",
        ".####.",
        "......"
    ]],
    ["Kanne", [
        "..#...",
        ".##...",
        "######",
        "######",
        "######",
        "######",
        "######",
        ".####."
    ]],
    ["Krone", [
        "#.##.#",
        "#.##.#",
        "######",
        "######",
        "######",
        "######",
        ".####.",
        "######"
    ]],
    ["Schleife", [
        "##..##",
        "######",
        "######",
        ".####.",
        "..##..",
        ".####.",
        "######",
        "##..##"
    ]],
    ["Fisch", [
        "......",
        ".####.",
        "######",
        "##.###",
        "######",
        ".####.",
        "..#...",
        "......"
    ]],
    ["Vogel", [
        "..###.",
        ".#####",
        "######",
        "######",
        ".####.",
        "..##..",
        "..#...",
        ".#...."
    ]],
    ["Katze", [
        "#....#",
        "##..##",
        "######",
        "######",
        "######",
        "######",
        ".####.",
        "..##.."
    ]],
    ["Pfote", [
        ".#..#.",
        "#....#",
        "......",
        ".####.",
        "######",
        "######",
        ".####.",
        "..##.."
    ]],
    ["Schnecke", [
        ".####.",
        "##..##",
        "#.##.#",
        "#.##.#",
        "##..##",
        ".####.",
        "...##.",
        "######"
    ]],
    ["Schmetterling", [
        "##..##",
        "######",
        ".####.",
        "..##..",
        "..##..",
        ".####.",
        "######",
        "##..##"
    ]],
    ["Marienkaefer", [
        "..##..",
        ".####.",
        "######",
        "#.##.#",
        "#.##.#",
        "######",
        ".####.",
        "..##.."
    ]],
    ["Auto", [
        "......",
        ".####.",
        "######",
        "#.##.#",
        "######",
        "######",
        "#.##.#",
        "......"
    ]],
    ["Segelboot", [
        "...##.",
        "..###.",
        ".####.",
        "######",
        "......",
        "######",
        "#....#",
        "######"
    ]],
    ["Flugzeug", [
        "..##..",
        "..##..",
        "######",
        "######",
        "######",
        "..##..",
        "..##..",
        ".####."
    ]],
    ["Rakete", [
        "..##..",
        ".####.",
        ".####.",
        ".####.",
        "######",
        ".####.",
        ".#..#.",
        "#....#"
    ]],
    ["Ufo", [
        "..##..",
        "######",
        "######",
        "#....#",
        "######",
        "######",
        ".####.",
        "..##.."
    ]],
    ["Geist", [
        ".####.",
        "######",
        "######",
        "######",
        "######",
        "######",
        "######",
        "#.##.#"
    ]],
    ["Roboter", [
        ".#..#.",
        "######",
        "#.##.#",
        "######",
        "######",
        ".####.",
        "#.##.#",
        "#....#"
    ]],
    ["Fernseher", [
        "..##..",
        ".####.",
        "######",
        "#....#",
        "#....#",
        "######",
        "######",
        "..##.."
    ]],
    ["Brief", [
        "######",
        "##..##",
        "#.##.#",
        "#.##.#",
        "##..##",
        "######",
        "######",
        "######"
    ]],
    ["Fahne", [
        "##....",
        "######",
        "######",
        "##....",
        "##....",
        "##....",
        "##....",
        "##...."
    ]],
    ["Buch", [
        "......",
        "######",
        "#.##.#",
        "######",
        "######",
        "#.##.#",
        "######",
        "......"
    ]],
    ["Kerze", [
        "..##..",
        ".####.",
        "..##..",
        ".####.",
        ".####.",
        ".####.",
        ".####.",
        "######"
    ]],
    ["Labyrinth", [
        "######",
        "#.##.#",
        "#.##.#",
        "#.##.#",
        "#.##.#",
        "#.##.#",
        "#.##.#",
        "######"
    ]],
    ["Block", [
        "######",
        "######",
        "######",
        "######",
        "######",
        "######",
        "######",
        "######"
    ]]
];

// --------------------------------------------------
// DER ZWEITE SATZ
// --------------------------------------------------

const BOARD_LAYOUTS_2 = {};

FORMEN_SATZ2.forEach(([name, bild], index) => {
    BOARD_LAYOUTS_2[index + 1] = formAusBild(index + 1, name, bild);
});

// Name einer Form des zweiten Satzes (fuer Tooltips und Titel).
function formNameSatz2(level) {
    const form = ((level - 1) % FORMEN_SATZ2.length) + 1;

    return FORMEN_SATZ2[form - 1][0];
}
