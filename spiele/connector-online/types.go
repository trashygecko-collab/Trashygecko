package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// maxNameLen ist die maximale Länge eines Spielernamens in Zeichen.
const maxNameLen = 24

// presenceTimeout bestimmt, wie lange ein Spieler ohne Lebenszeichen im Raum
// bleibt. Handys sperren den Bildschirm, wechseln kurz die App oder verlieren
// im WLAN ein paar Pakete – mit zu kurzer Frist flogen Spieler dabei aus dem
// Raum und ihre Partie wurde weggeworfen.
const presenceTimeout = 150 * time.Second

// emptyRoomGameGrace: Sind alle Geräte weg, bleiben Einladungen und Partien
// noch so lange liegen und gehen erst danach verloren.
const emptyRoomGameGrace = 5 * time.Minute

// gameTimeout: Ein Teilnehmer einer laufenden Partie ohne Lebenszeichen gilt
// nach dieser Frist als weg. Die Partie wird dann beendet und der verbliebene
// Spieler erfährt es sofort – er soll nicht auf jemanden warten, der weg ist.
//
// Die Frist muss über der Drosselperiode des Browsers liegen: Ein Tab, der
// lange im Hintergrund liegt, fragt nur noch etwa einmal pro Minute – bei 60 s
// wurde die Partie gemessen fälschlich beendet. 90 s lässt diesen Takt zu und
// meldet einen wirklich verschwundenen Spieler nach anderthalb Minuten.
// Die Mitgliedschaft im Raum bleibt länger bestehen (presenceTimeout), damit
// ein gesperrtes Handy niemanden aus dem Raum wirft.
const gameTimeout = 90 * time.Second

// spielSeiteFrist: Wie lange eine laufende Partie stehen bleiben darf, wenn
// keine Spielseite mehr nach ihr fragt. Wer mit dem Zurück-Knopf in den Raum
// geht, löst kein "Spiel verlassen" aus; ohne diese Frist bliebe die Partie
// für immer stehen und würde jede neue Runde blockieren.
const spielSeiteFrist = 90 * time.Second

// verstecktFrist: Ein Spieltab, der in den Hintergrund geht, meldet sich mit
// seite=versteckt. Solange dieses Signal nicht älter als die Frist ist, bleibt
// der Spieler im Raum und seine Partie stehen – ein eingefrorener Tab ist von
// einem geschlossenen nicht zu unterscheiden, und ein Handy im Hintergrund
// kommt meist zurück. Danach gelten wieder die normalen Fristen.
const verstecktFrist = 5 * time.Minute

// pauseNach: So lange nach dem Verstecken erfahren die Mitspieler einmal, dass
// die Partie steht. Sie sollen nicht ahnungslos auf ein stummes Gerät warten.
const pauseNach = 90 * time.Second

// Hinweis ist eine einmalige Nachricht an genau einen Spieler, zum Beispiel
// "Anna hat das Spiel verlassen – Tic Tac Toe beendet." Die Spielseite holt
// sie über /api/imspiel ab; dabei wird sie gelöscht (nur einmal).
// SpielerZeile ist ein Eintrag aus /api/players. Anwesend unterscheidet
// „ist gerade da“ von „steht nur noch im Raum“.
type SpielerZeile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Anwesend bool   `json:"anwesend"`
	// Spiel ist die Kennung der laufenden Partie dieses Spielers (leer, wenn er
	// an keinem Tisch sitzt). Damit kann der Connector auch Zuschauern den
	// Spielstand nennen, ohne die Spiele einzeln abzufragen.
	Spiel string `json:"spiel,omitempty"`
}

type Hinweis struct {
	Text string
	Zeit time.Time
}

// afkAfter: Wer so lange nichts tut (kein Chat, kein Klick, kein Zug), gilt
// als kurz weg – das automatische Lebenszeichen zählt dabei nicht mit.
const afkAfter = time.Minute

