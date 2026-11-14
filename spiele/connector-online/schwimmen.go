package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 31 (Schwimmen): drei Karten auf der Hand, drei offen auf dem Tisch. Tauschen,
// schieben oder klopfen – wer am Ende die wenigsten Punkte hat, verliert ein
// Leben. Drei Leben, wer alle verliert, ist raus.
const (
	schwimmMaxPlayers = 9
	schwimmLives      = 3
	schwimmHandSize   = 3
)

var schwimmSuits = []string{"♥", "♦", "♠", "♣"}
var schwimmValues = []string{"7", "8", "9", "10", "B", "D", "K", "A"}

type SchwimmCard struct {
	Suit  string `json:"suit"`
	Value string `json:"value"`
}

type SchwimmPlayer struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Hand  []SchwimmCard `json:"-"`
	Lives int           `json:"lives"`
	Out   bool          `json:"out"`
	Ready bool          `json:"ready"`
}

type SchwimmenGame struct {
	ID         string           `json:"id"`
	HostID     string           `json:"hostId"`
	Players    []*SchwimmPlayer `json:"-"`
	Deck       []SchwimmCard    `json:"-"`
	Table      []SchwimmCard    `json:"table"`
	Current    int              `json:"current"`
	Direction  int              `json:"direction"`
	Passes     int              `json:"passes"`
	FinalTurns int              `json:"finalTurns"`
	Round      int              `json:"round"`
	Phase      string           `json:"phase"`  // lobby, playing, between, finished
	Status     string           `json:"status"` // invited, active, finished
	Winner     string           `json:"winner,omitempty"`
	Note       string           `json:"note,omitempty"`
	Result     []string         `json:"result,omitempty"`
	DeckCount  int              `json:"deckCount"`
}

type SchwimmSeatView struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Count int     `json:"count"`
	Lives int     `json:"lives"`
	Out   bool    `json:"out"`
	Ready bool    `json:"ready"`
	Score float64 `json:"score"`
	Shown bool    `json:"shown"`
	// Cards wird erst nach der Runde gefüllt: Dann liegen alle Karten offen,
	// damit jeder die Wertung nachvollziehen kann.
	Cards []SchwimmCard `json:"cards,omitempty"`
}

type SchwimmenView struct {
	ID         string            `json:"id"`
	HostID     string            `json:"hostId"`
	Status     string            `json:"status"`
	Phase      string            `json:"phase"`
	Seats      []SchwimmSeatView `json:"seats"`
	Current    int               `json:"current"`
	Direction  int               `json:"direction"`
	Table      []SchwimmCard     `json:"table"`
	Passes     int               `json:"passes"`
	FinalTurns int               `json:"finalTurns"`
	Round      int               `json:"round"`
	DeckCount  int               `json:"deckCount"`
	Winner     string            `json:"winner,omitempty"`
	Note       string            `json:"note,omitempty"`
	Result     []string          `json:"result,omitempty"`
	You        int               `json:"you"`
	Hand       []SchwimmCard     `json:"hand"`
	Score      float64           `json:"score"`
	MyTurn     bool              `json:"myTurn"`
}

