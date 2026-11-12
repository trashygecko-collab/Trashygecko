package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Go auf 9x9: Schwarz beginnt, zwei Pässe in Folge beenden die Partie,
// dann zählen Steine plus umschlossene Gebiete (Weiß bekommt 6,5 Ausgleich).
const goSize = 9

var goNeighborSteps = [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

type goPos struct {
	R int `json:"r"`
	C int `json:"c"`
}

type GoGame struct {
	ID         string      `json:"id"`
	RedID      string      `json:"redId"`
	RedName    string      `json:"redName"`
	BlueID     string      `json:"blueId"`
	BlueName   string      `json:"blueName"`
	Board      [][]*string `json:"board"`
	Turn       string      `json:"turn"` // B oder W
	CapturedB  int         `json:"capturedB"`
	CapturedW  int         `json:"capturedW"`
	Passes     int         `json:"passes"`
	Positions  []string    `json:"-"`
	LastMove   *goPos      `json:"lastMove"`
	Status     string      `json:"status"` // invited, active, finished
	Winner     string      `json:"winner,omitempty"`
	ScoreBlack int         `json:"scoreBlack"`
	ScoreWhite float64     `json:"scoreWhite"`
	WhiteBase  int         `json:"whiteBase"`
	Neutral    int         `json:"neutral"`
	Note       string      `json:"note,omitempty"`
}

func goEmptyBoard() [][]*string {
	board := make([][]*string, goSize)
	for r := range board {
		board[r] = make([]*string, goSize)
	}
	return board
}

func goCloneBoard(board [][]*string) [][]*string {
	next := make([][]*string, goSize)
	for r := 0; r < goSize; r++ {
		next[r] = make([]*string, goSize)
		copy(next[r], board[r])
	}
	return next
}

func goOther(color string) string {
	if color == "B" {
		return "W"
	}
	return "B"
}

func goWord(color string) string {
	if color == "B" {
		return "Schwarz"
	}
	return "Weiß"
}

func goInside(r, c int) bool {
	return r >= 0 && r < goSize && c >= 0 && c < goSize
}

func goKey(board [][]*string) string {
	var b strings.Builder
	for r := 0; r < goSize; r++ {
		for c := 0; c < goSize; c++ {
			if board[r][c] == nil {
				b.WriteByte('.')
			} else {
				b.WriteString(*board[r][c])
			}
		}
	}
	return b.String()
}

func goSet(board [][]*string, r, c int, value string) {
	if value == "" {
		board[r][c] = nil
		return
	}
	v := value
	board[r][c] = &v
}

// goGroup liefert die Steine einer Gruppe und die Anzahl ihrer Freiheiten.
func goGroup(board [][]*string, r, c int) ([][2]int, int) {
	if board[r][c] == nil {
		return nil, 0
	}
	color := *board[r][c]
	stones := [][2]int{}
	liberties := map[[2]int]bool{}
	visited := map[[2]int]bool{}
	stack := [][2]int{{r, c}}

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[cur] {
			continue
		}
		visited[cur] = true
		stones = append(stones, cur)

		for _, step := range goNeighborSteps {
			nr, nc := cur[0]+step[0], cur[1]+step[1]
			if !goInside(nr, nc) {
				continue
			}
			cell := board[nr][nc]
			if cell == nil {
				liberties[[2]int{nr, nc}] = true
			} else if *cell == color && !visited[[2]int{nr, nc}] {
				stack = append(stack, [2]int{nr, nc})
			}
		}
	}

	return stones, len(liberties)
}

// goTry spielt einen Stein auf ein Feld und gibt das neue Brett, die Zahl der
// gefangenen Steine und ob der Zug erlaubt ist zurück.
func goTry(board [][]*string, positions []string, r, c int, color string) ([][]*string, int, bool) {
	next := goCloneBoard(board)
	goSet(next, r, c, color)
	taken := 0

	for _, step := range goNeighborSteps {
		nr, nc := r+step[0], c+step[1]
		if !goInside(nr, nc) || next[nr][nc] == nil || *next[nr][nc] != goOther(color) {
			continue
		}
		stones, liberties := goGroup(next, nr, nc)
		if liberties == 0 {
			for _, stone := range stones {
				goSet(next, stone[0], stone[1], "")
				taken++
			}
		}
	}

	if _, liberties := goGroup(next, r, c); liberties == 0 {
		return nil, 0, false
	}

	key := goKey(next)
	for _, seen := range positions {
		if seen == key {
			return nil, 0, false
		}
	}

	return next, taken, true
}

func goCalculateScore(board [][]*string) (black, white, neutral int) {
	visited := make([][]bool, goSize)
	for r := range visited {
		visited[r] = make([]bool, goSize)
	}

	for r := 0; r < goSize; r++ {
		for c := 0; c < goSize; c++ {
			if board[r][c] == nil {
				continue
			}
			if *board[r][c] == "B" {
				black++
			} else {
				white++
			}
		}
	}

	for r := 0; r < goSize; r++ {
		for c := 0; c < goSize; c++ {
			if board[r][c] != nil || visited[r][c] {
				continue
			}

			stack := [][2]int{{r, c}}
			visited[r][c] = true
			size := 0
			borders := map[string]bool{}

			for len(stack) > 0 {
				cur := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				size++

				for _, step := range goNeighborSteps {
					nr, nc := cur[0]+step[0], cur[1]+step[1]
					if !goInside(nr, nc) {
						continue
					}
					if board[nr][nc] == nil {
						if !visited[nr][nc] {
							visited[nr][nc] = true
							stack = append(stack, [2]int{nr, nc})
						}
					} else {
						borders[*board[nr][nc]] = true
					}
				}
			}

			if len(borders) == 1 {
				if borders["B"] {
					black += size
				} else {
					white += size
				}
			} else {
				neutral += size
			}
		}
	}

	return black, white, neutral
}

