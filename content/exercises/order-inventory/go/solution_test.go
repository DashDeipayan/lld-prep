package main

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

type fakeGW struct {
	mu      sync.Mutex
	approve bool
	log     []string
}

func (g *fakeGW) Charge(id string, a int64) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.log = append(g.log, fmt.Sprintf("charge:%s:%d", id, a))
	return g.approve
}

func (g *fakeGW) Refund(id string, a int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.log = append(g.log, fmt.Sprintf("refund:%s:%d", id, a))
}

func shop(t *testing.T) (*OrderService, *fakeGW) {
	t.Helper()
	gw := &fakeGW{approve: true}
	s := NewOrderService(gw)
	s.AddProduct("A", 100, 5)
	s.AddProduct("B", 250, 2)
	return s, gw
}

func place(t *testing.T, s *OrderService, c string, lines ...OrderLine) string {
	t.Helper()
	id, err := s.PlaceOrder(c, lines)
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	return id
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func is(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected %v, got %v", msg, target, err)
	}
}

func eqv(t *testing.T, want, got any, msg string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func TestPlaceOrderReservesAndTotals(t *testing.T) {
	defer guard(t)
	s, _ := shop(t)
	o := place(t, s, "asha", OrderLine{"A", 2}, OrderLine{"B", 1}, OrderLine{"A", 1})
	eqv(t, "O1", o, "id")
	total, _ := s.Total(o)
	eqv(t, int64(550), total, "total")
	eqv(t, 2, s.Stock("A"), "A reserved")
	eqv(t, 1, s.Stock("B"), "B reserved")
	st, _ := s.Status(o)
	eqv(t, Created, st, "created")
}

func TestAllOrNothingReservation(t *testing.T) {
	defer guard(t)
	s, _ := shop(t)
	_, err := s.PlaceOrder("asha", []OrderLine{{"A", 1}, {"B", 3}})
	is(t, err, ErrOutOfStock, "B short")
	eqv(t, 5, s.Stock("A"), "A untouched")
	_, err = s.PlaceOrder("asha", []OrderLine{{"Z", 1}})
	is(t, err, ErrNotFound, "unknown sku")
	_, err = s.PlaceOrder("asha", []OrderLine{{"A", 0}})
	is(t, err, ErrInvalid, "zero qty")
	_, err = s.PlaceOrder("asha", nil)
	is(t, err, ErrInvalid, "empty")
	must(t, s.Restock("B", 1))
	eqv(t, 3, s.Stock("B"), "restocked")
}

func TestHappyPathLifecycle(t *testing.T) {
	defer guard(t)
	s, gw := shop(t)
	o := place(t, s, "asha", OrderLine{"B", 2})
	is(t, s.Ship(o), ErrIllegalTransition, "ship before pay")
	must(t, s.Pay(o))
	eqv(t, []string{"charge:O1:500"}, gw.log, "charged")
	must(t, s.Ship(o))
	must(t, s.Deliver(o))
	is(t, s.Cancel(o), ErrIllegalTransition, "cancel delivered")
	is(t, s.Pay(o), ErrIllegalTransition, "pay twice")
}

func TestPaymentFailureKeepsOrderCreated(t *testing.T) {
	defer guard(t)
	s, gw := shop(t)
	o := place(t, s, "asha", OrderLine{"A", 1})
	gw.approve = false
	is(t, s.Pay(o), ErrPaymentFailed, "declined")
	st, _ := s.Status(o)
	eqv(t, Created, st, "still created")
	eqv(t, 4, s.Stock("A"), "still reserved")
	gw.approve = true
	must(t, s.Pay(o))
}

func TestCancellationRules(t *testing.T) {
	defer guard(t)
	s, gw := shop(t)
	unpaid := place(t, s, "asha", OrderLine{"A", 2})
	must(t, s.Cancel(unpaid))
	eqv(t, 5, s.Stock("A"), "released")
	eqv(t, 0, len(gw.log), "no refund")
	paid := place(t, s, "asha", OrderLine{"B", 1})
	must(t, s.Pay(paid))
	must(t, s.Cancel(paid))
	eqv(t, []string{"charge:O2:250", "refund:O2:250"}, gw.log, "refund")
	eqv(t, 2, s.Stock("B"), "released")
	shipped := place(t, s, "ravi", OrderLine{"A", 1})
	must(t, s.Pay(shipped))
	must(t, s.Ship(shipped))
	is(t, s.Cancel(shipped), ErrIllegalTransition, "cancel shipped")
	eqv(t, []string{"O1", "O2"}, s.OrdersOf("asha"), "orders of customer")
}

func TestNoOversellingUnderContention(t *testing.T) {
	defer guard(t)
	s, _ := shop(t)
	s.AddProduct("HOT", 10, 30)
	var ok atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			if _, err := s.PlaceOrder(fmt.Sprint("c", i), []OrderLine{{"HOT", 1}}); err == nil {
				ok.Add(1)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	eqv(t, int32(30), ok.Load(), "exactly 30")
	eqv(t, 0, s.Stock("HOT"), "stock 0")
}
