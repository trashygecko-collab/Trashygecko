/* ===========================================================================
   HALMA – Computergegner

   Drei Stufen, alle mit denselben Regeln wie Menschen:
   * leicht – bewertet jeden Zug grob und wählt unter den brauchbaren einen
     aus, damit Anfänger mitkommen.
   * mittel – nimmt den besten Zug nach vollständiger Bewertung (Fortschritt,
     Sprungketten, Figuren im Ziel).
   * schwer – prüft zusätzlich die beste Antwort des Gegners und meidet Züge,
     die ihm helfen (kleine Zwei-Zug-Suche).

   Der Zufall ist ein fester Generator: dieselbe Stellung ergibt dieselbe
   Wahl, damit sich Fehler nachstellen lassen.
   =========================================================================== */
(function (HB) {
  "use strict";

  function zufall(seed) {
    var a = (seed || 1) >>> 0;
    return function () {
      a = (a + 0x6D2B79F5) >>> 0;
      var t = a;
      t = Math.imul(t ^ (t >>> 15), t | 1);
      t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
      return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
    };
  }

  /* Bewertung einer Stellung aus Sicht von spieler (höher ist besser). */
  function bewerten(brett, belegung, spieler) {
    var R = HB.Regeln;
    var wert = 0;
    var drin = 0;
    for (var f = 0; f < belegung.length; f++) {
      if (belegung[f] !== spieler) continue;
      var d = R.zielAbstand(brett, f, spieler);
      wert -= d * 6;
      if (d === 0) {
        drin++;
        wert += 14;                       // im Ziel stehen ist mehr als nur nah dran
      }
    }
    wert += drin * drin * 2;              // geschlossene Reihe im Ziel zählt extra
    return wert;
  }

  function zugWert(brett, vorher, nachher, spieler, zug) {
    var w = bewerten(brett, nachher, spieler) - bewerten(brett, vorher, spieler);
    if (zug.art === "sprung") w += 4 + zug.weg.length * 2;   // Ketten lohnen
    return w;
  }

  /* Alle Züge mit Wert, absteigend sortiert. */
  function kandidaten(brett, belegung, varianteId, spieler) {
    var R = HB.Regeln;
    var zuege = R.alleZuege(brett, belegung, varianteId, spieler);
    var liste = zuege.map(function (zug) {
      var nachher = R.ziehe(belegung, zug);
      return { zug: zug, wert: zugWert(brett, belegung, nachher, spieler, zug), nachher: nachher };
    });
    liste.sort(function (a, b) { return b.wert - a.wert; });
    return liste;
  }

  function waehleZug(brett, belegung, varianteId, spieler, stufe, seed) {
    var R = HB.Regeln;
    var zuf = zufall((seed || 1) + spieler * 7919 + belegung.filter(function (v) { return v >= 0; }).length);
    var liste = kandidaten(brett, belegung, varianteId, spieler);
    if (!liste.length) return null;
    if (stufe === "leicht") {
      /* Nicht der beste Zug jedes Mal – aber nie ein schlechter. */
      var grenze = Math.max(1, Math.ceil(liste.length * 0.45));
      var i = Math.floor(zuf() * grenze);
      return liste[i].zug;
    }
    if (stufe !== "schwer") return liste[0].zug;

    /* Schwer: die besten Kandidaten gegen die beste Gegnerantwort prüfen. */
    var gegner = [];
    for (var s = 0; s < 6; s++) {
      if (s !== spieler && belegung.indexOf(s) >= 0) gegner.push(s);
    }
    var beste = null, besteWert = -Infinity;
    var enge = liste.slice(0, 10);
    for (var k = 0; k < enge.length; k++) {
      var kandidat = enge[k];
      var wert = kandidat.wert;
      if (gegner.length && kandidat.zug.art !== "sprung") {
        /* Was kann der stärkste Gegner als Antwort tun? */
        for (var g = 0; g < gegner.length; g++) {
          var antworten = kandidaten(brett, kandidat.nachher, varianteId, gegner[g]);
          if (antworten.length) {
            var besteAntwort = antworten[0];
            wert -= Math.max(0, besteAntwort.wert) * 0.35;
            break;
          }
        }
      }
      wert += zuf() * 0.4;                       // Gleichstand auflösen
      if (wert > besteWert) { besteWert = wert; beste = kandidat.zug; }
    }
    return beste || liste[0].zug;
  }

  /* Denkzeit je Stufe – reine Anzeige, die Regeln bleiben gleich. */
  var TEMPO = { leicht: 260, mittel: 420, schwer: 620 };

  HB.KI = {
    waehleZug: waehleZug,
    bewerten: bewerten,
    kandidaten: kandidaten,
    tempo: function (stufe) { return TEMPO[stufe] || 400; },
    zufall: zufall
  };
})(window.HB = window.HB || {});
