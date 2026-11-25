package main

// Etappe 5: zentrale Bereitschaft.
//
// Bisher hatte jedes Spiel seinen eigenen Bereit-Knopf, seine eigene Merkliste
// und seine eigene Fehlermeldung; die Connector-Seite musste dafür jedes Spiel
// einzeln fragen. Jetzt gibt es eine Tür (/api/bereit) und einen Stand im
// zentralen Ruf (/api/zustand). Die Wahrheit bleibt beim Spiel: Der Connector
// schaltet über die Aktion des Spiels und liest danach, wer bereit ist –
// gestartet wird weiterhin über das Spiel, damit keine Regel doppelt gilt.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"
)

// bereitSpiele nennt die Spiele, deren Aktionstabelle ein "ready" kennt – nur
// sie können eine Bereitschaft melden, alle anderen starten direkt. Die Liste
// steht hier bewusst einmal: Die Aktionen liegen in 34 Spiel-Dateien, und die
// Karte soll schon vor dem Klick wissen, ob es etwas zu drücken gibt.
var bereitSpiele = map[string]bool{
	"arschloch":       true,
	"baccara":         true,
	"bingo":           true,
	"bomberman":       true,
	"boston":          true,
	"bubble":          true,
	"cabo":            true,
	"chicago":         true,
	"domino":          true,
	"einsxeins":       true,
	"gaensespiel":     true,
	"galgen":          true,
	"halma":           true,
	"kniffel":         true,
	"labyrinth":       true,
	"maexchen":        true,
	"maumau":          true,
	"menschaerger":    true,
	"pig":             true,
	"poker":           true,
	"pong":            true,
	"quizshow":        true,
	"risiko":          true,
	"schiffe":         true,
	"schlangen":       true,
	"schwimmen":       true,
	"sechsundsechzig": true,
	"shipcrew":        true,
	"siebzehn":        true,
	"snake":           true,
	"tron":            true,
	"uno":             true,
	"wortsuche":       true,
	"zehntausend":     true,
}

// bereitFrist: So lange gilt ein gelesener Stand, bevor er neu geholt wird.
// Der zentrale Ruf kommt alle fünf Sekunden; drei Sekunden sparen Mehrfach-
// abrufe, wenn mehrere Spieler gleichzeitig fragen, ohne dass ein Klick lange
// unsichtbar bleibt (nach jeder Aktion wird der Stand frisch geholt).
const bereitFrist = 3 * time.Second

type bereitEintrag struct {
	zeit  time.Time
	namen []string
}

var (
	bereitMu    sync.Mutex
	bereitCache = map[string]bereitEintrag{}
)

// spielStandHolen ruft den Spielstand von innen ab – mit dem Token eines
// Spielers am Tisch, so wie es die Spielseite täte.
func spielStandHolen(api, code, id string) map[string]any {
	token := ""
	mu.Lock()
	room := rooms[code]
	_, sitzt := Player{}, false
	if room != nil {
		var p Player
		p, sitzt = room.Players[id]
		token = p.Token
	}
	mu.Unlock()
	if !sitzt {
		return nil
	}
	// Ein Platz ohne Token (alte Verbindung) wird ohne Token gelesen – so
	// hält es die Spielseite auch.
	ziel := "/api/" + api + "?room=" + url.QueryEscape(code) + "&id=" + url.QueryEscape(id)
	if token != "" {
		ziel += "&token=" + url.QueryEscape(token)
	}
	rec := httptest.NewRecorder()
	routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ziel, nil))
	if rec.Code != http.StatusOK {
		return nil
	}
	var stand map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &stand); err != nil {
		return nil
	}
	return stand
}

