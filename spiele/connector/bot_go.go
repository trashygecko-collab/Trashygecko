package main

/* ===========================================================================
   BOT FÜR GO (9 × 9)

   Der Bot setzt einen Stein nach derselben Prüfung wie ein Mensch
   (goAktion): belegt, selbstmörderisch und Ko sind verboten. Er sieht nur das
   Brett, das auch ein Mensch sieht.

   Was er bewertet:
     · gefangene Steine zählen viel,
     · eine eigene Gruppe mit nur einer Freiheit (Selbst-Atari) ist schlecht,
     · eigene Steine in der Nachbarschaft und die dritte/vierte Linie sind gut,
     · die erste Linie ist meist schlecht,
     · ein eigenes Auge wird nicht zugefüllt.

   Die Stufen:
     · leicht setzt irgendeinen erlaubten Punkt (zufällig durchprobiert),
     · mittel nimmt den besten nach dieser Bewertung (die 24 besten Punkte
       werden geprüft),
     · schwer prüft alle freien Punkte, füllt kein eigenes Gebiet zu und zieht
       die beste Schlag-Antwort des Gegners ab, bevor es sich entscheidet.

   Gepasst wird, wenn kein Zug mehr etwas bringt – zwei Pässe beenden die
   Partie, wie im Spiel.
   =========================================================================== */

import (
	"math/rand"
	"sort"
	"time"
)

// goBotZufall liefert den eigenen Zufallsgenerator des Bots. Er hängt am
// Spielstand statt am globalen Zufall: so bleibt ein Test reproduzierbar und
// stört andere Zufallsspieler im selben Lauf nicht.
func (g *GoGame) goBotZufall() *rand.Rand {
	if g.BotZufall == nil {
		saat := int64(len(g.Positions))*2654435761 + int64(len(g.ID))*97
		for i := 0; i < len(g.ID) && i < 8; i++ {
			saat = saat*31 + int64(g.ID[i])
		}
		g.BotZufall = rand.New(rand.NewSource(saat))
	}
	return g.BotZufall
}

// goBotKandidat ist ein Punkt mit seiner Vorbewertung (für die Reihenfolge).
type goBotKandidat struct {
	r, c  int
	vorab float64
}

// goBotFarbe ist die Farbe des Bots ("" wenn keiner mitspielt).
func goBotFarbe(g *GoGame) string {
	if g == nil || len(g.Bots) == 0 {
		return ""
	}
	return goColorOf(g, g.Bots[0])
}

// goBotAuge erkennt ein eigenes Auge: alle Nachbarn sind eigene Steine.
func goBotAuge(g *GoGame, r, c int, farbe string) bool {
	for _, st := range goNeighborSteps {
		nr, nc := r+st[0], c+st[1]
		if !goInside(nr, nc) {
			return false
		}
		if g.Board[nr][nc] == nil || *g.Board[nr][nc] != farbe {
			return false
		}
	}
	return true
}

// goBotEigenesGebiet markiert alle leeren Punkte, deren Gebiet nur von
// eigenen Steinen umschlossen ist – die füllt ein guter Spieler nicht zu.
func goBotEigenesGebiet(g *GoGame, farbe string) [][]bool {
	gebiet := make([][]bool, goSize)
	for r := range gebiet {
		gebiet[r] = make([]bool, goSize)
	}
	gesehen := make([][]bool, goSize)
	for r := range gesehen {
		gesehen[r] = make([]bool, goSize)
	}

	for r := 0; r < goSize; r++ {
		for c := 0; c < goSize; c++ {
			if g.Board[r][c] != nil || gesehen[r][c] {
				continue
			}
			feld := [][2]int{{r, c}}
			gesehen[r][c] = true
			eigen := true
			for i := 0; i < len(feld); i++ {
				cur := feld[i]
				for _, st := range goNeighborSteps {
					nr, nc := cur[0]+st[0], cur[1]+st[1]
					if !goInside(nr, nc) {
						eigen = false // offenes Brett: kein Gebiet
						continue
					}
					if g.Board[nr][nc] == nil {
						if !gesehen[nr][nc] {
							gesehen[nr][nc] = true
							feld = append(feld, [2]int{nr, nc})
						}
						continue
					}
					if *g.Board[nr][nc] != farbe {
						eigen = false
					}
				}
			}
			if eigen {
				for _, p := range feld {
					gebiet[p[0]][p[1]] = true
				}
			}
		}
	}
	return gebiet
}

