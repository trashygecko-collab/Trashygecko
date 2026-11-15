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

// Cabo nach den Regeln von Smiling Monster: 4 Karten vor jedem, zwei davon
// am Rundenstart anschauen, dann reihum tauschen, ziehen oder Cabo ansagen.
// Wer am Ende die wenigsten Punkte hat, gewinnt die Runde.

const (
	caboMaxSpieler = 4
	caboKarten     = 4
	caboStartBlick = 2
	caboZiel       = 101
)

type CaboSpieler struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Karten   []int          `json:"-"`
	Bekannt  map[string]int `json:"-"`
	Punkte   int            `json:"punkte"`
	Runden   int            `json:"runden"`
	Ready    bool           `json:"ready"`
	Geschaut int            `json:"geschaut"`
}

type CaboGame struct {
	ID           string        `json:"id"`
	HostID       string        `json:"hostId"`
	Status       string        `json:"status"` // invited, anschauen, active, wertung, ende
	Runde        int           `json:"runde"`
	Zug          int           `json:"zug"`
	Start        int           `json:"start"`
	Spieler      []CaboSpieler `json:"spieler"`
	Stapel       []int         `json:"-"`
	Ablage       []int         `json:"ablage"`
	Gezogen      *int          `json:"gezogen"`
	Quelle       string        `json:"quelle"`
	CaboVon      int           `json:"caboVon"`
	RestZuege    int           `json:"restZuege"`
	Note         string        `json:"note"`
	Sieger       string        `json:"sieger,omitempty"`
	RundenSieger int           `json:"rundenSieger"`
	RundenPunkte []int         `json:"rundenPunkte"`
}

type CaboSpielerSicht struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Anzahl   int            `json:"anzahl"`
	Bekannt  map[string]int `json:"bekannt"`
	Punkte   int            `json:"punkte"`
	Runden   int            `json:"runden"`
	Ready    bool           `json:"ready"`
	Geschaut int            `json:"geschaut"`
	Werte    []int          `json:"werte,omitempty"`
}

type CaboAnsicht struct {
	ID           string             `json:"id"`
	HostID       string             `json:"hostId"`
	Status       string             `json:"status"`
	Runde        int                `json:"runde"`
	Zug          int                `json:"zug"`
	Start        int                `json:"start"`
	Spieler      []CaboSpielerSicht `json:"spieler"`
	AblageOben   *int               `json:"ablageOben"`
	AblageAnzahl int                `json:"ablageAnzahl"`
	Stapel       int                `json:"stapel"`
	Gezogen      *int               `json:"gezogen"`
	Quelle       string             `json:"quelle"`
	CaboVon      int                `json:"caboVon"`
	RestZuege    int                `json:"restZuege"`
	Note         string             `json:"note"`
	Sieger       string             `json:"sieger,omitempty"`
	RundenPunkte []int              `json:"rundenPunkte"`
}

// ---------------------------------------------------------------- Aufbau

