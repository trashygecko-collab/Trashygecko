package main

// Risiko im Raum (Phase 2, 22.11.2026): Lobby, geprüfte Aktionen, Bot-Takt,
// dauerhafte Spielstände.
//
// Die Handler-Funktion nimmt nur Absichten entgegen („greife X an“). Alles
// Weitere – Reihenfolge, Besitz, Würfel, Karten, Sieg – entscheidet
// risiko.go. Der Spielstand wird nach jeder Änderung als JSON gespeichert und
// beim Start des Connectors wieder geladen, damit eine Partie einen Neustart
// überlebt.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RisikoEinstellung beschreibt den Aufbau einer Partie.
type RisikoEinstellung struct {
	Modus  string
	Anzahl int
	Namen  []string
	IDs    []string
	Bots   []bool
	Stufen []string
	Wesen  []string
	Pause  string
}

// risikoStaendeOrdner liegt neben der Exe; Tests setzen ihn auf ein
// Wegwerf-Verzeichnis.
var risikoStaendeOrdner = "risiko-staende"

func risikoTakt(pause string) time.Duration {
	switch pause {
	case "schnell":
		return 150 * time.Millisecond
	case "langsam":
		return 900 * time.Millisecond
	}
	return 400 * time.Millisecond
}

// ---------- Lobby ----------------------------------------------------------

func risikoNeueLobby(hostID, hostName string, anzahl int) *RisikoSpiel {
	if anzahl < risikoMinSpieler {
		anzahl = risikoMinSpieler
	}
	if anzahl > risikoMaxSpieler {
		anzahl = risikoMaxSpieler
	}
	return &RisikoSpiel{
		ID:      randomCode(),
		Anzahl:  anzahl,
		Modus:   "welt",
		Status:  risikoStatusLobby,
		Sieger:  -1,
		Version: 1,
		Pause:   "normal",
		HostID:  hostID,
		Spieler: []RisikoSpieler{{
			ID: hostID, Name: hostName, Farbe: risikoFarben[0], Team: -1, Bereit: true,
		}},
		Gebiete: map[string]*RisikoGebiet{},
	}
}

func risikoSitzVon(g *RisikoSpiel, id string) int {
	if g == nil {
		return -1
	}
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			return i
		}
	}
	return -1
}

// ---------- Handler --------------------------------------------------------

type risikoEingabe struct {
	ID     string `json:"id"`
	Anzahl int    `json:"anzahl"`
	Modus  string `json:"modus"`
	Stufe  string `json:"stufe"`
	Wesen  string `json:"wesen"`
	Platz  int    `json:"platz"`
	Gebiet string `json:"gebiet"`
	Von    string `json:"von"`
	Nach   string `json:"nach"`
	Armeen int    `json:"armeen"`
	Karten []int  `json:"karten"`
	Pause  string `json:"pause"`
}

