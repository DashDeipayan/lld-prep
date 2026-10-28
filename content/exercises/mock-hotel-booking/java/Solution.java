import java.util.*;

enum RoomType { SINGLE, DOUBLE, SUITE }

final class NoRoomAvailableException extends RuntimeException {
    NoRoomAvailableException(String m) { super(m); }
}

final class Hotel {
    void addRoom(String number, RoomType type, long nightlyRate) { throw new UnsupportedOperationException("TODO"); }
    Map<RoomType, Integer> availability(int checkIn, int checkOut) { throw new UnsupportedOperationException("TODO"); }
    String book(String guest, RoomType type, int checkIn, int checkOut) { throw new UnsupportedOperationException("TODO"); }
    String roomOf(String bookingId) { throw new UnsupportedOperationException("TODO"); }
    void checkIn(String bookingId, int day) { throw new UnsupportedOperationException("TODO"); }
    void addCharge(String bookingId, long amount) { throw new UnsupportedOperationException("TODO"); }
    long checkOut(String bookingId) { throw new UnsupportedOperationException("TODO"); }
    void cancel(String bookingId) { throw new UnsupportedOperationException("TODO"); }
    String status(String bookingId) { throw new UnsupportedOperationException("TODO"); }
}
