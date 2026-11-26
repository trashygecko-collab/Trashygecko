package main

// Doppelte Spielzüge durch eine erneute Zustellung verhindern.
//
// Die Seite schickt zu jedem POST eine eigene Kennung (Kopfzeile X-Aktion,
// gesetzt in spiele/token.js). Kommt dieselbe Kennung ein zweites Mal an – etwa
// weil die Antwort unterwegs verloren ging und die Seite den Ruf nach einer
// Wiederverbindung wiederholt –, wird nicht noch einmal gespielt: Der
// gespeicherte Antworttext geht unverändert zurück.
//
// Gespeichert werden nur erfolgreiche Antworten (2xx) bis 128 KB, höchstens
// 400 Stück und höchstens 90 Sekunden lang. Fehlgeschlagene Rufe werden nicht
// gespeichert – ein zweiter Versuch darf dort wirklich etwas tun.
//
// Zwischengespeichert wird nur, was ein Spielzug sein kann. Lebenszeichen,
// Chat, Raumverwaltung und Einladungen sind ausgenommen: Sie tragen ihre
// eigenen Schutzmechanismen (msgId, Token) und sollen immer frisch antworten.

import (
	"bytes"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	aktionFrist   = 90 * time.Second
	aktionMax     = 400
	aktionGroesse = 128 << 10
)

type aktionsAntwort struct {
	code  int
	rumpf []byte
	zeit  time.Time
}

var aktionen = struct {
	sync.Mutex
	m map[string]aktionsAntwort
}{m: map[string]aktionsAntwort{}}

// ohneAktionsschutz nennt die Endpunkte, die nicht zwischengespeichert werden.
func ohneAktionsschutz(pfad string) bool {
	switch pfad {
	case "/api/presence", "/api/zustand", "/api/messages", "/api/rooms",
		"/api/ich", "/api/sitzung", "/api/leave", "/api/raum/verlassen",
		"/api/angebot", "/api/bereit", "/api/einladung", "/api/moderation",
		"/api/umbenennen", "/api/tone", "/api/smilies", "/api/qr",
		"/api/players", "/api/imspiel":
		return true
	}
	return !strings.HasPrefix(pfad, "/api/")
}

// aktionSchluessel bildet den Schlüssel aus Raum, Spieler, Endpunkt, Aktion und
// der Marke der Seite. Ohne geprüfte Kennung oder ohne Marke gibt es keinen
// Schutz – dann läuft der Ruf wie bisher.
func aktionSchluessel(r *http.Request) string {
	if r.Method != http.MethodPost || ohneAktionsschutz(r.URL.Path) {
		return ""
	}
	marke := strings.TrimSpace(r.Header.Get("X-Aktion"))
	if marke == "" || len(marke) > 64 {
		return ""
	}
	id, code := kennung(r), raumAusKontext(r)
	if id == "" || code == "" {
		return ""
	}
	return code + "|" + id + "|" + r.URL.Path + "?" +
		strings.ToLower(r.URL.Query().Get("action")) + "|" + marke
}

func aktionHolen(schluessel string) (aktionsAntwort, bool) {
	jetzt := time.Now()
	aktionen.Lock()
	defer aktionen.Unlock()
	for k, a := range aktionen.m {
		if jetzt.Sub(a.zeit) > aktionFrist {
			delete(aktionen.m, k)
		}
	}
	a, da := aktionen.m[schluessel]
	return a, da
}

func aktionMerken(schluessel string, code int, rumpf []byte) {
	aktionen.Lock()
	defer aktionen.Unlock()
	if len(aktionen.m) >= aktionMax {
		// Ältesten Eintrag zuerst weg – der Speicher bleibt beschränkt.
		aeltester, zeit := "", time.Time{}
		for k, a := range aktionen.m {
			if zeit.IsZero() || a.zeit.Before(zeit) {
				aeltester, zeit = k, a.zeit
			}
		}
		if aeltester != "" {
			delete(aktionen.m, aeltester)
		}
	}
	aktionen.m[schluessel] = aktionsAntwort{code: code, rumpf: append([]byte(nil), rumpf...), zeit: time.Now()}
}

// antwortMitschrift merkt sich Status und Rumpf, damit eine Wiederholung genau
// dieselbe Antwort bekommt.
type antwortMitschrift struct {
	http.ResponseWriter
	code  int
	rumpf bytes.Buffer
}

func (a *antwortMitschrift) WriteHeader(code int) {
	a.code = code
	a.ResponseWriter.WriteHeader(code)
}

func (a *antwortMitschrift) Write(b []byte) (int, error) {
	if a.rumpf.Len() < aktionGroesse {
		a.rumpf.Write(b)
	}
	return a.ResponseWriter.Write(b)
}

// mitIdempotenz läuft innerhalb von withIdentity, damit Kennung und Raum schon
// geprüft im Kontext stehen.
func mitIdempotenz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		schluessel := aktionSchluessel(r)
		if schluessel == "" {
			next.ServeHTTP(w, r)
			return
		}
		if antwort, da := aktionHolen(schluessel); da {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("X-Aktion-Wiederholt", "1")
			w.WriteHeader(antwort.code)
			w.Write(antwort.rumpf)
			return
		}
		mit := &antwortMitschrift{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(mit, r)
		if mit.code >= 200 && mit.code < 300 && mit.rumpf.Len() <= aktionGroesse {
			aktionMerken(schluessel, mit.code, mit.rumpf.Bytes())
		}
	})
}
