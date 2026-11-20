/* =========================================================================
   SOLITÄR – gemeinsames Kartenmodul

   Karten sind schlichte Objekte {id, r (1–13), f (0–3), offen}. Alle fünf
   Modi benutzen dasselbe Kartenbild, dieselben Farben und dieselbe
   Bewegungsanimation.
   ========================================================================= */
(function () {
  "use strict";

  var TG = window.TG = window.TG || {};

  var RANG = { 1: "A", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7",
    8: "8", 9: "9", 10: "10", 11: "B", 12: "D", 13: "K" };

  var FARBEN = [
    { sym: "♠", rot: false, name: "Pik" },
    { sym: "♥", rot: true, name: "Herz" },
    { sym: "♦", rot: true, name: "Karo" },
    { sym: "♣", rot: false, name: "Kreuz" }
  ];

  /* Augenbilder (x/y in Prozent der Kartenfläche) – wie im Spider-Solitär. */
  var PIPS = {
    1: [[.5, .5]],
    2: [[.5, .16], [.5, .84]],
    3: [[.5, .16], [.5, .5], [.5, .84]],
    4: [[.28, .16], [.72, .16], [.28, .84], [.72, .84]],
    5: [[.28, .16], [.72, .16], [.5, .5], [.28, .84], [.72, .84]],
    6: [[.28, .16], [.72, .16], [.28, .5], [.72, .5], [.28, .84], [.72, .84]],
    7: [[.28, .16], [.72, .16], [.5, .35], [.28, .5], [.72, .5], [.28, .84], [.72, .84]],
    8: [[.28, .16], [.72, .16], [.5, .35], [.28, .5], [.72, .5], [.5, .65], [.28, .84], [.72, .84]],
    9: [[.28, .13], [.72, .13], [.28, .38], [.72, .38], [.5, .5], [.28, .62], [.72, .62], [.28, .87], [.72, .87]],
    10: [[.28, .1], [.72, .1], [.5, .24], [.28, .36], [.72, .36], [.28, .64], [.72, .64], [.5, .76], [.28, .9], [.72, .9]]
  };

  var nr = 0;

  function karte(r, f, offen) {
    return { id: ++nr, r: r, f: f, offen: !!offen };
  }

  /* farben: Liste von Farben, jede Farbe ergibt ein volles Ass–König-Blatt.
     Klondike/Pyramid/TriPeaks: [0,1,2,3]; Spider: achtmal eine Farbe usw. */
  function deck(farben) {
    var out = [];
    farben.forEach(function (f) {
      for (var r = 1; r <= 13; r++) out.push(karte(r, f, false));
    });
    return out;
  }

  function mischen(a) {
    for (var i = a.length - 1; i > 0; i--) {
      var j = Math.floor(Math.random() * (i + 1));
      var t = a[i]; a[i] = a[j]; a[j] = t;
    }
    return a;
  }

  function rot(k) { return FARBEN[k.f].rot; }
  function rangText(r) { return RANG[r]; }
  function farbText(f) { return FARBEN[f].name; }
  function kartenText(k) { return RANG[k.r] + " " + FARBEN[k.f].sym; }

  function innen(k) {
    var rang = RANG[k.r], sym = FARBEN[k.f].sym;
    var ecke = "<b>" + rang + "</b><i>" + sym + "</i>";
    var mitte;
    if (k.r <= 10) {
      mitte = '<span class="pips">' + PIPS[k.r].map(function (p) {
        return '<i class="pip' + (p[1] > .5 ? " unten" : "") + '" style="left:' +
          (p[0] * 100).toFixed(1) + "%;top:" + (p[1] * 100).toFixed(1) + '%">' + sym + "</i>";
      }).join("") + "</span>";
    } else {
      mitte = '<span class="bild"><b>' + rang + "</b><i>" + sym + "</i></span>";
    }
    return '<span class="ecke oben">' + ecke + "</span>" + mitte +
      '<span class="ecke unten">' + ecke + "</span>";
  }

  /* Kartenbild. opts: {hinten, klasse, art, i, index, titel} */
  function el(k, opts) {
    opts = opts || {};
    var d = document.createElement("div");
    var hinten = opts.hinten || (k && !k.offen);
    d.className = "karte " + (k && !hinten ? (rot(k) ? "rot" : "schwarz") : "hinten") +
      (opts.klasse ? " " + opts.klasse : "");
    if (k) d.dataset.id = String(k.id);
    if (opts.art) d.dataset.art = opts.art;
    if (opts.i !== undefined) d.dataset.i = String(opts.i);
    if (opts.index !== undefined) d.dataset.index = String(opts.index);
    if (opts.titel !== false && k && !hinten) d.title = kartenText(k);
    if (!hinten && k) d.innerHTML = innen(k);
    if (hinten) d.setAttribute("aria-hidden", "true");
    return d;
  }

  /* ---------------------------------------------------------- Bewegung -- */
  /* Vor dem Neuzeichnen die Lage aller Karten merken, danach bewegen() –
     die Karten gleiten dann von ihrer alten an die neue Stelle. */
  function positionen() {
    var map = new Map();
    document.querySelectorAll(".karte[data-id]").forEach(function (e) {
      var r = e.getBoundingClientRect();
      map.set(e.dataset.id, { x: r.left, y: r.top });
    });
    return map;
  }

  function bewege(vorher) {
    if (!vorher || !vorher.size) return;
    if (window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
    document.querySelectorAll(".karte[data-id]").forEach(function (e) {
      var alt = vorher.get(e.dataset.id);
      if (!alt) return;
      var r = e.getBoundingClientRect();
      var dx = Math.round(alt.x - r.left), dy = Math.round(alt.y - r.top);
      if (Math.abs(dx) < 2 && Math.abs(dy) < 2) return;
      e.style.zIndex = "6";
      var lauf = e.animate(
        [{ transform: "translate(" + dx + "px," + dy + "px)" }, { transform: "none" }],
        { duration: 190, easing: "cubic-bezier(.2,.85,.3,1)" });
      lauf.onfinish = function () { e.style.zIndex = ""; };
    });
  }

  function gruppe(a, b) { return a === b; }

  TG.Karten = {
    FARBEN: FARBEN, PIPS: PIPS,
    karte: karte, deck: deck, mischen: mischen,
    rot: rot, rangText: rangText, farbText: farbText, kartenText: kartenText,
    innen: innen, el: el, positionen: positionen, bewege: bewege, gruppe: gruppe
  };
})();
