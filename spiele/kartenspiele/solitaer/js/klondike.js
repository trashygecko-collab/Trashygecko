/* =========================================================================
   KLONDIKE – das klassische Solitär
   7 Tableau-Spalten, 4 Ablagen Ass bis König, Nachziehen 1 oder 3 Karten,
   leere Spalten nur mit Königen.
   ========================================================================= */
(function () {
  "use strict";
  var TG = window.TG, K = TG.Karten, Sp = TG.Spalte;
  var MODI = TG.MODI = TG.MODI || {};

  var z = null;
  var ziele = [];       // gültige Zielzonen zur aktuellen Auswahl

  var MOD = MODI.klondike = {
    id: "klondike",
    name: "Klondike",
    kurz: "Das klassische Solitär: sieben Spalten, rot und schwarz absteigend, " +
      "vier Ablagen von Ass bis König. Nachziehen mit einer oder drei Karten.",
    vorschau: '<svg viewBox="0 0 120 84" aria-hidden="true">' +
      '<g fill="#2a3138" stroke="#4b5661" stroke-width="1.2">' +
      '<rect x="4" y="4" width="30" height="42" rx="4"/><rect x="40" y="4" width="30" height="42" rx="4"/>' +
      '<rect x="76" y="4" width="30" height="42" rx="4"/>' +
      '<rect x="4" y="54" width="20" height="26" rx="3"/><rect x="28" y="54" width="20" height="26" rx="3"/>' +
      '<rect x="52" y="54" width="20" height="26" rx="3"/><rect x="76" y="54" width="20" height="26" rx="3"/>' +
      '<rect x="100" y="54" width="16" height="22" rx="3"/></g>' +
      '<text x="12" y="24" fill="#d9d3c6" font-size="12" font-family="Georgia,serif">A</text>' +
      '<text x="84" y="24" fill="#c9705f" font-size="12" font-family="Georgia,serif">K</text>' +
      '<text x="10" y="70" fill="#d9d3c6" font-size="10" font-family="Georgia,serif">♠</text>' +
      '<text x="34" y="70" fill="#c9705f" font-size="10" font-family="Georgia,serif">♥</text>' +
      "</svg>",
    optionen: [{ key: "ziehen", label: "Nachziehen", werte: [
      { wert: 1, text: "eine Karte" }, { wert: 3, text: "drei Karten" }
    ] }],
    standard: { ziehen: 1 },
    regeln: [
      "<b>Ziel:</b> alle 52 Karten auf die vier Ablagen bringen – jede Farbe von Ass bis König.",
      "<b>Aufbau:</b> sieben Spalten mit 1 bis 7 Karten, nur die oberste liegt offen.",
      "<b>Tableau:</b> absteigend und abwechselnd rot und schwarz anlegen. Eine leere Spalte darf nur mit einem König beginnen.",
      "<b>Nachziehen:</b> oben links eine Karte (oder drei) aufdecken. Ist der Stapel leer, wird die Ablage zurückgedreht.",
      "<b>Ablage:</b> Asse beginnen die vier Stapel, danach aufsteigend in derselben Farbe.",
      "<b>Punkte:</b> +10 je Karte auf der Ablage, +5 für eine aufgedeckte Karte."
    ],

    /* ------------------------------------------------------------ Aufbau */
    neu: function (opt) {
      var ziehen = opt && Number(opt.ziehen) === 3 ? 3 : 1;
      var d = K.mischen(K.deck([0, 1, 2, 3]));
      z = {
        t: [[], [], [], [], [], [], []],
        f: [[], [], [], []],
        stock: [], waste: [], ziehen: ziehen,
        punkte: 0, zuege: 0, sieg: false
      };
      var i = 0, c, r;
      for (c = 0; c < 7; c++) {
        for (r = 0; r <= c; r++) {
          var k = d[i++];
          k.offen = (r === c);
          z.t[c].push(k);
        }
      }
      z.stock = d.slice(i);
      ziele = [];
      MOD.zeichne();
    },

    snap: function () {
      return JSON.parse(JSON.stringify({
        t: z.t, f: z.f, stock: z.stock, waste: z.waste,
        ziehen: z.ziehen, punkte: z.punkte, zuege: z.zuege
      }));
    },

    restore: function (s) {
      z = s;
      z.sieg = false;
      ziele = [];
      MOD.zeichne();
    },

    info: function () {
      var n = z ? z.f.reduce(function (a, f) { return a + f.length; }, 0) : 0;
      return { punkte: z ? z.punkte : 0, zuege: z ? z.zuege : 0, extra: "Ablagen " + n + "/52" };
    },

    /* ----------------------------------------------------------- Zeichnen */
    zeichne: function () {
      if (!z) return;
      var brett = document.getElementById("brett");
      brett.innerHTML = "";
      var wrap = document.createElement("div");
      wrap.className = "klondike";

      var kopf = document.createElement("div");
      kopf.className = "kopfreihe";

      var stock = document.createElement("div");
      stock.className = "stapelfeld" + (z.stock.length ? "" : " leer");
      stock.dataset.zone = "stock";
      stock.setAttribute("aria-label", "Nachziehstapel, " + z.stock.length + " Karten");
      stock.title = z.stock.length ? "Karte aufdecken" : "Ablage zurückdrehen";
      if (z.stock.length) {
        var oben = z.stock[z.stock.length - 1];
        stock.appendChild(K.el(oben, { hinten: true, klasse: "stapelkarte" }));
        var zahl = document.createElement("span");
        zahl.className = "stapelzahl";
        zahl.textContent = z.stock.length;
        stock.appendChild(zahl);
      } else {
        stock.innerHTML = '<span class="leerzeichen">↻</span>';
      }

      var waste = document.createElement("div");
      waste.className = "stapelfeld" + (z.waste.length ? "" : " leer");
      waste.dataset.zone = "waste";
      waste.setAttribute("aria-label", "Ablagestapel");
      if (z.waste.length) {
        var wTop = z.waste[z.waste.length - 1];
        waste.appendChild(K.el(wTop, {
          art: "waste", index: z.waste.length - 1,
          klasse: (ShellAuswahl() && ShellAuswahl().art === "waste" ? "gewaehlt " : "") +
            (ziele.indexOf("waste") >= 0 ? "" : "")
        }));
        if (z.waste.length > 1) {
          var w2 = document.createElement("div");
          w2.className = "waste-rest";
          w2.appendChild(K.el(z.waste[z.waste.length - 2], { hinten: true, klasse: "klein" }));
          waste.insertBefore(w2, waste.firstChild);
        }
      } else {
        waste.innerHTML = '<span class="leerzeichen">Ablage</span>';
      }

      var luft = document.createElement("div");
      luft.className = "luft";

      var ablagen = document.createElement("div");
      ablagen.className = "ablagen";
      z.f.forEach(function (stapel, i) {
        var feld = document.createElement("div");
        feld.className = "stapelfeld ziel";
        feld.dataset.zone = "f:" + i;
        feld.setAttribute("aria-label", "Ablage " + K.farbText(i));
        if (stapel.length) {
          feld.appendChild(K.el(stapel[stapel.length - 1], {
            art: "f", i: i, index: stapel.length - 1,
            klasse: ShellAuswahl() && ShellAuswahl().art === "f" && ShellAuswahl().i === i ? "gewaehlt" : ""
          }));
        } else {
          feld.innerHTML = '<span class="leerzeichen farbe' + (K.FARBEN[i].rot ? " rot" : "") +
            '">' + K.FARBEN[i].sym + "</span>";
        }
        if (ziele.indexOf("f:" + i) >= 0) feld.classList.add("moeglich");
        ablagen.appendChild(feld);
      });

      kopf.append(stock, waste, luft, ablagen);

      var tab = document.createElement("div");
      tab.className = "tableau";
      z.t.forEach(function (spalte, i) {
        var feld = document.createElement("div");
        feld.className = "spalte";
        feld.dataset.zone = "t:" + i;
        var aus = ShellAuswahl();
        Sp.fuellen(feld, spalte, {
          art: "t", i: i,
          leerText: "K",
          abstandOffen: 0.34, abstandZu: 0.16,
          auswahl: aus && aus.art === "t" && aus.i === i ? aus.index : undefined
        });
        if (zielZone(i)) feld.classList.add("moeglich");
        tab.appendChild(feld);
      });

      wrap.append(kopf, tab);
      brett.appendChild(wrap);
    },

    /* ------------------------------------------------------------- Regeln */
    waehlen: function (was) {
      if (!z || !was) return;
      var karten = quelle(was);
      if (!karten) { TG.Shell.auswahl = null; ziele = []; MOD.zeichne(); return; }
      TG.Shell.auswahl = was;
      ziele = MOD.zieleFuer(was);
      MOD.zeichne();
      if (!ziele.length) {
        TG.Shell.hinweis("Für " + K.kartenText(karten[0]) + " liegt gerade kein Ziel frei.", true);
      } else {
        TG.Shell.hinweis("Ziel wählen – die möglichen Ablagen sind hervorgehoben.");
      }
    },

    /* Welche Zonen passen zur Auswahl? (auch für die Markierung) */
    zieleFuer: function (was) {
      var karten = quelle(was);
      if (!karten) return [];
      var out = [];
      var k = karten[0];
      if (karten.length === 1 && ablagePasst(k)) out.push("f:" + k.f);
      for (var i = 0; i < 7; i++) {
        if (was.art === "t" && was.i === i) continue;
        if (tableauPasst(karten, i)) out.push("t:" + i);
      }
      return out;
    },

    ablegen: function (was, zone) {
      if (!z || !was || !zone) return false;
      var karten = quelle(was);
      if (!karten) { TG.Shell.hinweis("Diese Karte lässt sich nicht bewegen.", true); return false; }
      if (zone.indexOf("f:") === 0) {
        var fi = Number(zone.slice(2));
        if (karten.length !== 1 || !ablagePasst(karten[0], fi)) {
          TG.Shell.hinweis("Auf die Ablage passt nur die nächste Karte derselben Farbe.", true);
          return false;
        }
        TG.Shell.merken();
        wegnehmen(was, karten.length);
        z.f[fi].push(karten[0]);
        z.punkte += 10;
      } else if (zone.indexOf("t:") === 0) {
        var ti = Number(zone.slice(2));
        if (!tableauPasst(karten, ti)) {
          TG.Shell.hinweis("Dort passt die Karte nicht – absteigend und abwechselnd rot/schwarz.", true);
          return false;
        }
        TG.Shell.merken();
        wegnehmen(was, karten.length);
        z.t[ti].push.apply(z.t[ti], karten);
        aufdecken(was);
      } else {
        return false;
      }
      z.zuege++;
      TG.Shell.auswahl = null;
      ziele = [];
      var vorher = K.positionen();
      MOD.zeichne();
      K.bewege(vorher);
      TG.Shell.zugFertig();
      pruefeSieg();
      return true;
    },

    zoneKlick: function (zone) {
      if (!z) return;
      if (zone !== "stock" && zone !== "waste") { TG.Shell.auswahl = null; ziele = []; MOD.zeichne(); return; }
      if (zone === "waste") {
        TG.Shell.auswahl = null; ziele = []; MOD.zeichne();
        return;
      }
      if (!z.stock.length && !z.waste.length) { TG.Shell.hinweis("Beide Stapel sind leer."); return; }
      TG.Shell.merken();
      var vorher = K.positionen();
      if (!z.stock.length) {
        z.stock = z.waste.reverse();
        z.stock.forEach(function (k) { k.offen = false; });
        z.waste = [];
        TG.Shell.hinweis("Ablage zurückgedreht – " + z.stock.length + " Karten im Stapel.");
      } else {
        var n = Math.min(z.ziehen, z.stock.length);
        for (var i = 0; i < n; i++) {
          var k = z.stock.pop();
          k.offen = true;
          z.waste.push(k);
        }
        TG.Shell.hinweis(n + (n === 1 ? " Karte" : " Karten") + " aufgedeckt.");
      }
      z.zuege++;
      TG.Shell.auswahl = null;
      ziele = [];
      MOD.zeichne();
      K.bewege(vorher);
      TG.Shell.zugFertig();
      pruefeSieg();
    },

    tipp: function () {
      if (!z) return "";
      var i, j, k;
      // 1. Ablage vom Nachziehstapel
      if (z.waste.length) {
        k = z.waste[z.waste.length - 1];
        if (ablagePasst(k)) return vorschlag({ art: "waste", index: z.waste.length - 1 }, "f:" + k.f, "auf die Ablage");
      }
      // 2. Tableau-Spitze auf die Ablage
      for (i = 0; i < 7; i++) {
        var s = z.t[i];
        if (!s.length) continue;
        k = s[s.length - 1];
        if (k.offen && ablagePasst(k)) {
          return vorschlag({ art: "t", i: i, index: s.length - 1 }, "f:" + k.f, "auf die Ablage");
        }
      }
      // 3. Tableau umschichten, möglichst mit aufgedeckter Karte darunter
      for (i = 0; i < 7; i++) {
        for (j = 0; j < z.t[i].length; j++) {
          var kk = z.t[i][j];
          if (!kk.offen) continue;
          var folge = z.t[i].slice(j);
          if (!folgeFrei(folge)) continue;
          for (var t = 0; t < 7; t++) {
            if (t === i) continue;
            if (!tableauPasst(folge, t)) continue;
            var bringt = j > 0 && !z.t[i][j - 1].offen;
            if (!bringt && !z.t[t].length) continue;
            return vorschlag({ art: "t", i: i, index: j }, "t:" + t, "in die markierte Spalte");
          }
        }
      }
      if (z.stock.length) {
        return "Tipp: links oben neue Karten aufdecken.";
      }
      return "Kein Zug möglich – nimm einen Zug zurück oder beginne neu.";
    },

    brettKlick: function () {
      TG.Shell.auswahl = null;
      ziele = [];
      MOD.zeichne();
    }
  };

  function ShellAuswahl() { return TG.Shell.auswahl; }
  function zielZone(i) { return ziele.indexOf("t:" + i) >= 0; }

  function vorschlag(was, zone, text) {
    TG.Shell.auswahl = was;
    ziele = [zone];
    MOD.zeichne();
    var k = quelle(was)[0];
    return "Tipp: " + K.kartenText(k) + " " + text + ".";
  }

  /* Karten hinter einer Auswahl (Array) oder null. */
  function quelle(was) {
    if (!z || !was) return null;
    if (was.art === "waste") {
      if (!z.waste.length || was.index !== z.waste.length - 1) return null;
      return [z.waste[z.waste.length - 1]];
    }
    if (was.art === "f") {
      var f = z.f[was.i];
      if (!f || !f.length || was.index !== f.length - 1) return null;
      return [f[f.length - 1]];
    }
    if (was.art === "t") {
      var s = z.t[was.i];
      if (!s || was.index < 0 || was.index >= s.length) return null;
      var folge = s.slice(was.index);
      if (!folgeFrei(folge)) return null;
      return folge;
    }
    return null;
  }

  function folgeFrei(folge) {
    for (var i = 0; i < folge.length; i++) {
      if (!folge[i].offen) return false;
      if (i > 0) {
        if (folge[i].r !== folge[i - 1].r - 1) return false;
        if (K.rot(folge[i]) === K.rot(folge[i - 1])) return false;
      }
    }
    return true;
  }

  function ablagePasst(k, fi) {
    if (fi === undefined) fi = k.f;
    if (fi !== k.f) return false;
    var f = z.f[fi];
    if (!f.length) return k.r === 1;
    return f[f.length - 1].r === k.r - 1;
  }

  function tableauPasst(karten, i) {
    var s = z.t[i];
    var k = karten[0];
    if (!s.length) return k.r === 13;
    var oben = s[s.length - 1];
    return oben.offen && oben.r === k.r + 1 && K.rot(oben) !== K.rot(k);
  }

  function wegnehmen(was, anzahl) {
    if (was.art === "waste") z.waste.pop();
    else if (was.art === "f") z.f[was.i].pop();
    else z.t[was.i].splice(z.t[was.i].length - anzahl, anzahl);
  }

  function aufdecken(was) {
    if (was.art !== "t") return;
    var s = z.t[was.i];
    if (s.length && !s[s.length - 1].offen) {
      s[s.length - 1].offen = true;
      z.punkte += 5;
    }
  }

  function pruefeSieg() {
    if (!z) return;
    if (z.f.every(function (f) { return f.length === 13; })) {
      z.sieg = true;
      TG.Shell.sieg("Punkte " + z.punkte + " – sauber abgelegt.");
      return;
    }
    if (z.stock.length || z.waste.length) return;
    if (hatZug()) return;
    z.ende = true;
    TG.Shell.verloren("Nachziehstapel und Ablage sind leer, und keine Karte lässt sich mehr bewegen.");
  }

  /* Gibt es überhaupt noch einen regelgerechten Zug? */
  function hatZug() {
    var i, j, t;
    for (i = 0; i < 7; i++) {
      var s = z.t[i];
      for (j = 0; j < s.length; j++) {
        if (!s[j].offen) continue;
        var folge = s.slice(j);
        if (!folgeFrei(folge)) continue;
        if (folge.length === 1 && ablagePasst(folge[0])) return true;
        for (t = 0; t < 7; t++) {
          if (t === i) continue;
          if (!tableauPasst(folge, t)) continue;
          if (s.length === folge.length && !z.t[t].length) continue;
          return true;
        }
      }
    }
    return false;
  }
})();
