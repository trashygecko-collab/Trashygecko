package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 17 und 4: Ein Spieler ist die Bank, alle anderen spielen dagegen. Die Bank
// liegt offen auf dem Tisch und deckt die Einsätze – mehr kann sie nicht
// verlieren. Bei Gleichstand gewinnt die Bank.

const (
	siebzehnMaxSpieler = 6
	siebzehnKasse      = 1000 // 10 € Startgeld
	siebzehnSchritt    = 10   // kleinste Einheit: 10 Cent
	siebzehnBankMin    = 100  // 1 € kleinste Bank
)

type SiebzehnSpieler struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Kasse   int    `json:"kasse"`
	Einsatz int    `json:"einsatz"`
	Karten  []int  `json:"karten"`
	Status  string `json:"status"` // wartet, setzt, spielt, haelt, ueberkauft, pleite
	Ready   bool   `json:"ready"`
	Feuer   bool   `json:"feuer"`
}

type SiebzehnErgebnis struct {
	Name    string `json:"name"`
	Einsatz int    `json:"einsatz"`
	Karten  []int  `json:"karten"`
	Summe   int    `json:"summe"`
	Ausgang string `json:"ausgang"` // gewonnen, blackjack, verloren, gleichstand, ueberkauft
	Netto   int    `json:"netto"`
}

type SiebzehnGame struct {
	ID         string             `json:"id"`
	HostID     string             `json:"hostId"`
	Status     string             `json:"status"` // invited, bank, setzen, spiel, runde, ende
	Runde      int                `json:"runde"`
	Bank       int                `json:"bank"`
	BankSitz   int                `json:"bankSitz"`
	BankKarten []int              `json:"-"`
	Spieler    []SiebzehnSpieler  `json:"spieler"`
	Stapel     []int              `json:"-"`
	Zug        int                `json:"zug"`
	Note       string             `json:"note"`
	Sieger     string             `json:"sieger,omitempty"`
	Ergebnisse []SiebzehnErgebnis `json:"ergebnisse"`
	FeuerSitz  int                `json:"feuerSitz"`
}

type SiebzehnSpielerSicht struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Kasse   int    `json:"kasse"`
	Einsatz int    `json:"einsatz"`
	Karten  []int  `json:"karten"`
	Summe   int    `json:"summe"`
	Status  string `json:"status"`
	Ready   bool   `json:"ready"`
	Feuer   bool   `json:"feuer"`
}

type SiebzehnAnsicht struct {
	ID         string                 `json:"id"`
	HostID     string                 `json:"hostId"`
	Status     string                 `json:"status"`
	Runde      int                    `json:"runde"`
	Bank       int                    `json:"bank"`
	BankSitz   int                    `json:"bankSitz"`
	BankKarten []int                  `json:"bankKarten"`
	BankDeckt  bool                   `json:"bankDeckt"`
	Belegt     int                    `json:"belegt"`
	Spieler    []SiebzehnSpielerSicht `json:"spieler"`
	Zug        int                    `json:"zug"`
	Note       string                 `json:"note"`
	Sieger     string                 `json:"sieger,omitempty"`
	Ergebnisse []SiebzehnErgebnis     `json:"ergebnisse"`
	FeuerSitz  int                    `json:"feuerSitz"`
}

// ---------------------------------------------------------------- Karten

