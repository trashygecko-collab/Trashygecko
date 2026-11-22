package main

// Würfelspiele online (spiele/wuerfelspiele).
//
// Die 18 Würfelspiele bringen ihre Regeln als JavaScript-Module mit
// (start/sicht/aktion). Damit die Regeln nicht ein zweites Mal in Go stehen,
// lädt der Connector genau diese Dateien in eine kleine JavaScript-Maschine
// (goja) und führt sie hier aus. Gewürfelt wird über z.wurf() also auf dem
// Server, und ein Zug gilt nur, wenn ihn der Spieler macht, der am Zug ist.
// Die Spieler bekommen den Spielstand; gezeichnet wird mit demselben Rahmen
// wie lokal.
//
// Die Partie liegt in room.Wuerfel. Alle Zugriffe darauf laufen unter
// wuerfelMu (goja ist nicht nebenläufig). Sperrreihenfolge: erst mu (Raum),
// dann wuerfelMu – nie umgekehrt.

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// wuerfelKatalog sind die Würfelspiele, die der Connector anbietet. Die Namen
// stehen in den Modulen selbst (titel), hier steht nur die Kennung.
var wuerfelKatalog = []string{
	"beelzebub", "boese-drei", "boese-eins", "der-turm", "dufte-hundert",
	"eins-bis-sechzehn", "fenster", "gestrichene-zwoelf", "hin-und-her",
	"hohe-hausnummer", "jule", "paerchen", "plus-minus", "rentmeister",
	"streichhoelzer", "teure-sieben", "unter-oder-ueber", "wuerfelpoker",
}

func wuerfelBekannt(id string) bool {
	for _, k := range wuerfelKatalog {
		if k == id {
			return true
		}
	}
	return false
}

// --- Regeln laden -----------------------------------------------------------

// wuerfelVariante ist eine Zusatzregel, die das Regelmodul anbietet
// (varianten:[{id,text}]). Der Gastgeber wählt sie vor dem Start.
type wuerfelVariante struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type wuerfelRegeln struct {
	id        string
	vm        *goja.Runtime
	def       goja.Value
	titel     string
	min       int
	max       int
	varianten []wuerfelVariante
}

var (
	// wuerfelMu schützt Regelmaschinen und Partien.
	wuerfelMu     sync.Mutex
	wuerfelRegler = map[string]*wuerfelRegeln{}
)

// wuerfelUmgebung ist der kleinste Ersatz für einen Browser. Der Rahmen baut
// seine Seite erst bei DOMContentLoaded; hier bleibt document im Zustand
// "loading", damit genau das nicht passiert.
const wuerfelUmgebung = `
var window = this;
var document = {
  readyState: "loading",
  addEventListener: function () {},
  removeEventListener: function () {},
  createElement: function () {
    return {
      style: {}, classList: { add: function () {}, remove: function () {}, contains: function () { return false; } },
      dataset: {}, children: [],
      setAttribute: function () {}, removeAttribute: function () {},
      appendChild: function () {}, append: function () {}, insertBefore: function () {},
      addEventListener: function () {}, removeEventListener: function () {},
      querySelector: function () { return null; }, querySelectorAll: function () { return []; },
      replaceWith: function () {}, remove: function () {}, focus: function () {},
      textContent: "", innerHTML: "", hidden: false, value: ""
    };
  },
  querySelector: function () { return null; },
  querySelectorAll: function () { return []; },
  getElementById: function () { return null; },
  addEventListener: function () {},
  body: { appendChild: function () {}, textContent: "", classList: { add: function () {} } }
};
var localStorage = {
  getItem: function () { return null; },
  setItem: function () {}, removeItem: function () {}
};
var matchMedia = function () { return { matches: false, addEventListener: function () {} }; };
var setTimeout = function () { return 0; };
var clearTimeout = function () {};
`

