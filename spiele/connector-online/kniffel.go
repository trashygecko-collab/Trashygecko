package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Kniffel: 5 Würfel, drei Würfe pro Runde, 13 Kategorien, 2 bis 6 Spieler.
const kniffelMaxPlayers = 6

var kniffelCategories = []string{
	"ones", "twos", "threes", "fours", "fives", "sixes",
	"three", "four", "full", "small", "large", "kniffel", "chance",
}

var kniffelColors = []string{"#42f5ff", "#ff4fd8", "#ffe66d", "#62ff9b", "#b18cff", "#ff985e"}

type KniffelPlayer struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Color        string          `json:"color"`
	Scores       map[string]*int `json:"scores"`
	KniffelBonus int             `json:"kniffelBonus"`
	Ready        bool            `json:"ready"`
}

type KniffelGame struct {
	ID       string          `json:"id"`
	HostID   string          `json:"hostId"`
	BotID    string          `json:"botId,omitempty"`
	BotStufe string          `json:"-"`
	BotFrist time.Time       `json:"-"`
	Players  []KniffelPlayer `json:"players"`
	Current  int             `json:"current"`
	Dice     []int           `json:"dice"`
	Rolls    int             `json:"rolls"`
	Status   string          `json:"status"` // invited, active, finished
	Winner   string          `json:"winner,omitempty"`
	Note     string          `json:"note,omitempty"`
}

func kniffelSum(dice []int) int {
	sum := 0
	for _, v := range dice {
		sum += v
	}
	return sum
}

func kniffelSumFace(dice []int, face int) int {
	count := 0
	for _, v := range dice {
		if v == face {
			count++
		}
	}
	return count * face
}

func kniffelSameAtLeast(dice []int, n int) bool {
	for face := 1; face <= 6; face++ {
		count := 0
		for _, v := range dice {
			if v == face {
				count++
			}
		}
		if count >= n {
			return true
		}
	}
	return false
}

func kniffelFullHouse(dice []int) bool {
	counts := []int{}
	for face := 1; face <= 6; face++ {
		count := 0
		for _, v := range dice {
			if v == face {
				count++
			}
		}
		if count > 0 {
			counts = append(counts, count)
		}
	}
	if len(counts) != 2 {
		return false
	}
	return (counts[0] == 3 && counts[1] == 2) || (counts[0] == 2 && counts[1] == 3)
}

func kniffelStraight(dice []int, need int) bool {
	seen := make([]bool, 7)
	for _, v := range dice {
		if v >= 1 && v <= 6 {
			seen[v] = true
		}
	}
	best := 0
	run := 0
	for face := 1; face <= 6; face++ {
		if seen[face] {
			run++
			if run > best {
				best = run
			}
		} else {
			run = 0
		}
	}
	return best >= need
}

func kniffelScoreFor(category string, dice []int) int {
	switch category {
	case "ones":
		return kniffelSumFace(dice, 1)
	case "twos":
		return kniffelSumFace(dice, 2)
	case "threes":
		return kniffelSumFace(dice, 3)
	case "fours":
		return kniffelSumFace(dice, 4)
	case "fives":
		return kniffelSumFace(dice, 5)
	case "sixes":
		return kniffelSumFace(dice, 6)
	case "three":
		if kniffelSameAtLeast(dice, 3) {
			return kniffelSum(dice)
		}
		return 0
	case "four":
		if kniffelSameAtLeast(dice, 4) {
			return kniffelSum(dice)
		}
		return 0
	case "full":
		if kniffelFullHouse(dice) {
			return 25
		}
		return 0
	case "small":
		if kniffelStraight(dice, 4) {
			return 30
		}
		return 0
	case "large":
		if kniffelStraight(dice, 5) {
			return 40
		}
		return 0
	case "kniffel":
		if kniffelSameAtLeast(dice, 5) {
			return 50
		}
		return 0
	case "chance":
		return kniffelSum(dice)
	}
	return 0
}

func kniffelIsCategory(id string) bool {
	for _, c := range kniffelCategories {
		if c == id {
			return true
		}
	}
	return false
}

func kniffelNewPlayer(id, name string, seat int) KniffelPlayer {
	scores := map[string]*int{}
	for _, c := range kniffelCategories {
		scores[c] = nil
	}
	return KniffelPlayer{
		ID:     id,
		Name:   name,
		Color:  kniffelColors[seat%len(kniffelColors)],
		Scores: scores,
	}
}

func kniffelUpper(p KniffelPlayer) int {
	sum := 0
	for _, c := range kniffelCategories[:6] {
		if p.Scores[c] != nil {
			sum += *p.Scores[c]
		}
	}
	return sum
}

