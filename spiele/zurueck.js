/*
 * TRASHYGECKO – Rückweg für Spiele (zentrale Navigation)
 *
 * Eine Datei für alle Spiele: sie setzt einen kleinen Knopf in die Seite und
 * führt zu der Stelle zurück, von der das Spiel gestartet wurde.
 *
 *   aus der Spieleübersicht  →  zurück zur Übersicht (lokal oder veröffentlicht)
 *   aus dem Connector        →  zurück zum Connector (derselbe, mit Raum)
 *   im Rahmen des Connectors →  der Connector leert nur den Rahmen
 *   direkt geöffnet          →  zurück zur Spieleübersicht
 *
 * Die Herkunft kommt aus ?zurueck=…, sonst aus dem Referrer, sonst aus dem
 * Rahmen. Jedes Ziel wird geprüft: nur eigene Adresse, die bekannten
 * Projektadressen (Render, GitHub Pages) und Pfade innerhalb von /spiele.
 * Fremde URLs werden nie übernommen.
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
 *   data-kein-knopf="1"        nur die Funktionen bereitstellen
 *
 * Eigene Zurück-Links bekommen data-tg-zurueck; ihr href wird auf das
 * berechnete Ziel gesetzt und der Klick läuft über dieselbe Logik.
 *
 * Für Spiele mit eigenem Zurück-Knopf: window.tgZurueck() aufrufen.
 * Für eigene Verweise: window.TGNAV.zurueckZiel() / TGNAV.connectorZiel().
 */
