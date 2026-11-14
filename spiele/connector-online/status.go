package main

import "net/http"

// Online ist die LAN-Adresse des Servers uninteressant. Die Liste bleibt aber
// vorhanden, weil die Admin-Prüfung in lobby.go (isServerMachine) sie liest –
// als leere Liste ist dort niemand "der Server-PC".
func localIPv4s() []string {
	return nil
}

// status liefert der Connector-Oberfläche die Adresse, die Mitspieler
// eintippen sollen.
func status(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]any{
		"port":     "",
		"ip":       "",
		"ips":      []string{},
		"url":      publicURL(r) + "/connector",
		"homepage": home,
		"online":   true,
	})
}
