package main

import (
	"math"
	"testing"
	"time"
)

func bomberTestGame(t *testing.T) *BomberGame {
	t.Helper()
	g := bomberNeu("TG-TEST", "p1", []BomberSpieler{
		bomberNeuerSpieler(1, "p1", "Anna"),
		bomberNeuerSpieler(2, "p2", "Ben"),
	})
	g.Status = "spiel"
	bomberNeueRunde(g)
	g.Vorlauf = 0
	return g
}

func TestBomberTischUndStart(t *testing.T) {
	g := bomberTestGame(t)
	if len(g.Raster) != bomberSpalten*bomberZeilen {
		t.Fatalf("Raster hat %d Felder, erwartet %d", len(g.Raster), bomberSpalten*bomberZeilen)
	}
	if g.Runde != 1 || g.Gewinner != -1 {
		t.Fatalf("Startwerte falsch: Runde %d, Gewinner %d", g.Runde, g.Gewinner)
	}
	p1, p2 := g.Spieler[0], g.Spieler[1]
	if p1.X != 1.5 || p1.Y != 1.5 || p2.X != 13.5 || p2.Y != 11.5 {
		t.Fatalf("Startecken falsch: P1 (%.1f,%.1f) P2 (%.1f,%.1f)", p1.X, p1.Y, p2.X, p2.Y)
	}
	if !p1.Leben || !p2.Leben || p1.BombMax != 1 || p1.Feuer != 1 || p1.Tempo != 1 {
		t.Fatalf("Startausrüstung falsch: %+v", p1)
	}
	// Rand und Pfeiler sind Wände, die Startecken frei.
	for x := 0; x < bomberSpalten; x++ {
		if g.kachel(x, 0) != 1 || g.kachel(x, bomberZeilen-1) != 1 {
			t.Fatalf("Rand bei x=%d nicht dicht", x)
		}
	}
	for _, e := range [][2]int{{1, 1}, {2, 1}, {1, 2}, {13, 11}} {
		if g.kachel(e[0], e[1]) != 0 {
			t.Fatalf("Startfeld (%d,%d) ist nicht frei", e[0], e[1])
		}
	}
	if g.kachel(0, 0) != 1 || g.kachel(2, 2) != 1 {
		t.Fatal("Pfeiler fehlen")
	}
}

func TestBomberBewegungUndWand(t *testing.T) {
	g := bomberTestGame(t)
	// Leeres Feld, damit nur Wände im Weg stehen können.
	for i := range g.Raster {
		g.Raster[i] = 0
	}
	p := &g.Spieler[0]
	p.DX = 1
	for i := 0; i < 60; i++ {
		g.schritt(bomberTakt)
	}
	if math.Abs(p.X-4.55) > 0.2 {
		t.Fatalf("nach einer Sekunde x=%.2f, erwartet rund 4,55", p.X)
	}
	if p.BlickDX != 1 || p.BlickDY != 0 {
		t.Fatalf("Blickrichtung falsch: %d/%d", p.BlickDX, p.BlickDY)
	}

	// Wand direkt vor die Nase und weiterlaufen.
	g.Raster[1*bomberSpalten+int(math.Floor(p.X))+1] = 1
	vorher := p.X
	for i := 0; i < 30; i++ {
		g.schritt(bomberTakt)
	}
	if p.X-vorher > 0.7 {
		t.Fatalf("Wand bremst nicht: x %.2f → %.2f", vorher, p.X)
	}

	// Ohne gehaltene Richtung bleibt der Bomber stehen.
	p.DX = 0
	ruhe := p.X
	for i := 0; i < 30; i++ {
		g.schritt(bomberTakt)
	}
	if p.X != ruhe {
		t.Fatalf("Bomber läuft ohne Eingabe weiter: %.2f → %.2f", ruhe, p.X)
	}
}

