package main

import "errors"

// ======================= Provided =======================

var (
	ErrInvalidSplit  = errors.New("invalid split")
	ErrUnknownUser   = errors.New("unknown user")
	ErrDuplicateUser = errors.New("duplicate user")
	ErrInvalidAmount = errors.New("invalid amount")
	ErrOverSettle    = errors.New("settling more than owed")
)

type Transfer struct {
	From, To string
	Amount   int64
}

// Split returns participant → share; shares must sum to amount.
type Split interface {
	Shares(amount int64, participants []string) (map[string]int64, error)
}

// ======================= Step 1 =======================

type EqualSplit struct{}

func (EqualSplit) Shares(amount int64, participants []string) (map[string]int64, error) {
	panic("TODO")
}

func NewExactSplit(amounts map[string]int64) Split { panic("TODO") }

func NewPercentSplit(percents map[string]int) Split { panic("TODO") }

// ======================= Steps 2-5 =======================

type Splitwise struct {
	// TODO
}

func NewSplitwise() *Splitwise { panic("TODO") }

func (s *Splitwise) AddUser(id string) error { panic("TODO") }
func (s *Splitwise) AddExpense(paidBy string, amount int64, participants []string, split Split) (string, error) {
	panic("TODO")
}
func (s *Splitwise) Owes(debtor, creditor string) int64         { panic("TODO") }
func (s *Splitwise) Balance(user string) int64                  { panic("TODO") }
func (s *Splitwise) Settle(from, to string, amount int64) error { panic("TODO") }
func (s *Splitwise) Simplify() []Transfer                       { panic("TODO") }
