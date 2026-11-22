package main

import "time"

/* ===========================================================================
   BOT FÜR DAS GÄNSESPIEL

   Auch hier entscheidet nur der Würfel; der Bot würfelt wie ein Mensch. Die
   Stufe ändert nichts am Ergebnis und ist deshalb nur Anzeige.
   =========================================================================== */

func gaenseBotSitz(g *GaenseGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return g.sitz(g.BotID)
}

func gaenseBotTreiben(g *GaenseGame, jetzt time.Time) bool {
	sitz := gaenseBotSitz(g)
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
	g.gaenseWuerfeln(sitz)
	return true
}
