package main

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type env struct {
	now     time.Time
	approve bool
	charges []string
	mu      sync.Mutex
}

func setup(t *testing.T) (*BookingService, *env) {
	t.Helper()
	e := &env{now: time.Unix(0, 0), approve: true}
	gw := GatewayFunc(func(u string, a int64) bool {
		e.mu.Lock()
		defer e.mu.Unlock()
		e.charges = append(e.charges, fmt.Sprintf("%s:%d", u, a))
		return e.approve
	})
	b := NewBookingService(gw, func() time.Time { return e.now }, 5*time.Minute)
	for _, err := range []error{
		b.AddTheatre("T1", "Bengaluru"), b.AddTheatre("T2", "Mumbai"),
		b.AddScreen("T1", "SC1", 2, 1, 3), b.AddScreen("T2", "SC2", 1, 1, 2),
		b.AddMovie("M1", "Dune", 120), b.AddMovie("M2", "Up", 90),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	return b, e
}

func at(min int) time.Time { return time.Unix(0, 0).Add(time.Duration(min) * time.Minute) }

func addShow(t *testing.T, b *BookingService, movie, screen string, min int) string {
	t.Helper()
	id, err := b.AddShow(movie, screen, at(min), 200, 350)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func is(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected %v, got %v", msg, target, err)
	}
}

func deep(t *testing.T, want, got any, msg string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func book(t *testing.T, b *BookingService, show, user string, seats ...string) *Booking {
	t.Helper()
	h, err := b.Hold(show, user, seats)
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	bk, err := b.Confirm(h)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	return bk
}

func TestScreenLayoutAndAvailability(t *testing.T) {
	defer guard(t)
	b, _ := setup(t)
	s := addShow(t, b, "M1", "SC1", 60)
	seats, _ := b.AvailableSeats(s)
	deep(t, []string{"A1", "A2", "A3", "B1", "B2", "B3", "C1", "C2", "C3"}, seats, "layout")
	is(t, b.AddScreen("T9", "X", 1, 1, 1), ErrNotFound, "unknown theatre")
}

func TestShowOverlapOnSameScreen(t *testing.T) {
	defer guard(t)
	b, _ := setup(t)
	addShow(t, b, "M1", "SC1", 60)
	_, err := b.AddShow("M2", "SC1", at(179), 200, 350)
	is(t, err, ErrOverlap, "overlap")
	deep(t, "S2", addShow(t, b, "M2", "SC1", 180), "back-to-back ok")
	deep(t, "S3", addShow(t, b, "M1", "SC2", 60), "other screen")
}

func TestFindShowsByCityMovieAndTime(t *testing.T) {
	defer guard(t)
	b, e := setup(t)
	late := addShow(t, b, "M1", "SC1", 300)
	early := addShow(t, b, "M1", "SC1", 60)
	addShow(t, b, "M1", "SC2", 60)
	addShow(t, b, "M2", "SC1", 500)
	deep(t, []string{early, late}, b.FindShows("bengaluru", "M1"), "case-insensitive, sorted")
	e.now = at(61)
	deep(t, []string{late}, b.FindShows("Bengaluru", "M1"), "started excluded")
}

func TestHoldConfirmPricing(t *testing.T) {
	defer guard(t)
	b, e := setup(t)
	s := addShow(t, b, "M1", "SC1", 60)
	h, err := b.Hold(s, "asha", []string{"C1", "A1"})
	if err != nil || h != "H1" {
		t.Fatalf("hold: %s %v", h, err)
	}
	bk, err := b.Confirm(h)
	if err != nil {
		t.Fatal(err)
	}
	deep(t, Booking{"B1", s, "asha", []string{"A1", "C1"}, 550, "CONFIRMED"}, *bk, "booking")
	deep(t, []string{"asha:550"}, e.charges, "charged once")
	_, err = b.Hold(s, "ravi", []string{"A1"})
	is(t, err, ErrSeatUnavailable, "booked seat")
	_, err = b.Confirm(h)
	is(t, err, ErrHoldNotActive, "hold reused")
}

func TestPaymentFailureReleasesSeats(t *testing.T) {
	defer guard(t)
	b, e := setup(t)
	s := addShow(t, b, "M1", "SC1", 60)
	h, _ := b.Hold(s, "asha", []string{"A1", "A2"})
	e.approve = false
	_, err := b.Confirm(h)
	is(t, err, ErrPaymentFailed, "declined")
	seats, _ := b.AvailableSeats(s)
	deep(t, 9, len(seats), "released")
}

func TestHoldExpiryAndShowStart(t *testing.T) {
	defer guard(t)
	b, e := setup(t)
	s := addShow(t, b, "M1", "SC1", 60)
	h, _ := b.Hold(s, "asha", []string{"A1"})
	e.now = e.now.Add(5 * time.Minute)
	_, err := b.Confirm(h)
	is(t, err, ErrHoldNotActive, "expired")
	deep(t, 0, len(e.charges), "no charge")
	if _, err := b.Hold(s, "ravi", []string{"A1"}); err != nil {
		t.Errorf("re-hold after expiry: %v", err)
	}
	e.now = at(60)
	_, err = b.Hold(s, "x", []string{"B1"})
	is(t, err, ErrShowStarted, "show started")
}

func TestCancellation(t *testing.T) {
	defer guard(t)
	b, e := setup(t)
	s := addShow(t, b, "M1", "SC1", 60)
	bk := book(t, b, s, "asha", "B2")
	c, err := b.Cancel(bk.ID)
	if err != nil || c.Status != "CANCELLED" {
		t.Fatalf("cancel: %v %v", c, err)
	}
	seats, _ := b.AvailableSeats(s)
	if !reflect.DeepEqual(seats, []string{"A1", "A2", "A3", "B1", "B2", "B3", "C1", "C2", "C3"}) {
		t.Errorf("seat back on sale: %v", seats)
	}
	_, err = b.Cancel(bk.ID)
	is(t, err, ErrInvalid, "double cancel")
	bk2 := book(t, b, s, "ravi", "B2")
	e.now = at(61)
	_, err = b.Cancel(bk2.ID)
	is(t, err, ErrShowStarted, "after start")
}

func TestNoDoubleBookingUnderContention(t *testing.T) {
	defer guard(t)
	b, _ := setup(t)
	s := addShow(t, b, "M1", "SC1", 60)
	var wg sync.WaitGroup
	var mu sync.Mutex
	owner := map[string]string{}
	var conflict string
	start := make(chan struct{})
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			want := []string{fmt.Sprintf("A%d", 1+i%3), fmt.Sprintf("C%d", 1+(i/3)%3)}
			h, err := b.Hold(s, fmt.Sprint("u", i), want)
			if err != nil {
				return
			}
			bk, err := b.Confirm(h)
			if err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for _, seat := range bk.Seats {
				if prev, dup := owner[seat]; dup {
					conflict = fmt.Sprintf("seat %s sold to %s and %s", seat, prev, bk.ID)
				}
				owner[seat] = bk.ID
			}
		}(i)
	}
	close(start)
	wg.Wait()
	if conflict != "" {
		t.Fatal(conflict)
	}
	if len(owner) == 0 {
		t.Error("some bookings should succeed")
	}
}