// goBotVorwert sortiert die Punkte grob vor: Nachbarschaft, Linie, Ecken.
func goBotVorwert(g *GoGame, r, c int) float64 {
	wert := 0.0
	for _, st := range goNeighborSteps {
		nr, nc := r+st[0], c+st[1]
		if !goInside(nr, nc) {
			wert += 0.5 // Rand
			continue
		}
		if g.Board[nr][nc] != nil {
			wert += 1.5
		}
	}
	// Dritte und vierte Linie sind die guten; die erste ist schwach.
	for _, v := range []int{r, c} {
		switch {
		case v == 0 || v == goSize-1:
			wert -= 1
		case v == 2 || v == goSize-3:
			wert += 0.8
		}
	}
	return wert
}

// goBotWert bewertet einen Zug: gefangene Steine, eigene Freiheiten, Atari
// beim Gegner, Nachbarschaft. legal ist false, wenn der Zug verboten ist.
func goBotWert(g *GoGame, r, c int, farbe string) (wert float64, legal bool) {
	next, taken, ok := goTry(g.Board, g.Positions, r, c, farbe)
	if !ok {
		return 0, false
	}
	wert = float64(taken) * 12

	_, frei := goGroup(next, r, c)
	switch {
	case frei == 1:
		wert -= 30 // Selbst-Atari: der Gegner schlägt sofort
	case frei == 2:
		wert -= 3
	default:
		wert += float64(frei) * 0.2
	}

	gesehen := map[[2]int]bool{}
	for _, st := range goNeighborSteps {
		nr, nc := r+st[0], c+st[1]
		if !goInside(nr, nc) || next[nr][nc] == nil || *next[nr][nc] != goOther(farbe) {
			continue
		}
		if gesehen[[2]int{nr, nc}] {
			continue
		}
		steine, gegnerFrei := goGroup(next, nr, nc)
		for _, s := range steine {
			gesehen[s] = true
		}
		if gegnerFrei == 1 {
			wert += 8 // dieser Zug bringt eine Gegnergruppe in Atari
		}
	}

	for _, st := range goNeighborSteps {
		nr, nc := r+st[0], c+st[1]
		if !goInside(nr, nc) {
			continue
		}
		if next[nr][nc] != nil {
			wert += 1.5
		}
	}

	if r == 0 || c == 0 || r == goSize-1 || c == goSize-1 {
		wert -= 1.5
	}
	return wert, true
}

// goBotBest sucht den besten Zug unter den besten grenze Kandidaten.
// verboten sind Punkte, die nicht gespielt werden sollen (eigenes Gebiet);
// ohneR/ohneC überspringt einen Punkt (-1 = keiner).
func goBotBest(g *GoGame, farbe string, grenze int, verboten [][]bool, ohneR, ohneC int) (best float64, br, bc int, ok bool) {
	kandidaten := make([]goBotKandidat, 0, goSize*goSize)
	for r := 0; r < goSize; r++ {
		for c := 0; c < goSize; c++ {
			if g.Board[r][c] != nil || (r == ohneR && c == ohneC) || goBotAuge(g, r, c, farbe) {
				continue
			}
			if verboten != nil && verboten[r][c] {
				continue
			}
			kandidaten = append(kandidaten, goBotKandidat{r, c, goBotVorwert(g, r, c)})
		}
	}
	sort.SliceStable(kandidaten, func(i, j int) bool {
		return kandidaten[i].vorab > kandidaten[j].vorab
	})
	if len(kandidaten) > grenze {
		kandidaten = kandidaten[:grenze]
	}

	best = -1e9
	for _, k := range kandidaten {
		wert, legal := goBotWert(g, k.r, k.c, farbe)
		if !legal {
			continue
		}
		if wert > best || !ok {
			best, br, bc, ok = wert, k.r, k.c, true
		}
	}
	return best, br, bc, ok
}

