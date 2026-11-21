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

// Domino: Zwei Spieler, je sieben Steine aus dem Doppelsechs-Satz (28 Steine),
// der Rest liegt im Vorrat. Wer nicht legen kann, zieht einen Stein; ist der
// Vorrat leer, wird gepasst. Wer zuerst alle Steine los ist, bekommt die Augen
// des Gegners als Punkte. Können beide nicht mehr legen, gewinnt die kleinere
// Augensumme.
const (
	dominoMaxSpieler = 2
	dominoHand       = 7
	dominoAugen      = 6
)

type DominoSpieler struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Ready  bool     `json:"ready"`
	Punkte int      `json:"punkte"`
	Runden int      `json:"runden"`
	Hand   [][2]int `json:"hand,omitempty"` // nur die eigene Hand wird ausgegeben
	Anzahl int      `json:"anzahl"`         // Steine auf der Hand
}

type DominoGame struct {
	ID      string          `json:"id"`
	HostID  string          `json:"hostId"`
	Status  string          `json:"status"` // invited, spiel, ende
	Runde   int             `json:"runde"`
	Kette   [][2]int        `json:"kette"` // in Anlege-Reihenfolge, links nach rechts
	Enden   [2]int          `json:"enden"` // offene Werte links/rechts, -1 wenn leer
	Talon   int             `json:"talon"` // Steine im Vorrat
	Zug     int             `json:"zug"`   // Sitz, der legen muss
	Spieler []DominoSpieler `json:"spieler"`
	Sieger  string          `json:"sieger,omitempty"`
	Note    string          `json:"note,omitempty"`

	vorrat [][2]int // der echte Vorrat bleibt auf dem Server
}

// dominoSatz liefert alle 28 Steine des Doppelsechs-Satzes.
func dominoSatz() [][2]int {
	satz := make([][2]int, 0, 28)
	for a := 0; a <= dominoAugen; a++ {
		for b := a; b <= dominoAugen; b++ {
			satz = append(satz, [2]int{a, b})
		}
	}
	return satz
}

func dominoNeuesSpiel(hostID, hostName, id string) *DominoGame {
	return &DominoGame{
		ID: id, HostID: hostID, Status: "invited",
		Kette: [][2]int{}, Enden: [2]int{-1, -1},
		// Auch der Gastgeber erklärt sich bereit.
		Spieler: []DominoSpieler{{ID: hostID, Name: hostName}},
	}
}

func (g *DominoGame) sitz(id string) int {
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			return i
		}
	}
	return -1
}

// dominoRunde teilt neu aus und setzt die Kette zurück. Punkte bleiben stehen.
func (g *DominoGame) dominoRunde() {
	satz := dominoSatz()
	rand.Shuffle(len(satz), func(i, j int) { satz[i], satz[j] = satz[j], satz[i] })
	g.Kette = [][2]int{}
	g.Enden = [2]int{-1, -1}
	g.Sieger = ""
	g.Zug = 0
	for i := range g.Spieler {
		hand := append([][2]int{}, satz[i*dominoHand:(i+1)*dominoHand]...)
		sort.Slice(hand, func(a, b int) bool {
			if hand[a][0] != hand[b][0] {
				return hand[a][0] < hand[b][0]
			}
			return hand[a][1] < hand[b][1]
		})
		g.Spieler[i].Hand = hand
	}
	g.vorrat = append([][2]int{}, satz[dominoMaxSpieler*dominoHand:]...)
	g.Talon = len(g.vorrat)
	g.Status = "spiel"
	g.Note = "Wer zuerst alle Steine los ist, bekommt die Augen des Gegners als Punkte."
}

// dominoAugen zählt die Punkte auf der Hand eines Sitzes.
func (g *DominoGame) dominoAugen(sitz int) int {
	summe := 0
	for _, stein := range g.Spieler[sitz].Hand {
		summe += stein[0] + stein[1]
	}
	return summe
}

// dominoPasstAn meldet, ob ein Stein an eine der offenen Seiten passt.
func (g *DominoGame) dominoPasstAn(stein [2]int) bool {
	if len(g.Kette) == 0 {
		return true
	}
	for _, ende := range g.Enden {
		if stein[0] == ende || stein[1] == ende {
			return true
		}
	}
	return false
}

