package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Sim: sechs Punkte, jede Linie zwischen zwei Punkten ist ein Zug. Wer ein
// Dreieck in der eigenen Farbe vervollständigt, hat verloren.
const simEdgeCount = 15

type SimGame struct {
	ID       string   `json:"id"`
	RedID    string   `json:"redId"`
	RedName  string   `json:"redName"`
	BlueID   string   `json:"blueId"`
	BlueName string   `json:"blueName"`
	Edges    []string `json:"edges"` // 15 Linien: "", "red" oder "blue"
	Current  string   `json:"current"`
	Status   string   `json:"status"` // invited, active, finished
	Winner   string   `json:"winner,omitempty"`
	Loser    string   `json:"loser,omitempty"`
	Triangle []int    `json:"triangle,omitempty"`
	Note     string   `json:"note,omitempty"`
}

func simNewGame(hostID, hostName, guestID, guestName, id string) *SimGame {
	return &SimGame{
		ID:       id,
		RedID:    hostID,
		RedName:  hostName,
		BlueID:   guestID,
		BlueName: guestName,
		Edges:    make([]string, simEdgeCount),
		Current:  "red",
		Status:   "invited",
	}
}

// simEdge liefert den Index einer Linie zwischen zwei Punkten – in derselben
// Reihenfolge, in der die Spielseite ihre Linien aufbaut.
func simEdge(a, b int) int {
	if a > b {
		a, b = b, a
	}
	index := 0
	for i := 0; i < 6; i++ {
		for j := i + 1; j < 6; j++ {
			if i == a && j == b {
				return index
			}
			index++
		}
	}
	return -1
}

// simTriangle sucht ein Dreieck, bei dem alle drei Linien diese Farbe haben.
func simTriangle(edges []string, color string) []int {
	for a := 0; a < 6; a++ {
		for b := a + 1; b < 6; b++ {
			for c := b + 1; c < 6; c++ {
				ab := simEdge(a, b)
				ac := simEdge(a, c)
				bc := simEdge(b, c)
				if ab < 0 || ac < 0 || bc < 0 {
					continue
				}
				if edges[ab] == color && edges[ac] == color && edges[bc] == color {
					return []int{ab, ac, bc}
				}
			}
		}
	}
	return nil
}

func simColorOf(g *SimGame, id string) string {
	if g.RedID == id {
		return "red"
	}
	if g.BlueID == id {
		return "blue"
	}
	return ""
}

func cloneSim(g *SimGame) *SimGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Edges = append([]string(nil), g.Edges...)
	if g.Triangle != nil {
		c.Triangle = append([]int(nil), g.Triangle...)
	}
	return &c
}

func simHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneSim(room.Sim)
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
		if room.Sim != nil && !room.gameSlotFree(room.Sim.Status, room.Sim.RedID, room.Sim.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Sim-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Sim = simNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Sim ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "sim",
			GameID:     room.Sim.ID,
		})

	case "accept":
		game := room.Sim
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = ""

	case "decline":
		game := room.Sim
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "line":
		game := room.Sim
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := simColorOf(game, input.ID)
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
		if input.Index < 0 || input.Index >= simEdgeCount {
			mu.Unlock()
			http.Error(w, "Diese Linie gibt es nicht.", http.StatusBadRequest)
			return
		}
		if game.Edges[input.Index] != "" {
			mu.Unlock()
			http.Error(w, "Diese Linie ist schon gesetzt.", http.StatusConflict)
			return
		}

		game.Edges[input.Index] = color
		if triangle := simTriangle(game.Edges, color); triangle != nil {
			// Eigenes Dreieck gebaut – der andere gewinnt.
			game.Triangle = triangle
			game.Status = "finished"
			game.Loser = color
			if color == "red" {
				game.Winner = "blue"
			} else {
				game.Winner = "red"
			}
			game.Note = player.Name + " baut ein Dreieck und verliert!"
		} else {
			if color == "red" {
				game.Current = "blue"
			} else {
				game.Current = "red"
			}
			game.Note = ""
		}

	case "reset":
		game := room.Sim
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		game.Edges = make([]string, simEdgeCount)
		game.Triangle = nil
		game.Current = "red"
		game.Winner = ""
		game.Loser = ""
		game.Note = ""
		game.Status = "active"

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Sim-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneSim(room.Sim)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
