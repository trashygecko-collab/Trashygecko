package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"
)

// Angebot ist ein offener Spieltisch: Jemand bietet eine Runde an, andere
// nehmen Platz. Ein Raum kann mehrere Tische gleichzeitig haben – einen je
// Spiel, denn je Spiel gibt es einen Spielstand im Raum.
//
// ZielID/ZielName machen daraus eine direkte Einladung: Dann darf nur dieser
// eine Spieler annehmen oder ablehnen, und die Einladung verfällt nach
// angebotFrist – sonst blockierte eine unbeantwortete Einladung das Spiel.
type Angebot struct {
	ID         string    `json:"id"`
	Game       string    `json:"game"`
	GameName   string    `json:"gameName"`
	VonID      string    `json:"vonId"`
	VonName    string    `json:"vonName"`
	ZielID     string    `json:"zielId,omitempty"`
	ZielName   string    `json:"zielName,omitempty"`
	Pairs      int       `json:"pairs,omitempty"`
	Kind       string    `json:"kind,omitempty"`
	MaxPlayers int       `json:"maxPlayers,omitempty"`
	Zeit       time.Time `json:"-"`
	// Bereit ist abgeleitet: Das Spiel kennt einen Bereit-Knopf. Gesetzt wird
	// der Wert beim Kopieren (bereitSpiele), gespeichert nicht – so weiß die
	// Tischkarte vor dem Klick, ob sie überhaupt einen Knopf zeigen soll.
	Bereit bool `json:"bereit"`
	// Beigetreten sind die Spieler, die diesen Tisch angenommen haben. Der
	// Ersteller sitzt immer mit; die Liste wächst nur, solange der Tisch offen
	// ist. Wer den Raum verlässt, verschwindet daraus (angeboteVonSpieler).
	Beigetreten []string `json:"-"`
	// Sitze ist abgeleitet: Ersteller und Beigetretene in ihrer Reihenfolge.
	// Damit zeigt die Tischkarte, wer schon sitzt, ohne den Spielstand zu
	// lesen – und ein Beigetretener bekommt kein zweites "Mitspielen".
	Sitze []string `json:"sitze"`
}

// angebotFrist: So lange wartet eine direkte Einladung auf Antwort. Danach
// fällt der Tisch weg; sonst bliebe das Spiel besetzt, ohne dass jemand
// antworten kann.
const angebotFrist = 5 * time.Minute

// spielApi ordnet der Kennung aus der Spieleliste den API-Pfad zu.
var spielApi = map[string]string{
	"tictactoe":          "tictactoe",
	"viergewinnt":        "viergewinnt",
	"dame":               "dame",
	"paare":              "paare",
	"menschaerger":       "maensync", // Standweg mit Regeln im Browser
	"kaesekaestche":      "kaesekaestche",
	"sim":                "sim",
	"kalaha":             "kalaha",
	"streichholzspiel":   "streichholz",
	"muehle":             "muehle",
	"reversi":            "reversi",
	"hold":               "hold",
	"memory":             "memory",
	"mastermind":         "mastermind",
	"skyjo":              "skyjo",
	"kniffel":            "kniffel",
	"zehntausend":        "zehntausend",
	"maexchen":           "maexchen",
	"boston":             "boston",
	"chicago":            "chicago",
	"shipcrew":           "shipcrew",
	"halma":              "halma",
	"schach":             "schach",
	"go":                 "go",
	"backgammon":         "backgammon",
	"maumau":             "maumau",
	"uno":                "uno",
	"schwimmen":          "schwimmen",
	"arschloch":          "arschloch",
	"einsxeins":          "einsxeins",
	"poker":              "poker",
	"cabo":               "cabo",
	"siebzehn":           "siebzehn",
	"baccara":            "baccara",
	"schiffe":            "schiffe",
	"pong":               "pong",
	"tron":               "tron",
	"bomberman":          "bomberman",
	"bubble":             "bubble",
	"snake":              "snake",
	"tetris":             "tetris",
	"sechsundsechzig":    "sechsundsechzig",
	"galgen":             "galgen",
	"quizshow":           "quiz",
	"bingo":              "bingo",
	"wortsuche":          "wortsuche",
	"domino":             "domino",
	"schlangen":          "schlangen",
	"pig":                "pig",
	"gaensespiel":        "gaensespiel",
	"risiko":             "risiko",
	"labyrinth":          "labyrinth",
	"malen":              "malen",
	"der-turm":           "wuerfel",
	"dufte-hundert":      "wuerfel",
	"eins-bis-sechzehn":  "wuerfel",
	"fenster":            "wuerfel",
	"gestrichene-zwoelf": "wuerfel",
	"hin-und-her":        "wuerfel",
	"jule":               "wuerfel",
	"paerchen":           "wuerfel",
	"plus-minus":         "wuerfel",
	"rentmeister":        "wuerfel",
	"streichhoelzer":     "wuerfel",
	"teure-sieben":       "wuerfel",
	"unter-oder-ueber":   "wuerfel",
	"wuerfelpoker":       "wuerfel",
}

