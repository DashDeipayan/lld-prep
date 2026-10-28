package main

import (
	"fmt"
	"reflect"
	"testing"
)

func same(t *testing.T, want, got []string, msg string) {
	t.Helper()
	if len(want) == 0 && len(got) == 0 {
		return
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func TestNotifiesOnlyOnChangeInOrder(t *testing.T) {
	defer guard(t)
	tk := NewStockTicker()
	var log []string
	tk.Subscribe(ListenerFunc(func(s string, p int64) { log = append(log, fmt.Sprintf("a:%s%d", s, p)) }))
	tk.Subscribe(ListenerFunc(func(s string, p int64) { log = append(log, fmt.Sprintf("b:%s%d", s, p)) }))
	tk.Update("X", 1)
	tk.Update("X", 1)
	tk.Update("Y", 1)
	tk.Update("X", 2)
	same(t, []string{"a:X1", "b:X1", "a:Y1", "b:Y1", "a:X2", "b:X2"}, log, "same price twice is not a change; order = subscription order")
}

func TestUnsubscribeHandle(t *testing.T) {
	defer guard(t)
	tk := NewStockTicker()
	var got []string
	off := tk.Subscribe(ListenerFunc(func(s string, p int64) { got = append(got, fmt.Sprint(p)) }))
	if tk.ListenerCount() != 1 {
		t.Errorf("expected 1 listener, got %d", tk.ListenerCount())
	}
	tk.Update("X", 1)
	off()
	off()
	tk.Update("X", 2)
	same(t, []string{"1"}, got, "no events after unsubscribe")
	if tk.ListenerCount() != 0 {
		t.Errorf("expected 0 listeners after unsubscribe, got %d", tk.ListenerCount())
	}
}

func TestFaultyListenerIsolated(t *testing.T) {
	defer guard(t)
	tk := NewStockTicker()
	var got []string
	tk.Subscribe(ListenerFunc(func(string, int64) { panic("boom") }))
	tk.Subscribe(ListenerFunc(func(s string, p int64) { got = append(got, fmt.Sprint(p)) }))
	tk.Update("X", 7)
	same(t, []string{"7"}, got, "listener after the panicking one still notified")
}

func TestUnsubscribeDuringNotification(t *testing.T) {
	defer guard(t)
	tk := NewStockTicker()
	var log []string
	var off func()
	off = tk.Subscribe(ListenerFunc(func(s string, p int64) { log = append(log, fmt.Sprint("once", p)); off() }))
	tk.Subscribe(ListenerFunc(func(s string, p int64) { log = append(log, fmt.Sprint("always", p)) }))
	tk.Update("X", 1)
	tk.Update("X", 2)
	same(t, []string{"once1", "always1", "always2"}, log, "self-unsubscribing listener gets one event, others unaffected")
}

func TestAlertEdgeTriggeredAbove(t *testing.T) {
	defer guard(t)
	tk := NewStockTicker()
	var sink []string
	tk.Subscribe(NewThresholdAlert("AAPL", 150, true, &sink))
	for _, p := range []int64{149, 151, 155, 148, 150} {
		tk.Update("AAPL", p)
	}
	same(t, []string{"AAPL>=150 @151", "AAPL>=150 @150"}, sink, "fires on crossing, re-arms after dropping below")
}

func TestAlertBelowAndOtherSymbols(t *testing.T) {
	defer guard(t)
	tk := NewStockTicker()
	var sink []string
	tk.Subscribe(NewThresholdAlert("TSLA", 100, false, &sink))
	tk.Update("AAPL", 10)
	tk.Update("TSLA", 99)
	tk.Update("TSLA", 98)
	same(t, []string{"TSLA<=100 @99"}, sink, "below-threshold alert, other symbols ignored, no repeat while still below")
}
