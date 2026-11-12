package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// port wird beim Start ermittelt, damit ein belegter Port die App nicht lahmlegt.
var port = "8765"

var (
	home string
	page string
)

func randomCode() string {
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	return "TG-" + strings.ToUpper(hex.EncodeToString(b))
}

// findHomepage ermittelt die Homepage-Wurzel ausgehend vom Ordner dieser Datei.
// Dadurch funktioniert der Connector unabhängig davon, wohin die Homepage
// kopiert wurde, und ohne dass ein Pfad fest eingebaut ist.
func findHomepage(connectorDir string) string {
	// 1. Der zuverlässigste Marker: <Homepage>/spiele/connector. Er findet
	//    auch dann die richtige Wurzel, wenn index.html fehlt oder umbenannt ist.
	for dir := connectorDir; ; dir = filepath.Dir(dir) {
		if info, err := os.Stat(filepath.Join(dir, "spiele", "connector")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	// 2. Sonst das nächste Elternverzeichnis mit index.html.
	for dir := connectorDir; ; dir = filepath.Dir(dir) {
		if info, err := os.Stat(filepath.Join(dir, "index.html")); err == nil && !info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	// Fallback: der Connector liegt immer unter <Homepage>/spiele/connector.
	// Nur akzeptieren, wenn dort auch der spiele-Ordner liegt, sonst würde
	// irgendein Großelternverzeichnis als Homepage herhalten.
	fallback := filepath.Clean(filepath.Join(connectorDir, "..", ".."))
	if fallback != connectorDir {
		if info, err := os.Stat(filepath.Join(fallback, "spiele")); err == nil && info.IsDir() {
			return fallback
		}
	}
	return connectorDir
}

// freePort liefert den ersten freien Port ab 8765.
func freePort() string {
	for p := 8765; p <= 8785; p++ {
		candidate := strconv.Itoa(p)
		listener, err := net.Listen("tcp", ":"+candidate)
		if err != nil {
			continue
		}
		_ = listener.Close()
		return candidate
	}
	return port
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	// Resolve any symlinks to get the actual executable location
	exePath, _ = filepath.EvalSymlinks(exePath)
	connectorDir := filepath.Dir(exePath)

	home = findHomepage(connectorDir)
	// Ein Argument wird nur übernommen, wenn es ein echtes Verzeichnis ist.
	// Sonst würde ein versehentlich übergebenes Argument (Drag & Drop,
	// Verknüpfung mit Parameter) die Homepage unbrauchbar machen.
	if len(os.Args) > 1 {
		if info, statErr := os.Stat(os.Args[1]); statErr == nil && info.IsDir() {
			home = os.Args[1]
		} else {
			log.Printf("Argument %q ist kein Verzeichnis und wird ignoriert.", os.Args[1])
		}
	}
	abs, err := filepath.Abs(home)
	if err != nil {
		log.Fatal(err)
	}
	home = abs

	pagePath := filepath.Join(connectorDir, "connector.html")
	pageData, err := os.ReadFile(pagePath)
	if err != nil {
		log.Fatalf("Connector-Oberfläche konnte nicht geladen werden (%s): %v", pagePath, err)
	}
	page = string(pageData)

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
	mux.HandleFunc("/api/moderation", moderationHandler)
	mux.HandleFunc("/api/tictactoe", ticTacToeHandler)
	mux.HandleFunc("/api/viergewinnt", vierGewinntHandler)
	mux.HandleFunc("/api/dame", dameHandler)
	mux.HandleFunc("/api/paare", paareHandler)
	mux.HandleFunc("/api/menschaerger", menschaergerHandler)
	mux.Handle("/", http.FileServer(http.Dir(home)))

	ensureMainRoom()
	startRoomJanitor()

	port = freePort()
	addr := "http://localhost:" + port + "/connector"
	go func() {
		time.Sleep(700 * time.Millisecond)
		if openErr := exec.Command("rundll32", "url.dll,FileProtocolHandler", addr).Start(); openErr != nil {
			log.Printf("Browser konnte nicht automatisch geöffnet werden: %v", openErr)
		}
	}()

	log.Printf("TrashyGecko Connector startet: %s", addr)
	log.Printf("Homepage: %s", home)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
