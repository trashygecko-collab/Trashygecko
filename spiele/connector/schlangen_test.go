package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSchlangenBrett(t *testing.T) {
	for lauf := 0; lauf < 50; lauf++ {
		leitern, schlangen := schlangenBrett()
		if len(leitern) != schlangenLeitern || len(schlangen) != schlangenSchlangen {
			t.Fatalf("Brett: %d Leitern, %d Schlangen", len(leitern), len(schlangen))
		}
		belegt := map[int]bool{}
		for _, paar := range leitern {
			if paar[1] <= paar[0] {
				t.Fatalf("Leiter zeigt nicht nach oben: %v", paar)
			}
			if paar[0] < 2 || paar[1] > 99 || belegt[paar[0]] || belegt[paar[1]] {
				t.Fatalf("Leiter doppelt oder am Rand: %v", paar)
			}
			belegt[paar[0]], belegt[paar[1]] = true, true
		}
		for _, paar := range schlangen {
			if paar[1] >= paar[0] {
				t.Fatalf("Schlange zeigt nicht nach unten: %v", paar)
			}
			if paar[1] < 2 || paar[0] > 99 || belegt[paar[0]] || belegt[paar[1]] {
				t.Fatalf("Schlange doppelt oder am Rand: %v", paar)
			}
			belegt[paar[0]], belegt[paar[1]] = true, true
		}
	}
}

func TestSchlangenZuege(t *testing.T) {
	g := &SchlangenGame{
		Status: "spiel", Zug: 0,
		Leitern:   [][2]int{{3, 22}},
		Schlangen: [][2]int{{20, 4}},
		Spieler: []SchlangenSpieler{
			{ID: "a", Name: "Anna"},
			{ID: "b", Name: "Ben"},
		},
		Verlauf: []string{},
	}
	// Leiter: 1 + 2 = 3 → 22.
	g.Spieler[0].Pos = 1
	if text := g.schlangenZug(0, 2); text != "" {
		t.Fatalf("Zug: %q", text)
	}
	if g.Spieler[0].Pos != 22 {
		t.Fatalf("Leiter nicht genutzt: %d", g.Spieler[0].Pos)
	}
	if !strings.Contains(g.Letzte, "Leiter") {
		t.Fatalf("Leiter nicht erklärt: %q", g.Letzte)
	}
	if g.Zug != 1 {
		t.Fatalf("Zug nicht weitergegeben: %d", g.Zug)
	}
	// Zu weit: von 98 mit 5 über 100 – die Figur bleibt stehen.
	g.Spieler[1].Pos = 98
	g.Zug = 1
	if text := g.schlangenZug(1, 5); text != "" {
		t.Fatalf("Zug: %q", text)
	}
	if g.Spieler[1].Pos != 98 {
		t.Fatalf("Figur lief über das Ziel: %d", g.Spieler[1].Pos)
	}
	if !strings.Contains(g.Letzte, "zu weit") {
		t.Fatalf("Überlauf nicht erklärt: %q", g.Letzte)
	}
	// Exakt auf 100 gewinnt.
	g.Zug = 1
	if text := g.schlangenZug(1, 2); text != "" {
		t.Fatalf("Zug: %q", text)
	}
	if g.Status != "ende" || g.Sieger != "b" {
		t.Fatalf("Sieg: status=%s sieger=%s", g.Status, g.Sieger)
	}
	// Falscher Spieler darf nicht würfeln.
	if text := g.schlangenZug(0, 3); text != "Gerade läuft keine Runde." {
		t.Fatalf("Zug nach dem Ende: %q", text)
	}
}

