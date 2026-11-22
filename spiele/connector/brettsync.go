package main

// Online-Mehrspieler für Das verrückte Labyrinth (Risiko hat seit dem
// 22.11.2026 ein eigenes, servergeprüftes Spiel in risiko.go/risiko_raum.go).
//
// Beide Spiele rechnen ihre Regeln im Browser (dort stehen Karten, Armeen,
// Kacheln, Missionen – dieselben Regeln wie im Einzelspiel). Der Server führt
// den verbindlichen Stand:
//   - wer am Tisch sitzt und in welcher Reihenfolge,
//   - wer am Zug ist (nur dieser darf einen Stand schicken),
//   - den letzten gültigen Spielstand mit Versionsnummer.
// Jeder Zug kommt als vollständiger Stand an, bekommt eine neue Version und
// wird allen Spielern beim nächsten Abruf geliefert. Damit sind Zugreihenfolge,
// Wiederverbinden und Neustart abgedeckt, ohne die Regeln doppelt zu pflegen
// (die Spiele bleiben die einzige Quelle der Regeln).

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SyncSpieler ist ein Sitz am Tisch.
type SyncSpieler struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Platz  int    `json:"platz"`
	Bereit bool   `json:"bereit"`
}

// SyncGame ist eine Partie eines Spiels mit verteiltem Spielstand.
type SyncGame struct {
	ID      string          `json:"id"`
	Status  string          `json:"status"` // invited | active | finished
	Spieler []SyncSpieler   `json:"spieler"`
	Anzahl  int             `json:"anzahl"`
	Current int             `json:"current"`
	Sieger  int             `json:"sieger"`
	ZugNr   int             `json:"zugNr"`
	Stand   json.RawMessage `json:"stand"`
	Version int             `json:"version"`
	Note    string          `json:"note"`
	HostID  string          `json:"-"`
}

const syncStandMax = 256 * 1024 // ein Stand bleibt weit darunter

type syncArt struct {
	Name     string
	Min, Max int
}

var syncArten = map[string]syncArt{
	"labyrinth": {Name: "Das verrückte Labyrinth", Min: 2, Max: 4},
}

// syncFeld liefert den Platz im Raum, an dem die Partie dieser Art liegt.
func syncFeld(room *Room, art string) **SyncGame {
	return &room.Labyrinth
}

func syncSicht(g *SyncGame) *SyncGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Spieler = append([]SyncSpieler{}, g.Spieler...)
	if len(g.Stand) > 0 {
		c.Stand = append(json.RawMessage{}, g.Stand...)
	}
	return &c
}

