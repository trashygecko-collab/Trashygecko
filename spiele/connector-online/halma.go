package main

// Halma im Spielraum: klassisch, Sternhalma und Sprunghalma.
//
// Der Server würfelt nicht, er rechnet – jeder Zug wird hier nach den Regeln
// geprüft. Die Brettgeometrie steht doppelt: in js/brett.js für die Anzeige und
// hier für die verbindliche Prüfung. Beide bauen die Felder aus derselben
// Konstruktion (16×16 mit Eckdreiecken, Hexgitter mit Radius 4 und sechs
// Zacken), damit Anzeige und Regel nicht auseinanderlaufen.

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	halmaKlassisch = "klassisch"
	halmaStern     = "stern"
	halmaSprung    = "sprung"
)

type HalmaBrett struct {
	Art        string `json:"art"`
	Plaetze    []int  `json:"plaetze"`
	zonen      [][]int
	Gegenseite []int `json:"gegenseite"`
	Anzahl     int   `json:"felder"`
	N          int   `json:"n"`
	Figuren    int   `json:"figuren"`

	x, y       []int
	nachbar    [][]int
	sprung     [][][2]int
	index      map[string]int
	richtungen [][2]int
}

type HalmaZug struct {
	Von     int    `json:"von"`
	Nach    int    `json:"nach"`
	Weg     []int  `json:"weg"`
	Art     string `json:"art"`
	Spieler int    `json:"spieler"`
}

// HalmaSpieler ist ein Sitz am Tisch. Platz ist die Ecke des Bretts und
// bleibt dem Spieler erhalten; die Reihenfolge in HalmaGame.Spieler ist nur
// die Zugreihenfolge. Beides fällt auseinander, sobald jemand die Partie
// verlässt – deshalb tragen die Figuren in der Belegung den Platz, nicht die
// Listenposition.
type HalmaSpieler struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Platz  int    `json:"platz"`
	Art    string `json:"art"`
	Bereit bool   `json:"bereit"`
	ImZiel int    `json:"imZiel"`
}

type HalmaGame struct {
	ID       string         `json:"id"`
	HostID   string         `json:"hostId"`
	Status   string         `json:"status"` // invited, active, finished
	Variante string         `json:"variante"`
	Anzahl   int            `json:"anzahl"`
	Brett    string         `json:"brett"`
	Spieler  []HalmaSpieler `json:"spieler"`
	Plaetze  []int          `json:"plaetze"`
	Belegung []int          `json:"belegung"`
	Current  int            `json:"current"`
	ZugNr    int            `json:"zugNr"`
	Sieger   int            `json:"sieger"`
	Letzter  *HalmaZug      `json:"letzter,omitempty"`
	Verlauf  []string       `json:"verlauf"`
	Note     string         `json:"note,omitempty"`

	brett *HalmaBrett
}

var halmaR8 = [][2]int{{1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}, {0, -1}, {1, -1}}
var halmaR6 = [][2]int{{1, 0}, {0, 1}, {-1, 1}, {-1, 0}, {0, -1}, {1, -1}}

// ---------------------------------------------------------------- Brett

