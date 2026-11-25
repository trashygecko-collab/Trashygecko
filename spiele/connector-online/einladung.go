package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Ein privater Raum entsteht nur auf Einladung. Er hat einen Code wie jeder
// Raum, aber niemand sieht ihn: Der Server gibt ihn allein an die beiden
// Programme weiter, die ihn brauchen. Der gemeinsame Hauptraum bleibt davon
// unberührt – dort laufen Chat, Angebote und Partien weiter.
type Einladung struct {
	VonID    string    `json:"vonId"`
	VonName  string    `json:"vonName"`
	ZielID   string    `json:"zielId"`
	ZielName string    `json:"zielName"`
	Raum     string    `json:"raum,omitempty"`
	Stand    string    `json:"stand"` // offen | angenommen | abgelehnt
	Zeit     time.Time `json:"-"`
}

// einladungen hängt am Ziel: eine Einladung je Spieler, eine neue ersetzt die
// alte. Der Schlüssel ist die Kennung des Eingeladenen.
var einladungen = map[string]*Einladung{}

// Zwei Minuten reichen zum Antworten; länger soll niemand auf eine offene
// Einladung warten.
const einladungFrist = 2 * time.Minute

// einladungAufraeumen wirft alte Einladungen weg und räumt den privaten Raum
// ab, den niemand mehr betreten hat. Muss unter mu laufen.
func einladungAufraeumen(jetzt time.Time) {
	for zielID, e := range einladungen {
		if jetzt.Sub(e.Zeit) <= einladungFrist {
			continue
		}
		if e.Raum != "" && e.Raum != mainRoomCode {
			if raum := rooms[e.Raum]; raum != nil && len(raum.Players) == 0 {
				delete(rooms, e.Raum)
			}
		}
		delete(einladungen, zielID)
	}
}

// privaterRaum legt den Raum für genau diese zwei Spieler an. Er startet leer;
// die Sitze entstehen, sobald die beiden ihn betreten (wie in jedem Raum).
func privaterRaum(vonVon, vonBis string) *Room {
	mu.Lock()
	defer mu.Unlock()
	code := randomCode()
	for rooms[code] != nil || code == mainRoomCode {
		code = randomCode()
	}
	rooms[code] = &Room{
		Code:       code,
		Privat:     true,
		OwnerID:    vonVon,
		OwnerName:  vonBis,
		Name:       "Privater Raum",
		Created:    time.Now(),
		LastActive: time.Now(),
		Players:    map[string]Player{},
		Messages:   []Message{},
		Mutes:      map[string]time.Time{},
		MsgTimes:   map[string][]time.Time{},
	}
	return rooms[code]
}

// einladungenFuer liefert die Einladungen dieses Spielers: die eine offene an
// ihn und alle, die er selbst verschickt hat. Muss unter mu laufen, nachdem
// einladungAufraeumen gelaufen ist.
func einladungenFuer(ich string) ([]Einladung, []Einladung) {
	eingehend := []Einladung{}
	ausgehend := []Einladung{}
	if ich == "" {
		return eingehend, ausgehend
	}
	if e := einladungen[ich]; e != nil && e.Stand == "offen" {
		eingehend = append(eingehend, *e)
	}
	for _, e := range einladungen {
		if e.VonID == ich {
			ausgehend = append(ausgehend, *e)
		}
	}
	return eingehend, ausgehend
}

