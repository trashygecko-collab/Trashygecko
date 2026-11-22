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

// Chicago (bayerische Fassung mit Bierfilzen): drei Würfel, bis zu drei Würfe.
// Eins zählt 100, Sechs zählt 60, zwei bis fünf ihren Wert. Wer vorlegt,
// bestimmt mit seiner Wurfzahl, wie viele Würfe die Runde hat; jeder Nachfolger
// muss den Wert überbieten („mit ist Shit“ – Gleichstand genügt nicht).
//
// Phase 1: Es sind ein Bierfilz mehr im Spiel als Spieler. Jede Runde bekommt
// der Tiefste einen Filz. Phase 2: Wer den höchsten Wurf hat, darf einen Filz
// abgeben. Wer zuletzt Filze hat, verliert. Ein „Chic“ (drei Einser in einem
// Wurf) befreit sofort von allem.
const chMaxPlayers = 6

var chColors = []string{"#e8cd8e", "#7fb98a", "#d98878", "#8aa9d6", "#b18cff", "#ff985e"}

type ChPlayer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Ready  bool   `json:"ready"`
	Deckel int    `json:"deckel"` // Bierfilze
	Halten []int  `json:"halten"` // liegende Würfel dieser Runde
	Wuerfe int    `json:"wuerfe"` // verbrauchte Würfe dieser Runde
	Wert   int    `json:"wert"`   // Rundenergebnis
	Fertig bool   `json:"fertig"`
	Chic   bool   `json:"chic"`
	Raus   bool   `json:"raus"` // in Phase 2 ohne Filz
}

// ChErgebnis ist eine Zeile der Rundenübersicht.
type ChErgebnis struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Wert   int    `json:"wert"`
	Wuerfe int    `json:"wuerfe"`
	Deckel int    `json:"deckel"`
	Chic   bool   `json:"chic"`
}

type ChRunde struct {
	Nr         int          `json:"nr"`
	Phase      int          `json:"phase"`
	Wurfzahl   int          `json:"wurfzahl"`
	Ergebnisse []ChErgebnis `json:"ergebnisse"`
	Verlierer  string       `json:"verlierer,omitempty"` // Phase 1
	Bester     string       `json:"bester,omitempty"`    // Phase 2
	Chic       string       `json:"chic,omitempty"`
	Filze      int          `json:"filze"` // wie viele Filze nach der Runde verteilt sind
}

type ChicagoGame struct {
	ID        string     `json:"id"`
	HostID    string     `json:"hostId"`
	Players   []ChPlayer `json:"players"`
	Current   int        `json:"current"`
	Status    string     `json:"status"` // invited, active, runde, finished
	Phase     int        `json:"phase"`  // 1 oder 2
	Runde     int        `json:"runde"`
	Wurfzahl  int        `json:"wurfzahl"` // 0 = noch offen (der Vorleger bestimmt sie)
	Wuerfel   []int      `json:"wuerfel"`  // der letzte Wurf
	WurfNr    int        `json:"wurfNr"`
	Wer       string     `json:"wer"`
	Hoch      int        `json:"hoch"` // bester Wert der Runde
	HochID    string     `json:"hochId"`
	HochName  string     `json:"hochName"`
	TiefID    string     `json:"tiefId"`
	TiefName  string     `json:"tiefName"`
	Filze     int        `json:"filze"`    // verteilte Bierfilze
	FilzeMax  int        `json:"filzeMax"` // Spieler + 1
	Letzte    *ChRunde   `json:"letzte,omitempty"`
	Verlierer string     `json:"verlierer,omitempty"`
	Note      string     `json:"note,omitempty"`
	// Bot: Kennung, Stufe und die Zeit, zu der der Bot würfelt.
	BotID    string    `json:"botId,omitempty"`
	Bots     []string  `json:"bots,omitempty"`
	BotStufe string    `json:"botStufe,omitempty"`
	BotFrist time.Time `json:"-"`
}

