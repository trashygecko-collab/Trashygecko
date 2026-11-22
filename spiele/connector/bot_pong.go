package main

/* ===========================================================================
   BOT FÜR PONG

   Pong rechnet der Server selbst; der Bot ist der zweite Schläger und bewegt
   ihn im selben Takt wie die Physik (pongWeiter). Er sieht nur Ball und
   Schläger – also genau das, was auch ein Mensch auf dem Schirm sieht – und
   trifft je nach Stufe genauer und schneller.

   Stufen:
     leicht  – langsamer Schläger, zielt mit großem Fehler.
     mittel  – mittleres Tempo, kleiner Fehler.
     schwer  – schnell und fast ohne Fehler; rechnet den Ballweg bis zum
               Schläger mit Seitenwänden voraus.
   =========================================================================== */

import (
	"math"
	"math/rand"
)

// pongBotSitz ist der Platz des Bots (-1, wenn keiner mitspielt).
func pongBotSitz(g *PongGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return pongSitz(g, g.BotID)
}

// pongBotTempo ist die Höchstgeschwindigkeit des Schlägers in Pixel je Sekunde.
func pongBotTempo(stufe string) float64 {
	switch botStufe(stufe) {
	case botLeicht:
		return 250
	case botSchwer:
		return 480
	}
	return 350
}

// pongBotFehler ist der größte Zielfehler je Ballwechsel.
func pongBotFehler(stufe string) float64 {
	switch botStufe(stufe) {
	case botLeicht:
		return 55
	case botSchwer:
		return 8
	}
	return 26
}

// pongSpiegelX faltet eine x-Position an den Seitenwänden, so wie der Ball
// abprallt.
func pongSpiegelX(x, min, max float64) float64 {
	spanne := max - min
	if spanne <= 0 {
		return min
	}
	x = math.Mod(x-min, 2*spanne)
	if x < 0 {
		x += 2 * spanne
	}
	if x > spanne {
		x = 2*spanne - x
	}
	return min + x
}

// pongBotZiel rechnet aus, wo der Ball die Schlägerlinie erreicht. Fliegt er
// vom Bot weg, kehrt der Schläger in die Mitte zurück.
func pongBotZiel(g *PongGame, sitz int) float64 {
	mitte := pongBreite / 2
	if sitz == 0 {
		if g.BallVY <= 0 {
			return mitte
		}
	} else if g.BallVY >= 0 {
		return mitte
	}
	ebene := pongSchlaegerY(sitz)
	if sitz == 0 {
		ebene += pongSchlaegerHoehe
	}
	if g.BallVY == 0 {
		return mitte
	}
	t := (ebene - g.BallY) / g.BallVY
	if t <= 0 {
		return mitte
	}
	return pongSpiegelX(g.BallX+g.BallVX*t, pongBallR, pongBreite-pongBallR)
}

// pongBotBewegen rückt den Bot-Schläger um dt Sekunden in Richtung Ziel.
func pongBotBewegen(g *PongGame, dt float64) {
	if g == nil || g.Status != "spiel" || dt <= 0 {
		return
	}
	sitz := pongBotSitz(g)
	if sitz < 0 {
		return
	}
	if g.Rallye != g.BotRallye {
		g.BotRallye = g.Rallye
		g.BotFehler = (rand.Float64()*2 - 1) * pongBotFehler(g.BotStufe)
	}
	ziel := pongBotZiel(g, sitz) + g.BotFehler
	x := g.Spieler[sitz].X
	schritt := pongBotTempo(g.BotStufe) * dt
	delta := ziel - x
	if delta > schritt {
		delta = schritt
	} else if delta < -schritt {
		delta = -schritt
	}
	x += delta
	halb := pongSchlaegerBreite / 2
	if x < halb {
		x = halb
	}
	if x > pongBreite-halb {
		x = pongBreite - halb
	}
	g.Spieler[sitz].X = x
}
