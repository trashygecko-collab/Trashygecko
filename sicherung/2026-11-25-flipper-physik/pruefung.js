/* ==========================================================================
   TrashyGecko – Flipper · Physikprüfung
   Simuliert alle vier Tische ohne Browser und zählt:
     - fehlerhafte Abschüsse (Kugel hinter der Feder, im Kanal, verloren)
     - Kugelfallen (unter 2 s Bewegung < 10 px)
     - Wanddurchbrüche (Kugel verlässt das Spielfeld)
     - Kugelrettung (Impuls und Zurücksetzen)
   Aufruf:  node pruefung.js [--kurz]
   ========================================================================== */
"use strict";

var FL = require("../../spiele/flipper/flipper-kern.js");

var DT = 1 / 240;
var RK = FL.RK, W = FL.W, H = FL.H;
var KURZ = process.argv.indexOf("--kurz") >= 0;

function neuesSpiel(nr) {
  var sp = FL.neuesSpiel(nr);
  FL.partieStarten(sp);
  return sp;
}

function setzeKugel(sp, x, y, vx, vy, ebene) {
  var k = FL.neuerBall(x, y, vx, vy, ebene || "haupt");
  sp.kugeln.push(k);
  sp.zustand = "spiel";
  return k;
}

/* Ein Bildschritt mit Sammelstelle für Auffälligkeiten. */
function schritt(sp, beob, i) {
  FL.schritt(sp, DT);
  for (var n = 0; n < sp.kugeln.length; n++) {
    var k = sp.kugeln[n];
    if (!k || k.zustand === "weg") continue;
    if (k.zustand !== "feld") continue;
    if (k.ebene === "deck") {
      if (k.x < 78 + RK - 2 || k.x > 356 - RK + 2 || k.y < 70 + RK - 2 || k.y > 320 + 2) {
        beob.deckAus.push({ x: Math.round(k.x), y: Math.round(k.y), i: i });
      }
    } else if (k.x > 372) {
      /* im Abschusskanal: nur der Schacht ist erlaubt */
      if (k.x < 372 + RK - 1 || k.x > 420 - RK + 1 || k.y > 749) beob.kanalAus.push({ x: Math.round(k.x), y: Math.round(k.y) });
    } else if (k.x < 22 || k.x > 366 || k.y < 325) {
      beob.wandAus.push({ x: Math.round(k.x), y: Math.round(k.y), i: i });
    }
  }
}

/* Kugelfalle: 2 s lang weniger als 10 px Ortsveränderung (ohne Kanal). */
function fallePruefen(k, stand, i) {
  if (!stand.anker) { stand.anker = { x: k.x, y: k.y, i: i }; return false; }
  if (Math.hypot(k.x - stand.anker.x, k.y - stand.anker.y) > 10) {
    stand.anker = { x: k.x, y: k.y, i: i };
    return false;
  }
  return (i - stand.anker.i) * DT >= 2;
}

function leer() {
  return { fangen: [], wandAus: [], deckAus: [], kanalAus: [], hinterFeder: 0 };
}

function tischName(nr) { return FL.TISCHE[nr].name; }

FL.rettungAn(false);          // Geometrie ohne Netz prüfen

/* ------------------------------------------------------------- 1. Abschuss */
function abschuss() {
  var ergebnis = [];
  for (var nr = 0; nr < 4; nr++) {
    var r = { id: tischName(nr), n: 0, feld: 0, verloren: 0, kanal: 0, feder: 0, wand: 0, rest: [] };
    [0.28, 0.55, 1].forEach(function (kraft) {
      for (var w = 0; w < (KURZ ? 5 : 15); w++) {
        var sp = neuesSpiel(nr);
        FL.abschuss(sp, kraft);
        var k = sp.kugeln[sp.kugeln.length - 1];
        if (!k) { r.rest.push("kein Schuss"); return; }
        r.n++;
        var beob = leer();
        var hatFeld = false, feder = false, zurueck = false, raus = false;
        for (var i = 0; i < 240 * 4 && sp.kugeln.indexOf(k) >= 0; i++) {
          schritt(sp, beob, i);
          if (k.ebene === "haupt" && k.x > 372) { if (k.y > 748) feder = true; }
          else if (!hatFeld) hatFeld = true;
          if (hatFeld && k.ebene === "haupt" && k.x > 372 && k.y > 600) zurueck = true;
          if (beob.kanalAus.length) raus = true;
        }
        if (beob.wandAus.length) r.wand++;
        if (feder) r.feder++;
        if (zurueck) r.kanal++;
        if (raus) r.rest.push("aus dem Kanal");
        if (hatFeld) r.feld++;
        else r.verloren++;
      }
    });
    ergebnis.push(r);
  }
  return ergebnis;
}