// dominoKannLegen meldet, ob der Sitz überhaupt einen passenden Stein hat.
func (g *DominoGame) dominoKannLegen(sitz int) bool {
	for _, stein := range g.Spieler[sitz].Hand {
		if g.dominoPasstAn(stein) {
			return true
		}
	}
	return false
}

func (g *DominoGame) handIndex(sitz int, stein [2]int) int {
	for i, k := range g.Spieler[sitz].Hand {
		if (k[0] == stein[0] && k[1] == stein[1]) || (k[0] == stein[1] && k[1] == stein[0]) {
			return i
		}
	}
	return -1
}

func (g *DominoGame) handNehmen(sitz, index int) {
	hand := g.Spieler[sitz].Hand
	g.Spieler[sitz].Hand = append(hand[:index], hand[index+1:]...)
	g.Spieler[sitz].Anzahl = len(g.Spieler[sitz].Hand)
}

func (g *DominoGame) naechsterZug() {
	if len(g.Spieler) == 0 {
		return
	}
	g.Zug = (g.Zug + 1) % len(g.Spieler)
}

// dominoBeenden schließt die Runde ab. gewinner < 0 heißt: gesperrt, es
// entscheidet die kleinere Augensumme.
func (g *DominoGame) dominoBeenden(gewinner int, note string) {
	g.Status = "ende"
	g.Sieger = ""
	if gewinner < 0 {
		best, bestAugen := -1, 0
		gleich := false
		for i := range g.Spieler {
			augen := g.dominoAugen(i)
			switch {
			case best < 0 || augen < bestAugen:
				best, bestAugen, gleich = i, augen, false
			case augen == bestAugen:
				gleich = true
			}
		}
		if best >= 0 && !gleich {
			gewinner = best
		}
		if gleich {
			note += " Es bleibt unentschieden."
			g.Note = note
			return
		}
	}
	if gewinner < 0 {
		g.Note = note
		return
	}
	punkte := 0
	for i := range g.Spieler {
		if i != gewinner {
			punkte += g.dominoAugen(i)
		}
	}
	g.Spieler[gewinner].Punkte += punkte
	g.Spieler[gewinner].Runden++
	g.Sieger = g.Spieler[gewinner].ID
	g.Note = note + " " + g.Spieler[gewinner].Name + " bekommt " + strconv.Itoa(punkte) + " Punkte."
}

// dominoLegen prüft einen Zug und legt den Stein an die gewählte Seite.
func (g *DominoGame) dominoLegen(sitz int, stein [2]int, seite string) string {
	if g.Status != "spiel" {
		return "Gerade läuft keine Runde."
	}
	if sitz != g.Zug {
		return "Du bist nicht an der Reihe."
	}
	index := g.handIndex(sitz, stein)
	if index < 0 {
		return "Dieser Stein ist nicht auf deiner Hand."
	}
	leer := len(g.Kette) == 0
	passtLinks := leer || stein[0] == g.Enden[0] || stein[1] == g.Enden[0]
	passtRechts := leer || stein[0] == g.Enden[1] || stein[1] == g.Enden[1]
	seite = strings.ToLower(strings.TrimSpace(seite))
	switch seite {
	case "links":
		if !passtLinks {
			return "Dieser Stein passt nicht an die linke Seite."
		}
	case "rechts":
		if !passtRechts {
			return "Dieser Stein passt nicht an die rechte Seite."
		}
	default:
		switch {
		case passtLinks && passtRechts:
			seite = "rechts"
		case passtLinks:
			seite = "links"
		case passtRechts:
			seite = "rechts"
		default:
			return "Dieser Stein passt an keine Seite."
		}
	}
	// Anlegen: links liegt der Stein vor der Kette, rechts dahinter. Die
	// passende Hälfte zeigt immer zur Kette.
	if leer {
		g.Kette = [][2]int{{stein[0], stein[1]}}
		g.Enden = [2]int{stein[0], stein[1]}
	} else if seite == "links" {
		// Die rechte Hälfte muss an die bisherige linke Kante zeigen.
		neu := [2]int{stein[0], stein[1]}
		if stein[1] != g.Enden[0] {
			neu = [2]int{stein[1], stein[0]}
		}
		g.Kette = append([][2]int{neu}, g.Kette...)
		g.Enden[0] = neu[0]
	} else {
		neu := [2]int{stein[0], stein[1]}
		if neu[0] != g.Enden[1] {
			neu = [2]int{neu[1], neu[0]}
		}
		g.Kette = append(g.Kette, neu)
		g.Enden[1] = neu[1]
	}
	g.handNehmen(sitz, index)
	g.Note = g.Spieler[sitz].Name + " legt " + dominoText(stein) + "."
	if len(g.Spieler[sitz].Hand) == 0 {
		g.dominoBeenden(sitz, g.Spieler[sitz].Name+" hat alle Steine weg.")
		return ""
	}
	g.naechsterZug()
	return ""
}

