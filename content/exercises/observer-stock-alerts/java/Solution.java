import java.util.*;

interface PriceListener {
    void onPrice(String symbol, long price);
}

final class StockTicker {
    /** @return a handle that unsubscribes the listener */
    Runnable subscribe(PriceListener l) {
        throw new UnsupportedOperationException("TODO");
    }

    void update(String symbol, long price) {
        throw new UnsupportedOperationException("TODO");
    }

    int listenerCount() {
        throw new UnsupportedOperationException("TODO");
    }
}

final class ThresholdAlert implements PriceListener {
    ThresholdAlert(String symbol, long threshold, boolean above, List<String> sink) {
        // TODO
    }

    public void onPrice(String symbol, long price) {
        throw new UnsupportedOperationException("TODO");
    }
}
