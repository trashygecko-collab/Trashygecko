
/*
 * TRASHYGECKO MAHJONG
 * Hauptsteuerung mit Zeit, Punkten und Spielstand
 */

const MAHJONG_SYMBOLS = Array.from(
    { length: 72 },
    (_, index) => `motiv-${String(index + 1).padStart(2, "0")}`
);

let currentBoard = [];
let selectedTile = null;
let foundPairs = 0;
let mistakes = 0;
let gameSeconds = 0;
let gameTimer = null;
let gameRunning = false;
let currentLevel = 1;
let score = 0;
let gameData = null;
let levelWechselTimer = null;
let hinweisTiles = [];

// Zeichnet das bestehende Brett neu, ohne alle Steine zu ersetzen.
function aktualisiereBrett() {
    updateBoardView(currentBoard, selectedTile, hinweisTiles, handleTileClick);
}

// --------------------------------------------------
// HTML-ELEMENTE
// --------------------------------------------------

function getElement(...ids) {
    for (const id of ids) {
        const element = document.getElementById(id);
        if (element) return element;
    }

    return null;
}

function setText(ids, value) {
    const element = getElement(...ids);

    if (element) {
        element.textContent = value;
    }
}

// Fehlende Anzeigen automatisch ergänzen
function createGameInformation() {
    const board = getElement("board");

    if (!board || !board.parentElement) return;

    if (!getElement("gameStatus")) {
        const status = document.createElement("div");

        status.id = "gameStatus";
        status.style.cssText =
            "text-align:center;color:#f5edda;font-size:18px;" +
            "font-weight:bold;margin:12px auto;min-height:28px;";

        board.parentElement.insertBefore(status, board);
    }

    if (!getElement("scoreDisplay")) {
        const scoreBox = document.createElement("div");

        scoreBox.id = "scoreDisplay";
        scoreBox.style.cssText =
            "text-align:center;color:#c9ad7d;font-size:18px;" +
            "font-weight:bold;margin:8px auto;";

        board.parentElement.insertBefore(scoreBox, board);
    }
}

// --------------------------------------------------
// ANZEIGEN
// --------------------------------------------------

