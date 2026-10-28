package main

import "errors"

type Color int

const (
	White Color = iota
	Black
)

type Kind int

const (
	King Kind = iota
	Queen
	Rook
	Bishop
	Knight
	Pawn
)

var (
	ErrIllegalMove   = errors.New("illegal move")
	ErrWrongTurn     = errors.New("wrong turn")
	ErrInvalidSquare = errors.New("invalid square")
)

// TODO: a Square value type and a Piece interface with implementations

type Board struct {
	// TODO
}

func NewEmptyBoard() *Board    { panic("TODO") }
func NewStandardBoard() *Board { panic("TODO") }

func (b *Board) Place(sq string, c Color, k Kind) error { panic("TODO") }
func (b *Board) PieceAt(sq string) string               { panic("TODO") }
func (b *Board) MovesFrom(sq string) ([]string, error)  { panic("TODO") }
func (b *Board) Move(from, to string) error             { panic("TODO") }
func (b *Board) Turn() Color                            { panic("TODO") }
