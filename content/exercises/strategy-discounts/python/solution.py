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
        pass  # TODO

    def discount(self, cart: Cart) -> int:
        raise NotImplementedError("TODO")


class FlatOff(DiscountStrategy):
    def __init__(self, amount: int, min_subtotal: int) -> None:
        pass  # TODO

    def discount(self, cart: Cart) -> int:
        raise NotImplementedError("TODO")


class BuyNGetOneFree(DiscountStrategy):
    def __init__(self, sku: str, n: int) -> None:
        pass  # TODO

    def discount(self, cart: Cart) -> int:
        raise NotImplementedError("TODO")


class Checkout:
    def __init__(self, strategies: List[DiscountStrategy]) -> None:
        pass  # TODO

    def total(self, cart: Cart) -> int:
        raise NotImplementedError("TODO")
