/* Rotlicht – Balance-Simulator (nur zum Abstimmen, nicht im Repo).
   Aufruf: node rotlicht_tune.js [-v]
           BAL='{"lohn":0.5}' node rotlicht_tune.js

   Drei Spielweisen, Bild für Bild (1 s Raster):
     aktiv        klickt jede freie Sekunde, kauft sofort das Beste
     normal       schaut alle 4 s nach, klickt dabei, kauft meist das Beste
     vorsichtig   schaut alle 45 s nach, hält Rücklage, mag Personal

   „Regulärer Abschluss“ = Stufe 10 erreicht und jede Verbesserung gekauft. */
"use strict";

var RL = require("path").join ? require(require("path").join(__dirname, "..", "..", "spiele", "rotlicht", "rotlicht-kern.js")) : null;

var T0 = 1700000000000;
var MAX_H = 7;

/* Tabellen-Anpassungen für Suchläufe (verändern die echten Spieltabellen). */
if (false) {
  var t = JSON.parse(process.env.TAB);
  if (t.stageF) RL.STUFEN.forEach(function (s) { s.faktor = 1 + (s.faktor - 1) * t.stageF; });
  if (t.stageK) RL.STUFEN.forEach(function (s) { s.kosten = s.kosten * t.stageK; });
  if (t.upF) RL.UPGRADES.forEach(function (u) { if (u.faktor) u.faktor = 1 + (u.faktor - 1) * t.upF; });
  if (t.upK) RL.UPGRADES.forEach(function (u) { u.preis = u.preis * t.upK; });
  if (t.wage) RL.PERSONAL.forEach(function (p) { p.stunde = p.stunde * t.wage; });
  if (t.price) RL.PERSONAL.forEach(function (p) { p.preis = p.preis * t.price; });
  if (Array.isArray(t.preise)) RL.PERSONAL.forEach(function (p, i) { if (t.preise[i]) p.preis = Math.round(p.preis * t.preise[i]); });
  if (t.wuchs) RL.PERSONAL.forEach(function (p) { p.wuchs = p.wuchs + t.wuchs; });
  if (t.stufen) {
    var r = t.stufen;
    RL.STUFEN.forEach(function (st, i) {
      if (i === 0) { st.faktor = r.f0; return; }
      st.kosten = Math.round(r.start * Math.pow(r.ratio, i - 1) * 100) / 100;
      st.faktor = Math.round((r.f0 + r.df * i) * 1000) / 1000;
    });
  }
}

/* Kaufkandidaten: Nutzen (Einnahmen je Sekunde) je Euro. */
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
  return liste;
}

function kaufe(s, t, k) {
  var r;
  if (k.art === "stufe") r = RL.stufeKaufen(s, t);
  else if (k.art === "personal") r = RL.personalKaufen(s, k.id, 1, t);
  else if (k.art === "upgrade") r = RL.upgradeKaufen(s, k.id, t);
  else r = RL.standortKaufen(s, t);
  return !!(r && r.ok);
}

/* Eine Spielweise als Funktion. */
function spieler(o) {
  return function (s, t, sek) {
    if (sek % o.blick !== 0) return;
    if (s.kondome < 1 && s.geld >= 5) RL.kondomeKaufen(s, 1, t);
    if (sek % o.klickHand === 0) RL.klick(s, "hand", t);
    if (o.klickExtra && sek % o.klickExtra === 0) RL.klick(s, "extra", t);

    var l = kandidaten(s);
    if (!l.length) return;
    /* Stufe zuerst, wenn sie sich lohnt und bezahlt ist. */
    if (o.stufeZuerst) {
      var st = l.filter(function (x) { return x.art === "stufe"; })[0];
      if (st && s.geld >= st.kosten * (1 + o.ruecklage)) { kaufe(s, t, st); return; }
    }
    l = l.filter(function (x) { return s.geld >= x.kosten * (1 + o.ruecklage); });
    if (!l.length) return;
    l.sort(function (a, b) { return b.nutzen - a.nutzen; });
    var griff = l[0];
    if (o.zufall && Math.random() < o.zufall) griff = l[Math.min(l.length - 1, 1 + Math.floor(Math.random() * 2))];
    if (o.zufallPersonal && Math.random() < o.zufallPersonal) {
      var pers = l.filter(function (x) { return x.art === "personal"; });
      if (pers.length) griff = pers[0];
    }
    kaufe(s, t, griff);
  };
}

function lauf(name, fn, maxH) {
  var s = RL.neuerStand(T0);
  var t = T0, stufe10 = null, alles = null, protokoll = [];
  var letzteStufe = 1;
  var sekunden = Math.round((maxH || MAX_H) * 3600);
  for (var sek = 1; sek <= sekunden; sek++) {
    t = T0 + sek * 1000;
    RL.abrechnen(s, t);
    RL.ereignis(s, t);
    fn(s, t, sek);
    if (s.stufe !== letzteStufe) {
      letzteStufe = s.stufe;
      protokoll.push(RL.formatDauer(sek) + " → " + s.stufe);
    }
    if (stufe10 === null && s.stufe === 10) stufe10 = sek;
    if (alles === null && s.stufe === 10 && RL.anzahlUpgrades(s) === RL.UPGRADES.length) alles = sek;
  }
  return {
    name: name, stufe10: stufe10, alles: alles, protokoll: protokoll, stand: s,
    personal: RL.anzahlPersonal(s), standorte: s.standorte,
    upgrades: RL.anzahlUpgrades(s), erfolge: Object.keys(s.erfolge).length,
    stunden: RL.werte(s).proStunde
  };
}

var aktiv = spieler({ blick: 1, klickHand: 1, klickExtra: 1, stufeZuerst: true, ruecklage: 0, zufall: 0 });
var normal = spieler({ blick: 4, klickHand: 4, klickExtra: 8, stufeZuerst: true, ruecklage: 0.1, zufall: 0.25 });
var vorsichtig = spieler({ blick: 60, klickHand: 60, klickExtra: 120, stufeZuerst: false, ruecklage: 0.5, zufallPersonal: 0.5 });

function zeile(e) {
  return e.name.padEnd(12) +
    "Stufe 10: " + (e.stufe10 ? RL.formatDauer(e.stufe10) : "> " + MAX_H + " h").padEnd(11) +
    " komplett: " + (e.alles ? RL.formatDauer(e.alles) : "> " + MAX_H + " h").padEnd(11) +
    " (" + e.upgrades + "/" + RL.UPGRADES.length + " Upgrades, " +
    e.erfolge + "/" + RL.ERFOLGE.length + " Erfolge, " + e.personal + " Leute, " +
    e.standorte + " Standorte, " + RL.formatGeld(e.stunden) + "/h)";
}

var erg = [];
erg.push(lauf("aktiv", aktiv));
erg.push(lauf("normal", normal));
erg.push(lauf("vorsichtig", vorsichtig));
erg.forEach(function (e) { console.log(zeile(e)); });
erg.forEach(function (e) {
  var fehlt = RL.ERFOLGE.filter(function (a) { return !e.stand.erfolge[a.id]; }).map(function (a) { return a.id; });
  if (fehlt.length) console.log("   offen bei " + e.name + ": " + fehlt.join(", "));
});
if (process.argv[2] === "-v") {
  erg.forEach(function (e) { console.log("\n" + e.name + ": " + e.protokoll.join(" · ")); });
}
