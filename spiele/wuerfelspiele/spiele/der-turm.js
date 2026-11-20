/* Der Turm – eigene Stichzahl ermitteln, dann Treffer in Turmsteine tauschen. */
TGW.anmelden({
  id: "der-turm",
  titel: "Der Turm",
  unter: "Drei Würfel · eigene Stichzahl · gemeinsamer Turm",
  spieler: { min: 2, max: 6, standard: 2 },
  regeln: [
    "Zuerst würfelt jeder Spieler einmal mit einem Würfel – das ist seine persönliche Stichzahl.",
    "Aus allen Steinen wird ein gemeinsamer Turm gebaut: drei Steine je Spieler.",
    "Wer am Zug ist, würfelt mit drei Würfeln. Jeder Würfel, der die eigene Stichzahl zeigt, ist ein Treffer.",
    "Solange der Turm steht, nimmt man für jeden Treffer einen Stein aus dem Turm zu sich.",
    "Ist der Turm leer, gibt man für jeden Treffer einen eigenen Stein an den nächsten Spieler weiter.",
    "Wer keinen Stein mehr hat, während der Turm leer ist, gewinnt."
  ],
  start: function (z) {
    z.spieler.forEach(function (p) { p.stich = null; p.steine = 0; });
    z.turm = 0;
    z.phase = "stich";
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var stichphase = z.phase === "stich";
    return {
      wuerfel: z.wuerfel,
      summe: stichphase ? null : z.turm,
      summeText: "im Turm",
      hinweis: stichphase
        ? p.name + " ermittelt die Stichzahl"
        : "Deine Stichzahl: " + p.stich + " · Turm " + z.turm,
      klein: stichphase
        ? "Ein Wurf mit einem Würfel je Spieler."
        : (z.turm > 0 ? "Treffer nehmen Steine aus dem Turm." : "Turm leer: Treffer geben Steine weiter."),
      knopf: stichphase ? { text: "Stichzahl würfeln", art: "wurf" } : { text: "Würfeln", art: "wurf" },
      spieler: z.spieler.map(function (q) {
        return {
          name: q.name,
          wert: q.stich === null ? "—" : q.stich,
          neben: q.stich === null ? "ohne Stichzahl" : (q.steine + " Steine")
        };
      }),
      tabelle: {
        titel: "Turm",
        spalten: ["Spieler", "Stichzahl", "Steine"],
        zeilen: z.spieler.map(function (q) {
          return [q.name, q.stich === null ? "—" : q.stich, q.steine];
        })
      }
    };
  },
  aktion: function (z, art) {
    if (art !== "wurf") return false;
    var p = z.spieler[z.dran];
    if (z.phase === "stich") {
      var zahl = z.wurf(1)[0];
      p.stich = zahl;
      TGW.note(z, p.name + " hat die Stichzahl " + zahl);
      TGW.naechster(z);
      if (z.spieler.every(function (q) { return q.stich !== null; })) {
        z.phase = "spiel";
        z.turm = 3 * z.spieler.length;
        z.dran = 0;
        TGW.note(z, "Der Turm hat " + z.turm + " Steine");
      }
      return;
    }
    var w = z.wurf(3);
    var treffer = w.filter(function (x) { return x === p.stich; }).length;
    if (z.turm > 0) {
      var nehmen = Math.min(treffer, z.turm);
      z.turm -= nehmen;
      p.steine += nehmen;
      TGW.note(z, p.name + ": " + treffer + " Treffer · " + nehmen + " Steine aus dem Turm");
    } else {
      var geben = Math.min(treffer, p.steine);
      p.steine -= geben;
      var naechster = z.spieler[(z.dran + 1) % z.spieler.length];
      if (naechster !== p) naechster.steine += geben;
      TGW.note(z, p.name + ": " + treffer + " Treffer · " + geben + " Steine weiter");
    }
    if (z.turm === 0 && p.steine === 0) {
      TGW.sieger(z, [z.dran], "Turm leer und keine Steine mehr.");
      return;
    }
    TGW.naechster(z);
  }
});
