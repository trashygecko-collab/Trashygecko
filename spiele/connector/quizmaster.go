package main

/* ===========================================================================
   DER QUIZMASTER – Kommentare zur Scherzshow

   Der Moderator spricht ausschließlich hier, also am Server. Jede Ansage steht
   danach im Spielstand (Feld "moderator") und ist damit für alle Spieler einer
   Partie dieselbe – kein Client würfelt eigene Sprüche.

   Auswahl: Jede Situation (Art) hat ihren eigenen Vorrat. Innerhalb des
   Vorrats werden zuletzt benutzte Texte gemieden (Fenster über
   quizModFenster Ansagen), damit sich nichts kurzfristig wiederholt.

   Platzhalter in den Texten:
     {name}    – die Hauptperson der Situation
     {andere}  – die zweite Person (Verfolger, altes Zugpferd, Verlierer)
     {punkte}  – Punkte mit Vorzeichen ("+100", "-25")
     {abstand} – Vorsprung als Zahl (der Text schreibt "Punkte" selbst dazu)
     {falsch}  – die tatsächlich gegebene falsche Antwort, im O-Ton zitiert
   =========================================================================== */

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// QuizAnsage ist das, was der Moderator gerade sagt.
type QuizAnsage struct {
	Name string `json:"name"`
	Text string `json:"text"`
	Art  string `json:"art"`
}

// quizModeratorName steht auf dem Schild am Tresen.
const quizModeratorName = "Kalle Klabauter"

// quizModFenster: so viele Ansagen lang wird ein Text nicht wiederholt.
const quizModFenster = 18

// quizModStand ist der Auswahlzustand einer Partie (nicht im JSON).
type quizModStand struct {
	rng    *rand.Rand
	letzte map[string]int
	tick   int
}

// quizModLage beschreibt die Situation, in die ein Spruch eingesetzt wird.
type quizModLage struct {
	Name    string
	Andere  string
	Punkte  string
	Falsch  string
	Abstand int
}

func (g *QuizGame) modStand() *quizModStand {
	if g.ModStand == nil {
		saat := time.Now().UnixNano()
		for i := 0; i < len(g.ID); i++ {
			saat += int64(g.ID[i]) * int64(31*(i+1))
		}
		g.ModStand = &quizModStand{rng: rand.New(rand.NewSource(saat)), letzte: map[string]int{}}
	}
	return g.ModStand
}

// quizModSage wählt einen Spruch der Art aus und legt ihn in den Spielstand.
func (g *QuizGame) quizModSage(art string, lage quizModLage) {
	texte := quizModTexte[art]
	if len(texte) == 0 {
		return
	}
	st := g.modStand()
	st.tick++
	frisch := make([]int, 0, len(texte))
	for i, text := range texte {
		if zuletzt, ok := st.letzte[text]; !ok || st.tick-zuletzt > quizModFenster {
			frisch = append(frisch, i)
		}
	}
	if len(frisch) == 0 {
		for i := range texte {
			frisch = append(frisch, i)
		}
	}
	text := texte[frisch[st.rng.Intn(len(frisch))]]
	st.letzte[text] = st.tick
	g.Moderator = &QuizAnsage{
		Name: quizModeratorName,
		Text: quizModFuelle(text, lage),
		Art:  art,
	}
}

// quizModFuelle setzt die Platzhalter ein. Fehlt eine Angabe, steht dort ein
// neutraler Ersatz – lieber ungenau als eine Lücke im Satz.
func quizModFuelle(text string, lage quizModLage) string {
	name := strings.TrimSpace(lage.Name)
	andere := strings.TrimSpace(lage.Andere)
	punkte := strings.TrimSpace(lage.Punkte)
	falsch := strings.TrimSpace(lage.Falsch)
	if name == "" {
		name = "wer auch immer"
	}
	if andere == "" {
		andere = "der Rest"
	}
	if punkte == "" {
		punkte = "nichts"
	}
	if falsch == "" {
		falsch = "irgendwas"
	}
	return strings.NewReplacer(
		"{name}", name,
		"{andere}", andere,
		"{punkte}", punkte,
		"{falsch}", falsch,
		"{abstand}", strconv.Itoa(lage.Abstand),
	).Replace(text)
}

// quizSpitze nennt alle Sitze mit der höchsten Punktzahl.
func (g *QuizGame) quizSpitze() []int {
	oben := make([]int, 0, len(g.Spieler))
	best := 0
	for i := range g.Spieler {
		if len(oben) == 0 || g.Spieler[i].Punkte > best {
			oben = oben[:0]
			best = g.Spieler[i].Punkte
		}
		if g.Spieler[i].Punkte == best {
			oben = append(oben, i)
		}
	}
	return oben
}

