package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Skyjo: je 12 Karten in einem 4x3-Feld, zwei Startkarten offen,
// drei gleiche Werte in einer Spalte zählen 0, wer am Ende die wenigsten
// Punkte hat, gewinnt. Gespielt wird bis 100 Punkte.
const skyjoCards = 12

var skyjoCounts = [][2]int{
	{-2, 5}, {-1, 10}, {0, 15},
	{1, 10}, {2, 10}, {3, 10}, {4, 10},
	{5, 10}, {6, 10}, {7, 10}, {8, 10},
	{9, 10}, {10, 10}, {11, 10}, {12, 10},
}

type SkyjoGame struct {
	ID          string    `json:"id"`
	RedID       string    `json:"redId"`
	RedName     string    `json:"redName"`
	BlueID      string    `json:"blueId"`
	BlueName    string    `json:"blueName"`
	Cards       [2][]int  `json:"cards"`
	Open        [2][]bool `json:"open"`
	Removed     [2][]bool `json:"removed"`
	Deck        []int     `json:"-"`
	DeckCount   int       `json:"deckCount"`
	Discard     []int     `json:"discard"`
	Current     int       `json:"current"`
	SetupPlayer int       `json:"setupPlayer"`
	Phase       string    `json:"phase"` // setup, play, reveal, between
	Drawn       *int      `json:"drawn"`
	DrawnSource string    `json:"drawnSource"`
	Finisher    int       `json:"finisher"` // -1 = noch keiner
	LastTurn    int       `json:"lastTurn"`
	// StartNext ist der Spieler, der die nächste Runde beginnt – laut Regel
	// der, der die letzte Runde beendet hat (-1 = keine Vorgabe).
	StartNext   int    `json:"startNext"`
	Round       int    `json:"round"`
	RoundScores [2]int `json:"roundScores"`
	Totals      [2]int `json:"totals"`
	Status      string `json:"status"` // invited, active, finished
	Winner      string `json:"winner,omitempty"`
	Note        string `json:"note,omitempty"`
	RoundWinner string `json:"roundWinner,omitempty"`
}

func skyjoMakeDeck() []int {
	deck := []int{}
	for _, entry := range skyjoCounts {
		for i := 0; i < entry[1]; i++ {
			deck = append(deck, entry[0])
		}
	}
	for i := len(deck) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
	return deck
}

func skyjoResetRound(g *SkyjoGame) {
	g.Deck = skyjoMakeDeck()
	g.Discard = []int{}
	for p := 0; p < 2; p++ {
		g.Cards[p] = make([]int, skyjoCards)
		g.Open[p] = make([]bool, skyjoCards)
		g.Removed[p] = make([]bool, skyjoCards)
		for i := 0; i < skyjoCards; i++ {
			g.Cards[p][i] = g.Deck[len(g.Deck)-1]
			g.Deck = g.Deck[:len(g.Deck)-1]
		}
	}
	g.Discard = append(g.Discard, g.Deck[len(g.Deck)-1])
	g.Deck = g.Deck[:len(g.Deck)-1]
	g.DeckCount = len(g.Deck)

	g.Current = 0
	g.SetupPlayer = 0
	g.Phase = "setup"
	g.Drawn = nil
	g.DrawnSource = ""
	g.Finisher = -1
	g.LastTurn = -1
	g.RoundScores = [2]int{}
	g.RoundWinner = ""
	g.Note = "Decke zwei Karten auf."
}

func skyjoNewGame(hostID, hostName, guestID, guestName, id string) *SkyjoGame {
	g := &SkyjoGame{
		ID:        id,
		RedID:     hostID,
		RedName:   hostName,
		BlueID:    guestID,
		BlueName:  guestName,
		Round:     1,
		Status:    "invited",
		Finisher:  -1,
		LastTurn:  -1,
		StartNext: -1,
	}
	skyjoResetRound(g)
	return g
}

func cloneSkyjo(g *SkyjoGame) *SkyjoGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Deck = append([]int(nil), g.Deck...)
	c.Discard = append([]int(nil), g.Discard...)
	for p := 0; p < 2; p++ {
		c.Cards[p] = append([]int(nil), g.Cards[p]...)
		c.Open[p] = append([]bool(nil), g.Open[p]...)
		c.Removed[p] = append([]bool(nil), g.Removed[p]...)
	}
	if g.Drawn != nil {
		v := *g.Drawn
		c.Drawn = &v
	}
	return &c
}

