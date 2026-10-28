import java.util.*;

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

// ======================= Your design =======================
final class BookingService {
    BookingService(PaymentGateway gateway, Clock clock, long holdTtlMillis) { /* TODO */ }

    void addTheatre(String theatreId, String city) { throw new UnsupportedOperationException("TODO"); }
    void addScreen(String theatreId, String screenId, int regularRows, int premiumRows, int seatsPerRow) { throw new UnsupportedOperationException("TODO"); }
    void addMovie(String id, String title, int durationMinutes) { throw new UnsupportedOperationException("TODO"); }
    String addShow(String movieId, String screenId, long startMillis, long regularPrice, long premiumPrice) { throw new UnsupportedOperationException("TODO"); }
    List<String> findShows(String city, String movieId) { throw new UnsupportedOperationException("TODO"); }
    List<String> availableSeats(String showId) { throw new UnsupportedOperationException("TODO"); }
    String hold(String showId, String user, List<String> seatIds) { throw new UnsupportedOperationException("TODO"); }
    Booking confirm(String holdId) { throw new UnsupportedOperationException("TODO"); }
    Booking cancel(String bookingId) { throw new UnsupportedOperationException("TODO"); }
}
