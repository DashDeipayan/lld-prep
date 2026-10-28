import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

// ======================= Provided =======================
interface PaymentGateway { boolean charge(String user, long amount); }

interface Clock { long nowMillis(); }

record Booking(String id, String showId, String user, List<String> seats, long amount, String status) {}

final class SeatUnavailableException extends RuntimeException {
    SeatUnavailableException(String m) { super(m); }
}

final class PaymentFailedException extends RuntimeException {
    PaymentFailedException(String m) { super(m); }
}

// ======================= Catalogue =======================
enum SeatType { REGULAR, PREMIUM }

record Theatre(String id, String city) {}

record Screen(String id, Theatre theatre, Map<String, SeatType> layout) {}

record Movie(String id, String title, int durationMinutes) {}

// ======================= Inventory =======================
final class Show {
    enum Status { AVAILABLE, HELD, BOOKED }

    static final class SeatState {
        Status status = Status.AVAILABLE;
        String holdId;
        long expiresAt;
    }

    final String id;
    final Movie movie;
    final Screen screen;
    final long start, end;
    final Map<SeatType, Long> prices;
    final SortedMap<String, SeatState> seats = new TreeMap<>();

    Show(String id, Movie movie, Screen screen, long start, Map<SeatType, Long> prices) {
        this.id = id;
        this.movie = movie;
        this.screen = screen;
        this.start = start;
        this.end = start + movie.durationMinutes() * 60_000L;
        this.prices = prices;
        for (String s : screen.layout().keySet()) seats.put(s, new SeatState());
    }

    boolean available(SeatState s, long now) {
        return s.status == Status.AVAILABLE || (s.status == Status.HELD && s.expiresAt <= now);
    }

    long price(String seatId) { return prices.get(screen.layout().get(seatId)); }
}

final class BookingService {
    private record Hold(String id, Show show, String user, List<String> seats, long expiresAt) {}

    private final PaymentGateway gateway;
    private final Clock clock;
    private final long ttl;
    private final Map<String, Theatre> theatres = new ConcurrentHashMap<>();
    private final Map<String, Screen> screens = new ConcurrentHashMap<>();
    private final Map<String, Movie> movies = new ConcurrentHashMap<>();
    private final Map<String, Show> shows = new ConcurrentHashMap<>();
    private final Map<String, Hold> holds = new ConcurrentHashMap<>();
    private final Map<String, Booking> bookings = new ConcurrentHashMap<>();
    private int nextShow = 1, nextHold = 1, nextBooking = 1;

    BookingService(PaymentGateway gateway, Clock clock, long holdTtlMillis) {
        this.gateway = gateway;
        this.clock = clock;
        this.ttl = holdTtlMillis;
    }

    // ---------- catalogue ----------
    void addTheatre(String theatreId, String city) {
        if (theatres.putIfAbsent(theatreId, new Theatre(theatreId, city)) != null) throw new IllegalArgumentException("duplicate theatre");
    }

    void addScreen(String theatreId, String screenId, int regularRows, int premiumRows, int seatsPerRow) {
        Theatre t = get(theatres, theatreId, "theatre");
        if (regularRows < 0 || premiumRows < 0 || regularRows + premiumRows == 0 || regularRows + premiumRows > 26 || seatsPerRow < 1)
            throw new IllegalArgumentException("invalid layout");
        Map<String, SeatType> layout = new LinkedHashMap<>();
        for (int r = 0; r < regularRows + premiumRows; r++)
            for (int n = 1; n <= seatsPerRow; n++)
                layout.put("" + (char) ('A' + r) + n, r < regularRows ? SeatType.REGULAR : SeatType.PREMIUM);
        if (screens.putIfAbsent(screenId, new Screen(screenId, t, Map.copyOf(layout))) != null) throw new IllegalArgumentException("duplicate screen");
    }

    void addMovie(String id, String title, int durationMinutes) {
        if (durationMinutes <= 0) throw new IllegalArgumentException("duration must be positive");
        if (movies.putIfAbsent(id, new Movie(id, title, durationMinutes)) != null) throw new IllegalArgumentException("duplicate movie");
    }

    synchronized String addShow(String movieId, String screenId, long startMillis, long regularPrice, long premiumPrice) {
        Movie m = get(movies, movieId, "movie");
        Screen sc = get(screens, screenId, "screen");
        long end = startMillis + m.durationMinutes() * 60_000L;
        for (Show s : shows.values())
            if (s.screen == sc && startMillis < s.end && s.start < end) throw new IllegalStateException("overlaps show " + s.id);
        String id = "S" + nextShow++;
        shows.put(id, new Show(id, m, sc, startMillis, Map.of(SeatType.REGULAR, regularPrice, SeatType.PREMIUM, premiumPrice)));
        return id;
    }

