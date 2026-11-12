package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Arschloch (auch „Präsident"): Wer seine Karten zuerst loswird, steigt auf.
// Der Letzte wird Arschloch und muss in der nächsten Runde seine besten Karten
// abgeben. 3–4 Spieler mit einem Blatt, 5–9 mit zwei Blättern.
const (
	arschMaxPlayers = 9
	arschMinPlayers = 3
	arschMaxTrick   = 4
)

var arschSuits = []string{"♣", "♠", "♥", "♦"}

var arschValues = []struct {
	Value string
	Rank  int
}{
	{"2", 2}, {"3", 3}, {"4", 4}, {"5", 5}, {"6", 6}, {"7", 7},
	{"8", 8}, {"9", 9}, {"10", 10}, {"B", 11}, {"D", 12}, {"K", 13}, {"A", 14},
}

type ArschCard struct {
	Suit  string `json:"suit"`
	Value string `json:"value"`
	Rank  int    `json:"rank"`
}

type ArschPlayer struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Hand   []ArschCard `json:"-"`
	Active bool        `json:"active"`
	Passed bool        `json:"passed"`
	Rank   int         `json:"rank"`
	Ready  bool        `json:"ready"`
}

type ArschlochGame struct {
	ID        string         `json:"id"`
	HostID    string         `json:"hostId"`
	Players   []*ArschPlayer `json:"-"`
	Deck      []ArschCard    `json:"-"`
	Trick     []ArschCard    `json:"trick"`
	TrickRank int            `json:"trickRank"`
	TrickSeat int            `json:"trickSeat"`
	Current   int            `json:"current"`
	Round     int            `json:"round"`
	Phase     string         `json:"phase"`  // lobby, playing, between
	Status    string         `json:"status"` // invited, active
	Winner    string         `json:"winner,omitempty"`
	Note      string         `json:"note,omitempty"`
	Result    []string       `json:"result,omitempty"`
}

type ArschSeatView struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Count  int    `json:"count"`
	Active bool   `json:"active"`
	Passed bool   `json:"passed"`
	Rank   int    `json:"rank"`
	Ready  bool   `json:"ready"`
	// Cards liegt erst nach der Runde offen: Dann sieht jeder, wer was hatte.
	Cards []ArschCard `json:"cards,omitempty"`
}

type ArschlochView struct {
	ID         string          `json:"id"`
	HostID     string          `json:"hostId"`
	Status     string          `json:"status"`
	Phase      string          `json:"phase"`
	Seats      []ArschSeatView `json:"seats"`
	Current    int             `json:"current"`
	Trick      []ArschCard     `json:"trick"`
	TrickRank  int             `json:"trickRank"`
	TrickSeat  int             `json:"trickSeat"`
	Round      int             `json:"round"`
	Need       int             `json:"need"`
	Note       string          `json:"note,omitempty"`
	Result     []string        `json:"result,omitempty"`
	You        int             `json:"you"`
	Hand       []ArschCard     `json:"hand"`
	Playable   []int           `json:"playable"`
	MyTurn     bool            `json:"myTurn"`
	Deck       int             `json:"deck,omitempty"`
	RoundOver  bool            `json:"roundOver"`
	IsHost     bool            `json:"isHost"`
	MinPlayers int             `json:"minPlayers"`
	MaxPlayers int             `json:"maxPlayers"`
}

func arschRankName(rank, players int) string {
	switch {
	case rank == 1:
		return "Präsident"
	case rank == 2:
		return "Vizepräsident"
	case rank == players:
		return "Arschloch"
	case rank == players-1:
		return "Vize-Arschloch"
	}
	return "Bürger"
}

