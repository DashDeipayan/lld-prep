package main

import "errors"

var (
	ErrInvalidCoin       = errors.New("invalid coin")
	ErrInvalidQuantity   = errors.New("quantity must be positive")
	ErrBusy              = errors.New("busy: transaction in progress")
	ErrSoldOut           = errors.New("sold out")
	ErrInsertMoneyFirst  = errors.New("insert money first")
	ErrUnknownProduct    = errors.New("unknown product")
	ErrOutOfStock        = errors.New("out of stock")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

var Coins = map[int]bool{5: true, 10: true, 25: true, 100: true}

type Product struct {
	Code  string
	Name  string
	Price int
}

type Dispense struct {
	Product Product
	Change  int
}

// TODO: a State interface with one implementation per state

type VendingMachine struct {
	// TODO
}

func NewVendingMachine() *VendingMachine { panic("TODO") }

func (m *VendingMachine) Stock(p Product, qty int) error       { panic("TODO") }
func (m *VendingMachine) InsertCoin(coin int) error            { panic("TODO") }
func (m *VendingMachine) Select(code string) (Dispense, error) { panic("TODO") }
func (m *VendingMachine) Cancel() int                          { panic("TODO") }
func (m *VendingMachine) Balance() int                         { panic("TODO") }
func (m *VendingMachine) State() string                        { panic("TODO") }
func (m *VendingMachine) Quantity(code string) int             { panic("TODO") }
