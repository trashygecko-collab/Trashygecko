/* Beelzebub – drei positive Würfel gegen einen negativen, fünf Runden.
   Derselbe Ablauf wie Pärchen, nur mit einem Minuswürfel. */
TGW.anmelden(TGW.plusMinusRunden({
  id: "beelzebub",
  titel: "Beelzebub",
  unter: "Vier Würfel · drei gegen einen · fünf Runden",
  minus: 1,
  klein: "Drei Pluswürfel, ein Minuswürfel (rot).",
  kurz: "Drei Würfel zählen plus, ein roter zählt minus. Fünf Runden, höchste Summe gewinnt.",
  erklaerung: {
    ziel: "Nach fünf Runden die höchste Gesamtsumme haben.",
    zug: "Einmal mit vier Würfeln würfeln. Drei helle Würfel zählen plus, der rote zählt minus.",
    wuerfel: "Beispiel: 5 + 3 + 2 − 6 = 4. Das Rundenergebnis kommt zur Gesamtsumme.",
    nochmal: "Nein – ein Wurf pro Runde, dann ist der Nächste dran.",
    sieg: "Höchste Gesamtsumme nach fünf Runden.",
    besonders: [
      "Eine Eins auf dem roten Würfel ist Glück: Dann wird nur 1 abgezogen.",
      "Runden können negativ sein, die Gesamtsumme kann sinken."
    ]
  },
  regeln: [
    "Jeder Spieler würfelt in jeder Runde mit vier Würfeln: drei helle Pluswürfel und ein roter Minuswürfel.",
    "Die drei Pluswürfel werden addiert, der rote Würfel wird abgezogen.",
    "Das Ergebnis der Runde wird zur Gesamtsumme addiert – Minusrunden sind möglich.",
    "Nach fünf Runden gewinnt die höchste Gesamtsumme."
  ]
}));
