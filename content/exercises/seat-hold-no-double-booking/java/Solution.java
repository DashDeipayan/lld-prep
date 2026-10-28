import java.util.*;

interface Clock {
    long nowMillis();
}

final class SeatUnavailableException extends RuntimeException {
    SeatUnavailableException(String m) { super(m); }
}

final class SeatInventory {
    SeatInventory(List<String> seatIds, long holdTtlMillis, Clock clock) {
        // TODO
    }

    String hold(String userId, List<String> seatIds) { throw new UnsupportedOperationException("TODO"); }

    void confirm(String holdId) { throw new UnsupportedOperationException("TODO"); }

    void release(String holdId) { throw new UnsupportedOperationException("TODO"); }

    String status(String seatId) { throw new UnsupportedOperationException("TODO"); }

    List<String> available() { throw new UnsupportedOperationException("TODO"); }
}
