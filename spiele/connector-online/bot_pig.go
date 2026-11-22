package main

import "time"

/* ===========================================================================
   BOT FÜR PIG

   Der Bot würfelt und sichert über dieselben Funktionen wie ein Mensch.

   Stufen:
     leicht  – sichert früh und zufällig (14 Punkte, sonst oft schon bei 8).
     mittel  – würfelt bis 20 Punkte, sichert aber sofort, wenn das Ziel
               erreicht ist.
     schwer  – rechnet den Erwartungswert: Weiterwürfeln bringt im Schnitt
               3,33 Punkte, kostet aber mit 1/6 die ganze Rundensumme.
   =========================================================================== */

func pigBotSitz(g *PigGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return g.sitz(g.BotID)
}

func pigBotTreiben(g *PigGame, jetzt time.Time) bool {
	sitz := pigBotSitz(g)
	if sitz < 0 || g.Status != "spiel" || g.Zug != sitz {
		return false
	}
	if g.BotFrist.IsZero() {
		g.BotFrist = jetzt.Add(botWartezeit(g.BotStufe))
		return false
	}
	if jetzt.Before(g.BotFrist) {
		return false
	}
	g.BotFrist = time.Time{}
	sp := &g.Spieler[sitz]
	if sp.Punkte+sp.Runde >= pigZiel || pigBotSichern(g, sp, g.BotStufe) {
		g.pigSichern(sitz)
		return true
	}
	g.pigWuerfeln(sitz)
	return true
}

func pigBotSichern(g *PigGame, sp *PigSpieler, stufe string) bool {
	if sp.Runde == 0 {
		return false
	}
	switch botStufe(stufe) {
	case botLeicht:
		if botZufall().Intn(100) < 25 && sp.Runde >= 8 {
			return true
		}
		return sp.Runde >= 14
	case botSchwer:
		// Erwartungswert: (5/6)*(+4) - (1/6)*Runde > 0  ⇔  Runde < 20.
		// Kurz vor dem Ziel zählt der sichere Abschluss mehr.
		if sp.Punkte+sp.Runde >= pigZiel {
			return true
		}
		return sp.Runde >= 20 || sp.Punkte+sp.Runde >= pigZiel-6 && sp.Runde >= 12
	}
	return sp.Runde >= 20
}
