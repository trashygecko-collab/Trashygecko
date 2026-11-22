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

	// Bots am Tisch (action=bot). Sie sitzen wie Menschen, ziehen aber durch
	// arschAktion und werden beim Abholen des Stands angetrieben.
	BotID    string    `json:"botId,omitempty"`
	Bots     []string  `json:"-"`
	BotStufe string    `json:"botStufe,omitempty"`
	BotFrist time.Time `json:"-"`
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
	BotID      string          `json:"botId,omitempty"`
	BotStufe   string          `json:"botStufe,omitempty"`
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

// arschWeiter gibt den Zug an den Nächsten, der noch mitspielt. Ist keiner
// mehr da, endet der Stich.
func (g *ArschlochGame) arschWeiter(seat int) {
	next := arschNext(g, seat)
	if next < 0 {
		g.arschFinishTrick()
		return
	}
	g.Current = next
	g.arschCheckTrickEnd()
}

// arschAktion ist die eine Prüfung für Mensch und Bot: Karten legen (indices)
// oder passen (pass). Rückgabe ist die Meldung ("" heißt ok) und der
// HTTP-Status für den Menschen.
func (g *ArschlochGame) arschAktion(seat int, indices []int, pass bool, name string) (string, int) {
	if g == nil {
		return "Es läuft gerade keine Partie.", http.StatusConflict
	}
	if seat < 0 || seat >= len(g.Players) {
		return "Du gehörst nicht zu dieser Partie.", http.StatusForbidden
	}
	if g.Phase != "playing" {
		return "Es läuft gerade keine Runde.", http.StatusConflict
	}
	if g.Current != seat {
		return "Du bist nicht am Zug.", http.StatusConflict
	}
	me := g.Players[seat]

	if pass {
		if len(g.Trick) == 0 {
			return "Beim Eröffnen kannst du nicht passen.", http.StatusConflict
		}
		me.Passed = true
		g.Note = name + " passt."
		g.arschWeiter(seat)
		return "", 0
	}

	if len(indices) == 0 {
		return "Wähle zuerst Karten aus.", http.StatusBadRequest
	}
	if len(indices) > arschMaxTrick {
		return "Mehr als vier gleiche Karten gehen nicht.", http.StatusBadRequest
	}
	seen := map[int]bool{}
	cards := []ArschCard{}
	for _, index := range indices {
		if index < 0 || index >= len(me.Hand) || seen[index] {
			return "Diese Karten hast du nicht auf der Hand.", http.StatusBadRequest
		}
		seen[index] = true
		cards = append(cards, me.Hand[index])
	}
	rank := cards[0].Rank
	for _, card := range cards {
		if card.Rank != rank {
			return "Es müssen Karten mit gleichem Wert sein.", http.StatusBadRequest
		}
	}
	if len(g.Trick) > 0 {
		if len(cards) != len(g.Trick) {
			return "Du musst genau " + strconv.Itoa(len(g.Trick)) + " Karte(n) legen.", http.StatusBadRequest
		}
		if rank <= g.TrickRank {
			return "Der Wert muss höher sein als " + g.Trick[0].Value + ".", http.StatusBadRequest
		}
	}

	// Karten von der Hand nehmen – von hinten, damit die Indizes stimmen.
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))
	for _, index := range indices {
		me.Hand = append(me.Hand[:index], me.Hand[index+1:]...)
	}
	sort.SliceStable(cards, func(i, j int) bool { return cards[i].Rank < cards[j].Rank })

	g.Trick = cards
	g.TrickRank = rank
	g.TrickSeat = seat
	me.Passed = true
	g.Note = name + " legt " + cards[0].Value + "."

	if len(me.Hand) == 0 {
		me.Active = false
		me.Rank = arschRanked(g) + 1
		g.Note = name + " ist fertig und wird " + arschRankName(me.Rank, len(g.Players)) + "."
	}

	if arschActiveCount(g) == 1 {
		g.arschEndRound()
		return "", 0
	}
	g.arschWeiter(seat)
	return "", 0
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
	view.BotID = g.BotID
	view.BotStufe = g.BotStufe

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
	c.Bots = append([]string(nil), g.Bots...)
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
		if room.Arschloch != nil {
			arschBotTreiben(room.Arschloch, time.Now())
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
		Stufe string `json:"stufe"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.ID = kennung(r)
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
	case "bot":
		room.botNamenNeu()
		stufe := botStufe(input.Stufe)
		if room.Arschloch != nil && !room.gameSlotFree(room.Arschloch.Status, arschIDs(room.Arschloch)...) {
			fail("In diesem Spielraum läuft bereits eine Arschloch-Runde.", http.StatusConflict)
			return
		}
		g := arschNewGame(input.ID, player.Name, randomCode())
		g.Players[0].Ready = true
		for n := 1; n <= 2; n++ {
			g.Players = append(g.Players, &ArschPlayer{ID: botID(n), Name: room.botName(n), Active: true, Ready: true})
			g.Bots = append(g.Bots, botID(n))
		}
		g.BotID, g.BotStufe = botID(1), stufe
		g.Round = 1
		g.arschDeal()
		g.Note = player.Name + " spielt gegen " + room.botName(1) + " und " + room.botName(2) + " (" + botStufeText(stufe) + ")."
		room.Arschloch = g
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Arschloch gegen zwei Bots (" + botStufeText(stufe) + ").",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})
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
		game.BotFrist = time.Time{}
		game.arschDeal()

	case "play", "pass":
		if game == nil {
			fail("Es läuft gerade keine Partie.", http.StatusConflict)
			return
		}
		seat := arschSeat(game, input.ID)
		if seat < 0 {
			fail("Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		msg, status := game.arschAktion(seat, input.Cards, action == "pass", player.Name)
		if status != 0 {
			fail(msg, status)
			return
		}

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
