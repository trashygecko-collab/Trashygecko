/* Fenster würfeln – drei Würfel, nur Zweien zählen (1 / 4 / 9), Ziel 30. */
TGW.anmelden({
  id: "fenster",
  titel: "Fenster würfeln",
  unter: "Drei Würfel · bis zu drei Würfe · nur Zweien zählen · Ziel 30",
  spieler: { min: 1, max: 6, standard: 2 },
  regeln: [
    "Wer am Zug ist, hat bis zu drei Würfe mit drei Würfeln.",
    "Es zählen nur Zweien: eine einzelne Zwei bringt 1 Punkt, ein Zweierpasch 4 Punkte, drei Zweien 9 Punkte.",
    "Alle anderen Augenzahlen bringen nichts.",
    "Nach jedem Wurf darf man aufhören („Halten“) oder weiterwürfeln.",
    "Wer zuerst 30 Zweierpunkte erreicht, gewinnt."
  ],
  ZIEL: 30,
  wert: function (w) {
    return [0, 1, 4, 9][w.filter(function (x) { return x === 2; }).length];
  },
  start: function (z) {
    z.spieler.forEach(function (p) { p.punkte = 0; p.wuerfe = 0; });
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var wert = z.wuerfel.length === 3 ? this.wert(z.wuerfel) : null;
    return {
      wuerfel: z.wuerfel,
      summe: wert,
      summeText: "Zweier",
      hinweis: "Wurf " + (p.wuerfe + 1) + " von 3 · " + p.punkte + " Punkte",
      klein: "Eine Zwei = 1 · Pasch = 4 · drei Zweien = 9",
      knopf: { text: "Würfeln", art: "wurf", an: p.wuerfe < 3 },
      knopf2: z.wuerfel.length ? { text: "Halten", art: "halten" } : null,
      tabelle: {
        titel: "Zweierpunkte",
        spalten: ["Spieler", "Punkte"],
        zeilen: z.spieler.map(function (q) { return [q.name, q.punkte]; })
      }
    };
  },
  aktion: function (z, art) {
    var p = z.spieler[z.dran];
    if (art === "halten") {
      p.wuerfe = 0;
      z.wuerfel = [];
      TGW.naechster(z);
      return;
    }
    if (art !== "wurf" || p.wuerfe >= 3) return false;
    var w = z.wurf(3);
    var wert = this.wert(w);
    p.wuerfe += 1;
    p.punkte += wert;
    TGW.note(z, p.name + ": " + w.join(" ") + " · " + (wert ? "+" + wert : "nichts") + " (" + p.punkte + ")");
    if (p.punkte >= this.ZIEL) {
      TGW.sieger(z, [z.dran], p.punkte + " Zweierpunkte erreicht.");
      return;
    }
    if (p.wuerfe >= 3) {
      p.wuerfe = 0;
      TGW.naechster(z);
    }
  }
});
