package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Der eigene Smiley-Bestand liegt als Dateien neben dem Connector:
//
//	spiele/connector/smilies/katze.gif
//	spiele/connector/smilies/tiere/hund.png
//
// Der Ordner wird bei jeder Abfrage frisch gelesen – neue Dateien sind sofort
// im Chat verfügbar, ohne dass hier irgendwo eine Liste gepflegt wird. Ein
// Unterordner ist die Kategorie; Dateien direkt im Ordner haben keine.
const (
	smileyOrdner    = "spiele/connector/smilies"
	smileyMaxBytes  = 1 << 20 // 1 MiB je Datei
	smileyMaxAnzahl = 500     // mehr wird nicht angeboten
)

// smileyArten sind die erlaubten Dateiendungen (klein geschrieben).
var smileyArten = map[string]string{
	".png": "png", ".jpg": "jpg", ".jpeg": "jpg", ".webp": "webp", ".gif": "gif",
}

// Smiley ist ein Eintrag aus dem Bestand.
type Smiley struct {
	Name      string `json:"name"`      // "katze.gif" oder "tiere/hund.png"
	Pfad      string `json:"pfad"`      // "/spiele/connector/smilies/tiere/hund.png"
	Kategorie string `json:"kategorie"` // "" = direkt im Ordner
	Art       string `json:"art"`       // png, jpg, webp, gif
	Groesse   int64  `json:"groesse"`   // Bytes
}

func smileyPfad() string {
	return filepath.Join(home, filepath.FromSlash(smileyOrdner))
}

// smileyListe liest den Bestand: erst die Dateien im Ordner, dann je einen
// Unterordner als Kategorie.
func smileyListe() []Smiley {
	liste := []Smiley{}
	wurzel := smileyPfad()
	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		return liste
	}
	for _, e := range eintraege {
		if len(liste) >= smileyMaxAnzahl {
			break
		}
		if e.IsDir() {
			unter, err := os.ReadDir(filepath.Join(wurzel, e.Name()))
			if err != nil {
				continue
			}
			for _, u := range unter {
				if len(liste) >= smileyMaxAnzahl {
					break
				}
				if s, ok := smileyEintrag(e.Name() + "/" + u.Name()); ok {
					liste = append(liste, s)
				}
			}
			continue
		}
		if s, ok := smileyEintrag(e.Name()); ok {
			liste = append(liste, s)
		}
	}
	sort.Slice(liste, func(i, j int) bool {
		return strings.ToLower(liste[i].Name) < strings.ToLower(liste[j].Name)
	})
	return liste
}

// smileyEintrag prüft eine einzelne Datei und macht daraus einen Eintrag.
func smileyEintrag(rel string) (Smiley, bool) {
	name, art, ok := smileyNamePruefen(rel)
	if !ok {
		return Smiley{}, false
	}
	info, err := os.Stat(filepath.Join(smileyPfad(), filepath.FromSlash(name)))
	if err != nil || !info.Mode().IsRegular() || info.Size() > smileyMaxBytes || info.Size() == 0 {
		return Smiley{}, false
	}
	return Smiley{
		Name:      name,
		Pfad:      "/" + smileyOrdner + "/" + name,
		Kategorie: kategorieVon(name),
		Art:       art,
		Groesse:   info.Size(),
	}, true
}

// smileyNamePruefen nimmt nur harmlose relative Namen mit erlaubter Endung an:
// kein "..", kein führender Schrägstrich, höchstens ein Unterordner.
func smileyNamePruefen(rel string) (string, string, bool) {
	name := strings.TrimSpace(strings.ReplaceAll(rel, "\\", "/"))
	if name == "" || len(name) > 120 || strings.HasPrefix(name, "/") || strings.Contains(name, "..") {
		return "", "", false
	}
	teile := strings.Split(name, "/")
	if len(teile) > 2 {
		return "", "", false
	}
	for _, t := range teile {
		if t == "" || t == "." {
			return "", "", false
		}
		for _, r := range t {
			if r < 0x20 || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' || r == '\\' {
				return "", "", false
			}
		}
	}
	art, ok := smileyArten[strings.ToLower(filepath.Ext(name))]
	if !ok {
		return "", "", false
	}
	return name, art, true
}

func kategorieVon(name string) string {
	if i := strings.LastIndex(name, "/"); i > 0 {
		return name[:i]
	}
	return ""
}

// smileyErlaubt prüft einen Namen aus einer Chatnachricht: Er muss auf eine
// vorhandene, erlaubte Grafik im Bestand zeigen. Alles andere fliegt raus.
func smileyErlaubt(name string) (string, bool) {
	s, ok := smileyEintrag(name)
	if !ok {
		return "", false
	}
	return s.Name, true
}

// smiliesHandler liefert den Bestand als JSON.
func smiliesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET erforderlich", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(smileyListe())
}
