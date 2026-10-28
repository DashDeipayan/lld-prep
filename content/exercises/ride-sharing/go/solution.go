package main

import (
	"errors"
	"time"
)

// ======================= Provided =======================

type VehicleClass int

const (
	Mini VehicleClass = iota
	Sedan
	SUV
)

type TripStatus int

const (
	Assigned TripStatus = iota
	Arrived
	InProgress
	Completed
	Cancelled
)

type Location struct{ X, Y int }

func (l Location) Distance(o Location) int { return abs(l.X-o.X) + abs(l.Y-o.Y) }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type DriverInfo struct {
	ID     string
	Class  VehicleClass
	Loc    Location
	Rating float64
}

type Rate struct{ Base, PerKm, PerMin int64 }

type MatchingStrategy interface {
	Match(pickup Location, cands []DriverInfo) (string, bool)
}

type PricingStrategy interface {
	Fare(cls VehicleClass, distance int, d time.Duration) int64
}

var (
	ErrNotFound          = errors.New("not found")
	ErrNoDriver          = errors.New("no driver available")
	ErrIllegalTransition = errors.New("illegal transition")
	ErrRiderBusy         = errors.New("rider already has an active trip")
	ErrDriverBusy        = errors.New("driver is on a trip")
)

// ======================= Your design =======================

func NewNearestDriver(maxDistance int) MatchingStrategy { panic("TODO") }

func NewStandardPricing(rates map[VehicleClass]Rate, surgePercent int64) PricingStrategy {
	panic("TODO")
}

type RideService struct {
	// TODO
}

func NewRideService(m MatchingStrategy, p PricingStrategy, clock func() time.Time) *RideService {
	panic("TODO")
}

func (s *RideService) AddDriver(id string, cls VehicleClass, loc Location, rating float64) error {
	panic("TODO")
}
func (s *RideService) SetOnline(id string, online bool) error       { panic("TODO") }
func (s *RideService) UpdateLocation(id string, loc Location) error { panic("TODO") }
func (s *RideService) RequestRide(rider string, from, to Location, cls VehicleClass) (string, error) {
	panic("TODO")
}
func (s *RideService) DriverOf(tripID string) (string, error)   { panic("TODO") }
func (s *RideService) Status(tripID string) (TripStatus, error) { panic("TODO") }
func (s *RideService) Arrived(tripID string) error              { panic("TODO") }
func (s *RideService) Start(tripID string) error                { panic("TODO") }
func (s *RideService) End(tripID string) (int64, error)         { panic("TODO") }
func (s *RideService) Cancel(tripID string) error               { panic("TODO") }
