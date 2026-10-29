from __future__ import annotations

import threading
from abc import ABC, abstractmethod
from collections import defaultdict, deque
from typing import Callable, Deque, Dict, List, Optional

# A clock is a zero-argument function returning the current time in milliseconds.
Clock = Callable[[], int]


class Sender(ABC):
    @abstractmethod
    def send(self, to: str, msg: str) -> None:
        """Raises an exception on failure."""


class RateLimitedError(Exception):
    """Provided: raised when a recipient has used up their quota."""


class RetryingSender(Sender):
    def __init__(self, inner: Sender, max_attempts: int) -> None:
        if max_attempts < 1:
            raise ValueError("max_attempts must be >= 1")
        self._inner = inner
        self._max_attempts = max_attempts

    def send(self, to: str, msg: str) -> None:
        last: Optional[Exception] = None
        for _ in range(self._max_attempts):
            try:
                self._inner.send(to, msg)
                return
            except Exception as e:  # noqa: B902
                last = e
        raise last


class RateLimitedSender(Sender):
    def __init__(self, inner: Sender, clock: Clock, max_per_window: int, window_ms: int) -> None:
        self._inner = inner
        self._clock = clock
        self._max = max_per_window
        self._window = window_ms
        self._sent: Dict[str, Deque[int]] = defaultdict(deque)  # recipient -> times of successful sends
        self._lock = threading.Lock()

    def send(self, to: str, msg: str) -> None:
        with self._lock:
            now = self._clock()
            times = self._sent[to]
            while times and now - times[0] >= self._window:
                times.popleft()  # slid out of the window
            if len(times) >= self._max:
                raise RateLimitedError("rate limited: " + to)
            self._inner.send(to, msg)  # raises => quota not consumed
            times.append(now)


class LoggingSender(Sender):
    def __init__(self, inner: Sender, log: List[str]) -> None:
        self._inner = inner
        self._log = log

    def send(self, to: str, msg: str) -> None:
        try:
            self._inner.send(to, msg)
        except Exception:
            self._log.append("FAIL %s: %s" % (to, msg))
            raise  # pass the error through unchanged
        self._log.append("OK %s: %s" % (to, msg))
