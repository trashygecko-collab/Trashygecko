/* Prüfprogramm ROTLICHT – Regelprüfung und Durchspiel-Simulation.
   Aufruf: node pruefung.js
   Braucht: ../../spiele/rotlicht/rotlicht-kern.js */
"use strict";

var pfad = require("path");
var RL = require(pfad.join(__dirname, "..", "..", "spiele", "rotlicht", "rotlicht-kern.js"));

var fehler = 0, geprueft = 0;
function pruefe(name, bedingung, zusatz) {
  geprueft++;
  if (!bedingung) {
    fehler++;
    console.log("FEHLER  " + name + (zusatz !== undefined ? "  → " + zusatz : ""));
  } else {
    console.log("ok      " + name + (zusatz !== undefined ? "  (" + zusatz + ")" : ""));
  }
}
function nah(a, b, tol) { return Math.abs(a - b) <= (tol || 1e-6); }

var T0 = 1700000000000;
function neu() { return RL.neuerStand(T0); }

/* ------------------------------------------------------------ Startzustand */
var s = neu();
pruefe("Start: 0 €", s.geld === 0);
pruefe("Start: Stufe 1", s.stufe === 1);
pruefe("Start: kein Personal", RL.anzahlPersonal(s) === 0);
pruefe("Start: keine Kondome", s.kondome === 0);
pruefe("Start: 0 € je Stunde", RL.werte(s).proStunde === 0);

/* ------------------------------------------------------------- Handarbeit */
var r = RL.klick(s, "hand", T0 + 1000);
pruefe("Handarbeit gibt 20 €", r.ok && r.betrag === 20 && s.geld === 20, RL.formatGeld(s.geld));
r = RL.klick(s, "hand", T0 + 2000);
pruefe("Handarbeit gesperrt bis zur nächsten Spielstunde", !r.ok && r.grund === "warte", r.rest + " s");
r = RL.klick(s, "hand", T0 + 1000 + 60000);
pruefe("Handarbeit nach 60 s wieder frei", r.ok && s.geld === 40, RL.formatGeld(s.geld));

/* --------------------------------------------------------- Zusatzgeschäft */
r = RL.klick(s, "extra", T0 + 61000);
pruefe("Zusatzgeschäft ohne Kondom abgelehnt", !r.ok && r.grund === "keine-kondome");
r = RL.kondomeKaufen(s, 1, T0 + 62000);
pruefe("Viererpack: 4 Kondome für 5 €", r.ok && r.anzahl === 4 && r.preis === 5 && s.kondome === 4 && nah(s.geld, 35), RL.formatGeld(s.geld));
r = RL.klick(s, "extra", T0 + 63000);
pruefe("Zusatzgeschäft gibt 50 € und verbraucht ein Kondom", r.ok && r.betrag === 50 && s.kondome === 3, RL.formatGeld(s.geld));
r = RL.klick(s, "extra", T0 + 64000);
pruefe("Zusatzgeschäft hat eigenen 60-s-Takt", !r.ok && r.grund === "warte");

/* ------------------------------------------------------- Kauf ohne Geld */
var s2 = neu();
r = RL.personalKaufen(s2, "mia", 1, T0);
pruefe("Kauf ohne Geld scheitert", !r.ok && r.grund === "geld");
pruefe("Kontostand bleibt 0 (nicht negativ)", s2.geld === 0);
r = RL.upgradeKaufen(s2, "tafel", 1, T0);
pruefe("Upgrade ohne Geld scheitert", !r.ok);
r = RL.stufeKaufen(s2, T0);
pruefe("Stufe 2 ohne Geld scheitert", !r.ok);

