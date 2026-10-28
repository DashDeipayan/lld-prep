package main

import (
	"errors"
	"time"
)

var ErrClosed = errors.New("queue closed")

type BoundedQueue[T any] struct {
	// TODO
}

func NewBoundedQueue[T any](capacity int) *BoundedQueue[T] { panic("TODO") }

func (q *BoundedQueue[T]) Put(x T) error                                  { panic("TODO") }
func (q *BoundedQueue[T]) Take() (T, bool)                                { panic("TODO") }
func (q *BoundedQueue[T]) Offer(x T, timeout time.Duration) (bool, error) { panic("TODO") }
func (q *BoundedQueue[T]) Poll(timeout time.Duration) (T, bool)           { panic("TODO") }
func (q *BoundedQueue[T]) Close()                                         { panic("TODO") }
func (q *BoundedQueue[T]) Len() int                                       { panic("TODO") }
