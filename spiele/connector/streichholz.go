package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// 21 Streichhölzer, abwechselnd 1 bis 3 nehmen – wer das letzte nimmt, verliert.
const streichholzTotal = 21

type StreichholzGame struct {
	ID        string   `json:"id"`
	RedID     string   `json:"redId"`
	RedName   string   `json:"redName"`
	BlueID    string   `json:"blueId"`
	BlueName  string   `json:"blueName"`
	Total     int      `json:"total"`
	Remaining int      `json:"remaining"`
	Taken     []string `json:"taken"` // pro Holz: "", "red" oder "blue"
	Current   string   `json:"current"`
	Status    string   `json:"status"` // invited, active, finished
	Loser     string   `json:"loser,omitempty"`
	Winner    string   `json:"winner,omitempty"`
	Note      string   `json:"note,omitempty"`
	// Bot: Kennung, Stufe und Ziehzeit (nicht im JSON).
	BotID    string    `json:"botId,omitempty"`
	BotStufe string    `json:"-"`
	BotFrist time.Time `json:"-"`
}

func streichholzNewGame(hostID, hostName, guestID, guestName, id string) *StreichholzGame {
	return &StreichholzGame{
		ID:        id,
		RedID:     hostID,
		RedName:   hostName,
		BlueID:    guestID,
		BlueName:  guestName,
		Total:     streichholzTotal,
		Remaining: streichholzTotal,
		Taken:     make([]string, streichholzTotal),
		Current:   "red",
		Status:    "invited",
	}
}

func streichholzName(g *StreichholzGame, color string) string {
	if color == "red" {
		return g.RedName
	}
	return g.BlueName
}

func streichholzColorOf(g *StreichholzGame, id string) string {
	if g.RedID == id {
		return "red"
	}
	if g.BlueID == id {
		return "blue"
	}
	return ""
}

func cloneStreichholz(g *StreichholzGame) *StreichholzGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Taken = append([]string(nil), g.Taken...)
	return &c
}

// streichholzNehmen prüft einen Zug und führt ihn aus – für Mensch und Bot
// derselbe Weg. Rückgabe: HTTP-Nummer und Text (0 = genommen).
func streichholzNehmen(g *StreichholzGame, color string, amount int) (int, string) {
	if g == nil || g.Status != "active" {
		return http.StatusConflict, "Es läuft gerade keine aktive Partie."
	}
	if g.Current != color {
		return http.StatusConflict, "Der andere Spieler ist an der Reihe."
	}
	if amount < 1 || amount > 3 {
		return http.StatusBadRequest, "Es dürfen nur 1, 2 oder 3 Hölzer sein."
	}
	if amount > g.Remaining {
		return http.StatusConflict, "So viele Hölzer liegen nicht mehr."
	}
	for i := g.Remaining - amount; i < g.Remaining; i++ {
		g.Taken[i] = color
	}
	g.Remaining -= amount
	name := g.RedName
	if color == "blue" {
		name = g.BlueName
	}
	if g.Remaining == 0 {
		g.Status = "finished"
		g.Loser = color
		if color == "red" {
			g.Winner = "blue"
		} else {
			g.Winner = "red"
		}
		g.Note = name + " nimmt das letzte Streichholz und verliert."
		return 0, ""
	}
	if color == "red" {
		g.Current = "blue"
	} else {
		g.Current = "red"
	}
	g.Note = name + " nimmt " + streichholzAmount(amount) + "."
	g.BotFrist = time.Time{}
	return 0, ""
}

func streichholzHandler(w http.ResponseWriter, r *http.Request) {
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
		streichholzBotTreiben(room.Streichholz, time.Now())
		game := cloneStreichholz(room.Streichholz)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID     string `json:"id"`
		ToID   string `json:"toId"`
		Amount int    `json:"amount"`
		Stufe  string `json:"stufe"` // nur bei action=bot
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
		if room.Streichholz != nil && !room.gameSlotFree(room.Streichholz.Status, room.Streichholz.RedID, room.Streichholz.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Streichholz-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Streichholz = streichholzNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zum Streichholzspiel ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "streichholzspiel",
			GameID:     room.Streichholz.ID,
		})

	case "accept":
		game := room.Streichholz
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = ""

	case "decline":
		game := room.Streichholz
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "take":
		game := room.Streichholz
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := streichholzColorOf(game, input.ID)
		if color == "" {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if nr, text := streichholzNehmen(game, color, input.Amount); nr != 0 {
			mu.Unlock()
			http.Error(w, text, nr)
			return
		}

	case "bot":
		room.botNamenNeu()
		stufe := botStufe(input.Stufe)
		if room.Streichholz != nil && !room.gameSlotFree(room.Streichholz.Status, room.Streichholz.RedID, room.Streichholz.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Streichholz-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Streichholz = streichholzNewGame(input.ID, player.Name, botID(1), room.botName(1), randomCode())
		room.Streichholz.Status = "active"
		room.Streichholz.BotID = botID(1)
		room.Streichholz.BotStufe = stufe
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})
	case "reset":
		game := room.Streichholz
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		fresh := streichholzNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
		fresh.Status = "active"
		fresh.BotID = game.BotID
		fresh.BotStufe = game.BotStufe
		room.Streichholz = fresh

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Streichholz-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneStreichholz(room.Streichholz)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

func streichholzAmount(n int) string {
	if n == 1 {
		return "ein Holz"
	}
	return string(rune('0'+n)) + " Hölzer"
}
