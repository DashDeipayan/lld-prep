import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

public class Tests {
    static Scheduler sched() {
        Scheduler s = new Scheduler();
        s.addRoom("R1", 4);
        s.addRoom("R2", 8);
        s.addRoom("R0", 4);
        s.addRoom("R3", 2);
        return s;
    }

    static void testBestFitRoomSelection() {
        Scheduler s = sched();
        T.eq("R0", s.book("a", 600, 660, 3).room(), "smallest capacity >= 3, tie -> lowest id (R0)");
        T.eq("R1", s.book("b", 600, 660, 3).room(), "R0 busy -> R1");
        T.eq("R2", s.book("c", 600, 660, 3).room(), "both 4-seaters busy -> R2");
        T.throwsEx(NoRoomAvailableException.class, () -> s.book("d", 600, 660, 3), "nothing left that fits");
        T.eq("R3", s.book("e", 600, 660, 2).room(), "2-seater still free");
    }

    static void testHalfOpenIntervals() {
        Scheduler s = sched();
        s.bookRoom("R1", "a", 600, 660);
        T.eq("M2", s.bookRoom("R1", "b", 660, 720).id(), "back-to-back is fine");
        T.throwsEx(IllegalStateException.class, () -> s.bookRoom("R1", "c", 659, 661), "overlap");
        T.throwsEx(IllegalStateException.class, () -> s.bookRoom("R1", "c", 500, 800), "enclosing overlap");
        T.throwsEx(IllegalStateException.class, () -> s.bookRoom("R1", "c", 610, 620), "inside overlap");
        T.eq("M3", s.bookRoom("R1", "c", 540, 600).id(), "ends exactly at next start");
    }

    static void testValidation() {
        Scheduler s = sched();
        T.throwsEx(IllegalArgumentException.class, () -> s.book("a", 600, 600, 2), "empty interval");
        T.throwsEx(IllegalArgumentException.class, () -> s.book("a", -1, 60, 2), "negative");
        T.throwsEx(IllegalArgumentException.class, () -> s.book("a", 1400, 1441, 2), "past midnight");
        T.throwsEx(IllegalArgumentException.class, () -> s.book("a", 0, 60, 0), "no attendees");
        T.throwsEx(IllegalArgumentException.class, () -> s.bookRoom("RX", "a", 0, 60), "unknown room");
        T.throwsEx(IllegalArgumentException.class, () -> s.addRoom("R1", 3), "duplicate room");
    }

    static void testCancelAndSchedule() {
        Scheduler s = sched();
        Booking b1 = s.bookRoom("R2", "a", 700, 760);
        Booking b2 = s.bookRoom("R2", "b", 600, 630);
        T.eq(List.of(b2, b1), s.schedule("R2"), "sorted by start");
        s.cancel(b1.id());
        T.eq(List.of(b2), s.schedule("R2"), "cancelled removed");
        T.throwsEx(IllegalStateException.class, () -> s.cancel(b1.id()), "double cancel");
        T.eq("R2", s.bookRoom("R2", "c", 700, 760).room(), "slot reusable after cancel");
    }

    static void testFreeRoomsAndSlots() {
        Scheduler s = sched();
        s.bookRoom("R0", "a", 600, 660);
        T.eq(List.of("R3", "R1", "R2"), s.freeRooms(630, 690, 2), "fit + free, ordered by capacity then id");
        s.bookRoom("R1", "a", 540, 570);
        s.bookRoom("R1", "b", 600, 630);
        s.bookRoom("R1", "c", 630, 700);
        T.eq(List.of(new Slot(480, 540), new Slot(570, 600), new Slot(700, 720)), s.freeSlots("R1", 480, 720), "gaps, adjacent bookings merge");
        T.eq(List.of(), s.freeSlots("R1", 600, 700), "fully booked window");
        T.eq(List.of(new Slot(0, 1440)), s.freeSlots("R2", 0, 1440), "empty room");
    }

    static void testSingleWinnerUnderContention() throws Exception {
        Scheduler s = sched();
        AtomicInteger ok = new AtomicInteger();
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<?>> fs = new ArrayList<>();
        for (int i = 0; i < 50; i++) {
            String who = "u" + i;
            fs.add(pool.submit(() -> {
                go.await();
                try { s.bookRoom("R1", who, 600, 660); ok.incrementAndGet(); } catch (IllegalStateException ignored) {}
                return null;
            }));
        }
        go.countDown();
        for (Future<?> f : fs) f.get(10, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.eq(1, ok.get(), "exactly one booking wins");
        T.eq(1, s.schedule("R1").size(), "one booking stored");
    }
}
