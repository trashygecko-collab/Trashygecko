package main

import (
	"encoding/json"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

// hauptRaumCode ist die feste Kennung des einen Raums, den der Connector
// anbietet. Sie ist bewusst fest und nicht zufällig: Nach einem Neustart des
// Servers (Render startet die Instanz neu, der Speicher ist weg) finden die
// offenen Seiten denselben Raum wieder, statt still in einem neuen Raum zu
// landen – vorher leerte das den Chat und ließ die Partie im Rahmen an einem
// toten Raum hängen.
const hauptRaumCode = "TG-HALLE"

// mainRoomCode ist der eine Raum, den der Connector anbietet.
var mainRoomCode string

// ensureMainRoom legt diesen Raum an. Es gibt bewusst nur einen: der Spieler
// am Server-PC ist Admin, alle anderen kommen in denselben Raum.
func ensureMainRoom() {
	mu.Lock()
	defer mu.Unlock()
	if mainRoomCode == "" {
		mainRoomCode = hauptRaumCode
	}
	if rooms[mainRoomCode] != nil {
		return
	}
	code := mainRoomCode
	rooms[code] = &Room{
		Code:       code,
		Created:    time.Now(),
		LastActive: time.Now(),
		Players:    map[string]Player{},
		Messages:   []Message{},
		Mutes:      map[string]time.Time{},
		MsgTimes:   map[string][]time.Time{},
	}
}

// roomTitle bildet den Raumnamen aus dem Spielernamen: "Anna" -> "Annas Abenteuerland".
// Namenlose Spieler ("Namenloser 1", …) heißen dort nur "Abenteuerland" –
// aus der Anzeigenummer soll kein Besitzername werden.
func roomTitle(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || istNamenlosBezeichnung(name) {
		return "Abenteuerland"
	}
	// Bei Namen auf s, x, z oder ß klingt ein zweites s falsch: "Klaus' Abenteuerland".
	last := name[len(name)-1:]
	if last == "s" || last == "x" || last == "z" || last == "ß" {
		return name + "' Abenteuerland"
	}
	return name + "s Abenteuerland"
}

// istStandardname erkennt die Namen, die der Server früher selbst vergeben
// hat: "Boss" am Server-PC und "Spieler1", "Spieler2", … für alle anderen.
// Wer so heißt, hat sich keinen Namen ausgedacht und gilt als namenlos.
func istStandardname(name string) bool {
	name = strings.TrimSpace(name)
	if strings.EqualFold(name, "Boss") {
		return true
	}
	const basis = "Spieler"
	if len(name) <= len(basis) || !strings.EqualFold(name[:len(basis)], basis) {
		return false
	}
	for _, z := range name[len(basis):] {
		if z < '0' || z > '9' {
			return false
		}
	}
	return true
}

// bossErkennen bestimmt, wer als Boss (Admin) gilt. Lokal ist das der PC, auf
// dem der Connector läuft; die Online-Fassung setzt das beim Start um.
var bossErkennen = func(r *http.Request, room *Room) bool {
	return isServerMachine(r)
}

// isServerMachine meldet, ob die Anfrage von diesem PC kommt. Nur dort läuft der
// Connector – also sitzt dort der Host, der automatisch Admin ist.
func isServerMachine(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	for _, local := range localIPv4s() {
		if local == ip.String() {
			return true
		}
	}
	return false
}

// roomsHandler handles GET (list rooms) and POST (create room).
func roomsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		mu.Lock()
		list := []RoomInfo{}
		for code, room := range rooms {
			room.cleanPlayers()
			// Determine status
			status := room.gameStatus()
			// Determine current game name if any
			game := ""
			if room.TicTacToe != nil && (room.TicTacToe.Status == "active" || room.TicTacToe.Status == "invited") {
				game = "TicTacToe"
			} else if room.VierGewinnt != nil && (room.VierGewinnt.Status == "active" || room.VierGewinnt.Status == "invited") {
				game = "Vier gewinnt"
			} else if room.Dame != nil && (room.Dame.Status == "active" || room.Dame.Status == "invited") {
				game = "Dame"
			} else if room.Paare != nil && (room.Paare.Status == "active" || room.Paare.Status == "invited") {
				game = "Paare finden"
			} else if room.Kaese != nil && (room.Kaese.Status == "active" || room.Kaese.Status == "invited") {
				game = "Käsekästchen"
			} else if room.Sim != nil && (room.Sim.Status == "active" || room.Sim.Status == "invited") {
				game = "Sim"
			} else if room.Kalaha != nil && (room.Kalaha.Status == "active" || room.Kalaha.Status == "invited") {
				game = "Kalaha"
			} else if room.Streichholz != nil && (room.Streichholz.Status == "active" || room.Streichholz.Status == "invited") {
				game = "Streichholzspiel"
			} else if room.Muehle != nil && (room.Muehle.Status == "active" || room.Muehle.Status == "invited") {
				game = "Mühle"
			} else if room.Reversi != nil && (room.Reversi.Status == "active" || room.Reversi.Status == "invited") {
				game = "Reversi"
			} else if room.Hold != nil && (room.Hold.Status == "active" || room.Hold.Status == "invited") {
				game = "Hold the Line"
			} else if room.Skyjo != nil && (room.Skyjo.Status == "active" || room.Skyjo.Status == "invited") {
				game = "Skyjo"
			} else if room.Kniffel != nil && (room.Kniffel.Status == "active" || room.Kniffel.Status == "invited") {
				game = "Kniffel"
			} else if room.ZehnTausend != nil && (room.ZehnTausend.Status == "active" || room.ZehnTausend.Status == "invited") {
				game = "10.000"
			} else if room.Maexchen != nil && (room.Maexchen.Status == "active" || room.Maexchen.Status == "invited") {
				game = "Lügen Mäxchen"
			} else if room.Boston != nil && (room.Boston.Status == "active" || room.Boston.Status == "invited") {
				game = "Auf nach Boston"
			} else if room.Chicago != nil && (room.Chicago.Status == "active" || room.Chicago.Status == "invited" || room.Chicago.Status == "runde") {
				game = "Chicago"
			} else if room.ShipCrew != nil && (room.ShipCrew.Status == "active" || room.ShipCrew.Status == "invited" || room.ShipCrew.Status == "runde") {
				game = "Ship Captain Crew"
			} else if room.Halma != nil && (room.Halma.Status == "active" || room.Halma.Status == "invited") {
				game = "Halma"
			} else if room.Schach != nil && (room.Schach.Status == "active" || room.Schach.Status == "invited") {
				game = "Schach"
			} else if room.Go != nil && (room.Go.Status == "active" || room.Go.Status == "invited") {
				game = "Go"
			} else if room.Backgammon != nil && (room.Backgammon.Status == "active" || room.Backgammon.Status == "invited") {
				game = "Backgammon"
			} else if room.Sechsundsechzig != nil && (room.Sechsundsechzig.Status == "spiel" || room.Sechsundsechzig.Status == "invited") {
				game = "Sechsundsechzig"
			}
			list = append(list, RoomInfo{Code: code, Players: len(room.Players), Created: uhrzeitKurz(room.Created), CreatedAt: room.Created, Owner: room.OwnerName, Status: status, Game: game, Name: room.Name})
		}
		mu.Unlock()
		sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.Before(list[j].CreatedAt) })
		jsonResponse(w, list)
		return
	}

	if r.Method == http.MethodPost {
		var input struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&input)
		input.ID = kennung(r)
		pName := strings.TrimSpace(input.Name)
		if pName == "" {
			pName = "Namenloser"
		}
		pName = limitName(pName)

		// Es gibt genau einen gemeinsamen Raum. Weitere Räume entstehen nur
		// noch auf private Einladung (/api/einladung) – nie auf Zuruf.
		ensureMainRoom()
		mu.Lock()
		room := rooms[mainRoomCode]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht verfügbar", http.StatusServiceUnavailable)
			return
		}
		code := room.Code
		displayName := room.Name
		if displayName == "" {
			// Nur für die Anzeige – der Name wird beim Beitreten gesetzt.
			displayName = roomTitle(pName)
		}
		mu.Unlock()
		jsonResponse(w, map[string]string{"code": code, "name": displayName})
		return
	}

	http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
}

