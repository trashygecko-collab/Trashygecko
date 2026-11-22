package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Pig: Wer am Zug ist, würfelt beliebig oft. Jede Zahl von 2 bis 6 zählt zur
// Rundensumme, eine 1 löscht sie und beendet den Zug. Mit „Sichern" wandert
// die Rundensumme aufs Konto. Wer zuerst 100 Punkte auf dem Konto hat, gewinnt.
const (
	pigMaxSpieler = 4
	pigZiel       = 100
)

type PigSpieler struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Ready  bool   `json:"ready"`
	Farbe  int    `json:"farbe"`
	Punkte int    `json:"punkte"` // gesichert
	Runde  int    `json:"runde"`  // im laufenden Zug gesammelt
}

type PigGame struct {
	ID       string       `json:"id"`
	HostID   string       `json:"hostId"`
	Status   string       `json:"status"` // invited, spiel, ende
	Runde    int          `json:"runde"`
	Zug      int          `json:"zug"`
	Spieler  []PigSpieler `json:"spieler"`
	Wurf     int          `json:"wurf"`
	Letzte   string       `json:"letzte,omitempty"`
	Verlauf  []string     `json:"verlauf"`
	Sieger   string       `json:"sieger,omitempty"`
	Note     string       `json:"note,omitempty"`
	BotID    string       `json:"botId,omitempty"`
	BotStufe string       `json:"-"`
	BotFrist time.Time    `json:"-"`
}

func pigNeuesSpiel(hostID, hostName, id string) *PigGame {
	return &PigGame{
		ID: id, HostID: hostID, Status: "invited", Verlauf: []string{},
		Spieler: []PigSpieler{{ID: hostID, Name: hostName}},
	}
}

func (g *PigGame) sitz(id string) int {
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			return i
		}
	}
	return -1
}

func (g *PigGame) pigRunde() {
	g.Wurf = 0
	g.Letzte = ""
	g.Verlauf = []string{}
	g.Sieger = ""
	g.Zug = 0
	for i := range g.Spieler {
		g.Spieler[i].Punkte = 0
		g.Spieler[i].Runde = 0
	}
	g.Status = "spiel"
	g.Note = "Würfle so oft du willst – eine 1 nimmt alles aus diesem Zug. Bei 100 Punkten auf dem Konto gewinnst du."
}

func (g *PigGame) verlauf(text string) {
	g.Letzte = text
	g.Verlauf = append(g.Verlauf, text)
	if len(g.Verlauf) > 8 {
		g.Verlauf = g.Verlauf[len(g.Verlauf)-8:]
	}
}

func (g *PigGame) naechster() {
	g.Zug = (g.Zug + 1) % len(g.Spieler)
}

// pigWuerfeln zählt den Wurf zur Rundensumme oder löscht sie bei einer 1.
func (g *PigGame) pigWuerfeln(sitz int) string {
	return g.pigWurf(sitz, 1+rand.Intn(6))
}

// pigWurf rechnet einen Zug mit vorgegebener Augenzahl.
func (g *PigGame) pigWurf(sitz, wurf int) string {
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
	if wurf == 1 {
		verloren := sp.Runde
		sp.Runde = 0
		g.verlauf(sp.Name + " würfelt eine 1 – " + strconv.Itoa(verloren) + " Punkte aus diesem Zug sind weg.")
		g.naechster()
		return ""
	}
	sp.Runde += wurf
	if sp.Punkte+sp.Runde >= pigZiel {
		g.verlauf(sp.Name + " würfelt " + strconv.Itoa(wurf) + " – " + strconv.Itoa(sp.Runde) + " Punkte in diesem Zug. Mit Sichern ist das Spiel gewonnen.")
		return ""
	}
	g.verlauf(sp.Name + " würfelt " + strconv.Itoa(wurf) + " – " + strconv.Itoa(sp.Runde) + " Punkte in diesem Zug.")
	return ""
}

// pigSichern übernimmt die Rundensumme aufs Konto und beendet den Zug.
func (g *PigGame) pigSichern(sitz int) string {
	if g.Status != "spiel" {
		return "Gerade läuft keine Runde."
	}
	if sitz != g.Zug {
		return "Du bist nicht an der Reihe."
	}
	sp := &g.Spieler[sitz]
	if sp.Runde == 0 {
		return "In diesem Zug ist noch nichts zu sichern."
	}
	sp.Punkte += sp.Runde
	g.verlauf(sp.Name + " sichert " + strconv.Itoa(sp.Runde) + " Punkte und steht bei " + strconv.Itoa(sp.Punkte) + ".")
	sp.Runde = 0
	if sp.Punkte >= pigZiel {
		g.Status = "ende"
		g.Sieger = sp.ID
		g.Note = sp.Name + " hat " + strconv.Itoa(sp.Punkte) + " Punkte und gewinnt."
		return ""
	}
	g.naechster()
	return ""
}

