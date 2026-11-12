package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

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
	Code      string
	OwnerID   string
	OwnerName string
	Name      string // e.g. "Spielername Raum"
	Created   time.Time
	Players   map[string]Player
	Messages    []Message
	TicTacToe   *TicTacToeGame
	VierGewinnt *VierGewinntGame
	Dame        *DameGame
	Mutes       map[string]time.Time   // playerID -> unmute time
	MsgTimes  map[string][]time.Time // for spam protection
}

type RoomInfo struct {
	Code    string `json:"code"`
	Players int    `json:"players"`
	Created string `json:"created"`
	Owner   string `json:"owner,omitempty"`
	Status  string `json:"status,omitempty"`
	Game    string `json:"game,omitempty"`
	Name    string `json:"name,omitempty"`
}

var (
	rooms = map[string]*Room{}
	mu    sync.Mutex
)

func (r *Room) cleanPlayers() {
	now := time.Now()
	for id, p := range r.Players {
		if now.Sub(p.Last) > 20*time.Second {
			delete(r.Players, id)
		}
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
	return "Verfügbar"
}

func jsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
