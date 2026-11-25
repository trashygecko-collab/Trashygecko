package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Wortsuche: Alle sehen dasselbe Buchstabengitter und dieselbe Wortliste.
// Wer ein Wort zuerst findet, bekommt Punkte dafür; danach ist es für alle
// abgehakt. Die Runde endet, wenn alle Wörter gefunden sind oder die Zeit
// abgelaufen ist – wer dann die meisten Punkte hat, gewinnt.
const (
	wortsucheMaxSpieler = 6
	wortsucheKante      = 10
	wortsucheWorte      = 8
	wortsucheDauer      = 180 // Sekunden je Runde
	wortsuchePunktFak   = 10  // Punkte je Buchstabe
)

// Wörter ohne Umlaute, damit das Gitter aus einfachen Buchstaben besteht.
var wortsucheListe = []string{
	"garten", "fenster", "brunnen", "schule", "hafen", "insel", "wolke", "regen",
	"schnee", "sonne", "mond", "stern", "rakete", "planet", "galaxie", "komet",
	"apfel", "banane", "kirsche", "zitrone", "salat", "brot", "kaese", "milch",
	"tisch", "stuhl", "schrank", "teppich", "spiegel", "lampe", "vorhang", "kissen",
	"fahrrad", "motor", "kabel", "batterie", "schraube", "gitarre", "trommel", "klavier",
	"geige", "floete", "konzert", "lehrer", "arzt", "pilot", "maler", "gaertner",
	"elefant", "katze", "hund", "pinguin", "fuchs", "igel", "wal", "adler",
	"insel", "strand", "hotel", "koffer", "urlaub", "berg", "tal", "fluss",
	"wiese", "blume", "stein", "sand", "meer", "bucht", "wald", "baum",
	"zeitung", "brief", "brille", "tasche", "uhr", "schluessel", "kuchen", "keks",
	"honig", "eimer", "schaufel", "leiter", "haken", "seil", "laterne", "kerze",
	"teller", "messer", "gabel", "loffel", "topf", "pfanne", "ofen", "herd",
}

var wortsucheRichtungen = [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}

type WortsucheSpieler struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Ready       bool   `json:"ready"`
	Punkte      int    `json:"punkte"`
	Treffer     int    `json:"treffer"`
	Fehlversuch int    `json:"fehlversuch"`
}

// WortsucheFund ist ein gefundenes Wort samt seinen Feldern im Gitter.
type WortsucheFund struct {
	Wort    string   `json:"wort"`
	VonID   string   `json:"vonId"`
	VonName string   `json:"vonName"`
	Punkte  int      `json:"punkte"`
	Zellen  [][2]int `json:"zellen"`
}

type WortsucheGame struct {
	ID       string             `json:"id"`
	HostID   string             `json:"hostId"`
	Status   string             `json:"status"` // invited, spiel, ende
	Runde    int                `json:"runde"`
	Kante    int                `json:"kante"`
	Gitter   [][]string         `json:"gitter"`
	Worte    []string           `json:"worte"`
	Funde    []WortsucheFund    `json:"funde"`
	Spieler  []WortsucheSpieler `json:"spieler"`
	Rest     int                `json:"rest"` // Sekunden
	Sieger   string             `json:"sieger,omitempty"`
	Note     string             `json:"note,omitempty"`
	BotID    string             `json:"botId,omitempty"`
	BotStufe string             `json:"botStufe,omitempty"`

	start    time.Time // Beginn der Runde (nur serverseitig)
	BotFrist time.Time // nächster Fund des Computers (nur serverseitig)
}

func wortsucheNeuesSpiel(hostID, hostName, id string) *WortsucheGame {
	return &WortsucheGame{
		ID: id, HostID: hostID, Status: "invited",
		Kante: wortsucheKante, Gitter: [][]string{}, Worte: []string{}, Funde: []WortsucheFund{},
		Rest: wortsucheDauer,
		// Auch der Gastgeber erklärt sich bereit.
		Spieler: []WortsucheSpieler{{ID: hostID, Name: hostName}},
	}
}

