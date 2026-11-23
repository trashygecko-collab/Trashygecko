package main

/* ===========================================================================
   SCHERZSHOW – das Quizspiel

   Ein reines Scherzfragen-Quiz: keine Wissensfragen, nur Wortspiele,
   Denkfallen und absurde Situationen. Die Fragen stehen im Katalog
   (quizfragen.go, spiele/quizshow/fragen.json) – im Spielcode steht keine
   einzige.

   Zwei Modi:
     tempo  – alle sehen dieselbe Frage, wer zuerst richtig antwortet, bekommt
              die meisten Punkte (100/75/50/25 nach Reihenfolge des Eintreffens
              am SERVER), falsch kostet 25.
     punkte – feste Antwortzeit, Geschwindigkeit zählt nicht: richtig +100,
              falsch -25, keine Antwort 0.

   Gerechnet wird ausschließlich hier: Der Browser schickt nur den Index der
   gewählten Antwort, nie Punkte und nie eine Zeit. Die Reihenfolge im
   Tempo-Modus ergibt sich aus der Serveruhr beim Eintreffen, doppelt antworten
   geht nicht, nach Ablauf der Frist zählt keine Antwort mehr.

   Eine Runde endet, wenn alle geantwortet haben oder die Zeit abgelaufen ist.
   Danach zeigt die Auflösung die richtige Antwort samt Spruch; nach der Pause
   kommt die nächste Frage, nach der letzten die Rangliste. Angetrieben wird
   das – wie überall hier – vom GET der Spielseite.
   =========================================================================== */

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	quizMaxSpieler  = 6
	quizMaxRunden   = 20
	quizStandardRun = 5
	quizZeitTempo   = 20 * time.Second
	quizZeitPunkte  = 15 * time.Second
	quizPauseZeit   = 7 * time.Second
)

// QuizPunkte sind je Partie einstellbar (die Spielseite reicht Werte aus dem
// Eröffnen-Dialog durch); die Vorgaben sind die aus der Spielidee.
type QuizPunkte struct {
	Erste   int `json:"erste"`
	Zweite  int `json:"zweite"`
	Dritte  int `json:"dritte"`
	Weitere int `json:"weitere"`
	Falsch  int `json:"falsch"` // positiver Betrag, wird abgezogen
}

// QuizFrageSicht ist die Frage, wie sie während der Runde übertragen wird –
// ohne richtige Antwort und ohne Auflösung.
type QuizFrageSicht struct {
	ID        string   `json:"id"`
	Kat       string   `json:"kat"`
	KatName   string   `json:"katName"`
	Stufe     int      `json:"stufe"`
	StufeText string   `json:"stufeText"`
	Frage     string   `json:"frage"`
	Antworten []string `json:"antworten"`
}

// QuizLoesung erscheint erst in der Auflösung.
type QuizLoesung struct {
	Richtig    int    `json:"richtig"`
	Text       string `json:"text"`
	Aufloesung string `json:"aufloesung"`
}

type QuizSpieler struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Punkte      int       `json:"punkte"`
	Ready       bool      `json:"ready"`
	Antwort     *int      `json:"antwort,omitempty"`
	Richtig     bool      `json:"richtig"`
	PunkteRunde int       `json:"punkteRunde"`
	Zeit        time.Time `json:"-"` // Serverzeit der Antwort (Reihenfolge)
}

type QuizGame struct {
	ID        string          `json:"id"`
	HostID    string          `json:"hostId"`
	Status    string          `json:"status"` // invited, spiel, aufloesung, ende
	Modus     string          `json:"modus"`  // tempo, punkte
	ModusName string          `json:"modusName"`
	Kat       string          `json:"kat,omitempty"`
	KatName   string          `json:"katName,omitempty"`
	Runden    int             `json:"runden"` // Zahl der Fragen dieser Partie
	Runde     int             `json:"runde"`  // laufende Frage (1…Runden)
	Frage     *QuizFrageSicht `json:"frage,omitempty"`
	Loesung   *QuizLoesung    `json:"loesung,omitempty"`
	Spieler   []QuizSpieler   `json:"spieler"`
	RestMs    int             `json:"restMs"`
	ZeitMs    int             `json:"zeitMs"`
	Punkte    QuizPunkte      `json:"punkte"`
	Sieger    []string        `json:"sieger,omitempty"`
	Note      string          `json:"note,omitempty"`
	// Moderator ist der aktuelle Spruch des Quizmasters. Er entsteht nur am
	// Server, damit alle Spieler derselben Partie dasselbe hören.
	Moderator *QuizAnsage `json:"moderator,omitempty"`

	// Nur am Server:
	Frist    time.Time     `json:"-"`
	PauseBis time.Time     `json:"-"`
	Zeit     time.Duration `json:"-"`
	Pause    time.Duration `json:"-"`
	Vorrat   []QuizFrage   `json:"-"`
	Aktuell  *QuizFrage    `json:"-"`
	// KatNamen kommt beim Start aus dem Katalog, damit die Frage ihre
	// Kategorie auch dann zeigen kann, wenn gemischt gespielt wird.
	KatNamen map[string]string `json:"-"`
	// Bank und Pool sind die Antwortdatenbank der Partie: aus ihr entstehen
	// die drei falschen Antworten jeder Frage.
	Bank QuizBank            `json:"-"`
	Pool map[string][]string `json:"-"`
	// ModStand ist der Auswahlzustand des Moderators (zuletzt benutzte Texte).
	ModStand *quizModStand `json:"-"`
}

