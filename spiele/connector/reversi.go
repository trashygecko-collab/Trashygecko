package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Reversi: 8x8, Schwarz (Gastgeber) beginnt, 1 = schwarz, -1 = weiß.
var reversiDirs = [][2]int{
	{-1, -1}, {-1, 0}, {-1, 1},
	{0, -1}, {0, 1},
	{1, -1}, {1, 0}, {1, 1},
}

type ReversiGame struct {
	ID       string  `json:"id"`
	RedID    string  `json:"redId"`
	RedName  string  `json:"redName"`
	BlueID   string  `json:"blueId"`
	BlueName string  `json:"blueName"`
	Board    [][]int `json:"board"`
	Turn     int     `json:"turn"` // 1 = schwarz, -1 = weiß
	LastR    int     `json:"lastR"`
	LastC    int     `json:"lastC"`
	HasLast  bool    `json:"hasLast"`
	Status   string  `json:"status"`           // invited, active, finished
	Winner   string  `json:"winner,omitempty"` // black, white, draw
	Note     string  `json:"note,omitempty"`
}

func reversiNewGame(hostID, hostName, guestID, guestName, id string) *ReversiGame {
	board := make([][]int, 8)
	for r := range board {
		board[r] = make([]int, 8)
	}
	board[3][3] = -1
	board[3][4] = 1
	board[4][3] = 1
	board[4][4] = -1
	return &ReversiGame{
		ID:       id,
		RedID:    hostID,
		RedName:  hostName,
		BlueID:   guestID,
		BlueName: guestName,
		Board:    board,
		Turn:     1,
		Status:   "invited",
	}
}

func reversiClone(g *ReversiGame) *ReversiGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Board = make([][]int, len(g.Board))
	for r := range g.Board {
		c.Board[r] = append([]int(nil), g.Board[r]...)
	}
	return &c
}

func reversiColorOf(g *ReversiGame, id string) string {
	if g.RedID == id {
		return "black"
	}
	if g.BlueID == id {
		return "white"
	}
	return ""
}

func reversiStone(color string) int {
	if color == "black" {
		return 1
	}
	return -1
}

func reversiName(g *ReversiGame, stone int) string {
	if stone == 1 {
		return g.RedName
	}
	return g.BlueName
}

func reversiWord(stone int) string {
	if stone == 1 {
		return "Schwarz"
	}
	return "Weiß"
}

func reversiInside(r, c int) bool {
	return r >= 0 && r < 8 && c >= 0 && c < 8
}

func reversiFlips(board [][]int, r, c, color int) [][2]int {
	flips := [][2]int{}
	if board[r][c] != 0 {
		return flips
	}
	for _, dir := range reversiDirs {
		rr := r + dir[0]
		cc := c + dir[1]
		line := [][2]int{}
		for reversiInside(rr, cc) && board[rr][cc] == -color {
			line = append(line, [2]int{rr, cc})
			rr += dir[0]
			cc += dir[1]
		}
		if len(line) > 0 && reversiInside(rr, cc) && board[rr][cc] == color {
			flips = append(flips, line...)
		}
	}
	return flips
}

func reversiHasMove(board [][]int, color int) bool {
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if len(reversiFlips(board, r, c, color)) > 0 {
				return true
			}
		}
	}
	return false
}

func reversiCount(board [][]int) (black, white int) {
	for _, row := range board {
		for _, v := range row {
			if v == 1 {
				black++
			} else if v == -1 {
				white++
			}
		}
	}
	return black, white
}

func reversiFinish(g *ReversiGame) {
	black, white := reversiCount(g.Board)
	g.Status = "finished"
	if black > white {
		g.Winner = "black"
	} else if white > black {
		g.Winner = "white"
	} else {
		g.Winner = "draw"
	}
}

func reversiHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	if code == "" {
		http.Error(w, "Spielraum-Code fehlt", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		mu.Lock()
		room := rooms[code]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		if _, ok := room.Players[id]; !ok {
			mu.Unlock()
			http.Error(w, "Bitte zuerst dem Spielraum beitreten.", http.StatusForbidden)
			return
		}
		game := reversiClone(room.Reversi)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID   string `json:"id"`
		ToID string `json:"toId"`
		R    int    `json:"r"`
		C    int    `json:"c"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.ID = strings.TrimSpace(input.ID)
	input.ToID = strings.TrimSpace(input.ToID)
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

	mu.Lock()
	room := rooms[code]
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	room.cleanPlayers()
	player, ok := room.Players[input.ID]
	if !ok {
		mu.Unlock()
		http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
		return
	}

	switch action {
	case "invite":
		target, ok := room.Players[input.ToID]
		if !ok || input.ToID == input.ID {
			mu.Unlock()
			http.Error(w, "Dieser Spieler ist nicht verfügbar.", http.StatusBadRequest)
			return
		}
		if room.Reversi != nil && !room.gameSlotFree(room.Reversi.Status, room.Reversi.RedID, room.Reversi.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Reversi-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Reversi = reversiNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Reversi ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "reversi",
			GameID:     room.Reversi.ID,
		})

	case "accept":
		game := room.Reversi
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = ""

	case "decline":
		game := room.Reversi
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "move":
		game := room.Reversi
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := reversiColorOf(game, input.ID)
		if color == "" {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if reversiStone(color) != game.Turn {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		if !reversiInside(input.R, input.C) {
			mu.Unlock()
			http.Error(w, "Dieses Feld liegt nicht auf dem Brett.", http.StatusBadRequest)
			return
		}
		stone := reversiStone(color)
		flips := reversiFlips(game.Board, input.R, input.C, stone)
		if len(flips) == 0 {
			mu.Unlock()
			http.Error(w, "Dort ist kein gültiger Zug möglich.", http.StatusConflict)
			return
		}

		game.Board[input.R][input.C] = stone
		for _, f := range flips {
			game.Board[f[0]][f[1]] = stone
		}
		game.LastR, game.LastC, game.HasLast = input.R, input.C, true
		game.Turn = -stone
		game.Note = player.Name + " setzt und dreht " + reversiNumber(len(flips)) + "."

		// Aussetzen und Spielende prüfen.
		if !reversiHasMove(game.Board, game.Turn) {
			if !reversiHasMove(game.Board, -game.Turn) {
				reversiFinish(game)
				break
			}
			game.Note = reversiWord(game.Turn) + " kann nicht setzen und muss aussetzen."
			game.Turn = -game.Turn
		}

	case "pass":
		game := room.Reversi
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := reversiColorOf(game, input.ID)
		if color == "" {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if reversiStone(color) != game.Turn {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		if reversiHasMove(game.Board, game.Turn) {
			mu.Unlock()
			http.Error(w, "Du kannst setzen und musst nicht aussetzen.", http.StatusConflict)
			return
		}
		if !reversiHasMove(game.Board, -game.Turn) {
			reversiFinish(game)
			break
		}
		game.Turn = -game.Turn
		game.Note = player.Name + " muss aussetzen."

	case "reset":
		game := room.Reversi
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		fresh := reversiNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
		fresh.Status = "active"
		room.Reversi = fresh

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Reversi-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := reversiClone(room.Reversi)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

func reversiNumber(n int) string {
	if n == 1 {
		return "einen Stein"
	}
	digits := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	if n >= 0 && n < len(digits) {
		return digits[n] + " Steine"
	}
	return "viele Steine"
}
