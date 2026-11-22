package main

/* ===========================================================================
   BOT FÜR SIM

   Sim ist ein Duell: 15 Linien zwischen sechs Punkten, wer ein Dreieck in der
   eigenen Farbe vollendet, hat verloren. Der Bot setzt seine Linien durch
   simLinie – dieselbe Prüfung wie der Mensch (Reihenfolge, freie Linie).
   Er sieht nur das Brett, das auch der Mensch sieht.

   Stufen:
     leicht  – setzt irgendeine freie Linie, auch wenn sie das eigene Dreieck
               schließt (verliert dadurch oft sofort).
     mittel  – meidet das eigene Dreieck, wählt unter den sicheren Linien aber
               zufällig.
     schwer  – meidet das eigene Dreieck und wählt die Linie, die dem Gegner
               die wenigsten sicheren Antworten lässt (erzwungenes Dreieck).
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// simBotIstBot sagt, ob diese Kennung ein Bot ist.
func simBotIstBot(g *SimGame, id string) bool {
	if g == nil {
		return false
	}
	for _, b := range g.Bots {
		if b == id {
			return true
		}
	}
	return false
}

// simBotDran sagt, ob der Bot am Zug ist.
func simBotDran(g *SimGame) bool {
	if g == nil || g.Status != "active" || len(g.Bots) == 0 {
		return false
	}
	return simBotIstBot(g, simIDVonFarbe(g, g.Current))
}

// simIDVonFarbe liefert die Kennung hinter einer Farbe.
func simIDVonFarbe(g *SimGame, farbe string) string {
	if farbe == "red" {
		return g.RedID
	}
	return g.BlueID
}

// simBotTreiben lässt den Bot nach der Denkpause ziehen; aufgerufen im
// GET-Abruf, damit die Partie auch läuft, wenn der Mensch nur zusieht.
func simBotTreiben(g *SimGame, jetzt time.Time) bool {
	if !simBotDran(g) {
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
	return simBotSchritt(g)
}

// simBotSchritt zieht genau eine Linie.
func simBotSchritt(g *SimGame) bool {
	if !simBotDran(g) {
		return false
	}
	index := simBotWahl(g)
	if index < 0 {
		return false
	}
	msg, _ := simLinie(g, simIDVonFarbe(g, g.Current), index)
	return msg == ""
}

// simMitLinie liefert das Brett mit einer zusätzlichen Linie.
func simMitLinie(edges []string, index int, farbe string) []string {
	e := append([]string(nil), edges...)
	e[index] = farbe
	return e
}

// simSicher sagt, ob diese Linie kein eigenes Dreieck schließt.
func simSicher(edges []string, index int, farbe string) bool {
	return simTriangle(simMitLinie(edges, index, farbe), farbe) == nil
}

// simBotWahl sucht die Linie für den aktuellen Zug.
func simBotWahl(g *SimGame) int {
	farbe := g.Current
	gegner := "red"
	if farbe == "red" {
		gegner = "blue"
	}
	frei := []int{}
	sicher := []int{}
	for i, v := range g.Edges {
		if v != "" {
			continue
		}
		frei = append(frei, i)
		if simSicher(g.Edges, i, farbe) {
			sicher = append(sicher, i)
		}
	}
	if len(frei) == 0 {
		return -1
	}

	switch botStufe(g.BotStufe) {
	case botLeicht:
		return frei[rand.Intn(len(frei))]
	case botSchwer:
		if len(sicher) == 0 {
			return frei[rand.Intn(len(frei))]
		}
		bester, besteZahl := -1, 1<<30
		for _, i := range sicher {
			nachher := simMitLinie(g.Edges, i, farbe)
			antworten := 0
			for j, v := range nachher {
				if v == "" && simSicher(nachher, j, gegner) {
					antworten++
				}
			}
			if antworten < besteZahl {
				bester, besteZahl = i, antworten
			}
		}
		if bester >= 0 {
			return bester
		}
	}

	if len(sicher) == 0 {
		// Jede Linie schließt ein eigenes Dreieck – dann geht es nicht anders.
		return frei[rand.Intn(len(frei))]
	}
	return sicher[rand.Intn(len(sicher))]
}