func quizModusName(modus string) string {
	if modus == "punkte" {
		return "Punktequiz"
	}
	return "Geschwindigkeitsquiz"
}

func quizNeuesSpiel(hostID, hostName, modus, kat string, runden int, id string) *QuizGame {
	modus = strings.ToLower(strings.TrimSpace(modus))
	if modus != "punkte" {
		modus = "tempo"
	}
	if runden < 1 || runden > quizMaxRunden {
		runden = quizStandardRun
	}
	zeit := quizZeitTempo
	if modus == "punkte" {
		zeit = quizZeitPunkte
	}
	return &QuizGame{
		ID: id, HostID: hostID, Status: "invited",
		Modus: modus, ModusName: quizModusName(modus),
		Kat: kat, Runden: runden,
		Zeit: zeit, ZeitMs: int(zeit.Milliseconds()), Pause: quizPauseZeit,
		Punkte: QuizPunkte{Erste: 100, Zweite: 75, Dritte: 50, Weitere: 25, Falsch: 25},
		// Auch der Gastgeber erklärt sich bereit – sonst könnte er allein
		// starten und die anderen müssten hinterherspringen.
		Spieler: []QuizSpieler{{ID: hostID, Name: hostName}},
	}
}

func (g *QuizGame) quizSitz(id string) int {
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			return i
		}
	}
	return -1
}

// quizZeit fuehrt eine eingesendete Zeit in den erlaubten Bereich.
func quizZeit(sekunden int) time.Duration {
	if sekunden < 5 {
		sekunden = int(quizZeitTempo / time.Second)
	}
	if sekunden > 120 {
		sekunden = 120
	}
	return time.Duration(sekunden) * time.Second
}

// quizStart zieht den Vorrat aus dem Katalog: nur die gewählte Kategorie,
// keine Frage zweimal, höchstens so viele wie gewünscht. Gibt es zu wenige,
// spielt die Partie eben kürzer – sie startet nicht mit doppelten Fragen.
func (g *QuizGame) quizStart(bank QuizBank) string {
	pool := quizAntwortPool(bank)
	vorrat := make([]QuizFrage, 0, len(bank.Fragen))
	gesperrt := 0
	for _, f := range bank.Fragen {
		if g.Kat != "" && f.Kat != g.Kat {
			continue
		}
		// Ohne drei fremde Antworten derselben Gruppe wird eine Frage nicht
		// freigegeben – lieber eine Runde weniger als vier halbe Antworten.
		if _, _, err := quizOptionen(f, pool); err != nil {
			gesperrt++
			continue
		}
		vorrat = append(vorrat, f)
	}
	if len(vorrat) == 0 {
		if len(bank.Fragen) == 0 {
			return "Der Fragenkatalog ist leer."
		}
		if g.Kat != "" {
			if gesperrt > 0 {
				return "Keine Frage dieser Kategorie ist spielbar – es fehlen Antworten in den Antwortgruppen."
			}
			return "Zu dieser Kategorie gibt es noch keine Fragen."
		}
		return "Keine Frage ist spielbar – es fehlen Antworten in den Antwortgruppen."
	}
	g.KatNamen = make(map[string]string, len(bank.Kategorien))
	for _, k := range bank.Kategorien {
		g.KatNamen[k.ID] = k.Name
	}
	g.Bank = bank
	g.Pool = pool
	rand.Shuffle(len(vorrat), func(i, j int) { vorrat[i], vorrat[j] = vorrat[j], vorrat[i] })
	if len(vorrat) > g.Runden {
		vorrat = vorrat[:g.Runden]
	}
	g.Vorrat = vorrat
	g.Runden = len(vorrat)
	g.Runde = 0
	for i := range g.Spieler {
		g.Spieler[i].Punkte = 0
		g.Spieler[i].PunkteRunde = 0
		g.Spieler[i].Richtig = false
		g.Spieler[i].Antwort = nil
	}
	g.Sieger = nil
	g.Loesung = nil
	g.quizNaechste(time.Now())
	g.quizModSage("auftakt", quizModLage{Name: g.Spieler[0].Name})
	return ""
}