func goNewGame(hostID, hostName, guestID, guestName, id string) *GoGame {
	board := goEmptyBoard()
	g := &GoGame{
		ID:        id,
		RedID:     hostID,
		RedName:   hostName,
		BlueID:    guestID,
		BlueName:  guestName,
		Board:     board,
		Turn:      "B",
		Positions: []string{goKey(board)},
		Status:    "invited",
	}
	return g
}

func cloneGo(g *GoGame) *GoGame {
	if g == nil {
		return nil
	}
	c := *g
	c.Board = goCloneBoard(g.Board)
	c.Positions = append([]string(nil), g.Positions...)
	if g.LastMove != nil {
		p := *g.LastMove
		c.LastMove = &p
	}
	return &c
}

func goColorOf(g *GoGame, id string) string {
	if g.RedID == id {
		return "B"
	}
	if g.BlueID == id {
		return "W"
	}
	return ""
}

func goHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("room")))
	if code == "" {
		http.Error(w, "Spielraum-Code fehlt", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		mu.Lock()
		room := rooms[code]
		if room == nil {
			mu.Unlock()
			http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
			return
		}
		if _, ok := room.Players[id]; !ok {
			mu.Unlock()
			http.Error(w, "Bitte zuerst dem Spielraum beitreten.", http.StatusForbidden)
			return
		}
		game := cloneGo(room.Go)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID   string `json:"id"`
		ToID string `json:"toId"`
		R    int    `json:"r"`
		C    int    `json:"c"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ungültige Daten", http.StatusBadRequest)
		return
	}
	input.ID = strings.TrimSpace(input.ID)
	input.ToID = strings.TrimSpace(input.ToID)
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

	mu.Lock()
	room := rooms[code]
	if room == nil {
		mu.Unlock()
		http.Error(w, "Spielraum nicht gefunden", http.StatusNotFound)
		return
	}
	room.cleanPlayers()
	player, ok := room.Players[input.ID]
	if !ok {
		mu.Unlock()
		http.Error(w, "Bitte erneut dem Spielraum beitreten.", http.StatusForbidden)
		return
	}

	switch action {
	case "invite":
		target, ok := room.Players[input.ToID]
		if !ok || input.ToID == input.ID {
			mu.Unlock()
			http.Error(w, "Dieser Spieler ist nicht verfügbar.", http.StatusBadRequest)
			return
		}
		if room.Go != nil && !room.gameSlotFree(room.Go.Status, room.Go.RedID, room.Go.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Go-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Go = goNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Go ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "go",
			GameID:     room.Go.ID,
		})

	case "accept":
		game := room.Go
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = "Schwarz beginnt!"

	case "decline":
		game := room.Go
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "play", "pass", "reset":
		game := room.Go
		if game == nil || (action != "reset" && game.Status != "active") {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := goColorOf(game, input.ID)
		if color == "" {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}

		switch action {
		case "play":
			if game.Turn != color {
				mu.Unlock()
				http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
				return
			}
			if !goInside(input.R, input.C) {
				mu.Unlock()
				http.Error(w, "Diese Kreuzung liegt nicht auf dem Brett.", http.StatusBadRequest)
				return
			}
			if game.Board[input.R][input.C] != nil {
				mu.Unlock()
				http.Error(w, "Auf dieser Kreuzung liegt schon ein Stein.", http.StatusConflict)
				return
			}

			next, taken, legal := goTry(game.Board, game.Positions, input.R, input.C, color)
			if !legal {
				mu.Unlock()
				http.Error(w, "Dieser Zug ist nicht erlaubt!", http.StatusConflict)
				return
			}

			game.Board = next
			game.Positions = append(game.Positions, goKey(next))
			if color == "B" {
				game.CapturedB += taken
			} else {
				game.CapturedW += taken
			}
			game.Passes = 0
			game.LastMove = &goPos{R: input.R, C: input.C}
			game.Turn = goOther(color)

			if taken > 0 {
				game.Note = goWord(color) + " fängt " + goCount(taken) + " gegnerische Steine!"
			} else {
				game.Note = goWord(game.Turn) + " ist am Zug."
			}

		case "pass":
			if game.Turn != color {
				mu.Unlock()
				http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
				return
			}
			game.Passes++
			game.LastMove = nil
			game.Turn = goOther(color)

			if game.Passes >= 2 {
				black, white, neutral := goCalculateScore(game.Board)
				whiteTotal := float64(white) + 6.5
				game.Status = "finished"
				game.ScoreBlack = black
				game.WhiteBase = white
				game.ScoreWhite = whiteTotal
				game.Neutral = neutral
				if float64(black) > whiteTotal {
					game.Winner = "B"
				} else {
					game.Winner = "W"
				}
				game.Note = "Partie beendet – " + goWord(game.Winner) + " gewinnt mit " +
					goScoreText(black, whiteTotal) + "."
				break
			}

			game.Note = goWord(color) + " passt. Ein weiterer Pass beendet die Partie."

		case "reset":
			if game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
				mu.Unlock()
				http.Error(w, "Eine neue Partie ist erst nach Spielende möglich.", http.StatusConflict)
				return
			}
			fresh := goNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
			fresh.Status = "active"
			fresh.Note = "Schwarz beginnt!"
			room.Go = fresh
		}

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Go-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := cloneGo(room.Go)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}

func goCount(n int) string {
	digits := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	if n >= 0 && n < len(digits) {
		return digits[n]
	}
	return "viele"
}

func goScoreText(black int, white float64) string {
	half := ""
	if white == float64(int(white)) {
		half = ".0"
	}
	return goCount(black) + " : " + goCount(int(white)) + half
}
