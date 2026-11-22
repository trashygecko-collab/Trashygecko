package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// UNO: 108 Karten, sieben auf die Hand, gleiche Farbe oder gleicher Wert.
// Aussetzen, Richtungswechsel, +2, Farbwahl und +4 mit Anfechten. Wer nur noch
// eine Karte hat, muss UNO rufen, sonst gibt es zwei Strafkarten.
const (
	unoMaxPlayers = 4
	unoHandSize   = 7
	// unoWindow ist die Zeit, in der UNO gerufen werden muss.
	unoWindow = 10 * time.Second
	// unoChallengeWindow: So lange darf eine +4 angefochten werden.
	unoChallengeWindow = 15 * time.Second
)

var unoColors = []string{"red", "blue", "green", "yellow"}
var unoActions = []string{"skip", "reverse", "draw2"}

type UnoCard struct {
	Color string `json:"color"`
	Value string `json:"value"`
}

type UnoPlayer struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Hand  []UnoCard `json:"-"`
	Ready bool      `json:"ready"`
}

// UnoChallenge merkt sich eine +4, die noch angefochten werden kann.
type UnoChallenge struct {
	Offender   int       `json:"offender"`
	Challenger int       `json:"challenger"`
	PrevColor  string    `json:"prevColor"`
	Until      time.Time `json:"-"`
}

type UnoGame struct {
	ID           string        `json:"id"`
	HostID       string        `json:"hostId"`
	Players      []*UnoPlayer  `json:"-"`
	Deck         []UnoCard     `json:"-"`
	Discard      []UnoCard     `json:"-"`
	Top          *UnoCard      `json:"top"`
	Current      int           `json:"current"`
	Direction    int           `json:"direction"`
	ActiveColor  string        `json:"activeColor"`
	Drawn        *UnoCard      `json:"-"`
	Phase        string        `json:"phase"`  // lobby, playing, chooseColor, challenge
	Status       string        `json:"status"` // invited, active, finished
	PendingUno   int           `json:"pendingUno"`
	UnoUntil     time.Time     `json:"-"`
	Challenge    *UnoChallenge `json:"challenge"`
	Winner       string        `json:"winner,omitempty"`
	BotID        string        `json:"botId,omitempty"`
	BotStufe     string        `json:"-"`
	BotFrist     time.Time     `json:"-"`
	Note         string        `json:"note,omitempty"`
	LastPlayed   string        `json:"lastPlayed,omitempty"`
	DeckCount    int           `json:"deckCount"`
	DiscardCount int           `json:"discardCount"`
}

type UnoSeatView struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Ready bool   `json:"ready"`
}

type UnoView struct {
	ID              string        `json:"id"`
	HostID          string        `json:"hostId"`
	Status          string        `json:"status"`
	Phase           string        `json:"phase"`
	Seats           []UnoSeatView `json:"seats"`
	Current         int           `json:"current"`
	Direction       int           `json:"direction"`
	Top             *UnoCard      `json:"top"`
	ActiveColor     string        `json:"activeColor"`
	DeckCount       int           `json:"deckCount"`
	DiscardCount    int           `json:"discardCount"`
	Drawn           *UnoCard      `json:"drawn,omitempty"`
	PendingUno      int           `json:"pendingUno"`
	Challenge       *UnoChallenge `json:"challenge"`
	Winner          string        `json:"winner,omitempty"`
	Note            string        `json:"note,omitempty"`
	LastPlayed      string        `json:"lastPlayed,omitempty"`
	You             int           `json:"you"`
	Hand            []UnoCard     `json:"hand"`
	Playable        []bool        `json:"playable"`
	MustChooseColor bool          `json:"mustChooseColor"`
	MustChallenge   bool          `json:"mustChallenge"`
}

