package main

/* ===========================================================================
   BOT FÜR DOMINO

   Domino ist ein Legespiel für zwei: je sieben Steine, der Rest im Vorrat, wer
   zuerst alle Steine weg hat, bekommt die Augen des Gegners als Punkte. Der
   Bot spielt über dieselben Funktionen wie ein Mensch (`dominoLegen`,
   `dominoZiehen`, `dominoPassen`) – er kann also nichts legen, was nicht
   passt, und zieht wie jeder andere, wenn nichts passt.

   Was er sieht, ist die eigene Hand und die offene Kette; die Hand des Gegners
   kennt er nicht. Er bewertet jeden erlaubten Zug danach, wie viele seiner
   eigenen Steine danach noch passen:

     · leicht  – legt irgendeinen passenden Stein (auch den letzten Anschluss weg)
     · mittel  – legt den Stein, nach dem die meisten eigenen Steine passen
     · schwer  – wie mittel, dazu hohe Steine zuerst (die zählen am Ende gegen
                 einen) und ein Doppelkopf, der dem Gegner wenig lässt

   Passt nichts, zieht er einen Stein aus dem Vorrat; ist der Vorrat leer,
   passt er. Damit kommt jede Runde voran.
   =========================================================================== */

import (
	"time"
)

// dominoBotZugWahl ist ein Zug: Stein und Seite.
type dominoBotZugWahl struct {
	stein [2]int
	seite string
}

// dominoBotEnden rechnet aus, welche zwei Werte nach dem Anlegen offen liegen.
func dominoBotEnden(g *DominoGame, stein [2]int, seite string) ([2]int, bool) {
	if len(g.Kette) == 0 {
		return [2]int{stein[0], stein[1]}, true
	}
	if seite == "links" {
		switch {
		case stein[1] == g.Enden[0]:
			return [2]int{stein[0], g.Enden[1]}, true
		case stein[0] == g.Enden[0]:
			return [2]int{stein[1], g.Enden[1]}, true
		}
		return [2]int{}, false
	}
	switch {
	case stein[0] == g.Enden[1]:
		return [2]int{g.Enden[0], stein[1]}, true
	case stein[1] == g.Enden[1]:
		return [2]int{g.Enden[0], stein[0]}, true
	}
	return [2]int{}, false
}

// dominoBotAnschluss zählt, wie viele eigene Steine nach diesem Zug noch
// passen – das ist die einzige Sicherheit, die ein Spieler hier hat.
func dominoBotAnschluss(g *DominoGame, sitz int, gespielt [2]int, enden [2]int) int {
	n := 0
	for _, k := range g.Spieler[sitz].Hand {
		if k == gespielt {
			continue
		}
		if k[0] == enden[0] || k[1] == enden[0] || k[0] == enden[1] || k[1] == enden[1] {
			n++
		}
	}
	return n
}

// dominoBotWahl sucht den besten erlaubten Zug der Stufe.
func dominoBotWahl(g *DominoGame, sitz int, stufe string) (dominoBotZugWahl, bool) {
	if g == nil || sitz < 0 || sitz >= len(g.Spieler) {
		return dominoBotZugWahl{}, false
	}
	leer := len(g.Kette) == 0
	kandidaten := make([]dominoBotZugWahl, 0, len(g.Spieler[sitz].Hand)*2)
	for _, stein := range g.Spieler[sitz].Hand {
		seiten := []string{"rechts"}
		if !leer {
			seiten = []string{"links", "rechts"}
		}
		for _, seite := range seiten {
			if _, ok := dominoBotEnden(g, stein, seite); ok {
				kandidaten = append(kandidaten, dominoBotZugWahl{stein: stein, seite: seite})
			}
		}
	}
	if len(kandidaten) == 0 {
		return dominoBotZugWahl{}, false
	}
	if botStufe(stufe) == botLeicht {
		return kandidaten[botZufall().Intn(len(kandidaten))], true
	}
	best, bestWert := kandidaten[0], -1<<30
	for _, k := range kandidaten {
		enden, _ := dominoBotEnden(g, k.stein, k.seite)
		wert := dominoBotAnschluss(g, sitz, k.stein, enden) * 10
		if botStufe(stufe) == botSchwer {
			wert += k.stein[0] + k.stein[1] // hohe Steine zuerst loswerden
			if enden[0] == enden[1] {
				wert += 4 // ein Doppelkopf lässt dem Gegner wenig Wege
			}
		}
		if wert > bestWert {
			best, bestWert = k, wert
		}
	}
	return best, true
}

// dominoBotTreiben lässt den Bot einen Schritt machen, sobald seine Denkpause
// um ist. Ein Zug je Abruf, damit die Mitspieler mitlesen können.
func dominoBotTreiben(g *DominoGame, jetzt time.Time) bool {
	if g == nil || g.Status != "spiel" || g.BotID == "" {
		return false
	}
	if !g.BotFrist.IsZero() && jetzt.Before(g.BotFrist) {
		return false
	}
	sitz := g.Zug
	if sitz < 0 || sitz >= len(g.Spieler) || g.Spieler[sitz].ID != g.BotID {
		return false
	}
	g.BotFrist = jetzt.Add(botWartezeit(g.BotStufe))
	if wahl, ok := dominoBotWahl(g, sitz, g.BotStufe); ok {
		return g.dominoLegen(sitz, wahl.stein, wahl.seite) == ""
	}
	if len(g.vorrat) > 0 {
		return g.dominoZiehen(sitz) == ""
	}
	return g.dominoPassen(sitz) == ""
}