type Message struct {
	Name       string `json:"name"`
	Text       string `json:"text"`
	Time       string `json:"time"`
	SenderID   string `json:"senderId"`
	ToID       string `json:"toId,omitempty"`
	ToName     string `json:"toName,omitempty"`
	Whisper    bool   `json:"whisper"`
	GameInvite bool   `json:"gameInvite,omitempty"`
	// Tone färbt eine Nachricht ein: love (rosa), joke (grün), evil (rot).
	Tone   string `json:"tone,omitempty"`
	Game   string `json:"game,omitempty"`
	GameID string `json:"gameId,omitempty"`
	System bool   `json:"system,omitempty"`
	// Gone markiert "verlässt den Raum" – der Name ist dann nicht anklickbar.
	Gone bool `json:"gone,omitempty"`
	// Bild ist der Dateiname einer Grafik aus dem eigenen Smiley-Bestand
	// ("katze.gif" oder "tiere/katze.gif"). Leer = reine Textnachricht;
	// alte Nachrichten ohne dieses Feld bleiben unverändert lesbar.
	Bild string `json:"bild,omitempty"`
	// MsgID ist die Kennung, die die Seite beim Senden vergibt. Kommt dieselbe
	// Nachricht durch einen Wiederholungsversuch ein zweites Mal an, erkennt
	// der Raum sie daran und legt keine zweite Zeile an.
	MsgID string `json:"msgId,omitempty"`
}

type Player struct {
	Name string    `json:"name"`
	Last time.Time `json:"-"`
	// LastAction ist die letzte echte Aktion des Spielers, AfkWarned merkt
	// sich, ob für ihn schon eine „ist wohl eingeschlafen“-Meldung kam.
	LastAction time.Time `json:"-"`
	// Token ist das Geheimnis dieses Platzes. Es entsteht beim Beitritt und
	// muss bei jeder Änderung mitgeschickt werden – die Kennung allein (in
	// /api/players für alle sichtbar) reicht damit zu nichts mehr.
	Token string `json:"-"`
	// Geraet ist das zweite Geheimnis dieses Browsers. Geht das Token
	// verloren, holt es den Platz innerhalb von zehn Minuten zurück – aber
	// nur für genau dieses Gerät, nie für eine geratene Kennung.
	Geraet string `json:"-"`
	// ImSpiel ist der letzte Abruf von einer Spielseite. Die Connector-Seite
	// fragt mit seite=raum und setzt das nicht – daran erkennt der Raum, dass
	// eine Partie ohne Spielseite zurückgeblieben ist.
	ImSpiel time.Time `json:"-"`
	// Versteckt ist der letzte Ruf mit seite=versteckt: Der Tab liegt im
	// Hintergrund und darf einfrieren, ohne die Partie zu verlieren.
	Versteckt time.Time `json:"-"`
	// VerstecktSeit ist der Beginn dieser Hintergrundphase. Die Pause-Meldung
	// zählt ab hier – ein weiterer versteckter Ruf darf sie nicht verschieben.
	VerstecktSeit time.Time `json:"-"`
	// PauseGemeldet merkt sich, ob die Mitspieler schon von dieser Pause
	// gehört haben – die Meldung kommt je Hintergrundphase genau einmal.
	PauseGemeldet bool `json:"-"`
	AfkWarned     bool `json:"-"`
}

// verstecktAktiv meldet, ob dieser Spieler gerade mit einem Tab im Hintergrund
// sitzt und die Frist dafür noch läuft.
func (p Player) verstecktAktiv(now time.Time) bool {
	return !p.Versteckt.IsZero() && now.Sub(p.Versteckt) <= verstecktFrist
}

// anwesend sagt, ob dieser Spieler gerade wirklich da ist. Wer länger als
// gameTimeout nichts geschickt hat, zählt nicht mit – auch wenn sein Platz
// noch bis presenceTimeout freigehalten wird. Für Regeln wie „nur zu zweit“
// zählt, wer da ist, nicht wer nur noch im Raum steht.
//
// Ein eingefrorener Hintergrund-Tab zählt als da: Solange die Versteckt-Frist
// läuft (seite=versteckt, fünf Minuten), ist er von einem geschlossenen Tab
// nicht zu unterscheiden, und seine Partie darf nicht auseinanderfallen.
// Ohne diese Ausnahme wurde eine Zwei-Fenster-Partie auf einem PC nach 90 s
// als „Spieler weg“ aufgelöst.
func (p Player) anwesend(now time.Time) bool {
	if p.verstecktAktiv(now) {
		return true
	}
	return now.Sub(p.Last) <= gameTimeout
}

