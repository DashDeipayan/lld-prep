package main

import "errors"

// ======================= Provided =======================

type OrderLine struct {
	SKU string
	Qty int
}

type OrderStatus int

const (
	Created OrderStatus = iota
	Paid
	Shipped
	Delivered
	Cancelled
)

type PaymentGateway interface {
	Charge(orderID string, amount int64) bool
	Refund(orderID string, amount int64)
}

var (
	ErrInvalid           = errors.New("invalid request")
	ErrNotFound          = errors.New("not found")
	ErrOutOfStock        = errors.New("out of stock")
	ErrPaymentFailed     = errors.New("payment failed")
	ErrIllegalTransition = errors.New("illegal transition")
)

// ======================= Your design =======================

type OrderService struct {
	// TODO
}

func NewOrderService(gw PaymentGateway) *OrderService { panic("TODO") }

func (s *OrderService) AddProduct(sku string, price int64, stock int) error           { panic("TODO") }
func (s *OrderService) Restock(sku string, qty int) error                             { panic("TODO") }
func (s *OrderService) Stock(sku string) int                                          { panic("TODO") }
func (s *OrderService) PlaceOrder(customer string, lines []OrderLine) (string, error) { panic("TODO") }
func (s *OrderService) Total(id string) (int64, error)                                { panic("TODO") }
func (s *OrderService) Status(id string) (OrderStatus, error)                         { panic("TODO") }
func (s *OrderService) Pay(id string) error                                           { panic("TODO") }
func (s *OrderService) Ship(id string) error                                          { panic("TODO") }
func (s *OrderService) Deliver(id string) error                                       { panic("TODO") }
func (s *OrderService) Cancel(id string) error                                        { panic("TODO") }
func (s *OrderService) OrdersOf(customer string) []string                             { panic("TODO") }
