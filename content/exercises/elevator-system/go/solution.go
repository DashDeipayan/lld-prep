package main

import "errors"

type Direction int

const (
	Idle Direction = iota
	Up
	Down
)

var (
	ErrInvalidFloor     = errors.New("invalid floor")
	ErrInvalidElevator  = errors.New("invalid elevator")
	ErrInvalidDirection = errors.New("invalid direction")
)

type DispatchStrategy interface {
	Choose(cars []*Elevator, floor int, dir Direction) *Elevator
}

type StrategyFunc func([]*Elevator, int, Direction) *Elevator

func (f StrategyFunc) Choose(c []*Elevator, fl int, d Direction) *Elevator { return f(c, fl, d) }

type Elevator struct {
	// TODO
}

func NewElevator(id int) *Elevator { panic("TODO") }

func (e *Elevator) ID() int           { panic("TODO") }
func (e *Elevator) Floor() int        { panic("TODO") }
func (e *Elevator) Dir() Direction    { panic("TODO") }
func (e *Elevator) Stops() []int      { panic("TODO") }
func (e *Elevator) AddStop(floor int) { panic("TODO") }

// Step advances one tick; returns the floor where doors opened, if any.
func (e *Elevator) Step() (int, bool) { panic("TODO") }

type NearestCar struct{}

func (NearestCar) Choose(cars []*Elevator, floor int, dir Direction) *Elevator { panic("TODO") }

type ElevatorSystem struct {
	// TODO
}

func NewElevatorSystem(floors, cars int, s DispatchStrategy) *ElevatorSystem { panic("TODO") }

func (s *ElevatorSystem) RequestPickup(floor int, dir Direction) (int, error) { panic("TODO") }
func (s *ElevatorSystem) RequestDestination(id, floor int) error              { panic("TODO") }
func (s *ElevatorSystem) Step() []string                                      { panic("TODO") }
func (s *ElevatorSystem) Car(id int) *Elevator                                { panic("TODO") }
