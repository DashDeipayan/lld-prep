package main

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"
)

func wallet(t *testing.T, ids ...string) *Wallet {
	t.Helper()
	w := NewWallet()
	for _, id := range ids {
		if err := w.OpenAccount(id); err != nil {
			t.Fatal(err)
		}
	}
	return w
}

func bal(t *testing.T, w *Wallet, id string) int64 {
	t.Helper()
	b, err := w.Balance(id)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestDepositTransferAndHistory(t *testing.T) {
	defer guard(t)
	w := wallet(t, "a", "b")
	if id, _ := w.Deposit("a", 500); id != "T1" {
		t.Errorf("deposit id %s", id)
	}
	if id, err := w.Transfer("k1", "a", "b", 200); id != "T2" || err != nil {
		t.Errorf("transfer: %s %v", id, err)
	}
	if bal(t, w, "a") != 300 || bal(t, w, "b") != 200 {
		t.Errorf("balances: %d %d", bal(t, w, "a"), bal(t, w, "b"))
	}
	ha, _ := w.History("a")
	if !reflect.DeepEqual(ha, []Txn{{"T1", "EXTERNAL", "a", 500}, {"T2", "a", "b", 200}}) {
		t.Errorf("a history: %+v", ha)
	}
	hb, _ := w.History("b")
	if !reflect.DeepEqual(hb, []Txn{{"T2", "a", "b", 200}}) {
		t.Errorf("b history: %+v", hb)
	}
}

func TestValidation(t *testing.T) {
	defer guard(t)
	w := wallet(t, "a", "b")
	w.Deposit("a", 100)
	checks := []struct {
		err  error
		want error
		msg  string
	}{
		{w.OpenAccount("a"), ErrDuplicateAccount, "duplicate"},
		{func() error { _, e := w.Deposit("a", 0); return e }(), ErrInvalidAmount, "zero deposit"},
		{func() error { _, e := w.Transfer("k1", "a", "zz", 10); return e }(), ErrUnknownAccount, "unknown"},
		{func() error { _, e := w.Transfer("k2", "a", "a", 10); return e }(), ErrSameAccount, "self"},
		{func() error { _, e := w.Transfer("k3", "a", "b", -1); return e }(), ErrInvalidAmount, "negative"},
		{func() error { _, e := w.Transfer("k4", "a", "b", 101); return e }(), ErrInsufficientFunds, "insufficient"},
	}
	for _, c := range checks {
		if !errors.Is(c.err, c.want) {
			t.Errorf("%s: expected %v, got %v", c.msg, c.want, c.err)
		}
	}
	if bal(t, w, "a") != 100 {
		t.Error("nothing should move")
	}
	if h, _ := w.History("a"); len(h) != 1 {
		t.Errorf("failed transfers recorded: %v", h)
	}
}

func TestIdempotentRetries(t *testing.T) {
	defer guard(t)
	w := wallet(t, "a", "b")
	w.Deposit("a", 100)
	id1, _ := w.Transfer("pay-1", "a", "b", 60)
	id2, _ := w.Transfer("pay-1", "a", "b", 60)
	if id1 != id2 {
		t.Errorf("same key -> same id: %s %s", id1, id2)
	}
	if bal(t, w, "a") != 40 {
		t.Errorf("charged once: %d", bal(t, w, "a"))
	}
	if _, err := w.Transfer("pay-2", "a", "b", 60); !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("second payment: %v", err)
	}
	w.Deposit("a", 1000)
	if _, err := w.Transfer("pay-2", "a", "b", 60); !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("retry of failed key must return the same failure: %v", err)
	}
	if bal(t, w, "a") != 1040 {
		t.Errorf("failed key re-executed: %d", bal(t, w, "a"))
	}
}

func TestConcurrentRandomTransfersConserveMoney(t *testing.T) {
	defer guard(t)
	w := NewWallet()
	ids := make([]string, 10)
	for i := range ids {
		ids[i] = fmt.Sprint("acc", i)
		w.OpenAccount(ids[i])
		w.Deposit(ids[i], 1000)
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(int64(seed)))
			for i := 0; i < 125; i++ {
				a := r.Intn(10)
				b := (a + 1 + r.Intn(9)) % 10
				w.Transfer(fmt.Sprintf("t%d-%d", seed, i), ids[a], ids[b], int64(1+r.Intn(300)))
			}
		}(g)
	}
	wg.Wait()
	var total int64
	for _, id := range ids {
		b := bal(t, w, id)
		if b < 0 {
			t.Errorf("%s negative: %d", id, b)
		}
		total += b
	}
	if total != 10000 {
		t.Errorf("money not conserved: %d", total)
	}
}

func TestOppositeDirectionTransfersDoNotDeadlock(t *testing.T) {
	defer guard(t)
	w := wallet(t, "A", "B")
	w.Deposit("A", 1_000_000)
	w.Deposit("B", 1_000_000)
	done := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				w.Transfer(fmt.Sprint("ab", i), "A", "B", 1)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				w.Transfer(fmt.Sprint("ba", i), "B", "A", 1)
			}
		}()
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(6 * time.Second):
		t.Fatal("deadlock: opposite-direction transfers did not finish")
	}
	if bal(t, w, "A")+bal(t, w, "B") != 2_000_000 {
		t.Error("not conserved")
	}
}

func TestConcurrentSameKeyExecutesOnce(t *testing.T) {
	defer guard(t)
	w := wallet(t, "a", "b")
	w.Deposit("a", 1000)
	var wg sync.WaitGroup
	var mu sync.Mutex
	ids := map[string]bool{}
	start := make(chan struct{})
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			id, _ := w.Transfer("same", "a", "b", 10)
			mu.Lock()
			ids[id] = true
			mu.Unlock()
		}()
	}
	close(start)
	wg.Wait()
	if len(ids) != 1 {
		t.Errorf("expected one txn id, got %v", ids)
	}
	if bal(t, w, "a") != 990 {
		t.Errorf("moved more than once: %d", bal(t, w, "a"))
	}
}
