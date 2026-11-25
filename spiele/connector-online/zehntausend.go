package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 10.000: sechs Würfel, 2 bis 6 Spieler, Ziel 10.000 Punkte, Speichern erst
// ab 350 Punkten im Zug. Dieselben Regeln stehen in der Spielseite
// (spiele/brettspiele/zehntausend/zehntausend.html) – hier zählt der Server.
const ztMaxPlayers = 6
const ztZiel = 10000
const ztEinstieg = 350

var ztColors = []string{"#e8cd8e", "#7fb98a", "#d98878", "#8aa9d6", "#b18cff", "#ff985e"}

type ZTPlayer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Punkte int    `json:"punkte"`
	Ready  bool   `json:"ready"`
}

// ZTFehlwurf hält den letzten Fehlwurf fest: die gewürfelten Würfel bleiben
// im Stand, bis der Nächste handelt. Die Seite zeigt sie in ihrer Meldung
// („Leider Pech gehabt!“) und wechselt erst danach sichtbar den Spieler.
type ZTFehlwurf struct {
	Nr        int    `json:"nr"`
	Name      string `json:"name"`
	Wuerfel   []int  `json:"wuerfel"`
	Verloren  int    `json:"verloren"`
	Naechster string `json:"naechster"`
}

type ZehnTausendGame struct {
	ID            string      `json:"id"`
	HostID        string      `json:"hostId"`
	BotID         string      `json:"botId,omitempty"`
	BotStufe      string      `json:"-"`
	BotFrist      time.Time   `json:"-"`
	Players       []ZTPlayer  `json:"players"`
	Current       int         `json:"current"`
	Wuerfel       []int       `json:"wuerfel"`
	Gesichert     []bool      `json:"gesichert"`
	WurfGesichert []bool      `json:"wurfGesichert"`
	WurfPunkte    int         `json:"wurfPunkte"`
	ZugPunkte     int         `json:"zugPunkte"`
	Wuerfe        int         `json:"wuerfe"`
	Status        string      `json:"status"` // invited, active, finished
	Winner        string      `json:"winner,omitempty"`
	Note          string      `json:"note,omitempty"`
	Fehlwurf      *ZTFehlwurf `json:"fehlwurf,omitempty"`

	// zaehlt die Fehlwürfe; die Seite erkennt daran eine neue Meldung.
	fehlwurfZaehler int
}

// ztDreier ist der Wert von drei gleichen: drei 1er = 1.000, drei 6er = 600.
func ztDreier(wert int) int {
	if wert == 1 {
		return 1000
	}
	return wert * 100
}

// ztZaehlt sagt, ob diese Augenzahl in dieser Anzahl Punkte bringt.
func ztZaehlt(wert, anzahl int) bool {
	return wert == 1 || wert == 5 || anzahl >= 3
}

// ztPunkte rechnet eine Würfelmenge aus. Jeder gleiche Würfel über drei
// hinaus zählt den Satz noch einmal (vier 2er = 400).
func ztPunkte(werte []int) int {
	zahl := make([]int, 7)
	for _, w := range werte {
		if w >= 1 && w <= 6 {
			zahl[w]++
		}
	}
	summe := 0
	for w := 1; w <= 6; w++ {
		n := zahl[w]
		if n == 0 {
			continue
		}
		switch {
		case n >= 3:
			summe += ztDreier(w) * (n - 2)
		case w == 1:
			summe += n * 100
		case w == 5:
			summe += n * 50
		}
	}
	return summe
}

// ztHatPunkte: Gibt es in diesem Wurf überhaupt einen Würfel, der zählt?
func ztHatPunkte(werte []int) bool {
	zahl := make([]int, 7)
	for _, w := range werte {
		if w >= 1 && w <= 6 {
			zahl[w]++
		}
	}
	if zahl[1] > 0 || zahl[5] > 0 {
		return true
	}
	for w := 2; w <= 6; w++ {
		if zahl[w] >= 3 {
			return true
		}
	}
	return false
}

