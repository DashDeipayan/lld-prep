from lldtest import eq, raises
from solution import LoggingSender, RateLimitedError, RateLimitedSender, RetryingSender, Sender


class Flaky(Sender):
    """Fails the first `failures` calls, then succeeds; records successful deliveries."""

    def __init__(self, failures):
        self.calls = 0
        self.failures = failures
        self.delivered = []

    def send(self, to, msg):
        self.calls += 1
        if self.calls <= self.failures:
            raise RuntimeError("fail #%d" % self.calls)
        self.delivered.append("%s:%s" % (to, msg))


class FakeClock:
    def __init__(self, now=0):
        self.now = now

    def __call__(self):
        return self.now


def test_retry_succeeds_within_attempts():
    f = Flaky(2)
    RetryingSender(f, 3).send("a", "hi")
    eq(3, f.calls, "two failures then success")
    eq(["a:hi"], f.delivered, "delivered once")


def test_retry_gives_up_with_last_error():
    f = Flaky(5)
    e = raises(RuntimeError, lambda: RetryingSender(f, 3).send("a", "hi"), "all attempts fail")
    eq("fail #3", str(e), "last error is re-raised")
    eq(3, f.calls, "exactly max_attempts calls")
    raises(ValueError, lambda: RetryingSender(f, 0), "max_attempts < 1")


def test_rate_limit_per_recipient_sliding_window():
    clock = FakeClock(1_000)
    f = Flaky(0)
    s = RateLimitedSender(f, clock, 2, 1_000)
    s.send("a", "1")
    clock.now += 400
    s.send("a", "2")
    raises(RateLimitedError, lambda: s.send("a", "3"), "third within window")
    s.send("b", "1")
    clock.now = 2_000
    s.send("a", "4")
    raises(RateLimitedError, lambda: s.send("a", "5"), "send at 1400 still in window at 2000")
    clock.now = 2_400
    s.send("a", "6")
    eq(["a:1", "a:2", "b:1", "a:4", "a:6"], f.delivered, "deliveries")
    eq(5, f.calls, "rejected sends never reach inner")


def test_failed_sends_do_not_consume_quota():
    f = Flaky(1)
    s = RateLimitedSender(f, FakeClock(0), 1, 1_000)
    raises(RuntimeError, lambda: s.send("a", "x"), "inner failure passes through")
    s.send("a", "y")
    eq(["a:y"], f.delivered, "quota still available after a failure")


def test_logging_passes_through():
    log = []
    s = LoggingSender(Flaky(1), log)
    raises(RuntimeError, lambda: s.send("a", "x"), "error re-raised")
    s.send("a", "y")
    eq(["FAIL a: x", "OK a: y"], log, "log lines")


def test_composition_retry_inside_rate_limit():
    f = Flaky(2)
    log = []
    s = LoggingSender(RateLimitedSender(RetryingSender(f, 3), FakeClock(0), 1, 1_000), log)
    s.send("a", "m1")
    raises(RateLimitedError, lambda: s.send("a", "m2"), "retries counted once against quota")
    eq(["OK a: m1", "FAIL a: m2"], log, "logging sees the decorated result")
    eq(3, f.calls, "inner called only for m1")
