package main

/* ===========================================================================
   BOT FÜR HALMA

   Halma ist ein Laufspiel: jede Figur soll in das gegenüberliegende Haus. Ein
   Schritt geht auf ein freies Nachbarfeld, ein Sprung über eine Figur (auch
   kettenweise) bleibt auf demselben Feld stehen, muss also nicht direkt hinter
   dem übersprungenen Feld landen.

   Der Bot spielt über `halmaZuegeVon` und damit über genau die Zugliste, die
   auch das Spiel prüft – er kann nichts ziehen, was die Regeln nicht erlauben.

   Bewertet wird mit dem Abstand zum Ziel: ein einmaliger Breitendurchlauf über
   das Brett liefert für jedes Feld die Zahl der Schritte bis zum Haus, und der
   Wert einer Stellung ist die Summe dieser Abstände über die eigenen Figuren.
   Für alle drei Brettarten (klassisch, Stern, Sprunghalma) ist das derselbe
   Maßstab, weil er nur den Nachbarschaftsgraphen benutzt.

     · leicht  – zieht irgendeinen erlaubten Zug
     · mittel  – nimmt den Zug, der die eigene Abstandssumme am stärksten senkt
     · schwer  – rechnet wie mittel; gemessen brachte eine tiefere Rechnung auf
                 diesem Brett nichts (Gegenzug-Modell 0:8, eigener Folgezug
                 ebenfalls 0:8 gegen die flache Variante), deshalb unterscheidet
                 sich die Stufe nur in der Denkpause

   Ohne erlaubten Zug (kann bei Halma nicht vorkommen, weil immer ein Feld
   frei ist) gibt der Bot auf, statt die Partie zu blockieren.
   =========================================================================== */

import (
	"time"
)

type halmaBotZug struct {
	von  int
	nach int
}

// halmaBotAbstand rechnet für jedes Feld die Zahl der Schritte bis zur
// Zielzone eines Platzes (Breitensuche über die Nachbarschaft).
func halmaBotAbstand(b *HalmaBrett, platz int) []int {
	fern := make([]int, b.Anzahl)
	for i := range fern {
		fern[i] = -1
	}
	var rand []int
	for _, i := range b.zielZone(platz) {
		if i >= 0 && i < len(fern) && fern[i] < 0 {
			fern[i] = 0
			rand = append(rand, i)
		}
	}
	for len(rand) > 0 {
		var naechste []int
		for _, f := range rand {
			for _, n := range b.nachbar[f] {
				if n >= 0 && n < len(fern) && fern[n] < 0 {
					fern[n] = fern[f] + 1
					naechste = append(naechste, n)
				}
			}
		}
		rand = naechste
	}
	// Felder, die die Suche nicht erreicht, sind unerreichbar – sie sollen
	// nicht als „nah“ durchgehen.
	for i := range fern {
		if fern[i] < 0 {
			fern[i] = b.Anzahl
		}
	}
	return fern
}

// halmaBotSumme ist der Wert einer Stellung für einen Platz: die Summe der
// Abstände aller eigenen Figuren zu ihrem Zielhaus.
func halmaBotSumme(abstand, belegung []int, platz int) int {
	summe := 0
	for i, p := range belegung {
		if p == platz {
			summe += abstand[i]
		}
	}
	return summe
}

// halmaBotZahl zählt die eigenen Figuren, die noch im eigenen Haus stehen.
// Wer sein Haus nicht räumt, versperrt dem Gegner das Ziel – deshalb wird das
// Räumen im Wert einer Stellung mitgezählt.
func halmaBotZahl(b *HalmaBrett, belegung []int, platz int) int {
	n := 0
	for _, i := range b.zonen[platz] {
		if belegung[i] == platz {
			n++
		}
	}
	return n
}

