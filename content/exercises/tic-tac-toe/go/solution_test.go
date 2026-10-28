package main

import (
	"errors"
	"testing"
)

func play(t *testing.T, g *Game, moves [][2]int) Status {
	t.Helper()
	p, s := X, InProgress
	for _, m := range moves {
		var err error
		if s, err = g.Move(p, m[0], m[1]); err != nil {
			t.Fatalf("move %v: %v", m, err)
		}
		p = 1 - p
	}
	return s
}

func st(t *testing.T, want, got Status, msg string) {
	t.Helper()
	if want != got {
		t.Errorf("%s: expected status %d, got %d", msg, want, got)
	}
}

func TestRowWin(t *testing.T) {
	defer guard(t)
	g := NewGame(3)
	st(t, XWon, play(t, g, [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {0, 2}}), "X wins top row")
	if g.Cell(0, 2) != 'X' || g.Cell(2, 2) != '.' {
		t.Errorf("cells: %c %c", g.Cell(0, 2), g.Cell(2, 2))
	}
}

func TestColumnAndDiagonalWins(t *testing.T) {
	defer guard(t)
	st(t, OWon, play(t, NewGame(3), [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {2, 2}, {2, 1}}), "O middle column")
	st(t, XWon, play(t, NewGame(3), [][2]int{{0, 0}, {0, 1}, {1, 1}, {0, 2}, {2, 2}}), "X main diagonal")
	st(t, XWon, play(t, NewGame(3), [][2]int{{0, 2}, {0, 0}, {1, 1}, {0, 1}, {2, 0}}), "X anti-diagonal")
}

func TestLargerBoard(t *testing.T) {
	defer guard(t)
	st(t, XWon, play(t, NewGame(4), [][2]int{{0, 3}, {0, 0}, {1, 2}, {1, 0}, {2, 1}, {2, 0}, {3, 0}}), "4x4 anti-diagonal")
}

func TestDraw(t *testing.T) {
	defer guard(t)
	st(t, Draw, play(t, NewGame(3), [][2]int{{0, 0}, {0, 1}, {0, 2}, {1, 1}, {1, 0}, {1, 2}, {2, 1}, {2, 0}, {2, 2}}), "full board")
}

func TestInvalidMovesChangeNothing(t *testing.T) {
	defer guard(t)
	g := NewGame(3)
	if _, err := g.Move(O, 0, 0); !errors.Is(err, ErrWrongTurn) {
		t.Errorf("X moves first: %v", err)
	}
	g.Move(X, 1, 1)
	if _, err := g.Move(O, 1, 1); !errors.Is(err, ErrOccupied) {
		t.Errorf("occupied: %v", err)
	}
	if _, err := g.Move(O, 3, 0); !errors.Is(err, ErrOutOfBounds) {
		t.Errorf("bounds: %v", err)
	}
	if _, err := g.Move(X, 0, 0); !errors.Is(err, ErrWrongTurn) {
		t.Errorf("X twice: %v", err)
	}
	if s, err := g.Move(O, 0, 0); err != nil || s != InProgress {
		t.Errorf("valid O move: %v %v", s, err)
	}
}

func TestNoMovesAfterGameOver(t *testing.T) {
	defer guard(t)
	g := NewGame(3)
	play(t, g, [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}, {0, 2}})
	if _, err := g.Move(O, 2, 2); !errors.Is(err, ErrGameOver) {
		t.Errorf("game over: %v", err)
	}
	st(t, XWon, g.Status(), "status sticks")
}
