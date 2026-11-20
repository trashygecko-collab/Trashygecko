/* Eins bis sechzehn – Zahlen 1…16 vorwärts und wieder zurück bilden. */
TGW.anmelden({
  id: "eins-bis-sechzehn",
  titel: "Eins bis Sechzehn",
  unter: "Drei Würfel · Zahlen 1–16 vorwärts und rückwärts",
  spieler: { min: 1, max: 6, standard: 2 },
  regeln: [
    "Jeder Spieler hat ein Feld mit den Zahlen 1 bis 16.",
    "Wer am Zug ist, würfelt mit drei Würfeln.",
    "Die gesuchte Zahl ist die kleinste noch offene Zahl – zuerst 1, dann 2 und so weiter bis 16.",
    "Die Zahl muss aus den Würfeln gebildet werden: ein einzelner Würfel oder eine Addition bzw. Subtraktion mehrerer Würfel.",
    "Gelingt das, wird die Zahl gestrichen und es geht sofort mit dem nächsten Wurf weiter.",
    "Gelingt es nicht, ist der Nächste dran.",
    "Sind alle sechzehn Zahlen gestrichen, beginnt das Feld von vorn – jetzt von 16 rückwärts bis 1.",
    "Wer zuerst vorwärts und rückwärts alle Zahlen gebildet hat, gewinnt."
  ],
  start: function (z) {
    z.spieler.forEach(function (p) {
      p.feld = [];
      for (var i = 0; i < 16; i++) p.feld.push(false);
      p.rueck = false;
    });
    z.phase = "wurf";
    z.formeln = [];
  },
  ziel: function (p) {
    var i;
    if (!p.rueck) {
      for (i = 0; i < 16; i++) if (!p.feld[i]) return i + 1;
      return 0;
    }
    for (i = 15; i >= 0; i--) if (!p.feld[i]) return i + 1;
    return 0;
  },
  formeln: function (werte, ziel) {
    var liste = [];
    var gesehen = {};
    var n = werte.length;
    for (var maske = 1; maske < (1 << n); maske++) {
      var idx = [];
      for (var i = 0; i < n; i++) if (maske & (1 << i)) idx.push(i);
      for (var vorz = 0; vorz < (1 << idx.length); vorz++) {
        var summe = 0;
        var plus = [];
        var minus = [];
        for (var j = 0; j < idx.length; j++) {
          var istPlus = !!(vorz & (1 << j));
          var wert = werte[idx[j]];
          if (istPlus) { summe += wert; plus.push(wert); }
          else { summe -= wert; minus.push(wert); }
        }
        if (summe !== ziel) continue;
        plus.sort(function (a, b) { return b - a; });
        minus.sort(function (a, b) { return b - a; });
        var text = plus.length ? plus.join(" + ") : "0";
        minus.forEach(function (w) { text += " − " + w; });
        if (gesehen[text]) continue;
        gesehen[text] = true;
        liste.push({ text: text + " = " + ziel, art: "bilden", daten: { zahl: ziel } });
      }
    }
    return liste.slice(0, 10);
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var gesucht = this.ziel(p);
    return {
      wuerfel: z.wuerfel,
      hinweis: gesucht ? "Gesucht: " + gesucht + (p.rueck ? " rückwärts" : " vorwärts") : "Feld voll",
      klein: "Offen: " + p.feld.filter(function (x) { return !x; }).length + " von 16 · " + (p.rueck ? "Rückweg" : "Hinweg"),
      knopf: z.phase === "wurf" ? { text: "Würfeln", art: "wurf" } : null,
      auswahl: z.phase === "waehlen" ? z.formeln : [],
      leisten: [{
        titel: p.name + (p.rueck ? " – rückwärts" : " – vorwärts"),
        zahlen: p.feld.map(function (aus, i) { return { wert: i + 1, aus: aus }; })
      }],
      tabelle: {
        titel: "Stand",
        spalten: ["Spieler", "Hinweg", "Rückweg"],
        zeilen: z.spieler.map(function (q) {
          var offen = q.feld.filter(function (x) { return !x; }).length;
          return [q.name, q.rueck ? "16/16" : (16 - offen) + "/16", q.rueck ? (16 - offen) + "/16" : "—"];
        })
      }
    };
  },
  aktion: function (z, art, daten) {
    var p = z.spieler[z.dran];
    if (art === "wurf") {
      if (z.phase !== "wurf") return false;
      var gesucht = this.ziel(p);
      var w = z.wurf(3);
      z.formeln = this.formeln(w, gesucht);
      if (!z.formeln.length) {
        TGW.note(z, p.name + " kann " + gesucht + " nicht bilden: " + w.join(" "));
        TGW.naechster(z);
        return;
      }
      z.phase = "waehlen";
      return;
    }
    if (art === "bilden") {
      if (z.phase !== "waehlen") return false;
      var zahl = daten.zahl;
      p.feld[zahl - 1] = true;
      TGW.note(z, p.name + " bildet " + zahl + (p.rueck ? " rückwärts" : ""));
      z.formeln = [];
      z.phase = "wurf";
      var offen = p.feld.filter(function (x) { return !x; }).length;
      if (offen === 0) {
        if (!p.rueck) {
          p.rueck = true;
          p.feld = p.feld.map(function () { return false; });
          TGW.note(z, p.name + " beginnt den Rückweg");
        } else {
          TGW.sieger(z, [z.dran], "Vorwärts und rückwärts alle Zahlen gebildet.");
        }
      }
      return;
    }
  }
});
