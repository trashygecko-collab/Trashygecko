package main

import "time"

/* ===========================================================================
   BOT FÜR SCHLANGEN UND LEITERN

   Mehr als würfeln gibt das Spiel nicht her: Es gibt keine Entscheidung, die
   ein Mensch treffen könnte, und damit auch keine Spielstärke. Der Bot würfelt
   deshalb in jeder Stufe gleich – die Stufe steht nur in der Anzeige.
   =========================================================================== */

func schlangenBotSitz(g *SchlangenGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return g.sitz(g.BotID)
}

func schlangenBotTreiben(g *SchlangenGame, jetzt time.Time) bool {
	sitz := schlangenBotSitz(g)
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
	g.schlangenWuerfeln(sitz)
	return true
}
