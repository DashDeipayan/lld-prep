package main

import (
	"regexp"
	"runtime"
	"sync"
	"testing"
)

// Runs first (source order), before anything has created the instance.
func TestAConcurrentFirstAccessReturnsOneInstance(t *testing.T) {
	defer guard(t)
	const n = 32
	var wg sync.WaitGroup
	start := make(chan struct{})
	got := make([]*IDGenerator, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			got[i] = GetIDGenerator()
		}(i)
	}
	close(start)
	wg.Wait()
	for i := 1; i < n; i++ {
		if got[i] != got[0] {
			t.Fatal("all goroutines must get the same instance")
		}
	}
}

func TestBSequentialIdsIncrease(t *testing.T) {
	defer guard(t)
	g := GetIDGenerator()
	a, b, c := g.NextID(), g.NextID(), g.NextID()
	if b != a+1 || c != b+1 {
		t.Errorf("ids should be consecutive: %d %d %d", a, b, c)
	}
}

func TestCConcurrentIdsAreUnique(t *testing.T) {
	defer guard(t)
	if runtime.GOMAXPROCS(0) < 2 {
		runtime.GOMAXPROCS(2)
	}
	g := GetIDGenerator()
	const workers, per = 16, 5000
	var mu sync.Mutex
	seen := make(map[int64]bool, workers*per)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			local := make([]int64, 0, per)
			for i := 0; i < per; i++ {
				local = append(local, g.NextID())
				if i%64 == 0 {
					runtime.Gosched()
				}
			}
			mu.Lock()
			for _, id := range local {
				seen[id] = true
			}
			mu.Unlock()
		}()
	}
	close(start)
	wg.Wait()
	if len(seen) != workers*per {
		t.Errorf("expected %d unique ids, got %d (lost updates => duplicates)", workers*per, len(seen))
	}
}

func TestDPrefixFormat(t *testing.T) {
	defer guard(t)
	id := GetIDGenerator().NextIDWithPrefix("ORD")
	if !regexp.MustCompile(`^ORD-\d{6,}$`).MatchString(id) {
		t.Errorf("expected ORD- followed by 6+ digits, got %q", id)
	}
}
