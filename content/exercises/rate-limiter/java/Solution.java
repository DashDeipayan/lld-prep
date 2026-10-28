import java.util.*;
import java.util.concurrent.*;

interface RateLimiter {
    boolean allow(String key);
}

interface Clock {
    long nowMillis();
}

final class TokenBucketLimiter implements RateLimiter {
    TokenBucketLimiter(long capacity, double refillPerSecond, Clock clock) { /* TODO */ }
    public boolean allow(String key) { throw new UnsupportedOperationException("TODO"); }
}

final class FixedWindowLimiter implements RateLimiter {
    FixedWindowLimiter(int max, long windowMillis, Clock clock) { /* TODO */ }
    public boolean allow(String key) { throw new UnsupportedOperationException("TODO"); }
}

final class SlidingWindowLimiter implements RateLimiter {
    SlidingWindowLimiter(int max, long windowMillis, Clock clock) { /* TODO */ }
    public boolean allow(String key) { throw new UnsupportedOperationException("TODO"); }
}
