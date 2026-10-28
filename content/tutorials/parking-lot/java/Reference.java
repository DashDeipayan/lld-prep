import java.util.*;
import java.util.concurrent.CopyOnWriteArrayList;

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
    private final String id;
    private final int floor;
    private final SpotSize size;
    private Vehicle vehicle;

    Spot(String id, int floor, SpotSize size) {
        this.id = id;
        this.floor = floor;
        this.size = size;
    }

    String id() { return id; }
    int floor() { return floor; }
    SpotSize size() { return size; }
    boolean canFit(Vehicle v) { return size.compareTo(v.type().minSize()) >= 0; }
    boolean isFree() { return vehicle == null; }

    void occupy(Vehicle v) {
        if (!isFree()) throw new IllegalStateException(id + " is occupied");
        if (!canFit(v)) throw new IllegalStateException(v.type() + " does not fit " + size);
        vehicle = v;
    }

    Vehicle vacate() {
        Vehicle v = vehicle;
        vehicle = null;
        return v;
    }
}

// ============================ Step 3 ============================
final class NearestSpotSelector implements SpotSelector {
    public Spot select(List<Spot> candidates, Vehicle v) {
        return candidates.isEmpty() ? null : candidates.get(0); // candidates arrive in lot order
    }
}

// ============================ Step 4 ============================
final class HourlyPricing implements PricingStrategy {
    private final Map<VehicleType, Long> rate;

    HourlyPricing(Map<VehicleType, Long> ratePerHour) { this.rate = new EnumMap<>(ratePerHour); }

    public long fee(VehicleType type, long durationMillis) {
        long hours = Math.max(1, (durationMillis + 3_599_999) / 3_600_000);
        return hours * rate.getOrDefault(type, 0L);
    }
}

// ======================= Steps 2, 3, 4, 5, 6 =======================
final class ParkingLot {
    private final PricingStrategy pricing;
    private final SpotSelector selector;
    private final Clock clock;
    private final List<Spot> spots = new ArrayList<>();             // lot order
    private final Map<String, Spot> spotsById = new HashMap<>();
    private final Set<Integer> floors = new HashSet<>();
    private final Map<String, Ticket> activeTickets = new HashMap<>();
    private final Set<String> parkedPlates = new HashSet<>();
    private final List<ParkingListener> listeners = new CopyOnWriteArrayList<>();
    private long nextTicket = 1;

    ParkingLot(PricingStrategy pricing, SpotSelector selector, Clock clock) {
        this.pricing = pricing;
        this.selector = selector;
        this.clock = clock;
    }

    synchronized void addFloor(int floor, int small, int medium, int large) {
        if (!floors.add(floor)) throw new IllegalArgumentException("duplicate floor " + floor);
        addSpots(floor, SpotSize.SMALL, "S", small);
        addSpots(floor, SpotSize.MEDIUM, "M", medium);
        addSpots(floor, SpotSize.LARGE, "L", large);
    }

    private void addSpots(int floor, SpotSize size, String code, int count) {
        for (int i = 1; i <= count; i++) {
            Spot s = new Spot("F" + floor + "-" + code + i, floor, size);
            spots.add(s);
            spotsById.put(s.id(), s);
        }
    }

    synchronized int available(SpotSize size) {
        int n = 0;
        for (Spot s : spots) if (s.size() == size && s.isFree()) n++;
        return n;
    }

    Ticket park(Vehicle v) {
        Ticket t;
        Spot spot;
        synchronized (this) {
            if (parkedPlates.contains(v.plate())) throw new IllegalStateException(v.plate() + " already parked");
            List<Spot> candidates = new ArrayList<>();
            for (Spot s : spots) if (s.isFree() && s.canFit(v)) candidates.add(s);
            spot = selector.select(candidates, v);
            if (spot == null) throw new ParkingFullException("no spot for " + v.type());
            spot.occupy(v);
            t = new Ticket("T" + nextTicket++, v.plate(), spot.id(), clock.nowMillis());
            activeTickets.put(t.id(), t);
            parkedPlates.add(v.plate());
        }
        emit(new Event("PARKED", t.id(), spot.id(), spot.size()));
        return t;
    }

    Receipt unpark(String ticketId) {
        Receipt r;
        Spot spot;
        synchronized (this) {
            Ticket t = activeTickets.remove(ticketId);
            if (t == null) throw new IllegalArgumentException("unknown or used ticket " + ticketId);
            spot = spotsById.get(t.spotId());
            Vehicle v = spot.vacate();
            parkedPlates.remove(v.plate());
            long duration = clock.nowMillis() - t.entryMillis();
            r = new Receipt(t.id(), t.plate(), spot.id(), duration / 60_000, pricing.fee(v.type(), duration));
        }
        emit(new Event("LEFT", ticketId, spot.id(), spot.size()));
        return r;
    }

    void addListener(ParkingListener l) { listeners.add(l); }

    private void emit(Event e) {
        for (ParkingListener l : listeners) l.onEvent(e);
    }
}

// ============================ Step 5 ============================
final class DisplayBoard implements ParkingListener {
    private final Map<SpotSize, Integer> free = new EnumMap<>(SpotSize.class);

    DisplayBoard(ParkingLot lot) {
        for (SpotSize s : SpotSize.values()) free.put(s, lot.available(s));
        lot.addListener(this);
    }

    public synchronized void onEvent(Event e) {
        free.merge(e.size(), e.kind().equals("PARKED") ? -1 : 1, Integer::sum);
    }

    synchronized int free(SpotSize size) { return free.get(size); }
}
