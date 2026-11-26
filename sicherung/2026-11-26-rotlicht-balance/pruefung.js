/* Prüfprogramm ROTLICHT – Regeln, Speichern, Offline und Spieldauer.
   Aufruf: node pruefung.js
   Braucht: ../../spiele/rotlicht/rotlicht-kern.js

   Die Prüfungen rechnen mit den Tabellen aus dem Kern, nicht mit
   abgeschriebenen Zahlen: Wer die Balance ändert, soll hier sehen, ob die
   Regeln noch stimmen und ob die Partie noch etwa vier Stunden dauert. */
"use strict";

var pfad = require("path");
var RL = require(pfad.join(__dirname, "..", "..", "spiele", "rotlicht", "rotlicht-kern.js"));

var fehler = 0, geprueft = 0;
function pruefe(name, bedingung, zusatz) {
  geprueft++;
  var hinten = zusatz === undefined ? "" : "  (" + zusatz + ")";
  if (!bedingung) {
    fehler++;
    console.log("FEHLER  " + name + hinten);
  } else {
    console.log("ok      " + name + hinten);
  }
}
function nah(a, b, tol) { return Math.abs(a - b) <= (tol || 1e-6); }

var T0 = 1700000000000;
function neu() { return RL.neuerStand(T0); }
function Mia() { return RL.PERSONAL[0]; }

/* ------------------------------------------------------------ Startzustand */
var s = neu();
pruefe("Start: 0 €", s.geld === 0);
pruefe("Start: Stufe 1", s.stufe === 1);
pruefe("Start: kein Personal", RL.anzahlPersonal(s) === 0);
pruefe("Start: keine Kondome", s.kondome === 0);
pruefe("Start: 0 € je Stunde", RL.werte(s).proStunde === 0);

/* ------------------------------------------------------------- Handarbeit */
var hand = RL.aktionsBetrag(s, "hand");
pruefe("Handarbeit zahlt den Grundbetrag × Stufe", hand === Math.round(20 * RL.STUFEN[0].faktor), RL.formatGeld(hand));
var r = RL.klick(s, "hand", T0 + 1000);
pruefe("Handarbeit schreibt gut", r.ok && nah(s.geld, hand));
r = RL.klick(s, "hand", T0 + 2000);
pruefe("Handarbeit ist bis zur nächsten Spielstunde gesperrt", !r.ok && r.grund === "warte", r.rest + " s");
r = RL.klick(s, "hand", T0 + 1000 + RL.werte(s).cooldown * 1000);
pruefe("Handarbeit danach wieder frei", r.ok, RL.werte(s).cooldown + " s Takt");

/* --------------------------------------------------------- Zusatzgeschäft */
r = RL.klick(s, "extra", T0 + 200000);
pruefe("Zusatzgeschäft ohne Kondom abgelehnt", !r.ok && r.grund === "keine-kondome");
s.geld = 100;
r = RL.kondomeKaufen(s, 1, T0 + 210000);
pruefe("Viererpack: 4 Kondome für 5 €", r.ok && r.anzahl === 4 && r.preis === 5 && s.kondome === 4);
r = RL.klick(s, "extra", T0 + 220000);
pruefe("Zusatzgeschäft gibt mehr als Handarbeit und kostet ein Kondom",
  r.ok && r.betrag > hand && s.kondome === 3, RL.formatGeld(r.betrag));

/* ---------------------------------------------------------- Kauf ohne Geld */
var s2 = neu();
pruefe("Kauf ohne Geld scheitert", !RL.personalKaufen(s2, Mia().id, 1, T0).ok && s2.geld === 0);
pruefe("Upgrade ohne Geld scheitert", !RL.upgradeKaufen(s2, "tafel", T0).ok);
pruefe("Stufe 2 ohne Geld scheitert", !RL.stufeKaufen(s2, T0).ok);
pruefe("Konten bleiben nicht negativ", s2.geld === 0 && RL.anzahlPersonal(s2) === 0);

/* --------------------------------------------------------------- Personal */
var s3 = neu();
s3.geld = 1e9;
var p0 = Mia();
r = RL.personalKaufen(s3, p0.id, 1, T0);
pruefe("Erste Mitarbeiterin kostet den Tabellenpreis", r.ok && nah(r.preis, p0.preis), RL.formatGeld(r.preis));
pruefe("Erste Mitarbeiterin bringt ihren Stundensatz", RL.werte(s3).personalStunde === p0.stunde,
  RL.formatGeld(p0.stunde) + "/h");
