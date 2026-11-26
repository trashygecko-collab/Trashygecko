package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// Diese Datei enthält alles, was die lokale und die Online-Fassung teilen:
// die Endpunkte, die Verbindungs-Zählung und die Kennungen für Spiele. Beide
// Fassungen rufen routes() auf – ein neues Spiel taucht dadurch automatisch in
// beiden auf.

func randomCode() string {
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	return "TG-" + strings.ToUpper(hex.EncodeToString(b))
}

// zufallsToken erzeugt das Geheimnis, das ein Spieler beim Beitritt bekommt.
// 16 Zufallsbytes als Hex sind nicht erratbar; geraten werden kann höchstens
// die Spielerkennung, und die allein reicht damit zu nichts mehr.
func zufallsToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// kontextSchluessel benennt die Werte, die withIdentity an die Handler gibt.
type kontextSchluessel int

const (
	kontextKennung kontextSchluessel = iota
	kontextRaum
)

// kennung liefert die von withIdentity geprüfte Spielerkennung. Die Handler
// lesen sie nur noch hier – der JSON-Rumpf kann eine Aktion damit nicht mehr
// auf einen fremden Platz umlenken.
func kennung(r *http.Request) string {
	wert, _ := r.Context().Value(kontextKennung).(string)
	return wert
}

// raumAusKontext liefert den geprüften Raumcode (Großbuchstaben).
func raumAusKontext(r *http.Request) string {
	wert, _ := r.Context().Value(kontextRaum).(string)
	return wert
}

// withIdentity prüft bei jedem API-Aufruf mit Spielerkennung, ob das
// mitgeschickte Token zu diesem Spieler gehört. Damit kann niemand mit einer
// fremden (etwa aus /api/players ausgelesenen) Kennung einen anderen
// umbenennen, aus einer Partie werfen oder für ihn ziehen. Das Token entsteht
// beim Beitritt (/api/presence) und wird nur dem Gerät ausgegeben, das den
// Platz angelegt hat.
//
// Nebenbei merkt sich die Hülle, wer wirklich etwas tut: Jeder POST außer dem
// automatischen Lebenszeichen zählt als Aktion (Chat, Spielzug, Einladung).
func withIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		geraet := strings.TrimSpace(r.URL.Query().Get("geraet"))

		// Kennung und Raum aus dem Rumpf füllen nur Lücken in der Adresse.
		// Verbindlich ist, was hier geprüft wird: Die Handler holen die
		// Kennung danach aus dem Kontext (kennung(r)) und nicht mehr aus dem
		// Rumpf – ein Rumpf kann also keine fremde Aktion mehr auslösen.
		rumpfID, rumpfRoom := "", ""
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			if err == nil {
				r.Body = io.NopCloser(bytes.NewReader(body))
				var probe struct {
					Room   string `json:"room"`
					ID     string `json:"id"`
					Token  string `json:"token"`
					Geraet string `json:"geraet"`
				}
				if json.Unmarshal(body, &probe) == nil {
					rumpfID = strings.TrimSpace(probe.ID)
					rumpfRoom = strings.ToUpper(strings.TrimSpace(probe.Room))
					if code == "" {
						code = rumpfRoom
					}
					if id == "" {
						id = rumpfID
					}
					if token == "" {
						token = strings.TrimSpace(probe.Token)
					}
					if geraet == "" {
						geraet = strings.TrimSpace(probe.Geraet)
					}
					// Lebenszeichen und der zentrale Statusruf melden selbst, ob
					// der Spieler etwas getan hat (Feld active). Würden sie hier
					// als Aktion zählen, bliebe die „eingeschlafen“-Meldung aus.
					if r.URL.Path != "/api/presence" && r.URL.Path != "/api/zustand" && id != "" {
						noteActivity(code, id)
					}
				}
			}
		}
		if id != "" && code != "" {
			mu.Lock()
			gespeichert := ""
			var platz Player
			bekannt := false
			if room := rooms[code]; room != nil {
				if p, ok := room.Players[id]; ok {
					gespeichert = p.Token
					platz = p
					bekannt = true
				}
			}
			// Eine gültige Sitzung dieses Spielers gilt auch dann, wenn der
			// Platz noch ein älteres Token trägt (zweiter Tab, ein anderes
			// Gerät desselben Spielers).
			sitzungGilt := sitzungPasst(id, token)
			mu.Unlock()
			// Nur ein bereits bekannter Platz mit Token kann abgewiesen werden;
			// ein neuer Spieler entsteht im Beitritt.
			if bekannt && !sitzungGilt && gespeichert != "" && gespeichert != token {
				// Ein verlorenes Token erneuert allein der Beitritt: Dort wird
				// die Gerätekennung geprüft und das neue Token gesetzt. Alle
				// anderen Endpunkte bleiben streng.
				darfErneuern := r.URL.Path == "/api/presence" &&
					geraetPasst(platz, geraet, time.Now()) && tokenBrauchbar(token)
				if !darfErneuern {
					http.Error(w, "Diese Spielerkennung gehört zu einem anderen Gerät.", http.StatusForbidden)
					return
				}
			}
		}
		// Geprüfte Identität an die Handler: Sie lesen Kennung und Raum nur
		// noch von hier, nie aus dem Rumpf.
		if id != "" {
			r = r.WithContext(context.WithValue(r.Context(), kontextKennung, id))
		}
		if code != "" {
			r = r.WithContext(context.WithValue(r.Context(), kontextRaum, code))
		}
		next.ServeHTTP(w, r)
	})
}

