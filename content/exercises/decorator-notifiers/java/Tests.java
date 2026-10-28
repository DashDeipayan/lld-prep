import java.util.*;

public class Tests {
    /** Fails the first `failures` calls, then succeeds; records successful deliveries. */
    static final class Flaky implements Sender {
        int calls, failures;
        final List<String> delivered = new ArrayList<>();
        Flaky(int failures) { this.failures = failures; }
        public void send(String to, String msg) {
            calls++;
            if (calls <= failures) throw new IllegalStateException("fail #" + calls);
            delivered.add(to + ":" + msg);
        }
    }

    static void testRetrySucceedsWithinAttempts() {
        Flaky f = new Flaky(2);
        new RetryingSender(f, 3).send("a", "hi");
        T.eq(3, f.calls, "two failures then success");
        T.eq(List.of("a:hi"), f.delivered, "delivered once");
    }

    static void testRetryGivesUpWithLastError() {
        Flaky f = new Flaky(5);
        IllegalStateException e = T.throwsEx(IllegalStateException.class, () -> new RetryingSender(f, 3).send("a", "hi"), "all attempts fail");
        T.eq("fail #3", e.getMessage(), "last error is rethrown");
        T.eq(3, f.calls, "exactly maxAttempts calls");
        T.throwsEx(IllegalArgumentException.class, () -> new RetryingSender(f, 0), "maxAttempts < 1");
    }

    static void testRateLimitPerRecipientSlidingWindow() {
        long[] now = {1_000};
        Flaky f = new Flaky(0);
        Sender s = new RateLimitedSender(f, () -> now[0], 2, 1_000);
        s.send("a", "1");
        now[0] += 400;
        s.send("a", "2");
        T.throwsEx(RateLimitedException.class, () -> s.send("a", "3"), "third within window");
        s.send("b", "1");
        now[0] = 2_000;
        s.send("a", "4");
        T.throwsEx(RateLimitedException.class, () -> s.send("a", "5"), "send at 1400 still in window at 2000");
        now[0] = 2_400;
        s.send("a", "6");
        T.eq(List.of("a:1", "a:2", "b:1", "a:4", "a:6"), f.delivered, "deliveries");
        T.eq(5, f.calls, "rejected sends never reach inner");
    }

    static void testFailedSendsDoNotConsumeQuota() {
        long[] now = {0};
        Flaky f = new Flaky(1);
        Sender s = new RateLimitedSender(f, () -> now[0], 1, 1_000);
        T.throwsEx(IllegalStateException.class, () -> s.send("a", "x"), "inner failure passes through");
        s.send("a", "y");
        T.eq(List.of("a:y"), f.delivered, "quota still available after a failure");
    }

    static void testLoggingPassesThrough() {
        List<String> log = new ArrayList<>();
        Flaky f = new Flaky(1);
        Sender s = new LoggingSender(f, log);
        T.throwsEx(IllegalStateException.class, () -> s.send("a", "x"), "error rethrown");
        s.send("a", "y");
        T.eq(List.of("FAIL a: x", "OK a: y"), log, "log lines");
    }

    static void testCompositionRetryInsideRateLimit() {
        long[] now = {0};
        Flaky f = new Flaky(2);
        List<String> log = new ArrayList<>();
        Sender s = new LoggingSender(new RateLimitedSender(new RetryingSender(f, 3), () -> now[0], 1, 1_000), log);
        s.send("a", "m1");
        T.throwsEx(RateLimitedException.class, () -> s.send("a", "m2"), "retries counted once against quota");
        T.eq(List.of("OK a: m1", "FAIL a: m2"), log, "logging sees the decorated result");
        T.eq(3, f.calls, "inner called only for m1");
    }
}
