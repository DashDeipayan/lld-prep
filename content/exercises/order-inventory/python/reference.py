from __future__ import annotations

import threading
from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from typing import Dict, FrozenSet, List


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


# ======================= Model =======================
@dataclass(frozen=True)
class Product:
    sku: str
    price: int


class _Order:
    def __init__(self, order_id: str, customer: str, lines: Dict[str, int], total: int) -> None:
        self.id, self.customer, self.lines, self.total = order_id, customer, lines, total
        self.status = OrderStatus.CREATED


class Inventory:
    """Owns stock counts; every multi-SKU change is atomic under its lock."""

    def __init__(self) -> None:
        self._available: Dict[str, int] = {}
        self._lock = threading.Lock()

    def add(self, sku: str, qty: int) -> None:
        with self._lock:
            self._available[sku] = self._available.get(sku, 0) + qty

    def get(self, sku: str) -> int:
        with self._lock:
            return self._available.get(sku, 0)

    def reserve(self, lines: Dict[str, int]) -> None:
        with self._lock:
            for sku, qty in lines.items():  # validate all...
                if self._available.get(sku, 0) < qty:
                    raise OutOfStockError(sku)
            for sku, qty in lines.items():  # ...then mutate all
                self._available[sku] -= qty

    def release(self, lines: Dict[str, int]) -> None:
        with self._lock:
            for sku, qty in lines.items():
                self._available[sku] = self._available.get(sku, 0) + qty


ALLOWED: Dict[OrderStatus, FrozenSet[OrderStatus]] = {
    OrderStatus.CREATED: frozenset({OrderStatus.PAID, OrderStatus.CANCELLED}),
    OrderStatus.PAID: frozenset({OrderStatus.SHIPPED, OrderStatus.CANCELLED}),
    OrderStatus.SHIPPED: frozenset({OrderStatus.DELIVERED}),
    OrderStatus.DELIVERED: frozenset(),
    OrderStatus.CANCELLED: frozenset(),
}


# ======================= Service =======================
class OrderService:
    def __init__(self, gateway: PaymentGateway) -> None:
        self._gateway = gateway
        self._inventory = Inventory()
        self._products: Dict[str, Product] = {}
        self._orders: Dict[str, _Order] = {}
        self._by_customer: Dict[str, List[str]] = {}
        self._next_order = 1
        self._lock = threading.RLock()

    def add_product(self, sku: str, price: int, stock: int) -> None:
        with self._lock:
            if sku in self._products or price < 0 or stock < 0:
                raise ValueError("invalid product " + sku)
            self._products[sku] = Product(sku, price)
            self._inventory.add(sku, stock)

    def restock(self, sku: str, qty: int) -> None:
        self._product(sku)
        if qty <= 0:
            raise ValueError("qty must be positive")
        self._inventory.add(sku, qty)

    def stock(self, sku: str) -> int:
        self._product(sku)
        return self._inventory.get(sku)

    def place_order(self, customer: str, lines: List[OrderLine]) -> str:
        if not lines:
            raise ValueError("empty order")
        merged: Dict[str, int] = {}
        total = 0
        for line in lines:
            if line.qty <= 0:
                raise ValueError("qty must be positive")
            total += self._product(line.sku).price * line.qty  # price as of placement
            merged[line.sku] = merged.get(line.sku, 0) + line.qty
        self._inventory.reserve(merged)  # atomic, all-or-nothing
        with self._lock:
            o = _Order("O%d" % self._next_order, customer, merged, total)
            self._next_order += 1
            self._orders[o.id] = o
            self._by_customer.setdefault(customer, []).append(o.id)
            return o.id

    def total(self, order_id: str) -> int:
        with self._lock:
            return self._order(order_id).total

    def status(self, order_id: str) -> OrderStatus:
        with self._lock:
            return self._order(order_id).status

    def pay(self, order_id: str) -> None:
        with self._lock:
            o = self._order(order_id)
            self._check(o, OrderStatus.PAID)
            if not self._gateway.charge(o.id, o.total):
                raise PaymentFailedError("charge declined for " + o.id)  # stays CREATED, stock stays reserved
            o.status = OrderStatus.PAID

    def ship(self, order_id: str) -> None:
        with self._lock:
            self._move(self._order(order_id), OrderStatus.SHIPPED)

    def deliver(self, order_id: str) -> None:
        with self._lock:
            self._move(self._order(order_id), OrderStatus.DELIVERED)

    def cancel(self, order_id: str) -> None:
        with self._lock:
            o = self._order(order_id)
            was_paid = o.status is OrderStatus.PAID
            self._move(o, OrderStatus.CANCELLED)
            self._inventory.release(o.lines)
            if was_paid:
                self._gateway.refund(o.id, o.total)

    def orders_of(self, customer: str) -> List[str]:
        with self._lock:
            return list(self._by_customer.get(customer, []))

    def _move(self, o: _Order, to: OrderStatus) -> None:
        self._check(o, to)
        o.status = to

    @staticmethod
    def _check(o: _Order, to: OrderStatus) -> None:
        if to not in ALLOWED[o.status]:
            raise IllegalTransitionError("%s -> %s" % (o.status.value, to.value))

    def _product(self, sku: str) -> Product:
        with self._lock:
            p = self._products.get(sku)
        if p is None:
            raise ValueError("unknown sku " + sku)
        return p

    def _order(self, order_id: str) -> _Order:
        o = self._orders.get(order_id)
        if o is None:
            raise ValueError("unknown order " + order_id)
        return o
