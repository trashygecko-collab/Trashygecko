package main

/* ===========================================================================
   FRAGENKATALOG UND ANTWORTDATENBANK DER SCHERZSHOW

   Die Fragen liegen NICHT im Spielcode, sondern in einer eigenen Datei neben
   der Spielseite (spiele/quizshow/fragen.json). Jede Frage speichert nur noch
   ihre RICHTIGE Antwort; die drei falschen Antworten entstehen zur Laufzeit
   aus dem Pool der richtigen Antworten aller anderen Fragen.

   Antwortgruppen (automatisch erkannt, im Editor änderbar):
     einzelwort  – Keks, Kühlschrank, Klopfsalat
     name        – Klaus, Peter, Rüdiger
     kurzgruppe  – „im Keller“, „unter dem Tisch“
     kurzaussage – „Weil er keine Schuhe hatte.“
     satz        – „Weil der Keks unter dem Baum lag.“
     zahl        – 7, 42, 100
     sonstig     – alles, was in keine Gruppe passt

   Format (eine Zeile je Frage, damit Diffs lesbar bleiben):
     { "id": "q001", "kat": "scherz", "stufe": 1, "typ": "kurzgruppe",
       "frage": "…", "antwort": "…", "aufloesung": "…" }

   Der alte Stand mit vier festen Antworten wird beim Laden übernommen:
   Antworten[Richtig] wird zur richtigen Antwort, die drei falschen entfallen.

   Geschrieben wird nur mit gültigem Platz-Token (dieselbe Prüfung wie bei
   jedem Spielzug): Ein fremder Browser kann den Katalog nicht leeren.
   =========================================================================== */

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// Antwortgruppen der zentralen Antwortdatenbank.
const (
	quizTypWort     = "einzelwort"
	quizTypName     = "name"
	quizTypGruppe   = "kurzgruppe"
	quizTypAussage  = "kurzaussage"
	quizTypSatz     = "satz"
	quizTypZahl     = "zahl"
	quizTypSonstig  = "sonstig"
	quizAntwortMind = 3 // so viele falsche Antworten braucht eine Runde
)

// QuizAntwortTyp ist ein Eintrag der Antwortgruppen (für den Editor).
type QuizAntwortTyp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var quizAntwortTypen = []QuizAntwortTyp{
	{quizTypWort, "Einzelwörter"},
	{quizTypName, "Namen"},
	{quizTypGruppe, "Kurze Wortgruppen"},
	{quizTypAussage, "Kurze Aussagen"},
	{quizTypSatz, "Ganze Sätze"},
	{quizTypZahl, "Zahlen"},
	{quizTypSonstig, "Sonstige Antworten"},
}

// quizErsatzgruppen: reicht eine Gruppe nicht, wird in dieser Reihenfolge
// weiter gesucht. Namen und Zahlen bleiben unter sich – ein Name darf nie
// gegen ein Ding antreten, sonst ist die Frage geschenkt.
var quizErsatzgruppen = map[string][]string{
	quizTypWort:    {quizTypGruppe, quizTypSonstig},
	quizTypGruppe:  {quizTypWort, quizTypAussage, quizTypSonstig},
	quizTypAussage: {quizTypSatz, quizTypGruppe, quizTypSonstig},
	quizTypSatz:    {quizTypAussage, quizTypGruppe, quizTypSonstig},
	quizTypSonstig: {quizTypWort, quizTypGruppe},
	quizTypName:    nil,
	quizTypZahl:    nil,
}