/* ------------------------------------------------- 2. Zufällige Freiflüge */
function freiflug(anzahl, sekunden) {
  var ergebnis = [];
  for (var nr = 0; nr < 4; nr++) {
    var r = { id: tischName(nr), n: 0, fallen: [], wandAus: 0, deckAus: 0, kanalAus: 0 };
    var zufall = 12345 + nr * 7919;
    function rnd() { zufall = (zufall * 1103515245 + 12345) % 2147483648; return zufall / 2147483648; }
    for (var t = 0; t < anzahl; t++) {
      var sp = neuesSpiel(nr);
      var x, y, versuche = 0;
      /* nur freie Startplätze: die Kugel wird im Spiel nie in ein Bauteil gesetzt */
      do { x = 45 + rnd() * 315; y = 345 + rnd() * 330; versuche++; }
      while (FL.freiheit(sp.tisch, x, y) < 8 && versuche < 80);
      var winkel = rnd() * Math.PI * 2, tempo = 120 + rnd() * 1200;
      var k = setzeKugel(sp, x, y, Math.cos(winkel) * tempo, Math.sin(winkel) * tempo);
      var beob = leer(), stand = { anker: null };
      r.n++;
      for (var i = 0; i < 240 * sekunden && sp.kugeln.indexOf(k) >= 0; i++) {
        schritt(sp, beob, i);
        if (k.x < 372 && fallePruefen(k, stand, i)) {
          r.fallen.push({ x: Math.round(k.x), y: Math.round(k.y), ebene: k.ebene });
          break;
        }
      }
      r.wandAus += beob.wandAus.length;
      r.deckAus += beob.deckAus.length;
      r.kanalAus += beob.kanalAus.length;
    }
    ergebnis.push(r);
  }
  return ergebnis;
}

/* ------------------------------------------- 3. Gerichtete Höchsttempo-Schüsse */
function durchbruch() {
  var stellen = [[202, 400], [100, 360], [320, 360], [70, 500], [330, 520], [202, 470], [150, 640], [280, 640]];
  var ergebnis = [];
  for (var nr = 0; nr < 4; nr++) {
    var r = { id: tischName(nr), n: 0, wandAus: [] };
    for (var s = 0; s < stellen.length; s++) {
      for (var d = 0; d < 12; d++) {
        var sp = neuesSpiel(nr);
        var a = d / 12 * Math.PI * 2;
        var k = setzeKugel(sp, stellen[s][0], stellen[s][1], Math.cos(a) * FL.MAXV, Math.sin(a) * FL.MAXV);
        var beob = leer();
        for (var i = 0; i < 240 * 4 && sp.kugeln.indexOf(k) >= 0; i++) schritt(sp, beob, i);
        r.n++;
        if (beob.wandAus.length) r.wandAus.push({ von: stellen[s], a: Math.round(a * 57), wo: beob.wandAus[0] });
      }
    }
    ergebnis.push(r);
  }
  return ergebnis;
}

