package main

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Bubble Shooter als Online-Duell über den Connector: Jeder Spieler hat sein
// eigenes Feld, und der Server rechnet beide Felder selbst – Kugelflug,
// Anhängen, Dreiergruppen, Absturz, Nachschub und die goldene Linie.
// Regeln und Maße sind dieselben wie in der lokalen Duell-Fassung
// (spiele/bubbleshooter/bubbleshooter.html): Feld 196 × 560, Kugelradius 13,
// sieben Spalten, sechs Startreihen, goldene Linie 66 px über dem Boden,
// alle sechs Schüsse eine neue Reihe, vier Farben, Seed 7.

const (
	bubW            = 196.0
	bubH            = 560.0
	bubR            = 13.0
	bubRh           = 22.516660498395405 // R * sqrt(3)
	bubSp           = 7                  // floor((W-2)/(2R))
	bubReihen       = 6
	bubTod          = 66.0 // Abstand der goldenen Linie vom unteren Rand
	bubReihenBisNeu = 6
	bubFarben       = 4
	bubTempo        = 8.5  // Pixel je 1/60 s
	bubWinkelMax    = 1.32 // Radiant
	bubSchussY      = 534.0
	bubTakt         = 1.0 / 60.0
	bubMaxSchritte  = 240 // höchstens vier Sekunden nachrechnen
	bubVorsprung    = 3.0
)

// ---------------------------------------------------------------- Zufall

// bubMulberry bildet mulberry32 aus der Seite nach – beide Felder starten
// dadurch mit derselben Karte, und der Verlauf bleibt nachvollziehbar.
type bubMulberry struct{ a uint32 }

func (m *bubMulberry) next() float64 {
	m.a += 0x6D2B79F5
	t := m.a
	t = (t ^ (t >> 15)) * (1 | t)
	u := (t ^ (t >> 7)) * (61 | t)
	// In der Vorlage steht „t + Math.imul(...) ^ t“; die Addition bindet
	// stärker als das exklusive Oder.
	t = (t + u) ^ t
	return float64(t^(t>>14)) / 4294967296.0
}

// ---------------------------------------------------------------- Brett

type bubReihe struct {
	V int   `json:"v"`
	Z []int `json:"z"` // 0 = leer, sonst Farbe 1..4
}

type bubKugel struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	VX    float64 `json:"-"`
	VY    float64 `json:"-"`
	Farbe int     `json:"farbe"` // 0..3
}

type bubBrett struct {
	Feld     []bubReihe
	Zufall   bubMulberry
	Punkte   int
	Schuesse int
	Aktuell  int
	Naechste int
	Winkel   float64
	Kugel    *bubKugel
	Zustand  string // start, laeuft, ende
}

func bubNeuesBrett() *bubBrett {
	b := &bubBrett{Zufall: bubMulberry{a: 7}, Zustand: "start"}
	b.neu()
	return b
}

func (b *bubBrett) farbe() int {
	return int(math.Floor(b.Zufall.next() * bubFarben))
}

func (b *bubBrett) neu() {
	b.Zufall = bubMulberry{a: 7}
	b.Feld = nil
	for r := 0; r < bubReihen; r++ {
		z := make([]int, bubSp)
		for c := 0; c < bubSp; c++ {
			z[c] = b.farbe() + 1
		}
		b.Feld = append(b.Feld, bubReihe{V: r % 2, Z: z})
	}
	b.Punkte = 0
	b.Kugel = nil
	b.Schuesse = 0
	b.Winkel = 0
	b.Zustand = "start"
	b.Aktuell = b.farbe()
	b.Naechste = b.farbe()
}

// reihenV: Versatz einer Reihe; unter dem Raster gilt die gedachte nächste Reihe.
func (b *bubBrett) reihenV(r int) int {
	if r < len(b.Feld) {
		return b.Feld[r].V
	}
	if len(b.Feld) > 0 {
		return 1 - b.Feld[len(b.Feld)-1].V
	}
	return 0
}