func unoDeck() []UnoCard {
	deck := make([]UnoCard, 0, 108)
	for _, color := range unoColors {
		deck = append(deck, UnoCard{Color: color, Value: "0"})
		for n := 1; n <= 9; n++ {
			value := strconv.Itoa(n)
			deck = append(deck, UnoCard{Color: color, Value: value}, UnoCard{Color: color, Value: value})
		}
		for _, action := range unoActions {
			deck = append(deck, UnoCard{Color: color, Value: action}, UnoCard{Color: color, Value: action})
		}
	}
	for i := 0; i < 4; i++ {
		deck = append(deck, UnoCard{Color: "wild", Value: "wild"})
		deck = append(deck, UnoCard{Color: "wild", Value: "wild4"})
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

// unoDrawCard zieht eine Karte. Ist der Stapel leer, wird die Ablage bis auf
// die oberste Karte neu gemischt.
func unoDrawCard(g *UnoGame) (UnoCard, bool) {
	if len(g.Deck) == 0 {
		if len(g.Discard) <= 1 {
			return UnoCard{}, false
		}
		top := g.Discard[len(g.Discard)-1]
		rest := append([]UnoCard(nil), g.Discard[:len(g.Discard)-1]...)
		rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
		g.Deck = rest
		g.Discard = []UnoCard{top}
	}
	card := g.Deck[len(g.Deck)-1]
	g.Deck = g.Deck[:len(g.Deck)-1]
	return card, true
}

func unoNewGame(hostID, hostName, id string) *UnoGame {
	g := &UnoGame{
		ID:         id,
		HostID:     hostID,
		Status:     "invited",
		Phase:      "lobby",
		Direction:  1,
		PendingUno: -1,
	}
	g.Players = append(g.Players, &UnoPlayer{ID: hostID, Name: hostName})
	return g
}

func unoSeat(g *UnoGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// unoIDs listet die Plätze einer Runde – damit lässt sich prüfen, ob ein neues
// Spiel den Tisch übernehmen darf (gameSlotFree).
func unoIDs(g *UnoGame) []string {
	ids := make([]string, 0, len(g.Players))
	for _, p := range g.Players {
		ids = append(ids, p.ID)
	}
	return ids
}

func unoNext(g *UnoGame, seat int) int {
	n := len(g.Players)
	if n == 0 {
		return 0
	}
	return ((seat+g.Direction)%n + n) % n
}

func (g *UnoGame) refreshCounts() {
	g.DeckCount = len(g.Deck)
	g.DiscardCount = len(g.Discard)
	if len(g.Discard) > 0 {
		top := g.Discard[len(g.Discard)-1]
		g.Top = &top
	}
}

func unoPlayable(g *UnoGame, card UnoCard) bool {
	if g.Top == nil {
		return true
	}
	if card.Color == "wild" {
		return true
	}
	return card.Color == g.ActiveColor || card.Value == g.Top.Value
}

func unoGive(g *UnoGame, seat, count int) int {
	given := 0
	for i := 0; i < count; i++ {
		card, ok := unoDrawCard(g)
		if !ok {
			break
		}
		g.Players[seat].Hand = append(g.Players[seat].Hand, card)
		given++
	}
	return given
}

// unoDeal verteilt sieben Karten und deckt eine Zahlenkarte auf.
func unoDeal(g *UnoGame) {
	g.Deck = unoDeck()
	g.Discard = nil
	for _, p := range g.Players {
		p.Hand = nil
		p.Ready = false
		unoGive(g, unoSeat(g, p.ID), unoHandSize)
	}
	first, ok := unoDrawCard(g)
	for ok && first.Color == "wild" {
		g.Deck = append([]UnoCard{first}, g.Deck...)
		first, ok = unoDrawCard(g)
	}
	g.Discard = []UnoCard{first}
	g.ActiveColor = first.Color
	g.Current = 0
	g.Direction = 1
	g.Drawn = nil
	g.PendingUno = -1
	g.Challenge = nil
	g.Winner = ""
	g.Phase = "playing"
	g.Status = "active"
	g.Note = g.Players[0].Name + " beginnt."
	g.refreshCounts()
}

// unoAdvance zählt im Uhrzeigersinn bzw. gegen die Richtung weiter.
func (g *UnoGame) unoAdvance(steps int) {
	for i := 0; i < steps; i++ {
		g.Current = unoNext(g, g.Current)
	}
}

// unoExpire räumt abgelaufene Fristen ab: ein vergessenes UNO kostet zwei
// Karten, eine unbeantwortete +4 wird angenommen.
func (g *UnoGame) unoExpire(now time.Time) {
	if g.PendingUno >= 0 && !g.UnoUntil.IsZero() && now.After(g.UnoUntil) {
		seat := g.PendingUno
		g.PendingUno = -1
		given := unoGive(g, seat, 2)
		g.Note = g.Players[seat].Name + " hat UNO vergessen und zieht " + unoCountText(given) + "."
	}
	if g.Challenge != nil && !g.Challenge.Until.IsZero() && now.After(g.Challenge.Until) {
		challenger := g.Challenge.Challenger
		g.Challenge = nil
		given := unoGive(g, challenger, 4)
		g.Phase = "playing"
		g.Current = challenger
		g.unoAdvance(1)
		g.Note = g.Players[challenger].Name + " nimmt die +4 und zieht " + unoCountText(given) + "."
	}
}

func unoCountText(n int) string {
	switch n {
	case 0:
		return "keine Karte"
	case 1:
		return "eine Karte"
	case 2:
		return "zwei Karten"
	}
	return strconv.Itoa(n) + " Karten"
}

// unoEffect wendet Aussetzen, Richtungswechsel und +2 an und gibt den Zug weiter.
func (g *UnoGame) unoEffect(card UnoCard) {
	switch card.Value {
	case "reverse":
		g.Direction *= -1
		if len(g.Players) == 2 {
			g.unoAdvance(2)
		} else {
			g.unoAdvance(1)
		}
		g.Note = "Richtungswechsel – " + g.Players[g.Current].Name + " ist am Zug."
	case "skip":
		g.unoAdvance(1)
		g.Note = g.Players[g.Current].Name + " setzt aus."
		g.unoAdvance(1)
	case "draw2":
		victim := unoNext(g, g.Current)
		given := unoGive(g, victim, 2)
		g.Note = g.Players[victim].Name + " zieht " + unoCountText(given) + " und setzt aus."
		g.unoAdvance(2)
	default:
		g.unoAdvance(1)
		g.Note = g.Players[g.Current].Name + " ist am Zug."
	}
}

func (g *UnoGame) viewFor(id string) UnoView {
	g.refreshCounts()
	view := UnoView{
		ID:              g.ID,
		HostID:          g.HostID,
		Status:          g.Status,
		Phase:           g.Phase,
		Current:         g.Current,
		Direction:       g.Direction,
		Top:             g.Top,
		ActiveColor:     g.ActiveColor,
		DeckCount:       g.DeckCount,
		DiscardCount:    g.DiscardCount,
		PendingUno:      g.PendingUno,
		Challenge:       g.Challenge,
		Winner:          g.Winner,
		Note:            g.Note,
		LastPlayed:      g.LastPlayed,
		You:             unoSeat(g, id),
		Hand:            []UnoCard{},
		Playable:        []bool{},
		MustChooseColor: g.Phase == "chooseColor" && unoSeat(g, id) == g.Current,
		MustChallenge:   g.Challenge != nil && unoSeat(g, id) == g.Challenge.Challenger,
	}
	for _, p := range g.Players {
		view.Seats = append(view.Seats, UnoSeatView{ID: p.ID, Name: p.Name, Count: len(p.Hand), Ready: p.Ready})
	}
	if view.You >= 0 {
		me := g.Players[view.You]
		view.Hand = append([]UnoCard(nil), me.Hand...)
		canAct := g.Status == "active" && g.Phase == "playing" && g.Current == view.You
		for _, card := range me.Hand {
			view.Playable = append(view.Playable, canAct && unoPlayable(g, card))
		}
		if g.Current == view.You && g.Drawn != nil {
			card := *g.Drawn
			view.Drawn = &card
		}
	}
	return view
}

func (g *UnoGame) allGone(r *Room) bool {
	present := 0
	for _, p := range g.Players {
		if _, ok := r.Players[p.ID]; ok {
			present++
		}
	}
	return present == 0
}

func cloneUno(g *UnoGame) *UnoGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Deck = append([]UnoCard(nil), g.Deck...)
	c.Discard = append([]UnoCard(nil), g.Discard...)
	c.Players = make([]*UnoPlayer, len(g.Players))
	for i, p := range g.Players {
		copied := *p
		copied.Hand = append([]UnoCard(nil), p.Hand...)
		c.Players[i] = &copied
	}
	if g.Drawn != nil {
		card := *g.Drawn
		c.Drawn = &card
	}
	if g.Top != nil {
		top := *g.Top
		c.Top = &top
	}
	if g.Challenge != nil {
		ch := *g.Challenge
		c.Challenge = &ch
	}
	return &c
}

// unoAktion führt eine Spielaktion aus. Mensch (HTTP) und Bot (unoBotSchritt)
// gehen durch dieselbe Prüfung: Karte passt, Spieler ist dran, Fristen und
// Folgezustand entstehen hier und nicht im Handler. Rückgabe 0 heißt
// „ausgeführt“; sonst HTTP-Status und Meldung für den Aufrufer.
func (g *UnoGame) unoAktion(seat int, akt string, index int, color, name string) (int, string) {
	if g == nil || g.Status != "active" {
		return http.StatusConflict, "Es läuft gerade keine aktive Partie."
	}
	if seat < 0 || seat >= len(g.Players) {
		return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
	}
	if strings.TrimSpace(name) == "" {
		name = g.Players[seat].Name
	}
	switch akt {
	case "draw":
		if g.Current != seat || g.Phase != "playing" {
			return http.StatusConflict, "Du bist nicht am Zug."
		}
		if g.Drawn != nil {
			return http.StatusConflict, "Du hast schon eine Karte gezogen."
		}
		card, available := unoDrawCard(g)
		if !available {
			return http.StatusConflict, "Es sind keine Karten mehr zum Ziehen da."
		}
		g.Players[seat].Hand = append(g.Players[seat].Hand, card)
		g.Drawn = &card
		g.Note = "Karte gezogen – legen oder passen."

	case "pass":
		if g.Current != seat || g.Phase != "playing" || g.Drawn == nil {
			return http.StatusConflict, "Du hast keine gezogene Karte zum Passen."
		}
		g.Drawn = nil
		g.unoAdvance(1)
		g.Note = name + " passt – " + g.Players[g.Current].Name + " ist am Zug."

	case "play":
		if g.Current != seat || g.Phase != "playing" {
			return http.StatusConflict, "Du bist nicht am Zug."
		}
		hand := g.Players[seat].Hand
		if index < 0 || index >= len(hand) {
			return http.StatusBadRequest, "Diese Karte hast du nicht."
		}
		if g.Drawn != nil && index != len(hand)-1 {
			return http.StatusConflict, "Nach dem Ziehen kannst du nur die gezogene Karte spielen."
		}
		card := hand[index]
		if !unoPlayable(g, card) {
			return http.StatusConflict, "Diese Karte passt nicht."
		}
		g.Players[seat].Hand = append(hand[:index], hand[index+1:]...)
		g.Discard = append(g.Discard, card)
		g.Drawn = nil
		g.LastPlayed = name
		g.refreshCounts()

		if len(g.Players[seat].Hand) == 0 {
			g.Status = "finished"
			g.Phase = "finished"
			g.Winner = name
			g.PendingUno = -1
			g.Challenge = nil
			g.Note = name + " legt die letzte Karte und gewinnt!"
			break
		}

		if len(g.Players[seat].Hand) == 1 {
			g.PendingUno = seat
			g.UnoUntil = time.Now().Add(unoWindow)
			g.Note = name + " hat nur noch eine Karte – UNO rufen!"
		}

		if card.Color == "wild" {
			g.Phase = "chooseColor"
			g.Note = "Wunschfarbe wählen!"
			break
		}

		g.ActiveColor = card.Color
		g.unoEffect(card)
		if g.PendingUno == seat {
			g.Note += " UNO nicht vergessen!"
		}

	case "color":
		if g.Current != seat || g.Phase != "chooseColor" {
			return http.StatusConflict, "Du darfst gerade keine Farbe wählen."
		}
		farbe := strings.ToLower(strings.TrimSpace(color))
		valid := false
		for _, c := range unoColors {
			if c == farbe {
				valid = true
				break
			}
		}
		if !valid {
			return http.StatusBadRequest, "Unbekannte Farbe."
		}
		// Die Farbe vor dem Buben steht noch in ActiveColor, weil wilde
		// Karten sie nicht ändern – daraus wird die +4 geprüft.
		prevColor := g.ActiveColor
		g.ActiveColor = farbe
		g.Phase = "playing"

		if g.Top != nil && g.Top.Value == "wild4" {
			challenger := unoNext(g, seat)
			g.Challenge = &UnoChallenge{
				Offender:   seat,
				Challenger: challenger,
				PrevColor:  prevColor,
				Until:      time.Now().Add(unoChallengeWindow),
			}
			g.Phase = "challenge"
			g.Current = challenger
			g.Note = g.Players[challenger].Name + ": +4 annehmen oder anzweifeln?"
			break
		}
		g.unoAdvance(1)
		g.Note = "Farbe " + farbe + " – " + g.Players[g.Current].Name + " ist am Zug."

	case "uno":
		if g.PendingUno != seat {
			return http.StatusConflict, "Du musst gerade kein UNO melden."
		}
		g.PendingUno = -1
		g.Note = name + " ruft UNO!"

	case "accept4":
		ch := g.Challenge
		if ch == nil || ch.Challenger != seat {
			return http.StatusConflict, "Es liegt keine +4 für dich bereit."
		}
		given := unoGive(g, seat, 4)
		g.Challenge = nil
		g.Phase = "playing"
		g.Current = seat
		g.unoAdvance(1)
		g.Note = name + " nimmt die +4 und zieht " + unoCountText(given) + "."

	case "challenge4":
		ch := g.Challenge
		if ch == nil || ch.Challenger != seat {
			return http.StatusConflict, "Es liegt keine +4 zum Anzweifeln für dich bereit."
		}
		offender := ch.Offender
		hadColor := false
		for _, card := range g.Players[offender].Hand {
			if card.Color == ch.PrevColor {
				hadColor = true
				break
			}
		}
		g.Challenge = nil
		g.Phase = "playing"
		if hadColor {
			given := unoGive(g, offender, 4)
			g.Current = seat
			g.Note = "Anzweifeln erfolgreich! " + g.Players[offender].Name +
				" zieht " + unoCountText(given) + "."
		} else {
			given := unoGive(g, seat, 6)
			g.Current = seat
			g.unoAdvance(1)
			g.Note = "Anzweifeln fehlgeschlagen – " + name +
				" zieht " + unoCountText(given) + "."
		}
	}
	return 0, ""
}

func unoHandler(w http.ResponseWriter, r *http.Request) {
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
		if room.Uno != nil {
			// Fristen (UNO-Ruf, +4) gelten im Raum, nicht nur in der Kopie –
			// sonst zeigt der Stand eine Strafe, die nie gebucht wurde.
			room.Uno.unoExpire(time.Now())
			unoBotTreiben(room.Uno, time.Now())
		}
		game := cloneUno(room.Uno)
		if game != nil {
			game.unoExpire(time.Now())
		}
		var view *UnoView
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
		Index int    `json:"index"`
		Color string `json:"color"`
		Stufe string `json:"stufe"`
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

	game := room.Uno
	if game != nil {
		game.unoExpire(time.Now())
	}

	switch action {
	case "bot":
		stufe := botStufe(input.Stufe)
		if room.Uno != nil {
			slot := room.Uno.Status
			if slot == "spiel" {
				slot = "active"
			}
			if !room.gameSlotFree(slot, unoIDs(room.Uno)...) {
				fail("In diesem Spielraum läuft bereits eine UNO-Runde.", http.StatusConflict)
				return
			}
		}
		g := unoNewGame(input.ID, player.Name, randomCode())
		g.Players = append(g.Players, &UnoPlayer{ID: botID(1), Name: botName(1), Ready: true})
		g.Players[0].Ready = true
		g.BotID, g.BotStufe = botID(1), stufe
		unoDeal(g)
		g.Note = player.Name + " spielt gegen " + botName(1) + " (" + botStufeText(stufe) + ")."
		room.Uno = g
	case "create":
		if game != nil && game.Status != "finished" {
			fail("An diesem Tisch läuft schon eine UNO-Runde.", http.StatusConflict)
			return
		}
		room.Uno = unoNewGame(input.ID, player.Name, randomCode())
		room.Uno.Note = player.Name + " eröffnet eine UNO-Runde – bis zu 4 Spieler, mitspielen über die Spielkarte."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine UNO-Runde – bis zu 4 Spieler, mitspielen über die Spielkarte.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "join":
		if game == nil || game.Status != "invited" {
			fail("Es läuft gerade keine offene UNO-Runde.", http.StatusConflict)
			return
		}
		if unoSeat(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= unoMaxPlayers {
			fail("Diese UNO-Runde ist mit 4 Spielern voll.", http.StatusConflict)
			return
		}
		game.Players = append(game.Players, &UnoPlayer{ID: input.ID, Name: player.Name})
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zur UNO-Runde.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "ready":
		if game == nil || game.Status != "invited" {
			fail("Bereit drücken geht nur, solange die Runde noch nicht läuft.", http.StatusConflict)
			return
		}
		seat := unoSeat(game, input.ID)
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
			fail("Es wartet keine UNO-Runde auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			fail("Nur der Gastgeber kann die Runde starten.", http.StatusForbidden)
			return
		}
		if len(game.Players) < 2 {
			fail("UNO braucht mindestens zwei Spieler am Tisch.", http.StatusConflict)
			return
		}
		for _, p := range game.Players {
			if !p.Ready {
				fail("Erst müssen alle Spieler BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		unoDeal(game)
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "startet die UNO-Runde.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "reset":
		if game == nil || game.Status != "finished" {
			fail("Ein neues Spiel ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		if unoSeat(game, input.ID) < 0 {
			fail("Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		fresh := &UnoGame{ID: game.ID, HostID: game.HostID, Status: "invited", Phase: "lobby", Direction: 1, PendingUno: -1}
		for _, p := range game.Players {
			fresh.Players = append(fresh.Players, &UnoPlayer{ID: p.ID, Name: p.Name})
		}
		fresh.Note = "Neue Runde – alle drücken BEREIT, dann startet der Gastgeber."
		room.Uno = fresh

	case "draw", "play", "pass", "color", "uno", "accept4", "challenge4":
		if game == nil || game.Status != "active" {
			fail("Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		seat := unoSeat(game, input.ID)
		if seat < 0 {
			fail("Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}

		if status, meldung := game.unoAktion(seat, action, input.Index, input.Color, player.Name); status != 0 {
			fail(meldung, status)
			return
		}
	default:
		fail("Unbekannte UNO-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	if room.Uno != nil {
		room.Uno.refreshCounts()
	}
	var view *UnoView
	if room.Uno != nil {
		v := room.Uno.viewFor(input.ID)
		view = &v
	}
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": view})
}
