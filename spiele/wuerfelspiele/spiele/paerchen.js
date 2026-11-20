/* Pärchen – zwei positive und zwei negative Würfel, fünf Runden.
   Würfel, Runden und Wertung kommen aus TGW.plusMinusRunden. */
TGW.anmelden(TGW.plusMinusRunden({
  id: "paerchen",
  titel: "Pärchen",
  unter: "Vier Würfel · zwei zählen plus, zwei zählen minus · fünf Runden",
  minus: 2,
  klein: "Zwei Pluswürfel, zwei Minuswürfel (rot).",
  regeln: [
    "Jeder Spieler würfelt in jeder Runde mit vier Würfeln: zwei helle Pluswürfel und zwei rote Minuswürfel.",
    "Die beiden Pluswürfel werden addiert, die beiden Minuswürfel abgezogen.",
    "Das Ergebnis der Runde wird zur Gesamtsumme addiert – Minusrunden sind möglich.",
    "Nach fünf Runden gewinnt die höchste Gesamtsumme."
  ]
}));