// quizNaechste zieht die nächste Frage aus dem Vorrat und baut ihre vier
// Antworten aus der Datenbank. Die Auswahl entsteht EINMAL am Server und
// steht damit für alle Spieler dieser Runde fest; jede neue Runde mischt neu.
func (g *QuizGame) quizNaechste(jetzt time.Time) bool {
	if g.Pool == nil {
		g.Pool = quizAntwortPool(g.Bank)
	}
	for len(g.Vorrat) > 0 {
		i := rand.Intn(len(g.Vorrat))
		f := g.Vorrat[i]
		g.Vorrat = append(g.Vorrat[:i], g.Vorrat[i+1:]...)
		optionen, _, err := quizOptionen(f, g.Pool)
		if err != nil {
			// Gesperrte Frage (zu wenige Antworten) wird übersprungen.
			continue
		}
		richtig := 0
		for j, a := range optionen {
			if strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(f.Antwort)) {
				richtig = j
			}
		}
		f.Antworten = optionen
		f.Richtig = richtig
		g.Aktuell = &f
		g.Frage = &QuizFrageSicht{
			ID: f.ID, Kat: f.Kat, KatName: g.KatNamen[f.Kat], Stufe: f.Stufe, StufeText: quizStufeText(f.Stufe),
			Frage: quizFrageAnzeigen(f.Frage), Antworten: append([]string(nil), optionen...),
		}
		g.Loesung = nil
		g.Runde++
		g.Status = "spiel"
		g.Frist = jetzt.Add(g.Zeit)
		for i := range g.Spieler {
			g.Spieler[i].Antwort = nil
			g.Spieler[i].Zeit = time.Time{}
			g.Spieler[i].Richtig = false
			g.Spieler[i].PunkteRunde = 0
		}
		g.Note = "Frage " + strconv.Itoa(g.Runde) + " von " + strconv.Itoa(g.Runden) + "."
		// Ab der zweiten Frage wirft der Moderator gelegentlich etwas ein. Der
		// Würfel liegt am Server, also bekommt es die ganze Partie gleich.
		if g.Runde >= 2 && g.modStand().rng.Intn(3) == 0 {
			g.quizModSage("zwischen", quizModLage{})
		}
		return true
	}
	g.quizEnde()
	return false
}

// quizAlleDran meldet, ob jeder Spieler eine Antwort abgegeben hat.
func (g *QuizGame) quizAlleDran() bool {
	for i := range g.Spieler {
		if g.Spieler[i].Antwort == nil {
			return false
		}
	}
	return true
}

// quizAuswerten rechnet die Punkte. Die Reihenfolge im Tempo-Modus ist die
// Reihenfolge der Serverzeit – der Client kann sie nicht beeinflussen.
func (g *QuizGame) quizAuswerten(jetzt time.Time) {
	if g.Aktuell == nil {
		g.quizEnde()
		return
	}
	richtig := g.Aktuell.Richtig
	// Wer stand vor dieser Frage oben? Der Moderator kommentiert danach den
	// Unterschied – dafür muss der alte Stand vor der Punktevergabe feststehen.
	fuehrerVorher := g.quizFuehrer()
	punkteVorher := make(map[string]int, len(g.Spieler))
	for i := range g.Spieler {
		punkteVorher[g.Spieler[i].ID] = g.Spieler[i].Punkte
	}
	reihe := make([]int, 0, len(g.Spieler))
	for i := range g.Spieler {
		if g.Spieler[i].Antwort != nil {
			reihe = append(reihe, i)
		}
	}
	sort.SliceStable(reihe, func(a, b int) bool {
		return g.Spieler[reihe[a]].Zeit.Before(g.Spieler[reihe[b]].Zeit)
	})
	platz := 0
	for _, i := range reihe {
		if *g.Spieler[i].Antwort != richtig {
			g.Spieler[i].Punkte -= g.Punkte.Falsch
			g.Spieler[i].PunkteRunde = -g.Punkte.Falsch
			continue
		}
		platz++
		punkte := g.Punkte.Weitere
		if g.Modus == "punkte" {
			// Punktequiz: Geschwindigkeit zählt nicht, alle Richtigen gleich.
			punkte = g.Punkte.Erste
		} else {
			switch platz {
			case 1:
				punkte = g.Punkte.Erste
			case 2:
				punkte = g.Punkte.Zweite
			case 3:
				punkte = g.Punkte.Dritte
			}
		}
		g.Spieler[i].Punkte += punkte
		g.Spieler[i].PunkteRunde = punkte
		g.Spieler[i].Richtig = true
	}
	g.Loesung = &QuizLoesung{
		Richtig: richtig, Text: g.Aktuell.Antworten[richtig],
		Aufloesung: g.Aktuell.Aufloesung,
	}
	g.Status = "aufloesung"
	g.PauseBis = jetzt.Add(g.Pause)
	g.Note = "Auflösung – " + g.Aktuell.Antworten[richtig]
	g.quizModeratorRunde(fuehrerVorher, punkteVorher)
}

