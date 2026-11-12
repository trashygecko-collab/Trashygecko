package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	rechenTasks      = 8
	rechenMaxPlayers = 7
	rechenMinPlayers = 2
)

// RechenTask ist eine Aufgabe, wie sie auf dem Tisch liegt: nur die Zeilen,
// die alle sehen. Die Lösung bleibt auf dem Server.
type RechenTask struct {
	Lines  []string `json:"lines"`
	Answer int      `json:"-"`
}

type RechenPlayer struct {
	ID       string
	Name     string
	Ready    bool
	Index    int // so viele Aufgaben sind beantwortet
	Correct  int
	Done     bool
	Finished time.Time
}

type RechenGame struct {
	ID        string
	HostID    string
	Mode      string // arithmetic oder equations
	Level     int    // 1..3
	Players   []*RechenPlayer
	Tasks     []RechenTask
	Status    string // invited, active, finished
	Round     int
	StartedAt time.Time
	Winner    string
	Note      string
}

// --- Aufgaben erzeugen (gleiche Grenzen wie auf der Spielseite) -------------

func rechenRand(min, max int) int {
	if max < min {
		max = min
	}
	return rand.Intn(max-min+1) + min
}

func rechenArithmetic(level int) RechenTask {
	addMax := []int{10, 100, 1000}[level-1]
	mulMax := []int{5, 10, 20}[level-1]

	switch rechenRand(1, 4) {
	case 1:
		a := rechenRand(1, addMax-1)
		b := rechenRand(1, addMax-a)
		return RechenTask{Lines: []string{fmt.Sprintf("%d + %d", a, b)}, Answer: a + b}
	case 2:
		a := rechenRand(2, addMax)
		b := rechenRand(1, a-1)
		return RechenTask{Lines: []string{fmt.Sprintf("%d − %d", a, b)}, Answer: a - b}
	case 3:
		a := rechenRand(1, mulMax)
		b := rechenRand(1, mulMax)
		return RechenTask{Lines: []string{fmt.Sprintf("%d × %d", a, b)}, Answer: a * b}
	default:
		teiler := rechenRand(1, mulMax)
		quotient := rechenRand(1, mulMax)
		return RechenTask{Lines: []string{fmt.Sprintf("%d ÷ %d", teiler*quotient, teiler)}, Answer: quotient}
	}
}

func rechenEquation(level int) RechenTask {
	if level == 1 {
		switch rechenRand(1, 4) {
		case 1:
			x := rechenRand(1, 5)
			y := rechenRand(1, 5)
			z := rechenRand(0, x+y-1)
			return RechenTask{
				Lines:  []string{fmt.Sprintf("a = %d + %d − %d", x, y, z), "a = ?"},
				Answer: x + y - z,
			}
		case 2:
			x := rechenRand(1, 5)
			y := rechenRand(1, 10/x)
			return RechenTask{Lines: []string{fmt.Sprintf("a = %d × %d", x, y), "a = ?"}, Answer: x * y}
		case 3:
			teiler := rechenRand(1, 10)
			quotient := rechenRand(1, 10)
			return RechenTask{
				Lines:  []string{fmt.Sprintf("a = %d ÷ %d", teiler*quotient, teiler), "a = ?"},
				Answer: quotient,
			}
		default:
			x := rechenRand(1, 3)
			y := rechenRand(1, 3)
			z := rechenRand(1, 3)
			return RechenTask{
				Lines:  []string{fmt.Sprintf("a = %d + %d + %d", x, y, z), "a = ?"},
				Answer: x + y + z,
			}
		}
	}

	a := rechenRand(1, 10)
	faktor := rechenRand(2, 10)
	b := faktor * a

	if level == 2 {
		switch rechenRand(1, 4) {
		case 1:
			return RechenTask{Lines: []string{fmt.Sprintf("a = %d", a), fmt.Sprintf("b = %d × a", faktor), "b = ?"}, Answer: b}
		case 2:
			zusatz := rechenRand(1, 100-a)
			return RechenTask{Lines: []string{fmt.Sprintf("a = %d", a), fmt.Sprintf("b = a + %d", zusatz), "b = ?"}, Answer: a + zusatz}
		case 3:
			abzug := rechenRand(1, 20)
			start := rechenRand(abzug+1, 100)
			return RechenTask{Lines: []string{fmt.Sprintf("a = %d", start), fmt.Sprintf("b = a − %d", abzug), "b = ?"}, Answer: start - abzug}
		default:
			teiler := rechenRand(2, 10)
			quotient := rechenRand(1, 100/teiler)
			return RechenTask{Lines: []string{fmt.Sprintf("a = %d", teiler*quotient), fmt.Sprintf("b = a ÷ %d", teiler), "b = ?"}, Answer: quotient}
		}
	}

	switch rechenRand(1, 3) {
	case 1:
		mal := rechenRand(2, 10)
		return RechenTask{
			Lines: []string{
				fmt.Sprintf("a = %d", a),
				fmt.Sprintf("b = %d × a", faktor),
				fmt.Sprintf("c = b + %d × a", mal),
				"c = ?",
			},
			Answer: b + mal*a,
		}
	case 2:
		zusatz := rechenRand(5, 300)
		return RechenTask{
			Lines: []string{
				fmt.Sprintf("a = %d", a),
				fmt.Sprintf("b = %d × a", faktor),
				fmt.Sprintf("c = a + b + %d", zusatz),
				"c = ?",
			},
			Answer: a + b + zusatz,
		}
	default:
		zusatz := rechenRand(5, 200)
		return RechenTask{
			Lines: []string{
				fmt.Sprintf("a = %d", a),
				fmt.Sprintf("b = %d × a", faktor),
				fmt.Sprintf("c = b − a + %d", zusatz),
				"c = ?",
			},
			Answer: b - a + zusatz,
		}
	}
}

