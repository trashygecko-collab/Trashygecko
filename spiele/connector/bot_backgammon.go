package main

import (
	"math/rand"
	"strconv"
	"time"
)

/* ===========================================================================
   BOT FÜR BACKGAMMON

   Der Bot würfelt mit demselben Wurf wie ein Mensch und zieht nur erlaubte
   Züge (bgLegalMoves).

   Stufen:
     leicht  – nimmt zu 45 % irgendeinen erlaubten Zug.
     mittel  – bewertet jeden Zug: schlagen, sichern, ins Ziel bringen.
     schwer  – wie mittel, meidet aber zusätzlich Blots in der gegnerischen
               Heimat und baut die eigene Heimat zu.
   =========================================================================== */

func bgBotSpieler(g *BackgammonGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	if g.RedID == g.BotID {
		return 0
	}
	if g.BlueID == g.BotID {
		return 1
	}
	return -1
}

func bgBotTreiben(g *BackgammonGame, jetzt time.Time) bool {
	me := bgBotSpieler(g)
	if me < 0 || g.Status != "active" || g.Player != me {
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
	if !g.Rolled {
		bgRoll(g)
		return true
	}
	// Den ganzen Zug in einem Rutsch spielen, damit niemand auf vier
	// einzelne Abrufe warten muss.
	for i := 0; i < 8; i++ {
		if g.Status != "active" || g.Player != me || !bgHasMove(g.Points, g.Bar, me, g.Dice) {
			break
		}
		zug, ok := bgBotZug(g, me, g.BotStufe)
		if !ok {
			break
		}
		if !bgBotAnwenden(g, me, zug) {
			break
		}
	}
	if g.Status != "active" {
		return true
	}
	if !bgHasMove(g.Points, g.Bar, me, g.Dice) {
		bgNextTurn(g)
	}
	return true
}

// bgRoll ist der Würfelwurf aus dem Handler – auch für den Bot.
func bgRoll(g *BackgammonGame) {
	a := rand.Intn(6) + 1
	b := rand.Intn(6) + 1
	werte := []int{a, b}
	if a == b {
		werte = []int{a, a, a, a}
	}
	wuerfel := make([]bgDie, len(werte))
	for i, v := range werte {
		wuerfel[i] = bgDie{ID: "d" + strconv.Itoa(i), Value: v}
	}
	g.Dice = wuerfel
	g.Rolled = true
	if a == b {
		g.Note = "Pasch! Vier Züge mit der " + strconv.Itoa(a) + "."
	} else {
		g.Note = "Gewürfelt: " + strconv.Itoa(a) + " und " + strconv.Itoa(b) + "."
	}
	if !bgHasMove(g.Points, g.Bar, g.Player, g.Dice) {
		g.Note += " Kein Zug möglich."
		bgNextTurn(g)
		g.Note = bgName(g, g.Player) + " ist am Zug (kein Zug möglich)."
	}
}

func bgBotZug(g *BackgammonGame, me int, stufe string) (bgMove, bool) {
	kandidaten := []bgMove{}
	gesehen := map[string]bool{}
	for _, die := range g.Dice {
		if die.Used {
			continue
		}
		for _, zug := range bgLegalMoves(g.Points, g.Bar, me, die.Value) {
			schluessel := strconv.Itoa(zug.From) + ">" + strconv.Itoa(zug.To) + "/" + strconv.Itoa(zug.Die)
			if !gesehen[schluessel] {
				gesehen[schluessel] = true
				kandidaten = append(kandidaten, zug)
			}
		}
	}
	if len(kandidaten) == 0 {
		return bgMove{}, false
	}
	if botStufe(stufe) == botLeicht && botZufall().Intn(100) < 45 {
		return kandidaten[botZufall().Intn(len(kandidaten))], true
	}
	beste := kandidaten[0]
	besterWert := -1 << 30
	for _, zug := range kandidaten {
		points, bar, off := bgApply(g.Points, g.Bar, g.Off, me, zug)
		wert := bgBotWert(points, bar, off, me, botStufe(stufe) == botSchwer)
		if wert > besterWert || (wert == besterWert && botZufall().Intn(2) == 0) {
			beste, besterWert = zug, wert
		}
	}
	return beste, true
}

func bgBotAnwenden(g *BackgammonGame, me int, zug bgMove) bool {
	if g == nil || g.Status != "active" || g.Player != me {
		return false
	}
	erlaubt := false
	for _, m := range bgLegalMoves(g.Points, g.Bar, me, zug.Die) {
		if m.From == zug.From && m.To == zug.To {
			zug = m
			erlaubt = true
			break
		}
	}
	if !erlaubt {
		return false
	}
	wuerfelOK := false
	for i := range g.Dice {
		if !g.Dice[i].Used && g.Dice[i].Value == zug.Die {
			g.Dice[i].Used = true
			wuerfelOK = true
			break
		}
	}
	if !wuerfelOK {
		return false
	}
	g.Points, g.Bar, g.Off = bgApply(g.Points, g.Bar, g.Off, me, zug)
	if g.Off[me] == 15 {
		bgFinish(g)
		return true
	}
	if !bgHasMove(g.Points, g.Bar, me, g.Dice) {
		bgNextTurn(g)
		return true
	}
	g.Note = "Gezogen – " + bgName(g, me) + " ist weiter am Zug."
	return true
}

func bgBotWert(points []int, bar [2]int, off [2]int, me int, vorsichtig bool) int {
	gegner := 1 - me
	wert := off[me]*1000 - off[gegner]*1000
	wert -= bar[me] * 120
	wert += bar[gegner] * 90

	// Pip-Abstand: je weniger Augen, desto besser.
	eigenePips, gegnerPips := 0, 0
	eigeneBlots, gegnerBlots := 0, 0
	eigenePunkte, gegnerPunkte := 0, 0
	gefaehrdet := 0
	for i := 0; i < 24; i++ {
		eigen := bgCountOn(points, i, me)
		fremd := bgCountOn(points, i, gegner)
		if me == 0 {
			eigenePips += eigen * (i + 1)
			gegnerPips += fremd * (24 - i)
		} else {
			eigenePips += eigen * (24 - i)
			gegnerPips += fremd * (i + 1)
		}
		if eigen == 1 {
			eigeneBlots++
			// Ein gegnerischer Stein 1–6 Felder davor kann schlagen.
			richtung := -1
			if me == 1 {
				richtung = 1
			}
			for schritt := 1; schritt <= 6; schritt++ {
				j := i + richtung*schritt
				if j < 0 || j >= 24 {
					break
				}
				if bgCountOn(points, j, gegner) > 0 {
					gefaehrdet++
					break
				}
			}
		}
		if fremd == 1 {
			gegnerBlots++
		}
		if eigen >= 2 {
			eigenePunkte++
		}
		if fremd >= 2 {
			gegnerPunkte++
		}
	}
	wert += (gegnerPips - eigenePips) * 2
	wert += (eigenePunkte - gegnerPunkte) * 12
	wert -= eigeneBlots * 8
	wert += gegnerBlots * 5
	if vorsichtig {
		wert -= gefaehrdet * 22
		// Die eigene Heimat ist der sicherste Ort.
		for i := 0; i < 6; i++ {
			if bgCountOn(points, i, me) >= 2 {
				wert += 6
			}
		}
	}
	return wert
}
