package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// roomsHandler handles GET (list rooms) and POST (create room).
func roomsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		mu.Lock()
		list := []RoomInfo{}
		for code, room := range rooms {
			room.cleanPlayers()
			// Determine status
			status := room.gameStatus()
			// Determine current game name if any
			game := ""
			if room.TicTacToe != nil && (room.TicTacToe.Status == "active" || room.TicTacToe.Status == "invited") {
				game = "TicTacToe"
			} else if room.VierGewinnt != nil && (room.VierGewinnt.Status == "active" || room.VierGewinnt.Status == "invited") {
				game = "Vier gewinnt"
			} else if room.Dame != nil && (room.Dame.Status == "active" || room.Dame.Status == "invited") {
				game = "Dame"
			}
			list = append(list, RoomInfo{Code: code, Players: len(room.Players), Created: room.Created.Format("15:04"), Owner: room.OwnerName, Status: status, Game: game, Name: room.Name})
		}
		mu.Unlock()
		sort.Slice(list, func(i, j int) bool { return list[i].Created < list[j].Created })
		jsonResponse(w, list)
		return
	}

	if r.Method == http.MethodPost {
		var input struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&input)
		pName := strings.TrimSpace(input.Name)
		if pName == "" {
			pName = "Spieler"
		}
		if len(pName) > 24 {
			pName = pName[:24]
		}

		mu.Lock()
		code := randomCode()
		for rooms[code] != nil {
			code = randomCode()
		}

		// Calculate room name with numbering: Spielername Raum, Spielername2 Raum, etc.
		baseName := pName + " Raum"
		targetName := baseName
		counter := 2
		// Check for duplicate names among active rooms
		nameExists := func(n string) bool {
			for _, rm := range rooms {
				if strings.EqualFold(rm.Name, n) {
					return true
				}
			}
			return false
		}
		for nameExists(targetName) {
			targetName = pName + strconv.Itoa(counter) + " Raum"
			counter++
		}

		rooms[code] = &Room{
			Code:      code,
			OwnerID:   strings.TrimSpace(input.ID),
			OwnerName: pName,
			Name:      targetName,
			Created:   time.Now(),
			Players:   map[string]Player{},
			Messages:  []Message{},
			Mutes:     map[string]time.Time{},
			MsgTimes:  map[string][]time.Time{},
		}
		mu.Unlock()
		jsonResponse(w, map[string]string{"code": code, "name": targetName})
		return
	}

	http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
}

// presence registers or updates a player's presence in a room.
func presence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erforderlich", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		Room string `json:"room"`
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.Room = strings.ToUpper(strings.TrimSpace(input.Room))
	input.ID = strings.TrimSpace(input.ID)
	input.Name = strings.TrimSpace(input.Name)
	if input.ID == "" || input.Room == "" {
		http.Error(w, "Spielraum oder Spielerkennung fehlt", http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		input.Name = "Spieler"
	}
	if len(input.Name) > 24 {
		input.Name = input.Name[:24]
	}

	mu.Lock()
	room := rooms[input.Room]
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	room.cleanPlayers()
	// Assign owner if not set
	if room.OwnerID == "" {
		room.OwnerID = input.ID
		room.OwnerName = input.Name
		if room.Name == "" {
			room.Name = input.Name + " Raum"
		}
	}
	// Ensure player is only in one room: others will be auto-removed when they join another via presence.
	room.Players[input.ID] = Player{Name: input.Name, Last: time.Now()}
	mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// playersHandler returns list of players in a room and room meta (like ownerId).
func playersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET erforderlich", http.StatusMethodNotAllowed)
		return
	}
	code := strings.ToUpper(r.URL.Query().Get("room"))
	mu.Lock()
	room := rooms[code]
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	room.cleanPlayers()
	list := make([]map[string]string, 0, len(room.Players))
	for id, player := range room.Players {
		list = append(list, map[string]string{"id": id, "name": player.Name})
	}
	ownerID := room.OwnerID
	mu.Unlock()
	sort.Slice(list, func(i, j int) bool { return strings.ToLower(list[i]["name"]) < strings.ToLower(list[j]["name"]) })
	w.Header().Set("X-Room-Owner", ownerID)
	jsonResponse(w, list)
}