func risikoSpielHandler(w http.ResponseWriter, r *http.Request) {
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
		p, ok := room.Players[id]
		if !ok {
			mu.Unlock()
			http.Error(w, "Bitte zuerst dem Spielraum beitreten.", http.StatusForbidden)
			return
		}
		// Wer den Stand abruft, ist da: ohne dieses Lebenszeichen würde der
		// Sitz nach 150 s verfallen und die Partie verschwinden.
		p.Last = time.Now()
		room.Players[id] = p
		risikoNamenAuffrischen(room)
		risikoBotTreiben(room, time.Now())
		sicht := risikoSichtFuer(room.Risiko, id)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": sicht})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input risikoEingabe
	roh, err := io.ReadAll(http.MaxBytesReader(w, r.Body, risikoStandMax))
	if err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	if len(strings.TrimSpace(string(roh))) > 0 {
		if err := json.Unmarshal(roh, &input); err != nil {
			http.Error(w, "Ungültige Daten", http.StatusBadRequest)
			return
		}
	}
	input.ID = strings.TrimSpace(input.ID)
	// Die Aktionsnamen sind gemischt geschrieben (aufstellenFertig);
	// ein ToLower würde sie unkenntlich machen.
	action := strings.TrimSpace(r.URL.Query().Get("action"))

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
	player.Last = time.Now()
	room.Players[input.ID] = player
	risikoNamenAuffrischen(room)
	jetzt := time.Now()

	fehler := func(status int, err error) {
		mu.Unlock()
		http.Error(w, err.Error(), status)
	}

	g := room.Risiko

	switch action {
	case "create":
		if g != nil && g.Status != risikoStatusBeendet {
			fehler(http.StatusConflict, errors.New("In diesem Spielraum läuft bereits eine Risiko-Partie."))
			return
		}
		if input.Anzahl >= risikoMinSpieler && input.Anzahl <= risikoMaxSpieler {
			// Anzahl sitzt in der Lobby, nicht im Spielstand.
		}
		lobby := risikoNeueLobby(input.ID, player.Name, input.Anzahl)
		if _, bekannt := risikoModi[input.Modus]; bekannt {
			lobby.Modus = input.Modus
		}
		lobby.Pause = risikoPauseWert(input.Pause)
		room.Risiko = lobby
		room.botNamenNeu()
		room.addMessage(Message{Name: player.Name, Text: "eröffnet Risiko.",
			Time: jetzt.Format("15:04:05"), SenderID: input.ID})

	case "join", "rejoin":
		if g == nil {
			fehler(http.StatusConflict, errors.New("Es läuft gerade keine Risiko-Partie."))
			return
		}
		if risikoSitzVon(g, input.ID) >= 0 {
			break
		}
		if g.Status == risikoStatusBeendet {
			fehler(http.StatusConflict, errors.New("Diese Partie ist beendet."))
			return
		}
		if g.Status != risikoStatusLobby || action == "join" && g.Status != risikoStatusLobby {
			fehler(http.StatusConflict, errors.New("Die Partie läuft schon."))
			return
		}
		if len(g.Spieler) >= g.Anzahl {
			fehler(http.StatusConflict, errors.New("Der Tisch ist voll."))
			return
		}
		g.Spieler = append(g.Spieler, RisikoSpieler{
			ID: input.ID, Name: player.Name, Farbe: risikoFarben[len(g.Spieler)%len(risikoFarben)],
			Team: -1,
		})
		g.Version++

	case "ready":
		if g == nil || g.Status != risikoStatusLobby {
			fehler(http.StatusConflict, errors.New("Bereit drücken geht nur in der Lobby."))
			return
		}
		k := risikoSitzVon(g, input.ID)
		if k < 0 {
			fehler(http.StatusForbidden, errors.New("Du sitzt nicht an diesem Tisch."))
			return
		}
		g.Spieler[k].Bereit = !g.Spieler[k].Bereit
		g.Version++

	case "modus":
		if g == nil || g.Status != risikoStatusLobby || input.ID != g.HostID {
			fehler(http.StatusForbidden, errors.New("Nur der Gastgeber stellt die Partie ein."))
			return
		}
		if _, bekannt := risikoModi[input.Modus]; bekannt {
			g.Modus = input.Modus
			g.Version++
		}

	case "botadd":
		if g == nil || g.Status != risikoStatusLobby || input.ID != g.HostID {
			fehler(http.StatusForbidden, errors.New("Nur der Gastgeber fügt Bots hinzu."))
			return
		}
		if len(g.Spieler) >= g.Anzahl {
			fehler(http.StatusConflict, errors.New("Der Tisch ist voll."))
			return
		}
		n := 1
		for _, sp := range g.Spieler {
			if sp.Bot {
				n++
			}
		}
		name := room.botName(n)
		g.Spieler = append(g.Spieler, RisikoSpieler{
			ID: botID(n), Name: name, Farbe: risikoFarben[len(g.Spieler)%len(risikoFarben)],
			Bot: true, Stufe: risikoBotStufe(input.Stufe), Wesen: risikoBotWesen(input.Wesen),
			Team: -1, Bereit: true,
		})
		g.Version++

	case "botremove":
		if g == nil || g.Status != risikoStatusLobby || input.ID != g.HostID {
			fehler(http.StatusForbidden, errors.New("Nur der Gastgeber entfernt Bots."))
			return
		}
		if input.Platz < 0 || input.Platz >= len(g.Spieler) || !g.Spieler[input.Platz].Bot {
			fehler(http.StatusBadRequest, errors.New("Dort sitzt kein Bot."))
			return
		}
		g.Spieler = append(g.Spieler[:input.Platz], g.Spieler[input.Platz+1:]...)
		g.Version++

	case "pause":
		if g == nil || g.Status != risikoStatusLobby || input.ID != g.HostID {
			fehler(http.StatusForbidden, errors.New("Nur der Gastgeber stellt das Tempo ein."))
			return
		}
		g.Pause = risikoPauseWert(input.Pause)
		g.Version++

	case "start":
		if g == nil || g.Status != risikoStatusLobby {
			fehler(http.StatusConflict, errors.New("Es wartet keine Partie auf den Start."))
			return
		}
		if input.ID != g.HostID {
			fehler(http.StatusForbidden, errors.New("Nur der Gastgeber startet die Partie."))
			return
		}
		if len(g.Spieler) < g.Anzahl {
			fehler(http.StatusConflict, errors.New(fmt.Sprintf("Es fehlen noch %d Spieler oder Bots.", g.Anzahl-len(g.Spieler))))
			return
		}
		for _, sp := range g.Spieler {
			if !sp.Bereit && !sp.Bot {
				fehler(http.StatusConflict, errors.New(sp.Name+" ist noch nicht bereit."))
				return
			}
		}
		if g.Modus == "team" && len(g.Spieler)%2 == 1 {
			fehler(http.StatusConflict, errors.New("Teamkrieg braucht eine gerade Spielerzahl."))
			return
		}
		for i := range g.Spieler {
			g.Spieler[i].Team = -1
			if g.Modus == "team" {
				g.Spieler[i].Team = i % 2
			}
			g.Spieler[i].Farbe = risikoFarben[i%len(risikoFarben)]
		}
		risikoAufbau(g, rand.New(rand.NewSource(jetzt.UnixNano())))
		g.HostID = input.ID
		g.Version++
		risikoLog(g, "Partie gestartet ("+risikoModi[g.Modus]+").")

	case "neu":
		if g == nil || g.Status != risikoStatusBeendet {
			fehler(http.StatusConflict, errors.New("Eine neue Runde geht erst nach dem Spielende."))
			return
		}
		if risikoSitzVon(g, input.ID) < 0 {
			fehler(http.StatusForbidden, errors.New("Du gehörst nicht zu dieser Partie."))
			return
		}
		for i := range g.Spieler {
			g.Spieler[i].Bereit = g.Spieler[i].Bot
			g.Spieler[i].Hand = nil
			g.Spieler[i].Ausgeschieden = false
			g.Spieler[i].StartPool = 0
		}
		g.Status = risikoStatusLobby
		g.Gebiete = map[string]*RisikoGebiet{}
		g.Hauptstaedte = nil
		g.Missionen = nil
		g.Log = nil
		g.Sieger = -1
		g.SiegGrund = ""
		g.AmZug = 0
		g.Phase = ""
		g.Pool = 0
		g.Nachruecken = nil
		g.LetzterKampf = nil
		g.Version++

	default:
		if g == nil || g.Status != risikoStatusAktiv {
			fehler(http.StatusConflict, errors.New("Es läuft gerade keine aktive Risiko-Partie."))
			return
		}
		k := risikoSitzVon(g, input.ID)
		if k < 0 {
			fehler(http.StatusForbidden, errors.New("Du gehörst nicht zu dieser Partie."))
			return
		}
		würfel := rand.New(rand.NewSource(jetzt.UnixNano()))
		var err error
		switch action {
		case "setzen":
			n := input.Armeen
			if n < 1 {
				n = 1
			}
			for i := 0; i < n && err == nil; i++ {
				err = risikoSetzeArmee(g, k, input.Gebiet)
			}
		case "aufstellenFertig":
			err = risikoAufstellenFertig(g, k)
		case "karten":
			_, err = risikoKartenTauschen(g, k, input.Karten)
		case "verstaerkenEnde":
			err = risikoEndeVerstaerken(g, k)
		case "angriff":
			_, err = risikoAngriff(g, k, input.Von, input.Nach, würfel)
		case "nachruecken":
			err = risikoNachrueckenArmeen(g, k, input.Armeen)
		case "aufgeben":
			err = risikoAufgeben(g, k)
		case "angriffEnde":
			err = risikoEndeAngriff(g, k)
		case "verschieben":
			err = risikoVerschiebe(g, k, input.Von, input.Nach, input.Armeen)
		case "zugEnde":
			err = risikoZugEnde(g, k)
		default:
			err = errors.New("Unbekannte Risiko-Aktion.")
		}
		if err != nil {
			fehler(http.StatusConflict, err)
			return
		}
		risikoBotTreiben(room, jetzt)
	}

	room.LastActive = jetzt
	risikoSpeichern(room)
	sicht := risikoSichtFuer(room.Risiko, input.ID)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": sicht})
}

