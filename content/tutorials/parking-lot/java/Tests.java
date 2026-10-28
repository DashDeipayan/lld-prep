import java.util.*;
import java.util.concurrent.*;

public class Tests {
    static long[] now = {0};
    static final Map<VehicleType, Long> RATES = Map.of(VehicleType.BIKE, 10L, VehicleType.CAR, 20L, VehicleType.TRUCK, 50L);
    static Vehicle bike(String p) { return new Vehicle(p, VehicleType.BIKE); }
    static Vehicle car(String p) { return new Vehicle(p, VehicleType.CAR); }
    static Vehicle truck(String p) { return new Vehicle(p, VehicleType.TRUCK); }

    static ParkingLot lot() {
        now[0] = 0;
        ParkingLot l = new ParkingLot(new HourlyPricing(RATES), new NearestSpotSelector(), () -> now[0]);
        l.addFloor(1, 2, 3, 1);
        l.addFloor(2, 1, 1, 1);
        return l;
    }

    // ---------------- step 1 ----------------
    static void s1_fittingRules() {
        Spot small = new Spot("x", 1, SpotSize.SMALL), medium = new Spot("y", 1, SpotSize.MEDIUM), large = new Spot("z", 1, SpotSize.LARGE);
        T.check(small.canFit(bike("b")) && medium.canFit(bike("b")) && large.canFit(bike("b")), "bike fits anywhere");
        T.check(!small.canFit(car("c")) && medium.canFit(car("c")) && large.canFit(car("c")), "car needs medium+");
        T.check(!small.canFit(truck("t")) && !medium.canFit(truck("t")) && large.canFit(truck("t")), "truck needs large");
        T.eq("y", medium.id(), "id");
        T.eq(SpotSize.MEDIUM, medium.size(), "size");
    }

    static void s1_occupyAndVacate() {
        Spot s = new Spot("F1-M1", 1, SpotSize.MEDIUM);
        T.check(s.isFree(), "new spot is free");
        s.occupy(car("KA1"));
        T.check(!s.isFree(), "occupied");
        T.throwsEx(IllegalStateException.class, () -> s.occupy(bike("KA2")), "double occupy");
        T.eq("KA1", s.vacate().plate(), "vacate returns the vehicle");
        T.check(s.isFree(), "free again");
        T.throwsEx(IllegalStateException.class, () -> new Spot("z", 1, SpotSize.SMALL).occupy(truck("T")), "occupy with non-fitting vehicle");
    }

    // ---------------- step 2 ----------------
    static void s2_floorsAndAvailability() {
        ParkingLot l = lot();
        T.eq(3, l.available(SpotSize.SMALL), "small across floors");
        T.eq(4, l.available(SpotSize.MEDIUM), "medium across floors");
        T.eq(2, l.available(SpotSize.LARGE), "large across floors");
        T.throwsEx(IllegalArgumentException.class, () -> l.addFloor(1, 1, 1, 1), "duplicate floor");
    }

    // ---------------- step 3 ----------------
    static void s3_parkIssuesTicketsNearestFirst() {
        ParkingLot l = lot();
        now[0] = 5_000;
        Ticket t1 = l.park(bike("B1"));
        T.eq("T1", t1.id(), "ticket id");
        T.eq("F1-S1", t1.spotId(), "bike takes the first small spot");
        T.eq(5_000L, t1.entryMillis(), "entry time from clock");
        T.eq("F1-M1", l.park(car("C1")).spotId(), "car takes first medium");
        T.eq("F1-L1", l.park(truck("TR1")).spotId(), "truck takes first large");
        T.eq("F2-L1", l.park(truck("TR2")).spotId(), "second truck goes to floor 2");
        T.eq(3, l.available(SpotSize.MEDIUM), "availability updated");
    }

    static void s3_fullAndDuplicate() {
        ParkingLot l = lot();
        l.park(truck("TR1"));
        l.park(truck("TR2"));
        T.throwsEx(ParkingFullException.class, () -> l.park(truck("TR3")), "no large spot left");
        T.throwsEx(IllegalStateException.class, () -> l.park(truck("TR1")), "plate already parked");
    }

