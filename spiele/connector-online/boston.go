package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Auf nach Boston: ein Würfelrennen über mehrere Runden. Drei Würfel, drei
// Würfe je Spieler – nach jedem Wurf bleibt der höchste Würfel liegen. Wer am
// Ende die höchste Summe hat, gewinnt die Runde und damit einen Siegpunkt.
// Am Schluss gewinnt, wer die meisten Siegpunkte hat (auch mehrere möglich).
//
// Die Würfel erzeugt nur der Server; der aktuelle Spieler löst seinen Wurf aus,
// alle anderen sehen ihn sofort mit (hier wird nichts verdeckt).
const bostonMaxPlayers = 6
const bostonWuerfe = 3

var bostonColors = []string{"#e8cd8e", "#7fb98a", "#d98878", "#8aa9d6", "#b18cff", "#ff985e"}

// erlaubte Rundenzahlen (der „Einsatz“ in Euro ist nur das Etikett dazu)
var bostonRundenWahl = []int{1, 3, 5, 10}

type BostonPlayer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Ready  bool   `json:"ready"`
	Siege  int    `json:"siege"`  // Siegpunkte aus allen Runden
	Punkte int    `json:"punkte"` // Summe der aktuellen (oder letzten) Runde
	Wurf   int    `json:"wurf"`   // wie viele Würfe in dieser Runde (0..3)
	Halten []int  `json:"halten"` // behaltene Würfel in der Reihenfolge des Wurfs
	Done   bool   `json:"done"`   // Zug dieser Runde abgeschlossen
}

// BostonRunde ist das Ergebnis einer abgeschlossenen Runde.
type BostonRunde struct {
	Nr          int            `json:"nr"`
	Summen      map[string]int `json:"summen"`
	Sieger      []string       `json:"sieger"`
	SiegerNamen []string       `json:"siegerNamen"`
	Best        int            `json:"best"`
}

type BostonGame struct {
	ID          string         `json:"id"`
	HostID      string         `json:"hostId"`
	Players     []BostonPlayer `json:"players"`
	Current     int            `json:"current"`
	Status      string         `json:"status"`  // invited, active, runde, finished
	Runden      int            `json:"runden"`  // gewählt: 1, 3, 5 oder 10
	Runde       int            `json:"runde"`   // 1-basiert
	Wuerfel     []int          `json:"wuerfel"` // der letzte Wurf (für alle sichtbar)
	WurfNr      int            `json:"wurfNr"`  // Zähler, damit die Seite neu animiert
	Wer         string         `json:"wer"`     // wer zuletzt gewürfelt hat
	Letzte      *BostonRunde   `json:"letzte,omitempty"`
	Sieger      []string       `json:"sieger,omitempty"`
	SiegerNamen []string       `json:"siegerNamen,omitempty"`
	Note        string         `json:"note,omitempty"`
	// Bot: Kennung, Stufe und die Zeit, zu der der Bot würfelt.
	BotID    string    `json:"botId,omitempty"`
	Bots     []string  `json:"bots,omitempty"`
	BotStufe string    `json:"botStufe,omitempty"`
	BotFrist time.Time `json:"-"`
}

func bostonNewPlayer(id, name string, seat int) BostonPlayer {
	return BostonPlayer{ID: id, Name: name, Color: bostonColors[seat%len(bostonColors)], Halten: []int{}}
}

// bostonSicht ist eine Kopie fürs JSON. Listen sind nie nil, damit die Seite
// immer ein Feld vorfindet (leer statt null).
func bostonSicht(g *BostonGame) *BostonGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Players = make([]BostonPlayer, len(g.Players))
	for i, p := range g.Players {
		c.Players[i] = p
		c.Players[i].Halten = append([]int{}, p.Halten...)
	}
	c.Wuerfel = append([]int{}, g.Wuerfel...)
	if g.Letzte != nil {
		l := *g.Letzte
		l.Sieger = append([]string{}, g.Letzte.Sieger...)
		l.SiegerNamen = append([]string{}, g.Letzte.SiegerNamen...)
		l.Summen = map[string]int{}
		for k, v := range g.Letzte.Summen {
			l.Summen[k] = v
		}
		c.Letzte = &l
	}
	c.Sieger = append([]string{}, g.Sieger...)
	c.SiegerNamen = append([]string{}, g.SiegerNamen...)
	return &c
}

