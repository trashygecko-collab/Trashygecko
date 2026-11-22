package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Baccara (Chemin de fer): Ein Spieler legt die Bank, alle anderen setzen
// dagegen. Es gibt nur eine Ponte-Hand, die allen Pointeuren gemeinsam gehört;
// die dritte Karte folgt dem festen Tableau. Gleichstand ist ungültig.

const (
	baccaraMaxSpieler = 6
	baccaraKasse      = 1000 // 10 € Startgeld
	baccaraSchritt    = 10   // kleinste Einheit: 10 Cent
	baccaraBankMin    = 100  // 1 € kleinste Bank
	baccaraPakete     = 6    // sechs Spiele a 52 Karten
)

type BaccaraSpieler struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Kasse   int    `json:"kasse"`
	Einsatz int    `json:"einsatz"`
	Status  string `json:"status"` // wartet, setzt, ponte, pleite
	Ready   bool   `json:"ready"`
}

type BaccaraErgebnis struct {
	Name    string `json:"name"`
	Einsatz int    `json:"einsatz"`
	Ausgang string `json:"ausgang"` // gewonnen, verloren, gleichstand
	Netto   int    `json:"netto"`
}

type BaccaraGame struct {
	ID          string            `json:"id"`
	HostID      string            `json:"hostId"`
	Status      string            `json:"status"` // invited, bank, setzen, zug, runde, ende
	Runde       int               `json:"runde"`
	Bank        int               `json:"bank"`
	BankSitz    int               `json:"bankSitz"`
	PonteSitz   int               `json:"ponteSitz"`
	PonteKarten []int             `json:"-"`
	BankKarten  []int             `json:"-"`
	Stapel      []int             `json:"-"`
	Zug         int               `json:"zug"`
	Spieler     []BaccaraSpieler  `json:"spieler"`
	Note        string            `json:"note"`
	Sieger      string            `json:"sieger,omitempty"`
	Ergebnisse  []BaccaraErgebnis `json:"ergebnisse"`
	BankWechsel bool              `json:"-"`

	// Bots am Tisch (action=bot). Sie handeln über baccaraAktion – dieselbe
	// Prüfung wie ein menschlicher Klick.
	BotID    string    `json:"botId,omitempty"`
	Bots     []string  `json:"-"`
	BotStufe string    `json:"botStufe,omitempty"`
	BotFrist time.Time `json:"-"`
}

type BaccaraAnsicht struct {
	ID          string            `json:"id"`
	HostID      string            `json:"hostId"`
	Status      string            `json:"status"`
	Runde       int               `json:"runde"`
	Bank        int               `json:"bank"`
	BankSitz    int               `json:"bankSitz"`
	PonteSitz   int               `json:"ponteSitz"`
	PonteName   string            `json:"ponteName"`
	PonteKarten []int             `json:"ponteKarten"`
	PontePunkte int               `json:"pontePunkte"`
	BankKarten  []int             `json:"bankKarten"`
	BankPunkte  int               `json:"bankPunkte"`
	Belegt      int               `json:"belegt"`
	Zug         int               `json:"zug"`
	Spieler     []BaccaraSpieler  `json:"spieler"`
	Note        string            `json:"note"`
	Sieger      string            `json:"sieger,omitempty"`
	Ergebnisse  []BaccaraErgebnis `json:"ergebnisse"`
	BankWechsel bool              `json:"bankWechsel"`
	BotStufe    string            `json:"botStufe,omitempty"`
}

// ---------------------------------------------------------------- Karten

