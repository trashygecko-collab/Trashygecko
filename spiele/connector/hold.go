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

	Bots     []string  `json:"bots,omitempty"`
	BotStufe string    `json:"botStufe,omitempty"`
	BotFrist time.Time `json:"-"`
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
		holdBotTreiben(room, time.Now())
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
		ID    string `json:"id"`
		ToID  string `json:"toId"`
		A     int    `json:"a"`
		B     int    `json:"b"`
		Stufe string `json:"stufe"`
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
	case "bot":
		room.botNamenNeu()
		if room.Hold != nil && !room.gameSlotFree(room.Hold.Status, room.Hold.RedID, room.Hold.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Hold-the-Line-Partie oder Einladung.", http.StatusConflict)
			return
		}
		stufe := botStufe(input.Stufe)
		game := holdNewGame(input.ID, player.Name, botID(1), room.botName(1), randomCode())
		game.Bots, game.BotStufe = []string{botID(1)}, stufe
		game.Status = "active"
		game.Note = player.Name + " spielt gegen " + room.botName(1) +
			" (" + botStufeText(stufe) + ")."
		game.BotFrist = time.Now().Add(botWartezeit(stufe))
		room.Hold = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Hold the Line gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     uhrzeit(time.Now()),
			SenderID: input.ID,
		})

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
			Time:       uhrzeit(time.Now()),
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

	case "line", "reset":
		if status, meldung := holdAktion(room, input.ID, action, input.A, input.B); meldung != "" {
			mu.Unlock()
			http.Error(w, meldung, status)
			return
		}

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

/* ---------------------------------------------------------------------------
   Gemeinsame Prüfung für Menschen und Bots
   ------------------------------------------------------------------------- */

// holdAktion zieht eine Linie oder startet eine neue Partie – für Menschen und
// Bots dieselbe Prüfung.
func holdAktion(room *Room, id, aktion string, a, b int) (int, string) {
	game := room.Hold
	if game == nil || (aktion != "reset" && game.Status != "active") {
		return http.StatusConflict, "Es läuft gerade keine aktive Partie."
	}
	color := holdColorOf(game, id)
	if color == "" {
		return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
	}

	switch aktion {
	case "line":
		if game.Current != color {
			return http.StatusConflict, "Der andere Spieler ist an der Reihe."
		}
		if a < 0 || a >= holdSize || b < 0 || b >= holdSize {
			return http.StatusBadRequest, "Dieser Punkt liegt nicht auf dem Brett."
		}
		if a == b {
			return http.StatusBadRequest, "Eine Linie braucht zwei verschiedene Punkte."
		}
		if !holdCanDraw(game.Lines, a, b) {
			return http.StatusConflict, "Diese Linie ist nicht erlaubt!"
		}

		game.Lines = append(game.Lines, HoldLine{A: a, B: b, Color: color})
		game.Note = holdName(game, color) + " zieht eine Linie."

		if !holdHasMove(game.Lines) {
			game.Status = "finished"
			game.Winner = color
			game.Note = holdName(game, holdOther(color)) + " hat keinen gültigen Zug mehr."
			return 0, ""
		}
		game.Current = holdOther(color)

	case "reset":
		if game.Status != "finished" || (id != game.RedID && id != game.BlueID) {
			return http.StatusConflict, "Eine neue Runde ist erst nach Spielende möglich."
		}
		fresh := holdNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
		fresh.Bots, fresh.BotStufe = game.Bots, game.BotStufe
		fresh.Status = "active"
		room.Hold = fresh

	default:
		return http.StatusBadRequest, "Unbekannte Hold-the-Line-Aktion."
	}

	return 0, ""
}
