import java.util.*;
import java.util.concurrent.*;

public class Tests {
    static long[] now = {0};
    static final Map<VehicleClass, Rate> RATES = Map.of(VehicleClass.MINI, new Rate(50, 10, 2), VehicleClass.SEDAN, new Rate(80, 15, 3), VehicleClass.SUV, new Rate(100, 20, 5));

    static RideService svc() {
        now[0] = 0;
        RideService s = new RideService(new NearestDriverStrategy(10), new StandardPricing(RATES, 100), () -> now[0]);
        s.addDriver("d1", VehicleClass.MINI, new Location(0, 0), 4.5);
        s.addDriver("d2", VehicleClass.MINI, new Location(5, 5), 4.9);
        s.addDriver("d3", VehicleClass.SEDAN, new Location(1, 1), 4.7);
        for (String d : List.of("d1", "d2", "d3")) s.setOnline(d, true);
        return s;
    }

    static void testNearestStrategyTieBreaks() {
        MatchingStrategy m = new NearestDriverStrategy(5);
        Location p = new Location(0, 0);
        List<DriverInfo> c = List.of(
                new DriverInfo("a", VehicleClass.MINI, new Location(2, 1), 4.1),
                new DriverInfo("b", VehicleClass.MINI, new Location(1, 2), 4.8),
                new DriverInfo("c", VehicleClass.MINI, new Location(3, 0), 4.8),
                new DriverInfo("z", VehicleClass.MINI, new Location(0, 9), 5.0));
        T.eq(Optional.of("b"), m.match(p, c), "distance 3 tie: higher rating wins");
        T.eq(Optional.of("b"), m.match(p, List.of(c.get(2), c.get(1))), "same distance and rating: lower id wins");
        T.eq(Optional.empty(), m.match(p, List.of(c.get(3))), "out of range");
    }

    static void testPricing() {
        PricingStrategy p = new StandardPricing(RATES, 100);
        T.eq(50L + 10 * 7 + 2 * 3, p.fare(VehicleClass.MINI, 7, 2 * 60_000 + 1), "minutes rounded up");
        T.eq((80L + 15 * 4 + 3 * 10) * 150 / 100, new StandardPricing(RATES, 150).fare(VehicleClass.SEDAN, 4, 10 * 60_000), "1.5x surge");
    }

    static void testRequestMatchesNearestAvailableOfClass() {
        RideService s = svc();
        String r1 = s.requestRide("asha", new Location(1, 0), new Location(9, 9), VehicleClass.MINI);
        T.eq("R1", r1, "trip id");
        T.eq("d1", s.driverOf(r1), "nearest mini");
        T.eq(TripStatus.ASSIGNED, s.status(r1), "assigned");
        String r2 = s.requestRide("ravi", new Location(1, 0), new Location(9, 9), VehicleClass.MINI);
        T.eq("d2", s.driverOf(r2), "d1 is busy, d2 next");
        T.throwsEx(NoDriverAvailableException.class, () -> s.requestRide("kim", new Location(0, 0), new Location(1, 1), VehicleClass.MINI), "no minis left");
        T.throwsEx(NoDriverAvailableException.class, () -> s.requestRide("kim", new Location(0, 0), new Location(1, 1), VehicleClass.SUV), "no SUVs at all");
        T.throwsEx(IllegalStateException.class, () -> s.requestRide("asha", new Location(0, 0), new Location(1, 1), VehicleClass.SEDAN), "rider already on a trip");
    }

    static void testOfflineDriversNotMatched() {
        RideService s = svc();
        s.setOnline("d1", false);
        T.eq("d2", s.driverOf(s.requestRide("asha", new Location(0, 0), new Location(3, 3), VehicleClass.MINI)), "offline d1 skipped");
        T.throwsEx(IllegalStateException.class, () -> s.setOnline("d2", false), "can't go offline mid-trip");
    }

    static void testLifecycleAndFare() {
        RideService s = svc();
        String r = s.requestRide("asha", new Location(0, 1), new Location(4, 4), VehicleClass.MINI);
        T.throwsEx(IllegalStateException.class, () -> s.start(r), "must arrive first");
        s.arrived(r);
        now[0] = 1_000;
        s.start(r);
        now[0] = 1_000 + 9 * 60_000 + 30_000;
        T.eq(50L + 10 * 7 + 2 * 10, s.end(r), "fare: distance 7, 9.5 min -> 10");
        T.eq(TripStatus.COMPLETED, s.status(r), "completed");
        T.throwsEx(IllegalStateException.class, () -> s.cancel(r), "can't cancel a completed trip");
        String r2 = s.requestRide("ravi", new Location(4, 4), new Location(0, 0), VehicleClass.MINI);
        T.eq("d1", s.driverOf(r2), "d1 free again and now located at (4,4)");
    }

    static void testCancellationFreesDriverAndRider() {
        RideService s = svc();
        String r = s.requestRide("asha", new Location(0, 0), new Location(2, 2), VehicleClass.SEDAN);
        s.arrived(r);
        s.cancel(r);
        T.eq(TripStatus.CANCELLED, s.status(r), "cancelled");
        String r2 = s.requestRide("asha", new Location(0, 0), new Location(2, 2), VehicleClass.SEDAN);
        T.eq("d3", s.driverOf(r2), "driver and rider both free again");
        s.arrived(r2);
        s.start(r2);
        T.throwsEx(IllegalStateException.class, () -> s.cancel(r2), "can't cancel in progress");
    }

    static void testNoDoubleAssignmentUnderContention() throws Exception {
        now[0] = 0;
        RideService s = new RideService(new NearestDriverStrategy(100), new StandardPricing(RATES, 100), () -> now[0]);
        for (int i = 0; i < 10; i++) { s.addDriver("d" + i, VehicleClass.MINI, new Location(i, 0), 4.0); s.setOnline("d" + i, true); }
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<String>> fs = new ArrayList<>();
        for (int i = 0; i < 50; i++) {
            String rider = "u" + i;
            fs.add(pool.submit(() -> {
                go.await();
                try { return s.driverOf(s.requestRide(rider, new Location(0, 0), new Location(5, 5), VehicleClass.MINI)); }
                catch (NoDriverAvailableException e) { return null; }
            }));
        }
        go.countDown();
        List<String> assigned = new ArrayList<>();
        for (Future<String> f : fs) { String d = f.get(10, TimeUnit.SECONDS); if (d != null) assigned.add(d); }
        pool.shutdownNow();
        T.eq(10, assigned.size(), "exactly 10 trips");
        T.eq(10, new HashSet<>(assigned).size(), "10 distinct drivers");
    }
}
