"Design a rate limiter" shows up as both an LLD and an HLD question. In LLD the interviewer wants **pluggable algorithms**, correct maths at window boundaries, **per-key** isolation, and thread safety.

## Interface

`allow(key)` returns `true` if the request may proceed, and `false` if it's rate-limited. Each key (user id, API key, IP) has its own independent budget. Time comes **only** from the injected clock (milliseconds).

## Algorithms

**Token bucket(capacity, refillPerSecond)**
Each key starts **full**. Tokens refill continuously at `refillPerSecond` (fractional tokens accumulate) up to `capacity`. A request takes one token.
*Good for:* allowing bursts while enforcing an average rate.

> **Floating-point trap:** with refill 1/s, calls at 300, 600, 900 and 1000 ms accumulate `0.3 + 0.3 + 0.3 + 0.1`, which is *slightly less* than `1.0` in IEEE-754. Compare with a small epsilon, or store integer micro-tokens. The tests check this.

**Fixed window(max, windowMs)**
Windows are aligned: `[0, w)`, `[w, 2w)`, … At most `max` requests per key per window.
*Weakness:* up to `2 × max` requests around a boundary. Know how to explain this.

**Sliding window log(max, windowMs)**
At most `max` requests in **any** span of `windowMs`. A request at time `t` counts while `now − t < windowMs`. Rejected requests aren't recorded.
*Trade-off:* exact, but O(max) memory per key.

## Test scenarios

- token bucket `capacity 3, refill 2/s`: 3 immediate requests OK, the 4th rejected. After 500 ms, one more OK.
- fixed window `2 per 1000 ms`: requests at 900, 950 OK, 999 rejected, 1000 OK (new window).
- sliding `2 per 1000 ms`: requests at 900, 950 OK, 1000 rejected, 1900 OK (900 expired), 1949 rejected.
- 100 concurrent callers on one key of a token bucket with capacity 10 and no refill → **exactly 10** allowed.

<div class="lang-note" data-lang="java">

Java: `interface RateLimiter { boolean allow(String key); }` and `interface Clock { long nowMillis(); }` are provided. Implement `TokenBucketLimiter(long capacity, double refillPerSecond, Clock)`, `FixedWindowLimiter(int max, long windowMillis, Clock)` and `SlidingWindowLimiter(int max, long windowMillis, Clock)`.

</div>

<div class="lang-note" data-lang="go">

Go: `type RateLimiter interface { Allow(key string) bool }` and `type Clock func() time.Time` are provided. Implement `NewTokenBucket(capacity int, refillPerSecond float64, c Clock) RateLimiter`, `NewFixedWindow(max int, window time.Duration, c Clock) RateLimiter` and `NewSlidingWindow(max int, window time.Duration, c Clock) RateLimiter`. Fixed windows align to multiples of `window` since the Unix epoch.

</div>

<div class="lang-note" data-lang="python">

Python: `RateLimiter` is an abstract base class with `allow(key) -> bool`, and a clock is a zero-argument callable returning milliseconds. Implement `TokenBucketLimiter(capacity, refill_per_second, clock)`, `FixedWindowLimiter(max_requests, window_ms, clock)` and `SlidingWindowLimiter(max_requests, window_ms, clock)`. Python floats are IEEE-754 doubles too, so the epsilon trap applies.

</div>