(function () {
    "use strict";

    // In der Android-App (WebView mit Bruecke TMahjong) gibt es weder Connector
    // noch Rahmen: die App hat ihren eigenen Knopf "SPIEL VERLASSEN".  Hier also
    // gar nichts tun - kein Knopf, kein window.tgZurueck.
    if (window.TMahjong) return;

    const skript = document.currentScript;
    const daten = skript ? skript.dataset : {};
    const ONLINE = "https://trashygecko.onrender.com";

    // Nie zwei Knöpfe in einer Seite.
    if (document.querySelector(".tg-zurueck")) return;

    // In einer laufenden Mehrspieler-Partie gehört das Verlassen dem Spiel.
    if (daten.nur === "ohne-room") {
        try {
            if (new URLSearchParams(location.search).has("room")) return;
        } catch (fehler) {
            return;
        }
    }

    // --------------------------------------------------
    // WOHER KAM DAS SPIEL, WOHIN GEHT ES ZURÜCK?
    // --------------------------------------------------

    // Die Spiele liegen unter …/spiele/…; die Übersicht ist …/spiele/spiele.html.
    function spieleWurzel() {
        const pfad = location.pathname;
        const i = pfad.toLowerCase().indexOf("/spiele/");

        if (i >= 0) return location.origin + pfad.slice(0, i + "/spiele/".length);
        if (/\/spiele\/?$/.test(pfad)) return location.origin + pfad.replace(/\/?$/, "/");

        // file:// oder ein einzelner Ordner: von der Seite aus eine Ebene hoch.
        try {
            return new URL("./", new URL(".", location.href)).href;
        } catch (fehler) {
            return location.href.replace(/[^/]*$/, "");
        }
    }

    function uebersichtZiel() {
        try {
            return new URL("spiele.html", spieleWurzel()).href;
        } catch (fehler) {
            return null;
        }
    }

    // Der Connector braucht einen Server. Auf GitHub Pages liegt nur die
    // statische Seite, dort gilt die laufende Adresse; im APK (file://) ebenso.
    function connectorZiel() {
        if (location.protocol === "file:") return ONLINE + "/connector";
        if (/\.github\.io$/.test(location.hostname)) return ONLINE + "/connector";
        if (location.protocol !== "http:" && location.protocol !== "https:") return ONLINE + "/connector";

        try {
            return new URL("connector/connector.html", spieleWurzel()).href;
        } catch (fehler) {
            return ONLINE + "/connector";
        }
    }

    // Nur eigene und bekannte Projektziele, nur innerhalb der Spiele/Connector.
    function geprueft(ziel) {
        if (!ziel) return null;

        let u;

        try {
            u = new URL(String(ziel), location.href);
        } catch (fehler) {
            return null;
        }

        if (u.protocol === "file:") {
            if (location.protocol !== "file:") return null;

            return /\/spiele\//i.test(u.pathname) || /\/spiele$/i.test(u.pathname) ? u.href : null;
        }

        if (u.protocol !== "http:" && u.protocol !== "https:") return null;
        // Kein Sprung in einen fremden Host, auch nicht mit gültigem Pfad.
        const eigen = u.origin === location.origin;
        const render = u.hostname === "trashygecko.onrender.com";
        const pages = u.hostname === "trashygecko-collab.github.io" && /^\/Trashygecko\//i.test(u.pathname);

        if (!eigen && !render && !pages) return null;
        // Und nur in Bereiche, die zum Projekt gehören (kein /admin usw.).
        if (!/\/spiele(\/|$)/i.test(u.pathname) && !/\/connector(\/|\.html|$)/i.test(u.pathname)) return null;

        return u.href;
    }

    function herkunft() {
        // 1. Ausdrücklich übergeben (?zurueck=…).
        let ausAdresse = null;

        try {
            ausAdresse = geprueft(new URLSearchParams(location.search).get("zurueck"));
        } catch (fehler) {
            ausAdresse = null;
        }

        if (ausAdresse) return ausAdresse;

        // 2. Der Referrer sagt, welche Seite das Spiel geöffnet hat.
        const referrer = geprueft(document.referrer);

        if (referrer) return referrer;

        // 3. Im Rahmen des Connectors: dort bleibt der Rahmen, kein Seitenwechsel.
        if (imRahmen()) return "rahmen";

        // 4. Direkt geöffnet: zurück zur Spieleübersicht.
        return uebersichtZiel();
    }

    function istConnector(ziel) {
        if (ziel === "rahmen") return true;

        try {
            return /\/connector(?:[/?#.]|$)/i.test(new URL(ziel, location.href).pathname);
        } catch (fehler) {
            return false;
        }
    }

    function imRahmen() {
        try {
            return Boolean(window.parent) && window.parent !== window;
        } catch (fehler) {
            // Fremder Rahmen: dann lieber wie ein eigener Tab behandeln.
            return false;
        }
    }

    // --------------------------------------------------
    // ZURÜCK
    // --------------------------------------------------

    function zurueck() {
        const ziel = herkunft();

        // PC/Handy im Connector: nur den rechten Rahmen schließen. Der
        // Connector hört zu und leert ihn; Chat und Liste bleiben stehen.
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

        // Der Browser lässt das Schließen nicht zu: zum tatsächlichen
        // Herkunftsziel. replace statt href – der Verlauf bleibt sauber.
        const adresse = ziel === "rahmen" ? connectorZiel() : ziel;

        setTimeout(function () {
            try {
                if (!window.closed && adresse) location.replace(adresse);
            } catch (fehler) {
                // Dann bleibt nur der Connector als letzter Ausweg.
                location.replace(connectorZiel());
            }
        }, 60);
    }

    // Eigene Zurück-Links ([data-tg-zurueck]) versorgen: href richtig setzen,
    // Klick über dieselbe Logik (im Rahmen also ohne Seitenwechsel).
    function verweiseVersorgen() {
        const ziel = herkunft();
        const adresse = ziel === "rahmen" ? connectorZiel() : ziel;
        const liste = document.querySelectorAll("[data-tg-zurueck]");

        for (let i = 0; i < liste.length; i++) {
            const a = liste[i];

            if (adresse) a.setAttribute("href", adresse);
            if (a.dataset.tgText === "connector" || !a.dataset.tgText) {
                if (a.dataset.tgText === "connector") a.textContent = istConnector(ziel) ? "Zurück zum Connector" : "Zurück zur Spieleübersicht";
            }
            if (!a.__tgZurueck) {
                a.__tgZurueck = true;
                a.addEventListener("click", function (ereignis) {
                    // Neuer Tab / eigenes Fenster: normaler Link.
                    if (ereignis.metaKey || ereignis.ctrlKey || ereignis.shiftKey || a.target === "_blank") return;

                    ereignis.preventDefault();
                    zurueck();
                });
            }
        }
    }

    window.tgZurueck = zurueck;
    window.TGNAV = {
        zurueckZiel: function () {
            const ziel = herkunft();

            return ziel === "rahmen" ? connectorZiel() : ziel;
        },
        connectorZiel: connectorZiel,
        uebersichtZiel: uebersichtZiel,
        herkunft: function () {
            const ziel = herkunft();

            return ziel === "rahmen" ? "connector" : istConnector(ziel) ? "connector" : "uebersicht";
        },
        geprueft: geprueft
    };

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
            // Mädchen-Ansicht: das helle Bonbonrosa der Spielseiten, nicht das
            // dunkle Grün des Klassikers (fehlte bisher ganz).
            '[data-stil="maedchen"] .tg-zurueck{',
            "  color:#5b0f2e;background:#fff6fa;border-color:#ef7aa8;opacity:.95;",
            "}",
            '[data-stil="maedchen"] .tg-zurueck:hover,[data-stil="maedchen"] .tg-zurueck:focus-visible{',
            "  background:#ffe3ef;color:#3a0f24;border-color:#b06bff;",
            "}",
            ".tg-zurueck--rechts{margin-left:auto}",
            // Eigene Zeile in einer Karte (1×1-Rennen, Paare).
            ".tg-zurueck--karte{display:block;margin:0 0 10px auto}",
            // Notnagel, wenn ein Spiel keine passende Leiste hat.
            ".tg-zurueck--ecke{position:fixed;top:8px;right:8px;z-index:2147483000}",
            // Lange Titel bekommen Platz, statt die Leiste zu sprengen.
            ".tg-mit-zurueck .titel{min-width:0;overflow:hidden;text-overflow:ellipsis;",
            "  white-space:nowrap}",
            // Am Handy wird der Knopf mit dem Finger getroffen: mindestens
            // 30 px hoch statt der 23 px der dezenten PC-Fassung.
            "@media(max-width:520px){.tg-zurueck{font-size:11px;padding:4px 9px;min-height:30px;box-sizing:border-box}}"
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
        const ziel = herkunft();
        const zumConnector = istConnector(ziel);

        knopf.type = "button";
        knopf.className = "tg-zurueck";
        knopf.id = "tgZurueck";
        knopf.textContent = zumConnector ? "Zurück zum Connector" : "Zurück zur Spieleübersicht";
        knopf.title = zumConnector
            ? "Spiel schließen und zurück zum Connector"
            : "Spiel schließen und zurück zur Spieleübersicht";
        knopf.addEventListener("click", zurueck);

        const zielBehaelter = daten.ziel ? document.querySelector(daten.ziel) : null;
        const vor = daten.vor ? document.querySelector(daten.vor) : null;
        const leiste = document.querySelector(".titelleiste");

        // 1. Ausdrücklich genannter Behälter.
        if (zielBehaelter) {
            if (daten.anfang === "1") {
                zielBehaelter.insertBefore(knopf, zielBehaelter.firstChild);
                if (daten.auf === "karte") knopf.classList.add("tg-zurueck--karte");
            } else if (vor && vor.parentElement === zielBehaelter) {
                zielBehaelter.insertBefore(knopf, vor);
                if (istReihe(zielBehaelter)) knopf.classList.add("tg-zurueck--rechts");
            } else {
                zielBehaelter.appendChild(knopf);
                if (istReihe(zielBehaelter)) knopf.classList.add("tg-zurueck--rechts");
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

        // 2b. Keine Fensterleiste mehr (Titelzeile entfernt): die oberste
        // Bedienzeile des Spiels nehmen, damit der Knopf nicht über dem Feld
        // schwebt. Rechtsbündig, wenn die Zeile eine Reihe ist.
        const reihe = document.querySelector(".menueleiste") || document.querySelector(".kopf");

        if (reihe) {
            reihe.classList.add("tg-mit-zurueck");
            reihe.appendChild(knopf);

            if (istReihe(reihe)) knopf.classList.add("tg-zurueck--rechts");

            return;
        }

        // 3. Notnagel: Ecke oben rechts.
        knopf.classList.add("tg-zurueck--ecke");
        document.body.appendChild(knopf);
    }

    // Spiele mit eigenem Zurück-Knopf (Mahjong, Tetris) rufen das hier auf.
    window.tgZurueck = zurueck;

    // data-kein-knopf="1": nur die Funktionen bereitstellen, nichts einbauen.
    function start() {
        verweiseVersorgen();

        if (daten.keinKnopf === "1") return;

        stilEinsetzen();
        knopfEinsetzen();
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", start);
    } else {
        start();
    }

    // Später eingebaute Verweise (Spiele bauen ihre Oberfläche selbst).
    document.addEventListener("DOMContentLoaded", verweiseVersorgen);
})();