// quizEnde setzt die Rangliste. Gleichstand an der Spitze heißt: kein Sieger.
func (g *QuizGame) quizEnde() {
	g.Status = "ende"
	g.Frage = nil
	// Der Bestwert startet beim ersten Spieler, nicht bei 0 – sonst hätte
	// eine Runde, in der alle nur Minuspunkte gesammelt haben, keinen Sieger
	// und die Meldung nennte „mit 0 Punkten“.
	best := 0
	if len(g.Spieler) > 0 {
		best = g.Spieler[0].Punkte
		for i := range g.Spieler {
			if g.Spieler[i].Punkte > best {
				best = g.Spieler[i].Punkte
			}
		}
	}
	g.Sieger = nil
	for i := range g.Spieler {
		if g.Spieler[i].Punkte == best {
			g.Sieger = append(g.Sieger, g.Spieler[i].ID)
		}
	}
	if len(g.Sieger) == len(g.Spieler) || len(g.Spieler) < 2 {
		g.Sieger = nil
		g.Note = "Unentschieden – niemand hat sich abgesetzt."
		g.quizModEnde()
		return
	}
	namen := make([]string, 0, len(g.Sieger))
	for _, id := range g.Sieger {
		if sitz := g.quizSitz(id); sitz >= 0 {
			namen = append(namen, g.Spieler[sitz].Name)
		}
	}
	g.Note = "Gewonnen: " + strings.Join(namen, ", ") + " mit " +
		strconv.Itoa(best) + " Punkten."
	g.quizModEnde()
}

// quizAntwort nimmt eine Antwort an. Erst die Prüfungen, dann die Serverzeit –
// ein Client kann weder doppelt antworten noch nach Ablauf der Frist.
func (g *QuizGame) quizAntwort(sitz, wahl int, jetzt time.Time) (int, string) {
	if g.Status != "spiel" || g.Frage == nil {
		return http.StatusConflict, "Gerade läuft keine Frage."
	}
	if sitz < 0 || sitz >= len(g.Spieler) {
		return http.StatusForbidden, "Du spielst nicht mit."
	}
	if wahl < 0 || wahl > 3 {
		return http.StatusBadRequest, "Bitte A, B, C oder D wählen."
	}
	if g.Spieler[sitz].Antwort != nil {
		return http.StatusConflict, "Deine Antwort steht schon."
	}
	if !jetzt.Before(g.Frist) {
		return http.StatusConflict, "Die Zeit ist um."
	}
	w := wahl
	g.Spieler[sitz].Antwort = &w
	g.Spieler[sitz].Zeit = jetzt
	if g.quizAlleDran() {
		g.quizAuswerten(jetzt)
	}
	return 0, ""
}

// quizTreiben schließt abgelaufene Runden und startet die nächste Frage.
func quizTreiben(g *QuizGame, jetzt time.Time) {
	if g == nil {
		return
	}
	if g.Status == "spiel" && !jetzt.Before(g.Frist) {
		g.quizAuswerten(jetzt)
	}
	if g.Status == "aufloesung" && !jetzt.Before(g.PauseBis) {
		if len(g.Vorrat) == 0 {
			g.quizEnde()
			return
		}
		g.quizNaechste(jetzt)
	}
}

