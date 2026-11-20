/* ===========================================================================
   HALMA – Brettgeometrie (unabhängig von der Darstellung)

   Zwei Bretter:
   * klassisch – 16×16 Felder, Startbereiche sind die Eckhäuser: zu zweit
     19 Felder (Reihen 1-2-3-4-5-4), zu viert 13 Felder (1-2-3-4-3), Ziel ist
     das gegenüberliegende Eckhaus. Das 13er-Haus steckt im 19er.
   * stern     – 121 Felder im Hexgitter: Mittelhexagon mit Radius 4 (61 Felder)
     und sechs Zacken mit je 10 Feldern (4+3+2+1). Ziel ist die
     gegenüberliegende Zacke (Nummer + 3).

   Ein Brett liefert:
     felder[i] = { x, y, px, py }      px/py sind Anzeigekoordinaten
     nachbar[i] = [j, …]               Felder für einen Schritt
     sprung[i]  = [[ueber, ziel], …]   ein Sprung über ein Nachbarfeld
     zonen[k]   = [i, …]               Start-/Zielfelder von Platz k
     id(x, y)                          Feldindex oder -1
     schritt(i, dx, dy, weite)         Feld „weite Schritte“ entfernt oder -1
     richtungen                        Einheitsrichtungen des Bretts
   =========================================================================== */
(function (HB) {
  "use strict";

  var R8 = [[1, 0], [1, 1], [0, 1], [-1, 1], [-1, 0], [-1, -1], [0, -1], [1, -1]];
  var R6 = [[1, 0], [0, 1], [-1, 1], [-1, 0], [0, -1], [1, -1]];

  /* ---------------------------------------------------------- klassisch -- */
  /* anzahl = Spielerzahl: 2 gibt das große Haus (19), 4 das kleine (13).
     Ohne Angabe gilt die Zweierpartie. */
  function klassisch(anzahl) {
    var N = 16, felder = [], nachbar = [], sprung = [], i, x, y;
    var gross = (anzahl || 2) === 2;

    for (y = 0; y < N; y++) {
      for (x = 0; x < N; x++) {
        felder.push({ x: x, y: y, px: x, py: y });
      }
    }

    function id(x, y) {
      if (x < 0 || y < 0 || x >= N || y >= N) return -1;
      return y * N + x;
    }

    for (i = 0; i < felder.length; i++) {
      var n = [], s = [];
      for (var d = 0; d < R8.length; d++) {
        var dx = R8[d][0], dy = R8[d][1];
        var j = id(felder[i].x + dx, felder[i].y + dy);
        if (j >= 0) {
          n.push(j);
          var k = id(felder[i].x + 2 * dx, felder[i].y + 2 * dy);
          if (k >= 0) s.push([j, k]);
        }
      }
      nachbar.push(n);
      sprung.push(s);
    }

    /* Ein Haus, von seiner Ecke aus gezählt (a, b): Reihen 1-2-3-4, die
       nächste Reihe ist um die beiden Eckfelder gekürzt – bei 19 Feldern
       eine Reihe mehr (1-2-3-4-5-4), bei 13 die kurze (1-2-3-4-3). */
    function imHaus(a, b) {
      var s = a + b;
      if (s <= 3) return true;
      if (gross) return s === 4 || (s === 5 && a >= 1 && a <= 4);
      return s === 4 && a >= 1 && a <= 3;
    }

    var zonen = [[], [], [], []];
    for (i = 0; i < felder.length; i++) {
      x = felder[i].x; y = felder[i].y;
      if (imHaus(x, y)) zonen[0].push(i);                             // oben links
      if (imHaus(N - 1 - x, y)) zonen[1].push(i);                     // oben rechts
      if (imHaus(x, N - 1 - y)) zonen[2].push(i);                     // unten links
      if (imHaus(N - 1 - x, N - 1 - y)) zonen[3].push(i);             // unten rechts
    }

    return {
      art: "klassisch",
      n: N,
      figuren: gross ? 19 : 13,
      felder: felder,
      nachbar: nachbar,
      sprung: sprung,
      zonen: zonen,
      richtungen: R8,
      gegenseite: [3, 2, 1, 0],
      id: id,
      schritt: function (i, dx, dy, weite) {
        return id(felder[i].x + dx * weite, felder[i].y + dy * weite);
      },
      /* Schlüssel für Anzeige/Prüfsummen */
      schluessel: function (i) { return felder[i].x + ":" + felder[i].y; }
    };
  }

  /* --------------------------------------------------------------- Stern -- */
  function stern() {
    var R = 4;                       // Hexagon-Radius → 61 Mittelfelder
    var felder = [], nachbar = [], sprung = [];
    var index = {};
    var zonen = [[], [], [], [], [], []];

    function schluessel(q, r) { return q + ":" + r; }

    function add(q, r, zone) {
      var k = schluessel(q, r);
      if (index[k] !== undefined) return index[k];
      var i = felder.length;
      index[k] = i;
      felder.push({ q: q, r: r, px: q + r / 2, py: r * 0.8660254037844386 });
      if (zone >= 0) zonen[zone].push(i);
      return i;
    }

    var q, r, d, i, j;
    for (q = -R; q <= R; q++) {
      for (r = -R; r <= R; r++) {
        if (Math.max(Math.abs(q), Math.abs(r), Math.abs(q + r)) <= R) add(q, r, -1);
      }
    }
    for (d = 0; d < 6; d++) {
      var b = R6[d], u = R6[(d + 1) % 6];
      for (i = 0; i < 4; i++) {
        for (j = 0; j <= 3 - i; j++) {
          add(5 * b[0] + i * b[0] + j * u[0],
              5 * b[1] + i * b[1] + j * u[1], d);
        }
      }
    }

    for (i = 0; i < felder.length; i++) {
      var nn = [], ss = [];
      for (d = 0; d < 6; d++) {
        var nq = felder[i].q + R6[d][0], nr = felder[i].r + R6[d][1];
        var jj = index[schluessel(nq, nr)];
        if (jj !== undefined) {
          nn.push(jj);
          var zq = felder[i].q + 2 * R6[d][0], zr = felder[i].r + 2 * R6[d][1];
          var kk = index[schluessel(zq, zr)];
          if (kk !== undefined) ss.push([jj, kk]);
        }
      }
      nachbar.push(nn);
      sprung.push(ss);
    }

    return {
      art: "stern",
      n: 8,
      figuren: 10,
      felder: felder,
      nachbar: nachbar,
      sprung: sprung,
      zonen: zonen,
      richtungen: R6,
      gegenseite: [3, 4, 5, 0, 1, 2],
      id: function (q, r) {
        var k = index[schluessel(q, r)];
        return k === undefined ? -1 : k;
      },
      schritt: function (i, dx, dy, weite) {
        return this.id(felder[i].q + dx * weite, felder[i].r + dy * weite);
      },
      schluessel: function (i) { return felder[i].q + ":" + felder[i].r; }
    };
  }

  HB.Brett = {
    klassisch: klassisch,
    stern: stern,
    /* Anzeigekoordinaten auf 0…1 normieren (Seitenverhältnis bleibt erhalten). */
    rahmen: function (brett) {
      var minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;
      brett.felder.forEach(function (f) {
        if (f.px < minX) minX = f.px;
        if (f.px > maxX) maxX = f.px;
        if (f.py < minY) minY = f.py;
        if (f.py > maxY) maxY = f.py;
      });
      return { minX: minX, maxX: maxX, minY: minY, maxY: maxY,
               breite: maxX - minX + 1, hoehe: maxY - minY + 1 };
    }
  };
})(window.HB = window.HB || {});
