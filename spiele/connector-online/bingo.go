package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Bingo: Jeder hat eine 5×5-Karte mit 24 Zahlen und einem freien Mittelfeld.
// Reihum zieht ein Spieler eine Zahl aus den 75 möglichen; alle suchen sie auf
// ihrer Karte und markieren sie. Wer zuerst eine volle Reihe, Spalte oder
// Diagonale hat und „Bingo“ ruft, gewinnt die Runde. Der Server prüft dabei,
// dass wirklich nur gezogene Zahlen markiert wurden – ein Fehlgriff kostet die
// Runde.
const (
	bingoMaxSpieler = 6
	bingoFeld       = 25
	bingoFrei       = 12 // Mittelfeld: ohne Zahl und immer markiert
	bingoSpalten    = 15 // Zahlen je Spalte: B 1–15, I 16–30, …
	bingoMaxZahl    = 75
)

var bingoBuchstaben = [5]string{"B", "I", "N", "G", "O"}

type BingoSpieler struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Ready     bool   `json:"ready"`
	Fehlgriff int    `json:"fehlgriff"`
	Verpasst  bool   `json:"verpasst"` // nach falschem Bingo für diese Runde gesperrt
	Gewonnen  bool   `json:"gewonnen"`
	Karte     []int  `json:"karte,omitempty"`    // nur die eigene Karte geht hinaus
	Markiert  []int  `json:"markiert,omitempty"` // nur die eigene Auswahl geht hinaus
	Marken    []bool `json:"-"`                  // ein Eintrag je Feld
	Anzahl    int    `json:"anzahl"`             // markierte Felder, für alle sichtbar
	Linien    int    `json:"linien"`             // volle Reihen/Spalten/Diagonalen
}

type BingoGame struct {
	ID      string         `json:"id"`
	HostID  string         `json:"hostId"`
	Status  string         `json:"status"` // invited, spiel, ende
	Zug     int            `json:"zug"`
	Runde   int            `json:"runde"`
	Gezogen []int          `json:"gezogen"`
	Rest    int            `json:"rest"` // noch nicht gezogene Zahlen
	Spieler []BingoSpieler `json:"spieler"`
	Sieger  string         `json:"sieger,omitempty"`
	Note    string         `json:"note,omitempty"`
	Letzte  []string       `json:"letzte"`
}

// karteBauen legt eine Karte an: je Spalte fünf Zahlen aus ihrem Bereich.
func karteBauen() []int {
	karte := make([]int, bingoFeld)
	for spalte := 0; spalte < 5; spalte++ {
		zahlen := make([]int, bingoSpalten)
		for i := range zahlen {
			zahlen[i] = spalte*bingoSpalten + i + 1
		}
		rand.Shuffle(len(zahlen), func(i, j int) { zahlen[i], zahlen[j] = zahlen[j], zahlen[i] })
		for zeile := 0; zeile < 5; zeile++ {
			karte[zeile*5+spalte] = zahlen[zeile]
		}
	}
	karte[bingoFrei] = 0
	return karte
}

func bingoNeueMarken() []bool {
	marken := make([]bool, bingoFeld)
	marken[bingoFrei] = true
	return marken
}

// bingoLinien zählt volle Reihen, Spalten und Diagonalen.
func bingoLinien(marken []bool) int {
	linien := 0
	for zeile := 0; zeile < 5; zeile++ {
		voll := true
		for spalte := 0; spalte < 5; spalte++ {
			if !marken[zeile*5+spalte] {
				voll = false
			}
		}
		if voll {
			linien++
		}
	}
	for spalte := 0; spalte < 5; spalte++ {
		voll := true
		for zeile := 0; zeile < 5; zeile++ {
			if !marken[zeile*5+spalte] {
				voll = false
			}
		}
		if voll {
			linien++
		}
	}
	abwaerts, aufwaerts := true, true
	for i := 0; i < 5; i++ {
		if !marken[i*5+i] {
			abwaerts = false
		}
		if !marken[i*5+(4-i)] {
			aufwaerts = false
		}
	}
	for _, voll := range []bool{abwaerts, aufwaerts} {
		if voll {
			linien++
		}
	}
	return linien
}

// bingoAuffrischen hält Anzahl und Linien je Spieler aktuell.
func (g *BingoGame) bingoAuffrischen() {
	for i := range g.Spieler {
		s := &g.Spieler[i]
		s.Anzahl = 0
		s.Markiert = nil
		for feld, marke := range s.Marken {
			if !marke {
				continue
			}
			s.Anzahl++
			if s.Karte[feld] > 0 {
				s.Markiert = append(s.Markiert, s.Karte[feld])
			}
		}
		s.Linien = bingoLinien(s.Marken)
	}
	g.Rest = bingoMaxZahl - len(g.Gezogen)
}