func dominoText(stein [2]int) string {
	return "[" + strconv.Itoa(stein[0]) + "|" + strconv.Itoa(stein[1]) + "]"
}

// dominoZiehen nimmt einen Stein aus dem Vorrat. Passt er nicht, geht der Zug
// weiter.
func (g *DominoGame) dominoZiehen(sitz int) string {
	if g.Status != "spiel" {
		return "Gerade läuft keine Runde."
	}
	if sitz != g.Zug {
		return "Du bist nicht an der Reihe."
	}
	if g.dominoKannLegen(sitz) {
		return "Du kannst noch legen."
	}
	if len(g.vorrat) == 0 {
		return "Der Vorrat ist leer – bitte passen."
	}
	stein := g.vorrat[0]
	g.vorrat = g.vorrat[1:]
	g.Talon = len(g.vorrat)
	hand := append(g.Spieler[sitz].Hand, stein)
	sort.Slice(hand, func(a, b int) bool {
		if hand[a][0] != hand[b][0] {
			return hand[a][0] < hand[b][0]
		}
		return hand[a][1] < hand[b][1]
	})
	g.Spieler[sitz].Hand = hand
	g.Spieler[sitz].Anzahl = len(hand)
	if g.dominoPasstAn(stein) {
		g.Note = g.Spieler[sitz].Name + " zieht " + dominoText(stein) + " – der Stein passt."
		return ""
	}
	g.Note = g.Spieler[sitz].Name + " zieht " + dominoText(stein) + " – der Stein passt nicht, der Zug geht weiter."
	g.naechsterZug()
	return ""
}

// dominoPassen gibt den Zug weiter, wenn kein Stein passt und der Vorrat leer
// ist. Sind danach beide blockiert, entscheidet die Augensumme.
func (g *DominoGame) dominoPassen(sitz int) string {
	if g.Status != "spiel" {
		return "Gerade läuft keine Runde."
	}
	if sitz != g.Zug {
		return "Du bist nicht an der Reihe."
	}
	if g.dominoKannLegen(sitz) {
		return "Du kannst noch legen."
	}
	if len(g.vorrat) > 0 {
		return "Du musst erst ziehen."
	}
	g.Note = g.Spieler[sitz].Name + " kann nicht legen und passt."
	g.naechsterZug()
	blockiert := true
	for i := range g.Spieler {
		if g.dominoKannLegen(i) {
			blockiert = false
			break
		}
	}
	if blockiert {
		g.dominoBeenden(-1, "Niemand kann mehr legen.")
	}
	return ""
}

func dominoAnzeige(g *DominoGame, id string) *DominoGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Kette = append([][2]int{}, g.Kette...)
	c.Spieler = make([]DominoSpieler, len(g.Spieler))
	for i := range g.Spieler {
		s := g.Spieler[i]
		s.Anzahl = len(s.Hand)
		if s.ID == id {
			s.Hand = append([][2]int{}, s.Hand...)
		} else {
			s.Hand = nil
		}
		c.Spieler[i] = s
	}
	return &c
}

