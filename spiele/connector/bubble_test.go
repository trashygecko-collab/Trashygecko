package main

import (
	"math"
	"testing"
	"time"
)

// Die Farben der Startreihen stammen aus mulberry32(7) – dieselbe Folge wie
// in der lokalen Fassung, damit beide Felder gleich aussehen.
var bubStartReihen = []string{"0032211", "0221032", "0112130", "3000123", "0322232", "3300110"}

func TestBubbleZufallUndStart(t *testing.T) {
	b := bubNeuesBrett()
	if len(b.Feld) != bubReihen {
		t.Fatalf("%d Reihen statt %d", len(b.Feld), bubReihen)
	}
	for r := range b.Feld {
		if len(b.Feld[r].Z) != bubSp {
			t.Fatalf("Reihe %d hat %d Spalten", r, len(b.Feld[r].Z))
		}
		if b.Feld[r].V != r%2 {
			t.Fatalf("Reihe %d hat Versatz %d", r, b.Feld[r].V)
		}
		got := ""
		for _, v := range b.Feld[r].Z {
			if v < 1 || v > bubFarben {
				t.Fatalf("Farbe %d außerhalb 1..%d", v, bubFarben)
			}
			got += string(rune('0' + v - 1))
		}
		if got != bubStartReihen[r] {
			t.Fatalf("Reihe %d: %s, erwartet %s", r, got, bubStartReihen[r])
		}
	}
	if b.kugeln() != bubReihen*bubSp {
		t.Fatalf("%d Kugeln auf dem Startfeld", b.kugeln())
	}
	if b.Zustand != "start" {
		t.Fatalf("Zustand %q", b.Zustand)
	}
	// Zwei Felder starten gleich.
	z := bubNeuesBrett()
	for r := range b.Feld {
		for c := 0; c < bubSp; c++ {
			if z.Feld[r].Z[c] != b.Feld[r].Z[c] {
				t.Fatal("zwei Felder starten unterschiedlich")
			}
		}
	}
	// Goldene Linie: 66 px über dem Boden.
	if bubH-bubTod != 494 {
		t.Fatalf("goldene Linie bei %v", bubH-bubTod)
	}
}

func TestBubbleSchussHaengtAn(t *testing.T) {
	b := bubNeuesBrett()
	b.los()
	vorher := b.kugeln()
	nachschub := b.Naechste
	if !b.schiessen(0) {
		t.Fatal("Schuss ging nicht los")
	}
	if b.schiessen(0) {
		t.Fatal("zweiter Schuss während des Flugs erlaubt")
	}
	if b.Schuesse != 1 || b.Aktuell != nachschub {
		t.Fatalf("Nachschub falsch: Schüsse %d, Kugel %d statt %d", b.Schuesse, b.Aktuell, nachschub)
	}
	for i := 0; i < 200 && b.Kugel != nil; i++ {
		b.schritt()
	}
	if b.Kugel != nil {
		t.Fatal("Kugel hängt in der Luft")
	}
	nach := b.kugeln()
	// Entweder eine Kugel mehr oder eine Dreiergruppe geplatzt.
	if nach != vorher+1 && nach > vorher {
		t.Fatalf("Kugeln %d → %d (erwartet +1 oder weniger)", vorher, nach)
	}
	if b.Zustand != "laeuft" {
		t.Fatalf("Feld ist nach einem Schuss schon vorbei: %q", b.Zustand)
	}
}

func TestBubbleDreiergruppeUndAbsturz(t *testing.T) {
	b := bubNeuesBrett()
	b.los()
	// Ein leeres Feld mit drei gleichfarbigen Kugeln in der obersten Reihe.
	b.Feld = []bubReihe{{V: 0, Z: make([]int, bubSp)}}
	b.Feld[0].Z[0] = 1
	b.Feld[0].Z[1] = 1
	// Eine Kugel hängt allein darunter (Reihe 1) und muss mit abstürzen.
	b.Feld = append(b.Feld, bubReihe{V: 1, Z: make([]int, bubSp)})
	b.Feld[1].Z[0] = 2
	vorher := b.kugeln()
	b.Kugel = &bubKugel{X: b.zelleX(0, 2), Y: b.zelleY(0), Farbe: 0}
	b.anhaengen(b.zelleX(0, 2), b.zelleY(0))
	if b.Punkte != 30+20 {
		t.Fatalf("Punkte %d (Dreiergruppe 30 + Absturz 20 erwartet)", b.Punkte)
	}
	if b.kugeln() != 0 {
		t.Fatalf("%d Kugeln übrig (vorher %d)", b.kugeln(), vorher)
	}
}

