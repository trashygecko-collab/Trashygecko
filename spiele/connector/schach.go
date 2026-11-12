package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Schach: Weiß (Gastgeber) beginnt. Die Stellung hat dieselbe Form wie in
// der Spielseite: 8x8 Brett, Zugfarbe, Rochaderechte, En-passant-Feld.
type chessPiece struct {
	Color string `json:"color"`
	Type  string `json:"type"`
}

type chessPos struct {
	R int `json:"r"`
	C int `json:"c"`
}

type chessState struct {
	Board    [8][8]*chessPiece `json:"board"`
	Turn     string            `json:"turn"`
	Castling map[string]bool   `json:"castling"`
	EP       *chessPos         `json:"ep"`
	Fullmove int               `json:"fullmove"`
}

type chessMove struct {
	From      chessPos
	To        chessPos
	Promotion bool
	Castle    string
	EnPassant bool
}

type ChessGame struct {
	ID        string       `json:"id"`
	RedID     string       `json:"redId"`
	RedName   string       `json:"redName"`
	BlueID    string       `json:"blueId"`
	BlueName  string       `json:"blueName"`
	State     chessState   `json:"state"`
	Moves     []string     `json:"moves"`
	CapturedW []chessPiece `json:"capturedW"`
	CapturedB []chessPiece `json:"capturedB"`
	LastFrom  *chessPos    `json:"lastFrom"`
	LastTo    *chessPos    `json:"lastTo"`
	Status    string       `json:"status"` // invited, active, finished
	Winner    string       `json:"winner,omitempty"`
	Note      string       `json:"note,omitempty"`
}

var chessKnightSteps = [][2]int{
	{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
	{1, -2}, {1, 2}, {2, -1}, {2, 1},
}

func chessInside(r, c int) bool {
	return r >= 0 && r < 8 && c >= 0 && c < 8
}

func chessOpposite(color string) string {
	if color == "w" {
		return "b"
	}
	return "w"
}

func chessInitialState() chessState {
	s := chessState{
		Turn: "w",
		Castling: map[string]bool{
			"wK": true, "wQ": true, "bK": true, "bQ": true,
		},
	}
	back := []string{"R", "N", "B", "Q", "K", "B", "N", "R"}
	for c := 0; c < 8; c++ {
		s.Board[0][c] = &chessPiece{Color: "b", Type: back[c]}
		s.Board[1][c] = &chessPiece{Color: "b", Type: "P"}
		s.Board[6][c] = &chessPiece{Color: "w", Type: "P"}
		s.Board[7][c] = &chessPiece{Color: "w", Type: back[c]}
	}
	return s
}

func chessCloneState(s chessState) chessState {
	n := s
	n.Castling = map[string]bool{}
	for k, v := range s.Castling {
		n.Castling[k] = v
	}
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if s.Board[r][c] != nil {
				piece := *s.Board[r][c]
				n.Board[r][c] = &piece
			}
		}
	}
	if s.EP != nil {
		ep := *s.EP
		n.EP = &ep
	}
	return n
}

func chessFindKing(s chessState, color string) *chessPos {
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := s.Board[r][c]
			if p != nil && p.Type == "K" && p.Color == color {
				return &chessPos{R: r, C: c}
			}
		}
	}
	return nil
}

