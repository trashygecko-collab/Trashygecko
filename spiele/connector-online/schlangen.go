package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Schlangen und Leitern: 100 Felder in zehn Reihen, abwechselnd von links nach
// rechts und zurück nummeriert. Gewürfelt wird mit einem Würfel; Leitern und
// Schlangen liegen pro Runde fest und werden auf dem Feld gezeichnet. Wer 100
// mit der passenden Zahl erreicht, gewinnt – zu weit gewürfelt heißt stehen
// bleiben.
const (
	schlangenMaxSpieler = 4
	schlangenFelder     = 100
	schlangenLeitern    = 6
	schlangenSchlangen  = 6
)

type SchlangenSpieler struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
	Farbe int    `json:"farbe"`
	Pos   int    `json:"pos"` // 0 = noch nicht auf dem Feld
	Zuege int    `json:"zuege"`
}

type SchlangenGame struct {
	ID        string             `json:"id"`
	HostID    string             `json:"hostId"`
	Status    string             `json:"status"` // invited, spiel, ende
	Runde     int                `json:"runde"`
	Zug       int                `json:"zug"`
	Spieler   []SchlangenSpieler `json:"spieler"`
	Wurf      int                `json:"wurf"`
	Leitern   [][2]int           `json:"leitern"`
	Schlangen [][2]int           `json:"schlangen"`
	Letzte    string             `json:"letzte,omitempty"`
	Verlauf   []string           `json:"verlauf"`
	Sieger    string             `json:"sieger,omitempty"`
	Note      string             `json:"note,omitempty"`
}

func schlangenNeuesSpiel(hostID, hostName, id string) *SchlangenGame {
	return &SchlangenGame{
		ID: id, HostID: hostID, Status: "invited",
		Leitern: [][2]int{}, Schlangen: [][2]int{}, Verlauf: []string{},
		Spieler: []SchlangenSpieler{{ID: hostID, Name: hostName}},
	}
}

func (g *SchlangenGame) sitz(id string) int {
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			return i
		}
	}
	return -1
}

// schlangenBrett legt Leitern und Schlangen für eine Runde fest. Start und Ziel
// sind eindeutig, damit keine Kette entsteht und keine Figur doppelt springt.
func schlangenBrett() ([][2]int, [][2]int) {
	belegt := map[int]bool{1: true, 100: true}
	leitern := [][2]int{}
	for versuch := 0; versuch < 400 && len(leitern) < schlangenLeitern; versuch++ {
		start := 2 + rand.Intn(88) // 2..89
		ende := start + 11 + rand.Intn(18)
		if ende > 99 || belegt[start] || belegt[ende] {
			continue
		}
		belegt[start], belegt[ende] = true, true
		leitern = append(leitern, [2]int{start, ende})
	}
	schlangen := [][2]int{}
	// Zwei Schlangen beginnen in den letzten zwei Reihen (81..97): auch dort ist
	// man noch nicht sicher vor dem Rutschen, wie auf einem echten Brett.
	for versuch := 0; versuch < 400 && len(schlangen) < 2; versuch++ {
		start := 81 + rand.Intn(17) // 81..97
		ende := start - (11 + rand.Intn(18))
		if ende < 2 || belegt[start] || belegt[ende] {
			continue
		}
		belegt[start], belegt[ende] = true, true
		schlangen = append(schlangen, [2]int{start, ende})
	}
	for versuch := 0; versuch < 400 && len(schlangen) < schlangenSchlangen; versuch++ {
		start := 12 + rand.Intn(86) // 12..97
		ende := start - (11 + rand.Intn(18))
		if ende < 2 || belegt[start] || belegt[ende] {
			continue
		}
		belegt[start], belegt[ende] = true, true
		schlangen = append(schlangen, [2]int{start, ende})
	}
	return leitern, schlangen
}

func (g *SchlangenGame) schlangenRunde() {
	g.Leitern, g.Schlangen = schlangenBrett()
	g.Wurf = 0
	g.Letzte = ""
	g.Verlauf = []string{}
	g.Sieger = ""
	g.Zug = 0
	for i := range g.Spieler {
		g.Spieler[i].Pos = 0
		g.Spieler[i].Zuege = 0
	}
	g.Status = "spiel"
	g.Note = "Wer Feld " + strconv.Itoa(schlangenFelder) + " genau erreicht, gewinnt. Leitern hoch, Schlangen hinunter."
}

func schlangenZiel(liste [][2]int, feld int) (int, bool) {
	for _, paar := range liste {
		if paar[0] == feld {
			return paar[1], true
		}
	}
	return 0, false
}

func (g *SchlangenGame) verlauf(text string) {
	g.Letzte = text
	g.Verlauf = append(g.Verlauf, text)
	if len(g.Verlauf) > 8 {
		g.Verlauf = g.Verlauf[len(g.Verlauf)-8:]
	}
}

func (g *SchlangenGame) naechster() {
	g.Zug = (g.Zug + 1) % len(g.Spieler)
}