func schwimmDeck() []SchwimmCard {
	deck := make([]SchwimmCard, 0, len(schwimmSuits)*len(schwimmValues))
	for _, suit := range schwimmSuits {
		for _, value := range schwimmValues {
			deck = append(deck, SchwimmCard{Suit: suit, Value: value})
		}
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

func schwimmDraw(g *SchwimmenGame) (SchwimmCard, bool) {
	if len(g.Deck) == 0 {
		return SchwimmCard{}, false
	}
	card := g.Deck[len(g.Deck)-1]
	g.Deck = g.Deck[:len(g.Deck)-1]
	return card, true
}

func schwimmCardPoints(card SchwimmCard) int {
	switch card.Value {
	case "A":
		return 11
	case "K", "D", "B":
		return 10
	}
	if n, err := strconv.Atoi(card.Value); err == nil {
		return n
	}
	return 0
}

// schwimmPoints zählt die beste Farbe. Drei gleiche Werte zählen 30,5.
func schwimmPoints(hand []SchwimmCard) float64 {
	if len(hand) == 3 && hand[0].Value == hand[1].Value && hand[1].Value == hand[2].Value {
		return 30.5
	}
	best := 0
	for _, suit := range schwimmSuits {
		sum := 0
		for _, card := range hand {
			if card.Suit == suit {
				sum += schwimmCardPoints(card)
			}
		}
		if sum > best {
			best = sum
		}
	}
	return float64(best)
}

// schwimmIsFeuer: drei Asse – das "Feuer". Wer es hat, lässt alle anderen
// ein Leben verlieren, die Runde ist sofort vorbei.
func schwimmIsFeuer(hand []SchwimmCard) bool {
	if len(hand) != 3 {
		return false
	}
	for _, card := range hand {
		if card.Value != "A" {
			return false
		}
	}
	return true
}

// schwimmFeuer wertet drei Asse aus: Alle anderen verlieren ein Leben.
func (g *SchwimmenGame) schwimmFeuer(seat int) {
	name := g.Players[seat].Name
	verloren := []string{}
	for i, p := range g.Players {
		if i == seat || p.Out {
			continue
		}
		p.Lives--
		if p.Lives <= 0 {
			p.Lives = 0
			p.Out = true
		}
		verloren = append(verloren, p.Name)
	}

	g.Result = []string{
		name + " hat Feuer – drei Asse!",
		"Alle anderen verlieren ein Leben.",
		"Verloren: " + strings.Join(verloren, ", "),
	}
	g.FinalTurns = -1
	g.Passes = 0

	if stillIn := schwimmActive(g); stillIn <= 1 {
		g.Status = "finished"
		g.Phase = "finished"
		if stillIn == 1 {
			for _, p := range g.Players {
				if !p.Out {
					g.Winner = p.Name
				}
			}
			g.Result = append(g.Result, "🏆 "+g.Winner+" gewinnt das Spiel!")
			g.Note = g.Winner + " gewinnt das Spiel!"
		} else {
			g.Note = "Kein Spieler ist mehr im Spiel."
			g.Result = append(g.Result, "Kein Spieler ist mehr im Spiel.")
		}
		return
	}

	g.Phase = "between"
	g.Note = name + " hat Feuer – alle anderen verlieren ein Leben."
}

func schwimmScoreText(points float64) string {
	if points == 30.5 {
		return "30½"
	}
	return strconv.Itoa(int(points))
}

func schwimmNewGame(hostID, hostName, id string) *SchwimmenGame {
	g := &SchwimmenGame{
		ID:         id,
		HostID:     hostID,
		Status:     "invited",
		Phase:      "lobby",
		Direction:  1,
		FinalTurns: -1,
		Round:      1,
	}
	g.Players = append(g.Players, &SchwimmPlayer{ID: hostID, Name: hostName, Lives: schwimmLives})
	return g
}

func schwimmSeat(g *SchwimmenGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func schwimmActive(g *SchwimmenGame) int {
	count := 0
	for _, p := range g.Players {
		if !p.Out {
			count++
		}
	}
	return count
}

// schwimmNext sucht den nächsten Spieler, der noch im Spiel ist.
func schwimmNext(g *SchwimmenGame, from int) int {
	n := len(g.Players)
	for step := 1; step <= n; step++ {
		next := ((from+step*g.Direction)%n + n) % n
		if !g.Players[next].Out {
			return next
		}
	}
	return from
}

// schwimmDeal gibt jedem im Spiel drei Karten und legt drei Tischkarten aus.
func schwimmDeal(g *SchwimmenGame) {
	g.Deck = schwimmDeck()
	g.Table = nil
	g.Passes = 0
	g.FinalTurns = -1
	g.Result = nil
	for _, p := range g.Players {
		p.Hand = nil
		p.Ready = false
		if p.Out {
			continue
		}
		for i := 0; i < schwimmHandSize; i++ {
			if card, ok := schwimmDraw(g); ok {
				p.Hand = append(p.Hand, card)
			}
		}
	}
	for i := 0; i < schwimmHandSize; i++ {
		if card, ok := schwimmDraw(g); ok {
			g.Table = append(g.Table, card)
		}
	}
	if g.Players[g.Current].Out {
		g.Current = schwimmNext(g, g.Current)
	}
	g.Phase = "playing"
	g.Status = "active"
	g.Note = g.Players[g.Current].Name + " beginnt."
	g.DeckCount = len(g.Deck)
}

// schwimmAdvance zählt die Klopfrunde herunter und gibt den Zug weiter.
func (g *SchwimmenGame) schwimmAdvance() {
	if g.FinalTurns > 0 {
		g.FinalTurns--
		if g.FinalTurns == 0 {
			g.schwimmFinishRound("Die Klopfrunde ist beendet.")
			return
		}
	}
	g.Current = schwimmNext(g, g.Current)
}

// schwimmFinishRound wertet aus: Wer die wenigsten Punkte hat, verliert ein
// Leben.
func (g *SchwimmenGame) schwimmFinishRound(reason string) {
	lowest := 0.0
	first := true
	for _, p := range g.Players {
		if p.Out {
			continue
		}
		points := schwimmPoints(p.Hand)
		if first || points < lowest {
			lowest = points
			first = false
		}
	}

	losers := []string{}
	for _, p := range g.Players {
		if p.Out || schwimmPoints(p.Hand) != lowest {
			continue
		}
		p.Lives--
		if p.Lives <= 0 {
			p.Lives = 0
			p.Out = true
		}
		losers = append(losers, p.Name)
	}

	// Die Karten der Mitspieler stehen in der Auswertung; hier nur die Summe.
	lines := []string{reason, "Niedrigste Punktzahl: " + schwimmScoreText(lowest), "Verloren: " + strings.Join(losers, ", ")}

	g.Result = lines
	g.FinalTurns = -1
	g.Passes = 0

	if stillIn := schwimmActive(g); stillIn <= 1 {
		g.Status = "finished"
		g.Phase = "finished"
		if stillIn == 1 {
			for _, p := range g.Players {
				if !p.Out {
					g.Winner = p.Name
				}
			}
			g.Note = g.Winner + " gewinnt das Spiel!"
			g.Result = append(g.Result, "🏆 "+g.Winner+" gewinnt das Spiel!")
		} else {
			g.Note = "Kein Spieler ist mehr im Spiel."
			g.Result = append(g.Result, "Kein Spieler ist mehr im Spiel.")
		}
		return
	}

	verb := "verliert"
	if len(losers) > 1 {
		verb = "verlieren"
	}
	g.Phase = "between"
	g.Note = "Runde beendet – " + strings.Join(losers, ", ") + " " + verb + " ein Leben."
}

func (g *SchwimmenGame) viewFor(id string) SchwimmenView {
	view := SchwimmenView{
		ID:         g.ID,
		HostID:     g.HostID,
		Status:     g.Status,
		Phase:      g.Phase,
		Current:    g.Current,
		Direction:  g.Direction,
		Table:      append([]SchwimmCard(nil), g.Table...),
		Passes:     g.Passes,
		FinalTurns: g.FinalTurns,
		Round:      g.Round,
		DeckCount:  len(g.Deck),
		Winner:     g.Winner,
		Note:       g.Note,
		Result:     append([]string(nil), g.Result...),
		You:        schwimmSeat(g, id),
		Hand:       []SchwimmCard{},
	}
	for _, p := range g.Players {
		seat := SchwimmSeatView{
			ID:    p.ID,
			Name:  p.Name,
			Count: len(p.Hand),
			Lives: p.Lives,
			Out:   p.Out,
			Ready: p.Ready,
		}
		// Punkte und Karten der anderen bleiben verborgen, solange die Runde läuft.
		if p.Out || g.Phase != "playing" {
			seat.Score = schwimmPoints(p.Hand)
			seat.Shown = true
		}
		if g.Phase != "playing" {
			seat.Cards = append([]SchwimmCard(nil), p.Hand...)
		}
		view.Seats = append(view.Seats, seat)
	}
	if view.You >= 0 {
		me := g.Players[view.You]
		view.Hand = append([]SchwimmCard(nil), me.Hand...)
		view.Score = schwimmPoints(me.Hand)
		view.MyTurn = g.Status == "active" && g.Phase == "playing" && g.Current == view.You
		view.Seats[view.You].Score = view.Score
		view.Seats[view.You].Shown = true
	}
	return view
}

func (g *SchwimmenGame) allGone(r *Room) bool {
	present := 0
	for _, p := range g.Players {
		if _, ok := r.Players[p.ID]; ok {
			present++
		}
	}
	return present == 0
}

func cloneSchwimmen(g *SchwimmenGame) *SchwimmenGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Deck = append([]SchwimmCard(nil), g.Deck...)
	c.Table = append([]SchwimmCard(nil), g.Table...)
	c.Result = append([]string(nil), g.Result...)
	c.Players = make([]*SchwimmPlayer, len(g.Players))
	for i, p := range g.Players {
		copied := *p
		copied.Hand = append([]SchwimmCard(nil), p.Hand...)
		c.Players[i] = &copied
	}
	return &c
}

func schwimmenHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneSchwimmen(room.Schwimmen)
		var view *SchwimmenView
		if game != nil {
			v := game.viewFor(id)
			view = &v
		}
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": view})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID         string `json:"id"`
		HandIndex  int    `json:"handIndex"`
		TableIndex int    `json:"tableIndex"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.ID = strings.TrimSpace(input.ID)
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

	fail := func(message string, status int) {
		mu.Unlock()
		http.Error(w, message, status)
	}

	game := room.Schwimmen

	switch action {
	case "create":
		if game != nil && game.Status != "finished" {
			fail("An diesem Tisch läuft schon eine Runde 31.", http.StatusConflict)
			return
		}
		room.Schwimmen = schwimmNewGame(input.ID, player.Name, randomCode())
		room.Schwimmen.Note = player.Name + " eröffnet eine Runde 31 – bis zu 9 Spieler, mitspielen über die Spielkarte."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Runde 31 – bis zu 9 Spieler, mitspielen über die Spielkarte.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "join":
		if game == nil || game.Status != "invited" {
			fail("Es läuft gerade keine offene Runde 31.", http.StatusConflict)
			return
		}
		if schwimmSeat(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= schwimmMaxPlayers {
			fail("Diese Runde ist mit 9 Spielern voll.", http.StatusConflict)
			return
		}
		game.Players = append(game.Players, &SchwimmPlayer{ID: input.ID, Name: player.Name, Lives: schwimmLives})
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zur Runde 31.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "ready":
		if game == nil || game.Status != "invited" {
			fail("Bereit drücken geht nur, solange die Runde noch nicht läuft.", http.StatusConflict)
			return
		}
		seat := schwimmSeat(game, input.ID)
		if seat < 0 {
			fail("Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		game.Players[seat].Ready = !game.Players[seat].Ready
		if game.Players[seat].Ready {
			game.Note = player.Name + " ist bereit."
		} else {
			game.Note = player.Name + " wartet noch."
		}

	case "start":
		if game == nil || game.Status != "invited" {
			fail("Es wartet keine Runde auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			fail("Nur der Gastgeber kann die Runde starten.", http.StatusForbidden)
			return
		}
		if len(game.Players) < 2 {
			fail("31 braucht mindestens zwei Spieler am Tisch.", http.StatusConflict)
			return
		}
		for _, p := range game.Players {
			if !p.Ready {
				fail("Erst müssen alle Spieler BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.Current = 0
		game.Round = 1
		schwimmDeal(game)
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "startet die Runde 31.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "reset":
		if game == nil || game.Status != "finished" {
			fail("Ein neues Spiel ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		if schwimmSeat(game, input.ID) < 0 {
			fail("Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		fresh := &SchwimmenGame{ID: game.ID, HostID: game.HostID, Status: "invited", Phase: "lobby", Direction: 1, FinalTurns: -1, Round: 1}
		for _, p := range game.Players {
			fresh.Players = append(fresh.Players, &SchwimmPlayer{ID: p.ID, Name: p.Name, Lives: schwimmLives})
		}
		fresh.Note = "Neue Runde – alle drücken BEREIT, dann startet der Gastgeber."
		room.Schwimmen = fresh

	case "swap", "swapall", "pass", "knock", "next":
		if game == nil {
			fail("Es läuft gerade keine Partie.", http.StatusConflict)
			return
		}
		seat := schwimmSeat(game, input.ID)
		if seat < 0 {
			fail("Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}

		if action == "next" {
			if game.Phase != "between" {
				fail("Die nächste Runde ist gerade nicht möglich.", http.StatusConflict)
				return
			}
			game.Current = schwimmNext(game, game.Current)
			game.Round++
			schwimmDeal(game)
			break
		}

		if game.Status != "active" || game.Phase != "playing" {
			fail("Es läuft gerade keine Runde.", http.StatusConflict)
			return
		}
		if game.Current != seat {
			fail("Du bist nicht am Zug.", http.StatusConflict)
			return
		}
		me := game.Players[seat]

		// Drei Asse zählen sofort, egal welche Aktion gerade gewählt wurde.
		if schwimmIsFeuer(me.Hand) {
			game.schwimmFeuer(seat)
			break
		}

		switch action {
		case "swap":
			if input.HandIndex < 0 || input.HandIndex >= len(me.Hand) ||
				input.TableIndex < 0 || input.TableIndex >= len(game.Table) {
				fail("Wähle eine deiner Karten und eine Tischkarte.", http.StatusBadRequest)
				return
			}
			me.Hand[input.HandIndex], game.Table[input.TableIndex] =
				game.Table[input.TableIndex], me.Hand[input.HandIndex]
			game.Passes = 0
			game.Note = player.Name + " tauscht eine Karte."
			if schwimmIsFeuer(me.Hand) {
				game.schwimmFeuer(seat)
				break
			}
			if schwimmPoints(me.Hand) == 31 {
				game.schwimmFinishRound(player.Name + " hat 31 Punkte erreicht!")
				break
			}
			game.schwimmAdvance()

		case "swapall":
			me.Hand, game.Table = game.Table, me.Hand
			game.Passes = 0
			game.Note = player.Name + " tauscht alle drei Karten."
			if schwimmIsFeuer(me.Hand) {
				game.schwimmFeuer(seat)
				break
			}
			if schwimmPoints(me.Hand) == 31 {
				game.schwimmFinishRound(player.Name + " hat 31 Punkte erreicht!")
				break
			}
			game.schwimmAdvance()

		case "pass":
			game.Passes++
			game.Note = player.Name + " schiebt."
			if game.Passes >= schwimmActive(game) {
				neu := []SchwimmCard{}
				for i := 0; i < schwimmHandSize; i++ {
					if card, ok := schwimmDraw(game); ok {
						neu = append(neu, card)
					}
				}
				if len(neu) > 0 {
					game.Table = neu
					game.Note = "Alle schieben – neue Tischkarten."
				} else {
					game.Note = "Alle schieben – der Stapel ist leer."
				}
				game.Passes = 0
			}
			game.schwimmAdvance()

		case "knock":
			if game.FinalTurns != -1 {
				fail("Es wird schon geklopft.", http.StatusConflict)
				return
			}
			game.FinalTurns = schwimmActive(game) - 1
			game.Note = player.Name + " klopft – letzter Durchgang!"
			if game.FinalTurns == 0 {
				game.schwimmFinishRound("Die Klopfrunde ist beendet.")
				break
			}
			game.Current = schwimmNext(game, game.Current)
		}

	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	if room.Schwimmen != nil {
		room.Schwimmen.DeckCount = len(room.Schwimmen.Deck)
	}
	var view *SchwimmenView
	if room.Schwimmen != nil {
		v := room.Schwimmen.viewFor(input.ID)
		view = &v
	}
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": view})
}
