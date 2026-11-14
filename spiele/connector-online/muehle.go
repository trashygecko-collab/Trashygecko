package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Mühle: 24 Punkte in drei Ringen, je 9 Steine setzen, dann ziehen.
// Die Nummerierung entspricht der Spielseite: Ring 0 außen (0..7),
// Ring 1 in der Mitte (8..15), Ring 2 innen (16..23), jeweils im Uhrzeigersinn.
var (
	muehleMills     [][3]int
	muehleNeighbors [][]int
)

func init() {
	for ring := 0; ring < 3; ring++ {
		offset := ring * 8
		for i := 0; i < 8; i += 2 {
			muehleMills = append(muehleMills, [3]int{offset + i, offset + i + 1, offset + (i+2)%8})
		}
	}
	for _, i := range []int{1, 3, 5, 7} {
		muehleMills = append(muehleMills, [3]int{i, 8 + i, 16 + i})
	}

	muehleNeighbors = make([][]int, 24)
	addEdge := func(a, b int) {
		muehleNeighbors[a] = append(muehleNeighbors[a], b)
		muehleNeighbors[b] = append(muehleNeighbors[b], a)
	}
	for ring := 0; ring < 3; ring++ {
		offset := ring * 8
		for i := 0; i < 8; i++ {
			addEdge(offset+i, offset+(i+1)%8)
		}
	}
	for _, i := range []int{1, 3, 5, 7} {
		addEdge(i, 8+i)
		addEdge(8+i, 16+i)
	}
}

type MuehleGame struct {
	ID             string   `json:"id"`
	RedID          string   `json:"redId"`
	RedName        string   `json:"redName"`
	BlueID         string   `json:"blueId"`
	BlueName       string   `json:"blueName"`
	State          []string `json:"state"` // 24 Felder: "", "red" oder "blue"
	Current        string   `json:"current"`
	PlacedRed      int      `json:"placedRed"`
	PlacedBlue     int      `json:"placedBlue"`
	Phase          string   `json:"phase"` // placing, moving
	CapturePending bool     `json:"capturePending"`
	Status         string   `json:"status"` // invited, active, finished
	Winner         string   `json:"winner,omitempty"`
	Note           string   `json:"note,omitempty"`
}

func muehleNewGame(hostID, hostName, guestID, guestName, id string) *MuehleGame {
	return &MuehleGame{
		ID:       id,
		RedID:    hostID,
		RedName:  hostName,
		BlueID:   guestID,
		BlueName: guestName,
		State:    make([]string, 24),
		Current:  "red",
		Phase:    "placing",
		Status:   "invited",
	}
}

func muehleColorOf(g *MuehleGame, id string) string {
	if g.RedID == id {
		return "red"
	}
	if g.BlueID == id {
		return "blue"
	}
	return ""
}

func muehleOther(color string) string {
	if color == "red" {
		return "blue"
	}
	return "red"
}

func muehleName(g *MuehleGame, color string) string {
	if color == "red" {
		return g.RedName
	}
	return g.BlueName
}

func cloneMuehle(g *MuehleGame) *MuehleGame {
	if g == nil {
		return nil
	}
	c := *g
	c.State = append([]string(nil), g.State...)
	return &c
}

func muehleCount(g *MuehleGame, color string) int {
	count := 0
	for _, field := range g.State {
		if field == color {
			count++
		}
	}
	return count
}

func muehleIsMill(g *MuehleGame, index int, color string) bool {
	for _, mill := range muehleMills {
		if mill[0] != index && mill[1] != index && mill[2] != index {
			continue
		}
		if g.State[mill[0]] == color && g.State[mill[1]] == color && g.State[mill[2]] == color {
			return true
		}
	}
	return false
}

// muehleCanRemove: Steine in einer Mühle sind geschützt, solange der Gegner
// noch einen Stein außerhalb einer Mühle hat.
func muehleCanRemove(g *MuehleGame, index int, color string) bool {
	opponent := muehleOther(color)
	if g.State[index] != opponent {
		return false
	}
	outside := 0
	for i, field := range g.State {
		if field == opponent && !muehleIsMill(g, i, opponent) {
			outside++
		}
	}
	if outside == 0 {
		return true
	}
	return !muehleIsMill(g, index, opponent)
}

func muehleHasMove(g *MuehleGame, color string) bool {
	if muehleCount(g, color) == 3 {
		for _, field := range g.State {
			if field == "" {
				return true
			}
		}
		return false
	}
	for i, field := range g.State {
		if field != color {
			continue
		}
		for _, n := range muehleNeighbors[i] {
			if g.State[n] == "" {
				return true
			}
		}
	}
	return false
}

func muehleFinish(g *MuehleGame, winner string) {
	g.Status = "finished"
	g.Winner = winner
	g.CapturePending = false
	g.Note = muehleName(g, winner) + " hat gewonnen."
}

