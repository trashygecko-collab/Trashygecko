/* =========================================================================
   FREECELL – 52 Karten, acht Spalten, vier freie Zellen, vier Ablagen.
   Es dürfen nur so viele Karten auf einmal umgelegt werden, wie freie
   Zellen und leere Spalten als Zwischenlager erlauben.
   ========================================================================= */
(function () {
  "use strict";
  var TG = window.TG, K = TG.Karten, Sp = TG.Spalte;
  var MODI = TG.MODI = TG.MODI || {};

  var z = null;
  var ziele = [];

  var MOD = MODI.freecell = {
    id: "freecell",
    name: "FreeCell",
    kurz: "Alle 52 Karten offen, vier freie Zellen als Zwischenlager. " +
      "Jede Karte lässt sich genau überlegen – fast jedes Blatt ist lösbar.",
    vorschau: '<svg viewBox="0 0 120 84" aria-hidden="true">' +
      '<g style="fill:none;stroke:var(--vv-rand)" stroke-width="1.2">' +
      '<rect x="4" y="4" width="14" height="20" rx="3"/><rect x="20" y="4" width="14" height="20" rx="3"/>' +
      '<rect x="36" y="4" width="14" height="20" rx="3"/><rect x="52" y="4" width="14" height="20" rx="3"/>' +
      '<rect x="74" y="4" width="20" height="28" rx="3"/><rect x="96" y="4" width="20" height="28" rx="3"/></g>' +
      '<g style="fill:var(--vv-karte);stroke:var(--vv-rand)" stroke-width="1.2">' +
      '<rect x="4" y="40" width="20" height="28" rx="3"/><rect x="28" y="40" width="20" height="28" rx="3"/>' +
      '<rect x="52" y="40" width="20" height="28" rx="3"/><rect x="76" y="40" width="20" height="28" rx="3"/>' +
      '<rect x="100" y="40" width="16" height="24" rx="3"/></g>' +
      '<text x="8" y="56" style="fill:var(--vv-hell)" font-size="10" font-family="Georgia,serif">7</text>' +
      '<text x="32" y="56" style="fill:var(--vv-rot)" font-size="10" font-family="Georgia,serif">6</text>' +
      '<text x="56" y="56" style="fill:var(--vv-hell)" font-size="10" font-family="Georgia,serif">5</text>' +
      '<text x="78" y="18" style="fill:var(--vv-rot)" font-size="10" font-family="Georgia,serif">A</text>' +
      "</svg>",
    optionen: [],
    standard: {},
    regeln: [
      "<b>Ziel:</b> alle Karten auf die vier Ablagen – jede Farbe von Ass bis König.",
      "<b>Blatt:</b> alle 52 Karten offen in acht Spalten (viermal sieben, viermal sechs).",
      "<b>Tableau:</b> absteigend und abwechselnd rot und schwarz. Leere Spalten sind frei.",
      "<b>Freie Zellen:</b> vier Plätze für je eine Karte als Zwischenlager.",
      "<b>Bewegen:</b> höchstens (freie Zellen + 1) × 2^(leere Spalten) Karten auf einmal.",
      "<b>Punkte:</b> +10 für jede Karte, die auf einer Ablage landet."
    ],

    neu: function () {
      var d = K.mischen(K.deck([0, 1, 2, 3]));
      z = { t: [], zellen: [null, null, null, null], f: [[], [], [], []], punkte: 0, zuege: 0, sieg: false };
      for (var i = 0; i < 8; i++) z.t.push([]);
      var pos = 0;
      for (var c = 0; c < 8; c++) {
        var n = c < 4 ? 7 : 6;
        for (var r = 0; r < n; r++) {
          var k = d[pos++];
          k.offen = true;
          z.t[c].push(k);
        }
      }
      ziele = [];
      MOD.zeichne();
    },

    snap: function () {
      return JSON.parse(JSON.stringify({
        t: z.t, zellen: z.zellen, f: z.f, punkte: z.punkte, zuege: z.zuege
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
      var frei = z ? z.zellen.filter(function (c) { return !c; }).length : 4;
      return { punkte: z ? z.punkte : 0, zuege: z ? z.zuege : 0, extra: "Ablagen " + n + "/52 · Zellen " + frei + "/4" };
    },

    zeichne: function () {
      if (!z) return;
      var brett = document.getElementById("brett");
      brett.innerHTML = "";
      var wrap = document.createElement("div");
      wrap.className = "freecell";

      var kopf = document.createElement("div");
      kopf.className = "kopfreihe";
      var zellen = document.createElement("div");
      zellen.className = "zellen";
      var aus = TG.Shell.auswahl;

      z.zellen.forEach(function (k, i) {
        var feld = document.createElement("div");
        feld.className = "stapelfeld zelle" + (k ? " voll" : " leer");
        feld.dataset.zone = "zelle:" + i;
        feld.setAttribute("aria-label", "Freie Zelle " + (i + 1));
        if (k) {
          feld.appendChild(K.el(k, {
            art: "zelle", i: i, index: 0,
            klasse: aus && aus.art === "zelle" && aus.i === i ? "gewaehlt" : ""
          }));
        } else {
          feld.innerHTML = '<span class="leerzeichen">frei</span>';
        }
        if (ziele.indexOf("zelle:" + i) >= 0) feld.classList.add("moeglich");
        zellen.appendChild(feld);
      });

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
            klasse: aus && aus.art === "f" && aus.i === i ? "gewaehlt" : ""
          }));
        } else {
          feld.innerHTML = '<span class="leerzeichen farbe' + (K.FARBEN[i].rot ? " rot" : "") + '">' +
            K.FARBEN[i].sym + "</span>";
        }
        if (ziele.indexOf("f:" + i) >= 0) feld.classList.add("moeglich");
        ablagen.appendChild(feld);
      });

      kopf.append(zellen, luft, ablagen);

      var tab = document.createElement("div");
      tab.className = "tableau acht";
      z.t.forEach(function (spalte, i) {
        var feld = document.createElement("div");
        feld.className = "spalte";
        feld.dataset.zone = "t:" + i;
        Sp.fuellen(feld, spalte, {
          art: "t", i: i,
          abstandOffen: 0.30, abstandZu: 0.14,
          auswahl: aus && aus.art === "t" && aus.i === i ? aus.index : undefined
        });
        if (ziele.indexOf("t:" + i) >= 0) feld.classList.add("moeglich");
        tab.appendChild(feld);
      });

      wrap.append(kopf, tab);
      brett.appendChild(wrap);
    },

    waehlen: function (was) {
      if (!z || !was) return;
      var karten = quelle(was);
      if (!karten) { TG.Shell.auswahl = null; ziele = []; MOD.zeichne(); return; }
      TG.Shell.auswahl = was;
      ziele = MOD.zieleFuer(was);
      MOD.zeichne();
      TG.Shell.hinweis(ziele.length ? "Ziel wählen." : "Für diese Karte ist kein Platz frei.", !ziele.length);
    },

    zieleFuer: function (was) {
      var karten = quelle(was);
      if (!karten) return [];
      var out = [];
      var k = karten[0];
      if (karten.length === 1 && ablagePasst(k)) out.push("f:" + k.f);
      for (var i = 0; i < 4; i++) {
        if (was.art === "zelle" && was.i === i) continue;
        if (!z.zellen[i] && karten.length === 1) out.push("zelle:" + i);
      }
      for (var t = 0; t < 8; t++) {
        if (was.art === "t" && was.i === t) continue;
        if (!tableauPasst(karten, t)) continue;
        if (karten.length > maxFolge(t)) continue;
        out.push("t:" + t);
      }
      return out;
    },

    ablegen: function (was, zone) {
      if (!z || !was) return false;
      var karten = quelle(was);
      if (!karten) { TG.Shell.hinweis("Diese Karten lassen sich nicht bewegen.", true); return false; }
      if (zone.indexOf("f:") === 0) {
        var fi = Number(zone.slice(2));
        if (karten.length !== 1 || !ablagePasst(karten[0], fi)) {
          TG.Shell.hinweis("Auf die Ablage passt nur die nächste Karte derselben Farbe.", true);
          return false;
        }
        TG.Shell.merken();
        weg(was, 1);
        z.f[fi].push(karten[0]);
        z.punkte += 10;
      } else if (zone.indexOf("zelle:") === 0) {
        var zi = Number(zone.slice(6));
        if (karten.length !== 1 || z.zellen[zi]) {
          TG.Shell.hinweis("In eine freie Zelle passt genau eine Karte.", true);
          return false;
        }
        TG.Shell.merken();
        weg(was, 1);
        z.zellen[zi] = karten[0];
      } else if (zone.indexOf("t:") === 0) {
        var ti = Number(zone.slice(2));
        if (!tableauPasst(karten, ti)) {
          TG.Shell.hinweis("Dort passt die Karte nicht – absteigend und abwechselnd rot/schwarz.", true);
          return false;
        }
        if (karten.length > maxFolge(ti)) {
          TG.Shell.hinweis("Zu viele Karten auf einmal – dafür fehlen freie Zellen oder leere Spalten.", true);
          return false;
        }
        TG.Shell.merken();
        weg(was, karten.length);
        z.t[ti].push.apply(z.t[ti], karten);
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

    zoneKlick: function () {
      TG.Shell.auswahl = null; ziele = []; MOD.zeichne();
    },

    doppelklick: function (was) {
      var ks = quelle(was);
      if (!ks || ks.length !== 1 || ks[0].id !== was.id || !ablagePasst(ks[0])) return false;
      return MOD.ablegen(was, "f:" + ks[0].f);
    },

    autoMoeglich: function () { return autoPlan() !== null; },
    autoVollenden: function () {
      if (!autoPlan()) return;
      var weiter = true;
      while (weiter) {
        weiter = false;
        for (var i = 0; i < 4 && !weiter; i++) {
          var c = z.zellen[i];
          if (c && ablagePasst(c)) weiter = MOD.ablegen({ art: "zelle", i: i, index: 0 }, "f:" + c.f);
        }
        for (var j = 0; j < 8 && !weiter; j++) {
          var s = z.t[j]; if (!s.length) continue;
          var k = s[s.length - 1];
          if (ablagePasst(k)) weiter = MOD.ablegen({ art: "t", i: j, index: s.length - 1 }, "f:" + k.f);
        }
      }
    },

    tipp: function () {
      if (!z) return "";
      var i, j;
      for (i = 0; i < 8; i++) {
        var s = z.t[i];
        if (!s.length) continue;
        var k = s[s.length - 1];
        if (ablagePasst(k)) {
          TG.Shell.auswahl = { art: "t", i: i, index: s.length - 1 };
          ziele = ["f:" + k.f];
          MOD.zeichne();
          return "Tipp: " + K.kartenText(k) + " auf die Ablage legen.";
        }
      }
      for (i = 0; i < 4; i++) {
        var c = z.zellen[i];
        if (c && ablagePasst(c)) {
          TG.Shell.auswahl = { art: "zelle", i: i, index: 0 };
          ziele = ["f:" + c.f];
          MOD.zeichne();
          return "Tipp: " + K.kartenText(c) + " aus der Zelle auf die Ablage legen.";
        }
      }
      for (i = 0; i < 8; i++) {
        for (j = 0; j < z.t[i].length; j++) {
          var folge = z.t[i].slice(j);
          if (!folgeFrei(folge)) continue;
          for (var t = 0; t < 8; t++) {
            if (t === i || !tableauPasst(folge, t) || folge.length > maxFolge(t)) continue;
            if (!z.t[t].length && !z.t[i].length) continue;
            TG.Shell.auswahl = { art: "t", i: i, index: j };
            ziele = ["t:" + t];
            MOD.zeichne();
            return "Tipp: die markierte Reihe in die hervorgehobene Spalte legen.";
          }
        }
      }
      return "Kein Zug möglich – nimm einen Zug zurück.";
    }
  };

  function autoPlan() {
    if (!z) return null;
    var spalten = z.t.map(function (s) { return s.slice(); });
    var zellen = z.zellen.slice();
    var rang = z.f.map(function (s) { return s.length; });
    var plan = [], fortschritt = true;
    while (fortschritt) {
      fortschritt = false;
      for (var i = 0; i < zellen.length; i++) {
        var c = zellen[i];
        if (c && c.r === rang[c.f] + 1) {
          plan.push({ was: { art: "zelle", i: i, index: 0 }, f: c.f });
          zellen[i] = null; rang[c.f]++; fortschritt = true;
        }
      }
      for (var j = 0; j < spalten.length; j++) {
        var s = spalten[j]; if (!s.length) continue;
        var k = s[s.length - 1];
        if (k.r === rang[k.f] + 1) {
          plan.push({ was: { art: "t", i: j, index: s.length - 1 }, f: k.f });
          s.pop(); rang[k.f]++; fortschritt = true;
        }
      }
    }
    return zellen.every(function (c) { return !c; }) && spalten.every(function (s) { return !s.length; }) ? plan : null;
  }



  /* Höchstzahl gleichzeitig bewegbarer Karten für Zielspalte zielIndex.
     Eine leere Zielspalte zählt nicht als Zwischenlager. */
  function maxFolge(zielIndex) {
    var frei = z.zellen.filter(function (c) { return !c; }).length;
    var leer = z.t.filter(function (s) { return !s.length; }).length;
    if (zielIndex !== undefined && !z.t[zielIndex].length) leer--;
    return (frei + 1) * Math.pow(2, Math.max(0, leer));
  }

  function quelle(was) {
    if (!z || !was) return null;
    if (was.art === "zelle") {
      var c = z.zellen[was.i];
      return c ? [c] : null;
    }
    if (was.art === "f") {
      var f = z.f[was.i];
      if (!f || !f.length) return null;
      return [f[f.length - 1]];
    }
    if (was.art === "t") {
      var s = z.t[was.i];
      if (!s || was.index < 0 || was.index >= s.length) return null;
      var folge = s.slice(was.index);
      return folgeFrei(folge) ? folge : null;
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

  function tableauPasst(karten, t) {
    var s = z.t[t];
    var k = karten[0];
    if (!s.length) return true;
    var oben = s[s.length - 1];
    return oben.r === k.r + 1 && K.rot(oben) !== K.rot(k);
  }

  function weg(was, anzahl) {
    if (was.art === "zelle") z.zellen[was.i] = null;
    else if (was.art === "f") z.f[was.i].pop();
    else z.t[was.i].splice(z.t[was.i].length - anzahl, anzahl);
  }

  function pruefeSieg() {
    if (!z || z.sieg) return;
    if (z.f.every(function (f) { return f.length === 13; })) {
      z.sieg = true;
      TG.Shell.sieg("Alle 52 Karten liegen auf den Ablagen.");
      return;
    }
    if (hatZug()) return;
    z.ende = true;
    TG.Shell.verloren("Zellen, Spalten und Ablagen lassen keinen regelgerechten Zug mehr zu.");
  }

  function hatZug() {
    var i, j, t;
    for (i = 0; i < 8; i++) {
      var s = z.t[i];
      if (s.length && ablagePasst(s[s.length - 1])) return true;
    }
    for (i = 0; i < 4; i++) {
      if (z.zellen[i] && ablagePasst(z.zellen[i])) return true;
      if (!z.zellen[i]) return true;                       // Karte in eine freie Zelle
    }
    for (i = 0; i < 8; i++) {
      for (j = 0; j < z.t[i].length; j++) {
        var folge = z.t[i].slice(j);
        if (!folgeFrei(folge)) continue;
        for (t = 0; t < 8; t++) {
          if (t === i || !tableauPasst(folge, t) || folge.length > maxFolge(t)) continue;
          if (!z.t[t].length && !j) continue;
          return true;
        }
      }
    }
    return false;
  }
})();
