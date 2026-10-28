package main

import (
	"errors"
	"reflect"
	"testing"
)

func pay(t *testing.T, f *PaymentFactory, method string, amount int64) Receipt {
	t.Helper()
	p, err := f.Create(method)
	if err != nil {
		t.Fatalf("create %q: %v", method, err)
	}
	r, err := p.Pay(amount)
	if err != nil {
		t.Fatalf("pay: %v", err)
	}
	return r
}

func TestDefaultsAndFees(t *testing.T) {
	defer guard(t)
	f := NewDefaultFactory()
	cases := []struct {
		m    string
		amt  int64
		want Receipt
	}{
		{"card", 10_000, Receipt{"card", 10_000, 200}},
		{"upi", 10_000, Receipt{"upi", 10_000, 0}},
		{"wallet", 10_050, Receipt{"wallet", 10_050, 100}},
	}
	for _, c := range cases {
		if got := pay(t, f, c.m, c.amt); got != c.want {
			t.Errorf("%s: expected %+v, got %+v", c.m, c.want, got)
		}
	}
}

func TestCaseInsensitive(t *testing.T) {
	defer guard(t)
	f := NewDefaultFactory()
	if pay(t, f, "CARD", 100).Method != "card" || pay(t, f, " Upi ", 100).Method != "upi" {
		t.Error("lookups should be case-insensitive and trimmed")
	}
}

func TestSupportedSorted(t *testing.T) {
	defer guard(t)
	if got := NewDefaultFactory().SupportedMethods(); !reflect.DeepEqual(got, []string{"card", "upi", "wallet"}) {
		t.Errorf("expected sorted [card upi wallet], got %v", got)
	}
}

func TestFreshInstances(t *testing.T) {
	defer guard(t)
	f := NewDefaultFactory()
	a, _ := f.Create("card")
	b, _ := f.Create("card")
	if reflect.ValueOf(a).Kind() == reflect.Ptr && a == b {
		t.Error("Create must return a new instance each call")
	}
}

type bnpl struct{}

func (bnpl) Method() string               { return "bnpl" }
func (bnpl) Pay(a int64) (Receipt, error) { return Receipt{"bnpl", a, 99}, nil }

func TestRegisterNewMethodWithoutEditingFactory(t *testing.T) {
	defer guard(t)
	f := NewDefaultFactory()
	if err := f.Register("BNPL", func() PaymentProcessor { return bnpl{} }); err != nil {
		t.Fatalf("register: %v", err)
	}
	if pay(t, f, "bnpl", 1000).Fee != 99 {
		t.Error("plugged-in processor not used")
	}
	if got := f.SupportedMethods(); !reflect.DeepEqual(got, []string{"bnpl", "card", "upi", "wallet"}) {
		t.Errorf("expected lower-cased sorted names, got %v", got)
	}
}

func TestErrors(t *testing.T) {
	defer guard(t)
	f := NewDefaultFactory()
	if err := f.Register("Card", NewDefaultFactory().mustCreator("card")); !errors.Is(err, ErrDuplicateMethod) {
		t.Errorf("duplicate register: expected ErrDuplicateMethod, got %v", err)
	}
	if _, err := f.Create("bitcoin"); !errors.Is(err, ErrUnsupportedMethod) {
		t.Errorf("unknown method: expected ErrUnsupportedMethod, got %v", err)
	}
	p, _ := f.Create("card")
	if _, err := p.Pay(0); !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("zero amount: expected ErrInvalidAmount, got %v", err)
	}
	if _, err := p.Pay(-5); !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("negative amount: expected ErrInvalidAmount, got %v", err)
	}
}

// mustCreator returns a constructor that reuses whatever the factory builds for method.
func (f *PaymentFactory) mustCreator(method string) func() PaymentProcessor {
	return func() PaymentProcessor {
		p, _ := f.Create(method)
		return p
	}
}
