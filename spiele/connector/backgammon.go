package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Backgammon: 24 Punkte, Spieler 0 (Gastgeber) zieht von 23 nach 0,
// Spieler 1 von 0 nach 23. Positive Werte gehören Spieler 0.
type bgDie struct {
	ID    string `json:"id"`
	Value int    `json:"value"`
	Used  bool   `json:"used"`
}

type BackgammonGame struct {
	ID         string  `json:"id"`
	RedID      string  `json:"redId"`
	RedName    string  `json:"redName"`
	BlueID     string  `json:"blueId"`
	BlueName   string  `json:"blueName"`
	Points     []int   `json:"points"`
	Bar        [2]int  `json:"bar"`
	Off        [2]int  `json:"off"`
	Player     int     `json:"player"`
	Dice       []bgDie `json:"dice"`
	Rolled     bool    `json:"rolled"`
	Status     string  `json:"status"` // invited, active, finished
	Winner     string  `json:"winner,omitempty"`
	Multiplier int     `json:"multiplier"`
	Note       string  `json:"note,omitempty"`
}

func bgInitialPoints() []int {
	points := make([]int, 24)
	points[23] = 2
	points[12] = 5
	points[7] = 3
	points[5] = 5
	points[0] = -2
	points[11] = -5
	points[16] = -3
	points[18] = -5
	return points
}

func bgCountOn(points []int, point, player int) int {
	if player == 0 {
		if points[point] > 0 {
			return points[point]
		}
		return 0
	}
	if points[point] < 0 {
		return -points[point]
	}
	return 0
}

func bgIsHome(points []int, bar [2]int, player int) bool {
	if bar[player] > 0 {
		return false
	}
	for i := 0; i < 24; i++ {
		if bgCountOn(points, i, player) > 0 {
			if player == 0 && i > 5 {
				return false
			}
			if player == 1 && i < 18 {
				return false
			}
		}
	}
	return true
}

type bgMove struct {
	From int // -1 = von der Bar
	To   int // -1 = auswürfeln
	Die  int
}

func bgLegalMoves(points []int, bar [2]int, player, die int) []bgMove {
	moves := []bgMove{}

	add := func(from, to int) {
		if to >= 0 {
			if player == 0 && points[to] <= -2 {
				return
			}
			if player == 1 && points[to] >= 2 {
				return
			}
		}
		moves = append(moves, bgMove{From: from, To: to, Die: die})
	}

	if bar[player] > 0 {
		to := 24 - die
		if player == 1 {
			to = die - 1
		}
		add(-1, to)
		return moves
	}

	home := bgIsHome(points, bar, player)

	for from := 0; from < 24; from++ {
		if bgCountOn(points, from, player) == 0 {
			continue
		}
		to := from - die
		if player == 1 {
			to = from + die
		}

		if to >= 0 && to < 24 {
			add(from, to)
			continue
		}

		if !home {
			continue
		}

		if to == -1 || to == 24 {
			add(from, -1)
			continue
		}

		if player == 0 && to < 0 {
			farther := false
			for j := from + 1; j <= 5; j++ {
				if points[j] > 0 {
					farther = true
				}
			}
			if !farther {
				add(from, -1)
			}
		} else if player == 1 && to > 23 {
			farther := false
			for j := 18; j < from; j++ {
				if points[j] < 0 {
					farther = true
				}
			}
			if !farther {
				add(from, -1)
			}
		}
	}

	return moves
}

func bgApply(points []int, bar [2]int, off [2]int, player int, move bgMove) ([]int, [2]int, [2]int) {
	np := append([]int(nil), points...)
	nb := bar
	no := off
	sign := 1
	if player == 1 {
		sign = -1
	}
	opponent := 1 - player

	if move.From == -1 {
		nb[player]--
	} else {
		np[move.From] -= sign
	}

	if move.To == -1 {
		no[player]++
		return np, nb, no
	}

	if (player == 0 && np[move.To] == -1) || (player == 1 && np[move.To] == 1) {
		np[move.To] = 0
		nb[opponent]++
	}
	np[move.To] += sign
	return np, nb, no
}

func bgHasMove(points []int, bar [2]int, player int, dice []bgDie) bool {
	for _, die := range dice {
		if die.Used {
			continue
		}
		if len(bgLegalMoves(points, bar, player, die.Value)) > 0 {
			return true
		}
	}
	return false
}

func bgNewGame(hostID, hostName, guestID, guestName, id string) *BackgammonGame {
	return &BackgammonGame{
		ID:       id,
		RedID:    hostID,
		RedName:  hostName,
		BlueID:   guestID,
		BlueName: guestName,
		Points:   bgInitialPoints(),
		Player:   0,
		Dice:     []bgDie{},
		Status:   "invited",
	}
}

func bgClone(g *BackgammonGame) *BackgammonGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Points = append([]int(nil), g.Points...)
	c.Dice = append([]bgDie(nil), g.Dice...)
	return &c
}

func bgColorOf(g *BackgammonGame, id string) string {
	if g.RedID == id {
		return "red"
	}
	if g.BlueID == id {
		return "blue"
	}
	return ""
}

func bgPlayerOf(color string) int {
	if color == "red" {
		return 0
	}
	return 1
}

