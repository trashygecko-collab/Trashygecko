package main

import "time"

/* ===========================================================================
   BOT FÜR NEON TIC TAC TOE

   Der Bot spielt „O“, der Mensch „X“ (oder umgekehrt, wenn der Mensch als
   Zweiter beitritt – die Kennung entscheidet). Er sieht nur das Brett, also
   genau das, was ein Mensch sieht.

   Stufen:
     leicht  – gewinnt, wenn es direkt geht; sonst spielt er zu 35 % irgendwo.
     mittel  – gewinnt sofort, blockt sonst, mag Mitte und Ecken.
     schwer  – vollständiges Minimax: verliert nie, gewinnt, wenn der Gegner
               einen Fehler macht.
   =========================================================================== */

// ticTacToeBotTreiben lässt den Bot ziehen, sobald seine Denkpause vorbei ist.
// Aufgerufen beim Abholen des Spielstands – der Browser fragt im Sekundentakt.
func ticTacToeBotTreiben(g *TicTacToeGame, jetzt time.Time) bool {
	if g == nil || g.Status != "active" || g.BotID == "" {
		return false
	}
	mark := ""
	if g.Current == "X" && g.XID == g.BotID {
		mark = "X"
	} else if g.Current == "O" && g.OID == g.BotID {
		mark = "O"
	}
	if mark == "" {
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
	index := ticTacToeBotZug(g.Board, mark, g.BotStufe)
	if index < 0 {
		return false
	}
	nr, _ := ticTacToeSetzen(g, mark, index)
	return nr == 0
}

// ticTacToeBotZug wählt ein Feld. Nur freie Felder, nie ein fremdes Zeichen.
func ticTacToeBotZug(board []string, mark, stufe string) int {
	stufe = botStufe(stufe)
	frei := []int{}
	for i, f := range board {
		if f == "" {
			frei = append(frei, i)
		}
	}
	if len(frei) == 0 {
		return -1
	}
	if stufe == botLeicht && botZufall().Intn(100) < 35 {
		return frei[botZufall().Intn(len(frei))]
	}

	gegner := "O"
	if mark == "O" {
		gegner = "X"
	}
	// Gewinnen geht immer vor.
	if i := ticTacToeGewinnfeld(board, mark); i >= 0 {
		return i
	}
	// Blocken, was der Gegner sonst gewinnt.
	if i := ticTacToeGewinnfeld(board, gegner); i >= 0 {
		return i
	}
	if stufe == botMittel || stufe == botLeicht {
		return ticTacToeMittelZug(board, frei)
	}
	// Schwer: Minimax. Gleich gute Züge werden gemischt, damit nicht jede
	// Partie gleich aussieht.
	beste, besteWerte := -1, -1000
	for _, i := range frei {
		board[i] = mark
		wert := ticTacToeMinimax(board, mark, gegner, 0, false)
		board[i] = ""
		if wert > besteWerte {
			beste, besteWerte = i, wert
		} else if wert == besteWerte {
			z := botZufall()
			if z.Intn(2) == 0 {
				beste = i
			}
		}
	}
	return beste
}

// ticTacToeGewinnfeld liefert das Feld, mit dem mark sofort gewinnt (–1: keins).
func ticTacToeGewinnfeld(board []string, mark string) int {
	for i := range board {
		if board[i] != "" {
			continue
		}
		board[i] = mark
		gewonnen := ticTacToeWinner(board) == mark
		board[i] = ""
		if gewonnen {
			return i
		}
	}
	return -1
}

// ticTacToeMittelZug: Mitte, dann Ecken, dann Kanten.
func ticTacToeMittelZug(board []string, frei []int) int {
	for _, i := range []int{4, 0, 2, 6, 8, 1, 3, 5, 7} {
		if board[i] == "" {
			return i
		}
	}
	return frei[botZufall().Intn(len(frei))]
}

// ticTacToeMinimax bewertet aus Sicht von mark. Tiefe zählt mit, damit ein
// schneller Sieg besser ist als ein späterer und eine Niederlage möglichst
// spät kommt.
func ticTacToeMinimax(board []string, mark, gegner string, tiefe int, amZug bool) int {
	if w := ticTacToeWinner(board); w != "" {
		if w == mark {
			return 10 - tiefe
		}
		return tiefe - 10
	}
	voll := true
	for _, f := range board {
		if f == "" {
			voll = false
			break
		}
	}
	if voll {
		return 0
	}
	wer := gegner
	if amZug {
		wer = mark
	}
	best := 100
	if amZug {
		best = -100
	}
	for i := range board {
		if board[i] != "" {
			continue
		}
		board[i] = wer
		wert := ticTacToeMinimax(board, mark, gegner, tiefe+1, !amZug)
		board[i] = ""
		if amZug {
			if wert > best {
				best = wert
			}
		} else if wert < best {
			best = wert
		}
	}
	return best
}