// presence registers or updates a player's presence in a room.
func presence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erforderlich", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		Room string `json:"room"`
		ID   string `json:"id"`
		Name string `json:"name"`
		// Active meldet, ob der Spieler seit dem letzten Lebenszeichen etwas
		// getan hat (Klick, Tastendruck). Ohne das zählt er als abwesend.
		Active bool `json:"active"`
		// Token ist das Geheimnis des Platzes; beim ersten Beitritt leer.
		Token string `json:"token"`
		// Umbenennen setzt die Seite nur, wenn der Spieler seinen Namen
		// wirklich ändert. Ohne das Zeichen bleibt der Name stehen – sonst
		// würde jedes Lebenszeichen eine Änderung des Admins zurücksetzen.
		Umbenennen bool `json:"umbenennen"`
		// Neu legt einen unbekannten Platz ausdrücklich an. Fehlt das Zeichen,
		// weist der Server eine unbekannte Kennung ab (409), statt still einen
		// weiteren Spieler in den Raum zu setzen.
		Neu bool `json:"neu"`
		// Geraet ist das zweite Geheimnis dieses Browsers (spiele/token.js).
		// Geht das Token verloren, holt es denselben Platz zurück – aber nur
		// innerhalb von zehn Minuten und nur für genau dieses Gerät.
		Geraet string `json:"geraet"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.Room = raumAusKontext(r)
	input.ID = kennung(r)
	input.Name = strings.TrimSpace(input.Name)
	// Das Gerät hängt sein Token an die Adresse (spiele/token.js), nicht in
	// den Rumpf. Ohne diese Zeile galt jeder bekannte Platz als fremdes Gerät
	// – jedes Lebenszeichen und jedes Neuladen lief in eine 403.
	if input.Token == "" {
		input.Token = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if input.Geraet == "" {
		input.Geraet = strings.TrimSpace(r.URL.Query().Get("geraet"))
	}
	if input.ID == "" || input.Room == "" {
		http.Error(w, "Spielraum oder Spielerkennung fehlt", http.StatusBadRequest)
		return
	}
	input.Name = limitName(input.Name)
	// Alte Automatiknamen ("Boss", "Spieler1", "Spieler2" …) zählen wie ein
	// leerer Name: damit hat sich niemand vorgestellt. Solche Plätze heißen
	// jetzt "Namenloser 1", "Namenloser 2" … – auch wenn der Name schon
	// gespeichert ist.
	if istStandardname(input.Name) {
		input.Name = ""
		input.Umbenennen = true
	}

	mu.Lock()
	room := rooms[input.Room]
	if room == nil && input.Room == mainRoomCode {
		// Der gemeinsame Raum überlebt einen Neustart des Servers: Er wird
		// hier neu angelegt, damit die offenen Seiten mit derselben Kennung
		// wieder hineinkommen (Chat und Partie sind mit dem Speicher weg).
		mu.Unlock()
		ensureMainRoom()
		mu.Lock()
		room = rooms[input.Room]
	}
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	// Ein Spieler kann in mehreren Räumen Mitglied sein: Die Halle bleibt, wer
	// eine private Runde betritt, sitzt dort zusätzlich. Ein Lebenszeichen
	// löscht deshalb hier keinen Platz in einem anderen Raum und beendet dort
	// auch keine Partie. Aufgeräumt wird erst beim ausdrücklichen Verlassen
	// (/api/leave, /api/einladung "zurueck") oder wenn eine Frist abläuft.
	room.cleanPlayers()
	amServer := bossErkennen(r, room)
	player, known := room.Players[input.ID]
	// Ein bekannter Platz gehört dem Gerät, das ihn angelegt hat: Wer die
	// Kennung errät oder aus /api/players abschreibt, kommt hier nicht weiter.
	// Eine gültige Sitzung desselben Spielers zählt als Ausweis – so darf ein
	// zweiter Tab oder ein zweites Gerät denselben Platz benutzen, ohne ihn zu
	// übernehmen.
	if known && (player.Token == "" || player.Token != strings.TrimSpace(input.Token)) {
		// Das Token kann verloren gehen (Speicher geleert, alter Tab, neues
		// Browserfenster). Dann holt die Gerätekennung denselben Platz zurück:
		// gleiche Kennung, gleicher Name, kein Nummernanhang. Ohne passende
		// Gerätekennung bleibt es bei der Abweisung – und es entsteht dabei
		// keine neue Identität.
		if !sitzungPasst(input.ID, strings.TrimSpace(input.Token)) &&
			(!geraetPasst(player, input.Geraet, time.Now()) || !tokenBrauchbar(strings.TrimSpace(input.Token))) {
			mu.Unlock()
			http.Error(w, "Diese Spielerkennung gehört zu einem anderen Gerät.", http.StatusForbidden)
			return
		}
		if player.Token != strings.TrimSpace(input.Token) {
			player.Token = strings.TrimSpace(input.Token)
		}
	}

	// Ein unbekannter Platz entsteht nur auf ausdrücklichen Wunsch (neu).
	// Sonst legte jeder Ruf mit einer alten oder erfundenen Kennung still einen
	// weiteren Spieler an – nach einem Server-Neustart oder einem verlorenen
	// Token sammelten sich so Plätze an, und der eigene Name war "vergeben".
	if !known && !input.Neu {
		mu.Unlock()
		http.Error(w, "Diese Spielerkennung ist hier nicht bekannt – bitte neu beitreten.", http.StatusConflict)
		return
	}
	// Wer ohne eigenen Namen hereinkommt, heißt "Namenloser 1"; der nächste
	// Namenlose bekommt "Namenloser 2" und so fort. Umbenennen geht jederzeit.
	// Ein bekannter Spieler behält seinen Namen; geändert wird er nur, wenn
	// er es selbst verlangt (umbenennen) oder der Admin ihn umbenannt hat.
	name := player.Name
	if !known || input.Umbenennen {
		if input.Name == "" || istNamenlosBezeichnung(input.Name) {
			// Kein eigener Name – auch nicht die Bezeichnung, die der Server
			// selbst vergeben hat und die die Seite im Feld anzeigt. Ein
			// bekannter Namenloser behält seine Nummer, solange er online ist;
			// sonst nummerierte jeder Ruf ihn neu durch.
			if !known || !istNamenlosBezeichnung(player.Name) {
				name = room.namenlosName(input.ID)
			}
		} else {
			// Doppelte Namen bekommen eine Nummer – sonst könnte sich jeder
			// als der Admin ausgeben.
			name = room.uniqueName(input.Name, input.ID)
		}
	}
	// Admin ist der Spieler am Server-PC. Seine Anfrage kommt von einer lokalen
	// Adresse, egal ob über localhost oder die LAN-IP aufgerufen wird.
	if room.OwnerID == "" || (room.OwnerID != input.ID && amServer) {
		room.OwnerID = input.ID
		room.OwnerName = name
		room.Name = roomTitle(name)
	}
	if !known {
		// Mit Absender, damit der Name im Chat steht und angeklickt werden kann.
		text := "betritt den Raum."
		if amServer {
			text = "ist da!"
		}
		room.addMessage(Message{
			Name:     name,
			Text:     text,
			Time:     uhrzeit(time.Now()),
			SenderID: input.ID,
		})
		// Begrüßung nur für den Neuen – der Raum flüstert ihn an, die anderen
		// sehen davon nichts.
		room.addMessage(Message{
			Name:    "Der Raum",
			Text:    "Hallo und herzlich willkommen! Spieler anflüstern mit /f Name",
			Time:    uhrzeit(time.Now()),
			ToID:    input.ID,
			ToName:  name,
			Whisper: true,
		})
	} else if player.Name != "" && player.Name != name {
		// Umbenennen einmal im Chat bekanntgeben – mit dem alten Namen als
		// Absender, damit die Zeile "Anna: heißt jetzt Alex." lesbar bleibt.
		room.addMessage(Message{
			Name:     player.Name,
			Text:     "heißt jetzt " + name + ".",
			Time:     uhrzeit(time.Now()),
			SenderID: input.ID,
		})
	}
	player.Name = name
	if !known {
		// Bringt das Gerät sein Geheimnis selbst mit (spiele/token.js erzeugt es
		// vor dem ersten Ruf), gilt das. Zwei fast gleichzeitige Beitritte
		// desselben Geräts schicken dann dasselbe Token – der zweite Ruf läuft
		// nicht mehr in „gehört zu einem anderen Gerät“.
		if mitgebracht := strings.TrimSpace(input.Token); tokenBrauchbar(mitgebracht) {
			player.Token = mitgebracht
		} else {
			player.Token = zufallsToken()
		}
	}
	// Die Gerätekennung gehört zum Platz, nicht zur Sitzung: Sie bleibt auch
	// nach einem neuen Token stehen (gleiches Browser-Profil).
	if geraetBrauchbar(input.Geraet) {
		player.Geraet = input.Geraet
	}
	player.Last = time.Now()
	if player.LastAction.IsZero() || input.Active {
		player.LastAction = time.Now()
		player.AfkWarned = false
	}
	room.Players[input.ID] = player
	// Ein Lebenszeichen hält den Spieler in jedem Raum frisch, in dem er
	// Mitglied ist. Wer eine private Runde spielt, bleibt damit in der Halle
	// anwesend, und ein einziger Ruf hält alle seine Tische am Leben – ohne
	// das fiel er nach presenceTimeout aus der Halle, sobald er den Raum
	// wechselte, mitten aus Chat und Partie.
	jetzt := time.Now()
	fuerAllePlaetze(input.ID, func(andere *Room, platz *Player) {
		if andere.Code == room.Code {
			return
		}
		platz.Last = jetzt
		if input.Active {
			platz.LastAction = jetzt
			platz.AfkWarned = false
		}
		andere.LastActive = jetzt
	})
	// Identität und Sitzung getrennt vom Platz führen: Der Platz kann verfallen
	// (Raum voll, Tab zu), das Profil und die Sitzung bleiben. Damit ist ein
	// zweiter Tab derselbe Spieler und kein Fremder.
	spielerMerken(input.ID, player.Name)
	sitzungMerken(input.ID, player.Geraet, player.Token)
	room.LastActive = time.Now()
	code := room.Code
	mu.Unlock()
	// Der vergebene Name geht zurück, damit die Seite ihn anzeigen kann.
	// Das Token geht nur an dieses Gerät zurück – nie an andere Spieler.
	jsonResponse(w, map[string]string{"name": name, "room": code, "token": player.Token})
}

// geraetFrist: so lange holt die Gerätekennung einen Platz zurück, dessen
// Token verloren ging. Danach gilt nur noch das Token.
const geraetFrist = 10 * time.Minute

// geraetPasst sagt, ob dieses Gerät den Platz schon einmal benutzt hat und die
// Rückgabefrist noch läuft. Die Kennung allein (für alle sichtbar) genügt nie.
func geraetPasst(player Player, geraet string, now time.Time) bool {
	if !geraetBrauchbar(geraet) || player.Geraet != geraet {
		return false
	}
	return !player.Last.IsZero() && now.Sub(player.Last) <= geraetFrist
}

// geraetBrauchbar prüft eine Gerätekennung: nur Buchstaben und Ziffern, 16 bis
// 128 Zeichen. Der Browser erzeugt sie wie ein Token.
func geraetBrauchbar(geraet string) bool {
	if len(geraet) < 16 || len(geraet) > 128 {
		return false
	}
	for _, r := range geraet {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

// tokenBrauchbar prüft ein vom Gerät mitgebrachtes Geheimnis: nur Buchstaben
// und Ziffern, 16 bis 64 Zeichen. Alles andere ersetzt der Server selbst.
func tokenBrauchbar(token string) bool {
	if len(token) < 16 || len(token) > 64 {
		return false
	}
	for _, r := range token {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

// playersHandler returns list of players in a room and room meta (like ownerId).
func playersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET erforderlich", http.StatusMethodNotAllowed)
		return
	}
	code := strings.ToUpper(r.URL.Query().Get("room"))
	mu.Lock()
	room := rooms[code]
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	room.cleanPlayers()
	// Anwesend unterscheidet „ist da“ von „steht nur noch im Raum“ – der
	// Connector braucht das zum Beispiel für die Zwei-Spieler-Regel.
	jetzt := time.Now()
	list := make([]SpielerZeile, 0, len(room.Players))
	for id, player := range room.Players {
		list = append(list, SpielerZeile{ID: id, Name: player.Name, Anwesend: player.anwesend(jetzt),
			Spiel: room.aktivesSpiel(id)})
	}
	ownerID := room.OwnerID
	mu.Unlock()
	sort.Slice(list, func(i, j int) bool { return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name) })
	w.Header().Set("X-Room-Owner", ownerID)
	jsonResponse(w, list)
}
