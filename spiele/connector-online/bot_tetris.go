package main

/* ===========================================================================
   COMPUTERGEGNER FÜR TETRIS

   Tetris ist das einzige Spiel, in dem der Server das Feld nicht kennt: Jede
   Seite rechnet ihr Brett selbst und meldet nur Linien, Leben und Müll
   (tetris.go). Damit ein Bot am Server mitspielen kann, steht hier die
   Brett-Engine der Spielseite (spiele/tetris/tetris.html) in Go – dieselben
   sieben Steine, dieselbe 7er-Tüte aus mulberry32(seed), dieselben Drehungen
   samt Versatz, dieselbe Wertung der Reihen und derselbe Müll (2 → 1,
   3 → 2, 4 → 4, Loch zufällig).

   Der Bot sieht damit genau das, was ein Mensch auf seinem Gerät sieht:
   sein eigenes Feld und die gemeinsame Steinfolge. Er setzt seine Steine im
   Tempo eines flotten Menschen (vier je Sekunde): Das Fallintervall der
   Seite ist nur die Schwerkraft, ein Mensch lässt jede Figur sofort fallen.
   Die Entscheidung fällt über eine Bewertung des Bretts nach dem Absetzen:
   Löcher, Höhe, Buckel und geräumte Reihen (Viererreihe zählt extra).
   leicht rechnet ohne Lochstrafe und würfelt aus den besten fünf Zügen,
   mittel nimmt den besten, schwer rechnet zusätzlich den nächsten Stein mit.
   =========================================================================== */

import (
	"sort"
	"time"
)

const (
	tetSpalten = 10
	tetZeilen  = 20
	tetMuell   = 8 // Farbwert einer Müllreihe (wie auf der Spielseite)
)

var tetAlleSteine = []string{"I", "O", "T", "S", "Z", "J", "L"}

// tetFormen sind die Startformen der Spielseite, 1:1 übernommen.
var tetFormen = map[string][][]int{
	"I": {{1, 1, 1, 1}},
	"O": {{1, 1}, {1, 1}},
	"T": {{0, 1, 0}, {1, 1, 1}},
	"S": {{0, 1, 1}, {1, 1, 0}},
	"Z": {{1, 1, 0}, {0, 1, 1}},
	"J": {{1, 0, 0}, {1, 1, 1}},
	"L": {{0, 0, 1}, {1, 1, 1}},
}

// tetFeld ist das eigene Brett des Bots samt gemeinsamer Steinfolge.
type tetFeld struct {
	Z      [][]int // tetZeilen Reihen à tetSpalten Felder, 0 = leer
	Folge  []string
	Nummer int
	Zufall bubMulberry // derselbe mulberry32 wie auf der Spielseite
}

// tetFeldNeu baut ein leeres Feld mit der Steinfolge aus dem Partie-Seed.
func tetFeldNeu(seed int64) *tetFeld {
	f := &tetFeld{Zufall: bubMulberry{a: uint32(seed)}}
	f.Z = make([][]int, tetZeilen)
	for r := range f.Z {
		f.Z[r] = make([]int, tetSpalten)
	}
	return f
}

// stein liefert den Stein Nummer nr. Die Tüte wird wie im Browser gemischt
// (Fisher-Yates rückwärts über alle sieben Steine).
func (f *tetFeld) stein(nr int) string {
	for len(f.Folge) <= nr {
		beutel := append([]string{}, tetAlleSteine...)
		for i := len(beutel) - 1; i > 0; i-- {
			j := int(f.Zufall.next() * float64(i+1))
			if j > i {
				j = i
			}
			beutel[i], beutel[j] = beutel[j], beutel[i]
		}
		f.Folge = append(f.Folge, beutel...)
	}
	return f.Folge[nr]
}

func tetFormKopie(form [][]int) [][]int {
	c := make([][]int, len(form))
	for y := range form {
		c[y] = append([]int{}, form[y]...)
	}
	return c
}

