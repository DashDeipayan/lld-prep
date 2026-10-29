from __future__ import annotations

from typing import Callable, List

# A listener is any callable taking (symbol, price).
PriceListener = Callable[[str, int], None]


class StockTicker:
    def subscribe(self, listener: PriceListener) -> Callable[[], None]:
        """Returns a function that unsubscribes the listener."""
        raise NotImplementedError("TODO")

    def update(self, symbol: str, price: int) -> None:
        raise NotImplementedError("TODO")

    def listener_count(self) -> int:
        raise NotImplementedError("TODO")


class ThresholdAlert:
    def __init__(self, symbol: str, threshold: int, above: bool, sink: List[str]) -> None:
        pass  # TODO

    def __call__(self, symbol: str, price: int) -> None:
        raise NotImplementedError("TODO")