func bingoNeuesSpiel(hostID, hostName, id string) *BingoGame {
	return &BingoGame{
		ID: id, HostID: hostID, Status: "invited",
		// Leere Listen statt null: die Seiten rechnen mit ihnen.
		Gezogen: []int{}, Letzte: []string{}, Rest: bingoMaxZahl,
		// Auch der Gastgeber erklärt sich bereit: niemand startet im Alleingang.
		Spieler: []BingoSpieler{{ID: hostID, Name: hostName}},
	}
}

// bingoKarten verteilt neue Karten und räumt die Runde auf.
func (g *BingoGame) bingoKarten() {
	for i := range g.Spieler {
		g.Spieler[i].Karte = karteBauen()
		g.Spieler[i].Marken = bingoNeueMarken()
		g.Spieler[i].Fehlgriff = 0
		g.Spieler[i].Verpasst = false
		g.Spieler[i].Gewonnen = false
	}
	g.Gezogen = []int{}
	g.Letzte = []string{}
	g.Sieger = ""
	g.Zug = rand.Intn(len(g.Spieler))
	g.Status = "spiel"
	g.bingoAuffrischen()
}

func (g *BingoGame) sitz(id string) int {
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			return i
		}
	}
	return -1
}

func bingoZahlText(zahl int) string {
	if zahl < 1 || zahl > bingoMaxZahl {
		return strconv.Itoa(zahl)
	}
	return bingoBuchstaben[(zahl-1)/bingoSpalten] + strconv.Itoa(zahl)
}

func enthaeltZahl(zahlen []int, zahl int) bool {
	for _, z := range zahlen {
		if z == zahl {
			return true
		}
	}
	return false
}

// bingoZiehen zieht eine noch nicht gezogene Zahl. Ist keine mehr übrig,
// endet die Runde ohne Sieger.
func (g *BingoGame) bingoZiehen() (int, string) {
	offen := make([]int, 0, bingoMaxZahl)
	for zahl := 1; zahl <= bingoMaxZahl; zahl++ {
		if !enthaeltZahl(g.Gezogen, zahl) {
			offen = append(offen, zahl)
		}
	}
	if len(offen) == 0 {
		return 0, "Alle 75 Zahlen sind gezogen."
	}
	zahl := offen[rand.Intn(len(offen))]
	g.Gezogen = append(g.Gezogen, zahl)
	g.Letzte = append(g.Letzte, bingoZahlText(zahl))
	if len(g.Letzte) > 8 {
		g.Letzte = g.Letzte[len(g.Letzte)-8:]
	}
	g.Note = g.Spieler[g.Zug].Name + " zieht " + bingoZahlText(zahl) + "."
	g.bingoWeiter()
	if len(g.Gezogen) == bingoMaxZahl {
		g.bingoBeenden(-1, "Alle 75 Zahlen sind gezogen – niemand hatte Bingo.")
	}
	g.bingoAuffrischen()
	return zahl, ""
}

func (g *BingoGame) bingoWeiter() {
	g.Zug = (g.Zug + 1) % len(g.Spieler)
}

// bingoMarkieren setzt oder nimmt eine Marke. Jede Zahl der eigenen Karte ist
// erlaubt – ob sie schon gezogen wurde, prüft erst die Bingo-Meldung.
func (g *BingoGame) bingoMarkieren(sitz, zahl int) string {
	s := &g.Spieler[sitz]
	for feld, wert := range s.Karte {
		if wert != zahl {
			continue
		}
		if feld == bingoFrei {
			return "Das Mittelfeld ist frei."
		}
		s.Marken[feld] = !s.Marken[feld]
		g.bingoAuffrischen()
		return ""
	}
	return "Diese Zahl steht nicht auf deiner Karte."
}

// bingoPruefen sagt, ob die Meldung gilt: eine volle Linie und keine Marke auf
// einer Zahl, die noch nicht gezogen wurde.
func (g *BingoGame) bingoPruefen(sitz int) (bool, string) {
	s := g.Spieler[sitz]
	if bingoLinien(s.Marken) == 0 {
		return false, "Noch keine volle Reihe."
	}
	for feld, marke := range s.Marken {
		if !marke || feld == bingoFrei {
			continue
		}
		if !enthaeltZahl(g.Gezogen, s.Karte[feld]) {
			return false, bingoZahlText(s.Karte[feld]) + " wurde noch nicht gezogen."
		}
	}
	return true, ""
}

// bingoBeenden schließt die Runde; sitz < 0 heißt: niemand hat gewonnen.
func (g *BingoGame) bingoBeenden(sitz int, note string) {
	g.Status = "ende"
	g.Sieger = ""
	if sitz >= 0 {
		g.Spieler[sitz].Gewonnen = true
		g.Sieger = g.Spieler[sitz].ID
	}
	g.Note = note
	g.bingoAuffrischen()
}

// bingoAnzeige kopiert den Stand für einen Betrachter: fremde Karten und
// Marken bleiben draußen.
func bingoAnzeige(g *BingoGame, id string) *BingoGame {
	if g == nil {
		return nil
	}
	c := *g
	// append auf nil ergäbe null in der Antwort – die Seiten erwarten Listen.
	c.Gezogen = append([]int{}, g.Gezogen...)
	c.Letzte = append([]string{}, g.Letzte...)
	c.Spieler = append([]BingoSpieler(nil), g.Spieler...)
	for i := range c.Spieler {
		c.Spieler[i].Marken = nil
		if c.Spieler[i].ID != id {
			c.Spieler[i].Karte = nil
			c.Spieler[i].Markiert = nil
		}
	}
	return &c
}

