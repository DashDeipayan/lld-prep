package main

import "errors"

type CarType int

const (
	Hatchback CarType = iota
	Sedan
	SUV
)

type Reservation struct {
	ID, Customer, CarID string
	StartDay, EndDay    int
	Cost                int64
	Status              string
}

var (
	ErrInvalid   = errors.New("invalid request")
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
	ErrBadStatus = errors.New("bad status")
)

type RentalService struct {
	// TODO
}

func NewRentalService() *RentalService { panic("TODO") }

func (r *RentalService) AddBranch(id, city string) error                               { panic("TODO") }
func (r *RentalService) AddCar(branch, carID string, t CarType, dailyRate int64) error { panic("TODO") }
func (r *RentalService) Search(city string, t CarType, start, end int) ([]string, error) {
	panic("TODO")
}
func (r *RentalService) Reserve(customer, carID string, start, end int) (Reservation, error) {
	panic("TODO")
}
func (r *RentalService) Cancel(id string) (Reservation, error)             { panic("TODO") }
func (r *RentalService) ReturnCar(id string, actualDay int) (int64, error) { panic("TODO") }
