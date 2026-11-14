package main

import (
	"encoding/json"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

// mainRoomCode ist der eine Raum, den der Connector anbietet.
var mainRoomCode string

// ensureMainRoom legt diesen Raum an. Es gibt bewusst nur einen: der Spieler
// am Server-PC ist Admin, alle anderen kommen in denselben Raum.
func ensureMainRoom() {
	mu.Lock()
	defer mu.Unlock()
	if mainRoomCode != "" && rooms[mainRoomCode] != nil {
		return
	}
	code := randomCode()
	for rooms[code] != nil {
		code = randomCode()
	}
	mainRoomCode = code
	rooms[code] = &Room{
		Code:       code,
		Created:    time.Now(),
		LastActive: time.Now(),
		Players:    map[string]Player{},
		Messages:   []Message{},
		Mutes:      map[string]time.Time{},
		MsgTimes:   map[string][]time.Time{},
	}
}

// roomTitle bildet den Raumnamen aus dem Spielernamen: "Anna" -> "Annas Abenteuerland".
func roomTitle(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Abenteuerland"
	}
	// Bei Namen auf s, x, z oder ß klingt ein zweites s falsch: "Klaus' Abenteuerland".
	last := name[len(name)-1:]
	if last == "s" || last == "x" || last == "z" || last == "ß" {
		return name + "' Abenteuerland"
	}
	return name + "s Abenteuerland"
}

