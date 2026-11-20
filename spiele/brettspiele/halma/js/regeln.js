/* ===========================================================================
   HALMA – Regelwerk (unabhängig von Darstellung, Netz und KI)

   Regelversionen (dokumentiert, Standard):
   * klassisch – 16×16, 19 Figuren je Spieler zu zweit, 13 zu viert, im
     Eckhaus; Ziel ist das gegenüberliegende Haus. Schritt auf ein
     angrenzendes freies Feld (8 Richtungen), Sprung über genau ein
     benachbartes Feld auf das freie Feld dahinter, beliebig viele Sprünge
     hintereinander mit derselben Figur. Figuren dürfen aus dem Zielbereich
     wieder herausziehen (kein Sperrfeld).
   * stern     – 121 Felder, 10 Figuren je Zacke, Ziel ist die gegenüberliegende
     Zacke, sonst dieselben Zugregeln im Hexgitter (6 Richtungen).
   * sprung    – Weitsprung-Halma auf dem 16×16-Brett: zusätzlich zu Schritt
     und Nahsprung sind weite Sprünge erlaubt. Übersprungen wird genau eine
     Figur, die auf der geraden Linie liegt; das Zielfeld liegt im selben
     Abstand hinter ihr (Spiegelpunkt). Jede Weite und alle 8 Richtungen,
     Ketten erlaubt. Das ist der Unterschied zum klassischen Halma, das nur
     über direkt benachbarte Figuren springt.

   Ein Zug ist { von, nach, weg: [übersprungene Felder], art }.
   =========================================================================== */
