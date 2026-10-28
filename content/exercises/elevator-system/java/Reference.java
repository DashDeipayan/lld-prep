import java.util.*;

enum Direction { UP, DOWN, IDLE }

interface DispatchStrategy {
    Elevator choose(List<Elevator> cars, int floor, Direction dir);
}

final class Elevator {
    private final int id;
    private int floor;
    private Direction direction = Direction.IDLE;
    private final TreeSet<Integer> stops = new TreeSet<>();

    Elevator(int id) { this.id = id; }

    int id() { return id; }
    int floor() { return floor; }
    Direction direction() { return direction; }
    List<Integer> stops() { return new ArrayList<>(stops); }
    void addStop(int f) { stops.add(f); }

    Optional<Integer> step() {
        if (stops.isEmpty()) {                                     // 1
            direction = Direction.IDLE;
            return Optional.empty();
        }
        if (stops.contains(floor)) return Optional.of(arrive());   // 2
        if (direction == Direction.IDLE) {                         // 3
            Integer up = stops.higher(floor), down = stops.lower(floor);
            direction = down == null || (up != null && up - floor <= floor - down) ? Direction.UP : Direction.DOWN;
        }
        if (direction == Direction.UP && stops.higher(floor) == null) direction = Direction.DOWN;   // 4
        else if (direction == Direction.DOWN && stops.lower(floor) == null) direction = Direction.UP;
        floor += direction == Direction.UP ? 1 : -1;               // 5
        return stops.contains(floor) ? Optional.of(arrive()) : Optional.empty();
    }

    private int arrive() {
        stops.remove(floor);
        if (stops.isEmpty()) direction = Direction.IDLE;
        return floor;
    }
}

final class NearestCarStrategy implements DispatchStrategy {
    public Elevator choose(List<Elevator> cars, int floor, Direction dir) {
        Elevator best = null;
        int bestCost = Integer.MAX_VALUE;
        for (Elevator c : cars) {
            int cost = cost(c, floor, dir);
            if (cost < bestCost) { best = c; bestCost = cost; }
        }
        return best;
    }

    static int cost(Elevator c, int floor, Direction dir) {
        int cur = c.floor();
        if (c.direction() == Direction.IDLE) return Math.abs(cur - floor);
        if (c.direction() == Direction.UP && dir == Direction.UP && floor >= cur) return floor - cur;
        if (c.direction() == Direction.DOWN && dir == Direction.DOWN && floor <= cur) return cur - floor;
        List<Integer> s = c.stops();
        int turn = s.isEmpty() ? cur : (c.direction() == Direction.UP ? s.get(s.size() - 1) : s.get(0));
        return Math.abs(cur - turn) + Math.abs(turn - floor);
    }
}

final class ElevatorSystem {
    private final int floors;
    private final List<Elevator> cars = new ArrayList<>();
    private final DispatchStrategy strategy;

    ElevatorSystem(int floors, int cars, DispatchStrategy strategy) {
        if (floors < 2 || cars < 1) throw new IllegalArgumentException("need >= 2 floors and >= 1 car");
        this.floors = floors;
        this.strategy = strategy;
        for (int i = 0; i < cars; i++) this.cars.add(new Elevator(i));
    }

    synchronized int requestPickup(int floor, Direction dir) {
        checkFloor(floor);
        if (dir == Direction.IDLE) throw new IllegalArgumentException("pickup direction must be UP or DOWN");
        Elevator e = strategy.choose(Collections.unmodifiableList(cars), floor, dir);
        e.addStop(floor);
        return e.id();
    }

    synchronized void requestDestination(int carId, int floor) {
        checkFloor(floor);
        car(carId).addStop(floor);
    }

    synchronized List<String> step() {
        List<String> events = new ArrayList<>();
        for (Elevator e : cars) e.step().ifPresent(f -> events.add("E" + e.id() + "@" + f));
        return events;
    }

    synchronized Elevator car(int id) {
        if (id < 0 || id >= cars.size()) throw new IllegalArgumentException("no elevator " + id);
        return cars.get(id);
    }

    private void checkFloor(int f) {
        if (f < 0 || f >= floors) throw new IllegalArgumentException("invalid floor " + f);
    }
}