// risikoNamenAuffrischen übernimmt einen geänderten Spielernamen aus dem
// Raum in den laufenden Tisch (umbenennen im Chat oder im Feld der Leiste).
func risikoNamenAuffrischen(room *Room) {
	g := room.Risiko
	if g == nil {
		return
	}
	for i := range g.Spieler {
		p, ok := room.Players[g.Spieler[i].ID]
		if !ok || p.Name == "" || p.Name == g.Spieler[i].Name {
			continue
		}
		g.Spieler[i].Name = p.Name
		g.Version++
	}
}

// risikoSichtFuer baut den Stand für einen bestimmten Spieler: Die eigenen
// Karten sind sichtbar, fremde Hände und der Nachziehstapel nicht.
func risikoSichtFuer(g *RisikoSpiel, id string) *RisikoSpiel {
	if g == nil {
		return nil
	}
	sicht := *g
	sicht.Spieler = append([]RisikoSpieler(nil), g.Spieler...)
	for i := range sicht.Spieler {
		if sicht.Spieler[i].ID != id {
			sicht.Spieler[i].Hand = nil
		}
	}
	sicht.Stapel = nil
	return &sicht
}

func risikoPauseWert(pause string) string {
	switch strings.ToLower(strings.TrimSpace(pause)) {
	case "schnell", "langsam":
		return strings.ToLower(strings.TrimSpace(pause))
	}
	return "normal"
}

