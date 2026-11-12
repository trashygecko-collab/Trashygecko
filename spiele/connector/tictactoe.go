package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ticTacToeWinner determines the winner of a TicTacToe board.
func ticTacToeWinner(board []string) string {
	combos := [][3]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {0, 3, 6}, {1, 4, 7}, {2, 5, 8}, {0, 4, 8}, {2, 4, 6}}
	for _, combo := range combos {
		mark := board[combo[0]]
		if mark != "" && board[combo[1]] == mark && board[combo[2]] == mark {
			return mark
		}
	}
	return ""
}

// ticTacToeHandler handles game state requests and actions.
func ticTacToeHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneTicTacToe(room.TicTacToe)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID    string `json:"id"`
		ToID  string `json:"toId"`
		Index int    `json:"index"`
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
		if room.TicTacToe != nil && !room.gameSlotFree(room.TicTacToe.Status, room.TicTacToe.XID, room.TicTacToe.OID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Tic-Tac-Toe-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.TicTacToe = &TicTacToeGame{ID: randomCode(), XID: input.ID, XName: player.Name, OID: input.ToID, OName: target.Name, Board: make([]string, 9), Current: "X", Status: "invited"}
		room.addMessage(Message{Name: player.Name, Text: "lädt dich zu Neon Tic Tac Toe ein.", Time: time.Now().Format("15:04:05"), SenderID: input.ID, ToID: input.ToID, ToName: target.Name, Whisper: true, GameInvite: true, Game: "tictactoe", GameID: room.TicTacToe.ID})
	case "accept":
		game := room.TicTacToe
		if game == nil || game.Status != "invited" || game.OID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
	case "decline":
		game := room.TicTacToe
		if game == nil || game.Status != "invited" || game.OID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"
	case "move":
		game := room.TicTacToe
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		mark := ""
		if input.ID == game.XID {
			mark = "X"
		} else if input.ID == game.OID {
			mark = "O"
		}
		if mark == "" {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if input.Index < 0 || input.Index > 8 || game.Board[input.Index] != "" {
			mu.Unlock()
			http.Error(w, "Dieses Feld ist nicht verfügbar.", http.StatusBadRequest)
			return
		}
		if game.Current != mark {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		game.Board[input.Index] = mark
		if winner := ticTacToeWinner(game.Board); winner != "" {
			game.Winner = winner
			game.Status = "finished"
		} else {
			// check draw
			full := true
			for _, f := range game.Board {
				if f == "" {
					full = false
					break
				}
			}
			if full {
				game.Winner = "draw"
				game.Status = "finished"
			} else if mark == "X" {
				game.Current = "O"
			} else {
				game.Current = "X"
			}
		}
	case "reset":
		game := room.TicTacToe
		if game == nil || game.Status != "finished" || (input.ID != game.XID && input.ID != game.OID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		game.Board = make([]string, 9)
		game.Current = "X"
		game.Winner = ""
		game.Status = "active"
	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Tic-Tac-Toe-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneTicTacToe(room.TicTacToe)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
