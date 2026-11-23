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

// Cabo nach den Regeln von Smiling Monster: 4 Karten vor jedem, zwei davon
// am Rundenstart anschauen, dann reihum tauschen, ziehen oder Cabo ansagen.
// Wer am Ende die wenigsten Punkte hat, gewinnt die Runde.

const (
	caboMaxSpieler = 4
	caboKarten     = 4
	caboStartBlick = 2
	caboZiel       = 101
)

type CaboSpieler struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Karten   []int          `json:"-"`
	Bekannt  map[string]int `json:"-"`
	Punkte   int            `json:"punkte"`
	Runden   int            `json:"runden"`
	Ready    bool           `json:"ready"`
	Geschaut int            `json:"geschaut"`
	Bot      bool           `json:"bot,omitempty"`
	Stufe    string         `json:"stufe,omitempty"`
	Zuege    int            `json:"-"`
}

type CaboGame struct {
	ID           string        `json:"id"`
	HostID       string        `json:"hostId"`
	Status       string        `json:"status"` // invited, anschauen, active, wertung, ende
	Runde        int           `json:"runde"`
	Zug          int           `json:"zug"`
	Start        int           `json:"start"`
	Spieler      []CaboSpieler `json:"spieler"`
	Stapel       []int         `json:"-"`
	Ablage       []int         `json:"ablage"`
	Gezogen      *int          `json:"gezogen"`
	Quelle       string        `json:"quelle"`
	CaboVon      int           `json:"caboVon"`
	RestZuege    int           `json:"restZuege"`
	Note         string        `json:"note"`
	Sieger       string        `json:"sieger,omitempty"`
	RundenSieger int           `json:"rundenSieger"`
	RundenPunkte []int         `json:"rundenPunkte"`
	BotMoment    time.Time     `json:"-"`
}

type CaboSpielerSicht struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Anzahl   int            `json:"anzahl"`
	Bekannt  map[string]int `json:"bekannt"`
	Punkte   int            `json:"punkte"`
	Runden   int            `json:"runden"`
	Ready    bool           `json:"ready"`
	Geschaut int            `json:"geschaut"`
	Werte    []int          `json:"werte,omitempty"`
	Bot      bool           `json:"bot,omitempty"`
	Stufe    string         `json:"stufe,omitempty"`
}

type CaboAnsicht struct {
	ID           string             `json:"id"`
	HostID       string             `json:"hostId"`
	Status       string             `json:"status"`
	Runde        int                `json:"runde"`
	Zug          int                `json:"zug"`
	Start        int                `json:"start"`
	Spieler      []CaboSpielerSicht `json:"spieler"`
	AblageOben   *int               `json:"ablageOben"`
	AblageAnzahl int                `json:"ablageAnzahl"`
	Stapel       int                `json:"stapel"`
	Gezogen      *int               `json:"gezogen"`
	Quelle       string             `json:"quelle"`
	CaboVon      int                `json:"caboVon"`
	RestZuege    int                `json:"restZuege"`
	Note         string             `json:"note"`
	Sieger       string             `json:"sieger,omitempty"`
	RundenPunkte []int              `json:"rundenPunkte"`
}

// ---------------------------------------------------------------- Aufbau

