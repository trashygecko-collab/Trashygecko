package main

import (
	"net/http"
	"os"
	"strings"
	"time"
)

// startZeit ist der Beginn dieses Prozesses. Zusammen mit dem Commit zeigt sie,
// ob wirklich der neue Stand läuft: Ein Deploy startet den Dienst neu, ein
// Kaltstart nach dem Einschlafen (Free-Plan) ändert nur die Startzeit.
var startZeit = time.Now()

// bauCommit liefert den Commit-Kurz-Hash des laufenden Abbilds. Render setzt
// RENDER_GIT_COMMIT bei jedem Deploy; lokal oder im eigenen Container kann
// BAU_COMMIT gesetzt werden. Ohne beides bleibt es bei "unbekannt" – dann sagt
// wenigstens die Startzeit, ob der Dienst neu gestartet wurde.
func bauCommit() string {
	for _, name := range []string{"BAU_COMMIT", "RENDER_GIT_COMMIT"} {
		if wert := strings.TrimSpace(os.Getenv(name)); wert != "" {
			if len(wert) > 7 {
				wert = wert[:7]
			}
			return wert
		}
	}
	return "unbekannt"
}

// Online ist die LAN-Adresse des Servers uninteressant. Die Liste bleibt aber
// vorhanden, weil die Admin-Prüfung in lobby.go (isServerMachine) sie liest –
// als leere Liste ist dort niemand "der Server-PC".
func localIPv4s() []string {
	return nil
}

// status liefert der Connector-Oberfläche die Adresse, die Mitspieler
// eintippen sollen – und die Build-Kennung für die Veröffentlichungsprüfung.
func status(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]any{
		"port":     "",
		"ip":       "",
		"ips":      []string{},
		"url":      publicURL(r) + "/connector",
		"homepage": home,
		"online":   true,
		"commit":   bauCommit(),
		"started":  startZeit.UTC().Format(time.RFC3339),
	})
}