func rechenNewTasks(mode string, level int) []RechenTask {
	tasks := make([]RechenTask, 0, rechenTasks)
	for i := 0; i < rechenTasks; i++ {
		if mode == "equations" {
			tasks = append(tasks, rechenEquation(level))
		} else {
			tasks = append(tasks, rechenArithmetic(level))
		}
	}
	return tasks
}

func rechenNewGame(hostID, name string) *RechenGame {
	return &RechenGame{
		ID:      randomCode(),
		HostID:  hostID,
		Mode:    "arithmetic",
		Level:   1,
		Status:  "invited",
		Players: []*RechenPlayer{{ID: hostID, Name: name}},
	}
}

func rechenSeat(g *RechenGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// rechenZeit schreibt Millisekunden als "24,3 s" oder "1:05,2 min".
func rechenZeit(ms int64) string {
	if ms <= 0 {
		return "–"
	}
	sekunden := float64(ms) / 1000
	if sekunden < 60 {
		return strings.Replace(strconv.FormatFloat(sekunden, 'f', 1, 64), ".", ",", 1) + " s"
	}
	minuten := int(sekunden) / 60
	rest := strings.Replace(fmt.Sprintf("%04.1f", sekunden-float64(minuten*60)), ".", ",", 1)
	return strconv.Itoa(minuten) + ":" + rest + " min"
}

// rechenMs ist die gebrauchte Zeit eines Spielers. Wer nicht fertig geworden
// ist, zählt als langsam, damit die Rangfolge stimmt.
func (g *RechenGame) rechenMs(p *RechenPlayer) int64 {
	if p.Done && !p.Finished.IsZero() {
		return p.Finished.Sub(g.StartedAt).Milliseconds()
	}
	return int64(^uint64(0) >> 2) // sehr groß: hinten einsortieren
}

// ranking sortiert nach richtigen Antworten, bei Gleichstand nach Zeit.
func (g *RechenGame) ranking() []*RechenPlayer {
	list := append([]*RechenPlayer(nil), g.Players...)
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Correct != list[j].Correct {
			return list[i].Correct > list[j].Correct
		}
		return g.rechenMs(list[i]) < g.rechenMs(list[j])
	})
	return list
}

func (g *RechenGame) resultLines() []string {
	list := g.ranking()
	lines := make([]string, 0, len(list))
	for i, p := range list {
		zeit := ""
		if p.Done {
			zeit = rechenZeit(g.rechenMs(p))
		}
		lines = append(lines, fmt.Sprintf("%d. %s – %d von %d richtig%s",
			i+1, p.Name, p.Correct, len(g.Tasks), func() string {
				if zeit == "" {
					return ""
				}
				return " · " + zeit
			}()))
	}
	return lines
}

// --- Ansicht für einen Spieler ---------------------------------------------

type RechenSeatView struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Ready    bool   `json:"ready"`
	Done     bool   `json:"done"`
	Correct  int    `json:"correct"`
	Answered int    `json:"answered"`
	TimeMs   int64  `json:"timeMs"`
}

type RechenView struct {
	ID      string           `json:"id"`
	Status  string           `json:"status"`
	Mode    string           `json:"mode"`
	Level   int              `json:"level"`
	Count   int              `json:"count"`
	Seats   []RechenSeatView `json:"seats"`
	You     int              `json:"you"`
	Index   int              `json:"index"`
	Lines   []string         `json:"lines,omitempty"`
	Correct int              `json:"correct"`
	Done    bool             `json:"done"`
	IsHost  bool             `json:"isHost"`
	Note    string           `json:"note,omitempty"`
	Result  []string         `json:"result,omitempty"`
	Winner  string           `json:"winner,omitempty"`
	Elapsed int64            `json:"elapsedMs"`
	Round   int              `json:"round"`
}

