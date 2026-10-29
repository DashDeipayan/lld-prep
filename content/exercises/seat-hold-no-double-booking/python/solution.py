from __future__ import annotations

from typing import Callable, List

# A clock is a zero-argument function returning the current time in milliseconds.
Clock = Callable[[], int]


class SeatUnavailableError(Exception):
    """Provided: one of the requested seats is held or booked."""


class HoldNotActiveError(Exception):
    """Provided: the hold is unknown, expired, released or already confirmed."""


class SeatInventory:
    def __init__(self, seat_ids: List[str], hold_ttl_ms: int, clock: Clock) -> None:
        pass  # TODO

    def hold(self, user_id: str, seat_ids: List[str]) -> str:
        raise NotImplementedError("TODO")

    def confirm(self, hold_id: str) -> None:
        raise NotImplementedError("TODO")

    def release(self, hold_id: str) -> None:
        raise NotImplementedError("TODO")

    def status(self, seat_id: str) -> str:
        raise NotImplementedError("TODO")

    def available(self) -> List[str]:
        raise NotImplementedError("TODO")
