package main

/* ===========================================================================
   BOT FÜR SKYJO

   Der Bot ist der blaue Platz und zieht ausschließlich über skyjoAktion –
   dieselbe Prüfung wie der Mensch (Phase, Zug, offene Karten, Spaltenregel,
   Rundenende). Er sieht nur seine eigenen offenen Karten; verdeckte Karten
   kennt er nicht, sie werden nur als „irgendwo verdeckt“ gewählt.

   Stufen:
     leicht  – nimmt nur eine ganz niedrige Ablagekarte und wirft mittlere
               Stapelkarten weg.
     mittel  – tauscht, solange die gezogene Karte unter 5 liegt, und nimmt
               eine Ablagekarte, wenn sie die höchste offene Karte schlägt.
     schwer  – vervollständigt Spalten (drei gleiche zählen 0), nimmt auch
               mittlere Ablagekarten und hält die Hand dadurch niedrig.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// skyjoBotIndex ist der Platz des Bots (0 rot, 1 blau, -1 = keiner).
func skyjoBotIndex(g *SkyjoGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	farbe := skyjoColorOf(g, g.BotID)
	if farbe == "" {
		return -1
	}
	return skyjoIndex(g, farbe)
}

// skyjoBotAmZug sagt, ob der Bot jetzt etwas tun muss. Zwischen den Runden
// darf jeder die nächste anstoßen – das übernimmt der Bot nach kurzer Pause.
func skyjoBotAmZug(g *SkyjoGame, me int) bool {
	if g == nil || g.Status != "active" || me < 0 {
		return false
	}
	switch g.Phase {
	case "setup":
		return g.SetupPlayer == me
	case "play", "reveal":
		return g.Current == me
	case "between":
		return true
	}
	return false
}

// skyjoBotTreiben lässt den Bot nach der Denkpause genau einen Schritt machen.
// Aufgerufen wird es beim Abholen des Spielstands (GET).
func skyjoBotTreiben(g *SkyjoGame, jetzt time.Time) bool {
	me := skyjoBotIndex(g)
	if !skyjoBotAmZug(g, me) {
		return false
	}
	if g.BotFrist.IsZero() {
		warte := botWartezeit(g.BotStufe)
		if g.Phase == "between" {
			// Erst soll der Mensch die Abrechnung sehen.
			warte *= 2
		}
		g.BotFrist = jetzt.Add(warte)
		return false
	}
	if jetzt.Before(g.BotFrist) {
		return false
	}
	g.BotFrist = time.Time{}
	return skyjoBotSchritt(g, me)
}

// skyjoBotSchritt macht genau einen Schritt: aufdecken, ziehen, tauschen oder
// die nächste Runde anstoßen.
func skyjoBotSchritt(g *SkyjoGame, me int) bool {
	switch g.Phase {
	case "setup":
		return skyjoBotTu(g, me, "reveal", skyjoBotVerdeckt(g, me), "")
	case "between":
		return skyjoBotTu(g, me, "nextround", 0, "")
	case "reveal":
		return skyjoBotTu(g, me, "opencard", skyjoBotAufdecken(g, me), "")
	case "play":
		if g.Drawn == nil {
			return skyjoBotTu(g, me, "draw", 0, skyjoBotQuelle(g, me))
		}
		return skyjoBotNachziehen(g, me)
	}
	return false
}

// skyjoBotTu schickt eine Aktion durch dieselbe Prüfung wie der Mensch.
func skyjoBotTu(g *SkyjoGame, me int, akt string, index int, source string) bool {
	status, _ := g.skyjoAktion(me, akt, index, source)
	return status == 0
}

// skyjoBotQuelle entscheidet, ob der Bot die Ablagekarte nimmt oder zieht.
func skyjoBotQuelle(g *SkyjoGame, me int) string {
	if len(g.Discard) == 0 {
		return "deck"
	}
	top := g.Discard[len(g.Discard)-1]
	hoechste, hatOffen := 0, false
	for i := 0; i < skyjoCards; i++ {
		if !g.Open[me][i] || g.Removed[me][i] {
			continue
		}
		if !hatOffen || g.Cards[me][i] > hoechste {
			hoechste, hatOffen = g.Cards[me][i], true
		}
	}
	if !hatOffen {
		return "deck"
	}
	// Eine Ablagekarte, die eine Spalte vollendet, ist immer richtig.
	if skyjoBotSpalte(g, me, top) >= 0 {
		return "discard"
	}
	switch botStufe(g.BotStufe) {
	case botLeicht:
		if top <= 0 && hoechste > top {
			return "discard"
		}
	case botSchwer:
		if hoechste > top || top <= 1 {
			return "discard"
		}
	default:
		if hoechste > top+1 {
			return "discard"
		}
	}
	return "deck"
}

// skyjoBotNachziehen entscheidet nach dem Ziehen: tauschen oder abwerfen.
func skyjoBotNachziehen(g *SkyjoGame, me int) bool {
	v := *g.Drawn
	ziel, spalte := skyjoBotTauschziel(g, me, v)
	if g.DrawnSource == "discard" {
		// Von der Ablage genommen: Es muss getauscht werden.
		if ziel < 0 {
			return false
		}
		return skyjoBotTu(g, me, "swap", ziel, "")
	}
	schwelle := 5
	if botStufe(g.BotStufe) == botLeicht {
		schwelle = 7
	}
	if ziel >= 0 && spalte {
		return skyjoBotTu(g, me, "swap", ziel, "")
	}
	if ziel >= 0 && g.Open[me][ziel] && v < schwelle {
		return skyjoBotTu(g, me, "swap", ziel, "")
	}
	if ziel >= 0 && !g.Open[me][ziel] && v < 3 {
		// Sehr niedrige Karte offen hinlegen ist besser als sie wegzuwerfen.
		return skyjoBotTu(g, me, "swap", ziel, "")
	}
	return skyjoBotTu(g, me, "discard", 0, "")
}

// skyjoBotTauschziel sucht das beste Feld für die gezogene Karte: erst eine
// Spalte vollenden, dann die höchste offene Karte schlagen, sonst verdeckt.
func skyjoBotTauschziel(g *SkyjoGame, me, v int) (int, bool) {
	if i := skyjoBotSpalte(g, me, v); i >= 0 {
		return i, true
	}
	best, bestWert := -1, -1<<30
	for i := 0; i < skyjoCards; i++ {
		if !g.Open[me][i] || g.Removed[me][i] {
			continue
		}
		if g.Cards[me][i] > v && g.Cards[me][i] > bestWert {
			best, bestWert = i, g.Cards[me][i]
		}
	}
	if best >= 0 {
		return best, false
	}
	verdeckt := []int{}
	for i := 0; i < skyjoCards; i++ {
		if !g.Open[me][i] && !g.Removed[me][i] {
			verdeckt = append(verdeckt, i)
		}
	}
	if len(verdeckt) == 0 {
		return -1, false
	}
	return verdeckt[rand.Intn(len(verdeckt))], false
}

// skyjoBotSpalte findet ein verdecktes Feld, das mit zwei offenen gleichen
// Karten eine Spalte vollendet – nur mit der passenden Karte v.
func skyjoBotSpalte(g *SkyjoGame, me, v int) int {
	for col := 0; col < 4; col++ {
		idx := [3]int{col, col + 4, col + 8}
		gleich, verdeckt, offen := 0, -1, 0
		weg := false
		for _, i := range idx {
			if g.Removed[me][i] {
				weg = true
				break
			}
			if g.Open[me][i] {
				offen++
				if g.Cards[me][i] == v {
					gleich++
				}
			} else {
				verdeckt = i
			}
		}
		if !weg && offen == 2 && verdeckt >= 0 && gleich == 2 {
			return verdeckt
		}
	}
	return -1
}

// skyjoBotVerdeckt wählt eine verdeckte Karte (beim Aufdecken am Start).
func skyjoBotVerdeckt(g *SkyjoGame, me int) int {
	kandidaten := []int{}
	for i := 0; i < skyjoCards; i++ {
		if !g.Open[me][i] && !g.Removed[me][i] {
			kandidaten = append(kandidaten, i)
		}
	}
	if len(kandidaten) == 0 {
		return -1
	}
	return kandidaten[rand.Intn(len(kandidaten))]
}

// skyjoBotAufdecken wählt die Karte, die nach dem Abwerfen aufgedeckt wird.
// Sie kann eine Spalte vollenden; sonst entscheidet der Zufall.
func skyjoBotAufdecken(g *SkyjoGame, me int) int {
	for col := 0; col < 4; col++ {
		idx := [3]int{col, col + 4, col + 8}
		verdeckt, wert, offen, weg := -1, 0, 0, false
		for _, i := range idx {
			if g.Removed[me][i] {
				weg = true
				break
			}
			if g.Open[me][i] {
				if offen == 0 {
					wert = g.Cards[me][i]
				} else if g.Cards[me][i] != wert {
					wert = -1 << 30
				}
				offen++
			} else {
				verdeckt = i
			}
		}
		if !weg && offen == 2 && verdeckt >= 0 && wert != -1<<30 {
			return verdeckt
		}
	}
	return skyjoBotVerdeckt(g, me)
}
