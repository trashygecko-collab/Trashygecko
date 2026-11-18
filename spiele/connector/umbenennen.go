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
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.Room = strings.ToUpper(strings.TrimSpace(input.Room))
	input.ID = strings.TrimSpace(input.ID)
	input.TargetID = strings.TrimSpace(input.TargetID)
	input.Name = limitName(strings.TrimSpace(input.Name))
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