// wortsucheGitter legt ein Gitter an und verteilt die Wörter darauf.
func wortsucheGitter(kante, anzahl int) ([][]string, []string) {
	gitter := make([][]string, kante)
	for i := range gitter {
		gitter[i] = make([]string, kante)
	}
	worte := []string{}
	reihenfolge := rand.Perm(len(wortsucheListe))
	for _, i := range reihenfolge {
		if len(worte) >= anzahl {
			break
		}
		wort := strings.ToUpper(wortsucheListe[i])
		if len([]rune(wort)) > kante || enthaelt(worte, wort) {
			continue
		}
		// Ein paar Plätze ausprobieren, ob das Wort passt.
		for versuch := 0; versuch < 60; versuch++ {
			richtung := wortsucheRichtungen[rand.Intn(len(wortsucheRichtungen))]
			zeichen := []rune(wort)
			zeile := rand.Intn(kante)
			spalte := rand.Intn(kante)
			endeZ := zeile + (len(zeichen)-1)*richtung[0]
			endeS := spalte + (len(zeichen)-1)*richtung[1]
			if endeZ < 0 || endeS < 0 || endeZ >= kante || endeS >= kante {
				continue
			}
			// Nur freie Felder belegen (gleiche Buchstaben sind erlaubt).
			belegt := false
			for k, z := range zeichen {
				feld := gitter[zeile+k*richtung[0]][spalte+k*richtung[1]]
				if feld != "" && feld != string(z) {
					belegt = true
					break
				}
			}
			if belegt {
				continue
			}
			for k, z := range zeichen {
				gitter[zeile+k*richtung[0]][spalte+k*richtung[1]] = string(z)
			}
			worte = append(worte, wort)
			break
		}
	}
	// Restliche Felder mit zufälligen Buchstaben füllen.
	alphabet := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for z := 0; z < kante; z++ {
		for s := 0; s < kante; s++ {
			if gitter[z][s] == "" {
				gitter[z][s] = string(alphabet[rand.Intn(len(alphabet))])
			}
		}
	}
	sort.Strings(worte)
	return gitter, worte
}

func (g *WortsucheGame) sitz(id string) int {
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			return i
		}
	}
	return -1
}

// wortsucheRunde startet ein frisches Gitter mit neuen Wörtern.
func (g *WortsucheGame) wortsucheRunde() {
	g.Gitter, g.Worte = wortsucheGitter(g.Kante, wortsucheWorte)
	g.Funde = []WortsucheFund{}
	for i := range g.Spieler {
		g.Spieler[i].Punkte = 0
		g.Spieler[i].Treffer = 0
		g.Spieler[i].Fehlversuch = 0
	}
	g.Status = "spiel"
	g.start = time.Now()
	g.Rest = wortsucheDauer
	g.Sieger = ""
	g.Note = "Findet die " + strconv.Itoa(len(g.Worte)) + " Wörter – waagerecht, senkrecht oder schräg."
}

// wortsuchePruefeZeit hält die Restzeit aktuell und beendet die Runde, wenn
// die Zeit um ist.
func (g *WortsucheGame) wortsuchePruefeZeit() {
	if g.Status != "spiel" {
		return
	}
	verstrichen := int(time.Since(g.start).Seconds())
	g.Rest = wortsucheDauer - verstrichen
	if g.Rest <= 0 {
		g.Rest = 0
		g.wortsucheBeenden("Die Zeit ist um.")
	}
}

func (g *WortsucheGame) wortsucheBeenden(note string) {
	g.Status = "ende"
	g.Sieger = ""
	best := -1
	for i := range g.Spieler {
		if best < 0 || g.Spieler[i].Punkte > g.Spieler[best].Punkte {
			best = i
		}
	}
	if best >= 0 && g.Spieler[best].Punkte > 0 {
		g.Sieger = g.Spieler[best].ID
		note += " " + g.Spieler[best].Name + " gewinnt mit " + strconv.Itoa(g.Spieler[best].Punkte) + " Punkten."
	}
	g.Note = note
}

func (g *WortsucheGame) gefunden(wort string) bool {
	for _, f := range g.Funde {
		if f.Wort == wort {
			return true
		}
	}
	return false
}

