package main

import (
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

/* ===========================================================================
   ZENTRALE BOT-VERWALTUNG

   Ein Bot ist auf dem Server ein ganz normaler Mitspieler mit einer Kennung
   („bot-1“ …), einem erkennbaren Namen samt Avatar und einer Stufe. Er sieht
   nur den Spielstand, den auch ein Mensch sieht, und schickt seine Züge durch
   dieselbe Prüfung wie ein Mensch – schummeln kann er also nicht.

   Was hier steht, gilt für alle Spiele:
     · Namen (botName: zufällig je Partie, ohne Dopplung),
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

// botSpitznamen sind die Namen, aus denen jeder Computergegner seinen
// bekommt. Die Liste ist die einzige Quelle – ein neues Spiel erbt die
// Vergabe ohne Zusatzarbeit, weil jeder Bot seinen Namen über botName holt.
var botSpitznamen = []string{
	"Klappstuhl", "Denkpause", "Toastbrot", "BinGleichWeg", "WarIchNicht",
	"KeinPlan", "FastWach", "Huch", "Gießkanne", "GuckNurZu",
	"Blumentopf", "BisschenMüde", "Büroklammer", "SendHelp", "Oopsie",
	"Kunibert", "Rüdiger", "Gottfried", "Detlef", "Waltraud",
	"Brunhilde", "Kunigunde", "Gertrud", "Pumpernickel", "Bimmel",
	"Schimmelpfennig", "Krawinkel", "Mumpitz", "Kevin", "Chantal",
	"Jacqueline", "Bärbel", "Horst", "Karlheinz", "Herbert",
	"Dieter", "Willi", "Hannelore", "Anneliese", "Ingeborg",
	"Brigitte", "Hanspeter", "Klausdieter", "Annegret", "Lieselotte",
	"Sepp", "Schorsch", "Vroni", "Resi", "Zensi",
	"Traudl", "Frauke", "Poppo", "Busenwunder", "Kurvenwunder",
	"Schürzenjäger", "Lustmolch", "Fummel-Vroni", "Nippel-Norbert", "Betthupferl",
	"Pralle-Resi", "Sitzpinkler", "Pantoffelheld", "Schmuddelkind", "Lustknabe",
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

// botName liefert den Namen des n-ten Bots einer Partie (ab 1).
//
// Der Name wird beim ersten Aufruf ausgelost und liegt danach am Raum: ein
// Neuladen, eine Wiederverbindung oder ein späterer Zug ändert ihn nicht
// mehr. Vor einer neuen Partie ruft „action=bot“ botNamenNeu – dann werden
// alle Namen frisch ausgelost. Innerhalb einer Partie kommt kein Name doppelt
// vor, und kein Bot bekommt den Namen eines Menschen, der im Raum sitzt.
func (r *Room) botName(n int) string {
	if n < 1 {
		n = 1
	}
	if r == nil {
		return botSpitznamen[n-1]
	}
	if r.botNamen == nil {
		r.botNamen = make(map[string]string, 4)
	}
	schluessel := strconv.Itoa(n)
	if name, da := r.botNamen[schluessel]; da {
		return name
	}
	name := r.botNameAuslosen()
	r.botNamen[schluessel] = name
	return name
}

// botNamenNeu verwirft die Namen der letzten Partie. Jede neue Partie würfelt
// damit frisch, während eine laufende ihre Namen behält.
func (r *Room) botNamenNeu() {
	if r == nil {
		return
	}
	r.botNamen = make(map[string]string, 4)
}

// botNameAuslosen zieht einen Namen, der in dieser Partie noch frei ist und
// den kein Mensch im Raum trägt. Jeder freie Name hat dieselbe Chance.
func (r *Room) botNameAuslosen() string {
	belegt := make(map[string]bool, len(r.botNamen)+len(r.Players))
	for _, name := range r.botNamen {
		belegt[name] = true
	}
	for _, p := range r.Players {
		belegt[p.Name] = true
	}
	frei := make([]string, 0, len(botSpitznamen))
	for _, name := range botSpitznamen {
		if !belegt[name] {
			frei = append(frei, name)
		}
	}
	if len(frei) == 0 {
		frei = botSpitznamen
	}
	// rand.Intn teilt sich einen Zufall, der beim Start gesät wird – zwei
	// Namen kurz hintereinander sind damit wirklich unabhängig.
	return frei[rand.Intn(len(frei))]
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
// botSaat stellt den Bot-Zufall auf einen festen Wert (Tests). 0 heißt:
// echter Zufall. Der Zähler sorgt dafür, dass jeder Aufruf eine neue Zahl
// aus derselben festen Folge bekommt.
var botSaat atomic.Int64

var botZaehler atomic.Int64

func botZufall() *rand.Rand {
	if s := botSaat.Load(); s != 0 {
		return rand.New(rand.NewSource(s + botZaehler.Add(1)))
	}
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// botIstFerngesteuert erkennt die Kennungen, die als Bot gelten. So kann ein
// Spiel prüfen, ob eine Aktion von einem Bot kommt.
func botIstFerngesteuert(id string) bool {
	return strings.HasPrefix(id, "bot-")
}
