package main

import "time"

/* ===========================================================================
   BOT FÜR REVERSI

   Stufen:
     leicht  – nimmt zu 40 % irgendeinen erlaubten Zug, sonst Ecken und
               möglichst viele Steine.
     mittel  – zwei Züge voraus (Alpha-Beta).
     schwer  – vier Züge voraus; Ecken und Kanten zählen dabei besonders,
               weil sie nicht mehr umgedreht werden können.
   =========================================================================== */

var reversiFeldWert = [8][8]int{
	{120, -20, 20, 5, 5, 20, -20, 120},
	{-20, -40, -5, -5, -5, -5, -40, -20},
	{20, -5, 15, 3, 3, 15, -5, 20},
	{5, -5, 3, 3, 3, 3, -5, 5},
	{5, -5, 3, 3, 3, 3, -5, 5},
	{20, -5, 15, 3, 3, 15, -5, 20},
	{-20, -40, -5, -5, -5, -5, -40, -20},
	{120, -20, 20, 5, 5, 20, -20, 120},
}

type reversiBotZug struct{ r, c int }

func reversiBotFarbe(g *ReversiGame) string {
	if g == nil || g.BotID == "" {
		return ""
	}
	return reversiColorOf(g, g.BotID)
}

func reversiBotTreiben(g *ReversiGame, jetzt time.Time) bool {
	farbe := reversiBotFarbe(g)
	if farbe == "" || g.Status != "active" || reversiStone(farbe) != g.Turn {
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
	zug, ok := reversiBotZugWahl(g, farbe, g.BotStufe)
	if !ok {
		return false
	}
	return reversiBotSetzen(g, farbe, zug)
}

func reversiBotZuege(board [][]int, stein int) []reversiBotZug {
	zuege := []reversiBotZug{}
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if len(reversiFlips(board, r, c, stein)) > 0 {
				zuege = append(zuege, reversiBotZug{r: r, c: c})
			}
		}
	}
	return zuege
}

func reversiBotZugWahl(g *ReversiGame, farbe, stufe string) (reversiBotZug, bool) {
	stein := reversiStone(farbe)
	zuege := reversiBotZuege(g.Board, stein)
	if len(zuege) == 0 {
		return reversiBotZug{}, false
	}
	stufe = botStufe(stufe)
	if stufe == botLeicht && botZufall().Intn(100) < 40 {
		return zuege[botZufall().Intn(len(zuege))], true
	}
	tiefe := 2
	switch stufe {
	case botLeicht:
		tiefe = 1
	case botSchwer:
		tiefe = 4
	}
	beste := zuege[0]
	besterWert := -1 << 30
	for _, zug := range zuege {
		probe := reversiClone(g)
		reversiBotAnwenden(probe, zug)
		wert := reversiBotSuche(probe, stein, tiefe-1, -1<<30, 1<<30)
		if wert > besterWert || (wert == besterWert && botZufall().Intn(2) == 0) {
			beste, besterWert = zug, wert
		}
	}
	return beste, true
}

// reversiBotAnwenden setzt einen Stein und dreht – dieselben Regeln wie im
// Handler, inklusive Aussetzen, wenn der Gegner nicht setzen kann.
func reversiBotAnwenden(g *ReversiGame, zug reversiBotZug) bool {
	stein := g.Turn
	flips := reversiFlips(g.Board, zug.r, zug.c, stein)
	if len(flips) == 0 {
		return false
	}
	g.Board[zug.r][zug.c] = stein
	for _, f := range flips {
		g.Board[f[0]][f[1]] = stein
	}
	g.LastR, g.LastC, g.HasLast = zug.r, zug.c, true
	g.Turn = -stein
	g.Note = reversiName(g, stein) + " setzt und dreht " + reversiNumber(len(flips)) + "."
	if !reversiHasMove(g.Board, g.Turn) {
		if !reversiHasMove(g.Board, -g.Turn) {
			reversiFinish(g)
			return true
		}
		g.Note = reversiWord(g.Turn) + " kann nicht setzen und muss aussetzen."
		g.Turn = -g.Turn
	}
	return true
}

func reversiBotSetzen(g *ReversiGame, farbe string, zug reversiBotZug) bool {
	if g == nil || g.Status != "active" || reversiStone(farbe) != g.Turn {
		return false
	}
	return reversiBotAnwenden(g, zug)
}

func reversiBotSuche(g *ReversiGame, stein, tiefe, alpha, beta int) int {
	if g.Status == "finished" || tiefe <= 0 {
		return reversiBotWert(g, stein)
	}
	zuege := reversiBotZuege(g.Board, g.Turn)
	if len(zuege) == 0 {
		// Aussetzen ist im Zustand schon erledigt; hier nur die Bewertung.
		return reversiBotWert(g, stein)
	}
	if g.Turn == stein {
		best := -1 << 30
		for _, zug := range zuege {
			probe := reversiClone(g)
			reversiBotAnwenden(probe, zug)
			wert := reversiBotSuche(probe, stein, tiefe-1, alpha, beta)
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
	for _, zug := range zuege {
		probe := reversiClone(g)
		reversiBotAnwenden(probe, zug)
		wert := reversiBotSuche(probe, stein, tiefe-1, alpha, beta)
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

func reversiBotWert(g *ReversiGame, stein int) int {
	if g.Status == "finished" {
		schwarz, weiss := reversiCount(g.Board)
		eigen, fremd := schwarz, weiss
		if stein == -1 {
			eigen, fremd = weiss, schwarz
		}
		if eigen > fremd {
			return 1000000 + (eigen-fremd)*100
		}
		if eigen < fremd {
			return -1000000 - (fremd-eigen)*100
		}
		return 0
	}
	wert := 0
	eigen, fremd := 0, 0
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			switch g.Board[r][c] {
			case stein:
				eigen++
				wert += reversiFeldWert[r][c]
			case -stein:
				fremd++
				wert -= reversiFeldWert[r][c]
			}
		}
	}
	wert += (eigen - fremd) * 2
	// Beweglichkeit: wer mehr Möglichkeiten hat, bestimmt das Spiel.
	wert += (len(reversiBotZuege(g.Board, stein)) - len(reversiBotZuege(g.Board, -stein))) * 8
	// Am Ende zählen die Steine, vorher die Stellung.
	if eigen+fremd > 54 {
		wert += (eigen - fremd) * 12
	}
	return wert
}
