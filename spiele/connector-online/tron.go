package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Tron (Lichtrennen) für zwei Spieler über den Connector: Der Server führt das
// Raster, die Spuren und die Köpfe. Die Geräte schicken nur ihre Richtung und
// holen den Stand ab – so laufen beide exakt dieselbe Partie, auch wenn ein
// Handy kurz hängt.
//
// Feld: 32 × 44 Felder hochkant (wie das Handy). Spieler 1 startet links und
// fährt nach rechts, Spieler 2 rechts und nach links. Drei Rundensiege
// gewinnen die Partie; fahren beide Köpfe ineinander, wird die Runde
// wiederholt.

const (
	tronSpalten      = 32
	tronZeilen       = 44
	tronZiel         = 3
	tronTakt         = 0.12 // Sekunden je Feld (gut acht Felder je Sekunde)
	tronMaxSpieler   = 2
	tronMaxVorsprung = 3.0
	tronMaxSchritte  = 40
)

type TronSpieler struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Ready    bool   `json:"ready"`
	X        int    `json:"-"`
	Y        int    `json:"-"`
	DX       int    `json:"-"`
	DY       int    `json:"-"`
	WunschDX int    `json:"-"`
	WunschDY int    `json:"-"`
	Tot      bool   `json:"-"`
	Sieg     int    `json:"-"`
	Spur     uint8  `json:"-"`
}

type TronGame struct {
	ID       string        `json:"id"`
	HostID   string        `json:"hostId"`
	Status   string        `json:"status"` // invited, spiel, runde, ende
	Spieler  []TronSpieler `json:"spieler"`
	Raster   []uint8       `json:"-"`
	Runde    int           `json:"runde"`
	Gewinner int           `json:"-"`
	Grund    string        `json:"-"`
	Meldung  string        `json:"meldung"`
	Note     string        `json:"note"`
	Sieger   string        `json:"sieger,omitempty"`
	Zuletzt  time.Time     `json:"-"`
	Rest     float64       `json:"-"`
	BotID    string        `json:"botId,omitempty"`
	BotStufe string        `json:"-"`
	BotFrist time.Time     `json:"-"`
}

type TronSpielerSicht struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	DX    int    `json:"dx"`
	DY    int    `json:"dy"`
	Tot   bool   `json:"tot"`
	Sieg  int    `json:"sieg"`
}

type TronAnsicht struct {
	ID       string             `json:"id"`
	HostID   string             `json:"hostId"`
	Status   string             `json:"status"`
	Spalten  int                `json:"spalten"`
	Zeilen   int                `json:"zeilen"`
	Runde    int                `json:"runde"`
	Ziel     int                `json:"ziel"`
	Gewinner int                `json:"gewinner"` // -1 keiner, 0 unentschieden, 1/2
	Raster   string             `json:"raster"`   // Lauflängen, z. B. "0x120,1x4"
	Spieler  []TronSpielerSicht `json:"spieler"`
	Du       int                `json:"du"`
	SiegZiel int                `json:"-"`
	Meldung  string             `json:"meldung"`
	Note     string             `json:"note"`
	Sieger   string             `json:"sieger,omitempty"`
	Tick     int64              `json:"tick"`
}

// ---------------------------------------------------------------- Aufbau

func tronNeuerSpieler(nr int, id, name string) TronSpieler {
	mitte := tronZeilen/2 - 5
	if nr == 2 {
		mitte = tronZeilen/2 + 5
	}
	s := TronSpieler{
		ID:   id,
		Name: name,
		Y:    mitte,
		Spur: uint8(nr),
	}
	if nr == 1 {
		s.X = 3
		s.DX = 1
	} else {
		s.X = tronSpalten - 4
		s.DX = -1
	}
	return s
}

