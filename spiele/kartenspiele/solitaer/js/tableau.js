/* =========================================================================
   SOLITÄR – Spaltenaufbau für Klondike, Spider und FreeCell.
   Karten liegen absolut übereinander; offene Karten haben einen größeren
   Abstand als verdeckte, damit man die Werte lesen kann.
   ========================================================================= */
(function () {
  "use strict";
  var TG = window.TG = window.TG || {};
  var K = TG.Karten;

  function breite(el) {
    var w = el.clientWidth;
    if (!w) {
      var r = el.getBoundingClientRect();
      w = r.width;
    }
    return w || 72;
  }

  /* el: position:relative; karten: Array; o: {art,i,offenAb,auswahl,abstandOffen,
     abstandZu,karteKlasse} */
  function fuellen(el, karten, o) {
    o = o || {};
    var w = breite(el);
    var h = w * 1.4;
    var offen = o.abstandOffen !== undefined ? o.abstandOffen : 0.30;
    var zu = o.abstandZu !== undefined ? o.abstandZu : 0.14;
    el.innerHTML = "";
    el.classList.toggle("leer", !karten.length);
    if (!karten.length) {
      var leer = document.createElement("div");
      leer.className = "leerfeld";
      leer.textContent = o.leerText || "";
      el.appendChild(leer);
      el.style.height = h + "px";
      return;
    }
    var y = 0;
    karten.forEach(function (k, i) {
      var c = K.el(k, {
        art: o.art,
        i: o.i,
        index: i,
        klasse: (o.auswahl !== undefined && i >= o.auswahl ? "gewaehlt " : "") +
          (o.karteKlasse || "")
      });
      c.style.position = "absolute";
      c.style.left = "0";
      c.style.right = "0";
      c.style.top = y + "px";
      if (o.tippIndex === i) c.classList.add("tipp");
      el.appendChild(c);
      y += (k.offen ? offen : zu) * w;
    });
    el.style.height = (y - (karten[karten.length - 1].offen ? offen : zu) * w + h) + "px";
  }

  TG.Spalte = { fuellen: fuellen, breite: breite };

  /* Nach Größenänderung neu zeichnen, damit die Abstände stimmen. */
  var zeit = null;
  window.addEventListener("resize", function () {
    if (!TG.Shell || !TG.Shell.spiel) return;
    clearTimeout(zeit);
    zeit = setTimeout(function () {
      if (TG.Shell.spiel) TG.Shell.spiel.zeichne();
    }, 150);
  });
})();
