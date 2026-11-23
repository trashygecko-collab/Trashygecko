package main

/* ===========================================================================
   BOT FÜR BOMBERMAN

   Bomberman läuft in Echtzeit: der Server rechnet sechzig Schritte je Sekunde
   (bomberWeiter), die Geräte schicken nur Richtung und „Bombe“. Der Bot setzt
   also regelmäßig eine Richtung statt eines einzelnen Zuges. Er sieht dabei
   nur, was auch ein Mensch sieht – Karte, Bomben, Flammen, Extras, Gegner –
   und geht durch dieselbe Prüfung (bomberAktion).

   Was er tut:
     · brennt es auf seinem Feld, läuft er zum nächsten sicheren Feld,
     · steht er neben einer Kiste (oder hat den Gegner in Reichweite), legt er
       eine Bombe – aber nur, wenn danach noch ein Fluchtweg bleibt,
     · sonst läuft er über das Feld auf das nächste Extra oder den Gegner zu,
       Kisten im Weg werden weggesprengt.

   Die Stufen unterscheiden sich in der Voraussicht: leicht merkt die Gefahr
   erst, wenn es schon brennt, läuft oft ziellos und legt Bomben ohne
   Fluchtweg; mittel rechnet gut eine Sekunde voraus, schwer anderthalb.
   =========================================================================== */

import (
	"math"
	"math/rand"
	"time"
)

// bomberBotZufall liefert den eigenen Zufallsgenerator des Bots. Er hängt am
// Spielstand statt am globalen Zufall, damit Messläufe reproduzierbar bleiben
// und andere Zufallsspieler im selben Lauf nicht gestört werden.
func (g *BomberGame) bomberBotZufall() *rand.Rand {
	if g.BotZufall == nil {
		// Saat aus der Stufe: derselbe Ablauf ergibt dieselben Würfe. Damit
		// sind Messläufe reproduzierbar, ohne den globalen Zufall zu stören.
		saat := int64(7)
		for i := 0; i < len(g.BotStufe); i++ {
			saat = saat*31 + int64(g.BotStufe[i])
		}
		g.BotZufall = rand.New(rand.NewSource(saat))
	}
	return g.BotZufall
}

// bomberBotTakt ist der Abstand zwischen zwei Entscheidungen. Das Spiel läuft
// in Echtzeit, die Richtung gilt bis zur nächsten Entscheidung.
func bomberBotTakt(g *BomberGame, stufe string) time.Duration {
	base := 170 * time.Millisecond
	switch botStufe(stufe) {
	case botLeicht:
		base = 240 * time.Millisecond
	case botSchwer:
		base = 120 * time.Millisecond
	}
	rnd := g.bomberBotZufall()
	return base + time.Duration(rnd.Int63n(int64(60*time.Millisecond)))
}

// bomberBotLane sagt, ab wie vielen Sekunden Restzeit eine Schusslinie beim
// Laufen gemieden wird. Alle drei Stufen gehen nicht durch eine Bombe: die
// Flamme kommt schneller als jede Korrektur.
func bomberBotLane(stufe string) float64 {
	if botStufe(stufe) == botLeicht {
		// Auch leicht meidet die eigene Schusslinie – die Stufe liegt im
		// Tempo und im Zufall, nicht im Selbstmord.
		return bomberBombenzeit
	}
	return bomberBombenzeit
}

// bomberBotSpieler liefert den Bot (nil, wenn keiner mitspielt).
func bomberBotSpieler(g *BomberGame) *BomberSpieler {
	if g == nil || len(g.Bots) == 0 {
		return nil
	}
	sitz := bomberSitz(g, g.Bots[0])
	if sitz < 0 {
		return nil
	}
	return &g.Spieler[sitz]
}

// bomberBotGegner liefert den anderen Spieler.
func bomberBotGegner(g *BomberGame, ich *BomberSpieler) *BomberSpieler {
	for i := range g.Spieler {
		if g.Spieler[i].Nr != ich.Nr {
			return &g.Spieler[i]
		}
	}
	return nil
}

func bomberBotFeld(s *BomberSpieler) [2]int {
	return [2]int{int(math.Floor(s.X)), int(math.Floor(s.Y))}
}

// bomberBotGefahr sammelt je Feld die Sekunden bis zum Feuer. Steht eine
// Flamme schon, sind es 0.
func bomberBotGefahr(g *BomberGame) map[[2]int]float64 {
	return bomberBotGefahrMit(g, nil)
}

