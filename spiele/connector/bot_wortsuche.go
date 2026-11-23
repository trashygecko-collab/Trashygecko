package main

/* ===========================================================================
   BOT FÜR DIE WORTSUCHE

   Die Wortsuche ist kein Brett, an dem beide abwechselnd ziehen: alle sehen
   dasselbe Buchstabengitter und dieselbe Wortliste, und wer ein Wort zuerst
   findet, bekommt die Punkte. Der Server kennt Gitter und Wörter – der Bot
   arbeitet also auf demselben Stand wie ein Mensch und schickt seinen Fund
   durch dieselbe Prüfung (`wortsucheFinden`). Schummeln kann er nicht: was
   nicht als Linie im Gitter steht, kommt nicht durch.

   Ein Bot ist kein Sucher, der sofort alles sieht: er braucht je Wort eine
   Denkpause. Die Stufen unterscheiden sich darin und in der Wortwahl:

     · leicht  – 14–17 s je Wort, nimmt irgendeinen Fund (oft ein kurzes Wort)
     · mittel  –  9–11 s je Wort, nimmt das längste offene Wort
     · schwer  –  5–6 s je Wort, nimmt das längste offene Wort

   In einer Runde von 180 Sekunden schafft leicht damit höchstens elf Wörter,
   mittel achtzehn, schwer über dreißig – die acht Wörter einer Runde sind für
   jede Stufe zu schaffen, die Stufen unterscheiden sich im Tempo.
   =========================================================================== */

import (
	"time"
)

// wortsucheBotPause ist die Denkpause bis zum nächsten gefundenen Wort.
func wortsucheBotPause(stufe string) time.Duration {
	min, span := 9*time.Second, 2*time.Second
	switch botStufe(stufe) {
	case botLeicht:
		min, span = 14*time.Second, 3*time.Second
	case botSchwer:
		min, span = 5*time.Second, time.Second
	}
	return min + time.Duration(botZufall().Int63n(int64(span)))
}

// wortsucheZug ist ein Fund: das Wort und die beiden Enden seiner Linie im
// Gitter (wie ein Mensch sie mit dem Finger zieht).
type wortsucheZug struct {
	Wort string
	Von  [2]int
	Bis  [2]int
}

// wortsucheBotKandidaten sucht für jedes noch offene Wort eine Linie im
// Gitter. Wörter, die schon gefunden sind, kommen nicht vor – sie bringen
// keine Punkte mehr.
func wortsucheBotKandidaten(g *WortsucheGame) []wortsucheZug {
	if g == nil || g.Kante <= 0 || len(g.Gitter) != g.Kante {
		return nil
	}
	aus := make([]wortsucheZug, 0, len(g.Worte))
	for _, wort := range g.Worte {
		if g.gefunden(wort) {
			continue
		}
		if zug, ok := wortsucheBotLinie(g, wort); ok {
			aus = append(aus, zug)
		}
	}
	return aus
}

// wortsucheBotLinie sucht ein Wort waagerecht, senkrecht und in beiden
// Schrägen – vorwärts wie rückwärts, so wie ein Mensch liest.
func wortsucheBotLinie(g *WortsucheGame, wort string) (wortsucheZug, bool) {
	zeichen := []rune(wort)
	if len(zeichen) < 2 || len(zeichen) > g.Kante {
		return wortsucheZug{}, false
	}
	for _, richtung := range wortsucheRichtungen {
		for z := 0; z < g.Kante; z++ {
			for s := 0; s < g.Kante; s++ {
				if !wortsucheBotPasst(g, zeichen, z, s, richtung) {
					continue
				}
				endeZ := z + (len(zeichen)-1)*richtung[0]
				endeS := s + (len(zeichen)-1)*richtung[1]
				return wortsucheZug{Wort: wort, Von: [2]int{z, s}, Bis: [2]int{endeZ, endeS}}, true
			}
		}
	}
	return wortsucheZug{}, false
}

// wortsucheBotPasst prüft, ob das Wort ab (z,s) in Richtung passt.
func wortsucheBotPasst(g *WortsucheGame, zeichen []rune, z, s int, richtung [2]int) bool {
	for k, r := range zeichen {
		zz := z + k*richtung[0]
		ss := s + k*richtung[1]
		if zz < 0 || ss < 0 || zz >= g.Kante || ss >= g.Kante {
			return false
		}
		if g.Gitter[zz][ss] != string(r) {
			return false
		}
	}
	return true
}

// wortsucheBotWahl nimmt je Stufe einen der Funde: leicht irgendeinen, mittel
// und schwer den längsten (mehr Buchstaben, mehr Punkte).
func wortsucheBotWahl(kandidaten []wortsucheZug, stufe string) (wortsucheZug, bool) {
	if len(kandidaten) == 0 {
		return wortsucheZug{}, false
	}
	if botStufe(stufe) == botLeicht {
		return kandidaten[botZufall().Intn(len(kandidaten))], true
	}
	best := kandidaten[0]
	for _, k := range kandidaten[1:] {
		if len([]rune(k.Wort)) > len([]rune(best.Wort)) {
			best = k
		}
	}
	return best, true
}

// wortsucheBotTreiben lässt den Bot genau ein Wort finden, sobald seine
// Denkpause um ist. Liefert true, wenn in diesem Aufruf ein Wort dazukam.
func wortsucheBotTreiben(g *WortsucheGame, jetzt time.Time) bool {
	if g == nil || g.Status != "spiel" || g.BotID == "" {
		return false
	}
	if !g.BotFrist.IsZero() && jetzt.Before(g.BotFrist) {
		return false
	}
	sitz := g.sitz(g.BotID)
	if sitz < 0 {
		return false
	}
	g.BotFrist = jetzt.Add(wortsucheBotPause(g.BotStufe))
	zug, ok := wortsucheBotWahl(wortsucheBotKandidaten(g), g.BotStufe)
	if !ok {
		return false
	}
	text, _ := g.wortsucheFinden(sitz, zug.Wort, zug.Von, zug.Bis)
	return text == ""
}
