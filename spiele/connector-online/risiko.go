package main

// Risiko-Regelwerk für den Server (Phase 2, 22.11.2026).
//
// Alles, was den Spielstand ändert, steht hier – der Browser schickt nur
// Absichten („greife X von Y an“), nie einen fertigen Stand. Damit prüft der
// Connector jede Aktion selbst: Reihenfolge, Phase, Besitz, Nachbarschaft,
// Armeen, Karten. Die Karte (47 Gebiete) steht in risiko_gebiete.go.

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
)

const (
	risikoPhaseAufstellen  = "aufstellen"
	risikoPhaseVerstaerken = "verstaerken"
	risikoPhaseAngreifen   = "angreifen"
	risikoPhaseVerschieben = "verschieben"

	risikoStatusLobby    = "invited"
	risikoStatusAktiv    = "active"
	risikoStatusBeendet  = "finished"
	risikoMaxSpieler     = 6
	risikoMinSpieler     = 2
	risikoLogMax         = 80
	risikoStartArmeen2   = 40
	risikoStartArmeen3   = 35
	risikoStartArmeen4   = 30
	risikoStartArmeen5   = 25
	risikoStartArmeen6   = 20
	risikoStandMax       = 256 * 1024
	risikoSatzWertLetzte = 60
)

var risikoSatzWerte = []int{4, 6, 8, 10, 12, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60}

var risikoFarben = []string{"#c8684f", "#4f7fa8", "#c9a44f", "#5f9160", "#8a74b4", "#b07d55"}

var risikoModi = map[string]string{
	"welt":       "Weltherrschaft",
	"schnell":    "Schnellspiel",
	"kontinent":  "Kontinentkrieg",
	"mission":    "Geheime Mission",
	"hauptstadt": "Hauptstadtkrieg",
	"team":       "Teamkrieg",
}

// RisikoGebiet ist ein Feld auf der Karte.
type RisikoGebiet struct {
	Besitzer int `json:"besitzer"`
	Armeen   int `json:"armeen"`
}

// RisikoKarte ist eine Gebietskarte; Gebiet ist bei Jokern leer.
type RisikoKarte struct {
	ID     string `json:"id"`
	Gebiet string `json:"gebiet,omitempty"`
	Sym    string `json:"sym"`
}

// RisikoSpieler ist ein Sitz. ID ist bei Menschen die Gerätekennung, bei Bots
// bot-1 … bot-6.
type RisikoSpieler struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Farbe         string        `json:"farbe"`
	Bot           bool          `json:"bot"`
	Stufe         string        `json:"stufe,omitempty"`
	Wesen         string        `json:"wesen,omitempty"`
	Hand          []RisikoKarte `json:"hand,omitempty"`
	Ausgeschieden bool          `json:"ausgeschieden"`
	Team          int           `json:"team"`
	StartPool     int           `json:"startPool,omitempty"`
	Bereit        bool          `json:"bereit"`
}

// RisikoMission ist das persönliche Ziel im Modus „mission“.
type RisikoMission struct {
	Art     string   `json:"art"`
	Spieler int      `json:"spieler,omitempty"`
	Kont    []string `json:"kont,omitempty"`
	Anzahl  int      `json:"anzahl,omitempty"`
	Text    string   `json:"text"`
}

// RisikoKampf ist das Ergebnis einer Angriffsrunde (für Anzeige/Animation).
type RisikoKampf struct {
	Von       string `json:"von"`
	Nach      string `json:"nach"`
	AW        []int  `json:"aw"`
	DW        []int  `json:"dw"`
	VerlustA  int    `json:"verlustA"`
	VerlustV  int    `json:"verlustV"`
	Erobert   bool   `json:"erobert"`
	Vorher    int    `json:"vorher"`
	Besitzer  int    `json:"besitzer"`
	ArmeenVon int    `json:"armeenVon"`
}

// RisikoNachruecken merkt sich eine gerade eroberte Nachbarschaft.
type RisikoNachruecken struct {
	Von  string `json:"von"`
	Nach string `json:"nach"`
	Max  int    `json:"max"`
}

type RisikoEintrag struct {
	Text string `json:"text"`
}

// RisikoSpiel ist der vollständige, verbindliche Spielstand.
type RisikoSpiel struct {
	ID           string                   `json:"id"`
	Modus        string                   `json:"modus"`
	Spieler      []RisikoSpieler          `json:"spieler"`
	Gebiete      map[string]*RisikoGebiet `json:"gebiete"`
	Anzahl       int                      `json:"anzahl"`
	AmZug        int                      `json:"amZug"`
	Phase        string                   `json:"phase"`
	Pool         int                      `json:"pool"`
	KartenSatz   int                      `json:"kartenSatz"`
	Runde        int                      `json:"runde"`
	Status       string                   `json:"status"`
	Sieger       int                      `json:"sieger"`
	SiegGrund    string                   `json:"siegGrund"`
	Hauptstaedte []string                 `json:"hauptstaedte"`
	Missionen    []RisikoMission          `json:"missionen"`
	Teams        []int                    `json:"teams"`
	Stapel       []RisikoKarte            `json:"stapel"`
	Ablage       []RisikoKarte            `json:"ablage"`
	LetzterKampf *RisikoKampf             `json:"letzterKampf,omitempty"`
	Nachruecken  *RisikoNachruecken       `json:"nachruecken,omitempty"`
	Log          []RisikoEintrag          `json:"log"`
	ZugErobert   bool                     `json:"zugErobert"`
	Verschoben   bool                     `json:"verschoben"`
	HostID       string                   `json:"hostId"`
	Version      int                      `json:"version"`
	Pause        string                   `json:"pause"` // tempo: schnell|normal|langsam
	BotPlan      []RisikoAktion           `json:"-"`
	BotRechnet   *RisikoSpiel             `json:"-"`
	BotFrist     int64                    `json:"-"`
	offen        map[string]bool          // nur beim Aufbau benutzt
}

