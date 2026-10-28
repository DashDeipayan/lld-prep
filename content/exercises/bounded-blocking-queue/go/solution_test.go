package main

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestFifoAndSize(t *testing.T) {
	defer guard(t)
	q := NewBoundedQueue[string](3)
	for _, s := range []string{"a", "b", "c"} {
		if err := q.Put(s); err != nil {
			t.Fatal(err)
		}
	}
	if q.Len() != 3 {
		t.Errorf("len: %d", q.Len())
	}
	if x, _ := q.Take(); x != "a" {
		t.Errorf("fifo 1: %q", x)
	}
	if x, _ := q.Take(); x != "b" {
		t.Errorf("fifo 2: %q", x)
	}
	if q.Len() != 1 {
		t.Errorf("len after takes: %d", q.Len())
	}
}

func TestTimeoutsWhenFullOrEmpty(t *testing.T) {
	defer guard(t)
	q := NewBoundedQueue[int](1)
	if _, ok := q.Poll(50 * time.Millisecond); ok {
		t.Error("poll on empty should time out")
	}
	if ok, err := q.Offer(1, 50*time.Millisecond); !ok || err != nil {
		t.Errorf("offer with room: %v %v", ok, err)
	}
	start := time.Now()
	if ok, _ := q.Offer(2, 80*time.Millisecond); ok {
		t.Error("offer on full should time out")
	}
	if time.Since(start) < 60*time.Millisecond {
		t.Error("offer should actually wait")
	}
}

func TestPutBlocksUntilTake(t *testing.T) {
	defer guard(t)
	q := NewBoundedQueue[int](1)
	q.Put(1)
	done := make(chan struct{})
	go func() { q.Put(2); close(done) }()
	select {
	case <-done:
		t.Fatal("put must block while full")
	case <-time.After(80 * time.Millisecond):
	}
	if x, _ := q.Take(); x != 1 {
		t.Errorf("take: %d", x)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("blocked put should complete after take")
	}
	if x, _ := q.Take(); x != 2 {
		t.Errorf("second item: %d", x)
	}
}

func TestProducerConsumerNoLossNoDuplication(t *testing.T) {
	defer guard(t)
	q := NewBoundedQueue[int](8)
	const producers, per = 4, 2000
	var pw, cw sync.WaitGroup
	var mu sync.Mutex
	seen := map[int]int{}
	for p := 0; p < producers; p++ {
		pw.Add(1)
		go func(base int) {
			defer pw.Done()
			for i := 0; i < per; i++ {
				q.Put(base + i)
			}
		}(p * per)
	}
	for c := 0; c < 4; c++ {
		cw.Add(1)
		go func() {
			defer cw.Done()
			for {
				x, ok := q.Take()
				if !ok {
					return
				}
				mu.Lock()
				seen[x]++
				mu.Unlock()
			}
		}()
	}
	pw.Wait()
	q.Close()
	cw.Wait()
	if len(seen) != producers*per {
		t.Errorf("expected %d distinct items, got %d", producers*per, len(seen))
	}
	for k, n := range seen {
		if n != 1 {
			t.Fatalf("item %d consumed %d times", k, n)
		}
	}
}

func TestCloseWakesBlockedCallers(t *testing.T) {
	defer guard(t)
	empty := NewBoundedQueue[int](1)
	takeRes := make(chan bool, 1)
	go func() { _, ok := empty.Take(); takeRes <- ok }()

	full := NewBoundedQueue[int](1)
	full.Put(1)
	putRes := make(chan error, 1)
	go func() { putRes <- full.Put(2) }()

	time.Sleep(80 * time.Millisecond)
	empty.Close()
	full.Close()
	select {
	case ok := <-takeRes:
		if ok {
			t.Error("blocked take should return ok=false after close")
		}
	case <-time.After(time.Second):
		t.Fatal("close must wake a blocked Take")
	}
	select {
	case err := <-putRes:
		if !errors.Is(err, ErrClosed) {
			t.Errorf("blocked put should fail with ErrClosed, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("close must wake a blocked Put")
	}
}

func TestDrainAfterClose(t *testing.T) {
	defer guard(t)
	q := NewBoundedQueue[int](3)
	q.Put(1)
	q.Put(2)
	q.Close()
	q.Close()
	if err := q.Put(3); !errors.Is(err, ErrClosed) {
		t.Errorf("put after close: %v", err)
	}
	if x, ok := q.Take(); !ok || x != 1 {
		t.Errorf("remaining items still available: %d %v", x, ok)
	}
	if x, ok := q.Poll(10 * time.Millisecond); !ok || x != 2 {
		t.Errorf("poll drains too: %d %v", x, ok)
	}
	done := make(chan bool, 1)
	go func() { _, ok := q.Take(); done <- ok }()
	select {
	case ok := <-done:
		if ok {
			t.Error("drained closed queue should return ok=false")
		}
	case <-time.After(time.Second):
		t.Fatal("take on drained closed queue must not block")
	}
}