func caboDeck() []int {
	deck := []int{0, 0}
	for wert := 1; wert <= 12; wert++ {
		for i := 0; i < 4; i++ {
			deck = append(deck, wert)
		}
	}
	deck = append(deck, 13, 13)
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

func caboNeuerSpieler(id, name string) CaboSpieler {
	return CaboSpieler{ID: id, Name: name, Karten: []int{}, Bekannt: map[string]int{}}
}

func caboNeueRunde(g *CaboGame) {
	deck := caboDeck()
	for i := range g.Spieler {
		p := &g.Spieler[i]
		p.Karten = make([]int, 0, caboKarten)
		for k := 0; k < caboKarten && len(deck) > 0; k++ {
			p.Karten = append(p.Karten, deck[0])
			deck = deck[1:]
		}
		p.Bekannt = map[string]int{}
		p.Geschaut = 0
		p.Zuege = 0
	}
	g.Stapel = deck
	g.Ablage = []int{g.Stapel[0]}
	g.Stapel = g.Stapel[1:]
	g.Gezogen = nil
	g.Quelle = ""
	g.CaboVon = -1
	g.RestZuege = 0
	g.RundenPunkte = nil
	g.Zug = g.Start
	g.Runde++
	g.Status = "anschauen"
	g.BotMoment = time.Now()
	g.Note = "Runde " + strconv.Itoa(g.Runde) + ": Jeder schaut sich zwei eigene Karten an."
}

// ---------------------------------------------------------------- Sätze

func caboHatWert(karten []int, wert int) int {
	n := 0
	for _, k := range karten {
		if k == wert {
			n++
		}
	}
	return n
}

func caboSatz(g *CaboGame, sitz int, ziele []int) bool {
	p := &g.Spieler[sitz]
	if len(ziele) < 2 || len(ziele) > len(p.Karten) {
		return false
	}
	gesehen := map[int]bool{}
	for _, pos := range ziele {
		if pos < 0 || pos >= len(p.Karten) || gesehen[pos] {
			return false
		}
		gesehen[pos] = true
	}
	wert := p.Karten[ziele[0]]
	for _, pos := range ziele {
		if p.Karten[pos] != wert {
			return false
		}
	}
	sort.Ints(ziele)
	// Satz abwerfen, neue Karte an die erste frei gewordene Stelle.
	for _, pos := range ziele {
		g.Ablage = append(g.Ablage, p.Karten[pos])
	}
	neu := make([]int, 0, len(p.Karten)-len(ziele)+1)
	neuBekannt := map[string]int{}
	zielIt := 0
	eingesetzt := false
	for pos, karte := range p.Karten {
		if zielIt < len(ziele) && ziele[zielIt] == pos {
			if !eingesetzt {
				neu = append(neu, *g.Gezogen)
				neuBekannt[caboSchluessel(sitz, len(neu)-1)] = *g.Gezogen
				eingesetzt = true
			}
			zielIt++
			continue
		}
		neu = append(neu, karte)
		if wert, ok := p.Bekannt[caboSchluessel(sitz, pos)]; ok {
			neuBekannt[caboSchluessel(sitz, len(neu)-1)] = wert
		}
	}
	p.Karten = neu
	p.Bekannt = neuBekannt
	return true
}

func caboSchluessel(sitz, pos int) string {
	return strconv.Itoa(sitz) + ":" + strconv.Itoa(pos)
}

// ---------------------------------------------------------------- Zug

func caboNachziehen(g *CaboGame) {
	if len(g.Stapel) == 0 {
		// Ablage bis auf die oberste Karte neu mischen – die bleibt liegen.
		oben := g.Ablage[len(g.Ablage)-1]
		rest := append([]int{}, g.Ablage[:len(g.Ablage)-1]...)
		rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
		g.Ablage = []int{oben}
		g.Stapel = rest
	}
}

// caboZugWeiter beendet den Zug. Liegt noch eine gezogene Karte da, wandert
// sie offen auf den Ablagestapel.
func caboZugWeiter(g *CaboGame) {
	if g.Gezogen != nil {
		g.Ablage = append(g.Ablage, *g.Gezogen)
	}
	g.Gezogen = nil
	g.Quelle = ""
	if g.Status != "active" {
		return
	}
	if g.CaboVon >= 0 {
		g.RestZuege--
		if g.RestZuege <= 0 {
			caboWertung(g)
			return
		}
	}
	g.Zug = (g.Zug + 1) % len(g.Spieler)
	if g.Spieler[g.Zug].ID == "" {
		g.Zug = 0
	}
	if g.Spieler[g.Zug].Bot {
		g.BotMoment = time.Now()
	}
}

func caboWertung(g *CaboGame) {
	summen := make([]int, len(g.Spieler))
	for i := range g.Spieler {
		summe := 0
		for _, karte := range g.Spieler[i].Karten {
			summe += karte
		}
		summen[i] = summe
	}

	kamikaze := -1
	for i := range g.Spieler {
		if caboHatWert(g.Spieler[i].Karten, 12) >= 2 && caboHatWert(g.Spieler[i].Karten, 13) >= 2 {
			kamikaze = i
		}
	}

	rundenPunkte := make([]int, len(g.Spieler))
	sieger := -1

	if kamikaze >= 0 {
		for i := range g.Spieler {
			if i == kamikaze {
				rundenPunkte[i] = 0
			} else {
				rundenPunkte[i] = 50
			}
		}
		sieger = kamikaze
		g.Note = "Kamikaze von " + g.Spieler[kamikaze].Name + "! Beide 12 und beide 13 – 0 Punkte, alle anderen 50."
	} else {
		minimum := summen[0]
		for _, summe := range summen {
			if summe < minimum {
				minimum = summe
			}
		}
		gewinner := []int{}
		for i, summe := range summen {
			if summe == minimum {
				gewinner = append(gewinner, i)
			}
		}
		// Cabo-Ansager gewinnt einen Gleichstand für sich allein.
		if g.CaboVon >= 0 && summen[g.CaboVon] == minimum {
			gewinner = []int{g.CaboVon}
		}
		for i := range g.Spieler {
			istGewinner := false
			for _, k := range gewinner {
				if k == i {
					istGewinner = true
				}
			}
			if !istGewinner {
				rundenPunkte[i] = summen[i]
			}
		}
		if g.CaboVon >= 0 && summen[g.CaboVon] != minimum {
			rundenPunkte[g.CaboVon] += 5
		}
		sieger = gewinner[0]
		namen := []string{}
		for _, i := range gewinner {
			namen = append(namen, g.Spieler[i].Name)
		}
		g.Note = "Wertung – am wenigsten Punkte: " + strings.Join(namen, ", ") + "."
		if g.CaboVon >= 0 && summen[g.CaboVon] != minimum {
			g.Note += " " + g.Spieler[g.CaboVon].Name + " hat Cabo gerufen und bekommt 5 Punkte extra."
		}
	}

	for i := range g.Spieler {
		g.Spieler[i].Runden = rundenPunkte[i]
		g.Spieler[i].Punkte += rundenPunkte[i]
		if g.Spieler[i].Punkte == 100 {
			g.Spieler[i].Punkte = 50
		}
	}

	g.RundenPunkte = rundenPunkte
	g.RundenSieger = sieger
	g.Status = "wertung"

	// Neuer Startspieler: Rundensieger, bei Gleichstand der mit den
	// wenigsten Gesamtpunkten.
	g.Start = sieger
	best := -1
	for i := range g.Spieler {
		if g.Spieler[i].Runden != 0 {
			continue
		}
		if best < 0 || g.Spieler[i].Punkte < g.Spieler[best].Punkte {
			best = i
		}
	}
	if best >= 0 {
		g.Start = best
	}

	// Spielende bei 101 oder mehr Punkten.
	ende := false
	for i := range g.Spieler {
		if g.Spieler[i].Punkte >= caboZiel {
			ende = true
		}
	}
	if ende {
		best := g.Spieler[0].Punkte
		gewinner := 0
		for i := range g.Spieler {
			if g.Spieler[i].Punkte < best {
				best = g.Spieler[i].Punkte
				gewinner = i
			}
		}
		g.Status = "ende"
		g.Sieger = g.Spieler[gewinner].Name
		g.Note += " Spiel beendet – " + g.Sieger + " gewinnt mit " + strconv.Itoa(best) + " Punkten."
	}
}

// ---------------------------------------------------------------- Sicht

func caboAnsicht(g *CaboGame, id string) *CaboAnsicht {
	if g == nil {
		return nil
	}
	a := &CaboAnsicht{
		ID:        g.ID,
		HostID:    g.HostID,
		Status:    g.Status,
		Runde:     g.Runde,
		Zug:       g.Zug,
		Start:     g.Start,
		Spieler:   make([]CaboSpielerSicht, 0, len(g.Spieler)),
		Stapel:    len(g.Stapel),
		Quelle:    g.Quelle,
		CaboVon:   g.CaboVon,
		RestZuege: g.RestZuege,
		Note:      g.Note,
		Sieger:    g.Sieger,
	}
	if len(g.Ablage) > 0 {
		oben := g.Ablage[len(g.Ablage)-1]
		a.AblageOben = &oben
	}
	a.AblageAnzahl = len(g.Ablage)
	if g.Gezogen != nil {
		wert := *g.Gezogen
		a.Gezogen = &wert
	}
	if g.RundenPunkte != nil {
		a.RundenPunkte = append([]int{}, g.RundenPunkte...)
	}
	offen := g.Status == "wertung" || g.Status == "ende"

	for i, p := range g.Spieler {
		sicht := CaboSpielerSicht{
			ID:       p.ID,
			Name:     p.Name,
			Anzahl:   len(p.Karten),
			Punkte:   p.Punkte,
			Runden:   p.Runden,
			Ready:    p.Ready,
			Geschaut: p.Geschaut,
			Bot:      p.Bot,
			Stufe:    p.Stufe,
			Bekannt:  map[string]int{},
		}
		if p.ID == id {
			for k, v := range p.Bekannt {
				sicht.Bekannt[k] = v
			}
		}
		if offen {
			sicht.Werte = append([]int{}, p.Karten...)
		}
		_ = i
		a.Spieler = append(a.Spieler, sicht)
	}
	return a
}

// ---------------------------------------------------------------- Aktionen

type caboEingabe struct {
	Pos   int
	Pos2  int
	Ziel  int
	Ziele []int
}

type caboFehler struct {
	Text string
	Nr   int
}

func (f *caboFehler) Error() string { return f.Text }

func caboNeu(text string, nr int) *caboFehler { return &caboFehler{Text: text, Nr: nr} }

// caboPruefeAnschauen schaltet weiter, sobald jeder seine zwei Karten gesehen
// hat. Mensch und Computer laufen beide hier durch.
func caboPruefeAnschauen(g *CaboGame) {
	for _, sp := range g.Spieler {
		if sp.Geschaut < caboStartBlick {
			return
		}
	}
	g.Status = "active"
	g.Zug = g.Start
	g.Note = "Los geht's – " + g.Spieler[g.Zug].Name + " beginnt."
	if g.Spieler[g.Zug].Bot {
		g.BotMoment = time.Now()
	}
}

// caboAktion führt eine Spielaktion aus. Menschen kommen über den Endpunkt
// hierher, der Computer über caboBotSchritt – die Regeln stehen nur einmal da.
func caboAktion(g *CaboGame, sitz int, action string, in caboEingabe) *caboFehler {
	if g == nil {
		return caboNeu("Gerade läuft keine Runde.", http.StatusConflict)
	}
	if sitz < 0 || sitz >= len(g.Spieler) {
		return caboNeu("Du gehörst nicht zu dieser Runde.", http.StatusForbidden)
	}
	p := &g.Spieler[sitz]

	switch action {
	case "anschauen":
		if g.Status != "anschauen" {
			return caboNeu("Gerade wird nicht angeschaut.", http.StatusConflict)
		}
		if p.Geschaut >= caboStartBlick {
			return caboNeu("Du hast schon zwei Karten gesehen.", http.StatusConflict)
		}
		if in.Pos < 0 || in.Pos >= len(p.Karten) {
			return caboNeu("Diese Karte gibt es nicht.", http.StatusBadRequest)
		}
		scl := caboSchluessel(sitz, in.Pos)
		if _, schon := p.Bekannt[scl]; schon {
			return caboNeu("Diese Karte kennst du schon – nimm eine andere.", http.StatusConflict)
		}
		p.Bekannt[scl] = p.Karten[in.Pos]
		p.Geschaut++
		g.Note = p.Name + " schaut sich Karten an."
		caboPruefeAnschauen(g)
		return nil

	case "nehmen", "ziehen":
		if g.Status != "active" {
			return caboNeu("Gerade läuft kein Zug.", http.StatusConflict)
		}
		if sitz != g.Zug {
			return caboNeu("Du bist nicht an der Reihe.", http.StatusConflict)
		}
		if g.Gezogen != nil {
			return caboNeu("Du hast schon eine Karte gezogen.", http.StatusConflict)
		}
		if action == "nehmen" {
			if len(g.Ablage) == 0 {
				return caboNeu("Der Ablagestapel ist leer.", http.StatusConflict)
			}
			wert := g.Ablage[len(g.Ablage)-1]
			g.Ablage = g.Ablage[:len(g.Ablage)-1]
			g.Gezogen = &wert
			g.Quelle = "ablage"
			g.Note = p.Name + " nimmt die " + strconv.Itoa(wert) + " vom Ablagestapel und tauscht sie ein."
		} else {
			caboNachziehen(g)
			wert := g.Stapel[len(g.Stapel)-1]
			g.Stapel = g.Stapel[:len(g.Stapel)-1]
			g.Gezogen = &wert
			g.Quelle = "stapel"
			g.Note = p.Name + " zieht eine Karte vom Stapel."
		}
		return nil

	case "legen":
		if g.Gezogen == nil {
			return caboNeu("Du hast keine Karte in der Hand.", http.StatusConflict)
		}
		if sitz != g.Zug {
			return caboNeu("Du bist nicht an der Reihe.", http.StatusConflict)
		}
		if in.Pos < 0 || in.Pos >= len(p.Karten) {
			return caboNeu("Diese Karte gibt es nicht.", http.StatusBadRequest)
		}
		alt := p.Karten[in.Pos]
		p.Karten[in.Pos] = *g.Gezogen
		p.Bekannt[caboSchluessel(sitz, in.Pos)] = *g.Gezogen
		g.Ablage = append(g.Ablage, alt)
		g.Note = p.Name + " legt " + strconv.Itoa(*g.Gezogen) + " ab und wirft " + strconv.Itoa(alt) + " weg."
		g.Gezogen = nil
		caboZugWeiter(g)
		return nil

	case "abwerfen":
		if g.Gezogen == nil {
			return caboNeu("Du hast keine Karte in der Hand.", http.StatusConflict)
		}
		if sitz != g.Zug {
			return caboNeu("Du bist nicht an der Reihe.", http.StatusConflict)
		}
		g.Note = p.Name + " wirft " + strconv.Itoa(*g.Gezogen) + " ab."
		caboZugWeiter(g)
		return nil

	case "peek", "spy", "swap":
		if g.Gezogen == nil {
			return caboNeu("Dafür musst du zuerst eine Karte gezogen.", http.StatusConflict)
		}
		if g.Quelle != "stapel" {
			return caboNeu("Vom Ablagestapel genommene Karten haben keine Aktion.", http.StatusConflict)
		}
		if sitz != g.Zug {
			return caboNeu("Du bist nicht an der Reihe.", http.StatusConflict)
		}
		wert := *g.Gezogen
		moeglich := false
		switch action {
		case "peek":
			moeglich = wert == 7 || wert == 8
		case "spy":
			moeglich = wert == 9 || wert == 10
		case "swap":
			moeglich = wert == 11 || wert == 12
		}
		if !moeglich {
			return caboNeu("Diese Karte hat keine solche Aktion.", http.StatusConflict)
		}

		switch action {
		case "peek":
			if in.Pos < 0 || in.Pos >= len(p.Karten) {
				return caboNeu("Diese Karte gibt es nicht.", http.StatusBadRequest)
			}
			p.Bekannt[caboSchluessel(sitz, in.Pos)] = p.Karten[in.Pos]
			g.Note = p.Name + " nutzt die Aktion " + strconv.Itoa(wert) + " und schaut sich eine eigene Karte an."

		case "spy":
			if in.Ziel < 0 || in.Ziel >= len(g.Spieler) || in.Ziel == sitz {
				return caboNeu("Wähle einen Mitspieler.", http.StatusBadRequest)
			}
			z := &g.Spieler[in.Ziel]
			if in.Pos < 0 || in.Pos >= len(z.Karten) {
				return caboNeu("Diese Karte gibt es nicht.", http.StatusBadRequest)
			}
			p.Bekannt[caboSchluessel(in.Ziel, in.Pos)] = z.Karten[in.Pos]
			g.Note = p.Name + " späht bei " + z.Name + "."

		case "swap":
			if in.Ziel < 0 || in.Ziel >= len(g.Spieler) || in.Ziel == sitz {
				return caboNeu("Wähle einen Mitspieler.", http.StatusBadRequest)
			}
			z := &g.Spieler[in.Ziel]
			if in.Pos < 0 || in.Pos >= len(p.Karten) || in.Pos2 < 0 || in.Pos2 >= len(z.Karten) {
				return caboNeu("Diese Karte gibt es nicht.", http.StatusBadRequest)
			}
			p.Karten[in.Pos], z.Karten[in.Pos2] = z.Karten[in.Pos2], p.Karten[in.Pos]
			delete(p.Bekannt, caboSchluessel(sitz, in.Pos))
			delete(p.Bekannt, caboSchluessel(in.Ziel, in.Pos2))
			delete(z.Bekannt, caboSchluessel(in.Ziel, in.Pos2))
			delete(z.Bekannt, caboSchluessel(sitz, in.Pos))
			g.Note = p.Name + " tauscht eine Karte mit " + z.Name + "."
		}
		caboZugWeiter(g)
		return nil

	case "satz":
		if g.Gezogen == nil {
			return caboNeu("Dafür musst du zuerst eine Karte gezogen haben.", http.StatusConflict)
		}
		if sitz != g.Zug {
			return caboNeu("Du bist nicht an der Reihe.", http.StatusConflict)
		}
		if !caboSatz(g, sitz, in.Ziele) {
			g.Note = p.Name + " behauptet einen Satz – falsch, der Zug ist vorbei."
			caboZugWeiter(g)
			return nil
		}
		g.Note = p.Name + " wirft einen Satz ab und legt " + strconv.Itoa(*g.Gezogen) + " hinein."
		g.Gezogen = nil
		caboZugWeiter(g)
		return nil

	case "cabo":
		if g.Status != "active" {
			return caboNeu("Gerade läuft kein Zug.", http.StatusConflict)
		}
		if sitz != g.Zug {
			return caboNeu("Du bist nicht an der Reihe.", http.StatusConflict)
		}
		if g.CaboVon >= 0 {
			return caboNeu("Cabo wurde in dieser Runde schon gerufen.", http.StatusConflict)
		}
		if g.Gezogen != nil {
			return caboNeu("Cabo ruft man, bevor man eine Karte zieht.", http.StatusConflict)
		}
		g.CaboVon = sitz
		// Der Rufende zählt mit: sein eigener Zug wird gleich beendet, danach
		// kommt jeder andere noch genau einmal dran.
		g.RestZuege = len(g.Spieler)
		g.Note = p.Name + " ruft CABO! Alle anderen sind noch einmal dran."
		caboZugWeiter(g)
		return nil

	case "naechste":
		if g.Status != "wertung" {
			return caboNeu("Die Runde ist noch nicht gewertet.", http.StatusConflict)
		}
		caboNeueRunde(g)
		return nil
	}
	return caboNeu("Unbekannte Cabo-Aktion.", http.StatusBadRequest)
}

// ---------------------------------------------------------------- Computer

const (
	// Kurze Pause zwischen zwei Schritten des Computers: Der Mensch sieht
	// ziehen und handeln getrennt, ohne auf etwas zu warten.
	caboBotVerzug = 520 * time.Millisecond
	// Erwartungswert einer unbekannten Karte (54 Karten, Summe 338).
	caboKartenWert = 6.26
)

func (g *CaboGame) hatBot() bool {
	for _, p := range g.Spieler {
		if p.Bot {
			return true
		}
	}
	return false
}

func caboBotAmZug(g *CaboGame) bool {
	if g == nil {
		return false
	}
	if g.Status == "anschauen" {
		for _, p := range g.Spieler {
			if p.Bot && p.Geschaut < caboStartBlick {
				return true
			}
		}
		return false
	}
	if g.Status != "active" || g.Zug < 0 || g.Zug >= len(g.Spieler) {
		return false
	}
	return g.Spieler[g.Zug].Bot
}

// caboBotTreiben lässt den Computer genau einen Schritt machen, sobald die
// kurze Pause vorbei ist. Aufgerufen wird es beim Abholen des Spielstands –
// der Browser fragt im Sekundentakt nach, damit der Mensch jeden Schritt sieht.
func caboBotTreiben(g *CaboGame, jetzt time.Time) bool {
	if !caboBotAmZug(g) || jetzt.Sub(g.BotMoment) < caboBotVerzug {
		return false
	}
	if !caboBotSchritt(g) {
		return false
	}
	g.BotMoment = jetzt
	return true
}

func caboBotSchritt(g *CaboGame) bool {
	if g == nil {
		return false
	}
	if g.Status == "anschauen" {
		for i := range g.Spieler {
			p := &g.Spieler[i]
			if !p.Bot || p.Geschaut >= caboStartBlick || len(p.Karten) == 0 {
				continue
			}
			pos := caboBotUnbekannt(p, i, -1)
			p.Bekannt[caboSchluessel(i, pos)] = p.Karten[pos]
			p.Geschaut++
			g.Note = p.Name + " schaut sich Karten an."
			caboPruefeAnschauen(g)
			return true
		}
		return false
	}
	if g.Status != "active" || g.Zug < 0 || g.Zug >= len(g.Spieler) {
		return false
	}
	sitz := g.Zug
	if !g.Spieler[sitz].Bot {
		return false
	}
	return caboBotZug(g, sitz)
}

// caboBotZug entscheidet einen Zug. Der Computer sieht nur, was ein Mensch
// auch sieht: den Ablagestapel, die Zahl der Karten, seine eigenen gemerkten
// Karten und alles, was er sich per Anschauen oder Spähen gemerkt hat.
func caboBotZug(g *CaboGame, sitz int) bool {
	p := &g.Spieler[sitz]
	stufe := caboStufe(p.Stufe)

	if g.Gezogen == nil {
		p.Zuege++
		if g.CaboVon < 0 && caboBotRuftCabo(p, sitz, stufe) {
			if caboAktion(g, sitz, "cabo", caboEingabe{}) == nil {
				return true
			}
		}
		if len(g.Ablage) > 0 {
			oben := g.Ablage[len(g.Ablage)-1]
			if caboBotNimmtAblage(p, sitz, oben, stufe) {
				if caboAktion(g, sitz, "nehmen", caboEingabe{}) == nil {
					return true
				}
			}
		}
		return caboAktion(g, sitz, "ziehen", caboEingabe{}) == nil
	}

	wert := *g.Gezogen
	if g.Quelle == "stapel" {
		if akt, ziel, pos, pos2, ok := caboBotAktion(p, g, sitz, wert, stufe); ok {
			if caboAktion(g, sitz, akt, caboEingabe{Ziel: ziel, Pos: pos, Pos2: pos2}) == nil {
				return true
			}
		}
		if ziele, ok := caboBotSatzWahl(p, sitz, wert, stufe); ok {
			if caboAktion(g, sitz, "satz", caboEingabe{Ziele: ziele}) == nil {
				return true
			}
		}
	}
	if pos, legen := caboBotLegen(p, sitz, wert, stufe); legen {
		if caboAktion(g, sitz, "legen", caboEingabe{Pos: pos}) == nil {
			return true
		}
	}
	if g.Quelle == "ablage" {
		// Vom Ablagestapel genommene Karten landen immer in der Auslage.
		return caboAktion(g, sitz, "legen", caboEingabe{Pos: caboBotUnbekannt(p, sitz, -1)}) == nil
	}
	return caboAktion(g, sitz, "abwerfen", caboEingabe{}) == nil
}

func caboStufe(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "leicht":
		return "leicht"
	case "schwer":
		return "schwer"
	}
	return "mittel"
}

