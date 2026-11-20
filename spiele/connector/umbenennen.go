package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// umbenennenHandler lässt den Admin (Raum-Besitzer) den Namen eines Spielers
// ändern – auch mitten in einer laufenden Partie. Der neue Name gilt sofort:
// er steht im Raum, in jeder Partie des Raums und damit beim nächsten Abruf
// auf jeder Spielseite. Die Partie bleibt stehen, es wird nichts neu verteilt.
func umbenennenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erforderlich", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		Room     string `json:"room"`
		ID       string `json:"id"`
		TargetID string `json:"targetId"`
		Name     string `json:"name"`
		// Vorher ist der Name, den der Admin im Dialog gesehen hat. Weicht er
		// inzwischen ab, hat jemand anders denselben Spieler schon umbenannt –
		// dann bleibt die spätere Anfrage ohne Wirkung, statt die Änderung zu
		// überschreiben. Leer heißt: keine Prüfung (alte Aufrufer).
		Vorher string `json:"vorher"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.Room = strings.ToUpper(strings.TrimSpace(input.Room))
	input.ID = strings.TrimSpace(input.ID)
	input.TargetID = strings.TrimSpace(input.TargetID)
	input.Name = limitName(strings.TrimSpace(input.Name))
	input.Vorher = strings.TrimSpace(input.Vorher)
	if input.Room == "" || input.ID == "" || input.TargetID == "" {
		http.Error(w, "Parameter fehlen", http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		http.Error(w, "Der Name darf nicht leer sein.", http.StatusBadRequest)
		return
	}

	mu.Lock()
	room := rooms[input.Room]
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	room.cleanPlayers()
	if room.OwnerID != input.ID {
		mu.Unlock()
		http.Error(w, "Nur der Admin kann Namen ändern.", http.StatusForbidden)
		return
	}
	target, ok := room.Players[input.TargetID]
	if !ok {
		mu.Unlock()
		http.Error(w, "Diesen Spieler gibt es nicht mehr.", http.StatusNotFound)
		return
	}
	alt := target.Name
	if input.Vorher != "" && input.Vorher != alt {
		mu.Unlock()
		http.Error(w, "Der Name wurde inzwischen geändert – bitte neu öffnen.", http.StatusConflict)
		return
	}
	// Doppelte Namen bekommen eine Nummer – sonst könnte sich der Admin als
	// Mitspieler ausgeben.
	neu := room.uniqueName(input.Name, input.TargetID)
	if neu != alt {
		target.Name = neu
		room.Players[input.TargetID] = target
		if room.OwnerID == input.TargetID {
			room.OwnerName = neu
			room.Name = roomTitle(neu)
		}
		// Partien tragen den Namen mit: Die Spielseiten fragen im Sekundentakt
		// und zeigen ihn damit sofort, ohne dass die Partie neu beginnt.
		room.spielerUmbenennen(input.TargetID, neu)
		// Der Chatverlauf bleibt derselbe – nur der Name an den alten Zeilen
		// wird nachgezogen, damit nicht zwei Namen für einen Spieler stehen.
		room.messagesUmbenennen(input.TargetID, alt, neu)
		room.addMessage(Message{
			Name:     alt,
			Text:     "heißt jetzt " + neu + ".",
			Time:     time.Now().Format("15:04:05"),
			SenderID: input.TargetID,
		})
		room.LastActive = time.Now()
	}
	mu.Unlock()
	jsonResponse(w, map[string]string{"name": neu, "vorher": alt})
}

// messagesUmbenennen zieht den neuen Namen durch den vorhandenen Chatverlauf:
// Absenderzeilen (auch „verlässt den Raum“) tragen die Spielerkennung, ebenso
// Flüsterzeilen als Empfänger. Systemzeilen ohne Kennung, die mit dem alten
// Namen beginnen (etwa „X ist wohl eingeschlafen“), werden am Wortanfang
// ersetzt. Die Nachricht selbst und ihr Zeitpunkt bleiben unverändert.
func (r *Room) messagesUmbenennen(id, alt, neu string) {
	if id == "" || neu == "" || alt == "" || alt == neu {
		return
	}
	for i := range r.Messages {
		m := &r.Messages[i]
		if m.SenderID == id {
			m.Name = neu
		}
		if m.ToID == id {
			m.ToName = neu
		}
		if m.SenderID == "" && m.Name == "" && strings.HasPrefix(m.Text, alt+" ") {
			m.Text = neu + strings.TrimPrefix(m.Text, alt)
		}
	}
}

// spielerUmbenennen schreibt einen neuen Namen in alle Partien des Raums. Es
// ist die einzige Stelle, die Namen in laufenden Spielen ändert – der
// Spielstand selbst (Figuren, Karten, Züge) bleibt unberührt.
func (r *Room) spielerUmbenennen(id, neu string) {
	if id == "" || neu == "" {
		return
	}
	// duell: zwei feste Plätze (Rot/Blau).
	duell := func(pid, pname, qid, qname *string) {
		if *pid == id {
			*pname = neu
		}
		if *qid == id {
			*qname = neu
		}
	}
	if g := r.TicTacToe; g != nil {
		duell(&g.XID, &g.XName, &g.OID, &g.OName)
	}
	if g := r.VierGewinnt; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Dame; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Paare; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Kaese; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Sim; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Kalaha; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Streichholz; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Muehle; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Reversi; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Hold; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Skyjo; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Schach; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Go; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Backgammon; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	if g := r.Tetris; g != nil {
		duell(&g.RedID, &g.RedName, &g.BlueID, &g.BlueName)
	}
	// sitze: Tische mit Spielerlisten.
	if g := r.MauMau; g != nil {
		benenneZeiger(g.Players, id, neu, func(p *MauPlayer) string { return p.ID }, func(p *MauPlayer) *string { return &p.Name })
	}
	if g := r.Uno; g != nil {
		benenneZeiger(g.Players, id, neu, func(p *UnoPlayer) string { return p.ID }, func(p *UnoPlayer) *string { return &p.Name })
	}
	if g := r.Schwimmen; g != nil {
		benenneZeiger(g.Players, id, neu, func(p *SchwimmPlayer) string { return p.ID }, func(p *SchwimmPlayer) *string { return &p.Name })
	}
	if g := r.Arschloch; g != nil {
		benenneZeiger(g.Players, id, neu, func(p *ArschPlayer) string { return p.ID }, func(p *ArschPlayer) *string { return &p.Name })
	}
	if g := r.Kniffel; g != nil {
		benenne(g.Players, id, neu, func(p *KniffelPlayer) string { return p.ID }, func(p *KniffelPlayer) *string { return &p.Name })
	}
	if g := r.ZehnTausend; g != nil {
		benenne(g.Players, id, neu, func(p *ZTPlayer) string { return p.ID }, func(p *ZTPlayer) *string { return &p.Name })
	}
	if g := r.Maexchen; g != nil {
		benenne(g.Players, id, neu, func(p *MXPlayer) string { return p.ID }, func(p *MXPlayer) *string { return &p.Name })
		if g.AnsageVon == id {
			g.AnsageVonName = neu
		}
	}
	if g := r.Boston; g != nil {
		benenne(g.Players, id, neu, func(p *BostonPlayer) string { return p.ID }, func(p *BostonPlayer) *string { return &p.Name })
	}
	if g := r.ShipCrew; g != nil {
		benenne(g.Players, id, neu, func(p *ScPlayer) string { return p.ID }, func(p *ScPlayer) *string { return &p.Name })
	}
	if g := r.Halma; g != nil {
		benenne(g.Spieler, id, neu, func(p *HalmaSpieler) string { return p.ID }, func(p *HalmaSpieler) *string { return &p.Name })
	}
	if g := r.Chicago; g != nil {
		benenne(g.Players, id, neu, func(p *ChPlayer) string { return p.ID }, func(p *ChPlayer) *string { return &p.Name })
		if g.HochID == id {
			g.HochName = neu
		}
		if g.TiefID == id {
			g.TiefName = neu
		}
	}
	if g := r.Menschaerger; g != nil {
		benenne(g.Players, id, neu, func(p *MaenPlayer) string { return p.ID }, func(p *MaenPlayer) *string { return &p.Name })
	}
	if g := r.Rechen; g != nil {
		benenneZeiger(g.Players, id, neu, func(p *RechenPlayer) string { return p.ID }, func(p *RechenPlayer) *string { return &p.Name })
	}
	if g := r.Poker; g != nil {
		benenne(g.Spieler, id, neu, func(p *PokerSpieler) string { return p.ID }, func(p *PokerSpieler) *string { return &p.Name })
	}
	if g := r.Cabo; g != nil {
		benenne(g.Spieler, id, neu, func(p *CaboSpieler) string { return p.ID }, func(p *CaboSpieler) *string { return &p.Name })
	}
	if g := r.Siebzehn; g != nil {
		benenne(g.Spieler, id, neu, func(p *SiebzehnSpieler) string { return p.ID }, func(p *SiebzehnSpieler) *string { return &p.Name })
	}
	if g := r.Baccara; g != nil {
		benenne(g.Spieler, id, neu, func(p *BaccaraSpieler) string { return p.ID }, func(p *BaccaraSpieler) *string { return &p.Name })
	}
	if g := r.Tron; g != nil {
		benenne(g.Spieler, id, neu, func(p *TronSpieler) string { return p.ID }, func(p *TronSpieler) *string { return &p.Name })
	}
	if g := r.Bomberman; g != nil {
		benenne(g.Spieler, id, neu, func(p *BomberSpieler) string { return p.ID }, func(p *BomberSpieler) *string { return &p.Name })
	}
	if g := r.Bubble; g != nil {
		benenne(g.Spieler, id, neu, func(p *BubbleSpieler) string { return p.ID }, func(p *BubbleSpieler) *string { return &p.Name })
	}
	if g := r.Snake; g != nil {
		benenne(g.Spieler, id, neu, func(p *SnakeSpieler) string { return p.ID }, func(p *SnakeSpieler) *string { return &p.Name })
	}
	if g := r.Pong; g != nil {
		benenne(g.Spieler, id, neu, func(p *PongSpieler) string { return p.ID }, func(p *PongSpieler) *string { return &p.Name })
	}
	if g := r.Schiffe; g != nil {
		benenne(g.Spieler, id, neu, func(p *SchiffeSpieler) string { return p.ID }, func(p *SchiffeSpieler) *string { return &p.Name })
	}
	if g := r.Galgen; g != nil {
		benenne(g.Spieler, id, neu, func(p *GalgenSpieler) string { return p.ID }, func(p *GalgenSpieler) *string { return &p.Name })
	}
	if g := r.Bingo; g != nil {
		benenne(g.Spieler, id, neu, func(p *BingoSpieler) string { return p.ID }, func(p *BingoSpieler) *string { return &p.Name })
	}
	if g := r.Wortsuche; g != nil {
		benenne(g.Spieler, id, neu, func(p *WortsucheSpieler) string { return p.ID }, func(p *WortsucheSpieler) *string { return &p.Name })
	}
	if g := r.Domino; g != nil {
		benenne(g.Spieler, id, neu, func(p *DominoSpieler) string { return p.ID }, func(p *DominoSpieler) *string { return &p.Name })
	}
	if g := r.Schlangen; g != nil {
		benenne(g.Spieler, id, neu, func(p *SchlangenSpieler) string { return p.ID }, func(p *SchlangenSpieler) *string { return &p.Name })
	}
	if g := r.Pig; g != nil {
		benenne(g.Spieler, id, neu, func(p *PigSpieler) string { return p.ID }, func(p *PigSpieler) *string { return &p.Name })
	}
	if g := r.Gaense; g != nil {
		benenne(g.Spieler, id, neu, func(p *GaenseSpieler) string { return p.ID }, func(p *GaenseSpieler) *string { return &p.Name })
	}
	if g := r.Sechsundsechzig; g != nil {
		benenne(g.Spieler, id, neu, func(p *SechsundsechzigSpieler) string { return p.ID }, func(p *SechsundsechzigSpieler) *string { return &p.Name })
	}
	// Ein offenes Angebot nennt den Ersteller beim Namen.
	if r.Angebot != nil && r.Angebot.VonID == id {
		r.Angebot.VonName = neu
	}
}

// benenne setzt den Namen des Spielers mit dieser Kennung in einer Spielerliste.
// Der Zeiger zeigt auf den Listeneintrag selbst – eine Kopie zu ändern würde
// den Spielstand nicht erreichen.
func benenne[T any](list []T, id, neu string, kennung func(*T) string, name func(*T) *string) {
	for i := range list {
		eintrag := &list[i]
		if kennung(eintrag) == id {
			*name(eintrag) = neu
		}
	}
}

// benenneZeiger ist dasselbe für Runden, die ihre Spieler als Zeiger halten.
func benenneZeiger[T any](list []*T, id, neu string, kennung func(*T) string, name func(*T) *string) {
	for _, p := range list {
		if p != nil && kennung(p) == id {
			*name(p) = neu
		}
	}
}
