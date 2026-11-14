package main

import (
	"net/http"

	"github.com/skip2/go-qrcode"
)

// qrCode zeigt einen QR-Code zur öffentlichen Adresse des Connectors. Der
// Parameter ip der lokalen Fassung wird hier nicht gebraucht.
func qrCode(w http.ResponseWriter, r *http.Request) {
	png, err := qrcode.Encode(publicURL(r)+"/connector", qrcode.Medium, 8)
	if err != nil {
		http.Error(w, "QR-Code konnte nicht erstellt werden", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(png)
}
