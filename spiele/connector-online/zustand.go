package main

// Ein Ruf statt sieben: /api/zustand liefert alles, was die Connector-Seite im
// Takt braucht – Lebenszeichen, Chat, Spieler, Einladungen, offenes Angebot,
// Stummschaltungen, laufende Partie und Hinweise. Vorher fragte die Seite mit
// sieben Timern einzeln (am sichtbaren Fenster gemessen 205–225 Rufe je
// Minute); jetzt ist es ein Ruf alle fünf Sekunden.
//
// Der Ruf legt nichts an: Wer hier nicht im Raum steht, bekommt 409 und meldet
// sich über den Beitritt (/api/presence) an. Damit gilt weiter: eine neue
// Identität entsteht ausschließlich beim ausdrücklichen Beitritt.

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"
)

// serverStart ist der Beginn dieses Serverprozesses. Er steht in jeder Antwort
// des zentralen Statusrufs: Ändert er sich, wurde der Server neu gestartet –
// dann sind Chat und laufende Partien aus dem Speicher weg. Die Seite sagt das
// ausdrücklich, statt es stillschweigend hinzunehmen.
var serverStart = time.Now().Unix()

func zustandHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erforderlich", http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		// Active meldet, ob der Spieler seit dem letzten Ruf etwas getan hat.
		// Nur dann gilt er als wach – sonst bliebe die „eingeschlafen“-Meldung
		// für immer aus.
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	id := kennung(r)
	code := raumAusKontext(r)
	if id == "" || code == "" {
		http.Error(w, "Spielraum oder Spielerkennung fehlt", http.StatusBadRequest)
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
	player, bekannt := room.Players[id]
	if !bekannt {
		mu.Unlock()
		http.Error(w, "Diese Spielerkennung ist hier nicht bekannt – bitte neu beitreten.", http.StatusConflict)
		return
	}
	jetzt := time.Now()
	player.Last = jetzt
	if in.Active {
		player.LastAction = jetzt
		player.AfkWarned = false
	}
	room.Players[id] = player
	// Ein Lebenszeichen gilt für alle Räume, in denen der Spieler Mitglied ist
	// (Etappe 2): Die Halle bleibt frisch, während er an einem privaten Tisch
	// sitzt.
	fuerAllePlaetze(id, func(andere *Room, platz *Player) {
		if andere.Code == room.Code {
			return
		}
		platz.Last = jetzt
		if in.Active {
			platz.LastAction = jetzt
			platz.AfkWarned = false
		}
		andere.LastActive = jetzt
	})

	// Ab hier wird nur gelesen. Alles in einem Zug, damit die Antwort einen
	// einzigen, in sich stimmigen Stand zeigt.
	// Flüstern sieht nur, wer es schreibt oder bekommt. /api/messages hat das
	// immer so gehalten; der zentrale Ruf darf es nicht umgehen – sonst stünde
	// jede private Einladung im Chat aller.
	msgs := make([]Message, 0, len(room.Messages))
	for _, m := range room.Messages {
		if !m.Whisper || m.SenderID == id || m.ToID == id {
			msgs = append(msgs, m)
		}
	}

	zeilen := make([]SpielerZeile, 0, len(room.Players))
	for pid, p := range room.Players {
		zeilen = append(zeilen, SpielerZeile{ID: pid, Name: p.Name,
			Anwesend: p.anwesend(jetzt), Spiel: room.aktivesSpiel(pid)})
	}
	sort.Slice(zeilen, func(i, j int) bool {
		return strings.ToLower(zeilen[i].Name) < strings.ToLower(zeilen[j].Name)
	})

	owner := room.OwnerID
	angebote := angeboteKopie(room)
	einladungAufraeumen(jetzt)
	eingehend, ausgehend := einladungenFuer(id)
	// Stummschaltungen sieht nur der Raum-Besitzer – die Liste ist eine
	// Verwaltungsansicht, kein Aushang für alle.
	mutes := []map[string]string{}
	if owner == id {
		mutes = mutesListe(room, jetzt)
	}
	hinweis := room.hinweisNehmen(id)
	inGame := room.playerInGame(id)
	aktiv := room.aktivesSpiel(id)
	mu.Unlock()

	// Bereitschaft je offenem Tisch – eine Abfrage, alle Tische. Der Stand
	// kommt aus den Spielen selbst (Etappe 5), damit er hier nicht doppelt
	// gepflegt wird und nicht auseinanderläuft.
	bereit := map[string][]string{}
	for _, a := range angebote {
		if _, schon := bereit[a.Game]; schon {
			continue
		}
		// Spiele ohne Bereit-Knopf brauchen keinen Spielstandsabruf.
		if !bereitSpiele[a.Game] {
			bereit[a.Game] = []string{}
			continue
		}
		bereit[a.Game] = bereitFuer(code, a.Game, a.VonID)
	}

	jsonResponse(w, map[string]any{
		"ok":       true,
		"name":     player.Name,
		"ownerId":  owner,
		"messages": msgs,
		"players":  zeilen,
		"angebote": angebote,
		"bereit":   bereit,
		"einladungen": map[string]any{
			"eingehend": eingehend,
			"ausgehend": ausgehend,
		},
		"mutes":   mutes,
		"inGame":  inGame,
		"start":   serverStart,
		"aktiv":   aktiv,
		"hinweis": hinweis,
	})
}
