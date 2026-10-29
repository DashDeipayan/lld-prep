from __future__ import annotations

import threading
from abc import ABC, abstractmethod
from enum import Enum
from typing import List, Optional, Set


class Direction(Enum):
    UP = "UP"
    DOWN = "DOWN"
    IDLE = "IDLE"


class DispatchStrategy(ABC):
    @abstractmethod
    def choose(self, cars: List["Elevator"], floor: int, direction: Direction) -> "Elevator": ...


class Elevator:
    def __init__(self, car_id: int) -> None:
        self._id = car_id
        self._floor = 0
        self._direction = Direction.IDLE
        self._stops: Set[int] = set()

    @property
    def id(self) -> int:
        return self._id

    @property
    def floor(self) -> int:
        return self._floor

    @property
    def direction(self) -> Direction:
        return self._direction

    def stops(self) -> List[int]:
        return sorted(self._stops)

    def add_stop(self, floor: int) -> None:
        self._stops.add(floor)

    def _above(self) -> bool:
        return any(s > self._floor for s in self._stops)

    def _below(self) -> bool:
        return any(s < self._floor for s in self._stops)

    def step(self) -> Optional[int]:
        if not self._stops:  # 1
            self._direction = Direction.IDLE
            return None
        if self._floor in self._stops:  # 2: open doors, don't move
            return self._arrive()
        if self._direction is Direction.IDLE:  # 3: nearest stop, UP on ties
            up = min((s for s in self._stops if s > self._floor), default=None)
            down = max((s for s in self._stops if s < self._floor), default=None)
            go_up = down is None or (up is not None and up - self._floor <= self._floor - down)
            self._direction = Direction.UP if go_up else Direction.DOWN
        if self._direction is Direction.UP and not self._above():  # 4: LOOK reverses at the last stop
            self._direction = Direction.DOWN
        elif self._direction is Direction.DOWN and not self._below():
            self._direction = Direction.UP
        self._floor += 1 if self._direction is Direction.UP else -1  # 5
        return self._arrive() if self._floor in self._stops else None

    def _arrive(self) -> int:
        self._stops.discard(self._floor)
        if not self._stops:
            self._direction = Direction.IDLE
        return self._floor


class NearestCarStrategy(DispatchStrategy):
    def choose(self, cars, floor, direction):
        # min() keeps the first of equal costs, and cars are in id order: ties -> lowest id.
        return min(cars, key=lambda c: self.cost(c, floor, direction))

    @staticmethod
    def cost(car: Elevator, floor: int, direction: Direction) -> int:
        cur = car.floor
        if car.direction is Direction.IDLE:
            return abs(cur - floor)
        if car.direction is Direction.UP and direction is Direction.UP and floor >= cur:
            return floor - cur
        if car.direction is Direction.DOWN and direction is Direction.DOWN and floor <= cur:
            return cur - floor
        stops = car.stops()
        turn = cur if not stops else (stops[-1] if car.direction is Direction.UP else stops[0])
        return abs(cur - turn) + abs(turn - floor)


class ElevatorSystem:
    def __init__(self, floors: int, cars: int, strategy: DispatchStrategy) -> None:
        if floors < 2 or cars < 1:
            raise ValueError("need >= 2 floors and >= 1 car")
        self._floors = floors
        self._cars = [Elevator(i) for i in range(cars)]
        self._strategy = strategy
        self._lock = threading.RLock()

    def request_pickup(self, floor: int, direction: Direction) -> int:
        with self._lock:
            self._check_floor(floor)
            if direction is Direction.IDLE:
                raise ValueError("pickup direction must be UP or DOWN")
            car = self._strategy.choose(list(self._cars), floor, direction)  # a copy: strategies can't reorder ours
            car.add_stop(floor)
            return car.id

    def request_destination(self, car_id: int, floor: int) -> None:
        with self._lock:
            self._check_floor(floor)
            self.car(car_id).add_stop(floor)

    def step(self) -> List[str]:
        with self._lock:
            events = []
            for car in self._cars:
                f = car.step()
                if f is not None:
                    events.append("E%d@%d" % (car.id, f))
            return events

    def car(self, car_id: int) -> Elevator:
        with self._lock:
            if not 0 <= car_id < len(self._cars):
                raise ValueError("no elevator %d" % car_id)
            return self._cars[car_id]

    def _check_floor(self, f: int) -> None:
        if not 0 <= f < self._floors:
            raise ValueError("invalid floor %d" % f)