// muehleEndTurn wechselt den Zug und prüft, ob der neue Spieler ziehen kann.
func muehleEndTurn(g *MuehleGame, color string) {
	next := muehleOther(color)
	if g.Phase == "moving" && !muehleHasMove(g, next) {
		muehleFinish(g, color)
		return
	}
	g.Current = next
}

func muehleHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneMuehle(room.Muehle)
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
		From  int    `json:"from"`
		To    int    `json:"to"`
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
		if room.Muehle != nil && !room.gameSlotFree(room.Muehle.Status, room.Muehle.RedID, room.Muehle.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Mühle-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Muehle = muehleNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Mühle ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "muehle",
			GameID:     room.Muehle.ID,
		})

	case "accept":
		game := room.Muehle
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = ""

	case "decline":
		game := room.Muehle
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "place", "move", "capture":
		game := room.Muehle
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := muehleColorOf(game, input.ID)
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

		switch action {
		case "place":
			if game.CapturePending {
				mu.Unlock()
				http.Error(w, "Du musst zuerst einen gegnerischen Stein entfernen.", http.StatusConflict)
				return
			}
			if game.Phase != "placing" {
				mu.Unlock()
				http.Error(w, "Alle Steine sind gesetzt, jetzt wird gezogen.", http.StatusConflict)
				return
			}
			if input.Index < 0 || input.Index >= 24 {
				mu.Unlock()
				http.Error(w, "Dieser Punkt liegt nicht auf dem Brett.", http.StatusBadRequest)
				return
			}
			if game.State[input.Index] != "" {
				mu.Unlock()
				http.Error(w, "Dieser Punkt ist schon besetzt.", http.StatusConflict)
				return
			}

			game.State[input.Index] = color
			if color == "red" {
				game.PlacedRed++
			} else {
				game.PlacedBlue++
			}

			if muehleIsMill(game, input.Index, color) {
				game.CapturePending = true
				game.Note = "Mühle! " + player.Name + " darf einen Stein entfernen."
				break
			}

			game.Note = player.Name + " setzt einen Stein."
			if game.PlacedRed == 9 && game.PlacedBlue == 9 {
				game.Phase = "moving"
				game.Current = muehleOther(color)
			} else {
				game.Current = muehleOther(color)
			}

		case "move":
			if game.CapturePending {
				mu.Unlock()
				http.Error(w, "Du musst zuerst einen gegnerischen Stein entfernen.", http.StatusConflict)
				return
			}
			if game.Phase != "moving" {
				mu.Unlock()
				http.Error(w, "Es wird noch gesetzt.", http.StatusConflict)
				return
			}
			if input.From < 0 || input.From >= 24 || input.To < 0 || input.To >= 24 {
				mu.Unlock()
				http.Error(w, "Dieser Punkt liegt nicht auf dem Brett.", http.StatusBadRequest)
				return
			}
			if game.State[input.From] != color {
				mu.Unlock()
				http.Error(w, "Dort steht kein eigener Stein.", http.StatusConflict)
				return
			}
			if game.State[input.To] != "" {
				mu.Unlock()
				http.Error(w, "Der Zielpunkt ist besetzt.", http.StatusConflict)
				return
			}
			if muehleCount(game, color) > 3 && !muehleAdjacent(input.From, input.To) {
				mu.Unlock()
				http.Error(w, "Du darfst nur entlang einer Linie zum Nachbarpunkt ziehen.", http.StatusConflict)
				return
			}

			game.State[input.To] = color
			game.State[input.From] = ""
			game.Note = player.Name + " zieht einen Stein."

			if muehleIsMill(game, input.To, color) {
				game.CapturePending = true
				game.Note = "Mühle! " + player.Name + " darf einen Stein entfernen."
				break
			}

			muehleEndTurn(game, color)

		case "capture":
			if !game.CapturePending {
				mu.Unlock()
				http.Error(w, "Es ist keine Mühle zu schlagen.", http.StatusConflict)
				return
			}
			if input.Index < 0 || input.Index >= 24 {
				mu.Unlock()
				http.Error(w, "Dieser Punkt liegt nicht auf dem Brett.", http.StatusBadRequest)
				return
			}
			if !muehleCanRemove(game, input.Index, color) {
				mu.Unlock()
				http.Error(w, "Du darfst diesen Stein nicht schlagen.", http.StatusConflict)
				return
			}

			game.State[input.Index] = ""
			game.CapturePending = false
			game.Note = player.Name + " schlägt einen Stein."

			if muehleCount(game, muehleOther(color)) <= 2 {
				muehleFinish(game, color)
				break
			}
			muehleEndTurn(game, color)
		}

	case "reset":
		game := room.Muehle
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		fresh := muehleNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
		fresh.Status = "active"
		room.Muehle = fresh

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Mühle-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneMuehle(room.Muehle)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

func muehleAdjacent(from, to int) bool {
	for _, n := range muehleNeighbors[from] {
		if n == to {
			return true
		}
	}
	return false
}
