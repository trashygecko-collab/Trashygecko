/* ==========================================================================
   TrashyGecko – Arcade Flipper · Kern
   Reine Spiellogik ohne DOM: Tische, Physik, Punkte, Spielablauf.
   Läuft im Browser (window.FL) und in Node (module.exports).
   ========================================================================== */
(function (root) {
  "use strict";

  var W = 440, H = 820, RK = 8.5;
  var TAKT = 1 / 240;          // fester Physik-Takt
  var MAXV = 1500;             // Geschwindigkeitsdeckel
  var FLIPPER_R = 6.5;         // halbe Dicke des Flipperarms
  var GRAV = 1160;

  /* ------------------------------------------------------------- Abschusskanal */
  var KANAL = { x1: 372, x2: 420, mitte: 396, boden: 770, stift: 748, oben: 306 };

  /* --------------------------------------------------------------- Oberdeck
     Eigene Spielebene über dem Hauptfeld. Sie hat einen eigenen Rahmen, eigene
     Bumper, ein eigenes Zielband, ein Jackpot-Loch, eine Halteklammer und zwei
     Ausgänge (Mitte unten ins Hauptfeld, rechts unten in die rechte Inlane). */
  var DECK = {
    x1: 96, y1: 124, x2: 344, y2: 302,
    ein: { x: 330, y: 172 },            // hier kommt die Kugel aus dem Aufzug an
    rinne: { x1: 160, x2: 240 },        // Auslass in der Decke (Mitte unten)
    loch: { x: 306, y: 272, r: 14 },    // Seitenausgang rechts unten
    ausRinne: { x: 200, y: 330, vx: 0, vy: 140 },
    ausLoch: { x: 326, y: 632, vx: -30, vy: 200 }   // in die rechte Inlane
  };

  /* ---------------------------------------------------------------- Rahmen
     Hauptfeld: y 300 (Decke, Lücke 340–372 für den Abschuss) bis 778.
     Unten wie am echten Automaten: je eine Inlane innen (Trennstab, dann eine
     Schräge, die die Kugel auf den Flipper legt) und außen eine schmale
     Outlane an der Bande. Der Trennstab endet oben in einem Pfosten, an dem
     sich der Weg teilt – bisher war hier eine 50–60 px breite offene Rinne. */
  var LANES = {
    links:  { wand: [56, 556],  knie: [56, 652],  ende: [110, 690], sling: [92, 600, 92, 645] },
    rechts: { wand: [340, 556], knie: [340, 652], ende: [282, 690], sling: [310, 600, 310, 645] }
  };

  function laneStrecken(l) {
    return [
      [l.wand[0], l.wand[1], l.knie[0], l.knie[1]],
      [l.knie[0], l.knie[1], l.ende[0], l.ende[1]],
      [l.sling[0], l.sling[1], l.sling[2], l.sling[3]]
    ];
  }

  var RAHMEN = [
    [30, 300, 340, 300],          // Decke des Hauptfelds
    [30, 300, 30, 772],           // linke Bande
    [372, 306, 372, 806],         // Kanalwand
    [420, 120, 420, 806],         // rechte Außenwand
    [372, 806, 420, 806]          // Kanalboden
  ].concat(
    laneStrecken(LANES.links),    // Inlane/Outlane links
    laneStrecken(LANES.rechts)    // Inlane/Outlane rechts
  );

  /* Der Weg aus dem Kanal ins Feld: oben rechts hinein und nach links weg. */
  var AUSLAUF = {
    name: "Auslauf", tempo: 900, punkte: 0,
    ein: { x: 396, y: 306, r: 30 },
    bahn: [{ x: 396, y: 304 }, { x: 366, y: 300 }, { x: 336, y: 304 }, { x: 312, y: 316 }],
    aus: { x: 300, y: 322, vx: -330, vy: 50 }
  };

  var SLING_L = { x1: 92, y1: 600, x2: 158, y2: 655, nx: 0.7071, ny: -0.7071 };
  var SLING_R = { x1: 310, y1: 600, x2: 244, y2: 655, nx: -0.7071, ny: -0.7071 };

  /* --------------------------------------------------------------- Tische
     Gemeinsame Architektur, je Tisch eigene Bumper, Bänke, Rampen und Werte. */
  var TISCHE = [
    {
      id: "zwerg", name: "Zwergenwerk", satz: 1, grav: GRAV, bonus: 3000,
      tipp: "Stollenwerk mit sechs Bumpern, Aufzug zum Oberdeck und Schloss links.",
      bumper: [[152, 372, 19, 130, 280], [252, 372, 19, 130, 280], [202, 442, 20, 150, 285],
               [124, 478, 16, 170, 270], [280, 478, 16, 170, 270]],
      deckBumper: [[254, 176, 14], [198, 212, 15], [142, 176, 14]],
      deckZiele: [[196, 268, 20, 9], [220, 268, 20, 9], [244, 268, 20, 9]],
      zieleL: [[31, 380], [31, 412], [31, 444]],
      zieleR: [[353, 380], [353, 412]],
      fallen: [[124, 494], [152, 494], [180, 494], [208, 494]],
      spinner: { x: 202, y: 536, len: 28 },
      scoop: { x: 202, y: 600, r: 14 },
      schloss: { x: 324, y: 508, r: 13 },
      magnet: { x: 202, y: 688, r: 20 },
      rampen: [
        { name: "Aufzug", punkte: 1200, tempo: 620, ziel: "deck",
          ein: { x: 330, y: 466, r: 14 },
          bahn: [[330, 466], [356, 430], [360, 350], [350, 300], [344, 240], [336, 196]],
          aus: { x: DECK.ein.x, y: DECK.ein.y, vx: -280, vy: 20 } },
        { name: "Schleife", punkte: 900, tempo: 600, ziel: "weiche",
          ein: { x: 96, y: 466, r: 14 },
          bahn: [[96, 466], [70, 430], [62, 350], [72, 300], [92, 320]],
          aus: { x: 78, y: 344, vx: 210, vy: 150 },
          aus2: { x: 326, y: 632, vx: -30, vy: 210 } }
      ]
    },
    {
      id: "neon", name: "Neon-Stadt", satz: 2, grav: GRAV + 40, bonus: 3500,
      tipp: "Sechs Bumper im Rautenmuster, schnelle Schleife, breites Zielband.",
      bumper: [[150, 360, 17, 140, 290], [252, 360, 17, 140, 290], [201, 352, 18, 140, 290],
               [176, 438, 16, 150, 280], [226, 438, 16, 150, 280], [201, 494, 15, 160, 285]],
      deckBumper: [[258, 180, 13], [200, 164, 15], [142, 180, 13]],
      deckZiele: [[196, 268, 20, 9], [220, 268, 20, 9], [244, 268, 20, 9]],
      zieleL: [[31, 376], [31, 408], [31, 440]],
      zieleR: [[353, 376], [353, 408]],
      fallen: [[118, 496], [146, 496], [174, 496], [202, 496]],
      spinner: { x: 202, y: 540, len: 30 },
      scoop: { x: 202, y: 606, r: 14 },
      schloss: { x: 326, y: 514, r: 13 },
      magnet: { x: 202, y: 690, r: 20 },
      rampen: [
        { name: "Aufzug", punkte: 1300, tempo: 660, ziel: "deck",
          ein: { x: 330, y: 462, r: 14 },
          bahn: [[330, 462], [358, 424], [362, 344], [352, 296], [344, 236], [336, 194]],
          aus: { x: DECK.ein.x, y: DECK.ein.y, vx: -300, vy: 10 } },
        { name: "Datenloop", punkte: 1000, tempo: 640, ziel: "weiche",
          ein: { x: 96, y: 462, r: 14 },
          bahn: [[96, 462], [68, 424], [60, 344], [70, 296], [92, 318]],
          aus: { x: 78, y: 342, vx: 220, vy: 160 },
          aus2: { x: 336, y: 640, vx: -40, vy: 260 } }
      ]
    },
    {
      id: "pirat", name: "Piratenbucht", satz: 3, grav: GRAV - 30, bonus: 3800,
      tipp: "Fünf Bumper, zwei Zielbänke, Planke oben rechts, Schloss am Heck.",
      bumper: [[152, 366, 18, 130, 285], [250, 366, 18, 130, 285],
               [112, 438, 15, 170, 275], [292, 438, 15, 170, 275], [201, 476, 19, 150, 280]],
      deckBumper: [[250, 172, 14], [196, 210, 15], [144, 172, 14]],
      deckZiele: [[200, 266, 20, 9], [224, 266, 20, 9], [248, 266, 20, 9]],
      zieleL: [[31, 386], [31, 418], [31, 450]],
      zieleR: [[353, 386], [353, 418]],
      fallen: [[126, 498], [154, 498], [182, 498], [210, 498]],
      spinner: { x: 202, y: 538, len: 29 },
      scoop: { x: 120, y: 600, r: 14 },
      schloss: { x: 316, y: 512, r: 13 },
      magnet: { x: 202, y: 688, r: 20 },
      rampen: [
        { name: "Planke", punkte: 1100, tempo: 630, ziel: "deck",
          ein: { x: 330, y: 458, r: 14 },
          bahn: [[330, 458], [358, 420], [362, 340], [352, 292], [344, 232], [336, 190]],
          aus: { x: DECK.ein.x, y: DECK.ein.y, vx: -290, vy: 15 } },
        { name: "Kombüse", punkte: 950, tempo: 610, ziel: "weiche",
          ein: { x: 96, y: 458, r: 14 },
          bahn: [[96, 458], [68, 420], [60, 340], [70, 292], [92, 314]],
          aus: { x: 78, y: 340, vx: 215, vy: 155 },
          aus2: { x: 326, y: 632, vx: -30, vy: 205 } }
      ]
    },
    {
      id: "stern", name: "Sternenstaub", satz: 4, grav: GRAV - 60, bonus: 4500,
      tipp: "Meisterklasse: große Bumper, sieben Ziele, hohe Werte, lange Wege.",
      bumper: [[150, 368, 21, 180, 310], [252, 368, 21, 180, 310], [201, 440, 18, 200, 300],
               [118, 484, 15, 210, 295], [284, 484, 15, 210, 295], [201, 498, 14, 220, 300]],
      deckBumper: [[252, 174, 15], [198, 212, 16], [144, 174, 15]],
      deckZiele: [[194, 268, 20, 9], [218, 268, 20, 9], [242, 268, 20, 9]],
      zieleL: [[31, 374], [31, 406], [31, 438]],
      zieleR: [[353, 374], [353, 406]],
      fallen: [[122, 500], [150, 500], [178, 500], [206, 500]],
      spinner: { x: 202, y: 542, len: 32 },
      scoop: { x: 202, y: 610, r: 15 },
      schloss: { x: 322, y: 510, r: 13 },
      magnet: { x: 202, y: 692, r: 21 },
      rampen: [
        { name: "Sternenbahn", punkte: 1500, tempo: 680, ziel: "deck",
          ein: { x: 330, y: 470, r: 14 },
          bahn: [[330, 470], [360, 430], [364, 348], [354, 298], [346, 238], [336, 194]],
          aus: { x: DECK.ein.x, y: DECK.ein.y, vx: -310, vy: 5 } },
        { name: "Kometenschleife", punkte: 1200, tempo: 660, ziel: "weiche",
          ein: { x: 96, y: 470, r: 14 },
          bahn: [[96, 470], [66, 430], [58, 348], [68, 298], [90, 320]],
          aus: { x: 76, y: 344, vx: 225, vy: 165 },
          aus2: { x: 326, y: 632, vx: -30, vy: 215 } }
      ]
    }
  ];

  /* ------------------------------------------------------------ Geometrie */
  var zufallStand = 20261120;
  function zufall() {
    zufallStand = (zufallStand * 1664525 + 1013904223) % 4294967296;
    return zufallStand / 4294967296;
  }
  function zufallSetzen(s) { zufallStand = s >>> 0; }

  function naechsterPunkt(px, py, ax, ay, bx, by) {
    var dx = bx - ax, dy = by - ay;
    var l = dx * dx + dy * dy;
    var t = l > 0 ? ((px - ax) * dx + (py - ay) * dy) / l : 0;
    if (t < 0) t = 0; else if (t > 1) t = 1;
    return { x: ax + t * dx, y: ay + t * dy };
  }
  function abstandSegment(px, py, ax, ay, bx, by) {
    var c = naechsterPunkt(px, py, ax, ay, bx, by);
    return Math.hypot(px - c.x, py - c.y);
  }

  /* --------------------------------------------------------------- Aufbau */
  function strecke(s) { return { x1: s[0], y1: s[1], x2: s[2], y2: s[3] }; }

  /* Deckrahmen: drei Wände plus Boden mit Auslass in der Mitte. */
  var DECK_WAENDE = [
    [96, 302, 96, 124], [96, 124, 140, 104], [140, 104, 300, 104],
    [300, 104, 344, 124], [344, 124, 344, 302],
    [96, 302, 160, 302], [240, 302, 344, 302]
  ].map(strecke);

  var DECK_RAHMEN = [
    [96, 110, 96, 300], [96, 108, 344, 108], [344, 110, 344, 300],
    [96, 300, 160, 300], [240, 300, 344, 300]
  ].map(strecke);

  function tischKopie(def) {
    var t = {
      id: def.id, name: def.name, satz: def.satz, grav: def.grav, tipp: def.tipp, bonus: def.bonus,
      waende: RAHMEN.concat(def.waende || []).map(strecke),
      eigene: (def.waende || []).map(strecke),
      bumper: def.bumper.map(function (b) {
        return { x: b[0], y: b[1], r: b[2], punkte: b[3], kick: b[4], blitz: 0, cooldown: 0 };
      }),
      schleudern: [
        { x1: SLING_L.x1, y1: SLING_L.y1, x2: SLING_L.x2, y2: SLING_L.y2, nx: SLING_L.nx, ny: SLING_L.ny, punkte: 80, kick: 320, cooldown: 0, blitz: 0 },
        { x1: SLING_R.x1, y1: SLING_R.y1, x2: SLING_R.x2, y2: SLING_R.y2, nx: SLING_R.nx, ny: SLING_R.ny, punkte: 80, kick: 320, cooldown: 0, blitz: 0 }
      ],
      zieleOben: def.zieleL.map(function (p) { return { x: p[0], y: p[1], w: 19, h: 26, punkte: 250, runter: false, blitz: 0, sperre: 0 }; })
        .concat(def.zieleR.map(function (p) { return { x: p[0], y: p[1], w: 19, h: 26, punkte: 300, runter: false, blitz: 0, sperre: 0 }; })),
      fallen: def.fallen.map(function (p) {
        return { x: p[0], y: p[1], w: 24, h: 11, punkte: 320, runter: false, blitz: 0, wecker: 0, sperre: 0 };
      }),
      spinner: { x: def.spinner.x, y: def.spinner.y, len: def.spinner.len, punkte: 110 },
      scoop: { x: def.scoop.x, y: def.scoop.y, r: def.scoop.r, punkte: 1200, sperre: 0, blitz: 0 },
      schloss: { x: def.schloss.x, y: def.schloss.y, r: def.schloss.r, sperre: 0, blitz: 0 },
      magnet: { x: def.magnet.x, y: def.magnet.y, r: def.magnet.r, an: 0 },
      rollover: [
        { x: 152, y: 314, w: 30, h: 24, an: false, blitz: 0 },
        { x: 202, y: 310, w: 30, h: 26, an: false, blitz: 0 },
        { x: 252, y: 314, w: 30, h: 24, an: false, blitz: 0 }
      ],
      rampen: def.rampen.map(function (r) {
        return {
          name: r.name, punkte: r.punkte, tempo: r.tempo, ziel: r.ziel,
          ein: { x: r.ein.x, y: r.ein.y, r: r.ein.r },
          bahn: r.bahn.map(function (p) { return { x: p[0], y: p[1] }; }),
          aus: { x: r.aus.x, y: r.aus.y, vx: r.aus.vx, vy: r.aus.vy },
          aus2: r.aus2 ? { x: r.aus2.x, y: r.aus2.y, vx: r.aus2.vx, vy: r.aus2.vy } : null,
          treffer: 0, sperre: 0, blitz: 0, laenge: 0
        };
      }),
      deck: {
        waende: DECK_RAHMEN.slice(),
        bumper: def.deckBumper.map(function (b) {
          return { x: b[0], y: b[1], r: b[2], punkte: 400, kick: 230, blitz: 0, cooldown: 0 };
        }),
        ziele: def.deckZiele.map(function (p) {
          return { x: p[0], y: p[1], w: p[2], h: p[3], punkte: 600, runter: false, blitz: 0, wecker: 0, sperre: 0 };
        }),
        gleiter: { x: 210, y: 238, w: 34, h: 10, x0: 130, x1: 288, phase: 0, richtung: 1, punkte: 700, blitz: 0, sperre: 0 },
        loch: { x: DECK.loch.x, y: DECK.loch.y, r: DECK.loch.r, sperre: 0, blitz: 0 },
        klammer: { x: 104, y: 264, r: 15, sperre: 0, haelt: 0, blitz: 0 }
      }
    };
    t.rampen.forEach(function (r) {
      var l = 0;
      for (var i = 1; i < r.bahn.length; i++) l += Math.hypot(r.bahn[i].x - r.bahn[i - 1].x, r.bahn[i].y - r.bahn[i - 1].y);
      r.laenge = l;
    });
    return t;
  }

  function neueKugel(x, y, vx, vy, ebene) {
    return {
      x: x, y: y, vx: vx || 0, vy: vy || 0,
      ebene: ebene || "haupt", zustand: "feld",
      bahn: null, sperre: 0, saver: 0, ruhig: 0,
      anker: null, still: 0, stups: 0
    };
  }

  function neuesSpiel(tischNr, kugeln) {
    var def = TISCHE[tischNr] || TISCHE[0];
    var t = tischKopie(def);
    var sp = {
      tischNr: TISCHE.indexOf(def), tisch: t,
      kugelnGesamt: 3, kugelnUebrig: 3,
      kugeln: [],                    // alle Kugeln im Spiel (Feld, Bahn, Schloss)
      zustand: "start",              // start | bereit | spiel | vorbei
      punkte: 0, multi: 1, zeit: 0, verloren: 0, zuege: 0,
      flipper: [
        { x: 108, y: 710, len: 81, ruhe: 0.52, aktiv: -0.56, winkel: 0.52, an: false, links: true, omega: 0 },
        { x: 284, y: 710, len: 81, ruhe: Math.PI - 0.52, aktiv: Math.PI + 0.56, winkel: Math.PI - 0.52, an: false, links: false, omega: 0 }
      ],
      spinnerWinkel: 0, spinnerDreh: 0,
      abschuss: { laedt: false, kraft: 0 },
      weiche: 0,                    // 0 = Auslauf oben, 1 = rechte Inlane
      bonusZeit: 0, saverZeit: 0,
      schloss: [],                  // gefangene Kugeln {x,y}
      multiball: 0,
      mission: { folge: [0, 2, 4], stand: 0, fertig: 0 },
      zieleRunter: 0,
      kickback: 1,                  // 1 = geladen
      extraBall: false,
      jackpot: false, schlossOffen: false,
      popups: [], ereignisse: [], meldung: "",
      auslaufSperre: 0, rampenSperre: 0, klemme: null, ruhig: 0
    };
    if (typeof kugeln === "number") { sp.kugelnGesamt = kugeln; sp.kugelnUebrig = kugeln; }
    return sp;
  }

  /* -------------------------------------------------------------- Punkte */
  function multiWert(sp) {
    var m = sp.multi + (sp.bonusZeit > 0 ? 1 : 0);
    return Math.min(10, m);
  }
  function addiere(sp, wert, x, y, art) {
    var echt = Math.round(wert * multiWert(sp));
    sp.punkte += echt;
    sp.popups.push({ x: x, y: y, text: "+" + echt, leben: 1 });
    if (sp.popups.length > 18) sp.popups.shift();
    sp.ereignisse.push({ art: art || "punkte", x: x, y: y, wert: echt });
    return echt;
  }

  /* ------------------------------------------------------------- Kollision */
  function stoss(k, c, nx, ny, e, reibung, extra) {
    k.x = c.x + nx * (RK + 0.35);
    k.y = c.y + ny * (RK + 0.35);
    var vn = k.vx * nx + k.vy * ny;
    if (vn < 0) {
      var j = (1 + e) * vn;
      k.vx -= j * nx;
      k.vy -= j * ny;
      if (reibung) {
        var tx = -ny, ty = nx;
        var vt = k.vx * tx + k.vy * ty;
        k.vx -= vt * reibung * tx;
        k.vy -= vt * reibung * ty;
      }
    }
    if (extra) { k.vx += nx * extra; k.vy += ny * extra; }
  }

  function wand(k, a, e) {
    var d = abstandSegment(k.x, k.y, a.x1, a.y1, a.x2, a.y2);
    if (d >= RK) return false;
    var c = naechsterPunkt(k.x, k.y, a.x1, a.y1, a.x2, a.y2);
    var dx = k.x - c.x, dy = k.y - c.y;
    if (d < 0.0001) { dx = 0; dy = -1; }
    var l = Math.hypot(dx, dy) || 1;
    stoss(k, c, dx / l, dy / l, e, 0.03, 0);
    return true;
  }

  function kreis(sp, k, o, r, punkte, kick, art) {
    var dx = k.x - o.x, dy = k.y - o.y;
    var d = Math.hypot(dx, dy);
    if (d >= r + RK) return false;
    if (d < 0.0001) { dx = 0; dy = -1; d = 1; }
    var nx = dx / d, ny = dy / d;
    stoss(k, { x: o.x + nx * r, y: o.y + ny * r }, nx, ny, 0.55, 0.02, 0);
    if (kick) { k.vx += nx * kick; k.vy += ny * kick; }
    if (o.cooldown !== undefined) {
      if (o.cooldown > 0) return true;
      o.cooldown = 0.11;
      if (o.blitz !== undefined) o.blitz = 1;
    }
    if (punkte) addiere(sp, punkte, o.x, o.y - r * 0.4, art || "bumper");
    return true;
  }

  function rechteck(sp, k, z, punkte, art) {
    if (z.runter) return false;
    var c = naechsterPunkt(k.x, k.y, z.x, z.y, z.x + z.w, z.y + z.h);
    var dx = k.x - c.x, dy = k.y - c.y;
    var d = Math.hypot(dx, dy);
    if (d >= RK) return false;
    if (d < 0.0001) { dx = 0; dy = -1; d = 1; }
    stoss(k, c, dx / d, dy / d, 0.45, 0.02, 0);
    if (z.sperre > 0) return false;
    if (z.sperre !== undefined) z.sperre = 0.45;
    if (punkte) addiere(sp, punkte, z.x + z.w / 2, z.y - 4, art || "ziel");
    return true;
  }

  function flipperStoss(sp, k, f, e) {
    var ex = f.x + Math.cos(f.winkel) * f.len;
    var ey = f.y + Math.sin(f.winkel) * f.len;
    var c = naechsterPunkt(k.x, k.y, f.x, f.y, ex, ey);
    var dx = k.x - c.x, dy = k.y - c.y;
    var d = Math.hypot(dx, dy);
    var dick = RK + FLIPPER_R;
    if (d >= dick) return false;
    if (d < 0.0001) { dx = 0; dy = -1; d = 1; }
    var nx = dx / d, ny = dy / d;
    k.x = c.x + nx * (dick + 0.3);
    k.y = c.y + ny * (dick + 0.3);
    var rx = c.x - f.x, ry = c.y - f.y;
    var vpx = -f.omega * ry, vpy = f.omega * rx;
    var rvx = k.vx - vpx, rvy = k.vy - vpy;
    var vn = rvx * nx + rvy * ny;
    if (vn < 0) {
      var j = (1 + e) * vn;
      k.vx = vpx + (rvx - j * nx);
      k.vy = vpy + (rvy - j * ny);
    }
    if (f.an && f.omega > 2) {
      var vz = k.vx * nx + k.vy * ny;
      var soll = Math.max(0, vpx * nx + vpy * ny) * 0.85;
      if (vz < soll) { k.vx += (soll - vz) * nx; k.vy += (soll - vz) * ny; }
    }
    return true;
  }

  /* --------------------------------------------------------------- Bahnen */
  function bahnPunkt(bahn, s) {
    var rest = s, letzte = bahn[0];
    for (var i = 1; i < bahn.length; i++) {
      var p = bahn[i - 1], q = bahn[i];
      var l = Math.hypot(q.x - p.x, q.y - p.y);
      if (rest <= l || i === bahn.length - 1) {
        var t = l > 0 ? Math.min(1, rest / l) : 0;
        return { x: p.x + (q.x - p.x) * t, y: p.y + (q.y - p.y) * t };
      }
      rest -= l; letzte = q;
    }
    return letzte;
  }

  function bahnStarten(sp, k, r) {
    var weg = r.laenge || 0;
    if (!weg) {
      for (var i = 1; i < r.bahn.length; i++) weg += Math.hypot(r.bahn[i].x - r.bahn[i - 1].x, r.bahn[i].y - r.bahn[i - 1].y);
    }
    k.bahn = { rampe: r, s: 0, weg: weg, tempo: r.tempo };
    k.zustand = "bahn";
    k.x = r.bahn[0].x; k.y = r.bahn[0].y; k.vx = 0; k.vy = 0;
  }

  function rampeStarten(sp, k, r) {
    bahnStarten(sp, k, r);
    r.treffer++;
    r.blitz = 1;
    var bonus = r.punkte + (r.treffer - 1) * 200;
    var p = addiere(sp, bonus, r.ein.x, r.ein.y - 14, "rampe");
    sp.meldung = r.name + " +" + p;
    if (r.ziel === "deck") sp.ereignisse.push({ art: "aufzug", x: r.ein.x, y: r.ein.y, wert: 0 });
    if (r.ziel === "weiche") {
      sp.weiche = 1 - sp.weiche;
      sp.ereignisse.push({ art: "weiche", x: 200, y: 470, wert: sp.weiche });
    }
  }

  function bahnSchritt(sp, k, dt) {
    var b = k.bahn;
    b.s += b.tempo * dt;
    var stelle = bahnPunkt(b.rampe.bahn, b.s);
    k.x = stelle.x; k.y = stelle.y;
    if (b.s < b.weg) return;
    var r = b.rampe, a = r.aus;
    if (r.ziel === "weiche" && sp.weiche === 1 && r.aus2) a = r.aus2;
    k.x = a.x; k.y = a.y; k.vx = a.vx; k.vy = a.vy;
    k.bahn = null; k.zustand = "feld";
    k.ebene = (a.y < DECK.y2 - 4) ? "deck" : "haupt";
    k.sperre = 0.9;
    sp.rampenSperre = 0.7;
    addiere(sp, 200, a.x, a.y - 10, "bahnende");
  }

  /* -------------------------------------------------------------- Abschuss */
  function abschussLaden(sp, an) {
    if (sp.zustand !== "bereit") return;
    sp.abschuss.laedt = !!an;
    if (!an) sp.abschuss.kraft = 0;
  }

  /* Erste Kugel der Partie bereitstellen (verbraucht einen der drei Bälle). */
  function partieStarten(sp) {
    sp.zustand = "bereit";
    if (sp.kugelnUebrig > 0) sp.kugelnUebrig--;
    sp.abschuss.kraft = 0;
    sp.meldung = "Kugel " + (sp.kugelnGesamt - sp.kugelnUebrig) + " von " + sp.kugelnGesamt + ".";
    return sp;
  }

  function abschuss(sp, kraft) {
    if (sp.zustand !== "bereit") return false;
    var kk = Math.max(0.28, Math.min(1, kraft === undefined ? sp.abschuss.kraft : kraft));
    var k = neueKugel(KANAL.mitte, KANAL.stift, (zufall() - 0.5) * 16, -(980 + 470 * kk), "haupt");
    k.sperre = 0.2;
    sp.kugeln.push(k);
    sp.zustand = "spiel";
    sp.abschuss.laedt = false;
    sp.abschuss.kraft = 0;
    sp.zuege++;
    sp.ereignisse.push({ art: "abschuss", x: KANAL.mitte, y: KANAL.stift, wert: kk });
    return true;
  }

  function fluegel(sp, links, an) {
    var f = sp.flipper[links ? 0 : 1];
    if (!f) return;
    f.an = !!an;
  }

  function flipperDrehen(sp, dt) {
    for (var i = 0; i < sp.flipper.length; i++) {
      var f = sp.flipper[i];
      var ziel = f.an ? f.aktiv : f.ruhe;
      var d = ziel - f.winkel;
      while (d > Math.PI) d -= Math.PI * 2;
      while (d < -Math.PI) d += Math.PI * 2;
      var max = 22 * dt;
      var neu = f.winkel + Math.max(-max, Math.min(max, d));
      f.omega = (neu - f.winkel) / dt;
      f.winkel = neu;
    }
  }

  /* ---------------------------------------------------- Spielelemente (Haupt) */
  function schleudern(sp, k, dt) {
    for (var i = 0; i < sp.tisch.schleudern.length; i++) {
      var s = sp.tisch.schleudern[i];
      if (s.cooldown > 0) s.cooldown -= dt;
      if (s.blitz > 0) s.blitz -= dt * 3;
      var d = abstandSegment(k.x, k.y, s.x1, s.y1, s.x2, s.y2);
      if (d >= RK + 4) continue;
      var l2 = (s.x2 - s.x1) * (s.x2 - s.x1) + (s.y2 - s.y1) * (s.y2 - s.y1);
      var t = l2 > 0 ? ((k.x - s.x1) * (s.x2 - s.x1) + (k.y - s.y1) * (s.y2 - s.y1)) / l2 : 0.5;
      if (t < 0.1 || t > 0.9) continue;
      var c = naechsterPunkt(k.x, k.y, s.x1, s.y1, s.x2, s.y2);
      var dx = k.x - c.x, dy = k.y - c.y, dl = Math.hypot(dx, dy);
      var nx = dl > 0.001 ? dx / dl : s.nx, ny = dl > 0.001 ? dy / dl : s.ny;
      if (Math.hypot(c.x - s.x1, c.y - s.y1) < 1.2 || Math.hypot(c.x - s.x2, c.y - s.y2) < 1.2) { nx = s.nx; ny = s.ny; }
      stoss(k, c, nx, ny, 0.45, 0, 0);
      if (nx * s.nx + ny * s.ny < 0.2) { nx = s.nx; ny = s.ny; }
      if (s.cooldown <= 0) {
        s.cooldown = 0.13; s.blitz = 1;
        k.vx = nx * (Math.abs(k.vx) * 0.25 + s.kick);
        k.vy = ny * (Math.abs(k.vy) * 0.25 + s.kick);
        addiere(sp, s.punkte, c.x, c.y - 6, "schleuder");
      }
    }
  }

  function spinner(sp, k, dt) {
    var spn = sp.tisch.spinner;
    sp.spinnerWinkel += dt * (2.4 + Math.min(10, Math.abs(k.vx) * 0.01 + Math.abs(k.vy) * 0.01));
    if (sp.spinnerDreh > 0) { sp.spinnerDreh -= dt; return; }
    var sx1 = spn.x - Math.cos(sp.spinnerWinkel) * spn.len;
    var sy1 = spn.y - Math.sin(sp.spinnerWinkel) * spn.len;
    var sx2 = spn.x + Math.cos(sp.spinnerWinkel) * spn.len;
    var sy2 = spn.y + Math.sin(sp.spinnerWinkel) * spn.len;
    if (abstandSegment(k.x, k.y, sx1, sy1, sx2, sy2) >= RK + 3) return;
    var sc = naechsterPunkt(k.x, k.y, sx1, sy1, sx2, sy2);
    var sdx = k.x - sc.x, sdy = k.y - sc.y, sd = Math.hypot(sdx, sdy);
    var snx = sd > 0.001 ? sdx / sd : 0, sny = sd > 0.001 ? sdy / sd : -1;
    stoss(k, sc, snx, sny, 0.5, 0, 0);
    k.vx += -Math.sin(sp.spinnerWinkel) * 150;
    k.vy += Math.cos(sp.spinnerWinkel) * 150 - 40;
    k.vx += (k.x < spn.x ? -1 : 1) * 220;
    k.vy -= 120;
    sp.spinnerDreh = 0.26;
    addiere(sp, spn.punkte, spn.x, spn.y - 18, "spinner");
  }

  function kugelfang(sp, k, dt) {
    var so = sp.tisch.scoop;
    if (so.sperre > 0) return;
    if (Math.hypot(k.x - so.x, k.y - so.y) < so.r + RK - 2 && k.vy > 20) {
      so.sperre = 1.4; so.blitz = 1;
      /* Auswurf schräg nach rechts oben – so fliegt die Kugel nicht in die
         Fallziele oder den Spinner über dem Loch. */
      k.x = so.x + 24; k.y = so.y - so.r - RK - 2;
      k.vx = 560; k.vy = -700;
      var p = addiere(sp, so.punkte, so.x, so.y - 26, "scoop");
      sp.meldung = "Kugelfang +" + p;
    }
  }

  function schlossLoch(sp, k, dt) {
    var sc = sp.tisch.schloss;
    if (sc.sperre > 0) return;
    if (Math.hypot(k.x - sc.x, k.y - sc.y) >= sc.r + RK - 2 || k.vy < 10) return;
    sc.sperre = 1.2; sc.blitz = 1;
    if (sp.schlossOffen && sp.schloss.length < 2) {
      k.zustand = "schloss";
      k.vx = 0; k.vy = 0;
      sp.schloss.push(k);
      sp.kugeln = sp.kugeln.filter(function (x) { return x !== k; });
      sp.schlossOffen = false;
      addiere(sp, 2000, sc.x, sc.y - 20, "schloss");
      sp.meldung = "Kugel eingeschlossen (" + sp.schloss.length + "/2)";
      sp.ereignisse.push({ art: "schloss", x: sc.x, y: sc.y, wert: sp.schloss.length });
      if (sp.schloss.length >= 2) multiballStarten(sp, "schloss");
      /* Letzte Kugel im Spiel eingesperrt: nächste kommt auf den Stift. */
      else if (imSpiel(sp) === 0) nachDrain(sp);
      return;
    }
    k.x = sc.x; k.y = sc.y - sc.r - RK - 4;
    k.vx = (zufall() - 0.5) * 220; k.vy = -900;
    addiere(sp, 700, sc.x, sc.y - 24, "loch");
  }

  function zielGetroffen(sp, i, x, y) {
    var m = sp.mission;
    if (m.folge[m.stand] === i) {
      m.stand++;
      addiere(sp, 300, x, y - 10, "mission");
      if (m.stand >= m.folge.length) {
        m.stand = 0;
        m.fertig++;
        m.folge = m.folge.map(function (v) { return (v + 1) % sp.tisch.zieleOben.length; });
        sp.schlossOffen = true;
        addiere(sp, 2500, 200, 360, "mission");
        sp.meldung = "Mission " + m.fertig + " erfüllt – Schloss offen!";
        sp.ereignisse.push({ art: "mission", x: 200, y: 380, wert: m.fertig });
        if (m.fertig === 3) { sp.extraBall = true; sp.ereignisse.push({ art: "extra", x: 200, y: 400, wert: 1 }); }
        if (m.fertig >= 2) { sp.jackpot = true; sp.ereignisse.push({ art: "licht", x: 200, y: 420, wert: 2 }); }
      }
    }
  }

  function bonusPhase(sp) {
    if (sp.bonusZeit > 0) return;
    sp.bonusZeit = 15;
    sp.meldung = "BONUSPHASE – alle Punkte ×2 (15 s)";
    sp.ereignisse.push({ art: "bonusphase", x: 200, y: 430, wert: 15 });
  }

  function multiballStarten(sp, grund) {
    var los = sp.schloss.slice();
    sp.schloss = [];
    sp.multiball++;
    los.forEach(function (s, i) {
      s.zustand = "feld";
      s.ebene = "haupt";
      s.x = 200 + (i ? 60 : -60);
      s.y = 360;
      s.vx = (i ? 220 : -220);
      s.vy = 180;
      s.saver = 0.6;
    });
    sp.kugeln = sp.kugeln.concat(los);
    if (grund !== "extra") {
      var k = neueKugel(DECK.ein.x, DECK.ein.y, -280, 10, "deck");
      k.saver = 0.6;
      sp.kugeln.push(k);
    }
    sp.meldung = "MULTIBALL! " + sp.kugeln.length + " Kugeln im Spiel";
    sp.ereignisse.push({ art: "multiball", x: 200, y: 380, wert: sp.kugeln.length });
    sp.zustand = "spiel";
  }

  /* ------------------------------------------------------ Spielelemente Deck */
  function deckSchritt(sp, k, dt) {
    var d = sp.tisch.deck, i;
    for (i = 0; i < d.bumper.length; i++) {
      var b = d.bumper[i];
      if (b.cooldown > 0) b.cooldown -= dt;
      if (b.blitz > 0) b.blitz -= dt * 3;
      kreis(sp, k, b, b.r, b.punkte, b.kick, "bumper");
    }
    for (i = 0; i < d.ziele.length; i++) {
      var z = d.ziele[i];
      if (z.blitz > 0) z.blitz -= dt * 3;
      if (z.wecker > 0) {
        z.wecker -= dt;
        if (z.wecker <= 0) z.runter = false;
        continue;
      }
      if (rechteck(sp, k, z, z.punkte, "ziel")) {
        z.runter = true; z.blitz = 1; z.wecker = 9;
        addiere(sp, 800, 200, 250, "deckbank");
        sp.ereignisse.push({ art: "ziel", x: z.x, y: z.y, wert: 0 });
      }
    }
    /* Gleiter: fährt langsam hin und her, ein bewegliches Ziel */
    var g = d.gleiter;
    if (g.blitz > 0) g.blitz -= dt * 3;
    if (rechteck(sp, k, g, g.punkte, "gleiter")) { g.blitz = 1; }
    /* Jackpot oder Schloss */
    var lo = d.loch;
    if (lo.sperre > 0) lo.sperre -= dt;
    if (lo.blitz > 0) lo.blitz -= dt * 2;
    if (lo.sperre <= 0 && Math.hypot(k.x - lo.x, k.y - lo.y) < lo.r + RK - 2) {
      lo.sperre = 1.2; lo.blitz = 1;
      var az = DECK.ausLoch;
      k.x = az.x; k.y = az.y; k.vx = az.vx; k.vy = az.vy;
      k.ebene = "haupt"; k.sperre = 0.4;
      sp.ereignisse.push({ art: "deckausgang", x: az.x, y: az.y, wert: 0 });
      if (sp.jackpot) {
        sp.jackpot = false;
        var jp = addiere(sp, sp.tisch.bonus * 4, lo.x, lo.y - 20, "jackpot");
        sp.meldung = "JACKPOT +" + jp;
        sp.ereignisse.push({ art: "jackpot", x: lo.x, y: lo.y, wert: jp });
      } else {
        var p = addiere(sp, 900, lo.x, lo.y - 20, "loch");
        sp.meldung = "Oberdeck-Ausgang +" + p;
      }
      return true;
    }
    /* Halteklammer: hält die Kugel kurz und lässt sie dann wieder los */
    var kl = d.klammer;
    if (kl.sperre > 0) kl.sperre -= dt;
    if (kl.blitz > 0) kl.blitz -= dt * 2;
    if (kl.haelt > 0) {
      kl.haelt -= dt;
      k.x = kl.x; k.y = kl.y; k.vx = 0; k.vy = 0;
      if (kl.haelt <= 0) {
        k.vx = 220 + zufall() * 120;
        k.vy = -120;
        kl.sperre = 2;
        sp.ereignisse.push({ art: "klammer", x: kl.x, y: kl.y, wert: 0 });
      }
      return true;
    }
    if (kl.sperre <= 0 && Math.hypot(k.x - kl.x, k.y - kl.y) < kl.r + RK) {
      kl.haelt = 1.1; kl.blitz = 1;
      k.x = kl.x; k.y = kl.y; k.vx = 0; k.vy = 0;
      addiere(sp, 600, kl.x, kl.y - 20, "klammer");
      sp.meldung = "Klammer hält die Kugel …";
      return true;
    }
    /* Ausgänge: Mittelloch in der Decke und Seitenloch rechts */
    if (k.y > DECK.y2 - 6 && k.x > DECK.rinne.x1 && k.x < DECK.rinne.x2) {
      var a = DECK.ausRinne;
      k.x = a.x; k.y = a.y; k.vx = a.vx + (k.x - 200) * 0.2; k.vy = a.vy;
      k.ebene = "haupt";
      k.sperre = 0.4;
      sp.ereignisse.push({ art: "decksprung", x: a.x, y: a.y, wert: 0 });
      addiere(sp, 400, a.x, a.y - 8, "decksprung");
      return true;
    }
    if (Math.hypot(k.x - DECK.loch.x, k.y - DECK.loch.y) < DECK.loch.r + RK - 3) {
      var b2 = DECK.ausLoch;
      k.x = b2.x; k.y = b2.y; k.vx = b2.vx; k.vy = b2.vy;
      k.ebene = "haupt";
      k.sperre = 0.4;
      sp.ereignisse.push({ art: "deckausgang", x: b2.x, y: b2.y, wert: 0 });
      addiere(sp, 700, b2.x, b2.y - 8, "deckausgang");
      return true;
    }
    return false;
  }

  /* ------------------------------------------------------------ Mikroschritt */
  function mikroschrittHaupt(sp, k, dt) {
    var i;
    k.vy += sp.tisch.grav * dt;
    k.x += k.vx * dt;
    k.y += k.vy * dt;

    for (i = 0; i < sp.tisch.waende.length; i++) wand(k, sp.tisch.waende[i], 0.55);

    /* Rücklaufklappe am Kopf der Outlanes: was nach oben geschossen wird (Kick-
       back, Flipper), fällt ins Feld statt in dieselbe Gasse zurück. */
    if (k.vy < -120 && k.y < 606 && k.y > 518) {
      if (k.x < LANES.links.wand[0]) { k.vx = Math.max(k.vx, 260); k.vy = Math.min(k.vy, -240); }
      else if (k.x > LANES.rechts.wand[0]) { k.vx = Math.min(k.vx, -260); k.vy = Math.min(k.vy, -240); }
    }

    schleudern(sp, k, dt);

    /* Kanalauslauf: nur wenn die Kugel im Kanal nach oben schießt */
    if (sp.auslaufSperre > 0) sp.auslaufSperre -= dt;
    if (sp.auslaufSperre <= 0 && k.vy < -120 && k.x > KANAL.x1 + 2 && k.y < AUSLAUF.ein.y + 16 && k.y > AUSLAUF.ein.y - 40) {
      sp.auslaufSperre = 1.1;
      bahnStarten(sp, k, AUSLAUF);
      return;
    }

    for (i = 0; i < sp.tisch.bumper.length; i++) {
      var b = sp.tisch.bumper[i];
      if (b.cooldown > 0) b.cooldown -= dt;
      if (b.blitz > 0) b.blitz -= dt * 3;
      kreis(sp, k, b, b.r, b.punkte, b.kick, "bumper");
    }

    for (i = 0; i < sp.tisch.zieleOben.length; i++) {
      var z = sp.tisch.zieleOben[i];
      if (z.blitz > 0) z.blitz -= dt * 3;
      if (rechteck(sp, k, z, z.punkte, "ziel")) { z.blitz = 1; zielGetroffen(sp, i, z.x + z.w / 2, z.y); }
    }

    for (i = 0; i < sp.tisch.fallen.length; i++) {
      var f = sp.tisch.fallen[i];
      if (f.blitz > 0) f.blitz -= dt * 3;
      if (f.wecker > 0) {
        f.wecker -= dt;
        if (f.wecker <= 0) f.runter = false;
        continue;
      }
      if (rechteck(sp, k, f, f.punkte, "fallziel")) {
        f.runter = true; f.blitz = 1; f.wecker = 10;
        sp.zieleRunter++;
        var alle = true;
        for (var q = 0; q < sp.tisch.fallen.length; q++) if (!sp.tisch.fallen[q].runter) alle = false;
        if (alle) {
          for (q = 0; q < sp.tisch.fallen.length; q++) { sp.tisch.fallen[q].wecker = 1.2; }
          addiere(sp, sp.tisch.bonus, 200, 500, "bonus");
          sp.meldung = "Fallziele komplett – Bonusphase!";
          sp.ereignisse.push({ art: "fallen", x: 200, y: 520, wert: 1 });
          bonusPhase(sp);
        }
      }
    }

    spinner(sp, k, dt);
    kugelfang(sp, k, dt);
    schlossLoch(sp, k, dt);

    /* Rampen */
    if (sp.rampenSperre > 0) sp.rampenSperre -= dt;
    if (k.sperre > 0) k.sperre -= dt;
    for (i = 0; i < sp.tisch.rampen.length; i++) {
      var r = sp.tisch.rampen[i];
      if (r.sperre > 0) r.sperre -= dt;
      if (r.blitz > 0) r.blitz -= dt * 2;
      if (k.sperre > 0 || r.sperre > 0 || sp.rampenSperre > 0) continue;
      if (Math.hypot(k.x - r.ein.x, k.y - r.ein.y) < r.ein.r + RK * 0.6) {
        r.sperre = 0.6;
        rampeStarten(sp, k, r);
        return;
      }
    }

    /* Magnet: zieht die Kugel in der Bonusphase an, stößt sie sonst ab */
    var mg = sp.tisch.magnet;
    var mdx = mg.x - k.x, mdy = mg.y - k.y;
    var md = Math.hypot(mdx, mdy) || 1;
    if (sp.bonusZeit > 0 && md < 74) {
      var f2 = 1500 * (1 - md / 74);
      k.vx += mdx / md * f2 * dt;
      k.vy += mdy / md * f2 * dt;
      mg.an = Math.min(1, mg.an + dt * 4);
    } else if (md < mg.r + RK + 3) {
      k.vx -= mdx / md * 2400 * dt;
      k.vy -= mdy / md * 2400 * dt;
      mg.an = Math.min(1, mg.an + dt * 4);
    } else {
      mg.an = Math.max(0, mg.an - dt * 3);
    }

    /* Rollover */
    for (i = 0; i < sp.tisch.rollover.length; i++) {
      var ro = sp.tisch.rollover[i];
      if (ro.blitz > 0) ro.blitz -= dt * 3;
      if (ro.an) continue;
      if (k.x > ro.x && k.x < ro.x + ro.w && k.y > ro.y && k.y < ro.y + ro.h) {
        ro.an = true; ro.blitz = 1;
        addiere(sp, 150, ro.x + ro.w / 2, ro.y - 6, "licht");
        var alleR = true;
        for (var q2 = 0; q2 < sp.tisch.rollover.length; q2++) if (!sp.tisch.rollover[q2].an) alleR = false;
        if (alleR) {
          sp.multi = Math.min(9, sp.multi + 1);
          addiere(sp, 1200, 200, 330, "licht");
          for (q2 = 0; q2 < sp.tisch.rollover.length; q2++) sp.tisch.rollover[q2].an = false;
          sp.meldung = "Alle Lichter an – Multiplikator ×" + sp.multi;
          sp.ereignisse.push({ art: "lichter", x: 200, y: 330, wert: sp.multi });
        }
      }
    }

    /* Kickback in der linken Outlane – schießt die Kugel gerade nach oben,
       damit sie nicht an der Inlane-Schräge hängen bleibt. */
    if (sp.kickback > 0 && k.ebene === "haupt" && k.y > 700 && k.x < 56 && k.vy > 0) {
      sp.kickback = 0;
      k.vy = -1250; k.vx = -10;
      sp.meldung = "Kickback!";
      sp.ereignisse.push({ art: "kickback", x: 60, y: 720, wert: 0 });
    }

    for (i = 0; i < sp.flipper.length; i++) flipperStoss(sp, k, sp.flipper[i], 0.55);

    var v = Math.hypot(k.vx, k.vy);
    if (v > MAXV) { k.vx *= MAXV / v; k.vy *= MAXV / v; }

    /* Auslass */
    if (k.y > H - 2 || k.y < -260 || k.x < -80 || k.x > W + 80) kugelVerloren(sp, k);
  }

  function mikroschrittDeck(sp, k, dt) {
    var i;
    k.vy += sp.tisch.grav * dt;
    k.x += k.vx * dt;
    k.y += k.vy * dt;
    var d = sp.tisch.deck;
    for (i = 0; i < d.waende.length; i++) wand(k, d.waende[i], 0.5);
    deckSchritt(sp, k, dt);
    var v = Math.hypot(k.vx, k.vy);
    if (v > MAXV) { k.vx *= MAXV / v; k.vy *= MAXV / v; }
  }

  function mikroschritt(sp, k, dt) {
    if (k.zustand === "bahn") { bahnSchritt(sp, k, dt); return; }
    if (k.ebene === "deck") mikroschrittDeck(sp, k, dt);
    else mikroschrittHaupt(sp, k, dt);
  }

  /* ------------------------------------------------------------ Ballverlust */
  function kugelVerloren(sp, k) {
    if (k.zustand === "weg") return;
    k.zustand = "weg";
    sp.verloren++;
    sp.ereignisse.push({ art: "verlust", x: Math.max(20, Math.min(W - 20, k.x)), y: Math.min(H - 10, k.y), wert: 0 });
    sp.kugeln = sp.kugeln.filter(function (x) { return x !== k; });
    nachDrain(sp);
  }

  function imSpiel(sp) {
    var n = 0;
    for (var i = 0; i < sp.kugeln.length; i++) if (sp.kugeln[i].zustand === "feld" || sp.kugeln[i].zustand === "bahn") n++;
    return n;
  }

  function nachDrain(sp) {
    if (imSpiel(sp) > 0) { sp.zustand = "spiel"; return; }
    if (sp.schloss.length >= 2) { multiballStarten(sp, "schloss"); return; }
    /* Letzte Kugel und noch eine im Schloss: freigeben statt beenden */
    if (sp.kugelnUebrig <= 0 && sp.schloss.length === 1) {
      multiballStarten(sp, "schloss");
      sp.kugelnUebrig = 0;
      return;
    }
    if (sp.kugelnUebrig > 0) {
      sp.tisch.fallen.forEach(function (f) { f.runter = false; f.wecker = 0; });
      if (sp.extraBall) { sp.extraBall = false; sp.kugelnUebrig++; sp.kugelnGesamt++; }
      sp.kugelnUebrig--;
      sp.zustand = "bereit";
      sp.abschuss.kraft = 0;
      sp.saverZeit = 1;
      sp.meldung = "Kugel " + (sp.kugelnGesamt - sp.kugelnUebrig + 1) + " von " + sp.kugelnGesamt + ".";
      return;
    }
    sp.zustand = "vorbei";
  }

  /* ---------------------------------------------------------------- Schritt */
  /* Deck-Elemente ticken auch ohne Kugel auf dem Deck (Gleiter, Fallziele). */
  function zielTakt(sp, dt) {
    var i;
    var so = sp.tisch.scoop, sc = sp.tisch.schloss;
    if (so.sperre > 0) so.sperre -= dt;
    if (so.blitz > 0) so.blitz -= dt * 2;
    if (sc.sperre > 0) sc.sperre -= dt;
    if (sc.blitz > 0) sc.blitz -= dt * 2;
    for (i = 0; i < sp.tisch.zieleOben.length; i++) {
      var zo = sp.tisch.zieleOben[i];
      if (zo.sperre > 0) zo.sperre -= dt;
    }
    for (i = 0; i < sp.tisch.fallen.length; i++) {
      var f = sp.tisch.fallen[i];
      if (f.wecker > 0) { f.wecker -= dt; if (f.wecker <= 0) f.runter = false; }
    }
  }

  function deckTakt(sp, dt) {
    var d = sp.tisch.deck, g = d.gleiter;
    if (g.sperre > 0) g.sperre -= dt;
    g.phase += dt * 0.75;
    g.x = g.x0 + (g.x1 - g.x0) * (0.5 - 0.5 * Math.cos(g.phase));
    for (var i = 0; i < d.ziele.length; i++) {
      var z = d.ziele[i];
      if (z.wecker > 0) { z.wecker -= dt; if (z.wecker <= 0) z.runter = false; }
    }
    var k = d.klammer;
    if (k.sperre > 0) k.sperre -= dt;
    if (k.blitz > 0) k.blitz -= dt * 2;
  }

  function schritt(sp, dt) {
    if (!(dt > 0)) return;
    if (dt > 1 / 30) dt = 1 / 30;
    sp.zeit += dt;
    flipperDrehen(sp, dt);
    if (sp.bonusZeit > 0) sp.bonusZeit = Math.max(0, sp.bonusZeit - dt);
    deckTakt(sp, dt);
    zielTakt(sp, dt);
    if (sp.saverZeit > 0) sp.saverZeit -= dt;

    if (sp.zustand === "bereit") {
      if (sp.abschuss.laedt) sp.abschuss.kraft = Math.min(1, sp.abschuss.kraft + dt * 1.5);
      return;
    }
    if (sp.zustand !== "spiel") return;

    var i, k, n = Math.max(1, Math.min(12, Math.ceil(dt / (1 / 480))));
    var h = dt / n;
    for (var s = 0; s < n; s++) {
      for (i = sp.kugeln.length - 1; i >= 0; i--) {
        k = sp.kugeln[i];
        if (!k || k.zustand === "weg" || k.zustand === "schloss") continue;
        if (k.sperre > 0) k.sperre -= h;
        mikroschritt(sp, k, h);
        /* Ballsuche wie am Automaten: bleibt die Kugel auf der Stelle (auch
           wenn sie nur auf einem Blechrand zittert), wird sie angestupst und
           nach dem dritten Versuch oben im Feld neu eingesetzt. */
        if (k.zustand === "feld") {
          if (!k.anker) k.anker = { x: k.x, y: k.y };
          if (Math.hypot(k.x - k.anker.x, k.y - k.anker.y) > 4) {
            k.anker.x = k.x; k.anker.y = k.y; k.still = 0; k.stups = 0;
          } else {
            k.still += h;
          }
          if (k.still > 6) {
            k.still = 0;
            k.stups++;
            k.anker.x = k.x; k.anker.y = k.y;
            if (k.stups >= 3) {
              k.x = 200 + (zufall() - 0.5) * 50;
              k.y = 400;
              k.vx = (zufall() - 0.5) * 160;
              k.vy = -140;
              k.stups = 0;
            } else {
              /* zur Feldmitte anstoßen – dort liegt der Flipper */
              k.vx += (k.x < W / 2 ? 1 : -1) * (140 + zufall() * 120);
              k.vy -= 220;
            }
            sp.ereignisse.push({ art: "stups", x: k.x, y: k.y, wert: k.stups });
          }
        }
      }
      if (sp.zustand !== "spiel") return;
    }
  }

  /* Ereignisse abholen (Sound/Effekte) – der Kern bleibt DOM-frei. */
  function ereignisseHolen(sp) {
    var e = sp.ereignisse;
    sp.ereignisse = [];
    return e;
  }

  var FL = {
    W: W, H: H, RK: RK, TAKT: TAKT, MAXV: MAXV, KANAL: KANAL, DECK: DECK,
    AUSLAUF: AUSLAUF, TISCHE: TISCHE, DECK_WAENDE: DECK_WAENDE, LANES: LANES,
    zufallSetzen: zufallSetzen, neuesSpiel: neuesSpiel, schritt: schritt,
    abschuss: abschuss, abschussLaden: abschussLaden, fluegel: fluegel,
    ereignisseHolen: ereignisseHolen, naechsterPunkt: naechsterPunkt,
    abstandSegment: abstandSegment, bahnPunkt: bahnPunkt,
    imSpiel: imSpiel, multiWert: multiWert, neuerBall: neueKugel, partieStarten: partieStarten
  };
  root.FL = FL;
  if (typeof module === "object" && module.exports) module.exports = FL;
})(typeof window !== "undefined" ? window : globalThis);
