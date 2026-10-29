from __future__ import annotations

import threading
from typing import Callable, Dict, Generic, List, Optional, TypeVar

K = TypeVar("K")
V = TypeVar("V")


class _Node:
    __slots__ = ("key", "value", "prev", "next")

    def __init__(self, key=None, value=None) -> None:
        self.key = key
        self.value = value
        self.prev: Optional[_Node] = None
        self.next: Optional[_Node] = None


class LRUCache(Generic[K, V]):
    def __init__(self, capacity: int, on_evict: Optional[Callable[[K, V], None]] = None) -> None:
        if capacity < 1:
            raise ValueError("capacity must be >= 1")
        self._capacity = capacity
        self._on_evict = on_evict
        self._map: Dict[K, _Node] = {}
        # Sentinels: head.next is the most recent, tail.prev the least recent. No None checks needed.
        self._head, self._tail = _Node(), _Node()
        self._head.next, self._tail.prev = self._tail, self._head
        self._lock = threading.Lock()

    def get(self, key: K, default: Optional[V] = None) -> Optional[V]:
        with self._lock:
            node = self._map.get(key)
            if node is None:
                return default
            self._move_to_front(node)
            return node.value

    def put(self, key: K, value: V) -> None:
        evicted = None
        with self._lock:
            node = self._map.get(key)
            if node is not None:
                node.value = value  # update: refresh, never evict
                self._move_to_front(node)
                return
            node = _Node(key, value)
            self._map[key] = node
            self._add_front(node)
            if len(self._map) > self._capacity:
                evicted = self._tail.prev
                self._unlink(evicted)
                del self._map[evicted.key]
        # Callback outside the lock: user code must not run while we hold it.
        if evicted is not None and self._on_evict is not None:
            self._on_evict(evicted.key, evicted.value)

    def size(self) -> int:
        with self._lock:
            return len(self._map)

    def keys(self) -> List[K]:
        with self._lock:
            out = []
            node = self._head.next
            while node is not self._tail:
                out.append(node.key)
                node = node.next
            return out

    # ---- list surgery, O(1) each ----
    def _move_to_front(self, node: _Node) -> None:
        self._unlink(node)
        self._add_front(node)

    def _add_front(self, node: _Node) -> None:
        node.prev, node.next = self._head, self._head.next
        self._head.next.prev = node
        self._head.next = node

    @staticmethod
    def _unlink(node: _Node) -> None:
        node.prev.next = node.next
        node.next.prev = node.prev
