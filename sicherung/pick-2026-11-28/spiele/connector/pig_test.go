package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPigZuege(t *testing.T) {
	g := &PigGame{
		Status: "spiel", Zug: 0,
		Spieler: []PigSpieler{
			{ID: "a", Name: "Anna"},
			{ID: "b", Name: "Ben"},
		},
		Verlauf: []string{},
	}
	// 6, 5, 2 sammeln – dann sichern.
	for _, wurf := range []int{6, 5, 2} {
		if text := g.pigWurf(0, wurf); text != "" {
			t.Fatalf("Wurf %d: %q", wurf, text)
		}
	}
	if g.Spieler[0].Runde != 13 {
		t.Fatalf("Rundensumme: %d", g.Spieler[0].Runde)
	}
	if g.Spieler[0].Punkte != 0 {
		t.Fatalf("Punkte vor dem Sichern: %d", g.Spieler[0].Punkte)
	}
	if text := g.pigSichern(0); text != "" {
		t.Fatalf("Sichern: %q", text)
	}
	if g.Spieler[0].Punkte != 13 || g.Spieler[0].Runde != 0 {
		t.Fatalf("nach dem Sichern: %d / %d", g.Spieler[0].Punkte, g.Spieler[0].Runde)
	}
	if g.Zug != 1 {
		t.Fatalf("Zug nicht weitergegeben: %d", g.Zug)
	}
	// Eine 1 löscht die Rundensumme und beendet den Zug.
	if text := g.pigWurf(1, 4); text != "" {
		t.Fatalf("Wurf: %q", text)
	}
	if text := g.pigWurf(1, 1); text != "" {
		t.Fatalf("Wurf 1: %q", text)
	}
	if g.Spieler[1].Runde != 0 || g.Spieler[1].Punkte != 0 {
		t.Fatalf("nach der 1: %d / %d", g.Spieler[1].Runde, g.Spieler[1].Punkte)
	}
	if g.Zug != 0 {
		t.Fatalf("nach der 1 nicht weitergegeben: %d", g.Zug)
	}
	// Ohne Punkte im Zug lässt sich nicht sichern.
	if text := g.pigSichern(0); text != "In diesem Zug ist noch nichts zu sichern." {
		t.Fatalf("Sichern ohne Punkte: %q", text)
	}
	// Punkte müssen gesichert werden: 99 + 6 ergibt 105 im Zug, aber noch
	// kein Sieg – erst das Sichern gewinnt.
	g.Spieler[0].Punkte = 99
	if text := g.pigWurf(0, 6); text != "" {
		t.Fatalf("Wurf: %q", text)
	}
	if g.Status != "spiel" || g.Spieler[0].Punkte != 99 {
		t.Fatalf("Sieg ohne Sichern: status=%s punkte=%d", g.Status, g.Spieler[0].Punkte)
	}
	if text := g.pigSichern(0); text != "" {
		t.Fatalf("Sichern: %q", text)
	}
	if g.Status != "ende" || g.Sieger != "a" || g.Spieler[0].Punkte != 105 {
		t.Fatalf("Sieg: status=%s sieger=%s punkte=%d", g.Status, g.Sieger, g.Spieler[0].Punkte)
	}
}

func TestPigRegeln(t *testing.T) {
	now := time.Now()
	room := &Room{Code: "TG-PIG1", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
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

	ruf := func(id, aktion string) *httptest.ResponseRecorder {
		rumpf, _ := json.Marshal(map[string]any{"id": id})
		req := httptest.NewRequest(http.MethodPost, mitToken(room.Code, id, "/api/pig?room="+room.Code+"&action="+aktion), strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		withIdentity(http.HandlerFunc(pigHandler)).ServeHTTP(rec, req)
		return rec
	}
	sicht := func(id string) *PigGame {
		req := httptest.NewRequest(http.MethodGet, mitToken(room.Code, id, "/api/pig?room="+room.Code+"&id="+id), nil)
		rec := httptest.NewRecorder()
		withIdentity(http.HandlerFunc(pigHandler)).ServeHTTP(rec, req)
		var antwort struct {
			Game *PigGame `json:"game"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &antwort)
		return antwort.Game
	}

	if r := ruf("a", "create"); r.Code != http.StatusOK {
		t.Fatalf("create: %d %s", r.Code, r.Body.String())
	}
	ruf("b", "join")
	ruf("c", "join")
	for _, id := range []string{"a", "b", "c"} {
		ruf(id, "ready")
	}
	if r := ruf("a", "start"); r.Code != http.StatusOK {
		t.Fatalf("start: %d %s", r.Code, r.Body.String())
	}
	g := sicht("a")
	if g.Status != "spiel" || len(g.Spieler) != 3 || g.Zug != 0 {
		t.Fatalf("Start: status=%s spieler=%d zug=%d", g.Status, len(g.Spieler), g.Zug)
	}
	if r := ruf("b", "wuerfeln"); r.Code != http.StatusConflict {
		t.Fatalf("falscher Spieler würfelt: %d", r.Code)
	}
	if r := ruf("a", "sichern"); r.Code != http.StatusConflict {
		t.Fatalf("Sichern ohne Punkte: %d", r.Code)
	}

	// Partie zu Ende spielen: ab 20 Punkten im Zug wird gesichert.
	zuege := 0
	for versuche := 0; versuche < 5000; versuche++ {
		g = sicht("a")
		if g.Status != "spiel" {
			break
		}
		amZug := g.Spieler[g.Zug]
		if amZug.Runde >= 20 {
			if r := ruf(amZug.ID, "sichern"); r.Code != http.StatusOK {
				t.Fatalf("sichern: %d %s", r.Code, r.Body.String())
			}
		} else {
			if r := ruf(amZug.ID, "wuerfeln"); r.Code != http.StatusOK {
				t.Fatalf("würfeln: %d %s", r.Code, r.Body.String())
			}
		}
		zuege++
	}
	g = sicht("a")
	if g.Status != "ende" || g.Sieger == "" {
		t.Fatalf("Partie endet nicht: status=%s (%.0f Züge)", g.Status, float64(zuege))
	}
	for _, s := range g.Spieler {
		if s.Punkte > 200 {
			t.Fatalf("Punkte unplausibel: %s %d", s.Name, s.Punkte)
		}
		if s.ID == g.Sieger && s.Punkte < pigZiel {
			t.Fatalf("Sieger unter 100: %d", s.Punkte)
		}
		if s.ID != g.Sieger && s.Punkte >= pigZiel {
			t.Fatalf("Verlierer über 100: %s %d", s.Name, s.Punkte)
		}
	}
	if r := ruf("a", "naechste"); r.Code != http.StatusOK {
		t.Fatalf("naechste: %d", r.Code)
	}
	g = sicht("a")
	if g.Runde != 2 || g.Status != "spiel" || g.Spieler[0].Punkte != 0 {
		t.Fatalf("Runde 2: %d / %s / %d", g.Runde, g.Status, g.Spieler[0].Punkte)
	}
	if drin, fertig, rest := pigRemoveSeat(room.Pig, "c", "Cem"); !drin || fertig || rest != 2 {
		t.Fatalf("Cem raus: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	if drin, fertig, rest := pigRemoveSeat(room.Pig, "b", "Ben"); !drin || !fertig || rest != 1 {
		t.Fatalf("Ben raus: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	if r := ruf("a", "reset"); r.Code != http.StatusOK {
		t.Fatalf("reset: %d", r.Code)
	}
}
