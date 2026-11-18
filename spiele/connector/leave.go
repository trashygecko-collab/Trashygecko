package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// leaveGameHandler beendet sofort alle Partien, in denen dieser Spieler sitzt.
// Der Spieler bleibt im Raum; verlassen wird ausschließlich das Spiel.
func leaveGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erforderlich", http.StatusMethodNotAllowed)
		return
	}
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	var input struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	id := strings.TrimSpace(input.ID)
	if code == "" || id == "" {
		http.Error(w, "Spielraum oder Spieler fehlt", http.StatusBadRequest)
		return
	}

	mu.Lock()
	room := rooms[code]
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	room.cleanPlayers()
	player, ok := room.Players[id]
	if !ok {
		mu.Unlock()
		http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
		return
	}
	ende, text := room.abgangBeenden(id, player.Name, "verlassen")
	if ende {
		room.LastActive = time.Now()
	}
	mu.Unlock()

	links := 0
	if ende {
		links = 1
	}
	jsonResponse(w, map[string]any{"ok": true, "left": links, "ended": ende, "text": text})
}

// partie ist ein Eintrag in der Liste aller Partien eines Raums. Alles, was
// mit Abgängen zu tun hat, läuft über diese eine Liste – kein Spiel wird an
// zwei Stellen behandelt.
type partie struct {
	name string
	// betroffen meldet, ob dieser Spieler in der Partie sitzt, und ob sie
	// danach zu Ende ist. Bei Tischen mit Sitzplatzverwaltung entfernt der
	// Aufruf zugleich den Platz; der Tisch kann mit genug Spielern weiterlaufen.
	betroffen func(id, name string) (drin, fertig bool)
	ende      func()
}

