import java.util.*;

// ======================= Provided =======================
enum VehicleClass { MINI, SEDAN, SUV }

enum TripStatus { ASSIGNED, ARRIVED, IN_PROGRESS, COMPLETED, CANCELLED }

record Location(int x, int y) {
    int distance(Location o) { return Math.abs(x - o.x) + Math.abs(y - o.y); }
}

record DriverInfo(String id, VehicleClass cls, Location location, double rating) {}

record Rate(long base, long perKm, long perMin) {}

interface MatchingStrategy { Optional<String> match(Location pickup, List<DriverInfo> candidates); }

interface PricingStrategy { long fare(VehicleClass cls, int distance, long durationMillis); }

interface Clock { long nowMillis(); }

final class NoDriverAvailableException extends RuntimeException {
    NoDriverAvailableException(String m) { super(m); }
}

// ======================= Strategies =======================
final class NearestDriverStrategy implements MatchingStrategy {
    private final int maxDistance;
    NearestDriverStrategy(int maxDistance) { this.maxDistance = maxDistance; }

    public Optional<String> match(Location pickup, List<DriverInfo> candidates) {
        return candidates.stream()
                .filter(d -> d.location().distance(pickup) <= maxDistance)
                .min(Comparator.comparingInt((DriverInfo d) -> d.location().distance(pickup))
                        .thenComparing(DriverInfo::rating, Comparator.reverseOrder())
                        .thenComparing(DriverInfo::id))
                .map(DriverInfo::id);
    }
}

final class StandardPricing implements PricingStrategy {
    private final Map<VehicleClass, Rate> rates;
    private final long surgePercent;
    StandardPricing(Map<VehicleClass, Rate> rates, long surgePercent) { this.rates = Map.copyOf(rates); this.surgePercent = surgePercent; }

    public long fare(VehicleClass cls, int distance, long durationMillis) {
        Rate r = rates.get(cls);
        long minutes = (durationMillis + 59_999) / 60_000;
        return (r.base() + r.perKm() * distance + r.perMin() * minutes) * surgePercent / 100;
    }
}

// ======================= Model =======================
final class Driver {
    final String id;
    final VehicleClass cls;
    final double rating;
    Location location;
    boolean online, busy;
    Driver(String id, VehicleClass cls, Location location, double rating) { this.id = id; this.cls = cls; this.location = location; this.rating = rating; }
    DriverInfo info() { return new DriverInfo(id, cls, location, rating); }
}

final class Trip {
    final String id, rider;
    final Driver driver;
    final Location from, to;
    final VehicleClass cls;
    TripStatus status = TripStatus.ASSIGNED;
    long startedAt;
    Trip(String id, String rider, Driver driver, Location from, Location to, VehicleClass cls) {
        this.id = id; this.rider = rider; this.driver = driver; this.from = from; this.to = to; this.cls = cls;
    }
}

// ======================= Service =======================
final class RideService {
    private static final Map<TripStatus, Set<TripStatus>> ALLOWED = Map.of(
            TripStatus.ASSIGNED, EnumSet.of(TripStatus.ARRIVED, TripStatus.CANCELLED),
            TripStatus.ARRIVED, EnumSet.of(TripStatus.IN_PROGRESS, TripStatus.CANCELLED),
            TripStatus.IN_PROGRESS, EnumSet.of(TripStatus.COMPLETED),
            TripStatus.COMPLETED, EnumSet.noneOf(TripStatus.class),
            TripStatus.CANCELLED, EnumSet.noneOf(TripStatus.class));

    private final MatchingStrategy matching;
    private final PricingStrategy pricing;
    private final Clock clock;
    private final Map<String, Driver> drivers = new TreeMap<>();
    private final Map<String, Trip> trips = new HashMap<>();
    private final Map<String, String> activeTripByRider = new HashMap<>();
    private int nextTrip = 1;

    RideService(MatchingStrategy matching, PricingStrategy pricing, Clock clock) {
        this.matching = matching;
        this.pricing = pricing;
        this.clock = clock;
    }

    synchronized void addDriver(String id, VehicleClass cls, Location location, double rating) {
        if (drivers.containsKey(id)) throw new IllegalArgumentException("duplicate driver");
        drivers.put(id, new Driver(id, cls, location, rating));
    }

    synchronized void setOnline(String driverId, boolean online) {
        Driver d = driver(driverId);
        if (!online && d.busy) throw new IllegalStateException("driver is on a trip");
        d.online = online;
    }

    synchronized void updateLocation(String driverId, Location location) { driver(driverId).location = location; }

    /** Matching + marking the driver busy happen under one lock: no double assignment. */
    synchronized String requestRide(String riderId, Location from, Location to, VehicleClass cls) {
        if (activeTripByRider.containsKey(riderId)) throw new IllegalStateException("rider already has an active trip");
        List<DriverInfo> candidates = new ArrayList<>();
        for (Driver d : drivers.values()) if (d.online && !d.busy && d.cls == cls) candidates.add(d.info());
        String chosen = matching.match(from, candidates).orElseThrow(() -> new NoDriverAvailableException("no " + cls + " nearby"));
        Driver d = driver(chosen);
        d.busy = true;
        Trip t = new Trip("R" + nextTrip++, riderId, d, from, to, cls);
        trips.put(t.id, t);
        activeTripByRider.put(riderId, t.id);
        return t.id;
    }

    synchronized String driverOf(String tripId) { return trip(tripId).driver.id; }
    synchronized TripStatus status(String tripId) { return trip(tripId).status; }

    synchronized void arrived(String tripId) { transition(trip(tripId), TripStatus.ARRIVED); }

    synchronized void start(String tripId) {
        Trip t = trip(tripId);
        transition(t, TripStatus.IN_PROGRESS);
        t.startedAt = clock.nowMillis();
    }

    synchronized long end(String tripId) {
        Trip t = trip(tripId);
        transition(t, TripStatus.COMPLETED);
        t.driver.location = t.to;
        return pricing.fare(t.cls, t.from.distance(t.to), clock.nowMillis() - t.startedAt);
    }

    synchronized void cancel(String tripId) { transition(trip(tripId), TripStatus.CANCELLED); }

    private void transition(Trip t, TripStatus to) {
        if (!ALLOWED.get(t.status).contains(to)) throw new IllegalStateException(t.status + " -> " + to + " not allowed");
        t.status = to;
        if (to == TripStatus.COMPLETED || to == TripStatus.CANCELLED) {
            t.driver.busy = false;
            activeTripByRider.remove(t.rider);
        }
    }

    private Driver driver(String id) {
        Driver d = drivers.get(id);
        if (d == null) throw new IllegalArgumentException("unknown driver " + id);
        return d;
    }

    private Trip trip(String id) {
        Trip t = trips.get(id);
        if (t == null) throw new IllegalArgumentException("unknown trip " + id);
        return t;
    }
}
