Build the booking core of BookMyShow. Read the walkthrough lesson first. It explains every decision tested here.

## Catalogue

- `addTheatre(theatreId, city)`
- `addScreen(theatreId, screenId, regularRows, premiumRows, seatsPerRow)`: rows are lettered from `A`, **regular rows first**, then premium rows. Seat ids are `row + number` (for example `A1`). With 2 regular and 1 premium row of 3 seats, the regular seats are `A1–A3` and `B1–B3`, and the premium seats are `C1–C3`.
- `addMovie(id, title, durationMinutes)`

## Shows

`addShow(movieId, screenId, startMillis, regularPrice, premiumPrice)` → `"S1"`, `"S2"`, …

- A show occupies `[start, start + duration)`. Overlapping shows on the **same screen** are rejected.
- `findShows(city, movieId)` → ids of shows of that movie in that city that **haven't started** (`start > now`), sorted by start time.
- `availableSeats(showId)` → sorted seat ids.

## Booking

- `hold(showId, user, seatIds)` → `"H1"`, … All-or-nothing, valid for `holdTtl` (lazy expiry, as on Day 5). You can't hold seats for a show that has started.
- `confirm(holdId)` → `Booking(id "B1"…, showId, user, seats, amount, "CONFIRMED")`. The amount is the sum of seat prices by type. The **payment gateway** is called with `(user, amount)`:
  - charge fails → the seats are **released**, *payment failed* error
  - hold expired, already used or unknown → *hold not active*
- `cancel(bookingId)` → the booking becomes `"CANCELLED"` and its seats are available again. Only before the show starts. Cancelling twice is an error.

## Errors

| Situation | Java | Go | Python |
|---|---|---|---|
| unknown theatre/screen/movie/show/seat, invalid input | `IllegalArgumentException` | `ErrNotFound` / `ErrInvalid` | `ValueError` |
| show overlap | `IllegalStateException` | `ErrOverlap` | `BookingStateError` |
| seat unavailable | `SeatUnavailableException` | `ErrSeatUnavailable` | `SeatUnavailableError` |
| hold not active | `IllegalStateException` | `ErrHoldNotActive` | `HoldNotActiveError` |
| payment failed | `PaymentFailedException` | `ErrPaymentFailed` | `PaymentFailedError` |
| show already started (hold/cancel), double cancel | `IllegalStateException` | `ErrShowStarted` / `ErrInvalid` | `BookingStateError` |

<div class="lang-note" data-lang="java">

Provided: `interface PaymentGateway { boolean charge(String user, long amount); }`, `interface Clock { long nowMillis(); }`, `record Booking(String id, String showId, String user, List<String> seats, long amount, String status)` (seats sorted), and the two exception classes. Constructor: `BookingService(PaymentGateway gateway, Clock clock, long holdTtlMillis)`.

</div>

<div class="lang-note" data-lang="go">

Provided: `type PaymentGateway interface { Charge(user string, amount int64) bool }`, `type Booking struct { ID, ShowID, User string; Seats []string; Amount int64; Status string }`. `NewBookingService(gw PaymentGateway, clock func() time.Time, holdTTL time.Duration)`. Times are `time.Time`, duration in `time.Duration`. Methods mirror Java with Go errors: `AddTheatre`, `AddScreen`, `AddMovie(id, title string, minutes int)`, `AddShow(movieID, screenID string, start time.Time, regular, premium int64) (string, error)`, `FindShows`, `AvailableSeats`, `Hold`, `Confirm`, `Cancel`.

</div>

<div class="lang-note" data-lang="python">

Provided: the `PaymentGateway` abstract base class (`charge(user, amount) -> bool`), the `Booking(id, show_id, user, seats, amount, status)` dataclass, and the error classes in the table above. The clock is a zero-argument callable returning milliseconds. Constructor: `BookingService(gateway, clock, hold_ttl_ms)`. Methods are snake_case: `add_theatre`, `add_screen`, `add_movie`, `add_show(movie_id, screen_id, start_ms, regular_price, premium_price)`, `find_shows`, `available_seats`, `hold`, `confirm`, `cancel`.

</div>