// RisikoAktion ist ein einzelner Bot-Schritt (auch für Pläne benutzt).
type RisikoAktion struct {
	Art    string `json:"art"`
	Gebiet string `json:"gebiet,omitempty"`
	Von    string `json:"von,omitempty"`
	Nach   string `json:"nach,omitempty"`
	Anzahl int    `json:"anzahl,omitempty"`
	Karten []int  `json:"karten,omitempty"`
}

var errRisikoVorbei = errors.New("Die Partie ist beendet.")

// risikoNeuesSpiel baut eine verteilte Partie auf (Aufstellung läuft danach
// über Aktionen der Spieler/Bots).
func risikoNeuesSpiel(einst RisikoEinstellung, r *rand.Rand) *RisikoSpiel {
	modus := einst.Modus
	if _, ok := risikoModi[modus]; !ok {
		modus = "welt"
	}
	anzahl := einst.Anzahl
	if anzahl < risikoMinSpieler {
		anzahl = risikoMinSpieler
	}
	if anzahl > risikoMaxSpieler {
		anzahl = risikoMaxSpieler
	}
	if modus == "team" && anzahl%2 == 1 {
		anzahl++
		if anzahl > risikoMaxSpieler {
			anzahl = risikoMaxSpieler
		}
	}
	g := &RisikoSpiel{
		ID:      randomCode(),
		Modus:   modus,
		Spieler: make([]RisikoSpieler, anzahl),
		Gebiete: make(map[string]*RisikoGebiet, len(risikoGebiete)),
		Phase:   risikoPhaseAufstellen,
		Status:  risikoStatusAktiv,
		Sieger:  -1,
		Runde:   1,
		Version: 1,
		Pause:   "normal",
	}
	if einst.Pause != "" {
		g.Pause = einst.Pause
	}
	startArmeen := risikoStartVorrat(anzahl)
	for k := 0; k < anzahl; k++ {
		name := ""
		if k < len(einst.Namen) {
			name = strings.TrimSpace(einst.Namen[k])
		}
		if name == "" {
			name = fmt.Sprintf("Spieler %d", k+1)
		}
		bot := k < len(einst.Bots) && einst.Bots[k]
		stufe, wesen := "", ""
		if bot {
			if k < len(einst.Stufen) {
				stufe = risikoBotStufe(einst.Stufen[k])
			} else {
				stufe = botMittel
			}
			if k < len(einst.Wesen) {
				wesen = risikoBotWesen(einst.Wesen[k])
			} else {
				wesen = "ausgewogen"
			}
		} else {
			bot = false
		}
		g.Spieler[k] = RisikoSpieler{
			ID:        einst.IDs[k],
			Name:      name,
			Farbe:     risikoFarben[k%len(risikoFarben)],
			Bot:       bot,
			Stufe:     stufe,
			Wesen:     wesen,
			Team:      -1,
			StartPool: startArmeen,
			Bereit:    true,
		}
		if modus == "team" {
			g.Spieler[k].Team = k % 2
		}
	}

	risikoAufbau(g, r)
	return g
}

// risikoAufbau verteilt Karte, Armeen, Hauptstädte und Missionen auf die
// bereits feststehenden Sitze (Lobby-Start und Testspiele teilen sich diesen
// Weg).
func risikoAufbau(g *RisikoSpiel, r *rand.Rand) {
	n := len(g.Spieler)
	g.Gebiete = make(map[string]*RisikoGebiet, len(risikoGebiete))
	g.Hauptstaedte = nil
	g.Missionen = nil
	g.Teams = make([]int, n)
	for k := range g.Spieler {
		g.Spieler[k].Hand = nil
		g.Spieler[k].Ausgeschieden = false
		g.Teams[k] = g.Spieler[k].Team
	}

	// Gebiete reihum verteilen, je 3 Armeen.
	ids := make([]string, 0, len(risikoGebiete))
	for _, def := range risikoGebiete {
		ids = append(ids, def.ID)
	}
	r.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
	for i, id := range ids {
		g.Gebiete[id] = &RisikoGebiet{Besitzer: i % n, Armeen: 3}
	}

	// Hauptstädte: je Spieler ein eigenes Gebiet.
	for k := 0; k < n; k++ {
		eigene := risikoGebieteVon(g, k)
		g.Hauptstaedte = append(g.Hauptstaedte, eigene[r.Intn(len(eigene))])
	}
	// Missionen (nur im Modus „mission“ sichtbar, sonst trotzdem gezogen).
	for k := 0; k < n; k++ {
		g.Missionen = append(g.Missionen, risikoMissionFuer(g, k, r))
	}

	vorrat := risikoStartVorrat(n)
	for k := range g.Spieler {
		g.Spieler[k].StartPool = vorrat
	}

	g.Stapel = risikoKartenstapel(r)
	g.Ablage = nil
	g.AmZug = 0
	g.Phase = risikoPhaseAufstellen
	g.Pool = 0
	g.KartenSatz = 0
	g.Runde = 1
	g.Status = risikoStatusAktiv
	g.Sieger = -1
	g.SiegGrund = ""
	g.ZugErobert = false
	g.Verschoben = false
	g.Nachruecken = nil
	g.LetzterKampf = nil
	g.Log = nil
	g.BotPlan = nil
	g.BotRechnet = nil
	g.BotFrist = 0
	risikoLog(g, "Die Weltkarte ist verteilt – "+g.Spieler[0].Name+" setzt zuerst.")
}

