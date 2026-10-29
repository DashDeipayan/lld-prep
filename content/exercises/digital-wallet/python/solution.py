from __future__ import annotations

from dataclasses import dataclass
from typing import List


@dataclass(frozen=True)
class Txn:
    id: str
    sender: str  # "EXTERNAL" for deposits
    receiver: str
    amount: int


class InsufficientFundsError(Exception):
    """Provided."""


class Wallet:
    def open_account(self, account_id: str) -> None:
        raise NotImplementedError("TODO")

    def deposit(self, account_id: str, amount: int) -> str:
        raise NotImplementedError("TODO")

    def transfer(self, idempotency_key: str, sender: str, receiver: str, amount: int) -> str:
        raise NotImplementedError("TODO")

    def balance(self, account_id: str) -> int:
        raise NotImplementedError("TODO")

    def history(self, account_id: str) -> List[Txn]:
        raise NotImplementedError("TODO")
