package main

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func mustPoll(t *testing.T, b *Broker, topic, g string, max int) []Message {
	t.Helper()
	m, err := b.Poll(topic, g, max)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func msgs(t *testing.T, want, got []Message, msg string) {
	t.Helper()
	if len(want) == 0 && len(got) == 0 {
		return
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func TestPublishAssignsSequentialOffsets(t *testing.T) {
	defer guard(t)
	b := NewBroker()
	b.CreateTopic("t")
	if o, _ := b.Publish("t", "a"); o != 0 {
		t.Errorf("first offset %d", o)
	}
	if o, _ := b.Publish("t", "b"); o != 1 {
		t.Errorf("second offset %d", o)
	}
	if err := b.CreateTopic("t"); !errors.Is(err, ErrTopicExists) {
		t.Errorf("duplicate: %v", err)
	}
	if _, err := b.Publish("nope", "x"); !errors.Is(err, ErrUnknownTopic) {
		t.Errorf("unknown: %v", err)
	}
}

func TestFromBeginningVsLatest(t *testing.T) {
	defer guard(t)
	b := NewBroker()
	b.CreateTopic("t")
	b.Publish("t", "old")
	b.Subscribe("t", "g1", true)
	b.Subscribe("t", "g2", false)
	b.Publish("t", "new")
	msgs(t, []Message{{0, "old"}, {1, "new"}}, mustPoll(t, b, "t", "g1", 10), "from beginning")
	msgs(t, []Message{{1, "new"}}, mustPoll(t, b, "t", "g2", 10), "latest only")
	msgs(t, nil, mustPoll(t, b, "t", "g1", 10), "nothing new")
}

func TestPollBatchesAndLag(t *testing.T) {
	defer guard(t)
	b := NewBroker()
	b.CreateTopic("t")
	b.Subscribe("t", "g", true)
	for i := 0; i < 5; i++ {
		b.Publish("t", fmt.Sprint("m", i))
	}
	lag := func(want int64) {
		t.Helper()
		if l, _ := b.Lag("t", "g"); l != want {
			t.Errorf("lag: expected %d, got %d", want, l)
		}
	}
	lag(5)
	msgs(t, []Message{{0, "m0"}, {1, "m1"}}, mustPoll(t, b, "t", "g", 2), "batch of 2")
	lag(3)
	if n := len(mustPoll(t, b, "t", "g", 100)); n != 3 {
		t.Errorf("rest: %d", n)
	}
	lag(0)
}

func TestResubscribeKeepsOffsetAndSeek(t *testing.T) {
	defer guard(t)
	b := NewBroker()
	b.CreateTopic("t")
	b.Subscribe("t", "g", true)
	b.Publish("t", "a")
	b.Publish("t", "b")
	mustPoll(t, b, "t", "g", 1)
	b.Subscribe("t", "g", true)
	msgs(t, []Message{{1, "b"}}, mustPoll(t, b, "t", "g", 10), "re-subscribe is a no-op")
	if err := b.Seek("t", "g", 0); err != nil {
		t.Fatal(err)
	}
	if n := len(mustPoll(t, b, "t", "g", 10)); n != 2 {
		t.Errorf("replay: %d", n)
	}
	if err := b.Seek("t", "g", 3); !errors.Is(err, ErrInvalidOffset) {
		t.Errorf("past end: %v", err)
	}
	if _, err := b.Poll("t", "unknown", 1); !errors.Is(err, ErrUnknownGroup) {
		t.Errorf("unknown group: %v", err)
	}
}

func TestGroupsShareWorkExactlyOnceUnderConcurrency(t *testing.T) {
	defer guard(t)
	b := NewBroker()
	b.CreateTopic("events")
	b.Subscribe("events", "billing", true)
	b.Subscribe("events", "audit", true)
	var pw, cw sync.WaitGroup
	start := make(chan struct{})
	for p := 0; p < 4; p++ {
		pw.Add(1)
		go func(id int) {
			defer pw.Done()
			<-start
			for i := 0; i < 500; i++ {
				b.Publish("events", fmt.Sprintf("%d:%d", id, i))
			}
		}(p)
	}
	var mu sync.Mutex
	seen := map[int64]int{}
	for c := 0; c < 4; c++ {
		cw.Add(1)
		go func() {
			defer cw.Done()
			<-start
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				mu.Lock()
				done := len(seen) >= 2000
				mu.Unlock()
				if done {
					return
				}
				ms, _ := b.Poll("events", "billing", 10)
				mu.Lock()
				for _, m := range ms {
					seen[m.Offset]++
				}
				mu.Unlock()
			}
		}()
	}
	close(start)
	pw.Wait()
	cw.Wait()
	if len(seen) != 2000 {
		t.Errorf("group consumed %d of 2000", len(seen))
	}
	for off, n := range seen {
		if n != 1 {
			t.Fatalf("offset %d delivered %d times within one group", off, n)
		}
	}
	if n := len(mustPoll(t, b, "events", "audit", 5000)); n != 2000 {
		t.Errorf("audit group got %d", n)
	}
}
