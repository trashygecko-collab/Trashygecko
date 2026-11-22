package main

import "time"

/* ===========================================================================
   BOT FÜR 21 STREICHHÖLZER

   Wer das letzte Holz nimmt, verliert. Verloren ist damit jede Lage, in der
   beim eigenen Zug 1, 5, 9, 13, 17 oder 21 Hölzer liegen (Rest 1 bei Vier).

   Stufen:
     leicht  – nimmt meist zufällig 1 bis 3 Hölzer.
     mittel  – spielt zu 80 % die richtige Zahl, sonst zufällig.
     schwer  – lässt immer 1, 5, 9, 13, 17 liegen, wenn das geht; sonst nimmt
               er ein Holz und hofft auf einen Fehler.
   =========================================================================== */

// streichholzBotTreiben lässt den Bot nehmen, sobald seine Denkpause vorbei ist.
func streichholzBotTreiben(g *StreichholzGame, jetzt time.Time) bool {
	if g == nil || g.Status != "active" || g.BotID == "" {
		return false
	}
	color := ""
	if g.Current == "red" && g.RedID == g.BotID {
		color = "red"
	} else if g.Current == "blue" && g.BlueID == g.BotID {
		color = "blue"
	}
	if color == "" {
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
	amount := streichholzBotZug(g.Remaining, g.BotStufe)
	if amount < 1 {
		return false
	}
	nr, _ := streichholzNehmen(g, color, amount)
	return nr == 0
}

// streichholzBotZug wählt 1 bis 3 Hölzer.
func streichholzBotZug(remaining int, stufe string) int {
	if remaining <= 0 {
		return 0
	}
	max := 3
	if remaining < max {
		max = remaining
	}
	stufe = botStufe(stufe)

	richtig := (remaining - 1) % 4
	if richtig < 1 || richtig > max {
		richtig = 1
	}

	switch stufe {
	case botLeicht:
		if botZufall().Intn(100) < 25 {
			return richtig
		}
	case botMittel:
		if botZufall().Intn(100) < 80 {
			return richtig
		}
	}
	// Schwer: immer die richtige Zahl, wenn sie möglich ist; in einer
	// verlorenen Lage (Rest 1 bei Vier) nimmt er ein Holz.
	if remaining%4 == 1 {
		return 1
	}
	return richtig
}
