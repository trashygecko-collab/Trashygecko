package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type VierGewinntGame struct {
	ID      string     `json:"id"`
	RedID   string     `json:"redId"`
	RedName string     `json:"redName"`
	BlueID  string     `json:"blueId"`
	BlueName string    `json:"blueName"`
	Grid    [][]string `json:"grid"` // 6 rows, 7 cols; "" for empty, "red" or "blue"
	Current string     `json:"current"` // "red" or "blue"
	Status  string     `json:"status"` // "invited", "active", "declined", "finished"
	Winner  string     `json:"winner,omitempty"` // "red", "blue", or "draw"
	Winning [][]int    `json:"winning,omitempty"` // winning cell coordinates [[r,c],...]
}

// checkVierGewinntWin checks if placing a disc at (r, c) leads to 4 in a row.
func checkVierGewinntWin(grid [][]string, r, c int, color string) [][]int {
	directions := [][2]int{
		{0, 1},  // horizontal
		{1, 0},  // vertical
		{1, 1},  // diagonal down-right
		{1, -1}, // diagonal down-left
	}

	for _, d := range directions {
		dr, dc := d[0], d[1]
		cells := [][]int{{r, c}}

		for _, sign := range []int{-1, 1} {
			nr := r + dr*sign
			nc := c + dc*sign
			for nr >= 0 && nr < 6 && nc >= 0 && nc < 7 && grid[nr][nc] == color {
				cells = append(cells, []int{nr, nc})
				nr += dr * sign
				nc += dc * sign
			}
		}

		if len(cells) >= 4 {
			return cells[:4]
		}
	}
	return nil
}

func newVierGewinntGrid() [][]string {
	grid := make([][]string, 6)
	for i := range grid {
		grid[i] = make([]string, 7)
	}
	return grid
}

func vierGewinntHandler(w http.ResponseWriter, r *http.Request) {
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
		game := room.VierGewinnt
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID     string `json:"id"`
		ToID   string `json:"toId"`
		Column int    `json:"column"`
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
		if room.VierGewinnt != nil && (room.VierGewinnt.Status == "invited" || room.VierGewinnt.Status == "active") {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Vier-gewinnt-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.VierGewinnt = &VierGewinntGame{
			ID:       randomCode(),
			RedID:    input.ID,
			RedName:  player.Name,
			BlueID:   input.ToID,
			BlueName: target.Name,
			Grid:     newVierGewinntGrid(),
			Current:  "red",
			Status:   "invited",
		}
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Vier gewinnt ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "viergewinnt",
			GameID:     room.VierGewinnt.ID,
		})

	case "accept":
		game := room.VierGewinnt
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"

	case "decline":
		game := room.VierGewinnt
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "move":
		game := room.VierGewinnt
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := ""
		if input.ID == game.RedID {
			color = "red"
		} else if input.ID == game.BlueID {
			color = "blue"
		}
		if color == "" {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Current != color {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		col := input.Column
		if col < 0 || col > 6 {
			mu.Unlock()
			http.Error(w, "Ungültige Spalte.", http.StatusBadRequest)
			return
		}

		// Find lowest free row in column
		row := -1
		for rIdx := 5; rIdx >= 0; rIdx-- {
			if game.Grid[rIdx][col] == "" {
				row = rIdx
				break
			}
		}
		if row == -1 {
			mu.Unlock()
			http.Error(w, "Diese Spalte ist voll.", http.StatusBadRequest)
			return
		}

		game.Grid[row][col] = color

		// Check win
		if winCells := checkVierGewinntWin(game.Grid, row, col, color); len(winCells) >= 4 {
			game.Winner = color
			game.Winning = winCells
			game.Status = "finished"
		} else {
			// Check board full
			full := true
			for cIdx := 0; cIdx < 7; cIdx++ {
				if game.Grid[0][cIdx] == "" {
					full = false
					break
				}
			}
			if full {
				game.Winner = "draw"
				game.Status = "finished"
			} else if color == "red" {
				game.Current = "blue"
			} else {
				game.Current = "red"
			}
		}

	case "reset":
		game := room.VierGewinnt
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		game.Grid = newVierGewinntGrid()
		game.Current = "red"
		game.Winner = ""
		game.Winning = nil
		game.Status = "active"

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Vier-gewinnt-Aktion.", http.StatusBadRequest)
		return
	}

	result := room.VierGewinnt
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
