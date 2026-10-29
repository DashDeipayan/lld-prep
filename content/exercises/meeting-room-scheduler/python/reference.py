from __future__ import annotations

import bisect
import threading
from dataclasses import dataclass
from typing import Dict, List


# ======================= Provided =======================
@dataclass(frozen=True)
class Booking:
    id: str
    room: str
    organizer: str
    start: int
    end: int


@dataclass(frozen=True)
class Slot:
    start: int
    end: int


class NoRoomAvailableError(Exception):
    pass


class ConflictError(Exception):
    """The room is taken, or the booking was already cancelled."""


# ======================= Model =======================
class _Room:
    def __init__(self, room_id: str, capacity: int) -> None:
        self.id = room_id
        self.capacity = capacity
        self.starts: List[int] = []  # sorted; parallel to bookings
        self.bookings: List[Booking] = []

    def is_free(self, start: int, end: int) -> bool:
        """Bookings never overlap, so only the latest one starting before `end` can clash with [start, end)."""
        i = bisect.bisect_left(self.starts, end) - 1
        return i < 0 or self.bookings[i].end <= start

    def add(self, b: Booking) -> None:
        i = bisect.bisect_left(self.starts, b.start)
        self.starts.insert(i, b.start)
        self.bookings.insert(i, b)

    def remove(self, b: Booking) -> None:
        i = bisect.bisect_left(self.starts, b.start)
        del self.starts[i]
        del self.bookings[i]


def _best_fit(room: _Room):
    return (room.capacity, room.id)  # smallest capacity, ties -> lowest id


def _validate(start: int, end: int) -> None:
    if start < 0 or end > 1440 or start >= end:
        raise ValueError("invalid interval [%d, %d)" % (start, end))


class Scheduler:
    def __init__(self) -> None:
        self._rooms: Dict[str, _Room] = {}
        self._bookings: Dict[str, Booking] = {}
        self._next = 1
        self._lock = threading.Lock()  # check-then-book must be atomic

    def add_room(self, room_id: str, capacity: int) -> None:
        with self._lock:
            if capacity < 1 or room_id in self._rooms:
                raise ValueError("invalid room " + room_id)
            self._rooms[room_id] = _Room(room_id, capacity)

    def book(self, organizer: str, start: int, end: int, attendees: int) -> Booking:
        _validate(start, end)
        if attendees < 1:
            raise ValueError("attendees must be >= 1")
        with self._lock:
            fits = [r for r in self._rooms.values() if r.capacity >= attendees and r.is_free(start, end)]
            if not fits:
                raise NoRoomAvailableError("no room for %d at %d" % (attendees, start))
            return self._place(min(fits, key=_best_fit), organizer, start, end)

    def book_room(self, room_id: str, organizer: str, start: int, end: int) -> Booking:
        _validate(start, end)
        with self._lock:
            room = self._room(room_id)
            if not room.is_free(start, end):
                raise ConflictError("conflict in " + room_id)
            return self._place(room, organizer, start, end)

    def cancel(self, booking_id: str) -> None:
        with self._lock:
            b = self._bookings.pop(booking_id, None)
            if b is None:
                raise ConflictError("unknown or cancelled booking " + booking_id)
            self._rooms[b.room].remove(b)

    def free_rooms(self, start: int, end: int, attendees: int) -> List[str]:
        _validate(start, end)
        with self._lock:
            fits = [r for r in self._rooms.values() if r.capacity >= attendees and r.is_free(start, end)]
            return [r.id for r in sorted(fits, key=_best_fit)]

    def schedule(self, room_id: str) -> List[Booking]:
        with self._lock:
            return list(self._room(room_id).bookings)

    def free_slots(self, room_id: str, start: int, end: int) -> List[Slot]:
        _validate(start, end)
        with self._lock:
            out = []
            cursor = start
            for b in self._room(room_id).bookings:
                if b.end <= cursor:
                    continue
                if b.start >= end:
                    break
                if b.start > cursor:
                    out.append(Slot(cursor, b.start))
                cursor = max(cursor, b.end)
            if cursor < end:
                out.append(Slot(cursor, end))
            return out

    def _place(self, room: _Room, organizer: str, start: int, end: int) -> Booking:
        b = Booking("M%d" % self._next, room.id, organizer, start, end)
        self._next += 1
        room.add(b)
        self._bookings[b.id] = b
        return b

    def _room(self, room_id: str) -> _Room:
        room = self._rooms.get(room_id)
        if room is None:
            raise ValueError("unknown room " + room_id)
        return room
