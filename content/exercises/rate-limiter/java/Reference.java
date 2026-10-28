import java.util.*;
import java.util.concurrent.*;

interface RateLimiter {
    boolean allow(String key);
}

interface Clock {
    long nowMillis();
}

final class TokenBucketLimiter implements RateLimiter {
    private static final class Bucket { double tokens; long last; }

    private final long capacity;
    private final double perMilli;
    private final Clock clock;
    private final ConcurrentHashMap<String, Bucket> buckets = new ConcurrentHashMap<>();

    TokenBucketLimiter(long capacity, double refillPerSecond, Clock clock) {
        this.capacity = capacity;
        this.perMilli = refillPerSecond / 1000.0;
        this.clock = clock;
    }

    public boolean allow(String key) {
        long now = clock.nowMillis();
        Bucket b = buckets.computeIfAbsent(key, k -> { Bucket nb = new Bucket(); nb.tokens = capacity; nb.last = now; return nb; });
        synchronized (b) {                                       // per-key lock: keys never contend
            b.tokens = Math.min(capacity, b.tokens + (now - b.last) * perMilli);
            b.last = Math.max(b.last, now);
            if (b.tokens >= 1 - 1e-9) { b.tokens = Math.max(0, b.tokens - 1); return true; }   // epsilon: 0.3+0.3+0.3+0.1 < 1.0 in floating point
            return false;
        }
    }
}

final class FixedWindowLimiter implements RateLimiter {
    private static final class Window { long start = -1; int count; }

    private final int max;
    private final long window;
    private final Clock clock;
    private final ConcurrentHashMap<String, Window> windows = new ConcurrentHashMap<>();

    FixedWindowLimiter(int max, long windowMillis, Clock clock) {
        this.max = max;
        this.window = windowMillis;
        this.clock = clock;
    }

    public boolean allow(String key) {
        long now = clock.nowMillis();
        long start = now - Math.floorMod(now, window);
        Window w = windows.computeIfAbsent(key, k -> new Window());
        synchronized (w) {
            if (w.start != start) { w.start = start; w.count = 0; }
            if (w.count >= max) return false;
            w.count++;
            return true;
        }
    }
}

final class SlidingWindowLimiter implements RateLimiter {
    private final int max;
    private final long window;
    private final Clock clock;
    private final ConcurrentHashMap<String, Deque<Long>> logs = new ConcurrentHashMap<>();

    SlidingWindowLimiter(int max, long windowMillis, Clock clock) {
        this.max = max;
        this.window = windowMillis;
        this.clock = clock;
    }

    public boolean allow(String key) {
        long now = clock.nowMillis();
        Deque<Long> log = logs.computeIfAbsent(key, k -> new ArrayDeque<>());
        synchronized (log) {
            while (!log.isEmpty() && now - log.peekFirst() >= window) log.pollFirst();
            if (log.size() >= max) return false;
            log.addLast(now);
            return true;
        }
    }
}