/* --------------------------------------------------------- 4. Oberdeck */
function oberdeck(anzahl) {
  var ergebnis = [];
  for (var nr = 0; nr < 4; nr++) {
    var r = { id: tischName(nr), n: 0, fallen: [], aus: 0, haengt: 0 };
    var zufall = 777 + nr * 31;
    function rnd() { zufall = (zufall * 1103515245 + 12345) % 2147483648; return zufall / 2147483648; }
    for (var t = 0; t < anzahl; t++) {
      var sp = neuesSpiel(nr);
      var x = 95 + rnd() * 245, y = 95 + rnd() * 215;
      var winkel = rnd() * Math.PI * 2, tempo = 60 + rnd() * 700;
      var k = setzeKugel(sp, x, y, Math.cos(winkel) * tempo, Math.sin(winkel) * tempo, "deck");
      var beob = leer(), stand = { anker: null };
      r.n++;
      var i;
      for (i = 0; i < 240 * 14 && sp.kugeln.indexOf(k) >= 0; i++) {
        schritt(sp, beob, i);
        if (k.ebene === "haupt") break;
        if (fallePruefen(k, stand, i)) { r.fallen.push({ x: Math.round(k.x), y: Math.round(k.y) }); break; }
      }
      r.aus += beob.deckAus.length;
      if (i >= 240 * 14) r.haengt++;
    }
    ergebnis.push(r);
  }
  return ergebnis;
}

/* --------------------------------------------------------- 5. Kugelrettung */
function rettungTest() {
  var ergebnis = [];
  for (var nr = 0; nr < 4; nr++) {
    var sp = neuesSpiel(nr);
    sp.tisch.grav = 0;                       // Kugel bleibt sonst nie liegen
    var start = FL.sichererPunkt(sp.tisch);
    var k = setzeKugel(sp, start.x, start.y, 0, 0);
    sp.punkte = 4711; sp.kugelnUebrig = 2;
    var punkteVorher = sp.punkte, kugelnVorher = sp.kugelnUebrig;
    var impulsiert = false, gesetzt = false, tImpuls = 0, tSetzt = 0;
    var punkteBeimSetzen = punkteVorher, kugelnBeimSetzen = kugelnVorher;
    FL.rettungAn(true);
    for (var i = 1; i <= 240 * 6; i++) {
      FL.schritt(sp, DT);
      FL.ereignisseHolen(sp).forEach(function (e) {
        if (e.art === "stups") { impulsiert = true; tImpuls = i * DT; }
        if (e.art === "rettung") {
          gesetzt = true; tSetzt = i * DT;
          punkteBeimSetzen = sp.punkte; kugelnBeimSetzen = sp.kugelnUebrig;
        }
      });
      /* Nach dem Impuls sofort wieder festhalten: der zweite Versuch muss greifen. */
      if (impulsiert) { k.vx = 0; k.vy = 0; }
      if (gesetzt) break;          // danach läuft wieder normales Spiel
    }
    FL.rettungAn(false);
    ergebnis.push({
      id: tischName(nr), impulsiert: impulsiert, gesetzt: gesetzt, tImpuls: tImpuls, tSetzt: tSetzt,
      punkteGleich: punkteBeimSetzen === punkteVorher, kugelnGleich: kugelnBeimSetzen === kugelnVorher,
      lebt: sp.kugeln.indexOf(k) >= 0, zustand: k.zustand, frei: sp.tisch.rettungspunkt
    });
  }
  return ergebnis;
}

/* ------------------------------------------------------------------ Bericht */
function zeile(s) { console.log(s); }
function prozent(a, b) { return b ? (100 * a / b).toFixed(1) + " %" : "–"; }

var A = abschuss();
zeile("== 1. Abschuss (Kraft 0,28 / 0,55 / 1,0) ==");
A.forEach(function (r) {
  zeile("  " + r.id + ": " + r.n + " Schüsse · im Feld " + r.feld + " (" + prozent(r.feld, r.n) + ")" +
    " · verloren " + r.verloren + " · zurück im Kanal " + r.kanal + " · hinter der Feder " + r.feder +
    " · Wandaus " + r.wand + (r.rest.length ? " · " + r.rest.join(", ") : ""));
});

