
/*
 * TRASHYGECKO MAHJONG
 * Spielbrett mit Zoom und Verschiebung
 */

const TILE_WIDTH = 64;
const TILE_HEIGHT = 76;
const STEP_X = 52;
const STEP_Y = 62;
const LAYER_OFFSET = 8;

let boardZoom = 1;
let boardOffsetX = 0;
let boardOffsetY = 0;
let boardViewport = null;
let boardStage = null;
let boardPointerStart = null;
let boardInitialFit = true;
let letzteBuehne = "";
let boardSuppressClick = false;
let boardPinch = null;
let letzteBreite = 0;
let letzteHoehe = 0;

function updateBoardTransform() {
    if (!boardStage) return;

    boardStage.style.transform =
        `translate(${boardOffsetX}px, ${boardOffsetY}px) scale(${boardZoom})`;
}

// Das Spielfeld nimmt immer den ganzen Bereich ein - hier ist nichts
// nachzujustieren.
function passeBuehneAn() {}

function fitBoard() {
    if (!boardViewport || !boardStage) return;

    const width = parseFloat(boardStage.style.width);
    const height = parseFloat(boardStage.style.height);

    const availableWidth = boardViewport.clientWidth - 16;
    const availableHeight = boardViewport.clientHeight - 16;

    const breitenZoom = availableWidth / width;
    const hoehenZoom = availableHeight / height;

    // Beim Start und nach dem Drehen ist das ganze Muster zu sehen.
    // Vergroessern geht mit zwei Fingern bzw. dem Mausrad.
    boardZoom = Math.max(
        0.25,
        Math.min(breitenZoom, hoehenZoom, 1.2),
        // Im Querformat ist die Hoehe sehr knapp - dann lieber groesser
        // bleiben und wischen, statt auf Briefmarkengroesse zu schrumpfen.
        Math.min(breitenZoom, 0.5)
    );

    passeBuehneAn();

    boardOffsetX =
        (boardViewport.clientWidth - width * boardZoom) / 2;
    boardOffsetY =
        (boardViewport.clientHeight - height * boardZoom) / 2;

    letzteBreite = boardViewport.clientWidth;
    letzteHoehe = boardViewport.clientHeight;

    updateBoardTransform();
}

// Nach Drehen des Geraets oder Groessenaenderung neu einpassen. Der Zoom des
// Spielers bleibt erhalten, solange das Brett weiterhin passt.
let groessenTimer = null;

function passeNachGroesseAn() {
    if (!boardViewport || !boardStage) return;

    const geaendert =
        boardViewport.clientWidth !== letzteBreite ||
        boardViewport.clientHeight !== letzteHoehe;

    if (geaendert) {
        fitBoard();
        return;
    }

    updateBoardTransform();
}

function planeGroessenPruefung() {
    clearTimeout(groessenTimer);
    groessenTimer = setTimeout(passeNachGroesseAn, 120);
}

window.addEventListener("resize", planeGroessenPruefung);
window.addEventListener("orientationchange", planeGroessenPruefung);

if (window.visualViewport) {
    window.visualViewport.addEventListener("resize", planeGroessenPruefung);
}

function changeBoardZoom(amount) {
    if (!boardViewport || !boardStage) return;

    const oldZoom = boardZoom;
    const newZoom = Math.min(2.5, Math.max(0.3, oldZoom + amount));

    if (newZoom === oldZoom) return;

    const centerX = boardViewport.clientWidth / 2;
    const centerY = boardViewport.clientHeight / 2;

    const boardX = (centerX - boardOffsetX) / oldZoom;
    const boardY = (centerY - boardOffsetY) / oldZoom;

    boardZoom = newZoom;

    boardOffsetX = centerX - boardX * newZoom;
    boardOffsetY = centerY - boardY * newZoom;

    passeBuehneAn();
    updateBoardTransform();
}

// Doppelklick / Doppeltipp: das ganze Brett ins Bild holen. Der Startzoom
// fuellt die Flaeche, diese Ansicht zeigt alles auf einmal.
function resetBoardView() {
    if (!boardViewport || !boardStage) return;

    const width = parseFloat(boardStage.style.width);
    const height = parseFloat(boardStage.style.height);

    const breitenZoom = (boardViewport.clientWidth - 16) / width;
    const hoehenZoom = (boardViewport.clientHeight - 16) / height;

    boardZoom = Math.max(
        0.25,
        Math.min(breitenZoom, hoehenZoom, 1.2)
    );

    boardOffsetX =
        (boardViewport.clientWidth - width * boardZoom) / 2;
    boardOffsetY =
        (boardViewport.clientHeight - height * boardZoom) / 2;

    updateBoardTransform();
}

