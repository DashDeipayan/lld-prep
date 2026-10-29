from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Callable, List


# ======================= Provided =======================
Clock = Callable[[], int]  # returns today's day number
Notifier = Callable[[str, str], None]  # notifier(member_id, message)


class FinePolicy(ABC):
    @abstractmethod
    def fine(self, days_late: int) -> int: ...


@dataclass(frozen=True)
class Loan:
    id: str
    barcode: str
    isbn: str
    member_id: str
    due_day: int


class LibraryRuleError(Exception):
    """A rule was broken: loan limit, nothing available, not on loan, reserving when available, duplicate reservation."""


# ======================= Your design =======================
class PerDayFine(FinePolicy):
    def __init__(self, per_day: int, cap: int) -> None:
        pass  # TODO

    def fine(self, days_late: int) -> int:
        raise NotImplementedError("TODO")


class Library:
    def __init__(self, fines: FinePolicy, clock: Clock, notifier: Notifier) -> None:
        pass  # TODO

    def add_book(self, isbn: str, title: str, author: str) -> None:
        raise NotImplementedError("TODO")

    def add_copy(self, isbn: str, barcode: str) -> None:
        raise NotImplementedError("TODO")

    def add_member(self, member_id: str, max_loans: int) -> None:
        raise NotImplementedError("TODO")

    def borrow(self, member_id: str, isbn: str) -> Loan:
        raise NotImplementedError("TODO")

    def give_back(self, barcode: str) -> int:
        """Returns the fine."""
        raise NotImplementedError("TODO")

    def reserve(self, member_id: str, isbn: str) -> None:
        raise NotImplementedError("TODO")

    def available_copies(self, isbn: str) -> int:
        raise NotImplementedError("TODO")

    def search(self, query: str) -> List[str]:
        raise NotImplementedError("TODO")
