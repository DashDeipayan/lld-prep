import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

public class Tests {
    static RentalService svc() {
        RentalService r = new RentalService();
        r.addBranch("B1", "Bengaluru");
        r.addBranch("B2", "Pune");
        r.addCar("B1", "C1", CarType.SEDAN, 2000);
        r.addCar("B1", "C2", CarType.SEDAN, 1500);
        r.addCar("B1", "C3", CarType.SUV, 3000);
        r.addCar("B2", "C4", CarType.SEDAN, 1000);
        return r;
    }

    static void testSearchFiltersAndSorts() {
        RentalService r = svc();
        T.eq(List.of("C2", "C1"), r.search("bengaluru", CarType.SEDAN, 1, 3), "city + type, sorted by rate");
        T.eq(List.of("C3"), r.search("Bengaluru", CarType.SUV, 1, 3), "SUV");
        T.eq(List.of(), r.search("Delhi", CarType.SEDAN, 1, 3), "no branch");
    }

    static void testReserveCostAndOverlap() {
        RentalService r = svc();
        Reservation a = r.reserve("asha", "C2", 10, 13);
        T.eq(new Reservation("RS1", "asha", "C2", 10, 13, 4500, "RESERVED"), a, "3 days x 1500");
        T.throwsEx(IllegalStateException.class, () -> r.reserve("ravi", "C2", 12, 15), "overlap");
        T.eq("RS2", r.reserve("ravi", "C2", 13, 15).id(), "back-to-back ok");
        T.eq(List.of("C1"), r.search("Bengaluru", CarType.SEDAN, 11, 12), "reserved car hidden from search");
        T.throwsEx(IllegalArgumentException.class, () -> r.reserve("x", "C2", 5, 5), "empty range");
        T.throwsEx(IllegalArgumentException.class, () -> r.reserve("x", "ZZ", 1, 2), "unknown car");
    }

    static void testCancel() {
        RentalService r = svc();
        Reservation a = r.reserve("asha", "C1", 1, 4);
        T.eq("CANCELLED", r.cancel(a.id()).status(), "cancelled");
        T.eq(List.of("C2", "C1"), r.search("Bengaluru", CarType.SEDAN, 1, 4), "free again");
        T.throwsEx(IllegalStateException.class, () -> r.cancel(a.id()), "double cancel");
    }

    static void testReturnWithLateFee() {
        RentalService r = svc();
        Reservation a = r.reserve("asha", "C1", 1, 4);
        T.eq(6000L + 2 * 2000 * 3 / 2, r.returnCar(a.id(), 6), "2 days late at 1.5x");
        Reservation b = r.reserve("ravi", "C1", 6, 8);
        T.eq(4000L, r.returnCar(b.id(), 7), "early return: no refund");
        T.throwsEx(IllegalStateException.class, () -> r.returnCar(b.id(), 9), "already completed");
        T.throwsEx(IllegalStateException.class, () -> r.cancel(b.id()), "can't cancel completed");
    }

    static void testSingleWinnerUnderContention() throws Exception {
        RentalService r = svc();
        AtomicInteger ok = new AtomicInteger();
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<?>> fs = new ArrayList<>();
        for (int i = 0; i < 30; i++) {
            int s = 10 + i % 3;
            fs.add(pool.submit(() -> {
                go.await();
                try { r.reserve("u", "C3", s, s + 3); ok.incrementAndGet(); } catch (IllegalStateException ignored) {}
                return null;
            }));
        }
        go.countDown();
        for (Future<?> f : fs) f.get(10, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.eq(1, ok.get(), "overlapping requests: exactly one wins");
    }
}
