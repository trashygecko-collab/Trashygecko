package main

/* ===========================================================================
   COMPUTERGEGNER FÜR DAS VERRÜCKTE LABYRINTH

   Das Labyrinth ist das einzige Spiel, dessen Regeln im Browser liegen: Der
   Server führt nur den verbindlichen Stand (brettsync.go), jede Seite rechnet
   ihren Zug selbst. Für einen Bot am Server braucht es die Regeln deshalb
   einmal in Go – und zwar als eine Prüffunktion, durch die jeder Bot-Zug
   geht:

     labyrinthSchieben   – Linie wählen, Stein drehen, einschieben
     labyrinthZiehen     – Figur über offene Gänge zum Ziel bewegen und
                           dabei den gesuchten Gegenstand einsammeln

   Beide liefern false, wenn der Zug nach den Regeln nicht erlaubt ist. Der
   Bot setzt seinen Stand danach über dieselbe Stelle im Server zusammen wie
   ein Mensch (syncStandUebernehmen): Zugreihenfolge, Version und „beendet“
   prüft der Server für beide gleich.

   Absichtlich nicht geprüft: der Stand, den ein menschlicher Browser
   schickt. Dort bleiben die Regeln wie bisher im Browser, damit eine alte
   Spielseite weiter funktioniert; der Bot kann aber nur über die Funktionen
   hier ziehen und damit nicht schummeln.
   =========================================================================== */

import (
	"encoding/json"
	"sort"
	"time"
)

// labKachel ist ein Brettfeld: vier Gangrichtungen (N, O, S, W) und der
// Gegenstand darauf (0–7, null = keiner).
type labKachel struct {
	Offen []int `json:"offen"`
	G     *int  `json:"g"`
}

// labLetzter merkt sich den letzten Einschub – dieselbe Linie darf nicht
// sofort zurückgeschoben werden.
type labLetzter struct {
	Achse string `json:"achse"`
	Index int    `json:"index"`
	Seite string `json:"seite"`
}

// labFigur ist eine Spielfigur samt Aufgabenliste.
type labFigur struct {
	I        int    `json:"i"`
	Heim     int    `json:"heim"`
	Name     string `json:"name"`
	KI       bool   `json:"ki"`
	Farbe    string `json:"farbe"`
	Erledigt int    `json:"erledigt"`
	Auf      []int  `json:"auf"`
}

// labStand ist genau der Stand, den die Spielseite über stand() liefert und
// über laden() wieder annimmt.
type labStand struct {
	Groesse int         `json:"groesse"`
	Thema   string      `json:"thema"`
	Modus   string      `json:"modus"`
	Phase   string      `json:"phase"`
	AmZug   int         `json:"amZug"`
	Zuege   int         `json:"zuege"`
	Vorbei  bool        `json:"vorbei"`
	Letzter *labLetzter `json:"letzter"`
	Frei    labKachel   `json:"frei"`
	Drehung int         `json:"drehung"`
	Fest    []bool      `json:"fest"`
	Feld    []labKachel `json:"feld"`
	Figuren []labFigur  `json:"figuren"`
}

// labWahl ist ein Einschub-Kandidat des Bots.
type labWahl struct {
	Achse   string
	Index   int
	Seite   string
	Drehung int
	Wert    float64
}

var labGegen = map[string]string{"links": "rechts", "rechts": "links", "oben": "unten", "unten": "oben"}

func labIdx(n, r, c int) int { return r*n + c }

func labRC(n, i int) (int, int) { return i / n, i % n }

// labDrehe dreht die offenen Seiten um k Schritte im Uhrzeigersinn.
func labDrehe(offen []int, k int) []int {
	k = ((k % 4) + 4) % 4
	aus := make([]int, 4)
	for i := 0; i < 4; i++ {
		aus[(i+k)%4] = offen[i]
	}
	return aus
}

