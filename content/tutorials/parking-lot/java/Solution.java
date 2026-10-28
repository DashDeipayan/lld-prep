import java.util.*;

// ======================= Provided (do not change) =======================
enum SpotSize { SMALL, MEDIUM, LARGE }

enum VehicleType {
    BIKE(SpotSize.SMALL), CAR(SpotSize.MEDIUM), TRUCK(SpotSize.LARGE);
    private final SpotSize minSize;
    VehicleType(SpotSize minSize) { this.minSize = minSize; }
    SpotSize minSize() { return minSize; }
}

record Vehicle(String plate, VehicleType type) {}
record Ticket(String id, String plate, String spotId, long entryMillis) {}
record Receipt(String ticketId, String plate, String spotId, long minutes, long fee) {}
record Event(String kind, String ticketId, String spotId, SpotSize size) {}

interface Clock { long nowMillis(); }
interface PricingStrategy { long fee(VehicleType type, long durationMillis); }
interface SpotSelector { Spot select(List<Spot> candidates, Vehicle v); }
interface ParkingListener { void onEvent(Event e); }

final class ParkingFullException extends RuntimeException {
    ParkingFullException(String m) { super(m); }
}

// ============================ Step 1 ============================
final class Spot {
    Spot(String id, int floor, SpotSize size) { /* TODO */ }
    String id() { throw new UnsupportedOperationException("TODO"); }
    int floor() { throw new UnsupportedOperationException("TODO"); }
    SpotSize size() { throw new UnsupportedOperationException("TODO"); }
    boolean canFit(Vehicle v) { throw new UnsupportedOperationException("TODO"); }
    boolean isFree() { throw new UnsupportedOperationException("TODO"); }
    void occupy(Vehicle v) { throw new UnsupportedOperationException("TODO"); }
    Vehicle vacate() { throw new UnsupportedOperationException("TODO"); }
}

// ============================ Step 3 ============================
final class NearestSpotSelector implements SpotSelector {
    public Spot select(List<Spot> candidates, Vehicle v) { throw new UnsupportedOperationException("TODO"); }
}

// ============================ Step 4 ============================
final class HourlyPricing implements PricingStrategy {
    HourlyPricing(Map<VehicleType, Long> ratePerHour) { /* TODO */ }
    public long fee(VehicleType type, long durationMillis) { throw new UnsupportedOperationException("TODO"); }
}

// ======================= Steps 2, 3, 4, 5, 6 =======================
final class ParkingLot {
    ParkingLot(PricingStrategy pricing, SpotSelector selector, Clock clock) { /* TODO */ }

    void addFloor(int floor, int small, int medium, int large) { throw new UnsupportedOperationException("TODO"); }
    int available(SpotSize size) { throw new UnsupportedOperationException("TODO"); }
    Ticket park(Vehicle v) { throw new UnsupportedOperationException("TODO"); }
    Receipt unpark(String ticketId) { throw new UnsupportedOperationException("TODO"); }
    void addListener(ParkingListener l) { throw new UnsupportedOperationException("TODO"); }
}

// ============================ Step 5 ============================
final class DisplayBoard implements ParkingListener {
    DisplayBoard(ParkingLot lot) { /* TODO */ }
    public void onEvent(Event e) { throw new UnsupportedOperationException("TODO"); }
    int free(SpotSize size) { throw new UnsupportedOperationException("TODO"); }
}
