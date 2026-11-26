/* Rotlicht – Offline-Ertrag und Pausen prüfen (nur zur Kontrolle, nicht im Repo).
   Aufruf: node rotlicht_offline.js */
"use strict";

var RL = require(require("path").join(__dirname, "..", "..", "spiele", "rotlicht", "rotlicht-kern.js"));
var T0 = 1700000000000;

function kandidaten(s) {
  var w = RL.werte(s);
  var liste = [];
  function dazu(art, id, kosten, zuwachs) {
    if (!(kosten > 0) || s.geld + 1e-9 < kosten) return;
    liste.push({ art: art, id: id, kosten: kosten, nutzen: zuwachs / kosten });
  }
  var ns = RL.naechsteStufe(s);
  if (ns) dazu("stufe", null, ns.kosten, (ns.faktor / w.stufenFaktor - 1) * w.proSekunde);
  for (var i = 0; i < RL.PERSONAL.length; i++) {
    var p = RL.PERSONAL[i];
    if (s.stufe < p.ab) continue;
    dazu("personal", p.id, RL.personalKosten(s, p, 1), p.stunde * w.gesamtFaktor / RL.STUNDE_SEK);
  }
  for (i = 0; i < RL.UPGRADES.length; i++) {
    var u = RL.UPGRADES[i];
    if (s.upgrades[u.id] || s.stufe < u.ab) continue;
    if (u.art === "einnahme") dazu("upgrade", u.id, u.preis, (u.faktor - 1) * w.proSekunde);
    else if (u.art === "tempo") dazu("upgrade", u.id, u.preis, (1 - u.faktor) * w.proSekunde * 0.25);
    else dazu("upgrade", u.id, u.preis, 0.05 * w.proSekunde);
  }
  dazu("standort", null, RL.standortKosten(s), 0.22 * w.proSekunde);
  return liste.sort(function (a, b) { return b.nutzen - a.nutzen; });
}

function kaufe(s, t, k) {
  if (!k) return;
  if (k.art === "stufe") RL.stufeKaufen(s, t);
  else if (k.art === "personal") RL.personalKaufen(s, k.id, 1, t);
  else if (k.art === "upgrade") RL.upgradeKaufen(s, k.id, t);
  else RL.standortKaufen(s, t);
}

function tick(s, t, mitKlicks) {
  RL.abrechnen(s, t);
  RL.ereignis(s, t);
  if (mitKlicks) {
    if (s.kondome < 1 && s.geld >= 5) RL.kondomeKaufen(s, 1, t);
    RL.klick(s, "hand", t);
    RL.klick(s, "extra", t);
  }
  var k = kandidaten(s)[0];
  if (k && s.geld >= k.kosten) kaufe(s, t, k);
}

/* 1. Was wird nach einer Abwesenheit gutgeschrieben? */
var s = RL.neuerStand(T0);
for (var sek = 1; sek <= 20 * 60; sek++) tick(s, T0 + sek * 1000, true);
console.log("Nach 20 Minuten: Stufe " + s.stufe + ", " + RL.formatGeld(RL.werte(s).proStunde) + "/h, " +
  RL.formatGeld(s.geld) + " auf dem Konto");

[30, 120, 480, 1440].forEach(function (m) {
  var kopie = RL.laden(RL.speicherText(s));
  var zu = RL.abrechnen(kopie, T0 + (20 * 60 + m * 60) * 1000);
  var bericht = kopie.bericht || {};
  console.log("  weg " + String(m).padStart(5) + " min → gutgeschrieben " + RL.formatGeld(zu) +
    " (vergütet: " + Math.round((bericht.sek || 0) / 60) + " min" + (bericht.gekappt ? ", gekappt" : "") + ")");
});

/* 2. Verkürzt eine Pause das Spiel? Gleiche Spielweise, Pause zählt mit. */
function durchlauf(pauseNach, pauseMinuten) {
  var st = RL.neuerStand(T0);
  var sek = 0;
  while (sek < 8 * 3600) {
    sek++;
    if (pauseNach && sek === pauseNach) {
      RL.abrechnen(st, T0 + (sek + pauseMinuten * 60) * 1000);
      st.bericht = null;
      sek += pauseMinuten * 60;
      continue;
    }
    tick(st, T0 + sek * 1000, true);
    if (st.stufe === 10) return sek;
  }
  return null;
}

var ohne = durchlauf(0, 0);
console.log("\nOhne Pause: Stufe 10 nach " + RL.formatDauer(ohne));
[30, 120, 480].forEach(function (m) {
  var mit = durchlauf(20 * 60, m);
  console.log("Pause " + String(m).padStart(4) + " min nach 20 min: Stufe 10 nach " +
    (mit ? RL.formatDauer(mit) : "> 8 h") + " (Pause zählt mit)");
});
