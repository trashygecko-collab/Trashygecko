package main

// Diese Datei trennt drei Dinge, die vorher in einem einzigen Platz im Raum
// steckten:
//
//  1. Spielerprofil – die dauerhafte Identität eines Spielers. Sie gilt über
//     Räume, Tabs und Wiederverbindungen hinweg.
//  2. Sitzung – die konkrete Verbindung eines Browsers oder Geräts. Ein Spieler
//     kann mehrere Sitzungen haben (zwei Tabs, Handy und Rechner).
//  3. Teilnahme – die Mitgliedschaft an einem Tisch. Sie ist eine Zugabe zur
//     Halle und ersetzt sie nie.
//
// Die Räume selbst bleiben unverändert: Der Platz im Raum (Room.Players)
// bleibt die Quelle für Chat, Partien und Anwesenheit. Die Sitzung ist die
// Identität dahinter – sie entscheidet, wer zu diesem Platz gehört.

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"
)

// sitzungFrist: So lange gilt eine Sitzung ohne Lebenszeichen. Sie ist bewusst
// großzügig – ein Handy liegt über Nacht in der Tasche und soll am nächsten
// Morgen derselbe Spieler sein. Der Platz im Raum verfällt weiterhin nach
// presenceTimeout; die Sitzung bleibt davon unberührt.
const sitzungFrist = 24 * time.Hour

// profilFrist: So lange bleibt ein Spielerprofil ohne Sitzung und ohne Platz
// erhalten, bevor es aus dem Speicher fällt. Danach ist es ein verwaister
// Eintrag – niemand kann sich mehr darauf berufen.
const profilFrist = 24 * time.Hour

// Spielerprofil ist die dauerhafte Spieleridentität.
type Spielerprofil struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Erstellt time.Time `json:"erstellt"`
	Zuletzt  time.Time `json:"zuletzt"`
}

// Sitzung ist die Verbindung eines Browsers oder Geräts zu einem Profil.
// Geheimnis ist das Token, das die Seite bei jedem Ruf mitschickt; Kennung ist
// eine kurze, ungefährliche Kennung zum Anzeigen und gezielten Abmelden.
type Sitzung struct {
	Kennung   string    `json:"kennung"`
	Geheimnis string    `json:"-"`
	SpielerID string    `json:"-"`
	Geraet    string    `json:"-"`
	Angelegt  time.Time `json:"angelegt"`
	Zuletzt   time.Time `json:"zuletzt"`
}

// Tisch ist eine laufende Partie dieses Spielers – die Teilnahme, getrennt von
// der Halle. Ohne Partie ist die Liste leer, die Halle bleibt trotzdem.
type Tisch struct {
	Raum   string `json:"raum"`
	Spiel  string `json:"spiel"`
	Privat bool   `json:"privat,omitempty"`
}

var (
	// spielerprofile hängt am Spieler (Kennung -> Profil).
	spielerprofile = map[string]*Spielerprofil{}
	// sitzungen hängt am Geheimnis (Token -> Sitzung). Der Schlüssel ist das
	// Geheimnis, damit die Prüfung eines Rufs ein einziger Nachschlag ist.
	sitzungen = map[string]*Sitzung{}
)

// neueSpielerID erzeugt eine Kennung für ein neues Profil. Sie entsteht nur
// bei einer ausdrücklichen Anmeldung (/api/sitzung), nie bei einer Abweisung.
func neueSpielerID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// kurzeKennung ist die öffentliche Kennung einer Sitzung: zufällig, aber kein
// Geheimnis. Sie steht in /api/ich, damit man einzelne Geräte abmelden kann.
func kurzeKennung() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return "S-" + strings.ToUpper(hex.EncodeToString(b))
}

// spielerMerken legt ein Profil an oder frischt es auf. Ein vorhandener Name
// bleibt stehen, solange kein neuer mitkommt – ein Lebenszeichen ohne Namen
// darf keinen Namen löschen. Muss unter mu laufen.
func spielerMerken(id, name string) *Spielerprofil {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	jetzt := time.Now()
	profil := spielerprofile[id]
	if profil == nil {
		profil = &Spielerprofil{ID: id, Erstellt: jetzt}
		spielerprofile[id] = profil
	}
	if name = strings.TrimSpace(name); name != "" {
		profil.Name = name
	}
	profil.Zuletzt = jetzt
	return profil
}