func tronSitz(g *TronGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// tronIDs listet die Plätze einer Partie – für die Platzprüfung.
func tronIDs(g *TronGame) []string {
	ids := make([]string, 0, len(g.Spieler))
	for _, p := range g.Spieler {
		ids = append(ids, p.ID)
	}
	return ids
}

// tronStart setzt das Match auf Anfang (Startknopf und Bot).
func tronStart(g *TronGame) {
	for i := range g.Spieler {
		g.Spieler[i].Sieg = 0
	}
	g.Runde = 1
	g.Sieger = ""
	g.Meldung = ""
	tronNeueRunde(g)
	g.Status = "spiel"
	g.Note = "Los! Nur rechtswinklige Kurven – umdrehen geht nicht."
}

// tronNaechsteRunde startet die nächste Runde; nach einem Unentschieden bleibt
// die Rundenzahl stehen.
func tronNaechsteRunde(g *TronGame) {
	if g.Gewinner != 0 {
		g.Runde++
	}
	tronNeueRunde(g)
	g.Status = "spiel"
	g.Meldung = ""
	g.Note = "Runde " + strconv.Itoa(g.Runde) + " läuft."
}

func tronNeueRunde(g *TronGame) {
	g.Raster = make([]uint8, tronSpalten*tronZeilen)
	for i := range g.Spieler {
		var nr int
		if g.Spieler[i].Spur == 2 {
			nr = 2
		} else {
			nr = 1
		}
		sieg := g.Spieler[i].Sieg
		ready := g.Spieler[i].Ready
		g.Spieler[i] = tronNeuerSpieler(nr, g.Spieler[i].ID, g.Spieler[i].Name)
		g.Spieler[i].Sieg = sieg
		g.Spieler[i].Ready = ready
		g.Raster[g.Spieler[i].Y*tronSpalten+g.Spieler[i].X] = g.Spieler[i].Spur
	}
	g.Gewinner = -1
	g.Grund = ""
	g.Rest = 0
	g.Zuletzt = time.Now()
}

func tronNeu(ID, hostID string, spieler []TronSpieler) *TronGame {
	g := &TronGame{
		ID:       ID,
		HostID:   hostID,
		Status:   "invited",
		Spieler:  spieler,
		Runde:    1,
		Gewinner: -1,
		Zuletzt:  time.Now(),
		Note:     "Tisch eröffnet – gesucht wird ein zweiter Spieler.",
	}
	g.Raster = make([]uint8, tronSpalten*tronZeilen)
	return g
}

// ---------------------------------------------------------------- Regeln

func tronFeldFrei(x, y int) bool {
	return x >= 0 && x < tronSpalten && y >= 0 && y < tronZeilen
}

// tronRichtung übersetzt "hoch", "runter", "links", "rechts".
func tronRichtung(name string) (int, int, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "hoch":
		return 0, -1, true
	case "runter":
		return 0, 1, true
	case "links":
		return -1, 0, true
	case "rechts":
		return 1, 0, true
	}
	return 0, 0, false
}

// tronLenken merkt die nächste Richtung. Umdrehen und die aktuelle Richtung
// sind verboten – wie im lokalen Spiel.
func tronLenken(s *TronSpieler, dx, dy int) bool {
	if s.Tot {
		return false
	}
	if dx == -s.DX && dy == -s.DY {
		return false
	}
	if dx == s.DX && dy == s.DY {
		return false
	}
	s.WunschDX = dx
	s.WunschDY = dy
	return true
}

// tronSchritt ist ein Feld weit: erst lenken, dann Köpfe setzen, Kollisionen
// prüfen und Spuren ziehen.
func tronSchritt(g *TronGame) {
	for i := range g.Spieler {
		s := &g.Spieler[i]
		if s.WunschDX != 0 || s.WunschDY != 0 {
			s.DX, s.DY = s.WunschDX, s.WunschDY
			s.WunschDX, s.WunschDY = 0, 0
		}
	}

	nx := make([]int, len(g.Spieler))
	ny := make([]int, len(g.Spieler))
	for i := range g.Spieler {
		nx[i] = g.Spieler[i].X + g.Spieler[i].DX
		ny[i] = g.Spieler[i].Y + g.Spieler[i].DY
	}

	stirbt := make([]bool, len(g.Spieler))
	for i := range g.Spieler {
		if g.Spieler[i].Tot {
			stirbt[i] = true
			continue
		}
		wand := !tronFeldFrei(nx[i], ny[i])
		spur := false
		if !wand {
			spur = g.Raster[ny[i]*tronSpalten+nx[i]] != 0
		}
		stirbt[i] = wand || spur
	}
	if len(g.Spieler) == 2 && !stirbt[0] && !stirbt[1] {
		// Kopf an Kopf oder Kreuztausch: beide sterben.
		if nx[0] == nx[1] && ny[0] == ny[1] {
			stirbt[0], stirbt[1] = true, true
		} else if nx[0] == g.Spieler[1].X && ny[0] == g.Spieler[1].Y &&
			nx[1] == g.Spieler[0].X && ny[1] == g.Spieler[0].Y {
			stirbt[0], stirbt[1] = true, true
		}
	}

	for i := range g.Spieler {
		if stirbt[i] {
			g.Spieler[i].Tot = true
			continue
		}
		g.Spieler[i].X = nx[i]
		g.Spieler[i].Y = ny[i]
		g.Raster[ny[i]*tronSpalten+nx[i]] = g.Spieler[i].Spur
	}

	tronRundePruefen(g)
}