// bomberBotGefahrMit rechnet zusätzlich eine noch nicht gelegte Bombe ein.
// Gezählt wird in Zündreihenfolge: eine Bombe, die in eine fremde Flamme
// gerät, zündet früher; eine Kiste, die eine Bombe wegräumt, macht den Weg
// für eine später zündende Bombe frei. Genau diese beiden Fälle hat die alte
// Schätzung übersehen – der Bot hielt Felder für sicher, die es nicht waren.
func bomberBotGefahrMit(g *BomberGame, zusatz *BomberBombe) map[[2]int]float64 {
	gefahr := map[[2]int]float64{}
	setze := func(x, y int, rest float64) {
		feld := [2]int{x, y}
		if alt, ok := gefahr[feld]; !ok || rest < alt {
			gefahr[feld] = rest
		}
	}
	for _, f := range g.Flammen {
		setze(f.X, f.Y, 0)
	}

	type zünder struct {
		x, y, radius int
		rest         float64
		weg          bool
	}
	bomben := make([]zünder, 0, len(g.Bomben)+1)
	for _, b := range g.Bomben {
		rest := b.Rest
		if rest < 0 {
			rest = 0
		}
		bomben = append(bomben, zünder{b.X, b.Y, b.Radius, rest, false})
	}
	if zusatz != nil {
		bomben = append(bomben, zünder{zusatz.X, zusatz.Y, zusatz.Radius, zusatz.Rest, false})
	}
	// Kisten: Zeitpunkt, zu dem sie fallen.
	gefallen := map[[2]int]float64{}

	for {
		nächste := -1
		for i := range bomben {
			if bomben[i].weg {
				continue
			}
			if nächste < 0 || bomben[i].rest < bomben[nächste].rest {
				nächste = i
			}
		}
		if nächste < 0 {
			break
		}
		b := &bomben[nächste]
		b.weg = true
		t := b.rest
		setze(b.x, b.y, t)
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			for i := 1; i <= b.radius; i++ {
				x, y := b.x+d[0]*i, b.y+d[1]*i
				k := g.kachel(x, y)
				if k == 1 {
					break
				}
				if k == 2 {
					// Steht die Kiste schon (frühere Explosion)? Dann geht
					// die Flamme weiter. Sonst brennt sie hier und fällt.
					if fällt, ok := gefallen[[2]int{x, y}]; ok && fällt < t {
						setze(x, y, t)
						continue
					}
					setze(x, y, t)
					gefallen[[2]int{x, y}] = t
					break
				}
				setze(x, y, t)
				// Bombe in der Flamme: zündet jetzt, nicht später.
				for j := range bomben {
					if bomben[j].weg || bomben[j].x != x || bomben[j].y != y {
						continue
					}
					if bomben[j].rest > t {
						bomben[j].rest = t
					}
				}
			}
		}
	}
	return gefahr
}

// bomberBotBombeSteht sagt, ob auf dem Feld schon eine Bombe liegt.
func bomberBotBombeSteht(g *BomberGame, x, y int) bool {
	for _, b := range g.Bomben {
		if b.X == x && b.Y == y {
			return true
		}
	}
	return false
}

// bomberBotEigeneBomben zählt die Bomben des Spielers auf dem Feld.
func bomberBotEigeneBomben(g *BomberGame, s *BomberSpieler) int {
	n := 0
	for _, b := range g.Bomben {
		if b.Nr == s.Nr {
			n++
		}
	}
	return n
}

// bomberBotWegSuche sucht den kürzesten Weg zu einem Feld, das ziel erfüllt.
// Gemieden werden Wände, Bomben und alles, was sperre meldet (brennende
// Felder). Zurück kommt der Weg ohne das Startfeld; nil heißt „kein Weg“,
// ein leerer Weg heißt „steht schon auf dem Zielfeld“.
func bomberBotWegSuche(g *BomberGame, von [2]int, ziel func(x, y int) bool,
	sperre func(x, y int) bool) [][2]int {
	if ziel(von[0], von[1]) {
		return [][2]int{}
	}
	frei := func(x, y int) bool {
		if g.kachel(x, y) != 0 {
			return false
		}
		if bomberBotBombeSteht(g, x, y) && !(x == von[0] && y == von[1]) {
			return false
		}
		if sperre != nil && sperre(x, y) {
			return false
		}
		return true
	}

	vorher := map[[2]int][2]int{}
	gesehen := map[[2]int]bool{von: true}
	schlange := [][2]int{von}
	for len(schlange) > 0 {
		jetzt := schlange[0]
		schlange = schlange[1:]
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nach := [2]int{jetzt[0] + d[0], jetzt[1] + d[1]}
			if gesehen[nach] || !frei(nach[0], nach[1]) {
				continue
			}
			gesehen[nach] = true
			vorher[nach] = jetzt
			if ziel(nach[0], nach[1]) {
				weg := [][2]int{nach}
				for feld := nach; feld != von; {
					feld = vorher[feld]
					if feld == von {
						break
					}
					weg = append([][2]int{feld}, weg...)
				}
				return weg
			}
			schlange = append(schlange, nach)
		}
	}
	return nil
}

