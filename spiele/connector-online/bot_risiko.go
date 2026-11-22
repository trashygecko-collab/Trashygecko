package main

// Bots für Risiko (Phase 2, 22.11.2026).
//
// Vier Stufen (leicht, mittel, schwer, experte) und sieben Persönlichkeiten.
// Ein Bot plant einen vollständigen Zug auf einer Kopie des Standes
// (risikoBotPlanen) und gibt die Schritte als Liste zurück; der Raum spielt
// sie einen nach dem anderen am echten Stand ab und prüft dabei jede Aktion
// mit denselben Regeln wie für Menschen. Der Bot benutzt nur, was auch ein
// Mensch sieht – keine verdeckten Karten, keine zusätzlichen Armeen.

import (
	"math/rand"
	"strings"
)

var risikoBotWesenListe = []string{
	"aggressiv", "defensiv", "expansion", "kontinent", "taktiker", "risikofreudig", "ausgewogen",
}

func risikoBotStufe(wert string) string {
	switch strings.ToLower(strings.TrimSpace(wert)) {
	case "leicht":
		return "leicht"
	case "schwer":
		return "schwer"
	case "experte":
		return "experte"
	}
	return "mittel"
}

func risikoBotWesen(wert string) string {
	w := strings.ToLower(strings.TrimSpace(wert))
	for _, bekannt := range risikoBotWesenListe {
		if w == bekannt {
			return w
		}
	}
	return "ausgewogen"
}

func risikoBotStufeText(stufe string) string {
	switch risikoBotStufe(stufe) {
	case "leicht":
		return "leicht"
	case "schwer":
		return "schwer"
	case "experte":
		return "experte"
	}
	return "mittel"
}

func risikoBotWesenText(wesen string) string {
	switch risikoBotWesen(wesen) {
	case "aggressiv":
		return "aggressiv"
	case "defensiv":
		return "defensiv"
	case "expansion":
		return "expansionsorientiert"
	case "kontinent":
		return "Kontinentjäger"
	case "taktiker":
		return "Taktiker"
	case "risikofreudig":
		return "risikofreudig"
	}
	return "ausgewogen"
}

// risikoBotGewicht sind die Stellschrauben einer Persönlichkeit. Sie ändern
// nur die Bewertung, nie die Regeln.
type risikoBotGewicht struct {
	angriff     float64 // niedriger = greift eher an
	garrison    int     // so viele Armeen bleiben mindestens hinten stehen
	kontinent   float64 // Gewicht für fast vollständige Kontinente
	expansion   float64 // Gewicht für schwache Nachbarn
	verteidigen float64 // Gewicht für bedrohte eigene Grenzen
	risiko      float64 // Abzug bei unsicheren Angriffen
}

func risikoBotGewichte(stufe, wesen string) risikoBotGewicht {
	g := risikoBotGewicht{angriff: 0.30, garrison: 2, kontinent: 1, expansion: 1, verteidigen: 1, risiko: 1}
	switch risikoBotStufe(stufe) {
	case "leicht":
		g.angriff, g.risiko = 0.05, 0.3
	case "schwer":
		g.angriff, g.kontinent, g.verteidigen = 0.42, 1.6, 1.5
	case "experte":
		g.angriff, g.kontinent, g.expansion, g.verteidigen, g.risiko = 0.48, 1.8, 1.4, 1.8, 1.6
	}
	switch risikoBotWesen(wesen) {
	case "aggressiv":
		g.angriff -= 0.12
		g.garrison = 1
	case "defensiv":
		g.angriff += 0.15
		g.garrison = 3
		g.verteidigen *= 1.8
	case "expansion":
		g.expansion *= 1.8
		g.kontinent *= 0.8
	case "kontinent":
		g.kontinent *= 2.2
		g.expansion *= 0.7
	case "taktiker":
		g.kontinent *= 1.2
		g.verteidigen *= 1.3
	case "risikofreudig":
		g.angriff -= 0.1
		g.risiko *= 0.5
	}
	return g
}

// risikoBotNachbarn liefert eigene Nachbarn bzw. Feinde eines Gebiets.
func risikoBotFeinde(g *RisikoSpiel, id string, k int) []string {
	var aus []string
	for _, n := range risikoGebietNach[id] {
		if g.Gebiete[n].Besitzer != k {
			aus = append(aus, n)
		}
	}
	return aus
}

func risikoBotEigene(g *RisikoSpiel, id string, k int) []string {
	var aus []string
	for _, n := range risikoGebietNach[id] {
		if g.Gebiete[n].Besitzer == k {
			aus = append(aus, n)
		}
	}
	return aus
}

