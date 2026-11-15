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

// Sechsundsechzig („66"): Zwei Spieler, 24 Karten, zwei Hände zu sechs Blatt,
// der Rest liegt als Talon mit einer offenen Trumpfkarte. Wer zuerst 66 Augen
// meldet, gewinnt die Runde. Jeder startet mit sieben Bummerl, wer auf null
// kommt gewinnt die Partie. Geld kommt hier nicht vor – nur Punkte.

const (
	sechsundsechzigStart = 7
	sechsundsechzigHand  = 6
)

// Eine Karte ist Farbe*100 + Wert. Farben: 0 Herz, 1 Karo, 2 Pik, 3 Kreuz.
// Werte: 9, 10, 11 Bube, 12 Dame, 13 König, 14 Ass.
type SechsundsechzigSpieler struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Hand     []int   `json:"-"`
	Angesagt [4]bool `json:"-"`
	Stiche   int     `json:"stiche"`
	Augen    int     `json:"augen"`
	Ansage   int     `json:"ansage"`
	Rest     int     `json:"rest"`
	Ready    bool    `json:"ready"`
}

type SechsundsechzigZug struct {
	Karte int `json:"karte"`
	Von   int `json:"von"`
}

type SechsundsechzigErgebnis struct {
	Sieger  int    `json:"sieger"`
	Punkte  int    `json:"punkte"`
	Grund   string `json:"grund"`
	Augen   [2]int `json:"augen"`
	Stiche  [2]int `json:"stiche"`
	Ansagen [2]int `json:"ansagen"`
	Rest    [2]int `json:"rest"`
}

type SechsundsechzigGame struct {
	ID                      string                   `json:"id"`
	HostID                  string                   `json:"hostId"`
	Status                  string                   `json:"status"` // invited, spiel, runde, ende
	Runde                   int                      `json:"runde"`
	Geber                   int                      `json:"geber"`
	Vorhand                 int                      `json:"vorhand"`
	Trumpf                  int                      `json:"trumpf"`
	TrumpfKarte             int                      `json:"trumpfKarte"`
	Talon                   []int                    `json:"-"`
	Gesperrt                bool                     `json:"gesperrt"`
	Sperrer                 int                      `json:"-"`
	GegnerSticheBeimSperren int                      `json:"-"`
	Spieler                 []SechsundsechzigSpieler `json:"spieler"`
	Zug                     int                      `json:"zug"`
	Tisch                   []SechsundsechzigZug     `json:"tisch"`
	AnsagePflicht           int                      `json:"-"`
	Note                    string                   `json:"note"`
	Sieger                  string                   `json:"sieger,omitempty"`
	Ergebnis                *SechsundsechzigErgebnis `json:"ergebnis,omitempty"`
}

type SechsundsechzigSpielerSicht struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Hand     []int  `json:"hand"`
	Anzahl   int    `json:"anzahl"`
	Stiche   int    `json:"stiche"`
	Augen    int    `json:"augen"`
	Ansage   int    `json:"ansage"`
	Rest     int    `json:"rest"`
	Ready    bool   `json:"ready"`
	Ansagbar []int  `json:"ansagbar"` // Farben, in denen König und Dame auf der Hand liegen
}

type SechsundsechzigAnsicht struct {
	ID            string                        `json:"id"`
	HostID        string                        `json:"hostId"`
	Status        string                        `json:"status"`
	Runde         int                           `json:"runde"`
	Geber         int                           `json:"geber"`
	Vorhand       int                           `json:"vorhand"`
	Trumpf        int                           `json:"trumpf"`
	TrumpfKarte   int                           `json:"trumpfKarte"`
	TalonRest     int                           `json:"talonRest"`
	Gesperrt      bool                          `json:"gesperrt"`
	Zwang         bool                          `json:"zwang"`
	Spieler       []SechsundsechzigSpielerSicht `json:"spieler"`
	Zug           int                           `json:"zug"`
	Tisch         []SechsundsechzigZug          `json:"tisch"`
	AnsagePflicht int                           `json:"ansagePflicht"`
	Note          string                        `json:"note"`
	Sieger        string                        `json:"sieger,omitempty"`
	Ergebnis      *SechsundsechzigErgebnis      `json:"ergebnis,omitempty"`
}

