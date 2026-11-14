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

// Angebot ist eine offene Runde: Der Ersteller bietet ein Spiel für den ganzen
// Raum an. Jeder sieht die Karte im Chat und tritt mit einem Klick bei – es
// braucht keine Einladung an einzelne Spieler mehr.
type Angebot struct {
	ID         string `json:"id"`
	Game       string `json:"game"`
	GameName   string `json:"gameName"`
	VonID      string `json:"vonId"`
	VonName    string `json:"vonName"`
	Pairs      int    `json:"pairs,omitempty"`
	Kind       string `json:"kind,omitempty"`
	MaxPlayers int    `json:"maxPlayers,omitempty"`
}

// spielApi ordnet der Kennung aus der Spieleliste den API-Pfad zu.
var spielApi = map[string]string{
	"tictactoe":        "tictactoe",
	"viergewinnt":      "viergewinnt",
	"dame":             "dame",
	"paare":            "paare",
	"menschaerger":     "menschaerger",
	"kaesekaestche":    "kaesekaestche",
	"sim":              "sim",
	"kalaha":           "kalaha",
	"streichholzspiel": "streichholz",
	"muehle":           "muehle",
	"reversi":          "reversi",
	"hold":             "hold",
	"skyjo":            "skyjo",
	"kniffel":          "kniffel",
	"schach":           "schach",
	"go":               "go",
	"backgammon":       "backgammon",
	"maumau":           "maumau",
	"uno":              "uno",
	"schwimmen":        "schwimmen",
	"arschloch":        "arschloch",
	"einsxeins":        "einsxeins",
}

// spielAnzeige sind die Namen, wie sie im Chat stehen.
var spielAnzeige = map[string]string{
	"tictactoe":        "Neon Tic Tac Toe",
	"viergewinnt":      "Vier gewinnt",
	"dame":             "Dame",
	"paare":            "Paare finden",
	"menschaerger":     "Mensch ärgere dich nicht",
	"kaesekaestche":    "Käsekästchen",
	"sim":              "Sim",
	"kalaha":           "Kalaha",
	"streichholzspiel": "21 Streichhölzer",
	"muehle":           "Mühle",
	"reversi":          "Reversi",
	"hold":             "Hold the Line",
	"skyjo":            "Skyjo",
	"kniffel":          "Kniffel",
	"schach":           "Schach",
	"go":               "Go",
	"backgammon":       "Backgammon",
	"maumau":           "Mau Mau",
	"uno":              "UNO",
	"schwimmen":        "31 Schwimmen",
	"arschloch":        "Arschloch",
	"einsxeins":        "1×1-Rennen",
}

// Tischspiele: dort sitzen mehrere, die Runde wird beim Anbieten eröffnet.
var tischspiel = map[string]bool{
	"kniffel":   true,
	"maumau":    true,
	"uno":       true,
	"schwimmen": true,
	"arschloch": true,
	"einsxeins": true,
}

func spielName(id string) string {
	if name, ok := spielAnzeige[id]; ok {
		return name
	}
	return id
}

// spielAufruf ruft einen Spiel-Endpunkt von innen auf – wie ein Klick eines
// Spielers. So gelten beim Beitritt genau dieselben Prüfungen wie sonst.
func spielAufruf(api, code, aktion string, felder map[string]any) (int, string) {
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
		mu.Lock()
		var angebot *Angebot
		if room := rooms[code]; room != nil && room.Angebot != nil {
			kopie := *room.Angebot
			angebot = &kopie
		}
		mu.Unlock()
		jsonResponse(w, angebot)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID         string `json:"id"`
		Action     string `json:"action"`
		Game       string `json:"game"`
		Pairs      int    `json:"pairs"`
		Kind       string `json:"kind"`
		MaxPlayers int    `json:"maxPlayers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.ID = strings.TrimSpace(input.ID)
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
		room.Angebot = &Angebot{
			ID:         randomCode(),
			Game:       spiel,
			GameName:   spielName(spiel),
			VonID:      input.ID,
			VonName:    player.Name,
			Pairs:      input.Pairs,
			Kind:       input.Kind,
			MaxPlayers: input.MaxPlayers,
		}
		room.LastActive = time.Now()
		mu.Unlock()

		// Tischspiele sofort eröffnen: der Ersteller sitzt schon, die anderen
		// kommen über den Knopf im Chat dazu.
		if tischspiel[spiel] {
			felder := map[string]any{"id": input.ID}
			if input.Pairs > 0 {
				felder["pairs"] = input.Pairs
			}
			if input.Kind != "" {
				felder["kind"] = input.Kind
			}
			if input.MaxPlayers > 0 {
				felder["maxPlayers"] = input.MaxPlayers
			}
			if nr, text := spielAufruf(api, code, "create", felder); nr >= 300 {
				mu.Lock()
				if room := rooms[code]; room != nil && room.Angebot != nil && room.Angebot.VonID == input.ID {
					room.Angebot = nil
				}
				mu.Unlock()
				http.Error(w, "Das Spiel lässt sich gerade nicht öffnen: "+text, http.StatusConflict)
				return
			}
		}

		mu.Lock()
		if room := rooms[code]; room != nil && room.Angebot != nil {
			room.addMessage(Message{
				Name:       player.Name,
				Text:       "bietet " + spielName(spiel) + " an.",
				Time:       time.Now().Format("15:04:05"),
				SenderID:   input.ID,
				GameInvite: true,
				Game:       spiel,
				GameID:     room.Angebot.ID,
			})
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
		if room.Angebot == nil {
			mu.Unlock()
			http.Error(w, "Es steht gerade kein Angebot offen.", http.StatusConflict)
			return
		}
		if room.Angebot.VonID == input.ID {
			mu.Unlock()
			http.Error(w, "Du hast die Runde selbst angeboten.", http.StatusBadRequest)
			return
		}
		angebot := *room.Angebot
		mu.Unlock()

		api := spielApi[angebot.Game]
		if tischspiel[angebot.Game] {
			if nr, text := spielAufruf(api, code, "join", map[string]any{"id": input.ID}); nr >= 300 {
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
			room.addMessage(Message{
				Text:   player.Name + " spielt jetzt " + angebot.GameName + " mit.",
				Time:   time.Now().Format("15:04:05"),
				System: true,
			})
			// Spiele für zwei sind damit besetzt – das Angebot ist erledigt.
			// An Tischen bleiben freie Plätze offen.
			if !tischspiel[angebot.Game] && room.Angebot != nil && room.Angebot.ID == angebot.ID {
				room.Angebot = nil
			}
		}
		mu.Unlock()
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
		if room.Angebot == nil || room.Angebot.VonID != input.ID {
			http.Error(w, "Es steht kein eigenes Angebot offen.", http.StatusConflict)
			return
		}
		room.Angebot = nil
		room.addMessage(Message{
			Text:   player.Name + " zieht das Angebot zurück.",
			Time:   time.Now().Format("15:04:05"),
			System: true,
		})
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Unbekannte Aktion", http.StatusBadRequest)
}
