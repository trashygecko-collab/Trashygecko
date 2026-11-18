package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Der Admin benennt mitten in einer laufenden Partie um: der Name muss im
// Raum, in jeder Partie und im Chat stehen, ohne den Spielstand anzutasten.
func TestUmbenennenWaehrendPartie(t *testing.T) {
	now := time.Now()
	code := "TG-UMB9"
	room := &Room{
		Code: code, OwnerID: "a", OwnerName: "Anna", Name: roomTitle("Anna"),
		Created: now, LastActive: now,
		Players: map[string]Player{
			"a": {Name: "Anna", Token: "ta", Last: now, LastAction: now},
			"b": {Name: "Ben", Token: "tb", Last: now, LastAction: now},
			"c": {Name: "Cem", Token: "tc", Last: now, LastAction: now},
		},
		Mutes: map[string]time.Time{}, MsgTimes: map[string][]time.Time{},
		TicTacToe: &TicTacToeGame{ID: "TG-1", XID: "a", XName: "Anna", OID: "b", OName: "Ben",
			Board: []string{"X", "", "", "", "O", "", "", "", ""}, Current: "red", Status: "active"},
		Galgen: &GalgenGame{ID: "TG-2", HostID: "b", Status: "spiel", Variante: "klassisch",
			VarianteName: "Klassisch", Wort: "insel", Anzeige: []string{"", "", "", "", ""},
			MaxFehler: 8, Spieler: []GalgenSpieler{{ID: "b", Name: "Ben", Punkte: 3}, {ID: "c", Name: "Cem", Punkte: 1}}},
	}
	mu.Lock()
	rooms[code] = room
	mu.Unlock()
	defer func() {
		mu.Lock()
		delete(rooms, code)
		mu.Unlock()
	}()

	umbenennen := func(wer, ziel, neu string) *httptest.ResponseRecorder {
		rumpf, _ := json.Marshal(map[string]string{"room": code, "id": wer, "targetId": ziel, "name": neu})
		req := httptest.NewRequest(http.MethodPost, "/api/umbenennen", strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		umbenennenHandler(rec, req)
		return rec
	}

	// 1) Nur der Besitzer darf umbenennen.
	if rec := umbenennen("b", "c", "Carl"); rec.Code != http.StatusForbidden {
		t.Fatalf("Nicht-Admin: %d statt 403 (%s)", rec.Code, rec.Body.String())
	}

	// 2) Mitten in der Partie umbenennen.
	if rec := umbenennen("a", "b", "Bernhard"); rec.Code != http.StatusOK {
		t.Fatalf("Admin: %d %s", rec.Code, rec.Body.String())
	}
	mu.Lock()
	neuerName := rooms[code].Players["b"].Name
	ttt := rooms[code].TicTacToe
	galgen := rooms[code].Galgen
	letzteZeile := ""
	if nachrichten := rooms[code].Messages; len(nachrichten) > 0 {
		letzteZeile = nachrichten[len(nachrichten)-1].Name + ": " + nachrichten[len(nachrichten)-1].Text
	}
	mu.Unlock()
	if neuerName != "Bernhard" {
		t.Fatalf("Raumname: %q", neuerName)
	}
	if ttt.OName != "Bernhard" || ttt.XName != "Anna" || ttt.OID != "b" || ttt.XID != "a" {
		t.Fatalf("TicTacToe: %+v", ttt)
	}
	if len(ttt.Board) != 9 || ttt.Board[0] != "X" || ttt.Board[4] != "O" || ttt.Status != "active" {
		t.Fatalf("Spielstand verändert: %+v", ttt)
	}
	if galgen.Spieler[0].Name != "Bernhard" || galgen.Spieler[0].Punkte != 3 ||
		galgen.Spieler[1].Name != "Cem" || galgen.Wort != "insel" {
		t.Fatalf("Galgen: %+v", galgen.Spieler)
	}
	if letzteZeile != "Ben: heißt jetzt Bernhard." {
		t.Fatalf("Chatzeile: %q", letzteZeile)
	}

	// 3) Ein Lebenszeichen setzt den neuen Namen nicht zurück (Umbenennen-Flag).
	presenceRuf := func(id, name string, umbenennen bool) *httptest.ResponseRecorder {
		rumpf, _ := json.Marshal(map[string]any{"room": code, "id": id, "name": name, "token": map[string]string{
			"a": "ta", "b": "tb", "c": "tc"}[id], "umbenennen": umbenennen, "neu": true})
		req := httptest.NewRequest(http.MethodPost, "/api/presence", strings.NewReader(string(rumpf)))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.168.9.4:4444"
		rec := httptest.NewRecorder()
		presence(rec, req)
		return rec
	}
	if rec := presenceRuf("b", "Ben", false); rec.Code != http.StatusOK {
		t.Fatalf("Lebenszeichen: %d %s", rec.Code, rec.Body.String())
	}
	mu.Lock()
	nachLebenszeichen := rooms[code].Players["b"].Name
	mu.Unlock()
	if nachLebenszeichen != "Bernhard" {
		t.Fatalf("Lebenszeichen hat zurückgesetzt: %q", nachLebenszeichen)
	}

	// 4) Doppelter Name bekommt eine Nummer.
	if rec := umbenennen("a", "c", "Bernhard"); rec.Code != http.StatusOK {
		t.Fatalf("Doppelt: %d %s", rec.Code, rec.Body.String())
	}
	mu.Lock()
	doppelt := rooms[code].Players["c"].Name
	mu.Unlock()
	if doppelt != "Bernhard1" {
		t.Fatalf("Doppelter Name: %q", doppelt)
	}

	// 5) Unbekannter Spieler und leerer Name werden abgelehnt.
	if rec := umbenennen("a", "gibtsnicht", "Eva"); rec.Code != http.StatusNotFound {
		t.Fatalf("Unbekannt: %d statt 404", rec.Code)
	}
	if rec := umbenennen("a", "c", "   "); rec.Code != http.StatusBadRequest {
		t.Fatalf("Leerer Name: %d statt 400", rec.Code)
	}
}
