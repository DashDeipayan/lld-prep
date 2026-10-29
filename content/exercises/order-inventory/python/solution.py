from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from typing import List


# ======================= Provided =======================
@dataclass(frozen=True)
class OrderLine:
    sku: str
    qty: int


class OrderStatus(Enum):
    CREATED = "CREATED"
    PAID = "PAID"
    SHIPPED = "SHIPPED"
    DELIVERED = "DELIVERED"
    CANCELLED = "CANCELLED"


class PaymentGateway(ABC):
    @abstractmethod
    def charge(self, order_id: str, amount: int) -> bool: ...

    @abstractmethod
    def refund(self, order_id: str, amount: int) -> None: ...


class OutOfStockError(Exception):
    pass


class PaymentFailedError(Exception):
    pass


class IllegalTransitionError(Exception):
    pass


# ======================= Your design =======================
class OrderService:
    def __init__(self, gateway: PaymentGateway) -> None:
        pass  # TODO

    def add_product(self, sku: str, price: int, stock: int) -> None:
        raise NotImplementedError("TODO")

    def restock(self, sku: str, qty: int) -> None:
        raise NotImplementedError("TODO")

    def stock(self, sku: str) -> int:
        raise NotImplementedError("TODO")

    def place_order(self, customer: str, lines: List[OrderLine]) -> str:
        raise NotImplementedError("TODO")

    def total(self, order_id: str) -> int:
        raise NotImplementedError("TODO")

    def status(self, order_id: str) -> OrderStatus:
        raise NotImplementedError("TODO")

    def pay(self, order_id: str) -> None:
        raise NotImplementedError("TODO")

    def ship(self, order_id: str) -> None:
        raise NotImplementedError("TODO")

    def deliver(self, order_id: str) -> None:
        raise NotImplementedError("TODO")

    def cancel(self, order_id: str) -> None:
        raise NotImplementedError("TODO")

    def orders_of(self, customer: str) -> List[str]:
        raise NotImplementedError("TODO")
