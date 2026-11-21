/* Pärchen – zwei positive und zwei negative Würfel, fünf Runden.
   Würfel, Runden und Wertung kommen aus TGW.plusMinusRunden. */
TGW.anmelden(TGW.plusMinusRunden({
  id: "paerchen",
  titel: "Pärchen",
  unter: "Vier Würfel · zwei zählen plus, zwei zählen minus · fünf Runden",
  minus: 2,
  klein: "Zwei Pluswürfel, zwei Minuswürfel (rot).",
  kurz: "Vier Würfel: zwei zählen plus, zwei rot und minus. Fünf Runden, höchste Summe gewinnt.",
  erklaerung: {
    ziel: "Nach fünf Runden die höchste Gesamtsumme haben.",
    zug: "Einmal mit vier Würfeln würfeln. Zwei helle Würfel zählen plus, die zwei roten zählen minus.",
    wuerfel: "Beispiel: 5 + 3 − 2 − 6 = 0. Das Rundenergebnis kommt zur Gesamtsumme.",
    nochmal: "Nein – ein Wurf pro Runde, dann ist der Nächste dran.",
    sieg: "Höchste Gesamtsumme nach fünf Runden.",
    besonders: [
      "Runden können negativ sein – hohe rote Würfel ziehen Punkte ab.",
      "Die Gesamtsumme kann im Lauf der Partie sinken."
    ]
  },
  regeln: [
    "Jeder Spieler würfelt in jeder Runde mit vier Würfeln: zwei helle Pluswürfel und zwei rote Minuswürfel.",
    "Die beiden Pluswürfel werden addiert, die beiden Minuswürfel abgezogen.",
    "Das Ergebnis der Runde wird zur Gesamtsumme addiert – Minusrunden sind möglich.",
    "Nach fünf Runden gewinnt die höchste Gesamtsumme."
  ]
}));