func skyjoColorOf(g *SkyjoGame, id string) string {
	if g.RedID == id {
		return "red"
	}
	if g.BlueID == id {
		return "blue"
	}
	return ""
}

func skyjoIndex(g *SkyjoGame, color string) int {
	if color == "red" {
		return 0
	}
	return 1
}

func skyjoName(g *SkyjoGame, player int) string {
	if player == 0 {
		return g.RedName
	}
	return g.BlueName
}

func skyjoCheckColumns(g *SkyjoGame, player int) {
	cards := g.Cards[player]
	open := g.Open[player]
	removed := g.Removed[player]

	for col := 0; col < 4; col++ {
		a, b, c := col, col+4, col+8
		if removed[a] || removed[b] || removed[c] {
			continue
		}
		if !open[a] || !open[b] || !open[c] {
			continue
		}
		if cards[a] == cards[b] && cards[b] == cards[c] {
			g.Removed[player][a] = true
			g.Removed[player][b] = true
			g.Removed[player][c] = true
		}
	}
}

func skyjoAllRevealed(g *SkyjoGame, player int) bool {
	for i := 0; i < skyjoCards; i++ {
		if !g.Open[player][i] && !g.Removed[player][i] {
			return false
		}
	}
	return true
}

func skyjoScore(g *SkyjoGame, player int) int {
	sum := 0
	for i := 0; i < skyjoCards; i++ {
		if g.Removed[player][i] {
			continue
		}
		sum += g.Cards[player][i]
	}
	return sum
}

func skyjoStartPlaying(g *SkyjoGame) {
	// Ab der zweiten Runde beginnt, wer die vorige Runde beendet hat.
	if g.StartNext >= 0 {
		g.Current = g.StartNext
		g.StartNext = -1
		g.Phase = "play"
		g.Note = skyjoName(g, g.Current) + " beginnt!"
		return
	}

	sum0 := 0
	sum1 := 0
	for i := 0; i < skyjoCards; i++ {
		if g.Open[0][i] {
			sum0 += g.Cards[0][i]
		}
		if g.Open[1][i] {
			sum1 += g.Cards[1][i]
		}
	}
	switch {
	case sum0 > sum1:
		g.Current = 0
	case sum1 > sum0:
		g.Current = 1
	case rand.Intn(2) == 0:
		g.Current = 0
	default:
		g.Current = 1
	}
	g.Phase = "play"
	g.Note = skyjoName(g, g.Current) + " beginnt!"
}

func skyjoFinishTurn(g *SkyjoGame) {
	justPlayed := g.Current

	if g.Finisher >= 0 && justPlayed == g.LastTurn {
		skyjoFinishRound(g)
		return
	}

	if g.Finisher < 0 && skyjoAllRevealed(g, justPlayed) {
		g.Finisher = justPlayed
		g.LastTurn = 1 - justPlayed
		g.Current = g.LastTurn
		g.Phase = "play"
		g.Note = skyjoName(g, g.Current) + " hat noch den letzten Zug!"
		return
	}

	g.Current = 1 - g.Current
	g.Phase = "play"
	g.Note = skyjoName(g, g.Current) + " ist am Zug."
}