func ztNewPlayer(id, name string, seat int) ZTPlayer {
	return ZTPlayer{ID: id, Name: name, Color: ztColors[seat%len(ztColors)]}
}

func ztFindPlayer(g *ZehnTausendGame, id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func ztIDs(g *ZehnTausendGame) []string {
	ids := make([]string, 0, len(g.Players))
	for _, p := range g.Players {
		ids = append(ids, p.ID)
	}
	return ids
}

func ztNeuerZug(g *ZehnTausendGame) {
	g.Wuerfel = make([]int, 6)
	g.Gesichert = make([]bool, 6)
	g.WurfGesichert = make([]bool, 6)
	g.WurfPunkte = 0
	g.ZugPunkte = 0
	g.Wuerfe = 0
}

// ztNaechsterWurf räumt nur die Würfel ab (alle sechs kommen zurück);
// die im Zug gesammelten Punkte bleiben erhalten.
func ztNaechsterWurf(g *ZehnTausendGame) {
	g.Wuerfel = make([]int, 6)
	g.Gesichert = make([]bool, 6)
	g.WurfGesichert = make([]bool, 6)
	g.WurfPunkte = 0
}

func ztNaechster(g *ZehnTausendGame) {
	g.Current = (g.Current + 1) % len(g.Players)
	ztNeuerZug(g)
}

// ztSichere legt die Auswahl fest: nur vollständige Sätze und einzelne 1er/5er
// zählen, unvollständige Auswahlen bleiben liegen. Gerechnet wird die ganze
// Gruppe des laufenden Wurfs neu; gutgeschrieben wird nur der Unterschied,
// damit auch der vierte gleiche Würfel seinen vollen Satz bringt.
func ztSichere(g *ZehnTausendGame, auswahl []bool) (int, bool) {
	if len(g.WurfGesichert) != 6 {
		g.WurfGesichert = make([]bool, 6)
	}
	// Gezählt wird nur, was ausgewählt ist: eine einzelne 2 zählt auch dann
	// nichts, wenn drei 2er auf dem Tisch liegen.
	gewaehlt := make([]int, 7)
	for i := 0; i < 6; i++ {
		if i >= len(auswahl) || !auswahl[i] || g.Gesichert[i] {
			continue
		}
		if w := g.Wuerfel[i]; w >= 1 && w <= 6 {
			gewaehlt[w]++
		}
	}
	for i := 0; i < 6; i++ {
		if i >= len(auswahl) || !auswahl[i] || g.Gesichert[i] {
			continue
		}
		w := g.Wuerfel[i]
		if w == 0 || !ztZaehlt(w, gewaehlt[w]) {
			continue
		}
		g.Gesichert[i] = true
		g.WurfGesichert[i] = true
	}
	werte := []int{}
	for i := 0; i < 6; i++ {
		if g.Gesichert[i] && g.WurfGesichert[i] && g.Wuerfel[i] >= 1 && g.Wuerfel[i] <= 6 {
			werte = append(werte, g.Wuerfel[i])
		}
	}
	punkte := ztPunkte(werte)
	dazu := punkte - g.WurfPunkte
	g.ZugPunkte += dazu
	g.WurfPunkte = punkte
	alle := true
	for _, s := range g.Gesichert {
		if !s {
			alle = false
			break
		}
	}
	return dazu, alle
}

// ztWuerfle würfelt alle nicht gesicherten Würfel neu.
func ztWuerfle(g *ZehnTausendGame) []int {
	// Ein neuer Wurf beginnt eine eigene Punktgruppe.
	g.WurfGesichert = make([]bool, 6)
	g.WurfPunkte = 0
	geworfen := []int{}
	for i := 0; i < 6; i++ {
		if g.Gesichert[i] {
			continue
		}
		g.Wuerfel[i] = rand.Intn(6) + 1
		geworfen = append(geworfen, g.Wuerfel[i])
	}
	g.Wuerfe++
	return geworfen
}

// ztFehlwurf prüft einen Wurf und beendet den Zug, wenn nichts zählt.
func ztFehlwurf(g *ZehnTausendGame, geworfen []int, name string) bool {
	if ztHatPunkte(geworfen) {
		return false
	}
	verloren := g.ZugPunkte
	ztNaechster(g)
	g.fehlwurfZaehler++
	g.Fehlwurf = &ZTFehlwurf{
		Nr:        g.fehlwurfZaehler,
		Name:      name,
		Wuerfel:   append([]int(nil), geworfen...),
		Verloren:  verloren,
		Naechster: g.Players[g.Current].Name,
	}
	g.Note = name + " hat einen Fehlwurf (" + ztListe(geworfen) + ")"
	if verloren > 0 {
		g.Note += " – " + strconv.Itoa(verloren) + " Punkte sind weg."
	} else {
		g.Note += " – der Zug wechselt."
	}
	g.Note += " Jetzt ist " + g.Players[g.Current].Name + " dran."
	return true
}

func ztListe(werte []int) string {
	teile := make([]string, 0, len(werte))
	for _, w := range werte {
		teile = append(teile, strconv.Itoa(w))
	}
	return strings.Join(teile, " · ")
}

func cloneZehnTausend(g *ZehnTausendGame) *ZehnTausendGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Players = append([]ZTPlayer(nil), g.Players...)
	c.Wuerfel = append([]int(nil), g.Wuerfel...)
	c.Gesichert = append([]bool(nil), g.Gesichert...)
	c.WurfGesichert = append([]bool(nil), g.WurfGesichert...)
	if g.Fehlwurf != nil {
		f := *g.Fehlwurf
		f.Wuerfel = append([]int(nil), g.Fehlwurf.Wuerfel...)
		c.Fehlwurf = &f
	}
	return &c
}

