/* Jule – erst 1 bis 6, dann 6 bis 1. Treffer: weiterwürfeln. Fehlwurf: nächster. */
TGW.anmelden({
  id: "jule",
  titel: "Jule",
  unter: "Ein Würfel · 1 bis 6 vorwärts, dann 6 bis 1 rückwärts",
  spieler: { min: 1, max: 6, standard: 2 },
  varianten: [{ id: "stumm", text: "Stumme Jule (es wird nicht gesprochen)" }],
  regeln: [
    "Gebraucht werden der Reihe nach die Zahlen 1, 2, 3, 4, 5, 6 – danach 6, 5, 4, 3, 2, 1.",
    "Wer am Zug ist, würfelt einmal.",
    "Trifft der Wurf die gesuchte Zahl, geht es sofort mit dem nächsten Wurf weiter.",
    "Ein Fehlwurf beendet den Zug, der Nächste ist dran.",
    "Wer zuerst alle zwölf Stufen geschafft hat, gewinnt.",
    "Stumme Jule: Während der ganzen Partie wird nicht gesprochen – der Verlauf bleibt aus."
  ],
  ZIELE: [1, 2, 3, 4, 5, 6, 6, 5, 4, 3, 2, 1],
  start: function (z) {
    z.spieler.forEach(function (p) { p.stufe = 0; p.fertig = false; });
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var stumm = !!(z.varianten && z.varianten.stumm);
    var ziel = this.ZIELE[Math.min(p.stufe, this.ZIELE.length - 1)];
    return {
      wuerfel: z.wuerfel,
      hinweis: p.fertig ? "Fertig" : "Gesucht: " + ziel,
      klein: "Stufe " + p.stufe + " von 12" + (stumm ? " · Stumme Jule: bitte nicht sprechen" : ""),
      knopf: { text: "Würfeln", art: "wurf" },
      verlauf: stumm ? [] : z.log,
      spieler: z.spieler.map(function (q) {
        return {
          name: q.name,
          wert: q.stufe + "/12",
          neben: q.fertig ? "durch" : "sucht " + this.ZIELE[Math.min(q.stufe, 11)]
        };
      }, this)
    };
  },
  aktion: function (z, art) {
    if (art !== "wurf") return false;
    var p = z.spieler[z.dran];
    if (p.fertig) return false;
    var wert = z.wurf(1)[0];
    var ziel = this.ZIELE[p.stufe];
    if (wert === ziel) {
      p.stufe += 1;
      TGW.note(z, p.name + " trifft die " + ziel);
      if (p.stufe >= this.ZIELE.length) {
        p.fertig = true;
        TGW.sieger(z, [z.dran], "Alle zwölf Stufen in einem Zug zu Ende gebracht.");
        return;
      }
      return;                       // Treffer: derselbe Spieler würfelt weiter
    }
    TGW.note(z, p.name + " würfelt " + wert + " statt " + ziel);
    TGW.naechster(z);
  }
});
