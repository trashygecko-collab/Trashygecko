/* ===========================================================================
   ROTLICHT – Spielregeln und Wirtschaft (ohne DOM)

   Diese Datei kennt kein Fenster, keinen Speicher und keine Zeitquelle: alles
   wird hereingereicht. Dadurch kann dasselbe Rechenwerk im Browser laufen und
   in Node Bild für Bild durchgespielt werden (sicherung/…/pruefung.js).

   Grundeinheit: eine Spielstunde = 60 Sekunden Echtzeit.
   =========================================================================== */
(function (root, fabrik) {
  "use strict";
  var RL = fabrik();
  if (typeof module === "object" && module.exports) module.exports = RL;
  if (root) root.RLKern = RL;
})(typeof window !== "undefined" ? window : (typeof globalThis !== "undefined" ? globalThis : this), function () {
  "use strict";

  var VERSION = 1;
  var STUNDE_SEK = 60;              /* eine Spielstunde in Echtzeitsekunden */
  var KONDOM = { anzahl: 4, preis: 5 };
  var OFFLINE_GRUND_STUNDEN = 8;

  /* ---------------------------------------------------------------- Stufen */
  var STUFEN = [
    { nr: 1, name: "Straßenarbeit", kosten: 0, faktor: 1,
      text: "Ein Laternenpfahl, eine Decke, große Träume. Mehr ist gerade nicht drin." },
    { nr: 2, name: "Kleine private Unterkunft", kosten: 500, faktor: 1.5,
      text: "Zum ersten Mal ein Dach über dem Kopf – und ein Türschild, das jeder sieht." },
    { nr: 3, name: "Erste eigene Geschäftsräume", kosten: 2000, faktor: 1.9,
      text: "Schaufenster, Neonherz, Kopierer. Die Nachbarschaft nennt es „Studio“." },
    { nr: 4, name: "Kleines Freudenhaus", kosten: 8647.26, faktor: 2.2,
      text: "Drei Zimmer, ein Wartebereich und ein Staubsauger mit sehr viel Leistung." },
    { nr: 5, name: "Größeres Etablissement", kosten: 6.41302e+06, faktor: 2.6,
      text: "Bar, Musik, Kaffeemaschine. Die Kaffeemaschine ist der wahre Star." },
    { nr: 6, name: "Mehrstöckiges Rotlichtgebäude", kosten: 2.72076e+11, faktor: 3,
      text: "Aufzug, Etagenklingel, Sicherheitsdienst. Der Aufzug fährt jetzt rund um die Uhr." },
    { nr: 7, name: "Luxus-Freudenhaus", kosten: 2.19104e+13, faktor: 3.5,
      text: "Marmor, Samt und ein Pissoir mit Beleuchtung. Der Hausmeister weint vor Glück." },
    { nr: 8, name: "Großes Stundenhotel", kosten: 4.61924e+15, faktor: 4,
      text: "Zimmerkarten, Zimmerservice, Zimmer mit Aussicht auf den Parkplatz." },
    { nr: 9, name: "Luxushotel in der Großstadt", kosten: 1.14919e+18, faktor: 4.5,
      text: "Die Skyline kennt euren Namen. Der Concierge kennt ihn schon lange." },
    { nr: 10, name: "Internationales Luxusimperium", kosten: 2.22149e+20, faktor: 6,
      text: "Von der Straße an die Börse. Die erste Laterne steht jetzt im Foyer." }
  ];

  /* -------------------------------------------------------------- Personal */
  /* preis = erster Mitarbeiter, stunde = Euro je Spielstunde und Kopf,
     wuchs = Preissteigerung je weiterem Kopf, ab = nötige Stufe. */
  var PERSONAL = [
    { id: "mia", name: "Mia", rolle: "Praktikantin", preis: 200, stunde: 20, wuchs: 1.23, ab: 1,
      text: "Bringt Kaffee, gute Laune und den ersten zahlenden Gast." },
    { id: "ruby", name: "Ruby", rolle: "Hostess", preis: 500, stunde: 100, wuchs: 1.23, ab: 1,
      text: "Hat ein Händchen für Stammkunden und ein zweites für Trinkgeld." },
    { id: "bea", name: "Bea", rolle: "Studioleitung", preis: 3000, stunde: 550, wuchs: 1.24, ab: 2,
      text: "Führt Dienstplan, Preistafel und die Nerven aller anderen." },
    { id: "kevin", name: "Kevin", rolle: "Türsteher", preis: 12000, stunde: 2200, wuchs: 1.24, ab: 2,
      text: "Sagt „kein Zutritt“ so freundlich, dass es wie ein Kompliment klingt." },
    { id: "krueger", name: "Frau Krüger", rolle: "Buchhaltung", preis: 60000, stunde: 12000, wuchs: 1.25, ab: 3,
      text: "Findet jede Quittung. Auch die, die nie geschrieben wurde." },
    { id: "ali", name: "Ali", rolle: "Chauffeur", preis: 350000, stunde: 80000, wuchs: 1.25, ab: 3,
      text: "Fährt diskret im Kreis, bis der Gast fertig ist. Tanken zahlt die Firma." },
    { id: "bohne", name: "Dr. Bohne", rolle: "Anwalt", preis: 2.5e6, stunde: 700000, wuchs: 1.26, ab: 4,
      text: "Liest das Kleingedruckte und ergänzt es um ein zweites Kleingedrucktes." },
    { id: "steffi", name: "Steffi", rolle: "Marketing", preis: 2e7, stunde: 6e6, wuchs: 1.26, ab: 5,
      text: "Macht aus einem roten Fenster eine Marke mit eigenem Hashtag." },
    { id: "conny", name: "Conny", rolle: "Hotelmanagerin", preis: 3e8, stunde: 9e7, wuchs: 1.27, ab: 6,
      text: "Checkt Gäste ein, Beschwerden aus und Personal nach." },
    { id: "glanz", name: "Agentur Glanz", rolle: "PR", preis: 6e9, stunde: 2e9, wuchs: 1.27, ab: 7,
      text: "Presse, Podcasts, Preisverleihungen – alles mit demselben roten Teppich." },
    { id: "guenther", name: "Günther", rolle: "Immobilienhai", preis: 1.5e11, stunde: 5e10, wuchs: 1.28, ab: 8,
      text: "Kauft Häuser, bevor sie gebaut sind. Manchmal auch die Straße davor." },
    { id: "family", name: "Family Office", rolle: "Vermögensverwaltung", preis: 4e12, stunde: 1.4e12, wuchs: 1.28, ab: 9,
      text: "Verwaltet das Geld so diskret, dass selbst ihr Chef es nicht mehr findet." },
    { id: "concierge", name: "Jet-Set-Concierge", rolle: "Concierge", preis: 2e14, stunde: 8e13, wuchs: 1.29, ab: 10,
      text: "Bucht Suite, Jet und Alibi – in dieser Reihenfolge." },
    { id: "filz", name: "Dr. Filz", rolle: "Lobbyist", preis: 1e16, stunde: 4e15, wuchs: 1.29, ab: 10,
      text: "Sorgt dafür, dass die Steuererklärung so aussieht wie ein Gartenzaun." }
  ];

  /* -------------------------------------------------------------- Upgrades */
  var UPGRADES = [
    { id: "tafel", name: "Reklametafel", art: "einnahme", faktor: 1.15, preis: 250, ab: 1,
      text: "Ein Schild an der Laterne: diskret, aber mit Pfeil." },
    { id: "ausstattung", name: "Bessere Ausstattung", art: "einnahme", faktor: 1.2, preis: 900, ab: 1,
      text: "Frische Bezüge, weiche Decke, ein Spiegel ohne Sprung." },
    { id: "kalender", name: "Terminkalender", art: "tempo", faktor: 0.85, preis: 1500, ab: 1,
      text: "Doppelbuchungen werden weniger. Die Wartezeit auch." },
    { id: "eingang", name: "Zweiter Eingang", art: "einnahme", faktor: 1.25, preis: 24000, ab: 2,
      text: "Zwei Türen, zwei Wege, halb so viel Gedränge im Flur." },
    { id: "blatt", name: "Werbung im Lokalblatt", art: "einnahme", faktor: 1.3, preis: 80000, ab: 2,
      text: "Vier Zeilen, ein Foto vom Hinterkopf, erstaunlich viele Anrufe." },
    { id: "reinigung", name: "Reinigungspersonal", art: "einnahme", faktor: 1.2, preis: 180000, ab: 3,
      text: "Riecht nach Zitrone statt nach Entscheidungen." },
    { id: "sicherheit", name: "Sicherheitsdienst", art: "einnahme", faktor: 1.25, preis: 360000, ab: 3,
      text: "Zwei kräftige Kollegen und ein Funkgerät mit Taschenlampe." },
    { id: "buchhaltung", name: "Automatische Buchhaltung", art: "offline", faktor: 1.25, stunden: 2,
      preis: 600000, ab: 3,
      text: "Zählt weiter, während niemand hinsieht – und findet das völlig normal." },
    { id: "raeume", name: "Größere Räumlichkeiten", art: "einnahme", faktor: 1.35, preis: 1.6e+06, ab: 4,
      text: "Mehr Zimmer heißt mehr Umsatz. Und mehr Flur, der gewischt werden will." },
    { id: "marketing", name: "Professionelles Marketing", art: "einnahme", faktor: 1.4, preis: 4.8e+06, ab: 4,
      text: "Zielgruppe, Reichweite, rote Zahlen nur in der Bilanz." },
    { id: "kapazitaet", name: "Kundenkapazität verdoppelt", art: "einnahme", faktor: 1.5, preis: 1.6e+07, ab: 5,
      text: "Statt Wartezimmer jetzt Wartebereich mit Sitzsack und Bildschirm." },
    { id: "etage", name: "Zusätzliche Etage", art: "einnahme", faktor: 1.3, preis: 6e+07, ab: 5,
      text: "Eine Treppe mehr, ein Aufzug weniger. Beschwerden: zwei." },
    { id: "organisation", name: "Bessere Organisation", art: "tempo", faktor: 0.85, preis: 2e+08, ab: 6,
      text: "Schichtplan aus der Cloud. Der Kalender hat jetzt einen Kalender." },
    { id: "neon", name: "Neonfassade", art: "einnahme", faktor: 1.35, preis: 8e+08, ab: 6,
      text: "Von drei Autobahnen aus sichtbar. Die Nachbarn haben aufgehört zu zählen." },
    { id: "vip", name: "VIP-Bereich", art: "einnahme", faktor: 1.5, preis: 4e+09, ab: 7,
      text: "Samt, Sekt und eine Tür, die nur nach oben aufgeht." },
    { id: "waescherei", name: "Eigene Wäscherei", art: "einnahme", faktor: 1.2, preis: 2.4e+10, ab: 7,
      text: "Drei Maschinen, ein Trockner und ein Mitarbeiter mit starkem Rücken." },
    { id: "chauffeur", name: "Chauffeur-Service", art: "einnahme", faktor: 1.4, preis: 1.6e+11, ab: 8,
      text: "Abholung an der Haustür, Rückweg ohne Umwege, Scheiben undurchsichtig." },
    { id: "interieur", name: "Designer-Interieur", art: "einnahme", faktor: 1.5, preis: 1.2e+12, ab: 8,
      text: "Alles in Roségold. Sogar der Aschenbecher hat einen Namen." },
    { id: "pr", name: "Eigene PR-Abteilung", art: "einnahme", faktor: 1.6, preis: 8e+12, ab: 9,
      text: "Schreibt die Schlagzeilen selbst, bevor es jemand anders tut." },
    { id: "lounge", name: "Luxus-Lounge", art: "tempo", faktor: 0.8, preis: 8e+13, ab: 9,
      text: "Wer wartet, wartet jetzt auf einer Ledercouch mit Blick auf die Stadt." },
    { id: "buchung", name: "Globales Buchungssystem", art: "offline", faktor: 1.3, stunden: 2,
      preis: 1.2e+15, ab: 10,
      text: "Bucht rund um die Uhr, in jeder Zeitzone, ohne ein einziges Telefonat." },
    { id: "jet", name: "Privatjet-Flotte", art: "einnahme", faktor: 1.8, preis: 2e+16, ab: 10,
      text: "Zwei Jets, eine Crew, sehr viele rote Koffer." },
    { id: "steuern", name: "Steueroptimierung Deluxe", art: "einnahme", faktor: 1.7, preis: 3.2e+17, ab: 10,
      text: "Der Firmensitz ist jetzt ein Briefkasten mit Meerblick." }
  ];

  /* ------------------------------------------------------------- Erfolge */
  var ERFOLGE = [
    { id: "erstes-geld", name: "Erste Einnahme", bonus: 0.02,
      text: "Die ersten 20 € sind verdient – ehrlicher als sie aussehen.",
      pruef: function (s) { return s.gesamtEinnahmen >= 20; } },
    { id: "hundert", name: "Der erste Hunni", bonus: 0.02,
      text: "100 € auf dem Konto. Die Laterne ist jetzt Stammgast.",
      pruef: function (s) { return s.gesamtEinnahmen >= 100; } },
    { id: "erste-kraft", name: "Erste Mitarbeiterin", bonus: 0.03,
      text: "Zum ersten Mal verdient jemand anderes das Geld mit.",
      pruef: function (s) { return anzahlPersonal(s) >= 1; } },
    { id: "team-10", name: "Zehn Leute", bonus: 0.03,
      text: "Ein Team, ein Dienstplan, ein Kaffeeautomat.",
      pruef: function (s) { return anzahlPersonal(s) >= 10; } },
    { id: "team-50", name: "Fünfzig Leute", bonus: 0.05,
      text: "Die Personalabteilung ist jetzt größer als das erste Zimmer.",
      pruef: function (s) { return anzahlPersonal(s) >= 50; } },
    { id: "stufe-3", name: "Eigene Räume", bonus: 0.03,
      text: "Ein Schaufenster mit Neonherz. Kein Umweg mehr über die Straße.",
      pruef: function (s) { return s.stufe >= 3; } },
    { id: "stufe-5", name: "Etablissement", bonus: 0.04,
      text: "Mit Bar. Die Bar war teurer als das Gebäude.",
      pruef: function (s) { return s.stufe >= 5; } },
    { id: "stufe-7", name: "Luxus", bonus: 0.05,
      text: "Marmor im Flur. Der Hausmeister trägt jetzt Handschuhe.",
      pruef: function (s) { return s.stufe >= 7; } },
    { id: "stufe-10", name: "Weltimperium", bonus: 0.08,
      text: "Die Laterne aus Stufe 1 steht im Foyer unter Denkmalschutz.",
      pruef: function (s) { return s.stufe >= 10; } },
    { id: "million", name: "Millionär", bonus: 0.04,
      text: "Eine Million Einnahmen. Die Bank fragt nicht mehr, woher.",
      pruef: function (s) { return s.gesamtEinnahmen >= 1e6; } },
    { id: "milliarde", name: "Milliardär", bonus: 0.05,
      text: "Eine Milliarde. Der Steuerberater hat jetzt einen Steuerberater.",
      pruef: function (s) { return s.gesamtEinnahmen >= 1e9; } },
    { id: "billion", name: "Billionär", bonus: 0.06,
      text: "Eine Billion. Das Finanzamt schickt Grußkarten.",
      pruef: function (s) { return s.gesamtEinnahmen >= 1e12; } },
    { id: "billiarde", name: "Billiardär", bonus: 0.07,
      text: "Zahlen mit 16 Stellen. Die Buchhaltung nennt es „Montag“.",
      pruef: function (s) { return s.gesamtEinnahmen >= 1e15; } },
    { id: "klicks-100", name: "100 Handgriffe", bonus: 0.02,
      text: "Einhundertmal selbst angepackt. Die Hände erinnern sich.",
      pruef: function (s) { return s.klicks >= 100; } },
    { id: "klicks-1000", name: "1000 Handgriffe", bonus: 0.03,
      text: "Tausendmal. Langsam übernimmt das Personal – wie es soll.",
      pruef: function (s) { return s.klicks >= 1000; } },
    { id: "standorte-5", name: "Fünf Standorte", bonus: 0.05,
      text: "Fünf Städte, fünf Neonherzen, eine Rechnung.",
      pruef: function (s) { return s.standorte >= 5; } },
    { id: "standorte-25", name: "Fünfundzwanzig Standorte", bonus: 0.06,
      text: "Die Karte im Büro braucht eine zweite Karte.",
      pruef: function (s) { return s.standorte >= 25; } },
    { id: "upgrades-10", name: "Zehn Verbesserungen", bonus: 0.04,
      text: "Zehnmal nachgerüstet. Der Handwerker kennt den Weg blind.",
      pruef: function (s) { return anzahlUpgrades(s) >= 10; } },
    { id: "upgrades-alle", name: "Alles ausgebaut", bonus: 0.05,
      text: "Keine einzige Baustelle mehr. Unheimlich.",
      pruef: function (s) { return anzahlUpgrades(s) >= UPGRADES.length; } },
    { id: "kondome-100", name: "Großeinkauf", bonus: 0.02,
      text: "Hundert Kondome gekauft. Der Großhandel grüßt freundlich.",
      pruef: function (s) { return s.kondomeGekauft >= 100; } }
  ];
  var ERFOLG_BONUS = {};
  for (var e = 0; e < ERFOLGE.length; e++) ERFOLG_BONUS[ERFOLGE[e].id] = ERFOLGE[e].bonus;

  /* ------------------------------------------------------------- Ereignisse */
  var EREIGNISSE = [
    "Ein Stammkunde zahlt freiwillig das Doppelte und will keine Quittung.",
    "Die Nachbarin beschwert sich über das rote Licht. Sie bleibt als Kundin.",
    "Ein Influencer dreht ein Video vor der Tür – die Schlange wird länger.",
    "Der Lieferant bringt versehentlich zwei Paletten Handtücher mehr.",
    "Ein Gast verliert beim Gehen einen Umschlag. Der Umschlag bleibt da.",
    "Die Lokalzeitung bringt eine Reportage mit dem Titel „Rot, aber rentabel“.",
    "Ein Betriebsausflug der Bank landet versehentlich bei euch.",
    "Jemand hat die Preistafel heimlich um 10 % erhöht. Niemand beschwert sich.",
    "Ein Hochzeitspaar bucht kurzfristig alle Zimmer für die Nacht.",
    "Der Kaffeautomat hat einen guten Tag. Das ganze Team auch.",
    "Ein Prominenter wird gesichtet. Es ist der zweite Prominente, der so aussieht.",
    "Die Feuerwehr prüft die Fluchtwege und lässt ein Trinkgeld da.",
    "Ein Gast bucht zwei Stunden extra, weil der Regen nicht aufhört.",
    "Der Nachbar verlangt Schmerzensgeld für das Neonlicht. Er wird Teilhaber.",
    "Ein Bus mit Junggesellinnen bleibt vor der Tür stehen. Der Motor läuft weiter.",
    "Die Konkurrenz schließt. Ihre Kunden stehen jetzt bei euch in der Reihe."
  ];

  /* --------------------------------------------------------------- Helfer */
  function zahl(x, ersatz) {
    return (typeof x === "number" && isFinite(x)) ? x : (ersatz || 0);
  }
  function ganz(x) {
    var n = Math.floor(zahl(x, 0));
    return n < 0 ? 0 : n;
  }
  function runde(x) { return Math.round(x * 100) / 100; }

  function neuerStand(jetzt) {
    var t = zahl(jetzt, Date.now());
    var s = {
      v: VERSION,
      geld: 0,
      gesamtEinnahmen: 0,
      gesamtAusgaben: 0,
      personal: {},
      upgrades: {},
      erfolge: {},
      kondome: 0,
      kondomeGekauft: 0,
      stufe: 1,
      standorte: 0,
      klicks: 0,
      gespielt: 0,
      bereit: { hand: 0, extra: 0 },
      letzte: t,
      start: t,
      naechstesEreignis: t + 120000,
      chronik: [],
      bericht: null
    };
    for (var i = 0; i < PERSONAL.length; i++) s.personal[PERSONAL[i].id] = 0;
    return s;
  }

  function personalDef(id) {
    for (var i = 0; i < PERSONAL.length; i++) if (PERSONAL[i].id === id) return PERSONAL[i];
    return null;
  }
  function upgradeDef(id) {
    for (var i = 0; i < UPGRADES.length; i++) if (UPGRADES[i].id === id) return UPGRADES[i];
    return null;
  }
  function anzahlPersonal(s) {
    var n = 0;
    for (var i = 0; i < PERSONAL.length; i++) n += ganz(s.personal[PERSONAL[i].id]);
    return n;
  }
  function anzahlUpgrades(s) {
    var n = 0;
    for (var i = 0; i < UPGRADES.length; i++) if (s.upgrades[UPGRADES[i].id]) n++;
    return n;
  }

  /* --------------------------------------------------------------- Werte */
  function werte(s) {
    var lohn = 0, anzahl = 0, i;
    for (i = 0; i < PERSONAL.length; i++) {
      var n = ganz(s.personal[PERSONAL[i].id]);
      lohn += n * PERSONAL[i].stunde;
      anzahl += n;
    }
    var ein = 1, tempo = 1, offlineBonus = 1, offlineStunden = OFFLINE_GRUND_STUNDEN;
    for (i = 0; i < UPGRADES.length; i++) {
      var u = UPGRADES[i];
      if (!s.upgrades[u.id]) continue;
      if (u.art === "einnahme") ein *= u.faktor;
      else if (u.art === "tempo") tempo *= u.faktor;
      else if (u.art === "offline") {
        offlineBonus *= u.faktor;
        offlineStunden += zahl(u.stunden, 0);
      }
    }
    var stufe = STUFEN[Math.min(STUFEN.length, Math.max(1, ganz(s.stufe))) - 1];
    var standort = Math.pow(1.22, ganz(s.standorte));
    var erfolg = 1;
    for (var id in s.erfolge) if (s.erfolge[id] && ERFOLG_BONUS[id]) erfolg += ERFOLG_BONUS[id];
    var gesamt = stufe.faktor * ein * standort * erfolg;
    var proStunde = lohn * gesamt;
    /* Handarbeit und Zusatzgeschäft bleiben bei 20/50 €: nur echte
       Geschäftserweiterungen (Stufe, Upgrades, Standorte) heben sie an,
       keine versteckten Erfolgsboni. */
    return {
      personalStunde: lohn,
      mitarbeiter: anzahl,
      einnahmeFaktor: ein,
      stufenFaktor: stufe.faktor,
      standortFaktor: standort,
      erfolgFaktor: erfolg,
      aktionsFaktor: stufe.faktor * ein * standort,
      gesamtFaktor: gesamt,
      proStunde: proStunde,
      proSekunde: proStunde / STUNDE_SEK,
      tempoFaktor: tempo,
      cooldown: Math.max(15, STUNDE_SEK * tempo),
      offlineBonus: offlineBonus,
      offlineStunden: offlineStunden,
      stufe: stufe
    };
  }

  /* ------------------------------------------------------------ Fortschritt */
  function abrechnen(s, jetzt) {
    var t = zahl(jetzt, Date.now());
    var dt = (t - zahl(s.letzte, t)) / 1000;
    if (!(dt > 0)) { s.letzte = t; return 0; }
    var w = werte(s);
    var maxSek = w.offlineStunden * 3600;
    var fern = dt > 5;
    var sek = Math.min(dt, maxSek);
    s.letzte = t;
    if (w.proSekunde <= 0) { s.gespielt += Math.min(dt, 3600); return 0; }
    var betrag = w.proSekunde * sek * (fern ? w.offlineBonus : 1);
    if (!(betrag > 0)) return 0;
    s.geld += betrag;
    s.gesamtEinnahmen += betrag;
    s.gespielt += Math.min(dt, maxSek);
    if (fern && dt > 120) {
      s.bericht = { sek: dt, geld: betrag, gekappt: dt > maxSek, bonus: w.offlineBonus };
    }
    return betrag;
  }

  function aktionsBetrag(s, art) {
    var w = werte(s);
    var basis = art === "extra" ? 50 : 20;
    return runde(basis * w.aktionsFaktor);
  }

  function klick(s, art, jetzt) {
    var t = zahl(jetzt, Date.now());
    abrechnen(s, t);
    if (art !== "extra") art = "hand";
    var w = werte(s);
    if (!(s.bereit && typeof s.bereit[art] === "number")) s.bereit = { hand: 0, extra: 0 };
    if (t < s.bereit[art]) {
      return { ok: false, grund: "warte", rest: Math.ceil((s.bereit[art] - t) / 1000) };
    }
    if (art === "extra" && ganz(s.kondome) < 1) {
      return { ok: false, grund: "keine-kondome", betrag: aktionsBetrag(s, art) };
    }
    var betrag = aktionsBetrag(s, art);
    if (art === "extra") s.kondome = ganz(s.kondome) - 1;
    s.geld += betrag;
    s.gesamtEinnahmen += betrag;
    s.klicks = ganz(s.klicks) + 1;
    s.bereit[art] = t + w.cooldown * 1000;
    erfolgePruefen(s, t);
    return { ok: true, betrag: betrag, art: art, cooldown: w.cooldown };
  }

  function kondomeKaufen(s, pakete, jetzt) {
    abrechnen(s, zahl(jetzt, Date.now()));
    var anzahl = Math.max(1, Math.floor(zahl(pakete, 1)));
    var preis = runde(KONDOM.preis * anzahl);
    if (s.geld + 1e-9 < preis) return { ok: false, grund: "geld", preis: preis };
    s.geld -= preis;
    s.gesamtAusgaben += preis;
    s.kondome = ganz(s.kondome) + KONDOM.anzahl * anzahl;
    s.kondomeGekauft = ganz(s.kondomeGekauft) + KONDOM.anzahl * anzahl;
    erfolgePruefen(s, jetzt);
    return { ok: true, anzahl: KONDOM.anzahl * anzahl, preis: preis };
  }

  function personalKosten(s, p, anzahl) {
    var n = ganz(s.personal[p.id]);
    var preis = p.preis * Math.pow(p.wuchs, n);
    var summe = 0;
    for (var i = 0; i < anzahl; i++) { summe += preis; preis *= p.wuchs; }
    return runde(summe);
  }

  function maxPersonal(s, p) {
    var n = ganz(s.personal[p.id]);
    var preis = p.preis * Math.pow(p.wuchs, n);
    var rest = s.geld, k = 0;
    while (k < 100000 && rest + 1e-9 >= preis) { rest -= preis; preis *= p.wuchs; k++; }
    return k;
  }

  function personalKaufen(s, id, anzahl, jetzt) {
    abrechnen(s, zahl(jetzt, Date.now()));
    var p = personalDef(id);
    if (!p) return { ok: false, grund: "unbekannt" };
    if (s.stufe < p.ab) return { ok: false, grund: "gesperrt", ab: p.ab };
    anzahl = Math.max(1, Math.floor(zahl(anzahl, 1)));
    var preis = personalKosten(s, p, anzahl);
    if (s.geld + 1e-9 < preis) return { ok: false, grund: "geld", preis: preis };
    s.geld -= preis;
    s.gesamtAusgaben += preis;
    s.personal[id] = ganz(s.personal[id]) + anzahl;
    erfolgePruefen(s, jetzt);
    return { ok: true, anzahl: anzahl, preis: preis };
  }

  function upgradeKaufen(s, id, jetzt) {
    abrechnen(s, zahl(jetzt, Date.now()));
    var u = upgradeDef(id);
    if (!u) return { ok: false, grund: "unbekannt" };
    if (s.upgrades[id]) return { ok: false, grund: "gekauft" };
    if (s.stufe < u.ab) return { ok: false, grund: "gesperrt", ab: u.ab };
    if (s.geld + 1e-9 < u.preis) return { ok: false, grund: "geld", preis: u.preis };
    s.geld -= u.preis;
    s.gesamtAusgaben += u.preis;
    s.upgrades[id] = true;
    erfolgePruefen(s, jetzt);
    return { ok: true, preis: u.preis };
  }

  function standortKosten(s) {
    return runde(5000 * Math.pow(2.4, ganz(s.standorte)));
  }

  function standortKaufen(s, jetzt) {
    abrechnen(s, zahl(jetzt, Date.now()));
    var preis = standortKosten(s);
    if (s.geld + 1e-9 < preis) return { ok: false, grund: "geld", preis: preis };
    s.geld -= preis;
    s.gesamtAusgaben += preis;
    s.standorte = ganz(s.standorte) + 1;
    erfolgePruefen(s, jetzt);
    return { ok: true, preis: preis, standorte: s.standorte };
  }

  function naechsteStufe(s) {
    return s.stufe < STUFEN.length ? STUFEN[s.stufe] : null;
  }

  function stufeKaufen(s, jetzt) {
    abrechnen(s, zahl(jetzt, Date.now()));
    var n = naechsteStufe(s);
    if (!n) return { ok: false, grund: "max" };
    if (s.geld + 1e-9 < n.kosten) return { ok: false, grund: "geld", preis: n.kosten };
    s.geld -= n.kosten;
    s.gesamtAusgaben += n.kosten;
    s.stufe = n.nr;
    erfolgePruefen(s, jetzt);
    return { ok: true, preis: n.kosten, stufe: n.nr, name: n.name };
  }

  function erfolgePruefen(s, jetzt) {
    var neu = [], i;
    for (i = 0; i < ERFOLGE.length; i++) {
      var a = ERFOLGE[i];
      if (s.erfolge[a.id]) continue;
      if (a.pruef(s)) {
        s.erfolge[a.id] = true;
        neu.push(a);
        chronik(s, "Erfolg: " + a.name, jetzt);
      }
    }
    return neu;
  }

  function chronik(s, text, jetzt) {
    if (!s.chronik) s.chronik = [];
    s.chronik.unshift({ t: zahl(jetzt, Date.now()), text: String(text) });
    if (s.chronik.length > 40) s.chronik.length = 40;
  }

  function ereignis(s, jetzt) {
    var t = zahl(jetzt, Date.now());
    abrechnen(s, t);
    if (t < zahl(s.naechstesEreignis, 0)) return null;
    s.naechstesEreignis = t + (70000 + Math.random() * 90000);
    var text = EREIGNISSE[Math.floor(Math.random() * EREIGNISSE.length)];
    var w = werte(s);
    var geld = 0;
    if (w.proSekunde > 0 && Math.random() < 0.6) {
      geld = runde(w.proSekunde * (20 + Math.random() * 70));
      if (geld < 5) geld = 5;
      s.geld += geld;
      s.gesamtEinnahmen += geld;
    }
    chronik(s, text + (geld ? " (+" + geld + " €)" : ""), t);
    erfolgePruefen(s, t);
    return { text: text, geld: geld };
  }

  /* ---------------------------------------------------------------- Speichern */
  function speicherText(s) {
    return JSON.stringify({
      v: VERSION,
      geld: runde(zahl(s.geld, 0)),
      gesamtEinnahmen: runde(zahl(s.gesamtEinnahmen, 0)),
      gesamtAusgaben: runde(zahl(s.gesamtAusgaben, 0)),
      personal: s.personal,
      upgrades: s.upgrades,
      erfolge: s.erfolge,
      kondome: s.kondome,
      kondomeGekauft: s.kondomeGekauft,
      stufe: s.stufe,
      standorte: s.standorte,
      klicks: s.klicks,
      gespielt: runde(zahl(s.gespielt, 0)),
      bereit: s.bereit,
      letzte: s.letzte,
      start: s.start,
      naechstesEreignis: s.naechstesEreignis,
      chronik: s.chronik
    });
  }

  function laden(roh) {
    var jetzt = Date.now();
    var s = neuerStand(jetzt);
    if (!roh) return s;
    var o;
    try { o = typeof roh === "string" ? JSON.parse(roh) : roh; } catch (fehler) { return s; }
    if (!o || typeof o !== "object") return s;

    s.geld = Math.max(0, zahl(o.geld, 0));
    s.gesamtEinnahmen = Math.max(0, zahl(o.gesamtEinnahmen, 0));
    s.gesamtAusgaben = Math.max(0, zahl(o.gesamtAusgaben, 0));
    s.kondome = ganz(o.kondome);
    s.kondomeGekauft = ganz(o.kondomeGekauft);
    s.klicks = ganz(o.klicks);
    s.gespielt = Math.max(0, zahl(o.gespielt, 0));
    s.stufe = Math.min(STUFEN.length, Math.max(1, ganz(o.stufe) || 1));
    s.standorte = ganz(o.standorte);
    s.start = zahl(o.start, jetzt);
    s.letzte = zahl(o.letzte, jetzt);
    s.naechstesEreignis = zahl(o.naechstesEreignis, jetzt + 120000);
    s.bereit = {
      hand: zahl(o.bereit && o.bereit.hand, 0),
      extra: zahl(o.bereit && o.bereit.extra, 0)
    };
    for (var i = 0; i < PERSONAL.length; i++) {
      var id = PERSONAL[i].id;
      s.personal[id] = ganz(o.personal && o.personal[id]);
    }
    for (i = 0; i < UPGRADES.length; i++) {
      var uid = UPGRADES[i].id;
      if (o.upgrades && o.upgrades[uid]) s.upgrades[uid] = true;
    }
    for (i = 0; i < ERFOLGE.length; i++) {
      var aid = ERFOLGE[i].id;
      if (o.erfolge && o.erfolge[aid]) s.erfolge[aid] = true;
    }
    if (Array.isArray(o.chronik)) {
      for (i = 0; i < o.chronik.length && i < 40; i++) {
        var z = o.chronik[i];
        if (z && typeof z.text === "string") s.chronik.push({ t: zahl(z.t, jetzt), text: z.text.slice(0, 300) });
      }
    }
    return s;
  }

  /* ------------------------------------------------------------- Darstellung */
  var EINHEITEN = [
    [1e24, "Qtl."], [1e21, "Trd."], [1e18, "Trl."], [1e15, "Brd."],
    [1e12, "Bio."], [1e9, "Mrd."], [1e6, "Mio."], [1e3, "Tsd."]
  ];

  function zahlText(x, stellen) {
    var n = Math.abs(x);
    var s = n.toFixed(stellen);
    var teile = s.split(".");
    teile[0] = teile[0].replace(/\B(?=(\d{3})+(?!\d))/g, ".");
    return teile.length > 1 ? teile[0] + "," + teile[1] : teile[0];
  }

  function formatGeld(x) {
    if (typeof x !== "number" || isNaN(x)) return "0 €";
    if (!isFinite(x)) return "∞ €";
    var vz = x < 0 ? "−" : "";
    var n = Math.abs(x);
    if (n < 1e3) return vz + zahlText(n, 2) + " €";
    if (n < 1e6) return vz + zahlText(n, 0) + " €";
    for (var i = 0; i < EINHEITEN.length; i++) {
      var wert = EINHEITEN[i][0], kurz = EINHEITEN[i][1];
      if (n >= wert) {
        var v = n / wert;
        var st = v < 10 ? 2 : (v < 100 ? 1 : 0);
        return vz + zahlText(v, st) + " " + kurz + " €";
      }
    }
    return vz + n.toExponential(2).replace(".", ",").replace("e+", " × 10^") + " €";
  }

  function formatZahl(x) {
    var n = Math.floor(Math.abs(zahl(x, 0)));
    var s = String(n).replace(/\B(?=(\d{3})+(?!\d))/g, ".");
    return (x < 0 ? "−" : "") + s;
  }

  function formatDauer(sekunden) {
    var s = Math.max(0, Math.round(zahl(sekunden, 0)));
    if (s < 60) return s + " s";
    var m = Math.floor(s / 60), r = s % 60;
    if (m < 60) return m + ":" + (r < 10 ? "0" : "") + r + " min";
    var h = Math.floor(m / 60);
    m = m % 60;
    if (h < 24) return h + " h " + (m < 10 ? "0" : "") + m + " min";
    var t = Math.floor(h / 24);
    return t + " T " + (h % 24) + " h";
  }

  /* ------------------------------------------------------------------ Export */
  return {
    VERSION: VERSION,
    STUNDE_SEK: STUNDE_SEK,
    KONDOM: KONDOM,
    STUFEN: STUFEN,
    PERSONAL: PERSONAL,
    UPGRADES: UPGRADES,
    ERFOLGE: ERFOLGE,
    EREIGNISSE: EREIGNISSE,
    neuerStand: neuerStand,
    werte: werte,
    abrechnen: abrechnen,
    aktionsBetrag: aktionsBetrag,
    klick: klick,
    kondomeKaufen: kondomeKaufen,
    personalDef: personalDef,
    personalKosten: personalKosten,
    maxPersonal: maxPersonal,
    personalKaufen: personalKaufen,
    upgradeDef: upgradeDef,
    upgradeKaufen: upgradeKaufen,
    standortKosten: standortKosten,
    standortKaufen: standortKaufen,
    naechsteStufe: naechsteStufe,
    stufeKaufen: stufeKaufen,
    erfolgePruefen: erfolgePruefen,
    ereignis: ereignis,
    anzahlPersonal: anzahlPersonal,
    anzahlUpgrades: anzahlUpgrades,
    speicherText: speicherText,
    laden: laden,
    formatGeld: formatGeld,
    formatZahl: formatZahl,
    formatDauer: formatDauer
  };
});