// risikoBotDruck schätzt, wie stark ein Gebiet bedroht ist (feindliche Armeen
// ringsum minus eigene).
func risikoBotDruck(g *RisikoSpiel, id string, k int) int {
	druck := 0
	for _, n := range risikoGebietNach[id] {
		if g.Gebiete[n].Besitzer != k {
			druck += g.Gebiete[n].Armeen
		}
	}
	return druck - g.Gebiete[id].Armeen
}

// risikoBotWert bewertet ein Zielgebiet für einen Angriff.
func risikoBotWert(g *RisikoSpiel, von, nach string, k int, gew risikoBotGewicht) float64 {
	def := risikoGebieteDef(nach)
	wert := 1.0
	if def.Kont != "" {
		gesamt, eigen := 0, 0
		for _, gd := range risikoGebiete {
			if gd.Kont == def.Kont {
				gesamt++
				if g.Gebiete[gd.ID].Besitzer == k {
					eigen++
				}
			}
		}
		if eigen == gesamt-1 {
			wert += float64(risikoKontinente[def.Kont].Bonus) * 0.45 * gew.kontinent
			if g.Modus == "kontinent" {
				wert += 2 * gew.kontinent
			}
		} else if eigen >= gesamt-2 {
			wert += float64(risikoKontinente[def.Kont].Bonus) * 0.15 * gew.kontinent
		}
	}
	if g.Modus == "hauptstadt" {
		for _, hs := range g.Hauptstaedte {
			if hs == nach {
				wert += 3 * gew.kontinent
			}
		}
	}
	if g.Modus == "mission" {
		if m := g.Missionen[k]; m.Art == "kontinente" && def.Kont != "" {
			for _, kont := range m.Kont {
				if kont == def.Kont {
					wert += 2
				}
			}
		}
	}
	// Schwache Nachbarn sind lohnender als starke.
	wert += gew.expansion * (2.0 / float64(1+g.Gebiete[nach].Armeen))
	// Ziele, die selbst unter Druck stehen, sind leichter zu halten.
	_ = von
	return wert
}

// risikoBotAngriff wählt einen Angriff oder meldet keinen.
func risikoBotAngriff(g *RisikoSpiel, k int, r *rand.Rand) (string, string, bool) {
	sp := g.Spieler[k]
	gew := risikoBotGewichte(sp.Stufe, sp.Wesen)
	type option struct {
		von, nach string
		note      float64
	}
	var optionen []option
	for _, von := range risikoGebieteVon(g, k) {
		a := g.Gebiete[von].Armeen
		if a < 2 || a <= gew.garrison {
			continue
		}
		for _, nach := range risikoBotFeinde(g, von, k) {
			d := g.Gebiete[nach].Armeen
			p := risikoSiegChance(a, d)
			if p <= 0.01 {
				continue
			}
			wert := risikoBotWert(g, von, nach, k, gew)
			verlust := (1 - p) * float64(a-1) * 0.9 * gew.risiko
			note := p*(wert+float64(d)*0.35) - verlust
			// Eigene Grenze danach nicht nackt lassen.
			if a-1 <= gew.garrison {
				note -= 0.4 * gew.verteidigen
			}
			if risikoBotDruck(g, nach, k) > 6 {
				note += 0.15 // viel Gegenwehr erwartet
			}
			optionen = append(optionen, option{von: von, nach: nach, note: note})
		}
	}
	if len(optionen) == 0 {
		return "", "", false
	}

	switch risikoBotStufe(sp.Stufe) {
	case "leicht":
		// Nimmt eine der mittleren Optionen, lässt auch mal eine Runde aus.
		if r.Intn(100) < 35 {
			return "", "", false
		}
		o := optionen[r.Intn(len(optionen))]
		if p := risikoSiegChance(g.Gebiete[o.von].Armeen, g.Gebiete[o.nach].Armeen); p < 0.35 {
			return "", "", false
		}
		return o.von, o.nach, true
	}

	// mittel/schwer/experte: beste Note, Schwellen je Stufe.
	beste := optionen[0]
	for _, o := range optionen[1:] {
		if o.note > beste.note {
			beste = o
		}
	}
	schwelle := 0.25
	if risikoBotStufe(sp.Stufe) == "schwer" {
		schwelle = 0.35
	}
	if risikoBotStufe(sp.Stufe) == "experte" {
		schwelle = 0.40
		if beste.note < schwelle && len(optionen) > 1 {
			// Zweitbeste probieren, wenn die erste zu knapp ist.
			for _, o := range optionen {
				if o.von != beste.von || o.nach != beste.nach {
					if o.note >= schwelle {
						beste = o
						break
					}
				}
			}
		}
	}
	if beste.note < schwelle {
		return "", "", false
	}
	if risikoBotStufe(sp.Stufe) == "experte" {
		// Grobe Gegenrechnung: lohnt das Ziel auch nach einem Gegenangriff?
		probe := risikoKopie(g)
		probe.Gebiete[beste.nach].Besitzer = k
		probe.Gebiete[beste.nach].Armeen = 1
		gegen := 0.0
		for _, n := range risikoGebietNach[beste.nach] {
			if probe.Gebiete[n].Besitzer != k && probe.Gebiete[n].Armeen >= 3 {
				gegen += risikoSiegChance(probe.Gebiete[n].Armeen, 1)
			}
		}
		if gegen >= 2.2 {
			return "", "", false
		}
	}
	return beste.von, beste.nach, true
}

