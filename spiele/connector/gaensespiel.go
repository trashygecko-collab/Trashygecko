package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Gänsespiel: 63 Felder auf einem Rundkurs, zwei Würfel je Zug. Wer 63 genau
// erreicht, gewinnt; wer darüber würfelt, läuft die überzähligen Schritte
// zurück. Die Sonderfelder sind fest und gelten für alle:
//
//	6 Brücke → 12, 9/18/27/36/45/54 Gans → noch einmal die Augenzahl,
//	19 Gasthaus → eine Runde aussetzen, 31 Brunnen → warten, bis jemand
//	anders auf 31 landet, 42 Labyrinth → zurück auf 30, 52 Gefängnis →
//	drei Runden aussetzen, 58 Tod → zurück auf 1.
const (
	gaenseMaxSpieler       = 4
	gaenseFelder           = 63
	gaenseBruecke          = 6
	gaenseBrueckeZiel      = 12
	gaenseGasthaus         = 19
	gaenseBrunnen          = 31
	gaenseLabyrinth        = 42
	gaenseLabyrinthZiel    = 30
	gaenseGefaengnis       = 52
	gaenseGefaengnisRunden = 3
	gaenseTod              = 58
	gaenseTodZiel          = 1
)

type GaenseSpieler struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Ready     bool   `json:"ready"`
	Farbe     int    `json:"farbe"`
	Pos       int    `json:"pos"` // 0 = noch am Start
	Aussetzen int    `json:"aussetzen"`
	ImBrunnen bool   `json:"imBrunnen"`
}

type GaenseGame struct {
	ID      string          `json:"id"`
	HostID  string          `json:"hostId"`
	Status  string          `json:"status"` // invited, spiel, ende
	Runde   int             `json:"runde"`
	Zug     int             `json:"zug"`
	Spieler []GaenseSpieler `json:"spieler"`
	Wuerfel [2]int          `json:"wuerfel"`
	Letzte  string          `json:"letzte,omitempty"`
	Verlauf []string        `json:"verlauf"`
	Sieger  string          `json:"sieger,omitempty"`
	Note    string          `json:"note,omitempty"`
}

// gaenseGans meldet, ob ein Feld ein Gänsefeld ist.
func gaenseGans(feld int) bool {
	return feld > 0 && feld%9 == 0 && feld <= 54
}

func gaenseNeuesSpiel(hostID, hostName, id string) *GaenseGame {
	return &GaenseGame{
		ID: id, HostID: hostID, Status: "invited", Verlauf: []string{},
		Spieler: []GaenseSpieler{{ID: hostID, Name: hostName}},
	}
}

func (g *GaenseGame) sitz(id string) int {
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			return i
		}
	}
	return -1
}

func (g *GaenseGame) gaenseRunde() {
	g.Wuerfel = [2]int{}
	g.Letzte = ""
	g.Verlauf = []string{}
	g.Sieger = ""
	g.Zug = 0
	for i := range g.Spieler {
		g.Spieler[i].Pos = 0
		g.Spieler[i].Aussetzen = 0
		g.Spieler[i].ImBrunnen = false
	}
	g.Status = "spiel"
	g.Note = "63 genau erreichen – wer darüber würfelt, läuft zurück. Gänsefelder lassen euch noch einmal ziehen."
}

func (g *GaenseGame) verlauf(text string) {
	g.Letzte = text
	g.Verlauf = append(g.Verlauf, text)
	if len(g.Verlauf) > 10 {
		g.Verlauf = g.Verlauf[len(g.Verlauf)-10:]
	}
}

// gaenseZiehen rechnet einen Zug mit Rücklauf über das Ziel hinaus.
func gaenseZiehen(pos, schritte int) (int, string) {
	ziel := pos + schritte
	if ziel <= gaenseFelder {
		return ziel, ""
	}
	zurueck := ziel - gaenseFelder
	return gaenseFelder - zurueck, " Über das Ziel hinaus – " + strconv.Itoa(zurueck) + " Felder zurück."
}

// gaenseNaechster gibt den Zug weiter und überspringt, wer aussetzen muss oder
// im Brunnen sitzt.
func (g *GaenseGame) gaenseNaechster() {
	if len(g.Spieler) == 0 {
		return
	}
	for i := 0; i < len(g.Spieler)*8; i++ {
		g.Zug = (g.Zug + 1) % len(g.Spieler)
		sp := &g.Spieler[g.Zug]
		if sp.ImBrunnen {
			g.verlauf(sp.Name + " sitzt im Brunnen und setzt aus.")
			continue
		}
		if sp.Aussetzen > 0 {
			sp.Aussetzen--
			if sp.Aussetzen > 0 {
				g.verlauf(sp.Name + " setzt aus – noch " + strconv.Itoa(sp.Aussetzen) + " Runden.")
			} else {
				g.verlauf(sp.Name + " setzt diese Runde aus.")
			}
			continue
		}
		return
	}
}

// gaenseWuerfeln zieht und löst alle Sonderfelder der Reihe nach aus.
func (g *GaenseGame) gaenseWuerfeln(sitz int) string {
	a, b := 1+rand.Intn(6), 1+rand.Intn(6)
	g.Wuerfel = [2]int{a, b}
	return g.gaenseZug(sitz, a+b)
}

