package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

/* Mastermind zu zweit.

   Beide raten denselben geheimen Farbcode. Wer ihn zuerst knackt, gewinnt;
   reißt niemand die Höchstzahl an Versuchen, entscheidet die Zahl der
   schwarzen Treffer. Ratet einer, ist der andere dran. */

const mastermindPlatz = 4

type MastermindReihe struct {
	Farbe   string `json:"farbe"`
	Rate    []int  `json:"rate"`
	Schwarz int    `json:"schwarz"`
	Weiss   int    `json:"weiss"`
	Treffer bool   `json:"treffer"`
}

type MastermindGame struct {
	ID       string `json:"id"`
	RedID    string `json:"redId"`
	RedName  string `json:"redName"`
	BlueID   string `json:"blueId"`
	BlueName string `json:"blueName"`

	Farben int   `json:"farben"`
	Geheim []int `json:"-"`

	Reihen []MastermindReihe `json:"reihen"`
	Zuege  [2]int            `json:"zuege"`
	Max    int               `json:"max"`
	Turn   string            `json:"turn"`
	Status string            `json:"status"`
	Winner string            `json:"winner,omitempty"`
	Note   string            `json:"note,omitempty"`

	Bots     []string  `json:"bots,omitempty"`
	BotStufe string    `json:"botStufe,omitempty"`
	BotFrist time.Time `json:"-"`

	// BotZug ist ein fertig gerechneter Versuch. BotRechnet zeigt auf die
	// Partie, für die die Suche außerhalb der Sperre gerade läuft – ein Reset
	// ersetzt die Partie, damit gehört eine alte Rechnung nicht zur neuen.
	BotZug     []int           `json:"-"`
	BotRechnet *MastermindGame `json:"-"`
}

func mastermindNewGame(hostID, hostName, guestID, guestName, id string, farben int) *MastermindGame {
	if farben < 4 || farben > 7 {
		farben = 6
	}
	rnd := botZufall()
	geheim := make([]int, mastermindPlatz)
	for i := range geheim {
		geheim[i] = rnd.Intn(farben)
	}
	return &MastermindGame{
		ID:       id,
		RedID:    hostID,
		RedName:  hostName,
		BlueID:   guestID,
		BlueName: guestName,
		Farben:   farben,
		Geheim:   geheim,
		Reihen:   []MastermindReihe{},
		Max:      12,
		Turn:     "red",
		Status:   "active",
		Note:     "Rot fängt an: vier Farben raten.",
	}
}

func cloneMastermind(g *MastermindGame) *MastermindGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Geheim = append([]int(nil), g.Geheim...)
	c.Reihen = append([]MastermindReihe(nil), g.Reihen...)
	c.BotRechnet = nil // der Schnappschuss rechnet nicht selbst
	return &c
}

func mastermindSitz(g *MastermindGame, id string) int {
	switch id {
	case g.RedID:
		return 0
	case g.BlueID:
		return 1
	}
	return -1
}

// mastermindWerten zählt schwarze (richtige Stelle) und weiße (richtige Farbe,
// falsche Stelle) Treffer.
func mastermindWerten(rate, geheim []int) (int, int) {
	schwarz := 0
	restRate := []int{}
	restGeheim := []int{}
	for i := 0; i < mastermindPlatz && i < len(rate) && i < len(geheim); i++ {
		if rate[i] == geheim[i] {
			schwarz++
			continue
		}
		restRate = append(restRate, rate[i])
		restGeheim = append(restGeheim, geheim[i])
	}
	weiss := 0
	zaehler := map[int]int{}
	for _, g := range restGeheim {
		zaehler[g]++
	}
	for _, r := range restRate {
		if zaehler[r] > 0 {
			zaehler[r]--
			weiss++
		}
	}
	return schwarz, weiss
}

// mastermindSicht liefert den Spielstand so, dass der geheime Code nicht
// verraten wird – außer die Partie ist vorbei.
func mastermindSicht(g *MastermindGame, id string) map[string]any {
	if g == nil {
		return nil
	}
	sicht := map[string]any{
		"id":       g.ID,
		"farben":   g.Farben,
		"reihen":   g.Reihen,
		"zuege":    g.Zuege,
		"max":      g.Max,
		"turn":     g.Turn,
		"status":   g.Status,
		"winner":   g.Winner,
		"note":     g.Note,
		"redId":    g.RedID,
		"redName":  g.RedName,
		"blueId":   g.BlueID,
		"blueName": g.BlueName,
		"bots":     g.Bots,
		"botStufe": g.BotStufe,
		"ich":      id,
	}
	if g.Status == "finished" {
		sicht["geheim"] = g.Geheim
	}
	return sicht
}

