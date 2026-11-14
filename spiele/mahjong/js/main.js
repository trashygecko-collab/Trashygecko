
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

// Einmal pro Level darf kostenlos gemischt werden, wenn kein Zug mehr geht.
let mischenMoeglich = true;

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

// So viele Formen gibt es (jede Form in board-layout.js ist eine).
function hoechstesLevel() {
    return Math.max(...Object.keys(BOARD_LAYOUTS).map(Number));
}

// Nach der letzten Form beginnt die naechste Runde mit Form 1 - dann in einer
// anderen Steinfarbe. So laesst sich endlos weiterspielen.
// Von hell nach dunkel - die letzte Runde ist Schwarz.
const RUNDEN_FARBEN = [
    "Elfenbein",
    "Bernstein",
    "Rosé",
    "Jade",
    "Azur",
    "Amethyst",
    "Schiefer",
    "Weinrot",
    "Tiefschwarz",
    // Versteckte Zugabe: eine Runde mit allen Farben gleichzeitig.
    "Regenbogen"
];

// Nur diese Farben stehen in der Farbwahl - die Regenbogen-Runde bleibt
// eine Ueberraschung und taucht erst nach Level 504 auf.
const RUNDEN_FARBEN_SICHTBAR = RUNDEN_FARBEN.slice(0, RUNDEN_FARBEN.length - 1);

// Laeuft gerade die Regenbogen-Runde (Level 505-560, 1065-1120, ...)?
function istRegenbogenRunde(runde) {
    const nummer = ((runde - 1) % RUNDEN_FARBEN.length) + 1;

    return nummer === RUNDEN_FARBEN.length;
}

// Welche Form gehoert zu diesem Level?
function formVon(level) {
    return ((level - 1) % hoechstesLevel()) + 1;
}

// In welcher Runde liegt dieses Level?
function rundeVon(level) {
    return Math.floor((level - 1) / hoechstesLevel()) + 1;
}

// Name der Steinfarbe dieser Runde.
function farbenName(runde) {
    return RUNDEN_FARBEN[(runde - 1) % RUNDEN_FARBEN.length];
}

// Letztes Level einer vollen Farbrunde? (Level 504, 1008, ...)
// Kurzer Hinweis, wenn ein Level den Anfang einer neuen Farbrunde bildet.
function rundeStartHinweis(level) {
    const runde = rundeVon(level);

    if (formVon(level) !== 1 || runde < 2) return "";

    if (istRegenbogenRunde(runde)) {
        return 'Runde ' + runde + ' – alle Farben. Jedes Paar hat seine eigene Farbe.';
    }

    return 'Neue Runde ' + runde + ' – Steinfarbe ' + farbenName(runde) + '.';
}

function istRundenFinale(level) {
    if (!Number.isFinite(level)) return false;

    return (
        formVon(level) === hoechstesLevel() &&
        rundeVon(level) % RUNDEN_FARBEN.length === 0
    );
}

