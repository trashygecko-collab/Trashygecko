package main

import (
	"math/rand"
	"time"
)

/* ===========================================================================
   BOT FÜR MENSCH ÄRGERE DICH NICHT

   Der Bot würfelt und zieht über dieselben Funktionen wie ein Mensch.

   Stufen:
     leicht  – nimmt zu 40 % irgendeine erlaubte Figur.
     mittel  – schlägt, wenn es geht, bringt Figuren ins Ziel und räumt das
               Startfeld.
     schwer  – wie mittel; zusätzlich meidet er Felder, auf denen der Gegner
               im nächsten Wurf schlagen könnte.
   =========================================================================== */

func maenBotSitz(g *MenschaergerGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return g.seatOf(g.BotID)
}

func maenBotTreiben(g *MenschaergerGame, jetzt time.Time) bool {
	sitz := maenBotSitz(g)
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
	if g.Dice == 0 {
		maenRoll(g)
		return true
	}
	figur, ok := maenBotFigur(g, sitz, g.BotStufe)
	if !ok {
		// Kein Zug möglich: wie im Handler den Zug abgeben.
		maenNextPlayer(g)
		return false
	}
	player := &g.Players[sitz]
	notiz := maenApplyMove(g, figur)
	if maenAllInGoal(player.Pieces) {
		g.Status = "finished"
		g.Winner = player.ID
		g.Dice = 0
		g.ExtraRoll = false
		g.Note = player.Name + " bringt alle Figuren ins Ziel!"
		return true
	}
	sechs := g.Dice == 6
	g.Dice = 0
	if sechs {
		g.ExtraRoll = true
		g.RollAttempts = 0
		g.Note = "Sechs! Noch einmal würfeln."
	} else {
		maenNextPlayer(g)
		g.Note = notiz
	}
	return true
}

// maenRoll ist der Würfelwurf aus dem Handler – auch für den Bot.
func maenRoll(g *MenschaergerGame) {
	player := &g.Players[g.Current]
	dice := rand.Intn(6) + 1
	g.Dice = dice
	g.LastDice = dice
	g.Note = ""

	onTrack, laneCanAdvance := false, false
	for _, progress := range player.Pieces {
		if progress >= 0 && progress < 40 {
			onTrack = true
		}
		if progress >= 40 && progress < 43 {
			laneCanAdvance = true
		}
	}
	if dice != 6 && !onTrack && !laneCanAdvance && !g.ExtraRoll {
		g.RollAttempts++
		if g.RollAttempts < 3 {
			g.Dice = 0
			g.Note = "Keine Sechs! Noch ein Wurfversuch."
			return
		}
		g.Dice = 0
		maenNextPlayer(g)
		g.Note = "Dreimal keine Sechs – Spielerwechsel …"
		return
	}
	if dice == 6 {
		g.RollAttempts = 0
	}
	if len(maenLegalMoves(player.Pieces, player.Color, dice)) == 0 {
		g.Dice = 0
		maenNextPlayer(g)
		g.Note = "Kein gültiger Zug möglich – nächster Spieler …"
	}
}

func maenBotFigur(g *MenschaergerGame, sitz int, stufe string) (int, bool) {
	player := g.Players[sitz]
	erlaubt := maenLegalMoves(player.Pieces, player.Color, g.Dice)
	if len(erlaubt) == 0 {
		return 0, false
	}
	if botStufe(stufe) == botLeicht && botZufall().Intn(100) < 40 {
		return erlaubt[botZufall().Intn(len(erlaubt))], true
	}
	beste := erlaubt[0]
	besterWert := -1 << 30
	for _, figur := range erlaubt {
		wert := maenBotWert(g, sitz, figur, botStufe(stufe) == botSchwer)
		if wert > besterWert || (wert == besterWert && botZufall().Intn(2) == 0) {
			beste, besterWert = figur, wert
		}
	}
	return beste, true
}

func maenBotWert(g *MenschaergerGame, sitz, figur int, vorsichtig bool) int {
	player := g.Players[sitz]
	fortschritt := player.Pieces[figur]
	ziel := 0
	if fortschritt >= 0 {
		ziel = fortschritt + g.Dice
	}
	wert := ziel * 2
	if fortschritt == -1 {
		wert += 55 // aus dem Haus kommen ist viel wert
	}
	if ziel > 39 {
		wert += 220 // ins Ziel
	}
	if g.Dice == 6 && fortschritt == 0 {
		wert += 25 // Startfeld räumen
	}

	if ziel <= 39 {
		feld := (maenStartIndex[player.Color] + ziel) % 40
		for i := range g.Players {
			if i == sitz {
				continue
			}
			for _, other := range g.Players[i].Pieces {
				if other < 0 || other > 39 {
					continue
				}
				if (maenStartIndex[g.Players[i].Color]+other)%40 == feld {
					wert += 300 // schlagen
				}
			}
		}
		if vorsichtig {
			// Steht der Gegner 1–6 Felder dahinter, ist das Feld gefährlich.
			for i := range g.Players {
				if i == sitz {
					continue
				}
				gegnerStart := maenStartIndex[g.Players[i].Color]
				for _, other := range g.Players[i].Pieces {
					if other < 0 || other > 39 {
						continue
					}
					abstand := (feld - (gegnerStart+other)%40 + 40) % 40
					if abstand >= 1 && abstand <= 6 {
						wert -= 35
					}
				}
			}
		}
	}
	return wert
}
