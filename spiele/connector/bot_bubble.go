package main

/* ===========================================================================
   COMPUTERGEGNER FÜR BUBBLE SHOOTER

   Das Duell rechnet der Server ohnehin selbst (bubble.go: Kugelflug, Anhängen,
   Dreiergruppen, Absturz, goldene Linie). Der Bot braucht deshalb keine eigene
   Regel – er probiert Ziele auf einer Kopie des Brettes durch und nimmt den
   Schuss mit dem besten Ergebnis. Geschossen wird über dieselbe Prüfung wie
   beim Menschen: bubBrett.schiessen(winkel) (nur bei laufendem Brett und ohne
   Kugel in der Luft) und danach der echte Takt (schritt).

   Bewertet wird, was der Schuss auf dem Feld ändert: Punkte, weniger Kugeln,
   niedrigere Wand. Ein Schuss, der das Brett über die goldene Linie bringt,
   ist verboten teuer. Die Stufen unterscheiden sich in der Zahl der geprüften
   Winkel (leicht 7, mittel 15, schwer 25 plus Nachjustierung) und in der
   Denkpause.
   =========================================================================== */

import (
	"math"
	"time"
)

// bubKopie ist ein Brett zum Durchrechnen – samt Kugel und Zufallsstand, damit
// die Fortsetzung dieselbe wäre wie im Spiel.
func bubKopie(b *bubBrett) *bubBrett {
	c := *b
	c.Feld = make([]bubReihe, len(b.Feld))
	for i, r := range b.Feld {
		c.Feld[i] = bubReihe{V: r.V, Z: append([]int{}, r.Z...)}
	}
	if b.Kugel != nil {
		k := *b.Kugel
		c.Kugel = &k
	}
	return &c
}

// bubHoehe ist die unterste belegte Reihe – je höher sie liegt, desto näher
// ist das Brett der goldenen Linie.
func bubHoehe(b *bubBrett) int {
	for r := len(b.Feld) - 1; r >= 0; r-- {
		for c := 0; c < bubSp; c++ {
			if b.Feld[r].Z[c] != 0 {
				return r
			}
		}
	}
	return -1
}

// bubSimuliert schießt auf einer Kopie und lässt die Kugel bis zum Anhängen
// fliegen. false heißt: kein gültiger Schuss (Kugel fliegt nicht zu Ende).
func bubSimuliert(b *bubBrett, winkel float64) (*bubBrett, bool) {
	sim := bubKopie(b)
	if !sim.schiessen(winkel) {
		return nil, false
	}
	for i := 0; i < 900 && sim.Kugel != nil; i++ {
		sim.schritt()
	}
	if sim.Kugel != nil {
		return nil, false
	}
	return sim, true
}

// bubWert bewertet ein Brett nach einem Schuss. Punkte bringen vor allem
// Gruppen und Abstürze; jede gesparte Kugel und jede freie Reihe zählen
// zusätzlich. Ein verlorenes Brett ist ein Fehlschuss.
func bubWert(vorher, nach *bubBrett) float64 {
	if nach.Zustand == "ende" {
		return -100000
	}
	wert := float64(nach.Punkte - vorher.Punkte)
	wert += 12 * float64(vorher.kugeln()-nach.kugeln())
	wert -= 6 * float64(bubHoehe(nach))
	return wert
}

// bubWinkelListe liefert gleichmäßig verteilte Zielwinkel.
func bubWinkelListe(n int) []float64 {
	if n < 2 {
		return []float64{0}
	}
	aus := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		aus = append(aus, -bubWinkelMax+2*bubWinkelMax*float64(i)/float64(n-1))
	}
	return aus
}

// bubbleBotZiel sucht den besten Winkel für das eigene Brett.
func bubbleBotZiel(b *bubBrett, stufe string) (float64, bool) {
	anzahl := 15
	switch stufe {
	case botLeicht:
		anzahl = 7
	case botSchwer:
		anzahl = 25
	}
	type treffer struct {
		winkel float64
		wert   float64
	}
	var liste []treffer
	for _, w := range bubWinkelListe(anzahl) {
		sim, ok := bubSimuliert(b, w)
		if !ok {
			continue
		}
		liste = append(liste, treffer{winkel: w, wert: bubWert(b, sim)})
	}
	if len(liste) == 0 {
		return 0, false
	}
	// leicht schießt irgendeinen geprüften Winkel, mittel und schwer den besten.
	zufall := botZufall()
	bestes, besterWert := liste[zufall.Intn(len(liste))].winkel, math.Inf(-1)
	for _, t := range liste {
		if t.wert > besterWert {
			bestes, besterWert = t.winkel, t.wert
		}
	}
	if stufe == botLeicht && zufall.Intn(100) < 55 {
		return liste[zufall.Intn(len(liste))].winkel, true
	}
	// Schwer justiert um den besten Winkel nach: engere Winkel, feinere Suche.
	if stufe == botSchwer {
		for _, d := range []float64{-0.09, -0.06, -0.03, 0.03, 0.06, 0.09} {
			w := bestes + d
			if w < -bubWinkelMax || w > bubWinkelMax {
				continue
			}
			sim, ok := bubSimuliert(b, w)
			if !ok {
				continue
			}
			if wert := bubWert(b, sim); wert > besterWert {
				bestes, besterWert = w, wert
			}
		}
	}
	return bestes, true
}

// bubbleBotTreiben lässt den Bot schießen, sobald sein Feld frei ist und die
// Denkpause vorbei ist.
func bubbleBotTreiben(g *BubbleGame, jetzt time.Time) {
	if g == nil || g.Status != "spiel" || g.BotID == "" {
		return
	}
	sitz := bubbleSitz(g, g.BotID)
	if sitz < 0 {
		return
	}
	b := g.Spieler[sitz].Brett
	if b == nil || b.Zustand != "laeuft" || b.Kugel != nil {
		return
	}
	if g.BotFrist.IsZero() {
		g.BotFrist = jetzt.Add(botWartezeit(g.BotStufe))
		return
	}
	if jetzt.Before(g.BotFrist) {
		return
	}
	winkel, ok := bubbleBotZiel(b, g.BotStufe)
	if !ok {
		return
	}
	if b.schiessen(winkel) {
		g.BotFrist = time.Time{}
	}
}