// gaenseZug rechnet einen Zug mit vorgegebener Augensumme.
func (g *GaenseGame) gaenseZug(sitz, summe int) string {
	if g.Status != "spiel" {
		return "Gerade läuft keine Runde."
	}
	if sitz != g.Zug {
		return "Du bist nicht an der Reihe."
	}
	sp := &g.Spieler[sitz]
	a, b := g.Wuerfel[0], g.Wuerfel[1]
	if a == 0 && b == 0 {
		a, b = 0, summe
	}
	teile := []string{sp.Name + " würfelt " + strconv.Itoa(a) + " + " + strconv.Itoa(b) + " = " + strconv.Itoa(summe) + "."}
	pos, zusatz := gaenseZiehen(sp.Pos, summe)
	teile = append(teile, "Von Feld "+strconv.Itoa(sp.Pos)+" auf "+strconv.Itoa(pos)+"."+zusatz)
	sp.Pos = pos

	for schritt := 0; schritt < 8; schritt++ {
		switch feld := sp.Pos; {
		case feld == gaenseBruecke:
			sp.Pos = gaenseBrueckeZiel
			teile = append(teile, "Die Brücke trägt "+sp.Name+" auf Feld "+strconv.Itoa(gaenseBrueckeZiel)+".")
		case gaenseGans(feld):
			neu, mehr := gaenseZiehen(feld, summe)
			sp.Pos = neu
			teile = append(teile, "Gänsefeld "+strconv.Itoa(feld)+" – "+sp.Name+" zieht noch einmal "+strconv.Itoa(summe)+" Felder auf "+strconv.Itoa(neu)+"."+mehr)
		case feld == gaenseGasthaus:
			sp.Aussetzen = 1
			teile = append(teile, "Im Gasthaus muss "+sp.Name+" eine Runde aussetzen.")
			schritt = 8
		case feld == gaenseBrunnen:
			for i := range g.Spieler {
				if i != sitz && g.Spieler[i].ImBrunnen {
					g.Spieler[i].ImBrunnen = false
					teile = append(teile, g.Spieler[i].Name+" ist aus dem Brunnen frei.")
				}
			}
			sp.ImBrunnen = true
			teile = append(teile, sp.Name+" fällt in den Brunnen und wartet, bis jemand auf Feld 31 landet.")
			schritt = 8
		case feld == gaenseLabyrinth:
			sp.Pos = gaenseLabyrinthZiel
			teile = append(teile, "Das Labyrinth schickt "+sp.Name+" zurück auf Feld "+strconv.Itoa(gaenseLabyrinthZiel)+".")
		case feld == gaenseGefaengnis:
			sp.Aussetzen = gaenseGefaengnisRunden
			teile = append(teile, "Im Gefängnis setzt "+sp.Name+" "+strconv.Itoa(gaenseGefaengnisRunden)+" Runden aus.")
			schritt = 8
		case feld == gaenseTod:
			sp.Pos = gaenseTodZiel
			teile = append(teile, "Der Tod schickt "+sp.Name+" zurück zum Anfang.")
		default:
			schritt = 8
		}
	}
	g.verlauf(strings.Join(teile, " "))
	if sp.Pos >= gaenseFelder {
		g.Status = "ende"
		g.Sieger = sp.ID
		g.Note = sp.Name + " erreicht Feld " + strconv.Itoa(gaenseFelder) + " und gewinnt."
		return ""
	}
	g.gaenseNaechster()
	return ""
}

func gaenseAnzeige(g *GaenseGame) *GaenseGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Spieler = append([]GaenseSpieler{}, g.Spieler...)
	c.Verlauf = append([]string{}, g.Verlauf...)
	return &c
}

func gaenseHandler(w http.ResponseWriter, r *http.Request) {
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
		game := gaenseAnzeige(room.Gaense)
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
	amTisch := func(status, fehlt, nichtDabei string) (*GaenseGame, int, string, int) {
		g := room.Gaense
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
		if room.Gaense != nil {
			slot := room.Gaense.Status
			if slot == "spiel" {
				slot = "active"
			}
			if !room.gameSlotFree(slot, gaenseIDs(room.Gaense)...) {
				fail("In diesem Spielraum läuft bereits eine Runde Gänsespiel.", http.StatusConflict)
				return
			}
		}
		room.Gaense = gaenseNeuesSpiel(input.ID, player.Name, randomCode())
		room.Gaense.Note = player.Name + " eröffnet eine Runde Gänsespiel."
	case "join":
		g := room.Gaense
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Runde.", http.StatusConflict)
			return
		}
		if g.sitz(input.ID) >= 0 {
			fail("Du sitzt schon am Tisch.", http.StatusConflict)
			return
		}
		if len(g.Spieler) >= gaenseMaxSpieler {
			fail("Der Tisch ist voll (höchstens "+strconv.Itoa(gaenseMaxSpieler)+" Spieler).", http.StatusConflict)
			return
		}
		g.Spieler = append(g.Spieler, GaenseSpieler{ID: input.ID, Name: player.Name, Farbe: len(g.Spieler)})
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
			fail("Das Gänsespiel braucht mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, s := range g.Spieler {
			if !s.Ready {
				fail("Noch nicht alle sind bereit.", http.StatusConflict)
				return
			}
		}
		g.Runde = 1
		g.gaenseRunde()
	case "wuerfeln":
		g, sitz, problem, status := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, status)
			return
		}
		if text := g.gaenseWuerfeln(sitz); text != "" {
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
		g.gaenseRunde()
	case "reset":
		if room.Gaense != nil && room.Gaense.HostID != input.ID {
			fail("Nur der Gastgeber kann die Runde schließen.", http.StatusForbidden)
			return
		}
		room.Gaense = nil
	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}
	room.LastActive = time.Now()
	game := gaenseAnzeige(room.Gaense)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": game})
}

func gaenseIDs(g *GaenseGame) []string {
	ids := make([]string, 0, len(g.Spieler))
	for _, s := range g.Spieler {
		ids = append(ids, s.ID)
	}
	return ids
}

func gaenseRemoveSeat(g *GaenseGame, id, name string) (bool, bool, int) {
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