// risikoStartVorrat nennt die Startarmeen je Spielerzahl.
func risikoStartVorrat(anzahl int) int {
	switch anzahl {
	case 2:
		return risikoStartArmeen2
	case 3:
		return risikoStartArmeen3
	case 4:
		return risikoStartArmeen4
	case 5:
		return risikoStartArmeen5
	}
	return risikoStartArmeen6
}

func risikoMissionFuer(g *RisikoSpiel, k int, r *rand.Rand) RisikoMission {
	arten := []string{"zerstoeren", "kontinente", "gebiete"}
	art := arten[r.Intn(len(arten))]
	if art == "zerstoeren" {
		var andere []int
		for i := range g.Spieler {
			if i != k {
				andere = append(andere, i)
			}
		}
		if len(andere) > 0 {
			ziel := andere[r.Intn(len(andere))]
			return RisikoMission{Art: art, Spieler: ziel,
				Text: "Zerstöre " + g.Spieler[ziel].Name + " (alle seine Gebiete erobern)."}
		}
	}
	if art == "kontinente" {
		ks := make([]string, 0, len(risikoKontinente))
		for id := range risikoKontinente {
			ks = append(ks, id)
		}
		sort.Strings(ks)
		r.Shuffle(len(ks), func(i, j int) { ks[i], ks[j] = ks[j], ks[i] })
		return RisikoMission{Art: art, Kont: ks[:2],
			Text: "Halte " + risikoKontinente[ks[0]].Name + " und " + risikoKontinente[ks[1]].Name + " vollständig."}
	}
	return RisikoMission{Art: "gebiete", Anzahl: 26, Text: "Halte 26 Gebiete gleichzeitig."}
}

func risikoKartenstapel(r *rand.Rand) []RisikoKarte {
	sym := []string{"infanterie", "kavallerie", "artillerie"}
	stapel := make([]RisikoKarte, 0, len(risikoGebiete)+2)
	for i, def := range risikoGebiete {
		stapel = append(stapel, RisikoKarte{ID: fmt.Sprintf("k%d", i), Gebiet: def.ID, Sym: sym[i%3]})
	}
	stapel = append(stapel, RisikoKarte{ID: "j1", Sym: "joker"}, RisikoKarte{ID: "j2", Sym: "joker"})
	r.Shuffle(len(stapel), func(i, j int) { stapel[i], stapel[j] = stapel[j], stapel[i] })
	return stapel
}

// ---------- Abfragen -------------------------------------------------------

func risikoGebieteVon(g *RisikoSpiel, k int) []string {
	var aus []string
	for _, def := range risikoGebiete {
		f := g.Gebiete[def.ID]
		if f != nil && f.Besitzer == k {
			aus = append(aus, def.ID)
		}
	}
	return aus
}

func risikoArmeenGesamt(g *RisikoSpiel, k int) int {
	summe := 0
	for _, def := range risikoGebiete {
		f := g.Gebiete[def.ID]
		if f != nil && f.Besitzer == k {
			summe += f.Armeen
		}
	}
	return summe
}

func risikoKontinentVoll(g *RisikoSpiel, kont string, k int) bool {
	for _, def := range risikoGebiete {
		if def.Kont == kont && g.Gebiete[def.ID].Besitzer != k {
			return false
		}
	}
	return true
}

func risikoVerstaerkung(g *RisikoSpiel, k int) int {
	land := len(risikoGebieteVon(g, k))
	basis := land / 3
	if basis < 3 {
		basis = 3
	}
	bonus := 0
	for kont, def := range risikoKontinente {
		if risikoKontinentVoll(g, kont, k) {
			bonus += def.Bonus
		}
	}
	return basis + bonus
}

func risikoSatzWert(g *RisikoSpiel) int {
	i := g.KartenSatz
	if i >= len(risikoSatzWerte) {
		i = len(risikoSatzWerte) - 1
	}
	return risikoSatzWerte[i]
}

func risikoIstSatz(karten []RisikoKarte) bool {
	if len(karten) != 3 {
		return false
	}
	joker := 0
	var rest []string
	for _, k := range karten {
		if k.Sym == "joker" {
			joker++
		} else {
			rest = append(rest, k.Sym)
		}
	}
	if len(rest) <= 2 {
		return true
	}
	if len(rest) != 3 {
		return false
	}
	if rest[0] == rest[1] && rest[1] == rest[2] {
		return true
	}
	return rest[0] != rest[1] && rest[1] != rest[2] && rest[0] != rest[2]
}

