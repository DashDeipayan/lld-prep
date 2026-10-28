package main

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

func hotel(t *testing.T) *Hotel {
	t.Helper()
	h := NewHotel()
	for _, err := range []error{
		h.AddRoom("102", Double, 4000), h.AddRoom("101", Double, 4000),
		h.AddRoom("103", Single, 2500), h.AddRoom("301", Suite, 9000),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	return h
}

func book(t *testing.T, h *Hotel, g string, rt RoomType, in, out int) (string, string) {
	t.Helper()
	id, err := h.Book(g, rt, in, out)
	if err != nil {
		t.Fatalf("book: %v", err)
	}
	r, _ := h.RoomOf(id)
	return id, r
}

func is(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected %v, got %v", msg, target, err)
	}
}

func avail(t *testing.T, h *Hotel, in, out int) map[RoomType]int {
	t.Helper()
	m, err := h.Availability(in, out)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestAvailabilityAndAssignment(t *testing.T) {
	defer guard(t)
	h := hotel(t)
	if got := avail(t, h, 1, 3); !reflect.DeepEqual(got, map[RoomType]int{Single: 1, Double: 2, Suite: 1}) {
		t.Errorf("all free: %v", got)
	}
	id, r := book(t, h, "a", Double, 1, 3)
	if id != "BK1" || r != "101" {
		t.Errorf("first: %s %s", id, r)
	}
	if _, r := book(t, h, "b", Double, 2, 4); r != "102" {
		t.Errorf("overlap gets next room: %s", r)
	}
	if got := avail(t, h, 2, 3); got[Double] != 0 {
		t.Errorf("doubles full: %v", got)
	}
	if _, r := book(t, h, "c", Double, 3, 5); r != "101" {
		t.Errorf("101 free from night 3: %s", r)
	}
	_, err := h.Book("d", Double, 2, 4)
	is(t, err, ErrNoRoom, "none left")
	_, err = h.Book("d", Single, 4, 4)
	is(t, err, ErrInvalid, "empty stay")
}

func TestLifecycleAndBill(t *testing.T) {
	defer guard(t)
	h := hotel(t)
	b, _ := book(t, h, "a", Suite, 5, 8)
	is(t, h.AddCharge(b, 100), ErrBadStatus, "charge before check-in")
	is(t, h.CheckIn(b, 4), ErrBadStatus, "wrong day")
	if err := h.CheckIn(b, 5); err != nil {
		t.Fatal(err)
	}
	h.AddCharge(b, 1200)
	h.AddCharge(b, 300)
	is(t, h.AddCharge(b, 0), ErrInvalid, "positive charge")
	if bill, err := h.CheckOut(b); err != nil || bill != 3*9000+1500 {
		t.Errorf("bill: %d %v", bill, err)
	}
	_, err := h.CheckOut(b)
	is(t, err, ErrBadStatus, "double checkout")
	if _, r := book(t, h, "b", Suite, 6, 7); r != "301" {
		t.Errorf("freed after checkout: %s", r)
	}
}

func TestCancel(t *testing.T) {
	defer guard(t)
	h := hotel(t)
	b, _ := book(t, h, "a", Single, 1, 2)
	if err := h.Cancel(b); err != nil {
		t.Fatal(err)
	}
	if st, _ := h.Status(b); st != "CANCELLED" {
		t.Errorf("status %s", st)
	}
	if avail(t, h, 1, 2)[Single] != 1 {
		t.Error("room free again")
	}
	is(t, h.Cancel(b), ErrBadStatus, "double cancel")
	c, _ := book(t, h, "b", Single, 1, 2)
	h.CheckIn(c, 1)
	is(t, h.Cancel(c), ErrBadStatus, "cancel after check-in")
}

func TestLastSuiteSingleWinner(t *testing.T) {
	defer guard(t)
	h := hotel(t)
	var ok atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := h.Book("g", Suite, 10, 12); err == nil {
				ok.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()
	if ok.Load() != 1 {
		t.Errorf("expected exactly 1 booking, got %d", ok.Load())
	}
}
