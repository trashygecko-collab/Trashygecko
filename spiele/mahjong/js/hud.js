/*
 * TRASHYGECKO MAHJONG
 * Anzeige: Punkte, Zeit, Meldungen und Sprueche
 *
 * Aus main.js ausgelagert. Alle Teile haengen an globalen Zustaenden
 * (score, gameSeconds, gameData), die weiterhin in main.js stehen.
 */

// Was ein Tipp kostet und ab wann es ueberhaupt einen gibt.
const TIPP_KOSTEN = 200;

// Kraeftige Farben fuer den Spruch nach einem Paar: satt genug, um ueber dem
// Brett aufzufallen, und hell genug, um auf dem dunklen Feld zu tragen.
const SPRUCH_FARBEN = [
    "#ffc22e",   // Gold
    "#4ddc86",   // Mint
    "#4fb6ff",   // Himmelblau
    "#ff6f9c",   // Rosa
    "#b48cff",   // Lavendel
    "#ffb340",   // Sand
    "#2fd6c6",   // Aqua
    "#ff8f5e"    // Apricot
];

// Warme Farben fuer den Spruch auf der Siegtafel - nie zweimal dieselbe.
// Kraeftiger als frueher (eine Stufe tiefer und dadurch mehr Farbkraft), aber
// weiter hell genug: auf der Tafel bleibt jede Farbe bei mindestens 5.4:1.
const SIEG_SPRUCH_FARBEN = [
    "#ff9d33",   // Orange
    "#ffc94c",   // Gold
    "#ffa82d",   // Amber
    "#ff8b54",   // Coral
    "#ff758e",   // Rosa
    "#ffb854"    // Honig
];

// Zieht eine Farbe aus der Liste, aber nie zweimal dieselbe hintereinander.
function farbwahl(farben) {
    let letzte = "";

    return () => {
        const auswahl = farben.filter(farbe => farbe !== letzte);

        letzte = auswahl[Math.floor(Math.random() * auswahl.length)];

        return letzte;
    };
}

const spruchFarbe = farbwahl(SPRUCH_FARBEN);
const siegSpruchFarbe = farbwahl(SIEG_SPRUCH_FARBEN);

// --------------------------------------------------
// TIMER
// --------------------------------------------------

// Die Zeit kommt aus der Uhr, nicht aus der Zahl der Ticks: ein gebremster
// oder gebuendelter Timer (Hintergrund-Tab, Lastspitze) verfaelscht damit
// weder Spielzeit noch Bestzeiten.
let timerStart = 0;

function updateStats() {
    // Der eine Level-Waehler zeigt das Level des laufenden Formensatzes.
    setText("levelNumber", String(currentLevel).padStart(2, "0"));

    // Ohne den Satz sieht man in der Taste nur "01" und weiss nicht, ob
    // Handy- oder Normal-Level laufen.
    setText("levelSatz", FORMEN_SATZ === 2 ? "HANDY" : "NORMAL");

    const levelButton = getElement("levelButton");

    if (levelButton) {
        levelButton.setAttribute("aria-label", `Levelauswahl: ${satzName()}`);
    }

    setText(
        "remainingTiles",
        144 - foundPairs * 2
    );

    setText(
        "remainingPairs",
        72 - foundPairs
    );

    setText(
        "gameTime",
        formatTime(gameSeconds)
    );

    setText(
        "scoreDisplay",
        `Punkte: ${score.toLocaleString("de-DE")}`
    );

    // Tipp-Knopf: unter 200 Punkten gibt es keinen Tipp - das sieht man
    // schon am Knopf, nicht erst nach dem Klick.
    const hintButton = getElement("hintButton");

    if (hintButton) {
        const keinTipp = score < TIPP_KOSTEN || !gameRunning;

        hintButton.disabled = keinTipp;
        hintButton.title = keinTipp
            ? `Tipp erst ab ${TIPP_KOSTEN} Punkten`
            : `Markiert ein Paar und kostet ${TIPP_KOSTEN} Punkte`;
    }
}

