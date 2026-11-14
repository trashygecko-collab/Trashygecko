package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// maxNameLen ist die maximale Länge eines Spielernamens in Zeichen.
const maxNameLen = 24

// presenceTimeout bestimmt, wie lange ein Spieler ohne Lebenszeichen im Raum
// bleibt. Handys sperren den Bildschirm, wechseln kurz die App oder verlieren
// im WLAN ein paar Pakete – mit zu kurzer Frist flogen Spieler dabei aus dem
// Raum und ihre Partie wurde weggeworfen.
const presenceTimeout = 150 * time.Second

// emptyRoomGameGrace: Sind alle Geräte weg, bleiben Einladungen und Partien
// noch so lange liegen und gehen erst danach verloren.
const emptyRoomGameGrace = 5 * time.Minute

// afkAfter: Wer so lange nichts tut (kein Chat, kein Klick, kein Zug), gilt
// als kurz weg – das automatische Lebenszeichen zählt dabei nicht mit.
const afkAfter = time.Minute

type Message struct {
	Name       string `json:"name"`
	Text       string `json:"text"`
	Time       string `json:"time"`
	SenderID   string `json:"senderId"`
	ToID       string `json:"toId,omitempty"`
	ToName     string `json:"toName,omitempty"`
	Whisper    bool   `json:"whisper"`
	GameInvite bool   `json:"gameInvite,omitempty"`
	// Tone färbt eine Nachricht ein: love (rosa), joke (grün), evil (rot).
	Tone   string `json:"tone,omitempty"`
	Game   string `json:"game,omitempty"`
	GameID string `json:"gameId,omitempty"`
	System bool   `json:"system,omitempty"`
	// Gone markiert "verlässt den Raum" – der Name ist dann nicht anklickbar.
	Gone bool `json:"gone,omitempty"`
}

type Player struct {
	Name string    `json:"name"`
	Last time.Time `json:"-"`
	// LastAction ist die letzte echte Aktion des Spielers, AfkWarned merkt
	// sich, ob für ihn schon eine „ist wohl eingeschlafen“-Meldung kam.
	LastAction time.Time `json:"-"`
	AfkWarned  bool      `json:"-"`
}

type TicTacToeGame struct {
	ID      string   `json:"id"`
	XID     string   `json:"xId"`
	XName   string   `json:"xName"`
	OID     string   `json:"oId"`
	OName   string   `json:"oName"`
	Board   []string `json:"board"`
	Current string   `json:"current"`
	Status  string   `json:"status"`
	Winner  string   `json:"winner,omitempty"`
}

type Room struct {
	Code      string
	OwnerID   string
	OwnerName string
	// EmptySince merkt sich, seit wann niemand mehr im Raum ist. Solange das
	// nicht lange her ist, bleiben laufende Partien liegen (Handy gesperrt,
	// Rechner zu, kurze Pause).
	EmptySince   time.Time
	Name         string // e.g. "Spielername Raum"
	Created      time.Time
	LastActive   time.Time
	Players      map[string]Player
	Messages     []Message
	TicTacToe    *TicTacToeGame
	VierGewinnt  *VierGewinntGame
	Dame         *DameGame
	Paare        *PaareGame
	Menschaerger *MenschaergerGame
	Kaese        *KaeseGame
	Sim          *SimGame
	Kalaha       *KalahaGame
	Streichholz  *StreichholzGame
	Muehle       *MuehleGame
	Reversi      *ReversiGame
	Hold         *HoldGame
	Skyjo        *SkyjoGame
	Kniffel      *KniffelGame
	Schach       *ChessGame
	Go           *GoGame
	Backgammon   *BackgammonGame
	MauMau       *MauMauGame
	Uno          *UnoGame
	Schwimmen    *SchwimmenGame
	Arschloch    *ArschlochGame
	Rechen       *RechenGame
	Mal          *Malbrett              // gemeinsame Zeichenfläche
	Angebot      *Angebot               // offene Runde für den ganzen Raum
	Mutes        map[string]time.Time   // playerID -> unmute time
	MsgTimes     map[string][]time.Time // for spam protection
}

type RoomInfo struct {
	Code      string    `json:"code"`
	Players   int       `json:"players"`
	Created   string    `json:"created"`
	CreatedAt time.Time `json:"-"`
	Owner     string    `json:"owner,omitempty"`
	Status    string    `json:"status,omitempty"`
	Game      string    `json:"game,omitempty"`
	Name      string    `json:"name,omitempty"`
}

