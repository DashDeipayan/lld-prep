package main

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func lines(t *testing.T, want, got []string, msg string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %q\n  actual: %q", msg, want, got)
	}
}

func TestFormatters(t *testing.T) {
	defer guard(t)
	if got := (SimpleFormatter{}).Format("orders", Warn, "payment slow"); got != "WARN [orders] payment slow" {
		t.Errorf("simple: %q", got)
	}
	want := `{"level":"INFO","logger":"app","msg":"say \"hi\" \\o/"}`
	if got := (JSONFormatter{}).Format("app", Info, `say "hi" \o/`); got != want {
		t.Errorf("json:\nexpected %s\n  actual %s", want, got)
	}
}

func TestLoggerAndHandlerLevels(t *testing.T) {
	defer guard(t)
	var console, alerts []string
	log := NewLogger("orders", Info).
		AddHandler(NewMemoryHandler(Debug, &console, SimpleFormatter{})).
		AddHandler(NewMemoryHandler(Error, &alerts, SimpleFormatter{}))
	log.Debug("d")
	log.Info("i")
	log.Warn("w")
	log.Error("e")
	lines(t, []string{"INFO [orders] i", "WARN [orders] w", "ERROR [orders] e"}, console, "logger drops DEBUG")
	lines(t, []string{"ERROR [orders] e"}, alerts, "handler keeps only ERROR")
}

func TestSetLevelAtRuntime(t *testing.T) {
	defer guard(t)
	var sink []string
	log := NewLogger("x", Error).AddHandler(NewMemoryHandler(Debug, &sink, SimpleFormatter{}))
	log.Info("before")
	log.SetLevel(Debug)
	log.Debug("after")
	lines(t, []string{"DEBUG [x] after"}, sink, "level change applies immediately")
}

func TestOrderAndFailureIsolation(t *testing.T) {
	defer guard(t)
	var order []string
	log := NewLogger("x", Debug).
		AddHandler(HandlerFunc(func(string, Level, string) error { order = append(order, "first"); return nil })).
		AddHandler(HandlerFunc(func(string, Level, string) error { return errors.New("disk full") })).
		AddHandler(HandlerFunc(func(string, Level, string) error { panic("boom") })).
		AddHandler(HandlerFunc(func(string, Level, string) error { order = append(order, "fourth"); return nil }))
	log.Info("m")
	lines(t, []string{"first", "fourth"}, order, "handlers in order, failures isolated")
}

func TestConcurrentLogging(t *testing.T) {
	defer guard(t)
	var sink []string
	log := NewLogger("x", Debug).AddHandler(NewMemoryHandler(Debug, &sink, SimpleFormatter{}))
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				log.Info(fmt.Sprintf("%d-%d", id, i))
			}
		}(g)
	}
	wg.Wait()
	if len(sink) != 4000 {
		t.Errorf("expected 4000 lines, got %d", len(sink))
	}
}
