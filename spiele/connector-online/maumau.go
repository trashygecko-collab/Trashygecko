package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Mau Mau: 32 Karten, fünf auf die Hand, gleiche Farbe oder gleicher Wert
// legen. 7 lässt den Nächsten zwei Karten ziehen und aussetzen, 8 setzt den
// Nächsten aus, Bube wünscht eine Farbe. Bei der letzten Karte muss MAU gesagt
// werden, sonst gibt es zwei Strafkarten.
const (
	maumauMaxPlayers = 7
	maumauHandSize   = 5
	// maumauMauWindow ist die Zeit, in der die anderen „Erwischt!“ rufen
	// können, bevor das MAU als gesagt gilt.
	maumauMauWindow = 10 * time.Second
)

var maumauSuits = []string{"♥", "♦", "♠", "♣"}
var maumauValues = []string{"7", "8", "9", "10", "B", "D", "K", "A"}

type MauCard struct {
	Suit  string `json:"suit"`
	Value string `json:"value"`
}

type MauPlayer struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Hand  []MauCard `json:"-"`
	Ready bool      `json:"ready"`
}

type MauMauGame struct {
	ID           string       `json:"id"`
	HostID       string       `json:"hostId"`
	Players      []*MauPlayer `json:"-"`
	Deck         []MauCard    `json:"-"`
	Discard      []MauCard    `json:"-"`
	Top          *MauCard     `json:"top"`
	Current      int          `json:"current"`
	Wished       string       `json:"wished"`
	Drawn        *MauCard     `json:"-"`
	DrawnSource  string       `json:"drawnSource"`
	Phase        string       `json:"phase"`  // lobby, playing, chooseSuit
	Status       string       `json:"status"` // invited, active, finished
	PendingMau   int          `json:"pendingMau"`
	MauUntil     time.Time    `json:"-"`
	Winner       string       `json:"winner,omitempty"`
	Note         string       `json:"note,omitempty"`
	LastPlayed   string       `json:"lastPlayed,omitempty"`
	DeckCount    int          `json:"deckCount"`
	DiscardCount int          `json:"discardCount"`
}

// MauSeatView zeigt einen Mitspieler, ohne dessen Karten zu verraten.
type MauSeatView struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Ready bool   `json:"ready"`
}

// MauMauView ist der Spielstand aus Sicht eines Spielers: eigenes Blatt,
// welche Karten davon passen, sonst nur die Anzahl der Mitspieler.
type MauMauView struct {
	ID             string        `json:"id"`
	HostID         string        `json:"hostId"`
	Status         string        `json:"status"`
	Phase          string        `json:"phase"`
	Seats          []MauSeatView `json:"seats"`
	Current        int           `json:"current"`
	Top            *MauCard      `json:"top"`
	DeckCount      int           `json:"deckCount"`
	DiscardCount   int           `json:"discardCount"`
	Wished         string        `json:"wished"`
	Drawn          *MauCard      `json:"drawn,omitempty"`
	DrawnSource    string        `json:"drawnSource,omitempty"`
	PendingMau     int           `json:"pendingMau"`
	Winner         string        `json:"winner,omitempty"`
	Note           string        `json:"note,omitempty"`
	LastPlayed     string        `json:"lastPlayed,omitempty"`
	You            int           `json:"you"`
	Hand           []MauCard     `json:"hand"`
	Playable       []bool        `json:"playable"`
	MustChooseSuit bool          `json:"mustChooseSuit"`
}

func maumauDeck() []MauCard {
	deck := make([]MauCard, 0, len(maumauSuits)*len(maumauValues))
	for _, suit := range maumauSuits {
		for _, value := range maumauValues {
			deck = append(deck, MauCard{Suit: suit, Value: value})
		}
	}
	for i := len(deck) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
	return deck
}

// maumauDrawCard nimmt eine Karte vom Nachziehstapel. Ist er leer, wird die
// Ablage bis auf die oberste Karte neu gemischt.
func maumauDrawCard(g *MauMauGame) (MauCard, bool) {
	if len(g.Deck) == 0 {
		if len(g.Discard) <= 1 {
			return MauCard{}, false
		}
		top := g.Discard[len(g.Discard)-1]
		rest := append([]MauCard(nil), g.Discard[:len(g.Discard)-1]...)
		for i := len(rest) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			rest[i], rest[j] = rest[j], rest[i]
		}
		g.Deck = rest
		g.Discard = []MauCard{top}
	}
	card := g.Deck[len(g.Deck)-1]
	g.Deck = g.Deck[:len(g.Deck)-1]
	return card, true
}

