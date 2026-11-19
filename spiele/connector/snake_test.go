package main

import "testing"

func snakeTestSpiel(t *testing.T) *SnakeGame {
	t.Helper()
	g := snakeNeu("SN-TEST", "p1", []SnakeSpieler{
		snakeNeuerSpieler(1, "p1", "Anna"),
		snakeNeuerSpieler(2, "p2", "Ben"),
	}, 7)
	g.Spieler[0].Ready = true
	g.Spieler[1].Ready = true
	g.Runde = 1
	g.Status = "spiel"
	g.snakeNeueRunde()
	return g
}

func TestSnakeTischUndStart(t *testing.T) {
	g := snakeTestSpiel(t)
	if len(g.Spieler) != 2 {
		t.Fatalf("%d Spieler", len(g.Spieler))
	}
	a, b := g.Spieler[0], g.Spieler[1]
	if a.Segmente[0] != (SnaPunkt{5, 11}) || a.Richtung != (SnaPunkt{1, 0}) {
		t.Fatalf("Start von Spieler 1: Kopf %v, Richtung %v", a.Segmente[0], a.Richtung)
	}
	if b.Segmente[0] != (SnaPunkt{15, 11}) || b.Richtung != (SnaPunkt{-1, 0}) {
		t.Fatalf("Start von Spieler 2: Kopf %v, Richtung %v", b.Segmente[0], b.Richtung)
	}
	if len(a.Segmente) != 3 || len(b.Segmente) != 3 {
		t.Fatalf("Startlängen %d und %d", len(a.Segmente), len(b.Segmente))
	}
	if !g.HatApfel {
		t.Fatal("kein Apfel auf dem Feld")
	}
	// Der Apfel liegt auf einer freien Kachel.
	for _, p := range g.Spieler {
		for _, teil := range p.Segmente {
			if teil == g.Apfel {
				t.Fatalf("Apfel liegt auf einer Schlange: %v", g.Apfel)
			}
		}
	}
	r := g.raster()
	if len(r) != snaSpalten*snaZeilen {
		t.Fatalf("Raster hat %d Felder", len(r))
	}
	if r[11*snaSpalten+5] != 1 || r[11*snaSpalten+15] != 2 {
		t.Fatalf("Raster zeigt die Köpfe nicht: %d/%d", r[11*snaSpalten+5], r[11*snaSpalten+15])
	}
}

func TestSnakeBewegenUndUmkehren(t *testing.T) {
	g := snakeTestSpiel(t)
	g.snakeSchritt()
	if g.Spieler[0].Segmente[0] != (SnaPunkt{6, 11}) {
		t.Fatalf("Kopf nach einem Schritt: %v", g.Spieler[0].Segmente[0])
	}
	if g.Spieler[1].Segmente[0] != (SnaPunkt{14, 11}) {
		t.Fatalf("Kopf von Spieler 2: %v", g.Spieler[1].Segmente[0])
	}
	if len(g.Spieler[0].Segmente) != 3 {
		t.Fatalf("Länge ohne Apfel: %d", len(g.Spieler[0].Segmente))
	}

	// Umkehren ist verboten, gleiche Richtung auch.
	if g.Spieler[0].richtungMerken(SnaPunkt{-1, 0}) {
		t.Fatal("Umkehren wurde angenommen")
	}
	if g.Spieler[0].richtungMerken(SnaPunkt{1, 0}) {
		t.Fatal("gleiche Richtung wurde angenommen")
	}
	// Höchstens drei Wechsel auf Vorrat.
	if !g.Spieler[0].richtungMerken(SnaPunkt{0, -1}) ||
		!g.Spieler[0].richtungMerken(SnaPunkt{-1, 0}) ||
		!g.Spieler[0].richtungMerken(SnaPunkt{0, 1}) {
		t.Fatal("Richtungswechsel wurden zu früh abgelehnt")
	}
	if g.Spieler[0].richtungMerken(SnaPunkt{1, 0}) {
		t.Fatal("vierter Wechsel passt nicht in die Warteschlange")
	}
}

func TestSnakeApfelWaechst(t *testing.T) {
	g := snakeTestSpiel(t)
	// Apfel direkt vor Spieler 1 legen.
	g.Apfel = SnaPunkt{6, 11}
	g.HatApfel = true
	g.snakeSchritt()
	if len(g.Spieler[0].Segmente) != 4 {
		t.Fatalf("Länge nach dem Apfel: %d", len(g.Spieler[0].Segmente))
	}
	if g.Spieler[0].Punkte != 10 {
		t.Fatalf("Punkte nach dem Apfel: %d", g.Spieler[0].Punkte)
	}
	if g.HatApfel && g.Apfel == (SnaPunkt{6, 11}) {
		t.Fatal("der gefressene Apfel liegt noch da")
	}
}