// partien listet alle Partien des Raums mit ihrer Abgangsregel.
func (r *Room) partien() []partie {
	// duell: zwei feste Kennungen – fehlt einer, ist die Partie vorbei.
	duell := func(name string, ids func() (string, string), ende func()) partie {
		return partie{
			name: name,
			betroffen: func(id, _ string) (bool, bool) {
				a, b := ids()
				return id != "" && (a == id || b == id), true
			},
			ende: ende,
		}
	}
	// sitz: Tisch mit Sitzplatzverwaltung.
	sitz := func(name string, weg func(id, name string) (bool, bool, int), ende func()) partie {
		return partie{
			name: name,
			betroffen: func(id, n string) (bool, bool) {
				drin, fertig, _ := weg(id, n)
				return drin, fertig
			},
			ende: ende,
		}
	}
	// ohneSitz: Tisch ohne sichere Sitzplatzverwaltung – endet ganz.
	ohneSitz := func(name string, ids func() []string, ende func()) partie {
		return partie{
			name:      name,
			betroffen: func(id, _ string) (bool, bool) { return enthaelt(ids(), id), true },
			ende:      ende,
		}
	}

	return []partie{
		duell("Tic Tac Toe",
			func() (string, string) {
				if g := r.TicTacToe; g != nil {
					return g.XID, g.OID
				}
				return "", ""
			},
			func() { r.TicTacToe = nil }),
		duell("Vier gewinnt",
			func() (string, string) {
				if g := r.VierGewinnt; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.VierGewinnt = nil }),
		duell("Dame",
			func() (string, string) {
				if g := r.Dame; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Dame = nil }),
		duell("Paare",
			func() (string, string) {
				if g := r.Paare; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Paare = nil }),
		duell("Käsekästchen",
			func() (string, string) {
				if g := r.Kaese; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Kaese = nil }),
		duell("Sim",
			func() (string, string) {
				if g := r.Sim; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Sim = nil }),
		duell("Kalaha",
			func() (string, string) {
				if g := r.Kalaha; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Kalaha = nil }),
		duell("Streichholzspiel",
			func() (string, string) {
				if g := r.Streichholz; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Streichholz = nil }),
		duell("Mühle",
			func() (string, string) {
				if g := r.Muehle; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Muehle = nil }),
		duell("Reversi",
			func() (string, string) {
				if g := r.Reversi; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Reversi = nil }),
		duell("Hold",
			func() (string, string) {
				if g := r.Hold; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Hold = nil }),
		duell("Skyjo",
			func() (string, string) {
				if g := r.Skyjo; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Skyjo = nil }),
		duell("Schach",
			func() (string, string) {
				if g := r.Schach; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Schach = nil }),
		duell("Go",
			func() (string, string) {
				if g := r.Go; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Go = nil }),
		duell("Backgammon",
			func() (string, string) {
				if g := r.Backgammon; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Backgammon = nil }),
		duell("Tetris",
			func() (string, string) {
				if g := r.Tetris; g != nil {
					return g.RedID, g.BlueID
				}
				return "", ""
			},
			func() { r.Tetris = nil }),

		sitz("Mau-Mau",
			func(id, n string) (bool, bool, int) {
				if g := r.MauMau; g != nil {
					return maumauRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.MauMau = nil }),
		sitz("UNO",
			func(id, n string) (bool, bool, int) {
				if g := r.Uno; g != nil {
					return unoRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Uno = nil }),
		sitz("Schwimmen",
			func(id, n string) (bool, bool, int) {
				if g := r.Schwimmen; g != nil {
					return schwimmRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Schwimmen = nil }),
		sitz("Kniffel",
			func(id, n string) (bool, bool, int) {
				if g := r.Kniffel; g != nil {
					return kniffelRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Kniffel = nil }),
		sitz("Arschloch",
			func(id, n string) (bool, bool, int) {
				if g := r.Arschloch; g != nil {
					return arschlochRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Arschloch = nil }),
		sitz("Galgenmännchen",
			func(id, n string) (bool, bool, int) {
				if g := r.Galgen; g != nil {
					return galgenRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Galgen = nil }),
		sitz("Bingo",
			func(id, n string) (bool, bool, int) {
				if g := r.Bingo; g != nil {
					return bingoRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Bingo = nil }),
		sitz("Wortsuche",
			func(id, n string) (bool, bool, int) {
				if g := r.Wortsuche; g != nil {
					return wortsucheRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Wortsuche = nil }),
		sitz("Domino",
			func(id, n string) (bool, bool, int) {
				if g := r.Domino; g != nil {
					return dominoRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Domino = nil }),
		sitz("Schlangen und Leitern",
			func(id, n string) (bool, bool, int) {
				if g := r.Schlangen; g != nil {
					return schlangenRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Schlangen = nil }),
		sitz("Pig",
			func(id, n string) (bool, bool, int) {
				if g := r.Pig; g != nil {
					return pigRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Pig = nil }),
		sitz("Gänsespiel",
			func(id, n string) (bool, bool, int) {
				if g := r.Gaense; g != nil {
					return gaenseRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Gaense = nil }),
		sitz("Mensch ärgere dich nicht",
			func(id, n string) (bool, bool, int) {
				if g := r.Menschaerger; g != nil {
					return maenRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Menschaerger = nil }),
		sitz("1x1",
			func(id, n string) (bool, bool, int) {
				if g := r.Rechen; g != nil {
					return rechenRemoveSeat(g, id, n)
				}
				return false, false, 0
			},
			func() { r.Rechen = nil }),

		ohneSitz("Poker",
			func() []string {
				if g := r.Poker; g != nil {
					return idsVon(g.Spieler, func(p PokerSpieler) string { return p.ID })
				}
				return nil
			},
			func() { r.Poker = nil }),
		ohneSitz("Cabo",
			func() []string {
				if g := r.Cabo; g != nil {
					return idsVon(g.Spieler, func(p CaboSpieler) string { return p.ID })
				}
				return nil
			},
			func() { r.Cabo = nil }),
		ohneSitz("Siebzehn",
			func() []string {
				if g := r.Siebzehn; g != nil {
					return idsVon(g.Spieler, func(p SiebzehnSpieler) string { return p.ID })
				}
				return nil
			},
			func() { r.Siebzehn = nil }),
		ohneSitz("Baccara",
			func() []string {
				if g := r.Baccara; g != nil {
					return idsVon(g.Spieler, func(p BaccaraSpieler) string { return p.ID })
				}
				return nil
			},
			func() { r.Baccara = nil }),
		ohneSitz("Schiffe versenken",
			func() []string {
				if g := r.Schiffe; g != nil {
					return idsVon(g.Spieler, func(p SchiffeSpieler) string { return p.ID })
				}
				return nil
			},
			func() { r.Schiffe = nil }),
		ohneSitz("Sechsundsechzig",
			func() []string {
				if g := r.Sechsundsechzig; g != nil {
					return idsVon(g.Spieler, func(p SechsundsechzigSpieler) string { return p.ID })
				}
				return nil
			},
			func() { r.Sechsundsechzig = nil }),
	}
}

// abgangBeenden löst alle Partien, in denen dieser Spieler sitzt. Es ist die
// einzige Stelle, an der Partien wegen eines Abgangs enden – gleich ob er
// ausdrücklich "Spiel verlassen" gedrückt hat (grund "verlassen"), ob seine
// Lebenszeichen ausgeblieben sind (grund "verbindung") oder ob alle wieder im
// Raum sind (grund "zurueck").
func (r *Room) abgangBeenden(id, name, grund string) (bool, string) {
	if id == "" {
		return false, ""
	}
	var namen []string
	for _, p := range r.partien() {
		if drin, fertig := p.betroffen(id, name); drin && fertig {
			p.ende()
			namen = append(namen, p.name)
		}
	}
	// Ein offenes Angebot, dessen Ersteller weg ist, kann niemand mehr annehmen.
	if r.Angebot != nil && r.Angebot.VonID == id {
		r.Angebot = nil
	}
	if len(namen) == 0 {
		return false, ""
	}

	// Beim Zurück-Knopf ist niemand gegangen – die Meldung nennt deshalb
	// keinen Spieler und keine Trennung. Hinweise entfallen: alle sind im Raum.
	if grund == "zurueck" {
		text := strings.Join(namen, ", ") + " beendet – niemand war mehr am Spielbrett."
		r.addMessage(Message{Text: text, System: true})
		r.LastActive = time.Now()
		r.imSpielVergessen()
		return true, text
	}

	verb := "hat das Spiel verlassen"
	if grund == "verbindung" {
		verb = "hat die Verbindung verloren"
	}
	text := name + " " + verb + " – " + strings.Join(namen, ", ") + " beendet."
	r.addMessage(Message{Text: text, System: true})
	r.LastActive = time.Now()
	r.imSpielVergessen()

	// Wer danach in keiner Partie mehr sitzt, wird sofort herausgeführt.
	// Der Hinweis geht an genau diesen Spieler und wird beim Abholen gelöscht.
	for pid := range r.Players {
		if pid == id {
			continue
		}
		if !r.playerInGame(pid) {
			r.hinweisFuer(pid, text)
		}
	}
	return true, text
}

// allePartienLoeschen räumt einen Raum ohne Spieler vollständig auf.
func (r *Room) allePartienLoeschen() {
	for _, p := range r.partien() {
		p.ende()
	}
}

// enthaelt meldet, ob die Kennung in der Liste steht.
func enthaelt(ids []string, id string) bool {
	for _, x := range ids {
		if x != "" && x == id {
			return true
		}
	}
	return false
}

// idsVon zieht die Kennungen aus einer Spielerliste.
func idsVon[T any](list []T, kennung func(T) string) []string {
	out := make([]string, 0, len(list))
	for _, e := range list {
		out = append(out, kennung(e))
	}
	return out
}

// imSpielHandler beantwortet die Frage "sitze ich noch in einer Partie?".
// Die Spielseiten fragen das im Hintergrund und gehen zurück in den Raum,
// sobald die eigene Partie vorbei ist – niemand wartet auf einen Mitspieler,
// der längst weg ist. Ein hinterlegter Hinweis erklärt dabei, warum.
func imSpielHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	imSpiel := false
	aktiv := ""
	hinweis := ""
	if id != "" {
		mu.Lock()
		if room := rooms[code]; room != nil {
			// Wer gerade auf einer Spielseite sitzt, ist da: das gilt als
			// Lebenszeichen. Ohne das würden Seiten ohne eigenen Takt (Schach,
			// Reversi) nach 150 s aus dem Raum fallen und ihre Partie verlieren.
			// Die Connector-Seite meldet sich mit seite=raum: sie hält den
			// Spieler im Raum, gilt aber nicht als lebende Spielseite.
			// seite=versteckt meldet einen Tab im Hintergrund: Lebenszeichen
			// ja, Spielseite nein – so bleibt eine eingefrorene Partie stehen.
			if p, ok := room.Players[id]; ok {
				seite := strings.TrimSpace(r.URL.Query().Get("seite"))
				p.Last = time.Now()
				if seite == "versteckt" {
					// Nur der erste Ruf beginnt die Phase; weitere versteckte Rufe
					// halten sie am Leben, ohne die Pause-Meldung zu verschieben.
					if p.Versteckt.IsZero() {
						p.VerstecktSeit = p.Last
					}
					p.Versteckt = p.Last
				} else {
					if seite != "raum" {
						p.ImSpiel = p.Last
					}
					// Wieder sichtbar (oder zurück im Raum): die Pause ist vorbei.
					if !p.Versteckt.IsZero() {
						p.Versteckt = time.Time{}
						p.VerstecktSeit = time.Time{}
						p.PauseGemeldet = false
					}
				}
				room.Players[id] = p
			}
			room.cleanPlayers()
			imSpiel = room.playerInGame(id)
			aktiv = room.aktivesSpiel(id)
			hinweis = room.hinweisNehmen(id)
		}
		mu.Unlock()
	}
	jsonResponse(w, map[string]any{"ok": true, "inGame": imSpiel, "aktiv": aktiv, "hinweis": hinweis})
}

// aktivesSpiel liefert die Kennung der Partie, die gerade läuft und in der
// dieser Spieler sitzt. Damit springt die Connector-Seite von selbst ins Spiel,
// sobald eine Einladung angenommen oder ein Tisch gestartet wurde.
func (r *Room) aktivesSpiel(id string) string {
	zwei := func(name, status, a, b string) string {
		if status == "active" && (a == id || b == id) {
			return name
		}
		return ""
	}
	if g := r.TicTacToe; g != nil {
		if s := zwei("tictactoe", g.Status, g.XID, g.OID); s != "" {
			return s
		}
	}
	if g := r.VierGewinnt; g != nil {
		if s := zwei("viergewinnt", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Dame; g != nil {
		if s := zwei("dame", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Paare; g != nil {
		if s := zwei("paare", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Kaese; g != nil {
		if s := zwei("kaesekaestche", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Sim; g != nil {
		if s := zwei("sim", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Kalaha; g != nil {
		if s := zwei("kalaha", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Streichholz; g != nil {
		if s := zwei("streichholzspiel", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Muehle; g != nil {
		if s := zwei("muehle", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Reversi; g != nil {
		if s := zwei("reversi", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Hold; g != nil {
		if s := zwei("hold", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Skyjo; g != nil {
		if s := zwei("skyjo", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Schach; g != nil {
		if s := zwei("schach", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Go; g != nil {
		if s := zwei("go", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.Backgammon; g != nil {
		if s := zwei("backgammon", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	if g := r.MauMau; g != nil && g.Status == "active" {
		for _, p := range g.Players {
			if p.ID == id {
				return "maumau"
			}
		}
	}
	if g := r.Uno; g != nil && g.Status == "active" {
		for _, p := range g.Players {
			if p.ID == id {
				return "uno"
			}
		}
	}
	if g := r.Schwimmen; g != nil && g.Status == "active" {
		for _, p := range g.Players {
			if p.ID == id {
				return "schwimmen"
			}
		}
	}
	if g := r.Kniffel; g != nil && g.Status == "active" {
		for _, p := range g.Players {
			if p.ID == id {
				return "kniffel"
			}
		}
	}
	if g := r.Arschloch; g != nil && g.Status == "active" {
		for _, p := range g.Players {
			if p.ID == id {
				return "arschloch"
			}
		}
	}
	if g := r.Menschaerger; g != nil && g.Status == "active" {
		for _, p := range g.Players {
			if p.ID == id {
				return "menschaerger"
			}
		}
	}
	// Spiele, die in keiner Liste standen: Sie laufen, sobald ein Tisch
	// offen ist – also alles außer Einladung und Ende.
	laeuft := func(status string) bool {
		return status != "" && status != "invited" && status != "ende"
	}
	if g := r.Tetris; g != nil {
		if s := zwei("tetris", g.Status, g.RedID, g.BlueID); s != "" {
			return s
		}
	}
	fuerSitz := func(name string, status string, ids []string) string {
		if laeuft(status) && enthaelt(ids, id) {
			return name
		}
		return ""
	}
	if g := r.Poker; g != nil {
		if s := fuerSitz("poker", g.Status, idsVon(g.Spieler, func(p PokerSpieler) string { return p.ID })); s != "" {
			return s
		}
	}
	if g := r.Cabo; g != nil {
		if s := fuerSitz("cabo", g.Status, idsVon(g.Spieler, func(p CaboSpieler) string { return p.ID })); s != "" {
			return s
		}
	}
	if g := r.Baccara; g != nil {
		if s := fuerSitz("baccara", g.Status, idsVon(g.Spieler, func(p BaccaraSpieler) string { return p.ID })); s != "" {
			return s
		}
	}
	if g := r.Schiffe; g != nil {
		if s := fuerSitz("schiffe", g.Status, idsVon(g.Spieler, func(p SchiffeSpieler) string { return p.ID })); s != "" {
			return s
		}
	}
	if g := r.Siebzehn; g != nil {
		if s := fuerSitz("siebzehn", g.Status, idsVon(g.Spieler, func(p SiebzehnSpieler) string { return p.ID })); s != "" {
			return s
		}
	}
	if g := r.Sechsundsechzig; g != nil {
		if s := fuerSitz("sechsundsechzig", g.Status, idsVon(g.Spieler, func(p SechsundsechzigSpieler) string { return p.ID })); s != "" {
			return s
		}
	}
	if g := r.Galgen; g != nil {
		if s := fuerSitz("galgen", g.Status, galgenIDs(g)); s != "" {
			return s
		}
	}
	if g := r.Bingo; g != nil {
		if s := fuerSitz("bingo", g.Status, bingoIDs(g)); s != "" {
			return s
		}
	}
	if g := r.Wortsuche; g != nil {
		if s := fuerSitz("wortsuche", g.Status, wortsucheIDs(g)); s != "" {
			return s
		}
	}
	if g := r.Domino; g != nil {
		if s := fuerSitz("domino", g.Status, dominoIDs(g)); s != "" {
			return s
		}
	}
	if g := r.Schlangen; g != nil {
		if s := fuerSitz("schlangen", g.Status, schlangenIDs(g)); s != "" {
			return s
		}
	}
	if g := r.Pig; g != nil {
		if s := fuerSitz("pig", g.Status, pigIDs(g)); s != "" {
			return s
		}
	}
	if g := r.Gaense; g != nil {
		if s := fuerSitz("gaensespiel", g.Status, gaenseIDs(g)); s != "" {
			return s
		}
	}
	if g := r.Rechen; g != nil && g.Status == "active" {
		for _, p := range g.Players {
			if p.ID == id {
				return "einsxeins"
			}
		}
	}
	return ""
}

// playerInGame prüft alle Partien des Raums auf diesen Spieler.
func (r *Room) playerInGame(id string) bool {
	if g := r.TicTacToe; g != nil && (g.XID == id || g.OID == id) {
		return true
	}
	if g := r.VierGewinnt; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Dame; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Paare; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Kaese; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Sim; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Kalaha; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Streichholz; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Muehle; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Reversi; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Hold; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Skyjo; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Schach; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Go; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Backgammon; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	if g := r.Tetris; g != nil && (g.RedID == id || g.BlueID == id) {
		return true
	}
	// Kartenspiele mit Sitzliste: Ohne sie hielte sich die Spielseite für
	// „nicht im Spiel“ und schickte den Spieler nach wenigen Sekunden
	// zurück in den Raum, obwohl er noch am Tisch sitzt.
	sitzliste := []func() []string{
		func() []string {
			if g := r.Baccara; g != nil {
				return idsVon(g.Spieler, func(p BaccaraSpieler) string { return p.ID })
			}
			return nil
		},
		func() []string {
			if g := r.Cabo; g != nil {
				return idsVon(g.Spieler, func(p CaboSpieler) string { return p.ID })
			}
			return nil
		},
		func() []string {
			if g := r.Poker; g != nil {
				return idsVon(g.Spieler, func(p PokerSpieler) string { return p.ID })
			}
			return nil
		},
		func() []string {
			if g := r.Schiffe; g != nil {
				return idsVon(g.Spieler, func(p SchiffeSpieler) string { return p.ID })
			}
			return nil
		},
		func() []string {
			if g := r.Siebzehn; g != nil {
				return idsVon(g.Spieler, func(p SiebzehnSpieler) string { return p.ID })
			}
			return nil
		},
		func() []string {
			if g := r.Sechsundsechzig; g != nil {
				return idsVon(g.Spieler, func(p SechsundsechzigSpieler) string { return p.ID })
			}
			return nil
		},
	}
	for _, liste := range sitzliste {
		if enthaelt(liste(), id) {
			return true
		}
	}
	if g := r.MauMau; g != nil {
		for _, p := range g.Players {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Uno; g != nil {
		for _, p := range g.Players {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Schwimmen; g != nil {
		for _, p := range g.Players {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Kniffel; g != nil {
		for _, p := range g.Players {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Arschloch; g != nil {
		for _, p := range g.Players {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Menschaerger; g != nil {
		for _, p := range g.Players {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Galgen; g != nil {
		for _, p := range g.Spieler {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Bingo; g != nil {
		for _, p := range g.Spieler {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Wortsuche; g != nil {
		for _, p := range g.Spieler {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Domino; g != nil {
		for _, p := range g.Spieler {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Schlangen; g != nil {
		for _, p := range g.Spieler {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Pig; g != nil {
		for _, p := range g.Spieler {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Gaense; g != nil {
		for _, p := range g.Spieler {
			if p.ID == id {
				return true
			}
		}
	}
	if g := r.Rechen; g != nil {
		for _, p := range g.Players {
			if p.ID == id {
				return true
			}
		}
	}
	return false
}

// --- Tischspiele: einen Platz entfernen ------------------------------------
// Alle Helfer melden, ob der Spieler überhaupt dabei war, ob die Partie damit
// beendet ist und wie viele Spieler übrig bleiben.

func maumauRemoveSeat(g *MauMauGame, id, name string) (bool, bool, int) {
	idx := -1
	for i, p := range g.Players {
		if p.ID == id {
			idx = i
		}
	}
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
	g.Drawn = nil
	g.DrawnSource = ""
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}

func unoRemoveSeat(g *UnoGame, id, name string) (bool, bool, int) {
	idx := -1
	for i, p := range g.Players {
		if p.ID == id {
			idx = i
		}
	}
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
	g.Drawn = nil
	g.Challenge = nil
	if g.Phase == "chooseColor" || g.Phase == "challenge" {
		g.Phase = "playing"
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}

func schwimmRemoveSeat(g *SchwimmenGame, id, name string) (bool, bool, int) {
	idx := schwimmSeat(g, id)
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
	if g.Phase == "playing" && g.Players[g.Current].Out {
		g.Current = schwimmNext(g, g.Current)
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}

func kniffelRemoveSeat(g *KniffelGame, id, name string) (bool, bool, int) {
	idx := -1
	for i, p := range g.Players {
		if p.ID == id {
			idx = i
		}
	}
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
	g.Dice = nil
	g.Rolls = 0
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}

func arschlochRemoveSeat(g *ArschlochGame, id, name string) (bool, bool, int) {
	idx := arschSeat(g, id)
	if idx < 0 {
		return false, false, len(g.Players)
	}
	rank := g.Players[idx].Rank
	g.Players = append(g.Players[:idx], g.Players[idx+1:]...)
	if len(g.Players) < 2 {
		return true, true, len(g.Players)
	}
	if g.HostID == id {
		g.HostID = g.Players[0].ID
	}
	// Plätze nachrücken lassen, damit die Rangfolge lückenlos bleibt.
	if rank > 0 {
		for _, p := range g.Players {
			if p.Rank > rank {
				p.Rank--
			}
		}
	}
	if g.TrickSeat == idx {
		g.Trick = nil
		g.TrickRank = 0
		g.TrickSeat = -1
	}
	if g.TrickSeat > idx {
		g.TrickSeat--
	}
	if g.Current > idx {
		g.Current--
	}
	if g.Current >= len(g.Players) {
		g.Current = 0
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}

func maenRemoveSeat(g *MenschaergerGame, id, name string) (bool, bool, int) {
	idx := -1
	for i, p := range g.Players {
		if p.ID == id {
			idx = i
		}
	}
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
	g.Dice = 0
	g.ExtraRoll = false
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}
