package main

import (
	"math/rand"
	"time"
)

/* Memory-Bot.

   Der Bot spielt mit demselben Brett wie der Mensch, sieht aber nur, was
   aufgedeckt wurde. Wie viel er davon behält, hängt von der Stufe ab:

     leicht – merkt sich drei Karten und greift oft daneben,
     mittel – merkt sich jede Karte des Bretts, irrt gelegentlich,
     schwer – merkt sich alles und spielt ohne Fehler.

   Er würfelt aus einer stellungsgebundenen Quelle, damit Messungen mit
   gleichem Brett vergleichbar bleiben. */

func memoryBotGedaechtnis(stufe string) int {
	switch stufe {
	case botLeicht:
		return 3
	case botSchwer:
		return 99
	}
	return 8
}

func memoryBotFehler(stufe string) int {
	switch stufe {
	case botLeicht:
		return 35
	case botSchwer:
		return 0
	}
	return 8
}

// memoryBotZufall bindet den Würfel an Stellung und Gedächtnis des Bots.
func memoryBotZufall(g *MemoryGame, salz int) *rand.Rand {
	saat := int64(1)
	for _, k := range g.Karten {
		saat = saat*31 + int64(k) + 7
	}
	for _, b := range g.Besitzer {
		saat = saat*31 + int64(b+2)
	}
	for _, o := range g.Offen {
		saat = saat*17 + int64(o+1)
	}
	for feld := range g.BotWeiss {
		saat = saat*13 + int64(feld+1)
	}
	if g.BotStufe == botLeicht {
		saat = saat*7 + 3
	}
	return rand.New(rand.NewSource(saat*101 + int64(salz)))
}

// memoryBotMerken nimmt eine aufgedeckte Karte ins Gedächtnis (mit Grenze).
func memoryBotMerken(g *MemoryGame, feld int) {
	if g == nil || len(g.Bots) == 0 {
		return
	}
	if g.BotWeiss == nil {
		g.BotWeiss = map[int]int{}
		g.BotReihe = nil
	}
	if _, schon := g.BotWeiss[feld]; schon {
		return
	}
	g.BotWeiss[feld] = g.Karten[feld]
	g.BotReihe = append(g.BotReihe, feld)
	for len(g.BotReihe) > memoryBotGedaechtnis(g.BotStufe) {
		alt := g.BotReihe[0]
		g.BotReihe = g.BotReihe[1:]
		delete(g.BotWeiss, alt)
	}
}

// memoryBotFrei liefert alle Felder, die noch keinem gehören.
func memoryBotFrei(g *MemoryGame) []int {
	frei := []int{}
	for i, b := range g.Besitzer {
		if b < 0 {
			frei = append(frei, i)
		}
	}
	return frei
}

func memoryBotOffen(g *MemoryGame, feld int) bool {
	for _, o := range g.Offen {
		if o == feld {
			return true
		}
	}
	return false
}

func memoryBotWaehle(rnd *rand.Rand, felder []int) int {
	return felder[rnd.Intn(len(felder))]
}

// memoryBotFlip entscheidet die nächste Karte, die der Bot aufdeckt.
func memoryBotFlip(g *MemoryGame) (int, bool) {
	frei := []int{}
	for _, feld := range memoryBotFrei(g) {
		if !memoryBotOffen(g, feld) {
			frei = append(frei, feld)
		}
	}
	if len(frei) == 0 {
		return 0, false
	}
	rnd := memoryBotZufall(g, len(g.BotWeiss)*3+len(g.Offen))
	fehler := memoryBotFehler(g.BotStufe)

	// Liegt schon eine Karte offen, passt sie zu einer bekannten?
	if len(g.Offen) == 1 {
		wert := g.Karten[g.Offen[0]]
		partner := []int{}
		for _, feld := range frei {
			if g.BotWeiss[feld] == wert {
				partner = append(partner, feld)
			}
		}
		if len(partner) > 0 && rnd.Intn(100) >= fehler {
			return memoryBotWaehle(rnd, partner), true
		}
	}

	// Ein bekanntes Paar aufdecken.
	paar := memoryBotBekanntesPaar(g, frei)
	if paar >= 0 && rnd.Intn(100) >= fehler {
		return paar, true
	}

	// Sonst eine Karte, die der Bot noch nicht kennt (Erkunden).
	unbekannt := []int{}
	for _, feld := range frei {
		if _, kennt := g.BotWeiss[feld]; !kennt {
			unbekannt = append(unbekannt, feld)
		}
	}
	if len(unbekannt) > 0 {
		return memoryBotWaehle(rnd, unbekannt), true
	}
	return memoryBotWaehle(rnd, frei), true
}

// memoryBotBekanntesPaar sucht zwei freie Felder mit demselben bekannten Wert
// und liefert das erste davon (-1, wenn es keins gibt).
func memoryBotBekanntesPaar(g *MemoryGame, frei []int) int {
	gesehen := map[int]int{}
	for _, feld := range frei {
		wert, kennt := g.BotWeiss[feld]
		if !kennt {
			continue
		}
		if erstes, ok := gesehen[wert]; ok {
			_ = erstes
			return feld
		}
		gesehen[wert] = feld
	}
	return -1
}

// memoryBotTreiben lässt den Bot nach seiner Denkpause eine Karte aufdecken.
// Die Sperre hält der Aufrufer.
func memoryBotTreiben(room *Room, now time.Time) {
	g := room.Memory
	if g == nil || len(g.Bots) == 0 || g.Status != "active" {
		return
	}
	if g.Turn != "blue" || !now.After(g.BotFrist) {
		return
	}
	feld, ok := memoryBotFlip(g)
	if !ok {
		return
	}
	if _, meldung := memoryAktion(room, g.Bots[0], "flip", feld); meldung != "" {
		g.BotFrist = now.Add(time.Second)
		return
	}
	g.BotFrist = now.Add(botWartezeit(g.BotStufe))
}

// memoryBotSchritt deckt eine Karte auf, ohne auf die Denkpause zu warten –
// für Tests und für den GET-Antrieb im schnellen Takt.
func memoryBotSchritt(room *Room) bool {
	g := room.Memory
	if g == nil || len(g.Bots) == 0 || g.Status != "active" || g.Turn != "blue" {
		return false
	}
	feld, ok := memoryBotFlip(g)
	if !ok {
		return false
	}
	_, meldung := memoryAktion(room, g.Bots[0], "flip", feld)
	return meldung == ""
}