/* ---------------------------------------------------------- Personal */
var s3 = neu();
s3.geld = 10000;
r = RL.personalKaufen(s3, "mia", 1, T0);
pruefe("Erste Mitarbeiterin kostet 200 €", r.ok && r.preis === 200, RL.formatGeld(s3.geld));
pruefe("Erste Mitarbeiterin bringt 20 €/h", RL.werte(s3).personalStunde === 20);
r = RL.personalKaufen(s3, "ruby", 1, T0);
pruefe("Zweite Mitarbeiterin kostet 500 €", r.ok && r.preis === 500);
pruefe("Beide zusammen 120 €/h Grundlohn", RL.werte(s3).personalStunde === 120, RL.formatGeld(RL.werte(s3).proStunde) + "/h mit Erfolgen");
pruefe("Zweite Mia teurer als die erste", RL.personalKosten(s3, RL.personalDef("mia"), 1) > 200,
  RL.formatGeld(RL.personalKosten(s3, RL.personalDef("mia"), 1)));
r = RL.personalKaufen(s3, "bea", 1, T0);
pruefe("Gesperrte Mitarbeiterin (Stufe 2 nötig) abgelehnt", !r.ok && r.grund === "gesperrt");

/* ------------------------------------------------------------ Abrechnen */
var s4 = neu();
s4.geld = 1000;
RL.personalKaufen(s4, "ruby", 1, T0);         /* 100 €/h plus Erfolgsbonus */
s4.geld = 0;
var stunde4 = RL.werte(s4).proSekunde * 60;
var zu = RL.abrechnen(s4, T0 + 60000);         /* eine Spielstunde */
pruefe("Eine Spielstunde zahlt den Stundensatz", nah(zu, stunde4) && nah(s4.geld, stunde4), RL.formatGeld(zu));
zu = RL.abrechnen(s4, T0 + 120000);
pruefe("Zweite Spielstunde nochmal derselbe Satz", nah(zu, stunde4), RL.formatGeld(s4.geld));
pruefe("Ruby allein bringt 100 €/h plus Erfolg", RL.werte(s4).personalStunde === 100, RL.formatGeld(RL.werte(s4).proStunde) + "/h");
var vor = s4.geld;
zu = RL.abrechnen(s4, T0 + 120000 - 5000);
pruefe("Uhr rückwärts bringt kein Geld", zu === 0 && s4.geld === vor);

/* --------------------------------------------------------------- Upgrades */
var s5 = neu();
s5.geld = 20000;
var vorh = RL.werte(s5).proStunde;
r = RL.upgradeKaufen(s5, "tafel", T0);
pruefe("Upgrade Reklametafel kostet 250 €", r.ok && r.preis === 250, RL.formatGeld(s5.geld));
pruefe("Reklametafel erhöht Einnahmen ×1,15", nah(RL.werte(s5).proStunde, vorh * 1.15));
r = RL.upgradeKaufen(s5, "tafel", T0);
pruefe("Upgrade nicht doppelt kaufbar", !r.ok && r.grund === "gekauft");
var cdVor = RL.werte(s5).cooldown;
RL.upgradeKaufen(s5, "kalender", T0);
pruefe("Terminkalender senkt Wartezeit auf 51 s", nah(RL.werte(s5).cooldown, 51) && RL.werte(s5).cooldown < cdVor,
  RL.werte(s5).cooldown + " s");
s5.stufe = 4;
RL.upgradeKaufen(s5, "reinigung", T0);
RL.upgradeKaufen(s5, "sicherheit", T0);
var w5 = RL.werte(s5);
pruefe("Wartezeit bleibt über 15 s", w5.cooldown >= 15, w5.cooldown + " s");

/* ---------------------------------------------------------------- Stufen */
var s6 = neu();
s6.geld = 1e21;
var namen = [];
for (var i = 0; i < 9; i++) {
  var nn = RL.naechsteStufe(s6);
  r = RL.stufeKaufen(s6, T0);
  if (r.ok) namen.push(r.name);
}
pruefe("Alle neun Ausbaustufen kaufbar", s6.stufe === 10 && namen.length === 9, namen.join(" → "));
pruefe("Stufe 10 versechsfacht die Einnahmen", RL.werte(s6).stufenFaktor === 6);
r = RL.stufeKaufen(s6, T0);
pruefe("Über Stufe 10 hinaus kein Kauf", !r.ok && r.grund === "max");
var gesamtStufen = 0;
for (var k = 1; k < RL.STUFEN.length; k++) gesamtStufen += RL.STUFEN[k].kosten;
pruefe("Ausbaukosten steigen streng", (function () {
  for (var j = 1; j < RL.STUFEN.length; j++) if (!(RL.STUFEN[j].kosten > RL.STUFEN[j - 1].kosten * 3)) return false;
  return true;
})(), RL.formatGeld(gesamtStufen) + " gesamt");

