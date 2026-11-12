
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
let boardSuppressClick = false;

function updateBoardTransform() {
    if (!boardStage) return;

    boardStage.style.transform =
        `translate(${boardOffsetX}px, ${boardOffsetY}px) scale(${boardZoom})`;
}

function fitBoard() {
    if (!boardViewport || !boardStage) return;

    const width = parseFloat(boardStage.style.width);
    const height = parseFloat(boardStage.style.height);

    const availableWidth = boardViewport.clientWidth - 30;
    const availableHeight = boardViewport.clientHeight - 30;

    boardZoom = Math.min(
        availableWidth / width,
        availableHeight / height,
        1
    );

    boardZoom = Math.max(0.3, boardZoom);

    boardOffsetX = (boardViewport.clientWidth - width * boardZoom) / 2;
    boardOffsetY = (boardViewport.clientHeight - height * boardZoom) / 2;

    updateBoardTransform();
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

    updateBoardTransform();
}

function resetBoardView() {
    fitBoard();
}

function createZoomControls(viewport) {
    const controls = document.createElement("div");

    controls.style.cssText = `
        position:absolute;
        top:10px;
        right:10px;
        z-index:10000;
        display:flex;
        gap:6px;
    `;

    function makeButton(label, title, action) {
        const button = document.createElement("button");

        button.type = "button";
        button.textContent = label;
        button.title = title;

        button.style.cssText = `
            width:42px;
            height:38px;
            border:1px solid #d5b36a;
            border-radius:7px;
            background:#102b27;
            color:#f5edda;
            font-size:22px;
            font-weight:bold;
            cursor:pointer;
            box-shadow:0 2px 6px #0008;
        `;

        button.addEventListener("click", event => {
            event.stopPropagation();
            action();
        });

        button.addEventListener("pointerdown", event => {
            event.stopPropagation();
        });

        return button;
    }

    controls.appendChild(
        makeButton("+", "Vergrößern", () => changeBoardZoom(0.15))
    );

    controls.appendChild(
        makeButton("−", "Verkleinern", () => changeBoardZoom(-0.15))
    );

    controls.appendChild(
        makeButton("↺", "Ansicht zurücksetzen", resetBoardView)
    );

    viewport.appendChild(controls);
}

function enableBoardMovement(viewport) {
    viewport.addEventListener("pointerdown", event => {
        if (event.target.closest(".board-zoom-controls")) return;

        boardPointerStart = {
            x: event.clientX,
            y: event.clientY,
            offsetX: boardOffsetX,
            offsetY: boardOffsetY,
            moved: false
        };

        viewport.setPointerCapture(event.pointerId);
    });

    viewport.addEventListener("pointermove", event => {
        if (!boardPointerStart) return;

        const dx = event.clientX - boardPointerStart.x;
        const dy = event.clientY - boardPointerStart.y;

        if (Math.abs(dx) > 4 || Math.abs(dy) > 4) {
            boardPointerStart.moved = true;
        }

        if (boardPointerStart.moved) {
            boardOffsetX = boardPointerStart.offsetX + dx;
            boardOffsetY = boardPointerStart.offsetY + dy;

            updateBoardTransform();
        }
    });

    function finishMovement() {
        if (boardPointerStart?.moved) {
            boardSuppressClick = true;

            setTimeout(() => {
                boardSuppressClick = false;
            }, 100);
        }

        boardPointerStart = null;
    }

    viewport.addEventListener("pointerup", finishMovement);
    viewport.addEventListener("pointercancel", finishMovement);

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

function renderBoard(board, selectedId = null, onTileClick = null) {
    const container = document.getElementById("board");

    if (!container) return;

    container.innerHTML = "";

    if (!board || board.length === 0) return;

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
        width:100%;
        height:70vh;
        min-height:440px;
        max-height:850px;
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

    createZoomControls(viewport);
    enableBoardMovement(viewport);

    if (boardInitialFit) {
        fitBoard();
        boardInitialFit = false;
    } else {
        updateBoardTransform();
    }

    viewport.addEventListener("dblclick", event => {
        if (!event.target.closest(".mahjong-tile")) {
            resetBoardView();
        }
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
