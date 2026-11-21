package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Lügen Mäxchen: zwei Würfel, Ansage mit Bluff. Der Server würfelt und merkt
// sich den echten Wurf; nur der Würfler selbst bekommt ihn zu sehen, bis
// jemand aufdeckt. Jeder startet mit 3 €, wer eine Runde verliert zahlt 1 €,
// bei 0 € ist er raus – der Letzte gewinnt.
const mxMaxPlayers = 6
const mxStartEuro = 3

var mxColors = []string{"#e8cd8e", "#7fb98a", "#d98878", "#8aa9d6", "#b18cff", "#ff985e"}

type MXPlayer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Euro  int    `json:"euro"`
	Ready bool   `json:"ready"`
	Out   bool   `json:"out"`
}

// MXCombo ist eine Ansage: Hoch ist die größere, Tief die kleinere Augenzahl.
type MXCombo struct {
	Hoch int `json:"hoch"`
	Tief int `json:"tief"`
}

func (c MXCombo) String() string {
	return strconv.Itoa(c.Hoch) + "-" + strconv.Itoa(c.Tief)
}

// MXReveal hält das Ergebnis eines Aufdeckens fest.
type MXReveal struct {
	Nr          int     `json:"nr"`
	VonID       string  `json:"vonId"`
	VonName     string  `json:"vonName"`
	Angesagt    MXCombo `json:"angesagt"`
	Gewuerfelt  MXCombo `json:"gewuerfelt"`
	Gelogen     bool    `json:"gelogen"`
	Aufdecker   string  `json:"aufdecker"`
	AufdeckerID string  `json:"aufdeckerId"`
	Verlierer   string  `json:"verlierer"`
	VerliererID string  `json:"verliererId"`
	Aus         bool    `json:"aus"`
	Euro        int     `json:"euro"`
	Start       string  `json:"start,omitempty"` // wer die neue Runde beginnt
}

type MaexchenGame struct {
	ID            string     `json:"id"`
	HostID        string     `json:"hostId"`
	Players       []MXPlayer `json:"players"`
	Current       int        `json:"current"`
	Status        string     `json:"status"` // invited, active, finished
	Phase         string     `json:"phase"`  // wurf, ansage
	Runde         int        `json:"runde"`
	Ansage        *MXCombo   `json:"ansage,omitempty"`
	AnsageVon     string     `json:"ansageVon,omitempty"`
	AnsageVonName string     `json:"ansageVonName,omitempty"`
	Dice          []int      `json:"dice,omitempty"` // nur für den Würfler
	DiceOwner     string     `json:"diceOwner,omitempty"`
	Reveal        *MXReveal  `json:"reveal,omitempty"`
	LastLiar      string     `json:"lastLiar,omitempty"`
	Winner        string     `json:"winner,omitempty"`
	Note          string     `json:"note,omitempty"`

	// Wurf zur stehenden Ansage – für die Prüfung beim Aufdecken, nie im JSON
	// an andere Spieler.
	wurf     []int
	revealNr int
}

// mxCombo ordnet ein Würfelpaar so, dass Hoch >= Tief ist.
func mxCombo(a, b int) MXCombo {
	if a < b {
		a, b = b, a
	}
	return MXCombo{Hoch: a, Tief: b}
}

// mxRang ist die Rangfolge: Mäxchen (2-1) schlägt alles, dann die Pasche,
// dann die übrigen Kombinationen nach Augenzahl (6-5 vor 6-4).
func mxRang(c MXCombo) int {
	if c.Hoch == 2 && c.Tief == 1 {
		return 1000
	}
	if c.Hoch == c.Tief {
		return 500 + c.Hoch
	}
	return c.Hoch*10 + c.Tief
}

func mxGueltig(c MXCombo) bool {
	return c.Hoch >= 1 && c.Hoch <= 6 && c.Tief >= 1 && c.Tief <= 6 && c.Hoch >= c.Tief
}

func mxNewPlayer(id, name string, seat int) MXPlayer {
	return MXPlayer{ID: id, Name: name, Color: mxColors[seat%len(mxColors)], Euro: mxStartEuro}
}

