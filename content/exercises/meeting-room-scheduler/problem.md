Design a meeting room booking system. Times are **minutes of the day** (`0–1440`), and intervals are half-open `[start, end)`: a meeting ending at 10:00 doesn't conflict with one starting at 10:00.

## API

| Operation | Behaviour |
|---|---|
| `addRoom(id, capacity)` | duplicate → error |
| `book(organizer, start, end, attendees)` → booking | **best fit:** among rooms with `capacity ≥ attendees` that are free for the whole interval, choose the **smallest capacity**, ties → lowest id. No room → *no room available* |
| `bookRoom(roomId, organizer, start, end)` → booking | book a specific room. Conflict → *conflict* |
| `cancel(bookingId)` | frees the slot. Unknown or already cancelled → error |
| `freeRooms(start, end, attendees)` | ids of rooms that fit and are free, ordered by (capacity, id) |
| `schedule(roomId)` | that room's active bookings sorted by start |
| `freeSlots(roomId, from, to)` | the gaps within `[from, to)` not covered by bookings, as `Slot(start, end)` in order. Empty gaps are omitted |

Booking ids are `"M1"`, `"M2"`, … An invalid interval (`start < 0`, `end > 1440`, `start ≥ end`) or `attendees < 1` is invalid input.

## Concurrency (tested)

50 threads try `bookRoom("R1", …, 600, 660)` at the same moment. Exactly **one** succeeds.

<div class="lang-note" data-lang="java">

Provided: `record Booking(String id, String room, String organizer, int start, int end)` and `record Slot(int start, int end)`, plus `NoRoomAvailableException`. Invalid input or unknown ids throw `IllegalArgumentException`. Conflict or double cancel throws `IllegalStateException`.

</div>

<div class="lang-note" data-lang="go">

Provided: `Booking{ID, Room, Organizer string; Start, End int}`, `Slot{Start, End int}`. `NewScheduler()`. Errors: `ErrInvalid`, `ErrNotFound`, `ErrConflict`, `ErrNoRoom`.

</div>

<div class="lang-note" data-lang="python">

Provided: the `Booking(id, room, organizer, start, end)` and `Slot(start, end)` dataclasses, plus `NoRoomAvailableError` and `ConflictError`. Methods are snake_case (`add_room`, `book_room`, `free_rooms`, `free_slots(room_id, start, end)`). Invalid input or unknown ids raise `ValueError`. A conflict or double cancel raises `ConflictError`. The `bisect` module keeps each room's bookings sorted by start.

</div>

## Follow-ups

- Recurring meetings? (Expand them into occurrences, and either book all or none, the same all-or-nothing idea as seat holds.)
- Multiple days? (Use `LocalDateTime` / `time.Time`. The interval logic is identical.)
- Equipment requirements (projector, VC)? (A room-filter Specification chain in front of the best-fit strategy.)
