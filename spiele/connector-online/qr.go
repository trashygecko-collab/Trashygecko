package main

import (
	"github.com/skip2/go-qrcode"
	"net"
	"net/http"
	"strings"
)

// qrCode generates a QR code for the connector URL if the IP belongs to this host.
func qrCode(w http.ResponseWriter, r *http.Request) {
	ip := net.ParseIP(strings.TrimSpace(r.URL.Query().Get("ip"))).To4()
	if ip == nil {
		http.Error(w, "Ungültige IP-Adresse", http.StatusBadRequest)
		return
	}
	// Verify the IP belongs to a local interface.
	allowed := false
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if netIP, ok := addr.(*net.IPNet); ok && netIP.IP.To4() != nil && netIP.IP.Equal(ip) {
				allowed = true
				break
			}
		}
		if allowed {
			break
		}
	}
	if !allowed {
		http.Error(w, "IP-Adresse gehört nicht zu diesem PC", http.StatusBadRequest)
		return
	}
	png, err := qrcode.Encode("http://"+net.JoinHostPort(ip.String(), port)+"/connector", qrcode.Medium, 8)
	if err != nil {
		http.Error(w, "QR-Code konnte nicht erstellt werden", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(png)
}