// caboBotSumme schätzt die eigene Auslage: bekannte Karten zählen genau,
// unbekannte mit dem Durchschnittswert des Decks.
func caboBotSumme(p *CaboSpieler, sitz int) float64 {
	summe := 0.0
	for pos := 0; pos < len(p.Karten); pos++ {
		if wert, ok := p.Bekannt[caboSchluessel(sitz, pos)]; ok {
			summe += float64(wert)
		} else {
			summe += caboKartenWert
		}
	}
	return summe
}

// caboBotUnbekannt liefert eine eigene Kartenposition, deren Wert der Computer
// nicht kennt; ohne solche eine zufällige andere (nicht die ausgenommene).
func caboBotUnbekannt(p *CaboSpieler, sitz, ausser int) int {
	frei := []int{}
	for pos := 0; pos < len(p.Karten); pos++ {
		if pos == ausser {
			continue
		}
		if _, bekannt := p.Bekannt[caboSchluessel(sitz, pos)]; !bekannt {
			frei = append(frei, pos)
		}
	}
	if len(frei) == 0 {
		alle := []int{}
		for pos := 0; pos < len(p.Karten); pos++ {
			if pos != ausser {
				alle = append(alle, pos)
			}
		}
		if len(alle) == 0 {
			return 0
		}
		return alle[rand.Intn(len(alle))]
	}
	return frei[rand.Intn(len(frei))]
}

