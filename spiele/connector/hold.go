package main

import (
	"encoding/json"
	"math"
	"net/http"
	"strings"
	"time"
)

// Hold the Line: 5x5 Punkte. Abwechselnd eine gerade Linie zwischen zwei
// freien Punkten ziehen, die keine vorhandene Linie berührt oder kreuzt.
// Wer keinen Zug mehr hat, verliert.
const holdSize = 25

type HoldLine struct {
	A     int    `json:"a"`
	B     int    `json:"b"`
	Color string `json:"color"` // red oder blue
}

type HoldGame struct {
	ID       string     `json:"id"`
	RedID    string     `json:"redId"`
	RedName  string     `json:"redName"`
	BlueID   string     `json:"blueId"`
	BlueName string     `json:"blueName"`
	Lines    []HoldLine `json:"lines"`
	Current  string     `json:"current"`
	Status   string     `json:"status"`
	Winner   string     `json:"winner,omitempty"`
	Note     string     `json:"note,omitempty"`
}

type holdPoint struct {
	x float64
	y float64
}

func holdPointOf(index int) holdPoint {
	return holdPoint{
		x: float64(index % 5),
		y: float64(index / 5),
	}
}

func holdCross(a, b, c holdPoint) float64 {
	return (b.x-a.x)*(c.y-a.y) - (b.y-a.y)*(c.x-a.x)
}

func holdOnSegment(a, b, p holdPoint) bool {
	const eps = 0.000001
	return math.Abs(holdCross(a, b, p)) < eps &&
		p.x >= math.Min(a.x, b.x)-eps && p.x <= math.Max(a.x, b.x)+eps &&
		p.y >= math.Min(a.y, b.y)-eps && p.y <= math.Max(a.y, b.y)+eps
}

func holdSegmentsTouch(a, b, c, d holdPoint) bool {
	const eps = 0.000001
	o1 := holdCross(a, b, c)
	o2 := holdCross(a, b, d)
	o3 := holdCross(c, d, a)
	o4 := holdCross(c, d, b)

	if ((o1 > eps && o2 < -eps) || (o1 < -eps && o2 > eps)) &&
		((o3 > eps && o4 < -eps) || (o3 < -eps && o4 > eps)) {
		return true
	}

	if math.Abs(o1) < eps && holdOnSegment(a, b, c) {
		return true
	}
	if math.Abs(o2) < eps && holdOnSegment(a, b, d) {
		return true
	}
	if math.Abs(o3) < eps && holdOnSegment(c, d, a) {
		return true
	}
	if math.Abs(o4) < eps && holdOnSegment(c, d, b) {
		return true
	}
	return false
}

func holdCanDraw(lines []HoldLine, start, end int) bool {
	if start == end || start < 0 || start >= holdSize || end < 0 || end >= holdSize {
		return false
	}
	a := holdPointOf(start)
	b := holdPointOf(end)
	for _, line := range lines {
		if holdSegmentsTouch(a, b, holdPointOf(line.A), holdPointOf(line.B)) {
			return false
		}
	}
	return true
}

func holdHasMove(lines []HoldLine) bool {
	for a := 0; a < holdSize; a++ {
		for b := a + 1; b < holdSize; b++ {
			if holdCanDraw(lines, a, b) {
				return true
			}
		}
	}
	return false
}

func holdNewGame(hostID, hostName, guestID, guestName, id string) *HoldGame {
	return &HoldGame{
		ID:       id,
		RedID:    hostID,
		RedName:  hostName,
		BlueID:   guestID,
		BlueName: guestName,
		Lines:    []HoldLine{},
		Current:  "red",
		Status:   "invited",
	}
}

func holdColorOf(g *HoldGame, id string) string {
	if g.RedID == id {
		return "red"
	}
	if g.BlueID == id {
		return "blue"
	}
	return ""
}

func holdOther(color string) string {
	if color == "red" {
		return "blue"
	}
	return "red"
}

func holdName(g *HoldGame, color string) string {
	if color == "red" {
		return g.RedName
	}
	return g.BlueName
}

func cloneHold(g *HoldGame) *HoldGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Lines = append([]HoldLine(nil), g.Lines...)
	return &c
}

func holdHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneHold(room.Hold)
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
		A    int    `json:"a"`
		B    int    `json:"b"`
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
		if room.Hold != nil && !room.gameSlotFree(room.Hold.Status, room.Hold.RedID, room.Hold.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Hold-the-Line-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Hold = holdNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Hold the Line ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "hold",
			GameID:     room.Hold.ID,
		})

	case "accept":
		game := room.Hold
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = ""

	case "decline":
		game := room.Hold
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "line":
		game := room.Hold
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := holdColorOf(game, input.ID)
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
		if input.A < 0 || input.A >= holdSize || input.B < 0 || input.B >= holdSize {
			mu.Unlock()
			http.Error(w, "Dieser Punkt liegt nicht auf dem Brett.", http.StatusBadRequest)
			return
		}
		if input.A == input.B {
			mu.Unlock()
			http.Error(w, "Eine Linie braucht zwei verschiedene Punkte.", http.StatusBadRequest)
			return
		}
		if !holdCanDraw(game.Lines, input.A, input.B) {
			mu.Unlock()
			http.Error(w, "Diese Linie ist nicht erlaubt!", http.StatusConflict)
			return
		}

		game.Lines = append(game.Lines, HoldLine{A: input.A, B: input.B, Color: color})
		game.Note = player.Name + " zieht eine Linie."

		if !holdHasMove(game.Lines) {
			game.Status = "finished"
			game.Winner = color
			game.Note = holdName(game, holdOther(color)) + " hat keinen gültigen Zug mehr."
			break
		}
		game.Current = holdOther(color)

	case "reset":
		game := room.Hold
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		fresh := holdNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
		fresh.Status = "active"
		room.Hold = fresh

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Hold-the-Line-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneHold(room.Hold)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
