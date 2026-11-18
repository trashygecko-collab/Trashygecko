package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func kannFuerTest(g *DominoGame, hand [][2]int) bool {
	for _, k := range hand {
		if dominoPasstTest(g, k) {
			return true
		}
	}
	return false
}

func dominoPasstTest(g *DominoGame, stein [2]int) bool {
	if len(g.Kette) == 0 {
		return true
	}
	for _, e := range g.Enden {
		if stein[0] == e || stein[1] == e {
			return true
		}
	}
	return false
}

// TestDominoRegeln spielt eine ganze Runde über die Schnittstelle: Ausgeben,
// Beitreten, Bereit, Start, Legen/Ziehen/Passen bis zum Ende, Punkte, neue
// Runde, Zeit- und Sonderfälle.
func TestDominoRegeln(t *testing.T) {
	now := time.Now()
	room := &Room{Code: "TG-DOM1", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
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
		req := httptest.NewRequest(http.MethodPost, "/api/domino?room="+room.Code+"&action="+aktion, strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		dominoHandler(rec, req)
		return rec
	}
	sicht := func(id string) *DominoGame {
		req := httptest.NewRequest(http.MethodGet, "/api/domino?room="+room.Code+"&id="+id, nil)
		rec := httptest.NewRecorder()
		dominoHandler(rec, req)
		var antwort struct {
			Game *DominoGame `json:"game"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &antwort)
		return antwort.Game
	}

	if r := ruf("a", "create", nil); r.Code != http.StatusOK {
		t.Fatalf("create: %d %s", r.Code, r.Body.String())
	}
	if r := ruf("b", "join", nil); r.Code != http.StatusOK {
		t.Fatalf("join: %d %s", r.Code, r.Body.String())
	}
	// Der Tisch ist voll: ein dritter Spieler darf nicht mehr mit.
	if r := ruf("c", "join", nil); r.Code != http.StatusConflict {
		t.Fatalf("dritter Spieler: %d", r.Code)
	}
	if r := ruf("a", "start", nil); r.Code != http.StatusConflict {
		t.Fatalf("Start ohne Bereitschaft: %d", r.Code)
	}
	ruf("a", "ready", nil)
	ruf("b", "ready", nil)
	if r := ruf("a", "start", nil); r.Code != http.StatusOK {
		t.Fatalf("start: %d %s", r.Code, r.Body.String())
	}

	g := sicht("a")
	if g.Status != "spiel" || g.Runde != 1 {
		t.Fatalf("Start: status=%s runde=%d", g.Status, g.Runde)
	}
	if len(g.Spieler) != 2 || len(g.Kette) != 0 || g.Talon != 14 {
		t.Fatalf("Tisch: %d Spieler, %d Steine in der Kette, %d im Vorrat", len(g.Spieler), len(g.Kette), g.Talon)
	}
	if len(g.Spieler[0].Hand) != 7 || g.Spieler[1].Anzahl != 7 {
		t.Fatalf("Hände: %d eigene, %d fremde", len(g.Spieler[0].Hand), g.Spieler[1].Anzahl)
	}
	// Fremde Hände bleiben verdeckt, nur ihre Anzahl steht da.
	if len(sicht("a").Spieler[1].Hand) != 0 {
		t.Fatalf("fremde Hand ist sichtbar: %d Steine", len(sicht("a").Spieler[1].Hand))
	}
	if sicht("a").Spieler[1].Anzahl != 7 {
		t.Fatalf("Anzahl fremder Steine: %d", sicht("a").Spieler[1].Anzahl)
	}
	if len(sicht("b").Spieler[0].Hand) != 0 {
		t.Fatalf("Annas Hand ist für Ben sichtbar")
	}

	// Alle 28 Steine sind einmal im Spiel: Kette + Hände + Vorrat.
	alle := map[[2]int]int{}
	add := func(s [2]int) {
		if s[0] > s[1] {
			s[0], s[1] = s[1], s[0]
		}
		alle[s]++
	}
	for _, s := range g.Kette {
		add(s)
	}
	for _, s := range sicht("a").Spieler[0].Hand {
		add(s)
	}
	for _, s := range sicht("b").Spieler[1].Hand {
		add(s)
	}
	if len(alle)+g.Talon != 28 {
		t.Fatalf("Steine: %d verschiedene + %d Vorrat", len(alle), g.Talon)
	}
	for s, n := range alle {
		if n != 1 {
			t.Fatalf("Stein %v kommt %dx vor", s, n)
		}
	}

	// Falscher Spieler, fremder Stein, unpassende Seite: alles abgewiesen.
	wer := g.Spieler[g.Zug].ID
	anderer := ""
	for _, s := range g.Spieler {
		if s.ID != wer {
			anderer = s.ID
		}
	}
	erster := sicht(wer).Spieler[g.Zug].Hand[0]
	if r := ruf(anderer, "legen", map[string]any{"stein": []int{erster[0], erster[1]}}); r.Code != http.StatusConflict {
		t.Fatalf("falscher Spieler durfte legen: %d", r.Code)
	}
	// Ziehen ist nur erlaubt, wenn wirklich kein Stein passt.
	if eigen := sicht(wer).Spieler[g.Zug].Hand; kannFuerTest(g, eigen) {
		if r := ruf(wer, "ziehen", nil); r.Code != http.StatusConflict {
			t.Fatalf("Ziehen mit passendem Stein: %d", r.Code)
		}
	}
	g = sicht("a")

	// Ganze Runde durchspielen: legen, sonst ziehen, sonst passen.
	zuege := 0
	for i := 0; i < 400; i++ {
		g = sicht("a")
		if g.Status != "spiel" {
			break
		}
		amZug := g.Spieler[g.Zug].ID
		eigen := sicht(amZug)
		var stein [2]int
		gefunden := false
		for _, k := range eigen.Spieler[eigen.Zug].Hand {
			if dominoPasstTest(eigen, k) {
				stein, gefunden = k, true
				break
			}
		}
		var r *httptest.ResponseRecorder
		switch {
		case gefunden:
			r = ruf(amZug, "legen", map[string]any{"stein": []int{stein[0], stein[1]}})
		case eigen.Talon > 0:
			r = ruf(amZug, "ziehen", nil)
		default:
			r = ruf(amZug, "passen", nil)
		}
		if r.Code != http.StatusOK {
			t.Fatalf("Zug %d (%s): %d %s", i, amZug, r.Code, r.Body.String())
		}
		zuege++
	}
	g = sicht("a")
	if g.Status != "ende" {
		t.Fatalf("Runde endet nicht: status=%s, %d Züge", g.Status, zuege)
	}
	if len(g.Kette) == 0 {
		t.Fatal("keine Steine in der Kette")
	}
	// Punkte: Wer gewinnt, bekommt genau die Augen des Gegners.
	voll := room.Domino
	if voll.Sieger != "" {
		gewonnen, punkte := -1, 0
		for i := range voll.Spieler {
			if voll.Spieler[i].ID == voll.Sieger {
				gewonnen = i
			} else {
				punkte += voll.dominoAugen(i)
			}
		}
		if gewonnen >= 0 && voll.Spieler[gewonnen].Punkte != punkte {
			t.Fatalf("Punkte %d, erwartet %d", voll.Spieler[gewonnen].Punkte, punkte)
		}
	}
	// Kette + beide Hände + Vorrat ergeben wieder 28 Steine.
	if len(voll.Kette)+len(voll.Spieler[0].Hand)+len(voll.Spieler[1].Hand)+len(voll.vorrat) != 28 {
		t.Fatalf("Steine verloren: %d + %d + %d + %d", len(voll.Kette), len(voll.Spieler[0].Hand),
			len(voll.Spieler[1].Hand), len(voll.vorrat))
	}
	// Die Kette ist lückenlos: jeder Stein passt an seinen Nachbarn.
	for i := 1; i < len(g.Kette); i++ {
		if g.Kette[i-1][1] != g.Kette[i][0] {
			t.Fatalf("Kette reißt bei %d: %v | %v", i, g.Kette[i-1], g.Kette[i])
		}
	}
	if len(g.Kette) > 0 && (g.Enden[0] != g.Kette[0][0] || g.Enden[1] != g.Kette[len(g.Kette)-1][1]) {
		t.Fatalf("Enden %v passen nicht zur Kette", g.Enden)
	}
	// Punkte: der Sieger bekommt genau die Augen des Gegners.
	for i := range g.Spieler {
		if g.Spieler[i].ID == g.Sieger {
			if g.Spieler[i].Runden != 1 || g.Spieler[i].Punkte < 0 {
				t.Fatalf("Sieger: %+v", g.Spieler[i])
			}
			if len(g.Spieler[i].Hand) != 0 && g.Spieler[i].Punkte == 0 {
				t.Fatalf("Sieger ohne Punkte: %+v", g.Spieler[i])
			}
		}
	}

	// Neue Runde: frische Steine, Punkte bleiben.
	vorher := g.Spieler[0].Punkte + g.Spieler[1].Punkte
	if r := ruf("a", "naechste", nil); r.Code != http.StatusOK {
		t.Fatalf("naechste: %d %s", r.Code, r.Body.String())
	}
	g = sicht("a")
	if g.Status != "spiel" || g.Runde != 2 || len(g.Kette) != 0 || g.Talon != 14 {
		t.Fatalf("Runde 2: status=%s runde=%d kette=%d talon=%d", g.Status, g.Runde, len(g.Kette), g.Talon)
	}
	if g.Spieler[0].Punkte+g.Spieler[1].Punkte != vorher {
		t.Fatalf("Punkte verloren: %d → %d", vorher, g.Spieler[0].Punkte+g.Spieler[1].Punkte)
	}
	if len(g.Spieler[0].Hand) != 7 || g.Spieler[1].Anzahl != 7 {
		t.Fatalf("neue Hände: %d eigene, %d fremde", len(g.Spieler[0].Hand), g.Spieler[1].Anzahl)
	}

	// Ein Spieler verlässt den Tisch – die Runde ist zu Ende.
	if drin, fertig, rest := dominoRemoveSeat(room.Domino, "b", "Ben"); !drin || !fertig || rest != 1 {
		t.Fatalf("Ben raus: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	if r := ruf("a", "reset", nil); r.Code != http.StatusOK {
		t.Fatalf("reset: %d", r.Code)
	}
	if sicht("a") != nil {
		t.Fatal("Runde blieb stehen")
	}
}

// TestDominoBlockiert prüft das Ende, wenn keiner mehr legen kann, und die
// Regel, dass erst gezogen und dann gepasst wird.
func TestDominoBlockiert(t *testing.T) {
	g := &DominoGame{
		ID: "x", HostID: "a", Status: "spiel", Runde: 1, Zug: 0,
		Kette:  [][2]int{{1, 2}},
		Enden:  [2]int{1, 2},
		Spieler: []DominoSpieler{
			{ID: "a", Name: "Anna", Hand: [][2]int{{4, 5}}},
			{ID: "b", Name: "Ben", Hand: [][2]int{{6, 6}, {5, 5}}},
		},
		vorrat: [][2]int{},
	}
	g.Talon = 0

	if text := g.dominoZiehen(0); text != "Der Vorrat ist leer – bitte passen." {
		t.Fatalf("Ziehen ohne Vorrat: %q", text)
	}
	// Anna passt – Ben kann auch nicht legen, also ist die Runde blockiert.
	if text := g.dominoPassen(0); text != "" {
		t.Fatalf("Passen: %q", text)
	}
	if g.Status != "ende" {
		t.Fatalf("blockiert, aber status=%s", g.Status)
	}
	// Anna hat 9 Augen, Ben 22 – Anna gewinnt und bekommt Bens Augen.
	if g.Sieger != "a" || g.Spieler[0].Punkte != 22 {
		t.Fatalf("Sieger %q, Punkte %d (erwartet a/22)", g.Sieger, g.Spieler[0].Punkte)
	}

	// Wer legen kann, darf nicht passen.
	g2 := &DominoGame{
		Status: "spiel", Zug: 0, Kette: [][2]int{{1, 2}}, Enden: [2]int{1, 2},
		Spieler: []DominoSpieler{
			{ID: "a", Name: "Anna", Hand: [][2]int{{2, 6}, {4, 4}}},
			{ID: "b", Name: "Ben", Hand: [][2]int{{0, 3}}},
		},
		vorrat: [][2]int{{3, 3}},
	}
	g2.Talon = 1
	if text := g2.dominoPassen(0); text != "Du kannst noch legen." {
		t.Fatalf("Passen mit passendem Stein: %q", text)
	}
	if text := g2.dominoZiehen(0); text != "Du kannst noch legen." {
		t.Fatalf("Ziehen mit passendem Stein: %q", text)
	}
	if text := g2.dominoLegen(0, [2]int{2, 6}, ""); text != "" {
		t.Fatalf("Legen rechts: %q", text)
	}
	if len(g2.Kette) != 2 || g2.Kette[1] != [2]int{2, 6} || g2.Enden[1] != 6 {
		t.Fatalf("rechts angelegt: %v / %v", g2.Kette, g2.Enden)
	}
	// Ben hat 0|3: passt an keine Seite (links 1, rechts 6).
	if text := g2.dominoLegen(1, [2]int{0, 3}, "links"); text != "Dieser Stein passt nicht an die linke Seite." {
		t.Fatalf("falsche Seite: %q", text)
	}
	if text := g2.dominoLegen(1, [2]int{0, 3}, "rechts"); text != "Dieser Stein passt nicht an die rechte Seite." {
		t.Fatalf("falsche Seite rechts: %q", text)
	}
	if text := g2.dominoZiehen(1); text != "" {
		t.Fatalf("Ben zieht: %q", text)
	}
	if len(g2.Spieler[1].Hand) != 2 || g2.Talon != 0 {
		t.Fatalf("nach dem Ziehen: %d Steine, Vorrat %d", len(g2.Spieler[1].Hand), g2.Talon)
	}
	// Passt der gezogene Stein nicht, geht der Zug weiter.
	if g2.Zug != 0 {
		t.Fatalf("Zug nach dem Ziehen: %d", g2.Zug)
	}
}
