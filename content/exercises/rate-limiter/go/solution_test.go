package main

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fclock struct{ ms int64 }

func (c *fclock) now() time.Time { return time.UnixMilli(c.ms) }

func calls(r RateLimiter, c *fclock, key string, at ...int64) []bool {
	out := []bool{}
	for _, t := range at {
		c.ms = t
		out = append(out, r.Allow(key))
	}
	return out
}

func bools(t *testing.T, want, got []bool, msg string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func TestTokenBucketBurstAndRefill(t *testing.T) {
	defer guard(t)
	c := &fclock{}
	r := NewTokenBucket(3, 2, c.now)
	bools(t, []bool{true, true, true, false}, calls(r, c, "k", 0, 0, 0, 0), "burst then reject")
	bools(t, []bool{true, false}, calls(r, c, "k", 500, 500), "one token after 500ms")
	bools(t, []bool{true, true, true, false}, calls(r, c, "k", 10000, 10000, 10000, 10000), "refill caps at capacity")
}

func TestTokenBucketFractionalAccumulation(t *testing.T) {
	defer guard(t)
	c := &fclock{}
	r := NewTokenBucket(1, 1, c.now)
	bools(t, []bool{true, false, false, false, true}, calls(r, c, "k", 0, 300, 600, 900, 1000), "fractions accumulate")
}

func TestFixedWindowBoundaries(t *testing.T) {
	defer guard(t)
	c := &fclock{}
	r := NewFixedWindow(2, time.Second, c.now)
	bools(t, []bool{true, true, false, true, true, false}, calls(r, c, "k", 900, 950, 999, 1000, 1500, 1999), "aligned windows")
}

func TestSlidingWindowLog(t *testing.T) {
	defer guard(t)
	c := &fclock{}
	r := NewSlidingWindow(2, time.Second, c.now)
	bools(t, []bool{true, true, false, true, false, true}, calls(r, c, "k", 900, 950, 1000, 1900, 1949, 1950), "sliding log")
}

func TestKeysAreIndependent(t *testing.T) {
	defer guard(t)
	c := &fclock{}
	for i, r := range []RateLimiter{NewTokenBucket(1, 0, c.now), NewFixedWindow(1, time.Second, c.now), NewSlidingWindow(1, time.Second, c.now)} {
		if !(r.Allow("a") && !r.Allow("a") && r.Allow("b")) {
			t.Errorf("limiter %d: keys must have independent budgets", i)
		}
	}
}

func TestConcurrentCallersExactBudget(t *testing.T) {
	defer guard(t)
	r := NewTokenBucket(10, 0, func() time.Time { return time.UnixMilli(0) })
	var ok atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if r.Allow("hot") {
				ok.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()
	if ok.Load() != 10 {
		t.Errorf("expected exactly 10 allowed, got %d", ok.Load())
	}
	_ = fmt.Sprint
}
