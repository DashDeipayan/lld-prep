package main

import (
	"fmt"
	"time"
)

// Current production code: works single-threaded, breaks under load.

var instance *IDGenerator

type IDGenerator struct {
	counter int64
}

func newIDGenerator() *IDGenerator {
	time.Sleep(20 * time.Millisecond) // simulate expensive init; keep this
	return &IDGenerator{}
}

func GetIDGenerator() *IDGenerator {
	if instance == nil {
		instance = newIDGenerator()
	}
	return instance
}

func (g *IDGenerator) NextID() int64 {
	v := g.counter
	v++
	g.counter = v
	return v
}

func (g *IDGenerator) NextIDWithPrefix(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, g.NextID())
}