func TestBubbleNachschubUndLinie(t *testing.T) {
	g := bubbleNeu("BG-TEST", "p1", []BubbleSpieler{
		{ID: "p1", Name: "Anna", Nr: 1},
		{ID: "p2", Name: "Ben", Nr: 2},
	})
	g.Spieler[0].Ready = true
	g.Spieler[1].Ready = true
	g.starten()
	b := g.Spieler[0].Brett
	reihen := len(b.Feld)

	// Sechs Schüsse ohne Flug: jeder Schuss zählt, nach dem sechsten kommt
	// eine neue Reihe – genau wie lokal.
	for i := 0; i < bubReihenBisNeu; i++ {
		if !b.schiessen(0) {
			t.Fatalf("Schuss %d ging nicht los", i+1)
		}
		b.Kugel = nil
		b.schritt()
	}
	if len(b.Feld) != reihen+1 {
		t.Fatalf("%d Reihen nach sechs Schüssen, erwartet %d", len(b.Feld), reihen+1)
	}
	if b.Schuesse != 0 {
		t.Fatalf("Schusszähler steht bei %d", b.Schuesse)
	}
}

func TestBubbleGoldeneLinieVerliert(t *testing.T) {
	g := bubbleNeu("BG-TEST", "p1", []BubbleSpieler{
		{ID: "p1", Name: "Anna", Nr: 1},
		{ID: "p2", Name: "Ben", Nr: 2},
	})
	g.starten()
	anna := g.Spieler[0].Brett

	// 22 Reihen reichen bis an die goldene Linie (zelleY(21)+R ≥ 494).
	anna.Feld = nil
	for r := 0; r < 22; r++ {
		anna.Feld = append(anna.Feld, bubReihe{V: r % 2, Z: make([]int, bubSp)})
	}
	anna.pruefeEnde()
	if anna.Zustand != "laeuft" {
		t.Fatalf("leere Reihen beenden das Feld: Zustand %q", anna.Zustand)
	}

	// Jetzt eine Kugel in der untersten Reihe – sie berührt die Linie.
	anna.Feld[len(anna.Feld)-1].Z[0] = 1
	anna.pruefeEnde()
	if anna.Zustand != "ende" {
		t.Fatal("Kugel unter der goldenen Linie beendet das Feld nicht")
	}

	g.verloren(0)
	if g.Status != "ende" || g.Gewinner != 2 || g.Sieger != "Ben" {
		t.Fatalf("Duellende falsch: Status %s, Gewinner %d, Sieger %q", g.Status, g.Gewinner, g.Sieger)
	}
	if g.Meldung == "" {
		t.Fatal("Meldung fehlt")
	}
	// Ein zweiter Verlust ändert nichts mehr.
	g.verloren(1)
	if g.Gewinner != 2 {
		t.Fatal("Gewinner wurde nachträglich überschrieben")
	}
}

func TestBubbleWandUndAnsicht(t *testing.T) {
	g := bubbleNeu("BG-TEST", "p1", []BubbleSpieler{
		{ID: "p1", Name: "Anna", Nr: 1},
		{ID: "p2", Name: "Ben", Nr: 2},
	})
	g.starten()
	b := g.Spieler[0].Brett
	if !b.schiessen(1.25) {
		t.Fatal("Schuss ging nicht los")
	}
	min, max, gesehen := math.Inf(1), math.Inf(-1), 0
	for i := 0; i < 300 && b.Kugel != nil; i++ {
		b.schritt()
		if b.Kugel != nil {
			min = math.Min(min, b.Kugel.X)
			max = math.Max(max, b.Kugel.X)
			gesehen++
		}
	}
	if b.Kugel != nil {
		t.Fatal("Kugel hängt in der Luft")
	}
	if gesehen == 0 {
		t.Fatal("kein Flug beobachtet")
	}
	if min < bubR || max > bubW-bubR {
		t.Fatalf("Kugel verlässt das Feld: x %.1f … %.1f", min, max)
	}

	a := bubbleAnsicht(g, "p2")
	if a == nil || a.Du != 1 || a.Spalten != bubSp || a.Linie != 494 {
		t.Fatalf("Ansicht unvollständig: %+v", a)
	}
	if a.Brett == nil || len(a.Brett.Reihen) == 0 {
		t.Fatal("eigenes Brett fehlt in der Ansicht")
	}
	if len(a.Spieler) != 2 {
		t.Fatalf("%d Spieler in der Ansicht", len(a.Spieler))
	}
	// Jeder bekommt sein eigenes Brett: nach einem Schuss von Anna unterscheidet
	// sich die Kugelzahl.
	anna := bubbleAnsicht(g, "p1")
	ben := bubbleAnsicht(g, "p2")
	if anna.Brett.Schuesse != 1 || ben.Brett.Schuesse != 0 {
		t.Fatalf("Schüsse vertauscht: Anna %d, Ben %d", anna.Brett.Schuesse, ben.Brett.Schuesse)
	}

	// Vorsprung: drei Sekunden werden nachgerechnet, mehr nicht.
	g.Zuletzt = time.Now().Add(-10 * time.Second)
	bubbleWeiter(g)
	if g.Rest > bubTakt {
		t.Fatalf("Restzeit nach dem Nachrechnen: %v", g.Rest)
	}
}
