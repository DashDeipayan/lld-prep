package main

import (
	"errors"
	"time"
)

type AuctionListener interface {
	OnEvent(auctionID, event string)
}

type ListenerFunc func(auctionID, event string)

func (f ListenerFunc) OnEvent(a, e string) { f(a, e) }

type Result struct {
	AuctionID, Winner string
	Price             int64
}

var (
	ErrInvalid        = errors.New("invalid request")
	ErrNotFound       = errors.New("not found")
	ErrClosed         = errors.New("auction closed")
	ErrAlreadyLeading = errors.New("already leading")
	ErrBidTooLow      = errors.New("bid too low")
	ErrNotEnded       = errors.New("auction still running")
)

type AuctionHouse struct {
	// TODO
}

func NewAuctionHouse(clock func() time.Time) *AuctionHouse { panic("TODO") }

func (h *AuctionHouse) Create(seller, item string, startingPrice, minIncrement int64, end time.Time) (string, error) {
	panic("TODO")
}
func (h *AuctionHouse) Subscribe(id string, l AuctionListener) error   { panic("TODO") }
func (h *AuctionHouse) PlaceBid(id, bidder string, amount int64) error { panic("TODO") }
func (h *AuctionHouse) Leader(id string) (string, bool)                { panic("TODO") }
func (h *AuctionHouse) Highest(id string) int64                        { panic("TODO") }
func (h *AuctionHouse) EndTime(id string) time.Time                    { panic("TODO") }
func (h *AuctionHouse) Close(id string) (Result, error)                { panic("TODO") }
