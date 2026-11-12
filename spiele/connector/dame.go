package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type GenericGameSession struct {
	ID        string `json:"id"`
	GameType  string `json:"gameType"` // "tictactoe", "viergewinnt", "dame", etc.
	Player1ID string `json:"player1Id"`
	Player1   string `json:"player1"`
	Player2ID string `json:"player2Id"`
	Player2   string `json:"player2"`
	Status    string `json:"status"` // "invited", "active", "declined", "finished"
	State     any    `json:"state"`  // arbitrary game state
}

type DameGame struct {
	ID       string         `json:"id"`
	RedID    string         `json:"redId"`
	RedName  string         `json:"redName"`
	BlueID   string         `json:"blueId"`
	BlueName string         `json:"blueName"`
	Board    [][]map[string]any `json:"board"` // 8x8 cells: null or {color: "red"|"blue", king: bool}
	Current  string         `json:"current"`  // "red" or "blue"
	Status   string         `json:"status"`   // "invited", "active", "declined", "finished"
	Winner   string         `json:"winner,omitempty"`
}

func newDameBoard() [][]map[string]any {
	board := make([][]map[string]any, 8)
	for r := 0; r < 8; r++ {
		board[r] = make([]map[string]any, 8)
		for c := 0; c < 8; c++ {
			if (r+c)%2 != 0 {
				if r < 3 {
					board[r][c] = map[string]any{"color": "blue", "king": false}
				} else if r > 4 {
					board[r][c] = map[string]any{"color": "red", "king": false}
				}
			}
		}
	}
	return board
}

func dameHandler(w http.ResponseWriter, r *http.Request) {
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
		game := room.Dame
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID      string             `json:"id"`
		ToID    string             `json:"toId"`
		Board   [][]map[string]any `json:"board"`
		Current string             `json:"current"`
		Winner  string             `json:"winner"`
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
		if room.Dame != nil && (room.Dame.Status == "invited" || room.Dame.Status == "active") {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Dame-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Dame = &DameGame{
			ID:       randomCode(),
			RedID:    input.ID,
			RedName:  player.Name,
			BlueID:   input.ToID,
			BlueName: target.Name,
			Board:    newDameBoard(),
			Current:  "red",
			Status:   "invited",
		}
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Dame ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "dame",
			GameID:     room.Dame.ID,
		})

	case "accept":
		game := room.Dame
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"

	case "decline":
		game := room.Dame
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "move":
		game := room.Dame
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

		if input.Board != nil {
			game.Board = input.Board
		}
		if input.Current != "" {
			game.Current = input.Current
		}
		if input.Winner != "" {
			game.Winner = input.Winner
			game.Status = "finished"
		}

	case "reset":
		game := room.Dame
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		game.Board = newDameBoard()
		game.Current = "red"
		game.Winner = ""
		game.Status = "active"

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Dame-Aktion.", http.StatusBadRequest)
		return
	}

	result := room.Dame
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