// labKopie ist ein tiefer Stand, damit der Bot auf einer Kopie rechnen kann.
func labKopie(st labStand) labStand {
	c := st
	c.Fest = append([]bool{}, st.Fest...)
	c.Feld = make([]labKachel, len(st.Feld))
	for i, t := range st.Feld {
		c.Feld[i] = labKachel{Offen: append([]int{}, t.Offen...), G: t.G}
	}
	c.Frei = labKachel{Offen: append([]int{}, st.Frei.Offen...), G: st.Frei.G}
	c.Figuren = make([]labFigur, len(st.Figuren))
	for i, f := range st.Figuren {
		c.Figuren[i] = f
		c.Figuren[i].Auf = append([]int{}, f.Auf...)
	}
	if st.Letzter != nil {
		l := *st.Letzter
		c.Letzter = &l
	}
	return c
}

// labyrinthLinieErlaubt: nur die ungeraden Linien sind verschiebbar, und die
// zuletzt geschobene Linie nicht sofort zurück.
func labyrinthLinieErlaubt(st *labStand, achse string, index int, seite string) bool {
	if st.Vorbei || st.Phase != "schieben" {
		return false
	}
	if achse != "r" && achse != "c" {
		return false
	}
	if index < 0 || index >= st.Groesse || index%2 == 0 {
		return false
	}
	if achse == "r" && seite != "links" && seite != "rechts" {
		return false
	}
	if achse == "c" && seite != "oben" && seite != "unten" {
		return false
	}
	if l := st.Letzter; l != nil && l.Achse == achse && l.Index == index && l.Seite == labGegen[seite] {
		return false
	}
	return true
}

// labyrinthSchieben führt einen Einschub aus. Die Figuren in der Linie wandern
// mit, der herausgeschobene Stein wird der neue freie Stein.
func labyrinthSchieben(st *labStand, achse string, index int, seite string, drehung int) bool {
	if !labyrinthLinieErlaubt(st, achse, index, seite) {
		return false
	}
	n := st.Groesse
	neu := labKachel{Offen: labDrehe(st.Frei.Offen, drehung), G: st.Frei.G}
	var raus labKachel

	verschiebeFigur := func(f *labFigur) {
		r, c := labRC(n, f.I)
		if achse == "r" {
			if r != index {
				return
			}
			if seite == "links" {
				c = (c + 1) % n
			} else {
				c = (c - 1 + n) % n
			}
		} else {
			if c != index {
				return
			}
			if seite == "oben" {
				r = (r + 1) % n
			} else {
				r = (r - 1 + n) % n
			}
		}
		f.I = labIdx(n, r, c)
	}

	if achse == "r" {
		if seite == "links" {
			raus = st.Feld[labIdx(n, index, n-1)]
			for c := n - 1; c > 0; c-- {
				st.Feld[labIdx(n, index, c)] = st.Feld[labIdx(n, index, c-1)]
			}
			st.Feld[labIdx(n, index, 0)] = neu
		} else {
			raus = st.Feld[labIdx(n, index, 0)]
			for c := 0; c < n-1; c++ {
				st.Feld[labIdx(n, index, c)] = st.Feld[labIdx(n, index, c+1)]
			}
			st.Feld[labIdx(n, index, n-1)] = neu
		}
	} else {
		if seite == "oben" {
			raus = st.Feld[labIdx(n, n-1, index)]
			for r := n - 1; r > 0; r-- {
				st.Feld[labIdx(n, r, index)] = st.Feld[labIdx(n, r-1, index)]
			}
			st.Feld[labIdx(n, 0, index)] = neu
		} else {
			raus = st.Feld[labIdx(n, 0, index)]
			for r := 0; r < n-1; r++ {
				st.Feld[labIdx(n, r, index)] = st.Feld[labIdx(n, r+1, index)]
			}
			st.Feld[labIdx(n, n-1, index)] = neu
		}
	}

	for i := range st.Figuren {
		verschiebeFigur(&st.Figuren[i])
	}
	st.Frei = labKachel{Offen: append([]int{}, raus.Offen...), G: raus.G}
	st.Drehung = 0
	st.Letzter = &labLetzter{Achse: achse, Index: index, Seite: seite}
	st.Phase = "ziehen"
	st.Zuege++
	return true
}