func TestSchlangenRegeln(t *testing.T) {
	now := time.Now()
	room := &Room{Code: "TG-SCH1", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
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
		req := httptest.NewRequest(http.MethodPost, "/api/schlangen?room="+room.Code+"&action="+aktion, strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		schlangenHandler(rec, req)
		return rec
	}
	sicht := func(id string) *SchlangenGame {
		req := httptest.NewRequest(http.MethodGet, "/api/schlangen?room="+room.Code+"&id="+id, nil)
		rec := httptest.NewRecorder()
		schlangenHandler(rec, req)
		var antwort struct {
			Game *SchlangenGame `json:"game"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &antwort)
		return antwort.Game
	}

	if r := ruf("a", "create"); r.Code != http.StatusOK {
		t.Fatalf("create: %d %s", r.Code, r.Body.String())
	}
	for _, id := range []string{"b", "c", "d"} {
		if r := ruf(id, "join"); r.Code != http.StatusOK {
			t.Fatalf("join %s: %d", id, r.Code)
		}
	}
	if r := ruf("e", "join"); r.Code != http.StatusConflict {
		t.Fatalf("fünfter Spieler: %d", r.Code)
	}
	if r := ruf("a", "start"); r.Code != http.StatusConflict {
		t.Fatalf("Start ohne Bereitschaft: %d", r.Code)
	}
	for _, id := range []string{"a", "b", "c", "d"} {
		ruf(id, "ready")
	}
	if r := ruf("a", "start"); r.Code != http.StatusOK {
		t.Fatalf("start: %d %s", r.Code, r.Body.String())
	}
	g := sicht("a")
	if g.Status != "spiel" || len(g.Spieler) != 4 || len(g.Leitern) != 6 || len(g.Schlangen) != 6 {
		t.Fatalf("Start: status=%s spieler=%d leitern=%d schlangen=%d", g.Status, len(g.Spieler), len(g.Leitern), len(g.Schlangen))
	}
	if g.Spieler[0].Pos != 0 || g.Spieler[3].Pos != 0 {
		t.Fatalf("Startpositionen: %d / %d", g.Spieler[0].Pos, g.Spieler[3].Pos)
	}
	// Nur der Spieler am Zug darf würfeln.
	if r := ruf("b", "wuerfeln"); r.Code != http.StatusConflict {
		t.Fatalf("falscher Spieler würfelt: %d", r.Code)
	}

	// Partie durchspielen und dabei die Regeln prüfen.
	zuege := 0
	versuche := 0
	for ; versuche < 4000; versuche++ {
		g = sicht("a")
		if g.Status != "spiel" {
			break
		}
		amZug := g.Spieler[g.Zug].ID
		vorher := sicht(amZug).Spieler[g.Zug].Pos
		if r := ruf(amZug, "wuerfeln"); r.Code != http.StatusOK {
			t.Fatalf("würfeln %s: %d %s", amZug, r.Code, r.Body.String())
		}
		after := sicht(amZug)
		danach := vorher
		for _, s := range after.Spieler {
			if s.ID == amZug {
				danach = s.Pos
			}
		}
		wurf := after.Wurf
		if wurf < 1 || wurf > 6 {
			t.Fatalf("Würfel: %d", wurf)
		}
		if danach != 0 && danach < vorher && danach != vorher {
			// Nur Schlangen dürfen zurückführen – und dann muss eine Schlange
			// genau auf diesem Feld liegen.
			to, ok := schlangenZiel(after.Schlangen, vorher+wurf)
			if !ok || to != danach {
				t.Fatalf("Rücksprung ohne Schlange: %d → %d (Wurf %d)", vorher, danach, wurf)
			}
		}
		if danach > schlangenFelder {
			t.Fatalf("Figur über 100: %d", danach)
		}
		if after.Status == "spiel" {
			if after.Spieler[after.Zug].ID == amZug {
				t.Fatalf("Spieler war zweimal hintereinander dran")
			}
		} else if after.Sieger != amZug {
			t.Fatalf("Sieger %q, erwartet %q", after.Sieger, amZug)
		}
		zuege++
	}
	g = sicht("a")
	if g.Status != "ende" || g.Sieger == "" {
		t.Fatalf("Partie endet nicht: status=%s, %d Züge", g.Status, zuege)
	}
	gewinner := g.Spieler[0]
	for _, s := range g.Spieler {
		if s.ID == g.Sieger {
			gewinner = s
		}
	}
	if gewinner.Pos != schlangenFelder {
		t.Fatalf("Sieger steht auf %d", gewinner.Pos)
	}
	if r := ruf("a", "naechste"); r.Code != http.StatusOK {
		t.Fatalf("naechste: %d", r.Code)
	}
	g = sicht("a")
	if g.Runde != 2 || g.Status != "spiel" || g.Spieler[0].Pos != 0 {
		t.Fatalf("Runde 2: runde=%d status=%s pos=%d", g.Runde, g.Status, g.Spieler[0].Pos)
	}
	if drin, fertig, rest := schlangenRemoveSeat(room.Schlangen, "b", "Ben"); !drin || fertig || rest != 3 {
		t.Fatalf("Ben raus: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	if r := ruf("a", "reset"); r.Code != http.StatusOK {
		t.Fatalf("reset: %d", r.Code)
	}
	if sicht("a") != nil {
		t.Fatal("Runde blieb stehen")
	}
}
