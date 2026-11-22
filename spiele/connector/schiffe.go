package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Schiffe versenken für zwei Spieler: Jeder setzt eine Flotte auf sein eigenes
// 10x10-Feld, danach wird abwechselnd geschossen. Ein Schiff wird erst komplett
// aufgedeckt, wenn es versenkt ist – bis dahin sieht der Gegner nur Treffer.

const (
	schiffeGroesse    = 10
	schiffeMaxSpieler = 2
)

// schiffeFlotte ist die Flotte: Schlachtschiff, Kreuzer, zwei Zerstörer, U-Boot.
var schiffeFlotte = []int{5, 4, 3, 3, 2}

type SchiffeSchiff struct {
	X          int  `json:"x"`
	Y          int  `json:"y"`
	Laenge     int  `json:"laenge"`
	Waagerecht bool `json:"waagerecht"`
	Treffer    int  `json:"treffer"`
	Versenkt   bool `json:"versenkt"`
}

type SchiffeSpieler struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Ready    bool            `json:"ready"`
	Fertig   bool            `json:"fertig"`
	Schiffe  []SchiffeSchiff `json:"-"`
	Feld     [][]int         `json:"-"` // -1 Wasser, sonst Index des Schiffes
	Beschuss [][]int         `json:"-"` // 0 nichts, 1 Wasser, 2 Treffer, 3 versenkt
}

type SchiffeGame struct {
	ID       string           `json:"id"`
	HostID   string           `json:"hostId"`
	Status   string           `json:"status"` // invited, setzen, spiel, ende
	Spieler  []SchiffeSpieler `json:"spieler"`
	Zug      int              `json:"zug"`
	Meldung  string           `json:"meldung"`
	Art      string           `json:"art"` // treffer, versenkt, wasser, info
	Note     string           `json:"note"`
	Sieger   string           `json:"sieger,omitempty"`
	BotID    string           `json:"botId,omitempty"`
	BotStufe string           `json:"-"`
	BotFrist time.Time        `json:"-"`
}

type SchiffeSpielerSicht struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Ready    bool   `json:"ready"`
	Fertig   bool   `json:"fertig"`
	Schiffe  int    `json:"schiffe"`
	Treffer  int    `json:"treffer"`
	Versenkt int    `json:"versenkt"`
}

type SchiffeAnsicht struct {
	ID      string                `json:"id"`
	HostID  string                `json:"hostId"`
	Status  string                `json:"status"`
	Zug     int                   `json:"zug"`
	ZugID   string                `json:"zugId"`
	Spieler []SchiffeSpielerSicht `json:"spieler"`
	Eigen   SchiffeEigen          `json:"eigen"`
	Feind   SchiffeFeind          `json:"feind"`
	Meldung string                `json:"meldung"`
	Art     string                `json:"art"`
	Note    string                `json:"note"`
	Sieger  string                `json:"sieger,omitempty"`
}

type SchiffeEigen struct {
	Schiffe  []SchiffeSchiff `json:"schiffe"`
	Beschuss [][]int         `json:"beschuss"`
	Fertig   bool            `json:"fertig"`
}

type SchiffeFeind struct {
	Name     string  `json:"name"`
	Beschuss [][]int `json:"beschuss"`
}

// ---------------------------------------------------------------- Aufbau

func schiffeLeeresFeld() [][]int {
	feld := make([][]int, schiffeGroesse)
	for i := range feld {
		feld[i] = make([]int, schiffeGroesse)
		for j := range feld[i] {
			feld[i][j] = -1
		}
	}
	return feld
}

func schiffeLeererBeschuss() [][]int {
	feld := make([][]int, schiffeGroesse)
	for i := range feld {
		feld[i] = make([]int, schiffeGroesse)
	}
	return feld
}

func schiffeKopie(feld [][]int) [][]int {
	kopie := make([][]int, len(feld))
	for i := range feld {
		kopie[i] = append([]int{}, feld[i]...)
	}
	return kopie
}

func schiffeNeuerSpieler(id, name string) SchiffeSpieler {
	return SchiffeSpieler{
		ID:       id,
		Name:     name,
		Feld:     schiffeLeeresFeld(),
		Beschuss: schiffeLeererBeschuss(),
	}
}

