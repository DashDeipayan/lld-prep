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

// ======================= Your design =======================
final class NearestDriverStrategy implements MatchingStrategy {
    NearestDriverStrategy(int maxDistance) { /* TODO */ }
    public Optional<String> match(Location pickup, List<DriverInfo> candidates) { throw new UnsupportedOperationException("TODO"); }
}

final class StandardPricing implements PricingStrategy {
    StandardPricing(Map<VehicleClass, Rate> rates, long surgePercent) { /* TODO */ }
    public long fare(VehicleClass cls, int distance, long durationMillis) { throw new UnsupportedOperationException("TODO"); }
}

final class RideService {
    RideService(MatchingStrategy matching, PricingStrategy pricing, Clock clock) { /* TODO */ }

    void addDriver(String id, VehicleClass cls, Location location, double rating) { throw new UnsupportedOperationException("TODO"); }
    void setOnline(String driverId, boolean online) { throw new UnsupportedOperationException("TODO"); }
    void updateLocation(String driverId, Location location) { throw new UnsupportedOperationException("TODO"); }
    String requestRide(String riderId, Location from, Location to, VehicleClass cls) { throw new UnsupportedOperationException("TODO"); }
    String driverOf(String tripId) { throw new UnsupportedOperationException("TODO"); }
    TripStatus status(String tripId) { throw new UnsupportedOperationException("TODO"); }
    void arrived(String tripId) { throw new UnsupportedOperationException("TODO"); }
    void start(String tripId) { throw new UnsupportedOperationException("TODO"); }
    long end(String tripId) { throw new UnsupportedOperationException("TODO"); }
    void cancel(String tripId) { throw new UnsupportedOperationException("TODO"); }
}
