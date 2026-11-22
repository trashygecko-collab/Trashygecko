package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Galgenmännchen: Ein gemeinsames Wort, die Spieler raten streng abwechselnd
// je einen Buchstaben. Ein Treffer bringt Punkte, ein Fehlgriff kostet ein
// Leben (für alle). Wer mit seinem Buchstaben das letzte offene Feld füllt,
// hat das Wort erraten und gewinnt; sind vorher die Leben weg, haben alle
// verloren (kein Sieger).
const galgenMaxSpieler = 6

// galgenWorte: Wort, Kategorie und Stufe (1 leicht, 2 mittel, 3 schwer).
type galgenWort struct {
	Wort      string
	Kategorie string
	Stufe     int
}

var galgenListe = []galgenWort{
	{"sonne", "Natur", 1}, {"wolke", "Natur", 1}, {"stein", "Natur", 1},
	{"blume", "Natur", 1}, {"fluss", "Natur", 1}, {"wiese", "Natur", 1},
	{"apfel", "Essen", 1}, {"brot", "Essen", 1}, {"kaese", "Essen", 1},
	{"milch", "Essen", 1}, {"salat", "Essen", 1}, {"nudel", "Essen", 1},
	{"tisch", "Haus", 1}, {"stuhl", "Haus", 1}, {"fenster", "Haus", 2},
	{"garten", "Haus", 1}, {"teppich", "Haus", 2}, {"spiegel", "Haus", 2},
	{"fahrrad", "Technik", 2}, {"motor", "Technik", 1}, {"schraube", "Technik", 2},
	{"kabel", "Technik", 1}, {"lautsprecher", "Technik", 3}, {"batterie", "Technik", 2},
	{"gitarre", "Musik", 2}, {"trommel", "Musik", 2}, {"klavier", "Musik", 2},
	{"geige", "Musik", 1}, {"floete", "Musik", 2}, {"konzert", "Musik", 2},
	{"planet", "Weltraum", 2}, {"rakete", "Weltraum", 2}, {"stern", "Weltraum", 1},
	{"galaxie", "Weltraum", 3}, {"mond", "Weltraum", 1}, {"teleskop", "Weltraum", 3},
	{"schwimmen", "Sport", 2}, {"laufen", "Sport", 1}, {"springen", "Sport", 2},
	{"ball", "Sport", 1}, {"torwart", "Sport", 2}, {"stadion", "Sport", 2},
	{"lehrer", "Beruf", 1}, {"arzt", "Beruf", 1}, {"baecker", "Beruf", 2},
	{"maler", "Beruf", 1}, {"pilot", "Beruf", 1}, {"gaertner", "Beruf", 2},
	{"himmel", "Wetter", 2}, {"regen", "Wetter", 1}, {"schnee", "Wetter", 1},
	{"gewitter", "Wetter", 3}, {"nebel", "Wetter", 1}, {"sonnenschein", "Wetter", 3},
	{"insel", "Reisen", 1}, {"berg", "Reisen", 1}, {"strand", "Reisen", 2},
	{"hotel", "Reisen", 1}, {"koffer", "Reisen", 2}, {"urlaub", "Reisen", 2},
	{"bibliothek", "Orte", 3}, {"schule", "Orte", 1}, {"markt", "Orte", 1},
	{"brunnen", "Orte", 2}, {"bahnhof", "Orte", 2}, {"museum", "Orte", 2},
	{"elefant", "Tiere", 2}, {"katze", "Tiere", 1}, {"hund", "Tiere", 1},
	{"pinguin", "Tiere", 2}, {"schmetterling", "Tiere", 3}, {"eichhoernchen", "Tiere", 3},
	{"fuchs", "Tiere", 1}, {"igel", "Tiere", 1}, {"wal", "Tiere", 1},
	{"zitrone", "Obst", 2}, {"banane", "Obst", 2}, {"kirsche", "Obst", 2},
	{"erdbeere", "Obst", 2}, {"pflaume", "Obst", 2}, {"ananas", "Obst", 2},
	{"schokolade", "Suess", 3}, {"kuchen", "Suess", 1}, {"keks", "Suess", 1},
	{"honig", "Suess", 1}, {"eis", "Suess", 1}, {"marmelade", "Suess", 3},
	{"fernseher", "Wohnen", 3}, {"sofaecke", "Wohnen", 3}, {"lampe", "Wohnen", 1},
	{"vorhang", "Wohnen", 2}, {"kissen", "Wohnen", 2}, {"schrank", "Wohnen", 2},
	{"zeitung", "Alltag", 2}, {"brief", "Alltag", 1}, {"schluessel", "Alltag", 3},
	{"brille", "Alltag", 2}, {"tasche", "Alltag", 2}, {"uhr", "Alltag", 1},
}

