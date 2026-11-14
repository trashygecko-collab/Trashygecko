package main

import (
	"encoding/json"
	"errors"
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
	ID       string             `json:"id"`
	RedID    string             `json:"redId"`
	RedName  string             `json:"redName"`
	BlueID   string             `json:"blueId"`
	BlueName string             `json:"blueName"`
	Board    [][]map[string]any `json:"board"`   // 8x8 cells: null or {color: "red"|"blue", king: bool}
	Current  string             `json:"current"` // "red" or "blue"
	Status   string             `json:"status"`  // "invited", "active", "declined", "finished"
	Winner   string             `json:"winner,omitempty"`
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

func dameColor(cell map[string]any) string {
	if cell == nil {
		return ""
	}
	color, _ := cell["color"].(string)
	if color != "red" && color != "blue" {
		return ""
	}
	return color
}

func dameIsKing(cell map[string]any) bool {
	if cell == nil {
		return false
	}
	king, _ := cell["king"].(bool)
	return king
}

func dameCellValid(cell map[string]any) bool {
	if cell == nil {
		return true
	}
	if _, ok := cell["color"].(string); !ok {
		return false
	}
	if dameColor(cell) == "" {
		return false
	}
	_, ok := cell["king"].(bool)
	return ok
}

func otherDameColor(color string) string {
	if color == "red" {
		return "blue"
	}
	return "red"
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// validateDameBoard prüft, ob next ein legaler Zug von color auf prev ist.
// Das Brett bleibt damit die Wahrheit: ein Client kann keine Figuren
// verschieben, hinzufügen oder sich selbst zum Sieger erklären.
func validateDameBoard(prev, next [][]map[string]any, color string) error {
	if len(prev) != 8 || len(next) != 8 {
		return errors.New("ungültiges Brett")
	}
	for r := 0; r < 8; r++ {
		if len(prev[r]) != 8 || len(next[r]) != 8 {
			return errors.New("ungültiges Brett")
		}
		for c := 0; c < 8; c++ {
			if !dameCellValid(prev[r][c]) || !dameCellValid(next[r][c]) {
				return errors.New("ungültiges Brett")
			}
		}
	}

	type square struct{ r, c int }
	var from, to *square
	balance := 0 // eigene Figuren: prev minus next
	oppPrev, oppNext := 0, 0

	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			prevColor := dameColor(prev[r][c])
			nextColor := dameColor(next[r][c])

			if prevColor == color {
				balance++
			}
			if nextColor == color {
				balance--
			}
			if prevColor != "" && prevColor != color {
				oppPrev++
			}
			if nextColor != "" && nextColor != color {
				oppNext++
			}

			if prevColor == color && nextColor == "" {
				if from != nil {
					return errors.New("nur eine Figur darf ziehen")
				}
				from = &square{r, c}
			}
			if prevColor == "" && nextColor == color {
				if to != nil {
					return errors.New("nur eine Figur darf ziehen")
				}
				to = &square{r, c}
			}
			// Figuren, die stehen bleiben, dürfen nicht verändert werden
			// (auch eine Dame darf nicht wieder zur einfachen Figur werden).
			if prevColor != "" && prevColor == nextColor {
				if dameIsKing(prev[r][c]) != dameIsKing(next[r][c]) {
					return errors.New("Figuren dürfen nicht verändert werden")
				}
			}
		}
	}

	if balance != 0 {
		return errors.New("Figurenanzahl stimmt nicht")
	}
	if from == nil || to == nil {
		return errors.New("kein Zug erkannt")
	}

	dr, dc := to.r-from.r, to.c-from.c
	if dr == 0 || absInt(dr) != absInt(dc) {
		return errors.New("Züge sind nur diagonal erlaubt")
	}
	distance := absInt(dr)
	stepR, stepC := 1, 1
	if dr < 0 {
		stepR = -1
	}
	if dc < 0 {
		stepC = -1
	}
	king := dameIsKing(prev[from.r][from.c])

	captures := 0
	for i := 1; i < distance; i++ {
		r, c := from.r+stepR*i, from.c+stepC*i
		cell := prev[r][c]
		if cell == nil {
			continue // nur eine Dame darf über leere Felder gleiten
		}
		if dameColor(cell) == color {
			return errors.New("eigene Figur im Weg")
		}
		if next[r][c] != nil {
			return errors.New("geschlagene Figur wurde nicht entfernt")
		}
		captures++
	}

	if captures != oppPrev-oppNext {
		return errors.New("Schlagzug stimmt nicht")
	}
	if !king {
		if distance > 2 {
			return errors.New("diese Figur darf nicht so weit ziehen")
		}
		if distance == 2 && captures != 1 {
			return errors.New("ungültiger Sprung")
		}
	}
	return nil
}

// dameHasAnyMove meldet, ob color überhaupt noch ziehen kann.
func dameHasAnyMove(board [][]map[string]any, color string) bool {
	if len(board) != 8 {
		return false
	}
	for r := 0; r < 8; r++ {
		if len(board[r]) != 8 {
			return false
		}
		for c := 0; c < 8; c++ {
			if dameColor(board[r][c]) != color {
				continue
			}
			king := dameIsKing(board[r][c])
			maxStep := 1
			if king {
				maxStep = 7
			}
			for _, d := range [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
				for step := 1; step <= maxStep; step++ {
					nr, nc := r+d[0]*step, c+d[1]*step
					if nr < 0 || nr > 7 || nc < 0 || nc > 7 {
						break
					}
					cell := board[nr][nc]
					if cell == nil {
						// Ein Schritt auf ein freies Feld ist immer ein Zug.
						return true
					}
					if dameColor(cell) == color {
						break
					}
					// Gegnerische Figur: schlagen, wenn dahinter ein Feld frei ist.
					lr, lc := nr+d[0], nc+d[1]
					if lr >= 0 && lr <= 7 && lc >= 0 && lc <= 7 && board[lr][lc] == nil {
						return true
					}
					break
				}
			}
		}
	}
	return false
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
		game := cloneDame(room.Dame)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID    string             `json:"id"`
		ToID  string             `json:"toId"`
		Board [][]map[string]any `json:"board"`
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
		if room.Dame != nil && !room.gameSlotFree(room.Dame.Status, room.Dame.RedID, room.Dame.BlueID) {
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
		switch input.ID {
		case game.RedID:
			color = "red"
		case game.BlueID:
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
		if input.Board == nil {
			mu.Unlock()
			http.Error(w, "Kein Spielbrett übermittelt.", http.StatusBadRequest)
			return
		}
		if err := validateDameBoard(game.Board, input.Board, color); err != nil {
			mu.Unlock()
			http.Error(w, "Ungültiger Zug: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Der Server entscheidet, wer als Nächstes zieht und wer gewonnen hat.
		game.Board = input.Board
		game.Current = otherDameColor(color)
		if !dameHasAnyMove(game.Board, game.Current) {
			game.Winner = color
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

	room.LastActive = time.Now()
	result := cloneDame(room.Dame)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