// tetDrehen dreht eine Form wie drehen() auf der Spielseite.
func tetDrehen(form [][]int, richtung int) [][]int {
	hoehe := len(form)
	breite := len(form[0])
	ergebnis := make([][]int, breite)
	for y := range ergebnis {
		ergebnis[y] = make([]int, hoehe)
	}
	for y := 0; y < hoehe; y++ {
		for x := 0; x < breite; x++ {
			if form[y][x] == 0 {
				continue
			}
			if richtung > 0 {
				ergebnis[x][hoehe-1-y] = 1
			} else {
				ergebnis[breite-1-x][y] = 1
			}
		}
	}
	return ergebnis
}

// tetPasst prüft eine Form an einer Stelle – dieselbe Prüfung wie passt().
func tetPasst(z [][]int, form [][]int, px, py int) bool {
	for y := range form {
		for x := range form[y] {
			if form[y][x] == 0 {
				continue
			}
			bx, by := px+x, py+y
			if bx < 0 || bx >= tetSpalten || by >= tetZeilen {
				return false
			}
			if by >= 0 && z[by][bx] != 0 {
				return false
			}
		}
	}
	return true
}

func tetStartSpalte(form [][]int) int {
	return (tetSpalten - len(form[0])) / 2
}

// tetEinsetzen setzt eine Form ab, räumt volle Reihen und liefert das neue
// Brett mit der Zahl der geräumten Reihen. Das alte Brett bleibt unberührt.
func tetEinsetzen(z [][]int, form [][]int, px, py int) ([][]int, int) {
	neu := make([][]int, tetZeilen)
	for r := range z {
		neu[r] = append([]int{}, z[r]...)
	}
	for y := range form {
		for x := range form[y] {
			if form[y][x] == 0 {
				continue
			}
			bx, by := px+x, py+y
			if by >= 0 && by < tetZeilen && bx >= 0 && bx < tetSpalten {
				neu[by][bx] = 1
			}
		}
	}
	volle := 0
	rest := make([][]int, 0, tetZeilen)
	for _, reihe := range neu {
		frei := false
		for _, wert := range reihe {
			if wert == 0 {
				frei = true
				break
			}
		}
		if frei {
			rest = append(rest, reihe)
		} else {
			volle++
		}
	}
	for len(rest) < tetZeilen {
		rest = append([][]int{make([]int, tetSpalten)}, rest...)
	}
	return rest, volle
}

// tetKennzahlen liefert Gesamthöhe, Löcher und Buckel eines Bretts.
func tetKennzahlen(z [][]int) (hoehen, loecher, buckel int) {
	h := make([]int, tetSpalten)
	for c := 0; c < tetSpalten; c++ {
		for r := 0; r < tetZeilen; r++ {
			if z[r][c] != 0 {
				h[c] = tetZeilen - r
				break
			}
		}
	}
	for c := 0; c < tetSpalten; c++ {
		hoehen += h[c]
		gesehen := false
		for r := 0; r < tetZeilen; r++ {
			if z[r][c] != 0 {
				gesehen = true
				continue
			}
			if gesehen {
				loecher++
			}
		}
	}
	for c := 0; c+1 < tetSpalten; c++ {
		d := h[c] - h[c+1]
		if d < 0 {
			d = -d
		}
		buckel += d
	}
	return
}

// tetWert bewertet ein Brett nach dem Absetzen. Geräumte Reihen zählen mit
// Bonus, damit eine Viererreihe mehr wert ist als vier einzelne.
func tetWert(z [][]int, reihen int) float64 {
	hoehen, loecher, buckel := tetKennzahlen(z)
	w := -0.51*float64(hoehen) - 0.36*float64(loecher) - 0.18*float64(buckel)
	switch reihen {
	case 1:
		w += 0.76
	case 2:
		w += 2.4
	case 3:
		w += 5.2
	case 4:
		w += 9.5
	}
	return w
}

