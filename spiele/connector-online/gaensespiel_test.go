package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGaenseZiehen(t *testing.T) {
	faelle := []struct{ pos, schritte, erwartet int }{
		{1, 1, 2},
		{60, 3, 63},
		{60, 5, 61},
		{55, 10, 61},
		{62, 12, 52},
		{0, 9, 9},
	}
	for _, f := range faelle {
		got, _ := gaenseZiehen(f.pos, f.schritte)
		if got != f.erwartet {
			t.Fatalf("%d + %d = %d, erwartet %d", f.pos, f.schritte, got, f.erwartet)
		}
	}
}

func neuesGaenseTest(spieler ...string) *GaenseGame {
	g := &GaenseGame{Status: "spiel", Zug: 0, Verlauf: []string{}}
	for i, name := range spieler {
		g.Spieler = append(g.Spieler, GaenseSpieler{ID: strings.ToLower(name), Name: name, Farbe: i})
	}
	return g
}

func TestGaenseSonderfelder(t *testing.T) {
	// Brücke: 3 + 3 = 6, von dort auf 12.
	g := neuesGaenseTest("Anna", "Ben")
	g.Spieler[0].Pos = 3
	if text := g.gaenseZug(0, 3); text != "" {
		t.Fatalf("Zug: %q", text)
	}
	if g.Spieler[0].Pos != gaenseBrueckeZiel {
		t.Fatalf("Brücke: %d", g.Spieler[0].Pos)
	}
	if !strings.Contains(g.Letzte, "Brücke") {
		t.Fatalf("Brücke nicht erklärt: %q", g.Letzte)
	}

	// Gans: 7 + 2 = 9, dann noch einmal 2 auf 11.
	g = neuesGaenseTest("Anna", "Ben")
	g.Spieler[0].Pos = 7
	if text := g.gaenseZug(0, 2); text != "" {
		t.Fatalf("Zug: %q", text)
	}
	if g.Spieler[0].Pos != 11 {
		t.Fatalf("Gans: %d", g.Spieler[0].Pos)
	}
	if !strings.Contains(g.Letzte, "Gänsefeld") {
		t.Fatalf("Gans nicht erklärt: %q", g.Letzte)
	}

	// Gasthaus: 15 + 4 = 19, eine Runde aussetzen.
	g = neuesGaenseTest("Anna", "Ben")
	g.Spieler[0].Pos = 15
	g.gaenseZug(0, 4)
	if g.Spieler[0].Aussetzen != 1 || !strings.Contains(g.Letzte, "Gasthaus") {
		t.Fatalf("Gasthaus: aussetzen=%d text=%q", g.Spieler[0].Aussetzen, g.Letzte)
	}
	if g.Zug != 1 {
		t.Fatalf("Zug nach dem Gasthaus: %d", g.Zug)
	}
	// Bens Zug – danach wird Anna übersprungen und Ben ist wieder dran.
	g.Spieler[1].Pos = 0
	g.gaenseZug(1, 2)
	if g.Spieler[0].Aussetzen != 0 {
		t.Fatalf("Aussetzen nicht abgebaut: %d", g.Spieler[0].Aussetzen)
	}
	if g.Zug != 1 {
		t.Fatalf("Übersprungen, aber Zug bei %d", g.Zug)
	}

	// Brunnen: 27 + 4 = 31 – warten, bis jemand anders auf 31 landet.
	g = neuesGaenseTest("Anna", "Ben")
	g.Spieler[0].Pos = 27
	g.gaenseZug(0, 4)
	if !g.Spieler[0].ImBrunnen || g.Zug != 1 {
		t.Fatalf("Brunnen: drin=%v zug=%d", g.Spieler[0].ImBrunnen, g.Zug)
	}
	// Ben landet ebenfalls auf 31: Anna ist frei, Ben sitzt fest.
	g.Spieler[1].Pos = 27
	g.gaenseZug(1, 4)
	if g.Spieler[0].ImBrunnen {
		t.Fatal("Anna blieb im Brunnen")
	}
	if !g.Spieler[1].ImBrunnen {
		t.Fatal("Ben sitzt nicht im Brunnen")
	}
	if !strings.Contains(g.Letzte, "frei") || !strings.Contains(g.Letzte, "Brunnen") {
		t.Fatalf("Brunnen nicht erklärt: %q", g.Letzte)
	}

	// Labyrinth: 39 + 3 = 42 → 30.
	g = neuesGaenseTest("Anna", "Ben")
	g.Spieler[0].Pos = 39
	g.gaenseZug(0, 3)
	if g.Spieler[0].Pos != gaenseLabyrinthZiel || !strings.Contains(g.Letzte, "Labyrinth") {
		t.Fatalf("Labyrinth: %d %q", g.Spieler[0].Pos, g.Letzte)
	}

	// Gefängnis: 49 + 3 = 52, drei Runden aussetzen.
	g = neuesGaenseTest("Anna", "Ben")
	g.Spieler[0].Pos = 49
	g.gaenseZug(0, 3)
	if g.Spieler[0].Aussetzen != gaenseGefaengnisRunden {
		t.Fatalf("Gefängnis: %d", g.Spieler[0].Aussetzen)
	}

	// Tod: 55 + 3 = 58 → zurück auf 1.
	g = neuesGaenseTest("Anna", "Ben")
	g.Spieler[0].Pos = 55
	g.gaenseZug(0, 3)
	if g.Spieler[0].Pos != gaenseTodZiel || !strings.Contains(g.Letzte, "Tod") {
		t.Fatalf("Tod: %d %q", g.Spieler[0].Pos, g.Letzte)
	}

	// Über das Ziel: von 54 mit 12 → 66 → 60 (Feld 60 ist keine Gans).
	g = neuesGaenseTest("Anna", "Ben")
	g.Spieler[0].Pos = 54
	g.gaenseZug(0, 12)
	if g.Spieler[0].Pos != 60 {
		t.Fatalf("Rücklauf über das Ziel: %d", g.Spieler[0].Pos)
	}

	// Gans mit Rücklauf: von 42 mit 12 auf 54 (Gans), noch einmal 12 → 66 →
	// 60 zurück.
	g = neuesGaenseTest("Anna", "Ben")
	g.Spieler[0].Pos = 42
	g.gaenseZug(0, 12)
	if g.Spieler[0].Pos != 60 {
		t.Fatalf("Gans 42 + 12 über das Ziel: %d", g.Spieler[0].Pos)
	}
	if !strings.Contains(g.Letzte, "Gänsefeld") || !strings.Contains(g.Letzte, "zurück") {
		t.Fatalf("Gans mit Rücklauf nicht erklärt: %q", g.Letzte)
	}

	// Genau auf 63 gewinnt.
	g = neuesGaenseTest("Anna", "Ben")
	g.Spieler[0].Pos = 60
	g.gaenseZug(0, 3)
	if g.Status != "ende" || g.Sieger != "anna" {
		t.Fatalf("Sieg: status=%s sieger=%q", g.Status, g.Sieger)
	}
}

