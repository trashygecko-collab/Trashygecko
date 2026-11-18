/*
 * TRASHYGECKO MAHJONG
 * Schwere Formen – zwei eigene Sammlungen
 *
 * Der schwere Modus (36 Motive zu je vier Steinen) bekommt eigene Spielfelder.
 * Es gibt zwei Sammlungen, die unabhaengig voneinander arbeiten:
 *
 *   BOARD_LAYOUTS_SCHWER        PC-Level: 12 Spalten, 8 Reihen,
 *                               hoechstens vier Ebenen (z 0..3) - genau der
 *                               Bereich, den die PC-Formen aus
 *                               board-layout.js belegen (12 x 8, maxZ 3).
 *
 *   BOARD_LAYOUTS_SCHWER_HANDY  Handy-Level: 8 Spalten, bis zu 10 Reihen,
 *                               hoechstens fuenf Ebenen (z 0..4).  25 Umrisse
 *                               zu je vier Profilen ergeben 100 Felder.
 *
 * Welche Sammlung gilt, entscheidet der Formen-Umschalter (FORMEN_SATZ):
 * Satz 1 (PC-LEVEL) nimmt die PC-Sammlung, Satz 2 (HANDY-LEVEL) die
 * Handy-Sammlung - unabhaengig von der Fensterbreite.
 *
 * Gebaut wird jede Form aus zwei Teilen:
 *   - dem Umriss (ein Bild aus 12 Spalten bzw. 8 Spalten, "#" = Stein)
 *   - einem Hoehenprofil, das bestimmt, wo die Steine uebereinander liegen
 * Jede Zelle eines Umrisses traegt mindestens eine Ebene; die restlichen der
 * 144 Steine kommen nach Profil nach oben.  So geht die Steinzahl immer genau
 * auf, jede Saeule ist von unten durchgehend gefuellt und kein Stein schwebt.
 */

// --------------------------------------------------
// GERAETEERKENNUNG
// --------------------------------------------------

// Dieselbe Grenze wie im CSS: Handy und kleine Tablets bis 900 px Breite,
// darueber Rechner und grosse Tablets.  Kein eigenes Mass, nur die
// vorhandene Aufteilung des Layouts.
function istKleinesGeraet() {
    if (typeof window === "undefined" || typeof window.matchMedia !== "function") {
        return false;
    }

    return window.matchMedia("(max-width: 900px)").matches;
}

// --------------------------------------------------
// HOEHENPROFILE
// --------------------------------------------------

// Jedes Profil liefert ein Gewicht je Zelle: grosse Werte bekommen mehr
// Ebenen.  Die vier Profile verteilen die Steine ganz unterschiedlich -
// Kuppe in der Mitte, Grat in der Mitte, Treppe von links oben nach rechts
// unten, oder umgekehrt ein hoher Rand mit flacher Mitte.
const SCHWER_PROFILE = {
    kuppe(x, y, breite, hoehe) {
        const cx = (breite - 1) / 2;
        const cy = (hoehe - 1) / 2;
        const weit = Math.max(1, Math.hypot(cx, cy));
        const d = Math.hypot(x - cx, y - cy) / weit;

        return Math.pow(Math.max(0, 1 - d), 1.4) + 0.02;
    },

    grat(x, y, breite, hoehe) {
        const cx = (breite - 1) / 2;
        const d = Math.abs(x - cx) / Math.max(1, cx);

        // Der Grat laeuft senkrecht durch die Mitte und wird nach oben breiter.
        return Math.pow(Math.max(0, 1 - d), 1.2) + 0.3 * (y / Math.max(1, hoehe - 1)) + 0.02;
    },

    terrasse(x, y, breite, hoehe) {
        const nachRechts = x / Math.max(1, breite - 1);
        const nachUnten = y / Math.max(1, hoehe - 1);

        return Math.pow(nachRechts * 0.62 + nachUnten * 0.38 + 0.05, 1.7) + 0.02;
    },

    mulde(x, y, breite, hoehe) {
        const cx = (breite - 1) / 2;
        const cy = (hoehe - 1) / 2;
        const weit = Math.max(1, Math.hypot(cx, cy));

        // Umkehrung der Kuppe: der Rand liegt hoch, die Mitte bleibt flach.
        return Math.hypot(x - cx, y - cy) / weit + 0.02;
    }
};

const SCHWER_PROFIL_NAMEN = {
    kuppe: "Kuppe",
    grat: "Grat",
    terrasse: "Terrassen",
    mulde: "Mulde"
};

// --------------------------------------------------
// AUS UMRISS UND PROFIL WIRD EIN FELD
// --------------------------------------------------

