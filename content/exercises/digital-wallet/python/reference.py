from __future__ import annotations

import itertools
import threading
from dataclasses import dataclass, field
from typing import Dict, List, Optional


@dataclass(frozen=True)
class Txn:
    id: str
    sender: str  # "EXTERNAL" for deposits
    receiver: str
    amount: int


class InsufficientFundsError(Exception):
    """Provided."""


@dataclass
class _Account:
    id: str
    balance: int = 0
    history: List[Txn] = field(default_factory=list)
    lock: threading.Lock = field(default_factory=threading.Lock)


class _Outcome:
    """The result of the first call for an idempotency key: a txn id or the exception to re-raise."""

    def __init__(self) -> None:
        self.done = threading.Event()
        self.txn_id: Optional[str] = None
        self.error: Optional[BaseException] = None

    def get(self) -> str:
        self.done.wait()
        if self.error is not None:
            raise self.error
        return self.txn_id


class Wallet:
    def __init__(self) -> None:
        self._accounts: Dict[str, _Account] = {}
        self._accounts_lock = threading.Lock()
        self._outcomes: Dict[str, _Outcome] = {}
        self._outcomes_lock = threading.Lock()
        self._ids = itertools.count(1)
        self._ids_lock = threading.Lock()

    def open_account(self, account_id: str) -> None:
        with self._accounts_lock:
            if account_id in self._accounts:
                raise ValueError("account exists: " + account_id)
            self._accounts[account_id] = _Account(account_id)

    def deposit(self, account_id: str, amount: int) -> str:
        if amount <= 0:
            raise ValueError("amount must be positive")
        a = self._account(account_id)
        with a.lock:
            a.balance += amount
            t = Txn(self._next_id(), "EXTERNAL", account_id, amount)
            a.history.append(t)
            return t.id

    def transfer(self, idempotency_key: str, sender: str, receiver: str, amount: int) -> str:
        with self._outcomes_lock:
            outcome = self._outcomes.get(idempotency_key)
            first = outcome is None
            if first:
                outcome = self._outcomes[idempotency_key] = _Outcome()
        if first:  # only the first caller executes; concurrent retries wait for its outcome
            try:
                outcome.txn_id = self._do_transfer(sender, receiver, amount)
            except Exception as e:  # noqa: B902 - remembered and re-raised for every retry
                outcome.error = e
            finally:
                outcome.done.set()
        return outcome.get()

    def _do_transfer(self, sender: str, receiver: str, amount: int) -> str:
        if amount <= 0:
            raise ValueError("amount must be positive")
        if sender == receiver:
            raise ValueError("cannot transfer to the same account")
        a, b = self._account(sender), self._account(receiver)
        first, second = sorted((a, b), key=lambda acc: acc.id)  # global lock order => no deadlock
        with first.lock, second.lock:
            if a.balance < amount:
                raise InsufficientFundsError("%s has %d" % (sender, a.balance))
            a.balance -= amount
            b.balance += amount
            t = Txn(self._next_id(), sender, receiver, amount)
            a.history.append(t)
            b.history.append(t)
            return t.id

    def balance(self, account_id: str) -> int:
        a = self._account(account_id)
        with a.lock:
            return a.balance

    def history(self, account_id: str) -> List[Txn]:
        a = self._account(account_id)
        with a.lock:
            return list(a.history)

    def _account(self, account_id: str) -> _Account:
        with self._accounts_lock:
            a = self._accounts.get(account_id)
        if a is None:
            raise ValueError("unknown account " + account_id)
        return a

    def _next_id(self) -> str:
        with self._ids_lock:
            return "T%d" % next(self._ids)