// bereitAusStand liest aus einem Spielstand, wer bereit ist. Die Spiele nennen
// ihre Sitze je nach Alter "players", "spieler" oder "seats", bereit heißt
// "ready" oder "bereit" – beides steht so in ihren Antworten.
func bereitAusStand(stand map[string]any) []string {
	namen := []string{}
	if stand == nil {
		return namen
	}
	// Die Spiel-Endpunkte liefern ihren Stand unter "game".
	if spiel, ok := stand["game"].(map[string]any); ok {
		stand = spiel
	}
	for _, feld := range []string{"players", "spieler", "seats"} {
		liste, ok := stand[feld].([]any)
		if !ok {
			continue
		}
		for _, eintrag := range liste {
			sitz, ok := eintrag.(map[string]any)
			if !ok {
				continue
			}
			bereit, hat := sitz["ready"].(bool)
			if !hat {
				bereit, hat = sitz["bereit"].(bool)
			}
			if !hat || !bereit {
				continue
			}
			if id, ok := sitz["id"].(string); ok && id != "" {
				namen = append(namen, id)
			}
		}
	}
	return namen
}

// bereitFuer liefert die Bereiten eines Spiels, kurz zwischengespeichert.
// leser ist ein Spieler am Tisch; sein Token öffnet den Spielstand.
func bereitFuer(code, spiel, leser string) []string {
	schluessel := code + "|" + spiel
	bereitMu.Lock()
	if e, da := bereitCache[schluessel]; da && time.Since(e.zeit) < bereitFrist {
		namen := e.namen
		bereitMu.Unlock()
		return namen
	}
	bereitMu.Unlock()

	namen := []string{}
	if api, bekannt := spielApi[spiel]; bekannt && leser != "" {
		namen = bereitAusStand(spielStandHolen(api, code, leser))
	}
	bereitMu.Lock()
	bereitCache[schluessel] = bereitEintrag{zeit: time.Now(), namen: namen}
	bereitMu.Unlock()
	return namen
}

// bereitVergessen wirft den Stand eines Spiels weg (nach einer Aktion).
func bereitVergessen(code, spiel string) {
	bereitMu.Lock()
	delete(bereitCache, code+"|"+spiel)
	bereitMu.Unlock()
}

// bereitHandler: POST /api/bereit?room=…&id=… mit {"game":"kniffel"} schaltet
// die eigene Bereitschaft um und antwortet mit dem frischen Stand.
func bereitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erforderlich", http.StatusMethodNotAllowed)
		return
	}
	id := kennung(r)
	code := raumAusKontext(r)
	if id == "" || code == "" {
		http.Error(w, "Spielraum oder Spielerkennung fehlt", http.StatusBadRequest)
		return
	}
	var input struct {
		Game string `json:"game"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	spiel := strings.ToLower(strings.TrimSpace(input.Game))
	api, bekannt := spielApi[spiel]
	if !bekannt {
		http.Error(w, "Dieses Spiel lässt sich nicht anbieten.", http.StatusBadRequest)
		return
	}
	mu.Lock()
	room := rooms[code]
	sitzt := false
	if room != nil {
		_, sitzt = room.Players[id]
	}
	mu.Unlock()
	if room == nil {
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	if !sitzt {
		http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
		return
	}
	// Die Aktion des Spiels schaltet um – sie kennt die Sitze und die Phase.
	if nr, text := spielAufruf(api, code, "ready", map[string]any{"id": id}); nr >= 300 {
		// Spiele ohne Bereit-Knopf antworten mit "Unbekannte …-Aktion", weil ihre
		// Aktionstabelle kein "ready" kennt. Das ist kein Fehler des Spielers.
		if strings.Contains(text, "Unbekannte") {
			http.Error(w, "Dieses Spiel braucht kein Bereit – es startet direkt.", http.StatusConflict)
			return
		}
		if nr < 400 {
			nr = http.StatusConflict
		}
		http.Error(w, text, nr)
		return
	}
	bereitVergessen(code, spiel)
	jsonResponse(w, map[string]any{"game": spiel, "bereit": bereitFuer(code, spiel, id)})
}
