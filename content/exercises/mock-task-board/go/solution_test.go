package main

import (
	"errors"
	"reflect"
	"testing"
)

func create(t *testing.T, b *TaskBoard, title string, prio int) string {
	t.Helper()
	id, err := b.Create(title, "pm", prio)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func move(t *testing.T, b *TaskBoard, id string, to Status, actor string) {
	t.Helper()
	if err := b.Move(id, to, actor); err != nil {
		t.Fatalf("move %s -> %v by %s: %v", id, to, actor, err)
	}
}

func is(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected %v, got %v", msg, target, err)
	}
}

func deep(t *testing.T, want, got any, msg string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func TestWorkflowAndActors(t *testing.T) {
	defer guard(t)
	b := NewTaskBoard()
	id := create(t, b, "Fix login", 3)
	deep(t, "T-1", id, "id")
	is(t, b.Move(id, InProgress, "bob"), ErrIllegalTransition, "unassigned")
	b.Assign(id, "bob")
	is(t, b.Move(id, InProgress, "eve"), ErrUnauthorized, "only assignee starts")
	move(t, b, id, InProgress, "bob")
	is(t, b.Move(id, Done, "alice"), ErrIllegalTransition, "skip review")
	move(t, b, id, InReview, "bob")
	is(t, b.Move(id, Done, "bob"), ErrUnauthorized, "self-approval")
	move(t, b, id, InProgress, "alice")
	move(t, b, id, InReview, "bob")
	move(t, b, id, Done, "alice")
	st, _ := b.Status(id)
	deep(t, Done, st, "done")
	is(t, b.Move(id, Todo, "bob"), ErrIllegalTransition, "terminal")
	is(t, b.Assign(id, "eve"), ErrIllegalTransition, "reassign done")
}

func TestHistory(t *testing.T) {
	defer guard(t)
	b := NewTaskBoard()
	id := create(t, b, "X", 1)
	b.Assign(id, "bob")
	move(t, b, id, InProgress, "bob")
	move(t, b, id, InReview, "bob")
	h, _ := b.History(id)
	deep(t, []string{"created by pm", "assigned to bob", "TODO->IN_PROGRESS by bob", "IN_PROGRESS->IN_REVIEW by bob"}, h, "audit")
}

func TestDependenciesBlockStart(t *testing.T) {
	defer guard(t)
	b := NewTaskBoard()
	api, ui := create(t, b, "API", 2), create(t, b, "UI", 2)
	if err := b.AddDependency(ui, api); err != nil {
		t.Fatal(err)
	}
	b.Assign(api, "bob")
	b.Assign(ui, "carol")
	deep(t, []string{ui}, b.BlockedTasks(), "ui blocked")
	is(t, b.Move(ui, InProgress, "carol"), ErrBlocked, "blocked start")
	move(t, b, api, InProgress, "bob")
	move(t, b, api, InReview, "bob")
	move(t, b, api, Done, "alice")
	deep(t, []string{}, b.BlockedTasks(), "unblocked")
	move(t, b, ui, InProgress, "carol")
}

func TestCycleDetection(t *testing.T) {
	defer guard(t)
	b := NewTaskBoard()
	a, c, d := create(t, b, "A", 1), create(t, b, "B", 1), create(t, b, "C", 1)
	b.AddDependency(a, c)
	b.AddDependency(c, d)
	is(t, b.AddDependency(d, a), ErrCycle, "cycle")
	is(t, b.AddDependency(a, a), ErrCycle, "self")
	if err := b.AddDependency(a, d); err != nil {
		t.Errorf("non-cyclic edge: %v", err)
	}
	is(t, b.AddDependency(a, "T-99"), ErrNotFound, "unknown")
}

func TestTasksForUserSorted(t *testing.T) {
	defer guard(t)
	b := NewTaskBoard()
	low, high, mid1, mid2 := create(t, b, "low", 1), create(t, b, "high", 5), create(t, b, "m1", 3), create(t, b, "m2", 3)
	for _, id := range []string{low, high, mid1, mid2} {
		b.Assign(id, "bob")
	}
	b.Assign(mid2, "carol")
	deep(t, []string{high, mid1, low}, b.TasksFor("bob"), "priority desc then creation")
	_, err := b.Create("bad", "pm", 6)
	is(t, err, ErrInvalid, "priority range")
}
