package main

import "time"

type RateLimiter interface {
	Allow(key string) bool
}

type Clock func() time.Time

func NewTokenBucket(capacity int, refillPerSecond float64, c Clock) RateLimiter { panic("TODO") }

func NewFixedWindow(max int, window time.Duration, c Clock) RateLimiter { panic("TODO") }

func NewSlidingWindow(max int, window time.Duration, c Clock) RateLimiter { panic("TODO") }