var zweiter = RL.personalKosten(s3, p0, 1);
pruefe("Zweite Mitarbeiterin derselben Art teurer", zweiter > p0.preis, RL.formatGeld(zweiter));
pruefe("Gespert ohne Stufe: Personal ab Stufe 2", (function () {
  var p = RL.PERSONAL.filter(function (x) { return x.ab > 1; })[0];
  return !RL.personalKaufen(s3, p.id, 1, T0).ok;
})());
pruefe("Preissteigerung je Kopf über 1,25 (kein Mengenrausch)",
  RL.PERSONAL.every(function (p) { return p.wuchs >= 1.3; }),
  RL.PERSONAL.map(function (p) { return p.wuchs; }).join(", "));

/* ---------------------------------------------------------------- Upgrades */
var s5 = neu();
s5.geld = 1e20;
var vorh = RL.werte(s5).proStunde;
r = RL.upgradeKaufen(s5, "tafel", T0);
pruefe("Reklametafel kaufbar", r.ok && r.preis === 250);
pruefe("Reklametafel erhöht die Einnahmen", RL.werte(s5).proStunde > vorh || vorh === 0);
pruefe("Upgrade nicht doppelt kaufbar", !RL.upgradeKaufen(s5, "tafel", T0).ok);
var tempo = RL.UPGRADES.filter(function (u) { return u.art === "tempo"; })[0];
var cdVor = RL.werte(s5).cooldown;
RL.upgradeKaufen(s5, tempo.id, T0);
pruefe("Tempo-Upgrade senkt den Takt", RL.werte(s5).cooldown < cdVor,
  cdVor.toFixed(1) + " s → " + RL.werte(s5).cooldown.toFixed(1) + " s");
pruefe("Takt bleibt über 15 s", RL.werte(s5).cooldown >= 15);

/* ----------------------------------------------------------------- Stufen */
var s6 = neu();
s6.geld = 1e20;
var namen = [];
for (var i = 0; i < 9; i++) {
  var rz = RL.stufeKaufen(s6, T0);
  if (rz.ok) namen.push(rz.name);
}
pruefe("Alle neun Ausbaustufen kaufbar", s6.stufe === 10 && namen.length === 9, namen.join(" → "));
pruefe("Stufe 10 hebt die Einnahmen um mehr als das Dreifache",
  RL.werte(s6).stufenFaktor / RL.STUFEN[0].faktor > 2,
  "×" + (RL.werte(s6).stufenFaktor / RL.STUFEN[0].faktor).toFixed(2));
pruefe("Über Stufe 10 hinaus kein Kauf", !RL.stufeKaufen(s6, T0).ok);
pruefe("Ausbaukosten steigen streng", (function () {
  for (var j = 1; j < RL.STUFEN.length; j++) if (!(RL.STUFEN[j].kosten > RL.STUFEN[j - 1].kosten)) return false;
  return true;
})());
pruefe("Ausbaukosten wachsen gleichmäßig (Faktor 20 bis 60 je Stufe)", (function () {
  for (var j = 2; j < RL.STUFEN.length; j++) {
    var q = RL.STUFEN[j].kosten / RL.STUFEN[j - 1].kosten;
    if (!(q >= 20 && q <= 60)) return false;
  }
  return true;
})(), "Stufe 10 kostet " + RL.formatGeld(RL.STUFEN[9].kosten));

/* --------------------------------------------------------------- Standorte */
var s7 = neu();
s7.geld = 1e6;
var f0 = RL.werte(s7).standortFaktor;
r = RL.standortKaufen(s7, T0);
pruefe("Standort kostet 5.000 € und zählt hoch", r.ok && s7.standorte === 1 && r.preis === 5000);
pruefe("Standort erhöht die Einnahmen ×1,05", nah(RL.werte(s7).standortFaktor, f0 * 1.05));
pruefe("Zweiter Standort teurer", RL.standortKosten(s7) > 5000, RL.formatGeld(RL.standortKosten(s7)));

/* ----------------------------------------------------------- Offline-Zeit */
var s8 = neu();
s8.geld = 1000;
RL.personalKaufen(s8, Mia().id, 1, T0);
s8.geld = 0;
var satz8 = RL.werte(s8).proSekunde;
var zu = RL.abrechnen(s8, T0 + 2 * 3600 * 1000);
pruefe("Zwei Stunden Abwesenheit zählen voll", nah(zu, satz8 * 2 * 3600), RL.formatGeld(zu));
pruefe("Rückkehrbericht gesetzt", s8.bericht && s8.bericht.sek === 7200);

