package main

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type clk struct{ t time.Time }

func (c *clk) now() time.Time { return c.t }

func newInv(ids ...string) (*SeatInventory, *clk) {
	c := &clk{t: time.UnixMilli(1000)}
	return NewSeatInventory(ids, time.Second, c.now), c
}

func mustHold(t *testing.T, s *SeatInventory, seats ...string) string {
	t.Helper()
	id, err := s.Hold("u", seats)
	if err != nil {
		t.Fatalf("hold %v: %v", seats, err)
	}
	return id
}

func status(t *testing.T, s *SeatInventory, seat, want, msg string) {
	t.Helper()
	if got := s.Status(seat); got != want {
		t.Errorf("%s: status(%s) expected %s, got %s", msg, seat, want, got)
	}
}

func errIs(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected %v, got %v", msg, target, err)
	}
}

func TestHoldAndStatus(t *testing.T) {
	defer guard(t)
	s, _ := newInv("A1", "A2", "A3")
	if h := mustHold(t, s, "A1", "A3"); h != "H1" {
		t.Errorf("first hold id: %s", h)
	}
	status(t, s, "A1", "HELD", "held")
	if got := s.Available(); !reflect.DeepEqual(got, []string{"A2"}) {
		t.Errorf("available: %v", got)
	}
	if h := mustHold(t, s, "A2"); h != "H2" {
		t.Errorf("ids increase: %s", h)
	}
}

func TestAllOrNothing(t *testing.T) {
	defer guard(t)
	s, _ := newInv("A1", "A2", "A3")
	mustHold(t, s, "A2")
	_, err := s.Hold("v", []string{"A1", "A2", "A3"})
	errIs(t, err, ErrSeatUnavailable, "one seat taken")
	status(t, s, "A1", "AVAILABLE", "not held by failed attempt")
	status(t, s, "A3", "AVAILABLE", "not held by failed attempt")
}

func TestExpiryIsLazyAndExact(t *testing.T) {
	defer guard(t)
	s, c := newInv("A1")
	h := mustHold(t, s, "A1")
	c.t = c.t.Add(999 * time.Millisecond)
	status(t, s, "A1", "HELD", "1ms before expiry")
	c.t = c.t.Add(time.Millisecond)
	status(t, s, "A1", "AVAILABLE", "expired at now == expiresAt")
	errIs(t, s.Confirm(h), ErrHoldNotActive, "confirm expired hold")
	if h2 := mustHold(t, s, "A1"); h2 != "H2" {
		t.Errorf("re-hold after expiry: %s", h2)
	}
}

func TestConfirmAndRelease(t *testing.T) {
	defer guard(t)
	s, c := newInv("A1", "A2")
	h1 := mustHold(t, s, "A1")
	h2 := mustHold(t, s, "A2")
	if err := s.Confirm(h1); err != nil {
		t.Fatal(err)
	}
	if err := s.Release(h2); err != nil {
		t.Fatal(err)
	}
	c.t = c.t.Add(10 * time.Second)
	status(t, s, "A1", "BOOKED", "booked never expires")
	status(t, s, "A2", "AVAILABLE", "released")
	errIs(t, s.Confirm(h1), ErrHoldNotActive, "double confirm")
	errIs(t, s.Release(h2), ErrHoldNotActive, "double release")
	errIs(t, s.Confirm("H99"), ErrHoldNotActive, "unknown hold")
	_, err := s.Hold("w", []string{"A1"})
	errIs(t, err, ErrSeatUnavailable, "booked seat cannot be held")
}

func TestInvalidInput(t *testing.T) {
	defer guard(t)
	s, _ := newInv("A1", "A2")
	for _, req := range [][]string{{}, {"A1", "A1"}, {"Z9"}} {
		_, err := s.Hold("u", req)
		errIs(t, err, ErrUnknownSeat, fmt.Sprint("invalid request ", req))
	}
	if got := s.Available(); !reflect.DeepEqual(got, []string{"A1", "A2"}) {
		t.Errorf("nothing should change: %v", got)
	}
}

func TestNoDoubleBookingUnderContention(t *testing.T) {
	defer guard(t)
	var ids []string
	for i := 0; i < 12; i++ {
		ids = append(ids, fmt.Sprintf("S%c", 'a'+i))
	}
	s := NewSeatInventory(ids, time.Minute, func() time.Time { return time.Unix(0, 0) })
	type res struct {
		hold  string
		seats []string
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan res, 64)
	for g := 0; g < 64; g++ {
		uniq := []string{}
		seen := map[string]bool{}
		for _, k := range []int{g % 12, (g*5 + 1) % 12, (g*7 + 3) % 12} {
			if !seen[ids[k]] {
				seen[ids[k]] = true
				uniq = append(uniq, ids[k])
			}
		}
		wg.Add(1)
		go func(want []string) {
			defer wg.Done()
			<-start
			if h, err := s.Hold("u", want); err == nil {
				results <- res{h, want}
			}
		}(uniq)
	}
	close(start)
	wg.Wait()
	close(results)
	owner := map[string]string{}
	ok := 0
	for r := range results {
		ok++
		for _, seat := range r.seats {
			if prev, dup := owner[seat]; dup {
				t.Fatalf("seat %s held by both %s and %s", seat, prev, r.hold)
			}
			owner[seat] = r.hold
			status(t, s, seat, "HELD", "successful hold's seat")
		}
	}
	if ok == 0 {
		t.Error("at least one hold must succeed")
	}
}
