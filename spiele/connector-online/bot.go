package main

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

/* ===========================================================================
   ZENTRALE BOT-VERWALTUNG

   Ein Bot ist auf dem Server ein ganz normaler Mitspieler mit einer Kennung
   („bot-1“ …), einem erkennbaren Namen samt Avatar und einer Stufe. Er sieht
   nur den Spielstand, den auch ein Mensch sieht, und schickt seine Züge durch
   dieselbe Prüfung wie ein Mensch – schummeln kann er also nicht.

   Was hier steht, gilt für alle Spiele:
     · Namen und Avatare (botName),
     · die drei Stufen (botStufe, botStufeText),
     · die Denkpause (botWartezeit): kurz, aber nicht hektisch,
     · die Kennung (botID).

   Was ein Bot in einem bestimmten Spiel tut, steht in der Datei des Spiels
   (bot_tictactoe.go, bot_viergewinnt.go, bot_streichholz.go …). Jedes Spiel
   entscheidet selbst, welche Züge es kennt – die Mechanik bleibt dort, wo sie
   schon war.
   =========================================================================== */

const (
	botLeicht = "leicht"
	botMittel = "mittel"
	botSchwer = "schwer"
)

// botNamen sind die erkennbaren Namen. Der Avatar steht davor, damit ein Bot
// auch in einer Namensliste sofort auffällt.
var botNamen = []string{
	"🤖 Robi",
	"🤖 Nova",
	"🤖 Pixel",
	"🤖 Tessa",
	"🤖 Kalle",
	"🤖 Mira",
	"🤖 Bolt",
	"🤖 Fips",
}

// botStufe prüft die gewünschte Stufe. Unbekanntes heißt „mittel“ – so spielt
// ein Bot nie versehentlich in einer Stufe, die es nicht gibt.
func botStufe(wert string) string {
	switch strings.ToLower(strings.TrimSpace(wert)) {
	case botLeicht:
		return botLeicht
	case botSchwer:
		return botSchwer
	}
	return botMittel
}

// botStufeText ist die Anzeige zur Stufe.
func botStufeText(stufe string) string {
	switch botStufe(stufe) {
	case botLeicht:
		return "leicht"
	case botSchwer:
		return "schwer"
	}
	return "mittel"
}

// botName liefert den Namen für den n-ten Bot eines Tisches (ab 1).
func botName(n int) string {
	if n < 1 {
		n = 1
	}
	return botNamen[(n-1)%len(botNamen)]
}

// botID ist die Kennung des n-ten Bots. Sie steht fest, damit Züge und
// Spielstand nach einem Neuladen dieselben bleiben.
func botID(n int) string {
	if n < 1 {
		n = 1
	}
	return "bot-" + strconv.Itoa(n)
}

// botWartezeit ist die Denkpause vor einem Zug. Sie hängt an der Stufe und
// streut ein wenig, damit sich die Züge nicht im Takt anfühlen: lang genug,
// dass man den Zug sieht, kurz genug, dass niemand wartet.
func botWartezeit(stufe string) time.Duration {
	base := 620 * time.Millisecond
	switch botStufe(stufe) {
	case botLeicht:
		base = 800 * time.Millisecond
	case botSchwer:
		base = 520 * time.Millisecond
	}
	return base + time.Duration(botZufall().Int63n(int64(260*time.Millisecond)))
}

// botZufall ist der Zufall eines Bot-Zuges (Streuung bei gleichen Zügen).
func botZufall() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// botIstFerngesteuert erkennt die Kennungen, die als Bot gelten. So kann ein
// Spiel prüfen, ob eine Aktion von einem Bot kommt.
func botIstFerngesteuert(id string) bool {
	return strings.HasPrefix(id, "bot-")
}
