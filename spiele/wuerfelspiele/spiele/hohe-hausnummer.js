/* Hohe Hausnummer – ein Würfel, drei Würfe, höchste dreistellige Zahl.
   Der Ablauf „würfeln, dann eine Stelle belegen“ steckt in TGW.wurfDannWahl. */
TGW.anmelden(TGW.wurfDannWahl({
  id: "hohe-hausnummer",
  titel: "Hohe Hausnummer",
  unter: "Ein Würfel · drei Würfe · höchste dreistellige Zahl",
  regeln: [
    "Jeder Spieler würfelt dreimal.",
    "Nach jedem Wurf muss die Zahl einer noch freien Stelle zugeordnet werden: Hunderter, Zehner oder Einer.",
    "Die drei Stellen ergeben zusammen die Hausnummer. Eine gesetzte Stelle bleibt stehen.",
    "Wer am Ende die höchste Hausnummer hat, gewinnt."
  ],
  STELLEN: ["Hunderter", "Zehner", "Einer"],
  start: function (z) {
    z.spieler.forEach(function (p) {
      p.stellen = [null, null, null];
      p.wurfwert = 0;
      p.wuerfe = 0;
      p.zahl = null;
      p.fertig = false;
    });
  },
  frei: function (p) {
    var liste = [];
    p.stellen.forEach(function (wert, i) { if (wert === null) liste.push(i); });
    return liste;
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var frei = this.frei(p);
    return {
      wuerfel: z.wuerfel,
      hinweis: z.phase === "waehlen" ? "Wohin mit der " + p.wurfwert + "?" : "Wurf " + (p.wuerfe + 1) + " von 3",
      klein: z.phase === "waehlen"
        ? "Noch frei: " + frei.map(function (i) { return this.STELLEN[i]; }, this).join(", ")
        : "Nach jedem Wurf eine Stelle belegen.",
      spieler: z.spieler.map(function (q) {
        return {
          name: q.name,
          wert: q.zahl === null ? "—" : q.zahl,
          neben: q.fertig ? "fertig" : "Wurf " + q.wuerfe + "/3"
        };
      }),
      tabelle: {
        titel: "Hausnummern",
        spalten: ["Spieler", "H", "Z", "E", "Zahl"],
        zeilen: z.spieler.map(function (q) {
          return [q.name].concat(q.stellen.map(function (w) { return w === null ? "·" : w; }))
            .concat([q.zahl === null ? "—" : q.zahl]);
        })
      }
    };
  },
  wahl: function (z) {
    var p = z.spieler[z.dran];
    return this.frei(p).map(function (i) {
      return { text: this.STELLEN[i] + ": " + p.wurfwert, art: "waehlen", daten: { stelle: i } };
    }, this);
  },
  wurf: function (z) {
    var p = z.spieler[z.dran];
    p.wurfwert = z.wurf(1)[0];
    p.wuerfe += 1;
  },
  setze: function (z, daten) {
    var p = z.spieler[z.dran];
    var stelle = daten.stelle;
    if (!(stelle >= 0 && stelle <= 2) || p.stellen[stelle] !== null) return false;
    p.stellen[stelle] = p.wurfwert;
    TGW.note(z, p.name + " setzt " + p.wurfwert + " auf " + this.STELLEN[stelle]);
    if (p.stellen.indexOf(null) >= 0) return;
    p.zahl = Number(p.stellen.join(""));
    p.fertig = true;
    if (TGW.endeNachHoechstem(z, function (q) { return q.zahl; }, "Höchste Hausnummer")) return;
    TGW.naechster(z);
  }
}));
