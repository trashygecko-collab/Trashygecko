package main

import "time"

/* ===========================================================================
   BOT FÜR VIER GEWINNT

   Stufen:
     leicht  – nimmt meist eine zufällige Spalte; nur selten sieht er einen
               direkten Sieg. So macht er sichtbare Fehler.
     mittel  – gewinnt sofort, blockt sonst, prüft danach zwei Züge voraus.
     schwer  – Suchbaum fünf Züge tief (Alpha-Beta) mit Bewertung der
               Viererfenster; verliert dadurch praktisch nicht.
   =========================================================================== */

// vierGewinntBotTreiben lässt den Bot setzen, wenn seine Denkpause vorbei ist.
func vierGewinntBotTreiben(g *VierGewinntGame, jetzt time.Time) bool {
	if g == nil || g.Status != "active" || g.BotID == "" {
		return false
	}
	color := ""
	if g.Current == "red" && g.RedID == g.BotID {
		color = "red"
	} else if g.Current == "blue" && g.BlueID == g.BotID {
		color = "blue"
	}
	if color == "" {
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
	col := vierGewinntBotSpalte(g.Grid, color, g.BotStufe)
	if col < 0 {
		return false
	}
	nr, _ := vierGewinntSetzen(g, color, col)
	return nr == 0
}

// vierGewinntBotSpalte wählt eine Spalte (0–6). Volle Spalten fallen weg.
func vierGewinntBotSpalte(grid [][]string, color, stufe string) int {
	stufe = botStufe(stufe)
	frei := []int{}
	for c := 0; c < 7; c++ {
		if grid[0][c] == "" {
			frei = append(frei, c)
		}
	}
	if len(frei) == 0 {
		return -1
	}
	gegner := "blue"
	if color == "blue" {
		gegner = "red"
	}

	if stufe == botLeicht {
		z := botZufall()
		if z.Intn(100) < 20 {
			for _, c := range frei {
				if vierGewinntGewinnProbe(grid, c, color) {
					return c
				}
			}
		}
		return frei[botZufall().Intn(len(frei))]
	}

	for _, c := range frei {
		if vierGewinntGewinnProbe(grid, c, color) {
			return c
		}
	}
	for _, c := range frei {
		if vierGewinntGewinnProbe(grid, c, gegner) {
			return c
		}
	}

	tiefe := 2
	if stufe == botSchwer {
		tiefe = 4
	}
	beste, besteWert := -1, -(1 << 30)
	for _, c := range frei {
		probe := vierGewinntKopie(grid)
		zeile := vierGewinntFallzeile(probe, c)
		probe[zeile][c] = color
		wert := -vierGewinntSuche(probe, gegner, color, tiefe-1, -(1 << 30), 1<<30)
		if wert > besteWert {
			beste, besteWert = c, wert
		} else if wert == besteWert && botZufall().Intn(2) == 0 {
			beste = c
		}
	}
	return beste
}

// vierGewinntGewinnProbe: gewinnt ein Stein in dieser Spalte sofort?
func vierGewinntGewinnProbe(grid [][]string, col int, color string) bool {
	zeile := vierGewinntFallzeile(grid, col)
	if zeile < 0 {
		return false
	}
	grid[zeile][col] = color
	gewonnen := checkVierGewinntWin(grid, zeile, col, color) != nil
	grid[zeile][col] = ""
	return gewonnen
}

func vierGewinntFallzeile(grid [][]string, col int) int {
	for r := 5; r >= 0; r-- {
		if grid[r][col] == "" {
			return r
		}
	}
	return -1
}

func vierGewinntKopie(grid [][]string) [][]string {
	k := make([][]string, 6)
	for r := range grid {
		k[r] = append([]string(nil), grid[r]...)
	}
	return k
}

// vierGewinntSuche bewertet aus Sicht von „wer“ (der gerade ziehen will).
func vierGewinntSuche(grid [][]string, wer, gegner string, tiefe, alpha, beta int) int {
	if tiefe == 0 {
		return vierGewinntBewerten(grid, wer, gegner)
	}
	zuege := 0
	best := -(1 << 29)
	for c := 0; c < 7; c++ {
		if grid[0][c] != "" {
			continue
		}
		zuege++
		r := vierGewinntFallzeile(grid, c)
		grid[r][c] = wer
		var wert int
		if checkVierGewinntWin(grid, r, c, wer) != nil {
			wert = 100000 + tiefe
		} else {
			wert = -vierGewinntSuche(grid, gegner, wer, tiefe-1, -beta, -alpha)
		}
		grid[r][c] = ""
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
	if zuege == 0 {
		return 0
	}
	return best
}

// vierGewinntBewerten zählt Viererfenster: eigene Steine positiv, fremde
// negativ; ein Fenster mit beiden Farben zählt nicht.
func vierGewinntBewerten(grid [][]string, wer, gegner string) int {
	summe := 0
	fenster := func(zellen [4][2]int) {
		eigen, fremd := 0, 0
		for _, z := range zellen {
			switch grid[z[0]][z[1]] {
			case wer:
				eigen++
			case gegner:
				fremd++
			}
		}
		switch {
		case eigen > 0 && fremd > 0:
			return
		case eigen == 3:
			summe += 50
		case eigen == 2:
			summe += 10
		case eigen == 1:
			summe += 1
		case fremd == 3:
			summe -= 60
		case fremd == 2:
			summe -= 12
		case fremd == 1:
			summe -= 1
		}
	}
	for r := 0; r < 6; r++ {
		for c := 0; c < 7; c++ {
			if c+3 < 7 {
				fenster([4][2]int{{r, c}, {r, c + 1}, {r, c + 2}, {r, c + 3}})
			}
			if r+3 < 6 {
				fenster([4][2]int{{r, c}, {r + 1, c}, {r + 2, c}, {r + 3, c}})
			}
			if r+3 < 6 && c+3 < 7 {
				fenster([4][2]int{{r, c}, {r + 1, c + 1}, {r + 2, c + 2}, {r + 3, c + 3}})
			}
			if r+3 < 6 && c-3 >= 0 {
				fenster([4][2]int{{r, c}, {r + 1, c - 1}, {r + 2, c - 2}, {r + 3, c - 3}})
			}
		}
	}
	return summe
}
