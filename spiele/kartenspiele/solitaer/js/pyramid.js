/* =========================================================================
   PYRAMID – 28 Karten in sieben Reihen. Zwei freie Karten mit der Summe 13
   verschwinden, Könige (13) allein. Nachziehstapel und Ablage wie beim
   klassischen Pyramid-Solitär.
   ========================================================================= */
(function () {
  "use strict";
  var TG = window.TG, K = TG.Karten;
  var MODI = TG.MODI = TG.MODI || {};

  var z = null;          // {pyr:[28], stock, waste, punkte, zuege, sieg}
  var auswahl = null;    // {typ:"pyr",i} | {typ:"waste"}

  var MOD = MODI.pyramid = {
    id: "pyramid",
    name: "Pyramid",
    kurz: "28 Karten in sieben Reihen. Immer zwei freie Karten mit der Summe 13 " +
      "abräumen – ein König zählt 13 und geht allein.",
    vorschau: '<svg viewBox="0 0 120 84" aria-hidden="true">' +
      '<g style="fill:var(--vv-karte);stroke:var(--vv-rand)" stroke-width="1.2">' +
      '<rect x="52" y="4" width="18" height="14" rx="2"/><rect x="43" y="20" width="18" height="14" rx="2"/>' +
      '<rect x="61" y="20" width="18" height="14" rx="2"/><rect x="34" y="36" width="18" height="14" rx="2"/>' +
      '<rect x="52" y="36" width="18" height="14" rx="2"/><rect x="70" y="36" width="18" height="14" rx="2"/>' +
      '<rect x="25" y="52" width="18" height="14" rx="2"/><rect x="43" y="52" width="18" height="14" rx="2"/>' +
      '<rect x="61" y="52" width="18" height="14" rx="2"/><rect x="79" y="52" width="18" height="14" rx="2"/>' +
      '<rect x="4" y="70" width="16" height="12" rx="2"/><rect x="100" y="70" width="16" height="12" rx="2"/></g>' +
      '<text x="57" y="15" style="fill:var(--vv-rot)" font-size="9" font-family="Georgia,serif">K</text>' +
      '<text x="48" y="31" style="fill:var(--vv-hell)" font-size="9" font-family="Georgia,serif">A</text>' +
      '<text x="66" y="31" style="fill:var(--vv-hell)" font-size="9" font-family="Georgia,serif">D</text>' +
      "</svg>",
    optionen: [],
    standard: {},
    regeln: [
      "<b>Ziel:</b> alle 28 Karten der Pyramide abräumen.",
      "<b>Aufbau:</b> sieben Reihen mit 1 bis 7 Karten, alle offen.",
      "<b>Paare:</b> zwei freie Karten, deren Werte zusammen 13 ergeben (Ass = 1, Bube = 11, Dame = 12), verschwinden zusammen.",
      "<b>König:</b> zählt 13 und wird allein abgeräumt.",
      "<b>Frei:</b> eine Karte ist frei, wenn keine Karte mehr auf ihr liegt – die unterste Reihe ist immer frei.",
      "<b>Nachziehen:</b> links unten eine Karte aufdecken; sie darf mit einer freien Pyramidenkarte ein Paar bilden.",
      "<b>Punkte:</b> +10 für jedes abgeräumte Paar."
    ],

    neu: function () {
      var d = K.mischen(K.deck([0, 1, 2, 3]));
      z = { pyr: new Array(28).fill(null), stock: [], waste: [], punkte: 0, zuege: 0, sieg: false };
      for (var i = 0; i < 28; i++) {
        var k = d[i];
        k.offen = true;
        z.pyr[i] = k;
      }
      z.waste = [d[28]];
      z.stock = d.slice(29);
      auswahl = null;
      MOD.zeichne();
    },

    snap: function () {
      return JSON.parse(JSON.stringify({
        pyr: z.pyr, stock: z.stock, waste: z.waste, punkte: z.punkte, zuege: z.zuege
      }));
    },

    restore: function (s) {
      z = s;
      z.sieg = false;
      auswahl = null;
      MOD.zeichne();
    },

    info: function () {
      var weg = z ? z.pyr.filter(function (k) { return !k; }).length : 0;
      return {
        punkte: z ? z.punkte : 0, zuege: z ? z.zuege : 0,
        extra: "Karten " + weg + "/28 · Vorrat " + (z ? z.stock.length : 0)
      };
    },

    zeichne: function () {
      if (!z) return;
      var brett = document.getElementById("brett");
      brett.innerHTML = "";
      var wrap = document.createElement("div");
      wrap.className = "pyramid";

      var zielKarten = {}, zielWaste = false;
      if (auswahl) {
        MOD.zieleFuer({ art: auswahl.typ === "waste" ? "waste" : "pyr", i: auswahl.i }).forEach(function (zn) {
          if (zn.indexOf("karte:") === 0) zielKarten[Number(zn.slice(6))] = true;
          if (zn === "waste") zielWaste = true;
        });
      }
      var feld = document.createElement("div");
      feld.className = "pyramide";
      var w = feld.clientWidth || brett.clientWidth || 560;
      var kartenBreite = Math.min(w / 7.4, 92);
      /* Auch die Hoehe muss reichen: Pyramide, 14 px Abstand und die beiden
         Stapel darunter (Karte = --sfw breit, 7/5 hoch). Sonst muss am PC in
         der Bahn gerollt werden (bei 1280 x 800 fehlten 133 px). Die 4 px
         Luft verhindern, dass ein Rundungsrest eine Bildlaufleiste ausloest. */
      var bahn = document.getElementById("bahn");
      if (bahn) {
        var ps = getComputedStyle(bahn);
        var platz = bahn.clientHeight - parseFloat(ps.paddingTop) - parseFloat(ps.paddingBottom);
        var sfw = parseFloat(getComputedStyle(document.documentElement).getPropertyValue("--sfw")) || 82;
        if (platz > 240) kartenBreite = Math.min(kartenBreite, (platz - 18 - sfw * 1.4) / 6.272);
      }
      var kartenHoehe = kartenBreite * 1.4;
      var schrittY = kartenHoehe * 0.58;
      feld.style.width = (kartenBreite * 7.4) + "px";
      feld.style.height = (kartenHoehe + 6 * schrittY) + "px";
      feld.style.margin = "0 auto";

      var idx = 0;
      for (var r = 0; r < 7; r++) {
        var anzahl = r + 1;
        var startX = (7 - anzahl) / 2 * kartenBreite;
        for (var p = 0; p < anzahl; p++, idx++) {
          var k = z.pyr[idx];
          var frei = istFrei(idx);
          var x = startX + p * kartenBreite;
          var y = r * schrittY;
          if (!k) {
            var leer = document.createElement("div");
            leer.className = "pyr-platz";
            leer.style.cssText = "left:" + x + "px;top:" + y + "px;width:" + kartenBreite + "px;height:" + kartenHoehe + "px";
            feld.appendChild(leer);
            continue;
          }
          var c = K.el(k, {
            art: "pyr", i: idx, index: idx,
            klasse: (frei ? "frei" : "verdeckt") +
              (auswahl && auswahl.typ === "pyr" && auswahl.i === idx ? " gewaehlt" : "") +
              (frei ? " zielbar" : "") +
              (zielKarten[idx] ? " moeglich" : "")
          });
          c.style.cssText = "left:" + x + "px;top:" + y + "px;width:" + kartenBreite + "px;height:" + kartenHoehe + "px";
          if (frei) c.dataset.zone = "karte:" + idx;
          feld.appendChild(c);
        }
      }

      var fuss = document.createElement("div");
      fuss.className = "pyramid-fuss";

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
      waste.className = "stapelfeld" + (z.waste.length ? "" : " leer") + (zielWaste ? " moeglich" : "");
      waste.dataset.zone = "waste";
      waste.setAttribute("aria-label", "Ablagekarte");
      if (z.waste.length) {
        var wk = z.waste[z.waste.length - 1];
        waste.appendChild(K.el(wk, {
          art: "waste", index: 0,
          klasse: auswahl && auswahl.typ === "waste" ? "gewaehlt" : ""
        }));
      } else {
        waste.innerHTML = '<span class="leerzeichen">Ablage</span>';
      }

      fuss.append(stock, waste);
      wrap.append(feld, fuss);
      brett.appendChild(wrap);
    },

    waehlen: function (was) {
      if (!z || !was) return;
      if (was.art === "waste") {
        if (!z.waste.length) return;
        if (auswahl && auswahl.typ === "pyr") {
          var pk = z.pyr[auswahl.i], wk = z.waste[z.waste.length - 1];
          if (pk && wk && pk.r + wk.r === 13) { raeumen(auswahl.i, { waste: true }); return; }
        }
        if (auswahl && auswahl.typ === "waste") { auswahl = null; TG.Shell.auswahl = null; MOD.zeichne(); return; }
        auswahl = { typ: "waste" };
        TG.Shell.auswahl = was;
        MOD.zeichne();
        TG.Shell.hinweis("Jetzt eine freie Karte mit passender Summe wählen (13 ergibt das Paar).");
        return;
      }
      if (was.art !== "pyr") return;
      var i = was.i;
      if (!z.pyr[i]) return;
      if (!istFrei(i)) {
        TG.Shell.hinweis("Diese Karte ist noch bedeckt.", true);
        return;
      }
      if (auswahl && auswahl.typ === "pyr" && auswahl.i === i) {
        auswahl = null; TG.Shell.auswahl = null; MOD.zeichne(); return;
      }
      if (!auswahl) {
        if (z.pyr[i].r === 13) { raeumen(i, null); return; }
        auswahl = { typ: "pyr", i: i };
        TG.Shell.auswahl = was;
        MOD.zeichne();
        TG.Shell.hinweis("Zweite freie Karte wählen – zusammen müssen die Werte 13 ergeben.");
        return;
      }
      var a = auswahl.typ === "waste" ? z.waste[z.waste.length - 1] : z.pyr[auswahl.i];
      var b = z.pyr[i];
      if (!a) { auswahl = null; TG.Shell.auswahl = null; MOD.zeichne(); return; }
      if (a.r + b.r === 13) {
        raeumen(i, auswahl.typ === "waste" ? { waste: true } : { pyr: auswahl.i });
        return;
      }
      TG.Shell.hinweis("Die Werte ergeben " + (a.r + b.r) + " – nötig sind 13.", true);
      if (b.r === 13) { raeumen(i, null); return; }
      auswahl = { typ: "pyr", i: i };
      TG.Shell.auswahl = { art: "pyr", i: i, index: i };
      MOD.zeichne();
    },

    /* Dieselbe Prüfung wie ablegen: freie, offene Karten, deren Wert mit der
       Auswahl 13 ergibt; aus dem Berg zusätzlich die Ablagekarte. */
    zieleFuer: function (was) {
      if (!z || !was) return [];
      var k = was.art === "waste" ? z.waste[z.waste.length - 1] :
        (was.art === "pyr" ? z.pyr[was.i] : null);
      if (!k || !k.offen) return [];
      if (was.art === "pyr" && !istFrei(was.i)) return [];
      var out = [];
      for (var i = 0; i < 28; i++) {
        if (was.art === "pyr" && i === was.i) continue;
        if (!z.pyr[i] || !z.pyr[i].offen || !istFrei(i)) continue;
        if (k.r + z.pyr[i].r === 13) out.push("karte:" + i);
      }
      if (was.art === "pyr" && z.waste.length) {
        var w = z.waste[z.waste.length - 1];
        if (w && w.offen && k.r + w.r === 13) out.push("waste");
      }
      return out;
    },

    ablegen: function (was, zone) {
      if (!z || !was) return false;
      if (zone.indexOf("karte:") === 0) {
        var zi = Number(zone.slice(6));
        if (!z.pyr[zi] || !istFrei(zi) || !z.pyr[zi].offen) return false;
        if (was.art === "waste") {
          var wk = z.waste[z.waste.length - 1];
          if (!wk) return false;
          if (wk.r + z.pyr[zi].r !== 13) return false;
          raeumen(zi, { waste: true });
          return true;
        }
        if (was.art !== "pyr" || was.i === zi) return false;
        var a = z.pyr[was.i];
        if (!a || !istFrei(was.i) || !a.offen) return false;
        if (a.r + z.pyr[zi].r !== 13) return false;
        raeumen(zi, { pyr: was.i });
        return true;
      }
      if (zone === "waste" && was.art === "pyr") {
        var b = z.pyr[was.i];
        var w = z.waste[z.waste.length - 1];
        if (!b || !w || !istFrei(was.i) || !b.offen || b.r + w.r !== 13) return false;
        raeumen(was.i, { waste: true });
        return true;
      }
      return false;
    },

    zoneKlick: function (zone) {
      if (!z) return;
      if (zone === "stock") { MOD.nachziehen(); return; }
      auswahl = null; TG.Shell.auswahl = null; MOD.zeichne();
    },

    nachziehen: function () {
      if (!z.stock.length) {
        TG.Shell.hinweis("Der Nachziehstapel ist leer.", true);
        pruefeEnde();
        return;
      }
      TG.Shell.merken();
      var k = z.stock.pop();
      k.offen = true;
      z.waste.push(k);
      z.zuege++;
      auswahl = null;
      TG.Shell.auswahl = null;
      var vorher = K.positionen();
      MOD.zeichne();
      K.bewege(vorher);
      TG.Shell.zugFertig("Neue Ablagekarte: " + K.kartenText(k) + ".");
      pruefeEnde();
    },

    tipp: function () {
      if (!z) return "";
      var frei = [];
      for (var i = 0; i < 28; i++) if (z.pyr[i] && istFrei(i)) frei.push(i);
      var wk = z.waste.length ? z.waste[z.waste.length - 1] : null;
      for (var a = 0; a < frei.length; a++) {
        if (z.pyr[frei[a]].r === 13) {
          auswahl = { typ: "pyr", i: frei[a] };
          TG.Shell.auswahl = { art: "pyr", i: frei[a], index: frei[a] };
          MOD.zeichne();
          return "Tipp: der König liegt frei und kann allein abgeräumt werden.";
        }
        if (wk && z.pyr[frei[a]].r + wk.r === 13) {
          auswahl = { typ: "waste" };
          TG.Shell.auswahl = { art: "waste", index: 0 };
          MOD.zeichne();
          return "Tipp: " + K.kartenText(z.pyr[frei[a]]) + " passt zur Ablagekarte " + K.kartenText(wk) + ".";
        }
        for (var b = a + 1; b < frei.length; b++) {
          if (z.pyr[frei[a]].r + z.pyr[frei[b]].r === 13) {
            auswahl = { typ: "pyr", i: frei[a] };
            TG.Shell.auswahl = { art: "pyr", i: frei[a], index: frei[a] };
            MOD.zeichne();
            return "Tipp: " + K.kartenText(z.pyr[frei[a]]) + " und " + K.kartenText(z.pyr[frei[b]]) + " ergeben 13.";
          }
        }
      }
      if (z.stock.length) return "Tipp: eine neue Karte nachziehen.";
      return "Kein Paar mehr möglich – nimm einen Zug zurück.";
    }
  };

  /* Reihen: 1, 2, 3 … 7 Karten; Karte r/p wird von r+1/p und r+1/p+1 bedeckt. */
  function zeileIndex(r) { return r * (r + 1) / 2; }

  function istFrei(i) {
    if (!z.pyr[i]) return false;
    var r = 0, rest = i;
    while (rest >= r + 1) { rest -= r + 1; r++; }
    if (r === 6) return true;
    var unten = zeileIndex(r + 1) + rest;
    return !z.pyr[unten] && !z.pyr[unten + 1];
  }

  function raeumen(i, partner) {
    TG.Shell.merken();
    z.pyr[i] = null;
    if (partner && partner.waste) {
      if (z.waste.length) z.waste.pop();
    } else if (partner && partner.pyr !== undefined) {
      z.pyr[partner.pyr] = null;
    }
    z.punkte += 10;
    z.zuege++;
    auswahl = null;
    TG.Shell.auswahl = null;
    var vorher = K.positionen();
    MOD.zeichne();
    K.bewege(vorher);
    TG.Shell.zugFertig();
    if (z.pyr.every(function (k) { return !k; })) {
      z.sieg = true;
      TG.Shell.sieg("Die ganze Pyramide ist abgeräumt.");
      return;
    }
    pruefeEnde();
  }

  function pruefeEnde() {
    if (!z || z.sieg) return;
    if (z.pyr.every(function (k) { return !k; })) return;
    if (z.stock.length) return;
    var frei = [];
    for (var i = 0; i < 28; i++) if (z.pyr[i] && istFrei(i)) frei.push(i);
    var wk = z.waste.length ? z.waste[z.waste.length - 1] : null;
    var moeglich = false;
    for (var a = 0; a < frei.length && !moeglich; a++) {
      if (z.pyr[frei[a]].r === 13) moeglich = true;
      if (wk && z.pyr[frei[a]].r + wk.r === 13) moeglich = true;
      for (var b = a + 1; b < frei.length && !moeglich; b++) {
        if (z.pyr[frei[a]].r + z.pyr[frei[b]].r === 13) moeglich = true;
      }
    }
    if (!moeglich) {
      TG.Shell.verloren("Der Nachziehstapel ist leer und es liegt kein Paar mit der Summe 13 mehr frei.");
    }
  }
})();
