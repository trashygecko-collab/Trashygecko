
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

// Formensatz des laufenden Spiels (1 = Normal, 2 = Handy). Jedes Spiel beginnt
// mit den Handy-Leveln - auch wenn zuletzt ein Normal-Level gespielt wurde.
let FORMEN_SATZ = 2;

// Schwierigkeitsgrad: "leicht" ist das bisherige Spiel (72 Motive zu je zwei
// Steinen), "schwer" teilt 36 Motive zu je vier Steinen aus.  Der Grad gilt fuer
// die laufende Partie und wird fuer die naechste aus den Einstellungen geladen.
let spielModus = "leicht";

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

function getElement(id) {
    return document.getElementById(id);
}

function setText(id, value) {
    const element = getElement(id);

    if (element) {
        element.textContent = value;
    }
}

// --------------------------------------------------
// ANZEIGEN
// --------------------------------------------------

// So viele Formen gibt es (jede Form in board-layout.js ist eine).
function hoechstesLevel(satz = FORMEN_SATZ) {
    const formen = satz === 2 ? BOARD_LAYOUTS_2 : BOARD_LAYOUTS;

    return Math.max(...Object.keys(formen).map(Number));
}

// Der Spielstand liegt getrennt vor: Formen 1 wie bisher ganz oben,
// Formen 2 in einem eigenen Abschnitt mit eigenen Haken und Bestzeiten.
function satzDaten(satz = FORMEN_SATZ) {
    if (!gameData) gameData = createDefaultStorage();

    if (satz !== 2) return gameData;

    if (!gameData.satz2) {
        gameData.satz2 = neueSatzDaten();
    }

    return gameData.satz2;
}

// Name der Form – im zweiten Satz hat jede Form einen.
function formName(level, satz = FORMEN_SATZ) {
    return satz === 2 ? formNameSatz2(formVon(level, satz)) : "";
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
    // Zugabe: eine Runde mit allen Farben gleichzeitig - jedes Paar in eigener
    // Farbe.  Steht wie die anderen Farben in der Farbwahl (Runde 10, ab 505).
    "Regenbogen"
];

// Laeuft gerade die Regenbogen-Runde (Level 505-560, 1065-1120, ...)?
function istRegenbogenRunde(runde) {
    const nummer = ((runde - 1) % RUNDEN_FARBEN.length) + 1;

    return nummer === RUNDEN_FARBEN.length;
}

// Welche Form gehoert zu diesem Level?
function formVon(level, satz = FORMEN_SATZ) {
    return ((level - 1) % hoechstesLevel(satz)) + 1;
}

// In welcher Runde liegt dieses Level?
function rundeVon(level, satz = FORMEN_SATZ) {
    return Math.floor((level - 1) / hoechstesLevel(satz)) + 1;
}

// Name der Steinfarbe dieser Runde.
function farbenName(runde) {
    return RUNDEN_FARBEN[(runde - 1) % RUNDEN_FARBEN.length];
}

// So heissen die beiden Levelsysteme - genau wie die Knoepfe am Umschalter.
const SATZ_NAMEN = { 1: "Normal-Level", 2: "Handy-Level" };

