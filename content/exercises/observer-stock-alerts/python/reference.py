from __future__ import annotations

import threading
from typing import Callable, Dict, List

# A listener is any callable taking (symbol, price).
PriceListener = Callable[[str, int], None]


class _Subscription:
    """Wraps each subscription so the same listener can be subscribed twice and removed independently."""

    __slots__ = ("listener",)

    def __init__(self, listener: PriceListener) -> None:
        self.listener = listener


class StockTicker:
    def __init__(self) -> None:
        self._last: Dict[str, int] = {}
        self._subs: List[_Subscription] = []
        self._lock = threading.Lock()

    def subscribe(self, listener: PriceListener) -> Callable[[], None]:
        if listener is None:
            raise ValueError("listener is required")
        sub = _Subscription(listener)
        with self._lock:
            self._subs.append(sub)

        def unsubscribe() -> None:
            with self._lock:
                if sub in self._subs:  # calling it twice is harmless
                    self._subs.remove(sub)

        return unsubscribe

    def update(self, symbol: str, price: int) -> None:
        with self._lock:
            changed = self._last.get(symbol) != price
            self._last[symbol] = price
            snapshot = list(self._subs)  # iterate a copy: listeners may (un)subscribe while notified
        if not changed:
            return
        for sub in snapshot:
            with self._lock:
                still_subscribed = sub in self._subs  # removed earlier in this round?
            if not still_subscribed:
                continue
            try:
                sub.listener(symbol, price)  # called outside the lock, so a listener can't deadlock us
            except Exception:
                pass  # isolate a faulty listener

    def listener_count(self) -> int:
        with self._lock:
            return len(self._subs)


class ThresholdAlert:
    def __init__(self, symbol: str, threshold: int, above: bool, sink: List[str]) -> None:
        self._symbol = symbol
        self._threshold = threshold
        self._above = above
        self._sink = sink
        self._triggered = False

    def __call__(self, symbol: str, price: int) -> None:
        if symbol != self._symbol:
            return
        hit = price >= self._threshold if self._above else price <= self._threshold
        if hit and not self._triggered:  # edge-triggered: fire on the transition only
            self._sink.append("%s%s%d @%d" % (self._symbol, ">=" if self._above else "<=", self._threshold, price))
        self._triggered = hit
