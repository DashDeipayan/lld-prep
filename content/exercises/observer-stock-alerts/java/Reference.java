import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;

interface PriceListener {
    void onPrice(String symbol, long price);
}

final class StockTicker {
    private final Map<String, Long> last = new ConcurrentHashMap<>();
    // Snapshot-on-iterate: safe for (un)subscribe during notification.
    private final List<PriceListener> listeners = new CopyOnWriteArrayList<>();

    Runnable subscribe(PriceListener l) {
        Objects.requireNonNull(l);
        // Wrap so that the same listener object can be subscribed twice and removed independently.
        PriceListener entry = l::onPrice;
        listeners.add(entry);
        return () -> listeners.remove(entry);
    }

    void update(String symbol, long price) {
        Long prev = last.put(symbol, price);
        if (prev != null && prev == price) return;
        for (PriceListener l : listeners) {
            if (!listeners.contains(l)) continue; // removed earlier in this round
            try {
                l.onPrice(symbol, price);
            } catch (RuntimeException e) {
                // isolate faulty listener
            }
        }
    }

    int listenerCount() { return listeners.size(); }
}

final class ThresholdAlert implements PriceListener {
    private final String symbol;
    private final long threshold;
    private final boolean above;
    private final List<String> sink;
    private boolean triggered;

    ThresholdAlert(String symbol, long threshold, boolean above, List<String> sink) {
        this.symbol = symbol;
        this.threshold = threshold;
        this.above = above;
        this.sink = sink;
    }

    public void onPrice(String sym, long price) {
        if (!symbol.equals(sym)) return;
        boolean hit = above ? price >= threshold : price <= threshold;
        if (hit && !triggered) {
            sink.add(symbol + (above ? ">=" : "<=") + threshold + " @" + price);
        }
        triggered = hit;
    }
}
