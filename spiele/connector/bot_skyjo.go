package main

/* ===========================================================================
   BOTS FÜR SKYJO

   Jeder Bot ist ein eigener Platz mit eigenen Karten und eigenen Zügen. Er
   handelt ausschließlich über skyjoAktion und sieht – wie ein Mensch – nur
   seine offenen Karten. Verdeckte Karten kennt er nicht.

   Stufen (die Entscheidungsqualität unterscheidet sich messbar):
     leicht – nimmt fast nie eine Ablagekarte, tauscht nur sehr niedrige
              Stapelkarten und deckt wahllos auf.
     mittel – nimmt eine Ablagekarte, wenn sie die höchste offene Karte
              schlägt, wirft hohe Stapelkarten ab.
     schwer – rechnet Spalten mit (drei gleiche zählen 0), wählt das
              günstigste Tauschziel und deckt die passende Spaltenkarte auf.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

func skyjoBotPause(stufe string) time.Duration {
	switch stufe {
	case "leicht":
		return 780 * time.Millisecond
	case "schwer":
		return 1150 * time.Millisecond
	}
	return 950 * time.Millisecond
}

// skyjoBotAmZug sagt, ob der Bot jetzt selbst handeln muss.
func skyjoBotAmZug(g *SkyjoGame) bool {
	if g == nil || g.Status != "active" {
		return false
	}
	if g.Phase != "setup" && g.Phase != "play" && g.Phase != "reveal" {
		return false
	}
	return g.Zug >= 0 && g.Zug < len(g.Spieler) && g.Spieler[g.Zug].Bot
}

// skyjoBotTreiben lässt den aktiven Bot nach seiner Denkpause genau einen
// Schritt machen. Aufgerufen beim Abholen des Stands.
func skyjoBotTreiben(g *SkyjoGame, jetzt time.Time) bool {
	if !skyjoBotAmZug(g) {
		if g != nil {
			g.BotFrist = time.Time{}
		}
		return false
	}
	if !g.BotFrist.IsZero() && jetzt.Before(g.BotFrist) {
		return false
	}
	p := &g.Spieler[g.Zug]
	if skyjoBotSchritt(g) {
		g.BotFrist = jetzt.Add(skyjoBotPause(p.Stufe))
		return true
	}
	g.BotFrist = time.Time{}
	return false
}

// skyjoBotSchritt führt eine einzelne Bot-Aktion aus.
func skyjoBotSchritt(g *SkyjoGame) bool {
	if !skyjoBotAmZug(g) {
		return false
	}
	me := g.Zug
	p := &g.Spieler[me]
	stufe := p.Stufe
	if stufe == "" {
		stufe = "mittel"
	}
	switch g.Phase {
	case "setup":
		idx := skyjoBotVerdecktWahl(p, stufe)
		if idx < 0 {
			return false
		}
		code, _ := g.skyjoAktion(me, "reveal", idx, "")
		return code == 0
	case "play":
		if g.Gezogen == nil {
			quelle := "deck"
			if wert, ok := skyjoAblageWert(g); ok && skyjoBotNimmtAblage(p, wert, stufe) {
				quelle = "ablage"
			}
			code, _ := g.skyjoAktion(me, "draw", 0, quelle)
			return code == 0
		}
		gezogen := *g.Gezogen
		if g.Quelle == "deck" && skyjoBotWirftAb(p, gezogen, stufe) {
			if code, _ := g.skyjoAktion(me, "discard", 0, ""); code != 0 {
				return false
			}
			idx := skyjoBotVerdecktWahl(p, stufe)
			if idx < 0 {
				return true
			}
			code, _ := g.skyjoAktion(me, "opencard", idx, "")
			return code == 0
		}
		idx := skyjoBotTauschZiel(p, gezogen, stufe)
		if idx < 0 {
			// Kein guter Tausch: die Karte möglichst dorthin, wo sie am
			// wenigsten kostet (niedrigste offene Karte).
			idx = skyjoBotNiedrigsterPlatz(p)
		}
		if idx < 0 {
			return false
		}
		code, _ := g.skyjoAktion(me, "swap", idx, "")
		return code == 0
	case "reveal":
		idx := skyjoBotVerdecktWahl(p, stufe)
		if idx < 0 {
			return false
		}
		code, _ := g.skyjoAktion(me, "opencard", idx, "")
		return code == 0
	}
	return false
}

func skyjoAblageWert(g *SkyjoGame) (int, bool) {
	if g == nil || len(g.Ablage) == 0 {
		return 0, false
	}
	return g.Ablage[len(g.Ablage)-1], true
}

func skyjoHoechsteOffen(p *SkyjoSpieler) (int, int) {
	wert, idx := -1<<30, -1
	for i, k := range p.Karten {
		if k.Weg || !k.Offen {
			continue
		}
		if k.Wert > wert {
			wert, idx = k.Wert, i
		}
	}
	return wert, idx
}

// skyjoSpalteFuer liefert die Spaltenmitglieder einer Karte.
func skyjoSpalteFuer(idx int) []int {
	s := idx % skyjoSpalten
	return []int{s, s + skyjoSpalten, s + skyjoSpalten*2}
}

// skyjoSpalteHilft sagt, ob ein Wert die Spalte vervollständigt: liegen dort
// schon zwei offene gleiche Karten, zählt die Spalte danach 0. Verdeckte
// Karten zählen nicht mit – der Bot kennt sie nicht.
func skyjoSpalteHilft(p *SkyjoSpieler, idx, neuer int) bool {
	gleich := 0
	for _, j := range skyjoSpalteFuer(idx) {
		if j == idx {
			continue
		}
		k := p.Karten[j]
		if k.Weg || !k.Offen {
			continue
		}
		if k.Wert == neuer {
			gleich++
		}
	}
	return gleich >= 2
}

func skyjoSpalteFertig(p *SkyjoSpieler, wert int) bool {
	for i, k := range p.Karten {
		if k.Weg || !k.Offen || k.Wert != wert {
			continue
		}
		if skyjoSpalteHilft(p, i, wert) {
			return true
		}
	}
	return false
}

func skyjoBotNimmtAblage(p *SkyjoSpieler, wert int, stufe string) bool {
	hoechste, _ := skyjoHoechsteOffen(p)
	if hoechste < 0 {
		return wert <= 3
	}
	switch stufe {
	case "leicht":
		return wert <= -1
	case "schwer":
		if skyjoSpalteFertig(p, wert) {
			return true
		}
		return wert < hoechste
	}
	return wert <= hoechste-3
}

func skyjoBotWirftAb(p *SkyjoSpieler, wert int, stufe string) bool {
	hoechste, _ := skyjoHoechsteOffen(p)
	if hoechste < 0 {
		return false
	}
	switch stufe {
	case "leicht":
		return wert > hoechste+3
	case "schwer":
		// Abwerfen ändert die eigene Summe nicht – es lohnt nur, wenn kein
		// Tausch die Hand verbessert.
		return wert >= hoechste
	}
	return wert >= 10 && wert > hoechste
}

func skyjoBotTauschZiel(p *SkyjoSpieler, wert int, stufe string) int {
	hoechste, ziel := skyjoHoechsteOffen(p)
	if hoechste < 0 {
		return -1
	}
	if stufe == "schwer" {
		for i, k := range p.Karten {
			if k.Weg || !k.Offen {
				continue
			}
			if skyjoSpalteHilft(p, i, wert) {
				return i
			}
		}
		if wert < hoechste {
			return ziel
		}
		return -1
	}
	if stufe == "leicht" {
		if hoechste >= 5 && wert <= hoechste-4 {
			return ziel
		}
		return -1
	}
	if wert < hoechste-2 {
		return ziel
	}
	return -1
}

// skyjoBotVerdecktWahl wählt eine verdeckte Karte; aufgedeckt wird, was
// darunter liegt – der Bot kann dabei nicht mehr wissen als ein Mensch.
func skyjoBotVerdecktWahl(p *SkyjoSpieler, stufe string) int {
	verdeckt := []int{}
	for i, k := range p.Karten {
		if !k.Offen && !k.Weg {
			verdeckt = append(verdeckt, i)
		}
	}
	if len(verdeckt) == 0 {
		return -1
	}
	return verdeckt[rand.Intn(len(verdeckt))]
}

// skyjoBotNiedrigsterPlatz liefert die offene Karte mit dem kleinsten Wert;
// dort richtet eine unpassende Karte den geringsten Schaden an.
func skyjoBotNiedrigsterPlatz(p *SkyjoSpieler) int {
	wert, idx := 1<<30, -1
	for i, k := range p.Karten {
		if k.Weg || !k.Offen {
			continue
		}
		if k.Wert < wert {
			wert, idx = k.Wert, i
		}
	}
	if idx >= 0 {
		return idx
	}
	return skyjoBotVerdecktWahl(p, "")
}