// ---------------------------------------------------------------- Karten

func sechsundsechzigFarbe(karte int) int { return karte / 100 }
func sechsundsechzigWert(karte int) int  { return karte % 100 }

func sechsundsechzigPunkte(karte int) int {
	switch sechsundsechzigWert(karte) {
	case 14:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	}
	return 0
}

func sechsundsechzigDeck() []int {
	deck := make([]int, 0, 24)
	for farbe := 0; farbe < 4; farbe++ {
		for _, wert := range []int{9, 10, 11, 12, 13, 14} {
			deck = append(deck, farbe*100+wert)
		}
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

// ---------------------------------------------------------------- Aufbau

func sechsundsechzigNeuerSpieler(id, name string) SechsundsechzigSpieler {
	return SechsundsechzigSpieler{
		ID: id, Name: name, Hand: []int{}, Rest: sechsundsechzigStart,
	}
}

func sechsundsechzigSitz(g *SechsundsechzigGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func sechsundsechzigZwang(g *SechsundsechzigGame) bool {
	return g.Gesperrt || len(g.Talon) == 0
}

func sechsundsechzigGeben(g *SechsundsechzigGame) {
	deck := sechsundsechzigDeck()
	g.Spieler[0].Hand = append([]int{}, deck[0:6]...)
	g.Spieler[1].Hand = append([]int{}, deck[6:12]...)
	g.TrumpfKarte = deck[12]
	g.Trumpf = sechsundsechzigFarbe(deck[12])
	// Die offene Trumpfkarte liegt unter dem Talon: sie wird als letzte gehoben.
	talon := append([]int{}, deck[13:]...)
	talon = append(talon, g.TrumpfKarte)
	g.Talon = talon
	g.Gesperrt = false
	g.Sperrer = -1
	g.GegnerSticheBeimSperren = 0
	g.Tisch = []SechsundsechzigZug{}
	g.AnsagePflicht = -1
	g.Ergebnis = nil
	for i := range g.Spieler {
		g.Spieler[i].Stiche = 0
		g.Spieler[i].Augen = 0
		g.Spieler[i].Ansage = 0
		g.Spieler[i].Angesagt = [4]bool{}
	}
	g.Vorhand = (g.Geber + 1) % 2
	g.Zug = g.Vorhand
	g.Status = "spiel"
	g.Note = g.Spieler[g.Vorhand].Name + " ist Vorhand und spielt aus · Trumpf " +
		sechsundsechzigFarbname(g.Trumpf) + " (" + sechsundsechzigKartenname(g.TrumpfKarte) + ")."
}

func sechsundsechzigFarbname(farbe int) string {
	namen := []string{"Herz", "Karo", "Pik", "Kreuz"}
	if farbe < 0 || farbe > 3 {
		return "?"
	}
	return namen[farbe]
}

func sechsundsechzigKartenname(karte int) string {
	farbe := sechsundsechzigFarbe(karte)
	wert := sechsundsechzigWert(karte)
	farben := []string{"♥", "♦", "♠", "♣"}
	werte := map[int]string{9: "9", 10: "10", 11: "B", 12: "D", 13: "K", 14: "A"}
	name := werte[wert]
	if name == "" {
		name = strconv.Itoa(wert)
	}
	return farben[farbe] + name
}

// ---------------------------------------------------------------- Zugregeln

// sechsundsechzigFolge prüft Farb- und Stichzwang. farbe ist die angespielte
// Farbe, leadWert der Wert der angespielten Karte.
func sechsundsechzigFolge(g *SechsundsechzigGame, hand []int, karte, farbe, leadWert int) bool {
	hatFarbe, hoeher := false, false
	for _, k := range hand {
		if sechsundsechzigFarbe(k) == farbe {
			hatFarbe = true
			if sechsundsechzigWert(k) > leadWert {
				hoeher = true
			}
		}
	}
	if hatFarbe {
		if sechsundsechzigFarbe(karte) != farbe {
			return false
		}
		// Stichzwang: eine höhere Karte der Farbe muss kommen, wenn sie da ist.
		if hoeher && sechsundsechzigWert(karte) <= leadWert {
			return false
		}
		return true
	}
	// Farbzwang geht vor Stichzwang: ohne die Farbe muss Trumpf gestochen werden.
	if farbe != g.Trumpf {
		hatTrumpf := false
		for _, k := range hand {
			if sechsundsechzigFarbe(k) == g.Trumpf {
				hatTrumpf = true
			}
		}
		if hatTrumpf && sechsundsechzigFarbe(karte) != g.Trumpf {
			return false
		}
	}
	return true
}

func sechsundsechzigErlaubt(g *SechsundsechzigGame, sitz, karte int) bool {
	hand := g.Spieler[sitz].Hand
	if len(g.Tisch) == 0 {
		// Nach einer Ansage muss eine der beiden Karten ausgespielt werden.
		if g.AnsagePflicht >= 0 {
			if sechsundsechzigFarbe(karte) != g.AnsagePflicht {
				return false
			}
			wert := sechsundsechzigWert(karte)
			if wert != 12 && wert != 13 {
				return false
			}
		}
		return true
	}
	if !sechsundsechzigZwang(g) {
		return true
	}
	lead := g.Tisch[0].Karte
	return sechsundsechzigFolge(g, hand, karte, sechsundsechzigFarbe(lead), sechsundsechzigWert(lead))
}

func sechsundsechzigStich(g *SechsundsechzigGame) int {
	erste, zweite := g.Tisch[0], g.Tisch[1]
	ersteTrumpf := sechsundsechzigFarbe(erste.Karte) == g.Trumpf
	zweiteTrumpf := sechsundsechzigFarbe(zweite.Karte) == g.Trumpf
	if zweiteTrumpf && !ersteTrumpf {
		return zweite.Von
	}
	if zweiteTrumpf && ersteTrumpf {
		if sechsundsechzigWert(zweite.Karte) > sechsundsechzigWert(erste.Karte) {
			return zweite.Von
		}
		return erste.Von
	}
	if ersteTrumpf {
		return erste.Von
	}
	if sechsundsechzigFarbe(zweite.Karte) == sechsundsechzigFarbe(erste.Karte) &&
		sechsundsechzigWert(zweite.Karte) > sechsundsechzigWert(erste.Karte) {
		return zweite.Von
	}
	return erste.Von
}

// ---------------------------------------------------------------- Bewertung

func sechsundsechzigPunkteFuer(g *SechsundsechzigGame, sieger int) int {
	verlierer := 1 - sieger
	if g.Gesperrt && g.Sperrer == verlierer {
		if g.GegnerSticheBeimSperren == 0 {
			return 3
		}
		return 2
	}
	if g.Spieler[verlierer].Stiche == 0 {
		return 3
	}
	if g.Spieler[verlierer].Augen+g.Spieler[verlierer].Ansage <= 32 {
		return 2
	}
	return 1
}

func sechsundsechzigWerten(g *SechsundsechzigGame, sieger int, grund string, punkte int) {
	// Ansagen zählen nur, wenn der Ansager im Spiel mindestens einen Stich hat.
	for i := range g.Spieler {
		if g.Spieler[i].Ansage > 0 && g.Spieler[i].Stiche == 0 {
			g.Spieler[i].Ansage = 0
		}
	}
	if punkte <= 0 {
		punkte = sechsundsechzigPunkteFuer(g, sieger)
	}
	ergebnis := &SechsundsechzigErgebnis{Punkte: punkte, Grund: grund}
	for i := 0; i < 2; i++ {
		ergebnis.Augen[i] = g.Spieler[i].Augen + g.Spieler[i].Ansage
		ergebnis.Stiche[i] = g.Spieler[i].Stiche
		ergebnis.Ansagen[i] = g.Spieler[i].Ansage
	}
	ergebnis.Sieger = sieger
	g.Spieler[sieger].Rest -= punkte
	if g.Spieler[sieger].Rest <= 0 {
		g.Spieler[sieger].Rest = 0
	}
	for i := 0; i < 2; i++ {
		ergebnis.Rest[i] = g.Spieler[i].Rest
	}
	g.Ergebnis = ergebnis
	g.Tisch = []SechsundsechzigZug{}
	g.AnsagePflicht = -1
	g.Status = "runde"
	g.Note = g.Spieler[sieger].Name + " gewinnt die Runde mit " +
		strconv.Itoa(ergebnis.Augen[sieger]) + ":" + strconv.Itoa(ergebnis.Augen[1-sieger]) +
		" Augen · " + strconv.Itoa(punkte) + " Punkt(e)."
	if g.Spieler[sieger].Rest == 0 {
		g.Status = "ende"
		g.Sieger = g.Spieler[sieger].Name
		g.Note += " Damit ist die Partie gewonnen – " + g.Sieger + " hat alle Bummerl abgegeben."
	}
}

// ---------------------------------------------------------------- Sicht

func sechsundsechzigAnsicht(g *SechsundsechzigGame, wer string) *SechsundsechzigAnsicht {
	if g == nil {
		return nil
	}
	a := &SechsundsechzigAnsicht{
		ID:            g.ID,
		HostID:        g.HostID,
		Status:        g.Status,
		Runde:         g.Runde,
		Geber:         g.Geber,
		Vorhand:       g.Vorhand,
		Trumpf:        g.Trumpf,
		TrumpfKarte:   g.TrumpfKarte,
		TalonRest:     len(g.Talon),
		Gesperrt:      g.Gesperrt,
		Zwang:         sechsundsechzigZwang(g),
		Zug:           g.Zug,
		Note:          g.Note,
		Sieger:        g.Sieger,
		Ergebnis:      g.Ergebnis,
		Tisch:         append([]SechsundsechzigZug{}, g.Tisch...),
		AnsagePflicht: g.AnsagePflicht,
	}
	if g.Status == "invited" {
		a.TrumpfKarte = 0
	}
	for i, p := range g.Spieler {
		sicht := SechsundsechzigSpielerSicht{
			ID:     p.ID,
			Name:   p.Name,
			Anzahl: len(p.Hand),
			Stiche: p.Stiche,
			Augen:  p.Augen,
			Ansage: p.Ansage,
			Rest:   p.Rest,
			Ready:  p.Ready,
		}
		if p.ID == wer {
			hand := append([]int{}, p.Hand...)
			sort.Slice(hand, func(x, y int) bool {
				if sechsundsechzigFarbe(hand[x]) != sechsundsechzigFarbe(hand[y]) {
					return sechsundsechzigFarbe(hand[x]) < sechsundsechzigFarbe(hand[y])
				}
				return sechsundsechzigWert(hand[x]) > sechsundsechzigWert(hand[y])
			})
			sicht.Hand = hand
			for farbe := 0; farbe < 4; farbe++ {
				hatKoenig, hatDame := false, false
				for _, k := range p.Hand {
					if sechsundsechzigFarbe(k) != farbe {
						continue
					}
					if sechsundsechzigWert(k) == 13 {
						hatKoenig = true
					}
					if sechsundsechzigWert(k) == 12 {
						hatDame = true
					}
				}
				if hatKoenig && hatDame && !p.Angesagt[farbe] {
					sicht.Ansagbar = append(sicht.Ansagbar, farbe)
				}
			}
		}
		if i == g.Zug && g.Status == "spiel" {
			sicht.Stiche = p.Stiche
		}
		a.Spieler = append(a.Spieler, sicht)
	}
	return a
}

// ---------------------------------------------------------------- Endpunkt

func sechsundsechzigHandler(w http.ResponseWriter, r *http.Request) {
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
		game := sechsundsechzigAnsicht(room.Sechsundsechzig, id)
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
		Karte int    `json:"karte"`
		Farbe int    `json:"farbe"`
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
		if room.Sechsundsechzig != nil && room.Sechsundsechzig.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Runde Sechsundsechzig.", http.StatusConflict)
			return
		}
		room.Sechsundsechzig = &SechsundsechzigGame{
			ID:      randomCode(),
			HostID:  input.ID,
			Status:  "invited",
			Spieler: []SechsundsechzigSpieler{sechsundsechzigNeuerSpieler(input.ID, player.Name)},
			Zug:     -1,
			Sperrer: -1,
			Note:    "Tisch eröffnet – Sechsundsechzig ist ein Spiel für zwei.",
		}
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Runde Sechsundsechzig – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Sechsundsechzig
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if sechsundsechzigSitz(game, input.ID) >= 0 {
			break
		}
		if len(game.Spieler) >= 2 {
			mu.Unlock()
			http.Error(w, "Sechsundsechzig ist ein Spiel für zwei – der Tisch ist voll.", http.StatusConflict)
			return
		}
		game.Spieler = append(game.Spieler, sechsundsechzigNeuerSpieler(input.ID, player.Name))
		game.Note = player.Name + " sitzt jetzt mit am Tisch."

	case "ready":
		game := room.Sechsundsechzig
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := sechsundsechzigSitz(game, input.ID)
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
		game := room.Sechsundsechzig
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Runde auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber gibt die Karten.", http.StatusForbidden)
			return
		}
		if len(game.Spieler) < 2 {
			mu.Unlock()
			http.Error(w, "Sechsundsechzig braucht zwei Spieler.", http.StatusConflict)
			return
		}
		for _, p := range game.Spieler {
			if !p.Ready {
				mu.Unlock()
				http.Error(w, "Erst müssen beide BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.Runde = 0
		game.Geber = 1 // Vorhand ist damit Platz 0 in der ersten Runde
		sechsundsechzigNeueRunde(game)

	case "spielen":
		game := room.Sechsundsechzig
		if game == nil || game.Status != "spiel" {
			mu.Unlock()
			http.Error(w, "Gerade läuft kein Spiel.", http.StatusConflict)
			return
		}
		sitz := sechsundsechzigSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		if game.Zug != sitz {
			mu.Unlock()
			http.Error(w, "Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		gefunden := -1
		for i, k := range game.Spieler[sitz].Hand {
			if k == input.Karte {
				gefunden = i
				break
			}
		}
		if gefunden < 0 {
			mu.Unlock()
			http.Error(w, "Diese Karte hast du nicht auf der Hand.", http.StatusBadRequest)
			return
		}
		if !sechsundsechzigErlaubt(game, sitz, input.Karte) {
			mu.Unlock()
			http.Error(w, "Diese Karte ist nicht erlaubt – Farb- und Stichzwang.", http.StatusConflict)
			return
		}
		game.Spieler[sitz].Hand = append(game.Spieler[sitz].Hand[:gefunden], game.Spieler[sitz].Hand[gefunden+1:]...)
		game.Tisch = append(game.Tisch, SechsundsechzigZug{Karte: input.Karte, Von: sitz})
		if game.AnsagePflicht >= 0 && len(game.Tisch) == 1 {
			game.AnsagePflicht = -1
		}

		if len(game.Tisch) < 2 {
			game.Zug = 1 - sitz
			game.Note = game.Spieler[sitz].Name + " spielt " + sechsundsechzigKartenname(input.Karte) + "."
			break
		}

		gewinner := sechsundsechzigStich(game)
		augen := 0
		for _, zug := range game.Tisch {
			augen += sechsundsechzigPunkte(zug.Karte)
		}
		game.Spieler[gewinner].Stiche++
		game.Spieler[gewinner].Augen += augen
		game.Note = game.Spieler[gewinner].Name + " nimmt den Stich (" + strconv.Itoa(augen) +
			" Augen, zusammen " + strconv.Itoa(game.Spieler[gewinner].Augen+game.Spieler[gewinner].Ansage) + ")."
		game.Tisch = []SechsundsechzigZug{}
		sechsundsechzigHeben(game, gewinner)
		game.Zug = gewinner

		if len(game.Spieler[0].Hand) == 0 && len(game.Spieler[1].Hand) == 0 {
			// Letzter Stich: der Gewinner bekommt zehn Augen dazu – außer der
			// Talon wurde zugedreht.
			if !game.Gesperrt {
				game.Spieler[gewinner].Augen += 10
				game.Note += " Letzter Stich: zehn Augen extra für " + game.Spieler[gewinner].Name + "."
			}
			augen0 := game.Spieler[0].Augen + game.Spieler[0].Ansage
			augen1 := game.Spieler[1].Augen + game.Spieler[1].Ansage
			sieger := gewinner
			if augen0 != augen1 {
				if augen0 > augen1 {
					sieger = 0
				} else {
					sieger = 1
				}
			}
			sechsundsechzigWerten(game, sieger, "letzter", 0)
		}

	case "heiraten":
		game := room.Sechsundsechzig
		if game == nil || game.Status != "spiel" {
			mu.Unlock()
			http.Error(w, "Gerade läuft kein Spiel.", http.StatusConflict)
			return
		}
		sitz := sechsundsechzigSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		if game.Zug != sitz || len(game.Tisch) != 0 {
			mu.Unlock()
			http.Error(w, "Ansagen geht nur, wenn du selbst ausspielst.", http.StatusConflict)
			return
		}
		if input.Farbe < 0 || input.Farbe > 3 {
			mu.Unlock()
			http.Error(w, "Unbekannte Farbe.", http.StatusBadRequest)
			return
		}
		if game.Spieler[sitz].Angesagt[input.Farbe] {
			mu.Unlock()
			http.Error(w, "Diese Ansage hast du schon gemacht.", http.StatusConflict)
			return
		}
		hatKoenig, hatDame := false, false
		for _, k := range game.Spieler[sitz].Hand {
			if sechsundsechzigFarbe(k) != input.Farbe {
				continue
			}
			if sechsundsechzigWert(k) == 13 {
				hatKoenig = true
			}
			if sechsundsechzigWert(k) == 12 {
				hatDame = true
			}
		}
		if !hatKoenig || !hatDame {
			mu.Unlock()
			http.Error(w, "Für eine Ansage brauchst du König und Dame einer Farbe.", http.StatusConflict)
			return
		}
		game.Spieler[sitz].Angesagt[input.Farbe] = true
		wert := 20
		if input.Farbe == game.Trumpf {
			wert = 40
		}
		game.Spieler[sitz].Ansage += wert
		game.AnsagePflicht = input.Farbe
		game.Note = game.Spieler[sitz].Name + " sagt " + sechsundsechzigFarbname(input.Farbe) +
			" an: +" + strconv.Itoa(wert) + " Augen. Jetzt muss König oder Dame folgen."

	case "sperren":
		game := room.Sechsundsechzig
		if game == nil || game.Status != "spiel" {
			mu.Unlock()
			http.Error(w, "Gerade läuft kein Spiel.", http.StatusConflict)
			return
		}
		sitz := sechsundsechzigSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		if game.Zug != sitz || len(game.Tisch) != 0 {
			mu.Unlock()
			http.Error(w, "Zudrehen geht nur, wenn du selbst ausspielst.", http.StatusConflict)
			return
		}
		if game.Gesperrt {
			mu.Unlock()
			http.Error(w, "Der Talon ist schon zu.", http.StatusConflict)
			return
		}
		if len(game.Talon) <= 1 {
			mu.Unlock()
			http.Error(w, "Zu wenig im Talon – jetzt wird nur noch gespielt.", http.StatusConflict)
			return
		}
		game.Gesperrt = true
		game.Sperrer = sitz
		game.GegnerSticheBeimSperren = game.Spieler[1-sitz].Stiche
		game.Note = game.Spieler[sitz].Name + " dreht den Talon zu – ab jetzt gilt Farb- und Stichzwang, " +
			"der letzte Stich bringt keine zehn Augen."

	case "melden":
		game := room.Sechsundsechzig
		if game == nil || game.Status != "spiel" {
			mu.Unlock()
			http.Error(w, "Gerade läuft kein Spiel.", http.StatusConflict)
			return
		}
		sitz := sechsundsechzigSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		augen := game.Spieler[sitz].Augen + game.Spieler[sitz].Ansage
		if augen < 66 {
			// Falsch gemeldet: der Gegner bekommt die Punkte, die der Melder
			// bekommen hätte.
			punkte := sechsundsechzigPunkteFuer(game, sitz)
			sechsundsechzigWerten(game, 1-sitz, "falsch", punkte)
			game.Note = game.Spieler[sitz].Name + " meldet 66 mit nur " + strconv.Itoa(augen) +
				" Augen – der Gegner bekommt " + strconv.Itoa(punkte) + " Punkt(e)."
			break
		}
		sechsundsechzigWerten(game, sitz, "gemeldet", 0)
		game.Note = game.Spieler[sitz].Name + " meldet 66 mit " + strconv.Itoa(augen) + " Augen und gewinnt die Runde."

	case "naechste":
		game := room.Sechsundsechzig
		if game == nil || game.Status != "runde" {
			mu.Unlock()
			http.Error(w, "Die Runde ist noch nicht abgerechnet.", http.StatusConflict)
			return
		}
		if sechsundsechzigSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Runde.", http.StatusForbidden)
			return
		}
		game.Geber = 1 - game.Geber
		sechsundsechzigNeueRunde(game)

	case "reset":
		game := room.Sechsundsechzig
		if game == nil || game.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Eine neue Partie geht erst nach dem Ende.", http.StatusConflict)
			return
		}
		if sechsundsechzigSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Runde.", http.StatusForbidden)
			return
		}
		frisch := &SechsundsechzigGame{
			ID:      game.ID,
			HostID:  game.HostID,
			Status:  "invited",
			Zug:     -1,
			Sperrer: -1,
			Note:    "Neue Partie – erst BEREIT drücken.",
		}
		for _, p := range game.Spieler {
			frisch.Spieler = append(frisch.Spieler, sechsundsechzigNeuerSpieler(p.ID, p.Name))
		}
		room.Sechsundsechzig = frisch

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := sechsundsechzigAnsicht(room.Sechsundsechzig, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

// sechsundsechzigHeben legt die obersten Talonkarten nach: erst der
// Stichgewinner, dann der andere Spieler.
func sechsundsechzigNeueRunde(g *SechsundsechzigGame) {
	g.Runde++
	g.Status = "spiel"
	sechsundsechzigGeben(g)
}

func sechsundsechzigHeben(g *SechsundsechzigGame, gewinner int) {
	verlierer := 1 - gewinner
	if len(g.Talon) > 0 {
		g.Spieler[gewinner].Hand = append(g.Spieler[gewinner].Hand, g.Talon[0])
		g.Talon = g.Talon[1:]
	}
	if len(g.Talon) > 0 {
		g.Spieler[verlierer].Hand = append(g.Spieler[verlierer].Hand, g.Talon[0])
		g.Talon = g.Talon[1:]
	}
}