// quizNamensListe: kleine Hilfe für die automatische Zuordnung. Wer einen
// Namen einträgt, der hier fehlt, kann den Typ im Editor umstellen – das
// Spiel nimmt dann den gespeicherten Typ.
var quizNamensListe = map[string]bool{
	"anna": true, "anton": true, "bert": true, "bernd": true, "bruno": true,
	"carla": true, "carl": true, "dieter": true, "doris": true, "emil": true,
	"emma": true, "erich": true, "eva": true, "felix": true, "fritz": true,
	"gerd": true, "greta": true, "gustav": true, "hans": true, "heidi": true,
	"heinz": true, "helga": true, "hugo": true, "ines": true, "jakob": true,
	"jan": true, "jens": true, "jonas": true, "julia": true, "jürgen": true,
	"kai": true, "karin": true, "karl": true, "klaus": true, "kunibert": true,
	"lars": true, "lena": true, "leo": true, "lisa": true, "lotte": true,
	"lukas": true, "manfred": true, "maria": true, "martin": true, "max": true,
	"mia": true, "monika": true, "moritz": true, "nina": true, "olaf": true,
	"otto": true, "paul": true, "peter": true, "petra": true, "ralf": true,
	"rita": true, "rüdiger": true, "sabine": true, "sarah": true, "sven": true,
	"theo": true, "tina": true, "tom": true, "ulrich": true, "ursula": true,
	"ute": true, "vera": true, "walter": true, "werner": true, "wilma": true,
	"wolfgang": true,
	// Weitere Vornamen, damit neue Alle-Kinder-Reime ohne Handarbeit
	// in der richtigen Gruppe landen.
	"annette": true, "christiane": true, "gitte": true, "jenny": true,
	"lutz": true, "rainer": true, "renate": true, "rolf": true,
	"rosel": true, "thorsten": true, "torsten": true, "uli": true,
	"ulli": true, "vroni": true,
}

// QuizKategorie ist ein Eintrag der Fragenkategorien.
type QuizKategorie struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// QuizFrage ist eine Frage des Katalogs. Gespeichert wird nur die richtige
// Antwort; Antworten und Richtig sind der alte Stand (vier feste Antworten)
// und werden beim Laden übernommen.
type QuizFrage struct {
	ID         string `json:"id"`
	Kat        string `json:"kat"`
	Stufe      int    `json:"stufe"` // 1 leicht, 2 mittel, 3 gemein
	Typ        string `json:"typ,omitempty"`
	Frage      string `json:"frage"`
	Antwort    string `json:"antwort"`
	Aufloesung string `json:"aufloesung"`

	Antworten []string `json:"antworten,omitempty"`
	Richtig   int      `json:"richtig,omitempty"`
}

// QuizBank ist die ganze Datei.
type QuizBank struct {
	Stand      string          `json:"stand"`
	Kategorien []QuizKategorie `json:"kategorien"`
	Fragen     []QuizFrage     `json:"fragen"`
}

// QuizFrageVorschau zeigt dem Editor, welche vier Antworten das System für
// diese Frage gerade erzeugen würde.
type QuizFrageVorschau struct {
	Typ       string   `json:"typ"`
	TypName   string   `json:"typName"`
	Antworten []string `json:"antworten"`
	Richtig   int      `json:"richtig"`
	Fehler    string   `json:"fehler,omitempty"`
}

// QuizBankSicht ist der Katalog samt Antwortgruppen und Vorschau.
type QuizBankSicht struct {
	QuizBank
	Typen    []QuizAntwortTyp             `json:"typen"`
	Pool     map[string]int               `json:"pool"`
	Vorschau map[string]QuizFrageVorschau `json:"vorschau"`
}

// quizBankMu schützt Lesen und Schreiben der Datei. Der Handel ist ein Prozess,
// deshalb genügt ein Mutex im Speicher.
var quizBankMu sync.Mutex

// quizStufeText benennt die Schwierigkeit für die Anzeige.
func quizStufeText(stufe int) string {
	switch stufe {
	case 1:
		return "leicht"
	case 3:
		return "gemein"
	}
	return "mittel"
}

// quizBankPfad sucht fragen.json neben der Spielseite – wie registerDatei
// neben der Connector-Seite.
func quizBankPfad() (string, error) {
	var kandidaten []string
	if home != "" {
		kandidaten = append(kandidaten, filepath.Join(home, "spiele", "quizshow", "fragen.json"))
	}
	if hier, err := os.Getwd(); err == nil {
		kandidaten = append(kandidaten,
			filepath.Join(hier, "..", "quizshow", "fragen.json"),
			filepath.Join(hier, "spiele", "quizshow", "fragen.json"),
			filepath.Join(hier, "quizshow", "fragen.json"))
	}
	for _, k := range kandidaten {
		if info, err := os.Stat(k); err == nil && !info.IsDir() {
			return k, nil
		}
	}
	return "", errors.New("Fragenkatalog nicht gefunden (spiele/quizshow/fragen.json)")
}

