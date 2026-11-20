/* Die böse Drei – beliebig oft würfeln, eine Drei vernichtet die Rundensumme. */
TGW.anmelden({
  id: "boese-drei",
  titel: "Die böse Drei",
  unter: "Ein Würfel · beliebig viele Würfe · eine Drei löscht alles",
  spieler: { min: 1, max: 6, standard: 2 },
  regeln: [
    "Wer am Zug ist, darf beliebig oft würfeln.",
    "Alle gewürfelten Augen werden zur Rundensumme addiert.",
    "Eine gewürfelte Drei vernichtet die bisherige Rundensumme und beendet den Zug sofort.",
    "Mit „Halten“ beendet man den Zug freiwillig und sichert die Summe.",
    "Der Spieler mit dem höchsten gesicherten Ergebnis gewinnt den Durchgang."
  ],
  start: function (z) {
    z.spieler.forEach(function (p) { p.zug = 0; p.best = 0; p.fertig = false; });
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    return {
      wuerfel: z.wuerfel,
      summe: p.zug,
      summeText: "im Zug",
      hinweis: p.fertig ? p.name + " hat " + p.best + " gesichert" : "Weiterwürfeln oder halten?",
      klein: "Eine Drei löscht die Zugsumme.",
      knopf: p.fertig ? null : { text: "Würfeln", art: "wurf" },
      knopf2: !p.fertig && p.zug > 0 ? { text: "Halten (" + p.zug + ")", art: "halten" } : null,
      spieler: z.spieler.map(function (q) {
        return {
          name: q.name,
          wert: q.best,
          neben: q.fertig ? "gesichert" : (q === p ? "im Zug " + q.zug : "wartet")
        };
      })
    };
  },
  aktion: function (z, art) {
    var p = z.spieler[z.dran];
    if (p.fertig) return false;
    if (art === "wurf") {
      var wert = z.wurf(1)[0];
      if (wert === 3) {
        TGW.note(z, p.name + " würfelt eine Drei – " + p.zug + " Punkte weg");
        p.zug = 0;
        this.zugBeenden(z, p);
      } else {
        p.zug += wert;
        TGW.note(z, p.name + " würfelt " + wert + " (" + p.zug + ")");
      }
      return;
    }
    if (art === "halten") {
      TGW.note(z, p.name + " hält bei " + p.zug);
      this.zugBeenden(z, p);
    }
  },
  zugBeenden: function (z, p) {
    p.best = Math.max(p.best, p.zug);
    p.fertig = true;
    if (TGW.endeNachHoechstem(z, function (q) { return q.best; }, "Höchstes Ergebnis")) return;
    TGW.naechster(z);
    z.spieler[z.dran].zug = 0;
  }
});
