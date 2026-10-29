from __future__ import annotations

import heapq
import threading
from abc import ABC, abstractmethod
from collections import defaultdict
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
def _check(amount: int, participants: List[str]) -> None:
    if amount <= 0:
        raise ValueError("amount must be positive")
    if not participants:
        raise ValueError("no participants")


def _same_keys(keys, participants: List[str]) -> None:
    if set(keys) != set(participants) or len(keys) != len(participants):
        raise ValueError("split keys must match participants")


class EqualSplit(Split):
    def shares(self, amount, participants):
        _check(amount, participants)
        base, rem = divmod(amount, len(participants))
        # the first `rem` participants pay one extra cent, so shares always sum to amount
        return {p: base + (1 if i < rem else 0) for i, p in enumerate(participants)}


class ExactSplit(Split):
    def __init__(self, amounts: Dict[str, int]) -> None:
        self._amounts = dict(amounts)

    def shares(self, amount, participants):
        _check(amount, participants)
        _same_keys(self._amounts, participants)
        out = {p: self._amounts[p] for p in participants}
        if any(v < 0 for v in out.values()):
            raise ValueError("negative share")
        if sum(out.values()) != amount:
            raise ValueError("exact shares sum to %d, expected %d" % (sum(out.values()), amount))
        return out


class PercentSplit(Split):
    def __init__(self, percents: Dict[str, int]) -> None:
        self._percents = dict(percents)

    def shares(self, amount, participants):
        _check(amount, participants)
        _same_keys(self._percents, participants)
        if any(p < 0 for p in self._percents.values()):
            raise ValueError("negative percent")
        if sum(self._percents.values()) != 100:
            raise ValueError("percents sum to %d" % sum(self._percents.values()))
        out = {p: amount * self._percents[p] // 100 for p in participants}
        left = amount - sum(out.values())
        i = 0
        while left > 0:  # hand out the leftover cents in participant order
            out[participants[i % len(participants)]] += 1
            left -= 1
            i += 1
        return out


# ======================= Steps 2-5 =======================
class Splitwise:
    def __init__(self) -> None:
        self._users: Dict[str, None] = {}  # an ordered set
        # net[a][b] = how much a owes b (signed: negative means b owes a)
        self._net: Dict[str, Dict[str, int]] = defaultdict(lambda: defaultdict(int))
        self._next_expense = 1
        self._lock = threading.RLock()

    def add_user(self, user_id: str) -> None:
        with self._lock:
            if user_id in self._users:
                raise ValueError("duplicate user " + user_id)
            self._users[user_id] = None

    def add_expense(self, paid_by: str, amount: int, participants: List[str], split: Split) -> str:
        with self._lock:
            self._require(paid_by)
            if participants is None or len(set(participants)) != len(participants):
                raise ValueError("duplicate participants")
            for p in participants:
                self._require(p)
            shares = split.shares(amount, participants)  # validate fully before touching any balance
            for p, share in shares.items():
                if p != paid_by and share > 0:
                    self._add_debt(p, paid_by, share)
            eid = "E%d" % self._next_expense
            self._next_expense += 1
            return eid

    def owes(self, debtor: str, creditor: str) -> int:
        with self._lock:
            return max(0, self._net.get(debtor, {}).get(creditor, 0))

    def balance(self, user: str) -> int:
        """Positive: others owe this user. Negative: this user owes others."""
        with self._lock:
            return -sum(self._net.get(user, {}).values())

    def settle(self, sender: str, receiver: str, amount: int) -> None:
        with self._lock:
            self._require(sender)
            self._require(receiver)
            if amount <= 0:
                raise ValueError("amount must be positive")
            if amount > self.owes(sender, receiver):
                raise SettlementError("settling more than owed")
            self._add_debt(receiver, sender, amount)  # paying back is a debt in the other direction

    def simplify(self) -> List[Transfer]:
        with self._lock:
            creditors, debtors = [], []  # max-heaps via negated amounts; ties broken by id
            for u in sorted(self._users):
                b = self.balance(u)
                if b > 0:
                    heapq.heappush(creditors, (-b, u))
                elif b < 0:
                    heapq.heappush(debtors, (b, u))
            out = []
            while creditors and debtors:
                c_neg, c = heapq.heappop(creditors)
                d_neg, d = heapq.heappop(debtors)
                x = min(-c_neg, -d_neg)
                out.append(Transfer(d, c, x))
                if -c_neg > x:
                    heapq.heappush(creditors, (c_neg + x, c))
                if -d_neg > x:
                    heapq.heappush(debtors, (d_neg + x, d))
            return out

    def _add_debt(self, debtor: str, creditor: str, amount: int) -> None:
        self._net[debtor][creditor] += amount
        self._net[creditor][debtor] -= amount

    def _require(self, user_id: str) -> None:
        if user_id not in self._users:
            raise ValueError("unknown user " + user_id)
