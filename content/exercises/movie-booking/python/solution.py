from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Callable, List

# ======================= Provided =======================
Clock = Callable[[], int]  # milliseconds


class PaymentGateway(ABC):
    @abstractmethod
    def charge(self, user: str, amount: int) -> bool: ...


@dataclass(frozen=True)
class Booking:
    id: str
    show_id: str
    user: str
    seats: List[str]  # sorted
    amount: int
    status: str  # "CONFIRMED" or "CANCELLED"


class SeatUnavailableError(Exception):
    pass


class PaymentFailedError(Exception):
    pass


class HoldNotActiveError(Exception):
    pass


class BookingStateError(Exception):
    """Show overlap, show already started, or a double cancel."""


# ======================= Your design =======================
class BookingService:
    def __init__(self, gateway: PaymentGateway, clock: Clock, hold_ttl_ms: int) -> None:
        pass  # TODO

    def add_theatre(self, theatre_id: str, city: str) -> None:
        raise NotImplementedError("TODO")

    def add_screen(self, theatre_id: str, screen_id: str, regular_rows: int, premium_rows: int, seats_per_row: int) -> None:
        raise NotImplementedError("TODO")

    def add_movie(self, movie_id: str, title: str, duration_minutes: int) -> None:
        raise NotImplementedError("TODO")

    def add_show(self, movie_id: str, screen_id: str, start_ms: int, regular_price: int, premium_price: int) -> str:
        raise NotImplementedError("TODO")

    def find_shows(self, city: str, movie_id: str) -> List[str]:
        raise NotImplementedError("TODO")

    def available_seats(self, show_id: str) -> List[str]:
        raise NotImplementedError("TODO")

    def hold(self, show_id: str, user: str, seat_ids: List[str]) -> str:
        raise NotImplementedError("TODO")

    def confirm(self, hold_id: str) -> Booking:
        raise NotImplementedError("TODO")

    def cancel(self, booking_id: str) -> Booking:
        raise NotImplementedError("TODO")