func zehnTausendRemoveSeat(g *ZehnTausendGame, id, name string) (bool, bool, int) {
	idx := ztFindPlayer(g, id)
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
	ztNeuerZug(g)
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Players)) + " Spielern weiter."
	return true, false, len(g.Players)
}

// ztBank schreibt die Zugpunkte gut und wechselt den Spieler. false heißt:
// Die 350 Punkte fürs Speichern sind noch nicht erreicht. Bei 10.000 endet
// die Partie; der Aufrufer setzt keine weitere Meldung mehr.
func ztBank(g *ZehnTausendGame, seat int) bool {
	if seat < 0 || seat >= len(g.Players) || g.ZugPunkte < ztEinstieg {
		return false
	}
	summe := g.ZugPunkte
	g.Players[seat].Punkte += summe
	g.Note = g.Players[seat].Name + " sichert " + strconv.Itoa(summe) + " Punkte. Konto: " +
		strconv.Itoa(g.Players[seat].Punkte) + " von " + strconv.Itoa(ztZiel) + "."
	if g.Players[seat].Punkte >= ztZiel {
		g.Status = "finished"
		g.Winner = g.Players[seat].Name
		g.Note = g.Players[seat].Name + " gewinnt mit " + strconv.Itoa(g.Players[seat].Punkte) + " Punkten!"
		return true
	}
	ztNaechster(g)
	g.Note += " Jetzt ist " + g.Players[g.Current].Name + " dran."
	return true
}

