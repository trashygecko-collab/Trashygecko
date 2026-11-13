
import { BOARD_LAYOUT } from "./board-layout.js";

// Erstellt einen neuen Editor-Spielstand
export function createEditorLayout() {
    return BOARD_LAYOUT.positions.map((position, index) => ({
        id: index,
        x: position.x,
        y: position.y,
        z: position.z
    }));
}

// Fügt einen Stein an einer Position hinzu
export function addEditorTile(layout, x, y, z) {
    if (!Number.isInteger(x) ||
        !Number.isInteger(y) ||
        !Number.isInteger(z) ||
        x < 0 || y < 0 || z < 0) {
        return false;
    }

    if (layout.some(tile =>
        tile.x === x &&
        tile.y === y &&
        tile.z === z
    )) {
        return false;
    }

    const nextId = layout.length > 0
        ? Math.max(...layout.map(tile => tile.id)) + 1
        : 0;

    layout.push({ id: nextId, x, y, z });

    return true;
}

// Entfernt einen Stein anhand seiner ID
export function removeEditorTile(layout, id) {
    const index = layout.findIndex(tile => tile.id === id);

    if (index === -1) {
        return false;
    }

    layout.splice(index, 1);

    return true;
}

// Verschiebt einen Stein
export function moveEditorTile(layout, id, x, y, z) {
    const tile = layout.find(item => item.id === id);

    if (!tile ||
        !Number.isInteger(x) ||
        !Number.isInteger(y) ||
        !Number.isInteger(z) ||
        x < 0 || y < 0 || z < 0) {
        return false;
    }

    const occupied = layout.some(other =>
        other.id !== id &&
        other.x === x &&
        other.y === y &&
        other.z === z
    );

    if (occupied) {
        return false;
    }

    tile.x = x;
    tile.y = y;
    tile.z = z;

    return true;
}

// Prüft ein Level auf grundlegende Fehler
export function validateEditorLayout(layout) {
    const errors = [];

    if (!Array.isArray(layout) || layout.length === 0) {
        return {
            valid: false,
            errors: ["Das Level enthält keine Steine."]
        };
    }

    const positions = new Set();
    const ids = new Set();

    for (const tile of layout) {
        if (![tile.x, tile.y, tile.z, tile.id].every(Number.isInteger)) {
            errors.push("Ein Stein besitzt ungültige Positionsdaten.");
            continue;
        }

        const position = `${tile.x},${tile.y},${tile.z}`;

        if (positions.has(position)) {
            errors.push(`Position doppelt belegt: ${position}`);
        }

        if (ids.has(tile.id)) {
            errors.push(`Stein-ID doppelt vergeben: ${tile.id}`);
        }

        positions.add(position);
        ids.add(tile.id);
    }

    if (layout.length !== 144) {
        errors.push(`Das Level hat ${layout.length} statt 144 Steine.`);
    }

    return {
        valid: errors.length === 0,
        errors
    };
}

// Exportiert ein Level als JSON-Text
export function exportEditorLayout(layout) {
    const validation = validateEditorLayout(layout);

    if (!validation.valid) {
        return {
            success: false,
            errors: validation.errors
        };
    }

    return {
        success: true,
        json: JSON.stringify({
            name: "Eigenes Mahjong-Level",
            tileCount: layout.length,
            positions: layout
        }, null, 2)
    };
}

// Erstellt eine unabhängige Kopie des Levels
export function copyEditorLayout(layout) {
    return layout.map(tile => ({ ...tile }));
}