// GalgenVarianten bestimmen Fehlerzahl und Hinweise.
var galgenVarianten = map[string]struct {
	Name     string
	Fehler   int
	Hinweis  bool
	MaxStufe int
}{
	"klassisch": {"Klassisch", 8, true, 3},
	"streng":    {"Streng", 6, true, 3},
	"ohnehilfe": {"Ohne Hinweis", 8, false, 3},
	"leicht":    {"Leicht", 10, true, 1},
}

type GalgenSpieler struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Punkte  int    `json:"punkte"`
	Treffer int    `json:"treffer"`
	Ready   bool   `json:"ready"`
}

type GalgenGame struct {
	ID            string          `json:"id"`
	HostID        string          `json:"hostId"`
	Status        string          `json:"status"` // invited, spiel, ende
	Variante      string          `json:"variante"`
	VarianteName  string          `json:"varianteName"`
	Kategorie     string          `json:"kategorie,omitempty"`
	HinweisZeigen bool            `json:"hinweisZeigen"`
	Wort          string          `json:"-"`
	Anzeige       []string        `json:"anzeige"` // ein Eintrag je Buchstabe
	Geloest       bool            `json:"geloest"`
	Buchstaben    []string        `json:"buchstaben"` // geratene Buchstaben
	Fehler        int             `json:"fehler"`
	MaxFehler     int             `json:"maxFehler"`
	Spieler       []GalgenSpieler `json:"spieler"`
	Zug           int             `json:"zug"`
	Runde         int             `json:"runde"`
	Sieger        string          `json:"sieger,omitempty"`
	Note          string          `json:"note,omitempty"`
	Letzte        []string        `json:"letzte"` // Protokoll der letzten Züge
	BotID         string          `json:"botId,omitempty"`
	BotStufe      string          `json:"-"`
	BotFrist      time.Time       `json:"-"`
}

func galgenNeuesSpiel(hostID, hostName, variante, id string) *GalgenGame {
	v, ok := galgenVarianten[variante]
	if !ok {
		variante = "klassisch"
		v = galgenVarianten[variante]
	}
	return &GalgenGame{
		ID: id, HostID: hostID, Status: "invited", Variante: variante, VarianteName: v.Name,
		MaxFehler: v.Fehler, HinweisZeigen: v.Hinweis,
		// Auch der Gastgeber erklärt sich bereit – so ist der Bereit-Knopf für
		// alle gleich und niemand startet eine Runde im Alleingang.
		Spieler: []GalgenSpieler{{ID: hostID, Name: hostName, Ready: false}},
	}
}

// galgenWaehlen zieht ein Wort passend zur Variante und setzt die Anzeige.
func (g *GalgenGame) galgenWaehlen() {
	stufe := galgenVarianten[g.Variante].MaxStufe
	passend := make([]galgenWort, 0, len(galgenListe))
	for _, w := range galgenListe {
		if w.Stufe <= stufe {
			passend = append(passend, w)
		}
	}
	w := passend[rand.Intn(len(passend))]
	g.Wort = w.Wort
	g.Kategorie = w.Kategorie
	g.Anzeige = make([]string, len([]rune(w.Wort)))
	g.Buchstaben = []string{}
	g.Fehler = 0
	g.Geloest = false
	g.Zug = rand.Intn(len(g.Spieler))
	g.Letzte = []string{}
	for i := range g.Spieler {
		g.Spieler[i].Punkte = 0
		g.Spieler[i].Treffer = 0
	}
	g.Note = "Das Wort hat " + strconv.Itoa(len([]rune(w.Wort))) + " Buchstaben."
}

func (g *GalgenGame) galgenSitz(id string) int {
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			return i
		}
	}
	return -1
}

// galgenRaten prüft einen Buchstaben und liefert die Trefferzahl; ein
// Problemtext sagt, warum der Zug nicht gilt.
func (g *GalgenGame) galgenRaten(buchstabe string) (int, string) {
	buchstabe = strings.ToLower(strings.TrimSpace(buchstabe))
	if len([]rune(buchstabe)) != 1 {
		return 0, "Bitte genau einen Buchstaben raten."
	}
	for _, b := range g.Buchstaben {
		if b == buchstabe {
			return 0, "Dieser Buchstabe wurde schon geraten."
		}
	}
	g.Buchstaben = append(g.Buchstaben, buchstabe)
	treffer := 0
	for i, r := range []rune(g.Wort) {
		if strings.ToLower(string(r)) == buchstabe {
			g.Anzeige[i] = string(r)
			treffer++
		}
	}
	return treffer, ""
}

