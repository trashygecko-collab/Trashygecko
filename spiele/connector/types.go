package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
	"unicode/utf8"
)

// maxNameLen ist die maximale Länge eines Spielernamens in Zeichen.
const maxNameLen = 24

type Message struct {
	Name       string `json:"name"`
	Text       string `json:"text"`
	Time       string `json:"time"`
	SenderID   string `json:"senderId"`
	ToID       string `json:"toId,omitempty"`
	ToName     string `json:"toName,omitempty"`
	Whisper    bool   `json:"whisper"`
	GameInvite bool   `json:"gameInvite,omitempty"`
	Game       string `json:"game,omitempty"`
	GameID     string `json:"gameId,omitempty"`
	System     bool   `json:"system,omitempty"`
}

type Player struct {
	Name string    `json:"name"`
	Last time.Time `json:"-"`
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
	Code         string
	OwnerID      string
	OwnerName    string
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
		if now.Sub(p.Last) > 20*time.Second {
			delete(r.Players, id)
			r.addMessage(Message{
				Text:   p.Name + " verlässt die Räuberhöhle",
				Time:   now.Format("15:04:05"),
				System: true,
			})
		}
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
				// Der eine Hauptraum bleibt immer bestehen.
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
	_ = json.NewEncoder(w).Encode(v)
}