// quizFuehrer nennt die alleinige Spitze oder "" (Gleichstand/niemand).
func (g *QuizGame) quizFuehrer() string {
	if oben := g.quizSpitze(); len(oben) == 1 {
		return g.Spieler[oben[0]].ID
	}
	return ""
}

func (g *QuizGame) quizName(id string) string {
	if sitz := g.quizSitz(id); sitz >= 0 {
		return g.Spieler[sitz].Name
	}
	return ""
}

// quizPunkteText schreibt Punkte mit Vorzeichen.
func quizPunkteText(wert int) string {
	if wert > 0 {
		return "+" + strconv.Itoa(wert)
	}
	return strconv.Itoa(wert)
}

// quizModKandidat ist eine Situation, die zu dieser Runde passt.
type quizModKandidat struct {
	art  string
	lage quizModLage
}

// quizModeratorRunde wählt nach der Auswertung einer Frage den passenden
// Kommentar. Es werden alle passenden Situationen gesammelt; den Zuschlag
// bekommt die erste, die nicht schon in der Vorrunde dran war – so wechselt
// der Ton, statt ewig dieselbe Leier zu spielen.
func (g *QuizGame) quizModeratorRunde(fuehrerVorher string, punkteVorher map[string]int) {
	// Die Art der Vorrunde steht im Spielstand – kein zweites Gedächtnis dafür.
	vermeiden := ""
	if g.Moderator != nil {
		vermeiden = g.Moderator.Art
	}
	art, lage := g.quizModLageRunde(fuehrerVorher, punkteVorher, vermeiden)
	g.quizModSage(art, lage)
}

// quizModLageRunde baut die Kandidatenliste der Runde. vermeiden ist die Art
// der Vorrunde – sie wird nur genommen, wenn es keine Alternative gibt.
func (g *QuizGame) quizModLageRunde(fuehrerVorher string, punkteVorher map[string]int, vermeiden string) (string, quizModLage) {
	nachher := g.quizFuehrer()
	kandidaten := make([]quizModKandidat, 0, 4)
	merke := func(art string, lage quizModLage) {
		kandidaten = append(kandidaten, quizModKandidat{art, lage})
	}

	// 1. Jemand ist in dieser Runde neu ins Minus gerutscht.
	frischUnten := -1
	for i := range g.Spieler {
		s := &g.Spieler[i]
		vorher, bekannt := punkteVorher[s.ID]
		if !bekannt || vorher < 0 || s.Punkte >= 0 || s.PunkteRunde >= 0 {
			continue
		}
		if frischUnten < 0 || s.Punkte < g.Spieler[frischUnten].Punkte {
			frischUnten = i
		}
	}
	if frischUnten >= 0 {
		merke("grottig", quizModLage{
			Name:   g.Spieler[frischUnten].Name,
			Punkte: quizPunkteText(g.Spieler[frischUnten].Punkte),
		})
	}

	// 2. Führungswechsel.
	if fuehrerVorher != "" && nachher != "" && fuehrerVorher != nachher {
		alt, neu := g.quizSitz(fuehrerVorher), g.quizSitz(nachher)
		if alt >= 0 && neu >= 0 {
			merke("fuehrung", quizModLage{
				Name:    g.Spieler[neu].Name,
				Andere:  g.Spieler[alt].Name,
				Abstand: g.Spieler[neu].Punkte - g.Spieler[alt].Punkte,
			})
		}
	}

	// 3. Niemand hat geantwortet – das ist eine eigene Situation.
	niemand := true
	for i := range g.Spieler {
		if g.Spieler[i].Antwort != nil {
			niemand = false
		}
	}
	if niemand {
		merke("leer", quizModLage{})
	} else if nachher == "" {
		// 4. Gleichstand an der Spitze (mindestens zwei gleichauf). Nach einer
		// Runde ohne Antwort ist "leer" die bessere Ansage.
		if name, andere, ja := g.quizModSpitze(); ja {
			merke("gleichstand", quizModLage{Name: name, Andere: andere})
		}
	}

	// 5. Treffer: der beste Richtige der Runde.
	bester, besterWert := -1, 0
	for i := range g.Spieler {
		s := &g.Spieler[i]
		if s.PunkteRunde > besterWert {
			bester, besterWert = i, s.PunkteRunde
		}
	}
	if bester >= 0 {
		merke("richtig", quizModLage{
			Name:   g.Spieler[bester].Name,
			Punkte: quizPunkteText(besterWert),
		})
	}

	// 6. Punktverlust ohne Treffer.
	verlierer := -1
	for i := range g.Spieler {
		s := &g.Spieler[i]
		if s.PunkteRunde < 0 && (verlierer < 0 || s.PunkteRunde < g.Spieler[verlierer].PunkteRunde) {
			verlierer = i
		}
	}
	if verlierer >= 0 {
		s := &g.Spieler[verlierer]
		// Gelegentlich zitiert er, was da wirklich stand, statt nur allgemein
		// zu spotten. Das Zitat ist die Antwort aus dem Spielstand – jeder
		// Spieler liest also denselben Wortlaut.
		if falsch, ja := g.quizModFalscheAntwort(s); ja && g.modStand().rng.Intn(3) == 0 {
			merke("falschantwort", quizModLage{Name: s.Name, Falsch: falsch})
		}
		merke("verlust", quizModLage{
			Name:   s.Name,
			Punkte: quizPunkteText(s.PunkteRunde),
		})
	}

	// 7. Rückfall: niemand hat gepunktet und niemand verlor.
	if len(kandidaten) == 0 {
		merke("leer", quizModLage{})
	}

	// Zuschlag: erste passende Art, die nicht die der Vorrunde ist.
	for _, k := range kandidaten {
		if k.art != vermeiden {
			return k.art, k.lage
		}
	}
	return kandidaten[0].art, kandidaten[0].lage
}