// wortsucheFinden prüft eine Auswahl: gerade Linie, passende Länge, Wort aus
// der Liste, Buchstaben im Gitter – und noch nicht gefunden.
func (g *WortsucheGame) wortsucheFinden(sitz int, wort string, von, bis [2]int) (string, int) {
	wort = strings.ToUpper(strings.TrimSpace(wort))
	if wort == "" {
		return "Kein Wort ausgewählt.", 0
	}
	if !enthaelt(g.Worte, wort) {
		return "Dieses Wort steht nicht auf der Liste.", 0
	}
	if g.gefunden(wort) {
		return "Das Wort wurde schon gefunden.", 0
	}
	dz := bis[0] - von[0]
	ds := bis[1] - von[1]
	if dz == 0 && ds == 0 {
		return "Bitte Anfang und Ende der Wortlinie wählen.", 0
	}
	if !(dz == 0 || ds == 0 || dz == ds || dz == -ds) {
		return "Wörter liegen nur waagerecht, senkrecht oder schräg.", 0
	}
	laenge := absInt(dz)
	if absInt(ds) > laenge {
		laenge = absInt(ds)
	}
	laenge++
	if laenge != len([]rune(wort)) {
		return "Die Auswahl passt nicht zur Wortlänge.", 0
	}
	// Vom Startfeld in Richtung Endfeld laufen: die Vorzeichen geben sie vor.
	schrittZ, schrittS := vorzeichen(dz), vorzeichen(ds)
	// Buchstaben vom Startfeld in Richtung Endfeld lesen – das Wort darf auch
	// rückwärts ausgewählt werden (Ende zuerst).
	zellen := make([][2]int, 0, laenge)
	gelesen := make([]rune, 0, laenge)
	for k := 0; k < laenge; k++ {
		z := von[0] + k*schrittZ
		s := von[1] + k*schrittS
		if z < 0 || s < 0 || z >= g.Kante || s >= g.Kante {
			return "Die Auswahl verlässt das Gitter.", 0
		}
		zellen = append(zellen, [2]int{z, s})
		gelesen = append(gelesen, []rune(g.Gitter[z][s])[0])
	}
	if string(gelesen) != wort {
		umgekehrt := make([]rune, len(gelesen))
		for i := range gelesen {
			umgekehrt[len(gelesen)-1-i] = gelesen[i]
		}
		if string(umgekehrt) != wort {
			return "Die Buchstaben passen nicht zu „" + wort + "“.", 0
		}
		for i, j := 0, len(zellen)-1; i < j; i, j = i+1, j-1 {
			zellen[i], zellen[j] = zellen[j], zellen[i]
		}
	}
	// Gültig: Punkte nach Wortlänge, Wort ist für alle gefunden.
	punkte := len([]rune(wort)) * wortsuchePunktFak
	g.Spieler[sitz].Punkte += punkte
	g.Spieler[sitz].Treffer++
	g.Funde = append(g.Funde, WortsucheFund{
		Wort: wort, VonID: g.Spieler[sitz].ID, VonName: g.Spieler[sitz].Name,
		Punkte: punkte, Zellen: zellen,
	})
	g.Note = g.Spieler[sitz].Name + " findet „" + wort + "“ (+" + strconv.Itoa(punkte) + ")."
	if len(g.Funde) == len(g.Worte) {
		g.wortsucheBeenden("Alle Wörter gefunden.")
	}
	return "", punkte
}

func wortsucheAnzeige(g *WortsucheGame) *WortsucheGame {
	if g == nil {
		return nil
	}
	g.wortsuchePruefeZeit()
	c := *g
	c.Gitter = make([][]string, len(g.Gitter))
	for i := range g.Gitter {
		c.Gitter[i] = append([]string{}, g.Gitter[i]...)
	}
	c.Worte = append([]string{}, g.Worte...)
	c.Funde = append([]WortsucheFund{}, g.Funde...)
	for i := range c.Funde {
		c.Funde[i].Zellen = append([][2]int{}, g.Funde[i].Zellen...)
	}
	c.Spieler = append([]WortsucheSpieler{}, g.Spieler...)
	return &c
}