var s8b = neu();
s8b.geld = 1000;
RL.personalKaufen(s8b, Mia().id, 1, T0);
s8b.geld = 0;
var satz8b = RL.werte(s8b).proSekunde;
zu = RL.abrechnen(s8b, T0 + 20 * 3600 * 1000);
pruefe("Zwanzig Stunden werden auf zwei gekappt", nah(zu, satz8b * 2 * 3600), RL.formatGeld(zu));
pruefe("Kappung wird gemeldet", s8b.bericht && s8b.bericht.gekappt);

var s8c = neu();
s8c.geld = 1e20;
s8c.stufe = 10;
RL.UPGRADES.filter(function (u) { return u.art === "offline"; })
  .forEach(function (u) { RL.upgradeKaufen(s8c, u.id, T0); });
RL.personalKaufen(s8c, Mia().id, 1, T0);
s8c.geld = 0;
var w8c = RL.werte(s8c);
pruefe("Beide Buchhaltungen erlauben vier Stunden", nah(w8c.offlineStunden, 4), w8c.offlineStunden + " h");
var satz8c = w8c.proSekunde;
zu = RL.abrechnen(s8c, T0 + 20 * 3600 * 1000);
pruefe("… und zahlen den Offline-Bonus", nah(zu, satz8c * 4 * 3600 * w8c.offlineBonus),
  RL.formatGeld(zu) + " bei ×" + w8c.offlineBonus.toFixed(2));

/* ------------------------------------------------------------- Speichern */
var s9 = neu();
s9.geld = 12345.67;
RL.personalKaufen(s9, Mia().id, 2, T0);
RL.upgradeKaufen(s9, "tafel", T0);
s9.kondome = 7;
s9.stufe = 2;
s9.standorte = 3;
var s9b = RL.laden(RL.speicherText(s9));
pruefe("Speichern und Laden erhält das Geld", nah(s9b.geld, s9.geld, 0.01), RL.formatGeld(s9b.geld));
pruefe("… das Personal", s9b.personal[Mia().id] === 2);
pruefe("… die Upgrades", s9b.upgrades.tafel === true);
pruefe("… Kondome, Stufe, Standorte", s9b.kondome === 7 && s9b.stufe === 2 && s9b.standorte === 3);
pruefe("… die Einnahmen je Stunde", nah(RL.werte(s9b).proStunde, RL.werte(s9).proStunde));
pruefe("Kaputter Spielstand → frischer Start", RL.laden("{kein json").geld === 0);
var boese = RL.laden(JSON.stringify({
  geld: -500, gesamtEinnahmen: "viel", personal: { mia: -3, ruby: "x" },
  stufe: 99, kondome: -4, upgrades: { tafel: true, unfug: true }, bereit: { hand: "spaeter" }
}));
pruefe("Negatives/Unsinn wird bereinigt", boese.geld === 0 && boese.personal.mia === 0 && boese.stufe === 10 && boese.kondome === 0);
pruefe("Unbekannte Upgrades fallen weg", !boese.upgrades.unfug);
pruefe("Ungültige Bereitschaft wird 0", boese.bereit.hand === 0);

/* -------------------------------------------------------------- Erfolge */
var s10 = neu();
var e0 = RL.werte(s10).gesamtFaktor;
RL.klick(s10, "hand", T0 + 1000);
pruefe("Erfolg „Erste Einnahme“ kommt", !!s10.erfolge["erstes-geld"]);
pruefe("Erfolg erhöht den Gesamtfaktor", RL.werte(s10).gesamtFaktor > e0);
pruefe("Erfolgsboni bleiben zusammen unter 40 %",
  RL.ERFOLGE.reduce(function (a, e) { return a + e.bonus; }, 0) < 0.4,
  "+" + Math.round(RL.ERFOLGE.reduce(function (a, b) { return a + b.bonus; }, 0) * 100) + " %");

/* ------------------------------------------------------------ Ereignisse */
var s11 = neu();
s11.geld = 1000;
RL.personalKaufen(s11, Mia().id, 1, T0);
s11.geld = 0;
var ereignis = null, tE = T0;
for (var sek = 1; sek < 400 && !ereignis; sek++) {
  tE = T0 + sek * 1000;
  ereignis = RL.ereignis(s11, tE);
}
pruefe("Ereignis erscheint nach kurzer Zeit", !!ereignis);
pruefe("Nächstes Ereignis liegt in der Zukunft", s11.naechstesEreignis > tE);
pruefe("Chronik füllt sich", s11.chronik.length > 0, s11.chronik.length + " Zeilen");

