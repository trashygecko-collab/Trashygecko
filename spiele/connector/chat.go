package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// messages handles GET (fetch) and POST (send) chat messages.
func messages(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(r.URL.Query().Get("room"))
	if code == "" {
		http.Error(w, "Spielraum-Code fehlt", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		viewerID := r.URL.Query().Get("id")
		mu.Lock()
		room := rooms[code]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		// Clean old data
		room.cleanPlayers()
		// Prepare copy respecting whispers
		copyMessages := make([]Message, 0, len(room.Messages))
		for _, msg := range room.Messages {
			if !msg.Whisper || msg.SenderID == viewerID || msg.ToID == viewerID {
				copyMessages = append(copyMessages, msg)
			}
		}
		mu.Unlock()
		jsonResponse(w, copyMessages)
		return
	}

	if r.Method == http.MethodPost {
		var input struct {
			ID   string `json:"id"`
			ToID string `json:"toId"`
			Name string `json:"name"`
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Ungültige Nachricht", http.StatusBadRequest)
			return
		}
		input.ID = strings.TrimSpace(input.ID)
		input.ToID = strings.TrimSpace(input.ToID)
		input.Name = strings.TrimSpace(input.Name)
		input.Text = strings.TrimSpace(input.Text)
		if input.Name == "" {
			input.Name = "Spieler"
		}
		input.Name = limitName(input.Name)
		if input.Text == "" || utf8.RuneCountInString(input.Text) > 1000 {
			http.Error(w, "Nachricht leer oder zu lang", http.StatusBadRequest)
			return
		}

		mu.Lock()
		room := rooms[code]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		// Verify sender exists
		sender, ok := room.Players[input.ID]
		if !ok {
			mu.Unlock()
			http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
			return
		}
		// Update sender name if changed
		if sender.Name != input.Name {
			sender.Name = input.Name
			room.Players[input.ID] = sender
		}
		// Check mute status
		if muted, remaining := room.isMuted(input.ID); muted {
			mu.Unlock()
			http.Error(w, "Du bist stummgeschaltet für "+remaining.Truncate(time.Second).String(), http.StatusForbidden)
			return
		}
		// Spam protection: max 5 messages per 10 seconds
		now := time.Now()
		timestamps := room.MsgTimes[input.ID]
		// Keep only recent timestamps
		recent := []time.Time{}
		for _, t := range timestamps {
			if now.Sub(t) < 10*time.Second {
				recent = append(recent, t)
			}
		}
		if len(recent) >= 5 {
			mu.Unlock()
			http.Error(w, "Zu viele Nachrichten in kurzer Zeit", http.StatusTooManyRequests)
			return
		}
		recent = append(recent, now)
		room.MsgTimes[input.ID] = recent

		// Determine recipient name if whisper
		recipientName := ""
		if input.ToID != "" {
			recipient, ok := room.Players[input.ToID]
			if !ok || input.ToID == input.ID {
				mu.Unlock()
				http.Error(w, "Empfänger nicht verfügbar", http.StatusBadRequest)
				return
			}
			recipientName = recipient.Name
		}
		// Append message
		room.addMessage(Message{
			Name:     sender.Name,
			Text:     input.Text,
			Time:     now.Format("15:04:05"),
			SenderID: input.ID,
			ToID:     input.ToID,
			ToName:   recipientName,
			Whisper:  input.ToID != "",
		})
		room.LastActive = now
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
}