func caboDeck() []int {
	deck := []int{0, 0}
	for wert := 1; wert <= 12; wert++ {
		for i := 0; i < 4; i++ {
			deck = append(deck, wert)
		}
	}
	deck = append(deck, 13, 13)
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

func caboNeuerSpieler(id, name string) CaboSpieler {
	return CaboSpieler{ID: id, Name: name, Karten: []int{}, Bekannt: map[string]int{}}
}

func caboNeueRunde(g *CaboGame) {
	deck := caboDeck()
	for i := range g.Spieler {
		p := &g.Spieler[i]
		p.Karten = make([]int, 0, caboKarten)
		for k := 0; k < caboKarten && len(deck) > 0; k++ {
			p.Karten = append(p.Karten, deck[0])
			deck = deck[1:]
		}
		p.Bekannt = map[string]int{}
		p.Geschaut = 0
	}
	g.Stapel = deck
	g.Ablage = []int{g.Stapel[0]}
	g.Stapel = g.Stapel[1:]
	g.Gezogen = nil
	g.Quelle = ""
	g.CaboVon = -1
	g.RestZuege = 0
	g.RundenPunkte = nil
	g.Zug = g.Start
	g.Runde++
	g.Status = "anschauen"
	g.Note = "Runde " + strconv.Itoa(g.Runde) + ": Jeder schaut sich zwei eigene Karten an."
}

// ---------------------------------------------------------------- Sätze

func caboHatWert(karten []int, wert int) int {
	n := 0
	for _, k := range karten {
		if k == wert {
			n++
		}
	}
	return n
}

func caboSatz(g *CaboGame, sitz int, ziele []int) bool {
	p := &g.Spieler[sitz]
	if len(ziele) < 2 || len(ziele) > len(p.Karten) {
		return false
	}
	gesehen := map[int]bool{}
	for _, pos := range ziele {
		if pos < 0 || pos >= len(p.Karten) || gesehen[pos] {
			return false
		}
		gesehen[pos] = true
	}
	wert := p.Karten[ziele[0]]
	for _, pos := range ziele {
		if p.Karten[pos] != wert {
			return false
		}
	}
	sort.Ints(ziele)
	// Satz abwerfen, neue Karte an die erste frei gewordene Stelle.
	for _, pos := range ziele {
		g.Ablage = append(g.Ablage, p.Karten[pos])
	}
	neu := make([]int, 0, len(p.Karten)-len(ziele)+1)
	neuBekannt := map[string]int{}
	zielIt := 0
	eingesetzt := false
	for pos, karte := range p.Karten {
		if zielIt < len(ziele) && ziele[zielIt] == pos {
			if !eingesetzt {
				neu = append(neu, *g.Gezogen)
				neuBekannt[caboSchluessel(sitz, len(neu)-1)] = *g.Gezogen
				eingesetzt = true
			}
			zielIt++
			continue
		}
		neu = append(neu, karte)
		if wert, ok := p.Bekannt[caboSchluessel(sitz, pos)]; ok {
			neuBekannt[caboSchluessel(sitz, len(neu)-1)] = wert
		}
	}
	p.Karten = neu
	p.Bekannt = neuBekannt
	return true
}

func caboSchluessel(sitz, pos int) string {
	return strconv.Itoa(sitz) + ":" + strconv.Itoa(pos)
}

// ---------------------------------------------------------------- Zug

func caboNachziehen(g *CaboGame) {
	if len(g.Stapel) == 0 {
		// Ablage bis auf die oberste Karte neu mischen – die bleibt liegen.
		oben := g.Ablage[len(g.Ablage)-1]
		rest := append([]int{}, g.Ablage[:len(g.Ablage)-1]...)
		rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
		g.Ablage = []int{oben}
		g.Stapel = rest
	}
}

// caboZugWeiter beendet den Zug. Liegt noch eine gezogene Karte da, wandert
// sie offen auf den Ablagestapel.
func caboZugWeiter(g *CaboGame) {
	if g.Gezogen != nil {
		g.Ablage = append(g.Ablage, *g.Gezogen)
	}
	g.Gezogen = nil
	g.Quelle = ""
	if g.Status != "active" {
		return
	}
	if g.CaboVon >= 0 {
		g.RestZuege--
		if g.RestZuege <= 0 {
			caboWertung(g)
			return
		}
	}
	g.Zug = (g.Zug + 1) % len(g.Spieler)
	if g.Spieler[g.Zug].ID == "" {
		g.Zug = 0
	}
}

func caboWertung(g *CaboGame) {
	summen := make([]int, len(g.Spieler))
	for i := range g.Spieler {
		summe := 0
		for _, karte := range g.Spieler[i].Karten {
			summe += karte
		}
		summen[i] = summe
	}

	kamikaze := -1
	for i := range g.Spieler {
		if caboHatWert(g.Spieler[i].Karten, 12) >= 2 && caboHatWert(g.Spieler[i].Karten, 13) >= 2 {
			kamikaze = i
		}
	}

	rundenPunkte := make([]int, len(g.Spieler))
	sieger := -1

	if kamikaze >= 0 {
		for i := range g.Spieler {
			if i == kamikaze {
				rundenPunkte[i] = 0
			} else {
				rundenPunkte[i] = 50
			}
		}
		sieger = kamikaze
		g.Note = "Kamikaze von " + g.Spieler[kamikaze].Name + "! Beide 12 und beide 13 – 0 Punkte, alle anderen 50."
	} else {
		minimum := summen[0]
		for _, summe := range summen {
			if summe < minimum {
				minimum = summe
			}
		}
		gewinner := []int{}
		for i, summe := range summen {
			if summe == minimum {
				gewinner = append(gewinner, i)
			}
		}
		// Cabo-Ansager gewinnt einen Gleichstand für sich allein.
		if g.CaboVon >= 0 && summen[g.CaboVon] == minimum {
			gewinner = []int{g.CaboVon}
		}
		for i := range g.Spieler {
			istGewinner := false
			for _, k := range gewinner {
				if k == i {
					istGewinner = true
				}
			}
			if !istGewinner {
				rundenPunkte[i] = summen[i]
			}
		}
		if g.CaboVon >= 0 && summen[g.CaboVon] != minimum {
			rundenPunkte[g.CaboVon] += 5
		}
		sieger = gewinner[0]
		namen := []string{}
		for _, i := range gewinner {
			namen = append(namen, g.Spieler[i].Name)
		}
		g.Note = "Wertung – am wenigsten Punkte: " + strings.Join(namen, ", ") + "."
		if g.CaboVon >= 0 && summen[g.CaboVon] != minimum {
			g.Note += " " + g.Spieler[g.CaboVon].Name + " hat Cabo gerufen und bekommt 5 Punkte extra."
		}
	}

	for i := range g.Spieler {
		g.Spieler[i].Runden = rundenPunkte[i]
		g.Spieler[i].Punkte += rundenPunkte[i]
		if g.Spieler[i].Punkte == 100 {
			g.Spieler[i].Punkte = 50
		}
	}

	g.RundenPunkte = rundenPunkte
	g.RundenSieger = sieger
	g.Status = "wertung"

	// Neuer Startspieler: Rundensieger, bei Gleichstand der mit den
	// wenigsten Gesamtpunkten.
	g.Start = sieger
	best := -1
	for i := range g.Spieler {
		if g.Spieler[i].Runden != 0 {
			continue
		}
		if best < 0 || g.Spieler[i].Punkte < g.Spieler[best].Punkte {
			best = i
		}
	}
	if best >= 0 {
		g.Start = best
	}

	// Spielende bei 101 oder mehr Punkten.
	ende := false
	for i := range g.Spieler {
		if g.Spieler[i].Punkte >= caboZiel {
			ende = true
		}
	}
	if ende {
		best := g.Spieler[0].Punkte
		gewinner := 0
		for i := range g.Spieler {
			if g.Spieler[i].Punkte < best {
				best = g.Spieler[i].Punkte
				gewinner = i
			}
		}
		g.Status = "ende"
		g.Sieger = g.Spieler[gewinner].Name
		g.Note += " Spiel beendet – " + g.Sieger + " gewinnt mit " + strconv.Itoa(best) + " Punkten."
	}
}

// ---------------------------------------------------------------- Sicht

func caboAnsicht(g *CaboGame, id string) *CaboAnsicht {
	if g == nil {
		return nil
	}
	a := &CaboAnsicht{
		ID:        g.ID,
		HostID:    g.HostID,
		Status:    g.Status,
		Runde:     g.Runde,
		Zug:       g.Zug,
		Start:     g.Start,
		Spieler:   make([]CaboSpielerSicht, 0, len(g.Spieler)),
		Stapel:    len(g.Stapel),
		Quelle:    g.Quelle,
		CaboVon:   g.CaboVon,
		RestZuege: g.RestZuege,
		Note:      g.Note,
		Sieger:    g.Sieger,
	}
	if len(g.Ablage) > 0 {
		oben := g.Ablage[len(g.Ablage)-1]
		a.AblageOben = &oben
	}
	a.AblageAnzahl = len(g.Ablage)
	if g.Gezogen != nil {
		wert := *g.Gezogen
		a.Gezogen = &wert
	}
	if g.RundenPunkte != nil {
		a.RundenPunkte = append([]int{}, g.RundenPunkte...)
	}
	offen := g.Status == "wertung" || g.Status == "ende"

	for i, p := range g.Spieler {
		sicht := CaboSpielerSicht{
			ID:       p.ID,
			Name:     p.Name,
			Anzahl:   len(p.Karten),
			Punkte:   p.Punkte,
			Runden:   p.Runden,
			Ready:    p.Ready,
			Geschaut: p.Geschaut,
			Bekannt:  map[string]int{},
		}
		if p.ID == id {
			for k, v := range p.Bekannt {
				sicht.Bekannt[k] = v
			}
		}
		if offen {
			sicht.Werte = append([]int{}, p.Karten...)
		}
		_ = i
		a.Spieler = append(a.Spieler, sicht)
	}
	return a
}

// ---------------------------------------------------------------- Endpunkt

func caboHandler(w http.ResponseWriter, r *http.Request) {
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
		game := caboAnsicht(room.Cabo, id)
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
		Pos    int    `json:"pos"`
		Pos2   int    `json:"pos2"`
		Ziel   int    `json:"ziel"`
		Ziele  []int  `json:"ziele"`
		Quelle string `json:"quelle"`
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
		if room.Cabo != nil && room.Cabo.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Cabo-Runde.", http.StatusConflict)
			return
		}
		room.Cabo = &CaboGame{
			ID:      randomCode(),
			HostID:  input.ID,
			Status:  "invited",
			Spieler: []CaboSpieler{caboNeuerSpieler(input.ID, player.Name)},
			Note:    "Tisch eröffnet – 2 bis 4 Spieler, vier Karten vor jedem.",
		}
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Cabo-Runde – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Cabo
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Cabo-Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		drin := false
		for _, p := range game.Spieler {
			if p.ID == input.ID {
				drin = true
				break
			}
		}
		if !drin {
			if len(game.Spieler) >= caboMaxSpieler {
				mu.Unlock()
				http.Error(w, "Dieser Tisch ist mit 4 Spielern voll.", http.StatusConflict)
				return
			}
			game.Spieler = append(game.Spieler, caboNeuerSpieler(input.ID, player.Name))
			game.Note = player.Name + " setzt sich dazu."
			room.addMessage(Message{
				Name:     player.Name,
				Text:     "spielt Cabo mit.",
				Time:     zeit,
				SenderID: input.ID,
			})
		}

	case "ready":
		game := room.Cabo
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := caboFinde(game, input.ID)
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
		game := room.Cabo
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Runde auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber kann geben.", http.StatusForbidden)
			return
		}
		if len(game.Spieler) < 2 {
			mu.Unlock()
			http.Error(w, "Cabo braucht mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, p := range game.Spieler {
			if !p.Ready {
				mu.Unlock()
				http.Error(w, "Erst müssen alle BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.Status = "anschauen"
		game.Start = 0
		game.Runde = 0
		caboNeueRunde(game)
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "gibt die Karten – jeder schaut sich zwei an.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "anschauen":
		game := room.Cabo
		if game == nil || game.Status != "anschauen" {
			mu.Unlock()
			http.Error(w, "Gerade wird nicht angeschaut.", http.StatusConflict)
			return
		}
		sitz := caboFinde(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Runde.", http.StatusForbidden)
			return
		}
		p := &game.Spieler[sitz]
		if p.Geschaut >= caboStartBlick {
			mu.Unlock()
			http.Error(w, "Du hast schon zwei Karten gesehen.", http.StatusConflict)
			return
		}
		if input.Pos < 0 || input.Pos >= len(p.Karten) {
			mu.Unlock()
			http.Error(w, "Diese Karte gibt es nicht.", http.StatusBadRequest)
			return
		}
		scl := caboSchluessel(sitz, input.Pos)
		if _, schon := p.Bekannt[scl]; schon {
			mu.Unlock()
			http.Error(w, "Diese Karte kennst du schon – nimm eine andere.", http.StatusConflict)
			return
		}
		p.Bekannt[scl] = p.Karten[input.Pos]
		p.Geschaut++
		game.Note = player.Name + " schaut sich Karten an."
		alleFertig := true
		for _, sp := range game.Spieler {
			if sp.Geschaut < caboStartBlick {
				alleFertig = false
			}
		}
		if alleFertig {
			game.Status = "active"
			game.Zug = game.Start
			game.Note = "Los geht's – " + game.Spieler[game.Zug].Name + " beginnt."
		}

	case "nehmen", "ziehen":
		game := room.Cabo
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Gerade läuft kein Zug.", http.StatusConflict)
			return
		}
		if caboFinde(game, input.ID) != game.Zug {
			mu.Unlock()
			http.Error(w, "Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		if game.Gezogen != nil {
			mu.Unlock()
			http.Error(w, "Du hast schon eine Karte gezogen.", http.StatusConflict)
			return
		}
		if action == "nehmen" {
			if len(game.Ablage) == 0 {
				mu.Unlock()
				http.Error(w, "Der Ablagestapel ist leer.", http.StatusConflict)
				return
			}
			wert := game.Ablage[len(game.Ablage)-1]
			game.Ablage = game.Ablage[:len(game.Ablage)-1]
			game.Gezogen = &wert
			game.Quelle = "ablage"
			game.Note = player.Name + " nimmt die " + strconv.Itoa(wert) + " vom Ablagestapel und tauscht sie ein."
		} else {
			caboNachziehen(game)
			wert := game.Stapel[len(game.Stapel)-1]
			game.Stapel = game.Stapel[:len(game.Stapel)-1]
			game.Gezogen = &wert
			game.Quelle = "stapel"
			game.Note = player.Name + " zieht eine Karte vom Stapel."
		}

	case "legen":
		game := room.Cabo
		if game == nil || game.Gezogen == nil {
			mu.Unlock()
			http.Error(w, "Du hast keine Karte in der Hand.", http.StatusConflict)
			return
		}
		sitz := caboFinde(game, input.ID)
		if sitz != game.Zug {
			mu.Unlock()
			http.Error(w, "Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		p := &game.Spieler[sitz]
		if input.Pos < 0 || input.Pos >= len(p.Karten) {
			mu.Unlock()
			http.Error(w, "Diese Karte gibt es nicht.", http.StatusBadRequest)
			return
		}
		alt := p.Karten[input.Pos]
		p.Karten[input.Pos] = *game.Gezogen
		p.Bekannt[caboSchluessel(sitz, input.Pos)] = *game.Gezogen
		game.Ablage = append(game.Ablage, alt)
		game.Note = player.Name + " legt " + strconv.Itoa(*game.Gezogen) + " ab und wirft " + strconv.Itoa(alt) + " weg."
		game.Gezogen = nil
		caboZugWeiter(game)

	case "abwerfen":
		game := room.Cabo
		if game == nil || game.Gezogen == nil {
			mu.Unlock()
			http.Error(w, "Du hast keine Karte in der Hand.", http.StatusConflict)
			return
		}
		if caboFinde(game, input.ID) != game.Zug {
			mu.Unlock()
			http.Error(w, "Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		game.Note = player.Name + " wirft " + strconv.Itoa(*game.Gezogen) + " ab."
		caboZugWeiter(game)

	case "peek", "spy", "swap":
		game := room.Cabo
		if game == nil || game.Gezogen == nil {
			mu.Unlock()
			http.Error(w, "Dafür musst du zuerst eine Karte ziehen.", http.StatusConflict)
			return
		}
		if game.Quelle != "stapel" {
			mu.Unlock()
			http.Error(w, "Vom Ablagestapel genommene Karten haben keine Aktion.", http.StatusConflict)
			return
		}
		sitz := caboFinde(game, input.ID)
		if sitz != game.Zug {
			mu.Unlock()
			http.Error(w, "Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		wert := *game.Gezogen
		moeglich := false
		switch action {
		case "peek":
			moeglich = wert == 7 || wert == 8
		case "spy":
			moeglich = wert == 9 || wert == 10
		case "swap":
			moeglich = wert == 11 || wert == 12
		}
		if !moeglich {
			mu.Unlock()
			http.Error(w, "Diese Karte hat keine solche Aktion.", http.StatusConflict)
			return
		}
		p := &game.Spieler[sitz]

		switch action {
		case "peek":
			if input.Pos < 0 || input.Pos >= len(p.Karten) {
				mu.Unlock()
				http.Error(w, "Diese Karte gibt es nicht.", http.StatusBadRequest)
				return
			}
			p.Bekannt[caboSchluessel(sitz, input.Pos)] = p.Karten[input.Pos]
			game.Note = player.Name + " schaut sich eine eigene Karte an (" + strconv.Itoa(wert) + " abgeworfen)."

		case "spy":
			if input.Ziel < 0 || input.Ziel >= len(game.Spieler) || input.Ziel == sitz {
				mu.Unlock()
				http.Error(w, "Wähle einen Mitspieler.", http.StatusBadRequest)
				return
			}
			ziel := input.Ziel
			z := &game.Spieler[ziel]
			if input.Pos < 0 || input.Pos >= len(z.Karten) {
				mu.Unlock()
				http.Error(w, "Diese Karte gibt es nicht.", http.StatusBadRequest)
				return
			}
			p.Bekannt[caboSchluessel(ziel, input.Pos)] = z.Karten[input.Pos]
			game.Note = player.Name + " späht bei " + z.Name + "."

		case "swap":
			if input.Ziel < 0 || input.Ziel >= len(game.Spieler) || input.Ziel == sitz {
				mu.Unlock()
				http.Error(w, "Wähle einen Mitspieler.", http.StatusBadRequest)
				return
			}
			z := &game.Spieler[input.Ziel]
			if input.Pos < 0 || input.Pos >= len(p.Karten) || input.Pos2 < 0 || input.Pos2 >= len(z.Karten) {
				mu.Unlock()
				http.Error(w, "Diese Karte gibt es nicht.", http.StatusBadRequest)
				return
			}
			p.Karten[input.Pos], z.Karten[input.Pos2] = z.Karten[input.Pos2], p.Karten[input.Pos]
			delete(p.Bekannt, caboSchluessel(sitz, input.Pos))
			delete(p.Bekannt, caboSchluessel(input.Ziel, input.Pos2))
			delete(z.Bekannt, caboSchluessel(input.Ziel, input.Pos2))
			delete(z.Bekannt, caboSchluessel(sitz, input.Pos))
			game.Note = player.Name + " tauscht eine Karte mit " + z.Name + "."
		}
		caboZugWeiter(game)

	case "satz":
		game := room.Cabo
		if game == nil || game.Gezogen == nil {
			mu.Unlock()
			http.Error(w, "Dafür musst du zuerst eine Karte gezogen haben.", http.StatusConflict)
			return
		}
		sitz := caboFinde(game, input.ID)
		if sitz != game.Zug {
			mu.Unlock()
			http.Error(w, "Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		if !caboSatz(game, sitz, input.Ziele) {
			game.Note = player.Name + " behauptet einen Satz – falsch, der Zug ist vorbei."
			caboZugWeiter(game)
			break
		}
		game.Note = player.Name + " wirft einen Satz ab und legt " + strconv.Itoa(*game.Gezogen) + " hinein."
		game.Gezogen = nil
		caboZugWeiter(game)

	case "cabo":
		game := room.Cabo
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Gerade läuft kein Zug.", http.StatusConflict)
			return
		}
		sitz := caboFinde(game, input.ID)
		if sitz != game.Zug {
			mu.Unlock()
			http.Error(w, "Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		if game.CaboVon >= 0 {
			mu.Unlock()
			http.Error(w, "Cabo wurde in dieser Runde schon gerufen.", http.StatusConflict)
			return
		}
		if game.Gezogen != nil {
			mu.Unlock()
			http.Error(w, "Cabo ruft man, bevor man eine Karte zieht.", http.StatusConflict)
			return
		}
		game.CaboVon = sitz
		// Der Rufende zählt mit: sein eigener Zug wird gleich beendet, danach
		// kommt jeder andere noch genau einmal dran.
		game.RestZuege = len(game.Spieler)
		game.Note = player.Name + " ruft CABO! Alle anderen sind noch einmal dran."
		caboZugWeiter(game)

	case "naechste":
		game := room.Cabo
		if game == nil || game.Status != "wertung" {
			mu.Unlock()
			http.Error(w, "Die Runde ist noch nicht gewertet.", http.StatusConflict)
			return
		}
		if caboFinde(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Runde.", http.StatusForbidden)
			return
		}
		caboNeueRunde(game)

	case "reset":
		game := room.Cabo
		if game == nil || game.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Eine neue Partie geht erst nach dem Ende.", http.StatusConflict)
			return
		}
		if caboFinde(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Runde.", http.StatusForbidden)
			return
		}
		frisch := &CaboGame{
			ID:      game.ID,
			HostID:  game.HostID,
			Status:  "invited",
			Spieler: []CaboSpieler{},
			Note:    "Neue Partie – erst BEREIT drücken.",
		}
		for _, p := range game.Spieler {
			frisch.Spieler = append(frisch.Spieler, caboNeuerSpieler(p.ID, p.Name))
		}
		room.Cabo = frisch

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Cabo-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := caboAnsicht(room.Cabo, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

func caboFinde(g *CaboGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}
