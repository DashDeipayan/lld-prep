from __future__ import annotations

from dataclasses import dataclass

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


class VendingMachine:
    # TODO: model the states explicitly (a State base class + one class per state)

    def stock(self, product: Product, qty: int) -> None:
        raise NotImplementedError("TODO")

    def insert_coin(self, coin: int) -> None:
        raise NotImplementedError("TODO")

    def select(self, code: str) -> Dispense:
        raise NotImplementedError("TODO")

    def cancel(self) -> int:
        raise NotImplementedError("TODO")

    def balance(self) -> int:
        raise NotImplementedError("TODO")

    def state(self) -> str:
        """"SOLD_OUT", "IDLE" or "HAS_MONEY"."""
        raise NotImplementedError("TODO")

    def quantity(self, code: str) -> int:
        raise NotImplementedError("TODO")