// anwesend zählt die wirklich anwesenden Spieler des Raums.
func (r *Room) anwesend() int {
	now := time.Now()
	n := 0
	for _, p := range r.Players {
		if p.anwesend(now) {
			n++
		}
	}
	return n
}

type TicTacToeGame struct {
	ID      string   `json:"id"`
	XID     string   `json:"xId"`
	XName   string   `json:"xName"`
	OID     string   `json:"oId"`
	OName   string   `json:"oName"`
	Board   []string `json:"board"`
	Current string   `json:"current"`
	Status  string   `json:"status"`
	Winner  string   `json:"winner,omitempty"`
	// Bot: Kennung („bot-1“), Stufe und die Zeit, zu der der Bot zieht.
	// Nur die Kennung geht an den Browser – Zeit und Stufe bleiben am Server.
	BotID    string    `json:"botId,omitempty"`
	BotStufe string    `json:"-"`
	BotFrist time.Time `json:"-"`
}

type Room struct {
	Code      string
	OwnerID   string
	OwnerName string
	// Privat kennzeichnet den Raum zweier eingeladener Spieler. Der
	// gemeinsame Hauptraum hat das Kennzeichen nie.
	Privat bool
	// EmptySince merkt sich, seit wann niemand mehr im Raum ist. Solange das
	// nicht lange her ist, bleiben laufende Partien liegen (Handy gesperrt,
	// Rechner zu, kurze Pause).
	EmptySince time.Time
	// botNamen hält die Namen der laufenden Partie (Bot-Nummer → Name), damit
	// ein Bot bei jedem Abruf gleich heißt. Vor jeder neuen Partie geleert.
	botNamen        map[string]string
	Name            string // e.g. "Spielername Raum"
	Created         time.Time
	LastActive      time.Time
	Players         map[string]Player
	Messages        []Message
	TicTacToe       *TicTacToeGame
	VierGewinnt     *VierGewinntGame
	Dame            *DameGame
	Paare           *PaareGame
	Menschaerger    *MenschaergerGame
	Kaese           *KaeseGame
	Sim             *SimGame
	Tetris          *TetrisGame
	Kalaha          *KalahaGame
	Streichholz     *StreichholzGame
	Muehle          *MuehleGame
	Reversi         *ReversiGame
	Hold            *HoldGame
	Memory          *MemoryGame
	Mastermind      *MastermindGame
	Skyjo           *SkyjoGame
	Kniffel         *KniffelGame
	ZehnTausend     *ZehnTausendGame // 10.000 (Würfelspiel online)
	Maexchen        *MaexchenGame    // Lügen Mäxchen (Bluff-Würfelspiel online)
	Boston          *BostonGame      // Auf nach Boston (Würfelrennen online)
	Chicago         *ChicagoGame     // Chicago (Würfelspiel mit Bierfilzen)
	ShipCrew        *ShipCrewGame    // Schiffe versenken (Ship Captain Crew, Würfel)
	Halma           *HalmaGame       // Halma (klassisch, Stern, Sprung)
	Schach          *ChessGame
	Go              *GoGame
	Backgammon      *BackgammonGame
	MauMau          *MauMauGame
	Uno             *UnoGame
	Poker           *PokerGame
	Cabo            *CaboGame
	Siebzehn        *SiebzehnGame
	Baccara         *BaccaraGame
	Schiffe         *SchiffeGame
	Pong            *PongGame   // Pong (Online-Duell)
	Tron            *TronGame   // Tron (Lichtrennen online)
	Bomberman       *BomberGame // Bomberman (Bomben-Duell online)
	Bubble          *BubbleGame // Bubble Shooter (Duell online)
	Snake           *SnakeGame  // Snake (Duell online)
	Schwimmen       *SchwimmenGame
	Arschloch       *ArschlochGame
	Sechsundsechzig *SechsundsechzigGame
	Rechen          *RechenGame
	Mal             *Malbrett              // gemeinsame Zeichenfläche
	Galgen          *GalgenGame            // Galgenmännchen
	Quiz            *QuizGame              // Scherzshow (humorvolles Quiz)
	Bingo           *BingoGame             // Bingo
	Wortsuche       *WortsucheGame         // Wortsuche
	Domino          *DominoGame            // Domino
	Schlangen       *SchlangenGame         // Schlangen und Leitern
	Pig             *PigGame               // Pig
	Gaense          *GaenseGame            // Gänsespiel
	Risiko          *RisikoSpiel           // Risiko (servergeprüfter Stand)
	Labyrinth       *SyncGame              // Das verrückte Labyrinth (Online)
	Wuerfel         *wuerfelPartie         // Würfelspiele der Sammlung (18 Regelmodule, Server würfelt)
	Angebot         *Angebot               // offene Runde für den ganzen Raum
	Mutes           map[string]time.Time   // playerID -> unmute time
	MsgTimes        map[string][]time.Time // for spam protection
	// Hinweise sind einmalige Nachrichten an einzelne Spieler (Abgang eines
	// Mitspielers). Sie werden beim Abholen gelöscht.
	Hinweise map[string]Hinweis
}