// quizBankLaden liest und prüft den Katalog.
func quizBankLaden() (QuizBank, error) {
	quizBankMu.Lock()
	defer quizBankMu.Unlock()
	return quizBankLadenOhneSperre()
}

func quizBankLadenOhneSperre() (QuizBank, error) {
	pfad, err := quizBankPfad()
	if err != nil {
		return QuizBank{}, err
	}
	roh, err := os.ReadFile(pfad)
	if err != nil {
		return QuizBank{}, err
	}
	var bank QuizBank
	if err := json.Unmarshal(roh, &bank); err != nil {
		return QuizBank{}, errors.New("Fragenkatalog ist kein gültiges JSON: " + err.Error())
	}
	for i := range bank.Fragen {
		bank.Fragen[i] = quizFrageNormal(bank.Fragen[i])
	}
	if err := quizBankPruefen(bank); err != nil {
		return QuizBank{}, err
	}
	return bank, nil
}

// quizFrageNormal übernimmt den alten Stand (vier Antworten, eine markiert)
// und ergänzt den Antworttyp, wenn keiner gespeichert ist.
func quizFrageNormal(f QuizFrage) QuizFrage {
	if strings.TrimSpace(f.Antwort) == "" && len(f.Antworten) > 0 {
		richtig := f.Richtig
		if richtig < 0 || richtig >= len(f.Antworten) {
			richtig = 0
		}
		f.Antwort = strings.TrimSpace(f.Antworten[richtig])
	}
	f.Antworten = nil
	f.Richtig = 0
	if !quizTypBekannt(f.Typ) {
		f.Typ = quizAntwortTypErkennen(f)
	}
	return f
}

// quizBankPruefen hält den Katalog spielbar: Frage, genau eine richtige
// Antwort, eine bekannte Kategorie, eine Auflösung.
func quizBankPruefen(bank QuizBank) error {
	katOK := map[string]bool{}
	for _, k := range bank.Kategorien {
		katOK[k.ID] = true
	}
	gesehen := map[string]bool{}
	for i, f := range bank.Fragen {
		wo := "Frage " + strconv.Itoa(i+1)
		if f.ID != "" {
			wo = "Frage " + f.ID
		}
		if strings.TrimSpace(f.ID) == "" {
			return errors.New(wo + ": Kennung fehlt")
		}
		if gesehen[f.ID] {
			return errors.New(wo + ": Kennung doppelt")
		}
		gesehen[f.ID] = true
		if strings.TrimSpace(f.Frage) == "" {
			return errors.New(wo + ": Fragetext fehlt")
		}
		if strings.TrimSpace(f.Antwort) == "" {
			return errors.New(wo + ": richtige Antwort fehlt")
		}
		if f.Stufe < 1 || f.Stufe > 3 {
			return errors.New(wo + ": Stufe muss 1–3 sein")
		}
		if !katOK[f.Kat] {
			return errors.New(wo + ": unbekannte Kategorie " + f.Kat)
		}
		if f.Typ != "" && !quizTypBekannt(f.Typ) {
			return errors.New(wo + ": unbekannter Antworttyp " + f.Typ)
		}
		if strings.TrimSpace(f.Aufloesung) == "" {
			return errors.New(wo + ": Auflösung fehlt")
		}
	}
	return nil
}

// quizBankSpeichern schreibt die Datei in einem Stück: erst eine
// Nachbardatei, dann umbenennen. Ein Absturz mitten im Schreiben lässt damit
// nie einen halben Katalog zurück.
func quizBankSpeichern(bank QuizBank) error {
	pfad, err := quizBankPfad()
	if err != nil {
		return err
	}
	if bank.Stand == "" {
		bank.Stand = datum(time.Now())
	}
	if bank.Fragen == nil {
		bank.Fragen = []QuizFrage{}
	}
	for i := range bank.Fragen {
		bank.Fragen[i].Antworten = nil
		bank.Fragen[i].Richtig = 0
	}
	daten, err := json.MarshalIndent(bank, "", "  ")
	if err != nil {
		return err
	}
	daten = append(daten, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(pfad), "fragen-*.json")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(daten); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	return os.Rename(name, pfad)
}

