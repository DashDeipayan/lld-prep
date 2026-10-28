package main

import (
	"errors"
	"time"
)

// ======================= Provided =======================

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalid         = errors.New("invalid request")
	ErrOverlap         = errors.New("show overlaps")
	ErrSeatUnavailable = errors.New("seat unavailable")
	ErrHoldNotActive   = errors.New("hold not active")
	ErrPaymentFailed   = errors.New("payment failed")
	ErrShowStarted     = errors.New("show already started")
)

type PaymentGateway interface {
	Charge(user string, amount int64) bool
}

type GatewayFunc func(user string, amount int64) bool

func (f GatewayFunc) Charge(u string, a int64) bool { return f(u, a) }

type Booking struct {
	ID, ShowID, User string
	Seats            []string // sorted
	Amount           int64
	Status           string // CONFIRMED / CANCELLED
}

// ======================= Your design =======================

type BookingService struct {
	// TODO
}

func NewBookingService(gw PaymentGateway, clock func() time.Time, holdTTL time.Duration) *BookingService {
	panic("TODO")
}

func (b *BookingService) AddTheatre(id, city string) error { panic("TODO") }
func (b *BookingService) AddScreen(theatreID, screenID string, regularRows, premiumRows, seatsPerRow int) error {
	panic("TODO")
}
func (b *BookingService) AddMovie(id, title string, minutes int) error { panic("TODO") }
func (b *BookingService) AddShow(movieID, screenID string, start time.Time, regular, premium int64) (string, error) {
	panic("TODO")
}
func (b *BookingService) FindShows(city, movieID string) []string                  { panic("TODO") }
func (b *BookingService) AvailableSeats(showID string) ([]string, error)           { panic("TODO") }
func (b *BookingService) Hold(showID, user string, seats []string) (string, error) { panic("TODO") }
func (b *BookingService) Confirm(holdID string) (*Booking, error)                  { panic("TODO") }
func (b *BookingService) Cancel(bookingID string) (*Booking, error)                { panic("TODO") }