// caboBotRuftCabo: Der Computer ruft, wenn seine Auslage nach eigener Schätzung
// klein ist. Damit eine Runde überhaupt endet, ruft er nach einer Weile
// Geduld auch ohne sichere Schätzung – wie ein Mensch, der die Runde beenden will.
func caboBotRuftCabo(p *CaboSpieler, sitz int, stufe string) bool {
	summe := caboBotSumme(p, sitz)
	schwelle, geduld := 12.0, 6
	switch stufe {
	case "schwer":
		schwelle, geduld = 7.0, 12
	case "mittel":
		schwelle, geduld = 9.5, 9
	}
	if summe <= schwelle {
		return true
	}
	return p.Zuege >= geduld && summe <= 18
}

func caboBotNimmtAblage(p *CaboSpieler, sitz, oben int, stufe string) bool {
	if stufe == "leicht" {
		return oben == 0 || (oben <= 2 && rand.Intn(100) < 40)
	}
	bester := -1
	for pos := 0; pos < len(p.Karten); pos++ {
		if wert, ok := p.Bekannt[caboSchluessel(sitz, pos)]; ok && wert > bester {
			bester = wert
		}
	}
	if stufe == "schwer" {
		return oben <= 4 && (bester < 0 || bester > oben)
	}
	return oben <= 3 && bester > oben
}

