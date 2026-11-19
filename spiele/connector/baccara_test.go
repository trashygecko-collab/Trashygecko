package main

import (
	"strconv"
	"testing"
)

// baccaraTestTisch baut einen Tisch in der Abrechnung; Kasse je Spieler in Cent.
func baccaraTestTisch(kassen []int) *BaccaraGame {
	g := &BaccaraGame{ID: "BA-TEST", HostID: "p1", Status: "runde", Runde: 1, BankSitz: 0, PonteSitz: -1}
	for i, kasse := range kassen {
		p := baccaraNeuerSpieler("p"+strconv.Itoa(i+1), "Spieler"+strconv.Itoa(i+1))
		p.Kasse = kasse
		g.Spieler = append(g.Spieler, p)
	}
	return g
}

// Wer die kleinste Bank nicht legen kann, bekommt sie nicht – sonst bliebe der
// Tisch stehen, weil „Bank legen“ dauerhaft abgelehnt wird.
func TestBaccaraBankUeberspringtArme(t *testing.T) {
	g := baccaraTestTisch([]int{0, 90, 200})
	baccaraBankWeiter(g)
	if g.Status != "bank" {
		t.Fatalf("Status falsch: %q", g.Status)
	}
	if g.BankSitz != 2 {
		t.Fatalf("Bank ging an Sitz %d statt an den zahlungsfähigen Sitz 2", g.BankSitz)
	}
}

// Kann niemand die Bank legen, ist die Partie vorbei statt festgefahren.
func TestBaccaraBankOhneGeldBeendetPartie(t *testing.T) {
	g := baccaraTestTisch([]int{0, 90})
	baccaraBankWeiter(g)
	if g.Status != "ende" {
		t.Fatalf("Status falsch: %q statt ende", g.Status)
	}
	if g.Note == "" {
		t.Fatal("kein Hinweis, warum die Partie vorbei ist")
	}
}

// Die normale Vergabe bleibt: der Nächste mit genug Geld bekommt die Bank.
func TestBaccaraBankGehtAnDenNaechsten(t *testing.T) {
	g := baccaraTestTisch([]int{0, 500, 500})
	baccaraBankWeiter(g)
	if g.Status != "bank" || g.BankSitz != 1 {
		t.Fatalf("Bank falsch vergeben: Status %q, Sitz %d", g.Status, g.BankSitz)
	}
	if g.Bank != 0 {
		t.Fatalf("Bank sollte leer sein, ist %d", g.Bank)
	}
}

// Setzt kein Pointeur mehr, ist die Partie vorbei statt in „setzen“ zu hängen.
func TestBaccaraSetzenOhneGeldBeendetPartie(t *testing.T) {
	g := baccaraTestTisch([]int{2000, 0})
	g.Bank = 2000
	baccaraNeueRunde(g)
	if g.Status != "ende" {
		t.Fatalf("Status falsch: %q statt ende", g.Status)
	}
	if g.Sieger != "Spieler1" {
		t.Fatalf("Sieger falsch: %q", g.Sieger)
	}
}

// Solange ein Pointeur Geld hat, geht die Runde normal weiter.
func TestBaccaraSetzenMitGeldBleibtRunde(t *testing.T) {
	g := baccaraTestTisch([]int{1000, 200})
	g.Bank = 1000
	baccaraNeueRunde(g)
	if g.Status != "setzen" {
		t.Fatalf("Status falsch: %q statt setzen", g.Status)
	}
}
