/* ===========================================================================
   Ergebnis-Sprüche für alle Spiele
   ---------------------------------------------------------------------------
   Lob, Gewinn und Pech stehen ausschließlich in den zentralen Listen des
   Connectors:

     spiele/lob.js       (connector/lob.txt)        -> LOB_SPRUECHE, lobSpruch()
     spiele/gewonnen.js  (connector/gewonnen.txt)   -> GEWONNEN_SPRUECHE, gewonnenSpruch()
     spiele/pech.js      (connector/pech.txt)       -> PECH_SPRUECHE, pechSpruch()

   Hier steht nur die Auswahl und die Anzeige – kein einziger Spruch. Die drei
   Dateien werden beim ersten Aufruf nachgeladen (zusammen rund 330 kB), damit
   eine Seite, die nie ein Ergebnis meldet, sie nicht mitlädt.

   Benutzung in einem Spiel:

     <script src="../ergebnis.js"></script>      (Pfad je nach Tiefe)
     TGErgebnis.gewonnen();   // Sieg: Gewinn- oder Lobspruch
     TGErgebnis.verloren();   // Niederlage: Pechspruch
     TGErgebnis.lob();        // nur Anerkennung, ohne Sieg

   Die Anzeige ist eine kleine Karte am unteren Rand, die nach ein paar
   Sekunden von selbst verschwindet. Sie nimmt keine Klicks an
   (pointer-events:none), verdeckt also nie eine Bedienung, und wiederholt
   keinen Spruch direkt hintereinander (die Mischbeutel in den Listen sorgen
   dafür, hier zusätzlich geprüft).
   ========================================================================= */
(function () {
  "use strict";

  var WURZEL = (function () {
    var s = document.currentScript;
    if (!s || !s.src) return "";
    return s.src.slice(0, s.src.lastIndexOf("/") + 1);
  })();

  var DATEIEN = ["lob.js", "gewonnen.js", "pech.js"];
  var ladeversprechen = null;
  var feld = null;
  var timer = null;
  var letzterText = "";
  var gesperrtBis = 0;

  function lade() {
    if (ladeversprechen) return ladeversprechen;
    ladeversprechen = Promise.all(DATEIEN.map(function (name) {
      return new Promise(function (fertig) {
        var s = document.createElement("script");
        s.src = WURZEL + name;
        s.async = false;          // Reihenfolge bleibt: lob, gewonnen, pech
        s.onload = function () { fertig(true); };
        s.onerror = function () { fertig(false); };
        (document.head || document.documentElement).appendChild(s);
      });
    })).then(function (ergebnis) {
      return ergebnis.every(Boolean);
    });
    return ladeversprechen;
  }

  /* Zieht einen Spruch aus der passenden zentralen Liste. Die Funktionen dort
     mischen einen Beutel und legen ihn erst neu an, wenn er leer ist – damit
     kommt keiner zweimal direkt hintereinander. */
  function ziehe(art) {
    try {
      if (art === "pech" && typeof window.pechSpruch === "function") {
        return window.pechSpruch();
      }
      if (art === "gewonnen" && typeof window.gewonnenSpruch === "function") {
        return window.gewonnenSpruch();
      }
      if (typeof window.lobSpruch === "function") {
        return window.lobSpruch(false);
      }
    } catch (e) {
      /* Liste fehlt oder ist kaputt – dann bleibt die Anzeige leer. */
    }
    return "";
  }

  function karte() {
    if (feld && feld.isConnected) return feld;
    feld = document.createElement("div");
    feld.id = "tg-spruch";
    feld.setAttribute("role", "status");
    feld.setAttribute("aria-live", "polite");
    feld.style.cssText = [
      "position:fixed",
      "left:50%",
      "bottom:calc(10px + env(safe-area-inset-bottom,0px))",
      "transform:translate(-50%,10px)",
      "z-index:9998",
      "pointer-events:none",
      "max-width:min(520px,92vw)",
      "box-sizing:border-box",
      "padding:11px 16px",
      "border-radius:12px",
      "background:var(--tg-spruch-bg,rgba(10,26,20,.9))",
      "color:var(--tg-spruch-text,#f6eeda)",
      "border:1px solid var(--tg-spruch-rand,rgba(232,205,142,.6))",
      "box-shadow:0 10px 26px rgba(0,0,0,.38)",
      "font:600 15px/1.35 system-ui,-apple-system,'Segoe UI',sans-serif",
      "text-align:center",
      "opacity:0",
      "transition:opacity .22s ease,transform .22s ease",
      "backdrop-filter:blur(3px)",
      "-webkit-backdrop-filter:blur(3px)"
    ].join(";");
    (document.body || document.documentElement).appendChild(feld);
    return feld;
  }

  function zeige(text) {
    if (!text) return;
    var d = karte();
    d.textContent = text;
    d.style.opacity = "1";
    d.style.transform = "translate(-50%,0)";
    clearTimeout(timer);
    /* Kurze Sprüche 3,5 s, lange höchstens 7 s – der Spielfluss läuft weiter,
       weil die Karte keine Klicks annimmt. */
    var dauer = Math.min(7000, Math.max(3500, 2200 + text.length * 45));
    timer = setTimeout(function () {
      d.style.opacity = "0";
      d.style.transform = "translate(-50%,10px)";
    }, dauer);
  }

  /* art: "gewonnen" | "pech" | "lob". Ein Aufruf je Ergebnis: innerhalb von
     1,2 s wird ein zweiter Aufruf geschluckt, damit Spiele, die ihr Ende
     mehrfach zeichnen (Neuzeichnen, Serverstand), nicht doppelt melden. */
  function melde(art) {
    var jetzt = Date.now();
    if (jetzt < gesperrtBis) return false;
    gesperrtBis = jetzt + 1200;
    lade().then(function (ok) {
      if (!ok) return;
      var text = ziehe(art);
      if (text && text === letzterText) text = ziehe(art);
      if (!text) return;
      letzterText = text;
      zeige(text);
    });
    return true;
  }

  window.TGErgebnis = {
    gewonnen: function () { return melde("gewonnen"); },
    /* Für Spiele, die nur den Namen des Siegers kennen: Ist der eigene Name
       dabei, gibt es den Gewinnspruch, sonst den Pechspruch. Ohne eigenen
       Namen (Ein-Gerät-Partie) gilt der Sieger. */
    nachName: function (sieger, meiner) {
      if (meiner && sieger && String(meiner) !== String(sieger)) return melde("pech");
      return melde("gewonnen");
    },
    verloren: function () { return melde("pech"); },
    lob: function () { return melde("lob"); },
    /* Für Prüfungen und Spiele, die den Text selbst setzen wollen. */
    bereit: lade,
    ziehe: ziehe,
    zeige: zeige
  };
})();
