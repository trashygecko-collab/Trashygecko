package main

/* ===========================================================================
   BOT FÜR LÜGEN MÄXCHEN

   Der Bot würfelt, sagt an und deckt auf – alles über dieselben drei
   Prüffunktionen wie der Mensch (mxWurf, mxSage, mxAufdecken). Er sieht nur
   seine eigenen Würfel; was der Vorgänger gewürfelt hat, weiß er nicht.

   Stufen:
     leicht  – deckt nur selten auf (12 %) und blufft gern mit großen
               Sprüngen nach oben; fällt damit selbst auf.
     mittel  – deckt nur Pasche und das Mäxchen auf, sagt sonst ehrlich an,
               wenn die eigenen Würfel reichen, und blufft minimal.
     schwer  – riecht Lügen früher (ab 6-2), sagt ehrlich an, wenn es geht,
               und blufft nur um eine Stufe nach oben.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// mxBotStufen sind die Stufen der Bots am Tisch – in der Reihenfolge, in der
// sie in g.Bots stehen. Ohne Eintrag gilt die Stufe der Partie.
func mxBotStufeFuer(g *MaexchenGame, seat int) string {
	if g == nil || seat < 0 || seat >= len(g.Players) {
		return botMittel
	}
	id := g.Players[seat].ID
	for i, b := range g.Bots {
		if b != id {
			continue
		}
		if i < len(g.BotStufen) {
			return botStufe(g.BotStufen[i])
		}
		return botStufe(g.BotStufe)
	}
	return botMittel
}

// mxBotIstBot sagt, ob dieser Platz ein Bot ist.
func mxBotIstBot(g *MaexchenGame, seat int) bool {
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

// mxBotDran sagt, ob ein Bot jetzt handeln muss.
func mxBotDran(g *MaexchenGame) bool {
	if g == nil || g.Status != "active" || len(g.Bots) == 0 {
		return false
	}
	return g.Current >= 0 && g.Current < len(g.Players) && mxBotIstBot(g, g.Current)
}

// mxBotTreiben lässt genau einen Bot nach der Denkpause handeln. Aufgerufen
// beim Abholen des Spielstands (GET), damit die Partie auch läuft, wenn der
// Mensch nur zusieht.
func mxBotTreiben(g *MaexchenGame, jetzt time.Time) bool {
	if !mxBotDran(g) {
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
	return mxBotSchritt(g)
}

// mxAlleCombos sind alle 21 Ansagen, aufsteigend nach Rang.
func mxAlleCombos() []MXCombo {
	liste := make([]MXCombo, 0, 21)
	for hoch := 1; hoch <= 6; hoch++ {
		for tief := 1; tief <= hoch; tief++ {
			liste = append(liste, MXCombo{Hoch: hoch, Tief: tief})
		}
	}
	for i := 0; i < len(liste); i++ {
		for j := i + 1; j < len(liste); j++ {
			if mxRang(liste[j]) < mxRang(liste[i]) {
				liste[i], liste[j] = liste[j], liste[i]
			}
		}
	}
	return liste
}

// mxNaechsteCombo liefert die kleinste Ansage über c.
func mxNaechsteCombo(c MXCombo) (MXCombo, bool) {
	rang := mxRang(c)
	for _, kandidat := range mxAlleCombos() {
		if mxRang(kandidat) > rang {
			return kandidat, true
		}
	}
	return MXCombo{}, false
}

// mxBotSchritt macht genau einen Schritt: aufdecken, würfeln oder ansagen.
func mxBotSchritt(g *MaexchenGame) bool {
	if g == nil || g.Status != "active" {
		return false
	}
	seat := g.Current
	if !mxBotIstBot(g, seat) {
		return false
	}
	name := g.Players[seat].Name

	if g.Phase == "wurf" {
		if mxBotDecktAuf(g, seat) {
			meldung, _ := mxAufdecken(g, seat, name)
			return meldung == ""
		}
		if meldung, _ := mxWurf(g, seat, name); meldung != "" {
			return false
		}
	}

	if g.Phase == "ansage" && g.DiceOwner == g.Players[seat].ID {
		wahl, ok := mxBotAnsage(g, seat)
		if !ok {
			return false
		}
		meldung, _ := mxSage(g, seat, name, wahl)
		return meldung == ""
	}
	return false
}

// mxBotDecktAuf entscheidet, ob der Bot lieber aufdeckt statt zu würfeln. Je
// höher die Ansage, desto wahrscheinlicher ist sie gelogen.
func mxBotDecktAuf(g *MaexchenGame, seat int) bool {
	if g == nil || g.Ansage == nil || g.AnsageVon == "" {
		return false
	}
	if g.AnsageVon == g.Players[seat].ID {
		return false // die eigene Ansage kann man nicht aufdecken
	}
	rang := mxRang(*g.Ansage)
	if rang >= mxRang(MXCombo{Hoch: 2, Tief: 1}) {
		return true // 2-1 lässt sich nicht überbieten
	}
	switch mxBotStufeFuer(g, seat) {
	case botLeicht:
		return rand.Intn(100) < 40
	case botSchwer:
		// Über 6-5 gibt es nur noch Pasche – wer da noch erhöhen will, muss
		// einen Pasch behaupten, und die fliegen fast immer auf. Also lieber
		// gleich aufdecken.
		if rang >= mxRang(MXCombo{Hoch: 6, Tief: 5}) {
			return true
		}
	}
	return rang >= mxRang(MXCombo{Hoch: 5, Tief: 5}) // Pasch oder Mäxchen
}

// mxBotAnsage wählt die Ansage: ehrlich, wenn die eigenen Würfel über der
// bisherigen Ansage liegen, sonst ein Bluff.
func mxBotAnsage(g *MaexchenGame, seat int) (MXCombo, bool) {
	if g == nil || len(g.wurf) != 2 {
		return MXCombo{}, false
	}
	eigen := mxCombo(g.wurf[0], g.wurf[1])
	if g.Ansage == nil {
		return eigen, true
	}
	vor := mxRang(*g.Ansage)
	if mxRang(eigen) > vor {
		return eigen, true
	}
	naechste, ok := mxNaechsteCombo(*g.Ansage)
	if !ok {
		return MXCombo{}, false
	}
	stufe := mxBotStufeFuer(g, seat)
	if stufe == botSchwer {
		// Beim Bluffen keinen Pasch behaupten, wenn es einen normalen Wert
		// darüber gibt: Pasche werden eher aufgedeckt.
		for _, c := range mxAlleCombos() {
			if mxRang(c) <= vor {
				continue
			}
			if c.Hoch == c.Tief {
				continue
			}
			if mxRang(c)-vor <= 12 {
				return c, true
			}
			break
		}
		return naechste, true
	}
	if stufe == botLeicht && rand.Intn(100) < 70 {
		// Großer Sprung: leichter zu durchschauen, aber einschüchternd.
		hoeher := []MXCombo{}
		for _, c := range mxAlleCombos() {
			if mxRang(c) > vor {
				hoeher = append(hoeher, c)
			}
		}
		if len(hoeher) > 0 {
			return hoeher[rand.Intn(len(hoeher))], true
		}
	}
	return naechste, true
}
