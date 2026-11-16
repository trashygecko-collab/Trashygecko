/*
 * TRASHYGECKO MAHJONG
 * Anzeige: Punkte, Zeit, Meldungen und Sprueche
 *
 * Aus main.js ausgelagert. Alle Teile haengen an globalen Zustaenden
 * (score, gameSeconds, gameData), die weiterhin in main.js stehen.
 */

// Was ein Tipp kostet und ab wann es ueberhaupt einen gibt.
const TIPP_KOSTEN = 200;

// Gut lesbare Farben auf dem dunklen Feld - fuer den Spruch nach einem Paar.
const SPRUCH_FARBEN = [
    "#ffd166",
    "#8fe3ab",
    "#9ad4ff",
    "#ffb3c7",
    "#c9b6ff",
    "#ffe6a8",
    "#8ee8dc",
    "#ffc49a"
];

let letzteSpruchFarbe = "";

// Warme Farben fuer den Spruch auf der Siegtafel - nie zweimal dieselbe.
const SIEG_SPRUCH_FARBEN = [
    "#ffa94d",
    "#ffd166",
    "#ffb347",
    "#ff9c6e",
    "#ff8fa3",
    "#f6c177"
];

let letzteSiegSpruchFarbe = "";

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
    element.textContent = "";

    const text = document.createElement("span");

    text.className = "status-text";
    text.textContent = message;

    if (!rot && punkte) {
        text.style.color = spruchFarbe();
    }

    element.appendChild(text);

    if (punkte) {
        const zusatz = document.createElement("span");

        zusatz.className = "status-punkte";
        zusatz.textContent = " " + punkte;

        element.appendChild(zusatz);
    }
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

// Zufaellige Spruchfarbe, aber nie zweimal dieselbe hintereinander.
function spruchFarbe() {
    const auswahl = SPRUCH_FARBEN.filter(farbe => farbe !== letzteSpruchFarbe);
    const farbe = auswahl[Math.floor(Math.random() * auswahl.length)];

    letzteSpruchFarbe = farbe;

    return farbe;
}

function siegSpruchFarbe() {
    const auswahl = SIEG_SPRUCH_FARBEN.filter(
        farbe => farbe !== letzteSiegSpruchFarbe
    );
    const farbe = auswahl[Math.floor(Math.random() * auswahl.length)];

    letzteSiegSpruchFarbe = farbe;

    return farbe;
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
