import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

public class Tests {
    static Hotel hotel() {
        Hotel h = new Hotel();
        h.addRoom("102", RoomType.DOUBLE, 4000);
        h.addRoom("101", RoomType.DOUBLE, 4000);
        h.addRoom("103", RoomType.SINGLE, 2500);
        h.addRoom("301", RoomType.SUITE, 9000);
        return h;
    }

    static void testAvailabilityAndAssignment() {
        Hotel h = hotel();
        T.eq(Map.of(RoomType.SINGLE, 1, RoomType.DOUBLE, 2, RoomType.SUITE, 1), h.availability(1, 3), "all free");
        String b1 = h.book("a", RoomType.DOUBLE, 1, 3);
        T.eq("BK1", b1, "id");
        T.eq("101", h.roomOf(b1), "lowest number first");
        T.eq("102", h.roomOf(h.book("b", RoomType.DOUBLE, 2, 4)), "overlapping stay gets the next room");
        T.eq(Map.of(RoomType.SINGLE, 1, RoomType.DOUBLE, 0, RoomType.SUITE, 1), h.availability(2, 3), "doubles full on night 2");
        T.eq("101", h.roomOf(h.book("c", RoomType.DOUBLE, 3, 5)), "101 free from night 3");
        T.throwsEx(NoRoomAvailableException.class, () -> h.book("d", RoomType.DOUBLE, 2, 4), "none left");
        T.throwsEx(IllegalArgumentException.class, () -> h.book("d", RoomType.SINGLE, 4, 4), "empty stay");
    }

    static void testLifecycleAndBill() {
        Hotel h = hotel();
        String b = h.book("a", RoomType.SUITE, 5, 8);
        T.throwsEx(IllegalStateException.class, () -> h.addCharge(b, 100), "no charges before check-in");
        T.throwsEx(IllegalStateException.class, () -> h.checkIn(b, 4), "wrong day");
        h.checkIn(b, 5);
        T.eq("CHECKED_IN", h.status(b), "checked in");
        h.addCharge(b, 1200);
        h.addCharge(b, 300);
        T.throwsEx(IllegalArgumentException.class, () -> h.addCharge(b, 0), "positive charge");
        T.eq(3 * 9000L + 1500, h.checkOut(b), "3 nights + charges");
        T.eq("CHECKED_OUT", h.status(b), "checked out");
        T.throwsEx(IllegalStateException.class, () -> h.checkOut(b), "double checkout");
        T.eq("301", h.roomOf(h.book("b", RoomType.SUITE, 6, 7)), "room freed after checkout");
    }

    static void testCancel() {
        Hotel h = hotel();
        String b = h.book("a", RoomType.SINGLE, 1, 2);
        h.cancel(b);
        T.eq("CANCELLED", h.status(b), "cancelled");
        T.eq(1, h.availability(1, 2).get(RoomType.SINGLE), "room free again");
        T.throwsEx(IllegalStateException.class, () -> h.cancel(b), "double cancel");
        String c = h.book("b", RoomType.SINGLE, 1, 2);
        h.checkIn(c, 1);
        T.throwsEx(IllegalStateException.class, () -> h.cancel(c), "can't cancel after check-in");
    }

    static void testLastSuiteSingleWinner() throws Exception {
        Hotel h = hotel();
        AtomicInteger ok = new AtomicInteger();
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<?>> fs = new ArrayList<>();
        for (int i = 0; i < 20; i++) fs.add(pool.submit(() -> {
            go.await();
            try { h.book("g", RoomType.SUITE, 10, 12); ok.incrementAndGet(); } catch (NoRoomAvailableException ignored) {}
            return null;
        }));
        go.countDown();
        for (Future<?> f : fs) f.get(10, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.eq(1, ok.get(), "one suite, one booking");
    }
}