func (g *GalgenGame) galgenFertig() bool {
	for _, a := range g.Anzeige {
		if a == "" {
			return false
		}
	}
	return true
}

func (g *GalgenGame) galgenWeiter() {
	g.Zug = (g.Zug + 1) % len(g.Spieler)
}

// galgenBeenden schließt die Runde. Sieger ist, wer das Wort erraten hat;
// eine verlorene Runde hat keinen Sieger. Zurück kommt der Sitz mit den
// meisten Punkten (nur für den Text).
func (g *GalgenGame) galgenBeenden(geloest bool, siegerID string) int {
	g.Status = "ende"
	best := 0
	for i := range g.Spieler {
		if g.Spieler[i].Punkte > g.Spieler[best].Punkte {
			best = i
		}
	}
	if geloest {
		g.Sieger = siegerID
	} else {
		g.Sieger = ""
	}
	return best
}

func cloneGalgen(g *GalgenGame) *GalgenGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Anzeige = append([]string(nil), g.Anzeige...)
	c.Buchstaben = append([]string(nil), g.Buchstaben...)
	c.Letzte = append([]string(nil), g.Letzte...)
	c.Spieler = append([]GalgenSpieler(nil), g.Spieler...)
	return &c
}

// galgenAktion prüft einen Ratezug. Mensch (HTTP) und Bot (galgenBotSchritt)
// gehen durch dieselbe Prüfung: Reihe, ein Buchstabe, keine Doppelung, Treffer,
// Punkte, Rundenende.
func (g *GalgenGame) galgenAktion(sitz int, buchstabe string) (int, string) {
	if g == nil || g.Status != "spiel" {
		return http.StatusConflict, "Gerade läuft keine Runde."
	}
	if sitz < 0 || sitz >= len(g.Spieler) {
		return http.StatusForbidden, "Du spielst nicht mit."
	}
	if sitz != g.Zug {
		return http.StatusConflict, "Du bist nicht an der Reihe."
	}
	anzahl, problem := g.galgenRaten(buchstabe)
	if problem != "" {
		return http.StatusConflict, problem
	}
	b := strings.ToLower(strings.TrimSpace(buchstabe))
	name := g.Spieler[sitz].Name
	if anzahl > 0 {
		g.Spieler[sitz].Punkte += anzahl
		g.Spieler[sitz].Treffer++
		g.Note = name + " trifft „" + strings.ToUpper(b) + "“ – " + strconv.Itoa(anzahl) + " Buchstabe(n)."
		g.Letzte = append(g.Letzte, "✔ "+name+": "+strings.ToUpper(b))
		if g.galgenFertig() {
			g.Geloest = true
			g.galgenBeenden(true, g.Spieler[sitz].ID)
			g.Note = "Wort gelöst: " + strings.ToUpper(g.Wort) + ". " +
				name + " hat das letzte Feld gefüllt und gewinnt."
		} else {
			g.galgenWeiter()
		}
	} else {
		g.Fehler++
		g.Note = name + " rät „" + strings.ToUpper(b) + "“ – daneben (" +
			strconv.Itoa(g.Fehler) + "/" + strconv.Itoa(g.MaxFehler) + ")."
		g.Letzte = append(g.Letzte, "✖ "+name+": "+strings.ToUpper(b))
		if g.Fehler >= g.MaxFehler {
			g.galgenBeenden(false, "")
			g.Note = "Verloren! Das Wort war „" + strings.ToUpper(g.Wort) + "“ – diesmal hat es keiner erraten."
		} else {
			g.galgenWeiter()
		}
	}
	if len(g.Letzte) > 8 {
		g.Letzte = g.Letzte[len(g.Letzte)-8:]
	}
	return 0, ""
}

// galgenNaechste startet die nächste Runde.
func (g *GalgenGame) galgenNaechste(sitz int) (int, string) {
	if g == nil || g.Status != "ende" {
		return http.StatusConflict, "Es läuft keine beendete Runde."
	}
	if sitz < 0 || sitz >= len(g.Spieler) {
		return http.StatusForbidden, "Du spielst nicht mit."
	}
	g.Status = "spiel"
	g.Runde++
	g.galgenWaehlen()
	return 0, ""
}