// spielAnzeige sind die Namen, wie sie im Chat stehen.
var spielAnzeige = map[string]string{
	"tictactoe":          "Tic Tac Toe",
	"viergewinnt":        "Vier gewinnt",
	"dame":               "Dame",
	"paare":              "Paare finden",
	"menschaerger":       "Mensch ärgere dich nicht",
	"kaesekaestche":      "Käsekästchen",
	"sim":                "Sim",
	"kalaha":             "Kalaha",
	"streichholzspiel":   "21 Streichhölzer",
	"muehle":             "Mühle",
	"reversi":            "Reversi",
	"hold":               "Hold the Line",
	"memory":             "Memory",
	"mastermind":         "Mastermind",
	"skyjo":              "SKYJO",
	"kniffel":            "Kniffel",
	"zehntausend":        "10.000",
	"maexchen":           "Lügen Mäxchen",
	"boston":             "Auf nach Boston",
	"chicago":            "Chicago",
	"shipcrew":           "Ship Captain Crew",
	"halma":              "Halma",
	"schach":             "Schach",
	"go":                 "Go",
	"backgammon":         "Backgammon",
	"maumau":             "Mau Mau",
	"uno":                "UNO",
	"schwimmen":          "31 Schwimmen",
	"arschloch":          "Arschloch",
	"einsxeins":          "1×1-Rennen",
	"poker":              "Poker",
	"cabo":               "Cabo",
	"siebzehn":           "17 und 4",
	"baccara":            "Baccara",
	"schiffe":            "Schiffe versenken",
	"pong":               "Pong",
	"tron":               "Tron",
	"bomberman":          "Bomberman",
	"bubble":             "Bubble Shooter",
	"snake":              "Snake",
	"sechsundsechzig":    "Sechsundsechzig",
	"galgen":             "Galgenmännchen",
	"quizshow":           "Scherzshow",
	"bingo":              "Bingo",
	"wortsuche":          "Wortsuche",
	"domino":             "Domino",
	"schlangen":          "Schlangen und Leitern",
	"pig":                "Pig",
	"gaensespiel":        "Gänsespiel",
	"risiko":             "Risiko",
	"labyrinth":          "Das verrückte Labyrinth",
	"tetris":             "Tetris",
	"malen":              "Malen",
	"der-turm":           "Der Turm",
	"dufte-hundert":      "Dufte Hundert",
	"eins-bis-sechzehn":  "Eins bis Sechzehn",
	"fenster":            "Fenster würfeln",
	"gestrichene-zwoelf": "Gestrichene Zwölf",
	"hin-und-her":        "Hin und her",
	"jule":               "Jule",
	"paerchen":           "Pärchen",
	"plus-minus":         "Plus-Minus",
	"rentmeister":        "Rentmeister",
	"streichhoelzer":     "Streichhölzer auswürfeln",
	"teure-sieben":       "Teure Sieben",
	"unter-oder-ueber":   "Unter oder über sieben",
	"wuerfelpoker":       "Würfelpoker",
}

