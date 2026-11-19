package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// umbenennenTestRaum baut einen Raum mit Admin, Mitspieler, Chatverlauf und
// laufender Partie – so wie er im Betrieb aussieht.
func umbenennenTestRaum() (*Room, string) {
	now := time.Now()
	code := "TG-UMB1"
	player := func(name string) Player {
		return Player{Name: name, Token: "token-" + name, Last: now, LastAction: now}
	}
	room := &Room{
		Code: code, OwnerID: "admin", OwnerName: "Anna", Name: roomTitle("Anna"),
		Created: now, LastActive: now,
		Players: map[string]Player{"admin": player("Anna"), "b": player("Ben")},
		Mutes:   map[string]time.Time{}, MsgTimes: map[string][]time.Time{},
		Messages: []Message{
			{Name: "Ben", Text: "Hallo zusammen", Time: "10:00:00", SenderID: "b"},
			{Name: "Anna", Text: "psst", Time: "10:00:05", SenderID: "admin", ToID: "b", ToName: "Ben", Whisper: true},
			{Text: "Ben ist wohl gerade auf dem Klo oder eingeschlafen.", Time: "10:01:00", System: true},
			{Name: "Anna", Text: "bis gleich", Time: "10:02:00", SenderID: "admin"},
		},
		Angebot: &Angebot{ID: "A1", Game: "pig", GameName: "Pig", VonID: "b", VonName: "Ben"},
		Gaense: &GaenseGame{ID: "G1", HostID: "b", Status: "spiel", Runde: 1, Zug: 1,
			Spieler: []GaenseSpieler{
				{ID: "admin", Name: "Anna", Farbe: 0, Pos: 12},
				{ID: "b", Name: "Ben", Farbe: 1, Pos: 17, Aussetzen: 1},
			}},
		Tron: &TronGame{ID: "T1", HostID: "b", Status: "spiel", Runde: 2,
			Spieler: []TronSpieler{
				{ID: "b", Name: "Ben", X: 7, Y: 3, DX: 1},
				{ID: "admin", Name: "Anna", X: 20, Y: 9},
			}},
	}
	mu.Lock()
	rooms[code] = room
	mu.Unlock()
	return room, code
}