func chessIsAttacked(s chessState, r, c int, byColor string) bool {
	b := &s.Board

	pawnRow := r + 1
	if byColor == "b" {
		pawnRow = r - 1
	}
	for _, dc := range []int{-1, 1} {
		pc := c + dc
		if chessInside(pawnRow, pc) {
			p := b[pawnRow][pc]
			if p != nil && p.Color == byColor && p.Type == "P" {
				return true
			}
		}
	}

	for _, step := range chessKnightSteps {
		nr, nc := r+step[0], c+step[1]
		if chessInside(nr, nc) {
			p := b[nr][nc]
			if p != nil && p.Color == byColor && p.Type == "N" {
				return true
			}
		}
	}

	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			nr, nc := r+dr, c+dc
			if chessInside(nr, nc) {
				p := b[nr][nc]
				if p != nil && p.Color == byColor && p.Type == "K" {
					return true
				}
			}
		}
	}

	rookDirs := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	bishopDirs := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}

	for _, dir := range rookDirs {
		nr, nc := r+dir[0], c+dir[1]
		for chessInside(nr, nc) {
			p := b[nr][nc]
			if p != nil {
				if p.Color == byColor && (p.Type == "R" || p.Type == "Q") {
					return true
				}
				break
			}
			nr += dir[0]
			nc += dir[1]
		}
	}

	for _, dir := range bishopDirs {
		nr, nc := r+dir[0], c+dir[1]
		for chessInside(nr, nc) {
			p := b[nr][nc]
			if p != nil {
				if p.Color == byColor && (p.Type == "B" || p.Type == "Q") {
					return true
				}
				break
			}
			nr += dir[0]
			nc += dir[1]
		}
	}

	return false
}

func chessInCheck(s chessState, color string) bool {
	king := chessFindKing(s, color)
	if king == nil {
		return false
	}
	return chessIsAttacked(s, king.R, king.C, chessOpposite(color))
}

func chessPseudoMoves(s chessState, r, c int) []chessMove {
	p := s.Board[r][c]
	if p == nil {
		return nil
	}
	moves := []chessMove{}
	color := p.Color
	enemy := chessOpposite(color)

	add := func(nr, nc int, move chessMove) {
		if !chessInside(nr, nc) {
			return
		}
		target := s.Board[nr][nc]
		if target != nil && (target.Color == color || target.Type == "K") {
			return
		}
		move.To = chessPos{R: nr, C: nc}
		moves = append(moves, move)
	}

	switch p.Type {
	case "P":
		dir := -1
		start := 6
		promotionRow := 0
		if color == "b" {
			dir = 1
			start = 1
			promotionRow = 7
		}

		if chessInside(r+dir, c) && s.Board[r+dir][c] == nil {
			add(r+dir, c, chessMove{From: chessPos{R: r, C: c}, Promotion: r+dir == promotionRow})
			if r == start && chessInside(r+2*dir, c) && s.Board[r+2*dir][c] == nil {
				add(r+2*dir, c, chessMove{From: chessPos{R: r, C: c}})
			}
		}

		for _, dc := range []int{-1, 1} {
			nr, nc := r+dir, c+dc
			if !chessInside(nr, nc) {
				continue
			}
			target := s.Board[nr][nc]
			if target != nil && target.Color == enemy && target.Type != "K" {
				add(nr, nc, chessMove{From: chessPos{R: r, C: c}, Promotion: nr == promotionRow})
			} else if s.EP != nil && s.EP.R == nr && s.EP.C == nc {
				moves = append(moves, chessMove{
					From:      chessPos{R: r, C: c},
					To:        chessPos{R: nr, C: nc},
					EnPassant: true,
				})
			}
		}

	case "N":
		for _, step := range chessKnightSteps {
			add(r+step[0], c+step[1], chessMove{From: chessPos{R: r, C: c}})
		}

	case "B", "R", "Q":
		dirs := [][2]int{}
		if p.Type != "R" {
			dirs = append(dirs, [2]int{-1, -1}, [2]int{-1, 1}, [2]int{1, -1}, [2]int{1, 1})
		}
		if p.Type != "B" {
			dirs = append(dirs, [2]int{-1, 0}, [2]int{1, 0}, [2]int{0, -1}, [2]int{0, 1})
		}
		for _, dir := range dirs {
			nr, nc := r+dir[0], c+dir[1]
			for chessInside(nr, nc) {
				target := s.Board[nr][nc]
				if target == nil {
					moves = append(moves, chessMove{From: chessPos{R: r, C: c}, To: chessPos{R: nr, C: nc}})
				} else {
					if target.Color == enemy && target.Type != "K" {
						moves = append(moves, chessMove{From: chessPos{R: r, C: c}, To: chessPos{R: nr, C: nc}})
					}
					break
				}
				nr += dir[0]
				nc += dir[1]
			}
		}

	case "K":
		for dr := -1; dr <= 1; dr++ {
			for dc := -1; dc <= 1; dc++ {
				if dr == 0 && dc == 0 {
					continue
				}
				add(r+dr, c+dc, chessMove{From: chessPos{R: r, C: c}})
			}
		}

		home := 0
		if color == "w" {
			home = 7
		}
		if r == home && c == 4 && !chessInCheck(s, color) {
			if s.Castling[color+"K"] &&
				s.Board[home][5] == nil && s.Board[home][6] == nil &&
				s.Board[home][7] != nil && s.Board[home][7].Type == "R" && s.Board[home][7].Color == color &&
				!chessIsAttacked(s, home, 5, enemy) && !chessIsAttacked(s, home, 6, enemy) {
				moves = append(moves, chessMove{
					From:   chessPos{R: r, C: c},
					To:     chessPos{R: home, C: 6},
					Castle: "K",
				})
			}
			if s.Castling[color+"Q"] &&
				s.Board[home][1] == nil && s.Board[home][2] == nil && s.Board[home][3] == nil &&
				s.Board[home][0] != nil && s.Board[home][0].Type == "R" && s.Board[home][0].Color == color &&
				!chessIsAttacked(s, home, 3, enemy) && !chessIsAttacked(s, home, 2, enemy) {
				moves = append(moves, chessMove{
					From:   chessPos{R: r, C: c},
					To:     chessPos{R: home, C: 2},
					Castle: "Q",
				})
			}
		}
	}

	return moves
}

