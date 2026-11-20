/*
 * TRASHYGECKO – „Zurück zum Connector“ für Einzelspieler-Spiele
 *
 * Eine Datei für alle Spiele: sie setzt einen kleinen Knopf in die Seite und
 * regelt die Rückkehr je nach Umgebung.
 *
 *   PC (Connector):        Das Spiel läuft im rechten Rahmen. Der Knopf meldet
 *                          sich beim Connector; der leert nur den Rahmen. Chat
 *                          und Spieleliste links bleiben, kein Neuladen.
 *   Handy / Tablet:        Das Spiel läuft im eigenen Tab. Der Knopf schließt
 *                          ihn (window.close). Verhindert der Browser das,
 *                          geht es wie bisher zurück zum Connector.
 *
 * Einbau: eine Zeile vor </body>, Pfad je nach Tiefe der Spielseite:
 *
 *   <script src="../zurueck.js"></script>
 *   <script src="../../zurueck.js"></script>
 *
 * Platzwahl über data-Attribute am Script-Tag (alle freiwillig):
 *
 *   data-ziel=".titelleiste"   Behälter, in den der Knopf kommt
 *   data-vor=".kreuz"          darin vor dieses Element
 *   data-anfang="1"            als erstes Kind des Behälters
 *   data-auf="karte"           als eigene Zeile oben in einer Karte
 *   data-nur="ohne-room"       nicht zeigen, wenn ?room= in der Adresse steht
 *   data-kein-knopf="1"        nur window.tgZurueck() bereitstellen
 *
 * Ohne data-ziel landet der Knopf in der Titelleiste des Fensters
 * (.titelleiste), sonst fest in der Ecke oben rechts.
 *
 * Für Spiele mit eigenem Zurück-Knopf: window.tgZurueck() aufrufen.
 */