// labOffenZu meldet, ob von Feld i ein Gang in Richtung seite führt (und
// innerhalb des Bretts liegt).
func labOffenZu(st *labStand, i, seite int) bool {
	if i < 0 || i >= len(st.Feld) {
		return false
	}
	r, c := labRC(st.Groesse, i)
	switch seite {
	case 0:
		return r > 0 && st.Feld[i].Offen[0] == 1
	case 1:
		return c < st.Groesse-1 && st.Feld[i].Offen[1] == 1
	case 2:
		return r < st.Groesse-1 && st.Feld[i].Offen[2] == 1
	case 3:
		return c > 0 && st.Feld[i].Offen[3] == 1
	}
	return false
}

func labFigurAuf(st *labStand, i, ausser int) int {
	for k, f := range st.Figuren {
		if k != ausser && f.I == i {
			return k
		}
	}
	return -1
}

// labKarte ist eine Breitensuche über die offenen Gänge. mitFiguren sperrt
// Felder, auf denen eine andere Figur steht (fürs Ziehen); ohne zählt nur das
// Brett (für die Bewertung eines Einschubs, wie im Browser).
func labKarte(st *labStand, von, ausser int, mitFiguren bool) map[int]int {
	dist := map[int]int{von: 0}
	q := []int{von}
	dr := []int{-1, 0, 1, 0}
	dc := []int{0, 1, 0, -1}
	for kopf := 0; kopf < len(q); kopf++ {
		a := q[kopf]
		r, c := labRC(st.Groesse, a)
		for s := 0; s < 4; s++ {
			if !labOffenZu(st, a, s) {
				continue
			}
			nr, nc := r+dr[s], c+dc[s]
			if nr < 0 || nc < 0 || nr >= st.Groesse || nc >= st.Groesse {
				continue
			}
			b := labIdx(st.Groesse, nr, nc)
			if st.Feld[b].Offen[(s+2)%4] != 1 {
				continue
			}
			if mitFiguren && labFigurAuf(st, b, ausser) != -1 {
				continue
			}
			if _, da := dist[b]; da {
				continue
			}
			dist[b] = dist[a] + 1
			q = append(q, b)
		}
	}
	return dist
}

// labZielFeld ist das Feld des gesuchten Gegenstands; -1, wenn er im freien
// Stein liegt (dann hilft nur ein Einschub).
func labZielFeld(st *labStand) int { return labZielFuer(st, st.Figuren[st.AmZug]) }

func labZielFuer(st *labStand, f labFigur) int {
	if f.Erledigt >= len(f.Auf) {
		return f.Heim
	}
	wunsch := f.Auf[f.Erledigt]
	for i, t := range st.Feld {
		if t.G != nil && *t.G == wunsch {
			return i
		}
	}
	return -1
}

// labAbschluss sammelt den Gegenstand ein und gibt den Zug weiter. Wer alle
// Aufgaben hat und auf seinem Startfeld steht, gewinnt.
func labAbschluss(st *labStand) {
	if st.Vorbei || len(st.Figuren) == 0 {
		return
	}
	f := &st.Figuren[st.AmZug]
	if f.Erledigt < len(f.Auf) && st.Feld[f.I].G != nil && *st.Feld[f.I].G == f.Auf[f.Erledigt] {
		st.Feld[f.I].G = nil
		f.Erledigt++
	}
	if f.Erledigt >= len(f.Auf) && f.I == f.Heim {
		st.Vorbei = true
		st.Phase = "vorbei"
		return
	}
	st.AmZug = (st.AmZug + 1) % len(st.Figuren)
	st.Phase = "schieben"
	st.Drehung = 0
}

// labyrinthZiehen bewegt die Figur des Spielers am Zug auf ein erreichbares
// Feld (auch das eigene: dann wird nur eingesammelt) und schließt den Zug ab.
func labyrinthZiehen(st *labStand, ziel int) bool {
	if st.Vorbei || st.Phase != "ziehen" || len(st.Figuren) == 0 {
		return false
	}
	if ziel < 0 || ziel >= len(st.Feld) {
		return false
	}
	f := &st.Figuren[st.AmZug]
	karte := labKarte(st, f.I, st.AmZug, true)
	if _, da := karte[ziel]; !da {
		return false
	}
	f.I = ziel
	labAbschluss(st)
	return true
}