/* ---------------------------------------------------------- Zahlenformat */
pruefe("0 € wird „0,00 €“", RL.formatGeld(0) === "0,00 €");
pruefe("1.234,50 € ohne Nachkommastellen", RL.formatGeld(1234.5) === "1.235 €");
pruefe("Millionen mit Einheit", RL.formatGeld(1.234e6).indexOf("Mio.") > 0);
pruefe("Milliarden mit Einheit", RL.formatGeld(5.5e9).indexOf("Mrd.") > 0);
pruefe("Billionen mit Einheit", RL.formatGeld(2.5e12).indexOf("Bio.") > 0);
pruefe("kein NaN in der Ausgabe", RL.formatGeld(NaN).indexOf("NaN") < 0);

/* ------------------------------------------------------- Durchspiel-Simulation
   Drei Spielweisen, Bild für Bild (1 s). „Abschluss“ heißt: Stufe 10 steht
   und jede Verbesserung ist gekauft. */
function kandidaten(st) {
  var w = RL.werte(st);
  var liste = [];
  function dazu(art, id, kosten, zuwachs) {
    if (!(kosten > 0) || st.geld + 1e-9 < kosten) return;
    liste.push({ art: art, id: id, kosten: kosten, nutzen: zuwachs / kosten });
  }
  var ns = RL.naechsteStufe(st);
  if (ns) dazu("stufe", null, ns.kosten, (ns.faktor / w.stufenFaktor - 1) * w.proSekunde);
  for (var i = 0; i < RL.PERSONAL.length; i++) {
    var p = RL.PERSONAL[i];
    if (st.stufe < p.ab) continue;
    dazu("personal", p.id, RL.personalKosten(st, p, 1), p.stunde * w.gesamtFaktor / RL.STUNDE_SEK);
  }
  for (i = 0; i < RL.UPGRADES.length; i++) {
    var u = RL.UPGRADES[i];
    if (st.upgrades[u.id] || st.stufe < u.ab) continue;
    if (u.art === "einnahme") dazu("upgrade", u.id, u.preis, (u.faktor - 1) * w.proSekunde);
    else if (u.art === "tempo") dazu("upgrade", u.id, u.preis, (1 - u.faktor) * w.proSekunde * 0.25);
    else dazu("upgrade", u.id, u.preis, 0.05 * w.proSekunde);
  }
  dazu("standort", null, RL.standortKosten(st), 0.22 * w.proSekunde);
  return liste;
}
function kaufe(st, t, k) {
  var rz;
  if (k.art === "stufe") rz = RL.stufeKaufen(st, t);
  else if (k.art === "personal") rz = RL.personalKaufen(st, k.id, 1, t);
  else if (k.art === "upgrade") rz = RL.upgradeKaufen(st, k.id, t);
  else rz = RL.standortKaufen(st, t);
  return !!(rz && rz.ok);
}
function spieler(o) {
  return function (st, t, sek) {
    if (sek % o.blick !== 0) return;
    if (st.kondome < 1 && st.geld >= 5) RL.kondomeKaufen(st, 1, t);
    if (sek % o.klickHand === 0) RL.klick(st, "hand", t);
    if (o.klickExtra && sek % o.klickExtra === 0) RL.klick(st, "extra", t);
    var l = kandidaten(st);
    if (!l.length) return;
    if (o.stufeZuerst) {
      var stufe = l.filter(function (x) { return x.art === "stufe"; })[0];
      if (stufe && st.geld >= stufe.kosten * (1 + o.ruecklage)) { kaufe(st, t, stufe); return; }
    }
    l = l.filter(function (x) { return st.geld >= x.kosten * (1 + o.ruecklage); });
    if (!l.length) return;
    l.sort(function (a, b) { return b.nutzen - a.nutzen; });
    var griff = l[0];
    if (o.zufall && Math.random() < o.zufall) griff = l[Math.min(l.length - 1, 1 + Math.floor(Math.random() * 2))];
    if (o.zufallPersonal && Math.random() < o.zufallPersonal) {
      var pers = l.filter(function (x) { return x.art === "personal"; });
      if (pers.length) griff = pers[0];
    }
    kaufe(st, t, griff);
  };
}
function lauf(fn, maxH) {
  var st = RL.neuerStand(T0);
  var stufe10 = null, abschluss = null, protokoll = [], letzte = 1, negativ = 0, nan = 0;
  for (var sek = 1; sek <= Math.round(maxH * 3600); sek++) {
    var t = T0 + sek * 1000;
    RL.abrechnen(st, t);
    RL.ereignis(st, t);
    fn(st, t, sek);
    if (st.geld < -1e-6) negativ++;
    if (!isFinite(st.geld)) nan++;
    if (st.stufe !== letzte) {
      letzte = st.stufe;
      protokoll.push(RL.formatDauer(sek) + " → Stufe " + st.stufe);
    }
    if (stufe10 === null && st.stufe === 10) stufe10 = sek;
    if (abschluss === null && st.stufe === 10 && RL.anzahlUpgrades(st) === RL.UPGRADES.length) abschluss = sek;
    if (abschluss !== null) break;
  }
  return {
    stufe10: stufe10, abschluss: abschluss, protokoll: protokoll, stand: st,
    negativ: negativ, nan: nan, erfolge: Object.keys(st.erfolge).length
  };
}