// tetWertLeicht übersieht Löcher und Buckel und freut sich über jede Reihe –
// die schwache Stufe baut sich damit zu.
func tetWertLeicht(z [][]int, reihen int) float64 {
	hoehen, _, _ := tetKennzahlen(z)
	return -0.5*float64(hoehen) + 0.9*float64(reihen)
}

type tetZug struct {
	Form    [][]int
	Spalte  int
	Reihe   int
	Drehung int
	Wert    float64
}

// tetKandidaten probiert alle Drehungen und Spalten auf einer Kopie durch und
// bewertet das Ergebnis. Nicht erreichbare Stellen (die Form passt schon oben
// nicht) fallen weg.
func tetKandidaten(z [][]int, form [][]int, leicht bool) []tetZug {
	var liste []tetZug
	for d := 0; d < 4; d++ {
		m := tetFormKopie(form)
		for i := 0; i < d; i++ {
			m = tetDrehen(m, 1)
		}
		for px := -3; px < tetSpalten+3; px++ {
			if !tetPasst(z, m, px, 0) {
				continue
			}
			py := 0
			for tetPasst(z, m, px, py+1) {
				py++
			}
			neu, reihen := tetEinsetzen(z, m, px, py)
			wert := tetWert(neu, reihen)
			if leicht {
				wert = tetWertLeicht(neu, reihen)
			}
			liste = append(liste, tetZug{Form: m, Spalte: px, Reihe: py, Drehung: d, Wert: wert})
		}
	}
	sort.SliceStable(liste, func(i, j int) bool { return liste[i].Wert > liste[j].Wert })
	return liste
}

// tetBotWahl bestimmt den Zug für den aktuellen Stein des Feldes.
func tetBotWahl(f *tetFeld, form [][]int, stufe string) (tetZug, bool) {
	liste := tetKandidaten(f.Z, form, stufe == botLeicht)
	if len(liste) == 0 {
		return tetZug{}, false
	}
	if stufe == botLeicht {
		// Aus den fünf besten einen würfeln: gut, aber nicht verlässlich.
		n := len(liste)
		if n > 5 {
			n = 5
		}
		return liste[botZufall().Intn(n)], true
	}
	if stufe != botSchwer {
		return liste[0], true
	}
	// schwer: den nächsten Stein der gemeinsamen Folge mitrechnen.
	naechste := tetFormKopie(tetFormen[f.stein(f.Nummer+1)])
	bestes, besterWert := liste[0], liste[0].Wert
	for i := 0; i < len(liste) && i < 12; i++ {
		neu, _ := tetEinsetzen(f.Z, liste[i].Form, liste[i].Spalte, liste[i].Reihe)
		folge := tetKandidaten(neu, naechste, false)
		wert := liste[i].Wert
		if len(folge) > 0 {
			wert += 0.55 * folge[0].Wert
		}
		if wert > besterWert {
			bestes, besterWert = liste[i], wert
		}
	}
	return bestes, true
}

// tetBotZug setzt einen Stein und meldet, ob danach kein Platz mehr für den
// nächsten ist (dann ist die Partie vorbei).
func tetBotZug(f *tetFeld, stufe string) (reihen int, vorbei bool) {
	art := f.stein(f.Nummer)
	form := tetFormKopie(tetFormen[art])
	zug, ok := tetBotWahl(f, form, stufe)
	if !ok {
		return 0, true
	}
	neu, geraeumt := tetEinsetzen(f.Z, zug.Form, zug.Spalte, zug.Reihe)
	f.Z = neu
	f.Nummer++
	if !tetBotPasst(f) {
		return geraeumt, true
	}
	return geraeumt, false
}