func TestGaenseRegeln(t *testing.T) {
	now := time.Now()
	room := &Room{Code: "TG-GAN1", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
		Created: now, LastActive: now,
		Players: map[string]Player{
			"a": {Name: "Anna", Token: "ta", Last: now, LastAction: now},
			"b": {Name: "Ben", Token: "tb", Last: now, LastAction: now},
			"c": {Name: "Cem", Token: "tc", Last: now, LastAction: now},
			"d": {Name: "Dana", Token: "td", Last: now, LastAction: now},
			"e": {Name: "Emil", Token: "te", Last: now, LastAction: now},
		},
		Mutes: map[string]time.Time{}, MsgTimes: map[string][]time.Time{}}
	mu.Lock()
	rooms[room.Code] = room
	mu.Unlock()
	defer func() { mu.Lock(); delete(rooms, room.Code); mu.Unlock() }()

	ruf := func(id, aktion string) *httptest.ResponseRecorder {
		rumpf, _ := json.Marshal(map[string]any{"id": id})
		req := httptest.NewRequest(http.MethodPost, "/api/gaensespiel?room="+room.Code+"&action="+aktion, strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		gaenseHandler(rec, req)
		return rec
	}
	sicht := func(id string) *GaenseGame {
		req := httptest.NewRequest(http.MethodGet, "/api/gaensespiel?room="+room.Code+"&id="+id, nil)
		rec := httptest.NewRecorder()
		gaenseHandler(rec, req)
		var antwort struct {
			Game *GaenseGame `json:"game"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &antwort)
		return antwort.Game
	}

	if r := ruf("a", "create"); r.Code != http.StatusOK {
		t.Fatalf("create: %d %s", r.Code, r.Body.String())
	}
	for _, id := range []string{"b", "c", "d"} {
		ruf(id, "join")
	}
	if r := ruf("e", "join"); r.Code != http.StatusConflict {
		t.Fatalf("fünfter Spieler: %d", r.Code)
	}
	for _, id := range []string{"a", "b", "c", "d"} {
		ruf(id, "ready")
	}
	if r := ruf("a", "start"); r.Code != http.StatusOK {
		t.Fatalf("start: %d %s", r.Code, r.Body.String())
	}
	g := sicht("a")
	if g.Status != "spiel" || len(g.Spieler) != 4 {
		t.Fatalf("Start: status=%s spieler=%d", g.Status, len(g.Spieler))
	}
	if r := ruf("b", "wuerfeln"); r.Code != http.StatusConflict {
		t.Fatalf("falscher Spieler würfelt: %d", r.Code)
	}
	// Im Brunnen darf man nicht würfeln, weil man nicht am Zug ist.
	g.Spieler[0].ImBrunnen = true
	g.gaenseNaechster()
	if g.Zug == 0 {
		t.Fatal("Brunnenspieler blieb am Zug")
	}

	zuege := 0
	for versuche := 0; versuche < 6000; versuche++ {
		g = sicht("a")
		if g.Status != "spiel" {
			break
		}
		amZug := g.Spieler[g.Zug]
		if amZug.ImBrunnen || amZug.Aussetzen > 0 {
			t.Fatalf("%s ist am Zug, obwohl blockiert", amZug.Name)
		}
		if r := ruf(amZug.ID, "wuerfeln"); r.Code != http.StatusOK {
			t.Fatalf("würfeln %s: %d %s", amZug.Name, r.Code, r.Body.String())
		}
		zuege++
	}
	g = sicht("a")
	if g.Status != "ende" || g.Sieger == "" {
		t.Fatalf("Partie endet nicht: status=%s, %d Züge", g.Status, zuege)
	}
	for _, s := range g.Spieler {
		if s.Pos < 0 || s.Pos > gaenseFelder {
			t.Fatalf("%s steht auf %d", s.Name, s.Pos)
		}
		if s.ID == g.Sieger && s.Pos != gaenseFelder {
			t.Fatalf("Sieger auf %d", s.Pos)
		}
	}
	if r := ruf("a", "naechste"); r.Code != http.StatusOK {
		t.Fatalf("naechste: %d", r.Code)
	}
	g = sicht("a")
	if g.Runde != 2 || g.Status != "spiel" || g.Spieler[0].Pos != 0 || g.Spieler[0].ImBrunnen {
		t.Fatalf("Runde 2: %d / %s / %d", g.Runde, g.Status, g.Spieler[0].Pos)
	}
	if drin, fertig, rest := gaenseRemoveSeat(room.Gaense, "b", "Ben"); !drin || fertig || rest != 3 {
		t.Fatalf("Ben raus: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	if r := ruf("a", "reset"); r.Code != http.StatusOK {
		t.Fatalf("reset: %d", r.Code)
	}
}
