package main

import "errors"

// ======================= Provided =======================

type Booking struct {
	ID, Room, Organizer string
	Start, End          int
}

type Slot struct{ Start, End int }

var (
	ErrInvalid  = errors.New("invalid request")
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrNoRoom   = errors.New("no room available")
)

// ======================= Your design =======================

type Scheduler struct {
	// TODO
}

func NewScheduler() *Scheduler { panic("TODO") }

func (s *Scheduler) AddRoom(id string, capacity int) error                             { panic("TODO") }
func (s *Scheduler) Book(organizer string, start, end, attendees int) (Booking, error) { panic("TODO") }
func (s *Scheduler) BookRoom(room, organizer string, start, end int) (Booking, error)  { panic("TODO") }
func (s *Scheduler) Cancel(id string) error                                            { panic("TODO") }
func (s *Scheduler) FreeRooms(start, end, attendees int) ([]string, error)             { panic("TODO") }
func (s *Scheduler) Schedule(room string) ([]Booking, error)                           { panic("TODO") }
func (s *Scheduler) FreeSlots(room string, from, to int) ([]Slot, error)               { panic("TODO") }
