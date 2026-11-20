/* Dufte Hundert – zwei Würfel, Pasch wird multipliziert, erster mit 100 Punkten. */
TGW.anmelden({
  id: "dufte-hundert",
  titel: "Dufte Hundert",
  unter: "Zwei Würfel · Pasch zählt doppelt · Ziel 100 Punkte",
  spieler: { min: 1, max: 6, standard: 2 },
  regeln: [
    "Wer am Zug ist, würfelt mit zwei Würfeln.",
    "Ohne Pasch werden die beiden Augenzahlen addiert.",
    "Bei einem Pasch werden die beiden Augenzahlen multipliziert (4 und 4 ergeben also 16).",
    "Die Punkte werden addiert; wer zuerst 100 Punkte erreicht oder überschreitet, gewinnt."
  ],
  ZIEL: 100,
  start: function (z) {
    z.spieler.forEach(function (p) { p.punkte = 0; });
  },
  sicht: function (z) {
    var werte = z.wuerfel || [];
    var pasch = werte.length === 2 && werte[0] === werte[1];
    var wert = werte.length === 2 ? (pasch ? werte[0] * werte[1] : werte[0] + werte[1]) : 0;
    return {
      wuerfel: z.wuerfel,
      summe: z.wuerfel.length === 2 ? wert : null,
      summeText: pasch ? "Pasch ×" : "Augen",
      hinweis: "Würfeln – Ziel ist " + this.ZIEL + " Punkte",
      klein: "Pasch wird multipliziert, sonst wird addiert.",
      knopf: { text: "Würfeln", art: "wurf" },
      tabelle: {
        titel: "Punkte",
        spalten: ["Spieler", "Punkte"],
        zeilen: z.spieler.map(function (q) { return [q.name, q.punkte]; })
      }
    };
  },
  aktion: function (z, art) {
    if (art !== "wurf") return false;
    var p = z.spieler[z.dran];
    var w = z.wurf(2);
    var pasch = w[0] === w[1];
    var wert = pasch ? w[0] * w[1] : w[0] + w[1];
    p.punkte += wert;
    TGW.note(z, p.name + ": " + w[0] + (pasch ? " × " : " + ") + w[1] + " = " + wert + " → " + p.punkte);
    if (p.punkte >= this.ZIEL) {
      TGW.sieger(z, [z.dran], p.punkte + " Punkte erreicht.");
      return;
    }
    TGW.naechster(z);
  }
});