func risikoNachbar(von, nach string) bool {
	for _, n := range risikoGebietNach[von] {
		if n == nach {
			return true
		}
	}
	return false
}

func risikoPhaseText(g *RisikoSpiel) string {
	switch g.Phase {
	case risikoPhaseAufstellen:
		return "Aufstellen"
	case risikoPhaseVerstaerken:
		return "Verstärken"
	case risikoPhaseAngreifen:
		return "Angreifen"
	case risikoPhaseVerschieben:
		return "Verschieben"
	}
	return g.Phase
}

func risikoSpielerIndex(g *RisikoSpiel, id string) int {
	for i := range g.Spieler {
		if g.Spieler[i].ID == id {
			return i
		}
	}
	return -1
}

func risikoLog(g *RisikoSpiel, text string) {
	g.Log = append(g.Log, RisikoEintrag{Text: text})
	if len(g.Log) > risikoLogMax {
		g.Log = g.Log[len(g.Log)-risikoLogMax:]
	}
}

func risikoGeaendert(g *RisikoSpiel) {
	g.Version++
}

// ---------- Karten ---------------------------------------------------------

func risikoZieheKarte(g *RisikoSpiel, k int, r *rand.Rand) *RisikoKarte {
	if len(g.Stapel) == 0 {
		// Ablage neu mischen, erst dann ist der Stapel leer.
		if len(g.Ablage) == 0 {
			return nil
		}
		g.Stapel, g.Ablage = g.Ablage, nil
		r.Shuffle(len(g.Stapel), func(i, j int) { g.Stapel[i], g.Stapel[j] = g.Stapel[j], g.Stapel[i] })
	}
	karte := g.Stapel[len(g.Stapel)-1]
	g.Stapel = g.Stapel[:len(g.Stapel)-1]
	g.Spieler[k].Hand = append(g.Spieler[k].Hand, karte)
	return &karte
}

// risikoKartenTauschen tauscht drei Karten gegen Verstärkung.
func risikoKartenTauschen(g *RisikoSpiel, k int, indizes []int) (int, error) {
	if err := risikoAmZug(g, k); err != nil {
		return 0, err
	}
	if g.Phase != risikoPhaseVerstaerken {
		return 0, errors.New("Karten tauscht man beim Verstärken.")
	}
	if len(indizes) != 3 {
		return 0, errors.New("Es müssen genau drei Karten sein.")
	}
	gesehen := map[int]bool{}
	var karten []RisikoKarte
	for _, i := range indizes {
		if i < 0 || i >= len(g.Spieler[k].Hand) || gesehen[i] {
			return 0, errors.New("Diese Karten hat der Spieler nicht.")
		}
		gesehen[i] = true
		karten = append(karten, g.Spieler[k].Hand[i])
	}
	if !risikoIstSatz(karten) {
		return 0, errors.New("Diese drei Karten sind kein Satz.")
	}
	wert := risikoSatzWert(g)
	extra := 0
	for _, karte := range karten {
		if karte.Gebiet != "" && g.Gebiete[karte.Gebiet].Besitzer == k {
			extra += 2
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(indizes)))
	for _, i := range indizes {
		g.Ablage = append(g.Ablage, g.Spieler[k].Hand[i])
		g.Spieler[k].Hand = append(g.Spieler[k].Hand[:i], g.Spieler[k].Hand[i+1:]...)
	}
	g.KartenSatz++
	g.Pool += wert + extra
	risikoLog(g, fmt.Sprintf("%s tauscht einen Kartensatz: +%d Armeen.", g.Spieler[k].Name, wert+extra))
	risikoGeaendert(g)
	return wert + extra, nil
}

// ---------- Züge -----------------------------------------------------------

func risikoAmZug(g *RisikoSpiel, k int) error {
	if g.Status != risikoStatusAktiv {
		return errRisikoVorbei
	}
	if k != g.AmZug {
		return fmt.Errorf("Jetzt ist %s am Zug.", g.Spieler[g.AmZug].Name)
	}
	if g.Spieler[k].Ausgeschieden {
		return errors.New("Dieser Spieler ist ausgeschieden.")
	}
	return nil
}

// risikoSetzeArmee setzt eine Armee aus dem Vorrat (Aufstellen/Verstärken).
func risikoSetzeArmee(g *RisikoSpiel, k int, gebiet string) error {
	if err := risikoAmZug(g, k); err != nil {
		return err
	}
	if g.Phase != risikoPhaseAufstellen && g.Phase != risikoPhaseVerstaerken {
		return errors.New("Jetzt werden keine Armeen gesetzt.")
	}
	feld := g.Gebiete[gebiet]
	if feld == nil {
		return errors.New("Dieses Gebiet gibt es nicht.")
	}
	if feld.Besitzer != k {
		return errors.New("Nur eigene Gebiete dürfen verstärkt werden.")
	}
	if g.Phase == risikoPhaseVerstaerken && g.Pool <= 0 {
		return errors.New("Der Vorrat ist leer.")
	}
	pool := g.Pool
	if g.Phase == risikoPhaseAufstellen {
		pool = g.Spieler[k].StartPool
		if pool <= 0 {
			return errors.New("Dieser Spieler hat keine Startarmeen mehr.")
		}
		g.Spieler[k].StartPool--
	} else {
		g.Pool--
	}
	feld.Armeen++
	risikoLog(g, fmt.Sprintf("%s setzt eine Armee nach %s (%d).", g.Spieler[k].Name, risikoGebietName(gebiet), feld.Armeen))
	if g.Phase == risikoPhaseAufstellen && g.Spieler[k].StartPool == 0 {
		risikoNaechsterAufsteller(g)
	}
	risikoGeaendert(g)
	return nil
}

