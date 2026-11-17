// Groessen je Bildschirm: Knoepfe, Schrift und Bedienleiste wachsen mit dem
// Geraet.  Die Masse der Bedien-Oberflaeche stehen in css/mahjong.css in rem,
// hier wird nur die Grundschrift gesetzt.
//
// Bezug ist die Diagonale in CSS-Pixeln: 854x480 (5 Zoll) = 1,00, damit
// 6 Zoll (960x540) = 1,09, 7 Zoll (1024x600) = 1,22 und 10 Zoll (1280x800) =
// 1,55.  Hochkant aendert sich nichts, weil die Diagonale gleich bleibt.
(function () {
    let BEZUG = Math.sqrt(854 * 854 + 480 * 480);
    let GRUND = 16;                 // Grundschrift in Pixeln bei Faktor 1
    let MIN = 0.85;
    let MAX = 1.8;
    let zuletzt = 0;

    // Liefert den Faktor, den auch renderer.js fuer die Zoom-Obergrenze nutzt.
    function groesseFaktor() {
        let w = window.innerWidth || document.documentElement.clientWidth || 854;
        let h = window.innerHeight || document.documentElement.clientHeight || 480;
        let f = Math.sqrt(w * w + h * h) / BEZUG;
        return Math.max(MIN, Math.min(MAX, f || 1));
    }

    function grundschriftSetzen() {
        let px = Math.round(GRUND * groesseFaktor() * 100) / 100;
        if (px === zuletzt) return;          // nichts tun, wenn sich nichts aendert
        zuletzt = px;
        document.documentElement.style.fontSize = px + "px";
    }

    window.groesseFaktor = groesseFaktor;
    grundschriftSetzen();
    // Drehen loest ebenfalls resize aus; ein zweiter Listener waere doppelt.
    window.addEventListener("resize", grundschriftSetzen);
})();

/* --- Kopfzeile auf kleinen Geraeten ---------------------------------------
   Reicht der Platz nicht, rutscht die Knopfreihe unter die Zahlen.  Dann
   bekommt die Zahlenzeile drei Zonen: Steine und Paare links, Uhrzeit in der
   Mitte, Punkte rechts an der Uhrzeit.  Das CSS haengt an der Klasse
   "gestapelt" - so gilt es genau dann, wenn wirklich gestapelt wird, egal ob
   das an der Breite, an langen Zahlen (Punkte: 100.000) oder am Drehen liegt. */
(function () {
    function stapelPruefen() {
        let stats = document.querySelector(".stats");
        let knoepfe = document.querySelector(".game-buttons");

        if (!stats || !knoepfe) return;

        // 4 px Toleranz: dieselbe Zeile heisst (fast) dieselbe Oberkante.
        let gestapelt =
            knoepfe.getBoundingClientRect().top >
            stats.getBoundingClientRect().top + 4;

        stats.classList.toggle("gestapelt", gestapelt);
    }

    function start() {
        stapelPruefen();
        window.addEventListener("resize", stapelPruefen);
        // Auch beim Aufbau des Bretts koennen andere Zahlen stehen.
        window.addEventListener("load", stapelPruefen);
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", start);
    } else {
        start();
    }
})();
