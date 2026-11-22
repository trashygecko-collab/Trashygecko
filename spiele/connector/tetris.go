package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Tetris: jeder spielt auf seinem Gerät sein eigenes Feld. Beide bekommen
// dieselbe Steinfolge (gemeinsamer Seed). Geräumte Reihen schicken Müll zum
// Gegner: 2 Reihen -> 1, 3 -> 2, 4 -> 4. Die Felder selbst werden nicht
// übertragen – nur Linien, Leben und der offene Müll.
type TetrisGame struct {
	ID        string `json:"id"`
	RedID     string `json:"redId"`
	RedName   string `json:"redName"`
	BlueID    string `json:"blueId"`
	BlueName  string `json:"blueName"`
	Status    string `json:"status"` // invited, active, finished, declined
	Seed      int64  `json:"seed"`
	RedLines  int    `json:"redLines"`
	BlueLines int    `json:"blueLines"`
	RedAlive  bool   `json:"redAlive"`
	BlueAlive bool   `json:"blueAlive"`
	// Offener Müll, den der jeweilige Spieler noch abholen muss.
	RedGarbage  int    `json:"redGarbage"`
	BlueGarbage int    `json:"blueGarbage"`
	Winner      string `json:"winner,omitempty"`
	Note        string `json:"note,omitempty"`
}

func tetrisNewGame(hostID, hostName, guestID, guestName, id string) *TetrisGame {
	return &TetrisGame{
		ID:       id,
		RedID:    hostID,
		RedName:  hostName,
		BlueID:   guestID,
		BlueName: guestName,
		Status:   "invited",
	}
}

func cloneTetris(g *TetrisGame) *TetrisGame {
	if g == nil {
		return nil
	}
	c := *g
	return &c
}

// tetrisSeite sagt, ob die Kennung rot oder blau spielt.
func tetrisSeite(g *TetrisGame, id string) string {
	if g.RedID == id {
		return "red"
	}
	if g.BlueID == id {
		return "blue"
	}
	return ""
}

func tetrisBegrenzen(wert, min, max int) int {
	if wert < min {
		return min
	}
	if wert > max {
		return max
	}
	return wert
}

func tetrisHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneTetris(room.Tetris)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID      string `json:"id"`
		ToID    string `json:"toId"`
		Lines   int    `json:"lines"`
		Alive   *bool  `json:"alive"`
		Garbage int    `json:"garbage"`
		Taken   int    `json:"taken"`
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
		if room.Tetris != nil && !room.gameSlotFree(room.Tetris.Status, room.Tetris.RedID, room.Tetris.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Tetris-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Tetris = tetrisNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Tetris ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "tetris",
			GameID:     room.Tetris.ID,
		})

	case "accept":
		game := room.Tetris
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Seed = rand.Int63n(1 << 30)
		game.RedAlive = true
		game.BlueAlive = true
		game.RedLines = 0
		game.BlueLines = 0
		game.RedGarbage = 0
		game.BlueGarbage = 0
		game.Winner = ""
		game.Note = ""

	case "decline":
		game := room.Tetris
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "sync":
		game := room.Tetris
		if game == nil {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine Tetris-Partie.", http.StatusConflict)
			return
		}
		seite := tetrisSeite(game, input.ID)
		if seite == "" {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Status != "active" {
			mu.Unlock()
			jsonResponse(w, map[string]any{"game": cloneTetris(game)})
			return
		}

		muell := tetrisBegrenzen(input.Garbage, 0, 20)
		abgeholt := tetrisBegrenzen(input.Taken, 0, 40)
		if seite == "red" {
			game.RedLines = tetrisBegrenzen(input.Lines, 0, 9999)
			game.RedGarbage = tetrisBegrenzen(game.RedGarbage-abgeholt, 0, 9999)
			game.BlueGarbage = tetrisBegrenzen(game.BlueGarbage+muell, 0, 999)
			if input.Alive != nil && !*input.Alive && game.RedAlive {
				game.RedAlive = false
				game.Status = "finished"
				game.Winner = "blue"
				game.Note = player.Name + " hat oben keinen Platz mehr."
			}
		} else {
			game.BlueLines = tetrisBegrenzen(input.Lines, 0, 9999)
			game.BlueGarbage = tetrisBegrenzen(game.BlueGarbage-abgeholt, 0, 9999)
			game.RedGarbage = tetrisBegrenzen(game.RedGarbage+muell, 0, 999)
			if input.Alive != nil && !*input.Alive && game.BlueAlive {
				game.BlueAlive = false
				game.Status = "finished"
				game.Winner = "red"
				game.Note = player.Name + " hat oben keinen Platz mehr."
			}
		}

	case "reset":
		game := room.Tetris
		if game == nil || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		if game.Status != "finished" && game.Status != "declined" {
			mu.Unlock()
			http.Error(w, "Die laufende Partie läuft noch.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Seed = rand.Int63n(1 << 30)
		game.RedAlive = true
		game.BlueAlive = true
		game.RedLines = 0
		game.BlueLines = 0
		game.RedGarbage = 0
		game.BlueGarbage = 0
		game.Winner = ""
		game.Note = ""

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Tetris-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneTetris(room.Tetris)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
