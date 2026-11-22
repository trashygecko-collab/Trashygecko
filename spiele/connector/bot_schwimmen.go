package main

/* ===========================================================================
   BOT FÜR 31 (SCHWIMMEN)

   Der Bot sitzt als normaler Mitspieler am Tisch und zieht ausschließlich über
   schwimmAktion – dieselbe Prüfung wie der Mensch (Zug, Klopfrunde, Tausch,
   Schieben, Feuer, Rundenende). Er sieht sein eigenes Blatt und die drei
   offenen Tischkarten, nie fremde Karten.

   Ziel im Spiel: möglichst viele Punkte in einer Farbe (31 ist die Bestmarke),
   denn wer am Ende die wenigsten Punkte hat, verliert ein Leben.

   Stufen:
     leicht  – tauscht nur, wenn es sofort besser wird, und klopft spät.
     mittel  – sucht den besten Einzeltausch oder tauscht alle drei, klopft
               bei 24 Punkten.
     schwer  – wie mittel, klopft aber schon bei 26 Punkten und meidet
               Tauschzüge, die die beste Farbe zerschneiden.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// schwimmBotSitz ist der Platz des Bots (-1, wenn keiner mitspielt).
func schwimmBotSitz(g *SchwimmenGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return schwimmSeat(g, g.BotID)
}

// schwimmBotAmZug sagt, ob der Bot jetzt etwas tun muss. Zwischen den Runden
// darf jeder die nächste anstoßen – das übernimmt der Bot nach kurzer Pause.
func schwimmBotAmZug(g *SchwimmenGame, sitz int) bool {
	if g == nil || g.Status != "active" || sitz < 0 {
		return false
	}
	if g.Phase == "between" {
		return true
	}
	return g.Phase == "playing" && g.Current == sitz
}

// schwimmBotTreiben lässt den Bot nach der Denkpause genau einen Schritt
// machen. Aufgerufen wird es beim Abholen des Spielstands (GET).
func schwimmBotTreiben(g *SchwimmenGame, jetzt time.Time) bool {
	sitz := schwimmBotSitz(g)
	if !schwimmBotAmZug(g, sitz) {
		return false
	}
	if g.BotFrist.IsZero() {
		warte := botWartezeit(g.BotStufe)
		if g.Phase == "between" {
			warte *= 2
		}
		g.BotFrist = jetzt.Add(warte)
		return false
	}
	if jetzt.Before(g.BotFrist) {
		return false
	}
	g.BotFrist = time.Time{}
	return schwimmBotSchritt(g, sitz)
}

// schwimmBotSchritt macht genau einen Schritt: nächste Runde, klopfen, tauschen
// oder schieben.
func schwimmBotSchritt(g *SchwimmenGame, sitz int) bool {
	if g.Phase == "between" {
		return schwimmBotTu(g, sitz, "next", 0, 0)
	}
	if g.Phase != "playing" || g.Current != sitz {
		return false
	}
	hand := g.Players[sitz].Hand
	if len(hand) == 0 || len(g.Table) == 0 {
		return schwimmBotTu(g, sitz, "pass", 0, 0)
	}
	// Drei Asse: Ein Pass löst das Feuer aus.
	if schwimmIsFeuer(hand) {
		return schwimmBotTu(g, sitz, "pass", 0, 0)
	}

	basis := schwimmPoints(hand)
	schwelle := 26.0
	switch botStufe(g.BotStufe) {
	case botLeicht:
		schwelle = 21
	case botMittel:
		schwelle = 24
	}
	if g.FinalTurns == -1 && basis >= schwelle {
		if botStufe(g.BotStufe) != botLeicht || rand.Intn(100) < 60 {
			return schwimmBotTu(g, sitz, "knock", 0, 0)
		}
	}

	bestH, bestT, bestP := -1, -1, basis
	for h := range hand {
		for t := range g.Table {
			if p := schwimmMitKarte(hand, h, g.Table[t]); p > bestP {
				bestH, bestT, bestP = h, t, p
			}
		}
	}
	alle := 0.0
	if len(g.Table) == schwimmHandSize {
		alle = schwimmPoints(g.Table)
	}
	if alle > bestP {
		bestH, bestT, bestP = -1, -1, alle
	}

	if bestP > basis {
		if bestH < 0 {
			return schwimmBotTu(g, sitz, "swapall", 0, 0)
		}
		return schwimmBotTu(g, sitz, "swap", bestH, bestT)
	}
	return schwimmBotTu(g, sitz, "pass", 0, 0)
}

// schwimmBotTu schickt eine Aktion durch dieselbe Prüfung wie der Mensch.
func schwimmBotTu(g *SchwimmenGame, sitz int, akt string, handIndex, tableIndex int) bool {
	status, _ := g.schwimmAktion(sitz, akt, handIndex, tableIndex)
	return status == 0
}

// schwimmMitKarte rechnet die Punktzahl, wenn Handkarte h gegen karte getauscht
// wird – ohne den Spielstand anzufassen.
func schwimmMitKarte(hand []SchwimmCard, h int, karte SchwimmCard) float64 {
	probe := append([]SchwimmCard(nil), hand...)
	probe[h] = karte
	return schwimmPoints(probe)
}
