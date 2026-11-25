package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// moderationHandler allows room owners to mute players and view mute status.
// POST with {room, ownerId, targetId, minutes} to set mute (minutes 1-3).
// GET with ?room=CODE returns list of muted players and remaining seconds.
// mutesListe liefert die laufenden Stummschaltungen mit Restsekunden und wirft
// abgelaufene weg. Muss unter mu laufen.
func mutesListe(room *Room, jetzt time.Time) []map[string]string {
	list := []map[string]string{}
	for id, bis := range room.Mutes {
		if jetzt.After(bis) {
			delete(room.Mutes, id)
			continue
		}
		name := ""
		if p, ok := room.Players[id]; ok {
			name = p.Name
		}
		list = append(list, map[string]string{"id": id, "name": name,
			"seconds": strconv.Itoa(int(bis.Sub(jetzt).Seconds()))})
	}
	return list
}

func moderationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var input struct {
			Room     string `json:"room"`
			OwnerID  string `json:"ownerId"`
			TargetID string `json:"targetId"`
			Minutes  int    `json:"minutes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Ungültige Daten", http.StatusBadRequest)
			return
		}
		input.Room = raumAusKontext(r)
		input.OwnerID = kennung(r)
		input.TargetID = strings.TrimSpace(input.TargetID)
		if input.Room == "" || input.OwnerID == "" || input.TargetID == "" {
			http.Error(w, "Parameter fehlen", http.StatusBadRequest)
			return
		}
		if input.Minutes < 1 || input.Minutes > 3 {
			http.Error(w, "Ungültige Dauer", http.StatusBadRequest)
			return
		}
		mu.Lock()
		room := rooms[input.Room]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		if room.OwnerID != input.OwnerID {
			mu.Unlock()
			http.Error(w, "Nur der Raum‑Besitzer kann stummschalten", http.StatusForbidden)
			return
		}
		// Set mute until time
		room.Mutes[input.TargetID] = time.Now().Add(time.Duration(input.Minutes) * time.Minute)
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == http.MethodGet {
		roomCode := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
		if roomCode == "" {
			http.Error(w, "Raum‑Code fehlt", http.StatusBadRequest)
			return
		}
		mu.Lock()
		room := rooms[roomCode]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		list := mutesListe(room, time.Now())
		mu.Unlock()
		jsonResponse(w, list)
		return
	}

	http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
}