var (
	rooms = map[string]*Room{}
	mu    sync.Mutex
)

// limitName kürzt einen Namen auf maxNameLen Zeichen. Ein Schnitt nach Bytes
// würde Umlaute zerteilen und ungültiges UTF-8 erzeugen.
func limitName(s string) string {
	if utf8.RuneCountInString(s) <= maxNameLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxNameLen])
}

func (r *Room) cleanPlayers() {
	now := time.Now()
	for id, p := range r.Players {
		if now.Sub(p.Last) > presenceTimeout {
			delete(r.Players, id)
			// Mit Absender, damit der Name im Chat steht und angeklickt werden kann.
			r.addMessage(Message{
				Name:     p.Name,
				Text:     "verlässt den Raum.",
				Time:     now.Format("15:04:05"),
				SenderID: id,
				Gone:     true,
			})
		}
	}
	// Ein Angebot, dessen Ersteller weg ist, kann niemand mehr annehmen.
	if r.Angebot != nil {
		if _, da := r.Players[r.Angebot.VonID]; !da {
			r.Angebot = nil
		}
	}
	if len(r.Players) == 0 {
		if r.EmptySince.IsZero() {
			r.EmptySince = now
		}
	} else {
		r.EmptySince = time.Time{}
	}
	// Sind alle Geräte kurz weg (gesperrte Handys, geschlossene Tabs), bleibt
	// die Partie noch eine Weile stehen, damit sie beim Zurückkommen weitergeht.
	if len(r.Players) == 0 && now.Sub(r.EmptySince) <= emptyRoomGameGrace {
		return
	}
	// Eine Einladung an jemanden, der den Raum verlassen hat, ist wertlos und
	// würde jede weitere Einladung blockieren.
	if r.TicTacToe != nil && r.gameAbandoned(r.TicTacToe.Status, r.TicTacToe.XID, r.TicTacToe.OID) {
		r.TicTacToe = nil
	}
	if r.VierGewinnt != nil && r.gameAbandoned(r.VierGewinnt.Status, r.VierGewinnt.RedID, r.VierGewinnt.BlueID) {
		r.VierGewinnt = nil
	}
	if r.Dame != nil && r.gameAbandoned(r.Dame.Status, r.Dame.RedID, r.Dame.BlueID) {
		r.Dame = nil
	}
	if r.Paare != nil && r.gameAbandoned(r.Paare.Status, r.Paare.RedID, r.Paare.BlueID) {
		r.Paare = nil
	}
	// Ein Tisch mit bis zu vier Sitzen hat eigene Regeln: er lebt, solange
	// der Gastgeber da ist.
	if r.Menschaerger != nil && r.Menschaerger.abandoned(r) {
		r.Menschaerger = nil
	}
	if r.Kaese != nil && r.gameAbandoned(r.Kaese.Status, r.Kaese.RedID, r.Kaese.BlueID) {
		r.Kaese = nil
	}
	if r.Sim != nil && r.gameAbandoned(r.Sim.Status, r.Sim.RedID, r.Sim.BlueID) {
		r.Sim = nil
	}
	if r.Kalaha != nil && r.gameAbandoned(r.Kalaha.Status, r.Kalaha.RedID, r.Kalaha.BlueID) {
		r.Kalaha = nil
	}
	if r.Streichholz != nil && r.gameAbandoned(r.Streichholz.Status, r.Streichholz.RedID, r.Streichholz.BlueID) {
		r.Streichholz = nil
	}
	if r.Muehle != nil && r.gameAbandoned(r.Muehle.Status, r.Muehle.RedID, r.Muehle.BlueID) {
		r.Muehle = nil
	}
	if r.Reversi != nil && r.gameAbandoned(r.Reversi.Status, r.Reversi.RedID, r.Reversi.BlueID) {
		r.Reversi = nil
	}
	if r.Hold != nil && r.gameAbandoned(r.Hold.Status, r.Hold.RedID, r.Hold.BlueID) {
		r.Hold = nil
	}
	if r.Skyjo != nil && r.gameAbandoned(r.Skyjo.Status, r.Skyjo.RedID, r.Skyjo.BlueID) {
		r.Skyjo = nil
	}
	if r.Schach != nil && r.gameAbandoned(r.Schach.Status, r.Schach.RedID, r.Schach.BlueID) {
		r.Schach = nil
	}
	if r.Go != nil && r.gameAbandoned(r.Go.Status, r.Go.RedID, r.Go.BlueID) {
		r.Go = nil
	}
	if r.Backgammon != nil && r.gameAbandoned(r.Backgammon.Status, r.Backgammon.RedID, r.Backgammon.BlueID) {
		r.Backgammon = nil
	}
	if r.Kniffel != nil && r.Kniffel.Status != "finished" {
		present := 0
		for _, p := range r.Kniffel.Players {
			if _, ok := r.Players[p.ID]; ok {
				present++
			}
		}
		if present == 0 {
			r.Kniffel = nil
		}
	}
	if r.MauMau != nil && r.MauMau.allGone(r) {
		r.MauMau = nil
	}
	if r.Uno != nil && r.Uno.allGone(r) {
		r.Uno = nil
	}
	if r.Schwimmen != nil && r.Schwimmen.allGone(r) {
		r.Schwimmen = nil
	}
	if r.Arschloch != nil && r.Arschloch.allGone(r) {
		r.Arschloch = nil
	}
	if r.Rechen != nil && r.Rechen.allGone(r) {
		r.Rechen = nil
	}
	// clean old timestamps for spam
	for id, ts := range r.MsgTimes {
		filtered := []time.Time{}
		for _, t := range ts {
			if now.Sub(t) < 10*time.Second {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) == 0 {
			delete(r.MsgTimes, id)
		} else {
			r.MsgTimes[id] = filtered
		}
	}
}

// uniqueName macht Spielernamen im Raum eindeutig: Wer mit einem schon
// belegten Namen kommt, wird „Anna1“, „Anna2“ usw. Sonst könnte sich jeder als
// Admin ausgeben und dessen Platz übernehmen.
func (r *Room) uniqueName(name, id string) string {
	taken := func(candidate string) bool {
		for otherID, p := range r.Players {
			if otherID != id && strings.EqualFold(p.Name, candidate) {
				return true
			}
		}
		return false
	}
	if !taken(name) {
		return name
	}
	base := limitName(name)
	for i := 1; ; i++ {
		candidate := limitName(base + strconv.Itoa(i))
		if !taken(candidate) {
			return candidate
		}
	}
}

// freierName findet den ersten freien Namen der Form "Spieler1", "Spieler2", …
func (r *Room) freierName(basis string) string {
	belegt := func(kandidat string) bool {
		for _, p := range r.Players {
			if strings.EqualFold(p.Name, kandidat) {
				return true
			}
		}
		return false
	}
	for i := 1; i < 100; i++ {
		if kandidat := basis + strconv.Itoa(i); !belegt(kandidat) {
			return kandidat
		}
	}
	return basis
}

// warnIdle meldet Spieler, die seit über einer Minute nichts getan haben. Der
// Admin bleibt außen vor: er sitzt am Server und schaut oft nur zu.
func (r *Room) warnIdle(now time.Time) {
	for id, p := range r.Players {
		if id == r.OwnerID || p.AfkWarned {
			continue
		}
		if now.Sub(p.LastAction) <= afkAfter {
			continue
		}
		p.AfkWarned = true
		r.Players[id] = p
		r.addMessage(Message{
			Text:   p.Name + " ist wohl gerade auf dem Klo oder eingeschlafen.",
			Time:   now.Format("15:04:05"),
			System: true,
		})
	}
}

// noteActivity hält fest, dass ein Spieler etwas getan hat (Chat, Spielzug,
// Klick auf der Oberfläche).
func noteActivity(code, id string) {
	if code == "" || id == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	room := rooms[strings.ToUpper(strings.TrimSpace(code))]
	if room == nil {
		return
	}
	player, ok := room.Players[id]
	if !ok {
		return
	}
	player.LastAction = time.Now()
	player.AfkWarned = false
	room.Players[id] = player
}

// gameSlotFree meldet, ob ein neues Spiel das aktuelle ersetzen darf. Eine
// abgeschlossene Partie blockiert nie, eine aufgegebene auch nicht.
func (r *Room) gameSlotFree(status string, ids ...string) bool {
	if status != "invited" && status != "active" {
		return true
	}
	return r.gameAbandoned(status, ids...)
}

// gameAbandoned meldet, ob niemand mehr da ist, der die Partie spielen kann:
// eine Einladung an einen abwesenden Spieler oder eine laufende Partie, aus
// der beide Spieler verschwunden sind. Eine laufende Partie mit einem
// anwesenden Spieler gilt weiter als belegt.
func (r *Room) gameAbandoned(status string, ids ...string) bool {
	present := 0
	for _, id := range ids {
		if _, ok := r.Players[id]; ok {
			present++
		}
	}
	if status == "invited" {
		return present < len(ids)
	}
	return present == 0
}

func (r *Room) isMuted(id string) (bool, time.Duration) {
	until, ok := r.Mutes[id]
	if !ok {
		return false, 0
	}
	remaining := time.Until(until)
	if remaining <= 0 {
		delete(r.Mutes, id)
		return false, 0
	}
	return true, remaining
}

func (r *Room) addMessage(msg Message) {
	r.Messages = append(r.Messages, msg)
	if len(r.Messages) > 100 {
		r.Messages = r.Messages[len(r.Messages)-100:]
	}
}

func (r *Room) gameStatus() string {
	if r.TicTacToe != nil {
		switch r.TicTacToe.Status {
		case "active":
			return "Spiel läuft"
		case "invited":
			return "Einladung ausstehend"
		}
	}
	if r.VierGewinnt != nil {
		switch r.VierGewinnt.Status {
		case "active":
			return "Spiel läuft"
		case "invited":
			return "Einladung ausstehend"
		}
	}
	if r.Dame != nil {
		switch r.Dame.Status {
		case "active":
			return "Spiel läuft"
		case "invited":
			return "Einladung ausstehend"
		}
	}
	if r.Paare != nil {
		switch r.Paare.Status {
		case "active":
			return "Spiel läuft"
		case "invited":
			return "Einladung ausstehend"
		}
	}
	if r.Menschaerger != nil {
		switch r.Menschaerger.Status {
		case "active":
			return "Spiel läuft"
		case "invited":
			return "Einladung ausstehend"
		}
	}
	return "Verfügbar"
}

// cloneTicTacToe kopiert einen Spielstand. Ohne Kopie würde der JSON-Encoder
// das Brett lesen, während ein anderer Request es beschreibt (Datenrennen).
func cloneTicTacToe(g *TicTacToeGame) *TicTacToeGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Board = append([]string(nil), g.Board...)
	return &c
}