(function () {
    "use strict";

    // In der Android-App (WebView mit Bruecke TMahjong) gibt es weder Connector
    // noch Rahmen: die App hat ihren eigenen Knopf "SPIEL VERLASSEN".  Hier also
    // gar nichts tun - kein Knopf, kein window.tgZurueck.
    if (window.TMahjong) return;

    // Nie zwei Knöpfe in einer Seite.
    if (document.querySelector(".tg-zurueck")) return;

    const skript = document.currentScript;
    const daten = skript ? skript.dataset : {};
    const CONNECTOR = "https://trashygecko.onrender.com/connector";

    // In einer laufenden Mehrspieler-Partie gehört das Verlassen dem Spiel.
    if (daten.nur === "ohne-room") {
        try {
            if (new URLSearchParams(location.search).has("room")) return;
        } catch (fehler) {
            return;
        }
    }

    // --------------------------------------------------
    // WOHIN ZURÜCK?
    // --------------------------------------------------

    // Connector-Adresse: die des Tabs, der das Spiel geöffnet hat (Referrer),
    // sonst dieselbe Herkunft wie die Seite, sonst die bekannte Online-Adresse.
    function connectorAdresse() {
        const herkunft = String(document.referrer || "");

        if (/\/connector(?:[/?#]|$)/.test(herkunft)) {
            return herkunft.replace(/[?#].*$/, "");
        }

        if (location.protocol === "http:" || location.protocol === "https:") {
            return location.origin + "/connector";
        }

        return CONNECTOR;
    }

    function imRahmen() {
        try {
            return Boolean(window.parent) && window.parent !== window;
        } catch (fehler) {
            // Fremder Rahmen: dann lieber wie ein eigener Tab behandeln.
            return false;
        }
    }

    function zurueck() {
        // PC: nur den rechten Rahmen schließen. Der Connector hört zu.
        if (imRahmen()) {
            try {
                window.parent.postMessage(
                    { tg: "spiel", art: "verlassen", hinweis: "Spiel geschlossen." },
                    location.origin === "null" ? "*" : location.origin
                );

                return;
            } catch (fehler) {
                // Weiter unten mit window.close / Adresse versuchen.
            }
        }

        // Eigener Tab: erst schließen versuchen.
        let zu = false;

        try {
            window.close();
            zu = window.closed;
        } catch (fehler) {
            zu = false;
        }

        if (zu) return;

        // Der Browser lässt das Schließen nicht zu: zurück zum Connector,
        // ohne den Tab dahinter (die Homepage) anzufassen.
        setTimeout(function () {
            try {
                if (!window.closed) location.href = connectorAdresse();
            } catch (fehler) {
                location.href = CONNECTOR;
            }
        }, 60);
    }

    // --------------------------------------------------
    // AUSSEHEN
    // --------------------------------------------------

    function stilEinsetzen() {
        if (document.getElementById("tg-zurueck-stil")) return;

        const stil = document.createElement("style");

        stil.id = "tg-zurueck-stil";
        stil.textContent = [
            // Dezent: kleine Schrift, dunkles Grün, goldener Rand wie in den
            // Spielen selbst. Erst beim Berühren wird er kräftiger.
            ".tg-zurueck{",
            "  font:inherit;font-size:11px;line-height:1.35;letter-spacing:.02em;",
            "  padding:3px 8px;border-radius:5px;cursor:pointer;white-space:nowrap;",
            "  color:#e8cd8e;background:#0b241d;border:1px solid #d9bc7c66;opacity:.85;",
            "  transition:background .15s,color .15s,opacity .15s,border-color .15s;",
            "}",
            ".tg-zurueck:hover,.tg-zurueck:focus-visible{",
            "  background:#184536;color:#f6edda;border-color:#e8cd8e;opacity:1;outline:none;",
            "}",
            // In einer Reihe (Titelleiste, HUD) rutscht der Knopf nach rechts.
            // Red Reactor (Buben-Ansicht): Metallfläche mit roter LED-Kante.
            '[data-stil="buben"] .tg-zurueck,[data-stil="jungs"] .tg-zurueck,',
            '[data-stil="maenner"] .tg-zurueck{',
            "  color:#e6e9ec;background:#16181c;border-color:#e0242f;opacity:.9;",
            "}",
            '[data-stil="buben"] .tg-zurueck:hover,[data-stil="buben"] .tg-zurueck:focus-visible,',
            '[data-stil="jungs"] .tg-zurueck:hover,[data-stil="jungs"] .tg-zurueck:focus-visible,',
            '[data-stil="maenner"] .tg-zurueck:hover,[data-stil="maenner"] .tg-zurueck:focus-visible{',
            "  background:#23262c;color:#ffffff;border-color:#ff5560;",
            "}",
            ".tg-zurueck--rechts{margin-left:auto}",
            // Eigene Zeile in einer Karte (1×1-Rennen, Paare).
            ".tg-zurueck--karte{display:block;margin:0 0 10px auto}",
            // Notnagel, wenn ein Spiel keine passende Leiste hat.
            ".tg-zurueck--ecke{position:fixed;top:8px;right:8px;z-index:2147483000}",
            // Lange Titel bekommen Platz, statt die Leiste zu sprengen.
            ".tg-mit-zurueck .titel{min-width:0;overflow:hidden;text-overflow:ellipsis;",
            "  white-space:nowrap}",
            "@media(max-width:520px){.tg-zurueck{font-size:10px;padding:3px 6px}}"
        ].join("");

        (document.head || document.documentElement).appendChild(stil);
    }

    function istReihe(element) {
        try {
            const anzeige = getComputedStyle(element).display;

            return anzeige.indexOf("flex") !== -1 || anzeige.indexOf("grid") !== -1;
        } catch (fehler) {
            return false;
        }
    }

    function knopfEinsetzen() {
        const knopf = document.createElement("button");

        knopf.type = "button";
        knopf.className = "tg-zurueck";
        knopf.id = "tgZurueck";
        knopf.textContent = "Zurück zum Connector";
        knopf.title = "Spiel schließen und zurück zum Connector";
        knopf.addEventListener("click", zurueck);

        const ziel = daten.ziel ? document.querySelector(daten.ziel) : null;
        const vor = daten.vor ? document.querySelector(daten.vor) : null;
        const leiste = document.querySelector(".titelleiste");

        // 1. Ausdrücklich genannter Behälter.
        if (ziel) {
            if (daten.anfang === "1") {
                ziel.insertBefore(knopf, ziel.firstChild);
                if (daten.auf === "karte") knopf.classList.add("tg-zurueck--karte");
            } else if (vor && vor.parentElement === ziel) {
                ziel.insertBefore(knopf, vor);
                if (istReihe(ziel)) knopf.classList.add("tg-zurueck--rechts");
            } else {
                ziel.appendChild(knopf);
                if (istReihe(ziel)) knopf.classList.add("tg-zurueck--rechts");
            }

            return;
        }

        // 2. Die Fensterleiste der Brett- und Denkspiele.
        if (leiste) {
            leiste.classList.add("tg-mit-zurueck");

            const kreuz = leiste.querySelector(".kreuz");

            if (kreuz) {
                kreuz.parentElement.insertBefore(knopf, kreuz);
            } else {
                leiste.appendChild(knopf);
            }

            knopf.classList.add("tg-zurueck--rechts");

            return;
        }

        // 3. Notnagel: Ecke oben rechts.
        knopf.classList.add("tg-zurueck--ecke");
        document.body.appendChild(knopf);
    }

    // Spiele mit eigenem Zurück-Knopf (Mahjong, Tetris) rufen das hier auf.
    window.tgZurueck = zurueck;

    // data-kein-knopf="1": nur die Funktion bereitstellen, nichts einbauen.
    function start() {
        if (daten.keinKnopf === "1") return;

        stilEinsetzen();
        knopfEinsetzen();
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", start);
    } else {
        start();
    }
})();