// So viele Level gibt es (jede Form in board-layout.js ist eins).
function hoechstesLevel() {
    return Math.max(...Object.keys(BOARD_LAYOUTS).map(Number));
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

function updateStats() {
    // Nach dem letzten Level gibt es kein "Level 16" in der Anzeige.
    const angezeigtesLevel = Math.min(currentLevel, hoechstesLevel());

    setText(
        ["levelNumber", "level", "currentLevel", "levelDisplay"],
        String(angezeigtesLevel).padStart(2, "0")
    );

    setText(
        ["remainingTiles"],
        144 - foundPairs * 2
    );

    setText(
        ["remainingPairs", "pairs", "pairsDisplay"],
        72 - foundPairs
    );

    setText(
        ["gameTime", "timer", "gameTimer", "timeDisplay"],
        formatTime(gameSeconds)
    );

    setText(
        ["scoreDisplay"],
        `Punkte: ${score}`
    );
}

function setMessage(message) {
    const element = getElement(
        "gameStatus",
        "message",
        "gameMessage",
        "statusMessage"
    );

    if (element) {
        element.textContent = message;
    }
}

// --------------------------------------------------
// TIMER
// --------------------------------------------------

function startGameTimer() {
    stopGameTimer();

    gameTimer = setInterval(() => {
        if (!gameRunning) return;

        gameSeconds++;
        updateStats();
    }, 1000);
}

function stopGameTimer() {
    if (gameTimer !== null) {
        clearInterval(gameTimer);
        gameTimer = null;
    }
}

// --------------------------------------------------
// SPIEL STARTEN
// --------------------------------------------------

function startGame() {
    stopGameTimer();

    // Ein noch laufender Wechsel ins nächste Level ist damit erledigt.
    if (levelWechselTimer !== null) {
        clearTimeout(levelWechselTimer);
        levelWechselTimer = null;
    }

    gameRunning = false;
    selectedTile = null;
    hinweisTiles = [];
    foundPairs = 0;
    mistakes = 0;
    gameSeconds = 0;
    score = 0;

    if (gameData) {
        currentLevel = gameData.currentLevel || 1;
    }

    // Nach dem letzten Level nicht in eine Form ohne Vorlage laufen.
    if (!Number.isFinite(currentLevel) || currentLevel < 1) {
        currentLevel = 1;
    }
    if (currentLevel > hoechstesLevel()) {
        currentLevel = hoechstesLevel();
    }

    updateStats();
    setMessage("Spielbrett wird vorbereitet ...");
    leereGewonnenSpruch();
    leereGewonnenSymbol();
    zeigeWeiterKnopf("");

    try {
        currentBoard = generateSolvableBoard(MAHJONG_SYMBOLS, currentLevel);

        if (!Array.isArray(currentBoard) || currentBoard.length !== 144) {
            throw new Error("Das Spielbrett enthält nicht genau 144 Steine.");
        }

        gameRunning = true;

        renderBoard(currentBoard, null, handleTileClick);

        updateStats();
        setMessage("Viel Spaß beim Spielen!");

        startGameTimer();

    } catch (error) {
        gameRunning = false;

        console.error("Mahjong-Startfehler:", error);

        setMessage("Das Spielbrett konnte nicht erstellt werden.");
    }
}

// --------------------------------------------------
// STEINE AUSWÄHLEN
// --------------------------------------------------

function handleTileClick(tileId) {
    if (!gameRunning) return;

    const tile = currentBoard.find(
        item => item.id === tileId && !item.removed
    );

    if (!tile || !isTileFree(tile, currentBoard)) return;

    if (selectedTile === null) {
        selectedTile = tileId;
        hinweisTiles = [];

        aktualisiereBrett();

        return;
    }

    if (selectedTile === tileId) {
        selectedTile = null;
        hinweisTiles = [];

        aktualisiereBrett();

        return;
    }

    const firstTile = currentBoard.find(
        item => item.id === selectedTile && !item.removed
    );

    if (!firstTile) {
        selectedTile = null;
        hinweisTiles = [];

        aktualisiereBrett();

        return;
    }

    if (firstTile.symbolId === tile.symbolId) {
        removeTilePair(currentBoard, firstTile.id, tile.id);

        foundPairs++;
        score += 100;
        selectedTile = null;
        hinweisTiles = [];

        aktualisiereBrett();

        updateStats();

        if (foundPairs === 72) {
            finishGame();
        } else {
            setMessage(lobSpruch() + " +100 Punkte");
        }

        return;
    }

    mistakes++;
    selectedTile = null;
    hinweisTiles = [];

    aktualisiereBrett();

    updateStats();

    blinkeFalscheSteine([firstTile.id, tile.id]);

    setMessage(pechSpruch());
}

// --------------------------------------------------
// HINWEIS
// --------------------------------------------------

function showHint() {
    if (!gameRunning) return;

    const freeTiles = getFreeTiles(currentBoard);

    for (let i = 0; i < freeTiles.length; i++) {
        for (let j = i + 1; j < freeTiles.length; j++) {
            if (
                freeTiles[i].symbolId ===
                freeTiles[j].symbolId
            ) {
                // Beide Steine des Paares markieren - nicht nur einen.
                selectedTile = null;
                hinweisTiles = [freeTiles[i].id, freeTiles[j].id];

                aktualisiereBrett();

                setMessage("Diese beiden Steine passen zusammen.");

                return;
            }
        }
    }

    setMessage("Momentan ist kein passendes Paar frei.");
}

// --------------------------------------------------
// SPIEL GEWONNEN
// --------------------------------------------------

function finishGame() {
    gameRunning = false;

    stopGameTimer();

    const timeBonus = Math.max(0, 1000 - gameSeconds * 2);
    const mistakePenalty = mistakes * 25;

    score = Math.max(0, score + timeBonus - mistakePenalty);

    updateStats();

    setMessage(
        `GEWONNEN! Glückwunsch! ${score} Punkte – Zeit: ${formatTime(gameSeconds)}`
    );

    zeigeGewonnenSpruch();
    zeigeGewonnenSymbol();

    // Spielstand und Bestzeit speichern
    let gespeichert = false;

    try {
        if (gameData) {
            saveLevelResult(gameData, currentLevel, gameSeconds);

            gameData.totalScore =
                (gameData.totalScore || 0) + score;

            gameData.highScore = Math.max(
                gameData.highScore || 0,
                score
            );

            saveStorage(gameData);

            currentLevel = gameData.currentLevel;

            updateStats();

            gespeichert = true;
        }
    } catch (error) {
        console.error("Spielstand konnte nicht gespeichert werden:", error);
    }

    if (!gespeichert) {
        setMessage("Gewonnen! Der Spielstand konnte aber nicht gespeichert werden.");
        return;
    }

    // Nach dem Sieg wartet das Spiel auf "Weiter", damit die Meldung und der
    // Spruch in Ruhe gelesen werden koennen.
    const hoechstes = hoechstesLevel();

    if (currentLevel > hoechstes) {
        setMessage(
            `GEWONNEN! ${score} Punkte – Zeit: ${formatTime(gameSeconds)}. ` +
            `Alle ${hoechstes} Level geschafft!`
        );

        zeigeWeiterKnopf("");
        return;
    }

    setMessage(
        `GEWONNEN! ${score} Punkte – Zeit: ${formatTime(gameSeconds)}`
    );

    zeigeWeiterKnopf(`WEITER · LEVEL ${currentLevel}`);
}

// --------------------------------------------------
// WEITER-KNOPF NACH DEM SIEG
// --------------------------------------------------

// Ohne Beschriftung wird der Knopf versteckt.
function zeigeWeiterKnopf(beschriftung) {
    const knopf = getElement("continueButton");

    if (!knopf) return;

    if (!beschriftung) {
        knopf.hidden = true;
        return;
    }

    knopf.textContent = beschriftung;
    knopf.hidden = false;
    knopf.focus();
}

// --------------------------------------------------
// LEVEL-AUSWAHL
// --------------------------------------------------

function baueLevelAuswahl() {
    const raster = getElement("levelGrid");

    if (!raster) return;

    raster.innerHTML = "";

    for (let level = 1; level <= hoechstesLevel(); level++) {
        const feld = document.createElement("button");

        feld.type = "button";
        feld.className = "level-feld";
        feld.dataset.level = String(level);
        feld.textContent = String(level).padStart(2, "0");

        feld.addEventListener("click", () => waehleLevel(level));

        raster.appendChild(feld);
    }
}

function markiereAktuellesLevel() {
    const angezeigt = Math.min(currentLevel, hoechstesLevel());

    document.querySelectorAll(".level-feld").forEach(feld => {
        feld.classList.toggle(
            "aktuell",
            Number(feld.dataset.level) === angezeigt
        );
    });
}

function oeffneLevelAuswahl() {
    const auswahl = getElement("levelPicker");

    if (!auswahl) return;

    baueLevelAuswahl();
    markiereAktuellesLevel();

    auswahl.hidden = false;

    const aktuell =
        auswahl.querySelector(".level-feld.aktuell") ||
        auswahl.querySelector(".level-feld");

    if (aktuell) {
        aktuell.focus();
        aktuell.scrollIntoView({ block: "center" });
    }
}

function schliesseLevelAuswahl() {
    const auswahl = getElement("levelPicker");

    if (!auswahl) return;

    auswahl.hidden = true;
}

function waehleLevel(level) {
    const ziel = Math.min(Math.max(1, level), hoechstesLevel());

    if (gameData) {
        gameData.currentLevel = ziel;

        try {
            saveStorage(gameData);
        } catch (fehler) {}
    }

    currentLevel = ziel;
    gameData.currentLevel = ziel;

    schliesseLevelAuswahl();
    startGame();

    setMessage(`Level ${ziel} gestartet.`);
}

// --------------------------------------------------
// INITIALISIERUNG
// --------------------------------------------------

function initializeGame() {
    createGameInformation();

    try {
        gameData = loadStorage();
        currentLevel = gameData.currentLevel || 1;
    } catch (error) {
        console.error("Speicherfehler:", error);

        gameData = createDefaultStorage();
    }

    const newGameButton = getElement("newGameButton");

    if (newGameButton) {
        newGameButton.addEventListener("click", startGame);
    }

    const continueButton = getElement("continueButton");

    if (continueButton) {
        continueButton.addEventListener("click", () => {
            continueButton.hidden = true;
            startGame();
        });
    }

    const hintButton = getElement("hintButton");

    if (hintButton) {
        hintButton.addEventListener("click", showHint);
    }

    // Zurück zur Spiele-Übersicht (Connector).
    const backButton = getElement("backButton");

    if (backButton) {
        backButton.addEventListener("click", () => {
            window.location.href = "../spiele.html";
        });
    }

    // Vollbild an und aus – Optik wie die übrigen Knöpfe.
    const fullscreenButton = getElement("fullscreenButton");

    if (fullscreenButton) {
        const imVollbild = () =>
            !!(document.fullscreenElement || document.webkitFullscreenElement);

        const zeigeVollbild = () => {
            fullscreenButton.textContent =
                imVollbild() ? "⛶ VOLLBILD AUS" : "⛶ VOLLBILD";
        };

        fullscreenButton.addEventListener("click", () => {
            try {
                if (imVollbild()) {
                    const beenden = document.exitFullscreen
                        ? document.exitFullscreen()
                        : (document.webkitExitFullscreen
                            ? document.webkitExitFullscreen()
                            : null);

                    if (beenden && beenden.catch) beenden.catch(() => {});
                } else {
                    const ziel = document.documentElement;
                    const start = ziel.requestFullscreen
                        ? ziel.requestFullscreen()
                        : (ziel.webkitRequestFullscreen
                            ? ziel.webkitRequestFullscreen()
                            : null);

                    if (!start) {
                        setMessage("Dieser Browser kann kein Vollbild.");
                        return;
                    }

                    if (start.catch) {
                        start.catch(() =>
                            setMessage("Vollbild wurde vom Browser blockiert.")
                        );
                    }
                }
            } catch (error) {
                setMessage("Vollbild ging nicht: " + error.message);
            }

            setTimeout(zeigeVollbild, 300);
        });

        document.addEventListener("fullscreenchange", zeigeVollbild);
        document.addEventListener("webkitfullscreenchange", zeigeVollbild);
    }

    // Level-Auswahl öffnen und schließen.
    const levelButton = getElement("levelButton");
    const levelPicker = getElement("levelPicker");
    const levelPickerClose = getElement("levelPickerClose");

    if (levelButton) {
        levelButton.addEventListener("click", oeffneLevelAuswahl);
    }

    if (levelPickerClose) {
        levelPickerClose.addEventListener("click", schliesseLevelAuswahl);
    }

    if (levelPicker) {
        levelPicker.addEventListener("click", event => {
            if (event.target === levelPicker) {
                schliesseLevelAuswahl();
            }
        });
    }

    document.addEventListener("keydown", event => {
        if (event.key === "Escape") {
            schliesseLevelAuswahl();
        }
    });

    // Zurück auf Level 1 – der übrige Spielstand bleibt erhalten.
    const resetLevelButton = getElement("resetLevelButton");

    if (resetLevelButton) {
        resetLevelButton.addEventListener("click", () => {
            if (gameData) {
                gameData.currentLevel = 1;
                saveStorage(gameData);
            }

            currentLevel = 1;
            startGame();
            setMessage("Zurück auf Level 1 – der übrige Spielstand bleibt erhalten.");
        });
    }

    updateStats();

    startGame();
}

// --------------------------------------------------
// START
// --------------------------------------------------

// Auf dem Handy laeuft die Zeit sonst weiter, waehrend die App im
// Hintergrund ist.
document.addEventListener("visibilitychange", () => {
    if (document.hidden) {
        stopGameTimer();
        return;
    }

    if (gameRunning) {
        startGameTimer();
    }
});

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initializeGame);
} else {
    initializeGame();
}