// sitzungMerken trägt eine Sitzung ein. Dasselbe Geheimnis bleibt dieselbe
// Sitzung – ein Lebenszeichen erzeugt keine zweite. Gehört das Geheimnis
// bereits zu einem anderen Spieler, wird nichts geändert: Ein fremdes Token
// darf keinen Platz übernehmen. Muss unter mu laufen.
func sitzungMerken(spielerID, geraet, geheimnis string) *Sitzung {
	spielerID = strings.TrimSpace(spielerID)
	geheimnis = strings.TrimSpace(geheimnis)
	if spielerID == "" || geheimnis == "" {
		return nil
	}
	jetzt := time.Now()
	sitzung := sitzungen[geheimnis]
	if sitzung != nil && sitzung.SpielerID != spielerID {
		return nil
	}
	if sitzung == nil {
		sitzung = &Sitzung{
			Kennung:   kurzeKennung(),
			Geheimnis: geheimnis,
			SpielerID: spielerID,
			Angelegt:  jetzt,
		}
		sitzungen[geheimnis] = sitzung
	}
	if geraetBrauchbar(geraet) {
		sitzung.Geraet = geraet
	}
	sitzung.Zuletzt = jetzt
	return sitzung
}

// sitzungPasst meldet, ob dieses Geheimnis eine gültige Sitzung genau dieses
// Spielers ist. Muss unter mu laufen.
func sitzungPasst(spielerID, geheimnis string) bool {
	if spielerID == "" || geheimnis == "" {
		return false
	}
	sitzung := sitzungen[geheimnis]
	if sitzung == nil || sitzung.SpielerID != spielerID {
		return false
	}
	return time.Since(sitzung.Zuletzt) <= sitzungFrist
}

// sitzungVonGeraet sucht die jüngste Sitzung dieses Spielers auf genau diesem
// Gerät. Sie holt ein verlorenes Token zurück – aber nur innerhalb der
// Gerätefrist und nie für eine geratene Kennung. Muss unter mu laufen.
func sitzungVonGeraet(spielerID, geraet string) *Sitzung {
	if spielerID == "" || !geraetBrauchbar(geraet) {
		return nil
	}
	var beste *Sitzung
	for _, sitzung := range sitzungen {
		if sitzung.SpielerID != spielerID || sitzung.Geraet != geraet {
			continue
		}
		if time.Since(sitzung.Zuletzt) > geraetFrist {
			continue
		}
		if beste == nil || sitzung.Zuletzt.After(beste.Zuletzt) {
			beste = sitzung
		}
	}
	return beste
}

// platzVonSpieler sucht den Platz dieses Spielers in einem Raum. Muss unter mu
// laufen.
func platzVonSpieler(spielerID string) (Player, bool) {
	for _, raum := range rooms {
		if platz, da := raum.Players[spielerID]; da {
			return platz, true
		}
	}
	return Player{}, false
}

// fuerAllePlaetze ruft aendern für jeden Platz dieses Spielers auf – in jedem
// Raum, in dem er Mitglied ist. Die Halle und eine private Runde schließen
// sich damit nicht mehr aus; ein Lebenszeichen gilt für alle. Muss unter mu
// laufen.
func fuerAllePlaetze(id string, aendern func(raum *Room, platz *Player)) {
	if id == "" || aendern == nil {
		return
	}
	for _, raum := range rooms {
		platz, da := raum.Players[id]
		if !da {
			continue
		}
		aendern(raum, &platz)
		// Nur zurückschreiben, wenn aendern den Platz nicht gerade entfernt hat.
		if _, noch := raum.Players[id]; noch {
			raum.Players[id] = platz
		}
	}
}