// quizSicht ist der Stand für den Browser: Restzeit aus der Serveruhr,
// Frage ohne Lösung, Lösung erst in der Auflösung.
func quizSicht(g *QuizGame, jetzt time.Time) *QuizGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Spieler = append([]QuizSpieler(nil), g.Spieler...)
	if g.Frage != nil {
		f := *g.Frage
		f.Antworten = append([]string(nil), g.Frage.Antworten...)
		c.Frage = &f
	}
	if g.Loesung != nil {
		l := *g.Loesung
		c.Loesung = &l
	}
	if g.Moderator != nil {
		m := *g.Moderator
		c.Moderator = &m
	}
	c.Sieger = append([]string(nil), g.Sieger...)
	c.RestMs = 0
	if g.Status == "spiel" && !g.Frist.IsZero() {
		c.RestMs = int(g.Frist.Sub(jetzt).Milliseconds())
	} else if g.Status == "aufloesung" && !g.PauseBis.IsZero() {
		c.RestMs = int(g.PauseBis.Sub(jetzt).Milliseconds())
	}
	if c.RestMs < 0 {
		c.RestMs = 0
	}
	return &c
}

func quizIDs(g *QuizGame) []string {
	ids := make([]string, 0, len(g.Spieler))
	for _, s := range g.Spieler {
		ids = append(ids, s.ID)
	}
	return ids
}

// quizRemoveSeat nimmt einen Spieler aus der Partie (Verlassen/Abbruch).
func quizRemoveSeat(g *QuizGame, id, name string) (bool, bool, int) {
	sitz := g.quizSitz(id)
	if sitz < 0 {
		return false, false, len(g.Spieler)
	}
	g.Spieler = append(g.Spieler[:sitz], g.Spieler[sitz+1:]...)
	if len(g.Spieler) < 2 {
		g.Status = "ende"
		g.Note = name + " ist gegangen – für eine Show braucht es mindestens zwei."
		return true, true, len(g.Spieler)
	}
	if g.HostID == id {
		g.HostID = g.Spieler[0].ID
	}
	g.Note = name + " ist raus – es geht mit " + strconv.Itoa(len(g.Spieler)) + " weiter."
	g.quizModSage("abgang", quizModLage{Name: name})
	if g.Status == "spiel" && g.quizAlleDran() {
		g.quizAuswerten(time.Now())
	}
	return true, false, len(g.Spieler)
}

