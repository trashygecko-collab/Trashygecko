/* Würfelpoker – fünf Würfel, zwei Nachwürfe, drei Blätter je Spieler. */
TGW.anmelden({
  id: "wuerfelpoker",
  titel: "Würfelpoker",
  unter: "Fünf Würfel · zwei Nachwürfe · drei Blätter je Spieler",
  spieler: { min: 1, max: 6, standard: 2 },
  regeln: [
    "Jeder Spieler spielt drei Blätter.",
    "Ein Blatt besteht aus fünf Würfeln; es darf zweimal nachgewürfelt werden.",
    "Zum Nachwürfeln werden Würfel angetippt und damit behalten – nur die übrigen werden neu geworfen.",
    "Gewertet wird von oben nach unten: Fünfling, Vierling, Full House, Straße, Drilling, zwei Paare, Paar.",
    "Bei gleicher Kombination entscheidet die höhere Kombinationszahl, danach die höchste verbleibende Augenzahl.",
    "Die drei Blätter werden addiert; wer die höchste Summe hat, gewinnt.",
    "Es wird ausschließlich um Spielpunkte gespielt."
  ],
  HAEND: 3,
  NACHWUERFE: 2,
  start: function (z) {
    z.spieler.forEach(function (p) {
      p.punkte = 0;
      p.blatt = 0;
      p.fertig = false;
      p.haende = [];
      this.neuesBlatt(p);
    }, this);
  },
  neuesBlatt: function (p) {
    p.hand = [];
    p.halten = [false, false, false, false, false];
    p.nachwuerfe = 0;
    p.ergebnis = null;
  },
  werten: function (w) {
    var zaehl = {};
    w.forEach(function (x) { zaehl[x] = (zaehl[x] || 0) + 1; });
    var paare = [];
    var drilling = 0;
    var vierling = 0;
    var fuenfling = 0;
    Object.keys(zaehl).map(Number).sort(function (a, b) { return b - a; }).forEach(function (augen) {
      var n = zaehl[augen];
      if (n === 5) fuenfling = augen;
      else if (n === 4) vierling = augen;
      else if (n === 3) drilling = augen;
      else if (n === 2) paare.push(augen);
    });
    var klein = [1, 2, 3, 4, 5].every(function (x) { return w.indexOf(x) >= 0; });
    var gross = [2, 3, 4, 5, 6].every(function (x) { return w.indexOf(x) >= 0; });
    var ergebnis;
    if (fuenfling) ergebnis = { rang: 7, name: "Fünfling", kombi: fuenfling };
    else if (vierling) ergebnis = { rang: 6, name: "Vierling", kombi: vierling };
    else if (drilling && paare.length >= 1) ergebnis = { rang: 5, name: "Full House", kombi: drilling };
    else if (klein || gross) ergebnis = { rang: 4, name: "Straße", kombi: klein ? 5 : 6 };
    else if (drilling) ergebnis = { rang: 3, name: "Drilling", kombi: drilling };
    else if (paare.length >= 2) ergebnis = { rang: 2, name: "Zwei Paare", kombi: Math.max(paare[0], paare[1]) };
    else if (paare.length === 1) ergebnis = { rang: 1, name: "Paar", kombi: paare[0] };
    else ergebnis = { rang: 0, name: "Nichts", kombi: Math.max.apply(null, w) };
    var rest = 0;
    w.forEach(function (x) { if (x !== ergebnis.kombi && x > rest) rest = x; });
    ergebnis.rest = rest;
    ergebnis.punkte = ergebnis.rang * 100000 + ergebnis.kombi * 1000 + rest * 10;
    return ergebnis;
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var w = p.hand;
    return {
      wuerfel: w.map(function (wert, i) {
        return { wert: wert, waehlbar: p.hand.length > 0, gewaehlt: p.halten[i] };
      }),
      summe: p.ergebnis ? p.ergebnis.name : null,
      summeText: p.ergebnis ? "Blatt" : "",
      hinweis: p.hand.length === 0
        ? "Blatt " + (p.blatt + 1) + " von " + this.HAEND + " · erster Wurf"
        : (p.ergebnis ? p.ergebnis.name : "Würfel antippen, um sie zu behalten"),
      klein: p.hand.length === 0
        ? "Fünf Würfel, danach zwei Nachwürfe."
        : "Nachwürfe übrig: " + (this.NACHWUERFE - p.nachwuerfe),
      knopf: p.hand.length === 0
        ? { text: "Würfeln", art: "wurf" }
        : { text: "Nachwürfeln (" + (this.NACHWUERFE - p.nachwuerfe) + ")", art: "nochmal", an: p.nachwuerfe < this.NACHWUERFE },
      knopf2: p.hand.length === 0 ? null : { text: "Blatt werten", art: "fertig" },
      spieler: z.spieler.map(function (q) {
        return {
          name: q.name,
          wert: q.punkte,
          neben: q.fertig ? "fertig" : ("Blatt " + (q.blatt + 1) + "/" + this.HAEND)
        };
      }, this),
      tabelle: {
        titel: "Blätter",
        spalten: ["Spieler", "1", "2", "3"],
        zeilen: z.spieler.map(function (q) {
          var zeile = [q.name];
          for (var i = 0; i < this.HAEND; i++) {
            zeile.push(q.haende[i] ? q.haende[i].name + " (" + q.haende[i].punkte + ")" : "—");
          }
          return zeile;
        }, this)
      }
    };
  },
  aktion: function (z, art, daten) {
    var p = z.spieler[z.dran];
    if (p.fertig) return false;
    if (art === "wurf") {
      if (p.hand.length > 0) return false;
      p.hand = z.wurf(5).slice();
      return;
    }
    if (art === "wuerfel") {
      if (p.hand.length === 0) return false;
      var i = daten.index;
      if (i < 0 || i > 4) return false;
      p.halten[i] = !p.halten[i];
      return;
    }
    if (art === "nochmal") {
      if (p.hand.length === 0 || p.nachwuerfe >= this.NACHWUERFE) return false;
      var neu = z.wurf(5);
      for (var k = 0; k < 5; k++) if (!p.halten[k]) p.hand[k] = neu[k];
      p.nachwuerfe += 1;
      TGW.note(z, p.name + " würfelt nach (" + p.nachwuerfe + "/" + this.NACHWUERFE + ")");
      return;
    }
    if (art === "fertig") {
      if (p.hand.length === 0) return false;
      p.ergebnis = this.werten(p.hand);
      p.punkte += p.ergebnis.punkte;
      p.haende.push(p.ergebnis);
      p.blatt += 1;
      TGW.note(z, p.name + ": " + p.ergebnis.name + " → " + p.ergebnis.punkte);
      z.wuerfel = [];
      if (p.blatt >= this.HAEND) {
        p.fertig = true;
        if (TGW.endeNachHoechstem(z, function (q) { return q.punkte; }, "Höchste Blattsumme")) return;
        TGW.naechster(z);
      } else {
        this.neuesBlatt(p);
      }
    }
  }
});
