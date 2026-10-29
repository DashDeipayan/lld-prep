from solution import FixedWindowLimiter, SlidingWindowLimiter, TokenBucketLimiter

now = [0]
clock = lambda: now[0]  # noqa: E731
limiters = {
    "token-bucket(3, 2/s)": TokenBucketLimiter(3, 2, clock),
    "fixed-window(3/s)": FixedWindowLimiter(3, 1000, clock),
    "sliding-log(3/s)": SlidingWindowLimiter(3, 1000, clock),
}
times = [0, 100, 200, 300, 600, 1000, 1100, 1250]
for name, limiter in limiters.items():
    row = []
    for t in times:
        now[0] = t
        row.append("%d%s" % (t, "✓" if limiter.allow("user-1") else "✗"))
    print("%-22s%s" % (name, " ".join(row)))
