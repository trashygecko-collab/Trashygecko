package main

/* ===========================================================================
   BOT FÜR POKER (TEXAS HOLD'EM)

   Am Tisch sitzt der Mensch auf Platz 0, dazu zwei Bots. Jede Bot-Aktion läuft
   durch pokerAktion – dieselbe Prüfung wie ein menschlicher Klick. Der Bot
   sieht nur seine eigenen Karten und was offen auf dem Tisch liegt.

   Stufen:
     leicht  – geht billig mit, passt sonst, erhöht nur mit sehr starker Hand.
     mittel  – rechnet die Pot-Quoten: mitgehen, wenn die Hand besser ist als
               der Preis, erhöhen bei sehr starker Hand.
     schwer  – wie mittel, setzt mit starker Hand aber den ganzen Pot und
               lässt schwache Hände gegen große Einsätze konsequent weg.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// pokerBotIstBot sagt, ob diese Kennung ein Bot ist.
func pokerBotIstBot(g *PokerGame, id string) bool {
	if g == nil || id == "" {
		return false
	}
	for _, b := range g.Bots {
		if b == id {
			return true
		}
	}
	return false
}

// pokerBotAuftrag sagt, was gerade zu tun ist.
func pokerBotAuftrag(g *PokerGame) (string, int, bool) {
	if g == nil || g.Status != "active" {
		return "", 0, false
	}
	if g.Runde == "showdown" {
		// Die Bots geben die nächste Hand frei.
		return "naechste", 0, true
	}
	sitz := g.AmZug
	if sitz < 0 || sitz >= len(g.Spieler) {
		return "", 0, false
	}
	p := &g.Spieler[sitz]
	if !pokerBotIstBot(g, p.ID) || p.Raus || p.AllIn {
		return "", 0, false
	}
	aktion, betrag := pokerBotEntscheidung(g, sitz)
	return aktion, betrag, true
}

// pokerBotTreiben lässt den Bot nach der Denkpause handeln.
func pokerBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Poker
	if _, _, ok := pokerBotAuftrag(g); !ok {
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
	return pokerBotSchritt(room)
}

// pokerBotSchritt führt genau eine Bot-Aktion aus.
func pokerBotSchritt(room *Room) bool {
	g := room.Poker
	aktion, betrag, ok := pokerBotAuftrag(g)
	if !ok {
		return false
	}
	if _, meldung := pokerAktion(room, g.Spieler[g.AmZug].ID, aktion, betrag); meldung != "" {
		return false
	}
	return true
}

// ---------------------------------------------------------------- Bewertung

// pokerBotStaerke schätzt die eigene Hand zwischen 0 (nichts) und 1 (Nuts).
func pokerBotStaerke(hand, tisch []string) float64 {
	if len(hand) < 2 {
		return 0
	}
	if len(tisch) < 3 {
		r1, r2 := pokerRang(hand[0]), pokerRang(hand[1])
		hoch, tief := r1, r2
		if tief > hoch {
			hoch, tief = tief, hoch
		}
		if r1 == r2 {
			// Ein Paar: Asse 0,95 – Zweien 0,48.
			return 0.45 + float64(hoch)/14*0.5
		}
		staerke := 0.12 + float64(hoch+tief)/28*0.36
		if hand[0][1] == hand[1][1] {
			staerke += 0.05 // gleiche Farbe
		}
		if hoch-tief <= 2 {
			staerke += 0.04 // zusammenhängend
		}
		return staerke
	}
	karten := append([]string{}, hand...)
	karten = append(karten, tisch...)
	best := pokerBest(karten)
	if best <= 0 {
		return 0.2
	}
	kategorie := best >> 20
	kicker := float64(best&0xFFFFF) / float64(1<<20)
	switch kategorie {
	case 0:
		return 0.18 + 0.12*kicker
	case 1:
		return 0.45 + 0.05*kicker
	case 2:
		return 0.6 + 0.05*kicker
	case 3:
		return 0.72 + 0.04*kicker
	case 4:
		return 0.78 + 0.04*kicker
	case 5:
		return 0.82 + 0.04*kicker
	case 6:
		return 0.88 + 0.03*kicker
	case 7:
		return 0.94 + 0.03*kicker
	}
	return 1
}

// pokerBotZiel rechnet den Betrag aus, auf den erhöht wird: erst mitgehen,
// dann Zuschlag, immer in 10-Cent-Schritten und nie über die eigenen Chips.
func pokerBotZiel(g *PokerGame, sitz, zuschlag int) int {
	p := &g.Spieler[sitz]
	ziel := g.Hoechster + zuschlag
	if ziel < g.Hoechster+pokerSchritt {
		ziel = g.Hoechster + pokerSchritt
	}
	ziel = (ziel / pokerSchritt) * pokerSchritt
	if ziel > p.Einsatz+p.Chips {
		ziel = p.Einsatz + p.Chips
	}
	return ziel
}

// pokerBotErhoehe packt eine Erhöhung; reicht das Geld nicht (Rest-Chips sind
// kein Vielfaches von 10 Cent), wird stattdessen mitgegangen.
func pokerBotErhoehe(g *PokerGame, sitz, zuschlag int) (string, int) {
	p := &g.Spieler[sitz]
	ziel := pokerBotZiel(g, sitz, zuschlag)
	if ziel <= p.Einsatz || ziel-p.Einsatz < pokerSchritt {
		if g.Hoechster > p.Einsatz {
			return "call", 0
		}
		return "check", 0
	}
	if ziel == p.Einsatz+p.Chips {
		// Die ganzen Chips sind selten ein Vielfaches von 10 Cent. Als
		// Erhöhung würde der Betrag abgerundet und läge dann unter dem
		// Mindestbetrag – der Tisch stünde. All-in geht ohne Rundung.
		return "allin", 0
	}
	return "raise", ziel
}

// pokerBotEntscheidung wählt Aktion und Betrag nach Stufe.
func pokerBotEntscheidung(g *PokerGame, sitz int) (string, int) {
	p := &g.Spieler[sitz]
	stufe := botStufe(g.BotStufe)
	staerke := pokerBotStaerke(p.Hand, g.Tisch)
	zu := g.Hoechster - p.Einsatz
	if zu < 0 {
		zu = 0
	}
	pot := g.Pot
	for i := range g.Spieler {
		pot += g.Spieler[i].Einsatz
	}
	wurf := rand.Float64()
	allIn := zu >= p.Chips

	switch stufe {
	case botLeicht:
		if zu == 0 {
			if staerke > 0.8 && wurf < 0.5 {
				return pokerBotErhoehe(g, sitz, pot/4)
			}
			return "check", 0
		}
		if allIn {
			if staerke > 0.75 {
				return "call", 0
			}
			return "fold", 0
		}
		if zu <= 40 || staerke > 0.7 || wurf < 0.15 {
			return "call", 0
		}
		return "fold", 0

	case botMittel:
		if zu == 0 {
			if staerke > 0.62 && wurf < 0.7 {
				return pokerBotErhoehe(g, sitz, pot/2)
			}
			return "check", 0
		}
		quote := float64(zu) / float64(pot+zu)
		if staerke >= 0.85 && wurf < 0.8 && !allIn {
			return pokerBotErhoehe(g, sitz, pot*3/4)
		}
		if staerke >= quote+0.18 {
			return "call", 0
		}
		return "fold", 0
	}

	// schwer: setzt mit starken Händen den ganzen Pot, geht mit, solange der
	// Preis zur Hand passt, und lässt schwache Hände gegen große Einsätze weg.
	if zu == 0 {
		if staerke > 0.66 {
			return pokerBotErhoehe(g, sitz, pot)
		}
		return "check", 0
	}
	quote := float64(zu) / float64(pot+zu)
	if allIn {
		if staerke > 0.78 {
			return "call", 0
		}
		return "fold", 0
	}
	if staerke >= 0.8 {
		return pokerBotErhoehe(g, sitz, pot)
	}
	if staerke >= quote+0.2 {
		return "call", 0
	}
	if zu <= 20 && staerke > 0.35 {
		return "call", 0
	}
	return "fold", 0
}