// identitaetPruefen beantwortet, ob dieser Aufruf zu diesem Spieler gehört.
// Dabei wird nichts angelegt: Eine Abweisung erzeugt keine Identität. Wieder-
// hergestellt ist true, wenn das Token fehlte und allein die Gerätekennung die
// bestehende Sitzung zurückgeholt hat. Muss unter mu laufen.
func identitaetPruefen(id, token, geraet string) (profil Spielerprofil, kennung, geheimnis string, wiederhergestellt, ok bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Spielerprofil{}, "", "", false, false
	}
	jetzt := time.Now()

	// 1. Gültige Sitzung dieses Spielers.
	if sitzungPasst(id, token) {
		sitzung := sitzungen[token]
		sitzung.Zuletzt = jetzt
		return profilKopie(id), sitzung.Kennung, sitzung.Geheimnis, false, true
	}

	// 2. Alter Platz im Raum: Sein Token gilt weiter, auch wenn die Sitzung
	//    noch nicht eingetragen ist (Stand vor dieser Änderung).
	if platz, da := platzVonSpieler(id); da && platz.Token != "" && platz.Token == token {
		spielerMerken(id, platz.Name)
		sitzung := sitzungMerken(id, platz.Geraet, platz.Token)
		if sitzung != nil {
			return profilKopie(id), sitzung.Kennung, sitzung.Geheimnis, false, true
		}
	}

	// 3. Verlorenes Token: dasselbe Gerät holt dieselbe Sitzung zurück.
	if sitzung := sitzungVonGeraet(id, geraet); sitzung != nil {
		sitzung.Zuletzt = jetzt
		spielerMerken(id, "")
		return profilKopie(id), sitzung.Kennung, sitzung.Geheimnis, true, true
	}
	if platz, da := platzVonSpieler(id); da && geraetPasst(platz, geraet, jetzt) {
		spielerMerken(id, platz.Name)
		sitzung := sitzungMerken(id, platz.Geraet, platz.Token)
		if sitzung != nil {
			return profilKopie(id), sitzung.Kennung, sitzung.Geheimnis, true, true
		}
	}

	return Spielerprofil{}, "", "", false, false
}

// profilKopie liefert das Profil als Wert. Muss unter mu laufen.
func profilKopie(id string) Spielerprofil {
	if profil := spielerprofile[id]; profil != nil {
		return *profil
	}
	return Spielerprofil{ID: id}
}

// sitzungenAufraeumen wirft abgelaufene Sitzungen weg und vergisst Profile, an
// denen weder Sitzung noch Platz hängt. Ohne das wüchse die Karte bei jedem
// Besuch weiter. Muss unter mu laufen.
func sitzungenAufraeumen(jetzt time.Time) {
	for geheimnis, sitzung := range sitzungen {
		if jetzt.Sub(sitzung.Zuletzt) > sitzungFrist {
			delete(sitzungen, geheimnis)
		}
	}
	for id, profil := range spielerprofile {
		if jetzt.Sub(profil.Zuletzt) <= profilFrist {
			continue
		}
		hatSitzung := false
		for _, sitzung := range sitzungen {
			if sitzung.SpielerID == id {
				hatSitzung = true
				break
			}
		}
		if hatSitzung {
			continue
		}
		if _, da := platzVonSpieler(id); da {
			continue
		}
		delete(spielerprofile, id)
	}
}

// ichHandler beantwortet GET /api/ich: Wer bin ich, wo bin ich, an welchen
// Tischen sitze ich, welche Geräte sind angemeldet? Der Ruf legt nichts an –
// ohne gültige Sitzung antwortet er mit 401.
func ichHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET erforderlich", http.StatusMethodNotAllowed)
		return
	}
	abfrage := r.URL.Query()
	mu.Lock()
	profil, kennung, geheimnis, wiederhergestellt, ok := identitaetPruefen(
		kennung(r), abfrage.Get("token"), abfrage.Get("geraet"))
	if !ok {
		mu.Unlock()
		http.Error(w, "Keine gültige Sitzung – bitte neu anmelden.", http.StatusUnauthorized)
		return
	}

	halle := ""
	if raum := rooms[mainRoomCode]; raum != nil {
		if _, da := raum.Players[profil.ID]; da {
			halle = raum.Code
		}
	}
	tische := []Tisch{}
	for code, raum := range rooms {
		if _, da := raum.Players[profil.ID]; !da {
			continue
		}
		if spiel := raum.aktivesSpiel(profil.ID); spiel != "" {
			tische = append(tische, Tisch{Raum: code, Spiel: spiel, Privat: raum.Privat})
		}
	}
	sort.Slice(tische, func(i, j int) bool { return tische[i].Raum < tische[j].Raum })

	eigene := []map[string]any{}
	for _, sitzung := range sitzungen {
		if sitzung.SpielerID != profil.ID {
			continue
		}
		eigene = append(eigene, map[string]any{
			"kennung":  sitzung.Kennung,
			"angelegt": sitzung.Angelegt,
			"zuletzt":  sitzung.Zuletzt,
			"dieses":   sitzung.Kennung == kennung,
		})
	}
	sort.Slice(eigene, func(i, j int) bool {
		return eigene[i]["angelegt"].(time.Time).Before(eigene[j]["angelegt"].(time.Time))
	})
	mu.Unlock()

	antwort := map[string]any{
		"spieler": map[string]any{
			"id":       profil.ID,
			"name":     profil.Name,
			"erstellt": profil.Erstellt,
		},
		"halle":     halle,
		"tische":    tische,
		"sitzungen": eigene,
	}
	// Ein wiedergewonnenes Token geht nur an dieses Gerät zurück – es hat sich
	// mit seiner Gerätekennung ausgewiesen.
	if wiederhergestellt {
		antwort["token"] = geheimnis
	}
	jsonResponse(w, antwort)
}

