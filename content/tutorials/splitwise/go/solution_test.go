package main

import (
	"errors"
	"reflect"
	"testing"
)

func sw(t *testing.T, users ...string) *Splitwise {
	t.Helper()
	s := NewSplitwise()
	for _, u := range users {
		if err := s.AddUser(u); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func shares(t *testing.T, sp Split, amount int64, ps ...string) []int64 {
	t.Helper()
	m, err := sp.Shares(amount, ps)
	if err != nil {
		t.Fatalf("shares: %v", err)
	}
	out := []int64{}
	for _, p := range ps {
		out = append(out, m[p])
	}
	return out
}

func eqv(t *testing.T, want, got any, msg string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func invalid(t *testing.T, sp Split, amount int64, ps []string, msg string) {
	t.Helper()
	if _, err := sp.Shares(amount, ps); !errors.Is(err, ErrInvalidSplit) {
		t.Errorf("%s: expected ErrInvalidSplit, got %v", msg, err)
	}
}

func expense(t *testing.T, s *Splitwise, payer string, amount int64, ps []string, sp Split) string {
	t.Helper()
	id, err := s.AddExpense(payer, amount, ps, sp)
	if err != nil {
		t.Fatalf("expense: %v", err)
	}
	return id
}

// ---------------- step 1 ----------------
func TestS1_EqualSplitDistributesRemainder(t *testing.T) {
	defer guard(t)
	eqv(t, []int64{34, 33, 33}, shares(t, EqualSplit{}, 100, "a", "b", "c"), "100 / 3")
	eqv(t, []int64{7}, shares(t, EqualSplit{}, 7, "x"), "single")
}

func TestS1_ExactSplitValidates(t *testing.T) {
	defer guard(t)
	eqv(t, []int64{70, 30}, shares(t, NewExactSplit(map[string]int64{"a": 70, "b": 30}), 100, "a", "b"), "exact")
	invalid(t, NewExactSplit(map[string]int64{"a": 70, "b": 20}), 100, []string{"a", "b"}, "sum mismatch")
	invalid(t, NewExactSplit(map[string]int64{"a": 100}), 100, []string{"a", "b"}, "missing participant")
}

func TestS1_PercentSplitRemainders(t *testing.T) {
	defer guard(t)
	p := NewPercentSplit(map[string]int{"a": 33, "b": 33, "c": 34})
	eqv(t, []int64{4, 3, 3}, shares(t, p, 10, "a", "b", "c"), "leftover in participant order")
	eqv(t, []int64{33, 33, 34}, shares(t, p, 100, "a", "b", "c"), "no leftover")
	invalid(t, NewPercentSplit(map[string]int{"a": 50, "b": 40}), 100, []string{"a", "b"}, "sum != 100")
	invalid(t, EqualSplit{}, 0, []string{"a"}, "amount must be positive")
	invalid(t, EqualSplit{}, 10, nil, "no participants")
}

// ---------------- step 2 ----------------
func TestS2_UsersAndExpenses(t *testing.T) {
	defer guard(t)
	s := sw(t, "a", "b", "c")
	if err := s.AddUser("a"); !errors.Is(err, ErrDuplicateUser) {
		t.Errorf("duplicate: %v", err)
	}
	eqv(t, "E1", expense(t, s, "a", 300, []string{"a", "b", "c"}, EqualSplit{}), "expense id")
	eqv(t, int64(100), s.Owes("b", "a"), "b owes a")
	eqv(t, int64(100), s.Owes("c", "a"), "c owes a")
	eqv(t, int64(0), s.Owes("a", "a"), "own share not a debt")
	eqv(t, "E2", expense(t, s, "b", 50, []string{"c"}, NewExactSplit(map[string]int64{"c": 50})), "payer need not participate")
	eqv(t, int64(50), s.Owes("c", "b"), "c owes b")
}

func TestS2_ExpenseValidation(t *testing.T) {
	defer guard(t)
	s := sw(t, "a", "b")
	if _, err := s.AddExpense("z", 10, []string{"a"}, EqualSplit{}); !errors.Is(err, ErrUnknownUser) {
		t.Errorf("unknown payer: %v", err)
	}
	if _, err := s.AddExpense("a", 10, []string{"a", "q"}, EqualSplit{}); !errors.Is(err, ErrUnknownUser) {
		t.Errorf("unknown participant: %v", err)
	}
	if _, err := s.AddExpense("a", 10, []string{"b", "b"}, EqualSplit{}); !errors.Is(err, ErrInvalidSplit) {
		t.Errorf("duplicate participant: %v", err)
	}
	eqv(t, int64(0), s.Owes("b", "a"), "failed expenses change nothing")
}

// ---------------- step 3 ----------------
func TestS3_NettingBothDirections(t *testing.T) {
	defer guard(t)
	s := sw(t, "a", "b")
	expense(t, s, "b", 30, []string{"a"}, EqualSplit{})
	expense(t, s, "a", 10, []string{"b"}, EqualSplit{})
	eqv(t, int64(20), s.Owes("a", "b"), "netted")
	eqv(t, int64(0), s.Owes("b", "a"), "never negative")
}

func TestS3_BalancesSumToZero(t *testing.T) {
	defer guard(t)
	s := sw(t, "a", "b", "c", "d")
	expense(t, s, "a", 1000, []string{"a", "b", "c", "d"}, EqualSplit{})
	expense(t, s, "c", 333, []string{"a", "b", "c"}, EqualSplit{})
	eqv(t, int64(750-111), s.Balance("a"), "a")
	eqv(t, int64(-250-111), s.Balance("b"), "b")
	var sum int64
	for _, u := range []string{"a", "b", "c", "d"} {
		sum += s.Balance(u)
	}
	eqv(t, int64(0), sum, "conserved")
}

// ---------------- step 4 ----------------
func TestS4_SettleUp(t *testing.T) {
	defer guard(t)
	s := sw(t, "a", "b")
	expense(t, s, "a", 100, []string{"a", "b"}, EqualSplit{})
	if err := s.Settle("b", "a", 30); err != nil {
		t.Fatal(err)
	}
	eqv(t, int64(20), s.Owes("b", "a"), "partial")
	if err := s.Settle("b", "a", 21); !errors.Is(err, ErrOverSettle) {
		t.Errorf("overpay: %v", err)
	}
	if err := s.Settle("b", "a", 0); !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("zero: %v", err)
	}
	s.Settle("b", "a", 20)
	eqv(t, int64(0), s.Balance("a"), "settled")
}

// ---------------- step 5 ----------------
func TestS5_SimplifyChain(t *testing.T) {
	defer guard(t)
	s := sw(t, "a", "b", "c")
	expense(t, s, "b", 50, []string{"a"}, NewExactSplit(map[string]int64{"a": 50}))
	expense(t, s, "c", 50, []string{"b"}, NewExactSplit(map[string]int64{"b": 50}))
	eqv(t, []Transfer{{"a", "c", 50}}, s.Simplify(), "chain collapses")
	eqv(t, int64(50), s.Owes("a", "b"), "simplify doesn't mutate")
}

func TestS5_SimplifyGreedyOrderAndBound(t *testing.T) {
	defer guard(t)
	s := sw(t, "a", "b", "c", "d", "e")
	expense(t, s, "a", 400, []string{"b", "c", "d", "e"}, EqualSplit{})
	expense(t, s, "b", 300, []string{"c", "d", "e"}, EqualSplit{})
	eqv(t, []Transfer{{"c", "a", 200}, {"d", "a", 200}, {"e", "b", 200}}, s.Simplify(), "greedy order")
	eqv(t, []Transfer{}, sw(t, "x", "y").Simplify(), "nothing to settle")
}