// wuerfelOrdner findet die Regelmodule. Im Betrieb steht die Homepage-Wurzel
// fest (home); in den Tests läuft der Prozess im Ordner spiele/connector und
// greift über ../wuerfelspiele zu. Beides sind dieselben Dateien.
func wuerfelOrdner() (string, error) {
	var kandidaten []string
	if home != "" {
		kandidaten = append(kandidaten, filepath.Join(home, "spiele", "wuerfelspiele"))
	}
	if hier, err := os.Getwd(); err == nil {
		kandidaten = append(kandidaten,
			filepath.Join(hier, "..", "wuerfelspiele"),
			filepath.Join(hier, "spiele", "wuerfelspiele"),
			filepath.Join(hier, "wuerfelspiele"))
	}
	for _, k := range kandidaten {
		if info, err := os.Stat(filepath.Join(k, "rahmen.js")); err == nil && !info.IsDir() {
			return k, nil
		}
	}
	return "", errors.New("Regelmodule der Würfelspiele nicht gefunden")
}

func wuerfelDatei(teile ...string) (string, error) {
	ordner, err := wuerfelOrdner()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{ordner}, teile...)...), nil
}

// wuerfelRegelnHolen lädt Rahmen und Regelmodul eines Spiels einmal und behält
// die Maschine. Aufruf nur unter wuerfelMu.
func wuerfelRegelnHolen(spiel string) (*wuerfelRegeln, error) {
	if r := wuerfelRegler[spiel]; r != nil {
		return r, nil
	}
	if !wuerfelBekannt(spiel) {
		return nil, errors.New("unbekanntes Würfelspiel")
	}
	rahmenPfad, err := wuerfelDatei("rahmen.js")
	if err != nil {
		return nil, err
	}
	modulPfad, err := wuerfelDatei("spiele", spiel+".js")
	if err != nil {
		return nil, err
	}
	rahmen, err := os.ReadFile(rahmenPfad)
	if err != nil {
		return nil, err
	}
	modul, err := os.ReadFile(modulPfad)
	if err != nil {
		return nil, err
	}
	vm := goja.New()
	if _, err = vm.RunString(wuerfelUmgebung); err != nil {
		return nil, err
	}
	if _, err = vm.RunString(string(rahmen)); err != nil {
		return nil, err
	}
	if _, err = vm.RunString(string(modul)); err != nil {
		return nil, err
	}
	tgw := vm.Get("TGW")
	if tgw == nil || goja.IsUndefined(tgw) || goja.IsNull(tgw) {
		return nil, errors.New("Rahmen meldet sich nicht an")
	}
	def := tgw.ToObject(vm).Get("spiele").ToObject(vm).Get(spiel)
	if def == nil || goja.IsUndefined(def) || goja.IsNull(def) {
		return nil, errors.New("Regelmodul fehlt: " + spiel)
	}
	r := &wuerfelRegeln{id: spiel, vm: vm, def: def, min: 2, max: 6}
	o := def.ToObject(vm)
	if t := o.Get("titel"); t != nil && !goja.IsUndefined(t) {
		r.titel = t.String()
	}
	if s := o.Get("spieler"); s != nil && !goja.IsUndefined(s) && !goja.IsNull(s) {
		so := s.ToObject(vm)
		if m := so.Get("min"); m != nil && !goja.IsUndefined(m) {
			r.min = int(m.ToInteger())
		}
		if m := so.Get("max"); m != nil && !goja.IsUndefined(m) {
			r.max = int(m.ToInteger())
		}
	}
	// Online braucht es einen Gegner: Spiele, die allein gehen, verlangen am
	// Tisch trotzdem zwei Spieler. Das ist keine Spielregel, sondern die
	// Bedingung einer gemeinsamen Runde.
	if r.min < 2 {
		r.min = 2
	}
	if r.max < r.min {
		r.max = r.min
	}
	if vs := o.Get("varianten"); vs != nil && !goja.IsUndefined(vs) && !goja.IsNull(vs) {
		liste := vs.ToObject(vm)
		for i := 0; ; i++ {
			eintrag := liste.Get(strconv.Itoa(i))
			if eintrag == nil || goja.IsUndefined(eintrag) || goja.IsNull(eintrag) {
				break
			}
			eo := eintrag.ToObject(vm)
			id := strings.TrimSpace(eo.Get("id").String())
			if id == "" {
				continue
			}
			text := ""
			if t := eo.Get("text"); t != nil && !goja.IsUndefined(t) && !goja.IsNull(t) {
				text = t.String()
			}
			r.varianten = append(r.varianten, wuerfelVariante{ID: id, Text: text})
		}
	}
	wuerfelRegler[spiel] = r
	return r, nil
}

