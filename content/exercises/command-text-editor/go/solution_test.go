package main

import (
	"errors"
	"testing"
)

func text(t *testing.T, e *TextEditor, want, msg string) {
	t.Helper()
	if got := e.Text(); got != want {
		t.Errorf("%s\nexpected: %q\n  actual: %q", msg, want, got)
	}
}

func ok(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func truth(t *testing.T, cond bool, msg string) {
	t.Helper()
	if !cond {
		t.Error(msg)
	}
}

func TestInsertAndDelete(t *testing.T) {
	defer guard(t)
	e := NewTextEditor(10)
	ok(t, e.Insert(0, "hello"))
	ok(t, e.Insert(5, " world"))
	ok(t, e.Insert(0, ">"))
	text(t, e, ">hello world", "inserts")
	ok(t, e.Delete(0, 1))
	ok(t, e.Delete(5, 11))
	text(t, e, "hello", "deletes")
}

func TestUndoRedoSequence(t *testing.T) {
	defer guard(t)
	e := NewTextEditor(10)
	ok(t, e.Insert(0, "hello"))
	ok(t, e.Insert(5, " world"))
	ok(t, e.Delete(0, 6))
	text(t, e, "world", "after edits")
	truth(t, e.Undo(), "undo 1")
	text(t, e, "hello world", "undo delete restores removed text")
	truth(t, e.Undo(), "undo 2")
	text(t, e, "hello", "undo insert")
	truth(t, e.Redo(), "redo")
	text(t, e, "hello world", "redo insert")
}

func TestNewEditClearsRedo(t *testing.T) {
	defer guard(t)
	e := NewTextEditor(10)
	ok(t, e.Insert(0, "ab"))
	e.Undo()
	ok(t, e.Insert(0, "x"))
	truth(t, !e.Redo(), "redo must be cleared by a new edit")
	text(t, e, "x", "text")
}

func TestEmptyStacks(t *testing.T) {
	defer guard(t)
	e := NewTextEditor(10)
	truth(t, !e.Undo(), "nothing to undo")
	truth(t, !e.Redo(), "nothing to redo")
	ok(t, e.Insert(0, "a"))
	e.Undo()
	truth(t, !e.Undo(), "undo past beginning")
	text(t, e, "", "empty")
}

func TestInvalidOpsChangeNothing(t *testing.T) {
	defer guard(t)
	e := NewTextEditor(10)
	ok(t, e.Insert(0, "abc"))
	for _, err := range []error{e.Insert(4, "x"), e.Insert(-1, "x"), e.Delete(2, 1), e.Delete(0, 4)} {
		if !errors.Is(err, ErrOutOfRange) {
			t.Errorf("expected ErrOutOfRange, got %v", err)
		}
	}
	text(t, e, "abc", "text unchanged")
	truth(t, e.Undo(), "only the valid insert is in history")
	truth(t, !e.Undo(), "invalid ops were not recorded")
}

func TestNoOpsNotRecorded(t *testing.T) {
	defer guard(t)
	e := NewTextEditor(10)
	ok(t, e.Insert(0, "abc"))
	ok(t, e.Insert(1, ""))
	ok(t, e.Delete(2, 2))
	e.Undo()
	text(t, e, "", "single undo reverts the only real edit")
	truth(t, !e.Undo(), "no-ops not in history")
}

func TestBoundedHistoryDropsOldest(t *testing.T) {
	defer guard(t)
	e := NewTextEditor(2)
	ok(t, e.Insert(0, "a"))
	ok(t, e.Insert(1, "b"))
	ok(t, e.Insert(2, "c"))
	truth(t, e.Undo(), "undo c")
	truth(t, e.Undo(), "undo b")
	truth(t, !e.Undo(), "a dropped from history")
	text(t, e, "a", "oldest edit can no longer be undone")
}
