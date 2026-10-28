package main

import "errors"

var (
	ErrUnknownAccount    = errors.New("unknown account")
	ErrDuplicateAccount  = errors.New("duplicate account")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrSameAccount       = errors.New("cannot transfer to the same account")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Txn struct {
	ID, From, To string
	Amount       int64
}

type Wallet struct {
	// TODO
}

func NewWallet() *Wallet { panic("TODO") }

func (w *Wallet) OpenAccount(id string) error                                 { panic("TODO") }
func (w *Wallet) Deposit(id string, amount int64) (string, error)             { panic("TODO") }
func (w *Wallet) Transfer(key, from, to string, amount int64) (string, error) { panic("TODO") }
func (w *Wallet) Balance(id string) (int64, error)                            { panic("TODO") }
func (w *Wallet) History(id string) ([]Txn, error)                            { panic("TODO") }