func kniffelTotal(p KniffelPlayer) int {
	upper := kniffelUpper(p)
	total := upper + p.KniffelBonus
	if upper >= 63 {
		total += 35
	}
	for _, c := range kniffelCategories[6:] {
		if p.Scores[c] != nil {
			total += *p.Scores[c]
		}
	}
	return total
}

func kniffelRemaining(p KniffelPlayer) int {
	count := 0
	for _, c := range kniffelCategories {
		if p.Scores[c] == nil {
			count++
		}
	}
	return count
}

// kniffelWurf würfelt alle nicht gehaltenen Würfel neu. Die Prüfungen
// (Reihenfolge, Anzahl) bleiben im Handler; der Bot nutzt dieselbe Mechanik.
func kniffelWurf(g *KniffelGame, kept []bool) {
	if g.Rolls == 0 {
		g.Dice = make([]int, 5)
	}
	for i := 0; i < 5 && i < len(kept); i++ {
		if !kept[i] {
			g.Dice[i] = rand.Intn(6) + 1
		}
	}
	g.Rolls++
}

// kniffelEintrag trägt eine Kategorie ein, vergibt den Kniffel-Zusatzbonus und
// wechselt zum nächsten Spieler bzw. beendet die Partie. wert erzwingt beim
// Streichen eine 0.
func kniffelEintrag(g *KniffelGame, seat int, category string, wert *int) int {
	points := kniffelScoreFor(category, g.Dice)
	if wert != nil {
		points = *wert
	}
	g.Players[seat].Scores[category] = &points
	if category != "kniffel" && g.Players[seat].Scores["kniffel"] != nil &&
		*g.Players[seat].Scores["kniffel"] == 50 && kniffelSameAtLeast(g.Dice, 5) {
		g.Players[seat].KniffelBonus += 100
	}
	g.Dice = make([]int, 5)
	g.Rolls = 0

	allDone := true
	for _, p := range g.Players {
		if kniffelRemaining(p) > 0 {
			allDone = false
			break
		}
	}
	if allDone {
		g.Status = "finished"
		best := -1
		for _, p := range g.Players {
			if total := kniffelTotal(p); total > best {
				best = total
				g.Winner = p.Name
			}
		}
		g.Note = "Spiel beendet – " + g.Winner + " gewinnt."
		return points
	}
	for step := 1; step <= len(g.Players); step++ {
		next := (seat + step) % len(g.Players)
		if kniffelRemaining(g.Players[next]) > 0 {
			g.Current = next
			break
		}
	}
	return points
}

func kniffelFindPlayer(g *KniffelGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func cloneKniffel(g *KniffelGame) *KniffelGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Dice = append([]int(nil), g.Dice...)
	c.Players = make([]KniffelPlayer, len(g.Players))
	for i, p := range g.Players {
		scores := map[string]*int{}
		for k, v := range p.Scores {
			if v == nil {
				scores[k] = nil
				continue
			}
			value := *v
			scores[k] = &value
		}
		p.Scores = scores
		c.Players[i] = p
	}
	return &c
}

