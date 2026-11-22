/* =========================================================================
   TRIPEAKS – 28 Karten in drei Bergen (Reihen 3/6/9/10), eine offene
   Ablagekarte, 23 Karten Vorrat. Freie Karten mit einem Wert ±1 wandern auf
   die Ablage; König und Ass schließen aneinander an.
   ========================================================================= */
(function () {
  "use strict";
  var TG = window.TG, K = TG.Karten;
  var MODI = TG.MODI = TG.MODI || {};

  var z = null;
  var AUSWAHL = null;

  /* Welche Karte der nächsten Reihe bedeckt welche Karte? */
  var UEBER = [
    [[0, 1], [2, 3], [4, 5]],
    [[0, 1], [1, 2], [3, 4], [4, 5], [6, 7], [7, 8]],
    [[0, 1], [1, 2], [2, 3], [3, 4], [4, 5], [5, 6], [6, 7], [7, 8], [8, 9]]
  ];
  var REIHEN = [3, 6, 9, 10];

  /* Linke Kante in Kartenbreiten – die Berge überlappen halb. */
  var X = [
    [1.5, 4.5, 7.5],
    [1, 2, 4, 5, 7, 8],
    [0.5, 1.5, 2.5, 3.5, 4.5, 5.5, 6.5, 7.5, 8.5],
    [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
  ];

  function start(r) {
    var n = 0;
    for (var i = 0; i < r; i++) n += REIHEN[i];
    return n;
  }
  var GESAMT = 28;

  var MOD = MODI.tripeaks = {
    id: "tripeaks",
    name: "TriPeaks",
    kurz: "Drei Berge, 28 Karten. Freie Karten mit einem Wert höher oder tiefer " +
      "auf die Ablage legen – die Farbe zählt nicht.",
    vorschau: '<svg viewBox="0 0 120 84" aria-hidden="true">' +
      '<g style="fill:var(--vv-karte);stroke:var(--vv-rand)" stroke-width="1.2">' +
      '<rect x="18" y="4" width="18" height="14" rx="2"/><rect x="51" y="4" width="18" height="14" rx="2"/>' +
      '<rect x="84" y="4" width="18" height="14" rx="2"/>' +
      '<rect x="12" y="20" width="18" height="14" rx="2"/><rect x="24" y="20" width="18" height="14" rx="2"/>' +
      '<rect x="45" y="20" width="18" height="14" rx="2"/><rect x="57" y="20" width="18" height="14" rx="2"/>' +
      '<rect x="78" y="20" width="18" height="14" rx="2"/><rect x="90" y="20" width="18" height="14" rx="2"/>' +
      '<rect x="8" y="36" width="18" height="14" rx="2"/><rect x="22" y="36" width="18" height="14" rx="2"/>' +
      '<rect x="36" y="36" width="18" height="14" rx="2"/><rect x="50" y="36" width="18" height="14" rx="2"/>' +
      '<rect x="64" y="36" width="18" height="14" rx="2"/><rect x="78" y="36" width="18" height="14" rx="2"/>' +
      '<rect x="92" y="36" width="18" height="14" rx="2"/>' +
      '<rect x="8" y="54" width="18" height="14" rx="2"/><rect x="30" y="54" width="18" height="14" rx="2"/>' +
      '<rect x="52" y="54" width="18" height="14" rx="2"/><rect x="74" y="54" width="18" height="14" rx="2"/>' +
      '<rect x="96" y="54" width="18" height="14" rx="2"/></g>' +
      '<rect x="6" y="72" width="20" height="10" rx="2" style="fill:var(--vv-rot)" opacity=".55"/>' +
      '<text x="20" y="15" style="fill:var(--vv-hell)" font-size="9" font-family="Georgia,serif">A</text>' +
      '<text x="53" y="15" style="fill:var(--vv-rot)" font-size="9" font-family="Georgia,serif">7</text>' +
      '<text x="86" y="15" style="fill:var(--vv-hell)" font-size="9" font-family="Georgia,serif">D</text>' +
      "</svg>",
    optionen: [],
    standard: {},
    regeln: [
      "<b>Ziel:</b> alle 28 Karten der drei Berge auf die Ablage abräumen.",
      "<b>Aufbau:</b> vier Reihen mit 3, 6, 9 und 10 Karten. Nur die unterste Reihe liegt offen; die übrigen Karten deckt die nächste Reihe auf.",
      "<b>Spielen:</b> eine freie Karte darf auf die Ablage, wenn ihr Wert genau eins höher oder tiefer liegt – die Farbe ist egal. Ass und König schließen aneinander an.",
      "<b>Frei:</b> eine Karte ist frei, wenn keine Karte mehr auf ihr liegt.",
      "<b>Nachziehen:</b> unten links eine Karte aufdecken, sie wird die neue Ablagekarte.",
      "<b>Punkte:</b> +10 für jede abgeräumte Karte."
    ],

    neu: function () {
      var d = K.mischen(K.deck([0, 1, 2, 3]));
      z = { t: [], stock: [], waste: [], punkte: 0, zuege: 0, sieg: false };
      for (var i = 0; i < GESAMT; i++) {
        var k = d[i];
        k.offen = i >= start(3);
        z.t.push(k);
      }
      z.waste = [d[28]];
      z.stock = d.slice(29);
      AUSWAHL = null;
      MOD.zeichne();
    },

    snap: function () {
      return JSON.parse(JSON.stringify({
        t: z.t, stock: z.stock, waste: z.waste, punkte: z.punkte, zuege: z.zuege
      }));
    },

    restore: function (s) {
      z = s;
      z.sieg = false;
      AUSWAHL = null;
      MOD.zeichne();
    },

    info: function () {
      var weg = z ? z.t.filter(function (k) { return !k; }).length : 0;
      return {
        punkte: z ? z.punkte : 0, zuege: z ? z.zuege : 0,
        extra: "Berge " + weg + "/28 · Vorrat " + (z ? z.stock.length : 0)
      };
    },

    zeichne: function () {
      if (!z) return;
      var brett = document.getElementById("brett");
      brett.innerHTML = "";
      var wrap = document.createElement("div");
      wrap.className = "tripeaks";

      var feld = document.createElement("div");
      feld.className = "berge";
      var w = brett.clientWidth || 560;
      var kw = Math.min(w / 10.6, 88);
      var kh = kw * 1.4;
      var dy = kh * 0.56;
      feld.style.width = (kw * 10) + "px";
      feld.style.height = (kh + 3 * dy) + "px";
      feld.style.margin = "0 auto";

      for (var r = 0; r < 4; r++) {
        for (var p = 0; p < REIHEN[r]; p++) {
          var i = start(r) + p;
          var k = z.t[i];
          var x = X[r][p] * kw;
          var y = r * dy;
          if (!k) {
            var leer = document.createElement("div");
            leer.className = "berg-platz";
            leer.style.cssText = "left:" + x + "px;top:" + y + "px;width:" + kw + "px;height:" + kh + "px";
            feld.appendChild(leer);
            continue;
          }
          var frei = istFrei(i);
          var c = K.el(k, {
            art: "tp", i: i, index: i,
            klasse: (frei ? "frei" : "bedeckt") +
              (frei && spielbar(k) ? " spielbar" : "") +
              (AUSWAHL === i ? " gewaehlt" : "")
          });
          c.style.cssText = "left:" + x + "px;top:" + y + "px;width:" + kw + "px;height:" + kh + "px";
          if (frei) c.dataset.zone = "karte:" + i;
          feld.appendChild(c);
        }
      }

      var fuss = document.createElement("div");
      fuss.className = "tripeaks-fuss";

      var stock = document.createElement("div");
      stock.className = "stapelfeld" + (z.stock.length ? "" : " leer");
      stock.dataset.zone = "stock";
      stock.setAttribute("aria-label", "Nachziehstapel");
      if (z.stock.length) {
        stock.appendChild(K.el(z.stock[z.stock.length - 1], { hinten: true, klasse: "stapelkarte" }));
        var zahl = document.createElement("span");
        zahl.className = "stapelzahl";
        zahl.textContent = z.stock.length;
        stock.appendChild(zahl);
      } else {
        stock.innerHTML = '<span class="leerzeichen">leer</span>';
      }

      var waste = document.createElement("div");
      waste.className = "stapelfeld waste" + (z.waste.length ? "" : " leer");
      waste.dataset.zone = "waste";
      waste.setAttribute("aria-label", "Ablagekarte");
      if (z.waste.length) {
        var wk = z.waste[z.waste.length - 1];
        waste.appendChild(K.el(wk, { klasse: "ohne-klick" }));
        if (z.waste.length > 1) {
          var rest = document.createElement("span");
          rest.className = "waste-zahl";
          rest.textContent = z.waste.length;
          waste.appendChild(rest);
        }
      } else {
        waste.innerHTML = '<span class="leerzeichen">Ablage</span>';
      }

      fuss.append(stock, waste);
      wrap.append(feld, fuss);
      brett.appendChild(wrap);
    },

    waehlen: function (was) {
      if (!z || !was || was.art !== "tp") return;
      var i = was.i;
      if (!z.t[i]) return;
      if (!istFrei(i)) { TG.Shell.hinweis("Diese Karte ist noch bedeckt.", true); return; }
      if (!spielbar(z.t[i])) {
        AUSWAHL = null;
        TG.Shell.auswahl = null;
        MOD.zeichne();
        TG.Shell.hinweis("Auf " + K.kartenText(ablage()) + " passt nur ein Wert " +
          (ablage().r === 13 ? "König oder 2" : "eins höher oder tiefer") + ".", true);
        return;
      }
      MOD.spielen(i);
    },

    ablegen: function (was, zone) {
      if (!z || !was) return false;
      if (was.art === "tp") {
        /* Nur auf die Ablage ziehen – ein Loslassen auf einem anderen Berg
           soll keine Karte spielen. */
        if (zone !== "waste" && zone !== "karte:" + was.i) return false;
        MOD.waehlen(was);
        return true;
      }
      if (zone === "stock") { MOD.nachziehen(); return true; }
      return false;
    },

    zoneKlick: function (zone) {
      if (!z) return;
      if (zone === "stock") { MOD.nachziehen(); return; }
      AUSWAHL = null;
      TG.Shell.auswahl = null;
      MOD.zeichne();
    },

    spielen: function (i) {
      var k = z.t[i];
      if (!k || !istFrei(i) || !spielbar(k)) return false;
      TG.Shell.merken();
      z.t[i] = null;
      z.waste.push(k);
      k.offen = true;
      z.punkte += 10;
      z.zuege++;
      AUSWAHL = null;
      TG.Shell.auswahl = null;
      aufdecken();
      var vorher = K.positionen();
      MOD.zeichne();
      K.bewege(vorher);
      TG.Shell.zugFertig();
      if (z.t.every(function (c) { return !c; })) {
        z.sieg = true;
        TG.Shell.sieg("Alle drei Berge sind abgeräumt.");
        return true;
      }
      pruefeEnde();
      return true;
    },

    nachziehen: function () {
      if (!z.stock.length) {
        TG.Shell.hinweis("Der Vorrat ist leer.", true);
        pruefeEnde();
        return;
      }
      TG.Shell.merken();
      var k = z.stock.pop();
      k.offen = true;
      z.waste.push(k);
      z.zuege++;
      AUSWAHL = null;
      TG.Shell.auswahl = null;
      var vorher = K.positionen();
      MOD.zeichne();
      K.bewege(vorher);
      TG.Shell.zugFertig("Neue Ablagekarte: " + K.kartenText(k) + ".");
      pruefeEnde();
    },

    tipp: function () {
      if (!z) return "";
      var w = ablage();
      for (var i = 0; i < GESAMT; i++) {
        if (z.t[i] && istFrei(i) && spielbar(z.t[i])) {
          AUSWAHL = i;
          TG.Shell.auswahl = { art: "tp", i: i, index: i };
          MOD.zeichne();
          return "Tipp: " + K.kartenText(z.t[i]) + " passt auf " + K.kartenText(w) + ".";
        }
      }
      if (z.stock.length) return "Tipp: eine neue Karte nachziehen.";
      return "Kein Zug mehr möglich – nimm einen Zug zurück.";
    }
  };

  function ablage() { return z.waste[z.waste.length - 1]; }

  function spielbar(k) {
    var w = ablage();
    if (!w) return true;
    if (Math.abs(k.r - w.r) === 1) return true;
    return (k.r === 1 && w.r === 13) || (k.r === 13 && w.r === 1);
  }

  function reihe(i) {
    var r = 0;
    while (i >= start(r + 1)) r++;
    return { r: r, p: i - start(r) };
  }

  function istFrei(i) {
    if (!z.t[i]) return false;
    var pos = reihe(i);
    if (pos.r === 3) return true;
    var unten = start(pos.r + 1);
    var paare = UEBER[pos.r][pos.p];
    return !z.t[unten + paare[0]] && !z.t[unten + paare[1]];
  }

  function aufdecken() {
    for (var i = 0; i < GESAMT; i++) {
      if (z.t[i] && !z.t[i].offen && istFrei(i)) z.t[i].offen = true;
    }
  }

  function pruefeEnde() {
    if (!z || z.sieg) return;
    if (z.stock.length) return;
    var w = ablage();
    for (var i = 0; i < GESAMT; i++) {
      if (z.t[i] && istFrei(i) && spielbar(z.t[i])) return;
    }
    if (w) TG.Shell.verloren("Der Vorrat ist leer und keine freie Karte passt mehr auf " + K.kartenText(w) + ".");
  }
})();
