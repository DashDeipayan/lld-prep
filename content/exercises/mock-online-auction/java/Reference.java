import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;

interface Clock { long nowMillis(); }

interface AuctionListener { void onEvent(String auctionId, String event); }

record Result(String auctionId, String winner, long price) {}

final class BidRejectedException extends RuntimeException {
    final String reason;
    BidRejectedException(String reason) { super(reason); this.reason = reason; }
}

final class AuctionHouse {
    private static final long SNIPE_WINDOW = 60_000;

    private static final class Auction {
        final String id, seller, item;
        final long startingPrice, minIncrement;
        long end;
        String leader;
        long highest;
        Result result;
        final List<AuctionListener> listeners = new CopyOnWriteArrayList<>();
        Auction(String id, String seller, String item, long start, long inc, long end) {
            this.id = id; this.seller = seller; this.item = item; this.startingPrice = start; this.minIncrement = inc; this.end = end;
        }
    }

    private final Clock clock;
    private final Map<String, Auction> auctions = new ConcurrentHashMap<>();
    private int next = 1;

    AuctionHouse(Clock clock) { this.clock = clock; }

    synchronized String create(String seller, String item, long startingPrice, long minIncrement, long endMillis) {
        if (endMillis <= clock.nowMillis() || startingPrice <= 0 || minIncrement <= 0) throw new IllegalArgumentException("invalid auction");
        Auction a = new Auction("A" + next++, seller, item, startingPrice, minIncrement, endMillis);
        auctions.put(a.id, a);
        return a.id;
    }

    void subscribe(String auctionId, AuctionListener l) { auction(auctionId).listeners.add(l); }

    void placeBid(String auctionId, String bidder, long amount) {
        Auction a = auction(auctionId);
        List<String> events = new ArrayList<>();
        synchronized (a) {                                        // one lock per auction
            long now = clock.nowMillis();
            if (a.result != null || now >= a.end) throw new BidRejectedException("CLOSED");
            if (bidder.equals(a.seller)) throw new IllegalArgumentException("seller cannot bid");
            if (bidder.equals(a.leader)) throw new BidRejectedException("LEADING");
            long min = a.leader == null ? a.startingPrice : a.highest + a.minIncrement;
            if (amount < min) throw new BidRejectedException("TOO_LOW");
            String previous = a.leader;
            a.leader = bidder;
            a.highest = amount;
            if (a.end - now < SNIPE_WINDOW) a.end = now + SNIPE_WINDOW;
            events.add("BID:" + bidder + ":" + amount);
            if (previous != null) events.add("OUTBID:" + previous);
            for (String e : events) for (AuctionListener l : a.listeners) l.onEvent(a.id, e);   // in-order per auction
        }
    }

    Optional<String> leader(String auctionId) {
        Auction a = auction(auctionId);
        synchronized (a) { return Optional.ofNullable(a.leader); }
    }

    long highest(String auctionId) {
        Auction a = auction(auctionId);
        synchronized (a) { return a.highest; }
    }

    long endTime(String auctionId) {
        Auction a = auction(auctionId);
        synchronized (a) { return a.end; }
    }

    Result close(String auctionId) {
        Auction a = auction(auctionId);
        synchronized (a) {
            if (a.result != null) return a.result;                  // idempotent
            if (clock.nowMillis() < a.end) throw new IllegalStateException("auction still running");
            a.result = new Result(a.id, a.leader, a.leader == null ? 0 : a.highest);
            String ev = a.leader == null ? "CLOSED:UNSOLD" : "CLOSED:" + a.leader + ":" + a.highest;
            for (AuctionListener l : a.listeners) l.onEvent(a.id, ev);
            return a.result;
        }
    }

    private Auction auction(String id) {
        Auction a = auctions.get(id);
        if (a == null) throw new IllegalArgumentException("unknown auction " + id);
        return a;
    }
}
