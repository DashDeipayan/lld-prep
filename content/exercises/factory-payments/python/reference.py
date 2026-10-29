from __future__ import annotations

import threading
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Callable, Dict, List


@dataclass(frozen=True)
class Receipt:
    method: str
    amount: int
    fee: int


# ---- Provided errors ----
class DuplicateMethodError(Exception):
    """register() was called twice for the same method name."""


class UnsupportedMethodError(ValueError):
    """create() was asked for a method nobody registered."""


class PaymentProcessor(ABC):
    @abstractmethod
    def method(self) -> str: ...

    @abstractmethod
    def pay(self, amount_cents: int) -> Receipt:
        """Raises ValueError when amount_cents <= 0."""


class BaseProcessor(PaymentProcessor):
    """Template Method: validation and receipt building are shared, only the fee differs."""

    def pay(self, amount_cents: int) -> Receipt:
        if amount_cents <= 0:
            raise ValueError("amount must be positive")
        return Receipt(self.method(), amount_cents, self.fee(amount_cents))

    @abstractmethod
    def fee(self, amount_cents: int) -> int: ...


class CardProcessor(BaseProcessor):
    def method(self) -> str:
        return "card"

    def fee(self, amount_cents: int) -> int:
        return amount_cents * 2 // 100


class UpiProcessor(BaseProcessor):
    def method(self) -> str:
        return "upi"

    def fee(self, amount_cents: int) -> int:
        return 0


class WalletProcessor(BaseProcessor):
    def method(self) -> str:
        return "wallet"

    def fee(self, amount_cents: int) -> int:
        return amount_cents // 100


def _normalise(method: str) -> str:
    if method is None:
        raise ValueError("method required")
    return method.strip().lower()


class PaymentFactory:
    def __init__(self) -> None:
        self._registry: Dict[str, Callable[[], PaymentProcessor]] = {}
        self._lock = threading.Lock()  # plugins may register from several threads at startup

    def register(self, method: str, ctor: Callable[[], PaymentProcessor]) -> None:
        key = _normalise(method)
        with self._lock:
            if key in self._registry:
                raise DuplicateMethodError("already registered: " + key)
            self._registry[key] = ctor

    def create(self, method: str) -> PaymentProcessor:
        with self._lock:
            ctor = self._registry.get(_normalise(method))
        if ctor is None:
            raise UnsupportedMethodError("unsupported payment method: %r" % method)
        return ctor()  # a constructor, not a shared instance: processors may hold per-payment state

    def supported_methods(self) -> List[str]:
        with self._lock:
            return sorted(self._registry)

    @classmethod
    def with_defaults(cls) -> "PaymentFactory":
        f = cls()
        f.register("card", CardProcessor)
        f.register("upi", UpiProcessor)
        f.register("wallet", WalletProcessor)
        return f
