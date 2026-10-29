from __future__ import annotations

import threading
from abc import ABC, abstractmethod
from collections import deque
from dataclasses import dataclass, field
from enum import Enum
from typing import Callable, Deque, Dict, List, Optional


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


# ======================= Model =======================
class PerDayFine(FinePolicy):
    def __init__(self, per_day: int, cap: int) -> None:
        self._per_day = per_day
        self._cap = cap

    def fine(self, days_late: int) -> int:
        return min(self._cap, max(0, days_late) * self._per_day)


@dataclass(frozen=True)
class Book:
    isbn: str
    title: str
    author: str


class CopyStatus(Enum):
    AVAILABLE = "AVAILABLE"
    ON_LOAN = "ON_LOAN"
    HELD = "HELD"


@dataclass
class Copy:
    barcode: str
    isbn: str
    status: CopyStatus = CopyStatus.AVAILABLE
    held_for: Optional[str] = None
    held_until: int = 0

    def lapsed(self, today: int) -> bool:
        return self.status is CopyStatus.HELD and today > self.held_until

    def available_to(self, member_id: str, today: int) -> bool:
        if self.status is CopyStatus.AVAILABLE:
            return True
        if self.status is CopyStatus.HELD:
            return self.lapsed(today) or self.held_for == member_id
        return False


@dataclass(frozen=True)
class Member:
    id: str
    max_loans: int


# ======================= Service (facade) =======================
class Library:
    LOAN_DAYS = 14
    HOLD_DAYS = 2

    def __init__(self, fines: FinePolicy, clock: Clock, notifier: Notifier) -> None:
        self._fines = fines
        self._clock = clock
        self._notifier = notifier
        self._books: Dict[str, Book] = {}
        self._copies_by_isbn: Dict[str, Dict[str, Copy]] = {}
        self._copies: Dict[str, Copy] = {}
        self._members: Dict[str, Member] = {}
        self._active_loans: Dict[str, Loan] = {}  # by barcode
        self._reservations: Dict[str, Deque[str]] = {}
        self._next_loan = 1
        self._lock = threading.RLock()

    def add_book(self, isbn: str, title: str, author: str) -> None:
        with self._lock:
            if isbn in self._books:
                raise ValueError("duplicate isbn " + isbn)
            self._books[isbn] = Book(isbn, title, author)
            self._copies_by_isbn[isbn] = {}

    def add_copy(self, isbn: str, barcode: str) -> None:
        with self._lock:
            self._book(isbn)
            if barcode in self._copies:
                raise ValueError("duplicate barcode " + barcode)
            c = self._copies[barcode] = Copy(barcode, isbn)
            self._copies_by_isbn[isbn][barcode] = c

    def add_member(self, member_id: str, max_loans: int) -> None:
        with self._lock:
            if member_id in self._members:
                raise ValueError("duplicate member " + member_id)
            self._members[member_id] = Member(member_id, max_loans)

    def borrow(self, member_id: str, isbn: str) -> Loan:
        with self._lock:
            m = self._member(member_id)
            self._book(isbn)
            active = sum(1 for loan in self._active_loans.values() if loan.member_id == member_id)
            if active >= m.max_loans:
                raise LibraryRuleError("loan limit reached")
            today = self._clock()
            copies = [self._copies_by_isbn[isbn][b] for b in sorted(self._copies_by_isbn[isbn])]
            held = [c for c in copies if c.status is CopyStatus.HELD and c.held_for == member_id and not c.lapsed(today)]
            free = [c for c in copies if c.available_to(member_id, today)]
            chosen = (held or free or [None])[0]  # a copy held for this member first, else the lowest barcode
            if chosen is None:
                raise LibraryRuleError("no copy available")
            chosen.status, chosen.held_for = CopyStatus.ON_LOAN, None
            loan = Loan("L%d" % self._next_loan, chosen.barcode, isbn, member_id, today + self.LOAN_DAYS)
            self._next_loan += 1
            self._active_loans[chosen.barcode] = loan
            return loan

    def give_back(self, barcode: str) -> int:
        with self._lock:
            c = self._copies.get(barcode)
            if c is None:
                raise ValueError("unknown copy " + barcode)
            loan = self._active_loans.pop(barcode, None)
            if loan is None:
                raise LibraryRuleError("copy is not on loan")
            today = self._clock()
            fine = self._fines.fine(max(0, today - loan.due_day))
            queue = self._reservations.get(c.isbn)
            if queue:
                nxt = queue.popleft()
                c.status, c.held_for, c.held_until = CopyStatus.HELD, nxt, today + self.HOLD_DAYS
                self._notifier(nxt, c.isbn + " is ready for pickup")
            else:
                c.status = CopyStatus.AVAILABLE
            return fine

    def reserve(self, member_id: str, isbn: str) -> None:
        with self._lock:
            self._member(member_id)
            self._book(isbn)
            if self.available_copies(isbn) > 0:
                raise LibraryRuleError("a copy is available, borrow it instead")
            queue = self._reservations.setdefault(isbn, deque())
            if member_id in queue:
                raise LibraryRuleError("already reserved")
            queue.append(member_id)

    def available_copies(self, isbn: str) -> int:
        with self._lock:
            self._book(isbn)
            today = self._clock()
            return sum(1 for c in self._copies_by_isbn[isbn].values()
                       if c.status is CopyStatus.AVAILABLE or c.lapsed(today))

    def search(self, query: str) -> List[str]:
        q = query.lower()
        with self._lock:
            hits = [b for b in self._books.values() if q in b.title.lower() or q in b.author.lower()]
            return [b.isbn for b in sorted(hits, key=lambda b: b.title)]

    def _book(self, isbn: str) -> Book:
        b = self._books.get(isbn)
        if b is None:
            raise ValueError("unknown isbn " + isbn)
        return b

    def _member(self, member_id: str) -> Member:
        m = self._members.get(member_id)
        if m is None:
            raise ValueError("unknown member " + member_id)
        return m
