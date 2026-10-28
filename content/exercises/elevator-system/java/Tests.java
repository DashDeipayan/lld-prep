import java.util.*;

public class Tests {
    /** runs `n` ticks and returns "tick:event" strings */
    static List<String> run(ElevatorSystem s, int n) {
        List<String> out = new ArrayList<>();
        for (int t = 1; t <= n; t++) for (String e : s.step()) out.add(t + ":" + e);
        return out;
    }

    static void testSingleCarLookOrder() {
        ElevatorSystem s = new ElevatorSystem(10, 1, new NearestCarStrategy());
        s.requestDestination(0, 5);
        s.requestDestination(0, 2);
        s.requestDestination(0, 8);
        T.eq(List.of("2:E0@2", "5:E0@5", "8:E0@8"), run(s, 10), "serves 2, 5, 8 on the way up");
        T.eq(Direction.IDLE, s.car(0).direction(), "idle after last stop");
        T.eq(8, s.car(0).floor(), "rests at 8");
    }

    static void testContinuesSweepThenReverses() {
        ElevatorSystem s = new ElevatorSystem(10, 1, new NearestCarStrategy());
        s.requestDestination(0, 5);
        run(s, 3);
        T.eq(3, s.car(0).floor(), "at 3 going up");
        T.eq(Direction.UP, s.car(0).direction(), "direction up");
        s.requestDestination(0, 1);
        s.requestDestination(0, 6);
        T.eq(List.of("2:E0@5", "3:E0@6", "8:E0@1"), run(s, 8), "keeps going up (5, 6), then reverses to 1");
    }

    static void testStopAtCurrentFloorOpensDoorsWithoutMoving() {
        ElevatorSystem s = new ElevatorSystem(5, 1, new NearestCarStrategy());
        s.requestDestination(0, 0);
        T.eq(List.of("E0@0"), s.step(), "doors open on the same floor");
        T.eq(0, s.car(0).floor(), "did not move");
        T.eq(List.of(), s.step(), "nothing left");
    }

    static void testIdleTieBreakGoesUp() {
        ElevatorSystem s = new ElevatorSystem(10, 1, new NearestCarStrategy());
        s.requestDestination(0, 4);
        run(s, 4);
        s.requestDestination(0, 2);
        s.requestDestination(0, 6);
        T.eq(List.of("2:E0@6", "6:E0@2"), run(s, 6), "equidistant stops: go UP first");
    }

    static void testNearestCarDispatch() {
        ElevatorSystem s = new ElevatorSystem(10, 2, new NearestCarStrategy());
        T.eq(0, s.requestPickup(5, Direction.UP), "both idle at 0: tie -> lowest id");
        run(s, 2);
        T.eq(0, s.requestPickup(4, Direction.UP), "car 0 at 2 moving up, 4 UP is on its way (cost 2 vs 4)");
        T.eq(1, s.requestPickup(1, Direction.UP), "car 0 must finish its sweep (cost 3+4), idle car 1 costs 1");
        T.eq(List.of(4, 5), s.car(0).stops(), "car 0 stops");
        T.eq(List.of(1), s.car(1).stops(), "car 1 stops");
    }

    static void testOppositeDirectionUsesTurnaround() {
        ElevatorSystem s = new ElevatorSystem(20, 2, new NearestCarStrategy());
        s.requestDestination(0, 10);
        s.requestDestination(1, 1);
        run(s, 1);
        // car0 at 1 going UP to 10; car1 at 1 -> idle at 1 after doors open
        T.eq(1, s.requestPickup(8, Direction.DOWN), "car0: 8 DOWN costs |1-10|+|10-8|=11, idle car1 costs 7");
    }

    static void testCustomStrategyAndValidation() {
        DispatchStrategy lastCar = (cars, f, d) -> cars.get(cars.size() - 1);
        ElevatorSystem s = new ElevatorSystem(10, 3, lastCar);
        T.eq(2, s.requestPickup(3, Direction.DOWN), "strategy injected");
        T.throwsEx(IllegalArgumentException.class, () -> s.requestPickup(10, Direction.UP), "floor out of range");
        T.throwsEx(IllegalArgumentException.class, () -> s.requestPickup(3, Direction.IDLE), "pickup needs a direction");
        T.throwsEx(IllegalArgumentException.class, () -> s.requestDestination(3, 1), "no such car");
        T.throwsEx(IllegalArgumentException.class, () -> s.requestDestination(0, -1), "negative floor");
    }
}