// schlangenWuerfeln zieht eine Figur und löst Leitern und Schlangen aus.
func (g *SchlangenGame) schlangenWuerfeln(sitz int) string {
	return g.schlangenZug(sitz, 1+rand.Intn(6))
}

// schlangenZug rechnet einen Zug mit vorgegebener Augenzahl.
func (g *SchlangenGame) schlangenZug(sitz, wurf int) string {
	if g.Status != "spiel" {
		return "Gerade läuft keine Runde."
	}
	if sitz != g.Zug {
		return "Du bist nicht an der Reihe."
	}
	if wurf < 1 || wurf > 6 {
		return "Ungültige Augenzahl."
	}
	g.Wurf = wurf
	sp := &g.Spieler[sitz]
	sp.Zuege++
	ziel := sp.Pos + wurf
	if ziel > schlangenFelder {
		g.verlauf(sp.Name + " würfelt " + strconv.Itoa(wurf) + " – das ist zu weit, die Figur bleibt auf " + strconv.Itoa(sp.Pos) + ".")
		g.naechster()
		return ""
	}
	sp.Pos = ziel
	text := sp.Name + " würfelt " + strconv.Itoa(wurf) + " und zieht auf Feld " + strconv.Itoa(ziel) + "."
	if to, ok := schlangenZiel(g.Leitern, ziel); ok {
		sp.Pos = to
		text += " Die Leiter führt hinauf auf " + strconv.Itoa(to) + "."
	} else if to, ok := schlangenZiel(g.Schlangen, ziel); ok {
		sp.Pos = to
		text += " Die Schlange rutscht hinunter auf " + strconv.Itoa(to) + "."
	}
	g.verlauf(text)
	if sp.Pos >= schlangenFelder {
		g.Status = "ende"
		g.Sieger = sp.ID
		g.Note = sp.Name + " erreicht Feld " + strconv.Itoa(schlangenFelder) + " und gewinnt."
		return ""
	}
	g.naechster()
	return ""
}

func schlangenAnzeige(g *SchlangenGame) *SchlangenGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Spieler = append([]SchlangenSpieler{}, g.Spieler...)
	c.Leitern = append([][2]int{}, g.Leitern...)
	c.Schlangen = append([][2]int{}, g.Schlangen...)
	c.Verlauf = append([]string{}, g.Verlauf...)
	return &c
}

func schlangenHandler(w http.ResponseWriter, r *http.Request) {
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
		game := schlangenAnzeige(room.Schlangen)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID string `json:"id"`
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
	amTisch := func(status, fehlt, nichtDabei string) (*SchlangenGame, int, string, int) {
		g := room.Schlangen
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
		if room.Schlangen != nil {
			slot := room.Schlangen.Status
			if slot == "spiel" {
				slot = "active"
			}
			if !room.gameSlotFree(slot, schlangenIDs(room.Schlangen)...) {
				fail("In diesem Spielraum läuft bereits Schlangen und Leitern.", http.StatusConflict)
				return
			}
		}
		room.Schlangen = schlangenNeuesSpiel(input.ID, player.Name, randomCode())
		room.Schlangen.Note = player.Name + " eröffnet eine Runde Schlangen und Leitern."
	case "join":
		g := room.Schlangen
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Runde.", http.StatusConflict)
			return
		}
		if g.sitz(input.ID) >= 0 {
			fail("Du sitzt schon am Tisch.", http.StatusConflict)
			return
		}
		if len(g.Spieler) >= schlangenMaxSpieler {
			fail("Der Tisch ist voll (höchstens "+strconv.Itoa(schlangenMaxSpieler)+" Spieler).", http.StatusConflict)
			return
		}
		g.Spieler = append(g.Spieler, SchlangenSpieler{ID: input.ID, Name: player.Name, Farbe: len(g.Spieler)})
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
			fail("Schlangen und Leitern braucht mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, s := range g.Spieler {
			if !s.Ready {
				fail("Noch nicht alle sind bereit.", http.StatusConflict)
				return
			}
		}
		g.Runde = 1
		g.schlangenRunde()
	case "wuerfeln":
		g, sitz, problem, status := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, status)
			return
		}
		if text := g.schlangenWuerfeln(sitz); text != "" {
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
		g.schlangenRunde()
	case "reset":
		if room.Schlangen != nil && room.Schlangen.HostID != input.ID {
			fail("Nur der Gastgeber kann die Runde schließen.", http.StatusForbidden)
			return
		}
		room.Schlangen = nil
	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}
	room.LastActive = time.Now()
	game := schlangenAnzeige(room.Schlangen)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": game})
}

func schlangenIDs(g *SchlangenGame) []string {
	ids := make([]string, 0, len(g.Spieler))
	for _, s := range g.Spieler {
		ids = append(ids, s.ID)
	}
	return ids
}

func schlangenRemoveSeat(g *SchlangenGame, id, name string) (bool, bool, int) {
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
