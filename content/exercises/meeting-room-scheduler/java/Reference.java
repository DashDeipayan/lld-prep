import java.util.*;

// ======================= Provided =======================
record Booking(String id, String room, String organizer, int start, int end) {}

record Slot(int start, int end) {}

final class NoRoomAvailableException extends RuntimeException {
    NoRoomAvailableException(String m) { super(m); }
}

// ======================= Model =======================
final class Room {
    final String id;
    final int capacity;
    final TreeMap<Integer, Booking> byStart = new TreeMap<>();
    Room(String id, int capacity) { this.id = id; this.capacity = capacity; }

    /** Only the neighbours in start order can overlap [s, e). */
    boolean isFree(int s, int e) {
        Map.Entry<Integer, Booking> before = byStart.lowerEntry(e);   // latest booking starting before e
        return before == null || before.getValue().end() <= s;
    }
}

final class Scheduler {
    private static final Comparator<Room> BEST_FIT = Comparator.comparingInt((Room r) -> r.capacity).thenComparing(r -> r.id);
    private final Map<String, Room> rooms = new HashMap<>();
    private final Map<String, Booking> bookings = new HashMap<>();
    private int next = 1;

    synchronized void addRoom(String id, int capacity) {
        if (capacity < 1 || rooms.containsKey(id)) throw new IllegalArgumentException("invalid room " + id);
        rooms.put(id, new Room(id, capacity));
    }

    synchronized Booking book(String organizer, int start, int end, int attendees) {
        validate(start, end);
        if (attendees < 1) throw new IllegalArgumentException("attendees must be >= 1");
        Room r = rooms.values().stream().filter(x -> x.capacity >= attendees && x.isFree(start, end)).min(BEST_FIT)
                .orElseThrow(() -> new NoRoomAvailableException("no room for " + attendees + " at " + start));
        return place(r, organizer, start, end);
    }

    synchronized Booking bookRoom(String roomId, String organizer, int start, int end) {
        validate(start, end);
        Room r = room(roomId);
        if (!r.isFree(start, end)) throw new IllegalStateException("conflict in " + roomId);
        return place(r, organizer, start, end);
    }

    synchronized void cancel(String bookingId) {
        Booking b = bookings.remove(bookingId);
        if (b == null) throw new IllegalStateException("unknown or cancelled booking " + bookingId);
        rooms.get(b.room()).byStart.remove(b.start());
    }

    synchronized List<String> freeRooms(int start, int end, int attendees) {
        validate(start, end);
        return rooms.values().stream().filter(r -> r.capacity >= attendees && r.isFree(start, end)).sorted(BEST_FIT).map(r -> r.id).toList();
    }

    synchronized List<Booking> schedule(String roomId) { return List.copyOf(room(roomId).byStart.values()); }

    synchronized List<Slot> freeSlots(String roomId, int from, int to) {
        validate(from, to);
        List<Slot> out = new ArrayList<>();
        int cursor = from;
        for (Booking b : room(roomId).byStart.values()) {
            if (b.end() <= cursor) continue;
            if (b.start() >= to) break;
            if (b.start() > cursor) out.add(new Slot(cursor, b.start()));
            cursor = Math.max(cursor, b.end());
        }
        if (cursor < to) out.add(new Slot(cursor, to));
        return out;
    }

    private Booking place(Room r, String organizer, int start, int end) {
        Booking b = new Booking("M" + next++, r.id, organizer, start, end);
        r.byStart.put(start, b);
        bookings.put(b.id(), b);
        return b;
    }

    private Room room(String id) {
        Room r = rooms.get(id);
        if (r == null) throw new IllegalArgumentException("unknown room " + id);
        return r;
    }

    private static void validate(int start, int end) {
        if (start < 0 || end > 1440 || start >= end) throw new IllegalArgumentException("invalid interval [" + start + "," + end + ")");
    }
}
