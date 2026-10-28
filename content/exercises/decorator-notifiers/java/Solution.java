import java.util.*;

interface Sender {
    /** @throws RuntimeException on failure */
    void send(String to, String msg);
}

interface Clock {
    long nowMillis();
}

final class RateLimitedException extends RuntimeException {
    RateLimitedException(String m) { super(m); }
}

final class RetryingSender implements Sender {
    RetryingSender(Sender inner, int maxAttempts) { /* TODO */ }
    public void send(String to, String msg) { throw new UnsupportedOperationException("TODO"); }
}

final class RateLimitedSender implements Sender {
    RateLimitedSender(Sender inner, Clock clock, int maxPerWindow, long windowMillis) { /* TODO */ }
    public void send(String to, String msg) { throw new UnsupportedOperationException("TODO"); }
}

final class LoggingSender implements Sender {
    LoggingSender(Sender inner, List<String> log) { /* TODO */ }
    public void send(String to, String msg) { throw new UnsupportedOperationException("TODO"); }
}