// labyrinthZugKandidaten listet alle erlaubten Einschübe (Linie, Seite,
// Drehung) – dieselbe Auswahl, die auch der Browser rechnet.
func labyrinthZugKandidaten(st *labStand) []labWahl {
	seiten := [][2]string{{"r", "links"}, {"r", "rechts"}, {"c", "oben"}, {"c", "unten"}}
	var aus []labWahl
	for i := 1; i < st.Groesse; i += 2 {
		for _, s := range seiten {
			if !labyrinthLinieErlaubt(st, s[0], i, s[1]) {
				continue
			}
			for d := 0; d < 4; d++ {
				aus = append(aus, labWahl{Achse: s[0], Index: i, Seite: s[1], Drehung: d})
			}
		}
	}
	return aus
}

func labManhattan(n, a, b int) int {
	ar, ac := labRC(n, a)
	br, bc := labRC(n, b)
	dr, dc := ar-br, ac-bc
	if dr < 0 {
		dr = -dr
	}
	if dc < 0 {
		dc = -dc
	}
	return dr + dc
}

// labyrinthWert bewertet einen Einschub: kurzer Weg zum Ziel zählt viel, ist
// das Ziel nicht erreichbar, zählen Nähe und Größe der Inseln. Gegner, die
// kurz vor ihrem Ziel stehen, kosten Punkte.
func labyrinthWert(st labStand, w labWahl, stufe string) float64 {
	sim := labKopie(st)
	if !labyrinthSchieben(&sim, w.Achse, w.Index, w.Seite, w.Drehung) {
		return -1e9
	}
	ich := sim.AmZug
	f := sim.Figuren[ich]
	ziel := labZielFuer(&sim, f)
	karte := labKarte(&sim, f.I, ich, false)
	wert := 0.0
	if ziel >= 0 {
		if d, da := karte[ziel]; da {
			wert = 120 - float64(d)*8
			if d == 0 {
				wert += 60
				if f.Erledigt >= len(f.Auf) {
					wert += 200
				}
			}
		} else {
			wert = -float64(labManhattan(sim.Groesse, f.I, ziel))*4 +
				float64(len(labKarte(&sim, ziel, -1, false)))*2.2 +
				float64(len(karte))*1.1
		}
	} else {
		wert = float64(len(karte)) * 1.2
		if len(f.Auf) > 0 && f.Erledigt < len(f.Auf) && sim.Frei.G != nil && *sim.Frei.G == f.Auf[f.Erledigt] {
			wert += 30
		}
	}
	for k, o := range sim.Figuren {
		if k == ich {
			continue
		}
		oz := labZielFuer(&sim, o)
		if oz < 0 {
			continue
		}
		ok := labKarte(&sim, o.I, k, false)
		od := 60
		if d, da := ok[oz]; da {
			od = d
		}
		schwer := stufe == botSchwer
		switch {
		case od == 0:
			if schwer {
				wert -= 90
			} else {
				wert -= 45
			}
		case od == 1:
			if schwer {
				wert -= 55
			} else {
				wert -= 22
			}
		case od <= 4:
			if schwer {
				wert -= 14
			} else {
				wert -= 6
			}
		}
	}
	return wert
}

// labyrinthWahl sucht den Einschub. leicht nimmt irgendeinen aus dem oberen
// Drittel, mittel meist den besten, schwer immer den besten.
func labyrinthWahl(st labStand, stufe string) (labWahl, bool) {
	kandidaten := labyrinthZugKandidaten(&st)
	if len(kandidaten) == 0 {
		return labWahl{}, false
	}
	for i := range kandidaten {
		kandidaten[i].Wert = labyrinthWert(st, kandidaten[i], stufe)
	}
	sort.SliceStable(kandidaten, func(a, b int) bool { return kandidaten[a].Wert > kandidaten[b].Wert })
	zufall := botZufall()
	switch stufe {
	case botLeicht:
		n := len(kandidaten) * 35 / 100
		if n < 3 {
			n = 3
		}
		if n > len(kandidaten) {
			n = len(kandidaten)
		}
		return kandidaten[zufall.Intn(n)], true
	case botMittel:
		n := len(kandidaten)
		if n > 5 {
			n = 5
		}
		if zufall.Intn(5) == 0 {
			return kandidaten[zufall.Intn(n)], true
		}
	}
	return kandidaten[0], true
}

