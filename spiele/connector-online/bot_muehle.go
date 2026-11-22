package main

import "time"

/* ===========================================================================
   BOT FÜR MÜHLE

   Der Bot sieht nur das Brett und den Zugstand – dieselben Daten wie ein
   Mensch. Gesetzt, gezogen und geschlagen wird über die Funktionen des
   Spiels, also mit genau derselben Prüfung.

   Stufen:
     leicht  – spielt zu 35 % irgendeinen erlaubten Zug, sonst einen Schritt
               voraus (Mühle bauen, Mühle des Gegners verhindern).
     mittel  – zwei Züge voraus, mit ruhiger Bewertung.
     schwer  – vier Züge voraus (Alpha-Beta) und schlägt den Stein, der dem
               Gegner am meisten nimmt.
   =========================================================================== */

type muehleBotAktion struct {
	art  string // place, move, capture
	von  int
	nach int
}

func muehleBotFarbe(g *MuehleGame) string {
	if g == nil || g.BotID == "" {
		return ""
	}
	return muehleColorOf(g, g.BotID)
}

// muehleBotTreiben lässt den Bot ziehen, sobald die Denkpause vorbei ist.
func muehleBotTreiben(g *MuehleGame, jetzt time.Time) bool {
	farbe := muehleBotFarbe(g)
	if farbe == "" || g.Status != "active" || g.Current != farbe {
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
	aktion, ok := muehleBotZug(g, farbe, g.BotStufe)
	if !ok {
		return false
	}
	return muehleBotAnwenden(g, farbe, aktion)
}

// muehleBotZug wählt eine Aktion. Gibt es keine, ist die Partie vorbei.
func muehleBotZug(g *MuehleGame, farbe, stufe string) (muehleBotAktion, bool) {
	aktionen := muehleBotAktionen(g, farbe)
	if len(aktionen) == 0 {
		return muehleBotAktion{}, false
	}
	stufe = botStufe(stufe)
	if stufe == botLeicht && botZufall().Intn(100) < 35 {
		return aktionen[botZufall().Intn(len(aktionen))], true
	}

	tiefe := 2
	switch stufe {
	case botLeicht:
		tiefe = 1
	case botSchwer:
		tiefe = 4
	}
	beste := aktionen[0]
	besterWert := -1 << 30
	for _, aktion := range aktionen {
		probe := cloneMuehle(g)
		if !muehleBotAnwenden(probe, farbe, aktion) {
			continue
		}
		wert := muehleBotSuche(probe, farbe, tiefe-1, -1<<30, 1<<30)
		if wert > besterWert || (wert == besterWert && botZufall().Intn(2) == 0) {
			beste, besterWert = aktion, wert
		}
	}
	return beste, true
}

func muehleBotAktionen(g *MuehleGame, farbe string) []muehleBotAktion {
	if g == nil || g.Status != "active" || g.Current != farbe {
		return nil
	}
	gegner := muehleOther(farbe)
	aktionen := []muehleBotAktion{}
	if g.CapturePending {
		for i, feld := range g.State {
			if feld == gegner && muehleCanRemove(g, i, farbe) {
				aktionen = append(aktionen, muehleBotAktion{art: "capture", nach: i})
			}
		}
		return aktionen
	}
	if g.Phase == "placing" {
		for i, feld := range g.State {
			if feld == "" {
				aktionen = append(aktionen, muehleBotAktion{art: "place", nach: i})
			}
		}
		return aktionen
	}
	eigene := muehleCount(g, farbe)
	for i, feld := range g.State {
		if feld != farbe {
			continue
		}
		if eigene == 3 {
			for j, ziel := range g.State {
				if ziel == "" {
					aktionen = append(aktionen, muehleBotAktion{art: "move", von: i, nach: j})
				}
			}
			continue
		}
		for _, n := range muehleNeighbors[i] {
			if g.State[n] == "" {
				aktionen = append(aktionen, muehleBotAktion{art: "move", von: i, nach: n})
			}
		}
	}
	return aktionen
}

// muehleBotAnwenden führt die Aktion mit denselben Regeln wie der Handler aus.
func muehleBotAnwenden(g *MuehleGame, farbe string, aktion muehleBotAktion) bool {
	if g == nil || g.Status != "active" || g.Current != farbe {
		return false
	}
	switch aktion.art {
	case "capture":
		if !g.CapturePending || aktion.nach < 0 || aktion.nach >= 24 || !muehleCanRemove(g, aktion.nach, farbe) {
			return false
		}
		g.State[aktion.nach] = ""
		g.CapturePending = false
		g.Note = muehleName(g, farbe) + " schlägt einen Stein."
		if muehleCount(g, muehleOther(farbe)) <= 2 {
			muehleFinish(g, farbe)
			return true
		}
		muehleAdvance(g, farbe)
		return true
	case "place":
		if g.CapturePending || g.Phase != "placing" || aktion.nach < 0 || aktion.nach >= 24 || g.State[aktion.nach] != "" {
			return false
		}
		g.State[aktion.nach] = farbe
		if farbe == "red" {
			g.PlacedRed++
		} else {
			g.PlacedBlue++
		}
		g.Note = muehleName(g, farbe) + " setzt einen Stein."
		if muehleIsMill(g, aktion.nach, farbe) {
			g.CapturePending = true
			g.Note = "Mühle! " + muehleName(g, farbe) + " darf einen Stein entfernen."
			return true
		}
		muehleAdvance(g, farbe)
		return true
	case "move":
		if g.CapturePending || g.Phase != "moving" {
			return false
		}
		if aktion.von < 0 || aktion.von >= 24 || aktion.nach < 0 || aktion.nach >= 24 {
			return false
		}
		if g.State[aktion.von] != farbe || g.State[aktion.nach] != "" {
			return false
		}
		if muehleCount(g, farbe) > 3 && !muehleAdjacent(aktion.von, aktion.nach) {
			return false
		}
		g.State[aktion.nach] = farbe
		g.State[aktion.von] = ""
		g.Note = muehleName(g, farbe) + " zieht einen Stein."
		if muehleIsMill(g, aktion.nach, farbe) {
			g.CapturePending = true
			g.Note = "Mühle! " + muehleName(g, farbe) + " darf einen Stein entfernen."
			return true
		}
		muehleEndTurn(g, farbe)
		return true
	}
	return false
}

func muehleBotSuche(g *MuehleGame, farbe string, tiefe, alpha, beta int) int {
	if g.Status == "finished" || tiefe <= 0 {
		return muehleBotWert(g, farbe)
	}
	aktionen := muehleBotAktionen(g, g.Current)
	if len(aktionen) == 0 {
		return muehleBotWert(g, farbe)
	}
	if g.Current == farbe {
		best := -1 << 30
		for _, aktion := range aktionen {
			probe := cloneMuehle(g)
			if !muehleBotAnwenden(probe, g.Current, aktion) {
				continue
			}
			wert := muehleBotSuche(probe, farbe, tiefe-1, alpha, beta)
			if wert > best {
				best = wert
			}
			if best > alpha {
				alpha = best
			}
			if alpha >= beta {
				break
			}
		}
		return best
	}
	best := 1 << 30
	for _, aktion := range aktionen {
		probe := cloneMuehle(g)
		if !muehleBotAnwenden(probe, g.Current, aktion) {
			continue
		}
		wert := muehleBotSuche(probe, farbe, tiefe-1, alpha, beta)
		if wert < best {
			best = wert
		}
		if best < beta {
			beta = best
		}
		if alpha >= beta {
			break
		}
	}
	return best
}

// muehleBotWert bewertet aus Sicht von farbe. Steine in der Hand zählen mit,
// damit der Bot beim Setzen nicht blind Material hergibt.
func muehleBotWert(g *MuehleGame, farbe string) int {
	if g.Status == "finished" {
		if g.Winner == farbe {
			return 1000000
		}
		if g.Winner == "" {
			return 0
		}
		return -1000000
	}
	gegner := muehleOther(farbe)
	eigeneHand, gegnerHand := 0, 0
	if farbe == "red" {
		eigeneHand = 9 - g.PlacedRed
		gegnerHand = 9 - g.PlacedBlue
	} else {
		eigeneHand = 9 - g.PlacedBlue
		gegnerHand = 9 - g.PlacedRed
	}
	wert := (muehleCount(g, farbe) + eigeneHand - muehleCount(g, gegner) - gegnerHand) * 100

	eigeneMuehlen, gegnerMuehlen := 0, 0
	for _, mill := range muehleMills {
		if g.State[mill[0]] == farbe && g.State[mill[1]] == farbe && g.State[mill[2]] == farbe {
			eigeneMuehlen++
		}
		if g.State[mill[0]] == gegner && g.State[mill[1]] == gegner && g.State[mill[2]] == gegner {
			gegnerMuehlen++
		}
	}
	wert += (eigeneMuehlen - gegnerMuehlen) * 18

	// Zwei Steine in einer Mühle sind fast eine Mühle.
	offenEigen, offenGegner := 0, 0
	for _, mill := range muehleMills {
		eigen, fremd := 0, 0
		for _, i := range mill {
			switch g.State[i] {
			case farbe:
				eigen++
			case gegner:
				fremd++
			}
		}
		if eigen == 2 && fremd == 0 {
			offenEigen++
		}
		if fremd == 2 && eigen == 0 {
			offenGegner++
		}
	}
	wert += (offenEigen - offenGegner) * 6

	return wert
}