func chNewPlayer(id, name string, seat int) ChPlayer {
	return ChPlayer{ID: id, Name: name, Color: chColors[seat%len(chColors)], Halten: []int{}}
}

func chFind(g *ChicagoGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// chWert rechnet die liegenden Würfel: Eins = 100, Sechs = 60, sonst Augenzahl.
func chWert(w []int) int {
	s := 0
	for _, a := range w {
		switch a {
		case 1:
			s += 100
		case 6:
			s += 60
		default:
			s += a
		}
	}
	return s
}

// chIstChic: drei Einser in einem Wurf.
func chIstChic(w []int) bool {
	return len(w) == 3 && w[0] == 1 && w[1] == 1 && w[2] == 1
}

// chWertText erklärt einen Wert in Worten ("zwei Einser und eine 4").
func chWertText(w []int) string {
	if len(w) == 0 {
		return "noch nichts"
	}
	teile := make([]string, 0, len(w))
	for _, a := range w {
		switch a {
		case 1:
			teile = append(teile, "1")
		case 6:
			teile = append(teile, "6")
		default:
			teile = append(teile, strconv.Itoa(a))
		}
	}
	return strings.Join(teile, " ") + " = " + strconv.Itoa(chWert(w))
}

// chSicht kopiert den Stand fürs JSON (Listen nie nil).
func chSicht(g *ChicagoGame) *ChicagoGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Players = make([]ChPlayer, len(g.Players))
	for i, p := range g.Players {
		c.Players[i] = p
		c.Players[i].Halten = append([]int{}, p.Halten...)
	}
	c.Wuerfel = append([]int{}, g.Wuerfel...)
	if g.Letzte != nil {
		l := *g.Letzte
		l.Ergebnisse = append([]ChErgebnis{}, g.Letzte.Ergebnisse...)
		c.Letzte = &l
	}
	return &c
}

// chAktiv sind die Spieler, die in dieser Runde noch würfeln müssen.
func chAktiv(g *ChicagoGame) []int {
	var out []int
	for i, p := range g.Players {
		if !p.Raus {
			out = append(out, i)
		}
	}
	return out
}

// chMaxWuerfe: wie viele Würfe der aktuelle Spieler noch hat.
func chMaxWuerfe(g *ChicagoGame) int {
	if g.Wurfzahl > 0 {
		return g.Wurfzahl
	}
	return 3
}

// chRundeStarten setzt Würfel und Ergebnisse zurück und bestimmt den Vorleger.
func chRundeStarten(g *ChicagoGame, start int) {
	for i := range g.Players {
		g.Players[i].Halten = []int{}
		g.Players[i].Wuerfe = 0
		g.Players[i].Wert = 0
		g.Players[i].Fertig = false
		g.Players[i].Chic = false
	}
	g.Wuerfel = []int{}
	g.Wer = ""
	g.Wurfzahl = 0
	g.Hoch = 0
	g.HochID = ""
	g.HochName = ""
	g.TiefID = ""
	g.TiefName = ""
	g.Status = "active"
	g.Current = start
	if g.Current < 0 || g.Current >= len(g.Players) || g.Players[g.Current].Raus {
		if n := chNaechster(g, start); n >= 0 {
			g.Current = n
		}
	}
	g.Note = "Runde " + strconv.Itoa(g.Runde) + " (Phase " + strconv.Itoa(g.Phase) + "): " +
		g.Players[g.Current].Name + " legt vor."
}

// chNaechster liefert den nächsten Spieler, der noch nicht fertig ist.
func chNaechster(g *ChicagoGame, ab int) int {
	for i := 1; i <= len(g.Players); i++ {
		k := (ab + i) % len(g.Players)
		if !g.Players[k].Raus && !g.Players[k].Fertig {
			return k
		}
	}
	return -1
}

func chAlleFertig(g *ChicagoGame) bool {
	for _, p := range g.Players {
		if !p.Raus && !p.Fertig {
			return false
		}
	}
	return true
}