// caboBotAktion nutzt die Aktion einer gezogenen Karte, wenn sie etwas bringt.
func caboBotAktion(p *CaboSpieler, g *CaboGame, sitz, wert int, stufe string) (string, int, int, int, bool) {
	switch {
	case wert == 7 || wert == 8:
		pos := -1
		for k := 0; k < len(p.Karten); k++ {
			if _, bekannt := p.Bekannt[caboSchluessel(sitz, k)]; !bekannt {
				pos = k
				break
			}
		}
		if pos < 0 {
			return "", 0, 0, 0, false
		}
		if stufe == "leicht" && rand.Intn(100) < 35 {
			return "", 0, 0, 0, false
		}
		return "peek", 0, pos, 0, true

	case wert == 9 || wert == 10:
		ziel, pos := -1, -1
		// Erst beim Menschen spähen, dann bei anderen; unbekannte Karten zuerst.
		for i := range g.Spieler {
			if i == sitz || i == ziel {
				continue
			}
			for k := 0; k < len(g.Spieler[i].Karten); k++ {
				if _, kennt := p.Bekannt[caboSchluessel(i, k)]; kennt {
					continue
				}
				if ziel < 0 || (!g.Spieler[i].Bot && g.Spieler[ziel].Bot) {
					ziel, pos = i, k
				}
			}
		}
		if ziel < 0 {
			return "", 0, 0, 0, false
		}
		if stufe == "leicht" && rand.Intn(100) < 35 {
			return "", 0, 0, 0, false
		}
		return "spy", ziel, pos, 0, true

	case wert == 11 || wert == 12:
		// Eigene höchste bekannte Karte weggeben, möglichst gegen eine beim
		// Mitspieler, die der Computer als klein kennt.
		eigen, eigenWert := -1, -1
		for k := 0; k < len(p.Karten); k++ {
			if v, ok := p.Bekannt[caboSchluessel(sitz, k)]; ok && v > eigenWert {
				eigen, eigenWert = k, v
			}
		}
		grenze := 7
		if stufe == "leicht" {
			grenze = 5
		}
		if eigen < 0 || eigenWert < grenze {
			return "", 0, 0, 0, false
		}
		ziel, pos, posWert := -1, -1, 99
		for i := range g.Spieler {
			if i == sitz {
				continue
			}
			for k := 0; k < len(g.Spieler[i].Karten); k++ {
				v, kennt := p.Bekannt[caboSchluessel(i, k)]
				if kennt && v < posWert {
					ziel, pos, posWert = i, k, v
				}
			}
		}
		if ziel < 0 {
			// Nichts bekannt: irgendeine Karte eines Mitspielers, bevorzugt
			// beim Menschen (dort ist die Auslage am ehesten interessant).
			for i := range g.Spieler {
				if i == sitz || len(g.Spieler[i].Karten) == 0 {
					continue
				}
				if ziel < 0 || (!g.Spieler[i].Bot && g.Spieler[ziel].Bot) {
					ziel = i
				}
			}
			if ziel < 0 {
				return "", 0, 0, 0, false
			}
			pos = rand.Intn(len(g.Spieler[ziel].Karten))
		}
		if stufe == "leicht" && rand.Intn(100) < 40 {
			return "", 0, 0, 0, false
		}
		return "swap", ziel, eigen, pos, true
	}
	return "", 0, 0, 0, false
}

