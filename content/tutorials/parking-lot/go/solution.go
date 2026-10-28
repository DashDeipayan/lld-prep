package main

import (
	"errors"
	"time"
)

// ======================= Provided (do not change) =======================

type SpotSize int

const (
	Small SpotSize = iota
	Medium
	Large
)

type VehicleType int

const (
	Bike VehicleType = iota
	Car
	Truck
)

// MinSize is the smallest spot a vehicle type fits in.
func (t VehicleType) MinSize() SpotSize { return [...]SpotSize{Small, Medium, Large}[t] }

type Vehicle struct {
	Plate string
	Type  VehicleType
}

type Ticket struct {
	ID, Plate, SpotID string
	Entry             time.Time
}

type Receipt struct {
	TicketID, Plate, SpotID string
	Minutes                 int64
	Fee                     int64
}

type Event struct {
	Kind     string // "PARKED" or "LEFT"
	TicketID string
	SpotID   string
	Size     SpotSize
}

type PricingStrategy interface {
	Fee(t VehicleType, d time.Duration) int64
}

// SpotSelector receives free, fitting spots in lot order; return nil if none is acceptable.
type SpotSelector interface {
	Select(candidates []*Spot, v Vehicle) *Spot
}

type ParkingListener interface {
	OnEvent(e Event)
}

type ListenerFunc func(Event)

func (f ListenerFunc) OnEvent(e Event) { f(e) }

type SelectorFunc func([]*Spot, Vehicle) *Spot

func (f SelectorFunc) Select(c []*Spot, v Vehicle) *Spot { return f(c, v) }

var (
	ErrSpotUnavailable = errors.New("spot unavailable")
	ErrDuplicateFloor  = errors.New("duplicate floor")
	ErrLotFull         = errors.New("parking lot full")
	ErrAlreadyParked   = errors.New("vehicle already parked")
	ErrUnknownTicket   = errors.New("unknown or used ticket")
)

// ============================ Step 1 ============================

type Spot struct {
	// TODO
}

func NewSpot(id string, floor int, size SpotSize) *Spot { panic("TODO") }

func (s *Spot) ID() string              { panic("TODO") }
func (s *Spot) Floor() int              { panic("TODO") }
func (s *Spot) Size() SpotSize          { panic("TODO") }
func (s *Spot) CanFit(v Vehicle) bool   { panic("TODO") }
func (s *Spot) IsFree() bool            { panic("TODO") }
func (s *Spot) Occupy(v Vehicle) error  { panic("TODO") }
func (s *Spot) Vacate() (Vehicle, bool) { panic("TODO") }

// ============================ Step 3 ============================

type NearestSelector struct{}

func (NearestSelector) Select(candidates []*Spot, v Vehicle) *Spot { panic("TODO") }

// ============================ Step 4 ============================

func NewHourlyPricing(ratePerHour map[VehicleType]int64) PricingStrategy { panic("TODO") }

// ======================= Steps 2, 3, 4, 5, 6 =======================

type ParkingLot struct {
	// TODO
}

func NewParkingLot(pricing PricingStrategy, selector SpotSelector, clock func() time.Time) *ParkingLot {
	panic("TODO")
}

func (l *ParkingLot) AddFloor(floor, small, medium, large int) error { panic("TODO") }
func (l *ParkingLot) Available(size SpotSize) int                    { panic("TODO") }
func (l *ParkingLot) Park(v Vehicle) (*Ticket, error)                { panic("TODO") }
func (l *ParkingLot) Unpark(ticketID string) (*Receipt, error)       { panic("TODO") }
func (l *ParkingLot) AddListener(pl ParkingListener)                 { panic("TODO") }

// ============================ Step 5 ============================

type DisplayBoard struct {
	// TODO
}

func NewDisplayBoard(lot *ParkingLot) *DisplayBoard { panic("TODO") }

func (b *DisplayBoard) OnEvent(e Event)        { panic("TODO") }
func (b *DisplayBoard) Free(size SpotSize) int { panic("TODO") }
