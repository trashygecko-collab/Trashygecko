package main

/* ===========================================================================
   BOT FÜR KNIFFEL

   Der Bot würfelt und trägt über dieselben Helfer ein wie ein Mensch
   (kniffelWurf, kniffelEintrag) – er kann also nicht schummeln und keine
   Kategorie doppelt belegen.

   Stufen:
     leicht  – ein Nachwurf, hält zufällig; nimmt danach das beste offene Feld.
     mittel  – zwei Würfe, hält die Würfel zum besten offenen Feld.
     schwer  – drei Würfe, hört bei einer fertigen Figur (Full House, Straße,
                Kniffel) früher auf und hält gezielt darauf.

   Die Feldwahl ist bei allen Stufen dieselbe: das offene Feld mit den meisten
   Punkten; bringt keines etwas, wird das erste offene (eine Einser-Reihe)
   gestrichen.
   =========================================================================== */

import (
	"math/rand"
	"strconv"
	"time"
)

func kniffelBotSitz(g *KniffelGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return kniffelFindPlayer(g, g.BotID)
}

// kniffelBotTreiben lässt den Bot nach der Denkpause einen Zug machen.
func kniffelBotTreiben(g *KniffelGame, jetzt time.Time) bool {
	sitz := kniffelBotSitz(g)
	if sitz < 0 || g.Status != "active" || g.Current != sitz {
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
	return kniffelBotZug(g, sitz)
}

// kniffelBotZug macht genau einen Schritt: würfeln, nachwürfeln oder eintragen.
func kniffelBotZug(g *KniffelGame, sitz int) bool {
	if g == nil || g.Status != "active" || sitz < 0 || sitz >= len(g.Players) {
		return false
	}
	if g.Current != sitz {
		return false
	}
	name := g.Players[sitz].Name
	if g.Rolls == 0 {
		kniffelWurf(g, make([]bool, 5))
		g.Note = name + " würfelt."
		return true
	}
	ziel, wert := kniffelBotZiel(g.Players[sitz], g.Dice)
	if kniffelBotWeiter(g, wert) {
		kniffelWurf(g, kniffelBotHalten(g.Dice, ziel, g.BotStufe))
		g.Note = name + " würfelt noch einmal."
		return true
	}
	kniffelEintrag(g, sitz, ziel, nil)
	if g.Status == "active" {
		g.Note = name + " trägt " + strconv.Itoa(wert) + " Punkte ein."
	}
	return true
}

// kniffelBotZiel sucht das offene Feld mit den meisten Punkten. Ist überall
// null zu holen, wird das erste offene Feld gestrichen.
func kniffelBotZiel(p KniffelPlayer, dice []int) (string, int) {
	besCat, bes := "", -1
	for _, c := range kniffelCategories {
		if p.Scores[c] != nil {
			continue
		}
		pt := kniffelScoreFor(c, dice)
		if pt > bes {
			bes, besCat = pt, c
		}
	}
	if bes <= 0 {
		for _, c := range kniffelCategories {
			if p.Scores[c] == nil {
				return c, 0
			}
		}
	}
	return besCat, bes
}

// kniffelBotWeiter entscheidet, ob noch einmal gewürfelt wird.
func kniffelBotWeiter(g *KniffelGame, wert int) bool {
	max := 2
	switch botStufe(g.BotStufe) {
	case botLeicht:
		max = 2
	case botSchwer:
		max = 3
	}
	if g.Rolls >= max {
		return false
	}
	// Eine fertige Figur (Full House, Straße, Kniffel) nicht wieder zerwürfeln.
	return wert < 30
}

// kniffelBotHalten wählt die Würfel, die liegen bleiben.
func kniffelBotHalten(dice []int, ziel, stufe string) []bool {
	behalten := make([]bool, 5)
	if botStufe(stufe) == botLeicht {
		for i := range behalten {
			behalten[i] = rand.Intn(100) < 25
		}
		return behalten
	}
	anzahl := [7]int{}
	for _, w := range dice {
		if w >= 1 && w <= 6 {
			anzahl[w]++
		}
	}
	haltenWert := func(wert int) {
		for i, w := range dice {
			if w == wert {
				behalten[i] = true
			}
		}
	}
	switch ziel {
	case "ones", "twos", "threes", "fours", "fives", "sixes":
		face := map[string]int{"ones": 1, "twos": 2, "threes": 3, "fours": 4, "fives": 5, "sixes": 6}[ziel]
		haltenWert(face)
	case "three", "four", "kniffel":
		best := 0
		for w := 1; w <= 6; w++ {
			if anzahl[w] > anzahl[best] {
				best = w
			}
		}
		haltenWert(best)
	case "full":
		erst, zweit := 0, 0
		for w := 1; w <= 6; w++ {
			if anzahl[w] > anzahl[erst] {
				zweit, erst = erst, w
			} else if anzahl[w] > anzahl[zweit] {
				zweit = w
			}
		}
		haltenWert(erst)
		if anzahl[zweit] >= 2 {
			haltenWert(zweit)
		}
	case "small", "large":
		for _, w := range kniffelBotLauf(dice) {
			haltenWert(w)
		}
	case "chance":
		// Chance belohnt nur die Augensumme – alles neu würfeln.
	}
	return behalten
}

// kniffelBotLauf liefert die Werte der längsten Straße im Wurf.
func kniffelBotLauf(dice []int) []int {
	da := [7]bool{}
	for _, w := range dice {
		if w >= 1 && w <= 6 {
			da[w] = true
		}
	}
	best, bestStart := 0, 1
	for start := 1; start <= 6; start++ {
		n := 0
		for w := start; w <= 6 && da[w]; w++ {
			n++
		}
		if n >= best {
			best, bestStart = n, start
		}
	}
	werte := []int{}
	for w := bestStart; w < bestStart+best; w++ {
		werte = append(werte, w)
	}
	return werte
}