type RoomInfo struct {
	Code      string    `json:"code"`
	Players   int       `json:"players"`
	Created   string    `json:"created"`
	CreatedAt time.Time `json:"-"`
	Owner     string    `json:"owner,omitempty"`
	Status    string    `json:"status,omitempty"`
	Game      string    `json:"game,omitempty"`
	Name      string    `json:"name,omitempty"`
}

var (
	rooms = map[string]*Room{}
	mu    sync.Mutex
)

// limitName kürzt einen Namen auf maxNameLen Zeichen. Ein Schnitt nach Bytes
// würde Umlaute zerteilen und ungültiges UTF-8 erzeugen.
func limitName(s string) string {
	if utf8.RuneCountInString(s) <= maxNameLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxNameLen])
}

func (r *Room) cleanPlayers() {
	now := time.Now()
	for id, p := range r.Players {
		if now.Sub(p.Last) > presenceTimeout {
			// Ein Tab im Hintergrund darf einfrieren, ohne den Raum zu
			// verlieren – solange die Versteckt-Frist läuft, bleibt der
			// Spieler drin (Bildschirm gesperrt, App im Hintergrund).
			if p.verstecktAktiv(now) {
				continue
			}
			// Wer den Raum verlässt, darf keine Partie offen lassen.
			r.abgangBeenden(id, p.Name, "verbindung")
			delete(r.Players, id)
			// Mit Absender, damit der Name im Chat steht und angeklickt werden kann.
			r.addMessage(Message{
				Name:     p.Name,
				Text:     "verlässt den Raum.",
				Time:     now.Format("15:04:05"),
				SenderID: id,
				Gone:     true,
			})
		}
	}
	// Ein Angebot, dessen Ersteller weg ist, kann niemand mehr annehmen.
	if r.Angebot != nil {
		if _, da := r.Players[r.Angebot.VonID]; !da {
			r.Angebot = nil
		}
	}
	// Laufende Partien: Wer 60 s nichts mehr von sich hören lässt, ist weg.
	// Die Partie wird beendet, damit der Mitspieler nicht endlos wartet.
	for id, p := range r.Players {
		if now.Sub(p.Last) > gameTimeout {
			// Liegt der Tab im Hintergrund, ist Stille erwartbar: Die Partie
			// bleibt bis zum Ende der Versteckt-Frist stehen.
			if p.verstecktAktiv(now) {
				continue
			}
			r.abgangBeenden(id, p.Name, "verbindung")
		}
	}
	// Pause melden: Ein Spieltab im Hintergrund meldet sich eine Weile
	// gedrosselt weiter und friert dann ein. Nach pauseNach erfahren die
	// Mitspieler einmal, dass die Partie steht.
	for id, p := range r.Players {
		if p.PauseGemeldet || !p.verstecktAktiv(now) || now.Sub(p.VerstecktSeit) <= pauseNach {
			continue
		}
		if r.aktivesSpiel(id) == "" {
			continue
		}
		p.PauseGemeldet = true
		r.Players[id] = p
		r.addMessage(Message{
			Text:   p.Name + " pausiert – das Gerät ist im Hintergrund. Die Partie bleibt stehen, bis " + p.Name + " zurück ist.",
			Time:   now.Format("15:04:05"),
			System: true,
		})
	}
	// Alle wieder im Raum, aber die Partie steht noch (Zurück-Knopf): nach
	// der Frist auflösen, damit sie keine neue Runde blockiert.
	if r.hatLaufendesSpiel() && r.seitenAlleWeg(now) {
		for id, p := range r.Players {
			r.abgangBeenden(id, p.Name, "zurueck")
		}
	}
	if len(r.Players) == 0 {
		if r.EmptySince.IsZero() {
			r.EmptySince = now
		}
	} else {
		r.EmptySince = time.Time{}
	}
	// Sind alle Geräte kurz weg (gesperrte Handys, geschlossene Tabs), bleibt
	// die Partie noch eine Weile stehen, damit sie beim Zurückkommen weitergeht.
	if len(r.Players) == 0 && now.Sub(r.EmptySince) <= emptyRoomGameGrace {
		return
	}
	// Niemand mehr da: Nach der Frist bleibt keine Partie zurück, die eine
	// spätere Runde blockieren würde.
	if len(r.Players) == 0 {
		r.allePartienLoeschen()
	}
	// clean old timestamps for spam
	for id, ts := range r.MsgTimes {
		filtered := []time.Time{}
		for _, t := range ts {
			if now.Sub(t) < 10*time.Second {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) == 0 {
			delete(r.MsgTimes, id)
		} else {
			r.MsgTimes[id] = filtered
		}
	}
}

// uniqueName macht Spielernamen im Raum eindeutig: Wer mit einem schon
// belegten Namen kommt, wird „Anna1“, „Anna2“ usw. Sonst könnte sich jeder als
// Admin ausgeben und dessen Platz übernehmen.
func (r *Room) uniqueName(name, id string) string {
	taken := func(candidate string) bool {
		for otherID, p := range r.Players {
			if otherID != id && strings.EqualFold(p.Name, candidate) {
				return true
			}
		}
		return false
	}
	if !taken(name) {
		return name
	}
	base := limitName(name)
	for i := 1; ; i++ {
		candidate := limitName(base + strconv.Itoa(i))
		if !taken(candidate) {
			return candidate
		}
	}
}

// freierName findet den ersten freien Namen der Form "Spieler1", "Spieler2", …
func (r *Room) freierName(basis string) string {
	belegt := func(kandidat string) bool {
		for _, p := range r.Players {
			if strings.EqualFold(p.Name, kandidat) {
				return true
			}
		}
		return false
	}
	for i := 1; i < 100; i++ {
		if kandidat := basis + strconv.Itoa(i); !belegt(kandidat) {
			return kandidat
		}
	}
	return basis
}

// warnIdle meldet Spieler, die seit über einer Minute nichts getan haben. Der
// Admin bleibt außen vor: er sitzt am Server und schaut oft nur zu.
func (r *Room) warnIdle(now time.Time) {
	for id, p := range r.Players {
		if id == r.OwnerID || p.AfkWarned {
			continue
		}
		if now.Sub(p.LastAction) <= afkAfter {
			continue
		}
		p.AfkWarned = true
		r.Players[id] = p
		r.addMessage(Message{
			Text:   p.Name + " ist wohl gerade auf dem Klo oder eingeschlafen.",
			Time:   now.Format("15:04:05"),
			System: true,
		})
	}
}

// noteActivity hält fest, dass ein Spieler etwas getan hat (Chat, Spielzug,
// Klick auf der Oberfläche).
func noteActivity(code, id string) {
	if code == "" || id == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	room := rooms[strings.ToUpper(strings.TrimSpace(code))]
	if room == nil {
		return
	}
	player, ok := room.Players[id]
	if !ok {
		return
	}
	player.LastAction = time.Now()
	player.AfkWarned = false
	room.Players[id] = player
}

// gameSlotFree meldet, ob ein neues Spiel das aktuelle ersetzen darf. Eine
// abgeschlossene Partie blockiert nie, eine aufgegebene auch nicht.
func (r *Room) gameSlotFree(status string, ids ...string) bool {
	if status != "invited" && status != "active" {
		return true
	}
	return r.gameAbandoned(status, ids...)
}

// gameAbandoned meldet, ob die Partie keinen spielbaren Platz mehr hat: eine
// Einladung an einen Abwesenden oder eine laufende Partie, aus der ein Spieler
// verschwunden ist. Sie blockiert dann keine neue Runde (gameSlotFree).
func (r *Room) gameAbandoned(status string, ids ...string) bool {
	present := 0
	for _, id := range ids {
		if _, ok := r.Players[id]; ok {
			present++
		}
	}
	return present < len(ids)
}

// hatLaufendesSpiel meldet, ob irgendein Spieler des Raums in einer
// laufenden Partie sitzt.
func (r *Room) hatLaufendesSpiel() bool {
	for id := range r.Players {
		if r.aktivesSpiel(id) != "" {
			return true
		}
	}
	return false
}

// seitenAlleWeg meldet, ob alle, die je am Spielbrett saßen, seit der Frist
// keine Spielseite mehr geöffnet haben. Wer nie gespielt hat (Zuschauer),
// zählt nicht mit; gab es noch nie eine Spielseite, bleibt alles stehen.
func (r *Room) seitenAlleWeg(now time.Time) bool {
	gespielt := false
	for _, p := range r.Players {
		// Wer im Hintergrund sitzt, ist nicht "zurück im Raum": Die Partie
		// bleibt stehen, solange die Versteckt-Frist läuft.
		if p.verstecktAktiv(now) {
			return false
		}
		if p.ImSpiel.IsZero() {
			continue
		}
		gespielt = true
		if now.Sub(p.ImSpiel) < spielSeiteFrist {
			return false
		}
	}
	return gespielt
}

// imSpielVergessen löscht die Marke "war auf einer Spielseite". Ohne das gälte
// die nächste Partie sofort als zurückgelassen.
func (r *Room) imSpielVergessen() {
	for id, p := range r.Players {
		if !p.ImSpiel.IsZero() {
			p.ImSpiel = time.Time{}
			r.Players[id] = p
		}
	}
}

// hinweisFuer hinterlegt eine einmalige Nachricht für genau diesen Spieler.
func (r *Room) hinweisFuer(id, text string) {
	if id == "" || text == "" {
		return
	}
	if r.Hinweise == nil {
		r.Hinweise = map[string]Hinweis{}
	}
	r.Hinweise[id] = Hinweis{Text: text, Zeit: time.Now()}
}

// hinweisNehmen liefert den Hinweis genau einmal und löscht ihn dabei.
func (r *Room) hinweisNehmen(id string) string {
	if id == "" || r.Hinweise == nil {
		return ""
	}
	h, ok := r.Hinweise[id]
	if !ok {
		return ""
	}
	delete(r.Hinweise, id)
	if time.Since(h.Zeit) > 2*time.Minute {
		return ""
	}
	return h.Text
}

func (r *Room) isMuted(id string) (bool, time.Duration) {
	until, ok := r.Mutes[id]
	if !ok {
		return false, 0
	}
	remaining := time.Until(until)
	if remaining <= 0 {
		delete(r.Mutes, id)
		return false, 0
	}
	return true, remaining
}

func (r *Room) addMessage(msg Message) {
	r.Messages = append(r.Messages, msg)
	if len(r.Messages) > 100 {
		r.Messages = r.Messages[len(r.Messages)-100:]
	}
}

// nachrichtMitID sucht eine bereits angenommene Nachricht derselben Seite.
// Der Chat schickt bei einem Netzwerkfehler dieselbe msgId noch einmal – ohne
// diese Prüfung stünde die Zeile zweimal im Verlauf.
func (r *Room) nachrichtMitID(msgID, senderID string) (Message, bool) {
	if msgID == "" {
		return Message{}, false
	}
	for i := len(r.Messages) - 1; i >= 0; i-- {
		if r.Messages[i].MsgID == msgID && r.Messages[i].SenderID == senderID {
			return r.Messages[i], true
		}
	}
	return Message{}, false
}

func (r *Room) gameStatus() string {
	if r.TicTacToe != nil {
		switch r.TicTacToe.Status {
		case "active":
			return "Spiel läuft"
		case "invited":
			return "Einladung ausstehend"
		}
	}
	if r.VierGewinnt != nil {
		switch r.VierGewinnt.Status {
		case "active":
			return "Spiel läuft"
		case "invited":
			return "Einladung ausstehend"
		}
	}
	if r.Dame != nil {
		switch r.Dame.Status {
		case "active":
			return "Spiel läuft"
		case "invited":
			return "Einladung ausstehend"
		}
	}
	if r.Paare != nil {
		switch r.Paare.Status {
		case "active":
			return "Spiel läuft"
		case "invited":
			return "Einladung ausstehend"
		}
	}
	if r.Menschaerger != nil {
		switch r.Menschaerger.Status {
		case "active":
			return "Spiel läuft"
		case "invited":
			return "Einladung ausstehend"
		}
	}
	return "Verfügbar"
}

// cloneTicTacToe kopiert einen Spielstand. Ohne Kopie würde der JSON-Encoder
// das Brett lesen, während ein anderer Request es beschreibt (Datenrennen).
func cloneTicTacToe(g *TicTacToeGame) *TicTacToeGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Board = append([]string(nil), g.Board...)
	return &c
}

