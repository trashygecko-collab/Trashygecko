package main

/* ===========================================================================
   BOT FÜR „1×1-RENNEN“

   Alle am Tisch bekommen dieselben Aufgaben (Grundrechenarten, Gleichungen,
   Zahlen merken oder Zahlenketten) und tippen die Lösung. Der Bot rechnet
   wie ein Mensch: er kennt die Lösung der Aufgabe, die gerade auf dem Tisch
   liegt, und schickt sie durch dieselbe Prüfung wie ein Mensch
   (rechenAntwort). Die Stufe entscheidet nur, wie lange er überlegt und wie
   oft er sich vertut – leicht braucht am längsten und patzt am häufigsten.

   Der Antrieb läuft über botFrist: GET auf /api/einsxeins ruft
   rechenBotTreiben; ist die Denkpause vorbei, antwortet der Bot genau eine
   Aufgabe.
   =========================================================================== */

import (
	"math/rand"
	"time"
)

// rechenBotDenkzeit ist die Pause vor einer Antwort. Sie hängt an der Stufe
// und streut, damit sich der Bot nicht im Takt anfühlt.
func rechenBotDenkzeit(stufe string) time.Duration {
	base := 1500 * time.Millisecond
	streu := 900 * time.Millisecond
	switch botStufe(stufe) {
	case botLeicht:
		base, streu = 2600*time.Millisecond, 1400*time.Millisecond
	case botSchwer:
		base, streu = 750*time.Millisecond, 550*time.Millisecond
	}
	// Der Zufall des Pakets (beim Start gesät) statt botZufall: dessen Quelle
	// ist die Uhr, und unter Windows tickt sie nur im Millisekundentakt – acht
	// Antworten hintereinander bekämen sonst denselben Wurf.
	return base + time.Duration(rand.Int63n(int64(streu)))
}

// rechenBotFehlerquote sagt, wie oft der Bot daneben tippt (in Prozent).
func rechenBotFehlerquote(stufe string) int {
	switch botStufe(stufe) {
	case botLeicht:
		return 28
	case botSchwer:
		return 2
	}
	return 9
}

// rechenBotSitz liefert den Platz des Bots (-1, wenn keiner mitspielt).
func rechenBotSitz(g *RechenGame) int {
	if g == nil || len(g.Bots) == 0 {
		return -1
	}
	return rechenSeat(g, g.Bots[0])
}

// rechenBotAuftrag sagt, welche Antwort der Bot jetzt abgeben würde:
// (Aufgabennummer, Wert, ok). Vertippt er sich, weicht der Wert um 1 bis 9 ab
// – so daneben, wie ein Mensch sich vertut, nicht zufällig weit weg.
func rechenBotAuftrag(g *RechenGame) (int, int, bool) {
	if g == nil || g.Status != "active" {
		return 0, 0, false
	}
	sitz := rechenBotSitz(g)
	if sitz < 0 {
		return 0, 0, false
	}
	p := g.Players[sitz]
	if p.Done || p.Index >= len(g.Tasks) {
		return 0, 0, false
	}

	aufgabe := g.Tasks[p.Index]
	wert := aufgabe.Answer
	if rand.Intn(100) < rechenBotFehlerquote(g.BotStufe) {
		abweichung := rand.Intn(9) + 1
		if rand.Intn(2) == 0 {
			abweichung = -abweichung
		}
		wert = aufgabe.Answer + abweichung
		if wert == aufgabe.Answer {
			wert++
		}
	}
	return p.Index, wert, true
}

// rechenBotTreiben lässt den Bot nach der Denkpause antworten.
func rechenBotTreiben(room *Room, jetzt time.Time) bool {
	g := room.Rechen
	if g == nil {
		return false
	}
	if _, _, ok := rechenBotAuftrag(g); !ok {
		g.BotFrist = time.Time{}
		return false
	}
	if g.BotFrist.IsZero() {
		g.BotFrist = jetzt.Add(rechenBotDenkzeit(g.BotStufe))
		return false
	}
	if jetzt.Before(g.BotFrist) {
		return false
	}
	if !rechenBotSchritt(room) {
		return false
	}
	g.BotFrist = jetzt.Add(rechenBotDenkzeit(g.BotStufe))
	return true
}

// rechenBotSchrittWert gibt genau eine Antwort ab (dieselbe Prüfung wie für
// einen Menschen) und meldet den getippten Wert und ob er stimmte.
func rechenBotSchrittWert(room *Room) (int, bool, bool) {
	g := room.Rechen
	index, wert, ok := rechenBotAuftrag(g)
	if !ok {
		return 0, false, false
	}
	zug := rechenAntwort(room, g, g.Bots[0], index, wert)
	return wert, zug.Richtig, zug.Meldung == ""
}

// rechenBotSchritt ist dasselbe ohne Rückmeldung.
func rechenBotSchritt(room *Room) bool {
	_, _, ok := rechenBotSchrittWert(room)
	return ok
}
