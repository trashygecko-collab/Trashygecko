/* Anzeige der Gewinnsprüche im Mahjong. Die Liste selbst steht seit dem
   25.11.2026 zentral in spiele/gewonnen.js (vor dieser Datei einbinden). */

let gewonnenVorrat = [];
let letzterGewonnen = "";

function gewonnenSpruch() {
    if (gewonnenVorrat.length === 0) {
        gewonnenVorrat = GEWONNEN_SPRUECHE.slice();

        for (let i = gewonnenVorrat.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [gewonnenVorrat[i], gewonnenVorrat[j]] = [gewonnenVorrat[j], gewonnenVorrat[i]];
        }
    }

    let spruch = gewonnenVorrat.pop();

    if (spruch === letzterGewonnen && gewonnenVorrat.length > 0) {
        const tausch = gewonnenVorrat.pop();
        gewonnenVorrat.push(spruch);
        spruch = tausch;
    }

    letzterGewonnen = spruch;

    return spruch;
}

// Zeigt den Spruch unter der Siegmeldung
function zeigeGewonnenSpruch() {
    const feld = document.getElementById("gewonnenText");

    if (!feld) return;

    feld.textContent = gewonnenSpruch();
    feld.style.color = siegSpruchFarbe();
}

// Beim Start eines neuen Bretts wieder leeren
function leereGewonnenSpruch() {
    const feld = document.getElementById("gewonnenText");

    if (feld) feld.textContent = "";
}

// --------------------------------------------------
// SONDERZEICHEN ZWISCHEN SPRUCH UND WEITER-KNOPF
// --------------------------------------------------

const GEWONNEN_SYMBOLE = [
    "👑", "🏆", "🥇", "🎯", "☆", "💎", "✨", "✔", "⚡", "🔥",
    "❤", "💖", "👏", "👍", "💯", "🚀", "💥", "💫", "🔮"
];

let symbolVorrat = [];
let letztesSymbol = "";

function gewonnenSymbol() {
    if (symbolVorrat.length === 0) {
        symbolVorrat = GEWONNEN_SYMBOLE.slice();

        for (let i = symbolVorrat.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [symbolVorrat[i], symbolVorrat[j]] = [symbolVorrat[j], symbolVorrat[i]];
        }
    }

    let symbol = symbolVorrat.pop();

    if (symbol === letztesSymbol && symbolVorrat.length > 0) {
        const tausch = symbolVorrat.pop();
        symbolVorrat.push(symbol);
        symbol = tausch;
    }

    letztesSymbol = symbol;

    return symbol;
}

function zeigeGewonnenSymbol() {
    const feld = document.getElementById("gewonnenSymbol");

    if (feld) feld.textContent = gewonnenSymbol();
}

function leereGewonnenSymbol() {
    const feld = document.getElementById("gewonnenSymbol");

    if (feld) feld.textContent = "";
}