// Tischspiele: dort sitzen mehrere, die Runde wird beim Anbieten eröffnet.
var tischspiel = map[string]bool{
	"kniffel":            true,
	"maumau":             true,
	"uno":                true,
	"schwimmen":          true,
	"arschloch":          true,
	"einsxeins":          true,
	"poker":              true,
	"cabo":               true,
	"siebzehn":           true,
	"baccara":            true,
	"schiffe":            true,
	"pong":               true,
	"tron":               true,
	"bomberman":          true,
	"bubble":             true,
	"snake":              true,
	"sechsundsechzig":    true,
	"galgen":             true,
	"quizshow":           true,
	"bingo":              true,
	"wortsuche":          true,
	"domino":             true,
	"schlangen":          true,
	"pig":                true,
	"gaensespiel":        true,
	"risiko":             true,
	"labyrinth":          true,
	"menschaerger":       true,
	"zehntausend":        true,
	"maexchen":           true,
	"boston":             true,
	"chicago":            true,
	"shipcrew":           true,
	"halma":              true,
	"der-turm":           true,
	"dufte-hundert":      true,
	"eins-bis-sechzehn":  true,
	"fenster":            true,
	"gestrichene-zwoelf": true,
	"hin-und-her":        true,
	"jule":               true,
	"paerchen":           true,
	"plus-minus":         true,
	"rentmeister":        true,
	"streichhoelzer":     true,
	"teure-sieben":       true,
	"unter-oder-ueber":   true,
	"wuerfelpoker":       true,
}

// Gemeinsame Fläche: Malen hat kein Gegenüber und keinen Spielstand, der
// geprüft werden müsste. Wer beitritt, malt auf demselben Brett mit – das
// Angebot bleibt offen, damit nach und nach alle dazukommen können.
var gemeinsam = map[string]bool{
	"malen": true,
}

func spielName(id string) string {
	if name, ok := spielAnzeige[id]; ok {
		return name
	}
	return id
}

// spielStatus liefert den Zustand der Runde, die zu diesem Spiel im Raum
// steht. ok=false heißt: es steht gar keine Runde. Nur Tischspiele legen beim
// Anbieten sofort eine an – bei Einladungen entsteht sie erst beim Annehmen.
func (r *Room) spielStatus(spiel string) (string, bool) {
	switch spiel {
	case "kniffel":
		if g := r.Kniffel; g != nil {
			return g.Status, true
		}
	case "maumau":
		if g := r.MauMau; g != nil {
			return g.Status, true
		}
	case "uno":
		if g := r.Uno; g != nil {
			return g.Status, true
		}
	case "schwimmen":
		if g := r.Schwimmen; g != nil {
			return g.Status, true
		}
	case "arschloch":
		if g := r.Arschloch; g != nil {
			return g.Status, true
		}
	case "einsxeins":
		if g := r.Rechen; g != nil {
			return g.Status, true
		}
	case "poker":
		if g := r.Poker; g != nil {
			return g.Status, true
		}
	case "cabo":
		if g := r.Cabo; g != nil {
			return g.Status, true
		}
	case "siebzehn":
		if g := r.Siebzehn; g != nil {
			return g.Status, true
		}
	case "baccara":
		if g := r.Baccara; g != nil {
			return g.Status, true
		}
	case "schiffe":
		if g := r.Schiffe; g != nil {
			return g.Status, true
		}
	case "pong":
		if g := r.Pong; g != nil {
			return g.Status, true
		}
	case "tron":
		if g := r.Tron; g != nil {
			return g.Status, true
		}
	case "bomberman":
		if g := r.Bomberman; g != nil {
			return g.Status, true
		}
	case "bubble":
		if g := r.Bubble; g != nil {
			return g.Status, true
		}
	case "snake":
		if g := r.Snake; g != nil {
			return g.Status, true
		}
	case "sechsundsechzig":
		if g := r.Sechsundsechzig; g != nil {
			return g.Status, true
		}
	case "galgen":
		if g := r.Galgen; g != nil {
			return g.Status, true
		}
	case "quizshow":
		if g := r.Quiz; g != nil {
			return g.Status, true
		}
	case "bingo":
		if g := r.Bingo; g != nil {
			return g.Status, true
		}
	case "wortsuche":
		if g := r.Wortsuche; g != nil {
			return g.Status, true
		}
	case "domino":
		if g := r.Domino; g != nil {
			return g.Status, true
		}
	case "schlangen":
		if g := r.Schlangen; g != nil {
			return g.Status, true
		}
	case "pig":
		if g := r.Pig; g != nil {
			return g.Status, true
		}
	case "gaensespiel":
		if g := r.Gaense; g != nil {
			return g.Status, true
		}
	case "risiko":
		if g := r.Risiko; g != nil {
			return g.Status, true
		}
	case "labyrinth":
		if g := r.Labyrinth; g != nil {
			return g.Status, true
		}
	case "menschaerger":
		if g := r.Maen; g != nil {
			return g.Status, true
		}
	case "zehntausend":
		if g := r.ZehnTausend; g != nil {
			return g.Status, true
		}
	case "maexchen":
		if g := r.Maexchen; g != nil {
			return g.Status, true
		}
	case "boston":
		if g := r.Boston; g != nil {
			return g.Status, true
		}
	case "chicago":
		if g := r.Chicago; g != nil {
			return g.Status, true
		}
	case "shipcrew":
		if g := r.ShipCrew; g != nil {
			return g.Status, true
		}
	case "halma":
		if g := r.Halma; g != nil {
			return g.Status, true
		}
	default:
		// Die 18 Würfelspiele teilen sich einen Tisch – gemeint ist nur der,
		// der auch angeboten wurde.
		if spielApi[spiel] == "wuerfel" {
			if g := r.Wuerfel; g != nil && g.Spiel == spiel {
				return g.Status, true
			}
		}
	}
	return "", false
}