function satzName(satz = FORMEN_SATZ) {
    return SATZ_NAMEN[satz === 2 ? 2 : 1];
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
        currentLevel = satzDaten().currentLevel || 1;
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
    zeigeHinweis("");
    leereGewonnenSpruch();
    leereGewonnenSymbol();
    zeigeWeiterKnopf("");
    zeigeSiegel(false);
    versteckeSiegEhrung();

    try {
        currentBoard = spielModus === "schwer"
            ? generateHardBoard(MAHJONG_SYMBOLS, formVon(currentLevel))
            : generateSolvableBoard(MAHJONG_SYMBOLS, formVon(currentLevel));

        if (!Array.isArray(currentBoard) || currentBoard.length !== 144) {
            throw new Error("Das Spielbrett enthält nicht genau 144 Steine.");
        }

        gameRunning = true;

        renderBoard(currentBoard, null, handleTileClick);

        zeigeMixKnopf();
        aktualisiereMixKnopf();

        updateStats();
        // Runden-Hinweis, wenn es einen gibt (Farbe/Regenbogen), sonst ein
        // Spruch aus der gemeinsamen Liste - wie nach einem Treffer.
        setMessage(rundeStartHinweis(currentLevel) || lobText(true));

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

    // Stein wieder abgewaehlt: ein Spruch aus pech.txt, aber kein Punktabzug -
    // das ist ja nur ein zweiter Tipp auf denselben Stein, kein Fehlgriff.
    if (selectedTile === tileId) {
        selectedTile = null;
        hinweisTiles = [];

        aktualisiereBrett();
        setMessage(pechText(), true);

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
            setMessage(lobText(true), false, "+100 Punkte");
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
        setMessage(pechText(), true, "- 100 Punkte");
    } else {
        // Bei 0 Punkten gibt es nichts mehr zu holen - dann sachlich bleiben.
        setMessage("Die 2 Steine passen nicht zusammen.");
    }
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

// Verteilt die Motive der restlichen Steine neu - auf denselben Positionen und
// nach einer Abbaureihenfolge, die mit den neuen Motiven garantiert wieder
// aufgeht.  Entfernte Steine bleiben entfernt.  Meldung: keine, das soll
// niemanden unterbrechen.
function verteileMotiveNeu() {
    const rest = currentBoard.filter(tile => !tile.removed);

    if (rest.length === 0) return false;

    // Wie oft kommt jedes Motiv noch vor?  In Leicht zweimal, in Schwer bis zu
    // viermal - ungerade Zahlen waeren ein Fehler.
    const anzahl = new Map();

    for (const tile of rest) {
        anzahl.set(tile.symbolId, (anzahl.get(tile.symbolId) || 0) + 1);
    }

    for (const zahl of anzahl.values()) {
        if (zahl % 2 !== 0) return false;
    }

    let order;

    try {
        order = findRemovalOrder(currentBoard);
    } catch (error) {
        console.error("Neuverteilen fehlgeschlagen:", error);

        return false;
    }

    if (!Array.isArray(order) || order.length * 2 !== rest.length) return false;

    // Ein Motiv mit vier Steinen belegt zwei Schritte, eines mit zwei Steinen
    // einen.  Die Motive werden gemischt, damit jede Neuverteilung anders ist.
    const vorrat = [];

    for (const [motiv, zahl] of anzahl) {
        for (let k = 0; k < zahl / 2; k++) vorrat.push(motiv);
    }

    if (vorrat.length !== order.length) return false;

    const verteilt = shuffle(vorrat);
    const alt = new Map(rest.map(tile => [tile.id, tile.symbolId]));
    const nachId = new Map(currentBoard.map(tile => [tile.id, tile]));

    order.forEach((paar, index) => {
        const motiv = verteilt[index];

        nachId.get(paar[0]).symbolId = motiv;
        nachId.get(paar[1]).symbolId = motiv;
    });

    // Die neue Verteilung muss die Reihenfolge wirklich aufgehen lassen.  Erst
    // dann gilt sie - sonst bleibt das Brett unangetastet.
    if (!pruefeAbbau(currentBoard, order)) {
        for (const [id, motiv] of alt) nachId.get(id).symbolId = motiv;

        console.error("Neuverteilung liess sich nicht abbauen - Brett bleibt.");

        return false;
    }

    currentBoard.loesung = order;

    selectedTile = null;
    hinweisTiles = [];

    aktualisiereBrett();

    return true;
}

// Nach jedem Treffer prüfen, ob das Brett noch spielbar ist.  In Leicht werden
// die restlichen Motive still neu verteilt, damit das Level nicht verloren ist.
// In Schwer bleibt die Sackgasse stehen: dort ist Mischen die Hilfe, und genau
// dann leuchtet der Knopf auf.
function pruefeFestgefahren() {
    if (!gameRunning) return;

    if (gibtEsNochEinPaar()) {
        aktualisiereMixKnopf();

        return;
    }

    if (spielModus !== "schwer" && verteileMotiveNeu()) return;

    setMessage(
        spielModus === "schwer"
            ? "Kein passendes Paar mehr frei – mit MISCHEN geht es weiter."
            : "Kein passendes Paar mehr frei. Über die Level-Taste neu starten."
    );

    aktualisiereMixKnopf();
}

// --------------------------------------------------
// MISCHEN (nur im schweren Modus)
// --------------------------------------------------

// Der Knopf gehoert zum schweren Modus - in Leicht bleibt er versteckt.
function zeigeMixKnopf() {
    const knopf = getElement("mixButton");

    if (!knopf) return;

    knopf.hidden = spielModus !== "schwer";
}

// Sobald kein passendes Paar mehr frei ist, faellt der Knopf auf: dann ist er
// die einzige Hilfe.  Sonst bleibt er dezent.
function aktualisiereMixKnopf() {
    const knopf = getElement("mixButton");

    if (!knopf || spielModus !== "schwer") return;

    const dringend = gameRunning && !gibtEsNochEinPaar();

    knopf.classList.toggle("dringend", dringend);
    knopf.disabled = !gameRunning;
    knopf.title = dringend
        ? "Kein passendes Paar mehr frei - Motive neu mischen"
        : "Motive der verbliebenen Steine neu mischen";
}

// Mischen: Motive der restlichen Steine neu verteilen.  Positionen, entfernte
// Steine, Punkte und Zeit bleiben, wie sie sind - nur die Motive wandern.
function mischeMotive() {
    if (!gameRunning || spielModus !== "schwer") return;

    if (!verteileMotiveNeu()) {
        setMessage("Die Motive liessen sich nicht neu mischen.");

        return;
    }

    updateStats();
    aktualisiereMixKnopf();
    setMessage("Motive neu gemischt.");
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
                setMessage(pechText(), true, `- ${TIPP_KOSTEN} Punkte`);

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

    // Spielstand und Bestzeit speichern. Scheitert das, sagt die Hinweiszeile
    // darunter Bescheid - die Siegmeldung bleibt stehen.
    try {
        if (gameData) {
            saveLevelResult(gameData, currentLevel, gameSeconds, satzDaten());

            gameData.totalScore =
                (gameData.totalScore || 0) + score;

            gameData.highScore = Math.max(
                gameData.highScore || 0,
                score
            );

            speichereSpielstand(
                "Der Spielstand lässt sich nicht speichern – der Sieg zählt nur " +
                "für diese Sitzung."
            );

            currentLevel = satzDaten().currentLevel;

            updateStats();
        }
    } catch (error) {
        console.error("Spielstand konnte nicht gespeichert werden:", error);
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

// Zeile auf der Sieg-Tafel: Levelsystem wie am Umschalter, Level, Runde, Farbe.
function siegLevelText(level) {
    const runde = rundeVon(level);

    return `${satzName()} · Level ${level} · Runde ${runde} · ${farbenName(runde)}`;
}

// Zeigt die goldene Tafel mit Level, Punkten und Spruch.
function zeigeSiegEhrung(level, punkte) {
    const tafel = getElement("siegEhrung");

    if (!tafel) return;

    const levelFeld = getElement("siegLevel");
    const punkteFeld = getElement("siegPunkte");

    if (levelFeld) {
        levelFeld.textContent = Number.isFinite(level) ? siegLevelText(level) : "";
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
// INITIALISIERUNG
// --------------------------------------------------

function initializeGame() {

    try {
        gameData = loadStorage();
        // loadStorage ergaenzt fehlende Einstellungen aus den Standardwerten.
        // Die Level-Auswahl zeigt denselben Satz wie das laufende Spiel.
        auswahlSatz = FORMEN_SATZ;
        currentLevel = satzDaten().currentLevel || 1;
    } catch (error) {
        console.error("Speicherfehler:", error);

        gameData = createDefaultStorage();
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

    const mixButton = getElement("mixButton");

    if (mixButton) {
        mixButton.addEventListener("click", mischeMotive);
    }

    // Zurück zur Spiele-Übersicht: die Online-Fassung auf Render, derselbe
    // Link, den spiele.html im Eintrag "Trashygeckos Connector" verwendet.
    const SPIELE_ONLINE = "https://trashygecko.onrender.com/connector";

    const backButton = getElement("backButton");

    if (backButton) {
        backButton.addEventListener("click", () => {
            window.location.href = SPIELE_ONLINE;
        });
    }

    // Nur in der Android-App: "Spiel verlassen" beendet die App ueber die
    // native Bruecke (MainActivity). Im Browser gibt es kein Verlassen -
    // dort bleibt der Knopf versteckt, er ist im HTML mit hidden angelegt.
    const verlassenButton = getElement("verlassenButton");
    const appBruecke = window.TMahjong;

    if (verlassenButton && appBruecke && typeof appBruecke.verlassen === "function") {
        verlassenButton.hidden = false;
        verlassenButton.addEventListener("click", () => appBruecke.verlassen());
    }

    // Level-Auswahl öffnen und schließen: ein Knopf, darin der Umschalter
    // zwischen Handy- und Normal-Leveln.
    const levelButton = getElement("levelButton");
    const levelPicker = getElement("levelPicker");
    const levelPickerClose = getElement("levelPickerClose");
    const satzUmschalter = getElement("satzUmschalter");

    if (levelButton) {
        levelButton.addEventListener("click", oeffneLevelAuswahl);
    }

    if (satzUmschalter) {
        satzUmschalter.addEventListener("click", event => {
            const knopf = event.target.closest("button[data-satz]");

            if (!knopf) return;

            setzeAuswahlSatz(Number(knopf.dataset.satz));
        });
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

    // Einstellungen oeffnen und schliessen.
    const settingsButton = getElement("settingsButton");
    const einstellungen = getElement("einstellungen");
    const einstellungenClose = getElement("einstellungenClose");

    if (settingsButton) {
        settingsButton.addEventListener("click", oeffneEinstellungen);
    }

    if (einstellungenClose) {
        einstellungenClose.addEventListener("click", schliesseEinstellungen);
    }

    if (einstellungen) {
        einstellungen.addEventListener("click", event => {
            if (event.target === einstellungen) {
                schliesseEinstellungen();
            }
        });
    }

    document.addEventListener("keydown", event => {
        if (event.key === "Escape") {
            schliesseLevelAuswahl();
            schliesseEinstellungen();
        }
    });

    // Steinform, Alterung und Hintergrund aus dem Spielstand uebernehmen.
    baueStilwahl();
    baueHintergrundwahl();
    baueModuswahl();
    setzeModus(modusWert(), false);
    zeigeAuswahlSatz();
    setzeStil(aktuellerStil().form, aktuellerStil().zustand, false);
    setzeHintergrund(hintergrundWert(), false);

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

