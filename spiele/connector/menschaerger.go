package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Ein Sitz am Tisch. Ready heißt: hat die Einladung angenommen.
type MaenPlayer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Ready  bool   `json:"ready"`
	Pieces []int  `json:"pieces"` // 4 Figuren: -1 Haus, 0..39 Rundweg, 40..43 Ziellauf, 44 Ziel
}

type MenschaergerGame struct {
	ID           string       `json:"id"`
	HostID       string       `json:"hostId"`
	MaxPlayers   int          `json:"maxPlayers"`
	Players      []MaenPlayer `json:"players"`
	Current      int          `json:"current"`
	Dice         int          `json:"dice"`     // 0 = kein offener Wurf
	LastDice     int          `json:"lastDice"` // zuletzt gewürfelte Zahl, bleibt sichtbar
	RollAttempts int          `json:"-"`
	ExtraRoll    bool         `json:"extraRoll"`
	Note         string       `json:"note,omitempty"`
	Status       string       `json:"status"` // invited, active, finished
	Winner       string       `json:"winner,omitempty"`
}

var maenStartIndex = map[string]int{"green": 0, "yellow": 10, "blue": 20, "red": 30}

// Farbreihenfolge je Spielerzahl – identisch zum lokalen Spiel.
var maenColorsByCount = map[int][]string{
	2: {"red", "yellow"},
	3: {"red", "green", "yellow"},
	4: {"red", "green", "yellow", "blue"},
}

func maenNewPieces() []int {
	return []int{-1, -1, -1, -1}
}

// maenLegalMoves liefert die Figuren, die mit diesem Wurf ziehen dürfen.
// Regeln wie im lokalen Spiel: Rauskommen nur mit einer Sechs, nicht über das
// Ziel hinaus, und nie zwei eigene Figuren auf demselben Feld.
func maenLegalMoves(pieces []int, color string, dice int) []int {
	if dice < 1 || dice > 6 || len(pieces) != 4 {
		return nil
	}
	start, ok := maenStartIndex[color]
	if !ok {
		return nil
	}
	moves := []int{}
	for index, progress := range pieces {
		var destination int
		if progress == -1 {
			if dice != 6 {
				continue
			}
			destination = 0
		} else {
			if progress >= 44 {
				continue
			}
			destination = progress + dice
			if destination > 44 {
				continue
			}
		}
		blocked := false
		if destination <= 39 {
			target := (start + destination) % 40
			for otherIndex, other := range pieces {
				if otherIndex == index || other < 0 || other > 39 {
					continue
				}
				if (start+other)%40 == target {
					blocked = true
					break
				}
			}
		} else if destination <= 43 {
			// Zielfelder: nur eine Figur pro Feld. Im Ziel (44) dürfen mehrere stehen.
			for otherIndex, other := range pieces {
				if otherIndex == index {
					continue
				}
				if other == destination {
					blocked = true
					break
				}
			}
		}
		if !blocked {
			moves = append(moves, index)
		}
	}
	return moves
}

func maenContains(list []int, value int) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func maenAllInGoal(pieces []int) bool {
	for _, progress := range pieces {
		if progress != 44 {
			return false
		}
	}
	return true
}

func maenNextPlayer(game *MenschaergerGame) {
	if len(game.Players) == 0 {
		return
	}
	game.Current = (game.Current + 1) % len(game.Players)
	game.Dice = 0
	game.ExtraRoll = false
	game.RollAttempts = 0
}

// maenApplyMove zieht eine Figur und schickt geschlagene Gegner nach Hause.
func maenApplyMove(game *MenschaergerGame, piece int) string {
	player := &game.Players[game.Current]
	progress := player.Pieces[piece]
	destination := 0
	if progress >= 0 {
		destination = progress + game.Dice
	}
	player.Pieces[piece] = destination

	note := ""
	if destination > 39 {
		return note
	}
	target := (maenStartIndex[player.Color] + destination) % 40
	for i := range game.Players {
		if i == game.Current {
			continue
		}
		other := &game.Players[i]
		for k, otherProgress := range other.Pieces {
			if otherProgress < 0 || otherProgress > 39 {
				continue
			}
			if (maenStartIndex[other.Color]+otherProgress)%40 == target {
				other.Pieces[k] = -1
				note = other.Name + " wurde geschlagen!"
			}
		}
	}
	return note
}

