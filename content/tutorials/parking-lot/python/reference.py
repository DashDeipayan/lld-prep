from __future__ import annotations

import threading
from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum, IntEnum
from typing import Callable, Dict, List, Optional, Set


# ======================= Provided (do not change) =======================
class SpotSize(IntEnum):  # IntEnum, so sizes compare: SMALL < MEDIUM < LARGE
    SMALL = 1
    MEDIUM = 2
    LARGE = 3


class VehicleType(Enum):
    BIKE = SpotSize.SMALL
    CAR = SpotSize.MEDIUM
    TRUCK = SpotSize.LARGE

    @property
    def min_size(self) -> SpotSize:
        return self.value


@dataclass(frozen=True)
class Vehicle:
    plate: str
    type: VehicleType


@dataclass(frozen=True)
class Ticket:
    id: str
    plate: str
    spot_id: str
    entry_ms: int


@dataclass(frozen=True)
class Receipt:
    ticket_id: str
    plate: str
    spot_id: str
    minutes: int
    fee: int


@dataclass(frozen=True)
class Event:
    kind: str  # "PARKED" or "LEFT"
    ticket_id: str
    spot_id: str
    size: SpotSize


Clock = Callable[[], int]  # milliseconds


class PricingStrategy(ABC):
    @abstractmethod
    def fee(self, vehicle_type: VehicleType, duration_ms: int) -> int: ...


class SpotSelector(ABC):
    @abstractmethod
    def select(self, candidates: List["Spot"], vehicle: Vehicle) -> Optional["Spot"]: ...


class ParkingListener(ABC):
    @abstractmethod
    def on_event(self, event: Event) -> None: ...


class ParkingFullError(Exception):
    pass


class ParkingStateError(Exception):
    """The spot is taken or too small, or the plate is already parked."""


# ============================ Step 1 ============================
class Spot:
    def __init__(self, spot_id: str, floor: int, size: SpotSize) -> None:
        self._id = spot_id
        self._floor = floor
        self._size = size
        self._vehicle: Optional[Vehicle] = None

    @property
    def id(self) -> str:
        return self._id

    @property
    def floor(self) -> int:
        return self._floor

    @property
    def size(self) -> SpotSize:
        return self._size

    def can_fit(self, vehicle: Vehicle) -> bool:
        return self._size >= vehicle.type.min_size  # the fitting rule is data, not a switch

    def is_free(self) -> bool:
        return self._vehicle is None

    def occupy(self, vehicle: Vehicle) -> None:
        if not self.is_free():
            raise ParkingStateError(self._id + " is occupied")
        if not self.can_fit(vehicle):
            raise ParkingStateError("%s does not fit %s" % (vehicle.type.name, self._size.name))
        self._vehicle = vehicle

    def vacate(self) -> Optional[Vehicle]:
        v, self._vehicle = self._vehicle, None
        return v


# ============================ Step 3 ============================
class NearestSpotSelector(SpotSelector):
    def select(self, candidates, vehicle):
        return candidates[0] if candidates else None  # candidates arrive in lot order


# ============================ Step 4 ============================
class HourlyPricing(PricingStrategy):
    HOUR_MS = 3_600_000

    def __init__(self, rate_per_hour: Dict[VehicleType, int]) -> None:
        self._rate = dict(rate_per_hour)

    def fee(self, vehicle_type, duration_ms):
        hours = max(1, -(-duration_ms // self.HOUR_MS))  # started hours (ceiling), at least 1
        return hours * self._rate.get(vehicle_type, 0)


# ======================= Steps 2, 3, 4, 5, 6 =======================
class ParkingLot:
    def __init__(self, pricing: PricingStrategy, selector: SpotSelector, clock: Clock) -> None:
        self._pricing = pricing
        self._selector = selector
        self._clock = clock
        self._spots: List[Spot] = []  # lot order: floor, then S -> M -> L, then number
        self._spots_by_id: Dict[str, Spot] = {}
        self._floors: Set[int] = set()
        self._active: Dict[str, Ticket] = {}
        self._parked_plates: Set[str] = set()
        self._listeners: List[ParkingListener] = []
        self._next_ticket = 1
        self._lock = threading.Lock()

    def add_floor(self, floor: int, small: int, medium: int, large: int) -> None:
        with self._lock:
            if floor in self._floors:
                raise ValueError("duplicate floor %d" % floor)
            self._floors.add(floor)
            for size, code, count in ((SpotSize.SMALL, "S", small), (SpotSize.MEDIUM, "M", medium), (SpotSize.LARGE, "L", large)):
                for n in range(1, count + 1):
                    spot = Spot("F%d-%s%d" % (floor, code, n), floor, size)
                    self._spots.append(spot)
                    self._spots_by_id[spot.id] = spot

    def available(self, size: SpotSize) -> int:
        with self._lock:
            return sum(1 for s in self._spots if s.size == size and s.is_free())

    def park(self, vehicle: Vehicle) -> Ticket:
        with self._lock:  # one lock: choosing and occupying a spot is one atomic step
            if vehicle.plate in self._parked_plates:
                raise ParkingStateError(vehicle.plate + " already parked")
            candidates = [s for s in self._spots if s.is_free() and s.can_fit(vehicle)]
            spot = self._selector.select(candidates, vehicle)  # the strategy decides
            if spot is None:
                raise ParkingFullError("no spot for " + vehicle.type.name)
            spot.occupy(vehicle)
            t = Ticket("T%d" % self._next_ticket, vehicle.plate, spot.id, self._clock())
            self._next_ticket += 1
            self._active[t.id] = t
            self._parked_plates.add(vehicle.plate)
            listeners = list(self._listeners)
        self._emit(listeners, Event("PARKED", t.id, spot.id, spot.size))  # outside the lock
        return t

    def unpark(self, ticket_id: str) -> Receipt:
        with self._lock:
            t = self._active.pop(ticket_id, None)
            if t is None:
                raise ValueError("unknown or used ticket " + ticket_id)
            spot = self._spots_by_id[t.spot_id]
            v = spot.vacate()
            self._parked_plates.discard(v.plate)
            duration = self._clock() - t.entry_ms
            r = Receipt(t.id, t.plate, spot.id, duration // 60_000, self._pricing.fee(v.type, duration))
            listeners = list(self._listeners)
        self._emit(listeners, Event("LEFT", ticket_id, spot.id, spot.size))
        return r

    def add_listener(self, listener: ParkingListener) -> None:
        with self._lock:
            self._listeners.append(listener)

    @staticmethod
    def _emit(listeners: List[ParkingListener], event: Event) -> None:
        for listener in listeners:  # in the order they were added
            listener.on_event(event)


# ============================ Step 5 ============================
class DisplayBoard(ParkingListener):
    def __init__(self, lot: ParkingLot) -> None:
        self._free = {s: lot.available(s) for s in SpotSize}
        self._lock = threading.Lock()  # events arrive from several gates at once
        lot.add_listener(self)

    def on_event(self, event: Event) -> None:
        with self._lock:
            self._free[event.size] += -1 if event.kind == "PARKED" else 1

    def free(self, size: SpotSize) -> int:
        with self._lock:
            return self._free[size]