// sitzungHandler beantwortet POST /api/sitzung. Hier – und nur hier – entsteht
// eine neue Spieleridentität; jede andere Abweisung legt nichts an.
func sitzungHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST erforderlich", http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		Aktion    string `json:"aktion"`
		Name      string `json:"name"`
		Geraet    string `json:"geraet"`
		SpielerID string `json:"spielerId"`
		Kennung   string `json:"kennung"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	// Die geprüfte Kennung aus der Adresse gilt; der Rumpf füllt nur Lücken.
	if in.SpielerID == "" {
		in.SpielerID = kennung(r)
	}
	in.SpielerID = strings.TrimSpace(in.SpielerID)
	in.Geraet = strings.TrimSpace(in.Geraet)

	switch strings.ToLower(strings.TrimSpace(in.Aktion)) {
	case "anmelden":
		mu.Lock()
		// Bekanntes Gerät mit bekannter Kennung: dieselbe Sitzung zurückgeben,
		// statt eine zweite Identität anzulegen.
		if sitzung := sitzungVonGeraet(in.SpielerID, in.Geraet); sitzung != nil {
			sitzung.Zuletzt = time.Now()
			profil := spielerMerken(in.SpielerID, "")
			mu.Unlock()
			jsonResponse(w, map[string]any{
				"spielerId": profil.ID,
				"name":      profil.Name,
				"token":     sitzung.Geheimnis,
				"kennung":   sitzung.Kennung,
			})
			return
		}
		id := neueSpielerID()
		geheimnis := zufallsToken()
		if id == "" || geheimnis == "" {
			mu.Unlock()
			http.Error(w, "Anmeldung gerade nicht möglich.", http.StatusServiceUnavailable)
			return
		}
		name := limitName(strings.TrimSpace(in.Name))
		if name == "" {
			name = "Namenloser"
		}
		profil := spielerMerken(id, name)
		sitzung := sitzungMerken(id, in.Geraet, geheimnis)
		mu.Unlock()
		if profil == nil || sitzung == nil {
			http.Error(w, "Anmeldung gerade nicht möglich.", http.StatusServiceUnavailable)
			return
		}
		jsonResponse(w, map[string]any{
			"spielerId": profil.ID,
			"name":      profil.Name,
			"token":     sitzung.Geheimnis,
			"kennung":   sitzung.Kennung,
		})

	case "wiederherstellen":
		// Holt eine bestehende Sitzung zurück. Legt nichts an – ohne passende
		// Sitzung bleibt es bei 401.
		mu.Lock()
		sitzung := sitzungVonGeraet(in.SpielerID, in.Geraet)
		if sitzung == nil {
			mu.Unlock()
			http.Error(w, "Keine Sitzung für dieses Gerät.", http.StatusUnauthorized)
			return
		}
		sitzung.Zuletzt = time.Now()
		profil := profilKopie(sitzung.SpielerID)
		mu.Unlock()
		jsonResponse(w, map[string]any{
			"spielerId": profil.ID,
			"name":      profil.Name,
			"token":     sitzung.Geheimnis,
			"kennung":   sitzung.Kennung,
		})

	case "abmelden":
		// Beendet genau eine Sitzung. Profil und laufende Partien bleiben
		// bestehen – das Aufräumen der Plätze gehört zum Verlassen-Vorgang.
		ich := kennung(r)
		mu.Lock()
		sitzung := sitzungen[strings.TrimSpace(r.URL.Query().Get("token"))]
		if sitzung == nil || sitzung.SpielerID != ich || sitzung.Kennung != strings.TrimSpace(in.Kennung) {
			mu.Unlock()
			http.Error(w, "Diese Sitzung gehört nicht zu dir.", http.StatusForbidden)
			return
		}
		delete(sitzungen, sitzung.Geheimnis)
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Unbekannte Aktion", http.StatusBadRequest)
	}
}
