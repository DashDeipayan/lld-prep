from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum, IntEnum
from typing import Callable, Dict, List, Optional


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
        pass  # TODO

    @property
    def id(self) -> str:
        raise NotImplementedError("TODO")

    @property
    def floor(self) -> int:
        raise NotImplementedError("TODO")

    @property
    def size(self) -> SpotSize:
        raise NotImplementedError("TODO")

    def can_fit(self, vehicle: Vehicle) -> bool:
        raise NotImplementedError("TODO")

    def is_free(self) -> bool:
        raise NotImplementedError("TODO")

    def occupy(self, vehicle: Vehicle) -> None:
        raise NotImplementedError("TODO")

    def vacate(self) -> Optional[Vehicle]:
        raise NotImplementedError("TODO")


# ============================ Step 3 ============================
class NearestSpotSelector(SpotSelector):
    def select(self, candidates, vehicle):
        raise NotImplementedError("TODO")


# ============================ Step 4 ============================
class HourlyPricing(PricingStrategy):
    def __init__(self, rate_per_hour: Dict[VehicleType, int]) -> None:
        pass  # TODO

    def fee(self, vehicle_type, duration_ms):
        raise NotImplementedError("TODO")


# ======================= Steps 2, 3, 4, 5, 6 =======================
class ParkingLot:
    def __init__(self, pricing: PricingStrategy, selector: SpotSelector, clock: Clock) -> None:
        pass  # TODO

    def add_floor(self, floor: int, small: int, medium: int, large: int) -> None:
        raise NotImplementedError("TODO")

    def available(self, size: SpotSize) -> int:
        raise NotImplementedError("TODO")

    def park(self, vehicle: Vehicle) -> Ticket:
        raise NotImplementedError("TODO")

    def unpark(self, ticket_id: str) -> Receipt:
        raise NotImplementedError("TODO")

    def add_listener(self, listener: ParkingListener) -> None:
        raise NotImplementedError("TODO")


# ============================ Step 5 ============================
class DisplayBoard(ParkingListener):
    def __init__(self, lot: ParkingLot) -> None:
        pass  # TODO

    def on_event(self, event: Event) -> None:
        raise NotImplementedError("TODO")

    def free(self, size: SpotSize) -> int:
        raise NotImplementedError("TODO")