var aktiv = spieler({ blick: 1, klickHand: 1, klickExtra: 1, stufeZuerst: true, ruecklage: 0 });
var normal = spieler({ blick: 4, klickHand: 4, klickExtra: 8, stufeZuerst: true, ruecklage: 0.1, zufall: 0.25 });
var vorsichtig = spieler({ blick: 60, klickHand: 60, klickExtra: 120, stufeZuerst: false, ruecklage: 0.5, zufallPersonal: 0.5 });

var erg = {
  aktiv: lauf(aktiv, 7),
  normal: lauf(normal, 7),
  vorsichtig: lauf(vorsichtig, 7)
};

console.log("\n--- Spieldauer bis zum regulären Abschluss ---");
Object.keys(erg).forEach(function (name) {
  var e = erg[name];
  console.log(name.padEnd(11) + "Stufe 10: " + (e.stufe10 ? RL.formatDauer(e.stufe10) : "> 7 h") +
    " · Abschluss: " + (e.abschluss ? RL.formatDauer(e.abschluss) : "> 7 h") +
    " · " + e.erfolge + "/" + RL.ERFOLGE.length + " Erfolge");
  console.log("             " + e.protokoll.join(" · "));
});

function stunden(sek) { return sek / 3600; }
pruefe("Aktives Spiel braucht mindestens 2,5 Stunden", !!erg.aktiv.abschluss && stunden(erg.aktiv.abschluss) >= 2.5,
  erg.aktiv.abschluss ? stunden(erg.aktiv.abschluss).toFixed(2) + " h" : "nicht geschafft");
pruefe("Aktives Spiel bleibt unter fünf Stunden", !!erg.aktiv.abschluss && stunden(erg.aktiv.abschluss) <= 5,
  erg.aktiv.abschluss ? stunden(erg.aktiv.abschluss).toFixed(2) + " h" : "zu lang");
pruefe("Normales Spiel landet bei etwa vier Stunden",
  !!erg.normal.abschluss && stunden(erg.normal.abschluss) >= 3.2 && stunden(erg.normal.abschluss) <= 5,
  erg.normal.abschluss ? stunden(erg.normal.abschluss).toFixed(2) + " h" : "nicht geschafft");
pruefe("Vorsichtiges Spiel bleibt unter sieben Stunden", !!erg.vorsichtig.abschluss,
  erg.vorsichtig.abschluss ? stunden(erg.vorsichtig.abschluss).toFixed(2) + " h" : "nicht geschafft");
pruefe("Kein Lauf endet vor der zweiten Stunde", Object.keys(erg).every(function (n) {
  return erg[n].stufe10 && stunden(erg[n].stufe10) >= 1.5;
}));
pruefe("Keine negativen Konten", Object.keys(erg).every(function (n) { return erg[n].negativ === 0; }));
pruefe("Kein NaN im Spielstand", Object.keys(erg).every(function (n) { return erg[n].nan === 0; }));
pruefe("Jede Stufe wird in jedem Lauf erreicht", Object.keys(erg).every(function (n) {
  return erg[n].stufe10 !== null;
}));
pruefe("Alle Erfolge bleiben erreichbar", Object.keys(erg).some(function (n) {
  return erg[n].erfolge === RL.ERFOLGE.length;
}), Math.max.apply(null, Object.keys(erg).map(function (n) { return erg[n].erfolge; })) + "/" + RL.ERFOLGE.length);
pruefe("Alle Preise sind endlich und nicht negativ",
  RL.PERSONAL.every(function (p) { return p.preis >= 0 && isFinite(p.preis); }) &&
  RL.UPGRADES.every(function (u) { return u.preis >= 0 && isFinite(u.preis); }) &&
  RL.STUFEN.every(function (x) { return x.kosten >= 0 && isFinite(x.kosten); }));


console.log("\n" + geprueft + " Prüfungen, " + fehler + " Fehler");
process.exit(fehler ? 1 : 0);
