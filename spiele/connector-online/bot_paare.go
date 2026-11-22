package main

/* ===========================================================================
   BOT FÜR PAARE FINDEN

   Paare finden ist kein Brett, das beide gemeinsam sehen: jeder Spieler löst
   dieselben Zeichen in eigener Anordnung, und wer zuerst fertig ist, meldet
   Zeit und Fehlversuche. Der Server kennt das Brett gar nicht (es entsteht im
   Browser aus der Partie-Kennung).

   Deshalb löst der Bot ein eigenes Brett derselben Größe im Kopf und meldet
   erst, wenn seine gerechnete Zeit wirklich verstrichen ist (BotFrist). Er
   würfelt nicht einfach eine Zahl: er merkt sich aufgedeckte Karten, sucht
   bekannte Paare, deckt Unbekanntes auf und zählt dabei seine Fehlversuche
   wie ein Mensch – inklusive der 800 ms, die das Spiel nach einem Fehlgriff
   sperrt.

   Die Stufen unterscheiden sich im Gedächtnis (leicht merkt sich nur fünf
   Karten), im Tempo je Aufdecken und darin, wie oft der Bot sich vergreift.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// paareBotKarten ist die Zahl der Karten auf dem Brett (wie im Spiel:
// 25 Karten, davon so viele doppelt, wie eingestellt sind).
const paareBotKarten = 25

// paareBotSperre ist die Zeit, die das Spiel nach einem Fehlgriff sperrt.
const paareBotSperre = 800

// paareBotTempo liefert Millisekunden je Aufdecken, die Zahl der gemerkten
// Karten und die Fehlgriff-Quote je Stufe.
func paareBotTempo(stufe string) (ms, gedaechtnis, vergreifen int) {
	switch botStufe(stufe) {
	case botLeicht:
		return 1500, 5, 25
	case botSchwer:
		return 700, paareBotKarten, 0
	}
	return 1050, paareBotKarten, 6
}

// paareBotPlan rechnet eine ganze Runde im Kopf durch und liefert Zeit und
// Fehlversuche.
func paareBotPlan(g *PaareGame, stufe string) *PaareResult {
	paare := g.Pairs
	if paare < 1 {
		paare = 8
	}
	if paare > 12 {
		paare = 12
	}
	ms, gedaechtnis, vergreifen := paareBotTempo(stufe)

	// Brett bauen: jede Paarzahl doppelt, der Rest sind Einzelkarten.
	felder := make([]int, 0, paareBotKarten)
	for i := 0; i < paare; i++ {
		felder = append(felder, i, i)
	}
	for i := len(felder); i < paareBotKarten; i++ {
		felder = append(felder, i)
	}
	rnd := rand.New(rand.NewSource(int64(len(g.ID))*7919 + int64(paare)*131 + int64(len(stufe))*17))
	rnd.Shuffle(len(felder), func(i, j int) { felder[i], felder[j] = felder[j], felder[i] })

	gefunden := make([]bool, len(felder))
	erinnert := map[int]int{} // Feld → Wert
	reihenfolge := make([]int, 0, paareBotKarten)
	merke := func(feld int) {
		if _, da := erinnert[feld]; da {
			return
		}
		erinnert[feld] = felder[feld]
		reihenfolge = append(reihenfolge, feld)
		for len(reihenfolge) > gedaechtnis {
			alt := reihenfolge[0]
			reihenfolge = reihenfolge[1:]
			delete(erinnert, alt)
		}
	}

	// unbekannt liefert ein noch nicht aufgedecktes, nicht gefundenes Feld.
	unbekannt := func() int {
		frei := make([]int, 0, len(felder))
		for i := range felder {
			if gefunden[i] {
				continue
			}
			if _, da := erinnert[i]; da {
				continue
			}
			frei = append(frei, i)
		}
		if len(frei) == 0 {
			return -1
		}
		return frei[rnd.Intn(len(frei))]
	}

	// bekanntesPaar sucht zwei gemerkte Felder mit demselben Wert.
	bekanntesPaar := func() (int, int, bool) {
		for a, wa := range erinnert {
			if gefunden[a] {
				continue
			}
			for b, wb := range erinnert {
				if b == a || gefunden[b] || wb != wa {
					continue
				}
				return a, b, true
			}
		}
		return 0, 0, false
	}

	// bekannterPartner sucht ein gemerktes Feld mit demselben Wert.
	bekannterPartner := func(feld int) (int, bool) {
		for b, wb := range erinnert {
			if b == feld || gefunden[b] || wb != felder[feld] {
				continue
			}
			return b, true
		}
		return 0, false
	}

	flips, fehler, treffer := 0, 0, 0
	for treffer < paare && flips < 600 {
		// 1. Ein bekanntes Paar aus der Erinnerung – außer der Bot vergreift
		//    sich gerade.
		if a, b, ok := bekanntesPaar(); ok && rnd.Intn(100) >= vergreifen {
			flips += 2
			gefunden[a], gefunden[b] = true, true
			treffer++
			continue
		}

		// 2. Ein unbekanntes Feld aufdecken.
		a := unbekannt()
		if a < 0 {
			break
		}
		flips++
		merke(a)

		// 3. Ist der Partner schon bekannt? Dann gleich nachlegen.
		if b, ok := bekannterPartner(a); ok {
			flips++
			if felder[b] == felder[a] {
				gefunden[a], gefunden[b] = true, true
				treffer++
			} else {
				fehler++
				merke(b)
			}
			continue
		}

		// 4. Sonst ein zweites unbekanntes Feld.
		c := unbekannt()
		if c < 0 {
			break
		}
		flips++
		if felder[c] == felder[a] {
			gefunden[a], gefunden[c] = true, true
			treffer++
			continue
		}
		fehler++
		merke(c)
	}

	zeit := int64(flips*ms + fehler*paareBotSperre)
	if zeit < 1 {
		zeit = 1
	}
	return &PaareResult{Time: zeit, Mistakes: fehler}
}

// paareBotTreiben meldet das Ergebnis, sobald die gerechnete Zeit um ist
// (über GET gerufen).
func paareBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Paare
	if g == nil || len(g.Bots) == 0 || g.Status != "active" || g.BotPlan == nil {
		return false
	}
	if g.BlueDone != nil || jetzt.Before(g.BotFrist) {
		return false
	}
	return paareBotSchritt(room)
}

// paareBotSchritt meldet genau ein Ergebnis (dieselbe Prüfung wie für einen
// Menschen).
func paareBotSchritt(room *Room) bool {
	g := room.Paare
	_, meldung := paareAktion(room, g.Bots[0], "done", g.BotPlan.Time, g.BotPlan.Mistakes)
	return meldung == ""
}
