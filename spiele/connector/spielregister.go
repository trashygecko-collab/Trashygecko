package main

// Rückfall für die Spielauswahl: Die Seite holt ihre Liste aus
// /spiele/multiplayer-register.js. Kommt dieses Skript im Browser nicht an
// (404, Netzfehler, blockiert), fragt der Connector /api/spiele – dieser
// Handler liest dieselbe Datei und gibt ihr JSON-Objekt unverändert zurück.
// Es bleibt also bei einer einzigen Liste; der Server ist nur eine zweite Tür
// dahinter. Fehlt die Datei wirklich, antwortet er mit 404 und die Seite zeigt
// den Grund an, statt eine leere Auswahl zu lassen.

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// registerDatei sucht /spiele/multiplayer-register.js neben der Seite.
func registerDatei() (string, error) {
	var kandidaten []string
	if home != "" {
		kandidaten = append(kandidaten, filepath.Join(home, "spiele", "multiplayer-register.js"))
	}
	if hier, err := os.Getwd(); err == nil {
		kandidaten = append(kandidaten,
			filepath.Join(hier, "..", "multiplayer-register.js"),
			filepath.Join(hier, "spiele", "multiplayer-register.js"),
			filepath.Join(hier, "multiplayer-register.js"))
	}
	for _, k := range kandidaten {
		if info, err := os.Stat(k); err == nil && !info.IsDir() {
			return k, nil
		}
	}
	return "", errors.New("Spieleliste nicht gefunden")
}

// spielRegister liefert { kategorien, spiele } als JSON.
func spielRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	pfad, err := registerDatei()
	if err != nil {
		http.Error(w, `{"fehler":"Spieleliste nicht gefunden"}`, http.StatusNotFound)
		return
	}
	roh, err := os.ReadFile(pfad)
	if err != nil {
		http.Error(w, `{"fehler":"Spieleliste nicht lesbar"}`, http.StatusInternalServerError)
		return
	}
	text := string(roh)
	i := strings.Index(text, "window.TGMultiplayer")
	if i < 0 {
		http.Error(w, `{"fehler":"Spieleliste hat ein unerwartetes Format"}`, http.StatusInternalServerError)
		return
	}
	rest := text[i:]
	j := strings.Index(rest, "=")
	k := strings.LastIndex(rest, "};")
	if j < 0 || k < j {
		http.Error(w, `{"fehler":"Spieleliste hat ein unerwartetes Format"}`, http.StatusInternalServerError)
		return
	}
	objekt := strings.TrimSpace(rest[j+1 : k+1])
	if !json.Valid([]byte(objekt)) {
		http.Error(w, `{"fehler":"Spieleliste ist kein gültiges JSON"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	io.WriteString(w, objekt)
}