func chessApplyMove(s chessState, move chessMove, promotionChoice string) chessState {
	n := chessCloneState(s)
	piece := n.Board[move.From.R][move.From.C]
	target := n.Board[move.To.R][move.To.C]

	n.Board[move.From.R][move.From.C] = nil

	if move.EnPassant {
		n.Board[move.From.R][move.To.C] = nil
	}

	pieceType := piece.Type
	if move.Promotion {
		pieceType = promotionChoice
	}
	n.Board[move.To.R][move.To.C] = &chessPiece{Color: piece.Color, Type: pieceType}

	if move.Castle != "" {
		row := move.From.R
		if move.Castle == "K" {
			n.Board[row][5] = n.Board[row][7]
			n.Board[row][7] = nil
		} else {
			n.Board[row][3] = n.Board[row][0]
			n.Board[row][0] = nil
		}
	}

	if piece.Type == "K" {
		n.Castling[piece.Color+"K"] = false
		n.Castling[piece.Color+"Q"] = false
	}

	if piece.Type == "R" {
		if move.From.R == 7 && move.From.C == 0 {
			n.Castling["wQ"] = false
		}
		if move.From.R == 7 && move.From.C == 7 {
			n.Castling["wK"] = false
		}
		if move.From.R == 0 && move.From.C == 0 {
			n.Castling["bQ"] = false
		}
		if move.From.R == 0 && move.From.C == 7 {
			n.Castling["bK"] = false
		}
	}

	if target != nil && target.Type == "R" {
		if move.To.R == 7 && move.To.C == 0 {
			n.Castling["wQ"] = false
		}
		if move.To.R == 7 && move.To.C == 7 {
			n.Castling["wK"] = false
		}
		if move.To.R == 0 && move.To.C == 0 {
			n.Castling["bQ"] = false
		}
		if move.To.R == 0 && move.To.C == 7 {
			n.Castling["bK"] = false
		}
	}

	n.EP = nil
	if piece.Type == "P" && abs(move.To.R-move.From.R) == 2 {
		n.EP = &chessPos{R: (move.From.R + move.To.R) / 2, C: move.From.C}
	}

	if piece.Color == "b" {
		n.Fullmove++
	}
	n.Turn = chessOpposite(piece.Color)
	return n
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func chessLegalMovesAt(s chessState, r, c int) []chessMove {
	p := s.Board[r][c]
	if p == nil {
		return nil
	}
	legal := []chessMove{}
	for _, move := range chessPseudoMoves(s, r, c) {
		next := chessApplyMove(s, move, "Q")
		if !chessInCheck(next, p.Color) {
			legal = append(legal, move)
		}
	}
	return legal
}

func chessHasAnyMove(s chessState, color string) bool {
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			p := s.Board[r][c]
			if p != nil && p.Color == color && len(chessLegalMovesAt(s, r, c)) > 0 {
				return true
			}
		}
	}
	return false
}