    List<String> findShows(String city, String movieId) {
        long now = clock.nowMillis();
        return shows.values().stream()
                .filter(s -> s.movie.id().equals(movieId) && s.screen.theatre().city().equalsIgnoreCase(city) && s.start > now)
                .sorted(Comparator.comparingLong((Show s) -> s.start).thenComparing(s -> s.id))
                .map(s -> s.id).toList();
    }

    List<String> availableSeats(String showId) {
        Show s = get(shows, showId, "show");
        synchronized (s) {
            long now = clock.nowMillis();
            List<String> out = new ArrayList<>();
            s.seats.forEach((id, st) -> { if (s.available(st, now)) out.add(id); });
            return out;
        }
    }

    // ---------- booking ----------
    String hold(String showId, String user, List<String> seatIds) {
        Show s = get(shows, showId, "show");
        if (seatIds == null || seatIds.isEmpty() || new HashSet<>(seatIds).size() != seatIds.size())
            throw new IllegalArgumentException("invalid seat list");
        synchronized (s) {                                     // one lock per show
            long now = clock.nowMillis();
            if (now >= s.start) throw new IllegalStateException("show already started");
            for (String id : seatIds) {
                Show.SeatState st = s.seats.get(id);
                if (st == null) throw new IllegalArgumentException("unknown seat " + id);
                if (!s.available(st, now)) throw new SeatUnavailableException(id);
            }
            String holdId = nextHoldId();
            for (String id : seatIds) {
                Show.SeatState st = s.seats.get(id);
                st.status = Show.Status.HELD;
                st.holdId = holdId;
                st.expiresAt = now + ttl;
            }
            List<String> sorted = new ArrayList<>(seatIds);
            Collections.sort(sorted);
            holds.put(holdId, new Hold(holdId, s, user, sorted, now + ttl));
            return holdId;
        }
    }

    Booking confirm(String holdId) {
        Hold h = holds.get(holdId);
        if (h == null) throw new IllegalStateException("hold not active");
        Show s = h.show();
        synchronized (s) {
            long now = clock.nowMillis();
            boolean active = h.expiresAt() > now && h.seats().stream().allMatch(id -> {
                Show.SeatState st = s.seats.get(id);
                return st.status == Show.Status.HELD && holdId.equals(st.holdId);
            });
            holds.remove(holdId);
            if (!active) throw new IllegalStateException("hold not active");
            long amount = h.seats().stream().mapToLong(s::price).sum();
            // NOTE: in production, call the gateway outside the show lock (see lesson).
            if (!gateway.charge(h.user(), amount)) {
                setStatus(s, h.seats(), Show.Status.AVAILABLE, null);
                throw new PaymentFailedException("payment declined for " + h.user());
            }
            setStatus(s, h.seats(), Show.Status.BOOKED, holdId);
            Booking b = new Booking(nextBookingId(), s.id, h.user(), h.seats(), amount, "CONFIRMED");
            bookings.put(b.id(), b);
            return b;
        }
    }

    Booking cancel(String bookingId) {
        Booking b = get(bookings, bookingId, "booking");
        Show s = shows.get(b.showId());
        synchronized (s) {
            Booking cur = bookings.get(bookingId);
            if (!cur.status().equals("CONFIRMED")) throw new IllegalStateException("already cancelled");
            if (clock.nowMillis() >= s.start) throw new IllegalStateException("show already started");
            setStatus(s, cur.seats(), Show.Status.AVAILABLE, null);
            Booking cancelled = new Booking(cur.id(), cur.showId(), cur.user(), cur.seats(), cur.amount(), "CANCELLED");
            bookings.put(bookingId, cancelled);
            return cancelled;
        }
    }

    // ---------- helpers ----------
    private static void setStatus(Show s, List<String> seats, Show.Status status, String holdId) {
        for (String id : seats) {
            Show.SeatState st = s.seats.get(id);
            st.status = status;
            st.holdId = holdId;
        }
    }

    private synchronized String nextHoldId() { return "H" + nextHold++; }
    private synchronized String nextBookingId() { return "B" + nextBooking++; }

    private static <T> T get(Map<String, T> m, String id, String what) {
        T v = m.get(id);
        if (v == null) throw new IllegalArgumentException("unknown " + what + " " + id);
        return v;
    }
}