// risikoBotVerstaerkungZiel wählt ein eigenes Gebiet für eine Armee.
func risikoBotVerstaerkungZiel(g *RisikoSpiel, k int, r *rand.Rand) string {
	sp := g.Spieler[k]
	gew := risikoBotGewichte(sp.Stufe, sp.Wesen)
	eigene := risikoGebieteVon(g, k)
	if len(eigene) == 0 {
		return ""
	}
	bestes, besteNote := eigene[0], -1e9
	for _, id := range eigene {
		note := float64(risikoBotDruck(g, id, k))*1.2*gew.verteidigen + float64(len(risikoBotFeinde(g, id, k)))*0.6
		if id == g.Hauptstaedte[k] {
			note += 3
		}
		if g.Phase == risikoPhaseAufstellen && len(risikoBotFeinde(g, id, k)) > 0 {
			note += 1.5
		}
		def := risikoGebieteDef(id)
		if def.Kont != "" {
			gesamt, eigen := 0, 0
			for _, gd := range risikoGebiete {
				if gd.Kont == def.Kont {
					gesamt++
					if g.Gebiete[gd.ID].Besitzer == k {
						eigen++
					}
				}
			}
			if eigen == gesamt-1 {
				note += 3 * gew.kontinent
			}
		}
		if g.Gebiete[id].Armeen <= 1 {
			note += 1.5
		}
		if risikoBotStufe(sp.Stufe) == "leicht" {
			note += r.Float64() * 3
		}
		if note > besteNote {
			bestes, besteNote = id, note
		}
	}
	return bestes
}

// risikoBotKartenSatz sucht einen tauschbaren Satz, wenn es sich lohnt.
func risikoBotKartenSatz(g *RisikoSpiel, k int, r *rand.Rand) []int {
	hand := g.Spieler[k].Hand
	if len(hand) < 3 {
		return nil
	}
	for i := 0; i < len(hand); i++ {
		for j := i + 1; j < len(hand); j++ {
			for m := j + 1; m < len(hand); m++ {
				if risikoIstSatz([]RisikoKarte{hand[i], hand[j], hand[m]}) {
					// Leicht tauscht auch mal unnötig früh, die anderen erst,
					// wenn der Satzwert gut ist oder die Hand voll wird.
					if risikoBotStufe(g.Spieler[k].Stufe) == "leicht" {
						if r.Intn(100) < 70 {
							return []int{i, j, m}
						}
						continue
					}
					if g.KartenSatz >= 1 || len(hand) >= 5 || risikoSatzWert(g) >= 10 {
						return []int{i, j, m}
					}
				}
			}
		}
	}
	return nil
}