// Setzt die Steinfarbe fuer eine Runde (CSS: body[data-runde]).
function setzeRundenFarbe(runde) {
    const nummer = ((runde - 1) % RUNDEN_FARBEN.length) + 1;

    document.body.dataset.runde = String(nummer);

    // In der Regenbogen-Runde bekommt jeder Stein seine eigene Farbe.
    document.body.dataset.regenbogen = istRegenbogenRunde(runde) ? "ja" : "nein";
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
    // Level laufen endlos weiter - die Anzeige zeigt die volle Nummer.
    setText(
        ["levelNumber", "level", "currentLevel", "levelDisplay"],
        String(currentLevel).padStart(2, "0")
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

// Zufaellige Spruchfarbe, aber nie zweimal dieselbe hintereinander.
function spruchFarbe() {
    const auswahl = SPRUCH_FARBEN.filter(farbe => farbe !== letzteSpruchFarbe);
    const farbe = auswahl[Math.floor(Math.random() * auswahl.length)];

    letzteSpruchFarbe = farbe;

    return farbe;
}

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

function siegSpruchFarbe() {
    const auswahl = SIEG_SPRUCH_FARBEN.filter(
        farbe => farbe !== letzteSiegSpruchFarbe
    );
    const farbe = auswahl[Math.floor(Math.random() * auswahl.length)];

    letzteSiegSpruchFarbe = farbe;

    return farbe;
}

// rot = true faerbt die Zeile rot (Fehlgriff), sonst normale Farbe.
// punkte: Zusatz wie "+ 100 Punkte" - bleibt grau (bzw. rot beim Fehlgriff),
// damit der Spruch davor allein die Farbe traegt.
function setMessage(message, rot = false, punkte = "") {
    const element = getElement(
        "gameStatus",
        "message",
        "gameMessage",
        "statusMessage"
    );

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

    if (!Number.isFinite(currentLevel) || currentLevel < 1) {
        currentLevel = 1;
    }

    // Jede Runde spielt dieselben Formen in einer anderen Steinfarbe.
    setzeRundenFarbe(rundeVon(currentLevel));

    updateStats();

    // Neues Level: die Anzeige startet wieder in Gold.
    stoppePunkteBlitz();

    setMessage("Spielbrett wird vorbereitet ...");
    leereGewonnenSpruch();
    leereGewonnenSymbol();
    zeigeWeiterKnopf("");
    zeigeSiegel(false);
    versteckeSiegEhrung();
    versteckeMischenKnopf();
    mischenMoeglich = true;

    try {
        currentBoard = generateSolvableBoard(
            MAHJONG_SYMBOLS,
            formVon(currentLevel)
        );

        if (!Array.isArray(currentBoard) || currentBoard.length !== 144) {
            throw new Error("Das Spielbrett enthält nicht genau 144 Steine.");
        }

        gameRunning = true;

        renderBoard(currentBoard, null, handleTileClick);

        updateStats();
        setMessage(rundeStartHinweis(currentLevel) || "Viel Spaß beim Spielen!");

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
            setMessage(lobSpruch(), false, "+100 Punkte");
            pruefeFestgefahren();
        }

        return;
    }

    // Nur solange Punkte da sind, wird gespottet und abgezogen.
    const abzugMoeglich = score > 0;

    mistakes++;
    score = Math.max(0, score - 100);
    selectedTile = null;
    hinweisTiles = [];

    aktualisiereBrett();

    updateStats();

    blinkeFalscheSteine([firstTile.id, tile.id]);

    if (abzugMoeglich) {
        blinkePunkte();
        setMessage(pechSpruch(), true, "- 100 Punkte");
    } else {
        // Bei 0 Punkten gibt es nichts mehr zu holen - dann sachlich bleiben.
        setMessage("Die 2 Steine passen nicht zusammen.");
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

// --------------------------------------------------
// FESTGEFAHRENES BRETT
// --------------------------------------------------

// Gibt es überhaupt noch zwei freie Steine mit gleichem Motiv?
function gibtEsNochEinPaar() {
    const frei = getFreeTiles(currentBoard);
    const gesehen = new Set();

    for (const stein of frei) {
        if (gesehen.has(stein.symbolId)) return true;

        gesehen.add(stein.symbolId);
    }

    return false;
}

function zeigeMischenKnopf() {
    const knopf = getElement("mischenButton");

    if (knopf) knopf.hidden = false;
}

function versteckeMischenKnopf() {
    const knopf = getElement("mischenButton");

    if (knopf) knopf.hidden = true;
}

// Nach jedem Treffer prüfen, ob das Brett noch spielbar ist.
function pruefeFestgefahren() {
    if (!gameRunning) return;
    if (gibtEsNochEinPaar()) return;

    if (mischenMoeglich) {
        zeigeMischenKnopf();
        setMessage(
            "Kein passendes Paar mehr frei – du darfst einmal kostenlos mischen."
        );
    } else {
        versteckeMischenKnopf();
        setMessage(
            "Kein passendes Paar mehr frei. Mit NEU startest du das Level neu."
        );
    }
}

// Verteilt die Motive der restlichen Steine neu - nach einer Reihenfolge, die
// garantiert wieder aufgeht. Die Form des Bretts bleibt unverändert.
function mischeBrett() {
    let order;

    try {
        order = findRemovalOrder(currentBoard);
    } catch (error) {
        console.error("Mischen fehlgeschlagen:", error);

        versteckeMischenKnopf();
        setMessage("Mischen ging nicht – starte das Level mit NEU neu.");

        return;
    }

    const motive = [
        ...new Set(
            currentBoard
                .filter(tile => !tile.removed)
                .map(tile => tile.symbolId)
        )
    ];

    if (order.length !== motive.length) {
        versteckeMischenKnopf();
        setMessage("Mischen ging nicht – starte das Level mit NEU neu.");

        return;
    }

    const nachId = new Map(currentBoard.map(tile => [tile.id, tile]));

    order.forEach((paar, index) => {
        const symbolId = motive[index];

        nachId.get(paar[0]).symbolId = symbolId;
        nachId.get(paar[1]).symbolId = symbolId;
    });

    mischenMoeglich = false;
    versteckeMischenKnopf();
    selectedTile = null;
    hinweisTiles = [];

    aktualisiereBrett();
    setMessage("Neu gemischt – das war das eine Gratis-Mischen in diesem Level.");
}

// --------------------------------------------------
// HINWEIS
// --------------------------------------------------

function showHint() {
    if (!gameRunning) return;

    // Ein Tipp kostet 200 Punkte - darunter gibt es keinen mehr.
    if (score < TIPP_KOSTEN) {
        setMessage(`Für einen Tipp brauchst du mindestens ${TIPP_KOSTEN} Punkte.`);

        return;
    }

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

                // Ein Tipp kostet 200 Punkte - der Spruch kommt wie beim
                // Fehlgriff in Rot, nur milder.
                score = Math.max(0, score - TIPP_KOSTEN);

                updateStats();
                blinkePunkte();
                setMessage(pechSpruch(), true, `- ${TIPP_KOSTEN} Punkte`);

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

    // Level, das gerade geschafft wurde (currentLevel zaehlt gleich hoch).
    const geschafftesLevel = currentLevel;

    const timeBonus = Math.max(0, 1000 - gameSeconds * 2);
    const mistakePenalty = mistakes * 25;

    score = Math.max(0, score + timeBonus - mistakePenalty);

    updateStats();

    setMessage(
        `GEWONNEN! Glückwunsch! ${score} Punkte – Zeit: ${formatTime(gameSeconds)}`
    );

    zeigeGewonnenSpruch();
    zeigeGewonnenSymbol();
    zeigeSiegel(true);
    zeigeSiegWerte();
    zeigeSiegEhrung(geschafftesLevel, score);

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

    // Nach dem Sieg wartet das Spiel auf "Weiter", damit Meldung und Spruch in
    // Ruhe gelesen werden. Danach geht es endlos weiter: nach der letzten Form
    // beginnt die naechste Runde in neuer Steinfarbe.
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
// SIEGEREHRUNG
// --------------------------------------------------

// Zeigt die goldene Tafel mit Level, Punkten und Spruch.
function zeigeSiegEhrung(level, punkte) {
    const tafel = getElement("siegEhrung");

    if (!tafel) return;

    const levelFeld = getElement("siegLevel");
    const punkteFeld = getElement("siegPunkte");

    if (levelFeld) {
        levelFeld.textContent = Number.isFinite(level)
            ? `Level ${level} · Runde ${rundeVon(level)} · ${farbenName(rundeVon(level))}`
            : "";
    }

    if (punkteFeld) {
        punkteFeld.textContent = Number(punkte).toLocaleString("de-DE");
    }

    // Nach der letzten Farbe einer vollen Runde kommt die eigene Tafel dazu.
    const rundenFeld = getElement("siegRunde");

    if (rundenFeld) {
        rundenFeld.hidden = !istRundenFinale(level);
    }

    tafel.hidden = false;
}

function versteckeSiegEhrung() {
    const tafel = getElement("siegEhrung");

    if (tafel) {
        tafel.hidden = true;
    }

    const rundenFeld = getElement("siegRunde");

    if (rundenFeld) {
        rundenFeld.hidden = true;
    }
}

// --------------------------------------------------
// SIGNATUR IM SIEG-ZUSTAND
// --------------------------------------------------

// Goldene Zeile unter dem Weiter-Knopf - nur nach einem gewonnenen Level.
function zeigeSiegel(anzeigen) {
    const siegel = getElement("siegel");

    if (siegel) {
        siegel.hidden = !anzeigen;
    }
}

// Zeit und Highscore in der Signatur-Zeile unter dem Weiter-Knopf.
function zeigeSiegWerte() {
    const werte = getElement("siegelWerte");

    if (!werte) return;

    const highscore =
        gameData && Number.isFinite(gameData.highScore)
            ? gameData.highScore
            : score;

    werte.textContent =
        `ZEIT ${formatTime(gameSeconds)} · ` +
        `HIGHSCORE ${highscore.toLocaleString("de-DE")}`;
}

// --------------------------------------------------
// LEVEL-AUSWAHL
// --------------------------------------------------

// Runde, die in der Level-Auswahl gerade angezeigt wird.
let auswahlRunde = 1;

// Bestzeit eines Levels in Sekunden - null, wenn es noch nicht geschafft ist.
function bestzeitVon(level) {
    const zeiten = gameData ? gameData.bestTimes : null;
    const wert = zeiten ? zeiten[level] : undefined;

    return Number.isFinite(wert) ? wert : null;
}

// Wie viele Level dieser Runde sind schon geschafft?
function geschafftInRunde(runde) {
    const formen = hoechstesLevel();
    const start = (runde - 1) * formen;
    let anzahl = 0;

    for (let form = 1; form <= formen; form++) {
        if (bestzeitVon(start + form) !== null) anzahl++;
    }

    return anzahl;
}

// --------------------------------------------------
// STEIN-FORM UND ALTERUNG
// --------------------------------------------------

const STIL_FORMEN = ["rund", "eckig"];
const STIL_ZUSTAENDE = ["neu", "alt"];

// Setzt Form und Zustand der Steine (CSS: body[data-form], body[data-zustand]).
function setzeStil(form, zustand, speichern = true) {
    const formWert = STIL_FORMEN.includes(form) ? form : "rund";
    const zustandWert = STIL_ZUSTAENDE.includes(zustand) ? zustand : "neu";

    document.body.dataset.form = formWert;
    document.body.dataset.zustand = zustandWert;

    document.querySelectorAll(".stil-gruppe").forEach(gruppe => {
        const aktiv =
            gruppe.dataset.gruppe === "form" ? formWert : zustandWert;

        gruppe.querySelectorAll("button").forEach(knopf => {
            const an = knopf.dataset.wert === aktiv;

            knopf.classList.toggle("aktiv", an);
            knopf.setAttribute("aria-pressed", an ? "true" : "false");
        });
    });

    if (speichern && gameData) {
        gameData.settings.steinForm = formWert;
        gameData.settings.steinZustand = zustandWert;

        saveStorage(gameData);
    }
}

function aktuellerStil() {
    const settings = (gameData && gameData.settings) || {};

    return {
        form: STIL_FORMEN.includes(settings.steinForm) ? settings.steinForm : "rund",
        zustand: STIL_ZUSTAENDE.includes(settings.steinZustand)
            ? settings.steinZustand
            : "neu"
    };
}

function baueStilwahl() {
    const leiste = getElement("stilLeiste");

    if (!leiste) return;

    leiste.querySelectorAll(".stil-gruppe").forEach(gruppe => {
        const feld = gruppe.dataset.gruppe;

        gruppe.querySelectorAll("button").forEach(knopf => {
            knopf.addEventListener("click", () => {
                const jetzt = aktuellerStil();

                if (feld === "form") {
                    setzeStil(knopf.dataset.wert, jetzt.zustand);
                } else {
                    setzeStil(jetzt.form, knopf.dataset.wert);
                }
            });
        });
    });
}

function baueLevelAuswahl(runde = auswahlRunde) {
    const raster = getElement("levelGrid");

    if (!raster) return;

    auswahlRunde = Math.max(1, runde);

    raster.innerHTML = "";

    const formen = hoechstesLevel();
    const start = (auswahlRunde - 1) * formen;

    for (let form = 1; form <= formen; form++) {
        const level = start + form;
        const feld = document.createElement("button");
        const bestzeit = bestzeitVon(level);

        feld.type = "button";
        feld.className = "level-feld";
        feld.dataset.level = String(level);

        const nummer = document.createElement("span");

        nummer.className = "level-feld-nummer";
        nummer.textContent = String(level).padStart(2, "0");
        feld.appendChild(nummer);

        if (bestzeit !== null) {
            const zeit = document.createElement("span");

            zeit.className = "level-feld-zeit";
            zeit.textContent = formatTime(bestzeit);
            feld.appendChild(zeit);

            feld.classList.add("geschafft");
        }

        feld.title =
            bestzeit !== null
                ? `Form ${form} · Bestzeit ${formatTime(bestzeit)}`
                : `Form ${form}`;

        feld.addEventListener("click", () => waehleLevel(level));

        raster.appendChild(feld);
    }

    const anzeige = getElement("rundeAnzeige");

    if (anzeige) {
        anzeige.textContent =
            `Runde ${auswahlRunde} · ${farbenName(auswahlRunde)} · ` +
            `${geschafftInRunde(auswahlRunde)} von ${formen}`;
    }

    baueFarbwahl();
    markiereAktuellesLevel();
}

// Farbwahl am Ende der Level-Auswahl: jede Farbe ist eine Runde.
function baueFarbwahl() {
    const leiste = getElement("farbeLeiste");

    if (!leiste) return;

    leiste.innerHTML = "";

    RUNDEN_FARBEN_SICHTBAR.forEach((name, index) => {
        const runde = index + 1;
        const abLevel = (runde - 1) * hoechstesLevel() + 1;

        const feld = document.createElement("button");

        feld.type = "button";
        feld.className = "farbe-feld";
        feld.dataset.farbe = String(runde);
        const geschafft = geschafftInRunde(runde);

        feld.title =
            `Runde ${runde} · ${name} · ab Level ${abLevel} · ` +
            `${geschafft} von ${hoechstesLevel()} geschafft`;
        feld.setAttribute(
            "aria-label",
            `Runde ${runde}, ${name}, ab Level ${abLevel}`
        );

        feld.classList.toggle("aktiv", runde === auswahlRunde);

        feld.addEventListener("click", () => baueLevelAuswahl(runde));

        leiste.appendChild(feld);
    });
}

function markiereAktuellesLevel() {
    document.querySelectorAll(".level-feld").forEach(feld => {
        feld.classList.toggle(
            "aktuell",
            Number(feld.dataset.level) === currentLevel
        );
    });
}

function oeffneLevelAuswahl() {
    const auswahl = getElement("levelPicker");

    if (!auswahl) return;

    baueLevelAuswahl(rundeVon(currentLevel));

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
    // Level laufen endlos: nach Form 56 geht es in der naechsten Runde weiter.
    const ziel = Math.max(1, Math.floor(level));

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

    const mischenButton = getElement("mischenButton");

    if (mischenButton) {
        mischenButton.addEventListener("click", mischeBrett);
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

    // Steinform und Alterung aus dem Spielstand uebernehmen.
    baueStilwahl();
    setzeStil(aktuellerStil().form, aktuellerStil().zustand, false);

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

