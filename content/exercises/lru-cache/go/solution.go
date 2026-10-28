package main

// Do not use container/list: build the doubly linked list yourself.

type LRU[K comparable, V any] struct {
	// TODO
}

func NewLRU[K comparable, V any](capacity int, onEvict func(K, V)) *LRU[K, V] { panic("TODO") }

func (c *LRU[K, V]) Get(k K) (V, bool) { panic("TODO") }
func (c *LRU[K, V]) Put(k K, v V)      { panic("TODO") }
func (c *LRU[K, V]) Len() int          { panic("TODO") }

// Keys returns keys from most to least recently used.
func (c *LRU[K, V]) Keys() []K { panic("TODO") }