func risikoNaechsterAufsteller(g *RisikoSpiel) {
	for i := 1; i <= len(g.Spieler); i++ {
		k := (g.AmZug + i) % len(g.Spieler)
		if g.Spieler[k].StartPool > 0 {
			g.AmZug = k
			risikoGeaendert(g)
			return
		}
	}
	risikoAufstellungBeenden(g)
}

// risikoAufstellenFertig beendet die Startaufstellung, wenn der eigene Vorrat
// leer ist.
func risikoAufstellenFertig(g *RisikoSpiel, k int) error {
	if err := risikoAmZug(g, k); err != nil {
		return err
	}
	if g.Phase != risikoPhaseAufstellen {
		return errors.New("Die Aufstellung ist schon vorbei.")
	}
	if g.Spieler[k].StartPool > 0 {
		return fmt.Errorf("Es sind noch %d Startarmeen übrig.", g.Spieler[k].StartPool)
	}
	risikoAufstellungBeenden(g)
	risikoGeaendert(g)
	return nil
}

func risikoAufstellungBeenden(g *RisikoSpiel) {
	for k := range g.Spieler {
		g.Spieler[k].StartPool = 0
	}
	g.AmZug = 0
	for i := 0; i < len(g.Spieler); i++ {
		if !g.Spieler[i].Ausgeschieden && len(risikoGebieteVon(g, i)) > 0 {
			g.AmZug = i
			break
		}
	}
	g.Phase = risikoPhaseVerstaerken
	g.Pool = risikoVerstaerkung(g, g.AmZug)
	g.ZugErobert = false
	g.Verschoben = false
	g.Nachruecken = nil
	risikoLog(g, fmt.Sprintf("Aufstellung fertig – %s beginnt mit %d Armeen.", g.Spieler[g.AmZug].Name, g.Pool))
}

func risikoEndeVerstaerken(g *RisikoSpiel, k int) error {
	if err := risikoAmZug(g, k); err != nil {
		return err
	}
	if g.Phase != risikoPhaseVerstaerken {
		return errors.New("Gerade wird nicht verstärkt.")
	}
	if g.Pool > 0 {
		return fmt.Errorf("Es sind noch %d Armeen im Vorrat.", g.Pool)
	}
	g.Phase = risikoPhaseAngreifen
	risikoLog(g, g.Spieler[k].Name+" beginnt anzugreifen.")
	risikoGeaendert(g)
	return nil
}

func risikoEndeAngriff(g *RisikoSpiel, k int) error {
	if err := risikoAmZug(g, k); err != nil {
		return err
	}
	if g.Phase != risikoPhaseAngreifen {
		return errors.New("Gerade wird nicht angegriffen.")
	}
	if g.Nachruecken != nil {
		return errors.New("Zuerst Armeen nachrücken.")
	}
	g.Phase = risikoPhaseVerschieben
	risikoLog(g, g.Spieler[k].Name+" verschiebt Armeen.")
	risikoGeaendert(g)
	return nil
}

// risikoAngriff würfelt eine Runde und liefert das Ergebnis.
func risikoAngriff(g *RisikoSpiel, k int, von, nach string, r *rand.Rand) (RisikoKampf, error) {
	var kampf RisikoKampf
	if err := risikoAmZug(g, k); err != nil {
		return kampf, err
	}
	if g.Phase != risikoPhaseAngreifen {
		return kampf, errors.New("Gerade wird nicht angegriffen.")
	}
	if g.Nachruecken != nil {
		return kampf, errors.New("Zuerst Armeen nachrücken.")
	}
	a, b := g.Gebiete[von], g.Gebiete[nach]
	if a == nil || b == nil {
		return kampf, errors.New("Dieses Gebiet gibt es nicht.")
	}
	if a.Besitzer != k {
		return kampf, errors.New("Dieses Gebiet gehört nicht dem Angreifer.")
	}
	if b.Besitzer == k {
		return kampf, errors.New("Das Ziel gehört schon dem Angreifer.")
	}
	if !risikoNachbar(von, nach) {
		return kampf, errors.New("Die Gebiete grenzen nicht aneinander.")
	}
	if a.Armeen < 2 {
		return kampf, errors.New("Zum Angreifen braucht das Gebiet mindestens zwei Armeen.")
	}
	ka := a.Armeen - 1
	if ka > 3 {
		ka = 3
	}
	kd := b.Armeen
	if kd > 2 {
		kd = 2
	}
	kampf = RisikoKampf{Von: von, Nach: nach, Vorher: b.Besitzer, AW: risikoWuerfle(ka, r), DW: risikoWuerfle(kd, r)}
	va, vv := risikoVergleiche(kampf.AW, kampf.DW)
	kampf.VerlustA, kampf.VerlustV = va, vv
	a.Armeen -= va
	b.Armeen -= vv
	if b.Armeen <= 0 {
		b.Besitzer = k
		b.Armeen = 0
		kampf.Erobert = true
		risikoLog(g, fmt.Sprintf("%s erobert %s von %s.", g.Spieler[k].Name, risikoGebietName(nach), g.Spieler[kampf.Vorher].Name))
		if !g.ZugErobert {
			g.ZugErobert = true
			if karte := risikoZieheKarte(g, k, r); karte != nil {
				risikoLog(g, fmt.Sprintf("%s zieht eine Karte (%s).", g.Spieler[k].Name, risikoSymbolName(karte.Sym)))
			}
		}
		if len(risikoGebieteVon(g, kampf.Vorher)) == 0 {
			risikoAusscheiden(g, kampf.Vorher, k)
		}
		g.Nachruecken = &RisikoNachruecken{Von: von, Nach: nach, Max: a.Armeen - 1}
		if g.Nachruecken.Max < 1 {
			g.Nachruecken = nil
		}
	} else {
		risikoLog(g, fmt.Sprintf("%s greift %s an: %s gegen %s – Verluste %d:%d.",
			g.Spieler[k].Name, risikoGebietName(nach), risikoDiceText(kampf.AW), risikoDiceText(kampf.DW), va, vv))
	}
	kampf.Besitzer = b.Besitzer
	kampf.ArmeenVon = a.Armeen
	g.LetzterKampf = &kampf
	if !risikoPruefeSieg(g) {
		risikoGeaendert(g)
	}
	return kampf, nil
}

