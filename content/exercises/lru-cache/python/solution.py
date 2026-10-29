from __future__ import annotations

from typing import Callable, Generic, List, Optional, TypeVar

K = TypeVar("K")
V = TypeVar("V")


class LRUCache(Generic[K, V]):
    def __init__(self, capacity: int, on_evict: Optional[Callable[[K, V], None]] = None) -> None:
        pass  # TODO

    def get(self, key: K, default: Optional[V] = None) -> Optional[V]:
        raise NotImplementedError("TODO")

    def put(self, key: K, value: V) -> None:
        raise NotImplementedError("TODO")

    def size(self) -> int:
        raise NotImplementedError("TODO")

    def keys(self) -> List[K]:
        """Most recently used first."""
        raise NotImplementedError("TODO")