func (b *bubBrett) zelleX(r, c int) float64 {
	x := bubR + 2*bubR*float64(c)
	if b.reihenV(r) != 0 {
		x += bubR
	}
	return x
}

func (b *bubBrett) zelleY(r int) float64 {
	return bubR + bubRh*float64(r)
}

func (b *bubBrett) nachbarn(r, c int) [][2]int {
	v := b.reihenV(r)
	drs := [][2]int{{-1, -1}, {-1, 0}, {0, -1}, {0, 1}, {1, -1}, {1, 0}}
	if v != 0 {
		drs = [][2]int{{-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, 0}, {1, 1}}
	}
	liste := [][2]int{}
	for _, d := range drs {
		rr, cc := r+d[0], c+d[1]
		if rr < 0 || rr > len(b.Feld) {
			continue
		}
		if cc < 0 || cc >= bubSp {
			continue
		}
		liste = append(liste, [2]int{rr, cc})
	}
	return liste
}

func (b *bubBrett) kugelBei(r, c int) bool {
	return r < len(b.Feld) && r >= 0 && c >= 0 && c < bubSp && b.Feld[r].Z[c] != 0
}

func (b *bubBrett) los() {
	b.Zustand = "laeuft"
}

func (b *bubBrett) schiessen(winkel float64) bool {
	if b.Zustand != "laeuft" || b.Kugel != nil {
		return false
	}
	if winkel < -bubWinkelMax {
		winkel = -bubWinkelMax
	}
	if winkel > bubWinkelMax {
		winkel = bubWinkelMax
	}
	b.Winkel = winkel
	b.Kugel = &bubKugel{
		X:     bubW / 2,
		Y:     bubSchussY - bubR,
		VX:    math.Sin(winkel) * bubTempo,
		VY:    -math.Cos(winkel) * bubTempo,
		Farbe: b.Aktuell,
	}
	b.Aktuell = b.Naechste
	b.Naechste = b.farbe()
	b.Schuesse++
	return true
}

// freieZellen sammelt alle leeren Zellen, die an eine Kugel grenzen – in der
// Reihenfolge, in der sie gefunden werden (bei gleich weiten Kandidaten
// entscheidet sie).
func (b *bubBrett) freieZellen() [][2]int {
	gesehen := map[[2]int]bool{}
	liste := [][2]int{}
	merken := func(z [2]int) {
		if !gesehen[z] {
			gesehen[z] = true
			liste = append(liste, z)
		}
	}
	hatKugel := false
	for r := 0; r < len(b.Feld); r++ {
		for c := 0; c < bubSp; c++ {
			if !b.kugelBei(r, c) {
				continue
			}
			hatKugel = true
			for _, n := range b.nachbarn(r, c) {
				if !b.kugelBei(n[0], n[1]) {
					merken(n)
				}
			}
		}
	}
	if !hatKugel {
		if len(b.Feld) == 0 {
			b.Feld = append(b.Feld, bubReihe{V: 0, Z: make([]int, bubSp)})
		}
		for c := 0; c < bubSp; c++ {
			merken([2]int{0, c})
		}
	}
	return liste
}

func (b *bubBrett) anhaengen(x, y float64) {
	farbe := 0
	if b.Kugel != nil {
		farbe = b.Kugel.Farbe + 1
	}
	kandidaten := b.freieZellen()
	bestD := math.Inf(1)
	beste := [2]int{-1, -1}
	for _, k := range kandidaten {
		dx := b.zelleX(k[0], k[1]) - x
		dy := b.zelleY(k[0]) - y
		d := dx*dx + dy*dy
		if d < bestD {
			bestD = d
			beste = k
		}
	}
	b.Kugel = nil
	if beste[0] < 0 {
		return
	}
	r, c := beste[0], beste[1]
	if r >= len(b.Feld) {
		b.Feld = append(b.Feld, bubReihe{V: b.reihenV(r), Z: make([]int, bubSp)})
	}
	b.Feld[r].Z[c] = farbe
	b.pruefen(r, c)
}