var F = freiflug(KURZ ? 40 : 150, 20);
zeile("== 2. Freiflug (Zufallslage, bis 1320 px/s, 20 s) ==");
F.forEach(function (r) {
  var gruppen = {};
  r.fallen.forEach(function (f) {
    var key = f.ebene + " " + (Math.round(f.x / 20) * 20) + "/" + (Math.round(f.y / 20) * 20);
    gruppen[key] = (gruppen[key] || 0) + 1;
  });
  var top = Object.keys(gruppen).sort(function (a, b) { return gruppen[b] - gruppen[a]; }).slice(0, 6)
    .map(function (key) { return key + "×" + gruppen[key]; }).join(", ");
  zeile("  " + r.id + ": " + r.n + " Kugeln · " + r.fallen.length + " Fallen (" + prozent(r.fallen.length, r.n) + ")" +
    " · Wandaus " + r.wandAus + " · Kanalaus " + r.kanalAus + " · Deckaus " + r.deckAus + (top ? " · " + top : ""));
});

var D = durchbruch();
zeile("== 3. Höchsttempo gegen Wände (1500 px/s, 96 Schüsse je Tisch) ==");
D.forEach(function (r) {
  zeile("  " + r.id + ": " + (r.wandAus.length ? r.wandAus.length + " Durchbrüche · " +
    r.wandAus.slice(0, 3).map(function (w) { return "von " + w.von + " bei " + w.a + "° nach " + w.wo.x + "/" + w.wo.y; }).join(" | ")
    : "keine Durchbrüche"));
});

var O = oberdeck(KURZ ? 20 : 60);
zeile("== 4. Oberdeck (Zufallslage, 14 s) ==");
O.forEach(function (r) {
  var gruppen = {};
  r.fallen.forEach(function (f) {
    var key = Math.round(f.x / 20) * 20 + "/" + Math.round(f.y / 20) * 20;
    gruppen[key] = (gruppen[key] || 0) + 1;
  });
  var top = Object.keys(gruppen).sort(function (a, b) { return gruppen[b] - gruppen[a]; }).slice(0, 5)
    .map(function (key) { return key + "×" + gruppen[key]; }).join(", ");
  zeile("  " + r.id + ": " + r.n + " Kugeln · " + r.fallen.length + " Fallen · " + r.haengt + " ohne Ausgang nach 14 s" +
    " · Randaus " + r.aus + (top ? " · " + top : ""));
});

var R = rettungTest();
zeile("== 5. Kugelrettung (Kugel 2 s festgehalten) ==");
R.forEach(function (r) {
  zeile("  " + r.id + ": Impuls " + (r.impulsiert ? "nach " + r.tImpuls.toFixed(2) + " s" : "FEHLT") +
    " · Zurücksetzen " + (r.gesetzt ? "nach " + (r.tSetzt - r.tImpuls).toFixed(2) + " s" : "FEHLT") +
    " · Punkte " + (r.punkteGleich ? "unverändert" : "VERÄNDERT") +
    " · Kugeln " + (r.kugelnGleich ? "unverändert" : "VERÄNDERT") +
    " · Ziel " + (r.frei ? r.frei.x + "/" + r.frei.y : "–"));
});

var fehler = 0;
A.forEach(function (r) { fehler += r.feder + r.kanal + r.wand + (r.n - r.feld - r.verloren) + r.rest.length; });
F.forEach(function (r) { fehler += r.fallen.length + r.wandAus + r.deckAus + r.kanalAus; });
D.forEach(function (r) { fehler += r.wandAus.length; });
O.forEach(function (r) { fehler += r.fallen.length + r.aus + r.haengt; });
R.forEach(function (r) {
  fehler += (r.impulsiert ? 0 : 1) + (r.gesetzt ? 0 : 1) + (r.punkteGleich ? 0 : 1) + (r.kugelnGleich ? 0 : 1) + (r.lebt ? 0 : 1);
});
zeile(fehler === 0 ? "ERGEBNIS: keine Auffälligkeiten." : "ERGEBNIS: " + fehler + " Auffälligkeiten.");
