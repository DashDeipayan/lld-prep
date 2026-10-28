package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"
)

func TestBasicGetPut(t *testing.T) {
	defer guard(t)
	c := NewLRU[string, int](2, nil)
	c.Put("a", 1)
	c.Put("b", 2)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Errorf("hit: %v %v", v, ok)
	}
	if _, ok := c.Get("z"); ok {
		t.Error("miss expected")
	}
	if c.Len() != 2 {
		t.Errorf("len %d", c.Len())
	}
}

func TestEvictsLeastRecentlyUsed(t *testing.T) {
	defer guard(t)
	var ev []string
	c := NewLRU[string, int](2, func(k string, v int) { ev = append(ev, fmt.Sprintf("%s=%d", k, v)) })
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a")
	c.Put("c", 3)
	if !reflect.DeepEqual(ev, []string{"b=2"}) {
		t.Errorf("evictions: %v (b was LRU because a was read)", ev)
	}
	if _, ok := c.Get("b"); ok {
		t.Error("b should be gone")
	}
	if got := c.Keys(); !reflect.DeepEqual(got, []string{"c", "a"}) {
		t.Errorf("keys MRU first: %v", got)
	}
}

func TestUpdateRefreshesAndDoesNotEvict(t *testing.T) {
	defer guard(t)
	var ev []string
	c := NewLRU[string, int](2, func(k string, _ int) { ev = append(ev, k) })
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("a", 10)
	if len(ev) != 0 {
		t.Errorf("update is not an eviction: %v", ev)
	}
	if got := c.Keys(); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("update moves to front: %v", got)
	}
	c.Put("c", 3)
	if !reflect.DeepEqual(ev, []string{"b"}) {
		t.Errorf("b evicted next: %v", ev)
	}
	if v, _ := c.Get("a"); v != 10 {
		t.Errorf("updated value: %d", v)
	}
}

func TestCapacityOne(t *testing.T) {
	defer guard(t)
	c := NewLRU[int, int](1, nil)
	for i := 0; i < 5; i++ {
		c.Put(i, i)
	}
	if got := c.Keys(); !reflect.DeepEqual(got, []int{4}) {
		t.Errorf("only latest survives: %v", got)
	}
}

func TestConcurrentAccessKeepsInvariants(t *testing.T) {
	defer guard(t)
	c := NewLRU[int, int](64, nil)
	var wg sync.WaitGroup
	errs := make(chan string, 8)
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(seed int64) {
			defer wg.Done()
			r := rand.New(rand.NewSource(seed))
			for i := 0; i < 20000; i++ {
				k := r.Intn(200)
				if r.Intn(2) == 0 {
					c.Put(k, k)
				} else if v, ok := c.Get(k); ok && v != k {
					errs <- fmt.Sprintf("wrong value for %d: %d", k, v)
					return
				}
			}
		}(int64(g))
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Fatal(e)
	}
	keys := c.Keys()
	if c.Len() > 64 {
		t.Errorf("len %d > capacity", c.Len())
	}
	if len(keys) != c.Len() {
		t.Errorf("list (%d) and map (%d) disagree", len(keys), c.Len())
	}
	seen := map[int]bool{}
	for _, k := range keys {
		if seen[k] {
			t.Fatalf("duplicate node for key %d", k)
		}
		seen[k] = true
	}
}