// Verteilt die 144 Steine auf die Zellen eines Umrisses.  "hoechste" ist die
// hoechste erlaubte Ebene (also z = hoechste - 1); darueber wird nie gestapelt.
//
// Die oberste Reihe (y = 0) bleibt immer flach: Ein Stein, der dort hoeher
// liegt, ragt nach oben aus dem Feld heraus (der Versatz je Ebene geht nach
// oben) und wuerde abgeschnitten.  Gestapelt wird deshalb nur ab der zweiten
// Reihe - die Form bleibt so vollstaendig im Bild.
function schwerHoehen(zellen, gewicht, hoechste, name) {
    const rest = STEINE_PRO_LEVEL - zellen.length;

    if (rest < 0) {
        throw new Error("Schwere Form " + name + " hat mehr als 144 Zellen.");
    }

    const gewichte = zellen.map(zelle =>
        zelle.y > 0 ? Math.max(0.02, gewicht(zelle.x, zelle.y)) : 0
    );
    const summe = gewichte.reduce((a, b) => a + b, 0);

    const roh = gewichte.map(wert => (rest * wert) / summe);
    const zusatz = roh.map(wert => Math.min(hoechste - 1, Math.floor(wert)));

    let offen = rest - zusatz.reduce((a, b) => a + b, 0);

    if (offen < 0) {
        throw new Error("Schwere Form " + name + ": zu viele Ebenen vergeben.");
    }

    // Der Rest geht reihum an die Zellen mit dem groessten Nachkommaanteil,
    // die noch eine Ebene vertragen - so geht die Zahl exakt auf 144 auf.
    const reihe = zellen
        .map((_, index) => index)
        .filter(index => zusatz[index] < hoechste - 1 && zellen[index].y > 0)
        .sort((a, b) => (roh[b] - zusatz[b]) - (roh[a] - zusatz[a]));

    let schritt = 0;

    while (offen > 0) {
        if (reihe.length === 0) {
            throw new Error(
                "Schwere Form " + name + ": zu wenige Zellen unter der obersten Reihe fuer 144 Steine."
            );
        }

        const index = reihe[schritt % reihe.length];

        if (zusatz[index] < hoechste - 1) {
            zusatz[index]++;
            offen--;
        }

        schritt++;

        if (schritt > 20000) {
            throw new Error("Schwere Form " + name + ": Ebenen gehen nicht auf.");
        }
    }

    return zusatz.map(wert => 1 + wert);
}

// Baut aus einem Umriss (Zeilen aus "#" und ".") und einem Profilnamen ein
// fertiges Feld.  Die Ebenen entstehen von unten nach oben; jede Saeule ist
// lueckenlos gefuellt.
function schwerForm(level, name, bild, profilName, breite, hoechste) {
    const zeilen = bild.map(zeile => zeile.trim());
    const zellen = [];

    zeilen.forEach((zeile, y) => {
        if (zeile.length !== breite) {
            throw new Error(
                "Schwere Form " + name + ": Zeile " + (y + 1) + " ist " +
                zeile.length + " statt " + breite + " Zeichen breit."
            );
        }

        [...zeile].forEach((zeichen, x) => {
            if (zeichen === "#") zellen.push({ x, y });
        });
    });

    if (zellen.length === 0) {
        throw new Error("Schwere Form " + name + " hat keine Steine.");
    }

    const gewicht = SCHWER_PROFILE[profilName];

    if (!gewicht) {
        throw new Error("Unbekanntes Profil: " + profilName);
    }

    const hoehen = schwerHoehen(
        zellen,
        (x, y) => gewicht(x, y, breite, zeilen.length),
        hoechste,
        name
    );

    const positionen = [];

    zellen.forEach((zelle, index) => {
        for (let z = 0; z < hoehen[index]; z++) {
            positionen.push({ x: zelle.x, y: zelle.y, z });
        }
    });

    const ebenen = [];
    const hoechsteEbene = Math.max(...hoehen);

    for (let z = 0; z < hoechsteEbene; z++) {
        ebenen.push(positionen.filter(position => position.z === z));
    }

    return {
        id: "trashygecko-mahjong-schwer-" + (breite === 6 ? "handy-" : "") + level,
        level,
        name: name + " · " + SCHWER_PROFIL_NAMEN[profilName],
        columns: breite,
        rows: zeilen.length,
        tileCount: positionen.length,
        layers: ebenen,
        positions: positionen
    };
}

// --------------------------------------------------
// PC: 14 UMRISSE, JE VIER PROFILE = 56 FELDER
// --------------------------------------------------

