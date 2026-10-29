This is the heart of BookMyShow, Ticketmaster and any slot-booking system. During checkout a user **holds** seats for a few minutes, then pays (**confirm**) or abandons (**release** or expiry). Two users must **never** end up with the same seat.

## API

`SeatInventory(seatIds, holdTtl, clock)`. Time comes **only** from the injected clock.

| Operation | Behaviour |
|---|---|
| `hold(userId, seatIds)` → holdId | **All or nothing.** If every seat is available, they all become HELD until `now + ttl` and a new hold id is returned (`"H1"`, `"H2"`, …). If any seat is unavailable, nothing changes and it fails with *seat unavailable* |
| `confirm(holdId)` | the hold must be active (not expired, released or confirmed). Its seats become BOOKED permanently |
| `release(holdId)` | an active hold's seats become AVAILABLE again |
| `status(seatId)` | `"AVAILABLE"`, `"HELD"` or `"BOOKED"`. An expired hold reads as `"AVAILABLE"` |
| `available()` | available seat ids, **sorted** |

A seat is **available** at time `now` if it's AVAILABLE, or HELD with `expiresAt <= now`.

Invalid input (unknown seat, an empty list, the same seat listed twice) fails with *invalid / unknown seat*. `confirm` or `release` on an unknown, expired, released or already-confirmed hold fails with *hold not active*.

## The concurrency test

64 threads (goroutines in Go) start at the same instant, each trying to hold 3 of 12 seats chosen from overlapping sets. When they finish, **no seat may belong to two successful holds**, and every successful hold's seats must read as HELD.

<div class="lang-note" data-lang="java">

Java: `Clock` is `long nowMillis()`, TTL is in millis. Errors: `SeatUnavailableException` (provided) for unavailable seats. `IllegalArgumentException` for invalid or unknown seats. `IllegalStateException` for a hold that isn't active.

</div>

<div class="lang-note" data-lang="go">

Go: `NewSeatInventory(seatIDs []string, ttl time.Duration, clock func() time.Time) *SeatInventory`. `Hold(user string, seats []string) (string, error)`, `Confirm(id) error`, `Release(id) error`, `Status(seat) string`, `Available() []string`. Errors: `ErrSeatUnavailable`, `ErrUnknownSeat` (also used for empty or duplicate lists), `ErrHoldNotActive`.

</div>

<div class="lang-note" data-lang="python">

Python: `SeatInventory(seat_ids, hold_ttl_ms, clock)` where `clock` is a zero-argument callable returning milliseconds. Methods: `hold(user_id, seat_ids)`, `confirm(hold_id)`, `release(hold_id)`, `status(seat_id)` and `available()`. Errors: `SeatUnavailableError` and `HoldNotActiveError` (both provided), and `ValueError` for invalid or unknown seats.

</div>

## Discuss after coding

- Granularity: is one lock per show fine? When would you shard?
- How does this map to a database with several app servers? (`UPDATE … WHERE status='AVAILABLE' OR expires_at <= now()` and check the affected row count, or a unique constraint on bookings.)
- What if the payment callback arrives **after** the hold expired? (Refund, or extend the hold when payment starts.)