// angebotNochOffen sagt, ob das Angebot noch jemanden aufnehmen kann. Neue
// Spieler nimmt allein die Lobby auf (Status "invited"); sobald die Runde
// läuft – auch wenn sie erst auf der Spielseite gestartet wurde – oder zu Ende
// ist, ist das Angebot erledigt. Die Halle muss dafür nichts aufräumen.

// angeboteKopie liefert alle offenen Tische als Kopien und räumt erledigte
// weg: gestartete Runden, verfallene Einladungen und Tische ohne Spiel.
// Muss unter mu laufen.
func angeboteKopie(room *Room) []Angebot {
	offen := []Angebot{}
	if room == nil {
		return offen
	}
	jetzt := time.Now()
	behalten := room.Angebote[:0]
	for _, a := range room.Angebote {
		if a == nil {
			continue
		}
		// Eine direkte Einladung, die niemand beantwortet, verfällt: Sonst
		// bliebe ihr Spiel besetzt, obwohl niemand mehr antworten kann.
		if a.ZielID != "" && jetzt.Sub(a.Zeit) > angebotFrist {
			continue
		}
		if !room.angebotNochOffen(a) {
			continue
		}
		behalten = append(behalten, a)
		kopie := *a
		kopie.Bereit = bereitSpiele[a.Game]
		kopie.Sitze = sitzeVon(a)
		offen = append(offen, kopie)
	}
	room.Angebote = behalten
	return offen
}

// sitzeVon nennt Ersteller und Beigetretene eines Tisches – ohne Doppelte,
// auch wenn ein alter Tisch die Kennung zweimal trägt. Muss unter mu laufen.
func sitzeVon(a *Angebot) []string {
	sitze := []string{}
	if a == nil {
		return sitze
	}
	if a.VonID != "" {
		sitze = append(sitze, a.VonID)
	}
	for _, id := range a.Beigetreten {
		if id != "" && !enthaelt(sitze, id) {
			sitze = append(sitze, id)
		}
	}
	return sitze
}

// angebotMitID sucht einen offenen Tisch. Muss unter mu laufen.
func angebotMitID(room *Room, id string) *Angebot {
	if room == nil || id == "" {
		return nil
	}
	for _, a := range room.Angebote {
		if a != nil && a.ID == id {
			return a
		}
	}
	return nil
}