// isServerMachine meldet, ob die Anfrage von diesem PC kommt. Nur dort läuft der
// Connector – also sitzt dort der Host, der automatisch Admin ist.
func isServerMachine(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	for _, local := range localIPv4s() {
		if local == ip.String() {
			return true
		}
	}
	return false
}

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
			} else if room.Paare != nil && (room.Paare.Status == "active" || room.Paare.Status == "invited") {
				game = "Paare finden"
			} else if room.Kaese != nil && (room.Kaese.Status == "active" || room.Kaese.Status == "invited") {
				game = "Käsekästchen"
			} else if room.Sim != nil && (room.Sim.Status == "active" || room.Sim.Status == "invited") {
				game = "Sim"
			} else if room.Kalaha != nil && (room.Kalaha.Status == "active" || room.Kalaha.Status == "invited") {
				game = "Kalaha"
			} else if room.Streichholz != nil && (room.Streichholz.Status == "active" || room.Streichholz.Status == "invited") {
				game = "Streichholzspiel"
			} else if room.Muehle != nil && (room.Muehle.Status == "active" || room.Muehle.Status == "invited") {
				game = "Mühle"
			} else if room.Reversi != nil && (room.Reversi.Status == "active" || room.Reversi.Status == "invited") {
				game = "Reversi"
			} else if room.Hold != nil && (room.Hold.Status == "active" || room.Hold.Status == "invited") {
				game = "Hold the Line"
			} else if room.Skyjo != nil && (room.Skyjo.Status == "active" || room.Skyjo.Status == "invited") {
				game = "Skyjo"
			} else if room.Kniffel != nil && (room.Kniffel.Status == "active" || room.Kniffel.Status == "invited") {
				game = "Kniffel"
			} else if room.Schach != nil && (room.Schach.Status == "active" || room.Schach.Status == "invited") {
				game = "Schach"
			} else if room.Go != nil && (room.Go.Status == "active" || room.Go.Status == "invited") {
				game = "Go"
			} else if room.Backgammon != nil && (room.Backgammon.Status == "active" || room.Backgammon.Status == "invited") {
				game = "Backgammon"
			} else if room.Menschaerger != nil && (room.Menschaerger.Status == "active" || room.Menschaerger.Status == "invited") {
				game = "Mensch ärgere dich nicht"
			}
			list = append(list, RoomInfo{Code: code, Players: len(room.Players), Created: room.Created.Format("15:04"), CreatedAt: room.Created, Owner: room.OwnerName, Status: status, Game: game, Name: room.Name})
		}
		mu.Unlock()
		sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.Before(list[j].CreatedAt) })
		jsonResponse(w, list)
		return
	}

	if r.Method == http.MethodPost {
		var input struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&input)
		input.ID = strings.TrimSpace(input.ID)
		pName := strings.TrimSpace(input.Name)
		if pName == "" {
			pName = "Spieler"
		}
		pName = limitName(pName)

		ensureMainRoom()
		mu.Lock()
		room := rooms[mainRoomCode]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht verfügbar", http.StatusServiceUnavailable)
			return
		}
		code := room.Code
		displayName := room.Name
		if displayName == "" {
			// Nur für die Anzeige – der Name wird beim Beitreten gesetzt.
			displayName = roomTitle(pName)
		}
		mu.Unlock()
		jsonResponse(w, map[string]string{"code": code, "name": displayName})
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
		// Active meldet, ob der Spieler seit dem letzten Lebenszeichen etwas
		// getan hat (Klick, Tastendruck). Ohne das zählt er als abwesend.
		Active bool `json:"active"`
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
	input.Name = limitName(input.Name)
	// Wer ohne eigenen Namen hereinkommt, bekommt einen: am Server-PC "Boss",
	// alle anderen "Spieler1", "Spieler2", … – umbenennen geht jederzeit.
	amServer := isServerMachine(r)

	mu.Lock()
	room := rooms[input.Room]
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	// Ein Spieler gehört immer nur zu genau einem Spielraum.
	for otherCode, other := range rooms {
		if otherCode == input.Room {
			continue
		}
		if gewechselt, ok := other.Players[input.ID]; ok {
			delete(other.Players, input.ID)
			// Wer in einen anderen Raum wechselt, verlässt diesen hier.
			other.addMessage(Message{
				Name:     gewechselt.Name,
				Text:     "verlässt den Raum.",
				Time:     time.Now().Format("15:04:05"),
				SenderID: input.ID,
				Gone:     true,
			})
			other.cleanPlayers()
			other.LastActive = time.Now()
		}
	}
	room.cleanPlayers()
	if input.Name == "" {
		if amServer {
			input.Name = "Boss"
		} else {
			input.Name = room.freierName("Spieler")
		}
	}
	// Doppelte Namen bekommen eine Nummer – sonst könnte sich jeder als der
	// Admin ausgeben.
	name := room.uniqueName(input.Name, input.ID)
	// Admin ist der Spieler am Server-PC. Seine Anfrage kommt von einer lokalen
	// Adresse, egal ob über localhost oder die LAN-IP aufgerufen wird.
	if room.OwnerID == "" || (room.OwnerID != input.ID && amServer) {
		room.OwnerID = input.ID
		room.OwnerName = name
		room.Name = roomTitle(name)
	}
	player, known := room.Players[input.ID]
	if !known {
		// Mit Absender, damit der Name im Chat steht und angeklickt werden kann.
		text := "betritt den Raum."
		if amServer {
			text = "ist da!"
		}
		room.addMessage(Message{
			Name:     name,
			Text:     text,
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})
	} else if player.Name != "" && player.Name != name {
		// Umbenennen einmal im Chat bekanntgeben – mit dem alten Namen als
		// Absender, damit die Zeile "Anna: heißt jetzt Alex." lesbar bleibt.
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "heißt jetzt " + name + ".",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.ID,
		})
	}
	player.Name = name
	player.Last = time.Now()
	if player.LastAction.IsZero() || input.Active {
		player.LastAction = time.Now()
		player.AfkWarned = false
	}
	room.Players[input.ID] = player
	room.LastActive = time.Now()
	code := room.Code
	mu.Unlock()
	// Der vergebene Name geht zurück, damit die Seite ihn anzeigen kann.
	jsonResponse(w, map[string]string{"name": name, "room": code})
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
