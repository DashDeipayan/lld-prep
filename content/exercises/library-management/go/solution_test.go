package main

import (
	"errors"
	"reflect"
	"testing"
)

type world struct {
	today int
	notes []string
}

func lib(t *testing.T) (*Library, *world) {
	t.Helper()
	w := &world{today: 100}
	l := NewLibrary(NewPerDayFine(10, 50), func() int { return w.today },
		NotifierFunc(func(m, msg string) { w.notes = append(w.notes, m+":"+msg) }))
	for _, err := range []error{
		l.AddBook("111", "Effective Java", "Bloch"), l.AddBook("222", "Clean Code", "Martin"),
		l.AddCopy("111", "B-2"), l.AddCopy("111", "B-1"), l.AddCopy("222", "C-1"),
		l.AddMember("asha", 2), l.AddMember("ravi", 2),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	return l, w
}

func borrow(t *testing.T, l *Library, m, isbn string) Loan {
	t.Helper()
	ln, err := l.Borrow(m, isbn)
	if err != nil {
		t.Fatalf("borrow %s %s: %v", m, isbn, err)
	}
	return ln
}

func errIs(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected %v, got %v", msg, target, err)
	}
}

func fine(t *testing.T, l *Library, barcode string, want int64, msg string) {
	t.Helper()
	f, err := l.GiveBack(barcode)
	if err != nil || f != want {
		t.Errorf("%s: expected fine %d, got %d (%v)", msg, want, f, err)
	}
}

func TestBorrowPicksLowestBarcodeAndSetsDueDay(t *testing.T) {
	defer guard(t)
	l, _ := lib(t)
	if got := borrow(t, l, "asha", "111"); got != (Loan{"L1", "B-1", "111", "asha", 114}) {
		t.Errorf("loan: %+v", got)
	}
	if l.AvailableCopies("111") != 1 {
		t.Error("one copy left")
	}
	if borrow(t, l, "ravi", "111").Barcode != "B-2" {
		t.Error("next copy")
	}
	_, err := l.Borrow("asha", "111")
	errIs(t, err, ErrNoCopyAvailable, "none left")
}

func TestLoanLimit(t *testing.T) {
	defer guard(t)
	l, _ := lib(t)
	l.AddMember("solo", 1)
	borrow(t, l, "solo", "111")
	_, err := l.Borrow("solo", "222")
	errIs(t, err, ErrLimitReached, "limit")
}

func TestFinesViaStrategy(t *testing.T) {
	defer guard(t)
	l, w := lib(t)
	a := borrow(t, l, "asha", "111")
	r := borrow(t, l, "ravi", "111")
	w.today = 114
	fine(t, l, a.Barcode, 0, "on due date")
	w.today = 117
	fine(t, l, r.Barcode, 30, "3 days late")
	again := borrow(t, l, "asha", "222")
	w.today = 200
	fine(t, l, again.Barcode, 50, "capped")
}

func TestReturnErrors(t *testing.T) {
	defer guard(t)
	l, _ := lib(t)
	_, err := l.GiveBack("ZZZ")
	errIs(t, err, ErrNotFound, "unknown barcode")
	_, err = l.GiveBack("C-1")
	errIs(t, err, ErrNotOnLoan, "not on loan")
}

func TestReservationHoldsCopyAndNotifies(t *testing.T) {
	defer guard(t)
	l, w := lib(t)
	l.AddMember("kim", 2)
	errIs(t, l.Reserve("kim", "222"), ErrCopyAvailable, "copy available")
	ln := borrow(t, l, "asha", "222")
	if err := l.Reserve("kim", "222"); err != nil {
		t.Fatal(err)
	}
	errIs(t, l.Reserve("kim", "222"), ErrAlreadyReserved, "duplicate")
	l.GiveBack(ln.Barcode)
	if !reflect.DeepEqual(w.notes, []string{"kim:222 is ready for pickup"}) {
		t.Errorf("notes: %v", w.notes)
	}
	if l.AvailableCopies("222") != 0 {
		t.Error("held copy is not available")
	}
	_, err := l.Borrow("ravi", "222")
	errIs(t, err, ErrNoCopyAvailable, "held for kim")
	if borrow(t, l, "kim", "222").Barcode != "C-1" {
		t.Error("kim gets the held copy")
	}
}

func TestHoldLapsesAfterTwoDays(t *testing.T) {
	defer guard(t)
	l, w := lib(t)
	l.AddMember("kim", 2)
	ln := borrow(t, l, "asha", "222")
	l.Reserve("kim", "222")
	l.GiveBack(ln.Barcode)
	w.today += 2
	_, err := l.Borrow("ravi", "222")
	errIs(t, err, ErrNoCopyAvailable, "still held on day +2")
	w.today++
	if l.AvailableCopies("222") != 1 {
		t.Error("hold should lapse")
	}
	if borrow(t, l, "ravi", "222").Barcode != "C-1" {
		t.Error("anyone can borrow now")
	}
}

func TestSearchByTitleOrAuthor(t *testing.T) {
	defer guard(t)
	l, _ := lib(t)
	l.AddBook("333", "Java Concurrency in Practice", "Goetz")
	if got := l.Search("JAVA"); !reflect.DeepEqual(got, []string{"111", "333"}) {
		t.Errorf("title search: %v", got)
	}
	if got := l.Search("mart"); !reflect.DeepEqual(got, []string{"222"}) {
		t.Errorf("author search: %v", got)
	}
	if got := l.Search("python"); len(got) != 0 {
		t.Errorf("no match: %v", got)
	}
}
