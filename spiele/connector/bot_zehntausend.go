package main

/* ===========================================================================
   BOT FÜR 10.000

   Der Bot nutzt dieselben Helfer wie ein Mensch: ztWuerfle, ztSichere,
   ztFehlwurf und ztBank. Er sieht nur die Würfel und sein Konto – schummeln
   kann er nicht.

   Er sichert alle Pasche ab drei gleichen sowie jede 1 und 5. Danach würfelt
   er den Rest weiter. Gespeichert wird, sobald die Stufe es erlaubt:
     leicht  – sofort ab 350 (jeder Fehlwurf wäre teurer als der Gewinn).
     mittel  – ab 600 oder wenn nur noch ein Würfel übrig ist.
     schwer  – ab 1.000, bei nur noch einem Würfel oder kurz vor dem Ziel.
   =========================================================================== */

import (
	"strconv"
	"time"
)

func ztBotSitz(g *ZehnTausendGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return ztFindPlayer(g, g.BotID)
}

func ztBotTreiben(g *ZehnTausendGame, jetzt time.Time) bool {
	sitz := ztBotSitz(g)
	if sitz < 0 || g.Status != "active" || g.Current != sitz {
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
	return ztBotZug(g, sitz)
}

// ztBotZug macht einen Schritt: ersten Wurf, weiterwürfeln oder sichern.
func ztBotZug(g *ZehnTausendGame, sitz int) bool {
	if g == nil || g.Status != "active" || sitz < 0 || sitz >= len(g.Players) || g.Current != sitz {
		return false
	}
	name := g.Players[sitz].Name
	if g.Wuerfe == 0 {
		geworfen := ztWuerfle(g)
		if ztFehlwurf(g, geworfen, name) {
			return true
		}
		g.Note = name + " würfelt: " + ztListe(geworfen) + ". " +
			strconv.Itoa(ztPunkte(geworfen)) + " Punkte sind möglich."
		return true
	}
	if g.ZugPunkte >= ztBotGrenze(g, sitz) {
		return ztBank(g, sitz)
	}
	_, alle := ztSichere(g, ztBotAuswahl(g))
	if alle {
		ztNaechsterWurf(g)
	}
	geworfen := ztWuerfle(g)
	if ztFehlwurf(g, geworfen, name) {
		return true
	}
	g.Note = name + " würfelt weiter: " + ztListe(geworfen) + ". Zug: " +
		strconv.Itoa(g.ZugPunkte) + " Punkte."
	return true
}

// ztBotGrenze ist der Punktestand, ab dem der Bot sichert.
func ztBotGrenze(g *ZehnTausendGame, sitz int) int {
	gesichert := 0
	for _, s := range g.Gesichert {
		if s {
			gesichert++
		}
	}
	if gesichert >= 5 || g.Players[sitz].Punkte+g.ZugPunkte >= ztZiel {
		return ztEinstieg
	}
	switch botStufe(g.BotStufe) {
	case botLeicht:
		return ztEinstieg
	case botSchwer:
		return 1000
	}
	return 600
}

// ztBotAuswahl nimmt alle Pasche ab drei gleichen sowie einzelne 1er und 5er.
// Ein vierter gleicher Würfel zählt dabei als eigener Satz – das erledigt
// ztSichere über die Gruppenrechnung.
func ztBotAuswahl(g *ZehnTausendGame) []bool {
	auswahl := make([]bool, 6)
	anzahl := [7]int{}
	for i, w := range g.Wuerfel {
		if i < len(g.Gesichert) && g.Gesichert[i] {
			continue
		}
		if w >= 1 && w <= 6 {
			anzahl[w]++
		}
	}
	for w := 1; w <= 6; w++ {
		if anzahl[w] < 3 {
			continue
		}
		for i, wurf := range g.Wuerfel {
			if i < len(g.Gesichert) && g.Gesichert[i] {
				continue
			}
			if wurf == w {
				auswahl[i] = true
			}
		}
	}
	for i, w := range g.Wuerfel {
		if i < len(g.Gesichert) && g.Gesichert[i] {
			continue
		}
		if w == 1 || w == 5 {
			auswahl[i] = true
		}
	}
	return auswahl
}
