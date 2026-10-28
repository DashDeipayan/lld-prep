import java.util.*;
import java.util.concurrent.*;

public class Tests {
    static long[] now = {0};
    static List<String> events = Collections.synchronizedList(new ArrayList<>());

    static AuctionHouse house() {
        now[0] = 0;
        events.clear();
        return new AuctionHouse(() -> now[0]);
    }

    static String reason(Runnable r) {
        try { r.run(); return "accepted"; } catch (BidRejectedException e) { return e.reason; }
    }

    static void testBidRules() {
        AuctionHouse h = house();
        String a = h.create("sam", "lamp", 100, 10, 1_000_000);
        T.eq("TOO_LOW", reason(() -> h.placeBid(a, "asha", 99)), "below starting price");
        h.placeBid(a, "asha", 100);
        T.eq("LEADING", reason(() -> h.placeBid(a, "asha", 200)), "leader can't outbid self");
        T.eq("TOO_LOW", reason(() -> h.placeBid(a, "ravi", 109)), "needs highest + increment");
        h.placeBid(a, "ravi", 110);
        T.eq(Optional.of("ravi"), h.leader(a), "leader");
        T.eq(110L, h.highest(a), "highest");
        T.throwsEx(IllegalArgumentException.class, () -> h.placeBid(a, "sam", 500), "seller can't bid");
        T.throwsEx(IllegalArgumentException.class, () -> h.create("sam", "x", 100, 10, 0), "end must be in the future");
    }

    static void testOutbidNotificationsInOrder() {
        AuctionHouse h = house();
        String a = h.create("sam", "lamp", 100, 10, 1_000_000);
        h.subscribe(a, (id, e) -> events.add(id + " " + e));
        h.placeBid(a, "asha", 100);
        h.placeBid(a, "ravi", 150);
        h.placeBid(a, "asha", 160);
        T.eq(List.of("A1 BID:asha:100", "A1 BID:ravi:150", "A1 OUTBID:asha", "A1 BID:asha:160", "A1 OUTBID:ravi"), events, "events");
    }

    static void testAntiSnipingExtension() {
        AuctionHouse h = house();
        String a = h.create("sam", "lamp", 100, 10, 600_000);
        now[0] = 500_000;
        h.placeBid(a, "asha", 100);
        T.eq(600_000L, h.endTime(a), "100s left: no extension");
        now[0] = 570_000;
        h.placeBid(a, "ravi", 110);
        T.eq(630_000L, h.endTime(a), "30s left: extended to now + 60s");
        now[0] = 629_999;
        h.placeBid(a, "asha", 120);
        T.eq(689_999L, h.endTime(a), "extends again");
        now[0] = 689_999;
        T.eq("CLOSED", reason(() -> h.placeBid(a, "ravi", 500)), "closed at end time");
    }

    static void testCloseIsIdempotent() {
        AuctionHouse h = house();
        String a = h.create("sam", "lamp", 100, 10, 1_000_000);
        String b = h.create("sam", "rug", 100, 10, 1_000_000);
        h.subscribe(a, (id, e) -> events.add(e));
        h.subscribe(b, (id, e) -> events.add(e));
        h.placeBid(a, "asha", 100);
        T.throwsEx(IllegalStateException.class, () -> h.close(a), "can't close early");
        now[0] = 1_000_000;
        Result r = h.close(a);
        T.eq(new Result("A1", "asha", 100), r, "winner");
        T.check(r == h.close(a), "second close returns the same result");
        T.eq(new Result("A2", null, 0), h.close(b), "unsold");
        T.eq(List.of("BID:asha:100", "CLOSED:asha:100", "CLOSED:UNSOLD"), events, "close notifies once");
        T.eq("CLOSED", reason(() -> h.placeBid(a, "ravi", 999)), "no bids after close");
    }

    static void testConcurrentBidsStayConsistent() throws Exception {
        AuctionHouse h = house();
        String a = h.create("sam", "lamp", 100, 10, 10_000_000);
        List<Long> accepted = Collections.synchronizedList(new ArrayList<>());
        h.subscribe(a, (id, e) -> { if (e.startsWith("BID:")) accepted.add(Long.parseLong(e.substring(e.lastIndexOf(':') + 1))); });
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<?>> fs = new ArrayList<>();
        for (int i = 0; i < 40; i++) {
            int n = i;
            fs.add(pool.submit(() -> { go.await(); reason(() -> h.placeBid(a, "u" + n, 100 + 10L * n)); return null; }));
        }
        go.countDown();
        for (Future<?> f : fs) f.get(10, TimeUnit.SECONDS);
        pool.shutdownNow();
        for (int i = 1; i < accepted.size(); i++) T.check(accepted.get(i) > accepted.get(i - 1), "accepted bids strictly increasing: " + accepted);
        T.eq(Optional.of("u39"), h.leader(a), "highest bidder wins");
        T.eq(490L, h.highest(a), "highest amount");
    }
}
