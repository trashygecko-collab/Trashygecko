package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Snake als Online-Duell über den Connector: Zwei Schlangen auf einem Feld,
// ein Apfel, und der Server rechnet jeden Schritt. Wer den Rand, sich selbst
// oder die andere Schlange berührt, verliert die Runde; drei Rundensiege
// gewinnen das Match. Feld und Start wie in der lokalen Fassung
// (spiele/snake/snake.html): 21 × 23 Kacheln, Kopf bei (5,11) nach rechts,
// Richtungswechsel höchstens drei auf Vorrat, Umkehren verboten.

const (
	snaSpalten       = 21
	snaZeilen        = 23
	snaTakt          = 0.2 // Sekunden je Schritt (fünf Schritte/s)
	snaZiel          = 3
	snaMaxSchritte   = 240
	snaVorsprung     = 3.0
	snaMaxSpieler    = 2
	snaWarteschlange = 3
)

type SnaPunkt struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type SnakeSpieler struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Ready    bool       `json:"ready"`
	Nr       int        `json:"-"`
	Segmente []SnaPunkt `json:"-"` // [0] ist der Kopf
	Wartet   []SnaPunkt `json:"-"`
	Richtung SnaPunkt   `json:"-"`
	Leben    bool       `json:"-"`
	Sieg     int        `json:"-"`
	Punkte   int        `json:"-"`
}

type SnakeGame struct {
	ID       string         `json:"id"`
	HostID   string         `json:"hostId"`
	Status   string         `json:"status"` // invited, spiel, runde, ende
	Spieler  []SnakeSpieler `json:"spieler"`
	Apfel    SnaPunkt       `json:"-"`
	HatApfel bool           `json:"-"`
	Runde    int            `json:"-"`
	Gewinner int            `json:"-"`
	Meldung  string         `json:"meldung"`
	Note     string         `json:"note"`
	Sieger   string         `json:"sieger,omitempty"`
	Zuletzt  time.Time      `json:"-"`
	Rest     float64        `json:"-"`
	Zufall   *rand.Rand     `json:"-"`
	BotID    string         `json:"-"`
	BotStufe string         `json:"-"`
}

type SnakeSpielerSicht struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Ready    bool     `json:"ready"`
	Nr       int      `json:"nr"`
	Laenge   int      `json:"laenge"`
	Punkte   int      `json:"punkte"`
	Sieg     int      `json:"sieg"`
	Tot      bool     `json:"tot"`
	Richtung string   `json:"richtung"`
	Kopf     SnaPunkt `json:"kopf"`
}

type SnakeAnsicht struct {
	ID       string              `json:"id"`
	HostID   string              `json:"hostId"`
	Status   string              `json:"status"`
	Spalten  int                 `json:"spalten"`
	Zeilen   int                 `json:"zeilen"`
	Runde    int                 `json:"runde"`
	Ziel     int                 `json:"ziel"`
	Raster   string              `json:"raster"`
	Apfel    *SnaPunkt           `json:"apfel"`
	Spieler  []SnakeSpielerSicht `json:"spieler"`
	Gewinner int                 `json:"gewinner"`
	Meldung  string              `json:"meldung"`
	Note     string              `json:"note"`
	Sieger   string              `json:"sieger,omitempty"`
	Du       int                 `json:"du"`
	Tick     int64               `json:"tick"`
}

func snaName(p SnaPunkt) string {
	return strconv.Itoa(p.X) + "," + strconv.Itoa(p.Y)
}