func umbenennenRuf(code, id, ziel, name, vorher string) *httptest.ResponseRecorder {
	rumpf, _ := json.Marshal(map[string]any{
		"room": code, "id": id, "targetId": ziel, "name": name, "vorher": vorher,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/umbenennen", strings.NewReader(string(rumpf)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	umbenennenHandler(rec, req)
	return rec
}

func raumAufraeumen(code string) {
	mu.Lock()
	delete(rooms, code)
	mu.Unlock()
}

// Der Admin benennt um; Name, Chatverlauf, Angebot und alle Partien ziehen mit,
// Spielkennung, Figuren, Punkte und Spielstand bleiben unverändert.
func TestUmbenennenZiehtAllesNach(t *testing.T) {
	room, code := umbenennenTestRaum()
	defer raumAufraeumen(code)

	rec := umbenennenRuf(code, "admin", "b", "  Ben2  ", "Ben")
	if rec.Code != http.StatusOK {
		t.Fatalf("umbenennen: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	var antwort map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &antwort); err != nil {
		t.Fatalf("Antwort: %v", err)
	}
	if antwort["name"] != "Ben2" || antwort["vorher"] != "Ben" {
		t.Fatalf("Antwort: %v", antwort)
	}

	mu.Lock()
	defer mu.Unlock()
	ben := room.Players["b"]
	if ben.Name != "Ben2" {
		t.Fatalf("Raumname: %q", ben.Name)
	}
	if ben.Token != "token-Ben" {
		t.Fatalf("Token wurde verändert: %q", ben.Token)
	}
	// Chat: alte Zeilen tragen den neuen Namen, Absender und Empfänger.
	if room.Messages[0].Name != "Ben2" || room.Messages[0].Text != "Hallo zusammen" {
		t.Fatalf("alte Nachricht: %+v", room.Messages[0])
	}
	if room.Messages[1].ToName != "Ben2" || room.Messages[1].Name != "Anna" {
		t.Fatalf("Flüsterzeile: %+v", room.Messages[1])
	}
	if !strings.HasPrefix(room.Messages[2].Text, "Ben2 ") {
		t.Fatalf("Systemzeile nicht mitgezogen: %q", room.Messages[2].Text)
	}
	// Neue Meldung nennt alten und neuen Namen.
	letzte := room.Messages[len(room.Messages)-1]
	if letzte.SenderID != "b" || letzte.Name != "Ben" || letzte.Text != "heißt jetzt Ben2." {
		t.Fatalf("Meldung: %+v", letzte)
	}
	// Angebot und Partien.
	if room.Angebot.VonName != "Ben2" {
		t.Fatalf("Angebot: %q", room.Angebot.VonName)
	}
	g := room.Gaense
	if g.Spieler[1].Name != "Ben2" || g.Spieler[1].Pos != 17 || g.Spieler[1].Aussetzen != 1 ||
		g.Spieler[1].ID != "b" || g.Spieler[0].Name != "Anna" {
		t.Fatalf("Gänsespiel: %+v", g.Spieler[1])
	}
	if g.Status != "spiel" || g.Zug != 1 || g.Runde != 1 {
		t.Fatalf("Gänsespiel-Stand verändert: %s/%d/%d", g.Status, g.Zug, g.Runde)
	}
	tr := room.Tron
	if tr.Spieler[0].Name != "Ben2" || tr.Spieler[0].X != 7 || tr.Spieler[0].Y != 3 || tr.Spieler[0].DX != 1 {
		t.Fatalf("Tron: %+v", tr.Spieler[0])
	}
	if tr.Status != "spiel" || tr.Runde != 2 {
		t.Fatalf("Tron-Stand verändert: %s/%d", tr.Status, tr.Runde)
	}
}

// Nur der Admin darf umbenennen – auch mit gültigem Token eines Mitspielers.
func TestUmbenennenNurAdmin(t *testing.T) {
	room, code := umbenennenTestRaum()
	defer raumAufraeumen(code)

	rec := umbenennenRuf(code, "b", "admin", "Chef", "Anna")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Mitspieler durfte umbenennen: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	mu.Lock()
	name := room.Players["admin"].Name
	mu.Unlock()
	if name != "Anna" {
		t.Fatalf("Name trotz Absage geändert: %q", name)
	}
}

// Ungültige Namen werden abgewiesen, zu lange gekürzt, Doppelte bekommen eine
// Nummer, und ein veralteter Dialog überschreibt keine neuere Änderung.
func TestUmbenennenValidierung(t *testing.T) {
	room, code := umbenennenTestRaum()
	defer raumAufraeumen(code)

	for _, leer := range []string{"", "   "} {
		if rec := umbenennenRuf(code, "admin", "b", leer, "Ben"); rec.Code != http.StatusBadRequest {
			t.Fatalf("leerer Name %q: %d", leer, rec.Code)
		}
	}
	if rec := umbenennenRuf(code, "admin", "gibtsnicht", "Egon", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("unbekannter Spieler: %d", rec.Code)
	}
	if rec := umbenennenRuf(code, "admin", "b", "Anna", "Ben"); rec.Code != http.StatusOK {
		t.Fatalf("doppelter Name: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	mu.Lock()
	if room.Players["b"].Name != "Anna1" {
		mu.Unlock()
		t.Fatalf("doppelter Name: %q", room.Players["b"].Name)
	}
	mu.Unlock()

	lang := strings.Repeat("L", maxNameLen+8)
	if rec := umbenennenRuf(code, "admin", "b", lang, "Anna1"); rec.Code != http.StatusOK {
		t.Fatalf("langer Name: %d", rec.Code)
	}
	mu.Lock()
	gekuerzt := room.Players["b"].Name
	mu.Unlock()
	if len([]rune(gekuerzt)) != maxNameLen {
		t.Fatalf("Länge: %d Zeichen (%q)", len([]rune(gekuerzt)), gekuerzt)
	}

	// Zwei Dialoge auf demselben Stand: Der zweite (veraltete) darf nicht
	// gewinnen, und der Name bleibt, wie ihn der erste gesetzt hat.
	if rec := umbenennenRuf(code, "admin", "b", "Ben3", gekuerzt); rec.Code != http.StatusOK {
		t.Fatalf("zweiter Ruf: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	if rec := umbenennenRuf(code, "admin", "b", "Ben4", gekuerzt); rec.Code != http.StatusConflict {
		t.Fatalf("veralteter Dialog: %d statt 409", rec.Code)
	}
	mu.Lock()
	name := room.Players["b"].Name
	mu.Unlock()
	if name != "Ben3" {
		t.Fatalf("veralteter Dialog hat überschrieben: %q", name)
	}
}

// Wird der Admin selbst umbenannt, ziehen Raumname und Besitzername mit.
func TestUmbenennenAdminSelbst(t *testing.T) {
	room, code := umbenennenTestRaum()
	defer raumAufraeumen(code)

	if rec := umbenennenRuf(code, "admin", "admin", "Chefin", "Anna"); rec.Code != http.StatusOK {
		t.Fatalf("Admin umbenennen: %d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	mu.Lock()
	defer mu.Unlock()
	if room.Players["admin"].Name != "Chefin" || room.OwnerName != "Chefin" {
		t.Fatalf("Admin: %q / %q", room.Players["admin"].Name, room.OwnerName)
	}
	if room.Name != roomTitle("Chefin") {
		t.Fatalf("Raumname: %q", room.Name)
	}
	if room.OwnerID != "admin" {
		t.Fatalf("Besitz hat gewechselt: %q", room.OwnerID)
	}
}