func kniffelHandler(w http.ResponseWriter, r *http.Request) {
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
		if room.Kniffel != nil {
			kniffelBotTreiben(room.Kniffel, time.Now())
		}
		game := cloneKniffel(room.Kniffel)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID       string `json:"id"`
		Stufe    string `json:"stufe"`
		Category string `json:"category"`
		// Value erzwingt einen festen Wert – beim Streichen 0.
		Value *int   `json:"value"`
		Kept  []bool `json:"kept"`
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
	case "create":
		if room.Kniffel != nil && room.Kniffel.Status != "finished" {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Kniffel-Runde.", http.StatusConflict)
			return
		}
		game := &KniffelGame{
			ID:     randomCode(),
			HostID: input.ID,
			Status: "invited",
			Dice:   make([]int, 5),
		}
		game.Players = append(game.Players, kniffelNewPlayer(input.ID, player.Name, 0))
		game.Note = player.Name + " eröffnet eine Kniffel-Runde."
		room.Kniffel = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Kniffel-Runde – bis zu 6 Spieler, mitspielen über die Spielkarte.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "join":
		game := room.Kniffel
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine offene Kniffel-Runde.", http.StatusConflict)
			return
		}
		if kniffelFindPlayer(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= kniffelMaxPlayers {
			mu.Unlock()
			http.Error(w, "Diese Kniffel-Runde ist mit 6 Spielern voll.", http.StatusConflict)
			return
		}
		game.Players = append(game.Players, kniffelNewPlayer(input.ID, player.Name, len(game.Players)))
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zur Kniffel-Runde.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "bot":
		if room.Kniffel != nil && room.Kniffel.Status != "finished" {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Kniffel-Runde.", http.StatusConflict)
			return
		}
		stufe := botStufe(input.Stufe)
		g := &KniffelGame{
			ID:       randomCode(),
			HostID:   input.ID,
			Status:   "active",
			Dice:     make([]int, 5),
			BotID:    botID(1),
			BotStufe: stufe,
		}
		g.Players = append(g.Players, kniffelNewPlayer(input.ID, player.Name, 0))
		g.Players = append(g.Players, kniffelNewPlayer(g.BotID, botName(1), 1))
		g.Players[0].Ready = true
		g.Players[1].Ready = true
		g.Note = player.Name + " spielt gegen " + botName(1) + " (" + botStufeText(stufe) + ")."
		room.Kniffel = g
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Kniffel gegen " + botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "start":
		game := room.Kniffel
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Kniffel-Runde auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber kann die Runde starten.", http.StatusForbidden)
			return
		}
		if len(game.Players) < 2 {
			mu.Unlock()
			http.Error(w, "Kniffel braucht mindestens zwei Spieler am Tisch.", http.StatusConflict)
			return
		}
		for _, p := range game.Players {
			if !p.Ready {
				mu.Unlock()
				http.Error(w, "Erst müssen alle Spieler BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.Status = "active"
		game.Current = 0
		game.Dice = make([]int, 5)
		game.Rolls = 0
		game.Note = "Das Spiel beginnt! " + game.Players[0].Name + " würfelt zuerst."

	case "ready":
		game := room.Kniffel
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur, solange die Runde noch nicht läuft.", http.StatusConflict)
			return
		}
		seat := kniffelFindPlayer(game, input.ID)
		if seat < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		game.Players[seat].Ready = !game.Players[seat].Ready
		if game.Players[seat].Ready {
			game.Note = player.Name + " ist bereit."
		} else {
			game.Note = player.Name + " wartet noch."
		}

	case "roll":
		game := room.Kniffel
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		seat := kniffelFindPlayer(game, input.ID)
		if seat < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Current != seat {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		if game.Rolls >= 3 {
			mu.Unlock()
			http.Error(w, "Du hast schon dreimal gewürfelt – jetzt Punkte eintragen.", http.StatusConflict)
			return
		}
		if len(input.Kept) != 5 {
			mu.Unlock()
			http.Error(w, "Es müssen fünf Würfel sein.", http.StatusBadRequest)
			return
		}

		kniffelWurf(game, input.Kept)
		game.Note = player.Name + " würfelt (Wurf " + strconv.Itoa(game.Rolls) + " von 3)."

	case "score":
		game := room.Kniffel
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		seat := kniffelFindPlayer(game, input.ID)
		if seat < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Current != seat {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		if game.Rolls == 0 {
			mu.Unlock()
			http.Error(w, "Du musst zuerst würfeln.", http.StatusConflict)
			return
		}

		category := strings.ToLower(strings.TrimSpace(input.Category))
		if category == "" {
			// Ohne Auswahl: normalerweise das Feld mit den meisten Punkten,
			// beim Streichen (Wert 0) das Feld, das am wenigsten bringt.
			besterWert := -1
			for _, c := range kniffelCategories {
				if game.Players[seat].Scores[c] != nil {
					continue
				}
				punkte := kniffelScoreFor(c, game.Dice)
				nehmen := punkte > besterWert
				if input.Value != nil {
					nehmen = besterWert < 0 || punkte < besterWert
				}
				if nehmen {
					besterWert = punkte
					category = c
				}
			}
		}
		if !kniffelIsCategory(category) {
			mu.Unlock()
			http.Error(w, "Unbekannte Kategorie.", http.StatusBadRequest)
			return
		}
		if game.Players[seat].Scores[category] != nil {
			mu.Unlock()
			http.Error(w, "Diese Kategorie ist schon eingetragen.", http.StatusConflict)
			return
		}

		points := kniffelEintrag(game, seat, category, input.Value)
		if game.Status == "active" {
			game.Note = player.Name + " trägt " + strconv.Itoa(points) + " Punkte ein."
		}

	case "reset":
		game := room.Kniffel
		if game == nil || game.Status != "finished" {
			mu.Unlock()
			http.Error(w, "Ein neues Spiel ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		if kniffelFindPlayer(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		// Eine neue Runde beginnt wieder am Tisch: alle drücken erst BEREIT.
		fresh := &KniffelGame{
			ID:       game.ID,
			HostID:   game.HostID,
			Status:   "invited",
			Dice:     make([]int, 5),
			BotID:    game.BotID,
			BotStufe: game.BotStufe,
		}
		for i, p := range game.Players {
			fresh.Players = append(fresh.Players, kniffelNewPlayer(p.ID, p.Name, i))
		}
		for i := range fresh.Players {
			if fresh.Players[i].ID == fresh.BotID {
				fresh.Players[i].Ready = true
			}
		}
		fresh.Note = "Neue Runde – alle drücken BEREIT, dann startet der Gastgeber."
		room.Kniffel = fresh

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Kniffel-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneKniffel(room.Kniffel)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
