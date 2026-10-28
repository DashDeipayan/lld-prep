import java.util.*;

enum CarType { HATCHBACK, SEDAN, SUV }

record Reservation(String id, String customer, String carId, int startDay, int endDay, long cost, String status) {}

final class RentalService {
    private record Branch(String id, String city) {}

    private static final class Car {
        final String id;
        final Branch branch;
        final CarType type;
        final long rate;
        final List<Reservation> active = new ArrayList<>();
        Car(String id, Branch branch, CarType type, long rate) { this.id = id; this.branch = branch; this.type = type; this.rate = rate; }

        boolean free(int s, int e) {
            for (Reservation r : active) if (s < r.endDay() && r.startDay() < e) return false;
            return true;
        }
    }

    private final Map<String, Branch> branches = new HashMap<>();
    private final Map<String, Car> cars = new HashMap<>();
    private final Map<String, Reservation> reservations = new HashMap<>();
    private int next = 1;

    synchronized void addBranch(String branchId, String city) {
        if (branches.putIfAbsent(branchId, new Branch(branchId, city)) != null) throw new IllegalArgumentException("duplicate branch");
    }

    synchronized void addCar(String branchId, String carId, CarType type, long dailyRate) {
        Branch b = branches.get(branchId);
        if (b == null || cars.containsKey(carId) || dailyRate <= 0) throw new IllegalArgumentException("invalid car");
        cars.put(carId, new Car(carId, b, type, dailyRate));
    }

    synchronized List<String> search(String city, CarType type, int startDay, int endDay) {
        validate(startDay, endDay);
        return cars.values().stream()
                .filter(c -> c.type == type && c.branch.city().equalsIgnoreCase(city) && c.free(startDay, endDay))
                .sorted(Comparator.comparingLong((Car c) -> c.rate).thenComparing(c -> c.id))
                .map(c -> c.id).toList();
    }

    synchronized Reservation reserve(String customer, String carId, int startDay, int endDay) {
        validate(startDay, endDay);
        Car c = car(carId);
        if (!c.free(startDay, endDay)) throw new IllegalStateException("car " + carId + " is booked for those dates");
        Reservation r = new Reservation("RS" + next++, customer, carId, startDay, endDay, (endDay - startDay) * c.rate, "RESERVED");
        c.active.add(r);
        reservations.put(r.id(), r);
        return r;
    }

    synchronized Reservation cancel(String reservationId) {
        Reservation r = active(reservationId);
        return finish(r, "CANCELLED", r.cost());
    }

    synchronized long returnCar(String reservationId, int actualDay) {
        Reservation r = active(reservationId);
        Car c = car(r.carId());
        long late = Math.max(0, actualDay - r.endDay());
        long amount = r.cost() + late * c.rate * 3 / 2;
        finish(r, "COMPLETED", amount);
        return amount;
    }

    private Reservation finish(Reservation r, String status, long amount) {
        Car c = car(r.carId());
        c.active.remove(r);
        Reservation done = new Reservation(r.id(), r.customer(), r.carId(), r.startDay(), r.endDay(), amount, status);
        reservations.put(r.id(), done);
        return done;
    }

    private Reservation active(String id) {
        Reservation r = reservations.get(id);
        if (r == null) throw new IllegalArgumentException("unknown reservation " + id);
        if (!r.status().equals("RESERVED")) throw new IllegalStateException("reservation is " + r.status());
        return r;
    }

    private Car car(String id) {
        Car c = cars.get(id);
        if (c == null) throw new IllegalArgumentException("unknown car " + id);
        return c;
    }

    private static void validate(int s, int e) {
        if (s < 0 || s >= e) throw new IllegalArgumentException("invalid range");
    }
}
