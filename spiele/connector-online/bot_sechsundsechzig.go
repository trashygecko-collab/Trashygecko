package main

/* ===========================================================================
   BOT FÜR SECHSUNDSECHZIG

   Am Tisch sitzt der Mensch auf Platz 0, der Bot auf Platz 1. Jede Bot-Aktion
   läuft durch sechsundsechzigAktion – dieselbe Prüfung wie ein menschlicher
   Klick. Der Bot sieht nur, was offen liegt (Trumpf, Talon, Tisch, eigene
   Hand), und spielt keine Karte, die das Regelwerk verbietet.

   Stufen:
     leicht  – spielt irgendeine erlaubte Karte, sagt nur zufällig an und
               schließt den Talon nie.
     mittel  – nimmt den Stich, wenn es billig geht, wirft sonst die
               wertloseste Karte ab und sagt König und Dame an.
     schwer  – wie mittel, zieht aber mit den beiden höchsten Trümpfen Trumpf
               an, spart Trümpfe bei billigen Stichen auf, kämpft um den
               letzten Stich (zehn Augen) und dreht den Talon nur mit
               Trumpf-Ass und -Zehn und ab 40 Augen zu.
  =========================================================================== */

import (
	"math/rand"
	"time"
)

// sechsundsechzigBotIstBot sagt, ob diese Kennung ein Bot ist.
func sechsundsechzigBotIstBot(g *SechsundsechzigGame, id string) bool {
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

// sechsundsechzigBotSitz liefert den Sitz des ersten Bots (-1, wenn keiner).
func sechsundsechzigBotSitz(g *SechsundsechzigGame) int {
	if g == nil || len(g.Bots) == 0 {
		return -1
	}
	return sechsundsechzigSitz(g, g.Bots[0])
}

// sechsundsechzigBotAuftrag sagt, was der Bot jetzt tun soll.
func sechsundsechzigBotAuftrag(g *SechsundsechzigGame) (string, int, int, bool) {
	sitz := sechsundsechzigBotSitz(g)
	if sitz < 0 {
		return "", 0, 0, false
	}
	switch g.Status {
	case "spiel":
		if g.Zug != sitz {
			return "", 0, 0, false
		}
		if g.Spieler[sitz].Augen+g.Spieler[sitz].Ansage >= 66 {
			return "melden", 0, 0, true
		}
		if len(g.Tisch) == 0 {
			if farbe, ok := sechsundsechzigBotAnsage(g, sitz); ok {
				return "heiraten", 0, farbe, true
			}
			if sechsundsechzigBotSperren(g, sitz) {
				return "sperren", 0, 0, true
			}
		}
		karte := sechsundsechzigBotKarte(g, sitz)
		if karte == 0 {
			return "", 0, 0, false
		}
		return "spielen", karte, 0, true

	case "runde":
		// Nach der Abrechnung wartet der Bot auf einen Menschen: Die Runde
		// beginnt erst auf „Nächste Runde“, damit das Ergebnis stehen bleibt.
		return "", 0, 0, false
	}
	return "", 0, 0, false
}

// sechsundsechzigBotTreiben lässt den Bot nach der Denkpause handeln.
func sechsundsechzigBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Sechsundsechzig
	if _, _, _, ok := sechsundsechzigBotAuftrag(g); !ok {
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
	return sechsundsechzigBotSchritt(room)
}

// sechsundsechzigBotSchritt führt genau eine Bot-Aktion aus.
func sechsundsechzigBotSchritt(room *Room) bool {
	g := room.Sechsundsechzig
	aktion, karte, farbe, ok := sechsundsechzigBotAuftrag(g)
	if !ok {
		return false
	}
	sitz := sechsundsechzigBotSitz(g)
	if _, meldung := sechsundsechzigAktion(room, g.Spieler[sitz].ID, aktion, karte, farbe); meldung != "" {
		return false
	}
	return true
}

// ---------------------------------------------------------------- Kartenwahl

// sechsundsechzigBotErlaubte sammelt alle Karten, die das Regelwerk zulässt.
func sechsundsechzigBotErlaubte(g *SechsundsechzigGame, sitz int) []int {
	karten := []int{}
	for _, k := range g.Spieler[sitz].Hand {
		if sechsundsechzigErlaubt(g, sitz, k) {
			karten = append(karten, k)
		}
	}
	return karten
}

// sechsundsechzigBotGewinnt sagt, ob karte den Stich gegen die ausgespielte
// Karte gewinnt (der Bot sitzt immer an zweiter Stelle im Stich).
func sechsundsechzigBotGewinnt(lead, karte, trump int) bool {
	leadTrumpf := sechsundsechzigFarbe(lead) == trump
	kartenTrumpf := sechsundsechzigFarbe(karte) == trump
	if kartenTrumpf && !leadTrumpf {
		return true
	}
	if kartenTrumpf && leadTrumpf {
		return sechsundsechzigWert(karte) > sechsundsechzigWert(lead)
	}
	if leadTrumpf {
		return false
	}
	return sechsundsechzigFarbe(karte) == sechsundsechzigFarbe(lead) &&
		sechsundsechzigWert(karte) > sechsundsechzigWert(lead)
}

// sechsundsechzigBotKosten bewertet eine Karte: Augen zählen am meisten, dann
// der Rang. Trümpfe bekommen einen Zuschlag, wenn sie geschont werden sollen.
func sechsundsechzigBotKosten(karte, trump, trumpZuschlag int) int {
	kosten := sechsundsechzigPunkte(karte)*10 + sechsundsechzigWert(karte)
	if sechsundsechzigFarbe(karte) == trump {
		kosten += trumpZuschlag
	}
	return kosten
}

// sechsundsechzigBotGuenstigste liefert die Karte mit den kleinsten Kosten.
func sechsundsechzigBotGuenstigste(karten []int, trump, trumpZuschlag int) int {
	beste := karten[0]
	for _, k := range karten[1:] {
		if sechsundsechzigBotKosten(k, trump, trumpZuschlag) < sechsundsechzigBotKosten(beste, trump, trumpZuschlag) {
			beste = k
		}
	}
	return beste
}

// sechsundsechzigBotHoechste liefert die höchste Karte, Trumpf zählt dabei als
// eigene Farbe (Trumpf schlägt Fehlfarbe).
func sechsundsechzigBotHoechste(karten []int, trump int) int {
	beste := karten[0]
	for _, k := range karten[1:] {
		besteTrumpf := sechsundsechzigFarbe(beste) == trump
		kTrumpf := sechsundsechzigFarbe(k) == trump
		if kTrumpf != besteTrumpf {
			if kTrumpf {
				beste = k
			}
			continue
		}
		if sechsundsechzigWert(k) > sechsundsechzigWert(beste) {
			beste = k
		}
	}
	return beste
}

// sechsundsechzigBotKarte wählt die Karte nach Stufe.
func sechsundsechzigBotKarte(g *SechsundsechzigGame, sitz int) int {
	erlaubt := sechsundsechzigBotErlaubte(g, sitz)
	if len(erlaubt) == 0 {
		return 0
	}
	stufe := botStufe(g.BotStufe)
	if stufe == botLeicht {
		return erlaubt[rand.Intn(len(erlaubt))]
	}
	trump := g.Trumpf
	if len(g.Tisch) == 0 {
		return sechsundsechzigBotAusspielen(g, sitz, erlaubt, stufe)
	}

	lead := g.Tisch[0].Karte
	gewinner := []int{}
	for _, k := range erlaubt {
		if sechsundsechzigBotGewinnt(lead, k, trump) {
			gewinner = append(gewinner, k)
		}
	}
	letzter := len(g.Spieler[0].Hand) <= 1 && len(g.Spieler[1].Hand) <= 1
	stichAugen := sechsundsechzigPunkte(lead)
	if len(gewinner) == 0 {
		// Nichts zu holen: die billigste Karte abwerfen, Trümpfe schonen.
		return sechsundsechzigBotGuenstigste(erlaubt, trump, 100)
	}
	if stufe == botSchwer && !letzter && stichAugen < 10 {
		// Ein billiger Stich lohnt keinen Trumpf: dann lieber abwerfen.
		ohneTrumpf := []int{}
		for _, k := range erlaubt {
			if sechsundsechzigFarbe(k) != trump {
				ohneTrumpf = append(ohneTrumpf, k)
			}
		}
		if len(ohneTrumpf) > 0 {
			return sechsundsechzigBotGuenstigste(ohneTrumpf, trump, 100)
		}
	}
	return sechsundsechzigBotGuenstigste(gewinner, trump, 0)
}

// sechsundsechzigBotAusspielen wählt die erste Karte eines Stichs.
func sechsundsechzigBotAusspielen(g *SechsundsechzigGame, sitz int, erlaubt []int, stufe string) int {
	trump := g.Trumpf
	if stufe != botSchwer {
		return sechsundsechzigBotHoechste(erlaubt, trump)
	}
	// Mit den beiden höchsten Trümpfen Trumpf anziehen, sonst die höchste
	// Fehlfarbe ausspielen und Trümpfe für später behalten.
	hatAss, hatZehn, trümpfe := false, false, 0
	for _, k := range g.Spieler[sitz].Hand {
		if sechsundsechzigFarbe(k) != trump {
			continue
		}
		trümpfe++
		if sechsundsechzigWert(k) == 14 {
			hatAss = true
		}
		if sechsundsechzigWert(k) == 10 {
			hatZehn = true
		}
	}
	fehl := []int{}
	for _, k := range erlaubt {
		if sechsundsechzigFarbe(k) != trump {
			fehl = append(fehl, k)
		}
	}
	if (hatAss && hatZehn) || trümpfe >= 3 {
		return sechsundsechzigBotHoechste(erlaubt, trump)
	}
	if len(fehl) > 0 {
		return sechsundsechzigBotHoechste(fehl, trump)
	}
	return sechsundsechzigBotHoechste(erlaubt, trump)
}

// ---------------------------------------------------------------- Ansagen

// sechsundsechzigBotAnsage wählt eine Hochzeit (König und Dame einer Farbe).
// Trumpf zählt 40, die anderen Farben 20.
func sechsundsechzigBotAnsage(g *SechsundsechzigGame, sitz int) (int, bool) {
	p := g.Spieler[sitz]
	kandidaten := []int{}
	for farbe := 0; farbe < 4; farbe++ {
		hatKoenig, hatDame := false, false
		for _, k := range p.Hand {
			if sechsundsechzigFarbe(k) != farbe {
				continue
			}
			if sechsundsechzigWert(k) == 13 {
				hatKoenig = true
			}
			if sechsundsechzigWert(k) == 12 {
				hatDame = true
			}
		}
		if hatKoenig && hatDame && !p.Angesagt[farbe] {
			kandidaten = append(kandidaten, farbe)
		}
	}
	if len(kandidaten) == 0 {
		return 0, false
	}
	if botStufe(g.BotStufe) == botLeicht {
		if rand.Intn(2) == 0 {
			return 0, false
		}
		return kandidaten[rand.Intn(len(kandidaten))], true
	}
	for _, farbe := range kandidaten {
		if farbe == g.Trumpf {
			return farbe, true
		}
	}
	return kandidaten[0], true
}

// sechsundsechzigBotSperren: Nur der schwere Bot dreht den Talon zu, und nur
// mit Trumpf-Ass und Trumpf-Zehn ab 40 Augen.
func sechsundsechzigBotSperren(g *SechsundsechzigGame, sitz int) bool {
	if botStufe(g.BotStufe) != botSchwer || g.Gesperrt || len(g.Talon) <= 1 {
		return false
	}
	if g.Spieler[sitz].Augen+g.Spieler[sitz].Ansage < 40 {
		return false
	}
	hatAss, hatZehn := false, false
	for _, k := range g.Spieler[sitz].Hand {
		if sechsundsechzigFarbe(k) != g.Trumpf {
			continue
		}
		if sechsundsechzigWert(k) == 14 {
			hatAss = true
		}
		if sechsundsechzigWert(k) == 10 {
			hatZehn = true
		}
	}
	return hatAss && hatZehn
}
