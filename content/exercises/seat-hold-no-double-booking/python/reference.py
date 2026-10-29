from __future__ import annotations

import threading
from dataclasses import dataclass
from enum import Enum
from typing import Callable, Dict, List, Optional, Tuple

# A clock is a zero-argument function returning the current time in milliseconds.
Clock = Callable[[], int]


class SeatUnavailableError(Exception):
    """Provided: one of the requested seats is held or booked."""


class HoldNotActiveError(Exception):
    """Provided: the hold is unknown, expired, released or already confirmed."""


class Status(Enum):
    AVAILABLE = "AVAILABLE"
    HELD = "HELD"
    BOOKED = "BOOKED"


@dataclass
class _Seat:
    status: Status = Status.AVAILABLE
    hold_id: Optional[str] = None
    expires_at: int = 0

    def is_available(self, now: int) -> bool:
        # Lazy expiry: no background sweeper, a stale hold simply reads as available.
        return self.status is Status.AVAILABLE or (self.status is Status.HELD and self.expires_at <= now)


@dataclass
class _Hold:
    seats: Tuple[str, ...]
    expires_at: int
    active: bool = True


class SeatInventory:
    def __init__(self, seat_ids: List[str], hold_ttl_ms: int, clock: Clock) -> None:
        self._seats: Dict[str, _Seat] = {sid: _Seat() for sid in seat_ids}
        self._holds: Dict[str, _Hold] = {}
        self._ttl = hold_ttl_ms
        self._clock = clock
        self._next_hold = 1
        self._lock = threading.Lock()  # one lock per show: check-and-hold must be atomic

    def hold(self, user_id: str, seat_ids: List[str]) -> str:
        if not seat_ids:
            raise ValueError("no seats")
        if len(set(seat_ids)) != len(seat_ids):
            raise ValueError("duplicate seats")
        with self._lock:
            now = self._clock()
            for sid in seat_ids:  # 1. validate everything...
                seat = self._seats.get(sid)
                if seat is None:
                    raise ValueError("unknown seat " + sid)
                if not seat.is_available(now):
                    raise SeatUnavailableError(sid)
            hold_id = "H%d" % self._next_hold  # 2. ...then mutate everything
            self._next_hold += 1
            expires = now + self._ttl
            for sid in seat_ids:
                seat = self._seats[sid]
                seat.status, seat.hold_id, seat.expires_at = Status.HELD, hold_id, expires
            self._holds[hold_id] = _Hold(tuple(seat_ids), expires)
            return hold_id

    def confirm(self, hold_id: str) -> None:
        with self._lock:
            h = self._active_hold(hold_id)
            for sid in h.seats:
                self._seats[sid].status = Status.BOOKED
            h.active = False

    def release(self, hold_id: str) -> None:
        with self._lock:
            h = self._active_hold(hold_id)
            for sid in h.seats:
                seat = self._seats[sid]
                seat.status, seat.hold_id = Status.AVAILABLE, None
            h.active = False

    def status(self, seat_id: str) -> str:
        with self._lock:
            seat = self._seats.get(seat_id)
            if seat is None:
                raise ValueError("unknown seat " + seat_id)
            if seat.status is Status.HELD and seat.is_available(self._clock()):
                return Status.AVAILABLE.value
            return seat.status.value

    def available(self) -> List[str]:
        with self._lock:
            now = self._clock()
            return sorted(sid for sid, seat in self._seats.items() if seat.is_available(now))

    def _active_hold(self, hold_id: str) -> _Hold:
        h = self._holds.get(hold_id)
        if h is None or not h.active or h.expires_at <= self._clock():
            raise HoldNotActiveError(hold_id)
        return h
