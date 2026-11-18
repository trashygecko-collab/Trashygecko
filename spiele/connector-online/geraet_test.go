package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Geht das Token verloren, holt die Gerätekennung denselben Platz zurück –
// gleiche Kennung, gleicher Name, kein Nummernanhang und kein zweiter Platz.
// Eine fremde Gerätekennung kommt damit nicht durch, und nach zehn Minuten
// gilt nur noch das Token.
func TestGeraetHoltEigenenPlatzZurueck(t *testing.T) {
	now := time.Now()
	code := "TG-GER7"
	mu.Lock()
	rooms[code] = &Room{Code: code, Name: roomTitle("Anna"), Created: now, LastActive: now,
		Players: map[string]Player{}, Mutes: map[string]time.Time{}, MsgTimes: map[string][]time.Time{}}
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		delete(rooms, code)
		mu.Unlock()
	})

	geraet := "geraet0123456789abcdef0123456789"
	anderes := "geraet9876543210fedcba9876543210"
	tokenA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	tokenB := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	tokenC := "cccccccccccccccccccccccccccccccc"

	ruf := func(token, geraetWert string, neu bool) (*httptest.ResponseRecorder, map[string]string) {
		rumpf, _ := json.Marshal(map[string]any{
			"room": code, "id": "anna", "name": "Anna", "token": token, "geraet": geraetWert, "neu": neu,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/presence", strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.168.9.7:4000"
		rec := httptest.NewRecorder()
		withIdentity(http.HandlerFunc(presence)).ServeHTTP(rec, req)
		var antwort map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &antwort)
		return rec, antwort
	}
	spieler := func() (int, Player) {
		mu.Lock()
		defer mu.Unlock()
		return len(rooms[code].Players), rooms[code].Players["anna"]
	}

	// 1) Erster Beitritt merkt sich das Gerät.
	if rec, _ := ruf(tokenA, geraet, true); rec.Code != http.StatusOK {
		t.Fatalf("Beitritt: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}

	// 2) Token verloren, gleiches Gerät: derselbe Platz, neues Token.
	rec, antwort := ruf(tokenB, geraet, false)
	if rec.Code != http.StatusOK {
		t.Fatalf("Gerät zurückholen: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	if antwort["token"] != tokenB || antwort["name"] != "Anna" {
		t.Fatalf("Antwort: %v", antwort)
	}
	anzahl, platz := spieler()
	if anzahl != 1 || platz.Token != tokenB || platz.Name != "Anna" {
		t.Fatalf("Platz: %d Spieler, Token %q, Name %q", anzahl, platz.Token, platz.Name)
	}

	// 3) Fremdes Gerät mit geratener Kennung: kein Zugriff.
	if rec, _ := ruf(tokenC, anderes, false); rec.Code != http.StatusForbidden {
		t.Fatalf("fremdes Gerät: %d statt 403", rec.Code)
	}
	// 4) Gültiges Token ohne Gerätekennung funktioniert weiter.
	if rec, _ := ruf(tokenB, "", false); rec.Code != http.StatusOK {
		t.Fatalf("gültiges Token: %d statt 200", rec.Code)
	}
	// 5) Nach zehn Minuten gilt nur noch das Token.
	mu.Lock()
	alt := rooms[code].Players["anna"]
	alt.Last = time.Now().Add(-11 * time.Minute)
	rooms[code].Players["anna"] = alt
	mu.Unlock()
	if rec, _ := ruf(tokenC, geraet, false); rec.Code != http.StatusForbidden {
		t.Fatalf("nach der Frist: %d statt 403", rec.Code)
	}
	if anzahl, _ := spieler(); anzahl != 1 {
		t.Fatalf("es sind %d Plätze statt 1", anzahl)
	}

	// Unbrauchbare Gerätekennungen weist der Server ab, statt sie zu übernehmen.
	for _, schlecht := range []string{"kurz", strings.Repeat("a", 129), "abcdefghij klmnop", "abcdefghijklmnop;"} {
		if geraetBrauchbar(schlecht) {
			t.Fatalf("%q gilt als brauchbare Gerätekennung", schlecht)
		}
	}
	if !geraetBrauchbar(geraet) {
		t.Fatal("gültige Gerätekennung abgelehnt")
	}
}
