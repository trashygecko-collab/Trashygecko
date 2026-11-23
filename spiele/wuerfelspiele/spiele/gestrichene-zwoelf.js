/* Gestrichene Zwölf – Zahlen 2 bis 12 abstreichen, die Sieben zählt immer. */
TGW.anmelden({
  id: "gestrichene-zwoelf",
  titel: "Gestrichene Zwölf",
  unter: "Zwei Würfel · Zahlen 2–12 · drei Runden mit Bonus und Minus",
  spieler: { min: 1, max: 6, standard: 2 },
  kurz: "Würfle Summen von 2 bis 12 und streiche sie; die Sieben zählt immer.",
  erklaerung: {
    ziel: "Nach drei Runden die meisten Punkte haben.",
    zug: "Zwei Würfel werfen; die Summe bestimmt eine Zahl auf dem eigenen Blatt.",
    wuerfel: "Summen von 2 bis 12. Eine offene Zahl wird gestrichen und bringt ihre Zahl als Punkte. Die Sieben wird nie gestrichen und bringt bei jedem Wurf 7 Punkte.",
    nochmal: "Nein, ein Wurf pro Zug – sechs Züge je Spieler und Runde.",
    sieg: "Höchste Gesamtpunktzahl nach drei Runden.",
    besonders: [
      "Eine schon gestrichene Zahl (außer der Sieben) kostet 2 Minuspunkte.",
      "Am Rundenende: Sind alle zehn Zahlen von 2–6 und 8–12 gestrichen, gibt es 20 Bonuspunkte.",
      "Zu Beginn jeder Runde wird das Blatt neu ausgelegt, die Punkte bleiben."
    ]
  },
  regeln: [
    "Jeder Spieler hat ein Blatt mit den Zahlen 2 bis 12.",
    "Wer am Zug ist, würfelt zwei Würfel; die Summe bestimmt die Zahl.",
    "Eine noch offene Zahl wird gestrichen und zählt ihre Punkte.",
    "Die Sieben darf mehrfach gewertet werden – sie bringt bei jedem Wurf 7 Punkte.",
    "Eine bereits gestrichene Zahl (außer der Sieben) kostet 2 Minuspunkte.",
    "Eine Runde dauert sechs Züge je Spieler. Am Rundenende gibt jede vollständig gestrichene Zwölf 20 Bonuspunkte.",
    "Nach drei Runden gewinnt die höchste Gesamtpunktzahl."
  ],
  RUNDEN: 3,
  ZUEGE: 6,
  BEREICH: [2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12],
  start: function (z) {
    z.spieler.forEach(function (p) { this.blattLeeren(p); p.punkte = 0; p.zuege = 0; }, this);
    z.runde = 1;
  },
  blattLeeren: function (p) {
    p.aus = {};
    this.BEREICH.forEach(function (n) { p.aus[n] = false; });
  },
  vollstaendig: function (p) {
    return this.BEREICH.every(function (n) { return n === 7 || p.aus[n]; });
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var zahlen = this.BEREICH.map(function (n) {
      return { wert: n, aus: n === 7 ? false : !!p.aus[n], frisch: n === 7 };
    });
    return {
      wuerfel: z.wuerfel,
      summe: z.wuerfel.length === 2 ? z.wuerfel[0] + z.wuerfel[1] : null,
      summeText: "Summe",
      hinweis: "Runde " + z.runde + " von " + this.RUNDEN + " · Zug " + (p.zuege + 1) + " von " + this.ZUEGE,
      klein: "Die Sieben zählt immer. Schon gestrichene Zahlen kosten 2 Punkte.",
      knopf: { text: "Würfeln", art: "wurf" },
      leisten: [{ titel: p.name + " – Blatt", zahlen: zahlen }],
      tabelle: {
        titel: "Punkte",
        spalten: ["Spieler", "Gestrichen", "Punkte"],
        zeilen: z.spieler.map(function (q) {
          var anzahl = this.BEREICH.filter(function (n) { return n !== 7 && q.aus[n]; }).length;
          return [q.name, anzahl + "/10", q.punkte];
        }, this)
      }
    };
  },
  aktion: function (z, art) {
    if (art !== "wurf") return false;
    var p = z.spieler[z.dran];
    var summe = z.wurf(2)[0] + z.wuerfel[1];
    var schon = summe !== 7 && !!p.aus[summe];
    if (summe === 7) {
      p.punkte += 7;
      TGW.note(z, p.name + " würfelt die Sieben: +7 (" + p.punkte + ")");
    } else if (!p.aus[summe]) {
      p.aus[summe] = true;
      p.punkte += summe;
      TGW.note(z, p.name + " streicht " + summe + ": +" + summe + " (" + p.punkte + ")");
    } else {
      p.punkte -= 2;
      TGW.note(z, p.name + " hat die " + summe + " schon: −2 (" + p.punkte + ")");
    }
    p.zuege += 1;
    TGW.naechster(z);
    if (schon && !z.ende) {
      TGW.meldung(z, p.name + " hat die " + summe +
        " schon gestrichen – 2 Punkte weg. Jetzt ist " +
        z.spieler[z.dran].name + " dran.");
    }
    if (this.rundeZuEnde(z)) return;
  },
  rundeZuEnde: function (z) {
    if (!z.spieler.every(function (q) { return q.zuege >= this.ZUEGE; }, this)) return false;
    z.spieler.forEach(function (q) {
      if (this.vollstaendig(q)) {
        q.punkte += 20;
        TGW.note(z, q.name + " hat die Zwölf voll: +20 Bonus");
      }
    }, this);
    if (z.runde >= this.RUNDEN) {
      var r = TGW.hoechste(z.spieler.map(function (q) { return q.punkte; }));
      TGW.sieger(z, r.siege, "Nach drei Runden: " + r.max + " Punkte");
      return true;
    }
    z.runde += 1;
    z.spieler.forEach(function (q) { this.blattLeeren(q); q.zuege = 0; }, this);
    z.dran = 0;
    TGW.note(z, "Runde " + z.runde + " beginnt");
    return false;
  }
});