func tronRundePruefen(g *TronGame) {
	tot := 0
	for _, s := range g.Spieler {
		if s.Tot {
			tot++
		}
	}
	// Die Runde endet, sobald ein Renner gestorben ist – wie im lokalen
	// Spiel. Sind beide gestorben (Kopf an Kopf), ist die Runde unentschieden.
	if tot == 0 {
		return
	}
	gewinner := 0
	if len(g.Spieler) == 2 {
		switch {
		case g.Spieler[0].Tot && !g.Spieler[1].Tot:
			gewinner = 2
		case g.Spieler[1].Tot && !g.Spieler[0].Tot:
			gewinner = 1
		}
	}
	g.Gewinner = gewinner
	g.Status = "runde"
	if gewinner == 0 {
		g.Grund = "Kopf an Kopf – beide Renner sind gestorben."
		g.Meldung = "Runde " + strconv.Itoa(g.Runde) + " unentschieden und wird wiederholt."
	} else {
		g.Spieler[gewinner-1].Sieg++
		g.Grund = g.Spieler[gewinner-1].Name + " gewinnt die Runde."
		g.Meldung = g.Spieler[gewinner-1].Name + " gewinnt Runde " + strconv.Itoa(g.Runde) + "."
	}
	if g.Spieler[0].Sieg >= tronZiel || g.Spieler[1].Sieg >= tronZiel {
		g.Status = "ende"
		wer := 0
		if g.Spieler[1].Sieg > g.Spieler[0].Sieg {
			wer = 1
		}
		g.Sieger = g.Spieler[wer].Name
		g.Meldung = g.Spieler[wer].Name + " gewinnt das Match mit " +
			strconv.Itoa(g.Spieler[wer].Sieg) + " zu " + strconv.Itoa(g.Spieler[1-wer].Sieg) + " Runden."
		g.Note = "Neue Partie startet mit einem Klick."
		return
	}
	g.Note = "Weiter mit der nächsten Runde – wer zuerst drückt, startet sie."
}

// tronWeiter lässt die Partie bis jetzt laufen. Ohne Abfrage steht sie still.
func tronWeiter(g *TronGame) {
	if g == nil || g.Status != "spiel" {
		return
	}
	jetzt := time.Now()
	dt := jetzt.Sub(g.Zuletzt).Seconds()
	if dt < 0 {
		dt = 0
	}
	if dt > tronMaxVorsprung {
		dt = tronMaxVorsprung
	}
	g.Zuletzt = jetzt
	g.Rest += dt
	schritte := 0
	for g.Rest >= tronTakt && g.Status == "spiel" && schritte < tronMaxSchritte {
		g.Rest -= tronTakt
		schritte++
		tronBotDenken(g)
		tronSchritt(g)
	}
	if schritte >= tronMaxSchritte {
		g.Rest = 0
	}
}

// tronRLE presst das Raster in Lauflängen: "0x120,1x5,2x9".
func tronRLE(raster []uint8) string {
	if len(raster) == 0 {
		return ""
	}
	var b strings.Builder
	wert := raster[0]
	anzahl := 1
	schreiben := func() {
		b.WriteString(strconv.Itoa(int(wert)))
		b.WriteByte('x')
		b.WriteString(strconv.Itoa(anzahl))
	}
	for _, v := range raster[1:] {
		if v == wert {
			anzahl++
			continue
		}
		schreiben()
		b.WriteByte(',')
		wert = v
		anzahl = 1
	}
	schreiben()
	return b.String()
}

// ---------------------------------------------------------------- Sicht

func tronAnsicht(g *TronGame, id string) *TronAnsicht {
	if g == nil {
		return nil
	}
	a := &TronAnsicht{
		ID:       g.ID,
		HostID:   g.HostID,
		Status:   g.Status,
		Spalten:  tronSpalten,
		Zeilen:   tronZeilen,
		Runde:    g.Runde,
		Ziel:     tronZiel,
		Gewinner: g.Gewinner,
		Raster:   tronRLE(g.Raster),
		Du:       tronSitz(g, id),
		Meldung:  g.Meldung,
		Note:     g.Note,
		Sieger:   g.Sieger,
		Tick:     time.Now().UnixMilli(),
	}
	for _, p := range g.Spieler {
		a.Spieler = append(a.Spieler, TronSpielerSicht{
			ID: p.ID, Name: p.Name, Ready: p.Ready,
			X: p.X, Y: p.Y, DX: p.DX, DY: p.DY, Tot: p.Tot, Sieg: p.Sieg,
		})
	}
	return a
}