func pigAnzeige(g *PigGame) *PigGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Spieler = append([]PigSpieler{}, g.Spieler...)
	c.Verlauf = append([]string{}, g.Verlauf...)
	return &c
}

func pigHandler(w http.ResponseWriter, r *http.Request) {
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
		pigBotTreiben(room.Pig, time.Now())
		game := pigAnzeige(room.Pig)
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
	fail := func(text string, status int) {
		mu.Unlock()
		http.Error(w, text, status)
	}
	amTisch := func(status, fehlt, nichtDabei string) (*PigGame, int, string, int) {
		g := room.Pig
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
	case "bot":
		room.botNamenNeu()
		stufe := botStufe(input.Stufe)
		if room.Pig != nil {
			slot := room.Pig.Status
			if slot == "spiel" {
				slot = "active"
			}
			if !room.gameSlotFree(slot, pigIDs(room.Pig)...) {
				fail("In diesem Spielraum läuft bereits eine Runde Pig.", http.StatusConflict)
				return
			}
		}
		g := pigNeuesSpiel(input.ID, player.Name, randomCode())
		g.Spieler = append(g.Spieler, PigSpieler{ID: botID(1), Name: room.botName(1), Farbe: 1, Ready: true})
		g.Spieler[0].Ready = true
		g.Runde = 1
		g.pigRunde()
		g.BotID, g.BotStufe = botID(1), stufe
		g.Note = player.Name + " beginnt gegen " + room.botName(1) + " (" + botStufeText(stufe) + ")."
		room.Pig = g

	case "create":
		if room.Pig != nil {
			slot := room.Pig.Status
			if slot == "spiel" {
				slot = "active"
			}
			if !room.gameSlotFree(slot, pigIDs(room.Pig)...) {
				fail("In diesem Spielraum läuft bereits eine Runde Pig.", http.StatusConflict)
				return
			}
		}
		room.Pig = pigNeuesSpiel(input.ID, player.Name, randomCode())
		room.Pig.Note = player.Name + " eröffnet eine Runde Pig."
	case "join":
		g := room.Pig
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Runde.", http.StatusConflict)
			return
		}
		if g.sitz(input.ID) >= 0 {
			fail("Du sitzt schon am Tisch.", http.StatusConflict)
			return
		}
		if len(g.Spieler) >= pigMaxSpieler {
			fail("Der Tisch ist voll (höchstens "+strconv.Itoa(pigMaxSpieler)+" Spieler).", http.StatusConflict)
			return
		}
		g.Spieler = append(g.Spieler, PigSpieler{ID: input.ID, Name: player.Name, Farbe: len(g.Spieler)})
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
			fail("Pig braucht mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, s := range g.Spieler {
			if !s.Ready {
				fail("Noch nicht alle sind bereit.", http.StatusConflict)
				return
			}
		}
		g.Runde = 1
		g.pigRunde()
	case "wuerfeln":
		g, sitz, problem, status := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, status)
			return
		}
		if text := g.pigWuerfeln(sitz); text != "" {
			fail(text, http.StatusConflict)
			return
		}
	case "sichern":
		g, sitz, problem, status := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, status)
			return
		}
		if text := g.pigSichern(sitz); text != "" {
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
		g.pigRunde()
	case "reset":
		if room.Pig != nil && room.Pig.HostID != input.ID {
			fail("Nur der Gastgeber kann die Runde schließen.", http.StatusForbidden)
			return
		}
		room.Pig = nil
	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}
	if room.Pig != nil && room.Pig.BotID != "" && input.ID != room.Pig.BotID {
		room.Pig.BotFrist = time.Time{}
	}
	room.LastActive = time.Now()
	game := pigAnzeige(room.Pig)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": game})
}

func pigIDs(g *PigGame) []string {
	ids := make([]string, 0, len(g.Spieler))
	for _, s := range g.Spieler {
		ids = append(ids, s.ID)
	}
	return ids
}

func pigRemoveSeat(g *PigGame, id, name string) (bool, bool, int) {
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
