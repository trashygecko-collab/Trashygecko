package main

import "time"

/* ===========================================================================
   BOT FÜR SCHIFFE VERSENKEN

   Der Bot setzt eine feste, gültige Flotte und schießt nur auf Felder, die er
   noch nicht beschossen hat. Er liest ausschließlich sein eigenes Schussbild
   (Wasser/Treffer/versenkt) – nie die gegnerische Flotte.

   Stufen:
     leicht  – schießt irgendwo auf ein freies Feld.
     mittel  – sucht im Schachbrettmuster und verfolgt Treffer.
     schwer  – verfolgt Treffer, verlängert Trefferlinien und schießt zuerst
               auf die Nachbarfelder bereits getroffener Schiffe.
   =========================================================================== */

func schiffeBotSitz(g *SchiffeGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return schiffeSitz(g, g.BotID)
}

// schiffeBotFlotte ist eine feste Aufstellung, die die Prüfung akzeptiert.
func schiffeBotFlotte() []SchiffeSchiff {
	return []SchiffeSchiff{
		{X: 0, Y: 0, Laenge: 5, Waagerecht: true},
		{X: 0, Y: 2, Laenge: 4, Waagerecht: true},
		{X: 0, Y: 4, Laenge: 3, Waagerecht: true},
		{X: 0, Y: 7, Laenge: 3, Waagerecht: true},
		{X: 0, Y: 9, Laenge: 2, Waagerecht: true},
	}
}

func schiffeBotTreiben(g *SchiffeGame, jetzt time.Time) bool {
	sitz := schiffeBotSitz(g)
	if sitz < 0 {
		return false
	}
	if g.Status == "setzen" && !g.Spieler[sitz].Fertig {
		if fehler := schiffeSetzen(g, sitz, schiffeBotFlotte()); fehler != "" {
			return false
		}
		g.Note = g.Spieler[sitz].Name + " hat die Flotte gesetzt."
		if schiffeAlleFertig(g) {
			g.Status = "spiel"
			g.Zug = 0
			g.Art = "info"
			g.Meldung = "Beide Flotten stehen – " + g.Spieler[0].Name + " beginnt."
		}
		return true
	}
	if g.Status != "spiel" || g.Zug != sitz {
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
	for i := 0; i < 12; i++ {
		if g.Status != "spiel" || g.Zug != sitz {
			break
		}
		x, y, ok := schiffeBotZiel(g, sitz, g.BotStufe)
		if !ok {
			break
		}
		g.Meldung = schiffeSchuss(g, sitz, x, y)
		g.Note = g.Spieler[sitz].Name + " schießt auf " + schiffeFeldName(x, y) + "."
		if schiffeVersenktAlle(g, 1-sitz) {
			g.Status = "ende"
			g.Sieger = g.Spieler[sitz].Name
			g.Meldung = "Alle Schiffe versenkt – " + g.Spieler[sitz].Name + " gewinnt die Partie!"
			g.Art = "versenkt"
			g.Zug = -1
			break
		}
		if g.Art == "wasser" {
			g.Zug = 1 - sitz
			break
		}
	}
	return true
}

// schiffeBotZiel wählt ein noch unbeschossenes Feld. Das Schussbild liegt beim
// Gegner-Spieler (er speichert, wohin geschossen wurde).
func schiffeBotZiel(g *SchiffeGame, sitz int, stufe string) (int, int, bool) {
	beschuss := g.Spieler[1-sitz].Beschuss
	frei := func(x, y int) bool {
		return x >= 0 && y >= 0 && x < schiffeGroesse && y < schiffeGroesse && beschuss[y][x] == 0
	}
	kandidaten := [][2]int{}
	if botStufe(stufe) != botLeicht {
		// Nachbarfelder offener Treffer zuerst.
		for y := 0; y < schiffeGroesse; y++ {
			for x := 0; x < schiffeGroesse; x++ {
				if beschuss[y][x] != 2 {
					continue
				}
				for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
					nx, ny := x+d[0], y+d[1]
					if frei(nx, ny) {
						kandidaten = append(kandidaten, [2]int{nx, ny})
					}
				}
			}
		}
	}
	if len(kandidaten) > 0 {
		// Schwer: liegt ein zweiter Treffer in der Reihe, an der Linie bleiben.
		if botStufe(stufe) == botSchwer {
			for _, k := range kandidaten {
				x, y := k[0], k[1]
				if (frei(x-1, y) && beschuss[y][x-1] == 2) || (frei(x+1, y) && beschuss[y][x+1] == 2) ||
					(frei(x, y-1) && beschuss[y-1][x] == 2) || (frei(x, y+1) && beschuss[y+1][x] == 2) {
					return x, y, true
				}
			}
			// Eine Linie aus zwei Treffern verlängern.
			for y := 0; y < schiffeGroesse; y++ {
				for x := 0; x < schiffeGroesse; x++ {
					if beschuss[y][x] != 2 {
						continue
					}
					if x+1 < schiffeGroesse && beschuss[y][x+1] == 2 {
						for nx := x - 1; nx >= 0 && beschuss[y][nx] == 2; nx-- {
							if frei(nx-1, y) {
								return nx - 1, y, true
							}
						}
						nx := x + 2
						for nx < schiffeGroesse && beschuss[y][nx] == 2 {
							nx++
						}
						if frei(nx, y) {
							return nx, y, true
						}
					}
					if y+1 < schiffeGroesse && beschuss[y+1][x] == 2 {
						for ny := y - 1; ny >= 0 && beschuss[ny][x] == 2; ny-- {
							if frei(x, ny-1) {
								return x, ny - 1, true
							}
						}
						ny := y + 2
						for ny < schiffeGroesse && beschuss[ny][x] == 2 {
							ny++
						}
						if frei(x, ny) {
							return x, ny, true
						}
					}
				}
			}
		}
		k := kandidaten[botZufall().Intn(len(kandidaten))]
		return k[0], k[1], true
	}

	for y := 0; y < schiffeGroesse; y++ {
		for x := 0; x < schiffeGroesse; x++ {
			if !frei(x, y) {
				continue
			}
			if botStufe(stufe) == botLeicht || (x+y)%2 == 0 {
				kandidaten = append(kandidaten, [2]int{x, y})
			}
		}
	}
	if len(kandidaten) == 0 {
		for y := 0; y < schiffeGroesse; y++ {
			for x := 0; x < schiffeGroesse; x++ {
				if frei(x, y) {
					kandidaten = append(kandidaten, [2]int{x, y})
				}
			}
		}
	}
	if len(kandidaten) == 0 {
		return 0, 0, false
	}
	k := kandidaten[botZufall().Intn(len(kandidaten))]
	return k[0], k[1], true
}