func risikoNachrueckenArmeen(g *RisikoSpiel, k int, anzahl int) error {
	if err := risikoAmZug(g, k); err != nil {
		return err
	}
	if g.Nachruecken == nil {
		return errors.New("Es ist nichts nachzurücken.")
	}
	n := anzahl
	if n < 1 {
		n = 1
	}
	if n > g.Nachruecken.Max {
		n = g.Nachruecken.Max
	}
	g.Gebiete[g.Nachruecken.Von].Armeen -= n
	g.Gebiete[g.Nachruecken.Nach].Armeen += n
	risikoLog(g, fmt.Sprintf("→ %d Armeen rücken nach %s nach.", n, risikoGebietName(g.Nachruecken.Nach)))
	g.Nachruecken = nil
	risikoPruefeSieg(g)
	risikoGeaendert(g)
	return nil
}

func risikoVerschiebe(g *RisikoSpiel, k int, von, nach string, anzahl int) error {
	if err := risikoAmZug(g, k); err != nil {
		return err
	}
	if g.Phase != risikoPhaseVerschieben {
		return errors.New("Verschoben wird am Ende des Zuges.")
	}
	if g.Verschoben {
		return errors.New("In diesem Zug wurde schon verschoben.")
	}
	a, b := g.Gebiete[von], g.Gebiete[nach]
	if a == nil || b == nil {
		return errors.New("Dieses Gebiet gibt es nicht.")
	}
	if a.Besitzer != k || b.Besitzer != k {
		return errors.New("Beide Gebiete müssen dem Spieler gehören.")
	}
	if !risikoNachbar(von, nach) {
		return errors.New("Die Gebiete grenzen nicht aneinander.")
	}
	if a.Armeen < 2 {
		return errors.New("Es muss mindestens eine Armee zurückbleiben.")
	}
	n := anzahl
	if n < 1 {
		n = 1
	}
	if n > a.Armeen-1 {
		n = a.Armeen - 1
	}
	a.Armeen -= n
	b.Armeen += n
	g.Verschoben = true
	risikoLog(g, fmt.Sprintf("%s verschiebt %d Armeen nach %s.", g.Spieler[k].Name, n, risikoGebietName(nach)))
	risikoGeaendert(g)
	return nil
}

func risikoZugEnde(g *RisikoSpiel, k int) error {
	if err := risikoAmZug(g, k); err != nil {
		return err
	}
	if g.Phase == risikoPhaseAufstellen {
		return errors.New("Erst die Startarmeen setzen.")
	}
	if g.Nachruecken != nil {
		return errors.New("Zuerst Armeen nachrücken.")
	}
	if risikoPruefeSieg(g) {
		return nil
	}
	naechster := -1
	for i := 1; i <= len(g.Spieler); i++ {
		nk := (g.AmZug + i) % len(g.Spieler)
		if !g.Spieler[nk].Ausgeschieden && len(risikoGebieteVon(g, nk)) > 0 {
			naechster = nk
			break
		}
	}
	if naechster < 0 {
		return errors.New("Es ist niemand mehr am Tisch.")
	}
	if naechster <= g.AmZug {
		g.Runde++
	}
	risikoBeginneZug(g, naechster)
	risikoGeaendert(g)
	return nil
}

func risikoBeginneZug(g *RisikoSpiel, k int) {
	g.AmZug = k
	g.Phase = risikoPhaseVerstaerken
	g.Pool = risikoVerstaerkung(g, k)
	g.ZugErobert = false
	g.Verschoben = false
	g.Nachruecken = nil
	risikoLog(g, fmt.Sprintf("Runde %d: %s verstärkt mit %d Armeen.", g.Runde, g.Spieler[k].Name, g.Pool))
}

