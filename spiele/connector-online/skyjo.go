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
	StartNext int `json:"startNext"`
	// NextReady merkt, wer die nächste Runde schon bestätigt hat. Die Runde
	// beginnt erst, wenn alle bestätigt haben; der Bot gilt immer als bereit.
	NextReady   [2]bool   `json:"nextReady"`
	Round       int       `json:"round"`
	RoundScores [2]int    `json:"roundScores"`
	Totals      [2]int    `json:"totals"`
	Status      string    `json:"status"` // invited, active, finished
	Winner      string    `json:"winner,omitempty"`
	Note        string    `json:"note,omitempty"`
	RoundWinner string    `json:"roundWinner,omitempty"`
	BotID       string    `json:"botId,omitempty"`
	BotStufe    string    `json:"-"`
	BotFrist    time.Time `json:"-"`
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
	g.NextReady = [2]bool{}
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

// skyjoNextRound beginnt die nächste Runde. Aufgerufen wird es nur, wenn
// alle Spieler bestätigt haben – nie von selbst.
func skyjoNextRound(g *SkyjoGame) {
	g.Round++
	skyjoResetRound(g)
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

// skyjoAktion führt eine Spielaktion aus. Mensch (HTTP) und Bot
// (skyjoBotSchritt) gehen durch dieselbe Prüfung: Phase, Zug, offene Karten,
// Spaltenregel, Rundenende und Punktestand entstehen hier.
// Rückgabe 0 heißt „ausgeführt“; sonst HTTP-Status und Meldung.
func (g *SkyjoGame) skyjoAktion(me int, akt string, index int, source string) (int, string) {
	if g == nil || g.Status != "active" {
		return http.StatusConflict, "Es läuft gerade keine aktive Partie."
	}
	if me < 0 || me > 1 {
		return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
	}
	switch akt {
	case "reveal":
		if g.Phase != "setup" {
			return http.StatusConflict, "Das Aufdecken ist schon vorbei."
		}
		if g.SetupPlayer != me {
			return http.StatusConflict, "Der andere Spieler deckt gerade auf."
		}
		if index < 0 || index >= skyjoCards {
			return http.StatusBadRequest, "Diese Karte gibt es nicht."
		}
		if g.Open[me][index] || g.Removed[me][index] {
			return http.StatusConflict, "Diese Karte ist schon offen."
		}
		g.Open[me][index] = true
		opened := 0
		for _, v := range g.Open[me] {
			if v {
				opened++
			}
		}
		if opened >= 2 {
			if g.SetupPlayer == 0 {
				g.SetupPlayer = 1
				g.Note = skyjoName(g, 1) + ": Jetzt zwei Karten aufdecken."
			} else {
				skyjoStartPlaying(g)
			}
		} else {
			g.Note = "Noch eine Karte aufdecken."
		}

	case "draw":
		if g.Phase != "play" || g.Current != me {
			return http.StatusConflict, "Du bist nicht am Zug."
		}
		if g.Drawn != nil {
			return http.StatusConflict, "Du hast schon eine Karte gezogen."
		}
		quelle := strings.ToLower(strings.TrimSpace(source))
		if quelle == "discard" {
			if len(g.Discard) == 0 {
				return http.StatusConflict, "Der Ablagestapel ist leer."
			}
			value := g.Discard[len(g.Discard)-1]
			g.Discard = g.Discard[:len(g.Discard)-1]
			g.Drawn = &value
			g.DrawnSource = "discard"
			g.Note = "Ablagekarte genommen."
			break
		}

		if len(g.Deck) == 0 {
			if len(g.Discard) <= 1 {
				return http.StatusConflict, "Keine Karten mehr im Nachziehstapel!"
			}
			top := g.Discard[len(g.Discard)-1]
			rest := append([]int(nil), g.Discard[:len(g.Discard)-1]...)
			for i := len(rest) - 1; i > 0; i-- {
				j := rand.Intn(i + 1)
				rest[i], rest[j] = rest[j], rest[i]
			}
			g.Deck = rest
			g.Discard = []int{top}
		}

		value := g.Deck[len(g.Deck)-1]
		g.Deck = g.Deck[:len(g.Deck)-1]
		g.DeckCount = len(g.Deck)
		g.Drawn = &value
		g.DrawnSource = "deck"
		g.Note = "Karte gezogen."

	case "swap":
		if g.Phase != "play" || g.Current != me {
			return http.StatusConflict, "Du bist nicht am Zug."
		}
		if g.Drawn == nil {
			return http.StatusConflict, "Du hast noch keine Karte gezogen."
		}
		if index < 0 || index >= skyjoCards {
			return http.StatusBadRequest, "Diese Karte gibt es nicht."
		}
		if g.Removed[me][index] {
			return http.StatusConflict, "Diese Karte ist schon weg."
		}
		old := g.Cards[me][index]
		g.Cards[me][index] = *g.Drawn
		g.Open[me][index] = true
		g.Discard = append(g.Discard, old)
		g.Drawn = nil
		g.DrawnSource = ""
		skyjoCheckColumns(g, me)
		skyjoFinishTurn(g)

	case "discard":
		if g.Phase != "play" || g.Current != me {
			return http.StatusConflict, "Du bist nicht am Zug."
		}
		if g.Drawn == nil || g.DrawnSource != "deck" {
			return http.StatusConflict, "Es liegt keine gezogene Karte zum Abwerfen bereit."
		}
		g.Discard = append(g.Discard, *g.Drawn)
		g.Drawn = nil
		g.DrawnSource = ""
		hidden := 0
		for i := 0; i < skyjoCards; i++ {
			if !g.Open[me][i] && !g.Removed[me][i] {
				hidden++
			}
		}
		if hidden == 0 {
			skyjoFinishTurn(g)
			break
		}
		g.Phase = "reveal"
		g.Note = "Decke eine deiner verdeckten Karten auf."

	case "opencard":
		if g.Phase != "reveal" || g.Current != me {
			return http.StatusConflict, "Du bist nicht am Zug."
		}
		if index < 0 || index >= skyjoCards {
			return http.StatusBadRequest, "Diese Karte gibt es nicht."
		}
		if g.Open[me][index] || g.Removed[me][index] {
			return http.StatusConflict, "Diese Karte ist schon offen."
		}
		g.Open[me][index] = true
		skyjoCheckColumns(g, me)
		g.Phase = "play"
		skyjoFinishTurn(g)

	case "nextround":
		if g.Phase != "between" || g.Status != "active" {
			return http.StatusConflict, "Eine neue Runde ist gerade nicht möglich."
		}
		// Nicht einer allein: Jeder bestätigt die nächste Runde, erst dann
		// beginnt sie. Der Bot gilt als bereit, sonst wartete man ewig.
		g.NextReady[me] = true
		if b := skyjoBotIndex(g); b >= 0 {
			g.NextReady[b] = true
		}
		if g.NextReady[0] && g.NextReady[1] {
			skyjoNextRound(g)
			break
		}
		g.Note = skyjoName(g, me) + " möchte weiter – " +
			skyjoName(g, 1-me) + " bestätigt noch."

	case "start":
		// Kurzer Anstoß, damit die Anzeige nach dem Beitritt stimmt.
	default:
		return http.StatusBadRequest, "Unbekannte Skyjo-Aktion."
	}
	return 0, ""
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
		if room.Skyjo != nil {
			skyjoBotTreiben(room.Skyjo, time.Now())
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
		Stufe  string `json:"stufe"`
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
		if room.Skyjo != nil && !room.gameSlotFree(room.Skyjo.Status, room.Skyjo.RedID, room.Skyjo.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Skyjo-Partie oder Einladung.", http.StatusConflict)
			return
		}
		g := skyjoNewGame(input.ID, player.Name, botID(1), room.botName(1), randomCode())
		g.BotID, g.BotStufe = botID(1), botStufe(input.Stufe)
		g.Status = "active"
		g.Note = player.Name + " spielt gegen " + room.botName(1) + " (" + botStufeText(g.BotStufe) + "). Decke zwei Karten auf."
		room.Skyjo = g
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
			Time:       uhrzeit(time.Now()),
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

		if action == "reset" {
			if game.Status != "finished" {
				mu.Unlock()
				http.Error(w, "Ein neues Spiel ist erst nach Spielende möglich.", http.StatusConflict)
				return
			}
			fresh := skyjoNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
			fresh.Status = "active"
			fresh.BotID, fresh.BotStufe = game.BotID, game.BotStufe
			room.Skyjo = fresh
			break
		}
		if status, meldung := game.skyjoAktion(me, action, input.Index, input.Source); status != 0 {
			mu.Unlock()
			http.Error(w, meldung, status)
			return
		}
		// Ist der andere Platz nicht mehr im Raum (Zeit abgelaufen, Tab zu),
		// soll der verbliebene Spieler die nächste Runde nicht blockieren.
		if action == "nextround" && game.Phase == "between" && game.Status == "active" {
			other := 1 - me
			otherID := game.RedID
			if other == 1 {
				otherID = game.BlueID
			}
			if _, da := room.Players[otherID]; !da {
				game.NextReady[other] = true
				if game.NextReady[0] && game.NextReady[1] {
					skyjoNextRound(game)
				}
			}
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
