/*
 * TRASHYGECKO MAHJONG
 * Levelauswahl und Einstellungen samt ihrer Inhalte
 *
 * Beide Dialoge: Oeffnen/Schliessen, Fokusfuehrung, das Raster der Level, die
 * Farbwahl und die Einstellungen (Steinform, Steinalterung, Hintergrund).
 * Die Auswahl-Zustaende (auswahlSatz, auswahlRunde) liegen hier, die Dialoge
 * selbst stehen im HTML.
 */

// --------------------------------------------------
// LEVEL-AUSWAHL
// --------------------------------------------------

// Runde, die in der Level-Auswahl gerade angezeigt wird.
let auswahlRunde = 1;

// Formensatz, den die Level-Auswahl zeigt (1 = PC, 2 = Handy).
// Jedes Spiel startet mit den Handy-Leveln, die Auswahl beginnt also dort
// (Hauptsteuerung setzt auswahlSatz beim Start auf FORMEN_SATZ).
let auswahlSatz = 2;

// Fokus zurueck auf den Knopf, der den Dialog geoeffnet hat - sonst landet er
// nach dem Schliessen irgendwo im Nichts und die Tastaturbedienung haengt.
let fokusVorDialog = null;

function merkeFokus() {
    fokusVorDialog = document.activeElement;
}

function gebeFokusZurueck() {
    if (fokusVorDialog) fokusVorDialog.focus();

    fokusVorDialog = null;
}

function oeffneLevelAuswahl() {
    const auswahl = getElement("levelPicker");

    if (!auswahl) return;

    merkeFokus();
    setzeAuswahlSatz(auswahlSatz);

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

    if (!auswahl || auswahl.hidden) return;

    auswahl.hidden = true;

    gebeFokusZurueck();
}

// Einstellungen: Steine, Hintergrund und Rueckweg zur Spieleuebersicht.
function oeffneEinstellungen() {
    const dialog = getElement("einstellungen");

    if (!dialog) return;

    // Zeigt den tatsaechlich gespeicherten Hintergrund und markiert ihn.
    setzeHintergrund(hintergrundWert(), false);

    merkeFokus();

    dialog.hidden = false;

    // Erster Bedienschritt im Dialog, damit die Tastatur dort ankommt.
    const erstesFeld = dialog.querySelector(".stil-gruppe button");

    if (erstesFeld) erstesFeld.focus();
}

function schliesseEinstellungen() {
    const dialog = getElement("einstellungen");

    if (!dialog || dialog.hidden) return;

    dialog.hidden = true;

    gebeFokusZurueck();
}

function baueLevelAuswahl(runde = auswahlRunde) {
    const raster = getElement("levelGrid");

    if (!raster) return;

    auswahlRunde = Math.max(1, runde);

    raster.innerHTML = "";

    const formen = hoechstesLevel(auswahlSatz);
    const start = (auswahlRunde - 1) * formen;

    // Drei Zustaende: freigeschaltet, naechstes Level, gesperrt.  Gesperrte
    // Level bleiben stehen - abgedunkelt, mit Schloss und Rundenfarbe.
    const daten = satzDaten(auswahlSatz, spielModus);
    const frei = freiesLevel(daten, auswahlSatz);
    const naechstes = naechstesLevel(daten, auswahlSatz);

    for (let form = 1; form <= formen; form++) {
        const level = start + form;
        const feld = document.createElement("button");
        const bestzeit = bestzeitVon(level, auswahlSatz);
        // Das Raster zeigt genau eine Runde - levelRunde ist die gewaehlte.
        const levelRunde = auswahlRunde;
        const gesperrt = level > frei;

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
        } else if (level === naechstes) {
            // Das naechste offene Level: dezente Umrandung statt Goldfuellung.
            feld.classList.add("naechstes");
        }

        if (gesperrt) {
            feld.classList.add("gesperrt");
            feld.dataset.farbe = String(levelRunde);

            // Nicht anklickbar - und ohne Fokus, damit die Tastatur nicht auf
            // einem gesperrten Feld haengen bleibt.
            feld.disabled = true;
            feld.tabIndex = -1;

            const schloss = document.createElement("span");

            schloss.className = "level-feld-schloss";
            schloss.setAttribute("aria-hidden", "true");
            feld.appendChild(schloss);
        }

        // Tooltip in derselben Sprache wie der Umschalter: System, Runde, Farbe.
        const titel = [
            satzName(auswahlSatz),
            `Runde ${levelRunde}`,
            farbenName(levelRunde),
            formName(level, auswahlSatz),
            bestzeit !== null ? `Bestzeit ${formatTime(bestzeit)}` : null,
            gesperrt ? `gesperrt · ab Level ${frei} spielbar` : null,
        ];

        feld.title = titel.filter(Boolean).join(" · ");

        if (!gesperrt) feld.addEventListener("click", () => waehleLevel(level));

        raster.appendChild(feld);
    }

    setzeWeiterKnopf();

    const anzeige = getElement("rundeAnzeige");

    if (anzeige) {
        anzeige.textContent =
            `${variantenText(auswahlSatz, spielModus)} · Runde ${auswahlRunde} · ${farbenName(auswahlRunde)} · ` +
            `${geschafftInRunde(auswahlRunde, auswahlSatz)} von ${formen}`;
        anzeige.dataset.stufe = stufenWert(spielModus);
    }

    baueFarbwahl();
    markiereAktuellesLevel();
}

