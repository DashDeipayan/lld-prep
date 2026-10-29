from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from typing import Callable, Dict, List, Optional

# ======================= Provided =======================
Clock = Callable[[], int]  # milliseconds


class VehicleClass(Enum):
    MINI = "MINI"
    SEDAN = "SEDAN"
    SUV = "SUV"


class TripStatus(Enum):
    ASSIGNED = "ASSIGNED"
    ARRIVED = "ARRIVED"
    IN_PROGRESS = "IN_PROGRESS"
    COMPLETED = "COMPLETED"
    CANCELLED = "CANCELLED"


@dataclass(frozen=True)
class Location:
    x: int
    y: int

    def distance(self, other: "Location") -> int:
        return abs(self.x - other.x) + abs(self.y - other.y)


@dataclass(frozen=True)
class DriverInfo:
    id: str
    cls: VehicleClass
    location: Location
    rating: float


@dataclass(frozen=True)
class Rate:
    base: int
    per_km: int
    per_min: int


class MatchingStrategy(ABC):
    @abstractmethod
    def match(self, pickup: Location, candidates: List[DriverInfo]) -> Optional[str]: ...


class PricingStrategy(ABC):
    @abstractmethod
    def fare(self, cls: VehicleClass, distance: int, duration_ms: int) -> int: ...


class NoDriverAvailableError(Exception):
    pass


class RideStateError(Exception):
    """An illegal trip transition, a rider who already has a trip, or a driver going offline mid-trip."""


# ======================= Your design =======================
class NearestDriverStrategy(MatchingStrategy):
    def __init__(self, max_distance: int) -> None:
        pass  # TODO

    def match(self, pickup, candidates):
        raise NotImplementedError("TODO")


class StandardPricing(PricingStrategy):
    def __init__(self, rates: Dict[VehicleClass, Rate], surge_percent: int) -> None:
        pass  # TODO

    def fare(self, cls, distance, duration_ms):
        raise NotImplementedError("TODO")


class RideService:
    def __init__(self, matching: MatchingStrategy, pricing: PricingStrategy, clock: Clock) -> None:
        pass  # TODO

    def add_driver(self, driver_id: str, cls: VehicleClass, location: Location, rating: float) -> None:
        raise NotImplementedError("TODO")

    def set_online(self, driver_id: str, online: bool) -> None:
        raise NotImplementedError("TODO")

    def update_location(self, driver_id: str, location: Location) -> None:
        raise NotImplementedError("TODO")

    def request_ride(self, rider_id: str, src: Location, dst: Location, cls: VehicleClass) -> str:
        raise NotImplementedError("TODO")

    def driver_of(self, trip_id: str) -> str:
        raise NotImplementedError("TODO")

    def status(self, trip_id: str) -> TripStatus:
        raise NotImplementedError("TODO")

    def arrived(self, trip_id: str) -> None:
        raise NotImplementedError("TODO")

    def start(self, trip_id: str) -> None:
        raise NotImplementedError("TODO")

    def end(self, trip_id: str) -> int:
        raise NotImplementedError("TODO")

    def cancel(self, trip_id: str) -> None:
        raise NotImplementedError("TODO")