// quizFragePutzt räumt eine eingesendete Frage auf und vergibt bei Bedarf eine
// Kennung. Die Prüfung selbst macht quizBankPruefen über den ganzen Katalog.
func quizFragePutzt(f QuizFrage) QuizFrage {
	f.ID = strings.TrimSpace(f.ID)
	if f.ID == "" {
		f.ID = "q" + strconv.FormatInt(time.Now().UnixNano()%1e9, 10)
	}
	f.Kat = strings.TrimSpace(f.Kat)
	f.Frage = strings.TrimSpace(f.Frage)
	f.Antwort = strings.TrimSpace(f.Antwort)
	f.Aufloesung = strings.TrimSpace(f.Aufloesung)
	f.Typ = strings.ToLower(strings.TrimSpace(f.Typ))
	if !quizTypBekannt(f.Typ) {
		f.Typ = ""
	}
	f = quizFrageNormal(f)
	if f.Stufe < 1 || f.Stufe > 3 {
		f.Stufe = 2
	}
	return f
}

// ---- Antwortgruppen -------------------------------------------------------

var quizZahlMuster = regexp.MustCompile(`^[+-]?\d{1,9}([.,]\d{1,4})?$`)

func quizTypBekannt(id string) bool {
	for _, t := range quizAntwortTypen {
		if t.ID == id {
			return true
		}
	}
	return false
}

func quizTypName_(id string) string {
	for _, t := range quizAntwortTypen {
		if t.ID == id {
			return t.Name
		}
	}
	return id
}

// quizAntwortTypErkennen ordnet eine Antwort automatisch einer Gruppe zu.
func quizAntwortTypErkennen(f QuizFrage) string {
	text := strings.TrimSpace(f.Antwort)
	if text == "" {
		return quizTypSonstig
	}
	if quizZahlMuster.MatchString(text) {
		return quizTypZahl
	}
	if strings.Contains(f.Frage, "{NAME}") {
		return quizTypName
	}
	woerter := strings.Fields(text)
	if len(woerter) == 1 {
		wort := strings.ToLower(strings.Trim(text, ".,!?;:„“\"'"))
		if quizNamensListe[wort] {
			return quizTypName
		}
		if r, _ := utf8.DecodeRuneInString(text); unicode.IsUpper(r) {
			klein := strings.ToLower(f.Frage)
			if strings.Contains(klein, "name") || strings.Contains(klein, "heißt") {
				return quizTypName
			}
		}
		return quizTypWort
	}
	if len(woerter) <= 4 {
		return quizTypGruppe
	}
	if len(woerter) <= 5 {
		return quizTypAussage
	}
	return quizTypSatz
}

// quizAntwortTyp nimmt den gespeicherten Typ, sonst die Automatik.
func quizAntwortTyp(f QuizFrage) string {
	if quizTypBekannt(f.Typ) {
		return f.Typ
	}
	return quizAntwortTypErkennen(f)
}

// quizAntwortPool sammelt je Gruppe alle richtigen Antworten des Katalogs,
// doppelte Texte nur einmal.
func quizAntwortPool(bank QuizBank) map[string][]string {
	pool := map[string][]string{}
	gesehen := map[string]bool{}
	for _, f := range bank.Fragen {
		text := strings.TrimSpace(f.Antwort)
		if text == "" {
			continue
		}
		typ := quizAntwortTyp(f)
		schluessel := strings.ToLower(text)
		if gesehen[schluessel] {
			continue
		}
		gesehen[schluessel] = true
		pool[typ] = append(pool[typ], text)
	}
	return pool
}

