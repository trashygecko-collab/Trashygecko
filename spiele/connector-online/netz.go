package main

import (
	"net/http"
	"os"
	"strings"
)

// publicURL ist die Adresse, unter der die Seite im Internet erreichbar ist.
// Sie kommt aus PUBLIC_URL (z. B. https://spiele.example.de) oder wird aus der
// Anfrage abgeleitet – hinter einem Proxy steht das Protokoll in
// X-Forwarded-Proto, sonst wäre alles fälschlich http.
func publicURL(r *http.Request) string {
	if wert := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_URL")), "/"); wert != "" {
		return wert
	}
	proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := r.Host
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwarded != "" {
		host = forwarded
	}
	return proto + "://" + host
}

// Online gibt es keinen Rechner, auf dem der Connector "neben dem Bildschirm"
// läuft: Boss und Admin ist, wer den Raum als Erster betritt. Danach kommen
// alle als Spieler1, Spieler2, … – umbenennen geht wie gehabt.
func init() {
	bossErkennen = func(r *http.Request, room *Room) bool {
		return room.OwnerID == ""
	}
}
