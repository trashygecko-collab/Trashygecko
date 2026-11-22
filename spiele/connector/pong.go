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

// Pong für zwei Spieler über den Connector: Der Server führt Ball und Schläger
// selbst, die Geräte schicken nur ihre Schlägerposition und holen den Stand ab.
// So sehen beide denselben Ball, auch wenn ein Handy kurz hängt – und keiner
// kann am anderen vorbei mogeln.
//
// Das Feld ist hochkant wie auf der Spielseite: Spieler 0 spielt unten,
// Spieler 1 oben. Jeder sieht sich selbst unten (die Spielseite spiegelt).

const (
	pongBreite          = 400.0
	pongHoehe           = 600.0
	pongBallR           = 9.0
	pongSchlaegerBreite = 78.0
	pongSchlaegerHoehe  = 13.0
	pongFeldOben        = 24.0
	pongFeldUnten       = pongHoehe - 24.0
	pongSiegespunkte    = 7
	pongTempoStart      = 324.0 // Pixel je Sekunde
	pongTempoMax        = 690.0
	pongMaxSpieler      = 2
	pongPauseSekunden   = 0.9
	pongTakt            = 1.0 / 120.0 // kleine Schritte, damit nichts durchrutscht
	pongMaxVorsprung    = 2.0         // höchstens zwei Sekunden auf einmal nachrechnen
)

type PongSpieler struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Ready bool    `json:"ready"`
	X     float64 `json:"-"`
}

type PongGame struct {
	ID       string        `json:"id"`
	HostID   string        `json:"hostId"`
	Status   string        `json:"status"` // invited, spiel, ende
	Spieler  []PongSpieler `json:"spieler"`
	BallX    float64       `json:"-"`
	BallY    float64       `json:"-"`
	BallVX   float64       `json:"-"`
	BallVY   float64       `json:"-"`
	Punkte   [2]int        `json:"-"`
	Rallye   int           `json:"-"`
	Pause    float64       `json:"-"`
	Sieger   string        `json:"sieger,omitempty"`
	Meldung  string        `json:"meldung"`
	Note     string        `json:"note"`
	Zuletzt  time.Time     `json:"-"`
	BotID    string        `json:"botId,omitempty"`
	BotStufe string        `json:"-"`
	// BotFehler und BotRallye halten den Zielfehler je Ballwechsel fest,
	// damit der Schläger nicht bei jedem Bild neu zittert.
	BotFehler float64 `json:"-"`
	BotRallye int     `json:"-"`
}

type PongSpielerSicht struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Ready  bool   `json:"ready"`
	Punkte int    `json:"punkte"`
}

type PongBall struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	VX float64 `json:"vx"`
	VY float64 `json:"vy"`
}

type PongAnsicht struct {
	ID        string             `json:"id"`
	HostID    string             `json:"hostId"`
	Status    string             `json:"status"`
	Ball      PongBall           `json:"ball"`
	Schlaeger []float64          `json:"schlaeger"` // x-Position je Spieler
	Punkte    []int              `json:"punkte"`
	Rallye    int                `json:"rallye"`
	Pause     float64            `json:"pause"`
	Spieler   []PongSpielerSicht `json:"spieler"`
	Du        int                `json:"du"` // -1 = Zuschauer
	Sieger    string             `json:"sieger,omitempty"`
	Meldung   string             `json:"meldung"`
	Note      string             `json:"note"`
	Tick      int64              `json:"tick"` // Serverzeit in ms (nur fürs Gefühl)
}

// ---------------------------------------------------------------- Aufbau

func pongNeuerSpieler(id, name string) PongSpieler {
	return PongSpieler{ID: id, Name: name, X: pongBreite / 2}
}