func chessSquareName(pos chessPos) string {
	return string(rune('a'+pos.C)) + string(rune('0'+(8-pos.R)))
}

func chessNotation(move chessMove, piece *chessPiece, captured bool, promotionChoice string) string {
	if move.Castle == "K" {
		return "O-O"
	}
	if move.Castle == "Q" {
		return "O-O-O"
	}

	result := ""
	if piece.Type == "P" {
		if captured {
			result = string(rune('a'+move.From.C)) + "x"
		}
	} else {
		result = piece.Type
	}
	result += chessSquareName(move.To)
	if move.Promotion {
		result += "=" + promotionChoice
	}
	if move.EnPassant {
		result += " e.p."
	}
	return result
}

func chessNewGame(hostID, hostName, guestID, guestName, id string) *ChessGame {
	return &ChessGame{
		ID:        id,
		RedID:     hostID,
		RedName:   hostName,
		BlueID:    guestID,
		BlueName:  guestName,
		State:     chessInitialState(),
		Moves:     []string{},
		CapturedW: []chessPiece{},
		CapturedB: []chessPiece{},
		Status:    "invited",
	}
}

func chessClone(g *ChessGame) *ChessGame {
	if g == nil {
		return nil
	}
	c := *g
	c.State = chessCloneState(g.State)
	c.Moves = append([]string(nil), g.Moves...)
	c.CapturedW = append([]chessPiece(nil), g.CapturedW...)
	c.CapturedB = append([]chessPiece(nil), g.CapturedB...)
	if g.LastFrom != nil {
		p := *g.LastFrom
		c.LastFrom = &p
	}
	if g.LastTo != nil {
		p := *g.LastTo
		c.LastTo = &p
	}
	return &c
}

func chessColorOf(g *ChessGame, id string) string {
	if g.RedID == id {
		return "w"
	}
	if g.BlueID == id {
		return "b"
	}
	return ""
}

