package main

import "errors"

var (
	ErrDuplicateMethod   = errors.New("payment method already registered")
	ErrUnsupportedMethod = errors.New("unsupported payment method")
	ErrInvalidAmount     = errors.New("amount must be positive")
)

type Receipt struct {
	Method string
	Amount int64
	Fee    int64
}

type PaymentProcessor interface {
	Method() string
	Pay(amountCents int64) (Receipt, error)
}

// TODO: card, upi, wallet processors

type PaymentFactory struct {
	// TODO
}

func NewPaymentFactory() *PaymentFactory { panic("TODO") }

func (f *PaymentFactory) Register(method string, ctor func() PaymentProcessor) error { panic("TODO") }

func (f *PaymentFactory) Create(method string) (PaymentProcessor, error) { panic("TODO") }

func (f *PaymentFactory) SupportedMethods() []string { panic("TODO") }

func NewDefaultFactory() *PaymentFactory { panic("TODO") }
