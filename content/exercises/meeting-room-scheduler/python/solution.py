from __future__ import annotations

from dataclasses import dataclass
from typing import List


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


# ======================= Your design =======================
class Scheduler:
    def add_room(self, room_id: str, capacity: int) -> None:
        raise NotImplementedError("TODO")

    def book(self, organizer: str, start: int, end: int, attendees: int) -> Booking:
        raise NotImplementedError("TODO")

    def book_room(self, room_id: str, organizer: str, start: int, end: int) -> Booking:
        raise NotImplementedError("TODO")

    def cancel(self, booking_id: str) -> None:
        raise NotImplementedError("TODO")

    def free_rooms(self, start: int, end: int, attendees: int) -> List[str]:
        raise NotImplementedError("TODO")

    def schedule(self, room_id: str) -> List[Booking]:
        raise NotImplementedError("TODO")

    def free_slots(self, room_id: str, start: int, end: int) -> List[Slot]:
        raise NotImplementedError("TODO")
