package main

import "time"

/* ===========================================================================
   BOT FÜR SCHACH

   Der Bot rechnet auf demselben Zustand wie der Server und zieht durch
   dieselbe Prüfung (legalMovesAt). Er kennt keine versteckten Figuren.

   Stufen:
     leicht  – ein Zug voraus, zu 30 % ein zufälliger legaler Zug.
     mittel  – zwei Züge voraus (Alpha-Beta).
     schwer  – drei Züge voraus; Material, Figurenstellung und Bauernstruktur
               werden bewertet.
   =========================================================================== */

var chessBotWerte = map[string]int{"P": 100, "N": 320, "B": 330, "R": 500, "Q": 900, "K": 20000}

var chessBotPST = map[string][64]int{
	"P": {
		0, 0, 0, 0, 0, 0, 0, 0,
		50, 50, 50, 50, 50, 50, 50, 50,
		10, 10, 20, 30, 30, 20, 10, 10,
		5, 5, 10, 25, 25, 10, 5, 5,
		0, 0, 0, 20, 20, 0, 0, 0,
		5, -5, -10, 0, 0, -10, -5, 5,
		5, 10, 10, -20, -20, 10, 10, 5,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
	"N": {
		-50, -40, -30, -30, -30, -30, -40, -50,
		-40, -20, 0, 0, 0, 0, -20, -40,
		-30, 0, 10, 15, 15, 10, 0, -30,
		-30, 5, 15, 20, 20, 15, 5, -30,
		-30, 0, 15, 20, 20, 15, 0, -30,
		-30, 5, 10, 15, 15, 10, 5, -30,
		-40, -20, 0, 5, 5, 0, -20, -40,
		-50, -40, -30, -30, -30, -30, -40, -50,
	},
	"B": {
		-20, -10, -10, -10, -10, -10, -10, -20,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 5, 5, 10, 10, 5, 5, -10,
		-10, 0, 10, 10, 10, 10, 0, -10,
		-10, 10, 10, 10, 10, 10, 10, -10,
		-10, 5, 0, 0, 0, 0, 5, -10,
		-20, -10, -10, -10, -10, -10, -10, -20,
	},
	"R": {
		0, 0, 0, 0, 0, 0, 0, 0,
		5, 10, 10, 10, 10, 10, 10, 5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		0, 0, 0, 5, 5, 0, 0, 0,
	},
	"Q": {
		-20, -10, -10, -5, -5, -10, -10, -20,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-5, 0, 5, 5, 5, 5, 0, -5,
		0, 0, 5, 5, 5, 5, 0, -5,
		-10, 5, 5, 5, 5, 5, 0, -10,
		-10, 0, 5, 0, 0, 0, 0, -10,
		-20, -10, -10, -5, -5, -10, -10, -20,
	},
	"K": {
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-30, -40, -40, -50, -50, -40, -40, -30,
		-20, -30, -30, -40, -40, -30, -30, -20,
		-10, -20, -20, -20, -20, -20, -20, -10,
		20, 20, 0, 0, 0, 0, 20, 20,
		20, 30, 10, 0, 0, 10, 30, 20,
	},
}

func chessBotFarbe(g *ChessGame) string {
	if g == nil || g.BotID == "" {
		return ""
	}
	if g.RedID == g.BotID {
		return "w"
	}
	if g.BlueID == g.BotID {
		return "b"
	}
	return ""
}

func chessBotTreiben(g *ChessGame, jetzt time.Time) bool {
	farbe := chessBotFarbe(g)
	if farbe == "" || g.Status != "active" || g.State.Turn != farbe {
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
	zug, ok := chessBotZug(g, farbe, g.BotStufe)
	if !ok {
		return false
	}
	return chessBotSpielen(g, zug)
}

func chessBotZuege(s chessState, farbe string) []chessMove {
	zuege := []chessMove{}
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := s.Board[r][c]
			if p == nil || p.Color != farbe {
				continue
			}
			zuege = append(zuege, chessLegalMovesAt(s, r, c)...)
		}
	}
	return zuege
}

func chessBotZug(g *ChessGame, farbe, stufe string) (chessMove, bool) {
	zuege := chessBotZuege(g.State, farbe)
	if len(zuege) == 0 {
		return chessMove{}, false
	}
	stufe = botStufe(stufe)
	if stufe == botLeicht && botZufall().Intn(100) < 30 {
		return zuege[botZufall().Intn(len(zuege))], true
	}
	tiefe := 2
	switch stufe {
	case botLeicht:
		tiefe = 1
	case botSchwer:
		tiefe = 3
	}
	beste := zuege[0]
	besterWert := -1 << 30
	for _, zug := range zuege {
		next := chessApplyMove(g.State, zug, "Q")
		wert := -chessBotSuche(next, farbe, tiefe-1, -1<<30, 1<<30)
		if wert > besterWert || (wert == besterWert && botZufall().Intn(2) == 0) {
			beste, besterWert = zug, wert
		}
	}
	return beste, true
}

