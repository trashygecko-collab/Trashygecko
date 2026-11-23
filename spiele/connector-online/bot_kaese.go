package main

import "time"

/* ===========================================================================
   BOT FÜR KÄSEKÄSTCHEN

   Der Bot kennt nur die Linien und Kästchen, die auch ein Mensch sieht.

   Stufen:
     leicht  – zieht zu 40 % irgendeine freie Linie, sonst nimmt er ein
               geschenktes Kästchen und meidet die offensichtlichste Vorlage.
     mittel  – ein Zug voraus: Kästchen schließen, keine dritte Seite
               verschenken.
     schwer  – zwei Züge voraus: rechnet die Kette des Gegners mit, wenn er
               eine Vorlage bekommt.
   =========================================================================== */

type kaeseBotLinie struct {
	dir      string
	row, col int
}

func kaeseBotFarbe(g *KaeseGame) string {
	if g == nil || g.BotID == "" {
		return ""
	}
	return kaeseColorOf(g, g.BotID)
}

// kaeseBotTreiben lässt den Bot ziehen, sobald die Denkpause vorbei ist. Der
// Aufrufer hält die Sperre des Raums, deshalb rechnet diese Funktion nichts:
// Die Suche dauert gemessen 2,6–11,5 ms je Zug und lief früher viermal je
// Abruf. Sie läuft in einer eigenen Goroutine auf einem Schnappschuss und legt
// die Linie in BotZug ab; der nächste Aufruf spielt sie aus.
func kaeseBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Kaese
	if g == nil || g.Status != "active" {
		return false
	}
	farbe := kaeseBotFarbe(g)
	if farbe == "" || g.Current != farbe {
		return false
	}
	if g.BotZug != nil {
		linie := *g.BotZug
		g.BotZug = nil
		if !kaeseBotZiehen(g, farbe, linie) {
			g.BotFrist = jetzt.Add(time.Second)
			return false
		}
		return true
	}
	if g.BotRechnet == g {
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
	g.BotRechnet = g
	g.BotGen++
	gen := g.BotGen
	schnappschuss := cloneKaese(g)
	stufe := g.BotStufe
	wache := botWacheStarten(botRechenFrist,
		func() bool { return room.Kaese == g && g.BotRechnet == g && g.BotGen == gen },
		func() {
			g.BotRechnet = nil
			g.BotFrist = time.Now()
		})
	go func() {
		linie, ok := kaeseBotWahl(schnappschuss, farbe, stufe)
		mu.Lock()
		defer mu.Unlock()
		wache.Stop()
		kaeseBotErgebnis(room, g, gen, linie, ok)
	}()
	return false
}

// kaeseBotErgebnis übernimmt eine fertig gerechnete Linie. quelle ist die
// Partie, für die gerechnet wurde: Wurde sie inzwischen ersetzt (neue Runde
// oder neuer Bot-Start), verfällt das Ergebnis, ohne die neue Partie
// anzufassen – insbesondere bleibt deren Rechen-Zustand stehen. Der Aufrufer
// hält die Sperre.
func kaeseBotErgebnis(room *Room, quelle *KaeseGame, gen int, linie kaeseBotLinie, ok bool) bool {
	if quelle == nil || room.Kaese != quelle || quelle.BotGen != gen {
		return false
	}
	quelle.BotRechnet = nil
	if !ok {
		return false
	}
	quelle.BotZug = &linie
	return true
}

func kaeseBotLinien(g *KaeseGame) []kaeseBotLinie {
	linien := []kaeseBotLinie{}
	if g == nil || g.Status != "active" {
		return linien
	}
	for row := 0; row <= g.Rows; row++ {
		for col := 0; col < g.Cols; col++ {
			if g.H[row*g.Cols+col] == "" {
				linien = append(linien, kaeseBotLinie{dir: "h", row: row, col: col})
			}
		}
	}
	for row := 0; row < g.Rows; row++ {
		for col := 0; col <= g.Cols; col++ {
			if g.V[row*(g.Cols+1)+col] == "" {
				linien = append(linien, kaeseBotLinie{dir: "v", row: row, col: col})
			}
		}
	}
	return linien
}

func kaeseBotWahl(g *KaeseGame, farbe, stufe string) (kaeseBotLinie, bool) {
	linien := kaeseBotLinien(g)
	if len(linien) == 0 {
		return kaeseBotLinie{}, false
	}
	stufe = botStufe(stufe)
	if stufe == botLeicht && botZufall().Intn(100) < 40 {
		return linien[botZufall().Intn(len(linien))], true
	}

	beste := linien[0]
	besterWert := -1 << 30
	for _, linie := range linien {
		probe := cloneKaese(g)
		geschlossen := kaeseBotAnwenden(probe, farbe, linie)
		wert := 0
		if stufe == botSchwer {
			wert = kaeseBotStellung(probe, farbe, 6)
		} else {
			wert = geschlossen * 100
			// Dritte Seiten sind Einladungen an den Gegner.
			wert -= kaeseBotOffene(probe, otherDameColor(farbe)) * 25
			wert += kaeseBotOffene(probe, farbe) * 4
		}
		wert += geschlossen * 1000
		if wert > besterWert || (wert == besterWert && botZufall().Intn(2) == 0) {
			beste, besterWert = linie, wert
		}
	}
	return beste, true
}

