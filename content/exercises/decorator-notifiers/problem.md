Your notification `Sender` is flaky, and marketing keeps spamming people. Add behaviour **without modifying** any existing sender. Each concern is a decorator that wraps a `Sender` and is itself a `Sender`.

## Decorators

**Retrying(inner, maxAttempts)**
Calls `inner` up to `maxAttempts` times and stops at the first success. If every attempt fails, it fails with the **last** error. `maxAttempts < 1` is invalid.

**RateLimited(inner, clock, maxPerWindow, window)**
At most `maxPerWindow` **successful** sends per **recipient** in any sliding window of length `window`. A send at time `t` counts against later sends while `now - t < window`. When the limit is hit, it fails with a *rate limited* error and **does not call inner**. Failed sends don't consume quota.

**Logging(inner, log)**
Appends `"OK <to>: <msg>"` or `"FAIL <to>: <msg>"` to `log` and passes the result or error through unchanged.

## Composition matters

```text
RateLimited( Retrying( flakySms, 3 ), clock, 2, 1h )
```

A message that needs 3 attempts counts **once** against the rate limit. Your design should make that ordering obvious.

<div class="lang-note" data-lang="java">

Java: `Sender.send(to, msg)` throws `RuntimeException` on failure. `RateLimitedException` is provided. `Clock` is an interface: `long nowMillis()`. Window and timestamps are in milliseconds. Invalid `maxAttempts` throws `IllegalArgumentException`.

</div>

<div class="lang-note" data-lang="go">

Go: `Sender.Send(to, msg string) error`. Constructors: `WithRetry(inner Sender, maxAttempts int) Sender` (panics if `maxAttempts < 1`), `WithRateLimit(inner Sender, clock Clock, max int, window time.Duration) Sender`, `WithLogging(inner Sender, log *[]string) Sender`. `type Clock func() time.Time`. Rate limit errors wrap `ErrRateLimited`.

</div>

<div class="lang-note" data-lang="python">

Python: `Sender` is an abstract base class whose `send(to, msg)` raises an exception on failure. The classes are `RetryingSender(inner, max_attempts)` (raises `ValueError` if `max_attempts < 1`), `RateLimitedSender(inner, clock, max_per_window, window_ms)` and `LoggingSender(inner, log)`. A clock is any zero-argument callable returning milliseconds. `RateLimitedError` is provided. Python's `@decorator` functions are a different thing from this design pattern, although both wrap a callable.

</div>
