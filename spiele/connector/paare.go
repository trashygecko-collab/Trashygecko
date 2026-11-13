package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// PaareResult ist das Ergebnis eines Spielers: gebrauchte Zeit in Millisekunden
// und Anzahl der Fehlversuche.
type PaareResult struct {
	Time     int64 `json:"time"`
	Mistakes int   `json:"mistakes"`
}

// PaareGame ist eine Partie "Paare finden". Beide Spieler lösen dasselbe Brett
// gleichzeitig; wer schneller fertig ist, gewinnt.
type PaareGame struct {
	ID       string       `json:"id"`
	RedID    string       `json:"redId"`
	RedName  string       `json:"redName"`
	BlueID   string       `json:"blueId"`
	BlueName string       `json:"blueName"`
	Kind     string       `json:"kind"`  // "numbers", "letters", "symbols"
	Pairs    int          `json:"pairs"` // 1..12
	Status   string       `json:"status"`
	RedDone  *PaareResult `json:"redDone,omitempty"`
	BlueDone *PaareResult `json:"blueDone,omitempty"`
	Winner   string       `json:"winner,omitempty"` // "red", "blue", "draw"
}

var paareKinds = []string{"numbers", "letters", "symbols"}

func paareKindValid(kind string) bool {
	for _, known := range paareKinds {
		if kind == known {
			return true
		}
	}
	return false
}

// paareKindName ist die deutsche Bezeichnung der Zeichenart (für Meldungen).
func paareKindName(kind string) string {
	switch kind {
	case "numbers":
		return "Zahlen"
	case "letters":
		return "Buchstaben"
	case "symbols":
		return "Sonderzeichen"
	}
	return kind
}

// paareWinner ermittelt den Sieger. Bei gleicher Zeit entscheiden die Fehler.
func paareWinner(g *PaareGame) string {
	if g.RedDone == nil || g.BlueDone == nil {
		return ""
	}
	switch {
	case g.RedDone.Time < g.BlueDone.Time:
		return "red"
	case g.BlueDone.Time < g.RedDone.Time:
		return "blue"
	}
	switch {
	case g.RedDone.Mistakes < g.BlueDone.Mistakes:
		return "red"
	case g.BlueDone.Mistakes < g.RedDone.Mistakes:
		return "blue"
	}
	return "draw"
}

func paareHandler(w http.ResponseWriter, r *http.Request) {
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
		game := clonePaare(room.Paare)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID       string `json:"id"`
		ToID     string `json:"toId"`
		Kind     string `json:"kind"`
		Pairs    int    `json:"pairs"`
		Time     int64  `json:"time"`
		Mistakes int    `json:"mistakes"`
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
		if room.Paare != nil && !room.gameSlotFree(room.Paare.Status, room.Paare.RedID, room.Paare.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Paare-Partie oder Einladung.", http.StatusConflict)
			return
		}
		pairs := input.Pairs
		if pairs < 1 {
			pairs = 8
		}
		if pairs > 12 {
			pairs = 12
		}
		// Der Einladende bestimmt, welche Zeichen auf dem Brett liegen.
		// Ohne gültige Angabe wählt der Server zufällig.
		kind := strings.ToLower(strings.TrimSpace(input.Kind))
		if !paareKindValid(kind) {
			kind = paareKinds[int(time.Now().UnixNano())%len(paareKinds)]
		}
		room.Paare = &PaareGame{
			ID:       randomCode(),
			RedID:    input.ID,
			RedName:  player.Name,
			BlueID:   input.ToID,
			BlueName: target.Name,
			Kind:     kind,
			Pairs:    pairs,
			Status:   "invited",
		}
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Paare finden ein (" + strconv.Itoa(pairs) + " Paare · " + paareKindName(kind) + ").",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "paare",
			GameID:     room.Paare.ID,
		})

	case "accept":
		game := room.Paare
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"

	case "decline":
		game := room.Paare
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "done":
		game := room.Paare
		if game == nil || (game.Status != "active" && game.Status != "finished") {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine Partie.", http.StatusConflict)
			return
		}
		if input.Time <= 0 {
			input.Time = 1
		}
		if input.Mistakes < 0 {
			input.Mistakes = 0
		}
		result := &PaareResult{Time: input.Time, Mistakes: input.Mistakes}
		switch input.ID {
		case game.RedID:
			if game.RedDone != nil {
				mu.Unlock()
				http.Error(w, "Ergebnis wurde bereits gemeldet.", http.StatusConflict)
				return
			}
			game.RedDone = result
		case game.BlueID:
			if game.BlueDone != nil {
				mu.Unlock()
				http.Error(w, "Ergebnis wurde bereits gemeldet.", http.StatusConflict)
				return
			}
			game.BlueDone = result
		default:
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.RedDone != nil && game.BlueDone != nil {
			game.Winner = paareWinner(game)
			game.Status = "finished"
		}

	case "reset":
		game := room.Paare
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		game.ID = randomCode() // neuer Seed, damit beide ein frisches Brett bekommen
		game.RedDone = nil
		game.BlueDone = nil
		game.Winner = ""
		game.Status = "active"

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Paare-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := clonePaare(room.Paare)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
