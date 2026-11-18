package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Nach einem Server-Neustart kennt der Server den Platz nicht mehr. Ohne
// ausdrückliches "neu" wird er abgewiesen – so entstehen keine stillen
// Zweitplätze. Mit "neu" entsteht derselbe Platz wieder: gleiche Kennung,
// gleiches Token, gleicher Name, ein Spieler im Raum.
func TestNeustartLegtPlatzNurMitNeuAn(t *testing.T) {
	now := time.Now()
	code := "TG-NEU7"
	mu.Lock()
	rooms[code] = &Room{Code: code, Name: roomTitle("Anna"), Created: now, LastActive: now,
		Players: map[string]Player{}, Mutes: map[string]time.Time{}, MsgTimes: map[string][]time.Time{}}
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		delete(rooms, code)
		mu.Unlock()
	})

	token := "abcdef0123456789abcdef0123456789"
	ruf := func(neu bool) (*httptest.ResponseRecorder, map[string]string) {
		rumpf, _ := json.Marshal(map[string]any{
			"room": code, "id": "anna", "name": "Anna", "token": token, "neu": neu,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/presence", strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.168.5.5:2222"
		rec := httptest.NewRecorder()
		// Durch die echte Kette, damit auch die Hülle mitgeprüft wird.
		withIdentity(http.HandlerFunc(presence)).ServeHTTP(rec, req)
		var antwort map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &antwort)
		return rec, antwort
	}

	// 1) Unbekannter Platz ohne "neu": abgewiesen, nichts angelegt.
	if rec, _ := ruf(false); rec.Code != http.StatusConflict {
		t.Fatalf("ohne neu: %d statt 409 (%s)", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	mu.Lock()
	spieler := len(rooms[code].Players)
	mu.Unlock()
	if spieler != 0 {
		t.Fatalf("ohne neu wurden %d Plätze angelegt", spieler)
	}

	// 2) Mit "neu": der Platz entsteht.
	rec, antwort := ruf(true)
	if rec.Code != http.StatusOK {
		t.Fatalf("mit neu: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	if antwort["token"] != token || antwort["name"] != "Anna" {
		t.Fatalf("Antwort: %v", antwort)
	}

	// 3) Neustart nachstellen: der Platz ist weg, das Gerät hat sein Token noch.
	mu.Lock()
	delete(rooms[code].Players, "anna")
	mu.Unlock()

	// 4) Lebenszeichen ohne "neu": abgewiesen, kein stiller Zweitplatz.
	if rec, _ := ruf(false); rec.Code != http.StatusConflict {
		t.Fatalf("nach Neustart ohne neu: %d statt 409", rec.Code)
	}

	// 5) Wiederanmelden mit "neu": wieder genau ein Platz, gleiches Token.
	rec, antwort = ruf(true)
	if rec.Code != http.StatusOK {
		t.Fatalf("Wiederanmelden: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	mu.Lock()
	zurueck := rooms[code].Players["anna"]
	spieler = len(rooms[code].Players)
	mu.Unlock()
	if spieler != 1 || zurueck.Token != token || zurueck.Name != "Anna" {
		t.Fatalf("Platz: %d Spieler, Token %q, Name %q", spieler, zurueck.Token, zurueck.Name)
	}
	if antwort["token"] != token {
		t.Fatalf("Antwort-Token %q, erwartet %q", antwort["token"], token)
	}
}