func zehnTausendHandler(w http.ResponseWriter, r *http.Request) {
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
		if room.ZehnTausend != nil {
			ztBotTreiben(room.ZehnTausend, time.Now())
		}
		game := cloneZehnTausend(room.ZehnTausend)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID      string `json:"id"`
		Stufe   string `json:"stufe"`
		Auswahl []bool `json:"auswahl"`
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

	// Jede Aktion räumt die Fehlwurf-Meldung ab; ein neuer Fehlwurf setzt sie
	// gleich wieder. So bleibt sie genau bis zum nächsten Zug stehen.
	if room.ZehnTausend != nil {
		room.ZehnTausend.Fehlwurf = nil
	}

	switch action {
	case "bot":
		room.botNamenNeu()
		if room.ZehnTausend != nil && room.ZehnTausend.Status != "finished" {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine 10.000-Runde.", http.StatusConflict)
			return
		}
		stufe := botStufe(input.Stufe)
		g := &ZehnTausendGame{
			ID:       randomCode(),
			HostID:   input.ID,
			Status:   "active",
			BotID:    botID(1),
			BotStufe: stufe,
		}
		ztNeuerZug(g)
		g.Players = append(g.Players, ztNewPlayer(input.ID, player.Name, 0))
		g.Players = append(g.Players, ztNewPlayer(g.BotID, room.botName(1), 1))
		g.Players[0].Ready = true
		g.Players[1].Ready = true
		g.Note = player.Name + " spielt gegen " + room.botName(1) + " (" + botStufeText(stufe) + ")."
		room.ZehnTausend = g
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt 10.000 gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     uhrzeit(time.Now()),
			SenderID: input.ID,
		})

	case "create":
		if room.ZehnTausend != nil && room.ZehnTausend.Status != "finished" {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine 10.000-Runde.", http.StatusConflict)
			return
		}
		game := &ZehnTausendGame{
			ID:     randomCode(),
			HostID: input.ID,
			Status: "invited",
		}
		ztNeuerZug(game)
		game.Players = append(game.Players, ztNewPlayer(input.ID, player.Name, 0))
		game.Note = player.Name + " eröffnet eine 10.000-Runde."
		room.ZehnTausend = game
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine 10.000-Runde – bis zu 6 Spieler, mitspielen über die Spielkarte.",
			Time:     uhrzeit(time.Now()),
			SenderID: input.ID,
		})

	case "join":
		game := room.ZehnTausend
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine offene 10.000-Runde.", http.StatusConflict)
			return
		}
		if ztFindPlayer(game, input.ID) >= 0 {
			break
		}
		if len(game.Players) >= ztMaxPlayers {
			mu.Unlock()
			http.Error(w, "Diese 10.000-Runde ist mit 6 Spielern voll.", http.StatusConflict)
			return
		}
		game.Players = append(game.Players, ztNewPlayer(input.ID, player.Name, len(game.Players)))
		game.Note = player.Name + " spielt mit."
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "setzt sich zur 10.000-Runde.",
			Time:     uhrzeit(time.Now()),
			SenderID: input.ID,
		})

	case "ready":
		game := room.ZehnTausend
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur, solange die Runde noch nicht läuft.", http.StatusConflict)
			return
		}
		seat := ztFindPlayer(game, input.ID)
		if seat < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		game.Players[seat].Ready = !game.Players[seat].Ready
		if game.Players[seat].Ready {
			game.Note = player.Name + " ist bereit."
		} else {
			game.Note = player.Name + " wartet noch."
		}

	case "start":
		game := room.ZehnTausend
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine 10.000-Runde auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber kann die Runde starten.", http.StatusForbidden)
			return
		}
		if len(game.Players) < 2 {
			mu.Unlock()
			http.Error(w, "10.000 braucht mindestens zwei Spieler am Tisch.", http.StatusConflict)
			return
		}
		for _, p := range game.Players {
			if !p.Ready {
				mu.Unlock()
				http.Error(w, "Erst müssen alle Spieler BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.Status = "active"
		game.Current = 0
		ztNeuerZug(game)
		game.Note = "Das Spiel beginnt! " + game.Players[0].Name + " würfelt zuerst."

	case "roll":
		game := room.ZehnTausend
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		seat := ztFindPlayer(game, input.ID)
		if seat < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Current != seat {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		if game.Wuerfe > 0 {
			mu.Unlock()
			http.Error(w, "Du hast schon gewürfelt – nimm WEITERWÜRFELN.", http.StatusConflict)
			return
		}
		geworfen := ztWuerfle(game)
		if ztFehlwurf(game, geworfen, player.Name) {
			break
		}
		game.Note = player.Name + " würfelt: " + ztListe(geworfen) + ". " +
			strconv.Itoa(ztPunkte(geworfen)) + " Punkte sind möglich."

	case "weiter":
		game := room.ZehnTausend
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		seat := ztFindPlayer(game, input.ID)
		if seat < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Current != seat {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		if game.Wuerfe == 0 {
			mu.Unlock()
			http.Error(w, "Du musst zuerst würfeln.", http.StatusConflict)
			return
		}
		sicher, alle := ztSichere(game, input.Auswahl)
		if alle {
			// Alle sechs Würfel sind verbraucht und kommen zurück; die Zugpunkte bleiben.
			punkte := game.ZugPunkte
			ztNaechsterWurf(game)
			game.Note = player.Name + " hat alle sechs Würfel verbraucht – sie kommen alle zurück. Zug: " +
				strconv.Itoa(punkte) + " Punkte."
			break
		}
		geworfen := ztWuerfle(game)
		if ztFehlwurf(game, geworfen, player.Name) {
			break
		}
		game.Note = player.Name + " würfelt weiter: " + ztListe(geworfen) + "."
		if sicher > 0 {
			game.Note = strconv.Itoa(sicher) + " Punkte gesichert. " + game.Note
		}
		game.Note += " Zug: " + strconv.Itoa(game.ZugPunkte) + " Punkte."

	case "merken":
		// Einen Würfel sofort festlegen (1, 5 oder vollständiger Pasch).
		game := room.ZehnTausend
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		seat := ztFindPlayer(game, input.ID)
		if seat < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Current != seat {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		if game.Wuerfe == 0 {
			mu.Unlock()
			http.Error(w, "Du musst zuerst würfeln.", http.StatusConflict)
			return
		}
		punkte, alle := ztSichere(game, input.Auswahl)
		game.Note = player.Name + " wählt Würfel aus."
		if punkte > 0 {
			game.Note = player.Name + " legt " + strconv.Itoa(punkte) + " Punkte beiseite. Zug: " +
				strconv.Itoa(game.ZugPunkte) + " Punkte."
		}
		if alle {
			game.Note += " Alle sechs Würfel liegen – sichern oder weiterwürfeln."
		}

	case "sichern":
		game := room.ZehnTausend
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		seat := ztFindPlayer(game, input.ID)
		if seat < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.Current != seat {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		if game.Wuerfe == 0 {
			mu.Unlock()
			http.Error(w, "Du musst zuerst würfeln.", http.StatusConflict)
			return
		}
		ztSichere(game, input.Auswahl)
		if !ztBank(game, seat) {
			mu.Unlock()
			http.Error(w, "Erst ab "+strconv.Itoa(ztEinstieg)+" Punkten speichern – es fehlen noch "+
				strconv.Itoa(ztEinstieg-game.ZugPunkte)+".", http.StatusConflict)
			return
		}

	case "reset":
		game := room.ZehnTausend
		if game == nil || game.Status != "finished" {
			mu.Unlock()
			http.Error(w, "Ein neues Spiel ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		if ztFindPlayer(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		fresh := &ZehnTausendGame{
			ID:       game.ID,
			HostID:   game.HostID,
			Status:   "invited",
			BotID:    game.BotID,
			BotStufe: game.BotStufe,
		}
		ztNeuerZug(fresh)
		for i, p := range game.Players {
			fresh.Players = append(fresh.Players, ztNewPlayer(p.ID, p.Name, i))
		}
		for i := range fresh.Players {
			if fresh.Players[i].ID == fresh.BotID {
				fresh.Players[i].Ready = true
			}
		}
		fresh.Note = "Neue Runde – alle drücken BEREIT, dann startet der Gastgeber."
		room.ZehnTausend = fresh

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte 10.000-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneZehnTausend(room.ZehnTausend)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