func baccaraDeck() []int {
	deck := make([]int, 0, baccaraPakete*52)
	for p := 0; p < baccaraPakete; p++ {
		for wert := 1; wert <= 13; wert++ {
			for i := 0; i < 4; i++ {
				deck = append(deck, wert)
			}
		}
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

// baccaraWert ist der Zählwert einer Karte: Ass 1, Zehn und Bilder 0.
func baccaraWert(karte int) int {
	if karte >= 10 {
		return 0
	}
	return karte
}

// baccaraPunkte zählt nur die Einerstelle: 7 + 5 = 12 zählt 2.
func baccaraPunkte(karten []int) int {
	summe := 0
	for _, karte := range karten {
		summe += baccaraWert(karte)
	}
	return summe % 10
}

func baccaraName(karten []int) string {
	teile := make([]string, 0, len(karten))
	for _, karte := range karten {
		switch karte {
		case 1:
			teile = append(teile, "A")
		case 11:
			teile = append(teile, "B")
		case 12:
			teile = append(teile, "D")
		case 13:
			teile = append(teile, "K")
		default:
			teile = append(teile, strconv.Itoa(karte))
		}
	}
	return strings.Join(teile, " ")
}

func baccaraEuro(cent int) string {
	v := cent
	zeichen := ""
	if v < 0 {
		zeichen = "−"
		v = -v
	}
	return zeichen + strconv.Itoa(v/100) + "," + strconv.Itoa(100 + v%100)[1:] + " €"
}

// ---------------------------------------------------------------- Aufbau

func baccaraNeuerSpieler(id, name string) BaccaraSpieler {
	return BaccaraSpieler{ID: id, Name: name, Kasse: baccaraKasse, Status: "wartet"}
}

func baccaraSitz(g *BaccaraGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// baccaraNach liefert den nächsten Spieler, der die Bank legen kann. Wer nicht
// einmal die kleinste Bank in der Kasse hat, kommt nicht in Frage – sonst bleibt
// der Tisch stehen, weil „Bank legen“ dauerhaft abgelehnt wird und niemand mehr
// einen Knopf hat, der weiterhilft.
func baccaraNach(g *BaccaraGame, von int) int {
	n := len(g.Spieler)
	for i := 1; i <= n; i++ {
		k := (von + i) % n
		if g.Spieler[k].Status != "pleite" && g.Spieler[k].Kasse >= baccaraBankMin {
			return k
		}
	}
	return -1
}

// baccaraKannSetzen meldet, ob mindestens ein Pointeur noch Geld für den
// kleinsten Einsatz hat.
func baccaraKannSetzen(g *BaccaraGame) bool {
	for i := range g.Spieler {
		if i == g.BankSitz || g.Spieler[i].Status == "pleite" {
			continue
		}
		if g.Spieler[i].Kasse >= baccaraSchritt {
			return true
		}
	}
	return false
}

func baccaraBelegt(g *BaccaraGame) int {
	summe := 0
	for _, p := range g.Spieler {
		summe += p.Einsatz
	}
	return summe
}

// baccaraPonteSitz ist der Spieler mit dem höchsten Einsatz – er bekommt die
// Karten. Bei Gleichstand entscheidet die Sitzordnung ab der Bank.
func baccaraPonteSitz(g *BaccaraGame) int {
	best := -1
	for i := 1; i <= len(g.Spieler); i++ {
		k := (g.BankSitz + i) % len(g.Spieler)
		if g.Spieler[k].Einsatz <= 0 {
			continue
		}
		if best < 0 || g.Spieler[k].Einsatz > g.Spieler[best].Einsatz {
			best = k
		}
	}
	return best
}

func baccaraNeueRunde(g *BaccaraGame) {
	g.Runde++
	g.PonteKarten = []int{}
	g.BankKarten = []int{}
	g.Ergebnisse = nil
	g.PonteSitz = -1
	g.Zug = -1
	g.BankWechsel = false
	for i := range g.Spieler {
		g.Spieler[i].Einsatz = 0
		if g.Spieler[i].Kasse <= 0 && i != g.BankSitz {
			g.Spieler[i].Status = "pleite"
		} else {
			g.Spieler[i].Status = "wartet"
		}
	}
	if g.Bank <= 0 {
		baccaraBankWeiter(g)
		return
	}
	// Kann keiner der Pointeure mehr setzen, wäre die Runde nicht mehr
	// weiterzuspielen: der Bankhalter könnte keine Karten geben und niemand
	// hätte einen Knopf. Dann ist die Partie vorbei.
	if !baccaraKannSetzen(g) {
		g.Status = "ende"
		if g.BankSitz >= 0 && g.BankSitz < len(g.Spieler) {
			g.Sieger = g.Spieler[g.BankSitz].Name
			g.Note = "Niemand kann mehr setzen – " + g.Sieger + " hat die Partie gewonnen."
		}
		return
	}
	g.Status = "setzen"
	g.Note = "Runde " + strconv.Itoa(g.Runde) + ": Setzt gegen die Bank (" + baccaraEuro(g.Bank) + ")."
}

func baccaraBankWeiter(g *BaccaraGame) {
	if g.Bank > 0 {
		if g.BankSitz >= 0 && g.BankSitz < len(g.Spieler) {
			g.Spieler[g.BankSitz].Kasse += g.Bank
		}
		g.Bank = 0
	}
	naechster := baccaraNach(g, g.BankSitz)
	if naechster < 0 {
		g.Status = "ende"
		g.Note = "Niemand kann die Bank legen – die Partie ist vorbei."
		return
	}
	g.BankSitz = naechster
	g.PonteKarten = []int{}
	g.BankKarten = []int{}
	g.PonteSitz = -1
	g.Zug = -1
	g.BankWechsel = false
	for i := range g.Spieler {
		g.Spieler[i].Einsatz = 0
		if g.Spieler[i].Kasse <= 0 && i != naechster {
			g.Spieler[i].Status = "pleite"
		}
	}
	g.Status = "bank"
	g.Note = g.Spieler[naechster].Name + " hat die Bank – wie viel legt er auf den Tisch?"
}

// ---------------------------------------------------------------- Spiel

func baccaraZiehen(g *BaccaraGame) int {
	if len(g.Stapel) < 12 {
		g.Stapel = baccaraDeck()
	}
	karte := g.Stapel[0]
	g.Stapel = g.Stapel[1:]
	return karte
}

func baccaraGeben(g *BaccaraGame) {
	g.PonteSitz = baccaraPonteSitz(g)
	if g.PonteSitz < 0 {
		return
	}
	g.PonteKarten = []int{baccaraZiehen(g)}
	g.BankKarten = []int{baccaraZiehen(g)}
	g.PonteKarten = append(g.PonteKarten, baccaraZiehen(g))
	g.BankKarten = append(g.BankKarten, baccaraZiehen(g))
	g.Zug = -1

	ponte := baccaraPunkte(g.PonteKarten)
	bank := baccaraPunkte(g.BankKarten)
	g.Status = "spiel"
	note := "Ponte " + strconv.Itoa(ponte) + " – Bank " + strconv.Itoa(bank) + "."

	switch {
	case ponte >= 8 || bank >= 8:
		// Naturel: keine dritte Karte.
		if ponte >= 8 && bank >= 8 {
			note += " Naturel auf beiden Seiten."
		} else if ponte >= 8 {
			note += " Naturel beim Ponte – die Bank zieht nicht."
		} else {
			note += " Naturel bei der Bank – der Ponte zieht nicht."
		}
		g.Note = note
		baccaraAbrechnen(g)
	case ponte <= 4:
		g.PonteKarten = append(g.PonteKarten, baccaraZiehen(g))
		g.Note = note + " Der Ponte muss ziehen: " + strconv.Itoa(baccaraPunkte(g.PonteKarten)) + " Punkte."
		baccaraBankZug(g)
	case ponte == 5:
		// Freie Wahl: der Ponte entscheidet.
		g.Status = "zug"
		g.Zug = g.PonteSitz
		g.Note = note + " " + g.Spieler[g.PonteSitz].Name + " hat fünf – Karte oder stehen bleiben?"
	case ponte == 6 || ponte == 7:
		g.Note = note + " Der Ponte bleibt bei " + strconv.Itoa(ponte) + " stehen."
		baccaraBankZug(g)
	default:
		g.Note = note + " Der Ponte ist überkauft."
		baccaraBankZug(g)
	}
}

// baccaraBankZug zieht für die Bank nach dem Tableau des Chemin de fer. Hat der
// Ponte nicht gezogen, zieht die Bank bis sechs und bleibt bei sieben stehen.
func baccaraBankZug(g *BaccaraGame) {
	punkte := baccaraPunkte(g.BankKarten)
	ziehen := false
	if len(g.PonteKarten) >= 3 {
		dritte := baccaraWert(g.PonteKarten[2])
		switch {
		case punkte <= 2:
			ziehen = true
		case punkte == 3:
			ziehen = dritte != 8
		case punkte == 4:
			ziehen = dritte != 1 && dritte != 8 && dritte != 9 && dritte != 0
		case punkte == 5:
			ziehen = dritte == 5 || dritte == 6 || dritte == 7
		case punkte == 6:
			ziehen = dritte == 6 || dritte == 7
		default:
			ziehen = false
		}
	} else {
		ziehen = punkte <= 5
	}
	if ziehen && punkte <= 7 {
		g.BankKarten = append(g.BankKarten, baccaraZiehen(g))
		g.Note += " Die Bank zieht auf " + strconv.Itoa(baccaraPunkte(g.BankKarten)) + "."
	} else {
		g.Note += " Die Bank bleibt bei " + strconv.Itoa(punkte) + " stehen."
	}
	baccaraAbrechnen(g)
}

func baccaraAbrechnen(g *BaccaraGame) {
	ponte := baccaraPunkte(g.PonteKarten)
	bank := baccaraPunkte(g.BankKarten)
	verloren := ponte > bank
	gleich := ponte == bank
	ergebnisse := []BaccaraErgebnis{}

	for i := range g.Spieler {
		p := &g.Spieler[i]
		if p.Einsatz <= 0 {
			continue
		}
		einsatz := p.Einsatz
		ergebnis := BaccaraErgebnis{Name: p.Name, Einsatz: einsatz}
		switch {
		case gleich:
			p.Kasse += einsatz
			ergebnis.Ausgang = "gleichstand"
			ergebnis.Netto = 0
		case verloren:
			p.Kasse += einsatz * 2
			g.Bank -= einsatz
			ergebnis.Ausgang = "gewonnen"
			ergebnis.Netto = einsatz
		default:
			g.Bank += einsatz
			ergebnis.Ausgang = "verloren"
			ergebnis.Netto = -einsatz
		}
		ergebnisse = append(ergebnisse, ergebnis)
	}

	g.Ergebnisse = ergebnisse
	g.BankWechsel = verloren
	g.Zug = -1
	g.Status = "runde"
	ausgang := "Die Bank gewinnt"
	if verloren {
		ausgang = "Der Ponte gewinnt"
	} else if gleich {
		ausgang = "Gleichstand – Einsätze zurück"
	}
	g.Note = "Ponte " + strconv.Itoa(ponte) + " gegen Bank " + strconv.Itoa(bank) + ": " + ausgang + "."

	if verloren {
		g.Note += " Die Bank ist geschlagen und wandert weiter."
	}

	mitGeld := 0
	naechster := ""
	for i := range g.Spieler {
		if g.Spieler[i].Kasse > 0 || i == g.BankSitz {
			mitGeld++
			if g.Spieler[i].Kasse > 0 {
				naechster = g.Spieler[i].Name
			}
		}
	}
	if mitGeld < 2 {
		g.Status = "ende"
		g.Sieger = naechster
		g.Note += " Spiel beendet – " + g.Sieger + " hat alles gewonnen."
	}
}

// ---------------------------------------------------------------- Sicht

func baccaraAnsicht(g *BaccaraGame) *BaccaraAnsicht {
	if g == nil {
		return nil
	}
	a := &BaccaraAnsicht{
		ID:          g.ID,
		HostID:      g.HostID,
		Status:      g.Status,
		Runde:       g.Runde,
		Bank:        g.Bank,
		BankSitz:    g.BankSitz,
		PonteSitz:   g.PonteSitz,
		PonteKarten: append([]int{}, g.PonteKarten...),
		PontePunkte: baccaraPunkte(g.PonteKarten),
		BankKarten:  append([]int{}, g.BankKarten...),
		BankPunkte:  baccaraPunkte(g.BankKarten),
		Belegt:      baccaraBelegt(g),
		Zug:         g.Zug,
		Spieler:     append([]BaccaraSpieler{}, g.Spieler...),
		Note:        g.Note,
		Sieger:      g.Sieger,
		Ergebnisse:  append([]BaccaraErgebnis{}, g.Ergebnisse...),
		BankWechsel: g.BankWechsel,
		BotStufe:    g.BotStufe,
	}
	if g.PonteSitz >= 0 && g.PonteSitz < len(g.Spieler) {
		a.PonteName = g.Spieler[g.PonteSitz].Name
	}
	return a
}

// ---------------------------------------------------------------- Endpunkt

func baccaraHandler(w http.ResponseWriter, r *http.Request) {
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
		baccaraBotTreiben(room, time.Now())
		game := baccaraAnsicht(room.Baccara)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID     string `json:"id"`
		Betrag int    `json:"betrag"`
		Stufe  string `json:"stufe"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
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

	zeit := time.Now().Format("15:04:05")

	switch action {
	case "create":
		if room.Baccara != nil && room.Baccara.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Runde Baccara.", http.StatusConflict)
			return
		}
		room.Baccara = &BaccaraGame{
			ID:        randomCode(),
			HostID:    input.ID,
			Status:    "invited",
			BankSitz:  0,
			PonteSitz: -1,
			Zug:       -1,
			Spieler:   []BaccaraSpieler{baccaraNeuerSpieler(input.ID, player.Name)},
			Note:      "Tisch eröffnet – einer legt die Bank, die anderen setzen dagegen.",
		}
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Runde Baccara – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Baccara
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if baccaraSitz(game, input.ID) >= 0 {
			break
		}
		if len(game.Spieler) >= baccaraMaxSpieler {
			mu.Unlock()
			http.Error(w, "Dieser Tisch ist mit 6 Spielern voll.", http.StatusConflict)
			return
		}
		game.Spieler = append(game.Spieler, baccaraNeuerSpieler(input.ID, player.Name))
		game.Note = player.Name + " setzt sich dazu."

	case "bot":
		room.botNamenNeu()
		if room.Baccara != nil && room.Baccara.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Runde Baccara.", http.StatusConflict)
			return
		}
		stufe := botStufe(input.Stufe)
		spiel := &BaccaraGame{
			ID:        randomCode(),
			HostID:    input.ID,
			Status:    "invited",
			BankSitz:  0,
			PonteSitz: -1,
			Zug:       -1,
			BotID:     botID(1),
			Bots:      []string{botID(1)},
			BotStufe:  stufe,
			Note:      room.botName(1) + " sitzt bereit am Tisch (" + botStufeText(stufe) + ").",
		}
		spiel.Spieler = append(spiel.Spieler, baccaraNeuerSpieler(input.ID, player.Name))
		spiel.Spieler = append(spiel.Spieler, baccaraNeuerSpieler(botID(1), room.botName(1)))
		for i := range spiel.Spieler {
			spiel.Spieler[i].Ready = true
		}
		room.Baccara = spiel
		if status, meldung := baccaraAktion(room, input.ID, "start", 0); meldung != "" {
			mu.Unlock()
			http.Error(w, meldung, status)
			return
		}
		// Der Bot legt zuerst die Bank, damit der Mensch als Ponte setzt.
		room.Baccara.BankSitz = 1
		room.Baccara.Bank = 0
		room.Baccara.Status = "bank"
		room.Baccara.Note = room.botName(1) + " hat die Bank – wie viel legt er auf den Tisch?"
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "spielt Baccara gegen " + room.botName(1) + " (" + botStufeText(stufe) + ").",
			Time:     zeit,
			SenderID: input.ID,
		})

	default:
		if status, meldung := baccaraAktion(room, input.ID, action, input.Betrag); meldung != "" {
			mu.Unlock()
			http.Error(w, meldung, status)
			return
		}
	}

	room.LastActive = time.Now()
	result := baccaraAnsicht(room.Baccara)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

// ---------------------------------------------------------------- Aktionen

// baccaraAktion führt eine Aktion am Tisch aus. Mensch und Bot gehen durch
// dieselbe Prüfung; die Rückgabe ist HTTP-Status und Meldung ("" heißt
// erledigt).
func baccaraAktion(room *Room, id, aktion string, betrag int) (int, string) {
	game := room.Baccara

	switch aktion {
	case "ready":
		if game == nil || game.Status != "invited" {
			return http.StatusConflict, "Bereit drücken geht nur vor dem Start."
		}
		sitz := baccaraSitz(game, id)
		if sitz < 0 {
			return http.StatusForbidden, "Du sitzt nicht an diesem Tisch."
		}
		game.Spieler[sitz].Ready = !game.Spieler[sitz].Ready
		if game.Spieler[sitz].Ready {
			game.Note = game.Spieler[sitz].Name + " ist bereit."
		} else {
			game.Note = game.Spieler[sitz].Name + " wartet noch."
		}

	case "start":
		if game == nil || game.Status != "invited" {
			return http.StatusConflict, "Es wartet keine Runde auf den Start."
		}
		if game.HostID != id {
			return http.StatusForbidden, "Nur der Gastgeber kann die Bank übernehmen."
		}
		if len(game.Spieler) < 2 {
			return http.StatusConflict, "Baccara braucht mindestens zwei Spieler."
		}
		for _, p := range game.Spieler {
			if !p.Ready {
				return http.StatusConflict, "Erst müssen alle BEREIT drücken."
			}
		}
		game.BankSitz = 0
		game.Runde = 0
		game.Bank = 0
		game.Status = "bank"
		game.Note = game.Spieler[0].Name + " hat die Bank – wie viel legt er auf den Tisch?"

	case "bank":
		if game == nil || game.Status != "bank" {
			return http.StatusConflict, "Gerade legt niemand eine Bank."
		}
		if baccaraSitz(game, id) != game.BankSitz {
			return http.StatusForbidden, "Nur der Bankhalter legt die Bank."
		}
		halter := &game.Spieler[game.BankSitz]
		if betrag < baccaraBankMin {
			return http.StatusBadRequest, "Die Bank muss mindestens " + baccaraEuro(baccaraBankMin) + " haben."
		}
		betrag = (betrag / baccaraSchritt) * baccaraSchritt
		if betrag > halter.Kasse {
			return http.StatusConflict, "So viel Geld hast du nicht in der Kasse."
		}
		halter.Kasse -= betrag
		game.Bank = betrag
		baccaraNeueRunde(game)
		game.Note = "Bank gelegt: " + baccaraEuro(game.Bank) + " – jetzt setzen."

	case "setzen", "banco":
		if game == nil || game.Status != "setzen" {
			return http.StatusConflict, "Gerade wird nicht gesetzt."
		}
		sitz := baccaraSitz(game, id)
		if sitz < 0 || sitz == game.BankSitz {
			return http.StatusForbidden, "Die Bank setzt nicht gegen sich selbst."
		}
		p := &game.Spieler[sitz]
		if p.Status == "pleite" {
			return http.StatusConflict, "Deine Kasse ist leer."
		}
		if p.Einsatz > 0 {
			return http.StatusConflict, "Du hast für diese Runde schon gesetzt."
		}
		frei := game.Bank - baccaraBelegt(game)
		if frei <= 0 {
			return http.StatusConflict, "Die Bank ist schon voll abgedeckt – diese Runde setzt niemand mehr."
		}
		einsatz := (betrag / baccaraSchritt) * baccaraSchritt
		if aktion == "banco" {
			einsatz = frei
			if einsatz > p.Kasse {
				einsatz = (p.Kasse / baccaraSchritt) * baccaraSchritt
			}
			if einsatz <= 0 {
				return http.StatusConflict, "Für Banco fehlt dir Geld."
			}
		}
		if einsatz < baccaraSchritt {
			return http.StatusBadRequest, "Mindestens " + baccaraEuro(baccaraSchritt) + "."
		}
		if einsatz > p.Kasse {
			return http.StatusConflict, "So viel Geld hast du nicht."
		}
		if einsatz > frei {
			return http.StatusConflict, "Die Bank ist nur noch mit " + baccaraEuro(frei) + " gedeckt."
		}
		p.Kasse -= einsatz
		p.Einsatz = einsatz
		p.Status = "setzt"
		if aktion == "banco" {
			game.Note = p.Name + " ruft BANCO – " + baccaraEuro(einsatz) + " auf die ganze Bank."
		} else {
			game.Note = p.Name + " setzt " + baccaraEuro(einsatz) + "."
		}

	case "geben":
		if game == nil || game.Status != "setzen" {
			return http.StatusConflict, "Gerade wird nicht gegeben."
		}
		if baccaraSitz(game, id) != game.BankSitz {
			return http.StatusForbidden, "Nur der Bankhalter gibt die Karten."
		}
		if baccaraBelegt(game) <= 0 {
			return http.StatusConflict, "Es hat noch niemand gesetzt."
		}
		baccaraGeben(game)

	case "karte", "stehen":
		if game == nil || game.Status != "zug" {
			return http.StatusConflict, "Gerade entscheidet niemand."
		}
		if baccaraSitz(game, id) != game.Zug {
			return http.StatusForbidden, "Nur der Ponte mit fünf Punkten entscheidet."
		}
		if aktion == "karte" {
			game.PonteKarten = append(game.PonteKarten, baccaraZiehen(game))
			game.Note = game.Spieler[game.Zug].Name + " nimmt eine Karte: " + strconv.Itoa(baccaraPunkte(game.PonteKarten)) + " Punkte."
		} else {
			game.Note = game.Spieler[game.Zug].Name + " bleibt bei fünf stehen."
		}
		game.Zug = -1
		game.Status = "spiel"
		baccaraBankZug(game)

	case "naechste":
		if game == nil || game.Status != "runde" {
			return http.StatusConflict, "Die Runde ist noch nicht abgerechnet."
		}
		if baccaraSitz(game, id) < 0 {
			return http.StatusForbidden, "Du gehörst nicht zu dieser Runde."
		}
		if game.BankWechsel || game.Bank <= 0 {
			baccaraBankWeiter(game)
			break
		}
		baccaraNeueRunde(game)

	case "abgeben":
		if game == nil || game.Status != "runde" {
			return http.StatusConflict, "Die Bank lässt sich erst nach einer Runde abgeben."
		}
		if baccaraSitz(game, id) != game.BankSitz {
			return http.StatusForbidden, "Nur der Bankhalter kann abgeben."
		}
		baccaraBankWeiter(game)

	case "reset":
		if game == nil || game.Status != "ende" {
			return http.StatusConflict, "Eine neue Partie geht erst nach dem Ende."
		}
		if baccaraSitz(game, id) < 0 {
			return http.StatusForbidden, "Du gehörst nicht zu dieser Runde."
		}
		frisch := &BaccaraGame{
			ID:        game.ID,
			HostID:    game.HostID,
			Status:    "invited",
			BankSitz:  0,
			PonteSitz: -1,
			Zug:       -1,
			Spieler:   []BaccaraSpieler{},
			Note:      "Neue Partie – erst BEREIT drücken.",
		}
		for _, p := range game.Spieler {
			frisch.Spieler = append(frisch.Spieler, baccaraNeuerSpieler(p.ID, p.Name))
		}
		room.Baccara = frisch
		room.LastActive = time.Now()
		return 0, ""

	default:
		return http.StatusBadRequest, "Unbekannte Aktion."
	}

	room.LastActive = time.Now()
	return 0, ""
}