// ---------------------------------------------------------------- Endpunkt

func tronHandler(w http.ResponseWriter, r *http.Request) {
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
		tronBotTreiben(room.Tron, time.Now())
		tronWeiter(room.Tron)
		game := tronAnsicht(room.Tron, id)
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

	zeit := time.Now().Format("15:04:05")

	switch action {
	case "bot":
		room.botNamenNeu()
		if room.Tron != nil && !room.gameSlotFree(room.Tron.Status, tronIDs(room.Tron)...) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Partie Tron.", http.StatusConflict)
			return
		}
		g := tronNeu(randomCode(), input.ID, []TronSpieler{
			tronNeuerSpieler(1, input.ID, player.Name),
			tronNeuerSpieler(2, botID(1), room.botName(1)),
		})
		g.Spieler[0].Ready, g.Spieler[1].Ready = true, true
		g.BotID, g.BotStufe = botID(1), botStufe(input.Stufe)
		tronStart(g)
		g.Note = player.Name + " spielt gegen " + room.botName(1) + " (" + botStufeText(g.BotStufe) + ")."
		room.Tron = g
	case "create":
		if room.Tron != nil && room.Tron.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Partie Tron.", http.StatusConflict)
			return
		}
		room.Tron = tronNeu(randomCode(), input.ID,
			[]TronSpieler{tronNeuerSpieler(1, input.ID, player.Name)})
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Partie Tron – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Tron
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if tronSitz(game, input.ID) >= 0 {
			break
		}
		if len(game.Spieler) >= tronMaxSpieler {
			mu.Unlock()
			http.Error(w, "Tron ist ein Duell – der Tisch ist voll.", http.StatusConflict)
			return
		}
		game.Spieler = append(game.Spieler, tronNeuerSpieler(2, input.ID, player.Name))
		game.Note = player.Name + " sitzt mit am Tisch. Beide drücken BEREIT."

	case "ready":
		game := room.Tron
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := tronSitz(game, input.ID)
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
		game := room.Tron
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Partie auf den Start.", http.StatusConflict)
			return
		}
		if len(game.Spieler) < tronMaxSpieler {
			mu.Unlock()
			http.Error(w, "Tron braucht zwei Spieler.", http.StatusConflict)
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
		tronStart(game)

	case "richtung":
		game := room.Tron
		if game == nil || game.Status != "spiel" {
			mu.Unlock()
			http.Error(w, "Gerade läuft keine Runde.", http.StatusConflict)
			return
		}
		sitz := tronSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		dx, dy, gueltig := tronRichtung(input.Dir)
		if !gueltig {
			mu.Unlock()
			http.Error(w, "Unbekannte Richtung.", http.StatusBadRequest)
			return
		}
		if !tronLenken(&game.Spieler[sitz], dx, dy) {
			mu.Unlock()
			http.Error(w, "Diese Richtung geht nicht (umdrehen verboten).", http.StatusConflict)
			return
		}

	case "weiter":
		game := room.Tron
		if game == nil || game.Status != "runde" {
			mu.Unlock()
			http.Error(w, "Die Runde läuft noch.", http.StatusConflict)
			return
		}
		if tronSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		tronNaechsteRunde(game)

	case "reset":
		game := room.Tron
		if game == nil || game.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Eine neue Partie geht erst nach dem Ende.", http.StatusConflict)
			return
		}
		if tronSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		frisch := tronNeu(game.ID, game.HostID, []TronSpieler{})
		frisch.BotID, frisch.BotStufe = game.BotID, game.BotStufe
		frisch.Note = "Neue Partie – erst BEREIT drücken."
		for _, p := range game.Spieler {
			nr := 1
			if p.Spur == 2 {
				nr = 2
			}
			frisch.Spieler = append(frisch.Spieler, tronNeuerSpieler(nr, p.ID, p.Name))
		}
		room.Tron = frisch

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	tronWeiter(room.Tron)
	result := tronAnsicht(room.Tron, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
