/* TRASHYGECKO – gemeinsame Spiel-Logik für Startbildschirm, Vollbild und Feldmaße
 *
 * Kleine, abhängigkeitsfreie Hilfe für Spiele. Jedes Spiel bleibt eigenständig;
 * hier steht nur, was in jedem Spiel gleich funktionieren muss:
 *
 *   TGSpiel.vollbild.ist()                 – läuft gerade Vollbild?
 *   TGSpiel.vollbild.um()                  – ein-/ausschalten
 *   TGSpiel.vollbild.an() / .aus()         – gezielt
 *   TGSpiel.vollbild.knopf(el, texte)      – Knopf anschließen (Text + Zustand)
 *   TGSpiel.vollbild.beiWechsel(fn)        – fn(an) bei jedem Wechsel
 *   TGSpiel.feldPassen(canvas, behaelter, {rand, verhaeltnis})
 *   TGSpiel.zeige(el, an)                  – hidden setzen/entfernen
 *
 * Vollbild: echtes Fullscreen-API, wo es das gibt. Wo nicht (iPhone-Safari),
 * schaltet eine Pseudo-Vollbild-Klasse die Seite randlos – mit demselben
 * Knopf und mit Escape wieder aus. Nach dem Aussteigen wird die Darstellung
 * über ein Wechsel-Ereignis neu berechnet.
 */