/* --------------------------------------------------------------- Standorte */
var s7 = neu();
s7.geld = 1e6;
var st0 = s7.standorte, f0 = RL.werte(s7).standortFaktor;
r = RL.standortKaufen(s7, T0);
pruefe("Standort kostet 5.000 € und zählt hoch", r.ok && s7.standorte === st0 + 1 && r.preis === 5000);
pruefe("Standort erhöht Einnahmen ×1,22", nah(RL.werte(s7).standortFaktor, f0 * 1.22));
pruefe("Zweiter Standort teurer", RL.standortKosten(s7) > 5000, RL.formatGeld(RL.standortKosten(s7)));

/* ----------------------------------------------------------- Offline-Zeit */
var s8 = neu();
s8.geld = 1000;
RL.personalKaufen(s8, "ruby", 1, T0);       /* 100 €/h */
s8.geld = 0;
var satz8 = RL.werte(s8).proSekunde;
zu = RL.abrechnen(s8, T0 + 3 * 3600 * 1000);
pruefe("Drei Stunden Abwesenheit zahlen voll", nah(zu, satz8 * 3 * 3600), RL.formatGeld(zu));
pruefe("Rückkehrbericht gesetzt", s8.bericht && s8.bericht.sek === 10800);
var s8b = neu();
s8b.geld = 1000;
RL.personalKaufen(s8b, "ruby", 1, T0);
s8b.geld = 0;
var satz8b = RL.werte(s8b).proSekunde;
zu = RL.abrechnen(s8b, T0 + 20 * 3600 * 1000);
pruefe("Zwanzig Stunden werden auf acht gekappt", nah(zu, satz8b * 8 * 3600), RL.formatGeld(zu));
var s8c = neu();
s8c.geld = 1e9;
s8c.stufe = 3;
RL.upgradeKaufen(s8c, "buchhaltung", T0);
RL.personalKaufen(s8c, "ruby", 1, T0);
s8c.geld = 0;
var satz8c = RL.werte(s8c).proSekunde;
zu = RL.abrechnen(s8c, T0 + 20 * 3600 * 1000);
pruefe("Buchhaltung: zehn Stunden und +25 %", nah(zu, satz8c * 10 * 3600 * 1.25), RL.formatGeld(zu));

/* ------------------------------------------------------------ Speichern */
var s9 = neu();
s9.geld = 12345.67;
RL.personalKaufen(s9, "mia", 2, T0);
RL.upgradeKaufen(s9, "tafel", T0);
s9.kondome = 7;
s9.stufe = 2;
s9.standorte = 3;
var text = RL.speicherText(s9);
var s9b = RL.laden(text);
pruefe("Speichern und Laden erhält das Geld", nah(s9b.geld, s9.geld, 0.01), RL.formatGeld(s9b.geld));
pruefe("… das Personal", s9b.personal.mia === 2);
pruefe("… die Upgrades", s9b.upgrades.tafel === true);
pruefe("… Kondome, Stufe, Standorte", s9b.kondome === 7 && s9b.stufe === 2 && s9b.standorte === 3);
pruefe("… die Einnahmen je Stunde", nah(RL.werte(s9b).proStunde, RL.werte(s9).proStunde));
var kaputt = RL.laden("{das ist kein json");
pruefe("Kaputter Spielstand → frischer Start", kaputt.geld === 0 && kaputt.stufe === 1);
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
pruefe("Erfolg erhöht den Gesamtfaktor", RL.werte(s10).gesamtFaktor > e0, RL.werte(s10).gesamtFaktor.toFixed(3));

