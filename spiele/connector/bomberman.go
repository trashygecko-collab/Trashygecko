package main

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Bomberman für zwei Spieler über den Connector: Der Server führt Karte,
// Bomben, Flammen, Extras und Runden. Die Geräte schicken nur ihre Richtung
// und „Bombe legen“ und holen den Stand ab – so laufen beide exakt dieselbe
// Partie. Regeln und Maße sind dieselben wie in der lokalen Fassung
// (spiele/bomberman/bomberman.html): 15 × 13 Kacheln, Bombe nach 2,4 s,
// Flamme 0,62 s, Tempo 3,05 Kacheln je Sekunde, drei Rundensiege.

const (
	bomberSpalten      = 15
	bomberZeilen       = 13
	bomberRadius       = 0.33 // Spielerradius in Kacheln
	bomberBombenzeit   = 2.4
	bomberFlammenzeit  = 0.62
	bomberZiel         = 3
	bomberVorlauf      = 1.9
	bomberTempo        = 3.05
	bomberMaxSpieler   = 2
	bomberTakt         = 1.0 / 60.0
	bomberMaxSchritte  = 240 // höchstens vier Sekunden auf einmal nachrechnen
	bomberMaxVorsprung = 3.0
)

type BomberSpieler struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Ready   bool    `json:"ready"`
	Nr      int     `json:"-"`
	X       float64 `json:"-"`
	Y       float64 `json:"-"`
	DX      int     `json:"-"` // gehaltene Richtung
	DY      int     `json:"-"`
	BlickDX int     `json:"-"` // zuletzt gelaufene Richtung (für die Augen)
	BlickDY int     `json:"-"`
	BombMax int     `json:"-"`
	Feuer   int     `json:"-"`
	Tempo   float64 `json:"-"`
	Leben   bool    `json:"-"`
	Sieg    int     `json:"-"`
	LaufT   float64 `json:"-"`
	TotSeit float64 `json:"-"`
}

type BomberBombe struct {
	X      int     `json:"-"`
	Y      int     `json:"-"`
	Nr     int     `json:"-"`
	Rest   float64 `json:"-"`
	Radius int     `json:"-"`
	Frei   int     `json:"-"` // dieser Spieler darf noch herauslaufen
	Phase  float64 `json:"-"`
}

type BomberFlamme struct {
	X     int     `json:"-"`
	Y     int     `json:"-"`
	Rest  float64 `json:"-"`
	Phase float64 `json:"-"`
}

type BomberPowerup struct {
	X     int     `json:"-"`
	Y     int     `json:"-"`
	Art   string  `json:"-"`
	Phase float64 `json:"-"`
}

type BomberGame struct {
	ID       string          `json:"id"`
	HostID   string          `json:"hostId"`
	Status   string          `json:"status"` // invited, spiel, runde, ende
	Spieler  []BomberSpieler `json:"spieler"`
	Raster   []uint8         `json:"-"`
	Bomben   []*BomberBombe  `json:"-"`
	Flammen  []BomberFlamme  `json:"-"`
	Powerups []BomberPowerup `json:"-"`
	Runde    int             `json:"-"`
	Vorlauf  float64         `json:"-"`
	Gewinner int             `json:"-"`
	Meldung  string          `json:"meldung"`
	Note     string          `json:"note"`
	Sieger   string          `json:"sieger,omitempty"`
	Zuletzt  time.Time       `json:"-"`
	Rest     float64         `json:"-"`

	Bots      []string   `json:"bots,omitempty"`
	BotStufe  string     `json:"botStufe,omitempty"`
	BotFrist  time.Time  `json:"-"`
	BotZufall *rand.Rand `json:"-"`
}