func skyjoFinishRound(g *SkyjoGame) {
	for player := 0; player < 2; player++ {
		for i := 0; i < skyjoCards; i++ {
			g.Open[player][i] = true
		}
	}

	// Die Dreier-Regel gilt auch bei der Schlussabrechnung: Werden beim
	// Aufdecken drei gleiche in einer Spalte sichtbar, zählt die Spalte nicht.
	skyjoCheckColumns(g, 0)
	skyjoCheckColumns(g, 1)

	s0 := skyjoScore(g, 0)
	s1 := skyjoScore(g, 1)

	// Wer die Runde beendet, muss danach die streng niedrigste Punktzahl haben –
	// erreicht jemand weniger oder gleich viel, zählt die Summe doppelt (nur
	// positive Werte).
	if g.Finisher == 0 && s0 >= s1 && s0 > 0 {
		s0 *= 2
	}
	if g.Finisher == 1 && s1 >= s0 && s1 > 0 {
		s1 *= 2
	}

	g.RoundScores = [2]int{s0, s1}
	g.Totals[0] += s0
	g.Totals[1] += s1
	g.Phase = "between"
	g.Drawn = nil
	g.DrawnSource = ""

	switch {
	case s0 < s1:
		g.RoundWinner = "red"
	case s1 < s0:
		g.RoundWinner = "blue"
	default:
		g.RoundWinner = "draw"
	}

	// Der Rundenbeender beginnt die nächste Runde.
	g.StartNext = g.Finisher

	if g.Totals[0] >= 100 || g.Totals[1] >= 100 {
		g.Status = "finished"
		switch {
		case g.Totals[0] < g.Totals[1]:
			g.Winner = "red"
		case g.Totals[1] < g.Totals[0]:
			g.Winner = "blue"
		default:
			// Gleichstand: die bessere letzte Runde entscheidet.
			switch {
			case s0 < s1:
				g.Winner = "red"
			case s1 < s0:
				g.Winner = "blue"
			default:
				g.Winner = "draw"
			}
		}
	}

	g.Note = "Runde " + skyjoNumber(g.Round) + " beendet – " + skyjoRoundText(g)
}

func skyjoNumber(n int) string {
	return strconv.Itoa(n)
}

func skyjoRoundText(g *SkyjoGame) string {
	if g.RoundWinner == "draw" {
		return "unentschieden"
	}
	return skyjoName(g, skyjoIndex(g, g.RoundWinner)) + " gewinnt die Runde"
}

func skyjoHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneSkyjo(room.Skyjo)
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
		Index  int    `json:"index"`
		Source string `json:"source"`
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
		if room.Skyjo != nil && !room.gameSlotFree(room.Skyjo.Status, room.Skyjo.RedID, room.Skyjo.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Skyjo-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Skyjo = skyjoNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Skyjo ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "skyjo",
			GameID:     room.Skyjo.ID,
		})

	case "accept":
		game := room.Skyjo
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = "Decke zwei Karten auf."

	case "decline":
		game := room.Skyjo
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "start", "reveal", "draw", "swap", "discard", "opencard", "nextround", "reset":
		game := room.Skyjo
		if game == nil || (action != "reset" && game.Status != "active") {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := skyjoColorOf(game, input.ID)
		if color == "" {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		me := skyjoIndex(game, color)

		switch action {
		case "reveal":
			if game.Phase != "setup" {
				mu.Unlock()
				http.Error(w, "Das Aufdecken ist schon vorbei.", http.StatusConflict)
				return
			}
			if game.SetupPlayer != me {
				mu.Unlock()
				http.Error(w, "Der andere Spieler deckt gerade auf.", http.StatusConflict)
				return
			}
			if input.Index < 0 || input.Index >= skyjoCards {
				mu.Unlock()
				http.Error(w, "Diese Karte gibt es nicht.", http.StatusBadRequest)
				return
			}
			if game.Open[me][input.Index] || game.Removed[me][input.Index] {
				mu.Unlock()
				http.Error(w, "Diese Karte ist schon offen.", http.StatusConflict)
				return
			}

			game.Open[me][input.Index] = true
			opened := 0
			for _, v := range game.Open[me] {
				if v {
					opened++
				}
			}

			if opened >= 2 {
				if game.SetupPlayer == 0 {
					game.SetupPlayer = 1
					game.Note = skyjoName(game, 1) + ": Jetzt zwei Karten aufdecken."
				} else {
					skyjoStartPlaying(game)
				}
			} else {
				game.Note = "Noch eine Karte aufdecken."
			}

		case "draw":
			if game.Phase != "play" || game.Current != me {
				mu.Unlock()
				http.Error(w, "Du bist nicht am Zug.", http.StatusConflict)
				return
			}
			if game.Drawn != nil {
				mu.Unlock()
				http.Error(w, "Du hast schon eine Karte gezogen.", http.StatusConflict)
				return
			}

			source := strings.ToLower(strings.TrimSpace(input.Source))
			if source == "discard" {
				if len(game.Discard) == 0 {
					mu.Unlock()
					http.Error(w, "Der Ablagestapel ist leer.", http.StatusConflict)
					return
				}
				value := game.Discard[len(game.Discard)-1]
				game.Discard = game.Discard[:len(game.Discard)-1]
				game.Drawn = &value
				game.DrawnSource = "discard"
				game.Note = "Ablagekarte genommen."
				break
			}

			if len(game.Deck) == 0 {
				if len(game.Discard) <= 1 {
					mu.Unlock()
					http.Error(w, "Keine Karten mehr im Nachziehstapel!", http.StatusConflict)
					return
				}
				top := game.Discard[len(game.Discard)-1]
				rest := append([]int(nil), game.Discard[:len(game.Discard)-1]...)
				for i := len(rest) - 1; i > 0; i-- {
					j := rand.Intn(i + 1)
					rest[i], rest[j] = rest[j], rest[i]
				}
				game.Deck = rest
				game.Discard = []int{top}
			}

			value := game.Deck[len(game.Deck)-1]
			game.Deck = game.Deck[:len(game.Deck)-1]
			game.DeckCount = len(game.Deck)
			game.Drawn = &value
			game.DrawnSource = "deck"
			game.Note = "Karte gezogen."

		case "swap":
			if game.Phase != "play" || game.Current != me {
				mu.Unlock()
				http.Error(w, "Du bist nicht am Zug.", http.StatusConflict)
				return
			}
			if game.Drawn == nil {
				mu.Unlock()
				http.Error(w, "Du hast noch keine Karte gezogen.", http.StatusConflict)
				return
			}
			if input.Index < 0 || input.Index >= skyjoCards {
				mu.Unlock()
				http.Error(w, "Diese Karte gibt es nicht.", http.StatusBadRequest)
				return
			}
			if game.Removed[me][input.Index] {
				mu.Unlock()
				http.Error(w, "Diese Karte ist schon weg.", http.StatusConflict)
				return
			}

			old := game.Cards[me][input.Index]
			game.Cards[me][input.Index] = *game.Drawn
			game.Open[me][input.Index] = true
			game.Discard = append(game.Discard, old)
			game.Drawn = nil
			game.DrawnSource = ""

			skyjoCheckColumns(game, me)
			skyjoFinishTurn(game)

		case "discard":
			if game.Phase != "play" || game.Current != me {
				mu.Unlock()
				http.Error(w, "Du bist nicht am Zug.", http.StatusConflict)
				return
			}
			if game.Drawn == nil || game.DrawnSource != "deck" {
				mu.Unlock()
				http.Error(w, "Es liegt keine gezogene Karte zum Abwerfen bereit.", http.StatusConflict)
				return
			}

			game.Discard = append(game.Discard, *game.Drawn)
			game.Drawn = nil
			game.DrawnSource = ""

			hidden := 0
			for i := 0; i < skyjoCards; i++ {
				if !game.Open[me][i] && !game.Removed[me][i] {
					hidden++
				}
			}

			if hidden == 0 {
				skyjoFinishTurn(game)
				break
			}

			game.Phase = "reveal"
			game.Note = "Decke eine deiner verdeckten Karten auf."

		case "opencard":
			if game.Phase != "reveal" || game.Current != me {
				mu.Unlock()
				http.Error(w, "Du bist nicht am Zug.", http.StatusConflict)
				return
			}
			if input.Index < 0 || input.Index >= skyjoCards {
				mu.Unlock()
				http.Error(w, "Diese Karte gibt es nicht.", http.StatusBadRequest)
				return
			}
			if game.Open[me][input.Index] || game.Removed[me][input.Index] {
				mu.Unlock()
				http.Error(w, "Diese Karte ist schon offen.", http.StatusConflict)
				return
			}

			game.Open[me][input.Index] = true
			skyjoCheckColumns(game, me)
			game.Phase = "play"
			skyjoFinishTurn(game)

		case "nextround":
			if game.Phase != "between" || game.Status != "active" {
				mu.Unlock()
				http.Error(w, "Eine neue Runde ist gerade nicht möglich.", http.StatusConflict)
				return
			}
			game.Round++
			skyjoResetRound(game)

		case "reset":
			if game.Status != "finished" {
				mu.Unlock()
				http.Error(w, "Ein neues Spiel ist erst nach Spielende möglich.", http.StatusConflict)
				return
			}
			fresh := skyjoNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
			fresh.Status = "active"
			room.Skyjo = fresh

		case "start":
			// Kurzer Anstoß, damit die Anzeige nach dem Beitritt stimmt.
		}

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Skyjo-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneSkyjo(room.Skyjo)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
