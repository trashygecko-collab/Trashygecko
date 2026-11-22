package main

import "time"

/* ===========================================================================
   BOT FÜR DAME

   Der Bot erzeugt seine Züge aus dem Brett und schickt sie durch dieselbe
   Brettprüfung wie ein Mensch. Geschlagen wird wie auf der Spielseite: eine
   Figur springt über einen Gegner, eine Dame darf über das ganze Feld gleiten
   und dabei auch mehrere Gegner in einer Linie nehmen.

   Stufen:
     leicht  – nimmt zu 40 % irgendeinen erlaubten Zug, sonst einen Schritt
               voraus (schlagen vor laufen).
     mittel  – zwei Züge voraus.
     schwer  – drei Züge voraus mit Alpha-Beta; Damen und die eigene Grundreihe
               zählen dabei mehr.
   =========================================================================== */

type dameBotZug struct {
	vonR, vonC   int
	nachR, nachC int
}

func dameBotFarbe(g *DameGame) string {
	if g == nil || g.BotID == "" {
		return ""
	}
	if g.RedID == g.BotID {
		return "red"
	}
	if g.BlueID == g.BotID {
		return "blue"
	}
	return ""
}

func dameBotTreiben(g *DameGame, jetzt time.Time) bool {
	farbe := dameBotFarbe(g)
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
	zug, ok := dameBotZugWahl(g, farbe, g.BotStufe)
	if !ok {
		return false
	}
	return dameBotAnwenden(g, farbe, zug)
}

func dameBotZuege(board [][]map[string]any, farbe string) []dameBotZug {
	zuege := []dameBotZug{}
	if len(board) != 8 {
		return zuege
	}
	gegner := otherDameColor(farbe)
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if dameColor(board[r][c]) != farbe {
				continue
			}
			king := dameIsKing(board[r][c])
			for _, d := range [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
				if !king {
					// Ein Schritt auf ein freies Feld.
					nr, nc := r+d[0], c+d[1]
					if nr >= 0 && nr < 8 && nc >= 0 && nc < 8 && board[nr][nc] == nil {
						zuege = append(zuege, dameBotZug{vonR: r, vonC: c, nachR: nr, nachC: nc})
					}
					// Sprung über genau einen Gegner.
					mr, mc := r+d[0], c+d[1]
					nr, nc = r+2*d[0], c+2*d[1]
					if nr >= 0 && nr < 8 && nc >= 0 && nc < 8 &&
						dameColor(board[mr][mc]) == gegner && board[nr][nc] == nil {
						zuege = append(zuege, dameBotZug{vonR: r, vonC: c, nachR: nr, nachC: nc})
					}
					continue
				}
				// Dame: entlang der Diagonale gleiten und dabei Gegner sammeln.
				for step := 1; step < 8; step++ {
					nr, nc := r+d[0]*step, c+d[1]*step
					if nr < 0 || nr >= 8 || nc < 0 || nc >= 8 {
						break
					}
					cell := board[nr][nc]
					if cell == nil {
						zuege = append(zuege, dameBotZug{vonR: r, vonC: c, nachR: nr, nachC: nc})
						continue
					}
					if dameColor(cell) == farbe {
						break
					}
				}
			}
		}
	}
	return zuege
}

func dameBotZugWahl(g *DameGame, farbe, stufe string) (dameBotZug, bool) {
	zuege := dameBotZuege(g.Board, farbe)
	if len(zuege) == 0 {
		return dameBotZug{}, false
	}
	if botStufe(stufe) == botLeicht && botZufall().Intn(100) < 40 {
		return zuege[botZufall().Intn(len(zuege))], true
	}
	tiefe := 2
	switch botStufe(stufe) {
	case botLeicht:
		tiefe = 1
	case botSchwer:
		tiefe = 3
	}
	beste := zuege[0]
	besterWert := -1 << 30
	for _, zug := range zuege {
		probe := cloneDame(g)
		if !dameBotAnwenden(probe, farbe, zug) {
			continue
		}
		wert := dameBotSuche(probe, farbe, tiefe-1, -1<<30, 1<<30)
		if wert > besterWert || (wert == besterWert && botZufall().Intn(2) == 0) {
			beste, besterWert = zug, wert
		}
	}
	return beste, true
}