/* ------------------------------------------------------------ Ereignisse */
var s11 = neu();
s11.geld = 1000;
RL.personalKaufen(s11, "ruby", 1, T0);
s11.geld = 0;
var ereignis = null, tE = T0;
for (var sek = 1; sek < 400 && !ereignis; sek++) {
  tE = T0 + sek * 1000;
  ereignis = RL.ereignis(s11, tE);
}
pruefe("Ereignis erscheint nach kurzer Zeit", !!ereignis);
pruefe("Nächstes Ereignis liegt in der Zukunft", s11.naechstesEreignis > tE);
pruefe("Chronik füllt sich", s11.chronik.length > 0, s11.chronik.length + " Zeilen");

/* ------------------------------------------------------------ Zahlenformat */
pruefe("0 € wird „0,00 €“", RL.formatGeld(0) === "0,00 €", RL.formatGeld(0));
pruefe("1.234,50 € ohne Nachkommastellen", RL.formatGeld(1234.5) === "1.235 €", RL.formatGeld(1234.5));
pruefe("Millionen mit Einheit", RL.formatGeld(1.234e6).indexOf("Mio.") > 0, RL.formatGeld(1.234e6));
pruefe("Milliarden mit Einheit", RL.formatGeld(5.5e9).indexOf("Mrd.") > 0, RL.formatGeld(5.5e9));
pruefe("Billionen mit Einheit", RL.formatGeld(2.5e12).indexOf("Bio.") > 0, RL.formatGeld(2.5e12));
pruefe("Billiarden mit Einheit", RL.formatGeld(7e15).indexOf("Brd.") > 0, RL.formatGeld(7e15));
pruefe("kein NaN in der Ausgabe", RL.formatGeld(NaN).indexOf("NaN") < 0, RL.formatGeld(NaN));

/* ------------------------------------------------------- Durchspiel-Simulation */
function kaufBester(s, t) {
  var w = RL.werte(s);
  var bester = null;
  function biete(art, id, kosten, zuwachs) {
    if (kosten <= 0 || s.geld + 1e-9 < kosten) return;
    var nutzen = zuwachs / kosten;
    if (!bester || nutzen > bester.nutzen) bester = { art: art, id: id, nutzen: nutzen, kosten: kosten };
  }
  var ns = RL.naechsteStufe(s);
  if (ns) biete("stufe", null, ns.kosten, (ns.faktor / w.stufenFaktor - 1) * w.proSekunde);
  for (var i = 0; i < RL.PERSONAL.length; i++) {
    var p = RL.PERSONAL[i];
    if (s.stufe < p.ab) continue;
    biete("personal", p.id, RL.personalKosten(s, p, 1), p.stunde * w.gesamtFaktor / RL.STUNDE_SEK);
  }
  for (i = 0; i < RL.UPGRADES.length; i++) {
    var u = RL.UPGRADES[i];
    if (s.upgrades[u.id] || s.stufe < u.ab) continue;
    if (u.art === "einnahme") biete("upgrade", u.id, u.preis, (u.faktor - 1) * w.proSekunde);
    else if (u.art === "tempo") biete("upgrade", u.id, u.preis, (1 - u.faktor) * w.proSekunde * 0.25);
  }
  biete("standort", null, RL.standortKosten(s), 0.22 * w.proSekunde);
  if (!bester) return false;
  var r2;
  if (bester.art === "stufe") r2 = RL.stufeKaufen(s, t);
  else if (bester.art === "personal") r2 = RL.personalKaufen(s, bester.id, 1, t);
  else if (bester.art === "upgrade") r2 = RL.upgradeKaufen(s, bester.id, t);
  else r2 = RL.standortKaufen(s, t);
  return !!(r2 && r2.ok);
}

