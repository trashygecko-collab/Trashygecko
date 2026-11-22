package main

/* ===========================================================================
   BOT FÜR UNO

   Der Bot sitzt als normaler Mitspieler am Tisch und zieht ausschließlich über
   unoAktion – also durch dieselbe Prüfung wie der Mensch (Farbe oder Wert,
   Reihenfolge, Aussetzen, +2, +4, Farbe wählen, UNO-Frist). Er sieht nur sein
   eigenes Blatt und den offenen Ablagestapel, nie fremde Karten.

   Stufen:
     leicht  – legt irgendeine passende Karte, wählt eine zufällige Farbe und
               nimmt eine +4 immer an.
     mittel  – legt lieber Zahlen als wilde Karten und wählt die Farbe, die es
               am häufigsten hält.
     schwer  – legt Aussetzen, +2 und +4 bevorzugt, wenn der Nächste nur noch
               eine Karte hat, hält Mehrfachfarben zusammen und zweifelt eine
               +4 gelegentlich an.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// unoBotSitz ist der Platz des Bots (-1, wenn keiner mitspielt).
func unoBotSitz(g *UnoGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return unoSeat(g, g.BotID)
}

// unoBotAmZug sagt, ob der Bot jetzt etwas tun muss: UNO rufen, eine +4
// beantworten, Farbe wählen oder legen/ziehen.
func unoBotAmZug(g *UnoGame, sitz int) bool {
	if g == nil || g.Status != "active" || sitz < 0 {
		return false
	}
	if g.PendingUno == sitz {
		return true
	}
	if g.Phase == "challenge" && g.Challenge != nil && g.Challenge.Challenger == sitz {
		return true
	}
	if g.Current != sitz {
		return false
	}
	return g.Phase == "playing" || g.Phase == "chooseColor"
}

// unoBotTreiben lässt den Bot nach der Denkpause genau einen Schritt machen.
// Aufgerufen wird es beim Abholen des Spielstands (GET) – so zieht der Bot
// auch, wenn der Mensch nur zusieht.
func unoBotTreiben(g *UnoGame, jetzt time.Time) bool {
	sitz := unoBotSitz(g)
	if !unoBotAmZug(g, sitz) {
		return false
	}
	if g.BotFrist.IsZero() {
		g.BotFrist = jetzt.Add(botWartezeit(g.BotStufe))
		return false
	}
	if jetzt.Before(g.BotFrist) {
		return false
	}
	g.BotFrist = time.Time{}
	return unoBotSchritt(g, sitz)
}

// unoBotSchritt macht genau einen Schritt. Die Reihenfolge folgt den Fristen:
// erst UNO rufen (die Frist läuft), dann eine +4 beantworten, dann Farbe
// wählen, dann legen oder ziehen.
func unoBotSchritt(g *UnoGame, sitz int) bool {
	if g.PendingUno == sitz {
		return unoBotTu(g, sitz, "uno", 0, "")
	}
	if g.Phase == "challenge" && g.Challenge != nil && g.Challenge.Challenger == sitz {
		akt := "accept4"
		if botStufe(g.BotStufe) == botSchwer && rand.Intn(100) < 30 {
			akt = "challenge4"
		}
		return unoBotTu(g, sitz, akt, 0, "")
	}
	if g.Phase == "chooseColor" {
		if g.Current != sitz {
			return false
		}
		return unoBotTu(g, sitz, "color", 0, unoBotFarbe(g, sitz))
	}
	if g.Phase != "playing" || g.Current != sitz {
		return false
	}
	hand := g.Players[sitz].Hand
	if g.Drawn != nil {
		// Nach dem Ziehen darf nur die gezogene Karte gelegt werden.
		letzte := len(hand) - 1
		if letzte >= 0 && unoPlayable(g, hand[letzte]) {
			return unoBotTu(g, sitz, "play", letzte, "")
		}
		return unoBotTu(g, sitz, "pass", 0, "")
	}
	if i := unoBotKarte(g, sitz); i >= 0 {
		return unoBotTu(g, sitz, "play", i, "")
	}
	return unoBotTu(g, sitz, "draw", 0, "")
}

// unoBotTu schickt eine Aktion durch dieselbe Prüfung wie der Mensch.
func unoBotTu(g *UnoGame, sitz int, akt string, index int, farbe string) bool {
	status, _ := g.unoAktion(sitz, akt, index, farbe, g.Players[sitz].Name)
	return status == 0
}

// unoBotKarte wählt die passende Karte. -1 heißt: keine passt, also ziehen.
func unoBotKarte(g *UnoGame, sitz int) int {
	hand := g.Players[sitz].Hand
	kandidaten := []int{}
	for i, k := range hand {
		if unoPlayable(g, k) {
			kandidaten = append(kandidaten, i)
		}
	}
	if len(kandidaten) == 0 {
		return -1
	}
	if botStufe(g.BotStufe) == botLeicht || len(kandidaten) == 1 {
		return kandidaten[rand.Intn(len(kandidaten))]
	}

	anzahl := map[string]int{}
	for _, k := range hand {
		if k.Color != "wild" {
			anzahl[k.Color]++
		}
	}
	nachbar := unoNext(g, sitz)
	knapp := len(g.Players[nachbar].Hand) <= 1

	best, bestWert := kandidaten[0], -1<<30
	for _, i := range kandidaten {
		k := hand[i]
		wert := 6 // Zahl: legt Farbe weiter, ohne Sondervorteil
		switch k.Value {
		case "wild4":
			wert = 3
		case "wild":
			wert = 2
		case "draw2":
			wert = 9
		case "skip":
			wert = 8
		case "reverse":
			wert = 7
		}
		if knapp && (k.Value == "draw2" || k.Value == "skip" || k.Value == "wild4") {
			wert += 4
		}
		if k.Color != "wild" {
			// Mehrfach gehaltene Farben zusammenhalten, letzte Karte einer
			// Farbe dagegen gern loswerden.
			if anzahl[k.Color] > 1 {
				wert += anzahl[k.Color]
			} else {
				wert -= 2
			}
		}
		if wert > bestWert {
			best, bestWert = i, wert
		}
	}
	return best
}

// unoBotFarbe wählt die Farbe, die der Bot am häufigsten auf der Hand hat;
// bei Gleichstand oder nur wilden Karten entscheidet der Zufall.
func unoBotFarbe(g *UnoGame, sitz int) string {
	anzahl := map[string]int{}
	for _, k := range g.Players[sitz].Hand {
		if k.Color != "wild" {
			anzahl[k.Color]++
		}
	}
	best, bestWert := "", 0
	for _, c := range unoColors {
		if anzahl[c] > bestWert {
			best, bestWert = c, anzahl[c]
		}
	}
	if best == "" {
		return unoColors[rand.Intn(len(unoColors))]
	}
	return best
}