func TestSnakeWandUndRunde(t *testing.T) {
	g := snakeTestSpiel(t)
	// Spieler 1 steht vor dem rechten Rand.
	g.Spieler[0].Segmente = []SnaPunkt{{20, 11}, {19, 11}, {18, 11}}
	g.Spieler[0].Richtung = SnaPunkt{1, 0}
	g.snakeSchritt()
	if g.Spieler[0].Leben {
		t.Fatal("die Schlange überlebt den Rand")
	}
	if g.Status != "runde" || g.Gewinner != 2 {
		t.Fatalf("Runde falsch: Status %s, Gewinner %d", g.Status, g.Gewinner)
	}
	if g.Spieler[1].Sieg != 1 {
		t.Fatalf("Siege von Spieler 2: %d", g.Spieler[1].Sieg)
	}
	if g.Meldung == "" {
		t.Fatal("Meldung fehlt")
	}
}

func TestSnakeKopfAnKopf(t *testing.T) {
	g := snakeTestSpiel(t)
	g.Spieler[0].Segmente = []SnaPunkt{{5, 11}}
	g.Spieler[0].Richtung = SnaPunkt{1, 0}
	g.Spieler[1].Segmente = []SnaPunkt{{7, 11}}
	g.Spieler[1].Richtung = SnaPunkt{-1, 0}
	g.snakeSchritt()
	if g.Spieler[0].Leben || g.Spieler[1].Leben {
		t.Fatal("Kopf an Kopf überlebt jemand")
	}
	if g.Status != "runde" || g.Gewinner != 0 {
		t.Fatalf("Unentschieden falsch: Status %s, Gewinner %d", g.Status, g.Gewinner)
	}
	if g.Spieler[0].Sieg != 0 || g.Spieler[1].Sieg != 0 {
		t.Fatal("bei Kopf an Kopf darf niemand einen Sieg bekommen")
	}
}

func TestSnakeEigeneSpur(t *testing.T) {
	g := snakeTestSpiel(t)
	// Eine Schlange, die im Kreis läuft, beißt sich in den Schwanz – dort darf
	// der Kopf nur hin, wenn der Ring verschwindet.
	g.Spieler[0].Segmente = []SnaPunkt{{5, 11}, {5, 12}, {4, 12}, {4, 11}}
	g.Spieler[0].Richtung = SnaPunkt{0, -1}   // von (5,11) auf (5,10)
	g.Spieler[0].Wartet = []SnaPunkt{{-1, 0}} // danach links, das ist das eigene Ende
	g.snakeSchritt()
	if !g.Spieler[0].Leben {
		t.Fatal("der Zug auf das frei werdende Schwanzende gilt als Biss")
	}

	// In die eigene Mitte führt dagegen in den Tod.
	g2 := snakeTestSpiel(t)
	g2.Spieler[0].Segmente = []SnaPunkt{{5, 11}, {6, 11}, {6, 12}, {5, 12}, {4, 12}}
	g2.Spieler[0].Richtung = SnaPunkt{0, 1} // auf (5,12): eigener Körper
	g2.snakeSchritt()
	if g2.Spieler[0].Leben {
		t.Fatal("Biss in den eigenen Körper überlebt")
	}
}

func TestSnakeMatchUndAnsicht(t *testing.T) {
	g := snakeTestSpiel(t)
	for runde := 1; runde <= 3; runde++ {
		g.Runde = runde
		g.Status = "spiel"
		g.snakeNeueRunde()
		g.Spieler[0].Segmente = []SnaPunkt{{20, 5}, {19, 5}, {18, 5}}
		g.Spieler[0].Richtung = SnaPunkt{1, 0}
		g.snakeSchritt()
	}
	if g.Status != "ende" || g.Sieger != "Ben" || g.Runde != 3 {
		t.Fatalf("Match falsch: Status %s, Sieger %q, Runde %d", g.Status, g.Sieger, g.Runde)
	}
	if g.Spieler[1].Sieg != 3 {
		t.Fatalf("Siege: %d", g.Spieler[1].Sieg)
	}

	a := snakeAnsicht(g, "p2")
	if a == nil || a.Du != 1 || a.Spalten != 21 || a.Zeilen != 23 || a.Raster == "" {
		t.Fatalf("Ansicht unvollständig: %+v", a)
	}
	if len(a.Spieler) != 2 || a.Spieler[1].Richtung != "links" {
		t.Fatalf("Spielersicht falsch: %+v", a.Spieler)
	}
	if len(a.Raster) > 120 {
		t.Fatalf("Raster zu lang für ein Feld: %d Zeichen", len(a.Raster))
	}
}
