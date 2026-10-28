package main

import "errors"

// ======================= Provided =======================

var (
	ErrNotFound        = errors.New("not found")
	ErrDuplicate       = errors.New("duplicate")
	ErrLimitReached    = errors.New("loan limit reached")
	ErrNoCopyAvailable = errors.New("no copy available")
	ErrNotOnLoan       = errors.New("copy is not on loan")
	ErrCopyAvailable   = errors.New("a copy is available; borrow it instead")
	ErrAlreadyReserved = errors.New("already reserved")
)

type FinePolicy interface {
	Fine(daysLate int) int64
}

type Notifier interface {
	Notify(memberID, msg string)
}

type NotifierFunc func(memberID, msg string)

func (f NotifierFunc) Notify(m, s string) { f(m, s) }

type Loan struct {
	ID, Barcode, ISBN, MemberID string
	DueDay                      int
}

// ======================= Your design =======================

func NewPerDayFine(perDay, cap int64) FinePolicy { panic("TODO") }

type Library struct {
	// TODO
}

func NewLibrary(f FinePolicy, today func() int, n Notifier) *Library { panic("TODO") }

func (l *Library) AddBook(isbn, title, author string) error   { panic("TODO") }
func (l *Library) AddCopy(isbn, barcode string) error         { panic("TODO") }
func (l *Library) AddMember(id string, maxLoans int) error    { panic("TODO") }
func (l *Library) Borrow(memberID, isbn string) (Loan, error) { panic("TODO") }
func (l *Library) GiveBack(barcode string) (int64, error)     { panic("TODO") }
func (l *Library) Reserve(memberID, isbn string) error        { panic("TODO") }
func (l *Library) AvailableCopies(isbn string) int            { panic("TODO") }
func (l *Library) Search(query string) []string               { panic("TODO") }