func chessHandler(w http.ResponseWriter, r *http.Request) {
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
		game := chessClone(room.Schach)
		mu.Unlock()
		jsonResponse(w, map[string]any{"game": game})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID        string   `json:"id"`
		ToID      string   `json:"toId"`
		From      chessPos `json:"from"`
		To        chessPos `json:"to"`
		Promotion string   `json:"promotion"`
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
		if room.Schach != nil && !room.gameSlotFree(room.Schach.Status, room.Schach.RedID, room.Schach.BlueID) {
			mu.Unlock()
			http.Error(w, "In diesem Spielraum läuft bereits eine Schach-Partie oder Einladung.", http.StatusConflict)
			return
		}
		room.Schach = chessNewGame(input.ID, player.Name, input.ToID, target.Name, randomCode())
		room.addMessage(Message{
			Name:       player.Name,
			Text:       "lädt dich zu Schach ein.",
			Time:       time.Now().Format("15:04:05"),
			SenderID:   input.ID,
			ToID:       input.ToID,
			ToName:     target.Name,
			Whisper:    true,
			GameInvite: true,
			Game:       "schach",
			GameID:     room.Schach.ID,
		})

	case "accept":
		game := room.Schach
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "active"
		game.Note = "Weiß beginnt – wähle eine Figur."

	case "decline":
		game := room.Schach
		if game == nil || game.Status != "invited" || game.BlueID != input.ID {
			mu.Unlock()
			http.Error(w, "Diese Einladung ist nicht mehr verfügbar.", http.StatusConflict)
			return
		}
		game.Status = "declined"

	case "move":
		game := room.Schach
		if game == nil || game.Status != "active" {
			mu.Unlock()
			http.Error(w, "Es läuft gerade keine aktive Partie.", http.StatusConflict)
			return
		}
		color := chessColorOf(game, input.ID)
		if color == "" {
			mu.Unlock()
			http.Error(w, "Du gehörst nicht zu dieser Partie.", http.StatusForbidden)
			return
		}
		if game.State.Turn != color {
			mu.Unlock()
			http.Error(w, "Der andere Spieler ist an der Reihe.", http.StatusConflict)
			return
		}
		if !chessInside(input.From.R, input.From.C) || !chessInside(input.To.R, input.To.C) {
			mu.Unlock()
			http.Error(w, "Dieses Feld liegt nicht auf dem Brett.", http.StatusBadRequest)
			return
		}
		piece := game.State.Board[input.From.R][input.From.C]
		if piece == nil || piece.Color != color {
			mu.Unlock()
			http.Error(w, "Auf diesem Feld steht keine eigene Figur.", http.StatusConflict)
			return
		}

		var chosen *chessMove
		for _, move := range chessLegalMovesAt(game.State, input.From.R, input.From.C) {
			if move.To == input.To {
				m := move
				chosen = &m
				break
			}
		}
		if chosen == nil {
			mu.Unlock()
			http.Error(w, "Dieser Zug ist nicht erlaubt.", http.StatusConflict)
			return
		}

		promotionChoice := strings.ToUpper(strings.TrimSpace(input.Promotion))
		if promotionChoice != "Q" && promotionChoice != "R" && promotionChoice != "B" && promotionChoice != "N" {
			promotionChoice = "Q"
		}

		capturedPiece := game.State.Board[input.To.R][input.To.C]
		if chosen.EnPassant {
			capturedPiece = game.State.Board[input.From.R][input.To.C]
		}
		hadCapture := capturedPiece != nil || chosen.EnPassant
		notation := chessNotation(*chosen, piece, hadCapture, promotionChoice)

		if capturedPiece != nil {
			if color == "w" {
				game.CapturedW = append(game.CapturedW, *capturedPiece)
			} else {
				game.CapturedB = append(game.CapturedB, *capturedPiece)
			}
		}

		from := input.From
		to := input.To
		game.State = chessApplyMove(game.State, *chosen, promotionChoice)
		game.LastFrom = &from
		game.LastTo = &to
		game.Moves = append(game.Moves, notation)

		nextTurn := game.State.Turn
		inCheck := chessInCheck(game.State, nextTurn)
		hasMove := chessHasAnyMove(game.State, nextTurn)

		word := "Schwarz"
		if nextTurn == "w" {
			word = "Weiß"
		}
		moverWord := "Schwarz"
		if color == "w" {
			moverWord = "Weiß"
		}

		switch {
		case !hasMove && inCheck:
			game.Status = "finished"
			game.Winner = color
			game.Note = "SCHACHMATT! " + moverWord + " gewinnt!"
		case !hasMove:
			game.Status = "finished"
			game.Winner = ""
			game.Note = "PATT! Die Partie endet unentschieden."
		case inCheck:
			game.Note = "SCHACH! " + word + " muss reagieren!"
		default:
			game.Note = word + " ist am Zug."
		}

	case "reset":
		game := room.Schach
		if game == nil || game.Status != "finished" || (input.ID != game.RedID && input.ID != game.BlueID) {
			mu.Unlock()
			http.Error(w, "Eine neue Partie ist erst nach Spielende möglich.", http.StatusConflict)
			return
		}
		fresh := chessNewGame(game.RedID, game.RedName, game.BlueID, game.BlueName, game.ID)
		fresh.Status = "active"
		fresh.Note = "Weiß beginnt – wähle eine Figur."
		room.Schach = fresh

	default:
		mu.Unlock()
		http.Error(w, "Unbekannte Schach-Aktion.", http.StatusBadRequest)
		return
	}

	room.LastActive = time.Now()
	result := chessClone(room.Schach)
	mu.Unlock()
	jsonResponse(w, map[string]any{"game": result})
}