// einladungHandler: GET liefert meine Einladungen (eingehend und ausgehend),
// POST sendet, nimmt an, lehnt ab oder führt zurück in den Hauptraum.
func einladungHandler(w http.ResponseWriter, r *http.Request) {
	ich := kennung(r)
	if ich == "" {
		http.Error(w, "Spielerkennung fehlt", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		mu.Lock()
		einladungAufraeumen(time.Now())
		eingehend, ausgehend := einladungenFuer(ich)
		mu.Unlock()
		jsonResponse(w, map[string]any{"eingehend": eingehend, "ausgehend": ausgehend})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var in struct {
		Action string `json:"action"`
		ZielID string `json:"zielId"`
		VonID  string `json:"vonId"`
		Raum   string `json:"raum"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}

	switch strings.ToLower(strings.TrimSpace(in.Action)) {
	case "senden":
		ziel := strings.TrimSpace(in.ZielID)
		if ziel == "" || ziel == ich {
			http.Error(w, "Bitte einen anderen Spieler auswählen.", http.StatusBadRequest)
			return
		}
		mu.Lock()
		einladungAufraeumen(time.Now())
		// Beide müssen im selben Raum sein – der Einladende hier, der
		// Eingeladene ebenfalls.
		raum := rooms[mainRoomCode]
		ein := Player{}
		da := false
		if raum != nil {
			ein, da = raum.Players[ich]
		}
		zielSpieler, zielDa := Player{}, false
		if raum != nil {
			zielSpieler, zielDa = raum.Players[ziel]
		}
		mu.Unlock()
		if !da || !zielDa {
			http.Error(w, "Der Spieler ist nicht mehr im Spielraum.", http.StatusNotFound)
			return
		}
		// Ein Platz, dessen Browser längst zu ist, steht noch eine Weile in der
		// Liste. Eine Einladung dorthin käme nie an – also gleich sagen.
		if !zielSpieler.anwesend(time.Now()) {
			http.Error(w, "Der Spieler ist gerade nicht da.", http.StatusConflict)
			return
		}
		priv := privaterRaum(ich, ein.Name)
		mu.Lock()
		einladungen[ziel] = &Einladung{
			VonID:    ich,
			VonName:  ein.Name,
			ZielID:   ziel,
			ZielName: zielSpieler.Name,
			Raum:     priv.Code,
			Stand:    "offen",
			Zeit:     time.Now(),
		}
		mu.Unlock()
		jsonResponse(w, map[string]string{"stand": "offen"})

	case "annehmen":
		mu.Lock()
		einladungAufraeumen(time.Now())
		e := einladungen[ich]
		if e == nil || e.Stand != "offen" || (in.VonID != "" && e.VonID != in.VonID) {
			mu.Unlock()
			http.Error(w, "Diese Einladung gibt es nicht mehr.", http.StatusNotFound)
			return
		}
		raum := rooms[e.Raum]
		if raum == nil {
			delete(einladungen, ich)
			mu.Unlock()
			http.Error(w, "Der private Raum ist nicht mehr da.", http.StatusNotFound)
			return
		}
		e.Stand = "angenommen"
		raum.LastActive = time.Now()
		antwort := map[string]string{"raum": e.Raum, "vonName": e.VonName, "vonId": e.VonID}
		mu.Unlock()
		jsonResponse(w, antwort)

	case "ablehnen":
		mu.Lock()
		e := einladungen[ich]
		if e != nil {
			e.Stand = "abgelehnt"
			if raum := rooms[e.Raum]; raum != nil && len(raum.Players) == 0 {
				delete(rooms, e.Raum)
			}
		}
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)

	case "zurueck":
		// Zurück in den gemeinsamen Raum: Platz im privaten Raum freigeben und
		// die Einladung schließen. Die Rückkehr selbst ist ein normales
		// Betreten des Hauptraums.
		mu.Lock()
		code := strings.TrimSpace(in.Raum)
		if code == "" || code == mainRoomCode {
			for _, e := range einladungen {
				if e.VonID == ich || e.ZielID == ich {
					code = e.Raum
					break
				}
			}
		}
		if code != "" && code != mainRoomCode {
			if raum := rooms[code]; raum != nil {
				delete(raum.Players, ich)
				raum.cleanPlayers()
				raum.LastActive = time.Now()
				if len(raum.Players) == 0 {
					delete(rooms, code)
				}
			}
		}
		for ziel, e := range einladungen {
			if e.VonID == ich || e.ZielID == ich {
				delete(einladungen, ziel)
			}
		}
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Unbekannte Aktion", http.StatusBadRequest)
	}
}
