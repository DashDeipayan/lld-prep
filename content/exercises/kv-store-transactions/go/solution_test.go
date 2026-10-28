package main

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func get(t *testing.T, s *KVStore, k, want string, present bool, msg string) {
	t.Helper()
	v, ok := s.Get(k)
	if ok != present || (present && v != want) {
		t.Errorf("%s: get(%s) = %q,%v; expected %q,%v", msg, k, v, ok, want, present)
	}
}

func count(t *testing.T, s *KVStore, v string, want int, msg string) {
	t.Helper()
	if got := s.Count(v); got != want {
		t.Errorf("%s: count(%s) = %d, expected %d", msg, v, got, want)
	}
}

func TestBasicOpsAndCount(t *testing.T) {
	defer guard(t)
	s := NewKVStore()
	s.Set("a", "10")
	s.Set("b", "10")
	get(t, s, "a", "10", true, "get")
	count(t, s, "10", 2, "two keys")
	s.Set("b", "20")
	count(t, s, "10", 1, "overwrite")
	count(t, s, "20", 1, "new value")
	s.Delete("a")
	s.Delete("nope")
	get(t, s, "a", "", false, "deleted")
	count(t, s, "10", 0, "after delete")
}

func TestRollbackRestoresInnermostBlock(t *testing.T) {
	defer guard(t)
	s := NewKVStore()
	s.Set("a", "10")
	s.Begin()
	s.Set("a", "20")
	s.Set("c", "30")
	get(t, s, "a", "20", true, "uncommitted visible")
	if err := s.Rollback(); err != nil {
		t.Fatal(err)
	}
	get(t, s, "a", "10", true, "restored")
	get(t, s, "c", "", false, "created key gone")
	count(t, s, "10", 1, "counts restored")
	count(t, s, "30", 0, "counts restored")
}

func TestNestedBlocks(t *testing.T) {
	defer guard(t)
	s := NewKVStore()
	s.Set("a", "10")
	s.Begin()
	s.Set("a", "20")
	s.Begin()
	s.Delete("a")
	get(t, s, "a", "", false, "inner delete")
	s.Rollback()
	get(t, s, "a", "20", true, "inner rollback")
	s.Rollback()
	get(t, s, "a", "10", true, "outer rollback")
}

func TestMultipleWritesSameKeyInBlock(t *testing.T) {
	defer guard(t)
	s := NewKVStore()
	s.Set("x", "1")
	s.Begin()
	s.Set("x", "2")
	s.Set("x", "3")
	s.Delete("x")
	s.Set("x", "4")
	s.Rollback()
	get(t, s, "x", "1", true, "value before the block")
	count(t, s, "1", 1, "count")
	count(t, s, "4", 0, "count")
}

func TestCommitClosesAllBlocks(t *testing.T) {
	defer guard(t)
	s := NewKVStore()
	s.Begin()
	s.Set("a", "1")
	s.Begin()
	s.Set("b", "1")
	if err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	count(t, s, "1", 2, "committed")
	if err := s.Rollback(); !errors.Is(err, ErrNoTransaction) {
		t.Errorf("rollback after commit: %v", err)
	}
	if err := s.Commit(); !errors.Is(err, ErrNoTransaction) {
		t.Errorf("commit with nothing open: %v", err)
	}
}

func TestRollbackCostProportionalToChanges(t *testing.T) {
	defer guard(t)
	s := NewKVStore()
	for i := 0; i < 200000; i++ {
		s.Set(fmt.Sprint("k", i), "v")
	}
	start := time.Now()
	for i := 0; i < 2000; i++ {
		s.Begin()
		s.Set("k1", "w")
		s.Rollback()
	}
	if el := time.Since(start); el > 1500*time.Millisecond {
		t.Errorf("2000 begin/set/rollback cycles took %v; don't copy the whole map per transaction", el)
	}
	count(t, s, "v", 200000, "all restored")
}