func bgName(g *BackgammonGame, player int) string {
	if player == 0 {
		return g.RedName
	}
	return g.BlueName
}

// bgNextTurn wechselt den Spieler und räumt die Würfel ab.
func bgNextTurn(g *BackgammonGame) {
	g.Player = 1 - g.Player
	g.Dice = []bgDie{}
	g.Rolled = false
	g.Note = bgName(g, g.Player) + " ist am Zug."
}

func bgFinish(g *BackgammonGame) {
	winner := g.Player
	loser := 1 - winner

	multiplier := 1
	if g.Off[loser] == 0 {
		backgammon := g.Bar[loser] > 0
		for i := 0; i < 24 && !backgammon; i++ {
			if winner == 0 && g.Points[i] < 0 && i <= 5 {
				backgammon = true
			}
			if winner == 1 && g.Points[i] > 0 && i >= 18 {
				backgammon = true
			}
		}
		if backgammon {
			multiplier = 3
		} else {
			multiplier = 2
		}
	}

	g.Status = "finished"
	g.Winner = "red"
	if winner == 1 {
		g.Winner = "blue"
	}
	g.Multiplier = multiplier
	g.Note = bgName(g, winner) + " gewinnt."
}

func backgammonHandler(w http.ResponseWriter, r *http.Request) {
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
		game := bgClone(room.Backgammon)
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
		From *int   `json:"from"`
		To   *int   `json:"to"`
		Die  int    `json:"die"`
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
		if room.Backgammon != nil && !room.gameSlotFree(room.Backgammon.Status, room.Backgammon.RedID, room.Backgammon.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Backgammon-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Backgammon = bgNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Backgammon ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "backgammon",
			GameID:     room.Backgammon.ID,
		})

	case "accept":
		game := room.Backgammon
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = game.RedName + " ist am Zug."

	case "decline":
		game := room.Backgammon
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "roll", "move", "reset":
		game := room.Backgammon
		if game == nil || (action != "reset" && game.Status != "active") {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := bgColorOf(game, input.ID)
		if color == "" {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		me := bgPlayerOf(color)

		switch action {
		case "roll":
			if game.Player != me {
				mu.Unlock()
				http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
				return
			}
			if game.Rolled {
				mu.Unlock()
				http.Error(w, "Du hast in diesem Zug schon gewürfelt.", http.StatusConflict)
				return
			}

			a := rand.Intn(6) + 1
			b := rand.Intn(6) + 1
			values := []int{a, b}
			if a == b {
				values = []int{a, a, a, a}
			}
			dice := make([]bgDie, len(values))
			for i, v := range values {
				dice[i] = bgDie{ID: "d" + strconv.Itoa(i), Value: v}
			}
			game.Dice = dice
			game.Rolled = true

			if a == b {
				game.Note = "Pasch! Vier Züge mit der " + strconv.Itoa(a) + "."
			} else {
				game.Note = "Gewürfelt: " + strconv.Itoa(a) + " und " + strconv.Itoa(b) + "."
			}

			if !bgHasMove(game.Points, game.Bar, game.Player, game.Dice) {
				game.Note = game.Note + " Kein Zug möglich."
				bgNextTurn(game)
				game.Note = bgName(game, game.Player) + " ist am Zug (kein Zug möglich)."
			}

		case "move":
			if game.Player != me {
				mu.Unlock()
				http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
				return
			}
			if !game.Rolled || len(game.Dice) == 0 {
				mu.Unlock()
				http.Error(w, "Du musst zuerst würfeln.", http.StatusConflict)
				return
			}
			if input.From == nil || input.To == nil {
				mu.Unlock()
				http.Error(w, "Es fehlen Start- oder Zielfeld.", http.StatusBadRequest)
				return
			}

			var chosen *bgMove
			for _, move := range bgLegalMoves(game.Points, game.Bar, game.Player, input.Die) {
				if move.From == *input.From && move.To == *input.To {
					m := move
					chosen = &m
					break
				}
			}
			if chosen == nil {
				mu.Unlock()
				http.Error(w, "Dieser Zug ist nicht erlaubt.", http.StatusConflict)
				return
			}

			dieOK := false
			for i := range game.Dice {
				if !game.Dice[i].Used && game.Dice[i].Value == input.Die {
					game.Dice[i].Used = true
					dieOK = true
					break
				}
			}
			if !dieOK {
				mu.Unlock()
				http.Error(w, "Dieser Würfel ist nicht verfügbar.", http.StatusConflict)
				return
			}

			game.Points, game.Bar, game.Off = bgApply(game.Points, game.Bar, game.Off, game.Player, *chosen)

			if game.Off[game.Player] == 15 {
				bgFinish(game)
				break
			}

			if !bgHasMove(game.Points, game.Bar, game.Player, game.Dice) {
				bgNextTurn(game)
				break
			}

			game.Note = "Gezogen – " + bgName(game, game.Player) + " ist weiter am Zug."

		case "reset":
			if game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
				mu.Unlock()
				http.Error(w, "Eine neue Partie ist erst nach Spielende möglich.", http.StatusConflict)
				return
			}
			fresh := bgNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
			fresh.Status = "active"
			fresh.Note = fresh.RedName + " ist am Zug."
			room.Backgammon = fresh
		}

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Backgammon-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := bgClone(room.Backgammon)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