// Beschriftung des WEITER-Knopfs in der Auswahl: nennt das Ziel direkt, damit
// man vor dem Klick weiss, welches Level kommt.
function setzeWeiterKnopf() {
    const knopf = getElement("weiterButton");

    if (!knopf) return;

    const daten = satzDaten(auswahlSatz, spielModus);
    const ziel = naechstesLevel(daten, auswahlSatz);

    knopf.disabled = ziel === 0;
    knopf.textContent = ziel
        ? `WEITER · LEVEL ${String(ziel).padStart(2, "0")}`
        : "ALLES GESCHAFFT";
}

// Farbwahl am Ende der Level-Auswahl: jede Farbe ist eine Runde.
function baueFarbwahl() {
    const leiste = getElement("farbeLeiste");

    if (!leiste) return;

    leiste.innerHTML = "";

    RUNDEN_FARBEN.forEach((name, index) => {
        const runde = index + 1;
        const abLevel = (runde - 1) * hoechstesLevel(auswahlSatz) + 1;

        const feld = document.createElement("button");

        feld.type = "button";
        feld.className = "farbe-feld";
        feld.dataset.farbe = String(runde);
        const geschafft = geschafftInRunde(runde, auswahlSatz);

        feld.title =
            `${satzName(auswahlSatz)} · Runde ${runde} · ${name} · ab Level ${abLevel} · ` +
            `${geschafft} von ${hoechstesLevel(auswahlSatz)} geschafft`;
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
    // Nur der laufende Satz hat ein aktuelles Level - der andere zeigt nichts.
    const imSpiel = auswahlSatz === FORMEN_SATZ;

    document.querySelectorAll(".level-feld").forEach(feld => {
        feld.classList.toggle(
            "aktuell",
            imSpiel && Number(feld.dataset.level) === currentLevel
        );
    });
}

// Zeigt einen der beiden Levelsaetze in der Auswahl und markiert ihn am
// Umschalter.  Der laufende Satz (FORMEN_SATZ) bleibt davon unberuehrt.
function setzeAuswahlSatz(satz) {
    auswahlSatz = satz === 1 ? 1 : 2;

    const level = satzDaten(auswahlSatz).currentLevel || 1;

    baueLevelAuswahl(rundeVon(level, auswahlSatz));
    zeigeAuswahlSatz();
}

// Hebt den aktiven Umschalter hervor (Klasse wie bei der Steinform).
function zeigeAuswahlSatz() {
    document.querySelectorAll("#satzUmschalter button").forEach(knopf => {
        const an = Number(knopf.dataset.satz) === auswahlSatz;

        knopf.classList.toggle("aktiv", an);
        knopf.setAttribute("aria-pressed", an ? "true" : "false");
    });
}

function waehleLevel(level) {
    // Level laufen endlos: nach der letzten Form geht es in der naechsten Runde weiter.
    const ziel = Math.floor(Number(level));

    // Ungueltige Angaben (etwa aus einem alten Spielstand) nicht stillschweigend
    // als Level 1 starten, sondern sichtbar ablehnen.
    if (!Number.isFinite(ziel) || ziel < 1) {
        setMessage("Dieses Level gibt es nicht.");

        return;
    }

    // Gesperrte Level sind auch hier zu: der Knopf ist zwar abgeschaltet, die
    // Pruefung bleibt trotzdem, falls das Feld anders ausgeloest wird.
    if (ziel > freiesLevel(satzDaten(auswahlSatz, spielModus), auswahlSatz)) {
        setMessage("Dieses Level ist noch gesperrt.");

        return;
    }

    // Gewaehlt wird immer im Satz, den die Auswahl gerade zeigt.
    FORMEN_SATZ = auswahlSatz;
    satzDaten(auswahlSatz).currentLevel = ziel;
    speichereSpielstand();

    currentLevel = ziel;

    schliesseLevelAuswahl();
    startGame();
}

// Bestzeit eines Levels in Sekunden - null, wenn es noch nicht geschafft ist.
function bestzeitVon(level, satz = FORMEN_SATZ) {
    const daten = gameData ? satzDaten(satz) : null;
    const wert = daten && daten.bestTimes ? daten.bestTimes[level] : undefined;

    return Number.isFinite(wert) ? wert : null;
}

// Wie viele Level dieser Runde sind schon geschafft?
function geschafftInRunde(runde, satz = FORMEN_SATZ) {
    const formen = hoechstesLevel(satz);
    const start = (runde - 1) * formen;
    let anzahl = 0;

    for (let form = 1; form <= formen; form++) {
        if (bestzeitVon(start + form, satz) !== null) anzahl++;
    }

    return anzahl;
}

// --------------------------------------------------
// HINTERGRUND
// --------------------------------------------------

// Dunkle, augenschonende Farben - Mitte und Rand des Strahlenverlaufs.
// Schrift und Bedienelemente sind hell und bleiben auf allen Varianten lesbar.
const HINTERGRUENDE = [
    { wert: "gruen", name: "Dunkelgrün", mitte: "#1d4b3e", rand: "#0a211b" },
    { wert: "blau", name: "Dunkelblau", mitte: "#1c3550", rand: "#0a1521" },
    { wert: "anthrazit", name: "Anthrazit", mitte: "#2a2e33", rand: "#121417" },
    { wert: "schiefer", name: "Schiefer", mitte: "#26313f", rand: "#0e1319" },
    { wert: "petrol", name: "Petrol", mitte: "#173d44", rand: "#08191d" },
    { wert: "rot", name: "Dunkelrot", mitte: "#4a2024", rand: "#1c0b0d" },
    { wert: "violett", name: "Dunkelviolett", mitte: "#35244a", rand: "#150e1f" },
    { wert: "braun", name: "Dunkelbraun", mitte: "#3d2c1e", rand: "#18110a" },
];

// Gespeicherte Wahl; "gruen" ist der Standard aus dem Stylesheet.
function hintergrundWert() {
    const gespeichert = gameData && gameData.settings ? gameData.settings.hintergrund : null;

    return HINTERGRUENDE.some(thema => thema.wert === gespeichert) ? gespeichert : "gruen";
}

// Setzt den Hintergrund sofort - ohne Neuladen.
function setzeHintergrund(wert, speichern = true) {
    const thema = HINTERGRUENDE.find(eintrag => eintrag.wert === wert) || HINTERGRUENDE[0];

    // Der Standard steht im Stylesheet; alles andere liegt direkt am Element.
    // Die Spielflaeche hat einen eigenen Verlauf und wird mitgefaerbt.
    const standard = thema.wert === "gruen";
    const flaeche = getElement("boardArea");

    document.body.style.background = standard
        ? ""
        : `radial-gradient(ellipse at center, ${thema.mitte} 0%, ${thema.rand} 78%)`;

    if (flaeche) {
        flaeche.style.background = standard
            ? ""
            : `radial-gradient(ellipse at center, ${thema.mitte}, ${thema.rand} 80%)`;
    }

    document.querySelectorAll("#hintergrundLeiste button").forEach(knopf => {
        const an = knopf.dataset.wert === thema.wert;

        knopf.classList.toggle("aktiv", an);
        knopf.setAttribute("aria-pressed", an ? "true" : "false");
    });

    if (!speichern || !gameData) return;

    gameData.settings = gameData.settings || {};
    gameData.settings.hintergrund = thema.wert;

    speichereSpielstand();
}

// Farbkreise fuer die Hintergrundwahl.
function baueHintergrundwahl() {
    const leiste = getElement("hintergrundLeiste");

    if (!leiste) return;

    leiste.innerHTML = "";

    HINTERGRUENDE.forEach(thema => {
        const feld = document.createElement("button");

        feld.type = "button";
        feld.className = "hintergrund-feld";
        feld.dataset.wert = thema.wert;
        feld.title = thema.name;
        feld.setAttribute("aria-label", thema.name);
        feld.style.background =
            `radial-gradient(circle at 50% 40%, ${thema.mitte} 0%, ${thema.rand} 100%)`;

        feld.addEventListener("click", () => setzeHintergrund(thema.wert));

        leiste.appendChild(feld);
    });
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

        speichereSpielstand();
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

// --------------------------------------------------
// SCHWIERIGKEIT
// --------------------------------------------------

// "leicht" ist das bisherige Spiel, "schwer" der anspruchsvollere Aufbau.
const MODI = ["leicht", "schwer"];

// Gespeicherte Wahl; ohne Eintrag bleibt es beim bisherigen Spiel.
function modusWert() {
    const gespeichert =
        gameData && gameData.settings ? gameData.settings.schwierigkeit : null;

    return gespeichert === "schwer" ? "schwer" : "leicht";
}

// Setzt den Grad fuer die laufende Partie und merkt ihn fuer die naechste.
function setzeModus(wert, speichern = true) {
    const neu = MODI.includes(wert) ? wert : "leicht";
    const gewechselt = neu !== spielModus;

    spielModus = neu;

    // Das Stylesheet haengt die Knopfreihe daran: in Schwer steht dort ein
    // vierter Knopf (MISCHEN), in Leicht bleibt die Reihe wie bisher.
    document.body.dataset.modus = neu;

    document.querySelectorAll("#modusLeiste button, #stufenUmschalter button").forEach(knopf => {
        const an = knopf.dataset.wert === neu;

        knopf.classList.toggle("aktiv", an);
        knopf.setAttribute("aria-pressed", an ? "true" : "false");
    });

    zeigeMixKnopf();
    aktualisiereMixKnopf();

    // Die Varianten-Anzeige nennt die Stufe - sie wechselt mit.
    setzeVariantenAnzeige();

    // Ein Wechsel mitten in der Partie laesst das Brett stehen und gilt sofort.
    if (gewechselt && gameRunning) {
        setMessage(
            neu === "schwer"
                ? "Schwierigkeit Schwer – das Brett bleibt, MISCHEN ist jetzt verfügbar."
                : "Schwierigkeit Leicht – das Brett bleibt."
        );
    }

    if (!speichern || !gameData) return;

    gameData.settings = gameData.settings || {};
    gameData.settings.schwierigkeit = neu;

    speichereSpielstand();
}

// LEICHT/SCHWER gibt es an zwei Stellen: in den Einstellungen und - fuer alle
// vier Varianten an einer Stelle - direkt in der Levelauswahl.  Beide Knoepfe
// rufen denselben Wechsel auf; in der Auswahl wird danach das Raster neu
// gebaut, damit Haken und Bestzeiten zur gewaehlten Stufe gehoeren.
function baueModuswahl() {
    const leisten = [getElement("modusLeiste"), getElement("stufenUmschalter")];

    leisten.forEach(leiste => {
        if (!leiste) return;

        leiste.querySelectorAll("button[data-wert]").forEach(knopf => {
            knopf.addEventListener("click", () => {
                setzeModus(knopf.dataset.wert);

                if (leiste.id === "stufenUmschalter") baueLevelAuswahl(auswahlRunde);
            });
        });
    });

    const weiter = getElement("weiterButton");

    if (weiter) {
        weiter.addEventListener("click", () => weiterSpielen(auswahlSatz, spielModus));
    }
}
