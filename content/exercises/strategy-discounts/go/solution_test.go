package main

import "testing"

func sample() *Cart {
	c := NewCart()
	c.Add("BOOK", 500, 3)
	c.Add("PEN", 100, 1)
	return c
}

func eq(t *testing.T, want, got int64, msg string) {
	t.Helper()
	if want != got {
		t.Errorf("%s\nexpected: %d\n  actual: %d", msg, want, got)
	}
}

func TestPercentageOff(t *testing.T) {
	defer guard(t)
	eq(t, 160, NewPercentageOff(10).Discount(sample()), "10% of 1600")
	eq(t, 0, NewPercentageOff(0).Discount(sample()), "0% is nothing")
}

func TestFlatOffRespectsMinimum(t *testing.T) {
	defer guard(t)
	eq(t, 300, NewFlatOff(300, 1500).Discount(sample()), "subtotal 1600 >= 1500")
	eq(t, 300, NewFlatOff(300, 1600).Discount(sample()), "boundary: subtotal == minimum qualifies")
	eq(t, 0, NewFlatOff(300, 2000).Discount(sample()), "below minimum gives no discount")
}

func TestBuyNGetOneFree(t *testing.T) {
	defer guard(t)
	eq(t, 500, NewBuyNGetOneFree("BOOK", 2).Discount(sample()), "3 books, buy 2 get 1 -> one free")
	c := NewCart()
	c.Add("BOOK", 500, 7)
	eq(t, 1000, NewBuyNGetOneFree("BOOK", 2).Discount(c), "7 books -> two full groups of 3")
	eq(t, 0, NewBuyNGetOneFree("MUG", 1).Discount(sample()), "SKU not in cart")
}

func TestCheckoutPicksBestSingleDiscount(t *testing.T) {
	defer guard(t)
	co := NewCheckout(NewPercentageOff(10), NewFlatOff(300, 1500), NewBuyNGetOneFree("BOOK", 2))
	eq(t, 1100, co.Total(sample()), "best discount is 500, promotions do not stack")
}

func TestCheckoutWithoutStrategies(t *testing.T) {
	defer guard(t)
	eq(t, 1600, NewCheckout().Total(sample()), "no promotions -> subtotal")
	eq(t, 0, NewCheckout(NewPercentageOff(50)).Total(NewCart()), "empty cart")
}

func TestTotalNeverNegative(t *testing.T) {
	defer guard(t)
	eq(t, 0, NewCheckout(NewFlatOff(10_000, 0)).Total(sample()), "discount larger than subtotal clamps to 0")
}

type fixed int64

func (f fixed) Discount(*Cart) int64 { return int64(f) }

func TestNewStrategyNeedsNoCheckoutChange(t *testing.T) {
	defer guard(t)
	eq(t, 1558, NewCheckout(fixed(42)).Total(sample()), "a brand-new strategy works without touching Checkout")
}
