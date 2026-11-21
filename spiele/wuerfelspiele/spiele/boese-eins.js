/* Die böse Eins – fünf Würfe, eine Eins beendet den Durchgang sofort. */
TGW.anmelden({
  id: "boese-eins",
  titel: "Die böse Eins",
  unter: "Ein Würfel · fünf Würfe · eine Eins stoppt sofort",
  spieler: { min: 1, max: 6, standard: 2 },
  kurz: "Fünf Würfe, alles wird addiert – aber eine Eins stoppt dich sofort.",
  erklaerung: {
    ziel: "In fünf Würfen so viele Augen wie möglich sammeln.",
    zug: "Wer am Zug ist, würfelt fünfmal hintereinander. Alle Augen kommen auf die eigene Summe.",
    wuerfel: "Ein Würfel; 2 bis 6 zählen voll zur Summe.",
    nochmal: "Nach jedem Wurf wird weitergewürfelt – bis fünf Würfe gespielt sind oder eine Eins fällt.",
    sieg: "Höchste Summe nach fünf Würfen.",
    besonders: [
      "Eine Eins beendet den Durchgang sofort; die bis dahin gesammelte Summe bleibt stehen.",
      "Wer eine Eins würfelt, steht oft schlechter da als jemand mit lauter kleinen Zahlen."
    ]
  },
  regeln: [
    "Jeder Spieler hat fünf Würfe.",
    "Alle gewürfelten Augen werden zusammengezählt.",
    "Eine Eins beendet den Durchgang sofort – die bis dahin erreichte Summe bleibt bestehen.",
    "Nach fünf Würfen ohne Eins zählt die volle Summe. Wer die höchste hat, gewinnt."
  ],
  start: function (z) {
    z.spieler.forEach(function (p) {
      p.summe = 0;
      p.wuerfe = 0;
      p.gebissen = false;
      p.fertig = false;
    });
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    return {
      wuerfel: z.wuerfel,
      summe: p.summe,
      summeText: "Summe",
      hinweis: "Wurf " + (p.wuerfe + 1) + " von 5 · bisher " + p.summe,
      klein: "Eine Eins beendet den Durchgang sofort.",
      knopf: { text: "Würfeln", art: "wurf" },
      spieler: z.spieler.map(function (q) {
        return {
          name: q.name,
          wert: q.summe,
          neben: q.fertig ? (q.gebissen ? "Eins – gestoppt" : "5 Würfe") : "Wurf " + q.wuerfe + "/5"
        };
      })
    };
  },
  aktion: function (z, art) {
    if (art !== "wurf") return false;
    var p = z.spieler[z.dran];
    if (p.fertig) return false;
    var wert = z.wurf(1)[0];
    p.wuerfe += 1;
    if (wert === 1) {
      p.gebissen = true;
      p.fertig = true;
      TGW.note(z, p.name + " würfelt eine Eins – " + p.summe + " Punkte");
    } else {
      p.summe += wert;
      TGW.note(z, p.name + " würfelt " + wert + " (" + p.summe + ")");
      if (p.wuerfe >= 5) p.fertig = true;
    }
    if (p.fertig) {
      if (TGW.endeNachHoechstem(z, function (q) { return q.summe; }, "Höchste Summe")) return;
      TGW.naechster(z);
    }
  }
});