// caboBotSatzWahl sucht einen Satz aus Karten, deren Wert der Computer kennt –
// ein falscher Satz würde den Zug beenden, das riskiert er nicht.
func caboBotSatzWahl(p *CaboSpieler, sitz, wert int, stufe string) ([]int, bool) {
	if stufe == "leicht" && rand.Intn(100) >= 40 {
		return nil, false
	}
	gruppen := map[int][]int{}
	for pos := 0; pos < len(p.Karten); pos++ {
		if v, ok := p.Bekannt[caboSchluessel(sitz, pos)]; ok {
			gruppen[v] = append(gruppen[v], pos)
		}
	}
	besterWert, beste := -1, []int{}
	for v, pos := range gruppen {
		if len(pos) < 2 {
			continue
		}
		if float64(v*len(pos)) <= float64(wert) {
			continue
		}
		if v*len(pos) > besterWert {
			besterWert, beste = v*len(pos), pos
		}
	}
	if len(beste) < 2 {
		return nil, false
	}
	sort.Ints(beste)
	return beste, true
}

// caboBotLegen wählt die eigene Karte, die weichen soll: die höchste bekannte,
// wenn die gezogene kleiner ist. Sonst wandert die Karte weg.
func caboBotLegen(p *CaboSpieler, sitz, wert int, stufe string) (int, bool) {
	bester, besteWert := -1, -1
	for pos := 0; pos < len(p.Karten); pos++ {
		if v, ok := p.Bekannt[caboSchluessel(sitz, pos)]; ok && v > besteWert {
			bester, besteWert = pos, v
		}
	}
	if bester >= 0 && besteWert > wert {
		return bester, true
	}
	if stufe != "leicht" && wert <= 3 {
		if pos := caboBotUnbekannt(p, sitz, bester); pos >= 0 {
			return pos, true
		}
	}
	if stufe == "leicht" && rand.Intn(100) < 45 {
		return caboBotUnbekannt(p, sitz, -1), true
	}
	return 0, false
}