function simulation(maxStunden, mitKlicks) {
  var st = RL.neuerStand(T0);
  var t = T0;
  var stuferZeit = {}, letzteStufe = 1;
  var maxGeld = 0, kaeufe = 0, negatives = 0, nan = 0;
  var protokoll = [];
  var sekunden = Math.round(maxStunden * 3600);
  for (var sek = 1; sek <= sekunden; sek++) {
    t = T0 + sek * 1000;
    RL.abrechnen(st, t);
    if (mitKlicks) {
      if (st.kondome < 1 && st.geld >= 5) RL.kondomeKaufen(st, 1, t);
      RL.klick(st, "hand", t);
      RL.klick(st, "extra", t);
    }
    for (var n = 0; n < 400 && kaufBester(st, t); n++) kaeufe++;
    if (!isFinite(st.geld)) nan++;
    if (st.geld < -1e-6) negatives++;
    if (st.geld > maxGeld) maxGeld = st.geld;
    if (st.stufe !== letzteStufe) {
      letzteStufe = st.stufe;
      stuferZeit[st.stufe] = sek / 3600;
      protokoll.push("Stufe " + st.stufe + " (" + RL.STUFEN[st.stufe - 1].name + ") nach " +
        RL.formatDauer(sek) + ", " + RL.werte(st).proStunde.toFixed(0) + " €/h");
    }
  }
  return {
    stand: st, stuferZeit: stuferZeit, protokoll: protokoll, kaeufe: kaeufe,
    negatives: negatives, nan: nan, maxGeld: maxGeld
  };
}

console.log("\n--- Simulation mit Klicks (aktiv) ---");
var a = simulation(6, true);
console.log(a.protokoll.join("\n"));
console.log("Ende nach 6 h: " + RL.formatGeld(a.stand.geld) + ", " +
  RL.formatGeld(RL.werte(a.stand).proStunde) + "/h, " + RL.anzahlPersonal(a.stand) + " Mitarbeiter, " +
  a.stand.standorte + " Standorte");
pruefe("Simulation ohne negative Konten", a.negatives === 0);
pruefe("Simulation ohne NaN", a.nan === 0);
pruefe("Stufe 5 in unter 2 Stunden", (a.stuferZeit[5] || 99) < 2, (a.stuferZeit[5] || 99).toFixed(2) + " h");
pruefe("Stufe 10 wird erreicht", !!a.stuferZeit[10], a.stuferZeit[10] ? a.stuferZeit[10].toFixed(2) + " h" : "nicht in 6 h");
pruefe("Jede Stufe dauert länger als die vorige", (function(){var v=[];for(var q=2;q<=10;q++){if(!a.stuferZeit[q])return false;v.push(a.stuferZeit[q]);}for(q=1;q<v.length;q++)if(!(v[q]>v[q-1]))return false;return true;})(), JSON.stringify(a.stuferZeit));

console.log("\n--- Simulation ohne Klicks (nur zusehen) ---");
var b = simulation(8, false);
console.log(b.protokoll.join("\n") || "(keine Stufe erreicht)");
console.log("Ende nach 8 h: " + RL.formatGeld(b.stand.geld) + ", " + RL.formatGeld(RL.werte(b.stand).proStunde) + "/h");
pruefe("Ohne Klicks kein Fortschritt am Anfang (erwartet)", b.stand.stufe === 1);
pruefe("Ohne Klicks kein negatives Konto", b.negatives === 0);

console.log("\n--- Kaufmengen ---");
var s12 = neu();
s12.geld = 100000;
var p = RL.personalDef("mia");
var max = RL.maxPersonal(s12, p);
pruefe("maxPersonal passt zum Geld", RL.personalKosten(s12, p, max) <= s12.geld &&
  RL.personalKosten(s12, p, max + 1) > s12.geld, max + " Stück für " + RL.formatGeld(RL.personalKosten(s12, p, max)));
RL.personalKaufen(s12, "mia", max, T0);
pruefe("Mengenkauf bucht genau einmal ab", s12.geld >= 0 && s12.personal.mia === max, s12.personal.mia + " Mia");

console.log("\n" + geprueft + " Prüfungen, " + fehler + " Fehler");
process.exit(fehler ? 1 : 0);
