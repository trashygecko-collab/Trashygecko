
/*
 * TRASHYGECKO MAHJONG
 * Hauptsteuerung mit Zeit, Punkten und Spielstand
 */

const MAHJONG_SYMBOLS = Array.from(
    { length: 57 },
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
            "text-align:center;color:#d5b36a;font-size:18px;" +
            "font-weight:bold;margin:8px auto;";

        board.parentElement.insertBefore(scoreBox, board);
    }
}

// --------------------------------------------------
// ANZEIGEN
// --------------------------------------------------

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
    setText(
        ["levelNumber", "level", "currentLevel", "levelDisplay"],
        String(currentLevel).padStart(2, "0")
    );

    setText(
        ["remainingTiles"],
        114 - foundPairs * 2
    );

    setText(
        ["remainingPairs", "pairs", "pairsDisplay"],
        57 - foundPairs
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

    gameRunning = false;
    selectedTile = null;
    foundPairs = 0;
    mistakes = 0;
    gameSeconds = 0;
    score = 0;

    if (gameData) {
        currentLevel = gameData.currentLevel || 1;
    }

    updateStats();
    setMessage("Spielbrett wird vorbereitet ...");

    try {
        currentBoard = generateSolvableBoard(MAHJONG_SYMBOLS, currentLevel);

        if (!Array.isArray(currentBoard) || currentBoard.length !== 114) {
            throw new Error("Das Spielbrett enthält nicht genau 114 Steine.");
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

        renderBoard(currentBoard, selectedTile, handleTileClick);

        return;
    }

    if (selectedTile === tileId) {
        selectedTile = null;

        renderBoard(currentBoard, null, handleTileClick);

        return;
    }

    const firstTile = currentBoard.find(
        item => item.id === selectedTile && !item.removed
    );

    if (!firstTile) {
        selectedTile = null;
        renderBoard(currentBoard, null, handleTileClick);
        return;
    }

    if (firstTile.symbolId === tile.symbolId) {
        removeTilePair(currentBoard, firstTile.id, tile.id);

        foundPairs++;
        score += 100;
        selectedTile = null;

        renderBoard(currentBoard, null, handleTileClick);

        updateStats();

        if (foundPairs === 57) {
            finishGame();
        } else {
            setMessage("Paar gefunden! +100 Punkte");
        }

        return;
    }

    mistakes++;
    selectedTile = null;

    renderBoard(currentBoard, null, handleTileClick);

    updateStats();

    setMessage("Das sind keine gleichen Steine.");
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
                selectedTile = freeTiles[i].id;

                renderBoard(
                    currentBoard,
                    selectedTile,
                    handleTileClick
                );

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

    // Spielstand und Bestzeit speichern
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
        }
    } catch (error) {
        console.error("Spielstand konnte nicht gespeichert werden:", error);
        setMessage("Gewonnen! Der Spielstand konnte aber nicht gespeichert werden.");
    }
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

    const hintButton = getElement("hintButton");

    if (hintButton) {
        hintButton.addEventListener("click", showHint);
    }

    const shuffleButton = getElement("shuffleButton");

    if (shuffleButton) {
        shuffleButton.addEventListener("click", () => {
            setMessage("Mischen ist noch nicht verfügbar.");
        });
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

if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initializeGame);
} else {
    initializeGame();
}

