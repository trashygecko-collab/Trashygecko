package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Das Malbrett ist eine gemeinsame Fläche pro Raum. Alle Striche landen auf dem
// Server, damit Handy und Rechner dasselbe Bild sehen; geleert werden darf es
// nur vom Admin (dem Besitzer des Raums).

const (
	malSize       = 1200.0 // festes Raster, damit alle dasselbe Bild sehen
	malMaxStrokes = 3000
	malMaxPoints  = 90000
)

// MalStrich ist eine gezeichnete Linie mit ein paar Punkten.
type MalStrich struct {
	Author string       `json:"author"`
	Color  string       `json:"color"`
	Width  float64      `json:"width"`
	Points [][2]float64 `json:"points"`
}

// Malbrett haelt alle Striche eines Raums.
type Malbrett struct {
	Epoch   int         `json:"epoch"`
	Strokes []MalStrich `json:"strokes"`
	points  int
}

// addStroke haengt einen Strich an und wirft die aeltesten weg, wenn es zu viel
// wird – so bleibt der Speicher klein, auch wenn lange gemalt wird.
func (m *Malbrett) addStroke(s MalStrich) {
	m.Strokes = append(m.Strokes, s)
	m.points += len(s.Points)

	for len(m.Strokes) > 1 && (m.points > malMaxPoints || len(m.Strokes) > malMaxStrokes) {
		m.points -= len(m.Strokes[0].Points)
		m.Strokes = m.Strokes[1:]
	}
}

func (m *Malbrett) leeren() {
	m.Strokes = nil
	m.points = 0
	m.Epoch++
}

// malFarbe prueft eine Hex-Farbe wie "#1e88e5".
func malFarbe(wert string) string {
	wert = strings.ToLower(strings.TrimSpace(wert))
	if len(wert) != 7 || wert[0] != '#' {
		return "#111111"
	}
	for _, zeichen := range wert[1:] {
		if !strings.ContainsRune("0123456789abcdef", zeichen) {
			return "#111111"
		}
	}
	return wert
}

func malBreite(wert float64) float64 {
	if wert < 1 {
		return 8
	}
	if wert > 60 {
		return 60
	}
	return wert
}

func malPunkte(punkte [][2]float64) [][2]float64 {
	sauber := make([][2]float64, 0, len(punkte))
	for _, punkt := range punkte {
		x, y := punkt[0], punkt[1]
		if x < 0 {
			x = 0
		}
		if x > malSize {
			x = malSize
		}
		if y < 0 {
			y = 0
		}
		if y > malSize {
			y = malSize
		}
		sauber = append(sauber, [2]float64{x, y})
	}
	return sauber
}

// malenHandler beantwortet "was ist auf der Flaeche?" und nimmt neue Striche an.
func malenHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	if code == "" {
		http.Error(w, "Spielraum-Code fehlt", http.StatusBadRequest)
		return
	}
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

	if r.Method == http.MethodGet {
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		epoche, _ := strconv.Atoi(r.URL.Query().Get("epoch"))
		seit, _ := strconv.Atoi(r.URL.Query().Get("since"))

		mu.Lock()
		raum := rooms[code]
		if raum == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		if raum.Mal == nil {
			raum.Mal = &Malbrett{}
		}
		brett := raum.Mal

		var liste []MalStrich
		if epoche == brett.Epoch {
			if seit < 0 {
				seit = 0
			}
			if seit > len(brett.Strokes) {
				seit = len(brett.Strokes)
			}
			liste = append(liste, brett.Strokes[seit:]...)
		} else {
			liste = append(liste, brett.Strokes...)
		}

		if liste == nil {
			liste = []MalStrich{}
		}

		antwort := map[string]any{
			"ok":      true,
			"epoch":   brett.Epoch,
			"rev":     len(brett.Strokes),
			"strokes": liste,
			"admin":   id != "" && raum.OwnerID == id,
		}
		mu.Unlock()
		jsonResponse(w, antwort)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input struct {
		ID      string      `json:"id"`
		Strokes []MalStrich `json:"strokes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		http.Error(w, "Spieler fehlt", http.StatusBadRequest)
		return
	}

	mu.Lock()
	raum := rooms[code]
	if raum == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	if raum.Mal == nil {
		raum.Mal = &Malbrett{}
	}
	brett := raum.Mal

	switch action {
	case "clear":
		// Zurücksetzen darf nur der Admin, also der Besitzer des Raums.
		if raum.OwnerID != id {
			mu.Unlock()
			http.Error(w, "Nur der Admin kann die Fläche leeren.", http.StatusForbidden)
			return
		}
		brett.leeren()
		raum.LastActive = time.Now()
		epoche := brett.Epoch
		mu.Unlock()
		jsonResponse(w, map[string]any{"ok": true, "epoch": epoche, "rev": 0})

	case "draw":
		neu := 0
		for _, strich := range input.Strokes {
			if len(strich.Points) == 0 || len(strich.Points) > 4000 {
				continue
			}
			strich.Author = id
			strich.Color = malFarbe(strich.Color)
			strich.Width = malBreite(strich.Width)
			strich.Points = malPunkte(strich.Points)
			brett.addStroke(strich)
			neu++
		}
		raum.LastActive = time.Now()
		epoche, stand := brett.Epoch, len(brett.Strokes)
		mu.Unlock()
		jsonResponse(w, map[string]any{"ok": true, "angenommen": neu, "epoch": epoche, "rev": stand})

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Aktion", http.StatusBadRequest)
	}
}
