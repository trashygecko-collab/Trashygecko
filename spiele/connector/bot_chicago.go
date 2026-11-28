package main

/* ===========================================================================
   BOT FÜR CHICAGO

   Chicago ist Wert-Würfeln: Eins zählt 100, Sechs 60, alles andere seinen
   Wert. Drei Würfel, bis zu drei Würfe – man kann nach jedem Wurf Würfel
   liegen lassen und mit den übrigen weiterwürfeln. Der Bot entscheidet also,
   welche Würfel er behält; alles andere (Wurfzahl, Filze, Chic) macht die
   Regel im Spiel selbst.

   Stufen:
     leicht  – behält alles, was liegt (auch Zweien und Dreien).
     mittel  – behält nur 1, 5 und 6.
     schwer  – behält nur 1 und 6 (die einzig wirklich wertvollen Würfel) und
               dreht drei Sechser in zwei Einser.
   =========================================================================== */

import "time"

// chicagoBotSitz liefert den Platz des Bots (-1, wenn keiner).
func chicagoBotSitz(g *ChicagoGame) int {
	if g == nil || len(g.Bots) == 0 {
		return -1
	}
	return chFind(g, g.Bots[0])
}

// chicagoBotBehalten sagt, ob ein Würfel für die Stufe wertvoll genug ist.
func chicagoBotBehalten(stufe string, auge int) bool {
	switch botStufe(stufe) {
	case botLeicht:
		return true
	case botMittel:
		return auge == 1 || auge == 5 || auge == 6
	}
	return auge == 1 || auge == 6
}

// chicagoBotAuftrag sagt, was der Bot jetzt tun soll.
func chicagoBotAuftrag(g *ChicagoGame) (string, []int, bool) {
	sitz := chicagoBotSitz(g)
	if sitz < 0 {
		return "", nil, false
	}
	switch g.Status {
	case "runde":
		return "weiter", nil, true
	case "active":
	default:
		return "", nil, false
	}
	if g.Current != sitz {
		return "", nil, false
	}
	p := &g.Players[sitz]
	if p.Fertig || p.Raus {
		return "", nil, false
	}
	// Liegen noch Würfel eines Chic herum, sind sie nicht die eigenen.
	if len(g.Wuerfel) == 0 || g.Wer != p.ID {
		if p.Wuerfe < chMaxWuerfe(g) {
			return "wurf", nil, true
		}
		// Würfe sind aufgebraucht: abgeben. Früher gab der Bot hier auf, und
		// der Tisch blieb stehen, weil niemand mehr am Zug war.
		return "weitergeben", nil, true
	}

	// Drei Sechser sind 180 – gedreht werden daraus zwei Einser (200).
	sechsen := 0
	for _, a := range g.Wuerfel {
		if a == 6 {
			sechsen++
		}
	}
	if sechsen == 3 {
		return "drehen", nil, true
	}

	// Welche Würfel lohnen sich? Nur so viele, wie noch Platz ist.
	frei := 3 - len(p.Halten)
	behalten := []int{}
	for i, a := range g.Wuerfel {
		if len(behalten) >= frei {
			break
		}
		if chicagoBotBehalten(g.BotStufe, a) {
			behalten = append(behalten, i)
		}
	}
	if len(behalten) > 0 {
		return "halten", behalten, true
	}
	// Nichts wertvoll – mit den restlichen Würfen neu würfeln.
	if p.Wuerfe < chMaxWuerfe(g) {
		return "wurf", nil, true
	}
	return "", nil, false
}

// chicagoBotTreiben lässt den Bot nach der Denkpause handeln.
func chicagoBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Chicago
	if _, _, ok := chicagoBotAuftrag(g); !ok {
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
	return chicagoBotSchritt(room)
}

// chicagoBotSchritt führt genau eine Bot-Aktion aus.
func chicagoBotSchritt(room *Room) bool {
	g := room.Chicago
	aktion, indizes, ok := chicagoBotAuftrag(g)
	if !ok {
		return false
	}
	sitz := chicagoBotSitz(g)
	if _, meldung := chicagoAktion(room, g.Players[sitz].ID, aktion, indizes); meldung != "" {
		return false
	}
	return true
}