// regelFn holt eine Funktion des Regelmoduls (start, sicht, aktion).
func (p *wuerfelPartie) regelFn(name string) (goja.Callable, error) {
	f, ok := goja.AssertFunction(p.reg.def.ToObject(p.reg.vm).Get(name))
	if !ok {
		return nil, errors.New("Regelmodul " + p.Spiel + " hat kein " + name)
	}
	return f, nil
}

// rahmenFn holt eine Funktion des Rahmens (nur neuePartie).
func (p *wuerfelPartie) rahmenFn(name string) (goja.Callable, error) {
	vm := p.reg.vm
	f, ok := goja.AssertFunction(vm.Get("TGW").ToObject(vm).Get(name))
	if !ok {
		return nil, errors.New("Rahmen hat kein " + name)
	}
	return f, nil
}

// --- Partie -----------------------------------------------------------------

type wuerfelSitz struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type wuerfelPartie struct {
	ID     string // Kennung der Runde (bleibt über „Neue Runde“ gleich)
	Spiel  string
	Host   string
	Sitze  []wuerfelSitz
	Status string // invited, active, finished
	Ende   string
	reg    *wuerfelRegeln
	z      goja.Value
	// Varianten sind die Zusatzregeln, die der Gastgeber im Wartebereich
	// angekreuzt hat. Sie gehören zur Runde (nicht zum Spielstand), damit sie
	// beim Neuaufbau des Tisches – Join, Start, Neue Runde – erhalten bleiben.
	Varianten map[string]bool
	// Bot: Kennung („bot-1“), Stufe und die Zeit, zu der er zieht. Die
	// Kennung steht als Mitspieler im Spielstand, Zeit und Stufe bleiben am
	// Server.
	BotID    string
	BotStufe string
	BotFrist time.Time
}

// setzeVarianten übernimmt nur Regeln, die dieses Spiel wirklich anbietet.
func (p *wuerfelPartie) setzeVarianten(wahl map[string]bool) {
	gewaehlt := map[string]bool{}
	for _, v := range p.reg.varianten {
		if wahl[v.ID] {
			gewaehlt[v.ID] = true
		}
	}
	p.Varianten = gewaehlt
}

func wuerfelSaat() int64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UnixNano()
	}
	return int64(binary.LittleEndian.Uint64(b[:]) >> 1)
}

