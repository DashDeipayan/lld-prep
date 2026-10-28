package main

import (
	"errors"
	"testing"
)

func stocked(t *testing.T) *VendingMachine {
	t.Helper()
	vm := NewVendingMachine()
	must(t, vm.Stock(Product{"A1", "Cola", 125}, 1))
	must(t, vm.Stock(Product{"B2", "Chips", 50}, 2))
	return vm
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func is(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected %v, got %v", msg, target, err)
	}
}

func eqs(t *testing.T, want, got any, msg string) {
	t.Helper()
	if want != got {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func TestStartsSoldOutAndRejectsCoins(t *testing.T) {
	defer guard(t)
	vm := NewVendingMachine()
	eqs(t, "SOLD_OUT", vm.State(), "new machine")
	is(t, vm.InsertCoin(25), ErrSoldOut, "coin while sold out")
	_, err := vm.Select("A1")
	is(t, err, ErrSoldOut, "select while sold out")
	eqs(t, 0, vm.Cancel(), "cancel while sold out")
}

func TestStockingMovesToIdle(t *testing.T) {
	defer guard(t)
	vm := stocked(t)
	eqs(t, "IDLE", vm.State(), "after stocking")
	eqs(t, 2, vm.Quantity("B2"), "quantity")
	must(t, vm.Stock(Product{"B2", "Chips", 50}, 3))
	eqs(t, 5, vm.Quantity("B2"), "stocking adds")
	is(t, vm.Stock(Product{"C3", "Gum", 10}, 0), ErrInvalidQuantity, "qty must be positive")
}

func TestCoinsAndInvalidCoin(t *testing.T) {
	defer guard(t)
	vm := stocked(t)
	is(t, vm.InsertCoin(3), ErrInvalidCoin, "invalid denomination")
	eqs(t, "IDLE", vm.State(), "invalid coin changes nothing")
	must(t, vm.InsertCoin(25))
	must(t, vm.InsertCoin(10))
	eqs(t, "HAS_MONEY", vm.State(), "has money")
	eqs(t, 35, vm.Balance(), "balance")
}

func TestSelectInIdleRejected(t *testing.T) {
	defer guard(t)
	_, err := stocked(t).Select("A1")
	is(t, err, ErrInsertMoneyFirst, "select in idle")
}

func TestSuccessfulSaleWithChange(t *testing.T) {
	defer guard(t)
	vm := stocked(t)
	for _, c := range []int{100, 25, 25} {
		must(t, vm.InsertCoin(c))
	}
	d, err := vm.Select("A1")
	must(t, err)
	eqs(t, "A1", d.Product.Code, "product dispensed")
	eqs(t, 25, d.Change, "change")
	eqs(t, 0, vm.Balance(), "balance reset")
	eqs(t, 0, vm.Quantity("A1"), "one unit taken")
	eqs(t, "IDLE", vm.State(), "stock remains -> IDLE")
}

func TestFailedSelectsKeepState(t *testing.T) {
	defer guard(t)
	vm := stocked(t)
	must(t, vm.InsertCoin(25))
	_, err := vm.Select("A1")
	is(t, err, ErrInsufficientFunds, "insufficient funds")
	_, err = vm.Select("ZZ")
	is(t, err, ErrUnknownProduct, "unknown product")
	eqs(t, "HAS_MONEY", vm.State(), "still has money")
	eqs(t, 25, vm.Balance(), "balance untouched")
	is(t, vm.Stock(Product{"C3", "Gum", 10}, 1), ErrBusy, "cannot stock mid-transaction")
}

func TestOutOfStockAndSoldOut(t *testing.T) {
	defer guard(t)
	vm := NewVendingMachine()
	must(t, vm.Stock(Product{"A1", "Cola", 125}, 1))
	must(t, vm.Stock(Product{"B2", "Chips", 50}, 1))
	must(t, vm.InsertCoin(100))
	must(t, vm.InsertCoin(25))
	_, err := vm.Select("A1")
	must(t, err)
	must(t, vm.InsertCoin(100))
	_, err = vm.Select("A1")
	is(t, err, ErrOutOfStock, "A1 out of stock")
	d, err := vm.Select("B2")
	must(t, err)
	eqs(t, 50, d.Change, "change from 100 for 50")
	eqs(t, "SOLD_OUT", vm.State(), "last unit sold -> SOLD_OUT")
}

func TestCancelRefunds(t *testing.T) {
	defer guard(t)
	vm := stocked(t)
	must(t, vm.InsertCoin(100))
	must(t, vm.InsertCoin(5))
	eqs(t, 105, vm.Cancel(), "full refund")
	eqs(t, "IDLE", vm.State(), "back to idle")
	eqs(t, 0, vm.Balance(), "balance cleared")
}