// bomberBotVorzeichen ist -1, 0 oder 1.
func bomberBotVorzeichen(v float64) int {
	if v > 0 {
		return 1
	}
	if v < 0 {
		return -1
	}
	return 0
}

// bomberBotKnoten sagt, ob der Bot sauber in einer Feldmitte steht. Nur dort
// darf er die Richtung wechseln – sonst pendelt er zwischen zwei Achsen hin
// und her, weil das Spiel ihn beim Laufen zur Mitte zieht.
func bomberBotKnoten(s *BomberSpieler) bool {
	return math.Abs(s.X-(math.Floor(s.X)+0.5)) < 0.12 &&
		math.Abs(s.Y-(math.Floor(s.Y)+0.5)) < 0.12
}

// bomberBotRichtung macht aus zwei Feldern eine Richtung.
func bomberBotRichtung(von, nach [2]int) (int, int) {
	dx, dy := nach[0]-von[0], nach[1]-von[1]
	if dx > 1 {
		dx = 1
	}
	if dx < -1 {
		dx = -1
	}
	if dy > 1 {
		dy = 1
	}
	if dy < -1 {
		dy = -1
	}
	return dx, dy
}

// bomberBotIrgendwohin läuft in eine freie Richtung (Notausgang).
func bomberBotIrgendwohin(g *BomberGame, von [2]int, gefahr map[[2]int]float64) (int, int, bool) {
	richtungen := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	rnd := g.bomberBotZufall()
	rnd.Shuffle(len(richtungen), func(i, j int) {
		richtungen[i], richtungen[j] = richtungen[j], richtungen[i]
	})
	bestes := [2]int{}
	bestwert := math.Inf(-1)
	for _, d := range richtungen {
		x, y := von[0]+d[0], von[1]+d[1]
		if g.kachel(x, y) != 0 || bomberBotBombeSteht(g, x, y) {
			continue
		}
		wert := 99.0
		if rest, ok := gefahr[[2]int{x, y}]; ok {
			wert = rest
		}
		if wert > bestwert {
			bestwert, bestes = wert, [2]int{d[0], d[1]}
		}
	}
	if bestwert == math.Inf(-1) {
		return 0, 0, false
	}
	return bestes[0], bestes[1], true
}

// bomberBotTrifftZiel sagt, ob eine Bombe auf diesem Feld etwas lohnt: eine
// Kiste oder der Gegner in Reichweite.
func bomberBotTrifftZiel(g *BomberGame, s *BomberSpieler, von [2]int) bool {
	gegner := bomberBotGegner(g, s)
	for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		for i := 1; i <= s.Feuer; i++ {
			x, y := von[0]+d[0]*i, von[1]+d[1]*i
			k := g.kachel(x, y)
			if k == 1 {
				break
			}
			if k == 2 {
				return true
			}
			if gegner != nil && gegner.Leben && int(math.Floor(gegner.X)) == x && int(math.Floor(gegner.Y)) == y {
				return true
			}
		}
	}
	return false
}