(function () {
  "use strict";

  var wurzel = document.documentElement;
  var pseudoKlasse = "ist-vollbild";
  var hoerer = [];
  var letzterZustand = false;

  function echtAktiv() {
    return Boolean(document.fullscreenElement || document.webkitFullscreenElement || document.mozFullScreenElement);
  }

  function ist() {
    return echtAktiv() || document.body.classList.contains(pseudoKlasse);
  }

  function melden() {
    var an = ist();

    if (an === letzterZustand) return;
    letzterZustand = an;

    if (document.body) document.body.classList.toggle("vollbild-an", an);
    hoerer.forEach(function (fn) {
      try { fn(an); } catch (fehler) { /* ein Hörer darf die anderen nicht stoppen */ }
    });
  }

  function pseudoAn() {
    document.body.classList.add(pseudoKlasse);
    document.body.style.overflow = "hidden";
    melden();
  }

  function pseudoAus() {
    document.body.classList.remove(pseudoKlasse);
    document.body.style.overflow = "";
    melden();
  }

  function an() {
    if (ist()) return;

    var ziel = document.getElementById("app") || wurzel;
    var anfrage = ziel.requestFullscreen || ziel.webkitRequestFullscreen || ziel.mozRequestFullScreen;

    if (anfrage) {
      try {
        var ergebnis = anfrage.call(ziel, { navigationUI: "hide" });

        if (ergebnis && typeof ergebnis.then === "function") {
          ergebnis.then(melden, pseudoAn);
        } else {
          setTimeout(melden, 60);
        }
        return;
      } catch (fehler) { /* weiter unten als Pseudo-Vollbild */ }
    }

    pseudoAn();
  }

  function aus() {
    if (document.body.classList.contains(pseudoKlasse)) pseudoAus();

    var beenden = document.exitFullscreen || document.webkitExitFullscreen || document.mozCancelFullScreen;

    if ((echtAktiv()) && beenden) {
      try {
        var ergebnis = beenden.call(document);

        if (ergebnis && typeof ergebnis.then === "function") ergebnis.then(melden, melden);
        else setTimeout(melden, 60);
      } catch (fehler) {
        melden();
      }
    }
    setTimeout(melden, 80);
  }

  function um() {
    if (ist()) aus();
    else an();
  }

  function beiWechsel(fn) {
    if (typeof fn === "function") hoerer.push(fn);
  }

  /* Knopf mit Zustand: Beschriftung, aria-pressed, Symbol und Titel. */
  function knopf(el, texte) {
    if (!el) return;
    var t = texte || {};

    function beschriften() {
      var an = ist();
      var text = el.querySelector(".vb-text");

      el.setAttribute("aria-pressed", an ? "true" : "false");
      el.title = an ? (t.titelAn || "Vollbild beenden (Escape)") : (t.titelAus || "Vollbild einschalten");
      el.setAttribute("aria-label", an ? (t.textAn || "Vollbild beenden") : (t.textAus || "Vollbild"));
      if (text) text.textContent = an ? (t.textAn || "Vollbild beenden") : (t.textAus || "Vollbild");

      var symbol = el.querySelector(".vb-symbol");
      if (symbol) symbol.textContent = an ? "⤡" : "⛶";
    }

    el.addEventListener("click", function (e) {
      e.preventDefault();
      um();
      /* Nach dem Klick den Fokus nicht im Knopf lassen: Leertaste soll
         wieder dem Spiel gehören. */
      el.blur();
    });

    beiWechsel(beschriften);
    beschriften();
  }

  ["fullscreenchange", "webkitfullscreenchange", "mozfullscreenchange"].forEach(function (name) {
    document.addEventListener(name, melden);
  });

  document.addEventListener("keydown", function (e) {
    if (e.key !== "Escape") return;
    if (!ist()) return;
    e.preventDefault();
    aus();
  });

  /* ---------------------------------------------------------------- Feldmaße */

  /* Setzt die CSS-Größe einer Leinwand so, dass sie ohne Verzerrung in den
     Behälter passt, und berechnet bei jeder Änderung neu (Fenstergröße,
     Gerätedrehung, Vollbild, weiche Tastatur). Die Zeichenfläche selbst
     bleibt in ihren logischen Maßen – nur die Darstellung wächst. */
  function feldPassen(canvas, behaelter, optionen) {
    if (!canvas || !behaelter) return null;

    var o = optionen || {};
    var verhaeltnis = o.verhaeltnis > 0 ? o.verhaeltnis : 1;
    var rand = o.rand >= 0 ? o.rand : 0;
    var letzte = "";
    var geplant = false;
    var meldenFn = o.beiGroesse;

    function messen() {
      geplant = false;

      var box = behaelter.getBoundingClientRect();
      var platzB = Math.max(64, box.width - rand);
      var platzH = Math.max(64, box.height - rand);
      var breite = Math.min(platzB, platzH * verhaeltnis);
      var hoehe = breite / verhaeltnis;
      var schluessel = Math.round(breite) + "x" + Math.round(hoehe);

      if (schluessel === letzte) return;
      letzte = schluessel;

      canvas.style.width = Math.round(breite) + "px";
      canvas.style.height = Math.round(hoehe) + "px";

      if (typeof meldenFn === "function") meldenFn(Math.round(breite), Math.round(hoehe));
    }

    function anstossen() {
      if (geplant) return;
      geplant = true;
      requestAnimationFrame(messen);
    }

    if (window.ResizeObserver) new ResizeObserver(anstossen).observe(behaelter);
    window.addEventListener("resize", anstossen);
    window.addEventListener("orientationchange", anstossen);
    document.addEventListener("fullscreenchange", anstossen);
    document.addEventListener("webkitfullscreenchange", anstossen);
    if (window.visualViewport) {
      window.visualViewport.addEventListener("resize", anstossen);
      window.visualViewport.addEventListener("scroll", anstossen);
    }
    if (document.fonts && document.fonts.ready) document.fonts.ready.then(anstossen);

    anstossen();

    return { messen: messen, anstossen: anstossen };
  }

  /* ---------------------------------------------------------------- Kleinkram */

  function zeige(el, an) {
    if (!el) return;
    if (an) el.removeAttribute("hidden");
    else el.setAttribute("hidden", "");
  }

  window.TGSpiel = {
    vollbild: { ist: ist, an: an, aus: aus, um: um, knopf: knopf, beiWechsel: beiWechsel },
    feldPassen: feldPassen,
    zeige: zeige
  };
})();