// tetBotPasst: Steht der Stein, der gerade dran ist, an seinem Startplatz?
// Dort prüft die Seite beim neuen Stein (neuesStueck) und nach Müll
// (muellEmpfangen prüft die aktuelle Stelle, also die Startstelle, solange
// nichts bewegt wurde) – auch wenn eine andere Spalte oben noch Platz hätte.
func tetBotPasst(f *tetFeld) bool {
	form := tetFormen[f.stein(f.Nummer)]
	return tetPasst(f.Z, form, tetStartSpalte(form), 0)
}

// tetMuellFuerReihen ist die Müllregel der Spielseite.
func tetMuellFuerReihen(reihen int) int {
	switch reihen {
	case 2:
		return 1
	case 3:
		return 2
	case 4:
		return 4
	}
	return 0
}

// tetFeldMuell schiebt Müllreihen von unten herein, das Loch steht in Spalte
// loch. Ist die oberste Reihe schon belegt, ist das Feld tot.
func tetFeldMuell(z [][]int, anzahl, loch int) ([][]int, bool) {
	for i := 0; i < anzahl; i++ {
		for _, wert := range z[0] {
			if wert != 0 {
				return z, true
			}
		}
		reihe := make([]int, tetSpalten)
		for c := range reihe {
			reihe[c] = tetMuell
		}
		reihe[loch%tetSpalten] = 0
		neu := make([][]int, tetZeilen)
		for r := 0; r < tetZeilen-1; r++ {
			neu[r] = z[r+1]
		}
		neu[tetZeilen-1] = reihe
		z = neu
	}
	return z, false
}

const (
	tetBotDenkzeit  = 250 * time.Millisecond // vier Steine je Sekunde
	tetBotNachholen = 8                      // höchstens so viele Steine je Abruf
)

func tetBotVerliert(g *TetrisGame) {
	g.BlueAlive = false
	g.Status = "finished"
	g.Winner = "red"
	g.Note = g.BlueName + " hat oben keinen Platz mehr."
}

// tetBotRunde holt offenen Müll ab und setzt einen Stein. false = vorbei.
func tetBotRunde(g *TetrisGame) bool {
	if g.BlueGarbage > 0 {
		feld, tot := tetFeldMuell(g.BotFeld.Z, g.BlueGarbage, botZufall().Intn(tetSpalten))
		g.BotFeld.Z = feld
		g.BlueGarbage = 0
		if tot || !tetBotPasst(g.BotFeld) {
			tetBotVerliert(g)
			return false
		}
	}
	reihen, vorbei := tetBotZug(g.BotFeld, g.BotStufe)
	g.BlueLines += reihen
	if muell := tetMuellFuerReihen(reihen); muell > 0 {
		g.RedGarbage = tetrisBegrenzen(g.RedGarbage+muell, 0, 999)
	}
	if vorbei {
		tetBotVerliert(g)
		return false
	}
	return true
}

// tetBotTreiben setzt die Steine, die seit dem letzten Abruf fällig waren.
// Angetrieben wird es vom GET der Spielseite (alle 500 ms), deshalb holt der
// Bot je Abruf nach, statt nur einen Stein zu setzen.
func tetBotTreiben(g *TetrisGame, jetzt time.Time) {
	if g == nil || g.Status != "active" || g.BotID == "" {
		return
	}
	if g.BotFeld == nil {
		g.BotFeld = tetFeldNeu(g.Seed)
	}
	if g.BotFrist.IsZero() {
		g.BotFrist = jetzt.Add(tetBotDenkzeit)
		return
	}
	for n := 0; n < tetBotNachholen && !jetzt.Before(g.BotFrist); n++ {
		if !tetBotRunde(g) {
			return
		}
		g.BotFrist = g.BotFrist.Add(tetBotDenkzeit)
	}
	if jetzt.After(g.BotFrist) {
		// Weit im Rückstand (Tab war lange zu): nicht unbegrenzt nachholen.
		g.BotFrist = jetzt.Add(tetBotDenkzeit)
	}
}
