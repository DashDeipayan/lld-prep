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

type aclock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *aclock) now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *aclock) set(ms int64)   { c.mu.Lock(); c.t = time.UnixMilli(ms); c.mu.Unlock() }

func house() (*AuctionHouse, *aclock) {
	c := &aclock{t: time.UnixMilli(0)}
	return NewAuctionHouse(c.now), c
}

func create(t *testing.T, h *AuctionHouse, item string, endMs int64) string {
	t.Helper()
	id, err := h.Create("sam", item, 100, 10, time.UnixMilli(endMs))
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

func TestBidRules(t *testing.T) {
	defer guard(t)
	h, _ := house()
	a := create(t, h, "lamp", 1_000_000)
	is(t, h.PlaceBid(a, "asha", 99), ErrBidTooLow, "below start")
	if err := h.PlaceBid(a, "asha", 100); err != nil {
		t.Fatal(err)
	}
	is(t, h.PlaceBid(a, "asha", 200), ErrAlreadyLeading, "self outbid")
	is(t, h.PlaceBid(a, "ravi", 109), ErrBidTooLow, "increment")
	if err := h.PlaceBid(a, "ravi", 110); err != nil {
		t.Fatal(err)
	}
	if l, _ := h.Leader(a); l != "ravi" || h.Highest(a) != 110 {
		t.Errorf("leader %s highest %d", l, h.Highest(a))
	}
	is(t, h.PlaceBid(a, "sam", 500), ErrInvalid, "seller")
	_, err := h.Create("sam", "x", 100, 10, time.UnixMilli(0))
	is(t, err, ErrInvalid, "end in future")
}

func TestOutbidNotificationsInOrder(t *testing.T) {
	defer guard(t)
	h, _ := house()
	a := create(t, h, "lamp", 1_000_000)
	var events []string
	h.Subscribe(a, ListenerFunc(func(id, e string) { events = append(events, id+" "+e) }))
	h.PlaceBid(a, "asha", 100)
	h.PlaceBid(a, "ravi", 150)
	h.PlaceBid(a, "asha", 160)
	want := []string{"A1 BID:asha:100", "A1 BID:ravi:150", "A1 OUTBID:asha", "A1 BID:asha:160", "A1 OUTBID:ravi"}
	if !reflect.DeepEqual(want, events) {
		t.Errorf("events\nexpected: %v\n  actual: %v", want, events)
	}
}

func TestAntiSnipingExtension(t *testing.T) {
	defer guard(t)
	h, c := house()
	a := create(t, h, "lamp", 600_000)
	c.set(500_000)
	h.PlaceBid(a, "asha", 100)
	if got := h.EndTime(a).UnixMilli(); got != 600_000 {
		t.Errorf("no extension: %d", got)
	}
	c.set(570_000)
	h.PlaceBid(a, "ravi", 110)
	if got := h.EndTime(a).UnixMilli(); got != 630_000 {
		t.Errorf("extended: %d", got)
	}
	c.set(629_999)
	h.PlaceBid(a, "asha", 120)
	if got := h.EndTime(a).UnixMilli(); got != 689_999 {
		t.Errorf("extends again: %d", got)
	}
	c.set(689_999)
	is(t, h.PlaceBid(a, "ravi", 500), ErrClosed, "closed at end")
}

func TestCloseIsIdempotent(t *testing.T) {
	defer guard(t)
	h, c := house()
	a := create(t, h, "lamp", 1_000_000)
	b := create(t, h, "rug", 1_000_000)
	var events []string
	rec := ListenerFunc(func(_, e string) { events = append(events, e) })
	h.Subscribe(a, rec)
	h.Subscribe(b, rec)
	h.PlaceBid(a, "asha", 100)
	_, err := h.Close(a)
	is(t, err, ErrNotEnded, "early close")
	c.set(1_000_000)
	r, err := h.Close(a)
	if err != nil || r != (Result{"A1", "asha", 100}) {
		t.Errorf("winner: %+v %v", r, err)
	}
	if r2, _ := h.Close(a); r2 != r {
		t.Error("second close must return the same result")
	}
	if rb, _ := h.Close(b); rb != (Result{"A2", "", 0}) {
		t.Errorf("unsold: %+v", rb)
	}
	want := []string{"BID:asha:100", "CLOSED:asha:100", "CLOSED:UNSOLD"}
	if !reflect.DeepEqual(want, events) {
		t.Errorf("events\nexpected: %v\n  actual: %v", want, events)
	}
	is(t, h.PlaceBid(a, "ravi", 999), ErrClosed, "no bids after close")
}

func TestConcurrentBidsStayConsistent(t *testing.T) {
	defer guard(t)
	h, _ := house()
	a := create(t, h, "lamp", 10_000_000)
	var mu sync.Mutex
	var accepted []int64
	h.Subscribe(a, ListenerFunc(func(_, e string) {
		if strings.HasPrefix(e, "BID:") {
			v, _ := strconv.ParseInt(e[strings.LastIndex(e, ":")+1:], 10, 64)
			mu.Lock()
			accepted = append(accepted, v)
			mu.Unlock()
		}
	}))
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			h.PlaceBid(a, fmt.Sprint("u", i), 100+10*int64(i))
		}(i)
	}
	close(start)
	wg.Wait()
	for i := 1; i < len(accepted); i++ {
		if accepted[i] <= accepted[i-1] {
			t.Fatalf("accepted bids must strictly increase: %v", accepted)
		}
	}
	if l, _ := h.Leader(a); l != "u39" || h.Highest(a) != 490 {
		t.Errorf("leader %s highest %d", l, h.Highest(a))
	}
}
