from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Callable

# A clock is a zero-argument function returning the current time in milliseconds.
Clock = Callable[[], int]


class RateLimiter(ABC):
    @abstractmethod
    def allow(self, key: str) -> bool: ...


class TokenBucketLimiter(RateLimiter):
    def __init__(self, capacity: int, refill_per_second: float, clock: Clock) -> None:
        pass  # TODO

    def allow(self, key: str) -> bool:
        raise NotImplementedError("TODO")


class FixedWindowLimiter(RateLimiter):
    def __init__(self, max_requests: int, window_ms: int, clock: Clock) -> None:
        pass  # TODO

    def allow(self, key: str) -> bool:
        raise NotImplementedError("TODO")


class SlidingWindowLimiter(RateLimiter):
    def __init__(self, max_requests: int, window_ms: int, clock: Clock) -> None:
        pass  # TODO

    def allow(self, key: str) -> bool:
        raise NotImplementedError("TODO")