// bomberBotFlucht sucht nach dem Legen einer Bombe noch ein sicheres Feld.
// Gerechnet wird in Zeit, nicht in Schritten: der Bot muss die Mitte jedes
// Feldes vor dem Feuer erreichen, und die eigene Bombe zündet in 2,4 s.
// Kettenreaktionen und fallende Kisten stecken in bomberBotGefahrMit.
func bomberBotFlucht(g *BomberGame, s *BomberSpieler, von [2]int, gefahr map[[2]int]float64) bool {
	zusatz := &BomberBombe{X: von[0], Y: von[1], Nr: s.Nr, Rest: bomberBombenzeit, Radius: s.Feuer}
	mit := bomberBotGefahrMit(g, zusatz)

	kante := 1.0/(bomberTempo*s.Tempo) + 0.06
	best := map[[2]int]float64{von: 0}
	offen := [][2]int{von}
	for len(offen) > 0 {
		jetzt := offen[0]
		offen = offen[1:]
		t := best[jetzt]
		if _, brennt := mit[jetzt]; jetzt != von && !brennt {
			// Feld brennt nie: Ziel erreicht, wenn der Weg davor geklappt hat.
			return true
		}
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nach := [2]int{jetzt[0] + d[0], jetzt[1] + d[1]}
			if g.kachel(nach[0], nach[1]) != 0 {
				continue
			}
			if nach != von && bomberBotBombeSteht(g, nach[0], nach[1]) {
				continue
			}
			ankunft := t + kante
			if rest, brennt := mit[nach]; brennt && rest <= ankunft+0.25 {
				continue
			}
			if alt, ok := best[nach]; ok && alt <= ankunft {
				continue
			}
			best[nach] = ankunft
			offen = append(offen, nach)
		}
	}
	return false
}

// bomberBotZug entscheidet Richtung und Bombe. ok ist false, wenn der Bot
// gerade nichts tun kann (Vorlauf, tot, keine Runde).
func bomberBotZug(g *BomberGame) (int, int, bool, bool) {
	s := bomberBotSpieler(g)
	if s == nil || !s.Leben || g.Status != "spiel" || g.Vorlauf > 0 {
		return 0, 0, false, false
	}
	stufe := botStufe(g.BotStufe)
	meine := bomberBotFeld(s)
	gefahr := bomberBotGefahr(g)
	lane := bomberBotLane(stufe)
	brenntBald := func(x, y int) bool {
		rest, ok := gefahr[[2]int{x, y}]
		return ok && rest < lane
	}

	// 1. Hier brennt es gleich (eigene Bombe, fremde Bombe, Flamme): weg –
	// egal, wie viel Zeit bleibt. Wer in der Schusslinie stehen bleibt,
	// stirbt. Das gilt für jede Stufe.
	if rest, brennt := gefahr[meine]; brennt {
		budget := int(rest * bomberTempo * 0.8)
		if budget < 1 {
			budget = 1
		}
		weg := bomberBotWegSuche(g, meine, func(x, y int) bool {
			_, brennt := gefahr[[2]int{x, y}]
			return !brennt
		}, func(x, y int) bool {
			rest, brennt := gefahr[[2]int{x, y}]
			return brennt && rest <= 0 // steht schon in Flammen
		})
		if weg != nil && len(weg) > 0 && len(weg) <= budget {
			dx, dy := bomberBotRichtung(meine, weg[0])
			return dx, dy, false, true
		}
		dx, dy, ok := bomberBotIrgendwohin(g, meine, gefahr)
		return dx, dy, false, ok
	}

	// 2. Zwischen zwei Feldern gilt die Richtung, bis der nächste Knoten
	// erreicht ist (oder etwas im Weg steht – auch eine Bomben-Schusslinie:
	// sonst läuft der Bot am Rand der Flamme hin und her und wird erwischt).
	if !bomberBotKnoten(s) {
		if s.DX != 0 || s.DY != 0 {
			px, py := s.X+float64(s.DX)*0.3, s.Y+float64(s.DY)*0.3
			vorne := [2]int{int(math.Floor(px)), int(math.Floor(py))}
			if g.begehbar(px, py, s) && !brenntBald(vorne[0], vorne[1]) {
				return s.DX, s.DY, false, true
			}
		}
		if d := (math.Floor(s.X) + 0.5) - s.X; math.Abs(d) > 0.12 {
			return bomberBotVorzeichen(d), 0, false, true
		}
		if d := (math.Floor(s.Y) + 0.5) - s.Y; math.Abs(d) > 0.12 {
			return 0, bomberBotVorzeichen(d), false, true
		}
	}

	// 3. Leicht läuft gern ziellos durch die Gegend.
	if stufe == botLeicht && g.bomberBotZufall().Intn(100) < 25 {
		if dx, dy, ok := bomberBotIrgendwohin(g, meine, gefahr); ok {
			return dx, dy, false, true
		}
	}

	// 4. Bombe legen, wenn es etwas trifft und ein Fluchtweg bleibt.
	if bomberBotEigeneBomben(g, s) < s.BombMax && !bomberBotBombeSteht(g, meine[0], meine[1]) &&
		bomberBotTrifftZiel(g, s, meine) {
		// Auch die leichte Stufe legt keine Bombe ohne Fluchtweg. Sie irrt
		// weiterhin gern umher und reagiert spät – aber Selbstmord ist keine
		// Stufe, sondern ein Fehler.
		if bomberBotFlucht(g, s, meine, gefahr) {
			return 0, 0, true, true
		}
	}

	// Leicht folgt einem Plan nur in etwa der Hälfte der Fälle – es trödelt
	// und vertut sich, kommt aber voran.
	if stufe == botLeicht && g.bomberBotZufall().Intn(100) < 50 {
		if dx, dy, ok := bomberBotIrgendwohin(g, meine, gefahr); ok {
			return dx, dy, false, true
		}
	}

	var weg [][2]int
	{
		// 5. Sonst: Extra holen, sonst auf den Gegner zu, sonst eine Kiste
		// aufsprengen.
		ziel := func(x, y int) bool {
			for _, p := range g.Powerups {
				if p.X == x && p.Y == y {
					return true
				}
			}
			return false
		}
		weg = bomberBotWegSuche(g, meine, ziel, brenntBald)
		if weg == nil {
			gegner := bomberBotGegner(g, s)
			if gegner != nil && gegner.Leben {
				zielFeld := bomberBotFeld(gegner)
				weg = bomberBotWegSuche(g, meine, func(x, y int) bool {
					return x == zielFeld[0] && y == zielFeld[1]
				}, brenntBald)
			}
		}
		if weg == nil {
			// Kein Weg: an die nächste Kiste heran und sie wegräumen – aber nur an
			// einer Stelle, an der nach der Bombe noch ein Fluchtweg bleibt.
			// Sonst steht der Bot in einer Ecke und sprengt sich selbst.
			weg = bomberBotWegSuche(g, meine, func(x, y int) bool {
				nebenKiste := false
				for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					if g.kachel(x+d[0], y+d[1]) == 2 {
						nebenKiste = true
						break
					}
				}
				if !nebenKiste {
					return false
				}
				return bomberBotFlucht(g, s, [2]int{x, y}, gefahr)
			}, brenntBald)
		}

	}
	if len(weg) > 0 {
		dx, dy := bomberBotRichtung(meine, weg[0])
		return dx, dy, false, true
	}
	// Hier ist nichts zu holen (und bomben geht nicht): weiterlaufen.
	dx, dy, ok := bomberBotIrgendwohin(g, meine, gefahr)
	return dx, dy, false, ok
}