// chWurfMit würfelt für einen Spieler. augen wird nur von den Tests gesetzt.
func chWurfMit(g *ChicagoGame, idx int, augen []int) ([]int, bool) {
	p := &g.Players[idx]
	n := 3 - len(p.Halten)
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
	p.Wuerfe++
	g.Wuerfel = w
	g.WurfNr++
	g.Wer = p.ID

	chic := chIstChic(w)
	if chic {
		p.Chic = true
		p.Fertig = true
		p.Wert = chWert(w)
		g.Note = p.Name + " würfelt einen Chic – drei Einser auf einen Wurf!"
		return w, true
	}
	g.Note = p.Name + " würfelt " + chWertText(w)
	return w, false
}

// chWeitergeben schließt den Zug ab: alle noch liegenden Würfel des letzten
// Wurfs zählen mit (wer nichts herauslegt, lässt alles liegen).
func chWeitergeben(g *ChicagoGame, idx int) {
	p := &g.Players[idx]
	// Wer nichts herausgelegt hat, lässt alles liegen – aber nie mehr als drei.
	for _, a := range g.Wuerfel {
		if len(p.Halten) >= 3 {
			break
		}
		p.Halten = append(p.Halten, a)
	}
	g.Wuerfel = []int{}
	p.Wert = chWert(p.Halten)
	p.Fertig = true

	if g.Wurfzahl == 0 {
		// Der Vorleger bestimmt mit seiner Wurfzahl die Runde.
		g.Wurfzahl = p.Wuerfe
		if g.Wurfzahl < 1 {
			g.Wurfzahl = 1
		}
	}
	erster := g.HochName == ""
	ueberboten := erster || p.Wert > g.Hoch
	if ueberboten {
		g.Hoch = p.Wert
		g.HochID = p.ID
		g.HochName = p.Name
	}
	if erster || !ueberboten {
		// Nicht übertroffen – „Mit ist Shit“: Gleichstand genügt nicht.
		g.TiefID = p.ID
		g.TiefName = p.Name
	}
	g.Note = p.Name + " hat " + strconv.Itoa(p.Wert) + ". Es stehen " +
		strconv.Itoa(g.Hoch) + " (" + g.HochName + ")."
}

// chDrehenSechser dreht Sechser in Einsen um: aus zwei Sechsen wird eine Eins,
// aus drei Sechsen zwei Einsen. Die übrige Sechs wandert zurück in den Becher.
func chDrehenSechser(g *ChicagoGame) bool {
	sechsen := 0
	for _, a := range g.Wuerfel {
		if a == 6 {
			sechsen++
		}
	}
	if sechsen < 2 {
		return false
	}
	einsen := sechsen - 1
	neu := make([]int, 0, len(g.Wuerfel))
	for i := 0; i < einsen; i++ {
		neu = append(neu, 1)
	}
	for _, a := range g.Wuerfel {
		if a != 6 {
			neu = append(neu, a)
		}
	}
	g.Wuerfel = neu
	return true
}