func mxFind(g *MaexchenGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func mxAktiv(g *MaexchenGame) int {
	n := 0
	for _, p := range g.Players {
		if !p.Out {
			n++
		}
	}
	return n
}

// mxNaechsterAktiv liefert den nächsten Spieler im Uhrzeigersinn, der noch
// Euro hat.
func mxNaechsterAktiv(g *MaexchenGame, ab int) int {
	for i := 1; i <= len(g.Players); i++ {
		k := (ab + i) % len(g.Players)
		if !g.Players[k].Out {
			return k
		}
	}
	return ab
}

// mxNeueRunde räumt Ansage und Würfel ab; der Verlierer beginnt.
func mxNeueRunde(g *MaexchenGame, start int) {
	g.Runde++
	g.Ansage = nil
	g.AnsageVon = ""
	g.AnsageVonName = ""
	g.Dice = nil
	g.DiceOwner = ""
	g.wurf = nil
	g.Phase = "wurf"
	g.Current = start
}

func mxWuerfle(g *MaexchenGame) MXCombo {
	a := rand.Intn(6) + 1
	b := rand.Intn(6) + 1
	g.wurf = []int{a, b}
	g.Dice = []int{a, b}
	return mxCombo(a, b)
}

// mxView ist der Stand für einen bestimmten Spieler: fremde Würfel bleiben
// verborgen.
func mxView(g *MaexchenGame, viewer string) *MaexchenGame {
	c := g.clone()
	if c == nil {
		return nil
	}
	if c.DiceOwner != viewer {
		c.Dice = nil
	}
	return c
}

func (g *MaexchenGame) clone() *MaexchenGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Players = append([]MXPlayer(nil), g.Players...)
	c.Dice = append([]int(nil), g.Dice...)
	c.wurf = append([]int(nil), g.wurf...)
	if g.Ansage != nil {
		a := *g.Ansage
		c.Ansage = &a
	}
	if g.Reveal != nil {
		r := *g.Reveal
		c.Reveal = &r
	}
	return &c
}

