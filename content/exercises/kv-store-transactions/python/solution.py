from __future__ import annotations

from typing import Optional


class NoTransactionError(Exception):
    """Provided: rollback() or commit() with no open transaction."""


class KVStore:
    def set(self, key: str, value: str) -> None:
        raise NotImplementedError("TODO")

    def get(self, key: str) -> Optional[str]:
        raise NotImplementedError("TODO")

    def delete(self, key: str) -> None:
        raise NotImplementedError("TODO")

    def count(self, value: str) -> int:
        raise NotImplementedError("TODO")

    def begin(self) -> None:
        raise NotImplementedError("TODO")

    def rollback(self) -> None:
        raise NotImplementedError("TODO")

    def commit(self) -> None:
        raise NotImplementedError("TODO")
