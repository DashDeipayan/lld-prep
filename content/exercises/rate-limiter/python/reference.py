from __future__ import annotations

import threading
from abc import ABC, abstractmethod
from collections import deque
from typing import Callable, Deque, Dict, Generic, TypeVar

# A clock is a zero-argument function returning the current time in milliseconds.
Clock = Callable[[], int]

S = TypeVar("S")


class RateLimiter(ABC):
    @abstractmethod
    def allow(self, key: str) -> bool: ...


class _PerKey(Generic[S]):
    """Per-key state, each with its own lock, so different keys never contend."""

    def __init__(self, new_state: Callable[[], S]) -> None:
        self._new_state = new_state
        self._states: Dict[str, S] = {}
        self._locks: Dict[str, threading.Lock] = {}
        self._guard = threading.Lock()

    def get(self, key: str):
        with self._guard:  # only held to look up / create the entry
            if key not in self._states:
                self._states[key] = self._new_state()
                self._locks[key] = threading.Lock()
            return self._states[key], self._locks[key]


class _Bucket:
    __slots__ = ("tokens", "last")

    def __init__(self, tokens: float, last: int) -> None:
        self.tokens = tokens
        self.last = last


class TokenBucketLimiter(RateLimiter):
    EPSILON = 1e-9  # 0.3 + 0.3 + 0.3 + 0.1 < 1.0 in floating point

    def __init__(self, capacity: int, refill_per_second: float, clock: Clock) -> None:
        self._capacity = capacity
        self._per_ms = refill_per_second / 1000.0
        self._clock = clock
        self._buckets = _PerKey(lambda: _Bucket(capacity, clock()))  # each key starts full

    def allow(self, key: str) -> bool:
        bucket, lock = self._buckets.get(key)
        with lock:
            now = self._clock()
            bucket.tokens = min(self._capacity, bucket.tokens + (now - bucket.last) * self._per_ms)
            bucket.last = max(bucket.last, now)
            if bucket.tokens >= 1 - self.EPSILON:
                bucket.tokens = max(0.0, bucket.tokens - 1)
                return True
            return False


class _Window:
    __slots__ = ("start", "count")

    def __init__(self) -> None:
        self.start = -1
        self.count = 0


class FixedWindowLimiter(RateLimiter):
    def __init__(self, max_requests: int, window_ms: int, clock: Clock) -> None:
        self._max = max_requests
        self._window = window_ms
        self._clock = clock
        self._windows = _PerKey(_Window)

    def allow(self, key: str) -> bool:
        w, lock = self._windows.get(key)
        with lock:
            now = self._clock()
            start = now - now % self._window  # aligned: [0, w), [w, 2w), ...
            if w.start != start:
                w.start, w.count = start, 0
            if w.count >= self._max:
                return False
            w.count += 1
            return True


class SlidingWindowLimiter(RateLimiter):
    def __init__(self, max_requests: int, window_ms: int, clock: Clock) -> None:
        self._max = max_requests
        self._window = window_ms
        self._clock = clock
        self._logs: _PerKey[Deque[int]] = _PerKey(deque)

    def allow(self, key: str) -> bool:
        log, lock = self._logs.get(key)
        with lock:
            now = self._clock()
            while log and now - log[0] >= self._window:
                log.popleft()
            if len(log) >= self._max:
                return False  # rejected requests aren't recorded
            log.append(now)
            return True
