import java.util.*;

enum Direction { UP, DOWN, IDLE }

interface DispatchStrategy {
    Elevator choose(List<Elevator> cars, int floor, Direction dir);
}

final class Elevator {
    Elevator(int id) { /* TODO */ }

    int id() { throw new UnsupportedOperationException("TODO"); }
    int floor() { throw new UnsupportedOperationException("TODO"); }
    Direction direction() { throw new UnsupportedOperationException("TODO"); }
    /** sorted ascending */
    List<Integer> stops() { throw new UnsupportedOperationException("TODO"); }
    void addStop(int floor) { throw new UnsupportedOperationException("TODO"); }
    /** @return the floor where doors opened this tick, if any */
    Optional<Integer> step() { throw new UnsupportedOperationException("TODO"); }
}

final class NearestCarStrategy implements DispatchStrategy {
    public Elevator choose(List<Elevator> cars, int floor, Direction dir) { throw new UnsupportedOperationException("TODO"); }
}

final class ElevatorSystem {
    ElevatorSystem(int floors, int cars, DispatchStrategy strategy) { /* TODO */ }

    int requestPickup(int floor, Direction dir) { throw new UnsupportedOperationException("TODO"); }
    void requestDestination(int carId, int floor) { throw new UnsupportedOperationException("TODO"); }
    List<String> step() { throw new UnsupportedOperationException("TODO"); }
    Elevator car(int id) { throw new UnsupportedOperationException("TODO"); }
}
