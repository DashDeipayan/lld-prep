from __future__ import annotations

import threading
from typing import Generic, Optional, TypeVar

T = TypeVar("T")


class QueueClosedError(Exception):
    """Provided: put/offer on a closed queue."""


class BoundedBlockingQueue(Generic[T]):
    """None is not a valid item: take()/poll() use None to mean "closed and drained" / "timed out"."""

    def __init__(self, capacity: int) -> None:
        pass  # TODO

    def put(self, item: T) -> None:
        raise NotImplementedError("TODO")

    def take(self) -> Optional[T]:
        raise NotImplementedError("TODO")

    def offer(self, item: T, timeout: float) -> bool:
        """timeout is in seconds."""
        raise NotImplementedError("TODO")

    def poll(self, timeout: float) -> Optional[T]:
        """timeout is in seconds."""
        raise NotImplementedError("TODO")

    def close(self) -> None:
        raise NotImplementedError("TODO")

    def size(self) -> int:
        raise NotImplementedError("TODO")
