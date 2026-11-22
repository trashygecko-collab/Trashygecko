package main

import "time"

/* ===========================================================================
   BOT FÜR KALAHA

   Stufen:
     leicht  – nimmt zu 35 % eine beliebige Mulde, sonst einen Zug voraus.
     mittel  – fünf Züge voraus, mit Alpha-Beta.
     schwer  – sieben Züge voraus; zusätzlich zählen eigene Steine, die der
               Gegner als Nächstes fangen könnte, negativ.
   =========================================================================== */

func kalahaBotFarbe(g *KalahaGame) string {
	if g == nil || g.BotID == "" {
		return ""
	}
	return kalahaColorOf(g, g.BotID)
}

func kalahaBotTreiben(g *KalahaGame, jetzt time.Time) bool {
	farbe := kalahaBotFarbe(g)
	if farbe == "" || g.Status != "active" || g.Current != farbe {
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
	index, ok := kalahaBotWahl(g, farbe, g.BotStufe)
	if !ok {
		return false
	}
	return kalahaBotSaeen(g, farbe, index)
}

func kalahaBotMulden(pits []int, farbe string) []int {
	mulden := []int{}
	from, to := 0, 6
	if farbe == "blue" {
		from, to = 7, 13
	}
	for i := from; i < to; i++ {
		if pits[i] > 0 {
			mulden = append(mulden, i)
		}
	}
	return mulden
}

func kalahaBotWahl(g *KalahaGame, farbe, stufe string) (int, bool) {
	mulden := kalahaBotMulden(g.Pits, farbe)
	if len(mulden) == 0 {
		return 0, false
	}
	if botStufe(stufe) == botLeicht && botZufall().Intn(100) < 35 {
		return mulden[botZufall().Intn(len(mulden))], true
	}
	tiefe := 5
	switch botStufe(stufe) {
	case botLeicht:
		tiefe = 1
	case botSchwer:
		tiefe = 7
	}
	beste := mulden[0]
	besterWert := -1 << 30
	for _, index := range mulden {
		probe := cloneKalaha(g)
		kalahaPlay(probe, index)
		wert := kalahaBotSuche(probe, farbe, tiefe-1, -1<<30, 1<<30)
		if wert > besterWert || (wert == besterWert && botZufall().Intn(2) == 0) {
			beste, besterWert = index, wert
		}
	}
	return beste, true
}

func kalahaBotSaeen(g *KalahaGame, farbe string, index int) bool {
	if g == nil || g.Status != "active" || g.Current != farbe {
		return false
	}
	if !kalahaOwnPit(index, farbe) || g.Pits[index] == 0 {
		return false
	}
	g.Note = kalahaPlay(g, index)
	return true
}

func kalahaBotSuche(g *KalahaGame, farbe string, tiefe, alpha, beta int) int {
	if g.Status == "finished" || tiefe <= 0 {
		return kalahaBotWert(g, farbe)
	}
	mulden := kalahaBotMulden(g.Pits, g.Current)
	if len(mulden) == 0 {
		return kalahaBotWert(g, farbe)
	}
	if g.Current == farbe {
		best := -1 << 30
		for _, index := range mulden {
			probe := cloneKalaha(g)
			kalahaPlay(probe, index)
			wert := kalahaBotSuche(probe, farbe, tiefe-1, alpha, beta)
			if wert > best {
				best = wert
			}
			if best > alpha {
				alpha = best
			}
			if alpha >= beta {
				break
			}
		}
		return best
	}
	best := 1 << 30
	for _, index := range mulden {
		probe := cloneKalaha(g)
		kalahaPlay(probe, index)
		wert := kalahaBotSuche(probe, farbe, tiefe-1, alpha, beta)
		if wert < best {
			best = wert
		}
		if best < beta {
			beta = best
		}
		if alpha >= beta {
			break
		}
	}
	return best
}

func kalahaBotWert(g *KalahaGame, farbe string) int {
	eigen := kalahaOwnStore(farbe)
	fremd := kalahaOpponentStore(farbe)
	if g.Status == "finished" {
		if g.Pits[eigen] > g.Pits[fremd] {
			return 1000000 + (g.Pits[eigen]-g.Pits[fremd])*100
		}
		if g.Pits[eigen] < g.Pits[fremd] {
			return -1000000 - (g.Pits[fremd]-g.Pits[eigen])*100
		}
		return 0
	}
	wert := (g.Pits[eigen] - g.Pits[fremd]) * 20
	from, to := 0, 6
	if farbe == "blue" {
		from, to = 7, 13
	}
	for i := from; i < to; i++ {
		wert += g.Pits[i] * 2
		if g.Pits[i] == 1 {
			// Ein einzelner Stein gegenüber einer vollen Mulde wird gefangen.
			if g.Pits[12-i] > 0 {
				wert -= 6
			}
		}
	}
	for i := 0; i < 6; i++ {
		gegnerFach := i + 7
		if farbe == "blue" {
			gegnerFach = i
		}
		wert -= g.Pits[gegnerFach] * 2
	}
	return wert
}
