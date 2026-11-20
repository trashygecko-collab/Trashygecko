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

// Schiffe versenken (Ship Captain Crew): fünf Würfel, höchstens drei Würfe.
// Gesucht wird in dieser Reihenfolge 6 (Schiff), 5 (Kapitän), 4 (Maat).
// Erst wenn alle drei gesichert sind, zählen die beiden übrigen Würfel als
// Ladung. Der Spieler sichert selbst; der Server würfelt und führt den Stand.
const scMaxPlayers = 6
const scWuerfel = 5
const scWuerfe = 3

var scColors = []string{"#e8cd8e", "#7fb98a", "#d98878", "#8aa9d6", "#b18cff", "#ff985e"}

// scSonder sind die drei Spezialwürfel in der Reihenfolge, in der sie fallen müssen.
var scSonder = []struct {
	Wert  int
	Titel string
}{
	{6, "Schiff"},
	{5, "Kapitän"},
	{4, "Maat"},
}

type ScPlayer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Ready  bool   `json:"ready"`
	Halten []int  `json:"halten"`  // gesicherte Spezialwürfel: Schiff, Kapitän, Maat
	Wuerfel []int `json:"wuerfel"` // der letzte Wurf der noch freien Würfel
	Wuerfe int    `json:"wuerfe"`
	Wert   int    `json:"wert"` // Ladungssumme, 0 = keine Kombination
	Fertig bool   `json:"fertig"`
	Siege  int    `json:"siege"`
}

type ScErgebnis struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Wert   int    `json:"wert"`
	Halten []int  `json:"halten"`
	Ladung []int  `json:"ladung"`
	Voll   bool   `json:"voll"`
}

type ScRunde struct {
	Nr         int          `json:"nr"`
	Ergebnisse []ScErgebnis `json:"ergebnisse"`
	Sieger     []string     `json:"sieger"`
	SiegerIDs  []string     `json:"siegerIds"`
	Best       int          `json:"best"`
}

type ShipCrewGame struct {
	ID        string     `json:"id"`
	HostID    string     `json:"hostId"`
	Players   []ScPlayer `json:"players"`
	Current   int        `json:"current"`
	Status    string     `json:"status"` // invited, active, runde, finished
	Runde     int        `json:"runde"`
	Runden    int        `json:"runden"`
	WurfNr    int        `json:"wurfNr"` // zählt jeden Wurf der Partie (für die Animation)
	Wer       string     `json:"wer"`
	Letzte    *ScRunde   `json:"letzte,omitempty"`
	Sieger    []string   `json:"siegerNamen"`
	SiegerIDs []string   `json:"siegerIds"`
	Note      string     `json:"note,omitempty"`
}

func scNewPlayer(id, name string, seat int) ScPlayer {
	return ScPlayer{ID: id, Name: name, Color: scColors[seat%len(scColors)], Halten: []int{}, Wuerfel: []int{}}
}

func scFind(g *ShipCrewGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// scErwartet liefert den Wert, der als nächstes gesichert werden darf.
// 0 heißt: alles gesichert, es zählt nur noch die Ladung.
func scErwartet(p *ScPlayer) int {
	if len(p.Halten) >= len(scSonder) {
		return 0
	}
	return scSonder[len(p.Halten)].Wert
}

// scLadung ist die Summe der beiden Würfel, die nach den Spezialwürfeln übrig sind.
// scTitel nennt den Spezialwürfel zu einem Wert.
func scTitel(wert int) string {
	for _, s := range scSonder {
		if s.Wert == wert {
			return s.Titel
		}
	}
	return "Würfel"
}

func scLadung(p *ScPlayer) int {
	if len(p.Halten) < len(scSonder) {
		return 0
	}
	s := 0
	for _, a := range p.Wuerfel {
		s += a
	}
	return s
}

func scSicht(g *ShipCrewGame) *ShipCrewGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Players = make([]ScPlayer, len(g.Players))
	for i, p := range g.Players {
		c.Players[i] = p
		c.Players[i].Halten = append([]int{}, p.Halten...)
		c.Players[i].Wuerfel = append([]int{}, p.Wuerfel...)
	}
	if g.Letzte != nil {
		l := *g.Letzte
		l.Ergebnisse = make([]ScErgebnis, len(g.Letzte.Ergebnisse))
		for i, e := range g.Letzte.Ergebnisse {
			l.Ergebnisse[i] = e
			l.Ergebnisse[i].Halten = append([]int{}, e.Halten...)
			l.Ergebnisse[i].Ladung = append([]int{}, e.Ladung...)
		}
		l.Sieger = append([]string{}, g.Letzte.Sieger...)
		l.SiegerIDs = append([]string{}, g.Letzte.SiegerIDs...)
		c.Letzte = &l
	}
	c.Sieger = append([]string{}, g.Sieger...)
	c.SiegerIDs = append([]string{}, g.SiegerIDs...)
	return &c
}

