from __future__ import annotations

import threading
from dataclasses import dataclass
from enum import Enum
from typing import Dict


# ---- Provided ----
class RoomType(Enum):
    SINGLE = "SINGLE"
    DOUBLE = "DOUBLE"
    SUITE = "SUITE"


class NoRoomAvailableError(Exception):
    pass


class BookingStatusError(Exception):
    """The booking isn't in the right state for this step (or it's the wrong check-in day)."""


# ---- Model ----
@dataclass(frozen=True)
class _Room:
    number: str
    type: RoomType
    rate: int


class _Status(Enum):
    BOOKED = "BOOKED"
    CHECKED_IN = "CHECKED_IN"
    CHECKED_OUT = "CHECKED_OUT"
    CANCELLED = "CANCELLED"


class _Booking:
    def __init__(self, booking_id: str, guest: str, room: _Room, check_in: int, check_out: int) -> None:
        self.id, self.guest, self.room, self.check_in, self.check_out = booking_id, guest, room, check_in, check_out
        self.status = _Status.BOOKED
        self.charges = 0

    def blocks(self, start: int, end: int) -> bool:
        """Only live bookings hold the room; stays are half-open [check_in, check_out)."""
        return self.status in (_Status.BOOKED, _Status.CHECKED_IN) and start < self.check_out and self.check_in < end


def _validate(check_in: int, check_out: int) -> None:
    if check_in < 0 or check_in >= check_out:
        raise ValueError("invalid stay")


class Hotel:
    def __init__(self) -> None:
        self._rooms: Dict[str, _Room] = {}
        self._bookings: Dict[str, _Booking] = {}
        self._next = 1
        self._lock = threading.Lock()

    def add_room(self, number: str, room_type: RoomType, nightly_rate: int) -> None:
        with self._lock:
            if number in self._rooms or nightly_rate <= 0:
                raise ValueError("invalid room " + number)
            self._rooms[number] = _Room(number, room_type, nightly_rate)

    def availability(self, check_in: int, check_out: int) -> Dict[RoomType, int]:
        _validate(check_in, check_out)
        with self._lock:
            out = {t: 0 for t in RoomType}  # every type present, possibly 0
            for r in self._rooms.values():
                if self._free(r, check_in, check_out):
                    out[r.type] += 1
            return out

    def book(self, guest: str, room_type: RoomType, check_in: int, check_out: int) -> str:
        _validate(check_in, check_out)
        with self._lock:
            for number in sorted(self._rooms):  # lowest number first
                r = self._rooms[number]
                if r.type is room_type and self._free(r, check_in, check_out):
                    b = _Booking("BK%d" % self._next, guest, r, check_in, check_out)
                    self._next += 1
                    self._bookings[b.id] = b
                    return b.id
            raise NoRoomAvailableError("no %s for %d-%d" % (room_type.value, check_in, check_out))

    def room_of(self, booking_id: str) -> str:
        with self._lock:
            return self._booking(booking_id).room.number

    def check_in(self, booking_id: str, day: int) -> None:
        with self._lock:
            b = self._booking(booking_id)
            self._require(b, _Status.BOOKED)
            if day != b.check_in:
                raise BookingStatusError("check-in is on day %d" % b.check_in)
            b.status = _Status.CHECKED_IN

    def add_charge(self, booking_id: str, amount: int) -> None:
        with self._lock:
            b = self._booking(booking_id)
            self._require(b, _Status.CHECKED_IN)
            if amount <= 0:
                raise ValueError("amount must be positive")
            b.charges += amount

    def check_out(self, booking_id: str) -> int:
        with self._lock:
            b = self._booking(booking_id)
            self._require(b, _Status.CHECKED_IN)
            b.status = _Status.CHECKED_OUT
            return (b.check_out - b.check_in) * b.room.rate + b.charges

    def cancel(self, booking_id: str) -> None:
        with self._lock:
            b = self._booking(booking_id)
            self._require(b, _Status.BOOKED)
            b.status = _Status.CANCELLED

    def status(self, booking_id: str) -> str:
        with self._lock:
            return self._booking(booking_id).status.value

    def _free(self, room: _Room, start: int, end: int) -> bool:
        return not any(b.room is room and b.blocks(start, end) for b in self._bookings.values())

    @staticmethod
    def _require(b: _Booking, status: _Status) -> None:
        if b.status is not status:
            raise BookingStatusError("booking is %s, expected %s" % (b.status.value, status.value))

    def _booking(self, booking_id: str) -> _Booking:
        b = self._bookings.get(booking_id)
        if b is None:
            raise ValueError("unknown booking " + booking_id)
        return b