// angebotFuerSpiel sucht den offenen Tisch dieses Spiels. Je Spiel gibt es
// einen Spielstand im Raum, also auch höchstens einen Tisch. Muss unter mu
// laufen.
func angebotFuerSpiel(room *Room, spiel string) *Angebot {
	if room == nil || spiel == "" {
		return nil
	}
	for _, a := range room.Angebote {
		if a != nil && a.Game == spiel {
			return a
		}
	}
	return nil
}

// angebotEntfernen nimmt genau diesen Tisch aus dem Raum. Muss unter mu laufen.
func angebotEntfernen(room *Room, id string) {
	if room == nil || id == "" {
		return
	}
	rest := room.Angebote[:0]
	for _, a := range room.Angebote {
		if a != nil && a.ID != id {
			rest = append(rest, a)
		}
	}
	room.Angebote = rest
}

// angeboteVonSpieler entfernt alle Tische eines Spielers – beim Gehen kann
// niemand mehr annehmen – und streicht ihn aus den Sitzen fremder Tische.
// Sonst stünde auf einer Tischkarte noch jemand, der längst weg ist.
// Muss unter mu laufen.
func angeboteVonSpieler(room *Room, id string) {
	if room == nil || id == "" {
		return
	}
	rest := room.Angebote[:0]
	for _, a := range room.Angebote {
		if a == nil || a.VonID == id {
			continue
		}
		if enthaelt(a.Beigetreten, id) {
			bleibt := a.Beigetreten[:0]
			for _, x := range a.Beigetreten {
				if x != id {
					bleibt = append(bleibt, x)
				}
			}
			a.Beigetreten = bleibt
		}
		rest = append(rest, a)
	}
	room.Angebote = rest
}

func (r *Room) angebotNochOffen(a *Angebot) bool {
	if a == nil {
		return false
	}
	// Malen ist eine gemeinsame Fläche ohne Plätze: sie bleibt offen.
	if gemeinsam[a.Game] {
		return true
	}
	// Einladungen an einen Spieler erledigen sich beim Annehmen oder Ablehnen.
	if !tischspiel[a.Game] {
		return true
	}
	status, da := r.spielStatus(a.Game)
	if !da {
		return false
	}
	return status == "invited"
}

