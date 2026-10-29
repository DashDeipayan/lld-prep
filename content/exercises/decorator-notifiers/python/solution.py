from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Callable, List

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
        pass  # TODO

    def send(self, to: str, msg: str) -> None:
        raise NotImplementedError("TODO")


class RateLimitedSender(Sender):
    def __init__(self, inner: Sender, clock: Clock, max_per_window: int, window_ms: int) -> None:
        pass  # TODO

    def send(self, to: str, msg: str) -> None:
        raise NotImplementedError("TODO")


class LoggingSender(Sender):
    def __init__(self, inner: Sender, log: List[str]) -> None:
        pass  # TODO

    def send(self, to: str, msg: str) -> None:
        raise NotImplementedError("TODO")
