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
  /* tor = Oberkante der Auslauföffnung in der Kanalwand. Darüber ist der Kanal
     mit der Decke dicht; die Kugel verlässt ihn nur durch diese Öffnung. */
  var KANAL = { x1: 372, x2: 420, mitte: 396, boden: 770, stift: 748, oben: 322, tor: 352 };

  /* --------------------------------------------------------------- Oberdeck
     Eigene Spielebene über dem Hauptfeld. Sie hat einen eigenen Rahmen, eigene
     Bumper, ein eigenes Zielband, ein Jackpot-Loch, eine Halteklammer und zwei
     Ausgänge (Mitte unten ins Hauptfeld, rechts unten in die rechte Inlane). */
  var DECK = {
    x1: 78, y1: 70, x2: 356, y2: 320,
    ecke: { x: 32, y: 18 },             // abgeschrägte obere Ecken der Deckplatte
    ein: { x: 338, y: 116 },            // hier kommt die Kugel aus dem Aufzug an
    rinne: { x1: 150, x2: 214 },        // Auslass in der Decke (Mitte unten)
    /* Umlaufschiene: Mittellinie der Bahn, auf der die Kugel aus dem Aufzug
       einmal quer über das Deck läuft. Sie ist nur gezeichnet: als Wand wäre
       ihr gerades Mittelstück ein Blech, auf dem die Kugel liegen bliebe. */
    schiene: [{ x: 100, y: 156 }, { x: 132, y: 136 }, { x: 288, y: 136 }, { x: 320, y: 156 }],
    loch: { x: 330, y: 300, r: 17 },    // Seitenausgang rechts unten
    ausRinne: { x: 182, y: 386, vx: 0, vy: 130 },
    spinner: { x: 214, y: 104, len: 22, punkte: 150 },   // Umlauf-Spinner im Oberdeck
    ausLoch: { x: 326, y: 632, vx: -30, vy: 200 }   // in die rechte Inlane
  };

  /* ---------------------------------------------------------------- Rahmen
     Hauptfeld: oben dicht bei y 322, unten wie am echten Automaten je eine
     Inlane innen (Trennstab, dann eine Schräge, die die Kugel auf den Flipper
     legt) und außen eine schmale Outlane an der Bande. Der Trennstab endet
     oben in einem Pfosten, an dem sich der Weg teilt.
     Der Abschusskanal ist ein geschlossener Schacht: oben mit der Decke dicht,
     unten liegt die Kugel auf dem Kolbenkopf (KANAL.stift). Einziger Ausgang
     ist die Öffnung in der Kanalwand zwischen Decke und KANAL.tor. */
  var LANES = {
    links:  { wand: [56, 556],  knie: [56, 652],  ende: [110, 690],     sling: [92, 600, 92, 640], sohle: [92, 640, 158, 640] },
    rechts: { wand: [340, 556], knie: [340, 652], ende: [282, 690],     sling: [310, 600, 310, 640], sohle: [310, 640, 244, 640] }
  };

  function laneStrecken(l) {
    return [
      [l.wand[0], l.wand[1], l.knie[0], l.knie[1]],
      [l.knie[0], l.knie[1], l.ende[0], l.ende[1]],
      [l.sling[0], l.sling[1], l.sling[2], l.sling[3]],
      [l.sohle[0], l.sohle[1], l.sohle[2], l.sohle[3]]   // Sohle: der Schleuderkoerper ist massiv
    ];
  }

  var RAHMEN = [
    [30, 322, 372, 322],          // Decke des Hauptfelds bis an den Kanal
    [30, 322, 30, 772],           // linke Bande
    [372, KANAL.tor, 372, 806],   // Kanalwand, oben die Auslauföffnung
    [420, 120, 420, 806],         // rechte Außenwand
    [372, 322, 420, 322]          // Kanal oben dicht
  ].concat(
    laneStrecken(LANES.links),    // Inlane/Outlane links
    laneStrecken(LANES.rechts)    // Inlane/Outlane rechts
  );

  /* Einbahnklappe an der Auslauföffnung des Kanals: von innen offen, von außen
     dicht. Sie wird nicht gezeichnet, die Öffnung bleibt sichtbar. */
  var KANAL_TOR = { x1: KANAL.x1, y1: KANAL.oben, x2: KANAL.x1, y2: KANAL.tor, tor: true };

  /* Der Weg aus dem Kanal ins Feld: oben aus der Öffnung in der Kanalwand
     heraus und unter der Decke nach links. Die Bahn bleibt unter der Decke,
     damit die Kugel nicht über dem Spielfeld gezeichnet wird. */
  var AUSLAUF = {
    name: "Auslauf", tempo: 900, punkte: 0,
    ein: { x: 396, y: 336, r: 30 },
    bahn: [{ x: 396, y: 344 }, { x: 392, y: 336 }, { x: 372, y: 334 }, { x: 344, y: 332 }, { x: 316, y: 338 }],
    aus: { x: 300, y: 346, vx: -330, vy: 50 }
  };

  /* Schleudern: die Unterkante liegt bei y 640, nicht 655. Sonst bleibt
     zwischen ihrer Außenecke und der Inlane nur ein 14 px breiter Durchlass –
     zu wenig für eine 17 px große Kugel, die dort festklemmt. */
  var SLING_L = { x1: 92, y1: 600, x2: 158, y2: 640, nx: 0.519, ny: -0.855 };
  var SLING_R = { x1: 310, y1: 600, x2: 244, y2: 640, nx: -0.519, ny: -0.855 };

  /* Zielbank auf dem Oberdeck – auf allen vier Tischen dieselbe (linke Bande). */
  var DECK_ZIELE = [[84, 176, 16, 30], [84, 210, 16, 30], [84, 244, 16, 30]];

  /* --------------------------------------------------------------- Tische
     Gemeinsame Architektur, je Tisch eigene Bumper, Bänke, Rampen und Werte. */
  var TISCHE = [
    {
      id: "zwerg", name: "Zwergenwerk", satz: 1, grav: GRAV, bonus: 3000,
      tipp: "Stollenwerk mit sechs Bumpern, Aufzug zum Oberdeck und Schloss links.",
      bumper: [[152, 372, 19, 130, 280], [252, 372, 19, 130, 280], [202, 442, 20, 150, 285],
               [124, 478, 16, 170, 270], [280, 478, 16, 170, 270]],
      deckBumper: [[146, 196, 23], [226, 234, 26], [306, 196, 23], [226, 168, 15]],
      deckZiele: DECK_ZIELE,
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
          bahn: [[330, 466], [348, 436], [358, 400], [364, 340], [364, 280], [356, 220], [348, 160]],
          aus: { x: DECK.ein.x, y: DECK.ein.y, vx: -280, vy: 20 } },
        { name: "Schleife", punkte: 900, tempo: 600, ziel: "weiche",
          ein: { x: 96, y: 466, r: 14 },
          bahn: [[96, 466], [70, 430], [60, 370], [58, 300], [66, 292], [74, 330]],
          aus: { x: 78, y: 344, vx: 210, vy: 150 },
          aus2: { x: 326, y: 632, vx: -30, vy: 210 } }
      ]
    },
    {
      id: "neon", name: "Neon-Stadt", satz: 2, grav: GRAV + 40, bonus: 3500,
      tipp: "Sechs Bumper im Rautenmuster, schnelle Schleife, breites Zielband.",
      bumper: [[150, 378, 17, 140, 290], [252, 378, 17, 140, 290], [201, 368, 18, 140, 290],
               [176, 438, 16, 150, 280], [226, 438, 16, 150, 280], [201, 494, 15, 160, 285]],
      deckBumper: [[140, 200, 21], [226, 240, 24], [310, 200, 21], [226, 166, 15]],
      deckZiele: DECK_ZIELE,
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
          bahn: [[330, 462], [348, 432], [358, 396], [364, 336], [364, 276], [356, 216], [348, 158]],
          aus: { x: DECK.ein.x, y: DECK.ein.y, vx: -300, vy: 10 } },
        { name: "Datenloop", punkte: 1000, tempo: 640, ziel: "weiche",
          ein: { x: 96, y: 462, r: 14 },
          bahn: [[96, 462], [70, 428], [60, 368], [58, 298], [66, 290], [74, 330]],
          aus: { x: 78, y: 342, vx: 220, vy: 160 },
          aus2: { x: 336, y: 640, vx: -40, vy: 260 } }
      ]
    },
    {
      id: "pirat", name: "Piratenbucht", satz: 3, grav: GRAV - 30, bonus: 3800,
      tipp: "Fünf Bumper, zwei Zielbänke, Planke oben rechts, Schloss am Heck.",
      bumper: [[152, 366, 18, 130, 285], [250, 366, 18, 130, 285],
               [112, 438, 15, 170, 275], [292, 438, 15, 170, 275], [201, 476, 19, 150, 280]],
      deckBumper: [[146, 192, 24], [224, 232, 25], [304, 192, 24], [226, 168, 16]],
      deckZiele: DECK_ZIELE,
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
          bahn: [[330, 458], [348, 428], [358, 392], [364, 334], [364, 274], [356, 214], [348, 156]],
          aus: { x: DECK.ein.x, y: DECK.ein.y, vx: -290, vy: 15 } },
        { name: "Kombüse", punkte: 950, tempo: 610, ziel: "weiche",
          ein: { x: 96, y: 458, r: 14 },
          bahn: [[96, 458], [70, 424], [60, 364], [58, 296], [66, 290], [74, 330]],
          aus: { x: 78, y: 340, vx: 215, vy: 155 },
          aus2: { x: 326, y: 632, vx: -30, vy: 205 } }
      ]
    },
    {
      id: "stern", name: "Sternenstaub", satz: 4, grav: GRAV - 60, bonus: 4500,
      tipp: "Meisterklasse: große Bumper, sieben Ziele, hohe Werte, lange Wege.",
      bumper: [[150, 368, 21, 180, 310], [252, 368, 21, 180, 310], [201, 440, 18, 200, 300],
               [118, 484, 15, 210, 295], [284, 484, 15, 210, 295], [201, 498, 14, 220, 300]],
      deckBumper: [[144, 190, 25], [226, 238, 27], [308, 190, 25], [226, 164, 16]],
      deckZiele: DECK_ZIELE,
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
          bahn: [[330, 470], [350, 440], [360, 402], [366, 340], [366, 278], [358, 218], [350, 160]],
          aus: { x: DECK.ein.x, y: DECK.ein.y, vx: -310, vy: 5 } },
        { name: "Kometenschleife", punkte: 1200, tempo: 660, ziel: "weiche",
          ein: { x: 96, y: 470, r: 14 },
          bahn: [[96, 470], [68, 434], [58, 372], [56, 296], [66, 288], [74, 330]],
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

  function strecken(punkte) {
    var s = [];
    for (var i = 1; i < punkte.length; i++) {
      s.push([punkte[i - 1].x, punkte[i - 1].y, punkte[i].x, punkte[i].y]);
    }
    return s;
  }

  /* Deckrahmen: Außenkanten mit abgeschrägten Ecken, Boden mit Auslass in der
     Mitte und die Umlaufschiene unter der Decke (durch sie läuft die Kugel aus
     dem Aufzug einmal quer über das Deck). Alles aus DECK abgeleitet – der
     Zeichner nimmt dieselben Zahlen. */
  var DECK_WAENDE = [
    [DECK.x1, DECK.y2, DECK.x1, DECK.y1 + DECK.ecke.y],
    [DECK.x1, DECK.y1 + DECK.ecke.y, DECK.x1 + DECK.ecke.x, DECK.y1],
    [DECK.x1 + DECK.ecke.x, DECK.y1, DECK.x2 - DECK.ecke.x, DECK.y1],
    [DECK.x2 - DECK.ecke.x, DECK.y1, DECK.x2, DECK.y1 + DECK.ecke.y],
    [DECK.x2, DECK.y1 + DECK.ecke.y, DECK.x2, DECK.y2],
    /* Der Boden fällt zu beiden Seiten zur Rinne ab: eine Kugel rollt aus jeder
       Ecke in den Auslass, statt in der Ecke liegen zu bleiben. Die Rinne
       selbst bleibt offen, dort geht es zurück ins Hauptfeld. */
    [DECK.x1, DECK.y2 - 20, DECK.rinne.x1, DECK.y2],
    [DECK.rinne.x2, DECK.y2, DECK.x2, DECK.y2 - 20]
  ].map(strecke);

  /* Die Umlaufschiene (DECK.schiene) ist reine Zeichnung: als Wand wäre ihr
     gerades Mittelstück ein Blech, auf dem die Kugel liegen bleiben könnte. */

  function tischKopie(def) {
    var t = {
      id: def.id, name: def.name, satz: def.satz, grav: def.grav, tipp: def.tipp, bonus: def.bonus,
      waende: RAHMEN.concat(def.waende || []).map(strecke),
      tore: [KANAL_TOR],
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
        { x: 152, y: 326, w: 30, h: 24, an: false, blitz: 0 },
        { x: 202, y: 324, w: 30, h: 26, an: false, blitz: 0 },
        { x: 252, y: 326, w: 30, h: 24, an: false, blitz: 0 }
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
        waende: DECK_WAENDE.slice(),
        bumper: def.deckBumper.map(function (b) {
          return { x: b[0], y: b[1], r: b[2], punkte: 400, kick: 230, blitz: 0, cooldown: 0 };
        }),
        ziele: def.deckZiele.map(function (p) {
          return { x: p[0], y: p[1], w: p[2], h: p[3], punkte: 600, runter: false, blitz: 0, wecker: 0, sperre: 0 };
        }),
        gleiter: { x: 202, y: 276, w: 44, h: 14, x0: 138, x1: 258, phase: 0, richtung: 1, punkte: 700, blitz: 0, sperre: 0 },
        loch: { x: DECK.loch.x, y: DECK.loch.y, r: DECK.loch.r, sperre: 0, blitz: 0 },
        klammer: { x: 100, y: 296, r: 17, sperre: 0, haelt: 0, blitz: 0 },
        spinner: { x: DECK.spinner.x, y: DECK.spinner.y, len: DECK.spinner.len, punkte: DECK.spinner.punkte, winkel: 0, dreh: 0, treffer: 0, sperre: 0 }
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
      bahn: null, sperre: 0,
      anker: null, rettStufe: 0, rettFrei: 0,   // Kugelrettung
      vorX: x, vorY: y                           // Lage vor dem Mikroschritt
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
      kickback: 1,                  // 1 = geladen
      extraBall: false,
      jackpot: false, schlossOffen: false,
      popups: [], ereignisse: [], meldung: "",
      auslaufSperre: 0, rampenSperre: 0
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

  /* ------------------------------------------------------------- Kollision
     Ein Stoß setzt die Kugel auf die Fläche zurück und spiegelt nur die
     Normalkomponente; die Bewegung längs der Fläche bleibt erhalten (bis auf
     Reibung). Die Kugel sackt dadurch an keiner Wand auf null ab, und der
     Abprallwinkel folgt der Fläche. */
  function abprall(k, cx, cy, nx, ny, e, reibung, tief) {
    var vn = k.vx * nx + k.vy * ny;
    /* Nur eingreifen, wenn die Kugel in die Fläche hineinläuft (oder tief
       darin steckt). Sonst würde sie an einer gemeinsamen Ecke zweier
       Flächen in jedem Mikroschritt neu festgesetzt und käme nie wieder los. */
    if (vn >= 0 && !tief) return 0;
    k.x = cx + nx * (RK + 0.4);
    k.y = cy + ny * (RK + 0.4);
    if (vn >= 0) return 0;
    var j = -(1 + e) * vn;
    k.vx += j * nx;
    k.vy += j * ny;
    if (reibung) {
      var tx = -ny, ty = nx;
      var vt = k.vx * tx + k.vy * ty;
      var dmp = Math.min(1, reibung);
      k.vx -= vt * dmp * tx;
      k.vy -= vt * dmp * ty;
    }
    return -vn;
  }

  function stoss(k, c, nx, ny, e, reibung, extra) {
    abprall(k, c.x, c.y, nx, ny, e, reibung);
    if (extra) { k.vx += nx * extra; k.vy += ny * extra; }
  }

  /* Wand als Linie. Die Kugel wird auf die Seite zurückgeschoben, von der sie
     kam: so kann sie bei einem großen Schritt nicht hinter die Wand rutschen
     und auf der falschen Seite kleben bleiben. Ein "tor" wirkt nur von der
     Feldseite aus – die Einbahnklappe am Abschusskanal. */
  function wand(k, a, e) {
    if (a.tor && k.x >= a.x1) return false;
    var c = naechsterPunkt(k.x, k.y, a.x1, a.y1, a.x2, a.y2);
    var dx = k.x - c.x, dy = k.y - c.y;
    var d = Math.hypot(dx, dy);
    if (d >= RK) return false;
    var nx, ny;
    if (d > 0.0001) { nx = dx / d; ny = dy / d; }
    else {
      /* Genau auf der Linie: die Seite von vor dem Schritt entscheidet. */
      var px = k.vorX - c.x, py = k.vorY - c.y, pl = Math.hypot(px, py);
      if (pl > 0.0001) { nx = px / pl; ny = py / pl; }
      else { var l = Math.hypot(k.vx, k.vy) || 1; nx = -k.vx / l; ny = -k.vy / l; }
    }
    if (k.vorX !== undefined && (k.vorX - c.x) * nx + (k.vorY - c.y) * ny < 0) {
      nx = -nx; ny = -ny;
    }
    abprall(k, c.x, c.y, nx, ny, e, 0.02, d < RK * 0.7);
    return true;
  }

  /* Liegt die Kugel im massiven Schleuderkoerper, wird sie zur naechsten
     Kante hinausgesetzt. Sonst pendelte sie zwischen Sohle und Gummiband hin
     und her, ohne je wieder ins Feld zu kommen. */
  function slingRaus(k, s) {
    var ax = s.x1, ay = s.y1, bx = s.x2, by = s.y2, cx = s.x1, cy = s.y2;
    var d1 = (bx - ax) * (k.y - ay) - (by - ay) * (k.x - ax);
    var d2 = (cx - bx) * (k.y - by) - (cy - by) * (k.x - bx);
    var d3 = (ax - cx) * (k.y - cy) - (ay - cy) * (k.x - cx);
    if (!((d1 >= 0 && d2 >= 0 && d3 >= 0) || (d1 <= 0 && d2 <= 0 && d3 <= 0))) return false;
    var kanten = [[ax, ay, bx, by], [bx, by, cx, cy], [cx, cy, ax, ay]];
    var best = null, bd = 1e9;
    for (var i = 0; i < kanten.length; i++) {
      var e = kanten[i];
      var c = naechsterPunkt(k.x, k.y, e[0], e[1], e[2], e[3]);
      var dd = Math.hypot(k.x - c.x, k.y - c.y);
      if (dd < bd) { bd = dd; best = { c: c, dx: e[2] - e[0], dy: e[3] - e[1] }; }
    }
    var nl = Math.hypot(best.dx, best.dy) || 1;
    var nx = best.dy / nl, ny = -best.dx / nl;      // nach außen
    k.x = best.c.x + nx * (RK + 0.5);
    k.y = best.c.y + ny * (RK + 0.5);
    k.vx = nx * 90 + k.vx * 0.2;
    k.vy = ny * 90 + k.vy * 0.2;
    return true;
  }

  function kreis(sp, k, o, r, punkte, kick, art) {
    var dx = k.x - o.x, dy = k.y - o.y;
    var d = Math.hypot(dx, dy);
    if (d >= r + RK) return false;
    var nx, ny;
    if (d > 0.0001) { nx = dx / d; ny = dy / d; }
    else { nx = 0; ny = -1; }
    abprall(k, o.x + nx * r, o.y + ny * r, nx, ny, 0.55, 0.02);
    if (o.cooldown !== undefined && o.cooldown > 0) return true;
    /* Der Kick kommt einmal je Treffer, nicht in jedem Mikroschritt – sonst
       beschleunigt ein einziger Kontakt die Kugel schlagartig auf das Maximum. */
    if (kick) { k.vx += nx * kick; k.vy += ny * kick; }
    if (o.cooldown !== undefined) o.cooldown = 0.11;
    if (o.blitz !== undefined) o.blitz = 1;
    if (punkte) addiere(sp, punkte, o.x, o.y - r * 0.4, art || "bumper");
    return true;
  }

  /* Ziele, Fallziele und der Gleiter sind Kapseln entlang ihrer langen Achse:
     oben und unten rund, damit die Kugel nicht auf einem geraden Blech liegen
     bleibt. Die Außenmaße entsprechen genau dem gezeichneten Rechteck. */
  function rechteck(sp, k, z, punkte, art) {
    if (z.runter) return false;
    var rr = Math.min(z.w, z.h) / 2;
    var ax, ay, bx, by;
    if (z.h >= z.w) { ax = bx = z.x + z.w / 2; ay = z.y + rr; by = z.y + z.h - rr; }
    else { ay = by = z.y + z.h / 2; ax = z.x + rr; bx = z.x + z.w - rr; }
    var c = naechsterPunkt(k.x, k.y, ax, ay, bx, by);
    var dx = k.x - c.x, dy = k.y - c.y;
    var d = Math.hypot(dx, dy);
    if (d >= RK + rr) return false;
    var nx, ny;
    if (d > 0.0001) { nx = dx / d; ny = dy / d; }
    else {
      var l = Math.hypot(k.vx, k.vy) || 1;
      nx = -k.vx / l; ny = -k.vy / l;
    }
    if (k.vorX !== undefined && (k.vorX - c.x) * nx + (k.vorY - c.y) * ny < 0) {
      nx = -nx; ny = -ny;
    }
    abprall(k, c.x + nx * rr, c.y + ny * rr, nx, ny, 0.45, 0.02, d < RK * 0.7);
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
  /* Nur das Laden beginnt hier. Beim Loslassen bleibt die aufgebaute Kraft
     stehen, damit der Aufrufer sie an abschuss() weitergeben kann; erst der
     Schuss selbst setzt sie zurück. */
  function abschussLaden(sp, an) {
    if (an && sp.zustand !== "bereit") return;
    sp.abschuss.laedt = !!an;
    if (an) sp.abschuss.kraft = 0;
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
      if (slingRaus(k, s)) continue;
      var d = abstandSegment(k.x, k.y, s.x1, s.y1, s.x2, s.y2);
      if (d >= RK + 4) continue;
      var l2 = (s.x2 - s.x1) * (s.x2 - s.x1) + (s.y2 - s.y1) * (s.y2 - s.y1);
      var t = l2 > 0 ? ((k.x - s.x1) * (s.x2 - s.x1) + (k.y - s.y1) * (s.y2 - s.y1)) / l2 : 0.5;
      var c = naechsterPunkt(k.x, k.y, s.x1, s.y1, s.x2, s.y2);
      var dx = k.x - c.x, dy = k.y - c.y, dl = Math.hypot(dx, dy);
      /* Für Position und Abprall gilt der radiale Abstand zur Fläche. Ein
         fester Normalenvektor würde die Kugel an der Schleuderecke jeden
         Schritt auf denselben Punkt zurücksetzen. */
      var nx = dl > 0.001 ? dx / dl : s.nx, ny = dl > 0.001 ? dy / dl : s.ny;
      /* Auch an den Enden abprallen: sonst rutscht die Kugel dort durch die
         Schleuder in den Winkel dahinter. Gekickt und gezählt wird nur auf dem
         mittleren Stück und in Richtung der Schleuderfläche. */
      stoss(k, c, nx, ny, 0.45, 0, 0);
      if (nx * s.nx + ny * s.ny < 0.2) { nx = s.nx; ny = s.ny; }
      if (t >= 0.1 && t <= 0.9 && s.cooldown <= 0) {
        s.cooldown = 0.13; s.blitz = 1;
        /* Normalanteil neu, Seitenanteil zu 55 % behalten – kein Absacken. */
        var vt = k.vx * -ny + k.vy * nx;
        var vn = Math.max(s.kick, Math.abs(k.vx * nx + k.vy * ny) * 0.3 + s.kick);
        k.vx = nx * vn - ny * vt * 0.55;
        k.vy = ny * vn + nx * vt * 0.55;
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
      /* Frei zwischen Decke und Bumpern einsetzen, nicht in einem Bumper. */
      s.x = 182 + i * 40;
      s.y = 386;
      s.vx = (i ? 220 : -220);
      s.vy = 140;
    });
    sp.kugeln = sp.kugeln.concat(los);
    if (grund !== "extra") {
      var k = neueKugel(DECK.ein.x, DECK.ein.y, -280, 10, "deck");
      sp.kugeln.push(k);
    }
    sp.meldung = "MULTIBALL! " + sp.kugeln.length + " Kugeln im Spiel";
    sp.ereignisse.push({ art: "multiball", x: 200, y: 380, wert: sp.kugeln.length });
    sp.zustand = "spiel";
  }

  /* ------------------------------------------------------ Spielelemente Deck */
  function deckSchritt(sp, k, dt) {
    var d = sp.tisch.deck, i;
    /* Spinner in der Umlaufbahn: zählt jede Fahrt, bremst die Kugel leicht. */
    var sn = d.spinner;
    if (sn.sperre > 0) sn.sperre -= dt;
    sn.winkel += dt * (sn.dreh + Math.hypot(k.vx, k.vy) * 0.02);
    if (sn.dreh > 0) sn.dreh = Math.max(0, sn.dreh - dt * 2.4);
    if (sn.sperre <= 0 && Math.abs(k.y - sn.y) < 17 && Math.abs(k.x - sn.x) < sn.len) {
      sn.sperre = 0.13; sn.dreh = Math.min(22, sn.dreh + 10); sn.treffer++;
      k.vx *= 0.9;
      addiere(sp, sn.punkte, sn.x, sn.y - 18, "spinner");
      sp.ereignisse.push({ art: "spinner", x: sn.x, y: sn.y, wert: sn.treffer });
    }
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
    k.vorX = k.x; k.vorY = k.y;
    k.vy += sp.tisch.grav * dt;
    k.x += k.vx * dt;
    k.y += k.vy * dt;

    /* Abschusskanal: geschlossener Schacht. Die Kugel liegt auf dem Kolbenkopf
       auf (KANAL.stift) und bleibt in der Spur. Hinter den Kolben oder über
       die Decke hinaus kann sie nicht geraten; liegt sie auf dem Kolben, wird
       sie von dort gleich wieder abgeschossen. */
    if (k.x > KANAL.x1) {
      k.vx *= 0.97;
      if (k.x < KANAL.x1 + RK) { k.x = KANAL.x1 + RK; k.vx = Math.abs(k.vx) * 0.4; }
      if (k.x > KANAL.x2 - RK) { k.x = KANAL.x2 - RK; k.vx = -Math.abs(k.vx) * 0.4; }
      if (k.y > KANAL.stift) {
        k.y = KANAL.stift;
        if (k.vy > 0) k.vy = -k.vy * 0.22;
      }
      if (k.y >= KANAL.stift - 0.5 && Math.abs(k.vy) < 26) {
        k.vy = -1200;
        k.vx = (zufall() - 0.5) * 8;
        sp.ereignisse.push({ art: "abschuss", x: k.x, y: k.y, wert: 0.5 });
      }
    }

    for (i = 0; i < sp.tisch.waende.length; i++) wand(k, sp.tisch.waende[i], 0.55);
    /* Einbahnklappe an der Auslauföffnung: die Kugel kommt heraus, aber von
       draußen nicht wieder hinein. */
    for (i = 0; i < sp.tisch.tore.length; i++) wand(k, sp.tisch.tore[i], 0.4);

    /* Rücklaufklappe am Kopf der Outlanes: was nach oben geschossen wird (Kick-
       back, Flipper), fällt ins Feld statt in dieselbe Gasse zurück. */
    if (k.vy < -120 && k.y < 606 && k.y > 518) {
      if (k.x < LANES.links.wand[0]) { k.vx = Math.max(k.vx, 260); k.vy = Math.min(k.vy, -240); }
      else if (k.x > LANES.rechts.wand[0]) { k.vx = Math.min(k.vx, -260); k.vy = Math.min(k.vy, -240); }
    }

    schleudern(sp, k, dt);

    /* Kanalauslauf: die Kugel schießt im Kanal nach oben und verlässt ihn durch
       die Öffnung in der Wand, knapp unter der Decke. Der Fangbereich liegt
       unter dem Kanaldach, damit die Kugel nicht erst dagegen stößt. */
    if (sp.auslaufSperre > 0) sp.auslaufSperre -= dt;
    if (sp.auslaufSperre <= 0 && k.ebene === "haupt" && k.vy < -120 &&
        k.x > KANAL.x1 + 4 && k.y < KANAL.tor && k.y > KANAL.tor - 52) {
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

    /* Letzte Sicherung gegen jeden Rechen- oder Geometrieausreißer: im Feld
       bleibt die Kugel im Feld. (Der Kanal wird oben schon eigens geführt.) */
    if (k.x < KANAL.x1) {
      if (k.x < 30 + RK) { k.x = 30 + RK; if (k.vx < 0) k.vx = -k.vx * 0.45; }
      if (k.y < 322 + RK) { k.y = 322 + RK; if (k.vy < 0) k.vy = -k.vy * 0.45; }
    }

    var v = Math.hypot(k.vx, k.vy);
    if (v > MAXV) { k.vx *= MAXV / v; k.vy *= MAXV / v; }

    /* Auslass */
    if (k.y > H - 2 || k.y < -260 || k.x < -80 || k.x > W + 80) kugelVerloren(sp, k);
  }

  function mikroschrittDeck(sp, k, dt) {
    var i;
    k.vorX = k.x; k.vorY = k.y;
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

  /* --------------------------------------------------------- Kugelrettung
     Eine frei im Feld liegende Kugel, die zwei Sekunden lang praktisch
     unbeweglich war, bekommt zuerst einen kontrollierten Impuls. Bleibt sie
     danach wieder liegen, wird sie an eine freie Stelle im Feld gesetzt.
     Punktestand, Kugeln und Spielzustand bleiben unverändert. Ein Ball auf
     einer Bahn, im Schloss, im Kanal oder auf einem gedrückten Flipper ist ein
     vorgesehener Ruhezustand und wird nicht angefasst. */
  var RETTUNG_STILL = 2.0;
  var RETTUNG_IMPULS = 1.2;
  var rettungAktiv = true;                 // nur für Prüfläufe abschaltbar

  function flipperHaelt(sp, k) {
    for (var i = 0; i < sp.flipper.length; i++) {
      var f = sp.flipper[i];
      if (!f.an) continue;
      var ex = f.x + Math.cos(f.winkel) * f.len;
      var ey = f.y + Math.sin(f.winkel) * f.len;
      if (abstandSegment(k.x, k.y, f.x, f.y, ex, ey) < RK + FLIPPER_R + 3) return true;
    }
    return false;
  }

  /* Kleinster Abstand zu allen festen Bauteilen – negativ heißt: überlappt. */
  function freiheit(t, x, y) {
    var i, d = 9999;
    function naeher(w) { if (w < d) d = w; }
    for (i = 0; i < t.waende.length; i++) {
      var w = t.waende[i];
      naeher(abstandSegment(x, y, w.x1, w.y1, w.x2, w.y2) - RK);
    }
    for (i = 0; i < t.bumper.length; i++) {
      var b = t.bumper[i];
      naeher(Math.hypot(x - b.x, y - b.y) - (b.r + RK));
    }
    for (i = 0; i < t.zieleOben.length; i++) {
      var z = t.zieleOben[i];
      naeher(Math.hypot(x - (z.x + z.w / 2), y - (z.y + z.h / 2)) - (Math.max(z.w, z.h) / 2 + RK));
    }
    for (i = 0; i < t.fallen.length; i++) {
      var f = t.fallen[i];
      naeher(Math.hypot(x - (f.x + f.w / 2), y - (f.y + f.h / 2)) - (Math.max(f.w, f.h) / 2 + RK));
    }
    for (i = 0; i < t.schleudern.length; i++) {
      var s = t.schleudern[i];
      naeher(abstandSegment(x, y, s.x1, s.y1, s.x2, s.y2) - (RK + 6));
    }
    naeher(Math.hypot(x - t.spinner.x, y - t.spinner.y) - (RK + 16));
    naeher(Math.hypot(x - t.scoop.x, y - t.scoop.y) - (t.scoop.r + RK));
    naeher(Math.hypot(x - t.schloss.x, y - t.schloss.y) - (t.schloss.r + RK));
    naeher(Math.hypot(x - t.magnet.x, y - t.magnet.y) - (t.magnet.r + RK));
    return d;
  }

  /* Freiesten Platz im Feld einmal je Tisch bestimmen; weiter unten ist es
     günstiger, dort wartet der Flipper statt eines Bumpers. */
  function sichererPunkt(t) {
    if (t.rettungspunkt) return t.rettungspunkt;
    var best = { x: 202, y: 560 }, wert = -999;
    for (var y = 360; y <= 660; y += 20) {
      for (var x = 60; x <= 340; x += 20) {
        var w = freiheit(t, x, y) + (y - 360) * 0.06;
        if (w > wert) { wert = w; best = { x: x, y: y }; }
      }
    }
    t.rettungspunkt = best;
    return best;
  }

  function rettung(sp, k, h) {
    if (!rettungAktiv) { k.anker = null; return; }
    if (k.zustand !== "feld") { k.anker = null; return; }
    if (k.sperre > 0) { k.anker = null; return; }
    if (k.ebene === "haupt" && k.x > KANAL.x1) { k.anker = null; return; }
    if (flipperHaelt(sp, k)) { k.anker = null; k.rettFrei = 0; return; }
    if (!k.anker) { k.anker = { x: k.x, y: k.y, t: 0 }; return; }
    if (Math.hypot(k.x - k.anker.x, k.y - k.anker.y) > 7) {
      k.anker.x = k.x; k.anker.y = k.y; k.anker.t = 0;
      k.rettFrei += h;
      if (k.rettFrei > 3) k.rettStufe = 0;
      return;
    }
    k.anker.t += h;
    var grenze = k.rettStufe === 0 ? RETTUNG_STILL : RETTUNG_IMPULS;
    if (k.anker.t < grenze) return;
    if (k.rettStufe === 0) {
      /* Erster Versuch: ein kontrollierter Impuls schräg nach oben. */
      k.rettStufe = 1;
      k.rettFrei = 0;
      k.anker.t = 0; k.anker.x = k.x; k.anker.y = k.y;
      k.vx = (k.x < W / 2 ? 1 : -1) * (300 + zufall() * 120);
      k.vy = -(500 + zufall() * 140);
      sp.ereignisse.push({ art: "stups", x: k.x, y: k.y, wert: 1 });
      sp.meldung = "Kugel befreit.";
      return;
    }
    /* Zweiter Versuch: an eine freie Stelle setzen. Kein Leben, keine Punkte. */
    var p = sichererPunkt(sp.tisch);
    k.x = p.x; k.y = p.y;
    k.vx = 0; k.vy = 110;
    k.ebene = "haupt"; k.bahn = null;
    k.sperre = 0.4; k.rettStufe = 0; k.rettFrei = 0;
    k.anker = { x: k.x, y: k.y, t: 0 };
    sp.ereignisse.push({ art: "rettung", x: k.x, y: k.y, wert: 0 });
    sp.meldung = "Kugel zurückgesetzt.";
  }

  function schritt(sp, dt) {
    if (!(dt > 0)) return;
    if (dt > 1 / 30) dt = 1 / 30;
    flipperDrehen(sp, dt);
    if (sp.bonusZeit > 0) sp.bonusZeit = Math.max(0, sp.bonusZeit - dt);
    deckTakt(sp, dt);
    zielTakt(sp, dt);
    if (sp.saverZeit > 0) sp.saverZeit -= dt;

    if (sp.zustand === "bereit") {
      if (sp.abschuss.laedt) sp.abschuss.kraft = Math.min(1, sp.abschuss.kraft + dt * 1.15);
      return;
    }
    if (sp.zustand !== "spiel") return;

    /* Schrittweite so klein, dass die Kugel auch bei Höchsttempo (1500 px/s)
       keine Wand überspringen kann: höchstens 4 px je Mikroschritt. */
    var i, k, n = Math.max(1, Math.min(40, Math.ceil(dt * MAXV / 4)));
    var h = dt / n;
    for (var s = 0; s < n; s++) {
      for (i = sp.kugeln.length - 1; i >= 0; i--) {
        k = sp.kugeln[i];
        if (!k || k.zustand === "weg" || k.zustand === "schloss") continue;
        if (k.sperre > 0) k.sperre -= h;
        mikroschritt(sp, k, h);
        rettung(sp, k, h);
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
    imSpiel: imSpiel, multiWert: multiWert, neuerBall: neueKugel, partieStarten: partieStarten,
    rettungAn: function (an) { rettungAktiv = !!an; }, sichererPunkt: sichererPunkt,
    freiheit: function (t, x, y) { return freiheit(t, x, y); }
  };
  root.FL = FL;
  if (typeof module === "object" && module.exports) module.exports = FL;
})(typeof window !== "undefined" ? window : globalThis);