// ---------------------------------------------------------------- Endpunkt

func caboHandler(w http.ResponseWriter, r *http.Request) {
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
		// Der Computer zieht beim Abholen des Stands – ein Schritt je Abruf.
		if room.Cabo != nil {
			caboBotTreiben(room.Cabo, time.Now())
		}
		game := caboAnsicht(room.Cabo, id)
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
		Pos   int    `json:"pos"`
		Pos2  int    `json:"pos2"`
		Ziel  int    `json:"ziel"`
		Ziele []int  `json:"ziele"`
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
	var problem *caboFehler

	switch action {
	case "create", "solo", "bot":
		if room.Cabo != nil && room.Cabo.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Cabo-Runde.", http.StatusConflict)
			return
		}
		game := &CaboGame{
			ID:      randomCode(),
			HostID:  input.ID,
			Status:  "invited",
			Spieler: []CaboSpieler{caboNeuerSpieler(input.ID, player.Name)},
			Note:    "Tisch eröffnet – 2 bis 4 Spieler, vier Karten vor jedem.",
		}
		room.Cabo = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Cabo-Runde – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})
		if action == "solo" || action == "bot" {
			room.botNamenNeu()
			stufe := caboStufe(input.Stufe)
			game.Spieler[0].Ready = true
			game.Spieler = append(game.Spieler, CaboSpieler{
				ID: botID(1), Name: room.botName(1), Karten: []int{}, Bekannt: map[string]int{},
				Ready: true, Bot: true, Stufe: stufe,
			})
			game.Start = 0
			game.Runde = 0
			caboNeueRunde(game)
			game.Note = "Einzelspieler gegen " + room.botName(1) + " (" + stufe + "). " + game.Note
		}

	case "join":
		game := room.Cabo
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Cabo-Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if game.hatBot() {
			mu.Unlock()
			http.Error(w, "Diese Runde läuft gegen den Computer.", http.StatusConflict)
			return
		}
		drin := false
		for _, p := range game.Spieler {
			if p.ID == input.ID {
				drin = true
				break
			}
		}
		if !drin {
			if len(game.Spieler) >= caboMaxSpieler {
				mu.Unlock()
				http.Error(w, "Dieser Tisch ist mit 4 Spielern voll.", http.StatusConflict)
				return
			}
			game.Spieler = append(game.Spieler, caboNeuerSpieler(input.ID, player.Name))
			game.Note = player.Name + " setzt sich dazu."
			room.addMessage(Message{
				Name:     player.Name,
				Text:     "spielt Cabo mit.",
				Time:     zeit,
				SenderID: input.ID,
			})
		}

	case "ready":
		game := room.Cabo
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := caboFinde(game, input.ID)
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
		game := room.Cabo
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Runde auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber kann geben.", http.StatusForbidden)
			return
		}
		if len(game.Spieler) < 2 {
			mu.Unlock()
			http.Error(w, "Cabo braucht mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, p := range game.Spieler {
			if !p.Ready {
				mu.Unlock()
				http.Error(w, "Erst müssen alle BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.Runde = 0
		caboNeueRunde(game)
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "gibt die Karten – jeder schaut sich zwei an.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "reset":
		game := room.Cabo
		if game == nil || game.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Eine neue Partie geht erst nach dem Ende.", http.StatusConflict)
			return
		}
		if caboFinde(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Runde.", http.StatusForbidden)
			return
		}
		frisch := &CaboGame{
			ID:      game.ID,
			HostID:  game.HostID,
			Status:  "invited",
			Spieler: []CaboSpieler{},
			Note:    "Neue Partie – erst BEREIT drücken.",
		}
		for _, p := range game.Spieler {
			neu := caboNeuerSpieler(p.ID, p.Name)
			neu.Bot = p.Bot
			neu.Stufe = p.Stufe
			neu.Ready = p.Bot
			frisch.Spieler = append(frisch.Spieler, neu)
		}
		room.Cabo = frisch
		if frisch.hatBot() {
			// Gegen den Computer geht es ohne Wartezeit gleich weiter.
			frisch.Runde = 0
			caboNeueRunde(frisch)
			frisch.Note = "Neue Partie gegen den Computer. " + frisch.Note
		}

	default:
		sitz := caboFinde(room.Cabo, input.ID)
		problem = caboAktion(room.Cabo, sitz, action, caboEingabe{
			Pos: input.Pos, Pos2: input.Pos2, Ziel: input.Ziel, Ziele: input.Ziele,
		})
		if problem != nil {
			mu.Unlock()
			http.Error(w, problem.Text, problem.Nr)
			return
		}
	}

	room.LastActive = time.Now()
	if room.Cabo != nil {
		caboBotTreiben(room.Cabo, time.Now())
	}
	result := caboAnsicht(room.Cabo, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

func caboFinde(g *CaboGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}
