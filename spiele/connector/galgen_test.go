package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGalgenRegeln(t *testing.T) {
	now := time.Now()
	room := &Room{Code: "TG-GAL1", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
		Created: now, LastActive: now,
		Players: map[string]Player{
			"a": {Name: "Anna", Token: "ta", Last: now, LastAction: now},
			"b": {Name: "Ben", Token: "tb", Last: now, LastAction: now},
			"c": {Name: "Cem", Token: "tc", Last: now, LastAction: now},
		},
		Mutes: map[string]time.Time{}, MsgTimes: map[string][]time.Time{}}
	mu.Lock()
	rooms[room.Code] = room
	mu.Unlock()
	defer func() { mu.Lock(); delete(rooms, room.Code); mu.Unlock() }()

	ruf := func(id, aktion string, koerper map[string]any) *httptest.ResponseRecorder {
		if koerper == nil {
			koerper = map[string]any{}
		}
		koerper["id"] = id
		rumpf, _ := json.Marshal(koerper)
		req := httptest.NewRequest(http.MethodPost, "/api/galgen?room="+room.Code+"&action="+aktion, strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		galgenHandler(rec, req)
		return rec
	}
	// Das Wort steht nur serverseitig – der Test darf hineinschauen.
	geheim := func() string {
		mu.Lock()
		defer mu.Unlock()
		if g := rooms[room.Code].Galgen; g != nil {
			return g.Wort
		}
		return ""
	}
	spiel := func() *GalgenGame {
		req := httptest.NewRequest(http.MethodGet, "/api/galgen?room="+room.Code+"&id=a", nil)
		rec := httptest.NewRecorder()
		galgenHandler(rec, req)
		var antwort struct {
			Game *GalgenGame `json:"game"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &antwort)
		return antwort.Game
	}

	if r := ruf("a", "create", map[string]any{"variante": "klassisch"}); r.Code != http.StatusOK {
		t.Fatalf("create: %d %s", r.Code, r.Body.String())
	}
	ruf("b", "join", nil)
	ruf("c", "join", nil)
	ruf("a", "ready", nil)
	ruf("b", "ready", nil)
	ruf("c", "ready", nil)
	if r := ruf("a", "start", nil); r.Code != http.StatusOK {
		t.Fatalf("start: %d %s", r.Code, r.Body.String())
	}
	g := spiel()
	if g.Status != "spiel" || len(g.Spieler) != 3 || len(g.Anzeige) < 3 {
		t.Fatalf("Start: %+v", g)
	}
	// Falscher Spieler darf nicht raten.
	dran := g.Spieler[g.Zug].ID
	anderer := ""
	for _, s := range g.Spieler {
		if s.ID != dran {
			anderer = s.ID
		}
	}
	if r := ruf(anderer, "raten", map[string]any{"buchstabe": "e"}); r.Code != http.StatusConflict {
		t.Fatalf("fremder Zug: %d", r.Code)
	}
	// Richtige Buchstaben raten, bis das Wort steht – Punkte müssen stimmen.
	gesamt := 0
	for runde := 0; runde < 200; runde++ {
		g = spiel()
		if g.Status != "spiel" {
			break
		}
		sitz := g.Spieler[g.Zug]
		// Das Wort ist absichtlich nicht Teil der Ansicht – also der Reihe nach
		// das Alphabet raten, bis es steht.
		buchstabe := ""
		for _, k := range geheim() {
			b := strings.ToLower(string(k))
			if !enthaelt(g.Buchstaben, b) {
				buchstabe = b
				break
			}
		}
		if buchstabe == "" {
			break
		}
		vorher := sitz.Punkte
		rec := ruf(sitz.ID, "raten", map[string]any{"buchstabe": buchstabe})
		if rec.Code != http.StatusOK {
			t.Fatalf("raten %s: %d %s", buchstabe, rec.Code, rec.Body.String())
		}
		nach := spiel()
		neu := 0
		for _, s := range nach.Spieler {
			if s.ID == sitz.ID {
				neu = s.Punkte
			}
		}
		if neu < vorher {
			t.Fatalf("Punkte gesunken: %d -> %d", vorher, neu)
		}
		gesamt += neu - vorher
	}
	g = spiel()
	if g.Status != "ende" || !g.Geloest {
		t.Fatalf("Runde endet nicht: status=%s geloest=%v", g.Status, g.Geloest)
	}
	wort := strings.Join(g.Anzeige, "")
	wortlaenge := len([]rune(wort))
	if gesamt != wortlaenge {
		t.Fatalf("Punkte %d != Buchstaben %d", gesamt, wortlaenge)
	}
	if g.Sieger == "" {
		t.Fatal("kein Sieger")
	}
	// Nächste Runde startet mit neuem Wort.
	if r := ruf("b", "naechste", nil); r.Code != http.StatusOK {
		t.Fatalf("naechste: %d", r.Code)
	}
	g = spiel()
	if g.Status != "spiel" || g.Runde != 2 {
		t.Fatalf("Runde 2: %+v", g)
	}
	// Verlieren: falsche Buchstaben bis zum Ende.
	for i := 0; i < 40 && g.Status == "spiel"; i++ {
		g = spiel()
		sitz := g.Spieler[g.Zug]
		b := ""
		for _, k := range "abcdefghijklmnopqrstuvwxyzäöüß" {
			bs := string(k)
			if !enthaelt(g.Buchstaben, bs) && !strings.Contains(geheim(), bs) {
				b = bs
				break
			}
		}
		if b == "" {
			break
		}
		ruf(sitz.ID, "raten", map[string]any{"buchstabe": b})
	}
	g = spiel()
	if g.Status != "ende" || g.Geloest {
		t.Fatalf("Niederlage nicht erkannt: status=%s geloest=%v fehler=%d", g.Status, g.Geloest, g.Fehler)
	}
	// Verlassen entfernt den Platz.
	ruf("a", "reset", nil)
	ruf("a", "create", map[string]any{"variante": "streng"})
	ruf("b", "join", nil)
	if drin, fertig, rest := galgenRemoveSeat(room.Galgen, "c", "Cem"); drin || fertig {
		t.Fatalf("Cem war nicht am Tisch: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	if drin, fertig, rest := galgenRemoveSeat(room.Galgen, "b", "Ben"); !drin || !fertig || rest != 1 {
		t.Fatalf("Ben raus: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	t.Logf("OK: Wortlänge %d, Punkte %d, Sieger gesetzt, Varianten und Verlassen geprüft", wortlaenge, gesamt)
}
