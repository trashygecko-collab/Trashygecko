package main

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// spielFeld ordnet jeder Spielkennung das Feld im Raum zu, in dem die Partie
// liegt. Genau hier waren Lücken entstanden: „17 und 4“ und „Sechsundsechzig“
// fehlten in einzelnen Tabellen, weil jede Liste für sich gepflegt wurde.
//
// Der Test prüft, dass jedes Spiel aus SUPPORTED_GAMES (der Liste in
// connector.html) in allen Spieltabellen vorkommt. Kommt ein Spiel dazu oder
// wird eines umbenannt, schlägt er fehl, bis alle Stellen nachgezogen sind.
var spielFeld = map[string]string{
	"tictactoe":        "TicTacToe",
	"viergewinnt":      "VierGewinnt",
	"dame":             "Dame",
	"paare":            "Paare",
	"menschaerger":     "Menschaerger",
	"kaesekaestche":    "Kaese",
	"sim":              "Sim",
	"tetris":           "Tetris",
	"kalaha":           "Kalaha",
	"streichholzspiel": "Streichholz",
	"muehle":           "Muehle",
	"reversi":          "Reversi",
	"hold":             "Hold",
	"skyjo":            "Skyjo",
	"kniffel":          "Kniffel",
	"schach":           "Schach",
	"go":               "Go",
	"backgammon":       "Backgammon",
	"maumau":           "MauMau",
	"siebzehn":         "Siebzehn",
	"baccara":          "Baccara",
	"schiffe":          "Schiffe",
	"pong":             "Pong",
	"tron":             "Tron",
	"bomberman":        "Bomberman",
	"bubble":           "Bubble",
	"snake":            "Snake",
	"cabo":             "Cabo",
	"poker":            "Poker",
	"uno":              "Uno",
	"schwimmen":        "Schwimmen",
	"arschloch":        "Arschloch",
	"einsxeins":        "Rechen",
	"sechsundsechzig":  "Sechsundsechzig",
	"galgen":           "Galgen",
	"bingo":            "Bingo",
	"wortsuche":        "Wortsuche",
	"domino":           "Domino",
	"schlangen":        "Schlangen",
	"pig":              "Pig",
	"gaensespiel":      "Gaense",
	"malen":            "Mal",
}

// spielLautSeite ist ein Eintrag aus SUPPORTED_GAMES in connector.html.
type spielLautSeite struct {
	ID   string
	Name string
	API  string
	Pfad string
}

var spielZeileMuster = regexp.MustCompile(`id:\s*"([^"]+)",\s*name:\s*"([^"]*)",\s*path:\s*"([^"]*)"(?:,\s*api:\s*"([^"]*)")?`)

// spieleAusSeite liest die Spieleliste aus der Oberfläche – sie ist die Quelle,
// an der sich der Server orientieren muss.
func spieleAusSeite(t *testing.T) []spielLautSeite {
	t.Helper()
	roh, err := os.ReadFile("connector.html")
	if err != nil {
		t.Fatalf("connector.html nicht lesbar: %v", err)
	}
	text := string(roh)
	anfang := strings.Index(text, "const SUPPORTED_GAMES")
	if anfang < 0 {
		t.Fatal("SUPPORTED_GAMES nicht gefunden")
	}
	ende := strings.Index(text[anfang:], "];")
	if ende < 0 {
		t.Fatal("Ende von SUPPORTED_GAMES nicht gefunden")
	}
	block := text[anfang : anfang+ende]
	treffer := spielZeileMuster.FindAllStringSubmatch(block, -1)
	spiele := make([]spielLautSeite, 0, len(treffer))
	for _, m := range treffer {
		spiele = append(spiele, spielLautSeite{ID: m[1], Name: m[2], Pfad: m[3], API: m[4]})
	}
	if len(spiele) < 20 {
		t.Fatalf("nur %d Spiele in SUPPORTED_GAMES gelesen – Muster passt nicht mehr", len(spiele))
	}
	return spiele
}

// funktionstext schneidet den Rumpf einer Methode aus einer Quelldatei heraus.
func funktionstext(t *testing.T, datei, kopf string) string {
	t.Helper()
	roh, err := os.ReadFile(datei)
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", datei, err)
	}
	text := string(roh)
	i := strings.Index(text, kopf)
	if i < 0 {
		t.Fatalf("%q nicht in %s gefunden", kopf, datei)
	}
	rest := text[i:]
	if ende := strings.Index(rest, "\n}\n"); ende >= 0 {
		return rest[:ende]
	}
	return rest
}