func (b *bubBrett) gleiche(r, c int) [][2]int {
	farbe := b.Feld[r].Z[c]
	if farbe == 0 {
		return nil
	}
	gesehen := map[[2]int]bool{{r, c}: true}
	stapel := [][2]int{{r, c}}
	treffer := [][2]int{}
	for len(stapel) > 0 {
		zelle := stapel[len(stapel)-1]
		stapel = stapel[:len(stapel)-1]
		treffer = append(treffer, zelle)
		for _, n := range b.nachbarn(zelle[0], zelle[1]) {
			if gesehen[n] || !b.kugelBei(n[0], n[1]) {
				continue
			}
			if b.Feld[n[0]].Z[n[1]] != farbe {
				continue
			}
			gesehen[n] = true
			stapel = append(stapel, n)
		}
	}
	return treffer
}

func (b *bubBrett) schwebende() [][2]int {
	erreichbar := map[[2]int]bool{}
	stapel := [][2]int{}
	for c := 0; c < bubSp; c++ {
		if b.kugelBei(0, c) {
			erreichbar[[2]int{0, c}] = true
			stapel = append(stapel, [2]int{0, c})
		}
	}
	for len(stapel) > 0 {
		zelle := stapel[len(stapel)-1]
		stapel = stapel[:len(stapel)-1]
		for _, n := range b.nachbarn(zelle[0], zelle[1]) {
			if erreichbar[n] || !b.kugelBei(n[0], n[1]) {
				continue
			}
			erreichbar[n] = true
			stapel = append(stapel, n)
		}
	}
	weg := [][2]int{}
	for r := 0; r < len(b.Feld); r++ {
		for c := 0; c < bubSp; c++ {
			if !b.kugelBei(r, c) {
				continue
			}
			if !erreichbar[[2]int{r, c}] {
				weg = append(weg, [2]int{r, c})
			}
		}
	}
	return weg
}

func (b *bubBrett) pruefen(r, c int) {
	gruppe := b.gleiche(r, c)
	if len(gruppe) >= 3 {
		for _, z := range gruppe {
			b.Feld[z[0]].Z[z[1]] = 0
		}
		b.Punkte += 10 * len(gruppe)
		for _, z := range b.schwebende() {
			if b.Feld[z[0]].Z[z[1]] == 0 {
				continue
			}
			b.Feld[z[0]].Z[z[1]] = 0
			b.Punkte += 20
		}
	}
	b.pruefeEnde()
}

func (b *bubBrett) pruefeEnde() {
	for r := 0; r < len(b.Feld); r++ {
		for c := 0; c < bubSp; c++ {
			if !b.kugelBei(r, c) {
				continue
			}
			if b.zelleY(r)+bubR >= bubH-bubTod {
				b.ende()
				return
			}
		}
	}
}

func (b *bubBrett) neueReihe() {
	v := 0
	if len(b.Feld) > 0 {
		v = 1 - b.Feld[0].V
	}
	z := make([]int, bubSp)
	for c := 0; c < bubSp; c++ {
		z[c] = b.farbe() + 1
	}
	b.Feld = append([]bubReihe{{V: v, Z: z}}, b.Feld...)
	b.pruefeEnde()
}

func (b *bubBrett) trifftKugel(x, y float64) bool {
	for r := 0; r < len(b.Feld); r++ {
		for c := 0; c < bubSp; c++ {
			if !b.kugelBei(r, c) {
				continue
			}
			dx := b.zelleX(r, c) - x
			dy := b.zelleY(r) - y
			if dx*dx+dy*dy < bubR*bubR*3.1 {
				return true
			}
		}
	}
	return false
}