// scRundeStarten setzt alle Züge zurück und beginnt bei start.
func scRundeStarten(g *ShipCrewGame, start int) {
	for i := range g.Players {
		g.Players[i].Halten = []int{}
		g.Players[i].Wuerfel = []int{}
		g.Players[i].Wuerfe = 0
		g.Players[i].Wert = 0
		g.Players[i].Fertig = false
	}
	g.Status = "active"
	g.Wer = ""
	g.Current = start % len(g.Players)
	g.Note = "Runde " + strconv.Itoa(g.Runde) + ": " + g.Players[g.Current].Name + " beginnt."
}

// scWurfMit würfelt alle freien Würfel eines Spielers. augen dient den Tests.
func scWurfMit(g *ShipCrewGame, idx int, augen []int) []int {
	p := &g.Players[idx]
	n := scWuerfel - len(p.Halten)
	if n < 1 {
		n = 1
	}
	w := make([]int, n)
	if len(augen) >= n {
		copy(w, augen[:n])
	} else {
		for i := range w {
			w[i] = rand.Intn(6) + 1
		}
	}
	p.Wuerfel = w
	p.Wuerfe++
	g.WurfNr++
	g.Wer = p.ID
	return w
}

// scNaechster sucht den nächsten Spieler, der noch nicht fertig ist.
func scNaechster(g *ShipCrewGame, ab int) int {
	for i := 1; i <= len(g.Players); i++ {
		k := (ab + i) % len(g.Players)
		if !g.Players[k].Fertig {
			return k
		}
	}
	return -1
}

func scAlleFertig(g *ShipCrewGame) bool {
	for _, p := range g.Players {
		if !p.Fertig {
			return false
		}
	}
	return true
}

// scZugBeenden schließt den Zug ab und wechselt weiter.
func scZugBeenden(g *ShipCrewGame, idx int) {
	p := &g.Players[idx]
	p.Wert = scLadung(p)
	p.Fertig = true
	if p.Wert > 0 {
		g.Note = p.Name + " hat Schiff, Kapitän und Maat – Ladung " + strconv.Itoa(p.Wert) + "."
	} else {
		g.Note = "KEIN SCHIFF IN SICHT! " + p.Name + " bekommt 0 Punkte."
	}
	if scAlleFertig(g) {
		scAuswerten(g)
		return
	}
	if n := scNaechster(g, idx); n >= 0 {
		g.Current = n
	}
}

// scAuswerten wertet die Runde: höchste Ladung gewinnt, Gleichstand teilt.
func scAuswerten(g *ShipCrewGame) {
	r := &ScRunde{Nr: g.Runde, Best: 0}
	for _, p := range g.Players {
		voll := len(p.Halten) >= len(scSonder)
		ladung := []int{}
		if voll {
			ladung = append(ladung, p.Wuerfel...)
		}
		r.Ergebnisse = append(r.Ergebnisse, ScErgebnis{
			ID: p.ID, Name: p.Name, Wert: p.Wert,
			Halten: append([]int{}, p.Halten...), Ladung: ladung, Voll: voll,
		})
		if p.Wert > r.Best {
			r.Best = p.Wert
		}
	}
	for i := range g.Players {
		p := &g.Players[i]
		if p.Wert > 0 && p.Wert == r.Best {
			r.Sieger = append(r.Sieger, p.Name)
			r.SiegerIDs = append(r.SiegerIDs, p.ID)
			p.Siege++
		}
	}
	g.Letzte = r
	g.Status = "runde"

	if len(r.Sieger) == 0 {
		g.Note = "Runde " + strconv.Itoa(r.Nr) + ": niemand hatte eine Ladung."
	} else if len(r.Sieger) == 1 {
		g.Note = r.Sieger[0] + " gewinnt die Runde mit Ladung " + strconv.Itoa(r.Best) + "."
	} else {
		g.Note = strings.Join(r.Sieger, " und ") + " teilen sich die Runde mit Ladung " +
			strconv.Itoa(r.Best) + "."
	}

	if g.Runde >= g.Runden {
		// Die Rundenübersicht bleibt stehen; „weiter“ zeigt den Endstand.
		best := 0
		g.Sieger = []string{}
		g.SiegerIDs = []string{}
		for _, p := range g.Players {
			if p.Siege > best {
				best = p.Siege
			}
		}
		for _, p := range g.Players {
			if p.Siege == best && best > 0 {
				g.Sieger = append(g.Sieger, p.Name)
				g.SiegerIDs = append(g.SiegerIDs, p.ID)
			}
		}
		punkte := " Siegpunkten"
		if best == 1 {
			punkte = " Siegpunkt"
		}
		if len(g.Sieger) == 1 {
			g.Note = g.Sieger[0] + " gewinnt die Partie mit " + strconv.Itoa(best) + punkte + "."
		} else if len(g.Sieger) > 1 {
			g.Note = strings.Join(g.Sieger, " und ") + " teilen sich den Sieg mit " +
				strconv.Itoa(best) + punkte + "."
		}
	}
}