func (p *wuerfelPartie) sitz(id string) int {
	for i, s := range p.Sitze {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func (p *wuerfelPartie) dran() int {
	if p.z == nil {
		return 0
	}
	v := p.z.ToObject(p.reg.vm).Get("dran")
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return 0
	}
	return int(v.ToInteger())
}

// wurfAufbauen erzeugt den Spielstand aus den Sitzen. Das passiert beim
// Eröffnen und bei jeder Sitzänderung – vor dem Start geht so nichts verloren.
func (p *wuerfelPartie) wurfAufbauen() error {
	vm := p.reg.vm
	namen := make([]string, 0, len(p.Sitze))
	for _, s := range p.Sitze {
		namen = append(namen, s.Name)
	}
	fn, err := p.rahmenFn("neuePartie")
	if err != nil {
		return err
	}
	varianten := vm.NewObject()
	for id, an := range p.Varianten {
		if an {
			_ = varianten.Set(id, true)
		}
	}
	z, err := fn(goja.Undefined(), vm.ToValue(namen), varianten, vm.ToValue(wuerfelSaat()))
	if err != nil {
		return err
	}
	p.z = z
	p.Ende = ""
	return nil
}

// wuerfelSitzeFiltern lässt nur Spieler sitzen, die noch im Raum sind. Das
// passiert beim Start einer Runde – wer inzwischen gegangen ist, sitzt dann
// nicht mehr mit am Tisch. Während einer laufenden Runde bleibt der Platz
// dagegen stehen (kurze Verbindungspause, gesperrtes Handy).
func wuerfelSitzeFiltern(p *wuerfelPartie, anwesend map[string]bool) {
	behalten := p.Sitze[:0]
	for _, s := range p.Sitze {
		if anwesend[s.ID] {
			behalten = append(behalten, s)
		}
	}
	p.Sitze = behalten
	p.hostNachruecken()
}

// hostNachruecken gibt den Vorsitz an den ersten Sitz, wenn der Gastgeber
// nicht mehr am Tisch ist.
func (p *wuerfelPartie) hostNachruecken() {
	if p.Host == "" || len(p.Sitze) == 0 || p.sitz(p.Host) >= 0 {
		return
	}
	p.Host = p.Sitze[0].ID
}

// anwesendeSitze lässt den Bot am Tisch: Er steht nicht in room.Players und
// würde beim Filtern sonst als abwesend herausfallen. nil heißt „nicht
// filtern“ – so bleibt der Bot beim Start und bei „Neue Runde“ sitzen.
func (p *wuerfelPartie) anwesendeSitze(anwesend map[string]bool) map[string]bool {
	if p.BotID != "" {
		return nil
	}
	return anwesend
}

func (p *wuerfelPartie) starten(anwesend map[string]bool) error {
	if anwesend != nil {
		wuerfelSitzeFiltern(p, anwesend)
	}
	if err := p.wurfAufbauen(); err != nil {
		return err
	}
	fn, err := p.regelFn("start")
	if err != nil {
		return err
	}
	if _, err := fn(p.reg.def, p.z); err != nil {
		return err
	}
	p.Status = "active"
	return nil
}

// aktion gibt den Zug an das Regelmodul. false heißt: so nicht erlaubt.
func (p *wuerfelPartie) aktion(art string, daten json.RawMessage) (bool, error) {
	vm := p.reg.vm
	fn, err := p.regelFn("aktion")
	if err != nil {
		return false, err
	}
	var wert any
	if len(daten) > 0 && string(daten) != "null" {
		if err := json.Unmarshal(daten, &wert); err != nil {
			return false, err
		}
	}
	ergebnis, err := fn(p.reg.def, p.z, vm.ToValue(art), vm.ToValue(wert))
	if err != nil {
		return false, err
	}
	if ergebnis != nil && !goja.IsUndefined(ergebnis) && !goja.IsNull(ergebnis) {
		if b, ok := ergebnis.Export().(bool); ok && !b {
			return false, nil
		}
	}
	// Ende? Das Modul setzt z.ende auf einen Text.
	ende := p.z.ToObject(vm).Get("ende")
	if ende != nil && !goja.IsUndefined(ende) && !goja.IsNull(ende) {
		p.Status = "finished"
		if text := ende.ToObject(vm).Get("text"); text != nil && !goja.IsUndefined(text) {
			p.Ende = text.String()
		}
	}
	return true, nil
}

// sichtWerte ruft sicht(z) des Regelmoduls auf: Knopf, Auswahl, Hinweis.
func (p *wuerfelPartie) sichtWerte() json.RawMessage {
	fn, err := p.regelFn("sicht")
	if err != nil {
		return nil
	}
	wert, err := fn(p.reg.def, p.z)
	if err != nil {
		return nil
	}
	return wuerfelJson(p.reg.vm, wert)
}

func wuerfelJson(vm *goja.Runtime, wert goja.Value) json.RawMessage {
	if wert == nil || goja.IsUndefined(wert) || goja.IsNull(wert) {
		return nil
	}
	fn, ok := goja.AssertFunction(vm.Get("JSON").ToObject(vm).Get("stringify"))
	if !ok {
		return nil
	}
	out, err := fn(goja.Undefined(), wert)
	if err != nil || out == nil || goja.IsUndefined(out) {
		return nil
	}
	text := out.String()
	if text == "" || text == "null" {
		return nil
	}
	return json.RawMessage(text)
}

// --- Antwort an die Spieler -------------------------------------------------

type wuerfelSicht struct {
	ID      string          `json:"id"`
	Spiel   string          `json:"spiel"`
	Titel   string          `json:"titel"`
	Status  string          `json:"status"`
	Host    string          `json:"hostId"`
	Spieler []wuerfelSitz   `json:"spieler"`
	Dran    int             `json:"dran"`
	DranID  string          `json:"dranId"`
	Du      int             `json:"du"`
	Min     int             `json:"min"`
	Max     int             `json:"max"`
	Ende    string          `json:"ende,omitempty"`
	Zustand json.RawMessage `json:"zustand,omitempty"`
	// Sicht ist, was das Regelmodul gerade anbietet (Knopf, Auswahl, Hinweis).
	// Die Spielseite zeichnet daraus; Prüfungen können damit wie ein Spieler
	// ziehen, ohne die Regeln nachzubauen.
	Sicht json.RawMessage `json:"sicht,omitempty"`
	// Varianten sind die Zusatzregeln des Spiels samt Auswahl – der
	// Wartebereich zeigt daraus die Kästchen für den Gastgeber.
	Varianten []wuerfelVarianteSicht `json:"varianten,omitempty"`
}

// wuerfelVarianteSicht ist eine Zusatzregel mit dem Haken des Gastgebers.
type wuerfelVarianteSicht struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	An   bool   `json:"an"`
}

