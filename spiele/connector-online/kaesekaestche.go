package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Käsekästchen: 5 Spalten × 6 Reihen Kästchen, gespielt wird auf den Linien.
const (
	kaeseCols = 5
	kaeseRows = 6
)

// KaeseGame ist der Spielstand einer Partie. Linien und Kästchen tragen ""
// (noch frei), "red" oder "blue".
type KaeseGame struct {
	ID        string   `json:"id"`
	RedID     string   `json:"redId"`
	RedName   string   `json:"redName"`
	BlueID    string   `json:"blueId"`
	BlueName  string   `json:"blueName"`
	Cols      int      `json:"cols"`
	Rows      int      `json:"rows"`
	H         []string `json:"h"`     // (Rows+1)*Cols waagerechte Linien
	V         []string `json:"v"`     // Rows*(Cols+1) senkrechte Linien
	Boxes     []string `json:"boxes"` // Rows*Cols geschlossene Kästchen
	RedScore  int      `json:"redScore"`
	BlueScore int      `json:"blueScore"`
	Current   string   `json:"current"`
	Status    string   `json:"status"` // invited, active, finished
	Winner    string   `json:"winner,omitempty"`
	Note      string   `json:"note,omitempty"`
}

func kaeseNewGame(hostID, hostName, guestID, guestName, id string) *KaeseGame {
	return &KaeseGame{
		ID:       id,
		RedID:    hostID,
		RedName:  hostName,
		BlueID:   guestID,
		BlueName: guestName,
		Cols:     kaeseCols,
		Rows:     kaeseRows,
		H:        make([]string, (kaeseRows+1)*kaeseCols),
		V:        make([]string, kaeseRows*(kaeseCols+1)),
		Boxes:    make([]string, kaeseRows*kaeseCols),
		Current:  "red",
		Status:   "invited",
		RedScore: 0,
	}
}

func (g *KaeseGame) hIndex(row, col int) int   { return row*g.Cols + col }
func (g *KaeseGame) vIndex(row, col int) int   { return row*(g.Cols+1) + col }
func (g *KaeseGame) boxIndex(row, col int) int { return row*g.Cols + col }

func (g *KaeseGame) boxDone(row, col int) bool {
	return g.Boxes[g.boxIndex(row, col)] != ""
}

// boxComplete meldet, ob alle vier Linien um ein Kästchen liegen.
func (g *KaeseGame) boxComplete(row, col int) bool {
	return g.H[g.hIndex(row, col)] != "" &&
		g.H[g.hIndex(row+1, col)] != "" &&
		g.V[g.vIndex(row, col)] != "" &&
		g.V[g.vIndex(row, col+1)] != ""
}

// kaesePlayLine zieht eine Linie für den Spieler, der am Zug ist, und gibt die
// Anzahl der dabei geschlossenen Kästchen zurück.
func kaesePlayLine(g *KaeseGame, dir string, row, col int) int {
	color := g.Current
	if dir == "h" {
		g.H[g.hIndex(row, col)] = color
	} else {
		g.V[g.vIndex(row, col)] = color
	}

	completed := 0
	closeBox := func(r, c int) {
		if r < 0 || r >= g.Rows || c < 0 || c >= g.Cols {
			return
		}
		if g.boxDone(r, c) || !g.boxComplete(r, c) {
			return
		}
		g.Boxes[g.boxIndex(r, c)] = color
		completed++
	}
	if dir == "h" {
		closeBox(row-1, col)
		closeBox(row, col)
	} else {
		closeBox(row, col-1)
		closeBox(row, col)
	}

	if color == "red" {
		g.RedScore += completed
	} else {
		g.BlueScore += completed
	}
	return completed
}

func kaeseColorOf(g *KaeseGame, id string) string {
	if g.RedID == id {
		return "red"
	}
	if g.BlueID == id {
		return "blue"
	}
	return ""
}

func cloneKaese(g *KaeseGame) *KaeseGame {
	if g == nil {
		return nil
	}
	c := *g
	c.H = append([]string(nil), g.H...)
	c.V = append([]string(nil), g.V...)
	c.Boxes = append([]string(nil), g.Boxes...)
	return &c
}

// kaesekaestcheHandler verwaltet Einladung und Züge einer Partie.
func kaesekaestcheHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneKaese(room.Kaese)
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
		Dir  string `json:"dir"`
		Row  int    `json:"row"`
		Col  int    `json:"col"`
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
		if room.Kaese != nil && !room.gameSlotFree(room.Kaese.Status, room.Kaese.RedID, room.Kaese.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Käsekästchen-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Kaese = kaeseNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Käsekästchen ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "kaesekaestche",
			GameID:     room.Kaese.ID,
		})

	case "accept":
		game := room.Kaese
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = ""

	case "decline":
		game := room.Kaese
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "line":
		game := room.Kaese
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := kaeseColorOf(game, input.ID)
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
		dir := strings.ToLower(strings.TrimSpace(input.Dir))
		switch dir {
		case "h":
			if input.Row < 0 || input.Row > game.Rows || input.Col < 0 || input.Col >= game.Cols {
				mu.Unlock()
				http.Error(w, "Diese Linie gibt es nicht.", http.StatusBadRequest)
				return
			}
			if game.H[game.hIndex(input.Row, input.Col)] != "" {
				mu.Unlock()
				http.Error(w, "Diese Linie ist schon gezogen.", http.StatusConflict)
				return
			}
		case "v":
			if input.Row < 0 || input.Row >= game.Rows || input.Col < 0 || input.Col > game.Cols {
				mu.Unlock()
				http.Error(w, "Diese Linie gibt es nicht.", http.StatusBadRequest)
				return
			}
			if game.V[game.vIndex(input.Row, input.Col)] != "" {
				mu.Unlock()
				http.Error(w, "Diese Linie ist schon gezogen.", http.StatusConflict)
				return
			}
		default:
			mu.Unlock()
			http.Error(w, "Unbekannte Linie.", http.StatusBadRequest)
			return
		}

		completed := kaesePlayLine(game, dir, input.Row, input.Col)
		if completed > 0 {
			// Ein geschlossenes Kästchen bringt einen weiteren Zug.
			game.Note = player.Name + " schließt " + kaesePlural(completed) + " und darf noch einmal."
		} else {
			if game.Current == "red" {
				game.Current = "blue"
			} else {
				game.Current = "red"
			}
			game.Note = "Weiter geht's!"
		}
		if game.RedScore+game.BlueScore >= game.Rows*game.Cols {
			game.Status = "finished"
			switch {
			case game.RedScore > game.BlueScore:
				game.Winner = game.RedID
			case game.BlueScore > game.RedScore:
				game.Winner = game.BlueID
			default:
				game.Winner = "draw"
			}
			game.Note = "Endstand: " + game.RedName + " " + kaeseScore(game.RedScore) + " : " + kaeseScore(game.BlueScore) + " " + game.BlueName
		}

	case "reset":
		game := room.Kaese
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		game.H = make([]string, (game.Rows+1)*game.Cols)
		game.V = make([]string, game.Rows*(game.Cols+1))
		game.Boxes = make([]string, game.Rows*game.Cols)
		game.RedScore = 0
		game.BlueScore = 0
		game.Current = "red"
		game.Winner = ""
		game.Note = ""
		game.Status = "active"

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Käsekästchen-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneKaese(room.Kaese)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

func kaeseScore(n int) string {
	return strconv.Itoa(n)
}

func kaesePlural(n int) string {
	if n == 1 {
		return "ein Kästchen"
	}
	return strconv.Itoa(n) + " Kästchen"
}
