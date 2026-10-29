**Interviewer:** *"Design the reservation system for a single hotel. Guests book a room type for some nights, the front desk checks them in and out, and room-service charges go on the bill."*

## Requirements

| Operation | Rules |
|---|---|
| `addRoom(number, type, nightlyRate)` | type ∈ SINGLE, DOUBLE, SUITE. Room numbers are strings that compare lexicographically (`"101"`, `"102"`, …) |
| `availability(checkIn, checkOut)` | count of free rooms **per type** for the whole stay (every type present, possibly 0) |
| `book(guest, type, checkIn, checkOut)` → booking id | assigns the **lowest-numbered** free room of that type. Ids `"BK1"`, … No room → *unavailable* |
| `roomOf(bookingId)` | the assigned room number |
| `checkIn(bookingId, day)` | only from BOOKED, and only **on** the check-in day |
| `addCharge(bookingId, amount)` | only while CHECKED_IN, `amount > 0` |
| `checkOut(bookingId)` → bill | only from CHECKED_IN. `bill = nights × rate + charges`. The room is freed |
| `cancel(bookingId)` | only from BOOKED. The room is freed |
| `status(bookingId)` | `"BOOKED"`, `"CHECKED_IN"`, `"CHECKED_OUT"` or `"CANCELLED"` |

Nights are integer day numbers, and a stay covers `[checkIn, checkOut)`. Invalid ranges and unknown ids are invalid input.

**Concurrency:** 20 guests booking the last free SUITE at the same moment results in exactly **one** booking.

<div class="lang-note" data-lang="java">

Provided: `enum RoomType { SINGLE, DOUBLE, SUITE }`, `NoRoomAvailableException`. `availability` returns `Map<RoomType, Integer>`. Invalid input throws `IllegalArgumentException`. A wrong lifecycle step throws `IllegalStateException`.

</div>

<div class="lang-note" data-lang="go">

Provided: `RoomType` (`Single`, `Double`, `Suite`). `NewHotel()`. `Availability(in, out int) (map[RoomType]int, error)`. Errors: `ErrInvalid`, `ErrNotFound`, `ErrNoRoom`, `ErrBadStatus`.

</div>

<div class="lang-note" data-lang="python">

Provided: the `RoomType` enum, plus `NoRoomAvailableError` and `BookingStatusError` (for a wrong lifecycle step or wrong check-in day). `availability` returns a `dict` from every `RoomType` to a count. Invalid input raises `ValueError`. Methods are snake_case: `add_room`, `room_of`, `check_in`, `add_charge`, `check_out`.

</div>