func (g *RechenGame) viewFor(id string) RechenView {
	// In der Lobby stehen die Aufgaben noch nicht fest – dort trotzdem die
	// Zahl nennen, die auf alle wartet.
	anzahl := len(g.Tasks)
	if anzahl == 0 {
		anzahl = rechenTasks
	}
	view := RechenView{
		ID:      g.ID,
		Status:  g.Status,
		Mode:    g.Mode,
		Level:   g.Level,
		Count:   anzahl,
		You:     rechenSeat(g, id),
		IsHost:  g.HostID == id,
		Note:    g.Note,
		Winner:  g.Winner,
		Seats:   make([]RechenSeatView, 0, len(g.Players)),
		Elapsed: 0,
		Round:   g.Round,
	}

	for _, p := range g.Players {
		view.Seats = append(view.Seats, RechenSeatView{
			ID:       p.ID,
			Name:     p.Name,
			Ready:    p.Ready,
			Done:     p.Done,
			Correct:  p.Correct,
			Answered: p.Index,
			TimeMs:   rechenMsOrZero(g, p),
		})
	}

	if view.You >= 0 {
		me := g.Players[view.You]
		view.Index = me.Index
		view.Correct = me.Correct
		view.Done = me.Done
		if view.Status == "active" && !me.Done && me.Index < len(g.Tasks) {
			view.Lines = append([]string(nil), g.Tasks[me.Index].Lines...)
		}
	}

	if view.Status == "active" && !g.StartedAt.IsZero() {
		view.Elapsed = time.Since(g.StartedAt).Milliseconds()
	}
	if view.Status == "finished" {
		view.Result = g.resultLines()
	}
	return view
}

func rechenMsOrZero(g *RechenGame, p *RechenPlayer) int64 {
	if p.Done && !p.Finished.IsZero() {
		return p.Finished.Sub(g.StartedAt).Milliseconds()
	}
	return 0
}

// --- Spielraum-Aktionen -----------------------------------------------------

