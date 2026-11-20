/* Hin und her – gewürfelte Zahl umschalten, alle sechs gleichzeitig gestrichen. */
TGW.anmelden({
  id: "hin-und-her",
  titel: "Hin und her",
  unter: "Ein Würfel · Zahlen 1–6 umschalten",
  spieler: { min: 1, max: 6, standard: 2 },
  regeln: [
    "Jeder Spieler beginnt mit den Zahlen 1 bis 6, alle sind offen.",
    "Wer am Zug ist, würfelt einmal.",
    "Die gewürfelte Zahl wird umgeschaltet: offen wird gestrichen, gestrichen wird wieder offen.",
    "Wer nach seinem Wurf alle sechs Zahlen gleichzeitig gestrichen hat, gewinnt."
  ],
  start: function (z) {
    z.spieler.forEach(function (p) {
      p.aus = [false, false, false, false, false, false];
      p.fertig = false;
    });
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    function gestrichen(q) { return q.aus.filter(Boolean).length; }
    return {
      wuerfel: z.wuerfel,
      hinweis: "Würfeln und eine Zahl umschalten",
      klein: "Gestrichen: " + gestrichen(p) + " von 6",
      knopf: { text: "Würfeln", art: "wurf" },
      leisten: z.spieler.map(function (q) {
        return {
          titel: q.name + " – " + gestrichen(q) + "/6",
          zahlen: q.aus.map(function (aus, i) { return { wert: i + 1, aus: aus }; })
        };
      }),
      spieler: z.spieler.map(function (q) {
        return { name: q.name, wert: gestrichen(q) + "/6", neben: q.fertig ? "durch" : "" };
      })
    };
  },
  aktion: function (z, art) {
    if (art !== "wurf") return false;
    var p = z.spieler[z.dran];
    var wert = z.wurf(1)[0];
    p.aus[wert - 1] = !p.aus[wert - 1];
    var anzahl = p.aus.filter(Boolean).length;
    TGW.note(z, p.name + " würfelt " + wert + " – " + (p.aus[wert - 1] ? "gestrichen" : "wieder offen") + " (" + anzahl + "/6)");
    if (anzahl === 6) {
      p.fertig = true;
      TGW.sieger(z, [z.dran], "Alle sechs Zahlen gleichzeitig gestrichen.");
      return;
    }
    TGW.naechster(z);
  }
});
