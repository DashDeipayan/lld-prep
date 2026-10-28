package main

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

func sched(t *testing.T) *Scheduler {
	t.Helper()
	s := NewScheduler()
	for _, r := range []struct {
		id  string
		cap int
	}{{"R1", 4}, {"R2", 8}, {"R0", 4}, {"R3", 2}} {
		if err := s.AddRoom(r.id, r.cap); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

// roomOf turns a (Booking, error) result into the room id, or the error text.
func roomOf(b Booking, err error) string {
	if err != nil {
		return "error: " + err.Error()
	}
	return b.Room
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

func TestBestFitRoomSelection(t *testing.T) {
	defer guard(t)
	s := sched(t)
	deep(t, "R0", roomOf(s.Book("a", 600, 660, 3)), "best fit, tie -> lowest id")
	deep(t, "R1", roomOf(s.Book("b", 600, 660, 3)), "next 4-seater")
	deep(t, "R2", roomOf(s.Book("c", 600, 660, 3)), "8-seater")
	_, err := s.Book("d", 600, 660, 3)
	is(t, err, ErrNoRoom, "none left")
	deep(t, "R3", roomOf(s.Book("e", 600, 660, 2)), "2-seater")
}

func TestHalfOpenIntervals(t *testing.T) {
	defer guard(t)
	s := sched(t)
	s.BookRoom("R1", "a", 600, 660)
	if b, err := s.BookRoom("R1", "b", 660, 720); err != nil || b.ID != "M2" {
		t.Errorf("back-to-back: %v %v", b, err)
	}
	for _, iv := range [][2]int{{659, 661}, {500, 800}, {610, 620}} {
		_, err := s.BookRoom("R1", "c", iv[0], iv[1])
		is(t, err, ErrConflict, fmt.Sprint("overlap ", iv))
	}
	if b, err := s.BookRoom("R1", "c", 540, 600); err != nil || b.ID != "M3" {
		t.Errorf("ends at next start: %v %v", b, err)
	}
}

func TestValidation(t *testing.T) {
	defer guard(t)
	s := sched(t)
	for _, c := range []struct{ s, e, n int }{{600, 600, 2}, {-1, 60, 2}, {1400, 1441, 2}, {0, 60, 0}} {
		_, err := s.Book("a", c.s, c.e, c.n)
		is(t, err, ErrInvalid, fmt.Sprint("invalid ", c))
	}
	_, err := s.BookRoom("RX", "a", 0, 60)
	is(t, err, ErrNotFound, "unknown room")
	is(t, s.AddRoom("R1", 3), ErrInvalid, "duplicate room")
}

func TestCancelAndSchedule(t *testing.T) {
	defer guard(t)
	s := sched(t)
	b1, _ := s.BookRoom("R2", "a", 700, 760)
	b2, _ := s.BookRoom("R2", "b", 600, 630)
	got, _ := s.Schedule("R2")
	deep(t, []Booking{b2, b1}, got, "sorted by start")
	if err := s.Cancel(b1.ID); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Schedule("R2")
	deep(t, []Booking{b2}, got, "cancelled removed")
	is(t, s.Cancel(b1.ID), ErrNotFound, "double cancel")
	deep(t, "R2", roomOf(s.BookRoom("R2", "c", 700, 760)), "slot reusable")
}

func TestFreeRoomsAndSlots(t *testing.T) {
	defer guard(t)
	s := sched(t)
	s.BookRoom("R0", "a", 600, 660)
	free, _ := s.FreeRooms(630, 690, 2)
	deep(t, []string{"R3", "R1", "R2"}, free, "ordered by capacity then id")
	s.BookRoom("R1", "a", 540, 570)
	s.BookRoom("R1", "b", 600, 630)
	s.BookRoom("R1", "c", 630, 700)
	slots, _ := s.FreeSlots("R1", 480, 720)
	deep(t, []Slot{{480, 540}, {570, 600}, {700, 720}}, slots, "gaps")
	slots, _ = s.FreeSlots("R1", 600, 700)
	deep(t, []Slot{}, slots, "fully booked")
	slots, _ = s.FreeSlots("R2", 0, 1440)
	deep(t, []Slot{{0, 1440}}, slots, "empty room")
}

func TestSingleWinnerUnderContention(t *testing.T) {
	defer guard(t)
	s := sched(t)
	var ok atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			if _, err := s.BookRoom("R1", fmt.Sprint("u", i), 600, 660); err == nil {
				ok.Add(1)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	deep(t, int32(1), ok.Load(), "exactly one winner")
}