func einsxeinsHandler(w http.ResponseWriter, r *http.Request) {
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
		var view *RechenView
		if room.Rechen != nil {
			v := room.Rechen.viewFor(id)
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
		Mode  string `json:"mode"`
		Level int    `json:"level"`
		Index int    `json:"index"`
		Value *int   `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.ID = strings.TrimSpace(input.ID)
	input.Mode = strings.ToLower(strings.TrimSpace(input.Mode))
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
	antwort := func(extra map[string]any) {
		if room.Rechen != nil {
			v := room.Rechen.viewFor(input.ID)
			if extra == nil {
				extra = map[string]any{}
			}
			extra["game"] = v
			mu.Unlock()
			jsonResponse(w, extra)
			return
		}
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": nil})
	}

	game := room.Rechen

	switch action {
	case "create":
		if game != nil && game.Status != "finished" {
			fail("In diesem Spielraum läuft schon ein 1×1-Rennen.", http.StatusConflict)
			return
		}
		g := rechenNewGame(input.ID, player.Name)
		if input.Mode == "equations" || input.Mode == "arithmetic" {
			g.Mode = input.Mode
		}
		if input.Level >= 1 && input.Level <= 3 {
			g.Level = input.Level
		}
		g.Note = player.Name + " eröffnet ein 1×1-Rennen – bis zu 7 Spieler, mitspielen über die Spielkarte."
		room.Rechen = g
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet ein 1×1-Rennen – bis zu 7 Spieler, mitspielen über die Spielkarte.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "join":
		if game == nil || game.Status != "invited" {
			fail("Es läuft gerade keine offene 1×1-Runde.", http.StatusConflict)
			return
		}
		if rechenSeat(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= rechenMaxPlayers {
			fail("Diese Runde ist mit 7 Spielern voll.", http.StatusConflict)
			return
		}
		game.Players = append(game.Players, &RechenPlayer{ID: input.ID, Name: player.Name})
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zum 1×1-Rennen.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "ready":
		if game == nil || game.Status != "invited" {
			fail("Bereit drücken geht nur, solange das Rennen noch nicht läuft.", http.StatusConflict)
			return
		}
		seat := rechenSeat(game, input.ID)
		if seat < 0 {
			fail("Du bist bei diesem Rennen nicht dabei.", http.StatusForbidden)
			return
		}
		game.Players[seat].Ready = !game.Players[seat].Ready
		if game.Players[seat].Ready {
			game.Note = player.Name + " ist bereit."
		} else {
			game.Note = player.Name + " wartet noch."
		}

	case "settings":
		if game == nil || game.Status != "invited" {
			fail("Die Einstellungen lassen sich nur vor dem Start ändern.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			fail("Nur der Gastgeber stellt die Aufgaben ein.", http.StatusForbidden)
			return
		}
		if input.Mode == "equations" || input.Mode == "arithmetic" {
			game.Mode = input.Mode
		}
		if input.Level >= 1 && input.Level <= 3 {
			game.Level = input.Level
		}

	case "start":
		if game == nil || game.Status != "invited" {
			fail("Es wartet keine Runde auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			fail("Nur der Gastgeber kann das Rennen starten.", http.StatusForbidden)
			return
		}
		if len(game.Players) < rechenMinPlayers {
			fail("Für ein Rennen braucht es mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, p := range game.Players {
			if !p.Ready {
				fail("Erst müssen alle Spieler BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.Round++
		game.Tasks = rechenNewTasks(game.Mode, game.Level)
		game.StartedAt = time.Now()
		game.Status = "active"
		game.Winner = ""
		for _, p := range game.Players {
			p.Index, p.Correct, p.Done, p.Finished = 0, 0, false, time.Time{}
		}
		game.Note = "Los! Alle haben dieselben acht Aufgaben."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "startet das 1×1-Rennen.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "answer":
		if game == nil || game.Status != "active" {
			fail("Es läuft gerade kein Rennen.", http.StatusConflict)
			return
		}
		seat := rechenSeat(game, input.ID)
		if seat < 0 {
			fail("Du bist bei diesem Rennen nicht dabei.", http.StatusForbidden)
			return
		}
		me := game.Players[seat]
		if me.Done {
			fail("Du bist schon durch mit allen Aufgaben.", http.StatusConflict)
			return
		}
		if input.Value == nil {
			fail("Es fehlt eine Antwort.", http.StatusBadRequest)
			return
		}
		if input.Index != me.Index {
			fail("Diese Aufgabe ist schon beantwortet.", http.StatusConflict)
			return
		}
		task := game.Tasks[me.Index]
		richtig := *input.Value == task.Answer
		if richtig {
			me.Correct++
		}
		me.Index++
		if me.Index >= len(game.Tasks) {
			me.Done = true
			me.Finished = time.Now()
		}

		alleFertig := true
		for _, p := range game.Players {
			if !p.Done {
				alleFertig = false
				break
			}
		}
		if alleFertig {
			game.Status = "finished"
			best := game.ranking()[0]
			game.Winner = best.Name
			game.Note = best.Name + " gewinnt."
			room.addMessage(Message{
				Text:   "1×1-Rennen beendet – " + best.Name + " gewinnt.",
				System: true,
			})
		}

		mu.Unlock()
		jsonResponse(w, map[string]any{
			"game":    game.viewFor(input.ID),
			"correct": richtig,
			"answer":  task.Answer,
		})
		return

	case "reset":
		if game == nil {
			fail("Es läuft gerade keine Runde.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID && rechenSeat(game, input.ID) < 0 {
			fail("Nur Mitspieler können eine neue Runde eröffnen.", http.StatusForbidden)
			return
		}
		if game.Status == "active" {
			fail("Das Rennen läuft noch.", http.StatusConflict)
			return
		}
		game.Status = "invited"
		game.StartedAt = time.Time{}
		game.Winner = ""
		for _, p := range game.Players {
			p.Ready = false
			p.Index, p.Correct, p.Done, p.Finished = 0, 0, false, time.Time{}
		}
		game.Tasks = nil
		game.Note = player.Name + " will noch eine Runde – alle bitte BEREIT drücken."

	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	antwort(nil)
}

// allGone meldet, dass niemand mehr am Tisch sitzt.
func (g *RechenGame) allGone(r *Room) bool {
	present := 0
	for _, p := range g.Players {
		if _, ok := r.Players[p.ID]; ok {
			present++
		}
	}
	return present == 0
}

// removeSeat nimmt einen Spieler aus dem Rennen (für /api/leave).
func rechenRemoveSeat(g *RechenGame, id, name string) (bool, bool, int) {
	seat := rechenSeat(g, id)
	if seat < 0 {
		return false, false, len(g.Players)
	}
	g.Players = append(g.Players[:seat], g.Players[seat+1:]...)
	if len(g.Players) < rechenMinPlayers {
		return true, true, len(g.Players)
	}
	if g.HostID == id {
		g.HostID = g.Players[0].ID
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	if g.Status == "active" {
		alleFertig := true
		for _, p := range g.Players {
			if !p.Done {
				alleFertig = false
				break
			}
		}
		if alleFertig {
			g.Status = "finished"
			best := g.ranking()[0]
			g.Winner = best.Name
			g.Note = best.Name + " gewinnt."
		}
	}
	return true, false, len(g.Players)
}