// maenAbandoned meldet, ob der Tisch niemanden mehr hat, der spielen kann.
// Ein offener Tisch bleibt bestehen, solange der Gastgeber da ist.
func (g *MenschaergerGame) abandoned(room *Room) bool {
	if g == nil || (g.Status != "invited" && g.Status != "active") {
		return false
	}
	if _, ok := room.Players[g.HostID]; !ok {
		return true
	}
	if g.Status == "invited" {
		return false
	}
	for _, player := range g.Players {
		if _, ok := room.Players[player.ID]; ok {
			return false
		}
	}
	return true
}

func (g *MenschaergerGame) seatOf(id string) int {
	for i, player := range g.Players {
		if player.ID == id {
			return i
		}
	}
	return -1
}

func (g *MenschaergerGame) readyCount() int {
	ready := 0
	for _, player := range g.Players {
		if player.Ready {
			ready++
		}
	}
	return ready
}

func menschaergerHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneMenschaerger(room.Menschaerger)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID         string `json:"id"`
		ToID       string `json:"toId"`
		MaxPlayers int    `json:"maxPlayers"`
		Piece      int    `json:"piece"`
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
		// Erster Klick eröffnet den Tisch, jeder weitere füllt einen Sitz.
		if room.Menschaerger == nil || room.Menschaerger.Status == "finished" {
			limit := input.MaxPlayers
			if limit < 2 {
				limit = 4
			}
			if limit > 4 {
				limit = 4
			}
			room.Menschaerger = &MenschaergerGame{
				ID:         randomCode(),
				HostID:     input.ID,
				MaxPlayers: limit,
				Players: []MaenPlayer{{
					ID: input.ID, Name: player.Name, Ready: true, Pieces: maenNewPieces(),
				}},
				Status: "invited",
			}
		}
		game := room.Menschaerger
		if game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es läuft bereits eine Partie.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber kann weitere Spieler einladen.", http.StatusForbidden)
			return
		}
		target, ok := room.Players[input.ToID]
		if !ok || input.ToID == input.ID {
			mu.Unlock()
			http.Error(w, "Dieser Spieler ist nicht verfügbar.", http.StatusBadRequest)
			return
		}
		if game.seatOf(input.ToID) != -1 {
			mu.Unlock()
			http.Error(w, "Dieser Spieler sitzt schon am Tisch.", http.StatusConflict)
			return
		}
		if len(game.Players) >= game.MaxPlayers {
			mu.Unlock()
			http.Error(w, "Der Tisch ist voll.", http.StatusConflict)
			return
		}
		game.Players = append(game.Players, MaenPlayer{
			ID: input.ToID, Name: target.Name, Pieces: maenNewPieces(),
		})
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Mensch ärgere dich nicht ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "menschaerger",
			GameID:     game.ID,
		})

	case "accept":
		game := room.Menschaerger
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		seat := game.seatOf(input.ID)
		if seat == -1 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		game.Players[seat].Ready = true

	case "ready":
		// Bereit melden (oder wieder abmelden) – direkt aus der Chat-Karte.
		game := room.Menschaerger
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es gibt keinen offenen Tisch.", http.StatusConflict)
			return
		}
		seat := game.seatOf(input.ID)
		if seat == -1 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		game.Players[seat].Ready = !game.Players[seat].Ready
		text := player.Name + " ist bereit."
		if !game.Players[seat].Ready {
			text = player.Name + " ist noch nicht bereit."
		}
		room.addMessage(Message{
			Text:   text,
			Time:   time.Now().Format("15:04:05"),
			System: true,
		})

	case "decline":
		game := room.Menschaerger
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		seat := game.seatOf(input.ID)
		if seat == -1 || input.ID == game.HostID {
			mu.Unlock()
			http.Error(w, "Diese Einladung stammt nicht von dir.", http.StatusForbidden)
			return
		}
		game.Players = append(game.Players[:seat], game.Players[seat+1:]...)

	case "start":
		game := room.Menschaerger
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es gibt keinen offenen Tisch.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber startet die Partie.", http.StatusForbidden)
			return
		}
		seated := []MaenPlayer{}
		for _, seat := range game.Players {
			if seat.Ready {
				seated = append(seated, seat)
			}
		}
		if len(seated) < 2 {
			mu.Unlock()
			http.Error(w, "Mindestens zwei Spieler müssen bereit sein.", http.StatusConflict)
			return
		}
		if len(seated) > 4 {
			seated = seated[:4]
		}
		colors := maenColorsByCount[len(seated)]
		for i := range seated {
			seated[i].Color = colors[i]
			seated[i].Pieces = maenNewPieces()
			seated[i].Ready = true
		}
		game.Players = seated
		game.Current = 0
		game.Dice = 0
		game.LastDice = 0
		game.ExtraRoll = false
		game.RollAttempts = 0
		game.Note = ""
		game.Winner = ""
		game.Status = "active"

	case "roll":
		game := room.Menschaerger
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine Partie.", http.StatusConflict)
			return
		}
		player := &game.Players[game.Current]
		if player.ID != input.ID {
			mu.Unlock()
			http.Error(w, "Du bist nicht am Zug.", http.StatusConflict)
			return
		}
		if game.Dice != 0 {
			mu.Unlock()
			http.Error(w, "Ziehe zuerst deine Figur.", http.StatusConflict)
			return
		}
		dice := rand.Intn(6) + 1
		game.Dice = dice
		game.LastDice = dice
		game.Note = ""

		hasPieceInPlay := false
		for _, progress := range player.Pieces {
			if progress >= 0 && progress < 44 {
				hasPieceInPlay = true
			}
		}
		if dice != 6 && !hasPieceInPlay && !game.ExtraRoll {
			game.RollAttempts++
			if game.RollAttempts < 3 {
				game.Dice = 0
				game.Note = "Keine Sechs! Noch " + strconv.Itoa(3-game.RollAttempts) + " Wurfversuch(e)."
				break
			}
			game.Dice = 0
			maenNextPlayer(game)
			game.Note = "Dreimal keine Sechs – Spielerwechsel …"
			break
		}
		if dice == 6 {
			game.RollAttempts = 0
		}
		if len(maenLegalMoves(player.Pieces, player.Color, dice)) == 0 {
			game.Dice = 0
			maenNextPlayer(game)
			game.Note = "Kein gültiger Zug möglich – nächster Spieler …"
		}

	case "move":
		game := room.Menschaerger
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine Partie.", http.StatusConflict)
			return
		}
		player := &game.Players[game.Current]
		if player.ID != input.ID {
			mu.Unlock()
			http.Error(w, "Du bist nicht am Zug.", http.StatusConflict)
			return
		}
		if game.Dice == 0 {
			mu.Unlock()
			http.Error(w, "Würfle zuerst.", http.StatusConflict)
			return
		}
		if !maenContains(maenLegalMoves(player.Pieces, player.Color, game.Dice), input.Piece) {
			mu.Unlock()
			http.Error(w, "Mit dieser Figur kannst du nicht ziehen.", http.StatusBadRequest)
			return
		}
		note := maenApplyMove(game, input.Piece)
		if maenAllInGoal(player.Pieces) {
			game.Status = "finished"
			game.Winner = player.ID
			game.Dice = 0
			game.ExtraRoll = false
			game.Note = player.Name + " bringt alle Figuren ins Ziel!"
			break
		}
		rolledSix := game.Dice == 6
		game.Dice = 0
		if rolledSix {
			game.ExtraRoll = true
			game.RollAttempts = 0
			game.Note = "Sechs! Du darfst noch einmal würfeln – Würfel anklicken."
		} else {
			maenNextPlayer(game)
			game.Note = note
		}

	case "reset":
		game := room.Menschaerger
		if game == nil || game.Status != "finished" {
			mu.Unlock()
			http.Error(w, "Eine neue Runde ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber startet eine neue Runde.", http.StatusForbidden)
			return
		}
		colors := maenColorsByCount[len(game.Players)]
		for i := range game.Players {
			game.Players[i].Color = colors[i]
			game.Players[i].Pieces = maenNewPieces()
		}
		game.Current = 0
		game.Dice = 0
		game.LastDice = 0
		game.ExtraRoll = false
		game.RollAttempts = 0
		game.Note = ""
		game.Winner = ""
		game.Status = "active"

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneMenschaerger(room.Menschaerger)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
