from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from typing import List


# ---- Provided ----
class CarType(Enum):
    HATCHBACK = "HATCHBACK"
    SEDAN = "SEDAN"
    SUV = "SUV"


@dataclass(frozen=True)
class Reservation:
    id: str
    customer: str
    car_id: str
    start_day: int
    end_day: int
    cost: int
    status: str  # "RESERVED", "CANCELLED" or "COMPLETED"


class ReservationError(Exception):
    """The car is booked for those dates, or the reservation isn't RESERVED any more."""


# ---- Your design ----
class RentalService:
    def add_branch(self, branch_id: str, city: str) -> None:
        raise NotImplementedError("TODO")

    def add_car(self, branch_id: str, car_id: str, car_type: CarType, daily_rate: int) -> None:
        raise NotImplementedError("TODO")

    def search(self, city: str, car_type: CarType, start_day: int, end_day: int) -> List[str]:
        raise NotImplementedError("TODO")

    def reserve(self, customer: str, car_id: str, start_day: int, end_day: int) -> Reservation:
        raise NotImplementedError("TODO")

    def cancel(self, reservation_id: str) -> Reservation:
        raise NotImplementedError("TODO")

    def return_car(self, reservation_id: str, actual_day: int) -> int:
        raise NotImplementedError("TODO")
