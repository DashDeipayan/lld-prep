import java.util.*;

interface Clock { long nowMillis(); }

interface AuctionListener { void onEvent(String auctionId, String event); }

record Result(String auctionId, String winner, long price) {}

final class BidRejectedException extends RuntimeException {
    final String reason;
    BidRejectedException(String reason) { super(reason); this.reason = reason; }
}

final class AuctionHouse {
    AuctionHouse(Clock clock) { /* TODO */ }

    String create(String seller, String item, long startingPrice, long minIncrement, long endMillis) { throw new UnsupportedOperationException("TODO"); }
    void subscribe(String auctionId, AuctionListener l) { throw new UnsupportedOperationException("TODO"); }
    void placeBid(String auctionId, String bidder, long amount) { throw new UnsupportedOperationException("TODO"); }
    Optional<String> leader(String auctionId) { throw new UnsupportedOperationException("TODO"); }
    long highest(String auctionId) { throw new UnsupportedOperationException("TODO"); }
    long endTime(String auctionId) { throw new UnsupportedOperationException("TODO"); }
    Result close(String auctionId) { throw new UnsupportedOperationException("TODO"); }
}