// quizModFalscheAntwort liefert den Wortlaut, den dieser Spieler wirklich
// gewählt hat – aber nur, wenn er sich vorlesen lässt. Lange oder leere
// Antworten bleiben außen vor, dann kommt der allgemeine Spruch.
func (g *QuizGame) quizModFalscheAntwort(s *QuizSpieler) (string, bool) {
	if s == nil || s.Antwort == nil || s.Richtig || g.Aktuell == nil {
		return "", false
	}
	i := *s.Antwort
	if i < 0 || i >= len(g.Aktuell.Antworten) {
		return "", false
	}
	text := strings.Join(strings.Fields(g.Aktuell.Antworten[i]), " ")
	// Der Satzpunkt der Antwort gehört nicht ins Zitat: „Der Schaffner“ statt
	// „Der Schaffner.“? – die Satzzeichen setzt der Spruch selbst.
	text = strings.TrimRight(text, ".!?,;: ")
	if n := len([]rune(text)); n < 3 || n > 48 {
		return "", false
	}
	return text, true
}

// quizModSpitze liefert die zwei Namen an der Spitze, wenn dort mindestens
// zwei gleichauf liegen.
func (g *QuizGame) quizModSpitze() (string, string, bool) {
	oben := g.quizSpitze()
	if len(oben) < 2 {
		return "", "", false
	}
	return g.Spieler[oben[0]].Name, g.Spieler[oben[1]].Name, true
}

// quizModEnde kommentiert die Rangliste: Sieg mit Verfolger oder Gleichstand.
func (g *QuizGame) quizModEnde() {
	if len(g.Spieler) == 0 {
		return
	}
	if len(g.Sieger) == 0 {
		// Ohne Spitze bleiben die Platzhalter leer und werden beim Sprechen
		// durch neutrale Worte ersetzt.
		name, andere, _ := g.quizModSpitze()
		g.quizModSage("ende_gleich", quizModLage{Name: name, Andere: andere})
		return
	}
	// Zwei oder mehr auf Platz eins: geteilter Sieg, kein Einzelsieger.
	if len(g.Sieger) > 1 {
		g.quizModSage("ende_doppel", quizModLage{
			Name:   g.quizName(g.Sieger[0]),
			Andere: g.quizName(g.Sieger[1]),
		})
		return
	}
	sieger := g.quizSitz(g.Sieger[0])
	if sieger < 0 {
		return
	}
	// Verlierer = die Person mit den wenigsten Punkten (bei Gleichstand die
	// erste, die nicht gewonnen hat).
	letzter := -1
	for i := range g.Spieler {
		if g.Sieger[0] == g.Spieler[i].ID {
			continue
		}
		if letzter < 0 || g.Spieler[i].Punkte < g.Spieler[letzter].Punkte {
			letzter = i
		}
	}
	lage := quizModLage{
		Name:   g.Spieler[sieger].Name,
		Punkte: strconv.Itoa(g.Spieler[sieger].Punkte),
	}
	if letzter >= 0 {
		lage.Andere = g.Spieler[letzter].Name
		lage.Abstand = g.Spieler[sieger].Punkte - g.Spieler[letzter].Punkte
	}
	g.quizModSage("ende_sieg", lage)
}
