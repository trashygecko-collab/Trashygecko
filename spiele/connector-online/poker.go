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

// Texas Hold'em für den Connector. Jeder startet mit 10 € (1000 Cent),
// kleine Blinds 10 ct, große 20 ct. Gesetzt wird in 10-Cent-Schritten.

const (
	pokerStartChips = 1000
	pokerKleinerBB  = 10
	pokerGrosserBB  = 20
	pokerMaxSpieler = 6
	pokerSchritt    = 10
)

type PokerSpieler struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Chips     int      `json:"chips"`
	Einsatz   int      `json:"einsatz"` // in dieser Setzrunde schon gesetzt
	Gesetzt   int      `json:"gesetzt"` // in dieser Hand insgesamt
	Hand      []string `json:"hand"`    // zwei Karten, leer bei den anderen
	Raus      bool     `json:"raus"`    // gepasst
	AllIn     bool     `json:"allIn"`
	Ready     bool     `json:"ready"`
	Gezeigt   bool     `json:"gezeigt"`   // Karten offen (Showdown)
	Gehandelt bool     `json:"gehandelt"` // in dieser Setzrunde dran gewesen
}

type PokerGame struct {
	ID        string         `json:"id"`
	HostID    string         `json:"hostId"`
	Status    string         `json:"status"` // invited, active, ende
	Spieler   []PokerSpieler `json:"spieler"`
	Tisch     []string       `json:"tisch"`
	Pot       int            `json:"pot"`
	Hoechster int            `json:"hoechster"`
	AmZug     int            `json:"amZug"`
	Runde     string         `json:"runde"` // preflop, flop, turn, river, showdown
	Hand      int            `json:"hand"`
	Dealer    int            `json:"dealer"`
	Note      string         `json:"note"`
	Gewinner  string         `json:"gewinner,omitempty"`
	Blinds    []int          `json:"blinds"`
	Deck      []string       `json:"-"`
}

// ---------------------------------------------------------------- Karten