func risikoAusscheiden(g *RisikoSpiel, k, sieger int) {
	if g.Spieler[k].Ausgeschieden {
		return
	}
	g.Spieler[k].Ausgeschieden = true
	karten := append([]RisikoKarte(nil), g.Spieler[k].Hand...)
	g.Spieler[k].Hand = nil
	if sieger >= 0 {
		g.Spieler[sieger].Hand = append(g.Spieler[sieger].Hand, karten...)
	}
	risikoLog(g, fmt.Sprintf("%s ist ausgeschieden. %s erhält %d Karten.",
		g.Spieler[k].Name, g.Spieler[sieger].Name, len(karten)))
}

func risikoWuerfle(n int, r *rand.Rand) []int {
	aus := make([]int, n)
	for i := range aus {
		aus[i] = 1 + r.Intn(6)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(aus)))
	return aus
}

// risikoVergleiche: höchster gegen höchsten, Gleichstand für die Verteidigung.
func risikoVergleiche(a, d []int) (verlustA, verlustV int) {
	paar := len(a)
	if len(d) < paar {
		paar = len(d)
	}
	for i := 0; i < paar; i++ {
		if a[i] > d[i] {
			verlustV++
		} else {
			verlustA++
		}
	}
	return
}

func risikoDiceText(w []int) string {
	teile := make([]string, len(w))
	for i, v := range w {
		teile[i] = fmt.Sprint(v)
	}
	return strings.Join(teile, ",")
}

func risikoSymbolName(sym string) string {
	switch sym {
	case "infanterie":
		return "Infanterie"
	case "kavallerie":
		return "Kavallerie"
	case "artillerie":
		return "Artillerie"
	}
	return "Joker"
}

func risikoGebietName(id string) string {
	return risikoGebieteDef(id).Name
}

func risikoGebieteDef(id string) risikoGebietDef {
	for _, def := range risikoGebiete {
		if def.ID == id {
			return def
		}
	}
	return risikoGebietDef{ID: id, Name: id}
}

// ---------- Sieg -----------------------------------------------------------

// risikoAufgeben beendet eine laufende Partie auf Wunsch eines Spielers.
// Sieger ist, wer danach die meisten Gebiete hält (bei Gleichstand der
// erste); so bleibt auch eine Partie mit mehr als zwei Spielern sauber.
func risikoAufgeben(g *RisikoSpiel, k int) error {
	if g == nil || g.Status != risikoStatusAktiv {
		return errors.New("Es läuft gerade keine aktive Risiko-Partie.")
	}
	if k < 0 || k >= len(g.Spieler) {
		return errors.New("Du gehörst nicht zu dieser Partie.")
	}
	sieger, meiste := -1, -1
	for i := range g.Spieler {
		if i == k || g.Spieler[i].Ausgeschieden {
			continue
		}
		if n := len(risikoGebieteVon(g, i)); n > meiste {
			sieger, meiste = i, n
		}
	}
	g.Status = risikoStatusBeendet
	g.Sieger = sieger
	g.SiegGrund = g.Spieler[k].Name + " gibt auf."
	risikoLog(g, "SIEG: "+g.SiegGrund)
	risikoGeaendert(g)
	return nil
}

func risikoPruefeSieg(g *RisikoSpiel) bool {
	if g.Status == risikoStatusBeendet {
		return true
	}
	for k := range g.Spieler {
		if g.Spieler[k].Ausgeschieden {
			continue
		}
		if grund := risikoSiegGrund(g, k); grund != "" {
			g.Status = risikoStatusBeendet
			g.Sieger = k
			g.SiegGrund = grund
			g.Phase = risikoPhaseVerschieben
			risikoLog(g, "SIEG: "+grund)
			risikoGeaendert(g)
			return true
		}
	}
	return false
}

func risikoSiegGrund(g *RisikoSpiel, k int) string {
	land := len(risikoGebieteVon(g, k))
	aktiv := 0
	for i := range g.Spieler {
		if !g.Spieler[i].Ausgeschieden && len(risikoGebieteVon(g, i)) > 0 {
			aktiv++
		}
	}
	if aktiv == 1 {
		return g.Spieler[k].Name + " hält als Einziger noch Gebiete und gewinnt."
	}
	switch g.Modus {
	case "schnell":
		if land >= 24 {
			return g.Spieler[k].Name + " hält 24 Gebiete."
		}
	case "kontinent":
		voll := 0
		for kont := range risikoKontinente {
			if risikoKontinentVoll(g, kont, k) {
				voll++
			}
		}
		if voll >= 3 {
			return g.Spieler[k].Name + " hält drei vollständige Kontinente."
		}
	case "mission":
		if m := g.Missionen[k]; risikoMissionErfuellt(g, k, m) {
			return g.Spieler[k].Name + " erfüllt seine Mission: " + m.Text
		}
	case "hauptstadt":
		alle := true
		for _, id := range g.Hauptstaedte {
			if g.Gebiete[id].Besitzer != k {
				alle = false
				break
			}
		}
		if alle && len(g.Hauptstaedte) > 0 {
			return g.Spieler[k].Name + " hält alle Hauptstädte."
		}
	case "team":
		mitglieder := 0
		for i := range g.Spieler {
			if g.Spieler[i].Team == g.Spieler[k].Team && g.Gebiete[g.Hauptstaedte[i]].Besitzer == k {
				mitglieder++
			}
		}
		_ = mitglieder
		alle := true
		for _, def := range risikoGebiete {
			if g.Spieler[g.Gebiete[def.ID].Besitzer].Team != g.Spieler[k].Team {
				alle = false
				break
			}
		}
		if alle {
			return "Team " + risikoTeamName(g.Spieler[k].Team) + " hält die ganze Weltkarte."
		}
	default:
		if land == len(risikoGebiete) {
			return g.Spieler[k].Name + " hält die ganze Weltkarte."
		}
	}
	return ""
}