// ---------- Bots am Tisch --------------------------------------------------

// risikoBotTreiben lässt den Bot am Zug handeln. Der Plan wird einmal auf
// einer Kopie gerechnet und dann Schritt für Schritt am echten Stand
// abgespielt; jede Aktion läuft dabei durch dieselben Prüfungen wie bei einem
// Menschen. Aufstellen und Phasenwechsel gehen in einem Abruf durch, Kämpfe
// einzeln (eine Runde je Abruf, damit die Anzeige mitkommt).
func risikoBotTreiben(room *Room, jetzt time.Time) {
	g := room.Risiko
	if g == nil || g.Status != risikoStatusAktiv {
		return
	}
	k := g.AmZug
	if k < 0 || k >= len(g.Spieler) || !g.Spieler[k].Bot || g.Spieler[k].Ausgeschieden {
		return
	}
	if len(g.BotPlan) == 0 {
		if g.BotFrist != 0 && jetzt.UnixMilli() < g.BotFrist {
			return
		}
		g.BotPlan = risikoBotPlanen(g, k, rand.New(rand.NewSource(jetzt.UnixNano())))
		g.BotFrist = jetzt.Add(risikoTakt(g.Pause)).UnixMilli()
		if len(g.BotPlan) == 0 {
			// Kein Zug möglich (z. B. alles dicht): Zug beenden.
			if err := risikoZugEnde(g, k); err != nil {
				return
			}
		}
		return
	}

	budget := 120
	for len(g.BotPlan) > 0 && budget > 0 {
		aktion := g.BotPlan[0]
		g.BotPlan = g.BotPlan[1:]
		budget--
		if err := risikoAktionAusfuehren(g, k, aktion, rand.New(rand.NewSource(jetzt.UnixNano()))); err != nil {
			// Plan verworfen; der nächste Abruf rechnet neu.
			g.BotPlan = nil
			g.BotFrist = jetzt.Add(time.Second).UnixMilli()
			return
		}
		if aktion.Art == "angriff" {
			break // eine Kampfrunde je Abruf
		}
		if g.Status != risikoStatusAktiv || g.AmZug != k || g.Spieler[k].Ausgeschieden {
			g.BotPlan = nil
			break
		}
	}
	if len(g.BotPlan) > 0 {
		g.BotFrist = jetzt.Add(risikoTakt(g.Pause)).UnixMilli()
	}
}