// chAuswerten verteilt bzw. entfernt am Rundenende die Bierfilze.
func chAuswerten(g *ChicagoGame) {
	r := &ChRunde{Nr: g.Runde, Phase: g.Phase, Wurfzahl: g.Wurfzahl}
	for i := range g.Players {
		p := &g.Players[i]
		r.Ergebnisse = append(r.Ergebnisse, ChErgebnis{
			ID: p.ID, Name: p.Name, Wert: p.Wert, Wuerfe: p.Wuerfe,
			Deckel: p.Deckel, Chic: p.Chic,
		})
		if p.Chic {
			r.Chic = p.Name
			// Ein Chic nimmt alle bisherigen Filze aus dem Spiel.
			if p.Deckel > 0 {
				g.Filze -= p.Deckel
				if g.Filze < 0 {
					g.Filze = 0
				}
				p.Deckel = 0
			}
		}
	}

	if g.Phase == 1 {
		// Der Tiefste bekommt einen Filz (Chic-Spieler nie). Tief ist der, an
		// dem das „Tief“ hängen geblieben ist – nicht einfach der kleinste Wert,
		// denn wer nur gleichzieht, ist nach „Mit ist Shit“ selbst tief.
		best := chFind(g, g.TiefID)
		if best < 0 || g.Players[best].Chic || g.Players[best].Raus {
			best = -1
			bestWert := -1
			for i := range g.Players {
				p := &g.Players[i]
				if p.Chic || p.Raus {
					continue
				}
				if best < 0 || p.Wert < bestWert || (p.Wert == bestWert && i > best) {
					best = i
					bestWert = p.Wert
				}
			}
		}
		if best >= 0 {
			g.Players[best].Deckel++
			g.Filze++
			r.Verlierer = g.Players[best].Name
			g.Note = g.Players[best].Name + " ist tief und bekommt einen Bierfilz (" +
				strconv.Itoa(g.Players[best].Wert) + ")."
		}
	} else {
		// Phase 2: Der Beste darf einen Filz abgeben.
		best := chFind(g, g.HochID)
		if best < 0 || g.Players[best].Chic || g.Players[best].Raus || g.Players[best].Deckel == 0 {
			best = -1
			bestWert := -1
			for i := range g.Players {
				p := &g.Players[i]
				if p.Chic || p.Raus || p.Deckel == 0 {
					continue
				}
				if best < 0 || p.Wert > bestWert || (p.Wert == bestWert && i < best) {
					best = i
					bestWert = p.Wert
				}
			}
		}
		if best >= 0 {
			g.Players[best].Deckel--
			g.Filze--
			r.Bester = g.Players[best].Name
			g.Note = g.Players[best].Name + " hatte den besten Wurf und gibt einen Bierfilz ab."
		}
	}

	// Die Übersicht zeigt die Filze nach der Runde.
	for i := range r.Ergebnisse {
		for j := range g.Players {
			if g.Players[j].ID == r.Ergebnisse[i].ID {
				r.Ergebnisse[i].Deckel = g.Players[j].Deckel
			}
		}
	}
	r.Filze = g.Filze
	g.Letzte = r
	g.Status = "runde"

	// Wer ohne Filz ist, ist in Phase 2 raus.
	if g.Phase == 2 {
		for i := range g.Players {
			if g.Players[i].Deckel == 0 {
				g.Players[i].Raus = true
			}
		}
		uebrig := 0
		letzter := ""
		for i := range g.Players {
			if g.Players[i].Deckel > 0 {
				uebrig++
				letzter = g.Players[i].Name
			}
		}
		if uebrig <= 1 {
			// Die Rundenübersicht bleibt stehen; erst „weiter“ beendet die Partie.
			g.Verlierer = letzter
			if letzter != "" {
				g.Note = letzter + " hat als letzter noch einen Bierfilz – die Runde ist fällig."
			}
		}
	}
}

