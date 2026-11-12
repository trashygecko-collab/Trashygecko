package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Kalaha: 14 Fächer – 0..5 Rot, 6 rotes Sammelbecken, 7..12 Blau,
// 13 blaues Sammelbecken. Verteilt wird gegen den Uhrzeigersinn, also mit
// steigendem Index; das Becken des Gegners wird übersprungen.
const kalahaSlots = 14

type KalahaGame struct {
	ID       string `json:"id"`
	RedID    string `json:"redId"`
	RedName  string `json:"redName"`
	BlueID   string `json:"blueId"`
	BlueName string `json:"blueName"`
	Pits     []int  `json:"pits"`
	Current  string `json:"current"`
	Status   string `json:"status"` // invited, active, finished
	Winner   string `json:"winner,omitempty"`
	Note     string `json:"note,omitempty"`
}

func kalahaNewGame(hostID, hostName, guestID, guestName, id string) *KalahaGame {
	pits := make([]int, kalahaSlots)
	for i := 0; i < 6; i++ {
		pits[i] = 4
		pits[i+7] = 4
	}
	return &KalahaGame{
		ID:       id,
		RedID:    hostID,
		RedName:  hostName,
		BlueID:   guestID,
		BlueName: guestName,
		Pits:     pits,
		Current:  "red",
		Status:   "invited",
	}
}

func kalahaOwnStore(color string) int {
	if color == "red" {
		return 6
	}
	return 13
}

func kalahaOpponentStore(color string) int {
	if color == "red" {
		return 13
	}
	return 6
}

func kalahaOwnPit(index int, color string) bool {
	if color == "red" {
		return index >= 0 && index <= 5
	}
	return index >= 7 && index <= 12
}

func kalahaSideEmpty(pits []int, color string) bool {
	from, to := 0, 6
	if color == "blue" {
		from, to = 7, 13
	}
	for i := from; i < to; i++ {
		if pits[i] != 0 {
			return false
		}
	}
	return true
}

// kalahaPlay verteilt eine Mulde und gibt den Text für die Anzeige zurück.
func kalahaPlay(g *KalahaGame, index int) string {
	color := g.Current
	stones := g.Pits[index]
	g.Pits[index] = 0

	position := index
	for stones > 0 {
		position = (position + 1) % kalahaSlots
		if position == kalahaOpponentStore(color) {
			continue
		}
		g.Pits[position]++
		stones--
	}

	note := ""
	captured := 0
	if kalahaOwnPit(position, color) && g.Pits[position] == 1 {
		opposite := 12 - position
		if g.Pits[opposite] > 0 {
			captured = g.Pits[opposite] + g.Pits[position]
			g.Pits[kalahaOwnStore(color)] += captured
			g.Pits[position] = 0
			g.Pits[opposite] = 0
		}
	}
	if captured > 0 {
		note = kalahaName(g, color) + " fängt " + kalahaCount(captured) + "!"
	}

	// Eine leere Seite beendet die Partie, die Reste wandern in die Becken.
	if kalahaSideEmpty(g.Pits, "red") || kalahaSideEmpty(g.Pits, "blue") {
		for i := 0; i < 6; i++ {
			g.Pits[6] += g.Pits[i]
			g.Pits[i] = 0
			g.Pits[13] += g.Pits[i+7]
			g.Pits[i+7] = 0
		}
		g.Status = "finished"
		switch {
		case g.Pits[6] > g.Pits[13]:
			g.Winner = "red"
		case g.Pits[13] > g.Pits[6]:
			g.Winner = "blue"
		default:
			g.Winner = "draw"
		}
		g.Note = "Endstand – " + g.RedName + ": " + kalahaInt(g.Pits[6]) + " · " + g.BlueName + ": " + kalahaInt(g.Pits[13])
		return g.Note
	}

	if position == kalahaOwnStore(color) {
		if note != "" {
			return note + " Bonuszug!"
		}
		return "Bonuszug für " + kalahaName(g, color) + "!"
	}

	if color == "red" {
		g.Current = "blue"
	} else {
		g.Current = "red"
	}
	if note == "" {
		note = kalahaName(g, g.Current) + " ist dran."
	}
	return note
}

func kalahaName(g *KalahaGame, color string) string {
	if color == "red" {
		return g.RedName
	}
	return g.BlueName
}

func kalahaCount(n int) string {
	if n == 1 {
		return "einen Stein"
	}
	return kalahaInt(n) + " Steine"
}

func kalahaInt(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

func kalahaColorOf(g *KalahaGame, id string) string {
	if g.RedID == id {
		return "red"
	}
	if g.BlueID == id {
		return "blue"
	}
	return ""
}

func cloneKalaha(g *KalahaGame) *KalahaGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Pits = append([]int(nil), g.Pits...)
	return &c
}

func kalahaHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneKalaha(room.Kalaha)
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
		if room.Kalaha != nil && !room.gameSlotFree(room.Kalaha.Status, room.Kalaha.RedID, room.Kalaha.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Kalaha-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Kalaha = kalahaNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Kalaha ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "kalaha",
			GameID:     room.Kalaha.ID,
		})

	case "accept":
		game := room.Kalaha
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = ""

	case "decline":
		game := room.Kalaha
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "sow":
		game := room.Kalaha
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := kalahaColorOf(game, input.ID)
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
		if input.Index < 0 || input.Index >= kalahaSlots {
			mu.Unlock()
			http.Error(w, "Diese Mulde gibt es nicht.", http.StatusBadRequest)
			return
		}
		if !kalahaOwnPit(input.Index, color) {
			mu.Unlock()
			http.Error(w, "Das ist eine Mulde des anderen Spielers.", http.StatusBadRequest)
			return
		}
		if game.Pits[input.Index] == 0 {
			mu.Unlock()
			http.Error(w, "Diese Mulde ist leer.", http.StatusConflict)
			return
		}
		game.Note = kalahaPlay(game, input.Index)

	case "reset":
		game := room.Kalaha
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		fresh := kalahaNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
		fresh.Status = "active"
		room.Kalaha = fresh

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Kalaha-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneKalaha(room.Kalaha)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
