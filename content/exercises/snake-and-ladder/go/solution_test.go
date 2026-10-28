package main

import (
	"errors"
	"testing"
)

func script(rolls ...int) Dice {
	i := 0
	return DiceFunc(func() int { r := rolls[i]; i++; return r })
}

func mustBoard(t *testing.T, size int, snakes, ladders map[int]int) *Board {
	t.Helper()
	b, err := NewBoard(size, snakes, ladders)
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	return b
}

func game(t *testing.T, b *Board, d Dice) *Game {
	t.Helper()
	g, err := NewGame(b, []string{"A", "B"}, d)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func turn(t *testing.T, g *Game, want TurnResult, msg string) {
	t.Helper()
	got, err := g.PlayTurn()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
	if got != want {
		t.Errorf("%s\nexpected: %+v\n  actual: %+v", msg, want, got)
	}
}

func std(t *testing.T) *Board { return mustBoard(t, 20, map[int]int{17: 5}, map[int]int{3: 11}) }

func TestBoardValidation(t *testing.T) {
	defer guard(t)
	bad := []struct {
		name            string
		snakes, ladders map[int]int
	}{
		{"snake going up", map[int]int{5: 9}, nil},
		{"ladder going down", nil, map[int]int{9: 5}},
		{"out of range", map[int]int{25: 3}, nil},
		{"jump from final cell", map[int]int{20: 3}, nil},
		{"two jumps from same cell", map[int]int{9: 2}, map[int]int{9: 15}},
		{"chain", map[int]int{15: 4}, map[int]int{4: 12}},
	}
	for _, c := range bad {
		if _, err := NewBoard(20, c.snakes, c.ladders); !errors.Is(err, ErrInvalidBoard) {
			t.Errorf("%s: expected ErrInvalidBoard, got %v", c.name, err)
		}
	}
	if _, err := NewGame(std(t), []string{"solo"}, script(1)); !errors.Is(err, ErrNotEnoughPlayers) {
		t.Errorf("solo: %v", err)
	}
}

func TestBasicTurnsAlternate(t *testing.T) {
	defer guard(t)
	g := game(t, std(t), script(2, 4, 1))
	turn(t, g, TurnResult{"A", 2, 0, 2, false}, "A")
	turn(t, g, TurnResult{"B", 4, 0, 4, false}, "B")
	turn(t, g, TurnResult{"A", 1, 2, 11, false}, "A ladder 3->11")
	if g.Position("A") != 11 {
		t.Errorf("position %d", g.Position("A"))
	}
}

func TestSnake(t *testing.T) {
	defer guard(t)
	g := game(t, std(t), script(5, 1, 6, 6))
	g.PlayTurn()
	g.PlayTurn()
	turn(t, g, TurnResult{"A", 6, 5, 11, false}, "A 5->11 (six, rolls again)")
	turn(t, g, TurnResult{"A", 6, 11, 5, false}, "A 11->17, snake down to 5")
}

func TestSixGivesExtraRoll(t *testing.T) {
	defer guard(t)
	g := game(t, std(t), script(6, 1, 2))
	g.PlayTurn()
	if r, _ := g.PlayTurn(); r.Player != "A" {
		t.Errorf("A rolls again after a six, got %s", r.Player)
	}
	if r, _ := g.PlayTurn(); r.Player != "B" {
		t.Errorf("then B, got %s", r.Player)
	}
}

func TestThreeSixesPenalty(t *testing.T) {
	defer guard(t)
	g := game(t, mustBoard(t, 40, nil, nil), script(2, 1, 6, 6, 6, 3))
	for i := 0; i < 4; i++ {
		g.PlayTurn()
	}
	turn(t, g, TurnResult{"A", 6, 14, 2, false}, "third six -> back to turn start")
	if r, _ := g.PlayTurn(); r.Player != "B" {
		t.Errorf("turn passes after penalty, got %s", r.Player)
	}
}

func TestExactFinishAndWinner(t *testing.T) {
	defer guard(t)
	g := game(t, mustBoard(t, 10, nil, nil), script(5, 1, 6, 1, 4, 4))
	g.PlayTurn()
	g.PlayTurn()
	turn(t, g, TurnResult{"A", 6, 5, 5, false}, "overshoot stays")
	turn(t, g, TurnResult{"A", 1, 5, 6, false}, "A again after six")
	g.PlayTurn()
	turn(t, g, TurnResult{"A", 4, 6, 10, true}, "exact finish")
	if w, ok := g.Winner(); !ok || w != "A" {
		t.Errorf("winner %q %v", w, ok)
	}
	if _, err := g.PlayTurn(); !errors.Is(err, ErrGameOver) {
		t.Errorf("after game over: %v", err)
	}
}