// bomberBotTreiben lässt den Bot entscheiden (über GET aufgerufen). In der
// Rundenpause blättert er weiter, sonst setzt er Richtung und Bombe.
func bomberBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Bomberman
	if g == nil || len(g.Bots) == 0 {
		return false
	}

	if g.Status == "runde" {
		if g.BotFrist.IsZero() {
			g.BotFrist = jetzt.Add(botWartezeit(g.BotStufe))
			return false
		}
		if jetzt.Before(g.BotFrist) {
			return false
		}
		g.BotFrist = time.Time{}
		_, meldung := bomberAktion(room, g.Bots[0], "weiter", 0, 0)
		return meldung == ""
	}

	if g.Status != "spiel" {
		return false
	}
	if g.BotFrist.IsZero() {
		g.BotFrist = jetzt.Add(bomberBotTakt(g, g.BotStufe))
		return false
	}
	if jetzt.Before(g.BotFrist) {
		return false
	}
	g.BotFrist = jetzt.Add(bomberBotTakt(g, g.BotStufe))
	return bomberBotSchritt(room)
}

// bomberBotSchritt führt genau eine Entscheidung aus (dieselbe Prüfung wie für
// einen Menschen).
func bomberBotSchritt(room *Room) bool {
	g := room.Bomberman
	dx, dy, bombe, ok := bomberBotZug(g)
	if !ok {
		return false
	}
	getan := false
	if dx != 0 || dy != 0 {
		if _, meldung := bomberAktion(room, g.Bots[0], "richtung", dx, dy); meldung == "" {
			getan = true
		}
	}
	if bombe {
		if _, meldung := bomberAktion(room, g.Bots[0], "bombe", 0, 0); meldung == "" {
			getan = true
		}
	}
	return getan
}
