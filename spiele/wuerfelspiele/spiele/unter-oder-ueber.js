/* Unter oder über sieben – Ansage vor dem Wurf, virtuelle Punkte, keine Wetten. */
TGW.anmelden({
  id: "unter-oder-ueber",
  titel: "Unter oder über sieben",
  unter: "Zwei Würfel · Ansage vor dem Wurf · nur Spielpunkte",
  spieler: { min: 1, max: 6, standard: 2 },
  regeln: [
    "Jeder Spieler startet mit 50 virtuellen Punkten.",
    "Vor jedem Wurf wird angesagt: unter sieben (2–6), über sieben (8–12) oder genau sieben.",
    "Ein richtiger Tipp bringt 5 Punkte, ein falscher kostet 5.",
    "Genau sieben ist der Sonderfall: Der Einsatz ist verdoppelt – richtig bringt 10 Punkte, falsch kostet 10.",
    "Wer auf unter oder über setzt und eine Sieben würfelt, verliert ebenfalls 10 Punkte.",
    "Es wird ausschließlich mit Spielpunkten gespielt, niemals mit Geld. Nach acht Zügen je Spieler gewinnt die höchste Punktzahl."
  ],
  START: 50,
  ZUEGE: 8,
  start: function (z) {
    z.spieler.forEach(function (p) { p.punkte = this.START; p.zuege = 0; p.tipp = null; p.fertig = false; }, this);
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var namen = { unter: "unter sieben", ueber: "über sieben", sieben: "genau sieben" };
    return {
      wuerfel: z.wuerfel,
      summe: z.wuerfel.length === 2 ? z.wuerfel[0] + z.wuerfel[1] : null,
      summeText: "Summe",
      hinweis: p.tipp ? "Tipp: " + namen[p.tipp] : "Was sagst du voraus?",
      klein: "Zug " + (p.zuege + 1) + " von " + this.ZUEGE + " · Sieben zählt doppelt",
      knopf: { text: "Würfeln", art: "wurf", an: !!p.tipp },
      auswahl: [
        { text: "Unter sieben (+5)", art: "ansage", daten: { tipp: "unter" } },
        { text: "Über sieben (+5)", art: "ansage", daten: { tipp: "ueber" } },
        { text: "Genau sieben (±10)", art: "ansage", daten: { tipp: "sieben" } }
      ],
      tabelle: {
        titel: "Punkte",
        spalten: ["Spieler", "Züge", "Punkte"],
        zeilen: z.spieler.map(function (q) { return [q.name, q.zuege + "/" + this.ZUEGE, q.punkte]; }, this)
      }
    };
  },
  aktion: function (z, art, daten) {
    var p = z.spieler[z.dran];
    if (art === "ansage") {
      p.tipp = daten.tipp;
      return;
    }
    if (art !== "wurf" || !p.tipp) return false;
    var w = z.wurf(2);
    var summe = w[0] + w[1];
    var tipp = p.tipp;
    var punkte = 0;
    if (summe === 7) punkte = tipp === "sieben" ? 10 : -10;
    else if (summe < 7) punkte = tipp === "unter" ? 5 : (tipp === "ueber" ? -5 : -10);
    else punkte = tipp === "ueber" ? 5 : (tipp === "unter" ? -5 : -10);
    p.punkte += punkte;
    p.zuege += 1;
    TGW.note(z, p.name + ": " + w[0] + "+" + w[1] + " = " + summe + " · " + (punkte > 0 ? "+" : "") + punkte + " (" + p.punkte + ")");
    p.tipp = null;
    if (p.zuege >= this.ZUEGE) p.fertig = true;
    if (TGW.endeNachHoechstem(z, function (q) { return q.punkte; }, "Höchste Punktzahl")) return;
    TGW.naechster(z);
  }
});
