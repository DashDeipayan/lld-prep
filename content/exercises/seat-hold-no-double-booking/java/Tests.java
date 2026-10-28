import java.util.*;
import java.util.concurrent.*;

public class Tests {
    static long[] now = {0};
    static SeatInventory inv(String... ids) { now[0] = 1_000; return new SeatInventory(List.of(ids), 1_000, () -> now[0]); }

    static void testHoldAndStatus() {
        SeatInventory s = inv("A1", "A2", "A3");
        String h = s.hold("u", List.of("A1", "A3"));
        T.eq("H1", h, "first hold id");
        T.eq("HELD", s.status("A1"), "held");
        T.eq(List.of("A2"), s.available(), "available sorted");
        T.eq("H2", s.hold("v", List.of("A2")), "ids increase");
    }

    static void testAllOrNothing() {
        SeatInventory s = inv("A1", "A2", "A3");
        s.hold("u", List.of("A2"));
        T.throwsEx(SeatUnavailableException.class, () -> s.hold("v", List.of("A1", "A2", "A3")), "one seat taken");
        T.eq("AVAILABLE", s.status("A1"), "A1 not held by the failed attempt");
        T.eq("AVAILABLE", s.status("A3"), "A3 not held by the failed attempt");
    }

    static void testExpiryIsLazyAndExact() {
        SeatInventory s = inv("A1");
        String h = s.hold("u", List.of("A1"));
        now[0] += 999;
        T.eq("HELD", s.status("A1"), "still held 1ms before expiry");
        now[0] += 1;
        T.eq("AVAILABLE", s.status("A1"), "expired at exactly now == expiresAt");
        T.throwsEx(IllegalStateException.class, () -> s.confirm(h), "cannot confirm an expired hold");
        T.eq("H2", s.hold("v", List.of("A1")), "seat can be re-held after expiry");
    }

    static void testConfirmAndRelease() {
        SeatInventory s = inv("A1", "A2");
        String h1 = s.hold("u", List.of("A1"));
        String h2 = s.hold("v", List.of("A2"));
        s.confirm(h1);
        s.release(h2);
        now[0] += 10_000;
        T.eq("BOOKED", s.status("A1"), "booked seats never expire");
        T.eq("AVAILABLE", s.status("A2"), "released");
        T.throwsEx(IllegalStateException.class, () -> s.confirm(h1), "double confirm");
        T.throwsEx(IllegalStateException.class, () -> s.release(h2), "double release");
        T.throwsEx(IllegalStateException.class, () -> s.confirm("H99"), "unknown hold");
        T.throwsEx(SeatUnavailableException.class, () -> s.hold("w", List.of("A1")), "booked seat cannot be held");
    }

    static void testInvalidInput() {
        SeatInventory s = inv("A1", "A2");
        T.throwsEx(IllegalArgumentException.class, () -> s.hold("u", List.of()), "empty");
        T.throwsEx(IllegalArgumentException.class, () -> s.hold("u", List.of("A1", "A1")), "duplicate");
        T.throwsEx(IllegalArgumentException.class, () -> s.hold("u", List.of("Z9")), "unknown seat");
        T.eq(List.of("A1", "A2"), s.available(), "nothing changed");
    }

    static void testNoDoubleBookingUnderContention() throws Exception {
        List<String> ids = new ArrayList<>();
        for (int i = 0; i < 12; i++) ids.add("S" + (char) ('a' + i));
        SeatInventory s = new SeatInventory(ids, 60_000, () -> 0);
        int n = 64;
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<Map.Entry<String, List<String>>>> fs = new ArrayList<>();
        for (int t = 0; t < n; t++) {
            List<String> want = List.of(ids.get(t % 12), ids.get((t * 5 + 1) % 12), ids.get((t * 7 + 3) % 12));
            List<String> uniq = new ArrayList<>(new LinkedHashSet<>(want));
            fs.add(pool.submit(() -> {
                go.await();
                try { return Map.entry(s.hold("u", uniq), uniq); } catch (SeatUnavailableException e) { return null; }
            }));
        }
        go.countDown();
        Map<String, String> owner = new HashMap<>();
        int ok = 0;
        for (Future<Map.Entry<String, List<String>>> f : fs) {
            Map.Entry<String, List<String>> r = f.get(10, TimeUnit.SECONDS);
            if (r == null) continue;
            ok++;
            for (String seat : r.getValue()) {
                String prev = owner.put(seat, r.getKey());
                T.check(prev == null, "seat " + seat + " held by both " + prev + " and " + r.getKey());
                T.eq("HELD", s.status(seat), "successful hold's seat is HELD");
            }
        }
        pool.shutdownNow();
        T.check(ok >= 1, "at least one hold must succeed");
    }
}
