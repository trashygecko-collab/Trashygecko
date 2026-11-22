package main

/* ===========================================================================
   BOT FÜR SCHIFFE VERSENKEN (SHIP CAPTAIN CREW)

   Fünf Würfel, höchstens drei Würfe. Gesucht werden nacheinander 6 (Schiff),
   5 (Kapitän) und 4 (Maat). Erst danach zählen die beiden übrigen Würfel als
   Ladung. Der Bot sichert also den Würfel, der gerade dran ist – fehlt er,
   würfelt er weiter, solange Würfe übrig sind.

   Stufen (nur die Ladung unterscheidet sich, das Suchen ist reines Würfeln):
     leicht  – gibt auf, wenn der gesuchte Würfel im ersten Wurf fehlt, und
               nimmt die Ladung, wie sie liegt.
     mittel  – würfelt die Ladung neu, solange sie 7 oder weniger ergibt.
     schwer  – würfelt die Ladung neu, solange sie 9 oder weniger ergibt.
   =========================================================================== */

import "time"

// shipcrewBotSitz liefert den Platz des Bots (-1, wenn keiner).
func shipcrewBotSitz(g *ShipCrewGame) int {
	if g == nil || len(g.Bots) == 0 {
		return -1
	}
	return scFind(g, g.Bots[0])
}

// shipcrewBotLadungsgrenze: bis zu welcher Ladung der Bot noch einmal würfelt.
//
// Zwei Würfel bringen im Schnitt 7. Wer weniger hat, verbessert sich mit einem
// neuen Wurf. Schwer schaut zusätzlich auf die schon fertigen Mitspieler: eine
// Ladung, die nicht für die Führung reicht, ist wertlos – also nochmal würfeln.
func shipcrewBotLadungsgrenze(g *ShipCrewGame, stufe string) int {
	switch botStufe(stufe) {
	case botLeicht:
		return 0
	case botMittel:
		return 6
	}
	best := 0
	for _, p := range g.Players {
		if p.Fertig && p.Wert > best {
			best = p.Wert
		}
	}
	if best == 0 {
		best = 6
	}
	return best
}

// shipcrewBotAuftrag sagt, was der Bot jetzt tun soll.
func shipcrewBotAuftrag(g *ShipCrewGame) (string, int, bool) {
	sitz := shipcrewBotSitz(g)
	if sitz < 0 {
		return "", 0, false
	}
	switch g.Status {
	case "runde":
		return "weiter", 0, true
	case "active":
	default:
		return "", 0, false
	}
	if g.Current != sitz {
		return "", 0, false
	}
	p := &g.Players[sitz]
	if p.Fertig {
		return "", 0, false
	}
	// Fehlt noch ein Spezialwürfel, sucht der Bot genau den.
	if erwartet := scErwartet(p); erwartet != 0 {
		if len(p.Wuerfel) == 0 {
			if p.Wuerfe < scWuerfe {
				return "wurf", 0, true
			}
			return "bestaetigen", 0, true
		}
		for i, a := range p.Wuerfel {
			if a == erwartet {
				return "sichern", i, true
			}
		}
		// Leicht gibt nach dem ersten Wurf auf, wenn der gesuchte Würfel
		// fehlt; mittel und schwer würfeln ihre drei Würfe zu Ende.
		if botStufe(g.BotStufe) == botLeicht && p.Wuerfe >= 1 {
			return "bestaetigen", 0, true
		}
		if p.Wuerfe < scWuerfe {
			return "wurf", 0, true
		}
		return "bestaetigen", 0, true
	}

	// Schiff, Kapitän und Maat liegen: jetzt zählt nur noch die Ladung.
	if len(p.Wuerfel) == 0 {
		if p.Wuerfe < scWuerfe {
			return "wurf", 0, true
		}
		return "bestaetigen", 0, true
	}
	if p.Wuerfe < scWuerfe && scLadung(p) <= shipcrewBotLadungsgrenze(g, g.BotStufe) {
		return "wurf", 0, true
	}
	return "bestaetigen", 0, true
}

// shipcrewBotTreiben lässt den Bot nach der Denkpause handeln.
func shipcrewBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.ShipCrew
	if _, _, ok := shipcrewBotAuftrag(g); !ok {
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
	return shipcrewBotSchritt(room)
}

// shipcrewBotSchritt führt genau eine Bot-Aktion aus.
func shipcrewBotSchritt(room *Room) bool {
	g := room.ShipCrew
	aktion, index, ok := shipcrewBotAuftrag(g)
	if !ok {
		return false
	}
	sitz := shipcrewBotSitz(g)
	if _, meldung := shipcrewAktion(room, g.Players[sitz].ID, aktion, index); meldung != "" {
		return false
	}
	return true
}
