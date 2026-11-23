/* Streichhölzer auswürfeln – zehn Hölzer, Einsen wandern nach links, Sechsen in die Mitte. */
TGW.anmelden({
  id: "streichhoelzer",
  titel: "Streichhölzer auswürfeln",
  unter: "Drei Würfel · zehn Hölzer je Spieler · Einsen und Sechsen wandern",
  spieler: { min: 2, max: 6, standard: 3 },
  kurz: "Zehn Hölzer: Einsen wandern nach links, Sechsen in die Mitte. Wer zuletzt Hölzer hat, gewinnt.",
  erklaerung: {
    ziel: "Als Letzter noch Streichhölzer haben. Jeder startet mit zehn.",
    zug: "Mit drei Würfeln werfen: Jede Eins gibt ein Hölzchen an den nächsten Spieler, jede Sechs wandert in die Mitte.",
    wuerfel: "Eins = ein Hölzchen nach links (an den Nächsten in der Runde). Sechs = ein Hölzchen in die Mitte, also aus dem Spiel. 2 bis 5 bewirken nichts.",
    nochmal: "Nein, ein Wurf pro Zug.",
    sieg: "Wer als Einziger noch Hölzer hat.",
    besonders: [
      "Wer keine Hölzer mehr hat, scheidet aus.",
      "Kann man nicht mehr genug abgeben, wandert der Rest in die Mitte.",
      "Die Mitte ist der Ablageplatz für alle Hölzer, die aus dem Spiel sind."
    ]
  },
  regeln: [
    "Jeder Spieler beginnt mit zehn virtuellen Streichhölzern.",
    "Wer am Zug ist, würfelt mit drei Würfeln.",
    "Jede Eins gibt ein Hölzchen an den nächsten Spieler in der Runde (den linken Nachbarn am Tisch).",
    "Jede Sechs wandert in die gemeinsame Mitte und ist aus dem Spiel.",
    "Die übrigen Zahlen haben keine Wirkung.",
    "Wer keine Hölzer mehr hat, scheidet aus – kann aber nichts mehr abgeben, solche Hölzer gehen in die Mitte.",
    "Wer als Letzter noch Hölzer hat, gewinnt."
  ],
  START: 10,
  start: function (z) {
    z.spieler.forEach(function (p) { p.hoelzer = this.START; p.raus = false; }, this);
    z.mitte = 0;
  },
  naechsterAktiver: function (z) {
    for (var i = 0; i < z.spieler.length; i++) {
      TGW.naechster(z);
      if (!z.spieler[z.dran].raus) return;
    }
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    return {
      wuerfel: z.wuerfel,
      summe: z.mitte,
      summeText: "in der Mitte",
      hinweis: "Drei Würfel – Einsen wandern nach links, Sechsen in die Mitte",
      klein: p.name + " hat " + p.hoelzer + " Hölzer",
      knopf: p.raus ? null : { text: "Würfeln", art: "wurf" },
      spieler: z.spieler.map(function (q) {
        return {
          name: q.name,
          wert: q.raus ? "raus" : q.hoelzer,
          neben: q.raus ? "keine Hölzer" : (q.hoelzer + " Hölzer")
        };
      }),
      tabelle: {
        titel: "Hölzer",
        spalten: ["Spieler", "Hölzer"],
        zeilen: z.spieler.map(function (q) { return [q.name, q.raus ? "ausgeschieden" : q.hoelzer]; })
      }
    };
  },
  aktion: function (z, art) {
    if (art !== "wurf") return false;
    var p = z.spieler[z.dran];
    if (p.raus) return false;
    var w = z.wurf(3);
    var einsen = w.filter(function (x) { return x === 1; }).length;
    var sechsen = w.filter(function (x) { return x === 6; }).length;
    var ziel = z.spieler[(z.dran + 1) % z.spieler.length];
    var gegeben = Math.min(einsen, p.hoelzer);
    p.hoelzer -= gegeben;
    if (ziel === p) {
      z.mitte += gegeben;                       // Alleinspieler: nichts zu verschenken
    } else if (ziel.raus) {
      z.mitte += gegeben;
    } else {
      ziel.hoelzer += gegeben;
    }
    var sechsenWeg = Math.min(sechsen, p.hoelzer);
    p.hoelzer -= sechsenWeg;
    z.mitte += sechsenWeg;
    TGW.note(z, p.name + ": " + w.join(" ") + " · " + einsen + " nach links, " + sechsen + " in die Mitte");
    var rausJetzt = false;
    if (p.hoelzer <= 0 && !p.raus) {
      p.raus = true;
      rausJetzt = true;
      TGW.note(z, p.name + " ist ausgeschieden");
    }
    var uebrig = z.spieler.filter(function (q) { return !q.raus; });
    if (uebrig.length <= 1) {
      if (uebrig.length === 1) TGW.sieger(z, [z.spieler.indexOf(uebrig[0])], "Letzter mit Streichhölzern.");
      else TGW.sieger(z, z.spieler.map(function (q, i) { return i; }), "Alle haben keine Hölzer mehr.");
      return;
    }
    this.naechsterAktiver(z);
    if (rausJetzt && !z.ende) {
      TGW.meldung(z, p.name + " hat keine Streichhölzer mehr und ist " +
        "ausgeschieden. Jetzt ist " + z.spieler[z.dran].name + " dran.");
    }
  }
});
