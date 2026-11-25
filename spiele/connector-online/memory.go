package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Memory zu zweit: ein gemeinsames Brett, abwechselnd zwei Karten aufdecken.
// Wer ein Paar findet, behält es und ist noch einmal dran. Am Ende gewinnt,
// wer mehr Paare hat.
type MemoryGame struct {
	ID       string `json:"id"`
	RedID    string `json:"redId"`
	RedName  string `json:"redName"`
	BlueID   string `json:"blueId"`
	BlueName string `json:"blueName"`
	Paare    int    `json:"pairs"`

	Karten   []int  `json:"-"`        // Wert je Feld (geheim)
	Besitzer []int  `json:"besitzer"` // -1 frei, 0 rot, 1 blau
	Offen    []int  `json:"offen"`    // gerade aufgedeckte Felder
	Punkte   [2]int `json:"punkte"`
	Zuege    [2]int `json:"zuege"`
	Turn     string `json:"turn"`
	Status   string `json:"status"`
	Winner   string `json:"winner,omitempty"`
	Note     string `json:"note,omitempty"`
	Letzte   []int  `json:"letzte,omitempty"`
	Treffer  bool   `json:"treffer"`

	Bots     []string  `json:"bots,omitempty"`
	BotStufe string    `json:"botStufe,omitempty"`
	BotFrist time.Time `json:"-"`

	BotWeiss map[int]int `json:"-"` // Bot-Gedächtnis: Feld → Wert
	BotReihe []int       `json:"-"` // Reihenfolge fürs Vergessen
}

func memoryPaareNormal(n int) int {
	switch {
	case n < 8:
		return 6
	case n > 8:
		return 10
	}
	return 8
}

// memoryNewGame legt ein Brett an und mischt die Karten.
func memoryNewGame(hostID, hostName, guestID, guestName, id string, paare int) *MemoryGame {
	paare = memoryPaareNormal(paare)
	karten := make([]int, 0, paare*2)
	for i := 0; i < paare; i++ {
		karten = append(karten, i, i)
	}
	rand := botZufall()
	rand.Shuffle(len(karten), func(i, j int) { karten[i], karten[j] = karten[j], karten[i] })
	besitzer := make([]int, len(karten))
	for i := range besitzer {
		besitzer[i] = -1
	}
	return &MemoryGame{
		ID:       id,
		RedID:    hostID,
		RedName:  hostName,
		BlueID:   guestID,
		BlueName: guestName,
		Paare:    paare,
		Karten:   karten,
		Besitzer: besitzer,
		Offen:    []int{},
		Turn:     "red",
		Status:   "active",
		Note:     "Rot beginnt: zwei Karten aufdecken.",
	}
}

func cloneMemory(g *MemoryGame) *MemoryGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Karten = append([]int(nil), g.Karten...)
	c.Besitzer = append([]int(nil), g.Besitzer...)
	c.Offen = append([]int(nil), g.Offen...)
	c.Letzte = append([]int(nil), g.Letzte...)
	return &c
}

func memorySitz(g *MemoryGame, id string) int {
	switch id {
	case g.RedID:
		return 0
	case g.BlueID:
		return 1
	}
	return -1
}

