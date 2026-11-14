package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// port wird beim Start ermittelt, damit ein belegter Port die App nicht lahmlegt.
var port = "8765"

var (
	home string
	page string
)

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

// portFrei prüft, ob sich auf dieser Adresse lauschen lässt.
func portFrei(adresse string) bool {
	listener, err := net.Listen("tcp", adresse)
	if err != nil {
		return false
	}
	_ = listener.Close()
	return true
}

// freePort liefert den ersten freien Port ab 8765. Geprüft wird auch
// 127.0.0.1 – sonst kann ein anderer Dienst, der nur auf localhost lauscht,
// den Port für den Browser "gewinnen" und die Connector-Seite 404 liefern.
func freePort() string {
	for p := 8765; p <= 8785; p++ {
		candidate := strconv.Itoa(p)
		if !portFrei("127.0.0.1:"+candidate) || !portFrei(":"+candidate) {
			continue
		}
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
	// Ohne spiele-Ordner läuft der Chat, aber jede Spielseite wäre ein 404 –
	// das ist beim Kopieren der schnellste Weg in die Verzweiflung.
	if info, statErr := os.Stat(filepath.Join(home, "spiele")); statErr != nil || !info.IsDir() {
		log.Printf("WARNUNG: %s enthält keinen spiele-Ordner – die Spielseiten werden nicht gefunden.", home)
	}
	log.Fatal(http.ListenAndServe(":"+port, routes()))
}
