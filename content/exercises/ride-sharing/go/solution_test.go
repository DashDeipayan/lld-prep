package main

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

var rates = map[VehicleClass]Rate{Mini: {50, 10, 2}, Sedan: {80, 15, 3}, SUV: {100, 20, 5}}

type rclock struct{ t time.Time }

func (c *rclock) now() time.Time { return c.t }

func svc(t *testing.T) (*RideService, *rclock) {
	t.Helper()
	c := &rclock{t: time.Unix(0, 0)}
	s := NewRideService(NewNearestDriver(10), NewStandardPricing(rates, 100), c.now)
	s.AddDriver("d1", Mini, Location{0, 0}, 4.5)
	s.AddDriver("d2", Mini, Location{5, 5}, 4.9)
	s.AddDriver("d3", Sedan, Location{1, 1}, 4.7)
	for _, d := range []string{"d1", "d2", "d3"} {
		s.SetOnline(d, true)
	}
	return s, c
}

func ride(t *testing.T, s *RideService, rider string, from, to Location, cls VehicleClass) (string, string) {
	t.Helper()
	id, err := s.RequestRide(rider, from, to, cls)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	d, _ := s.DriverOf(id)
	return id, d
}

func is(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected %v, got %v", msg, target, err)
	}
}

func TestNearestStrategyTieBreaks(t *testing.T) {
	defer guard(t)
	m := NewNearestDriver(5)
	p := Location{0, 0}
	a := DriverInfo{"a", Mini, Location{2, 1}, 4.1}
	b := DriverInfo{"b", Mini, Location{1, 2}, 4.8}
	c := DriverInfo{"c", Mini, Location{3, 0}, 4.8}
	z := DriverInfo{"z", Mini, Location{0, 9}, 5.0}
	if id, _ := m.Match(p, []DriverInfo{a, b, c, z}); id != "b" {
		t.Errorf("rating tie-break: %s", id)
	}
	if id, _ := m.Match(p, []DriverInfo{c, b}); id != "b" {
		t.Errorf("id tie-break: %s", id)
	}
	if _, ok := m.Match(p, []DriverInfo{z}); ok {
		t.Error("out of range must not match")
	}
}

func TestPricing(t *testing.T) {
	defer guard(t)
	p := NewStandardPricing(rates, 100)
	if got := p.Fare(Mini, 7, 2*time.Minute+time.Millisecond); got != 50+10*7+2*3 {
		t.Errorf("minutes rounded up: %d", got)
	}
	if got := NewStandardPricing(rates, 150).Fare(Sedan, 4, 10*time.Minute); got != (80+15*4+3*10)*150/100 {
		t.Errorf("surge: %d", got)
	}
}

func TestRequestMatchesNearestAvailableOfClass(t *testing.T) {
	defer guard(t)
	s, _ := svc(t)
	r1, d := ride(t, s, "asha", Location{1, 0}, Location{9, 9}, Mini)
	if r1 != "R1" || d != "d1" {
		t.Errorf("first ride: %s %s", r1, d)
	}
	if st, _ := s.Status(r1); st != Assigned {
		t.Errorf("status %d", st)
	}
	if _, d2 := ride(t, s, "ravi", Location{1, 0}, Location{9, 9}, Mini); d2 != "d2" {
		t.Errorf("second ride driver %s", d2)
	}
	_, err := s.RequestRide("kim", Location{}, Location{1, 1}, Mini)
	is(t, err, ErrNoDriver, "no minis left")
	_, err = s.RequestRide("kim", Location{}, Location{1, 1}, SUV)
	is(t, err, ErrNoDriver, "no SUVs")
	_, err = s.RequestRide("asha", Location{}, Location{1, 1}, Sedan)
	is(t, err, ErrRiderBusy, "rider busy")
}

func TestOfflineDriversNotMatched(t *testing.T) {
	defer guard(t)
	s, _ := svc(t)
	s.SetOnline("d1", false)
	if _, d := ride(t, s, "asha", Location{}, Location{3, 3}, Mini); d != "d2" {
		t.Errorf("offline skipped: %s", d)
	}
	is(t, s.SetOnline("d2", false), ErrDriverBusy, "offline mid-trip")
}

func TestLifecycleAndFare(t *testing.T) {
	defer guard(t)
	s, c := svc(t)
	r, _ := ride(t, s, "asha", Location{0, 1}, Location{4, 4}, Mini)
	is(t, s.Start(r), ErrIllegalTransition, "must arrive first")
	if err := s.Arrived(r); err != nil {
		t.Fatal(err)
	}
	c.t = time.UnixMilli(1000)
	s.Start(r)
	c.t = c.t.Add(9*time.Minute + 30*time.Second)
	if fare, err := s.End(r); err != nil || fare != 50+10*7+2*10 {
		t.Errorf("fare: %d %v", fare, err)
	}
	if st, _ := s.Status(r); st != Completed {
		t.Errorf("status %d", st)
	}
	is(t, s.Cancel(r), ErrIllegalTransition, "cancel completed")
	if _, d := ride(t, s, "ravi", Location{4, 4}, Location{}, Mini); d != "d1" {
		t.Errorf("d1 should be free at (4,4): %s", d)
	}
}

func TestCancellationFreesDriverAndRider(t *testing.T) {
	defer guard(t)
	s, _ := svc(t)
	r, _ := ride(t, s, "asha", Location{}, Location{2, 2}, Sedan)
	s.Arrived(r)
	if err := s.Cancel(r); err != nil {
		t.Fatal(err)
	}
	r2, d := ride(t, s, "asha", Location{}, Location{2, 2}, Sedan)
	if d != "d3" {
		t.Errorf("driver free again: %s", d)
	}
	s.Arrived(r2)
	s.Start(r2)
	is(t, s.Cancel(r2), ErrIllegalTransition, "cancel in progress")
}

func TestNoDoubleAssignmentUnderContention(t *testing.T) {
	defer guard(t)
	s := NewRideService(NewNearestDriver(100), NewStandardPricing(rates, 100), time.Now)
	for i := 0; i < 10; i++ {
		id := fmt.Sprint("d", i)
		s.AddDriver(id, Mini, Location{i, 0}, 4.0)
		s.SetOnline(id, true)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var drivers []string
	start := make(chan struct{})
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			id, err := s.RequestRide(fmt.Sprint("u", i), Location{}, Location{5, 5}, Mini)
			if err != nil {
				return
			}
			d, _ := s.DriverOf(id)
			mu.Lock()
			drivers = append(drivers, d)
			mu.Unlock()
		}(i)
	}
	close(start)
	wg.Wait()
	uniq := map[string]bool{}
	for _, d := range drivers {
		uniq[d] = true
	}
	if len(drivers) != 10 || len(uniq) != 10 {
		t.Errorf("expected 10 trips with 10 distinct drivers, got %d trips / %d drivers", len(drivers), len(uniq))
	}
}
