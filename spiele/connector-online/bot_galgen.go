package main

/* ===========================================================================
   BOT FÜR GALGENMÄNNCHEN

   Der Bot ist ein normaler Mitspieler und rät über galgenAktion – dieselbe
   Prüfung wie der Mensch (Reihe, Doppelungen, Treffer, Leben, Rundenende).
   Er sieht nur, was auf dem Schirm steht: die Länge, die aufgedeckten
   Buchstaben und die schon geratenen. Das Wort selbst liest er nicht.

   Stufen:
     leicht  – rät zufällige Buchstaben.
     mittel  – rät nach deutscher Buchstabenhäufigkeit.
     schwer  – vergleicht das Muster mit der eingebauten Wortliste und rät den
               Buchstaben, der in den noch passenden Wörtern am häufigsten
               fehlt.
   =========================================================================== */

import (
	"math/rand"
	"strings"
	"time"
)

// galgenHaeufig ist die Reihenfolge der deutschen Buchstaben nach Häufigkeit.
const galgenHaeufig = "enisratdhulcgmobwfkzpvjyxq"

var galgenAlphabet = []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
	"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
}

// galgenBotSitz ist der Platz des Bots (-1, wenn keiner mitspielt).
func galgenBotSitz(g *GalgenGame) int {
	if g == nil || g.BotID == "" {
		return -1
	}
	return g.galgenSitz(g.BotID)
}

// galgenBotRang ist die Häufigkeitsstufe eines Buchstabens (klein = häufig).
func galgenBotRang(b string) int {
	if i := strings.Index(galgenHaeufig, b); i >= 0 {
		return i
	}
	return len(galgenHaeufig)
}

// galgenBotBuchstabe wählt den nächsten Ratebuchstaben.
func galgenBotBuchstabe(g *GalgenGame) string {
	belegt := map[string]bool{}
	for _, b := range g.Buchstaben {
		belegt[b] = true
	}
	frei := []string{}
	for _, b := range galgenAlphabet {
		if !belegt[b] {
			frei = append(frei, b)
		}
	}
	if len(frei) == 0 {
		return ""
	}
	switch botStufe(g.BotStufe) {
	case botLeicht:
		return frei[rand.Intn(len(frei))]
	case botMittel:
		for _, b := range []rune(galgenHaeufig) {
			if !belegt[string(b)] {
				return string(b)
			}
		}
		return frei[0]
	}

	// Schwer: nur Wörter, die zum Muster passen und keine Fehlversuche enthalten.
	falsch := map[string]bool{}
	for _, b := range g.Buchstaben {
		gefunden := false
		for _, a := range g.Anzeige {
			if strings.ToLower(a) == b {
				gefunden = true
				break
			}
		}
		if !gefunden {
			falsch[b] = true
		}
	}
	zaehler := map[string]int{}
	for _, w := range galgenListe {
		r := []rune(strings.ToLower(w.Wort))
		if len(r) != len(g.Anzeige) {
			continue
		}
		ok := true
		for i, a := range g.Anzeige {
			if a != "" && strings.ToLower(a) != string(r[i]) {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		for _, rn := range r {
			if falsch[string(rn)] {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		for i, rn := range r {
			b := string(rn)
			if g.Anzeige[i] != "" || belegt[b] {
				continue
			}
			zaehler[b]++
		}
	}
	best := ""
	for b, n := range zaehler {
		if n > zaehler[best] || (n == zaehler[best] && galgenBotRang(b) < galgenBotRang(best)) {
			best = b
		}
	}
	if best != "" {
		return best
	}
	for _, b := range []rune(galgenHaeufig) {
		if !belegt[string(b)] {
			return string(b)
		}
	}
	return frei[0]
}

// galgenBotTreiben lässt den Bot nach der Denkpause raten oder die nächste
// Runde starten. Aufgerufen wird es beim Abholen des Spielstands (GET).
func galgenBotTreiben(g *GalgenGame, jetzt time.Time) bool {
	if g == nil || g.BotID == "" {
		return false
	}
	sitz := galgenBotSitz(g)
	if sitz < 0 {
		return false
	}
	switch g.Status {
	case "spiel":
		if g.Zug != sitz {
			return false
		}
	case "ende":
		// nächste Runde anstoßen
	default:
		return false
	}
	if g.BotFrist.IsZero() {
		warte := botWartezeit(g.BotStufe)
		if g.Status == "ende" {
			warte *= 2
		}
		g.BotFrist = jetzt.Add(warte)
		return false
	}
	if jetzt.Before(g.BotFrist) {
		return false
	}
	g.BotFrist = time.Time{}
	if g.Status == "ende" {
		status, _ := g.galgenNaechste(sitz)
		return status == 0
	}
	b := galgenBotBuchstabe(g)
	if b == "" {
		return false
	}
	status, _ := g.galgenAktion(sitz, b)
	return status == 0
}
