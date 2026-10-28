package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func moves(t *testing.T, b *Board, sq string, want []string, msg string) {
	t.Helper()
	got, err := b.MovesFrom(sq)
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
	if len(want) == 0 && len(got) == 0 {
		return
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func place(b *Board, sq string, c Color, k Kind) { _ = b.Place(sq, c, k) }

func TestStandardOpeningMoves(t *testing.T) {
	defer guard(t)
	b := NewStandardBoard()
	moves(t, b, "b1", []string{"a3", "c3"}, "knight jumps over pawns")
	moves(t, b, "e2", []string{"e3", "e4"}, "pawn single + double")
	moves(t, b, "a1", nil, "rook boxed in")
	moves(t, b, "g8", []string{"f6", "h6"}, "black knight")
	moves(t, b, "d7", []string{"d5", "d6"}, "black pawn moves down")
	moves(t, b, "e5", nil, "empty square")
	if b.PieceAt("e1") != "wK" || b.PieceAt("b8") != "bN" {
		t.Errorf("pieceAt: %s %s", b.PieceAt("e1"), b.PieceAt("b8"))
	}
}

func TestSlidersOnEmptyBoard(t *testing.T) {
	defer guard(t)
	b := NewEmptyBoard()
	place(b, "d4", White, Rook)
	if got, _ := b.MovesFrom("d4"); len(got) != 14 {
		t.Errorf("rook in middle: %d moves", len(got))
	}
	place(b, "a1", White, Bishop)
	moves(t, b, "a1", []string{"b2", "c3"}, "bishop stops before own rook")
	place(b, "h8", Black, Queen)
	got, _ := b.MovesFrom("h8")
	var lines []string
	for _, s := range got {
		if strings.HasSuffix(s, "8") || strings.HasPrefix(s, "h") {
			lines = append(lines, s)
		}
	}
	want := []string{"a8", "b8", "c8", "d8", "e8", "f8", "g8", "h1", "h2", "h3", "h4", "h5", "h6", "h7"}
	if !reflect.DeepEqual(want, lines) {
		t.Errorf("queen rank+file\nexpected: %v\n  actual: %v", want, lines)
	}
	has := func(s string) bool {
		for _, x := range got {
			if x == s {
				return true
			}
		}
		return false
	}
	if !has("d4") || has("c3") {
		t.Error("queen diagonal must capture d4 and stop there")
	}
}

func TestBlockingAndCaptures(t *testing.T) {
	defer guard(t)
	b := NewEmptyBoard()
	place(b, "a1", White, Rook)
	place(b, "a4", Black, Pawn)
	place(b, "c1", White, Knight)
	moves(t, b, "a1", []string{"a2", "a3", "a4", "b1"}, "capture a4, blocked by own knight")
}

func TestKingAndKnight(t *testing.T) {
	defer guard(t)
	b := NewEmptyBoard()
	place(b, "e4", White, King)
	moves(t, b, "e4", []string{"d3", "d4", "d5", "e3", "e5", "f3", "f4", "f5"}, "king centre")
	place(b, "a1", Black, King)
	moves(t, b, "a1", []string{"a2", "b1", "b2"}, "king corner")
	place(b, "g1", White, Knight)
	moves(t, b, "g1", []string{"e2", "f3", "h3"}, "knight near edge")
}

func TestPawnRules(t *testing.T) {
	defer guard(t)
	b := NewEmptyBoard()
	place(b, "e2", White, Pawn)
	place(b, "e3", Black, Knight)
	moves(t, b, "e2", nil, "blocked pawn")
	place(b, "d3", Black, Pawn)
	place(b, "f3", White, Pawn)
	moves(t, b, "e2", []string{"d3"}, "captures enemy only")
	place(b, "c5", White, Pawn)
	moves(t, b, "c5", []string{"c6"}, "no double step off start rank")
	place(b, "h7", Black, Pawn)
	place(b, "h5", White, Rook)
	moves(t, b, "h7", []string{"h6"}, "double step needs both empty")
}

func TestMoveTurnsAndCapture(t *testing.T) {
	defer guard(t)
	b := NewStandardBoard()
	if b.Turn() != White {
		t.Fatal("white first")
	}
	if err := b.Move("e7", "e5"); !errors.Is(err, ErrWrongTurn) {
		t.Errorf("black first: %v", err)
	}
	for _, m := range [][2]string{{"e2", "e4"}, {"d7", "d5"}, {"e4", "d5"}} {
		if err := b.Move(m[0], m[1]); err != nil {
			t.Fatalf("move %v: %v", m, err)
		}
	}
	if b.PieceAt("d5") != "wP" || b.PieceAt("e4") != "." {
		t.Errorf("capture: d5=%s e4=%s", b.PieceAt("d5"), b.PieceAt("e4"))
	}
	if err := b.Move("d8", "d4"); !errors.Is(err, ErrIllegalMove) {
		t.Errorf("queen path stops at capture on d5: %v", err)
	}
	if err := b.Move("e5", "e4"); !errors.Is(err, ErrIllegalMove) {
		t.Errorf("no piece: %v", err)
	}
	if err := b.Move("z9", "e4"); !errors.Is(err, ErrInvalidSquare) {
		t.Errorf("invalid square: %v", err)
	}
}