// labyrinthZiel wählt das Feld, auf das die Figur zieht: das Ziel, wenn es
// erreichbar ist, sonst das erreichbare Feld, das ihm am nächsten liegt.
func labyrinthZiel(st *labStand, stufe string) int {
	f := st.Figuren[st.AmZug]
	ziel := labZielFuer(st, f)
	if ziel < 0 {
		return -1
	}
	karte := labKarte(st, f.I, st.AmZug, true)
	if _, da := karte[ziel]; da {
		return ziel
	}
	bestes, bester := -1, 1<<30
	for i := range karte {
		if i == f.I {
			continue
		}
		d := labManhattan(st.Groesse, i, ziel)
		if d < bester {
			bester, bestes = d, i
		}
	}
	if bestes >= 0 && stufe == botLeicht && botZufall().Intn(100) < 35 {
		for i := range karte {
			if i != f.I {
				return i
			}
		}
	}
	return bestes
}

// labyrinthBotZug rechnet einen vollständigen Zug: einschieben, dann ziehen
// (oder stehen bleiben). false heißt: kein erlaubter Zug gefunden.
func labyrinthBotZug(st labStand, stufe string) (labStand, bool) {
	w, ok := labyrinthWahl(st, stufe)
	if !ok {
		return st, false
	}
	neu := labKopie(st)
	if !labyrinthSchieben(&neu, w.Achse, w.Index, w.Seite, w.Drehung) {
		return st, false
	}
	ziel := labyrinthZiel(&neu, stufe)
	if ziel < 0 || !labyrinthZiehen(&neu, ziel) {
		labAbschluss(&neu)
	}
	return neu, true
}

// labyrinthBotTreiben lässt den Bot ziehen, wenn er am Zug ist und seine
// Denkpause vorbei ist. Der Stand wird über syncStandUebernehmen gesetzt –
// dieselbe Stelle, durch die auch ein menschlicher Zug geht.
func labyrinthBotTreiben(room *Room, g *SyncGame, jetzt time.Time) {
	if g == nil || g.Status != "active" || len(g.Stand) == 0 {
		return
	}
	k := g.Current
	if k < 0 || k >= len(g.Spieler) || !g.Spieler[k].Bot {
		return
	}
	if !g.BotFrist.IsZero() && jetzt.Before(g.BotFrist) {
		return
	}
	var st labStand
	if err := json.Unmarshal(g.Stand, &st); err != nil || st.Vorbei || len(st.Figuren) == 0 {
		return
	}
	if st.AmZug != k {
		return
	}

	if st.Phase == "ziehen" {
		// Nach einem Neustart mitten im Zug: nur noch die Figur bewegen.
		ziel := labyrinthZiel(&st, g.BotStufe)
		if ziel < 0 || !labyrinthZiehen(&st, ziel) {
			labAbschluss(&st)
		}
	} else {
		neu, ok := labyrinthBotZug(st, g.BotStufe)
		if !ok {
			return
		}
		st = neu
	}

	roh, err := json.Marshal(st)
	if err != nil {
		return
	}
	syncStandUebernehmen(g, roh, st.AmZug, st.Vorbei, true)
	if st.Vorbei {
		g.Note = g.Spieler[st.AmZug].Name + " gewinnt."
	}
	if g.Status == "active" && g.Current >= 0 && g.Current < len(g.Spieler) && g.Spieler[g.Current].Bot {
		g.BotFrist = jetzt.Add(botWartezeit(g.BotStufe))
	} else {
		g.BotFrist = time.Time{}
	}
	if room != nil {
		room.LastActive = jetzt
	}
}

// labyrinthBotFristSetzen startet die Denkpause, sobald ein Bot am Zug ist.
func labyrinthBotFristSetzen(g *SyncGame, jetzt time.Time) {
	if g == nil || g.Status != "active" || g.Current < 0 || g.Current >= len(g.Spieler) {
		return
	}
	if !g.Spieler[g.Current].Bot {
		g.BotFrist = time.Time{}
		return
	}
	if g.BotFrist.IsZero() {
		g.BotFrist = jetzt.Add(botWartezeit(g.BotStufe))
	}
}
