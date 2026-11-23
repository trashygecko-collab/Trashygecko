package main

import (
	"math/rand"
	"time"
)

/* Mastermind-Bot.

   Beide Spieler raten denselben Code, alle Rückmeldungen liegen offen auf dem
   Tisch. Der Bot nutzt das je nach Stufe:

     leicht – rät ohne Rücksicht auf die Rückmeldungen,
     mittel – behält alle Codes, die zu den Rückmeldungen passen, und rät einen
              davon,
     schwer – wählt den Versuch, der die Kandidaten im schlimmsten Fall am
              stärksten zusammenschrumpft (Minimax über alle Codes). */

// mastermindAlleCodes zählt alle möglichen Codes auf.
func mastermindAlleCodes(farben int) [][]int {
	codes := [][]int{}
	for a := 0; a < farben; a++ {
		for b := 0; b < farben; b++ {
			for c := 0; c < farben; c++ {
				for d := 0; d < farben; d++ {
					codes = append(codes, []int{a, b, c, d})
				}
			}
		}
	}
	return codes
}

// mastermindBotKandidaten lässt nur Codes übrig, die zu allen bisherigen
// Rückmeldungen passen.
func mastermindBotKandidaten(g *MastermindGame) [][]int {
	return mastermindBotKandidatenMit(g, false)
}

// mastermindBotKandidatenMit kann die weißen Treffer übergehen. Die mittlere
// Stufe tut das: sie achtet nur auf die richtige Stelle, nicht auf die richtige
// Farbe an falscher Stelle.
func mastermindBotKandidatenMit(g *MastermindGame, nurSchwarz bool) [][]int {
	kandidaten := [][]int{}
	for _, code := range mastermindAlleCodes(g.Farben) {
		passt := true
		for _, reihe := range g.Reihen {
			s, w := mastermindWerten(reihe.Rate, code)
			if s != reihe.Schwarz || (!nurSchwarz && w != reihe.Weiss) {
				passt = false
				break
			}
		}
		if passt {
			kandidaten = append(kandidaten, code)
		}
	}
	return kandidaten
}

func mastermindBotZufall(g *MastermindGame) *rand.Rand {
	saat := int64(len(g.Reihen)*7919 + g.Farben)
	for _, reihe := range g.Reihen {
		saat = saat*31 + int64(reihe.Schwarz*7+reihe.Weiss)
		for _, f := range reihe.Rate {
			saat = saat*13 + int64(f+1)
		}
	}
	return rand.New(rand.NewSource(saat))
}

// mastermindBotRate wählt den nächsten Versuch.
func mastermindBotRate(g *MastermindGame) []int {
	rnd := mastermindBotZufall(g)
	kandidaten := mastermindBotKandidaten(g)
	if g.BotStufe == botMittel {
		kandidaten = mastermindBotKandidatenMit(g, true)
	}

	switch g.BotStufe {
	case botLeicht:
		rate := make([]int, mastermindPlatz)
		for i := range rate {
			rate[i] = rnd.Intn(g.Farben)
		}
		return rate
	case botMittel:
		// Meist ein Kandidat, der zu den schwarzen Treffern passt, gelegentlich
		// ein danebener Versuch.
		if len(kandidaten) == 0 || rnd.Intn(100) < 15 {
			rate := make([]int, mastermindPlatz)
			for i := range rate {
				rate[i] = rnd.Intn(g.Farben)
			}
			return rate
		}
		return append([]int(nil), kandidaten[rnd.Intn(len(kandidaten))]...)
	}

	if len(kandidaten) == 0 {
		kandidaten = mastermindAlleCodes(g.Farben)
	}
	if len(kandidaten) == 1 {
		return append([]int(nil), kandidaten[0]...)
	}

	// Minimax: der Versuch mit dem kleinsten schlimmsten Rest. Bei
	// Gleichstand gewinnt ein Versuch, der selbst noch Kandidat ist, dann der,
	// der mehr Fälle unterscheidet.
	kandidat := map[string]bool{}
	for _, code := range kandidaten {
		kandidat[mastermindSchluessel(code)] = true
	}
	bestesWorst := len(kandidaten) + 1
	besteGruppen := -1
	besteKandidat := false
	var beste []int
	for _, versuch := range mastermindAlleCodes(g.Farben) {
		gruppen := map[int]int{}
		worst := 0
		for _, code := range kandidaten {
			s, w := mastermindWerten(versuch, code)
			gruppen[s*10+w]++
			if gruppen[s*10+w] > worst {
				worst = gruppen[s*10+w]
			}
		}
		istKandidat := kandidat[mastermindSchluessel(versuch)]
		besser := worst < bestesWorst
		if worst == bestesWorst {
			if istKandidat != besteKandidat {
				besser = istKandidat
			} else if len(gruppen) > besteGruppen {
				besser = true
			}
		}
		if besser {
			beste = append([]int(nil), versuch...)
			bestesWorst = worst
			besteGruppen = len(gruppen)
			besteKandidat = istKandidat
		}
	}
	if beste != nil {
		return beste
	}
	return append([]int(nil), kandidaten[0]...)
}