// spielAufruf ruft einen Spiel-Endpunkt von innen auf – wie ein Klick eines
// Spielers. So gelten beim Beitritt genau dieselben Prüfungen wie sonst.
func spielAufruf(api, code, aktion string, felder map[string]any) (int, string) {
	// Der Aufruf gilt als Klick des Spielers, also gehört sein Token dazu –
	// sonst weist die Kennungsprüfung ihn als fremdes Gerät ab.
	if id, ok := felder["id"].(string); ok && id != "" {
		mu.Lock()
		if room := rooms[code]; room != nil {
			if p, gefunden := room.Players[id]; gefunden {
				felder["token"] = p.Token
			}
		}
		mu.Unlock()
	}
	body, _ := json.Marshal(felder)
	ziel := "/api/" + api + "?room=" + url.QueryEscape(code) + "&action=" + url.QueryEscape(aktion)
	req := httptest.NewRequest(http.MethodPost, ziel, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	routes().ServeHTTP(rec, req)
	return rec.Code, strings.TrimSpace(rec.Body.String())
}

// angebotHandler: GET liefert das offene Angebot des Raums (oder null),
// POST bietet eine Runde an, tritt ihr bei oder zieht sie zurück.
func angebotHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	if code == "" {
		http.Error(w, "Spielraum-Code fehlt", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		// Eine Antwort für alle: die Liste der offenen Tische (leer, wenn keiner
		// steht). Wer einen bestimmten sucht, nimmt ihn aus der Liste – über die
		// Kennung oder das Spiel.
		mu.Lock()
		offen := angeboteKopie(rooms[code])
		mu.Unlock()
		jsonResponse(w, offen)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID     string `json:"id"`
		Action string `json:"action"`
		Game   string `json:"game"`
		// AngebotID nennt den Tisch. Ohne sie gilt der Tisch zum Spiel – das
		// reicht, solange je Spiel nur einer offen ist, und hält alte Aufrufer
		// am Leben.
		AngebotID string `json:"angebotId"`
		// ZielID macht aus dem Angebot eine direkte Einladung an genau diesen
		// Spieler.
		ZielID     string `json:"zielId"`
		Pairs      int    `json:"pairs"`
		Kind       string `json:"kind"`
		MaxPlayers int    `json:"maxPlayers"`
		// Zwei Spiele brauchen beim Eröffnen eine Angabe: Halma die Brettform,
		// Ship Captain Crew die Rundenzahl. Beides steht in der Spielauswahl und
		// wird hier an das Spiel durchgereicht.
		Variante string `json:"variante"`
		Anzahl   int    `json:"anzahl"`
		Runden   int    `json:"runden"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.ID = kennung(r)
	aktion := strings.ToLower(strings.TrimSpace(input.Action))

	switch aktion {
	case "bieten":
		spiel := strings.ToLower(strings.TrimSpace(input.Game))
		api, bekannt := spielApi[spiel]
		if !bekannt {
			http.Error(w, "Dieses Spiel lässt sich nicht anbieten.", http.StatusBadRequest)
			return
		}
		mu.Lock()
		room := rooms[code]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		room.cleanPlayers()
		player, ok := room.Players[input.ID]
		if !ok {
			mu.Unlock()
			http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
			return
		}
		// Je Spiel ein Tisch: Wer denselben Tisch noch einmal anbietet, ersetzt
		// seinen eigenen; ein fremder Tisch bleibt stehen (sonst würde ein
		// zweiter Spieler den ersten still verdrängen).
		if alt := angebotFuerSpiel(room, spiel); alt != nil {
			if alt.VonID != input.ID {
				mu.Unlock()
				http.Error(w, "Für dieses Spiel steht schon ein Tisch offen.", http.StatusConflict)
				return
			}
			angebotEntfernen(room, alt.ID)
		}
		// Direkte Einladung: Der Eingeladene muss da sein, sonst kommt sie nie an.
		ziel := strings.TrimSpace(input.ZielID)
		zielName := ""
		if ziel != "" {
			if ziel == input.ID {
				mu.Unlock()
				http.Error(w, "Bitte einen anderen Spieler auswählen.", http.StatusBadRequest)
				return
			}
			zielSpieler, da := room.Players[ziel]
			if !da {
				mu.Unlock()
				http.Error(w, "Der Spieler ist nicht mehr im Spielraum.", http.StatusNotFound)
				return
			}
			if !zielSpieler.anwesend(time.Now()) {
				mu.Unlock()
				http.Error(w, "Der Spieler ist gerade nicht da.", http.StatusConflict)
				return
			}
			zielName = zielSpieler.Name
		}
		room.Angebote = append(room.Angebote, &Angebot{
			ID:         randomCode(),
			Game:       spiel,
			GameName:   spielName(spiel),
			VonID:      input.ID,
			VonName:    player.Name,
			ZielID:     ziel,
			ZielName:   zielName,
			Pairs:      input.Pairs,
			Kind:       input.Kind,
			MaxPlayers: input.MaxPlayers,
			Zeit:       time.Now(),
		})
		tischID := room.Angebote[len(room.Angebote)-1].ID
		room.LastActive = time.Now()
		mu.Unlock()

		// Tischspiele sofort eröffnen: der Ersteller sitzt schon, die anderen
		// kommen über den Knopf im Chat dazu.
		if tischspiel[spiel] {
			felder := map[string]any{"id": input.ID}
			felder["spiel"] = spiel
			if input.Pairs > 0 {
				felder["pairs"] = input.Pairs
			}
			if input.Kind != "" {
				felder["kind"] = input.Kind
			}
			if input.MaxPlayers > 0 {
				felder["maxPlayers"] = input.MaxPlayers
			}
			// Kommt nichts mit (alte Seite im Zwischenspeicher, Angebot aus
			// einem alten Tab), gilt der Standard – sonst ließe sich das Spiel
			// gar nicht eröffnen.
			if spiel == "halma" {
				if input.Variante == "" {
					input.Variante = "klassisch"
				}
				if input.Anzahl == 0 {
					input.Anzahl = 2
				}
				felder["variante"] = input.Variante
				felder["anzahl"] = input.Anzahl
			}
			if spiel == "shipcrew" {
				if input.Runden == 0 {
					input.Runden = 5
				}
				felder["runden"] = input.Runden
			}
			// Mensch ärgere dich nicht: Die Sitzzahl steht in der Spielauswahl
			// (maxPlayers), das Spiel liest sie beim Eröffnen als "anzahl". Ohne
			// das Umlegen entstand jede Runde mit zwei Sitzen – egal, was im
			// Connector eingestellt war.
			if spiel == "menschaerger" {
				sitze := input.MaxPlayers
				if sitze == 0 {
					sitze = input.Anzahl
				}
				if sitze < 2 || sitze > 4 {
					sitze = 2
				}
				felder["anzahl"] = sitze
			}
			if nr, text := spielAufruf(api, code, "create", felder); nr >= 300 {
				mu.Lock()
				if room := rooms[code]; room != nil {
					angebotEntfernen(room, tischID)
				}
				mu.Unlock()
				http.Error(w, "Das Spiel lässt sich gerade nicht öffnen: "+text, http.StatusConflict)
				return
			}
		}

		mu.Lock()
		if room := rooms[code]; room != nil && angebotMitID(room, tischID) != nil {
			msg := Message{
				Name:       player.Name,
				Text:       "bietet " + spielName(spiel) + " an.",
				Time:       uhrzeit(time.Now()),
				SenderID:   input.ID,
				GameInvite: true,
				Angebot:    true,
				Game:       spiel,
				GameID:     tischID,
			}
			// Eine direkte Einladung ist ein Flüstern an genau diesen Spieler –
			// sie steht damit nur in seinem Chat und nur er kann sie annehmen.
			if ziel != "" {
				msg.Text = "lädt dich zu " + spielName(spiel) + " ein."
				msg.ToID, msg.ToName, msg.Whisper = ziel, zielName, true
			}
			room.addMessage(msg)
		}
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
		return

	case "annehmen":
		mu.Lock()
		room := rooms[code]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		room.cleanPlayers()
		player, ok := room.Players[input.ID]
		if !ok {
			mu.Unlock()
			http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
			return
		}
		tisch := angebotMitID(room, input.AngebotID)
		if tisch == nil {
			tisch = angebotFuerSpiel(room, strings.ToLower(strings.TrimSpace(input.Game)))
		}
		if tisch == nil {
			mu.Unlock()
			http.Error(w, "Es steht gerade kein Angebot offen.", http.StatusConflict)
			return
		}
		if tisch.VonID == input.ID {
			mu.Unlock()
			http.Error(w, "Du hast die Runde selbst angeboten.", http.StatusBadRequest)
			return
		}
		// Eine direkte Einladung gilt genau einem Spieler.
		if tisch.ZielID != "" && tisch.ZielID != input.ID {
			mu.Unlock()
			http.Error(w, "Dieser Tisch ist für "+tisch.ZielName+" reserviert.", http.StatusForbidden)
			return
		}
		if !room.angebotNochOffen(tisch) {
			angebotEntfernen(room, tisch.ID)
			mu.Unlock()
			http.Error(w, "Die Runde ist schon gestartet – hier ist nichts mehr offen.", http.StatusConflict)
			return
		}
		angebot := *tisch
		mu.Unlock()

		api := spielApi[angebot.Game]
		if gemeinsam[angebot.Game] {
			// Nichts zu öffnen: die Fläche liegt schon auf dem Server, jeder
			// Malt-Strich kommt dort an. Nur die Meldung im Chat fehlt noch.
		} else if tischspiel[angebot.Game] {
			if nr, text := spielAufruf(api, code, "join", map[string]any{"id": input.ID, "spiel": angebot.Game}); nr >= 300 {
				http.Error(w, text, http.StatusConflict)
				return
			}
		} else {
			felder := map[string]any{"id": angebot.VonID, "toId": input.ID}
			if angebot.Pairs > 0 {
				felder["pairs"] = angebot.Pairs
			}
			if angebot.Kind != "" {
				felder["kind"] = angebot.Kind
			}
			if angebot.MaxPlayers > 0 {
				felder["maxPlayers"] = angebot.MaxPlayers
			}
			if nr, text := spielAufruf(api, code, "invite", felder); nr >= 300 {
				http.Error(w, text, http.StatusConflict)
				return
			}
			if nr, text := spielAufruf(api, code, "accept", map[string]any{"id": input.ID}); nr >= 300 {
				http.Error(w, text, http.StatusConflict)
				return
			}
		}

		mu.Lock()
		if room := rooms[code]; room != nil {
			// Der Platz am Tisch gehört jetzt ihm: Die Karte zeigt ihn als
			// sitzend und bietet kein zweites Mitspielen mehr an.
			if tisch := angebotMitID(room, angebot.ID); tisch != nil && !enthaelt(tisch.Beigetreten, input.ID) {
				tisch.Beigetreten = append(tisch.Beigetreten, input.ID)
			}
			text := player.Name + " spielt jetzt " + angebot.GameName + " mit."
			if gemeinsam[angebot.Game] {
				text = player.Name + " malt jetzt mit."
			}
			room.addMessage(Message{
				Text:   text,
				Time:   uhrzeit(time.Now()),
				System: true,
			})
			// Spiele für zwei sind damit besetzt – der Tisch ist erledigt.
			// An Tischen mit mehreren Plätzen bleiben freie Plätze offen.
			if !tischspiel[angebot.Game] && !gemeinsam[angebot.Game] {
				angebotEntfernen(room, angebot.ID)
			}
		}
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
		return

	case "ablehnen":
		// Zu zweit ist ein Angebot eine Einladung an genau diesen einen
		// Spieler – der darf sie auch ablehnen. Sind mehr im Raum, kann
		// jemand anderes noch mitspielen; dann bleibt das Angebot stehen.
		mu.Lock()
		defer mu.Unlock()
		room := rooms[code]
		if room == nil {
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		room.cleanPlayers()
		player, ok := room.Players[input.ID]
		if !ok {
			http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
			return
		}
		tisch := angebotMitID(room, input.AngebotID)
		if tisch == nil {
			tisch = angebotFuerSpiel(room, strings.ToLower(strings.TrimSpace(input.Game)))
		}
		if tisch == nil {
			http.Error(w, "Es steht gerade kein Angebot offen.", http.StatusConflict)
			return
		}
		if tisch.VonID == input.ID {
			http.Error(w, "Du hast die Runde selbst angeboten.", http.StatusBadRequest)
			return
		}
		// Eine direkte Einladung darf nur der Eingeladene ablehnen – und sie ist
		// damit erledigt, egal wie viele sonst im Raum sind. Ein offenes Angebot
		// bleibt stehen, solange noch jemand anderes mitspielen könnte.
		if tisch.ZielID != "" {
			if tisch.ZielID != input.ID {
				http.Error(w, "Diese Einladung gilt "+tisch.ZielName+".", http.StatusForbidden)
				return
			}
		} else if room.anwesend() > 2 {
			http.Error(w, "Es sind noch andere im Raum – das Angebot bleibt offen.", http.StatusConflict)
			return
		}
		angebotEntfernen(room, tisch.ID)
		room.LastActive = time.Now()
		room.addMessage(Message{
			Text:   player.Name + " lehnt ab.",
			Time:   uhrzeit(time.Now()),
			System: true,
		})
		w.WriteHeader(http.StatusNoContent)
		return

	case "zurueck":
		mu.Lock()
		defer mu.Unlock()
		room := rooms[code]
		if room == nil {
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		player, ok := room.Players[input.ID]
		if !ok {
			http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
			return
		}
		tisch := angebotMitID(room, input.AngebotID)
		if tisch == nil {
			tisch = angebotFuerSpiel(room, strings.ToLower(strings.TrimSpace(input.Game)))
		}
		if tisch == nil || tisch.VonID != input.ID {
			http.Error(w, "Es steht kein eigenes Angebot offen.", http.StatusConflict)
			return
		}
		angebotEntfernen(room, tisch.ID)
		room.addMessage(Message{
			Text:   player.Name + " zieht das Angebot zurück.",
			Time:   uhrzeit(time.Now()),
			System: true,
		})
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Unbekannte Aktion", http.StatusBadRequest)
}