// scWeiter beginnt die nächste Runde – nach der letzten steht der Endstand.
func scWeiter(g *ShipCrewGame) {
	if g.Status != "runde" {
		return
	}
	if g.Runde >= g.Runden {
		g.Status = "finished"
		return
	}
	g.Runde++
	start := 0
	if g.Letzte != nil && len(g.Letzte.SiegerIDs) > 0 {
		if i := scFind(g, g.Letzte.SiegerIDs[0]); i >= 0 {
			start = i
		}
	}
	scRundeStarten(g, start)
}

// scRaus nimmt einen Platz aus der Partie.
func scRaus(g *ShipCrewGame, id, name string) (bool, bool, int) {
	idx := scFind(g, id)
	if idx < 0 {
		return false, false, len(g.Players)
	}
	g.Players = append(g.Players[:idx], g.Players[idx+1:]...)
	if len(g.Players) < 2 {
		return true, true, len(g.Players)
	}
	if g.HostID == id {
		g.HostID = g.Players[0].ID
	}
	if g.Current > idx {
		g.Current--
	}
	if g.Current >= len(g.Players) {
		g.Current = 0
	}
	if g.Status == "active" {
		if scAlleFertig(g) {
			scAuswerten(g)
		} else if g.Players[g.Current].Fertig {
			if n := scNaechster(g, g.Current); n >= 0 {
				g.Current = n
			}
		}
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}

// scSortiert liefert die Ergebnisse einer Runde (beste Ladung zuerst).
func scSortiert(g *ShipCrewGame) []ScErgebnis {
	out := []ScErgebnis{}
	if g.Letzte == nil {
		return out
	}
	out = append(out, g.Letzte.Ergebnisse...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Wert > out[j].Wert })
	return out
}

