package main

import "errors"

var (
	ErrInvalidBoard     = errors.New("invalid board")
	ErrNotEnoughPlayers = errors.New("need at least 2 players")
	ErrGameOver         = errors.New("game over")
)

type Dice interface {
	Roll() int
}

type DiceFunc func() int

func (f DiceFunc) Roll() int { return f() }

type TurnResult struct {
	Player   string
	Roll     int
	From, To int
	Won      bool
}

type Board struct {
	// TODO
}

func NewBoard(size int, snakes, ladders map[int]int) (*Board, error) { panic("TODO") }

type Game struct {
	// TODO
}

func NewGame(b *Board, players []string, d Dice) (*Game, error) { panic("TODO") }

func (g *Game) PlayTurn() (TurnResult, error) { panic("TODO") }
func (g *Game) Position(p string) int         { panic("TODO") }
func (g *Game) Winner() (string, bool)        { panic("TODO") }