// risikoBotSchritt wählt genau eine Aktion für die aktuelle Lage.
func risikoBotSchritt(g *RisikoSpiel, k int, r *rand.Rand) (RisikoAktion, bool) {
	if g.Status != risikoStatusAktiv || g.AmZug != k || g.Spieler[k].Ausgeschieden {
		return RisikoAktion{}, false
	}
	switch g.Phase {
	case risikoPhaseAufstellen:
		if g.Spieler[k].StartPool <= 0 {
			return RisikoAktion{Art: "aufstellenFertig"}, true
		}
		ziel := risikoBotVerstaerkungZiel(g, k, r)
		if ziel == "" {
			return RisikoAktion{}, false
		}
		return RisikoAktion{Art: "setzen", Gebiet: ziel}, true

	case risikoPhaseVerstaerken:
		if g.Pool > 0 {
			ziel := risikoBotVerstaerkungZiel(g, k, r)
			if ziel == "" {
				return RisikoAktion{}, false
			}
			return RisikoAktion{Art: "setzen", Gebiet: ziel}, true
		}
		if karten := risikoBotKartenSatz(g, k, r); karten != nil {
			return RisikoAktion{Art: "karten", Karten: karten}, true
		}
		return RisikoAktion{Art: "verstaerkenEnde"}, true

	case risikoPhaseAngreifen:
		if g.Nachruecken != nil {
			return RisikoAktion{Art: "nachruecken", Anzahl: g.Nachruecken.Max}, true
		}
		if von, nach, ok := risikoBotAngriff(g, k, r); ok {
			return RisikoAktion{Art: "angriff", Von: von, Nach: nach}, true
		}
		return RisikoAktion{Art: "angriffEnde"}, true

	case risikoPhaseVerschieben:
		if !g.Verschoben {
			if von, nach, anzahl, ok := risikoBotVerschieben(g, k); ok {
				return RisikoAktion{Art: "verschieben", Von: von, Nach: nach, Anzahl: anzahl}, true
			}
		}
		return RisikoAktion{Art: "zugEnde"}, true
	}
	return RisikoAktion{}, false
}

// risikoBotVerschieben bringt Armeen von hinten an die bedrohte Front.
func risikoBotVerschieben(g *RisikoSpiel, k int) (string, string, int, bool) {
	sp := g.Spieler[k]
	if sp.Stufe == "leicht" {
		return "", "", 0, false
	}
	gew := risikoBotGewichte(sp.Stufe, sp.Wesen)
	besterWert := 0.0
	var von, nach string
	for _, id := range risikoGebieteVon(g, k) {
		if len(risikoBotFeinde(g, id, k)) > 0 || g.Gebiete[id].Armeen < 3 {
			continue
		}
		for _, ziel := range risikoBotEigene(g, id, k) {
			if len(risikoBotFeinde(g, ziel, k)) == 0 {
				continue
			}
			druck := risikoBotDruck(g, ziel, k)
			wert := float64(druck)*gew.verteidigen - float64(g.Gebiete[ziel].Armeen)*0.4
			if wert > besterWert {
				besterWert, von, nach = wert, id, ziel
			}
		}
	}
	if von == "" {
		return "", "", 0, false
	}
	anzahl := g.Gebiete[von].Armeen - 1
	if anzahl > 8 {
		anzahl = 8
	}
	if anzahl < 1 {
		return "", "", 0, false
	}
	return von, nach, anzahl, true
}

// risikoAktionAusfuehren wendet eine Bot-Aktion mit denselben Regeln wie für
// Menschen an.
func risikoAktionAusfuehren(g *RisikoSpiel, k int, a RisikoAktion, r *rand.Rand) error {
	switch a.Art {
	case "setzen":
		return risikoSetzeArmee(g, k, a.Gebiet)
	case "aufstellenFertig":
		return risikoAufstellenFertig(g, k)
	case "karten":
		_, err := risikoKartenTauschen(g, k, a.Karten)
		return err
	case "verstaerkenEnde":
		return risikoEndeVerstaerken(g, k)
	case "angriff":
		_, err := risikoAngriff(g, k, a.Von, a.Nach, r)
		return err
	case "nachruecken":
		return risikoNachrueckenArmeen(g, k, a.Anzahl)
	case "angriffEnde":
		return risikoEndeAngriff(g, k)
	case "verschieben":
		return risikoVerschiebe(g, k, a.Von, a.Nach, a.Anzahl)
	case "zugEnde":
		return risikoZugEnde(g, k)
	}
	return nil
}

// risikoBotPlanen spielt einen vollständigen Zug auf einer Kopie durch und
// liefert die Schritte. Die echte Partie bleibt unberührt.
func risikoBotPlanen(g *RisikoSpiel, k int, r *rand.Rand) []RisikoAktion {
	probe := risikoKopie(g)
	var plan []RisikoAktion
	for i := 0; i < 500; i++ {
		if probe.Status != risikoStatusAktiv || probe.AmZug != k || probe.Spieler[k].Ausgeschieden {
			break
		}
		aktion, ok := risikoBotSchritt(probe, k, r)
		if !ok {
			break
		}
		plan = append(plan, aktion)
		if err := risikoAktionAusfuehren(probe, k, aktion, r); err != nil {
			break
		}
		if aktion.Art == "zugEnde" || aktion.Art == "aufstellenFertig" {
			if probe.AmZug != k {
				break
			}
		}
	}
	return plan
}
