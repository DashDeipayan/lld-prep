package main

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type flaky struct {
	calls, failures int
	delivered       []string
}

func (f *flaky) Send(to, msg string) error {
	f.calls++
	if f.calls <= f.failures {
		return fmt.Errorf("fail #%d", f.calls)
	}
	f.delivered = append(f.delivered, to+":"+msg)
	return nil
}

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) set(ms int64)            { c.t = time.UnixMilli(ms) }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func list(t *testing.T, want, got []string, msg string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func TestRetrySucceedsWithinAttempts(t *testing.T) {
	defer guard(t)
	f := &flaky{failures: 2}
	if err := WithRetry(f, 3).Send("a", "hi"); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if f.calls != 3 {
		t.Errorf("expected 3 calls, got %d", f.calls)
	}
	list(t, []string{"a:hi"}, f.delivered, "delivered once")
}

func TestRetryGivesUpWithLastError(t *testing.T) {
	defer guard(t)
	f := &flaky{failures: 5}
	err := WithRetry(f, 3).Send("a", "hi")
	if err == nil || err.Error() != "fail #3" {
		t.Errorf("expected last error 'fail #3', got %v", err)
	}
	if f.calls != 3 {
		t.Errorf("expected exactly 3 calls, got %d", f.calls)
	}
}

func TestRateLimitPerRecipientSlidingWindow(t *testing.T) {
	defer guard(t)
	c := &fakeClock{}
	c.set(1000)
	f := &flaky{}
	s := WithRateLimit(f, c.now, 2, time.Second)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
	}
	limited := func(err error, msg string) {
		t.Helper()
		if !errors.Is(err, ErrRateLimited) {
			t.Errorf("%s: expected ErrRateLimited, got %v", msg, err)
		}
	}
	must(s.Send("a", "1"))
	c.advance(400 * time.Millisecond)
	must(s.Send("a", "2"))
	limited(s.Send("a", "3"), "third within window")
	must(s.Send("b", "1"))
	c.set(2000)
	must(s.Send("a", "4"))
	limited(s.Send("a", "5"), "send at 1400 still in window at 2000")
	c.set(2400)
	must(s.Send("a", "6"))
	list(t, []string{"a:1", "a:2", "b:1", "a:4", "a:6"}, f.delivered, "deliveries")
	if f.calls != 5 {
		t.Errorf("rejected sends must not reach inner: %d calls", f.calls)
	}
}

func TestFailedSendsDoNotConsumeQuota(t *testing.T) {
	defer guard(t)
	c := &fakeClock{}
	f := &flaky{failures: 1}
	s := WithRateLimit(f, c.now, 1, time.Second)
	if err := s.Send("a", "x"); err == nil || errors.Is(err, ErrRateLimited) {
		t.Errorf("inner failure should pass through, got %v", err)
	}
	if err := s.Send("a", "y"); err != nil {
		t.Errorf("quota should still be available: %v", err)
	}
}

func TestLoggingPassesThrough(t *testing.T) {
	defer guard(t)
	var log []string
	f := &flaky{failures: 1}
	s := WithLogging(f, &log)
	if s.Send("a", "x") == nil {
		t.Error("error must be returned")
	}
	if err := s.Send("a", "y"); err != nil {
		t.Errorf("unexpected %v", err)
	}
	list(t, []string{"FAIL a: x", "OK a: y"}, log, "log lines")
}

func TestCompositionRetryInsideRateLimit(t *testing.T) {
	defer guard(t)
	c := &fakeClock{}
	f := &flaky{failures: 2}
	var log []string
	s := WithLogging(WithRateLimit(WithRetry(f, 3), c.now, 1, time.Second), &log)
	if err := s.Send("a", "m1"); err != nil {
		t.Fatalf("m1: %v", err)
	}
	if err := s.Send("a", "m2"); !errors.Is(err, ErrRateLimited) {
		t.Errorf("m2: expected ErrRateLimited, got %v", err)
	}
	list(t, []string{"OK a: m1", "FAIL a: m2"}, log, "logging sees the decorated result")
	if f.calls != 3 {
		t.Errorf("inner called only for m1: %d", f.calls)
	}
}