func chessBotSuche(s chessState, farbe string, tiefe, alpha, beta int) int {
	zuege := chessBotZuege(s, s.Turn)
	if len(zuege) == 0 {
		if chessInCheck(s, s.Turn) {
			if s.Turn == farbe {
				return -900000 - tiefe
			}
			return 900000 + tiefe
		}
		return 0
	}
	if tiefe <= 0 {
		return chessBotWert(s, farbe)
	}
	if s.Turn == farbe {
		best := -1 << 30
		for _, zug := range zuege {
			wert := chessBotSuche(chessApplyMove(s, zug, "Q"), farbe, tiefe-1, alpha, beta)
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
		wert := chessBotSuche(chessApplyMove(s, zug, "Q"), farbe, tiefe-1, alpha, beta)
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

func chessBotWert(s chessState, farbe string) int {
	wert := 0
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := s.Board[r][c]
			if p == nil {
				continue
			}
			index := r*8 + c
			if p.Color == "b" {
				index = (7-r)*8 + c
			}
			figur := chessBotWerte[p.Type] + chessBotPST[p.Type][index]
			if p.Color == farbe {
				wert += figur
			} else {
				wert -= figur
			}
		}
	}
	return wert
}

// chessBotSpielen führt den Zug wie der Handler aus: gleiche Prüfung, gleiche
// Anzeige, gleiches Spielende.
func chessBotSpielen(g *ChessGame, zug chessMove) bool {
	if g == nil || g.Status != "active" {
		return false
	}
	farbe := g.State.Turn
	if g.RedID != g.BotID && g.BlueID != g.BotID {
		return false
	}
	legal := false
	for _, m := range chessLegalMovesAt(g.State, zug.From.R, zug.From.C) {
		if m.To == zug.To {
			zug = m
			legal = true
			break
		}
	}
	if !legal {
		return false
	}
	piece := g.State.Board[zug.From.R][zug.From.C]
	if piece == nil || piece.Color != farbe {
		return false
	}
	captured := g.State.Board[zug.To.R][zug.To.C]
	if zug.EnPassant {
		captured = g.State.Board[zug.From.R][zug.To.C]
	}
	hatSchlag := captured != nil || zug.EnPassant
	notation := chessNotation(zug, piece, hatSchlag, "Q")
	if captured != nil {
		if farbe == "w" {
			g.CapturedW = append(g.CapturedW, *captured)
		} else {
			g.CapturedB = append(g.CapturedB, *captured)
		}
	}
	von, nach := zug.From, zug.To
	g.State = chessApplyMove(g.State, zug, "Q")
	g.LastFrom, g.LastTo = &von, &nach
	g.Moves = append(g.Moves, notation)
	if captured != nil || piece.Type == "P" {
		g.Ruhe = 0
	} else {
		g.Ruhe++
	}

	naechste := g.State.Turn
	inCheck := chessInCheck(g.State, naechste)
	hasMove := chessHasAnyMove(g.State, naechste)
	dreifach := chessRemisDreifach(g)
	switch {
	case dreifach:
		g.Status = "finished"
		g.Winner = ""
		g.Note = "Remis – dieselbe Stellung zum dritten Mal."
	case g.Ruhe >= 100:
		g.Status = "finished"
		g.Winner = ""
		g.Note = "Remis – 50 Züge ohne Schlag und ohne Bauernzug."
	case !hasMove && inCheck:
		g.Status = "finished"
		g.Winner = farbe
		g.Note = "SCHACHMATT! " + chessBotName(g, farbe) + " gewinnt!"
	case !hasMove:
		g.Status = "finished"
		g.Winner = ""
		g.Note = "PATT! Die Partie endet unentschieden."
	case inCheck:
		g.Note = "SCHACH! " + chessBotName(g, naechste) + " muss reagieren!"
	default:
		g.Note = chessBotName(g, naechste) + " ist am Zug."
	}
	return true
}

func chessBotName(g *ChessGame, farbe string) string {
	if farbe == "w" {
		return "Weiß"
	}
	return "Schwarz"
}
