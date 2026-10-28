package main

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

func svc(t *testing.T) *RentalService {
	t.Helper()
	r := NewRentalService()
	for _, err := range []error{
		r.AddBranch("B1", "Bengaluru"), r.AddBranch("B2", "Pune"),
		r.AddCar("B1", "C1", Sedan, 2000), r.AddCar("B1", "C2", Sedan, 1500),
		r.AddCar("B1", "C3", SUV, 3000), r.AddCar("B2", "C4", Sedan, 1000),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	return r
}

func search(t *testing.T, r *RentalService, city string, ct CarType, s, e int) []string {
	t.Helper()
	out, err := r.Search(city, ct, s, e)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func deep(t *testing.T, want, got any, msg string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func is(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected %v, got %v", msg, target, err)
	}
}

func TestSearchFiltersAndSorts(t *testing.T) {
	defer guard(t)
	r := svc(t)
	deep(t, []string{"C2", "C1"}, search(t, r, "bengaluru", Sedan, 1, 3), "sorted by rate")
	deep(t, []string{"C3"}, search(t, r, "Bengaluru", SUV, 1, 3), "SUV")
	deep(t, []string{}, search(t, r, "Delhi", Sedan, 1, 3), "no branch")
}

func TestReserveCostAndOverlap(t *testing.T) {
	defer guard(t)
	r := svc(t)
	a, err := r.Reserve("asha", "C2", 10, 13)
	if err != nil {
		t.Fatal(err)
	}
	deep(t, Reservation{"RS1", "asha", "C2", 10, 13, 4500, "RESERVED"}, a, "reservation")
	_, err = r.Reserve("ravi", "C2", 12, 15)
	is(t, err, ErrConflict, "overlap")
	b, err := r.Reserve("ravi", "C2", 13, 15)
	if err != nil || b.ID != "RS2" {
		t.Errorf("back-to-back: %v %v", b, err)
	}
	deep(t, []string{"C1"}, search(t, r, "Bengaluru", Sedan, 11, 12), "hidden from search")
	_, err = r.Reserve("x", "C2", 5, 5)
	is(t, err, ErrInvalid, "empty range")
	_, err = r.Reserve("x", "ZZ", 1, 2)
	is(t, err, ErrNotFound, "unknown car")
}

func TestCancel(t *testing.T) {
	defer guard(t)
	r := svc(t)
	a, _ := r.Reserve("asha", "C1", 1, 4)
	c, err := r.Cancel(a.ID)
	if err != nil || c.Status != "CANCELLED" {
		t.Fatalf("cancel: %v %v", c, err)
	}
	deep(t, []string{"C2", "C1"}, search(t, r, "Bengaluru", Sedan, 1, 4), "free again")
	_, err = r.Cancel(a.ID)
	is(t, err, ErrBadStatus, "double cancel")
}

func TestReturnWithLateFee(t *testing.T) {
	defer guard(t)
	r := svc(t)
	a, _ := r.Reserve("asha", "C1", 1, 4)
	if amt, _ := r.ReturnCar(a.ID, 6); amt != 6000+2*2000*3/2 {
		t.Errorf("late fee: %d", amt)
	}
	b, _ := r.Reserve("ravi", "C1", 6, 8)
	if amt, _ := r.ReturnCar(b.ID, 7); amt != 4000 {
		t.Errorf("early return: %d", amt)
	}
	_, err := r.ReturnCar(b.ID, 9)
	is(t, err, ErrBadStatus, "already completed")
	_, err = r.Cancel(b.ID)
	is(t, err, ErrBadStatus, "cancel completed")
}

func TestSingleWinnerUnderContention(t *testing.T) {
	defer guard(t)
	r := svc(t)
	var ok atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			s := 10 + i%3
			if _, err := r.Reserve("u", "C3", s, s+3); err == nil {
				ok.Add(1)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	deep(t, int32(1), ok.Load(), "exactly one wins")
}