func dominoHandler(w http.ResponseWriter, r *http.Request) {
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
		game := dominoAnzeige(room.Domino, id)
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
		Stein [2]int `json:"stein"`
		Seite string `json:"seite"`
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
	fail := func(text string, status int) {
		mu.Unlock()
		http.Error(w, text, status)
	}
	// amTisch liefert die Runde in der erwarteten Lage und den eigenen Platz.
	amTisch := func(status, fehlt, nichtDabei string) (*DominoGame, int, string, int) {
		g := room.Domino
		if g == nil || g.Status != status {
			return nil, -1, fehlt, http.StatusConflict
		}
		sitz := g.sitz(input.ID)
		if sitz < 0 {
			return nil, -1, nichtDabei, http.StatusForbidden
		}
		return g, sitz, "", 0
	}

	switch action {
	case "create":
		if room.Domino != nil {
			slot := room.Domino.Status
			if slot == "spiel" {
				slot = "active"
			}
			if !room.gameSlotFree(slot, dominoIDs(room.Domino)...) {
				fail("In diesem Spielraum läuft bereits eine Runde Domino.", http.StatusConflict)
				return
			}
		}
		room.Domino = dominoNeuesSpiel(input.ID, player.Name, randomCode())
		room.Domino.Note = player.Name + " eröffnet eine Runde Domino."
	case "join":
		g := room.Domino
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Runde.", http.StatusConflict)
			return
		}
		if g.sitz(input.ID) >= 0 {
			fail("Du sitzt schon am Tisch.", http.StatusConflict)
			return
		}
		if len(g.Spieler) >= dominoMaxSpieler {
			fail("Der Tisch ist voll (höchstens "+strconv.Itoa(dominoMaxSpieler)+" Spieler).", http.StatusConflict)
			return
		}
		g.Spieler = append(g.Spieler, DominoSpieler{ID: input.ID, Name: player.Name})
	case "ready":
		g, sitz, problem, status := amTisch("invited", "Es gibt keine offene Runde.", "Du sitzt nicht an diesem Tisch.")
		if problem != "" {
			fail(problem, status)
			return
		}
		g.Spieler[sitz].Ready = !g.Spieler[sitz].Ready
	case "start":
		g, _, problem, status := amTisch("invited", "Es gibt keine offene Runde.", "Du sitzt nicht an diesem Tisch.")
		if problem != "" {
			fail(problem, status)
			return
		}
		if len(g.Spieler) < 2 {
			fail("Domino braucht zwei Spieler.", http.StatusConflict)
			return
		}
		for _, s := range g.Spieler {
			if !s.Ready {
				fail("Noch nicht alle sind bereit.", http.StatusConflict)
				return
			}
		}
		g.Runde = 1
		g.dominoRunde()
	case "legen":
		g, sitz, problem, status := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, status)
			return
		}
		if text := g.dominoLegen(sitz, input.Stein, input.Seite); text != "" {
			fail(text, http.StatusConflict)
			return
		}
	case "ziehen":
		g, sitz, problem, status := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, status)
			return
		}
		if text := g.dominoZiehen(sitz); text != "" {
			fail(text, http.StatusConflict)
			return
		}
	case "passen":
		g, sitz, problem, status := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, status)
			return
		}
		if text := g.dominoPassen(sitz); text != "" {
			fail(text, http.StatusConflict)
			return
		}
	case "naechste":
		g, _, problem, status := amTisch("ende", "Es läuft keine beendete Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, status)
			return
		}
		g.Runde++
		g.dominoRunde()
	case "reset":
		if room.Domino != nil && room.Domino.HostID != input.ID {
			fail("Nur der Gastgeber kann die Runde schließen.", http.StatusForbidden)
			return
		}
		room.Domino = nil
	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}
	room.LastActive = time.Now()
	game := dominoAnzeige(room.Domino, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": game})
}

func dominoIDs(g *DominoGame) []string {
	ids := make([]string, 0, len(g.Spieler))
	for _, s := range g.Spieler {
		ids = append(ids, s.ID)
	}
	return ids
}

// dominoRemoveSeat nimmt einen Spieler aus der Runde (Verlassen/Abbruch).
func dominoRemoveSeat(g *DominoGame, id, name string) (bool, bool, int) {
	sitz := g.sitz(id)
	if sitz < 0 {
		return false, false, len(g.Spieler)
	}
	g.Spieler = append(g.Spieler[:sitz], g.Spieler[sitz+1:]...)
	if len(g.Spieler) < 2 {
		return true, true, len(g.Spieler)
	}
	if g.HostID == id {
		g.HostID = g.Spieler[0].ID
	}
	if g.Zug > sitz {
		g.Zug--
	}
	if g.Zug >= len(g.Spieler) {
		g.Zug = 0
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Spieler)) + " Spielern weiter."
	return true, false, len(g.Spieler)
}