func cloneVierGewinnt(g *VierGewinntGame) *VierGewinntGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Grid = make([][]string, len(g.Grid))
	for i, row := range g.Grid {
		c.Grid[i] = append([]string(nil), row...)
	}
	if g.Winning != nil {
		c.Winning = make([][]int, len(g.Winning))
		for i, cell := range g.Winning {
			c.Winning[i] = append([]int(nil), cell...)
		}
	}
	return &c
}

func cloneMenschaerger(g *MenschaergerGame) *MenschaergerGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Players = make([]MaenPlayer, len(g.Players))
	for i, player := range g.Players {
		c.Players[i] = player
		c.Players[i].Pieces = append([]int(nil), player.Pieces...)
	}
	return &c
}

func clonePaare(g *PaareGame) *PaareGame {
	if g == nil {
		return nil
	}
	c := *g
	if g.RedDone != nil {
		result := *g.RedDone
		c.RedDone = &result
	}
	if g.BlueDone != nil {
		result := *g.BlueDone
		c.BlueDone = &result
	}
	return &c
}

func cloneDame(g *DameGame) *DameGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Board = make([][]map[string]any, len(g.Board))
	for i, row := range g.Board {
		c.Board[i] = make([]map[string]any, len(row))
		for j, cell := range row {
			if cell == nil {
				continue
			}
			copied := make(map[string]any, len(cell))
			for k, v := range cell {
				copied[k] = v
			}
			c.Board[i][j] = copied
		}
	}
	return &c
}

// startRoomJanitor räumt Spielräume auf, in denen niemand mehr ist. Sonst
// wächst die Raumliste unbegrenzt und die Raumnamen zählen längst verlassene
// Räume mit ("Anna2 Raum", "Anna3 Raum", ...).
func startRoomJanitor() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for code, room := range rooms {
				room.cleanPlayers()
				room.warnIdle(time.Now())
				// Der eine Hauptraum bleibt immer bestehen – auch wenn der Host vom
				// Handy spielt und sein Rechner gerade keinen Tab offen hat.
				if code == mainRoomCode {
					continue
				}
				if len(room.Players) == 0 && time.Since(room.LastActive) > 2*time.Minute {
					delete(rooms, code)
				}
			}
			mu.Unlock()
		}
	}()
}

func jsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	// Spielstände dürfen nie aus dem Zwischenspeicher kommen: Bei wackligem
	// WLAN liefert ein Browser sonst einen alten Zug statt des neuen.
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}