func galgenHandler(w http.ResponseWriter, r *http.Request) {
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
		galgenBotTreiben(room.Galgen, time.Now())
		game := cloneGalgen(room.Galgen)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID        string `json:"id"`
		Variante  string `json:"variante"`
		Buchstabe string `json:"buchstabe"`
		Stufe     string `json:"stufe"`
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

	switch action {
	case "bot":
		if room.Galgen != nil && !room.gameSlotFree(room.Galgen.Status, galgenIDs(room.Galgen)...) {
			fail("In diesem Spielraum läuft bereits eine Galgenmännchen-Runde.", http.StatusConflict)
			return
		}
		g := galgenNeuesSpiel(input.ID, player.Name, input.Variante, randomCode())
		g.Spieler = append(g.Spieler, GalgenSpieler{ID: botID(1), Name: botName(1), Ready: true})
		g.Spieler[0].Ready = true
		g.BotID, g.BotStufe = botID(1), botStufe(input.Stufe)
		g.Status = "spiel"
		g.Runde = 1
		g.galgenWaehlen()
		g.Note = player.Name + " spielt gegen " + botName(1) + " (" + botStufeText(g.BotStufe) + "). " + g.Note
		room.Galgen = g
	case "create":
		if room.Galgen != nil && !room.gameSlotFree(room.Galgen.Status, galgenIDs(room.Galgen)...) {
			fail("In diesem Spielraum läuft bereits eine Galgenmännchen-Runde.", http.StatusConflict)
			return
		}
		room.Galgen = galgenNeuesSpiel(input.ID, player.Name, input.Variante, randomCode())
		room.Galgen.Note = player.Name + " eröffnet eine Runde Galgenmännchen (" + room.Galgen.VarianteName + ")."
	case "join":
		g := room.Galgen
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Runde.", http.StatusConflict)
			return
		}
		if g.galgenSitz(input.ID) >= 0 {
			fail("Du sitzt schon am Tisch.", http.StatusConflict)
			return
		}
		if len(g.Spieler) >= galgenMaxSpieler {
			fail("Der Tisch ist voll (höchstens "+strconv.Itoa(galgenMaxSpieler)+" Spieler).", http.StatusConflict)
			return
		}
		g.Spieler = append(g.Spieler, GalgenSpieler{ID: input.ID, Name: player.Name, Ready: false})
	case "ready":
		g := room.Galgen
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Runde.", http.StatusConflict)
			return
		}
		sitz := g.galgenSitz(input.ID)
		if sitz < 0 {
			fail("Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		g.Spieler[sitz].Ready = !g.Spieler[sitz].Ready
	case "start":
		g := room.Galgen
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Runde.", http.StatusConflict)
			return
		}
		if g.galgenSitz(input.ID) < 0 {
			fail("Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		for _, s := range g.Spieler {
			if !s.Ready {
				fail("Noch nicht alle sind bereit.", http.StatusConflict)
				return
			}
		}
		g.Status = "spiel"
		g.Runde = 1
		g.galgenWaehlen()
	case "raten":
		g := room.Galgen
		if g == nil {
			fail("Gerade läuft keine Runde.", http.StatusConflict)
			return
		}
		sitz := g.galgenSitz(input.ID)
		if sitz < 0 {
			fail("Du spielst nicht mit.", http.StatusForbidden)
			return
		}
		if status, meldung := g.galgenAktion(sitz, input.Buchstabe); status != 0 {
			fail(meldung, status)
			return
		}
	case "naechste":
		g := room.Galgen
		if g == nil {
			fail("Es läuft keine beendete Runde.", http.StatusConflict)
			return
		}
		sitz := g.galgenSitz(input.ID)
		if sitz < 0 {
			fail("Du spielst nicht mit.", http.StatusForbidden)
			return
		}
		if status, meldung := g.galgenNaechste(sitz); status != 0 {
			fail(meldung, status)
			return
		}
	case "reset":
		if room.Galgen != nil && room.Galgen.HostID != input.ID {
			fail("Nur der Gastgeber kann die Runde schließen.", http.StatusForbidden)
			return
		}
		room.Galgen = nil
	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}
	room.LastActive = time.Now()
	game := cloneGalgen(room.Galgen)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": game})
}

func galgenIDs(g *GalgenGame) []string {
	ids := make([]string, 0, len(g.Spieler))
	for _, s := range g.Spieler {
		ids = append(ids, s.ID)
	}
	return ids
}

// galgenRemoveSeat nimmt einen Spieler aus der Runde (Verlassen/Abbruch).
func galgenRemoveSeat(g *GalgenGame, id, name string) (bool, bool, int) {
	sitz := g.galgenSitz(id)
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