func (b *bubBrett) bewegen() {
	k := b.Kugel
	schritte := int(math.Ceil(math.Hypot(k.VX, k.VY) / 3))
	if schritte < 1 {
		schritte = 1
	}
	sx := k.VX / float64(schritte)
	sy := k.VY / float64(schritte)
	for i := 0; i < schritte; i++ {
		k.X += sx
		k.Y += sy
		if k.X < bubR+1 {
			k.X = bubR + 1
			k.VX = math.Abs(k.VX)
		}
		if k.X > bubW-bubR-1 {
			k.X = bubW - bubR - 1
			k.VX = -math.Abs(k.VX)
		}
		if k.Y <= bubR+1 {
			b.anhaengen(k.X, k.Y)
			return
		}
		if b.trifftKugel(k.X, k.Y) {
			b.anhaengen(k.X, k.Y)
			return
		}
	}
}

func (b *bubBrett) schritt() {
	if b.Zustand != "laeuft" {
		return
	}
	if b.Kugel != nil {
		b.bewegen()
	} else if b.Schuesse >= bubReihenBisNeu {
		b.Schuesse = 0
		b.neueReihe()
	}
}

func (b *bubBrett) ende() {
	if b.Zustand != "laeuft" {
		return
	}
	b.Zustand = "ende"
}

func (b *bubBrett) kugeln() int {
	n := 0
	for r := range b.Feld {
		for c := 0; c < bubSp; c++ {
			if b.Feld[r].Z[c] != 0 {
				n++
			}
		}
	}
	return n
}

// ---------------------------------------------------------------- Partie

type BubbleSpieler struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Ready bool      `json:"ready"`
	Nr    int       `json:"-"`
	Brett *bubBrett `json:"-"`
}

type BubbleGame struct {
	ID       string          `json:"id"`
	HostID   string          `json:"hostId"`
	Status   string          `json:"status"` // invited, spiel, ende
	Spieler  []BubbleSpieler `json:"spieler"`
	Gewinner int             `json:"-"`
	Meldung  string          `json:"meldung"`
	Note     string          `json:"note"`
	Sieger   string          `json:"sieger,omitempty"`
	Zuletzt  time.Time       `json:"-"`
	Rest     float64         `json:"-"`
}

type BubbleBrettSicht struct {
	Reihen           []bubReihe `json:"reihen"`
	Punkte           int        `json:"punkte"`
	Schuesse         int        `json:"schuesse"`
	Kugeln           int        `json:"kugeln"`
	Aktuell          int        `json:"aktuell"`
	Naechste         int        `json:"naechste"`
	Winkel           float64    `json:"winkel"`
	Kugel            *bubKugel  `json:"kugel"`
	SchuesseBisReihe int        `json:"schuesseBisReihe"`
	Zustand          string     `json:"zustand"`
}

type BubbleSpielerSicht struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Ready            bool   `json:"ready"`
	Nr               int    `json:"nr"`
	Punkte           int    `json:"punkte"`
	Kugeln           int    `json:"kugeln"`
	Reihen           int    `json:"reihen"`
	SchuesseBisReihe int    `json:"schuesseBisReihe"`
	Verloren         bool   `json:"verloren"`
}

type BubbleAnsicht struct {
	ID       string               `json:"id"`
	HostID   string               `json:"hostId"`
	Status   string               `json:"status"`
	Spalten  int                  `json:"spalten"`
	Radius   float64              `json:"radius"`
	Breite   float64              `json:"breite"`
	Hoehe    float64              `json:"hoehe"`
	Linie    float64              `json:"linie"`
	Brett    *BubbleBrettSicht    `json:"brett"`
	Spieler  []BubbleSpielerSicht `json:"spieler"`
	Gewinner int                  `json:"gewinner"`
	Meldung  string               `json:"meldung"`
	Note     string               `json:"note"`
	Sieger   string               `json:"sieger,omitempty"`
	Du       int                  `json:"du"`
	Tick     int64                `json:"tick"`
}