func dameBotAnwenden(g *DameGame, farbe string, zug dameBotZug) bool {
	if g == nil || g.Status != "active" || g.Current != farbe {
		return false
	}
	if zug.vonR < 0 || zug.vonR >= 8 || zug.vonC < 0 || zug.vonC >= 8 ||
		zug.nachR < 0 || zug.nachR >= 8 || zug.nachC < 0 || zug.nachC >= 8 {
		return false
	}
	if dameColor(g.Board[zug.vonR][zug.vonC]) != farbe || g.Board[zug.nachR][zug.nachC] != nil {
		return false
	}
	dr, dc := zug.nachR-zug.vonR, zug.nachC-zug.vonC
	if dr == 0 || absInt(dr) != absInt(dc) {
		return false
	}
	schritte := absInt(dr)
	sr, sc := 1, 1
	if dr < 0 {
		sr = -1
	}
	if dc < 0 {
		sc = -1
	}
	gegner := otherDameColor(farbe)
	opfer := [][2]int{}
	for i := 1; i < schritte; i++ {
		r, c := zug.vonR+sr*i, zug.vonC+sc*i
		if g.Board[r][c] == nil {
			continue
		}
		if dameColor(g.Board[r][c]) == farbe {
			return false
		}
		opfer = append(opfer, [2]int{r, c})
	}
	king := dameIsKing(g.Board[zug.vonR][zug.vonC])
	if !king {
		if schritte > 2 {
			return false
		}
		if schritte == 2 && len(opfer) != 1 {
			return false
		}
	}
	if len(opfer) != 1 && len(opfer) != 0 || schritte == 2 && len(opfer) == 0 {
		// Nur ein einzelner Gegner ist erlaubt; eine Dame darf mehrere nehmen.
		if !king {
			return false
		}
	}
	for _, o := range opfer {
		g.Board[o[0]][o[1]] = nil
	}
	g.Board[zug.nachR][zug.nachC] = map[string]any{"color": farbe, "king": king}
	if !king {
		if farbe == "red" && zug.nachR == 0 {
			g.Board[zug.nachR][zug.nachC] = map[string]any{"color": farbe, "king": true}
		}
		if farbe == "blue" && zug.nachR == 7 {
			g.Board[zug.nachR][zug.nachC] = map[string]any{"color": farbe, "king": true}
		}
	}
	g.Board[zug.vonR][zug.vonC] = nil
	if len(opfer) > 0 {
		g.Ruhe = 0
	} else {
		g.Ruhe++
	}
	g.Current = gegner
	if g.Ruhe >= 100 {
		g.Status = "finished"
		g.Winner = "draw"
		return true
	}
	if !dameHasAnyMove(g.Board, gegner) {
		g.Winner = farbe
		g.Status = "finished"
	}
	return true
}

func dameBotSuche(g *DameGame, farbe string, tiefe, alpha, beta int) int {
	if g.Status == "finished" || tiefe <= 0 {
		return dameBotWert(g, farbe)
	}
	zuege := dameBotZuege(g.Board, g.Current)
	if len(zuege) == 0 {
		return dameBotWert(g, farbe)
	}
	if g.Current == farbe {
		best := -1 << 30
		for _, zug := range zuege {
			probe := cloneDame(g)
			if !dameBotAnwenden(probe, g.Current, zug) {
				continue
			}
			wert := dameBotSuche(probe, farbe, tiefe-1, alpha, beta)
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
		probe := cloneDame(g)
		if !dameBotAnwenden(probe, g.Current, zug) {
			continue
		}
		wert := dameBotSuche(probe, farbe, tiefe-1, alpha, beta)
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

func dameBotWert(g *DameGame, farbe string) int {
	if g.Status == "finished" {
		if g.Winner == farbe {
			return 1000000
		}
		if g.Winner == "" {
			return 0
		}
		return -1000000
	}
	wert := 0
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			cell := g.Board[r][c]
			color := dameColor(cell)
			if color == "" {
				continue
			}
			faktor := 1
			if color != farbe {
				faktor = -1
			}
			figur := 100
			if dameIsKing(cell) {
				figur = 185
			} else if color == "red" {
				figur += (7 - r) * 5
			} else {
				figur += r * 5
			}
			// Die eigene Grundreihe ist Deckung.
			if color == "red" && r == 7 {
				figur += 8
			}
			if color == "blue" && r == 0 {
				figur += 8
			}
			// Mitte ist mehr wert als der Rand.
			if c >= 2 && c <= 5 {
				figur += 4
			}
			wert += faktor * figur
		}
	}
	return wert
}
