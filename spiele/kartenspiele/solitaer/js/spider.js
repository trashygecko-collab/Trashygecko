/* =========================================================================
   SPIDER – 104 Karten, zehn Spalten, ein bis vier Farben, fünf Nachlegen,
   acht Folgen König bis Ass. Regeln aus dem bisherigen Spider-Solitär.
   ========================================================================= */
(function () {
  "use strict";
  var TG = window.TG, K = TG.Karten, Sp = TG.Spalte;
  var MODI = TG.MODI = TG.MODI || {};

  var z = null;
  var ziele = [];

  var MOD = MODI.spider = {
    id: "spider",
    name: "Spider",
    kurz: "Zwei Blätter, zehn Spalten. Absteigend anlegen, vollständige Folgen " +
      "König bis Ass in einer Farbe verschwinden. Mit fünf Nachlege-Runden.",
    vorschau: '<svg viewBox="0 0 120 84" aria-hidden="true">' +
      '<g style="fill:var(--vv-karte);stroke:var(--vv-rand)" stroke-width="1.2">' +
      '<rect x="4" y="6" width="22" height="31" rx="3"/><rect x="29" y="6" width="22" height="31" rx="3"/>' +
      '<rect x="54" y="6" width="22" height="31" rx="3"/><rect x="79" y="6" width="22" height="31" rx="3"/>' +
      '<rect x="4" y="44" width="22" height="31" rx="3"/><rect x="29" y="44" width="22" height="31" rx="3"/>' +
      '<rect x="54" y="44" width="22" height="31" rx="3"/><rect x="79" y="44" width="22" height="31" rx="3"/>' +
      '<rect x="100" y="44" width="16" height="22" rx="3"/></g>' +
      '<text x="8" y="20" style="fill:var(--vv-hell)" font-size="11" font-family="Georgia,serif">K</text>' +
      '<text x="33" y="20" style="fill:var(--vv-rot)" font-size="11" font-family="Georgia,serif">D</text>' +
      '<text x="58" y="20" style="fill:var(--vv-hell)" font-size="11" font-family="Georgia,serif">B</text>' +
      '<text x="8" y="58" style="fill:var(--vv-hell)" font-size="11" font-family="Georgia,serif">A</text>' +
      '<text x="103" y="58" style="fill:var(--vv-leise)" font-size="10" font-family="Georgia,serif">8</text>' +
      "</svg>",
    optionen: [{ key: "farben", label: "Farben", werte: [
      { wert: 1, text: "eine (leicht)" }, { wert: 2, text: "zwei" }, { wert: 4, text: "vier (schwer)" }
    ] }],
    standard: { farben: 1 },
    regeln: [
      "<b>Ziel:</b> acht vollständige Folgen König bis Ass in einer Farbe bilden – sie verschwinden nach unten links.",
      "<b>Blatt:</b> 104 Karten in zehn Spalten, die ersten vier mit sechs, die übrigen sechs mit fünf Karten.",
      "<b>Anlegen:</b> absteigend auf eine Karte, die einen Rang höher liegt – die Farbe ist dabei egal. Bewegt werden nur absteigende Reihen gleicher Farbe.",
      "<b>Leere Spalte:</b> darf mit jeder Karte belegt werden.",
      "<b>Nachlegen:</b> unten rechts fünf Runden zu je zehn Karten – nur wenn keine Spalte leer ist.",
      "<b>Punkte:</b> Start 500, jeder Zug kostet einen Punkt, jede fertige Folge bringt 100."
    ],

    neu: function (opt) {
      var farbenAnzahl = opt && Number(opt.farben) ? Number(opt.farben) : 1;
      var farbenListe = [];
      var i;
      if (farbenAnzahl === 1) {
        for (i = 0; i < 8; i++) farbenListe.push(0);
      } else if (farbenAnzahl === 2) {
        for (i = 0; i < 4; i++) { farbenListe.push(0); farbenListe.push(1); }
      } else {
        for (i = 0; i < 2; i++) farbenListe.push(0, 1, 2, 3);
      }
      var karten = K.mischen(K.deck(farbenListe));
      z = { t: [], ablage: [], stock: [], farben: farbenAnzahl, punkte: 500, zuege: 0, sieg: false };
      for (i = 0; i < 10; i++) z.t.push([]);
      var pos = 0, s, n;
      for (s = 0; s < 10; s++) {
        n = s < 4 ? 6 : 5;
        for (i = 0; i < n; i++) z.t[s].push(karten[pos++]);
        z.t[s][z.t[s].length - 1].offen = true;
      }
      z.stock = karten.slice(pos);
      ziele = [];
      MOD.zeichne();
    },

    snap: function () {
      return JSON.parse(JSON.stringify({
        t: z.t, ablage: z.ablage, stock: z.stock, farben: z.farben,
        punkte: z.punkte, zuege: z.zuege
      }));
    },

    restore: function (s) {
      z = s;
      z.sieg = false;
      ziele = [];
      MOD.zeichne();
    },

    info: function () {
      return {
        punkte: z ? z.punkte : 500, zuege: z ? z.zuege : 0,
        extra: z ? "Folgen " + z.ablage.length + "/8 · Vorrat " + Math.floor(z.stock.length / 10) : ""
      };
    },

    zeichne: function () {
      if (!z) return;
      var brett = document.getElementById("brett");
      brett.innerHTML = "";
      var wrap = document.createElement("div");
      wrap.className = "spider";

      var tab = document.createElement("div");
      tab.className = "tableau zehn";
      z.t.forEach(function (spalte, i) {
        var feld = document.createElement("div");
        feld.className = "spalte";
        feld.dataset.zone = "t:" + i;
        var aus = TG.Shell.auswahl;
        Sp.fuellen(feld, spalte, {
          art: "t", i: i,
          abstandOffen: 0.30, abstandZu: 0.13,
          auswahl: aus && aus.art === "t" && aus.i === i ? aus.index : undefined
        });
        if (ziele.indexOf("t:" + i) >= 0) feld.classList.add("moeglich");
        tab.appendChild(feld);
      });

      var fuss = document.createElement("div");
      fuss.className = "spider-fuss";

      var fertig = document.createElement("div");
      fertig.className = "fertige";
      for (var i = 0; i < 8; i++) {
        var platz = document.createElement("div");
        platz.className = "fertigplatz";
        var farbe = z.ablage[i];
        if (farbe === undefined) {
          platz.innerHTML = '<span class="leerzeichen">K</span>';
          platz.title = "Hierhin wandert eine fertige Folge König bis Ass";
        } else {
          platz.className += " voll";
          platz.appendChild(K.el({ r: 13, f: farbe }, { klasse: "ohne-klick" }));
          platz.title = "Fertige Folge " + K.farbText(farbe);
        }
        fertig.appendChild(platz);
      }

      var vorrat = document.createElement("div");
      vorrat.className = "vorrat" + (z.stock.length ? "" : " leer");
      vorrat.dataset.zone = "stock";
      vorrat.setAttribute("aria-label", "Vorrat: " + z.stock.length + " Karten");
      if (z.stock.length) {
        vorrat.appendChild(K.el(z.stock[z.stock.length - 1], { hinten: true, klasse: "stapelkarte" }));
        var zahl = document.createElement("span");
        zahl.className = "stapelzahl";
        zahl.textContent = Math.floor(z.stock.length / 10) + "×";
        vorrat.appendChild(zahl);
        vorrat.title = "Zehn Karten nachlegen";
      } else {
        vorrat.innerHTML = '<span class="leerzeichen">leer</span>';
      }

      fuss.append(fertig, vorrat);
      wrap.append(tab, fuss);
      brett.appendChild(wrap);
    },

    waehlen: function (was) {
      if (!z || !was) return;
      var folge = quelle(was);
      if (!folge) {
        TG.Shell.auswahl = null; ziele = []; MOD.zeichne();
        TG.Shell.hinweis("Nur offene, absteigende Reihen gleicher Farbe lassen sich bewegen.", true);
        return;
      }
      TG.Shell.auswahl = was;
      ziele = MOD.zieleFuer(was);
      MOD.zeichne();
      TG.Shell.hinweis(ziele.length
        ? "Ziel wählen – markierte Spalten sind möglich."
        : "Für diese Reihe ist keine Spalte frei.", !ziele.length);
    },

    zieleFuer: function (was) {
      var folge = quelle(was);
      if (!folge) return [];
      var out = [];
      for (var i = 0; i < 10; i++) {
        if (was.art === "t" && was.i === i) continue;
        if (passtAuf(folge[0], i)) out.push("t:" + i);
      }
      return out;
    },

    ablegen: function (was, zone) {
      if (!z || !was || zone.indexOf("t:") !== 0) return false;
      var folge = quelle(was);
      var ti = Number(zone.slice(2));
      if (!folge) { TG.Shell.hinweis("Diese Reihe lässt sich nicht bewegen.", true); return false; }
      if (!passtAuf(folge[0], ti)) {
        TG.Shell.hinweis("Dort passt die Karte nicht – sie muss einen Rang niedriger sein.", true);
        return false;
      }
      TG.Shell.merken();
      z.t[was.i].length = was.index;
      z.t[ti].push.apply(z.t[ti], folge);
      var rest = z.t[was.i];
      if (rest.length && !rest[rest.length - 1].offen) rest[rest.length - 1].offen = true;
      z.punkte = Math.max(0, z.punkte - 1);
      z.zuege++;
      TG.Shell.auswahl = null;
      ziele = [];
      folgenPruefen();
      var vorher = K.positionen();
      MOD.zeichne();
      K.bewege(vorher);
      TG.Shell.zugFertig();
      pruefeSieg();
      return true;
    },

    zoneKlick: function (zone) {
      if (!z) return;
      if (zone === "stock") { MOD.nachlegen(); return; }
      TG.Shell.auswahl = null; ziele = []; MOD.zeichne();
    },

    nachlegen: function () {
      if (!z.stock.length) { TG.Shell.hinweis("Der Vorrat ist leer.", true); pruefeSieg(); return; }
      if (z.t.some(function (s) { return s.length === 0; })) {
        TG.Shell.hinweis("Nachlegen geht nur, wenn keine Spalte leer ist.", true);
        return;
      }
      TG.Shell.merken();
      for (var s = 0; s < 10; s++) {
        if (!z.stock.length) break;
        var k = z.stock.pop();
        k.offen = true;
        z.t[s].push(k);
      }
      z.zuege++;
      TG.Shell.auswahl = null;
      ziele = [];
      folgenPruefen();
      var vorher = K.positionen();
      MOD.zeichne();
      K.bewege(vorher);
      TG.Shell.zugFertig("Zehn neue Karten liegen auf dem Tisch.");
      pruefeSieg();
    },

    doppelklick: function (was) {
      var folge = quelle(was);
      if (!folge || folge[0].id !== was.id) return false;
      var ziele = MOD.zieleFuer(was);
      if (ziele.length !== 1) return false;
      return MOD.ablegen(was, ziele[0]);
    },

    tipp: function () {
      if (!z) return "";
      var bester = null, von, i, nach;
      for (von = 0; von < 10; von++) {
        for (i = 0; i < z.t[von].length; i++) {
          var folge = quelle({ art: "t", i: von, index: i });
          if (!folge) continue;
          for (nach = 0; nach < 10; nach++) {
            if (nach === von || !passtAuf(folge[0], nach)) continue;
            var ziel = z.t[nach];
            var gleicheFarbe = ziel.length > 0 && ziel[ziel.length - 1].f === folge[0].f;
            var aufLeer = ziel.length === 0;
            if (aufLeer && z.t[von].length === folge.length) continue;
            var wert = (gleicheFarbe ? 3 : 0) + (aufLeer ? 0 : 1);
            if (!bester || wert > bester.wert) bester = { von: von, index: i, nach: nach, wert: wert };
          }
        }
      }
      if (!bester) {
        if (z.stock.length && !z.t.some(function (s) { return !s.length; })) {
          return "Tipp: zehn neue Karten nachlegen.";
        }
        return "Kein Zug möglich – nimm einen Zug zurück.";
      }
      TG.Shell.auswahl = { art: "t", i: bester.von, index: bester.index };
      ziele = ["t:" + bester.nach];
      MOD.zeichne();
      return "Tipp: die markierte Reihe in die hervorgehobene Spalte legen.";
    }
  };

  /* Anlegen: eine Karte darf auf jede Karte einen Rang höher – Farbe egal. */
  function passtAuf(karte, zielSpalte) {
    var ziel = z.t[zielSpalte];
    if (!ziel.length) return true;
    var oben = ziel[ziel.length - 1];
    return oben.offen && oben.r === karte.r + 1;
  }

  function quelle(was) {
    if (!z || !was || was.art !== "t") return null;
    var s = z.t[was.i];
    if (!s || was.index < 0 || was.index >= s.length) return null;
    for (var i = was.index; i < s.length; i++) {
      if (!s[i].offen) return null;
      if (i > was.index) {
        if (s[i].f !== s[i - 1].f) return null;
        if (s[i].r !== s[i - 1].r - 1) return null;
      }
    }
    return s.slice(was.index);
  }

  function folgenPruefen() {
    for (var s = 0; s < 10; s++) {
      var karten = z.t[s];
      if (karten.length < 13) continue;
      var start = karten.length - 13;
      var passt = karten[start].offen && karten[start].r === 13;
      for (var i = start; i < karten.length && passt; i++) {
        if (!karten[i].offen) passt = false;
        if (karten[i].f !== karten[start].f) passt = false;
        if (karten[i].r !== 13 - (i - start)) passt = false;
      }
      if (!passt) continue;
      var folge = karten.splice(start, 13);
      z.ablage.push(folge[0].f);
      z.punkte += 100;
      if (karten.length && !karten[karten.length - 1].offen) karten[karten.length - 1].offen = true;
    }
  }

  function pruefeSieg() {
    if (z.sieg) return;
    if (z.ablage.length === 8 && z.t.every(function (s) { return s.length === 0; })) {
      z.sieg = true;
      TG.Shell.sieg("Acht Folgen – " + z.punkte + " Punkte.");
      return;
    }
    if (z.stock.length) return;
    if (hatZug()) return;
    z.ende = true;
    TG.Shell.verloren("Der Vorrat ist leer, und es lässt sich keine Reihe mehr umlegen.");
  }

  function hatZug() {
    for (var i = 0; i < 10; i++) {
      for (var j = 0; j < z.t[i].length; j++) {
        var folge = quelle({ art: "t", i: i, index: j });
        if (!folge) continue;
        for (var t = 0; t < 10; t++) {
          if (t === i || !passtAuf(folge[0], t)) continue;
          if (!z.t[t].length && z.t[i].length === folge.length) continue;
          return true;
        }
      }
    }
    return false;
  }
})();
