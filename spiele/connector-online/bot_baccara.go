package main

/* ===========================================================================
   BOT FÜR BACCARA

   Am Tisch sitzt der Mensch auf Platz 0, der Bot auf Platz 1. Jede Bot-Aktion
   läuft durch baccaraAktion – dieselbe Prüfung wie ein menschlicher Klick.
   Der Bot legt zuerst die Bank, setzt als Ponte gegen eine fremde Bank, gibt
   die Karten erst, wenn alle gesetzt haben, und entscheidet bei fünf Punkten
   selbst.

   Stufen:
     leicht  – legt die kleinste Bank und setzt zufällig hoch (bis zur
               Abdeckung).
     mittel  – legt 15 % der Kasse und setzt 10–40 Cent.
     schwer  – legt ein Viertel der Kasse und setzt als Ponte immer nur den
               kleinsten Betrag, weil die Bank den Vorteil hat.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// baccaraBotIstBot sagt, ob diese Kennung ein Bot ist.
func baccaraBotIstBot(id string) bool {
	return id == botID(1)
}

// baccaraBotSitz liefert den Sitz des Bots (-1, wenn keiner).
func baccaraBotSitz(g *BaccaraGame) int {
	if g == nil || len(g.Bots) == 0 {
		return -1
	}
	return baccaraSitz(g, g.Bots[0])
}

// baccaraAlleGesetzt: Alle, die setzen können, haben gesetzt – oder die Bank
// ist voll abgedeckt, dann kann niemand mehr.
func baccaraAlleGesetzt(g *BaccaraGame) bool {
	if baccaraBelegt(g) <= 0 {
		return false
	}
	if g.Bank-baccaraBelegt(g) < baccaraSchritt {
		return true
	}
	for i := range g.Spieler {
		if i == g.BankSitz {
			continue
		}
		p := &g.Spieler[i]
		if p.Status == "pleite" || p.Kasse < baccaraSchritt {
			continue
		}
		if p.Einsatz <= 0 {
			return false
		}
	}
	return true
}

// baccaraBotAuftrag sagt, was der Bot jetzt tun soll.
func baccaraBotAuftrag(g *BaccaraGame) (string, int, bool) {
	sitz := baccaraBotSitz(g)
	if sitz < 0 {
		return "", 0, false
	}
	switch g.Status {
	case "bank":
		if g.BankSitz == sitz {
			if betrag := baccaraBotBank(g, sitz); betrag > 0 {
				return "bank", betrag, true
			}
		}

	case "setzen":
		if g.BankSitz == sitz {
			if baccaraAlleGesetzt(g) {
				return "geben", 0, true
			}
			return "", 0, false
		}
		p := &g.Spieler[sitz]
		if p.Status == "pleite" || p.Einsatz > 0 || p.Kasse < baccaraSchritt {
			return "", 0, false
		}
		if g.Bank-baccaraBelegt(g) < baccaraSchritt {
			return "", 0, false
		}
		if betrag := baccaraBotEinsatz(g, sitz); betrag > 0 {
			return "setzen", betrag, true
		}

	case "zug":
		if g.Zug == sitz {
			if baccaraBotZiehen(g.BotStufe) {
				return "karte", 0, true
			}
			return "stehen", 0, true
		}

	case "runde":
		// Der Bot gibt die nächste Runde frei (und reicht die Bank weiter,
		// wenn sie geschlagen wurde).
		return "naechste", 0, true
	}
	return "", 0, false
}

// baccaraBotTreiben lässt den Bot nach der Denkpause handeln.
func baccaraBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Baccara
	if _, _, ok := baccaraBotAuftrag(g); !ok {
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
	return baccaraBotSchritt(room)
}

// baccaraBotSchritt führt genau eine Bot-Aktion aus.
func baccaraBotSchritt(room *Room) bool {
	g := room.Baccara
	aktion, betrag, ok := baccaraBotAuftrag(g)
	if !ok {
		return false
	}
	sitz := baccaraBotSitz(g)
	if _, meldung := baccaraAktion(room, g.Spieler[sitz].ID, aktion, betrag); meldung != "" {
		return false
	}
	return true
}

// baccaraBotBank wählt den Betrag, den der Bot als Bank auf den Tisch legt.
func baccaraBotBank(g *BaccaraGame, sitz int) int {
	kasse := (g.Spieler[sitz].Kasse / baccaraSchritt) * baccaraSchritt
	betrag := baccaraBankMin
	switch botStufe(g.BotStufe) {
	case botMittel:
		betrag = kasse * 3 / 20
	case botSchwer:
		betrag = kasse / 4
	}
	if betrag < baccaraBankMin {
		betrag = baccaraBankMin
	}
	if betrag > kasse {
		betrag = kasse
	}
	return (betrag / baccaraSchritt) * baccaraSchritt
}

// baccaraBotEinsatz wählt den Einsatz als Ponte.
func baccaraBotEinsatz(g *BaccaraGame, sitz int) int {
	kasse := g.Spieler[sitz].Kasse
	frei := g.Bank - baccaraBelegt(g)
	hoechstens := kasse
	if frei < hoechstens {
		hoechstens = frei
	}
	hoechstens = (hoechstens / baccaraSchritt) * baccaraSchritt
	if hoechstens < baccaraSchritt {
		return 0
	}
	betrag := baccaraSchritt
	switch botStufe(g.BotStufe) {
	case botLeicht:
		betrag = (1 + rand.Intn(hoechstens/baccaraSchritt)) * baccaraSchritt
	case botMittel:
		betrag = (1 + rand.Intn(4)) * baccaraSchritt
	}
	if betrag > hoechstens {
		betrag = hoechstens
	}
	return betrag
}

// baccaraBotZiehen entscheidet bei fünf Punkten: Karte oder stehen bleiben.
// Leicht würfelt, mittel und schwer ziehen – fünf verliert zu oft gegen sechs
// und sieben, und nur die dritte Karte bringt noch sechs bis neun Punkte.
func baccaraBotZiehen(stufe string) bool {
	if botStufe(stufe) == botLeicht {
		return rand.Intn(2) == 0
	}
	return true
}