func bostonFind(g *BostonGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// bostonSumme addiert die behaltenen Würfel.
func bostonSumme(halten []int) int {
	s := 0
	for _, w := range halten {
		s += w
	}
	return s
}

// bostonRundenWert prüft die gewählte Rundenzahl.
func bostonRundenWert(n int) (int, bool) {
	for _, w := range bostonRundenWahl {
		if w == n {
			return w, true
		}
	}
	return 0, false
}

// bostonRundeStarten setzt alle Spieler für eine neue Runde zurück.
func bostonRundeStarten(g *BostonGame) {
	for i := range g.Players {
		g.Players[i].Wurf = 0
		g.Players[i].Halten = []int{}
		g.Players[i].Punkte = 0
		g.Players[i].Done = false
	}
	g.Wuerfel = nil
	g.Wer = ""
	// Wer beginnt, rotiert von Runde zu Runde.
	if len(g.Players) > 0 {
		g.Current = (g.Runde - 1) % len(g.Players)
	} else {
		g.Current = 0
	}
	g.Status = "active"
	g.Note = "Runde " + strconv.Itoa(g.Runde) + " von " + strconv.Itoa(g.Runden) +
		": " + g.Players[g.Current].Name + " würfelt zuerst."
}

// bostonWurf würfelt die noch offenen Würfel und legt den höchsten beiseite.
func bostonWurf(g *BostonGame, idx int) []int {
	p := &g.Players[idx]
	n := bostonWuerfe - len(p.Halten)
	if n < 1 {
		n = 1
	}
	w := make([]int, n)
	best := 0
	for i := range w {
		w[i] = rand.Intn(6) + 1
		if w[i] > w[best] {
			best = i
		}
	}
	p.Halten = append(p.Halten, w[best])
	p.Wurf = len(p.Halten)
	g.Wuerfel = w
	g.WurfNr++
	g.Wer = p.ID

	if p.Wurf >= bostonWuerfe {
		p.Done = true
		p.Punkte = bostonSumme(p.Halten)
	}
	return w
}

// bostonNaechster liefert den nächsten Spieler, der in dieser Runde noch nicht
// fertig ist – oder -1, wenn alle gewürfelt haben.
func bostonNaechster(g *BostonGame, ab int) int {
	for i := 1; i <= len(g.Players); i++ {
		k := (ab + i) % len(g.Players)
		if !g.Players[k].Done {
			return k
		}
	}
	return -1
}

func bostonAlleFertig(g *BostonGame) bool {
	if len(g.Players) == 0 {
		return false
	}
	for _, p := range g.Players {
		if !p.Done {
			return false
		}
	}
	return true
}

// bostonNamen verbindet Namen für eine Meldung: "Anna", "Anna und Ben",
// "Anna, Ben und Carl".
func bostonNamen(namen []string) string {
	switch len(namen) {
	case 0:
		return "niemand"
	case 1:
		return namen[0]
	case 2:
		return namen[0] + " und " + namen[1]
	default:
		return strings.Join(namen[:len(namen)-1], ", ") + " und " + namen[len(namen)-1]
	}
}

// bostonAuswerten vergleicht die Summen, verteilt die Siegpunkte und zeigt die
// Ergebnisübersicht. Bei Gleichstand bekommen alle Besten einen Punkt.
func bostonAuswerten(g *BostonGame) {
	best := 0
	for _, p := range g.Players {
		if p.Punkte > best {
			best = p.Punkte
		}
	}
	r := &BostonRunde{Nr: g.Runde, Summen: map[string]int{}, Best: best}
	for i := range g.Players {
		p := &g.Players[i]
		r.Summen[p.ID] = p.Punkte
		if p.Punkte == best {
			p.Siege++
			r.Sieger = append(r.Sieger, p.ID)
			r.SiegerNamen = append(r.SiegerNamen, p.Name)
		}
	}
	g.Letzte = r
	g.Status = "runde"
	if len(r.SiegerNamen) == 1 {
		g.Note = "Runde " + strconv.Itoa(g.Runde) + ": " + r.SiegerNamen[0] +
			" gewinnt mit " + strconv.Itoa(best) + " Punkten."
	} else {
		g.Note = "Runde " + strconv.Itoa(g.Runde) + ": " + bostonNamen(r.SiegerNamen) +
			" teilen sich den Sieg mit " + strconv.Itoa(best) + " Punkten."
	}
}

// bostonWeiter schaltet zur nächsten Runde oder beendet die Partie.
func bostonWeiter(g *BostonGame) {
	if g.Runde >= g.Runden {
		best := 0
		for _, p := range g.Players {
			if p.Siege > best {
				best = p.Siege
			}
		}
		g.Sieger = nil
		g.SiegerNamen = nil
		for _, p := range g.Players {
			if p.Siege == best {
				g.Sieger = append(g.Sieger, p.ID)
				g.SiegerNamen = append(g.SiegerNamen, p.Name)
			}
		}
		g.Status = "finished"
		if len(g.SiegerNamen) == 1 {
			g.Note = "Partie beendet: " + g.SiegerNamen[0] + " gewinnt mit " +
				strconv.Itoa(best) + " Siegpunkten."
		} else {
			g.Note = "Partie beendet: " + bostonNamen(g.SiegerNamen) +
				" teilen sich den Sieg mit " + strconv.Itoa(best) + " Siegpunkten."
		}
		return
	}
	g.Runde++
	bostonRundeStarten(g)
}

// bostonRaus nimmt einen Platz aus der Partie.
func bostonRaus(g *BostonGame, id, name string) (bool, bool, int) {
	idx := bostonFind(g, id)
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
		if bostonAlleFertig(g) {
			bostonAuswerten(g)
		} else if g.Players[g.Current].Done {
			if n := bostonNaechster(g, g.Current); n >= 0 {
				g.Current = n
			}
		}
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}

func bostonHandler(w http.ResponseWriter, r *http.Request) {
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
		bostonBotTreiben(room, time.Now())
		sicht := bostonSicht(room.Boston)
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
		Runden int    `json:"runden"`
		Stufe  string `json:"stufe"`
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
		if room.Boston != nil && room.Boston.Status != "finished" {
			fehler(http.StatusConflict, "In diesem Spielraum läuft bereits eine Boston-Runde.")
			return
		}
		runden := 3
		if input.Runden != 0 {
			gewaehlt, ok := bostonRundenWert(input.Runden)
			if !ok {
				fehler(http.StatusBadRequest, "Bitte 1, 3, 5 oder 10 Runden wählen.")
				return
			}
			runden = gewaehlt
		}
		game := &BostonGame{ID: randomCode(), HostID: input.ID, Status: "invited", Runden: runden, Runde: 1}
		game.Players = append(game.Players, bostonNewPlayer(input.ID, player.Name, 0))
		game.Note = player.Name + " eröffnet Auf nach Boston (" +
			strconv.Itoa(runden) + " Runden)."
		room.Boston = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet Auf nach Boston – " + strconv.Itoa(runden) + " Runden, drei Würfel, drei Würfe je Spieler.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "join":
		game := room.Boston
		if game == nil || game.Status != "invited" {
			fehler(http.StatusConflict, "Es läuft gerade keine offene Boston-Runde.")
			return
		}
		if bostonFind(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= bostonMaxPlayers {
			fehler(http.StatusConflict, "Diese Runde ist mit 6 Spielern voll.")
			return
		}
		game.Players = append(game.Players, bostonNewPlayer(input.ID, player.Name, len(game.Players)))
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zu Auf nach Boston.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "bot":
		room.botNamenNeu()
		if room.Boston != nil && room.Boston.Status != "finished" {
			fehler(http.StatusConflict, "In diesem Spielraum läuft bereits eine Boston-Runde.")
			return
		}
		stufe := botStufe(input.Stufe)
		game := &BostonGame{
			ID: randomCode(), HostID: input.ID, Status: "invited",
			Runden: 3, Runde: 1,
			BotID: botID(1), Bots: []string{botID(1)}, BotStufe: stufe,
		}
		game.Players = append(game.Players, bostonNewPlayer(input.ID, player.Name, 0))
		game.Players = append(game.Players, bostonNewPlayer(botID(1), room.botName(1), 1))
		for i := range game.Players {
			game.Players[i].Ready = true
		}
		game.Note = player.Name + " spielt gegen " + room.botName(1) +
			" (" + botStufeText(stufe) + ")."
		room.Boston = game
		if status, meldung := bostonAktion(room, input.ID, "start"); meldung != "" {
			fehler(status, meldung)
			return
		}
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Auf nach Boston gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	default:
		if status, meldung := bostonAktion(room, input.ID, action); meldung != "" {
			fehler(status, meldung)
			return
		}
	}

	room.LastActive = time.Now()
	sicht := bostonSicht(room.Boston)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": sicht})
}

// bostonAktion führt einen Spielzug aus – für Menschen und Bots dieselbe
// Prüfung. create und join bleiben im Handler, weil sie Meldungen in den Raum
// schreiben; alles andere steht hier.
func bostonAktion(room *Room, id, aktion string) (int, string) {
	game := room.Boston
	switch aktion {
	case "ready":
		if game == nil || game.Status != "invited" {
			return http.StatusConflict, "Bereit drücken geht nur, solange die Runde noch nicht läuft."
		}
		seat := bostonFind(game, id)
		if seat < 0 {
			return http.StatusForbidden, "Du sitzt nicht an diesem Tisch."
		}
		game.Players[seat].Ready = !game.Players[seat].Ready
		if game.Players[seat].Ready {
			game.Note = game.Players[seat].Name + " ist bereit."
		} else {
			game.Note = game.Players[seat].Name + " wartet noch."
		}

	case "start":
		if game == nil || game.Status != "invited" {
			return http.StatusConflict, "Es wartet keine Boston-Runde auf den Start."
		}
		if game.HostID != id {
			return http.StatusForbidden, "Nur der Gastgeber kann die Runde starten."
		}
		if len(game.Players) < 2 {
			return http.StatusConflict, "Auf nach Boston braucht mindestens zwei Spieler am Tisch."
		}
		for _, p := range game.Players {
			if !p.Ready {
				return http.StatusConflict, "Erst müssen alle Spieler BEREIT drücken."
			}
		}
		for i := range game.Players {
			game.Players[i].Siege = 0
		}
		game.Runde = 1
		game.Letzte = nil
		game.Sieger = nil
		game.SiegerNamen = nil
		bostonRundeStarten(game)

	case "roll":
		if game == nil || game.Status != "active" {
			return http.StatusConflict, "Es läuft gerade keine aktive Boston-Runde."
		}
		seat := bostonFind(game, id)
		if seat < 0 {
			return http.StatusForbidden, "Du sitzt nicht an diesem Tisch."
		}
		if seat != game.Current {
			return http.StatusConflict, "Du bist nicht dran – " +
				game.Players[game.Current].Name + " würfelt gerade."
		}
		if game.Players[seat].Done {
			return http.StatusConflict, "Dein Zug in dieser Runde ist schon fertig."
		}
		bostonWurf(game, seat)
		p := &game.Players[seat]
		if p.Done {
			game.Note = p.Name + " hat " + strconv.Itoa(p.Punkte) + " Punkte."
		} else {
			game.Note = p.Name + " hält " + strconv.Itoa(p.Halten[len(p.Halten)-1]) +
				" – Zwischensumme " + strconv.Itoa(bostonSumme(p.Halten)) + "."
		}
		// Ein Spieler macht alle drei Würfe hintereinander – gewechselt wird
		// erst, wenn sein Zug fertig ist.
		if bostonAlleFertig(game) {
			bostonAuswerten(game)
		} else if p.Done {
			if n := bostonNaechster(game, seat); n >= 0 {
				game.Current = n
			}
		}

	case "weiter":
		if game == nil || game.Status != "runde" {
			return http.StatusConflict, "Die nächste Runde geht erst nach der Ergebnisübersicht."
		}
		if bostonFind(game, id) < 0 {
			return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
		}
		bostonWeiter(game)

	case "reset":
		if game == nil || game.Status != "finished" {
			return http.StatusConflict, "Eine neue Partie ist erst nach Spielende möglich."
		}
		if bostonFind(game, id) < 0 {
			return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
		}
		fresh := &BostonGame{ID: game.ID, HostID: game.HostID, Status: "invited", Runden: game.Runden, Runde: 1}
		for i, p := range game.Players {
			fresh.Players = append(fresh.Players, bostonNewPlayer(p.ID, p.Name, i))
		}
		fresh.Note = "Neue Partie – alle drücken BEREIT, dann startet der Gastgeber."
		room.Boston = fresh

	default:
		return http.StatusBadRequest, "Unbekannte Boston-Aktion."
	}
	return 0, ""
}
