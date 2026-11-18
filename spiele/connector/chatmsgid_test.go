package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Ein Wiederholungsversuch schickt dieselbe Nachricht noch einmal. Die msgId
// sorgt dafür, dass sie im Verlauf genau einmal steht und trotzdem eine
// Empfangsbestätigung zurückkommt.
func TestChatMsgIDVerhindertDoppelteZeile(t *testing.T) {
	now := time.Now()
	code := "TG-MSG7"
	token := "dddddddddddddddddddddddddddddddd"
	mu.Lock()
	rooms[code] = &Room{Code: code, Name: roomTitle("Anna"), Created: now, LastActive: now,
		Players: map[string]Player{"anna": {Name: "Anna", Token: token, Last: now}},
		Mutes:   map[string]time.Time{}, MsgTimes: map[string][]time.Time{}}
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		delete(rooms, code)
		mu.Unlock()
	})

	senden := func(msgID, text string) (*httptest.ResponseRecorder, map[string]string) {
		rumpf, _ := json.Marshal(map[string]any{"id": "anna", "text": text, "msgId": msgID})
		req := httptest.NewRequest(http.MethodPost, "/api/messages?room="+code+"&token="+token, strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.168.9.8:4100"
		rec := httptest.NewRecorder()
		withIdentity(http.HandlerFunc(messages)).ServeHTTP(rec, req)
		var antwort map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &antwort)
		return rec, antwort
	}
	verlauf := func() []Message {
		mu.Lock()
		defer mu.Unlock()
		return append([]Message{}, rooms[code].Messages...)
	}
	zeilen := func(text string) int {
		n := 0
		for _, m := range verlauf() {
			if m.Text == text && m.SenderID == "anna" {
				n++
			}
		}
		return n
	}

	// Erster Versuch.
	rec, antwort := senden("msg-1", "Hallo zusammen")
	if rec.Code != http.StatusOK {
		t.Fatalf("erster Versuch: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	if antwort["msgId"] != "msg-1" || antwort["time"] == "" {
		t.Fatalf("Bestätigung: %v", antwort)
	}

	// Derselbe Versuch noch einmal (Antwort war unterwegs verloren).
	rec, antwort = senden("msg-1", "Hallo zusammen")
	if rec.Code != http.StatusOK {
		t.Fatalf("Wiederholung: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	if antwort["msgId"] != "msg-1" {
		t.Fatalf("Bestätigung der Wiederholung: %v", antwort)
	}
	if n := zeilen("Hallo zusammen"); n != 1 {
		t.Fatalf("die Nachricht steht %d-mal im Verlauf", n)
	}

	// Ohne msgId bleibt der alte Weg unverändert (zwei Nachrichten).
	if rec, _ := senden("", "zwei"); rec.Code != http.StatusOK {
		t.Fatalf("ohne msgId: %d", rec.Code)
	}
	if rec, _ := senden("", "zwei"); rec.Code != http.StatusOK {
		t.Fatalf("ohne msgId (2): %d", rec.Code)
	}
	if n := zeilen("zwei"); n != 2 {
		t.Fatalf("ohne msgId: %d Zeilen statt 2", n)
	}

	// Zwei verschiedene Kennungen mit gleichem Text bleiben zwei Zeilen.
	if rec, _ := senden("msg-2", "gleich"); rec.Code != http.StatusOK {
		t.Fatalf("msg-2: %d", rec.Code)
	}
	if rec, _ := senden("msg-3", "gleich"); rec.Code != http.StatusOK {
		t.Fatalf("msg-3: %d", rec.Code)
	}
	if n := zeilen("gleich"); n != 2 {
		t.Fatalf("zwei Kennungen: %d Zeilen statt 2", n)
	}
}