// mxRaus nimmt einen Platz aus der Partie. Bei weniger als zwei Spielern ist
// der Tisch zu Ende.
func mxRaus(g *MaexchenGame, id, name string) (bool, bool, int) {
	idx := mxFind(g, id)
	if idx < 0 {
		return false, false, len(g.Players)
	}
	g.Players = append(g.Players[:idx], g.Players[idx+1:]...)
	if len(g.Players) < 2 {
		return true, true, len(g.Players)
	}
	if g.HostID == id {
		g.HostID = g.Players[0].ID
	}
	if g.Current > idx {
		g.Current--
	}
	if g.Current >= len(g.Players) {
		g.Current = 0
	}
	if g.Players[g.Current].Out {
		g.Current = mxNaechsterAktiv(g, g.Current)
	}
	if g.AnsageVon == id {
		// Der Ansager ist weg – die Ansage verfällt, es wird neu gewürfelt.
		mxNeueRunde(g, g.Current)
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}

func maexchenHandler(w http.ResponseWriter, r *http.Request) {
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
		game := mxView(room.Maexchen, id)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID     string   `json:"id"`
		Ansage *MXCombo `json:"ansage"`
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

	fehler := func(status int, text string) {
		mu.Unlock()
		http.Error(w, text, status)
	}

	switch action {
	case "create":
		if room.Maexchen != nil && room.Maexchen.Status != "finished" {
			fehler(http.StatusConflict, "In diesem Spielraum läuft bereits eine Mäxchen-Runde.")
			return
		}
		game := &MaexchenGame{ID: randomCode(), HostID: input.ID, Status: "invited", Phase: "wurf", Runde: 1}
		game.Players = append(game.Players, mxNewPlayer(input.ID, player.Name, 0))
		game.Note = player.Name + " eröffnet eine Runde Lügen Mäxchen."
		room.Maexchen = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Runde Lügen Mäxchen – 3 € Startgeld, wer zuerst pleite ist, fliegt.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "join":
		game := room.Maexchen
		if game == nil || game.Status != "invited" {
			fehler(http.StatusConflict, "Es läuft gerade keine offene Mäxchen-Runde.")
			return
		}
		if mxFind(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= mxMaxPlayers {
			fehler(http.StatusConflict, "Diese Runde ist mit 6 Spielern voll.")
			return
		}
		game.Players = append(game.Players, mxNewPlayer(input.ID, player.Name, len(game.Players)))
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zu Lügen Mäxchen.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "ready":
		game := room.Maexchen
		if game == nil || game.Status != "invited" {
			fehler(http.StatusConflict, "Bereit drücken geht nur, solange die Runde noch nicht läuft.")
			return
		}
		seat := mxFind(game, input.ID)
		if seat < 0 {
			fehler(http.StatusForbidden, "Du sitzt nicht an diesem Tisch.")
			return
		}
		game.Players[seat].Ready = !game.Players[seat].Ready
		if game.Players[seat].Ready {
			game.Note = player.Name + " ist bereit."
		} else {
			game.Note = player.Name + " wartet noch."
		}

	case "start":
		game := room.Maexchen
		if game == nil || game.Status != "invited" {
			fehler(http.StatusConflict, "Es wartet keine Mäxchen-Runde auf den Start.")
			return
		}
		if game.HostID != input.ID {
			fehler(http.StatusForbidden, "Nur der Gastgeber kann die Runde starten.")
			return
		}
		if len(game.Players) < 2 {
			fehler(http.StatusConflict, "Lügen Mäxchen braucht mindestens zwei Spieler am Tisch.")
			return
		}
		for _, p := range game.Players {
			if !p.Ready {
				fehler(http.StatusConflict, "Erst müssen alle Spieler BEREIT drücken.")
				return
			}
		}
		game.Status = "active"
		game.Runde = 0 // mxNeueRunde zählt hoch: die erste Runde ist die 1.
		mxNeueRunde(game, 0)
		game.Note = "Los geht's: " + game.Players[0].Name + " würfelt zuerst."

	case "roll":
		game := room.Maexchen
		if game == nil || game.Status != "active" {
			fehler(http.StatusConflict, "Es läuft gerade keine aktive Partie.")
			return
		}
		seat := mxFind(game, input.ID)
		if seat < 0 {
			fehler(http.StatusForbidden, "Du gehörst nicht zu dieser Partie.")
			return
		}
		if game.Current != seat {
			fehler(http.StatusConflict, "Der andere Spieler ist an der Reihe.")
			return
		}
		if game.Phase != "wurf" {
			fehler(http.StatusConflict, "Du hast schon gewürfelt – jetzt Ansage weitergeben.")
			return
		}
		if game.Ansage != nil && mxRang(*game.Ansage) >= mxRang(MXCombo{Hoch: 2, Tief: 1}) {
			fehler(http.StatusConflict, "Die Ansage 2-1 ist nicht zu überbieten – da hilft nur Aufdecken.")
			return
		}
		mxWuerfle(game)
		game.Phase = "ansage"
		game.DiceOwner = input.ID
		game.Note = player.Name + " würfelt und überlegt sich eine Ansage."
		game.Note += " Bisher angesagt: " + mxAnsageText(game) + "."

	case "ansage":
		game := room.Maexchen
		if game == nil || game.Status != "active" {
			fehler(http.StatusConflict, "Es läuft gerade keine aktive Partie.")
			return
		}
		seat := mxFind(game, input.ID)
		if seat < 0 {
			fehler(http.StatusForbidden, "Du gehörst nicht zu dieser Partie.")
			return
		}
		if game.Current != seat || game.Phase != "ansage" || game.DiceOwner != input.ID {
			fehler(http.StatusConflict, "Gerade ist keine Ansage von dir offen.")
			return
		}
		if input.Ansage == nil {
			fehler(http.StatusBadRequest, "Es fehlt die Ansage.")
			return
		}
		neu := mxCombo(input.Ansage.Hoch, input.Ansage.Tief)
		if !mxGueltig(neu) {
			fehler(http.StatusBadRequest, "Die Ansage muss aus zwei Würfeln von 1 bis 6 bestehen.")
			return
		}
		if game.Ansage != nil && mxRang(neu) <= mxRang(*game.Ansage) {
			fehler(http.StatusConflict, "Die Ansage muss höher sein als "+game.Ansage.String()+".")
			return
		}
		game.Ansage = &neu
		game.AnsageVon = input.ID
		game.AnsageVonName = player.Name
		game.Dice = nil
		game.DiceOwner = ""
		game.Phase = "wurf"
		game.Current = mxNaechsterAktiv(game, seat)
		game.Note = player.Name + " sagt " + neu.String() + " an. Jetzt ist " +
			game.Players[game.Current].Name + " dran – glauben oder aufdecken."

	case "aufdecken":
		game := room.Maexchen
		if game == nil || game.Status != "active" {
			fehler(http.StatusConflict, "Es läuft gerade keine aktive Partie.")
			return
		}
		seat := mxFind(game, input.ID)
		if seat < 0 {
			fehler(http.StatusForbidden, "Du gehörst nicht zu dieser Partie.")
			return
		}
		if game.Current != seat {
			fehler(http.StatusConflict, "Aufdecken kann nur, wer gerade dran ist.")
			return
		}
		if game.Phase != "wurf" {
			fehler(http.StatusConflict, "Du hast schon gewürfelt – jetzt Ansage weitergeben.")
			return
		}
		if game.Ansage == nil || game.AnsageVon == "" {
			fehler(http.StatusConflict, "Es steht keine Ansage zum Aufdecken im Raum.")
			return
		}
		if game.AnsageVon == input.ID {
			fehler(http.StatusConflict, "Die eigene Ansage kann man nicht aufdecken.")
			return
		}
		if len(game.wurf) != 2 {
			fehler(http.StatusConflict, "Zu dieser Ansage liegt kein Wurf vor.")
			return
		}
		echt := mxCombo(game.wurf[0], game.wurf[1])
		gelogen := echt != *game.Ansage
		verlierer := game.AnsageVon
		if !gelogen {
			verlierer = input.ID
		}
		vIdx := mxFind(game, verlierer)
		if vIdx < 0 {
			fehler(http.StatusConflict, "Der Spieler ist nicht mehr am Tisch.")
			return
		}
		game.Players[vIdx].Euro--
		if game.Players[vIdx].Euro <= 0 {
			game.Players[vIdx].Euro = 0
			game.Players[vIdx].Out = true
		}
		game.revealNr++
		game.Reveal = &MXReveal{
			Nr:          game.revealNr,
			VonID:       game.AnsageVon,
			VonName:     game.AnsageVonName,
			Angesagt:    *game.Ansage,
			Gewuerfelt:  echt,
			Gelogen:     gelogen,
			Aufdecker:   player.Name,
			AufdeckerID: input.ID,
			Verlierer:   game.Players[vIdx].Name,
			VerliererID: verlierer,
			Aus:         game.Players[vIdx].Out,
			Euro:        game.Players[vIdx].Euro,
		}
		if gelogen {
			game.LastLiar = game.AnsageVonName
			game.Note = game.AnsageVonName + " hat gelogen: " + game.Ansage.String() +
				" angesagt, " + echt.String() + " gewürfelt. " + game.Players[vIdx].Name + " zahlt 1 €."
		} else {
			game.Note = game.AnsageVonName + " war ehrlich: " + echt.String() +
				" gewürfelt. " + game.Players[vIdx].Name + " zahlt 1 € fürs falsche Aufdecken."
		}
		if game.Players[vIdx].Out {
			game.Note += " Damit ist " + game.Players[vIdx].Name + " pleite und raus."
		}
		if mxAktiv(game) <= 1 {
			for _, p := range game.Players {
				if !p.Out {
					game.Winner = p.Name
				}
			}
			game.Status = "finished"
			game.Phase = "ende"
			game.Dice = nil
			game.DiceOwner = ""
			game.wurf = nil
			game.Note = game.Winner + " gewinnt die Partie!"
			break
		}
		start := vIdx
		if game.Players[start].Out {
			start = mxNaechsterAktiv(game, start)
		}
		mxNeueRunde(game, start)
		game.Reveal.Start = game.Players[game.Current].Name
		game.Note += " Neue Runde: " + game.Players[game.Current].Name + " würfelt."

	case "reset":
		game := room.Maexchen
		if game == nil || game.Status != "finished" {
			fehler(http.StatusConflict, "Eine neue Runde ist erst nach Spielende möglich.")
			return
		}
		if mxFind(game, input.ID) < 0 {
			fehler(http.StatusForbidden, "Du gehörst nicht zu dieser Partie.")
			return
		}
		fresh := &MaexchenGame{ID: game.ID, HostID: game.HostID, Status: "invited", Phase: "wurf", Runde: 1}
		for i, p := range game.Players {
			fresh.Players = append(fresh.Players, mxNewPlayer(p.ID, p.Name, i))
		}
		fresh.Note = "Neue Runde – alle drücken BEREIT, dann startet der Gastgeber."
		room.Maexchen = fresh

	default:
		fehler(http.StatusBadRequest, "Unbekannte Mäxchen-Aktion.")
		return
	}

	room.LastActive = time.Now()
	result := mxView(room.Maexchen, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

func mxAnsageText(g *MaexchenGame) string {
	if g == nil || g.Ansage == nil {
		return "noch keine"
	}
	return g.Ansage.String() + " von " + g.AnsageVonName
}