// memorySicht baut den Spielstand für ein Gerät: nur aufgedeckte und
// gefundene Karten zeigen ihren Wert, der Rest bleibt verdeckt.
func memorySicht(g *MemoryGame, id string) map[string]any {
	if g == nil {
		return nil
	}
	sicht := make([]map[string]any, len(g.Karten))
	offen := map[int]bool{}
	for _, i := range g.Offen {
		offen[i] = true
	}
	for i, wert := range g.Karten {
		eintrag := map[string]any{"besitzer": g.Besitzer[i], "offen": offen[i]}
		if g.Besitzer[i] >= 0 || offen[i] {
			eintrag["wert"] = wert
		}
		sicht[i] = eintrag
	}
	return map[string]any{
		"id":       g.ID,
		"pairs":    g.Paare,
		"karten":   sicht,
		"offen":    g.Offen,
		"letzte":   g.Letzte,
		"treffer":  g.Treffer,
		"punkte":   g.Punkte,
		"zuege":    g.Zuege,
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
}

// memoryAktion deckt eine Karte auf oder startet neu – für Menschen und Bots
// dieselbe Prüfung.
func memoryAktion(room *Room, id, aktion string, feld int) (int, string) {
	g := room.Memory
	if g == nil {
		return http.StatusConflict, "Es läuft gerade keine Memory-Partie."
	}
	sitz := memorySitz(g, id)
	if sitz < 0 {
		return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
	}
	farbe := []string{"red", "blue"}[sitz]

	switch aktion {
	case "flip":
		if g.Status != "active" {
			return http.StatusConflict, "Die Partie ist beendet."
		}
		if g.Turn != farbe {
			return http.StatusConflict, "Der andere Spieler ist an der Reihe."
		}
		if feld < 0 || feld >= len(g.Karten) {
			return http.StatusBadRequest, "Dieses Feld gibt es nicht."
		}
		if g.Besitzer[feld] >= 0 {
			return http.StatusConflict, "Diese Karte ist schon gefunden."
		}
		for _, offen := range g.Offen {
			if offen == feld {
				return http.StatusConflict, "Diese Karte liegt schon offen."
			}
		}
		g.Letzte = nil
		g.Offen = append(g.Offen, feld)
		memoryBotMerken(g, feld)
		if len(g.Offen) < 2 {
			return 0, ""
		}

		a, b := g.Offen[0], g.Offen[1]
		g.Zuege[sitz]++
		g.Offen = []int{}
		g.Letzte = []int{a, b}
		if g.Karten[a] == g.Karten[b] {
			g.Besitzer[a], g.Besitzer[b] = sitz, sitz
			g.Punkte[sitz]++
			g.Treffer = true
			if memoryFertig(g) {
				g.Status = "finished"
				g.Winner = memorySieger(g)
				g.Note = "Alle Paare gefunden – " + memorySiegerName(g) + " gewinnt " +
					memoryStand(g) + "."
				return 0, ""
			}
			g.Note = memoryName(g, farbe) + " findet ein Paar (" + memoryStand(g) + ") und ist noch einmal dran."
			return 0, ""
		}
		g.Treffer = false
		g.Turn = memoryOther(farbe)
		g.Note = "Kein Paar – " + memoryName(g, g.Turn) + " ist dran (" + memoryStand(g) + ")."

	case "reset":
		if g.Status != "finished" || (id != g.RedID && id != g.BlueID) {
			return http.StatusConflict, "Eine neue Runde ist erst nach Spielende möglich."
		}
		fresh := memoryNewGame(g.RedID, g.RedName, g.BlueID, g.BlueName, randomCode(), g.Paare)
		fresh.Bots, fresh.BotStufe = g.Bots, g.BotStufe
		if len(fresh.Bots) > 0 {
			fresh.BotFrist = time.Now().Add(botWartezeit(fresh.BotStufe))
		}
		room.Memory = fresh

	default:
		return http.StatusBadRequest, "Unbekannte Memory-Aktion."
	}

	return 0, ""
}

func memoryOther(farbe string) string {
	if farbe == "red" {
		return "blue"
	}
	return "red"
}

func memoryName(g *MemoryGame, farbe string) string {
	if farbe == "red" {
		return g.RedName
	}
	return g.BlueName
}

func memorySieger(g *MemoryGame) string {
	switch {
	case g.Punkte[0] > g.Punkte[1]:
		return "red"
	case g.Punkte[1] > g.Punkte[0]:
		return "blue"
	}
	return "draw"
}

func memorySiegerName(g *MemoryGame) string {
	if g.Winner == "draw" {
		return "keiner"
	}
	return memoryName(g, g.Winner)
}

func memoryStand(g *MemoryGame) string {
	return goCount(g.Punkte[0]) + " : " + goCount(g.Punkte[1])
}

func memoryFertig(g *MemoryGame) bool {
	for _, b := range g.Besitzer {
		if b < 0 {
			return false
		}
	}
	return true
}

func memoryHandler(w http.ResponseWriter, r *http.Request) {
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
		memoryBotTreiben(room, time.Now())
		sicht := memorySicht(room.Memory, id)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": sicht})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID    string `json:"id"`
		ToID  string `json:"toId"`
		Feld  int    `json:"feld"`
		Stufe string `json:"stufe"`
		Paare int    `json:"pairs"`
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

	switch action {
	case "bot":
		room.botNamenNeu()
		if room.Memory != nil && !room.gameSlotFree(room.Memory.Status, room.Memory.RedID, room.Memory.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Memory-Partie.", http.StatusConflict)
			return
		}
		stufe := botStufe(input.Stufe)
		paare := input.Paare
		if paare == 0 {
			paare = 8
		}
		game := memoryNewGame(input.ID, player.Name, botID(1), room.botName(1), randomCode(), paare)
		game.Bots, game.BotStufe = []string{botID(1)}, stufe
		game.BotFrist = time.Now().Add(botWartezeit(stufe))
		game.Note = player.Name + " spielt gegen " + room.botName(1) +
			" (" + botStufeText(stufe) + ") – " + player.Name + " beginnt."
		room.Memory = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Memory gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     uhrzeit(time.Now()),
			SenderID: input.ID,
		})
		room.LastActive = time.Now()
		sicht := memorySicht(room.Memory, input.ID)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": sicht})
		return

	case "invite":
		target, ok := room.Players[input.ToID]
		if !ok || input.ToID == input.ID {
			mu.Unlock()
			http.Error(w, "Dieser Spieler ist nicht verfügbar.", http.StatusBadRequest)
			return
		}
		if room.Memory != nil && !room.gameSlotFree(room.Memory.Status, room.Memory.RedID, room.Memory.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Memory-Partie oder Einladung.", http.StatusConflict)
			return
		}
		paare := input.Paare
		if paare == 0 {
			paare = 8
		}
		room.Memory = memoryNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode(), paare)
		room.Memory.Status = "invited"
		room.Memory.Note = player.Name + " wartet auf " + target.Name + "."
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Memory ein.",
			Time:       uhrzeit(time.Now()),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "memory",
			GameID:     room.Memory.ID,
		})

	case "accept":
		game := room.Memory
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = "Rot beginnt: zwei Karten aufdecken."

	case "decline":
		game := room.Memory
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "flip", "reset":
		if status, meldung := memoryAktion(room, input.ID, action, input.Feld); meldung != "" {
			mu.Unlock()
			http.Error(w, meldung, status)
			return
		}

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Memory-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	sicht := memorySicht(room.Memory, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": sicht})
}
