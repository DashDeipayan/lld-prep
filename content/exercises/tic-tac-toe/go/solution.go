package main

import "errors"

type Player int

const (
	X Player = iota
	O
)

type Status int

const (
	InProgress Status = iota
	XWon
	OWon
	Draw
)

var (
	ErrGameOver    = errors.New("game over")
	ErrOutOfBounds = errors.New("out of bounds")
	ErrOccupied    = errors.New("cell occupied")
	ErrWrongTurn   = errors.New("wrong turn")
)

type Game struct {
	// TODO
}

func NewGame(n int) *Game { panic("TODO") }

func (g *Game) Move(p Player, row, col int) (Status, error) { panic("TODO") }
func (g *Game) Status() Status                              { panic("TODO") }
func (g *Game) Cell(row, col int) byte                      { panic("TODO") }