func pokerDeck() []string {
	raenge := []string{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
	deck := make([]string, 0, 52)
	for _, farbe := range []string{"s", "h", "d", "c"} {
		for _, rang := range raenge {
			deck = append(deck, rang+farbe)
		}
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

func pokerRang(karte string) int {
	switch karte[:1] {
	case "T":
		return 10
	case "J":
		return 11
	case "Q":
		return 12
	case "K":
		return 13
	case "A":
		return 14
	}
	n, _ := strconv.Atoi(karte[:1])
	return n
}

// pokerWert5 bewertet genau fünf Karten: Kategorie und Stichwerte, gepackt.
func pokerWert5(karten []string) int64 {
	if len(karten) != 5 {
		return 0
	}
	raenge := make([]int, 0, 5)
	anzahl := map[int]int{}
	farbe := karten[0][1]
	flush := true
	for _, karte := range karten {
		rang := pokerRang(karte)
		raenge = append(raenge, rang)
		anzahl[rang]++
		if karte[1] != farbe {
			flush = false
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(raenge)))

	straight, hoch := false, 0
	if len(anzahl) == 5 {
		if raenge[0]-raenge[4] == 4 {
			straight, hoch = true, raenge[0]
		}
		if raenge[0] == 14 && raenge[1] == 5 { // Ass zählt auch unten
			straight, hoch = true, 5
		}
	}

	gruppen := make([]int, 0, 5)
	for rang := range anzahl {
		gruppen = append(gruppen, rang)
	}
	sort.Slice(gruppen, func(i, j int) bool {
		if anzahl[gruppen[i]] != anzahl[gruppen[j]] {
			return anzahl[gruppen[i]] > anzahl[gruppen[j]]
		}
		return gruppen[i] > gruppen[j]
	})

	kategorie := 0
	folge := gruppen
	switch {
	case straight && flush:
		kategorie, folge = 8, []int{hoch}
	case anzahl[gruppen[0]] == 4:
		kategorie = 7
	case anzahl[gruppen[0]] == 3 && len(gruppen) > 1 && anzahl[gruppen[1]] == 2:
		kategorie = 6
	case flush:
		kategorie = 5
	case straight:
		kategorie, folge = 4, []int{hoch}
	case anzahl[gruppen[0]] == 3:
		kategorie = 3
	case anzahl[gruppen[0]] == 2 && len(gruppen) > 1 && anzahl[gruppen[1]] == 2:
		kategorie = 2
	case anzahl[gruppen[0]] == 2:
		kategorie = 1
	}

	wert := int64(kategorie)
	for i := 0; i < 5; i++ {
		rang := 0
		if i < len(folge) {
			rang = folge[i]
		}
		wert = wert*16 + int64(rang)
	}
	return wert
}

// pokerBest sucht aus bis zu sieben Karten die beste Fünf.
func pokerBest(karten []string) int64 {
	if len(karten) < 5 {
		return 0
	}
	var best int64
	n := len(karten)
	for a := 0; a < n-4; a++ {
		for b := a + 1; b < n-3; b++ {
			for c := b + 1; c < n-2; c++ {
				for d := c + 1; d < n-1; d++ {
					for e := d + 1; e < n; e++ {
						wert := pokerWert5([]string{karten[a], karten[b], karten[c], karten[d], karten[e]})
						if wert > best {
							best = wert
						}
					}
				}
			}
		}
	}
	return best
}

func pokerName(codes []string) string {
	namen := map[byte]string{'s': "♠", 'h': "♥", 'd': "♦", 'c': "♣"}
	teile := make([]string, 0, len(codes))
	for _, karte := range codes {
		rang := karte[:1]
		if rang == "T" {
			rang = "10"
		}
		teile = append(teile, rang+namen[karte[1]])
	}
	return strings.Join(teile, " ")
}

// ---------------------------------------------------------------- Tisch

func pokerNeuerSpieler(id, name string) PokerSpieler {
	return PokerSpieler{ID: id, Name: name, Chips: pokerStartChips, Hand: []string{}}
}

func pokerSitz(g *PokerGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func pokerAktiv(g *PokerGame) int {
	n := 0
	for _, p := range g.Spieler {
		if !p.Raus {
			n++
		}
	}
	return n
}

func pokerMitChips(g *PokerGame) int {
	n := 0
	for _, p := range g.Spieler {
		if p.Chips > 0 {
			n++
		}
	}
	return n
}

// pokerNaechster sucht ab "von" den nächsten Sitz, der noch mitspielt.
func pokerNaechster(g *PokerGame, von int) int {
	n := len(g.Spieler)
	for i := 1; i <= n; i++ {
		k := (von + i) % n
		if g.Spieler[k].Chips > 0 || !g.Spieler[k].Raus {
			return k
		}
	}
	return 0
}

// pokerNaechsterZumSetzen liefert den nächsten, der noch setzen kann.
func pokerNaechsterZumSetzen(g *PokerGame, von int) int {
	n := len(g.Spieler)
	for i := 1; i <= n; i++ {
		k := (von + i) % n
		p := g.Spieler[k]
		if p.Raus || p.AllIn || p.Chips <= 0 {
			continue
		}
		if p.Einsatz < g.Hoechster || !p.Gehandelt {
			return k
		}
	}
	return -1
}

func pokerLege(g *PokerGame, sitz, betrag int) int {
	if betrag <= 0 {
		return 0
	}
	if betrag > g.Spieler[sitz].Chips {
		betrag = g.Spieler[sitz].Chips
	}
	g.Spieler[sitz].Chips -= betrag
	g.Spieler[sitz].Einsatz += betrag
	g.Spieler[sitz].Gesetzt += betrag
	if g.Spieler[sitz].Chips == 0 {
		g.Spieler[sitz].AllIn = true
	}
	if g.Spieler[sitz].Einsatz > g.Hoechster {
		g.Hoechster = g.Spieler[sitz].Einsatz
	}
	return betrag
}

// pokerNeueHand teilt aus, setzt Blinds und bestimmt, wer zuerst dran ist.
func pokerNeueHand(room *Room, g *PokerGame) {
	g.Deck = pokerDeck()
	g.Tisch = []string{}
	g.Pot = 0
	g.Hoechster = 0
	g.Hand++
	g.Runde = "preflop"
	g.Gewinner = ""

	for i := range g.Spieler {
		p := &g.Spieler[i]
		p.Hand = []string{}
		p.Einsatz = 0
		p.Gesetzt = 0
		p.Raus = p.Chips <= 0
		p.AllIn = false
		p.Gezeigt = false
		p.Gehandelt = false
	}

	// Zwei Karten, reihum.
	for runde := 0; runde < 2; runde++ {
		for sitz := 0; sitz < len(g.Spieler); sitz++ {
			k := (g.Dealer + 1 + sitz) % len(g.Spieler)
			if g.Spieler[k].Raus {
				continue
			}
			g.Spieler[k].Hand = append(g.Spieler[k].Hand, g.Deck[0])
			g.Deck = g.Deck[1:]
		}
	}

	klein := pokerNaechster(g, g.Dealer)
	gross := pokerNaechster(g, klein)
	pokerLege(g, klein, pokerKleinerBB)
	pokerLege(g, gross, pokerGrosserBB)

	g.AmZug = pokerNaechster(g, gross)
	g.Note = "Neue Hand – " + g.Spieler[klein].Name + " setzt 10 ct, " +
		g.Spieler[gross].Name + " 20 ct. " + g.Spieler[g.AmZug].Name + " ist dran."
}

// pokerWeiter sammelt die Einsätze ein und öffnet die nächste Straße.
func pokerWeiter(g *PokerGame) {
	for i := range g.Spieler {
		g.Pot += g.Spieler[i].Einsatz
		g.Spieler[i].Einsatz = 0
		g.Spieler[i].Gehandelt = false
	}
	g.Hoechster = 0

	if pokerAktiv(g) <= 1 {
		pokerShowdown(g)
		return
	}

	switch g.Runde {
	case "preflop":
		g.Runde = "flop"
		pokerGebeTisch(g, 3)
	case "flop":
		g.Runde = "turn"
		pokerGebeTisch(g, 1)
	case "turn":
		g.Runde = "river"
		pokerGebeTisch(g, 1)
	case "river":
		pokerShowdown(g)
		return
	}

	g.AmZug = pokerNaechster(g, g.Dealer)
	if g.Spieler[g.AmZug].Raus || g.Spieler[g.AmZug].AllIn {
		// Alle anderen sind all-in: es wird nicht mehr gesetzt.
		pokerWeiter(g)
		return
	}
	g.Note = "Nächste Straße: " + pokerStrassenName(g.Runde) + " – " + g.Spieler[g.AmZug].Name + " beginnt."
}

func pokerGebeTisch(g *PokerGame, anzahl int) {
	for i := 0; i < anzahl && len(g.Deck) > 0; i++ {
		g.Tisch = append(g.Tisch, g.Deck[0])
		g.Deck = g.Deck[1:]
	}
}

func pokerStrassenName(runde string) string {
	switch runde {
	case "preflop":
		return "die ersten zwei Karten"
	case "flop":
		return "der Flop"
	case "turn":
		return "der Turn"
	case "river":
		return "der River"
	}
	return runde
}

// pokerShowdown deckt auf und verteilt den Pot – mit Seitenpötten, falls
// jemand mit weniger Chips all-in gegangen ist.
func pokerShowdown(g *PokerGame) {
	for i := range g.Spieler {
		g.Pot += g.Spieler[i].Einsatz
		g.Spieler[i].Einsatz = 0
	}
	g.Runde = "showdown"

	werte := make([]int64, len(g.Spieler))
	for i := range g.Spieler {
		if g.Spieler[i].Raus {
			continue
		}
		g.Spieler[i].Gezeigt = true
		werte[i] = pokerBest(append(append([]string{}, g.Spieler[i].Hand...), g.Tisch...))
	}

	// Stufen aus den Einsätzen: erst der kleinste Beitrag, dann der nächste.
	stufen := make([]int, 0, len(g.Spieler))
	for i := range g.Spieler {
		if g.Spieler[i].Gesetzt > 0 {
			stufen = append(stufen, g.Spieler[i].Gesetzt)
		}
	}
	sort.Ints(stufen)

	gewinne := map[int]int{}
	vorher := 0
	for _, stufe := range stufen {
		if stufe <= vorher {
			continue
		}
		summe := 0
		teilnehmer := []int{}
		for i := range g.Spieler {
			if g.Spieler[i].Gesetzt < stufe {
				continue
			}
			summe += stufe - vorher
			if !g.Spieler[i].Raus {
				teilnehmer = append(teilnehmer, i)
			}
		}
		vorher = stufe
		if summe == 0 {
			continue
		}
		if len(teilnehmer) == 0 {
			continue
		}
		best := int64(-1)
		sieger := []int{}
		for _, i := range teilnehmer {
			if werte[i] > best {
				best = werte[i]
				sieger = []int{i}
			} else if werte[i] == best {
				sieger = append(sieger, i)
			}
		}
		anteil := summe / len(sieger)
		rest := summe - anteil*len(sieger)
		for k, i := range sieger {
			g.Spieler[i].Chips += anteil
			if k == 0 {
				g.Spieler[i].Chips += rest
			}
			gewinne[i] += anteil
			if k == 0 {
				gewinne[i] += rest
			}
		}
	}

	g.Pot = 0
	siegerListe := make([]int, 0, len(gewinne))
	for i := range gewinne {
		siegerListe = append(siegerListe, i)
	}
	sort.Slice(siegerListe, func(a, b int) bool { return gewinne[siegerListe[a]] > gewinne[siegerListe[b]] })
	beschreibung := make([]string, 0, len(siegerListe))
	for _, i := range siegerListe {
		beschreibung = append(beschreibung, g.Spieler[i].Name+" gewinnt "+pokerEuro(gewinne[i]))
	}
	if len(beschreibung) > 0 {
		g.Note = "Showdown – " + strings.Join(beschreibung, ", ") + "."
	} else {
		g.Note = "Showdown."
	}

	// Karten der Gewinner kurz benennen.
	for i := range g.Spieler {
		if !g.Spieler[i].Raus && !g.Spieler[i].Gezeigt {
			g.Spieler[i].Gezeigt = true
		}
	}

	if pokerMitChips(g) <= 1 {
		g.Status = "ende"
		for i := range g.Spieler {
			if g.Spieler[i].Chips > 0 {
				g.Gewinner = g.Spieler[i].Name
			}
		}
		if g.Gewinner != "" {
			g.Note += " " + g.Gewinner + " hat den Tisch abgeräumt."
		}
	}
}

func pokerEuro(cent int) string {
	return strconv.Itoa(cent/100) + "," + strconv.Itoa(100 + cent%100)[1:] + " €"
}

// pokerAnsicht lässt die Karten der anderen weg, solange sie verdeckt sind.
func pokerAnsicht(g *PokerGame, id string) *PokerGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Spieler = make([]PokerSpieler, len(g.Spieler))
	copy(c.Spieler, g.Spieler)
	c.Tisch = append([]string{}, g.Tisch...)
	c.Blinds = append([]int{}, g.Blinds...)
	for i := range c.Spieler {
		if c.Spieler[i].ID == id || c.Spieler[i].Gezeigt {
			c.Spieler[i].Hand = append([]string{}, c.Spieler[i].Hand...)
		} else {
			c.Spieler[i].Hand = []string{}
		}
	}
	return &c
}

// ---------------------------------------------------------------- Endpunkt

func pokerHandler(w http.ResponseWriter, r *http.Request) {
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
		game := pokerAnsicht(room.Poker, id)
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
		Betrag int    `json:"betrag"`
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

	zeit := time.Now().Format("15:04:05")

	switch action {
	case "create":
		if room.Poker != nil && room.Poker.Status != "ende" {
			mu.Unlock()
			http.Error(w, "An diesem Tisch läuft schon eine Partie.", http.StatusConflict)
			return
		}
		game := &PokerGame{
			ID:      randomCode(),
			HostID:  input.ID,
			Status:  "invited",
			Blinds:  []int{pokerKleinerBB, pokerGrosserBB},
			Spieler: []PokerSpieler{pokerNeuerSpieler(input.ID, player.Name)},
			Note:    "Tisch eröffnet – bis zu 6 Spieler, jeder mit 10 €.",
		}
		room.Poker = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet einen Poker-Tisch – Setzrunde ab 10 ct, mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Poker
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Poker-Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if pokerSitz(game, input.ID) >= 0 {
			break
		}
		if len(game.Spieler) >= pokerMaxSpieler {
			mu.Unlock()
			http.Error(w, "Dieser Tisch ist mit 6 Spielern voll.", http.StatusConflict)
			return
		}
		game.Spieler = append(game.Spieler, pokerNeuerSpieler(input.ID, player.Name))
		game.Note = player.Name + " setzt sich an den Tisch."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich an den Poker-Tisch.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "ready":
		game := room.Poker
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := pokerSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		game.Spieler[sitz].Ready = !game.Spieler[sitz].Ready
		if game.Spieler[sitz].Ready {
			game.Note = player.Name + " ist bereit."
		} else {
			game.Note = player.Name + " wartet noch."
		}

	case "start":
		game := room.Poker
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Partie auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber kann geben.", http.StatusForbidden)
			return
		}
		if len(game.Spieler) < 2 {
			mu.Unlock()
			http.Error(w, "Poker braucht mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, p := range game.Spieler {
			if !p.Ready {
				mu.Unlock()
				http.Error(w, "Erst müssen alle BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.Status = "active"
		game.Dealer = len(game.Spieler) - 1
		game.Hand = 0
		pokerNeueHand(room, game)
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "gibt die Karten – Poker beginnt, jeder hat 10 € vor sich.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "fold", "check", "call", "raise", "allin":
		game := room.Poker
		if game == nil || game.Status != "active" || game.Runde == "showdown" {
			mu.Unlock()
			http.Error(w, "Gerade läuft keine Setzrunde.", http.StatusConflict)
			return
		}
		sitz := pokerSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.AmZug != sitz {
			mu.Unlock()
			http.Error(w, "Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		p := &game.Spieler[sitz]

		switch action {
		case "fold":
			p.Raus = true
			p.Gehandelt = true
			game.Note = player.Name + " passt."

		case "check":
			if p.Einsatz < game.Hoechster {
				mu.Unlock()
				http.Error(w, "Zum Mitgehen fehlen dir "+pokerEuro(game.Hoechster-p.Einsatz)+".", http.StatusConflict)
				return
			}
			p.Gehandelt = true
			game.Note = player.Name + " schiebt."

		case "call":
			fehlt := game.Hoechster - p.Einsatz
			if fehlt <= 0 {
				p.Gehandelt = true
				game.Note = player.Name + " schiebt."
				break
			}
			gezahlt := pokerLege(game, sitz, fehlt)
			p.Gehandelt = true
			if p.AllIn {
				game.Note = player.Name + " geht mit " + pokerEuro(gezahlt) + " all-in."
			} else {
				game.Note = player.Name + " geht mit (" + pokerEuro(gezahlt) + ")."
			}

		case "raise", "allin":
			ziel := input.Betrag
			if action == "allin" {
				ziel = p.Einsatz + p.Chips
			}
			if ziel <= p.Einsatz {
				mu.Unlock()
				http.Error(w, "Der Einsatz muss höher sein.", http.StatusBadRequest)
				return
			}
			ziel = (ziel / pokerSchritt) * pokerSchritt
			if ziel > p.Einsatz+p.Chips {
				ziel = p.Einsatz + p.Chips
			}
			minimum := game.Hoechster + pokerSchritt
			if ziel < minimum && ziel < p.Einsatz+p.Chips {
				mu.Unlock()
				http.Error(w, "Mindestens "+pokerEuro(minimum)+" – in 10-Cent-Schritten.", http.StatusConflict)
				return
			}
			gezahlt := pokerLege(game, sitz, ziel-p.Einsatz)
			p.Gehandelt = true
			// Nach einer Erhöhung sind alle anderen wieder dran.
			for i := range game.Spieler {
				if i != sitz {
					game.Spieler[i].Gehandelt = false
				}
			}
			if p.AllIn {
				game.Note = player.Name + " geht all-in (" + pokerEuro(p.Einsatz) + ")."
			} else {
				game.Note = player.Name + " erhöht auf " + pokerEuro(p.Einsatz) + " (+" + pokerEuro(gezahlt) + ")."
			}
		}

		// Passen alle anderen, ist die Hand sofort zu Ende – auch mitten in
		// einer Setzrunde.
		naechster := pokerNaechsterZumSetzen(game, sitz)
		if pokerAktiv(game) <= 1 || naechster < 0 {
			pokerWeiter(game)
		} else {
			game.AmZug = naechster
		}

	case "naechste":
		game := room.Poker
		if game == nil || game.Status != "active" || game.Runde != "showdown" {
			mu.Unlock()
			http.Error(w, "Die laufende Hand ist noch nicht zu Ende.", http.StatusConflict)
			return
		}
		if pokerSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if pokerMitChips(game) < 2 {
			game.Status = "ende"
			game.Note = "Nur noch ein Spieler hat Chips – die Partie ist vorbei."
			break
		}
		game.Dealer = pokerNaechster(game, game.Dealer)
		pokerNeueHand(room, game)

	case "reset":
		game := room.Poker
		if game == nil || game.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Eine neue Partie geht erst nach dem Ende.", http.StatusConflict)
			return
		}
		if pokerSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		frisch := &PokerGame{
			ID:      game.ID,
			HostID:  game.HostID,
			Status:  "invited",
			Blinds:  []int{pokerKleinerBB, pokerGrosserBB},
			Spieler: []PokerSpieler{},
			Note:    "Neue Partie – jeder wieder mit 10 €. Erst BEREIT drücken.",
		}
		for _, p := range game.Spieler {
			frisch.Spieler = append(frisch.Spieler, pokerNeuerSpieler(p.ID, p.Name))
		}
		room.Poker = frisch

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Poker-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := pokerAnsicht(room.Poker, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