const SCHWER_UMRISSE_PC = [
    ["Zitadelle", [
        "..##....##..",
        ".####..####.",
        "############",
        "##..####..##",
        "##..####..##",
        "############",
        ".####..####.",
        "..##....##.."
    ]],
    ["Zickzack", [
        "###.....####",
        "###.....####",
        "##########..",
        "..#########.",
        "..#########.",
        "####....####",
        "####....####",
        "####....####"
    ]],
    ["Doppelturm", [
        ".####..####.",
        ".####..####.",
        ".##########.",
        ".####..####.",
        ".####..####.",
        "..########..",
        "...######...",
        "....####...."
    ]],
    ["Treppe", [
        "##..........",
        "####........",
        "######......",
        "########....",
        "##########..",
        "############",
        "############",
        "############"
    ]],
    ["Sanduhr", [
        "############",
        ".##########.",
        "..########..",
        "...######...",
        "...######...",
        "..########..",
        ".##########.",
        "############"
    ]],
    ["Kreuz", [
        "...####.....",
        "...####.....",
        ".##########.",
        ".##########.",
        ".##########.",
        "...####.....",
        "...####.....",
        "...####....."
    ]],
    ["Hufeisen", [
        ".##......##.",
        ".##......##.",
        ".##......##.",
        ".##########.",
        ".##########.",
        ".##########.",
        ".##########.",
        "............"
    ]],
    ["Winkel", [
        "####........",
        "#####.......",
        "######......",
        "#######.....",
        "########....",
        "#######.....",
        "######......",
        "#####......."
    ]],
    ["Graben", [
        "##########..",
        "##########..",
        "##......##..",
        "##......##..",
        "##......##..",
        "##########..",
        "##########..",
        "............"
    ]],
    ["Kamm", [
        "#.#.#.#.#.#.",
        "############",
        "#.#.#.#.#.#.",
        "############",
        "#.#.#.#.#.#.",
        "############",
        "............",
        "............"
    ]],
    ["Vase", [
        "..########..",
        ".##########.",
        ".##########.",
        "..########..",
        "...######...",
        "....####....",
        "....####....",
        "...######..."
    ]],
    ["Labyrinth", [
        "###########.",
        "#.........#.",
        "#.#######.#.",
        "#.#.....#.#.",
        "#.#.###.#.#.",
        "#.#.....#.#.",
        "#.#######.#.",
        "#..........."
    ]],
    ["Zwillingsbogen", [
        ".####..####.",
        ".####..####.",
        "##########..",
        "##########..",
        ".####..####.",
        ".####..####.",
        "..########..",
        "...######..."
    ]],
    ["Amboss", [
        "############",
        "############",
        "....####....",
        "....####....",
        "....####....",
        "..########..",
        "..########..",
        "............"
    ]]
];

// --------------------------------------------------
// --------------------------------------------------
// --------------------------------------------------
// --------------------------------------------------
// HANDY: 25 UMRISSE (8 x 10), JE VIER PROFILE = 100 FELDER
// --------------------------------------------------

