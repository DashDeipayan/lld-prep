package main

import "errors"

type RoomType int

const (
	Single RoomType = iota
	Double
	Suite
)

var (
	ErrInvalid   = errors.New("invalid request")
	ErrNotFound  = errors.New("not found")
	ErrNoRoom    = errors.New("no room available")
	ErrBadStatus = errors.New("bad status")
)

type Hotel struct {
	// TODO
}

func NewHotel() *Hotel { panic("TODO") }

func (h *Hotel) AddRoom(number string, t RoomType, rate int64) error        { panic("TODO") }
func (h *Hotel) Availability(in, out int) (map[RoomType]int, error)         { panic("TODO") }
func (h *Hotel) Book(guest string, t RoomType, in, out int) (string, error) { panic("TODO") }
func (h *Hotel) RoomOf(id string) (string, error)                           { panic("TODO") }
func (h *Hotel) CheckIn(id string, day int) error                           { panic("TODO") }
func (h *Hotel) AddCharge(id string, amount int64) error                    { panic("TODO") }
func (h *Hotel) CheckOut(id string) (int64, error)                          { panic("TODO") }
func (h *Hotel) Cancel(id string) error                                     { panic("TODO") }
func (h *Hotel) Status(id string) (string, error)                           { panic("TODO") }
