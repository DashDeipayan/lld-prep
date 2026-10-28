package main

import (
	"errors"
	"time"
)

var (
	ErrSeatUnavailable = errors.New("seat unavailable")
	ErrUnknownSeat     = errors.New("unknown or invalid seat")
	ErrHoldNotActive   = errors.New("hold not active")
)

type SeatInventory struct {
	// TODO
}

func NewSeatInventory(seatIDs []string, ttl time.Duration, clock func() time.Time) *SeatInventory {
	panic("TODO")
}

func (s *SeatInventory) Hold(user string, seats []string) (string, error) { panic("TODO") }
func (s *SeatInventory) Confirm(holdID string) error                      { panic("TODO") }
func (s *SeatInventory) Release(holdID string) error                      { panic("TODO") }
func (s *SeatInventory) Status(seat string) string                        { panic("TODO") }
func (s *SeatInventory) Available() []string                              { panic("TODO") }