// arschDeck baut das Blatt: ein Kartenspiel bis vier Spieler, sonst zwei.
func arschDeck(players int) []ArschCard {
	packs := 1
	if players >= 5 {
		packs = 2
	}
	deck := make([]ArschCard, 0, 52*packs)
	for pack := 0; pack < packs; pack++ {
		for _, v := range arschValues {
			for _, suit := range arschSuits {
				deck = append(deck, ArschCard{Suit: suit, Value: v.Value, Rank: v.Rank})
			}
		}
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

func arschSortHand(p *ArschPlayer) {
	sort.SliceStable(p.Hand, func(i, j int) bool {
		if p.Hand[i].Rank != p.Hand[j].Rank {
			return p.Hand[i].Rank < p.Hand[j].Rank
		}
		return arschSuitIndex(p.Hand[i].Suit) < arschSuitIndex(p.Hand[j].Suit)
	})
}

func arschSuitIndex(suit string) int {
	for i, s := range arschSuits {
		if s == suit {
			return i
		}
	}
	return len(arschSuits)
}

func arschNewGame(hostID, hostName, id string) *ArschlochGame {
	g := &ArschlochGame{
		ID:        id,
		HostID:    hostID,
		Status:    "invited",
		Phase:     "lobby",
		Round:     1,
		TrickSeat: -1,
	}
	g.Players = append(g.Players, &ArschPlayer{ID: hostID, Name: hostName, Active: true})
	return g
}

func arschSeat(g *ArschlochGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func arschRanked(g *ArschlochGame) int {
	count := 0
	for _, p := range g.Players {
		if p.Rank > 0 {
			count++
		}
	}
	return count
}

func arschActiveCount(g *ArschlochGame) int {
	count := 0
	for _, p := range g.Players {
		if p.Active {
			count++
		}
	}
	return count
}

// arschNext sucht den nächsten Spieler, der noch Karten hat und nicht gepasst
// hat. -1 heißt: dieser Stich ist entschieden.
func arschNext(g *ArschlochGame, from int) int {
	n := len(g.Players)
	for step := 1; step <= n; step++ {
		i := (from + step) % n
		if g.Players[i].Active && !g.Players[i].Passed {
			return i
		}
	}
	return -1
}

// arschExchange gibt die hohen Karten nach oben: Präsident und Arschloch
// tauschen zwei Karten (bei drei Spielern eine), Vizepräsident und
// Vize-Arschloch eine.
func (g *ArschlochGame) arschExchange(oldRanks []int) {
	n := len(g.Players)
	if len(oldRanks) != n {
		return
	}
	find := func(rank int) int {
		for i, r := range oldRanks {
			if r == rank {
				return i
			}
		}
		return -1
	}

	if president, asshole := find(1), find(n); president >= 0 && asshole >= 0 {
		amount := 2
		if n == 3 {
			amount = 1
		}
		g.arschSwap(president, asshole, amount)
	}
	if n >= 4 {
		if vice, viceAss := find(2), find(n-1); vice >= 0 && viceAss >= 0 {
			g.arschSwap(vice, viceAss, 1)
		}
	}
}

// arschSwap: Der obere Platz gibt seine niedrigsten Karten ab, der untere seine
// höchsten.
func (g *ArschlochGame) arschSwap(high, low, count int) {
	a, b := g.Players[high], g.Players[low]
	for k := 0; k < count; k++ {
		arschSortHand(a)
		arschSortHand(b)
		if len(a.Hand) == 0 || len(b.Hand) == 0 {
			return
		}
		lowCard := a.Hand[0]
		highCard := b.Hand[len(b.Hand)-1]
		a.Hand = append(a.Hand[1:], highCard)
		b.Hand = append(b.Hand[:len(b.Hand)-1], lowCard)
	}
	arschSortHand(a)
	arschSortHand(b)
}

// arschDeal gibt neu, führt den Kartentausch aus und bestimmt, wer eröffnet.
func (g *ArschlochGame) arschDeal() {
	oldRanks := make([]int, len(g.Players))
	for i, p := range g.Players {
		oldRanks[i] = p.Rank
	}

	g.Deck = arschDeck(len(g.Players))
	for _, p := range g.Players {
		p.Hand = nil
		p.Active = true
		p.Passed = false
		p.Rank = 0
	}
	seat := 0
	for len(g.Deck) > 0 {
		card := g.Deck[len(g.Deck)-1]
		g.Deck = g.Deck[:len(g.Deck)-1]
		g.Players[seat%len(g.Players)].Hand = append(g.Players[seat%len(g.Players)].Hand, card)
		seat++
	}
	for _, p := range g.Players {
		arschSortHand(p)
	}

	g.Trick = nil
	g.TrickRank = 0
	g.TrickSeat = -1
	g.Result = nil

	if g.Round > 1 {
		g.arschExchange(oldRanks)
	}

	// Das Arschloch der letzten Runde darf eröffnen, in Runde 1 entscheidet das Los.
	g.Current = rand.Intn(len(g.Players))
	for i, rank := range oldRanks {
		if rank == len(g.Players) {
			g.Current = i
		}
	}
	if !g.Players[g.Current].Active {
		g.Current = arschNext(g, g.Current)
	}
	if g.Current < 0 {
		g.Current = 0
	}

	g.Phase = "playing"
	g.Status = "active"
	g.Note = g.Players[g.Current].Name + " eröffnet die Runde."
}

// arschPlayable liefert die Karten, die gerade gelegt werden dürfen.
func (g *ArschlochGame) arschPlayable(seat int) []int {
	out := []int{}
	p := g.Players[seat]
	if !p.Active || g.Phase != "playing" || seat != g.Current {
		return out
	}
	if len(g.Trick) == 0 {
		for i := range p.Hand {
			out = append(out, i)
		}
		return out
	}
	need := len(g.Trick)
	byRank := map[int][]int{}
	for i, card := range p.Hand {
		byRank[card.Rank] = append(byRank[card.Rank], i)
	}
	for rank, indices := range byRank {
		if rank > g.TrickRank && len(indices) >= need {
			out = append(out, indices...)
		}
	}
	sort.Ints(out)
	return out
}

// arschFinishTrick schließt den Stich: Wer die höchsten Karten gelegt hat,
// darf neu eröffnen.
func (g *ArschlochGame) arschFinishTrick() {
	winner := g.TrickSeat
	winnerName := ""
	if winner >= 0 && winner < len(g.Players) {
		winnerName = g.Players[winner].Name
	}
	g.Trick = nil
	g.TrickRank = 0
	g.TrickSeat = -1
	for _, p := range g.Players {
		if p.Active {
			p.Passed = false
		}
	}

	if winner >= 0 && g.Players[winner].Active {
		g.Current = winner
		g.Note = winnerName + " gewinnt den Stich und eröffnet neu."
		return
	}
	next := arschNext(g, winner)
	if next < 0 {
		g.arschEndRound()
		return
	}
	g.Current = next
	g.Note = winnerName + " ist raus – " + g.Players[next].Name + " eröffnet."
}

// arschCheckTrickEnd prüft, ob außer dem letzten Geber noch jemand spielen kann.
func (g *ArschlochGame) arschCheckTrickEnd() {
	if len(g.Trick) == 0 {
		return
	}
	for i, p := range g.Players {
		if i == g.TrickSeat {
			continue
		}
		if p.Active && !p.Passed {
			return
		}
	}
	g.arschFinishTrick()
}

// arschEndRound verteilt die letzten Plätze und zeigt die Rangliste.
func (g *ArschlochGame) arschEndRound() {
	next := arschRanked(g) + 1
	for _, p := range g.Players {
		if p.Rank == 0 {
			p.Rank = next
			next++
		}
		p.Active = false
		p.Passed = false
	}
	g.Trick = nil
	g.TrickRank = 0
	g.TrickSeat = -1
	g.Phase = "between"

	ordered := make([]*ArschPlayer, len(g.Players))
	copy(ordered, g.Players)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Rank < ordered[j].Rank })

	lines := []string{"Runde " + strconv.Itoa(g.Round) + " ist zu Ende."}
	for _, p := range ordered {
		lines = append(lines, strconv.Itoa(p.Rank)+". "+arschRankName(p.Rank, len(g.Players))+" – "+p.Name)
	}
	g.Result = lines

	last := ordered[len(ordered)-1]
	g.Winner = ordered[0].Name
	g.Note = last.Name + " ist Arschloch und gibt nächste Runde die besten Karten ab."
}

func (g *ArschlochGame) viewFor(id string) ArschlochView {
	view := ArschlochView{
		ID:         g.ID,
		HostID:     g.HostID,
		Status:     g.Status,
		Phase:      g.Phase,
		Current:    g.Current,
		Trick:      append([]ArschCard(nil), g.Trick...),
		TrickRank:  g.TrickRank,
		TrickSeat:  g.TrickSeat,
		Round:      g.Round,
		Need:       len(g.Trick),
		Note:       g.Note,
		Result:     append([]string(nil), g.Result...),
		You:        arschSeat(g, id),
		Hand:       []ArschCard{},
		Playable:   []int{},
		Deck:       len(g.Deck),
		RoundOver:  g.Phase != "playing",
		MinPlayers: arschMinPlayers,
		MaxPlayers: arschMaxPlayers,
	}
	view.IsHost = g.HostID == id

	for _, p := range g.Players {
		seat := ArschSeatView{
			ID:     p.ID,
			Name:   p.Name,
			Count:  len(p.Hand),
			Active: p.Active,
			Passed: p.Passed,
			Rank:   p.Rank,
			Ready:  p.Ready,
		}
		if g.Phase != "playing" {
			seat.Cards = append([]ArschCard(nil), p.Hand...)
		}
		view.Seats = append(view.Seats, seat)
	}

	if view.You >= 0 {
		me := g.Players[view.You]
		view.Hand = append([]ArschCard(nil), me.Hand...)
		view.Playable = g.arschPlayable(view.You)
		view.MyTurn = g.Phase == "playing" && g.Current == view.You
	}
	return view
}

func (g *ArschlochGame) allGone(r *Room) bool {
	present := 0
	for _, p := range g.Players {
		if _, ok := r.Players[p.ID]; ok {
			present++
		}
	}
	return present == 0
}

func cloneArschloch(g *ArschlochGame) *ArschlochGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Deck = append([]ArschCard(nil), g.Deck...)
	c.Trick = append([]ArschCard(nil), g.Trick...)
	c.Result = append([]string(nil), g.Result...)
	c.Players = make([]*ArschPlayer, len(g.Players))
	for i, p := range g.Players {
		copied := *p
		copied.Hand = append([]ArschCard(nil), p.Hand...)
		c.Players[i] = &copied
	}
	return &c
}

func arschlochHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneArschloch(room.Arschloch)
		var view *ArschlochView
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
		ID    string `json:"id"`
		Cards []int  `json:"cards"`
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

	game := room.Arschloch

	switch action {
	case "create":
		if game != nil {
			fail("An diesem Tisch läuft schon eine Arschloch-Runde.", http.StatusConflict)
			return
		}
		room.Arschloch = arschNewGame(input.ID, player.Name, randomCode())
		room.Arschloch.Note = player.Name + " eröffnet eine Arschloch-Runde – bis zu 9 Spieler, mitspielen über die Spielkarte."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Arschloch-Runde – bis zu 9 Spieler, mitspielen über die Spielkarte.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "join":
		if game == nil || game.Status != "invited" {
			fail("Es läuft gerade keine offene Arschloch-Runde.", http.StatusConflict)
			return
		}
		if arschSeat(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= arschMaxPlayers {
			fail("Diese Runde ist mit 9 Spielern voll.", http.StatusConflict)
			return
		}
		game.Players = append(game.Players, &ArschPlayer{ID: input.ID, Name: player.Name, Active: true})
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zur Arschloch-Runde.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "ready":
		if game == nil || game.Status != "invited" {
			fail("Bereit drücken geht nur, solange die Runde noch nicht läuft.", http.StatusConflict)
			return
		}
		seat := arschSeat(game, input.ID)
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
		if len(game.Players) < arschMinPlayers {
			fail("Arschloch braucht mindestens drei Spieler am Tisch.", http.StatusConflict)
			return
		}
		for _, p := range game.Players {
			if !p.Ready {
				fail("Erst müssen alle Spieler BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.Round = 1
		game.arschDeal()
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "startet die Arschloch-Runde.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "next":
		if game == nil {
			fail("Es läuft gerade keine Partie.", http.StatusConflict)
			return
		}
		if arschSeat(game, input.ID) < 0 {
			fail("Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Phase != "between" {
			fail("Die nächste Runde ist gerade nicht möglich.", http.StatusConflict)
			return
		}
		game.Round++
		game.arschDeal()

	case "play":
		if game == nil {
			fail("Es läuft gerade keine Partie.", http.StatusConflict)
			return
		}
		seat := arschSeat(game, input.ID)
		if seat < 0 {
			fail("Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Phase != "playing" {
			fail("Es läuft gerade keine Runde.", http.StatusConflict)
			return
		}
		if game.Current != seat {
			fail("Du bist nicht am Zug.", http.StatusConflict)
			return
		}
		me := game.Players[seat]
		if len(input.Cards) == 0 {
			fail("Wähle zuerst Karten aus.", http.StatusBadRequest)
			return
		}
		if len(input.Cards) > arschMaxTrick {
			fail("Mehr als vier gleiche Karten gehen nicht.", http.StatusBadRequest)
			return
		}
		seen := map[int]bool{}
		cards := []ArschCard{}
		for _, index := range input.Cards {
			if index < 0 || index >= len(me.Hand) || seen[index] {
				fail("Diese Karten hast du nicht auf der Hand.", http.StatusBadRequest)
				return
			}
			seen[index] = true
			cards = append(cards, me.Hand[index])
		}
		rank := cards[0].Rank
		for _, card := range cards {
			if card.Rank != rank {
				fail("Es müssen Karten mit gleichem Wert sein.", http.StatusBadRequest)
				return
			}
		}
		if len(game.Trick) > 0 {
			if len(cards) != len(game.Trick) {
				fail("Du musst genau "+strconv.Itoa(len(game.Trick))+" Karte(n) legen.", http.StatusBadRequest)
				return
			}
			if rank <= game.TrickRank {
				fail("Der Wert muss höher sein als "+game.Trick[0].Value+".", http.StatusBadRequest)
				return
			}
		}

		// Karten von der Hand nehmen – von hinten, damit die Indizes stimmen.
		sort.Sort(sort.Reverse(sort.IntSlice(input.Cards)))
		for _, index := range input.Cards {
			me.Hand = append(me.Hand[:index], me.Hand[index+1:]...)
		}
		sort.SliceStable(cards, func(i, j int) bool { return cards[i].Rank < cards[j].Rank })

		game.Trick = cards
		game.TrickRank = rank
		game.TrickSeat = seat
		me.Passed = true
		game.Note = player.Name + " legt " + cards[0].Value + "."

		if len(me.Hand) == 0 {
			me.Active = false
			me.Rank = arschRanked(game) + 1
			game.Note = player.Name + " ist fertig und wird " + arschRankName(me.Rank, len(game.Players)) + "."
		}

		if arschActiveCount(game) == 1 {
			game.arschEndRound()
			break
		}

		next := arschNext(game, seat)
		if next < 0 {
			game.arschFinishTrick()
			break
		}
		game.Current = next
		game.arschCheckTrickEnd()

	case "pass":
		if game == nil {
			fail("Es läuft gerade keine Partie.", http.StatusConflict)
			return
		}
		seat := arschSeat(game, input.ID)
		if seat < 0 {
			fail("Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Phase != "playing" {
			fail("Es läuft gerade keine Runde.", http.StatusConflict)
			return
		}
		if game.Current != seat {
			fail("Du bist nicht am Zug.", http.StatusConflict)
			return
		}
		if len(game.Trick) == 0 {
			fail("Beim Eröffnen kannst du nicht passen.", http.StatusConflict)
			return
		}
		game.Players[seat].Passed = true
		game.Note = player.Name + " passt."
		next := arschNext(game, seat)
		if next < 0 {
			game.arschFinishTrick()
			break
		}
		game.Current = next
		game.arschCheckTrickEnd()

	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	var view *ArschlochView
	if room.Arschloch != nil {
		v := room.Arschloch.viewFor(input.ID)
		view = &v
	}
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": view})
}