// quizOptionen baut die vier Antwortmöglichkeiten: die richtige plus drei
// fremde aus derselben Gruppe. Reicht die Gruppe nicht, hilft eine
// Ersatzgruppe; reicht auch das nicht, wird die Frage gesperrt.
func quizOptionen(f QuizFrage, pool map[string][]string) ([]string, string, error) {
	typ := quizAntwortTyp(f)
	richtig := strings.TrimSpace(f.Antwort)
	if richtig == "" {
		return nil, typ, errors.New("Die richtige Antwort fehlt.")
	}
	verboten := map[string]bool{strings.ToLower(richtig): true}
	kandidaten := make([]string, 0, 24)
	fuege := func(gruppe string) {
		for _, a := range pool[gruppe] {
			k := strings.ToLower(strings.TrimSpace(a))
			if k == "" || verboten[k] {
				continue
			}
			verboten[k] = true
			kandidaten = append(kandidaten, a)
		}
	}
	fuege(typ)
	if len(kandidaten) < quizAntwortMind {
		for _, g := range quizErsatzgruppen[typ] {
			fuege(g)
			if len(kandidaten) >= quizAntwortMind {
				break
			}
		}
	}
	if len(kandidaten) < quizAntwortMind {
		return nil, typ, errors.New("Zu wenige andere Antworten in der Gruppe „" +
			quizTypName_(typ) + "“ (" + strconv.Itoa(len(kandidaten)) + " von " +
			strconv.Itoa(quizAntwortMind) + " nötig).")
	}
	rand.Shuffle(len(kandidaten), func(i, j int) { kandidaten[i], kandidaten[j] = kandidaten[j], kandidaten[i] })
	optionen := append([]string{richtig}, kandidaten[:quizAntwortMind]...)
	rand.Shuffle(len(optionen), func(i, j int) { optionen[i], optionen[j] = optionen[j], optionen[i] })
	return optionen, typ, nil
}

// quizAntworten ist die bequeme Fassung für eine einzelne Frage.
func quizAntworten(f QuizFrage, bank QuizBank) ([]string, string, error) {
	return quizOptionen(f, quizAntwortPool(bank))
}

// quizFrageAnzeigen räumt den Platzhalter für fehlende Namen auf: Die Lücke
// bleibt eine Lücke, sonst stünde die Lösung schon im Reim.
func quizFrageAnzeigen(text string) string {
	text = strings.ReplaceAll(text, "{NAME}", "______")
	text = strings.ReplaceAll(text, "{name}", "______")
	return text
}

// quizBankSicht rechnet Vorschau und Poolgrößen für den Editor.
func quizBankSicht(bank QuizBank) QuizBankSicht {
	pool := quizAntwortPool(bank)
	sicht := QuizBankSicht{
		QuizBank: bank,
		Typen:    quizAntwortTypen,
		Pool:     map[string]int{},
		Vorschau: make(map[string]QuizFrageVorschau, len(bank.Fragen)),
	}
	for _, t := range quizAntwortTypen {
		sicht.Pool[t.ID] = len(pool[t.ID])
	}
	for _, f := range bank.Fragen {
		optionen, typ, err := quizOptionen(f, pool)
		v := QuizFrageVorschau{Typ: typ, TypName: quizTypName_(typ), Antworten: optionen}
		if err != nil {
			v.Fehler = err.Error()
		} else {
			for i, a := range optionen {
				if strings.EqualFold(a, strings.TrimSpace(f.Antwort)) {
					v.Richtig = i
				}
			}
		}
		sicht.Vorschau[f.ID] = v
	}
	return sicht
}

// quizEditorHatRecht: Schreiben darf nur ein Spieler aus einem bestehenden
// Raum mit gültigem Token. Die Kennung kommt aus der geprüften Hülle, den
// Platz selbst muss es im Raum geben.
func quizEditorHatRecht(r *http.Request) bool {
	code := raumAusKontext(r)
	id := kennung(r)
	if code == "" || id == "" {
		return false
	}
	mu.Lock()
	defer mu.Unlock()
	room := rooms[code]
	if room == nil {
		return false
	}
	_, ok := room.Players[id]
	return ok
}

