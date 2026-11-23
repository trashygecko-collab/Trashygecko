package main

/* ===========================================================================
   COMPUTERGEGNER FÜR SNAKE

   Snake rechnet der Server selbst (snake.go: Takt, Wachsen, Anstoßen, Runden).
   Der Bot gibt deshalb nur eine Richtung an – über dieselbe Prüfung wie ein
   Mensch: SnakeSpieler.richtungMerken (höchstens drei auf Vorrat, kein
   Umkehren). Er kann nicht schneller ziehen als der Takt, weil seine Richtung
   genau vor demselben snakeSchritt gesetzt wird, den auch ein Mensch auslöst.

   Wegfindung: Breitensuche vom Kopf zum Apfel über freie Kacheln (eigener und
   fremder Körper blockieren, der Schwanz fällt beim Ziehen weg). Ist kein Weg
   frei, geht es in die Richtung mit dem meisten Platz (Flutfüllung). Schwer
   meidet zusätzlich die Kacheln, auf die der gegnerische Kopf im nächsten
   Schritt ziehen könnte; leicht würfelt zu 30 % eine freie Richtung.
   =========================================================================== */

// snaRichtungen sind die vier Züge in fester Reihenfolge.
var snaRichtungen = []SnaPunkt{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

func snaImFeld(p SnaPunkt) bool {
	return p.X >= 0 && p.Y >= 0 && p.X < snaSpalten && p.Y < snaZeilen
}

// snaBelegt baut das Feld der blockierten Kacheln. Der Schwanz einer lebenden
// Schlange fällt weg, weil er in diesem Schritt weiterzieht (außer sie wächst
// gerade – dann blockiert er, was nur vorsichtiger ist).
func snaBelegt(g *SnakeGame, ausser int) map[SnaPunkt]bool {
	belegt := map[SnaPunkt]bool{}
	for i := range g.Spieler {
		p := &g.Spieler[i]
		if !p.Leben {
			continue
		}
		teile := p.Segmente
		for k, teil := range teile {
			if k == len(teile)-1 {
				continue
			}
			belegt[teil] = true
		}
	}
	return belegt
}

// snaFlaeche zählt die von start aus erreichbaren freien Kacheln.
func snaFlaeche(belegt map[SnaPunkt]bool, start SnaPunkt) int {
	if !snaImFeld(start) || belegt[start] {
		return 0
	}
	gesehen := map[SnaPunkt]bool{start: true}
	stapel := []SnaPunkt{start}
	for len(stapel) > 0 {
		p := stapel[len(stapel)-1]
		stapel = stapel[:len(stapel)-1]
		for _, d := range snaRichtungen {
			n := SnaPunkt{p.X + d.X, p.Y + d.Y}
			if !snaImFeld(n) || belegt[n] || gesehen[n] {
				continue
			}
			gesehen[n] = true
			stapel = append(stapel, n)
		}
	}
	return len(gesehen)
}

// snaWegSuch sucht den ersten Schritt vom Kopf zum Apfel. gefahr sind Kacheln,
// die der Gegner im nächsten Zug erreichen könnte (nur schwer).
func snaWegSuch(kopf, ziel SnaPunkt, belegt, gefahr map[SnaPunkt]bool) (SnaPunkt, bool) {
	if !snaImFeld(ziel) {
		return SnaPunkt{}, false
	}
	eltern := map[SnaPunkt]SnaPunkt{kopf: {-1, -1}}
	stapel := []SnaPunkt{kopf}
	for kopfIndex := 0; kopfIndex < len(stapel); kopfIndex++ {
		p := stapel[kopfIndex]
		for _, d := range snaRichtungen {
			n := SnaPunkt{p.X + d.X, p.Y + d.Y}
			if !snaImFeld(n) || belegt[n] {
				continue
			}
			if _, da := eltern[n]; da {
				continue
			}
			if gefahr[n] && n != ziel {
				continue
			}
			eltern[n] = p
			if n == ziel {
				// Weg zurück bis zum ersten Schritt nach dem Kopf.
				schritt := n
				for {
					vorgaenger := eltern[schritt]
					if vorgaenger == kopf {
						return SnaPunkt{schritt.X - kopf.X, schritt.Y - kopf.Y}, true
					}
					schritt = vorgaenger
				}
			}
			stapel = append(stapel, n)
		}
	}
	return SnaPunkt{}, false
}

// snakeBotWahl bestimmt die nächste Richtung des Bots auf Platz sitz.
func snakeBotWahl(g *SnakeGame, sitz int, stufe string) SnaPunkt {
	if sitz < 0 || sitz >= len(g.Spieler) {
		return SnaPunkt{1, 0}
	}
	s := &g.Spieler[sitz]
	if len(s.Segmente) == 0 {
		return s.Richtung
	}
	kopf := s.Segmente[0]
	belegt := snaBelegt(g, sitz)

	// Kacheln, auf die der Gegner im nächsten Schritt ziehen könnte.
	gefahr := map[SnaPunkt]bool{}
	if stufe == botSchwer && len(g.Spieler) == 2 {
		for i := range g.Spieler {
			if i == sitz || !g.Spieler[i].Leben || len(g.Spieler[i].Segmente) == 0 {
				continue
			}
			gk := g.Spieler[i].Segmente[0]
			for _, d := range snaRichtungen {
				gefahr[SnaPunkt{gk.X + d.X, gk.Y + d.Y}] = true
			}
		}
	}

	// 1. Weg zum Apfel.
	if g.HatApfel {
		if d, ok := snaWegSuch(kopf, g.Apfel, belegt, gefahr); ok {
			if stufe != botLeicht || botZufall().Intn(100) >= 30 {
				return d
			}
		}
	}

	// 2. Kein Weg: die Richtung mit dem meisten Platz, bei Gleichstand die
	//    nähere zum Apfel.
	bestes := s.Richtung
	besterPlatz, besterAbstand := -1, 1<<30
	for _, d := range snaRichtungen {
		n := SnaPunkt{kopf.X + d.X, kopf.Y + d.Y}
		if !snaImFeld(n) || belegt[n] {
			continue
		}
		platz := snaFlaeche(belegt, n)
		abstand := 0
		if g.HatApfel {
			abstand = snaBetrag(n.X-g.Apfel.X) + snaBetrag(n.Y-g.Apfel.Y)
		}
		if platz > besterPlatz || (platz == besterPlatz && abstand < besterAbstand) {
			bestes, besterPlatz, besterAbstand = d, platz, abstand
		}
	}
	if stufe == botLeicht && botZufall().Intn(100) < 30 {
		// Irgendeine freie Richtung – auch mal in die Sackgasse.
		var frei []SnaPunkt
		for _, d := range snaRichtungen {
			n := SnaPunkt{kopf.X + d.X, kopf.Y + d.Y}
			if snaImFeld(n) && !belegt[n] {
				frei = append(frei, d)
			}
		}
		if len(frei) > 0 {
			return frei[botZufall().Intn(len(frei))]
		}
	}
	return bestes
}

// snakeBotRichtung merkt die Richtung des Bots für den nächsten Schritt vor.
func snakeBotRichtung(g *SnakeGame, stufe string) {
	if g == nil || g.BotID == "" {
		return
	}
	sitz := snakeSitz(g, g.BotID)
	if sitz < 0 || !g.Spieler[sitz].Leben {
		return
	}
	g.Spieler[sitz].richtungMerken(snakeBotWahl(g, sitz, stufe))
}

func snaBetrag(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
