package main

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var rates = map[VehicleType]int64{Bike: 10, Car: 20, Truck: 50}

type tclock struct{ t time.Time }

func (c *tclock) now() time.Time { return c.t }

func newLot(t *testing.T) (*ParkingLot, *tclock) {
	t.Helper()
	c := &tclock{t: time.Unix(0, 0)}
	l := NewParkingLot(NewHourlyPricing(rates), NearestSelector{}, c.now)
	if err := l.AddFloor(1, 2, 3, 1); err != nil {
		t.Fatal(err)
	}
	if err := l.AddFloor(2, 1, 1, 1); err != nil {
		t.Fatal(err)
	}
	return l, c
}

func park(t *testing.T, l *ParkingLot, plate string, vt VehicleType) *Ticket {
	t.Helper()
	tk, err := l.Park(Vehicle{plate, vt})
	if err != nil {
		t.Fatalf("park %s: %v", plate, err)
	}
	return tk
}

func want[T comparable](t *testing.T, exp, got T, msg string) {
	t.Helper()
	if exp != got {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, exp, got)
	}
}

// ---------------- step 1 ----------------
func TestS1_FittingRules(t *testing.T) {
	defer guard(t)
	small, medium, large := NewSpot("x", 1, Small), NewSpot("y", 1, Medium), NewSpot("z", 1, Large)
	b, c, tr := Vehicle{"b", Bike}, Vehicle{"c", Car}, Vehicle{"t", Truck}
	if !(small.CanFit(b) && medium.CanFit(b) && large.CanFit(b)) {
		t.Error("bike fits anywhere")
	}
	if small.CanFit(c) || !medium.CanFit(c) || !large.CanFit(c) {
		t.Error("car needs medium+")
	}
	if small.CanFit(tr) || medium.CanFit(tr) || !large.CanFit(tr) {
		t.Error("truck needs large")
	}
	want(t, "y", medium.ID(), "id")
	want(t, Medium, medium.Size(), "size")
}

func TestS1_OccupyAndVacate(t *testing.T) {
	defer guard(t)
	s := NewSpot("F1-M1", 1, Medium)
	want(t, true, s.IsFree(), "new spot is free")
	if err := s.Occupy(Vehicle{"KA1", Car}); err != nil {
		t.Fatal(err)
	}
	want(t, false, s.IsFree(), "occupied")
	if err := s.Occupy(Vehicle{"KA2", Bike}); !errors.Is(err, ErrSpotUnavailable) {
		t.Errorf("double occupy: %v", err)
	}
	v, ok := s.Vacate()
	want(t, true, ok && v.Plate == "KA1", "vacate returns the vehicle")
	want(t, true, s.IsFree(), "free again")
	if err := NewSpot("z", 1, Small).Occupy(Vehicle{"T", Truck}); !errors.Is(err, ErrSpotUnavailable) {
		t.Errorf("non-fitting vehicle: %v", err)
	}
}

// ---------------- step 2 ----------------
func TestS2_FloorsAndAvailability(t *testing.T) {
	defer guard(t)
	l, _ := newLot(t)
	want(t, 3, l.Available(Small), "small across floors")
	want(t, 4, l.Available(Medium), "medium across floors")
	want(t, 2, l.Available(Large), "large across floors")
	if err := l.AddFloor(1, 1, 1, 1); !errors.Is(err, ErrDuplicateFloor) {
		t.Errorf("duplicate floor: %v", err)
	}
}

// ---------------- step 3 ----------------
func TestS3_ParkIssuesTicketsNearestFirst(t *testing.T) {
	defer guard(t)
	l, c := newLot(t)
	c.t = time.UnixMilli(5000)
	t1 := park(t, l, "B1", Bike)
	want(t, "T1", t1.ID, "ticket id")
	want(t, "F1-S1", t1.SpotID, "bike takes the first small spot")
	want(t, true, t1.Entry.Equal(time.UnixMilli(5000)), "entry time from clock")
	want(t, "F1-M1", park(t, l, "C1", Car).SpotID, "car takes first medium")
	want(t, "F1-L1", park(t, l, "TR1", Truck).SpotID, "truck takes first large")
	want(t, "F2-L1", park(t, l, "TR2", Truck).SpotID, "second truck to floor 2")
	want(t, 3, l.Available(Medium), "availability updated")
}

func TestS3_FullAndDuplicate(t *testing.T) {
	defer guard(t)
	l, _ := newLot(t)
	park(t, l, "TR1", Truck)
	park(t, l, "TR2", Truck)
	if _, err := l.Park(Vehicle{"TR3", Truck}); !errors.Is(err, ErrLotFull) {
		t.Errorf("no large spot left: %v", err)
	}
	if _, err := l.Park(Vehicle{"TR1", Truck}); !errors.Is(err, ErrAlreadyParked) {
		t.Errorf("plate already parked: %v", err)
	}
}