type BomberSpielerSicht struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Ready   bool    `json:"ready"`
	Nr      int     `json:"nr"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	DX      int     `json:"dx"`
	DY      int     `json:"dy"`
	BlickDX int     `json:"blickX"`
	BlickDY int     `json:"blickY"`
	BombMax int     `json:"bombenMax"`
	Feuer   int     `json:"feuer"`
	Tempo   float64 `json:"tempo"`
	Leben   bool    `json:"leben"`
	Sieg    int     `json:"sieg"`
	LaufT   float64 `json:"laufT"`
	TotSeit float64 `json:"totSeit"`
}

type BomberBombeSicht struct {
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Nr     int     `json:"nr"`
	Rest   float64 `json:"rest"`
	Radius int     `json:"radius"`
	Phase  float64 `json:"phase"`
}

type BomberFlammeSicht struct {
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Rest  float64 `json:"rest"`
	Phase float64 `json:"phase"`
}

type BomberPowerupSicht struct {
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Art   string  `json:"art"`
	Phase float64 `json:"phase"`
}

type BomberAnsicht struct {
	ID       string               `json:"id"`
	HostID   string               `json:"hostId"`
	Status   string               `json:"status"`
	Spalten  int                  `json:"spalten"`
	Zeilen   int                  `json:"zeilen"`
	Runde    int                  `json:"runde"`
	Ziel     int                  `json:"ziel"`
	Vorlauf  float64              `json:"vorlauf"`
	Raster   string               `json:"raster"`
	Spieler  []BomberSpielerSicht `json:"spieler"`
	Bomben   []BomberBombeSicht   `json:"bomben"`
	Flammen  []BomberFlammeSicht  `json:"flammen"`
	Powerups []BomberPowerupSicht `json:"powerups"`
	Gewinner int                  `json:"gewinner"`
	Meldung  string               `json:"meldung"`
	Note     string               `json:"note"`
	Sieger   string               `json:"sieger,omitempty"`
	Du       int                  `json:"du"`
	Tick     int64                `json:"tick"`
}

// ---------------------------------------------------------------- Aufbau

var bomberEcken = [4][2]int{{1, 1}, {13, 11}, {1, 11}, {13, 1}}

func bomberNeueKarte() []uint8 {
	raster := make([]uint8, bomberSpalten*bomberZeilen)
	for y := 0; y < bomberZeilen; y++ {
		for x := 0; x < bomberSpalten; x++ {
			randfrei := x == 0 || y == 0 || x == bomberSpalten-1 || y == bomberZeilen-1
			pfeiler := x%2 == 0 && y%2 == 0
			if randfrei || pfeiler {
				raster[y*bomberSpalten+x] = 1
			}
		}
	}
	for y := 1; y < bomberZeilen-1; y++ {
		for x := 1; x < bomberSpalten-1; x++ {
			if raster[y*bomberSpalten+x] != 0 {
				continue
			}
			if rand.Float64() < 0.58 {
				raster[y*bomberSpalten+x] = 2
			}
		}
	}
	// Startecken freihalten – wie im lokalen Spiel erst nach den Kisten.
	for _, e := range bomberEcken {
		for _, d := range [][2]int{{0, 0}, {1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
			x, y := e[0]+d[0], e[1]+d[1]
			if x > 0 && y > 0 && x < bomberSpalten-1 && y < bomberZeilen-1 {
				raster[y*bomberSpalten+x] = 0
			}
		}
	}
	return raster
}

func bomberNeuerSpieler(nr int, id, name string) BomberSpieler {
	x, y := bomberEcken[(nr-1)%4][0], bomberEcken[(nr-1)%4][1]
	return BomberSpieler{
		ID: id, Name: name, Nr: nr,
		X: float64(x) + 0.5, Y: float64(y) + 0.5,
		BombMax: 1, Feuer: 1, Tempo: 1, Leben: true,
	}
}

func bomberSitz(g *BomberGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// bomberNeueRunde setzt Karte, Bomben und Spieler zurück, behält die Siege.
func bomberNeueRunde(g *BomberGame) {
	g.Raster = bomberNeueKarte()
	g.Bomben = nil
	g.Flammen = nil
	g.Powerups = nil
	for i := range g.Spieler {
		nr := g.Spieler[i].Nr
		sieg := g.Spieler[i].Sieg
		ready := g.Spieler[i].Ready
		g.Spieler[i] = bomberNeuerSpieler(nr, g.Spieler[i].ID, g.Spieler[i].Name)
		g.Spieler[i].Sieg = sieg
		g.Spieler[i].Ready = ready
	}
	g.Vorlauf = bomberVorlauf
	g.Gewinner = -1
	g.Rest = 0
	g.Zuletzt = time.Now()
	g.Meldung = ""
}

func bomberNeu(ID, hostID string, spieler []BomberSpieler) *BomberGame {
	g := &BomberGame{
		ID:       ID,
		HostID:   hostID,
		Status:   "invited",
		Spieler:  spieler,
		Runde:    1,
		Gewinner: -1,
		Zuletzt:  time.Now(),
		Note:     "Tisch eröffnet – gesucht wird ein zweiter Bomber.",
	}
	g.Raster = bomberNeueKarte()
	return g
}

// ---------------------------------------------------------------- Regeln

func (g *BomberGame) kachel(x, y int) uint8 {
	if x < 0 || y < 0 || x >= bomberSpalten || y >= bomberZeilen {
		return 1
	}
	return g.Raster[y*bomberSpalten+x]
}

// begehbar prüft die vier Ecken des Bombers und alle Bomben. Die eigene
// Bombe darf man verlassen, solange man sie noch berührt.
func (g *BomberGame) begehbar(x, y float64, s *BomberSpieler) bool {
	for _, ox := range []float64{-bomberRadius, bomberRadius} {
		for _, oy := range []float64{-bomberRadius, bomberRadius} {
			if g.kachel(int(math.Floor(x+ox)), int(math.Floor(y+oy))) != 0 {
				return false
			}
		}
	}
	for _, b := range g.Bomben {
		bx, by := float64(b.X), float64(b.Y)
		if b.Frei == s.Nr {
			if x+bomberRadius <= bx || x-bomberRadius >= bx+1 ||
				y+bomberRadius <= by || y-bomberRadius >= by+1 {
				b.Frei = 0
			}
			continue
		}
		if x+bomberRadius > bx && x-bomberRadius < bx+1 &&
			y+bomberRadius > by && y-bomberRadius < by+1 {
			return false
		}
	}
	return true
}

func (g *BomberGame) bewege(s *BomberSpieler, dt float64) {
	if !s.Leben || g.Vorlauf > 0 {
		return
	}
	if s.DX == 0 && s.DY == 0 {
		return
	}
	schritt := bomberTempo * s.Tempo * dt
	nx, ny := s.X, s.Y
	glätten := func(v, dt float64) float64 {
		ziel := (math.Floor(v) + 0.5) - v
		if ziel > 8*dt {
			ziel = 8 * dt
		}
		if ziel < -8*dt {
			ziel = -8 * dt
		}
		return v + ziel
	}
	if s.DX != 0 {
		ny = glätten(s.Y, dt)
		ziel := s.X + float64(s.DX)*schritt
		if g.begehbar(ziel, ny, s) {
			nx = ziel
		}
	} else if s.DY != 0 {
		nx = glätten(s.X, dt)
		ziel := s.Y + float64(s.DY)*schritt
		if g.begehbar(nx, ziel, s) {
			ny = ziel
		}
	}
	if nx != s.X || ny != s.Y {
		s.X, s.Y = nx, ny
		s.LaufT += dt
		s.BlickDX, s.BlickDY = s.DX, s.DY
	}

	for i := len(g.Powerups) - 1; i >= 0; i-- {
		p := &g.Powerups[i]
		if p.X == int(math.Floor(s.X)) && p.Y == int(math.Floor(s.Y)) {
			switch p.Art {
			case "bombe":
				if s.BombMax < 6 {
					s.BombMax++
				}
			case "feuer":
				if s.Feuer < 6 {
					s.Feuer++
				}
			case "tempo":
				if s.Tempo*1.22 < 1.8 {
					s.Tempo *= 1.22
				} else {
					s.Tempo = 1.8
				}
			}
			g.Powerups = append(g.Powerups[:i], g.Powerups[i+1:]...)
		}
	}
}

func (g *BomberGame) legeBombe(s *BomberSpieler) bool {
	if g.Status != "spiel" || g.Vorlauf > 0 || !s.Leben {
		return false
	}
	bx := int(math.Floor(s.X))
	by := int(math.Floor(s.Y))
	eigene := 0
	for _, b := range g.Bomben {
		if b.X == bx && b.Y == by {
			return false
		}
		if b.Nr == s.Nr {
			eigene++
		}
	}
	if eigene >= s.BombMax {
		return false
	}
	g.Bomben = append(g.Bomben, &BomberBombe{
		X: bx, Y: by, Nr: s.Nr, Rest: bomberBombenzeit, Radius: s.Feuer,
		Frei: s.Nr, Phase: rand.Float64() * 6,
	})
	return true
}

func (g *BomberGame) vielleichtPowerup(x, y int) {
	if rand.Float64() > 0.34 {
		return
	}
	arten := []string{"feuer", "feuer", "bombe", "bombe", "tempo"}
	g.Powerups = append(g.Powerups, BomberPowerup{
		X: x, Y: y, Art: arten[rand.Intn(len(arten))], Phase: rand.Float64() * 6,
	})
}

// explodiere lässt eine Bombe hochgehen; Kettenreaktionen laufen sofort mit.
func (g *BomberGame) explodiere(b *BomberBombe) {
	zellen := [][2]int{{b.X, b.Y}}
	for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		for i := 1; i <= b.Radius; i++ {
			tx, ty := b.X+d[0]*i, b.Y+d[1]*i
			k := g.kachel(tx, ty)
			if k == 1 {
				break
			}
			zellen = append(zellen, [2]int{tx, ty})
			if k == 2 {
				g.Raster[ty*bomberSpalten+tx] = 0
				g.vielleichtPowerup(tx, ty)
				break
			}
		}
	}
	for _, z := range zellen {
		g.Flammen = append(g.Flammen, BomberFlamme{
			X: z[0], Y: z[1], Rest: bomberFlammenzeit, Phase: rand.Float64() * 6,
		})
	}
	// Kette: eine Flamme zündet andere Bomben sofort.
	for _, z := range zellen {
		for _, andere := range g.Bomben {
			if andere == b || andere.Rest <= 0 {
				continue
			}
			if andere.X == z[0] && andere.Y == z[1] {
				andere.Rest = 0
				g.explodiere(andere)
			}
		}
	}
}

// bomberSchritt rechnet ein Sechzigstel einer Sekunde weiter.
func (g *BomberGame) schritt(dt float64) {
	if g.Status != "spiel" {
		return
	}
	if g.Vorlauf > 0 {
		g.Vorlauf -= dt
		if g.Vorlauf <= 0 {
			g.Vorlauf = 0
			g.Note = "Los!"
		}
	}

	for i := range g.Spieler {
		g.bewege(&g.Spieler[i], dt)
		if !g.Spieler[i].Leben {
			g.Spieler[i].TotSeit += dt
		}
	}

	// Bomben zünden
	var offen []*BomberBombe
	for _, b := range g.Bomben {
		if b.Rest <= 0 {
			continue // ist schon in der Kette hochgegangen
		}
		b.Rest -= dt
		if b.Rest <= 0 {
			g.explodiere(b)
			continue
		}
		offen = append(offen, b)
	}
	g.Bomben = offen

	// Flammen brennen ab
	var brennt []BomberFlamme
	for _, f := range g.Flammen {
		f.Rest -= dt
		if f.Rest > 0 {
			brennt = append(brennt, f)
		}
	}
	g.Flammen = brennt

	// Wer in einer Flamme steht, ist raus.
	for i := range g.Spieler {
		s := &g.Spieler[i]
		if !s.Leben {
			continue
		}
		tx, ty := int(math.Floor(s.X)), int(math.Floor(s.Y))
		for _, f := range g.Flammen {
			if f.X == tx && f.Y == ty {
				s.Leben = false
				s.TotSeit = 0
				break
			}
		}
	}

	g.rundePruefen()
}

func (g *BomberGame) rundePruefen() {
	lebend := make([]*BomberSpieler, 0, len(g.Spieler))
	for i := range g.Spieler {
		if g.Spieler[i].Leben {
			lebend = append(lebend, &g.Spieler[i])
		}
	}
	if len(lebend) > 1 {
		return
	}
	g.Status = "runde"
	if len(lebend) == 1 {
		gewinner := lebend[0]
		g.Gewinner = gewinner.Nr
		gewinner.Sieg++
		g.Meldung = gewinner.Name + " gewinnt Runde " + strconv.Itoa(g.Runde) + " und steht bei " +
			strconv.Itoa(gewinner.Sieg) + " Siegen."
		if gewinner.Sieg >= bomberZiel {
			g.Status = "ende"
			g.Sieger = gewinner.Name
			g.Meldung = gewinner.Name + " gewinnt das Match mit " + strconv.Itoa(gewinner.Sieg) + " Rundensiegen."
			g.Note = "Neue Partie startet mit einem Klick."
			return
		}
	} else {
		g.Gewinner = 0
		g.Meldung = "Alle Bomber sind gleichzeitig hochgegangen. Runde " + strconv.Itoa(g.Runde) + " wird wiederholt."
	}
	g.Note = "Weiter mit der nächsten Runde – wer zuerst drückt, startet sie."
}

// bomberWeiter lässt die Partie bis jetzt laufen.
func bomberWeiter(g *BomberGame) {
	if g == nil || g.Status != "spiel" {
		return
	}
	jetzt := time.Now()
	dt := jetzt.Sub(g.Zuletzt).Seconds()
	if dt < 0 {
		dt = 0
	}
	if dt > bomberMaxVorsprung {
		dt = bomberMaxVorsprung
	}
	g.Zuletzt = jetzt
	g.Rest += dt
	schritte := 0
	for g.Rest >= bomberTakt && g.Status == "spiel" && schritte < bomberMaxSchritte {
		g.Rest -= bomberTakt
		schritte++
		g.schritt(bomberTakt)
	}
	if schritte >= bomberMaxSchritte {
		g.Rest = 0
	}
}

// ---------------------------------------------------------------- Sicht

func bomberAnsicht(g *BomberGame, id string) *BomberAnsicht {
	if g == nil {
		return nil
	}
	a := &BomberAnsicht{
		ID:       g.ID,
		HostID:   g.HostID,
		Status:   g.Status,
		Spalten:  bomberSpalten,
		Zeilen:   bomberZeilen,
		Runde:    g.Runde,
		Ziel:     bomberZiel,
		Vorlauf:  g.Vorlauf,
		Raster:   tronRLE(g.Raster),
		Gewinner: g.Gewinner,
		Meldung:  g.Meldung,
		Note:     g.Note,
		Sieger:   g.Sieger,
		Du:       bomberSitz(g, id),
		Tick:     time.Now().UnixMilli(),
	}
	// Leere Listen als [] statt null – die Seite zählt sie direkt durch.
	a.Spieler = []BomberSpielerSicht{}
	a.Bomben = []BomberBombeSicht{}
	a.Flammen = []BomberFlammeSicht{}
	a.Powerups = []BomberPowerupSicht{}
	for _, p := range g.Spieler {
		a.Spieler = append(a.Spieler, BomberSpielerSicht{
			ID: p.ID, Name: p.Name, Ready: p.Ready, Nr: p.Nr,
			X: p.X, Y: p.Y, DX: p.DX, DY: p.DY, BlickDX: p.BlickDX, BlickDY: p.BlickDY,
			BombMax: p.BombMax, Feuer: p.Feuer, Tempo: p.Tempo,
			Leben: p.Leben, Sieg: p.Sieg, LaufT: p.LaufT, TotSeit: p.TotSeit,
		})
	}
	for _, b := range g.Bomben {
		a.Bomben = append(a.Bomben, BomberBombeSicht{
			X: b.X, Y: b.Y, Nr: b.Nr, Rest: b.Rest, Radius: b.Radius, Phase: b.Phase,
		})
	}
	for _, f := range g.Flammen {
		a.Flammen = append(a.Flammen, BomberFlammeSicht{X: f.X, Y: f.Y, Rest: f.Rest, Phase: f.Phase})
	}
	for _, p := range g.Powerups {
		a.Powerups = append(a.Powerups, BomberPowerupSicht{X: p.X, Y: p.Y, Art: p.Art, Phase: p.Phase})
	}
	return a
}

// ---------------------------------------------------------------- Endpunkt

func bomberHandler(w http.ResponseWriter, r *http.Request) {
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
		bomberWeiter(room.Bomberman)
		bomberBotTreiben(room, time.Now())
		game := bomberAnsicht(room.Bomberman, id)
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
		DX    int    `json:"dx"`
		DY    int    `json:"dy"`
		Stufe string `json:"stufe"`
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

	zeit := time.Now().Format("15:04:05")

	switch action {
	case "bot":
		room.botNamenNeu()
		if room.Bomberman != nil && room.Bomberman.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Partie Bomberman.", http.StatusConflict)
			return
		}
		stufe := botStufe(input.Stufe)
		game := bomberNeu(randomCode(), input.ID, []BomberSpieler{
			bomberNeuerSpieler(1, input.ID, player.Name),
			bomberNeuerSpieler(2, botID(1), room.botName(1)),
		})
		game.Bots, game.BotStufe = []string{botID(1)}, stufe
		for i := range game.Spieler {
			game.Spieler[i].Ready = true
		}
		bomberStarten(game)
		game.Note = player.Name + " spielt gegen " + room.botName(1) +
			" (" + botStufeText(stufe) + ")."
		game.BotFrist = time.Now().Add(botWartezeit(stufe))
		room.Bomberman = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Bomberman gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "create":
		if room.Bomberman != nil && room.Bomberman.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Partie Bomberman.", http.StatusConflict)
			return
		}
		room.Bomberman = bomberNeu(randomCode(), input.ID,
			[]BomberSpieler{bomberNeuerSpieler(1, input.ID, player.Name)})
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Partie Bomberman – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Bomberman
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if bomberSitz(game, input.ID) >= 0 {
			break
		}
		if len(game.Spieler) >= bomberMaxSpieler {
			mu.Unlock()
			http.Error(w, "Bomberman ist hier ein Duell – der Tisch ist voll.", http.StatusConflict)
			return
		}
		game.Spieler = append(game.Spieler, bomberNeuerSpieler(2, input.ID, player.Name))
		game.Note = player.Name + " sitzt mit am Tisch. Beide drücken BEREIT."

	case "ready":
		game := room.Bomberman
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := bomberSitz(game, input.ID)
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
		game := room.Bomberman
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Partie auf den Start.", http.StatusConflict)
			return
		}
		if len(game.Spieler) < bomberMaxSpieler {
			mu.Unlock()
			http.Error(w, "Bomberman braucht zwei Spieler.", http.StatusConflict)
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
		bomberStarten(game)

	case "richtung", "bombe", "weiter", "reset":
		status, meldung := bomberAktion(room, input.ID, action, input.DX, input.DY)
		if meldung != "" {
			mu.Unlock()
			http.Error(w, meldung, status)
			return
		}

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	bomberWeiter(room.Bomberman)
	result := bomberAnsicht(room.Bomberman, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

/* ---------------------------------------------------------------------------
   Gemeinsame Prüfung für Menschen und Bots
   ------------------------------------------------------------------------- */

// bomberStarten setzt die Partie auf Runde 1 zurück und gibt sie frei.
func bomberStarten(g *BomberGame) {
	for i := range g.Spieler {
		g.Spieler[i].Sieg = 0
	}
	g.Runde = 1
	g.Sieger = ""
	bomberNeueRunde(g)
	g.Status = "spiel"
	g.Note = "Achtung, gleich geht's los …"
}

// bomberAktion führt einen Spielzug aus – für Menschen und Bots dieselbe
// Prüfung. create, join, ready und start bleiben im Handler, weil sie Namen
// und Meldungen des Raums brauchen.
func bomberAktion(room *Room, id, aktion string, dx, dy int) (int, string) {
	g := room.Bomberman

	switch aktion {
	case "richtung":
		if g == nil || g.Status != "spiel" {
			return http.StatusConflict, "Gerade läuft keine Runde."
		}
		sitz := bomberSitz(g, id)
		if sitz < 0 {
			return http.StatusForbidden, "Du sitzt nicht an diesem Tisch."
		}
		if dx < -1 || dx > 1 || dy < -1 || dy > 1 || (dx != 0 && dy != 0) {
			return http.StatusBadRequest, "Ungültige Richtung."
		}
		g.Spieler[sitz].DX = dx
		g.Spieler[sitz].DY = dy

	case "bombe":
		if g == nil || g.Status != "spiel" {
			return http.StatusConflict, "Gerade läuft keine Runde."
		}
		sitz := bomberSitz(g, id)
		if sitz < 0 {
			return http.StatusForbidden, "Du sitzt nicht an diesem Tisch."
		}
		if !g.legeBombe(&g.Spieler[sitz]) {
			return http.StatusConflict, "Jetzt geht keine Bombe."
		}

	case "weiter":
		if g == nil || g.Status != "runde" {
			return http.StatusConflict, "Die Runde läuft noch."
		}
		if bomberSitz(g, id) < 0 {
			return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
		}
		if g.Gewinner != 0 {
			g.Runde++
		}
		bomberNeueRunde(g)
		g.Status = "spiel"
		g.Note = "Achtung, gleich geht's los …"
		g.BotFrist = time.Time{}

	case "reset":
		if g == nil || g.Status != "ende" {
			return http.StatusConflict, "Eine neue Partie geht erst nach dem Ende."
		}
		if bomberSitz(g, id) < 0 {
			return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
		}
		frisch := bomberNeu(g.ID, g.HostID, []BomberSpieler{})
		frisch.Bots, frisch.BotStufe = g.Bots, g.BotStufe
		frisch.Note = "Neue Partie – erst BEREIT drücken."
		for _, p := range g.Spieler {
			frisch.Spieler = append(frisch.Spieler, bomberNeuerSpieler(p.Nr, p.ID, p.Name))
		}
		room.Bomberman = frisch

	default:
		return http.StatusBadRequest, "Unbekannte Aktion."
	}
	return 0, ""
}