func siebzehnDeck() []int {
	// Vier Spiele a 52 Karten reichen für viele Runden ohne Mischen.
	deck := make([]int, 0, 4*52)
	for s := 0; s < 4; s++ {
		for wert := 2; wert <= 14; wert++ {
			for i := 0; i < 4; i++ {
				deck = append(deck, wert)
			}
		}
	}
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

// siebzehnSumme zählt Asse erst als 11 und senkt sie, solange die Summe über
// 21 liegt.
func siebzehnSumme(karten []int) int {
	summe := 0
	asse := 0
	for _, karte := range karten {
		switch {
		case karte == 14:
			summe += 11
			asse++
		case karte >= 11:
			summe += 10
		default:
			summe += karte
		}
	}
	for summe > 21 && asse > 0 {
		summe -= 10
		asse--
	}
	return summe
}

func siebzehnName(karten []int) string {
	teile := make([]string, 0, len(karten))
	for _, karte := range karten {
		switch karte {
		case 14:
			teile = append(teile, "A")
		case 13:
			teile = append(teile, "K")
		case 12:
			teile = append(teile, "D")
		case 11:
			teile = append(teile, "B")
		default:
			teile = append(teile, strconv.Itoa(karte))
		}
	}
	return strings.Join(teile, " ")
}

func siebzehnEuro(cent int) string {
	if cent < 0 {
		cent = 0
	}
	return strconv.Itoa(cent/100) + "," + strconv.Itoa(100 + cent%100)[1:] + " €"
}

// ---------------------------------------------------------------- Aufbau

func siebzehnNeuerSpieler(id, name string) SiebzehnSpieler {
	return SiebzehnSpieler{ID: id, Name: name, Kasse: siebzehnKasse, Karten: []int{}, Status: "wartet", Feuer: false}
}

func siebzehnSitz(g *SiebzehnGame, id string) int {
	for i, p := range g.Spieler {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// siebzehnNach liefert den nächsten Spieler, der die Bank legen kann. Wer
// nicht einmal die kleinste Bank in der Kasse hat, kommt nicht in Frage – sonst
// bleibt der Tisch stehen, weil „Bank legen" dauerhaft abgelehnt wird und
// niemand mehr einen Knopf hat, der weiterhilft.
func siebzehnNach(g *SiebzehnGame, von int) int {
	n := len(g.Spieler)
	for i := 1; i <= n; i++ {
		k := (von + i) % n
		if g.Spieler[k].Status != "pleite" && g.Spieler[k].Kasse >= siebzehnBankMin {
			return k
		}
	}
	return -1
}

func siebzehnBelegt(g *SiebzehnGame) int {
	summe := 0
	for _, p := range g.Spieler {
		summe += p.Einsatz
	}
	return summe
}

func siebzehnNeueRunde(g *SiebzehnGame) {
	g.Runde++
	g.BankKarten = []int{}
	g.Ergebnisse = nil
	g.FeuerSitz = -1
	for i := range g.Spieler {
		g.Spieler[i].Einsatz = 0
		g.Spieler[i].Karten = []int{}
		g.Spieler[i].Feuer = false
		if g.Spieler[i].Kasse <= 0 && i != g.BankSitz {
			g.Spieler[i].Status = "pleite"
		} else {
			g.Spieler[i].Status = "wartet"
		}
	}
	g.Status = "setzen"
	g.Zug = -1
	g.Note = "Runde " + strconv.Itoa(g.Runde) + ": Setzt eure Einsätze – die Bank hat " + siebzehnEuro(g.Bank) + "."
	if g.Bank <= 0 {
		g.Status = "bank"
		g.Note = "Die Bank ist leer – der Nächste legt eine neue Bank."
		return
	}
	// Kann keiner der Pointeure mehr setzen, wäre die Runde nicht mehr
	// weiterzuspielen: der Bankhalter könnte keine Karten geben und niemand
	// hätte einen Knopf. Dann ist die Partie vorbei.
	if !siebzehnKannSetzen(g) {
		g.Status = "ende"
		if g.BankSitz >= 0 && g.BankSitz < len(g.Spieler) {
			g.Sieger = g.Spieler[g.BankSitz].Name
			g.Note = "Niemand kann mehr setzen – " + g.Sieger + " hat die Partie gewonnen."
		}
	}
}

// siebzehnKannSetzen meldet, ob mindestens ein Pointeur noch Geld für den
// kleinsten Einsatz hat.
func siebzehnKannSetzen(g *SiebzehnGame) bool {
	for i := range g.Spieler {
		if i == g.BankSitz || g.Spieler[i].Status == "pleite" {
			continue
		}
		if g.Spieler[i].Kasse >= siebzehnSchritt {
			return true
		}
	}
	return false
}

func siebzehnBankWeiter(g *SiebzehnGame) {
	if g.Bank > 0 {
		// Rest der Bank zurück in die Kasse des bisherigen Bankhalters.
		if g.BankSitz >= 0 && g.BankSitz < len(g.Spieler) {
			g.Spieler[g.BankSitz].Kasse += g.Bank
		}
		g.Bank = 0
	}
	naechster := siebzehnNach(g, g.BankSitz)
	if naechster < 0 {
		g.Status = "ende"
		g.Note = "Niemand kann die Bank legen – die Partie ist vorbei."
		return
	}
	g.BankSitz = naechster
	g.BankKarten = []int{}
	g.Status = "bank"
	for i := range g.Spieler {
		g.Spieler[i].Einsatz = 0
		g.Spieler[i].Karten = []int{}
	}
	g.Note = g.Spieler[naechster].Name + " hat die Bank – wie viel Geld legt er auf den Tisch?"
}

// ---------------------------------------------------------------- Spiel

func siebzehnGeben(g *SiebzehnGame) {
	g.Stapel = siebzehnDeck()
	g.BankKarten = []int{g.Stapel[0], g.Stapel[1]}
	g.Stapel = g.Stapel[2:]
	for i := range g.Spieler {
		if i == g.BankSitz {
			continue
		}
		g.Spieler[i].Karten = []int{}
		g.Spieler[i].Status = "wartet"
	}
	for runde := 0; runde < 2; runde++ {
		for i := range g.Spieler {
			if i == g.BankSitz || g.Spieler[i].Einsatz <= 0 {
				continue
			}
			g.Spieler[i].Karten = append(g.Spieler[i].Karten, g.Stapel[0])
			g.Stapel = g.Stapel[1:]
		}
	}
	g.Zug = -1
	for i := 0; i < len(g.Spieler); i++ {
		k := (g.BankSitz + 1 + i) % len(g.Spieler)
		if k == g.BankSitz || g.Spieler[k].Einsatz <= 0 {
			continue
		}
		g.Spieler[k].Status = "spielt"
		if g.Zug < 0 {
			g.Zug = k
		}
	}
	g.Status = "spiel"
	g.Note = "Die Karten liegen – " + g.Spieler[g.Zug].Name + " beginnt."

	// Alle Einsätze sind schon gefallen: Schwarzjack zählt sofort, gebustet
	// wird erst beim Zug.
	if g.Zug < 0 {
		siebzehnBankZug(g)
	}
}

func siebzehnZugWeiter(g *SiebzehnGame) {
	for i := 1; i <= len(g.Spieler); i++ {
		k := (g.Zug + i) % len(g.Spieler)
		if k == g.BankSitz {
			continue
		}
		if g.Spieler[k].Status == "spielt" {
			g.Zug = k
			return
		}
	}
	siebzehnBankZug(g)
}

func siebzehnBankZug(g *SiebzehnGame) {
	for siebzehnSumme(g.BankKarten) < 17 {
		g.BankKarten = append(g.BankKarten, g.Stapel[0])
		g.Stapel = g.Stapel[1:]
	}
	siebzehnAbrechnen(g)
}

func siebzehnAbrechnen(g *SiebzehnGame) {
	bankSumme := siebzehnSumme(g.BankKarten)
	ergebnisse := []SiebzehnErgebnis{}

	for i := range g.Spieler {
		p := &g.Spieler[i]
		if i == g.BankSitz || p.Einsatz <= 0 {
			continue
		}
		summe := siebzehnSumme(p.Karten)
		einsatz := p.Einsatz
		ergebnis := SiebzehnErgebnis{Name: p.Name, Einsatz: einsatz, Karten: append([]int{}, p.Karten...), Summe: summe}

		blackjack := len(p.Karten) == 2 && summe == 21

		switch {
		case summe > 21:
			g.Bank += einsatz
			ergebnis.Ausgang = "ueberkauft"
			ergebnis.Netto = -einsatz
			p.Status = "ueberkauft"
		case bankSumme > 21 || summe > bankSumme:
			gewinn := einsatz
			if blackjack {
				gewinn = einsatz * 2
			}
			if gewinn > g.Bank {
				gewinn = g.Bank
			}
			g.Bank -= gewinn
			p.Kasse += einsatz + gewinn
			ergebnis.Netto = gewinn
			if blackjack && gewinn > einsatz {
				ergebnis.Ausgang = "blackjack"
			} else {
				ergebnis.Ausgang = "gewonnen"
			}
			p.Status = "haelt"
		case summe == bankSumme:
			// Gleichstand: die Bank gewinnt den Einsatz.
			g.Bank += einsatz
			ergebnis.Ausgang = "gleichstand"
			ergebnis.Netto = -einsatz
			p.Status = "haelt"
		default:
			g.Bank += einsatz
			ergebnis.Ausgang = "verloren"
			ergebnis.Netto = -einsatz
			p.Status = "haelt"
		}

		p.Kasse -= einsatz // der Einsatz war nur vorgemerkt
		ergebnisse = append(ergebnisse, ergebnis)
	}

	// Rückrechnung: Einsatz wurde beim Setzen noch nicht abgebucht.
	g.Ergebnisse = ergebnisse
	g.Status = "runde"
	g.Note = "Bank " + siebzehnName(g.BankKarten) + " = " + strconv.Itoa(bankSumme) + " · Kasse der Bank: " + siebzehnEuro(g.Bank)

	// Feuer: zwei Asse beim Spieler – der übernimmt als Nächstes die Bank.
	for i := range g.Spieler {
		if i == g.BankSitz {
			continue
		}
		if len(g.Spieler[i].Karten) >= 2 && g.Spieler[i].Karten[0] == 14 && g.Spieler[i].Karten[1] == 14 {
			g.Spieler[i].Feuer = true
			g.FeuerSitz = i
			g.Note += " " + g.Spieler[i].Name + " hat Feuer – nächste Runde führt er die Bank."
		}
	}

	// Ende, wenn nur noch ein Spieler Geld hat.
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

func siebzehnAnsicht(g *SiebzehnGame) *SiebzehnAnsicht {
	if g == nil {
		return nil
	}
	a := &SiebzehnAnsicht{
		ID:         g.ID,
		HostID:     g.HostID,
		Status:     g.Status,
		Runde:      g.Runde,
		Bank:       g.Bank,
		BankSitz:   g.BankSitz,
		Belegt:     siebzehnBelegt(g),
		Zug:        g.Zug,
		Note:       g.Note,
		Sieger:     g.Sieger,
		Ergebnisse: append([]SiebzehnErgebnis{}, g.Ergebnisse...),
		FeuerSitz:  g.FeuerSitz,
	}
	// Die zweite Bankkarte bleibt verdeckt, bis die Bank zieht.
	bankKarten := append([]int{}, g.BankKarten...)
	if len(bankKarten) > 1 && g.Status == "spiel" {
		bankKarten = bankKarten[:1]
		a.BankDeckt = true
	}
	a.BankKarten = bankKarten

	for i, p := range g.Spieler {
		sicht := SiebzehnSpielerSicht{
			ID:      p.ID,
			Name:    p.Name,
			Kasse:   p.Kasse,
			Einsatz: p.Einsatz,
			Karten:  append([]int{}, p.Karten...),
			Summe:   siebzehnSumme(p.Karten),
			Status:  p.Status,
			Ready:   p.Ready,
			Feuer:   p.Feuer,
		}
		if i == g.BankSitz && p.Kasse <= 0 {
			sicht.Kasse = p.Kasse
		}
		a.Spieler = append(a.Spieler, sicht)
	}
	return a
}

// ---------------------------------------------------------------- Endpunkt

func siebzehnHandler(w http.ResponseWriter, r *http.Request) {
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
		game := siebzehnAnsicht(room.Siebzehn)
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
		if room.Siebzehn != nil && room.Siebzehn.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Es läuft schon eine Runde 17 und 4.", http.StatusConflict)
			return
		}
		room.Siebzehn = &SiebzehnGame{
			ID:        randomCode(),
			HostID:    input.ID,
			Status:    "invited",
			BankSitz:  0,
			Spieler:   []SiebzehnSpieler{siebzehnNeuerSpieler(input.ID, player.Name)},
			FeuerSitz: -1,
			Note:      "Tisch eröffnet – einer ist die Bank, die anderen spielen dagegen.",
		}
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "eröffnet eine Runde 17 und 4 – mitspielen über die Spielkarte.",
			Time:     zeit,
			SenderID: input.ID,
		})

	case "join":
		game := room.Siebzehn
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet kein Tisch auf Mitspieler.", http.StatusConflict)
			return
		}
		if siebzehnSitz(game, input.ID) >= 0 {
			break
		}
		if len(game.Spieler) >= siebzehnMaxSpieler {
			mu.Unlock()
			http.Error(w, "Dieser Tisch ist mit 6 Spielern voll.", http.StatusConflict)
			return
		}
		game.Spieler = append(game.Spieler, siebzehnNeuerSpieler(input.ID, player.Name))
		game.Note = player.Name + " setzt sich dazu."

	case "ready":
		game := room.Siebzehn
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Bereit drücken geht nur vor dem Start.", http.StatusConflict)
			return
		}
		sitz := siebzehnSitz(game, input.ID)
		if sitz < 0 {
			mu.Unlock()
			http.Error(w, "Du sitzt nicht an diesem Tisch.", http.StatusForbidden)
			return
		}
		game.Spieler[sitz].Ready = !game.Spieler[sitz].Ready
		if game.Spieler[sitz].Ready {
			game.Note = player.Name + " ist bereit."
		} else {
			game.Note = player.Name + " wartet noch."
		}

	case "start":
		game := room.Siebzehn
		if game == nil || game.Status != "invited" {
			mu.Unlock()
			http.Error(w, "Es wartet keine Runde auf den Start.", http.StatusConflict)
			return
		}
		if game.HostID != input.ID {
			mu.Unlock()
			http.Error(w, "Nur der Gastgeber kann die Bank übernehmen.", http.StatusForbidden)
			return
		}
		if len(game.Spieler) < 2 {
			mu.Unlock()
			http.Error(w, "17 und 4 braucht mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, p := range game.Spieler {
			if !p.Ready {
				mu.Unlock()
				http.Error(w, "Erst müssen alle BEREIT drücken.", http.StatusConflict)
				return
			}
		}
		game.BankSitz = 0
		game.Runde = 0
		game.Bank = 0
		game.Status = "bank"
		game.Note = game.Spieler[0].Name + " hat die Bank – wie viel Geld legt er auf den Tisch?"

	case "bank":
		game := room.Siebzehn
		if game == nil || game.Status != "bank" {
			mu.Unlock()
			http.Error(w, "Gerade legt niemand eine Bank.", http.StatusConflict)
			return
		}
		if siebzehnSitz(game, input.ID) != game.BankSitz {
			mu.Unlock()
			http.Error(w, "Nur der Bankhalter legt die Bank.", http.StatusForbidden)
			return
		}
		halter := &game.Spieler[game.BankSitz]
		if input.Betrag < siebzehnBankMin {
			mu.Unlock()
			http.Error(w, "Die Bank muss mindestens "+siebzehnEuro(siebzehnBankMin)+" haben.", http.StatusBadRequest)
			return
		}
		input.Betrag = (input.Betrag / siebzehnSchritt) * siebzehnSchritt
		if input.Betrag > halter.Kasse {
			mu.Unlock()
			http.Error(w, "So viel Geld hast du nicht in der Kasse.", http.StatusConflict)
			return
		}
		halter.Kasse -= input.Betrag
		game.Bank = input.Betrag
		siebzehnNeueRunde(game)
		game.Note = "Bank gelegt: " + siebzehnEuro(game.Bank) + " – jetzt setzen."

	case "setzen", "banco":
		game := room.Siebzehn
		if game == nil || game.Status != "setzen" {
			mu.Unlock()
			http.Error(w, "Gerade wird nicht gesetzt.", http.StatusConflict)
			return
		}
		sitz := siebzehnSitz(game, input.ID)
		if sitz < 0 || sitz == game.BankSitz {
			mu.Unlock()
			http.Error(w, "Die Bank setzt nicht gegen sich selbst.", http.StatusForbidden)
			return
		}
		p := &game.Spieler[sitz]
		if p.Status == "pleite" {
			mu.Unlock()
			http.Error(w, "Deine Kasse ist leer.", http.StatusConflict)
			return
		}
		if p.Einsatz > 0 {
			mu.Unlock()
			http.Error(w, "Du hast für diese Runde schon gesetzt.", http.StatusConflict)
			return
		}
		frei := game.Bank - siebzehnBelegt(game)
		if frei <= 0 {
			mu.Unlock()
			http.Error(w, "Die Bank ist schon voll abgedeckt – diese Runde setzt niemand mehr.", http.StatusConflict)
			return
		}
		betrag := (input.Betrag / siebzehnSchritt) * siebzehnSchritt
		if action == "banco" {
			betrag = frei
			if betrag > p.Kasse {
				betrag = (p.Kasse / siebzehnSchritt) * siebzehnSchritt
			}
			if betrag <= 0 {
				mu.Unlock()
				http.Error(w, "Für Banco fehlt dir Geld.", http.StatusConflict)
				return
			}
		}
		if betrag < siebzehnSchritt {
			mu.Unlock()
			http.Error(w, "Mindestens "+siebzehnEuro(siebzehnSchritt)+".", http.StatusBadRequest)
			return
		}
		if betrag > p.Kasse {
			mu.Unlock()
			http.Error(w, "So viel Geld hast du nicht.", http.StatusConflict)
			return
		}
		if betrag > frei {
			mu.Unlock()
			http.Error(w, "Die Bank ist nur noch mit "+siebzehnEuro(frei)+" gedeckt.", http.StatusConflict)
			return
		}
		p.Einsatz = betrag
		p.Status = "setzt"
		if action == "banco" {
			game.Note = player.Name + " ruft BANCO – " + siebzehnEuro(betrag) + " auf die ganze Bank."
		} else {
			game.Note = player.Name + " setzt " + siebzehnEuro(betrag) + "."
		}

	case "geben":
		game := room.Siebzehn
		if game == nil || game.Status != "setzen" {
			mu.Unlock()
			http.Error(w, "Gerade wird nicht gegeben.", http.StatusConflict)
			return
		}
		if siebzehnSitz(game, input.ID) != game.BankSitz {
			mu.Unlock()
			http.Error(w, "Nur der Bankhalter gibt die Karten.", http.StatusForbidden)
			return
		}
		if siebzehnBelegt(game) <= 0 {
			mu.Unlock()
			http.Error(w, "Es hat noch niemand gesetzt.", http.StatusConflict)
			return
		}
		siebzehnGeben(game)

	case "ziehen", "halten":
		game := room.Siebzehn
		if game == nil || game.Status != "spiel" {
			mu.Unlock()
			http.Error(w, "Gerade läuft kein Zug.", http.StatusConflict)
			return
		}
		sitz := siebzehnSitz(game, input.ID)
		if sitz < 0 || sitz == game.BankSitz {
			mu.Unlock()
			http.Error(w, "Die Bank zieht nach allen anderen.", http.StatusForbidden)
			return
		}
		if game.Zug != sitz || game.Spieler[sitz].Status != "spielt" {
			mu.Unlock()
			http.Error(w, "Du bist nicht an der Reihe.", http.StatusConflict)
			return
		}
		p := &game.Spieler[sitz]

		if action == "halten" {
			p.Status = "haelt"
			game.Note = player.Name + " hält bei " + strconv.Itoa(siebzehnSumme(p.Karten)) + "."
			siebzehnZugWeiter(game)
			break
		}

		p.Karten = append(p.Karten, game.Stapel[0])
		game.Stapel = game.Stapel[1:]
		summe := siebzehnSumme(p.Karten)
		if summe > 21 {
			p.Status = "ueberkauft"
			game.Note = player.Name + " überkauft sich mit " + strconv.Itoa(summe) + "."
			siebzehnZugWeiter(game)
			break
		}
		if summe == 21 {
			p.Status = "haelt"
			game.Note = player.Name + " hat 21."
			siebzehnZugWeiter(game)
			break
		}
		game.Note = player.Name + " zieht auf " + strconv.Itoa(summe) + "."

	case "naechste":
		game := room.Siebzehn
		if game == nil || game.Status != "runde" {
			mu.Unlock()
			http.Error(w, "Die Runde ist noch nicht abgerechnet.", http.StatusConflict)
			return
		}
		if siebzehnSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Runde.", http.StatusForbidden)
			return
		}
		if game.Bank <= 0 {
			siebzehnBankWeiter(game)
			break
		}
		if game.FeuerSitz >= 0 {
			// Feuer: dieser Spieler übernimmt die Bank, der Rest wird
			// ausgezahlt. Hat er nicht einmal die kleinste Bank, geht sie an
			// den Nächsten, der sie legen kann – sonst käme der Tisch nicht
			// mehr weiter.
			alterHalter := game.BankSitz
			ziel := game.FeuerSitz
			feuerName := game.Spieler[ziel].Name
			game.FeuerSitz = -1
			if game.Spieler[ziel].Status == "pleite" || game.Spieler[ziel].Kasse < siebzehnBankMin {
				ziel = siebzehnNach(game, alterHalter)
			}
			if ziel < 0 {
				game.Status = "ende"
				game.Note = feuerName + " hatte Feuer, kann die Bank aber nicht legen – die Partie ist vorbei."
				break
			}
			if game.Bank > 0 {
				game.Spieler[alterHalter].Kasse += game.Bank
			}
			game.Bank = 0
			game.BankSitz = ziel
			game.Status = "bank"
			game.Ergebnisse = nil
			for i := range game.Spieler {
				game.Spieler[i].Einsatz = 0
				game.Spieler[i].Karten = []int{}
				game.Spieler[i].Feuer = false
			}
			if game.Spieler[game.BankSitz].Name == feuerName {
				game.Note = feuerName + " hatte Feuer und übernimmt die Bank."
			} else {
				game.Note = feuerName + " hatte Feuer, kann die Bank aber nicht legen – " +
					game.Spieler[game.BankSitz].Name + " übernimmt."
			}
			break
		}
		siebzehnNeueRunde(game)

	case "abgeben":
		game := room.Siebzehn
		if game == nil || game.Status != "runde" {
			mu.Unlock()
			http.Error(w, "Die Bank lässt sich erst nach einer Runde abgeben.", http.StatusConflict)
			return
		}
		if siebzehnSitz(game, input.ID) != game.BankSitz {
			mu.Unlock()
			http.Error(w, "Nur der Bankhalter kann abgeben.", http.StatusForbidden)
			return
		}
		siebzehnBankWeiter(game)

	case "reset":
		game := room.Siebzehn
		if game == nil || game.Status != "ende" {
			mu.Unlock()
			http.Error(w, "Eine neue Partie geht erst nach dem Ende.", http.StatusConflict)
			return
		}
		if siebzehnSitz(game, input.ID) < 0 {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Runde.", http.StatusForbidden)
			return
		}
		frisch := &SiebzehnGame{
			ID:        game.ID,
			HostID:    game.HostID,
			Status:    "invited",
			BankSitz:  0,
			Spieler:   []SiebzehnSpieler{},
			FeuerSitz: -1,
			Note:      "Neue Partie – erst BEREIT drücken.",
		}
		for _, p := range game.Spieler {
			frisch.Spieler = append(frisch.Spieler, siebzehnNeuerSpieler(p.ID, p.Name))
		}
		room.Siebzehn = frisch

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := siebzehnAnsicht(room.Siebzehn)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