func wortsucheHandler(w http.ResponseWriter, r *http.Request) {
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
		// Der Computer findet beim Abholen des Stands höchstens ein Wort.
		wortsucheBotTreiben(room.Wortsuche, time.Now())
		game := wortsucheAnzeige(room.Wortsuche)
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
		Wort  string `json:"wort"`
		Von   [2]int `json:"von"`
		Bis   [2]int `json:"bis"`
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
	// amTisch liefert die Runde in der erwarteten Lage und den eigenen Platz.
	amTisch := func(status, fehlt, nichtDabei string) (*WortsucheGame, int, string, int) {
		g := room.Wortsuche
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
		// Mensch gegen Computer: beide sitzen, die Runde läuft sofort.
		room.botNamenNeu()
		if room.Wortsuche != nil && !room.gameSlotFree(room.Wortsuche.Status, wortsucheIDs(room.Wortsuche)...) {
			fail("In diesem Spielraum läuft bereits eine Wortsuche.", http.StatusConflict)
			return
		}
		stufe := botStufe(input.Stufe)
		g := wortsucheNeuesSpiel(input.ID, player.Name, randomCode())
		g.Spieler[0].Ready = true
		g.Spieler = append(g.Spieler, WortsucheSpieler{ID: botID(1), Name: room.botName(1), Ready: true})
		g.BotID, g.BotStufe = botID(1), stufe
		g.Runde = 1
		g.wortsucheRunde()
		g.BotFrist = time.Now().Add(wortsucheBotPause(stufe))
		g.Note = player.Name + " spielt gegen " + room.botName(1) + " (" + botStufeText(stufe) + "). " + g.Note
		room.Wortsuche = g
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Wortsuche gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     uhrzeit(time.Now()),
			SenderID: input.ID,
		})
	case "create":
		if room.Wortsuche != nil && !room.gameSlotFree(room.Wortsuche.Status, wortsucheIDs(room.Wortsuche)...) {
			fail("In diesem Spielraum läuft bereits eine Wortsuche.", http.StatusConflict)
			return
		}
		room.Wortsuche = wortsucheNeuesSpiel(input.ID, player.Name, randomCode())
		room.Wortsuche.Note = player.Name + " eröffnet eine Runde Wortsuche."
	case "join":
		g := room.Wortsuche
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Runde.", http.StatusConflict)
			return
		}
		if g.sitz(input.ID) >= 0 {
			fail("Du sitzt schon am Tisch.", http.StatusConflict)
			return
		}
		if len(g.Spieler) >= wortsucheMaxSpieler {
			fail("Der Tisch ist voll (höchstens "+strconv.Itoa(wortsucheMaxSpieler)+" Spieler).", http.StatusConflict)
			return
		}
		g.Spieler = append(g.Spieler, WortsucheSpieler{ID: input.ID, Name: player.Name})
	case "ready":
		g, sitz, problem, code := amTisch("invited", "Es gibt keine offene Runde.", "Du sitzt nicht an diesem Tisch.")
		if problem != "" {
			fail(problem, code)
			return
		}
		g.Spieler[sitz].Ready = !g.Spieler[sitz].Ready
	case "start":
		g, _, problem, code := amTisch("invited", "Es gibt keine offene Runde.", "Du sitzt nicht an diesem Tisch.")
		if problem != "" {
			fail(problem, code)
			return
		}
		if len(g.Spieler) < 2 {
			fail("Die Wortsuche braucht mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, s := range g.Spieler {
			if !s.Ready {
				fail("Noch nicht alle sind bereit.", http.StatusConflict)
				return
			}
		}
		g.Runde = 1
		g.wortsucheRunde()
	case "finden":
		g, sitz, problem, code := amTisch("spiel", "Gerade läuft keine Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, code)
			return
		}
		text, _ := g.wortsucheFinden(sitz, input.Wort, input.Von, input.Bis)
		if text != "" {
			g.Spieler[sitz].Fehlversuch++
			fail(text, http.StatusConflict)
			return
		}
	case "naechste":
		g, _, problem, code := amTisch("ende", "Es läuft keine beendete Runde.", "Du spielst nicht mit.")
		if problem != "" {
			fail(problem, code)
			return
		}
		g.Runde++
		g.wortsucheRunde()
	case "reset":
		if room.Wortsuche != nil && room.Wortsuche.HostID != input.ID {
			fail("Nur der Gastgeber kann die Runde schließen.", http.StatusForbidden)
			return
		}
		room.Wortsuche = nil
	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}
	room.LastActive = time.Now()
	game := wortsucheAnzeige(room.Wortsuche)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": game})
}

func vorzeichen(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	}
	return 0
}

func wortsucheIDs(g *WortsucheGame) []string {
	ids := make([]string, 0, len(g.Spieler))
	for _, s := range g.Spieler {
		ids = append(ids, s.ID)
	}
	return ids
}

// wortsucheRemoveSeat nimmt einen Spieler aus der Runde (Verlassen/Abbruch).
func wortsucheRemoveSeat(g *WortsucheGame, id, name string) (bool, bool, int) {
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
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Spieler)) + " Spielern weiter."
	return true, false, len(g.Spieler)
}