func (p *wuerfelPartie) sicht(fuerID string) wuerfelSicht {
	s := wuerfelSicht{
		ID:      p.ID,
		Spiel:   p.Spiel,
		Titel:   p.reg.titel,
		Status:  p.Status,
		Host:    p.Host,
		Spieler: p.Sitze,
		Dran:    p.dran(),
		Du:      p.sitz(fuerID),
		Min:     p.reg.min,
		Max:     p.reg.max,
		Ende:    p.Ende,
	}
	if s.Dran >= 0 && s.Dran < len(p.Sitze) {
		s.DranID = p.Sitze[s.Dran].ID
	}
	for _, v := range p.reg.varianten {
		s.Varianten = append(s.Varianten, wuerfelVarianteSicht{ID: v.ID, Text: v.Text, An: p.Varianten[v.ID]})
	}
	if p.Status != "invited" {
		// Zustand ist der Spielstand (Funktionen wie wurf/rng fallen beim
		// Serialisieren weg), Sicht der Vorschlag des Regelmoduls.
		s.Zustand = wuerfelJson(p.reg.vm, p.z)
		s.Sicht = p.sichtWerte()
	}
	return s
}

// --- Abgang und Umbenennen (leave.go) ---------------------------------------

// wuerfelRemoveSeat nimmt einen Spieler vom Tisch. Läuft die Runde, ist sie
// damit zu Ende: Ohne ihn könnte sie auf einen Zug warten, den niemand mehr
// macht. Vor dem Start geht es mit den Übrigen weiter.
func wuerfelRemoveSeat(g *wuerfelPartie, id, name string) (bool, bool, int) {
	if g == nil {
		return false, false, 0
	}
	wuerfelMu.Lock()
	defer wuerfelMu.Unlock()
	return wuerfelSitzWeg(g, id, name)
}

