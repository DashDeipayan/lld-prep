from __future__ import annotations

import threading
from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from typing import Callable, Dict, FrozenSet, List, Optional

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


# ======================= Strategies =======================
class NearestDriverStrategy(MatchingStrategy):
    def __init__(self, max_distance: int) -> None:
        self._max = max_distance

    def match(self, pickup, candidates):
        in_range = [d for d in candidates if d.location.distance(pickup) <= self._max]
        if not in_range:
            return None
        # nearest first, then higher rating (negated), then lower id
        best = min(in_range, key=lambda d: (d.location.distance(pickup), -d.rating, d.id))
        return best.id


class StandardPricing(PricingStrategy):
    def __init__(self, rates: Dict[VehicleClass, Rate], surge_percent: int) -> None:
        self._rates = dict(rates)
        self._surge = surge_percent

    def fare(self, cls, distance, duration_ms):
        r = self._rates[cls]
        minutes = -(-duration_ms // 60_000)  # ceiling division
        return (r.base + r.per_km * distance + r.per_min * minutes) * self._surge // 100


# ======================= Model =======================
class _Driver:
    def __init__(self, driver_id: str, cls: VehicleClass, location: Location, rating: float) -> None:
        self.id, self.cls, self.location, self.rating = driver_id, cls, location, rating
        self.online = False
        self.busy = False

    def info(self) -> DriverInfo:
        return DriverInfo(self.id, self.cls, self.location, self.rating)


class _Trip:
    def __init__(self, trip_id: str, rider: str, driver: _Driver, src: Location, dst: Location, cls: VehicleClass) -> None:
        self.id, self.rider, self.driver, self.src, self.dst, self.cls = trip_id, rider, driver, src, dst, cls
        self.status = TripStatus.ASSIGNED
        self.started_at = 0


ALLOWED: Dict[TripStatus, FrozenSet[TripStatus]] = {
    TripStatus.ASSIGNED: frozenset({TripStatus.ARRIVED, TripStatus.CANCELLED}),
    TripStatus.ARRIVED: frozenset({TripStatus.IN_PROGRESS, TripStatus.CANCELLED}),
    TripStatus.IN_PROGRESS: frozenset({TripStatus.COMPLETED}),
    TripStatus.COMPLETED: frozenset(),
    TripStatus.CANCELLED: frozenset(),
}


# ======================= Service =======================
class RideService:
    def __init__(self, matching: MatchingStrategy, pricing: PricingStrategy, clock: Clock) -> None:
        self._matching = matching
        self._pricing = pricing
        self._clock = clock
        self._drivers: Dict[str, _Driver] = {}
        self._trips: Dict[str, _Trip] = {}
        self._active_trip_by_rider: Dict[str, str] = {}
        self._next_trip = 1
        self._lock = threading.Lock()

    def add_driver(self, driver_id: str, cls: VehicleClass, location: Location, rating: float) -> None:
        with self._lock:
            if driver_id in self._drivers:
                raise ValueError("duplicate driver " + driver_id)
            self._drivers[driver_id] = _Driver(driver_id, cls, location, rating)

    def set_online(self, driver_id: str, online: bool) -> None:
        with self._lock:
            d = self._driver(driver_id)
            if not online and d.busy:
                raise RideStateError("driver is on a trip")
            d.online = online

    def update_location(self, driver_id: str, location: Location) -> None:
        with self._lock:
            self._driver(driver_id).location = location

    def request_ride(self, rider_id: str, src: Location, dst: Location, cls: VehicleClass) -> str:
        # Matching and marking the driver busy happen under one lock: no double assignment.
        with self._lock:
            if rider_id in self._active_trip_by_rider:
                raise RideStateError("rider already has an active trip")
            candidates = [d.info() for d in sorted(self._drivers.values(), key=lambda d: d.id)
                          if d.online and not d.busy and d.cls is cls]
            chosen = self._matching.match(src, candidates)
            if chosen is None:
                raise NoDriverAvailableError("no %s nearby" % cls.value)
            d = self._driver(chosen)
            d.busy = True
            t = _Trip("R%d" % self._next_trip, rider_id, d, src, dst, cls)
            self._next_trip += 1
            self._trips[t.id] = t
            self._active_trip_by_rider[rider_id] = t.id
            return t.id

    def driver_of(self, trip_id: str) -> str:
        with self._lock:
            return self._trip(trip_id).driver.id

    def status(self, trip_id: str) -> TripStatus:
        with self._lock:
            return self._trip(trip_id).status

    def arrived(self, trip_id: str) -> None:
        with self._lock:
            self._transition(self._trip(trip_id), TripStatus.ARRIVED)

    def start(self, trip_id: str) -> None:
        with self._lock:
            t = self._trip(trip_id)
            self._transition(t, TripStatus.IN_PROGRESS)
            t.started_at = self._clock()

    def end(self, trip_id: str) -> int:
        with self._lock:
            t = self._trip(trip_id)
            self._transition(t, TripStatus.COMPLETED)
            t.driver.location = t.dst
            return self._pricing.fare(t.cls, t.src.distance(t.dst), self._clock() - t.started_at)

    def cancel(self, trip_id: str) -> None:
        with self._lock:
            self._transition(self._trip(trip_id), TripStatus.CANCELLED)

    def _transition(self, t: _Trip, to: TripStatus) -> None:
        if to not in ALLOWED[t.status]:
            raise RideStateError("%s -> %s not allowed" % (t.status.value, to.value))
        t.status = to
        if to in (TripStatus.COMPLETED, TripStatus.CANCELLED):
            t.driver.busy = False
            self._active_trip_by_rider.pop(t.rider, None)

    def _driver(self, driver_id: str) -> _Driver:
        d = self._drivers.get(driver_id)
        if d is None:
            raise ValueError("unknown driver " + driver_id)
        return d

    def _trip(self, trip_id: str) -> _Trip:
        t = self._trips.get(trip_id)
        if t is None:
            raise ValueError("unknown trip " + trip_id)
        return t