// quizHandler: GET liefert den Stand (und treibt den Ablauf), POST sind die
// Aktionen der Spielseite und des Connectors.
func quizHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	if code == "" {
		http.Error(w, "Spielraum-Code fehlt", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		mu.Lock()
		room := rooms[code]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		if _, ok := room.Players[id]; !ok {
			mu.Unlock()
			http.Error(w, "Bitte zuerst dem Spielraum beitreten.", http.StatusForbidden)
			return
		}
		jetzt := time.Now()
		quizTreiben(room.Quiz, jetzt)
		stand := quizSicht(room.Quiz, jetzt)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": stand})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID       string `json:"id"`
		Modus    string `json:"modus"`
		Kat      string `json:"kat"`
		Runden   int    `json:"runden"`
		Sekunden int    `json:"sekunden"`
		Wahl     int    `json:"wahl"`
		Punkte   *struct {
			Erste   int `json:"erste"`
			Zweite  int `json:"zweite"`
			Dritte  int `json:"dritte"`
			Weitere int `json:"weitere"`
			Falsch  int `json:"falsch"`
		} `json:"punkte"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.ID = kennung(r)
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

	mu.Lock()
	room := rooms[code]
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	room.cleanPlayers()
	player, ok := room.Players[input.ID]
	if !ok {
		mu.Unlock()
		http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
		return
	}
	fail := func(text string, status int) {
		mu.Unlock()
		http.Error(w, text, status)
	}
	antwort := func(g *QuizGame) {
		room.LastActive = time.Now()
		stand := quizSicht(g, time.Now())
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": stand})
	}

	switch action {
	case "create":
		if room.Quiz != nil && !room.gameSlotFree(room.Quiz.Status, quizIDs(room.Quiz)...) {
			fail("In diesem Spielraum läuft bereits eine Scherzshow.", http.StatusConflict)
			return
		}
		bank, err := quizBankLaden()
		if err != nil {
			fail(err.Error(), http.StatusInternalServerError)
			return
		}
		kat := strings.ToLower(strings.TrimSpace(input.Kat))
		g := quizNeuesSpiel(input.ID, player.Name, input.Modus, kat, input.Runden, randomCode())
		if kat != "" {
			name := ""
			for _, k := range bank.Kategorien {
				if k.ID == kat {
					name = k.Name
				}
			}
			if name == "" {
				fail("Diese Kategorie gibt es nicht.", http.StatusBadRequest)
				return
			}
			g.KatName = name
		}
		if input.Sekunden > 0 {
			g.Zeit = quizZeit(input.Sekunden)
			g.ZeitMs = int(g.Zeit.Milliseconds())
		}
		if p := input.Punkte; p != nil {
			// Negative Werte wären absurd, und ein Falsch-Abzug über dem
			// ersten Platz würde die Show zur Bestrafung machen.
			g.Punkte = QuizPunkte{
				Erste:   quizBegrenzen(p.Erste, 1, 500),
				Zweite:  quizBegrenzen(p.Zweite, 0, 500),
				Dritte:  quizBegrenzen(p.Dritte, 0, 500),
				Weitere: quizBegrenzen(p.Weitere, 0, 500),
				Falsch:  quizBegrenzen(p.Falsch, 0, 200),
			}
		}
		g.Note = player.Name + " eröffnet eine Scherzshow (" + g.ModusName + ")."
		room.Quiz = g
		antwort(g)
	case "join":
		g := room.Quiz
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Show.", http.StatusConflict)
			return
		}
		if g.quizSitz(input.ID) >= 0 {
			fail("Du sitzt schon auf der Bühne.", http.StatusConflict)
			return
		}
		if len(g.Spieler) >= quizMaxSpieler {
			fail("Die Bühne ist voll (höchstens "+strconv.Itoa(quizMaxSpieler)+" Spieler).", http.StatusConflict)
			return
		}
		g.Spieler = append(g.Spieler, QuizSpieler{ID: input.ID, Name: player.Name})
		antwort(g)
	case "ready":
		g := room.Quiz
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Show.", http.StatusConflict)
			return
		}
		sitz := g.quizSitz(input.ID)
		if sitz < 0 {
			fail("Du sitzt nicht auf der Bühne.", http.StatusForbidden)
			return
		}
		g.Spieler[sitz].Ready = !g.Spieler[sitz].Ready
		antwort(g)
	case "start":
		g := room.Quiz
		if g == nil || g.Status != "invited" {
			fail("Es gibt keine offene Show.", http.StatusConflict)
			return
		}
		if g.quizSitz(input.ID) < 0 {
			fail("Du sitzt nicht auf der Bühne.", http.StatusForbidden)
			return
		}
		if len(g.Spieler) < 2 {
			fail("Für eine Show braucht es mindestens zwei Spieler.", http.StatusConflict)
			return
		}
		for _, s := range g.Spieler {
			if !s.Ready {
				fail("Noch nicht alle sind bereit.", http.StatusConflict)
				return
			}
		}
		bank, err := quizBankLaden()
		if err != nil {
			fail(err.Error(), http.StatusInternalServerError)
			return
		}
		if meldung := g.quizStart(bank); meldung != "" {
			fail(meldung, http.StatusConflict)
			return
		}
		antwort(g)
	case "antwort":
		g := room.Quiz
		if g == nil {
			fail("Gerade läuft keine Show.", http.StatusConflict)
			return
		}
		if status, meldung := g.quizAntwort(g.quizSitz(input.ID), input.Wahl, time.Now()); status != 0 {
			fail(meldung, status)
			return
		}
		antwort(g)
	case "weiter":
		g := room.Quiz
		if g == nil || g.Status != "aufloesung" {
			fail("Gerade läuft keine Auflösung.", http.StatusConflict)
			return
		}
		if g.quizSitz(input.ID) < 0 {
			fail("Du spielst nicht mit.", http.StatusForbidden)
			return
		}
		if len(g.Vorrat) == 0 {
			g.quizEnde()
		} else {
			g.quizNaechste(time.Now())
		}
		antwort(g)
	case "reset":
		if room.Quiz == nil {
			fail("Es läuft keine Show.", http.StatusConflict)
			return
		}
		if room.Quiz.HostID != input.ID {
			fail("Nur der Gastgeber kann die Show schließen.", http.StatusForbidden)
			return
		}
		room.Quiz = nil
		room.LastActive = time.Now()
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": nil})
	default:
		fail("Unbekannte Aktion.", http.StatusBadRequest)
	}
}

func quizBegrenzen(wert, min, max int) int {
	if wert < min {
		return min
	}
	if wert > max {
		return max
	}
	return wert
}