func risikoMissionErfuellt(g *RisikoSpiel, k int, m RisikoMission) bool {
	switch m.Art {
	case "zerstoeren":
		return g.Spieler[m.Spieler].Ausgeschieden || len(risikoGebieteVon(g, m.Spieler)) == 0
	case "kontinente":
		for _, kont := range m.Kont {
			if !risikoKontinentVoll(g, kont, k) {
				return false
			}
		}
		return true
	case "gebiete":
		return len(risikoGebieteVon(g, k)) >= m.Anzahl
	}
	return false
}

func risikoTeamName(t int) string {
	if t == 0 {
		return "A"
	}
	return "B"
}

// ---------- Würfelwahrscheinlichkeit (für die Bots) ------------------------

var (
	risikoWurfMutex sync.Mutex
	risikoWurfCache = map[[2]int]map[[2]int]float64{}
	risikoPMutex    sync.Mutex
	risikoPCache    = map[[2]int]float64{}
)

// risikoWurfVerteilung liefert für Angreifer-/Verteidigerwürfel die
// Wahrscheinlichkeit jedes Verlustpaars (Angreifer, Verteidiger).
func risikoWurfVerteilung(ka, kd int) map[[2]int]float64 {
	key := [2]int{ka, kd}
	risikoWurfMutex.Lock()
	if v, ok := risikoWurfCache[key]; ok {
		risikoWurfMutex.Unlock()
		return v
	}
	risikoWurfMutex.Unlock()

	verteilung := map[[2]int]float64{}
	gesamt := 1
	for i := 0; i < ka+kd; i++ {
		gesamt *= 6
	}
	roll := make([]int, ka+kd)
	var lauf func(pos int)
	lauf = func(pos int) {
		if pos == ka+kd {
			a := append([]int(nil), roll[:ka]...)
			d := append([]int(nil), roll[ka:]...)
			sort.Sort(sort.Reverse(sort.IntSlice(a)))
			sort.Sort(sort.Reverse(sort.IntSlice(d)))
			va, vv := risikoVergleiche(a, d)
			verteilung[[2]int{va, vv}] += 1 / float64(gesamt)
			return
		}
		for v := 1; v <= 6; v++ {
			roll[pos] = v
			lauf(pos + 1)
		}
	}
	lauf(0)

	risikoWurfMutex.Lock()
	risikoWurfCache[key] = verteilung
	risikoWurfMutex.Unlock()
	return verteilung
}

// risikoSiegChance ist die exakte Chance, dass a Armeen d Verteidiger
// aufreiben (ohne Verstärkung dazwischen).
func risikoSiegChance(a, d int) float64 {
	if d <= 0 {
		return 1
	}
	if a <= 1 {
		return 0
	}
	key := [2]int{a, d}
	risikoPMutex.Lock()
	if v, ok := risikoPCache[key]; ok {
		risikoPMutex.Unlock()
		return v
	}
	risikoPMutex.Unlock()

	ka := a - 1
	if ka > 3 {
		ka = 3
	}
	kd := d
	if kd > 2 {
		kd = 2
	}
	summe := 0.0
	for paar, p := range risikoWurfVerteilung(ka, kd) {
		summe += p * risikoSiegChance(a-paar[0], d-paar[1])
	}
	risikoPMutex.Lock()
	risikoPCache[key] = summe
	risikoPMutex.Unlock()
	return summe
}

// ---------- Kopie für Bot-Pläne --------------------------------------------

func risikoKopie(g *RisikoSpiel) *RisikoSpiel {
	if g == nil {
		return nil
	}
	c := *g
	c.Spieler = make([]RisikoSpieler, len(g.Spieler))
	for i, sp := range g.Spieler {
		c.Spieler[i] = sp
		c.Spieler[i].Hand = append([]RisikoKarte(nil), sp.Hand...)
	}
	c.Gebiete = make(map[string]*RisikoGebiet, len(g.Gebiete))
	for id, f := range g.Gebiete {
		kopie := *f
		c.Gebiete[id] = &kopie
	}
	c.Hauptstaedte = append([]string(nil), g.Hauptstaedte...)
	c.Missionen = append([]RisikoMission(nil), g.Missionen...)
	c.Teams = append([]int(nil), g.Teams...)
	c.Stapel = append([]RisikoKarte(nil), g.Stapel...)
	c.Ablage = append([]RisikoKarte(nil), g.Ablage...)
	c.Log = append([]RisikoEintrag(nil), g.Log...)
	if g.Nachruecken != nil {
		nr := *g.Nachruecken
		c.Nachruecken = &nr
	}
	if g.LetzterKampf != nil {
		kampf := *g.LetzterKampf
		c.LetzterKampf = &kampf
	}
	c.BotPlan = nil
	c.BotRechnet = nil
	return &c
}
