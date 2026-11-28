package main

/* ===========================================================================
   BOT FÜR 17 UND 4

   Am Tisch sitzt der Mensch als Pointeur, dazu zwei Bots: einer ist die Bank,
   einer spielt wie der Mensch gegen sie. Jede Bot-Aktion läuft durch
   siebzehnAktion – dieselbe Prüfung wie ein menschlicher Klick. Der Bot sieht
   nur, was offen auf dem Tisch liegt (Bank, offene Bankkarte, eigene Karten).

   Stufen:
     leicht  – legt die kleinste Bank, setzt zufällig hoch (bis 1 €) und zieht
               fast immer weiter, auch bei 17.
     mittel  – legt ein Zehntel der Kasse als Bank, setzt 10–40 Cent und zieht
               bis 16, hält ab 17.
     schwer  – legt ein Viertel der Kasse, setzt nur den kleinsten Betrag (die
               Bank hat den Vorteil) und zieht bei 17 nur gegen eine hohe
               offene Bankkarte weiter, sonst hält er.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// siebzehnBotIstBot sagt, ob diese Kennung ein Bot ist.
func siebzehnBotIstBot(g *SiebzehnGame, id string) bool {
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

// siebzehnBotAuftrag sagt, was gerade zu tun ist: Sitz, Aktion, Betrag.
// ok=false heißt, der Mensch ist dran oder die Partie wartet auf ihn.
func siebzehnBotAuftrag(g *SiebzehnGame) (int, string, int, bool) {
	if g == nil || len(g.Bots) == 0 || g.BankSitz < 0 || g.BankSitz >= len(g.Spieler) {
		return 0, "", 0, false
	}
	bankBot := siebzehnBotIstBot(g, g.Spieler[g.BankSitz].ID)

	switch g.Status {
	case "bank":
		if bankBot {
			return g.BankSitz, "bank", siebzehnBotBankBetrag(g, g.BankSitz), true
		}

	case "setzen":
		// Erst setzen die Pointeure, dann gibt die Bank.
		for i := range g.Spieler {
			if i == g.BankSitz || !siebzehnBotIstBot(g, g.Spieler[i].ID) {
				continue
			}
			p := &g.Spieler[i]
			if p.Status == "pleite" || p.Einsatz > 0 || p.Kasse < siebzehnSchritt {
				continue
			}
			if g.Bank-siebzehnBelegt(g) < siebzehnSchritt {
				continue
			}
			return i, "setzen", siebzehnBotEinsatz(g, i), true
		}
		if bankBot && siebzehnAlleGesetzt(g) {
			return g.BankSitz, "geben", 0, true
		}

	case "spiel":
		if g.Zug >= 0 && g.Zug < len(g.Spieler) && siebzehnBotIstBot(g, g.Spieler[g.Zug].ID) {
			if siebzehnBotZiehen(g, g.Zug) {
				return g.Zug, "ziehen", 0, true
			}
			return g.Zug, "halten", 0, true
		}

	case "runde":
		// Nach der Abrechnung wartet der Bot – auch als Bank. Die nächste
		// Runde beginnt erst, wenn ein Mensch „Nächste Runde“ drückt.
		return 0, "", 0, false
	}
	return 0, "", 0, false
}

// siebzehnAlleGesetzt: Alle, die setzen können, haben gesetzt – oder die Bank
// ist voll abgedeckt, dann kann niemand mehr.
func siebzehnAlleGesetzt(g *SiebzehnGame) bool {
	if siebzehnBelegt(g) <= 0 {
		return false
	}
	if g.Bank-siebzehnBelegt(g) < siebzehnSchritt {
		return true
	}
	for i := range g.Spieler {
		if i == g.BankSitz {
			continue
		}
		if g.Spieler[i].Status == "pleite" || g.Spieler[i].Kasse < siebzehnSchritt {
			continue
		}
		if g.Spieler[i].Einsatz <= 0 {
			return false
		}
	}
	return true
}

// siebzehnBotBankBetrag wählt die Bank: klein, mittel oder groß – je nach Stufe
// und Kasse, immer in 10-Cent-Schritten.
func siebzehnBotBankBetrag(g *SiebzehnGame, sitz int) int {
	kasse := (g.Spieler[sitz].Kasse / siebzehnSchritt) * siebzehnSchritt
	betrag := siebzehnBankMin
	switch botStufe(g.BotStufe) {
	case botLeicht:
		betrag = siebzehnBankMin
	case botSchwer:
		betrag = kasse / 4
	default:
		betrag = kasse / 10
	}
	betrag = (betrag / siebzehnSchritt) * siebzehnSchritt
	if betrag < siebzehnBankMin {
		betrag = siebzehnBankMin
	}
	if betrag > kasse {
		betrag = kasse
	}
	return betrag
}

// siebzehnBotEinsatz wählt den Einsatz gegen die Bank.
func siebzehnBotEinsatz(g *SiebzehnGame, sitz int) int {
	frei := g.Bank - siebzehnBelegt(g)
	max := g.Spieler[sitz].Kasse
	if frei < max {
		max = frei
	}
	max = (max / siebzehnSchritt) * siebzehnSchritt
	if max < siebzehnSchritt {
		return siebzehnSchritt
	}
	betrag := siebzehnSchritt
	spanne := func(grenze int) int {
		if grenze > max {
			grenze = max
		}
		schritte := (grenze - siebzehnSchritt) / siebzehnSchritt
		if schritte < 0 {
			schritte = 0
		}
		return siebzehnSchritt + rand.Intn(schritte+1)*siebzehnSchritt
	}
	switch botStufe(g.BotStufe) {
	case botLeicht:
		betrag = spanne(100)
	case botSchwer:
		// Die Bank gewinnt Gleichstände – kleine Einsätze halten den Verlust
		// klein.
		betrag = siebzehnSchritt
	default:
		betrag = spanne(40)
	}
	if betrag > max {
		betrag = max
	}
	if betrag < siebzehnSchritt {
		betrag = siebzehnSchritt
	}
	return betrag
}

// siebzehnBotZiehen sagt, ob der Bot eine Karte nimmt (sonst hält er).
func siebzehnBotZiehen(g *SiebzehnGame, sitz int) bool {
	summe := siebzehnSumme(g.Spieler[sitz].Karten)
	bankOffen := 0
	if len(g.BankKarten) > 0 {
		bankOffen = g.BankKarten[0]
	}
	switch botStufe(g.BotStufe) {
	case botLeicht:
		if summe >= 18 {
			return false
		}
		return rand.Intn(100) < 80
	case botSchwer:
		if summe >= 18 {
			return false
		}
		// Bei 17 nur weiter, wenn die Bank hoch zeigt und deshalb eher 18–21
		// erreicht: dagegen verliert oder teilt man, ein Gleichstand zählt für
		// die Bank.
		if summe == 17 {
			return bankOffen >= 10
		}
		return true
	}
	return summe <= 16
}

// siebzehnBotTreiben lässt den Bot nach der Denkpause handeln; aufgerufen im
// GET-Abruf, damit die Partie auch läuft, wenn der Mensch nur zusieht.
func siebzehnBotTreiben(g *SiebzehnGame, jetzt time.Time) bool {
	if _, _, _, ok := siebzehnBotAuftrag(g); !ok {
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
	return siebzehnBotSchritt(g)
}

// siebzehnBotSchritt führt genau eine Bot-Aktion aus – durch dieselbe Prüfung
// wie ein Mensch.
func siebzehnBotSchritt(g *SiebzehnGame) bool {
	sitz, aktion, betrag, ok := siebzehnBotAuftrag(g)
	if !ok {
		return false
	}
	_, meldung := siebzehnAktion(g, g.Spieler[sitz].ID, aktion, betrag)
	return meldung == ""
}