func pongSitz(g *PongGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func pongSchlaegerY(sitz int) float64 {
	if sitz == 0 {
		return pongFeldUnten - pongSchlaegerHoehe
	}
	return pongFeldOben
}

func pongMitte(g *PongGame) {
	g.BallX = pongBreite/2 + (rand.Float64()*80 - 40)
	g.BallY = pongHoehe / 2
	g.BallVX = 0
	g.BallVY = 0
	g.Rallye = 0
}

func pongAufschlag(g *PongGame, richtung float64) {
	winkel := rand.Float64() - 0.5
	g.BallX = pongBreite/2 + (rand.Float64()*80 - 40)
	g.BallY = pongHoehe / 2
	g.BallVX = math.Sin(winkel) * pongTempoStart
	g.BallVY = math.Cos(winkel) * pongTempoStart * richtung
	g.Rallye = 0
	g.Pause = 0
}

func pongNeu(ID, hostID string, spieler []PongSpieler) *PongGame {
	g := &PongGame{
		ID:      ID,
		HostID:  hostID,
		Status:  "invited",
		Spieler: spieler,
		Zuletzt: time.Now(),
		Note:    "Tisch eröffnet – gesucht wird ein zweiter Spieler.",
	}
	for i := range g.Spieler {
		g.Spieler[i].X = pongBreite / 2
	}
	pongMitte(g)
	return g
}

// ---------------------------------------------------------------- Physik

// pongTempo liefert die Geschwindigkeit des Balls in Pixel je Sekunde.
func pongTempo(g *PongGame) float64 {
	return math.Hypot(g.BallVX, g.BallVY)
}

// pongPunkt zählt einen Punkt und legt eine kurze Pause ein. Danach geht der
// Aufschlag zum Verlierer des Ballwechsels.
func pongPunkt(g *PongGame, wer int) {
	g.Punkte[wer]++
	g.Rallye = 0
	pongMitte(g)
	if g.Punkte[0] >= pongSiegespunkte || g.Punkte[1] >= pongSiegespunkte {
		g.Status = "ende"
		g.Sieger = g.Spieler[wer].Name
		g.Meldung = g.Spieler[wer].Name + " gewinnt mit " +
			strconv.Itoa(g.Punkte[wer]) + " : " + strconv.Itoa(g.Punkte[1-wer]) + "."
		g.Note = "Neue Partie startet mit einem Klick."
		return
	}
	g.Pause = pongPauseSekunden
	if wer == 0 {
		g.BallVY = -pongTempoStart // Aufschlag nach oben (Verlierer oben)
	} else {
		g.BallVY = pongTempoStart
	}
	g.BallVX = 0
}

// pongSchlaegerTreffer spiegelt den Ball am Schläger und macht ihn etwas
// schneller. Wo er trifft, bestimmt den Winkel.
func pongSchlaegerTreffer(g *PongGame, sitz int) {
	halb := pongSchlaegerBreite / 2
	versatz := (g.BallX - g.Spieler[sitz].X) / halb
	if versatz > 1 {
		versatz = 1
	}
	if versatz < -1 {
		versatz = -1
	}
	tempo := math.Min(pongTempoMax, pongTempo(g)*1.045)
	g.BallVX = versatz * tempo * 0.72
	rest := tempo*tempo - g.BallVX*g.BallVX
	if rest < 1 {
		rest = 1
	}
	if sitz == 0 {
		g.BallVY = -math.Sqrt(rest)
		g.BallY = pongSchlaegerY(0) - pongBallR
	} else {
		g.BallVY = math.Sqrt(rest)
		g.BallY = pongSchlaegerY(1) + pongSchlaegerHoehe + pongBallR
	}
	g.Rallye++
	if g.Rallye > 0 && g.Rallye%10 == 0 {
		g.Note = "Rallye " + strconv.Itoa(g.Rallye) + " – der Ball wird schneller."
	}
}

// pongPhysik rechnet einen kleinen Schritt weiter. dt steht in Sekunden.
func pongPhysik(g *PongGame, dt float64) {
	if g.Status != "spiel" {
		return
	}
	if g.Pause > 0 {
		g.Pause -= dt
		if g.Pause <= 0 {
			g.Pause = 0
			richtung := 1.0
			if g.BallVY < 0 {
				richtung = -1
			}
			if g.BallVY == 0 {
				richtung = -1
			}
			pongAufschlag(g, richtung)
			g.Note = "Aufschlag."
		}
		return
	}
	g.BallX += g.BallVX * dt
	g.BallY += g.BallVY * dt

	// Seitenwände
	if g.BallX-pongBallR < 0 {
		g.BallX = pongBallR
		g.BallVX = math.Abs(g.BallVX)
	}
	if g.BallX+pongBallR > pongBreite {
		g.BallX = pongBreite - pongBallR
		g.BallVX = -math.Abs(g.BallVX)
	}

	// Schläger unten (Spieler 0)
	if g.BallVY > 0 {
		oben := pongSchlaegerY(0)
		if g.BallY+pongBallR >= oben && g.BallY+pongBallR <= oben+pongSchlaegerHoehe+12 {
			if math.Abs(g.BallX-g.Spieler[0].X) < pongSchlaegerBreite/2+pongBallR*0.6 {
				pongSchlaegerTreffer(g, 0)
			}
		}
	}

	// Schläger oben (Spieler 1)
	if g.BallVY < 0 {
		unten := pongSchlaegerY(1) + pongSchlaegerHoehe
		if g.BallY-pongBallR <= unten && g.BallY-pongBallR >= unten-12-pongSchlaegerHoehe {
			if math.Abs(g.BallX-g.Spieler[1].X) < pongSchlaegerBreite/2+pongBallR*0.6 {
				pongSchlaegerTreffer(g, 1)
			}
		}
	}

	// Punkt: unten vorbei heißt Punkt für oben und umgekehrt.
	if g.BallY-pongBallR > pongHoehe {
		pongPunkt(g, 1)
		return
	}
	if g.BallY+pongBallR < 0 {
		pongPunkt(g, 0)
	}
}

// pongStart setzt eine Partie auf Anfang: Schläger mittig, kein Punkt, Aufschlag.
func pongStart(g *PongGame) {
	for i := range g.Spieler {
		g.Spieler[i].X = pongBreite / 2
	}
	g.Punkte = [2]int{}
	g.Sieger = ""
	g.Meldung = ""
	g.Status = "spiel"
	g.Note = "Los! Finger in der eigenen Hälfte ziehen, der Schläger folgt."
	g.Zuletzt = time.Now()
	g.BotRallye = -1
	pongAufschlag(g, -1)
}

// pongIDs listet die Plätze einer Partie – für die Platzprüfung.
func pongIDs(g *PongGame) []string {
	ids := make([]string, 0, len(g.Spieler))
	for _, p := range g.Spieler {
		ids = append(ids, p.ID)
	}
	return ids
}

// pongWeiter lässt die Partie bis jetzt laufen. Ohne Abfrage steht sie still –
// das ist gewollt: niemand schaut zu, also muss auch nichts rechnen.
func pongWeiter(g *PongGame) {
	if g == nil || g.Status != "spiel" {
		return
	}
	jetzt := time.Now()
	dt := jetzt.Sub(g.Zuletzt).Seconds()
	if dt < 0 {
		dt = 0
	}
	if dt > pongMaxVorsprung {
		dt = pongMaxVorsprung
	}
	g.Zuletzt = jetzt
	pongBotBewegen(g, dt)
	schritte := int(dt / pongTakt)
	if schritte > 600 {
		schritte = 600
	}
	for i := 0; i < schritte; i++ {
		pongPhysik(g, pongTakt)
		if g.Status != "spiel" {
			return
		}
	}
}

// ---------------------------------------------------------------- Sicht

func pongAnsicht(g *PongGame, id string) *PongAnsicht {
	if g == nil {
		return nil
	}
	a := &PongAnsicht{
		ID:      g.ID,
		HostID:  g.HostID,
		Status:  g.Status,
		Ball:    PongBall{X: g.BallX, Y: g.BallY, VX: g.BallVX, VY: g.BallVY},
		Rallye:  g.Rallye,
		Pause:   g.Pause,
		Du:      pongSitz(g, id),
		Sieger:  g.Sieger,
		Meldung: g.Meldung,
		Note:    g.Note,
		Tick:    time.Now().UnixMilli(),
	}
	a.Punkte = []int{g.Punkte[0], g.Punkte[1]}
	for i, p := range g.Spieler {
		a.Schlaeger = append(a.Schlaeger, p.X)
		a.Spieler = append(a.Spieler, PongSpielerSicht{
			ID: p.ID, Name: p.Name, Ready: p.Ready, Punkte: g.Punkte[i],
		})
	}
	return a
}

// ---------------------------------------------------------------- Endpunkt

func pongHandler(w http.ResponseWriter, r *http.Request) {
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
		pongWeiter(room.Pong)
		game := pongAnsicht(room.Pong, id)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID    string  `json:"id"`
		X     float64 `json:"x"`
		Stufe string  `json:"stufe"`
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
		if room.Pong != nil && !room.gameSlotFree(room.Pong.Status, pongIDs(room.Pong)...) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Partie Pong.", http.StatusConflict)
			return
		}
		stufe := botStufe(input.Stufe)
		g := pongNeu(randomCode(), input.ID, []PongSpieler{
			pongNeuerSpieler(input.ID, player.Name),
			pongNeuerSpieler(botID(1), room.botName(1)),
		})
		g.Spieler[0].Ready = true
		g.Spieler[1].Ready = true
		g.BotID, g.BotStufe = botID(1), stufe
		pongStart(g)
		g.Note = player.Name + " spielt gegen " + room.botName(1) + " (" + botStufeText(stufe) + ")."
		room.Pong = g
	case "create":
		if room.Pong != nil && room.Pong.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Partie Pong.", http.StatusConflict)
			return
		}
		room.Pong = pongNeu(randomCode(), input.ID, []PongSpieler{pongNeuerSpieler(input.ID, player.Name)})
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Partie Pong – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Pong
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if pongSitz(game, input.ID) >= 0 {
			break
		}
		if len(game.Spieler) >= pongMaxSpieler {
			mu.Unlock()
			http.Error(w, "Pong ist ein Duell – der Tisch ist voll.", http.StatusConflict)
			return
		}
		game.Spieler = append(game.Spieler, pongNeuerSpieler(input.ID, player.Name))
		game.Note = player.Name + " sitzt mit am Tisch. Beide drücken BEREIT."

	case "ready":
		game := room.Pong
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := pongSitz(game, input.ID)
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
		game := room.Pong
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Partie auf den Start.", http.StatusConflict)
			return
		}
		if len(game.Spieler) < pongMaxSpieler {
			mu.Unlock()
			http.Error(w, "Pong braucht zwei Spieler.", http.StatusConflict)
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
		pongStart(game)

	case "schlaeger":
		game := room.Pong
		if game == nil {
			mu.Unlock()
			http.Error(w, "Es läuft keine Partie Pong.", http.StatusConflict)
			return
		}
		sitz := pongSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		x := input.X
		halb := pongSchlaegerBreite / 2
		if math.IsNaN(x) {
			x = pongBreite / 2
		}
		if x < halb {
			x = halb
		}
		if x > pongBreite-halb {
			x = pongBreite - halb
		}
		game.Spieler[sitz].X = x
		pongWeiter(game)

	case "reset":
		game := room.Pong
		if game == nil || game.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Eine neue Partie geht erst nach dem Ende.", http.StatusConflict)
			return
		}
		if pongSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		frisch := pongNeu(game.ID, game.HostID, []PongSpieler{})
		frisch.BotID, frisch.BotStufe = game.BotID, game.BotStufe
		frisch.Note = "Neue Partie – erst BEREIT drücken."
		for _, p := range game.Spieler {
			frisch.Spieler = append(frisch.Spieler, pongNeuerSpieler(p.ID, p.Name))
		}
		room.Pong = frisch

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	pongWeiter(room.Pong)
	result := pongAnsicht(room.Pong, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
