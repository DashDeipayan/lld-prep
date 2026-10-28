import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

public class Tests {
    static long[] now = {0};

    static List<Boolean> calls(RateLimiter r, String key, long... at) {
        List<Boolean> out = new ArrayList<>();
        for (long t : at) { now[0] = t; out.add(r.allow(key)); }
        return out;
    }

    static void testTokenBucketBurstAndRefill() {
        now[0] = 0;
        RateLimiter r = new TokenBucketLimiter(3, 2, () -> now[0]);
        T.eq(List.of(true, true, true, false), calls(r, "k", 0, 0, 0, 0), "burst of capacity, then reject");
        T.eq(List.of(true, false), calls(r, "k", 500, 500), "2 tokens/s -> one token after 500ms");
        T.eq(List.of(true, true, true, false), calls(r, "k", 10_000, 10_000, 10_000, 10_000), "refill caps at capacity");
    }

    static void testTokenBucketFractionalAccumulation() {
        now[0] = 0;
        RateLimiter r = new TokenBucketLimiter(1, 1, () -> now[0]);
        T.eq(List.of(true, false, false, false, true), calls(r, "k", 0, 300, 600, 900, 1000), "fractions add up to one token at 1000ms");
    }

    static void testFixedWindowBoundaries() {
        RateLimiter r = new FixedWindowLimiter(2, 1000, () -> now[0]);
        T.eq(List.of(true, true, false, true, true, false), calls(r, "k", 900, 950, 999, 1000, 1500, 1999), "aligned windows reset at 1000");
    }

    static void testSlidingWindowLog() {
        RateLimiter r = new SlidingWindowLimiter(2, 1000, () -> now[0]);
        T.eq(List.of(true, true, false, true, false, true), calls(r, "k", 900, 950, 1000, 1900, 1949, 1950),
                "900 expires at 1900, 950 expires at 1950; rejected calls are not recorded");
    }

    static void testKeysAreIndependent() {
        now[0] = 0;
        for (RateLimiter r : List.of(new TokenBucketLimiter(1, 0, () -> now[0]), new FixedWindowLimiter(1, 1000, () -> now[0]), new SlidingWindowLimiter(1, 1000, () -> now[0]))) {
            T.check(r.allow("a") && !r.allow("a") && r.allow("b"), r.getClass().getSimpleName() + ": per-key budgets");
        }
    }

    static void testConcurrentCallersExactBudget() throws Exception {
        RateLimiter r = new TokenBucketLimiter(10, 0, () -> 0);
        AtomicInteger ok = new AtomicInteger();
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<?>> fs = new ArrayList<>();
        for (int i = 0; i < 100; i++) fs.add(pool.submit(() -> { go.await(); if (r.allow("hot")) ok.incrementAndGet(); return null; }));
        go.countDown();
        for (Future<?> f : fs) f.get(5, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.eq(10, ok.get(), "exactly capacity requests allowed under contention");
    }
}