// Die Handy-Masse sind 8 Spalten und bis zu 10 Reihen - dieselbe Flaeche, auf
// der auch die Handy-Level im leichten Modus liegen.  Jeder Umriss traegt eine
// Saeule je "#", die Hoehe kommt aus dem Profil (hoechstens vier Ebenen).
const SCHWER_UMRISSE_HANDY = [
    ["Herz", [
        "...##...",
        ".######.",
        "########",
        "########",
        ".######.",
        "..####..",
        "..####..",
        "...##..."
    ]],
    ["Stern", [
        "...##...",
        ".######.",
        "########",
        "####.###",
        "########",
        ".######.",
        "...##..."
    ]],
    ["Baum", [
        "...##...",
        "..####..",
        ".######.",
        "########",
        ".######.",
        "..####..",
        ".######.",
        "##.##.##"
    ]],
    ["Haus", [
        "...##...",
        "..####..",
        ".######.",
        "########",
        "##....##",
        "##.##.##",
        "##.##.##",
        "########"
    ]],
    ["Fisch", [
        "..#####.",
        ".#######",
        "########",
        "########",
        ".#####.#",
        "..######",
        "..####.."
    ]],
    ["Vogel", [
        "..###...",
        ".######.",
        "########",
        "########",
        ".######.",
        "..####..",
        ".######."
    ]],
    ["Katze", [
        "##....##",
        "########",
        "########",
        ".######.",
        ".######.",
        "##.##.##",
        "##....##"
    ]],
    ["Hund", [
        "##....##",
        "########",
        "########",
        "########",
        ".######.",
        "##.##.##",
        ".######."
    ]],
    ["Schmetterling", [
        "##....##",
        "########",
        "########",
        ".######.",
        ".######.",
        "########",
        "##....##"
    ]],
    ["Blume", [
        "...##...",
        "..####..",
        ".######.",
        "########",
        ".######.",
        "...##...",
        "##.##.##",
        ".######."
    ]],
    ["Pilz", [
        "..####..",
        ".######.",
        "########",
        "########",
        "..####..",
        "..####..",
        "..####..",
        ".######."
    ]],
    ["Krone", [
        "##.##.##",
        "########",
        "########",
        ".######.",
        ".######.",
        "########"
    ]],
    ["Kelch", [
        "########",
        "########",
        "########",
        ".######.",
        "...##...",
        "...##...",
        "..####..",
        ".######."
    ]],
    ["Rakete", [
        "...##...",
        "..####..",
        "..####..",
        ".######.",
        ".######.",
        "########",
        ".######.",
        "..####..",
        "...##..."
    ]],
    ["Segelboot", [
        "...##...",
        "..####..",
        ".#####..",
        "#######.",
        ".######.",
        "########",
        "########",
        ".######."
    ]],
    ["Auto", [
        "..####..",
        ".######.",
        "########",
        "########",
        ".######.",
        "##.##.##",
        "########"
    ]],
    ["Gitarre", [
        ".######.",
        "########",
        ".######.",
        ".######.",
        "..####..",
        "..####..",
        ".######.",
        "########"
    ]],
    ["Kamera", [
        "..####..",
        ".######.",
        "########",
        "########",
        "##.##.##",
        "########",
        "########"
    ]],
    ["Tasse", [
        ".######.",
        ".######.",
        ".######.",
        ".######.",
        "########",
        ".######.",
        "..####..",
        "########"
    ]],
    ["Kerze", [
        "..####..",
        "########",
        "########",
        ".######.",
        ".######.",
        "########",
        "..####.."
    ]],
    ["Mond", [
        ".######.",
        "########",
        "#####...",
        "####....",
        "####....",
        "#####...",
        "########",
        ".######."
    ]],
    ["Wolke", [
        "..####..",
        ".######.",
        "########",
        "########",
        "########",
        ".######.",
        "..####.."
    ]],
    ["Burg", [
        "#.####.#",
        "########",
        "########",
        "########",
        "##.##.##",
        "##.##.##",
        "########"
    ]],
    ["Leiter", [
        "###..###",
        "###..###",
        "########",
        "###..###",
        "###..###",
        "########",
        "###..###",
        "########"
    ]],
    ["Wuerfel", [
        "########",
        "########",
        "##.##.##",
        "########",
        "##.##.##",
        "########",
        "########"
    ]]
];

// --------------------------------------------------
// DIE BEIDEN SAMMLUNGEN
// --------------------------------------------------

const SCHWER_PROFIL_FOLGE = ["kuppe", "grat", "terrasse", "mulde"];

function baueSammlung(umrisse, breite, hoechste) {
    const sammlung = {};

    umrisse.forEach(([name, bild], index) => {
        const level = index * SCHWER_PROFIL_FOLGE.length + 1;

        SCHWER_PROFIL_FOLGE.forEach((profil, stelle) => {
            sammlung[level + stelle] = schwerForm(
                level + stelle,
                name,
                bild,
                profil,
                breite,
                hoechste
            );
        });
    });

    return sammlung;
}

// 56 Felder fuer den Rechner (12 x 8, bis Ebene 4).
const BOARD_LAYOUTS_SCHWER = baueSammlung(SCHWER_UMRISSE_PC, 12, 4);

// 100 Felder fuer das Handy (8 x 10, bis Ebene 5 - die Handy-Masse).
const BOARD_LAYOUTS_SCHWER_HANDY = baueSammlung(SCHWER_UMRISSE_HANDY, 8, 5);

// --------------------------------------------------
// NAME DER FORM (fuer Hinweise und die Level-Liste)
// --------------------------------------------------

function formNameSchwer(level, klein = (typeof FORMEN_SATZ === "number" ? FORMEN_SATZ === 2 : istKleinesGeraet())) {
    const sammlung = klein ? BOARD_LAYOUTS_SCHWER_HANDY : BOARD_LAYOUTS_SCHWER;
    const felder = Object.keys(sammlung).map(Number).sort((a, b) => a - b);
    const form = felder[(Math.max(1, level) - 1) % felder.length];

    return sammlung[form] ? sammlung[form].name : "";
}