// halmaBotZuege sammelt alle erlaubten Züge eines Platzes.
func halmaBotZuege(g *HalmaGame, platz int) []halmaBotZug {
	b := halmaBrettFuer(g)
	var out []halmaBotZug
	for i, p := range g.Belegung {
		if p != platz {
			continue
		}
		for _, z := range halmaZuegeVon(b, g.Belegung, g.Variante, i) {
			out = append(out, halmaBotZug{von: i, nach: z.Nach})
		}
	}
	return out
}

// halmaBotWahl wählt den Zug für einen Platz. ok=false heißt: kein Zug
// möglich (kommt bei Halma praktisch nicht vor).
func halmaBotWahl(g *HalmaGame, platz int, stufe string) (halmaBotZug, bool) {
	zuege := halmaBotZuege(g, platz)
	if len(zuege) == 0 {
		return halmaBotZug{}, false
	}
	// Den letzten eigenen Zug nicht sofort zurücknehmen: sonst schieben sich
	// zwei Bots in einer Stellung ohne Fortschritt ewig denselben Stein hin
	// und her. Bleibt nur dieser eine Zug, ist er natürlich erlaubt.
	if letzt, da := g.BotZug[platz]; da && len(zuege) > 1 {
		var ohne []halmaBotZug
		for _, z := range zuege {
			if z.von == letzt[1] && z.nach == letzt[0] {
				continue
			}
			ohne = append(ohne, z)
		}
		if len(ohne) > 0 {
			zuege = ohne
		}
	}
	stufe = botStufe(stufe)
	if stufe == botLeicht {
		return zuege[botZufall().Intn(len(zuege))], true
	}
	b := halmaBrettFuer(g)
	abstand := halmaBotAbstand(b, platz)
	type bewertet struct {
		zug  halmaBotZug
		wert int
	}
	liste := make([]bewertet, 0, len(zuege))
	for _, z := range zuege {
		belegung := g.Belegung[z.von]
		g.Belegung[z.von], g.Belegung[z.nach] = -1, belegung
		wert := halmaBotSumme(abstand, g.Belegung, platz) * 2
		wert += halmaBotZahl(b, g.Belegung, platz) * 5
		g.Belegung[z.von], g.Belegung[z.nach] = belegung, -1
		liste = append(liste, bewertet{zug: z, wert: wert})
	}
	// Sortierung: bester (kleinster) Wert zuerst.
	for i := 1; i < len(liste); i++ {
		for j := i; j > 0 && liste[j].wert < liste[j-1].wert; j-- {
			liste[j], liste[j-1] = liste[j-1], liste[j]
		}
	}
	best := liste[0].wert
	var kandidaten []halmaBotZug
	for _, l := range liste {
		if l.wert == best {
			kandidaten = append(kandidaten, l.zug)
		}
	}
	return kandidaten[botZufall().Intn(len(kandidaten))], true
}

// halmaBotTreiben lässt den Computer ziehen, sobald seine Frist abgelaufen ist.
// Aufgerufen aus dem GET des Handlers – derselbe Takt wie bei den Menschen.
func halmaBotTreiben(g *HalmaGame, jetzt time.Time) bool {
	if g == nil || g.Status != "active" || g.BotID == "" || len(g.Spieler) == 0 {
		return false
	}
	if g.Current < 0 || g.Current >= len(g.Spieler) || g.Spieler[g.Current].ID != g.BotID {
		return false
	}
	if jetzt.Before(g.BotFrist) {
		return false
	}
	platz := g.Spieler[g.Current].Platz
	zug, ok := halmaBotWahl(g, platz, g.BotStufe)
	if !ok {
		g.Current = (g.Current + 1) % len(g.Spieler)
		g.BotFrist = jetzt.Add(botWartezeit(g.BotStufe))
		return true
	}
	if _, ok := halmaZiehen(g, g.Current, zug.von, zug.nach); !ok {
		return false
	}
	g.BotFrist = jetzt.Add(botWartezeit(g.BotStufe))
	return true
}