// halmaBrettBauen baut das Brett. anzahl ist die Spielerzahl: beim klassischen
// Brett bestimmt sie die Größe des Eckhauses (zu zweit 19 Felder, zu viert 13).
func halmaBrettBauen(art string, anzahl int) *HalmaBrett {
	b := &HalmaBrett{Art: art, Gegenseite: []int{3, 2, 1, 0}, Figuren: 10, index: map[string]int{}}
	if art == halmaStern {
		b.Gegenseite = []int{3, 4, 5, 0, 1, 2}
		b.N = 8
		b.richtungen = halmaR6
		b.zonen = make([][]int, 6)
		add := func(q, r, zone int) {
			k := strconv.Itoa(q) + ":" + strconv.Itoa(r)
			if _, da := b.index[k]; da {
				return
			}
			b.index[k] = len(b.x)
			b.x = append(b.x, q)
			b.y = append(b.y, r)
			if zone >= 0 {
				b.zonen[zone] = append(b.zonen[zone], len(b.x)-1)
			}
		}
		for q := -4; q <= 4; q++ {
			for r := -4; r <= 4; r++ {
				m := halmaMax3(halmaAbs(q), halmaAbs(r), halmaAbs(q+r))
				if m <= 4 {
					add(q, r, -1)
				}
			}
		}
		for d := 0; d < 6; d++ {
			bx, by := halmaR6[d][0], halmaR6[d][1]
			ux, uy := halmaR6[(d+1)%6][0], halmaR6[(d+1)%6][1]
			for i := 0; i < 4; i++ {
				for j := 0; j <= 3-i; j++ {
					add(5*bx+i*bx+j*ux, 5*by+i*by+j*uy, d)
				}
			}
		}
	} else {
		b.N = 16
		b.richtungen = halmaR8
		b.zonen = make([][]int, 4)
		// Eckhaus, von der Ecke aus gezählt: Reihen 1-2-3-4, die nächste Reihe
		// ist um die beiden Eckfelder gekürzt – zu zweit eine Reihe mehr
		// (1-2-3-4-5-4 = 19), zu viert 1-2-3-4-3 = 13.
		gross := anzahl == 0 || anzahl == 2
		imHaus := func(a, c int) bool {
			s := a + c
			if s <= 3 {
				return true
			}
			if gross {
				return s == 4 || (s == 5 && a >= 1 && a <= 4)
			}
			return s == 4 && a >= 1 && a <= 3
		}
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				b.x = append(b.x, x)
				b.y = append(b.y, y)
				i := y*16 + x
				if imHaus(x, y) {
					b.zonen[0] = append(b.zonen[0], i)
				}
				if imHaus(15-x, y) {
					b.zonen[1] = append(b.zonen[1], i)
				}
				if imHaus(x, 15-y) {
					b.zonen[2] = append(b.zonen[2], i)
				}
				if imHaus(15-x, 15-y) {
					b.zonen[3] = append(b.zonen[3], i)
				}
			}
		}
		b.Figuren = 13
		if gross {
			b.Figuren = 19
		}
	}
	b.Anzahl = len(b.x)
	b.nachbar = make([][]int, b.Anzahl)
	b.sprung = make([][][2]int, b.Anzahl)
	for i := 0; i < b.Anzahl; i++ {
		for _, d := range b.richtungen {
			j := b.schritt(i, d[0], d[1], 1)
			if j < 0 {
				continue
			}
			b.nachbar[i] = append(b.nachbar[i], j)
			if k := b.schritt(i, d[0], d[1], 2); k >= 0 {
				b.sprung[i] = append(b.sprung[i], [2]int{j, k})
			}
		}
	}
	return b
}

func (b *HalmaBrett) schritt(i, dx, dy, weite int) int {
	if b.Art == halmaStern {
		k := strconv.Itoa(b.x[i]+dx*weite) + ":" + strconv.Itoa(b.y[i]+dy*weite)
		if v, ok := b.index[k]; ok {
			return v
		}
		return -1
	}
	x, y := b.x[i]+dx*weite, b.y[i]+dy*weite
	if x < 0 || y < 0 || x > 15 || y > 15 {
		return -1
	}
	return y*16 + x
}

func (b *HalmaBrett) zielZone(platz int) []int {
	return b.zonen[b.Gegenseite[platz]]
}

func halmaMax3(a, b, c int) int {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}

func halmaAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// ------------------------------------------------------------- Regelwerk

// halmaStartPlaetze sagt, welche Startbereiche die Spieler bekommen – die
// Reihenfolge ist die Platznummer. Die erlaubten Spielerzahlen stehen nur hier,
// damit Liste und Aufstellung nicht auseinanderlaufen können.
var halmaStartPlaetze = map[string]map[int][]int{
	halmaKlassisch: {2: {0, 3}, 4: {0, 1, 2, 3}},
	halmaStern:     {2: {0, 3}, 3: {0, 2, 4}, 4: {0, 1, 3, 4}, 6: {0, 1, 2, 3, 4, 5}},
}

