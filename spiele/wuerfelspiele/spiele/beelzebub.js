/* Beelzebub – drei positive Würfel gegen einen negativen, fünf Runden.
   Derselbe Ablauf wie Pärchen, nur mit einem Minuswürfel. */
TGW.anmelden(TGW.plusMinusRunden({
  id: "beelzebub",
  titel: "Beelzebub",
  unter: "Vier Würfel · drei gegen einen · fünf Runden",
  minus: 1,
  klein: "Drei Pluswürfel, ein Minuswürfel (rot).",
  regeln: [
    "Jeder Spieler würfelt in jeder Runde mit vier Würfeln: drei helle Pluswürfel und ein roter Minuswürfel.",
    "Die drei Pluswürfel werden addiert, der rote Würfel wird abgezogen.",
    "Das Ergebnis der Runde wird zur Gesamtsumme addiert – Minusrunden sind möglich.",
    "Nach fünf Runden gewinnt die höchste Gesamtsumme."
  ]
}));
