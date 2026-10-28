import java.util.*;

// ======================= Provided =======================
record Booking(String id, String room, String organizer, int start, int end) {}

record Slot(int start, int end) {}

final class NoRoomAvailableException extends RuntimeException {
    NoRoomAvailableException(String m) { super(m); }
}

// ======================= Your design =======================
final class Scheduler {
    void addRoom(String id, int capacity) { throw new UnsupportedOperationException("TODO"); }
    Booking book(String organizer, int start, int end, int attendees) { throw new UnsupportedOperationException("TODO"); }
    Booking bookRoom(String roomId, String organizer, int start, int end) { throw new UnsupportedOperationException("TODO"); }
    void cancel(String bookingId) { throw new UnsupportedOperationException("TODO"); }
    List<String> freeRooms(int start, int end, int attendees) { throw new UnsupportedOperationException("TODO"); }
    List<Booking> schedule(String roomId) { throw new UnsupportedOperationException("TODO"); }
    List<Slot> freeSlots(String roomId, int from, int to) { throw new UnsupportedOperationException("TODO"); }
}
