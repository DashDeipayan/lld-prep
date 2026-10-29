from lldtest import eq, raises
from solution import Direction, DispatchStrategy, ElevatorSystem, NearestCarStrategy


def run(s, n):
    """Runs n ticks and returns "tick:event" strings."""
    out = []
    for t in range(1, n + 1):
        out.extend("%d:%s" % (t, e) for e in s.step())
    return out


def test_single_car_look_order():
    s = ElevatorSystem(10, 1, NearestCarStrategy())
    s.request_destination(0, 5)
    s.request_destination(0, 2)
    s.request_destination(0, 8)
    eq(["2:E0@2", "5:E0@5", "8:E0@8"], run(s, 10), "serves 2, 5, 8 on the way up")
    eq(Direction.IDLE, s.car(0).direction, "idle after last stop")
    eq(8, s.car(0).floor, "rests at 8")


def test_continues_sweep_then_reverses():
    s = ElevatorSystem(10, 1, NearestCarStrategy())
    s.request_destination(0, 5)
    run(s, 3)
    eq(3, s.car(0).floor, "at 3 going up")
    eq(Direction.UP, s.car(0).direction, "direction up")
    s.request_destination(0, 1)
    s.request_destination(0, 6)
    eq(["2:E0@5", "3:E0@6", "8:E0@1"], run(s, 8), "keeps going up (5, 6), then reverses to 1")


def test_stop_at_current_floor_opens_doors_without_moving():
    s = ElevatorSystem(5, 1, NearestCarStrategy())
    s.request_destination(0, 0)
    eq(["E0@0"], s.step(), "doors open on the same floor")
    eq(0, s.car(0).floor, "did not move")
    eq([], s.step(), "nothing left")


def test_idle_tie_break_goes_up():
    s = ElevatorSystem(10, 1, NearestCarStrategy())
    s.request_destination(0, 4)
    run(s, 4)
    s.request_destination(0, 2)
    s.request_destination(0, 6)
    eq(["2:E0@6", "6:E0@2"], run(s, 6), "equidistant stops: go UP first")


def test_nearest_car_dispatch():
    s = ElevatorSystem(10, 2, NearestCarStrategy())
    eq(0, s.request_pickup(5, Direction.UP), "both idle at 0: tie -> lowest id")
    run(s, 2)
    eq(0, s.request_pickup(4, Direction.UP), "car 0 at 2 moving up, 4 UP is on its way (cost 2 vs 4)")
    eq(1, s.request_pickup(1, Direction.UP), "car 0 must finish its sweep (cost 3+4), idle car 1 costs 1")
    eq([4, 5], s.car(0).stops(), "car 0 stops")
    eq([1], s.car(1).stops(), "car 1 stops")


def test_opposite_direction_uses_turnaround():
    s = ElevatorSystem(20, 2, NearestCarStrategy())
    s.request_destination(0, 10)
    s.request_destination(1, 1)
    run(s, 1)
    # car0 at 1 going UP to 10; car1 at 1 -> idle at 1 after doors open
    eq(1, s.request_pickup(8, Direction.DOWN), "car0: 8 DOWN costs |1-10|+|10-8|=11, idle car1 costs 7")


def test_custom_strategy_and_validation():
    class LastCar(DispatchStrategy):
        def choose(self, cars, floor, direction):
            return cars[-1]

    s = ElevatorSystem(10, 3, LastCar())
    eq(2, s.request_pickup(3, Direction.DOWN), "strategy injected")
    raises(ValueError, lambda: s.request_pickup(10, Direction.UP), "floor out of range")
    raises(ValueError, lambda: s.request_pickup(3, Direction.IDLE), "pickup needs a direction")
    raises(ValueError, lambda: s.request_destination(3, 1), "no such car")
    raises(ValueError, lambda: s.request_destination(0, -1), "negative floor")
