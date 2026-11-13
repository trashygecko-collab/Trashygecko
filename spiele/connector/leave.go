package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// leaveGameHandler beendet die Partie, in der dieser Spieler sitzt – für alle.
// Wer "Spiel verlassen" drückt, löst die Runde auf: die Mitspieler merken beim
// nächsten Abgleich, dass sie in keiner Partie mehr sitzen, und landen wieder
// im Raum.
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
	name := player.Name

	type ergebnis struct {
		fertig bool
		rest   int
	}
	funde := []ergebnis{}
	merke := func(fertig bool, rest int) { funde = append(funde, ergebnis{fertig, rest}) }

	// --- Duelle: mit einem Spieler ist das Spiel vorbei -------------------
	if g := room.TicTacToe; g != nil && (g.XID == id || g.OID == id) {
		room.TicTacToe = nil
		merke(true, 0)
	}
	if g := room.VierGewinnt; g != nil && (g.RedID == id || g.BlueID == id) {
		room.VierGewinnt = nil
		merke(true, 0)
	}
	if g := room.Dame; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Dame = nil
		merke(true, 0)
	}
	if g := room.Paare; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Paare = nil
		merke(true, 0)
	}
	if g := room.Kaese; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Kaese = nil
		merke(true, 0)
	}
	if g := room.Sim; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Sim = nil
		merke(true, 0)
	}
	if g := room.Kalaha; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Kalaha = nil
		merke(true, 0)
	}
	if g := room.Streichholz; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Streichholz = nil
		merke(true, 0)
	}
	if g := room.Muehle; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Muehle = nil
		merke(true, 0)
	}
	if g := room.Reversi; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Reversi = nil
		merke(true, 0)
	}
	if g := room.Hold; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Hold = nil
		merke(true, 0)
	}
	if g := room.Skyjo; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Skyjo = nil
		merke(true, 0)
	}
	if g := room.Schach; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Schach = nil
		merke(true, 0)
	}
	if g := room.Go; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Go = nil
		merke(true, 0)
	}
	if g := room.Backgammon; g != nil && (g.RedID == id || g.BlueID == id) {
		room.Backgammon = nil
		merke(true, 0)
	}

	// --- Tischspiele: verlassen beendet den ganzen Tisch ----------------
	// Wer das Spiel verlässt, beendet es für alle: die übrigen Spieler
	// landen wieder im Raum, statt mit einem Platz weniger weiterzuspielen.
	if g := room.MauMau; g != nil {
		if imSpiel, _, _ := maumauRemoveSeat(g, id, name); imSpiel {
			room.MauMau = nil
			merke(true, 0)
		}
	}
	// Wer das Spiel verlässt, beendet es für alle: die übrigen Spieler
	// landen wieder im Raum, statt mit einem Platz weniger weiterzuspielen.
	if g := room.Uno; g != nil {
		if imSpiel, _, _ := unoRemoveSeat(g, id, name); imSpiel {
			room.Uno = nil
			merke(true, 0)
		}
	}
	// Wer das Spiel verlässt, beendet es für alle: die übrigen Spieler
	// landen wieder im Raum, statt mit einem Platz weniger weiterzuspielen.
	if g := room.Schwimmen; g != nil {
		if imSpiel, _, _ := schwimmRemoveSeat(g, id, name); imSpiel {
			room.Schwimmen = nil
			merke(true, 0)
		}
	}
	// Wer das Spiel verlässt, beendet es für alle: die übrigen Spieler
	// landen wieder im Raum, statt mit einem Platz weniger weiterzuspielen.
	if g := room.Kniffel; g != nil {
		if imSpiel, _, _ := kniffelRemoveSeat(g, id, name); imSpiel {
			room.Kniffel = nil
			merke(true, 0)
		}
	}
	// Wer das Spiel verlässt, beendet es für alle: die übrigen Spieler
	// landen wieder im Raum, statt mit einem Platz weniger weiterzuspielen.
	if g := room.Arschloch; g != nil {
		if imSpiel, _, _ := arschlochRemoveSeat(g, id, name); imSpiel {
			room.Arschloch = nil
			merke(true, 0)
		}
	}
	// Wer das Spiel verlässt, beendet es für alle: die übrigen Spieler
	// landen wieder im Raum, statt mit einem Platz weniger weiterzuspielen.
	if g := room.Menschaerger; g != nil {
		if imSpiel, _, _ := maenRemoveSeat(g, id, name); imSpiel {
			room.Menschaerger = nil
			merke(true, 0)
		}
	}
	// Wer das Spiel verlässt, beendet es für alle: die übrigen Spieler
	// landen wieder im Raum, statt mit einem Platz weniger weiterzuspielen.
	if g := room.Rechen; g != nil {
		if imSpiel, _, _ := rechenRemoveSeat(g, id, name); imSpiel {
			room.Rechen = nil
			merke(true, 0)
		}
	}

	for _, f := range funde {
		text := name + " hat das Spiel verlassen."
		switch {
		case f.fertig && f.rest <= 1:
			text += " Das Spiel ist beendet – zurück in den Raum."
		case f.fertig:
			text += " Das Spiel ist beendet."
		default:
			text += " Es geht mit " + strconv.Itoa(f.rest) + " Spielern weiter."
		}
		room.addMessage(Message{Text: text, System: true})
	}

	ende := len(funde) > 0
	if ende {
		room.LastActive = time.Now()
	}
	mu.Unlock()
	jsonResponse(w, map[string]any{"ok": true, "left": len(funde), "ended": ende})
}

// imSpielHandler beantwortet die Frage "sitze ich noch in einer Partie?".
// Die Spielseiten fragen das im Hintergrund und gehen zurück in den Raum,
// sobald die eigene Partie vorbei ist – niemand wartet auf einen Mitspieler,
// der längst weg ist.
func imSpielHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	imSpiel := false
	aktiv := ""
	if id != "" {
		mu.Lock()
		if room := rooms[code]; room != nil {
			// Wer gerade auf einer Spielseite sitzt, ist da: das gilt als
			// Lebenszeichen. Ohne das würden Seiten ohne eigenen Takt (Schach,
			// Reversi) nach 150 s aus dem Raum fallen und ihre Partie verlieren.
			if p, ok := room.Players[id]; ok {
				p.Last = time.Now()
				room.Players[id] = p
			}
			room.cleanPlayers()
			imSpiel = room.playerInGame(id)
			aktiv = room.aktivesSpiel(id)
		}
		mu.Unlock()
	}
	jsonResponse(w, map[string]any{"ok": true, "inGame": imSpiel, "aktiv": aktiv})
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
