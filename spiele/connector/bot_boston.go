package main

/* ===========================================================================
   BOT FÜR „AUF NACH BOSTON“

   Boston ist ein reines Würfelrennen: Drei Würfel, drei Würfe, nach jedem Wurf
   bleibt der höchste liegen, wer am Ende die höchste Summe hat, holt den
   Siegpunkt. Es gibt keine Entscheidung – also würfelt der Bot einfach, wenn
   er dran ist, und gibt die nächste Runde frei.

   Die drei Stufen unterscheiden sich deshalb nur in der Denkpause
   (botWartezeit) und im Namen; ein besserer Zug ist bei Boston nicht möglich.
   Das steht so auch im Test.
   =========================================================================== */

import "time"

// bostonBotSitz liefert den Platz des Bots (-1, wenn keiner).
func bostonBotSitz(g *BostonGame) int {
	if g == nil || len(g.Bots) == 0 {
		return -1
	}
	return bostonFind(g, g.Bots[0])
}

// bostonBotAuftrag sagt, was der Bot jetzt tun soll.
func bostonBotAuftrag(g *BostonGame) (string, bool) {
	sitz := bostonBotSitz(g)
	if sitz < 0 {
		return "", false
	}
	switch g.Status {
	case "active":
		if g.Current == sitz && !g.Players[sitz].Done {
			return "roll", true
		}
	case "runde":
		// Der Bot blättert die Ergebnisübersicht weiter.
		return "weiter", true
	}
	return "", false
}

// bostonBotTreiben lässt den Bot nach der Denkpause würfeln.
func bostonBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Boston
	if _, ok := bostonBotAuftrag(g); !ok {
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
	return bostonBotSchritt(room)
}

// bostonBotSchritt führt genau eine Bot-Aktion aus (dieselbe Prüfung wie für
// einen Menschen).
func bostonBotSchritt(room *Room) bool {
	g := room.Boston
	aktion, ok := bostonBotAuftrag(g)
	if !ok {
		return false
	}
	sitz := bostonBotSitz(g)
	if _, meldung := bostonAktion(room, g.Players[sitz].ID, aktion); meldung != "" {
		return false
	}
	return true
}
