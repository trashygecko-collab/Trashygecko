package main

/* ===========================================================================
   BOT FÜR TRON

   Der Server rechnet die Rennen; der Bot lenkt seinen Renner im selben Takt
   wie die Physik (tronWeiter). Er sieht nur das Raster – also genau das, was
   auch ein Mensch auf dem Schirm sieht.

   Stufen:
     leicht  – wählt oft irgendeine freie Richtung und sperrt sich dadurch
               selbst ein.
     mittel  – fährt dorthin, wo der meiste freie Platz erreichbar ist.
     schwer  – wie mittel, meidet aber den Kopf des Gegners und bevorzugt
               Geradeausfahren sowie Felder mit vielen Fluchtwegen.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// tronBotSitz ist der Platz des Bots (-1, wenn keiner mitspielt).
func tronBotSitz(g *TronGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return tronSitz(g, g.BotID)
}

// tronBotFrei sagt, ob ein Feld im Raster liegt und noch frei ist.
func tronBotFrei(g *TronGame, x, y int) bool {
	if !tronFeldFrei(x, y) {
		return false
	}
	return g.Raster[y*tronSpalten+x] == 0
}

// tronBotFlaeche zählt die zusammenhängenden freien Felder ab (x, y), höchstens
// bis zur Grenze. Damit erkennt der Bot Sackgassen, bevor er hineinfährt.
func tronBotFlaeche(g *TronGame, x, y int, grenze int) int {
	if !tronBotFrei(g, x, y) {
		return 0
	}
	gesehen := make([]bool, len(g.Raster))
	start := y*tronSpalten + x
	gesehen[start] = true
	stapel := []int{start}
	anzahl := 0
	for len(stapel) > 0 && anzahl < grenze {
		i := stapel[len(stapel)-1]
		stapel = stapel[:len(stapel)-1]
		anzahl++
		cx, cy := i%tronSpalten, i/tronSpalten
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := cx+d[0], cy+d[1]
			if !tronFeldFrei(nx, ny) {
				continue
			}
			j := ny*tronSpalten + nx
			if gesehen[j] || g.Raster[j] != 0 {
				continue
			}
			gesehen[j] = true
			stapel = append(stapel, j)
		}
	}
	return anzahl
}

// tronBotWahl sucht die nächste Richtung: geradeaus, links oder rechts.
func tronBotWahl(g *TronGame, sitz int) (int, int) {
	s := &g.Spieler[sitz]
	kandidaten := [3][2]int{
		{s.DX, s.DY},
		{s.DY, -s.DX},
		{-s.DY, s.DX},
	}
	gegner := -1
	for i := range g.Spieler {
		if i != sitz && !g.Spieler[i].Tot {
			gegner = i
		}
	}
	gegnerX, gegnerY := -1, -1
	if gegner >= 0 {
		gegnerX = g.Spieler[gegner].X + g.Spieler[gegner].DX
		gegnerY = g.Spieler[gegner].Y + g.Spieler[gegner].DY
	}

	frei := []int{}
	wert := [3]int{}
	for i, d := range kandidaten {
		nx, ny := s.X+d[0], s.Y+d[1]
		if !tronBotFrei(g, nx, ny) {
			wert[i] = -1 << 30
			continue
		}
		frei = append(frei, i)
		wert[i] = tronBotFlaeche(g, nx, ny, 600)
		if i == 0 {
			wert[i] += 4 // geradeaus ist keine Umleitung
		}
		if nx == gegnerX && ny == gegnerY {
			wert[i] -= 5000 // Kopf an Kopf endet unentschieden
		}
		if botStufe(g.BotStufe) == botSchwer {
			// Fluchtwege nach dem Zug mitzählen.
			wege := 0
			for _, dd := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if tronBotFrei(g, nx+dd[0], ny+dd[1]) {
					wege++
				}
			}
			wert[i] += wege * 6
		}
	}
	if len(frei) == 0 {
		return s.DX, s.DY // nichts geht mehr – die Runde endet ohnehin
	}
	if botStufe(g.BotStufe) == botLeicht {
		if rand.Intn(2) == 0 {
			i := frei[rand.Intn(len(frei))]
			return kandidaten[i][0], kandidaten[i][1]
		}
	}
	best := frei[0]
	for _, i := range frei {
		if wert[i] > wert[best] {
			best = i
		}
	}
	return kandidaten[best][0], kandidaten[best][1]
}

// tronBotDenken setzt vor jedem Feld die Wunschrichtung des Bots.
func tronBotDenken(g *TronGame) {
	if g == nil || g.Status != "spiel" || g.BotID == "" {
		return
	}
	sitz := tronBotSitz(g)
	if sitz < 0 {
		return
	}
	s := &g.Spieler[sitz]
	if s.Tot {
		return
	}
	dx, dy := tronBotWahl(g, sitz)
	if dx == s.DX && dy == s.DY {
		return
	}
	tronLenken(s, dx, dy)
}

// tronBotTreiben startet zwischen den Runden die nächste, wenn der Mensch
// nicht selbst drückt.
func tronBotTreiben(g *TronGame, jetzt time.Time) bool {
	if g == nil || g.Status != "runde" || g.BotID == "" {
		return false
	}
	if g.BotFrist.IsZero() {
		g.BotFrist = jetzt.Add(2 * botWartezeit(g.BotStufe))
		return false
	}
	if jetzt.Before(g.BotFrist) {
		return false
	}
	g.BotFrist = time.Time{}
	tronNaechsteRunde(g)
	return true
}
