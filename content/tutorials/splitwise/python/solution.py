from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Dict, List


# ======================= Provided =======================
@dataclass(frozen=True)
class Transfer:
    sender: str  # `from` is a Python keyword
    receiver: str
    amount: int


class Split(ABC):
    @abstractmethod
    def shares(self, amount: int, participants: List[str]) -> Dict[str, int]:
        """participant -> share, in participant order; shares must sum to amount."""


class SettlementError(Exception):
    """Settling more than is owed."""


# ======================= Step 1 =======================
class EqualSplit(Split):
    def shares(self, amount, participants):
        raise NotImplementedError("TODO")


class ExactSplit(Split):
    def __init__(self, amounts: Dict[str, int]) -> None:
        pass  # TODO

    def shares(self, amount, participants):
        raise NotImplementedError("TODO")


class PercentSplit(Split):
    def __init__(self, percents: Dict[str, int]) -> None:
        pass  # TODO

    def shares(self, amount, participants):
        raise NotImplementedError("TODO")


# ======================= Steps 2-5 =======================
class Splitwise:
    def add_user(self, user_id: str) -> None:
        raise NotImplementedError("TODO")

    def add_expense(self, paid_by: str, amount: int, participants: List[str], split: Split) -> str:
        raise NotImplementedError("TODO")

    def owes(self, debtor: str, creditor: str) -> int:
        raise NotImplementedError("TODO")

    def balance(self, user: str) -> int:
        raise NotImplementedError("TODO")

    def settle(self, sender: str, receiver: str, amount: int) -> None:
        raise NotImplementedError("TODO")

    def simplify(self) -> List[Transfer]:
        raise NotImplementedError("TODO")
