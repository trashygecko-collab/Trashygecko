package main

import (
	"github.com/skip2/go-qrcode"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// eigeneIP prueft, ob die Adresse zu einer Netzwerkkarte dieses Rechners gehoert.
func eigeneIP(ip net.IP) bool {
	if ip == nil || ip.To4() == nil {
		return false
	}
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if netIP, ok := addr.(*net.IPNet); ok && netIP.IP.To4() != nil && netIP.IP.Equal(ip) {
				return true
			}
		}
	}
	return false
}

// qrAdresse bestimmt, was im Code steht. Grundsatz: nur eine Adresse, die zu
// diesem Server gehoert - eine geratene waere schlimmer als keine.
//
//	?url=<adresse>  genau die Adresse, die die Seite anzeigt (LAN oder oeffentlich)
//	?ip=<IPv4>      http://<ip>:<port>/connector, nur fuer Adressen dieses PCs
func qrAdresse(r *http.Request) (string, bool) {
	if roh := strings.TrimSpace(r.URL.Query().Get("url")); roh != "" {
		u, err := url.Parse(roh)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return "", false
		}
		host := strings.ToLower(u.Hostname())
		eigener := strings.ToLower(r.Host)
		if fwd := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); fwd != "" {
			eigener = strings.ToLower(fwd)
		}
		eigenerHost := eigener
		if h, _, splitErr := net.SplitHostPort(eigener); splitErr == nil {
			eigenerHost = h
		}
		if host != eigenerHost && !eigeneIP(net.ParseIP(host)) {
			return "", false
		}
		// Der Anschluss wird bewusst nicht geprueft: /api/status liefert die
		// Adresse, unter der die Seite erreichbar ist - online die oeffentliche
		// (https, hinter dem Proxy auf 443, waehrend der Connector innen auf
		// einem freien Port lauscht), lokal die LAN-Adresse mit dem Lausch-Port.
		// Diese Adresse wird unveraendert in den Code uebernommen.
		return u.String(), true
	}
	ip := net.ParseIP(strings.TrimSpace(r.URL.Query().Get("ip"))).To4()
	if ip == nil || !eigeneIP(ip) {
		return "", false
	}
	return "http://" + net.JoinHostPort(ip.String(), port) + "/connector", true
}

// qrCode liefert den QR-Code als PNG. Die Kantenlaenge ist ein Vielfaches der
// Modulzahl (8 Pixel je Modul), damit das Bild scharf bleibt: vorher stand
// "8" fuer die Gesamtgroesse, go-qrcode schrieb daraufhin 1 Pixel je Modul,
// und der Browser zog 37 Pixel unscharf auf 176 px hoch.
func qrCode(w http.ResponseWriter, r *http.Request) {
	adresse, ok := qrAdresse(r)
	if !ok {
		http.Error(w, "Keine gueltige Adresse fuer den QR-Code", http.StatusBadRequest)
		return
	}
	q, err := qrcode.New(adresse, qrcode.Medium)
	if err != nil {
		http.Error(w, "QR-Code konnte nicht erstellt werden", http.StatusInternalServerError)
		return
	}
	module := len(q.Bitmap()) // inklusive Rand
	skala := 8
	if module*skala > 2000 {
		skala = 4
	}
	png, err := q.PNG(module * skala)
	if err != nil {
		http.Error(w, "QR-Code konnte nicht erstellt werden", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(png)
}
