from __future__ import annotations

import dataclasses
import threading
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List


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


# ---- Model ----
@dataclass(frozen=True)
class _Branch:
    id: str
    city: str


@dataclass
class _Car:
    id: str
    branch: _Branch
    type: CarType
    rate: int
    active: List[Reservation] = field(default_factory=list)

    def free(self, start: int, end: int) -> bool:
        return all(not (start < r.end_day and r.start_day < end) for r in self.active)  # half-open overlap


def _validate(start: int, end: int) -> None:
    if start < 0 or start >= end:
        raise ValueError("invalid range")


class RentalService:
    def __init__(self) -> None:
        self._branches: Dict[str, _Branch] = {}
        self._cars: Dict[str, _Car] = {}
        self._reservations: Dict[str, Reservation] = {}
        self._next = 1
        self._lock = threading.Lock()  # check-then-reserve must be atomic

    def add_branch(self, branch_id: str, city: str) -> None:
        with self._lock:
            if branch_id in self._branches:
                raise ValueError("duplicate branch")
            self._branches[branch_id] = _Branch(branch_id, city)

    def add_car(self, branch_id: str, car_id: str, car_type: CarType, daily_rate: int) -> None:
        with self._lock:
            b = self._branches.get(branch_id)
            if b is None or car_id in self._cars or daily_rate <= 0:
                raise ValueError("invalid car")
            self._cars[car_id] = _Car(car_id, b, car_type, daily_rate)

    def search(self, city: str, car_type: CarType, start_day: int, end_day: int) -> List[str]:
        _validate(start_day, end_day)
        with self._lock:
            hits = [c for c in self._cars.values()
                    if c.type is car_type and c.branch.city.lower() == city.lower() and c.free(start_day, end_day)]
            return [c.id for c in sorted(hits, key=lambda c: (c.rate, c.id))]

    def reserve(self, customer: str, car_id: str, start_day: int, end_day: int) -> Reservation:
        _validate(start_day, end_day)
        with self._lock:
            c = self._car(car_id)
            if not c.free(start_day, end_day):
                raise ReservationError("car %s is booked for those dates" % car_id)
            r = Reservation("RS%d" % self._next, customer, car_id, start_day, end_day, (end_day - start_day) * c.rate, "RESERVED")
            self._next += 1
            c.active.append(r)
            self._reservations[r.id] = r
            return r

    def cancel(self, reservation_id: str) -> Reservation:
        with self._lock:
            r = self._active(reservation_id)
            return self._finish(r, "CANCELLED", r.cost)

    def return_car(self, reservation_id: str, actual_day: int) -> int:
        with self._lock:
            r = self._active(reservation_id)
            c = self._car(r.car_id)
            late = max(0, actual_day - r.end_day)
            amount = r.cost + late * c.rate * 3 // 2
            self._finish(r, "COMPLETED", amount)
            return amount

    def _finish(self, r: Reservation, status: str, amount: int) -> Reservation:
        self._car(r.car_id).active.remove(r)
        done = dataclasses.replace(r, status=status, cost=amount)  # a new snapshot; the old one is untouched
        self._reservations[r.id] = done
        return done

    def _active(self, reservation_id: str) -> Reservation:
        r = self._reservations.get(reservation_id)
        if r is None:
            raise ValueError("unknown reservation " + reservation_id)
        if r.status != "RESERVED":
            raise ReservationError("reservation is " + r.status)
        return r

    def _car(self, car_id: str) -> _Car:
        c = self._cars.get(car_id)
        if c is None:
            raise ValueError("unknown car " + car_id)
        return c
