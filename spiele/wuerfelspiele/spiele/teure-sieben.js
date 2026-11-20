/* Teure Sieben – 2–6 und 8–12 abstreichen, die Sieben streicht doppelt. */
TGW.anmelden({
  id: "teure-sieben",
  titel: "Teure Sieben",
  unter: "Zwei Würfel · Zahlen 2–6 und 8–12 · Sieben streicht doppelt",
  spieler: { min: 2, max: 6, standard: 2 },
  regeln: [
    "Jeder Spieler hat ein Blatt mit den Zahlen 2 bis 6 und 8 bis 12 – die Sieben fehlt.",
    "Wer am Zug ist, würfelt zwei Würfel; die Summe bestimmt die Zahl.",
    "Eine offene Zahl wird gestrichen.",
    "Ist die Zahl schon gestrichen, wandert sie als Bonus zum nächsten Spieler: Hat der sie noch offen, wird sie dort gestrichen.",
    "Eine gewürfelte Sieben erlaubt einen Mehrfachstrich: bis zu zwei beliebige offene Zahlen werden gestrichen.",
    "Wer zuerst alle zehn Zahlen gestrichen hat, gewinnt – auch durch einen Bonus."
  ],
  ZAHLEN: [2, 3, 4, 5, 6, 8, 9, 10, 11, 12],
  start: function (z) {
    z.spieler.forEach(function (p) {
      p.aus = {};
      this.ZAHLEN.forEach(function (n) { p.aus[n] = false; }, this);
      p.wahl = 0;
    }, this);
    z.phase = "wurf";
  },
  frei: function (p) {
    return this.ZAHLEN.filter(function (n) { return !p.aus[n]; }, this);
  },
  voll: function (p) {
    return this.frei(p).length === 0;
  },
  sicht: function (z) {
    var p = z.spieler[z.dran];
    var frei = this.frei(p);
    return {
      wuerfel: z.wuerfel,
      summe: z.wuerfel.length === 2 ? z.wuerfel[0] + z.wuerfel[1] : null,
      summeText: "Summe",
      hinweis: z.phase === "waehlen"
        ? "Sieben! Bis zu zwei Zahlen streichen (" + Math.max(0, 2 - p.wahl) + " offen)"
        : "Würfeln",
      klein: "Schon gestrichene Zahlen gehen als Bonus an den Nächsten.",
      knopf: z.phase === "wurf" ? { text: "Würfeln", art: "wurf" } : null,
      auswahl: z.phase === "waehlen" ? frei.map(function (n) {
        return { text: "Streichen: " + n, art: "waehlen", daten: { zahl: n } };
      }) : [],
      leisten: z.spieler.map(function (q) {
        return {
          titel: q.name + " – " + (this.ZAHLEN.length - this.frei(q).length) + "/10",
          zahlen: this.ZAHLEN.map(function (n) { return { wert: n, aus: q.aus[n] }; }, this)
        };
      }, this),
      spieler: z.spieler.map(function (q) {
        return { name: q.name, wert: this.ZAHLEN.length - this.frei(q).length + "/10", neben: "" };
      }, this)
    };
  },
  aktion: function (z, art, daten) {
    var p = z.spieler[z.dran];
    if (art === "wurf") {
      if (z.phase !== "wurf") return false;
      var w = z.wurf(2);
      var summe = w[0] + w[1];
      if (summe === 7) {
        p.wahl = 0;
        z.phase = "waehlen";
        TGW.note(z, p.name + " würfelt die Sieben – Mehrfachstrich");
        if (!this.frei(p).length) this.zugBeenden(z);
        return;
      }
      if (!p.aus[summe]) {
        p.aus[summe] = true;
        TGW.note(z, p.name + " streicht " + summe);
        if (this.voll(p)) { TGW.sieger(z, [z.dran], "Alle zehn Zahlen gestrichen."); return; }
      } else {
        this.bonus(z, summe);
      }
      this.zugBeenden(z);
      return;
    }
    if (art === "waehlen") {
      if (z.phase !== "waehlen" || p.aus[daten.zahl]) return false;
      p.aus[daten.zahl] = true;
      p.wahl += 1;
      TGW.note(z, p.name + " streicht " + daten.zahl + " (Sieben)");
      if (this.voll(p)) { TGW.sieger(z, [z.dran], "Alle zehn Zahlen gestrichen."); return; }
      if (p.wahl >= 2 || !this.frei(p).length) this.zugBeenden(z);
    }
  },
  bonus: function (z, zahl) {
    if (z.spieler.length < 2) return;
    var naechster = z.spieler[(z.dran + 1) % z.spieler.length];
    if (naechster.aus[zahl]) {
      TGW.note(z, "Bonus " + zahl + ": " + naechster.name + " hat sie schon");
      return;
    }
    naechster.aus[zahl] = true;
    TGW.note(z, "Bonus " + zahl + " an " + naechster.name);
    if (this.voll(naechster)) {
      TGW.sieger(z, [(z.dran + 1) % z.spieler.length], "Durch den Bonus alle zehn Zahlen gestrichen.");
    }
  },
  zugBeenden: function (z) {
    z.phase = "wurf";
    if (!z.ende) TGW.naechster(z);
  }
});
