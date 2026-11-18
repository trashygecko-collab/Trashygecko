package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// wortImGitter sucht ein Wort im Gitter und liefert Anfang und Ende.
func wortImGitter(gitter [][]string, wort string) ([2]int, [2]int, bool) {
	kante := len(gitter)
	zeichen := []rune(wort)
	for _, richtung := range [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}, {0, -1}, {-1, 0}, {-1, -1}, {-1, 1}} {
		for z := 0; z < kante; z++ {
			for s := 0; s < kante; s++ {
				endeZ := z + (len(zeichen)-1)*richtung[0]
				endeS := s + (len(zeichen)-1)*richtung[1]
				if endeZ < 0 || endeS < 0 || endeZ >= kante || endeS >= kante {
					continue
				}
				passt := true
				for k, b := range zeichen {
					if gitter[z+k*richtung[0]][s+k*richtung[1]] != string(b) {
						passt = false
						break
					}
				}
				if passt {
					return [2]int{z, s}, [2]int{endeZ, endeS}, true
				}
			}
		}
	}
	return [2]int{}, [2]int{}, false
}

func TestWortsucheRegeln(t *testing.T) {
	now := time.Now()
	room := &Room{Code: "TG-WOR1", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
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
		req := httptest.NewRequest(http.MethodPost, "/api/wortsuche?room="+room.Code+"&action="+aktion, strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		wortsucheHandler(rec, req)
		return rec
	}
	sicht := func(id string) *WortsucheGame {
		req := httptest.NewRequest(http.MethodGet, "/api/wortsuche?room="+room.Code+"&id="+id, nil)
		rec := httptest.NewRecorder()
		wortsucheHandler(rec, req)
		var antwort struct {
			Game *WortsucheGame `json:"game"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &antwort)
		return antwort.Game
	}

	if r := ruf("a", "create", nil); r.Code != http.StatusOK {
		t.Fatalf("create: %d %s", r.Code, r.Body.String())
	}
	ruf("b", "join", nil)
	ruf("c", "join", nil)
	if r := ruf("a", "start", nil); r.Code != http.StatusConflict {
		t.Fatalf("Start ohne Bereitschaft: %d", r.Code)
	}
	for _, id := range []string{"a", "b", "c"} {
		ruf(id, "ready", nil)
	}
	if r := ruf("a", "start", nil); r.Code != http.StatusOK {
		t.Fatalf("start: %d %s", r.Code, r.Body.String())
	}

	g := sicht("a")
	if g.Status != "spiel" || g.Runde != 1 || g.Kante != 10 {
		t.Fatalf("Start: status=%s runde=%d kante=%d", g.Status, g.Runde, g.Kante)
	}
	if len(g.Gitter) != 10 || len(g.Gitter[0]) != 10 {
		t.Fatalf("Gitter: %dx%d", len(g.Gitter), len(g.Gitter[0]))
	}
	if len(g.Worte) != wortsucheWorte {
		t.Fatalf("Wörter: %d", len(g.Worte))
	}
	if g.Rest <= 0 || g.Rest > wortsucheDauer {
		t.Fatalf("Restzeit beim Start: %d", g.Rest)
	}
	// Jedes Wort muss im Gitter liegen, sonst wäre die Runde unlösbar.
	for _, wort := range g.Worte {
		if _, _, ok := wortImGitter(g.Gitter, wort); !ok {
			t.Fatalf("Wort %s liegt nicht im Gitter", wort)
		}
	}

	wort := g.Worte[0]
	von, bis, _ := wortImGitter(g.Gitter, wort)

	// Krumme Auswahl, falsches Wort, fremdes Wort: alles abgewiesen.
	if r := ruf("a", "finden", map[string]any{"wort": wort, "von": []int{von[0], von[1]}, "bis": []int{(von[0] + 2) % 10, (von[1] + 1) % 10}}); r.Code != http.StatusConflict {
		t.Fatalf("krumme Auswahl: %d", r.Code)
	}
	if r := ruf("a", "finden", map[string]any{"wort": "XYLOPHON", "von": []int{von[0], von[1]}, "bis": []int{bis[0], bis[1]}}); r.Code != http.StatusConflict {
		t.Fatalf("Wort nicht auf der Liste: %d", r.Code)
	}
	if r := ruf("a", "finden", map[string]any{"wort": wort, "von": []int{bis[0], bis[1]}, "bis": []int{von[0], von[1] - 1}}); r.Code == http.StatusOK {
		t.Fatal("falsche Länge wurde angenommen")
	}
	if sicht("a").Spieler[0].Fehlversuch < 3 {
		t.Fatalf("Fehlversuche nicht gezählt: %d", sicht("a").Spieler[0].Fehlversuch)
	}

	// Richtige Auswahl – auch umgekehrt herum (Ende zuerst) zählt.
	if r := ruf("b", "finden", map[string]any{"wort": wort, "von": []int{bis[0], bis[1]}, "bis": []int{von[0], von[1]}}); r.Code != http.StatusOK {
		t.Fatalf("richtiges Wort (umgekehrt): %d %s", r.Code, r.Body.String())
	}
	g = sicht("a")
	if len(g.Funde) != 1 || g.Funde[0].Wort != wort || g.Funde[0].VonID != "b" {
		t.Fatalf("Fund nicht vermerkt: %+v", g.Funde)
	}
	erwartetPunkte := len([]rune(wort)) * wortsuchePunktFak
	if g.Spieler[1].Punkte != erwartetPunkte || g.Spieler[1].Treffer != 1 {
		t.Fatalf("Punkte: %d (erwartet %d), Treffer %d", g.Spieler[1].Punkte, erwartetPunkte, g.Spieler[1].Treffer)
	}
	if len(g.Funde[0].Zellen) != len([]rune(wort)) {
		t.Fatalf("Zellen im Fund: %d", len(g.Funde[0].Zellen))
	}
	// Dasselbe Wort noch einmal geht nicht.
	if r := ruf("a", "finden", map[string]any{"wort": wort, "von": []int{von[0], von[1]}, "bis": []int{bis[0], bis[1]}}); r.Code != http.StatusConflict {
		t.Fatalf("doppelter Fund: %d", r.Code)
	}

	// Restliche Wörter reihum finden, bis die Runde endet.
	reihe := []string{"a", "c"}
	for i, w := range g.Worte[1:] {
		v, b, ok := wortImGitter(g.Gitter, w)
		if !ok {
			t.Fatalf("Wort %s fehlt im Gitter", w)
		}
		wer := reihe[i%len(reihe)]
		if r := ruf(wer, "finden", map[string]any{"wort": w, "von": []int{v[0], v[1]}, "bis": []int{b[0], b[1]}}); r.Code != http.StatusOK {
			t.Fatalf("finden %s: %d %s", w, r.Code, r.Body.String())
		}
	}
	g = sicht("a")
	if g.Status != "ende" || len(g.Funde) != len(g.Worte) {
		t.Fatalf("Runde endet nicht: status=%s funde=%d/%d", g.Status, len(g.Funde), len(g.Worte))
	}
	bester := ""
	best := -1
	for _, s := range g.Spieler {
		if s.Punkte > best {
			best, bester = s.Punkte, s.ID
		}
	}
	if g.Sieger != bester || best <= 0 {
		t.Fatalf("Sieger: %q (erwartet %q, %d Punkte)", g.Sieger, bester, best)
	}

	// Neue Runde: frisches Gitter, Punkte zurückgesetzt.
	if r := ruf("a", "naechste", nil); r.Code != http.StatusOK {
		t.Fatalf("naechste: %d %s", r.Code, r.Body.String())
	}
	g = sicht("a")
	if g.Status != "spiel" || g.Runde != 2 || len(g.Funde) != 0 {
		t.Fatalf("Runde 2: status=%s runde=%d funde=%d", g.Status, g.Runde, len(g.Funde))
	}
	for _, s := range g.Spieler {
		if s.Punkte != 0 || s.Treffer != 0 || s.Fehlversuch != 0 {
			t.Fatalf("Werte nicht zurückgesetzt: %+v", s)
		}
	}

	// Zeitablauf: der Start wird zurückdatiert, dann endet die Runde.
	mu.Lock()
	rooms[room.Code].Wortsuche.start = time.Now().Add(-wortsucheDauer*time.Second - 5)
	mu.Unlock()
	g = sicht("a")
	if g.Status != "ende" || g.Rest != 0 {
		t.Fatalf("Zeitablauf: status=%s rest=%d", g.Status, g.Rest)
	}

	// Verlassen nimmt den Platz, der letzte beendet die Runde.
	if drin, fertig, rest := wortsucheRemoveSeat(room.Wortsuche, "c", "Cem"); !drin || fertig || rest != 2 {
		t.Fatalf("Cem raus: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	if drin, fertig, rest := wortsucheRemoveSeat(room.Wortsuche, "b", "Ben"); !drin || !fertig || rest != 1 {
		t.Fatalf("Ben raus: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	if r := ruf("a", "reset", nil); r.Code != http.StatusOK {
		t.Fatalf("reset: %d", r.Code)
	}
	if sicht("a") != nil {
		t.Fatal("Runde blieb stehen")
	}
	t.Logf("OK: %d Wörter, Punkte je Buchstabe %d, Zeit %ds", wortsucheWorte, wortsuchePunktFak, wortsucheDauer)
}

// TestWortsucheTisch: anbieten, beitreten, umbenennen – wie die anderen Tischspiele.
func TestWortsucheTisch(t *testing.T) {
	now := time.Now()
	room := &Room{Code: "TG-WOR2", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
		Created: now, LastActive: now,
		Players: map[string]Player{
			"a": {Name: "Anna", Token: "ta", Last: now, LastAction: now},
			"b": {Name: "Ben", Token: "tb", Last: now, LastAction: now},
		},
		Mutes: map[string]time.Time{}, MsgTimes: map[string][]time.Time{}}
	mu.Lock()
	rooms[room.Code] = room
	mu.Unlock()
	defer func() { mu.Lock(); delete(rooms, room.Code); mu.Unlock() }()

	if !tischspiel["wortsuche"] {
		t.Fatal("Wortsuche fehlt bei den Tischspielen")
	}
	if nr, text := spielAufruf(spielApi["wortsuche"], room.Code, "create", map[string]any{"id": "a"}); nr >= 300 {
		t.Fatalf("create über den Connector: %d %s", nr, text)
	}
	if nr, text := spielAufruf(spielApi["wortsuche"], room.Code, "join", map[string]any{"id": "b"}); nr >= 300 {
		t.Fatalf("join über den Connector: %d %s", nr, text)
	}
	mu.Lock()
	rooms[room.Code].Players["b"] = Player{Name: "Berta", Token: "tb", Last: now, LastAction: now}
	rooms[room.Code].spielerUmbenennen("b", "Berta")
	spieler := len(rooms[room.Code].Wortsuche.Spieler)
	name := rooms[room.Code].Wortsuche.Spieler[1].Name
	mu.Unlock()
	if spieler != 2 || name != "Berta" {
		t.Fatalf("Tisch: %d Spieler, Name %q", spieler, name)
	}
}
