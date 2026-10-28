import java.util.*;

interface Sender {
    void send(String to, String msg);
}

interface Clock {
    long nowMillis();
}

final class RateLimitedException extends RuntimeException {
    RateLimitedException(String m) { super(m); }
}

final class RetryingSender implements Sender {
    private final Sender inner;
    private final int maxAttempts;

    RetryingSender(Sender inner, int maxAttempts) {
        if (maxAttempts < 1) throw new IllegalArgumentException("maxAttempts must be >= 1");
        this.inner = Objects.requireNonNull(inner);
        this.maxAttempts = maxAttempts;
    }

    public void send(String to, String msg) {
        RuntimeException last = null;
        for (int i = 0; i < maxAttempts; i++) {
            try {
                inner.send(to, msg);
                return;
            } catch (RuntimeException e) {
                last = e;
            }
        }
        throw last;
    }
}

final class RateLimitedSender implements Sender {
    private final Sender inner;
    private final Clock clock;
    private final int max;
    private final long window;
    private final Map<String, Deque<Long>> sent = new HashMap<>();

    RateLimitedSender(Sender inner, Clock clock, int maxPerWindow, long windowMillis) {
        this.inner = inner;
        this.clock = clock;
        this.max = maxPerWindow;
        this.window = windowMillis;
    }

    public synchronized void send(String to, String msg) {
        long now = clock.nowMillis();
        Deque<Long> times = sent.computeIfAbsent(to, k -> new ArrayDeque<>());
        while (!times.isEmpty() && now - times.peekFirst() >= window) times.pollFirst();
        if (times.size() >= max) throw new RateLimitedException("rate limited: " + to);
        inner.send(to, msg);          // throws => quota not consumed
        times.addLast(now);
    }
}

final class LoggingSender implements Sender {
    private final Sender inner;
    private final List<String> log;

    LoggingSender(Sender inner, List<String> log) {
        this.inner = inner;
        this.log = log;
    }

    public void send(String to, String msg) {
        try {
            inner.send(to, msg);
            log.add("OK " + to + ": " + msg);
        } catch (RuntimeException e) {
            log.add("FAIL " + to + ": " + msg);
            throw e;
        }
    }
}
