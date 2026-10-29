from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Callable, List


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


class CardProcessor(PaymentProcessor):
    def method(self) -> str:
        raise NotImplementedError("TODO")

    def pay(self, amount_cents: int) -> Receipt:
        raise NotImplementedError("TODO")


class UpiProcessor(PaymentProcessor):
    def method(self) -> str:
        raise NotImplementedError("TODO")

    def pay(self, amount_cents: int) -> Receipt:
        raise NotImplementedError("TODO")


class WalletProcessor(PaymentProcessor):
    def method(self) -> str:
        raise NotImplementedError("TODO")

    def pay(self, amount_cents: int) -> Receipt:
        raise NotImplementedError("TODO")


class PaymentFactory:
    def register(self, method: str, ctor: Callable[[], PaymentProcessor]) -> None:
        raise NotImplementedError("TODO")

    def create(self, method: str) -> PaymentProcessor:
        raise NotImplementedError("TODO")

    def supported_methods(self) -> List[str]:
        raise NotImplementedError("TODO")

    @classmethod
    def with_defaults(cls) -> "PaymentFactory":
        raise NotImplementedError("TODO")