// chWeiter beginnt die nächste Runde oder wechselt in Phase 2.
func chWeiter(g *ChicagoGame) {
	// Steht der Verlierer schon fest, ist die Partie vorbei.
	if g.Phase == 2 && g.Verlierer != "" {
		g.Status = "finished"
		return
	}
	// Ein Chic kann in Phase 2 alle Filze vom Tisch nehmen. Dann hat niemand
	// mehr einen – die Partie endet ohne Verlierer, statt in einer Runde zu
	// hängen, in der keiner mehr würfeln darf.
	if g.Phase == 2 {
		mitFilz := 0
		for _, p := range g.Players {
			if p.Deckel > 0 {
				mitFilz++
			}
		}
		if mitFilz == 0 {
			g.Status = "finished"
			g.Note = "Alle Bierfilze sind weg – die Partie endet ohne Verlierer."
			return
		}
	}
	// Phase 1 endet, sobald alle Filze verteilt sind.
	if g.Phase == 1 && g.Filze >= g.FilzeMax {
		g.Phase = 2
		g.Runde = 1
		for i := range g.Players {
			g.Players[i].Raus = g.Players[i].Deckel == 0
		}
		// Es beginnen die Spieler mit den meisten Filzen.
		start := 0
		meiste := -1
		uebrig := 0
		letzter := ""
		for i, p := range g.Players {
			if p.Deckel > meiste {
				meiste = p.Deckel
				start = i
			}
			if p.Deckel > 0 {
				uebrig++
				letzter = p.Name
			}
		}
		// Hat nur einer Filze, ist die Partie entschieden – dann wird nicht
		// mehr allein weitergewürfelt.
		if uebrig <= 1 {
			g.Status = "finished"
			g.Verlierer = letzter
			g.Note = letzter + " hat als einziger noch Bierfilze – die Runde ist fällig."
			return
		}
		g.Runde = 1
		chRundeStarten(g, start)
		return
	}

	// Nächster Vorleger: in Phase 1 der Verlierer, in Phase 2 der beste Spieler.
	start := g.Current
	if g.Letzte != nil {
		namen := g.Letzte.Verlierer
		if g.Phase == 2 {
			namen = g.Letzte.Bester
		}
		if namen != "" {
			for i, p := range g.Players {
				if p.Name == namen {
					start = i
				}
			}
		}
	}
	g.Runde++
	chRundeStarten(g, start)
}

// chRaus nimmt einen Platz aus der Partie.
func chRaus(g *ChicagoGame, id, name string) (bool, bool, int) {
	idx := chFind(g, id)
	if idx < 0 {
		return false, false, len(g.Players)
	}
	hatte := g.Players[idx]
	g.Players = append(g.Players[:idx], g.Players[idx+1:]...)
	if len(g.Players) < 2 {
		return true, true, len(g.Players)
	}
	g.FilzeMax = len(g.Players) + 1
	if hatte.Deckel > 0 {
		g.Filze -= hatte.Deckel
		if g.Filze < 0 {
			g.Filze = 0
		}
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
		if chAlleFertig(g) {
			chAuswerten(g)
		} else if g.Players[g.Current].Fertig || g.Players[g.Current].Raus {
			if n := chNaechster(g, g.Current); n >= 0 {
				g.Current = n
			}
		}
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}

func chicagoHandler(w http.ResponseWriter, r *http.Request) {
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
		chicagoBotTreiben(room, time.Now())
		sicht := chSicht(room.Chicago)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": sicht})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID      string `json:"id"`
		Indizes []int  `json:"indizes"`
		Stufe   string `json:"stufe"`
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
		if room.Chicago != nil && room.Chicago.Status != "finished" {
			fehler(http.StatusConflict, "In diesem Spielraum läuft bereits eine Chicago-Runde.")
			return
		}
		game := &ChicagoGame{ID: randomCode(), HostID: input.ID, Status: "invited", Phase: 1, Runde: 1}
		game.Players = append(game.Players, chNewPlayer(input.ID, player.Name, 0))
		game.FilzeMax = 2
		game.Note = player.Name + " eröffnet Chicago – es geht um die Bierfilze."
		room.Chicago = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet Chicago – drei Würfel, ein Bierfilz zu viel.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "join":
		game := room.Chicago
		if game == nil || game.Status != "invited" {
			fehler(http.StatusConflict, "Es läuft gerade keine offene Chicago-Runde.")
			return
		}
		if chFind(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= chMaxPlayers {
			fehler(http.StatusConflict, "Diese Runde ist mit 6 Spielern voll.")
			return
		}
		game.Players = append(game.Players, chNewPlayer(input.ID, player.Name, len(game.Players)))
		game.FilzeMax = len(game.Players) + 1
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zu Chicago.",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	case "bot":
		room.botNamenNeu()
		if room.Chicago != nil && room.Chicago.Status != "finished" {
			fehler(http.StatusConflict, "In diesem Spielraum läuft bereits eine Chicago-Runde.")
			return
		}
		stufe := botStufe(input.Stufe)
		game := &ChicagoGame{
			ID: randomCode(), HostID: input.ID, Status: "invited",
			Phase: 1, Runde: 1,
			BotID: botID(1), Bots: []string{botID(1)}, BotStufe: stufe,
		}
		game.Players = append(game.Players, chNewPlayer(input.ID, player.Name, 0))
		game.Players = append(game.Players, chNewPlayer(botID(1), room.botName(1), 1))
		for i := range game.Players {
			game.Players[i].Ready = true
		}
		game.FilzeMax = len(game.Players) + 1
		game.Note = player.Name + " spielt gegen " + room.botName(1) +
			" (" + botStufeText(stufe) + ")."
		room.Chicago = game
		if status, meldung := chicagoAktion(room, input.ID, "start", nil); meldung != "" {
			fehler(status, meldung)
			return
		}
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Chicago gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})

	default:
		if status, meldung := chicagoAktion(room, input.ID, action, input.Indizes); meldung != "" {
			fehler(status, meldung)
			return
		}
	}

	room.LastActive = time.Now()
	sicht := chSicht(room.Chicago)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": sicht})
}

