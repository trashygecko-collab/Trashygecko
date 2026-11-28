package main

/* ===========================================================================
   BOT FÜR ARSCHLOCH

   Arschloch braucht mindestens drei Spieler. Der Bot-Betrieb setzt deshalb
   zwei Bots an den Tisch; beide ziehen ausschließlich durch arschAktion – also
   durch dieselbe Prüfung wie der Mensch (gleicher Wert, gleiche Anzahl, höher
   als der Stich, Passen nur wenn ein Stich liegt). Sie sehen nur ihr eigenes
   Blatt und den offenen Stich.

   Stufen:
     leicht  – legt irgendeine erlaubte Gruppe und passt auch mal, obwohl es
               ginge; eröffnet mit einer beliebigen Gruppe.
     mittel  – legt immer die niedrigste erlaubte Gruppe, eröffnet mit einem
               niedrigen Paar/Drilling, wenn es eins hat.
     schwer  – eröffnet mit der größten niedrigen Gruppe (Karten loswerden),
               hält hohe Karten (ab Dame) zurück, solange der Stich auch ohne
               sie entschieden werden kann, und passt dann bewusst.
   =========================================================================== */

import (
	"math/rand"
	"sort"
	"time"
)

// arschIDs sind alle Plätze am Tisch – Menschen und Bots.
func arschIDs(g *ArschlochGame) []string {
	if g == nil {
		return nil
	}
	ids := make([]string, 0, len(g.Players))
	for _, p := range g.Players {
		ids = append(ids, p.ID)
	}
	return ids
}

// arschBotIstBot sagt, ob an diesem Platz ein Bot sitzt.
func arschBotIstBot(g *ArschlochGame, seat int) bool {
	if g == nil || seat < 0 || seat >= len(g.Players) {
		return false
	}
	id := g.Players[seat].ID
	for _, b := range g.Bots {
		if b == id {
			return true
		}
	}
	return false
}

// arschBotDran sagt, ob ein Bot handeln muss: er ist am Zug. Nach der Runde
// rührt er sich nicht – die Rangliste bleibt stehen, bis ein Mensch die
// nächste Runde startet.
func arschBotDran(g *ArschlochGame) bool {
	if g == nil || g.Status != "active" || len(g.Bots) == 0 {
		return false
	}
	return g.Phase == "playing" && arschBotIstBot(g, g.Current)
}

// arschBotPause ist die Denkpause zwischen zwei Zügen.
func arschBotPause(g *ArschlochGame) time.Duration {
	return botWartezeit(g.BotStufe)
}

// arschBotTreiben lässt genau einen Bot nach der Denkpause handeln. Aufgerufen
// wird es beim Abholen des Spielstands (GET), damit die Partie auch läuft,
// wenn der Mensch nur zusieht.
func arschBotTreiben(g *ArschlochGame, jetzt time.Time) bool {
	if !arschBotDran(g) {
		return false
	}
	if g.BotFrist.IsZero() {
		g.BotFrist = jetzt.Add(arschBotPause(g))
		return false
	}
	if jetzt.Before(g.BotFrist) {
		return false
	}
	g.BotFrist = time.Time{}
	return arschBotSchritt(g)
}

// arschBotSchritt macht genau einen Schritt: Karten legen oder passen.
func arschBotSchritt(g *ArschlochGame) bool {
	if g.Phase != "playing" {
		return false
	}
	seat := g.Current
	if !arschBotIstBot(g, seat) {
		return false
	}
	name := g.Players[seat].Name
	karten := arschBotWahl(g, seat)
	if len(karten) == 0 {
		if len(g.Trick) == 0 {
			// Sicherheitsnetz: Beim Eröffnen ist jede Karte erlaubt.
			karten = []int{0}
		} else {
			msg, _ := g.arschAktion(seat, nil, true, name)
			return msg == ""
		}
	}
	msg, _ := g.arschAktion(seat, karten, false, name)
	return msg == ""
}

// arschBotWahl sucht die Karten für diesen Zug. Eine leere Liste heißt: passen
// (nur erlaubt, wenn ein Stich liegt).
func arschBotWahl(g *ArschlochGame, seat int) []int {
	erlaubt := g.arschPlayable(seat)
	if len(erlaubt) == 0 {
		return nil
	}
	stufe := botStufe(g.BotStufe)

	// Gruppen nach Wert: gleiche Karten gehören zusammen.
	gruppen := map[int][]int{}
	for _, index := range erlaubt {
		rank := g.Players[seat].Hand[index].Rank
		gruppen[rank] = append(gruppen[rank], index)
	}
	raenge := make([]int, 0, len(gruppen))
	for rank := range gruppen {
		raenge = append(raenge, rank)
	}
	sort.Ints(raenge)

	// Beim Bedienen zählt genau die Anzahl des Stichs.
	need := len(g.Trick)
	if need > 0 {
		for _, rank := range raenge {
			if len(gruppen[rank]) > need {
				gruppen[rank] = gruppen[rank][:need]
			}
		}
	}

	if stufe == botLeicht {
		if need > 0 && rand.Intn(100) < 45 {
			return nil
		}
		rank := raenge[rand.Intn(len(raenge))]
		return gruppen[rank]
	}

	// mittel und schwer legen den niedrigsten erlaubten Wert.
	wahl := raenge[0]
	if need == 0 {
		switch stufe {
		case botSchwer:
			// Eröffnen: unter den drei niedrigsten Werten die größte Gruppe.
			for _, rank := range raenge {
				if rank > raenge[0]+2 {
					break
				}
				if len(gruppen[rank]) > len(gruppen[wahl]) {
					wahl = rank
				}
			}
		default:
			// Mittel: immer nur eine Karte eröffnen – sicher, aber langsam.
			return gruppen[wahl][:1]
		}
		return gruppen[wahl]
	}

	if stufe == botSchwer && wahl >= 14 && need > 1 && len(g.Players[seat].Hand) > 2*need {
		// Nur ein Ass mehrfach abgeben, obwohl noch andere Karten da sind: das
		// hebt man sich für den Schluss auf.
		return nil
	}
	return gruppen[wahl]
}