func schluessel(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func schluesselBool(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestSpieltabellenVollstaendig meldet jede Lücke zwischen der Spieleliste der
// Oberfläche und den Tabellen des Servers.
func TestSpieltabellenVollstaendig(t *testing.T) {
	spiele := spieleAusSeite(t)

	// Die Tabellen, in denen ein Spiel stehen muss.
	ip := funktionstext(t, "leave.go", "func (r *Room) playerInGame(")
	aktiv := funktionstext(t, "leave.go", "func (r *Room) aktivesSpiel(")
	partien := funktionstext(t, "leave.go", "func (r *Room) partien()")
	routen := funktionstext(t, "server.go", "func routes()")

	inSeite := map[string]bool{}
	for _, s := range spiele {
		inSeite[s.ID] = true

		if _, ok := spielFeld[s.ID]; !ok {
			t.Errorf("%s: fehlt in spielFeld (diesem Test) – neues Spiel?", s.ID)
		}
		api, ok := spielApi[s.ID]
		if !ok {
			t.Errorf("%s: fehlt in spielApi – „Runde anbieten“ scheitert", s.ID)
		} else {
			if s.API != "" && "/api/"+api != s.API {
				t.Errorf("%s: spielApi zeigt auf %q, die Oberfläche auf %q", s.ID, api, s.API)
			}
			if !strings.Contains(routen, `"/api/`+api+`"`) {
				t.Errorf("%s: kein Endpunkt /api/%s in routes()", s.ID, api)
			}
		}
		if name, ok := spielAnzeige[s.ID]; !ok {
			t.Errorf("%s: fehlt in spielAnzeige – im Chat steht nur die Kennung", s.ID)
		} else if s.Name != "" && name != s.Name {
			t.Errorf("%s: spielAnzeige heißt %q, die Oberfläche %q", s.ID, name, s.Name)
		}

		// Malen ist eine gemeinsame Fläche: niemand „sitzt“ darin, sie endet nie.
		feld := spielFeld[s.ID]
		if s.ID != "malen" && feld != "" {
			for name, text := range map[string]string{
				"playerInGame": ip,
				"aktivesSpiel": aktiv,
				"partien":      partien,
			} {
				if !strings.Contains(text, "r."+feld) {
					t.Errorf("%s: %s kennt das Spiel nicht (r.%s fehlt)", s.ID, name, feld)
				}
			}
		}
	}

	// Keine Waise: Was in den Tabellen steht, muss auch in der Oberfläche sein.
	for _, id := range schluessel(spielApi) {
		if !inSeite[id] {
			t.Errorf("spielApi hat %q, SUPPORTED_GAMES nicht", id)
		}
	}
	for _, id := range schluessel(spielAnzeige) {
		if !inSeite[id] {
			t.Errorf("spielAnzeige hat %q, SUPPORTED_GAMES nicht", id)
		}
	}
	for _, id := range schluesselBool(tischspiel) {
		if !inSeite[id] {
			t.Errorf("tischspiel hat %q, SUPPORTED_GAMES nicht", id)
		}
		if gemeinsam[id] && tischspiel[id] {
			t.Errorf("%s: zugleich Tischspiel und gemeinsame Fläche", id)
		}
	}
	for _, id := range schluesselBool(gemeinsam) {
		if !inSeite[id] {
			t.Errorf("gemeinsam hat %q, SUPPORTED_GAMES nicht", id)
		}
	}
	// Auch das Umbenennen während einer Partie muss jedes Spiel kennen.
	umbenennen := funktionstext(t, "umbenennen.go", "func (r *Room) spielerUmbenennen(")
	for _, id := range schluessel(spielFeld) {
		feld := spielFeld[id]
		if feld == "" || id == "malen" {
			continue
		}
		if !strings.Contains(umbenennen, "r."+feld) {
			t.Errorf("%s: spielerUmbenennen kennt r.%s nicht – Namen bleiben in der Partie alt", id, feld)
		}
	}
	if len(spielFeld) != len(spiele) {
		t.Errorf("spielFeld hat %d Einträge, SUPPORTED_GAMES %d", len(spielFeld), len(spiele))
	}
}
