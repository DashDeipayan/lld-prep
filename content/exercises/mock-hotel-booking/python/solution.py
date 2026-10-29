from __future__ import annotations

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


# ---- Your design ----
class Hotel:
    def add_room(self, number: str, room_type: RoomType, nightly_rate: int) -> None:
        raise NotImplementedError("TODO")

    def availability(self, check_in: int, check_out: int) -> Dict[RoomType, int]:
        raise NotImplementedError("TODO")

    def book(self, guest: str, room_type: RoomType, check_in: int, check_out: int) -> str:
        raise NotImplementedError("TODO")

    def room_of(self, booking_id: str) -> str:
        raise NotImplementedError("TODO")

    def check_in(self, booking_id: str, day: int) -> None:
        raise NotImplementedError("TODO")

    def add_charge(self, booking_id: str, amount: int) -> None:
        raise NotImplementedError("TODO")

    def check_out(self, booking_id: str) -> int:
        raise NotImplementedError("TODO")

    def cancel(self, booking_id: str) -> None:
        raise NotImplementedError("TODO")

    def status(self, booking_id: str) -> str:
        raise NotImplementedError("TODO")