func snakeSitz(g *SnakeGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// snakeNeuerSpieler setzt Startplatz und Blickrichtung: Spieler 1 links mit
// Kopf nach rechts, Spieler 2 rechts mit Kopf nach links.
func snakeNeuerSpieler(nr int, id, name string) SnakeSpieler {
	if nr == 1 {
		return SnakeSpieler{
			ID: id, Name: name, Nr: 1, Leben: true,
			Segmente: []SnaPunkt{{5, 11}, {4, 11}, {3, 11}},
			Richtung: SnaPunkt{1, 0},
		}
	}
	return SnakeSpieler{
		ID: id, Name: name, Nr: 2, Leben: true,
		Segmente: []SnaPunkt{{snaSpalten - 6, 11}, {snaSpalten - 5, 11}, {snaSpalten - 4, 11}},
		Richtung: SnaPunkt{-1, 0},
	}
}

func snakeNeu(ID, hostID string, spieler []SnakeSpieler, saat int64) *SnakeGame {
	g := &SnakeGame{
		ID:       ID,
		HostID:   hostID,
		Status:   "invited",
		Spieler:  spieler,
		Runde:    1,
		Gewinner: -1,
		Zuletzt:  time.Now(),
		Zufall:   rand.New(rand.NewSource(saat)),
		Note:     "Tisch eröffnet – gesucht wird eine zweite Schlange.",
	}
	return g
}

// snakeRaster baut das Feld: 0 frei, 1 erste Schlange, 2 zweite Schlange.
func (g *SnakeGame) raster() []uint8 {
	r := make([]uint8, snaSpalten*snaZeilen)
	for _, p := range g.Spieler {
		for _, teil := range p.Segmente {
			if teil.X < 0 || teil.Y < 0 || teil.X >= snaSpalten || teil.Y >= snaZeilen {
				continue
			}
			r[teil.Y*snaSpalten+teil.X] = uint8(p.Nr)
		}
	}
	return r
}

// apfelSetzen legt einen Apfel auf eine freie Kachel.
func (g *SnakeGame) apfelSetzen() {
	belegt := map[SnaPunkt]bool{}
	for _, p := range g.Spieler {
		for _, teil := range p.Segmente {
			belegt[teil] = true
		}
	}
	frei := make([]SnaPunkt, 0, snaSpalten*snaZeilen)
	for y := 0; y < snaZeilen; y++ {
		for x := 0; x < snaSpalten; x++ {
			if !belegt[SnaPunkt{x, y}] {
				frei = append(frei, SnaPunkt{x, y})
			}
		}
	}
	if len(frei) == 0 {
		g.HatApfel = false
		return
	}
	g.Apfel = frei[g.Zufall.Intn(len(frei))]
	g.HatApfel = true
}

// snakeNeueRunde setzt Schlangen, Apfel und Zähler zurück, behält die Siege.
func (g *SnakeGame) snakeNeueRunde() {
	for i := range g.Spieler {
		nr := g.Spieler[i].Nr
		sieg := g.Spieler[i].Sieg
		ready := g.Spieler[i].Ready
		g.Spieler[i] = snakeNeuerSpieler(nr, g.Spieler[i].ID, g.Spieler[i].Name)
		g.Spieler[i].Sieg = sieg
		g.Spieler[i].Ready = ready
	}
	g.apfelSetzen()
	g.Gewinner = -1
	g.Meldung = ""
	g.Rest = 0
	g.Zuletzt = time.Now()
}

// richtungMerken nimmt höchstens drei Richtungswechsel auf Vorrat an;
// Umkehren und Wiederholungen fallen weg.
func (s *SnakeSpieler) richtungMerken(d SnaPunkt) bool {
	letzte := s.Richtung
	if len(s.Wartet) > 0 {
		letzte = s.Wartet[len(s.Wartet)-1]
	}
	if letzte.X == -d.X && letzte.Y == -d.Y {
		return false
	}
	if letzte.X == d.X && letzte.Y == d.Y {
		return false
	}
	if len(s.Wartet) >= snaWarteschlange {
		return false
	}
	s.Wartet = append(s.Wartet, d)
	return true
}

// wuerdeWachsen sagt, ob der nächste Kopf auf dem Apfel landet.
func (g *SnakeGame) wuerdeWachsen(s *SnakeSpieler, kopf SnaPunkt) bool {
	return g.HatApfel && kopf.X == g.Apfel.X && kopf.Y == g.Apfel.Y
}

// snakeSchritt rechnet einen Schritt für beide Schlangen.
func (g *SnakeGame) snakeSchritt() {
	if g.Status != "spiel" {
		return
	}
	// Richtung übernehmen
	for i := range g.Spieler {
		if len(g.Spieler[i].Wartet) > 0 {
			g.Spieler[i].Richtung = g.Spieler[i].Wartet[0]
			g.Spieler[i].Wartet = g.Spieler[i].Wartet[1:]
		}
	}

	koepfe := make([]SnaPunkt, len(g.Spieler))
	wachsen := make([]bool, len(g.Spieler))
	for i := range g.Spieler {
		if !g.Spieler[i].Leben {
			continue
		}
		kopf := g.Spieler[i].Segmente[0]
		koepfe[i] = SnaPunkt{kopf.X + g.Spieler[i].Richtung.X, kopf.Y + g.Spieler[i].Richtung.Y}
		wachsen[i] = g.wuerdeWachsen(&g.Spieler[i], koepfe[i])
	}

	// Belegte Kacheln: der Schwanz fällt weg, wenn die Schlange nicht wächst.
	belegt := map[SnaPunkt]int{}
	for i := range g.Spieler {
		if !g.Spieler[i].Leben {
			continue
		}
		teile := g.Spieler[i].Segmente
		ende := len(teile)
		if !wachsen[i] && ende > 0 {
			ende--
		}
		for k := 0; k < ende; k++ {
			belegt[teile[k]] = g.Spieler[i].Nr
		}
	}

	tot := make([]bool, len(g.Spieler))
	for i := range g.Spieler {
		if !g.Spieler[i].Leben {
			continue
		}
		k := koepfe[i]
		if k.X < 0 || k.Y < 0 || k.X >= snaSpalten || k.Y >= snaZeilen {
			tot[i] = true
			continue
		}
		if _, da := belegt[k]; da {
			tot[i] = true
			continue
		}
	}
	// Kopf an Kopf und Kreuztausch: beide sterben.
	if len(g.Spieler) == 2 && g.Spieler[0].Leben && g.Spieler[1].Leben {
		if koepfe[0] == koepfe[1] {
			tot[0], tot[1] = true, true
		}
		k0, k1 := g.Spieler[0].Segmente[0], g.Spieler[1].Segmente[0]
		if koepfe[0] == k1 && koepfe[1] == k0 {
			tot[0], tot[1] = true, true
		}
	}

	for i := range g.Spieler {
		if !g.Spieler[i].Leben || tot[i] {
			if tot[i] {
				g.Spieler[i].Leben = false
			}
			continue
		}
		s := &g.Spieler[i]
		s.Segmente = append([]SnaPunkt{koepfe[i]}, s.Segmente...)
		if wachsen[i] {
			s.Punkte += 10
			g.apfelSetzen()
		} else {
			s.Segmente = s.Segmente[:len(s.Segmente)-1]
		}
	}

	g.rundePruefen()
}

// rundePruefen beendet die Runde, sobald höchstens eine Schlange lebt.
func (g *SnakeGame) rundePruefen() {
	lebend := 0
	letzte := -1
	for i := range g.Spieler {
		if g.Spieler[i].Leben {
			lebend++
			letzte = i
		}
	}
	if lebend > 1 {
		return
	}
	g.Status = "runde"
	if lebend == 1 {
		gewinner := &g.Spieler[letzte]
		g.Gewinner = gewinner.Nr
		gewinner.Sieg++
		g.Meldung = gewinner.Name + " gewinnt Runde " + strconv.Itoa(g.Runde) + " und steht bei " +
			strconv.Itoa(gewinner.Sieg) + " Siegen."
		if gewinner.Sieg >= snaZiel {
			g.Status = "ende"
			g.Sieger = gewinner.Name
			g.Meldung = gewinner.Name + " gewinnt das Match mit " + strconv.Itoa(gewinner.Sieg) + " Rundensiegen."
			g.Note = "Neue Partie startet mit einem Klick."
			return
		}
	} else {
		g.Gewinner = 0
		g.Meldung = "Beide Schlangen sind gleichzeitig gestorben. Runde " + strconv.Itoa(g.Runde) + " wird wiederholt."
	}
	g.Note = "Weiter mit der nächsten Runde – wer zuerst drückt, startet sie."
}

// snakeWeiter lässt die Partie bis jetzt laufen.
func snakeWeiter(g *SnakeGame) {
	if g == nil || g.Status != "spiel" {
		return
	}
	jetzt := time.Now()
	dt := jetzt.Sub(g.Zuletzt).Seconds()
	if dt < 0 {
		dt = 0
	}
	if dt > snaVorsprung {
		dt = snaVorsprung
	}
	g.Zuletzt = jetzt
	g.Rest += dt
	schritte := 0
	for g.Rest >= snaTakt && g.Status == "spiel" && schritte < snaMaxSchritte {
		g.Rest -= snaTakt
		schritte++
		// Der Bot setzt seine Richtung im selben Takt, in dem auch ein Mensch
		// sie setzen würde – vor dem Schritt, durch dieselbe Prüfung.
		snakeBotRichtung(g, g.BotStufe)
		g.snakeSchritt()
	}
	if schritte >= snaMaxSchritte {
		g.Rest = 0
	}
}

func snakeRichtungName(d SnaPunkt) string {
	switch {
	case d.X > 0:
		return "rechts"
	case d.X < 0:
		return "links"
	case d.Y > 0:
		return "runter"
	}
	return "hoch"
}

func snakeAnsicht(g *SnakeGame, id string) *SnakeAnsicht {
	if g == nil {
		return nil
	}
	a := &SnakeAnsicht{
		ID:       g.ID,
		HostID:   g.HostID,
		Status:   g.Status,
		Spalten:  snaSpalten,
		Zeilen:   snaZeilen,
		Runde:    g.Runde,
		Ziel:     snaZiel,
		Raster:   tronRLE(g.raster()),
		Gewinner: g.Gewinner,
		Meldung:  g.Meldung,
		Note:     g.Note,
		Sieger:   g.Sieger,
		Du:       snakeSitz(g, id),
		Tick:     time.Now().UnixMilli(),
	}
	if g.HatApfel {
		apfel := g.Apfel
		a.Apfel = &apfel
	}
	a.Spieler = []SnakeSpielerSicht{}
	for _, p := range g.Spieler {
		a.Spieler = append(a.Spieler, SnakeSpielerSicht{
			ID: p.ID, Name: p.Name, Ready: p.Ready, Nr: p.Nr,
			Laenge: len(p.Segmente), Punkte: p.Punkte, Sieg: p.Sieg,
			Tot: !p.Leben, Richtung: snakeRichtungName(p.Richtung), Kopf: p.Segmente[0],
		})
	}
	return a
}

// ---------------------------------------------------------------- Endpunkt

func snakeHandler(w http.ResponseWriter, r *http.Request) {
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
		snakeWeiter(room.Snake)
		game := snakeAnsicht(room.Snake, id)
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
		Dir   string `json:"dir"`
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
	zeit := uhrzeit(time.Now())

	switch action {
	// Gegen den Server-Bot: Mensch links, Computer rechts, beide bereit.
	case "bot":
		if room.Snake != nil && room.Snake.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Partie Snake.", http.StatusConflict)
			return
		}
		stufe := botStufe(input.Stufe)
		room.botNamenNeu()
		spiel := snakeNeu(randomCode(), input.ID, []SnakeSpieler{
			snakeNeuerSpieler(1, input.ID, player.Name),
			snakeNeuerSpieler(2, botID(1), room.botName(1)),
		}, time.Now().UnixNano())
		spiel.BotID, spiel.BotStufe = botID(1), stufe
		for i := range spiel.Spieler {
			spiel.Spieler[i].Ready = true
			spiel.Spieler[i].Sieg = 0
		}
		spiel.Runde = 1
		spiel.Status = "spiel"
		spiel.snakeNeueRunde()
		spiel.Note = player.Name + " spielt gegen " + room.botName(1) +
			" (" + botStufeText(stufe) + ")."
		room.Snake = spiel
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Snake gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "create":
		if room.Snake != nil && room.Snake.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Partie Snake.", http.StatusConflict)
			return
		}
		room.Snake = snakeNeu(randomCode(), input.ID,
			[]SnakeSpieler{snakeNeuerSpieler(1, input.ID, player.Name)}, time.Now().UnixNano())
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Partie Snake – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Snake
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if snakeSitz(game, input.ID) >= 0 {
			break
		}
		if len(game.Spieler) >= snaMaxSpieler {
			mu.Unlock()
			http.Error(w, "Snake ist ein Duell – der Tisch ist voll.", http.StatusConflict)
			return
		}
		game.Spieler = append(game.Spieler, snakeNeuerSpieler(2, input.ID, player.Name))
		game.Note = player.Name + " sitzt mit am Tisch. Beide drücken BEREIT."

	case "ready":
		game := room.Snake
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := snakeSitz(game, input.ID)
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
		game := room.Snake
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Partie auf den Start.", http.StatusConflict)
			return
		}
		if len(game.Spieler) < snaMaxSpieler {
			mu.Unlock()
			http.Error(w, "Snake braucht zwei Spieler.", http.StatusConflict)
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
		for i := range game.Spieler {
			game.Spieler[i].Sieg = 0
		}
		game.Runde = 1
		game.Sieger = ""
		game.Status = "spiel"
		game.snakeNeueRunde()
		game.Note = "Los! Apfel fressen, nicht anstoßen."

	case "richtung":
		game := room.Snake
		if game == nil || game.Status != "spiel" {
			mu.Unlock()
			http.Error(w, "Gerade läuft keine Runde.", http.StatusConflict)
			return
		}
		sitz := snakeSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		var d SnaPunkt
		switch strings.ToLower(strings.TrimSpace(input.Dir)) {
		case "hoch":
			d = SnaPunkt{0, -1}
		case "runter":
			d = SnaPunkt{0, 1}
		case "links":
			d = SnaPunkt{-1, 0}
		case "rechts":
			d = SnaPunkt{1, 0}
		default:
			mu.Unlock()
			http.Error(w, "Unbekannte Richtung.", http.StatusBadRequest)
			return
		}
		if !game.Spieler[sitz].richtungMerken(d) {
			mu.Unlock()
			http.Error(w, "Diese Richtung geht gerade nicht.", http.StatusConflict)
			return
		}

	case "weiter":
		game := room.Snake
		if game == nil || game.Status != "runde" {
			mu.Unlock()
			http.Error(w, "Die Runde läuft noch.", http.StatusConflict)
			return
		}
		if snakeSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Gewinner != 0 {
			game.Runde++
		}
		game.Status = "spiel"
		game.snakeNeueRunde()
		game.Note = "Los! Apfel fressen, nicht anstoßen."

	case "reset":
		game := room.Snake
		if game == nil || game.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Eine neue Partie geht erst nach dem Ende.", http.StatusConflict)
			return
		}
		if snakeSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		frisch := snakeNeu(game.ID, game.HostID, []SnakeSpieler{}, time.Now().UnixNano())
		frisch.Note = "Neue Partie – erst BEREIT drücken."
		for _, p := range game.Spieler {
			frisch.Spieler = append(frisch.Spieler, snakeNeuerSpieler(p.Nr, p.ID, p.Name))
		}
		frisch.BotID, frisch.BotStufe = game.BotID, game.BotStufe
		if frisch.BotID != "" {
			// Der Bot drückt nicht selbst: er ist bereit, die Partie startet
			// mit demselben Klick.
			for i := range frisch.Spieler {
				frisch.Spieler[i].Ready = true
				frisch.Spieler[i].Sieg = 0
			}
			frisch.Runde = 1
			frisch.Status = "spiel"
			frisch.snakeNeueRunde()
			frisch.Note = "Los! Apfel fressen, nicht anstoßen."
		}
		room.Snake = frisch

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	snakeWeiter(room.Snake)
	result := snakeAnsicht(room.Snake, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