func syncPlatzVon(g *SyncGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// syncIDs listet die Kennungen der Sitzenden – für leave.go.
func syncIDs(g *SyncGame) []string {
	if g == nil {
		return nil
	}
	ids := make([]string, 0, len(g.Spieler))
	for _, p := range g.Spieler {
		ids = append(ids, p.ID)
	}
	return ids
}

func syncName(art string) string {
	if a, ok := syncArten[art]; ok {
		return a.Name
	}
	return art
}

func labyrinthHandler(w http.ResponseWriter, r *http.Request) { syncHandler(w, r, "labyrinth") }

// syncHandler bedient beide Spiele: ?room=…&action=… (POST) bzw. GET.
func syncHandler(w http.ResponseWriter, r *http.Request, art string) {
	regeln, bekannt := syncArten[art]
	if !bekannt {
		http.Error(w, "Unbekanntes Spiel", http.StatusBadRequest)
		return
	}
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	if code == "" {
		http.Error(w, "Spielraum-Code fehlt", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		mu.Lock()
		room := rooms[code]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		if _, ok := room.Players[id]; !ok {
			mu.Unlock()
			http.Error(w, "Bitte zuerst dem Spielraum beitreten.", http.StatusForbidden)
			return
		}
		sicht := syncSicht(*syncFeld(room, art))
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": sicht})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID     string          `json:"id"`
		Anzahl int             `json:"anzahl"`
		Weiter int             `json:"weiter"`
		Ende   bool            `json:"ende"`
		Stand  json.RawMessage `json:"stand"`
	}
	roh, err := io.ReadAll(io.LimitReader(r.Body, syncStandMax+4096))
	if err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(roh, &input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.ID = strings.TrimSpace(input.ID)
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

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
	fehler := func(status int, text string) {
		mu.Unlock()
		http.Error(w, text, status)
	}
	feld := syncFeld(room, art)
	g := *feld

	switch action {
	case "create":
		if g != nil && g.Status != "finished" {
			fehler(http.StatusConflict, "In diesem Spielraum läuft bereits eine "+regeln.Name+"-Partie.")
			return
		}
		anzahl := input.Anzahl
		if anzahl < regeln.Min || anzahl > regeln.Max {
			anzahl = regeln.Min
		}
		neu := &SyncGame{
			ID: randomCode(), Status: "invited", Anzahl: anzahl, Current: 0,
			Sieger: -1, ZugNr: 0, Version: 1, HostID: input.ID,
			Spieler: []SyncSpieler{{ID: input.ID, Name: player.Name, Platz: 0}},
			Note:    player.Name + " eröffnet " + regeln.Name + ".",
		}
		*feld = neu
		room.addMessage(Message{
			Name: player.Name, Text: "eröffnet " + regeln.Name + ".",
			Time: time.Now().Format("15:04:05"), SenderID: input.ID,
		})

	// join und rejoin tun dasselbe: Platz nehmen, solange einer frei ist –
	// rejoin auch in einer laufenden Partie (Handy gesperrt, Neuladen).
	case "join", "rejoin":
		if g == nil {
			fehler(http.StatusConflict, "Es läuft gerade keine "+regeln.Name+"-Partie.")
			return
		}
		if g.Status == "finished" {
			fehler(http.StatusConflict, "Diese Partie ist beendet.")
			return
		}
		if syncPlatzVon(g, input.ID) >= 0 {
			break
		}
		if action == "join" && g.Status != "invited" {
			fehler(http.StatusConflict, "Die Partie läuft schon.")
			return
		}
		if len(g.Spieler) >= g.Anzahl {
			fehler(http.StatusConflict, "Diese Partie ist schon voll.")
			return
		}
		frei := -1
		for p := 0; p < g.Anzahl; p++ {
			belegt := false
			for _, s := range g.Spieler {
				if s.Platz == p {
					belegt = true
					break
				}
			}
			if !belegt {
				frei = p
				break
			}
		}
		if frei < 0 {
			fehler(http.StatusConflict, "Diese Partie ist schon voll.")
			return
		}
		g.Spieler = append(g.Spieler, SyncSpieler{ID: input.ID, Name: player.Name, Platz: frei})
		if action == "join" {
			g.Note = player.Name + " spielt mit."
		} else {
			g.Note = player.Name + " ist wieder dabei."
		}

	case "ready":
		if g == nil || g.Status != "invited" {
			fehler(http.StatusConflict, "Bereit drücken geht nur vor dem Start.")
			return
		}
		k := syncPlatzVon(g, input.ID)
		if k < 0 {
			fehler(http.StatusForbidden, "Du sitzt nicht an diesem Tisch.")
			return
		}
		g.Spieler[k].Bereit = !g.Spieler[k].Bereit

	case "start":
		if g == nil || g.Status != "invited" {
			fehler(http.StatusConflict, "Es wartet keine Partie auf den Start.")
			return
		}
		if syncPlatzVon(g, input.ID) != 0 {
			fehler(http.StatusForbidden, "Nur der Gastgeber startet die Partie.")
			return
		}
		if len(g.Spieler) < regeln.Min {
			fehler(http.StatusConflict, regeln.Name+" braucht mindestens "+strconv.Itoa(regeln.Min)+" Spieler.")
			return
		}
		for _, p := range g.Spieler {
			if !p.Bereit {
				fehler(http.StatusConflict, p.Name+" ist noch nicht bereit.")
				return
			}
		}
		g.Status = "active"
		g.Current = 0
		g.Sieger = -1
		g.ZugNr = 1
		g.Stand = nil // die erste Seite baut die Aufstellung und schickt sie
		g.Version++
		g.Note = "Los geht's – " + g.Spieler[0].Name + " beginnt."

	// zug und stand nehmen einen vollständigen Spielstand an. Nur wer am Zug
	// ist, darf ihn schicken; die Version verhindert, dass ein alter Stand
	// einen neueren überschreibt.
	case "zug", "stand":
		if g == nil || g.Status != "active" {
			fehler(http.StatusConflict, "Es läuft gerade keine aktive Partie.")
			return
		}
		k := syncPlatzVon(g, input.ID)
		if k < 0 {
			fehler(http.StatusForbidden, "Du gehörst nicht zu dieser Partie.")
			return
		}
		if g.Current != k {
			fehler(http.StatusConflict, "Jetzt ist "+g.Spieler[g.Current].Name+" am Zug.")
			return
		}
		if len(input.Stand) == 0 || len(input.Stand) > syncStandMax {
			fehler(http.StatusBadRequest, "Der Spielstand fehlt oder ist zu groß.")
			return
		}
		var probe map[string]any
		if err := json.Unmarshal(input.Stand, &probe); err != nil {
			fehler(http.StatusBadRequest, "Der Spielstand ist kein gültiges JSON.")
			return
		}
		weiter := input.Weiter
		if weiter < 0 || weiter >= len(g.Spieler) {
			weiter = (k + 1) % len(g.Spieler)
		}
		g.Stand = append(json.RawMessage{}, input.Stand...)
		g.Current = weiter
		g.Version++
		g.Note = ""
		if action == "stand" {
			g.Note = "Aufstellung steht."
		} else {
			g.ZugNr++
		}
		if input.Ende {
			g.Status = "finished"
			g.Sieger = weiter
			g.Note = g.Spieler[weiter].Name + " gewinnt."
		}

	case "neu":
		if g == nil || (g.Status != "finished" && g.Status != "active") {
			fehler(http.StatusConflict, "Eine neue Runde geht erst nach dem Start.")
			return
		}
		if syncPlatzVon(g, input.ID) < 0 {
			fehler(http.StatusForbidden, "Du gehörst nicht zu dieser Partie.")
			return
		}
		// Dieselben Sitze, neuer Stand: die Reihenfolge beginnt wieder vorn.
		for i := range g.Spieler {
			g.Spieler[i].Bereit = false
		}
		g.Status = "invited"
		g.Stand = nil
		g.Current = 0
		g.Sieger = -1
		g.ZugNr = 0
		g.Version++
		g.Note = "Neue Runde – bitte bereit drücken."

	default:
		fehler(http.StatusBadRequest, "Unbekannte Aktion.")
		return
	}

	room.LastActive = time.Now()
	result := syncSicht(*feld)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

// syncFehlerText macht aus einem Fehler einen kurzen Satz für den Client.
func syncFehlerText(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, io.EOF) {
		return "Keine Antwort"
	}
	return err.Error()
}
