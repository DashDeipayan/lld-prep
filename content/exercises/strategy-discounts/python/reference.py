from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Dict, List, Optional


# ---- Provided: do not change Cart's public API ----
@dataclass(frozen=True)
class Line:
    sku: str
    unit_price: int
    qty: int


class Cart:
    def __init__(self) -> None:
        self._lines: Dict[str, Line] = {}

    def add(self, sku: str, unit_price: int, qty: int) -> None:
        old = self._lines.get(sku)
        self._lines[sku] = Line(sku, unit_price, (old.qty if old else 0) + qty)

    def line(self, sku: str) -> Optional[Line]:
        return self._lines.get(sku)

    def subtotal(self) -> int:
        return sum(l.unit_price * l.qty for l in self._lines.values())


# ---- Your design ----
class DiscountStrategy(ABC):
    @abstractmethod
    def discount(self, cart: Cart) -> int:
        """Amount (cents) this promotion takes off the cart. Never negative."""


class PercentageOff(DiscountStrategy):
    def __init__(self, percent: int) -> None:
        if not 0 <= percent <= 100:
            raise ValueError("percent must be 0..100")
        self._percent = percent

    def discount(self, cart: Cart) -> int:
        return cart.subtotal() * self._percent // 100


class FlatOff(DiscountStrategy):
    def __init__(self, amount: int, min_subtotal: int) -> None:
        self._amount = amount
        self._min_subtotal = min_subtotal

    def discount(self, cart: Cart) -> int:
        return self._amount if cart.subtotal() >= self._min_subtotal else 0


class BuyNGetOneFree(DiscountStrategy):
    def __init__(self, sku: str, n: int) -> None:
        if n < 1:
            raise ValueError("n must be >= 1")
        self._sku = sku
        self._n = n

    def discount(self, cart: Cart) -> int:
        line = cart.line(self._sku)
        if line is None:
            return 0
        return (line.qty // (self._n + 1)) * line.unit_price


class Checkout:
    def __init__(self, strategies: List[DiscountStrategy]) -> None:
        self._strategies = tuple(strategies)  # defensive copy

    def total(self, cart: Cart) -> int:
        best = max((s.discount(cart) for s in self._strategies), default=0)
        return max(0, cart.subtotal() - best)
