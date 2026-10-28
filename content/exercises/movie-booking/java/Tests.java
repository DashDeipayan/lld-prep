import java.util.*;
import java.util.concurrent.*;

public class Tests {
    static long[] now = {0};
    static List<String> charges = new ArrayList<>();
    static boolean[] approve = {true};
    static final long MIN = 60_000;

    static BookingService svc() {
        now[0] = 0;
        charges.clear();
        approve[0] = true;
        BookingService b = new BookingService((u, a) -> { charges.add(u + ":" + a); return approve[0]; }, () -> now[0], 5 * MIN);
        b.addTheatre("T1", "Bengaluru");
        b.addTheatre("T2", "Mumbai");
        b.addScreen("T1", "SC1", 2, 1, 3);
        b.addScreen("T2", "SC2", 1, 1, 2);
        b.addMovie("M1", "Dune", 120);
        b.addMovie("M2", "Up", 90);
        return b;
    }

    static void testScreenLayoutAndAvailability() {
        BookingService b = svc();
        String s = b.addShow("M1", "SC1", 60 * MIN, 200, 350);
        T.eq(List.of("A1", "A2", "A3", "B1", "B2", "B3", "C1", "C2", "C3"), b.availableSeats(s), "rows A-B regular, C premium");
        T.throwsEx(IllegalArgumentException.class, () -> b.addScreen("T9", "X", 1, 1, 1), "unknown theatre");
    }

    static void testShowOverlapOnSameScreen() {
        BookingService b = svc();
        b.addShow("M1", "SC1", 60 * MIN, 200, 350);
        T.throwsEx(IllegalStateException.class, () -> b.addShow("M2", "SC1", 179 * MIN, 200, 350), "starts before previous ends");
        T.eq("S2", b.addShow("M2", "SC1", 180 * MIN, 200, 350), "back-to-back is fine");
        T.eq("S3", b.addShow("M1", "SC2", 60 * MIN, 200, 350), "same time, different screen");
    }

    static void testFindShowsByCityMovieAndTime() {
        BookingService b = svc();
        String late = b.addShow("M1", "SC1", 300 * MIN, 200, 350);
        String early = b.addShow("M1", "SC1", 60 * MIN, 200, 350);
        b.addShow("M1", "SC2", 60 * MIN, 200, 350);
        b.addShow("M2", "SC1", 500 * MIN, 200, 350);
        T.eq(List.of(early, late), b.findShows("bengaluru", "M1"), "city match is case-insensitive, sorted by start");
        now[0] = 61 * MIN;
        T.eq(List.of(late), b.findShows("Bengaluru", "M1"), "started shows excluded");
    }

    static void testHoldConfirmPricing() {
        BookingService b = svc();
        String s = b.addShow("M1", "SC1", 60 * MIN, 200, 350);
        String h = b.hold(s, "asha", List.of("C1", "A1"));
        T.eq("H1", h, "hold id");
        Booking bk = b.confirm(h);
        T.eq(new Booking("B1", s, "asha", List.of("A1", "C1"), 550, "CONFIRMED"), bk, "regular 200 + premium 350, seats sorted");
        T.eq(List.of("asha:550"), charges, "gateway charged once with the total");
        T.throwsEx(SeatUnavailableException.class, () -> b.hold(s, "ravi", List.of("A1")), "booked seat");
        T.throwsEx(IllegalStateException.class, () -> b.confirm(h), "hold already used");
    }

    static void testPaymentFailureReleasesSeats() {
        BookingService b = svc();
        String s = b.addShow("M1", "SC1", 60 * MIN, 200, 350);
        String h = b.hold(s, "asha", List.of("A1", "A2"));
        approve[0] = false;
        T.throwsEx(PaymentFailedException.class, () -> b.confirm(h), "declined");
        T.eq(9, b.availableSeats(s).size(), "seats released after failed payment");
    }

    static void testHoldExpiryAndShowStart() {
        BookingService b = svc();
        String s = b.addShow("M1", "SC1", 60 * MIN, 200, 350);
        String h = b.hold(s, "asha", List.of("A1"));
        now[0] += 5 * MIN;
        T.throwsEx(IllegalStateException.class, () -> b.confirm(h), "expired hold");
        T.eq(List.of(), charges, "no charge for expired hold");
        b.hold(s, "ravi", List.of("A1"));
        now[0] = 60 * MIN;
        T.throwsEx(IllegalStateException.class, () -> b.hold(s, "x", List.of("B1")), "show started");
    }

    static void testCancellation() {
        BookingService b = svc();
        String s = b.addShow("M1", "SC1", 60 * MIN, 200, 350);
        Booking bk = b.confirm(b.hold(s, "asha", List.of("B2")));
        Booking c = b.cancel(bk.id());
        T.eq("CANCELLED", c.status(), "cancelled");
        T.check(b.availableSeats(s).contains("B2"), "seat back on sale");
        T.throwsEx(IllegalStateException.class, () -> b.cancel(bk.id()), "double cancel");
        Booking bk2 = b.confirm(b.hold(s, "ravi", List.of("B2")));
        now[0] = 61 * MIN;
        T.throwsEx(IllegalStateException.class, () -> b.cancel(bk2.id()), "after show start");
    }

    static void testNoDoubleBookingUnderContention() throws Exception {
        BookingService b = svc();
        String s = b.addShow("M1", "SC1", 60 * MIN, 200, 350);
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<Booking>> fs = new ArrayList<>();
        for (int i = 0; i < 40; i++) {
            List<String> want = List.of("A" + (1 + i % 3), "C" + (1 + (i / 3) % 3));
            String user = "u" + i;
            fs.add(pool.submit(() -> {
                go.await();
                try { return b.confirm(b.hold(s, user, want)); } catch (SeatUnavailableException e) { return null; }
            }));
        }
        go.countDown();
        Map<String, String> owner = new HashMap<>();
        for (Future<Booking> f : fs) {
            Booking bk = f.get(10, TimeUnit.SECONDS);
            if (bk == null) continue;
            for (String seat : bk.seats()) T.check(owner.put(seat, bk.id()) == null, "seat " + seat + " sold twice");
        }
        pool.shutdownNow();
        T.check(!owner.isEmpty(), "some bookings succeed");
    }
}
