import java.util.*;

enum RoomType { SINGLE, DOUBLE, SUITE }

final class NoRoomAvailableException extends RuntimeException {
    NoRoomAvailableException(String m) { super(m); }
}

final class Hotel {
    private record Room(String number, RoomType type, long rate) {}

    private enum Status { BOOKED, CHECKED_IN, CHECKED_OUT, CANCELLED }

    private static final class Booking {
        final String id, guest;
        final Room room;
        final int in, out;
        Status status = Status.BOOKED;
        long charges;
        Booking(String id, String guest, Room room, int in, int out) { this.id = id; this.guest = guest; this.room = room; this.in = in; this.out = out; }
        boolean blocks(int s, int e) { return (status == Status.BOOKED || status == Status.CHECKED_IN) && s < out && in < e; }
    }

    private final TreeMap<String, Room> rooms = new TreeMap<>();
    private final Map<String, Booking> bookings = new LinkedHashMap<>();
    private int next = 1;

    synchronized void addRoom(String number, RoomType type, long nightlyRate) {
        if (rooms.containsKey(number) || nightlyRate <= 0) throw new IllegalArgumentException("invalid room " + number);
        rooms.put(number, new Room(number, type, nightlyRate));
    }

    synchronized Map<RoomType, Integer> availability(int checkIn, int checkOut) {
        validate(checkIn, checkOut);
        Map<RoomType, Integer> out = new EnumMap<>(RoomType.class);
        for (RoomType t : RoomType.values()) out.put(t, 0);
        for (Room r : rooms.values()) if (free(r, checkIn, checkOut)) out.merge(r.type(), 1, Integer::sum);
        return out;
    }

    synchronized String book(String guest, RoomType type, int checkIn, int checkOut) {
        validate(checkIn, checkOut);
        for (Room r : rooms.values()) {                       // TreeMap => lowest number first
            if (r.type() == type && free(r, checkIn, checkOut)) {
                Booking b = new Booking("BK" + next++, guest, r, checkIn, checkOut);
                bookings.put(b.id, b);
                return b.id;
            }
        }
        throw new NoRoomAvailableException("no " + type + " for " + checkIn + "-" + checkOut);
    }

    synchronized String roomOf(String bookingId) { return booking(bookingId).room.number(); }

    synchronized void checkIn(String bookingId, int day) {
        Booking b = booking(bookingId);
        require(b, Status.BOOKED);
        if (day != b.in) throw new IllegalStateException("check-in is on day " + b.in);
        b.status = Status.CHECKED_IN;
    }

    synchronized void addCharge(String bookingId, long amount) {
        Booking b = booking(bookingId);
        require(b, Status.CHECKED_IN);
        if (amount <= 0) throw new IllegalArgumentException("amount must be positive");
        b.charges += amount;
    }

    synchronized long checkOut(String bookingId) {
        Booking b = booking(bookingId);
        require(b, Status.CHECKED_IN);
        b.status = Status.CHECKED_OUT;
        return (long) (b.out - b.in) * b.room.rate() + b.charges;
    }

    synchronized void cancel(String bookingId) {
        Booking b = booking(bookingId);
        require(b, Status.BOOKED);
        b.status = Status.CANCELLED;
    }

    synchronized String status(String bookingId) { return booking(bookingId).status.name(); }

    private boolean free(Room r, int s, int e) {
        for (Booking b : bookings.values()) if (b.room == r && b.blocks(s, e)) return false;
        return true;
    }

    private static void require(Booking b, Status s) {
        if (b.status != s) throw new IllegalStateException("booking is " + b.status + ", expected " + s);
    }

    private Booking booking(String id) {
        Booking b = bookings.get(id);
        if (b == null) throw new IllegalArgumentException("unknown booking " + id);
        return b;
    }

    private static void validate(int in, int out) {
        if (in < 0 || in >= out) throw new IllegalArgumentException("invalid stay");
    }
}
