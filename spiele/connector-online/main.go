package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Online-Fassung des Connectors: Sie liegt auf einem Server im Internet, hört
// auf den Port, den der Hoster vorgibt, öffnet keinen Browser und findet die
// Webseite dort, wo sie hochgeladen wurde. Alles Spielerische (Räume, Chat,
// Spiele) kommt unverändert aus ../connector und wird mit sync.py übernommen.

var (
	home string
	page string
	// port kommt aus PORT; er steht nur in Log und Status.
	port = "8080"
)

func envOder(name, standard string) string {
	if wert := strings.TrimSpace(os.Getenv(name)); wert != "" {
		return wert
	}
	return standard
}

// websiteRoot sucht den Ordner mit der Webseite: SITE_DIR, sonst das
// Verzeichnis der Programmdatei oder dessen Elternordner, wenn dort eine
// index.html liegt.
func websiteRoot(exeDir string) string {
	if wert := strings.TrimSpace(os.Getenv("SITE_DIR")); wert != "" {
		if abs, err := filepath.Abs(wert); err == nil {
			return abs
		}
	}
	for _, dir := range []string{exeDir, filepath.Join(exeDir, "..")} {
		if info, err := os.Stat(filepath.Join(dir, "index.html")); err == nil && !info.IsDir() {
			if abs, err := filepath.Abs(dir); err == nil {
				return abs
			}
		}
	}
	return exeDir
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exePath, _ = filepath.EvalSymlinks(exePath)
	exeDir := filepath.Dir(exePath)

	home = websiteRoot(exeDir)
	port = envOder("PORT", port)
	binden := strings.TrimSpace(os.Getenv("ADDR"))

	pagePath := filepath.Join(exeDir, "connector.html")
	pageData, err := os.ReadFile(pagePath)
	if err != nil {
		log.Fatalf("Connector-Oberfläche fehlt neben der Programmdatei (%s): %v", pagePath, err)
	}
	page = string(pageData)

	ensureMainRoom()
	startRoomJanitor()

	adresse := ":" + port
	if binden != "" {
		adresse = binden + ":" + port
	}

	log.Printf("TrashyGecko Connector (online) hört auf %s", adresse)
	log.Printf("Webseite: %s", home)
	// Ohne spiele-Ordner läuft der Chat, aber jede Spielseite wäre ein 404.
	if info, statErr := os.Stat(filepath.Join(home, "spiele")); statErr != nil || !info.IsDir() {
		log.Printf("WARNUNG: %s enthält keinen spiele-Ordner – die Spielseiten werden nicht gefunden.", home)
	}
	log.Fatal(http.ListenAndServe(adresse, routes()))
}