// chNaechsterOderAuswerten wechselt zum nächsten Spieler oder wertet die Runde aus.
func chNaechsterOderAuswerten(g *ChicagoGame, ab int) {
	if chAlleFertig(g) {
		chAuswerten(g)
		return
	}
	if n := chNaechster(g, ab); n >= 0 {
		g.Current = n
	}
}

// chSortiert liefert die Ergebnisse einer Runde nach Wert (absteigend).
func chSortiert(g *ChicagoGame) []ChErgebnis {
	out := []ChErgebnis{}
	if g.Letzte == nil {
		return out
	}
	out = append(out, g.Letzte.Ergebnisse...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Wert > out[j].Wert })
	return out
}

// chicagoAktion führt einen Spielzug aus – für Menschen und Bots dieselbe
// Prüfung. create und join bleiben im Handler, weil sie Meldungen in den Raum
// schreiben; alles andere steht hier.
func chicagoAktion(room *Room, id, aktion string, indizes []int) (int, string) {
	game := room.Chicago
	switch aktion {
	case "ready":
		if game == nil || game.Status != "invited" {
			return http.StatusConflict, "Bereit drücken geht nur, solange die Runde noch nicht läuft."
		}
		seat := chFind(game, id)
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
			return http.StatusConflict, "Es wartet keine Chicago-Runde auf den Start."
		}
		if game.HostID != id {
			return http.StatusForbidden, "Nur der Gastgeber kann die Runde starten."
		}
		if len(game.Players) < 2 {
			return http.StatusConflict, "Chicago braucht mindestens zwei Spieler am Tisch."
		}
		for _, p := range game.Players {
			if !p.Ready {
				return http.StatusConflict, "Erst müssen alle Spieler BEREIT drücken."
			}
		}
		game.Phase = 1
		game.Runde = 1
		game.Filze = 0
		game.FilzeMax = len(game.Players) + 1
		game.Letzte = nil
		game.Verlierer = ""
		for i := range game.Players {
			game.Players[i].Deckel = 0
			game.Players[i].Raus = false
		}
		chRundeStarten(game, 0)

	case "wurf":
		if game == nil || game.Status != "active" {
			return http.StatusConflict, "Es läuft gerade keine aktive Chicago-Runde."
		}
		seat := chFind(game, id)
		if seat < 0 {
			return http.StatusForbidden, "Du sitzt nicht an diesem Tisch."
		}
		if seat != game.Current {
			return http.StatusConflict, "Du bist nicht dran - " +
				game.Players[game.Current].Name + " würfelt gerade."
		}
		p := &game.Players[seat]
		if p.Fertig || p.Raus {
			return http.StatusConflict, "Dein Zug in dieser Runde ist schon fertig."
		}
		if p.Wuerfe >= chMaxWuerfe(game) {
			return http.StatusConflict, "Mehr Würfe gibt diese Runde nicht her."
		}
		chWurfMit(game, seat, nil)
		p = &game.Players[seat]
		if p.Chic {
			if game.Wurfzahl == 0 {
				game.Wurfzahl = p.Wuerfe
			}
			chNaechsterOderAuswerten(game, seat)
			break
		}
		if p.Wuerfe >= chMaxWuerfe(game) {
			chWeitergeben(game, seat)
			chNaechsterOderAuswerten(game, seat)
		}

	case "halten":
		if game == nil || game.Status != "active" {
			return http.StatusConflict, "Es läuft gerade keine aktive Chicago-Runde."
		}
		seat := chFind(game, id)
		if seat < 0 {
			return http.StatusForbidden, "Du sitzt nicht an diesem Tisch."
		}
		if seat != game.Current || len(game.Wuerfel) == 0 || game.Wer != id {
			return http.StatusConflict, "Erst würfeln, dann Würfel liegen lassen."
		}
		p := &game.Players[seat]
		if p.Fertig {
			return http.StatusConflict, "Dein Zug ist schon fertig."
		}
		gesehen := map[int]bool{}
		for _, i := range indizes {
			if i < 0 || i >= len(game.Wuerfel) || gesehen[i] {
				continue
			}
			gesehen[i] = true
			p.Halten = append(p.Halten, game.Wuerfel[i])
		}
		if len(p.Halten) >= 3 {
			game.Wuerfel = []int{}
			chWeitergeben(game, seat)
			chNaechsterOderAuswerten(game, seat)
			break
		}
		game.Wuerfel = []int{}
		game.Note = p.Name + " lässt " + strconv.Itoa(len(p.Halten)) + " Würfel liegen."

	case "drehen":
		if game == nil || game.Status != "active" || game.Wer != id || len(game.Wuerfel) == 0 {
			return http.StatusConflict, "Drehen geht nur direkt nach dem eigenen Wurf."
		}
		if !chDrehenSechser(game) {
			return http.StatusConflict, "Zum Drehen braucht es zwei Sechser im Wurf."
		}
		game.Note = game.Players[game.Current].Name + " dreht Sechser zu Einsen."

	case "weitergeben":
		if game == nil || game.Status != "active" {
			return http.StatusConflict, "Es läuft gerade keine aktive Chicago-Runde."
		}
		seat := chFind(game, id)
		if seat < 0 {
			return http.StatusForbidden, "Du sitzt nicht an diesem Tisch."
		}
		if seat != game.Current || game.Wer != id || len(game.Wuerfel) == 0 {
			return http.StatusConflict, "Erst würfeln, dann weitergeben."
		}
		if game.Players[seat].Fertig {
			return http.StatusConflict, "Dein Zug ist schon fertig."
		}
		chWeitergeben(game, seat)
		chNaechsterOderAuswerten(game, seat)

	case "weiter":
		if game == nil || game.Status != "runde" {
			return http.StatusConflict, "Die nächste Runde geht erst nach der Rundenübersicht."
		}
		if chFind(game, id) < 0 {
			return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
		}
		chWeiter(game)

	case "reset":
		if game == nil || game.Status != "finished" {
			return http.StatusConflict, "Eine neue Partie ist erst nach Spielende möglich."
		}
		if chFind(game, id) < 0 {
			return http.StatusForbidden, "Du gehörst nicht zu dieser Partie."
		}
		fresh := &ChicagoGame{ID: game.ID, HostID: game.HostID, Status: "invited", Phase: 1, Runde: 1}
		for i, p := range game.Players {
			fresh.Players = append(fresh.Players, chNewPlayer(p.ID, p.Name, i))
		}
		fresh.FilzeMax = len(fresh.Players) + 1
		fresh.Note = "Neue Partie – alle drücken BEREIT, dann startet der Gastgeber."
		room.Chicago = fresh

	default:
		return http.StatusBadRequest, "Unbekannte Chicago-Aktion."
	}
	return 0, ""
}
