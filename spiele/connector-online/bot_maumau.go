package main

/* ===========================================================================
   BOT FÜR MAU MAU

   Der Bot sitzt als normaler Mitspieler am Tisch und zieht ausschließlich über
   maumauAktion – also durch dieselbe Prüfung wie der Mensch (Farbe oder Wert,
   7/8/Bube, MAU-Frist, Strafkarten). Er sieht nur sein eigenes Blatt und die
   offene oberste Karte, nie fremde Karten.

   Stufen:
     leicht  – legt irgendeine passende Karte, wünscht eine zufällige Farbe und
               erwischt ein vergessenes MAU nur selten.
     mittel  – legt 7, 8 und Bube lieber als Zahlen, hält eine mehrfach
               gehaltene Farbe zusammen und wünscht die häufigste Farbe.
     schwer  – spielt 7 und 8 besonders gern, wenn der Nächste knapp dran ist,
               wartet mit dem Buben auf eine starke Wunschfarbe und erwischt
               ein vergessenes MAU fast immer.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// maumauBotSitz ist der Platz des Bots (-1, wenn keiner mitspielt).
func maumauBotSitz(g *MauMauGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return maumauSeat(g, g.BotID)
}

// maumauBotAmZug sagt, ob der Bot selbst etwas tun muss: MAU rufen, eine
// Wunschfarbe wählen oder legen/ziehen/behalten.
func maumauBotAmZug(g *MauMauGame, sitz int) bool {
	if g == nil || g.Status != "active" || sitz < 0 {
		return false
	}
	if g.PendingMau == sitz {
		return true
	}
	if g.Current != sitz {
		return false
	}
	return g.Phase == "playing" || g.Phase == "chooseSuit"
}

// maumauBotErwischenFaellig sagt, ob der Bot jetzt über ein vergessenes MAU
// des Menschen entscheiden sollte. Entschieden wird pro Frist genau einmal
// (BotCatchUm), damit nicht jeder Abruf neu würfelt; die Frist bleibt lang
// genug, dass der Mensch sein MAU noch rechtzeitig rufen kann.
func maumauBotErwischenFaellig(g *MauMauGame, sitz int, jetzt time.Time) bool {
	if g.PendingMau < 0 || g.PendingMau == sitz || g.MauUntil.IsZero() {
		return false
	}
	if jetzt.Before(g.MauUntil.Add(-6 * time.Second)) {
		return false
	}
	return !g.BotCatchUm.Equal(g.MauUntil)
}

// maumauBotTreiben lässt den Bot nach der Denkpause genau einen Schritt machen.
// Aufgerufen wird es beim Abholen des Spielstands (GET) – so zieht der Bot
// auch, wenn der Mensch nur zusieht.
func maumauBotTreiben(g *MauMauGame, jetzt time.Time) bool {
	sitz := maumauBotSitz(g)
	if g == nil || g.Status != "active" || sitz < 0 {
		return false
	}
	if !maumauBotAmZug(g, sitz) && !maumauBotErwischenFaellig(g, sitz, jetzt) {
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
	return maumauBotSchritt(g, sitz, jetzt)
}

// maumauBotSchritt macht genau einen Schritt. Reihenfolge: eigenes MAU,
// vergessenes MAU des Menschen, Wunschfarbe, dann legen oder ziehen.
func maumauBotSchritt(g *MauMauGame, sitz int, jetzt time.Time) bool {
	if g.PendingMau == sitz {
		return maumauBotTu(g, sitz, "mau", 0, "")
	}
	if maumauBotErwischen(g, sitz, jetzt) {
		return true
	}
	if g.Phase == "chooseSuit" {
		if g.Current != sitz {
			return false
		}
		return maumauBotTu(g, sitz, "suit", 0, maumauBotFarbe(g, sitz))
	}
	if g.Phase != "playing" || g.Current != sitz {
		return false
	}
	hand := g.Players[sitz].Hand
	if g.Drawn != nil {
		// Nach dem Ziehen darf nur die gezogene Karte gelegt werden.
		letzte := len(hand) - 1
		if letzte >= 0 && maumauPlayable(g, hand[letzte]) {
			return maumauBotTu(g, sitz, "play", letzte, "")
		}
		return maumauBotTu(g, sitz, "keep", 0, "")
	}
	if i := maumauBotKarte(g, sitz); i >= 0 {
		return maumauBotTu(g, sitz, "play", i, "")
	}
	return maumauBotTu(g, sitz, "draw", 0, "")
}

// maumauBotErwischen entscheidet einmal je MAU-Frist, ob der Bot einschreitet.
// Ein JA kostet den Menschen zwei Strafkarten; ein NEIN wird gemerkt.
func maumauBotErwischen(g *MauMauGame, sitz int, jetzt time.Time) bool {
	if !maumauBotErwischenFaellig(g, sitz, jetzt) {
		return false
	}
	g.BotCatchUm = g.MauUntil
	chance := 70
	switch botStufe(g.BotStufe) {
	case botLeicht:
		chance = 35
	case botSchwer:
		chance = 100
	}
	if rand.Intn(100) >= chance {
		return false
	}
	return maumauBotTu(g, sitz, "catch", 0, "")
}

// maumauBotTu schickt eine Aktion durch dieselbe Prüfung wie der Mensch.
func maumauBotTu(g *MauMauGame, sitz int, akt string, index int, suit string) bool {
	status, _ := g.maumauAktion(sitz, akt, index, suit, g.Players[sitz].Name)
	return status == 0
}

// maumauBotKarte wählt die passende Karte. -1 heißt: keine passt, also ziehen.
func maumauBotKarte(g *MauMauGame, sitz int) int {
	hand := g.Players[sitz].Hand
	kandidaten := []int{}
	for i, k := range hand {
		if maumauPlayable(g, k) {
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
		anzahl[k.Suit]++
	}
	nachbar := maumauNext(g, sitz)
	knapp := len(g.Players[nachbar].Hand) <= 2
	staerkste := 0
	for _, n := range anzahl {
		if n > staerkste {
			staerkste = n
		}
	}

	best, bestWert := kandidaten[0], -1<<30
	for _, i := range kandidaten {
		k := hand[i]
		wert := 5 // Zahl: legt die Farbe weiter, ohne Sondervorteil
		switch k.Value {
		case "7":
			wert = 10
		case "8":
			wert = 8
		case "B":
			wert = 7
		case "A":
			wert = 6
		}
		if knapp {
			// Wer nur noch zwei Karten hat, soll aussetzen oder ziehen.
			switch k.Value {
			case "7":
				wert += 6
			case "8":
				wert += 5
			}
		}
		// Mehrfach gehaltene Farben zusammenhalten, letzte Karte einer Farbe
		// dagegen gern loswerden.
		if anzahl[k.Suit] > 1 {
			wert += anzahl[k.Suit]
		} else {
			wert--
		}
		if botStufe(g.BotStufe) == botSchwer && k.Value == "B" && staerkste < 3 {
			// Ohne starke Wunschfarbe ist der Bube nur eine Notlösung.
			wert -= 2
		}
		if wert > bestWert {
			best, bestWert = i, wert
		}
	}
	return best
}

// maumauBotFarbe wünscht die Farbe, die der Bot am häufigsten auf der Hand
// hat; bei Gleichstand oder leerer Hand entscheidet der Zufall.
func maumauBotFarbe(g *MauMauGame, sitz int) string {
	anzahl := map[string]int{}
	for _, k := range g.Players[sitz].Hand {
		anzahl[k.Suit]++
	}
	best, bestWert := "", 0
	for _, farbe := range maumauSuits {
		if anzahl[farbe] > bestWert {
			best, bestWert = farbe, anzahl[farbe]
		}
	}
	if best == "" {
		return maumauSuits[rand.Intn(len(maumauSuits))]
	}
	return best
}
