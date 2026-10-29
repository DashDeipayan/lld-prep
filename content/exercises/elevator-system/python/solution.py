from __future__ import annotations

from abc import ABC, abstractmethod
from enum import Enum
from typing import List, Optional


class Direction(Enum):
    UP = "UP"
    DOWN = "DOWN"
    IDLE = "IDLE"


class DispatchStrategy(ABC):
    @abstractmethod
    def choose(self, cars: List["Elevator"], floor: int, direction: Direction) -> "Elevator": ...


class Elevator:
    def __init__(self, car_id: int) -> None:
        pass  # TODO

    @property
    def id(self) -> int:
        raise NotImplementedError("TODO")

    @property
    def floor(self) -> int:
        raise NotImplementedError("TODO")

    @property
    def direction(self) -> Direction:
        raise NotImplementedError("TODO")

    def stops(self) -> List[int]:
        """Sorted ascending."""
        raise NotImplementedError("TODO")

    def add_stop(self, floor: int) -> None:
        raise NotImplementedError("TODO")

    def step(self) -> Optional[int]:
        """Returns the floor where doors opened this tick, if any."""
        raise NotImplementedError("TODO")


class NearestCarStrategy(DispatchStrategy):
    def choose(self, cars, floor, direction):
        raise NotImplementedError("TODO")


class ElevatorSystem:
    def __init__(self, floors: int, cars: int, strategy: DispatchStrategy) -> None:
        pass  # TODO

    def request_pickup(self, floor: int, direction: Direction) -> int:
        raise NotImplementedError("TODO")

    def request_destination(self, car_id: int, floor: int) -> None:
        raise NotImplementedError("TODO")

    def step(self) -> List[str]:
        raise NotImplementedError("TODO")

    def car(self, car_id: int) -> Elevator:
        raise NotImplementedError("TODO")
