package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type VierGewinntGame struct {
	ID       string     `json:"id"`
	RedID    string     `json:"redId"`
	RedName  string     `json:"redName"`
	BlueID   string     `json:"blueId"`
	BlueName string     `json:"blueName"`
	Grid     [][]string `json:"grid"`              // 6 rows, 7 cols; "" for empty, "red" or "blue"
	Current  string     `json:"current"`           // "red" or "blue"
	Status   string     `json:"status"`            // "invited", "active", "declined", "finished"
	Winner   string     `json:"winner,omitempty"`  // "red", "blue", or "draw"
	Winning  [][]int    `json:"winning,omitempty"` // winning cell coordinates [[r,c],...]
	// Bot: Kennung, Stufe und Ziehzeit (nicht im JSON).
	BotID    string    `json:"botId,omitempty"`
	BotStufe string    `json:"-"`
	BotFrist time.Time `json:"-"`
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

// vierGewinntSetzen prüft einen Zug und setzt ihn – für Mensch und Bot
// derselbe Weg. Rückgabe: HTTP-Nummer und Text (0 = gesetzt).
func vierGewinntSetzen(g *VierGewinntGame, color string, col int) (int, string) {
	if g == nil || g.Status != "active" {
		return http.StatusConflict, "Es läuft gerade keine aktive Partie."
	}
	if g.Current != color {
		return http.StatusConflict, "Der andere Spieler ist an der Reihe."
	}
	if col < 0 || col > 6 {
		return http.StatusBadRequest, "Ungültige Spalte."
	}
	row := -1
	for rIdx := 5; rIdx >= 0; rIdx-- {
		if g.Grid[rIdx][col] == "" {
			row = rIdx
			break
		}
	}
	if row == -1 {
		return http.StatusBadRequest, "Diese Spalte ist voll."
	}
	g.Grid[row][col] = color
	if winCells := checkVierGewinntWin(g.Grid, row, col, color); len(winCells) >= 4 {
		g.Winner = color
		g.Winning = winCells
		g.Status = "finished"
		return 0, ""
	}
	for cIdx := 0; cIdx < 7; cIdx++ {
		if g.Grid[0][cIdx] == "" {
			g.BotFrist = time.Time{}
			if color == "red" {
				g.Current = "blue"
			} else {
				g.Current = "red"
			}
			return 0, ""
		}
	}
	g.Winner = "draw"
	g.Status = "finished"
	return 0, ""
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
		vierGewinntBotTreiben(room.VierGewinnt, time.Now())
		game := cloneVierGewinnt(room.VierGewinnt)
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
		Stufe  string `json:"stufe"` // nur bei action=bot
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.ID = kennung(r)
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
		if room.VierGewinnt != nil && !room.gameSlotFree(room.VierGewinnt.Status, room.VierGewinnt.RedID, room.VierGewinnt.BlueID) {
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
		if nr, text := vierGewinntSetzen(game, color, input.Column); nr != 0 {
			mu.Unlock()
			http.Error(w, text, nr)
			return
		}

	case "bot":
		room.botNamenNeu()
		stufe := botStufe(input.Stufe)
		if room.VierGewinnt != nil && !room.gameSlotFree(room.VierGewinnt.Status, room.VierGewinnt.RedID, room.VierGewinnt.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Vier-gewinnt-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.VierGewinnt = &VierGewinntGame{
			ID: randomCode(), RedID: input.ID, RedName: player.Name,
			BlueID: botID(1), BlueName: room.botName(1),
			Grid: newVierGewinntGrid(), Current: "red", Status: "active",
			BotID: botID(1), BotStufe: stufe,
		}
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})
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
		game.BotFrist = time.Time{}

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Vier-gewinnt-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneVierGewinnt(room.VierGewinnt)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