func TestS3_CustomSelectorStrategy(t *testing.T) {
	defer guard(t)
	topFloorFirst := SelectorFunc(func(c []*Spot, _ Vehicle) *Spot {
		var best *Spot
		for _, s := range c {
			if best == nil || s.Floor() > best.Floor() {
				best = s
			}
		}
		return best
	})
	l := NewParkingLot(NewHourlyPricing(rates), topFloorFirst, time.Now)
	l.AddFloor(1, 1, 1, 1)
	l.AddFloor(2, 1, 1, 1)
	id := park(t, l, "C1", Car).SpotID
	floor, _ := strconv.Atoi(id[1:strings.Index(id, "-")])
	want(t, 2, floor, "custom strategy used without changing ParkingLot")
}

// ---------------- step 4 ----------------
func TestS4_HourlyPricing(t *testing.T) {
	defer guard(t)
	p := NewHourlyPricing(rates)
	want(t, int64(20), p.Fee(Car, 0), "minimum one hour")
	want(t, int64(20), p.Fee(Car, 59*time.Minute), "59 min")
	want(t, int64(20), p.Fee(Car, 60*time.Minute), "exactly 60 min")
	want(t, int64(40), p.Fee(Car, 61*time.Minute), "61 min -> 2 hours")
	want(t, int64(150), p.Fee(Truck, 3*time.Hour), "3h truck")
}

func TestS4_UnparkReturnsReceipt(t *testing.T) {
	defer guard(t)
	l, c := newLot(t)
	tk := park(t, l, "C1", Car)
	c.t = c.t.Add(90 * time.Minute)
	r, err := l.Unpark(tk.ID)
	if err != nil {
		t.Fatal(err)
	}
	want(t, Receipt{"T1", "C1", "F1-M1", 90, 40}, *r, "receipt")
	want(t, 4, l.Available(Medium), "spot freed")
	if _, err := l.Unpark(tk.ID); !errors.Is(err, ErrUnknownTicket) {
		t.Errorf("ticket reuse: %v", err)
	}
	if _, err := l.Unpark("T999"); !errors.Is(err, ErrUnknownTicket) {
		t.Errorf("unknown ticket: %v", err)
	}
	want(t, "F1-M1", park(t, l, "C1", Car).SpotID, "same plate can park again")
}

// ---------------- step 5 ----------------
func TestS5_EventsInOrder(t *testing.T) {
	defer guard(t)
	l, _ := newLot(t)
	var log []string
	l.AddListener(ListenerFunc(func(e Event) { log = append(log, fmt.Sprintf("a:%s:%s", e.Kind, e.SpotID)) }))
	l.AddListener(ListenerFunc(func(e Event) { log = append(log, "b:"+e.Kind) }))
	tk := park(t, l, "C1", Car)
	l.Park(Vehicle{"C1", Car})
	l.Unpark(tk.ID)
	exp := []string{"a:PARKED:F1-M1", "b:PARKED", "a:LEFT:F1-M1", "b:LEFT"}
	if !reflect.DeepEqual(exp, log) {
		t.Errorf("events (failed park emits nothing)\nexpected: %v\n  actual: %v", exp, log)
	}
}

func TestS5_DisplayBoardTracksAvailability(t *testing.T) {
	defer guard(t)
	l, _ := newLot(t)
	pre := park(t, l, "B0", Bike)
	board := NewDisplayBoard(l)
	want(t, 2, board.Free(Small), "board starts from current availability")
	c := park(t, l, "C1", Car)
	park(t, l, "TR1", Truck)
	l.Unpark(pre.ID)
	l.Unpark(c.ID)
	for _, s := range []SpotSize{Small, Medium, Large} {
		want(t, l.Available(s), board.Free(s), fmt.Sprint("board matches lot for size ", s))
	}
}

// ---------------- step 6 ----------------
func TestS6_ConcurrentParkingNeverDoubleAssigns(t *testing.T) {
	defer guard(t)
	l := NewParkingLot(NewHourlyPricing(rates), NearestSelector{}, time.Now)
	l.AddFloor(1, 0, 10, 0)
	l.AddFloor(2, 0, 10, 0)
	board := NewDisplayBoard(l)
	var wg sync.WaitGroup
	var mu sync.Mutex
	spots := map[string]int{}
	ok := 0
	start := make(chan struct{})
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			tk, err := l.Park(Vehicle{fmt.Sprint("C", i), Car})
			if err != nil {
				return
			}
			mu.Lock()
			ok++
			spots[tk.SpotID]++
			mu.Unlock()
		}(i)
	}
	close(start)
	wg.Wait()
	want(t, 20, ok, "exactly 20 parks succeed")
	want(t, 20, len(spots), "20 distinct spots")
	want(t, 0, l.Available(Medium), "lot full")
	want(t, 0, board.Free(Medium), "board consistent under concurrency")
}