func bubbleSitz(g *BubbleGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func bubbleNeu(ID, hostID string, spieler []BubbleSpieler) *BubbleGame {
	return &BubbleGame{
		ID:       ID,
		HostID:   hostID,
		Status:   "invited",
		Spieler:  spieler,
		Gewinner: -1,
		Zuletzt:  time.Now(),
		Note:     "Tisch eröffnet – gesucht wird ein zweiter Spieler.",
	}
}

func (g *BubbleGame) starten() {
	for i := range g.Spieler {
		g.Spieler[i].Brett = bubNeuesBrett()
		g.Spieler[i].Brett.los()
	}
	g.Status = "spiel"
	g.Gewinner = -1
	g.Sieger = ""
	g.Meldung = ""
	g.Note = "Los! Wer zuerst die goldene Linie überschreitet, verliert."
	g.Rest = 0
	g.Zuletzt = time.Now()
}

// bubbleWeiter rechnet die Kugeln beider Felder bis jetzt weiter.
func bubbleWeiter(g *BubbleGame) {
	if g == nil || g.Status != "spiel" {
		return
	}
	jetzt := time.Now()
	dt := jetzt.Sub(g.Zuletzt).Seconds()
	if dt < 0 {
		dt = 0
	}
	if dt > bubVorsprung {
		dt = bubVorsprung
	}
	g.Zuletzt = jetzt
	g.Rest += dt
	schritte := 0
	for g.Rest >= bubTakt && g.Status == "spiel" && schritte < bubMaxSchritte {
		g.Rest -= bubTakt
		schritte++
		for i := range g.Spieler {
			if g.Spieler[i].Brett != nil {
				g.Spieler[i].Brett.schritt()
			}
		}
		for i := range g.Spieler {
			if g.Spieler[i].Brett != nil && g.Spieler[i].Brett.Zustand == "ende" {
				g.verloren(i)
				break
			}
		}
	}
	if schritte >= bubMaxSchritte {
		g.Rest = 0
	}
}

func (g *BubbleGame) verloren(sitz int) {
	if g.Status != "spiel" {
		return
	}
	verlierer := g.Spieler[sitz]
	gewinner := g.Spieler[1-sitz]
	g.Status = "ende"
	g.Gewinner = gewinner.Nr
	g.Sieger = gewinner.Name
	g.Meldung = verlierer.Name + " hat die goldene Linie überschritten – " + gewinner.Name +
		" gewinnt mit " + strconv.Itoa(gewinner.Brett.Punkte) + " zu " + strconv.Itoa(verlierer.Brett.Punkte) + " Punkten."
	g.Note = "Neue Partie startet mit einem Klick."
}

func bubbleAnsicht(g *BubbleGame, id string) *BubbleAnsicht {
	if g == nil {
		return nil
	}
	du := bubbleSitz(g, id)
	a := &BubbleAnsicht{
		ID:       g.ID,
		HostID:   g.HostID,
		Status:   g.Status,
		Spalten:  bubSp,
		Radius:   bubR,
		Breite:   bubW,
		Hoehe:    bubH,
		Linie:    bubH - bubTod,
		Gewinner: g.Gewinner,
		Meldung:  g.Meldung,
		Note:     g.Note,
		Sieger:   g.Sieger,
		Du:       du,
		Tick:     time.Now().UnixMilli(),
	}
	a.Spieler = []BubbleSpielerSicht{}
	for _, p := range g.Spieler {
		s := BubbleSpielerSicht{ID: p.ID, Name: p.Name, Ready: p.Ready, Nr: p.Nr}
		if p.Brett != nil {
			s.Punkte = p.Brett.Punkte
			s.Kugeln = p.Brett.kugeln()
			s.Reihen = len(p.Brett.Feld)
			s.SchuesseBisReihe = bubReihenBisNeu - p.Brett.Schuesse
			s.Verloren = p.Brett.Zustand == "ende"
		}
		a.Spieler = append(a.Spieler, s)
	}
	if du >= 0 && g.Spieler[du].Brett != nil {
		b := g.Spieler[du].Brett
		sicht := &BubbleBrettSicht{
			Reihen:           b.Feld,
			Punkte:           b.Punkte,
			Schuesse:         b.Schuesse,
			Kugeln:           b.kugeln(),
			Aktuell:          b.Aktuell,
			Naechste:         b.Naechste,
			Winkel:           math.Round(b.Winkel*1000) / 1000,
			SchuesseBisReihe: bubReihenBisNeu - b.Schuesse,
			Zustand:          b.Zustand,
		}
		if b.Kugel != nil {
			sicht.Kugel = &bubKugel{
				X:     math.Round(b.Kugel.X),
				Y:     math.Round(b.Kugel.Y),
				Farbe: b.Kugel.Farbe,
			}
		}
		a.Brett = sicht
	}
	return a
}

// ---------------------------------------------------------------- Endpunkt

func bubbleHandler(w http.ResponseWriter, r *http.Request) {
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
		bubbleWeiter(room.Bubble)
		game := bubbleAnsicht(room.Bubble, id)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID     string  `json:"id"`
		Winkel float64 `json:"winkel"`
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
	case "create":
		if room.Bubble != nil && room.Bubble.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon ein Bubble-Shooter-Duell.", http.StatusConflict)
			return
		}
		room.Bubble = bubbleNeu(randomCode(), input.ID,
			[]BubbleSpieler{{ID: input.ID, Name: player.Name, Nr: 1}})
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet ein Duell in Bubble Shooter – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Bubble
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if bubbleSitz(game, input.ID) >= 0 {
			break
		}
		if len(game.Spieler) >= 2 {
			mu.Unlock()
			http.Error(w, "Der Tisch ist voll.", http.StatusConflict)
			return
		}
		game.Spieler = append(game.Spieler, BubbleSpieler{ID: input.ID, Name: player.Name, Nr: 2})
		game.Note = player.Name + " sitzt mit am Tisch. Beide drücken BEREIT."

	case "ready":
		game := room.Bubble
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := bubbleSitz(game, input.ID)
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
		game := room.Bubble
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Duell auf den Start.", http.StatusConflict)
			return
		}
		if len(game.Spieler) < 2 {
			mu.Unlock()
			http.Error(w, "Das Duell braucht zwei Spieler.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber startet das Duell.", http.StatusForbidden)
			return
		}
		for _, p := range game.Spieler {
			if !p.Ready {
				mu.Unlock()
				http.Error(w, "Erst müssen beide BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.starten()

	case "schiessen":
		game := room.Bubble
		if game == nil || game.Status != "spiel" {
			mu.Unlock()
			http.Error(w, "Gerade läuft kein Duell.", http.StatusConflict)
			return
		}
		sitz := bubbleSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		if input.Winkel != input.Winkel || input.Winkel > 10 || input.Winkel < -10 {
			mu.Unlock()
			http.Error(w, "Ungültiger Winkel.", http.StatusBadRequest)
			return
		}
		if !game.Spieler[sitz].Brett.schiessen(input.Winkel) {
			mu.Unlock()
			http.Error(w, "Jetzt geht kein Schuss.", http.StatusConflict)
			return
		}

	case "reset":
		game := room.Bubble
		if game == nil || game.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Ein neues Duell geht erst nach dem Ende.", http.StatusConflict)
			return
		}
		if bubbleSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu diesem Duell.", http.StatusForbidden)
			return
		}
		frisch := bubbleNeu(game.ID, game.HostID, []BubbleSpieler{})
		frisch.Note = "Neues Duell – erst BEREIT drücken."
		for _, p := range game.Spieler {
			frisch.Spieler = append(frisch.Spieler, BubbleSpieler{ID: p.ID, Name: p.Name, Nr: p.Nr})
		}
		room.Bubble = frisch

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	bubbleWeiter(room.Bubble)
	result := bubbleAnsicht(room.Bubble, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
