from __future__ import annotations

import threading
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Dict, Optional

COINS = frozenset({5, 10, 25, 100})


@dataclass(frozen=True)
class Product:
    code: str
    name: str
    price: int


@dataclass(frozen=True)
class Dispense:
    product: Product
    change: int


class MachineStateError(Exception):
    """Provided: the operation isn't allowed in the machine's current state."""


class State(ABC):
    """Each state decides what every operation means while the machine is in it."""

    name = ""

    def stock(self, m: "VendingMachine", product: Product, qty: int) -> None:
        m._add_stock(product, qty)
        m._state = IDLE

    @abstractmethod
    def insert_coin(self, m: "VendingMachine", coin: int) -> None: ...

    @abstractmethod
    def select(self, m: "VendingMachine", code: str) -> Dispense: ...

    def cancel(self, m: "VendingMachine") -> int:
        return 0


class SoldOutState(State):
    name = "SOLD_OUT"

    def insert_coin(self, m, coin):
        raise MachineStateError("sold out")

    def select(self, m, code):
        raise MachineStateError("sold out")


class IdleState(State):
    name = "IDLE"

    def insert_coin(self, m, coin):
        m._balance += coin
        m._state = HAS_MONEY

    def select(self, m, code):
        raise MachineStateError("insert money first")


class HasMoneyState(State):
    name = "HAS_MONEY"

    def stock(self, m, product, qty):
        raise MachineStateError("busy: transaction in progress")

    def insert_coin(self, m, coin):
        m._balance += coin

    def select(self, m, code):
        product = m._products.get(code)
        if product is None:
            raise ValueError("unknown product " + code)
        if m.quantity(code) == 0:
            raise MachineStateError("out of stock: " + code)
        if m._balance < product.price:
            raise MachineStateError("insufficient funds")
        # All checks passed: only now change anything.
        m._inventory[code] -= 1
        change = m._balance - product.price
        m._balance = 0
        m._state = SOLD_OUT if sum(m._inventory.values()) == 0 else IDLE
        return Dispense(product, change)

    def cancel(self, m):
        refund = m._balance
        m._balance = 0
        m._state = IDLE
        return refund


# States hold no data of their own, so one instance of each is shared.
SOLD_OUT, IDLE, HAS_MONEY = SoldOutState(), IdleState(), HasMoneyState()


class VendingMachine:
    def __init__(self) -> None:
        self._products: Dict[str, Product] = {}
        self._inventory: Dict[str, int] = {}
        self._state: State = SOLD_OUT
        self._balance = 0
        self._lock = threading.RLock()  # one physical device: one operation at a time

    def stock(self, product: Product, qty: int) -> None:
        if qty <= 0:
            raise ValueError("qty must be positive")
        with self._lock:
            self._state.stock(self, product, qty)

    def insert_coin(self, coin: int) -> None:
        if coin not in COINS:
            raise ValueError("invalid coin %r" % coin)
        with self._lock:
            self._state.insert_coin(self, coin)

    def select(self, code: str) -> Dispense:
        with self._lock:
            return self._state.select(self, code)

    def cancel(self) -> int:
        with self._lock:
            return self._state.cancel(self)

    def balance(self) -> int:
        with self._lock:
            return self._balance

    def state(self) -> str:
        with self._lock:
            return self._state.name

    def quantity(self, code: str) -> int:
        with self._lock:
            return self._inventory.get(code, 0)

    def _add_stock(self, product: Product, qty: int) -> None:
        self._products[product.code] = product  # latest price and name win
        self._inventory[product.code] = self._inventory.get(product.code, 0) + qty