func maumauNewGame(hostID, hostName, id string) *MauMauGame {
	g := &MauMauGame{
		ID:         id,
		HostID:     hostID,
		Status:     "invited",
		Phase:      "lobby",
		PendingMau: -1,
	}
	g.Players = append(g.Players, &MauPlayer{ID: hostID, Name: hostName})
	return g
}

func maumauSeat(g *MauMauGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func maumauNext(g *MauMauGame, seat int) int {
	return (seat + 1) % len(g.Players)
}

// maumauActiveSuit ist die Farbe, die gerade liegt – nach einem Buben die
// gewünschte.
func maumauActiveSuit(g *MauMauGame) string {
	if g.Wished != "" {
		return g.Wished
	}
	if g.Top != nil {
		return g.Top.Suit
	}
	return ""
}

func maumauPlayable(g *MauMauGame, card MauCard) bool {
	if g.Top == nil {
		return true
	}
	return card.Suit == maumauActiveSuit(g) || card.Value == g.Top.Value || card.Value == "B"
}

// maumauDeal gibt jedem fünf Karten und legt eine Karte auf. 7, 8 und Bube
// taugen nicht als Startkarte, weil sie sofort jemanden treffen würden.
func maumauDeal(g *MauMauGame) {
	g.Deck = maumauDeck()
	g.Discard = nil
	for _, p := range g.Players {
		p.Hand = nil
		p.Ready = false
		for i := 0; i < maumauHandSize; i++ {
			card, ok := maumauDrawCard(g)
			if !ok {
				break
			}
			p.Hand = append(p.Hand, card)
		}
	}
	first, ok := maumauDrawCard(g)
	for ok && (first.Value == "7" || first.Value == "8" || first.Value == "B") {
		g.Deck = append([]MauCard{first}, g.Deck...)
		first, ok = maumauDrawCard(g)
	}
	g.Discard = []MauCard{first}
	g.Top = &g.Discard[0]
	g.Current = 0
	g.Wished = ""
	g.Drawn = nil
	g.DrawnSource = ""
	g.PendingMau = -1
	g.Winner = ""
	g.Phase = "playing"
	g.Status = "active"
	g.Note = g.Players[0].Name + " beginnt."
	g.refreshCounts()
}

func (g *MauMauGame) refreshCounts() {
	g.DeckCount = len(g.Deck)
	g.DiscardCount = len(g.Discard)
	if len(g.Discard) > 0 {
		top := g.Discard[len(g.Discard)-1]
		g.Top = &top
	}
}

// maumauClearExpiredMau beendet das kurze Zeitfenster, in dem jemand beim
// MAU noch erwischt werden konnte.
func (g *MauMauGame) maumauClearExpiredMau(now time.Time) {
	if g.PendingMau >= 0 && !g.MauUntil.IsZero() && now.After(g.MauUntil) {
		g.PendingMau = -1
	}
}

// maumauGiveCards teilt einem Spieler Karten aus.
func maumauGiveCards(g *MauMauGame, seat, count int) int {
	given := 0
	for i := 0; i < count; i++ {
		card, ok := maumauDrawCard(g)
		if !ok {
			break
		}
		g.Players[seat].Hand = append(g.Players[seat].Hand, card)
		given++
	}
	return given
}

// maumauFinishTurn gibt den Zug weiter und wendet die Sonderregeln der 7 und
// der 8 an.
func maumauFinishTurn(g *MauMauGame, seat int, played MauCard) {
	if played.Value != "B" {
		g.Wished = ""
	}
	next := maumauNext(g, seat)
	switch played.Value {
	case "7":
		victim := g.Players[next]
		drawn := maumauGiveCards(g, next, 2)
		g.Note = victim.Name + " zieht " + maumauCountText(drawn) + " und setzt aus."
		next = maumauNext(g, next)
	case "8":
		g.Note = g.Players[next].Name + " setzt aus."
		next = maumauNext(g, next)
	default:
		g.Note = g.Players[next].Name + " ist am Zug."
	}
	g.Current = next
	g.Phase = "playing"
	g.Drawn = nil
	g.DrawnSource = ""
}

func maumauCountText(n int) string {
	switch n {
	case 0:
		return "keine Karte mehr"
	case 1:
		return "eine Karte"
	case 2:
		return "zwei Karten"
	}
	return "Karten"
}

func (g *MauMauGame) viewFor(id string) MauMauView {
	g.refreshCounts()
	view := MauMauView{
		ID:             g.ID,
		HostID:         g.HostID,
		Status:         g.Status,
		Phase:          g.Phase,
		Current:        g.Current,
		Top:            g.Top,
		DeckCount:      g.DeckCount,
		DiscardCount:   g.DiscardCount,
		Wished:         g.Wished,
		DrawnSource:    g.DrawnSource,
		PendingMau:     g.PendingMau,
		Winner:         g.Winner,
		Note:           g.Note,
		LastPlayed:     g.LastPlayed,
		You:            maumauSeat(g, id),
		Hand:           []MauCard{},
		Playable:       []bool{},
		MustChooseSuit: g.Phase == "chooseSuit" && maumauSeat(g, id) == g.Current,
	}
	for _, p := range g.Players {
		view.Seats = append(view.Seats, MauSeatView{ID: p.ID, Name: p.Name, Count: len(p.Hand), Ready: p.Ready})
	}
	if view.You >= 0 {
		me := g.Players[view.You]
		view.Hand = append([]MauCard(nil), me.Hand...)
		if g.Status == "active" && g.Current == view.You && g.Phase == "playing" {
			for _, card := range me.Hand {
				view.Playable = append(view.Playable, maumauPlayable(g, card))
			}
		} else {
			view.Playable = make([]bool, len(me.Hand))
		}
		if g.Current == view.You && g.Drawn != nil {
			card := *g.Drawn
			view.Drawn = &card
		}
	}
	return view
}

func (g *MauMauGame) allGone(r *Room) bool {
	present := 0
	for _, p := range g.Players {
		if _, ok := r.Players[p.ID]; ok {
			present++
		}
	}
	return present == 0
}

func cloneMauMau(g *MauMauGame) *MauMauGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Deck = append([]MauCard(nil), g.Deck...)
	c.Discard = append([]MauCard(nil), g.Discard...)
	c.Players = make([]*MauPlayer, len(g.Players))
	for i, p := range g.Players {
		copied := *p
		copied.Hand = append([]MauCard(nil), p.Hand...)
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
	return &c
}

func maumauHandler(w http.ResponseWriter, r *http.Request) {
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
		game := cloneMauMau(room.MauMau)
		if game != nil {
			game.maumauClearExpiredMau(time.Now())
		}
		var view *MauMauView
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
		Suit  string `json:"suit"`
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

	game := room.MauMau
	if game != nil {
		game.maumauClearExpiredMau(time.Now())
	}

	switch action {
	case "create":
		if game != nil && game.Status != "finished" {
			fail("An diesem Tisch läuft schon eine Mau-Mau-Runde.", http.StatusConflict)
			return
		}
		room.MauMau = maumauNewGame(input.ID, player.Name, randomCode())
		room.MauMau.Note = player.Name + " eröffnet eine Mau-Mau-Runde – bis zu 7 Spieler, mitspielen über die Spielkarte."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Mau-Mau-Runde – bis zu 7 Spieler, mitspielen über die Spielkarte.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "join":
		if game == nil || game.Status != "invited" {
			fail("Es läuft gerade keine offene Mau-Mau-Runde.", http.StatusConflict)
			return
		}
		if maumauSeat(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= maumauMaxPlayers {
			fail("Diese Mau-Mau-Runde ist mit 7 Spielern voll.", http.StatusConflict)
			return
		}
		game.Players = append(game.Players, &MauPlayer{ID: input.ID, Name: player.Name})
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zur Mau-Mau-Runde.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "ready":
		if game == nil || game.Status != "invited" {
			fail("Bereit drücken geht nur, solange die Runde noch nicht läuft.", http.StatusConflict)
			return
		}
		seat := maumauSeat(game, input.ID)
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
			fail("Es wartet keine Mau-Mau-Runde auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			fail("Nur der Gastgeber kann die Runde starten.", http.StatusForbidden)
			return
		}
		if len(game.Players) < 2 {
			fail("Mau Mau braucht mindestens zwei Spieler am Tisch.", http.StatusConflict)
			return
		}
		for _, p := range game.Players {
			if !p.Ready {
				fail("Erst müssen alle Spieler BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		maumauDeal(game)
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "startet die Mau-Mau-Runde.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "reset":
		if game == nil || game.Status != "finished" {
			fail("Ein neues Spiel ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		if maumauSeat(game, input.ID) < 0 {
			fail("Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		fresh := &MauMauGame{ID: game.ID, HostID: game.HostID, Status: "invited", Phase: "lobby", PendingMau: -1}
		for _, p := range game.Players {
			fresh.Players = append(fresh.Players, &MauPlayer{ID: p.ID, Name: p.Name})
		}
		fresh.Note = "Neue Runde – alle drücken BEREIT, dann startet der Gastgeber."
		room.MauMau = fresh

	case "draw", "play", "keep", "suit", "mau", "catch":
		if game == nil || game.Status != "active" {
			fail("Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		seat := maumauSeat(game, input.ID)
		if seat < 0 {
			fail("Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}

		switch action {
		case "draw":
			if game.Current != seat || game.Phase != "playing" {
				fail("Du bist nicht am Zug.", http.StatusConflict)
				return
			}
			if game.Drawn != nil {
				fail("Du hast schon eine Karte gezogen.", http.StatusConflict)
				return
			}
			card, available := maumauDrawCard(game)
			if !available {
				fail("Es sind keine Karten mehr zum Ziehen da.", http.StatusConflict)
				return
			}
			game.Players[seat].Hand = append(game.Players[seat].Hand, card)
			if maumauPlayable(game, card) {
				game.Drawn = &card
				game.DrawnSource = "deck"
				game.Note = "Karte gezogen – du kannst sie legen oder behalten."
			} else {
				game.Note = player.Name + " zieht eine Karte und kann sie nicht legen."
				maumauFinishTurn(game, seat, MauCard{Value: "?"})
			}

		case "play":
			if game.Current != seat || game.Phase != "playing" {
				fail("Du bist nicht am Zug.", http.StatusConflict)
				return
			}
			hand := game.Players[seat].Hand
			if input.Index < 0 || input.Index >= len(hand) {
				fail("Diese Karte hast du nicht.", http.StatusBadRequest)
				return
			}
			// Nach dem Ziehen darf nur die gezogene Karte gelegt werden.
			if game.Drawn != nil && input.Index != len(hand)-1 {
				fail("Nach dem Ziehen kannst du nur die gezogene Karte spielen.", http.StatusConflict)
				return
			}
			card := hand[input.Index]
			if !maumauPlayable(game, card) {
				fail("Diese Karte passt nicht.", http.StatusConflict)
				return
			}
			game.Players[seat].Hand = append(hand[:input.Index], hand[input.Index+1:]...)
			game.Discard = append(game.Discard, card)
			game.Drawn = nil
			game.DrawnSource = ""
			game.LastPlayed = player.Name
			game.refreshCounts()

			if len(game.Players[seat].Hand) == 0 {
				game.Status = "finished"
				game.Phase = "finished"
				game.Winner = player.Name
				game.PendingMau = -1
				game.Note = player.Name + " legt die letzte Karte und gewinnt!"
				break
			}
			if len(game.Players[seat].Hand) == 1 {
				game.PendingMau = seat
				game.MauUntil = time.Now().Add(maumauMauWindow)
				game.Note = player.Name + " hat nur noch eine Karte – MAU sagen!"
			}
			if card.Value == "B" {
				game.Phase = "chooseSuit"
				if game.PendingMau < 0 {
					game.Note = "Wunschfarbe wählen!"
				}
				break
			}
			maumauFinishTurn(game, seat, card)
			if game.PendingMau == seat {
				game.Note += " MAU nicht vergessen!"
			}

		case "keep":
			if game.Current != seat || game.Drawn == nil {
				fail("Du hast keine gezogene Karte zum Behalten.", http.StatusConflict)
				return
			}
			maumauFinishTurn(game, seat, MauCard{Value: "?"})

		case "suit":
			if game.Current != seat || game.Phase != "chooseSuit" {
				fail("Du darfst gerade keine Farbe wünschen.", http.StatusConflict)
				return
			}
			wish := strings.TrimSpace(input.Suit)
			valid := false
			for _, suit := range maumauSuits {
				if suit == wish {
					valid = true
					break
				}
			}
			if !valid {
				fail("Unbekannte Farbe.", http.StatusBadRequest)
				return
			}
			game.Wished = wish
			game.Phase = "playing"
			maumauFinishTurn(game, seat, MauCard{Value: "B"})
			game.Note = "Wunschfarbe " + wish + " – " + game.Note

		case "mau":
			if game.PendingMau != seat {
				fail("Du musst gerade kein MAU melden.", http.StatusConflict)
				return
			}
			game.PendingMau = -1
			game.Note = player.Name + " ruft MAU!"

		case "catch":
			if game.PendingMau < 0 {
				fail("Es gibt gerade nichts zu erwischen.", http.StatusConflict)
				return
			}
			if game.PendingMau == seat {
				fail("Du kannst dich nicht selbst erwischen.", http.StatusConflict)
				return
			}
			offender := game.Players[game.PendingMau]
			drawn := maumauGiveCards(game, game.PendingMau, 2)
			game.PendingMau = -1
			game.Note = player.Name + " erwischt " + offender.Name + " – " + maumauCountText(drawn) + " Strafkarten."
		}

	default:
		fail("Unbekannte Mau-Mau-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	if room.MauMau != nil {
		room.MauMau.refreshCounts()
	}
	var view *MauMauView
	if room.MauMau != nil {
		v := room.MauMau.viewFor(input.ID)
		view = &v
	}
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": view})
}
