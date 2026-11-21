/* Rentmeister – sechs Zahlen von 3 bis 8, jede mit einem Wurf multipliziert.
   Der Ablauf „würfeln, dann einen Würfel wählen“ steckt in TGW.wurfDannWahl. */
TGW.anmelden(TGW.wurfDannWahl({
  id: "rentmeister",
  titel: "Rentmeister",
  unter: "Drei Würfel · sechs Würfe · Zahlen 3 bis 8",
  kurz: "Sechs feste Zahlen von 3 bis 8, jede mal einem gewählten Würfel – Summe sammeln.",
  erklaerung: {
    ziel: "Aus sechs Würfen die höchste Gesamtsumme holen.",
    zug: "Die festen Zahlen 3, 4, 5, 6, 7 und 8 werden der Reihe nach gespielt. Für jede wird mit drei Würfeln gewürfelt, dann genau ein Würfel gewählt.",
    wuerfel: "Drei Würfel. Der gewählte Würfel wird mit der festen Zahl multipliziert (bei der 5 also 5 × 4 = 20).",
    nochmal: "Nein – jede Zahl kommt genau einmal, danach ist der Nächste dran.",
    sieg: "Höchste Gesamtsumme nach allen sechs Zahlen.",
    besonders: [
      "Bei großen Zahlen lohnt sich ein hoher Würfel: 8 × 6 = 48 Punkte.",
      "Die gewählte Zahl wird sofort addiert – man sieht nach jedem Wurf den neuen Stand."
    ]
  },
  regeln: [
    "Die festen Zahlen 3, 4, 5, 6, 7 und 8 werden der Reihe nach abgearbeitet.",
    "Für jede Zahl wird mit drei Würfeln gewürfelt; danach wird genau ein Würfel ausgewählt.",
    "Die feste Zahl wird mit dem gewählten Würfel multipliziert; das Ergebnis wird addiert.",
    "Nach sechs Würfen steht die Gesamtsumme. Wer die höchste hat, gewinnt."
  ],
  ZAHLEN: [3, 4, 5, 6, 7, 8],
  start: function (z) {
    z.spieler.forEach(function (p) { p.stelle = 0; p.summe = 0; p.fertig = false; });
  },
  zahl: function (p) {
    return this.ZAHLEN[Math.min(p.stelle, this.ZAHLEN.length - 1)];
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var zahl = this.zahl(p);
    return {
      wuerfel: z.wuerfel,
      summe: p.summe,
      summeText: "Summe",
      hinweis: "Zahl " + zahl + " · Wurf " + (p.stelle + 1) + " von 6",
      klein: "Einen Würfel auswählen: " + zahl + " × Zahl",
      tabelle: {
        titel: "Stände",
        spalten: ["Spieler", "Zahl", "Summe"],
        zeilen: z.spieler.map(function (q) {
          return [q.name, q.fertig ? "fertig" : this.zahl(q), q.summe];
        }, this)
      }
    };
  },
  wahl: function (z) {
    var zahl = this.zahl(z.spieler[z.dran]);
    return z.wuerfel.map(function (w, i) {
      return { text: zahl + " × " + w + " = " + zahl * w, art: "waehlen", daten: { index: i } };
    });
  },
  wurf: function (z) {
    z.wurf(3);
  },
  setze: function (z, daten) {
    var p = z.spieler[z.dran];
    var zahl = this.zahl(p);
    var wert = zahl * z.wuerfel[daten.index];
    p.summe += wert;
    TGW.note(z, p.name + ": " + zahl + " × " + z.wuerfel[daten.index] + " = " + wert + " (" + p.summe + ")");
    p.stelle += 1;
    p.punkte = p.summe;
    if (p.stelle < this.ZAHLEN.length) return;
    p.fertig = true;
    if (TGW.endeNachHoechstem(z, function (q) { return q.summe; }, "Höchste Gesamtsumme")) return;
    TGW.naechster(z);
  }
}));