    static void s3_customSelectorStrategy() {
        now[0] = 0;
        SpotSelector topFloorFirst = (cands, v) -> cands.stream().max(Comparator.comparingInt(Spot::floor)).orElse(null);
        ParkingLot l = new ParkingLot(new HourlyPricing(RATES), topFloorFirst, () -> now[0]);
        l.addFloor(1, 1, 1, 1);
        l.addFloor(2, 1, 1, 1);
        T.eq(2, spotFloor(l.park(car("C1")).spotId()), "custom strategy used without changing ParkingLot");
    }

    static int spotFloor(String id) { return Integer.parseInt(id.substring(1, id.indexOf('-'))); }

    // ---------------- step 4 ----------------
    static void s4_hourlyPricing() {
        HourlyPricing p = new HourlyPricing(RATES);
        long min = 60_000;
        T.eq(20L, p.fee(VehicleType.CAR, 0), "minimum one hour");
        T.eq(20L, p.fee(VehicleType.CAR, 59 * min), "59 min");
        T.eq(20L, p.fee(VehicleType.CAR, 60 * min), "exactly 60 min");
        T.eq(40L, p.fee(VehicleType.CAR, 61 * min), "61 min -> 2 hours");
        T.eq(150L, p.fee(VehicleType.TRUCK, 180 * min), "3h truck");
    }

    static void s4_unparkReturnsReceipt() {
        ParkingLot l = lot();
        Ticket t = l.park(car("C1"));
        now[0] += 90 * 60_000;
        Receipt r = l.unpark(t.id());
        T.eq(new Receipt("T1", "C1", "F1-M1", 90, 40), r, "receipt");
        T.eq(4, l.available(SpotSize.MEDIUM), "spot freed");
        T.throwsEx(IllegalArgumentException.class, () -> l.unpark(t.id()), "ticket can't be reused");
        T.throwsEx(IllegalArgumentException.class, () -> l.unpark("T999"), "unknown ticket");
        T.eq("F1-M1", l.park(car("C1")).spotId(), "same plate can park again");
    }

    // ---------------- step 5 ----------------
    static void s5_eventsInOrder() {
        ParkingLot l = lot();
        List<String> log = new ArrayList<>();
        l.addListener(e -> log.add("a:" + e.kind() + ":" + e.spotId()));
        l.addListener(e -> log.add("b:" + e.kind()));
        Ticket t = l.park(car("C1"));
        try { l.park(car("C1")); } catch (IllegalStateException ignored) {}
        l.unpark(t.id());
        T.eq(List.of("a:PARKED:F1-M1", "b:PARKED", "a:LEFT:F1-M1", "b:LEFT"), log, "events; failed park emits nothing");
    }

    static void s5_displayBoardTracksAvailability() {
        ParkingLot l = lot();
        Ticket pre = l.park(bike("B0"));
        DisplayBoard board = new DisplayBoard(l);
        T.eq(2, board.free(SpotSize.SMALL), "board starts from current availability");
        Ticket c = l.park(car("C1"));
        l.park(truck("TR1"));
        l.unpark(pre.id());
        l.unpark(c.id());
        for (SpotSize s : SpotSize.values()) T.eq(l.available(s), board.free(s), "board matches lot for " + s);
    }

    // ---------------- step 6 ----------------
    static void s6_concurrentParkingNeverDoubleAssigns() throws Exception {
        now[0] = 0;
        ParkingLot l = new ParkingLot(new HourlyPricing(RATES), new NearestSpotSelector(), () -> now[0]);
        l.addFloor(1, 0, 10, 0);
        l.addFloor(2, 0, 10, 0);
        DisplayBoard board = new DisplayBoard(l);
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<String>> fs = new ArrayList<>();
        for (int i = 0; i < 32; i++) {
            String plate = "C" + i;
            fs.add(pool.submit(() -> {
                go.await();
                try { return l.park(car(plate)).spotId(); } catch (ParkingFullException e) { return null; }
            }));
        }
        go.countDown();
        Set<String> spots = new HashSet<>();
        int ok = 0;
        for (Future<String> f : fs) {
            String s = f.get(10, TimeUnit.SECONDS);
            if (s != null) { ok++; spots.add(s); }
        }
        pool.shutdownNow();
        T.eq(20, ok, "exactly 20 parks succeed");
        T.eq(20, spots.size(), "20 distinct spots");
        T.eq(0, l.available(SpotSize.MEDIUM), "lot full");
        T.eq(0, board.free(SpotSize.MEDIUM), "board consistent under concurrency");
    }
}