(function (HB) {
  "use strict";

  var VARIANTEN = {
    klassisch: {
      id: "klassisch", name: "Klassisches Halma", brett: "klassisch",
      springen: "nah", spielerzahlen: [2, 4]
    },
    stern: {
      id: "stern", name: "Sternhalma", brett: "stern",
      springen: "nah", spielerzahlen: [2, 3, 4, 6]
    },
    sprung: {
      id: "sprung", name: "Sprunghalma", brett: "klassisch",
      springen: "weit", spielerzahlen: [2, 4]
    }
  };

  /* Welche Startbereiche bekommen die Spieler? */
  var PLAETZE = {
    klassisch: { 2: [0, 3], 4: [0, 1, 2, 3] },
    stern: { 2: [0, 3], 3: [0, 2, 4], 4: [0, 1, 3, 4], 6: [0, 1, 2, 3, 4, 5] }
  };

  function plaetze(variante, anzahl) {
    var t = PLAETZE[VARIANTEN[variante].brett];
    return t[anzahl] || t[2];
  }

  function zielZone(brett, platz) {
    return brett.zonen[brett.gegenseite[platz]];
  }

  function neueBelegung(n) {
    var b = new Array(n);
    for (var i = 0; i < n; i++) b[i] = -1;
    return b;
  }

  /* Startaufstellung: Platz k bekommt die Felder seiner Zone. Die Figuren
     tragen den Platz (die Ecke) als Wert – nicht die Reihenfolge am Tisch,
     denn die kann sich ändern, wenn jemand die Partie verlässt. */
  function startAufstellung(brett, varianteId, anzahl) {
    var b = neueBelegung(brett.felder.length);
    var p = plaetze(varianteId, anzahl);
    p.forEach(function (platz) {
      brett.zonen[platz].forEach(function (i) { b[i] = platz; });
    });
    return b;
  }

  /* ------------------------------------------------------- Sprungziele -- */
  function weiteSprünge(brett, belegung, feld) {
    var out = [];
    var richtungen = brett.richtungen;
    for (var d = 0; d < richtungen.length; d++) {
      var dx = richtungen[d][0], dy = richtungen[d][1];
      for (var w = 1; w < 16; w++) {
        var ueber = brett.schritt(feld, dx, dy, w);
        if (ueber < 0) break;
        if (belegung[ueber] < 0) continue;          // übersprungen wird nur eine Figur
        var ziel = brett.schritt(feld, dx, dy, 2 * w);
        if (ziel < 0) break;
        if (belegung[ziel] < 0) out.push({ ueber: ueber, ziel: ziel });
      }
    }
    return out;
  }

  /* Alle Felder, die die Figur auf feld mit Sprüngen erreicht (BFS).
     Rückgabe: Map ziel → Weg (Liste der übersprungenen Felder). */
  function sprungFelder(brett, belegung, feld, art, artigkeit) {
    var art2 = art || "nah";
    var weg = new Map();
    var start = belegung[feld];
    var belegt = belegung.slice();
    belegt[feld] = -1;                              // die eigene Figur steht nicht mehr im Weg
    var rand = [feld];
    var gesehen = {};
    gesehen[feld] = true;

    while (rand.length) {
      var naechste = [];
      for (var r = 0; r < rand.length; r++) {
        var f = rand[r];
        var zuege;
        if (art2 === "weit") {
          zuege = weiteSprünge(brett, belegt, f);
        } else {
          /* Nahsprung: übersprungen wird genau die benachbarte Figur. */
          zuege = [];
          brett.sprung[f].forEach(function (p) {
            if (belegt[p[0]] >= 0) zuege.push({ ueber: p[0], ziel: p[1] });
          });
        }
        for (var z = 0; z < zuege.length; z++) {
          var ziel = zuege[z].ziel;
          if (belegt[ziel] >= 0 || gesehen[ziel]) continue;
          gesehen[ziel] = true;
          var vorher = weg.get(f) || [];
          weg.set(ziel, vorher.concat([zuege[z].ueber]));
          belegt[ziel] = start;                      // gedanklich dort stehen
          naechste.push(ziel);
        }
      }
      rand = naechste;
    }
    /* belegt ist nur eine Kopie – der Aufrufer merkt nichts davon. */
    return weg;
  }

  /* Alle Züge der Figur auf feld. */
  function zuegeVon(brett, belegung, varianteId, feld) {
    var v = VARIANTEN[varianteId] || VARIANTEN.klassisch;
    var spieler = belegung[feld];
    if (spieler < 0) return [];
    var zuege = [];
    var n = brett.nachbar[feld];
    for (var i = 0; i < n.length; i++) {
      if (belegung[n[i]] < 0) zuege.push({ von: feld, nach: n[i], weg: [], art: "schritt" });
    }
    sprungFelder(brett, belegung, feld, v.springen).forEach(function (weg, ziel) {
      zuege.push({ von: feld, nach: ziel, weg: weg, art: "sprung" });
    });
    return zuege;
  }

  /* Alle Züge eines Spielers (für KI und Server). */
  function alleZuege(brett, belegung, varianteId, spieler) {
    var out = [];
    for (var f = 0; f < belegung.length; f++) {
      if (belegung[f] !== spieler) continue;
      var z = zuegeVon(brett, belegung, varianteId, f);
      for (var i = 0; i < z.length; i++) out.push(z[i]);
    }
    return out;
  }

  /* Ist der Zug nach den Regeln erlaubt? (unabhängig davon, wer dran ist) */
  function zugGueltig(brett, belegung, varianteId, zug) {
    if (!zug || typeof zug.von !== "number" || typeof zug.nach !== "number") return false;
    if (zug.von < 0 || zug.von >= belegung.length) return false;
    if (zug.nach < 0 || zug.nach >= belegung.length) return false;
    if (zug.von === zug.nach) return false;
    if (belegung[zug.von] < 0 || belegung[zug.nach] >= 0) return false;
    var zuege = zuegeVon(brett, belegung, varianteId, zug.von);
    for (var i = 0; i < zuege.length; i++) {
      if (zuege[i].nach === zug.nach) return true;
    }
    return false;
  }

  function ziehe(belegung, zug) {
    var b = belegung.slice();
    b[zug.nach] = b[zug.von];
    b[zug.von] = -1;
    return b;
  }

  /* ------------------------------------------------------------- Siegen -- */
  function imZiel(brett, belegung, platz) {
    var zone = zielZone(brett, platz);
    var drin = 0;
    for (var i = 0; i < zone.length; i++) if (belegung[zone[i]] === platz) drin++;
    return drin;
  }

  function sieg(brett, belegung, platz) {
    return imZiel(brett, belegung, platz) === brett.zonen[zielIndex(brett, platz)].length;
  }

  function zielIndex(brett, platz) {
    return brett.gegenseite[platz];
  }

  /* Gitterdistanz eines Feldes zur eigenen Zielzone (ohne Blockaden). */
  function zielAbstand(brett, feld, platz) {
    var zone = zielZone(brett, platz);
    var f = brett.felder[feld];
    var best = Infinity;
    for (var i = 0; i < zone.length; i++) {
      var g = brett.felder[zone[i]];
      var d;
      if (brett.art === "stern") {
        var dq = f.q - g.q, dr = f.r - g.r;
        d = (Math.abs(dq) + Math.abs(dr) + Math.abs(dq + dr)) / 2;
      } else {
        d = Math.max(Math.abs(f.x - g.x), Math.abs(f.y - g.y));
      }
      if (d < best) best = d;
    }
    return best;
  }

  /* Wie viele Figuren stehen noch außerhalb des Ziels? */
  function fehlen(brett, belegung, platz) {
    var zone = zielZone(brett, platz);
    return zone.length - imZiel(brett, belegung, platz);
  }

  HB.Regeln = {
    VARIANTEN: VARIANTEN,
    plaetze: plaetze,
    zielZone: zielZone,
    zielIndex: zielIndex,
    neueBelegung: neueBelegung,
    startAufstellung: startAufstellung,
    sprungFelder: sprungFelder,
    weiteSprünge: weiteSprünge,
    zuegeVon: zuegeVon,
    alleZuege: alleZuege,
    zugGueltig: zugGueltig,
    ziehe: ziehe,
    imZiel: imZiel,
    sieg: sieg,
    fehlen: fehlen,
    zielAbstand: zielAbstand
  };
})(window.HB = window.HB || {});