// kaeseBotAnwenden zieht die Linie und gibt die Zahl der geschlossenen
// Kästchen zurück. Zugwechsel und Spielende laufen wie im Handler.
func kaeseBotAnwenden(g *KaeseGame, farbe string, linie kaeseBotLinie) int {
	if g == nil || g.Status != "active" || g.Current != farbe {
		return 0
	}
	if linie.dir == "h" {
		if linie.row < 0 || linie.row > g.Rows || linie.col < 0 || linie.col >= g.Cols {
			return 0
		}
		if g.H[g.hIndex(linie.row, linie.col)] != "" {
			return 0
		}
	} else {
		if linie.row < 0 || linie.row >= g.Rows || linie.col < 0 || linie.col > g.Cols {
			return 0
		}
		if g.V[g.vIndex(linie.row, linie.col)] != "" {
			return 0
		}
	}
	geschlossen := kaesePlayLine(g, linie.dir, linie.row, linie.col)
	if geschlossen > 0 {
		g.Note = kaeseName(g, farbe) + " schließt " + kaesePlural(geschlossen) + "."
	} else {
		if g.Current == "red" {
			g.Current = "blue"
		} else {
			g.Current = "red"
		}
		g.Note = "Weiter geht's!"
	}
	if g.RedScore+g.BlueScore >= g.Rows*g.Cols {
		g.Status = "finished"
		switch {
		case g.RedScore > g.BlueScore:
			g.Winner = g.RedID
		case g.BlueScore > g.RedScore:
			g.Winner = g.BlueID
		default:
			g.Winner = "draw"
		}
	}
	return geschlossen
}

func kaeseBotZiehen(g *KaeseGame, farbe string, linie kaeseBotLinie) bool {
	if g == nil || g.Status != "active" || g.Current != farbe {
		return false
	}
	vorher := 0
	if linie.dir == "h" {
		vorher = len(g.H)
	} else {
		vorher = len(g.V)
	}
	_ = vorher
	kaeseBotAnwenden(g, farbe, linie)
	if linie.dir == "h" {
		return g.H[g.hIndex(linie.row, linie.col)] != ""
	}
	return g.V[g.vIndex(linie.row, linie.col)] != ""
}

func kaeseName(g *KaeseGame, farbe string) string {
	if farbe == "red" {
		return g.RedName
	}
	return g.BlueName
}

// kaeseBotOffene zählt Kästchen mit genau drei Seiten für eine Farbe – die
// kann der Gegner mit einem Zug schließen.
func kaeseBotOffene(g *KaeseGame, farbe string) int {
	offen := 0
	for row := 0; row < g.Rows; row++ {
		for col := 0; col < g.Cols; col++ {
			if g.boxDone(row, col) {
				continue
			}
			seiten := 0
			for _, v := range []string{
				g.H[g.hIndex(row, col)], g.H[g.hIndex(row+1, col)],
				g.V[g.vIndex(row, col)], g.V[g.vIndex(row, col+1)],
			} {
				if v == farbe {
					seiten++
				}
			}
			if seiten == 3 {
				offen++
			}
		}
	}
	return offen
}

// kaeseBotStellung bewertet, was nach dem Zug auf dem Brett steht. Der Gegner
// nimmt geschenkte Kästchen sofort – diese Kette wird mitgerechnet.
func kaeseBotStellung(g *KaeseGame, farbe string, tiefe int) int {
	if g.Status == "finished" {
		eigen, fremd := g.RedScore, g.BlueScore
		if farbe == "blue" {
			eigen, fremd = fremd, eigen
		}
		if eigen > fremd {
			return 100000
		}
		if eigen < fremd {
			return -100000
		}
		return 0
	}
	gegner := "blue"
	if farbe == "blue" {
		gegner = "red"
	}
	eigen, fremd := g.RedScore, g.BlueScore
	if farbe == "blue" {
		eigen, fremd = fremd, eigen
	}
	wert := (eigen - fremd) * 1000
	wert -= kaeseBotOffene(g, gegner) * 80
	wert += kaeseBotOffene(g, farbe) * 15
	if tiefe <= 0 {
		return wert
	}
	// Der Gegner schließt, was er schließen kann; jedes Kästchen gibt ihm
	// einen weiteren Zug.
	for i := 0; i < tiefe; i++ {
		gewinn := 0
		for _, linie := range kaeseBotLinien(g) {
			probe := cloneKaese(g)
			vorher := probe.RedScore + probe.BlueScore
			kaeseBotAnwenden(probe, g.Current, linie)
			if probe.RedScore+probe.BlueScore > vorher {
				gewinn++
			}
		}
		if gewinn == 0 {
			break
		}
		// Nur die Zahl der möglichen Kästchen zählt hier, nicht der Zug selbst.
		wert -= gewinn * 40
		break
	}
	return wert
}