// quizfragenHandler ist der Frageneditor: GET liefert den ganzen Katalog,
// POST ändert ihn (speichern, löschen, Kategorie, Import).
func quizfragenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		bank, err := quizBankLaden()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, quizBankSicht(bank))
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}
	if !quizEditorHatRecht(r) {
		http.Error(w, "Der Fragenkatalog lässt sich nur aus einem Spielraum heraus bearbeiten.", http.StatusForbidden)
		return
	}

	var input struct {
		Action     string          `json:"action"`
		Frage      *QuizFrage      `json:"frage"`
		ID         string          `json:"frageId"`
		Kategorie  *QuizKategorie  `json:"kategorie"`
		Fragen     []QuizFrage     `json:"fragen"`
		Kategorien []QuizKategorie `json:"kategorien"`
		Ersetzen   bool            `json:"ersetzen"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}

	quizBankMu.Lock()
	defer quizBankMu.Unlock()
	bank, err := quizBankLadenOhneSperre()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch strings.ToLower(strings.TrimSpace(input.Action)) {
	case "speichern":
		if input.Frage == nil {
			http.Error(w, "Keine Frage mitgeschickt.", http.StatusBadRequest)
			return
		}
		f := quizFragePutzt(*input.Frage)
		gefunden := false
		for i := range bank.Fragen {
			if bank.Fragen[i].ID == f.ID {
				bank.Fragen[i] = f
				gefunden = true
				break
			}
		}
		if !gefunden {
			bank.Fragen = append(bank.Fragen, f)
		}
	case "loeschen":
		id := strings.TrimSpace(input.ID)
		if id == "" {
			http.Error(w, "Welche Frage?", http.StatusBadRequest)
			return
		}
		rest := make([]QuizFrage, 0, len(bank.Fragen))
		for _, f := range bank.Fragen {
			if f.ID != id {
				rest = append(rest, f)
			}
		}
		if len(rest) == len(bank.Fragen) {
			http.Error(w, "Diese Frage gibt es nicht.", http.StatusNotFound)
			return
		}
		bank.Fragen = rest
	case "kategorie":
		if input.Kategorie == nil {
			http.Error(w, "Keine Kategorie mitgeschickt.", http.StatusBadRequest)
			return
		}
		k := QuizKategorie{ID: strings.TrimSpace(input.Kategorie.ID), Name: strings.TrimSpace(input.Kategorie.Name)}
		k.ID = strings.ToLower(strings.ReplaceAll(k.ID, " ", "-"))
		if k.ID == "" || k.Name == "" {
			http.Error(w, "Kategorie braucht Kennung und Namen.", http.StatusBadRequest)
			return
		}
		for _, vorhanden := range bank.Kategorien {
			if vorhanden.ID == k.ID {
				http.Error(w, "Diese Kategorie gibt es schon.", http.StatusConflict)
				return
			}
		}
		bank.Kategorien = append(bank.Kategorien, k)
	case "vorschau":
		// Live-Vorschau für eine noch nicht gespeicherte Frage: derselbe
		// Weg wie im Spiel, aber ohne zu schreiben.
		if input.Frage == nil {
			http.Error(w, "Keine Frage mitgeschickt.", http.StatusBadRequest)
			return
		}
		f := quizFragePutzt(*input.Frage)
		optionen, typ, err := quizOptionen(f, quizAntwortPool(bank))
		v := QuizFrageVorschau{Typ: typ, TypName: quizTypName_(typ), Antworten: optionen}
		if err != nil {
			v.Fehler = err.Error()
		} else {
			for i, a := range optionen {
				if strings.EqualFold(a, strings.TrimSpace(f.Antwort)) {
					v.Richtig = i
				}
			}
		}
		jsonResponse(w, v)
		return
	case "import":
		if input.Ersetzen {
			bank.Fragen = nil
			if len(input.Kategorien) > 0 {
				bank.Kategorien = nil
			}
		}
		// Kategorien zuerst: Fragen dürfen nur bekannte Kategorien tragen.
		for _, k := range input.Kategorien {
			k.Name = strings.TrimSpace(k.Name)
			k.ID = strings.ToLower(strings.TrimSpace(k.ID))
			if k.ID == "" || k.Name == "" {
				continue
			}
			da := false
			for _, vorhanden := range bank.Kategorien {
				if vorhanden.ID == k.ID {
					da = true
					break
				}
			}
			if !da {
				bank.Kategorien = append(bank.Kategorien, k)
			}
		}
		for _, f := range input.Fragen {
			f = quizFragePutzt(f)
			gefunden := false
			for i := range bank.Fragen {
				if bank.Fragen[i].ID == f.ID {
					bank.Fragen[i] = f
					gefunden = true
					break
				}
			}
			if !gefunden {
				bank.Fragen = append(bank.Fragen, f)
			}
		}
	default:
		http.Error(w, "Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	// Nach jeder Änderung gilt derselbe Maßstab: ein kaputter Katalog wird
	// nicht geschrieben.
	if err := quizBankPruefen(bank); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := quizBankSpeichern(bank); err != nil {
		http.Error(w, "Katalog ließ sich nicht speichern: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, quizBankSicht(bank))
}
