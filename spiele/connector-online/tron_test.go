package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// tronRaum legt einen Raum mit zwei Spielern und Tron-Werkzeug an.
func tronRaum(t *testing.T) (func(id, aktion string, koerper map[string]any) *httptest.ResponseRecorder, func(id string) *TronAnsicht) {
	t.Helper()
	now := time.Now()
	room := &Room{Code: "TG-TRN1", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
		Created: now, LastActive: now,
		Players: map[string]Player{
			"a": {Name: "Anna", Token: "ta", Last: now, LastAction: now},
			"b": {Name: "Ben", Token: "tb", Last: now, LastAction: now},
		},
		Mutes: map[string]time.Time{}, MsgTimes: map[string][]time.Time{}}
	mu.Lock()
	rooms[room.Code] = room
	mu.Unlock()
	t.Cleanup(func() { mu.Lock(); delete(rooms, room.Code); mu.Unlock() })

	ruf := func(id, aktion string, koerper map[string]any) *httptest.ResponseRecorder {
		if koerper == nil {
			koerper = map[string]any{}
		}
		koerper["id"] = id
		rumpf, _ := json.Marshal(koerper)
		req := httptest.NewRequest(http.MethodPost, "/api/tron?room="+room.Code+"&action="+aktion, strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		tronHandler(rec, req)
		return rec
	}
	sicht := func(id string) *TronAnsicht {
		req := httptest.NewRequest(http.MethodGet, "/api/tron?room="+room.Code+"&id="+id, nil)
		rec := httptest.NewRecorder()
		tronHandler(rec, req)
		var antwort struct {
			Game *TronAnsicht `json:"game"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &antwort); err != nil {
			t.Fatalf("Antwort unlesbar: %v (%s)", err, rec.Body.String())
		}
		return antwort.Game
	}
	return ruf, sicht
}

// starteTron spielt create→join→ready→start durch.
func starteTron(t *testing.T, ruf func(id, aktion string, koerper map[string]any) *httptest.ResponseRecorder) {
	t.Helper()
	if rec := ruf("a", "create", nil); rec.Code != http.StatusOK {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	if rec := ruf("b", "join", nil); rec.Code != http.StatusOK {
		t.Fatalf("join: %d %s", rec.Code, rec.Body.String())
	}
	ruf("a", "ready", nil)
	ruf("b", "ready", nil)
	if rec := ruf("a", "start", nil); rec.Code != http.StatusOK {
		t.Fatalf("start: %d %s", rec.Code, rec.Body.String())
	}
}

func TestTronTischUndStart(t *testing.T) {
	ruf, sicht := tronRaum(t)

	ruf("a", "create", nil)
	ruf("b", "join", nil)
	if rec := ruf("a", "start", nil); rec.Code != http.StatusConflict {
		t.Fatalf("start ohne bereit: erwartet 409, war %d", rec.Code)
	}
	ruf("a", "ready", nil)
	ruf("b", "ready", nil)
	if rec := ruf("a", "start", nil); rec.Code != http.StatusOK {
		t.Fatalf("start: %d %s", rec.Code, rec.Body.String())
	}

	g := sicht("a")
	if g == nil || g.Status != "spiel" {
		t.Fatalf("Status nach Start: %+v", g)
	}
	if g.Spalten != tronSpalten || g.Zeilen != tronZeilen || g.Ziel != tronZiel || g.Runde != 1 {
		t.Fatalf("Feld oder Ziel falsch: %+v", g)
	}
	if len(g.Spieler) != 2 || g.Du != 0 {
		t.Fatalf("Spieler/Sitz falsch: %+v", g)
	}
	if sicht("b").Du != 1 {
		t.Fatalf("Sitz von Ben ist nicht 1")
	}
	// Startpositionen wie im lokalen Spiel: links nach rechts, rechts nach links.
	if g.Spieler[0].X != 3 || g.Spieler[0].DX != 1 || g.Spieler[1].DX != -1 {
		t.Fatalf("Startrichtung falsch: %+v", g.Spieler)
	}
	// Raster enthält zwei Startfelder, Lauflängen lesbar.
	if !strings.Contains(g.Raster, "x") || strings.Contains(g.Raster, "0x0") {
		t.Fatalf("Raster unbrauchbar: %q", g.Raster)
	}
}

func TestTronRichtung(t *testing.T) {
	ruf, sicht := tronRaum(t)
	starteTron(t, ruf)

	// Umdrehen ist verboten (Anna fährt nach rechts).
	if rec := ruf("a", "richtung", map[string]any{"dir": "links"}); rec.Code != http.StatusConflict {
		t.Fatalf("Umdrehen: erwartet 409, war %d", rec.Code)
	}
	if rec := ruf("a", "richtung", map[string]any{"dir": "schräg"}); rec.Code != http.StatusBadRequest {
		t.Fatalf("unbekannte Richtung: erwartet 400, war %d", rec.Code)
	}
	// Gültige Kurve wird beim nächsten Feld übernommen.
	if rec := ruf("a", "richtung", map[string]any{"dir": "runter"}); rec.Code != http.StatusOK {
		t.Fatalf("Kurve runter: %d %s", rec.Code, rec.Body.String())
	}
	mu.Lock()
	room := rooms["TG-TRN1"]
	room.Tron.Zuletzt = time.Now().Add(-500 * time.Millisecond)
	tronWeiter(room.Tron)
	mu.Unlock()

	g := sicht("a")
	if g.Spieler[0].DY != 1 || g.Spieler[0].DX != 0 {
		t.Fatalf("Kurve nicht übernommen: %+v", g.Spieler[0])
	}
	if g.Spieler[0].Y <= tronZeilen/2-5 {
		t.Fatalf("Renner ist nicht nach unten gefahren: y=%d", g.Spieler[0].Y)
	}

	// Fremde Kennung darf nicht lenken.
	req := httptest.NewRequest(http.MethodPost, "/api/tron?room=TG-TRN1&action=richtung", strings.NewReader(`{"id":"zzz","dir":"hoch"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	tronHandler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("fremde Kennung: erwartet 403, war %d", rec.Code)
	}
}

func TestTronRundeUndMatch(t *testing.T) {
	ruf, _ := tronRaum(t)
	starteTron(t, ruf)

	mu.Lock()
	game := rooms["TG-TRN1"].Tron
	// Ben fährt an die rechte Wand und stirbt dort.
	game.Spieler[1].X = tronSpalten - 1
	game.Spieler[1].Y = 5
	game.Spieler[1].DX = 1
	game.Spieler[1].DY = 0
	game.Raster = make([]uint8, tronSpalten*tronZeilen)
	tronSchritt(game)
	status, gewinner, sieg := game.Status, game.Gewinner, game.Spieler[0].Sieg
	mu.Unlock()
	if status != "runde" || gewinner != 1 {
		t.Fatalf("Runde nicht beendet: status=%s gewinner=%d", status, gewinner)
	}
	if sieg != 1 {
		t.Fatalf("Rundensieg fehlt: %d", sieg)
	}

	// "weiter" startet die nächste Runde und zählt sie hoch.
	if rec := ruf("b", "weiter", nil); rec.Code != http.StatusOK {
		t.Fatalf("weiter: %d %s", rec.Code, rec.Body.String())
	}
	mu.Lock()
	runde, stand := rooms["TG-TRN1"].Tron.Runde, rooms["TG-TRN1"].Tron.Status
	mu.Unlock()
	if runde != 2 || stand != "spiel" {
		t.Fatalf("nächste Runde fehlt: runde=%d status=%s", runde, stand)
	}

	// Zwei weitere Rundensiege für Anna beenden das Match.
	for i := 0; i < 2; i++ {
		mu.Lock()
		game = rooms["TG-TRN1"].Tron
		game.Status = "spiel"
		game.Spieler[1].X = tronSpalten - 1
		game.Spieler[1].Y = 5
		game.Spieler[1].DX = 1
		game.Spieler[1].DY = 0
		game.Spieler[1].Tot = false
		game.Raster = make([]uint8, tronSpalten*tronZeilen)
		tronSchritt(game)
		mu.Unlock()
	}
	mu.Lock()
	status, sieger := rooms["TG-TRN1"].Tron.Status, rooms["TG-TRN1"].Tron.Sieger
	siegA, siegB := rooms["TG-TRN1"].Tron.Spieler[0].Sieg, rooms["TG-TRN1"].Tron.Spieler[1].Sieg
	mu.Unlock()
	if status != "ende" || sieger != "Anna" {
		t.Fatalf("Matchende fehlt: status=%s sieger=%s siege=%d/%d", status, sieger, siegA, siegB)
	}

	// Neue Partie nach dem Ende.
	if rec := ruf("a", "reset", nil); rec.Code != http.StatusOK {
		t.Fatalf("reset: %d %s", rec.Code, rec.Body.String())
	}
}