func TestBomberBombeFlammeUndKette(t *testing.T) {
	g := bomberTestGame(t)
	for i := range g.Raster {
		g.Raster[i] = 0
	}
	p1 := &g.Spieler[0]
	p1.Feuer = 2
	p1.BombMax = 2
	p1.X, p1.Y = 5.5, 5.5
	if !g.legeBombe(p1) {
		t.Fatal("erste Bombe wurde nicht gelegt")
	}
	// Zweite Bombe zwei Felder weiter – soll durch die Kette mitgehen.
	p1.X = 7.5
	if !g.legeBombe(p1) {
		t.Fatal("zweite Bombe wurde nicht gelegt")
	}
	p1.X, p1.Y = 0.5, 0.5
	if len(g.Bomben) != 2 {
		t.Fatalf("%d Bomben statt 2", len(g.Bomben))
	}
	// 2,4 s warten: die erste Bombe geht hoch und nimmt die zweite mit.
	for i := 0; i < 150; i++ {
		g.schritt(bomberTakt)
	}
	if len(g.Bomben) != 0 {
		t.Fatalf("Bomben übrig: %d", len(g.Bomben))
	}
	flammen := map[[2]int]bool{}
	for _, f := range g.Flammen {
		flammen[[2]int{f.X, f.Y}] = true
	}
	// Bombe bei (5,5) mit Reichweite 2: Kreuz bis zwei Felder.
	for _, z := range [][2]int{{5, 5}, {6, 5}, {7, 5}, {5, 4}, {5, 7}, {9, 5}} {
		if !flammen[z] {
			t.Fatalf("Flamme auf (%d,%d) fehlt", z[0], z[1])
		}
	}
	if flammen[[2]int{5, 8}] {
		t.Fatal("Flamme reicht weiter als die Reichweite")
	}
	// Nach 0,62 s ist das Feuer aus.
	for i := 0; i < 45; i++ {
		g.schritt(bomberTakt)
	}
	if len(g.Flammen) != 0 {
		t.Fatalf("%d Flammen brennen zu lange", len(g.Flammen))
	}
}

func TestBomberRundeUndMatch(t *testing.T) {
	g := bomberTestGame(t)
	// Ben steht in einer Flamme.
	g.Flammen = []BomberFlamme{{X: 13, Y: 11, Rest: bomberFlammenzeit}}
	g.schritt(bomberTakt)
	if g.Spieler[1].Leben {
		t.Fatal("Ben überlebt die Flamme")
	}
	if g.Status != "runde" || g.Gewinner != 1 || g.Spieler[0].Sieg != 1 {
		t.Fatalf("Runde falsch: Status %s, Gewinner %d, Siege %d", g.Status, g.Gewinner, g.Spieler[0].Sieg)
	}
	if g.Meldung == "" {
		t.Fatal("Meldung fehlt")
	}

	// Zwei weitere Runden für Anna → Match vorbei.
	for runde := 2; runde <= 3; runde++ {
		g.Runde = runde
		bomberNeueRunde(g)
		g.Status = "spiel"
		g.Vorlauf = 0
		g.Flammen = []BomberFlamme{{X: 13, Y: 11, Rest: bomberFlammenzeit}}
		g.schritt(bomberTakt)
	}
	if g.Status != "ende" {
		t.Fatalf("Match endet nicht: Status %s, Siege %d:%d", g.Status, g.Spieler[0].Sieg, g.Spieler[1].Sieg)
	}
	if g.Sieger != "Anna" || g.Runde != 3 {
		t.Fatalf("Sieger %q in Runde %d", g.Sieger, g.Runde)
	}
}

func TestBomberUnentschieden(t *testing.T) {
	g := bomberTestGame(t)
	g.Flammen = []BomberFlamme{
		{X: 1, Y: 1, Rest: bomberFlammenzeit},
		{X: 13, Y: 11, Rest: bomberFlammenzeit},
	}
	g.schritt(bomberTakt)
	if g.Status != "runde" || g.Gewinner != 0 {
		t.Fatalf("Unentschieden falsch: Status %s, Gewinner %d", g.Status, g.Gewinner)
	}
	if g.Spieler[0].Sieg != 0 || g.Spieler[1].Sieg != 0 {
		t.Fatal("Bei Unentschieden darf niemand einen Sieg bekommen")
	}
}

func TestBomberAnsichtUndVorsprung(t *testing.T) {
	g := bomberTestGame(t)
	a := bomberAnsicht(g, "p2")
	if a == nil || a.Du != 1 || len(a.Spieler) != 2 || a.Raster == "" {
		t.Fatalf("Ansicht unvollständig: %+v", a)
	}
	if a.Spalten != 15 || a.Zeilen != 13 || a.Ziel != 3 {
		t.Fatalf("Maße falsch: %dx%d, Ziel %d", a.Spalten, a.Zeilen, a.Ziel)
	}
	if len(a.Raster) > 700 {
		t.Fatalf("Raster zu lang für ein Feld: %d Zeichen", len(a.Raster))
	}

	// Drei Sekunden Vorsprung werden nachgerechnet, aber nicht mehr.
	g.Zuletzt = time.Now().Add(-10 * time.Second)
	bomberWeiter(g)
	if g.Vorlauf != 0 {
		t.Fatalf("Vorlauf nach 3 s noch %.2f", g.Vorlauf)
	}
}
