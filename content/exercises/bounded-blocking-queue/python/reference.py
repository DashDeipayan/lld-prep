from __future__ import annotations

import threading
import time
from collections import deque
from typing import Deque, Generic, Optional, TypeVar

T = TypeVar("T")


class QueueClosedError(Exception):
    """Provided: put/offer on a closed queue."""


class BoundedBlockingQueue(Generic[T]):
    """None is not a valid item: take()/poll() use None to mean "closed and drained" / "timed out"."""

    def __init__(self, capacity: int) -> None:
        if capacity < 1:
            raise ValueError("capacity must be >= 1")
        self._items: Deque[T] = deque()
        self._capacity = capacity
        self._lock = threading.Lock()
        # Two conditions on one lock: producers wait on not_full, consumers on not_empty.
        self._not_full = threading.Condition(self._lock)
        self._not_empty = threading.Condition(self._lock)
        self._closed = False

    def put(self, item: T) -> None:
        self._check_item(item)
        with self._lock:
            while len(self._items) == self._capacity and not self._closed:  # `while`, not `if`
                self._not_full.wait()
            self._enqueue(item)

    def offer(self, item: T, timeout: float) -> bool:
        self._check_item(item)
        deadline = time.monotonic() + timeout
        with self._lock:
            while len(self._items) == self._capacity and not self._closed:
                remaining = deadline - time.monotonic()
                if remaining <= 0:
                    return False
                self._not_full.wait(remaining)
            self._enqueue(item)
            return True

    def take(self) -> Optional[T]:
        with self._lock:
            while not self._items and not self._closed:
                self._not_empty.wait()
            return self._dequeue() if self._items else None

    def poll(self, timeout: float) -> Optional[T]:
        deadline = time.monotonic() + timeout
        with self._lock:
            while not self._items and not self._closed:
                remaining = deadline - time.monotonic()
                if remaining <= 0:
                    return None
                self._not_empty.wait(remaining)
            return self._dequeue() if self._items else None

    def close(self) -> None:
        with self._lock:
            self._closed = True
            self._not_full.notify_all()  # blocked producers wake up and fail
            self._not_empty.notify_all()  # blocked consumers wake up and drain, then get None

    def size(self) -> int:
        with self._lock:
            return len(self._items)

    # ---- called with the lock held ----
    def _enqueue(self, item: T) -> None:
        if self._closed:
            raise QueueClosedError("queue closed")
        self._items.append(item)
        self._not_empty.notify()

    def _dequeue(self) -> T:
        item = self._items.popleft()
        self._not_full.notify()
        return item

    @staticmethod
    def _check_item(item: T) -> None:
        if item is None:
            raise ValueError("None can't be queued")