func halmaSpielerzahlen(variante string) []int {
	t := halmaStartPlaetze[halmaBrettArt(variante)]
	out := make([]int, 0, len(t))
	for n := range t {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

func halmaPlaetze(variante string, anzahl int) []int {
	t := halmaStartPlaetze[halmaBrettArt(variante)]
	p, ok := t[anzahl]
	if !ok {
		p = t[2]
	}
	// Kopie: Partie und Neustart halten ihre eigene Liste.
	return append([]int{}, p...)
}

func halmaEnthaelt(zahlen []int, gesucht int) bool {
	for _, n := range zahlen {
		if n == gesucht {
			return true
		}
	}
	return false
}

func halmaZahlenText(zahlen []int) string {
	teile := make([]string, 0, len(zahlen))
	for _, n := range zahlen {
		teile = append(teile, strconv.Itoa(n))
	}
	return strings.Join(teile, ", ")
}

func halmaBrettArt(variante string) string {
	if variante == halmaStern {
		return halmaStern
	}
	return halmaKlassisch
}

// halmaSprungFelder liefert alle Felder, die auf feld mit Sprüngen erreichbar
// sind, samt der übersprungenen Felder. Die eigene Figur steht dabei nicht im Weg.
func halmaSprungFelder(b *HalmaBrett, belegung []int, feld int, weit bool) map[int][]int {
	weg := map[int][]int{}
	belegt := make([]int, len(belegung))
	copy(belegt, belegung)
	belegt[feld] = -1
	gesehen := map[int]bool{feld: true}
	rand := []int{feld}
	for len(rand) > 0 {
		var naechste []int
		for _, f := range rand {
			var zuege [][3]int // ueber, ziel, dummy
			if weit {
				for _, d := range b.richtungen {
					for w := 1; w < 16; w++ {
						ueber := b.schritt(f, d[0], d[1], w)
						if ueber < 0 {
							break
						}
						if belegt[ueber] < 0 {
							continue
						}
						ziel := b.schritt(f, d[0], d[1], 2*w)
						if ziel < 0 {
							break
						}
						if belegt[ziel] < 0 {
							zuege = append(zuege, [3]int{ueber, ziel, 0})
						}
					}
				}
			} else {
				for _, p := range b.sprung[f] {
					if belegt[p[0]] >= 0 {
						zuege = append(zuege, [3]int{p[0], p[1], 0})
					}
				}
			}
			for _, z := range zuege {
				ziel := z[1]
				if belegt[ziel] >= 0 || gesehen[ziel] {
					continue
				}
				gesehen[ziel] = true
				weg[ziel] = append(append([]int{}, weg[f]...), z[0])
				belegt[ziel] = belegung[feld]
				naechste = append(naechste, ziel)
			}
		}
		rand = naechste
	}
	return weg
}

func halmaZuegeVon(b *HalmaBrett, belegung []int, variante string, feld int) []HalmaZug {
	var out []HalmaZug
	for _, n := range b.nachbar[feld] {
		if belegung[n] < 0 {
			out = append(out, HalmaZug{Von: feld, Nach: n, Weg: []int{}, Art: "schritt"})
		}
	}
	for ziel, weg := range halmaSprungFelder(b, belegung, feld, variante == halmaSprung) {
		out = append(out, HalmaZug{Von: feld, Nach: ziel, Weg: weg, Art: "sprung"})
	}
	return out
}

func halmaZugPruefen(b *HalmaBrett, belegung []int, variante string, von, nach int) (HalmaZug, bool) {
	if von < 0 || nach < 0 || von >= len(belegung) || nach >= len(belegung) || von == nach {
		return HalmaZug{}, false
	}
	if belegung[von] < 0 || belegung[nach] >= 0 {
		return HalmaZug{}, false
	}
	for _, z := range halmaZuegeVon(b, belegung, variante, von) {
		if z.Nach == nach {
			return z, true
		}
	}
	return HalmaZug{}, false
}

func halmaImZiel(b *HalmaBrett, belegung []int, platz int) int {
	n := 0
	for _, i := range b.zielZone(platz) {
		if belegung[i] == platz {
			n++
		}
	}
	return n
}

// halmaImZielAlle zählt für jeden Platz die eigenen Figuren in der Zielzone –
// ein Durchgang für alle Spieler einer Sicht.
func halmaImZielAlle(b *HalmaBrett, belegung []int) []int {
	n := make([]int, len(b.Gegenseite))
	for platz := range n {
		for _, i := range b.zonen[b.Gegenseite[platz]] {
			if belegung[i] == platz {
				n[platz]++
			}
		}
	}
	return n
}

func halmaSieg(b *HalmaBrett, belegung []int, platz int) bool {
	return halmaImZiel(b, belegung, platz) == len(b.zielZone(platz))
}

// halmaAufstellung stellt die Figuren in die angegebenen Häuser. Der Wert in
// der Belegung ist der Platz (die Ecke) selbst – nicht die Sitzreihenfolge.
// Dadurch bleiben Figuren und Spieler auch dann beieinander, wenn jemand
// ausscheidet und die Zugreihenfolge kürzer wird.
func halmaAufstellung(b *HalmaBrett, plaetze []int) []int {
	belegung := make([]int, b.Anzahl)
	for i := range belegung {
		belegung[i] = -1
	}
	for _, p := range plaetze {
		for _, i := range b.zonen[p] {
			belegung[i] = p
		}
	}
	return belegung
}

func halmaStartAufstellung(b *HalmaBrett, variante string, anzahl int) ([]int, []int) {
	plaetze := halmaPlaetze(variante, anzahl)
	return halmaAufstellung(b, plaetze), plaetze
}

// halmaSitzplaetze liefert die Häuser der Spieler in Zugreihenfolge.
func halmaSitzplaetze(g *HalmaGame) []int {
	plaetze := make([]int, len(g.Spieler))
	for i, p := range g.Spieler {
		plaetze[i] = p.Platz
	}
	return plaetze
}

// halmaFreierPlatz sucht das erste Haus, in dem noch niemand sitzt. Ohne das
// bekäme der zweite Spieler im Zweierspiel die Nachbärecke statt der
// gegenüberliegenden.
func halmaFreierPlatz(g *HalmaGame) int {
	for _, p := range halmaPlaetze(g.Variante, g.Anzahl) {
		belegt := false
		for _, s := range g.Spieler {
			if s.Platz == p {
				belegt = true
				break
			}
		}
		if !belegt {
			return p
		}
	}
	return -1
}

// ---------------------------------------------------------------- Sicht

func halmaSicht(g *HalmaGame) *HalmaGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Spieler = make([]HalmaSpieler, len(g.Spieler))
	// Das Brett entsteht einmal je Partie (halmaBrettFuer) und der Stand im
	// Ziel einmal je Sicht – nicht je Spieler. Bei 1,5-Sekunden-Polling je
	// Client war das vorher der teuerste Teil jeder Anfrage.
	b := halmaBrettFuer(g)
	var imZiel []int
	if b != nil && len(g.Belegung) == b.Anzahl {
		imZiel = halmaImZielAlle(b, g.Belegung)
	}
	for i, p := range g.Spieler {
		c.Spieler[i] = p
		if imZiel != nil && p.Platz >= 0 && p.Platz < len(imZiel) {
			c.Spieler[i].ImZiel = imZiel[p.Platz]
		}
	}
	c.Belegung = append([]int{}, g.Belegung...)
	c.Plaetze = append([]int{}, g.Plaetze...)
	c.Verlauf = append([]string{}, g.Verlauf...)
	if g.Letzter != nil {
		l := *g.Letzter
		l.Weg = append([]int{}, g.Letzter.Weg...)
		c.Letzter = &l
	}
	return &c
}

func halmaNeu(hostID, name, variante string, anzahl int) *HalmaGame {
	b := halmaBrettBauen(halmaBrettArt(variante), anzahl)
	belegung, plaetze := halmaStartAufstellung(b, variante, anzahl)
	g := &HalmaGame{
		ID: randomCode(), HostID: hostID, Status: "invited",
		Variante: variante, Anzahl: anzahl, Brett: b.Art,
		Plaetze: plaetze, Belegung: belegung, Current: 0, ZugNr: 1, Sieger: -1,
		Verlauf: []string{}, brett: b,
	}
	g.Spieler = []HalmaSpieler{{ID: hostID, Name: name, Platz: 0, Art: "mensch"}}
	g.Note = name + " eröffnet " + halmaName(variante) + " für " + strconv.Itoa(anzahl) + " Spieler."
	return g
}

func halmaName(variante string) string {
	switch variante {
	case halmaStern:
		return "Sternhalma"
	case halmaSprung:
		return "Sprunghalma"
	}
	return "Klassisches Halma"
}

func halmaBrettFuer(g *HalmaGame) *HalmaBrett {
	if g.brett == nil {
		g.brett = halmaBrettBauen(g.Brett, g.Anzahl)
	}
	return g.brett
}

// --------------------------------------------------------------- Handler

func halmaHandler(w http.ResponseWriter, r *http.Request) {
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
		sicht := halmaSicht(room.Halma)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": sicht})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID       string `json:"id"`
		Variante string `json:"variante"`
		Anzahl   int    `json:"anzahl"`
		Von      int    `json:"von"`
		Nach     int    `json:"nach"`
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
	fehler := func(status int, text string) {
		mu.Unlock()
		http.Error(w, text, status)
	}

	switch action {
	case "create":
		if room.Halma != nil && room.Halma.Status != "finished" {
			fehler(http.StatusConflict, "In diesem Spielraum läuft bereits eine Halma-Partie.")
			return
		}
		variante := input.Variante
		if variante != halmaKlassisch && variante != halmaStern && variante != halmaSprung {
			fehler(http.StatusBadRequest, "Bitte klassisch, stern oder sprung wählen.")
			return
		}
		anzahl := input.Anzahl
		zahlen := halmaSpielerzahlen(variante)
		if !halmaEnthaelt(zahlen, anzahl) {
			fehler(http.StatusBadRequest, "Für "+halmaName(variante)+" sind "+
				halmaZahlenText(zahlen)+" Spieler erlaubt.")
			return
		}
		room.Halma = halmaNeu(input.ID, player.Name, variante, anzahl)
		room.addMessage(Message{
			Name: player.Name,
			Text: "eröffnet " + halmaName(variante) + " für " + strconv.Itoa(anzahl) + " Spieler.",
			Time: time.Now().Format("15:04:05"), SenderID: input.ID,
		})

	// join und rejoin tun dasselbe: Platz nehmen, solange einer frei ist.
	// join nur in der offenen Lobby, rejoin auch in einer laufenden Partie –
	// der Client ruft ihn bei jedem Seitenaufbau auf, ohne zu wissen, ob er
	// schon sitzt; ein bereits besetzter eigener Platz ist kein Fehler.
	case "join", "rejoin":
		g := room.Halma
		if g == nil || g.Status == "finished" || (action == "join" && g.Status != "invited") {
			fehler(http.StatusConflict, "Es läuft gerade keine offene Halma-Partie.")
			return
		}
		if halmaPlatzVon(g, input.ID) >= 0 {
			break
		}
		if len(g.Spieler) >= g.Anzahl {
			fehler(http.StatusConflict, "Diese Partie ist schon voll.")
			return
		}
		frei := halmaFreierPlatz(g)
		if frei < 0 {
			fehler(http.StatusConflict, "Diese Partie ist schon voll.")
			return
		}
		g.Spieler = append(g.Spieler, HalmaSpieler{ID: input.ID, Name: player.Name, Platz: frei, Art: "mensch"})
		g.Plaetze = halmaSitzplaetze(g)
		if action == "join" {
			g.Note = player.Name + " spielt mit."
		} else {
			g.Note = player.Name + " ist wieder dabei."
		}

	case "ready":
		g := room.Halma
		if g == nil || g.Status != "invited" {
			fehler(http.StatusConflict, "Bereit drücken geht nur vor dem Start.")
			return
		}
		k := halmaPlatzVon(g, input.ID)
		if k < 0 {
			fehler(http.StatusForbidden, "Du sitzt nicht an diesem Tisch.")
			return
		}
		g.Spieler[k].Bereit = !g.Spieler[k].Bereit

	case "start":
		g := room.Halma
		if g == nil || g.Status != "invited" {
			fehler(http.StatusConflict, "Es wartet keine Partie auf den Start.")
			return
		}
		if g.HostID != input.ID {
			fehler(http.StatusForbidden, "Nur der Gastgeber startet die Partie.")
			return
		}
		if len(g.Spieler) < 2 {
			fehler(http.StatusConflict, "Halma braucht mindestens zwei Spieler.")
			return
		}
		for _, p := range g.Spieler {
			if !p.Bereit {
				fehler(http.StatusConflict, p.Name+" ist noch nicht bereit.")
				return
			}
		}
		// Nur die Häuser der tatsächlich Sitzenden besetzen: wer nicht
		// mitgekommen ist, hat auch keine Figuren auf dem Brett.
		b := halmaBrettFuer(g)
		g.Plaetze = halmaSitzplaetze(g)
		g.Belegung = halmaAufstellung(b, g.Plaetze)
		g.Status = "active"
		g.ZugNr = 1
		g.Current = 0
		g.Sieger = -1
		g.Note = "Los geht's – " + g.Spieler[0].Name + " beginnt."

	case "zug":
		g := room.Halma
		if g == nil || g.Status != "active" {
			fehler(http.StatusConflict, "Es läuft gerade keine aktive Partie.")
			return
		}
		k := halmaPlatzVon(g, input.ID)
		if k < 0 {
			fehler(http.StatusForbidden, "Du gehörst nicht zu dieser Partie.")
			return
		}
		if g.Current != k {
			fehler(http.StatusConflict, "Jetzt ist "+g.Spieler[g.Current].Name+" am Zug.")
			return
		}
		b := halmaBrettFuer(g)
		platz := g.Spieler[k].Platz
		if input.Von < 0 || input.Von >= len(g.Belegung) || g.Belegung[input.Von] != platz {
			fehler(http.StatusConflict, "Das ist nicht deine Figur.")
			return
		}
		zug, gut := halmaZugPruefen(b, g.Belegung, g.Variante, input.Von, input.Nach)
		if !gut {
			fehler(http.StatusConflict, "Dieser Zug ist nach den Regeln nicht erlaubt.")
			return
		}
		zug.Spieler = k
		g.Belegung[zug.Nach] = platz
		g.Belegung[zug.Von] = -1
		g.Letzter = &HalmaZug{Von: zug.Von, Nach: zug.Nach, Weg: zug.Weg, Art: zug.Art, Spieler: k}
		g.ZugNr++
		text := g.Spieler[k].Name + ": "
		if zug.Art == "sprung" {
			text += "Sprung"
			if len(zug.Weg) > 1 {
				text += "kette über " + strconv.Itoa(len(zug.Weg)) + " Figuren"
			}
		} else {
			text += "ein Schritt"
		}
		g.Verlauf = append(g.Verlauf, text+feldernamen(b, zug.Nach))
		if len(g.Verlauf) > 40 {
			g.Verlauf = g.Verlauf[len(g.Verlauf)-40:]
		}
		if halmaSieg(b, g.Belegung, platz) {
			g.Status = "finished"
			g.Sieger = k
			g.Note = g.Spieler[k].Name + " hat alle Figuren im Ziel."
		} else {
			g.Current = (g.Current + 1) % len(g.Spieler)
			g.Note = g.Spieler[g.Current].Name + " ist am Zug."
		}

	case "reset":
		g := room.Halma
		if g == nil {
			fehler(http.StatusConflict, "Hier läuft keine Halma-Partie.")
			return
		}
		k := halmaPlatzVon(g, input.ID)
		if k < 0 {
			fehler(http.StatusForbidden, "Du gehörst nicht zu dieser Partie.")
			return
		}
		b := halmaBrettFuer(g)
		g.Plaetze = halmaSitzplaetze(g)
		g.Belegung = halmaAufstellung(b, g.Plaetze)
		g.Status = "active"
		g.Current = 0
		g.ZugNr = 1
		g.Sieger = -1
		g.Letzter = nil
		g.Verlauf = []string{}
		g.Note = "Neue Partie – " + g.Spieler[0].Name + " beginnt."

	default:
		fehler(http.StatusBadRequest, "Unbekannte Aktion.")
		return
	}

	sicht := halmaSicht(room.Halma)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": sicht})
}

// halmaRaus nimmt einen Spieler vom Tisch (Verbindung weg oder verlassen).
//
// Die Plätze sind die Ecken des Bretts; sie bleiben beim Ausscheiden stehen.
// Die Figuren des Gehenden verschwinden vom Brett, alle anderen behalten
// ihren Wert – sonst stünden Figuren in fremden Häusern, wechselten beim
// Weiterzählen den Besitzer und die Zielzonen passten nicht mehr zu den
// Spielern (Geisterfiguren).
func halmaRaus(g *HalmaGame, id, name string) (bool, bool, int) {
	k := halmaPlatzVon(g, id)
	if k < 0 {
		return false, false, len(g.Spieler)
	}
	platz := g.Spieler[k].Platz
	for i := range g.Belegung {
		if g.Belegung[i] == platz {
			g.Belegung[i] = -1
		}
	}
	g.Spieler = append(g.Spieler[:k], g.Spieler[k+1:]...)
	g.Plaetze = halmaSitzplaetze(g)
	// Der letzte Zug kann vom Gehenden stammen – Zugnummer und Figuren sind
	// danach nicht mehr dieselben.
	g.Letzter = nil
	if len(g.Spieler) < 2 {
		return true, true, len(g.Spieler)
	}
	if g.HostID == id {
		g.HostID = g.Spieler[0].ID
	}
	if g.Status == "active" {
		if g.Current > k {
			g.Current--
		}
		if g.Current >= len(g.Spieler) {
			g.Current = 0
		}
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Spieler)) + " Spielern weiter."
	return true, false, len(g.Spieler)
}

func halmaPlatzVon(g *HalmaGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func feldernamen(b *HalmaBrett, i int) string {
	if i < 0 || i >= b.Anzahl {
		return ""
	}
	if b.Art == halmaStern {
		return " auf (" + strconv.Itoa(b.x[i]) + "/" + strconv.Itoa(b.y[i]) + ")"
	}
	return " auf " + string(rune('A'+b.x[i])) + strconv.Itoa(b.y[i]+1)
}
