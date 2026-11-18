package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBingoRegeln(t *testing.T) {
	now := time.Now()
	room := &Room{Code: "TG-BIN1", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
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
		req := httptest.NewRequest(http.MethodPost, "/api/bingo?room="+room.Code+"&action="+aktion, strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		bingoHandler(rec, req)
		return rec
	}
	sicht := func(id string) *BingoGame {
		req := httptest.NewRequest(http.MethodGet, "/api/bingo?room="+room.Code+"&id="+id, nil)
		rec := httptest.NewRecorder()
		bingoHandler(rec, req)
		var antwort struct {
			Game *BingoGame `json:"game"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &antwort)
		return antwort.Game
	}

	if r := ruf("a", "create", nil); r.Code != http.StatusOK {
		t.Fatalf("create: %d %s", r.Code, r.Body.String())
	}
	ruf("b", "join", nil)
	ruf("c", "join", nil)
	// Vor der Bereitschaft startet niemand.
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
	if g.Status != "spiel" || g.Runde != 1 || len(g.Spieler) != 3 {
		t.Fatalf("Start: status=%s runde=%d spieler=%d", g.Status, g.Runde, len(g.Spieler))
	}
	if g.Rest != 75 || len(g.Gezogen) != 0 {
		t.Fatalf("Start mit gezogenen Zahlen: rest=%d gezogen=%d", g.Rest, len(g.Gezogen))
	}

	// Karte: 25 Felder, Mitte frei, je Spalte der richtige Bereich, keine Doppelten.
	eigene := g.Spieler[0]
	if len(eigene.Karte) != bingoFeld || eigene.Karte[bingoFrei] != 0 {
		t.Fatalf("Karte unvollständig: %v", eigene.Karte)
	}
	gesehen := map[int]bool{}
	for feld, zahl := range eigene.Karte {
		if feld == bingoFrei {
			continue
		}
		spalte := feld % 5
		if zahl < spalte*bingoSpalten+1 || zahl > spalte*bingoSpalten+bingoSpalten {
			t.Fatalf("Zahl %d passt nicht in Spalte %d", zahl, spalte)
		}
		if gesehen[zahl] {
			t.Fatalf("Zahl %d doppelt auf der Karte", zahl)
		}
		gesehen[zahl] = true
	}
	if eigene.Anzahl != 1 || eigene.Linien != 0 {
		t.Fatalf("ungemarkte Karte: anzahl=%d linien=%d", eigene.Anzahl, eigene.Linien)
	}
	// Fremde Karten bleiben verborgen.
	if len(g.Spieler[1].Karte) != 0 || len(g.Spieler[1].Markiert) != 0 {
		t.Fatalf("fremde Karte sichtbar: %v", g.Spieler[1].Karte)
	}
	if fremd := sicht("b"); len(fremd.Spieler[0].Karte) != 0 || len(fremd.Spieler[1].Karte) != bingoFeld {
		t.Fatal("Ansicht liefert die falsche Karte")
	}

	// Nur wer dran ist, zieht.
	dran := g.Spieler[g.Zug].ID
	anderer := ""
	for _, s := range g.Spieler {
		if s.ID != dran {
			anderer = s.ID
		}
	}
	if r := ruf(anderer, "ziehen", nil); r.Code != http.StatusConflict {
		t.Fatalf("fremdes Ziehen: %d", r.Code)
	}
	if r := ruf(dran, "ziehen", nil); r.Code != http.StatusOK {
		t.Fatalf("ziehen: %d %s", r.Code, r.Body.String())
	}
	g = sicht("a")
	if len(g.Gezogen) != 1 || g.Rest != 74 {
		t.Fatalf("nach dem Ziehen: gezogen=%d rest=%d", len(g.Gezogen), g.Rest)
	}
	gezogen := append([]int(nil), g.Gezogen...)

	// Markieren: nur Zahlen der eigenen Karte, Mittelfeld nicht.
	if r := ruf("a", "markieren", map[string]any{"zahl": 999}); r.Code != http.StatusConflict {
		t.Fatalf("fremde Zahl markiert: %d", r.Code)
	}
	meine := sicht("a").Spieler[0]
	if r := ruf("a", "markieren", map[string]any{"zahl": meine.Karte[0]}); r.Code != http.StatusOK {
		t.Fatalf("markieren: %d %s", r.Code, r.Body.String())
	}
	// Zurücknehmen geht auch.
	if r := ruf("a", "markieren", map[string]any{"zahl": meine.Karte[0]}); r.Code != http.StatusOK {
		t.Fatalf("Marke zurücknehmen: %d", r.Code)
	}
	if sicht("a").Spieler[0].Anzahl != 1 {
		t.Fatal("Marke blieb stehen")
	}

	// Volle erste Reihe ohne die fünfte gezogene Zahl: Bingo gilt nicht.
	reihe := meine.Karte[0:5]
	for _, zahl := range reihe {
		if r := ruf("a", "markieren", map[string]any{"zahl": zahl}); r.Code != http.StatusOK {
			t.Fatalf("Reihe markieren (%d): %d", zahl, r.Code)
		}
	}
	// Die Meldung gilt nicht: der Fehlgriff steht im Spielstand, die Antwort
	// bleibt erfolgreich (der Zug ist angekommen, nur daneben).
	rec := ruf("a", "bingo", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("Bingo ohne gezogene Zahl: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "wurde noch nicht gezogen") {
		t.Fatalf("falscher Grund: %s", rec.Body.String())
	}
	verpasst := sicht("a").Spieler[0]
	if !verpasst.Verpasst || verpasst.Fehlgriff != 1 {
		t.Fatalf("Fehlgriff nicht vermerkt: verpasst=%v fehlgriff=%d", verpasst.Verpasst, verpasst.Fehlgriff)
	}
	// Wer daneben ruft, ist für die Runde gesperrt – auch mit gültiger Reihe.
	if r := ruf("a", "bingo", nil); r.Code != http.StatusConflict {
		t.Fatalf("zweiter Versuch nach Fehlgriff: %d", r.Code)
	}

	// Bis alle fünf Zahlen der Reihe gezogen sind, zieht reihum jeder.
	for i := 0; i < 200; i++ {
		g = sicht("a")
		if g.Status != "spiel" {
			break
		}
		alle := true
		for _, zahl := range reihe {
			if !enthaeltZahl(g.Gezogen, zahl) {
				alle = false
			}
		}
		if alle {
			break
		}
		ruf(g.Spieler[g.Zug].ID, "ziehen", nil)
	}
	g = sicht("a")
	if g.Status != "spiel" {
		t.Fatalf("Runde endete zu früh: %s", g.Note)
	}
	for _, zahl := range reihe {
		if !enthaeltZahl(g.Gezogen, zahl) {
			t.Fatalf("Reihe nicht komplett gezogen: %v / %v", reihe, g.Gezogen)
		}
	}
	// „a" war gesperrt; also gewinnt jemand anders mit seiner eigenen Reihe.
	gewinner := ""
	for _, s := range g.Spieler {
		if s.ID == "a" {
			continue
		}
		ei := sicht(s.ID)
		var ich BingoSpieler
		for _, x := range ei.Spieler {
			if x.ID == s.ID {
				ich = x
			}
		}
		for zeile := 0; zeile < 5; zeile++ {
			fertig := true
			for _, zahl := range ich.Karte[zeile*5 : zeile*5+5] {
				if !enthaeltZahl(ei.Gezogen, zahl) {
					fertig = false
				}
			}
			if !fertig {
				continue
			}
			for _, zahl := range ich.Karte[zeile*5 : zeile*5+5] {
				ruf(s.ID, "markieren", map[string]any{"zahl": zahl})
			}
			if r := ruf(s.ID, "bingo", nil); r.Code == http.StatusOK {
				gewinner = s.ID
			}
			break
		}
		if gewinner != "" {
			break
		}
	}
	if gewinner == "" {
		// Keine komplette Reihe bei den anderen: weiter ziehen, bis eine da ist.
		for i := 0; i < 400 && gewinner == ""; i++ {
			g = sicht("a")
			if g.Status != "spiel" {
				break
			}
			ruf(g.Spieler[g.Zug].ID, "ziehen", nil)
			g = sicht("a")
			for _, s := range g.Spieler {
				if s.ID == "a" || s.Verpasst {
					continue
				}
				ei := sicht(s.ID)
				var ich BingoSpieler
				for _, x := range ei.Spieler {
					if x.ID == s.ID {
						ich = x
					}
				}
				for zeile := 0; zeile < 5; zeile++ {
					fertig := true
					for _, zahl := range ich.Karte[zeile*5 : zeile*5+5] {
						if !enthaeltZahl(ei.Gezogen, zahl) {
							fertig = false
						}
					}
					if !fertig {
						continue
					}
					for _, zahl := range ich.Karte[zeile*5 : zeile*5+5] {
						ruf(s.ID, "markieren", map[string]any{"zahl": zahl})
					}
					if r := ruf(s.ID, "bingo", nil); r.Code == http.StatusOK {
						gewinner = s.ID
					}
					break
				}
			}
		}
	}
	if gewinner == "" {
		t.Fatal("niemand konnte Bingo melden")
	}
	g = sicht("a")
	if g.Status != "ende" || g.Sieger != gewinner {
		t.Fatalf("Ende: status=%s sieger=%s erwartet=%s", g.Status, g.Sieger, gewinner)
	}
	if len(gezogen) != 1 {
		t.Fatalf("gezogene Zahlen verändert: %v", gezogen)
	}

	// Nächste Runde: neue Karten, leere Ziehung, Runde 2.
	if r := ruf(gewinner, "naechste", nil); r.Code != http.StatusOK {
		t.Fatalf("naechste: %d %s", r.Code, r.Body.String())
	}
	g = sicht("a")
	if g.Status != "spiel" || g.Runde != 2 || len(g.Gezogen) != 0 || g.Rest != 75 {
		t.Fatalf("Runde 2: status=%s runde=%d gezogen=%d rest=%d", g.Status, g.Runde, len(g.Gezogen), g.Rest)
	}
	if g.Spieler[0].Verpasst || g.Spieler[0].Anzahl != 1 || g.Spieler[0].Linien != 0 {
		t.Fatalf("Karte der neuen Runde nicht frisch: %+v", g.Spieler[0])
	}

	// Alle Meldungen daneben: die Runde endet ohne Sieger.
	for _, id := range []string{"a", "b", "c"} {
		if r := ruf(id, "bingo", nil); r.Code != http.StatusOK {
			t.Fatalf("Bingo ohne Marken (%s): %d %s", id, r.Code, r.Body.String())
		}
	}
	g = sicht("a")
	if g.Status != "ende" || g.Sieger != "" {
		t.Fatalf("Runde ohne Sieger: status=%s sieger=%q", g.Status, g.Sieger)
	}

	// Verlassen nimmt den Platz, der letzte beendet die Runde.
	if drin, fertig, rest := bingoRemoveSeat(room.Bingo, "c", "Cem"); !drin || fertig || rest != 2 {
		t.Fatalf("Cem raus: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	if drin, fertig, rest := bingoRemoveSeat(room.Bingo, "b", "Ben"); !drin || !fertig || rest != 1 {
		t.Fatalf("Ben raus: drin=%v fertig=%v rest=%d", drin, fertig, rest)
	}
	// Aufräumen durch den Gastgeber.
	if r := ruf("a", "reset", nil); r.Code != http.StatusOK {
		t.Fatalf("reset: %d", r.Code)
	}
	if sicht("a") != nil {
		t.Fatal("Runde blieb stehen")
	}
	t.Logf("OK: Karten geprüft, Ziehen/Markieren/Bingo, Sieg durch %s, Runde ohne Sieger, Verlassen", gewinner)
}

// TestBingoTisch: das Spiel lässt sich anbieten, beitreten und starten wie die
// anderen Tischspiele.
func TestBingoTisch(t *testing.T) {
	now := time.Now()
	room := &Room{Code: "TG-BIN2", OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
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

	if !tischspiel["bingo"] {
		t.Fatal("Bingo fehlt bei den Tischspielen")
	}
	if nr, text := spielAufruf(spielApi["bingo"], room.Code, "create", map[string]any{"id": "a"}); nr >= 300 {
		t.Fatalf("create über den Connector: %d %s", nr, text)
	}
	if nr, text := spielAufruf(spielApi["bingo"], room.Code, "join", map[string]any{"id": "b"}); nr >= 300 {
		t.Fatalf("join über den Connector: %d %s", nr, text)
	}
	mu.Lock()
	spieler := len(rooms[room.Code].Bingo.Spieler)
	mu.Unlock()
	if spieler != 2 {
		t.Fatalf("Plätze nach dem Beitritt: %d", spieler)
	}
	// Umbenennen erreicht auch den Bingo-Tisch.
	mu.Lock()
	rooms[room.Code].Players["b"] = Player{Name: "Berta", Token: "tb", Last: now, LastAction: now}
	rooms[room.Code].spielerUmbenennen("b", "Berta")
	name := rooms[room.Code].Bingo.Spieler[1].Name
	mu.Unlock()
	if name != "Berta" {
		t.Fatalf("Name am Tisch: %q", name)
	}
}