// mastermindSchluessel macht einen Code als Kartenschlüssel brauchbar.
func mastermindSchluessel(code []int) string {
	schluessel := make([]byte, len(code))
	for i, f := range code {
		schluessel[i] = byte('a' + f)
	}
	return string(schluessel)
}

// mastermindBotTreiben lässt den Bot nach der Denkpause raten. Der Aufrufer
// hält die Sperre des Raums, deshalb rechnet diese Funktion nichts: Die
// Minimax-Suche über alle Codes dauert je nach Stellung einige hundert
// Millisekunden und würde jeden anderen Abruf blockieren. Sie läuft in einer
// eigenen Goroutine auf einem Schnappschuss und legt den Versuch in BotZug ab;
// der nächste Abruf spielt ihn aus.
func mastermindBotTreiben(room *Room, now time.Time) {
	g := room.Mastermind
	if g == nil || len(g.Bots) == 0 || g.Status != "active" || g.Turn != "blue" {
		return
	}
	if g.BotZug != nil {
		rate := g.BotZug
		g.BotZug = nil
		if _, meldung := mastermindAktion(room, g.Bots[0], "rate", rate); meldung != "" {
			g.BotFrist = now.Add(time.Second)
			return
		}
		g.BotFrist = now.Add(botWartezeit(g.BotStufe))
		return
	}
	if g.BotRechnet == g || !now.After(g.BotFrist) {
		return
	}

	g.BotRechnet = g
	g.BotGen++
	gen := g.BotGen
	schnappschuss := cloneMastermind(g)
	wache := botWacheStarten(botRechenFrist,
		func() bool { return room.Mastermind == g && g.BotRechnet == g && g.BotGen == gen },
		func() {
			g.BotRechnet = nil
			g.BotFrist = time.Now()
		})
	go func() {
		rate := mastermindBotRate(schnappschuss)
		mu.Lock()
		defer mu.Unlock()
		wache.Stop()
		mastermindBotErgebnis(room, g, gen, rate)
	}()
}

// mastermindBotErgebnis übernimmt einen fertig gerechneten Versuch. quelle ist
// die Partie, für die gerechnet wurde: Wurde sie inzwischen ersetzt (Reset oder
// neue Bot-Partie), verfällt das Ergebnis, ohne die neue Partie anzufassen –
// insbesondere bleibt deren Rechen-Zustand stehen. Der Aufrufer hält die Sperre.
func mastermindBotErgebnis(room *Room, quelle *MastermindGame, gen int, rate []int) bool {
	if quelle == nil || room.Mastermind != quelle || quelle.BotGen != gen {
		return false
	}
	quelle.BotRechnet = nil
	quelle.BotZug = rate
	quelle.BotFrist = time.Now()
	return true
}

// mastermindBotSchritt rät ohne Denkpause – für Tests.
func mastermindBotSchritt(room *Room) bool {
	g := room.Mastermind
	if g == nil || len(g.Bots) == 0 || g.Status != "active" || g.Turn != "blue" {
		return false
	}
	_, meldung := mastermindAktion(room, g.Bots[0], "rate", mastermindBotRate(g))
	return meldung == ""
}
