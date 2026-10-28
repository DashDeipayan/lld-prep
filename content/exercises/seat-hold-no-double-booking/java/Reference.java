import java.util.*;

interface Clock {
    long nowMillis();
}

final class SeatUnavailableException extends RuntimeException {
    SeatUnavailableException(String m) { super(m); }
}

final class SeatInventory {
    private enum Status { AVAILABLE, HELD, BOOKED }

    private static final class Seat {
        Status status = Status.AVAILABLE;
        String holdId;
        long expiresAt;
    }

    private static final class Hold {
        final List<String> seats;
        final long expiresAt;
        boolean active = true;
        Hold(List<String> seats, long expiresAt) { this.seats = seats; this.expiresAt = expiresAt; }
    }

    private final Map<String, Seat> seats = new TreeMap<>();
    private final Map<String, Hold> holds = new HashMap<>();
    private final long ttl;
    private final Clock clock;
    private long nextHold = 1;

    SeatInventory(List<String> seatIds, long holdTtlMillis, Clock clock) {
        for (String id : seatIds) seats.put(id, new Seat());
        this.ttl = holdTtlMillis;
        this.clock = clock;
    }

    synchronized String hold(String userId, List<String> seatIds) {
        if (seatIds == null || seatIds.isEmpty()) throw new IllegalArgumentException("no seats");
        if (new HashSet<>(seatIds).size() != seatIds.size()) throw new IllegalArgumentException("duplicate seats");
        long now = clock.nowMillis();
        for (String id : seatIds) {                         // 1. validate everything
            Seat s = seats.get(id);
            if (s == null) throw new IllegalArgumentException("unknown seat " + id);
            if (!isAvailable(s, now)) throw new SeatUnavailableException(id);
        }
        String holdId = "H" + nextHold++;                   // 2. then mutate everything
        long expires = now + ttl;
        for (String id : seatIds) {
            Seat s = seats.get(id);
            s.status = Status.HELD;
            s.holdId = holdId;
            s.expiresAt = expires;
        }
        holds.put(holdId, new Hold(List.copyOf(seatIds), expires));
        return holdId;
    }

    synchronized void confirm(String holdId) {
        Hold h = activeHold(holdId);
        for (String id : h.seats) {
            Seat s = seats.get(id);
            s.status = Status.BOOKED;
            s.holdId = holdId;
        }
        h.active = false;
    }

    synchronized void release(String holdId) {
        Hold h = activeHold(holdId);
        for (String id : h.seats) {
            Seat s = seats.get(id);
            s.status = Status.AVAILABLE;
            s.holdId = null;
        }
        h.active = false;
    }

    synchronized String status(String seatId) {
        Seat s = seats.get(seatId);
        if (s == null) throw new IllegalArgumentException("unknown seat " + seatId);
        if (s.status == Status.HELD && isAvailable(s, clock.nowMillis())) return "AVAILABLE";
        return s.status.name();
    }

    synchronized List<String> available() {
        long now = clock.nowMillis();
        List<String> out = new ArrayList<>();
        for (Map.Entry<String, Seat> e : seats.entrySet()) if (isAvailable(e.getValue(), now)) out.add(e.getKey());
        return out;
    }

    private Hold activeHold(String holdId) {
        Hold h = holds.get(holdId);
        if (h == null || !h.active || h.expiresAt <= clock.nowMillis()) throw new IllegalStateException("hold not active: " + holdId);
        return h;
    }

    private static boolean isAvailable(Seat s, long now) {
        return s.status == Status.AVAILABLE || (s.status == Status.HELD && s.expiresAt <= now);
    }
}
