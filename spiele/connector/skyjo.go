package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Skyjo: je 12 Karten in einem 4x3-Feld, zwei Startkarten offen,
// drei gleiche Werte in einer Spalte zählen 0, wer am Ende die wenigsten
// Punkte hat, gewinnt. Gespielt wird bis 100 Punkte.
//
// Mehrspieler: 2 bis 8 Plätze, Menschen und Bots gemischt. Der Server hält
// den vollständigen Stand und gibt jedem Spieler nur seine eigenen Karten
// vollständig; fremde verdeckte Karten verlassen den Server nie.
const (
	skyjoKarten  = 12
	skyjoSpalten = 4
	skyjoZiel    = 100
	skyjoMax     = 8
)

var skyjoCounts = [][2]int{
	{-2, 5}, {-1, 10}, {0, 15},
	{1, 10}, {2, 10}, {3, 10}, {4, 10},
	{5, 10}, {6, 10}, {7, 10}, {8, 10},
	{9, 10}, {10, 10}, {11, 10}, {12, 10},
}

type SkyjoKarte struct {
	Wert  int  `json:"wert"`
	Offen bool `json:"offen"`
	Weg   bool `json:"weg"`
}

type SkyjoSpieler struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Bot    bool         `json:"bot"`
	Stufe  string       `json:"stufe,omitempty"`
	Karten []SkyjoKarte `json:"-"`
	Runden int          `json:"runden"`
	Gesamt int          `json:"gesamt"`
}

type SkyjoRunde struct {
	Runde   int      `json:"runde"`
	Werte   []int    `json:"werte"`
	Gesamte []int    `json:"gesamte"`
	Sieger  []string `json:"sieger"`
}

type SkyjoGame struct {
	ID          string         `json:"id"`
	HostID      string         `json:"hostId"`
	Spieler     []SkyjoSpieler `json:"-"`
	Deck        []int          `json:"-"`
	Ablage      []int          `json:"ablage"`
	Zug         int            `json:"zug"`
	Phase       string         `json:"phase"` // setup, play, reveal, between, fertig, warten
	Gezogen     *int           `json:"gezogen"`
	Quelle      string         `json:"quelle"`
	Finisher    int            `json:"finisher"`
	RestZuege   int            `json:"-"`
	Runde       int            `json:"runde"`
	Status      string         `json:"status"` // invited, active, finished
	Sieger      []string       `json:"sieger,omitempty"`
	Notiz       string         `json:"notiz"`
	LetzteRunde *SkyjoRunde    `json:"letzteRunde,omitempty"`
	Sitze       int            `json:"sitze"`
	BotFrist    time.Time      `json:"-"`
}

