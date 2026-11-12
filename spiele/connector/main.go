package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const port = "8765"

var (
	home string
	page string
)

func randomCode() string {
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	return "TG-" + strings.ToUpper(hex.EncodeToString(b))
}

func main() {
	home = `C:\Users\DarkP\Documents\Homepage`
	if len(os.Args) > 1 {
		home = os.Args[1]
	}
	abs, err := filepath.Abs(home)
	if err != nil {
		log.Fatal(err)
	}
	home = abs
	pageData, err := os.ReadFile(filepath.Join(home, "spiele", "connector", "connector.html"))
	if err != nil {
		log.Fatalf("Connector-Oberfläche konnte nicht geladen werden: %v", err)
	}
	page = string(pageData)

	mux := http.NewServeMux()
	mux.HandleFunc("/connector", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	})
	mux.HandleFunc("/api/status", status)
	mux.HandleFunc("/api/rooms", roomsHandler)
	mux.HandleFunc("/api/presence", presence)
	mux.HandleFunc("/api/players", playersHandler)
	mux.HandleFunc("/api/qr", qrCode)
	mux.HandleFunc("/api/messages", messages)
	mux.HandleFunc("/api/moderation", moderationHandler)
	mux.HandleFunc("/api/tictactoe", ticTacToeHandler)
	mux.HandleFunc("/api/viergewinnt", vierGewinntHandler)
	mux.HandleFunc("/api/dame", dameHandler)
	mux.Handle("/", http.FileServer(http.Dir(home)))

	addr := "http://localhost:" + port + "/connector"
	go func() {
		time.Sleep(700 * time.Millisecond)
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", addr).Start()
	}()

	log.Printf("TrashyGecko Connector startet: %s", addr)
	log.Printf("Homepage: %s", home)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
