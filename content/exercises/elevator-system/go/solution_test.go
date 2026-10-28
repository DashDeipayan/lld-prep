package main

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func run(s *ElevatorSystem, n int) []string {
	out := []string{}
	for t := 1; t <= n; t++ {
		for _, e := range s.Step() {
			out = append(out, fmt.Sprintf("%d:%s", t, e))
		}
	}
	return out
}

func same(t *testing.T, want, got any, msg string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func dest(t *testing.T, s *ElevatorSystem, id int, floors ...int) {
	t.Helper()
	for _, f := range floors {
		if err := s.RequestDestination(id, f); err != nil {
			t.Fatal(err)
		}
	}
}

func pickup(t *testing.T, s *ElevatorSystem, floor int, d Direction) int {
	t.Helper()
	id, err := s.RequestPickup(floor, d)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestSingleCarLookOrder(t *testing.T) {
	defer guard(t)
	s := NewElevatorSystem(10, 1, NearestCar{})
	dest(t, s, 0, 5, 2, 8)
	same(t, []string{"2:E0@2", "5:E0@5", "8:E0@8"}, run(s, 10), "serves 2, 5, 8 going up")
	same(t, Idle, s.Car(0).Dir(), "idle after last stop")
	same(t, 8, s.Car(0).Floor(), "rests at 8")
}

func TestContinuesSweepThenReverses(t *testing.T) {
	defer guard(t)
	s := NewElevatorSystem(10, 1, NearestCar{})
	dest(t, s, 0, 5)
	run(s, 3)
	same(t, 3, s.Car(0).Floor(), "at 3")
	same(t, Up, s.Car(0).Dir(), "going up")
	dest(t, s, 0, 1, 6)
	same(t, []string{"2:E0@5", "3:E0@6", "8:E0@1"}, run(s, 8), "5, 6 then reverse to 1")
}

func TestStopAtCurrentFloorOpensDoorsWithoutMoving(t *testing.T) {
	defer guard(t)
	s := NewElevatorSystem(5, 1, NearestCar{})
	dest(t, s, 0, 0)
	same(t, []string{"E0@0"}, s.Step(), "doors open")
	same(t, 0, s.Car(0).Floor(), "did not move")
	same(t, []string{}, s.Step(), "nothing left")
}

func TestIdleTieBreakGoesUp(t *testing.T) {
	defer guard(t)
	s := NewElevatorSystem(10, 1, NearestCar{})
	dest(t, s, 0, 4)
	run(s, 4)
	dest(t, s, 0, 2, 6)
	same(t, []string{"2:E0@6", "6:E0@2"}, run(s, 6), "tie -> up first")
}

func TestNearestCarDispatch(t *testing.T) {
	defer guard(t)
	s := NewElevatorSystem(10, 2, NearestCar{})
	same(t, 0, pickup(t, s, 5, Up), "tie -> lowest id")
	run(s, 2)
	same(t, 0, pickup(t, s, 4, Up), "on its way")
	same(t, 1, pickup(t, s, 1, Up), "idle car 1 is cheaper")
	same(t, []int{4, 5}, s.Car(0).Stops(), "car 0 stops")
	same(t, []int{1}, s.Car(1).Stops(), "car 1 stops")
}

func TestOppositeDirectionUsesTurnaround(t *testing.T) {
	defer guard(t)
	s := NewElevatorSystem(20, 2, NearestCar{})
	dest(t, s, 0, 10)
	dest(t, s, 1, 1)
	run(s, 1)
	same(t, 1, pickup(t, s, 8, Down), "car0 cost 11 vs idle car1 cost 7")
}

func TestCustomStrategyAndValidation(t *testing.T) {
	defer guard(t)
	last := StrategyFunc(func(c []*Elevator, _ int, _ Direction) *Elevator { return c[len(c)-1] })
	s := NewElevatorSystem(10, 3, last)
	same(t, 2, pickup(t, s, 3, Down), "strategy injected")
	if _, err := s.RequestPickup(10, Up); !errors.Is(err, ErrInvalidFloor) {
		t.Errorf("floor range: %v", err)
	}
	if _, err := s.RequestPickup(3, Idle); !errors.Is(err, ErrInvalidDirection) {
		t.Errorf("direction: %v", err)
	}
	if err := s.RequestDestination(3, 1); !errors.Is(err, ErrInvalidElevator) {
		t.Errorf("car id: %v", err)
	}
	if err := s.RequestDestination(0, -1); !errors.Is(err, ErrInvalidFloor) {
		t.Errorf("negative floor: %v", err)
	}
}
