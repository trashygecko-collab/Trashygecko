package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

// adminPasswort gehört zum Chat-Befehl "/Admin <Passwort>". Damit wird man
// auch dann Admin, wenn der Platz schon vergeben ist (online die Regel).
//
// Es steht bewusst nicht mehr fest im Quelltext: Der Connector liegt samt
// Quellen öffentlich, "Admin" war damit für jeden lesbar und jeder Besucher
// konnte sich zum Besitzer machen. Jetzt gilt TG_ADMIN_PASSWORT, sonst ein
// Zufallswert je Start, den der Betreiber beim Start im Protokoll sieht.
var adminPasswort = zufallsToken()

func init() {
	if eigen := strings.TrimSpace(os.Getenv("TG_ADMIN_PASSWORT")); eigen != "" {
		adminPasswort = eigen
		return
	}
	log.Printf("Admin-Passwort für /Admin (nur dieser Start): %s", adminPasswort)
}

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
			Text string `json:"text"`
			// Bild: Name einer Grafik aus dem eigenen Bestand (optional).
			Bild string `json:"bild"`
			// MsgID vergibt die Seite je Nachricht. Ein Wiederholungsversuch
			// schickt dieselbe Kennung – der Raum erkennt die doppelte Zeile.
			MsgID string `json:"msgId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Ungültige Nachricht", http.StatusBadRequest)
			return
		}
		input.ID = strings.TrimSpace(input.ID)
		input.ToID = strings.TrimSpace(input.ToID)
		input.Text = strings.TrimSpace(input.Text)
		input.MsgID = strings.TrimSpace(input.MsgID)
		if len(input.MsgID) > 64 {
			input.MsgID = input.MsgID[:64]
		}
		// Nur Grafiken aus dem eigenen Bestand: keine fremden Adressen, keine
		// Pfade aus dem Ordner heraus, keine anderen Dateitypen.
		input.Bild = strings.TrimSpace(input.Bild)
		if input.Bild != "" {
			name, ok := smileyErlaubt(input.Bild)
			if !ok {
				http.Error(w, "Diese Grafik gehört nicht zum Smiley-Bestand.", http.StatusBadRequest)
				return
			}
			input.Bild = name
		} else {
			input.Bild = ""
		}
		// Eine Nachricht braucht Text oder eine Grafik – beides zusammen ist
		// erlaubt, nur Text bleibt wie bisher.
		if (input.Text == "" && input.Bild == "") || utf8.RuneCountInString(input.Text) > 1000 {
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
		// Der Name kommt allein aus dem Raum (Beitritt oder /api/umbenennen).
		// Eine Seite, die noch den alten Namen kennt, macht ein Umbenennen sonst
		// mit jeder Nachricht rückgängig – dieselbe Regel wie bei /api/presence:
		// dort gilt für einen bekannten Platz ebenfalls nur der Raum.
		sender.LastAction = time.Now()
		sender.AfkWarned = false
		room.Players[input.ID] = sender
		// Check mute status
		if muted, remaining := room.isMuted(input.ID); muted {
			mu.Unlock()
			http.Error(w, "Du bist stummgeschaltet für "+remaining.Truncate(time.Second).String(), http.StatusForbidden)
			return
		}
		// "/Admin Admin": Im Internet kann der Boss-Platz schon vergeben sein.
		// Wer das Passwort kennt, wird damit Admin. Der Befehl selbst landet
		// nie im Chat, damit das Passwort nicht für alle sichtbar wird.
		if strings.HasPrefix(strings.ToLower(input.Text), "/admin") {
			teile := strings.Fields(input.Text)
			if len(teile) == 2 && strings.EqualFold(teile[1], adminPasswort) {
				room.OwnerID = input.ID
				room.OwnerName = sender.Name
				room.Name = roomTitle(sender.Name)
				jetzt := time.Now()
				room.addMessage(Message{
					Name:     sender.Name,
					Text:     "ist jetzt Admin.",
					Time:     jetzt.Format("15:04:05"),
					SenderID: input.ID,
				})
				room.LastActive = jetzt
				mu.Unlock()
				w.WriteHeader(http.StatusNoContent)
				return
			}
			mu.Unlock()
			http.Error(w, "Falsches Passwort für /Admin.", http.StatusForbidden)
			return
		}
		// Dieselbe Nachricht kommt bei einem Wiederholungsversuch ein zweites
		// Mal an. Die msgId der Seite entscheidet: einmal angenommen, danach
		// nur noch die Bestätigung – so entsteht keine zweite Zeile im Verlauf.
		if vorhanden, gefunden := room.nachrichtMitID(input.MsgID, input.ID); gefunden {
			mu.Unlock()
			jsonResponse(w, map[string]string{"msgId": input.MsgID, "time": vorhanden.Time})
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
		uhrzeit := now.Format("15:04:05")
		room.addMessage(Message{
			Name:     sender.Name,
			Text:     input.Text,
			Bild:     input.Bild,
			Time:     uhrzeit,
			SenderID: input.ID,
			ToID:     input.ToID,
			ToName:   recipientName,
			Whisper:  input.ToID != "",
			MsgID:    input.MsgID,
		})
		room.LastActive = now
		mu.Unlock()
		// Empfangsbestätigung: die Seite hängt ihre Zeile erst daran ab.
		jsonResponse(w, map[string]string{"msgId": input.MsgID, "time": uhrzeit})
		return
	}

	http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
}