// goBotZug entscheidet über Zug oder Pass. ok ist false, wenn der Bot nicht
// am Zug ist.
func goBotZug(g *GoGame, farbe string) (r, c int, pass bool, ok bool) {
	if g == nil || g.Status != "active" || farbe == "" || g.Turn != farbe {
		return 0, 0, false, false
	}

	// Kein Punkt mehr frei: passen.
	leer := 0
	for r := 0; r < goSize; r++ {
		for c := 0; c < goSize; c++ {
			if g.Board[r][c] == nil {
				leer++
			}
		}
	}
	if leer == 0 {
		return 0, 0, true, true
	}

	stufe := botStufe(g.BotStufe)

	// Leicht: irgendein erlaubter Punkt, zufällig durchprobiert.
	if stufe == botLeicht {
		start := g.goBotZufall().Intn(goSize * goSize)
		for i := 0; i < goSize*goSize; i++ {
			n := (start + i) % (goSize * goSize)
			pr, pc := n/goSize, n%goSize
			if g.Board[pr][pc] != nil || goBotAuge(g, pr, pc, farbe) {
				continue
			}
			if _, legal := goBotWert(g, pr, pc, farbe); legal {
				return pr, pc, false, true
			}
		}
		return 0, 0, true, true
	}

	grenze := 6
	var verboten [][]bool
	if stufe == botSchwer {
		grenze = goSize * goSize
		verboten = goBotEigenesGebiet(g, farbe)
	}

	best, br, bc, gefunden := goBotBest(g, farbe, grenze, verboten, -1, -1)
	if !gefunden && verboten != nil {
		// Nur noch eigenes Gebiet übrig: dann eben dort.
		best, br, bc, gefunden = goBotBest(g, farbe, grenze, nil, -1, -1)
	}
	if !gefunden {
		return 0, 0, true, true
	}

	// Schwer zieht die beste Schlag-Antwort des Gegners ab. Über das Passen
	// entscheidet die Bewertung ohne Abzug – sonst passt der Bot zu früh.
	if stufe == botSchwer {
		rohBest := best
		best, br, bc = goBotMitAntwort(g, farbe, best, br, bc, grenze)
		best = rohBest
	}

	// Nichts mehr zu holen: passen (zwei Pässe beenden die Partie).
	if best <= 0 && (g.Passes >= 1 || leer <= 6) {
		return 0, 0, true, true
	}
	if best < -12 {
		return 0, 0, true, true
	}
	return br, bc, false, true
}

// goBotMitAntwort prüft den besten und den zweitbesten Zug samt der besten
// Schlag-Antwort des Gegners und liefert den besseren.
func goBotMitAntwort(g *GoGame, farbe string, best float64, br, bc, grenze int) (float64, int, int) {
	wert, sr, sc := goBotMitSchlag(g, farbe, br, bc, best)
	if zweit, pr, pc, ok := goBotBest(g, farbe, grenze, nil, br, bc); ok {
		if w2, r2, c2 := goBotMitSchlag(g, farbe, pr, pc, zweit); w2 > wert {
			wert, sr, sc = w2, r2, c2
		}
	}
	return wert, sr, sc
}

// goBotMitSchlag bewertet einen Zug samt der besten Schlag-Antwort.
func goBotMitSchlag(g *GoGame, farbe string, r, c int, wert float64) (float64, int, int) {
	next, _, legal := goTry(g.Board, g.Positions, r, c, farbe)
	if !legal {
		return wert, r, c
	}
	temp := &GoGame{
		Board:     next,
		Positions: append(append([]string(nil), g.Positions...), goKey(next)),
		Turn:      goOther(farbe),
	}
	if geschlagen := goBotSchlag(temp, goOther(farbe)); geschlagen > 0 {
		wert -= float64(geschlagen) * 10
	}
	return wert, r, c
}

// goBotSchlag sucht die beste Schlag-Antwort (nur Gefangene zählen).
func goBotSchlag(g *GoGame, farbe string) int {
	best := 0
	for r := 0; r < goSize; r++ {
		for c := 0; c < goSize; c++ {
			if g.Board[r][c] != nil || goBotAuge(g, r, c, farbe) {
				continue
			}
			if _, taken, legal := goTry(g.Board, g.Positions, r, c, farbe); legal && taken > best {
				best = taken
			}
		}
	}
	return best
}

// goBotTreiben lässt den Bot nach der Denkpause ziehen (über GET gerufen).
func goBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Go
	if g == nil || len(g.Bots) == 0 || g.Status != "active" {
		return false
	}
	if g.Turn != goBotFarbe(g) {
		g.BotFrist = time.Time{}
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
	return goBotSchritt(room)
}

// goBotSchritt führt genau einen Bot-Zug aus (dieselbe Prüfung wie für einen
// Menschen).
func goBotSchritt(room *Room) bool {
	g := room.Go
	r, c, pass, ok := goBotZug(g, goBotFarbe(g))
	if !ok {
		return false
	}
	aktion := "play"
	if pass {
		aktion = "pass"
	}
	_, meldung := goAktion(room, g.Bots[0], aktion, r, c)
	return meldung == ""
}