// mastermindAktion prüft einen Rateversuch – für Menschen und Bots dieselbe
// Funktion.
func mastermindAktion(room *Room, id, aktion string, rate []int) (int, string) {
	g := room.Mastermind
	if g == nil {
		return http.StatusConflict, "Es läuft gerade keine Mastermind-Partie."
	}
	sitz := mastermindSitz(g, id)
	if sitz < 0 {
		return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
	}
	farbe := []string{"red", "blue"}[sitz]

	switch aktion {
	case "rate":
		if g.Status != "active" {
			return http.StatusConflict, "Die Partie ist beendet."
		}
		if g.Turn != farbe {
			return http.StatusConflict, "Der andere Spieler ist an der Reihe."
		}
		if len(rate) != mastermindPlatz {
			return http.StatusBadRequest, "Ein Versuch braucht vier Farben."
		}
		for _, f := range rate {
			if f < 0 || f >= g.Farben {
				return http.StatusBadRequest, "Diese Farbe gibt es nicht."
			}
		}
		schwarz, weiss := mastermindWerten(rate, g.Geheim)
		g.Zuege[sitz]++
		g.Reihen = append(g.Reihen, MastermindReihe{
			Farbe:   farbe,
			Rate:    append([]int(nil), rate...),
			Schwarz: schwarz,
			Weiss:   weiss,
			Treffer: schwarz == mastermindPlatz,
		})
		if schwarz == mastermindPlatz {
			g.Status = "finished"
			g.Winner = farbe
			g.Note = mastermindName(g, farbe) + " knackt den Code nach " +
				goCount(g.Zuege[sitz]) + " Versuchen."
			return 0, ""
		}
		if g.Zuege[0] >= g.Max && g.Zuege[1] >= g.Max {
			g.Status = "finished"
			g.Winner = mastermindBester(g)
			g.Note = "Versuche aufgebraucht – " + mastermindAusgang(g) + "."
			return 0, ""
		}
		g.Turn = mastermindOther(farbe)
		g.Note = goCount(g.Zuege[sitz]) + ". Versuch: " + goCount(schwarz) + " schwarz, " +
			goCount(weiss) + " weiß – " + mastermindName(g, g.Turn) + " ist dran."

	case "reset":
		if g.Status != "finished" {
			return http.StatusConflict, "Eine neue Runde ist erst nach Spielende möglich."
		}
		fresh := mastermindNewGame(g.RedID, g.RedName, g.BlueID, g.BlueName, randomCode(), g.Farben)
		fresh.Bots, fresh.BotStufe = g.Bots, g.BotStufe
		if len(fresh.Bots) > 0 {
			fresh.BotFrist = time.Now().Add(botWartezeit(fresh.BotStufe))
		}
		room.Mastermind = fresh

	default:
		return http.StatusBadRequest, "Unbekannte Mastermind-Aktion."
	}
	return 0, ""
}

// mastermindBester entscheidet bei aufgebrauchten Versuchen: bestes Einzel-
// ergebnis zählt, bei Gleichstand die kleinere Zahl Versuche.
func mastermindBester(g *MastermindGame) string {
	best := [2]int{-1, -1}
	anzahl := [2]int{}
	for _, r := range g.Reihen {
		sitz := 0
		if r.Farbe == "blue" {
			sitz = 1
		}
		if r.Schwarz > best[sitz] {
			best[sitz] = r.Schwarz
		}
		anzahl[sitz]++
	}
	if best[0] > best[1] {
		return "red"
	}
	if best[1] > best[0] {
		return "blue"
	}
	if anzahl[0] > 0 && anzahl[1] == 0 {
		return "red"
	}
	if anzahl[1] > 0 && anzahl[0] == 0 {
		return "blue"
	}
	return "draw"
}

func mastermindAusgang(g *MastermindGame) string {
	if g.Winner == "draw" {
		return "unentschieden"
	}
	return mastermindName(g, g.Winner) + " gewinnt"
}

func mastermindName(g *MastermindGame, farbe string) string {
	if farbe == "red" {
		return g.RedName
	}
	return g.BlueName
}

func mastermindOther(farbe string) string {
	if farbe == "red" {
		return "blue"
	}
	return "red"
}

func mastermindHandler(w http.ResponseWriter, r *http.Request) {
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
		mastermindBotTreiben(room, time.Now())
		sicht := mastermindSicht(room.Mastermind, id)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": sicht})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID     string `json:"id"`
		ToID   string `json:"toId"`
		Stufe  string `json:"stufe"`
		Farben int    `json:"farben"`
		Rate   []int  `json:"rate"`
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

	switch action {
	case "bot":
		room.botNamenNeu()
		if room.Mastermind != nil && !room.gameSlotFree(room.Mastermind.Status, room.Mastermind.RedID, room.Mastermind.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Mastermind-Partie.", http.StatusConflict)
			return
		}
		stufe := botStufe(input.Stufe)
		game := mastermindNewGame(input.ID, player.Name, botID(1), room.botName(1), randomCode(), input.Farben)
		game.Bots, game.BotStufe = []string{botID(1)}, stufe
		game.BotFrist = time.Now().Add(botWartezeit(stufe))
		game.Note = player.Name + " spielt gegen " + room.botName(1) +
			" (" + botStufeText(stufe) + ") – " + player.Name + " beginnt."
		room.Mastermind = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Mastermind gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "invite":
		target, ok := room.Players[input.ToID]
		if !ok || input.ToID == input.ID {
			mu.Unlock()
			http.Error(w, "Dieser Spieler ist nicht verfügbar.", http.StatusBadRequest)
			return
		}
		if room.Mastermind != nil && !room.gameSlotFree(room.Mastermind.Status, room.Mastermind.RedID, room.Mastermind.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Mastermind-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Mastermind = mastermindNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode(), input.Farben)
		room.Mastermind.Status = "invited"
		room.Mastermind.Note = player.Name + " wartet auf " + target.Name + "."
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Mastermind ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "mastermind",
			GameID:     room.Mastermind.ID,
		})

	case "accept":
		game := room.Mastermind
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = "Rot fängt an: vier Farben raten."

	case "decline":
		game := room.Mastermind
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "rate", "reset":
		if status, meldung := mastermindAktion(room, input.ID, action, input.Rate); meldung != "" {
			mu.Unlock()
			http.Error(w, meldung, status)
			return
		}

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Mastermind-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	sicht := mastermindSicht(room.Mastermind, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": sicht})
}
