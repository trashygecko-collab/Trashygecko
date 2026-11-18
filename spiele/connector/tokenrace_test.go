package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Der Beitritt darf nicht mehr an zwei fast gleichzeitigen Presence-Rufen
// scheitern: das Gerät bringt sein Token selbst mit, beide Rufe schicken es.
func TestBeitrittTokenRace(t *testing.T) {
	now := time.Now()
	code := "TG-TOK9"
	room := &Room{Code: code, Name: roomTitle("Anna"), Created: now, LastActive: now,
		Players: map[string]Player{}, Mutes: map[string]time.Time{}, MsgTimes: map[string][]time.Time{}}
	mu.Lock()
	rooms[code] = room
	mu.Unlock()
	defer func() {
		mu.Lock()
		delete(rooms, code)
		mu.Unlock()
	}()

	beitreten := func(id, token string) *httptest.ResponseRecorder {
		rumpf, _ := json.Marshal(map[string]any{"room": code, "id": id, "name": "Anna", "token": token, "neu": true})
		req := httptest.NewRequest(http.MethodPost, "/api/presence", strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.168.7.9:51234"
		rec := httptest.NewRecorder()
		// Durch die echte Kette: die 403 für ein fremdes Token entsteht in
		// withIdentity, nicht im Beitritt selbst. Ein Test, der die Hülle
		// überspringt, würde einen Rückfall dort nicht bemerken.
		withIdentity(http.HandlerFunc(presence)).ServeHTTP(rec, req)
		return rec
	}

	// Wie im Browser: erst Token erzeugen, dann zwei Rufe gleichzeitig starten.
	mitgebracht := "0123456789abcdef0123456789abcdef"
	var wg sync.WaitGroup
	antworten := make([]*httptest.ResponseRecorder, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			antworten[i] = beitreten("race-1", mitgebracht)
		}(i)
	}
	wg.Wait()
	for i, rec := range antworten {
		if rec.Code != http.StatusOK {
			t.Fatalf("Ruf %d: %d %s", i, rec.Code, strings.TrimSpace(rec.Body.String()))
		}
		var antwort map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &antwort); err != nil {
			t.Fatalf("Ruf %d: keine Antwort (%v)", i, err)
		}
		if antwort["token"] != mitgebracht {
			t.Fatalf("Ruf %d: Token %q, erwartet %q", i, antwort["token"], mitgebracht)
		}
	}
	mu.Lock()
	gespeichert := rooms[code].Players["race-1"].Token
	spielerZahl := len(rooms[code].Players)
	mu.Unlock()
	if gespeichert != mitgebracht || spielerZahl != 1 {
		t.Fatalf("Platz: Token %q, Spieler %d", gespeichert, spielerZahl)
	}

	// Fremdes Token bleibt gesperrt.
	if rec := beitreten("race-1", "ffffffffffffffffffffffffffffffff"); rec.Code != http.StatusForbidden {
		t.Fatalf("fremdes Token: %d statt 403", rec.Code)
	}
	// Ohne Token bleibt ein bekannter Platz ebenfalls gesperrt.
	if rec := beitreten("race-1", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("ohne Token: %d statt 403", rec.Code)
	}

	// Alte Wege unverändert: ohne Token erzeugt der Server eines, und der
	// nächste Ruf ohne Token kommt nicht mehr durch (so war es schon immer).
	erst := beitreten("race-2", "")
	if erst.Code != http.StatusOK {
		t.Fatalf("neuer Platz ohne Token: %d", erst.Code)
	}
	var antwort2 map[string]string
	_ = json.Unmarshal(erst.Body.Bytes(), &antwort2)
	if len(antwort2["token"]) < 16 {
		t.Fatalf("Server-Token zu kurz: %q", antwort2["token"])
	}
	if rec := beitreten("race-2", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("zweiter Ruf ohne Token: %d statt 403", rec.Code)
	}

	// Unbrauchbare Geheimnisse ersetzt der Server.
	for _, schlecht := range []string{"kurz", strings.Repeat("a", 65), "abcdefghij klmnop", "abcdefghijklmnop;drop"} {
		if tokenBrauchbar(schlecht) {
			t.Fatalf("%q gilt als brauchbar", schlecht)
		}
	}
	if !tokenBrauchbar(mitgebracht) {
		t.Fatal("gültiges Token abgelehnt")
	}
}
