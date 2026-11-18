package main

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// pongRaum legt einen Raum mit zwei Spielern und Pong-Werkzeug an.
func pongRaum(t *testing.T) (*Room, func(id, aktion string, koerper map[string]any) *httptest.ResponseRecorder, func(id string) *PongAnsicht) {
	t.Helper()
	now := time.Now()
	room := &Room{Code: "TG-PNG1", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
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
		req := httptest.NewRequest(http.MethodPost, "/api/pong?room="+room.Code+"&action="+aktion, strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		pongHandler(rec, req)
		return rec
	}
	sicht := func(id string) *PongAnsicht {
		req := httptest.NewRequest(http.MethodGet, "/api/pong?room="+room.Code+"&id="+id, nil)
		rec := httptest.NewRecorder()
		pongHandler(rec, req)
		var antwort struct {
			Game *PongAnsicht `json:"game"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &antwort); err != nil {
			t.Fatalf("Antwort unlesbar: %v (%s)", err, rec.Body.String())
		}
		return antwort.Game
	}
	return room, ruf, sicht
}

func TestPongTischUndStart(t *testing.T) {
	_, ruf, sicht := pongRaum(t)

	if rec := ruf("a", "create", nil); rec.Code != http.StatusOK {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	// Zweiter Spieler kommt über die Spielkarte dazu.
	if rec := ruf("b", "join", nil); rec.Code != http.StatusOK {
		t.Fatalf("join: %d %s", rec.Code, rec.Body.String())
	}
	// Start ohne Bereitschaft geht nicht.
	if rec := ruf("a", "start", nil); rec.Code != http.StatusConflict {
		t.Fatalf("start ohne bereit: erwartet 409, war %d", rec.Code)
	}
	ruf("a", "ready", nil)
	ruf("b", "ready", nil)
	if rec := ruf("b", "start", nil); rec.Code != http.StatusForbidden {
		t.Fatalf("start durch Gast: erwartet 403, war %d", rec.Code)
	}
	if rec := ruf("a", "start", nil); rec.Code != http.StatusOK {
		t.Fatalf("start: %d %s", rec.Code, rec.Body.String())
	}

	g := sicht("a")
	if g == nil || g.Status != "spiel" {
		t.Fatalf("Status nach Start: %+v", g)
	}
	if len(g.Spieler) != 2 || len(g.Schlaeger) != 2 || len(g.Punkte) != 2 {
		t.Fatalf("Ansicht unvollständig: %+v", g)
	}
	if g.Du != 0 {
		t.Fatalf("Sitz von Anna: %d", g.Du)
	}
	if sicht("b").Du != 1 {
		t.Fatalf("Sitz von Ben ist nicht 1")
	}
	if math.Hypot(g.Ball.VX, g.Ball.VY) < pongTempoStart-1 {
		t.Fatalf("Ball steht still: %+v", g.Ball)
	}
}

func TestPongSchlaegerGrenzen(t *testing.T) {
	_, ruf, sicht := pongRaum(t)
	ruf("a", "create", nil)
	ruf("b", "join", nil)
	ruf("a", "ready", nil)
	ruf("b", "ready", nil)
	ruf("a", "start", nil)

	ruf("a", "schlaeger", map[string]any{"x": -500.0})
	if x := sicht("a").Schlaeger[0]; math.Abs(x-pongSchlaegerBreite/2) > 0.01 {
		t.Fatalf("linker Anschlag: %.2f", x)
	}
	ruf("a", "schlaeger", map[string]any{"x": 5000.0})
	if x := sicht("a").Schlaeger[0]; math.Abs(x-(pongBreite-pongSchlaegerBreite/2)) > 0.01 {
		t.Fatalf("rechter Anschlag: %.2f", x)
	}
	// Ben darf Annas Schläger nicht verschieben – er bewegt seinen eigenen.
	ruf("b", "schlaeger", map[string]any{"x": 1000.0})
	if x := sicht("a").Schlaeger[0]; math.Abs(x-(pongBreite-pongSchlaegerBreite/2)) > 0.01 {
		t.Fatalf("fremder Schläger verschoben: %.2f", x)
	}

	// Unbekannte Kennung wird abgewiesen.
	req := httptest.NewRequest(http.MethodPost, "/api/pong?room=TG-PNG1&action=schlaeger", strings.NewReader(`{"id":"zzz","x":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	pongHandler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("fremde Kennung: erwartet 403, war %d", rec.Code)
	}
}

func TestPongPhysik(t *testing.T) {
	spiel := func() *PongGame {
		g := pongNeu("TG-X", "a", []PongSpieler{pongNeuerSpieler("a", "Anna"), pongNeuerSpieler("b", "Ben")})
		g.Status = "spiel"
		g.Pause = 0
		return g
	}

	// Seitenwand: Ball fliegt nach links, prallt ab.
	g := spiel()
	g.BallX = 20
	g.BallY = 300
	g.BallVX = -300
	g.BallVY = 0
	for i := 0; i < 30; i++ {
		pongPhysik(g, pongTakt)
	}
	if g.BallVX <= 0 || g.BallX < pongBallR-0.01 {
		t.Fatalf("Wandabpraller: x=%.1f vx=%.1f", g.BallX, g.BallVX)
	}

	// Schläger unten: Ball kommt senkrecht, Schläger steht genau darunter.
	g = spiel()
	g.BallX = 200
	g.BallY = pongSchlaegerY(0) - pongBallR - 2
	g.BallVX = 0
	g.BallVY = 320
	g.Spieler[0].X = 200
	for i := 0; i < 40 && g.Rallye == 0; i++ {
		pongPhysik(g, pongTakt)
	}
	if g.Rallye != 1 {
		t.Fatalf("kein Treffer am unteren Schläger (rallye=%d, y=%.1f)", g.Rallye, g.BallY)
	}
	if g.BallVY >= 0 {
		t.Fatalf("Ball fliegt nach Treffer nicht nach oben: vy=%.1f", g.BallVY)
	}

	// Daneben: Ball geht unten vorbei, der obere Spieler bekommt den Punkt.
	g = spiel()
	g.BallX = 20
	g.BallY = pongHoehe - 30
	g.BallVX = 0
	g.BallVY = 320
	g.Spieler[0].X = 350
	for i := 0; i < 60 && g.Punkte[1] == 0; i++ {
		pongPhysik(g, pongTakt)
	}
	if g.Punkte[1] != 1 || g.Pause <= 0 {
		t.Fatalf("Punkt fehlt: %v pause=%.2f", g.Punkte, g.Pause)
	}

	// Sieg bei 7 Punkten.
	g = spiel()
	g.Punkte[0] = pongSiegespunkte - 1
	g.BallX = 20
	g.BallY = 0
	g.BallVX = 0
	g.BallVY = -320
	for i := 0; i < 40 && g.Status == "spiel"; i++ {
		pongPhysik(g, pongTakt)
	}
	if g.Status != "ende" || g.Sieger != "Anna" {
		t.Fatalf("Sieg fehlt: status=%s sieger=%s punkte=%v", g.Status, g.Sieger, g.Punkte)
	}
}
