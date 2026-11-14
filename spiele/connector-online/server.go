package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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

// withActivity merkt sich, wer wirklich etwas tut: Jeder POST außer dem
// automatischen Lebenszeichen zählt als Aktion (Chat, Spielzug, Einladung).
// Die Kennung steht bei allen diesen Aufrufen im JSON-Feld "id".
func withActivity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/") &&
			r.URL.Path != "/api/presence" {
			body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			if err == nil {
				r.Body = io.NopCloser(bytes.NewReader(body))
				var probe struct {
					ID string `json:"id"`
				}
				if json.Unmarshal(body, &probe) == nil && probe.ID != "" {
					noteActivity(r.URL.Query().Get("room"), probe.ID)
				}
			}
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
	mux.HandleFunc("/api/status", status)
	mux.HandleFunc("/api/rooms", roomsHandler)
	mux.HandleFunc("/api/presence", presence)
	mux.HandleFunc("/api/players", playersHandler)
	mux.HandleFunc("/api/qr", qrCode)
	mux.HandleFunc("/api/messages", messages)
	mux.HandleFunc("/api/tone", toneHandler)
	mux.HandleFunc("/api/moderation", moderationHandler)
	mux.HandleFunc("/api/tictactoe", ticTacToeHandler)
	mux.HandleFunc("/api/viergewinnt", vierGewinntHandler)
	mux.HandleFunc("/api/dame", dameHandler)
	mux.HandleFunc("/api/paare", paareHandler)
	mux.HandleFunc("/api/menschaerger", menschaergerHandler)
	mux.HandleFunc("/api/kaesekaestche", kaesekaestcheHandler)
	mux.HandleFunc("/api/sim", simHandler)
	mux.HandleFunc("/api/kalaha", kalahaHandler)
	mux.HandleFunc("/api/streichholz", streichholzHandler)
	mux.HandleFunc("/api/muehle", muehleHandler)
	mux.HandleFunc("/api/reversi", reversiHandler)
	mux.HandleFunc("/api/hold", holdHandler)
	mux.HandleFunc("/api/skyjo", skyjoHandler)
	mux.HandleFunc("/api/kniffel", kniffelHandler)
	mux.HandleFunc("/api/schach", chessHandler)
	mux.HandleFunc("/api/go", goHandler)
	mux.HandleFunc("/api/backgammon", backgammonHandler)
	mux.HandleFunc("/api/maumau", maumauHandler)
	mux.HandleFunc("/api/uno", unoHandler)
	mux.HandleFunc("/api/schwimmen", schwimmenHandler)
	mux.HandleFunc("/api/arschloch", arschlochHandler)
	mux.HandleFunc("/api/einsxeins", einsxeinsHandler)
	mux.HandleFunc("/api/angebot", angebotHandler)
	mux.HandleFunc("/api/malen", malenHandler)
	mux.HandleFunc("/api/leave", leaveGameHandler)
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

	return withActivity(mux)
}