func skyjoMakeDeck() []int {
	deck := []int{}
	for _, e := range skyjoCounts {
		for i := 0; i < e[1]; i++ {
			deck = append(deck, e[0])
		}
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

func skyjoNeueKarten() []SkyjoKarte {
	k := make([]SkyjoKarte, skyjoKarten)
	return k
}

func skyjoAusteilen(g *SkyjoGame) {
	g.Deck = skyjoMakeDeck()
	g.Ablage = []int{}
	for i := range g.Spieler {
		g.Spieler[i].Runden = 0
		g.Spieler[i].Karten = skyjoNeueKarten()
		for k := 0; k < skyjoKarten; k++ {
			g.Spieler[i].Karten[k] = SkyjoKarte{Wert: g.Deck[len(g.Deck)-1]}
			g.Deck = g.Deck[:len(g.Deck)-1]
		}
	}
	g.Ablage = append(g.Ablage, g.Deck[len(g.Deck)-1])
	g.Deck = g.Deck[:len(g.Deck)-1]
	g.Zug = 0
	g.Phase = "setup"
	g.Gezogen = nil
	g.Quelle = ""
	g.Finisher = -1
	g.RestZuege = 0
	g.Notiz = g.Spieler[0].Name + " deckt zwei Karten auf."
}

func skyjoNeu(id, hostID, hostName string, sitze int) *SkyjoGame {
	g := &SkyjoGame{ID: id, HostID: hostID, Runde: 1, Status: "invited", Finisher: -1, Sitze: sitze}
	g.Spieler = []SkyjoSpieler{{ID: hostID, Name: hostName}}
	g.Ablage = []int{}
	g.Notiz = "Warte auf Mitspieler."
	return g
}

func (g *SkyjoGame) indexVon(id string) int {
	for i, s := range g.Spieler {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func (g *SkyjoGame) name(i int) string {
	if i < 0 || i >= len(g.Spieler) {
		return "—"
	}
	return g.Spieler[i].Name
}

func (g *SkyjoGame) clone() *SkyjoGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Deck = append([]int(nil), g.Deck...)
	c.Ablage = append([]int(nil), g.Ablage...)
	c.Sieger = append([]string(nil), g.Sieger...)
	if g.Gezogen != nil {
		v := *g.Gezogen
		c.Gezogen = &v
	}
	c.Spieler = make([]SkyjoSpieler, len(g.Spieler))
	for i, s := range g.Spieler {
		c.Spieler[i] = s
		c.Spieler[i].Karten = append([]SkyjoKarte(nil), s.Karten...)
	}
	if g.LetzteRunde != nil {
		r := *g.LetzteRunde
		r.Werte = append([]int(nil), g.LetzteRunde.Werte...)
		r.Gesamte = append([]int(nil), g.LetzteRunde.Gesamte...)
		r.Sieger = append([]string(nil), g.LetzteRunde.Sieger...)
		c.LetzteRunde = &r
	}
	return &c
}

func skyjoSpaltenPruefen(p *SkyjoSpieler) bool {
	for s := 0; s < skyjoSpalten; s++ {
		idx := []int{s, s + skyjoSpalten, s + skyjoSpalten*2}
		ok := true
		for _, i := range idx {
			if p.Karten[i].Weg || !p.Karten[i].Offen {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		w := p.Karten[idx[0]].Wert
		if p.Karten[idx[1]].Wert == w && p.Karten[idx[2]].Wert == w {
			for _, i := range idx {
				p.Karten[i].Weg = true
			}
			return true
		}
	}
	return false
}

func skyjoVerdeckt(p *SkyjoSpieler) int {
	n := 0
	for _, k := range p.Karten {
		if !k.Offen && !k.Weg {
			n++
		}
	}
	return n
}

func skyjoPunkte(p *SkyjoSpieler) int {
	summe := 0
	for _, k := range p.Karten {
		if !k.Weg {
			summe += k.Wert
		}
	}
	return summe
}

func (g *SkyjoGame) naechster(i int) int {
	return (i + 1) % len(g.Spieler)
}

func (g *SkyjoGame) zugWeiter() {
	if g.Finisher >= 0 {
		g.RestZuege--
		if g.RestZuege <= 0 {
			g.rundeBeenden()
			return
		}
	}
	g.Zug = g.naechster(g.Zug)
	g.Phase = "play"
	g.Gezogen = nil
	g.Quelle = ""
	g.Notiz = g.name(g.Zug) + " ist am Zug."
}

func (g *SkyjoGame) zugBeenden() {
	p := &g.Spieler[g.Zug]
	if g.Finisher < 0 && skyjoVerdeckt(p) == 0 {
		g.Finisher = g.Zug
		g.RestZuege = len(g.Spieler)
		g.Notiz = p.Name + " hat alle Karten offen – jeder andere hat noch einen Zug."
	}
	g.zugWeiter()
}

func (g *SkyjoGame) rundeBeenden() {
	for i := range g.Spieler {
		p := &g.Spieler[i]
		for k := range p.Karten {
			if !p.Karten[k].Weg {
				p.Karten[k].Offen = true
			}
		}
		for skyjoSpaltenPruefen(p) {
		}
	}
	werte := make([]int, len(g.Spieler))
	for i := range g.Spieler {
		werte[i] = skyjoPunkte(&g.Spieler[i])
	}
	if g.Finisher >= 0 {
		minAndere := 1 << 30
		for i, w := range werte {
			if i != g.Finisher && w < minAndere {
				minAndere = w
			}
		}
		if werte[g.Finisher] >= minAndere && werte[g.Finisher] > 0 {
			werte[g.Finisher] *= 2
		}
	}
	for i := range g.Spieler {
		g.Spieler[i].Runden = werte[i]
		g.Spieler[i].Gesamt += werte[i]
	}
	beste := werte[0]
	for _, w := range werte {
		if w < beste {
			beste = w
		}
	}
	sieger := []string{}
	for i, w := range werte {
		if w == beste {
			sieger = append(sieger, g.Spieler[i].Name)
		}
	}
	gesamte := make([]int, len(g.Spieler))
	for i := range g.Spieler {
		gesamte[i] = g.Spieler[i].Gesamt
	}
	g.LetzteRunde = &SkyjoRunde{Runde: g.Runde, Werte: werte, Gesamte: gesamte, Sieger: sieger}
	g.Phase = "between"
	if len(sieger) == 1 {
		g.Notiz = "Runde " + itoa(g.Runde) + " beendet – " + sieger[0] + " gewinnt die Runde."
	} else {
		g.Notiz = "Runde " + itoa(g.Runde) + " beendet – unentschieden."
	}
	for i := range g.Spieler {
		if g.Spieler[i].Gesamt >= skyjoZiel {
			g.spielBeenden()
			return
		}
	}
}

func (g *SkyjoGame) spielBeenden() {
	g.Phase = "fertig"
	g.Status = "finished"
	beste := g.Spieler[0].Gesamt
	for _, s := range g.Spieler {
		if s.Gesamt < beste {
			beste = s.Gesamt
		}
	}
	kandidaten := []int{}
	for i, s := range g.Spieler {
		if s.Gesamt == beste {
			kandidaten = append(kandidaten, i)
		}
	}
	if len(kandidaten) > 1 && g.LetzteRunde != nil {
		bestRunde := 1 << 30
		for _, i := range kandidaten {
			if g.LetzteRunde.Werte[i] < bestRunde {
				bestRunde = g.LetzteRunde.Werte[i]
			}
		}
		gefiltert := []int{}
		for _, i := range kandidaten {
			if g.LetzteRunde.Werte[i] == bestRunde {
				gefiltert = append(gefiltert, i)
			}
		}
		kandidaten = gefiltert
	}
	g.Sieger = []string{}
	for _, i := range kandidaten {
		g.Sieger = append(g.Sieger, g.Spieler[i].Name)
	}
	if len(g.Sieger) == 1 {
		g.Notiz = g.Sieger[0] + " gewinnt mit " + itoa(beste) + " Punkten."
	} else {
		g.Notiz = "Unentschieden bei " + itoa(beste) + " Punkten."
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	ziffern := []byte{}
	for n > 0 {
		ziffern = append([]byte{byte('0' + n%10)}, ziffern...)
		n /= 10
	}
	if neg {
		return "-" + string(ziffern)
	}
	return string(ziffern)
}

// skyjoAktion führt eine Aktion für Spieler me aus. Rückgabe 0 = ausgeführt.
func (g *SkyjoGame) skyjoAktion(me int, akt string, index int, quelle string) (int, string) {
	if g == nil || g.Status != "active" {
		return http.StatusConflict, "Es läuft gerade keine aktive Partie."
	}
	if me < 0 || me >= len(g.Spieler) {
		return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
	}
	p := &g.Spieler[me]
	switch akt {
	case "reveal":
		if g.Phase != "setup" {
			return http.StatusConflict, "Das Aufdecken ist schon vorbei."
		}
		if g.Zug != me {
			return http.StatusConflict, "Ein anderer Spieler deckt gerade auf."
		}
		if index < 0 || index >= skyjoKarten || p.Karten[index].Offen || p.Karten[index].Weg {
			return http.StatusBadRequest, "Diese Karte gibt es nicht."
		}
		p.Karten[index].Offen = true
		offen := 0
		for _, k := range p.Karten {
			if k.Offen {
				offen++
			}
		}
		if offen < 2 {
			g.Notiz = p.Name + " deckt noch eine Karte auf."
			break
		}
		fertig := true
		for i := range g.Spieler {
			anz := 0
			for _, k := range g.Spieler[i].Karten {
				if k.Offen {
					anz++
				}
			}
			if anz < 2 {
				fertig = false
			}
		}
		if !fertig {
			g.Zug = g.naechster(g.Zug)
			g.Notiz = g.name(g.Zug) + " deckt zwei Karten auf."
			break
		}
		// Höchste Startsumme beginnt.
		beste, start := -1<<30, 0
		for i := range g.Spieler {
			summe := 0
			for _, k := range g.Spieler[i].Karten {
				if k.Offen {
					summe += k.Wert
				}
			}
			if summe > beste {
				beste, start = summe, i
			} else if summe == beste && rand.Intn(2) == 0 {
				start = i
			}
		}
		g.Zug = start
		g.Phase = "play"
		g.Notiz = g.name(start) + " beginnt (höchste Startsumme)."

	case "draw":
		if g.Phase != "play" || g.Zug != me {
			return http.StatusConflict, "Du bist nicht am Zug."
		}
		if g.Gezogen != nil {
			return http.StatusConflict, "Du hast schon eine Karte gezogen."
		}
		if strings.ToLower(strings.TrimSpace(quelle)) == "ablage" {
			if len(g.Ablage) == 0 {
				return http.StatusConflict, "Der Ablagestapel ist leer."
			}
			w := g.Ablage[len(g.Ablage)-1]
			g.Ablage = g.Ablage[:len(g.Ablage)-1]
			g.Gezogen = &w
			g.Quelle = "ablage"
			g.Notiz = p.Name + " nimmt die Ablagekarte."
			break
		}
		if len(g.Deck) == 0 {
			if len(g.Ablage) <= 1 {
				return http.StatusConflict, "Keine Karten mehr im Nachziehstapel."
			}
			oben := g.Ablage[len(g.Ablage)-1]
			rest := append([]int(nil), g.Ablage[:len(g.Ablage)-1]...)
			rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
			g.Deck = rest
			g.Ablage = []int{oben}
		}
		w := g.Deck[len(g.Deck)-1]
		g.Deck = g.Deck[:len(g.Deck)-1]
		g.Gezogen = &w
		g.Quelle = "deck"
		g.Notiz = p.Name + " zieht vom Stapel."

	case "swap":
		if g.Phase != "play" || g.Zug != me {
			return http.StatusConflict, "Du bist nicht am Zug."
		}
		if g.Gezogen == nil {
			return http.StatusConflict, "Du hast noch keine Karte gezogen."
		}
		if index < 0 || index >= skyjoKarten || p.Karten[index].Weg {
			return http.StatusBadRequest, "Diese Karte gibt es nicht."
		}
		alt := p.Karten[index].Wert
		p.Karten[index].Wert = *g.Gezogen
		p.Karten[index].Offen = true
		g.Ablage = append(g.Ablage, alt)
		g.Gezogen = nil
		g.Quelle = ""
		skyjoSpaltenPruefen(p)
		g.zugBeenden()

	case "discard":
		if g.Phase != "play" || g.Zug != me {
			return http.StatusConflict, "Du bist nicht am Zug."
		}
		if g.Gezogen == nil || g.Quelle != "deck" {
			return http.StatusConflict, "Es liegt keine Karte zum Abwerfen bereit."
		}
		g.Ablage = append(g.Ablage, *g.Gezogen)
		g.Gezogen = nil
		g.Quelle = ""
		if skyjoVerdeckt(p) == 0 {
			g.zugBeenden()
			break
		}
		g.Phase = "reveal"
		g.Notiz = p.Name + " deckt eine eigene Karte auf."

	case "opencard":
		if (g.Phase != "reveal" && g.Phase != "play") || g.Zug != me {
			return http.StatusConflict, "Du bist nicht am Zug."
		}
		if g.Phase == "play" && g.Gezogen == nil {
			return http.StatusConflict, "Du hast noch keine Karte gezogen."
		}
		if index < 0 || index >= skyjoKarten || p.Karten[index].Offen || p.Karten[index].Weg {
			return http.StatusBadRequest, "Diese Karte gibt es nicht."
		}
		p.Karten[index].Offen = true
		skyjoSpaltenPruefen(p)
		g.Phase = "play"
		g.zugBeenden()

	case "nextround":
		if g.Phase != "between" {
			return http.StatusConflict, "Eine neue Runde ist gerade nicht möglich."
		}
		g.Runde++
		skyjoAusteilen(g)

	case "start":
		// nur Anstoß für die Anzeige

	default:
		return http.StatusBadRequest, "Unbekannte Skyjo-Aktion."
	}
	return 0, ""
}

// skyjoSicht baut die Antwort für einen Spieler: eigene Karten vollständig,
// fremde nur, wenn sie offen sind.
func skyjoSicht(g *SkyjoGame, id string) map[string]any {
	if g == nil {
		return nil
	}
	me := g.indexVon(id)
	spieler := make([]map[string]any, len(g.Spieler))
	for i, s := range g.Spieler {
		karten := make([]map[string]any, len(s.Karten))
		for k, karte := range s.Karten {
			eintrag := map[string]any{"offen": karte.Offen, "weg": karte.Weg}
			if karte.Offen || i == me {
				eintrag["wert"] = karte.Wert
			} else {
				eintrag["wert"] = nil
			}
			karten[k] = eintrag
		}
		offen := 0
		for _, k := range s.Karten {
			if k.Offen {
				offen++
			}
		}
		spieler[i] = map[string]any{
			"id": s.ID, "name": s.Name, "bot": s.Bot, "stufe": s.Stufe,
			"gesamt": s.Gesamt, "runden": s.Runden,
			"karten": karten, "offen": offen, "verdeckt": skyjoVerdeckt(&g.Spieler[i]),
		}
	}
	eigene := []map[string]any{}
	if me >= 0 {
		for _, k := range g.Spieler[me].Karten {
			eigene = append(eigene, map[string]any{"wert": k.Wert, "offen": k.Offen, "weg": k.Weg})
		}
	}
	var oben any
	if len(g.Ablage) > 0 {
		oben = g.Ablage[len(g.Ablage)-1]
	}
	return map[string]any{
		"id": g.ID, "hostId": g.HostID, "status": g.Status, "phase": g.Phase,
		"runde": g.Runde, "zug": g.Zug, "zugId": g.Spieler[g.Zug].ID, "zugName": g.name(g.Zug),
		"spieler": spieler, "deckCount": len(g.Deck), "ablageOben": oben,
		"gezogen": g.Gezogen, "quelle": g.Quelle, "finisher": g.Finisher,
		"sieger": g.Sieger, "notiz": g.Notiz, "letzteRunde": g.LetzteRunde,
		"eigeneKarten": eigene, "ich": me, "sitze": g.Sitze,
	}
}

func skyjoHandler(w http.ResponseWriter, r *http.Request) {
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
		if g := room.Skyjo; g != nil && g.indexVon(id) < 0 && g.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		if room.Skyjo != nil {
			skyjoBotTreiben(room.Skyjo, time.Now())
		}
		sicht := skyjoSicht(room.Skyjo, id)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": sicht})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID     string `json:"id"`
		Index  int    `json:"index"`
		Source string `json:"source"`
		Stufe  string `json:"stufe"`
		Bots   int    `json:"bots"`
		Sitze  int    `json:"spieler"`
		Anzahl int    `json:"anzahl"`
		Max    int    `json:"maxPlayers"`
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

	switch action {
	case "bot":
		// Ein Mensch gegen mehrere Bots, direkt startklar.
		room.botNamenNeu()
		if room.Skyjo != nil && room.Skyjo.Status == "active" {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Skyjo-Partie.", http.StatusConflict)
			return
		}
		bots := input.Bots
		if bots < 1 {
			bots = 1
		}
		if bots > skyjoMax-1 {
			bots = skyjoMax - 1
		}
		g := skyjoNeu(randomCode(), input.ID, player.Name, bots+1)
		for i := 1; i <= bots; i++ {
			g.Spieler = append(g.Spieler, SkyjoSpieler{
				ID: botID(i), Name: room.botName(i), Bot: true, Stufe: botStufe(input.Stufe),
			})
		}
		g.Status = "active"
		skyjoAusteilen(g)
		g.Notiz = player.Name + " spielt gegen " + itoa(bots) + " Bots. Decke zwei Karten auf."
		room.Skyjo = g

	case "create", "join":
		g := room.Skyjo
		if g == nil || g.Status == "finished" {
			g = skyjoNeu(randomCode(), input.ID, player.Name, 2)
			room.Skyjo = g
		}
		if g.indexVon(input.ID) < 0 {
			if len(g.Spieler) >= skyjoMax {
				mu.Unlock()
				http.Error(w, "Der Tisch ist voll.", http.StatusConflict)
				return
			}
			g.Spieler = append(g.Spieler, SkyjoSpieler{ID: input.ID, Name: player.Name})
			g.Notiz = player.Name + " sitzt jetzt mit am Tisch."
		}
		if g.Status == "active" {
			mu.Unlock()
			http.Error(w, "Die Partie läuft schon.", http.StatusConflict)
			return
		}

	case "start":
		g := room.Skyjo
		if g == nil || g.indexVon(input.ID) != 0 {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber kann starten.", http.StatusForbidden)
			return
		}
		if g.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Die Partie läuft schon.", http.StatusConflict)
			return
		}
		ziel := input.Sitze
		if ziel == 0 {
			ziel = input.Anzahl
		}
		if ziel == 0 {
			ziel = input.Max
		}
		if ziel < len(g.Spieler) {
			ziel = len(g.Spieler)
		}
		if ziel < 2 {
			ziel = 2
		}
		if ziel > skyjoMax {
			ziel = skyjoMax
		}
		room.botNamenNeu()
		for len(g.Spieler) < ziel {
			i := len(g.Spieler)
			g.Spieler = append(g.Spieler, SkyjoSpieler{
				ID: botID(i), Name: room.botName(i), Bot: true, Stufe: botStufe(input.Stufe),
			})
		}
		g.Sitze = ziel
		g.Status = "active"
		skyjoAusteilen(g)
		g.Notiz = "Los geht's – jeder deckt zwei Karten auf."

	case "leave":
		g := room.Skyjo
		if g == nil {
			mu.Unlock()
			http.Error(w, "Keine Partie am Tisch.", http.StatusConflict)
			return
		}
		if g.Status == "active" {
			mu.Unlock()
			http.Error(w, "Während der Partie kann niemand aufstehen.", http.StatusConflict)
			return
		}
		for i := range g.Spieler {
			if g.Spieler[i].ID == input.ID {
				g.Spieler = append(g.Spieler[:i], g.Spieler[i+1:]...)
				break
			}
		}
		if len(g.Spieler) == 0 {
			room.Skyjo = nil
		}

	case "ready":
		// Bereitschaft ist im Raum nur ein Zeichen; Skyjo startet der Gastgeber.

	case "reveal", "draw", "swap", "discard", "opencard", "nextround", "reset":
		g := room.Skyjo
		if g == nil || (action != "reset" && g.Status != "active") {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		me := g.indexVon(input.ID)
		if me < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if action == "reset" {
			if g.Status != "finished" {
				mu.Unlock()
				http.Error(w, "Ein neues Spiel ist erst nach Spielende möglich.", http.StatusConflict)
				return
			}
			frisch := skyjoNeu(g.ID, g.HostID, g.name(0), g.Sitze)
			frisch.Spieler = make([]SkyjoSpieler, len(g.Spieler))
			for i, s := range g.Spieler {
				frisch.Spieler[i] = SkyjoSpieler{ID: s.ID, Name: s.Name, Bot: s.Bot, Stufe: s.Stufe}
			}
			frisch.Status = "active"
			skyjoAusteilen(frisch)
			frisch.Notiz = "Neue Partie – jeder deckt zwei Karten auf."
			room.Skyjo = frisch
			break
		}
		if status, meldung := g.skyjoAktion(me, action, input.Index, input.Source); status != 0 {
			mu.Unlock()
			http.Error(w, meldung, status)
			return
		}

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Skyjo-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	sicht := skyjoSicht(room.Skyjo, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": sicht})
}

// skyjoRemoveSeat nimmt einen Platz aus einem offenen Tisch. Läuft die Partie
// schon, endet sie für alle – mitten im Spiel kann niemand aufstehen, sonst
// wären Karten und Zugfolge nicht mehr nachvollziehbar.
func skyjoRemoveSeat(g *SkyjoGame, id, name string) (bool, bool, int) {
	if g == nil {
		return false, false, 0
	}
	i := g.indexVon(id)
	if i < 0 {
		return false, false, 0
	}
	if g.Status == "active" {
		return true, true, 0
	}
	g.Spieler = append(g.Spieler[:i], g.Spieler[i+1:]...)
	if len(g.Spieler) == 0 {
		return true, true, 0
	}
	return true, false, len(g.Spieler)
}
