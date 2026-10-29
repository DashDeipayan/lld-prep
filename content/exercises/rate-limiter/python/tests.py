import threading

from lldtest import check, eq
from solution import FixedWindowLimiter, SlidingWindowLimiter, TokenBucketLimiter


class FakeClock:
    def __init__(self, now=0):
        self.now = now

    def __call__(self):
        return self.now


def calls(limiter, clock, key, *at):
    out = []
    for t in at:
        clock.now = t
        out.append(limiter.allow(key))
    return out


def test_token_bucket_burst_and_refill():
    c = FakeClock()
    r = TokenBucketLimiter(3, 2, c)
    eq([True, True, True, False], calls(r, c, "k", 0, 0, 0, 0), "burst of capacity, then reject")
    eq([True, False], calls(r, c, "k", 500, 500), "2 tokens/s -> one token after 500ms")
    eq([True, True, True, False], calls(r, c, "k", 10_000, 10_000, 10_000, 10_000), "refill caps at capacity")


def test_token_bucket_fractional_accumulation():
    c = FakeClock()
    r = TokenBucketLimiter(1, 1, c)
    eq([True, False, False, False, True], calls(r, c, "k", 0, 300, 600, 900, 1000), "fractions add up to one token at 1000ms")


def test_fixed_window_boundaries():
    c = FakeClock()
    r = FixedWindowLimiter(2, 1000, c)
    eq([True, True, False, True, True, False], calls(r, c, "k", 900, 950, 999, 1000, 1500, 1999), "aligned windows reset at 1000")


def test_sliding_window_log():
    c = FakeClock()
    r = SlidingWindowLimiter(2, 1000, c)
    eq([True, True, False, True, False, True], calls(r, c, "k", 900, 950, 1000, 1900, 1949, 1950),
       "900 expires at 1900, 950 expires at 1950; rejected calls are not recorded")


def test_keys_are_independent():
    c = FakeClock()
    for r in [TokenBucketLimiter(1, 0, c), FixedWindowLimiter(1, 1000, c), SlidingWindowLimiter(1, 1000, c)]:
        check(r.allow("a") and not r.allow("a") and r.allow("b"), type(r).__name__ + ": per-key budgets")


def test_concurrent_callers_exact_budget():
    r = TokenBucketLimiter(10, 0, lambda: 0)
    start = threading.Barrier(100)
    allowed = []

    def caller():
        start.wait()
        if r.allow("hot"):
            allowed.append(1)

    threads = [threading.Thread(target=caller) for _ in range(100)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(5)
    eq(10, len(allowed), "exactly capacity requests allowed under contention")