func cloneVierGewinnt(g *VierGewinntGame) *VierGewinntGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Grid = make([][]string, len(g.Grid))
	for i, row := range g.Grid {
		c.Grid[i] = append([]string(nil), row...)
	}
	if g.Winning != nil {
		c.Winning = make([][]int, len(g.Winning))
		for i, cell := range g.Winning {
			c.Winning[i] = append([]int(nil), cell...)
		}
	}
	return &c
}

func cloneMenschaerger(g *MenschaergerGame) *MenschaergerGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Players = make([]MaenPlayer, len(g.Players))
	for i, player := range g.Players {
		c.Players[i] = player
		c.Players[i].Pieces = append([]int(nil), player.Pieces...)
	}
	return &c
}

func clonePaare(g *PaareGame) *PaareGame {
	if g == nil {
		return nil
	}
	c := *g
	if g.RedDone != nil {
		result := *g.RedDone
		c.RedDone = &result
	}
	if g.BlueDone != nil {
		result := *g.BlueDone
		c.BlueDone = &result
	}
	if g.BotPlan != nil {
		plan := *g.BotPlan
		c.BotPlan = &plan
	}
	return &c
}

func cloneDame(g *DameGame) *DameGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Board = make([][]map[string]any, len(g.Board))
	for i, row := range g.Board {
		c.Board[i] = make([]map[string]any, len(row))
		for j, cell := range row {
			if cell == nil {
				continue
			}
			copied := make(map[string]any, len(cell))
			for k, v := range cell {
				copied[k] = v
			}
			c.Board[i][j] = copied
		}
	}
	return &c
}

// startRoomJanitor räumt Spielräume auf, in denen niemand mehr ist. Sonst
// wächst die Raumliste unbegrenzt und die Raumnamen zählen längst verlassene
// Räume mit ("Anna2 Raum", "Anna3 Raum", ...).
func startRoomJanitor() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for code, room := range rooms {
				room.cleanPlayers()
				room.warnIdle(time.Now())
				// Der eine Hauptraum bleibt immer bestehen – auch wenn der Host vom
				// Handy spielt und sein Rechner gerade keinen Tab offen hat.
				if code == mainRoomCode {
					continue
				}
				if len(room.Players) == 0 && time.Since(room.LastActive) > 2*time.Minute {
					delete(rooms, code)
				}
			}
			mu.Unlock()
		}
	}()
}

func jsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	// Spielstände dürfen nie aus dem Zwischenspeicher kommen: Bei wackligem
	// WLAN liefert ein Browser sonst einen alten Zug statt des neuen.
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}