func bingoHandler(w http.ResponseWriter, r *http.Request) {
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
		game := bingoAnzeige(room.Bingo, id)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID   string `json:"id"`
		Zahl int    `json:"zahl"`
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
	// Ohne Runde ist es ein Konflikt, ohne eigenen Platz eine Absage.
	amTisch := func(status, fehlt, nichtDabei string) (*BingoGame, int, string, int) {
		g := room.Bingo
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
		if room.Bingo != nil && !room.gameSlotFree(room.Bingo.Status, bingoIDs(room.Bingo)...) {
			fail("In diesem Spielraum läuft bereits eine Bingo-Runde.", http.StatusConflict)
			return
		}
		room.Bingo = bingoNeuesSpiel(input.ID, player.Name, randomCode())
		room.Bingo.Note = player.Name + " eröffnet eine Runde Bingo."
	case "join":
		g := room.Bingo
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Runde.", http.StatusConflict)
			return
		}
		if g.sitz(input.ID) >= 0 {
			fail("Du sitzt schon am Tisch.", http.StatusConflict)
			return
		}
		if len(g.Spieler) >= bingoMaxSpieler {
			fail("Der Tisch ist voll (höchstens "+strconv.Itoa(bingoMaxSpieler)+" Spieler).", http.StatusConflict)
			return
		}
		g.Spieler = append(g.Spieler, BingoSpieler{ID: input.ID, Name: player.Name})
	case "ready":
		g, sitz, problem, code := amTisch("invited", "Es gibt keine offene Runde.", "Du sitzt nicht an diesem Tisch.")
		if problem != "" {
			fail(problem, code)
			return
		}
		g.Spieler[sitz].Ready = !g.Spieler[sitz].Ready
	case "start":
		g, _, problem, code := amTisch("invited", "Es gibt keine offene Runde.", "Du sitzt nicht an diesem Tisch.")
		if problem != "" {
			fail(problem, code)
			return
		}
		if len(g.Spieler) < 2 {
			fail("Bingo braucht mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, s := range g.Spieler {
			if !s.Ready {
				fail("Noch nicht alle sind bereit.", http.StatusConflict)
				return
			}
		}
		g.Runde = 1
		g.bingoKarten()
	case "ziehen":
		g, sitz, problem, code := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, code)
			return
		}
		if sitz != g.Zug {
			fail("Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		if _, text := g.bingoZiehen(); text != "" {
			fail(text, http.StatusConflict)
			return
		}
	case "markieren":
		g, sitz, problem, code := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, code)
			return
		}
		if text := g.bingoMarkieren(sitz, input.Zahl); text != "" {
			fail(text, http.StatusConflict)
			return
		}
	case "bingo":
		g, sitz, problem, code := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, code)
			return
		}
		if g.Spieler[sitz].Verpasst {
			fail("Dein Bingo war in dieser Runde schon daneben.", http.StatusConflict)
			return
		}
		gut, grund := g.bingoPruefen(sitz)
		if gut {
			g.bingoBeenden(sitz, g.Spieler[sitz].Name+" ruft Bingo und gewinnt die Runde.")
			break
		}
		g.Spieler[sitz].Verpasst = true
		g.Spieler[sitz].Fehlgriff++
		g.Note = g.Spieler[sitz].Name + " ruft Bingo – daneben: " + grund
		offen := 0
		for _, s := range g.Spieler {
			if !s.Verpasst {
				offen++
			}
		}
		if offen == 0 {
			g.bingoBeenden(-1, "Alle Bingo-Meldungen waren daneben – die Runde ist vorbei.")
			break
		}
		g.bingoAuffrischen()
	case "naechste":
		g, _, problem, code := amTisch("ende", "Es läuft keine beendete Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, code)
			return
		}
		g.Runde++
		g.bingoKarten()
		g.Note = "Runde " + strconv.Itoa(g.Runde) + " – neue Karten."
	case "reset":
		if room.Bingo != nil && room.Bingo.HostID != input.ID {
			fail("Nur der Gastgeber kann die Runde schließen.", http.StatusForbidden)
			return
		}
		room.Bingo = nil
	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}
	room.LastActive = time.Now()
	game := bingoAnzeige(room.Bingo, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": game})
}

func bingoIDs(g *BingoGame) []string {
	ids := make([]string, 0, len(g.Spieler))
	for _, s := range g.Spieler {
		ids = append(ids, s.ID)
	}
	return ids
}

// bingoRemoveSeat nimmt einen Spieler aus der Runde (Verlassen/Abbruch).
func bingoRemoveSeat(g *BingoGame, id, name string) (bool, bool, int) {
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
	g.bingoAuffrischen()
	return true, false, len(g.Spieler)
}
