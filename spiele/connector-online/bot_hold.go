package main

/* ===========================================================================
   BOT FÜR HOLD THE LINE

   5 × 5 Punkte. Wer am Zug ist, zieht eine gerade Linie zwischen zwei freien
   Punkten, die keine vorhandene Linie berührt oder kreuzt. Wer keinen Zug
   mehr hat, verliert – es gewinnt also, wer die letzte Linie zieht.

   Der Bot arbeitet mit einem eigenen Zufall, der an der Stellung hängt: keine
   Würfe aus dem globalen Zufall, damit Messläufe reproduzierbar bleiben und
   andere Tests im selben Lauf nicht gestört werden.

   Die Stufen:
     · leicht zieht irgendeine erlaubte Linie,
     · mittel nimmt eine der zehn Linien, die dem Gegner am wenigsten Züge
       lassen (Bedrängen),
     · schwer spielt die besten Kandidaten im Kopf durch: der Gegner antwortet,
       und der Bot prüft, ob er danach noch einen Zug hat. Ein Zug, nach dem
       der Gegner gar keinen mehr hat, gewinnt sofort.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// holdBotVorschauTiefe ist die Tiefe der Vorschau auf schwer (eigener Zug
// zählt nicht mit).
const holdBotVorschauTiefe = 2

// holdBotKontakt sagt, ob sich zwei Linien (a,b) und (c,d) berühren oder
// kreuzen – dieselbe Geometrie wie holdCanDraw, nur ohne Linienliste.
func holdBotKontakt(a, b, c, d int) bool {
	return holdSegmentsTouch(holdPointOf(a), holdPointOf(b), holdPointOf(c), holdPointOf(d))
}

// holdBotZuege liefert alle erlaubten Linien der Stellung.
func holdBotZuege(lines []HoldLine) [][2]int {
	zuege := make([][2]int, 0, 64)
	for a := 0; a < holdSize; a++ {
		for b := a + 1; b < holdSize; b++ {
			if holdCanDraw(lines, a, b) {
				zuege = append(zuege, [2]int{a, b})
			}
		}
	}
	return zuege
}

// holdBotRest ist die Zugliste, nachdem (a,b) gezogen wurde: der Zug selbst
// und alles, was ihn berührt, fällt weg.
func holdBotRest(zuege [][2]int, a, b int) [][2]int {
	rest := make([][2]int, 0, len(zuege))
	for _, z := range zuege {
		if z[0] == a && z[1] == b {
			continue
		}
		if holdBotKontakt(a, b, z[0], z[1]) {
			continue
		}
		rest = append(rest, z)
	}
	return rest
}

// holdBotZufall ist der Zufall des Bots. Er hängt an der Stellung, damit ein
// Messlauf reproduzierbar bleibt und der globale Zufall anderer Tests
// unberührt bleibt.
func holdBotZufall(g *HoldGame, zuege int) *rand.Rand {
	saat := int64(zuege)*7919 + int64(len(g.Lines))*131 + int64(len(g.BotStufe))*17 + 7
	for _, l := range g.Lines {
		saat = saat*31 + int64(l.A)*7 + int64(l.B)
	}
	return rand.New(rand.NewSource(saat))
}

// holdBotVorschau bewertet eine Stellung aus Sicht des Spielers am Zug:
// -1 heißt „keine Linie mehr, verloren“, +1 „der Gegner hat keine mehr“, 0
// „weiß nicht“. Sie prüft höchstens acht Antworten je Stellung, damit sie
// schnell bleibt.
func holdBotVorschau(zuege [][2]int, tiefe int, rnd *rand.Rand) int {
	if len(zuege) == 0 {
		return -1
	}
	if tiefe == 0 {
		return 0
	}
	kinder := append([][2]int(nil), zuege...)
	rnd.Shuffle(len(kinder), func(i, j int) { kinder[i], kinder[j] = kinder[j], kinder[i] })
	if len(kinder) > 8 {
		kinder = kinder[:8]
	}
	best := -2
	for _, m := range kinder {
		wert := -holdBotVorschau(holdBotRest(zuege, m[0], m[1]), tiefe-1, rnd)
		if wert > best {
			best = wert
		}
		if best == 1 {
			break
		}
	}
	return best
}

// holdBotZug entscheidet über die nächste Linie. ok ist false, wenn der Bot
// nicht am Zug ist oder keine Linie mehr möglich ist.
func holdBotZug(g *HoldGame, farbe string) (a, b int, ok bool) {
	if g == nil || g.Status != "active" || farbe == "" || g.Current != farbe {
		return 0, 0, false
	}
	zuege := holdBotZuege(g.Lines)
	if len(zuege) == 0 {
		return 0, 0, false
	}
	stufe := botStufe(g.BotStufe)
	rnd := holdBotZufall(g, len(zuege))

	// Leicht: irgendeine erlaubte Linie.
	if stufe == botLeicht {
		z := zuege[rnd.Intn(len(zuege))]
		return z[0], z[1], true
	}

	// Bedrängung: für jeden Kandidaten die Zahl der Gegenzüge, kleinste
	// zuerst. Ein Zug, nach dem der Gegner nichts mehr hat, gewinnt sofort.
	art := make([]int, len(zuege))
	for i, z := range zuege {
		art[i] = len(holdBotRest(zuege, z[0], z[1]))
	}
	reihe := make([]int, len(zuege))
	for i := range reihe {
		reihe[i] = i
	}
	for i := 0; i < len(reihe); i++ {
		for j := i + 1; j < len(reihe); j++ {
			if art[reihe[j]] < art[reihe[i]] {
				reihe[i], reihe[j] = reihe[j], reihe[i]
			}
		}
	}
	for _, i := range reihe {
		if art[i] == 0 {
			return zuege[i][0], zuege[i][1], true
		}
	}

	// Mittel: die drei besten Bedrängungslinien der Reihe nach prüfen und die
	// erste nehmen, nach der der Gegner den Bot nicht ohne Zug zurücklässt.
	// Ohne solche Prüfung rechnet der Bot nur „wenige Gegenzüge“ – das kürzt
	// die Partie und verliert gegen blosses Zufallsspiel (gemessen 0 von 40).
	if stufe == botMittel {
		grenze := 3
		if len(reihe) < grenze {
			grenze = len(reihe)
		}
		var ersatz [2]int
		for _, i := range reihe[:grenze] {
			z := zuege[i]
			rest := holdBotRest(zuege, z[0], z[1])
			if len(rest) == 0 {
				return z[0], z[1], true
			}
			antworten := append([][2]int(nil), rest...)
			rnd.Shuffle(len(antworten), func(m, n int) { antworten[m], antworten[n] = antworten[n], antworten[m] })
			if len(antworten) > 12 {
				antworten = antworten[:12]
			}
			ersatz = z
			ueberlebt := true
			for _, r := range antworten {
				if len(holdBotRest(rest, r[0], r[1])) == 0 {
					ueberlebt = false
					break
				}
			}
			if ueberlebt {
				return z[0], z[1], true
			}
		}
		return ersatz[0], ersatz[1], true
	}

	// Schwer: die besten vierundzwanzig durchrechnen.
	grenze := 24
	if len(reihe) < grenze {
		grenze = len(reihe)
	}
	var bestZug [2]int
	bestWert := -1 << 30
	for _, i := range reihe[:grenze] {
		z := zuege[i]
		rest := holdBotRest(zuege, z[0], z[1])
		if len(rest) == 0 {
			return z[0], z[1], true
		}
		wert := -holdBotVorschau(rest, holdBotVorschauTiefe, rnd)
		if wert > bestWert {
			bestWert, bestZug = wert, z
		}
	}
	return bestZug[0], bestZug[1], true
}

// holdBotTreiben lässt den Bot nach der Denkpause ziehen (über GET gerufen).
func holdBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Hold
	if g == nil || len(g.Bots) == 0 || g.Status != "active" {
		return false
	}
	if g.Current != holdColorOf(g, g.Bots[0]) {
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
	return holdBotSchritt(room)
}

// holdBotSchritt führt genau einen Zug aus (dieselbe Prüfung wie für einen
// Menschen).
func holdBotSchritt(room *Room) bool {
	g := room.Hold
	a, b, ok := holdBotZug(g, holdColorOf(g, g.Bots[0]))
	if !ok {
		return false
	}
	_, meldung := holdAktion(room, g.Bots[0], "line", a, b)
	return meldung == ""
}