// wuerfelSitzWeg ist derselbe Schritt ohne Sperre – der Aufrufer hält sie.
func wuerfelSitzWeg(g *wuerfelPartie, id, name string) (bool, bool, int) {
	idx := g.sitz(id)
	if idx < 0 {
		return false, false, len(g.Sitze)
	}
	g.Sitze = append(g.Sitze[:idx], g.Sitze[idx+1:]...)
	// Bleibt nur der Bot übrig, räumt der Tisch sich selbst: Ohne den
	// Menschen, der ihn hergeholt hat, kann keine Runde mehr laufen.
	if wuerfelNurBots(g.Sitze) {
		g.Sitze = nil
		g.Status = "finished"
		return true, true, 0
	}
	g.hostNachruecken()
	if g.Status == "active" {
		g.Status = "finished"
		g.Ende = name + " hat den Tisch verlassen – die Runde ist zu Ende."
	}
	anzahl := len(g.Sitze)
	if anzahl < g.reg.min {
		g.Status = "finished"
		return true, true, anzahl
	}
	return true, g.Status == "finished", anzahl
}

// wuerfelNurBots sagt, ob am Tisch außer Bots niemand mehr sitzt.
func wuerfelNurBots(sitze []wuerfelSitz) bool {
	for _, s := range sitze {
		if !botIstFerngesteuert(s.ID) {
			return false
		}
	}
	return true
}

// wuerfelSpielerUmbenennen übernimmt einen neuen Anzeigenamen in Sitz und
// laufenden Spielstand (die Sitze stehen in derselben Reihenfolge wie
// z.spieler).
func wuerfelSpielerUmbenennen(p *wuerfelPartie, id, name string) {
	if p == nil {
		return
	}
	wuerfelMu.Lock()
	defer wuerfelMu.Unlock()
	i := p.sitz(id)
	if i < 0 {
		return
	}
	p.Sitze[i].Name = name
	if p.z == nil {
		return
	}
	spieler := p.z.ToObject(p.reg.vm).Get("spieler")
	if spieler == nil || goja.IsUndefined(spieler) || goja.IsNull(spieler) {
		return
	}
	platz := spieler.ToObject(p.reg.vm).Get(strconv.Itoa(i))
	if platz == nil || goja.IsUndefined(platz) || goja.IsNull(platz) {
		return
	}
	_ = platz.ToObject(p.reg.vm).Set("name", name)
}

// wuerfelLage liest Sitzplatz und Status unter der einen Sperre.
func wuerfelLage(p *wuerfelPartie, id string) (int, string) {
	if p == nil {
		return -1, ""
	}
	wuerfelMu.Lock()
	defer wuerfelMu.Unlock()
	return p.sitz(id), p.Status
}

// wuerfelAktiv sagt, ob der Spieler gerade in einer laufenden Runde sitzt.
func wuerfelAktiv(p *wuerfelPartie, id string) bool {
	sitz, status := wuerfelLage(p, id)
	return sitz >= 0 && status == "active"
}

// wuerfelSitzt sagt, ob der Spieler überhaupt an diesem Tisch sitzt.
func wuerfelSitzt(p *wuerfelPartie, id string) bool {
	sitz, _ := wuerfelLage(p, id)
	return sitz >= 0
}

// --- Endpunkt ---------------------------------------------------------------

func wuerfelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	aktion := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	var input struct {
		ID        string          `json:"id"`
		Spiel     string          `json:"spiel"`
		Art       string          `json:"art"`
		Daten     json.RawMessage `json:"daten"`
		Varianten map[string]bool `json:"varianten"`
		Stufe     string          `json:"stufe"`
	}
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Ungültige Daten", http.StatusBadRequest)
			return
		}
	} else {
		input.ID = strings.TrimSpace(r.URL.Query().Get("id"))
		input.Spiel = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("spiel")))
	}

	antwort := func(s *wuerfelSicht) {
		_ = json.NewEncoder(w).Encode(map[string]any{"game": s})
	}
	fehler := func(text string, nr int) {
		http.Error(w, text, nr)
	}
	// sende schickt den Stand des Tisches – oder null, wenn keiner läuft.
	var room *Room
	sende := func() {
		if room == nil || room.Wuerfel == nil {
			antwort(nil)
			return
		}
		// Der Bot zieht beim nächsten Abruf: Die Spielseite holt den Stand
		// alle 900 ms, eine eigene Schleife braucht es dafür nicht.
		wuerfelBotTreiben(room.Wuerfel, time.Now())
		s := room.Wuerfel.sicht(input.ID)
		antwort(&s)
	}
	// aufbauen erzeugt den Stand neu (Sitze geändert); im Fehlerfall ist die
	// Antwort schon geschrieben.
	aufbauen := func(p *wuerfelPartie) bool {
		if err := p.wurfAufbauen(); err != nil {
			fehler("Das Spiel lässt sich gerade nicht öffnen.", http.StatusInternalServerError)
			return false
		}
		return true
	}

	if code == "" || input.ID == "" {
		fehler("Raum und Spieler fehlen.", http.StatusBadRequest)
		return
	}

	// Erst der Raum (mu), dann die Regelmaschine (wuerfelMu). Die beiden
	// Sperren werden nie gleichzeitig gehalten.
	mu.Lock()
	room = rooms[code]
	var sitz wuerfelSitz
	anwesend := map[string]bool{}
	da := false
	if room != nil {
		room.cleanPlayers()
		if p, ok := room.Players[input.ID]; ok {
			sitz = wuerfelSitz{ID: input.ID, Name: p.Name}
			da = true
		}
		for id := range room.Players {
			anwesend[id] = true
		}
	}
	mu.Unlock()
	if room == nil {
		fehler("Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	if !da {
		fehler("Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
		return
	}

	wuerfelMu.Lock()
	defer wuerfelMu.Unlock()

	tisch := room.Wuerfel
	if tisch != nil && len(tisch.Sitze) == 0 {
		room.Wuerfel = nil
		tisch = nil
	}

	// Stand abholen: ohne Partie bleibt das Feld leer, die Seite zeigt dann
	// den Wartebereich.
	if aktion == "" || aktion == "stand" {
		sende()
		return
	}

	if aktion == "create" {
		if tisch != nil && tisch.Status != "finished" {
			// Schon offen: der Ersteller wollte nur seine Runde sehen.
			sende()
			return
		}
		spiel := strings.ToLower(strings.TrimSpace(input.Spiel))
		if spiel == "" && tisch != nil {
			spiel = tisch.Spiel
		}
		reg, err := wuerfelRegelnHolen(spiel)
		if err != nil {
			fehler("Dieses Würfelspiel lässt sich nicht öffnen: "+err.Error(), http.StatusInternalServerError)
			return
		}
		p := &wuerfelPartie{
			ID:     randomCode(),
			Spiel:  spiel,
			Host:   sitz.ID,
			Sitze:  []wuerfelSitz{sitz},
			Status: "invited",
			reg:    reg,
		}
		p.setzeVarianten(input.Varianten)
		if !aufbauen(p) {
			return
		}
		room.Wuerfel = p
		sende()
		return
	}

	if aktion == "bot" {
		// Gegen den Computer: Der Mensch sitzt, der Bot nimmt den zweiten
		// Platz, dann läuft die Runde sofort los. Der Bot muss nicht
		// beitreten – er sitzt schon.
		spiel := strings.ToLower(strings.TrimSpace(input.Spiel))
		if !wuerfelBekannt(spiel) {
			fehler("Dieses Würfelspiel gibt es nicht.", http.StatusBadRequest)
			return
		}
		if tisch != nil && tisch.Status != "finished" {
			fehler("Es läuft schon eine Würfelrunde.", http.StatusConflict)
			return
		}
		reg, err := wuerfelRegelnHolen(spiel)
		if err != nil {
			fehler("Dieses Würfelspiel lässt sich nicht öffnen: "+err.Error(), http.StatusInternalServerError)
			return
		}
		p := &wuerfelPartie{
			ID:    randomCode(),
			Spiel: spiel,
			Host:  sitz.ID,
			Sitze: []wuerfelSitz{sitz, wuerfelSitz{ID: botID(1), Name: botName(1)}},
			// Der Status wird von starten() auf „active“ gesetzt.
			Status:   "invited",
			reg:      reg,
			BotID:    botID(1),
			BotStufe: botStufe(input.Stufe),
		}
		p.setzeVarianten(input.Varianten)
		if err := p.starten(nil); err != nil {
			fehler("Das Spiel lässt sich gerade nicht starten.", http.StatusInternalServerError)
			return
		}
		room.Wuerfel = p
		sende()
		return
	}

	if tisch == nil {
		fehler("Es läuft keine Würfelrunde. Biete im Connector eine Runde an.", http.StatusNotFound)
		return
	}

	switch aktion {
	case "join":
		if tisch.sitz(sitz.ID) >= 0 {
			sende()
			return
		}
		if tisch.Status != "invited" {
			fehler("Die Runde läuft schon.", http.StatusConflict)
			return
		}
		if len(tisch.Sitze) >= tisch.reg.max {
			fehler("Der Tisch ist voll.", http.StatusConflict)
			return
		}
		tisch.Sitze = append(tisch.Sitze, sitz)
		if len(tisch.Sitze) >= tisch.reg.max {
			if err := tisch.starten(tisch.anwesendeSitze(anwesend)); err != nil {
				fehler("Das Spiel lässt sich gerade nicht starten.", http.StatusInternalServerError)
				return
			}
		} else if !aufbauen(tisch) {
			return
		}

	case "start", "neu":
		if tisch.Host != sitz.ID {
			fehler("Nur der Gastgeber startet die Runde.", http.StatusForbidden)
			return
		}
		if aktion == "start" && tisch.Status == "active" {
			sende()
			return
		}
		if aktion == "neu" && tisch.Status != "finished" {
			fehler("Die Runde läuft noch.", http.StatusConflict)
			return
		}
		if len(tisch.Sitze) < tisch.reg.min {
			fehler("Es fehlen noch Mitspieler.", http.StatusConflict)
			return
		}
		if err := tisch.starten(tisch.anwesendeSitze(anwesend)); err != nil {
			fehler("Das Spiel lässt sich gerade nicht starten.", http.StatusInternalServerError)
			return
		}

	case "varianten":
		// Nur der Gastgeber stellt die Zusatzregeln ein – und nur, solange
		// die Runde noch nicht läuft.
		if tisch.Host != sitz.ID {
			fehler("Nur der Gastgeber wählt die Zusatzregeln.", http.StatusForbidden)
			return
		}
		if tisch.Status != "invited" {
			fehler("Die Zusatzregeln stehen erst nach dem Start fest.", http.StatusConflict)
			return
		}
		tisch.setzeVarianten(input.Varianten)
		if !aufbauen(tisch) {
			return
		}

	case "verlassen":
		drin, fertig, anzahl := wuerfelSitzWeg(tisch, sitz.ID, sitz.Name)
		if !drin {
			fehler("Du sitzt nicht an diesem Tisch.", http.StatusConflict)
			return
		}
		if fertig || anzahl == 0 {
			room.Wuerfel = nil
		}

	case "aktion":
		if tisch.Status != "active" {
			fehler("Die Runde läuft nicht.", http.StatusConflict)
			return
		}
		if tisch.dran() != tisch.sitz(sitz.ID) {
			fehler("Du bist nicht am Zug.", http.StatusConflict)
			return
		}
		ok, err := tisch.aktion(strings.ToLower(strings.TrimSpace(input.Art)), input.Daten)
		if err != nil {
			fehler("Der Zug ließ sich nicht ausführen.", http.StatusInternalServerError)
			return
		}
		if !ok {
			fehler("Dieser Zug ist gerade nicht möglich.", http.StatusConflict)
			return
		}

	default:
		fehler("Unbekannte Aktion.", http.StatusBadRequest)
		return
	}

	sende()
}
