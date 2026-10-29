from __future__ import annotations

import threading
from collections import Counter
from typing import Dict, List, Optional


class NoTransactionError(Exception):
    """Provided: rollback() or commit() with no open transaction."""


_ABSENT = object()  # sentinel: "the key didn't exist" (clearer than overloading None)


class KVStore:
    def __init__(self) -> None:
        self._data: Dict[str, str] = {}
        self._counts: Counter = Counter()
        # One undo log per open block: key -> value before this block first touched it.
        self._undo: List[Dict[str, object]] = []
        self._lock = threading.RLock()

    def set(self, key: str, value: str) -> None:
        if value is None:
            raise ValueError("value is required")
        with self._lock:
            self._write(key, value)

    def get(self, key: str) -> Optional[str]:
        with self._lock:
            return self._data.get(key)

    def delete(self, key: str) -> None:
        with self._lock:
            if key in self._data:
                self._write(key, _ABSENT)

    def count(self, value: str) -> int:
        with self._lock:
            return self._counts[value]

    def begin(self) -> None:
        with self._lock:
            self._undo.append({})

    def rollback(self) -> None:
        with self._lock:
            if not self._undo:
                raise NoTransactionError("no transaction")
            for key, old in self._undo.pop().items():  # O(keys changed in this block)
                self._put(key, old)

    def commit(self) -> None:
        with self._lock:
            if not self._undo:
                raise NoTransactionError("no transaction")
            self._undo.clear()  # changes already live in _data

    def _write(self, key: str, value: object) -> None:
        """A user-visible change: record the old value in the innermost block (first touch only)."""
        if self._undo:
            self._undo[-1].setdefault(key, self._data.get(key, _ABSENT))
        self._put(key, value)

    def _put(self, key: str, value: object) -> None:
        """The single place that keeps _data and _counts consistent."""
        old = self._data.pop(key, _ABSENT) if value is _ABSENT else self._data.get(key, _ABSENT)
        if old is not _ABSENT:
            self._counts[old] -= 1
            if self._counts[old] == 0:
                del self._counts[old]
        if value is not _ABSENT:
            self._data[key] = value
            self._counts[value] += 1