// risikoSitzIDs werden beim Verlassen eines Raums gebraucht.
func risikoSitzIDs(g *RisikoSpiel) []string {
	if g == nil {
		return nil
	}
	ids := make([]string, 0, len(g.Spieler))
	for _, sp := range g.Spieler {
		if !sp.Bot {
			ids = append(ids, sp.ID)
		}
	}
	return ids
}

func risikoSitzUmbenennen(g *RisikoSpiel, id, name string) {
	if g == nil || name == "" {
		return
	}
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			g.Spieler[i].Name = name
			g.Version++
		}
	}
}

// ---------- Spielstände ----------------------------------------------------

type risikoStand struct {
	Code  string       `json:"code"`
	Name  string       `json:"name"`
	Spiel *RisikoSpiel `json:"spiel"`
}

func risikoStandPfad(code string) string {
	return filepath.Join(risikoStaendeOrdner, strings.ToUpper(code)+".json")
}

func risikoSpeichern(room *Room) {
	if room == nil || room.Risiko == nil || room.Code == "" {
		return
	}
	if err := os.MkdirAll(risikoStaendeOrdner, 0o755); err != nil {
		return
	}
	daten, err := json.MarshalIndent(risikoStand{Code: room.Code, Name: room.Name, Spiel: room.Risiko}, "", " ")
	if err != nil {
		return
	}
	_ = os.WriteFile(risikoStandPfad(room.Code), daten, 0o644)
}

func risikoStandWeg(code string) {
	_ = os.Remove(risikoStandPfad(code))
}

// risikoStaendeLaden holt gespeicherte Partien beim Start zurück in den
// Speicher. Räume ohne Spielstand bleiben unberührt.
func risikoStaendeLaden() {
	eintraege, err := os.ReadDir(risikoStaendeOrdner)
	if err != nil {
		return
	}
	for _, e := range eintraege {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		daten, err := os.ReadFile(filepath.Join(risikoStaendeOrdner, e.Name()))
		if err != nil {
			continue
		}
		var stand risikoStand
		if err := json.Unmarshal(daten, &stand); err != nil || stand.Spiel == nil || stand.Code == "" {
			continue
		}
		if _, da := rooms[stand.Code]; da {
			continue
		}
		rooms[stand.Code] = &Room{
			Code:       stand.Code,
			Name:       stand.Name,
			Created:    time.Now(),
			LastActive: time.Now(),
			Players:    map[string]Player{},
			Messages:   []Message{},
			Mutes:      map[string]time.Time{},
			MsgTimes:   map[string][]time.Time{},
			botNamen:   map[string]string{},
			Risiko:     stand.Spiel,
		}
	}
}

func init() {
	risikoStaendeLaden()
}