// routes hängt alle Endpunkte an einen Mux: Connector-Seite, Spiel-APIs und
// die Homepage als statische Dateien.
func routes() http.Handler {
	mux := http.NewServeMux()
	connectorPage := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(page))
	}
	mux.HandleFunc("/connector", connectorPage)
	mux.HandleFunc("/connector/", connectorPage)
	// Die Homepage verlinkt den Connector unter seinem Dateipfad – so heißt die
	// Seite auf GitHub Pages und in der Android-App. Im Docker-Abbild fehlt der
	// Ordner spiele/connector (Quelltext und Programme bleiben draußen), dort
	// antwortete diese Adresse bisher mit 404. Jetzt liefert der Server sie
	// selbst aus, damit ein Link auf allen drei Wegen ankommt.
	mux.HandleFunc("/spiele/connector/connector.html", connectorPage)
	mux.HandleFunc("/api/status", status)
	mux.HandleFunc("/api/spiele", spielRegister)
	mux.HandleFunc("/api/rooms", roomsHandler)
	// Private Einladungen: ohne sichtbaren Code, der Server gibt den Raum
	// allein an die beiden Beteiligten weiter.
	mux.HandleFunc("/api/einladung", einladungHandler)
	mux.HandleFunc("/api/presence", presence)
	// Ein Ruf, alle Zustände: Chat, Spieler, Einladungen, Angebot,
	// Stummschaltungen, laufende Partie und Lebenszeichen in einem. Die
	// Connector-Seite fragt damit im Takt statt mit sieben Timern.
	mux.HandleFunc("/api/zustand", zustandHandler)
	// Eine Tür für die Bereitschaft: Sie schaltet über die Aktion des Spiels
	// und antwortet mit dem frischen Stand (Etappe 5).
	mux.HandleFunc("/api/bereit", bereitHandler)
	// Identität und Sitzung: /api/ich sagt, wer man ist und wo man sitzt,
	// /api/sitzung meldet an, stellt eine Sitzung wieder her oder meldet ein
	// Gerät ab. Eine neue Identität entsteht nur bei "anmelden".
	mux.HandleFunc("/api/ich", ichHandler)
	mux.HandleFunc("/api/sitzung", sitzungHandler)
	mux.HandleFunc("/api/players", playersHandler)
	mux.HandleFunc("/api/qr", qrCode)
	mux.HandleFunc("/api/messages", messages)
	mux.HandleFunc("/api/smilies", smiliesHandler)
	mux.HandleFunc("/api/tone", toneHandler)
	mux.HandleFunc("/api/moderation", moderationHandler)
	mux.HandleFunc("/api/umbenennen", umbenennenHandler)
	mux.HandleFunc("/api/tictactoe", ticTacToeHandler)
	mux.HandleFunc("/api/viergewinnt", vierGewinntHandler)
	mux.HandleFunc("/api/dame", dameHandler)
	mux.HandleFunc("/api/paare", paareHandler)
	mux.HandleFunc("/api/menschaerger", menschaergerHandler)
	mux.HandleFunc("/api/maensync", maenSyncHandler)
	mux.HandleFunc("/api/kaesekaestche", kaesekaestcheHandler)
	mux.HandleFunc("/api/sim", simHandler)
	mux.HandleFunc("/api/tetris", tetrisHandler)
	mux.HandleFunc("/api/kalaha", kalahaHandler)
	mux.HandleFunc("/api/streichholz", streichholzHandler)
	mux.HandleFunc("/api/muehle", muehleHandler)
	mux.HandleFunc("/api/reversi", reversiHandler)
	mux.HandleFunc("/api/hold", holdHandler)
	mux.HandleFunc("/api/memory", memoryHandler)
	mux.HandleFunc("/api/mastermind", mastermindHandler)
	mux.HandleFunc("/api/skyjo", skyjoHandler)
	mux.HandleFunc("/api/kniffel", kniffelHandler)
	mux.HandleFunc("/api/zehntausend", zehnTausendHandler)
	mux.HandleFunc("/api/maexchen", maexchenHandler)
	mux.HandleFunc("/api/boston", bostonHandler)
	mux.HandleFunc("/api/chicago", chicagoHandler)
	mux.HandleFunc("/api/shipcrew", shipcrewHandler)
	mux.HandleFunc("/api/halma", halmaHandler)
	mux.HandleFunc("/api/wuerfel", wuerfelHandler)
	mux.HandleFunc("/api/schach", chessHandler)
	mux.HandleFunc("/api/go", goHandler)
	mux.HandleFunc("/api/backgammon", backgammonHandler)
	mux.HandleFunc("/api/maumau", maumauHandler)
	mux.HandleFunc("/api/uno", unoHandler)
	mux.HandleFunc("/api/poker", pokerHandler)
	mux.HandleFunc("/api/cabo", caboHandler)
	mux.HandleFunc("/api/siebzehn", siebzehnHandler)
	mux.HandleFunc("/api/baccara", baccaraHandler)
	mux.HandleFunc("/api/schiffe", schiffeHandler)
	mux.HandleFunc("/api/pong", pongHandler)
	mux.HandleFunc("/api/tron", tronHandler)
	mux.HandleFunc("/api/bomberman", bomberHandler)
	mux.HandleFunc("/api/bubble", bubbleHandler)
	mux.HandleFunc("/api/snake", snakeHandler)
	mux.HandleFunc("/api/schwimmen", schwimmenHandler)
	mux.HandleFunc("/api/arschloch", arschlochHandler)
	mux.HandleFunc("/api/sechsundsechzig", sechsundsechzigHandler)
	mux.HandleFunc("/api/galgen", galgenHandler)
	mux.HandleFunc("/api/quiz", quizHandler)
	// Fragenkatalog der Scherzshow: GET = lesen, POST = bearbeiten (Editor).
	mux.HandleFunc("/api/quizfragen", quizfragenHandler)
	mux.HandleFunc("/api/bingo", bingoHandler)
	mux.HandleFunc("/api/wortsuche", wortsucheHandler)
	mux.HandleFunc("/api/domino", dominoHandler)
	mux.HandleFunc("/api/schlangen", schlangenHandler)
	mux.HandleFunc("/api/pig", pigHandler)
	mux.HandleFunc("/api/gaensespiel", gaenseHandler)
	mux.HandleFunc("/api/risiko", risikoSpielHandler)
	mux.HandleFunc("/api/labyrinth", labyrinthHandler)
	mux.HandleFunc("/api/einsxeins", einsxeinsHandler)
	mux.HandleFunc("/api/angebot", angebotHandler)
	mux.HandleFunc("/api/malen", malenHandler)
	mux.HandleFunc("/api/leave", leaveGameHandler)
	mux.HandleFunc("/api/raum/verlassen", wegHandler)
	mux.HandleFunc("/api/imspiel", imSpielHandler)
	// Spielseiten und Skripte ändern sich zwischen den Runden. Seiten werden
	// gar nicht zwischengespeichert, Skripte müssen vor jedem Öffnen
	// nachfragen – sonst spielt ein Gerät mit einem alten Brett weiter (graue
	// Felder, roter Würfel), während der andere schon die neue Fassung hat.
	frisch := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, ".html") {
				w.Header().Set("Cache-Control", "no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
			} else if strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Set("Cache-Control", "no-cache")
			}
			h.ServeHTTP(w, r)
		})
	}
	mux.Handle("/", frisch(http.FileServer(http.Dir(home))))

	return withIdentity(mitIdempotenz(mux))
}