// rot = true faerbt die Zeile rot (Fehlgriff), sonst normale Farbe.
// punkte: Zusatz wie "+ 100 Punkte" - bleibt grau (bzw. rot beim Fehlgriff),
// damit der Spruch davor allein die Farbe traegt.
function setMessage(message, rot = false, punkte = "") {
    const element = getElement("gameStatus");

    if (!element) return;

    element.classList.toggle("fehler", rot);

    const text = document.createElement("span");

    text.className = "status-text";
    text.textContent = message;

    if (!rot && punkte) {
        text.style.color = spruchFarbe();
    }

    if (!punkte) {
        element.replaceChildren(text);
        return;
    }

    const zusatz = document.createElement("span");

    zusatz.className = "status-punkte";
    zusatz.textContent = punkte;

    element.replaceChildren(text, zusatz);
}

function formatTime(seconds) {
    const minutes = Math.floor(seconds / 60);
    const rest = seconds % 60;

    return (
        String(minutes).padStart(2, "0") +
        ":" +
        String(rest).padStart(2, "0")
    );
}

function startGameTimer() {
    stopGameTimer();

    timerStart = Date.now() - gameSeconds * 1000;

    gameTimer = setInterval(() => {
        if (!gameRunning) return;

        const sekunden = Math.floor((Date.now() - timerStart) / 1000);

        if (sekunden === gameSeconds) return;

        gameSeconds = sekunden;
        updateStats();
    }, 250);
}

function stopGameTimer() {
    if (gameTimer !== null) {
        clearInterval(gameTimer);
        gameTimer = null;
    }
}

// Setzt die Punkteanzeige sofort wieder auf Gold.
function stoppePunkteBlitz() {
    clearTimeout(blinkePunkte.timer);

    const anzeige = getElement("scoreDisplay");

    if (anzeige) {
        anzeige.classList.remove("punkte-fehler");
    }
}

// Laesst die Punkteanzeige kurz rot aufleuchten (Fehlgriff).
function blinkePunkte() {
    const anzeige = getElement("scoreDisplay");

    if (!anzeige) return;

    anzeige.classList.remove("punkte-fehler");

    // Kurz neu zeichnen lassen, damit die Animation neu startet.
    void anzeige.offsetWidth;

    anzeige.classList.add("punkte-fehler");

    clearTimeout(blinkePunkte.timer);

    blinkePunkte.timer = setTimeout(() => {
        anzeige.classList.remove("punkte-fehler");
    }, 750);
}

// Sprueche aus den gemeinsamen Listen. Fehlt eine Datei (etwa weil sie nicht
// geladen wurde), bleibt ein neutraler Ersatztext stehen - eine fehlende
// Spruchliste darf nie das Spiel anhalten.
function lobText(mitSteinen = true) {
    return typeof lobSpruch === "function" ? lobSpruch(mitSteinen) : "Sehr gut!";
}

function pechText() {
    return typeof pechSpruch === "function" ? pechSpruch() : "Nicht schlimm, weiter!";
}

// Nebenmeldung unter der Statuszeile. Sie ueberschreibt nie den Spruch oder
// die Siegmeldung - beide bleiben gleichzeitig lesbar.
function zeigeHinweis(text) {
    const feld = getElement("speicherHinweis");

    if (!feld) return;

    feld.textContent = text || "";
    feld.hidden = !text;
}

// Speichert den Spielstand und sagt es, wenn der Browser das verweigert
// (voller oder gesperrter Speicher). Ein stiller Fehlschlag waere schlimmer:
// Einstellungen und Fortschritt waeren weg, ohne dass es auffaellt.
function speichereSpielstand(hinweis) {
    if (!gameData) return false;

    if (saveStorage(gameData)) {
        // Ein alter Hinweis gilt nicht mehr, sobald es wieder geht.
        zeigeHinweis("");

        return true;
    }

    zeigeHinweis(
        hinweis || "Der Spielstand lässt sich in diesem Browser nicht speichern."
    );

    return false;
}