func schiffeSitz(g *SchiffeGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// schiffeZellen liefert alle Felder eines Schiffes.
func schiffeZellen(s SchiffeSchiff) [][2]int {
	zellen := make([][2]int, 0, s.Laenge)
	for i := 0; i < s.Laenge; i++ {
		if s.Waagerecht {
			zellen = append(zellen, [2]int{s.X + i, s.Y})
		} else {
			zellen = append(zellen, [2]int{s.X, s.Y + i})
		}
	}
	return zellen
}

// schiffeSetzen prüft eine Flotte und trägt sie ins eigene Feld ein.
func schiffeSetzen(g *SchiffeGame, sitz int, schiffe []SchiffeSchiff) string {
	if len(schiffe) != len(schiffeFlotte) {
		return "Die Flotte besteht aus " + strconv.Itoa(len(schiffeFlotte)) + " Schiffen."
	}
	soll := append([]int{}, schiffeFlotte...)
	ist := make([]int, 0, len(schiffe))
	for _, s := range schiffe {
		ist = append(ist, s.Laenge)
	}
	sort.Ints(soll)
	sort.Ints(ist)
	for i := range soll {
		if soll[i] != ist[i] {
			return "Die Schiffsgrößen passen nicht – erwartet sind " + schiffeFlotteText() + "."
		}
	}

	spieler := &g.Spieler[sitz]
	feld := schiffeLeeresFeld()
	for i, s := range schiffe {
		if s.Laenge < 1 {
			return "Ungültige Schiffsgröße."
		}
		for _, z := range schiffeZellen(s) {
			x, y := z[0], z[1]
			if x < 0 || y < 0 || x >= schiffeGroesse || y >= schiffeGroesse {
				return "Ein Schiff ragt über den Rand hinaus."
			}
			if feld[y][x] != -1 {
				return "Die Schiffe dürfen sich nicht berühren."
			}
			feld[y][x] = i
		}
	}
	spieler.Schiffe = make([]SchiffeSchiff, len(schiffe))
	for i, s := range schiffe {
		s.Treffer = 0
		s.Versenkt = false
		spieler.Schiffe[i] = s
	}
	spieler.Feld = feld
	spieler.Fertig = true
	return ""
}

func schiffeFlotteText() string {
	teile := make([]string, 0, len(schiffeFlotte))
	for _, l := range schiffeFlotte {
		teile = append(teile, strconv.Itoa(l))
	}
	return strings.Join(teile, ", ")
}

func schiffeAlleFertig(g *SchiffeGame) bool {
	if len(g.Spieler) < schiffeMaxSpieler {
		return false
	}
	for _, p := range g.Spieler {
		if !p.Fertig {
			return false
		}
	}
	return true
}

func schiffeVersenktAlle(g *SchiffeGame, sitz int) bool {
	for _, s := range g.Spieler[sitz].Schiffe {
		if !s.Versenkt {
			return false
		}
	}
	return true
}

// schiffeSchuss verarbeitet einen Schuss und liefert die Meldung.
func schiffeSchuss(g *SchiffeGame, sitz, x, y int) string {
	gegner := 1 - sitz
	ziel := &g.Spieler[gegner]
	if ziel.Beschuss[y][x] != 0 {
		return ""
	}
	idx := ziel.Feld[y][x]
	if idx < 0 {
		ziel.Beschuss[y][x] = 1
		g.Art = "wasser"
		return "Pech gehabt – nur Wasser bei " + schiffeFeldName(x, y) + ". Jetzt ist " + g.Spieler[gegner].Name + " dran."
	}

	schiff := &ziel.Schiffe[idx]
	schiff.Treffer++
	ziel.Beschuss[y][x] = 2
	g.Art = "treffer"
	meldung := "Volltreffer auf " + schiffeFeldName(x, y) + "! Stark geschossen."
	if schiff.Treffer >= schiff.Laenge {
		schiff.Versenkt = true
		for _, z := range schiffeZellen(*schiff) {
			if ziel.Beschuss[z[1]][z[0]] == 2 {
				ziel.Beschuss[z[1]][z[0]] = 3
			}
		}
		g.Art = "versenkt"
		meldung = "Versenkt! Das " + schiffeSchiffName(schiff.Laenge) + " liegt auf dem Grund – du darfst gleich noch einmal."
	} else {
		meldung += " Du darfst noch einmal."
	}
	return meldung
}

func schiffeSchiffName(laenge int) string {
	switch laenge {
	case 5:
		return "Schlachtschiff"
	case 4:
		return "Kreuzer"
	case 3:
		return "Zerstörer"
	case 2:
		return "U-Boot"
	}
	return "Schiff"
}

func schiffeFeldName(x, y int) string {
	return string(rune('A'+x)) + strconv.Itoa(y+1)
}

// ---------------------------------------------------------------- Sicht

func schiffeAnsicht(g *SchiffeGame, id string) *SchiffeAnsicht {
	if g == nil {
		return nil
	}
	sitz := schiffeSitz(g, id)
	a := &SchiffeAnsicht{
		ID:      g.ID,
		HostID:  g.HostID,
		Status:  g.Status,
		Zug:     g.Zug,
		Meldung: g.Meldung,
		Art:     g.Art,
		Note:    g.Note,
		Sieger:  g.Sieger,
	}
	if g.Zug >= 0 && g.Zug < len(g.Spieler) {
		a.ZugID = g.Spieler[g.Zug].ID
	}
	for _, p := range g.Spieler {
		sicht := SchiffeSpielerSicht{ID: p.ID, Name: p.Name, Ready: p.Ready, Fertig: p.Fertig, Schiffe: len(p.Schiffe)}
		for _, s := range p.Schiffe {
			sicht.Treffer += s.Treffer
			if s.Versenkt {
				sicht.Versenkt++
			}
		}
		a.Spieler = append(a.Spieler, sicht)
	}
	if sitz >= 0 {
		eigen := g.Spieler[sitz]
		a.Eigen = SchiffeEigen{
			Schiffe:  append([]SchiffeSchiff{}, eigen.Schiffe...),
			Beschuss: schiffeKopie(eigen.Beschuss),
			Fertig:   eigen.Fertig,
		}
		if len(g.Spieler) > 1 {
			feind := g.Spieler[1-sitz]
			a.Feind = SchiffeFeind{Name: feind.Name, Beschuss: schiffeKopie(feind.Beschuss)}
		}
	}
	return a
}

// ---------------------------------------------------------------- Endpunkt

func schiffeHandler(w http.ResponseWriter, r *http.Request) {
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
		schiffeBotTreiben(room.Schiffe, time.Now())
		game := schiffeAnsicht(room.Schiffe, id)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID      string          `json:"id"`
		X       int             `json:"x"`
		Y       int             `json:"y"`
		Schiffe []SchiffeSchiff `json:"schiffe"`
		Stufe   string          `json:"stufe"`
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

	zeit := time.Now().Format("15:04:05")

	switch action {
	case "create":
		if room.Schiffe != nil && room.Schiffe.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Partie Schiffe versenken.", http.StatusConflict)
			return
		}
		room.Schiffe = &SchiffeGame{
			ID:      randomCode(),
			HostID:  input.ID,
			Status:  "invited",
			Zug:     -1,
			Spieler: []SchiffeSpieler{schiffeNeuerSpieler(input.ID, player.Name)},
			Note:    "Tisch eröffnet – gesucht wird ein zweiter Spieler.",
		}
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Partie Schiffe versenken – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Schiffe
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if schiffeSitz(game, input.ID) >= 0 {
			break
		}
		if len(game.Spieler) >= schiffeMaxSpieler {
			mu.Unlock()
			http.Error(w, "Schiffe versenken ist ein Duell – der Tisch ist voll.", http.StatusConflict)
			return
		}
		game.Spieler = append(game.Spieler, schiffeNeuerSpieler(input.ID, player.Name))
		game.Note = player.Name + " sitzt mit am Tisch. Beide drücken BEREIT."

	case "ready":
		game := room.Schiffe
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := schiffeSitz(game, input.ID)
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
		game := room.Schiffe
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Partie auf den Start.", http.StatusConflict)
			return
		}
		if len(game.Spieler) < schiffeMaxSpieler {
			mu.Unlock()
			http.Error(w, "Schiffe versenken braucht zwei Spieler.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber startet die Partie.", http.StatusForbidden)
			return
		}
		for _, p := range game.Spieler {
			if !p.Ready {
				mu.Unlock()
				http.Error(w, "Erst müssen beide BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.Status = "setzen"
		game.Zug = -1
		game.Art = "info"
		game.Meldung = ""
		game.Note = "Setzt eure Flotten – Schiffe anklicken, mit DREHEN die Richtung wechseln."

	case "setzen":
		game := room.Schiffe
		if game == nil || game.Status != "setzen" {
			mu.Unlock()
			http.Error(w, "Gerade setzt niemand Schiffe.", http.StatusConflict)
			return
		}
		sitz := schiffeSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		if game.Spieler[sitz].Fertig {
			mu.Unlock()
			http.Error(w, "Deine Flotte steht schon.", http.StatusConflict)
			return
		}
		if fehler := schiffeSetzen(game, sitz, input.Schiffe); fehler != "" {
			mu.Unlock()
			http.Error(w, fehler, http.StatusBadRequest)
			return
		}
		game.Note = player.Name + " hat die Flotte gesetzt."
		if schiffeAlleFertig(game) {
			game.Status = "spiel"
			game.Zug = 0
			game.Art = "info"
			game.Meldung = "Beide Flotten stehen – " + game.Spieler[0].Name + " beginnt."
			game.Note = "Es wird abwechselnd geschossen: Treffer heißt noch einmal, Wasser gibt den Zug ab."
		}

	case "schiessen":
		game := room.Schiffe
		if game == nil || game.Status != "spiel" {
			mu.Unlock()
			http.Error(w, "Gerade läuft kein Gefecht.", http.StatusConflict)
			return
		}
		sitz := schiffeSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		if game.Zug != sitz {
			mu.Unlock()
			http.Error(w, "Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		if input.X < 0 || input.Y < 0 || input.X >= schiffeGroesse || input.Y >= schiffeGroesse {
			mu.Unlock()
			http.Error(w, "Dieses Feld gibt es nicht.", http.StatusBadRequest)
			return
		}
		if game.Spieler[1-sitz].Beschuss[input.Y][input.X] != 0 {
			mu.Unlock()
			http.Error(w, "Auf dieses Feld wurde schon geschossen.", http.StatusConflict)
			return
		}
		meldung := schiffeSchuss(game, sitz, input.X, input.Y)
		game.Meldung = meldung
		game.Note = player.Name + " schießt auf " + schiffeFeldName(input.X, input.Y) + "."
		if schiffeVersenktAlle(game, 1-sitz) {
			game.Status = "ende"
			game.Sieger = game.Spieler[sitz].Name
			game.Meldung = "Alle Schiffe versenkt – " + game.Spieler[sitz].Name + " gewinnt die Partie!"
			game.Art = "versenkt"
			game.Zug = -1
			game.Note = player.Name + " hat die gegnerische Flotte vernichtet."
		} else if game.Art == "wasser" {
			game.Zug = 1 - sitz
		}

	case "bot":
		room.botNamenNeu()
		stufe := botStufe(input.Stufe)
		if room.Schiffe != nil && room.Schiffe.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Partie Schiffe versenken.", http.StatusConflict)
			return
		}
		game := &SchiffeGame{
			ID: randomCode(), HostID: input.ID, Status: "setzen", Zug: -1,
			Spieler: []SchiffeSpieler{
				schiffeNeuerSpieler(input.ID, player.Name),
				schiffeNeuerSpieler(botID(1), room.botName(1)),
			},
			BotID: botID(1), BotStufe: stufe,
		}
		game.Spieler[0].Ready = true
		game.Spieler[1].Ready = true
		if fehler := schiffeSetzen(game, 1, schiffeBotFlotte()); fehler != "" {
			mu.Unlock()
			http.Error(w, fehler, http.StatusInternalServerError)
			return
		}
		game.Note = player.Name + " setzt gegen " + room.botName(1) + " (" + botStufeText(stufe) + ") die eigene Flotte."
		room.Schiffe = game
		room.addMessage(Message{
			Name: player.Name, Time: zeit, SenderID: input.ID,
			Text: "spielt Schiffe versenken gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
		})

	case "reset":
		game := room.Schiffe
		if game == nil || game.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Eine neue Partie geht erst nach dem Ende.", http.StatusConflict)
			return
		}
		if schiffeSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		frisch := &SchiffeGame{
			ID:      game.ID,
			HostID:  game.HostID,
			Status:  "invited",
			Zug:     -1,
			Spieler: []SchiffeSpieler{},
			Note:    "Neue Partie – erst BEREIT drücken.",
		}
		for _, p := range game.Spieler {
			frisch.Spieler = append(frisch.Spieler, schiffeNeuerSpieler(p.ID, p.Name))
		}
		room.Schiffe = frisch

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	if room.Schiffe != nil && room.Schiffe.BotID != "" && input.ID != room.Schiffe.BotID {
		room.Schiffe.BotFrist = time.Time{}
	}
	room.LastActive = time.Now()
	result := schiffeAnsicht(room.Schiffe, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
