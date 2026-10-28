package main

import (
	"errors"
	"time"
)

type Sender interface {
	Send(to, msg string) error
}

type Clock func() time.Time

var ErrRateLimited = errors.New("rate limited")

// SenderFunc adapts a function to Sender.
type SenderFunc func(to, msg string) error

func (f SenderFunc) Send(to, msg string) error { return f(to, msg) }

func WithRetry(inner Sender, maxAttempts int) Sender { panic("TODO") }

func WithRateLimit(inner Sender, clock Clock, max int, window time.Duration) Sender { panic("TODO") }

func WithLogging(inner Sender, log *[]string) Sender { panic("TODO") }