function enableBoardMovement(viewport) {
    const aktiveZeiger = new Map();

    function kasten() {
        return viewport.getBoundingClientRect();
    }

    function entfernung(a, b) {
        return Math.hypot(a.x - b.x, a.y - b.y);
    }

    function mittelpunkt(a, b) {
        return { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 };
    }

    viewport.addEventListener("pointerdown", event => {
        aktiveZeiger.set(event.pointerId, {
            x: event.clientX,
            y: event.clientY
        });

        if (aktiveZeiger.size === 1) {
            boardPointerStart = {
                x: event.clientX,
                y: event.clientY,
                offsetX: boardOffsetX,
                offsetY: boardOffsetY,
                moved: false
            };

            // Den Zeiger noch NICHT festhalten: Sonst landet der Klick beim
            // Verschiebe-Bereich statt beim Stein. Erst beim echten Ziehen.
            return;
        }

        // Ab dem zweiten Finger wird nicht mehr gewischt, sondern gezoomt.
        boardPointerStart = null;

        if (aktiveZeiger.size === 2) {
            const [a, b] = [...aktiveZeiger.values()];
            const rahmen = kasten();
            const mitte = mittelpunkt(a, b);

            boardPinch = {
                abstand: Math.max(1, entfernung(a, b)),
                zoom: boardZoom,
                x: mitte.x - rahmen.left,
                y: mitte.y - rahmen.top,
                offsetX: boardOffsetX,
                offsetY: boardOffsetY
            };
        }
    });

    viewport.addEventListener("pointermove", event => {
        if (!aktiveZeiger.has(event.pointerId)) return;

        aktiveZeiger.set(event.pointerId, {
            x: event.clientX,
            y: event.clientY
        });

        // Zwei Finger: um den Mittelpunkt der Finger zoomen.
        if (boardPinch && aktiveZeiger.size >= 2) {
            const [a, b] = [...aktiveZeiger.values()];
            const rahmen = kasten();
            const mitte = mittelpunkt(a, b);
            const faktor = entfernung(a, b) / boardPinch.abstand;

            const neuerZoom = Math.min(
                2.5,
                Math.max(0.3, boardPinch.zoom * faktor)
            );

            const brettX =
                (boardPinch.x - boardPinch.offsetX) / boardPinch.zoom;
            const brettY =
                (boardPinch.y - boardPinch.offsetY) / boardPinch.zoom;

            boardZoom = neuerZoom;

            boardOffsetX = (mitte.x - rahmen.left) - brettX * neuerZoom;
            boardOffsetY = (mitte.y - rahmen.top) - brettY * neuerZoom;

            passeBuehneAn();
            updateBoardTransform();

            boardSuppressClick = true;
            return;
        }

        if (!boardPointerStart || aktiveZeiger.size > 1) return;

        const dx = event.clientX - boardPointerStart.x;
        const dy = event.clientY - boardPointerStart.y;

        if (Math.abs(dx) > 4 || Math.abs(dy) > 4) {
            if (!boardPointerStart.moved) {
                boardPointerStart.moved = true;

                // Jetzt ziehen wir wirklich: Zeiger uebernehmen, damit das
                // Verschieben auch ausserhalb des Bretts weiterlaeuft.
                try {
                    viewport.setPointerCapture(event.pointerId);
                } catch (fehler) {}
            }
        }

        if (boardPointerStart.moved) {
            boardOffsetX = boardPointerStart.offsetX + dx;
            boardOffsetY = boardPointerStart.offsetY + dy;

            updateBoardTransform();
        }
    });

    function beendeBewegung(warAmZoomen = false) {
        if (boardPointerStart?.moved || warAmZoomen) {
            boardSuppressClick = true;

            setTimeout(() => {
                boardSuppressClick = false;
            }, 150);
        }

        boardPointerStart = null;
    }

    function zeigerWeg(event) {
        aktiveZeiger.delete(event.pointerId);

        const warAmZoomen = !!boardPinch;

        if (aktiveZeiger.size < 2) {
            boardPinch = null;
        }

        if (aktiveZeiger.size === 0) {
            beendeBewegung(warAmZoomen);
        }
    }

    viewport.addEventListener("pointerup", zeigerWeg);
    viewport.addEventListener("pointercancel", zeigerWeg);

    viewport.addEventListener("click", event => {
        if (boardSuppressClick) {
            event.preventDefault();
            event.stopPropagation();
            boardSuppressClick = false;
        }
    }, true);

    viewport.addEventListener("wheel", event => {
        event.preventDefault();

        changeBoardZoom(event.deltaY < 0 ? 0.1 : -0.1);
    }, { passive: false });
}

