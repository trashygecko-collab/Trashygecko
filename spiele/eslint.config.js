// Minimale Pruefung fuer die Mahjong-Skripte.
//
// Die Spiele sind klassische Skripte ohne Module: Funktionen und Zustaende
// liegen bewusst global, damit die Dateien sich gegenseitig sehen. Deshalb ist
// no-undef hier aus - es wuerde jede Uebergabe zwischen den Dateien als Fehler
// melden. Die Regeln unten finden dagegen die Fehler, die hier wirklich
// vorkamen: doppelte Namen, Schattenvariablen, loses Vergleichen, var,
// stille catch-Bloecke.
const regeln = {
    eqeqeq: "error",
    "no-var": "error",
    "no-empty": ["error", { allowEmptyCatch: false }],
    "no-shadow": "error",
    "no-redeclare": "error",
    "no-dupe-keys": "error",
    "no-duplicate-case": "error",
    "no-unreachable": "error",
    // prefer-const kann hier nicht greifen: der Zustand wird ueber mehrere
    // Dateien hinweg geschrieben (let in main.js, Zuweisung in hud.js).
    "prefer-const": "off",
    // Nur lokale Variablen: globale Funktionen und Zustaende sind hier normal.
    "no-unused-vars": ["warn", { vars: "local", args: "after-used", caughtErrors: "none" }],
};

module.exports = [
    {
        files: ["mahjong/js/**/*.js", "lob.js"],
        languageOptions: {
            ecmaVersion: 2022,
            sourceType: "script",
        },
        rules: regeln,
    },
];
