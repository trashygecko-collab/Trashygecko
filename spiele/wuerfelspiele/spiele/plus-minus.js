/* Plus-Minus – neun Würfe, Muster + + − + + − + + −. */
TGW.anmelden({
  id: "plus-minus",
  titel: "Plus-Minus",
  unter: "Ein Würfel · neun Würfe · plus, plus, minus",
  spieler: { min: 1, max: 6, standard: 2 },
  regeln: [
    "Jeder Spieler würfelt neunmal.",
    "Der erste und der zweite Wurf werden addiert, der dritte wird abgezogen.",
    "Danach wiederholt sich das Muster: plus, plus, minus – bis alle neun Würfe gespielt sind.",
    "Die höchste Endsumme gewinnt; Zwischenstände dürfen negativ werden."
  ],
  VORZEICHEN: [1, 1, -1, 1, 1, -1, 1, 1, -1],
  start: function (z) {
    z.spieler.forEach(function (p) { p.summe = 0; p.rolle = 0; p.fertig = false; });
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var v = this.VORZEICHEN[Math.min(p.rolle, 8)];
    return {
      wuerfel: z.wuerfel,
      summe: p.summe,
      summeText: "Stand",
      hinweis: "Wurf " + (p.rolle + 1) + " von 9 · " + (v > 0 ? "wird addiert" : "wird abgezogen"),
      klein: "Muster: + + − + + − + + −",
      knopf: { text: "Würfeln", art: "wurf" },
      tabelle: {
        titel: "Stände",
        spalten: ["Spieler", "Würfe", "Summe"],
        zeilen: z.spieler.map(function (q) { return [q.name, q.rolle + "/9", q.summe]; })
      }
    };
  },
  aktion: function (z, art) {
    if (art !== "wurf") return false;
    var p = z.spieler[z.dran];
    if (p.fertig) return false;
    var wert = z.wurf(1)[0];
    var vorzeichen = this.VORZEICHEN[p.rolle];
    p.summe += vorzeichen * wert;
    p.rolle += 1;
    TGW.note(z, p.name + ": " + (vorzeichen > 0 ? "+" : "−") + wert + " → " + p.summe);
    if (p.rolle >= this.VORZEICHEN.length) {
      p.fertig = true;
      p.punkte = p.summe;
      if (TGW.endeNachHoechstem(z, function (q) { return q.summe; }, "Höchste Summe")) return;
      TGW.naechster(z);
    }
  }
});