function renderBoard(board, selectedId = null, onTileClick = null, hinweisIds = []) {
    const container = document.getElementById("board");

    if (!container) return;

    container.innerHTML = "";

    if (!board || board.length === 0) return;

    // Sind alle Steine weg (Sieg), bleibt nur die Meldung stehen.
    if (!board.some(tile => !tile.removed)) {
        container.style.display = "none";
        setzeFeldGroesse(false);
        return;
    }

    container.style.display = "";
    setzeFeldGroesse(true);

    const maxX = Math.max(...board.map(tile => tile.x));
    const maxY = Math.max(...board.map(tile => tile.y));
    const maxZ = Math.max(...board.map(tile => tile.z));

    const boardWidth =
        maxX * STEP_X + TILE_WIDTH + maxZ * LAYER_OFFSET;

    const boardHeight =
        maxY * STEP_Y + TILE_HEIGHT + maxZ * LAYER_OFFSET;

    const viewport = document.createElement("div");

    viewport.className = "board-viewport";

    viewport.style.cssText = `
        position:relative;
        flex:1 1 auto;
        width:100%;
        height:100%;
        min-height:0;
        overflow:hidden;
        touch-action:none;
        cursor:grab;
        background:transparent;
        user-select:none;
    `;

    const stage = document.createElement("div");

    stage.className = "board-stage";

    stage.style.cssText = `
        position:absolute;
        left:0;
        top:0;
        width:${boardWidth}px;
        height:${boardHeight}px;
        transform-origin:top left;
        will-change:transform;
    `;

    board.forEach(tile => {
        if (tile.removed) return;

        const button = document.createElement("button");

        button.type = "button";
        button.className = "mahjong-tile";
        button.dataset.tileId = tile.id;

        const free = isTileFree(tile, board);

        if (!free) button.classList.add("blocked");

        if (tile.id === selectedId) {
            button.classList.add("selected");
        }

        if (hinweisIds.includes(tile.id)) {
            button.classList.add("hinweis");
        }

        button.style.left =
            `${tile.x * STEP_X + tile.z * LAYER_OFFSET}px`;

        button.style.top =
            `${tile.y * STEP_Y - tile.z * LAYER_OFFSET}px`;

        button.style.zIndex =
            String(tile.z * 100 + tile.y * 12 + tile.x);

        button.disabled = !free;

        const image = document.createElement("img");

        image.className = "tile-symbol";
        image.src =
            `assets/stein (${parseInt(String(tile.symbolId).replace(/\D/g, ""), 10)}).png`;

        image.alt = `Mahjong-Stein ${tile.symbolId}`;
        image.draggable = false;

        image.onerror = () => {
            image.replaceWith(document.createTextNode("?"));
        };

        button.appendChild(image);

        if (typeof onTileClick === "function") {
            button.addEventListener("click", () => {
                onTileClick(tile.id);
            });
        }

        stage.appendChild(button);
    });

    viewport.appendChild(stage);
    container.appendChild(viewport);

    boardViewport = viewport;
    boardStage = stage;

    enableBoardMovement(viewport);

    const buehnenMass = boardStage.style.width + "x" + boardStage.style.height;

    // Neu vermessen beim ersten Aufbau und immer dann, wenn sich die
    // Brettgroesse aendert (z. B. nach einem Levelwechsel). Innerhalb eines
    // Levels bleibt der Zoom des Spielers erhalten.
    if (boardInitialFit || buehnenMass !== letzteBuehne) {
        fitBoard();
        boardInitialFit = false;
    } else {
        passeBuehneAn();
        updateBoardTransform();
    }

    letzteBuehne = buehnenMass;

    viewport.addEventListener("dblclick", event => {
        if (!event.target.closest(".mahjong-tile")) {
            resetBoardView();
        }
    });
}

// Aktualisiert ein bestehendes Brett, ohne alle Steine neu zu bauen.
// Wichtig fuer Handys: Bei jedem Tipp wurden vorher 144 Steine ersetzt.
// Nach dem Sieg schrumpft der Rahmen, sonst fuellt er den ganzen Platz.
function setzeFeldGroesse(gross) {
    const panel = document.querySelector(".board-area");

    if (!panel) return;

    panel.classList.toggle("kompakt", !gross);
}

function updateBoardView(board, selectedId = null, hinweisIds = [], onTileClick = null) {
    const container = document.getElementById("board");

    if (!container || !boardStage) {
        renderBoard(board, selectedId, onTileClick, hinweisIds);
        return;
    }

    // Sind alle Steine weg (Sieg), bleibt nur die Meldung stehen.
    if (!board || !board.some(tile => !tile.removed)) {
        container.style.display = "none";
        setzeFeldGroesse(false);
        return;
    }

    container.style.display = "";
    setzeFeldGroesse(true);

    const elemente = new Map();

    boardStage.querySelectorAll(".mahjong-tile").forEach(element => {
        elemente.set(Number(element.dataset.tileId), element);
    });

    board.forEach(tile => {
        const element = elemente.get(tile.id);

        if (!element) return;

        if (tile.removed) {
            element.remove();
            return;
        }

        const frei = isTileFree(tile, board);

        element.disabled = !frei;
        element.classList.toggle("blocked", !frei);
        element.classList.toggle("selected", tile.id === selectedId);
        element.classList.toggle("hinweis", hinweisIds.includes(tile.id));
    });
}

function clearBoard() {
    const container = document.getElementById("board");

    if (container) {
        container.innerHTML = "";
    }

    boardViewport = null;
    boardStage = null;
}