func shipcrewHandler(w http.ResponseWriter, r *http.Request) {
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
		sicht := scSicht(room.ShipCrew)
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
		Runden int    `json:"runden"`
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

	fehler := func(status int, text string) {
		mu.Unlock()
		http.Error(w, text, status)
	}

	switch action {
	case "create":
		if room.ShipCrew != nil && room.ShipCrew.Status != "finished" {
			fehler(http.StatusConflict, "In diesem Spielraum läuft bereits eine Runde Schiffe versenken.")
			return
		}
		if input.Runden != 1 && input.Runden != 3 && input.Runden != 5 && input.Runden != 10 {
			fehler(http.StatusBadRequest, "Bitte 1, 3, 5 oder 10 Runden wählen.")
			return
		}
		game := &ShipCrewGame{ID: randomCode(), HostID: input.ID, Status: "invited", Runde: 1, Runden: input.Runden}
		game.Players = append(game.Players, scNewPlayer(input.ID, player.Name, 0))
		wort := " Runden"
		if input.Runden == 1 {
			wort = " Runde"
		}
		game.Note = player.Name + " eröffnet Schiffe versenken – " + strconv.Itoa(input.Runden) + wort + "."
		room.ShipCrew = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet Schiffe versenken – fünf Würfel, drei Würfe.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "modus":
		// Der Gastgeber darf die Rundenzahl ändern, solange nicht gespielt wird.
		game := room.ShipCrew
		if game == nil || game.Status != "invited" {
			fehler(http.StatusConflict, "Der Modus lässt sich nur vor dem Start ändern.")
			return
		}
		if game.HostID != input.ID {
			fehler(http.StatusForbidden, "Nur der Gastgeber wählt den Modus.")
			return
		}
		if input.Runden != 1 && input.Runden != 3 && input.Runden != 5 && input.Runden != 10 {
			fehler(http.StatusBadRequest, "Bitte 1, 3, 5 oder 10 Runden wählen.")
			return
		}
		game.Runden = input.Runden
		wort := " Runden"
		if input.Runden == 1 {
			wort = " Runde"
		}
		game.Note = player.Name + " stellt den Modus auf " + strconv.Itoa(input.Runden) + wort + "."

	case "join":
		game := room.ShipCrew
		if game == nil || game.Status != "invited" {
			fehler(http.StatusConflict, "Es läuft gerade keine offene Runde Schiffe versenken.")
			return
		}
		if scFind(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= scMaxPlayers {
			fehler(http.StatusConflict, "Diese Runde ist mit 6 Spielern voll.")
			return
		}
		game.Players = append(game.Players, scNewPlayer(input.ID, player.Name, len(game.Players)))
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zu Schiffe versenken.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "rejoin":
		// Zurück an den Tisch, wenn die Verbindung mitten in der Partie abriss.
		game := room.ShipCrew
		if game == nil {
			fehler(http.StatusConflict, "Hier läuft keine Partie Schiffe versenken.")
			return
		}
		if scFind(game, input.ID) >= 0 {
			break
		}
		if game.Status == "runde" {
			fehler(http.StatusConflict, "Die Runde ist gerade vorbei – gleich geht es weiter.")
			return
		}
		if game.Status == "finished" || len(game.Players) >= scMaxPlayers {
			fehler(http.StatusConflict, "Hier ist kein Platz mehr frei.")
			return
		}
		game.Players = append(game.Players, scNewPlayer(input.ID, player.Name, len(game.Players)))
		game.Note = player.Name + " ist wieder dabei."
		// Läuft die Runde schon und alle anderen sind fertig, ist jetzt er dran –
		// sonst bliebe die Runde ohne Auswertung stehen.
		if game.Status == "active" {
			neu := len(game.Players) - 1
			alleAnderenFertig := true
			for i := range game.Players {
				if i != neu && !game.Players[i].Fertig {
					alleAnderenFertig = false
				}
			}
			if alleAnderenFertig {
				game.Current = neu
			}
		}

	case "ready":
		game := room.ShipCrew
		if game == nil || game.Status != "invited" {
			fehler(http.StatusConflict, "Bereit drücken geht nur, solange die Runde noch nicht läuft.")
			return
		}
		seat := scFind(game, input.ID)
		if seat < 0 {
			fehler(http.StatusForbidden, "Du sitzt nicht an diesem Tisch.")
			return
		}
		game.Players[seat].Ready = !game.Players[seat].Ready
		if game.Players[seat].Ready {
			game.Note = player.Name + " ist bereit."
		} else {
			game.Note = player.Name + " wartet noch."
		}

	case "start":
		game := room.ShipCrew
		if game == nil || game.Status != "invited" {
			fehler(http.StatusConflict, "Es wartet keine Runde auf den Start.")
			return
		}
		if game.HostID != input.ID {
			fehler(http.StatusForbidden, "Nur der Gastgeber kann die Runde starten.")
			return
		}
		if len(game.Players) < 2 {
			fehler(http.StatusConflict, "Schiffe versenken braucht mindestens zwei Spieler.")
			return
		}
		for _, p := range game.Players {
			if !p.Ready {
				fehler(http.StatusConflict, "Erst müssen alle Spieler BEREIT drücken.")
				return
			}
		}
		game.Runde = 1
		game.Letzte = nil
		game.Sieger = nil
		game.SiegerIDs = nil
		for i := range game.Players {
			game.Players[i].Siege = 0
		}
		scRundeStarten(game, 0)

	case "wurf":
		game := room.ShipCrew
		if game == nil || game.Status != "active" {
			fehler(http.StatusConflict, "Es läuft gerade keine aktive Runde.")
			return
		}
		seat := scFind(game, input.ID)
		if seat < 0 {
			fehler(http.StatusForbidden, "Du sitzt nicht an diesem Tisch.")
			return
		}
		if seat != game.Current {
			fehler(http.StatusConflict, "Du bist nicht dran – "+game.Players[game.Current].Name+" würfelt.")
			return
		}
		p := &game.Players[seat]
		if p.Fertig {
			fehler(http.StatusConflict, "Dein Zug ist schon fertig.")
			return
		}
		if p.Wuerfe >= scWuerfe {
			fehler(http.StatusConflict, "Mehr als drei Würfe gibt es nicht.")
			return
		}
		scWurfMit(game, seat, nil)
		if p.Wuerfe >= scWuerfe {
			// Nach dem dritten Wurf ist nichts mehr zu entscheiden.
			scZugBeenden(game, seat)
		}

	case "sichern":
		game := room.ShipCrew
		if game == nil || game.Status != "active" {
			fehler(http.StatusConflict, "Es läuft gerade keine aktive Runde.")
			return
		}
		seat := scFind(game, input.ID)
		if seat < 0 {
			fehler(http.StatusForbidden, "Du sitzt nicht an diesem Tisch.")
			return
		}
		if seat != game.Current || game.Wer != input.ID {
			fehler(http.StatusConflict, "Erst würfeln, dann Würfel sichern.")
			return
		}
		p := &game.Players[seat]
		if p.Fertig {
			fehler(http.StatusConflict, "Dein Zug ist schon fertig.")
			return
		}
		if input.Index < 0 || input.Index >= len(p.Wuerfel) {
			fehler(http.StatusConflict, "Diesen Würfel gibt es nicht.")
			return
		}
		erwartet := scErwartet(p)
		if erwartet == 0 {
			fehler(http.StatusConflict, "Schiff, Kapitän und Maat sind schon gesichert.")
			return
		}
		if p.Wuerfel[input.Index] != erwartet {
			titel := "Schiff"
			for _, s := range scSonder {
				if s.Wert == erwartet {
					titel = s.Titel
				}
			}
			fehler(http.StatusConflict, "Als Nächstes zählt nur die "+strconv.Itoa(erwartet)+
				" ("+titel+").")
			return
		}
		// Der Würfel wandert zu den gesicherten und wird nicht mehr geworfen.
		p.Halten = append(p.Halten, p.Wuerfel[input.Index])
		p.Wuerfel = append(p.Wuerfel[:input.Index], p.Wuerfel[input.Index+1:]...)
		if len(p.Halten) == len(scSonder) {
			p.Wert = scLadung(p)
			game.Note = p.Name + " hat Schiff, Kapitän und Maat – die beiden übrigen Würfel sind die Ladung."
		} else {
			game.Note = p.Name + " sichert " + scTitel(p.Halten[len(p.Halten)-1]) + "."
		}

	case "bestaetigen":
		game := room.ShipCrew
		if game == nil || game.Status != "active" {
			fehler(http.StatusConflict, "Es läuft gerade keine aktive Runde.")
			return
		}
		seat := scFind(game, input.ID)
		if seat < 0 {
			fehler(http.StatusForbidden, "Du sitzt nicht an diesem Tisch.")
			return
		}
		if seat != game.Current {
			fehler(http.StatusConflict, "Du bist nicht dran.")
			return
		}
		if game.Players[seat].Fertig {
			fehler(http.StatusConflict, "Dein Zug ist schon fertig.")
			return
		}
		if game.Players[seat].Wuerfe == 0 {
			fehler(http.StatusConflict, "Erst würfeln.")
			return
		}
		scZugBeenden(game, seat)

	case "weiter":
		game := room.ShipCrew
		if game == nil || game.Status != "runde" {
			fehler(http.StatusConflict, "Die nächste Runde geht erst nach der Rundenübersicht.")
			return
		}
		if scFind(game, input.ID) < 0 {
			fehler(http.StatusForbidden, "Du gehörst nicht zu dieser Partie.")
			return
		}
		scWeiter(game)

	case "reset":
		game := room.ShipCrew
		if game == nil || game.Status != "finished" {
			fehler(http.StatusConflict, "Eine neue Partie ist erst nach Spielende möglich.")
			return
		}
		if scFind(game, input.ID) < 0 {
			fehler(http.StatusForbidden, "Du gehörst nicht zu dieser Partie.")
			return
		}
		fresh := &ShipCrewGame{ID: game.ID, HostID: game.HostID, Status: "invited", Runde: 1, Runden: game.Runden}
		for i, p := range game.Players {
			fresh.Players = append(fresh.Players, scNewPlayer(p.ID, p.Name, i))
		}
		fresh.Note = "Neue Partie – alle drücken BEREIT, dann startet der Gastgeber."
		room.ShipCrew = fresh

	default:
		fehler(http.StatusBadRequest, "Unbekannte Aktion für Schiffe versenken.")
		return
	}

	room.LastActive = time.Now()
	sicht := scSicht(room.ShipCrew)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": sicht})
}
